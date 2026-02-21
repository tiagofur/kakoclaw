package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sipeed/kakoclaw/pkg/bus"
	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/logger"
)

// SignalChannel implements Channel interface for Signal messenger
// Uses signal-cli REST API or local command
type SignalChannel struct {
	*BaseChannel
	config      config.SignalConfig
	phoneNumber string // Bot's phone number
}

// NewSignalChannel creates a new Signal channel
func NewSignalChannel(cfg config.SignalConfig, bus *bus.MessageBus) (*SignalChannel, error) {
	if cfg.PhoneNumber == "" {
		return nil, fmt.Errorf("signal phone_number is required")
	}

	base := NewBaseChannel("signal", cfg, bus, cfg.AllowFrom)

	return &SignalChannel{
		BaseChannel: base,
		config:      cfg,
		phoneNumber: cfg.PhoneNumber,
	}, nil
}

// Start begins listening for Signal messages
func (c *SignalChannel) Start(ctx context.Context) error {
	logger.InfoC("signal", "Starting Signal channel...")

	// Check if signal-cli is available
	if err := c.checkSignalCLI(); err != nil {
		return fmt.Errorf("signal-cli not available: %w", err)
	}

	c.setRunning(true)
	logger.InfoCF("signal", "Signal channel started", map[string]interface{}{
		"phone": c.phoneNumber,
	})

	// Start message polling loop
	go c.messageLoop(ctx)

	return nil
}

// Stop stops the Signal channel
func (c *SignalChannel) Stop(ctx context.Context) error {
	logger.InfoC("signal", "Stopping Signal channel...")
	c.setRunning(false)
	return nil
}

// Send sends a message via Signal
func (c *SignalChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !c.IsRunning() {
		return fmt.Errorf("signal channel not running")
	}

	// Extract phone number from chat ID
	phone := msg.ChatID
	if phone == "" {
		return fmt.Errorf("chat_id (phone number) is required")
	}

	// Use signal-cli to send message
	cmd := exec.CommandContext(ctx, "signal-cli", "-a", c.phoneNumber, "send", "-m", msg.Content, phone)
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.ErrorCF("signal", "Failed to send message", map[string]interface{}{
			"error":  err.Error(),
			"output": string(output),
			"to":     phone,
		})
		return fmt.Errorf("failed to send signal message: %w", err)
	}

	logger.DebugCF("signal", "Message sent", map[string]interface{}{
		"to": phone,
	})

	return nil
}

// messageLoop polls for new messages
func (c *SignalChannel) messageLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastTimestamp := int64(0)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !c.IsRunning() {
				return
			}

			messages, err := c.receiveMessages()
			if err != nil {
				logger.ErrorCF("signal", "Failed to receive messages", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}

			for _, msg := range messages {
				// Skip messages from ourselves
				if msg.Source == c.phoneNumber {
					continue
				}

				// Skip old messages
				if msg.Timestamp <= lastTimestamp {
					continue
				}
				lastTimestamp = msg.Timestamp

				c.handleMessage(msg)
			}
		}
	}
}

// handleMessage processes a received Signal message
func (c *SignalChannel) handleMessage(msg SignalMessage) {
	senderID := msg.Source

	if !c.IsAllowed(senderID) {
		logger.DebugCF("signal", "Message rejected by allowlist", map[string]interface{}{
			"sender": senderID,
		})
		return
	}

	content := msg.Message
	if content == "" {
		content = "[empty message]"
	}

	logger.DebugCF("signal", "Received message", map[string]interface{}{
		"sender":  senderID,
		"preview": truncate(content, 50),
	})

	metadata := map[string]string{
		"timestamp": fmt.Sprintf("%d", msg.Timestamp),
		"source":    msg.Source,
	}

	// Use phone number as chat ID for 1:1 conversations
	_ = c.HandleMessage(senderID, senderID, content, nil, metadata)
}

// receiveMessages fetches new messages from signal-cli
func (c *SignalChannel) receiveMessages() ([]SignalMessage, error) {
	// Use signal-cli receive command with JSON output
	cmd := exec.Command("signal-cli", "-a", c.phoneNumber, "receive", "--json")
	output, err := cmd.Output()

	if err != nil {
		// If no messages, signal-cli might return empty output
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) == 0 {
			return []SignalMessage{}, nil
		}
		return nil, fmt.Errorf("signal-cli receive failed: %w", err)
	}

	// Parse JSON array of messages
	var messages []SignalMessage
	if len(output) > 0 {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			var msg SignalMessage
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				logger.DebugCF("signal", "Failed to parse message", map[string]interface{}{
					"error": err.Error(),
					"line":  line,
				})
				continue
			}
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

// checkSignalCLI verifies signal-cli is installed and working
func (c *SignalChannel) checkSignalCLI() error {
	cmd := exec.Command("signal-cli", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("signal-cli not found in PATH. Please install signal-cli: https://github.com/AsamK/signal-cli")
	}

	// Check if account is registered
	cmd = exec.Command("signal-cli", "-a", c.phoneNumber, "listAccounts")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("signal account not registered for %s: %w. Output: %s",
			c.phoneNumber, err, string(output))
	}

	logger.InfoCF("signal", "signal-cli check passed", map[string]interface{}{
		"phone": c.phoneNumber,
	})

	return nil
}

// SignalMessage represents a message from signal-cli
type SignalMessage struct {
	Envelope struct {
		Source       string `json:"source"`
		SourceNumber string `json:"sourceNumber"`
		SourceUUID   string `json:"sourceUuid"`
		Timestamp    int64  `json:"timestamp"`
		DataMessage  *struct {
			Message     string       `json:"message"`
			Timestamp   int64        `json:"timestamp"`
			Attachments []Attachment `json:"attachments,omitempty"`
		} `json:"dataMessage,omitempty"`
	} `json:"envelope"`
	Source    string `json:"source"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID          string `json:"id"`
	ContentType string `json:"contentType"`
	Filename    string `json:"filename"`
	Size        int    `json:"size"`
}

// downloadAttachment downloads an attachment to a temporary file
func (c *SignalChannel) downloadAttachment(attachmentID string) (string, error) {
	// Create temp directory
	tmpDir := filepath.Join(os.TempDir(), "KakoClaw-signal")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Download attachment
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("attachment-%s", attachmentID))

	cmd := exec.Command("signal-cli", "-a", c.phoneNumber, "getAttachment", attachmentID, tmpFile)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to download attachment: %w", err)
	}

	return tmpFile, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
