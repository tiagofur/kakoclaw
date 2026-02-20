package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// KnowledgeDocument represents an uploaded document in the knowledge base.
type KnowledgeDocument struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	MimeType   string    `json:"mime_type"`
	Size       int64     `json:"size"`
	ChunkCount int       `json:"chunk_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// KnowledgeChunk represents a searchable text chunk from a document.
type KnowledgeChunk struct {
	ID         int64  `json:"id"`
	DocumentID int64  `json:"document_id"`
	Content    string `json:"content"`
	Position   int    `json:"position"` // Chunk order within the document
}

// KnowledgeSearchResult is a single result from a knowledge base search.
type KnowledgeSearchResult struct {
	DocumentID   int64   `json:"document_id"`
	DocumentName string  `json:"document_name"`
	Content      string  `json:"content"`
	Rank         float64 `json:"rank"`
}

// migrateKnowledge creates the knowledge base tables (FTS5).
// Called from the main migrate() function.
func (s *Storage) migrateKnowledge() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS knowledge_documents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			mime_type TEXT NOT NULL DEFAULT 'text/plain',
			size INTEGER NOT NULL DEFAULT 0,
			chunk_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		// Regular table to hold chunk data (needed because FTS5 virtual tables
		// cannot have extra non-text columns like document_id).
		`CREATE TABLE IF NOT EXISTS knowledge_chunks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			document_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			position INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (document_id) REFERENCES knowledge_documents(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_doc ON knowledge_chunks(document_id);`,
		// FTS5 virtual table for full-text search across chunks
		`CREATE VIRTUAL TABLE IF NOT EXISTS knowledge_fts USING fts5(
			content,
			content_rowid=id,
			tokenize='porter unicode61'
		);`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			// FTS5 table may already exist — virtual tables can't use IF NOT EXISTS
			// in all SQLite builds, so silently ignore "already exists" errors.
			if strings.Contains(err.Error(), "already exists") {
				continue
			}
			return fmt.Errorf("knowledge migration: %w", err)
		}
	}
	return nil
}

// SaveKnowledgeDocument stores a document record and its text chunks.
// Chunks are inserted into both the regular table and the FTS5 index.
func (s *Storage) SaveKnowledgeDocument(name, mimeType string, size int64, chunks []string) (*KnowledgeDocument, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO knowledge_documents (name, mime_type, size, chunk_count) VALUES (?, ?, ?, ?)`,
		name, mimeType, size, len(chunks),
	)
	if err != nil {
		return nil, fmt.Errorf("insert document: %w", err)
	}
	docID, _ := res.LastInsertId()

	for i, chunk := range chunks {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		cRes, err := tx.Exec(
			`INSERT INTO knowledge_chunks (document_id, content, position) VALUES (?, ?, ?)`,
			docID, chunk, i,
		)
		if err != nil {
			return nil, fmt.Errorf("insert chunk %d: %w", i, err)
		}
		chunkID, _ := cRes.LastInsertId()
		// Insert into FTS5 index (rowid must match knowledge_chunks.id)
		if _, err := tx.Exec(`INSERT INTO knowledge_fts (rowid, content) VALUES (?, ?)`, chunkID, chunk); err != nil {
			return nil, fmt.Errorf("insert fts chunk %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &KnowledgeDocument{
		ID:         docID,
		Name:       name,
		MimeType:   mimeType,
		Size:       size,
		ChunkCount: len(chunks),
		CreatedAt:  time.Now(),
	}, nil
}

// ListKnowledgeDocuments returns all documents in the knowledge base.
func (s *Storage) ListKnowledgeDocuments() ([]KnowledgeDocument, error) {
	rows, err := s.db.Query(
		`SELECT id, name, mime_type, size, chunk_count, created_at FROM knowledge_documents ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	var docs []KnowledgeDocument
	for rows.Next() {
		var d KnowledgeDocument
		if err := rows.Scan(&d.ID, &d.Name, &d.MimeType, &d.Size, &d.ChunkCount, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

// DeleteKnowledgeDocument removes a document and all its chunks from the knowledge base.
func (s *Storage) DeleteKnowledgeDocument(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete FTS entries for this document's chunks
	if _, err := tx.Exec(
		`DELETE FROM knowledge_fts WHERE rowid IN (SELECT id FROM knowledge_chunks WHERE document_id = ?)`, id,
	); err != nil {
		return fmt.Errorf("delete fts: %w", err)
	}

	// Delete chunks
	if _, err := tx.Exec(`DELETE FROM knowledge_chunks WHERE document_id = ?`, id); err != nil {
		return fmt.Errorf("delete chunks: %w", err)
	}

	// Delete document
	res, err := tx.Exec(`DELETE FROM knowledge_documents WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

// SearchKnowledge performs a full-text search across the knowledge base.
// Returns matching chunks ranked by relevance (BM25).
func (s *Storage) SearchKnowledge(query string, limit int) ([]KnowledgeSearchResult, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	rows, err := s.db.Query(`
		SELECT kc.document_id, kd.name, kc.content, rank
		FROM knowledge_fts
		JOIN knowledge_chunks kc ON kc.id = knowledge_fts.rowid
		JOIN knowledge_documents kd ON kd.id = kc.document_id
		WHERE knowledge_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search knowledge: %w", err)
	}
	defer rows.Close()

	var results []KnowledgeSearchResult
	for rows.Next() {
		var r KnowledgeSearchResult
		if err := rows.Scan(&r.DocumentID, &r.DocumentName, &r.Content, &r.Rank); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetKnowledgeDocumentChunks retrieves all ordered chunks for a specific document.
func (s *Storage) GetKnowledgeDocumentChunks(docID int64) ([]KnowledgeChunk, error) {
	rows, err := s.db.Query(`
		SELECT id, document_id, content, position
		FROM knowledge_chunks
		WHERE document_id = ?
		ORDER BY position ASC
	`, docID)
	if err != nil {
		return nil, fmt.Errorf("get document chunks: %w", err)
	}
	defer rows.Close()

	var chunks []KnowledgeChunk
	for rows.Next() {
		var chunk KnowledgeChunk
		if err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.Content, &chunk.Position); err != nil {
			return nil, fmt.Errorf("scan chunk: %w", err)
		}
		chunks = append(chunks, chunk)
	}
	return chunks, rows.Err()
}

// UpdateKnowledgeChunk updates the content of an existing document chunk and its search index.
func (s *Storage) UpdateKnowledgeChunk(chunkID int64, newContent string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Update the chunk text
	if _, err := tx.Exec(`UPDATE knowledge_chunks SET content = ? WHERE id = ?`, newContent, chunkID); err != nil {
		return fmt.Errorf("update chunk: %w", err)
	}

	// Update the FTS5 index text
	if _, err := tx.Exec(`UPDATE knowledge_fts SET content = ? WHERE rowid = ?`, newContent, chunkID); err != nil {
		return fmt.Errorf("update fts chunk: %w", err)
	}

	return tx.Commit()
}
