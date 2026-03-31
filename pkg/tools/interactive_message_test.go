package tools

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sipeed/makoclaw/pkg/channels"
)

// mockAction implements channels.ChannelAction for testing the tool.
type mockAction struct {
	lastReq   channels.InteractiveMessageRequest
	returnID  string
	returnErr error
}

func (m *mockAction) SendInteractiveMessage(ctx context.Context, req channels.InteractiveMessageRequest) (string, error) {
	m.lastReq = req
	if m.returnErr != nil {
		return "", m.returnErr
	}
	return m.returnID, nil
}

func (m *mockAction) PollReaction(ctx context.Context, messageID string, timeout time.Duration) (channels.ApprovalResult, error) {
	return channels.ApprovalResult{Approved: true}, nil
}

func TestInteractiveMessageToolDisabledReturnsError(t *testing.T) {
	tool := NewInteractiveMessageTool(nil, false, "")
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"message": "approve?",
		"actions": []interface{}{},
		"channel": "discord",
		"chat_id": "123",
	})
	if err == nil {
		t.Error("expected error when channel actions disabled")
	}
	if !strings.Contains(err.Error(), "channel actions") {
		t.Errorf("error should mention 'channel actions', got: %s", err.Error())
	}
}

func TestInteractiveMessageToolNoEndpointReturnsError(t *testing.T) {
	tool := NewInteractiveMessageTool(nil, true, "")
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"message": "approve?",
	})
	if err == nil {
		t.Error("expected error when no endpoint configured")
	}
	if !strings.Contains(err.Error(), "interactions_endpoint_url") {
		t.Errorf("error should mention 'interactions_endpoint_url', got: %s", err.Error())
	}
}

func TestInteractiveMessageToolName(t *testing.T) {
	tool := NewInteractiveMessageTool(nil, false, "")
	if tool.Name() != "interactive_message" {
		t.Errorf("Name() = %q, want %q", tool.Name(), "interactive_message")
	}
}

func TestInteractiveMessageToolSetContext(t *testing.T) {
	tool := NewInteractiveMessageTool(nil, false, "")
	tool.SetContext("discord", "12345")
	if tool.channel != "discord" {
		t.Errorf("channel = %q, want %q", tool.channel, "discord")
	}
	if tool.chatID != "12345" {
		t.Errorf("chatID = %q, want %q", tool.chatID, "12345")
	}
}

func TestInteractiveMessageToolHappyPath(t *testing.T) {
	mock := &mockAction{returnID: "msg-42"}
	tool := NewInteractiveMessageTool(mock, true, "https://example.com/interactions")
	tool.SetContext("discord", "ch-100")

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"message": "Deploy to prod?",
		"actions": []interface{}{
			map[string]interface{}{"label": "Approve", "value": "yes", "style": "primary"},
			map[string]interface{}{"label": "Reject", "value": "no", "style": "danger"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "msg-42") {
		t.Errorf("result should contain message ID, got: %s", result)
	}
	if mock.lastReq.Message != "Deploy to prod?" {
		t.Errorf("message = %q, want %q", mock.lastReq.Message, "Deploy to prod?")
	}
	if len(mock.lastReq.Actions) != 2 {
		t.Errorf("actions count = %d, want 2", len(mock.lastReq.Actions))
	}
	if mock.lastReq.ChatID != "ch-100" {
		t.Errorf("chatID = %q, want %q", mock.lastReq.ChatID, "ch-100")
	}
}

func TestInteractiveMessageToolEmptyMessageError(t *testing.T) {
	mock := &mockAction{returnID: "msg-1"}
	tool := NewInteractiveMessageTool(mock, true, "https://example.com/interactions")

	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"message": "",
	})
	if err == nil {
		t.Error("expected error for empty message")
	}
}

func TestInteractiveMessageToolNilActionHandler(t *testing.T) {
	tool := NewInteractiveMessageTool(nil, true, "https://example.com/interactions")

	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"message": "test",
	})
	if err == nil {
		t.Error("expected error when no action handler")
	}
	if !strings.Contains(err.Error(), "no channel action handler") {
		t.Errorf("error should mention 'no channel action handler', got: %s", err.Error())
	}
}

func TestInteractiveMessageToolSendError(t *testing.T) {
	mock := &mockAction{returnErr: fmt.Errorf("discord API down")}
	tool := NewInteractiveMessageTool(mock, true, "https://example.com/interactions")
	tool.SetContext("discord", "ch-1")

	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"message": "test",
	})
	if err == nil {
		t.Error("expected error on send failure")
	}
	if !strings.Contains(err.Error(), "discord API down") {
		t.Errorf("error should propagate, got: %s", err.Error())
	}
}

func TestInteractiveMessageToolExplicitChannelOverridesContext(t *testing.T) {
	mock := &mockAction{returnID: "msg-1"}
	tool := NewInteractiveMessageTool(mock, true, "https://example.com/interactions")
	tool.SetContext("discord", "default-ch")

	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"message": "test",
		"channel": "slack",
		"chat_id": "override-ch",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.lastReq.ChatID != "override-ch" {
		t.Errorf("chatID = %q, want %q (explicit should override context)", mock.lastReq.ChatID, "override-ch")
	}
}

func TestInteractiveMessageToolDescription(t *testing.T) {
	tool := NewInteractiveMessageTool(nil, false, "")
	if tool.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

func TestInteractiveMessageToolParameters(t *testing.T) {
	tool := NewInteractiveMessageTool(nil, false, "")
	params := tool.Parameters()
	if params == nil {
		t.Fatal("Parameters() should not be nil")
	}
	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Parameters should have properties")
	}
	for _, field := range []string{"message", "actions", "channel", "chat_id"} {
		if _, ok := props[field]; !ok {
			t.Errorf("missing parameter %q", field)
		}
	}
}
