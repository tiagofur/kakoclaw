package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/storage"
	"github.com/sipeed/makoclaw/pkg/tools"
)

func setupTestServerWithAudit(t *testing.T) (*Server, *storage.Storage) {
	t.Helper()
	dir := t.TempDir()
	store, err := storage.New(config.StorageConfig{Path: filepath.Join(dir, "test.db")})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	// Insert directly to audit table bypassing FK check
	_, _ = store.ExecRaw(`PRAGMA foreign_keys=OFF`)

	s := &Server{store: store}
	return s, store
}

func TestHandleToolAuditExportJSON(t *testing.T) {
	s, store := setupTestServerWithAudit(t)

	// Seed an audit log entry
	auditLogger, err := tools.NewSQLiteAuditLogger(store)
	if err != nil {
		t.Fatalf("NewSQLiteAuditLogger: %v", err)
	}
	if err := auditLogger.LogToolExecution(context.Background(), tools.ToolExecutionLog{
		UserID:    1,
		Username:  "testuser",
		Tool:      "exec",
		Arguments: map[string]interface{}{"cmd": "ls"},
		Success:   true,
		Duration:  42,
	}); err != nil {
		t.Fatalf("LogToolExecution: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tools/audit/export?format=json", nil)
	req = req.WithContext(context.WithValue(req.Context(), userClaimsKey, &jwtClaims{Role: "admin"}))
	w := httptest.NewRecorder()

	s.handleToolAuditExport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); cd == "" {
		t.Fatal("expected Content-Disposition header")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if result["count"].(float64) != 1 {
		t.Fatalf("expected 1 log entry, got %v", result["count"])
	}
}

func TestHandleToolAuditExportCSV(t *testing.T) {
	s, store := setupTestServerWithAudit(t)

	auditLogger, err := tools.NewSQLiteAuditLogger(store)
	if err != nil {
		t.Fatalf("NewSQLiteAuditLogger: %v", err)
	}
	if err := auditLogger.LogToolExecution(context.Background(), tools.ToolExecutionLog{
		UserID:    1,
		Username:  "testuser",
		Tool:      "web_search",
		Arguments: map[string]interface{}{"query": "test"},
		Success:   true,
		Duration:  100,
	}); err != nil {
		t.Fatalf("LogToolExecution: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tools/audit/export?format=csv", nil)
	req = req.WithContext(context.WithValue(req.Context(), userClaimsKey, &jwtClaims{Role: "admin"}))
	w := httptest.NewRecorder()

	s.handleToolAuditExport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/csv" {
		t.Fatalf("expected text/csv, got %s", ct)
	}
	body := w.Body.String()
	if len(body) == 0 {
		t.Fatal("expected non-empty CSV body")
	}
	// Should contain header and at least one data row
	if !contains(body, "id,timestamp,user_id,username,tool,success,error,duration_ms") {
		t.Fatal("CSV missing header row")
	}
	if !contains(body, "testuser") {
		t.Fatal("CSV missing data row")
	}
}

func TestHandleToolAuditExportForbidden(t *testing.T) {
	s, _ := setupTestServerWithAudit(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tools/audit/export", nil)
	req = req.WithContext(context.WithValue(req.Context(), userClaimsKey, &jwtClaims{Role: "user"}))
	w := httptest.NewRecorder()

	s.handleToolAuditExport(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestHandleToolAuditExportEmpty(t *testing.T) {
	s, _ := setupTestServerWithAudit(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tools/audit/export?format=json", nil)
	req = req.WithContext(context.WithValue(req.Context(), userClaimsKey, &jwtClaims{Role: "admin"}))
	w := httptest.NewRecorder()

	s.handleToolAuditExport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["count"].(float64) != 0 {
		t.Fatalf("expected 0 logs, got %v", result["count"])
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
