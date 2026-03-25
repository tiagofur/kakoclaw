package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/sipeed/makoclaw/pkg/auth"
)

type ClaudeProvider struct {
	client      *anthropic.Client
	tokenSource func() (string, error)
}

func NewClaudeProvider(token string) *ClaudeProvider {
	client := anthropic.NewClient(
		option.WithAuthToken(token),
		option.WithBaseURL("https://api.anthropic.com"),
	)
	return &ClaudeProvider{client: &client}
}

func NewClaudeProviderWithTokenSource(token string, tokenSource func() (string, error)) *ClaudeProvider {
	p := NewClaudeProvider(token)
	p.tokenSource = tokenSource
	return p
}

func (p *ClaudeProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	var opts []option.RequestOption
	if p.tokenSource != nil {
		tok, err := p.tokenSource()
		if err != nil {
			return nil, fmt.Errorf("refreshing token: %w", err)
		}
		opts = append(opts, option.WithAuthToken(tok))
	}

	params, err := buildClaudeParams(messages, tools, model, options)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Messages.New(ctx, params, opts...)
	if err != nil {
		return nil, fmt.Errorf("claude API call: %w", err)
	}

	return parseClaudeResponse(resp), nil
}

func (p *ClaudeProvider) GetDefaultModel() string {
	return "claude-sonnet-4-5-20250929"
}

func (p *ClaudeProvider) ChatStream(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (<-chan StreamChunk, error) {
	var opts []option.RequestOption
	if p.tokenSource != nil {
		tok, err := p.tokenSource()
		if err != nil {
			return nil, fmt.Errorf("refreshing token: %w", err)
		}
		opts = append(opts, option.WithAuthToken(tok))
	}

	params, err := buildClaudeParams(messages, tools, model, options)
	if err != nil {
		return nil, err
	}

	stream := p.client.Messages.NewStreaming(ctx, params, opts...)
	ch := make(chan StreamChunk, 64)

	go func() {
		defer close(ch)
		defer stream.Close()

		usage := &UsageInfo{}
		finishReason := "stop"

		for stream.Next() {
			event := stream.Current()

			switch event.Type {
			case "message_start":
				usage.PromptTokens = int(event.Message.Usage.InputTokens)
				usage.CompletionTokens = int(event.Message.Usage.OutputTokens)
				usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
			case "message_delta":
				usage.CompletionTokens = int(event.Usage.OutputTokens)
				usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
				if event.Delta.StopReason != "" {
					finishReason = mapClaudeStopReason(string(event.Delta.StopReason))
				}
			case "content_block_delta":
				chunk := StreamChunk{}
				switch event.Delta.Type {
				case "text_delta":
					chunk.Content = event.Delta.Text
				case "thinking_delta":
					chunk.ThinkingDelta = event.Delta.Thinking
				case "input_json_delta":
					chunk.ToolCalls = indexedToolCallChunk(int(event.Index), ToolCall{
						Type:     "function",
						Function: &FunctionCall{Arguments: event.Delta.PartialJSON},
					})
				}

				if chunk.Content != "" || chunk.ThinkingDelta != "" || len(chunk.ToolCalls) > 0 {
					select {
					case ch <- chunk:
					case <-ctx.Done():
						return
					}
				}
			case "content_block_start":
				if event.ContentBlock.Type != "tool_use" {
					continue
				}

				chunk := StreamChunk{
					ToolCalls: indexedToolCallChunk(int(event.Index), ToolCall{
						ID:   event.ContentBlock.ID,
						Name: event.ContentBlock.Name,
						Type: "function",
					}),
				}

				select {
				case ch <- chunk:
				case <-ctx.Done():
					return
				}
			case "message_stop":
				finalChunk := StreamChunk{
					Done:         true,
					FinishReason: finishReason,
					Usage:        usage,
				}
				select {
				case ch <- finalChunk:
				case <-ctx.Done():
				}
				return
			}
		}

		if err := stream.Err(); err != nil {
			if ctx.Err() != nil {
				return
			}
			select {
			case ch <- StreamChunk{Done: true, FinishReason: "error", Error: err.Error()}:
			case <-ctx.Done():
			}
			return
		}

		select {
		case ch <- StreamChunk{Done: true, FinishReason: finishReason, Usage: usage}:
		case <-ctx.Done():
		}
	}()

	return ch, nil
}

func buildClaudeParams(messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (anthropic.MessageNewParams, error) {
	var system []anthropic.TextBlockParam
	var anthropicMessages []anthropic.MessageParam

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			system = append(system, anthropic.TextBlockParam{Text: msg.Content})
		case "user":
			if msg.ToolCallID != "" {
				anthropicMessages = append(anthropicMessages,
					anthropic.NewUserMessage(anthropic.NewToolResultBlock(msg.ToolCallID, msg.Content, false)),
				)
			} else {
				anthropicMessages = append(anthropicMessages,
					anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)),
				)
			}
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				var blocks []anthropic.ContentBlockParamUnion
				if msg.Content != "" {
					blocks = append(blocks, anthropic.NewTextBlock(msg.Content))
				}
				for _, tc := range msg.ToolCalls {
					blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, tc.Arguments, tc.Name))
				}
				anthropicMessages = append(anthropicMessages, anthropic.NewAssistantMessage(blocks...))
			} else {
				anthropicMessages = append(anthropicMessages,
					anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)),
				)
			}
		case "tool":
			anthropicMessages = append(anthropicMessages,
				anthropic.NewUserMessage(anthropic.NewToolResultBlock(msg.ToolCallID, msg.Content, false)),
			)
		}
	}

	maxTokens := int64(4096)
	if mt, ok := options["max_tokens"].(int); ok {
		maxTokens = int64(mt)
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		Messages:  anthropicMessages,
		MaxTokens: maxTokens,
	}

	if len(system) > 0 {
		params.System = system
	}

	if temp, ok := options["temperature"].(float64); ok {
		params.Temperature = anthropic.Float(temp)
	}

	if enabled, ok := options["extended_thinking"].(bool); ok && enabled {
		budgetTokens := int64(1024)
		switch budget := options["thinking_budget_tokens"].(type) {
		case int:
			if budget > 0 {
				budgetTokens = int64(budget)
			}
		case int64:
			if budget > 0 {
				budgetTokens = budget
			}
		case float64:
			if budget > 0 {
				budgetTokens = int64(budget)
			}
		}
		if budgetTokens < 1024 {
			budgetTokens = 1024
		}
		params.Thinking = anthropic.ThinkingConfigParamOfEnabled(budgetTokens)
	}

	if len(tools) > 0 {
		params.Tools = translateToolsForClaude(tools)
	}

	return params, nil
}

func translateToolsForClaude(tools []ToolDefinition) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, 0, len(tools))
	for _, t := range tools {
		tool := anthropic.ToolParam{
			Name: t.Function.Name,
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: t.Function.Parameters["properties"],
			},
		}
		if desc := t.Function.Description; desc != "" {
			tool.Description = anthropic.String(desc)
		}
		if req, ok := t.Function.Parameters["required"].([]interface{}); ok {
			required := make([]string, 0, len(req))
			for _, r := range req {
				if s, ok := r.(string); ok {
					required = append(required, s)
				}
			}
			tool.InputSchema.Required = required
		}
		result = append(result, anthropic.ToolUnionParam{OfTool: &tool})
	}
	return result
}

func parseClaudeResponse(resp *anthropic.Message) *LLMResponse {
	var content string
	var toolCalls []ToolCall

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			tb := block.AsText()
			content += tb.Text
		case "tool_use":
			tu := block.AsToolUse()
			var args map[string]interface{}
			if err := json.Unmarshal(tu.Input, &args); err != nil {
				args = map[string]interface{}{"raw": string(tu.Input)}
			}
			toolCalls = append(toolCalls, ToolCall{
				ID:        tu.ID,
				Name:      tu.Name,
				Arguments: args,
			})
		}
	}

	finishReason := "stop"
	switch resp.StopReason {
	case anthropic.StopReasonToolUse:
		finishReason = "tool_calls"
	case anthropic.StopReasonMaxTokens:
		finishReason = "length"
	case anthropic.StopReasonEndTurn:
		finishReason = "stop"
	}

	return &LLMResponse{
		Content:      content,
		ToolCalls:    toolCalls,
		FinishReason: finishReason,
		Usage: &UsageInfo{
			PromptTokens:     int(resp.Usage.InputTokens),
			CompletionTokens: int(resp.Usage.OutputTokens),
			TotalTokens:      int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		},
	}
}

func createClaudeTokenSource() func() (string, error) {
	return func() (string, error) {
		cred, err := auth.GetCredential("anthropic")
		if err != nil {
			return "", fmt.Errorf("loading auth credentials: %w", err)
		}
		if cred == nil {
			return "", fmt.Errorf("no credentials for anthropic. Run: makoclaw auth login --provider anthropic")
		}
		return cred.AccessToken, nil
	}
}

func indexedToolCallChunk(index int, toolCall ToolCall) []ToolCall {
	if index < 0 {
		index = 0
	}
	chunk := make([]ToolCall, index+1)
	chunk[index] = toolCall
	return chunk
}

func mapClaudeStopReason(stopReason string) string {
	switch strings.TrimSpace(stopReason) {
	case "tool_use":
		return "tool_calls"
	case "max_tokens":
		return "length"
	case "end_turn", "stop_sequence", "pause_turn", "refusal", "":
		return "stop"
	default:
		return stopReason
	}
}
