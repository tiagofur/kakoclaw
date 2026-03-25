package storage

import (
	"fmt"
	"time"
)

type TaskLog struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	Event     string    `json:"event"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Storage) AddTaskLog(taskID int64, event, message string) error {
	query := `INSERT INTO task_logs (task_id, event, message) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, taskID, event, message)
	if err != nil {
		return fmt.Errorf("adding task log: %w", err)
	}
	return nil
}

func (s *Storage) GetTaskLogs(taskID, userID int64) ([]TaskLog, error) {
	var (
		query string
		args  []interface{}
	)

	if s.isUserDB {
		query = `SELECT id, task_id, event, message, created_at FROM task_logs WHERE task_id = ? ORDER BY created_at ASC`
		args = []interface{}{taskID}
	} else {
		query = `SELECT tl.id, tl.task_id, tl.event, tl.message, tl.created_at
			FROM task_logs tl
			JOIN tasks t ON t.id = tl.task_id AND t.user_id = ?
			WHERE tl.task_id = ?
			ORDER BY tl.created_at ASC`
		args = []interface{}{userID, taskID}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("getting task logs: %w", err)
	}
	defer rows.Close()

	var logs []TaskLog
	for rows.Next() {
		var l TaskLog
		if err := rows.Scan(&l.ID, &l.TaskID, &l.Event, &l.Message, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning task log: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating task logs: %w", err)
	}
	return logs, nil
}
