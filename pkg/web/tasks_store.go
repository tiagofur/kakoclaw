package web

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type taskStore struct {
	db *sql.DB
}

type taskLogItem struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

func newTaskStore(dataDir string) (*taskStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dataDir, "tasks.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL,
		result TEXT,
		created_at DATETIME NOT NULL
	);`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS task_logs (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL,
		action TEXT NOT NULL,
		details TEXT,
		created_at DATETIME NOT NULL
	);`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`ALTER TABLE tasks ADD COLUMN result TEXT`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
		_ = db.Close()
		return nil, err
	}
	return &taskStore{db: db}, nil
}

func (s *taskStore) close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *taskStore) list() ([]taskItem, error) {
	rows, err := s.db.Query(`SELECT id, title, description, status, result, created_at FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]taskItem, 0)
	for rows.Next() {
		var it taskItem
		if err := rows.Scan(&it.ID, &it.Title, &it.Description, &it.Status, &it.Result, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *taskStore) create(title, description, status string) (taskItem, error) {
	if status == "" {
		status = "backlog"
	}
	it := taskItem{
		ID:          generateID(),
		Title:       title,
		Description: description,
		Status:      status,
		Result:      "",
		CreatedAt:   time.Now().UTC(),
	}
	_, err := s.db.Exec(`INSERT INTO tasks(id, title, description, status, result, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		it.ID, it.Title, it.Description, it.Status, it.Result, it.CreatedAt)
	if err == nil {
		_ = s.addLog(it.ID, "created", "task created in "+it.Status)
	}
	return it, err
}

func (s *taskStore) get(id string) (taskItem, error) {
	var it taskItem
	err := s.db.QueryRow(`SELECT id, title, description, status, result, created_at FROM tasks WHERE id = ?`, id).
		Scan(&it.ID, &it.Title, &it.Description, &it.Status, &it.Result, &it.CreatedAt)
	if err != nil {
		return taskItem{}, err
	}
	return it, nil
}

func (s *taskStore) update(id, title, description, status, result string) (taskItem, error) {
	res, err := s.db.Exec(`UPDATE tasks SET title = ?, description = ?, status = ?, result = ? WHERE id = ?`, title, description, status, result, id)
	if err != nil {
		return taskItem{}, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return taskItem{}, err
	}
	if affected == 0 {
		return taskItem{}, sql.ErrNoRows
	}
	item, err := s.get(id)
	if err == nil {
		_ = s.addLog(id, "updated", "task updated")
	}
	return item, err
}

func (s *taskStore) updateStatus(id, status string) (taskItem, error) {
	if status == "" {
		return taskItem{}, errors.New("status is required")
	}
	res, err := s.db.Exec(`UPDATE tasks SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return taskItem{}, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return taskItem{}, err
	}
	if affected == 0 {
		return taskItem{}, sql.ErrNoRows
	}
	item, err := s.get(id)
	if err == nil {
		_ = s.addLog(id, "status_changed", "status => "+status)
	}
	return item, err
}

func (s *taskStore) updateResult(id, result string) (taskItem, error) {
	res, err := s.db.Exec(`UPDATE tasks SET result = ? WHERE id = ?`, result, id)
	if err != nil {
		return taskItem{}, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return taskItem{}, err
	}
	if affected == 0 {
		return taskItem{}, sql.ErrNoRows
	}
	item, err := s.get(id)
	if err == nil {
		_ = s.addLog(id, "result_updated", "result updated")
	}
	return item, err
}

func (s *taskStore) delete(id string) error {
	res, err := s.db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	_ = s.addLog(id, "deleted", "task deleted")
	return nil
}

func (s *taskStore) addLog(taskID, action, details string) error {
	_, err := s.db.Exec(`INSERT INTO task_logs(id, task_id, action, details, created_at) VALUES (?, ?, ?, ?, ?)`,
		generateID(), taskID, action, details, time.Now().UTC())
	return err
}

func (s *taskStore) listLogs(taskID string) ([]taskLogItem, error) {
	rows, err := s.db.Query(`SELECT id, task_id, action, details, created_at FROM task_logs WHERE task_id = ? ORDER BY created_at DESC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]taskLogItem, 0)
	for rows.Next() {
		var it taskLogItem
		if err := rows.Scan(&it.ID, &it.TaskID, &it.Action, &it.Details, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func generateID() string {
	return time.Now().UTC().Format("20060102150405.000000000")
}
