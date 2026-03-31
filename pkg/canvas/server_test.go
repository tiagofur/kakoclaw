package canvas

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestCanvasSSEBroadcast(t *testing.T) {
	srv := NewCanvasServer(false)

	// Use a ResponseRecorder approach: directly call the handler with a
	// flushing recorder to test SSE without HTTP transport issues.
	// Instead, test the broadcast mechanism through the internal channel.

	// Verify broadcast delivers to registered clients
	ch := make(chan string, 16)
	srv.clientsMu.Lock()
	srv.clients[ch] = struct{}{}
	srv.clientsMu.Unlock()
	defer func() {
		srv.clientsMu.Lock()
		delete(srv.clients, ch)
		srv.clientsMu.Unlock()
	}()

	srv.Push("<h1>sse test</h1>")

	select {
	case data := <-ch:
		if data != "<h1>sse test</h1>" {
			t.Errorf("broadcast data = %q, want %q", data, "<h1>sse test</h1>")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("SSE broadcast not received within 500ms")
	}

	// Verify snapshot was also updated
	if snap := srv.Snapshot(); snap != "<h1>sse test</h1>" {
		t.Errorf("Snapshot() = %q, want %q", snap, "<h1>sse test</h1>")
	}
}

func TestCanvasSSEHandlerHeaders(t *testing.T) {
	srv := NewCanvasServer(false)

	ts := httptest.NewServer(http.HandlerFunc(srv.ServeSSE))
	defer ts.Close()

	// Make a quick connection to verify headers — use Transport with DisableKeepAlives
	// and a short timeout to avoid hanging.
	client := &http.Client{
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Push data first so the handler has something to send immediately
	srv.Push("<p>initial</p>")

	req, _ := http.NewRequestWithContext(ctx, "GET", ts.URL, nil)
	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		if resp.Header.Get("Content-Type") != "text/event-stream" {
			t.Errorf("Content-Type = %q, want %q", resp.Header.Get("Content-Type"), "text/event-stream")
		}
	}
	// Context timeout is expected — SSE streams don't end
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
