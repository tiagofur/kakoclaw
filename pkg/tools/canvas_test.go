package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/canvas"
)

func TestCanvasToolDisabledReturnsError(t *testing.T) {
	tool := NewCanvasTool(nil)
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "push",
		"content":   "<h1>hi</h1>",
		"format":    "html",
	})
	if err == nil {
		t.Error("expected error when canvas is disabled")
	}
	if !strings.Contains(err.Error(), "Canvas is disabled") {
		t.Errorf("error should mention 'Canvas is disabled', got: %s", err.Error())
	}
}

func TestCanvasToolPushDelegates(t *testing.T) {
	srv := canvas.NewCanvasServer(false)
	tool := NewCanvasTool(srv)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "push",
		"content":   "<p>test</p>",
		"format":    "html",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(strings.ToLower(result), "pushed") {
		t.Errorf("result should mention 'pushed', got: %s", result)
	}
	if snap := srv.Snapshot(); snap != "<p>test</p>" {
		t.Errorf("server snapshot = %q, want %q", snap, "<p>test</p>")
	}
}

func TestCanvasToolSnapshot(t *testing.T) {
	srv := canvas.NewCanvasServer(false)
	srv.Push("<div>content</div>")
	tool := NewCanvasTool(srv)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "snapshot",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "<div>content</div>" {
		t.Errorf("snapshot = %q, want %q", result, "<div>content</div>")
	}
}

func TestCanvasToolReset(t *testing.T) {
	srv := canvas.NewCanvasServer(false)
	srv.Push("<p>data</p>")
	tool := NewCanvasTool(srv)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "reset",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(strings.ToLower(result), "reset") {
		t.Errorf("result should mention 'reset', got: %s", result)
	}
	if snap := srv.Snapshot(); snap != "" {
		t.Errorf("snapshot after reset = %q, want empty", snap)
	}
}

func TestCanvasToolName(t *testing.T) {
	tool := NewCanvasTool(nil)
	if tool.Name() != "canvas" {
		t.Errorf("Name() = %q, want %q", tool.Name(), "canvas")
	}
}

func TestCanvasToolEvalDevModeOff(t *testing.T) {
	srv := canvas.NewCanvasServer(false)
	tool := NewCanvasTool(srv)
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "eval",
		"content":   "alert(1)",
	})
	if err == nil {
		t.Error("expected error when dev_mode is off")
	}
	if !strings.Contains(err.Error(), "dev_mode") {
		t.Errorf("error should mention dev_mode, got: %s", err.Error())
	}
}

func TestCanvasToolEvalDevModeOn(t *testing.T) {
	srv := canvas.NewCanvasServer(true)
	tool := NewCanvasTool(srv)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "eval",
		"content":   "document.title",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "eval sent" {
		t.Errorf("result = %q, want %q", result, "eval sent")
	}
}

func TestCanvasToolSnapshotEmpty(t *testing.T) {
	srv := canvas.NewCanvasServer(false)
	tool := NewCanvasTool(srv)
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "snapshot",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(strings.ToLower(result), "empty") {
		t.Errorf("empty snapshot should say 'empty', got: %s", result)
	}
}

func TestCanvasToolUnknownOperation(t *testing.T) {
	srv := canvas.NewCanvasServer(false)
	tool := NewCanvasTool(srv)
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "delete",
	})
	if err == nil {
		t.Error("expected error for unknown operation")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention 'unknown', got: %s", err.Error())
	}
}

func TestCanvasToolPushEmptyContent(t *testing.T) {
	srv := canvas.NewCanvasServer(false)
	tool := NewCanvasTool(srv)
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "push",
		"content":   "",
	})
	if err == nil {
		t.Error("expected error for empty content on push")
	}
}

func TestCanvasToolEvalEmptyContent(t *testing.T) {
	srv := canvas.NewCanvasServer(true)
	tool := NewCanvasTool(srv)
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"operation": "eval",
		"content":   "",
	})
	if err == nil {
		t.Error("expected error for empty content on eval")
	}
}

func TestCanvasToolDescription(t *testing.T) {
	tool := NewCanvasTool(nil)
	if tool.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

func TestCanvasToolParameters(t *testing.T) {
	tool := NewCanvasTool(nil)
	params := tool.Parameters()
	if params == nil {
		t.Fatal("Parameters() should not be nil")
	}
	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Parameters should have properties")
	}
	for _, field := range []string{"operation", "content", "format"} {
		if _, ok := props[field]; !ok {
			t.Errorf("missing parameter %q", field)
		}
	}
}
