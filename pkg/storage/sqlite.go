package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sipeed/picoclaw/pkg/config"
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

	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrating database: %w", err)
	}

	return s, nil
}

func (s *Storage) Close() error {
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
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			// Ignore "duplicate column" error for ALTER TABLE
			if err.Error() == "duplicate column name: archived" {
				continue
			}
			// For modernc.org/sqlite, error message might vary, but let's try to proceed
			// or check specific error codes if available.
			// A simple way is to check if it contains "duplicate column name"
			if strings.HasPrefix(query, "ALTER TABLE") {
				// We can try to check if column exists first, but ignoring error is simpler for now
				// given the constraints.
				fmt.Printf("Migration warning (safe to ignore if column exists): %v\n", err)
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

	// Backfill sessions table from existing chats
	if err := s.migrateSessions(); err != nil {
		return fmt.Errorf("session migration: %w", err)
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
