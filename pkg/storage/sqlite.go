package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sipeed/kakoclaw/pkg/config"
	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func New(cfg config.StorageConfig) (*Storage, error) {
	path := expandHome(cfg.Path)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("creating storage directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	// SQLite optimizations for single-user personal app:
	// WAL mode gives ~10x faster writes and is crash-safe.
	// synchronous=NORMAL is safe with WAL (only FULL adds extra fsync per commit).
	// foreign_keys enables ON DELETE CASCADE for knowledge/workflow tables.
	// busy_timeout avoids immediate SQLITE_BUSY errors from concurrent goroutines.
	// cache_size=-8000 sets ~8MB page cache (negligible for a desktop app).
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
		"PRAGMA cache_size=-8000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return nil, fmt.Errorf("setting %s: %w", p, err)
		}
	}

	// SQLite performs best with a single writer connection.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrating database: %w", err)
	}

	return s, nil
}

func (s *Storage) Close() error {
	// Consolidate WAL into the main database file for a clean single-file state.
	_, _ = s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return s.db.Close()
}

func (s *Storage) migrate() error {
	// Simple migration: create tables if not exist
	// In the future, we can use a proper migration tool or versioning.

	queries := []string{
		`CREATE TABLE IF NOT EXISTS chats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_chats_session_id ON chats(session_id);`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'todo',
			result TEXT NOT NULL DEFAULT '',
			archived BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		// Migration for existing tables
		`ALTER TABLE tasks ADD COLUMN archived BOOLEAN DEFAULT 0;`,
		`ALTER TABLE tasks ADD COLUMN result TEXT;`,
		// Ensure no NULLs for fields that are scanned into strings
		`UPDATE tasks SET description = '' WHERE description IS NULL;`,
		`UPDATE tasks SET result = '' WHERE result IS NULL;`,
		`UPDATE tasks SET status = 'todo' WHERE status IS NULL;`,
		`UPDATE tasks SET archived = 0 WHERE archived IS NULL;`,
		// Task logs table for tracking task execution events
		`CREATE TABLE IF NOT EXISTS task_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			event TEXT NOT NULL,
			message TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_task_logs_task_id ON task_logs(task_id);`,
		// Sessions table for chat session management
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL DEFAULT '',
			archived BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		// Migration for existing sessions tables
		`ALTER TABLE sessions ADD COLUMN title TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE sessions ADD COLUMN archived BOOLEAN NOT NULL DEFAULT 0;`,
		`ALTER TABLE sessions ADD COLUMN created_at DATETIME DEFAULT CURRENT_TIMESTAMP;`,
		`ALTER TABLE sessions ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP;`,
		`UPDATE sessions SET title = '' WHERE title IS NULL;`,
		`UPDATE sessions SET archived = 0 WHERE archived IS NULL;`,
		`UPDATE sessions SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;`,
		`UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);`,
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		// Migration: Add uuid column if it doesn't exist
		`ALTER TABLE users ADD COLUMN uuid TEXT UNIQUE;`,
		// Settings table for global configuration
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
		// Channel user mapping table
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
		// Appending user_id to existing tables
		`ALTER TABLE chats ADD COLUMN user_id INTEGER DEFAULT 1;`,
		`ALTER TABLE tasks ADD COLUMN user_id INTEGER DEFAULT 1;`,
		`ALTER TABLE sessions ADD COLUMN user_id INTEGER DEFAULT 1;`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			// ALTER TABLE errors (duplicate column, etc.) are safe to ignore
			// since we use idempotent CREATE IF NOT EXISTS + additive ALTERs.
			if strings.HasPrefix(query, "ALTER TABLE") {
				continue
			}
			return fmt.Errorf("executing migration query: %w", err)
		}
	}

	// Knowledge base tables (FTS5)
	if err := s.migrateKnowledge(); err != nil {
		return fmt.Errorf("knowledge migration: %w", err)
	}

	// Workflow builder tables
	if err := s.migrateWorkflows(); err != nil {
		return fmt.Errorf("workflow migration: %w", err)
	}

	// Prompt templates table
	if err := s.migratePrompts(); err != nil {
		return fmt.Errorf("prompts migration: %w", err)
	}

	// Backfill sessions table from existing chats
	if err := s.migrateSessions(); err != nil {
		return fmt.Errorf("session migration: %w", err)
	}

	// Observability metrics tables
	if err := s.migrateMetrics(); err != nil {
		return fmt.Errorf("metrics migration: %w", err)
	}

	return nil
}

func (s *Storage) migrateMetrics() error {
	queries := []string{
		// Aggregated counters (one row per metric key)
		`CREATE TABLE IF NOT EXISTS metrics_counters (
			key   TEXT PRIMARY KEY,
			value INTEGER NOT NULL DEFAULT 0
		);`,
		// Recent events ring buffer (JSON payloads)
		`CREATE TABLE IF NOT EXISTS metrics_events (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			payload    TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_events_id ON metrics_events(id DESC);`,
		// Model and Tool breakdowns as JSON blobs
		`CREATE TABLE IF NOT EXISTS metrics_breakdowns (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("metrics migration query: %w", err)
		}
	}
	return nil
}

func expandHome(path string) string {
	if path == "" {
		return path
	}
	if path[0] == '~' {
		home, _ := os.UserHomeDir()
		if len(path) > 1 && path[1] == '/' {
			return home + path[1:]
		}
		return home
	}
	return path
}
