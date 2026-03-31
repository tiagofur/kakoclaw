package devmemory_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sipeed/makoclaw/pkg/devmemory"
)

func TestStore_Inject(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "dev_memory_inject_test.db")
	store, err := devmemory.NewStore(dbPath, &MockEmbedder{})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = store.Add(ctx, "project follows clean architecture", "{}")
	_ = store.Add(ctx, "banana split config used in dev", "{}")

	block, err := store.Inject(ctx, "banana", 1)
	if err != nil {
		t.Fatalf("Inject failed: %v", err)
	}

	if !strings.Contains(block, "## 🧠 Relevant Developer Workspace Memories") {
		t.Errorf("Missing header in injected block: \n%s", block)
	}
	if !strings.Contains(block, "banana split config used in dev") {
		t.Errorf("Missing content in injected block: \n%s", block)
	}
	if strings.Contains(block, "clean architecture") {
		t.Errorf("Injected block contained irrelevant result: \n%s", block)
	}
}

func TestStore_InjectEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "dev_memory_inject_empty_test.db")
	store, _ := devmemory.NewStore(dbPath, &MockEmbedder{})
	defer store.Close()

	ctx := context.Background()

	block, err := store.Inject(ctx, "banana", 5)
	if err != nil {
		t.Fatalf("Inject empty failed: %v", err)
	}

	if block != "" {
		t.Errorf("Expected empty string when no memories found, got: %q", block)
	}
}
