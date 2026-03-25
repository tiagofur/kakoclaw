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
	Steps       json.RawMessage `json:"steps"`      // JSON array of WorkflowStep
	Parameters  json.RawMessage `json:"parameters"` // JSON array of WorkflowParameter
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

// WorkflowApproval tracks pending approval gates in workflow execution.
type WorkflowApproval struct {
	ID         int64      `json:"id"`
	WorkflowID int64      `json:"workflow_id"`
	RunID      int64      `json:"run_id"`
	StepID     string     `json:"step_id"`
	Message    string     `json:"message"`
	Status     string     `json:"status"` // "pending", "approved", "rejected"
	ApprovedBy string     `json:"approved_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

func (s *Storage) migrateWorkflows() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS workflows (
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
		`CREATE TABLE IF NOT EXISTS workflow_runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workflow_id INTEGER NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
			status TEXT NOT NULL DEFAULT 'running',
			results TEXT NOT NULL DEFAULT '[]',
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			finished_at DATETIME
		);`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow_id ON workflow_runs(workflow_id);`,
		`CREATE TABLE IF NOT EXISTS workflow_approvals (
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
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("workflow migration: %w", err)
		}
	}

	// Add parameters column if it doesn't exist (for existing databases)
	var columnExists int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('workflows') WHERE name='parameters'`).Scan(&columnExists)
	if err == nil && columnExists == 0 {
		if _, err := s.db.Exec(`ALTER TABLE workflows ADD COLUMN parameters TEXT NOT NULL DEFAULT '[]'`); err != nil {
			return fmt.Errorf("adding parameters column: %w", err)
		}
	}
	if s.isUserDB {
		return nil
	}

	workflowTables := []struct {
		table string
		index string
	}{
		{table: "workflows", index: "idx_workflows_user_id"},
		{table: "workflow_runs", index: "idx_workflow_runs_user_id"},
		{table: "workflow_approvals", index: "idx_workflow_approvals_user_id"},
	}

	for _, item := range workflowTables {
		var userIDColumnExists int
		query := fmt.Sprintf(`SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name='user_id'`, item.table)
		if err := s.db.QueryRow(query).Scan(&userIDColumnExists); err != nil {
			return fmt.Errorf("checking %s.user_id column: %w", item.table, err)
		}
		if userIDColumnExists == 0 {
			alter := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0`, item.table)
			if _, err := s.db.Exec(alter); err != nil {
				return fmt.Errorf("adding user_id column to %s: %w", item.table, err)
			}
		}
		index := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %s ON %s(user_id)`, item.index, item.table)
		if _, err := s.db.Exec(index); err != nil {
			return fmt.Errorf("creating %s: %w", item.index, err)
		}
	}

	return nil
}

// CreateWorkflow inserts a new workflow and returns its ID.
func (s *Storage) CreateWorkflow(userID int64, name, description string, steps, parameters, schedule json.RawMessage) (int64, error) {
	if steps == nil {
		steps = json.RawMessage("[]")
	}
	if parameters == nil {
		parameters = json.RawMessage("[]")
	}
	var query string
	var args []interface{}
	var scheduleStr *string
	if schedule != nil && string(schedule) != "null" {
		sv := string(schedule)
		scheduleStr = &sv
	}
	if s.isUserDB {
		query = `INSERT INTO workflows (name, description, steps, parameters, schedule) VALUES (?, ?, ?, ?, ?)`
		args = []interface{}{name, description, string(steps), string(parameters), scheduleStr}
	} else {
		uid := normalizeUserKey(userID)
		query = `INSERT INTO workflows (user_id, name, description, steps, parameters, schedule) VALUES (?, ?, ?, ?, ?, ?)`
		args = []interface{}{uid, name, description, string(steps), string(parameters), scheduleStr}
	}
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("creating workflow: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting insert id: %w", err)
	}
	return id, nil
}

// GetWorkflow returns a single workflow by ID.
func (s *Storage) GetWorkflow(id, userID int64) (*Workflow, error) {
	var query string
	var args []interface{}
	if s.isUserDB {
		query = `SELECT id, name, COALESCE(description, ''), enabled, COALESCE(steps, '[]'), COALESCE(parameters, '[]'), schedule, created_at, updated_at FROM workflows WHERE id = ?`
		args = []interface{}{id}
	} else {
		uid := normalizeUserKey(userID)
		query = `SELECT id, name, COALESCE(description, ''), enabled, COALESCE(steps, '[]'), COALESCE(parameters, '[]'), schedule, created_at, updated_at FROM workflows WHERE id = ? AND user_id = ?`
		args = []interface{}{id, uid}
	}
	var w Workflow
	var stepsStr, parametersStr, scheduleStr string
	var schedulePtr *string
	err := s.db.QueryRow(query, args...).Scan(&w.ID, &w.Name, &w.Description, &w.Enabled, &stepsStr, &parametersStr, &schedulePtr, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting workflow: %w", err)
	}
	w.Steps = json.RawMessage(stepsStr)
	w.Parameters = json.RawMessage(parametersStr)
	if schedulePtr != nil {
		scheduleStr = *schedulePtr
		w.Schedule = json.RawMessage(scheduleStr)
	}
	return &w, nil
}

// ListWorkflows returns all workflows.
func (s *Storage) ListWorkflows(userID int64) ([]Workflow, error) {
	var query string
	var args []interface{}
	if s.isUserDB {
		query = `SELECT id, name, COALESCE(description, ''), enabled, COALESCE(steps, '[]'), COALESCE(parameters, '[]'), schedule, created_at, updated_at FROM workflows ORDER BY updated_at DESC`
	} else {
		uid := normalizeUserKey(userID)
		query = `SELECT id, name, COALESCE(description, ''), enabled, COALESCE(steps, '[]'), COALESCE(parameters, '[]'), schedule, created_at, updated_at FROM workflows WHERE user_id = ? ORDER BY updated_at DESC`
		args = []interface{}{uid}
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing workflows: %w", err)
	}
	defer rows.Close()

	workflows := make([]Workflow, 0)
	for rows.Next() {
		var w Workflow
		var stepsStr, parametersStr string
		var schedulePtr *string
		if err := rows.Scan(&w.ID, &w.Name, &w.Description, &w.Enabled, &stepsStr, &parametersStr, &schedulePtr, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning workflow: %w", err)
		}
		w.Steps = json.RawMessage(stepsStr)
		w.Parameters = json.RawMessage(parametersStr)
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
func (s *Storage) UpdateWorkflow(id, userID int64, name, description string, enabled bool, steps, parameters, schedule json.RawMessage) (*Workflow, error) {
	if steps == nil {
		steps = json.RawMessage("[]")
	}
	if parameters == nil {
		parameters = json.RawMessage("[]")
	}
	var scheduleStr *string
	if schedule != nil && string(schedule) != "null" {
		sv := string(schedule)
		scheduleStr = &sv
	}
	var query string
	var args []interface{}
	if s.isUserDB {
		query = `UPDATE workflows SET name = ?, description = ?, enabled = ?, steps = ?, parameters = ?, schedule = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
		args = []interface{}{name, description, enabled, string(steps), string(parameters), scheduleStr, id}
	} else {
		uid := normalizeUserKey(userID)
		query = `UPDATE workflows SET name = ?, description = ?, enabled = ?, steps = ?, parameters = ?, schedule = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
		args = []interface{}{name, description, enabled, string(steps), string(parameters), scheduleStr, id, uid}
	}
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("updating workflow: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("getting rows affected: %w", err)
	}
	if n == 0 {
		return nil, fmt.Errorf("workflow not found")
	}
	return s.GetWorkflow(id, userID)
}

// DeleteWorkflow removes a workflow.
func (s *Storage) DeleteWorkflow(id, userID int64) error {
	var query string
	var args []interface{}
	if s.isUserDB {
		query = `DELETE FROM workflows WHERE id = ?`
		args = []interface{}{id}
	} else {
		uid := normalizeUserKey(userID)
		query = `DELETE FROM workflows WHERE id = ? AND user_id = ?`
		args = []interface{}{id, uid}
	}
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("deleting workflow: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}
	if n == 0 {
		if !s.isUserDB {
			return nil
		}
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
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting insert id: %w", err)
	}
	return id, nil
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

// CreateWorkflowApproval inserts a pending approval record.
func (s *Storage) CreateWorkflowApproval(workflowID, runID int64, stepID, message string) (int64, error) {
	result, err := s.db.Exec(
		`INSERT INTO workflow_approvals (workflow_id, run_id, step_id, message) VALUES (?, ?, ?, ?)`,
		workflowID, runID, stepID, message,
	)
	if err != nil {
		return 0, fmt.Errorf("creating approval: %w", err)
	}
	return result.LastInsertId()
}

// ResolveWorkflowApproval approves or rejects a pending approval.
func (s *Storage) ResolveWorkflowApproval(id int64, status, approvedBy string) error {
	result, err := s.db.Exec(
		`UPDATE workflow_approvals SET status = ?, approved_by = ?, resolved_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'pending'`,
		status, approvedBy, id,
	)
	if err != nil {
		return fmt.Errorf("resolving approval: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("approval not found or already resolved")
	}
	return nil
}

// GetPendingApprovals returns all pending workflow approvals.
func (s *Storage) GetPendingApprovals() ([]WorkflowApproval, error) {
	rows, err := s.db.Query(
		`SELECT id, workflow_id, run_id, step_id, COALESCE(message, ''), status, COALESCE(approved_by, ''), created_at, resolved_at
		FROM workflow_approvals WHERE status = 'pending' ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing approvals: %w", err)
	}
	defer rows.Close()

	approvals := make([]WorkflowApproval, 0)
	for rows.Next() {
		var a WorkflowApproval
		if err := rows.Scan(&a.ID, &a.WorkflowID, &a.RunID, &a.StepID, &a.Message, &a.Status, &a.ApprovedBy, &a.CreatedAt, &a.ResolvedAt); err != nil {
			return nil, fmt.Errorf("scanning approval: %w", err)
		}
		approvals = append(approvals, a)
	}
	return approvals, rows.Err()
}
