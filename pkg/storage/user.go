package storage

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                int64      `json:"id"`
	UUID              string     `json:"uuid"` // Unique identifier for filesystem/workspace paths
	Username          string     `json:"username"`
	Email             string     `json:"email,omitempty"`
	PasswordHash      string     `json:"-"`
	Role              string     `json:"role"` // 'admin' or 'user'
	FullName          string     `json:"full_name,omitempty"`
	DateOfBirth       *time.Time `json:"date_of_birth,omitempty"`
	Timezone          string     `json:"timezone,omitempty"`
	PreferredLanguage string     `json:"preferred_language,omitempty"`
	AvatarURL         string     `json:"avatar_url,omitempty"`
	Blocked           bool       `json:"blocked"`
	BlockedReason     string     `json:"blocked_reason,omitempty"`
	BlockedBy         *int64     `json:"blocked_by,omitempty"`
	BlockedAt         *time.Time `json:"blocked_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

var ErrUserNotFound = errors.New("user not found")
var ErrUserExists = errors.New("user already exists")
var ErrUserBlocked = errors.New("user is blocked")
var ErrCannotBlockSelf = errors.New("cannot block yourself")
var ErrCannotBlockLastAdmin = errors.New("cannot block the last admin")

// CountUsers returns the total number of users.
func (s *Storage) CountUsers() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// CountAdmins returns the total number of admin users.
func (s *Storage) CountAdmins() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	return count, err
}

// CreateUser creates a new user with an automatically generated UUID.
func (s *Storage) CreateUser(username, password, role string) (*User, error) {
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

	// Generate UUID for user
	userUUID := uuid.New().String()

	res, err := s.db.Exec(`
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

	return s.GetUserByID(id)
}

// CreateUserWithEmail creates a new user with email and auto-generated UUID.
func (s *Storage) CreateUserWithEmail(username, email, password, role string) (*User, error) {
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

	res, err := s.db.Exec(`
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

	return s.GetUserByID(id)
}

// legacyUserSelectCols is the SELECT column list for the legacy (non-central) users table.
const legacyUserSelectCols = `id, COALESCE(uuid, ''), username, COALESCE(email, ''), password_hash, role,
	COALESCE(full_name, ''), date_of_birth, COALESCE(timezone, ''), COALESCE(preferred_language, ''), COALESCE(avatar_url, ''),
	COALESCE(blocked, 0), COALESCE(blocked_reason, ''), blocked_by, blocked_at,
	created_at, updated_at`

func (s *Storage) scanUser(scanner interface {
	Scan(dest ...interface{}) error
}) (*User, error) {
	u := &User{}
	var email sql.NullString
	var fullName sql.NullString
	var dob sql.NullTime
	var tz sql.NullString
	var lang sql.NullString
	var avatar sql.NullString
	var blocked sql.NullBool
	var blockedReason sql.NullString
	var blockedBy sql.NullInt64
	var blockedAt sql.NullTime
	err := scanner.Scan(&u.ID, &u.UUID, &u.Username, &email, &u.PasswordHash, &u.Role,
		&fullName, &dob, &tz, &lang, &avatar,
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

func (s *Storage) GetUserByID(id int64) (*User, error) {
	row := s.db.QueryRow(`SELECT `+legacyUserSelectCols+` FROM users WHERE id = ?`, id)
	return s.scanUser(row)
}

func (s *Storage) GetUserByUUID(userUUID string) (*User, error) {
	row := s.db.QueryRow(`SELECT `+legacyUserSelectCols+` FROM users WHERE uuid = ?`, userUUID)
	return s.scanUser(row)
}

func (s *Storage) GetUserByUsername(username string) (*User, error) {
	row := s.db.QueryRow(`SELECT `+legacyUserSelectCols+` FROM users WHERE username = ? COLLATE NOCASE`, username)
	return s.scanUser(row)
}

func (s *Storage) GetUserByEmail(email string) (*User, error) {
	row := s.db.QueryRow(`SELECT `+legacyUserSelectCols+` FROM users WHERE email = ? COLLATE NOCASE`, email)
	return s.scanUser(row)
}

func (s *Storage) ListUsers() ([]*User, error) {
	rows, err := s.db.Query(`SELECT ` + legacyUserSelectCols + ` FROM users ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u, err := s.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *Storage) UpdateUserPassword(id int64, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, string(hash), id)
	return err
}

func (s *Storage) UpdateUserRole(id int64, role string) error {
	_, err := s.db.Exec(`UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, role, id)
	return err
}

func (s *Storage) DeleteUser(id int64) error {
	_, err := s.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

// UpdateUserProfile updates profile fields for a user.
func (s *Storage) UpdateUserProfile(id int64, username, email, fullName string, dateOfBirth *time.Time, timezone, preferredLanguage, avatarURL string) error {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if username == "" {
		return errors.New("username cannot be empty")
	}

	_, err := s.db.Exec(`
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
func (s *Storage) BlockUser(userID int64, adminID int64, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return errors.New("block reason cannot be empty")
	}

	// Prevent blocking self
	if userID == adminID {
		return ErrCannotBlockSelf
	}

	// Prevent blocking last admin
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}
	if user.Role == "admin" {
		// Count remaining active admins
		var adminCount int
		err := s.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'admin' AND (blocked = 0 OR blocked IS NULL)`).Scan(&adminCount)
		if err != nil {
			return err
		}
		if adminCount <= 1 {
			return ErrCannotBlockLastAdmin
		}
	}

	_, err = s.db.Exec(`
		UPDATE users 
		SET blocked = 1, blocked_reason = ?, blocked_by = ?, blocked_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		reason, adminID, userID)
	return err
}

// UnblockUser unblocks a user by resetting blocked fields.
func (s *Storage) UnblockUser(userID int64) error {
	_, err := s.db.Exec(`
		UPDATE users 
		SET blocked = 0, blocked_reason = NULL, blocked_by = NULL, blocked_at = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		userID)
	return err
}

// IsUserBlocked checks if a user is blocked and returns their full user record.
func (s *Storage) IsUserBlocked(userID int64) (bool, *User, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return false, nil, err
	}
	return user.Blocked, user, nil
}

// IsEmailBlocked checks if an email belongs to a blocked user.
func (s *Storage) IsEmailBlocked(email string) (bool, error) {
	if email == "" {
		return false, nil
	}
	var blocked bool
	err := s.db.QueryRow(`SELECT COALESCE(blocked, 0) FROM users WHERE email = ? COLLATE NOCASE`, email).Scan(&blocked)
	if err == sql.ErrNoRows {
		return false, nil // Email doesn't exist
	}
	if err != nil {
		return false, err
	}
	return blocked, nil
}

// Settings logic
func (s *Storage) GetSetting(key string) (string, error) {
	var val string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil // Not found is not an error, return empty string
	}
	return val, err
}

func (s *Storage) SetSetting(key, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value)
	return err
}
