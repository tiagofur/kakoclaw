package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Message struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Session represents a chat session record in the sessions table.
type Session struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	Title     string    `json:"title"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SessionSummary is the API response type combining session metadata with chat data.
type SessionSummary struct {
	SessionID    string    `json:"session_id"`
	Title        string    `json:"title"`
	Archived     bool      `json:"archived"`
	LastMessage  string    `json:"last_message"`
	UpdatedAt    time.Time `json:"updated_at"`
	MessageCount int       `json:"message_count"`
}

// ImportMessage is used for bulk-importing messages into a session.
type ImportMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// migrateSessions backfills the sessions table from existing chats that don't have a session record yet.
func (s *Storage) migrateSessions() error {
	query := `
		INSERT OR IGNORE INTO sessions (session_id, title, created_at, updated_at)
		SELECT 
			c.session_id,
			'',
			MIN(c.created_at),
			MAX(c.created_at)
		FROM chats c
		LEFT JOIN sessions s ON s.session_id = c.session_id
		WHERE s.id IS NULL
		GROUP BY c.session_id
	`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("backfilling sessions: %w", err)
	}
	return nil
}

// ensureSession creates a session record if one doesn't exist for the given sessionID.
func (s *Storage) ensureSession(sessionID string) error {
	query := `INSERT OR IGNORE INTO sessions (session_id) VALUES (?)`
	_, err := s.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("ensuring session: %w", err)
	}
	return nil
}

func (s *Storage) SaveMessage(sessionID, role, content string) error {
	if err := s.ensureSession(sessionID); err != nil {
		return err
	}
	query := `INSERT INTO chats (session_id, role, content) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, sessionID, role, content)
	if err != nil {
		return fmt.Errorf("saving message: %w", err)
	}
	// Touch session updated_at
	_, _ = s.db.Exec(`UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`, sessionID)
	return nil
}

func (s *Storage) ImportMessages(sessionID string, msgs []ImportMessage) (int, error) {
	if err := s.ensureSession(sessionID); err != nil {
		return 0, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmtWithTime, err := tx.Prepare(`INSERT INTO chats (session_id, role, content, created_at) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("prepare with time: %w", err)
	}
	defer stmtWithTime.Close()

	stmtNoTime, err := tx.Prepare(`INSERT INTO chats (session_id, role, content) VALUES (?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("prepare no time: %w", err)
	}
	defer stmtNoTime.Close()

	count := 0
	for _, m := range msgs {
		if m.Role == "" || m.Content == "" {
			continue
		}
		if !m.CreatedAt.IsZero() {
			if _, err := stmtWithTime.Exec(sessionID, m.Role, m.Content, m.CreatedAt); err != nil {
				return count, fmt.Errorf("insert message %d: %w", count, err)
			}
		} else {
			if _, err := stmtNoTime.Exec(sessionID, m.Role, m.Content); err != nil {
				return count, fmt.Errorf("insert message %d: %w", count, err)
			}
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	// Touch session updated_at
	_, _ = s.db.Exec(`UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`, sessionID)
	return count, nil
}

func (s *Storage) GetMessages(sessionID string) ([]Message, error) {
	query := `SELECT id, session_id, role, content, created_at FROM chats WHERE session_id = ? ORDER BY created_at ASC`
	rows, err := s.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("querying messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning message: %w", err)
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating messages: %w", err)
	}
	return messages, nil
}

func (s *Storage) SearchMessages(query string) ([]Message, error) {
	sqlQuery := `SELECT id, session_id, role, content, created_at FROM chats WHERE content LIKE ? ORDER BY created_at DESC LIMIT 50`
	rows, err := s.db.Query(sqlQuery, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("searching messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning message: %w", err)
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating search results: %w", err)
	}
	return messages, nil
}

// ListSessions returns session summaries with optional archived filter and pagination.
// If archived is nil, returns only non-archived sessions.
// If archived is non-nil, filters by that value.
func (s *Storage) ListSessions(archived *bool, limit, offset int) ([]SessionSummary, error) {
	if limit <= 0 {
		limit = 50
	}

	archivedFilter := false
	if archived != nil {
		archivedFilter = *archived
	}

	query := `
		SELECT 
			sess.session_id,
			COALESCE(sess.title, ''),
			sess.archived,
			COALESCE(c.content, ''),
			COALESCE(c.created_at, sess.updated_at),
			COALESCE(counts.msg_count, 0)
		FROM sessions sess
		LEFT JOIN (
			SELECT session_id, MAX(id) AS max_id, COUNT(*) AS msg_count
			FROM chats
			GROUP BY session_id
		) counts ON sess.session_id = counts.session_id
		LEFT JOIN chats c ON c.session_id = counts.session_id AND c.id = counts.max_id
		WHERE sess.archived = ?
		ORDER BY COALESCE(c.created_at, sess.updated_at) DESC
		LIMIT ? OFFSET ?
	`
	rows, err := s.db.Query(query, archivedFilter, limit, offset)
	if err != nil {
		if !isSessionSchemaCompatibilityError(err) {
			return nil, fmt.Errorf("listing sessions: %w", err)
		}
		// Legacy fallback for databases without sessions table/columns.
		if archivedFilter {
			return []SessionSummary{}, nil
		}
		legacyQuery := `
			SELECT
				counts.session_id,
				'' AS title,
				0 AS archived,
				COALESCE(c.content, ''),
				COALESCE(c.created_at, counts.last_created_at),
				COALESCE(counts.msg_count, 0)
			FROM (
				SELECT session_id, MAX(id) AS max_id, MAX(created_at) AS last_created_at, COUNT(*) AS msg_count
				FROM chats
				GROUP BY session_id
			) counts
			LEFT JOIN chats c ON c.session_id = counts.session_id AND c.id = counts.max_id
			ORDER BY COALESCE(c.created_at, counts.last_created_at) DESC
			LIMIT ? OFFSET ?
		`
		rows, err = s.db.Query(legacyQuery, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("listing sessions (legacy fallback): %w", err)
		}
	}
	defer rows.Close()

	var sessions []SessionSummary
	for rows.Next() {
		var ss SessionSummary
		var updatedAtStr string
		if err := rows.Scan(&ss.SessionID, &ss.Title, &ss.Archived, &ss.LastMessage, &updatedAtStr, &ss.MessageCount); err != nil {
			return nil, fmt.Errorf("scanning session: %w", err)
		}
		if updatedAtStr != "" {
			for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", time.RFC3339Nano, time.DateTime} {
				if t, err := time.Parse(layout, updatedAtStr); err == nil {
					ss.UpdatedAt = t
					break
				}
			}
		}
		sessions = append(sessions, ss)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating sessions: %w", err)
	}
	return sessions, nil
}

func isSessionSchemaCompatibilityError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such table: sessions") || strings.Contains(msg, "no such column:")
}

// GetSession returns a single session by session_id.
func (s *Storage) GetSession(sessionID string) (*Session, error) {
	query := `SELECT id, session_id, COALESCE(title, ''), archived, created_at, updated_at FROM sessions WHERE session_id = ?`
	var sess Session
	err := s.db.QueryRow(query, sessionID).Scan(&sess.ID, &sess.SessionID, &sess.Title, &sess.Archived, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("getting session: %w", err)
	}
	return &sess, nil
}

// UpdateSession updates title and/or archived status for a session.
func (s *Storage) UpdateSession(sessionID string, title *string, archived *bool) (*Session, error) {
	if title != nil {
		if _, err := s.db.Exec(`UPDATE sessions SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`, *title, sessionID); err != nil {
			return nil, fmt.Errorf("updating session title: %w", err)
		}
	}
	if archived != nil {
		if _, err := s.db.Exec(`UPDATE sessions SET archived = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`, *archived, sessionID); err != nil {
			return nil, fmt.Errorf("updating session archived: %w", err)
		}
	}
	return s.GetSession(sessionID)
}

// DeleteSession permanently removes a session and all its messages.
func (s *Storage) DeleteSession(sessionID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM chats WHERE session_id = ?`, sessionID); err != nil {
		return fmt.Errorf("deleting messages: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM sessions WHERE session_id = ?`, sessionID); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return tx.Commit()
}

// ForkSession copies messages from sourceSession up to (and including) messageID
// into a new session with the given newSessionID.
// If messageID is 0, all messages are copied.
func (s *Storage) ForkSession(sourceSession, newSessionID string, messageID int64) (int, error) {
	messages, err := s.GetMessages(sourceSession)
	if err != nil {
		return 0, fmt.Errorf("get source messages: %w", err)
	}

	if len(messages) == 0 {
		return 0, fmt.Errorf("source session is empty")
	}

	// Filter up to messageID
	var toFork []Message
	if messageID > 0 {
		for _, m := range messages {
			toFork = append(toFork, m)
			if m.ID == messageID {
				break
			}
		}
		if len(toFork) == 0 {
			return 0, fmt.Errorf("message %d not found in session", messageID)
		}
	} else {
		toFork = messages
	}

	// Ensure session record exists for the new fork
	if err := s.ensureSession(newSessionID); err != nil {
		return 0, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO chats (session_id, role, content, created_at) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, m := range toFork {
		if _, err := stmt.Exec(newSessionID, m.Role, m.Content, m.CreatedAt); err != nil {
			return 0, fmt.Errorf("insert: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return len(toFork), nil
}
