package storage

import (
	"database/sql"
	"fmt"
)

const variantSelectCols = `id, campaign_id, COALESCE(name, ''), COALESCE(subject, ''), COALESCE(body_html, ''), COALESCE(body_text, ''), split_percent, is_winner, created_at`

func scanVariant(sc rowScanner, v *ExperimentVariant) error {
	var isWinner int
	err := sc.Scan(
		&v.ID, &v.CampaignID, &v.Name, &v.Subject, &v.BodyHTML,
		&v.BodyText, &v.SplitPercent, &isWinner, &v.CreatedAt,
	)
	if err != nil {
		return err
	}
	v.IsWinner = isWinner == 1
	return nil
}

func (s *Storage) CreateVariant(v *ExperimentVariant) (*ExperimentVariant, error) {
	res, err := s.db.Exec(
		`INSERT INTO experiment_variants (campaign_id, name, subject, body_html, body_text, split_percent) VALUES (?, ?, ?, ?, ?, ?)`,
		v.CampaignID, v.Name, v.Subject, v.BodyHTML, v.BodyText, v.SplitPercent,
	)
	if err != nil {
		return nil, fmt.Errorf("creating variant: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting variant insert id: %w", err)
	}
	return s.GetVariantByID(id)
}

func (s *Storage) ListVariants(campaignID int64) ([]ExperimentVariant, error) {
	rows, err := s.db.Query(`SELECT `+variantSelectCols+` FROM experiment_variants WHERE campaign_id = ? ORDER BY id`, campaignID)
	if err != nil {
		return nil, fmt.Errorf("listing variants: %w", err)
	}
	defer rows.Close()

	var variants []ExperimentVariant
	for rows.Next() {
		var v ExperimentVariant
		if err := scanVariant(rows, &v); err != nil {
			return nil, fmt.Errorf("scanning variant: %w", err)
		}
		variants = append(variants, v)
	}
	if variants == nil {
		variants = []ExperimentVariant{}
	}
	return variants, rows.Err()
}

func (s *Storage) GetVariantByID(id int64) (*ExperimentVariant, error) {
	var v ExperimentVariant
	row := s.db.QueryRow(`SELECT `+variantSelectCols+` FROM experiment_variants WHERE id = ?`, id)
	if err := scanVariant(row, &v); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting variant: %w", err)
	}
	return &v, nil
}

func (s *Storage) UpdateVariant(v *ExperimentVariant) error {
	_, err := s.db.Exec(
		`UPDATE experiment_variants SET name = ?, subject = ?, body_html = ?, body_text = ?, split_percent = ? WHERE id = ?`,
		v.Name, v.Subject, v.BodyHTML, v.BodyText, v.SplitPercent, v.ID,
	)
	return err
}

func (s *Storage) DeleteVariant(id int64) error {
	_, err := s.db.Exec(`DELETE FROM experiment_variants WHERE id = ?`, id)
	return err
}

func (s *Storage) SetWinner(id int64) error {
	var campaignID int64
	if err := s.db.QueryRow(`SELECT campaign_id FROM experiment_variants WHERE id = ?`, id).Scan(&campaignID); err != nil {
		return fmt.Errorf("finding variant campaign: %w", err)
	}
	if _, err := s.db.Exec(`UPDATE experiment_variants SET is_winner = 0 WHERE campaign_id = ?`, campaignID); err != nil {
		return fmt.Errorf("resetting winners: %w", err)
	}
	_, err := s.db.Exec(`UPDATE experiment_variants SET is_winner = 1 WHERE id = ?`, id)
	return err
}

func (s *Storage) GetVariantMetrics(variantID int64) (sent, opened, clicked int, err error) {
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM email_deliveries WHERE variant_id = ?`, variantID).Scan(&sent); err != nil {
		return 0, 0, 0, fmt.Errorf("counting sent: %w", err)
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM email_deliveries WHERE variant_id = ? AND opened_at IS NOT NULL`, variantID).Scan(&opened); err != nil {
		return 0, 0, 0, fmt.Errorf("counting opened: %w", err)
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM email_deliveries WHERE variant_id = ? AND clicked_at IS NOT NULL`, variantID).Scan(&clicked); err != nil {
		return 0, 0, 0, fmt.Errorf("counting clicked: %w", err)
	}
	return sent, opened, clicked, nil
}
