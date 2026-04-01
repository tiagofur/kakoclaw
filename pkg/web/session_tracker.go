package web

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sipeed/makoclaw/pkg/logger"
)

type SessionStats struct {
	SessionID  string  `json:"session_id"`
	TokensUsed int     `json:"tokens_used"`
	CostUSD    float64 `json:"cost_usd"`
	TokenLimit int     `json:"token_limit"`
	NumTurns   int     `json:"num_turns"`
}

type trackerStateFile struct {
	ActiveSessionID string                   `json:"active_session_id"`
	Sessions        map[string]*SessionStats `json:"sessions"`
}

type SessionTracker struct {
	mu             sync.Mutex
	sessions       map[string]*SessionStats
	activeSessions map[string]string
	stateDir       string
	maxTokens      int
}

func NewSessionTracker(stateDir string, maxTokens int) *SessionTracker {
	t := &SessionTracker{
		sessions:       make(map[string]*SessionStats),
		activeSessions: make(map[string]string),
		stateDir:       stateDir,
		maxTokens:      maxTokens,
	}
	t.loadAll()
	return t
}

func (t *SessionTracker) Record(userUUID, sessionID string, turns int, costUSD float64) {
	if strings.TrimSpace(userUUID) == "" || strings.TrimSpace(sessionID) == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	stats := t.getOrCreateLocked(userUUID, sessionID)
	stats.TokensUsed += turns
	stats.NumTurns += turns
	stats.CostUSD += costUSD
	t.activeSessions[userUUID] = sessionID
	t.persistUserLocked(userUUID)
}

func (t *SessionTracker) ShouldReset(userUUID, sessionID string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.maxTokens == 0 {
		return false
	}

	stats := t.lookupLocked(userUUID, sessionID)
	return stats != nil && stats.TokensUsed > t.maxTokens
}

func (t *SessionTracker) ZeroSession(userUUID, newSessionID string) {
	if strings.TrimSpace(userUUID) == "" || strings.TrimSpace(newSessionID) == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.sessions[t.sessionKey(userUUID, newSessionID)] = &SessionStats{
		SessionID:  newSessionID,
		TokenLimit: t.maxTokens,
	}
	t.activeSessions[userUUID] = newSessionID
	t.persistUserLocked(userUUID)
}

func (t *SessionTracker) Stats(userUUID, sessionID string) SessionStats {
	t.mu.Lock()
	defer t.mu.Unlock()

	if strings.TrimSpace(userUUID) == "" {
		return SessionStats{TokenLimit: t.maxTokens}
	}

	if strings.TrimSpace(sessionID) == "" {
		sessionID = t.activeSessions[userUUID]
	}

	stats := t.lookupLocked(userUUID, sessionID)
	if stats == nil {
		return SessionStats{SessionID: sessionID, TokenLimit: t.maxTokens}
	}

	copy := *stats
	copy.TokenLimit = t.maxTokens
	return copy
}

func (t *SessionTracker) sessionKey(userUUID, sessionID string) string {
	return userUUID + ":" + sessionID
}

func (t *SessionTracker) getOrCreateLocked(userUUID, sessionID string) *SessionStats {
	key := t.sessionKey(userUUID, sessionID)
	stats := t.sessions[key]
	if stats == nil {
		stats = &SessionStats{SessionID: sessionID, TokenLimit: t.maxTokens}
		t.sessions[key] = stats
	}
	stats.TokenLimit = t.maxTokens
	return stats
}

func (t *SessionTracker) lookupLocked(userUUID, sessionID string) *SessionStats {
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}
	stats := t.sessions[t.sessionKey(userUUID, sessionID)]
	if stats != nil {
		stats.TokenLimit = t.maxTokens
	}
	return stats
}

func (t *SessionTracker) loadAll() {
	if strings.TrimSpace(t.stateDir) == "" {
		return
	}

	entries, err := os.ReadDir(t.stateDir)
	if err != nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "-session-tokens.json") {
			continue
		}
		userUUID := strings.TrimSuffix(entry.Name(), "-session-tokens.json")
		data, err := os.ReadFile(filepath.Join(t.stateDir, entry.Name()))
		if err != nil {
			continue
		}
		var persisted trackerStateFile
		if err := json.Unmarshal(data, &persisted); err != nil {
			continue
		}
		if persisted.ActiveSessionID != "" {
			t.activeSessions[userUUID] = persisted.ActiveSessionID
		}
		for sessionID, stats := range persisted.Sessions {
			if stats == nil {
				continue
			}
			copied := *stats
			copied.SessionID = sessionID
			copied.TokenLimit = t.maxTokens
			t.sessions[t.sessionKey(userUUID, sessionID)] = &copied
		}
	}
}

func (t *SessionTracker) persistUserLocked(userUUID string) {
	if strings.TrimSpace(t.stateDir) == "" {
		return
	}
	if err := os.MkdirAll(t.stateDir, 0755); err != nil {
		logger.WarnCF("web", "failed to create session tracker state directory", map[string]interface{}{"user_uuid": userUUID, "error": err.Error()})
		return
	}

	persisted := trackerStateFile{
		ActiveSessionID: t.activeSessions[userUUID],
		Sessions:        make(map[string]*SessionStats),
	}
	prefix := userUUID + ":"
	for key, stats := range t.sessions {
		if !strings.HasPrefix(key, prefix) || stats == nil {
			continue
		}
		sessionID := strings.TrimPrefix(key, prefix)
		copied := *stats
		copied.TokenLimit = t.maxTokens
		persisted.Sessions[sessionID] = &copied
	}

	data, err := json.MarshalIndent(persisted, "", "  ")
	if err != nil {
		logger.WarnCF("web", "failed to marshal session tracker state", map[string]interface{}{"user_uuid": userUUID, "error": err.Error()})
		return
	}

	path := filepath.Join(t.stateDir, userUUID+"-session-tokens.json")
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		logger.WarnCF("web", "failed to write session tracker state", map[string]interface{}{"user_uuid": userUUID, "error": err.Error()})
		return
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		logger.WarnCF("web", "failed to persist session tracker state", map[string]interface{}{"user_uuid": userUUID, "error": err.Error()})
	}
}
