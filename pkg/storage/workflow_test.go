package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
)

func TestListWorkflowsLegacyModeFiltersByUser(t *testing.T) {
	s := newTestStorage(t)
	const (
		userAID int64 = 1
		userBID int64 = 2
	)

	steps := json.RawMessage(`[{"name":"step-a","action":"echo"}]`)
	params := json.RawMessage(`[{"name":"message","type":"string"}]`)

	workflowAID, err := s.CreateWorkflow(userAID, "workflow-a", "owned by user A", steps, params, nil)
	if err != nil {
		t.Fatalf("CreateWorkflow userA: %v", err)
	}
	workflowBID, err := s.CreateWorkflow(userBID, "workflow-b", "owned by user B", steps, params, nil)
	if err != nil {
		t.Fatalf("CreateWorkflow userB: %v", err)
	}

	workflows, err := s.ListWorkflows(userAID)
	if err != nil {
		t.Fatalf("ListWorkflows userA: %v", err)
	}
	if len(workflows) != 1 {
		t.Fatalf("expected 1 workflow for user A, got %d", len(workflows))
	}
	if workflows[0].ID != workflowAID {
		t.Fatalf("expected workflow %d for user A, got %d", workflowAID, workflows[0].ID)
	}
	if workflows[0].ID == workflowBID {
		t.Fatalf("expected user B workflow %d to be filtered out", workflowBID)
	}
}

func TestWorkflowMigrationIdempotentAndNonDestructive(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy-workflows.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	legacyQueries := []string{
		`CREATE TABLE workflows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			enabled BOOLEAN NOT NULL DEFAULT 1,
			steps TEXT NOT NULL DEFAULT '[]',
			parameters TEXT NOT NULL DEFAULT '[]',
			schedule TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE workflow_runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workflow_id INTEGER NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
			status TEXT NOT NULL DEFAULT 'running',
			results TEXT NOT NULL DEFAULT '[]',
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			finished_at DATETIME
		);`,
		`CREATE INDEX idx_workflow_runs_workflow_id ON workflow_runs(workflow_id);`,
		`CREATE TABLE workflow_approvals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workflow_id INTEGER NOT NULL,
			run_id INTEGER NOT NULL,
			step_id TEXT NOT NULL,
			message TEXT,
			status TEXT DEFAULT 'pending',
			approved_by TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			resolved_at DATETIME
		);`,
		`INSERT INTO workflows (name, description, enabled, steps, parameters, schedule) VALUES ('legacy-workflow', 'before migration', 1, '[]', '[]', NULL);`,
		`INSERT INTO workflow_runs (workflow_id, status, results) VALUES (1, 'completed', '[]');`,
		`INSERT INTO workflow_approvals (workflow_id, run_id, step_id, message, status) VALUES (1, 1, 'approve-1', 'review me', 'pending');`,
	}

	for _, query := range legacyQueries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("seed legacy workflow schema: %v", err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	s, err := New(config.StorageConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer func() { _ = s.Close() }()

	if err := s.migrateWorkflows(); err != nil {
		t.Fatalf("second migrateWorkflows: %v", err)
	}
	if err := s.migrateWorkflows(); err != nil {
		t.Fatalf("third migrateWorkflows: %v", err)
	}

	for _, table := range []string{"workflows", "workflow_runs", "workflow_approvals"} {
		if got := workflowTableCount(t, s, table); got != 1 {
			t.Fatalf("expected 1 row in %s after repeated migration, got %d", table, got)
		}
		if !workflowHasColumn(t, s, table, "user_id") {
			t.Fatalf("expected %s.user_id column to exist after migration", table)
		}
		var userID int64
		query := fmt.Sprintf("SELECT user_id FROM %s LIMIT 1", table)
		if err := s.db.QueryRow(query).Scan(&userID); err != nil {
			t.Fatalf("read %s.user_id: %v", table, err)
		}
		if userID != 0 {
			t.Fatalf("expected migrated %s row to keep default user_id=0, got %d", table, userID)
		}
	}
}

func workflowTableCount(t *testing.T, s *Storage, table string) int {
	t.Helper()
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if err := s.db.QueryRow(query).Scan(&count); err != nil {
		t.Fatalf("count rows in %s: %v", table, err)
	}
	return count
}

func workflowHasColumn(t *testing.T, s *Storage, table, column string) bool {
	t.Helper()
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name = ?", table)
	if err := s.db.QueryRow(query, column).Scan(&count); err != nil {
		t.Fatalf("check column %s.%s: %v", table, column, err)
	}
	return count > 0
}
