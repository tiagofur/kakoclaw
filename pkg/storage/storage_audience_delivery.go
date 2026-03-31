package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const deliverySelectCols = `id, COALESCE(campaign_account, ''), COALESCE(campaign_name, ''), contact_id, COALESCE(subject, ''), COALESCE(status, 'queued'), COALESCE(provider_message_id, ''), sent_at, delivered_at, opened_at, clicked_at, bounced_at, COALESCE(error, ''), created_at`

func scanDelivery(sc rowScanner, d *EmailDelivery) error {
	var contactID sql.NullInt64
	var sentAt, deliveredAt, openedAt, clickedAt, bouncedAt sql.NullString
	err := sc.Scan(
		&d.ID, &d.CampaignAccount, &d.CampaignName, &contactID,
		&d.Subject, &d.Status, &d.ProviderMessageID,
		&sentAt, &deliveredAt, &openedAt, &clickedAt, &bouncedAt,
		&d.Error, &d.CreatedAt,
	)
	if err != nil {
		return err
	}
	if contactID.Valid {
		cid := contactID.Int64
		d.ContactID = &cid
	}
	d.SentAt = parseNullTime(sentAt)
	d.DeliveredAt = parseNullTime(deliveredAt)
	d.OpenedAt = parseNullTime(openedAt)
	d.ClickedAt = parseNullTime(clickedAt)
	d.BouncedAt = parseNullTime(bouncedAt)
	return nil
}

func (s *Storage) CreateDelivery(d *EmailDelivery) (*EmailDelivery, error) {
	if d.Status == "" {
		d.Status = "queued"
	}
	now := time.Now().Format(time.RFC3339)
	res, err := s.db.Exec(
		`INSERT INTO email_deliveries (campaign_account, campaign_name, contact_id, subject, status, provider_message_id, error, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		d.CampaignAccount, d.CampaignName, d.ContactID, d.Subject, d.Status, d.ProviderMessageID, d.Error, now,
	)
	if err != nil {
		return nil, fmt.Errorf("creating delivery: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting insert id: %w", err)
	}
	return s.GetDeliveryByID(id)
}

func (s *Storage) GetDeliveryByID(id int64) (*EmailDelivery, error) {
	var d EmailDelivery
	row := s.db.QueryRow(`SELECT `+deliverySelectCols+` FROM email_deliveries WHERE id = ?`, id)
	if err := scanDelivery(row, &d); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting delivery: %w", err)
	}
	return &d, nil
}

func (s *Storage) ListDeliveries(account, campaign string, page, limit int) ([]EmailDelivery, int, error) {
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
		conditions = append(conditions, `campaign_account = ?`)
		args = append(args, account)
	}
	if campaign != "" {
		conditions = append(conditions, `campaign_name = ?`)
		args = append(args, campaign)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int
	countQuery := `SELECT COUNT(*) FROM email_deliveries` + whereClause
	if err := s.db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("counting deliveries: %w", err)
	}

	query := `SELECT ` + deliverySelectCols + ` FROM email_deliveries` + whereClause + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	queryArgs := append(args, limit, offset)

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []EmailDelivery
	for rows.Next() {
		var d EmailDelivery
		if err := scanDelivery(rows, &d); err != nil {
			return nil, 0, fmt.Errorf("scanning delivery: %w", err)
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, totalCount, rows.Err()
}

// validDeliveryStatuses is the whitelist of allowed delivery status values.
var validDeliveryStatuses = map[string]bool{
	"queued":    true,
	"sent":      true,
	"delivered": true,
	"opened":    true,
	"clicked":   true,
	"bounced":   true,
	"failed":    true,
	"complained": true,
}

// deliveryStatusTimestamp maps statuses to their corresponding timestamp column.
var deliveryStatusTimestamp = map[string]string{
	"sent":      "sent_at",
	"delivered": "delivered_at",
	"opened":    "opened_at",
	"clicked":   "clicked_at",
	"bounced":   "bounced_at",
}

func (s *Storage) UpdateDeliveryStatus(id int64, status string) error {
	if !validDeliveryStatuses[status] {
		return fmt.Errorf("invalid delivery status: %q", status)
	}

	now := time.Now().Format(time.RFC3339)
	timestampCol, hasTimestamp := deliveryStatusTimestamp[status]
	if !hasTimestamp {
		_, err := s.db.Exec(`UPDATE email_deliveries SET status = ? WHERE id = ?`, status, id)
		return err
	}
	_, err := s.db.Exec(
		fmt.Sprintf(`UPDATE email_deliveries SET status = ?, %s = ? WHERE id = ?`, timestampCol),
		status, now, id,
	)
	return err
}

func (s *Storage) SendToList(listID int64, account, campaignName, subject, body string, sendFunc func(to, subject, body string) error) (sent int, skipped int, err error) {
	rows, err := s.db.Query(`
		SELECT c.id, c.email, c.status
		FROM contacts c
		JOIN contact_list_members clm ON clm.contact_id = c.id
		WHERE clm.list_id = ?`, listID,
	)
	if err != nil {
		return 0, 0, fmt.Errorf("querying list members: %w", err)
	}
	defer rows.Close()

	type listMember struct {
		id     int64
		email  string
		status string
	}
	var members []listMember
	for rows.Next() {
		var m listMember
		if err := rows.Scan(&m.id, &m.email, &m.status); err != nil {
			return 0, 0, fmt.Errorf("scanning member: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return 0, 0, err
	}

	now := time.Now().Format(time.RFC3339)
	for _, m := range members {
		if m.status != "active" {
			_, dbErr := s.db.Exec(
				`INSERT INTO email_deliveries (campaign_account, campaign_name, contact_id, subject, status, error, created_at) VALUES (?, ?, ?, ?, 'failed', ?, ?)`,
				account, campaignName, m.id, subject, "unsubscribed", now,
			)
			if dbErr != nil {
				return sent, skipped, fmt.Errorf("creating failed delivery for contact %d: %w", m.id, dbErr)
			}
			skipped++
			continue
		}

		res, dbErr := s.db.Exec(
			`INSERT INTO email_deliveries (campaign_account, campaign_name, contact_id, subject, status, created_at) VALUES (?, ?, ?, ?, 'queued', ?)`,
			account, campaignName, m.id, subject, now,
		)
		if dbErr != nil {
			return sent, skipped, fmt.Errorf("creating delivery: %w", dbErr)
		}
		deliveryID, idErr := res.LastInsertId()
		if idErr != nil {
			return sent, skipped, fmt.Errorf("getting delivery id for contact %d: %w", m.id, idErr)
		}

		sendErr := sendFunc(m.email, subject, body)
		if sendErr != nil {
			if _, dbErr := s.db.Exec(
				`UPDATE email_deliveries SET status = 'failed', error = ? WHERE id = ?`,
				sendErr.Error(), deliveryID,
			); dbErr != nil {
				return sent, skipped, fmt.Errorf("updating failed delivery %d: %w", deliveryID, dbErr)
			}
			skipped++
		} else {
			if _, dbErr := s.db.Exec(
				`UPDATE email_deliveries SET status = 'sent', sent_at = ? WHERE id = ?`,
				time.Now().Format(time.RFC3339), deliveryID,
			); dbErr != nil {
				return sent, skipped, fmt.Errorf("updating sent delivery %d: %w", deliveryID, dbErr)
			}
			sent++
		}
	}
	return sent, skipped, nil
}
