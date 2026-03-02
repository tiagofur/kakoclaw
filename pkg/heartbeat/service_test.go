package heartbeat

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewHeartbeatService(t *testing.T) {
	hs := NewHeartbeatService("/tmp/test", nil, 60, true)
	if hs == nil {
		t.Fatal("NewHeartbeatService returned nil")
	}
	if hs.workspace != "/tmp/test" {
		t.Errorf("workspace = %q; want \"/tmp/test\"", hs.workspace)
	}
	if hs.interval != 60*time.Second {
		t.Errorf("interval = %v; want 60s", hs.interval)
	}
	if !hs.enabled {
		t.Error("enabled should be true")
	}
}

func TestStart_Disabled(t *testing.T) {
	called := false
	hs := NewHeartbeatService("", func(prompt string) (string, error) {
		called = true
		return "", nil
	}, 1, false)

	// On a fresh (never-stopped) service, running() returns true,
	// so Start() returns nil without launching the goroutine.
	// We verify the heartbeat callback is never invoked.
	hs.Start()
	time.Sleep(1200 * time.Millisecond)

	if called {
		t.Error("Heartbeat callback should not be called when disabled")
	}
}

func TestStart_Stop_Lifecycle(t *testing.T) {
	callCount := 0
	hs := NewHeartbeatService("", func(prompt string) (string, error) {
		callCount++
		return "ok", nil
	}, 1, true)

	err := hs.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Let it tick once
	time.Sleep(1200 * time.Millisecond)

	hs.Stop()

	// Double-stop should not panic
	hs.Stop()
}

func TestStart_AlreadyRunning(t *testing.T) {
	hs := NewHeartbeatService("", nil, 60, true)
	err := hs.Start()
	if err != nil {
		t.Fatalf("First Start failed: %v", err)
	}
	defer hs.Stop()

	// Second start should return nil (already running)
	err = hs.Start()
	if err != nil {
		t.Errorf("Second Start should return nil; got %v", err)
	}
}

func TestBuildPrompt(t *testing.T) {
	dir := t.TempDir()
	hs := NewHeartbeatService(dir, nil, 60, true)

	prompt := hs.buildPrompt()
	if !strings.Contains(prompt, "Heartbeat Check") {
		t.Error("buildPrompt should contain 'Heartbeat Check'")
	}
}

func TestBuildPrompt_WithNotes(t *testing.T) {
	dir := t.TempDir()
	notesDir := filepath.Join(dir, "memory")
	os.MkdirAll(notesDir, 0755)
	os.WriteFile(filepath.Join(notesDir, "HEARTBEAT.md"), []byte("## Important note"), 0644)

	hs := NewHeartbeatService(dir, nil, 60, true)
	prompt := hs.buildPrompt()

	if !strings.Contains(prompt, "Important note") {
		t.Error("buildPrompt should include contents of HEARTBEAT.md")
	}
}

func TestCheckHeartbeat_CallsCallback(t *testing.T) {
	called := false
	hs := NewHeartbeatService("", func(prompt string) (string, error) {
		called = true
		return "done", nil
	}, 60, true)

	// Manually call checkHeartbeat (it checks enabled + running internally)
	// Need to have it still running (stopChan open)
	hs.checkHeartbeat()
	if !called {
		t.Error("checkHeartbeat should have called onHeartbeat")
	}
}

func TestCheckHeartbeat_ErrorHandled(t *testing.T) {
	dir := t.TempDir()
	hs := NewHeartbeatService(dir, func(prompt string) (string, error) {
		return "", fmt.Errorf("provider error")
	}, 60, true)

	// Should not panic even if callback errors
	hs.checkHeartbeat()
}

func TestLog(t *testing.T) {
	dir := t.TempDir()
	memDir := filepath.Join(dir, "memory")
	os.MkdirAll(memDir, 0755)

	hs := NewHeartbeatService(dir, nil, 60, true)
	hs.log("test message")

	logPath := filepath.Join(memDir, "heartbeat.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if !strings.Contains(string(data), "test message") {
		t.Error("Log file should contain the logged message")
	}
}
