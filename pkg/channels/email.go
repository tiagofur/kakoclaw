package channels

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	gomessage "github.com/emersion/go-message"
	gomail "github.com/emersion/go-message/mail"
	"github.com/google/uuid"
	"golang.org/x/net/html"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
)

// EmailChannel implements the Channel interface for email via IMAP/SMTP.
type EmailChannel struct {
	*BaseChannel
	config         config.EmailChannelConfig
	bus            *bus.MessageBus
	cancel         context.CancelFunc
	threads        sync.Map // chatID (sender email) -> *threadState
	workspace      string
	commandHandler *CommandHandler
}

// threadState tracks email threading state for a conversation.
type threadState struct {
	LastMessageID string
	References    []string
	Subject       string
}

// uidState persists IMAP UID tracking across restarts.
type uidState struct {
	Mailbox     string `json:"mailbox"`
	UIDValidity uint32 `json:"uid_validity"`
	LastUID     uint32 `json:"last_uid"`
}

// NewEmailChannel creates a new email channel with the given config.
func NewEmailChannel(cfg config.EmailChannelConfig, msgBus *bus.MessageBus, workspace string) *EmailChannel {
	base := NewBaseChannel("email", cfg, msgBus, cfg.AllowFrom)

	return &EmailChannel{
		BaseChannel: base,
		config:      cfg,
		bus:         msgBus,
		workspace:   workspace,
	}
}

// SetCommandHandler sets the command handler for this channel.
func (c *EmailChannel) SetCommandHandler(handler *CommandHandler) {
	c.commandHandler = handler
}

// Start begins polling the IMAP server for new emails.
func (c *EmailChannel) Start(ctx context.Context) error {
	logger.InfoCF("email", "Starting email channel (IMAP polling mode)", map[string]interface{}{
		"imap_host": c.config.IMAPHost,
		"mailbox":   c.mailbox(),
	})

	// Apply defaults
	pollInterval := c.config.PollIntervalSecs
	if pollInterval <= 0 {
		pollInterval = 60
	}

	c.setRunning(true)

	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	go func() {
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()

		var consecutiveFailures int

		// Do an initial fetch immediately
		if err := c.fetchNewEmails(ctx); err != nil {
			logger.WarnCF("email", "Initial email fetch failed", map[string]interface{}{
				"error": err.Error(),
			})
			consecutiveFailures++
		} else {
			consecutiveFailures = 0
		}

		for {
			select {
			case <-ctx.Done():
				logger.InfoC("email", "Email polling stopped (context cancelled)")
				return
			case <-ticker.C:
				// Exponential backoff on consecutive failures (max 5 min)
				if consecutiveFailures > 0 {
					backoff := time.Duration(1<<min(consecutiveFailures, 5)) * time.Second
					if backoff > 5*time.Minute {
						backoff = 5 * time.Minute
					}
					logger.DebugCF("email", "Backing off before next poll", map[string]interface{}{
						"backoff_seconds":      backoff.Seconds(),
						"consecutive_failures": consecutiveFailures,
					})
					select {
					case <-ctx.Done():
						return
					case <-time.After(backoff):
					}
				}

				if err := c.fetchNewEmails(ctx); err != nil {
					consecutiveFailures++
					logger.WarnCF("email", "Email fetch failed", map[string]interface{}{
						"error":                err.Error(),
						"consecutive_failures": consecutiveFailures,
					})
				} else {
					consecutiveFailures = 0
				}
			}
		}
	}()

	logger.InfoC("email", "Email channel started successfully")
	return nil
}

// Stop gracefully stops the email channel.
func (c *EmailChannel) Stop(ctx context.Context) error {
	logger.InfoC("email", "Stopping email channel...")
	if c.cancel != nil {
		c.cancel()
	}
	c.setRunning(false)
	return nil
}

// Send sends a reply email via SMTP with proper threading headers.
func (c *EmailChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !c.IsRunning() {
		return fmt.Errorf("email channel not running")
	}

	to := msg.ChatID // ChatID is the sender's email address
	if to == "" {
		return fmt.Errorf("no recipient email address (empty ChatID)")
	}

	// Validate recipient
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("invalid recipient email %q: %w", to, err)
	}

	from := c.config.From
	if from == "" {
		from = c.config.Username
	}
	if from == "" {
		return fmt.Errorf("no from address configured")
	}

	// Generate a unique Message-ID
	messageID := fmt.Sprintf("<%s@makoclaw>", uuid.New().String())

	// Build threading headers from thread state
	var inReplyTo string
	var references []string
	subject := "Re: (no subject)"

	if ts, ok := c.threads.Load(to); ok {
		state := ts.(*threadState)
		inReplyTo = state.LastMessageID
		references = append([]string{}, state.References...)
		if state.LastMessageID != "" && !contains(references, state.LastMessageID) {
			references = append(references, state.LastMessageID)
		}
		if state.Subject != "" {
			subject = state.Subject
			if !strings.HasPrefix(strings.ToLower(subject), "re:") {
				subject = "Re: " + subject
			}
		}
	}

	// Build RFC 2822 message
	var msgBuilder strings.Builder
	msgBuilder.WriteString("From: " + from + "\r\n")
	msgBuilder.WriteString("To: " + to + "\r\n")
	msgBuilder.WriteString("Subject: " + sanitizeHeader(subject) + "\r\n")
	msgBuilder.WriteString("Message-ID: " + messageID + "\r\n")
	if inReplyTo != "" {
		msgBuilder.WriteString("In-Reply-To: " + inReplyTo + "\r\n")
	}
	if len(references) > 0 {
		msgBuilder.WriteString("References: " + strings.Join(references, " ") + "\r\n")
	}
	msgBuilder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msgBuilder.WriteString("MIME-Version: 1.0\r\n")
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(msg.Content)
	msgBuilder.WriteString("\r\n")

	// Parse from address for envelope
	envelopeFrom, _, err := parseFromAddr(from, c.config.Username)
	if err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}

	// SMTP auth and send
	smtpHost := c.config.SMTPHost
	if smtpHost == "" {
		smtpHost = c.config.IMAPHost // Fallback: some providers share the host
	}
	smtpPort := c.config.SMTPPort
	if smtpPort <= 0 {
		smtpPort = 587
	}

	password := c.config.Password
	if strings.EqualFold(strings.TrimSpace(smtpHost), "smtp.gmail.com") {
		password = strings.ReplaceAll(password, " ", "")
	}

	addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", c.config.Username, password, smtpHost)

	if err := smtp.SendMail(addr, auth, envelopeFrom, []string{to}, []byte(msgBuilder.String())); err != nil {
		logger.ErrorCF("email", "Failed to send email", map[string]interface{}{
			"error": err.Error(),
			"to":    to,
			"host":  smtpHost,
		})
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	// Update thread state with our sent message
	c.threads.Store(to, &threadState{
		LastMessageID: messageID,
		References:    append(references, messageID),
		Subject:       subject,
	})

	logger.InfoCF("email", "Email sent successfully", map[string]interface{}{
		"to":      to,
		"subject": subject,
	})
	return nil
}

// fetchNewEmails connects to IMAP, searches for unseen messages, and processes them.
func (c *EmailChannel) fetchNewEmails(ctx context.Context) error {
	state := c.loadUIDState()

	imapPort := c.config.IMAPPort
	if imapPort <= 0 {
		imapPort = 993
	}
	addr := fmt.Sprintf("%s:%d", c.config.IMAPHost, imapPort)

	// Connect via TLS
	options := &imapclient.Options{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: c.config.InsecureSkipVerify,
		},
	}

	client, err := imapclient.DialTLS(addr, options)
	if err != nil {
		return fmt.Errorf("IMAP connect failed (%s): %w", addr, err)
	}
	defer func() {
		_ = client.Close()
	}()

	// Login
	if err := client.Login(c.config.Username, c.config.Password).Wait(); err != nil {
		return fmt.Errorf("IMAP login failed: %w", err)
	}

	// Select mailbox
	mailbox := c.mailbox()
	selectData, err := client.Select(mailbox, nil).Wait()
	if err != nil {
		return fmt.Errorf("IMAP SELECT %q failed: %w", mailbox, err)
	}

	// Check UID validity — reset tracking if it changed
	if state.UIDValidity != 0 && state.UIDValidity != selectData.UIDValidity {
		logger.WarnCF("email", "IMAP UIDValidity changed, resetting UID state", map[string]interface{}{
			"old_validity": state.UIDValidity,
			"new_validity": selectData.UIDValidity,
		})
		state.LastUID = 0
	}
	state.UIDValidity = selectData.UIDValidity
	state.Mailbox = mailbox

	// Search for unseen messages with UID > lastUID
	criteria := &imap.SearchCriteria{
		NotFlag: []imap.Flag{imap.FlagSeen},
	}
	if state.LastUID > 0 {
		uidSet := imap.UIDSet{imap.UIDRange{Start: imap.UID(state.LastUID + 1), Stop: 0}}
		criteria.UID = []imap.UIDSet{uidSet}
	}

	searchData, err := client.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return fmt.Errorf("IMAP SEARCH failed: %w", err)
	}

	uids := searchData.AllUIDs()
	if len(uids) == 0 {
		logger.DebugCF("email", "No new emails found", map[string]interface{}{
			"mailbox":  mailbox,
			"last_uid": state.LastUID,
		})
		return nil
	}

	logger.InfoCF("email", "Found new emails", map[string]interface{}{
		"count":   len(uids),
		"mailbox": mailbox,
	})

	// Fetch the messages
	uidSet := imap.UIDSetNum(uids...)
	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		UID:      true,
		Flags:    true,
		BodySection: []*imap.FetchItemBodySection{
			{Peek: true}, // Fetch entire body without marking as read
		},
	}

	fetchCmd := client.Fetch(uidSet, fetchOptions)
	defer fetchCmd.Close()

	var maxUID imap.UID
	for {
		msgData := fetchCmd.Next()
		if msgData == nil {
			break
		}

		var uid imap.UID
		var envelope *imap.Envelope
		var bodyReader io.Reader

		// Iterate over fetch items
		for {
			item := msgData.Next()
			if item == nil {
				break
			}

			switch data := item.(type) {
			case imapclient.FetchItemDataUID:
				uid = data.UID
			case imapclient.FetchItemDataEnvelope:
				envelope = data.Envelope
			case imapclient.FetchItemDataBodySection:
				bodyReader = data.Literal
			}
		}

		if uid > maxUID {
			maxUID = uid
		}

		if envelope == nil || bodyReader == nil {
			logger.WarnCF("email", "Skipping email with missing data", map[string]interface{}{
				"uid": uid,
			})
			continue
		}

		// Check size limit
		maxSize := c.config.MaxEmailSizeMB
		if maxSize <= 0 {
			maxSize = 10
		}
		limitedReader := io.LimitReader(bodyReader, int64(maxSize)*1024*1024)

		c.handleEmail(ctx, uid, envelope, limitedReader, client, uidSet)
	}

	// Update lastUID
	if maxUID > imap.UID(state.LastUID) {
		state.LastUID = uint32(maxUID)
	}
	c.saveUIDState(state)

	return nil
}

// handleEmail processes a single email message.
func (c *EmailChannel) handleEmail(ctx context.Context, uid imap.UID, envelope *imap.Envelope, bodyReader io.Reader, client *imapclient.Client, uidSet imap.UIDSet) {
	// Extract sender
	senderAddr := ""
	if len(envelope.From) > 0 {
		senderAddr = envelope.From[0].Addr()
	}
	if senderAddr == "" {
		logger.WarnCF("email", "Skipping email with no sender", map[string]interface{}{
			"uid": uid,
		})
		return
	}

	// Check allow-list
	if !c.IsAllowed(senderAddr) {
		logger.InfoCF("email", "Email rejected by allowlist", map[string]interface{}{
			"sender": senderAddr,
			"uid":    uid,
		})
		return
	}

	// Parse body
	body, err := c.extractPlainText(bodyReader)
	if err != nil {
		logger.WarnCF("email", "Failed to parse email body", map[string]interface{}{
			"error":  err.Error(),
			"uid":    uid,
			"sender": senderAddr,
		})
		return
	}

	body = strings.TrimSpace(body)
	if body == "" {
		logger.DebugCF("email", "Skipping empty email", map[string]interface{}{
			"uid":    uid,
			"sender": senderAddr,
		})
		return
	}

	// Resolve thread root for session key
	threadRoot := resolveThreadRoot(envelope.InReplyTo, envelope.MessageID)

	// Update thread state
	refs := make([]string, 0, len(envelope.InReplyTo)+1)
	refs = append(refs, envelope.InReplyTo...)
	if envelope.MessageID != "" {
		refs = append(refs, envelope.MessageID)
	}
	c.threads.Store(senderAddr, &threadState{
		LastMessageID: envelope.MessageID,
		References:    refs,
		Subject:       envelope.Subject,
	})

	// Build session key from thread root
	sessionKey := fmt.Sprintf("email:%s", threadRoot)

	logger.InfoCF("email", "Processing email", map[string]interface{}{
		"uid":         uid,
		"sender":      senderAddr,
		"subject":     envelope.Subject,
		"session_key": sessionKey,
	})

	// Mark as read if configured
	if c.config.MarkAsRead {
		singleUID := imap.UIDSetNum(uid)
		storeCmd := client.Store(singleUID, &imap.StoreFlags{
			Op:     imap.StoreFlagsAdd,
			Silent: true,
			Flags:  []imap.Flag{imap.FlagSeen},
		}, nil)
		if err := storeCmd.Close(); err != nil {
			logger.WarnCF("email", "Failed to mark email as read", map[string]interface{}{
				"error": err.Error(),
				"uid":   uid,
			})
		}
	}

	// Publish to message bus via BaseChannel.HandleMessage
	c.HandleMessage(senderAddr, senderAddr, body, nil, map[string]string{
		"subject":    envelope.Subject,
		"message_id": envelope.MessageID,
	})
}

// extractPlainText extracts plain text from an email body.
// Prefers text/plain parts, falls back to HTML-to-plaintext conversion.
func (c *EmailChannel) extractPlainText(r io.Reader) (string, error) {
	// Read the entire body first
	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("reading email body: %w", err)
	}

	// Try to parse as a MIME message using go-message
	entity, err := gomessage.Read(strings.NewReader(string(bodyBytes)))
	if err != nil {
		// If MIME parsing fails, treat the whole thing as plain text
		return string(bodyBytes), nil
	}

	// Check if it's multipart
	if mr := entity.MultipartReader(); mr != nil {
		return c.extractFromMultipart(mr)
	}

	// Single-part message
	contentType, _, _ := entity.Header.ContentType()
	bodyContent, err := io.ReadAll(entity.Body)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(contentType, "text/plain") {
		return string(bodyContent), nil
	}
	if strings.HasPrefix(contentType, "text/html") {
		return htmlToPlaintext(string(bodyContent)), nil
	}

	// Unknown content type, try as plain text
	return string(bodyContent), nil
}

// extractFromMultipart walks multipart MIME parts looking for text content.
func (c *EmailChannel) extractFromMultipart(mr gomessage.MultipartReader) (string, error) {
	var plainText string
	var htmlText string

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Try to return what we have
			break
		}

		contentType, _, _ := part.Header.ContentType()

		// Recurse into nested multipart
		if nestedMR := part.MultipartReader(); nestedMR != nil {
			nested, err := c.extractFromMultipart(nestedMR)
			if err == nil && nested != "" {
				if plainText == "" {
					plainText = nested
				}
			}
			continue
		}

		content, err := io.ReadAll(part.Body)
		if err != nil {
			continue
		}

		switch {
		case strings.HasPrefix(contentType, "text/plain"):
			if plainText == "" {
				plainText = string(content)
			}
		case strings.HasPrefix(contentType, "text/html"):
			if htmlText == "" {
				htmlText = string(content)
			}
		}
	}

	// Prefer plain text over HTML
	if plainText != "" {
		return plainText, nil
	}
	if htmlText != "" {
		return htmlToPlaintext(htmlText), nil
	}

	return "", nil
}

// resolveThreadRoot determines the thread root Message-ID for session grouping.
// Uses In-Reply-To references (first entry is the thread root), falling back to own Message-ID.
func resolveThreadRoot(inReplyTo []string, messageID string) string {
	if len(inReplyTo) > 0 && inReplyTo[0] != "" {
		return inReplyTo[0]
	}
	if messageID != "" {
		return messageID
	}
	// Fallback: generate a unique ID
	return uuid.New().String()
}

// htmlToPlaintext strips HTML tags and extracts visible text content.
func htmlToPlaintext(htmlContent string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlContent))
	var result strings.Builder
	var skipContent bool

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			text := strings.TrimSpace(result.String())
			// Collapse excessive whitespace
			lines := strings.Split(text, "\n")
			var cleaned []string
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				cleaned = append(cleaned, trimmed)
			}
			return strings.Join(cleaned, "\n")
		case html.StartTagToken:
			tn, _ := tokenizer.TagName()
			tag := strings.ToLower(string(tn))
			switch tag {
			case "script", "style", "head":
				skipContent = true
			case "br":
				result.WriteString("\n")
			case "p", "div", "li", "tr", "h1", "h2", "h3", "h4", "h5", "h6":
				result.WriteString("\n")
			}
		case html.EndTagToken:
			tn, _ := tokenizer.TagName()
			tag := strings.ToLower(string(tn))
			switch tag {
			case "script", "style", "head":
				skipContent = false
			}
		case html.TextToken:
			if !skipContent {
				result.Write(tokenizer.Text())
			}
		}
	}
}

// mailbox returns the configured mailbox name, defaulting to INBOX.
func (c *EmailChannel) mailbox() string {
	if c.config.Mailbox != "" {
		return c.config.Mailbox
	}
	return "INBOX"
}

// uidStatePath returns the path to the UID state file.
func (c *EmailChannel) uidStatePath() string {
	return filepath.Join(c.workspace, "email_uid.json")
}

// loadUIDState loads the persisted UID tracking state from disk.
func (c *EmailChannel) loadUIDState() uidState {
	var state uidState

	data, err := os.ReadFile(c.uidStatePath())
	if err != nil {
		// First run or missing file — start fresh
		return state
	}

	if err := json.Unmarshal(data, &state); err != nil {
		logger.WarnCF("email", "Failed to parse UID state file, starting fresh", map[string]interface{}{
			"error": err.Error(),
		})
		return uidState{}
	}

	return state
}

// saveUIDState persists the UID tracking state to disk.
func (c *EmailChannel) saveUIDState(state uidState) {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		logger.ErrorCF("email", "Failed to marshal UID state", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if err := os.WriteFile(c.uidStatePath(), data, 0600); err != nil {
		logger.ErrorCF("email", "Failed to save UID state", map[string]interface{}{
			"error": err.Error(),
			"path":  c.uidStatePath(),
		})
	}
}

// sanitizeHeader removes newlines from header values to prevent header injection.
func sanitizeHeader(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return value
}

// parseFromAddr parses the From address into envelope and header forms.
func parseFromAddr(from, fallback string) (envelope, header string, err error) {
	trimmed := strings.TrimSpace(from)
	if trimmed == "" {
		trimmed = strings.TrimSpace(fallback)
	}
	if trimmed == "" {
		return "", "", fmt.Errorf("from address is required")
	}

	parsed, parseErr := mail.ParseAddress(trimmed)
	if parseErr == nil && parsed != nil {
		return parsed.Address, parsed.String(), nil
	}
	if strings.Contains(trimmed, "<") || strings.Contains(trimmed, ">") {
		return "", "", fmt.Errorf("invalid from address %q", from)
	}
	return trimmed, trimmed, nil
}

// contains checks if a string slice contains a value.
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// Ensure EmailChannel satisfies the Channel interface at compile time.
var _ Channel = (*EmailChannel)(nil)

// Suppress unused import warnings.
var _ = gomail.Header{}
