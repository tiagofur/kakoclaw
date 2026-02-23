package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// UserStorageManager manages per-user SQLite database connections.
// Each user gets their own database at {dataRoot}/users/{uuid}/user.db.
type UserStorageManager struct {
	central  *CentralStorage
	dataRoot string             // e.g., /home/makoclaw/.MakoClaw
	stores   map[string]*Storage // uuid -> per-user Storage
	mu       sync.RWMutex
}

// NewUserStorageManager creates a new manager for per-user databases.
func NewUserStorageManager(central *CentralStorage, dataRoot string) *UserStorageManager {
	return &UserStorageManager{
		central:  central,
		dataRoot: dataRoot,
		stores:   make(map[string]*Storage),
	}
}

// UserDataRoot returns the root directory for a specific user's data.
func (m *UserStorageManager) UserDataRoot(userUUID string) string {
	return filepath.Join(m.dataRoot, "users", userUUID)
}

// UserDBPath returns the path to a specific user's database file.
func (m *UserStorageManager) UserDBPath(userUUID string) string {
	return filepath.Join(m.UserDataRoot(userUUID), "user.db")
}

// UserWorkspacePath returns the path to a specific user's workspace directory.
func (m *UserStorageManager) UserWorkspacePath(userUUID string) string {
	return filepath.Join(m.UserDataRoot(userUUID), "workspace")
}

// UserConfigPath returns the path to a specific user's config file.
func (m *UserStorageManager) UserConfigPath(userUUID string) string {
	return filepath.Join(m.UserDataRoot(userUUID), "config.json")
}

// GetOrCreate opens (or creates) a per-user Storage.
// The database is created at {dataRoot}/users/{uuid}/user.db.
// The user's directory structure is also ensured.
func (m *UserStorageManager) GetOrCreate(userUUID string) (*Storage, error) {
	if userUUID == "" {
		return nil, fmt.Errorf("user UUID is required")
	}

	// Fast path: check if already open
	m.mu.RLock()
	if store, ok := m.stores[userUUID]; ok {
		m.mu.RUnlock()
		return store, nil
	}
	m.mu.RUnlock()

	// Slow path: create or open
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if store, ok := m.stores[userUUID]; ok {
		return store, nil
	}

	dbPath := m.UserDBPath(userUUID)

	// Ensure user directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("creating user data directory: %w", err)
	}

	// Open per-user database using NewUserStorage (no user_id columns)
	store, err := NewUserStorage(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening user database for %s: %w", userUUID, err)
	}

	m.stores[userUUID] = store
	return store, nil
}

// Get returns an existing per-user Storage, or nil if not yet opened.
func (m *UserStorageManager) Get(userUUID string) *Storage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stores[userUUID]
}

// CloseUser closes a specific user's storage connection.
func (m *UserStorageManager) CloseUser(userUUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if store, ok := m.stores[userUUID]; ok {
		delete(m.stores, userUUID)
		return store.Close()
	}
	return nil
}

// Close closes all open per-user storage connections.
func (m *UserStorageManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for uuid, store := range m.stores {
		store.Close()
		delete(m.stores, uuid)
	}
}

// DataRoot returns the base data root path.
func (m *UserStorageManager) DataRoot() string {
	return m.dataRoot
}

// Central returns the central storage reference.
func (m *UserStorageManager) Central() *CentralStorage {
	return m.central
}

// EnsureUserDirectory creates the complete directory structure for a user.
func (m *UserStorageManager) EnsureUserDirectory(userUUID string) error {
	userRoot := m.UserDataRoot(userUUID)
	workspace := m.UserWorkspacePath(userUUID)

	// Core directories
	coreDirs := []string{
		workspace,
		filepath.Join(workspace, "memory"),
		filepath.Join(workspace, "sessions"),
		filepath.Join(workspace, "skills"),
		filepath.Join(workspace, "cron"),
		filepath.Join(workspace, "tasks"),
		filepath.Join(workspace, "temp"),
	}
	for _, dir := range coreDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	// Bootstrap workspace templates
	templates := map[string]string{
		"AGENTS.md": `# Agent Instructions

You are a helpful AI assistant. Be concise, accurate, and friendly.

## Guidelines

- Always explain what you're doing before taking actions
- Ask for clarification when request is ambiguous
- Use tools to help accomplish tasks
- Remember important information in your memory files
- Be proactive and helpful
- Learn from user feedback
`,
		"SOUL.md": `# Soul

I am makoclaw, a lightweight AI assistant powered by AI.

## Personality

- Helpful and friendly
- Concise and to the point
- Curious and eager to learn
- Honest and transparent

## Values

- Accuracy over speed
- User privacy and safety
- Transparency in actions
- Continuous improvement
`,
		"USER.md": `# User

Information about user goes here.

## Preferences

- Communication style: (casual/formal)
- Timezone: (your timezone)
- Language: (your preferred language)
`,
		"IDENTITY.md": `# Identity

## Name
makoclaw (The Apex AI Agent)

## Description
The ultimate evolution of the PicoClaw lineage. makoclaw is an ultra-efficient, Go-native personal AI assistant.

## Version
0.1.0
`,
	}

	for filename, content := range templates {
		filePath := filepath.Join(workspace, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("creating template %s: %w", filename, err)
			}
		}
	}

	// Memory file
	memoryFile := filepath.Join(workspace, "memory", "MEMORY.md")
	if _, err := os.Stat(memoryFile); os.IsNotExist(err) {
		memoryContent := `# Long-term Memory

This file stores important information that should persist across sessions.

## User Information

(Important facts about user)

## Preferences

(User preferences learned over time)

## Important Notes

(Things to remember)
`
		if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
			return fmt.Errorf("creating memory file: %w", err)
		}
	}

	_ = userRoot // used for future per-user files
	return nil
}
