package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/agent"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
	"github.com/sipeed/makoclaw/pkg/tools"
)

// StepType identifies what kind of step to execute.
type StepType string

const (
	StepPrompt    StepType = "prompt"
	StepTool      StepType = "tool"
	StepCondition StepType = "condition"
	StepLoop      StepType = "loop"
	StepParallel  StepType = "parallel"
	StepApproval  StepType = "approval"
)

// Step defines a single step in a workflow pipeline.
type Step struct {
	ID      string          `json:"id"`
	Type    StepType        `json:"type"`
	Label   string          `json:"label"`
	Config  json.RawMessage `json:"config"`
	OnError string          `json:"on_error"` // "stop" or "continue"
}

// PromptConfig holds settings for a prompt step.
type PromptConfig struct {
	Message string `json:"message"`
	Model   string `json:"model,omitempty"`
}

// ToolConfig holds settings for a tool step.
type ToolConfig struct {
	ToolName string                 `json:"tool_name"`
	Args     map[string]interface{} `json:"args"`
}

// ConditionConfig holds settings for a condition step.
type ConditionConfig struct {
	Operator  string `json:"operator"`  // "contains", "equals", "regex", "not_empty"
	Value     string `json:"value"`     // compare value
	Reference string `json:"reference"` // e.g. "{{step.1.output}}"
}

// LoopConfig holds settings for a loop step that iterates over items.
type LoopConfig struct {
	Items    string `json:"items"`    // comma-separated list or template ref
	Variable string `json:"variable"` // variable name for current item
	Steps    []Step `json:"steps"`    // nested steps to execute per item
}

// ParallelConfig holds settings for parallel step execution.
type ParallelConfig struct {
	Steps          []Step `json:"steps"`           // nested steps to run concurrently
	MaxConcurrency int    `json:"max_concurrency"` // 0 = unlimited
}

// ApprovalConfig holds settings for an approval gate step.
type ApprovalConfig struct {
	Message   string   `json:"message"`
	Approvers []string `json:"approvers,omitempty"`
	Timeout   string   `json:"timeout,omitempty"` // e.g. "24h"
}

// StepResult captures the outcome of a single step execution.
type StepResult struct {
	StepID   string `json:"step_id"`
	StepType string `json:"step_type"`
	Label    string `json:"label"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	Duration int64  `json:"duration_ms"`
	Skipped  bool   `json:"skipped,omitempty"`
}

// Engine executes workflows by running each step in sequence.
type Engine struct {
	agentLoop *agent.AgentLoop
	tools     *tools.ToolRegistry
	store     *storage.Storage
}

// NewEngine creates a workflow engine with access to the agent loop and tools.
func NewEngine(agentLoop *agent.AgentLoop, toolRegistry *tools.ToolRegistry, store *storage.Storage) *Engine {
	return &Engine{
		agentLoop: agentLoop,
		tools:     toolRegistry,
		store:     store,
	}
}

// Run executes a workflow and records the run in storage.
func (e *Engine) Run(ctx context.Context, wf *storage.Workflow) ([]StepResult, error) {
	return e.RunWithParams(ctx, wf, nil)
}

// RunWithParams executes a workflow with runtime parameters and records the run in storage.
func (e *Engine) RunWithParams(ctx context.Context, wf *storage.Workflow, params map[string]string) ([]StepResult, error) {
	var steps []Step
	if err := json.Unmarshal(wf.Steps, &steps); err != nil {
		return nil, fmt.Errorf("parsing workflow steps: %w", err)
	}

	// Create a run record
	runID, err := e.store.CreateWorkflowRun(wf.ID)
	if err != nil {
		return nil, fmt.Errorf("creating run record: %w", err)
	}

	logger.InfoCF("workflow", "Starting workflow",
		map[string]interface{}{"workflow_id": wf.ID, "name": wf.Name, "steps": len(steps), "run_id": runID})

	results := make([]StepResult, 0, len(steps))
	sessionKey := fmt.Sprintf("workflow:%d:run:%d", wf.ID, runID)
	conditionSkip := false
	stoppedOnError := false

	for i, step := range steps {
		if ctx.Err() != nil {
			break
		}

		// If previous condition step evaluated to false, skip until next non-condition step
		if conditionSkip && step.Type != StepCondition {
			results = append(results, StepResult{
				StepID:   step.ID,
				StepType: string(step.Type),
				Label:    step.Label,
				Skipped:  true,
			})
			continue
		}
		conditionSkip = false

		start := time.Now()
		var output string
		var stepErr error

		switch step.Type {
		case StepPrompt:
			output, stepErr = e.executePrompt(ctx, step, sessionKey, results, params)
		case StepTool:
			output, stepErr = e.executeTool(ctx, step, results, params)
		case StepCondition:
			matched, evalErr := e.evaluateCondition(step, results, params)
			if evalErr != nil {
				stepErr = evalErr
			} else if !matched {
				conditionSkip = true
				output = "condition: false — skipping next steps"
			} else {
				output = "condition: true — continuing"
			}
		case StepLoop:
			output, stepErr = e.executeLoop(ctx, step, sessionKey, results, params)
		case StepParallel:
			output, stepErr = e.executeParallel(ctx, step, sessionKey, results, params)
		case StepApproval:
			output, stepErr = e.executeApproval(ctx, step, wf.ID, runID)
		default:
			stepErr = fmt.Errorf("unknown step type: %s", step.Type)
		}

		duration := time.Since(start).Milliseconds()
		sr := StepResult{
			StepID:   step.ID,
			StepType: string(step.Type),
			Label:    step.Label,
			Output:   output,
			Duration: duration,
		}
		if stepErr != nil {
			sr.Error = stepErr.Error()
		}
		results = append(results, sr)

		logger.InfoCF("workflow", fmt.Sprintf("Step %d/%d completed", i+1, len(steps)),
			map[string]interface{}{
				"step_id":     step.ID,
				"type":        step.Type,
				"duration_ms": duration,
				"error":       sr.Error,
			})

		// Check on_error policy
		if stepErr != nil && step.OnError != "continue" {
			stoppedOnError = true
			break
		}
	}

	// Save results
	resultsJSON, marshalErr := json.Marshal(results)
	if marshalErr != nil {
		logger.ErrorCF("workflow", "Failed to marshal results", map[string]interface{}{"error": marshalErr.Error()})
		resultsJSON = []byte("[]") // fallback so UpdateWorkflowRun gets valid JSON
	}
	status := "completed"
	if ctx.Err() != nil {
		status = "failed"
	} else if stoppedOnError {
		status = "failed"
	} else {
		// Check if any steps had errors but continued
		for _, r := range results {
			if r.Error != "" {
				status = "completed_with_errors"
				break
			}
		}
	}
	if updateErr := e.store.UpdateWorkflowRun(runID, status, resultsJSON); updateErr != nil {
		logger.ErrorCF("workflow", "Failed to update workflow run", map[string]interface{}{"run_id": runID, "error": updateErr.Error()})
	}

	if ctx.Err() != nil {
		return results, ctx.Err()
	}
	return results, nil
}

// executePrompt sends a message to the AI agent and returns the response.
func (e *Engine) executePrompt(ctx context.Context, step Step, sessionKey string, prevResults []StepResult, params map[string]string) (string, error) {
	if e.agentLoop == nil {
		return "", fmt.Errorf("agent loop not available")
	}

	var cfg PromptConfig
	if err := json.Unmarshal(step.Config, &cfg); err != nil {
		return "", fmt.Errorf("parsing prompt config: %w", err)
	}

	message := interpolate(cfg.Message, prevResults, params)

	if cfg.Model != "" {
		return e.agentLoop.ProcessDirectWithModel(ctx, message, sessionKey, cfg.Model)
	}
	return e.agentLoop.ProcessDirect(ctx, message, sessionKey)
}

// executeTool calls a registered tool directly.
func (e *Engine) executeTool(ctx context.Context, step Step, prevResults []StepResult, params map[string]string) (string, error) {
	if e.tools == nil {
		return "", fmt.Errorf("tool registry not available")
	}

	var cfg ToolConfig
	if err := json.Unmarshal(step.Config, &cfg); err != nil {
		return "", fmt.Errorf("parsing tool config: %w", err)
	}

	// Interpolate template variables in string args
	args := make(map[string]interface{})
	for k, v := range cfg.Args {
		if s, ok := v.(string); ok {
			args[k] = interpolate(s, prevResults, params)
		} else {
			args[k] = v
		}
	}

	return e.tools.Execute(ctx, cfg.ToolName, args)
}

// evaluateCondition checks a condition against previous step outputs.
func (e *Engine) evaluateCondition(step Step, prevResults []StepResult, params map[string]string) (bool, error) {
	var cfg ConditionConfig
	if err := json.Unmarshal(step.Config, &cfg); err != nil {
		return false, fmt.Errorf("parsing condition config: %w", err)
	}

	subject := interpolate(cfg.Reference, prevResults, params)
	value := interpolate(cfg.Value, prevResults, params)

	switch cfg.Operator {
	case "contains":
		return strings.Contains(strings.ToLower(subject), strings.ToLower(value)), nil
	case "equals":
		return strings.TrimSpace(subject) == strings.TrimSpace(value), nil
	case "not_empty":
		return strings.TrimSpace(subject) != "", nil
	case "regex":
		re, err := regexp.Compile(value)
		if err != nil {
			return false, fmt.Errorf("invalid regex %q: %w", value, err)
		}
		return re.MatchString(subject), nil
	default:
		return false, fmt.Errorf("unknown operator: %s", cfg.Operator)
	}
}

// executeLoop runs nested steps for each item in the list.
func (e *Engine) executeLoop(ctx context.Context, step Step, sessionKey string, prevResults []StepResult, params map[string]string) (string, error) {
	var cfg LoopConfig
	if err := json.Unmarshal(step.Config, &cfg); err != nil {
		return "", fmt.Errorf("parsing loop config: %w", err)
	}

	itemsStr := interpolate(cfg.Items, prevResults, params)
	items := strings.Split(itemsStr, ",")
	variable := cfg.Variable
	if variable == "" {
		variable = "item"
	}

	var allOutputs []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		// Create params copy with loop variable
		loopParams := make(map[string]string)
		for k, v := range params {
			loopParams[k] = v
		}
		loopParams[variable] = item

		for _, nested := range cfg.Steps {
			if ctx.Err() != nil {
				return strings.Join(allOutputs, "\n"), ctx.Err()
			}

			var output string
			var stepErr error
			switch nested.Type {
			case StepPrompt:
				output, stepErr = e.executePrompt(ctx, nested, sessionKey, prevResults, loopParams)
			case StepTool:
				output, stepErr = e.executeTool(ctx, nested, prevResults, loopParams)
			default:
				stepErr = fmt.Errorf("unsupported nested step type in loop: %s", nested.Type)
			}

			if stepErr != nil {
				allOutputs = append(allOutputs, fmt.Sprintf("[%s] error: %v", item, stepErr))
				if nested.OnError != "continue" {
					return strings.Join(allOutputs, "\n"), stepErr
				}
			} else {
				allOutputs = append(allOutputs, fmt.Sprintf("[%s] %s", item, output))
			}
		}
	}

	return strings.Join(allOutputs, "\n"), nil
}

// executeParallel runs nested steps concurrently.
func (e *Engine) executeParallel(ctx context.Context, step Step, sessionKey string, prevResults []StepResult, params map[string]string) (string, error) {
	var cfg ParallelConfig
	if err := json.Unmarshal(step.Config, &cfg); err != nil {
		return "", fmt.Errorf("parsing parallel config: %w", err)
	}

	if len(cfg.Steps) == 0 {
		return "no steps to execute", nil
	}

	maxConc := cfg.MaxConcurrency
	if maxConc <= 0 {
		maxConc = len(cfg.Steps)
	}

	type result struct {
		index  int
		output string
		err    error
	}

	results := make(chan result, len(cfg.Steps))
	sem := make(chan struct{}, maxConc)

	for i, nested := range cfg.Steps {
		sem <- struct{}{} // acquire semaphore
		go func(idx int, s Step) {
			defer func() { <-sem }() // release semaphore
			var output string
			var stepErr error
			switch s.Type {
			case StepPrompt:
				output, stepErr = e.executePrompt(ctx, s, sessionKey, prevResults, params)
			case StepTool:
				output, stepErr = e.executeTool(ctx, s, prevResults, params)
			default:
				stepErr = fmt.Errorf("unsupported nested step type in parallel: %s", s.Type)
			}
			results <- result{index: idx, output: output, err: stepErr}
		}(i, nested)
	}

	outputs := make([]string, len(cfg.Steps))
	var firstErr error
	for range cfg.Steps {
		r := <-results
		if r.err != nil {
			outputs[r.index] = fmt.Sprintf("[step %d] error: %v", r.index+1, r.err)
			if firstErr == nil {
				firstErr = r.err
			}
		} else {
			outputs[r.index] = fmt.Sprintf("[step %d] %s", r.index+1, r.output)
		}
	}

	return strings.Join(outputs, "\n"), firstErr
}

// executeApproval creates an approval record and returns pending status.
func (e *Engine) executeApproval(ctx context.Context, step Step, workflowID, runID int64) (string, error) {
	var cfg ApprovalConfig
	if err := json.Unmarshal(step.Config, &cfg); err != nil {
		return "", fmt.Errorf("parsing approval config: %w", err)
	}

	if e.store == nil {
		return "", fmt.Errorf("storage not available for approval tracking")
	}

	message := cfg.Message
	if message == "" {
		message = "Approval required to continue workflow"
	}

	approvalID, err := e.store.CreateWorkflowApproval(workflowID, runID, step.ID, message)
	if err != nil {
		return "", fmt.Errorf("creating approval record: %w", err)
	}

	logger.InfoCF("workflow", "Approval gate created", map[string]interface{}{
		"approval_id": approvalID,
		"step_id":     step.ID,
		"message":     message,
	})

	return fmt.Sprintf("approval pending (id: %d) — %s", approvalID, message), nil
}

// templatePattern matches {{step.N.output}} or {{step.N.error}}
var templatePattern = regexp.MustCompile(`\{\{step\.(\d+)\.(output|error)\}\}`)

// paramPattern matches {{param.name}}
var paramPattern = regexp.MustCompile(`\{\{param\.(\w+)\}\}`)

// interpolate replaces {{step.N.output}}, {{step.N.error}}, and {{param.name}} placeholders
// with values from previous step results and runtime parameters. Step indices are 1-based.
func interpolate(text string, results []StepResult, params map[string]string) string {
	// Replace step references
	text = templatePattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := templatePattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		idx := 0
		fmt.Sscanf(parts[1], "%d", &idx)
		idx-- // convert 1-based to 0-based
		if idx < 0 || idx >= len(results) {
			return match
		}
		if parts[2] == "error" {
			return results[idx].Error
		}
		return results[idx].Output
	})

	// Replace parameter references
	if params != nil {
		text = paramPattern.ReplaceAllStringFunc(text, func(match string) string {
			parts := paramPattern.FindStringSubmatch(match)
			if len(parts) != 2 {
				return match
			}
			paramName := parts[1]
			if val, ok := params[paramName]; ok {
				return val
			}
			return match
		})
	}

	return text
}
