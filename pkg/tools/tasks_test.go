package tools

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/storage"
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
	// Create task directly with storage using same userID (1) as the tool default
	id, err := store.CreateTaskForUser(1, "to archive", "desc", "todo", "")
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
