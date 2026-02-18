package mcp

import (
	"context"
	"fmt"
	"strings"
)

// MCPTool wraps an MCP server tool as a tools.Tool interface implementation
type MCPTool struct {
	client   *Client
	toolInfo MCPToolInfo
	prefix   string // server name prefix for uniqueness
}

// NewMCPTool creates a tool proxy for an MCP server tool
func NewMCPTool(client *Client, info MCPToolInfo) *MCPTool {
	return &MCPTool{
		client:   client,
		toolInfo: info,
		prefix:   client.Name(),
	}
}

// Name returns the tool name, prefixed with the MCP server name for uniqueness
func (t *MCPTool) Name() string {
	return fmt.Sprintf("mcp_%s_%s", t.prefix, t.toolInfo.Name)
}

// Description returns the tool description
func (t *MCPTool) Description() string {
	desc := t.toolInfo.Description
	if desc == "" {
		desc = fmt.Sprintf("Tool from MCP server %q", t.prefix)
	}
	return fmt.Sprintf("[MCP:%s] %s", t.prefix, desc)
}

// Parameters returns the JSON Schema parameters for the tool
func (t *MCPTool) Parameters() map[string]interface{} {
	if t.toolInfo.InputSchema != nil {
		return t.toolInfo.InputSchema
	}
	// Fallback: empty object schema
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

// Execute calls the MCP server's tool and returns the result
func (t *MCPTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	if !t.client.Connected() {
		return "", fmt.Errorf("MCP server %q is not connected", t.prefix)
	}

	result, err := t.client.CallTool(ctx, t.toolInfo.Name, args)
	if err != nil {
		return "", fmt.Errorf("MCP tool %q call failed: %w", t.toolInfo.Name, err)
	}

	if result.IsError {
		// Collect error text
		var errTexts []string
		for _, c := range result.Content {
			if c.Text != "" {
				errTexts = append(errTexts, c.Text)
			}
		}
		return "", fmt.Errorf("MCP tool error: %s", strings.Join(errTexts, "; "))
	}

	// Collect text content
	var texts []string
	for _, c := range result.Content {
		if c.Type == "text" && c.Text != "" {
			texts = append(texts, c.Text)
		}
	}

	if len(texts) == 0 {
		return "(no output)", nil
	}
	return strings.Join(texts, "\n"), nil
}
