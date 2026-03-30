package storage

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var slugNonAlphaNum = regexp.MustCompile(`[^a-z0-9\s-]`)
var slugSpaces = regexp.MustCompile(`[\s]+`)
var slugMultiHyphen = regexp.MustCompile(`-+`)

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = slugNonAlphaNum.ReplaceAllString(slug, "")
	slug = slugSpaces.ReplaceAllString(slug, "-")
	slug = slugMultiHyphen.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

const listSelectCols = `id, name, COALESCE(slug, ''), COALESCE(description, ''), COALESCE(type, 'static'), COALESCE(account, ''), created_at, updated_at`

func (s *Storage) CreateContactList(l *ContactList) (*ContactList, error) {
	slug := generateSlug(l.Name)
	if slug == "" {
		slug = fmt.Sprintf("list-%d", time.Now().Unix())
	}
	now := time.Now().Format(time.RFC3339)
	res, err := s.db.Exec(
		`INSERT INTO contact_lists (name, slug, description, type, account, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		l.Name, slug, l.Description, l.Type, l.Account, now, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("contact list with slug %s already exists", slug)
		}
		return nil, fmt.Errorf("creating contact list: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting insert id: %w", err)
	}
	return s.GetContactListByID(id)
}

func (s *Storage) GetContactListByID(id int64) (*ContactList, error) {
	var l ContactList
	err := s.db.QueryRow(`SELECT `+listSelectCols+` FROM contact_lists WHERE id = ?`, id).
		Scan(&l.ID, &l.Name, &l.Slug, &l.Description, &l.Type, &l.Account, &l.CreatedAt, &l.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting contact list: %w", err)
	}
	return &l, nil
}

func (s *Storage) UpdateContactList(l *ContactList) error {
	_, err := s.db.Exec(
		`UPDATE contact_lists SET name=?, description=?, type=?, account=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		l.Name, l.Description, l.Type, l.Account, l.ID,
	)
	return err
}

func (s *Storage) DeleteContactList(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM contact_list_members WHERE list_id = ?`, id); err != nil {
		return fmt.Errorf("deleting list members: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM contact_lists WHERE id = ?`, id); err != nil {
		return fmt.Errorf("deleting contact list: %w", err)
	}
	return tx.Commit()
}

func (s *Storage) ListContactLists(account string) ([]ContactList, error) {
	rows, err := s.db.Query(
		`SELECT `+listSelectCols+` FROM contact_lists WHERE account = ? ORDER BY created_at DESC`, account,
	)
	if err != nil {
		return nil, fmt.Errorf("listing contact lists: %w", err)
	}
	defer rows.Close()

	var lists []ContactList
	for rows.Next() {
		var l ContactList
		if err := rows.Scan(&l.ID, &l.Name, &l.Slug, &l.Description, &l.Type, &l.Account, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning contact list: %w", err)
		}
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

func (s *Storage) AddListMember(contactID, listID int64) error {
	_, err := s.db.Exec(
		`INSERT INTO contact_list_members (contact_id, list_id) VALUES (?, ?)`,
		contactID, listID,
	)
	return err
}

func (s *Storage) RemoveListMember(contactID, listID int64) error {
	_, err := s.db.Exec(
		`DELETE FROM contact_list_members WHERE contact_id = ? AND list_id = ?`,
		contactID, listID,
	)
	return err
}
