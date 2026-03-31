package channels

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/utils"
	"github.com/sipeed/makoclaw/pkg/voice"
)

type TelegramChannel struct {
	*BaseChannel
	bot            *telego.Bot
	config         config.TelegramConfig
	chatIDs        sync.Map // senderID -> chatID (concurrent-safe)
	transcriber    *voice.GroqTranscriber
	placeholders   sync.Map // chatID -> messageID
	stopThinking   sync.Map // chatID -> thinkingCancel
	commandHandler *CommandHandler
}

type thinkingCancel struct {
	fn context.CancelFunc
}

func (c *thinkingCancel) Cancel() {
	if c != nil && c.fn != nil {
		c.fn()
	}
}

func NewTelegramChannel(cfg config.TelegramConfig, bus *bus.MessageBus) (*TelegramChannel, error) {
	var opts []telego.BotOption

	if cfg.Proxy != "" {
		proxyURL, parseErr := url.Parse(cfg.Proxy)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid proxy URL %q: %w", cfg.Proxy, parseErr)
		}
		opts = append(opts, telego.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}))
	}

	bot, err := telego.NewBot(cfg.Token, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	base := NewBaseChannel("telegram", cfg, bus, cfg.AllowFrom)

	return &TelegramChannel{
		BaseChannel:    base,
		bot:            bot,
		config:         cfg,
		chatIDs:        sync.Map{},
		transcriber:    nil,
		placeholders:   sync.Map{},
		stopThinking:   sync.Map{},
		commandHandler: nil,
	}, nil
}

func (c *TelegramChannel) SetTranscriber(transcriber *voice.GroqTranscriber) {
	c.transcriber = transcriber
}

// SetCommandHandler sets the command handler for this channel
func (c *TelegramChannel) SetCommandHandler(handler *CommandHandler) {
	c.commandHandler = handler
}

// GetCommandHandler returns the current command handler, or nil.
func (c *TelegramChannel) GetCommandHandler() *CommandHandler {
	return c.commandHandler
}

// cleanupThinking stops the thinking animation and cleans up resources (issue #36)
func (c *TelegramChannel) cleanupThinking(chatID string) {
	if stop, ok := c.stopThinking.Load(chatID); ok {
		if cf, ok := stop.(*thinkingCancel); ok && cf != nil {
			cf.Cancel()
		}
		c.stopThinking.Delete(chatID)
	}
}

func (c *TelegramChannel) Start(ctx context.Context) error {
	logger.InfoC("telegram", "Starting Telegram bot (polling mode)...")

	// Verify connectivity via getMe
	me, err := c.bot.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Telegram: %w", err)
	}

	c.setRunning(true)
	logger.InfoCF("telegram", "Telegram bot connected", map[string]interface{}{
		"username": me.Username,
		"id":       me.ID,
	})

	go func() {
		var offset int
		for {
			select {
			case <-ctx.Done():
				logger.InfoC("telegram", "Polling stopped (context cancelled)")
				return
			default:
			}

			logger.InfoCF("telegram", "Calling getUpdates", map[string]interface{}{
				"offset": offset,
			})

			result, err := c.bot.GetUpdates(ctx, &telego.GetUpdatesParams{
				Offset:  offset,
				Timeout: 30,
				Limit:   100,
			})
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				logger.WarnCF("telegram", "getUpdates error, retrying in 5s", map[string]interface{}{
					"error": err.Error(),
				})
				time.Sleep(5 * time.Second)
				continue
			}

			logger.InfoCF("telegram", "getUpdates returned", map[string]interface{}{
				"count": len(result),
			})

			for _, update := range result {
				if update.UpdateID >= offset {
					offset = update.UpdateID + 1
				}
				logger.InfoCF("telegram", "Processing update", map[string]interface{}{
					"update_id":    update.UpdateID,
					"has_message":  update.Message != nil,
					"has_edited":   update.EditedMessage != nil,
					"has_callback": update.CallbackQuery != nil,
				})
				if update.Message != nil {
					c.handleMessage(ctx, update)
				}
			}
		}
	}()

	return nil
}

func (c *TelegramChannel) Stop(ctx context.Context) error {
	logger.InfoC("telegram", "Stopping Telegram bot...")
	c.setRunning(false)
	return nil
}

func (c *TelegramChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !c.IsRunning() {
		return fmt.Errorf("telegram bot not running")
	}

	chatID, err := parseChatID(msg.ChatID)
	if err != nil {
		// Clean up thinking animation even on error (issue #36)
		c.cleanupThinking(msg.ChatID)
		return fmt.Errorf("invalid chat ID: %w", err)
	}

	// Stop thinking animation
	c.cleanupThinking(msg.ChatID)

	htmlContent := markdownToTelegramHTML(msg.Content)

	// Try to edit placeholder
	if pID, ok := c.placeholders.Load(msg.ChatID); ok {
		c.placeholders.Delete(msg.ChatID)
		editMsg := tu.EditMessageText(tu.ID(chatID), pID.(int), htmlContent)
		editMsg.ParseMode = telego.ModeHTML

		if _, err = c.bot.EditMessageText(ctx, editMsg); err == nil {
			return nil
		}
		// Fallback to new message if edit fails
	}

	tgMsg := tu.Message(tu.ID(chatID), htmlContent)
	tgMsg.ParseMode = telego.ModeHTML

	if _, err = c.bot.SendMessage(ctx, tgMsg); err != nil {
		logger.ErrorCF("telegram", "HTML parse failed, falling back to plain text", map[string]interface{}{
			"error": err.Error(),
		})
		tgMsg.ParseMode = ""
		_, err = c.bot.SendMessage(ctx, tgMsg)
		return err
	}

	return nil
}

func (c *TelegramChannel) handleMessage(ctx context.Context, update telego.Update) {
	message := update.Message
	if message == nil {
		return
	}

	user := message.From
	if user == nil {
		return
	}

	senderID := fmt.Sprintf("%d", user.ID)
	if user.Username != "" {
		senderID = fmt.Sprintf("%d|%s", user.ID, user.Username)
	}

	logger.InfoCF("telegram", "Incoming message", map[string]interface{}{
		"sender_id": senderID,
		"chat_id":   fmt.Sprintf("%d", message.Chat.ID),
		"text":      utils.Truncate(message.Text, 50),
	})

	chatID := message.Chat.ID
	c.chatIDs.Store(senderID, chatID)

	// Extract initial text content for command checking
	initialContent := ""
	if message.Text != "" {
		initialContent = message.Text
	} else if message.Caption != "" {
		initialContent = message.Caption
	}

	// Check for commands BEFORE any allowlist/policy gate so that /approve
	// works even for senders who are not yet paired or in the allowlist.
	if c.commandHandler != nil && IsCommand(initialContent) {
		handled, response, err := c.commandHandler.HandleCommand(ctx, "telegram", senderID, initialContent)
		if handled {
			if err != nil {
				logger.ErrorCF("telegram", "Command error", map[string]interface{}{
					"error": err.Error(),
				})
			}
			// Send command response
			if response != "" {
				_, _ = c.bot.SendMessage(ctx, tu.Message(tu.ID(chatID), response))
			}
			return
		}
	}

	// Gate non-command messages: check allowlist / DM policy before downloading media.
	if !c.ShouldDispatch(senderID) {
		if c.dmPolicy == "pairing" {
			c.issuePairingChallenge(senderID, fmt.Sprintf("%d", chatID))
		}
		logger.InfoCF("telegram", "Message rejected by policy", map[string]interface{}{
			"sender_id": senderID,
			"policy":    c.dmPolicy,
		})
		return
	}

	content := ""
	mediaPaths := []string{}
	localFiles := []string{} // 跟踪需要清理的本地文件

	// 确保临时文件在函数返回时被清理
	defer func() {
		for _, file := range localFiles {
			if err := os.Remove(file); err != nil {
				logger.DebugCF("telegram", "Failed to cleanup temp file", map[string]interface{}{
					"file":  file,
					"error": err.Error(),
				})
			}
		}
	}()

	if message.Text != "" {
		content += message.Text
	}

	if message.Caption != "" {
		if content != "" {
			content += "\n"
		}
		content += message.Caption
	}

	if message.Photo != nil && len(message.Photo) > 0 {
		photo := message.Photo[len(message.Photo)-1]
		photoPath := c.downloadPhoto(ctx, photo.FileID)
		if photoPath != "" {
			localFiles = append(localFiles, photoPath)
			mediaPaths = append(mediaPaths, photoPath)
			if content != "" {
				content += "\n"
			}
			content += fmt.Sprintf("[image: photo]")
		}
	}

	if message.Voice != nil {
		voicePath := c.downloadFile(ctx, message.Voice.FileID, ".ogg")
		if voicePath != "" {
			localFiles = append(localFiles, voicePath)
			mediaPaths = append(mediaPaths, voicePath)

			transcribedText := ""
			if c.transcriber != nil && c.transcriber.IsAvailable() {
				ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
				defer cancel()

				result, err := c.transcriber.Transcribe(ctx, voicePath)
				if err != nil {
					logger.ErrorCF("telegram", "Voice transcription failed", map[string]interface{}{
						"error": err.Error(),
						"path":  voicePath,
					})
					transcribedText = fmt.Sprintf("[voice (transcription failed)]")
				} else {
					transcribedText = fmt.Sprintf("[voice transcription: %s]", result.Text)
					logger.InfoCF("telegram", "Voice transcribed successfully", map[string]interface{}{
						"text": result.Text,
					})
				}
			} else {
				transcribedText = fmt.Sprintf("[voice]")
			}

			if content != "" {
				content += "\n"
			}
			content += transcribedText
		}
	}

	if message.Audio != nil {
		audioPath := c.downloadFile(ctx, message.Audio.FileID, ".mp3")
		if audioPath != "" {
			localFiles = append(localFiles, audioPath)
			mediaPaths = append(mediaPaths, audioPath)
			if content != "" {
				content += "\n"
			}
			content += fmt.Sprintf("[audio]")
		}
	}

	if message.Document != nil {
		docPath := c.downloadFile(ctx, message.Document.FileID, "")
		if docPath != "" {
			localFiles = append(localFiles, docPath)
			mediaPaths = append(mediaPaths, docPath)
			if content != "" {
				content += "\n"
			}
			content += fmt.Sprintf("[file]")
		}
	}

	if content == "" {
		content = "[empty message]"
	}

	logger.InfoCF("telegram", "Received message", map[string]interface{}{
		"sender_id": senderID,
		"chat_id":   fmt.Sprintf("%d", chatID),
		"preview":   utils.Truncate(content, 50),
	})

	// Thinking indicator
	err := c.bot.SendChatAction(ctx, tu.ChatAction(tu.ID(chatID), telego.ChatActionTyping))
	if err != nil {
		logger.ErrorCF("telegram", "Failed to send chat action", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Stop any previous thinking animation (issue #36)
	chatIDStr := fmt.Sprintf("%d", chatID)
	c.cleanupThinking(chatIDStr)

	// Create cancelable context for thinking animation.
	// Primary cancellation: cleanupThinking (called from Send when response arrives).
	// Safety net: 10-minute timeout catches edge cases where the agent loop dies
	// without sending a response (panic, context cancellation, bus failure, etc.).
	// This is intentionally longer than the HTTP provider timeout (300s) to avoid
	// false positives on slow providers.
	thinkCtx, thinkCancel := context.WithTimeout(ctx, 10*time.Minute)
	c.stopThinking.Store(chatIDStr, &thinkingCancel{fn: thinkCancel})

	pMsg, err := c.bot.SendMessage(ctx, tu.Message(tu.ID(chatID), "Thinking... 💭"))
	if err == nil {
		pID := pMsg.MessageID
		c.placeholders.Store(chatIDStr, pID)

		go func(cid int64, mid int) {
			dots := []string{".", "..", "..."}
			emotes := []string{"💭", "🤔", "☁️"}
			i := 0
			ticker := time.NewTicker(2000 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-thinkCtx.Done():
					// If safety-net timeout fired (not normal cancellation),
					// clean up and notify the user.
					if thinkCtx.Err() == context.DeadlineExceeded {
						c.placeholders.Delete(chatIDStr)
						c.stopThinking.Delete(chatIDStr)
						_, sendErr := c.bot.SendMessage(context.Background(), tu.Message(tu.ID(cid), "Sorry, the request took too long. Please try again."))
						if sendErr != nil {
							logger.ErrorCF("telegram", "Failed to send safety-net timeout message", map[string]interface{}{
								"error": sendErr.Error(),
							})
						}
					}
					return
				case <-ticker.C:
					i++
					text := fmt.Sprintf("Thinking%s %s", dots[i%len(dots)], emotes[i%len(emotes)])
					_, editErr := c.bot.EditMessageText(thinkCtx, tu.EditMessageText(tu.ID(chatID), mid, text))
					if editErr != nil {
						logger.DebugCF("telegram", "Failed to edit thinking message", map[string]interface{}{
							"error": editErr.Error(),
						})
						// If edit fails, stop the animation to prevent spam
						return
					}
				}
			}
		}(chatID, pID)
	}

	metadata := map[string]string{
		"message_id": fmt.Sprintf("%d", message.MessageID),
		"user_id":    fmt.Sprintf("%d", user.ID),
		"username":   user.Username,
		"first_name": user.FirstName,
		"is_group":   fmt.Sprintf("%t", message.Chat.Type != "private"),
	}

	_ = c.HandleMessage(senderID, fmt.Sprintf("%d", chatID), content, mediaPaths, metadata)
}

func (c *TelegramChannel) downloadPhoto(ctx context.Context, fileID string) string {
	file, err := c.bot.GetFile(ctx, &telego.GetFileParams{FileID: fileID})
	if err != nil {
		logger.ErrorCF("telegram", "Failed to get photo file", map[string]interface{}{
			"error": err.Error(),
		})
		return ""
	}

	return c.downloadFileWithInfo(file, ".jpg")
}

func (c *TelegramChannel) downloadFileWithInfo(file *telego.File, ext string) string {
	if file.FilePath == "" {
		return ""
	}

	url := c.bot.FileDownloadURL(file.FilePath)
	logger.DebugCF("telegram", "File URL", map[string]interface{}{"url": url})

	// Use FilePath as filename for better identification
	filename := file.FilePath + ext
	return utils.DownloadFile(url, filename, utils.DownloadOptions{
		LoggerPrefix: "telegram",
	})
}

func (c *TelegramChannel) downloadFile(ctx context.Context, fileID, ext string) string {
	file, err := c.bot.GetFile(ctx, &telego.GetFileParams{FileID: fileID})
	if err != nil {
		logger.ErrorCF("telegram", "Failed to get file", map[string]interface{}{
			"error": err.Error(),
		})
		return ""
	}

	return c.downloadFileWithInfo(file, ext)
}

func parseChatID(chatIDStr string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(chatIDStr, "%d", &id)
	return id, err
}

func markdownToTelegramHTML(text string) string {
	if text == "" {
		return ""
	}

	codeBlocks := extractCodeBlocks(text)
	text = codeBlocks.text

	inlineCodes := extractInlineCodes(text)
	text = inlineCodes.text

	text = regexp.MustCompile(`^#{1,6}\s+(.+)$`).ReplaceAllString(text, "$1")

	text = regexp.MustCompile(`^>\s*(.*)$`).ReplaceAllString(text, "$1")

	text = escapeHTML(text)

	text = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(text, `<a href="$2">$1</a>`)

	text = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(text, "<b>$1</b>")

	text = regexp.MustCompile(`__(.+?)__`).ReplaceAllString(text, "<b>$1</b>")

	reItalic := regexp.MustCompile(`_([^_]+)_`)
	text = reItalic.ReplaceAllStringFunc(text, func(s string) string {
		match := reItalic.FindStringSubmatch(s)
		if len(match) < 2 {
			return s
		}
		return "<i>" + match[1] + "</i>"
	})

	text = regexp.MustCompile(`~~(.+?)~~`).ReplaceAllString(text, "<s>$1</s>")

	text = regexp.MustCompile(`^[-*]\s+`).ReplaceAllString(text, "• ")

	for i, code := range inlineCodes.codes {
		escaped := escapeHTML(code)
		text = strings.ReplaceAll(text, fmt.Sprintf("\x00IC%d\x00", i), fmt.Sprintf("<code>%s</code>", escaped))
	}

	for i, code := range codeBlocks.codes {
		escaped := escapeHTML(code)
		text = strings.ReplaceAll(text, fmt.Sprintf("\x00CB%d\x00", i), fmt.Sprintf("<pre><code>%s</code></pre>", escaped))
	}

	return text
}

type codeBlockMatch struct {
	text  string
	codes []string
}

func extractCodeBlocks(text string) codeBlockMatch {
	re := regexp.MustCompile("```[\\w]*\\n?([\\s\\S]*?)```")
	matches := re.FindAllStringSubmatch(text, -1)

	codes := make([]string, 0, len(matches))
	for _, match := range matches {
		codes = append(codes, match[1])
	}

	text = re.ReplaceAllStringFunc(text, func(m string) string {
		return fmt.Sprintf("\x00CB%d\x00", len(codes)-1)
	})

	return codeBlockMatch{text: text, codes: codes}
}

type inlineCodeMatch struct {
	text  string
	codes []string
}

func extractInlineCodes(text string) inlineCodeMatch {
	re := regexp.MustCompile("`([^`]+)`")
	matches := re.FindAllStringSubmatch(text, -1)

	codes := make([]string, 0, len(matches))
	for _, match := range matches {
		codes = append(codes, match[1])
	}

	text = re.ReplaceAllStringFunc(text, func(m string) string {
		return fmt.Sprintf("\x00IC%d\x00", len(codes)-1)
	})

	return inlineCodeMatch{text: text, codes: codes}
}

func escapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	return text
}
