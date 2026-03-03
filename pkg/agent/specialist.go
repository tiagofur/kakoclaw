package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/providers"
	"github.com/sipeed/makoclaw/pkg/storage"
	"github.com/sipeed/makoclaw/pkg/tools"
	kakoclawContext "github.com/sipeed/makoclaw/pkg/utils"
)

// SpecialistAgent represents a specialized agent with specific tools and LLM configuration
type SpecialistAgent struct {
	*AgentLoop
	name         string
	description  string
	prompt       string
	allowedTools map[string]bool
	keywords     []string
	providerName string
	processMu    sync.Mutex // guards tool-swap in ProcessWithSpeciality
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
	Prompt      string   `json:"prompt"`
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
		Prompt:      specialist.prompt,
		Provider:    specialist.providerName,
		Model:       specialist.model,
		Tools:       tools,
		Keywords:    specialist.keywords,
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
		// Create a specialized provider config (copy by value to avoid sync.RWMutex copy)
		specCfg := config.Config{
			Agents:    globalCfg.Agents,
			Channels:  globalCfg.Channels,
			Providers: globalCfg.Providers,
			Gateway:   globalCfg.Gateway,
			Web:       globalCfg.Web,
			Tools:     globalCfg.Tools,
			Storage:   globalCfg.Storage,
		}
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

	// Configure specialist-specific context for lean token usage
	if cfg.Prompt != "" {
		al.contextBuilder.SetAgentSystemPrompt(cfg.Prompt)
	}
	if cfg.Skills != nil {
		al.contextBuilder.SetSkillFilter(cfg.Skills)
	}
	// Specialists run in lightweight mode: skip bootstrap files and memory
	al.contextBuilder.SetLightweightMode(true)

	specialist := &SpecialistAgent{
		AgentLoop:    al,
		name:         name,
		description:  cfg.Description,
		prompt:       cfg.Prompt,
		allowedTools: allowedTools,
		keywords:     cfg.Keywords,
		providerName: cfg.Provider,
	}

	logger.InfoCF("agent", "Specialist agent created", map[string]interface{}{
		"specialist": name,
		"provider":   cfg.Provider,
		"model":      cfg.Model,
		"tools":      len(allowedTools),
		"skills":     len(cfg.Skills),
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

// ProcessWithSpeciality processes a message using this specialist's configuration.
// The specialist's prompt is injected via the system prompt (SetAgentSystemPrompt),
// and tools are filtered to only allowed ones.
// Uses a mutex to serialize concurrent calls that swap the tool registry.
func (sa *SpecialistAgent) ProcessWithSpeciality(ctx context.Context, userMessage string) (string, error) {
	// Specialist prompt is now in the system prompt via SetAgentSystemPrompt,
	// no need to prepend it to the user message.
	fullMessage := userMessage

	// Serialize access to tool-swap to prevent concurrent modifications
	sa.processMu.Lock()
	originalTools := sa.tools
	sa.tools = sa.ToolFilter()
	sa.processMu.Unlock()

	defer func() {
		sa.processMu.Lock()
		sa.tools = originalTools
		sa.processMu.Unlock()
	}()

	// Inject the specialist's name into the context so tools know who is calling them
	agentCtx := kakoclawContext.WithAgentName(ctx, sa.name)

	// Process the message using ProcessDirect
	return sa.ProcessDirect(agentCtx, fullMessage, fmt.Sprintf("specialist_%s", sa.name))
}

// RequestColleagueTool allows specialists to request help from other specialists
type RequestColleagueTool struct {
	registry        *SpecialistRegistry
	currentAgent    *SpecialistAgent
	maxNestingDepth int
	currentDepth    int
}

// NewRequestColleagueTool creates a new colleague request tool for a specialist
func NewRequestColleagueTool(registry *SpecialistRegistry, current *SpecialistAgent) *RequestColleagueTool {
	return &RequestColleagueTool{
		registry:        registry,
		currentAgent:    current,
		maxNestingDepth: 2, // Prevent infinite recursion
		currentDepth:    0,
	}
}

func (t *RequestColleagueTool) Name() string {
	return "request_colleague"
}

func (t *RequestColleagueTool) Description() string {
	return "Request help from a colleague specialist when you need expertise outside your domain"
}

func (t *RequestColleagueTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"colleague": map[string]interface{}{
				"type":        "string",
				"description": "Name of the colleague specialist to consult (e.g., 'security', 'developer')",
			},
			"question": map[string]interface{}{
				"type":        "string",
				"description": "The specific question or task to ask the colleague",
			},
			"context": map[string]interface{}{
				"type":        "string",
				"description": "Additional context about why you need this help",
			},
		},
		"required": []string{"colleague", "question"},
	}
}

func (t *RequestColleagueTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	colleagueName, ok := args["colleague"].(string)
	if !ok {
		return "", fmt.Errorf("colleague must be a string")
	}

	question, ok := args["question"].(string)
	if !ok {
		return "", fmt.Errorf("question must be a string")
	}

	helpContext, _ := args["context"].(string)

	// Prevent infinite recursion
	if t.currentDepth >= t.maxNestingDepth {
		return fmt.Sprintf("⚠️ Cannot request help from %s: maximum collaboration depth reached. "+
			"Please report back to the orchestrator with partial results.", colleagueName), nil
	}

	// Can't ask yourself
	if colleagueName == t.currentAgent.name {
		return "⚠️ Cannot request help from yourself. Try a different colleague.", nil
	}

	// Get the colleague specialist
	colleague, err := t.registry.GetSpecialist(colleagueName)
	if err != nil {
		// Return available colleagues
		available := t.registry.ListSpecialists()
		names := make([]string, 0, len(available))
		for name := range available {
			if name != t.currentAgent.name {
				names = append(names, name)
			}
		}
		return fmt.Sprintf("⚠️ Colleague '%s' not found. Available colleagues: %v", colleagueName, names), nil
	}

	// Emit status event for UI
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:          t.currentAgent.name,
		Status:         "requesting_help",
		SpecialistName: colleagueName,
		Reason:         fmt.Sprintf("Consulting %s: %s", colleagueName, truncateString(question, 100)),
		Timestamp:      time.Now(),
	})

	// Record in team context if available
	if teamCtx := TeamContextFromCtx(ctx); teamCtx != nil {
		teamCtx.Communications = append(teamCtx.Communications, TeamComm{
			From:      t.currentAgent.name,
			To:        colleagueName,
			Message:   question,
			Type:      "request",
			Timestamp: time.Now(),
		})
	}

	// Build the help request with context
	var fullQuestion string
	if helpContext != "" {
		fullQuestion = fmt.Sprintf("A colleague (%s) needs your help:\n\nContext: %s\n\nQuestion: %s",
			t.currentAgent.name, helpContext, question)
	} else {
		fullQuestion = fmt.Sprintf("A colleague (%s) needs your help:\n\nQuestion: %s",
			t.currentAgent.name, question)
	}

	// Execute with incremented depth to track nesting
	t.currentDepth++
	defer func() { t.currentDepth-- }()

	logger.InfoCF("agent", "Specialist requesting colleague help", map[string]interface{}{
		"from":      t.currentAgent.name,
		"to":        colleagueName,
		"question":  truncateString(question, 100),
		"depth":     t.currentDepth,
	})

	// Get response from colleague
	response, err := colleague.ProcessWithSpeciality(ctx, fullQuestion)
	if err != nil {
		return fmt.Sprintf("⚠️ Colleague %s encountered an error: %s", colleagueName, err.Error()), nil
	}

	// Emit completion status
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:     colleagueName,
		Status:    "colleague_complete",
		Timestamp: time.Now(),
	})

	// Record response in team context
	if teamCtx := TeamContextFromCtx(ctx); teamCtx != nil {
		teamCtx.Communications = append(teamCtx.Communications, TeamComm{
			From:      colleagueName,
			To:        t.currentAgent.name,
			Message:   truncateString(response, 200),
			Type:      "response",
			Timestamp: time.Now(),
		})
	}

	return fmt.Sprintf("## Response from %s:\n\n%s", colleagueName, response), nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// LoadSpecialistsFromConfig initializes all configured specialists
func LoadSpecialistsFromConfig(
	cfg *config.Config,
	msgBus *bus.MessageBus,
	baseProvider providers.LLMProvider,
	store *storage.Storage,
) (*SpecialistRegistry, error) {
	registry := NewSpecialistRegistry()

	if len(cfg.Agents.Specialists) == 0 {
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

	// Wire up RequestColleagueTool to all specialists for inter-agent collaboration
	registry.WireColleagueTools()

	return registry, nil
}

// WireColleagueTools adds the request_colleague tool to all specialists
// This must be called after all specialists are registered
func (sr *SpecialistRegistry) WireColleagueTools() {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	for name, specialist := range sr.specialists {
		colleagueTool := NewRequestColleagueTool(sr, specialist)

		// Add to specialist's allowed tools
		specialist.allowedTools["request_colleague"] = true

		// Register the tool in the specialist's tool registry
		if specialist.tools != nil {
			specialist.tools.Register(colleagueTool)
		}

		logger.DebugCF("agent", "Wired colleague tool to specialist", map[string]interface{}{
			"specialist": name,
		})
	}
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
	return sa.keywords
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
