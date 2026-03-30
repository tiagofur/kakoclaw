package channels

import (
	"context"
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
