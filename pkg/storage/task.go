package storage

import (
	"fmt"
	"time"
)

type Task struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"` // User who owns this task
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Result      string    `json:"result"`
	Archived    bool      `json:"archived"`
	Agent       string    `json:"agent"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Storage) CreateTask(title, description, status, agent string) (int64, error) {
	return s.CreateTaskForUser(nil, title, description, status, agent)
}

func (s *Storage) CreateTaskForUser(userKey interface{}, title, description, status, agent string) (int64, error) {
	if status == "" {
		status = "todo"
	}
	var query string
	var args []interface{}

	if s.isUserDB {
		query = `INSERT INTO tasks (title, description, status, agent) VALUES (?, ?, ?, ?)`
		args = []interface{}{title, description, status, agent}
	} else {
		uid := normalizeUserKey(userKey)
		query = `INSERT INTO tasks (user_id, title, description, status, agent) VALUES (?, ?, ?, ?, ?)`
		args = []interface{}{uid, title, description, status, agent}
	}

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("creating task: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting insert id: %w", err)
	}
	return id, nil
}

func (s *Storage) GetTask(id int64) (*Task, error) {
	return s.GetTaskForUser(nil, id)
}

func (s *Storage) GetTaskForUser(userKey interface{}, id int64) (*Task, error) {
	var query string
	var args []interface{}

	if s.isUserDB {
		query = `SELECT id, 0 as user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks WHERE id = ?`
		args = []interface{}{id}
	} else {
		// If userKey is provided (not nil) and not 0, filter by it. Else, fetch without user_id filter.
		if userKey != nil && normalizeUserKey(userKey) != 0 {
			uid := normalizeUserKey(userKey)
			query = `SELECT id, user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks WHERE id = ? AND user_id = ?`
			args = []interface{}{id, uid}
		} else {
			query = `SELECT id, COALESCE(user_id, 0), title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks WHERE id = ?`
			args = []interface{}{id}
		}
	}

	var t Task
	err := s.db.QueryRow(query, args...).Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.Agent, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting task: %w", err)
	}
	return &t, nil
}

func (s *Storage) UpdateTask(id int64, title, description, status, result string) (*Task, error) {
	return s.UpdateTaskForUser(nil, id, title, description, status, result)
}

func (s *Storage) UpdateTaskForUser(userKey interface{}, id int64, title, description, status, result string) (*Task, error) {
	var query string
	var args []interface{}

	if s.isUserDB {
		query = `UPDATE tasks SET title = ?, description = ?, status = ?, result = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
		args = []interface{}{title, description, status, result, id}
	} else {
		if userKey != nil && normalizeUserKey(userKey) != 0 {
			uid := normalizeUserKey(userKey)
			query = `UPDATE tasks SET title = ?, description = ?, status = ?, result = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
			args = []interface{}{title, description, status, result, id, uid}
		} else {
			query = `UPDATE tasks SET title = ?, description = ?, status = ?, result = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
			args = []interface{}{title, description, status, result, id}
		}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("updating task: %w", err)
	}
	return s.GetTaskForUser(userKey, id)
}

func (s *Storage) RenameTaskAgentForUser(userKey interface{}, oldAgent, newAgent string) error {
	var query string
	var args []interface{}

	if s.isUserDB {
		query = `UPDATE tasks SET agent = ?, updated_at = CURRENT_TIMESTAMP WHERE agent = ?`
		args = []interface{}{newAgent, oldAgent}
	} else {
		uid := normalizeUserKey(userKey)
		query = `UPDATE tasks SET agent = ?, updated_at = CURRENT_TIMESTAMP WHERE agent = ? AND user_id = ?`
		args = []interface{}{newAgent, oldAgent, uid}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("renaming task agent: %w", err)
	}
	return nil
}

func (s *Storage) UpdateTaskStatus(id int64, status string) (*Task, error) {
	return s.UpdateTaskStatusForUser(nil, id, status)
}

func (s *Storage) UpdateTaskStatusForUser(userKey interface{}, id int64, status string) (*Task, error) {
	var query string
	var args []interface{}

	if s.isUserDB {
		query = `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
		args = []interface{}{status, id}
	} else {
		if userKey != nil && normalizeUserKey(userKey) != 0 {
			uid := normalizeUserKey(userKey)
			query = `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
			args = []interface{}{status, id, uid}
		} else {
			query = `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
			args = []interface{}{status, id}
		}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("updating task status: %w", err)
	}
	return s.GetTaskForUser(userKey, id)
}

func (s *Storage) ArchiveTask(id int64) error {
	return s.ArchiveTaskForUser(nil, id)
}

func (s *Storage) ArchiveTaskForUser(userKey interface{}, id int64) error {
	var query string
	var args []interface{}

	if s.isUserDB {
		query = `UPDATE tasks SET archived = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
		args = []interface{}{id}
	} else {
		if userKey != nil && normalizeUserKey(userKey) != 0 {
			uid := normalizeUserKey(userKey)
			query = `UPDATE tasks SET archived = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
			args = []interface{}{id, uid}
		} else {
			query = `UPDATE tasks SET archived = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
			args = []interface{}{id}
		}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("archiving task: %w", err)
	}
	return nil
}

func (s *Storage) UnarchiveTask(id int64) error {
	return s.UnarchiveTaskForUser(nil, id)
}

func (s *Storage) UnarchiveTaskForUser(userKey interface{}, id int64) error {
	var query string
	var args []interface{}

	if s.isUserDB {
		query = `UPDATE tasks SET archived = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
		args = []interface{}{id}
	} else {
		if userKey != nil && normalizeUserKey(userKey) != 0 {
			uid := normalizeUserKey(userKey)
			query = `UPDATE tasks SET archived = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
			args = []interface{}{id, uid}
		} else {
			query = `UPDATE tasks SET archived = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
			args = []interface{}{id}
		}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("unarchiving task: %w", err)
	}
	return nil
}

func (s *Storage) DeleteTask(id int64) error {
	return s.DeleteTaskForUser(nil, id)
}

func (s *Storage) DeleteTaskForUser(userKey interface{}, id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM task_logs WHERE task_id = ?`, id); err != nil {
		return fmt.Errorf("deleting task logs: %w", err)
	}

	var query string
	var args []interface{}
	if s.isUserDB {
		query = `DELETE FROM tasks WHERE id = ?`
		args = []interface{}{id}
	} else {
		if userKey != nil && normalizeUserKey(userKey) != 0 {
			uid := normalizeUserKey(userKey)
			query = `DELETE FROM tasks WHERE id = ? AND user_id = ?`
			args = []interface{}{id, uid}
		} else {
			query = `DELETE FROM tasks WHERE id = ?`
			args = []interface{}{id}
		}
	}

	if _, err := tx.Exec(query, args...); err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}
	return tx.Commit()
}

func (s *Storage) ListTasks(includeArchived bool) ([]Task, error) {
	return s.ListTasksForUser(nil, includeArchived)
}

func (s *Storage) ListTasksForUser(userKey interface{}, includeArchived bool) ([]Task, error) {
	var query string
	var args []interface{}

	if userKey == nil {
		if includeArchived {
			query = `SELECT id, 0 as user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks WHERE archived = 0 OR 1=1 ORDER BY created_at DESC`
		} else {
			query = `SELECT id, 0 as user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks WHERE archived = 0 ORDER BY created_at DESC`
		}
	} else if userID, ok := userKey.(int64); ok && !s.isUserDB {
		if includeArchived {
			query = `SELECT id, user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks WHERE user_id = ? ORDER BY created_at DESC`
		} else {
			query = `SELECT id, user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks WHERE user_id = ? AND archived = 0 ORDER BY created_at DESC`
		}
		args = []interface{}{userID}
	} else {
		// Fallback to original logic if userKey is not nil or int64, e.g., for s.isUserDB
		if s.isUserDB {
			query = `SELECT id, 0 as user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks WHERE (archived = ? OR ?) ORDER BY created_at DESC`
			args = []interface{}{false, includeArchived}
		} else {
			uid := normalizeUserKey(userKey)
			query = `SELECT id, user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks WHERE user_id = ? AND (archived = ? OR ?) ORDER BY created_at DESC`
			args = []interface{}{uid, false, includeArchived}
		}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.Agent, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (s *Storage) SearchTasks(query string) ([]Task, error) {
	return s.SearchTasksForUser(nil, query)
}

func (s *Storage) SearchTasksForUser(userKey interface{}, query string) ([]Task, error) {
	var sqlQuery string
	var args []interface{}
	searchTerm := "%" + escapeLikeQuery(query) + "%"

	if s.isUserDB {
		sqlQuery = `SELECT id, 0 as user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks WHERE (title LIKE ? ESCAPE '\' OR description LIKE ? ESCAPE '\') AND archived = 0 ORDER BY created_at DESC`
		args = []interface{}{searchTerm, searchTerm}
	} else {
		uid := normalizeUserKey(userKey)
		sqlQuery = `SELECT id, user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks WHERE user_id = ? AND (title LIKE ? ESCAPE '\' OR description LIKE ? ESCAPE '\') AND archived = 0 ORDER BY created_at DESC`
		args = []interface{}{uid, searchTerm, searchTerm}
	}

	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("searching tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.Agent, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// ListAllUsersTasks returns tasks for all users (for background worker).
func (s *Storage) ListAllUsersTasks(includeArchived bool) ([]Task, error) {
	var query string
	if includeArchived {
		query = `SELECT id, user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks ORDER BY created_at DESC`
	} else {
		query = `SELECT id, user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, agent, created_at, updated_at FROM tasks WHERE archived = 0 ORDER BY created_at DESC`
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.Agent, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating all tasks: %w", err)
	}
	return tasks, nil
}
