package session

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// sanitizeSessionKey
// ---------------------------------------------------------------------------

func TestSanitizeSessionKey_Normal(t *testing.T) {
	got := sanitizeSessionKey("user:1:web-chat")
	if got != "user:1:web-chat" {
		t.Errorf("sanitizeSessionKey(\"user:1:web-chat\") = %q; want \"user:1:web-chat\"", got)
	}
}

func TestSanitizeSessionKey_PathTraversal(t *testing.T) {
	got := sanitizeSessionKey("../../etc/passwd")
	if got == "" {
		t.Fatal("sanitizeSessionKey should return non-empty for traversal input")
	}
	// Must not contain ".."
	for i := 0; i < len(got)-1; i++ {
		if got[i] == '.' && got[i+1] == '.' {
			t.Errorf("sanitizeSessionKey result %q still contains \"..\"", got)
			break
		}
	}
}

func TestSanitizeSessionKey_Slashes(t *testing.T) {
	got := sanitizeSessionKey("a/b\\c")
	if got == "" {
		t.Fatal("sanitizeSessionKey should return non-empty")
	}
	for _, c := range got {
		if c == '/' || c == '\\' {
			t.Errorf("sanitizeSessionKey result %q contains path separator", got)
		}
	}
}

func TestSanitizeSessionKey_Dot(t *testing.T) {
	got := sanitizeSessionKey(".")
	if got != "" {
		t.Errorf("sanitizeSessionKey(\".\") = %q; want \"\"", got)
	}
}

func TestSanitizeSessionKey_Empty(t *testing.T) {
	got := sanitizeSessionKey("")
	if got != "" {
		t.Errorf("sanitizeSessionKey(\"\") = %q; want \"\"", got)
	}
}

// ---------------------------------------------------------------------------
// SessionManager.namespaceKey
// ---------------------------------------------------------------------------

func TestNamespaceKey_WithUserID(t *testing.T) {
	sm := NewSessionManager("")
	got := sm.namespaceKey(42, "chat")
	if got != "user:42:chat" {
		t.Errorf("namespaceKey(42, \"chat\") = %q; want \"user:42:chat\"", got)
	}
}

func TestNamespaceKey_ZeroUserID(t *testing.T) {
	sm := NewSessionManager("")
	got := sm.namespaceKey(0, "chat")
	if got != "chat" {
		t.Errorf("namespaceKey(0, \"chat\") = %q; want \"chat\"", got)
	}
}

// ---------------------------------------------------------------------------
// GetOrCreateForUser
// ---------------------------------------------------------------------------

func TestGetOrCreateForUser_NewSession(t *testing.T) {
	sm := NewSessionManager("")

	s := sm.GetOrCreateForUser(1, "test-session")
	if s == nil {
		t.Fatal("GetOrCreateForUser returned nil")
	}
	if s.Key != "user:1:test-session" {
		t.Errorf("Key = %q; want \"user:1:test-session\"", s.Key)
	}
}

func TestGetOrCreateForUser_ReturnsSameSession(t *testing.T) {
	sm := NewSessionManager("")

	s1 := sm.GetOrCreateForUser(1, "test")
	s2 := sm.GetOrCreateForUser(1, "test")
	if s1 != s2 {
		t.Error("GetOrCreateForUser should return the same session pointer on second call")
	}
}

func TestGetOrCreate_BackwardCompat(t *testing.T) {
	sm := NewSessionManager("")

	s := sm.GetOrCreate("legacy-key")
	if s.Key != "legacy-key" {
		t.Errorf("Key = %q; want \"legacy-key\"", s.Key)
	}
}

// ---------------------------------------------------------------------------
// AddMessageForUser / GetHistoryForUser
// ---------------------------------------------------------------------------

func TestAddAndGetHistory(t *testing.T) {
	sm := NewSessionManager("")

	sm.AddMessageForUser(1, "chat", "user", "hello")
	sm.AddMessageForUser(1, "chat", "assistant", "hi there")

	history := sm.GetHistoryForUser(1, "chat")
	if len(history) != 2 {
		t.Fatalf("history length = %d; want 2", len(history))
	}
	if history[0].Content != "hello" {
		t.Errorf("history[0].Content = %q; want \"hello\"", history[0].Content)
	}
	if history[1].Role != "assistant" {
		t.Errorf("history[1].Role = %q; want \"assistant\"", history[1].Role)
	}
}

func TestGetHistory_NonExistent(t *testing.T) {
	sm := NewSessionManager("")

	history := sm.GetHistoryForUser(99, "no-session")
	if len(history) != 0 {
		t.Errorf("history length = %d; want 0 for non-existent session", len(history))
	}
}

// ---------------------------------------------------------------------------
// Summary
// ---------------------------------------------------------------------------

func TestSetAndGetSummary(t *testing.T) {
	sm := NewSessionManager("")

	sm.GetOrCreateForUser(1, "chat")
	sm.SetSummaryForUser(1, "chat", "user asked about weather")

	got := sm.GetSummaryForUser(1, "chat")
	if got != "user asked about weather" {
		t.Errorf("Summary = %q; want \"user asked about weather\"", got)
	}
}

func TestGetSummary_NonExistent(t *testing.T) {
	sm := NewSessionManager("")

	got := sm.GetSummaryForUser(1, "missing")
	if got != "" {
		t.Errorf("Summary for non-existent session = %q; want \"\"", got)
	}
}

// ---------------------------------------------------------------------------
// TruncateHistoryForUser
// ---------------------------------------------------------------------------

func TestTruncateHistory(t *testing.T) {
	sm := NewSessionManager("")

	for i := 0; i < 10; i++ {
		sm.AddMessageForUser(1, "chat", "user", "msg")
	}

	sm.TruncateHistoryForUser(1, "chat", 3)
	history := sm.GetHistoryForUser(1, "chat")

	if len(history) != 3 {
		t.Errorf("history length after truncate = %d; want 3", len(history))
	}
}

func TestTruncateHistory_NoOp(t *testing.T) {
	sm := NewSessionManager("")

	sm.AddMessageForUser(1, "chat", "user", "only one")
	sm.TruncateHistoryForUser(1, "chat", 5)
	history := sm.GetHistoryForUser(1, "chat")

	if len(history) != 1 {
		t.Errorf("history length = %d; want 1 (truncate should be no-op)", len(history))
	}
}

// ---------------------------------------------------------------------------
// SaveForUser + persistence round-trip
// ---------------------------------------------------------------------------

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	sm := NewSessionManager(dir)

	// Use userID=0 (no namespace) to avoid ":" in filename (illegal on Windows)
	sm.AddMessage("test-chat", "user", "persisted message")
	session := sm.GetOrCreate("test-chat")
	if err := sm.Save(session); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file was created
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			found = true
		}
	}
	if !found {
		t.Fatal("No .json file created in storage directory")
	}

	// Create new manager from same dir to test loading
	sm2 := NewSessionManager(dir)
	history := sm2.GetHistory("test-chat")
	if len(history) != 1 {
		t.Fatalf("Reloaded history length = %d; want 1", len(history))
	}
	if history[0].Content != "persisted message" {
		t.Errorf("Reloaded content = %q; want \"persisted message\"", history[0].Content)
	}
}

func TestSave_NoStorageIsNoOp(t *testing.T) {
	sm := NewSessionManager("")

	sm.AddMessageForUser(1, "chat", "user", "test")
	session := sm.GetOrCreateForUser(1, "chat")

	err := sm.SaveForUser(1, session)
	if err != nil {
		t.Errorf("SaveForUser with no storage should return nil, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetOrCreateForSpecialist
// ---------------------------------------------------------------------------

func TestGetOrCreateForSpecialist(t *testing.T) {
	sm := NewSessionManager("")

	s := sm.GetOrCreateForSpecialist(1, "spec-session", "code-reviewer")
	if s.AgentProfile != "code-reviewer" {
		t.Errorf("AgentProfile = %q; want \"code-reviewer\"", s.AgentProfile)
	}
	if s.SpecialistMetadata == nil {
		t.Error("SpecialistMetadata should be initialized")
	}
}

// ---------------------------------------------------------------------------
// LinkParentSession
// ---------------------------------------------------------------------------

func TestLinkParentSession(t *testing.T) {
	sm := NewSessionManager("")

	sm.GetOrCreateForUser(1, "child")
	sm.LinkParentSession(1, "child", "parent-key")

	session := sm.GetOrCreateForUser(1, "child")
	if session.ParentSessionKey != "parent-key" {
		t.Errorf("ParentSessionKey = %q; want \"parent-key\"", session.ParentSessionKey)
	}
}

// ---------------------------------------------------------------------------
// SetStorage
// ---------------------------------------------------------------------------

func TestSetStorage(t *testing.T) {
	sm := NewSessionManager("")
	sm.AddMessageForUser(1, "chat", "user", "old message")

	dir := t.TempDir()
	sm.SetStorage(dir)

	// After SetStorage, sessions should be cleared
	history := sm.GetHistoryForUser(1, "chat")
	if len(history) != 0 {
		t.Errorf("After SetStorage, history should be empty; got %d messages", len(history))
	}
}

// ---------------------------------------------------------------------------
// User isolation: different users have separate sessions
// ---------------------------------------------------------------------------

func TestUserIsolation(t *testing.T) {
	sm := NewSessionManager("")

	sm.AddMessageForUser(1, "chat", "user", "user1 message")
	sm.AddMessageForUser(2, "chat", "user", "user2 message")

	h1 := sm.GetHistoryForUser(1, "chat")
	h2 := sm.GetHistoryForUser(2, "chat")

	if len(h1) != 1 {
		t.Fatalf("user 1 history length = %d; want 1", len(h1))
	}
	if len(h2) != 1 {
		t.Fatalf("user 2 history length = %d; want 1", len(h2))
	}
	if h1[0].Content != "user1 message" {
		t.Errorf("user 1 content = %q; want %q", h1[0].Content, "user1 message")
	}
	if h2[0].Content != "user2 message" {
		t.Errorf("user 2 content = %q; want %q", h2[0].Content, "user2 message")
	}
}

// ---------------------------------------------------------------------------
// Message cap at 500
// ---------------------------------------------------------------------------

func TestMessageCap(t *testing.T) {
	sm := NewSessionManager("")

	for i := 0; i < 510; i++ {
		sm.AddMessageForUser(1, "chat", "user", "msg")
	}

	history := sm.GetHistoryForUser(1, "chat")
	if len(history) != 500 {
		t.Errorf("history length = %d; want 500 (capped)", len(history))
	}
}

// ---------------------------------------------------------------------------
// GetHistory returns a copy, not the original slice
// ---------------------------------------------------------------------------

func TestGetHistory_ReturnsCopy(t *testing.T) {
	sm := NewSessionManager("")

	sm.AddMessageForUser(1, "chat", "user", "original")
	history := sm.GetHistoryForUser(1, "chat")

	// Mutate the returned slice
	history[0].Content = "mutated"

	// Original should be unchanged
	original := sm.GetHistoryForUser(1, "chat")
	if original[0].Content != "original" {
		t.Errorf("GetHistory did not return a copy; original was mutated to %q", original[0].Content)
	}
}

// ---------------------------------------------------------------------------
// TruncateHistory on nonexistent session
// ---------------------------------------------------------------------------

func TestTruncateHistory_NonExistent(t *testing.T) {
	sm := NewSessionManager("")

	// Should not panic
	sm.TruncateHistoryForUser(99, "nonexistent", 5)
}

// ---------------------------------------------------------------------------
// SetSummary on nonexistent session is a no-op
// ---------------------------------------------------------------------------

func TestSetSummary_NonExistent(t *testing.T) {
	sm := NewSessionManager("")

	// Should not panic or set anything
	sm.SetSummaryForUser(99, "nonexistent", "some summary")

	// Should still return empty
	got := sm.GetSummaryForUser(99, "nonexistent")
	if got != "" {
		t.Errorf("Summary for non-existent session = %q; want empty", got)
	}
}

// ---------------------------------------------------------------------------
// sanitizeSessionKey with various edge cases
// ---------------------------------------------------------------------------

func TestSanitizeSessionKey_DoubleDotsInPath(t *testing.T) {
	got := sanitizeSessionKey("foo..bar")
	for i := 0; i < len(got)-1; i++ {
		if got[i] == '.' && got[i+1] == '.' {
			t.Errorf("result %q still contains '..'", got)
			break
		}
	}
}

func TestSanitizeSessionKey_WhitespaceOnly(t *testing.T) {
	got := sanitizeSessionKey("   ")
	if got != "" {
		t.Errorf("sanitizeSessionKey(whitespace) = %q; want empty", got)
	}
}

func TestSanitizeSessionKey_MixedSeparators(t *testing.T) {
	got := sanitizeSessionKey("a/b\\c/../d")
	for _, c := range got {
		if c == '/' || c == '\\' {
			t.Errorf("result %q still contains path separator", got)
			break
		}
	}
}
