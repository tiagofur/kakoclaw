package storage

import (
	"encoding/json"
	"fmt"
	"time"
)

// Workflow represents a saved automation pipeline with ordered steps.
type Workflow struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled"`
	Steps       json.RawMessage `json:"steps"` // JSON array of WorkflowStep
	Schedule    json.RawMessage `json:"schedule,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// WorkflowRun records the result of a workflow execution.
type WorkflowRun struct {
	ID         int64           `json:"id"`
	WorkflowID int64           `json:"workflow_id"`
	Status     string          `json:"status"` // "running", "completed", "failed"
	Results    json.RawMessage `json:"results"`
	StartedAt  time.Time       `json:"started_at"`
	FinishedAt *time.Time      `json:"finished_at,omitempty"`
}

func (s *Storage) migrateWorkflows() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS workflows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			enabled BOOLEAN NOT NULL DEFAULT 1,
			steps TEXT NOT NULL DEFAULT '[]',
			schedule TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS workflow_runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workflow_id INTEGER NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
			status TEXT NOT NULL DEFAULT 'running',
			results TEXT NOT NULL DEFAULT '[]',
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			finished_at DATETIME
		);`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow_id ON workflow_runs(workflow_id);`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("workflow migration: %w", err)
		}
	}
	return nil
}

// CreateWorkflow inserts a new workflow and returns its ID.
func (s *Storage) CreateWorkflow(name, description string, steps, schedule json.RawMessage) (int64, error) {
	if steps == nil {
		steps = json.RawMessage("[]")
	}
	query := `INSERT INTO workflows (name, description, steps, schedule) VALUES (?, ?, ?, ?)`
	var scheduleStr *string
	if schedule != nil && string(schedule) != "null" {
		sv := string(schedule)
		scheduleStr = &sv
	}
	result, err := s.db.Exec(query, name, description, string(steps), scheduleStr)
	if err != nil {
		return 0, fmt.Errorf("creating workflow: %w", err)
	}
	return result.LastInsertId()
}

// GetWorkflow returns a single workflow by ID.
func (s *Storage) GetWorkflow(id int64) (*Workflow, error) {
	query := `SELECT id, name, COALESCE(description, ''), enabled, COALESCE(steps, '[]'), schedule, created_at, updated_at FROM workflows WHERE id = ?`
	var w Workflow
	var stepsStr, scheduleStr string
	var schedulePtr *string
	err := s.db.QueryRow(query, id).Scan(&w.ID, &w.Name, &w.Description, &w.Enabled, &stepsStr, &schedulePtr, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting workflow: %w", err)
	}
	w.Steps = json.RawMessage(stepsStr)
	if schedulePtr != nil {
		scheduleStr = *schedulePtr
		w.Schedule = json.RawMessage(scheduleStr)
	}
	return &w, nil
}

// ListWorkflows returns all workflows.
func (s *Storage) ListWorkflows() ([]Workflow, error) {
	query := `SELECT id, name, COALESCE(description, ''), enabled, COALESCE(steps, '[]'), schedule, created_at, updated_at FROM workflows ORDER BY updated_at DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("listing workflows: %w", err)
	}
	defer rows.Close()

	workflows := make([]Workflow, 0)
	for rows.Next() {
		var w Workflow
		var stepsStr string
		var schedulePtr *string
		if err := rows.Scan(&w.ID, &w.Name, &w.Description, &w.Enabled, &stepsStr, &schedulePtr, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning workflow: %w", err)
		}
		w.Steps = json.RawMessage(stepsStr)
		if schedulePtr != nil {
			w.Schedule = json.RawMessage(*schedulePtr)
		}
		workflows = append(workflows, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating workflows: %w", err)
	}
	return workflows, nil
}

// UpdateWorkflow updates a workflow's fields.
func (s *Storage) UpdateWorkflow(id int64, name, description string, enabled bool, steps, schedule json.RawMessage) (*Workflow, error) {
	if steps == nil {
		steps = json.RawMessage("[]")
	}
	var scheduleStr *string
	if schedule != nil && string(schedule) != "null" {
		sv := string(schedule)
		scheduleStr = &sv
	}
	query := `UPDATE workflows SET name = ?, description = ?, enabled = ?, steps = ?, schedule = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	result, err := s.db.Exec(query, name, description, enabled, string(steps), scheduleStr, id)
	if err != nil {
		return nil, fmt.Errorf("updating workflow: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return nil, fmt.Errorf("workflow not found")
	}
	return s.GetWorkflow(id)
}

// DeleteWorkflow removes a workflow.
func (s *Storage) DeleteWorkflow(id int64) error {
	result, err := s.db.Exec(`DELETE FROM workflows WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting workflow: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("workflow not found")
	}
	return nil
}

// CreateWorkflowRun inserts a new run record.
func (s *Storage) CreateWorkflowRun(workflowID int64) (int64, error) {
	result, err := s.db.Exec(`INSERT INTO workflow_runs (workflow_id, status) VALUES (?, 'running')`, workflowID)
	if err != nil {
		return 0, fmt.Errorf("creating workflow run: %w", err)
	}
	return result.LastInsertId()
}

// UpdateWorkflowRun updates a run's status and results.
func (s *Storage) UpdateWorkflowRun(id int64, status string, results json.RawMessage) error {
	var query string
	if status == "completed" || status == "completed_with_errors" || status == "failed" {
		query = `UPDATE workflow_runs SET status = ?, results = ?, finished_at = CURRENT_TIMESTAMP WHERE id = ?`
	} else {
		query = `UPDATE workflow_runs SET status = ?, results = ? WHERE id = ?`
	}
	_, err := s.db.Exec(query, status, string(results), id)
	if err != nil {
		return fmt.Errorf("updating workflow run: %w", err)
	}
	return nil
}

// ListWorkflowRuns returns recent runs for a workflow.
func (s *Storage) ListWorkflowRuns(workflowID int64, limit int) ([]WorkflowRun, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(`SELECT id, workflow_id, status, COALESCE(results, '[]'), started_at, finished_at FROM workflow_runs WHERE workflow_id = ? ORDER BY started_at DESC LIMIT ?`, workflowID, limit)
	if err != nil {
		return nil, fmt.Errorf("listing workflow runs: %w", err)
	}
	defer rows.Close()

	runs := make([]WorkflowRun, 0)
	for rows.Next() {
		var r WorkflowRun
		var resultsStr string
		if err := rows.Scan(&r.ID, &r.WorkflowID, &r.Status, &resultsStr, &r.StartedAt, &r.FinishedAt); err != nil {
			return nil, fmt.Errorf("scanning workflow run: %w", err)
		}
		r.Results = json.RawMessage(resultsStr)
		runs = append(runs, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating workflow runs: %w", err)
	}
	return runs, nil
}
