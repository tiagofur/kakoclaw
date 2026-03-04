package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sipeed/makoclaw/pkg/logger"
)

// OllamaProvider implements LLMProvider for local Ollama instance
type OllamaProvider struct {
	baseURL string
	client  *http.Client
}

// OllamaMessage represents a message in Ollama format
type OllamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []OllamaToolCall `json:"tool_calls,omitempty"`
}

// OllamaToolCall represents a tool call in Ollama format
type OllamaToolCall struct {
	Function OllamaFunctionCall `json:"function"`
}

// OllamaFunctionCall represents a function call in Ollama format
type OllamaFunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// OllamaRequest represents the request body for Ollama API
type OllamaRequest struct {
	Model    string           `json:"model"`
	Messages []OllamaMessage  `json:"messages"`
	Stream   bool             `json:"stream"`
	Options  OllamaOptions    `json:"options,omitempty"`
	Tools    []ToolDefinition `json:"tools,omitempty"`
}

// OllamaOptions represents model-specific options
type OllamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

// OllamaResponse represents the response from Ollama API
type OllamaResponse struct {
	Model   string `json:"model"`
	Message struct {
		Role      string           `json:"role"`
		Content   string           `json:"content"`
		ToolCalls []OllamaToolCall `json:"tool_calls,omitempty"`
	} `json:"message"`
	Done            bool `json:"done"`
	EvalCount       int  `json:"eval_count"`
	PromptEvalCount int  `json:"prompt_eval_count"`
}

// mustMarshalJSON marshals v to JSON string, returning "{}" on error
func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaProvider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				IdleConnTimeout: 30 * time.Second,
			},
		},
	}
}

// Chat implements LLMProvider.Chat
func (p *OllamaProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	logger.InfoCF("ollama", "Sending chat request", map[string]interface{}{
		"model":    model,
		"messages": len(messages),
		"tools":    len(tools),
	})

	// Convert messages to Ollama format
	ollamaMessages := make([]OllamaMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = OllamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Build request
	reqBody := OllamaRequest{
		Model:    model,
		Messages: ollamaMessages,
		Stream:   false,
	}

	// Add tools if provided (requires Ollama 0.3.0+)
	if len(tools) > 0 {
		reqBody.Tools = tools
	}

	// Add options if provided
	if temp, ok := options["temperature"].(float64); ok {
		reqBody.Options.Temperature = temp
	}
	if maxTokens, ok := options["max_tokens"].(int); ok {
		reqBody.Options.NumPredict = maxTokens
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Build response with tool calls if present
	llmResp := &LLMResponse{
		Content: ollamaResp.Message.Content,
		Usage: &UsageInfo{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		},
	}

	// Convert Ollama tool calls to standard format
	if len(ollamaResp.Message.ToolCalls) > 0 {
		llmResp.ToolCalls = make([]ToolCall, len(ollamaResp.Message.ToolCalls))
		for i, tc := range ollamaResp.Message.ToolCalls {
			llmResp.ToolCalls[i] = ToolCall{
				ID:   fmt.Sprintf("call_%d", i), // Ollama doesn't provide IDs, generate one
				Type: "function",
				Function: &FunctionCall{
					Name:      tc.Function.Name,
					Arguments: mustMarshalJSON(tc.Function.Arguments),
				},
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
		}
	}

	return llmResp, nil
}

// GetDefaultModel implements LLMProvider.GetDefaultModel
func (p *OllamaProvider) GetDefaultModel() string {
	return "llama3.2"
}

// ListModels returns available models from Ollama
func (p *OllamaProvider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API error (status %d)", resp.StatusCode)
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}

	return models, nil
}

// ChatStream implements StreamingLLMProvider for Ollama streaming responses.
func (p *OllamaProvider) ChatStream(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (<-chan StreamChunk, error) {
	logger.InfoCF("ollama", "Sending streaming chat request", map[string]interface{}{
		"model":    model,
		"messages": len(messages),
	})

	// Convert messages to Ollama format
	ollamaMessages := make([]OllamaMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = OllamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	reqBody := OllamaRequest{
		Model:    model,
		Messages: ollamaMessages,
		Stream:   true, // Enable streaming
	}

	// Add tools if provided (requires Ollama 0.3.0+)
	// Note: Streaming with tools may return tool calls only in the final chunk
	if len(tools) > 0 {
		reqBody.Tools = tools
	}

	if temp, ok := options["temperature"].(float64); ok {
		reqBody.Options.Temperature = temp
	}
	if maxTokens, ok := options["max_tokens"].(int); ok {
		reqBody.Options.NumPredict = maxTokens
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	ch := make(chan StreamChunk, 64)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			var ollamaResp OllamaResponse
			if err := json.Unmarshal([]byte(line), &ollamaResp); err != nil {
				logger.WarnCF("ollama", "Failed to parse streaming response", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}

			chunk := StreamChunk{
				Content: ollamaResp.Message.Content,
			}

			if ollamaResp.Done {
				chunk.Done = true
				chunk.FinishReason = "stop"
				chunk.Usage = &UsageInfo{
					PromptTokens:     ollamaResp.PromptEvalCount,
					CompletionTokens: ollamaResp.EvalCount,
					TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
				}

				// Convert tool calls if present in final chunk
				if len(ollamaResp.Message.ToolCalls) > 0 {
					chunk.ToolCalls = make([]ToolCall, len(ollamaResp.Message.ToolCalls))
					for i, tc := range ollamaResp.Message.ToolCalls {
						chunk.ToolCalls[i] = ToolCall{
							ID:   fmt.Sprintf("call_%d", i),
							Type: "function",
							Function: &FunctionCall{
								Name:      tc.Function.Name,
								Arguments: mustMarshalJSON(tc.Function.Arguments),
							},
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						}
					}
					chunk.FinishReason = "tool_calls"
				}
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

		// Check for scanner errors (network failures, etc.)
		if err := scanner.Err(); err != nil {
			logger.ErrorCF("ollama", "Stream scanner error", map[string]interface{}{
				"error": err.Error(),
			})
			select {
			case ch <- StreamChunk{Done: true, FinishReason: "error", Error: err.Error()}:
			case <-ctx.Done():
			}
			return
		}

		// If scanner exits without done, send final chunk
		ch <- StreamChunk{Done: true, FinishReason: "stop"}
	}()

	return ch, nil
}
