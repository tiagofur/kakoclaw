package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
)

func newLLMTaskTestConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Agents.Defaults.Provider = "mock"
	cfg.Agents.Defaults.Model = "mock"
	return cfg
}

func TestLLMTaskHappyPath(t *testing.T) {
	cfg := newLLMTaskTestConfig()
	tool := NewLLMTaskTool(cfg)

	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]interface{}{
		"model": "mock/mock",
		"task":  "hello world",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestLLMTaskHappyPathWithContext(t *testing.T) {
	cfg := newLLMTaskTestConfig()
	tool := NewLLMTaskTool(cfg)

	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]interface{}{
		"model":   "mock/mock",
		"task":    "summarize this",
		"context": "some background info",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestLLMTaskMissingArgs(t *testing.T) {
	cfg := newLLMTaskTestConfig()
	tool := NewLLMTaskTool(cfg)

	ctx := context.Background()
	_, err := tool.Execute(ctx, map[string]interface{}{
		"task": "hello",
		// model missing
	})
	if err == nil {
		t.Fatal("expected error when model is missing")
	}
}

func TestLLMTaskTimeout(t *testing.T) {
	cfg := newLLMTaskTestConfig()
	tool := NewLLMTaskTool(cfg)

	// Pass an already-cancelled context to simulate timeout
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := tool.Execute(ctx, map[string]interface{}{
		"model": "mock/mock",
		"task":  "hello",
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "context") {
		t.Fatalf("expected timeout or context error, got: %v", err)
	}
}
