package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/providers"
	"github.com/sipeed/makoclaw/pkg/session"
	"github.com/sipeed/makoclaw/pkg/tools"
)

// toolCallProvider returns tool calls on every Chat call.
type toolCallProvider struct {
	toolCalls []providers.ToolCall
}

func (p *toolCallProvider) Chat(_ context.Context, _ []providers.Message, _ []providers.ToolDefinition, _ string, _ map[string]interface{}) (*providers.LLMResponse, error) {
	return &providers.LLMResponse{ToolCalls: p.toolCalls, FinishReason: "tool_calls"}, nil
}

func (p *toolCallProvider) GetDefaultModel() string { return "test-model" }

// errorProvider always returns an error from Chat.
type errorProvider struct{}

func (p *errorProvider) Chat(_ context.Context, _ []providers.Message, _ []providers.ToolDefinition, _ string, _ map[string]interface{}) (*providers.LLMResponse, error) {
	return nil, fmt.Errorf("provider unavailable")
}
func (p *errorProvider) GetDefaultModel() string { return "error-model" }

// fakeFlusher is a minimal tool that calls fn when executed.
type fakeFlusher struct {
	name string
	fn   func()
}

func (f *fakeFlusher) Name() string        { return f.name }
func (f *fakeFlusher) Description() string { return "fake tool for flush testing" }
func (f *fakeFlusher) Parameters() map[string]interface{} {
	return map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
}
func (f *fakeFlusher) Execute(_ context.Context, _ map[string]interface{}) (string, error) {
	f.fn()
	return "ok", nil
}

func TestRunMemoryFlushTurn_NoMemoryTools(t *testing.T) {
	// With no memory tools registered, runMemoryFlushTurn should return nil silently.
	prov := &staticProvider{response: "ok"}
	workspace := t.TempDir()
	cfg := newAgentTestConfig(workspace)
	cfg.Agents.Defaults.MemoryFlushBeforeCompaction = true

	al := &AgentLoop{
		provider:      prov,
		workspace:     workspace,
		model:         "test-model",
		contextWindow: 512,
		sessions:      session.NewSessionManager(filepath.Join(workspace, "sessions")),
		tools:         tools.NewToolRegistry(),
		cfg:           cfg,
	}

	err := al.runMemoryFlushTurn(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRunMemoryFlushTurn_ExecutesToolCalls(t *testing.T) {
	// Provider returns a write_file tool call; verify the tool is executed.
	workspace := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspace, "memory"), 0755); err != nil {
		t.Fatal(err)
	}

	written := false
	fakeTool := &fakeFlusher{name: "write_file", fn: func() { written = true }}

	prov := &toolCallProvider{
		toolCalls: []providers.ToolCall{
			{Name: "write_file", Arguments: map[string]interface{}{"path": "memory/MEMORY.md", "content": "test"}},
		},
	}

	registry := tools.NewToolRegistry()
	registry.Register(fakeTool)

	cfg := newAgentTestConfig(workspace)
	cfg.Agents.Defaults.MemoryFlushBeforeCompaction = true

	al := &AgentLoop{
		provider:      prov,
		workspace:     workspace,
		model:         "test-model",
		contextWindow: 512,
		sessions:      session.NewSessionManager(filepath.Join(workspace, "sessions")),
		tools:         registry,
		cfg:           cfg,
	}

	err := al.runMemoryFlushTurn(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !written {
		t.Fatal("expected write_file tool to be executed")
	}
}

func TestRunMemoryFlushTurn_ProviderError(t *testing.T) {
	// If the LLM call fails, runMemoryFlushTurn returns nil (best-effort).
	workspace := t.TempDir()
	registry := tools.NewToolRegistry()
	registry.Register(&fakeFlusher{name: "write_file", fn: func() {}})

	cfg := newAgentTestConfig(workspace)
	cfg.Agents.Defaults.MemoryFlushBeforeCompaction = true

	al := &AgentLoop{
		provider:      &errorProvider{},
		workspace:     workspace,
		model:         "test-model",
		contextWindow: 512,
		sessions:      session.NewSessionManager(filepath.Join(workspace, "sessions")),
		tools:         registry,
		cfg:           cfg,
	}

	err := al.runMemoryFlushTurn(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("expected nil (best-effort), got %v", err)
	}
}

func TestRunMemoryFlushTurn_RejectsNonAllowlistedTools(t *testing.T) {
	// Provider returns a tool call for "exec" which is NOT in memoryFlushToolNames.
	// The exec tool must NOT be executed even if registered.
	executed := false
	execTool := &fakeFlusher{name: "exec", fn: func() { executed = true }}

	// Also register write_file so the method doesn't exit early (needs at least one memory tool).
	registry := tools.NewToolRegistry()
	registry.Register(execTool)
	registry.Register(&fakeFlusher{name: "write_file", fn: func() {}})

	prov := &toolCallProvider{
		toolCalls: []providers.ToolCall{
			{Name: "exec", Arguments: map[string]interface{}{"command": "rm -rf /"}},
		},
	}

	workspace := t.TempDir()
	cfg := newAgentTestConfig(workspace)
	cfg.Agents.Defaults.MemoryFlushBeforeCompaction = true

	al := &AgentLoop{
		provider:      prov,
		workspace:     workspace,
		model:         "test-model",
		contextWindow: 512,
		sessions:      session.NewSessionManager(filepath.Join(workspace, "sessions")),
		tools:         registry,
		cfg:           cfg,
	}

	err := al.runMemoryFlushTurn(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if executed {
		t.Fatal("exec tool must not be executed during memory flush (not in allowlist)")
	}
}
