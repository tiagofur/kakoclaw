package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/storage"
)

func newSOTestStorage(t *testing.T) *storage.Storage {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "so_test.db")
	s, err := storage.NewUserStorage(dbPath)
	if err != nil {
		t.Fatalf("NewUserStorage: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func newSOTool(t *testing.T) (*StandingOrderTool, *storage.Storage) {
	t.Helper()
	s := newSOTestStorage(t)
	tool, err := NewStandingOrderTool(s)
	if err != nil {
		t.Fatalf("NewStandingOrderTool: %v", err)
	}
	return tool, s
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestNewStandingOrderTool_NilStorage(t *testing.T) {
	_, err := NewStandingOrderTool(nil)
	if err == nil {
		t.Fatal("expected error for nil storage")
	}
}

func TestNewStandingOrderTool_Valid(t *testing.T) {
	tool, _ := newSOTool(t)
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
}

// ---------------------------------------------------------------------------
// SetUserID
// ---------------------------------------------------------------------------

func TestSetUserID(t *testing.T) {
	tool, _ := newSOTool(t)
	tool.SetUserID(42)
	if tool.userID != 42 {
		t.Errorf("expected userID 42, got %d", tool.userID)
	}
}

func TestSetUserID_ZeroIgnored(t *testing.T) {
	tool, _ := newSOTool(t)
	tool.SetUserID(99)
	tool.SetUserID(0) // should be ignored
	if tool.userID != 99 {
		t.Errorf("expected userID 99 (zero ignored), got %d", tool.userID)
	}
}

// ---------------------------------------------------------------------------
// Metadata
// ---------------------------------------------------------------------------

func TestStandingOrderTool_Name(t *testing.T) {
	tool, _ := newSOTool(t)
	if tool.Name() != "standing_orders" {
		t.Errorf("unexpected tool name: %q", tool.Name())
	}
}

func TestStandingOrderTool_Parameters(t *testing.T) {
	tool, _ := newSOTool(t)
	params := tool.Parameters()
	if params["type"] != "object" {
		t.Error("expected type 'object'")
	}
	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected properties map")
	}
	if _, ok := props["action"]; !ok {
		t.Error("expected 'action' in parameters")
	}
}

// ---------------------------------------------------------------------------
// list
// ---------------------------------------------------------------------------

func TestExecute_List_Empty(t *testing.T) {
	tool, _ := newSOTool(t)
	result, err := tool.Execute(context.Background(), map[string]interface{}{"action": "list"})
	if err != nil {
		t.Fatalf("Execute list: %v", err)
	}
	if result != "[]" && result != "null" {
		// Unmarshal to verify it's a valid JSON array
		var arr []interface{}
		if jsonErr := json.Unmarshal([]byte(result), &arr); jsonErr != nil {
			t.Fatalf("expected valid JSON array, got: %q", result)
		}
		if len(arr) != 0 {
			t.Errorf("expected empty array, got %d items", len(arr))
		}
	}
}

func TestExecute_List_WithOrders(t *testing.T) {
	tool, _ := newSOTool(t)

	_, _ = tool.Execute(context.Background(), map[string]interface{}{
		"action":  "create",
		"content": "Always be concise.",
		"label":   "style",
	})

	result, err := tool.Execute(context.Background(), map[string]interface{}{"action": "list"})
	if err != nil {
		t.Fatalf("Execute list: %v", err)
	}

	var arr []map[string]interface{}
	if err := json.Unmarshal([]byte(result), &arr); err != nil {
		t.Fatalf("list result not valid JSON: %v (got: %s)", err, result)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 order, got %d", len(arr))
	}
	if arr[0]["content"] != "Always be concise." {
		t.Errorf("unexpected content: %v", arr[0]["content"])
	}
	if _, ok := arr[0]["id"]; !ok {
		t.Error("expected 'id' field in list output")
	}
	if _, ok := arr[0]["created_at"]; !ok {
		t.Error("expected 'created_at' field in list output")
	}
}

// ---------------------------------------------------------------------------
// create
// ---------------------------------------------------------------------------

func TestExecute_Create_MissingContent(t *testing.T) {
	tool, _ := newSOTool(t)
	result, err := tool.Execute(context.Background(), map[string]interface{}{"action": "create"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "error") || !strings.Contains(result, "content") {
		t.Errorf("expected error about missing content, got: %q", result)
	}
}

func TestExecute_Create_Success(t *testing.T) {
	tool, _ := newSOTool(t)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"action":  "create",
		"content": "Respond only in bullet points.",
		"label":   "format",
	})
	if err != nil {
		t.Fatalf("Execute create: %v", err)
	}
	if !strings.Contains(result, "Standing order added") {
		t.Errorf("unexpected create result: %q", result)
	}
}

func TestExecute_Create_LimitExceeded(t *testing.T) {
	tool, _ := newSOTool(t)
	longContent := strings.Repeat("x", 201)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"action":  "create",
		"content": longContent,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(result, "error:") {
		t.Errorf("expected error result for long content, got: %q", result)
	}
}

// ---------------------------------------------------------------------------
// update
// ---------------------------------------------------------------------------

func TestExecute_Update_MissingID(t *testing.T) {
	tool, _ := newSOTool(t)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"action":  "update",
		"content": "new content",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "error") || !strings.Contains(result, "id") {
		t.Errorf("expected error about missing id, got: %q", result)
	}
}

func TestExecute_Update_Success(t *testing.T) {
	tool, _ := newSOTool(t)

	createResult, _ := tool.Execute(context.Background(), map[string]interface{}{
		"action":  "create",
		"content": "original content",
		"label":   "orig",
	})

	// Extract id from result message
	var id float64 = 1
	_ = createResult // id will be 1 from first insert

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"action":  "update",
		"id":      id,
		"content": "updated content",
		"label":   "upd",
	})
	if err != nil {
		t.Fatalf("Execute update: %v", err)
	}
	if result != "Updated." {
		t.Errorf("expected 'Updated.', got: %q", result)
	}
}

// ---------------------------------------------------------------------------
// delete
// ---------------------------------------------------------------------------

func TestExecute_Delete_MissingID(t *testing.T) {
	tool, _ := newSOTool(t)
	result, err := tool.Execute(context.Background(), map[string]interface{}{"action": "delete"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "error") || !strings.Contains(result, "id") {
		t.Errorf("expected error about missing id, got: %q", result)
	}
}

func TestExecute_Delete_Success(t *testing.T) {
	tool, _ := newSOTool(t)

	_, _ = tool.Execute(context.Background(), map[string]interface{}{
		"action":  "create",
		"content": "to be deleted",
	})

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"action": "delete",
		"id":     float64(1),
	})
	if err != nil {
		t.Fatalf("Execute delete: %v", err)
	}
	if result != "Removed." {
		t.Errorf("expected 'Removed.', got: %q", result)
	}

	listResult, _ := tool.Execute(context.Background(), map[string]interface{}{"action": "list"})
	var arr []interface{}
	_ = json.Unmarshal([]byte(listResult), &arr)
	if len(arr) != 0 {
		t.Errorf("expected empty list after delete, got %d items", len(arr))
	}
}

// ---------------------------------------------------------------------------
// toggle
// ---------------------------------------------------------------------------

func TestExecute_Toggle_MissingID(t *testing.T) {
	tool, _ := newSOTool(t)
	result, err := tool.Execute(context.Background(), map[string]interface{}{"action": "toggle"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "error") || !strings.Contains(result, "id") {
		t.Errorf("expected error about missing id, got: %q", result)
	}
}

func TestExecute_Toggle_Success(t *testing.T) {
	tool, _ := newSOTool(t)

	_, _ = tool.Execute(context.Background(), map[string]interface{}{
		"action":  "create",
		"content": "toggle this",
	})

	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"action": "toggle",
		"id":     float64(1),
	})
	if err != nil {
		t.Fatalf("Execute toggle: %v", err)
	}
	if result != "Toggled." {
		t.Errorf("expected 'Toggled.', got: %q", result)
	}
}

// ---------------------------------------------------------------------------
// unknown action
// ---------------------------------------------------------------------------

func TestExecute_UnknownAction(t *testing.T) {
	tool, _ := newSOTool(t)
	result, err := tool.Execute(context.Background(), map[string]interface{}{"action": "explode"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "error") || !strings.Contains(result, "explode") {
		t.Errorf("expected error mentioning unknown action 'explode', got: %q", result)
	}
}
