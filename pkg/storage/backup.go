package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ==================== BACKUP DATA TYPES ====================

// BackupUserData holds all exportable data for a single user.
type BackupUserData struct {
	ExportedAt time.Time              `json:"exported_at"`
	UserID     int64                  `json:"user_id"`
	Sessions   []BackupSession        `json:"sessions"`
	Messages   []BackupMessage        `json:"messages"`
	Tasks      []BackupTask           `json:"tasks"`
	TaskLogs   []BackupTaskLog        `json:"task_logs,omitempty"`
	Providers  *UserProvidersConfig   `json:"providers,omitempty"`
	Channels   []BackupChannelMapping `json:"channels,omitempty"`
}

// BackupChannelMapping is a portable channel-to-user mapping record.
type BackupChannelMapping struct {
	Channel  string `json:"channel"`
	SenderID string `json:"sender_id"`
}

// BackupSession is a portable session record (no internal DB id).
type BackupSession struct {
	SessionID string    `json:"session_id"`
	Title     string    `json:"title"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BackupMessage is a portable chat message (no internal DB id).
type BackupMessage struct {
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// BackupTask is a portable task record (no internal DB id).
type BackupTask struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Result      string    `json:"result"`
	Archived    bool      `json:"archived"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BackupTaskLog is a portable task log (references task by index in Tasks slice).
type BackupTaskLog struct {
	TaskTitle string    `json:"task_title"` // reference by title since IDs are not portable
	Event     string    `json:"event"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// ==================== EXPORT ====================

// ExportUserData extracts all data belonging to a user for backup purposes.
func (s *Storage) ExportUserData(userID int64) (*BackupUserData, error) {
	// Use userID as-is — callers that pass 0 mean "owner of this isolated per-user DB".
	// Do NOT normalise 0 → 1 here, because per-user DBs use 0 as the owner key.
	uid := userID
	data := &BackupUserData{
		ExportedAt: time.Now().UTC(),
		UserID:     uid,
		Sessions:   make([]BackupSession, 0),
		Messages:   make([]BackupMessage, 0),
		Tasks:      make([]BackupTask, 0),
		TaskLogs:   make([]BackupTaskLog, 0),
		Channels:   make([]BackupChannelMapping, 0),
	}

	// Detect if tables have user_id columns (for multi-user isolation support)
	hasSessUID := s.hasColumn("sessions", "user_id")
	hasChatUID := s.hasColumn("chats", "user_id")
	hasTaskUID := s.hasColumn("tasks", "user_id")
	hasChanTable := s.hasTable("channel_users")
	hasChanUID := hasChanTable && s.hasColumn("channel_users", "user_id")

	// Export sessions
	querySess := `SELECT session_id, COALESCE(title,''), archived, created_at, updated_at FROM sessions`
	if hasSessUID {
		querySess += ` WHERE user_id = ?`
	}
	querySess += ` ORDER BY created_at ASC`

	var rows *sql.Rows
	var err error
	if hasSessUID {
		rows, err = s.db.Query(querySess, uid)
	} else {
		rows, err = s.db.Query(querySess)
	}

	if err == nil {
		for rows.Next() {
			var sess BackupSession
			if err := rows.Scan(&sess.SessionID, &sess.Title, &sess.Archived, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
				continue
			}
			data.Sessions = append(data.Sessions, sess)
		}
		rows.Close()
	}

	// Export messages
	queryChat := `SELECT session_id, role, content, created_at FROM chats`
	if hasChatUID {
		queryChat += ` WHERE user_id = ?`
	}
	queryChat += ` ORDER BY created_at ASC`

	if hasChatUID {
		rows, err = s.db.Query(queryChat, uid)
	} else {
		rows, err = s.db.Query(queryChat)
	}

	if err == nil {
		for rows.Next() { // Note: variable was msgRows in original, changing to rows for consistency
			var msg BackupMessage
			if err := rows.Scan(&msg.SessionID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
				continue
			}
			data.Messages = append(data.Messages, msg)
		}
		rows.Close()
	}

	// Export tasks
	queryTask := `SELECT id, title, COALESCE(description,''), status, COALESCE(result,''), archived, created_at, updated_at FROM tasks`
	if hasTaskUID {
		queryTask += ` WHERE user_id = ?`
	}
	queryTask += ` ORDER BY created_at ASC`

	if hasTaskUID {
		rows, err = s.db.Query(queryTask, uid)
	} else {
		rows, err = s.db.Query(queryTask)
	}

	if err == nil {
		taskIDs := make(map[int64]string) // map task DB id -> title for log linking
		for rows.Next() {
			var t BackupTask
			var dbID int64
			if err := rows.Scan(&dbID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.CreatedAt, &t.UpdatedAt); err != nil {
				continue
			}
			data.Tasks = append(data.Tasks, t)
			taskIDs[dbID] = t.Title
		}
		rows.Close()

		// Export task logs for user's tasks
		if len(taskIDs) > 0 {
			ids := make([]int64, 0, len(taskIDs))
			for id := range taskIDs {
				ids = append(ids, id)
			}
			placeholders := ""
			args := make([]interface{}, len(ids))
			for i, id := range ids {
				if i > 0 {
					placeholders += ","
				}
				placeholders += "?"
				args[i] = id
			}
			logRows, err := s.db.Query(
				fmt.Sprintf(`SELECT task_id, event, message, created_at FROM task_logs WHERE task_id IN (%s) ORDER BY created_at ASC`, placeholders),
				args...,
			)
			if err == nil {
				defer logRows.Close()
				for logRows.Next() {
					var taskID int64
					var log BackupTaskLog
					if err := logRows.Scan(&taskID, &log.Event, &log.Message, &log.CreatedAt); err != nil {
						continue
					}
					log.TaskTitle = taskIDs[taskID]
					data.TaskLogs = append(data.TaskLogs, log)
				}
			}
		}
	}

	// Export per-user provider config
	if providers, err := s.GetUserProvidersConfig(uid); err == nil && providers != nil {
		data.Providers = providers
	}

	// Export channel mappings that point to this user (only if table exists)
	if hasChanTable {
		queryChan := `SELECT channel, sender_id FROM channel_users`
		var chanRows *sql.Rows
		if hasChanUID {
			chanRows, err = s.db.Query(queryChan+` WHERE user_id = ? ORDER BY channel, sender_id`, uid)
		} else {
			chanRows, err = s.db.Query(queryChan + ` ORDER BY channel, sender_id`)
		}

		if err == nil {
			defer chanRows.Close()
			for chanRows.Next() {
				var channel, senderID string
				if err := chanRows.Scan(&channel, &senderID); err != nil {
					continue
				}
				data.Channels = append(data.Channels, BackupChannelMapping{
					Channel:  channel,
					SenderID: senderID,
				})
			}
		}
	}

	return data, nil
}

// ==================== IMPORT ====================

// ImportUserData inserts backup data into the database for the given user.
// It uses INSERT OR IGNORE for sessions to avoid duplicates, and always appends messages/tasks.
// Returns counts of imported items.
// ImportUserData inserts backup data into the database for the given user.
// It uses INSERT OR IGNORE for sessions to avoid duplicates, and always appends messages/tasks.
// Returns counts of imported items.
func (s *Storage) ImportUserData(userID int64, data *BackupUserData) (sessions, messages, tasks int, err error) {
	if data == nil {
		return 0, 0, 0, nil
	}
	// Use userID as-is — callers that pass 0 mean "owner of this isolated per-user DB".
	// Do NOT normalise 0 → 1 here, because per-user DBs use 0 as the owner key.
	uid := userID

	// Detect table properties BEFORE opening the transaction to avoid deadlock with MaxOpenConns=1
	hasSessUID := s.hasColumn("sessions", "user_id")
	hasChatUID := s.hasColumn("chats", "user_id")
	hasTaskUID := s.hasColumn("tasks", "user_id")
	hasChanTable := s.hasTable("channel_users")
	hasChanUID := hasChanTable && s.hasColumn("channel_users", "user_id")

	tx, errTx := s.db.Begin()
	if errTx != nil {
		return 0, 0, 0, fmt.Errorf("begin tx: %w", errTx)
	}
	defer tx.Rollback()

	// Import sessions (skip duplicates)
	querySess := `INSERT OR IGNORE INTO sessions (session_id`
	if hasSessUID {
		querySess += `, user_id`
	}
	querySess += `, title, archived, created_at, updated_at) VALUES (?, `
	if hasSessUID {
		querySess += `?, `
	}
	querySess += `?, ?, ?, ?)`

	stmtSess, err := tx.Prepare(querySess)
	if err != nil {
		fmt.Printf("DEBUG: ImportUserData prepare sess FAILED: %v\n", err)
		return 0, 0, 0, fmt.Errorf("prepare sessions: %w", err)
	}
	defer stmtSess.Close()

	for _, sess := range data.Sessions {
		var args []interface{}
		args = append(args, sess.SessionID)
		if hasSessUID {
			args = append(args, uid)
		}
		args = append(args, sess.Title, sess.Archived, sess.CreatedAt, sess.UpdatedAt)

		res, err := stmtSess.Exec(args...)
		if err != nil {
			continue
		}
		if n, _ := res.RowsAffected(); n > 0 {
			sessions++
		}
	}

	// Import messages
	// First, collect existing message fingerprints to avoid exact duplicates
	existingMsgs := make(map[string]bool)
	queryExist := `SELECT session_id, role, created_at FROM chats`
	if hasChatUID {
		queryExist += ` WHERE user_id = ?`
	}

	var existRows *sql.Rows
	if hasChatUID {
		existRows, err = tx.Query(queryExist, uid)
	} else {
		existRows, err = tx.Query(queryExist)
	}

	if err == nil {
		for existRows.Next() {
			var sid, role string
			var cat time.Time
			if existRows.Scan(&sid, &role, &cat) == nil {
				key := fmt.Sprintf("%s|%s|%d", sid, role, cat.Unix())
				existingMsgs[key] = true
			}
		}
		existRows.Close()
	}

	queryMsg := `INSERT INTO chats (session_id`
	if hasChatUID {
		queryMsg += `, user_id`
	}
	queryMsg += `, role, content, created_at) VALUES (?, `
	if hasChatUID {
		queryMsg += `?, `
	}
	queryMsg += `?, ?, ?)`

	stmtMsg, err := tx.Prepare(queryMsg)
	if err != nil {
		fmt.Printf("DEBUG: ImportUserData prepare msg FAILED: %v\n", err)
		return sessions, 0, 0, fmt.Errorf("prepare messages: %w", err)
	}
	defer stmtMsg.Close()

	for _, msg := range data.Messages {
		key := fmt.Sprintf("%s|%s|%d", msg.SessionID, msg.Role, msg.CreatedAt.Unix())
		if existingMsgs[key] {
			continue // skip duplicate
		}

		// Ensure session exists
		var sessArgs []interface{}
		sessArgs = append(sessArgs, msg.SessionID)
		if hasSessUID {
			sessArgs = append(sessArgs, uid)
		}
		sessArgs = append(sessArgs, "", false, msg.CreatedAt, msg.CreatedAt)
		stmtSess.Exec(sessArgs...)

		// Insert message
		var msgArgs []interface{}
		msgArgs = append(msgArgs, msg.SessionID)
		if hasChatUID {
			msgArgs = append(msgArgs, uid)
		}
		msgArgs = append(msgArgs, msg.Role, msg.Content, msg.CreatedAt)

		if _, err := stmtMsg.Exec(msgArgs...); err != nil {
			continue
		}
		messages++
	}

	// Import tasks
	queryTask := `INSERT INTO tasks (`
	if hasTaskUID {
		queryTask += `user_id, `
	}
	queryTask += `title, description, status, result, archived, created_at, updated_at) VALUES (`
	if hasTaskUID {
		queryTask += `?, `
	}
	queryTask += `?, ?, ?, ?, ?, ?, ?)`

	stmtTask, err := tx.Prepare(queryTask)
	if err != nil {
		fmt.Printf("DEBUG: ImportUserData prepare task FAILED: %v\n", err)
		return sessions, messages, 0, fmt.Errorf("prepare tasks: %w", err)
	}
	defer stmtTask.Close()

	taskTitleToID := make(map[string]int64) // for linking logs
	for _, t := range data.Tasks {
		var taskArgs []interface{}
		if hasTaskUID {
			taskArgs = append(taskArgs, uid)
		}
		taskArgs = append(taskArgs, t.Title, t.Description, t.Status, t.Result, t.Archived, t.CreatedAt, t.UpdatedAt)

		res, err := stmtTask.Exec(taskArgs...)
		if err != nil {
			continue
		}
		tasks++
		if id, e := res.LastInsertId(); e == nil {
			taskTitleToID[t.Title] = id
		}
	}

	// Import task logs
	if len(data.TaskLogs) > 0 {
		stmtLog, err := tx.Prepare(`INSERT INTO task_logs (task_id, event, message, created_at) VALUES (?, ?, ?, ?)`)
		if err == nil {
			defer stmtLog.Close()
			for _, log := range data.TaskLogs {
				taskID, ok := taskTitleToID[log.TaskTitle]
				if !ok {
					continue
				}
				stmtLog.Exec(taskID, log.Event, log.Message, log.CreatedAt)
			}
		}
	}

	// Import per-user provider config
	if data.Providers != nil {
		providers := *data.Providers
		providers.UserID = uid
		providersJSON, marshalErr := json.Marshal(&providers)
		if marshalErr == nil {
			// user_providers_config always has user_id as PK, even in isolated DB.
			// (Though in isolated DB there's only one row with user_id=0 or similar)
			if _, err := tx.Exec(`
				INSERT INTO user_providers_config (user_id, config) VALUES (?, ?)
				ON CONFLICT(user_id) DO UPDATE SET config = excluded.config
			`, uid, string(providersJSON)); err != nil {
				return 0, 0, 0, fmt.Errorf("import providers: %w", err)
			}
		}
	}

	// Import channel mappings remapped to target user
	if hasChanTable && len(data.Channels) > 0 {
		for _, mapping := range data.Channels {
			if strings.TrimSpace(mapping.Channel) == "" || strings.TrimSpace(mapping.SenderID) == "" {
				continue
			}

			insertChan := `INSERT INTO channel_users (channel, sender_id`
			if hasChanUID {
				insertChan += `, user_id`
			}
			insertChan += `) VALUES (?, ?`
			if hasChanUID {
				insertChan += `, ?`
			}
			insertChan += `) ON CONFLICT(channel, sender_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP`
			if hasChanUID {
				insertChan += `, user_id = excluded.user_id`
			}

			var chanArgs []interface{}
			chanArgs = append(chanArgs, mapping.Channel, mapping.SenderID)
			if hasChanUID {
				chanArgs = append(chanArgs, uid)
			}

			if _, err := tx.Exec(insertChan, chanArgs...); err != nil {
				return 0, 0, 0, fmt.Errorf("import channel mapping: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, 0, fmt.Errorf("commit: %w", err)
	}

	return sessions, messages, tasks, nil
}

// hasTable checks if a table exists in the database.
func (s *Storage) hasTable(table string) bool {
	var count int
	err := s.db.QueryRow(
		`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table,
	).Scan(&count)
	return err == nil && count > 0
}

// hasColumn checks if a table has a specific column.
func (s *Storage) hasColumn(table, column string) bool {
	query := fmt.Sprintf("PRAGMA table_info(%s)", table)
	rows, err := s.db.Query(query)
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dtype string
		var notnull, pk int
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &dfltValue, &pk); err != nil {
			continue
		}
		if strings.EqualFold(name, column) {
			return true
		}
	}
	return false
}
