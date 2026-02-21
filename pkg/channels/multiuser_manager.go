package channels

import (
	"context"
	"fmt"
	"sync"

	"github.com/sipeed/kakoclaw/pkg/agent"
	"github.com/sipeed/kakoclaw/pkg/bus"
	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/logger"
	"github.com/sipeed/kakoclaw/pkg/storage"
)

// MultiUserChannelManager manages channel instances for multiple users.
// Each user can have their own channel configurations (Telegram bots, Discord bots, etc.)
type MultiUserChannelManager struct {
	managers  map[string]*Manager         // userUUID -> channel Manager
	agents    map[string]*agent.AgentLoop // userUUID -> agent loop
	globalCfg *config.Config
	bus       *bus.MessageBus
	storage   *storage.Storage
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewMultiUserChannelManager creates a manager that handles channels for all users
func NewMultiUserChannelManager(globalCfg *config.Config, bus *bus.MessageBus, store *storage.Storage) *MultiUserChannelManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &MultiUserChannelManager{
		managers:  make(map[string]*Manager),
		agents:    make(map[string]*agent.AgentLoop),
		globalCfg: globalCfg,
		bus:       bus,
		storage:   store,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// InitializeAllUsers loads all users from storage and initializes their channel managers
func (m *MultiUserChannelManager) InitializeAllUsers() error {
	if m.storage == nil {
		return fmt.Errorf("storage is required for multi-user channel manager")
	}

	users, err := m.storage.ListUsers()
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	logger.InfoCF("multiuser", "Initializing channels for users", map[string]interface{}{
		"user_count": len(users),
	})

	for _, user := range users {
		if _, err := m.GetOrCreateManagerForUser(user.UUID); err != nil {
			logger.ErrorCF("multiuser", "Failed to initialize channels for user", map[string]interface{}{
				"user_uuid": user.UUID,
				"username":  user.Username,
				"error":     err.Error(),
			})
		}
	}

	return nil
}

// GetOrCreateManagerForUser returns the channel manager for a user, creating it if necessary.
// This loads the user's config and creates their channel instances.
func (m *MultiUserChannelManager) GetOrCreateManagerForUser(userUUID string) (*Manager, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return existing manager if already initialized
	if mgr, exists := m.managers[userUUID]; exists {
		return mgr, nil
	}

	// Create agent loop for user
	agentLoop, err := agent.NewAgentLoopForUser(userUUID, m.globalCfg, m.bus, m.storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent loop for user: %w", err)
	}

	// Load user's merged config
	userCfg, err := config.LoadConfigForUser(userUUID)
	if err != nil {
		logger.WarnCF("multiuser", "Failed to load user config, using global", map[string]interface{}{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		userCfg = m.globalCfg
	}

	// Merge with global config
	mergedCfg := config.MergeConfigs(m.globalCfg, userCfg)

	// Create channel manager with user's merged config
	channelManager, err := NewManager(mergedCfg, m.bus, m.storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel manager: %w", err)
	}

	// Store references
	m.managers[userUUID] = channelManager
	m.agents[userUUID] = agentLoop

	logger.InfoCF("multiuser", "Created channel manager for user", map[string]interface{}{
		"user_uuid": userUUID,
		"channels":  mergedCfg.GetActiveChannels(),
	})

	return channelManager, nil
}

// GetManagerForUser returns the channel manager for a user without creating it
func (m *MultiUserChannelManager) GetManagerForUser(userUUID string) (*Manager, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mgr, exists := m.managers[userUUID]
	return mgr, exists
}

// GetAgentForUser returns the agent loop for a user
func (m *MultiUserChannelManager) GetAgentForUser(userUUID string) (*agent.AgentLoop, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	al, exists := m.agents[userUUID]
	return al, exists
}

// StartAll starts all channel managers for all initialized users
func (m *MultiUserChannelManager) StartAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logger.InfoCF("multiuser", "Starting all user channel managers", map[string]interface{}{
		"user_count": len(m.managers),
	})

	for userUUID, mgr := range m.managers {
		if err := mgr.StartAll(ctx); err != nil {
			logger.ErrorCF("multiuser", "Failed to start channels for user", map[string]interface{}{
				"user_uuid": userUUID,
				"error":     err.Error(),
			})
		} else {
			logger.InfoCF("multiuser", "Started channels for user", map[string]interface{}{
				"user_uuid": userUUID,
			})
		}

		// Start agent loop
		if al, exists := m.agents[userUUID]; exists {
			go al.Run(ctx)
		}
	}

	return nil
}

// StopAll stops all channel managers
func (m *MultiUserChannelManager) StopAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logger.InfoC("multiuser", "Stopping all user channel managers")

	for userUUID, mgr := range m.managers {
		if err := mgr.StopAll(ctx); err != nil {
			logger.ErrorCF("multiuser", "Failed to stop channels for user", map[string]interface{}{
				"user_uuid": userUUID,
				"error":     err.Error(),
			})
		}

		// Stop agent loop
		if al, exists := m.agents[userUUID]; exists {
			al.Stop()
		}
	}

	m.cancel()
	return nil
}

// RestartUserChannels restarts all channels for a specific user (e.g., after config change)
func (m *MultiUserChannelManager) RestartUserChannels(ctx context.Context, userUUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop existing manager/agent
	if mgr, exists := m.managers[userUUID]; exists {
		if err := mgr.StopAll(ctx); err != nil {
			logger.WarnCF("multiuser", "Error stopping channels during restart", map[string]interface{}{
				"user_uuid": userUUID,
				"error":     err.Error(),
			})
		}
	}
	if al, exists := m.agents[userUUID]; exists {
		al.Stop()
	}

	// Remove old references
	delete(m.managers, userUUID)
	delete(m.agents, userUUID)

	// Recreate with new config
	mgr, err := m.GetOrCreateManagerForUser(userUUID)
	if err != nil {
		return fmt.Errorf("failed to recreate manager: %w", err)
	}

	// Start new manager
	if err := mgr.StartAll(ctx); err != nil {
		return fmt.Errorf("failed to start channels: %w", err)
	}

	// Start new agent
	if al, exists := m.agents[userUUID]; exists {
		go al.Run(ctx)
	}

	logger.InfoCF("multiuser", "Restarted channels for user", map[string]interface{}{
		"user_uuid": userUUID,
	})

	return nil
}

// GetUserCount returns the number of initialized users
func (m *MultiUserChannelManager) GetUserCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.managers)
}
