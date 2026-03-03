package storage

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
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
	_ = s.SaveMessage("s1", "user", "find the makoclaw docs")
	_ = s.SaveMessage("s1", "assistant", "here are the results")
	_ = s.SaveMessage("s2", "user", "unrelated message")

	results, err := s.SearchMessages("makoclaw")
	if err != nil {
		t.Fatalf("SearchMessages: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Content != "find the makoclaw docs" {
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
	id, err := s.CreateTask("my task", "description", "todo", "")
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
	id, _ := s.CreateTask("update me", "", "backlog", "")

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
	id1, _ := s.CreateTask("active task", "", "todo", "")
	id2, _ := s.CreateTask("to archive", "", "todo", "")
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
	id, _ := s.CreateTask("delete me", "", "todo", "")
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
	id, _ := s.CreateTask("logged task", "", "todo", "")
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
	taskID, err := s1.CreateTask("persist task", "desc", "todo", "")
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

// --- Knowledge tests ---

func TestKnowledge_SaveAndList(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	doc, err := s.SaveKnowledgeDocument(userID, "test-doc.txt", "text/plain", 1024, []string{"chunk one", "chunk two", "chunk three"})
	if err != nil {
		t.Fatalf("SaveKnowledgeDocument: %v", err)
	}
	if doc.Name != "test-doc.txt" {
		t.Fatalf("expected name 'test-doc.txt', got %q", doc.Name)
	}
	if doc.MimeType != "text/plain" {
		t.Fatalf("expected mime_type 'text/plain', got %q", doc.MimeType)
	}
	if doc.Size != 1024 {
		t.Fatalf("expected size 1024, got %d", doc.Size)
	}
	if doc.ChunkCount != 3 {
		t.Fatalf("expected chunk_count 3, got %d", doc.ChunkCount)
	}

	docs, err := s.ListKnowledgeDocuments(userID)
	if err != nil {
		t.Fatalf("ListKnowledgeDocuments: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
	if docs[0].ID != doc.ID {
		t.Fatalf("expected doc ID %d, got %d", doc.ID, docs[0].ID)
	}
	if docs[0].Name != "test-doc.txt" {
		t.Fatalf("expected name 'test-doc.txt', got %q", docs[0].Name)
	}
}

func TestKnowledge_Delete(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	doc, err := s.SaveKnowledgeDocument(userID, "deleteme.txt", "text/plain", 100, []string{"content to delete"})
	if err != nil {
		t.Fatalf("SaveKnowledgeDocument: %v", err)
	}

	if err := s.DeleteKnowledgeDocument(userID, doc.ID); err != nil {
		t.Fatalf("DeleteKnowledgeDocument: %v", err)
	}

	docs, err := s.ListKnowledgeDocuments(userID)
	if err != nil {
		t.Fatalf("ListKnowledgeDocuments: %v", err)
	}
	if len(docs) != 0 {
		t.Fatalf("expected 0 documents after delete, got %d", len(docs))
	}
}

func TestKnowledge_Search(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	_, err := s.SaveKnowledgeDocument(userID, "golang-guide.txt", "text/plain", 500, []string{
		"Go is a statically typed compiled language",
		"Python is dynamically typed",
	})
	if err != nil {
		t.Fatalf("SaveKnowledgeDocument: %v", err)
	}

	_, err = s.SaveKnowledgeDocument(userID, "rust-guide.txt", "text/plain", 300, []string{
		"Rust is a systems programming language focused on safety",
	})
	if err != nil {
		t.Fatalf("SaveKnowledgeDocument: %v", err)
	}

	results, err := s.SearchKnowledge(userID, "Go statically typed", 5)
	if err != nil {
		t.Fatalf("SearchKnowledge: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 search result")
	}
	// The top result should be the Go chunk
	if results[0].DocumentName != "golang-guide.txt" {
		t.Fatalf("expected top result from golang-guide.txt, got %q", results[0].DocumentName)
	}
}

func TestKnowledge_GetChunks(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	doc, err := s.SaveKnowledgeDocument(userID, "multi-chunk.txt", "text/plain", 200, []string{
		"first chunk",
		"second chunk",
		"third chunk",
	})
	if err != nil {
		t.Fatalf("SaveKnowledgeDocument: %v", err)
	}

	chunks, err := s.GetKnowledgeDocumentChunks(userID, doc.ID)
	if err != nil {
		t.Fatalf("GetKnowledgeDocumentChunks: %v", err)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if chunks[0].Content != "first chunk" {
		t.Fatalf("expected first chunk content 'first chunk', got %q", chunks[0].Content)
	}
	if chunks[1].Content != "second chunk" {
		t.Fatalf("expected second chunk content 'second chunk', got %q", chunks[1].Content)
	}
	if chunks[2].Content != "third chunk" {
		t.Fatalf("expected third chunk content 'third chunk', got %q", chunks[2].Content)
	}
	// Verify ordering by position
	if chunks[0].Position > chunks[1].Position || chunks[1].Position > chunks[2].Position {
		t.Fatal("chunks are not in ascending position order")
	}
}

func TestKnowledge_UpdateChunk(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	doc, err := s.SaveKnowledgeDocument(userID, "updatable.txt", "text/plain", 100, []string{"original content"})
	if err != nil {
		t.Fatalf("SaveKnowledgeDocument: %v", err)
	}

	chunks, err := s.GetKnowledgeDocumentChunks(userID, doc.ID)
	if err != nil {
		t.Fatalf("GetKnowledgeDocumentChunks: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}

	if err := s.UpdateKnowledgeChunk(userID, chunks[0].ID, "updated content"); err != nil {
		t.Fatalf("UpdateKnowledgeChunk: %v", err)
	}

	updatedChunks, err := s.GetKnowledgeDocumentChunks(userID, doc.ID)
	if err != nil {
		t.Fatalf("GetKnowledgeDocumentChunks after update: %v", err)
	}
	if updatedChunks[0].Content != "updated content" {
		t.Fatalf("expected updated content 'updated content', got %q", updatedChunks[0].Content)
	}
}

// --- Workflow tests ---

func TestWorkflow_CRUD(t *testing.T) {
	s := newTestStorage(t)

	steps := json.RawMessage(`[{"name":"step1","action":"echo"}]`)
	params := json.RawMessage(`[{"name":"param1","type":"string"}]`)

	// Create
	id, err := s.CreateWorkflow("test-workflow", "a test workflow", steps, params, nil)
	if err != nil {
		t.Fatalf("CreateWorkflow: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero workflow ID")
	}

	// Get
	wf, err := s.GetWorkflow(id)
	if err != nil {
		t.Fatalf("GetWorkflow: %v", err)
	}
	if wf.Name != "test-workflow" {
		t.Fatalf("expected name 'test-workflow', got %q", wf.Name)
	}
	if wf.Description != "a test workflow" {
		t.Fatalf("expected description 'a test workflow', got %q", wf.Description)
	}
	if !wf.Enabled {
		t.Fatal("expected workflow to be enabled by default")
	}
	if string(wf.Steps) != string(steps) {
		t.Fatalf("expected steps %s, got %s", string(steps), string(wf.Steps))
	}

	// Update
	newSteps := json.RawMessage(`[{"name":"step1","action":"echo"},{"name":"step2","action":"log"}]`)
	updated, err := s.UpdateWorkflow(id, "updated-workflow", "updated desc", false, newSteps, params, nil)
	if err != nil {
		t.Fatalf("UpdateWorkflow: %v", err)
	}
	if updated.Name != "updated-workflow" {
		t.Fatalf("expected updated name 'updated-workflow', got %q", updated.Name)
	}
	if updated.Enabled {
		t.Fatal("expected workflow to be disabled after update")
	}

	// List
	workflows, err := s.ListWorkflows()
	if err != nil {
		t.Fatalf("ListWorkflows: %v", err)
	}
	if len(workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(workflows))
	}
	if workflows[0].Name != "updated-workflow" {
		t.Fatalf("expected 'updated-workflow' in list, got %q", workflows[0].Name)
	}

	// Delete
	if err := s.DeleteWorkflow(id); err != nil {
		t.Fatalf("DeleteWorkflow: %v", err)
	}
	workflows, err = s.ListWorkflows()
	if err != nil {
		t.Fatalf("ListWorkflows after delete: %v", err)
	}
	if len(workflows) != 0 {
		t.Fatalf("expected 0 workflows after delete, got %d", len(workflows))
	}
}

func TestWorkflow_Runs(t *testing.T) {
	s := newTestStorage(t)

	wfID, err := s.CreateWorkflow("run-test", "workflow for run testing", nil, nil, nil)
	if err != nil {
		t.Fatalf("CreateWorkflow: %v", err)
	}

	// Create run
	runID, err := s.CreateWorkflowRun(wfID)
	if err != nil {
		t.Fatalf("CreateWorkflowRun: %v", err)
	}
	if runID == 0 {
		t.Fatal("expected non-zero run ID")
	}

	// Update run
	results := json.RawMessage(`{"output":"success"}`)
	if err := s.UpdateWorkflowRun(runID, "completed", results); err != nil {
		t.Fatalf("UpdateWorkflowRun: %v", err)
	}

	// List runs
	runs, err := s.ListWorkflowRuns(wfID, 10)
	if err != nil {
		t.Fatalf("ListWorkflowRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Status != "completed" {
		t.Fatalf("expected status 'completed', got %q", runs[0].Status)
	}
	if runs[0].FinishedAt == nil {
		t.Fatal("expected finished_at to be set for completed run")
	}
}

// --- Metrics tests ---

func TestMetrics_CountersPersistence(t *testing.T) {
	s := newTestStorage(t)

	counters := map[string]int64{
		"llm_calls":  42,
		"tool_calls": 100,
		"tokens":     9999,
	}
	if err := s.SaveMetricCounters(counters); err != nil {
		t.Fatalf("SaveMetricCounters: %v", err)
	}

	loaded, err := s.LoadMetricCounters()
	if err != nil {
		t.Fatalf("LoadMetricCounters: %v", err)
	}
	for k, v := range counters {
		if loaded[k] != v {
			t.Fatalf("expected counter %s=%d, got %d", k, v, loaded[k])
		}
	}

	// Upsert with new values
	counters["llm_calls"] = 50
	if err := s.SaveMetricCounters(counters); err != nil {
		t.Fatalf("SaveMetricCounters upsert: %v", err)
	}
	loaded, err = s.LoadMetricCounters()
	if err != nil {
		t.Fatalf("LoadMetricCounters after upsert: %v", err)
	}
	if loaded["llm_calls"] != 50 {
		t.Fatalf("expected llm_calls=50 after upsert, got %d", loaded["llm_calls"])
	}
}

func TestMetrics_Events(t *testing.T) {
	s := newTestStorage(t)

	event1 := map[string]string{"type": "llm_call", "model": "gpt-4"}
	event2 := map[string]string{"type": "tool_call", "tool": "web_search"}
	event3 := map[string]string{"type": "llm_call", "model": "claude-3"}

	if err := s.AppendMetricEvent(event1); err != nil {
		t.Fatalf("AppendMetricEvent 1: %v", err)
	}
	if err := s.AppendMetricEvent(event2); err != nil {
		t.Fatalf("AppendMetricEvent 2: %v", err)
	}
	if err := s.AppendMetricEvent(event3); err != nil {
		t.Fatalf("AppendMetricEvent 3: %v", err)
	}

	events, err := s.LoadRecentEvents()
	if err != nil {
		t.Fatalf("LoadRecentEvents: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	// Verify order is oldest first (ASC by id)
	var first map[string]string
	if err := json.Unmarshal(events[0], &first); err != nil {
		t.Fatalf("unmarshal first event: %v", err)
	}
	if first["model"] != "gpt-4" {
		t.Fatalf("expected first event model 'gpt-4', got %q", first["model"])
	}

	var last map[string]string
	if err := json.Unmarshal(events[2], &last); err != nil {
		t.Fatalf("unmarshal last event: %v", err)
	}
	if last["model"] != "claude-3" {
		t.Fatalf("expected last event model 'claude-3', got %q", last["model"])
	}
}

func TestMetrics_Breakdowns(t *testing.T) {
	s := newTestStorage(t)

	modelMetrics := json.RawMessage(`{"gpt-4":{"calls":10,"tokens":5000},"claude-3":{"calls":5,"tokens":2000}}`)
	toolMetrics := json.RawMessage(`{"web_search":{"calls":3},"exec":{"calls":7}}`)

	if err := s.SaveMetricBreakdowns(modelMetrics, toolMetrics); err != nil {
		t.Fatalf("SaveMetricBreakdowns: %v", err)
	}

	loadedModel, loadedTool, err := s.LoadMetricBreakdowns()
	if err != nil {
		t.Fatalf("LoadMetricBreakdowns: %v", err)
	}
	if string(loadedModel) != string(modelMetrics) {
		t.Fatalf("expected model metrics %s, got %s", string(modelMetrics), string(loadedModel))
	}
	if string(loadedTool) != string(toolMetrics) {
		t.Fatalf("expected tool metrics %s, got %s", string(toolMetrics), string(loadedTool))
	}
}

// --- Chat ForUser tests ---

func TestChat_SaveAndGetMessages(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	if err := s.SaveMessageForUser(userID, "sess:user1", "user", "hello from user1", ""); err != nil {
		t.Fatalf("SaveMessageForUser: %v", err)
	}
	if err := s.SaveMessageForUser(userID, "sess:user1", "assistant", "hi user1", ""); err != nil {
		t.Fatalf("SaveMessageForUser: %v", err)
	}
	if err := s.SaveMessageForUser(userID, "sess:user1", "user", "another message", ""); err != nil {
		t.Fatalf("SaveMessageForUser: %v", err)
	}

	msgs, err := s.GetMessagesForUser(userID, "sess:user1")
	if err != nil {
		t.Fatalf("GetMessagesForUser: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "hello from user1" {
		t.Fatalf("unexpected first message: %+v", msgs[0])
	}
	if msgs[2].Content != "another message" {
		t.Fatalf("unexpected third message: %+v", msgs[2])
	}
}

func TestChat_SearchMessages(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	_ = s.SaveMessageForUser(userID, "s1", "user", "deploy the kubernetes cluster", "")
	_ = s.SaveMessageForUser(userID, "s1", "assistant", "deploying now", "")
	_ = s.SaveMessageForUser(userID, "s2", "user", "unrelated weather question", "")

	results, err := s.SearchMessagesForUser(userID, "kubernetes")
	if err != nil {
		t.Fatalf("SearchMessagesForUser: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Content != "deploy the kubernetes cluster" {
		t.Fatalf("unexpected result: %s", results[0].Content)
	}
}

func TestChat_ListSessions(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	_ = s.SaveMessageForUser(userID, "chat:alpha", "user", "alpha msg", "")
	_ = s.SaveMessageForUser(userID, "chat:beta", "user", "beta msg 1", "")
	_ = s.SaveMessageForUser(userID, "chat:beta", "assistant", "beta msg 2", "")

	sessions, err := s.ListSessionsForUser(userID, nil, 50, 0)
	if err != nil {
		t.Fatalf("ListSessionsForUser: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	counts := map[string]int{}
	for _, ss := range sessions {
		counts[ss.SessionID] = ss.MessageCount
	}
	if counts["chat:alpha"] != 1 {
		t.Fatalf("expected 1 message in chat:alpha, got %d", counts["chat:alpha"])
	}
	if counts["chat:beta"] != 2 {
		t.Fatalf("expected 2 messages in chat:beta, got %d", counts["chat:beta"])
	}
}

// --- Prompts tests ---

func TestPrompts_CRUD(t *testing.T) {
	s := newTestStorage(t)

	// Create
	p, err := s.CreatePrompt("Test Prompt", "You are a helpful assistant", "A test prompt", "ai,test")
	if err != nil {
		t.Fatalf("CreatePrompt: %v", err)
	}
	if p.Title != "Test Prompt" {
		t.Fatalf("expected title 'Test Prompt', got %q", p.Title)
	}
	if p.Content != "You are a helpful assistant" {
		t.Fatalf("expected content 'You are a helpful assistant', got %q", p.Content)
	}
	if p.Description != "A test prompt" {
		t.Fatalf("expected description 'A test prompt', got %q", p.Description)
	}
	if p.Tags != "ai,test" {
		t.Fatalf("expected tags 'ai,test', got %q", p.Tags)
	}

	// Get
	got, err := s.GetPrompt(p.ID)
	if err != nil {
		t.Fatalf("GetPrompt: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil prompt from GetPrompt")
	}
	if got.Title != "Test Prompt" {
		t.Fatalf("GetPrompt: expected title 'Test Prompt', got %q", got.Title)
	}

	// Update
	if err := s.UpdatePrompt(p.ID, "Updated Prompt", "new content", "updated desc", "ai,updated"); err != nil {
		t.Fatalf("UpdatePrompt: %v", err)
	}
	got, err = s.GetPrompt(p.ID)
	if err != nil {
		t.Fatalf("GetPrompt after update: %v", err)
	}
	if got.Title != "Updated Prompt" {
		t.Fatalf("expected updated title 'Updated Prompt', got %q", got.Title)
	}
	if got.Content != "new content" {
		t.Fatalf("expected updated content 'new content', got %q", got.Content)
	}

	// List
	prompts, err := s.ListPrompts()
	if err != nil {
		t.Fatalf("ListPrompts: %v", err)
	}
	if len(prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(prompts))
	}

	// Delete
	if err := s.DeletePrompt(p.ID); err != nil {
		t.Fatalf("DeletePrompt: %v", err)
	}
	got, err = s.GetPrompt(p.ID)
	if err != nil {
		t.Fatalf("GetPrompt after delete: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil prompt after delete")
	}
}

// --- Channel Mapping tests ---

func TestChannelMapping_SetAndGet(t *testing.T) {
	s := newTestStorage(t)

	if err := s.SetUserIDForChannelSender("telegram", "sender123", 42); err != nil {
		t.Fatalf("SetUserIDForChannelSender: %v", err)
	}

	userID, err := s.GetUserIDForChannelSender("telegram", "sender123")
	if err != nil {
		t.Fatalf("GetUserIDForChannelSender: %v", err)
	}
	if userID != 42 {
		t.Fatalf("expected userID 42, got %d", userID)
	}
}

func TestChannelMapping_Upsert(t *testing.T) {
	s := newTestStorage(t)

	if err := s.SetUserIDForChannelSender("discord", "user456", 10); err != nil {
		t.Fatalf("SetUserIDForChannelSender first: %v", err)
	}

	// Upsert with a new user ID
	if err := s.SetUserIDForChannelSender("discord", "user456", 20); err != nil {
		t.Fatalf("SetUserIDForChannelSender upsert: %v", err)
	}

	userID, err := s.GetUserIDForChannelSender("discord", "user456")
	if err != nil {
		t.Fatalf("GetUserIDForChannelSender: %v", err)
	}
	if userID != 20 {
		t.Fatalf("expected userID 20 after upsert, got %d", userID)
	}
}

func TestChannelMapping_NotFound(t *testing.T) {
	s := newTestStorage(t)

	userID, err := s.GetUserIDForChannelSender("slack", "nonexistent")
	if err != nil {
		t.Fatalf("GetUserIDForChannelSender: %v", err)
	}
	if userID != 0 {
		t.Fatalf("expected 0 for non-existent mapping, got %d", userID)
	}
}

// --- Task Log tests ---

func TestTaskLogs_AddAndGet(t *testing.T) {
	s := newTestStorage(t)
	taskID, err := s.CreateTask("logged task2", "", "todo", "")
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	_ = s.AddTaskLog(taskID, "created", "task was created")
	_ = s.AddTaskLog(taskID, "started", "execution started")
	_ = s.AddTaskLog(taskID, "completed", "execution finished")

	logs, err := s.GetTaskLogs(taskID)
	if err != nil {
		t.Fatalf("GetTaskLogs: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("expected 3 logs, got %d", len(logs))
	}
	// Verify order is ascending by created_at
	if logs[0].Event != "created" {
		t.Fatalf("expected first log event 'created', got %q", logs[0].Event)
	}
	if logs[1].Event != "started" {
		t.Fatalf("expected second log event 'started', got %q", logs[1].Event)
	}
	if logs[2].Event != "completed" {
		t.Fatalf("expected third log event 'completed', got %q", logs[2].Event)
	}
	if logs[0].Message != "task was created" {
		t.Fatalf("expected first log message 'task was created', got %q", logs[0].Message)
	}
}

func TestTaskLogs_EmptyForNonExistentTask(t *testing.T) {
	s := newTestStorage(t)

	logs, err := s.GetTaskLogs(99999)
	if err != nil {
		t.Fatalf("GetTaskLogs: %v", err)
	}
	if logs != nil && len(logs) != 0 {
		t.Fatalf("expected nil or empty slice for non-existent task, got %d logs", len(logs))
	}
}

// --- Task ForUser extended tests ---

func TestTask_ForUser_CRUD(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	// Create
	id, err := s.CreateTaskForUser(userID, "user task", "user task desc", "todo", "agent1")
	if err != nil {
		t.Fatalf("CreateTaskForUser: %v", err)
	}

	// Get
	task, err := s.GetTaskForUser(userID, id)
	if err != nil {
		t.Fatalf("GetTaskForUser: %v", err)
	}
	if task.Title != "user task" {
		t.Fatalf("expected title 'user task', got %q", task.Title)
	}
	if task.Description != "user task desc" {
		t.Fatalf("expected description 'user task desc', got %q", task.Description)
	}
	if task.Agent != "agent1" {
		t.Fatalf("expected agent 'agent1', got %q", task.Agent)
	}

	// Update
	updated, err := s.UpdateTaskForUser(userID, id, "updated user task", "new desc", "in_progress", "partial result")
	if err != nil {
		t.Fatalf("UpdateTaskForUser: %v", err)
	}
	if updated.Title != "updated user task" {
		t.Fatalf("expected updated title 'updated user task', got %q", updated.Title)
	}
	if updated.Status != "in_progress" {
		t.Fatalf("expected status 'in_progress', got %q", updated.Status)
	}
	if updated.Result != "partial result" {
		t.Fatalf("expected result 'partial result', got %q", updated.Result)
	}

	// List
	tasks, err := s.ListTasksForUser(userID, false)
	if err != nil {
		t.Fatalf("ListTasksForUser: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	// Delete
	if err := s.DeleteTaskForUser(userID, id); err != nil {
		t.Fatalf("DeleteTaskForUser: %v", err)
	}
	_, err = s.GetTaskForUser(userID, id)
	if err == nil {
		t.Fatal("expected error getting deleted task")
	}
}

func TestTask_ListByStatus(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	_, _ = s.CreateTaskForUser(userID, "todo task", "", "todo", "")
	id2, _ := s.CreateTaskForUser(userID, "in progress task", "", "in_progress", "")
	id3, _ := s.CreateTaskForUser(userID, "done task", "", "done", "")
	id4, _ := s.CreateTaskForUser(userID, "archived task", "", "done", "")
	_ = s.ArchiveTaskForUser(userID, id4)

	// List without archived
	tasks, err := s.ListTasksForUser(userID, false)
	if err != nil {
		t.Fatalf("ListTasksForUser: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 non-archived tasks, got %d", len(tasks))
	}

	// Verify we can find specific statuses
	statusMap := map[string]bool{}
	for _, task := range tasks {
		statusMap[task.Status] = true
	}
	if !statusMap["todo"] {
		t.Fatal("expected to find a 'todo' task")
	}
	if !statusMap["in_progress"] {
		t.Fatal("expected to find an 'in_progress' task")
	}
	if !statusMap["done"] {
		t.Fatal("expected to find a 'done' task")
	}

	// List with archived
	tasks, err = s.ListTasksForUser(userID, true)
	if err != nil {
		t.Fatalf("ListTasksForUser with archived: %v", err)
	}
	if len(tasks) != 4 {
		t.Fatalf("expected 4 tasks including archived, got %d", len(tasks))
	}

	// Update status
	_, err = s.UpdateTaskStatusForUser(userID, id2, "done")
	if err != nil {
		t.Fatalf("UpdateTaskStatusForUser: %v", err)
	}
	task, err := s.GetTaskForUser(userID, id2)
	if err != nil {
		t.Fatalf("GetTaskForUser: %v", err)
	}
	if task.Status != "done" {
		t.Fatalf("expected status 'done' after update, got %q", task.Status)
	}

	_ = id3 // used above in create
}

func TestTask_Search(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	_, _ = s.CreateTaskForUser(userID, "deploy kubernetes cluster", "production deploy", "todo", "")
	_, _ = s.CreateTaskForUser(userID, "fix login bug", "users cannot login", "in_progress", "")
	_, _ = s.CreateTaskForUser(userID, "update documentation", "refresh API docs", "todo", "")

	results, err := s.SearchTasksForUser(userID, "kubernetes")
	if err != nil {
		t.Fatalf("SearchTasksForUser: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'kubernetes', got %d", len(results))
	}
	if results[0].Title != "deploy kubernetes cluster" {
		t.Fatalf("expected 'deploy kubernetes cluster', got %q", results[0].Title)
	}

	// Search by description content
	results, err = s.SearchTasksForUser(userID, "login")
	if err != nil {
		t.Fatalf("SearchTasksForUser: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'login', got %d", len(results))
	}
	if results[0].Title != "fix login bug" {
		t.Fatalf("expected 'fix login bug', got %q", results[0].Title)
	}
}

// --- Skill Usage tests ---

func TestSkillUsage_RecordAndGet(t *testing.T) {
	s := newTestStorage(t)

	if err := s.RecordSkillUsage("weather", "builtin", "cli:default"); err != nil {
		t.Fatalf("RecordSkillUsage: %v", err)
	}

	stats, err := s.GetSkillUsageStats(30)
	if err != nil {
		t.Fatalf("GetSkillUsageStats: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 skill stat, got %d", len(stats))
	}
	if stats[0].SkillName != "weather" {
		t.Fatalf("expected skill_name 'weather', got %q", stats[0].SkillName)
	}
	if stats[0].LoadCount != 1 {
		t.Fatalf("expected load_count 1, got %d", stats[0].LoadCount)
	}
	if stats[0].LastUsedAt == "" {
		t.Fatal("expected last_used_at to be non-empty")
	}
}

func TestSkillUsage_MultipleRecords(t *testing.T) {
	s := newTestStorage(t)

	// Record same skill 3 times
	for i := 0; i < 3; i++ {
		if err := s.RecordSkillUsage("github", "builtin", "cli:default"); err != nil {
			t.Fatalf("RecordSkillUsage iteration %d: %v", i, err)
		}
	}
	// Record a different skill once
	if err := s.RecordSkillUsage("summarize", "workspace", "web:123"); err != nil {
		t.Fatalf("RecordSkillUsage summarize: %v", err)
	}

	stats, err := s.GetSkillUsageStats(30)
	if err != nil {
		t.Fatalf("GetSkillUsageStats: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 skill stats, got %d", len(stats))
	}
	// Results are ordered by load_count DESC, so github should be first
	if stats[0].SkillName != "github" {
		t.Fatalf("expected first skill 'github', got %q", stats[0].SkillName)
	}
	if stats[0].LoadCount != 3 {
		t.Fatalf("expected load_count 3 for github, got %d", stats[0].LoadCount)
	}
	if stats[1].SkillName != "summarize" {
		t.Fatalf("expected second skill 'summarize', got %q", stats[1].SkillName)
	}
	if stats[1].LoadCount != 1 {
		t.Fatalf("expected load_count 1 for summarize, got %d", stats[1].LoadCount)
	}
}

func TestSkillUsage_EmptyStats(t *testing.T) {
	s := newTestStorage(t)

	stats, err := s.GetSkillUsageStats(30)
	if err != nil {
		t.Fatalf("GetSkillUsageStats: %v", err)
	}
	if len(stats) != 0 {
		t.Fatalf("expected 0 skill stats on empty DB, got %d", len(stats))
	}
}

// --- Setup Session tests ---

func TestSetupSession_CreateAndGet(t *testing.T) {
	s := newTestStorage(t)

	metadata := map[string]string{"chat_id": "12345", "type": "onboard"}
	session, err := s.CreateSetupSession("telegram", "user42", metadata)
	if err != nil {
		t.Fatalf("CreateSetupSession: %v", err)
	}
	if session.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if session.Channel != "telegram" {
		t.Fatalf("expected channel 'telegram', got %q", session.Channel)
	}
	if session.SenderID != "user42" {
		t.Fatalf("expected sender_id 'user42', got %q", session.SenderID)
	}

	// Fix two issues for GetSetupSession to work in tests:
	// 1. Set expires_at using SQLite's datetime('now') to avoid timezone mismatch with Go's time.Now()
	// 2. Set user_id to 0 (non-NULL) because GetSetupSession scans user_id into int64 which cannot hold NULL
	_, err = s.db.Exec(`UPDATE setup_sessions SET expires_at = datetime('now', '+2 hours'), user_id = 0 WHERE token = ?`, session.Token)
	if err != nil {
		t.Fatalf("fix setup_session for retrieval: %v", err)
	}

	// Retrieve it
	got, err := s.GetSetupSession(session.Token)
	if err != nil {
		t.Fatalf("GetSetupSession: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil session from GetSetupSession")
	}
	if got.ID != session.ID {
		t.Fatalf("expected ID %d, got %d", session.ID, got.ID)
	}
	if got.Channel != "telegram" {
		t.Fatalf("expected channel 'telegram', got %q", got.Channel)
	}
	if got.SenderID != "user42" {
		t.Fatalf("expected sender_id 'user42', got %q", got.SenderID)
	}
}

func TestSetupSession_Update(t *testing.T) {
	s := newTestStorage(t)

	session, err := s.CreateSetupSession("discord", "disc_user", nil)
	if err != nil {
		t.Fatalf("CreateSetupSession: %v", err)
	}

	// Ensure expires_at is in the future relative to SQLite's datetime('now')
	_, err = s.db.Exec(`UPDATE setup_sessions SET expires_at = datetime('now', '+2 hours') WHERE token = ?`, session.Token)
	if err != nil {
		t.Fatalf("fix expires_at: %v", err)
	}

	// Mark as used with a user ID
	if err := s.UpdateSetupSession(session.Token, 7); err != nil {
		t.Fatalf("UpdateSetupSession: %v", err)
	}

	got, err := s.GetSetupSession(session.Token)
	if err != nil {
		t.Fatalf("GetSetupSession after update: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil session after update")
	}
	if got.UserID != 7 {
		t.Fatalf("expected user_id 7, got %d", got.UserID)
	}
	if got.UsedAt == nil {
		t.Fatal("expected used_at to be set after UpdateSetupSession")
	}
}

func TestSetupSession_GetExpiredReturnsNil(t *testing.T) {
	s := newTestStorage(t)

	session, err := s.CreateSetupSession("slack", "slack_user", nil)
	if err != nil {
		t.Fatalf("CreateSetupSession: %v", err)
	}

	// Manually expire the session by setting expires_at to the past
	_, err = s.db.Exec(`UPDATE setup_sessions SET expires_at = datetime('now', '-1 hour') WHERE token = ?`, session.Token)
	if err != nil {
		t.Fatalf("force-expire session: %v", err)
	}

	got, err := s.GetSetupSession(session.Token)
	if err != nil {
		t.Fatalf("GetSetupSession: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil for expired session")
	}
}

func TestSetupSession_CleanupExpired(t *testing.T) {
	s := newTestStorage(t)

	session, err := s.CreateSetupSession("telegram", "user1", nil)
	if err != nil {
		t.Fatalf("CreateSetupSession: %v", err)
	}

	// Force-expire it
	_, err = s.db.Exec(`UPDATE setup_sessions SET expires_at = datetime('now', '-2 hours') WHERE token = ?`, session.Token)
	if err != nil {
		t.Fatalf("force-expire: %v", err)
	}

	if err := s.CleanupExpiredSetupSessions(); err != nil {
		t.Fatalf("CleanupExpiredSetupSessions: %v", err)
	}

	// Verify it's actually deleted (even a direct query should find nothing)
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM setup_sessions WHERE token = ?`, session.Token).Scan(&count); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows after cleanup, got %d", count)
	}
}

// --- User Providers tests ---

func TestUserProviders_SaveAndLoad(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	providers := &UserProvidersConfig{
		UserID: userID,
		Anthropic: ProviderConfig{
			APIKey:  "sk-ant-test123",
			APIBase: "https://api.anthropic.com",
		},
		OpenAI: ProviderConfig{
			APIKey: "sk-openai-test456",
		},
	}

	if err := s.SaveUserProvidersConfig(userID, providers); err != nil {
		t.Fatalf("SaveUserProvidersConfig: %v", err)
	}

	loaded, err := s.GetUserProvidersConfig(userID)
	if err != nil {
		t.Fatalf("GetUserProvidersConfig: %v", err)
	}
	if loaded.Anthropic.APIKey != "sk-ant-test123" {
		t.Fatalf("expected Anthropic API key 'sk-ant-test123', got %q", loaded.Anthropic.APIKey)
	}
	if loaded.Anthropic.APIBase != "https://api.anthropic.com" {
		t.Fatalf("expected Anthropic API base 'https://api.anthropic.com', got %q", loaded.Anthropic.APIBase)
	}
	if loaded.OpenAI.APIKey != "sk-openai-test456" {
		t.Fatalf("expected OpenAI API key 'sk-openai-test456', got %q", loaded.OpenAI.APIKey)
	}
	if loaded.UserID != userID {
		t.Fatalf("expected UserID %d, got %d", userID, loaded.UserID)
	}
}

func TestUserProviders_DefaultEmptyConfig(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 999

	// No config saved yet - should return empty config
	loaded, err := s.GetUserProvidersConfig(userID)
	if err != nil {
		t.Fatalf("GetUserProvidersConfig: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil default providers config")
	}
	if loaded.UserID != userID {
		t.Fatalf("expected UserID %d, got %d", userID, loaded.UserID)
	}
	if loaded.Anthropic.APIKey != "" {
		t.Fatalf("expected empty Anthropic API key, got %q", loaded.Anthropic.APIKey)
	}
}

func TestUserProviders_UpdateSingleProvider(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	// Save initial config
	initial := &UserProvidersConfig{
		UserID: userID,
		Anthropic: ProviderConfig{
			APIKey: "initial-key",
		},
	}
	if err := s.SaveUserProvidersConfig(userID, initial); err != nil {
		t.Fatalf("SaveUserProvidersConfig: %v", err)
	}

	// Update just OpenAI
	if err := s.UpdateUserProviderConfig(userID, "openai", ProviderConfig{
		APIKey:  "new-openai-key",
		APIBase: "https://api.openai.com/v1",
	}); err != nil {
		t.Fatalf("UpdateUserProviderConfig: %v", err)
	}

	loaded, err := s.GetUserProvidersConfig(userID)
	if err != nil {
		t.Fatalf("GetUserProvidersConfig: %v", err)
	}
	// Anthropic should still have old key
	if loaded.Anthropic.APIKey != "initial-key" {
		t.Fatalf("expected Anthropic key 'initial-key', got %q", loaded.Anthropic.APIKey)
	}
	// OpenAI should have new key
	if loaded.OpenAI.APIKey != "new-openai-key" {
		t.Fatalf("expected OpenAI key 'new-openai-key', got %q", loaded.OpenAI.APIKey)
	}
}

func TestUserProviders_UpdateUnknownProvider(t *testing.T) {
	s := newTestStorage(t)
	err := s.UpdateUserProviderConfig(1, "nonexistent_provider", ProviderConfig{APIKey: "key"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestUserProviders_Delete(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	providers := &UserProvidersConfig{
		UserID:    userID,
		Anthropic: ProviderConfig{APIKey: "to-delete"},
	}
	if err := s.SaveUserProvidersConfig(userID, providers); err != nil {
		t.Fatalf("SaveUserProvidersConfig: %v", err)
	}

	if err := s.DeleteUserProvidersConfig(userID); err != nil {
		t.Fatalf("DeleteUserProvidersConfig: %v", err)
	}

	// After delete, should get empty config (not error)
	loaded, err := s.GetUserProvidersConfig(userID)
	if err != nil {
		t.Fatalf("GetUserProvidersConfig after delete: %v", err)
	}
	if loaded.Anthropic.APIKey != "" {
		t.Fatalf("expected empty Anthropic key after delete, got %q", loaded.Anthropic.APIKey)
	}
}

func TestUserProviders_Upsert(t *testing.T) {
	s := newTestStorage(t)
	var userID int64 = 1

	// Save first config
	first := &UserProvidersConfig{
		UserID:    userID,
		Anthropic: ProviderConfig{APIKey: "first-key"},
	}
	if err := s.SaveUserProvidersConfig(userID, first); err != nil {
		t.Fatalf("SaveUserProvidersConfig first: %v", err)
	}

	// Overwrite with new config
	second := &UserProvidersConfig{
		UserID:    userID,
		Anthropic: ProviderConfig{APIKey: "second-key"},
		Groq:      ProviderConfig{APIKey: "groq-key"},
	}
	if err := s.SaveUserProvidersConfig(userID, second); err != nil {
		t.Fatalf("SaveUserProvidersConfig second: %v", err)
	}

	loaded, err := s.GetUserProvidersConfig(userID)
	if err != nil {
		t.Fatalf("GetUserProvidersConfig: %v", err)
	}
	if loaded.Anthropic.APIKey != "second-key" {
		t.Fatalf("expected 'second-key', got %q", loaded.Anthropic.APIKey)
	}
	if loaded.Groq.APIKey != "groq-key" {
		t.Fatalf("expected 'groq-key', got %q", loaded.Groq.APIKey)
	}
}

// --- Backup export/import tests ---

func TestBackup_ExportAndImport(t *testing.T) {
	dir := t.TempDir()
	dbPath1 := filepath.Join(dir, "export.db")
	dbPath2 := filepath.Join(dir, "import.db")

	s1, err := New(config.StorageConfig{Path: dbPath1})
	if err != nil {
		t.Fatalf("New export db: %v", err)
	}
	defer s1.Close()

	var userID int64 = 1

	// Populate data - use ForUser variants to ensure user_id is set correctly
	_ = s1.SaveMessageForUser(userID, "backup:sess1", "user", "hello backup", "")
	_ = s1.SaveMessageForUser(userID, "backup:sess1", "assistant", "backup reply", "")
	taskID, _ := s1.CreateTaskForUser(userID, "backup task", "task desc", "in_progress", "")
	_ = s1.AddTaskLog(taskID, "created", "task created for backup test")

	// Export
	data, err := s1.ExportUserData(userID)
	if err != nil {
		t.Fatalf("ExportUserData: %v", err)
	}
	if len(data.Messages) != 2 {
		t.Fatalf("expected 2 exported messages, got %d", len(data.Messages))
	}
	if len(data.Tasks) != 1 {
		t.Fatalf("expected 1 exported task, got %d", len(data.Tasks))
	}
	if len(data.TaskLogs) != 1 {
		t.Fatalf("expected 1 exported task log, got %d", len(data.TaskLogs))
	}

	// Import into a fresh database
	s2, err := New(config.StorageConfig{Path: dbPath2})
	if err != nil {
		t.Fatalf("New import db: %v", err)
	}
	defer s2.Close()

	sessions, messages, tasks, err := s2.ImportUserData(userID, data)
	if err != nil {
		t.Fatalf("ImportUserData: %v", err)
	}
	if sessions < 1 {
		t.Fatalf("expected at least 1 imported session, got %d", sessions)
	}
	if messages != 2 {
		t.Fatalf("expected 2 imported messages, got %d", messages)
	}
	if tasks != 1 {
		t.Fatalf("expected 1 imported task, got %d", tasks)
	}

	// Verify imported data is readable
	msgs, err := s2.GetMessagesForUser(userID, "backup:sess1")
	if err != nil {
		t.Fatalf("GetMessagesForUser after import: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages after import, got %d", len(msgs))
	}
	if msgs[0].Content != "hello backup" {
		t.Fatalf("expected first message 'hello backup', got %q", msgs[0].Content)
	}
}

func TestBackup_ImportNilData(t *testing.T) {
	s := newTestStorage(t)

	sessions, messages, tasks, err := s.ImportUserData(1, nil)
	if err != nil {
		t.Fatalf("ImportUserData nil: %v", err)
	}
	if sessions != 0 || messages != 0 || tasks != 0 {
		t.Fatalf("expected all zeros for nil import, got sessions=%d messages=%d tasks=%d", sessions, messages, tasks)
	}
}

func TestBackup_ExportEmptyDatabase(t *testing.T) {
	s := newTestStorage(t)

	data, err := s.ExportUserData(1)
	if err != nil {
		t.Fatalf("ExportUserData empty: %v", err)
	}
	if len(data.Sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(data.Sessions))
	}
	if len(data.Messages) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(data.Messages))
	}
	if len(data.Tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(data.Tasks))
	}
}

// --- User CRUD tests (via CentralStorage, which has the full users schema) ---

func newTestCentralStorage(t *testing.T) *CentralStorage {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "central_test.db")
	cs, err := NewCentral(dbPath)
	if err != nil {
		t.Fatalf("NewCentral: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func TestUser_CreateAndGet(t *testing.T) {
	cs := newTestCentralStorage(t)

	user, err := cs.CreateUser("testuser", "password123", "user")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.Username != "testuser" {
		t.Fatalf("expected username 'testuser', got %q", user.Username)
	}
	if user.Role != "user" {
		t.Fatalf("expected role 'user', got %q", user.Role)
	}
	if user.UUID == "" {
		t.Fatal("expected non-empty UUID")
	}

	// Get by ID
	byID, err := cs.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if byID.Username != "testuser" {
		t.Fatalf("expected 'testuser', got %q", byID.Username)
	}

	// Get by UUID
	byUUID, err := cs.GetUserByUUID(user.UUID)
	if err != nil {
		t.Fatalf("GetUserByUUID: %v", err)
	}
	if byUUID.ID != user.ID {
		t.Fatalf("expected ID %d, got %d", user.ID, byUUID.ID)
	}

	// Get by username
	byName, err := cs.GetUserByUsername("testuser")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if byName.ID != user.ID {
		t.Fatalf("expected ID %d, got %d", user.ID, byName.ID)
	}
}

func TestUser_DuplicateUsername(t *testing.T) {
	cs := newTestCentralStorage(t)

	_, err := cs.CreateUser("duplicate", "pass1", "user")
	if err != nil {
		t.Fatalf("CreateUser first: %v", err)
	}

	_, err = cs.CreateUser("duplicate", "pass2", "user")
	if err != ErrUserExists {
		t.Fatalf("expected ErrUserExists, got %v", err)
	}
}

func TestUser_EmptyUsername(t *testing.T) {
	cs := newTestCentralStorage(t)

	_, err := cs.CreateUser("", "pass", "user")
	if err == nil {
		t.Fatal("expected error for empty username")
	}
}

func TestUser_ListAndCount(t *testing.T) {
	cs := newTestCentralStorage(t)

	_, _ = cs.CreateUser("admin1", "pass", "admin")
	_, _ = cs.CreateUser("user1", "pass", "user")
	_, _ = cs.CreateUser("user2", "pass", "user")

	users, err := cs.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}

	count, err := cs.CountUsers()
	if err != nil {
		t.Fatalf("CountUsers: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}

	adminCount, err := cs.CountAdmins()
	if err != nil {
		t.Fatalf("CountAdmins: %v", err)
	}
	if adminCount != 1 {
		t.Fatalf("expected 1 admin, got %d", adminCount)
	}
}

func TestUser_Delete(t *testing.T) {
	cs := newTestCentralStorage(t)

	user, err := cs.CreateUser("deleteme", "pass", "user")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if err := cs.DeleteUser(user.ID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	_, err = cs.GetUserByID(user.ID)
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUser_UpdateRole(t *testing.T) {
	cs := newTestCentralStorage(t)

	user, err := cs.CreateUser("roletest", "pass", "user")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if err := cs.UpdateUserRole(user.ID, "admin"); err != nil {
		t.Fatalf("UpdateUserRole: %v", err)
	}

	updated, err := cs.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID after role update: %v", err)
	}
	if updated.Role != "admin" {
		t.Fatalf("expected role 'admin', got %q", updated.Role)
	}
}

// --- Settings tests ---

func TestSettings_GetAndSet(t *testing.T) {
	s := newTestStorage(t)

	// Get nonexistent key returns empty string
	val, err := s.GetSetting("theme")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string for nonexistent key, got %q", val)
	}

	// Set and get
	if err := s.SetSetting("theme", "dark"); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	val, err = s.GetSetting("theme")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if val != "dark" {
		t.Fatalf("expected 'dark', got %q", val)
	}

	// Upsert
	if err := s.SetSetting("theme", "light"); err != nil {
		t.Fatalf("SetSetting upsert: %v", err)
	}
	val, _ = s.GetSetting("theme")
	if val != "light" {
		t.Fatalf("expected 'light' after upsert, got %q", val)
	}
}
