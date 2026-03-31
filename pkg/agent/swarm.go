package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
	"golang.org/x/sync/errgroup"
)

// SwarmRunner coordinates swarm executions using the specialist registry
type SwarmRunner struct {
	registry    *SpecialistRegistry
	costTracker *AgentCostTracker
	storage     *storage.Storage // optional, for persisting team contexts
	mu          sync.RWMutex
}

// SwarmExecution represents a single swarm run
type SwarmExecution struct {
	ID          string                        `json:"id"`
	SwarmName   string                        `json:"swarm_name"`
	Config      config.SwarmConfig            `json:"config"`
	Task        string                        `json:"task"`
	Status      string                        `json:"status"` // "running", "completed", "failed", "cancelled"
	TeamContext *TeamContext                  `json:"team_context"`
	Results     map[string]*SwarmMemberResult `json:"results"`
	StartTime   time.Time                     `json:"start_time"`
	EndTime     time.Time                     `json:"end_time,omitempty"`
	TotalCost   float64                       `json:"total_cost"`
	Error       string                        `json:"error,omitempty"`
	mu          sync.Mutex
}

// SwarmMemberResult holds the output of one member's execution
type SwarmMemberResult struct {
	Member    string    `json:"member"`
	Status    string    `json:"status"` // "pending", "running", "completed", "failed"
	Result    string    `json:"result"`
	Error     string    `json:"error,omitempty"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time,omitempty"`
	CostUSD   float64   `json:"cost_usd"`
}

// SwarmResult is the final output of a swarm execution
type SwarmResult struct {
	ExecutionID    string                        `json:"execution_id"`
	SwarmName      string                        `json:"swarm_name"`
	Status         string                        `json:"status"`
	AggregatedText string                        `json:"aggregated_text"`
	MemberResults  map[string]*SwarmMemberResult `json:"member_results"`
	TotalCost      float64                       `json:"total_cost"`
	Duration       time.Duration                 `json:"duration"`
	TeamContext    *TeamContext                  `json:"team_context"`
}

// NewSwarmRunner creates a new swarm runner
func NewSwarmRunner(registry *SpecialistRegistry, costTracker *AgentCostTracker) *SwarmRunner {
	return &SwarmRunner{
		registry:    registry,
		costTracker: costTracker,
	}
}

// SetStorage sets the storage backend for persisting team contexts.
func (sr *SwarmRunner) SetStorage(s *storage.Storage) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.storage = s
}

// RunSwarm executes a swarm with the given config and task
func (sr *SwarmRunner) RunSwarm(ctx context.Context, swarmCfg config.SwarmConfig, task string) (*SwarmResult, error) {
	if len(swarmCfg.Members) == 0 {
		return nil, fmt.Errorf("swarm '%s' has no members", swarmCfg.Name)
	}

	// Validate all members exist
	for _, member := range swarmCfg.Members {
		if _, err := sr.registry.GetSpecialist(member); err != nil {
			return nil, fmt.Errorf("swarm member '%s' not found in specialist registry", member)
		}
	}

	// Apply timeout
	timeout := time.Duration(swarmCfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 300 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	exec := &SwarmExecution{
		ID:        fmt.Sprintf("swarm-%s-%d", swarmCfg.Name, time.Now().UnixNano()),
		SwarmName: swarmCfg.Name,
		Config:    swarmCfg,
		Task:      task,
		Status:    "running",
		Results:   make(map[string]*SwarmMemberResult),
		StartTime: time.Now(),
	}

	parentTeamCtx := TeamContextFromCtx(ctx)

	// Initialize TeamContext for shared memory
	exec.TeamContext = &TeamContext{
		TaskID:         exec.ID,
		OriginalTask:   task,
		Progress:       make(map[string]string),
		SharedNotes:    []TeamNote{},
		Artifacts:      []string{},
		Decisions:      []TeamDecision{},
		Communications: []TeamComm{},
	}
	if parentTeamCtx != nil {
		exec.TeamContext.UserUUID = parentTeamCtx.UserUUID
		exec.TeamContext.UserID = parentTeamCtx.UserID
	}

	// Embed TeamContext into the context if shared memory is enabled (default true)
	if swarmCfg.SharedMemory {
		ctx = ContextWithTeamContext(ctx, exec.TeamContext)
	}

	// Initialize member results
	for _, member := range swarmCfg.Members {
		exec.Results[member] = &SwarmMemberResult{
			Member: member,
			Status: "pending",
		}
	}

	logger.InfoCF("agent", "Starting swarm execution", map[string]interface{}{
		"swarm":   swarmCfg.Name,
		"mode":    swarmCfg.Mode,
		"members": swarmCfg.Members,
		"task":    truncate(task, 100),
	})

	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:  "swarm:" + swarmCfg.Name,
		Status: "starting",
		Reason: fmt.Sprintf("Starting swarm '%s' with %d members in %s mode", swarmCfg.Name, len(swarmCfg.Members), swarmCfg.Mode),
	})

	var err error
	switch swarmCfg.Mode {
	case "parallel":
		err = sr.runParallel(ctx, exec)
	case "consensus":
		err = sr.runConsensus(ctx, exec)
	default: // "sequential" or empty
		err = sr.runSequential(ctx, exec)
	}

	exec.EndTime = time.Now()
	if err != nil {
		exec.Status = "failed"
		exec.Error = err.Error()
	} else {
		exec.Status = "completed"
	}

	// Calculate total cost from member results
	exec.mu.Lock()
	for _, r := range exec.Results {
		exec.TotalCost += r.CostUSD
	}
	exec.mu.Unlock()

	// Build aggregated text
	aggregated := sr.aggregateResults(exec)

	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:  "swarm:" + swarmCfg.Name,
		Status: exec.Status,
		Reason: fmt.Sprintf("Swarm '%s' %s (%.4f USD)", swarmCfg.Name, exec.Status, exec.TotalCost),
	})

	// Persist TeamContext for analysis and follow-up (best-effort)
	sr.persistTeamContext(exec)

	return &SwarmResult{
		ExecutionID:    exec.ID,
		SwarmName:      exec.SwarmName,
		Status:         exec.Status,
		AggregatedText: aggregated,
		MemberResults:  exec.Results,
		TotalCost:      exec.TotalCost,
		Duration:       exec.EndTime.Sub(exec.StartTime),
		TeamContext:    exec.TeamContext,
	}, err
}

// runSequential executes members one after another, passing shared notes between them
func (sr *SwarmRunner) runSequential(ctx context.Context, exec *SwarmExecution) error {
	for _, memberName := range exec.Config.Members {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err := sr.checkBudget(exec); err != nil {
			return err
		}

		taskWithContext := sr.buildTaskWithSharedNotes(exec, memberName)

		if err := sr.executeMember(ctx, exec, memberName, taskWithContext); err != nil {
			logger.WarnCF("agent", "Swarm member failed", map[string]interface{}{
				"swarm":  exec.SwarmName,
				"member": memberName,
				"error":  err.Error(),
			})
			// Continue with next member on failure
		}
	}
	return nil
}

// runParallel executes all members concurrently with shared TeamContext
func (sr *SwarmRunner) runParallel(ctx context.Context, exec *SwarmExecution) error {
	g, gCtx := errgroup.WithContext(ctx)

	for _, memberName := range exec.Config.Members {
		name := memberName // capture for goroutine
		g.Go(func() error {
			if err := sr.checkBudget(exec); err != nil {
				return err
			}
			return sr.executeMember(gCtx, exec, name, exec.Task)
		})
	}

	err := g.Wait()
	// errgroup returns first error; we still have results from successful members
	return err
}

// runConsensus runs all members independently on the same task, then synthesizes
func (sr *SwarmRunner) runConsensus(ctx context.Context, exec *SwarmExecution) error {
	// Run all members in parallel WITHOUT shared context
	g, gCtx := errgroup.WithContext(ctx)

	for _, memberName := range exec.Config.Members {
		name := memberName
		g.Go(func() error {
			if err := sr.checkBudget(exec); err != nil {
				return err
			}
			return sr.executeMember(gCtx, exec, name, exec.Task)
		})
	}

	_ = g.Wait() // Collect all results even if some fail

	// Synthesize results via LLM if we have multiple completed results
	return sr.synthesizeConsensus(ctx, exec)
}

// synthesizeConsensus uses an LLM call to synthesize results from multiple specialists
func (sr *SwarmRunner) synthesizeConsensus(ctx context.Context, exec *SwarmExecution) error {
	exec.mu.Lock()
	var completedResults []string
	for member, result := range exec.Results {
		if result.Status == "completed" && result.Result != "" {
			completedResults = append(completedResults, fmt.Sprintf("**%s**:\n%s", member, result.Result))
		}
	}
	exec.mu.Unlock()

	if len(completedResults) < 2 {
		return nil // Not enough results to synthesize
	}

	// Find the synthesizer agent
	synthesizerName := exec.Config.SynthesizerAgent
	if synthesizerName == "" && len(exec.Config.Members) > 0 {
		synthesizerName = exec.Config.Members[0] // Default to first member
	}

	synthesizer, err := sr.registry.GetSpecialist(synthesizerName)
	if err != nil {
		logger.WarnCF("agent", "Consensus synthesizer not found, skipping synthesis", map[string]interface{}{
			"synthesizer": synthesizerName,
			"error":       err.Error(),
		})
		return nil
	}

	synthesisPrompt := fmt.Sprintf(`You are synthesizing consensus from %d independent analyses of the same task.

Original task: %s

Here are the independent analyses:

%s

---

Synthesize a unified consensus answer:
1. Identify points of agreement across all analyses
2. Note any significant disagreements or different perspectives
3. Provide a final consolidated recommendation based on the majority view
4. Be concise but thorough`, len(completedResults), exec.Task, strings.Join(completedResults, "\n\n---\n\n"))

	result, synthErr := synthesizer.ProcessDirect(ctx, synthesisPrompt, "")
	if synthErr != nil {
		logger.WarnCF("agent", "Consensus synthesis failed", map[string]interface{}{
			"error": synthErr.Error(),
		})
		return nil // Non-fatal — individual results still available
	}

	// Store synthesis as a special result
	exec.mu.Lock()
	exec.Results["_consensus"] = &SwarmMemberResult{
		Member:    "_consensus",
		Status:    "completed",
		Result:    result,
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}
	exec.mu.Unlock()

	return nil
}

// executeMember runs a single specialist member
func (sr *SwarmRunner) executeMember(ctx context.Context, exec *SwarmExecution, memberName, task string) error {
	specialist, err := sr.registry.GetSpecialist(memberName)
	if err != nil {
		exec.mu.Lock()
		exec.Results[memberName].Status = "failed"
		exec.Results[memberName].Error = err.Error()
		exec.mu.Unlock()
		return err
	}
	if exec.TeamContext != nil {
		specialist.SetUserForAgent(exec.TeamContext.UserUUID, exec.TeamContext.UserID)
	}

	exec.mu.Lock()
	exec.Results[memberName].Status = "running"
	exec.Results[memberName].StartTime = time.Now()
	exec.TeamContext.Progress[memberName] = "in_progress"
	exec.mu.Unlock()

	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:  memberName,
		Status: "working",
		Reason: fmt.Sprintf("Specialist '%s' working on swarm task", memberName),
	})

	// Get cost before execution
	var costBefore float64
	if sr.costTracker != nil {
		if m := sr.costTracker.GetMetrics(memberName); m != nil {
			costBefore = m.EstimatedCostUSD
		}
	}

	// CRITICAL FIX for ISSUE-9: Pass user context directly to avoid reset in applyMessageUserContext
	// We can't use ProcessWithSpeciality/ProcessWithSpecialityForUser because they call ProcessDirect
	// which loses the UserID. Instead, we call ProcessDirectWithUser which
	// preserves user context through the entire execution path.
	specialist.SetUserForAgent(exec.TeamContext.UserUUID, exec.TeamContext.UserID)
	result, err := specialist.ProcessDirectWithUser(ctx, exec.TeamContext.UserID, task, "")

	// Get cost after execution
	var memberCost float64
	if sr.costTracker != nil {
		if m := sr.costTracker.GetMetrics(memberName); m != nil {
			memberCost = m.EstimatedCostUSD - costBefore
		}
	}

	exec.mu.Lock()
	defer exec.mu.Unlock()

	if err != nil {
		exec.Results[memberName].Status = "failed"
		exec.Results[memberName].Error = err.Error()
		exec.Results[memberName].EndTime = time.Now()
		exec.Results[memberName].CostUSD = memberCost
		exec.TeamContext.Progress[memberName] = "failed"

		// Add failure note to shared context
		exec.TeamContext.SharedNotes = append(exec.TeamContext.SharedNotes, TeamNote{
			Author:    memberName,
			Content:   fmt.Sprintf("Failed: %s", err.Error()),
			Timestamp: time.Now(),
		})
		return err
	}

	exec.Results[memberName].Status = "completed"
	exec.Results[memberName].Result = result
	exec.Results[memberName].EndTime = time.Now()
	exec.Results[memberName].CostUSD = memberCost
	exec.TeamContext.Progress[memberName] = "complete"

	// Add result summary to shared notes for subsequent members
	summary := result
	if len(summary) > 2000 {
		summary = summary[:2000] + "..."
	}
	exec.TeamContext.SharedNotes = append(exec.TeamContext.SharedNotes, TeamNote{
		Author:    memberName,
		Content:   summary,
		Timestamp: time.Now(),
	})

	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:  memberName,
		Status: "completed",
		Reason: fmt.Sprintf("Specialist '%s' completed swarm task", memberName),
	})

	return nil
}

// buildTaskWithSharedNotes prepends shared notes from previous members to the task
func (sr *SwarmRunner) buildTaskWithSharedNotes(exec *SwarmExecution, currentMember string) string {
	exec.mu.Lock()
	defer exec.mu.Unlock()

	if len(exec.TeamContext.SharedNotes) == 0 {
		return exec.Task
	}

	var sb strings.Builder
	sb.WriteString("## Context from team members\n\n")

	totalLen := 0
	for _, note := range exec.TeamContext.SharedNotes {
		if note.Author == currentMember {
			continue
		}
		entry := fmt.Sprintf("**%s**: %s\n\n", note.Author, note.Content)
		maxChars := exec.Config.SharedNotesMaxChars
		if maxChars <= 0 {
			maxChars = 4000
		}
		if totalLen+len(entry) > maxChars { // Cap shared context
			sb.WriteString("...(additional context truncated)\n\n")
			break
		}
		sb.WriteString(entry)
		totalLen += len(entry)
	}

	sb.WriteString("## Your task\n\n")
	sb.WriteString(exec.Task)

	return sb.String()
}

// checkBudget returns an error if the swarm has exceeded its cost budget
func (sr *SwarmRunner) checkBudget(exec *SwarmExecution) error {
	if exec.Config.MaxBudget <= 0 {
		return nil
	}

	exec.mu.Lock()
	defer exec.mu.Unlock()

	var totalCost float64
	for _, r := range exec.Results {
		totalCost += r.CostUSD
	}

	if totalCost >= exec.Config.MaxBudget {
		return fmt.Errorf("swarm '%s' exceeded budget limit (%.4f USD >= %.4f USD)", exec.SwarmName, totalCost, exec.Config.MaxBudget)
	}
	return nil
}

// aggregateResults builds a summary text from all member results
func (sr *SwarmRunner) aggregateResults(exec *SwarmExecution) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Swarm '%s' Results\n\n", exec.SwarmName))
	sb.WriteString(fmt.Sprintf("**Mode**: %s | **Duration**: %s | **Cost**: $%.4f\n\n",
		exec.Config.Mode, exec.EndTime.Sub(exec.StartTime).Round(time.Millisecond), exec.TotalCost))

	for _, memberName := range exec.Config.Members {
		r, ok := exec.Results[memberName]
		if !ok {
			continue
		}
		sb.WriteString(fmt.Sprintf("### %s (%s)\n", memberName, r.Status))
		if r.Error != "" {
			sb.WriteString(fmt.Sprintf("Error: %s\n\n", r.Error))
		} else if r.Result != "" {
			sb.WriteString(r.Result)
			sb.WriteString("\n\n")
		}
	}

	// Include team communications if any
	if exec.TeamContext != nil && len(exec.TeamContext.Communications) > 0 {
		sb.WriteString("### Team Communications\n")
		for _, comm := range exec.TeamContext.Communications {
			sb.WriteString(fmt.Sprintf("- %s → %s: %s\n", comm.From, comm.To, truncate(comm.Message, 100)))
		}
	}

	return sb.String()
}

// BuiltInSwarmTemplates provides preset swarm configurations
var BuiltInSwarmTemplates = map[string]config.SwarmConfig{
	"code-review-team": {
		Name:         "Code Review Team",
		Description:  "Sequential code review: developer writes, security analyst checks, reviewer finalizes",
		Members:      []string{"developer", "security_analyst", "code_reviewer"},
		Mode:         "sequential",
		SharedMemory: true,
		Timeout:      600,
	},
	"research-team": {
		Name:         "Research Team",
		Description:  "Independent research with consensus synthesis",
		Members:      []string{"researcher", "analyst"},
		Mode:         "consensus",
		SharedMemory: false,
		Timeout:      300,
	},
	"full-stack-team": {
		Name:         "Full Stack Team",
		Description:  "Parallel frontend and backend development",
		Members:      []string{"frontend_dev", "backend_dev", "designer"},
		Mode:         "parallel",
		SharedMemory: true,
		Timeout:      600,
	},
}

// persistTeamContext saves the team context to storage for analysis and follow-up.
func (sr *SwarmRunner) persistTeamContext(exec *SwarmExecution) {
	sr.mu.RLock()
	s := sr.storage
	sr.mu.RUnlock()

	if s == nil || exec.TeamContext == nil {
		return
	}

	var notes, decisions string
	if len(exec.TeamContext.SharedNotes) > 0 {
		var parts []string
		for _, n := range exec.TeamContext.SharedNotes {
			parts = append(parts, fmt.Sprintf("%s: %s", n.Author, n.Content))
		}
		notes = strings.Join(parts, "\n---\n")
	}
	if len(exec.TeamContext.Decisions) > 0 {
		var parts []string
		for _, d := range exec.TeamContext.Decisions {
			parts = append(parts, fmt.Sprintf("%s: %s", d.Author, d.Decision))
		}
		decisions = strings.Join(parts, "\n---\n")
	}

	if err := s.SaveTeamContext(exec.ID, exec.SwarmName, exec.Task, exec.Status, notes, decisions, exec.TotalCost); err != nil {
		logger.WarnCF("agent", "Failed to persist team context", map[string]interface{}{
			"swarm": exec.SwarmName,
			"error": err.Error(),
		})
	}
}
