package tools

import (
	"context"
	"fmt"
	"strings"
)

// BrowserTool provides high-level browser automation operations.
// Delegates to Playwright MCP server when available, falls back to web_fetch for basic ops.
type BrowserTool struct {
	workspace string
	registry  *ToolRegistry
}

func NewBrowserTool(workspace string, registry *ToolRegistry) *BrowserTool {
	return &BrowserTool{
		workspace: workspace,
		registry:  registry,
	}
}

func (t *BrowserTool) SetWorkspace(workspace string) {
	t.workspace = workspace
}

func (t *BrowserTool) Name() string { return "browser" }

func (t *BrowserTool) Description() string {
	return `Browser automation for testing and web interaction.
Actions:
  open       - Navigate to a URL and return page content/title
  screenshot - Take a screenshot of the current page (requires Playwright MCP)
  click      - Click an element by selector (requires Playwright MCP)
  fill       - Fill a form field (requires Playwright MCP)
  evaluate   - Execute JavaScript in the page (requires Playwright MCP)
  test_e2e   - Run end-to-end tests (detects Playwright/Cypress/Selenium in project)
  check_responsive - Test page at multiple viewport sizes (requires Playwright MCP)
  extract    - Extract text/data from a web page
  pdf        - Generate PDF from a page (requires Playwright MCP)
  status     - Check if Playwright MCP is available
Note: Advanced actions require Playwright MCP server. Install via:
  npx @anthropic/mcp-server-playwright
Without Playwright MCP, only open/extract/test_e2e are available via web_fetch.`
}

func (t *BrowserTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"open", "screenshot", "click", "fill", "evaluate", "test_e2e", "check_responsive", "extract", "pdf", "status"},
				"description": "The browser action to perform",
			},
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to navigate to (for open, screenshot, extract, pdf)",
			},
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for click/fill actions",
			},
			"value": map[string]interface{}{
				"type":        "string",
				"description": "Value to fill or JavaScript to evaluate",
			},
			"args": map[string]interface{}{
				"type":        "string",
				"description": "Additional arguments (test pattern for test_e2e, viewport sizes for check_responsive)",
			},
		},
		"required": []string{"action"},
	}
}

func (t *BrowserTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action is required")
	}

	switch action {
	case "status":
		return t.checkPlaywrightMCP(), nil

	case "open", "extract":
		return t.openOrExtract(ctx, args)

	case "test_e2e":
		return t.runE2ETests(ctx, args)

	case "screenshot", "click", "fill", "evaluate", "check_responsive", "pdf":
		return t.playwrightAction(ctx, action, args)

	default:
		return "", fmt.Errorf("unknown browser action: %s", action)
	}
}

func (t *BrowserTool) checkPlaywrightMCP() string {
	if t.registry == nil {
		return "Playwright MCP: Not available (no tool registry)\nBasic web operations available via web_fetch."
	}

	// Check for any MCP playwright tools
	tools := t.registry.List()
	playwrightTools := []string{}
	for _, tool := range tools {
		name := tool.Name()
		if strings.Contains(name, "playwright") || strings.Contains(name, "browser") {
			playwrightTools = append(playwrightTools, name)
		}
	}

	if len(playwrightTools) > 0 {
		return fmt.Sprintf("Playwright MCP: Available (%d tools)\nTools: %s",
			len(playwrightTools), strings.Join(playwrightTools, ", "))
	}

	return `Playwright MCP: Not configured
To enable full browser automation, add to your MCP config:
{
  "playwright": {
    "enabled": true,
    "command": "npx",
    "args": ["-y", "@anthropic/mcp-server-playwright"]
  }
}
Basic web operations (open/extract) are available via web_fetch.`
}

func (t *BrowserTool) openOrExtract(ctx context.Context, args map[string]interface{}) (string, error) {
	url, _ := args["url"].(string)
	if url == "" {
		return "", fmt.Errorf("url is required for open/extract action")
	}

	// Try Playwright MCP first
	if tool := t.findMCPTool("navigate", "goto"); tool != nil {
		result, err := tool.Execute(ctx, map[string]interface{}{"url": url})
		if err == nil {
			return result, nil
		}
	}

	// Fallback to web_fetch
	if t.registry != nil {
		if fetchTool, ok := t.registry.Get("web_fetch"); ok {
			return fetchTool.Execute(ctx, map[string]interface{}{"url": url})
		}
	}

	return "", fmt.Errorf("no web fetch tool available")
}

func (t *BrowserTool) runE2ETests(ctx context.Context, args map[string]interface{}) (string, error) {
	extraArgs, _ := args["args"].(string)

	// Detect test framework in the project
	projectTool := NewProjectTool(t.workspace)
	info := projectTool.detectProject(t.workspace)

	var cmd string
	switch {
	case info.PackageMgr == "npm" || info.PackageMgr == "yarn" || info.PackageMgr == "pnpm" || info.PackageMgr == "bun":
		// Check for common E2E frameworks
		if fileExists(t.workspace + "/playwright.config.ts") || fileExists(t.workspace + "/playwright.config.js") {
			cmd = info.PackageMgr + " exec playwright test"
		} else if fileExists(t.workspace + "/cypress.config.ts") || fileExists(t.workspace + "/cypress.config.js") {
			cmd = info.PackageMgr + " exec cypress run"
		} else {
			cmd = info.PackageMgr + " test"
		}
	default:
		cmd = info.TestCmd
	}

	if cmd == "" {
		return "Could not detect E2E test framework. Supported: Playwright, Cypress, or project test command.", nil
	}

	if extraArgs != "" {
		cmd = cmd + " " + extraArgs
	}

	return projectTool.runShell(ctx, t.workspace, cmd, 5*60*1000000000) // 5 min
}

func (t *BrowserTool) playwrightAction(ctx context.Context, action string, args map[string]interface{}) (string, error) {
	if t.registry == nil {
		return t.noPlaywrightMsg(action), nil
	}

	// Map our actions to Playwright MCP tool names
	mcpToolNames := map[string][]string{
		"screenshot":       {"screenshot", "take_screenshot"},
		"click":            {"click", "mouse_click"},
		"fill":             {"fill", "type_text", "input_text"},
		"evaluate":         {"evaluate", "execute_javascript", "run_javascript"},
		"check_responsive": {"screenshot", "take_screenshot"}, // We'll call multiple times
		"pdf":              {"pdf", "save_pdf", "generate_pdf"},
	}

	candidates, ok := mcpToolNames[action]
	if !ok {
		return "", fmt.Errorf("unknown playwright action: %s", action)
	}

	tool := t.findMCPTool(candidates...)
	if tool == nil {
		return t.noPlaywrightMsg(action), nil
	}

	// Build MCP args based on action
	mcpArgs := make(map[string]interface{})
	switch action {
	case "screenshot":
		if url, ok := args["url"].(string); ok {
			mcpArgs["url"] = url
		}
	case "click":
		if sel, ok := args["selector"].(string); ok {
			mcpArgs["selector"] = sel
		}
	case "fill":
		if sel, ok := args["selector"].(string); ok {
			mcpArgs["selector"] = sel
		}
		if val, ok := args["value"].(string); ok {
			mcpArgs["value"] = val
		}
	case "evaluate":
		if val, ok := args["value"].(string); ok {
			mcpArgs["expression"] = val
		}
	case "pdf":
		if url, ok := args["url"].(string); ok {
			mcpArgs["url"] = url
		}
	}

	return tool.Execute(ctx, mcpArgs)
}

func (t *BrowserTool) findMCPTool(names ...string) Tool {
	if t.registry == nil {
		return nil
	}
	tools := t.registry.List()
	for _, tool := range tools {
		toolName := strings.ToLower(tool.Name())
		for _, name := range names {
			if strings.Contains(toolName, name) {
				return tool
			}
		}
	}
	return nil
}

func (t *BrowserTool) noPlaywrightMsg(action string) string {
	return fmt.Sprintf(`Action "%s" requires Playwright MCP server which is not configured.
Install it by adding to your MCP config (~/.MakoClaw/mcp/playwright/mcp.json):
{
  "command": "npx",
  "args": ["-y", "@anthropic/mcp-server-playwright"]
}
Then restart MakoClaw.`, action)
}

var (
	_ Tool          = (*BrowserTool)(nil)
	_ WorkspaceTool = (*BrowserTool)(nil)
)
