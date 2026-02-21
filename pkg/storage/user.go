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
	ID           int64     `json:"id"`
	UUID         string    `json:"uuid"` // Unique identifier for filesystem/workspace paths
	Username     string    `json:"username"`
	Email        string    `json:"email,omitempty"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"` // 'admin' or 'user'
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

var ErrUserNotFound = errors.New("user not found")
var ErrUserExists = errors.New("user already exists")

// CountUsers returns the total number of users.
func (s *Storage) CountUsers() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
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

func (s *Storage) GetUserByID(id int64) (*User, error) {
	u := &User{}
	var email sql.NullString
	err := s.db.QueryRow(`
		SELECT id, uuid, username, COALESCE(email, ''), password_hash, role, created_at, updated_at
		FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.UUID, &u.Username, &email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if email.Valid {
		u.Email = email.String
	}
	return u, nil
}

func (s *Storage) GetUserByUUID(userUUID string) (*User, error) {
	u := &User{}
	var email sql.NullString
	err := s.db.QueryRow(`
		SELECT id, uuid, username, COALESCE(email, ''), password_hash, role, created_at, updated_at
		FROM users WHERE uuid = ?`, userUUID).
		Scan(&u.ID, &u.UUID, &u.Username, &email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if email.Valid {
		u.Email = email.String
	}
	return u, nil
}

func (s *Storage) GetUserByUsername(username string) (*User, error) {
	u := &User{}
	var email sql.NullString
	err := s.db.QueryRow(`
		SELECT id, uuid, username, COALESCE(email, ''), password_hash, role, created_at, updated_at
		FROM users WHERE username = ? COLLATE NOCASE`, username).
		Scan(&u.ID, &u.UUID, &u.Username, &email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if email.Valid {
		u.Email = email.String
	}
	return u, nil
}

func (s *Storage) ListUsers() ([]*User, error) {
	rows, err := s.db.Query(`
		SELECT id, uuid, username, COALESCE(email, ''), password_hash, role, created_at, updated_at
		FROM users ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		var email sql.NullString
		if err := rows.Scan(&u.ID, &u.UUID, &u.Username, &email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if email.Valid {
			u.Email = email.String
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

// UpdateUserProfile updates username and/or email for a user.
func (s *Storage) UpdateUserProfile(id int64, username, email string) error {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if username == "" {
		return errors.New("username cannot be empty")
	}

	_, err := s.db.Exec(`
		UPDATE users 
		SET username = ?, email = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?`,
		username, email, id)

	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ErrUserExists
	}
	return err
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
