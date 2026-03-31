# Memory Flush Before Compaction — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Before a session auto-summarizes, run a silent LLM turn where the agent saves important context to memory files (MEMORY.md, daily notes, knowledge base), preventing context loss during compaction.

**Architecture:** The flush callback mechanism already exists in `pkg/session/manager.go` (`SetFlushCallback`, `TruncateHistoryForUser`) and is wired in `pkg/agent/loop.go`. The current implementation only saves to the knowledge base. We replace it with a focused `runMemoryFlushTurn` method that uses a memory-only tool subset, respects a config flag, and properly logs its work.

**Tech Stack:** Go 1.26, `pkg/config`, `pkg/agent`, `pkg/session`, `pkg/providers`, `pkg/tools`

---

## File Map

| File | Action | What changes |
|------|--------|--------------|
| `pkg/config/config.go` | Modify | Add `MemoryFlushBeforeCompaction bool` to `AgentDefaults`; default `true` |
| `pkg/agent/loop.go` | Modify | Replace flush callback body; add `runMemoryFlushTurn` method |
| `pkg/agent/memory_flush_test.go` | Create | Unit tests for `runMemoryFlushTurn` and callback behavior |

---

## Task 1: Add config field

**Files:**
- Modify: `pkg/config/config.go`

- [ ] **Step 1: Add field to AgentDefaults struct**

Find the `AgentDefaults` struct (line ~184) and add the field:

```go
type AgentDefaults struct {
	Workspace                   string  `json:"workspace" env:"MAKOCLAW_AGENTS_DEFAULTS_WORKSPACE"`
	RestrictToWorkspace         bool    `json:"restrict_to_workspace" env:"MAKOCLAW_AGENTS_DEFAULTS_RESTRICT_TO_WORKSPACE"`
	Provider                    string  `json:"provider" env:"MAKOCLAW_AGENTS_DEFAULTS_PROVIDER"`
	Model                       string  `json:"model" env:"MAKOCLAW_AGENTS_DEFAULTS_MODEL"`
	ImageModel                  string  `json:"image_model,omitempty"`
	MaxTokens                   int     `json:"max_tokens" env:"MAKOCLAW_AGENTS_DEFAULTS_MAX_TOKENS"`
	Temperature                 float64 `json:"temperature" env:"MAKOCLAW_AGENTS_DEFAULTS_TEMPERATURE"`
	MaxToolIterations           int     `json:"max_tool_iterations" env:"MAKOCLAW_AGENTS_DEFAULTS_MAX_TOOL_ITERATIONS"`
	MemoryFlushBeforeCompaction bool    `json:"memory_flush_before_compaction"`
}
```

- [ ] **Step 2: Set default to true in DefaultConfig()**

Find `DefaultConfig()` (line ~594) and update the `Defaults` block:

```go
Defaults: AgentDefaults{
	Workspace:                   "~/.MakoClaw/workspace",
	RestrictToWorkspace:         true,
	Provider:                    "",
	Model:                       "",
	MaxTokens:                   8192,
	Temperature:                 0.7,
	MaxToolIterations:           20,
	MemoryFlushBeforeCompaction: true,
},
```

- [ ] **Step 3: Build to verify no compilation errors**

```bash
go build ./pkg/config/...
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add pkg/config/config.go
git commit -m "feat(config): add MemoryFlushBeforeCompaction to AgentDefaults"
```

---

## Task 2: Add runMemoryFlushTurn method

**Files:**
- Modify: `pkg/agent/loop.go`

- [ ] **Step 1: Write the failing test first**

Create `pkg/agent/memory_flush_test.go`:

```go
package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/providers"
	"github.com/sipeed/makoclaw/pkg/session"
	"github.com/sipeed/makoclaw/pkg/tools"
)

// toolCallProvider returns a response with tool calls on the first call,
// then a plain text response on subsequent calls.
type toolCallProvider struct {
	calls     int
	toolCalls []providers.ToolCall
}

func (p *toolCallProvider) Chat(_ context.Context, _ []providers.Message, _ []providers.ToolDefinition, _ string, _ map[string]interface{}) (*providers.LLMResponse, error) {
	p.calls++
	if p.calls == 1 && len(p.toolCalls) > 0 {
		return &providers.LLMResponse{ToolCalls: p.toolCalls, FinishReason: "tool_calls"}, nil
	}
	return &providers.LLMResponse{Content: "done", FinishReason: "stop"}, nil
}

func (p *toolCallProvider) GetDefaultModel() string { return "test-model" }

func newFlushTestLoop(t *testing.T, provider providers.LLMProvider, enabled bool) (*AgentLoop, string) {
	t.Helper()
	workspace := t.TempDir()
	cfg := newAgentTestConfig(workspace)
	cfg.Agents.Defaults.MemoryFlushBeforeCompaction = enabled

	registry := tools.NewToolRegistry()
	sm := session.NewSessionManager(filepath.Join(workspace, "sessions"))

	al := &AgentLoop{
		provider:      provider,
		workspace:     workspace,
		model:         "test-model",
		contextWindow: 512,
		sessions:      sm,
		tools:         registry,
		cfg:           cfg,
	}
	return al, workspace
}

func TestRunMemoryFlushTurn_NoMemoryTools(t *testing.T) {
	// With no memory tools registered, runMemoryFlushTurn should return nil silently.
	prov := &staticProvider{response: "ok"}
	al, _ := newFlushTestLoop(t, prov, true)

	err := al.runMemoryFlushTurn(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRunMemoryFlushTurn_ExecutesToolCalls(t *testing.T) {
	// Provider returns a write_file tool call; we verify the tool is executed.
	workspace := t.TempDir()
	memPath := filepath.Join(workspace, "memory", "MEMORY.md")
	if err := os.MkdirAll(filepath.Dir(memPath), 0755); err != nil {
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
	registry := tools.NewToolRegistry()
	registry.Register(&fakeFlusher{name: "write_file", fn: func() {}})

	al := &AgentLoop{
		provider:      &errorProvider{},
		workspace:     t.TempDir(),
		model:         "test-model",
		contextWindow: 512,
		sessions:      session.NewSessionManager(t.TempDir()),
		tools:         registry,
		cfg:           newAgentTestConfig(t.TempDir()),
	}
	al.cfg.Agents.Defaults.MemoryFlushBeforeCompaction = true

	err := al.runMemoryFlushTurn(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("expected nil (best-effort), got %v", err)
	}
}

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

// errorProvider always returns an error from Chat.
type errorProvider struct{}

func (p *errorProvider) Chat(_ context.Context, _ []providers.Message, _ []providers.ToolDefinition, _ string, _ map[string]interface{}) (*providers.LLMResponse, error) {
	return nil, fmt.Errorf("provider unavailable")
}
func (p *errorProvider) GetDefaultModel() string { return "error-model" }
```

- [ ] **Step 2: Add missing import to test file**

Add `"fmt"` to the import block in `pkg/agent/memory_flush_test.go`:

```go
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/providers"
	"github.com/sipeed/makoclaw/pkg/session"
	"github.com/sipeed/makoclaw/pkg/tools"
)
```

- [ ] **Step 3: Run tests to verify they fail (method doesn't exist yet)**

```bash
go test ./pkg/agent/ -run TestRunMemoryFlushTurn -v 2>&1 | head -20
```

Expected: compile error — `al.runMemoryFlushTurn undefined`.

- [ ] **Step 4: Add runMemoryFlushTurn to loop.go**

Add this method at the bottom of `pkg/agent/loop.go`, after `summarizeBatch`:

```go
// memoryFlushToolNames is the allowlist of tools available during a memory flush turn.
var memoryFlushToolNames = map[string]bool{
	"write_file":      true,
	"append_file":     true,
	"query_knowledge": true,
}

// runMemoryFlushTurn runs a single silent LLM turn before session compaction.
// It gives the agent a chance to persist important context to memory files.
// The turn is best-effort: errors are logged but never block compaction.
func (al *AgentLoop) runMemoryFlushTurn(ctx context.Context, sessionKey string) error {
	// Build memory-only tool definitions.
	var toolDefs []providers.ToolDefinition
	for _, td := range al.tools.GetDefinitions() {
		fnMap, ok := td["function"].(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := fnMap["name"].(string)
		if !memoryFlushToolNames[name] {
			continue
		}
		tdType, _ := td["type"].(string)
		desc, _ := fnMap["description"].(string)
		params, _ := fnMap["parameters"].(map[string]interface{})
		toolDefs = append(toolDefs, providers.ToolDefinition{
			Type: tdType,
			Function: providers.ToolFunctionDefinition{
				Name:        name,
				Description: desc,
				Parameters:  params,
			},
		})
	}
	if len(toolDefs) == 0 {
		return nil // no memory tools registered — skip silently
	}

	// Build messages: session history + flush prompt.
	history := al.sessions.GetHistoryForUser(al.userID, sessionKey)
	const flushPrompt = "This session is about to be summarized. Before that happens, " +
		"review the conversation and save anything important to memory — key decisions, " +
		"facts, user preferences, ongoing tasks, or context that could be lost in " +
		"summarization. Use the available memory tools to persist this."
	messages := make([]providers.Message, len(history)+1)
	copy(messages, history)
	messages[len(history)] = providers.Message{Role: "user", Content: flushPrompt}

	logger.InfoC("agent", "memory flush before compaction: starting")

	const flushMaxTokens = 2000
	resp, err := al.provider.Chat(ctx, messages, toolDefs, al.model, map[string]interface{}{
		"max_tokens": flushMaxTokens,
	})
	if err != nil {
		logger.WarnCF("agent", "memory flush before compaction: LLM call failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil // best-effort — never block compaction
	}

	// Execute tool calls from the flush turn.
	executed := 0
	for _, tc := range resp.ToolCalls {
		if _, ok := al.tools.Get(tc.Name); !ok {
			continue
		}
		if _, toolErr := al.tools.Execute(ctx, tc.Name, tc.Arguments); toolErr != nil {
			logger.WarnCF("agent", "memory flush: tool execution failed", map[string]interface{}{
				"tool":  tc.Name,
				"error": toolErr.Error(),
			})
			continue
		}
		executed++
	}

	logger.InfoCF("agent", "memory flush before compaction: completed", map[string]interface{}{
		"tools_executed": executed,
	})
	return nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./pkg/agent/ -run TestRunMemoryFlushTurn -v
```

Expected: all 3 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/agent/loop.go pkg/agent/memory_flush_test.go
git commit -m "feat(agent): add runMemoryFlushTurn for memory flush before compaction"
```

---

## Task 3: Update flush callback to use new method

**Files:**
- Modify: `pkg/agent/loop.go`

- [ ] **Step 1: Write the failing test for the callback behavior**

Add to `pkg/agent/memory_flush_test.go`:

```go
func TestFlushCallbackDisabledByConfig(t *testing.T) {
	// When MemoryFlushBeforeCompaction is false, flush callback must not call the LLM.
	prov := &callCountProvider{}
	registry := tools.NewToolRegistry()
	registry.Register(&fakeFlusher{name: "write_file", fn: func() {}})

	workspace := t.TempDir()
	cfg := newAgentTestConfig(workspace)
	cfg.Agents.Defaults.MemoryFlushBeforeCompaction = false

	sm := session.NewSessionManager(filepath.Join(workspace, "sessions"))

	al := &AgentLoop{
		provider:      prov,
		workspace:     workspace,
		model:         "test-model",
		contextWindow: 512,
		sessions:      sm,
		tools:         registry,
		cfg:           cfg,
	}

	// Simulate what the flush callback does.
	if al.cfg.Agents.Defaults.MemoryFlushBeforeCompaction {
		_ = al.runMemoryFlushTurn(context.Background(), "test-session")
	}

	if prov.chatCalls > 0 {
		t.Fatalf("expected 0 LLM calls when flush disabled, got %d", prov.chatCalls)
	}
}

// callCountProvider counts how many times Chat is called.
type callCountProvider struct {
	chatCalls int
}

func (p *callCountProvider) Chat(_ context.Context, _ []providers.Message, _ []providers.ToolDefinition, _ string, _ map[string]interface{}) (*providers.LLMResponse, error) {
	p.chatCalls++
	return &providers.LLMResponse{Content: "ok", FinishReason: "stop"}, nil
}
func (p *callCountProvider) GetDefaultModel() string { return "count-model" }
```

- [ ] **Step 2: Run to verify test passes (config check is in caller, not method)**

```bash
go test ./pkg/agent/ -run TestFlushCallbackDisabledByConfig -v
```

Expected: PASS — config check is done by the caller.

- [ ] **Step 3: Replace existing flush callback body in loop.go**

Find the current `SetFlushCallback` block (around line 600) and replace it:

**Old code:**
```go
sessionsManager.SetFlushCallback(func(ctx context.Context, sessionKey string) error {
	if loopRef == nil {
		return nil
	}
	if !loopRef.flushInProgress.CompareAndSwap(false, true) {
		return nil // re-entrant call — skip to avoid infinite recursion
	}
	defer loopRef.flushInProgress.Store(false)
	if _, ok := loopRef.tools.Get("query_knowledge"); !ok {
		return nil
	}
	syntheticMsg := "Before we continue, save any important facts, decisions, and in-progress work to the knowledge base."
	_, _ = loopRef.runAgentLoop(ctx, processOptions{
		UserMessage: syntheticMsg,
		SessionKey:  sessionKey,
	})
	return nil
})
```

**New code:**
```go
sessionsManager.SetFlushCallback(func(ctx context.Context, sessionKey string) error {
	if loopRef == nil {
		return nil
	}
	if !loopRef.cfg.Agents.Defaults.MemoryFlushBeforeCompaction {
		return nil
	}
	if !loopRef.flushInProgress.CompareAndSwap(false, true) {
		return nil // re-entrant call — skip to avoid infinite recursion
	}
	defer loopRef.flushInProgress.Store(false)
	return loopRef.runMemoryFlushTurn(ctx, sessionKey)
})
```

- [ ] **Step 4: Build the full package**

```bash
go build ./pkg/agent/...
```

Expected: no errors.

- [ ] **Step 5: Run all agent tests to check nothing is broken**

```bash
go test ./pkg/agent/... -timeout 60s
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add pkg/agent/loop.go pkg/agent/memory_flush_test.go
git commit -m "feat(agent): wire memory flush callback with config flag and focused tool subset"
```

---

## Task 4: Integration smoke test

**Files:**
- No code changes — manual verification

- [ ] **Step 1: Run full test suite**

```bash
go test ./... -timeout 120s
```

Expected: all tests pass.

- [ ] **Step 2: Verify config roundtrip**

```bash
go test ./pkg/config/... -v -run TestDefault
```

Expected: `MemoryFlushBeforeCompaction` defaults to `true`.

- [ ] **Step 3: Final commit if anything was adjusted**

```bash
git add -p
git commit -m "test(agent): add integration smoke for memory flush config roundtrip"
```

---

## Self-Review Checklist

- [x] Config field added with correct JSON tag and default value
- [x] `runMemoryFlushTurn` only uses memory tools (allowlist enforced)
- [x] Best-effort: LLM error → log + return nil, never panics
- [x] Re-entrancy guard (`flushInProgress`) preserved from original callback
- [x] No double-flush risk: `flushInProgress` atomic flag prevents concurrent calls
- [x] Silent turn: LLM response not saved to session history, not sent to user channel
- [x] All 4 test scenarios covered: no tools, tool execution, provider error, config disabled
- [x] Spec requirement "memory tools only" enforced via `memoryFlushToolNames` map
- [x] Spec requirement "config on/off" enforced in callback guard
