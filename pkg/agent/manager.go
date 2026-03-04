package agent

import (
	"context"
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
	swarmRunner    *SwarmRunner
	swarmConfigs   map[string]config.SwarmConfig
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

// InitializeSwarms sets up the swarm runner and loads swarm configs
func (am *AgentManager) InitializeSwarms(cfg *config.Config) {
	if am.specialistReg == nil {
		return
	}

	var costTracker *AgentCostTracker
	if am.defaultAgent != nil {
		costTracker = am.defaultAgent.costTracker
	}

	am.swarmRunner = NewSwarmRunner(am.specialistReg, costTracker)

	am.swarmConfigs = make(map[string]config.SwarmConfig)
	if cfg.Agents.Swarms != nil {
		for name, swarm := range cfg.Agents.Swarms {
			am.swarmConfigs[name] = swarm
		}
	}

	if len(am.swarmConfigs) > 0 {
		logger.InfoCF("agent", "Swarms loaded", map[string]interface{}{
			"count": len(am.swarmConfigs),
		})
	}
}

// RunSwarm executes a named swarm with the given task
func (am *AgentManager) RunSwarm(ctx context.Context, name, task string) (*SwarmResult, error) {
	if am.swarmRunner == nil {
		return nil, fmt.Errorf("swarm runner not initialized")
	}

	swarmCfg, ok := am.swarmConfigs[name]
	if !ok {
		return nil, fmt.Errorf("swarm '%s' not found", name)
	}

	return am.swarmRunner.RunSwarm(ctx, swarmCfg, task)
}

// RunSwarmWithConfig executes a swarm from a config directly (for ad-hoc runs)
func (am *AgentManager) RunSwarmWithConfig(ctx context.Context, swarmCfg config.SwarmConfig, task string) (*SwarmResult, error) {
	if am.swarmRunner == nil {
		return nil, fmt.Errorf("swarm runner not initialized")
	}

	return am.swarmRunner.RunSwarm(ctx, swarmCfg, task)
}

// GetSwarmsList returns info about all configured swarms
func (am *AgentManager) GetSwarmsList() []map[string]interface{} {
	if am == nil || am.swarmConfigs == nil {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, 0, len(am.swarmConfigs))
	for name, cfg := range am.swarmConfigs {
		result = append(result, map[string]interface{}{
			"name":          name,
			"description":   cfg.Description,
			"members":       cfg.Members,
			"mode":          cfg.Mode,
			"max_budget":    cfg.MaxBudget,
			"timeout":       cfg.Timeout,
			"shared_memory": cfg.SharedMemory,
		})
	}
	return result
}

// GetSwarmConfig returns a specific swarm config
func (am *AgentManager) GetSwarmConfig(name string) (config.SwarmConfig, bool) {
	if am == nil || am.swarmConfigs == nil {
		return config.SwarmConfig{}, false
	}
	cfg, ok := am.swarmConfigs[name]
	return cfg, ok
}

// SetSwarmConfig adds or updates a swarm config at runtime
func (am *AgentManager) SetSwarmConfig(name string, cfg config.SwarmConfig) {
	if am.swarmConfigs == nil {
		am.swarmConfigs = make(map[string]config.SwarmConfig)
	}
	am.swarmConfigs[name] = cfg
}

// DeleteSwarmConfig removes a swarm config
func (am *AgentManager) DeleteSwarmConfig(name string) bool {
	if am.swarmConfigs == nil {
		return false
	}
	if _, ok := am.swarmConfigs[name]; !ok {
		return false
	}
	delete(am.swarmConfigs, name)
	return true
}
