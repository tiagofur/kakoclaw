package tools

import (
	"context"
	"strings"
	"testing"
)

func TestTaskToolCreateAndList(t *testing.T) {
	tool, err := NewTaskTool(t.TempDir())
	if err != nil {
		t.Fatalf("NewTaskTool failed: %v", err)
	}
	defer func() { _ = tool.Close() }()
	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"action": "create",
		"title":  "test task",
		"status": "todo",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	out, err := tool.Execute(context.Background(), map[string]interface{}{
		"action": "list",
		"limit":  float64(5),
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(out, "test task") {
		t.Fatalf("expected created task in list, got: %s", out)
	}
}

