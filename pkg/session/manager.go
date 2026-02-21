package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sipeed/kakoclaw/pkg/providers"
)

type Session struct {
	Key      string              `json:"key"`
	Messages []providers.Message `json:"messages"`
	Summary  string              `json:"summary,omitempty"`
	Created  time.Time           `json:"created"`
	Updated  time.Time           `json:"updated"`
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
		os.MkdirAll(storage, 0755)
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

	sessionPath := filepath.Join(sm.storage, session.Key+".json")

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
