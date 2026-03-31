package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type CampaignMemoryEntry struct {
	ID         int64     `json:"id"`
	CampaignID int64     `json:"campaign_id"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

func (s *Storage) migrateCampaignMemory() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS campaign_memory (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			campaign_id INTEGER NOT NULL,
			role TEXT NOT NULL DEFAULT 'note',
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_campaign_memory_cid ON campaign_memory(campaign_id);`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("campaign_memory migration: %w", err)
		}
	}
	return nil
}

func (s *Storage) ListCampaignMemory(campaignID int64) ([]CampaignMemoryEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, campaign_id, role, content, created_at FROM campaign_memory WHERE campaign_id = ? ORDER BY created_at ASC`,
		campaignID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing campaign memory: %w", err)
	}
	defer rows.Close()

	var entries []CampaignMemoryEntry
	for rows.Next() {
		var e CampaignMemoryEntry
		if err := rows.Scan(&e.ID, &e.CampaignID, &e.Role, &e.Content, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning campaign memory entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *Storage) AddCampaignMemoryEntry(campaignID int64, role, content string) (*CampaignMemoryEntry, error) {
	if role == "" {
		role = "note"
	}
	res, err := s.db.Exec(
		`INSERT INTO campaign_memory (campaign_id, role, content) VALUES (?, ?, ?)`,
		campaignID, role, content,
	)
	if err != nil {
		return nil, fmt.Errorf("adding campaign memory entry: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting insert id: %w", err)
	}
	var e CampaignMemoryEntry
	row := s.db.QueryRow(`SELECT id, campaign_id, role, content, created_at FROM campaign_memory WHERE id = ?`, id)
	if err := row.Scan(&e.ID, &e.CampaignID, &e.Role, &e.Content, &e.CreatedAt); err != nil {
		return nil, fmt.Errorf("reading new entry: %w", err)
	}
	return &e, nil
}

func (s *Storage) DeleteCampaignMemoryEntry(id int64) error {
	res, err := s.db.Exec(`DELETE FROM campaign_memory WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting campaign memory entry: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
