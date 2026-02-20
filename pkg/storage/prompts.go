package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// Prompt represents a saved prompt template
type Prompt struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Description string    `json:"description"`
	Tags        string    `json:"tags"` // comma-separated
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Storage) migratePrompts() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS prompts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		tags TEXT NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

func (s *Storage) ListPrompts() ([]Prompt, error) {
	rows, err := s.db.Query(`SELECT id, title, content, description, tags, created_at, updated_at FROM prompts ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("listing prompts: %w", err)
	}
	defer rows.Close()

	var prompts []Prompt
	for rows.Next() {
		var p Prompt
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Description, &p.Tags, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		prompts = append(prompts, p)
	}
	return prompts, nil
}

func (s *Storage) CreatePrompt(title, content, description, tags string) (*Prompt, error) {
	res, err := s.db.Exec(`INSERT INTO prompts (title, content, description, tags) VALUES (?, ?, ?, ?)`,
		title, content, description, tags)
	if err != nil {
		return nil, fmt.Errorf("creating prompt: %w", err)
	}
	id, _ := res.LastInsertId()
	p, err := s.GetPrompt(id)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Storage) GetPrompt(id int64) (*Prompt, error) {
	var p Prompt
	err := s.db.QueryRow(`SELECT id, title, content, description, tags, created_at, updated_at FROM prompts WHERE id = ?`, id).
		Scan(&p.ID, &p.Title, &p.Content, &p.Description, &p.Tags, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (s *Storage) UpdatePrompt(id int64, title, content, description, tags string) error {
	_, err := s.db.Exec(`UPDATE prompts SET title=?, content=?, description=?, tags=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		title, content, description, tags, id)
	return err
}

func (s *Storage) DeletePrompt(id int64) error {
	_, err := s.db.Exec(`DELETE FROM prompts WHERE id=?`, id)
	return err
}
