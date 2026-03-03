package tools

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/utils"
)

func TestMessageToolName(t *testing.T) {
	tool := NewMessageTool()
	if got := tool.Name(); got != "message" {
		t.Errorf("Name() = %q, want %q", got, "message")
	}
}

func TestMessageToolDescription(t *testing.T) {
	tool := NewMessageTool()
	desc := tool.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestMessageToolParameters(t *testing.T) {
	tool := NewMessageTool()
	params := tool.Parameters()

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected properties to be a map")
	}

	expectedFields := []string{"content", "channel", "chat_id"}
	for _, field := range expectedFields {
		if _, ok := props[field]; !ok {
			t.Errorf("expected %q in parameters properties", field)
		}
	}

	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("expected required to be a string slice")
	}

	found := false
	for _, r := range required {
		if r == "content" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'content' in required list")
	}
}

func TestMessageToolSetContext(t *testing.T) {
	tool := NewMessageTool()
	tool.SetContext("telegram", "12345")

	if tool.defaultChannel != "telegram" {
		t.Errorf("defaultChannel = %q, want %q", tool.defaultChannel, "telegram")
	}
	if tool.defaultChatID != "12345" {
		t.Errorf("defaultChatID = %q, want %q", tool.defaultChatID, "12345")
	}
}

func TestMessageToolExecuteMissingContent(t *testing.T) {
	tool := NewMessageTool()
	_, err := tool.Execute(context.Background(), map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for missing content")
	}
	if !strings.Contains(err.Error(), "content is required") {
		t.Errorf("expected 'content is required' error, got %v", err)
	}
}

func TestMessageToolExecuteWrongContentType(t *testing.T) {
	tool := NewMessageTool()
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"content": 12345,
	})
	if err == nil {
		t.Fatal("expected error for non-string content")
	}
	if !strings.Contains(err.Error(), "content is required") {
		t.Errorf("expected 'content is required' error, got %v", err)
	}
}

func TestMessageToolExecuteNoTarget(t *testing.T) {
	tool := NewMessageTool()
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"content": "hello",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(result, "No target channel/chat") {
		t.Errorf("expected 'No target channel/chat' message, got %q", result)
	}
}

func TestMessageToolExecuteNoCallback(t *testing.T) {
	tool := NewMessageTool()
	tool.SetContext("telegram", "12345")
	// sendCallback is nil

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"content": "hello",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(result, "Message sending not configured") {
		t.Errorf("expected 'Message sending not configured' message, got %q", result)
	}
}

func TestMessageToolExecuteSuccess(t *testing.T) {
	tool := NewMessageTool()
	tool.SetContext("telegram", "12345")

	var sentChannel, sentChatID, sentContent string
	tool.SetSendCallback(func(channel, chatID, content string) error {
		sentChannel = channel
		sentChatID = chatID
		sentContent = content
		return nil
	})

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"content": "hello world",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(result, "Message sent to telegram:12345") {
		t.Errorf("expected success message, got %q", result)
	}
	if sentChannel != "telegram" {
		t.Errorf("callback channel = %q, want %q", sentChannel, "telegram")
	}
	if sentChatID != "12345" {
		t.Errorf("callback chatID = %q, want %q", sentChatID, "12345")
	}
	if sentContent != "hello world" {
		t.Errorf("callback content = %q, want %q", sentContent, "hello world")
	}
}

func TestMessageToolExecuteWithExplicitTarget(t *testing.T) {
	tool := NewMessageTool()
	tool.SetContext("telegram", "default_chat")

	var sentChannel, sentChatID string
	tool.SetSendCallback(func(channel, chatID, content string) error {
		sentChannel = channel
		sentChatID = chatID
		return nil
	})

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"content": "hello",
		"channel": "discord",
		"chat_id": "custom_chat",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(result, "Message sent to discord:custom_chat") {
		t.Errorf("expected success with explicit target, got %q", result)
	}
	if sentChannel != "discord" {
		t.Errorf("callback channel = %q, want %q", sentChannel, "discord")
	}
	if sentChatID != "custom_chat" {
		t.Errorf("callback chatID = %q, want %q", sentChatID, "custom_chat")
	}
}

func TestMessageToolExecuteCallbackError(t *testing.T) {
	tool := NewMessageTool()
	tool.SetContext("telegram", "12345")

	tool.SetSendCallback(func(channel, chatID, content string) error {
		return fmt.Errorf("network timeout")
	})

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"content": "hello",
	})
	if err != nil {
		t.Fatalf("expected no error (error returned in result), got %v", err)
	}
	if !strings.Contains(result, "Error sending message") {
		t.Errorf("expected error message in result, got %q", result)
	}
	if !strings.Contains(result, "network timeout") {
		t.Errorf("expected network timeout in result, got %q", result)
	}
}

func TestMessageToolExecuteWithAgentName(t *testing.T) {
	tool := NewMessageTool()
	tool.SetContext("telegram", "12345")

	var sentContent string
	tool.SetSendCallback(func(channel, chatID, content string) error {
		sentContent = content
		return nil
	})

	ctx := utils.WithAgentName(context.Background(), "researcher")

	_, err := tool.Execute(ctx, map[string]interface{}{
		"content": "Research findings here",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(sentContent, "Research findings here") {
		t.Error("sent content should contain original message")
	}
	if !strings.Contains(sentContent, "Sent by: researcher") {
		t.Errorf("sent content should contain agent signature, got %q", sentContent)
	}
}

func TestMessageToolExecuteWithoutAgentName(t *testing.T) {
	tool := NewMessageTool()
	tool.SetContext("telegram", "12345")

	var sentContent string
	tool.SetSendCallback(func(channel, chatID, content string) error {
		sentContent = content
		return nil
	})

	// Context without agent name
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"content": "Plain message",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should NOT have agent signature
	if strings.Contains(sentContent, "Sent by:") {
		t.Errorf("sent content should not have agent signature when no agent name, got %q", sentContent)
	}
	if sentContent != "Plain message" {
		t.Errorf("sent content = %q, want %q", sentContent, "Plain message")
	}
}

func TestMessageToolFallbackToDefaults(t *testing.T) {
	tool := NewMessageTool()
	tool.SetContext("slack", "general")

	var sentChannel, sentChatID string
	tool.SetSendCallback(func(channel, chatID, content string) error {
		sentChannel = channel
		sentChatID = chatID
		return nil
	})

	// Provide empty channel/chat_id - should fall back to defaults
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"content": "test",
		"channel": "",
		"chat_id": "",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sentChannel != "slack" {
		t.Errorf("should fall back to default channel, got %q", sentChannel)
	}
	if sentChatID != "general" {
		t.Errorf("should fall back to default chatID, got %q", sentChatID)
	}
}
