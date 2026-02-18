package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/storage"
	"github.com/sipeed/picoclaw/pkg/tools"
)

// StepType identifies what kind of step to execute.
type StepType string

const (
	StepPrompt    StepType = "prompt"
	StepTool      StepType = "tool"
	StepCondition StepType = "condition"
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
			output, stepErr = e.executePrompt(ctx, step, sessionKey, results)
		case StepTool:
			output, stepErr = e.executeTool(ctx, step, results)
		case StepCondition:
			matched, evalErr := e.evaluateCondition(step, results)
			if evalErr != nil {
				stepErr = evalErr
			} else if !matched {
				conditionSkip = true
				output = "condition: false — skipping next steps"
			} else {
				output = "condition: true — continuing"
			}
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
func (e *Engine) executePrompt(ctx context.Context, step Step, sessionKey string, prevResults []StepResult) (string, error) {
	if e.agentLoop == nil {
		return "", fmt.Errorf("agent loop not available")
	}

	var cfg PromptConfig
	if err := json.Unmarshal(step.Config, &cfg); err != nil {
		return "", fmt.Errorf("parsing prompt config: %w", err)
	}

	message := interpolate(cfg.Message, prevResults)

	if cfg.Model != "" {
		return e.agentLoop.ProcessDirectWithModel(ctx, message, sessionKey, cfg.Model)
	}
	return e.agentLoop.ProcessDirect(ctx, message, sessionKey)
}

// executeTool calls a registered tool directly.
func (e *Engine) executeTool(ctx context.Context, step Step, prevResults []StepResult) (string, error) {
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
			args[k] = interpolate(s, prevResults)
		} else {
			args[k] = v
		}
	}

	return e.tools.Execute(ctx, cfg.ToolName, args)
}

// evaluateCondition checks a condition against previous step outputs.
func (e *Engine) evaluateCondition(step Step, prevResults []StepResult) (bool, error) {
	var cfg ConditionConfig
	if err := json.Unmarshal(step.Config, &cfg); err != nil {
		return false, fmt.Errorf("parsing condition config: %w", err)
	}

	subject := interpolate(cfg.Reference, prevResults)
	value := interpolate(cfg.Value, prevResults)

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

// templatePattern matches {{step.N.output}} or {{step.N.error}}
var templatePattern = regexp.MustCompile(`\{\{step\.(\d+)\.(output|error)\}\}`)

// interpolate replaces {{step.N.output}} and {{step.N.error}} placeholders
// with values from previous step results. Step indices are 1-based.
func interpolate(text string, results []StepResult) string {
	return templatePattern.ReplaceAllStringFunc(text, func(match string) string {
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
}
