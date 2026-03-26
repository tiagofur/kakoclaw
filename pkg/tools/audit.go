package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/storage"
)

// ToolExecutionLog represents a single tool execution audit entry
type ToolExecutionLog struct {
	ID        int64                  `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    int64                  `json:"user_id"`
	Username  string                 `json:"username"`
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"` // Sanitized - no passwords/tokens
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Duration  int64                  `json:"duration_ms"` // Execution duration in milliseconds
}

// AuditLogger defines the interface for logging tool executions
type AuditLogger interface {
	LogToolExecution(ctx context.Context, log ToolExecutionLog) error
	QueryLogs(ctx context.Context, filters AuditQueryFilters) ([]ToolExecutionLog, error)
}

// AuditQueryFilters defines query parameters for audit log retrieval
type AuditQueryFilters struct {
	UserID    *int64     // Filter by user ID
	Tool      string     // Filter by tool name
	StartTime *time.Time // Filter by start time
	EndTime   *time.Time // Filter by end time
	Limit     int        // Maximum number of results
	Offset    int        // Pagination offset
}

// SQLiteAuditLogger implements AuditLogger using SQLite storage
type SQLiteAuditLogger struct {
	storage *storage.Storage
}

// NewSQLiteAuditLogger creates a new SQLite-based audit logger
func NewSQLiteAuditLogger(store *storage.Storage) (*SQLiteAuditLogger, error) {
	if store == nil {
		return nil, fmt.Errorf("storage cannot be nil")
	}

	// Ensure audit log table exists
	if _, err := store.ExecRaw(createAuditTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create audit_log table: %w", err)
	}

	return &SQLiteAuditLogger{storage: store}, nil
}

const createAuditTableSQL = `
CREATE TABLE IF NOT EXISTS tool_audit_log (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	user_id INTEGER NOT NULL,
	username TEXT NOT NULL,
	tool TEXT NOT NULL,
	arguments TEXT NOT NULL,
	success BOOLEAN NOT NULL,
	error TEXT,
	duration_ms INTEGER NOT NULL,
	FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_audit_user_id ON tool_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_tool ON tool_audit_log(tool);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON tool_audit_log(timestamp);
`

// LogToolExecution logs a tool execution to the audit trail
func (a *SQLiteAuditLogger) LogToolExecution(ctx context.Context, log ToolExecutionLog) error {
	// Sanitize arguments before storing
	sanitized := sanitizeArguments(log.Arguments)
	argsJSON, err := json.Marshal(sanitized)
	if err != nil {
		return fmt.Errorf("failed to marshal arguments: %w", err)
	}

	query := `
		INSERT INTO tool_audit_log (timestamp, user_id, username, tool, arguments, success, error, duration_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = a.storage.ExecRaw(query,
		log.Timestamp,
		log.UserID,
		log.Username,
		log.Tool,
		string(argsJSON),
		log.Success,
		log.Error,
		log.Duration,
	)

	return err
}

// QueryLogs retrieves audit logs matching the specified filters
func (a *SQLiteAuditLogger) QueryLogs(ctx context.Context, filters AuditQueryFilters) ([]ToolExecutionLog, error) {
	query := `SELECT id, timestamp, user_id, username, tool, arguments, success, error, duration_ms FROM tool_audit_log WHERE 1=1`
	args := []interface{}{}

	if filters.UserID != nil {
		query += ` AND user_id = ?`
		args = append(args, *filters.UserID)
	}

	if filters.Tool != "" {
		query += ` AND tool = ?`
		args = append(args, filters.Tool)
	}

	if filters.StartTime != nil {
		query += ` AND timestamp >= ?`
		args = append(args, *filters.StartTime)
	}

	if filters.EndTime != nil {
		query += ` AND timestamp <= ?`
		args = append(args, *filters.EndTime)
	}

	query += ` ORDER BY timestamp DESC`

	if filters.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filters.Limit)
	} else {
		query += ` LIMIT 1000` // Default limit
	}

	if filters.Offset > 0 {
		query += ` OFFSET ?`
		args = append(args, filters.Offset)
	}

	rows, err := a.storage.QueryRaw(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []ToolExecutionLog{}
	for rows.Next() {
		var log ToolExecutionLog
		var argsJSON string
		var errorStr sql.NullString

		err := rows.Scan(&log.ID, &log.Timestamp, &log.UserID, &log.Username,
			&log.Tool, &argsJSON, &log.Success, &errorStr, &log.Duration)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(argsJSON), &log.Arguments); err != nil {
			log.Arguments = map[string]interface{}{"_parse_error": err.Error()}
		}

		if errorStr.Valid {
			log.Error = errorStr.String
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// sanitizeArguments removes sensitive fields from tool arguments before logging
func sanitizeArguments(args map[string]interface{}) map[string]interface{} {
	if args == nil {
		return map[string]interface{}{}
	}

	// List of sensitive field names to redact
	sensitiveKeys := []string{
		"password", "token", "secret", "api_key", "apikey",
		"access_token", "refresh_token", "private_key",
	}

	sanitized := make(map[string]interface{}, len(args))
	for key, value := range args {
		keyLower := strings.ToLower(key)
		isSensitive := false

		for _, sensitive := range sensitiveKeys {
			if strings.Contains(keyLower, sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized[key] = "[REDACTED]"
		} else {
			sanitized[key] = value
		}
	}

	// Special handling for configure tool: check if "path" indicates a sensitive field
	// and redact "value" accordingly
	if path, ok := args["path"].(string); ok {
		if _, hasValue := args["value"]; hasValue {
			if isSensitiveConfigPath(path) {
				sanitized["value"] = "[REDACTED]"
			}
		}
	}

	return sanitized
}

// isSensitiveConfigPath checks if a config path points to a sensitive field.
// Used by sanitizeArguments to redact values in configure tool audit logs.
func isSensitiveConfigPath(path string) bool {
	// Extract the last segment of the path (the field name)
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return false
	}
	fieldName := parts[len(parts)-1]

	// List of sensitive field name patterns
	sensitivePatterns := []string{
		"api_key", "apikey", "token", "password", "secret",
		"private_key", "encrypt_key", "bot_token", "app_token",
		"app_secret", "client_secret", "access_token", "refresh_token",
		"verification_token",
	}

	nameLower := strings.ToLower(fieldName)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(nameLower, pattern) {
			return true
		}
	}
	return false
}

// RestrictedTools defines which tools should be audited (high-risk tools)
var RestrictedTools = map[string]bool{
	"exec":        true,
	"spawn":       true,
	"email":       true,
	"write_file":  true,
	"edit_file":   true,
	"append_file": true,
	"web_fetch":   true,
	"configure":    true,
	"system_setup": true,
}

// IsRestrictedTool checks if a tool should be audited
func IsRestrictedTool(toolName string) bool {
	return RestrictedTools[toolName]
}
