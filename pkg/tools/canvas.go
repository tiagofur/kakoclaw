package tools

import (
	"context"
	"fmt"

	"github.com/sipeed/makoclaw/pkg/canvas"
)

// CanvasTool provides the agent with canvas operations: push, eval, snapshot, reset.
type CanvasTool struct {
	server *canvas.CanvasServer
}

// NewCanvasTool creates a new CanvasTool. Pass nil server to indicate canvas is disabled.
func NewCanvasTool(server *canvas.CanvasServer) *CanvasTool {
	return &CanvasTool{server: server}
}

func (t *CanvasTool) Name() string {
	return "canvas"
}

func (t *CanvasTool) Description() string {
	return "Control the visual canvas workspace. Operations: push (replace HTML content), eval (execute JS in sandbox, requires dev_mode), snapshot (get current content), reset (clear canvas)."
}

func (t *CanvasTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "The canvas operation: push, eval, snapshot, or reset",
				"enum":        []string{"push", "eval", "snapshot", "reset"},
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "HTML content (for push) or JavaScript expression (for eval)",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "Content format: html (default) or text",
			},
		},
		"required": []string{"operation"},
	}
}

func (t *CanvasTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	if t.server == nil {
		return "", fmt.Errorf("Canvas is disabled. Set canvas.enabled = true to use this feature.")
	}

	operation, _ := args["operation"].(string)
	content, _ := args["content"].(string)

	switch operation {
	case "push":
		if content == "" {
			return "", fmt.Errorf("content is required for push operation")
		}
		t.server.Push(content)
		return "Content pushed to canvas.", nil

	case "eval":
		if content == "" {
			return "", fmt.Errorf("content is required for eval operation")
		}
		result, err := t.server.Eval(content)
		if err != nil {
			return "", err
		}
		return result, nil

	case "snapshot":
		snap := t.server.Snapshot()
		if snap == "" {
			return "Canvas is empty.", nil
		}
		return snap, nil

	case "reset":
		t.server.Reset()
		return "Canvas reset.", nil

	default:
		return "", fmt.Errorf("unknown canvas operation: %s", operation)
	}
}
