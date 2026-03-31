package bridge_test

import (
	"encoding/json"
	"testing"

	"github.com/sipeed/makoclaw/pkg/bridge"
)

func TestEvent_IsTerminal(t *testing.T) {
	tests := []struct {
		eventType string
		expected  bool
	}{
		{"result", true},
		{"error", true},
		{"pong", true},
		{"assistant", false},
		{"tool_use", false},
		{"system", false},
	}

	for _, tc := range tests {
		e := bridge.Event{Type: tc.eventType}
		if got := e.IsTerminal(); got != tc.expected {
			t.Errorf("IsTerminal(%q) = %v; want %v", tc.eventType, got, tc.expected)
		}
	}
}

func TestRequest_JSONMarshaling(t *testing.T) {
	req := bridge.Request{
		Command:   "prompt",
		Prompt:    "fix the bug",
		RequestID: "req-123",
		Options: bridge.RequestOptions{
			Model: "sonnet",
			Cwd:   "/fake/path",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var out bridge.Request
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if out.Command != req.Command {
		t.Errorf("got command %q, want %q", out.Command, req.Command)
	}
	if out.Prompt != req.Prompt {
		t.Errorf("got prompt %q, want %q", out.Prompt, req.Prompt)
	}
	if out.Options.Model != req.Options.Model {
		t.Errorf("got model %q, want %q", out.Options.Model, req.Options.Model)
	}
}
