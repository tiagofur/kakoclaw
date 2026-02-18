package storage

import (
	"fmt"
	"time"
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Result      string    `json:"result"`
	Archived    bool      `json:"archived"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Storage) CreateTask(title, description, status string) (int64, error) {
	if status == "" {
		status = "todo"
	}
	query := `INSERT INTO tasks (title, description, status) VALUES (?, ?, ?)`
	result, err := s.db.Exec(query, title, description, status)
	if err != nil {
		return 0, fmt.Errorf("creating task: %w", err)
	}
	return result.LastInsertId()
}

func (s *Storage) GetTask(id int64) (*Task, error) {
	query := `SELECT id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, created_at, updated_at FROM tasks WHERE id = ?`
	var t Task
	err := s.db.QueryRow(query, id).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("getting task: %w", err)
	}
	return &t, nil
}

func (s *Storage) UpdateTask(id int64, title, description, status, result string) (*Task, error) {
	query := `UPDATE tasks SET title = ?, description = ?, status = ?, result = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, title, description, status, result, id)
	if err != nil {
		return nil, fmt.Errorf("updating task: %w", err)
	}
	return s.GetTask(id)
}

func (s *Storage) UpdateTaskStatus(id int64, status string) (*Task, error) {
	query := `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, status, id)
	if err != nil {
		return nil, fmt.Errorf("updating task status: %w", err)
	}
	return s.GetTask(id)
}

func (s *Storage) ArchiveTask(id int64) error {
	query := `UPDATE tasks SET archived = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("archiving task: %w", err)
	}
	return nil
}

func (s *Storage) UnarchiveTask(id int64) error {
	query := `UPDATE tasks SET archived = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("unarchiving task: %w", err)
	}
	return nil
}

func (s *Storage) DeleteTask(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM task_logs WHERE task_id = ?`, id); err != nil {
		return fmt.Errorf("deleting task logs: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM tasks WHERE id = ?`, id); err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}
	return tx.Commit()
}

func (s *Storage) ListTasks(includeArchived bool) ([]Task, error) {
	query := `SELECT id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, created_at, updated_at FROM tasks WHERE archived = ? OR ? ORDER BY created_at DESC`
	rows, err := s.db.Query(query, false, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.CreatedAt, &t.UpdatedAt); err != nil {
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
	sqlQuery := `SELECT id, title, COALESCE(description, ''), status, COALESCE(result, ''), archived, created_at, updated_at FROM tasks WHERE (title LIKE ? OR description LIKE ?) AND archived = 0 ORDER BY created_at DESC`
	searchTerm := "%" + query + "%"
	rows, err := s.db.Query(sqlQuery, searchTerm, searchTerm)
	if err != nil {
		return nil, fmt.Errorf("searching tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating search results: %w", err)
	}
	return tasks, nil
}
