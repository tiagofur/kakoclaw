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
