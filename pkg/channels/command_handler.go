package channels

import (
	"context"
	"fmt"

	"github.com/sipeed/kakoclaw/pkg/logger"
	"github.com/sipeed/kakoclaw/pkg/storage"
)

// CommandHandler processes special commands like /setup
type CommandHandler struct {
	store *storage.Storage
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(store *storage.Storage) *CommandHandler {
	return &CommandHandler{
		store: store,
	}
}

// HandleCommand processes a command and returns true if handled
func (ch *CommandHandler) HandleCommand(ctx context.Context, channel, senderID, content string) (handled bool, response string, err error) {
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
	setupURL := fmt.Sprintf("https://kakoclaw.app/onboarding?token=%s", session.Token)

	response := fmt.Sprintf("🚀 Welcome to KakoClaw Setup!\n\nPlease visit this link to complete your configuration:\n%s\n\n⏱️ This link expires in 1 hour.", setupURL)

	logger.InfoCF("channels", "Setup session created", map[string]interface{}{
		"channel":   channel,
		"sender_id": senderID,
		"token":     session.Token,
	})

	return true, response, nil
}

// handleStatusCommand returns current status
func (ch *CommandHandler) handleStatusCommand(ctx context.Context) (bool, string, error) {
	response := "✅ KakoClaw is up and running!\n\nUse /setup to begin the onboarding process."
	return true, response, nil
}

// IsCommand checks if content starts with a command
func IsCommand(content string) bool {
	if len(content) == 0 {
		return false
	}
	return content[0] == '/'
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
