package storage

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"strings"
	"time"
)

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func parseNullTime(ns sql.NullString) *time.Time {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(layout, ns.String); err == nil {
			return &t
		}
	}
	return nil
}

const contactSelectCols = `id, email, COALESCE(first_name, ''), COALESCE(last_name, ''), COALESCE(phone, ''), COALESCE(company, ''), COALESCE(title, ''), COALESCE(tags, '[]'), COALESCE(custom_fields, '{}'), COALESCE(status, 'active'), COALESCE(source, 'manual'), subscribed_at, unsubscribed_at, created_at, updated_at`

func scanContact(sc rowScanner, c *Contact) error {
	var subAt, unsubAt sql.NullString
	err := sc.Scan(
		&c.ID, &c.Email, &c.FirstName, &c.LastName, &c.Phone, &c.Company,
		&c.Title, &c.Tags, &c.CustomFields, &c.Status, &c.Source,
		&subAt, &unsubAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return err
	}
	c.SubscribedAt = parseNullTime(subAt)
	c.UnsubscribedAt = parseNullTime(unsubAt)
	return nil
}

// validateEmail performs basic email format validation.
func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}
	atIdx := strings.Index(email, "@")
	if atIdx < 1 {
		return fmt.Errorf("invalid email format: missing @")
	}
	domain := email[atIdx+1:]
	if !strings.Contains(domain, ".") || strings.HasSuffix(domain, ".") {
		return fmt.Errorf("invalid email format: invalid domain")
	}
	return nil
}

func (s *Storage) CreateContact(c *Contact) (*Contact, error) {
	if err := validateEmail(c.Email); err != nil {
		return nil, err
	}
	if c.Tags == "" {
		c.Tags = "[]"
	}
	if c.CustomFields == "" {
		c.CustomFields = "{}"
	}
	if c.Status == "" {
		c.Status = "active"
	}
	if c.Source == "" {
		c.Source = "manual"
	}
	now := time.Now().Format(time.RFC3339)
	var subAt, unsubAt interface{}
	if c.SubscribedAt != nil {
		subAt = c.SubscribedAt.Format(time.RFC3339)
	}
	if c.UnsubscribedAt != nil {
		unsubAt = c.UnsubscribedAt.Format(time.RFC3339)
	}
	res, err := s.db.Exec(
		`INSERT INTO contacts (email, first_name, last_name, phone, company, title, tags, custom_fields, status, source, subscribed_at, unsubscribed_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Email, c.FirstName, c.LastName, c.Phone, c.Company, c.Title,
		c.Tags, c.CustomFields, c.Status, c.Source, subAt, unsubAt, now, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("contact with email %s already exists", c.Email)
		}
		return nil, fmt.Errorf("creating contact: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting insert id: %w", err)
	}
	return s.GetContactByID(id)
}

func (s *Storage) GetContactByID(id int64) (*Contact, error) {
	var c Contact
	row := s.db.QueryRow(`SELECT `+contactSelectCols+` FROM contacts WHERE id = ?`, id)
	if err := scanContact(row, &c); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting contact: %w", err)
	}
	return &c, nil
}

func (s *Storage) UpdateContact(c *Contact) error {
	var subAt, unsubAt interface{}
	if c.SubscribedAt != nil {
		subAt = c.SubscribedAt.Format(time.RFC3339)
	}
	if c.UnsubscribedAt != nil {
		unsubAt = c.UnsubscribedAt.Format(time.RFC3339)
	}
	_, err := s.db.Exec(
		`UPDATE contacts SET email=?, first_name=?, last_name=?, phone=?, company=?, title=?, tags=?, custom_fields=?, status=?, source=?, subscribed_at=?, unsubscribed_at=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		c.Email, c.FirstName, c.LastName, c.Phone, c.Company, c.Title,
		c.Tags, c.CustomFields, c.Status, c.Source, subAt, unsubAt, c.ID,
	)
	return err
}

func (s *Storage) DeleteContact(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM contact_list_members WHERE contact_id = ?`, id); err != nil {
		return fmt.Errorf("deleting contact list memberships: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM contacts WHERE id = ?`, id); err != nil {
		return fmt.Errorf("deleting contact: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing contact deletion: %w", err)
	}

	// Recalculate segment contact counts (best-effort, non-blocking)
	go func() {
		defer func() { recover() }()
		s.refreshAllSegmentCounts()
	}()
	return nil
}

func (s *Storage) ListContacts(filter ContactFilter) ([]Contact, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	var conditions []string
	var args []interface{}

	if filter.Q != "" {
		searchTerm := "%" + escapeLikeQuery(filter.Q) + "%"
		conditions = append(conditions, `(email LIKE ? ESCAPE '\' OR first_name LIKE ? ESCAPE '\' OR last_name LIKE ? ESCAPE '\')`)
		args = append(args, searchTerm, searchTerm, searchTerm)
	}
	if filter.Status != "" {
		conditions = append(conditions, `status = ?`)
		args = append(args, filter.Status)
	}
	if filter.Tags != "" {
		conditions = append(conditions, `tags LIKE ?`)
		args = append(args, "%"+escapeLikeQuery(filter.Tags)+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int
	countQuery := `SELECT COUNT(*) FROM contacts` + whereClause
	if err := s.db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("counting contacts: %w", err)
	}

	query := `SELECT ` + contactSelectCols + ` FROM contacts` + whereClause + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	queryArgs := append(args, filter.Limit, offset)

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing contacts: %w", err)
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var c Contact
		if err := scanContact(rows, &c); err != nil {
			return nil, 0, fmt.Errorf("scanning contact: %w", err)
		}
		contacts = append(contacts, c)
	}
	return contacts, totalCount, rows.Err()
}

func (s *Storage) ImportContactsFromCSV(records [][]string, headers []string) (inserted int, skipped int, err error) {
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	emailIdx, ok := headerMap["email"]
	if !ok {
		return 0, 0, fmt.Errorf("missing 'email' column in CSV headers")
	}

	firstNameIdx := headerMap["first_name"]
	lastNameIdx := headerMap["last_name"]
	phoneIdx := headerMap["phone"]
	companyIdx := headerMap["company"]
	titleIdx := headerMap["title"]

	getField := func(row []string, idx int) string {
		if idx >= 0 && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	now := time.Now().Format(time.RFC3339)
	for _, row := range records {
		email := getField(row, emailIdx)
		if email == "" || validateEmail(email) != nil {
			skipped++
			continue
		}

		_, dbErr := s.db.Exec(
			`INSERT INTO contacts (email, first_name, last_name, phone, company, title, tags, custom_fields, status, source, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, '[]', '{}', 'active', 'csv_import', ?, ?)`,
			email, getField(row, firstNameIdx), getField(row, lastNameIdx),
			getField(row, phoneIdx), getField(row, companyIdx), getField(row, titleIdx),
			now, now,
		)
		if dbErr != nil {
			if strings.Contains(dbErr.Error(), "UNIQUE constraint") {
				skipped++
				continue
			}
			return inserted, skipped, fmt.Errorf("importing contact %s: %w", email, dbErr)
		}
		inserted++
	}
	return inserted, skipped, nil
}

func (s *Storage) ExportContactsCSV(filter ContactFilter) ([]byte, error) {
	exportFilter := filter
	exportFilter.Page = 1
	exportFilter.Limit = 1000000
	contacts, _, err := s.ListContacts(exportFilter)
	if err != nil {
		return nil, fmt.Errorf("querying contacts for export: %w", err)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"email", "first_name", "last_name", "phone", "company", "title", "status", "tags"}); err != nil {
		return nil, fmt.Errorf("writing CSV header: %w", err)
	}
	for _, c := range contacts {
		if err := w.Write([]string{c.Email, c.FirstName, c.LastName, c.Phone, c.Company, c.Title, c.Status, c.Tags}); err != nil {
			return nil, fmt.Errorf("writing CSV row: %w", err)
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}
