package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/providers"
)

// LLMTaskTool delegates a sub-task to a specific LLM model in a single, non-streaming turn.
type LLMTaskTool struct {
	cfg *config.Config
}

// NewLLMTaskTool creates a new LLMTaskTool using the given config.
func NewLLMTaskTool(cfg *config.Config) *LLMTaskTool {
	return &LLMTaskTool{cfg: cfg}
}

func (t *LLMTaskTool) Name() string { return "llm_task" }

func (t *LLMTaskTool) Description() string {
	return "Delegate a sub-task to a specific LLM model. No tools, single turn, text response only."
}

func (t *LLMTaskTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"model": map[string]interface{}{
				"type":        "string",
				"description": "Provider/model string, e.g. openai/gpt-4o-mini",
			},
			"task": map[string]interface{}{
				"type":        "string",
				"description": "The task or question to send to the model",
			},
			"context": map[string]interface{}{
				"type":        "string",
				"description": "Optional background context to prepend",
			},
		},
		"required": []string{"model", "task"},
	}
}

func (t *LLMTaskTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	model, _ := args["model"].(string)
	task, _ := args["task"].(string)
	contextStr, _ := args["context"].(string)

	if model == "" || task == "" {
		return "", fmt.Errorf("llm_task: model and task are required")
	}

	// Build a minimal config with only the fields CreateProvider needs.
	// Copying *t.cfg would trigger the copylocks vet check because Config
	// embeds a sync.RWMutex. We only need Providers (API keys/bases) and
	// Agents.Defaults (provider name + model for selection logic).
	providerCfg := &config.Config{
		Providers: t.cfg.Providers,
		Agents: config.AgentsConfig{
			Defaults: t.cfg.Agents.Defaults,
		},
	}
	providerCfg.Agents.Defaults.Model = model

	provider, err := providers.CreateProvider(providerCfg)
	if err != nil {
		return "", fmt.Errorf("llm_task: cannot create provider for model %q: %w", model, err)
	}

	var messages []providers.Message
	if contextStr != "" {
		messages = append(messages, providers.Message{Role: "system", Content: contextStr})
	}
	messages = append(messages, providers.Message{Role: "user", Content: task})

	// Check if the parent context is already done before creating the timeout.
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("llm_task timeout: context cancelled before call: %w", err)
	}

	ctx30s, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := provider.Chat(ctx30s, messages, nil, model, map[string]interface{}{"max_tokens": 4096})
	if err != nil {
		if ctx30s.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("llm_task timeout after 30s")
		}
		return "", fmt.Errorf("llm_task: provider error: %w", err)
	}

	return resp.Content, nil
}
