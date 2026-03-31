package devmemory

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"

	_ "modernc.org/sqlite"
)

// Store provides sqlite persistence for DevStudio memories.
type Store struct {
	db       *sql.DB
	embedder Embedder
}

// NewStore initializes a new Dev Studio memory store.
func NewStore(dbPath string, embedder Embedder) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	if _, err := db.Exec(Schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create devmemory schema: %w", err)
	}

	return &Store{db: db, embedder: embedder}, nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// Add stores a new memory with its embedding.
func (s *Store) Add(ctx context.Context, content string, metadata string) error {
	var blob []byte
	if s.embedder != nil {
		embeddings, err := s.embedder.Embed(ctx, []string{content})
		if err != nil {
			return fmt.Errorf("generate embedding: %w", err)
		}
		if len(embeddings) > 0 {
			blob = serializeEmbedding(embeddings[0])
		}
	}

	// SQLite datetime format
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO dev_memories (content, metadata, embedding) VALUES (?, ?, ?)",
		content, metadata, blob,
	)
	if err != nil {
		return fmt.Errorf("insert memory: %w", err)
	}

	return nil
}

// Delete removes a memory by its ID.
func (s *Store) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM dev_memories WHERE id = ?", id)
	return err
}

func serializeEmbedding(vec []float32) []byte {
	buf := make([]byte, len(vec)*4)
	for i, v := range vec {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return buf
}

func deserializeEmbedding(blob []byte) []float32 {
	vec := make([]float32, len(blob)/4)
	for i := range vec {
		vec[i] = math.Float32frombits(binary.LittleEndian.Uint32(blob[i*4:]))
	}
	return vec
}
