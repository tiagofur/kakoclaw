package devmemory

import "context"

// Embedder defines the interface for generating vector embeddings from text.
// This allows memory injection and semantic search to work with different models
// or to be disabled entirely by providing a no-op implementation.
type Embedder interface {
	// Embed generates embeddings for a batch of texts.
	// Returns a slice of float32 vectors, one vector per input text.
	Embed(ctx context.Context, texts []string) ([][]float32, error)

	// Dimensions returns the size of the embedding vectors produced by this embedder.
	// Used for validating vectors before storing.
	Dimensions() int
}
