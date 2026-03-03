package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusOK, "OK"},
		{StatusWarning, "WARNING"},
		{StatusError, "ERROR"},
		{Status(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("Status.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusIcon(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusOK, "✓"},
		{StatusWarning, "⚠"},
		{StatusError, "✗"},
		{Status(999), "?"},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			if got := tt.status.Icon(); got != tt.want {
				t.Errorf("Status.Icon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name    string
		results []CheckResult
		want    bool
	}{
		{
			name: "no errors",
			results: []CheckResult{
				{Status: StatusOK},
				{Status: StatusWarning},
			},
			want: false,
		},
		{
			name: "has errors",
			results: []CheckResult{
				{Status: StatusOK},
				{Status: StatusError},
			},
			want: true,
		},
		{
			name:    "empty results",
			results: []CheckResult{},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasErrors(tt.results); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckDirectories(t *testing.T) {
	// This test might fail if .makoclaw doesn't exist
	// But it should still run without panic
	result := checkDirectories()

	// Just verify it returns a result
	if result.Name != "Directories" {
		t.Errorf("Expected check name 'Directories', got %s", result.Name)
	}
}

func TestCheckConfigFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Test with non-existent file
	result := checkConfigFile(configPath)
	if result.Status != StatusError {
		t.Errorf("Expected ERROR for non-existent config, got %s", result.Status)
	}

	// Test with valid config
	validConfig := `{
		"agents": {
			"defaults": {
				"model": "gpt-4"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	result = checkConfigFile(configPath)
	if result.Status != StatusOK {
		t.Errorf("Expected OK for valid config, got %s: %s", result.Status, result.Message)
	}

	// Test with invalid JSON
	invalidConfig := `{invalid json`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	result = checkConfigFile(configPath)
	if result.Status != StatusError {
		t.Errorf("Expected ERROR for invalid JSON, got %s", result.Status)
	}
}

func TestCheckPermissions(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create file with world-readable permissions (insecure)
	if err := os.WriteFile(configPath, []byte(`{}`), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result := checkPermissions(configPath)
	// Should warn about permissions on most systems
	if result.Name != "Permissions" {
		t.Errorf("Expected check name 'Permissions', got %s", result.Name)
	}
}

func TestCheckConfigFile_ValidWithProviders(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	validConfig := `{
		"agents": {
			"defaults": {
				"model": "claude-3-opus"
			}
		},
		"providers": {
			"anthropic": {
				"api_key": "sk-ant-1234567890abcdefg"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	result := checkConfigFile(configPath)
	if result.Status != StatusOK {
		t.Errorf("Expected OK for valid config with providers, got %s: %s", result.Status, result.Message)
	}
	if result.Name != "Configuration File" {
		t.Errorf("Expected check name 'Configuration File', got %s", result.Name)
	}
}

func TestCheckConfigFile_MissingModel(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Config with no model set
	noModelConfig := `{
		"agents": {
			"defaults": {}
		}
	}`
	if err := os.WriteFile(configPath, []byte(noModelConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	result := checkConfigFile(configPath)
	if result.Status != StatusWarning {
		t.Errorf("Expected WARNING for config with no model, got %s: %s", result.Status, result.Message)
	}
}

func TestCheckWorkspace_ValidWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		t.Fatalf("Failed to create workspace dir: %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.json")
	validConfig := `{
		"agents": {
			"defaults": {
				"model": "gpt-4",
				"workspace": "` + filepath.ToSlash(workspaceDir) + `"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	result := checkWorkspace(configPath)
	if result.Name != "Workspace" {
		t.Errorf("Expected check name 'Workspace', got %s", result.Name)
	}
	// Result depends on whether workspace path is resolved - just verify no panic
	// and the result is a valid CheckResult
	if result.Status != StatusOK && result.Status != StatusWarning && result.Status != StatusError {
		t.Errorf("Unexpected status: %v", result.Status)
	}
}

func TestCheckWorkspace_NonExistentWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Config pointing to a non-existent workspace directory
	nonExistentWorkspace := filepath.Join(tmpDir, "does", "not", "exist", "workspace")
	validConfig := `{
		"agents": {
			"defaults": {
				"model": "gpt-4",
				"workspace": "` + filepath.ToSlash(nonExistentWorkspace) + `"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	result := checkWorkspace(configPath)
	if result.Name != "Workspace" {
		t.Errorf("Expected check name 'Workspace', got %s", result.Name)
	}
	if result.Status != StatusError {
		t.Errorf("Expected ERROR for non-existent workspace directory, got %s: %s", result.Status, result.Message)
	}
}

func TestRunChecks_WithValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	validConfig := `{
		"agents": {
			"defaults": {
				"model": "gpt-4"
			}
		},
		"providers": {
			"openai": {
				"api_key": "sk-test1234567890abcdefg"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	results := RunChecks(configPath)
	if len(results) == 0 {
		t.Fatal("RunChecks returned empty results")
	}

	// Verify that we got results for expected check names
	checkNames := make(map[string]bool)
	for _, r := range results {
		checkNames[r.Name] = true
	}

	expectedChecks := []string{"Configuration File", "Workspace", "API Keys", "Providers", "Permissions", "Directories"}
	for _, name := range expectedChecks {
		if !checkNames[name] {
			t.Errorf("Missing check result for %q", name)
		}
	}
}

func TestRunChecks_MissingConfig(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "nonexistent", "config.json")

	results := RunChecks(nonExistentPath)
	if len(results) == 0 {
		t.Fatal("RunChecks returned empty results even with missing config")
	}

	// The config file check should be an error
	foundConfigError := false
	for _, r := range results {
		if r.Name == "Configuration File" && r.Status == StatusError {
			foundConfigError = true
			break
		}
	}
	if !foundConfigError {
		t.Error("Expected an ERROR status for 'Configuration File' check with missing config")
	}
}

func TestPrintResults_NoError(t *testing.T) {
	results := []CheckResult{
		{Name: "Test Check 1", Status: StatusOK, Message: "Everything is fine"},
		{Name: "Test Check 2", Status: StatusOK, Message: "All good"},
	}

	// Should not panic
	PrintResults(results)
}

func TestPrintResults_WithErrors(t *testing.T) {
	results := []CheckResult{
		{Name: "Test Check 1", Status: StatusOK, Message: "This is fine"},
		{Name: "Test Check 2", Status: StatusError, Message: "Something broke", Fix: "Fix it"},
		{Name: "Test Check 3", Status: StatusWarning, Message: "Be careful"},
	}

	// Should not panic
	PrintResults(results)
}

func TestPrintResults_Empty(t *testing.T) {
	// Should not panic with empty results
	PrintResults([]CheckResult{})
}
