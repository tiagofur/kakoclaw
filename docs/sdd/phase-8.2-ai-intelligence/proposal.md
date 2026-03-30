# Proposal: Phase 8.2 — AI Intelligence

**Change**: `phase-8.2-ai-intelligence`
**Status**: Draft
**Inspired by**: OpenClaw (https://github.com/openclaw/openclaw)
**Date**: 2026-03-30

---

## Intent

Three targeted AI capabilities to improve reasoning control, cost efficiency, and session continuity in MakoClaw:

1. Users can't control extended thinking budget per session — it's either always-on (hardcoded 1024 tokens) or off. No in-chat toggle exists.
2. No way for the agent to delegate a subtask to a cheaper/faster model — every tool call uses the primary model.
3. When context is summarized (session truncation), important in-flight facts are silently lost with no pre-flush to knowledge base.

## Scope

### In Scope
- `/think` chat command: `off | minimal | low | medium | high | xhigh` — sets thinking budget for the current session only
- `llm_task` tool: agent-callable tool that sends a prompt to an alternate model (no tools), returns text
- Context compaction memory flush: a pre-truncation hook that prompts the agent to save key facts to knowledge base before history is trimmed

### Out of Scope
- Global/persistent thinking level config (session-only by design)
- `llm_task` tool access to tools or structured JSON output (text-only, v1)
- Automatic fact extraction / NLP summarization (agent-driven, not rule-based)
- UI controls for thinking level (chat command only, v1)

## Approach

**Thinking Levels**: Extend `AgentLoopOptions` with `ThinkingBudget int` field. Parse `/think <level>` in the agent loop's message pre-processing (before LLM call). Map levels to token budgets: off=0, minimal=512, low=1024, medium=4096, high=8192, xhigh=16000. Store in session state (in-memory, not persisted). Replace the hardcoded `thinking_budget_tokens: 1024` in `loop.go` with the session value.

**llm_task Tool**: New `pkg/tools/llm_task.go` implementing the `Tool` interface. Takes `model` (provider/model string), `task`, `context` params. Calls the provider directly (no tools, no history) via `providers.CreateProvider`. Returns the text response. Register in `NewAgentLoop`.

**Context Compaction Memory Flush**: In `pkg/session/manager.go`, before truncating history, inject a synthetic user turn: `"Before we continue, please save any important facts, decisions, or in-progress work to your knowledge base using query_knowledge."` Run a single-turn agent loop iteration, then proceed with truncation.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/agent/loop.go` | Modified | Add `/think` command parsing, `ThinkingBudget` in opts, dynamic budget in LLM opts |
| `pkg/tools/llm_task.go` | New | `llm_task` tool implementation |
| `pkg/agent/loop.go` | Modified | Register `llm_task` in `NewAgentLoop` |
| `pkg/session/manager.go` | Modified | Pre-truncation memory flush hook |
| `pkg/config/config.go` | Modified | Optional `ThinkingLevels` budget map config |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| xhigh budget causes provider errors (Claude supports up to 16K) | Med | Cap at provider limit; graceful downgrade to `high` |
| `llm_task` creates infinite loops | Low | llm_task makes a single non-streaming call, no tool access |
| Pre-truncation flush increases latency for every compaction | Med | Cap flush turn to 30s timeout; skip if `query_knowledge` tool not registered |
| Session think level lost on restart | Low | Documented as intentional — session-only by design |

## Rollback Plan

- Thinking levels: revert `loop.go` to hardcoded `thinking_budget_tokens: 1024`
- `llm_task`: unregister from `NewAgentLoop`, delete `pkg/tools/llm_task.go`
- Compaction flush: remove pre-truncation hook from `manager.go`
- All three features are independently reversible; no DB migrations required

## Dependencies

- `llm_task` requires a valid secondary provider config (or uses primary with different model)
- Compaction flush requires `query_knowledge` tool to be registered (gracefully skipped otherwise)
- Extended thinking budget requires Claude provider (other providers silently ignore the option)

## Success Criteria

- [ ] `/think high` in chat sets budget to 8192 tokens; thinking blocks appear in responses
- [ ] `/think off` disables thinking blocks for that session
- [ ] `llm_task` tool callable from agent, returns text from alternate model within 30s
- [ ] Knowledge base contains agent-saved facts after a session truncation event
- [ ] No regressions in existing tests (`go test ./...`)

## Next Steps

- `sdd-spec` and `sdd-design` can run in parallel (both depend only on this proposal)
