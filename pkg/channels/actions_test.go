package channels

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestReactionPollerTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	poller := NewReactionPoller()
	result := poller.waitForSignal(ctx, make(chan ApprovalResult))
	if !result.TimedOut {
		t.Error("expected TimedOut to be true")
	}
}

func TestReactionPollerApproval(t *testing.T) {
	ctx := context.Background()
	poller := NewReactionPoller()
	ch := make(chan ApprovalResult, 1)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch <- ApprovalResult{Approved: true, Actor: "user1", Reaction: "👍"}
	}()

	result := poller.waitForSignal(ctx, ch)
	if !result.Approved {
		t.Error("expected Approved to be true")
	}
	if result.TimedOut {
		t.Error("expected TimedOut to be false")
	}
	if result.Actor != "user1" {
		t.Errorf("expected Actor 'user1', got '%s'", result.Actor)
	}
}

func TestReactionPollerRejection(t *testing.T) {
	ctx := context.Background()
	poller := NewReactionPoller()
	ch := make(chan ApprovalResult, 1)

	ch <- ApprovalResult{Approved: false, Actor: "user2", Reaction: "👎"}

	result := poller.waitForSignal(ctx, ch)
	if result.Approved {
		t.Error("expected Approved to be false")
	}
	if result.TimedOut {
		t.Error("expected TimedOut to be false")
	}
	if result.Actor != "user2" {
		t.Errorf("expected Actor 'user2', got '%s'", result.Actor)
	}
	if result.Reaction != "👎" {
		t.Errorf("expected Reaction '👎', got '%s'", result.Reaction)
	}
}

// MockChannelAction implements ChannelAction for testing.
type MockChannelAction struct {
	sentMessages []InteractiveMessageRequest
	approveAfter time.Duration
	shouldReject bool
	shouldError  bool
}

func (m *MockChannelAction) SendInteractiveMessage(ctx context.Context, req InteractiveMessageRequest) (string, error) {
	if m.shouldError {
		return "", fmt.Errorf("mock send error")
	}
	m.sentMessages = append(m.sentMessages, req)
	return fmt.Sprintf("msg-%d", len(m.sentMessages)), nil
}

func (m *MockChannelAction) PollReaction(ctx context.Context, messageID string, timeout time.Duration) (ApprovalResult, error) {
	if m.shouldError {
		return ApprovalResult{TimedOut: true}, fmt.Errorf("mock poll error")
	}

	select {
	case <-time.After(m.approveAfter):
		return ApprovalResult{
			Approved: !m.shouldReject,
			Actor:    "test-user",
			Reaction: "👍",
		}, nil
	case <-ctx.Done():
		return ApprovalResult{TimedOut: true}, nil
	}
}

func TestMockChannelActionSendInteractiveMessage(t *testing.T) {
	mock := &MockChannelAction{}
	ctx := context.Background()

	req := InteractiveMessageRequest{
		Message: "Approve deployment?",
		Actions: []ActionConfig{
			{Label: "Approve", Value: "approve", Style: "primary"},
			{Label: "Reject", Value: "reject", Style: "danger"},
		},
		ChatID: "channel-123",
	}

	msgID, err := mock.SendInteractiveMessage(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgID != "msg-1" {
		t.Errorf("messageID = %q, want %q", msgID, "msg-1")
	}
	if len(mock.sentMessages) != 1 {
		t.Fatalf("expected 1 sent message, got %d", len(mock.sentMessages))
	}
	if mock.sentMessages[0].Message != "Approve deployment?" {
		t.Errorf("message = %q, want %q", mock.sentMessages[0].Message, "Approve deployment?")
	}
	if len(mock.sentMessages[0].Actions) != 2 {
		t.Errorf("actions count = %d, want 2", len(mock.sentMessages[0].Actions))
	}
}

func TestMockChannelActionApprovalFlow(t *testing.T) {
	mock := &MockChannelAction{approveAfter: 50 * time.Millisecond}
	ctx := context.Background()

	// Send interactive message
	msgID, err := mock.SendInteractiveMessage(ctx, InteractiveMessageRequest{
		Message: "Deploy to production?",
		ChatID:  "ch-1",
	})
	if err != nil {
		t.Fatalf("send error: %v", err)
	}

	// Poll for approval
	result, err := mock.PollReaction(ctx, msgID, 1*time.Second)
	if err != nil {
		t.Fatalf("poll error: %v", err)
	}
	if !result.Approved {
		t.Error("expected approval")
	}
	if result.TimedOut {
		t.Error("expected no timeout")
	}
	if result.Actor != "test-user" {
		t.Errorf("Actor = %q, want %q", result.Actor, "test-user")
	}
}

func TestMockChannelActionRejectionFlow(t *testing.T) {
	mock := &MockChannelAction{approveAfter: 50 * time.Millisecond, shouldReject: true}
	ctx := context.Background()

	msgID, _ := mock.SendInteractiveMessage(ctx, InteractiveMessageRequest{
		Message: "Delete database?",
		ChatID:  "ch-1",
	})

	result, err := mock.PollReaction(ctx, msgID, 1*time.Second)
	if err != nil {
		t.Fatalf("poll error: %v", err)
	}
	if result.Approved {
		t.Error("expected rejection")
	}
}

func TestMockChannelActionPollTimeout(t *testing.T) {
	mock := &MockChannelAction{approveAfter: 5 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := mock.PollReaction(ctx, "msg-1", 5*time.Second)
	if err != nil {
		t.Fatalf("poll error: %v", err)
	}
	if !result.TimedOut {
		t.Error("expected timeout")
	}
}

func TestMockChannelActionSendError(t *testing.T) {
	mock := &MockChannelAction{shouldError: true}
	ctx := context.Background()

	_, err := mock.SendInteractiveMessage(ctx, InteractiveMessageRequest{Message: "test"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestInteractiveMessageRequestFields(t *testing.T) {
	req := InteractiveMessageRequest{
		Message: "Approve?",
		Actions: []ActionConfig{
			{Label: "Yes", Value: "yes", Style: "primary"},
			{Label: "No", Value: "no", Style: "danger"},
		},
		ChatID: "123",
	}

	if req.Message != "Approve?" {
		t.Errorf("Message = %q, want %q", req.Message, "Approve?")
	}
	if len(req.Actions) != 2 {
		t.Errorf("Actions length = %d, want 2", len(req.Actions))
	}
	if req.Actions[0].Style != "primary" {
		t.Errorf("Actions[0].Style = %q, want %q", req.Actions[0].Style, "primary")
	}
	if req.Actions[1].Style != "danger" {
		t.Errorf("Actions[1].Style = %q, want %q", req.Actions[1].Style, "danger")
	}
}

func TestApprovalResultFields(t *testing.T) {
	approved := ApprovalResult{Approved: true, Actor: "alice", Reaction: "👍", TimedOut: false}
	if !approved.Approved {
		t.Error("expected Approved")
	}

	rejected := ApprovalResult{Approved: false, Actor: "bob", Reaction: "👎", TimedOut: false}
	if rejected.Approved {
		t.Error("expected not Approved")
	}

	timedOut := ApprovalResult{TimedOut: true}
	if !timedOut.TimedOut {
		t.Error("expected TimedOut")
	}
}
