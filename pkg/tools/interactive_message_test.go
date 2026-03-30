package tools

import (
	"context"
	"strings"
	"testing"
)

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
