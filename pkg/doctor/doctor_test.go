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
	// This test might fail if .KakoClaw doesn't exist
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
