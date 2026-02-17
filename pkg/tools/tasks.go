package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type TaskTool struct {
	db *sql.DB
}

func (t *TaskTool) Close() error {
	if t == nil || t.db == nil {
		return nil
	}
	return t.db.Close()
}

func NewTaskTool(workspace string) (*TaskTool, error) {
	webDir := filepath.Join(workspace, "web")
	if err := os.MkdirAll(webDir, 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(webDir, "tasks.db"))
	if err != nil {
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
	return &TaskTool{db: db}, nil
}

func (t *TaskTool) Name() string { return "task_manager" }

func (t *TaskTool) Description() string {
	return "Manage Kanban tasks: create, list, get, update_status."
}

func (t *TaskTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{"type": "string", "enum": []string{"create", "list", "get", "update_status"}},
			"id":     map[string]interface{}{"type": "string"},
			"title":  map[string]interface{}{"type": "string"},
			"description": map[string]interface{}{
				"type": "string",
			},
			"status": map[string]interface{}{"type": "string"},
			"limit":  map[string]interface{}{"type": "integer"},
		},
		"required": []string{"action"},
	}
}

func (t *TaskTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	switch action {
	case "create":
		title, _ := args["title"].(string)
		if strings.TrimSpace(title) == "" {
			return "Error: title is required", nil
		}
		description, _ := args["description"].(string)
		status, _ := args["status"].(string)
		if status == "" {
			status = "todo"
		}
		id := time.Now().UTC().Format("20060102150405.000000000")
		_, err := t.db.ExecContext(ctx, `INSERT INTO tasks(id, title, description, status, result, created_at) VALUES (?, ?, ?, ?, '', ?)`,
			id, strings.TrimSpace(title), strings.TrimSpace(description), status, time.Now().UTC())
		if err != nil {
			return "", err
		}
		_, _ = t.db.ExecContext(ctx, `INSERT INTO task_logs(id, task_id, action, details, created_at) VALUES (?, ?, ?, ?, ?)`,
			time.Now().UTC().Format("20060102150405.000000001"), id, "created_via_tool", "created from tool", time.Now().UTC())
		return fmt.Sprintf("Task created: %s (%s)", title, id), nil
	case "list":
		limit := 10
		if raw, ok := args["limit"].(float64); ok && int(raw) > 0 {
			limit = int(raw)
		}
		rows, err := t.db.QueryContext(ctx, `SELECT id, title, status FROM tasks ORDER BY created_at DESC LIMIT ?`, limit)
		if err != nil {
			return "", err
		}
		defer rows.Close()
		items := make([]map[string]string, 0)
		for rows.Next() {
			var id, title, status string
			if err := rows.Scan(&id, &title, &status); err != nil {
				return "", err
			}
			items = append(items, map[string]string{"id": id, "title": title, "status": status})
		}
		b, _ := json.Marshal(items)
		return string(b), nil
	case "get":
		id, _ := args["id"].(string)
		if strings.TrimSpace(id) == "" {
			return "Error: id is required", nil
		}
		row := t.db.QueryRowContext(ctx, `SELECT id, title, description, status, result, created_at FROM tasks WHERE id = ?`, id)
		var tid, title, desc, status, result string
		var created time.Time
		if err := row.Scan(&tid, &title, &desc, &status, &result, &created); err != nil {
			return "Task not found", nil
		}
		out, _ := json.Marshal(map[string]interface{}{
			"id": tid, "title": title, "description": desc, "status": status, "result": result, "created_at": created,
		})
		return string(out), nil
	case "update_status":
		id, _ := args["id"].(string)
		status, _ := args["status"].(string)
		if strings.TrimSpace(id) == "" || strings.TrimSpace(status) == "" {
			return "Error: id and status are required", nil
		}
		res, err := t.db.ExecContext(ctx, `UPDATE tasks SET status = ? WHERE id = ?`, status, id)
		if err != nil {
			return "", err
		}
		aff, _ := res.RowsAffected()
		if aff == 0 {
			return "Task not found", nil
		}
		_, _ = t.db.ExecContext(ctx, `INSERT INTO task_logs(id, task_id, action, details, created_at) VALUES (?, ?, ?, ?, ?)`,
			time.Now().UTC().Format("20060102150405.000000002"), id, "status_changed_via_tool", "status => "+status, time.Now().UTC())
		return "Task status updated", nil
	default:
		return "Error: unsupported action", nil
	}
}

