package providers

import (
	"context"
	"fmt"
)

// MockProvider is a simple provider for testing without external API calls
type MockProvider struct{}

// NewMockProvider creates a new mock provider
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// Chat implements the LLMProvider interface with mock responses
func (m *MockProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	// Check if there are tool calls requested
	if len(tools) > 0 {
		// Return a mock tool call response
		return &LLMResponse{
			Content: "I'll help you with that. Let me use the available tools.",
			ToolCalls: []ToolCall{
				{
					ID:   "mock-call-1",
					Type: "function",
					Function: &FunctionCall{
						Name:      "message",
						Arguments: `{"channel": "cli", "text": "Mock response from agent"}`,
					},
				},
			},
			FinishReason: "tool_calls",
			Usage: &UsageInfo{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}, nil
	}

	// Return a simple mock text response
	return &LLMResponse{
		Content:      fmt.Sprintf("Mock response to: %s", getLastMessageContent(messages)),
		FinishReason: "stop",
		Usage: &UsageInfo{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}, nil
}

// GetDefaultModel returns the default model for mock provider
func (m *MockProvider) GetDefaultModel() string {
	return "mock"
}

// getLastMessageContent is a helper to get the content of the last message
func getLastMessageContent(messages []Message) string {
	if len(messages) == 0 {
		return ""
	}
	return messages[len(messages)-1].Content
}
