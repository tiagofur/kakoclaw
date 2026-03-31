package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type CampaignVersion struct {
	ID            int64     `json:"id"`
	CampaignID    int64     `json:"campaign_id"`
	VersionNumber int       `json:"version_number"`
	Subject       string    `json:"subject"`
	BodyHTML      string    `json:"body_html"`
	BodyText      string    `json:"body_text"`
	Note          string    `json:"note"`
	CreatedAt     time.Time `json:"created_at"`
}

func (s *Storage) migrateCampaignVersions() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS campaign_versions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			campaign_id INTEGER NOT NULL,
			version_number INTEGER NOT NULL DEFAULT 1,
			subject TEXT NOT NULL DEFAULT '',
			body_html TEXT NOT NULL DEFAULT '',
			body_text TEXT NOT NULL DEFAULT '',
			note TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_campaign_versions_cid ON campaign_versions(campaign_id, version_number DESC);`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("campaign_versions migration: %w", err)
		}
	}
	return nil
}

func (s *Storage) ListCampaignVersions(campaignID int64) ([]CampaignVersion, error) {
	rows, err := s.db.Query(
		`SELECT id, campaign_id, version_number, subject, body_html, body_text, note, created_at
		 FROM campaign_versions WHERE campaign_id = ? ORDER BY version_number DESC`,
		campaignID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing campaign versions: %w", err)
	}
	defer rows.Close()
	var versions []CampaignVersion
	for rows.Next() {
		var v CampaignVersion
		if err := rows.Scan(&v.ID, &v.CampaignID, &v.VersionNumber, &v.Subject, &v.BodyHTML, &v.BodyText, &v.Note, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning campaign version: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (s *Storage) SaveCampaignVersion(campaignID int64, subject, bodyHTML, bodyText, note string) (*CampaignVersion, error) {
	var nextNum int
	_ = s.db.QueryRow(`SELECT COALESCE(MAX(version_number), 0) + 1 FROM campaign_versions WHERE campaign_id = ?`, campaignID).Scan(&nextNum)
	if nextNum == 0 {
		nextNum = 1
	}

	res, err := s.db.Exec(
		`INSERT INTO campaign_versions (campaign_id, version_number, subject, body_html, body_text, note) VALUES (?, ?, ?, ?, ?, ?)`,
		campaignID, nextNum, subject, bodyHTML, bodyText, note,
	)
	if err != nil {
		return nil, fmt.Errorf("saving campaign version: %w", err)
	}
	id, _ := res.LastInsertId()

	var v CampaignVersion
	row := s.db.QueryRow(`SELECT id, campaign_id, version_number, subject, body_html, body_text, note, created_at FROM campaign_versions WHERE id = ?`, id)
	if err := row.Scan(&v.ID, &v.CampaignID, &v.VersionNumber, &v.Subject, &v.BodyHTML, &v.BodyText, &v.Note, &v.CreatedAt); err != nil {
		return nil, fmt.Errorf("reading saved version: %w", err)
	}
	return &v, nil
}

func (s *Storage) GetCampaignVersion(versionID int64) (*CampaignVersion, error) {
	var v CampaignVersion
	row := s.db.QueryRow(
		`SELECT id, campaign_id, version_number, subject, body_html, body_text, note, created_at FROM campaign_versions WHERE id = ?`,
		versionID,
	)
	if err := row.Scan(&v.ID, &v.CampaignID, &v.VersionNumber, &v.Subject, &v.BodyHTML, &v.BodyText, &v.Note, &v.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting campaign version: %w", err)
	}
	return &v, nil
}
