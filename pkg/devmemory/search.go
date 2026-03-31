package devmemory

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// Search finds the most similar memories to the query using cosine similarity.
func (s *Store) Search(ctx context.Context, query string, limit int) ([]Memory, error) {
	if s.embedder == nil {
		return nil, fmt.Errorf("search not available: embedder is nil")
	}

	embeddings, err := s.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("generate query embedding: %w", err)
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("query embedding returned 0 vectors")
	}
	queryEmbedding := embeddings[0]

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, content, metadata, embedding, created_at FROM dev_memories WHERE embedding IS NOT NULL`,
	)
	if err != nil {
		return nil, fmt.Errorf("query memories: %w", err)
	}
	defer rows.Close()

	type scored struct {
		memory Memory
		score  float64
	}

	var results []scored
	for rows.Next() {
		var m Memory
		var createdAt string
		var blob []byte

		if err := rows.Scan(&m.ID, &m.Content, &m.Metadata, &blob, &createdAt); err != nil {
			return nil, fmt.Errorf("scan memory: %w", err)
		}

		var t time.Time
		if tp, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
			t = tp
		} else if tp, err := time.Parse(time.RFC3339, createdAt); err == nil {
			t = tp
		}
		m.CreatedAt = t

		embedding := deserializeEmbedding(blob)
		score := cosineSimilarity(queryEmbedding, embedding)
		results = append(results, scored{memory: m, score: score})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	memories := make([]Memory, len(results))
	for i, res := range results {
		memories[i] = res.memory
	}

	return memories, nil
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}

	return dot / denom
}
