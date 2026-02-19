package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/sipeed/kakoclaw/pkg/storage"
)

type TaskTool struct {
	storage *storage.Storage
}

func (t *TaskTool) Close() error {
	return nil // Storage is closed by the agent/server
}

func NewTaskTool(s *storage.Storage) (*TaskTool, error) {
	if s == nil {
		return nil, fmt.Errorf("storage is nil")
	}
	return &TaskTool{storage: s}, nil
}

func (t *TaskTool) Name() string { return "task_manager" }

func (t *TaskTool) Description() string {
	return "Manage tasks: create, list, update_status, archive, unarchive. (ID is integer)"
}

func (t *TaskTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{"type": "string", "enum": []string{"create", "list", "update_status", "search", "archive", "unarchive"}},
			"id":     map[string]interface{}{"type": "integer"},
			"title":  map[string]interface{}{"type": "string"},
			"description": map[string]interface{}{
				"type": "string",
			},
			"status":           map[string]interface{}{"type": "string"},
			"query":            map[string]interface{}{"type": "string"},
			"include_archived": map[string]interface{}{"type": "boolean"},
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
		id, err := t.storage.CreateTask(title, description, status)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Task created: %s (ID: %d)", title, id), nil
	case "list":
		includeArchived, _ := args["include_archived"].(bool)
		tasks, err := t.storage.ListTasks(includeArchived)
		if err != nil {
			return "", err
		}
		b, _ := json.Marshal(tasks)
		return string(b), nil
	case "search":
		query, _ := args["query"].(string)
		tasks, err := t.storage.SearchTasks(query)
		if err != nil {
			return "", err
		}
		b, _ := json.Marshal(tasks)
		return string(b), nil
	case "update_status":
		var id int64
		if idFloat, ok := args["id"].(float64); ok {
			id = int64(idFloat)
		} else if idStr, ok := args["id"].(string); ok {
			var err error
			id, err = strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return "Error: invalid id format", nil
			}
		} else {
			return "Error: id is required", nil
		}

		status, _ := args["status"].(string)
		if strings.TrimSpace(status) == "" {
			return "Error: status is required", nil
		}
		_, err := t.storage.UpdateTaskStatus(id, status)
		if err != nil {
			return "", err
		}
		return "Task status updated", nil
	case "archive":
		var id int64
		if idFloat, ok := args["id"].(float64); ok {
			id = int64(idFloat)
		} else if idStr, ok := args["id"].(string); ok {
			var err error
			id, err = strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return "Error: invalid id format", nil
			}
		} else {
			return "Error: id is required", nil
		}
		err := t.storage.ArchiveTask(id)
		if err != nil {
			return "", err
		}
		return "Task archived", nil
	case "unarchive":
		var id int64
		if idFloat, ok := args["id"].(float64); ok {
			id = int64(idFloat)
		} else if idStr, ok := args["id"].(string); ok {
			var err error
			id, err = strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return "Error: invalid id format", nil
			}
		} else {
			return "Error: id is required", nil
		}
		err := t.storage.UnarchiveTask(id)
		if err != nil {
			return "", err
		}
		return "Task unarchived", nil
	default:
		return "Error: unsupported action", nil
	}
}


