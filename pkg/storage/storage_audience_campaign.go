package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const campaignSelectCols = `id, COALESCE(account, ''), name, COALESCE(subject, ''), COALESCE(body_html, ''), COALESCE(body_text, ''), COALESCE(template_slug, ''), list_id, segment_id, COALESCE(status, 'draft'), scheduled_at, sent_count, skipped_count, COALESCE(delivery_progress, ''), created_at, updated_at`

func scanCampaign(sc rowScanner, c *EmailCampaign) error {
	var scheduledAt sql.NullString
	err := sc.Scan(
		&c.ID, &c.Account, &c.Name, &c.Subject, &c.BodyHTML, &c.BodyText,
		&c.TemplateSlug, &c.ListID, &c.SegmentID, &c.Status,
		&scheduledAt, &c.SentCount, &c.SkippedCount, &c.DeliveryProgress,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return err
	}
	c.ScheduledAt = parseNullTime(scheduledAt)
	return nil
}

// UpdateCampaignDeliveryProgress writes live delivery progress JSON to the campaign row.
func (s *Storage) UpdateCampaignDeliveryProgress(id int64, progressJSON string) error {
	_, err := s.db.Exec(`UPDATE email_campaigns SET delivery_progress = ?, updated_at = ? WHERE id = ?`,
		progressJSON, time.Now().Format(time.RFC3339), id)
	return err
}

func (s *Storage) ListCampaigns(account, status string, page, limit int) ([]EmailCampaign, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var conditions []string
	var args []interface{}

	if account != "" {
		conditions = append(conditions, `account = ?`)
		args = append(args, account)
	}
	if status != "" {
		conditions = append(conditions, `status = ?`)
		args = append(args, status)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int
	countQuery := `SELECT COUNT(*) FROM email_campaigns` + whereClause
	if err := s.db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("counting campaigns: %w", err)
	}

	query := `SELECT ` + campaignSelectCols + ` FROM email_campaigns` + whereClause + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	queryArgs := append(args, limit, offset)

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []EmailCampaign
	for rows.Next() {
		var c EmailCampaign
		if err := scanCampaign(rows, &c); err != nil {
			return nil, 0, fmt.Errorf("scanning campaign: %w", err)
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, totalCount, rows.Err()
}

func (s *Storage) GetCampaignByID(id int64) (*EmailCampaign, error) {
	var c EmailCampaign
	row := s.db.QueryRow(`SELECT `+campaignSelectCols+` FROM email_campaigns WHERE id = ?`, id)
	if err := scanCampaign(row, &c); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting campaign: %w", err)
	}
	return &c, nil
}

func (s *Storage) CreateCampaign(c *EmailCampaign) (*EmailCampaign, error) {
	if c.Status == "" {
		c.Status = "draft"
	}
	now := time.Now().Format(time.RFC3339)
	res, err := s.db.Exec(
		`INSERT INTO email_campaigns (account, name, subject, body_html, body_text, template_slug, list_id, segment_id, status, scheduled_at, sent_count, skipped_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Account, c.Name, c.Subject, c.BodyHTML, c.BodyText,
		c.TemplateSlug, c.ListID, c.SegmentID, c.Status,
		nilIfZero(c.ScheduledAt), c.SentCount, c.SkippedCount, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("creating campaign: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting insert id: %w", err)
	}
	return s.GetCampaignByID(id)
}

func (s *Storage) UpdateCampaign(c *EmailCampaign) error {
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.Exec(
		`UPDATE email_campaigns SET account = ?, name = ?, subject = ?, body_html = ?, body_text = ?, template_slug = ?, list_id = ?, segment_id = ?, status = ?, scheduled_at = ?, sent_count = ?, skipped_count = ?, updated_at = ? WHERE id = ?`,
		c.Account, c.Name, c.Subject, c.BodyHTML, c.BodyText,
		c.TemplateSlug, c.ListID, c.SegmentID, c.Status,
		nilIfZero(c.ScheduledAt), c.SentCount, c.SkippedCount, now, c.ID,
	)
	return err
}

func (s *Storage) DeleteCampaign(id int64) error {
	_, err := s.db.Exec(`DELETE FROM email_campaigns WHERE id = ?`, id)
	return err
}

func (s *Storage) GetListMembers(listID int64) ([]Contact, error) {
	rows, err := s.db.Query(`
		SELECT c.id, COALESCE(c.email, ''), COALESCE(c.first_name, ''), COALESCE(c.last_name, ''), COALESCE(c.phone, ''), COALESCE(c.company, ''), COALESCE(c.title, ''), COALESCE(c.tags, ''), COALESCE(c.custom_fields, '{}'), COALESCE(c.status, ''), COALESCE(c.source, ''), c.subscribed_at, c.unsubscribed_at, c.created_at, c.updated_at
		FROM contacts c
		JOIN contact_list_members clm ON clm.contact_id = c.id
		WHERE clm.list_id = ?`, listID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying list members: %w", err)
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var c Contact
		if err := scanContact(rows, &c); err != nil {
			return nil, fmt.Errorf("scanning contact: %w", err)
		}
		contacts = append(contacts, c)
	}
	return contacts, rows.Err()
}

func nilIfZero(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339)
}
