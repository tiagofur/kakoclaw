package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type CompanyProfile struct {
	ID             int64      `json:"id"`
	Account        string     `json:"account"`
	Website        string     `json:"website"`
	Industry       string     `json:"industry"`
	Description    string     `json:"description"`
	TargetAudience string     `json:"target_audience"`
	SocialLinks    string     `json:"social_links"` // JSON
	ResearchNotes  string     `json:"research_notes"`
	ResearchedAt   *time.Time `json:"researched_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (s *Storage) migrateCompanyProfiles() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS company_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account TEXT NOT NULL UNIQUE,
			website TEXT NOT NULL DEFAULT '',
			industry TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			target_audience TEXT NOT NULL DEFAULT '',
			social_links TEXT NOT NULL DEFAULT '{}',
			research_notes TEXT NOT NULL DEFAULT '',
			researched_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("company_profiles migration: %w", err)
		}
	}
	return nil
}

func scanCompanyProfile(row interface{ Scan(...any) error }, p *CompanyProfile) error {
	var researchedAt sql.NullString
	err := row.Scan(
		&p.ID, &p.Account, &p.Website, &p.Industry, &p.Description,
		&p.TargetAudience, &p.SocialLinks, &p.ResearchNotes,
		&researchedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if researchedAt.Valid && researchedAt.String != "" {
		t, err := time.Parse(time.RFC3339, researchedAt.String)
		if err == nil {
			p.ResearchedAt = &t
		}
	}
	return nil
}

const profileCols = `id, account, website, industry, description, target_audience, social_links, research_notes, researched_at, created_at, updated_at`

func (s *Storage) ListCompanyProfiles() ([]CompanyProfile, error) {
	rows, err := s.db.Query(`SELECT ` + profileCols + ` FROM company_profiles ORDER BY account ASC`)
	if err != nil {
		return nil, fmt.Errorf("listing company profiles: %w", err)
	}
	defer rows.Close()
	var profiles []CompanyProfile
	for rows.Next() {
		var p CompanyProfile
		if err := scanCompanyProfile(rows, &p); err != nil {
			return nil, fmt.Errorf("scanning company profile: %w", err)
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

func (s *Storage) GetCompanyProfile(account string) (*CompanyProfile, error) {
	var p CompanyProfile
	row := s.db.QueryRow(`SELECT `+profileCols+` FROM company_profiles WHERE account = ?`, account)
	if err := scanCompanyProfile(row, &p); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting company profile: %w", err)
	}
	return &p, nil
}

func (s *Storage) UpsertCompanyProfile(p *CompanyProfile) (*CompanyProfile, error) {
	now := time.Now().Format(time.RFC3339)
	var researchedAt interface{}
	if p.ResearchedAt != nil {
		researchedAt = p.ResearchedAt.Format(time.RFC3339)
	}
	if p.SocialLinks == "" {
		p.SocialLinks = "{}"
	}
	_, err := s.db.Exec(`
		INSERT INTO company_profiles (account, website, industry, description, target_audience, social_links, research_notes, researched_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(account) DO UPDATE SET
			website = excluded.website,
			industry = excluded.industry,
			description = excluded.description,
			target_audience = excluded.target_audience,
			social_links = excluded.social_links,
			research_notes = excluded.research_notes,
			researched_at = excluded.researched_at,
			updated_at = excluded.updated_at`,
		p.Account, p.Website, p.Industry, p.Description,
		p.TargetAudience, p.SocialLinks, p.ResearchNotes,
		researchedAt, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("upserting company profile: %w", err)
	}
	return s.GetCompanyProfile(p.Account)
}
