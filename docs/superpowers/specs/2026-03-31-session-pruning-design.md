# Spec: Session Pruning (session-pruning)

**Date:** 2026-03-31
**Status:** Approved
**Change:** session-pruning

---

## 1. Context & Problem Statement

The MakoClaw agent loop re-sends the full session history to the LLM on every call inside `runLLMIteration`. Tool result messages (role=`"tool"`) accumulate in that history and are included verbatim every time. A single large tool result — for example a `web_fetch` returning 10,000 characters of HTML, or a `list_dir` over a large directory — is never discarded, so it inflates token usage on every subsequent iteration for the lifetime of the session.

The consequences are:

- Unnecessary token spend on every LLM call after the first large tool result.
- Premature session summarization (the auto-summarizer fires when history exceeds the configured token/message threshold).
- Increased latency as context size grows.

The fix is a lightweight, in-memory-only truncation pass applied to the history slice **just before** each `Chat()` call. Older tool results are compressed to a configurable maximum length; recent tool results are left untouched so the LLM can still reason about its latest actions.

---

## 2. Functional Requirements

**FR-1.** The system MUST provide a pure function `pruneHistoryToolResults` that accepts a slice of `providers.Message`, a `keepRecentN` integer, and a `maxChars` integer, and returns a new slice of `providers.Message`.

**FR-2.** The function MUST leave all non-tool messages (`role != "tool"`) completely unmodified.

**FR-3.** The function MUST leave the `keepRecentN` most-recently-encountered tool result messages (role=`"tool"`) completely unmodified.

**FR-4.** For every tool result message that is older than the `keepRecentN` most recent, the function MUST truncate the `Content` field to at most `maxChars` characters if its current length exceeds `maxChars`.

**FR-5.** When a tool result is truncated, the function MUST append the literal string `"\n[truncated]"` to the truncated content.

**FR-6.** When `maxChars` is `0`, the function MUST return the input slice unchanged (feature disabled). No copy is made.

**FR-7.** When `keepRecentN` is `0`, ALL tool result messages are candidates for truncation (none are protected by recency).

**FR-8.** The function MUST NOT mutate the original input slice or any of the `Message` structs it contains. It MUST return a new slice with new `Message` copies where modifications are needed.

**FR-9.** The function MUST be called from `runLLMIteration` (in `pkg/agent/loop.go`) on the `messages` slice, immediately before the `al.provider.Chat()` call. The result of the function MUST be passed to `Chat()`. The original `messages` variable MUST NOT be overwritten.

**FR-10.** The same pruning MUST also be applied inside `runLLMIterationStream` for the streaming code path, using the identical function.

**FR-11.** The two configuration values (`ToolResultMaxChars`, `ToolResultKeepRecent`) MUST be read from `AgentDefaults` in the agent config and passed into the iteration function at call time.

---

## 3. Non-Functional Requirements

**NFR-1. Performance.** The function iterates the message slice once in reverse to count tool results, then once forward to build the output. Time complexity is O(n) where n is the number of messages. No external I/O, no allocations beyond the new slice.

**NFR-2. Safety — no storage mutation.** Session history files on disk MUST NOT be modified. The pruned slice is ephemeral and exists only for the duration of the LLM call.

**NFR-3. Backward compatibility.** Existing configs without the new fields will receive the default values (`ToolResultMaxChars=2000`, `ToolResultKeepRecent=3`). No migration is required. Setting `ToolResultMaxChars=0` restores the pre-feature behavior.

**NFR-4. No streaming divergence.** The streaming path (`runLLMIterationStream`) MUST apply the same pruning so token behaviour is identical regardless of streaming mode.

**NFR-5. Testability.** The function MUST be a package-level (unexported) function in `pkg/agent/loop.go` or a dedicated file, making it directly unit-testable without instantiating an `AgentLoop`.

---

## 4. Interfaces & Types

### 4.1 Pure function signature

```go
// pruneHistoryToolResults returns a copy of messages where tool-result entries
// older than the keepRecentN most recent are truncated to maxChars characters.
// If maxChars == 0 the original slice is returned unchanged.
// The original slice is never mutated.
func pruneHistoryToolResults(
    messages    []providers.Message,
    keepRecentN int,
    maxChars    int,
) []providers.Message
```

`providers.Message` is defined in `pkg/providers/types.go`:

```go
type Message struct {
    Role       string     `json:"role"`
    Content    string     `json:"content"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
    ToolCallID string     `json:"tool_call_id,omitempty"`
}
```

### 4.2 Config fields added to `AgentDefaults`

```go
type AgentDefaults struct {
    // ... existing fields ...
    ToolResultMaxChars   int `json:"tool_result_max_chars"`   // 0 = disabled; default 2000
    ToolResultKeepRecent int `json:"tool_result_keep_recent"` // default 3
}
```

---

## 5. Behavior Specification

### 5.1 Disabled mode (`maxChars == 0`)

Return the input slice as-is. Do not allocate a new slice.

### 5.2 Message scan order

Tool results are identified by scanning the message slice **from the end** (newest first). The first `keepRecentN` tool result messages encountered are marked as protected. All subsequent tool result messages found while scanning are candidates for truncation.

Because the scan is newest-first, "recent" means closest to the end of the slice (the most recent messages in conversational order).

### 5.3 Truncation rule

For a candidate tool result message with content `c`:

```
if len(c) > maxChars:
    new content = c[:maxChars] + "\n[truncated]"
else:
    message is NOT modified (even though it is a candidate)
```

A message that is a candidate but whose content is already within `maxChars` is copied without modification.

### 5.4 Non-tool messages

Messages with `role != "tool"` are NEVER modified regardless of `Content` length or position.

### 5.5 Return value

The function returns a new `[]providers.Message` of the same length as the input. Individual `Message` structs that require no change are copied by value from the input slice. Structs that are truncated produce a new `Message` value with the modified `Content` field; all other fields (`Role`, `ToolCalls`, `ToolCallID`) are copied unchanged.

### 5.6 `keepRecentN == 0`

When `keepRecentN` is zero, no tool results are protected. Every tool result in the slice is a candidate (subject to the length check in 5.3).

### 5.7 Tool results with empty content

An empty `Content` string has length 0, which is ≤ any positive `maxChars`. It passes through unmodified.

### 5.8 Non-string content (future-proofing)

`providers.Message.Content` is always a `string` in the current type definition. No special handling for other types is needed. If future schema changes introduce structured content, this function will need revisiting, but that is out of scope for this change.

---

## 6. Integration Point

### 6.1 Location in `runLLMIteration`

In `pkg/agent/loop.go`, function `runLLMIteration` (starting at line ~1539), the following insertion is made immediately before the `al.provider.Chat()` call (currently line ~1646):

```go
// Prune old tool results to limit context size (in-memory only, not persisted)
prunedMessages := pruneHistoryToolResults(
    messages,
    al.cfg.Agents.Defaults.ToolResultKeepRecent,
    al.cfg.Agents.Defaults.ToolResultMaxChars,
)

response, err := al.provider.Chat(ctx, prunedMessages, providerToolDefs, model, map[string]interface{}{
    "max_tokens":  8192,
    "temperature": 0.7,
})
```

`messages` (the variable populated by `BuildMessages`) is NOT overwritten. `prunedMessages` is a local variable scoped to this iteration. All logging that reads `messages` (e.g., `formatMessagesForLog`) continues to log the original unpruned messages for debugging accuracy.

### 6.2 Location in `runLLMIterationStream`

The same pattern applies in `runLLMIterationStream`. The pruning call is inserted in the identical position before `al.provider.ChatStream()`.

### 6.3 Config access

`AgentLoop` already holds a reference to `al.cfg` (the loaded `*config.Config`). The two new fields are accessed as:

```go
al.cfg.Agents.Defaults.ToolResultMaxChars
al.cfg.Agents.Defaults.ToolResultKeepRecent
```

---

## 7. Configuration

### 7.1 JSON example

```json
{
  "agents": {
    "defaults": {
      "tool_result_max_chars": 2000,
      "tool_result_keep_recent": 3
    }
  }
}
```

### 7.2 Default values (in `DefaultConfig()`)

| Field | Default | Meaning |
|-------|---------|---------|
| `tool_result_max_chars` | `2000` | Older tool results truncated to 2000 chars |
| `tool_result_keep_recent` | `3` | The 3 most recent tool results are never truncated |

### 7.3 Disabling the feature

Set `tool_result_max_chars` to `0`. The function returns the original slice unchanged.

### 7.4 Truncating all tool results

Set `tool_result_keep_recent` to `0`. All tool results are candidates (still subject to the per-message length check).

---

## 8. Test Scenarios

All tests live in `pkg/agent/session_pruning_test.go`. They call `pruneHistoryToolResults` directly.

| ID | Name | Setup | Expected |
|----|------|-------|----------|
| T1 | Disabled (maxChars=0) | History with large tool results; maxChars=0 | Returns exact same slice (pointer equality) |
| T2 | No tool results | History of user + assistant messages only; maxChars=500 | Returns slice equal to input; no modification |
| T3 | All tool results within maxChars | 5 tool results each 100 chars; maxChars=200 | All tool results preserved verbatim |
| T4 | Old tool result exceeds maxChars | 1 old tool result of 500 chars; keepRecentN=0; maxChars=100 | Content truncated to 100 chars + `"\n[truncated]"` |
| T5 | Recent N preserved, older truncated | 5 tool results: indices 0–2 old (600 chars each), 3–4 recent; keepRecentN=2; maxChars=100 | Indices 3–4 untouched; indices 0–2 truncated to 100+suffix |
| T6 | keepRecentN=0 truncates all | 3 tool results of 300 chars each; keepRecentN=0; maxChars=50 | All 3 truncated |
| T7 | Non-tool messages never modified | User message of 5000 chars; assistant message of 5000 chars; maxChars=100 | Both returned unmodified |
| T8 | Original slice not mutated | Build input slice; call function; inspect original slice | Original Message structs are unchanged |

---

## 9. Files to Modify

| File | Change |
|------|--------|
| `pkg/config/config.go` | Add `ToolResultMaxChars int` and `ToolResultKeepRecent int` to `AgentDefaults` struct. Add defaults `ToolResultMaxChars: 2000` and `ToolResultKeepRecent: 3` in `DefaultConfig()`. |
| `pkg/agent/loop.go` | Add `pruneHistoryToolResults` function. Call it in `runLLMIteration` before `al.provider.Chat()`. Call it in `runLLMIterationStream` before `al.provider.ChatStream()`. |
| `pkg/agent/session_pruning_test.go` | New file. Unit tests T1–T8 for `pruneHistoryToolResults`. |
