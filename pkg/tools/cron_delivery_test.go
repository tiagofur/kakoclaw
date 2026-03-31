package tools

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/cron"
)

// mockExecutor satisfies JobExecutor for delivery tests.
type mockExecutor struct {
	response string
	err      error
}

func (m *mockExecutor) ProcessDirectWithChannel(_ context.Context, _, _, _, _ string) (string, error) {
	return m.response, m.err
}

func (m *mockExecutor) ProcessDirectWithChannelForUser(_ context.Context, _ int64, _, _, _, _ string) (string, error) {
	return m.response, m.err
}

// newDeliveryTestCronTool builds a CronTool wired with a real MessageBus and mock executor.
func newDeliveryTestCronTool(t *testing.T, response string) (*CronTool, *bus.MessageBus) {
	t.Helper()
	mb := bus.NewMessageBus()
	service := cron.NewCronService(filepath.Join(t.TempDir(), "cron", "jobs.json"), nil)
	executor := &mockExecutor{response: response}
	tool := NewCronTool(service, executor, mb)
	tool.SetUserID(1)
	return tool, mb
}

// buildTaskJob constructs a minimal CronJob with type "task".
func buildTaskJob(channel, chatID, name string) *cron.CronJob {
	return &cron.CronJob{
		ID:     "test-job-id",
		UserID: 1,
		Name:   name,
		Payload: cron.CronPayload{
			Channel: channel,
			To:      chatID,
			Message: "do something",
			JobType: "task",
		},
	}
}

// drainOutbound reads one message from the bus without blocking forever.
// Uses a short deadline so the outbound channel (capacity 100) has time to be written.
// Returns (msg, true) if a message was ready, (zero, false) if not.
func drainOutbound(mb *bus.MessageBus) (bus.OutboundMessage, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	return mb.SubscribeOutbound(ctx)
}

// TestExecuteJob_NonCliChannel verifies that a non-cli channel gets an OutboundMessage
// with the correct [Scheduled: name] prefix.
func TestExecuteJob_NonCliChannel(t *testing.T) {
	tool, mb := newDeliveryTestCronTool(t, "weather is sunny")
	defer mb.Close()

	job := buildTaskJob("telegram", "chat123", "morning-report")
	result := tool.ExecuteJob(context.Background(), job)

	if result != "task completed" {
		t.Fatalf("expected 'task completed', got %q", result)
	}

	msg, ok := drainOutbound(mb)
	if !ok {
		t.Fatal("expected an OutboundMessage to be published, but none was")
	}
	if msg.Channel != "telegram" {
		t.Errorf("expected channel 'telegram', got %q", msg.Channel)
	}
	if msg.ChatID != "chat123" {
		t.Errorf("expected chatID 'chat123', got %q", msg.ChatID)
	}
	if !strings.HasPrefix(msg.Content, "[Scheduled: morning-report] ") {
		t.Errorf("expected content to start with '[Scheduled: morning-report] ', got %q", msg.Content)
	}
	if !strings.Contains(msg.Content, "weather is sunny") {
		t.Errorf("expected content to contain response, got %q", msg.Content)
	}
}

// TestExecuteJob_CliChannel verifies that channel "cli" suppresses delivery.
func TestExecuteJob_CliChannel(t *testing.T) {
	tool, mb := newDeliveryTestCronTool(t, "some output")
	defer mb.Close()

	job := buildTaskJob("cli", "direct", "cli-job")
	result := tool.ExecuteJob(context.Background(), job)

	if result != "task completed" {
		t.Fatalf("expected 'task completed', got %q", result)
	}

	_, ok := drainOutbound(mb)
	if ok {
		t.Fatal("expected no OutboundMessage for cli channel, but one was published")
	}
}

// TestExecuteJob_EmptyChannel verifies that empty channel suppresses delivery.
func TestExecuteJob_EmptyChannel(t *testing.T) {
	tool, mb := newDeliveryTestCronTool(t, "some output")
	defer mb.Close()

	// Empty channel — ExecuteJob defaults it to "cli" internally, so delivery must be skipped.
	job := buildTaskJob("", "direct", "empty-channel-job")
	result := tool.ExecuteJob(context.Background(), job)

	if result != "task completed" {
		t.Fatalf("expected 'task completed', got %q", result)
	}

	_, ok := drainOutbound(mb)
	if ok {
		t.Fatal("expected no OutboundMessage for empty channel, but one was published")
	}
}

// TestExecuteJob_EmptyResponse verifies that empty/whitespace response suppresses delivery.
func TestExecuteJob_EmptyResponse(t *testing.T) {
	tool, mb := newDeliveryTestCronTool(t, "   ")
	defer mb.Close()

	job := buildTaskJob("telegram", "chat123", "quiet-job")
	result := tool.ExecuteJob(context.Background(), job)

	if result != "task completed" {
		t.Fatalf("expected 'task completed', got %q", result)
	}

	_, ok := drainOutbound(mb)
	if ok {
		t.Fatal("expected no OutboundMessage for whitespace response, but one was published")
	}
}

// TestExecuteJob_ResponseAtExactLimit verifies that a response exactly at 2000 chars
// is published without truncation.
func TestExecuteJob_ResponseAtExactLimit(t *testing.T) {
	// prefix = "[Scheduled: job] " = 17 chars; maxResponse = 2000 - 17 = 1983 chars
	prefix := "[Scheduled: job] "
	maxResponse := 2000 - len(prefix)
	response := strings.Repeat("x", maxResponse)

	tool, mb := newDeliveryTestCronTool(t, response)
	defer mb.Close()

	job := buildTaskJob("telegram", "chat123", "job")
	tool.ExecuteJob(context.Background(), job)

	msg, ok := drainOutbound(mb)
	if !ok {
		t.Fatal("expected an OutboundMessage to be published, but none was")
	}
	if strings.HasSuffix(msg.Content, "... [truncated]") {
		t.Errorf("expected no truncation for response at exact limit, but content ends with '[truncated]': %q", msg.Content)
	}
	if len(msg.Content) != 2000 {
		t.Errorf("expected total content length 2000, got %d", len(msg.Content))
	}
}

// TestExecuteJob_ResponseExceedsLimit verifies that an oversized response is truncated.
func TestExecuteJob_ResponseExceedsLimit(t *testing.T) {
	prefix := "[Scheduled: job] "
	maxResponse := 2000 - len(prefix)
	// One character over the per-response cap.
	response := strings.Repeat("y", maxResponse+1)

	tool, mb := newDeliveryTestCronTool(t, response)
	defer mb.Close()

	job := buildTaskJob("telegram", "chat123", "job")
	tool.ExecuteJob(context.Background(), job)

	msg, ok := drainOutbound(mb)
	if !ok {
		t.Fatal("expected an OutboundMessage to be published, but none was")
	}
	if !strings.HasSuffix(msg.Content, "... [truncated]") {
		t.Errorf("expected truncated suffix, got %q", msg.Content)
	}
	if !strings.HasPrefix(msg.Content, prefix) {
		t.Errorf("expected prefix %q to be preserved, got %q", prefix, msg.Content)
	}
}
