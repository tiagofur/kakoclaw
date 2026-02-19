package storage

import (
	"path/filepath"
	"testing"

	"github.com/sipeed/kakoclaw/pkg/config"
)

func newTestStorage(t *testing.T) *Storage {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := New(config.StorageConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// --- SQLite PRAGMAs ---

func TestNewEnablesWALMode(t *testing.T) {
	s := newTestStorage(t)
	var mode string
	if err := s.db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Fatalf("expected journal_mode=wal, got %s", mode)
	}
}

func TestNewEnablesForeignKeys(t *testing.T) {
	s := newTestStorage(t)
	var fk int
	if err := s.db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Fatalf("expected foreign_keys=1, got %d", fk)
	}
}

func TestNewSetsSynchronousNormal(t *testing.T) {
	s := newTestStorage(t)
	var sync int
	if err := s.db.QueryRow("PRAGMA synchronous").Scan(&sync); err != nil {
		t.Fatal(err)
	}
	// NORMAL = 1
	if sync != 1 {
		t.Fatalf("expected synchronous=1 (NORMAL), got %d", sync)
	}
}

// --- Migration idempotency ---

func TestMigrateIsIdempotent(t *testing.T) {
	s := newTestStorage(t)
	// Run migrate again — should not fail on existing tables/columns
	if err := s.migrate(); err != nil {
		t.Fatalf("second migrate() should be idempotent: %v", err)
	}
	// And a third time for good measure
	if err := s.migrate(); err != nil {
		t.Fatalf("third migrate() should be idempotent: %v", err)
	}
}

// --- Chat message CRUD ---

func TestSaveAndGetMessages(t *testing.T) {
	s := newTestStorage(t)
	sess := "test:session:1"

	if err := s.SaveMessage(sess, "user", "hello"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}
	if err := s.SaveMessage(sess, "assistant", "hi there"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	msgs, err := s.GetMessages(sess)
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "hello" {
		t.Fatalf("unexpected first message: %+v", msgs[0])
	}
	if msgs[1].Role != "assistant" || msgs[1].Content != "hi there" {
		t.Fatalf("unexpected second message: %+v", msgs[1])
	}
}

func TestGetMessagesEmptySession(t *testing.T) {
	s := newTestStorage(t)
	msgs, err := s.GetMessages("nonexistent")
	if err != nil {
		t.Fatalf("GetMessages: %v", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(msgs))
	}
}

func TestSearchMessages(t *testing.T) {
	s := newTestStorage(t)
	_ = s.SaveMessage("s1", "user", "find the KakoClaw docs")
	_ = s.SaveMessage("s1", "assistant", "here are the results")
	_ = s.SaveMessage("s2", "user", "unrelated message")

	results, err := s.SearchMessages("KakoClaw")
	if err != nil {
		t.Fatalf("SearchMessages: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Content != "find the KakoClaw docs" {
		t.Fatalf("unexpected result: %s", results[0].Content)
	}
}

// --- Session CRUD ---

func TestListSessionsEmpty(t *testing.T) {
	s := newTestStorage(t)
	sessions, err := s.ListSessions(nil, 50, 0)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestListSessionsWithMessages(t *testing.T) {
	s := newTestStorage(t)
	_ = s.SaveMessage("sess:a", "user", "hello a")
	_ = s.SaveMessage("sess:b", "user", "hello b")
	_ = s.SaveMessage("sess:b", "assistant", "reply b")

	sessions, err := s.ListSessions(nil, 50, 0)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	// Find each session by ID (order may vary with identical timestamps)
	counts := map[string]int{}
	for _, ss := range sessions {
		counts[ss.SessionID] = ss.MessageCount
	}
	if counts["sess:a"] != 1 {
		t.Fatalf("expected 1 message in sess:a, got %d", counts["sess:a"])
	}
	if counts["sess:b"] != 2 {
		t.Fatalf("expected 2 messages in sess:b, got %d", counts["sess:b"])
	}
}

func TestListSessionsArchivedFilter(t *testing.T) {
	s := newTestStorage(t)
	_ = s.SaveMessage("active:1", "user", "active")
	_ = s.SaveMessage("archived:1", "user", "archived")
	archivedTrue := true
	_, _ = s.UpdateSession("archived:1", nil, &archivedTrue)

	// Default (non-archived)
	sessions, _ := s.ListSessions(nil, 50, 0)
	if len(sessions) != 1 || sessions[0].SessionID != "active:1" {
		t.Fatalf("expected only active session, got %v", sessions)
	}

	// Archived only
	sessions, _ = s.ListSessions(&archivedTrue, 50, 0)
	if len(sessions) != 1 || sessions[0].SessionID != "archived:1" {
		t.Fatalf("expected only archived session, got %v", sessions)
	}
}

func TestListSessionsPagination(t *testing.T) {
	s := newTestStorage(t)
	for i := 0; i < 5; i++ {
		_ = s.SaveMessage("pg:"+string(rune('a'+i)), "user", "msg")
	}
	page1, _ := s.ListSessions(nil, 2, 0)
	page2, _ := s.ListSessions(nil, 2, 2)
	if len(page1) != 2 {
		t.Fatalf("expected 2 in page1, got %d", len(page1))
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 in page2, got %d", len(page2))
	}
	if page1[0].SessionID == page2[0].SessionID {
		t.Fatal("pages should not overlap")
	}
}

func TestGetAndUpdateSession(t *testing.T) {
	s := newTestStorage(t)
	_ = s.SaveMessage("update:me", "user", "hello")

	sess, err := s.GetSession("update:me")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if sess.Title != "" {
		t.Fatalf("expected empty title, got %q", sess.Title)
	}

	title := "My Chat"
	updated, err := s.UpdateSession("update:me", &title, nil)
	if err != nil {
		t.Fatalf("UpdateSession: %v", err)
	}
	if updated.Title != "My Chat" {
		t.Fatalf("expected title 'My Chat', got %q", updated.Title)
	}
}

func TestDeleteSession(t *testing.T) {
	s := newTestStorage(t)
	_ = s.SaveMessage("delete:me", "user", "bye")
	_ = s.SaveMessage("delete:me", "assistant", "goodbye")

	if err := s.DeleteSession("delete:me"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}

	msgs, _ := s.GetMessages("delete:me")
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages after delete, got %d", len(msgs))
	}

	_, err := s.GetSession("delete:me")
	if err == nil {
		t.Fatal("expected error getting deleted session")
	}
}

func TestForkSession(t *testing.T) {
	s := newTestStorage(t)
	_ = s.SaveMessage("origin", "user", "msg1")
	_ = s.SaveMessage("origin", "assistant", "msg2")
	_ = s.SaveMessage("origin", "user", "msg3")

	msgs, _ := s.GetMessages("origin")
	// Fork up to second message
	count, err := s.ForkSession("origin", "forked", msgs[1].ID)
	if err != nil {
		t.Fatalf("ForkSession: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 forked messages, got %d", count)
	}

	forkedMsgs, _ := s.GetMessages("forked")
	if len(forkedMsgs) != 2 {
		t.Fatalf("expected 2 messages in fork, got %d", len(forkedMsgs))
	}
	if forkedMsgs[1].Content != "msg2" {
		t.Fatalf("expected forked msg2, got %q", forkedMsgs[1].Content)
	}
}

func TestForkSessionAll(t *testing.T) {
	s := newTestStorage(t)
	_ = s.SaveMessage("full", "user", "a")
	_ = s.SaveMessage("full", "assistant", "b")

	count, err := s.ForkSession("full", "full-fork", 0)
	if err != nil {
		t.Fatalf("ForkSession all: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestImportMessages(t *testing.T) {
	s := newTestStorage(t)
	msgs := []ImportMessage{
		{Role: "user", Content: "imported 1"},
		{Role: "assistant", Content: "imported 2"},
		{Role: "", Content: "skipped"},
	}
	count, err := s.ImportMessages("import:sess", msgs)
	if err != nil {
		t.Fatalf("ImportMessages: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 imported, got %d", count)
	}

	got, _ := s.GetMessages("import:sess")
	if len(got) != 2 {
		t.Fatalf("expected 2 stored, got %d", len(got))
	}
}

// --- Task CRUD ---

func TestTaskCreateAndGet(t *testing.T) {
	s := newTestStorage(t)
	id, err := s.CreateTask("my task", "description", "todo")
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	task, err := s.GetTask(id)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if task.Title != "my task" || task.Status != "todo" {
		t.Fatalf("unexpected task: %+v", task)
	}
}

func TestTaskUpdateAndStatusChange(t *testing.T) {
	s := newTestStorage(t)
	id, _ := s.CreateTask("update me", "", "backlog")

	updated, err := s.UpdateTask(id, "updated title", "new desc", "in_progress", "partial")
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if updated.Title != "updated title" || updated.Status != "in_progress" {
		t.Fatalf("unexpected updated task: %+v", updated)
	}

	statusChanged, err := s.UpdateTaskStatus(id, "done")
	if err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}
	if statusChanged.Status != "done" {
		t.Fatalf("expected done, got %s", statusChanged.Status)
	}
}

func TestTaskArchiveAndList(t *testing.T) {
	s := newTestStorage(t)
	id1, _ := s.CreateTask("active task", "", "todo")
	id2, _ := s.CreateTask("to archive", "", "todo")
	_ = s.ArchiveTask(id2)

	// Without archived
	tasks, _ := s.ListTasks(false)
	if len(tasks) != 1 || tasks[0].ID != id1 {
		t.Fatalf("expected only active task, got %v", tasks)
	}

	// With archived
	tasks, _ = s.ListTasks(true)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	// Unarchive
	_ = s.UnarchiveTask(id2)
	tasks, _ = s.ListTasks(false)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 after unarchive, got %d", len(tasks))
	}
}

func TestTaskDelete(t *testing.T) {
	s := newTestStorage(t)
	id, _ := s.CreateTask("delete me", "", "todo")
	if err := s.DeleteTask(id); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	_, err := s.GetTask(id)
	if err == nil {
		t.Fatal("expected error getting deleted task")
	}
}

func TestTaskLogs(t *testing.T) {
	s := newTestStorage(t)
	id, _ := s.CreateTask("logged task", "", "todo")
	_ = s.AddTaskLog(id, "created", "task created")
	_ = s.AddTaskLog(id, "started", "execution began")

	logs, err := s.GetTaskLogs(id)
	if err != nil {
		t.Fatalf("GetTaskLogs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(logs))
	}
	if logs[0].Event != "created" {
		t.Fatalf("expected first log 'created', got %q", logs[0].Event)
	}
}

// --- Close / WAL checkpoint ---

func TestCloseCheckpointsWAL(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "wal_test.db")
	s, err := New(config.StorageConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_ = s.SaveMessage("wal:test", "user", "data")

	// Close should checkpoint WAL without error
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestDataPersistsAcrossRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "persist.db")
	s1, err := New(config.StorageConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("New first instance: %v", err)
	}
	if err := s1.SaveMessage("persist:s1", "user", "hello persisted"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}
	taskID, err := s1.CreateTask("persist task", "desc", "todo")
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("Close first instance: %v", err)
	}

	s2, err := New(config.StorageConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("New second instance: %v", err)
	}
	defer func() { _ = s2.Close() }()

	msgs, err := s2.GetMessages("persist:s1")
	if err != nil {
		t.Fatalf("GetMessages after restart: %v", err)
	}
	if len(msgs) != 1 || msgs[0].Content != "hello persisted" {
		t.Fatalf("unexpected messages after restart: %+v", msgs)
	}

	task, err := s2.GetTask(taskID)
	if err != nil {
		t.Fatalf("GetTask after restart: %v", err)
	}
	if task.Title != "persist task" {
		t.Fatalf("unexpected task after restart: %+v", task)
	}
}

func TestStoragePathIsolation(t *testing.T) {
	dir := t.TempDir()
	dbA := filepath.Join(dir, "a.db")
	dbB := filepath.Join(dir, "b.db")

	sA, err := New(config.StorageConfig{Path: dbA})
	if err != nil {
		t.Fatalf("New A: %v", err)
	}
	if err := sA.SaveMessage("session:a", "user", "only in a"); err != nil {
		t.Fatalf("SaveMessage A: %v", err)
	}
	if err := sA.Close(); err != nil {
		t.Fatalf("Close A: %v", err)
	}

	sB, err := New(config.StorageConfig{Path: dbB})
	if err != nil {
		t.Fatalf("New B: %v", err)
	}
	defer func() { _ = sB.Close() }()

	msgs, err := sB.GetMessages("session:a")
	if err != nil {
		t.Fatalf("GetMessages B: %v", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected isolated empty DB B, got %d messages", len(msgs))
	}
}
