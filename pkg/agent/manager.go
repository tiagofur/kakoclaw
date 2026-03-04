package agent

import (
	"fmt"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
)

// AgentManager coordinates the orchestrator and specialist agents
type AgentManager struct {
	defaultAgent   *AgentLoop
	orchestrator   *OrchestratorAgent
	specialistReg  *SpecialistRegistry
	isOrchestarted bool
}

// NewAgentManager creates a new agent manager.
// defaultAgent can be nil for degraded mode (no LLM provider configured).
func NewAgentManager(defaultAgent *AgentLoop) *AgentManager {
	return &AgentManager{
		defaultAgent:   defaultAgent,
		specialistReg:  NewSpecialistRegistry(),
		isOrchestarted: false,
	}
}

// IsReady returns true if the agent manager has a valid provider and can process requests.
// Returns false in degraded mode when no LLM provider is configured.
func (am *AgentManager) IsReady() bool {
	return am != nil && am.defaultAgent != nil && am.defaultAgent.provider != nil
}

// InitializeOrchestrator sets up the orchestrator and specialists if configured.
// Returns nil without error if agent manager is not ready (degraded mode).
func (am *AgentManager) InitializeOrchestrator(
	cfg *config.Config,
	msgBus *bus.MessageBus,
	storage *storage.Storage,
) error {
	// Skip initialization if not ready (degraded mode)
	if !am.IsReady() {
		logger.DebugCF("agent", "Orchestrator initialization skipped (degraded mode)", map[string]interface{}{})
		return nil
	}

	// Check if there are any specialists configured
	if len(cfg.Agents.Specialists) > 0 {
		// Get base provider from the default agent
		baseProvider := am.defaultAgent.provider

		// Load all specialists
		specialistReg, err := LoadSpecialistsFromConfig(cfg, msgBus, baseProvider, storage)
		if err != nil {
			return fmt.Errorf("failed to load specialists: %w", err)
		}

		am.specialistReg = specialistReg
		logger.InfoCF("agent", "Specialists loaded", map[string]interface{}{
			"count": len(specialistReg.ListSpecialists()),
		})
	}

	// Check if orchestrator is enabled
	if !cfg.Agents.Orchestrator.Enabled {
		logger.DebugCF("agent", "Orchestrator not enabled", map[string]interface{}{})
		return nil
	}

	if am.specialistReg == nil || len(am.specialistReg.ListSpecialists()) == 0 {
		logger.WarnCF("agent", "Orchestrator enabled but no specialists configured", map[string]interface{}{})
		return nil
	}

	baseProvider := am.defaultAgent.provider

	// Create orchestrator agent
	orchestrator, err := NewOrchestratorAgent(cfg, msgBus, baseProvider, am.specialistReg, storage)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	am.orchestrator = orchestrator
	am.isOrchestarted = true

	// Propagate user context dependencies so the orchestrator and specialists
	// write sessions/messages to the user's private workspace instead of the
	// global default workspace.
	if am.defaultAgent.centralStorage != nil {
		orchestrator.SetCentralStorage(am.defaultAgent.centralStorage)
		for _, spec := range am.specialistReg.ListSpecialists() {
			spec.SetCentralStorage(am.defaultAgent.centralStorage)
		}
	}
	if am.defaultAgent.storage != nil {
		orchestrator.SetStorage(am.defaultAgent.storage)
		for _, spec := range am.specialistReg.ListSpecialists() {
			spec.SetStorage(am.defaultAgent.storage)
		}
	}

	logger.InfoCF("agent", "Orchestrator initialized successfully", map[string]interface{}{
		"specialists": len(am.specialistReg.ListSpecialists()),
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
	if cfg == nil {
		return nil, fmt.Errorf("specialist config is required")
	}
	if cfg.Name == "" {
		cfg.Name = name
	}
	if cfg.Name == "" {
		return nil, fmt.Errorf("specialist name is required")
	}

	if !am.IsReady() {
		logger.InfoCF("agent", "Specialist config saved but runtime registration skipped (degraded mode)", map[string]interface{}{
			"specialist": cfg.Name,
		})
		return nil, nil
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

// GetSpecialistsList returns info about all registered specialists for the API
func (am *AgentManager) GetSpecialistsList() []map[string]interface{} {
	if am == nil || am.specialistReg == nil {
		return []map[string]interface{}{}
	}

	specialists := am.specialistReg.ListSpecialists()
	result := make([]map[string]interface{}, 0, len(specialists))

	for name, spec := range specialists {
		if name == "orchestrator" {
			continue // Skip orchestrator itself
		}

		tools := make([]string, 0, len(spec.allowedTools))
		for tool := range spec.allowedTools {
			tools = append(tools, tool)
		}

		result = append(result, map[string]interface{}{
			"name":        name,
			"description": spec.description,
			"tools":       tools,
			"keywords":    spec.keywords,
		})
	}

	return result
}
