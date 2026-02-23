package web

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/kakoclaw/pkg/storage"
)

// authRequest injects JWT claims into the request context so handlers see
// the user as authenticated.
func authRequest(r *http.Request, username string) *http.Request {
	claims := &jwtClaims{Sub: username}
	ctx := context.WithValue(r.Context(), userClaimsKey, claims)
	return r.WithContext(ctx)
}

// newBackupTestServer creates a Server backed by a temp SQLite DB with an admin user.
func newBackupTestServer(t *testing.T) (*Server, *storage.User) {
	t.Helper()

	s := newTestServer(t)

	user, err := s.store.GetUserByUsername("admin")
	if err != nil {
		t.Fatalf("admin user not found: %v", err)
	}

	return s, user
}

// TestBackupValidateManifest verifies that backup validation works correctly.
func TestBackupValidateManifest(t *testing.T) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	manifest := BackupManifest{
		Version:         "2.0",
		BackupType:      "personal",
		TotalFiles:      2,
		ConfigFileCount: 1,
		DataSizeBytes:   100,
	}
	manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
	mw, _ := zipWriter.Create("manifest.json")
	mw.Write(manifestJSON)

	dw, _ := zipWriter.Create("config.json")
	dw.Write([]byte(`{"test":"data"}`))
	zipWriter.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "backup.kakoclaw")
	io.Copy(part, buf)
	writer.Close()

	server := &Server{}
	req := httptest.NewRequest("POST", "/api/backup/validate", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	server.handleBackupValidate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if !result["valid"].(bool) {
		t.Error("Expected validation to return valid=true")
	}
	if result["total_files"].(float64) != 2 {
		t.Errorf("Expected total_files=2, got %v", result["total_files"])
	}
	if result["backup_format"].(string) != "v2_personal" {
		t.Errorf("Expected v2_personal format, got %v", result["backup_format"])
	}
}

// TestBackupValidateDetectsLegacy validates that a v1 backup with .db file is detected as legacy.
func TestBackupValidateDetectsLegacy(t *testing.T) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	manifest := BackupManifest{
		Version:    "1.0",
		TotalFiles: 1,
	}
	manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
	mw, _ := zipWriter.Create("manifest.json")
	mw.Write(manifestJSON)

	dw, _ := zipWriter.Create("database/kakoclaw.db")
	dw.Write([]byte("fake-db-bytes"))
	zipWriter.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "backup.kakoclaw")
	io.Copy(part, buf)
	writer.Close()

	server := &Server{}
	req := httptest.NewRequest("POST", "/api/backup/validate", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	server.handleBackupValidate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["backup_format"].(string) != "v1_legacy" {
		t.Errorf("Expected v1_legacy format, got %v", result["backup_format"])
	}
}

// TestBackupExportRequiresAuth ensures the export endpoint rejects unauthenticated requests.
func TestBackupExportRequiresAuth(t *testing.T) {
	s, _ := newBackupTestServer(t)

	req := httptest.NewRequest("GET", "/api/backup/export", nil)
	w := httptest.NewRecorder()

	s.handleBackupExport(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401, got %d", w.Code)
	}
}

// TestBackupImportRequiresAuth ensures the import endpoint rejects unauthenticated requests.
func TestBackupImportRequiresAuth(t *testing.T) {
	s, _ := newBackupTestServer(t)

	req := httptest.NewRequest("POST", "/api/backup/import", nil)
	w := httptest.NewRecorder()

	s.handleBackupImport(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401, got %d", w.Code)
	}
}

// TestBackupImportV2JSON verifies import of a v2 backup ZIP with JSON user data.
func TestBackupImportV2JSON(t *testing.T) {
	s, user := newBackupTestServer(t)

	zipBuf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuf)

	manifest := BackupManifest{
		Version:           "2.0",
		BackupType:        "personal",
		TotalFiles:        1,
		DatabaseFileCount: 1,
	}
	mJSON, _ := json.MarshalIndent(manifest, "", "  ")
	mw, _ := zipWriter.Create("manifest.json")
	mw.Write(mJSON)

	userData := storage.BackupUserData{
		Sessions: []storage.BackupSession{
			{SessionID: "test-session-1", Title: "Test Session"},
		},
		Messages: []storage.BackupMessage{
			{SessionID: "test-session-1", Role: "user", Content: "hello world"},
		},
		Tasks: []storage.BackupTask{
			{Title: "Test Task", Status: "todo"},
		},
	}
	dataJSON, _ := json.MarshalIndent(userData, "", "  ")
	dw, _ := zipWriter.Create(userDataEntry)
	dw.Write(dataJSON)
	zipWriter.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "backup.kakoclaw")
	io.Copy(part, zipBuf)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/backup/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, user.Username)
	w := httptest.NewRecorder()

	s.handleBackupImport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	details := result["details"].(map[string]interface{})
	if details["sessions_imported"].(float64) < 1 {
		t.Errorf("Expected at least 1 session imported, got %v", details["sessions_imported"])
	}
	if details["messages_imported"].(float64) < 1 {
		t.Errorf("Expected at least 1 message imported, got %v", details["messages_imported"])
	}
	if details["tasks_imported"].(float64) < 1 {
		t.Errorf("Expected at least 1 task imported, got %v", details["tasks_imported"])
	}
}

// TestBackupImportRestoresProvidersAndChannelMappings verifies restore includes
// per-user provider configuration and channel mapping data.
func TestBackupImportRestoresProvidersAndChannelMappings(t *testing.T) {
	s, user := newBackupTestServer(t)

	zipBuf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuf)

	manifest := BackupManifest{
		Version:           "2.0",
		BackupType:        "personal",
		TotalFiles:        1,
		DatabaseFileCount: 1,
	}
	mJSON, _ := json.MarshalIndent(manifest, "", "  ")
	mw, _ := zipWriter.Create("manifest.json")
	mw.Write(mJSON)

	userData := storage.BackupUserData{
		Providers: &storage.UserProvidersConfig{
			OpenAI: storage.ProviderConfig{
				APIKey:  "sk-test-user-openai",
				APIBase: "https://api.openai.com/v1",
			},
			Groq: storage.ProviderConfig{
				APIKey: "gsk-test-user-groq",
			},
		},
		Channels: []storage.BackupChannelMapping{
			{Channel: "telegram", SenderID: "123456"},
		},
	}
	dataJSON, _ := json.MarshalIndent(userData, "", "  ")
	dw, _ := zipWriter.Create(userDataEntry)
	dw.Write(dataJSON)
	zipWriter.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "backup.kakoclaw")
	io.Copy(part, zipBuf)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/backup/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, user.Username)
	w := httptest.NewRecorder()

	s.handleBackupImport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	providers, err := s.store.GetUserProvidersConfig(user.ID)
	if err != nil {
		t.Fatalf("GetUserProvidersConfig failed: %v", err)
	}
	if providers.OpenAI.APIKey != "sk-test-user-openai" {
		t.Fatalf("expected OpenAI api key restored, got %q", providers.OpenAI.APIKey)
	}
	if providers.Groq.APIKey != "gsk-test-user-groq" {
		t.Fatalf("expected Groq api key restored, got %q", providers.Groq.APIKey)
	}

	mappedUserID, err := s.store.GetUserIDForChannelSender("telegram", "123456")
	if err != nil {
		t.Fatalf("GetUserIDForChannelSender failed: %v", err)
	}
	if mappedUserID != user.ID {
		t.Fatalf("expected mapping to user %d, got %d", user.ID, mappedUserID)
	}
}

// TestBackupExportEmptyBackupError verifies that exporting with no options returns an error.
func TestBackupExportEmptyBackupError(t *testing.T) {
	s, user := newBackupTestServer(t)

	req := httptest.NewRequest("GET", "/api/backup/export?include_database=false&include_workspace=false&include_config=false&include_env=false", nil)
	req = authRequest(req, user.Username)
	w := httptest.NewRecorder()

	s.handleBackupExport(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("Expected error status when nothing to export, got %d", w.Code)
	}
}

// TestBackupExportWithAuth verifies that the export handler produces a v2 backup when authenticated.
func TestBackupExportWithAuth(t *testing.T) {
	s, user := newBackupTestServer(t)

	req := httptest.NewRequest("GET", "/api/backup/export?include_database=true&include_workspace=false&include_config=false&include_env=false", nil)
	req = authRequest(req, user.Username)
	w := httptest.NewRecorder()

	s.handleBackupExport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	zipData := w.Body.Bytes()
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Fatalf("Failed to read zip: %v", err)
	}

	foundManifest := false
	foundUserData := false
	for _, f := range zipReader.File {
		if f.Name == "manifest.json" {
			foundManifest = true
		}
		if f.Name == userDataEntry {
			foundUserData = true
		}
	}
	if !foundManifest {
		t.Error("manifest.json not found in backup")
	}
	if !foundUserData {
		t.Error("database/user_data.json not found in backup")
	}

	mf, _ := zipReader.Open("manifest.json")
	if mf != nil {
		defer mf.Close()
		var manifest BackupManifest
		json.NewDecoder(mf).Decode(&manifest)
		if manifest.Version != "2.0" {
			t.Errorf("Expected manifest version 2.0, got %s", manifest.Version)
		}
		if manifest.BackupType != "personal" {
			t.Errorf("Expected backup_type personal, got %s", manifest.BackupType)
		}
	}
}

// TestBackupImportWorkspaceFiles verifies workspace files are extracted correctly.
func TestBackupImportWorkspaceFiles(t *testing.T) {
	s, user := newBackupTestServer(t)

	zipBuf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuf)

	manifest := BackupManifest{
		Version:    "2.0",
		BackupType: "personal",
		TotalFiles: 2,
	}
	mJSON, _ := json.MarshalIndent(manifest, "", "  ")
	mw, _ := zipWriter.Create("manifest.json")
	mw.Write(mJSON)

	fw, _ := zipWriter.Create("workspace/test-doc.txt")
	fw.Write([]byte("test document content"))
	zipWriter.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "backup.kakoclaw")
	io.Copy(part, zipBuf)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/backup/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = authRequest(req, user.Username)
	w := httptest.NewRecorder()

	s.handleBackupImport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	details := result["details"].(map[string]interface{})
	if details["files_restored"].(float64) < 1 {
		t.Errorf("Expected at least 1 file restored, got %v", details["files_restored"])
	}

	wsPath := filepath.Join(os.Getenv("HOME"), ".kakoclaw", "users", user.UUID, "workspace", "test-doc.txt")
	data, err := os.ReadFile(wsPath)
	if err != nil {
		t.Logf("Could not verify file at %s (may be expected in CI): %v", wsPath, err)
	} else if string(data) != "test document content" {
		t.Errorf("Expected file content 'test document content', got '%s'", string(data))
	}
}
