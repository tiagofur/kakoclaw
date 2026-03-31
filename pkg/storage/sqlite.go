package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	_ "modernc.org/sqlite"
)

type Storage struct {
	db       *sql.DB
	isUserDB bool // true if this is an isolated per-user database (no user_id columns)
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

	// SQLite optimizations for single-user personal app
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

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrating database: %w", err)
	}

	return s, nil
}

func NewUserStorage(dbPath string) (*Storage, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("creating user storage directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening user database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging user database: %w", err)
	}

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

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	s := &Storage{db: db, isUserDB: true}
	if err := s.migrateUserDB(); err != nil {
		return nil, fmt.Errorf("migrating user database: %w", err)
	}

	return s, nil
}

func (s *Storage) migrateUserDB() error {
	queries := []string{
		// Chat messages
		`CREATE TABLE IF NOT EXISTS chats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			metadata TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_chats_session_id ON chats(session_id);`,
		`ALTER TABLE chats ADD COLUMN metadata TEXT DEFAULT '';`,

		// Sessions
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL DEFAULT '',
			archived BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);`,

		// Tasks
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'todo',
			result TEXT NOT NULL DEFAULT '',
			agent TEXT NOT NULL DEFAULT '',
			archived BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`ALTER TABLE tasks ADD COLUMN agent TEXT NOT NULL DEFAULT '';`,

		// Task logs
		`CREATE TABLE IF NOT EXISTS task_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			event TEXT NOT NULL,
			message TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_task_logs_task_id ON task_logs(task_id);`,

		// Prompts
		`CREATE TABLE IF NOT EXISTS prompts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,

		// Metrics
		`CREATE TABLE IF NOT EXISTS metrics_counters (
			key   TEXT PRIMARY KEY,
			value INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS metrics_events (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			payload    TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_events_id ON metrics_events(id DESC);`,
		`CREATE TABLE IF NOT EXISTS metrics_breakdowns (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS pairing_store (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			channel     TEXT    NOT NULL,
			sender_id   TEXT    NOT NULL,
			code        TEXT    NOT NULL,
			pending     BOOLEAN NOT NULL DEFAULT 1,
			created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			approved_at DATETIME,
			UNIQUE(channel, sender_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_pairing_code ON pairing_store(code);`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			errLower := strings.ToLower(err.Error())
			// Skip expected errors for idempotent migrations
			if strings.Contains(errLower, "duplicate column name") || strings.Contains(errLower, "no such column") {
				// Expected for idempotent migrations - log at debug level
				continue
			}
			if strings.HasPrefix(query, "ALTER TABLE") && !strings.Contains(err.Error(), "syntax error") {
				// Log non-critical ALTER TABLE failures for debugging schema drift
				queryPreview := query
				if len(queryPreview) > 80 {
					queryPreview = queryPreview[:80] + "..."
				}
				logger.WarnCF("storage", "ALTER TABLE migration skipped", map[string]interface{}{
					"query": queryPreview,
					"error": err.Error(),
				})
				continue
			}
			return fmt.Errorf("executing user migration: %w", err)
		}
	}

	if err := s.migrateKnowledge(); err != nil {
		return fmt.Errorf("knowledge migration: %w", err)
	}
	if err := s.migrateWorkflows(); err != nil {
		return fmt.Errorf("workflow migration: %w", err)
	}
	if err := s.migrateUserProviders(); err != nil {
		return fmt.Errorf("user providers migration: %w", err)
	}
	if err := s.migrateSkillUsage(); err != nil {
		return fmt.Errorf("skill usage migration: %w", err)
	}
	if err := s.migratePackages(); err != nil {
		return fmt.Errorf("packages migration: %w", err)
	}
	if err := s.migrateMarketingAudience(); err != nil {
		return fmt.Errorf("marketing audience migration: %w", err)
	}

	return nil
}

func (s *Storage) migrateMarketingAudience() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS contacts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			first_name TEXT DEFAULT '',
			last_name TEXT DEFAULT '',
			phone TEXT DEFAULT '',
			company TEXT DEFAULT '',
			title TEXT DEFAULT '',
			tags TEXT DEFAULT '[]',
			custom_fields TEXT DEFAULT '{}',
			status TEXT DEFAULT 'active',
			source TEXT DEFAULT 'manual',
			subscribed_at DATETIME,
			unsubscribed_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS contact_lists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT DEFAULT '',
			type TEXT DEFAULT 'static',
			account TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS contact_list_members (
			contact_id INTEGER NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
			list_id INTEGER NOT NULL REFERENCES contact_lists(id) ON DELETE CASCADE,
			added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (contact_id, list_id)
		);`,
		`CREATE TABLE IF NOT EXISTS segments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT DEFAULT '',
			rules TEXT DEFAULT '[]',
			account TEXT DEFAULT '',
			contact_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS email_deliveries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			campaign_account TEXT DEFAULT '',
			campaign_name TEXT DEFAULT '',
			contact_id INTEGER REFERENCES contacts(id) ON DELETE SET NULL,
			subject TEXT DEFAULT '',
			status TEXT DEFAULT 'queued',
			provider_message_id TEXT DEFAULT '',
			sent_at DATETIME,
			delivered_at DATETIME,
			opened_at DATETIME,
			clicked_at DATETIME,
			bounced_at DATETIME,
			error TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("marketing audience migration query: %w", err)
		}
	}

	// Performance indexes for frequently filtered columns
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_contacts_status ON contacts(status)`,
		`CREATE INDEX IF NOT EXISTS idx_contacts_created_at ON contacts(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_contact_lists_account ON contact_lists(account)`,
		`CREATE INDEX IF NOT EXISTS idx_segments_account ON segments(account)`,
		`CREATE INDEX IF NOT EXISTS idx_email_deliveries_campaign ON email_deliveries(campaign_account, campaign_name)`,
		`CREATE INDEX IF NOT EXISTS idx_email_deliveries_status ON email_deliveries(status)`,
		`CREATE INDEX IF NOT EXISTS idx_email_deliveries_contact ON email_deliveries(contact_id)`,
	}
	for _, idx := range indexes {
		if _, err := s.db.Exec(idx); err != nil {
			return fmt.Errorf("creating audience index: %w", err)
		}
	}

	return nil
}

func escapeLikeQuery(query string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(query)
}

func (s *Storage) Close() error {
	_, _ = s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return s.db.Close()
}

// Pairing returns a PairingStore backed by this storage's DB.
func (s *Storage) Pairing() *PairingStore {
	return NewPairingStore(s.db)
}

func (s *Storage) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS chats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			metadata TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_chats_session_id ON chats(session_id);`,
		`ALTER TABLE chats ADD COLUMN metadata TEXT DEFAULT '';`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'todo',
			result TEXT NOT NULL DEFAULT '',
			agent TEXT NOT NULL DEFAULT '',
			archived BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`ALTER TABLE tasks ADD COLUMN archived BOOLEAN DEFAULT 0;`,
		`ALTER TABLE tasks ADD COLUMN result TEXT;`,
		`ALTER TABLE tasks ADD COLUMN agent TEXT NOT NULL DEFAULT '';`,
		`UPDATE tasks SET description = '' WHERE description IS NULL;`,
		`UPDATE tasks SET result = '' WHERE result IS NULL;`,
		`UPDATE tasks SET status = 'todo' WHERE status IS NULL;`,
		`UPDATE tasks SET archived = 0 WHERE archived IS NULL;`,
		`CREATE TABLE IF NOT EXISTS task_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			event TEXT NOT NULL,
			message TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_task_logs_task_id ON task_logs(task_id);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL DEFAULT '',
			archived BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`ALTER TABLE sessions ADD COLUMN title TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE sessions ADD COLUMN archived BOOLEAN NOT NULL DEFAULT 0;`,
		`ALTER TABLE sessions ADD COLUMN created_at DATETIME DEFAULT CURRENT_TIMESTAMP;`,
		`ALTER TABLE sessions ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP;`,
		`UPDATE sessions SET title = '' WHERE title IS NULL;`,
		`UPDATE sessions SET archived = 0 WHERE archived IS NULL;`,
		`UPDATE sessions SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;`,
		`UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			extended_thinking BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`ALTER TABLE users ADD COLUMN uuid TEXT;`,
		`UPDATE users SET uuid = lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)),2) || '-a' || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6))) WHERE uuid IS NULL OR uuid = '';`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_uuid ON users(uuid);`,
		`ALTER TABLE users ADD COLUMN email TEXT;`,
		`ALTER TABLE users ADD COLUMN full_name TEXT;`,
		`ALTER TABLE users ADD COLUMN date_of_birth DATE;`,
		`ALTER TABLE users ADD COLUMN timezone TEXT;`,
		`ALTER TABLE users ADD COLUMN preferred_language TEXT;`,
		`ALTER TABLE users ADD COLUMN avatar_url TEXT;`,
		`ALTER TABLE users ADD COLUMN blocked BOOLEAN NOT NULL DEFAULT 0;`,
		`ALTER TABLE users ADD COLUMN blocked_reason TEXT;`,
		`ALTER TABLE users ADD COLUMN blocked_by INTEGER;`,
		`ALTER TABLE users ADD COLUMN blocked_at DATETIME;`,
		`CREATE INDEX IF NOT EXISTS idx_users_blocked ON users(blocked);`,
		`ALTER TABLE users ADD COLUMN allowed_tools TEXT;`,
		`ALTER TABLE users ADD COLUMN extended_thinking BOOLEAN NOT NULL DEFAULT 0;`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
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
		`ALTER TABLE chats ADD COLUMN user_id INTEGER DEFAULT 1;`,
		`ALTER TABLE tasks ADD COLUMN user_id INTEGER DEFAULT 1;`,
		`ALTER TABLE sessions ADD COLUMN user_id INTEGER DEFAULT 1;`,
		`ALTER TABLE knowledge_documents ADD COLUMN user_id INTEGER DEFAULT 1;`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate column name") || strings.Contains(strings.ToLower(err.Error()), "no such column") {
				continue
			}
			if strings.HasPrefix(query, "ALTER TABLE") && !strings.Contains(err.Error(), "syntax error") {
				continue
			}
			return fmt.Errorf("executing migration query: %w", err)
		}
	}

	if err := s.migrateKnowledge(); err != nil {
		return fmt.Errorf("knowledge migration: %w", err)
	}
	if err := s.migrateWorkflows(); err != nil {
		return fmt.Errorf("workflow migration: %w", err)
	}
	if err := s.migratePrompts(); err != nil {
		return fmt.Errorf("prompts migration: %w", err)
	}
	if err := s.migrateSessions(); err != nil {
		return fmt.Errorf("session migration: %w", err)
	}
	if err := s.migrateMetrics(); err != nil {
		return fmt.Errorf("metrics migration: %w", err)
	}
	if err := s.migrateSetupSessions(); err != nil {
		return fmt.Errorf("setup sessions migration: %w", err)
	}
	if err := s.migrateUserProviders(); err != nil {
		return fmt.Errorf("user providers migration: %w", err)
	}
	if err := s.migrateSkillUsage(); err != nil {
		return fmt.Errorf("skill usage migration: %w", err)
	}
	if err := s.migratePackages(); err != nil {
		return fmt.Errorf("packages migration: %w", err)
	}

	return nil
}

func (s *Storage) migrateMetrics() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS metrics_counters (
			key   TEXT PRIMARY KEY,
			value INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS metrics_events (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			payload    TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_events_id ON metrics_events(id DESC);`,
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

func (s *Storage) migrateSetupSessions() error {
	queries := []string{
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
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			if strings.HasPrefix(q, "ALTER TABLE") {
				continue
			}
			return fmt.Errorf("setup sessions migration query: %w", err)
		}
	}
	return nil
}

func normalizeUserKey(userKey interface{}) interface{} {
	switch v := userKey.(type) {
	case int64:
		if v <= 0 {
			return 1
		}
		return v
	case string:
		var id int64
		if _, err := fmt.Sscanf(v, "%d", &id); err == nil {
			if id <= 0 {
				return 1
			}
			return id
		}
		return v
	default:
		return v
	}
}

func (s *Storage) ExecRaw(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func (s *Storage) QueryRaw(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(query, args...)
}

// CountUsers returns the total number of users in the shared users table.
// In legacy shared-DB mode this is used to enforce the SINGLE-USER ONLY
// contract for components that are not safe once multiple accounts exist.
func (s *Storage) CountUsers() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (s *Storage) migrateSkillUsage() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS skill_usage_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			skill_name TEXT NOT NULL,
			skill_source TEXT NOT NULL,
			session_key TEXT NOT NULL,
			loaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_skill_usage_skill_name ON skill_usage_events(skill_name);`,
		`CREATE INDEX IF NOT EXISTS idx_skill_usage_loaded_at ON skill_usage_events(loaded_at);`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("skill usage migration query: %w", err)
		}
	}
	return nil
}

// RecordSkillUsage inserts a single usage event (fire-and-forget safe).
func (s *Storage) RecordSkillUsage(skillName, skillSource, sessionKey string) error {
	_, err := s.db.Exec(`
		INSERT INTO skill_usage_events (skill_name, skill_source, session_key)
		VALUES (?, ?, ?)`, skillName, skillSource, sessionKey)
	return err
}

// SkillUsageStat holds aggregated usage stats for a single skill.
type SkillUsageStat struct {
	SkillName  string `json:"skill_name"`
	LoadCount  int    `json:"load_count"`
	LastUsedAt string `json:"last_used_at"`
}

// GetSkillUsageStats returns top skills by load count over the past `days` days.
func (s *Storage) GetSkillUsageStats(days int) ([]SkillUsageStat, error) {
	rows, err := s.db.Query(`
		SELECT skill_name, COUNT(*) as load_count, MAX(loaded_at) as last_used_at
		FROM skill_usage_events
		WHERE loaded_at >= datetime('now', ?)
		GROUP BY skill_name
		ORDER BY load_count DESC
		LIMIT 20`,
		fmt.Sprintf("-%d days", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats []SkillUsageStat
	for rows.Next() {
		var st SkillUsageStat
		if err := rows.Scan(&st.SkillName, &st.LoadCount, &st.LastUsedAt); err != nil {
			return nil, err
		}
		stats = append(stats, st)
	}
	if stats == nil {
		stats = []SkillUsageStat{}
	}
	return stats, rows.Err()
}
