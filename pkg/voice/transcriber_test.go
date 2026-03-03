package voice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGroqTranscriber(t *testing.T) {
	tr := NewGroqTranscriber("test-key")
	if tr == nil {
		t.Fatal("NewGroqTranscriber returned nil")
	}
	if tr.apiBase != "https://api.groq.com/openai/v1" {
		t.Errorf("unexpected apiBase: %q", tr.apiBase)
	}
	if tr.apiKey != "test-key" {
		t.Errorf("unexpected apiKey: %q", tr.apiKey)
	}
	if tr.httpClient == nil {
		t.Fatal("httpClient should not be nil")
	}
}

func TestIsAvailable_WithKey(t *testing.T) {
	tr := NewGroqTranscriber("test-key")
	if !tr.IsAvailable() {
		t.Error("expected IsAvailable() == true when key is set")
	}
}

func TestIsAvailable_WithoutKey(t *testing.T) {
	tr := NewGroqTranscriber("")
	if tr.IsAvailable() {
		t.Error("expected IsAvailable() == false when key is empty")
	}
}

func TestTranscribe_FileNotFound(t *testing.T) {
	tr := NewGroqTranscriber("test-key")
	_, err := tr.Transcribe(context.Background(), "/nonexistent/path/audio.wav")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to open audio file") {
		t.Errorf("expected 'failed to open audio file' in error, got: %v", err)
	}
}

func TestTranscribe_Success(t *testing.T) {
	// Mock Groq API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request basics
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/audio/transcriptions") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("unexpected Authorization header: %s", auth)
		}
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "multipart/form-data") {
			t.Errorf("expected multipart/form-data content type, got: %s", contentType)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"text":"hello world","language":"en","duration":1.5}`))
	}))
	defer server.Close()

	tr := NewGroqTranscriber("test-key")
	tr.apiBase = server.URL
	tr.httpClient = server.Client()

	// Create a temporary audio file with dummy content
	tmpDir := t.TempDir()
	audioFile := filepath.Join(tmpDir, "test.wav")
	if err := os.WriteFile(audioFile, []byte("fake audio data"), 0644); err != nil {
		t.Fatalf("failed to create temp audio file: %v", err)
	}

	result, err := tr.Transcribe(context.Background(), audioFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Text != "hello world" {
		t.Errorf("expected text %q, got %q", "hello world", result.Text)
	}
	if result.Language != "en" {
		t.Errorf("expected language %q, got %q", "en", result.Language)
	}
	if result.Duration != 1.5 {
		t.Errorf("expected duration 1.5, got %f", result.Duration)
	}
}

func TestTranscribe_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":{"message":"invalid audio format","type":"invalid_request_error"}}`))
	}))
	defer server.Close()

	tr := NewGroqTranscriber("test-key")
	tr.apiBase = server.URL
	tr.httpClient = server.Client()

	tmpDir := t.TempDir()
	audioFile := filepath.Join(tmpDir, "bad.wav")
	if err := os.WriteFile(audioFile, []byte("not real audio"), 0644); err != nil {
		t.Fatalf("failed to create temp audio file: %v", err)
	}

	_, err := tr.Transcribe(context.Background(), audioFile)
	if err == nil {
		t.Fatal("expected error for API 400 response, got nil")
	}
	if !strings.Contains(err.Error(), "API error") {
		t.Errorf("expected 'API error' in error message, got: %v", err)
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected status code 400 in error message, got: %v", err)
	}
}
