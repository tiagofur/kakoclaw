package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/providers"
	"github.com/sipeed/makoclaw/pkg/storage"
)

type agentTrackerKey struct{}

// ContextWithAgentTracker embeds an AgentLoop into the context so that specialist
// delegations register agent names into the caller's tracker (e.g. the web chat AgentLoop)
// instead of the orchestrator's own internal loop.
func ContextWithAgentTracker(ctx context.Context, tracker *AgentLoop) context.Context {
	return context.WithValue(ctx, agentTrackerKey{}, tracker)
}

func agentTrackerFromCtx(ctx context.Context) *AgentLoop {
	if v, ok := ctx.Value(agentTrackerKey{}).(*AgentLoop); ok {
		return v
	}
	return nil
}

// Context keys for callbacks
type agentStatusCallbackKey struct{}
type contentSegmentCallbackKey struct{}

// ContextWithAgentStatusCallback embeds an AgentStatusCallback into the context
func ContextWithAgentStatusCallback(ctx context.Context, callback AgentStatusCallback) context.Context {
	return context.WithValue(ctx, agentStatusCallbackKey{}, callback)
}

func agentStatusCallbackFromCtx(ctx context.Context) AgentStatusCallback {
	if v, ok := ctx.Value(agentStatusCallbackKey{}).(AgentStatusCallback); ok {
		return v
	}
	return nil
}

// emitAgentStatus emits an agent status event if callback is available
func emitAgentStatus(ctx context.Context, event AgentStatusEvent) {
	if callback := agentStatusCallbackFromCtx(ctx); callback != nil {
		_ = callback(event) // Ignore errors to not break execution flow
	}
}

// ContextWithContentSegmentCallback embeds a ContentSegmentCallback into the context
func ContextWithContentSegmentCallback(ctx context.Context, callback ContentSegmentCallback) context.Context {
	return context.WithValue(ctx, contentSegmentCallbackKey{}, callback)
}

func contentSegmentCallbackFromCtx(ctx context.Context) ContentSegmentCallback {
	if v, ok := ctx.Value(contentSegmentCallbackKey{}).(ContentSegmentCallback); ok {
		return v
	}
	return nil
}

// emitContentSegment emits a content segment if callback is available
func emitContentSegment(ctx context.Context, segment ContentSegment) {
	if callback := contentSegmentCallbackFromCtx(ctx); callback != nil {
		_ = callback(segment)
	}
}

// OrchestratorAgent is a special agent that analyzes tasks and delegates to specialists
type OrchestratorAgent struct {
	*SpecialistAgent
	registry          *SpecialistRegistry
	delegationRetries int
	fallbackToDefault bool
}

// DelegationRequest represents a request to delegate to a specialist
type DelegationRequest struct {
	SpecialistName string `json:"specialist_name"`
	Task           string `json:"task"`
	Context        string `json:"context,omitempty"`
}

// DelegationResult represents the result of delegating to a specialist
type DelegationResult struct {
	SpecialistName string `json:"specialist_name"`
	Result         string `json:"result"`
	Success        bool   `json:"success"`
	Error          string `json:"error,omitempty"`
}

// NewOrchestratorAgent creates a new orchestrator agent
func NewOrchestratorAgent(
	cfg *config.Config,
	msgBus *bus.MessageBus,
	baseProvider providers.LLMProvider,
	registry *SpecialistRegistry,
	store *storage.Storage,
) (*OrchestratorAgent, error) {
	if registry == nil {
		return nil, fmt.Errorf("specialist registry is required")
	}

	// Create specialist agent with orchestrator config
	specCfg := config.SpecialistConfig{
		Name:              "orchestrator",
		Description:       cfg.Agents.Orchestrator.Description,
		Provider:          cfg.Agents.Orchestrator.Provider,
		Model:             cfg.Agents.Orchestrator.Model,
		MaxTokens:         cfg.Agents.Orchestrator.MaxTokens,
		Temperature:       cfg.Agents.Orchestrator.Temperature,
		MaxToolIterations: cfg.Agents.Orchestrator.MaxDelegationRetries,
		Tools:             []string{"delegate_to_specialist"}, // Special tool only for orchestrator
	}

	// Create the specialist agent wrapper
	specialist, err := NewSpecialistAgent("orchestrator", &specCfg, cfg, msgBus, baseProvider, store)
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestrator specialist: %w", err)
	}

	orchestrator := &OrchestratorAgent{
		SpecialistAgent:   specialist,
		registry:          registry,
		delegationRetries: cfg.Agents.Orchestrator.MaxDelegationRetries,
		fallbackToDefault: cfg.Agents.Orchestrator.FallbackToDefault,
	}

	// Register the delegation tool
	orchestrator.registerDelegationTool()

	// Inject orchestrator context into the system prompt so the LLM sees
	// the specialist catalog and delegation rules via the standard BuildSystemPrompt() path.
	orchestrator.contextBuilder.SetAgentSystemPrompt(orchestrator.BuildOrchestratorContext())
	// Orchestrator only delegates — no skills needed
	orchestrator.contextBuilder.SetSkillFilter([]string{})
	// Lean context: skip bootstrap files and memory
	orchestrator.contextBuilder.SetLightweightMode(true)

	logger.InfoCF("agent", "Orchestrator agent created", map[string]interface{}{
		"provider":    cfg.Agents.Orchestrator.Provider,
		"model":       cfg.Agents.Orchestrator.Model,
		"specialists": len(registry.ListSpecialists()),
	})

	return orchestrator, nil
}

// registerDelegationTool registers the delegation tool in the orchestrator's tool registry
func (oa *OrchestratorAgent) registerDelegationTool() {
	delegateTool := &DelegationTool{
		orchestrator: oa,
	}
	oa.tools.Register(delegateTool)
}

// DelegationTool is the tool used by the orchestrator to delegate tasks
type DelegationTool struct {
	orchestrator *OrchestratorAgent
}

func (dt *DelegationTool) Name() string {
	return "delegate_to_specialist"
}

func (dt *DelegationTool) Description() string {
	return "Delegates a task to an appropriate specialist agent"
}

func (dt *DelegationTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"specialist_name": map[string]interface{}{
				"type":        "string",
				"description": "The name of the specialist agent to delegate to",
			},
			"task": map[string]interface{}{
				"type":        "string",
				"description": "The task description to delegate",
			},
			"context": map[string]interface{}{
				"type":        "string",
				"description": "Additional context for the task (optional)",
			},
		},
		"required": []string{"specialist_name", "task"},
	}
}

func (dt *DelegationTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	// Parse delegation request
	specialistName, ok := args["specialist_name"].(string)
	if !ok {
		return "", fmt.Errorf("specialist_name must be a string")
	}

	task, ok := args["task"].(string)
	if !ok {
		return "", fmt.Errorf("task must be a string")
	}

	contextStr, _ := args["context"].(string)

	// Get specialist
	_, err := dt.orchestrator.registry.GetSpecialist(specialistName)
	if err != nil {
		logger.WarnCF("agent", "Specialist not found", map[string]interface{}{
			"specialist": specialistName,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("specialist '%s' not found: %w", specialistName, err)
	}

	// Build the task message with context
	fullTask := task
	if contextStr != "" {
		fullTask = fmt.Sprintf("%s\n\nAdditional Context:\n%s", task, contextStr)
	}

	// Emit status: orchestrator analyzing
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:     "orchestrator",
		Status:    "analyzing",
		Timestamp: time.Now(),
	})

	// Emit status: delegating with reason
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:          "orchestrator",
		Status:         "delegating",
		SpecialistName: specialistName,
		Reason:         generateDelegationReason(specialistName, task),
		Timestamp:      time.Now(),
	})

	// Execute task through specialist
	logger.DebugCF("agent", "Delegating to specialist", map[string]interface{}{
		"specialist": specialistName,
		"task":       truncate(task, 100),
	})

	// Emit status: specialist working
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:     specialistName,
		Status:    "working",
		Timestamp: time.Now(),
	})

	result, err := dt.orchestrator.processSpecialistTask(ctx, specialistName, fullTask)
	if err != nil {
		logger.WarnCF("agent", "Specialist execution failed", map[string]interface{}{
			"specialist": specialistName,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("specialist execution failed: %w", err)
	}

	// Emit status: specialist complete
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:     specialistName,
		Status:    "complete",
		Timestamp: time.Now(),
	})

	logger.DebugCF("agent", "Specialist task completed", map[string]interface{}{
		"specialist": specialistName,
		"result":     truncate(result, 100),
	})

	return result, nil
}

// processSpecialistTask executes a task through a specialist agent
func (oa *OrchestratorAgent) processSpecialistTask(ctx context.Context, specialistName, task string) (string, error) {
	specialist, err := oa.registry.GetSpecialist(specialistName)
	if err != nil {
		return "", err
	}

	// Timeout configurable (default: 5 minutes)
	timeout := 5 * time.Minute
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute in goroutine to detect timeout
	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := specialist.ProcessWithSpeciality(ctxWithTimeout, task)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for result or timeout
	var result string
	select {
	case result = <-resultChan:
		// Success - emit content segment
		emitContentSegment(ctx, ContentSegment{
			Agent:     specialistName,
			Content:   result,
			SegmentID: fmt.Sprintf("seg_%s_%d", specialistName, time.Now().UnixNano()),
			Timestamp: time.Now(),
		})

	case err := <-errChan:
		return "", fmt.Errorf("specialist processing error: %w", err)

	case <-ctxWithTimeout.Done():
		// Emit timeout status
		emitAgentStatus(ctx, AgentStatusEvent{
			Agent:     specialistName,
			Status:    "timeout",
			Timestamp: time.Now(),
		})
		return "", fmt.Errorf("specialist %s timed out after %v", specialistName, timeout)
	}

	// Use tracker from context if available (web chat passes activeAgentLoop as tracker)
	// so that orchestrator/specialist names are registered on the correct loop instance.
	tracker := agentTrackerFromCtx(ctx)
	if tracker == nil {
		tracker = oa.AgentLoop
	}
	if len(tracker.GetInvolvedAgents()) == 0 {
		tracker.AddInvolvedAgent("orchestrator")
	}
	tracker.AddInvolvedAgent(specialistName)

	return result, nil
}

// GetSpecialistsSummary returns a summary of available specialists for the orchestrator's context
func (oa *OrchestratorAgent) GetSpecialistsSummary() string {
	specialists := oa.registry.ListSpecialists()
	if len(specialists) == 0 {
		return "No specialists available."
	}

	var summary strings.Builder
	summary.WriteString("Available Specialists:\n")

	for name, specialist := range specialists {
		if name == "orchestrator" {
			continue // Skip ourselves
		}

		tools := make([]string, 0)
		for tool := range specialist.allowedTools {
			tools = append(tools, tool)
		}

		summary.WriteString(fmt.Sprintf("- **%s**: %s\n", name, specialist.description))
		if len(tools) > 0 {
			summary.WriteString(fmt.Sprintf("  Tools: %v\n", tools))
		}
	}

	return summary.String()
}

// BuildOrchestratorContext creates the system prompt for the orchestrator
func (oa *OrchestratorAgent) BuildOrchestratorContext() string {
	var context strings.Builder

	context.WriteString("You are a Project Manager AI that analyzes incoming tasks and delegates them to specialized AI agents.\n\n")

	context.WriteString("## Your Role\n")
	context.WriteString("- Analyze the user's request to understand what type of work is needed\n")
	context.WriteString("- Match the task to the most appropriate specialist agent\n")
	context.WriteString("- Delegate the task using the delegate_to_specialist tool\n")
	context.WriteString("- Return the specialist's result to the user\n\n")

	context.WriteString("## Delegation Rules\n")
	context.WriteString("1. Always use the delegate_to_specialist tool to assign work\n")
	context.WriteString("2. Include relevant context when delegating\n")
	context.WriteString("3. If a specialist fails, try an alternative if available\n")
	context.WriteString("4. Always explain to the user which specialist handled their request\n\n")

	context.WriteString(oa.GetSpecialistsSummary())

	return context.String()
}

// generateDelegationReason creates a human-readable explanation for delegation
func generateDelegationReason(specialistName, task string) string {
	reasonMap := map[string]string{
		"developer":       "requires code implementation or modification",
		"documentation":   "requires documentation updates",
		"testing":         "requires test creation or QA",
		"devops":          "requires infrastructure work",
		"analyst":         "requires data analysis",
		"researcher":      "requires investigation",
	}

	if reason, ok := reasonMap[specialistName]; ok {
		return fmt.Sprintf("This task %s", reason)
	}

	return fmt.Sprintf("Delegating to %s specialist", specialistName)
}

// ProcessOrchestratorMessage processes a user message through the orchestrator
// The orchestrator analyzes the message and delegates to appropriate specialists
func (oa *OrchestratorAgent) ProcessOrchestratorMessage(ctx context.Context, userMessage string) (string, error) {
	// Build orchestrator-specific context
	orchestratorContext := oa.BuildOrchestratorContext()

	// Prepend context to user message
	fullMessage := orchestratorContext + "\n\nUser Request:\n" + userMessage

	// Use orchestrator's agent loop to process
	// This will handle delegations via the delegation tool
	result, err := oa.ProcessDirect(ctx, fullMessage, "orchestrator_session")
	if err != nil {
		return "", fmt.Errorf("orchestrator processing failed: %w", err)
	}

	return result, nil
}

// TimeUnixNano returns current time in nanoseconds since epoch
// Used for generating unique session keys
func TimeUnixNano() int64 {
	return time.Now().UnixNano()
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
