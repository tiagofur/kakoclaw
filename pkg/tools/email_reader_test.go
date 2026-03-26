package tools

import (
	"strings"
	"testing"
)

func TestExtractEmailPlainText_MultipartAlternativePrefersPlainText(t *testing.T) {
	raw := strings.Join([]string{
		"From: sender@example.com",
		"To: receiver@example.com",
		"Subject: Test",
		"MIME-Version: 1.0",
		"Content-Type: multipart/alternative; boundary=\"b1\"",
		"",
		"--b1",
		"Content-Type: text/plain; charset=\"UTF-8\"",
		"",
		"Hola en texto plano",
		"--b1",
		"Content-Type: text/html; charset=\"UTF-8\"",
		"",
		"<html><body><p>Hola en <b>HTML</b></p></body></html>",
		"--b1--",
		"",
	}, "\r\n")

	body, err := extractEmailPlainText(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("extractEmailPlainText returned error: %v", err)
	}
	if got := strings.TrimSpace(body); got != "Hola en texto plano" {
		t.Fatalf("expected plain text part, got %q", got)
	}
}

func TestExtractEmailPlainText_HtmlOnlyFallback(t *testing.T) {
	raw := strings.Join([]string{
		"From: sender@example.com",
		"To: receiver@example.com",
		"Subject: HTML only",
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
		"",
		"<html><body><div>Proxima carrera F1</div></body></html>",
		"",
	}, "\r\n")

	body, err := extractEmailPlainText(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("extractEmailPlainText returned error: %v", err)
	}
	if !strings.Contains(body, "Proxima carrera F1") {
		t.Fatalf("expected html fallback text, got %q", body)
	}
}

func TestExtractEmailPlainText_RawFallbackWhenMalformed(t *testing.T) {
	raw := "not a valid mime\n\ncontenido de respaldo"

	body, err := extractEmailPlainText(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("extractEmailPlainText returned error: %v", err)
	}
	if got := strings.TrimSpace(body); got != "contenido de respaldo" {
		t.Fatalf("expected raw fallback body, got %q", got)
	}
}
