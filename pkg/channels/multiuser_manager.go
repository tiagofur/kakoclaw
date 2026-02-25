package channels

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sipeed/makoclaw/pkg/agent"
	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/cron"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
	"github.com/sipeed/makoclaw/pkg/tools"
)

// MultiUserChannelManager manages channel instances for multiple users.
// Each user can have their own channel configurations (Telegram bots, Discord bots, etc.)
type MultiUserChannelManager struct {
	managers     map[string]*Manager         // userUUID -> channel Manager
	agents       map[string]*agent.AgentLoop // userUUID -> agent loop
	cronTools    map[string]*tools.CronTool
	cronJobs     map[string]*cron.CronService
	globalCfg    *config.Config
	bus          *bus.MessageBus
	channelStore *storage.Storage        // For channel manager initialization
	centralStore *storage.CentralStorage // For user/cron management
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewMultiUserChannelManager creates a manager that handles channels for all users
func NewMultiUserChannelManager(globalCfg *config.Config, bus *bus.MessageBus, channelStore *storage.Storage, centralStore *storage.CentralStorage) *MultiUserChannelManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &MultiUserChannelManager{
		managers:     make(map[string]*Manager),
		agents:       make(map[string]*agent.AgentLoop),
		cronTools:    make(map[string]*tools.CronTool),
		cronJobs:     make(map[string]*cron.CronService),
		globalCfg:    globalCfg,
		bus:          bus,
		channelStore: channelStore,
		centralStore: centralStore,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// InitializeAllUsers loads all users from storage and initializes their channel managers
func (m *MultiUserChannelManager) InitializeAllUsers() error {
	if m.centralStore == nil {
		return fmt.Errorf("central storage is required for multi-user channel manager")
	}

	if err := m.migrateLegacyCronStore(); err != nil {
		logger.WarnCF("multiuser", "Failed to migrate legacy cron store", map[string]interface{}{
			"error": err.Error(),
		})
	}

	users, err := m.centralStore.ListUsers()
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
	return m.getOrCreateManagerForUserLocked(userUUID)
}

func (m *MultiUserChannelManager) getOrCreateManagerForUserLocked(userUUID string) (*Manager, error) {
	// Return existing manager if already initialized
	if mgr, exists := m.managers[userUUID]; exists {
		return mgr, nil
	}

	// Create agent loop for user
	agentLoop, err := agent.NewAgentLoopForUser(userUUID, m.globalCfg, m.bus, m.centralStore)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent loop for user: %w", err)
	}

	workspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure user workspace: %w", err)
	}

	cronStorePath := filepath.Join(workspace, "cron", "jobs.json")
	cronService := cron.NewCronService(cronStorePath, nil)
	cronTool := tools.NewCronTool(cronService, agentLoop, m.bus)
	agentLoop.RegisterTool(cronTool)
	cronService.SetOnJob(func(job *cron.CronJob) (string, error) {
		result := cronTool.ExecuteJob(m.ctx, job)
		return result, nil
	})

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
	channelManager, err := NewManager(mergedCfg, m.bus, m.channelStore)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel manager: %w", err)
	}

	// Store references
	m.managers[userUUID] = channelManager
	m.agents[userUUID] = agentLoop
	m.cronTools[userUUID] = cronTool
	m.cronJobs[userUUID] = cronService

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

// GetCronServiceForUser returns the cron service for a user
func (m *MultiUserChannelManager) GetCronServiceForUser(userUUID string) (*cron.CronService, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cs, exists := m.cronJobs[userUUID]
	return cs, exists
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

		if cronService, exists := m.cronJobs[userUUID]; exists {
			if err := cronService.Start(); err != nil {
				logger.WarnCF("multiuser", "Failed to start cron service for user", map[string]interface{}{
					"user_uuid": userUUID,
					"error":     err.Error(),
				})
			}
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

		if cronService, exists := m.cronJobs[userUUID]; exists {
			cronService.Stop()
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
	if cronService, exists := m.cronJobs[userUUID]; exists {
		cronService.Stop()
	}

	// Remove old references
	delete(m.managers, userUUID)
	delete(m.agents, userUUID)
	delete(m.cronTools, userUUID)
	delete(m.cronJobs, userUUID)

	// Recreate with new config
	mgr, err := m.getOrCreateManagerForUserLocked(userUUID)
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
	if cronService, exists := m.cronJobs[userUUID]; exists {
		if err := cronService.Start(); err != nil {
			logger.WarnCF("multiuser", "Failed to start cron service after restart", map[string]interface{}{
				"user_uuid": userUUID,
				"error":     err.Error(),
			})
		}
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

func (m *MultiUserChannelManager) migrateLegacyCronStore() error {
	if m.globalCfg == nil {
		return nil
	}
	legacyPath := filepath.Join(m.globalCfg.WorkspacePath(), "cron", "jobs.json")
	info, err := os.Stat(legacyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("stat legacy cron store: %w", err)
	}
	if info.IsDir() {
		return nil
	}

	store, err := readCronStore(legacyPath)
	if err != nil {
		return fmt.Errorf("read legacy cron store: %w", err)
	}
	if len(store.Jobs) == 0 {
		return nil
	}
	if m.centralStore == nil {
		return fmt.Errorf("central storage required for cron migration")
	}

	jobsByUser := make(map[string][]cron.CronJob)
	for _, job := range store.Jobs {
		if job.UserID == 0 {
			job.UserID = 1
		}
		user, err := m.centralStore.GetUserByID(job.UserID)
		if err != nil || user == nil || user.UUID == "" {
			// Fallback: assign to first available user rather than silently discard
			allUsers, listErr := m.centralStore.ListUsers()
			if listErr != nil || len(allUsers) == 0 {
				logger.WarnCF("multiuser", "Skipping cron job migration - user not found and no fallback available", map[string]interface{}{
					"user_id":  job.UserID,
					"job_id":   job.ID,
					"job_name": job.Name,
				})
				continue
			}
			fallback := allUsers[0]
			logger.WarnCF("multiuser", "Cron job user not found, migrating to first user as fallback", map[string]interface{}{
				"original_user_id": job.UserID,
				"fallback_uuid":    fallback.UUID,
				"job_id":           job.ID,
				"job_name":         job.Name,
			})
			job.UserID = fallback.ID
			jobsByUser[fallback.UUID] = append(jobsByUser[fallback.UUID], job)
			continue
		}
		jobsByUser[user.UUID] = append(jobsByUser[user.UUID], job)
	}

	for userUUID, jobs := range jobsByUser {
		workspace, err := config.EnsureUserWorkspace(userUUID)
		if err != nil {
			logger.WarnCF("multiuser", "Failed to ensure user workspace for cron migration", map[string]interface{}{
				"user_uuid": userUUID,
				"error":     err.Error(),
			})
			continue
		}
		storePath := filepath.Join(workspace, "cron", "jobs.json")
		perUserStore, err := readCronStore(storePath)
		if err != nil {
			return fmt.Errorf("read per-user cron store: %w", err)
		}

		existing := make(map[string]struct{})
		for _, job := range perUserStore.Jobs {
			existing[job.ID] = struct{}{}
		}
		for _, job := range jobs {
			if _, found := existing[job.ID]; found {
				continue
			}
			perUserStore.Jobs = append(perUserStore.Jobs, job)
		}
		if err := writeCronStore(storePath, perUserStore); err != nil {
			return fmt.Errorf("write per-user cron store: %w", err)
		}
	}

	migratedPath := legacyPath + ".migrated"
	if err := os.Rename(legacyPath, migratedPath); err != nil {
		logger.WarnCF("multiuser", "Failed to rename legacy cron store", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return nil
}

func readCronStore(path string) (cron.CronStore, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cron.CronStore{Version: 1, Jobs: []cron.CronJob{}}, nil
		}
		return cron.CronStore{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cron.CronStore{}, err
	}
	if len(data) == 0 {
		return cron.CronStore{Version: 1, Jobs: []cron.CronJob{}}, nil
	}
	var store cron.CronStore
	if err := json.Unmarshal(data, &store); err != nil {
		return cron.CronStore{}, err
	}
	if store.Version == 0 {
		store.Version = 1
	}
	return store, nil
}

func writeCronStore(path string, store cron.CronStore) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// GetOrCreateCronServiceForUser creates a cron service for a user even in degraded mode.
// This is used to allow per-user cron job scheduling without requiring a full agent loop.
// In degraded mode, only deliver=true jobs will execute (no agent processing).
func (m *MultiUserChannelManager) GetOrCreateCronServiceForUser(userUUID string) (*cron.CronService, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return existing cron service if already created
	if cronService, exists := m.cronJobs[userUUID]; exists {
		return cronService, nil
	}

	// Create cron service storage location
	workspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure user workspace for cron: %w", err)
	}

	cronStorePath := filepath.Join(workspace, "cron", "jobs.json")
	cronService := cron.NewCronService(cronStorePath, nil)

	// For degraded mode (no agent loop), we create a simplified executor that only does deliver
	// Create a noop agent loop stub that just publishes messages for deliver=true jobs
	ctx := context.Background()
	noop := &noopJobExecutor{bus: m.bus}

	cronTool := tools.NewCronTool(cronService, noop, m.bus)
	cronService.SetOnJob(func(job *cron.CronJob) (string, error) {
		result := cronTool.ExecuteJob(ctx, job)
		return result, nil
	})

	// Store cron service
	m.cronJobs[userUUID] = cronService

	logger.InfoCF("multiuser", "Created cron service for user (degraded mode or eager init)", map[string]interface{}{
		"user_uuid": userUUID,
	})

	return cronService, nil
}

// noopJobExecutor is a minimal executor for degraded mode cron that only supports deliver=true jobs
type noopJobExecutor struct {
	bus *bus.MessageBus
}

func (n *noopJobExecutor) ProcessDirectWithChannel(ctx context.Context, content, sessionKey, channel, chatID string) (string, error) {
	// Degraded mode: no agent processing
	return "", fmt.Errorf("agent processing unavailable in degraded mode")
}

func (n *noopJobExecutor) ProcessDirectWithChannelForUser(ctx context.Context, userID int64, content, sessionKey, channel, chatID string) (string, error) {
	// Degraded mode: no agent processing
	return "", fmt.Errorf("agent processing unavailable in degraded mode")
}
