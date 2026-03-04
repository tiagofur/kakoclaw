package voice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
)

func TestNewTTSSynthesizer(t *testing.T) {
	synth := NewTTSSynthesizer(config.TTSConfig{
		APIKey: "test-key",
		Voice:  "nova",
		Model:  "tts-1-hd",
	})
	if synth.voice != "nova" {
		t.Errorf("voice = %q, want %q", synth.voice, "nova")
	}
	if synth.model != "tts-1-hd" {
		t.Errorf("model = %q, want %q", synth.model, "tts-1-hd")
	}
}

func TestTTSSynthesizerDefaults(t *testing.T) {
	synth := NewTTSSynthesizer(config.TTSConfig{APIKey: "key"})
	if synth.voice != "alloy" {
		t.Errorf("default voice = %q, want %q", synth.voice, "alloy")
	}
	if synth.model != "tts-1" {
		t.Errorf("default model = %q, want %q", synth.model, "tts-1")
	}
}

func TestTTSSynthesizerIsAvailable(t *testing.T) {
	withKey := NewTTSSynthesizer(config.TTSConfig{APIKey: "key"})
	if !withKey.IsAvailable() {
		t.Error("expected available with API key")
	}

	withoutKey := NewTTSSynthesizer(config.TTSConfig{})
	if withoutKey.IsAvailable() {
		t.Error("expected not available without API key")
	}
}

func TestTTSSynthesizeEmptyText(t *testing.T) {
	synth := NewTTSSynthesizer(config.TTSConfig{APIKey: "key"})
	_, err := synth.Synthesize(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestTTSSynthesize_MockServer(t *testing.T) {
	fakeAudio := []byte("fake-mp3-data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audio/speech" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("bad auth header: %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write(fakeAudio)
	}))
	defer server.Close()

	synth := NewTTSSynthesizer(config.TTSConfig{APIKey: "test-key"})
	synth.apiBase = server.URL

	audio, err := synth.Synthesize(context.Background(), "Hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(audio) != string(fakeAudio) {
		t.Errorf("audio = %q, want %q", audio, fakeAudio)
	}
}

func TestTTSSynthesize_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid api key"}`))
	}))
	defer server.Close()

	synth := NewTTSSynthesizer(config.TTSConfig{APIKey: "bad-key"})
	synth.apiBase = server.URL

	_, err := synth.Synthesize(context.Background(), "Hello")
	if err == nil {
		t.Error("expected error for API failure")
	}
}
