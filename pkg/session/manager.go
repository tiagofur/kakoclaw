package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/makoclaw/pkg/providers"
)

type Session struct {
	Key      string              `json:"key"`
	Messages []providers.Message `json:"messages"`
	Summary  string              `json:"summary,omitempty"`
	Created  time.Time           `json:"created"`
	Updated  time.Time           `json:"updated"`
	// AgentProfile tracks which agent/specialist handled this session
	AgentProfile string `json:"agent_profile,omitempty"`
	// ParentSessionKey links to parent session for hierarchical sessions (e.g., orchestrator delegations)
	ParentSessionKey string `json:"parent_session_key,omitempty"`
	// SpecialistMetadata stores stats about specialist execution
	SpecialistMetadata map[string]interface{} `json:"specialist_metadata,omitempty"`
}

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	storage  string
}

func NewSessionManager(storage string) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		storage:  storage,
	}

	if storage != "" {
		if err := os.MkdirAll(storage, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create session storage directory %s: %v\n", storage, err)
		}
		sm.loadSessions()
	}

	return sm
}

// SetStorage resets the session storage directory and reloads sessions from disk.
func (sm *SessionManager) SetStorage(storage string) {
	sm.mu.Lock()
	sm.storage = storage
	sm.sessions = make(map[string]*Session)
	sm.mu.Unlock()

	if storage != "" {
		_ = os.MkdirAll(storage, 0755)
		_ = sm.loadSessions()
	}
}

// namespaceKey creates a user-scoped session key.
// If userID is 0, returns key unchanged for backward compatibility.
func (sm *SessionManager) namespaceKey(userID int64, key string) string {
	if userID == 0 {
		return key
	}
	return fmt.Sprintf("user:%d:%s", userID, key)
}

func (sm *SessionManager) GetOrCreate(key string) *Session {
	return sm.GetOrCreateForUser(0, key)
}

// GetOrCreateForUser retrieves or creates a session for a specific user with namespaced key.
// If userID is 0, falls back to non-namespaced behavior for backward compatibility.
func (sm *SessionManager) GetOrCreateForUser(userID int64, key string) *Session {
	nsKey := sm.namespaceKey(userID, key)
	sm.mu.RLock()
	session, ok := sm.sessions[nsKey]
	sm.mu.RUnlock()

	if !ok {
		sm.mu.Lock()
		// Double-check under write lock to prevent TOCTOU race
		if existing, exists := sm.sessions[nsKey]; exists {
			sm.mu.Unlock()
			return existing
		}
		session = &Session{
			Key:      nsKey,
			Messages: []providers.Message{},
			Created:  time.Now(),
			Updated:  time.Now(),
		}
		sm.sessions[nsKey] = session
		sm.mu.Unlock()
	}

	return session
}

// GetOrCreateForSpecialist creates a session for a specialist agent
func (sm *SessionManager) GetOrCreateForSpecialist(userID int64, sessionKey, agentProfile string) *Session {
	session := sm.GetOrCreateForUser(userID, sessionKey)
	session.AgentProfile = agentProfile
	if session.SpecialistMetadata == nil {
		session.SpecialistMetadata = make(map[string]interface{})
	}
	return session
}

// LinkParentSession sets the parent session for hierarchical tracking (e.g., orchestrator delegations)
func (sm *SessionManager) LinkParentSession(userID int64, childSessionKey, parentSessionKey string) {
	session := sm.GetOrCreateForUser(userID, childSessionKey)
	session.ParentSessionKey = parentSessionKey
	sm.mu.Lock()
	sm.sessions[sm.namespaceKey(userID, childSessionKey)] = session
	sm.mu.Unlock()
}

func (sm *SessionManager) AddMessage(sessionKey, role, content string) {
	sm.AddMessageForUser(0, sessionKey, role, content)
}

// AddMessageForUser adds a simple text message for a user's session.
func (sm *SessionManager) AddMessageForUser(userID int64, sessionKey, role, content string) {
	sm.AddFullMessageForUser(userID, sessionKey, providers.Message{
		Role:    role,
		Content: content,
	})
}

// AddFullMessage adds a complete message with tool calls to the session.
func (sm *SessionManager) AddFullMessage(sessionKey string, msg providers.Message) {
	sm.AddFullMessageForUser(0, sessionKey, msg)
}

// AddFullMessageForUser adds a complete message with tool calls and tool call ID to the session for a user.
// This is used to save the full conversation flow including tool calls and tool results.
func (sm *SessionManager) AddFullMessageForUser(userID int64, sessionKey string, msg providers.Message) {
	nsKey := sm.namespaceKey(userID, sessionKey)
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[nsKey]
	if !ok {
		session = &Session{
			Key:      nsKey,
			Messages: []providers.Message{},
			Created:  time.Now(),
		}
		sm.sessions[nsKey] = session
	}

	session.Messages = append(session.Messages, msg)
	session.Updated = time.Now()
}

func (sm *SessionManager) GetHistory(key string) []providers.Message {
	return sm.GetHistoryForUser(0, key)
}

// GetHistoryForUser retrieves the message history for a user's session.
func (sm *SessionManager) GetHistoryForUser(userID int64, key string) []providers.Message {
	nsKey := sm.namespaceKey(userID, key)
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[nsKey]
	if !ok {
		return []providers.Message{}
	}

	history := make([]providers.Message, len(session.Messages))
	copy(history, session.Messages)
	return history
}

func (sm *SessionManager) GetSummary(key string) string {
	return sm.GetSummaryForUser(0, key)
}

// GetSummaryForUser retrieves the summary for a user's session.
func (sm *SessionManager) GetSummaryForUser(userID int64, key string) string {
	nsKey := sm.namespaceKey(userID, key)
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[nsKey]
	if !ok {
		return ""
	}
	return session.Summary
}

func (sm *SessionManager) SetSummary(key string, summary string) {
	sm.SetSummaryForUser(0, key, summary)
}

// SetSummaryForUser sets the summary for a user's session.
func (sm *SessionManager) SetSummaryForUser(userID int64, key string, summary string) {
	nsKey := sm.namespaceKey(userID, key)
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[nsKey]
	if ok {
		session.Summary = summary
		session.Updated = time.Now()
	}
}

func (sm *SessionManager) TruncateHistory(key string, keepLast int) {
	sm.TruncateHistoryForUser(0, key, keepLast)
}

// TruncateHistoryForUser removes all but the last keepLast messages from a user's session.
func (sm *SessionManager) TruncateHistoryForUser(userID int64, key string, keepLast int) {
	nsKey := sm.namespaceKey(userID, key)
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[nsKey]
	if !ok {
		return
	}

	if len(session.Messages) <= keepLast {
		return
	}

	session.Messages = session.Messages[len(session.Messages)-keepLast:]
	session.Updated = time.Now()
}

func (sm *SessionManager) Save(session *Session) error {
	return sm.SaveForUser(0, session)
}

// SaveForUser persists a user's session to disk.
// Sessions are namespaced by key and stored in the base storage directory.
func (sm *SessionManager) SaveForUser(userID int64, session *Session) error {
	if sm.storage == "" {
		return nil
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Sanitize session key to prevent path traversal
	safeKey := sanitizeSessionKey(session.Key)
	if safeKey == "" {
		return fmt.Errorf("invalid session key")
	}
	sessionPath := filepath.Join(sm.storage, safeKey+".json")

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sessionPath, data, 0644)
}

func (sm *SessionManager) loadSessions() error {
	files, err := os.ReadDir(sm.storage)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		sessionPath := filepath.Join(sm.storage, file.Name())
		data, err := os.ReadFile(sessionPath)
		if err != nil {
			continue
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		sm.sessions[session.Key] = &session
	}

	return nil
}

// sanitizeSessionKey strips path separators and traversal sequences from a session key
// to ensure it cannot escape the sessions directory.
func sanitizeSessionKey(key string) string {
	// Replace path separators with underscores (session keys use ":" as separator)
	key = strings.ReplaceAll(key, "/", "_")
	key = strings.ReplaceAll(key, "\\", "_")
	key = strings.ReplaceAll(key, "..", "_")
	key = strings.TrimSpace(key)
	if key == "" || key == "." {
		return ""
	}
	return key
}
