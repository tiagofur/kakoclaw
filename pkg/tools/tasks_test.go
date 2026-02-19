package tools

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/storage"
)

func TestTaskToolCreateAndList(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := storage.New(config.StorageConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	defer func() { _ = store.Close() }()

	tool, err := NewTaskTool(store)
	if err != nil {
		t.Fatalf("NewTaskTool failed: %v", err)
	}
	defer func() { _ = tool.Close() }()

	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"action":      "create",
		"title":       "test task",
		"description": "desc",
		"status":      "todo",
	})
	if err != nil {
		t.Fatalf("execute create failed: %v", err)
	}

	// List (Active)
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

	// Archive
	// We need the ID, but created_via_tool doesn't return ID as int easily without parsing string "Task created: title (ID: 1)"
	// Simplification: We'll list again to get JSON and parse ID
	// Or we can just create another task via storage directly to hack it, ensuring we have ID.
	// But let's try to parse the output string from Create for now as it returns "Task created: ... (ID: <id>)"
	// Actually, the previous implementation returned "Task created: %s (ID: %d)".
	// Let's create a helper or just create via storage for test reliability.
	
	id, err := store.CreateTask("to archive", "desc", "todo")
	if err != nil {
		t.Fatalf("direct create failed: %v", err)
	}

	// Archive it
	_, err = tool.Execute(context.Background(), map[string]interface{}{
		"action": "archive",
		"id":     float64(id),
	})
	if err != nil {
		t.Fatalf("archive failed: %v", err)
	}

	// List (should not include archived)
	out, err = tool.Execute(context.Background(), map[string]interface{}{
		"action": "list",
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if strings.Contains(out, "to archive") {
		t.Fatalf("expected 'to archive' to be hidden, got: %s", out)
	}

	// List (include archived)
	out, err = tool.Execute(context.Background(), map[string]interface{}{
		"action":           "list",
		"include_archived": true,
	})
	if err != nil {
		t.Fatalf("list all failed: %v", err)
	}
	if !strings.Contains(out, "to archive") {
		t.Fatalf("expected 'to archive' to be shown, got: %s", out)
	}
}

