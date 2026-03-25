package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureUserWorkspaceCreatesExpectedStructure(t *testing.T) {
	baseDir := t.TempDir()
	InitDataDir(baseDir)
	defer InitDataDir("")

	workspace, err := EnsureUserWorkspace("user-workspace-test")
	if err != nil {
		t.Fatalf("EnsureUserWorkspace: %v", err)
	}

	expectedPaths := []string{
		workspace,
		filepath.Join(workspace, "memory"),
		filepath.Join(workspace, "sessions"),
		filepath.Join(workspace, "skills"),
		filepath.Join(workspace, "cron"),
		filepath.Join(workspace, "tasks"),
		filepath.Join(workspace, "temp"),
	}

	for _, expectedPath := range expectedPaths {
		info, err := os.Stat(expectedPath)
		if err != nil {
			t.Fatalf("expected path %s to exist: %v", expectedPath, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected path %s to be a directory", expectedPath)
		}
	}
}
