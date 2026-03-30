# Tasks: Phase 8.2 — AI Intelligence

## Phase 1: Foundation — Types and Struct Fields

- [ ] 1.1 In `pkg/agent/loop.go`, add `ThinkLevel` type and six constants (`ThinkOff=0`, `ThinkMinimal=512`, `ThinkLow=1024`, `ThinkMedium=4096`, `ThinkHigh=8192`, `ThinkXHigh=16000`) plus `thinkLevelMap map[string]ThinkLevel` above the `AgentLoop` struct.
- [ ] 1.2 Add `sessionThinkLevel map[string]ThinkLevel` and `sessionThinkMu sync.RWMutex` fields to the `AgentLoop` struct literal; initialize the map in `NewAgentLoop` (same pattern as `summarizing sync.Map`).
- [ ] 1.3 In `pkg/session/manager.go`, define `type FlushCallback func(ctx context.Context, sessionKey string) error`; add `flushCallback FlushCallback` field to `SessionManager` struct; add `func (sm *SessionManager) SetFlushCallback(fn FlushCallback)` setter.
- [ ] 1.4 **RED** — Write `TestThinkLevelMap` in `pkg/agent/loop_think_test.go` (new file): table-driven test asserting each level name maps to the correct budget and that `"ultra"` is not present. Run: `go test ./pkg/agent/ -run TestThinkLevelMap` — expect compile or assertion failure.

## Phase 2: `/think` Command Parsing

- [ ] 2.1 In `pkg/agent/loop.go`, inside `runAgentLoop` (and `runAgentLoopStream` if it exists as a separate entry point), add `/think` prefix check at the top of message processing, before any LLM call. Parse the level word, look up `thinkLevelMap`, write to `al.sessionThinkMu`-guarded `al.sessionThinkLevel[opts.SessionKey]`, and return an ACK string `"Thinking level set to <level> (<budget> tokens)"` without calling the LLM. Return descriptive error for unknown levels listing all valid names.
- [ ] 2.2 Replace the hardcoded `llmOpts["thinking_budget_tokens"] = 1024` at line ~1798 in `pkg/agent/loop.go` with a read from `al.sessionThinkLevel[opts.SessionKey]`; omit the key entirely when budget == 0.
- [ ] 2.3 **GREEN** — Run `TestThinkLevelMap`: `go test ./pkg/agent/ -run TestThinkLevelMap` — all assertions pass.
- [ ] 2.4 **RED** — Write `TestThinkCommandParsing` in `pkg/agent/loop_think_test.go`: use `MockProvider` to verify that sending `/think high` returns the ACK string with no LLM call, and that a subsequent message carries `thinking_budget_tokens=8192` in provider opts. Run — expect failure.
- [ ] 2.5 **GREEN** — Run `TestThinkCommandParsing` after 2.1–2.2: `go test ./pkg/agent/ -run TestThinkCommandParsing` — passes.
- [ ] 2.6 **RED** — Write `TestThinkInvalidLevel` in same file: send `/think ultra`, assert error message contains all valid level names, assert thinking budget unchanged. Run — expect failure.
- [ ] 2.7 **GREEN** — Run `TestThinkInvalidLevel`: `go test ./pkg/agent/ -run TestThinkInvalidLevel` — passes.
- [ ] 2.8 Commit: `feat(agent): add /think command with session-scoped thinking budget`

## Phase 3: `llm_task` Tool

- [ ] 3.1 **RED** — Create `pkg/tools/llm_task_test.go`: `TestLLMTaskHappyPath` uses a `MockProvider` (canned response `"result text"`); asserts `Execute` returns `"result text"` and no error. `TestLLMTaskTimeout` uses a provider that sleeps 35s; asserts error contains `"timeout"`. Run — compile error expected.
- [ ] 3.2 Create `pkg/tools/llm_task.go`:

```go
package tools

import (
    "context"
    "fmt"
    "time"

    "github.com/sipeed/makoclaw/pkg/config"
    "github.com/sipeed/makoclaw/pkg/providers"
)

type LLMTaskTool struct {
    cfg *config.Config
}

func NewLLMTaskTool(cfg *config.Config) *LLMTaskTool {
    return &LLMTaskTool{cfg: cfg}
}

func (t *LLMTaskTool) Name() string { return "llm_task" }

func (t *LLMTaskTool) Description() string {
    return "Delegate a sub-task to a specific LLM model. No tools, single turn, text response only."
}

func (t *LLMTaskTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "model": map[string]interface{}{
                "type":        "string",
                "description": "Provider/model string, e.g. openai/gpt-4o-mini",
            },
            "task": map[string]interface{}{
                "type":        "string",
                "description": "The task or question to send to the model",
            },
            "context": map[string]interface{}{
                "type":        "string",
                "description": "Optional background context to prepend",
            },
        },
        "required": []string{"model", "task"},
    }
}

func (t *LLMTaskTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    model, _ := args["model"].(string)
    task, _ := args["task"].(string)
    contextStr, _ := args["context"].(string)

    if model == "" || task == "" {
        return "", fmt.Errorf("llm_task: model and task are required")
    }

    clonedCfg := *t.cfg
    clonedCfg.Agents.Defaults.Model = model

    provider, err := providers.CreateProvider(&clonedCfg)
    if err != nil {
        return "", fmt.Errorf("llm_task: cannot create provider for model %q: %w", model, err)
    }

    var messages []providers.Message
    if contextStr != "" {
        messages = append(messages, providers.Message{Role: "system", Content: contextStr})
    }
    messages = append(messages, providers.Message{Role: "user", Content: task})

    ctx30s, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    resp, err := provider.Chat(ctx30s, messages, nil, model, map[string]interface{}{"max_tokens": 4096})
    if err != nil {
        if ctx30s.Err() == context.DeadlineExceeded {
            return "", fmt.Errorf("llm_task timeout after 30s")
        }
        return "", fmt.Errorf("llm_task: provider error: %w", err)
    }

    return resp.Content, nil
}
```

- [ ] 3.3 **GREEN** — Run `go test ./pkg/tools/ -run TestLLMTask` — both tests pass.
- [ ] 3.4 In `pkg/agent/loop.go` inside `NewAgentLoop`, register `tools.NewLLMTaskTool(cfg)` (same pattern as other tools). Verify `al.cfg` is accessible; if not, pass `cfg` through to where tools are registered.
- [ ] 3.5 **RED** — Add `TestLLMTaskRegistered` in `pkg/agent/loop_think_test.go` (or a new `loop_llmtask_test.go`): construct a real `AgentLoop` with minimal config and assert `al.tools.Get("llm_task") != nil`. Run — expect failure.
- [ ] 3.6 **GREEN** — Run `TestLLMTaskRegistered` after 3.4: `go test ./pkg/agent/ -run TestLLMTaskRegistered` — passes.
- [ ] 3.7 Commit: `feat(tools): add llm_task tool for single-turn model delegation`

## Phase 4: Compaction Memory Flush

- [ ] 4.1 **RED** — Write `TestFlushCallbackInvoked` in `pkg/session/manager_test.go`: create a `SessionManager`, call `SetFlushCallback` with a spy that records calls, add >keepLast messages, call `TruncateHistoryForUser`, assert spy was called once and history was truncated. Run — expect compile failure (field not yet defined).
- [ ] 4.2 In `pkg/session/manager.go`, update `TruncateHistoryForUser` to: acquire a 30s-bounded context, call `sm.flushCallback(ctx30s, key)` if set and `len(session.Messages) > keepLast`, then proceed with truncation regardless of flush error or timeout. Import `"context"` and `"time"`.
- [ ] 4.3 **GREEN** — Run `TestFlushCallbackInvoked`: `go test ./pkg/session/ -run TestFlushCallbackInvoked` — passes.
- [ ] 4.4 **RED** — Write `TestFlushSkippedWhenNil` in `pkg/session/manager_test.go`: call `TruncateHistoryForUser` without setting a callback; assert no panic and history is truncated. Run — expect failure.
- [ ] 4.5 **GREEN** — Run `TestFlushSkippedWhenNil` — passes.
- [ ] 4.6 **RED** — Write `TestFlushTimeout` in `pkg/session/manager_test.go`: callback sleeps 35s; assert truncation completes within 5s of the 30s deadline (use a 35s overall test timeout). Run — expect test to hang (failure).
- [ ] 4.7 **GREEN** — Run `TestFlushTimeout` after 4.2 implements the 30s-bounded context — passes within deadline.
- [ ] 4.8 In `pkg/agent/loop.go` inside `NewAgentLoop` (or `NewAgentLoopForUser`), after `sessions` is initialized, call `al.sessions.SetFlushCallback(func(ctx context.Context, sessionKey string) error { ... })`. The callback: check if `query_knowledge` is registered in `al.tools`; if not, return nil silently. Otherwise inject a synthetic user message `"Before we continue, save any important facts, decisions, and in-progress work to the knowledge base."`, run a single non-streaming LLM iteration against that session (do not append to history), return nil.
- [ ] 4.9 Verify CF-06: the flush synthetic turn must not appear in post-truncation history. Add `TestFlushTurnNotInHistory` in `pkg/session/manager_test.go`: after truncation, assert no message with the flush prompt text exists in `session.Messages`.
- [ ] 4.10 Commit: `feat(session): inject memory flush before compaction truncation`

## Phase 5: Integration and Cleanup

- [ ] 5.1 Run full test suite: `go test ./pkg/agent/... ./pkg/tools/... ./pkg/session/...` — all pass, no races: `go test -race ./pkg/agent/... ./pkg/tools/... ./pkg/session/...`
- [ ] 5.2 Remove any `// TODO` or placeholder comments introduced during implementation; ensure all new exported symbols have godoc comments.
- [ ] 5.3 Run `go vet ./pkg/agent/... ./pkg/tools/... ./pkg/session/...` — zero warnings.
- [ ] 5.4 Commit: `chore(agent): clean up phase-8.2 implementation`
