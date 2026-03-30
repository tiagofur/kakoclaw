package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/bus"
)

func TestThinkLevelMap(t *testing.T) {
	cases := []struct {
		name   string
		level  string
		budget ThinkLevel
	}{
		{"off", "off", ThinkOff},
		{"minimal", "minimal", ThinkMinimal},
		{"low", "low", ThinkLow},
		{"medium", "medium", ThinkMedium},
		{"high", "high", ThinkHigh},
		{"xhigh", "xhigh", ThinkXHigh},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := thinkLevelMap[tc.level]
			if !ok {
				t.Fatalf("level %q not found in thinkLevelMap", tc.level)
			}
			if got != tc.budget {
				t.Errorf("level %q: want %d, got %d", tc.level, tc.budget, got)
			}
		})
	}
	if _, ok := thinkLevelMap["ultra"]; ok {
		t.Error("thinkLevelMap must not contain 'ultra'")
	}
}

func TestThinkCommandParsing(t *testing.T) {
	cfg := newAgentTestConfig(t.TempDir())
	loop := NewAgentLoop(cfg, bus.NewMessageBus(), &staticProvider{response: "hello response"})

	// Send /think high — should return ACK without LLM call
	result, err := loop.runAgentLoop(context.Background(), processOptions{
		UserMessage: "/think high",
		SessionKey:  "test:session",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "8192") {
		t.Errorf("expected ACK to contain budget 8192, got: %s", result)
	}
	if !strings.Contains(result, "high") {
		t.Errorf("expected ACK to contain level name 'high', got: %s", result)
	}

	// Verify the session think level was set
	loop.sessionThinkMu.RLock()
	budget := loop.sessionThinkLevel["test:session"]
	loop.sessionThinkMu.RUnlock()
	if budget != ThinkHigh {
		t.Errorf("want ThinkHigh (%d), got %d", ThinkHigh, budget)
	}
}

func TestThinkInvalidLevel(t *testing.T) {
	cfg := newAgentTestConfig(t.TempDir())
	loop := NewAgentLoop(cfg, bus.NewMessageBus(), &staticProvider{response: "unused"})

	_, err := loop.runAgentLoop(context.Background(), processOptions{
		UserMessage: "/think ultra",
		SessionKey:  "test:session",
	})
	if err == nil {
		t.Fatal("expected error for invalid level, got nil")
	}
	// Must list valid level names
	for _, level := range []string{"off", "minimal", "low", "medium", "high", "xhigh"} {
		if !strings.Contains(err.Error(), level) {
			t.Errorf("error message missing level %q: %s", level, err.Error())
		}
	}
	// Budget must remain unchanged (zero = default)
	loop.sessionThinkMu.RLock()
	budget := loop.sessionThinkLevel["test:session"]
	loop.sessionThinkMu.RUnlock()
	if budget != 0 {
		t.Errorf("budget should be unchanged (0), got %d", budget)
	}
}
