# Design: Session Pruning (session-pruning)

**Date:** 2026-03-31
**Status:** Approved
**Change:** session-pruning
**Spec:** `docs/superpowers/specs/2026-03-31-session-pruning-design.md`

---

## 1. Architecture Overview

Session pruning is a lightweight, purely transformational pre-LLM-call filter inserted at the boundary between context assembly and the provider API call inside the agent loop. It operates entirely in-memory on an ephemeral copy of the message slice, has no side effects on stored session history, and requires no new packages, interfaces, or types beyond two configuration fields. The function is placed in `pkg/agent/loop.go` as a package-level unexported helper — co-located with the two call sites that use it — and is directly unit-testable without an `AgentLoop` instance.

---

## 2. Component: `pruneHistoryToolResults`

### Location

Package-level unexported function in `pkg/agent/loop.go`, appended at the bottom of the file after `runMemoryFlushTurn` (currently the last function, ending around line 2700).

### Signature

```go
// pruneHistoryToolResults returns a copy of messages where tool-result entries
// older than the keepRecentN most recent are truncated to maxChars characters.
// If maxChars == 0 the original slice is returned unchanged (no allocation).
// The original slice and its Message structs are never mutated.
func pruneHistoryToolResults(
    messages    []providers.Message,
    keepRecentN int,
    maxChars    int,
) []providers.Message
```

### Algorithm

```
1. if maxChars == 0:
       return messages  // early exit, no alloc, feature disabled

2. Scan messages from end to start (index len-1 down to 0).
   Maintain recentCount := 0 and a candidateSet map[int]struct{}.
   For each message where Role == "tool":
       if recentCount < keepRecentN:
           recentCount++           // this is a protected "recent" result
       else:
           candidateSet[i] = {}    // this is a truncation candidate

3. output := make([]providers.Message, len(messages))
   copy(output, messages)          // shallow copy of all Message structs

4. For each index i in candidateSet:
       if len(output[i].Content) > maxChars:
           output[i].Content = output[i].Content[:maxChars] + "\n[truncated]"
   // (If content <= maxChars, the copy from step 3 is already correct — no change needed.)

5. return output
```

### Safety

Step 3 uses `copy()` which copies struct values (Go copy semantics for slice elements). In step 4, we assign to `output[i].Content` — this modifies the copied struct in `output`, not the original struct in `messages`. The `Role`, `ToolCalls`, and `ToolCallID` fields are untouched.

---

## 3. Integration in `runLLMIteration`

The insertion point is in `pkg/agent/loop.go`, immediately before the primary `al.provider.Chat()` call at line 1646 (inside the iteration loop, after all debug logging of the original `messages` variable).

```go
// BEFORE (existing code at line 1644-1649):
// Call LLM
llmStart := time.Now()
response, err := al.provider.Chat(ctx, messages, providerToolDefs, model, map[string]interface{}{
    "max_tokens":  8192,
    "temperature": 0.7,
})

// AFTER:
// Prune old tool results to limit context size (in-memory only, not persisted)
prunedMessages := pruneHistoryToolResults(
    messages,
    al.cfg.Agents.Defaults.ToolResultKeepRecent,
    al.cfg.Agents.Defaults.ToolResultMaxChars,
)

// Call LLM
llmStart := time.Now()
response, err := al.provider.Chat(ctx, prunedMessages, providerToolDefs, model, map[string]interface{}{
    "max_tokens":  8192,
    "temperature": 0.7,
})
```

The `messages` variable is NOT overwritten. All logging that references `messages` (e.g., `formatMessagesForLog(messages)` at line 1640) continues to log the original, unpruned history — this is intentional for debugging accuracy. The `tokensIn` estimate at line 1651 (`al.estimateTokens(messages)`) also remains on the unpruned slice; a follow-up decision could change this to `prunedMessages` if token accounting accuracy is desired, but that is out of scope.

There is also a secondary `al.provider.Chat()` call in `runLLMIteration` at line 1886 (the "max iterations reached, force text" concluding call). That call already operates on a mutated `messages` slice that has an injected system instruction appended — it does not benefit from pruning and MUST NOT be wrapped (applying pruning to the concluding call would be premature; the agent is already being told to stop).

---

## 4. Integration in `runLLMIterationStream`

The `runLLMIterationStream` function has two call sites to cover:

**4a. Streaming path** (`streamingProvider.ChatStream()` at line 2020):

```go
// BEFORE (existing code at line 2017-2020):
// Try streaming for this iteration
if canStream {
    llmStart := time.Now()
    ch, err := streamingProvider.ChatStream(ctx, messages, providerToolDefs, model, llmOpts)

// AFTER:
// Try streaming for this iteration
if canStream {
    // Prune old tool results to limit context size (in-memory only, not persisted)
    prunedMessages := pruneHistoryToolResults(
        messages,
        al.cfg.Agents.Defaults.ToolResultKeepRecent,
        al.cfg.Agents.Defaults.ToolResultMaxChars,
    )
    llmStart := time.Now()
    ch, err := streamingProvider.ChatStream(ctx, prunedMessages, providerToolDefs, model, llmOpts)
```

**4b. Fallback non-streaming path** (`al.provider.Chat()` at line 2240):

```go
// BEFORE (existing code at line 2238-2240):
// Fallback: non-streaming Chat()
fallbackStart := time.Now()
response, err := al.provider.Chat(ctx, messages, providerToolDefs, model, llmOpts)

// AFTER:
// Fallback: non-streaming Chat()
// Prune old tool results to limit context size (in-memory only, not persisted)
prunedMessages := pruneHistoryToolResults(
    messages,
    al.cfg.Agents.Defaults.ToolResultKeepRecent,
    al.cfg.Agents.Defaults.ToolResultMaxChars,
)
fallbackStart := time.Now()
response, err := al.provider.Chat(ctx, prunedMessages, providerToolDefs, model, llmOpts)
```

Note: the variable name `prunedMessages` is scoped to the `if canStream` block and to the fallback block respectively — no name collision occurs. The `messages` variable continues to be used for `al.estimateTokens(messages)` calls after each path, preserving consistent (unpruned) token estimates for metrics.

The concluding Chat call in `runLLMIterationStream` at line 2401 (max iterations reached) is exempt for the same reason as in section 3.

---

## 5. Config Changes

### 5a. `AgentDefaults` struct (`pkg/config/config.go`, line 184)

```go
// BEFORE:
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

// AFTER (two fields appended):
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
    ToolResultMaxChars          int     `json:"tool_result_max_chars"`   // 0 = disabled; default 2000
    ToolResultKeepRecent        int     `json:"tool_result_keep_recent"` // default 3
}
```

### 5b. `DefaultConfig()` (`pkg/config/config.go`, line 598)

```go
// BEFORE:
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

// AFTER:
Defaults: AgentDefaults{
    Workspace:                   "~/.MakoClaw/workspace",
    RestrictToWorkspace:         true,
    Provider:                    "",
    Model:                       "",
    MaxTokens:                   8192,
    Temperature:                 0.7,
    MaxToolIterations:           20,
    MemoryFlushBeforeCompaction: true,
    ToolResultMaxChars:          2000,
    ToolResultKeepRecent:        3,
},
```

Existing configs that omit these fields will deserialize `ToolResultMaxChars` and `ToolResultKeepRecent` as `0` (Go zero value for `int`). However, because `ToolResultMaxChars == 0` is the "disabled" sentinel, this would silently disable pruning for users who don't have the fields in their config. To avoid that, the `MergeConfigs` function (or whichever function merges global defaults into user configs) must apply the default values when the user config has zero values for these fields. If `MergeConfigs` already handles this for `int` fields by checking `== 0`, no additional work is needed. If not, the implementer should verify and handle accordingly.

---

## 6. Test File Structure

File: `pkg/agent/session_pruning_test.go`

```go
package agent

import (
    "strings"
    "testing"

    "github.com/sipeed/makoclaw/pkg/providers"
)

// helper: build a providers.Message with role and content
func makeMsg(role, content string) providers.Message { ... }

// T1: maxChars=0 → returns exact same slice (pointer equality via &slice[0])
func TestPruneHistoryToolResults_Disabled(t *testing.T) { ... }

// T2: no tool messages → slice returned equal to input, nothing modified
func TestPruneHistoryToolResults_NoToolMessages(t *testing.T) { ... }

// T3: all tool results within maxChars → none truncated
func TestPruneHistoryToolResults_AllWithinLimit(t *testing.T) { ... }

// T4: keepRecentN=0, one tool result exceeds maxChars → truncated with suffix
func TestPruneHistoryToolResults_TruncatesOldResult(t *testing.T) { ... }

// T5: keepRecentN=2, 5 tool results (indices 0-2 old, 3-4 recent) → 0-2 truncated, 3-4 untouched
func TestPruneHistoryToolResults_KeepsRecentProtected(t *testing.T) { ... }

// T6: keepRecentN=0 → all tool results truncated
func TestPruneHistoryToolResults_ZeroKeepRecentTruncatesAll(t *testing.T) { ... }

// T7: non-tool messages (user, assistant) never modified even if > maxChars
func TestPruneHistoryToolResults_NonToolMessagesUnmodified(t *testing.T) { ... }

// T8: original input slice and Message structs are unchanged after call
func TestPruneHistoryToolResults_OriginalNotMutated(t *testing.T) { ... }
```

All 8 test functions call `pruneHistoryToolResults` directly. No `AgentLoop` instantiation is required. The `strings` package is used in tests that verify truncation suffix (`strings.HasSuffix(got, "\n[truncated]")`).

---

## 7. Design Decisions & Rationale

| Decision | Choice | Rationale |
|----------|--------|-----------|
| File placement | `loop.go` (not a new file) | The function is a pre-call filter tightly coupled to both `runLLMIteration` and `runLLMIterationStream`; it warrants no independent abstraction layer |
| Scan direction | Reverse (end → start) | "Recent" means conversationally latest, which is at the tail of the slice |
| Output strategy | Copy-on-write (new slice + modified struct values) | Preserves original for logging and debug output; safe for any concurrent reader of the session history |
| Feature toggle | `maxChars == 0` → disabled | Zero value is the natural disabled state in Go; no extra boolean field required |
| Default `maxChars` | `2000` | Accommodates typical tool outputs (file reads, search results) while preventing multi-KB accumulation across iterations |
| Default `keepRecentN` | `3` | Keeps the last 3 tool exchanges visible to the LLM, covering most multi-step tool patterns without exposing the full history |
| Concluding call exemption | Not pruned | The forced-text concluding call appends a system instruction to `messages` — pruning would be redundant and its truncation could remove context the LLM needs to produce its final summary |
| Token estimate unchanged | Still uses unpruned `messages` | Pruning is transparent to observability; changing the estimate scope is a separate concern and would break metrics consistency |

---

## 8. Non-Goals

The following are explicitly OUT OF SCOPE for this change:

- **No storage mutation.** Session history files on disk (`<workspace>/sessions/<key>.json`) are never touched. The pruned slice is ephemeral and discarded after the LLM call returns.
- **No truncation of non-tool messages.** User messages and assistant messages are never modified, regardless of length or position.
- **No semantic selection.** Tool results are selected for protection purely by recency (position from end), not by content relevance or tool type.
- **No cross-message compression.** Each tool result message is evaluated independently against `maxChars`. There is no aggregated or shared budget across multiple messages.
- **No streaming-specific divergence.** The same `pruneHistoryToolResults` function is reused verbatim for both streaming and non-streaming paths — no separate implementation.
- **No new package or interface.** This is a single unexported function. No public API surface is added.
- **No UI configuration.** The two new config fields are JSON-only. There is no Web UI panel for controlling them in this change.
