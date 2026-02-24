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
	Metadata  string    `json:"metadata,omitempty"` // JSON string storing agent attribution
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
	return s.ensureSessionForUser(sessionID, 0)
}

// ensureSessionForUser creates a session record if one doesn't exist for the given sessionID and user.
func (s *Storage) ensureSessionForUser(sessionID string, userKey interface{}) error {
	var query string
	var args []interface{}

	if s.isUserDB {
		query = `INSERT OR IGNORE INTO sessions (session_id) VALUES (?)`
		args = []interface{}{sessionID}
	} else {
		uid := normalizeUserKey(userKey)
		query = `INSERT OR IGNORE INTO sessions (session_id, user_id) VALUES (?, ?)`
		args = []interface{}{sessionID, uid}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("ensuring session: %w", err)
	}
	return nil
}

// SaveMessage saves a message with backward-compatible user ID (1).
func (s *Storage) SaveMessage(sessionID, role, content string) error {
	return s.SaveMessageForUser(0, sessionID, role, content, "")
}

// SaveMessageForUser saves a message scoped to a specific user.
func (s *Storage) SaveMessageForUser(userKey interface{}, sessionID, role, content, metadata string) error {
	if err := s.ensureSessionForUser(sessionID, userKey); err != nil {
		return err
	}

	var query string
	var args []interface{}
	var updateQuery string
	var updateArgs []interface{}

	if s.isUserDB {
		query = `INSERT INTO chats (session_id, role, content, metadata) VALUES (?, ?, ?, ?)`
		args = []interface{}{sessionID, role, content, metadata}
		updateQuery = `UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`
		updateArgs = []interface{}{sessionID}
	} else {
		uid := normalizeUserKey(userKey)
		query = `INSERT INTO chats (session_id, user_id, role, content, metadata) VALUES (?, ?, ?, ?, ?)`
		args = []interface{}{sessionID, uid, role, content, metadata}
		updateQuery = `UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE session_id = ? AND user_id = ?`
		updateArgs = []interface{}{sessionID, uid}
	}

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("saving message: %w", err)
	}
	// Touch session updated_at
	_, _ = s.db.Exec(updateQuery, updateArgs...)
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

	var stmtWithTime, stmtNoTime *sql.Stmt
	if s.isUserDB {
		stmtWithTime, err = tx.Prepare(`INSERT INTO chats (session_id, role, content, created_at) VALUES (?, ?, ?, ?)`)
		stmtNoTime, err = tx.Prepare(`INSERT INTO chats (session_id, role, content) VALUES (?, ?, ?)`)
	} else {
		// This is used for migration where we don't necessarily have the userID here
		// But ImportMessages signature currently doesn't take userID.
		// For backward compatibility with GetMessages() which uses userID=0, use 0 here.
		stmtWithTime, err = tx.Prepare(`INSERT INTO chats (session_id, user_id, role, content, created_at) VALUES (?, 0, ?, ?, ?)`)
		stmtNoTime, err = tx.Prepare(`INSERT INTO chats (session_id, user_id, role, content) VALUES (?, 0, ?, ?)`)
	}
	if err != nil {
		return 0, fmt.Errorf("prepare stmts: %w", err)
	}
	defer stmtWithTime.Close()
	defer stmtNoTime.Close()

	count := 0
	for _, m := range msgs {
		if m.Role == "" || m.Content == "" {
			continue
		}
		if !m.CreatedAt.IsZero() {
			var args []interface{}
			if s.isUserDB {
				args = []interface{}{sessionID, m.Role, m.Content, m.CreatedAt}
			} else {
				args = []interface{}{sessionID, m.Role, m.Content, m.CreatedAt}
			}
			if _, err := stmtWithTime.Exec(args...); err != nil {
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
	return s.GetMessagesForUser(0, sessionID)
}

func (s *Storage) GetMessagesForUser(userKey interface{}, sessionID string) ([]Message, error) {
	var query string
	var args []interface{}

	if s.isUserDB {
		query = `SELECT id, session_id, role, content, created_at, metadata FROM chats WHERE session_id = ? ORDER BY created_at ASC`
		args = []interface{}{sessionID}
	} else {
		uid := normalizeUserKey(userKey)
		query = `SELECT id, session_id, role, content, created_at, metadata FROM chats WHERE session_id = ? AND user_id = ? ORDER BY created_at ASC`
		args = []interface{}{sessionID, uid}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.CreatedAt, &msg.Metadata); err != nil {
			return nil, fmt.Errorf("scanning message: %w", err)
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (s *Storage) SearchMessages(query string) ([]Message, error) {
	return s.SearchMessagesForUser(0, query)
}

func (s *Storage) SearchMessagesForUser(userKey interface{}, query string) ([]Message, error) {
	var sqlQuery string
	var args []interface{}

	if s.isUserDB {
		sqlQuery = `SELECT id, session_id, role, content, created_at FROM chats WHERE content LIKE ? ORDER BY created_at DESC LIMIT 50`
		args = []interface{}{"%" + query + "%"}
	} else {
		uid := normalizeUserKey(userKey)
		sqlQuery = `SELECT id, session_id, role, content, created_at FROM chats WHERE user_id = ? AND content LIKE ? ORDER BY created_at DESC LIMIT 50`
		args = []interface{}{uid, "%" + query + "%"}
	}

	rows, err := s.db.Query(sqlQuery, args...)
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
	return messages, nil
}

// ListSessions returns session summaries with optional archived filter and pagination.
// If archived is nil, returns only non-archived sessions.
// If archived is non-nil, filters by that value.
func (s *Storage) ListSessions(archived *bool, limit, offset int) ([]SessionSummary, error) {
	return s.ListSessionsForUser(0, archived, limit, offset)
}

func (s *Storage) ListSessionsForUser(userKey interface{}, archived *bool, limit, offset int) ([]SessionSummary, error) {
	if limit <= 0 {
		limit = 50
	}

	archivedFilter := false
	if archived != nil {
		archivedFilter = *archived
	}

	var query string
	var args []interface{}

	if s.isUserDB {
		query = `
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
		args = []interface{}{archivedFilter, limit, offset}
	} else {
		uid := normalizeUserKey(userKey)
		query = `
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
				WHERE user_id = ?
				GROUP BY session_id
			) counts ON sess.session_id = counts.session_id
			LEFT JOIN chats c ON c.session_id = counts.session_id AND c.id = counts.max_id
			WHERE sess.user_id = ? AND sess.archived = ?
			ORDER BY COALESCE(c.created_at, sess.updated_at) DESC
			LIMIT ? OFFSET ?
		`
		args = []interface{}{uid, uid, archivedFilter, limit, offset}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		if !isSessionSchemaCompatibilityError(err) {
			return nil, fmt.Errorf("listing sessions: %w", err)
		}
		// Legacy fallback for databases without sessions table/columns.
		if archivedFilter {
			return []SessionSummary{}, nil
		}
		var legacyQuery string
		var legacyArgs []interface{}
		if s.isUserDB {
			legacyQuery = `
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
			legacyArgs = []interface{}{limit, offset}
		} else {
			uid := normalizeUserKey(userKey)
			legacyQuery = `
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
					WHERE user_id = ?
					GROUP BY session_id
				) counts
				LEFT JOIN chats c ON c.session_id = counts.session_id AND c.id = counts.max_id
				ORDER BY COALESCE(c.created_at, counts.last_created_at) DESC
				LIMIT ? OFFSET ?
			`
			legacyArgs = []interface{}{uid, limit, offset}
		}
		rows, err = s.db.Query(legacyQuery, legacyArgs...)
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
	return s.GetSessionForUser(0, sessionID)
}

func (s *Storage) GetSessionForUser(userKey interface{}, sessionID string) (*Session, error) {
	var query string
	var args []interface{}

	if s.isUserDB {
		query = `SELECT id, session_id, COALESCE(title, ''), archived, created_at, updated_at FROM sessions WHERE session_id = ?`
		args = []interface{}{sessionID}
	} else {
		uid := normalizeUserKey(userKey)
		query = `SELECT id, session_id, COALESCE(title, ''), archived, created_at, updated_at FROM sessions WHERE session_id = ? AND user_id = ?`
		args = []interface{}{sessionID, uid}
	}

	var sess Session
	err := s.db.QueryRow(query, args...).Scan(&sess.ID, &sess.SessionID, &sess.Title, &sess.Archived, &sess.CreatedAt, &sess.UpdatedAt)
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
	return s.UpdateSessionForUser(0, sessionID, title, archived)
}

func (s *Storage) UpdateSessionForUser(userKey interface{}, sessionID string, title *string, archived *bool) (*Session, error) {
	if title != nil {
		var query string
		var args []interface{}
		if s.isUserDB {
			query = `UPDATE sessions SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`
			args = []interface{}{*title, sessionID}
		} else {
			uid := normalizeUserKey(userKey)
			query = `UPDATE sessions SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ? AND user_id = ?`
			args = []interface{}{*title, sessionID, uid}
		}
		if _, err := s.db.Exec(query, args...); err != nil {
			return nil, fmt.Errorf("updating session title: %w", err)
		}
	}
	if archived != nil {
		var query string
		var args []interface{}
		if s.isUserDB {
			query = `UPDATE sessions SET archived = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ?`
			args = []interface{}{*archived, sessionID}
		} else {
			uid := normalizeUserKey(userKey)
			query = `UPDATE sessions SET archived = ?, updated_at = CURRENT_TIMESTAMP WHERE session_id = ? AND user_id = ?`
			args = []interface{}{*archived, sessionID, uid}
		}
		if _, err := s.db.Exec(query, args...); err != nil {
			return nil, fmt.Errorf("updating session archived: %w", err)
		}
	}
	return s.GetSessionForUser(userKey, sessionID)
}

// DeleteSession permanently removes a session and all its messages.
func (s *Storage) DeleteSession(sessionID string) error {
	return s.DeleteSessionForUser(0, sessionID)
}

func (s *Storage) DeleteSessionForUser(userKey interface{}, sessionID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if s.isUserDB {
		if _, err := tx.Exec(`DELETE FROM chats WHERE session_id = ?`, sessionID); err != nil {
			return fmt.Errorf("deleting messages: %w", err)
		}
		if _, err := tx.Exec(`DELETE FROM sessions WHERE session_id = ?`, sessionID); err != nil {
			return fmt.Errorf("deleting session: %w", err)
		}
	} else {
		uid := normalizeUserKey(userKey)
		if _, err := tx.Exec(`DELETE FROM chats WHERE session_id = ? AND user_id = ?`, sessionID, uid); err != nil {
			return fmt.Errorf("deleting messages: %w", err)
		}
		if _, err := tx.Exec(`DELETE FROM sessions WHERE session_id = ? AND user_id = ?`, sessionID, uid); err != nil {
			return fmt.Errorf("deleting session: %w", err)
		}
	}
	return tx.Commit()
}

// ForkSession copies messages from sourceSession up to (and including) messageID
// into a new session with the given newSessionID.
// If messageID is 0, all messages are copied.
func (s *Storage) ForkSession(sourceSession, newSessionID string, messageID int64) (int, error) {
	return s.ForkSessionForUser(0, sourceSession, newSessionID, messageID)
}

func (s *Storage) ForkSessionForUser(userKey interface{}, sourceSession, newSessionID string, messageID int64) (int, error) {
	messages, err := s.GetMessagesForUser(userKey, sourceSession)
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
	if err := s.ensureSessionForUser(newSessionID, userKey); err != nil {
		return 0, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var query string
	if s.isUserDB {
		query = `INSERT INTO chats (session_id, role, content, created_at) VALUES (?, ?, ?, ?)`
	} else {
		query = `INSERT INTO chats (session_id, user_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, m := range toFork {
		var args []interface{}
		if s.isUserDB {
			args = []interface{}{newSessionID, m.Role, m.Content, m.CreatedAt}
		} else {
			uid := normalizeUserKey(userKey)
			args = []interface{}{newSessionID, uid, m.Role, m.Content, m.CreatedAt}
		}
		if _, err := stmt.Exec(args...); err != nil {
			return 0, fmt.Errorf("insert: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return len(toFork), nil
}
