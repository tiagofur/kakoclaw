package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sipeed/makoclaw/pkg/config"
)

// UserStorageManager manages per-user SQLite database connections.
// Each user gets their own database at {dataRoot}/users/{uuid}/user.db.
type UserStorageManager struct {
	central  *CentralStorage
	dataRoot string              // e.g., /home/makoclaw/.MakoClaw
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
	_, err := config.EnsureUserWorkspace(userUUID)
	return err
}
