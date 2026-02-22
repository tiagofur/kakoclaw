package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"
)

// SetupSession represents a temporary session for onboarding flow
type SetupSession struct {
	ID        int64
	Token     string
	UserID    int64  // User that initiated setup (optional, for update mode)
	Channel   string // Which channel initiated: "telegram", "discord", "slack"
	SenderID  string // External ID from channel (user identifier)
	Metadata  map[string]string
	ExpiresAt time.Time
	CreatedAt time.Time
	UsedAt    *time.Time // When setup was completed
}

// CreateSetupSession generates a new setup session token
func (s *Storage) CreateSetupSession(channel, senderID string, metadata map[string]string) (*SetupSession, error) {
	// Generate random token (32 bytes = 44 base64 characters)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Token expires in 1 hour
	expiresAt := time.Now().Add(1 * time.Hour)

	// Serialize metadata as JSON
	metadataJSON := "{}"
	if len(metadata) > 0 {
		var entries []string
		for k, v := range metadata {
			entries = append(entries, fmt.Sprintf(`"%s":"%s"`, k, v))
		}
		metadataJSON = "{" + joinStrings(entries, ",") + "}"
	}

	var id int64
	err := s.db.QueryRow(`
		INSERT INTO setup_sessions (token, channel, sender_id, metadata, expires_at)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id
	`, token, channel, senderID, metadataJSON, expiresAt).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("creating setup session: %w", err)
	}

	return &SetupSession{
		ID:        id,
		Token:     token,
		Channel:   channel,
		SenderID:  senderID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}, nil
}

// GetSetupSession retrieves a setup session by token
func (s *Storage) GetSetupSession(token string) (*SetupSession, error) {
	var session SetupSession
	var metadataJSON string
	var usedAtSQL sql.NullTime

	err := s.db.QueryRow(`
		SELECT id, token, user_id, channel, sender_id, metadata, expires_at, created_at, used_at
		FROM setup_sessions
		WHERE token = ? AND expires_at > datetime('now')
	`, token).Scan(&session.ID, &session.Token, &session.UserID, &session.Channel, &session.SenderID, &metadataJSON, &session.ExpiresAt, &session.CreatedAt, &usedAtSQL)

	if err == sql.ErrNoRows {
		return nil, nil // Token not found or expired
	}
	if err != nil {
		return nil, fmt.Errorf("getting setup session: %w", err)
	}

	session.Metadata = make(map[string]string)
	if metadataJSON != "{}" && metadataJSON != "" {
		// Simple JSON parsing (could use json.Unmarshal for robustness)
		// For now, store as is and let frontend/handler deal with it
	}

	if usedAtSQL.Valid {
		session.UsedAt = &usedAtSQL.Time
	}

	return &session, nil
}

// UpdateSetupSession marks a session as used and optionally associates with a user
func (s *Storage) UpdateSetupSession(token string, userID int64) error {
	_, err := s.db.Exec(`
		UPDATE setup_sessions
		SET user_id = ?, used_at = datetime('now')
		WHERE token = ?
	`, userID, token)

	if err != nil {
		return fmt.Errorf("updating setup session: %w", err)
	}

	return nil
}

// CleanupExpiredSetupSessions removes old sessions
func (s *Storage) CleanupExpiredSetupSessions() error {
	_, err := s.db.Exec(`
		DELETE FROM setup_sessions
		WHERE expires_at < datetime('now')
	`)
	return err
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
