package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/sipeed/kakoclaw/pkg/bus"
	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/logger"
	"github.com/sipeed/kakoclaw/pkg/providers"
	"github.com/sipeed/kakoclaw/pkg/storage"
	"github.com/sipeed/kakoclaw/pkg/tools"
)

// SpecialistAgent represents a specialized agent with specific tools and LLM configuration
type SpecialistAgent struct {
	*AgentLoop
	name         string
	description  string
	prompt       string
	allowedTools map[string]bool
}

// SpecialistRegistry manages all configured specialist agents
type SpecialistRegistry struct {
	specialists map[string]*SpecialistAgent
	mu          sync.RWMutex
}

// NewSpecialistRegistry creates a new registry for specialist agents
func NewSpecialistRegistry() *SpecialistRegistry {
	return &SpecialistRegistry{
		specialists: make(map[string]*SpecialistAgent),
	}
}

// RegisterSpecialist adds a specialist agent to the registry
func (sr *SpecialistRegistry) RegisterSpecialist(specialist *SpecialistAgent) error {
	if specialist == nil {
		return fmt.Errorf("specialist cannot be nil")
	}
	if specialist.name == "" {
		return fmt.Errorf("specialist name cannot be empty")
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.specialists[specialist.name] = specialist
	logger.DebugCF("agent", "Specialist registered", map[string]interface{}{
		"specialist": specialist.name,
		"tools":      len(specialist.allowedTools),
	})

	return nil
}

// GetSpecialist retrieves a specialist by name
func (sr *SpecialistRegistry) GetSpecialist(name string) (*SpecialistAgent, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	specialist, exists := sr.specialists[name]
	if !exists {
		return nil, fmt.Errorf("specialist '%s' not found", name)
	}

	return specialist, nil
}

// ListSpecialists returns all registered specialists
func (sr *SpecialistRegistry) ListSpecialists() map[string]*SpecialistAgent {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	// Return a copy to prevent external modifications
	result := make(map[string]*SpecialistAgent, len(sr.specialists))
	for k, v := range sr.specialists {
		result[k] = v
	}
	return result
}

// RemoveSpecialist deletes a specialist from the registry
func (sr *SpecialistRegistry) RemoveSpecialist(name string) bool {
	if name == "" {
		return false
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	if _, exists := sr.specialists[name]; !exists {
		return false
	}

	delete(sr.specialists, name)
	return true
}

// GetSpecialistInfo returns metadata about a specialist
type SpecialistInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Provider    string   `json:"provider"`
	Model       string   `json:"model"`
	Tools       []string `json:"tools"`
	Keywords    []string `json:"keywords"`
}

func (sr *SpecialistRegistry) GetSpecialistInfo(name string) (*SpecialistInfo, error) {
	specialist, err := sr.GetSpecialist(name)
	if err != nil {
		return nil, err
	}

	tools := make([]string, 0)
	for tool := range specialist.allowedTools {
		tools = append(tools, tool)
	}

	return &SpecialistInfo{
		Name:        specialist.name,
		Description: specialist.description,
		Provider:    specialist.model,
		Tools:       tools,
	}, nil
}

// NewSpecialistAgent creates a new specialist agent with filtered tools
func NewSpecialistAgent(
	name string,
	cfg *config.SpecialistConfig,
	globalCfg *config.Config,
	msgBus *bus.MessageBus,
	baseProvider providers.LLMProvider,
	store *storage.Storage,
) (*SpecialistAgent, error) {
	if name == "" || cfg == nil {
		return nil, fmt.Errorf("name and config are required")
	}

	// Create or use provided LLM provider for specialist
	var provider providers.LLMProvider
	var err error

	// If specialist has custom provider config, use it; otherwise use base
	if cfg.Provider != "" {
		// Create a specialized provider config
		specCfg := *globalCfg
		specCfg.Agents.Defaults.Provider = cfg.Provider
		specCfg.Agents.Defaults.Model = cfg.Model
		specCfg.Agents.Defaults.MaxTokens = cfg.MaxTokens
		specCfg.Agents.Defaults.Temperature = cfg.Temperature
		specCfg.Agents.Defaults.MaxToolIterations = cfg.MaxToolIterations

		provider, err = providers.CreateProvider(&specCfg)
		if err != nil {
			logger.WarnCF("agent", "Failed to create specialized provider, using base", map[string]interface{}{
				"specialist": name,
				"error":      err.Error(),
			})
			provider = baseProvider
		}
	} else {
		provider = baseProvider
	}

	// Create base agent loop for specialist
	al := NewAgentLoop(globalCfg, msgBus, provider)

	// Build allowed tools map from config
	allowedTools := make(map[string]bool)
	for _, tool := range cfg.Tools {
		allowedTools[tool] = true
	}

	specialist := &SpecialistAgent{
		AgentLoop:    al,
		name:         name,
		description:  cfg.Description,
		prompt:       cfg.Prompt,
		allowedTools: allowedTools,
	}

	logger.InfoCF("agent", "Specialist agent created", map[string]interface{}{
		"specialist": name,
		"provider":   cfg.Provider,
		"model":      cfg.Model,
		"tools":      len(allowedTools),
	})

	return specialist, nil
}

// ToolFilter returns a filtered tool registry containing only allowed tools for this specialist
func (sa *SpecialistAgent) ToolFilter() *tools.ToolRegistry {
	filteredRegistry := tools.NewToolRegistry()

	// Register only allowed tools from the original registry
	if len(sa.allowedTools) == 0 {
		// If no restrictions, register all tools
		sa.tools.ForEach(func(tool tools.Tool) {
			filteredRegistry.Register(tool)
		})
	} else {
		// Register only allowed tools
		sa.tools.ForEach(func(tool tools.Tool) {
			if sa.allowedTools[tool.Name()] {
				filteredRegistry.Register(tool)
			}
		})
	}

	return filteredRegistry
}

// ProcessWithSpeciality processes a message using this specialist's configuration
func (sa *SpecialistAgent) ProcessWithSpeciality(ctx context.Context, userMessage string) (string, error) {
	// Build a specialized message with the specialist's prompt
	fullMessage := userMessage
	if sa.prompt != "" {
		fullMessage = sa.prompt + "\n\n" + userMessage
	}

	// Use the specialist's filtered tools
	originalTools := sa.tools
	sa.tools = sa.ToolFilter()
	defer func() {
		sa.tools = originalTools
	}()

	// Process the message using ProcessDirect
	return sa.ProcessDirect(ctx, fullMessage, fmt.Sprintf("specialist_%s", sa.name))
}

// LoadSpecialistsFromConfig initializes all configured specialists
func LoadSpecialistsFromConfig(
	cfg *config.Config,
	msgBus *bus.MessageBus,
	baseProvider providers.LLMProvider,
	store *storage.Storage,
) (*SpecialistRegistry, error) {
	registry := NewSpecialistRegistry()

	if cfg.Agents.Specialists == nil || len(cfg.Agents.Specialists) == 0 {
		logger.DebugCF("agent", "No specialists configured", map[string]interface{}{})
		return registry, nil
	}

	for name, specCfg := range cfg.Agents.Specialists {
		// Validate specialist config
		if specCfg.Name == "" {
			specCfg.Name = name
		}

		specialist, err := NewSpecialistAgent(name, &specCfg, cfg, msgBus, baseProvider, store)
		if err != nil {
			logger.WarnCF("agent", "Failed to create specialist", map[string]interface{}{
				"specialist": name,
				"error":      err.Error(),
			})
			continue
		}

		if err := registry.RegisterSpecialist(specialist); err != nil {
			logger.WarnCF("agent", "Failed to register specialist", map[string]interface{}{
				"specialist": name,
				"error":      err.Error(),
			})
			continue
		}
	}

	logger.InfoCF("agent", "Specialists loaded", map[string]interface{}{
		"count": len(registry.specialists),
	})

	return registry, nil
}

// SpecialistSelector helps select appropriate specialist agents for tasks
type SpecialistSelector struct {
	registry *SpecialistRegistry
}

// NewSpecialistSelector creates a new specialist selector
func NewSpecialistSelector(registry *SpecialistRegistry) *SpecialistSelector {
	return &SpecialistSelector{
		registry: registry,
	}
}

// SelectByKeywords finds specialists matching given keywords
func (ss *SpecialistSelector) SelectByKeywords(keywords []string) []*SpecialistAgent {
	var matches []*SpecialistAgent

	for _, specialist := range ss.registry.ListSpecialists() {
		for _, keyword := range keywords {
			// Check if keyword is in specialist's description or keywords
			if contains(specialist.GetSpecialistKeywords(), keyword) {
				matches = append(matches, specialist)
				break
			}
		}
	}

	return matches
}

// GetSpecialistKeywords returns the specialist's keywords
func (sa *SpecialistAgent) GetSpecialistKeywords() []string {
	// For now, extract from description or return empty
	// This could be expanded to include more metadata
	return []string{}
}

// GetDescription returns the specialist's description
func (sa *SpecialistAgent) GetDescription() string {
	return sa.description
}

// GetName returns the specialist's name
func (sa *SpecialistAgent) GetName() string {
	return sa.name
}

// GetAllowedTools returns the list of allowed tools for this specialist
func (sa *SpecialistAgent) GetAllowedTools() []string {
	tools := make([]string, 0, len(sa.allowedTools))
	for tool := range sa.allowedTools {
		tools = append(tools, tool)
	}
	return tools
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// SelectByName returns a single specialist by name
func (ss *SpecialistSelector) SelectByName(name string) (*SpecialistAgent, error) {
	return ss.registry.GetSpecialist(name)
}

// SelectAll returns all available specialists
func (ss *SpecialistSelector) SelectAll() []*SpecialistAgent {
	specialists := ss.registry.ListSpecialists()
	result := make([]*SpecialistAgent, 0, len(specialists))
	for _, s := range specialists {
		result = append(result, s)
	}
	return result
}
