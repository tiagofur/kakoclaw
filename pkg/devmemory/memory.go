package devmemory

import "time"

// Memory represents a stored semantic memory entry.
type Memory struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Metadata  string    `json:"metadata"` // JSON string representation
	Embedding []float32 `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
