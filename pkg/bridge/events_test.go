package bridge

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEventSessionResetConstantAndTokensUsedJSON(t *testing.T) {
	ev := Event{
		Type:       EventSessionReset,
		TokensUsed: 123,
	}

	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	jsonText := string(data)
	if !strings.Contains(jsonText, `"event":"session_reset"`) {
		t.Fatalf("expected session_reset event type, got %s", jsonText)
	}
	if !strings.Contains(jsonText, `"tokens_used":123`) {
		t.Fatalf("expected tokens_used field, got %s", jsonText)
	}
}

func TestEvent_IsTerminal(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		expected  bool
	}{
		{"result is terminal", "result", true},
		{"error is terminal", "error", true},
		{"pong is terminal", "pong", true},
		{"assistant is not terminal", "assistant", false},
		{"tool_use is not terminal", "tool_use", false},
		{"session_reset is not terminal", EventSessionReset, false},
		{"empty type is not terminal", "", false},
		{"random type is not terminal", "foo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := Event{Type: tt.eventType}
			if got := ev.IsTerminal(); got != tt.expected {
				t.Errorf("Event.IsTerminal() = %v, want %v", got, tt.expected)
			}
		})
	}
}
