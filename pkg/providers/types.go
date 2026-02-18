package providers

import "context"

type ToolCall struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type,omitempty"`
	Function  *FunctionCall          `json:"function,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type LLMResponse struct {
	Content      string     `json:"content"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason"`
	Usage        *UsageInfo `json:"usage,omitempty"`
}

type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type LLMProvider interface {
	Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error)
	GetDefaultModel() string
}

// StreamChunk represents a single token/chunk from a streaming response.
type StreamChunk struct {
	Content      string     `json:"content"` // Text delta (incremental token)
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason"`   // Empty until stream ends, then "stop"/"tool_calls"/"length"
	Done         bool       `json:"done"`            // True when stream is complete
	Usage        *UsageInfo `json:"usage,omitempty"` // Only set on final chunk
}

// StreamingLLMProvider is an optional interface for providers that support token-by-token streaming.
// If a provider implements this interface, the web handler can stream tokens to the client.
// Providers that don't implement it will fall back to the non-streaming Chat() method.
type StreamingLLMProvider interface {
	LLMProvider
	ChatStream(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (<-chan StreamChunk, error)
}

type ToolDefinition struct {
	Type     string                 `json:"type"`
	Function ToolFunctionDefinition `json:"function"`
}

type ToolFunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}
