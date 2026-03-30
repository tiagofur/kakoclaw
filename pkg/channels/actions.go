package channels

import (
	"context"
	"time"
)

// InteractiveMessageRequest describes an interactive message to send to a channel.
type InteractiveMessageRequest struct {
	Message string         `json:"message"`
	Actions []ActionConfig `json:"actions"`
	ChatID  string         `json:"chat_id"`
}

// ActionConfig defines a single interactive action (button, reaction, etc.).
type ActionConfig struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Style string `json:"style"` // "primary" | "danger"
}

// ApprovalResult holds the outcome of an interactive approval flow.
type ApprovalResult struct {
	Approved bool   `json:"approved"`
	Actor    string `json:"actor"`
	Reaction string `json:"reaction"`
	TimedOut bool   `json:"timed_out"`
}

// ChannelAction defines the interface for channels that support interactive messages.
type ChannelAction interface {
	SendInteractiveMessage(ctx context.Context, req InteractiveMessageRequest) (messageID string, err error)
	PollReaction(ctx context.Context, messageID string, timeout time.Duration) (ApprovalResult, error)
}

// ReactionPoller polls for user reactions on a message until approval or timeout.
type ReactionPoller struct {
	interval time.Duration
}

// NewReactionPoller creates a new ReactionPoller with a default 500ms polling interval.
func NewReactionPoller() *ReactionPoller {
	return &ReactionPoller{interval: 500 * time.Millisecond}
}

// waitForSignal waits for an approval signal or context cancellation.
func (p *ReactionPoller) waitForSignal(ctx context.Context, ch <-chan ApprovalResult) ApprovalResult {
	select {
	case result := <-ch:
		return result
	case <-ctx.Done():
		return ApprovalResult{TimedOut: true}
	}
}
