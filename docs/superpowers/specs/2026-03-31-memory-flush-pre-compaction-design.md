# Memory Flush Before Compaction — Design Spec

**Date:** 2026-03-31
**Status:** Approved
**Feature:** Automatic memory flush before session compaction

---

## Context

When a MakoClaw session exceeds the token/message threshold, it auto-summarizes (compacts) the history. This can silently discard important context — decisions made, facts learned, user preferences established during the session. OpenClaw solves this with a "silent turn" before compaction where the agent saves important context to memory. We're implementing the same pattern.

---

## Architecture

### Current flow (`pkg/agent/loop.go`)

```
user message
  → BuildContext()
  → LLM call
  → execute tools
  → save to session
  → [session auto-summarizes if over threshold]
  → repeat
```

### New flow

```
user message
  → BuildContext()
  → LLM call
  → execute tools
  → save to session
  → is session near threshold AND flush enabled AND !memoryFlushDone?
      → yes: run silent flush turn (LLM call + memory tools only)
      → set memoryFlushDone = true
  → session auto-summarizes
  → reset memoryFlushDone = false
  → repeat
```

The flush turn is a separate, silent LLM call. It does not get sent to the user channel, does not modify session history, and only writes to memory files.

---

## Configuration

New field in `AgentDefaults` in `pkg/config/config.go`:

```go
type AgentDefaults struct {
    // existing fields...
    MemoryFlushBeforeCompaction bool `json:"memory_flush_before_compaction"`
}
```

Default value: `true` (active by default — user opts out, not in).

Can be overridden per specialist agent. No separate threshold config — reuses the existing session summarization threshold.

Example `config.json`:

```json
{
  "agents": {
    "defaults": {
      "memory_flush_before_compaction": true
    }
  }
}
```

---

## The Flush Turn

### System prompt

Injected as a standalone system message, does NOT modify session history:

```
This session is about to be summarized. Before that happens, review
the conversation and save anything important to memory — key decisions,
facts, user preferences, ongoing tasks, or context that could be lost
in summarization. Use the available memory tools to persist this.
```

### Available tools (memory-only subset)

The full agent toolset is filtered down to memory tools only:

- `write_file` — restricted to memory paths (`memory/MEMORY.md`, `memory/YYYY-MM/YYYYMMDD.md`)
- `append_file` — same path restriction
- `query_knowledge` + knowledge base tools (if available in the agent's toolset)

Explicitly excluded: `exec`, `web_search`, `web_fetch`, `browser`, `message`, `spawn`, and all non-memory tools.

### Token budget

- Max response tokens: `2000` (hardcoded constant `flushMaxTokens`)
- If the model calls tools, they are executed
- If the model calls no tools, the flush is silently skipped

### Execution behavior

- LLM response is discarded (not shown to user, not saved to session)
- Tool calls ARE executed (this is the entire value — writes to disk)
- The flush turn is not added to session history

---

## State & Safety

### Preventing double flush

A `memoryFlushDone bool` field on the agent loop struct:

```go
// pseudo-logic in the loop
if session.NearThreshold() && cfg.MemoryFlushBeforeCompaction && !a.memoryFlushDone {
    a.runMemoryFlushTurn(ctx)
    a.memoryFlushDone = true
}

// after session summarizes:
a.memoryFlushDone = false
```

### Error handling

The flush is **best-effort** — it never blocks the summarization:

- Flush turn LLM call fails (timeout, provider error) → log warning, continue with compaction
- Agent calls no tools → silent skip, continue with compaction
- A memory tool fails to write → log error, continue with remaining tools and compaction

### Logging

```go
logger.InfoC("agent", "memory flush before compaction: starting")
logger.InfoC("agent", "memory flush before compaction: completed, tools executed: N")
logger.WarnC("agent", "memory flush before compaction: failed, proceeding with compaction")
```

---

## Session Interface

New method on the session manager to allow the loop to check threshold proximity without triggering summarization:

```go
// NearThreshold returns true if the session is approaching its compaction threshold
// (e.g., within 20% of the token limit or message count limit)
func (s *Session) NearThreshold() bool
```

---

## Testing

### Unit tests (`pkg/agent/`)

| Test | What it verifies |
|------|-----------------|
| `TestFlushTrigger` | Flush runs when `NearThreshold() = true` and config is enabled |
| `TestFlushDisabled` | No extra LLM call when `memory_flush_before_compaction: false` |
| `TestNoDoubleFlush` | Two loop iterations near threshold → flush runs only once |
| `TestFlushFlagReset` | After session summarizes, `memoryFlushDone` resets to `false` |
| `TestFlushErrorContinues` | Provider error during flush → compaction still proceeds |
| `TestFlushNoTools` | Agent calls no tools → flush silently skipped, compaction proceeds |

### Integration test

- Real session exceeding threshold with `memory_flush_before_compaction: true`
- Verify `MEMORY.md` was modified before session history is summarized

---

## Files to Modify

| File | Change |
|------|--------|
| `pkg/config/config.go` | Add `MemoryFlushBeforeCompaction bool` to `AgentDefaults` |
| `pkg/agent/loop.go` | Add flush turn logic + `memoryFlushDone` flag |
| `pkg/session/session.go` | Add `NearThreshold() bool` method |
| `pkg/agent/loop_test.go` | Add flush-related unit tests |
