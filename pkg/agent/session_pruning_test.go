package agent

import (
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/providers"
)

// makeToolMsg builds a providers.Message for use in tests.
func makeToolMsg(role, content string) providers.Message {
	return providers.Message{Role: role, Content: content}
}

// T1: maxChars=0 → feature disabled, original slice returned unchanged (pointer identity).
func TestPruneHistoryToolResults_Disabled(t *testing.T) {
	msgs := []providers.Message{
		makeToolMsg("tool", strings.Repeat("x", 5000)),
		makeToolMsg("tool", strings.Repeat("y", 5000)),
	}
	got := pruneHistoryToolResults(msgs, 1, 0)
	if &got[0] != &msgs[0] {
		t.Fatal("expected same underlying array (pointer identity) when maxChars==0")
	}
}

// T2: no tool result messages → output equals input (no modification).
func TestPruneHistoryToolResults_NoToolResults(t *testing.T) {
	msgs := []providers.Message{
		makeToolMsg("user", "hello"),
		makeToolMsg("assistant", "world"),
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
		makeToolMsg("tool", strings.Repeat("a", 100)),
		makeToolMsg("tool", strings.Repeat("b", 100)),
		makeToolMsg("tool", strings.Repeat("c", 100)),
		makeToolMsg("tool", strings.Repeat("d", 100)),
		makeToolMsg("tool", strings.Repeat("e", 100)),
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
		makeToolMsg("tool", content),
	}
	got := pruneHistoryToolResults(msgs, 0, 100)
	want := content[:100] + "\n[truncated]"
	if got[0].Content != want {
		t.Errorf("got %q, want %q", got[0].Content, want)
	}
}

// T5: keepRecentN=2, 5 tool results (indices 0-2 old, 3-4 recent) → 0-2 truncated, 3-4 untouched.
func TestPruneHistoryToolResults_RecentPreservedOldTruncated(t *testing.T) {
	longContent := strings.Repeat("x", 600)
	msgs := []providers.Message{
		makeToolMsg("tool", longContent), // index 0 — old
		makeToolMsg("tool", longContent), // index 1 — old
		makeToolMsg("tool", longContent), // index 2 — old
		makeToolMsg("tool", longContent), // index 3 — recent (2nd newest)
		makeToolMsg("tool", longContent), // index 4 — recent (newest)
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
		makeToolMsg("tool", content),
		makeToolMsg("tool", content),
		makeToolMsg("tool", content),
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
		makeToolMsg("user", longContent),
		makeToolMsg("assistant", longContent),
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
		makeToolMsg("tool", longContent),
		makeToolMsg("user", "hello"),
	}
	originalContent := msgs[0].Content

	_ = pruneHistoryToolResults(msgs, 0, 100)

	if msgs[0].Content != originalContent {
		t.Errorf("original slice was mutated: got %q, want %q", msgs[0].Content, originalContent)
	}
}
