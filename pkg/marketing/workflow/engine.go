package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/makoclaw/pkg/logger"
)

// Stage represents a single step in the campaign pipeline.
type Stage string

const (
	StageBrief    Stage = "brief"
	StageStrategy Stage = "strategy"
	StageCopy     Stage = "copy"
	StageVisuals  Stage = "visuals"
	StageSchedule Stage = "schedule"
	StageLaunch   Stage = "launch"
)

// StageState is the status of a single stage within the workflow.
type StageState string

const (
	StatePending          StageState = "pending"
	StateGenerating       StageState = "generating"
	StateAwaitingApproval StageState = "awaiting_approval"
	StateApproved         StageState = "approved"
	StateSkipped          StageState = "skipped"
)

// OrderedStages is the pipeline sequence.
var OrderedStages = []Stage{StageBrief, StageStrategy, StageCopy, StageVisuals, StageSchedule, StageLaunch}

// StageLabels provides human-readable names for each stage.
var StageLabels = map[Stage]string{
	StageBrief:    "Campaign Brief",
	StageStrategy: "Content Strategy",
	StageCopy:     "Marketing Copy",
	StageVisuals:  "Visual Assets",
	StageSchedule: "Posting Schedule",
	StageLaunch:   "Launch Report",
}

// StageIcons provides emoji icons for each stage.
var StageIcons = map[Stage]string{
	StageBrief:    "📋",
	StageStrategy: "🎯",
	StageCopy:     "✍️",
	StageVisuals:  "🎨",
	StageSchedule: "📅",
	StageLaunch:   "🚀",
}

// WorkflowState tracks the entire campaign workflow.
type WorkflowState struct {
	Account      string               `json:"account"`
	Campaign     string               `json:"campaign"`
	Stages       map[Stage]*StageInfo `json:"stages"`
	CurrentStage Stage                `json:"current_stage"`
	StartedAt    *time.Time           `json:"started_at,omitempty"`
	CompletedAt  *time.Time           `json:"completed_at,omitempty"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

// StageInfo tracks the state and output of a single stage.
type StageInfo struct {
	State       StageState `json:"state"`
	Output      string     `json:"output,omitempty"`       // AI raw output or summary
	OutputFiles []string   `json:"output_files,omitempty"` // Files generated
	Feedback    string     `json:"feedback,omitempty"`     // Rejection feedback
	Attempts    int        `json:"attempts"`               // Number of generation attempts
	GeneratedAt *time.Time `json:"generated_at,omitempty"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
}

// GenerateRequest is the payload for generating a stage.
type GenerateRequest struct {
	// Optional overrides per stage
	Objective      string `json:"objective,omitempty"`
	TargetAudience string `json:"target_audience,omitempty"`
	Platforms      string `json:"platforms,omitempty"`
	Tone           string `json:"tone,omitempty"`
	PostCount      int    `json:"post_count,omitempty"`
	ImageStyle     string `json:"image_style,omitempty"`
}

// RejectRequest is the payload for rejecting a stage output.
type RejectRequest struct {
	Feedback string `json:"feedback"`
}

// AIInvoker is the interface for calling the AI agent from Go.
// This decouples the workflow engine from the agent loop implementation.
type AIInvoker interface {
	Invoke(ctx context.Context, prompt, sessionKey string) (string, error)
}

// Engine manages the campaign workflow state machine.
type Engine struct {
	mu sync.Mutex
}

// NewEngine creates a new workflow engine.
func NewEngine() *Engine {
	return &Engine{}
}

// workflowPath returns the path to the workflow state file.
func workflowPath(campaignPath string) string {
	return filepath.Join(campaignPath, "workflow.json")
}

// Load reads the workflow state from disk. Returns nil if no workflow exists.
func (e *Engine) Load(campaignPath string) (*WorkflowState, error) {
	data, err := os.ReadFile(workflowPath(campaignPath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read workflow: %w", err)
	}
	var state WorkflowState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse workflow: %w", err)
	}
	return &state, nil
}

// Save persists the workflow state to disk.
func (e *Engine) Save(campaignPath string, state *WorkflowState) error {
	state.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal workflow: %w", err)
	}
	return os.WriteFile(workflowPath(campaignPath), data, 0644)
}

// Start initializes a new workflow for the campaign.
func (e *Engine) Start(account, campaign, campaignPath string) (*WorkflowState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	existing, err := e.Load(campaignPath)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil // already started
	}

	now := time.Now().UTC()
	stages := make(map[Stage]*StageInfo, len(OrderedStages))
	for _, s := range OrderedStages {
		stages[s] = &StageInfo{State: StatePending}
	}

	state := &WorkflowState{
		Account:      account,
		Campaign:     campaign,
		Stages:       stages,
		CurrentStage: StageBrief,
		StartedAt:    &now,
		UpdatedAt:    now,
	}

	// Auto-detect: if brief.md already exists, mark brief as approved
	briefPath := filepath.Join(campaignPath, "brief.md")
	if data, err := os.ReadFile(briefPath); err == nil && len(data) > 0 {
		stages[StageBrief].State = StateApproved
		stages[StageBrief].Output = "Pre-existing brief.md"
		stages[StageBrief].ApprovedAt = &now
		state.CurrentStage = StageStrategy
	}

	// Auto-detect strategy
	strategyPath := filepath.Join(campaignPath, "strategy.md")
	if data, err := os.ReadFile(strategyPath); err == nil && len(data) > 0 {
		stages[StageStrategy].State = StateApproved
		stages[StageStrategy].Output = "Pre-existing strategy.md"
		stages[StageStrategy].ApprovedAt = &now
		state.CurrentStage = StageCopy
	}

	if err := e.Save(campaignPath, state); err != nil {
		return nil, err
	}
	return state, nil
}

// Generate invokes AI to generate content for the current stage.
func (e *Engine) Generate(ctx context.Context, campaignPath string, invoker AIInvoker, req GenerateRequest) (*WorkflowState, string, error) {
	// Phase 1: Validate and mark as generating (short lock)
	e.mu.Lock()

	state, err := e.Load(campaignPath)
	if err != nil {
		e.mu.Unlock()
		return nil, "", fmt.Errorf("load workflow: %w", err)
	}
	if state == nil {
		e.mu.Unlock()
		return nil, "", fmt.Errorf("workflow not started")
	}

	stage := state.CurrentStage
	info := state.Stages[stage]

	// Allow re-generating if awaiting approval or pending
	if info.State != StatePending && info.State != StateAwaitingApproval {
		e.mu.Unlock()
		return nil, "", fmt.Errorf("stage %s is in state %s, cannot generate", stage, info.State)
	}

	info.State = StateGenerating
	if err := e.Save(campaignPath, state); err != nil {
		e.mu.Unlock()
		return nil, "", err
	}

	// Read existing context to build the prompt
	stageCtx := e.buildContext(campaignPath, state, req)

	// Build the prompt for this stage
	prompt := buildStagePrompt(stage, stageCtx, req)

	// Release lock before AI invocation (can take minutes)
	e.mu.Unlock()

	sessionKey := fmt.Sprintf("marketing_workflow_%s_%s_%s", state.Account, state.Campaign, stage)
	result, err := invoker.Invoke(ctx, prompt, sessionKey)

	// Phase 2: Update state with result (short lock)
	e.mu.Lock()
	defer e.mu.Unlock()

	// Re-load state in case it changed while we were unlocked
	state, loadErr := e.Load(campaignPath)
	if loadErr != nil {
		return nil, "", fmt.Errorf("reload workflow after generation: %w", loadErr)
	}
	info = state.Stages[stage]

	if err != nil {
		info.State = StatePending // rollback
		_ = e.Save(campaignPath, state)
		return nil, "", fmt.Errorf("AI generation failed: %w", err)
	}

	now := time.Now().UTC()
	info.Output = result
	info.State = StateAwaitingApproval
	info.Attempts++
	info.GeneratedAt = &now
	info.Feedback = ""

	// Parse output files from AI response
	info.OutputFiles = extractOutputFiles(stage, campaignPath)

	if err := e.Save(campaignPath, state); err != nil {
		return nil, "", err
	}

	logger.InfoCF("workflow", "stage generated", map[string]interface{}{
		"account":  state.Account,
		"campaign": state.Campaign,
		"stage":    string(stage),
		"attempt":  info.Attempts,
	})

	return state, result, nil
}

// Approve marks the current stage as approved and advances to the next stage.
func (e *Engine) Approve(campaignPath string) (*WorkflowState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	state, err := e.Load(campaignPath)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, fmt.Errorf("workflow not started")
	}

	stage := state.CurrentStage
	info := state.Stages[stage]
	if info.State != StateAwaitingApproval {
		return nil, fmt.Errorf("stage %s is in state %s, cannot approve", stage, info.State)
	}

	now := time.Now().UTC()
	info.State = StateApproved
	info.ApprovedAt = &now

	// Advance to next stage
	nextStage := nextStage(stage)
	if nextStage != "" {
		state.CurrentStage = nextStage
	} else {
		// All stages done!
		state.CompletedAt = &now
	}

	if err := e.Save(campaignPath, state); err != nil {
		return nil, err
	}

	logger.InfoCF("workflow", "stage approved", map[string]interface{}{
		"account":  state.Account,
		"campaign": state.Campaign,
		"stage":    string(stage),
	})

	return state, nil
}

// Reject marks the current stage back to pending with feedback for re-generation.
func (e *Engine) Reject(campaignPath string, feedback string) (*WorkflowState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	state, err := e.Load(campaignPath)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, fmt.Errorf("workflow not started")
	}

	stage := state.CurrentStage
	info := state.Stages[stage]
	if info.State != StateAwaitingApproval {
		return nil, fmt.Errorf("stage %s is in state %s, cannot reject", stage, info.State)
	}

	info.State = StatePending
	info.Feedback = feedback

	if err := e.Save(campaignPath, state); err != nil {
		return nil, err
	}

	logger.InfoCF("workflow", "stage rejected", map[string]interface{}{
		"account":  state.Account,
		"campaign": state.Campaign,
		"stage":    string(stage),
		"feedback": feedback,
	})

	return state, nil
}

// Skip marks the current stage as skipped and advances.
func (e *Engine) Skip(campaignPath string) (*WorkflowState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	state, err := e.Load(campaignPath)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, fmt.Errorf("workflow not started")
	}

	stage := state.CurrentStage
	info := state.Stages[stage]
	info.State = StateSkipped

	now := time.Now().UTC()
	next := nextStage(stage)
	if next != "" {
		state.CurrentStage = next
	} else {
		state.CompletedAt = &now
	}

	if err := e.Save(campaignPath, state); err != nil {
		return nil, err
	}
	return state, nil
}

// ResetStage resets a specific stage back to pending (allows going back).
func (e *Engine) ResetStage(campaignPath string, target Stage) (*WorkflowState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	state, err := e.Load(campaignPath)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, fmt.Errorf("workflow not started")
	}

	// Reset target and all subsequent stages
	found := false
	for _, s := range OrderedStages {
		if s == target {
			found = true
		}
		if found {
			state.Stages[s].State = StatePending
			state.Stages[s].Output = ""
			state.Stages[s].OutputFiles = nil
			state.Stages[s].Feedback = ""
			state.Stages[s].ApprovedAt = nil
			state.Stages[s].GeneratedAt = nil
		}
	}
	state.CurrentStage = target
	state.CompletedAt = nil

	if err := e.Save(campaignPath, state); err != nil {
		return nil, err
	}
	return state, nil
}

// buildContext gathers existing campaign artifacts to feed as context to the AI.
func (e *Engine) buildContext(campaignPath string, state *WorkflowState, req GenerateRequest) *StageContext {
	ctx := &StageContext{
		Account:        state.Account,
		Campaign:       state.Campaign,
		Objective:      req.Objective,
		TargetAudience: req.TargetAudience,
		Platforms:      req.Platforms,
		Tone:           req.Tone,
		PostCount:      req.PostCount,
		ImageStyle:     req.ImageStyle,
	}

	// Read brief if available
	if data, err := os.ReadFile(filepath.Join(campaignPath, "brief.md")); err == nil {
		ctx.Brief = string(data)
	}

	// Read strategy if available
	if data, err := os.ReadFile(filepath.Join(campaignPath, "strategy.md")); err == nil {
		ctx.Strategy = string(data)
	}

	// Read existing copy files
	copyDir := filepath.Join(campaignPath, "copy")
	if entries, err := os.ReadDir(copyDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				if data, err := os.ReadFile(filepath.Join(copyDir, entry.Name())); err == nil {
					ctx.CopyFiles = append(ctx.CopyFiles, CopyFile{
						Name:     entry.Name(),
						Contents: string(data),
					})
				}
			}
		}
	}

	// Include previous feedback if regenerating
	if info, ok := state.Stages[state.CurrentStage]; ok && info.Feedback != "" {
		ctx.PreviousFeedback = info.Feedback
	}

	return ctx
}

// extractOutputFiles finds files that were generated by the AI for a given stage.
func extractOutputFiles(stage Stage, campaignPath string) []string {
	var files []string
	var dir string

	switch stage {
	case StageBrief:
		if _, err := os.Stat(filepath.Join(campaignPath, "brief.md")); err == nil {
			return []string{"brief.md"}
		}
		return files
	case StageStrategy:
		if _, err := os.Stat(filepath.Join(campaignPath, "strategy.md")); err == nil {
			return []string{"strategy.md"}
		}
		return files
	case StageCopy:
		dir = filepath.Join(campaignPath, "copy")
	case StageVisuals:
		dir = filepath.Join(campaignPath, "assets", "images")
	case StageSchedule:
		dir = filepath.Join(campaignPath, "schedules")
	case StageLaunch:
		dir = filepath.Join(campaignPath, "analytics")
	}

	if dir == "" {
		return files
	}

	if entries, err := os.ReadDir(dir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, entry.Name())
			}
		}
	}
	return files
}

// nextStage returns the next stage in the pipeline, or empty string if done.
func nextStage(current Stage) Stage {
	for i, s := range OrderedStages {
		if s == current && i+1 < len(OrderedStages) {
			return OrderedStages[i+1]
		}
	}
	return ""
}

// Progress returns completion percentage and stage counts.
func (state *WorkflowState) Progress() (completed, total int, pct float64) {
	total = len(OrderedStages)
	for _, s := range OrderedStages {
		if info, ok := state.Stages[s]; ok {
			if info.State == StateApproved || info.State == StateSkipped {
				completed++
			}
		}
	}
	if total > 0 {
		pct = float64(completed) / float64(total) * 100
	}
	return
}
