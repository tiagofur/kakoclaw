package tools

import (
	"context"
	"strings"
	"testing"
)

// mockTool implements the Tool interface for testing.
type mockTool struct {
	name        string
	description string
	result      string
	executed    bool
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Description() string { return m.description }
func (m *mockTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}
func (m *mockTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	m.executed = true
	return m.result, nil
}

// mockContextualTool implements both Tool and ContextualTool for testing.
type mockContextualTool struct {
	mockTool
	channel string
	chatID  string
}

func (m *mockContextualTool) SetContext(channel, chatID string) {
	m.channel = channel
	m.chatID = chatID
}

func newMockTool(name, description, result string) *mockTool {
	return &mockTool{
		name:        name,
		description: description,
		result:      result,
	}
}

func TestToolRegistry_RegisterAndGet(t *testing.T) {
	registry := NewToolRegistry()
	tool := newMockTool("test_tool", "A test tool", "result")

	registry.Register(tool)

	got, ok := registry.Get("test_tool")
	if !ok {
		t.Fatal("expected tool to be found after registration")
	}
	if got.Name() != "test_tool" {
		t.Errorf("expected name %q, got %q", "test_tool", got.Name())
	}
	if got.Description() != "A test tool" {
		t.Errorf("expected description %q, got %q", "A test tool", got.Description())
	}
}

func TestToolRegistry_GetMissing(t *testing.T) {
	registry := NewToolRegistry()

	_, ok := registry.Get("nonexistent")
	if ok {
		t.Error("expected tool not to be found")
	}
}

func TestToolRegistry_List(t *testing.T) {
	registry := NewToolRegistry()
	names := []string{"alpha", "beta", "gamma"}
	for _, name := range names {
		registry.Register(newMockTool(name, "desc", "result"))
	}

	list := registry.List()
	if len(list) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(list))
	}

	// All names should be present (order is not guaranteed due to map iteration)
	nameSet := make(map[string]bool)
	for _, n := range list {
		nameSet[n] = true
	}
	for _, name := range names {
		if !nameSet[name] {
			t.Errorf("expected %q in list", name)
		}
	}
}

func TestToolRegistry_Count(t *testing.T) {
	registry := NewToolRegistry()

	if registry.Count() != 0 {
		t.Errorf("expected count 0 for empty registry, got %d", registry.Count())
	}

	registry.Register(newMockTool("one", "desc", "result"))
	registry.Register(newMockTool("two", "desc", "result"))
	registry.Register(newMockTool("three", "desc", "result"))

	if registry.Count() != 3 {
		t.Errorf("expected count 3, got %d", registry.Count())
	}
}

func TestToolRegistry_Execute(t *testing.T) {
	registry := NewToolRegistry()
	tool := newMockTool("exec_test", "desc", "execution result")
	registry.Register(tool)

	result, err := registry.Execute(context.Background(), "exec_test", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "execution result" {
		t.Errorf("expected %q, got %q", "execution result", result)
	}
	if !tool.executed {
		t.Error("expected tool.executed to be true")
	}
}

func TestToolRegistry_ExecuteNotFound(t *testing.T) {
	registry := NewToolRegistry()

	_, err := registry.Execute(context.Background(), "missing_tool", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for missing tool, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestToolRegistry_GetDefinitions(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register(newMockTool("def_tool", "A definition tool", "result"))

	defs := registry.GetDefinitions()
	if len(defs) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(defs))
	}

	def := defs[0]
	if def["type"] != "function" {
		t.Errorf("expected type 'function', got %v", def["type"])
	}

	fn, ok := def["function"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'function' key to be a map")
	}
	if fn["name"] != "def_tool" {
		t.Errorf("expected name 'def_tool', got %v", fn["name"])
	}
	if fn["description"] != "A definition tool" {
		t.Errorf("expected description 'A definition tool', got %v", fn["description"])
	}
	if fn["parameters"] == nil {
		t.Error("expected non-nil parameters")
	}
}

func TestToolRegistry_GetSummaries(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register(newMockTool("summary_tool", "Does things", "result"))
	registry.Register(newMockTool("another_tool", "Does other things", "result"))

	summaries := registry.GetSummaries()
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}

	// Check that each summary follows the expected format
	foundSummary := false
	foundAnother := false
	for _, s := range summaries {
		if strings.Contains(s, "summary_tool") && strings.Contains(s, "Does things") {
			foundSummary = true
		}
		if strings.Contains(s, "another_tool") && strings.Contains(s, "Does other things") {
			foundAnother = true
		}
		// Verify format: should contain backtick-wrapped name and description
		if !strings.Contains(s, "`") {
			t.Errorf("expected backtick-wrapped name in summary, got: %s", s)
		}
		if !strings.HasPrefix(s, "- ") {
			t.Errorf("expected summary to start with '- ', got: %s", s)
		}
	}
	if !foundSummary {
		t.Error("expected to find summary_tool in summaries")
	}
	if !foundAnother {
		t.Error("expected to find another_tool in summaries")
	}
}

func TestToolRegistry_ForEach(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register(newMockTool("fe_one", "desc", "result"))
	registry.Register(newMockTool("fe_two", "desc", "result"))
	registry.Register(newMockTool("fe_three", "desc", "result"))

	collected := make(map[string]bool)
	registry.ForEach(func(tool Tool) {
		collected[tool.Name()] = true
	})

	if len(collected) != 3 {
		t.Fatalf("expected 3 tools collected, got %d", len(collected))
	}
	for _, name := range []string{"fe_one", "fe_two", "fe_three"} {
		if !collected[name] {
			t.Errorf("expected %q to be collected by ForEach", name)
		}
	}
}

func TestToolRegistry_ExecuteWithContext_SetsContext(t *testing.T) {
	registry := NewToolRegistry()
	tool := &mockContextualTool{
		mockTool: mockTool{
			name:        "ctx_tool",
			description: "A contextual tool",
			result:      "ctx result",
		},
	}
	registry.Register(tool)

	result, err := registry.ExecuteWithContext(
		context.Background(),
		"ctx_tool",
		map[string]interface{}{},
		"telegram",
		"chat123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ctx result" {
		t.Errorf("expected %q, got %q", "ctx result", result)
	}
	if !tool.executed {
		t.Error("expected tool to be executed")
	}
	if tool.channel != "telegram" {
		t.Errorf("expected channel %q, got %q", "telegram", tool.channel)
	}
	if tool.chatID != "chat123" {
		t.Errorf("expected chatID %q, got %q", "chat123", tool.chatID)
	}
}

func TestToolRegistry_ExecuteWithContext_NoContextForNonContextualTool(t *testing.T) {
	registry := NewToolRegistry()
	tool := newMockTool("plain_tool", "A plain tool", "plain result")
	registry.Register(tool)

	// ExecuteWithContext should work fine even for non-contextual tools
	result, err := registry.ExecuteWithContext(
		context.Background(),
		"plain_tool",
		map[string]interface{}{},
		"discord",
		"chat456",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "plain result" {
		t.Errorf("expected %q, got %q", "plain result", result)
	}
	if !tool.executed {
		t.Error("expected tool to be executed")
	}
}

func TestToolRegistry_RegisterOverwrite(t *testing.T) {
	registry := NewToolRegistry()
	tool1 := newMockTool("overwrite_me", "first version", "result1")
	tool2 := newMockTool("overwrite_me", "second version", "result2")

	registry.Register(tool1)
	registry.Register(tool2)

	if registry.Count() != 1 {
		t.Errorf("expected count 1 after overwrite, got %d", registry.Count())
	}

	got, ok := registry.Get("overwrite_me")
	if !ok {
		t.Fatal("expected tool to be found")
	}
	if got.Description() != "second version" {
		t.Errorf("expected description %q (second registration), got %q", "second version", got.Description())
	}
}

func TestToolToSchema(t *testing.T) {
	tool := newMockTool("schema_test", "Schema description", "result")
	schema := ToolToSchema(tool)

	if schema["type"] != "function" {
		t.Errorf("expected type 'function', got %v", schema["type"])
	}

	fn, ok := schema["function"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'function' key to be a map")
	}
	if fn["name"] != "schema_test" {
		t.Errorf("expected name %q, got %v", "schema_test", fn["name"])
	}
	if fn["description"] != "Schema description" {
		t.Errorf("expected description %q, got %v", "Schema description", fn["description"])
	}
	if fn["parameters"] == nil {
		t.Error("expected non-nil parameters in schema")
	}
}
