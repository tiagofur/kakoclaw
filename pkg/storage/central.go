package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// CentralStorage manages the central database that holds ONLY authentication
// and global data: users, settings, channel-to-user mappings, and setup sessions.
// Per-user data (chats, tasks, knowledge, etc.) lives in per-user databases
// managed by UserStorageManager.
type CentralStorage struct {
	db *sql.DB
}

// NewCentral opens (or creates) the central database at the given path.
func NewCentral(dbPath string) (*CentralStorage, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("creating central storage directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening central database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging central database: %w", err)
	}

	// SQLite optimizations
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
		"PRAGMA cache_size=-4000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return nil, fmt.Errorf("setting %s: %w", p, err)
		}
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	cs := &CentralStorage{db: db}
	if err := cs.migrate(); err != nil {
		return nil, fmt.Errorf("migrating central database: %w", err)
	}

	return cs, nil
}

// Close closes the central database.
func (cs *CentralStorage) Close() error {
	_, _ = cs.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return cs.db.Close()
}

// DB returns the underlying database connection (for advanced use).
func (cs *CentralStorage) DB() *sql.DB {
	return cs.db
}

func (cs *CentralStorage) migrate() error {
	queries := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE,
			username TEXT NOT NULL UNIQUE,
			email TEXT,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			full_name TEXT,
			date_of_birth DATE,
			timezone TEXT,
			preferred_language TEXT,
			avatar_url TEXT,
			allowed_tools TEXT,
			blocked BOOLEAN NOT NULL DEFAULT 0,
			blocked_reason TEXT,
			blocked_by INTEGER,
			blocked_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_uuid ON users(uuid);`,
		`CREATE INDEX IF NOT EXISTS idx_users_blocked ON users(blocked);`,

		// Add new profile columns to existing deployments (safe: no-op if column exists)
		`ALTER TABLE users ADD COLUMN full_name TEXT;`,
		`ALTER TABLE users ADD COLUMN date_of_birth DATE;`,
		`ALTER TABLE users ADD COLUMN timezone TEXT;`,
		`ALTER TABLE users ADD COLUMN preferred_language TEXT;`,
		`ALTER TABLE users ADD COLUMN avatar_url TEXT;`,
		// Add tool permissions column
		`ALTER TABLE users ADD COLUMN allowed_tools TEXT;`,

		// Settings table (global settings: jwt_secret, etc.)
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,

		// Channel-to-user mapping table
		`CREATE TABLE IF NOT EXISTS channel_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			channel TEXT NOT NULL,
			sender_id TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(channel, sender_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_channel_users_channel_sender ON channel_users(channel, sender_id);`,

		// Setup sessions table for onboarding flow
		`CREATE TABLE IF NOT EXISTS setup_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token TEXT NOT NULL UNIQUE,
			user_id INTEGER,
			channel TEXT NOT NULL,
			sender_id TEXT NOT NULL,
			metadata TEXT NOT NULL DEFAULT '{}',
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			used_at DATETIME
		);`,
		`CREATE INDEX IF NOT EXISTS idx_setup_sessions_token ON setup_sessions(token);`,
		`CREATE INDEX IF NOT EXISTS idx_setup_sessions_expires_at ON setup_sessions(expires_at);`,
	}

	for _, query := range queries {
		if _, err := cs.db.Exec(query); err != nil {
			if strings.HasPrefix(query, "ALTER TABLE") {
				continue // column already exists — safe to ignore
			}
			return fmt.Errorf("executing central migration: %w", err)
		}
	}

	return nil
}

// ==================== USER CRUD ====================

// CountUsers returns the total number of users.
func (cs *CentralStorage) CountUsers() (int, error) {
	var count int
	err := cs.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// CountAdmins returns the total number of admin users.
func (cs *CentralStorage) CountAdmins() (int, error) {
	var count int
	err := cs.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	return count, err
}

// CreateUser creates a new user with an automatically generated UUID.
func (cs *CentralStorage) CreateUser(username, password, role string) (*User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if role == "" {
		role = "user"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userUUID := uuid.New().String()

	res, err := cs.db.Exec(`
		INSERT INTO users (username, password_hash, role, uuid)
		VALUES (?, ?, ?, ?)`,
		username, string(hash), role, userUUID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrUserExists
		}
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return cs.GetUserByID(id)
}

// CreateUserWithEmail creates a new user with email and auto-generated UUID.
func (cs *CentralStorage) CreateUserWithEmail(username, email, password, role string) (*User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if role == "" {
		role = "user"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userUUID := uuid.New().String()

	res, err := cs.db.Exec(`
		INSERT INTO users (username, email, password_hash, role, uuid)
		VALUES (?, ?, ?, ?, ?)`,
		username, email, string(hash), role, userUUID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrUserExists
		}
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return cs.GetUserByID(id)
}

func (cs *CentralStorage) scanUser(scanner interface {
	Scan(dest ...interface{}) error
}) (*User, error) {
	u := &User{}
	var email sql.NullString
	var fullName sql.NullString
	var dob sql.NullTime
	var tz sql.NullString
	var lang sql.NullString
	var avatar sql.NullString
	var allowedTools sql.NullString
	var blocked sql.NullBool
	var blockedReason sql.NullString
	var blockedBy sql.NullInt64
	var blockedAt sql.NullTime
	err := scanner.Scan(&u.ID, &u.UUID, &u.Username, &email, &u.PasswordHash, &u.Role,
		&fullName, &dob, &tz, &lang, &avatar,
		&allowedTools,
		&blocked, &blockedReason, &blockedBy, &blockedAt,
		&u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if email.Valid {
		u.Email = email.String
	}
	if fullName.Valid {
		u.FullName = fullName.String
	}
	if dob.Valid {
		u.DateOfBirth = &dob.Time
	}
	if tz.Valid {
		u.Timezone = tz.String
	}
	if lang.Valid {
		u.PreferredLanguage = lang.String
	}
	if avatar.Valid {
		u.AvatarURL = avatar.String
	}
	if allowedTools.Valid && allowedTools.String != "" {
		u.AllowedTools = &allowedTools.String
	}
	if blocked.Valid {
		u.Blocked = blocked.Bool
	}
	if blockedReason.Valid {
		u.BlockedReason = blockedReason.String
	}
	if blockedBy.Valid {
		val := blockedBy.Int64
		u.BlockedBy = &val
	}
	if blockedAt.Valid {
		u.BlockedAt = &blockedAt.Time
	}
	return u, nil
}

const userSelectCols = `id, COALESCE(uuid, ''), username, COALESCE(email, ''), password_hash, role,
	COALESCE(full_name, ''), date_of_birth, COALESCE(timezone, ''), COALESCE(preferred_language, ''), COALESCE(avatar_url, ''),
	COALESCE(allowed_tools, ''),
	COALESCE(blocked, 0), COALESCE(blocked_reason, ''), blocked_by, blocked_at,
	created_at, updated_at`

func (cs *CentralStorage) GetUserByID(id int64) (*User, error) {
	row := cs.db.QueryRow(`SELECT `+userSelectCols+` FROM users WHERE id = ?`, id)
	return cs.scanUser(row)
}

func (cs *CentralStorage) GetUserByUUID(userUUID string) (*User, error) {
	row := cs.db.QueryRow(`SELECT `+userSelectCols+` FROM users WHERE uuid = ?`, userUUID)
	return cs.scanUser(row)
}

func (cs *CentralStorage) GetUserByUsername(username string) (*User, error) {
	row := cs.db.QueryRow(`SELECT `+userSelectCols+` FROM users WHERE username = ? COLLATE NOCASE`, username)
	return cs.scanUser(row)
}

func (cs *CentralStorage) GetUserByEmail(email string) (*User, error) {
	row := cs.db.QueryRow(`SELECT `+userSelectCols+` FROM users WHERE email = ? COLLATE NOCASE`, email)
	return cs.scanUser(row)
}

func (cs *CentralStorage) ListUsers() ([]*User, error) {
	rows, err := cs.db.Query(`SELECT ` + userSelectCols + ` FROM users ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u, err := cs.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (cs *CentralStorage) UpdateUserPassword(id int64, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = cs.db.Exec(`UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, string(hash), id)
	return err
}

func (cs *CentralStorage) UpdateUserRole(id int64, role string) error {
	_, err := cs.db.Exec(`UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, role, id)
	return err
}

func (cs *CentralStorage) DeleteUser(id int64) error {
	_, err := cs.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

// UpdateUserProfile updates profile fields for a user.
func (cs *CentralStorage) UpdateUserProfile(id int64, username, email, fullName string, dateOfBirth *time.Time, timezone, preferredLanguage, avatarURL string) error {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if username == "" {
		return errors.New("username cannot be empty")
	}

	_, err := cs.db.Exec(`
		UPDATE users 
		SET username = ?, email = ?, full_name = ?, date_of_birth = ?, timezone = ?, preferred_language = ?, avatar_url = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`,
		username, email, fullName, dateOfBirth, timezone, preferredLanguage, avatarURL, id)

	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ErrUserExists
	}
	return err
}

// BlockUser blocks a user with a reason and tracks who blocked them.
func (cs *CentralStorage) BlockUser(userID int64, adminID int64, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return errors.New("block reason cannot be empty")
	}

	if userID == adminID {
		return ErrCannotBlockSelf
	}

	user, err := cs.GetUserByID(userID)
	if err != nil {
		return err
	}
	if user.Role == "admin" {
		var adminCount int
		err := cs.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'admin' AND (blocked = 0 OR blocked IS NULL)`).Scan(&adminCount)
		if err != nil {
			return err
		}
		if adminCount <= 1 {
			return ErrCannotBlockLastAdmin
		}
	}

	_, err = cs.db.Exec(`
		UPDATE users 
		SET blocked = 1, blocked_reason = ?, blocked_by = ?, blocked_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		reason, adminID, userID)
	return err
}

// UnblockUser unblocks a user.
func (cs *CentralStorage) UnblockUser(userID int64) error {
	_, err := cs.db.Exec(`
		UPDATE users 
		SET blocked = 0, blocked_reason = NULL, blocked_by = NULL, blocked_at = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		userID)
	return err
}

// IsUserBlocked checks if a user is blocked and returns their full user record.
func (cs *CentralStorage) IsUserBlocked(userID int64) (bool, *User, error) {
	user, err := cs.GetUserByID(userID)
	if err != nil {
		return false, nil, err
	}
	return user.Blocked, user, nil
}

// IsEmailBlocked checks if an email belongs to a blocked user.
func (cs *CentralStorage) IsEmailBlocked(email string) (bool, error) {
	if email == "" {
		return false, nil
	}
	var blocked bool
	err := cs.db.QueryRow(`SELECT COALESCE(blocked, 0) FROM users WHERE email = ? COLLATE NOCASE`, email).Scan(&blocked)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return blocked, nil
}

// ==================== SETTINGS ====================

// GetSetting returns a global setting value.
func (cs *CentralStorage) GetSetting(key string) (string, error) {
	var val string
	err := cs.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

// SetSetting stores a global setting value.
func (cs *CentralStorage) SetSetting(key, value string) error {
	_, err := cs.db.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value)
	return err
}

// ==================== CHANNEL MAPPINGS ====================

// GetUserIDForChannelSender returns the user ID mapped to a channel sender ID.
func (cs *CentralStorage) GetUserIDForChannelSender(channel, senderID string) (int64, error) {
	var userID int64
	err := cs.db.QueryRow(
		`SELECT user_id FROM channel_users WHERE channel = ? AND sender_id = ?`,
		channel,
		senderID,
	).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// SetUserIDForChannelSender upserts a mapping between a channel sender ID and user ID.
func (cs *CentralStorage) SetUserIDForChannelSender(channel, senderID string, userID int64) error {
	_, err := cs.db.Exec(
		`INSERT INTO channel_users (channel, sender_id, user_id)
		 VALUES (?, ?, ?)
		 ON CONFLICT(channel, sender_id) DO UPDATE SET user_id = excluded.user_id, updated_at = CURRENT_TIMESTAMP`,
		channel,
		senderID,
		userID,
	)
	return err
}

// ==================== SETUP SESSIONS ====================

// CreateSetupSession generates a new setup session token.
func (cs *CentralStorage) CreateSetupSession(channel, senderID string, metadata map[string]string) (*SetupSession, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(1 * time.Hour)

	metadataJSON := "{}"
	if len(metadata) > 0 {
		var entries []string
		for k, v := range metadata {
			entries = append(entries, fmt.Sprintf(`"%s":"%s"`, k, v))
		}
		metadataJSON = "{" + joinStrings(entries, ",") + "}"
	}

	var id int64
	err := cs.db.QueryRow(`
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

// GetSetupSession retrieves a setup session by token.
func (cs *CentralStorage) GetSetupSession(token string) (*SetupSession, error) {
	var session SetupSession
	var metadataJSON string
	var usedAtSQL sql.NullTime

	err := cs.db.QueryRow(`
		SELECT id, token, user_id, channel, sender_id, metadata, expires_at, created_at, used_at
		FROM setup_sessions
		WHERE token = ? AND expires_at > datetime('now')
	`, token).Scan(&session.ID, &session.Token, &session.UserID, &session.Channel, &session.SenderID, &metadataJSON, &session.ExpiresAt, &session.CreatedAt, &usedAtSQL)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting setup session: %w", err)
	}

	session.Metadata = make(map[string]string)

	if usedAtSQL.Valid {
		session.UsedAt = &usedAtSQL.Time
	}

	return &session, nil
}

// UpdateSetupSession marks a session as used and optionally associates with a user.
func (cs *CentralStorage) UpdateSetupSession(token string, userID int64) error {
	_, err := cs.db.Exec(`
		UPDATE setup_sessions
		SET user_id = ?, used_at = datetime('now')
		WHERE token = ?
	`, userID, token)

	if err != nil {
		return fmt.Errorf("updating setup session: %w", err)
	}

	return nil
}

// CleanupExpiredSetupSessions removes old sessions.
func (cs *CentralStorage) CleanupExpiredSetupSessions() error {
	_, err := cs.db.Exec(`
		DELETE FROM setup_sessions
		WHERE expires_at < datetime('now')
	`)
	return err
}
