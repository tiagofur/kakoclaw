// PicoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sipeed/picoclaw/pkg/auth"
	"github.com/sipeed/picoclaw/pkg/config"
)

type HTTPProvider struct {
	apiKey     string
	apiBase    string
	httpClient *http.Client
}

func NewHTTPProvider(apiKey, apiBase, proxy string) *HTTPProvider {
	client := &http.Client{
		Timeout: 0,
	}

	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	return &HTTPProvider{
		apiKey:     apiKey,
		apiBase:    apiBase,
		httpClient: client,
	}
}

func (p *HTTPProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	if p.apiBase == "" {
		return nil, fmt.Errorf("API base not configured")
	}

	// Strip provider prefix from model name (e.g., moonshot/kimi-k2.5 -> kimi-k2.5)
	if idx := strings.Index(model, "/"); idx != -1 {
		prefix := model[:idx]
		if prefix == "moonshot" || prefix == "nvidia" {
			model = model[idx+1:]
		}
	}

	requestBody := map[string]interface{}{
		"model":    model,
		"messages": messages,
	}

	if len(tools) > 0 {
		requestBody["tools"] = tools
		requestBody["tool_choice"] = "auto"
	}

	if maxTokens, ok := options["max_tokens"].(int); ok {
		lowerModel := strings.ToLower(model)
		if strings.Contains(lowerModel, "glm") || strings.Contains(lowerModel, "o1") {
			requestBody["max_completion_tokens"] = maxTokens
		} else {
			requestBody["max_tokens"] = maxTokens
		}
	}

	if temperature, ok := options["temperature"].(float64); ok {
		lowerModel := strings.ToLower(model)
		// Kimi k2 models only support temperature=1
		if strings.Contains(lowerModel, "kimi") && strings.Contains(lowerModel, "k2") {
			requestBody["temperature"] = 1.0
		} else {
			requestBody["temperature"] = temperature
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.apiBase+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	return p.parseResponse(body)
}

func (p *HTTPProvider) parseResponse(body []byte) (*LLMResponse, error) {
	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function *struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage *UsageInfo `json:"usage"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return &LLMResponse{
			Content:      "",
			FinishReason: "stop",
		}, nil
	}

	choice := apiResponse.Choices[0]

	toolCalls := make([]ToolCall, 0, len(choice.Message.ToolCalls))
	for _, tc := range choice.Message.ToolCalls {
		arguments := make(map[string]interface{})
		name := ""

		// Handle OpenAI format with nested function object
		if tc.Type == "function" && tc.Function != nil {
			name = tc.Function.Name
			if tc.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &arguments); err != nil {
					arguments["raw"] = tc.Function.Arguments
				}
			}
		} else if tc.Function != nil {
			// Legacy format without type field
			name = tc.Function.Name
			if tc.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &arguments); err != nil {
					arguments["raw"] = tc.Function.Arguments
				}
			}
		}

		toolCalls = append(toolCalls, ToolCall{
			ID:        tc.ID,
			Name:      name,
			Arguments: arguments,
		})
	}

	return &LLMResponse{
		Content:      choice.Message.Content,
		ToolCalls:    toolCalls,
		FinishReason: choice.FinishReason,
		Usage:        apiResponse.Usage,
	}, nil
}

func (p *HTTPProvider) GetDefaultModel() string {
	return ""
}

// ChatStream implements StreamingLLMProvider for OpenAI-compatible SSE streaming.
func (p *HTTPProvider) ChatStream(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (<-chan StreamChunk, error) {
	if p.apiBase == "" {
		return nil, fmt.Errorf("API base not configured")
	}

	// Strip provider prefix from model name (same as Chat)
	if idx := strings.Index(model, "/"); idx != -1 {
		prefix := model[:idx]
		if prefix == "moonshot" || prefix == "nvidia" {
			model = model[idx+1:]
		}
	}

	requestBody := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   true,
	}

	if len(tools) > 0 {
		requestBody["tools"] = tools
		requestBody["tool_choice"] = "auto"
	}

	if maxTokens, ok := options["max_tokens"].(int); ok {
		lowerModel := strings.ToLower(model)
		if strings.Contains(lowerModel, "glm") || strings.Contains(lowerModel, "o1") {
			requestBody["max_completion_tokens"] = maxTokens
		} else {
			requestBody["max_tokens"] = maxTokens
		}
	}

	if temperature, ok := options["temperature"].(float64); ok {
		lowerModel := strings.ToLower(model)
		if strings.Contains(lowerModel, "kimi") && strings.Contains(lowerModel, "k2") {
			requestBody["temperature"] = 1.0
		} else {
			requestBody["temperature"] = temperature
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.apiBase+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	ch := make(chan StreamChunk, 64)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		// Allow lines up to 256KB for large tool call arguments
		scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)

		for scanner.Scan() {
			line := scanner.Text()

			// SSE format: lines starting with "data: "
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- StreamChunk{Done: true, FinishReason: "stop"}
				return
			}

			var sseEvent struct {
				Choices []struct {
					Delta struct {
						Content   string `json:"content"`
						ToolCalls []struct {
							Index    int    `json:"index"`
							ID       string `json:"id"`
							Type     string `json:"type"`
							Function *struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						} `json:"tool_calls"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
				Usage *UsageInfo `json:"usage"`
			}

			if err := json.Unmarshal([]byte(data), &sseEvent); err != nil {
				continue
			}

			if len(sseEvent.Choices) == 0 {
				continue
			}

			choice := sseEvent.Choices[0]
			chunk := StreamChunk{
				Content: choice.Delta.Content,
				Usage:   sseEvent.Usage,
			}

			// Handle tool calls in delta
			for _, tc := range choice.Delta.ToolCalls {
				toolCall := ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
				}
				if tc.Function != nil {
					toolCall.Name = tc.Function.Name
					// Tool call arguments come in fragments; pass as-is
					if tc.Function.Arguments != "" {
						toolCall.Function = &FunctionCall{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						}
					}
				}
				chunk.ToolCalls = append(chunk.ToolCalls, toolCall)
			}

			if choice.FinishReason != nil {
				chunk.FinishReason = *choice.FinishReason
				chunk.Done = true
			}

			select {
			case ch <- chunk:
			case <-ctx.Done():
				return
			}

			if chunk.Done {
				return
			}
		}

		// If scanner exits without [DONE], send a final done chunk
		ch <- StreamChunk{Done: true, FinishReason: "stop"}
	}()

	return ch, nil
}

func createClaudeAuthProvider() (LLMProvider, error) {
	cred, err := auth.GetCredential("anthropic")
	if err != nil {
		return nil, fmt.Errorf("loading auth credentials: %w", err)
	}
	if cred == nil {
		return nil, fmt.Errorf("no credentials for anthropic. Run: picoclaw auth login --provider anthropic")
	}
	return NewClaudeProviderWithTokenSource(cred.AccessToken, createClaudeTokenSource()), nil
}

func createCodexAuthProvider() (LLMProvider, error) {
	cred, err := auth.GetCredential("openai")
	if err != nil {
		return nil, fmt.Errorf("loading auth credentials: %w", err)
	}
	if cred == nil {
		return nil, fmt.Errorf("no credentials for openai. Run: picoclaw auth login --provider openai")
	}
	return NewCodexProviderWithTokenSource(cred.AccessToken, cred.AccountID, createCodexTokenSource()), nil
}

func CreateProvider(cfg *config.Config) (LLMProvider, error) {
	model := cfg.Agents.Defaults.Model
	providerName := strings.ToLower(cfg.Agents.Defaults.Provider)

	var apiKey, apiBase, proxy string

	lowerModel := strings.ToLower(model)

	// Issue #43: Support explicit provider/model syntax
	// e.g., "openai/gpt-4" explicitly selects openai provider with gpt-4 model
	if idx := strings.Index(model, "/"); idx > 0 {
		explicitProvider := strings.ToLower(model[:idx])
		actualModel := model[idx+1:]

		switch explicitProvider {
		case "openai", "anthropic", "openrouter", "groq", "zhipu", "gemini", "moonshot", "nvidia", "ollama":
			providerName = explicitProvider
			model = actualModel
			lowerModel = strings.ToLower(model)
		}
	}

	// First, try to use explicitly configured provider
	if providerName != "" {
		switch providerName {
		case "mock":
			// Mock provider for testing without external APIs
			return NewMockProvider(), nil
		case "groq":
			if cfg.Providers.Groq.APIKey != "" {
				apiKey = cfg.Providers.Groq.APIKey
				apiBase = cfg.Providers.Groq.APIBase
				if apiBase == "" {
					apiBase = "https://api.groq.com/openai/v1"
				}
			}
		case "openai", "gpt":
			if cfg.Providers.OpenAI.APIKey != "" || cfg.Providers.OpenAI.AuthMethod != "" {
				if cfg.Providers.OpenAI.AuthMethod == "oauth" || cfg.Providers.OpenAI.AuthMethod == "token" {
					return createCodexAuthProvider()
				}
				apiKey = cfg.Providers.OpenAI.APIKey
				apiBase = cfg.Providers.OpenAI.APIBase
				if apiBase == "" {
					apiBase = "https://api.openai.com/v1"
				}
			}
		case "anthropic", "claude":
			if cfg.Providers.Anthropic.APIKey != "" || cfg.Providers.Anthropic.AuthMethod != "" {
				if cfg.Providers.Anthropic.AuthMethod == "oauth" || cfg.Providers.Anthropic.AuthMethod == "token" {
					return createClaudeAuthProvider()
				}
				apiKey = cfg.Providers.Anthropic.APIKey
				apiBase = cfg.Providers.Anthropic.APIBase
				if apiBase == "" {
					apiBase = "https://api.anthropic.com/v1"
				}
			}
		case "openrouter":
			if cfg.Providers.OpenRouter.APIKey != "" {
				apiKey = cfg.Providers.OpenRouter.APIKey
				if cfg.Providers.OpenRouter.APIBase != "" {
					apiBase = cfg.Providers.OpenRouter.APIBase
				} else {
					apiBase = "https://openrouter.ai/api/v1"
				}
			}
		case "zhipu", "glm":
			if cfg.Providers.Zhipu.APIKey != "" {
				apiKey = cfg.Providers.Zhipu.APIKey
				apiBase = cfg.Providers.Zhipu.APIBase
				if apiBase == "" {
					apiBase = "https://open.bigmodel.cn/api/paas/v4"
				}
			}
		case "gemini", "google":
			if cfg.Providers.Gemini.APIKey != "" {
				apiKey = cfg.Providers.Gemini.APIKey
				apiBase = cfg.Providers.Gemini.APIBase
				if apiBase == "" {
					apiBase = "https://generativelanguage.googleapis.com/v1beta"
				}
			}
		case "vllm":
			if cfg.Providers.VLLM.APIBase != "" {
				apiKey = cfg.Providers.VLLM.APIKey
				apiBase = cfg.Providers.VLLM.APIBase
			}
		case "ollama":
			// Issue #75: Ollama local LLM support
			apiBase = cfg.Providers.Ollama.APIBase
			if apiBase == "" {
				apiBase = "http://localhost:11434"
			}
			return NewOllamaProvider(apiBase), nil
		}
	}

	// Fallback: detect provider from model name
	if apiKey == "" && apiBase == "" {
		switch {
		case (strings.Contains(lowerModel, "kimi") || strings.Contains(lowerModel, "moonshot") || strings.HasPrefix(model, "moonshot/")) && cfg.Providers.Moonshot.APIKey != "":
			apiKey = cfg.Providers.Moonshot.APIKey
			apiBase = cfg.Providers.Moonshot.APIBase
			proxy = cfg.Providers.Moonshot.Proxy
			if apiBase == "" {
				apiBase = "https://api.moonshot.cn/v1"
			}

		case strings.HasPrefix(model, "openrouter/") || strings.HasPrefix(model, "anthropic/") || strings.HasPrefix(model, "openai/") || strings.HasPrefix(model, "meta-llama/") || strings.HasPrefix(model, "deepseek/") || strings.HasPrefix(model, "google/"):
			apiKey = cfg.Providers.OpenRouter.APIKey
			proxy = cfg.Providers.OpenRouter.Proxy
			if cfg.Providers.OpenRouter.APIBase != "" {
				apiBase = cfg.Providers.OpenRouter.APIBase
			} else {
				apiBase = "https://openrouter.ai/api/v1"
			}

		case (strings.Contains(lowerModel, "claude") || strings.HasPrefix(model, "anthropic/")) && (cfg.Providers.Anthropic.APIKey != "" || cfg.Providers.Anthropic.AuthMethod != ""):
			if cfg.Providers.Anthropic.AuthMethod == "oauth" || cfg.Providers.Anthropic.AuthMethod == "token" {
				return createClaudeAuthProvider()
			}
			apiKey = cfg.Providers.Anthropic.APIKey
			apiBase = cfg.Providers.Anthropic.APIBase
			proxy = cfg.Providers.Anthropic.Proxy
			if apiBase == "" {
				apiBase = "https://api.anthropic.com/v1"
			}

		case (strings.Contains(lowerModel, "gpt") || strings.HasPrefix(model, "openai/")) && (cfg.Providers.OpenAI.APIKey != "" || cfg.Providers.OpenAI.AuthMethod != ""):
			if cfg.Providers.OpenAI.AuthMethod == "oauth" || cfg.Providers.OpenAI.AuthMethod == "token" {
				return createCodexAuthProvider()
			}
			apiKey = cfg.Providers.OpenAI.APIKey
			apiBase = cfg.Providers.OpenAI.APIBase
			proxy = cfg.Providers.OpenAI.Proxy
			if apiBase == "" {
				apiBase = "https://api.openai.com/v1"
			}

		case (strings.Contains(lowerModel, "gemini") || strings.HasPrefix(model, "google/")) && cfg.Providers.Gemini.APIKey != "":
			apiKey = cfg.Providers.Gemini.APIKey
			apiBase = cfg.Providers.Gemini.APIBase
			proxy = cfg.Providers.Gemini.Proxy
			if apiBase == "" {
				apiBase = "https://generativelanguage.googleapis.com/v1beta"
			}

		case (strings.Contains(lowerModel, "glm") || strings.Contains(lowerModel, "zhipu") || strings.Contains(lowerModel, "zai")) && cfg.Providers.Zhipu.APIKey != "":
			apiKey = cfg.Providers.Zhipu.APIKey
			apiBase = cfg.Providers.Zhipu.APIBase
			proxy = cfg.Providers.Zhipu.Proxy
			if apiBase == "" {
				apiBase = "https://open.bigmodel.cn/api/paas/v4"
			}

		case (strings.Contains(lowerModel, "groq") || strings.HasPrefix(model, "groq/")) && cfg.Providers.Groq.APIKey != "":
			apiKey = cfg.Providers.Groq.APIKey
			apiBase = cfg.Providers.Groq.APIBase
			proxy = cfg.Providers.Groq.Proxy
			if apiBase == "" {
				apiBase = "https://api.groq.com/openai/v1"
			}

		case (strings.Contains(lowerModel, "nvidia") || strings.HasPrefix(model, "nvidia/")) && cfg.Providers.Nvidia.APIKey != "":
			apiKey = cfg.Providers.Nvidia.APIKey
			apiBase = cfg.Providers.Nvidia.APIBase
			proxy = cfg.Providers.Nvidia.Proxy
			if apiBase == "" {
				apiBase = "https://integrate.api.nvidia.com/v1"
			}

		case cfg.Providers.VLLM.APIBase != "":
			apiKey = cfg.Providers.VLLM.APIKey
			apiBase = cfg.Providers.VLLM.APIBase
			proxy = cfg.Providers.VLLM.Proxy

		case (strings.Contains(lowerModel, "ollama") || strings.HasPrefix(model, "ollama/")) && cfg.Providers.Ollama.APIBase != "":
			// Issue #75: Ollama support
			return NewOllamaProvider(cfg.Providers.Ollama.APIBase), nil

		default:
			if cfg.Providers.OpenRouter.APIKey != "" {
				apiKey = cfg.Providers.OpenRouter.APIKey
				proxy = cfg.Providers.OpenRouter.Proxy
				if cfg.Providers.OpenRouter.APIBase != "" {
					apiBase = cfg.Providers.OpenRouter.APIBase
				} else {
					apiBase = "https://openrouter.ai/api/v1"
				}
			} else {
				return nil, fmt.Errorf("no API key configured for model: %s", model)
			}
		}
	}

	if apiKey == "" && !strings.HasPrefix(model, "bedrock/") {
		return nil, fmt.Errorf("no API key configured for provider (model: %s)", model)
	}

	if apiBase == "" {
		return nil, fmt.Errorf("no API base configured for provider (model: %s)", model)
	}

	return NewHTTPProvider(apiKey, apiBase, proxy), nil
}

// GetProviderForModel returns the provider name for a given model (issue #43)
// Supports explicit provider/model syntax (e.g., "openai/gpt-4")
func GetProviderForModel(model string) (provider string, actualModel string) {
	// Check for explicit provider prefix
	if idx := strings.Index(model, "/"); idx > 0 {
		explicitProvider := strings.ToLower(model[:idx])
		switch explicitProvider {
		case "openai", "anthropic", "openrouter", "groq", "zhipu", "gemini", "moonshot", "nvidia", "ollama":
			return explicitProvider, model[idx+1:]
		}
	}

	// Auto-detect from model name
	lowerModel := strings.ToLower(model)
	switch {
	case strings.HasPrefix(lowerModel, "gpt-"), strings.HasPrefix(lowerModel, "text-"), strings.HasPrefix(lowerModel, "o1"):
		return "openai", model
	case strings.Contains(lowerModel, "claude"):
		return "anthropic", model
	case strings.Contains(lowerModel, "kimi") || strings.Contains(lowerModel, "moonshot"):
		return "moonshot", model
	case strings.Contains(lowerModel, "gemini") || strings.Contains(lowerModel, "gemma"):
		return "gemini", model
	case strings.Contains(lowerModel, "glm") || strings.Contains(lowerModel, "zai"):
		return "zhipu", model
	case strings.Contains(lowerModel, "llama") && !strings.Contains(lowerModel, "nvidia"):
		// Llama models could be from various providers
		return "openrouter", model
	case strings.Contains(lowerModel, "mixtral"), strings.Contains(lowerModel, "mistral"):
		return "openrouter", model
	default:
		// Default to openrouter for unknown models
		return "openrouter", model
	}
}
