package agent

import (
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/providers"
	"github.com/sipeed/makoclaw/pkg/tools"
)

// ---------------------------------------------------------------------------
// NewContextBuilder
// ---------------------------------------------------------------------------

func TestNewContextBuilder(t *testing.T) {
	cb := NewContextBuilder("/tmp/test-workspace")
	if cb == nil {
		t.Fatal("expected non-nil ContextBuilder")
	}
	if cb.workspace != "/tmp/test-workspace" {
		t.Errorf("expected workspace '/tmp/test-workspace', got %q", cb.workspace)
	}
	if cb.userUUID != "" {
		t.Errorf("expected empty userUUID, got %q", cb.userUUID)
	}
	if cb.userID != 0 {
		t.Errorf("expected userID 0, got %d", cb.userID)
	}
	if cb.skillsLoader == nil {
		t.Error("expected non-nil skillsLoader")
	}
	if cb.memory == nil {
		t.Error("expected non-nil memory")
	}
}

// ---------------------------------------------------------------------------
// SetAgentSystemPrompt
// ---------------------------------------------------------------------------

func TestSetAgentSystemPrompt(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetAgentSystemPrompt("You are an orchestrator.")

	if cb.agentSystemPrompt != "You are an orchestrator." {
		t.Errorf("expected agent system prompt to be set, got %q", cb.agentSystemPrompt)
	}
}

func TestSetAgentSystemPrompt_AppearsInSystemPrompt(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetAgentSystemPrompt("CUSTOM_AGENT_PROMPT_MARKER")
	cb.SetLightweightMode(true) // skip bootstrap and memory for cleaner test

	prompt := cb.BuildSystemPrompt()
	if !strings.Contains(prompt, "CUSTOM_AGENT_PROMPT_MARKER") {
		t.Error("expected agent system prompt to appear in BuildSystemPrompt output")
	}
}

// ---------------------------------------------------------------------------
// SetSkillFilter
// ---------------------------------------------------------------------------

func TestSetSkillFilter_Nil(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetSkillFilter(nil)
	if cb.skillFilter != nil {
		t.Error("expected nil skillFilter")
	}
}

func TestSetSkillFilter_Empty(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetSkillFilter([]string{})
	if cb.skillFilter == nil {
		t.Fatal("expected non-nil skillFilter")
	}
	if len(cb.skillFilter) != 0 {
		t.Errorf("expected empty skillFilter, got %d items", len(cb.skillFilter))
	}
}

func TestSetSkillFilter_Specific(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetSkillFilter([]string{"github", "weather"})
	if len(cb.skillFilter) != 2 {
		t.Errorf("expected 2 skill filter entries, got %d", len(cb.skillFilter))
	}
}

// ---------------------------------------------------------------------------
// SetLightweightMode
// ---------------------------------------------------------------------------

func TestSetLightweightMode(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")

	if cb.lightweightMode {
		t.Error("expected lightweightMode to be false by default")
	}

	cb.SetLightweightMode(true)
	if !cb.lightweightMode {
		t.Error("expected lightweightMode to be true after setting")
	}

	cb.SetLightweightMode(false)
	if cb.lightweightMode {
		t.Error("expected lightweightMode to be false after unsetting")
	}
}

// ---------------------------------------------------------------------------
// SetToolsRegistry / buildToolsSection
// ---------------------------------------------------------------------------

func TestSetToolsRegistry(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	registry := tools.NewToolRegistry()
	cb.SetToolsRegistry(registry)

	if cb.tools != registry {
		t.Error("expected tools registry to be set")
	}
}

func TestBuildToolsSection_NilRegistry(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	result := cb.buildToolsSection()
	if result != "" {
		t.Errorf("expected empty string for nil registry, got %q", result)
	}
}

func TestBuildToolsSection_EmptyRegistry(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetToolsRegistry(tools.NewToolRegistry())
	result := cb.buildToolsSection()
	if result != "" {
		t.Errorf("expected empty string for empty registry, got %q", result)
	}
}

func TestBuildToolsSection_WithTools(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	registry := tools.NewToolRegistry()
	registry.Register(tools.NewReadFileTool("/tmp", false))
	cb.SetToolsRegistry(registry)

	result := cb.buildToolsSection()
	if !strings.Contains(result, "Available Tools") {
		t.Error("expected 'Available Tools' header in tools section")
	}
	if !strings.Contains(result, "CRITICAL") {
		t.Error("expected CRITICAL instruction in tools section")
	}
}

// ---------------------------------------------------------------------------
// BuildSystemPrompt
// ---------------------------------------------------------------------------

func TestBuildSystemPrompt_ContainsIdentity(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetLightweightMode(true)

	prompt := cb.BuildSystemPrompt()
	if !strings.Contains(prompt, "makoclaw") {
		t.Error("expected 'makoclaw' in system prompt identity section")
	}
}

func TestBuildSystemPrompt_LightweightSkipsBootstrap(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetLightweightMode(true)

	prompt := cb.BuildSystemPrompt()
	// In lightweight mode, bootstrap files (AGENTS.md, etc.) should not be loaded
	// Since the workspace is /tmp/workspace and these files don't exist, we verify
	// that the prompt doesn't include bootstrap file headers
	if strings.Contains(prompt, "## AGENTS.md") {
		t.Error("expected no bootstrap files in lightweight mode")
	}
}

func TestBuildSystemPrompt_SeparatedBySeparators(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetAgentSystemPrompt("AGENT_PROMPT_HERE")
	cb.SetLightweightMode(true) // skip bootstrap and memory for cleaner test

	prompt := cb.BuildSystemPrompt()
	if !strings.Contains(prompt, "---") {
		t.Error("expected separator in system prompt between sections")
	}
}

func TestBuildSystemPrompt_EmptySkillFilterMeansNoSkills(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetSkillFilter([]string{}) // empty = no skills
	cb.SetLightweightMode(true)

	prompt := cb.BuildSystemPrompt()
	// Should not contain "# Skills" header if no skills are loaded
	if strings.Contains(prompt, "# Skills") {
		t.Error("expected no skills section when skill filter is empty")
	}
}

// ---------------------------------------------------------------------------
// getIdentity
// ---------------------------------------------------------------------------

func TestGetIdentity_ContainsRequiredSections(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	identity := cb.getIdentity()

	sections := []string{
		"makoclaw",
		"Current Time",
		"Runtime",
		"Workspace",
		"Important Rules",
		"Security & Permissions",
	}

	for _, section := range sections {
		if !strings.Contains(identity, section) {
			t.Errorf("expected identity to contain %q", section)
		}
	}
}

func TestGetIdentity_ContainsWorkspacePath(t *testing.T) {
	cb := NewContextBuilder("/tmp/my-workspace")
	identity := cb.getIdentity()

	// The workspace path should appear (it gets resolved to absolute)
	if !strings.Contains(identity, "my-workspace") {
		t.Error("expected identity to contain workspace path")
	}
}

// ---------------------------------------------------------------------------
// BuildMessages
// ---------------------------------------------------------------------------

func TestBuildMessages_BasicStructure(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetLightweightMode(true)

	messages := cb.BuildMessages(nil, "", "Hello!", nil, "", "")

	if len(messages) < 2 {
		t.Fatalf("expected at least 2 messages (system + user), got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("expected first message role 'system', got %q", messages[0].Role)
	}

	lastMsg := messages[len(messages)-1]
	if lastMsg.Role != "user" {
		t.Errorf("expected last message role 'user', got %q", lastMsg.Role)
	}
	if lastMsg.Content != "Hello!" {
		t.Errorf("expected user message 'Hello!', got %q", lastMsg.Content)
	}
}

func TestBuildMessages_WithHistory(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetLightweightMode(true)

	history := []providers.Message{
		{Role: "user", Content: "previous question"},
		{Role: "assistant", Content: "previous answer"},
	}

	messages := cb.BuildMessages(history, "", "new question", nil, "", "")

	// Should be: system + 2 history + 1 current = 4
	if len(messages) != 4 {
		t.Errorf("expected 4 messages, got %d", len(messages))
	}
}

func TestBuildMessages_WithSummary(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetLightweightMode(true)

	messages := cb.BuildMessages(nil, "Previous conversation about X.", "New question", nil, "", "")

	// System prompt should contain the summary
	if !strings.Contains(messages[0].Content, "Summary of Previous Conversation") {
		t.Error("expected summary in system message")
	}
	if !strings.Contains(messages[0].Content, "Previous conversation about X.") {
		t.Error("expected summary content in system message")
	}
}

func TestBuildMessages_WithChannelInfo(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetLightweightMode(true)

	messages := cb.BuildMessages(nil, "", "Hi", nil, "telegram", "12345")

	if !strings.Contains(messages[0].Content, "Current Session") {
		t.Error("expected current session info in system prompt")
	}
	if !strings.Contains(messages[0].Content, "telegram") {
		t.Error("expected channel name in system prompt")
	}
	if !strings.Contains(messages[0].Content, "12345") {
		t.Error("expected chat ID in system prompt")
	}
}

func TestBuildMessages_StripsOrphanedToolMessages(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.SetLightweightMode(true)

	// History starts with orphaned tool message (no preceding assistant with tool_use)
	history := []providers.Message{
		{Role: "tool", Content: "orphaned tool result", ToolCallID: "call_123"},
		{Role: "user", Content: "previous question"},
		{Role: "assistant", Content: "previous answer"},
	}

	messages := cb.BuildMessages(history, "", "new question", nil, "", "")

	// The orphaned tool message should be stripped
	for i, msg := range messages {
		if i == 0 {
			continue // skip system message
		}
		if msg.Role == "tool" && msg.ToolCallID == "call_123" {
			t.Error("expected orphaned tool message to be stripped from history")
		}
	}
}

// ---------------------------------------------------------------------------
// AddToolResult
// ---------------------------------------------------------------------------

func TestAddToolResult(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	messages := []providers.Message{
		{Role: "system", Content: "system prompt"},
	}

	messages = cb.AddToolResult(messages, "call_1", "read_file", "file contents here")

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}

	toolMsg := messages[1]
	if toolMsg.Role != "tool" {
		t.Errorf("expected role 'tool', got %q", toolMsg.Role)
	}
	if toolMsg.Content != "file contents here" {
		t.Errorf("expected tool result content, got %q", toolMsg.Content)
	}
	if toolMsg.ToolCallID != "call_1" {
		t.Errorf("expected ToolCallID 'call_1', got %q", toolMsg.ToolCallID)
	}
}

// ---------------------------------------------------------------------------
// AddAssistantMessage
// ---------------------------------------------------------------------------

func TestAddAssistantMessage(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	messages := []providers.Message{
		{Role: "system", Content: "system prompt"},
	}

	messages = cb.AddAssistantMessage(messages, "Here is my response.", nil)

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}

	assistantMsg := messages[1]
	if assistantMsg.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", assistantMsg.Role)
	}
	if assistantMsg.Content != "Here is my response." {
		t.Errorf("expected content 'Here is my response.', got %q", assistantMsg.Content)
	}
}

func TestAddAssistantMessage_EmptyContent(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	messages := []providers.Message{}

	messages = cb.AddAssistantMessage(messages, "", nil)

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	if messages[0].Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", messages[0].Role)
	}
}

// ---------------------------------------------------------------------------
// WithUser
// ---------------------------------------------------------------------------

func TestWithUser_SetsFields(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.WithUser("abc-123-uuid", 42)

	if cb.userUUID != "abc-123-uuid" {
		t.Errorf("expected userUUID 'abc-123-uuid', got %q", cb.userUUID)
	}
	if cb.userID != 42 {
		t.Errorf("expected userID 42, got %d", cb.userID)
	}
	// Workspace should be updated to user-specific path
	if !strings.Contains(cb.workspace, "abc-123-uuid") {
		t.Errorf("expected workspace to contain user UUID, got %q", cb.workspace)
	}
}

func TestWithUser_EmptyUUID(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	original := cb.workspace
	cb.WithUser("", 0)

	// Empty UUID should not change workspace
	if cb.workspace != original {
		t.Errorf("expected workspace unchanged for empty UUID, got %q", cb.workspace)
	}
}

// ---------------------------------------------------------------------------
// WithAnalytics
// ---------------------------------------------------------------------------

func TestWithAnalytics(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	result := cb.WithAnalytics(nil, "session_key")

	if result != cb {
		t.Error("expected WithAnalytics to return same builder for chaining")
	}
	if cb.sessionKey != "session_key" {
		t.Errorf("expected sessionKey 'session_key', got %q", cb.sessionKey)
	}
}

// ---------------------------------------------------------------------------
// WithCentralStore
// ---------------------------------------------------------------------------

func TestWithCentralStore(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	result := cb.WithCentralStore(nil)

	if result != cb {
		t.Error("expected WithCentralStore to return same builder for chaining")
	}
}

// ---------------------------------------------------------------------------
// getUserWorkspacePath
// ---------------------------------------------------------------------------

func TestGetUserWorkspacePath_NoUser(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	path := cb.getUserWorkspacePath()
	if path != "/tmp/workspace" {
		t.Errorf("expected '/tmp/workspace', got %q", path)
	}
}

func TestGetUserWorkspacePath_WithUser(t *testing.T) {
	cb := NewContextBuilder("/tmp/workspace")
	cb.userUUID = "test-uuid"
	path := cb.getUserWorkspacePath()
	if !strings.Contains(path, "test-uuid") {
		t.Errorf("expected path to contain user UUID, got %q", path)
	}
	if !strings.Contains(path, "workspace") {
		t.Errorf("expected path to contain 'workspace', got %q", path)
	}
}
