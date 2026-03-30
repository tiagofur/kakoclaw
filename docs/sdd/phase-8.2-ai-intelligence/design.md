# Design: Phase 8.2 — AI Intelligence

## Technical Approach

Three independent, additive changes to `pkg/agent/loop.go`, a new `pkg/tools/llm_task.go`, and a pre-truncation hook in `pkg/session/manager.go`. No DB migrations. All changes are reversible in isolation.

---

## Architecture Decisions

| Decision | Choice | Rejected | Rationale |
|---|---|---|---|
| ThinkLevel storage | In-memory field on `processOptions` + per-session map in `AgentLoop` | Config field, session JSON | Session-only by design; no persistence overhead |
| `/think` parsing | Pre-LLM in `runAgentLoop` / `runAgentLoopStream`, return early with ACK | Middleware / command handler | Both loop entry points share the parse; no new abstraction needed |
| `llm_task` provider resolution | Accept `model` string, call `CreateProvider` with cloned cfg | Inject provider at registration time | Allows `provider/model` syntax already supported by `CreateProvider` |
| Pre-truncation flush trigger | Wrap `TruncateHistoryForUser` — inject synthetic user turn, run one non-streaming LLM call, then truncate | Separate goroutine, post-truncation | Must complete before history is lost; blocking is acceptable |
| Flush guard | Check that `query_knowledge` is registered before injecting flush turn | Always inject | Avoids noisy no-op turn when knowledge base unavailable |

---

## Data Flow

### /think command

```
User input "/think high"
  └─ runAgentLoop / runAgentLoopStream
       ├─ strings.HasPrefix(opts.UserMessage, "/think")
       ├─ parse level → ThinkLevel(8192)
       ├─ store al.sessionThinkLevel[sessionKey] = 8192
       ├─ return ACK string (no LLM call)
       └─ next LLM call reads al.sessionThinkLevel[sessionKey]
            └─ llmOpts["thinking_budget_tokens"] = budget (0 = omit key)
```

### llm_task tool

```
Agent calls llm_task(model, task, context)
  └─ LLMTaskTool.Execute()
       ├─ clone al.cfg, override Model = args["model"]
       ├─ providers.CreateProvider(clonedCfg) → secondary provider
       ├─ providers.Chat(ctx30s, []Message{system+user}, nil, model, opts)
       └─ return response text (or error string)
```

### Pre-truncation memory flush

```
SessionManager.TruncateHistoryForUser(userID, key, keepLast)
  └─ if flushCallback != nil AND len(messages) > keepLast
       ├─ flushCallback(ctx, sessionKey)  // 30s timeout
       │    └─ AgentLoop: inject synthetic user turn
       │         "Before we continue, save important facts to knowledge base."
       │         runLLMIteration (non-streaming, no history append)
       └─ proceed with truncation
```

---

## Type and Struct Definitions

```go
// pkg/agent/loop.go

// ThinkLevel maps human-readable names to token budgets.
type ThinkLevel int

const (
    ThinkOff     ThinkLevel = 0
    ThinkMinimal ThinkLevel = 512
    ThinkLow     ThinkLevel = 1024
    ThinkMedium  ThinkLevel = 4096
    ThinkHigh    ThinkLevel = 8192
    ThinkXHigh   ThinkLevel = 16000
)

var thinkLevelMap = map[string]ThinkLevel{
    "off":     ThinkOff,
    "minimal": ThinkMinimal,
    "low":     ThinkLow,
    "medium":  ThinkMedium,
    "high":    ThinkHigh,
    "xhigh":   ThinkXHigh,
}

// Added to AgentLoop struct:
//   sessionThinkLevel map[string]ThinkLevel  // key = sessionKey
//   sessionThinkMu    sync.RWMutex
```

```go
// pkg/tools/llm_task.go

type LLMTaskTool struct {
    cfg *config.Config
}

func NewLLMTaskTool(cfg *config.Config) *LLMTaskTool

func (t *LLMTaskTool) Name() string        { return "llm_task" }
func (t *LLMTaskTool) Description() string { ... }

func (t *LLMTaskTool) Parameters() map[string]interface{} {
    // Required: "task" (string)
    // Required: "model" (string, provider/model format)
    // Optional: "context" (string)
}

func (t *LLMTaskTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    // 1. Clone cfg, set cfg.Agents.Defaults.Model = args["model"]
    // 2. providers.CreateProvider(clonedCfg)
    // 3. ctx30s, cancel := context.WithTimeout(ctx, 30*time.Second)
    // 4. provider.Chat(ctx30s, messages, nil, model, map{"max_tokens": 4096})
    // 5. return response.Content, nil
}
```

```go
// pkg/session/manager.go — flush callback injection

// FlushCallback is called before truncation with a 30s-bounded context.
type FlushCallback func(ctx context.Context, sessionKey string) error

// SessionManager gains one new field:
//   flushCallback FlushCallback

func (sm *SessionManager) SetFlushCallback(fn FlushCallback)
```

---

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/agent/loop.go` | Modify | Add `sessionThinkLevel` map + mutex; parse `/think` pre-LLM; wire budget into `llmOpts`; register `LLMTaskTool`; set flush callback on `SessionManager` |
| `pkg/tools/llm_task.go` | Create | `LLMTaskTool` implementation |
| `pkg/session/manager.go` | Modify | Add `FlushCallback` field + `SetFlushCallback`; invoke in `TruncateHistoryForUser` |

---

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | `thinkLevelMap` parse, budget injection into `llmOpts` | Table-driven test in `loop_test.go` using `MockProvider` |
| Unit | `LLMTaskTool.Execute` happy path + timeout | Mock provider; assert text returned within 30s |
| Unit | `TruncateHistoryForUser` calls `flushCallback` | Inject spy callback; assert called once before truncation |
| Integration | `/think high` → next LLM call carries `thinking_budget_tokens=8192` | Existing `shell_test.go` pattern with `MockProvider` |

---

## Migration / Rollout

No migration required. All three features are additive and independently gated:
- Thinking level: only active when `/think` is issued (default remains `off` = no budget injected unless `OnThinking != nil`)
- `llm_task`: only callable if agent uses the tool; no side effects on registration
- Flush: only fires if `flushCallback` is set AND `query_knowledge` is registered

---

## Open Questions

- [ ] Should `/think` persist across sessions (e.g., saved to session JSON)? Proposal says no — confirm with team.
- [ ] Cap `xhigh=16000` to provider limit dynamically, or hard-cap and document? Current plan: hard-cap + log warning if provider rejects.
