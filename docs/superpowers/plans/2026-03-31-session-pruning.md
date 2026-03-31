# Session Pruning Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add in-memory truncation of old tool result messages before each LLM API call to reduce token usage and delay session summarization.

**Architecture:** A pure function `pruneHistoryToolResults` is added to `pkg/agent/loop.go` that accepts the history slice and returns a copy with oversized older tool results truncated; it is called in both `runLLMIteration` and `runLLMIterationStream` immediately before each primary `Chat`/`ChatStream` call, leaving the original `messages` variable untouched for logging and metrics. Two config fields in `AgentDefaults` control the behavior: `ToolResultMaxChars` (default 2000, 0 = disabled) and `ToolResultKeepRecent` (default 3).

**Tech Stack:** Go 1.26, `pkg/config`, `pkg/agent`, `pkg/providers`

---

## File Map

| File | Action | What changes |
|------|--------|--------------|
| `pkg/config/config.go` | Modify | Add `ToolResultMaxChars` and `ToolResultKeepRecent` to `AgentDefaults`; add defaults; add to `GetUserConfigTemplate` |
| `pkg/agent/loop.go` | Modify | Add `pruneHistoryToolResults` function; call it in `runLLMIteration` and `runLLMIterationStream` |
| `pkg/agent/session_pruning_test.go` | Create | Unit tests T1–T8 for `pruneHistoryToolResults` |

---

## Task 1: Config fields

**Files:**
- Modify: `pkg/config/config.go`

- [ ] **Step 1: Write failing test for the new config fields**

Add a new test function to `pkg/config/config_test.go`:

```go
func TestAgentDefaultsToolResultFields(t *testing.T) {
	// DefaultConfig must have non-zero values for both fields
	cfg := DefaultConfig()
	if cfg.Agents.Defaults.ToolResultMaxChars == 0 {
		t.Error("expected ToolResultMaxChars default > 0, got 0 (would disable the feature)")
	}
	if cfg.Agents.Defaults.ToolResultKeepRecent == 0 {
		t.Error("expected ToolResultKeepRecent default > 0, got 0")
	}

	// Fields must round-trip through JSON
	data, err := json.Marshal(cfg.Agents.Defaults)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var roundtrip AgentDefaults
	if err := json.Unmarshal(data, &roundtrip); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if roundtrip.ToolResultMaxChars != cfg.Agents.Defaults.ToolResultMaxChars {
		t.Errorf("ToolResultMaxChars roundtrip: got %d, want %d", roundtrip.ToolResultMaxChars, cfg.Agents.Defaults.ToolResultMaxChars)
	}
	if roundtrip.ToolResultKeepRecent != cfg.Agents.Defaults.ToolResultKeepRecent {
		t.Errorf("ToolResultKeepRecent roundtrip: got %d, want %d", roundtrip.ToolResultKeepRecent, cfg.Agents.Defaults.ToolResultKeepRecent)
	}

	// Missing fields in JSON should unmarshal to zero (Go default) —
	// document this as the known zero-value limitation for user configs.
	partial := `{}`
	var empty AgentDefaults
	if err := json.Unmarshal([]byte(partial), &empty); err != nil {
		t.Fatalf("unmarshal partial: %v", err)
	}
	if empty.ToolResultMaxChars != 0 {
		t.Errorf("expected zero for absent field, got %d", empty.ToolResultMaxChars)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
go test ./pkg/config/... -run TestAgentDefaultsToolResultFields -v
```

Expected: FAIL — `ToolResultMaxChars` and `ToolResultKeepRecent` fields do not exist yet.

- [ ] **Step 3: Add fields to AgentDefaults struct**

Find `AgentDefaults` (line 184 in `pkg/config/config.go`) and append two fields after `MemoryFlushBeforeCompaction`:

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
	ToolResultMaxChars          int     `json:"tool_result_max_chars"`   // 0 = disabled; default 2000
	ToolResultKeepRecent        int     `json:"tool_result_keep_recent"` // default 3
}
```

- [ ] **Step 4: Add defaults in DefaultConfig()**

Find the `Defaults: AgentDefaults{...}` block inside `DefaultConfig()` (line ~598 in `pkg/config/config.go`) and append the two new fields:

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
	ToolResultMaxChars:          2000,
	ToolResultKeepRecent:        3,
},
```

- [ ] **Step 5: Add fields to GetUserConfigTemplate()**

Find the `Defaults: AgentDefaults{...}` block inside `GetUserConfigTemplate()` (line ~932 in `pkg/config/config.go`) and add the two inherited fields following the same pattern as `MemoryFlushBeforeCompaction`:

```go
Defaults: AgentDefaults{
	// Inherit non-sensitive defaults
	Workspace:                   filepath.Join("~", ".MakoClaw", "users", "{uuid}", "workspace"),
	RestrictToWorkspace:         globalConfig.Agents.Defaults.RestrictToWorkspace,
	Provider:                    "",                                             // Empty: user must choose
	Model:                       globalConfig.Agents.Defaults.Model,             // Inherit
	MaxTokens:                   globalConfig.Agents.Defaults.MaxTokens,         // Inherit
	Temperature:                 globalConfig.Agents.Defaults.Temperature,       // Inherit
	MaxToolIterations:           globalConfig.Agents.Defaults.MaxToolIterations, // Inherit
	MemoryFlushBeforeCompaction: globalConfig.Agents.Defaults.MemoryFlushBeforeCompaction, // Inherit
	ToolResultMaxChars:          globalConfig.Agents.Defaults.ToolResultMaxChars,          // Inherit
	ToolResultKeepRecent:        globalConfig.Agents.Defaults.ToolResultKeepRecent,        // Inherit
},
```

> **Note on MergeConfigs zero-value risk:** `MergeConfigs` merges the entire `Agents` section as a block — if a user config JSON has the agents section present but omits `tool_result_max_chars`, Go will deserialize it as `0`, silently disabling the feature for that user. `GetUserConfigTemplate` inherits the global defaults (Step 5 above), so new users created via the template will have correct values. Existing users who upgrade will have `0` until they update their config. This is a known limitation; a follow-up could add zero-value fallback logic to `MergeConfigs` for `int` fields, but that is out of scope for this change.

- [ ] **Step 6: Run the test to verify it passes**

```bash
go test ./pkg/config/... -run TestAgentDefaultsToolResultFields -v
```

Expected: PASS.

- [ ] **Step 7: Build to verify no compilation errors**

```bash
go build ./pkg/config/...
```

Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add pkg/config/config.go pkg/config/config_test.go
git commit -m "feat(config): add ToolResultMaxChars and ToolResultKeepRecent to AgentDefaults"
```

---

## Task 2: pruneHistoryToolResults function + tests

**Files:**
- Modify: `pkg/agent/loop.go`
- Create: `pkg/agent/session_pruning_test.go`

- [ ] **Step 1: Create test file with T1 (Disabled — fails because function does not exist yet)**

Create `pkg/agent/session_pruning_test.go`:

```go
package agent

import (
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/providers"
)

// makeMsg builds a providers.Message for use in tests.
func makeMsg(role, content string) providers.Message {
	return providers.Message{Role: role, Content: content}
}

// T1: maxChars=0 → feature disabled, original slice returned unchanged (pointer identity).
func TestPruneHistoryToolResults_Disabled(t *testing.T) {
	msgs := []providers.Message{
		makeMsg("tool", strings.Repeat("x", 5000)),
		makeMsg("tool", strings.Repeat("y", 5000)),
	}
	got := pruneHistoryToolResults(msgs, 1, 0)
	if &got[0] != &msgs[0] {
		t.Fatal("expected same underlying array (pointer identity) when maxChars==0")
	}
}
```

- [ ] **Step 2: Run T1 to verify it fails**

```bash
go test ./pkg/agent/... -run TestPruneHistoryToolResults_Disabled -v
```

Expected: compile error — `pruneHistoryToolResults undefined`.

- [ ] **Step 3: Add pruneHistoryToolResults to loop.go**

Append at the bottom of `pkg/agent/loop.go`, after `runMemoryFlushTurn`:

```go
// pruneHistoryToolResults returns a copy of messages where tool-result entries
// older than the keepRecentN most recent are truncated to maxChars characters.
// If maxChars == 0, the original slice is returned unchanged (feature disabled).
// The original slice is never mutated.
func pruneHistoryToolResults(messages []providers.Message, keepRecentN int, maxChars int) []providers.Message {
	if maxChars == 0 {
		return messages
	}
	// Scan newest-first to identify the keepRecentN most recent tool results.
	// All others are candidates for truncation.
	recentCount := 0
	recentSet := make(map[int]bool)
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "tool" {
			if recentCount < keepRecentN {
				recentSet[i] = true
				recentCount++
			}
		}
	}
	// Build output via shallow copy, then overwrite Content for truncation candidates.
	output := make([]providers.Message, len(messages))
	copy(output, messages)
	for i := range output {
		if output[i].Role != "tool" || recentSet[i] {
			continue
		}
		if len(output[i].Content) > maxChars {
			output[i].Content = output[i].Content[:maxChars] + "\n[truncated]"
		}
	}
	return output
}
```

- [ ] **Step 4: Run T1 to verify it passes**

```bash
go test ./pkg/agent/... -run TestPruneHistoryToolResults_Disabled -v
```

Expected: PASS.

- [ ] **Step 5: Add T2–T8 to session_pruning_test.go**

Append the following tests to `pkg/agent/session_pruning_test.go`:

```go
// T2: no tool result messages → output equals input (no modification).
func TestPruneHistoryToolResults_NoToolResults(t *testing.T) {
	msgs := []providers.Message{
		makeMsg("user", "hello"),
		makeMsg("assistant", "world"),
	}
	got := pruneHistoryToolResults(msgs, 1, 200)
	if len(got) != len(msgs) {
		t.Fatalf("length mismatch: got %d, want %d", len(got), len(msgs))
	}
	for i := range msgs {
		if got[i].Content != msgs[i].Content {
			t.Errorf("index %d: content changed unexpectedly", i)
		}
	}
}

// T3: all tool results within maxChars → none are truncated.
func TestPruneHistoryToolResults_AllWithinMaxChars(t *testing.T) {
	msgs := []providers.Message{
		makeMsg("tool", strings.Repeat("a", 100)),
		makeMsg("tool", strings.Repeat("b", 100)),
		makeMsg("tool", strings.Repeat("c", 100)),
		makeMsg("tool", strings.Repeat("d", 100)),
		makeMsg("tool", strings.Repeat("e", 100)),
	}
	got := pruneHistoryToolResults(msgs, 0, 200) // keepRecentN=0, all candidates
	for i, m := range got {
		if strings.HasSuffix(m.Content, "[truncated]") {
			t.Errorf("index %d: unexpected truncation (content was %d chars, maxChars=200)", i, len(msgs[i].Content))
		}
		if m.Content != msgs[i].Content {
			t.Errorf("index %d: content changed unexpectedly", i)
		}
	}
}

// T4: keepRecentN=0, one old tool result exceeds maxChars → truncated with suffix.
func TestPruneHistoryToolResults_OldResultTruncated(t *testing.T) {
	content := strings.Repeat("z", 500)
	msgs := []providers.Message{
		makeMsg("tool", content),
	}
	got := pruneHistoryToolResults(msgs, 0, 100)
	if len(got[0].Content) != len("z"*100+"\n[truncated]") {
		// Use explicit length check
	}
	want := content[:100] + "\n[truncated]"
	if got[0].Content != want {
		t.Errorf("got %q, want %q", got[0].Content, want)
	}
}

// T5: keepRecentN=2, 5 tool results (indices 0-2 old, 3-4 recent) → 0-2 truncated, 3-4 untouched.
func TestPruneHistoryToolResults_RecentPreservedOldTruncated(t *testing.T) {
	longContent := strings.Repeat("x", 600)
	msgs := []providers.Message{
		makeMsg("tool", longContent), // index 0 — old
		makeMsg("tool", longContent), // index 1 — old
		makeMsg("tool", longContent), // index 2 — old
		makeMsg("tool", longContent), // index 3 — recent (2nd newest)
		makeMsg("tool", longContent), // index 4 — recent (newest)
	}
	got := pruneHistoryToolResults(msgs, 2, 100)

	// Indices 3 and 4 must be untouched
	for _, i := range []int{3, 4} {
		if got[i].Content != longContent {
			t.Errorf("index %d: recent result was modified", i)
		}
	}

	// Indices 0, 1, 2 must be truncated
	want := longContent[:100] + "\n[truncated]"
	for _, i := range []int{0, 1, 2} {
		if got[i].Content != want {
			t.Errorf("index %d: got %q, want %q", i, got[i].Content, want)
		}
	}
}

// T6: keepRecentN=0 → all tool results are candidates and get truncated.
func TestPruneHistoryToolResults_KeepRecentZero(t *testing.T) {
	content := strings.Repeat("q", 300)
	msgs := []providers.Message{
		makeMsg("tool", content),
		makeMsg("tool", content),
		makeMsg("tool", content),
	}
	got := pruneHistoryToolResults(msgs, 0, 50)
	want := content[:50] + "\n[truncated]"
	for i, m := range got {
		if m.Content != want {
			t.Errorf("index %d: got %q, want %q", i, m.Content, want)
		}
	}
}

// T7: non-tool messages (user, assistant) are never modified even if > maxChars.
func TestPruneHistoryToolResults_NonToolMessagesNeverModified(t *testing.T) {
	longContent := strings.Repeat("w", 5000)
	msgs := []providers.Message{
		makeMsg("user", longContent),
		makeMsg("assistant", longContent),
	}
	got := pruneHistoryToolResults(msgs, 0, 100)
	for i, m := range got {
		if m.Content != longContent {
			t.Errorf("index %d (role=%q): non-tool message was modified", i, msgs[i].Role)
		}
	}
}

// T8: original slice and its Message structs are not mutated by the call.
func TestPruneHistoryToolResults_OriginalNotMutated(t *testing.T) {
	longContent := strings.Repeat("m", 500)
	msgs := []providers.Message{
		makeMsg("tool", longContent),
		makeMsg("user", "hello"),
	}
	originalContent := msgs[0].Content // snapshot before call

	_ = pruneHistoryToolResults(msgs, 0, 100)

	if msgs[0].Content != originalContent {
		t.Errorf("original slice was mutated: got %q, want %q", msgs[0].Content, originalContent)
	}
}
```

- [ ] **Step 6: Run all T1–T8 to verify they all pass**

```bash
go test ./pkg/agent/... -run TestPruneHistoryToolResults -v
```

Expected: all 8 tests PASS.

- [ ] **Step 7: Commit**

```bash
git add pkg/agent/loop.go pkg/agent/session_pruning_test.go
git commit -m "feat(agent): add pruneHistoryToolResults with full unit test coverage"
```

---

## Task 3: Integrate into runLLMIteration and runLLMIterationStream

**Files:**
- Modify: `pkg/agent/loop.go`

- [ ] **Step 1: Wrap the Chat() call in runLLMIteration**

In `pkg/agent/loop.go`, inside `runLLMIteration`, find the primary `al.provider.Chat()` call (around line 1646). It currently looks like:

```go
// Call LLM
llmStart := time.Now()
response, err := al.provider.Chat(ctx, messages, providerToolDefs, model, map[string]interface{}{
    "max_tokens":  8192,
    "temperature": 0.7,
})
```

Replace with:

```go
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

Do NOT modify the secondary `al.provider.Chat()` call at the end of `runLLMIteration` (the max-iterations concluding call around line 1886) — that call operates on a mutated slice and must not be wrapped.

- [ ] **Step 2: Wrap the ChatStream() call in runLLMIterationStream (streaming path)**

In `pkg/agent/loop.go`, inside `runLLMIterationStream`, find the `streamingProvider.ChatStream()` call inside the `if canStream {` block (around line 2020). It currently looks like:

```go
if canStream {
    llmStart := time.Now()
    ch, err := streamingProvider.ChatStream(ctx, messages, providerToolDefs, model, llmOpts)
```

Replace with:

```go
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

- [ ] **Step 3: Wrap the fallback non-streaming Chat() call in runLLMIterationStream**

Still in `runLLMIterationStream`, find the fallback `al.provider.Chat()` call (around line 2240). It currently looks like:

```go
// Fallback: non-streaming Chat()
fallbackStart := time.Now()
response, err := al.provider.Chat(ctx, messages, providerToolDefs, model, llmOpts)
```

Replace with:

```go
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

Note: `prunedMessages` is scoped inside the `if canStream` block (Step 2) and inside the fallback block (Step 3) respectively — no name collision. Do NOT wrap the concluding Chat call in `runLLMIterationStream` (max-iterations path around line 2401).

- [ ] **Step 4: Build to verify it compiles**

```bash
go build ./pkg/agent/...
```

Expected: no errors.

- [ ] **Step 5: Run all agent tests**

```bash
go test ./pkg/agent/... -v -timeout 60s
```

Expected: all tests PASS, including T1–T8 in `session_pruning_test.go`.

- [ ] **Step 6: Commit**

```bash
git add pkg/agent/loop.go
git commit -m "feat(agent): integrate pruneHistoryToolResults into LLM iteration paths"
```

---

## Self-Review Checklist

- [ ] `ToolResultMaxChars` and `ToolResultKeepRecent` added to struct, `DefaultConfig()`, and `GetUserConfigTemplate()`
- [ ] Zero-value limitation documented in plan (no silent behavior change for existing users)
- [ ] `pruneHistoryToolResults` never mutates the input slice (copy semantics verified by T8)
- [ ] `maxChars == 0` returns input slice unchanged with no allocation (T1)
- [ ] Non-tool messages are never modified (T7)
- [ ] Recent N tool results protected, older ones truncated (T5)
- [ ] Integration in `runLLMIteration`: primary Chat call only, not the concluding call
- [ ] Integration in `runLLMIterationStream`: streaming path + fallback path, not the concluding call
- [ ] Original `messages` variable never overwritten — logging and `estimateTokens` remain on unpruned slice
- [ ] All 8 tests pass before integration commit
