package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type SpawnTool struct {
	manager       *SubagentManager
	originChannel string
	originChatID  string
	mu            sync.Mutex
}

func NewSpawnTool(manager *SubagentManager) *SpawnTool {
	return &SpawnTool{
		manager:       manager,
		originChannel: "cli",
		originChatID:  "direct",
	}
}

func (t *SpawnTool) Name() string {
	return "spawn"
}

func (t *SpawnTool) Description() string {
	return "Spawn a subagent to handle a task in the background. Use this for complex or time-consuming tasks that can run independently. The subagent will complete the task and report back when done."
}

func (t *SpawnTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task": map[string]interface{}{
				"type":        "string",
				"description": "The task for subagent to complete",
			},
			"label": map[string]interface{}{
				"type":        "string",
				"description": "Optional short label for the task (for display)",
			},
		},
		"required": []string{"task"},
	}
}

func (t *SpawnTool) SetContext(channel, chatID string) {
	t.mu.Lock()
	t.originChannel = channel
	t.originChatID = chatID
	t.mu.Unlock()
}

func (t *SpawnTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	task, ok := args["task"].(string)
	if !ok {
		return "", fmt.Errorf("task is required")
	}

	label, _ := args["label"].(string)

	if t.manager == nil {
		return "Error: Subagent manager not configured", nil
	}

	t.mu.Lock()
	channel, chatID := t.originChannel, t.originChatID
	t.mu.Unlock()

	result, err := t.manager.Spawn(ctx, task, label, channel, chatID)
	if err != nil {
		return "", fmt.Errorf("failed to spawn subagent: %w", err)
	}

	return result, nil
}

// SpawnStatusTool allows checking on spawned subagent tasks.
type SpawnStatusTool struct {
	manager *SubagentManager
}

func NewSpawnStatusTool(manager *SubagentManager) *SpawnStatusTool {
	return &SpawnStatusTool{manager: manager}
}

func (t *SpawnStatusTool) Name() string {
	return "spawn_status"
}

func (t *SpawnStatusTool) Description() string {
	return "Check the status of spawned subagent tasks. Use without task_id to list all tasks, or with task_id to get details of a specific task. Can also cancel running tasks."
}

func (t *SpawnStatusTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional: specific task ID to check",
			},
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action to perform: 'status' (default) or 'cancel'",
				"enum":        []string{"status", "cancel"},
			},
		},
	}
}

func (t *SpawnStatusTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	if t.manager == nil {
		return "Error: Subagent manager not configured", nil
	}

	taskID, _ := args["task_id"].(string)
	action, _ := args["action"].(string)
	if action == "" {
		action = "status"
	}

	if action == "cancel" && taskID != "" {
		if err := t.manager.CancelTask(taskID); err != nil {
			return fmt.Sprintf("Failed to cancel: %s", err.Error()), nil
		}
		return fmt.Sprintf("Task %s cancelled.", taskID), nil
	}

	if taskID != "" {
		task, ok := t.manager.GetTask(taskID)
		if !ok {
			return fmt.Sprintf("Task %s not found.", taskID), nil
		}
		data, _ := json.Marshal(task)
		return string(data), nil
	}

	// List all tasks
	tasks := t.manager.ListTasks()
	if len(tasks) == 0 {
		return "No subagent tasks.", nil
	}

	data, _ := json.Marshal(tasks)
	return string(data), nil
}
