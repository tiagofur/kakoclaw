package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader("/tmp/test-data")
	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}
	if loader.dataDir != "/tmp/test-data" {
		t.Errorf("dataDir = %q, want %q", loader.dataDir, "/tmp/test-data")
	}
	if loader.globalMCPPath != filepath.Join("/tmp/test-data", "mcp") {
		t.Errorf("globalMCPPath = %q, want %q", loader.globalMCPPath, filepath.Join("/tmp/test-data", "mcp"))
	}
}

func TestLoadMCPFile_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	mcpFile := filepath.Join(tmpDir, "mcp.json")

	content := `{
		"command": "npx",
		"args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"],
		"env": {"NODE_ENV": "production"}
	}`
	if err := os.WriteFile(mcpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write mcp.json: %v", err)
	}

	loader := NewLoader(tmpDir)
	cfg, err := loader.loadMCPFile(mcpFile)
	if err != nil {
		t.Fatalf("loadMCPFile: %v", err)
	}

	if cfg.Command != "npx" {
		t.Errorf("Command = %q, want %q", cfg.Command, "npx")
	}
	if len(cfg.Args) != 3 {
		t.Fatalf("Args length = %d, want 3", len(cfg.Args))
	}
	if cfg.Args[0] != "-y" {
		t.Errorf("Args[0] = %q, want %q", cfg.Args[0], "-y")
	}
	if cfg.Env["NODE_ENV"] != "production" {
		t.Errorf("Env[NODE_ENV] = %q, want %q", cfg.Env["NODE_ENV"], "production")
	}
	// Default enabled is true
	if !cfg.Enabled {
		t.Error("Expected Enabled to default to true")
	}
}

func TestLoadMCPFile_ValidWithEnabledFalse(t *testing.T) {
	tmpDir := t.TempDir()
	mcpFile := filepath.Join(tmpDir, "mcp.json")

	content := `{
		"enabled": false,
		"command": "node",
		"args": ["server.js"]
	}`
	if err := os.WriteFile(mcpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write mcp.json: %v", err)
	}

	loader := NewLoader(tmpDir)
	cfg, err := loader.loadMCPFile(mcpFile)
	if err != nil {
		t.Fatalf("loadMCPFile: %v", err)
	}

	if cfg.Enabled {
		t.Error("Expected Enabled to be false when explicitly set")
	}
}

func TestLoadMCPFile_Invalid(t *testing.T) {
	tmpDir := t.TempDir()
	mcpFile := filepath.Join(tmpDir, "mcp.json")

	if err := os.WriteFile(mcpFile, []byte(`{invalid json`), 0644); err != nil {
		t.Fatalf("failed to write mcp.json: %v", err)
	}

	loader := NewLoader(tmpDir)
	_, err := loader.loadMCPFile(mcpFile)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadMCPFile_NotFound(t *testing.T) {
	loader := NewLoader(t.TempDir())
	_, err := loader.loadMCPFile(filepath.Join(t.TempDir(), "nonexistent", "mcp.json"))
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

func TestSaveToFolder(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir)

	cfg := config.MCPServerConfig{
		Enabled: true,
		Command: "npx",
		Args:    []string{"-y", "some-server"},
		Env:     map[string]string{"KEY": "value"},
	}

	if err := loader.SaveToFolder(tmpDir, "test-server", cfg); err != nil {
		t.Fatalf("SaveToFolder: %v", err)
	}

	// Verify the file was created
	mcpFile := filepath.Join(tmpDir, "test-server", "mcp.json")
	data, err := os.ReadFile(mcpFile)
	if err != nil {
		t.Fatalf("failed to read saved mcp.json: %v", err)
	}

	// Parse and verify content
	var saved map[string]interface{}
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}

	if saved["command"] != "npx" {
		t.Errorf("saved command = %v, want %q", saved["command"], "npx")
	}

	// When enabled is true (default), it should NOT be in the JSON
	if _, hasEnabled := saved["enabled"]; hasEnabled {
		t.Error("enabled=true should not be written to JSON (it's the default)")
	}
}

func TestSaveToFolder_DisabledIncludesEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir)

	cfg := config.MCPServerConfig{
		Enabled: false,
		Command: "node",
		Args:    []string{"server.js"},
	}

	if err := loader.SaveToFolder(tmpDir, "disabled-server", cfg); err != nil {
		t.Fatalf("SaveToFolder: %v", err)
	}

	mcpFile := filepath.Join(tmpDir, "disabled-server", "mcp.json")
	data, err := os.ReadFile(mcpFile)
	if err != nil {
		t.Fatalf("failed to read saved mcp.json: %v", err)
	}

	var saved map[string]interface{}
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}

	if saved["enabled"] != false {
		t.Errorf("saved enabled = %v, want false", saved["enabled"])
	}
}

func TestDeleteFromFolder(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir)

	cfg := config.MCPServerConfig{
		Enabled: true,
		Command: "node",
		Args:    []string{"server.js"},
	}

	// Save first
	if err := loader.SaveToFolder(tmpDir, "to-delete", cfg); err != nil {
		t.Fatalf("SaveToFolder: %v", err)
	}

	// Verify it exists
	serverDir := filepath.Join(tmpDir, "to-delete")
	if _, err := os.Stat(serverDir); os.IsNotExist(err) {
		t.Fatal("server directory should exist before delete")
	}

	// Delete
	if err := loader.DeleteFromFolder(tmpDir, "to-delete"); err != nil {
		t.Fatalf("DeleteFromFolder: %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(serverDir); !os.IsNotExist(err) {
		t.Fatal("server directory should not exist after delete")
	}
}

func TestDeleteFromFolder_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir)

	// Deleting a non-existent folder should not error (os.RemoveAll behavior)
	if err := loader.DeleteFromFolder(tmpDir, "nonexistent"); err != nil {
		t.Fatalf("DeleteFromFolder on non-existent should not error: %v", err)
	}
}

func TestLoadFromFolder(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir)

	// Create multiple server configs
	servers := map[string]config.MCPServerConfig{
		"server-a": {Enabled: true, Command: "node", Args: []string{"a.js"}},
		"server-b": {Enabled: true, Command: "python", Args: []string{"b.py"}},
	}

	for name, cfg := range servers {
		if err := loader.SaveToFolder(tmpDir, name, cfg); err != nil {
			t.Fatalf("SaveToFolder(%s): %v", name, err)
		}
	}

	// Load from folder
	loaded := loader.loadFromFolder(tmpDir, SourceGlobalFolder)
	if len(loaded) != 2 {
		t.Fatalf("loadFromFolder returned %d servers, want 2", len(loaded))
	}

	// Verify both servers are found
	foundNames := map[string]bool{}
	for _, s := range loaded {
		foundNames[s.Name] = true
		if s.Source != SourceGlobalFolder {
			t.Errorf("server %s source = %q, want %q", s.Name, s.Source, SourceGlobalFolder)
		}
	}

	if !foundNames["server-a"] || !foundNames["server-b"] {
		t.Errorf("expected both server-a and server-b, found: %v", foundNames)
	}
}

func TestLoadFromFolder_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir)

	loaded := loader.loadFromFolder(tmpDir, SourceGlobalFolder)
	if len(loaded) != 0 {
		t.Errorf("loadFromFolder on empty dir returned %d servers, want 0", len(loaded))
	}
}

func TestLoadFromFolder_NonExistentDir(t *testing.T) {
	loader := NewLoader(t.TempDir())
	loaded := loader.loadFromFolder(filepath.Join(t.TempDir(), "nonexistent"), SourceGlobalFolder)
	if len(loaded) != 0 {
		t.Errorf("loadFromFolder on non-existent dir returned %d servers, want 0", len(loaded))
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir)

	original := config.MCPServerConfig{
		Enabled: true,
		Command: "npx",
		Args:    []string{"-y", "@mcp/server", "--port", "3000"},
		Env:     map[string]string{"API_KEY": "secret123", "DEBUG": "true"},
	}

	if err := loader.SaveToFolder(tmpDir, "roundtrip", original); err != nil {
		t.Fatalf("SaveToFolder: %v", err)
	}

	// Load back
	mcpFile := filepath.Join(tmpDir, "roundtrip", "mcp.json")
	loaded, err := loader.loadMCPFile(mcpFile)
	if err != nil {
		t.Fatalf("loadMCPFile: %v", err)
	}

	if loaded.Command != original.Command {
		t.Errorf("Command = %q, want %q", loaded.Command, original.Command)
	}
	if len(loaded.Args) != len(original.Args) {
		t.Fatalf("Args length = %d, want %d", len(loaded.Args), len(original.Args))
	}
	for i, arg := range loaded.Args {
		if arg != original.Args[i] {
			t.Errorf("Args[%d] = %q, want %q", i, arg, original.Args[i])
		}
	}
	if loaded.Env["API_KEY"] != "secret123" {
		t.Errorf("Env[API_KEY] = %q, want %q", loaded.Env["API_KEY"], "secret123")
	}
	if !loaded.Enabled {
		t.Error("Expected Enabled to be true after round-trip")
	}
}

func TestEnsureMCPFolders(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir)
	loader.SetUserMCPPath(filepath.Join(tmpDir, "users", "test-user", "mcp"))

	if err := loader.EnsureMCPFolders(); err != nil {
		t.Fatalf("EnsureMCPFolders: %v", err)
	}

	// Check global folder exists
	if _, err := os.Stat(loader.GetGlobalMCPPath()); os.IsNotExist(err) {
		t.Error("global MCP folder should exist after EnsureMCPFolders")
	}

	// Check user folder exists
	if _, err := os.Stat(loader.GetUserMCPPath()); os.IsNotExist(err) {
		t.Error("user MCP folder should exist after EnsureMCPFolders")
	}
}
