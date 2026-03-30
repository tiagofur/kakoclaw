package voice

import (
	"testing"
)

func TestTalkModeSessionStopIdempotent(t *testing.T) {
	sess := NewTalkModeSession(nil, nil)
	if !sess.IsActive() {
		t.Error("expected session to be active after creation")
	}
	sess.Stop()
	if sess.IsActive() {
		t.Error("expected session to be inactive after Stop()")
	}
	// Must not panic on second Stop()
	sess.Stop()
	if sess.IsActive() {
		t.Error("expected session to remain inactive after second Stop()")
	}
}

func TestTalkModeSessionProviders(t *testing.T) {
	sess := NewTalkModeSession(nil, nil)
	if sess.stt != nil {
		t.Error("expected nil STT provider")
	}
	if sess.tts != nil {
		t.Error("expected nil TTS provider")
	}
}
