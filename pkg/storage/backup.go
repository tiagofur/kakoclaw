package storage

import (
	"fmt"
	"time"
)

// ==================== BACKUP DATA TYPES ====================

// BackupUserData holds all exportable data for a single user.
type BackupUserData struct {
	ExportedAt time.Time       `json:"exported_at"`
	UserID     int64           `json:"user_id"`
	Sessions   []BackupSession `json:"sessions"`
	Messages   []BackupMessage `json:"messages"`
	Tasks      []BackupTask    `json:"tasks"`
	TaskLogs   []BackupTaskLog `json:"task_logs,omitempty"`
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
	uid := normalizeUserID(userID)
	data := &BackupUserData{
		ExportedAt: time.Now().UTC(),
		UserID:     uid,
		Sessions:   make([]BackupSession, 0),
		Messages:   make([]BackupMessage, 0),
		Tasks:      make([]BackupTask, 0),
		TaskLogs:   make([]BackupTaskLog, 0),
	}

	// Export sessions
	rows, err := s.db.Query(`SELECT session_id, COALESCE(title,''), archived, created_at, updated_at FROM sessions WHERE user_id = ? ORDER BY created_at ASC`, uid)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sess BackupSession
			if err := rows.Scan(&sess.SessionID, &sess.Title, &sess.Archived, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
				continue
			}
			data.Sessions = append(data.Sessions, sess)
		}
	}

	// Export messages
	msgRows, err := s.db.Query(`SELECT session_id, role, content, created_at FROM chats WHERE user_id = ? ORDER BY created_at ASC`, uid)
	if err == nil {
		defer msgRows.Close()
		for msgRows.Next() {
			var msg BackupMessage
			if err := msgRows.Scan(&msg.SessionID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
				continue
			}
			data.Messages = append(data.Messages, msg)
		}
	}

	// Export tasks
	taskRows, err := s.db.Query(`SELECT id, title, COALESCE(description,''), status, COALESCE(result,''), archived, created_at, updated_at FROM tasks WHERE user_id = ? ORDER BY created_at ASC`, uid)
	if err == nil {
		defer taskRows.Close()
		taskIDs := make(map[int64]string) // map task DB id -> title for log linking
		for taskRows.Next() {
			var t BackupTask
			var dbID int64
			if err := taskRows.Scan(&dbID, &t.Title, &t.Description, &t.Status, &t.Result, &t.Archived, &t.CreatedAt, &t.UpdatedAt); err != nil {
				continue
			}
			data.Tasks = append(data.Tasks, t)
			taskIDs[dbID] = t.Title
		}

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

	return data, nil
}

// ==================== IMPORT ====================

// ImportUserData inserts backup data into the database for the given user.
// It uses INSERT OR IGNORE for sessions to avoid duplicates, and always appends messages/tasks.
// Returns counts of imported items.
func (s *Storage) ImportUserData(userID int64, data *BackupUserData) (sessions, messages, tasks int, err error) {
	if data == nil {
		return 0, 0, 0, nil
	}
	uid := normalizeUserID(userID)

	tx, errTx := s.db.Begin()
	if errTx != nil {
		return 0, 0, 0, fmt.Errorf("begin tx: %w", errTx)
	}
	defer tx.Rollback()

	// Import sessions (skip duplicates)
	stmtSess, err := tx.Prepare(`INSERT OR IGNORE INTO sessions (session_id, user_id, title, archived, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("prepare sessions: %w", err)
	}
	defer stmtSess.Close()
	for _, sess := range data.Sessions {
		res, err := stmtSess.Exec(sess.SessionID, uid, sess.Title, sess.Archived, sess.CreatedAt, sess.UpdatedAt)
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
	existRows, err := tx.Query(`SELECT session_id, role, created_at FROM chats WHERE user_id = ?`, uid)
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

	stmtMsg, err := tx.Prepare(`INSERT INTO chats (session_id, user_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return sessions, 0, 0, fmt.Errorf("prepare messages: %w", err)
	}
	defer stmtMsg.Close()
	for _, msg := range data.Messages {
		key := fmt.Sprintf("%s|%s|%d", msg.SessionID, msg.Role, msg.CreatedAt.Unix())
		if existingMsgs[key] {
			continue // skip duplicate
		}
		// Ensure session exists
		stmtSess.Exec(msg.SessionID, uid, "", false, msg.CreatedAt, msg.CreatedAt)
		if _, err := stmtMsg.Exec(msg.SessionID, uid, msg.Role, msg.Content, msg.CreatedAt); err != nil {
			continue
		}
		messages++
	}

	// Import tasks
	stmtTask, err := tx.Prepare(`INSERT INTO tasks (user_id, title, description, status, result, archived, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return sessions, messages, 0, fmt.Errorf("prepare tasks: %w", err)
	}
	defer stmtTask.Close()
	taskTitleToID := make(map[string]int64) // for linking logs
	for _, t := range data.Tasks {
		res, err := stmtTask.Exec(uid, t.Title, t.Description, t.Status, t.Result, t.Archived, t.CreatedAt, t.UpdatedAt)
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

	if err := tx.Commit(); err != nil {
		return 0, 0, 0, fmt.Errorf("commit: %w", err)
	}

	return sessions, messages, tasks, nil
}
