package agent

import (
	"fmt"

	"github.com/sipeed/kakoclaw/pkg/bus"
	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/logger"
	"github.com/sipeed/kakoclaw/pkg/storage"
)

// AgentManager coordinates the orchestrator and specialist agents
type AgentManager struct {
	defaultAgent   *AgentLoop
	orchestrator   *OrchestratorAgent
	specialistReg  *SpecialistRegistry
	isOrchestarted bool
}

// NewAgentManager creates a new agent manager
func NewAgentManager(defaultAgent *AgentLoop) *AgentManager {
	return &AgentManager{
		defaultAgent:   defaultAgent,
		specialistReg:  NewSpecialistRegistry(),
		isOrchestarted: false,
	}
}

// InitializeOrchestrator sets up the orchestrator and specialists if configured
func (am *AgentManager) InitializeOrchestrator(
	cfg *config.Config,
	msgBus *bus.MessageBus,
	storage *storage.Storage,
) error {
	// Check if orchestrator is enabled
	if !cfg.Agents.Orchestrator.Enabled {
		logger.DebugCF("agent", "Orchestrator not enabled", map[string]interface{}{})
		return nil
	}

	// Check if there are any specialists configured
	if cfg.Agents.Specialists == nil || len(cfg.Agents.Specialists) == 0 {
		logger.WarnCF("agent", "Orchestrator enabled but no specialists configured", map[string]interface{}{})
		return nil
	}

	// Get base provider from the default agent
	baseProvider := am.defaultAgent.provider

	// Load all specialists
	specialistReg, err := LoadSpecialistsFromConfig(cfg, msgBus, baseProvider, storage)
	if err != nil {
		return fmt.Errorf("failed to load specialists: %w", err)
	}

	am.specialistReg = specialistReg

	// Create orchestrator agent
	orchestrator, err := NewOrchestratorAgent(cfg, msgBus, baseProvider, specialistReg, storage)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	am.orchestrator = orchestrator
	am.isOrchestarted = true

	logger.InfoCF("agent", "Orchestrator initialized successfully", map[string]interface{}{
		"specialists": len(specialistReg.ListSpecialists()),
	})

	return nil
}

// GetActiveAgent returns the active agent (orchestrator if enabled, otherwise default)
func (am *AgentManager) GetActiveAgent() *AgentLoop {
	if am.isOrchestarted && am.orchestrator != nil {
		return am.orchestrator.AgentLoop
	}
	return am.defaultAgent
}

// GetOrchestrator returns the orchestrator if initialized
func (am *AgentManager) GetOrchestrator() *OrchestratorAgent {
	return am.orchestrator
}

// GetSpecialistRegistry returns the specialist registry
func (am *AgentManager) GetSpecialistRegistry() *SpecialistRegistry {
	return am.specialistReg
}

// IsOrchestrated returns whether the orchestrator is active
func (am *AgentManager) IsOrchestrated() bool {
	return am.isOrchestarted
}

// AddOrUpdateSpecialist registers a specialist from config, replacing any existing entry.
func (am *AgentManager) AddOrUpdateSpecialist(name string, cfg *config.SpecialistConfig, globalCfg *config.Config, store *storage.Storage) (*SpecialistAgent, error) {
	if am == nil || am.defaultAgent == nil {
		return nil, fmt.Errorf("agent manager not initialized")
	}
	if cfg == nil {
		return nil, fmt.Errorf("specialist config is required")
	}
	if cfg.Name == "" {
		cfg.Name = name
	}
	if cfg.Name == "" {
		return nil, fmt.Errorf("specialist name is required")
	}

	baseProvider := am.defaultAgent.provider
	msgBus := am.defaultAgent.bus

	specialist, err := NewSpecialistAgent(cfg.Name, cfg, globalCfg, msgBus, baseProvider, store)
	if err != nil {
		return nil, err
	}

	if err := am.specialistReg.RegisterSpecialist(specialist); err != nil {
		return nil, err
	}

	return specialist, nil
}

// RemoveSpecialist removes a specialist from the registry.
func (am *AgentManager) RemoveSpecialist(name string) bool {
	if am == nil || am.specialistReg == nil {
		return false
	}
	return am.specialistReg.RemoveSpecialist(name)
}
