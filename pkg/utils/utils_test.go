package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Truncate
// ---------------------------------------------------------------------------

func TestTruncate_Empty(t *testing.T) {
	if got := Truncate("", 10); got != "" {
		t.Errorf("Truncate(\"\", 10) = %q; want \"\"", got)
	}
}

func TestTruncate_WithinLimit(t *testing.T) {
	if got := Truncate("hello", 10); got != "hello" {
		t.Errorf("Truncate(\"hello\", 10) = %q; want \"hello\"", got)
	}
}

func TestTruncate_ExactLimit(t *testing.T) {
	if got := Truncate("hello", 5); got != "hello" {
		t.Errorf("Truncate(\"hello\", 5) = %q; want \"hello\"", got)
	}
}

func TestTruncate_Truncated(t *testing.T) {
	got := Truncate("hello world", 8)
	want := "hello..."
	if got != want {
		t.Errorf("Truncate(\"hello world\", 8) = %q; want %q", got, want)
	}
}

func TestTruncate_MaxLenLessThanOrEqual3(t *testing.T) {
	got := Truncate("hello", 3)
	want := "hel"
	if got != want {
		t.Errorf("Truncate(\"hello\", 3) = %q; want %q", got, want)
	}

	got = Truncate("hello", 1)
	want = "h"
	if got != want {
		t.Errorf("Truncate(\"hello\", 1) = %q; want %q", got, want)
	}
}

func TestTruncate_Unicode(t *testing.T) {
	// 5 runes: こんにちは
	got := Truncate("こんにちは世界", 5)
	want := "こん..."
	if got != want {
		t.Errorf("Truncate(\"こんにちは世界\", 5) = %q; want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// IsAudioFile
// ---------------------------------------------------------------------------

func TestIsAudioFile_ByExtension(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"track.mp3", true},
		{"track.MP3", true},
		{"sound.wav", true},
		{"voice.ogg", true},
		{"song.m4a", true},
		{"music.flac", true},
		{"audio.aac", true},
		{"clip.wma", true},
		{"photo.jpg", false},
		{"video.mp4", false},
		{"doc.pdf", false},
	}
	for _, tt := range tests {
		if got := IsAudioFile(tt.filename, ""); got != tt.want {
			t.Errorf("IsAudioFile(%q, \"\") = %v; want %v", tt.filename, got, tt.want)
		}
	}
}

func TestIsAudioFile_ByContentType(t *testing.T) {
	tests := []struct {
		contentType string
		want        bool
	}{
		{"audio/mpeg", true},
		{"audio/wav", true},
		{"Audio/OGG", true},
		{"application/ogg", true},
		{"application/x-ogg", true},
		{"text/plain", false},
		{"image/png", false},
		{"video/mp4", false},
	}
	for _, tt := range tests {
		if got := IsAudioFile("unknown", tt.contentType); got != tt.want {
			t.Errorf("IsAudioFile(\"unknown\", %q) = %v; want %v", tt.contentType, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// SanitizeFilename
// ---------------------------------------------------------------------------

func TestSanitizeFilename_Normal(t *testing.T) {
	if got := SanitizeFilename("report.pdf"); got != "report.pdf" {
		t.Errorf("SanitizeFilename(\"report.pdf\") = %q; want \"report.pdf\"", got)
	}
}

func TestSanitizeFilename_PathTraversal(t *testing.T) {
	got := SanitizeFilename("../../etc/passwd")
	if got == "../../etc/passwd" {
		t.Error("SanitizeFilename should strip path traversal sequences")
	}
	// Should not contain ".."
	if got == "" {
		t.Error("SanitizeFilename should return a non-empty result")
	}
}

func TestSanitizeFilename_Backslashes(t *testing.T) {
	got := SanitizeFilename(`C:\Users\evil\file.txt`)
	// filepath.Base on Windows returns "file.txt"
	if got == "" {
		t.Error("SanitizeFilename should return a non-empty result")
	}
}

// ---------------------------------------------------------------------------
// Context helpers: WithAgentName / AgentNameFrom
// ---------------------------------------------------------------------------

func TestAgentNameContext_RoundTrip(t *testing.T) {
	ctx := context.Background()
	ctx = WithAgentName(ctx, "code-reviewer")

	if got := AgentNameFrom(ctx); got != "code-reviewer" {
		t.Errorf("AgentNameFrom = %q; want \"code-reviewer\"", got)
	}
}

func TestAgentNameContext_Empty(t *testing.T) {
	ctx := context.Background()
	if got := AgentNameFrom(ctx); got != "" {
		t.Errorf("AgentNameFrom on empty context = %q; want \"\"", got)
	}
}

// ---------------------------------------------------------------------------
// DownloadFile / DownloadFileSimple
// ---------------------------------------------------------------------------

func TestDownloadFile(t *testing.T) {
	const fileContent = "hello, this is test file content"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fileContent))
	}))
	defer srv.Close()

	path := DownloadFile(srv.URL+"/testfile.txt", "testfile.txt", DownloadOptions{
		Timeout: 5 * time.Second,
	})

	if path == "" {
		t.Fatal("DownloadFile returned empty string; expected a valid file path")
	}
	t.Cleanup(func() { os.Remove(path) })

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != fileContent {
		t.Errorf("file contents = %q; want %q", string(data), fileContent)
	}
}

func TestDownloadFileSimple(t *testing.T) {
	const fileContent = "simple download content"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fileContent))
	}))
	defer srv.Close()

	path := DownloadFileSimple(srv.URL+"/data.bin", "data.bin")

	if path == "" {
		t.Fatal("DownloadFileSimple returned empty string; expected a valid file path")
	}
	t.Cleanup(func() { os.Remove(path) })

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != fileContent {
		t.Errorf("file contents = %q; want %q", string(data), fileContent)
	}
}

func TestDownloadFile_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer srv.Close()

	path := DownloadFile(srv.URL+"/missing.txt", "missing.txt", DownloadOptions{
		Timeout: 5 * time.Second,
	})

	if path != "" {
		os.Remove(path)
		t.Errorf("DownloadFile with 404 returned path %q; want empty string", path)
	}
}

func TestDownloadFile_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay longer than the client timeout
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("too late"))
	}))
	defer srv.Close()

	path := DownloadFile(srv.URL+"/slow.txt", "slow.txt", DownloadOptions{
		Timeout: 100 * time.Millisecond,
	})

	if path != "" {
		os.Remove(path)
		t.Errorf("DownloadFile with timeout returned path %q; want empty string", path)
	}
}
