package devmemory_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sipeed/makoclaw/pkg/devmemory"
)

type MockEmbedder struct{}

func (m *MockEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	vecs := make([][]float32, len(texts))
	for i, t := range texts {
		vec := []float32{0.0, 0.0, 0.0}
		if strings.Contains(t, "apple") {
			vec[0] = 1.0
		} else if strings.Contains(t, "banana") {
			vec[1] = 1.0
		} else if strings.Contains(t, "cherry") {
			vec[2] = 1.0
		}
		vecs[i] = vec
	}
	return vecs, nil
}

func (m *MockEmbedder) Dimensions() int {
	return 3
}

func TestStore_AddAndSearch(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "dev_memory_test.db")
	store, err := devmemory.NewStore(dbPath, &MockEmbedder{})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Add memories
	err = store.Add(ctx, "I like apple and cherry", `{"type":"fruit"}`)
	if err != nil {
		t.Fatalf("Failed to add memory 1: %v", err)
	}

	err = store.Add(ctx, "banana is yellow", `{"type":"fruit"}`)
	if err != nil {
		t.Fatalf("Failed to add memory 2: %v", err)
	}

	err = store.Add(ctx, "cherry is red", `{"type":"fruit"}`)
	if err != nil {
		t.Fatalf("Failed to add memory 3: %v", err)
	}

	// Search
	results, err := store.Search(ctx, "tell me about banana", 2)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Because of mock embedder, "banana" query should match memory 2 perfectly
	if !strings.Contains(results[0].Content, "banana") {
		t.Errorf("Expected top result to contain banana, got %q", results[0].Content)
	}
}

func TestStore_Delete(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "dev_memory_test.db")
	store, err := devmemory.NewStore(dbPath, &MockEmbedder{})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	_ = store.Add(ctx, "test content", "{}")

	// Verify we can find it
	results, _ := store.Search(ctx, "test", 10)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result")
	}

	// Delete it
	err = store.Delete(ctx, results[0].ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	// Verify gone
	results2, _ := store.Search(ctx, "test", 10)
	if len(results2) != 0 {
		t.Fatalf("Expected 0 results after delete")
	}
}
