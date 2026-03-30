package channels

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
)

// CommandHandler processes special commands like /setup
type CommandHandler struct {
	store        *storage.Storage
	cancelFunc   func(channel, senderID string) int
	pairingStore *storage.PairingStore
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(store *storage.Storage) *CommandHandler {
	return &CommandHandler{
		store: store,
	}
}

// SetCancelFunc sets the function used to cancel active executions for a channel/sender
func (ch *CommandHandler) SetCancelFunc(fn func(channel, senderID string) int) {
	ch.cancelFunc = fn
}

// SetPairingStore configures the pairing store used by the /approve command
func (ch *CommandHandler) SetPairingStore(ps *storage.PairingStore) {
	ch.pairingStore = ps
}

// HandleCommand processes a command and returns true if handled
func (ch *CommandHandler) HandleCommand(ctx context.Context, channel, senderID, content string) (handled bool, response string, err error) {
	// Check for stop keywords before command prefix check
	if ch.isStopCommand(content) {
		canceled := 0
		if ch.cancelFunc != nil {
			canceled = ch.cancelFunc(channel, senderID)
		}
		return true, fmt.Sprintf("Canceled %d active execution(s).", canceled), nil
	}

	// Check if it's a command
	if !IsCommand(content) {
		return false, "", nil
	}

	cmd, args := ParseCommand(content)

	switch cmd {
	case "setup":
		return ch.handleSetupCommand(ctx, channel, senderID, args)
	case "status":
		return ch.handleStatusCommand(ctx)
	case "approve":
		return ch.handleApproveCommand(ctx, channel, senderID, args)
	default:
		return false, "", nil
	}
}

// handleSetupCommand creates a setup session and returns a setup URL
func (ch *CommandHandler) handleSetupCommand(ctx context.Context, channel, senderID, args string) (bool, string, error) {
	// Create setup session
	session, err := ch.store.CreateSetupSession(channel, senderID, map[string]string{
		"initiated_from": channel,
	})
	if err != nil {
		logger.ErrorCF("channels", "Failed to create setup session", map[string]interface{}{
			"error": err.Error(),
		})
		return true, "Failed to create setup session. Please try again later.", err
	}

	// Generate setup URL (this will be customized per channel)
	setupURL := fmt.Sprintf("https://makoclaw.app/onboarding?token=%s", session.Token)

	response := fmt.Sprintf("🚀 Welcome to makoclaw Setup!\n\nPlease visit this link to complete your configuration:\n%s\n\n⏱️ This link expires in 1 hour.", setupURL)

	logger.InfoCF("channels", "Setup session created", map[string]interface{}{
		"channel":   channel,
		"sender_id": senderID,
		"token":     session.Token,
	})

	return true, response, nil
}

// handleStatusCommand returns current status
func (ch *CommandHandler) handleStatusCommand(ctx context.Context) (bool, string, error) {
	response := "✅ makoclaw is up and running!\n\nUse /setup to begin the onboarding process."
	return true, response, nil
}

// handleApproveCommand approves a pending pairing request by code
func (ch *CommandHandler) handleApproveCommand(_ context.Context, _, _, args string) (bool, string, error) {
	if ch.pairingStore == nil {
		return true, "Pairing store not configured.", nil
	}
	parts := strings.Fields(args)
	if len(parts) < 2 {
		return true, "Usage: /approve <channel> <code>", nil
	}
	targetChannel, code := parts[0], parts[1]
	senderID, err := ch.pairingStore.Approve(targetChannel, code)
	if errors.Is(err, storage.ErrCodeNotFound) {
		return true, fmt.Sprintf("Code %q not found or already used.", code), nil
	}
	if err != nil {
		return true, fmt.Sprintf("Approve failed: %v", err), nil
	}
	return true, fmt.Sprintf("✅ Sender %q approved on channel %q.", senderID, targetChannel), nil
}

// IsCommand checks if content starts with a command
func IsCommand(content string) bool {
	if len(content) == 0 {
		return false
	}
	return content[0] == '/'
}

// channelStopKeywords contains stop/cancel keywords in multiple languages.
// Duplicated from agent.stopKeywords to avoid circular import.
var channelStopKeywords = map[string]struct{}{
	"stop": {}, "cancel": {}, "abort": {}, "halt": {},
	"parar": {}, "cancelar": {}, "detener": {},
	"arrêter": {}, "annuler": {},
	"stopp": {}, "abbrechen": {},
	"abortar": {},
	"停止": {}, "取消": {},
	"中止": {}, "やめて": {}, "キャンセル": {},
	"توقف": {}, "إلغاء": {},
	"रुको": {}, "बंद करो": {}, "रद्द करो": {},
	"стоп": {}, "отмена": {}, "остановить": {},
}

// isStopCommand checks if the content is a stop/cancel command
func (ch *CommandHandler) isStopCommand(content string) bool {
	normalized := strings.ToLower(strings.TrimSpace(content))
	if normalized == "" {
		return false
	}
	_, ok := channelStopKeywords[normalized]
	return ok
}

// ParseCommand extracts the command and arguments
func ParseCommand(content string) (cmd, args string) {
	if len(content) == 0 || content[0] != '/' {
		return "", ""
	}

	// Remove leading /
	content = content[1:]

	// Split on space or newline
	for i, ch := range content {
		if ch == ' ' || ch == '\n' {
			return content[:i], content[i+1:]
		}
	}

	return content, ""
}
