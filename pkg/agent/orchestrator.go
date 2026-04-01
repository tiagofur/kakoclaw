package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/providers"
	"github.com/sipeed/makoclaw/pkg/storage"
)

type agentTrackerKey struct{}
type sessionKeyCtxKey struct{}
type modelOverrideCtxKey struct{}

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

func contextWithSessionKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, sessionKeyCtxKey{}, key)
}

// ContextWithSessionKey embeds the user's session key into the context.
func ContextWithSessionKey(ctx context.Context, key string) context.Context {
	return contextWithSessionKey(ctx, key)
}

func sessionKeyFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(sessionKeyCtxKey{}).(string); ok {
		return v
	}
	return ""
}

func contextWithModelOverride(ctx context.Context, model string) context.Context {
	return context.WithValue(ctx, modelOverrideCtxKey{}, model)
}

// ContextWithModelOverride embeds a per-request model override into context so
// orchestrator delegations and direct specialist invocations can reuse it.
func ContextWithModelOverride(ctx context.Context, model string) context.Context {
	return contextWithModelOverride(ctx, model)
}

func modelOverrideFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(modelOverrideCtxKey{}).(string); ok {
		return v
	}
	return ""
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

// SpecialistReportCallback is called when a specialist completes work and reports status
type SpecialistReportCallback func(report *SpecialistReport) error

type specialistReportCallbackKey struct{}

// ContextWithSpecialistReportCallback embeds a SpecialistReportCallback into the context
func ContextWithSpecialistReportCallback(ctx context.Context, callback SpecialistReportCallback) context.Context {
	return context.WithValue(ctx, specialistReportCallbackKey{}, callback)
}

func specialistReportCallbackFromCtx(ctx context.Context) SpecialistReportCallback {
	if v, ok := ctx.Value(specialistReportCallbackKey{}).(SpecialistReportCallback); ok {
		return v
	}
	return nil
}

// emitSpecialistReport emits a specialist report if callback is available
func emitSpecialistReport(ctx context.Context, report *SpecialistReport) {
	if callback := specialistReportCallbackFromCtx(ctx); callback != nil {
		_ = callback(report)
	}
}

// DelegationUpdate context helpers
type delegationUpdateCallbackKey struct{}

func ContextWithDelegationUpdateCallback(ctx context.Context, callback DelegationUpdateCallback) context.Context {
	return context.WithValue(ctx, delegationUpdateCallbackKey{}, callback)
}

func delegationUpdateCallbackFromCtx(ctx context.Context) DelegationUpdateCallback {
	if v, ok := ctx.Value(delegationUpdateCallbackKey{}).(DelegationUpdateCallback); ok {
		return v
	}
	return nil
}

func emitDelegationUpdate(ctx context.Context, update DelegationUpdate) {
	if callback := delegationUpdateCallbackFromCtx(ctx); callback != nil {
		_ = callback(update)
	}
}

// ToolCallback context helpers — used by specialists to emit tool events to the WebSocket
type toolCallbackKey struct{}

// ContextWithToolCallback embeds a ToolCallback into the context so specialists can emit tool events
func ContextWithToolCallback(ctx context.Context, callback ToolCallback) context.Context {
	return context.WithValue(ctx, toolCallbackKey{}, callback)
}

// ToolCallbackFromCtx retrieves the ToolCallback from context
func ToolCallbackFromCtx(ctx context.Context) ToolCallback {
	if v, ok := ctx.Value(toolCallbackKey{}).(ToolCallback); ok {
		return v
	}
	return nil
}

// SpecialistTokenCallback is called for each streamed token produced by a specialist agent
type SpecialistTokenCallback func(agentName, token string) error

type specialistTokenCallbackKey struct{}

// ContextWithSpecialistTokenCallback embeds a SpecialistTokenCallback into the context
func ContextWithSpecialistTokenCallback(ctx context.Context, callback SpecialistTokenCallback) context.Context {
	return context.WithValue(ctx, specialistTokenCallbackKey{}, callback)
}

// SpecialistTokenCallbackFromCtx retrieves the SpecialistTokenCallback from context
func SpecialistTokenCallbackFromCtx(ctx context.Context) SpecialistTokenCallback {
	if v, ok := ctx.Value(specialistTokenCallbackKey{}).(SpecialistTokenCallback); ok {
		return v
	}
	return nil
}

// Delegation chain context helpers
type delegationChainKey struct{}

func contextWithDelegationChain(ctx context.Context, chain []string) context.Context {
	return context.WithValue(ctx, delegationChainKey{}, chain)
}

func delegationChainFromCtx(ctx context.Context) []string {
	if v, ok := ctx.Value(delegationChainKey{}).([]string); ok {
		result := make([]string, len(v))
		copy(result, v)
		return result
	}
	return []string{"orchestrator"}
}

// SpecialistStreamEvent is emitted for each token produced by a specialist during streaming
type SpecialistStreamEvent struct {
	SpecialistName string    `json:"specialist_name"`
	Token          string    `json:"token"`
	Timestamp      time.Time `json:"timestamp"`
}

// OrchestratorDelegationEvent is emitted when the orchestrator delegates a task to a specialist
type OrchestratorDelegationEvent struct {
	SpecialistName string    `json:"specialist_name"`
	Task           string    `json:"task"`           // full prompt sent to specialist
	DelegationID   string    `json:"delegation_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// SynthesisEvent is emitted around the orchestrator's final synthesis pass
type SynthesisEvent struct {
	Phase     string    `json:"phase"`           // "start" | "token" | "end"
	Token     string    `json:"token,omitempty"` // only when Phase == "token"
	Timestamp time.Time `json:"timestamp"`
}

// ReportSavedEvent is emitted when a specialist saves a report artifact
type ReportSavedEvent struct {
	Title     string    `json:"title"`
	FilePath  string    `json:"file_path"`
	Summary   string    `json:"summary"`
	Timestamp time.Time `json:"timestamp"`
}

// SpecialistStreamCallback is called for each token produced by a specialist
type SpecialistStreamCallback func(event SpecialistStreamEvent) error

type specialistStreamCallbackKey struct{}

// ContextWithSpecialistStreamCallback embeds a SpecialistStreamCallback into the context
func ContextWithSpecialistStreamCallback(ctx context.Context, callback SpecialistStreamCallback) context.Context {
	return context.WithValue(ctx, specialistStreamCallbackKey{}, callback)
}

func specialistStreamCallbackFromCtx(ctx context.Context) SpecialistStreamCallback {
	if v, ok := ctx.Value(specialistStreamCallbackKey{}).(SpecialistStreamCallback); ok {
		return v
	}
	return nil
}

// emitSpecialistStream emits a SpecialistStreamEvent if callback is available
func emitSpecialistStream(ctx context.Context, specialistName, token string) {
	if cb := specialistStreamCallbackFromCtx(ctx); cb != nil {
		_ = cb(SpecialistStreamEvent{
			SpecialistName: specialistName,
			Token:          token,
			Timestamp:      time.Now(),
		})
	}
}

// OrchestratorDelegationCallback is called when the orchestrator delegates to a specialist
type OrchestratorDelegationCallback func(event OrchestratorDelegationEvent) error

type orchestratorDelegationCallbackKey struct{}

// ContextWithOrchestratorDelegationCallback embeds an OrchestratorDelegationCallback into the context
func ContextWithOrchestratorDelegationCallback(ctx context.Context, callback OrchestratorDelegationCallback) context.Context {
	return context.WithValue(ctx, orchestratorDelegationCallbackKey{}, callback)
}

func orchestratorDelegationCallbackFromCtx(ctx context.Context) OrchestratorDelegationCallback {
	if v, ok := ctx.Value(orchestratorDelegationCallbackKey{}).(OrchestratorDelegationCallback); ok {
		return v
	}
	return nil
}

// emitOrchestratorDelegation emits an OrchestratorDelegationEvent if callback is available
func emitOrchestratorDelegation(ctx context.Context, event OrchestratorDelegationEvent) {
	if cb := orchestratorDelegationCallbackFromCtx(ctx); cb != nil {
		_ = cb(event)
	}
}

// SynthesisCallback is called during the orchestrator's final synthesis pass
type SynthesisCallback func(event SynthesisEvent) error

type synthesisCallbackKey struct{}

// ContextWithSynthesisCallback embeds a SynthesisCallback into the context
func ContextWithSynthesisCallback(ctx context.Context, callback SynthesisCallback) context.Context {
	return context.WithValue(ctx, synthesisCallbackKey{}, callback)
}

func synthesisCallbackFromCtx(ctx context.Context) SynthesisCallback {
	if v, ok := ctx.Value(synthesisCallbackKey{}).(SynthesisCallback); ok {
		return v
	}
	return nil
}

// emitSynthesis emits a SynthesisEvent if callback is available
func emitSynthesis(ctx context.Context, phase, token string) {
	if cb := synthesisCallbackFromCtx(ctx); cb != nil {
		_ = cb(SynthesisEvent{
			Phase:     phase,
			Token:     token,
			Timestamp: time.Now(),
		})
	}
}

// ReportSavedCallback is called when a report artifact is saved by a specialist
type ReportSavedCallback func(event ReportSavedEvent) error

type reportSavedCallbackKey struct{}

// ContextWithReportSavedCallback embeds a ReportSavedCallback into the context
func ContextWithReportSavedCallback(ctx context.Context, callback ReportSavedCallback) context.Context {
	return context.WithValue(ctx, reportSavedCallbackKey{}, callback)
}

func reportSavedCallbackFromCtx(ctx context.Context) ReportSavedCallback {
	if v, ok := ctx.Value(reportSavedCallbackKey{}).(ReportSavedCallback); ok {
		return v
	}
	return nil
}

// emitReportSaved emits a ReportSavedEvent if callback is available
func emitReportSaved(ctx context.Context, event ReportSavedEvent) {
	if cb := reportSavedCallbackFromCtx(ctx); cb != nil {
		_ = cb(event)
	}
}

// OrchestratorAgent is a special agent that analyzes tasks and delegates to specialists
type OrchestratorAgent struct {
	*SpecialistAgent
	registry          *SpecialistRegistry
	delegationRetries int
	fallbackToDefault bool

	// Specialist report tracking for feedback loop
	lastReport       *SpecialistReport
	lastReportMu     sync.RWMutex
	currentReports   []SpecialistReport
	currentReportsMu sync.RWMutex
	currentTeamCtx   *TeamContext
	currentTeamCtxMu sync.RWMutex

	// Delegation limit tracking (per message)
	delegationCount   int
	delegationCountMu sync.Mutex
	maxDelegations    int // default: 3
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

// SpecialistReport represents structured feedback from a specialist
// This enables the orchestrator to make intelligent re-delegation decisions
type SpecialistReport struct {
	SpecialistName  string    `json:"specialist_name"`            // Who generated this report
	Status          string    `json:"status"`                     // "complete", "needs_help", "partial", "failed"
	Result          string    `json:"result"`                     // The actual output
	Confidence      float64   `json:"confidence"`                 // 0.0-1.0 confidence in result
	Suggestions     []string  `json:"suggestions"`                // Recommendations for improvement
	NeedsReview     bool      `json:"needs_review"`               // Request orchestrator review
	RequestHelp     string    `json:"request_help"`               // Request another specialist (e.g., "security")
	HelpContext     string    `json:"help_context"`               // Context for the help request
	Artifacts       []string  `json:"artifacts"`                  // Files created/modified
	TimeSpent       string    `json:"time_spent"`                 // Execution time
	Iteration       int       `json:"iteration"`                  // Which iteration this is from
	Timestamp       time.Time `json:"timestamp"`                  // When the report was generated
	DelegationChain []string  `json:"delegation_chain,omitempty"` // delegation path
	DelegationDepth int       `json:"delegation_depth,omitempty"` // depth in chain
	ToolsUsed       []string  `json:"tools_used,omitempty"`       // tools the specialist called
	IterationsUsed  int       `json:"iterations_used,omitempty"`  // LLM iterations consumed
}

// TeamContext holds shared state for multi-specialist collaboration
type TeamContext struct {
	mu             sync.Mutex        `json:"-"`
	TaskID         string            `json:"task_id"`
	OriginalTask   string            `json:"original_task"`
	UserUUID       string            `json:"user_uuid,omitempty"`
	UserID         int64             `json:"user_id,omitempty"`
	Subtasks       []Subtask         `json:"subtasks"`
	Progress       map[string]string `json:"progress"`       // specialist -> status
	SharedNotes    []TeamNote        `json:"shared_notes"`   // cross-specialist notes
	Artifacts      []string          `json:"artifacts"`      // files created
	Decisions      []TeamDecision    `json:"decisions"`      // key decisions made
	Communications []TeamComm        `json:"communications"` // inter-specialist messages
}

// Subtask represents a decomposed part of a larger task
type Subtask struct {
	ID           int    `json:"id"`
	Description  string `json:"description"`
	Specialist   string `json:"specialist"`
	Dependencies []int  `json:"dependencies"` // IDs of prerequisite subtasks
	Priority     int    `json:"priority"`
	Status       string `json:"status"` // "pending", "in_progress", "complete", "failed"
	Result       string `json:"result"`
}

// TeamNote represents a shared note between specialists
type TeamNote struct {
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	ForAgent  string    `json:"for_agent,omitempty"` // if directed at specific agent
}

// TeamDecision records a decision made during task execution
type TeamDecision struct {
	Author     string    `json:"author"`
	Decision   string    `json:"decision"`
	Rationale  string    `json:"rationale"`
	Timestamp  time.Time `json:"timestamp"`
	AffectedBy []string  `json:"affected_by"` // other agents affected
}

// TeamComm represents communication between specialists
type TeamComm struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // "request", "response", "info"
	Timestamp time.Time `json:"timestamp"`
}

// Context key for team context
type teamContextKey struct{}

// ContextWithTeamContext embeds TeamContext into the context
func ContextWithTeamContext(ctx context.Context, tc *TeamContext) context.Context {
	return context.WithValue(ctx, teamContextKey{}, tc)
}

// TeamContextFromCtx retrieves TeamContext from context
func TeamContextFromCtx(ctx context.Context) *TeamContext {
	if v, ok := ctx.Value(teamContextKey{}).(*TeamContext); ok {
		return v
	}
	return nil
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
		Tools:             []string{"delegate_to_specialist", "decompose_task"}, // Orchestrator delegation tools
	}

	// Create the specialist agent wrapper
	specialist, err := NewSpecialistAgent("orchestrator", &specCfg, cfg, msgBus, baseProvider, store)
	if err != nil {
		return nil, fmt.Errorf("failed to create orchestrator specialist: %w", err)
	}

	maxDel := cfg.Agents.Orchestrator.MaxDelegationsPerMessage
	if maxDel <= 0 {
		maxDel = 3
	}
	orchestrator := &OrchestratorAgent{
		SpecialistAgent:   specialist,
		registry:          registry,
		delegationRetries: cfg.Agents.Orchestrator.MaxDelegationRetries,
		fallbackToDefault: cfg.Agents.Orchestrator.FallbackToDefault,
		maxDelegations:    maxDel,
	}

	// Register the delegation tool
	orchestrator.registerDelegationTool()

	// Register built-in "general" specialist for fallback (if not already configured)
	if _, err := registry.GetSpecialist("general"); err != nil {
		generalCfg := &config.SpecialistConfig{
			Name:        "general",
			Description: "General-purpose agent with full capabilities. Handles tasks when no specialist matches.",
			Tools:       []string{}, // Empty = all tools available
		}
		generalSpec, gErr := NewSpecialistAgent("general", generalCfg, cfg, msgBus, baseProvider, store)
		if gErr == nil {
			registry.RegisterSpecialist(generalSpec)
			logger.InfoC("agent", "Registered built-in 'general' specialist for fallback")
		}
	}

	// Inject orchestrator context into the system prompt so the LLM sees
	// the specialist catalog and delegation rules via the standard BuildSystemPrompt() path.
	orchestrator.contextBuilder.SetAgentSystemPrompt(orchestrator.BuildOrchestratorContext())
	// Orchestrator needs skill-creator so it can inform users about skill creation
	orchestrator.contextBuilder.SetSkillFilter([]string{"skill-creator"})
	// Lean context: skip bootstrap files and memory
	orchestrator.contextBuilder.SetLightweightMode(true)
	orchestrator.AgentLoop.orchestratorOwner = orchestrator

	logger.InfoCF("agent", "Orchestrator agent created", map[string]interface{}{
		"provider":    cfg.Agents.Orchestrator.Provider,
		"model":       cfg.Agents.Orchestrator.Model,
		"specialists": len(registry.ListSpecialists()),
	})

	return orchestrator, nil
}

// registerDelegationTool registers the delegation and decomposition tools in the orchestrator's tool registry
func (oa *OrchestratorAgent) registerDelegationTool() {
	delegateTool := &DelegationTool{
		orchestrator: oa,
	}
	oa.tools.Register(delegateTool)

	// Register task decomposition tool for breaking complex tasks into subtasks
	decomposeTool := &TaskDecompositionTool{
		orchestrator: oa,
	}
	oa.tools.Register(decomposeTool)
}

// ResetDelegationCount resets the per-message delegation counter. Call before processing each new user message.
func (oa *OrchestratorAgent) ResetDelegationCount() {
	oa.delegationCountMu.Lock()
	oa.delegationCount = 0
	oa.delegationCountMu.Unlock()
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

	// Check delegation limit
	dt.orchestrator.delegationCountMu.Lock()
	dt.orchestrator.delegationCount++
	count := dt.orchestrator.delegationCount
	dt.orchestrator.delegationCountMu.Unlock()

	maxDel := dt.orchestrator.maxDelegations
	if maxDel <= 0 {
		maxDel = 3
	}

	if count > maxDel {
		emitAgentStatus(ctx, AgentStatusEvent{
			Agent:     "orchestrator",
			Status:    "max_delegations_reached",
			Reason:    fmt.Sprintf("Reached maximum %d delegations for this message. Synthesizing response with available results.", maxDel),
			Timestamp: time.Now(),
		})
		return fmt.Sprintf("[Maximum delegation limit (%d) reached. Please synthesize a response using the information gathered so far.]", maxDel), nil
	}

	// Get specialist (with fallback to general if enabled)
	_, err := dt.orchestrator.registry.GetSpecialist(specialistName)
	if err != nil {
		if dt.orchestrator.fallbackToDefault {
			// Fallback to "general" agent
			originalName := specialistName
			specialistName = "general"

			emitAgentStatus(ctx, AgentStatusEvent{
				Agent:           "orchestrator",
				Status:          "fallback",
				SpecialistName:  "general",
				Reason:          fmt.Sprintf("No specialist '%s' found, using general agent", originalName),
				DelegationChain: delegationChainFromCtx(ctx),
				DelegationDepth: 0,
				Timestamp:       time.Now(),
			})

			logger.InfoCF("agent", "Falling back to general agent", map[string]interface{}{
				"requested": originalName,
			})
		} else {
			logger.WarnCF("agent", "Specialist not found", map[string]interface{}{
				"specialist": specialistName,
				"error":      err.Error(),
			})
			return "", fmt.Errorf("specialist '%s' not found: %w", specialistName, err)
		}
	}

	// Build the task message with context
	fullTask := task
	if contextStr != "" {
		fullTask = fmt.Sprintf("%s\n\nAdditional Context:\n%s", task, contextStr)
	}

	// Emit status: orchestrator analyzing
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:           "orchestrator",
		Status:          "analyzing",
		DelegationChain: delegationChainFromCtx(ctx),
		DelegationDepth: 0,
		Timestamp:       time.Now(),
	})

	// Emit status: delegating with reason
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:           "orchestrator",
		Status:          "delegating",
		SpecialistName:  specialistName,
		Reason:          generateDelegationReason(specialistName, task),
		DelegationChain: delegationChainFromCtx(ctx),
		DelegationDepth: 0,
		Timestamp:       time.Now(),
	})

	// Execute task through specialist
	logger.DebugCF("agent", "Delegating to specialist", map[string]interface{}{
		"specialist": specialistName,
		"task":       truncate(task, 100),
	})

	result, err := dt.orchestrator.processSpecialistTask(ctx, specialistName, fullTask, sessionKeyFromCtx(ctx))
	if err != nil {
		logger.WarnCF("agent", "Specialist execution failed", map[string]interface{}{
			"specialist": specialistName,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("specialist execution failed: %w", err)
	}

	// Emit status: specialist complete with chain info
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:           specialistName,
		Status:          "complete",
		DelegationChain: append(delegationChainFromCtx(ctx), specialistName),
		DelegationDepth: 1,
		ParentAgent:     "orchestrator",
		Timestamp:       time.Now(),
	})

	logger.DebugCF("agent", "Specialist task completed", map[string]interface{}{
		"specialist": specialistName,
		"result":     truncate(result, 100),
	})

	return result, nil
}

// TaskDecompositionTool breaks complex tasks into subtasks with specialist assignments
type TaskDecompositionTool struct {
	orchestrator *OrchestratorAgent
}

func (tdt *TaskDecompositionTool) Name() string {
	return "decompose_task"
}

func (tdt *TaskDecompositionTool) Description() string {
	return "Breaks down a complex task into subtasks with specialist assignments and dependencies. Use this for tasks that require multiple specialists or have sequential steps."
}

func (tdt *TaskDecompositionTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task": map[string]interface{}{
				"type":        "string",
				"description": "The complex task to decompose into subtasks",
			},
			"subtasks": map[string]interface{}{
				"type":        "array",
				"description": "List of subtasks with their specialist assignments",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Description of the subtask",
						},
						"specialist": map[string]interface{}{
							"type":        "string",
							"description": "Name of the specialist to assign this subtask to",
						},
						"dependencies": map[string]interface{}{
							"type":        "array",
							"description": "Indices (0-based) of subtasks that must complete before this one",
							"items":       map[string]interface{}{"type": "integer"},
						},
						"priority": map[string]interface{}{
							"type":        "integer",
							"description": "Priority level (1=highest, 5=lowest)",
						},
					},
					"required": []string{"description", "specialist"},
				},
			},
		},
		"required": []string{"task", "subtasks"},
	}
}

func (tdt *TaskDecompositionTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	task, ok := args["task"].(string)
	if !ok {
		return "", fmt.Errorf("task must be a string")
	}

	subtasksRaw, ok := args["subtasks"].([]interface{})
	if !ok {
		return "", fmt.Errorf("subtasks must be an array")
	}

	// Get or create team context
	teamCtx := TeamContextFromCtx(ctx)
	userUUID := ""
	var userID int64
	if tdt.orchestrator != nil && tdt.orchestrator.AgentLoop != nil {
		userUUID = tdt.orchestrator.AgentLoop.userUUID
		userID = tdt.orchestrator.AgentLoop.userID
	}
	if teamCtx == nil {
		teamCtx = &TeamContext{
			TaskID:         fmt.Sprintf("task_%d", time.Now().UnixNano()),
			OriginalTask:   task,
			UserUUID:       userUUID,
			UserID:         userID,
			Progress:       make(map[string]string),
			SharedNotes:    []TeamNote{},
			Artifacts:      []string{},
			Decisions:      []TeamDecision{},
			Communications: []TeamComm{},
		}
		ctx = ContextWithTeamContext(ctx, teamCtx)
	}
	if teamCtx.UserUUID == "" {
		teamCtx.UserUUID = userUUID
	}
	if teamCtx.UserID == 0 {
		teamCtx.UserID = userID
	}

	// Parse subtasks
	var subtasks []Subtask
	for i, stRaw := range subtasksRaw {
		st, ok := stRaw.(map[string]interface{})
		if !ok {
			continue
		}

		subtask := Subtask{
			ID:       i,
			Status:   "pending",
			Priority: 3, // default medium priority
		}

		if desc, ok := st["description"].(string); ok {
			subtask.Description = desc
		}
		if spec, ok := st["specialist"].(string); ok {
			subtask.Specialist = spec
		}
		if deps, ok := st["dependencies"].([]interface{}); ok {
			for _, d := range deps {
				if dInt, ok := d.(float64); ok {
					subtask.Dependencies = append(subtask.Dependencies, int(dInt))
				}
			}
		}
		if prio, ok := st["priority"].(float64); ok {
			subtask.Priority = int(prio)
		}

		subtasks = append(subtasks, subtask)
	}

	teamCtx.Subtasks = subtasks

	// Emit status
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:     "orchestrator",
		Status:    "analyzing",
		Reason:    fmt.Sprintf("Decomposed task into %d subtasks", len(subtasks)),
		Timestamp: time.Now(),
	})

	// Execute subtasks respecting dependencies
	results := make([]string, len(subtasks))
	completed := make(map[int]bool)

	for len(completed) < len(subtasks) {
		// Find next executable subtasks (all dependencies satisfied)
		var ready []int
		for i, st := range subtasks {
			if completed[i] {
				continue
			}
			canRun := true
			for _, dep := range st.Dependencies {
				if !completed[dep] {
					canRun = false
					break
				}
			}
			if canRun {
				ready = append(ready, i)
			}
		}

		if len(ready) == 0 && len(completed) < len(subtasks) {
			return "", fmt.Errorf("circular dependency detected in subtasks")
		}

		// Execute ready subtasks (could parallelize in future)
		for _, idx := range ready {
			st := &subtasks[idx]
			st.Status = "in_progress"
			teamCtx.mu.Lock()
			teamCtx.Progress[st.Specialist] = "working"
			teamCtx.mu.Unlock()

			// Build context from previous results
			var contextParts []string
			for _, dep := range st.Dependencies {
				if results[dep] != "" {
					contextParts = append(contextParts, fmt.Sprintf("Previous step result:\n%s", results[dep]))
				}
			}
			contextStr := strings.Join(contextParts, "\n\n")

			// Emit delegating status
			emitAgentStatus(ctx, AgentStatusEvent{
				Agent:          "orchestrator",
				Status:         "delegating",
				SpecialistName: st.Specialist,
				Reason:         fmt.Sprintf("Subtask %d/%d: %s", idx+1, len(subtasks), truncate(st.Description, 50)),
				Timestamp:      time.Now(),
			})

			// Delegate to specialist
			fullTask := st.Description
			if contextStr != "" {
				fullTask = fmt.Sprintf("%s\n\nContext from previous steps:\n%s", st.Description, contextStr)
			}

			// Emit working status
			emitAgentStatus(ctx, AgentStatusEvent{
				Agent:     st.Specialist,
				Status:    "working",
				Reason:    truncate(st.Description, 50),
				Timestamp: time.Now(),
			})

			result, err := tdt.orchestrator.processSpecialistTask(ctx, st.Specialist, fullTask, sessionKeyFromCtx(ctx))
			if err != nil {
				st.Status = "failed"
				teamCtx.mu.Lock()
				teamCtx.Progress[st.Specialist] = "failed"
				teamCtx.mu.Unlock()
				logger.WarnCF("agent", "Subtask failed", map[string]interface{}{
					"subtask":    idx,
					"specialist": st.Specialist,
					"error":      err.Error(),
				})
				// Continue with other subtasks
				results[idx] = fmt.Sprintf("Error: %s", err.Error())
			} else {
				st.Status = "complete"
				st.Result = result
				teamCtx.mu.Lock()
				teamCtx.Progress[st.Specialist] = "complete"
				teamCtx.mu.Unlock()
				results[idx] = result
			}

			// Emit completion status
			emitAgentStatus(ctx, AgentStatusEvent{
				Agent:     st.Specialist,
				Status:    st.Status,
				Timestamp: time.Now(),
			})

			completed[idx] = true
		}
	}

	// Build aggregated result
	var resultBuilder strings.Builder
	resultBuilder.WriteString("## Task Decomposition Complete\n\n")
	resultBuilder.WriteString(fmt.Sprintf("Original task: %s\n\n", truncate(task, 100)))
	resultBuilder.WriteString(fmt.Sprintf("### Subtask Results (%d total)\n\n", len(subtasks)))

	for i, st := range subtasks {
		statusEmoji := "✅"
		if st.Status == "failed" {
			statusEmoji = "❌"
		}
		resultBuilder.WriteString(fmt.Sprintf("%s **%d. %s** (%s)\n", statusEmoji, i+1, st.Specialist, st.Description))
		if results[i] != "" {
			resultBuilder.WriteString(fmt.Sprintf("   Result: %s\n\n", truncate(results[i], 200)))
		}
	}

	return resultBuilder.String(), nil
}

// processSpecialistTask executes a task through a specialist agent.
// sessionKey preserves the caller's chat session when available.
func (oa *OrchestratorAgent) processSpecialistTask(ctx context.Context, specialistName, task, sessionKey string) (string, error) {
	specialist, err := oa.registry.GetSpecialist(specialistName)
	if err != nil {
		return "", err
	}

	// Build delegation chain from parent context (defensive copy to avoid slice aliasing)
	parentChain := delegationChainFromCtx(ctx)
	currentChain := make([]string, len(parentChain)+1)
	copy(currentChain, parentChain)
	currentChain[len(parentChain)] = specialistName
	currentDepth := len(parentChain) // orchestrator=0, first specialist=1, colleague=2
	parentAgent := "orchestrator"
	if len(parentChain) > 0 {
		parentAgent = parentChain[len(parentChain)-1]
	}

	// Generate a unique delegation ID
	delegationID := fmt.Sprintf("del_%s_%d", specialistName, time.Now().UnixNano())
	startTime := time.Now()

	// Store chain in context for nested delegations
	specialistCtx := contextWithDelegationChain(ctx, currentChain)

	// Inject response format instruction for structured feedback
	taskWithFormat := task + `

--- MANDATORY RESPONSE FORMAT ---
CRITICAL RULES:
1. You have a LIMITED number of tool iterations. Do NOT waste them.
2. Once you have enough information, STOP calling tools IMMEDIATELY and provide your final text response.
3. Do NOT loop — if you called the same tool twice with similar args, STOP and respond.
4. If a tool call fails, do NOT retry more than once — respond with what you have.
5. Your response MUST be a TEXT message (no more tool calls).

Your final response MUST contain TWO parts:

1. A JSON report block enclosed in markdown JSON tags:
` + "```json\n" + `{"status":"complete","confidence":0.9,"request_help":"","suggestion":""}
` + "```" + `
(Status: "complete", "partial", or "needs_help". Confidence: 0.0-1.0)

2. YOUR FULL DETAILED RESPONSE below the JSON block.
The user only sees text below the JSON block — include all findings there.
---`

	// Timeout from config (default: 5 minutes)
	timeoutSecs := oa.cfg.Agents.Orchestrator.SpecialistTimeoutSeconds
	if timeoutSecs <= 0 {
		timeoutSecs = 300
	}
	timeout := time.Duration(timeoutSecs) * time.Second
	ctxWithTimeout, cancel := context.WithTimeout(specialistCtx, timeout)
	defer cancel()
	effectiveSessionKey := sessionKey
	if effectiveSessionKey == "" {
		effectiveSessionKey = sessionKeyFromCtx(ctxWithTimeout)
	}

	// Emit delegation start with chain info (including active skills)
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:           specialistName,
		Status:          "working",
		DelegationChain: currentChain,
		DelegationDepth: currentDepth,
		ParentAgent:     parentAgent,
		ActiveSkills:    specialist.skills,
		MaxIterations:   specialist.maxIterations,
		Timestamp:       time.Now(),
	})

	emitDelegationUpdate(ctx, DelegationUpdate{
		DelegationID:  delegationID,
		From:          parentAgent,
		To:            specialistName,
		Status:        "started",
		Iteration:     0,
		MaxIterations: oa.delegationRetries,
		ElapsedMs:     0,
		Timestamp:     time.Now(),
	})

	// Emit orchestrator delegation event with full task sent to specialist
	emitOrchestratorDelegation(ctx, OrchestratorDelegationEvent{
		SpecialistName: specialistName,
		Task:           taskWithFormat,
		DelegationID:   delegationID,
		Timestamp:      time.Now(),
	})

	// Bridge SpecialistStreamCallback → SpecialistTokenCallback so specialist tokens
	// are forwarded as SpecialistStreamEvents to whoever registered the stream callback.
	streamCtx := ctxWithTimeout
	if streamCb := specialistStreamCallbackFromCtx(ctx); streamCb != nil {
		streamCtx = ContextWithSpecialistTokenCallback(ctxWithTimeout, func(agentName, token string) error {
			return streamCb(SpecialistStreamEvent{
				SpecialistName: agentName,
				Token:          token,
				Timestamp:      time.Now(),
			})
		})
	}

	// Execute in goroutine to detect timeout
	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := specialist.ProcessWithSpecialityForUserStream(streamCtx, oa.userUUID, oa.userID, taskWithFormat, effectiveSessionKey)
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
		elapsed := time.Since(startTime).Milliseconds()

		// Parse structured report from specialist response
		report := oa.parseSpecialistReport(specialistName, result)
		cleanResult := oa.cleanSpecialistResult(result)
		report.Result = cleanResult
		// Enrich report with chain metadata
		report.DelegationChain = currentChain
		report.DelegationDepth = currentDepth
		oa.appendCurrentReport(*report)
		oa.storeLastReport(report)

		// Emit specialist report event for frontend visibility
		emitSpecialistReport(ctx, report)

		// Log report details
		logger.InfoCF("agent", "Specialist report received", map[string]interface{}{
			"specialist": specialistName,
			"status":     report.Status,
			"confidence": report.Confidence,
			"help":       report.RequestHelp,
			"chain":      currentChain,
			"depth":      currentDepth,
		})

		// Emit content segment with clean result
		emitContentSegment(ctx, ContentSegment{
			Agent:     specialistName,
			Content:   cleanResult,
			SegmentID: fmt.Sprintf("seg_%s_%d", specialistName, time.Now().UnixNano()),
			Timestamp: time.Now(),
		})

		// Emit delegation complete
		emitDelegationUpdate(ctx, DelegationUpdate{
			DelegationID:  delegationID,
			From:          parentAgent,
			To:            specialistName,
			Status:        "complete",
			Iteration:     report.Iteration,
			MaxIterations: oa.delegationRetries,
			ElapsedMs:     elapsed,
			Timestamp:     time.Now(),
		})

		// Persist specialist result to user's original session for visibility
		if effectiveSessionKey != "" && oa.sessions != nil {
			summaryMsg := fmt.Sprintf("[Agent @%s]\n%s", specialistName, cleanResult)
			oa.sessions.AddMessage(effectiveSessionKey, "assistant", summaryMsg)
		}

		// Handle empty/insufficient response
		if len(strings.TrimSpace(cleanResult)) < 10 {
			logger.WarnCF("agent", "Specialist returned insufficient response", map[string]interface{}{
				"specialist": specialistName,
				"length":     len(cleanResult),
			})
			cleanResult = fmt.Sprintf("[Specialist %s returned an insufficient response. Confidence: %.0f%%. Consider delegating to another specialist or providing more context.]",
				specialistName, report.Confidence*100)
		}

		// Update result to clean version for return
		result = cleanResult

	case err := <-errChan:
		emitDelegationUpdate(ctx, DelegationUpdate{
			DelegationID: delegationID,
			From:         parentAgent,
			To:           specialistName,
			Status:       "error",
			ElapsedMs:    time.Since(startTime).Milliseconds(),
			Timestamp:    time.Now(),
		})
		return "", fmt.Errorf("specialist processing error: %w", err)

	case <-ctxWithTimeout.Done():
		// Emit timeout status with chain info
		emitAgentStatus(ctx, AgentStatusEvent{
			Agent:           specialistName,
			Status:          "timeout",
			DelegationChain: currentChain,
			DelegationDepth: currentDepth,
			ParentAgent:     parentAgent,
			Timestamp:       time.Now(),
		})
		emitDelegationUpdate(ctx, DelegationUpdate{
			DelegationID: delegationID,
			From:         parentAgent,
			To:           specialistName,
			Status:       "error",
			ElapsedMs:    time.Since(startTime).Milliseconds(),
			Timestamp:    time.Now(),
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

func (oa *OrchestratorAgent) aggregateCurrentExecutionReports(teamCtx *TeamContext) ([]SpecialistReport, string) {
	reports := oa.currentExecutionReports()
	if len(reports) == 0 {
		return nil, ""
	}

	return reports, oa.aggregateReports(reports, teamCtx)
}

func (oa *OrchestratorAgent) resetCurrentExecutionState(task, userUUID string, userID int64) *TeamContext {
	teamCtx := &TeamContext{
		TaskID:         fmt.Sprintf("task_%d", time.Now().UnixNano()),
		OriginalTask:   task,
		UserUUID:       userUUID,
		UserID:         userID,
		Progress:       make(map[string]string),
		SharedNotes:    []TeamNote{},
		Artifacts:      []string{},
		Decisions:      []TeamDecision{},
		Communications: []TeamComm{},
	}

	oa.currentReportsMu.Lock()
	oa.currentReports = nil
	oa.currentReportsMu.Unlock()

	oa.currentTeamCtxMu.Lock()
	oa.currentTeamCtx = teamCtx
	oa.currentTeamCtxMu.Unlock()

	return teamCtx
}

func (oa *OrchestratorAgent) appendCurrentReport(report SpecialistReport) {
	oa.currentReportsMu.Lock()
	defer oa.currentReportsMu.Unlock()
	oa.currentReports = append(oa.currentReports, report)
}

func (oa *OrchestratorAgent) currentExecutionReports() []SpecialistReport {
	oa.currentReportsMu.RLock()
	defer oa.currentReportsMu.RUnlock()
	result := make([]SpecialistReport, len(oa.currentReports))
	copy(result, oa.currentReports)
	return result
}

func specialistSegmentsFromReports(reports []SpecialistReport) []map[string]interface{} {
	segments := make([]map[string]interface{}, 0, len(reports))
	for _, report := range reports {
		if strings.TrimSpace(report.Result) == "" {
			continue
		}
		segments = append(segments, map[string]interface{}{
			"specialist": report.SpecialistName,
			"status":     report.Status,
			"content":    report.Result,
			"confidence": report.Confidence,
		})
	}
	return segments
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
	context.WriteString("4. Always explain to the user which specialist handled their request\n")
	context.WriteString("5. If no specialist matches the task, delegate to 'general'\n")
	context.WriteString("6. The 'general' agent has full capabilities and handles any task\n\n")

	context.WriteString("## Response Format\n")
	context.WriteString("- Start with a brief summary of what was done and by whom\n")
	context.WriteString("- Present the specialist's work directly - do not rewrite it significantly\n")
	context.WriteString("- If multiple specialists contributed, label each section with **[specialist_name]:**\n")
	context.WriteString("- Keep your own commentary brief - the specialist's work is the main content\n")
	context.WriteString("- NEVER list or display the available specialists to the user. The specialist catalog is for your internal decision-making only.\n")
	context.WriteString("- Do not echo specialist names, descriptions, or tools in your responses unless referencing who handled a specific task.\n\n")

	context.WriteString(oa.GetSpecialistsSummary())

	return context.String()
}

// generateDelegationReason creates a human-readable explanation for delegation
func generateDelegationReason(specialistName, task string) string {
	reasonMap := map[string]string{
		"developer":     "requires code implementation or modification",
		"documentation": "requires documentation updates",
		"testing":       "requires test creation or QA",
		"devops":        "requires infrastructure work",
		"analyst":       "requires data analysis",
		"researcher":    "requires investigation",
		"general":       "general-purpose task with full capabilities",
	}

	if reason, ok := reasonMap[specialistName]; ok {
		return fmt.Sprintf("This task %s", reason)
	}

	return fmt.Sprintf("Delegating to %s specialist", specialistName)
}

// ProcessOrchestratorMessage processes a user message through the orchestrator
// The orchestrator analyzes the message and delegates to appropriate specialists
func (oa *OrchestratorAgent) ProcessOrchestratorMessage(ctx context.Context, userMessage string) (string, error) {
	// Orchestrator context is already injected via SetAgentSystemPrompt (at init time).
	// Pass the user message directly to avoid the LLM echoing the specialist catalog.
	result, err := oa.ProcessDirect(ctx, userMessage, "orchestrator_session")
	if err != nil {
		return "", fmt.Errorf("orchestrator processing failed: %w", err)
	}

	return result, nil
}

// ProcessWithFeedbackLoop processes a user message with iterative feedback and re-delegation
// This enables true team collaboration where specialists can report back and request help
func (oa *OrchestratorAgent) ProcessWithFeedbackLoop(ctx context.Context, userMessage string) (string, error) {
	maxIterations := oa.cfg.Agents.Orchestrator.FeedbackLoopIterations
	if maxIterations <= 0 {
		maxIterations = 3
	}
	var reports []SpecialistReport

	// Initialize team context for this task
	teamCtx := &TeamContext{
		TaskID:         fmt.Sprintf("task_%d", time.Now().UnixNano()),
		OriginalTask:   userMessage,
		UserUUID:       oa.AgentLoop.userUUID,
		UserID:         oa.AgentLoop.userID,
		Progress:       make(map[string]string),
		SharedNotes:    []TeamNote{},
		Artifacts:      []string{},
		Decisions:      []TeamDecision{},
		Communications: []TeamComm{},
	}
	ctx = ContextWithTeamContext(ctx, teamCtx)

	// Emit initial status
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:     "orchestrator",
		Status:    "analyzing",
		Reason:    "Analyzing task and planning delegation strategy",
		Timestamp: time.Now(),
	})

	for iteration := 0; iteration < maxIterations; iteration++ {
		// Orchestrator context is already in the system prompt (SetAgentSystemPrompt at init).
		// Only include user request and previous reports here.
		var contextBuilder strings.Builder
		contextBuilder.WriteString(userMessage)

		if len(reports) > 0 {
			contextBuilder.WriteString("\n\n## Previous Specialist Reports:\n")
			for _, r := range reports {
				contextBuilder.WriteString(fmt.Sprintf("\n### Report from %s (Iteration %d):\n", r.SpecialistName, r.Iteration))
				contextBuilder.WriteString(fmt.Sprintf("- Status: %s\n", r.Status))
				contextBuilder.WriteString(fmt.Sprintf("- Confidence: %.0f%%\n", r.Confidence*100))
				if r.NeedsReview {
					contextBuilder.WriteString("- ⚠️ Requested orchestrator review\n")
				}
				if r.RequestHelp != "" {
					contextBuilder.WriteString(fmt.Sprintf("- 🤝 Requested help from: %s\n", r.RequestHelp))
					contextBuilder.WriteString(fmt.Sprintf("  Context: %s\n", r.HelpContext))
				}
				if len(r.Suggestions) > 0 {
					contextBuilder.WriteString("- Suggestions:\n")
					for _, s := range r.Suggestions {
						contextBuilder.WriteString(fmt.Sprintf("  - %s\n", s))
					}
				}
				contextBuilder.WriteString(fmt.Sprintf("- Result Preview: %s\n", truncate(r.Result, 500)))
			}

			contextBuilder.WriteString("\n## Instructions:\n")
			contextBuilder.WriteString("Based on the previous reports:\n")
			contextBuilder.WriteString("1. If a specialist requested help, delegate to that specialist first\n")
			contextBuilder.WriteString("2. If a specialist needs review, analyze their result and decide next steps\n")
			contextBuilder.WriteString("3. If results are partial or low confidence, consider re-delegating or getting additional input\n")
			contextBuilder.WriteString("4. If all tasks are complete with high confidence, aggregate and return the final result\n")
		}

		// Process through orchestrator
		result, err := oa.ProcessDirect(ctx, contextBuilder.String(), fmt.Sprintf("orchestrator_iteration_%d", iteration))
		if err != nil {
			return "", fmt.Errorf("orchestrator iteration %d failed: %w", iteration, err)
		}

		// Check if we received a structured report (from delegation tool)
		// The last report will be from the most recent delegation
		lastReport := oa.getLastSpecialistReport()
		if lastReport != nil {
			lastReport.Iteration = iteration
			reports = append(reports, *lastReport)
			teamCtx.mu.Lock()
			teamCtx.Progress[lastReport.SpecialistName] = lastReport.Status
			teamCtx.Communications = append(teamCtx.Communications, TeamComm{
				From:      lastReport.SpecialistName,
				To:        "orchestrator",
				Message:   fmt.Sprintf("Completed with status: %s, confidence: %.0f%%", lastReport.Status, lastReport.Confidence*100),
				Type:      "response",
				Timestamp: time.Now(),
			})
			teamCtx.mu.Unlock()

			// Check if we need to handle help request
			if lastReport.RequestHelp != "" && lastReport.Status != "complete" {
				// Emit help request status
				emitAgentStatus(ctx, AgentStatusEvent{
					Agent:          lastReport.SpecialistName,
					Status:         "requesting_help",
					SpecialistName: lastReport.RequestHelp,
					Reason:         lastReport.HelpContext,
					Timestamp:      time.Now(),
				})

				// Record communication
				teamCtx.mu.Lock()
				teamCtx.Communications = append(teamCtx.Communications, TeamComm{
					From:      lastReport.SpecialistName,
					To:        lastReport.RequestHelp,
					Message:   lastReport.HelpContext,
					Type:      "request",
					Timestamp: time.Now(),
				})
				teamCtx.mu.Unlock()
				continue // Next iteration will handle the help request
			}

			// Check if complete
			if lastReport.Status == "complete" && !lastReport.NeedsReview && lastReport.Confidence >= 0.8 {
				logger.InfoCF("agent", "Task completed with high confidence", map[string]interface{}{
					"specialist": lastReport.SpecialistName,
					"confidence": lastReport.Confidence,
					"iterations": iteration + 1,
				})
				break
			}
		} else {
			// No structured report - treat as final result
			return result, nil
		}
	}

	// Aggregate results from all reports
	return oa.aggregateReportsWithStatus(ctx, reports, teamCtx), nil
}

func (oa *OrchestratorAgent) aggregateReportsWithStatus(ctx context.Context, reports []SpecialistReport, teamCtx *TeamContext) string {
	if len(reports) == 0 {
		return oa.aggregateReports(reports, teamCtx)
	}

	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:     "orchestrator",
		Status:    "synthesis_start",
		Timestamp: time.Now(),
	})

	aggregated := oa.aggregateReports(reports, teamCtx)

	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:     "orchestrator",
		Status:    "synthesis_end",
		Timestamp: time.Now(),
	})

	return aggregated
}

// getLastSpecialistReport retrieves the last specialist report if available
// This is set by the delegation tool after a specialist completes
func (oa *OrchestratorAgent) getLastSpecialistReport() *SpecialistReport {
	oa.lastReportMu.RLock()
	defer oa.lastReportMu.RUnlock()
	return oa.lastReport
}

// storeLastReport saves the last specialist report for feedback loop access
func (oa *OrchestratorAgent) storeLastReport(report *SpecialistReport) {
	oa.lastReportMu.Lock()
	defer oa.lastReportMu.Unlock()
	oa.lastReport = report
}

// parseSpecialistReport extracts structured report data from specialist response
func (oa *OrchestratorAgent) parseSpecialistReport(specialistName, result string) *SpecialistReport {
	report := &SpecialistReport{
		SpecialistName: specialistName,
		Status:         "complete",
		Confidence:     0.9, // Default high confidence
		Result:         result,
		Timestamp:      time.Now(),
	}

	// Try to extract JSON from markdown block
	jsonStr := ""
	if start := strings.Index(result, "```json"); start != -1 {
		end := strings.Index(result[start+7:], "```")
		if end != -1 {
			jsonStr = strings.TrimSpace(result[start+7 : start+7+end])
		}
	} else if start := strings.Index(result, "{"); start != -1 {
		// Fallback for raw JSON - find matching closing brace
		depth := 0
		end := -1
		for i := start; i < len(result); i++ {
			switch result[i] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					end = i
					break
				}
			}
			if end != -1 {
				break
			}
		}
		if end != -1 {
			jsonStr = result[start : end+1]
		}
	}

	if jsonStr != "" {
		var parsed struct {
			Status      string  `json:"status"`
			Confidence  float64 `json:"confidence"`
			RequestHelp string  `json:"request_help"`
			Suggestion  string  `json:"suggestion"`
		}

		if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
			if parsed.Status != "" {
				report.Status = parsed.Status
			}
			if parsed.Confidence > 0 {
				report.Confidence = parsed.Confidence
			}
			if parsed.RequestHelp != "" {
				report.RequestHelp = parsed.RequestHelp
				report.NeedsReview = true
			}
			if parsed.Suggestion != "" {
				report.Suggestions = []string{parsed.Suggestion}
			}

			logger.DebugCF("agent", "Parsed specialist report", map[string]interface{}{
				"specialist": specialistName,
				"status":     report.Status,
				"confidence": report.Confidence,
				"help":       report.RequestHelp,
			})
		}
	}

	return report
}

// cleanSpecialistResult removes the JSON header from specialist response for user display
func (oa *OrchestratorAgent) cleanSpecialistResult(result string) string {
	// 1. Try to find and remove markdown json block
	if start := strings.Index(result, "```json"); start != -1 {
		end := strings.Index(result[start+7:], "```")
		if end != -1 {
			// Find the actual end index in the string
			blockEnd := start + 7 + end + 3
			cleaned := result[:start] + result[blockEnd:]
			return strings.TrimSpace(cleaned)
		}
	}

	// 2. Fallback to old behavior for unformatted JSON
	lines := strings.SplitN(result, "\n", 2)
	if len(lines) > 1 {
		firstLine := strings.TrimSpace(lines[0])

		// If first line is a JSON object, return the rest
		if strings.HasPrefix(firstLine, "{") && strings.Contains(firstLine, "}") {
			// Also checking if it parses as the report properly
			var parsed struct {
				Status string `json:"status"`
			}
			jsonEnd := strings.Index(firstLine, "}") + 1
			if err := json.Unmarshal([]byte(firstLine[:jsonEnd]), &parsed); err == nil && parsed.Status != "" {
				return strings.TrimSpace(lines[1])
			}
		}
	}

	return strings.TrimSpace(result)
}

// aggregateReports combines results from multiple specialist reports
func (oa *OrchestratorAgent) aggregateReports(reports []SpecialistReport, teamCtx *TeamContext) string {
	if len(reports) == 0 {
		return "No specialist reports to aggregate."
	}

	var result strings.Builder
	result.WriteString("## Task Completion Summary\n\n")

	// Add team activity summary
	teamCtx.mu.Lock()
	comms := make([]TeamComm, len(teamCtx.Communications))
	copy(comms, teamCtx.Communications)
	teamCtx.mu.Unlock()
	if len(comms) > 0 {
		result.WriteString("### Team Collaboration\n")
		for _, comm := range comms {
			result.WriteString(fmt.Sprintf("- %s → %s: %s\n", comm.From, comm.To, truncate(comm.Message, 100)))
		}
		result.WriteString("\n")
	}

	// Collect results from specialists
	result.WriteString("### Specialist Contributions\n\n")
	for _, r := range reports {
		if r.Status == "complete" || r.Status == "partial" {
			result.WriteString(fmt.Sprintf("**%s** (Confidence: %.0f%%):\n", r.SpecialistName, r.Confidence*100))
			result.WriteString(r.Result)
			result.WriteString("\n\n")
		}
	}

	// Add any artifacts created
	allArtifacts := make([]string, 0)
	for _, r := range reports {
		allArtifacts = append(allArtifacts, r.Artifacts...)
	}
	if len(allArtifacts) > 0 {
		result.WriteString("### Artifacts Created\n")
		for _, a := range allArtifacts {
			result.WriteString(fmt.Sprintf("- %s\n", a))
		}
	}

	return result.String()
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
