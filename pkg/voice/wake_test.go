package voice

import (
	"context"
	"testing"
	"time"
)

func TestWakeDetectorCancelStopsListen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wd := NewWakeDetector()
	done := make(chan struct{})
	go func() {
		wd.Listen(ctx, "hey mako", 0.5, func() {})
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Listen did not stop after context cancel")
	}
}
