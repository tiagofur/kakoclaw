package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// --- escapeLikeQuery tests ---

func TestEscapeLikeQuery_NoSpecialChars(t *testing.T) {
	input := "hello world"
	got := escapeLikeQuery(input)
	if got != input {
		t.Fatalf("expected %q, got %q", input, got)
	}
}

func TestEscapeLikeQuery_EscapesPercent(t *testing.T) {
	got := escapeLikeQuery("100%")
	want := `100\%`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestEscapeLikeQuery_EscapesUnderscore(t *testing.T) {
	got := escapeLikeQuery("user_name")
	want := `user\_name`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestEscapeLikeQuery_EscapesBackslash(t *testing.T) {
	got := escapeLikeQuery(`path\to\file`)
	want := `path\\to\\file`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestEscapeLikeQuery_AllSpecialCharsCombined(t *testing.T) {
	got := escapeLikeQuery(`50%_off\deal`)
	want := `50\%\_off\\deal`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestEscapeLikeQuery_EmptyString(t *testing.T) {
	got := escapeLikeQuery("")
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

// --- normalizeUserKey tests ---

func TestNormalizeUserKey_PositiveInt64_Unchanged(t *testing.T) {
	got := normalizeUserKey(int64(42))
	if got != int64(42) {
		t.Fatalf("expected int64(42), got %v (%T)", got, got)
	}
}

func TestNormalizeUserKey_ZeroInt64_ReturnsOne(t *testing.T) {
	got := normalizeUserKey(int64(0))
	if got != 1 {
		t.Fatalf("expected 1, got %v (%T)", got, got)
	}
}

func TestNormalizeUserKey_NegativeInt64_ReturnsOne(t *testing.T) {
	got := normalizeUserKey(int64(-5))
	if got != 1 {
		t.Fatalf("expected 1, got %v (%T)", got, got)
	}
}

func TestNormalizeUserKey_StringNumericPositive(t *testing.T) {
	got := normalizeUserKey("99")
	if got != int64(99) {
		t.Fatalf("expected int64(99), got %v (%T)", got, got)
	}
}

func TestNormalizeUserKey_StringNumericZero(t *testing.T) {
	got := normalizeUserKey("0")
	if got != 1 {
		t.Fatalf("expected 1, got %v (%T)", got, got)
	}
}

func TestNormalizeUserKey_StringNumericNegative(t *testing.T) {
	got := normalizeUserKey("-3")
	if got != 1 {
		t.Fatalf("expected 1, got %v (%T)", got, got)
	}
}

func TestNormalizeUserKey_NonNumericString_Passthrough(t *testing.T) {
	got := normalizeUserKey("alice")
	if got != "alice" {
		t.Fatalf("expected %q, got %v (%T)", "alice", got, got)
	}
}

func TestNormalizeUserKey_OtherType_Passthrough(t *testing.T) {
	got := normalizeUserKey(3.14)
	if got != 3.14 {
		t.Fatalf("expected 3.14, got %v (%T)", got, got)
	}
}

// --- expandHome tests ---

func TestExpandHome_TildeSlash(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home directory: %v", err)
	}
	got := expandHome("~/documents/file.txt")
	want := home + "/documents/file.txt"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestExpandHome_TildeOnly(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home directory: %v", err)
	}
	got := expandHome("~")
	if got != home {
		t.Fatalf("expected %q, got %q", home, got)
	}
}

func TestExpandHome_NoTilde(t *testing.T) {
	input := "/absolute/path/to/file"
	got := expandHome(input)
	if got != input {
		t.Fatalf("expected %q, got %q", input, got)
	}
}

func TestExpandHome_EmptyString(t *testing.T) {
	got := expandHome("")
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestExpandHome_RelativePath(t *testing.T) {
	input := "relative/path"
	got := expandHome(input)
	if got != input {
		t.Fatalf("expected %q, got %q", input, got)
	}
}

func TestExpandHome_TildeInMiddle(t *testing.T) {
	// tilde not at position 0 should not be expanded
	input := "/some/~/path"
	got := expandHome(input)
	if got != input {
		t.Fatalf("expected %q, got %q", input, got)
	}
}

func TestExpandHome_TildeWithWindowsSeparator(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}
	// On Windows, ~/path uses forward slash; ~ without / should return home
	// The function only checks for '/' after '~', so ~\path would NOT be expanded
	input := `~\documents`
	home, _ := os.UserHomeDir()
	got := expandHome(input)
	// Since expandHome checks path[1] == '/', this should return just home
	// because len > 1 and path[1] != '/', so it falls through to return home
	if got != home {
		t.Fatalf("expected %q, got %q", home, got)
	}
}

// --- newTestUserStorage helper ---

func newTestUserStorage(t *testing.T) *Storage {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "user_test.db")
	s, err := NewUserStorage(dbPath)
	if err != nil {
		t.Fatalf("NewUserStorage: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// --- NewUserStorage PRAGMA and migration tests ---

func TestNewUserStorage_WALMode(t *testing.T) {
	s := newTestUserStorage(t)
	var mode string
	if err := s.db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Fatalf("expected journal_mode=wal, got %s", mode)
	}
}

func TestNewUserStorage_IsUserDB(t *testing.T) {
	s := newTestUserStorage(t)
	if !s.isUserDB {
		t.Fatal("expected isUserDB to be true for NewUserStorage")
	}
}

func TestMigrateUserDB_Idempotent(t *testing.T) {
	s := newTestUserStorage(t)
	if err := s.migrateUserDB(); err != nil {
		t.Fatalf("second migrateUserDB should be idempotent: %v", err)
	}
	if err := s.migrateUserDB(); err != nil {
		t.Fatalf("third migrateUserDB should be idempotent: %v", err)
	}
}
