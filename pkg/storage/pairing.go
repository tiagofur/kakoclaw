package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// ErrCodeNotFound is returned when an approve code has no matching pending entry.
var ErrCodeNotFound = errors.New("pairing: code not found or already approved")

// PairingStore provides challenge-based allowlist CRUD backed by the user SQLite DB.
type PairingStore struct {
	db *sql.DB
}

// NewPairingStore wraps an existing *sql.DB (from Storage).
func NewPairingStore(db *sql.DB) *PairingStore {
	return &PairingStore{db: db}
}

// GenerateCode returns a 6-character hex challenge code.
func GenerateCode() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating pairing code: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// InsertPending upserts a pending challenge for (channel, senderID).
// If a pending row already exists it is a no-op (deduplication).
func (ps *PairingStore) InsertPending(channel, senderID, code string) error {
	_, err := ps.db.Exec(`
		INSERT INTO pairing_store (channel, sender_id, code, pending, created_at)
		VALUES (?, ?, ?, 1, ?)
		ON CONFLICT(channel, sender_id) DO NOTHING
	`, channel, senderID, code, time.Now().UTC())
	return err
}

// HasPending returns true when a pending (not yet approved) entry exists.
func (ps *PairingStore) HasPending(channel, senderID string) (bool, error) {
	var count int
	err := ps.db.QueryRow(`
		SELECT COUNT(*) FROM pairing_store
		WHERE channel = ? AND sender_id = ? AND pending = 1
	`, channel, senderID).Scan(&count)
	return count > 0, err
}

// IsApproved returns true when the sender has an approved entry.
func (ps *PairingStore) IsApproved(channel, senderID string) (bool, error) {
	var count int
	err := ps.db.QueryRow(`
		SELECT COUNT(*) FROM pairing_store
		WHERE channel = ? AND sender_id = ? AND pending = 0
	`, channel, senderID).Scan(&count)
	return count > 0, err
}

// Approve finds a pending row by code and marks it approved.
// Returns ErrCodeNotFound if the code is unknown or already used.
func (ps *PairingStore) Approve(channel, code string) (senderID string, err error) {
	tx, err := ps.db.Begin()
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = tx.QueryRow(`
		SELECT sender_id FROM pairing_store
		WHERE channel = ? AND code = ? AND pending = 1
	`, channel, code).Scan(&senderID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrCodeNotFound
	}
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(`
		UPDATE pairing_store SET pending = 0, approved_at = ?
		WHERE channel = ? AND code = ? AND pending = 1
	`, time.Now().UTC(), channel, code)
	if err != nil {
		return "", err
	}

	return senderID, tx.Commit()
}
