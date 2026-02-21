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
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Storage) CreateTask(title, description, status string) (int64, error) {
	return s.CreateTaskForUser(0, title, description, status)
}

func (s *Storage) CreateTaskForUser(userID int64, title, description, status string) (int64, error) {
	if status == "" {
		status = "todo"
	}
	uid := normalizeUserID(userID)
	query := `INSERT INTO tasks (user_id, title, description, status) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, uid, title, description, status)
	if err != nil {
		return 0, fmt.Errorf("creating task: %w", err)
	}
	return result.LastInsertId()
}

func (s *Storage) GetTask(id int64) (*Task, error) {
	return s.GetTaskForUser(0, id)
}

func (s *Storage) GetTaskForUser(userID, id int64) (*Task, error) {
	uid := normalizeUserID(userID)
	query := `SELECT id, user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, created_at, updated_at FROM tasks WHERE id = ? AND user_id = ?`
	var t Task
	err := s.db.QueryRow(query, id, uid).Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting task: %w", err)
	}
	return &t, nil
}

func (s *Storage) UpdateTask(id int64, title, description, status, result string) (*Task, error) {
	return s.UpdateTaskForUser(0, id, title, description, status, result)
}

func (s *Storage) UpdateTaskForUser(userID, id int64, title, description, status, result string) (*Task, error) {
	uid := normalizeUserID(userID)
	query := `UPDATE tasks SET title = ?, description = ?, status = ?, result = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
	_, err := s.db.Exec(query, title, description, status, result, id, uid)
	if err != nil {
		return nil, fmt.Errorf("updating task: %w", err)
	}
	return s.GetTaskForUser(uid, id)
}

func (s *Storage) UpdateTaskStatus(id int64, status string) (*Task, error) {
	return s.UpdateTaskStatusForUser(0, id, status)
}

func (s *Storage) UpdateTaskStatusForUser(userID, id int64, status string) (*Task, error) {
	uid := normalizeUserID(userID)
	query := `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
	_, err := s.db.Exec(query, status, id, uid)
	if err != nil {
		return nil, fmt.Errorf("updating task status: %w", err)
	}
	return s.GetTaskForUser(uid, id)
}

func (s *Storage) ArchiveTask(id int64) error {
	return s.ArchiveTaskForUser(0, id)
}

func (s *Storage) ArchiveTaskForUser(userID, id int64) error {
	uid := normalizeUserID(userID)
	query := `UPDATE tasks SET archived = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
	_, err := s.db.Exec(query, id, uid)
	if err != nil {
		return fmt.Errorf("archiving task: %w", err)
	}
	return nil
}

func (s *Storage) UnarchiveTask(id int64) error {
	return s.UnarchiveTaskForUser(0, id)
}

func (s *Storage) UnarchiveTaskForUser(userID, id int64) error {
	uid := normalizeUserID(userID)
	query := `UPDATE tasks SET archived = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
	_, err := s.db.Exec(query, id, uid)
	if err != nil {
		return fmt.Errorf("unarchiving task: %w", err)
	}
	return nil
}

func (s *Storage) DeleteTask(id int64) error {
	return s.DeleteTaskForUser(0, id)
}

func (s *Storage) DeleteTaskForUser(userID, id int64) error {
	uid := normalizeUserID(userID)
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM task_logs WHERE task_id = ?`, id); err != nil {
		return fmt.Errorf("deleting task logs: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM tasks WHERE id = ? AND user_id = ?`, id, uid); err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}
	return tx.Commit()
}

func (s *Storage) ListTasks(includeArchived bool) ([]Task, error) {
	return s.ListTasksForUser(0, includeArchived)
}

func (s *Storage) ListTasksForUser(userID int64, includeArchived bool) ([]Task, error) {
	uid := normalizeUserID(userID)
	query := `SELECT id, user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, created_at, updated_at FROM tasks WHERE user_id = ? AND (archived = ? OR ?) ORDER BY created_at DESC`
	rows, err := s.db.Query(query, uid, false, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating tasks: %w", err)
	}
	return tasks, nil
}

func (s *Storage) SearchTasks(query string) ([]Task, error) {
	return s.SearchTasksForUser(0, query)
}

func (s *Storage) SearchTasksForUser(userID int64, query string) ([]Task, error) {
	uid := normalizeUserID(userID)
	sqlQuery := `SELECT id, user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, created_at, updated_at FROM tasks WHERE user_id = ? AND (title LIKE ? OR description LIKE ?) AND archived = 0 ORDER BY created_at DESC`
	searchTerm := "%" + query + "%"
	rows, err := s.db.Query(sqlQuery, uid, searchTerm, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("searching tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating search results: %w", err)
	}
	return tasks, nil
}

// ListAllUsersTasks returns tasks for all users (for background worker).
func (s *Storage) ListAllUsersTasks(includeArchived bool) ([]Task, error) {
	query := `SELECT id, user_id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, created_at, updated_at FROM tasks WHERE (archived = ? OR ?) ORDER BY created_at DESC`
	rows, err := s.db.Query(query, false, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("listing all tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating all tasks: %w", err)
	}
	return tasks, nil
}
