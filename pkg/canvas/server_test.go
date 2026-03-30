package canvas

import (
	"testing"
)

func TestCanvasServerPushSnapshot(t *testing.T) {
	srv := NewCanvasServer(false)
	srv.Push("<h1>hello</h1>")
	if got := srv.Snapshot(); got != "<h1>hello</h1>" {
		t.Errorf("Snapshot() = %q, want %q", got, "<h1>hello</h1>")
	}
}

func TestCanvasServerEvalBlockedWhenDevModeOff(t *testing.T) {
	srv := NewCanvasServer(false)
	_, err := srv.Eval("alert(1)")
	if err == nil {
		t.Error("expected error when dev_mode is off")
	}
	if err != nil && !contains(err.Error(), "dev_mode") {
		t.Errorf("error should mention dev_mode, got: %s", err.Error())
	}
}

func TestCanvasServerEvalAllowedWhenDevModeOn(t *testing.T) {
	srv := NewCanvasServer(true)
	result, err := srv.Eval("alert(1)")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "eval sent" {
		t.Errorf("result = %q, want %q", result, "eval sent")
	}
}

func TestCanvasServerReset(t *testing.T) {
	srv := NewCanvasServer(false)
	srv.Push("<p>data</p>")
	srv.Reset()
	if got := srv.Snapshot(); got != "" {
		t.Errorf("Snapshot() after Reset() = %q, want empty", got)
	}
}

func TestCanvasServerMultiplePushes(t *testing.T) {
	srv := NewCanvasServer(false)
	srv.Push("<p>first</p>")
	srv.Push("<p>second</p>")
	if got := srv.Snapshot(); got != "<p>second</p>" {
		t.Errorf("Snapshot() = %q, want %q", got, "<p>second</p>")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
