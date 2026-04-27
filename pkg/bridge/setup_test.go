package bridge_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/bridge"
)

func TestEnsureBridge_CreatesAndUpdatesFile(t *testing.T) {
	dir := t.TempDir()

	content1 := []byte("console.log('v1')")
	backend := "claude-code"

	// Initial creation
	path1, err := bridge.EnsureBridge(dir, content1, backend)
	if err != nil {
		t.Fatalf("First EnsureBridge error: %v", err)
	}
	expectedPath := filepath.Join(dir, backend+".mjs")
	if path1 != expectedPath {
		t.Errorf("Path mismatch: got %s, want %s", path1, expectedPath)
	}

	read1, _ := os.ReadFile(path1)
	if string(read1) != string(content1) {
		t.Errorf("Content mismatch: got %s", string(read1))
	}

	// Update with new content
	content2 := []byte("console.log('v2')")
	path2, err := bridge.EnsureBridge(dir, content2, backend)
	if err != nil {
		t.Fatalf("Second EnsureBridge error: %v", err)
	}
	if path2 != expectedPath {
		t.Errorf("Path mismatch on update: got %s", path2)
	}

	read2, _ := os.ReadFile(path2)
	if string(read2) != string(content2) {
		t.Errorf("Content update failed: got %s", string(read2))
	}

	// Unchanged test
	fiList, _ := os.Stat(path2)
	modTime := fiList.ModTime()

	_, _ = bridge.EnsureBridge(dir, content2, backend)
	fiListNew, _ := os.Stat(path2)
	if modTime != fiListNew.ModTime() {
		t.Errorf("Expected file modification time to be unchanged since content is same")
	}
}
