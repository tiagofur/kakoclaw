package channels

import (
	"context"
	"fmt"
	"strings"

	"github.com/sipeed/kakoclaw/pkg/bus"
)

type Channel interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Send(ctx context.Context, msg bus.OutboundMessage) error
	IsRunning() bool
	IsAllowed(senderID string) bool
	GetUserIDForSender(senderID string) (int64, error) // Extract userID from senderID
}

type BaseChannel struct {
	config       interface{}
	bus          *bus.MessageBus
	running      bool
	name         string
	allowList    []string
	userResolver func(senderID string) (int64, error)
}

func NewBaseChannel(name string, config interface{}, bus *bus.MessageBus, allowList []string) *BaseChannel {
	return &BaseChannel{
		config:    config,
		bus:       bus,
		name:      name,
		allowList: allowList,
		running:   false,
	}
}

func (c *BaseChannel) Name() string {
	return c.name
}

func (c *BaseChannel) IsRunning() bool {
	return c.running
}

// SetUserResolver sets a resolver for senderID -> userID mappings.
func (c *BaseChannel) SetUserResolver(resolver func(senderID string) (int64, error)) {
	c.userResolver = resolver
}

// GetUserIDForSender is a default implementation that returns 0.
// Subclasses should override this to provide proper senderID -> userID mapping.
func (c *BaseChannel) GetUserIDForSender(senderID string) (int64, error) {
	if c.userResolver != nil {
		return c.userResolver(senderID)
	}
	return 0, nil // Default: no mapping, use userID 0
}

func (c *BaseChannel) IsAllowed(senderID string) bool {
	if len(c.allowList) == 0 {
		return true
	}

	// Extract parts from compound senderID like "123456|username"
	idPart := senderID
	userPart := ""
	if idx := strings.Index(senderID, "|"); idx > 0 {
		idPart = senderID[:idx]
		userPart = senderID[idx+1:]
	}

	for _, allowed := range c.allowList {
		// Strip leading "@" from allowed value for username matching
		trimmed := strings.TrimPrefix(allowed, "@")

		// Check exact matches
		if senderID == allowed || idPart == allowed {
			return true
		}

		// Check username matches (with or without @)
		if userPart != "" && (userPart == allowed || userPart == trimmed) {
			return true
		}

		// Check if allowed is a username and matches the userPart
		if trimmed != allowed && userPart == trimmed {
			return true
		}
	}

	return false
}

func (c *BaseChannel) HandleMessage(senderID, chatID, content string, media []string, metadata map[string]string) error {
	if !c.IsAllowed(senderID) {
		return nil
	}

	// Extract userID from senderID using channel-specific mapping
	userID, err := c.GetUserIDForSender(senderID)
	if err != nil {
		// Log error but continue with userID 0 as fallback
		userID = 0
	}

	// Build session key: channel:chatID (will be namespaced by SessionManager if userID > 0)
	sessionKey := fmt.Sprintf("%s:%s", c.name, chatID)

	msg := bus.InboundMessage{
		UserID:     userID,
		Channel:    c.name,
		SenderID:   senderID,
		ChatID:     chatID,
		Content:    content,
		Media:      media,
		SessionKey: sessionKey,
		Metadata:   metadata,
	}

	c.bus.PublishInbound(msg)
	return nil
}

func (c *BaseChannel) setRunning(running bool) {
	c.running = running
}
