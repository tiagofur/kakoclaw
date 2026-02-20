package web

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestBackupExportCorrectPath verifies that backup exports from the correct data directory
func TestBackupExportCorrectPath(t *testing.T) {
	// Create temporary directories to simulate workspace structure
	tempDir := t.TempDir()
	dataDir := tempDir
	workspaceDir := filepath.Join(dataDir, "workspace")

	// Create test files
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Create config.json
	configPath := filepath.Join(dataDir, "config.json")
	if err := ioutil.WriteFile(configPath, []byte(`{"test": "data"}`), 0644); err != nil {
		t.Fatalf("Failed to create config.json: %v", err)
	}

	// Create workspace files (e.g., sessions)
	sessionPath := filepath.Join(workspaceDir, "test-session.json")
	if err := ioutil.WriteFile(sessionPath, []byte(`{"session": "data"}`), 0644); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	// Create server with proper workspace
	server := &Server{
		workspace: workspaceDir,
	}

	// Make request
	req := httptest.NewRequest("GET", "/api/backup/export?include_database=false&include_workspace=true&include_config=true", nil)
	w := httptest.NewRecorder()

	server.handleBackupExport(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify it's a valid ZIP
	zipData := w.Body.Bytes()
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Fatalf("Failed to read zip: %v", err)
	}

	// Check that expected files are in the ZIP
	foundFiles := make(map[string]bool)
	for _, file := range zipReader.File {
		foundFiles[file.Name] = true
	}

	// Verify manifest exists
	if !foundFiles["manifest.json"] {
		t.Error("manifest.json not found in backup")
	}

	// Verify config.json exists
	if !foundFiles["config.json"] {
		t.Error("config.json not found in backup")
	}

	// Verify workspace files exist
	if !foundFiles["workspace/test-session.json"] {
		t.Error("workspace/test-session.json not found in backup")
	}

	// Verify manifest contents
	manifestFile, err := zipReader.Open("manifest.json")
	if err == nil {
		defer manifestFile.Close()
		var manifest BackupManifest
		if err := json.NewDecoder(manifestFile).Decode(&manifest); err != nil {
			t.Fatalf("Failed to decode manifest: %v", err)
		}

		if manifest.TotalFiles <= 0 {
			t.Error("Expected TotalFiles > 0 in manifest")
		}

		if manifest.ConfigFileCount != 1 {
			t.Errorf("Expected ConfigFileCount=1, got %d", manifest.ConfigFileCount)
		}

		if manifest.WorkspaceFileCount <= 0 {
			t.Errorf("Expected WorkspaceFileCount > 0, got %d", manifest.WorkspaceFileCount)
		}

		if len(manifest.ExportedFiles) == 0 {
			t.Error("Expected ExportedFiles to be populated in manifest")
		}
	}
}

// TestBackupExportEmptyBackupError verifies that exporting with no data returns error
func TestBackupExportEmptyBackupError(t *testing.T) {
	// Create temporary directory with no files
	tempDir := t.TempDir()
	workspaceDir := filepath.Join(tempDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)

	server := &Server{
		workspace: workspaceDir,
	}

	// Try to export with no options enabled (everything is false by default)
	req := httptest.NewRequest("GET", "/api/backup/export?include_database=false&include_workspace=false&include_config=false&include_env=false", nil)
	w := httptest.NewRecorder()

	server.handleBackupExport(w, req)

	// Should return error
	if w.Code == http.StatusOK {
		t.Fatalf("Expected error status when nothing to export, got %d", w.Code)
	}
}

// TestBackupExportIncludesAllDirectories verifies that skills, cron directories are exported
func TestBackupExportIncludesAllDirectories(t *testing.T) {
	// Create temporary directories with all structure
	tempDir := t.TempDir()
	dataDir := tempDir
	workspaceDir := filepath.Join(dataDir, "workspace")
	skillsDir := filepath.Join(dataDir, "skills")
	cronDir := filepath.Join(dataDir, "cron")

	// Create all directories
	for _, dir := range []string{workspaceDir, skillsDir, cronDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files in each directory
	testFiles := map[string]string{
		filepath.Join(skillsDir, "test-skill.js"):  `console.log("test");`,
		filepath.Join(cronDir, "test-cron.json"):   `{"cron": "test"}`,
		filepath.Join(workspaceDir, "AGENTS.md"):   `# Agents`,
		filepath.Join(workspaceDir, "SOUL.md"):     `# Soul`,
		filepath.Join(workspaceDir, "USER.md"):     `# User`,
		filepath.Join(workspaceDir, "IDENTITY.md"): `# Identity`,
	}

	for path, content := range testFiles {
		if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	server := &Server{
		workspace: workspaceDir,
	}

	// Export everything
	req := httptest.NewRequest("GET", "/api/backup/export?include_database=false&include_workspace=true&include_config=false", nil)
	w := httptest.NewRecorder()

	server.handleBackupExport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	// Verify ZIP contains all expected directories
	zipData := w.Body.Bytes()
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Fatalf("Failed to read zip: %v", err)
	}

	expectedPaths := []string{
		"skills/test-skill.js",
		"cron/test-cron.json",
		"workspace/AGENTS.md",
		"workspace/SOUL.md",
		"workspace/USER.md",
		"workspace/IDENTITY.md",
	}

	foundFiles := make(map[string]bool)
	for _, file := range zipReader.File {
		foundFiles[file.Name] = true
	}

	for _, path := range expectedPaths {
		if !foundFiles[path] {
			t.Errorf("Expected %s in backup, not found. Available files: %v", path, foundFiles)
		}
	}

	// Verify manifest reports correct file counts
	manifestFile, err := zipReader.Open("manifest.json")
	if err == nil {
		defer manifestFile.Close()
		var manifest BackupManifest
		if err := json.NewDecoder(manifestFile).Decode(&manifest); err != nil {
			t.Fatalf("Failed to decode manifest: %v", err)
		}

		if manifest.SkillsFileCount != 1 {
			t.Errorf("Expected SkillsFileCount=1, got %d", manifest.SkillsFileCount)
		}

		if manifest.CronFileCount != 1 {
			t.Errorf("Expected CronFileCount=1, got %d", manifest.CronFileCount)
		}

		if manifest.BootstrapFileCount != 4 {
			t.Errorf("Expected BootstrapFileCount=4, got %d", manifest.BootstrapFileCount)
		}
	}
}

// TestBackupValidateManifest verifies that backup validation works correctly
func TestBackupValidateManifest(t *testing.T) {
	// Create a test ZIP with manifest
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Create manifest
	manifest := BackupManifest{
		Version:         "1.0",
		TotalFiles:      2,
		ConfigFileCount: 1,
		DataSizeBytes:   100,
	}
	manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")

	manifestWriter, _ := zipWriter.Create("manifest.json")
	manifestWriter.Write(manifestJSON)

	// Add a dummy file
	dummyWriter, _ := zipWriter.Create("config.json")
	dummyWriter.Write([]byte(`{"test": "data"}`))

	zipWriter.Close()

	// Create multipart form with the ZIP
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", "backup.kakoclaw")
	io.Copy(part, buf)
	writer.Close()

	server := &Server{}

	// Make validate request
	req := httptest.NewRequest("POST", "/api/backup/validate", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	server.handleBackupValidate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
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
}

// TestBackupExportManifestAccuracy verifies manifest correctly reports what was exported
func TestBackupExportManifestAccuracy(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := tempDir
	workspaceDir := filepath.Join(dataDir, "workspace")

	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Create multiple workspace files to test counting
	for i := 0; i < 5; i++ {
		sessionPath := filepath.Join(workspaceDir, fmt.Sprintf("session-%d.json", i))
		ioutil.WriteFile(sessionPath, []byte(fmt.Sprintf(`{"id":%d}`, i)), 0644)
	}

	// Create config
	ioutil.WriteFile(filepath.Join(dataDir, "config.json"), []byte(`{}`), 0644)

	server := &Server{
		workspace: workspaceDir,
	}

	req := httptest.NewRequest("GET", "/api/backup/export?include_database=false&include_workspace=true&include_config=true", nil)
	w := httptest.NewRecorder()

	server.handleBackupExport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	zipData := w.Body.Bytes()
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Fatalf("Failed to read zip: %v", err)
	}

	manifestFile, err := zipReader.Open("manifest.json")
	if err != nil {
		t.Fatalf("Failed to open manifest: %v", err)
	}
	defer manifestFile.Close()

	var manifest BackupManifest
	if err := json.NewDecoder(manifestFile).Decode(&manifest); err != nil {
		t.Fatalf("Failed to decode manifest: %v", err)
	}

	// Verify counts match reality
	if manifest.WorkspaceFileCount != 5 {
		t.Errorf("Expected WorkspaceFileCount=5, got %d", manifest.WorkspaceFileCount)
	}

	if manifest.ConfigFileCount != 1 {
		t.Errorf("Expected ConfigFileCount=1, got %d", manifest.ConfigFileCount)
	}

	if manifest.TotalFiles != 6 {
		t.Errorf("Expected TotalFiles=6 (5 workspace + 1 config), got %d", manifest.TotalFiles)
	}

	// Verify ExportedFiles list is accurate
	hasConfig := false
	hasWorkspace := false
	for _, file := range manifest.ExportedFiles {
		if file == "config.json" {
			hasConfig = true
		}
		if file == "workspace/" {
			hasWorkspace = true
		}
	}

	if !hasConfig {
		t.Error("Expected 'config.json' in ExportedFiles")
	}

	if !hasWorkspace {
		t.Error("Expected 'workspace/' in ExportedFiles")
	}
}

// TestBackupExportZipIntegrity verifies the generated ZIP can be extracted
func TestBackupExportZipIntegrity(t *testing.T) {
	tempDir := t.TempDir()
	dataDir := tempDir
	workspaceDir := filepath.Join(dataDir, "workspace")

	os.MkdirAll(workspaceDir, 0755)
	ioutil.WriteFile(filepath.Join(dataDir, "config.json"), []byte(`{"test":true}`), 0644)
	ioutil.WriteFile(filepath.Join(workspaceDir, "test.json"), []byte(`{"data":"test"}`), 0644)

	server := &Server{
		workspace: workspaceDir,
	}

	req := httptest.NewRequest("GET", "/api/backup/export?include_database=false&include_workspace=true&include_config=true", nil)
	w := httptest.NewRecorder()

	server.handleBackupExport(w, req)

	// Extract ZIP to temp location and verify contents
	extractDir := t.TempDir()
	zipData := w.Body.Bytes()
	zipReader, _ := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))

	// Extract all files
	for _, file := range zipReader.File {
		filePath := filepath.Join(extractDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, 0755)
			continue
		}

		os.MkdirAll(filepath.Dir(filePath), 0755)

		src, _ := file.Open()
		defer src.Close()

		dst, _ := os.Create(filePath)
		defer dst.Close()

		io.Copy(dst, src)
	}

	// Verify extracted files exist and are readable
	if _, err := os.Stat(filepath.Join(extractDir, "manifest.json")); err != nil {
		t.Error("manifest.json not extracted properly")
	}

	if _, err := os.Stat(filepath.Join(extractDir, "config.json")); err != nil {
		t.Error("config.json not extracted properly")
	}

	if data, err := ioutil.ReadFile(filepath.Join(extractDir, "config.json")); err != nil || !bytes.Contains(data, []byte(`"test":true`)) {
		t.Error("config.json content corrupted during backup/extract")
	}
}
