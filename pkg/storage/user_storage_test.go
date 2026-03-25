package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
)

func TestEnsureUserDirectoryDelegatesToEnsureUserWorkspace(t *testing.T) {
	baseDir := t.TempDir()
	config.InitDataDir(baseDir)
	defer config.InitDataDir("")

	mgr := NewUserStorageManager(nil, baseDir)
	userUUID := "delegation-user"

	if err := mgr.EnsureUserDirectory(userUUID); err != nil {
		t.Fatalf("EnsureUserDirectory: %v", err)
	}

	workspace := filepath.Join(baseDir, "users", userUUID, "workspace")
	for _, dir := range []string{"memory", "sessions", "skills", "cron", "tasks", "temp"} {
		if info, err := os.Stat(filepath.Join(workspace, dir)); err != nil {
			t.Fatalf("expected delegated directory %s to exist: %v", dir, err)
		} else if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}

	identityBytes, err := os.ReadFile(filepath.Join(workspace, "IDENTITY.md"))
	if err != nil {
		t.Fatalf("read IDENTITY.md: %v", err)
	}
	if !strings.Contains(string(identityBytes), "Apex Efficiency") {
		t.Fatalf("expected canonical workspace template content from config.EnsureUserWorkspace")
	}
}
