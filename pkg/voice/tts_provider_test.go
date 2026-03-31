package voice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestElevenLabsProviderIsAvailable(t *testing.T) {
	p := &ElevenLabsProvider{}
	if p.IsAvailable() {
		t.Error("expected not available without API key")
	}

	p2 := &ElevenLabsProvider{apiKey: "key"}
	if !p2.IsAvailable() {
		t.Error("expected available with API key")
	}
}

func TestSystemTTSProviderIsAvailable(t *testing.T) {
	p := &SystemTTSProvider{}
	if !p.IsAvailable() {
		t.Error("expected system TTS to always be available")
	}
}

func TestElevenLabsSynthesizeEmptyText(t *testing.T) {
	p := NewElevenLabsProvider("test-key", "test-voice")
	_, err := p.Synthesize(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestElevenLabsSynthesizeCallsAPI(t *testing.T) {
	fakeAudio := []byte("fake-elevenlabs-mp3-data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path contains the voice ID
		if !strings.Contains(r.URL.Path, "/v1/text-to-speech/test-voice") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// Verify API key header
		if r.Header.Get("xi-api-key") != "test-key" {
			t.Errorf("bad xi-api-key header: %s", r.Header.Get("xi-api-key"))
		}
		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("bad content type: %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write(fakeAudio)
	}))
	defer server.Close()

	p := NewElevenLabsProvider("test-key", "test-voice")
	p.apiBase = server.URL

	audio, err := p.Synthesize(context.Background(), "Hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(audio) != string(fakeAudio) {
		t.Errorf("audio = %q, want %q", audio, fakeAudio)
	}
}

func TestElevenLabsSynthesizeAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid api key"}`))
	}))
	defer server.Close()

	p := NewElevenLabsProvider("bad-key", "voice")
	p.apiBase = server.URL

	_, err := p.Synthesize(context.Background(), "Hello")
	if err == nil {
		t.Error("expected error for API failure")
	}
}
