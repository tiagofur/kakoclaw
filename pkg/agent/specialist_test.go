package agent

import (
	"testing"
)

// ---------------------------------------------------------------------------
// NewSpecialistRegistry
// ---------------------------------------------------------------------------

func TestNewSpecialistRegistry(t *testing.T) {
	sr := NewSpecialistRegistry()
	if sr == nil {
		t.Fatal("expected non-nil SpecialistRegistry")
	}
	if len(sr.specialists) != 0 {
		t.Errorf("expected empty specialists map, got %d entries", len(sr.specialists))
	}
}

// ---------------------------------------------------------------------------
// RegisterSpecialist
// ---------------------------------------------------------------------------

func TestRegisterSpecialist_Success(t *testing.T) {
	sr := NewSpecialistRegistry()
	sa := &SpecialistAgent{
		name:         "developer",
		description:  "Writes code",
		allowedTools: map[string]bool{"read_file": true},
		keywords:     []string{"code", "develop"},
	}

	err := sr.RegisterSpecialist(sa)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := sr.GetSpecialist("developer")
	if err != nil {
		t.Fatalf("expected to find specialist: %v", err)
	}
	if got.name != "developer" {
		t.Errorf("expected name 'developer', got %q", got.name)
	}
}

func TestRegisterSpecialist_Nil(t *testing.T) {
	sr := NewSpecialistRegistry()
	err := sr.RegisterSpecialist(nil)
	if err == nil {
		t.Fatal("expected error for nil specialist")
	}
}

func TestRegisterSpecialist_EmptyName(t *testing.T) {
	sr := NewSpecialistRegistry()
	sa := &SpecialistAgent{
		name: "",
	}
	err := sr.RegisterSpecialist(sa)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegisterSpecialist_Overwrite(t *testing.T) {
	sr := NewSpecialistRegistry()

	sa1 := &SpecialistAgent{
		name:        "developer",
		description: "version 1",
	}
	sa2 := &SpecialistAgent{
		name:        "developer",
		description: "version 2",
	}

	if err := sr.RegisterSpecialist(sa1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := sr.RegisterSpecialist(sa2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := sr.GetSpecialist("developer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.description != "version 2" {
		t.Errorf("expected description 'version 2', got %q", got.description)
	}
}

// ---------------------------------------------------------------------------
// GetSpecialist
// ---------------------------------------------------------------------------

func TestGetSpecialist_Found(t *testing.T) {
	sr := NewSpecialistRegistry()
	sa := &SpecialistAgent{name: "tester", description: "Runs tests"}
	sr.RegisterSpecialist(sa)

	got, err := sr.GetSpecialist("tester")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.name != "tester" {
		t.Errorf("expected 'tester', got %q", got.name)
	}
}

func TestGetSpecialist_NotFound(t *testing.T) {
	sr := NewSpecialistRegistry()

	_, err := sr.GetSpecialist("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent specialist")
	}
}

// ---------------------------------------------------------------------------
// ListSpecialists
// ---------------------------------------------------------------------------

func TestListSpecialists_Empty(t *testing.T) {
	sr := NewSpecialistRegistry()
	result := sr.ListSpecialists()
	if len(result) != 0 {
		t.Errorf("expected 0 specialists, got %d", len(result))
	}
}

func TestListSpecialists_Multiple(t *testing.T) {
	sr := NewSpecialistRegistry()
	sr.RegisterSpecialist(&SpecialistAgent{name: "a"})
	sr.RegisterSpecialist(&SpecialistAgent{name: "b"})
	sr.RegisterSpecialist(&SpecialistAgent{name: "c"})

	result := sr.ListSpecialists()
	if len(result) != 3 {
		t.Errorf("expected 3 specialists, got %d", len(result))
	}
}

func TestListSpecialists_ReturnsCopy(t *testing.T) {
	sr := NewSpecialistRegistry()
	sr.RegisterSpecialist(&SpecialistAgent{name: "dev"})

	result := sr.ListSpecialists()
	// Modify the returned map
	result["injected"] = &SpecialistAgent{name: "injected"}

	// Original should be unchanged
	original := sr.ListSpecialists()
	if _, ok := original["injected"]; ok {
		t.Error("ListSpecialists should return a copy; modifying result affected the registry")
	}
	if len(original) != 1 {
		t.Errorf("expected 1 specialist in registry, got %d", len(original))
	}
}

// ---------------------------------------------------------------------------
// RemoveSpecialist
// ---------------------------------------------------------------------------

func TestRemoveSpecialist_Success(t *testing.T) {
	sr := NewSpecialistRegistry()
	sr.RegisterSpecialist(&SpecialistAgent{name: "dev"})

	removed := sr.RemoveSpecialist("dev")
	if !removed {
		t.Error("expected RemoveSpecialist to return true")
	}

	_, err := sr.GetSpecialist("dev")
	if err == nil {
		t.Error("expected error after removal")
	}
}

func TestRemoveSpecialist_NotFound(t *testing.T) {
	sr := NewSpecialistRegistry()
	removed := sr.RemoveSpecialist("nonexistent")
	if removed {
		t.Error("expected RemoveSpecialist to return false for nonexistent")
	}
}

func TestRemoveSpecialist_EmptyName(t *testing.T) {
	sr := NewSpecialistRegistry()
	removed := sr.RemoveSpecialist("")
	if removed {
		t.Error("expected RemoveSpecialist to return false for empty name")
	}
}

func TestRemoveSpecialist_DoesNotAffectOthers(t *testing.T) {
	sr := NewSpecialistRegistry()
	sr.RegisterSpecialist(&SpecialistAgent{name: "a"})
	sr.RegisterSpecialist(&SpecialistAgent{name: "b"})

	sr.RemoveSpecialist("a")

	if _, err := sr.GetSpecialist("b"); err != nil {
		t.Error("removing 'a' should not affect 'b'")
	}
	if len(sr.ListSpecialists()) != 1 {
		t.Errorf("expected 1 specialist after removal, got %d", len(sr.ListSpecialists()))
	}
}

// ---------------------------------------------------------------------------
// SpecialistAgent getters
// ---------------------------------------------------------------------------

func TestSpecialistAgent_GetName(t *testing.T) {
	sa := &SpecialistAgent{name: "developer"}
	if sa.GetName() != "developer" {
		t.Errorf("expected 'developer', got %q", sa.GetName())
	}
}

func TestSpecialistAgent_GetDescription(t *testing.T) {
	sa := &SpecialistAgent{description: "Writes code and tests"}
	if sa.GetDescription() != "Writes code and tests" {
		t.Errorf("expected 'Writes code and tests', got %q", sa.GetDescription())
	}
}

func TestSpecialistAgent_GetSpecialistKeywords(t *testing.T) {
	sa := &SpecialistAgent{keywords: []string{"code", "implement", "fix"}}
	kw := sa.GetSpecialistKeywords()
	if len(kw) != 3 {
		t.Errorf("expected 3 keywords, got %d", len(kw))
	}
	if kw[0] != "code" || kw[1] != "implement" || kw[2] != "fix" {
		t.Errorf("unexpected keywords: %v", kw)
	}
}

func TestSpecialistAgent_GetSpecialistKeywords_Empty(t *testing.T) {
	sa := &SpecialistAgent{}
	kw := sa.GetSpecialistKeywords()
	if kw != nil {
		t.Errorf("expected nil keywords, got %v", kw)
	}
}

func TestSpecialistAgent_GetAllowedTools(t *testing.T) {
	sa := &SpecialistAgent{
		allowedTools: map[string]bool{
			"read_file":  true,
			"write_file": true,
			"exec":       true,
		},
	}
	tools := sa.GetAllowedTools()
	if len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
	}

	// Check all expected tools are present
	toolSet := make(map[string]bool)
	for _, tool := range tools {
		toolSet[tool] = true
	}
	for _, expected := range []string{"read_file", "write_file", "exec"} {
		if !toolSet[expected] {
			t.Errorf("expected tool %q in allowed tools", expected)
		}
	}
}

func TestSpecialistAgent_GetAllowedTools_Empty(t *testing.T) {
	sa := &SpecialistAgent{allowedTools: map[string]bool{}}
	tools := sa.GetAllowedTools()
	if len(tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(tools))
	}
}

// ---------------------------------------------------------------------------
// GetSpecialistInfo
// ---------------------------------------------------------------------------

func TestGetSpecialistInfo_Success(t *testing.T) {
	sr := NewSpecialistRegistry()
	sa := &SpecialistAgent{
		name:         "developer",
		description:  "Code specialist",
		prompt:       "You are a code expert.",
		providerName: "openai",
		allowedTools: map[string]bool{"read_file": true, "write_file": true},
		keywords:     []string{"code", "develop"},
	}
	// Set model through the embedded AgentLoop (we use a zero-value AgentLoop)
	sa.AgentLoop = &AgentLoop{model: "gpt-4"}
	sr.RegisterSpecialist(sa)

	info, err := sr.GetSpecialistInfo("developer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Name != "developer" {
		t.Errorf("expected name 'developer', got %q", info.Name)
	}
	if info.Description != "Code specialist" {
		t.Errorf("expected description 'Code specialist', got %q", info.Description)
	}
	if info.Prompt != "You are a code expert." {
		t.Errorf("expected prompt, got %q", info.Prompt)
	}
	if info.Provider != "openai" {
		t.Errorf("expected provider 'openai', got %q", info.Provider)
	}
	if info.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got %q", info.Model)
	}
	if len(info.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(info.Tools))
	}
	if len(info.Keywords) != 2 {
		t.Errorf("expected 2 keywords, got %d", len(info.Keywords))
	}
}

func TestGetSpecialistInfo_NotFound(t *testing.T) {
	sr := NewSpecialistRegistry()
	_, err := sr.GetSpecialistInfo("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent specialist")
	}
}

// ---------------------------------------------------------------------------
// NewSpecialistSelector
// ---------------------------------------------------------------------------

func TestNewSpecialistSelector(t *testing.T) {
	sr := NewSpecialistRegistry()
	ss := NewSpecialistSelector(sr)
	if ss == nil {
		t.Fatal("expected non-nil SpecialistSelector")
	}
	if ss.registry != sr {
		t.Error("expected selector to reference the provided registry")
	}
}

// ---------------------------------------------------------------------------
// SpecialistSelector.SelectByName
// ---------------------------------------------------------------------------

func TestSelectByName_Found(t *testing.T) {
	sr := NewSpecialistRegistry()
	sr.RegisterSpecialist(&SpecialistAgent{name: "dev"})
	ss := NewSpecialistSelector(sr)

	got, err := ss.SelectByName("dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.name != "dev" {
		t.Errorf("expected 'dev', got %q", got.name)
	}
}

func TestSelectByName_NotFound(t *testing.T) {
	sr := NewSpecialistRegistry()
	ss := NewSpecialistSelector(sr)

	_, err := ss.SelectByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent specialist")
	}
}

// ---------------------------------------------------------------------------
// SpecialistSelector.SelectAll
// ---------------------------------------------------------------------------

func TestSelectAll_Empty(t *testing.T) {
	sr := NewSpecialistRegistry()
	ss := NewSpecialistSelector(sr)

	result := ss.SelectAll()
	if len(result) != 0 {
		t.Errorf("expected 0 specialists, got %d", len(result))
	}
}

func TestSelectAll_Multiple(t *testing.T) {
	sr := NewSpecialistRegistry()
	sr.RegisterSpecialist(&SpecialistAgent{name: "a"})
	sr.RegisterSpecialist(&SpecialistAgent{name: "b"})
	ss := NewSpecialistSelector(sr)

	result := ss.SelectAll()
	if len(result) != 2 {
		t.Errorf("expected 2 specialists, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// SpecialistSelector.SelectByKeywords
// ---------------------------------------------------------------------------

func TestSelectByKeywords_Match(t *testing.T) {
	sr := NewSpecialistRegistry()
	sr.RegisterSpecialist(&SpecialistAgent{name: "dev", keywords: []string{"code", "implement"}})
	sr.RegisterSpecialist(&SpecialistAgent{name: "tester", keywords: []string{"test", "qa"}})
	sr.RegisterSpecialist(&SpecialistAgent{name: "docs", keywords: []string{"docs", "documentation"}})
	ss := NewSpecialistSelector(sr)

	matches := ss.SelectByKeywords([]string{"code"})
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].name != "dev" {
		t.Errorf("expected 'dev', got %q", matches[0].name)
	}
}

func TestSelectByKeywords_NoMatch(t *testing.T) {
	sr := NewSpecialistRegistry()
	sr.RegisterSpecialist(&SpecialistAgent{name: "dev", keywords: []string{"code"}})
	ss := NewSpecialistSelector(sr)

	matches := ss.SelectByKeywords([]string{"security"})
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestSelectByKeywords_MultipleMatches(t *testing.T) {
	sr := NewSpecialistRegistry()
	sr.RegisterSpecialist(&SpecialistAgent{name: "dev", keywords: []string{"code", "review"}})
	sr.RegisterSpecialist(&SpecialistAgent{name: "reviewer", keywords: []string{"review", "audit"}})
	ss := NewSpecialistSelector(sr)

	matches := ss.SelectByKeywords([]string{"review"})
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

func TestSelectByKeywords_EmptyKeywords(t *testing.T) {
	sr := NewSpecialistRegistry()
	sr.RegisterSpecialist(&SpecialistAgent{name: "dev", keywords: []string{"code"}})
	ss := NewSpecialistSelector(sr)

	matches := ss.SelectByKeywords([]string{})
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for empty keywords, got %d", len(matches))
	}
}

func TestSelectByKeywords_NilKeywords(t *testing.T) {
	sr := NewSpecialistRegistry()
	sr.RegisterSpecialist(&SpecialistAgent{name: "dev", keywords: []string{"code"}})
	ss := NewSpecialistSelector(sr)

	matches := ss.SelectByKeywords(nil)
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for nil keywords, got %d", len(matches))
	}
}

// ---------------------------------------------------------------------------
// contains helper
// ---------------------------------------------------------------------------

func TestContains_Found(t *testing.T) {
	if !contains([]string{"a", "b", "c"}, "b") {
		t.Error("expected true for item in slice")
	}
}

func TestContains_NotFound(t *testing.T) {
	if contains([]string{"a", "b", "c"}, "d") {
		t.Error("expected false for item not in slice")
	}
}

func TestContains_EmptySlice(t *testing.T) {
	if contains([]string{}, "a") {
		t.Error("expected false for empty slice")
	}
}

func TestContains_NilSlice(t *testing.T) {
	if contains(nil, "a") {
		t.Error("expected false for nil slice")
	}
}

func TestContains_EmptyItem(t *testing.T) {
	if !contains([]string{"", "a"}, "") {
		t.Error("expected true for empty string in slice containing empty string")
	}
}

// ---------------------------------------------------------------------------
// truncateString helper
// ---------------------------------------------------------------------------

func TestTruncateString_Short(t *testing.T) {
	result := truncateString("hello", 10)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncateString_ExactLength(t *testing.T) {
	result := truncateString("hello", 5)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncateString_Long(t *testing.T) {
	result := truncateString("hello world", 5)
	if result != "hello..." {
		t.Errorf("expected 'hello...', got %q", result)
	}
}

func TestTruncateString_Empty(t *testing.T) {
	result := truncateString("", 10)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestTruncateString_ZeroMax(t *testing.T) {
	result := truncateString("hello", 0)
	if result != "..." {
		t.Errorf("expected '...', got %q", result)
	}
}

// ---------------------------------------------------------------------------
// RequestColleagueTool metadata
// ---------------------------------------------------------------------------

func TestRequestColleagueTool_Name(t *testing.T) {
	rct := &RequestColleagueTool{}
	if rct.Name() != "request_colleague" {
		t.Errorf("expected 'request_colleague', got %q", rct.Name())
	}
}

func TestRequestColleagueTool_Description(t *testing.T) {
	rct := &RequestColleagueTool{}
	if rct.Description() == "" {
		t.Error("expected non-empty description")
	}
}

func TestRequestColleagueTool_Parameters(t *testing.T) {
	rct := &RequestColleagueTool{}
	params := rct.Parameters()
	if params == nil {
		t.Fatal("expected non-nil parameters")
	}
	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected properties map")
	}
	if _, ok := props["colleague"]; !ok {
		t.Error("expected colleague property")
	}
	if _, ok := props["question"]; !ok {
		t.Error("expected question property")
	}
}

// ---------------------------------------------------------------------------
// NewRequestColleagueTool
// ---------------------------------------------------------------------------

func TestNewRequestColleagueTool(t *testing.T) {
	sr := NewSpecialistRegistry()
	sa := &SpecialistAgent{name: "dev"}
	tool := NewRequestColleagueTool(sr, sa)

	if tool == nil {
		t.Fatal("expected non-nil RequestColleagueTool")
	}
	if tool.registry != sr {
		t.Error("expected tool to reference provided registry")
	}
	if tool.currentAgent != sa {
		t.Error("expected tool to reference provided agent")
	}
	if tool.maxNestingDepth != 2 {
		t.Errorf("expected maxNestingDepth 2, got %d", tool.maxNestingDepth)
	}
	if tool.currentDepth != 0 {
		t.Errorf("expected currentDepth 0, got %d", tool.currentDepth)
	}
}

// ---------------------------------------------------------------------------
// SpecialistInfo struct
// ---------------------------------------------------------------------------

func TestSpecialistInfo_Fields(t *testing.T) {
	info := SpecialistInfo{
		Name:        "researcher",
		Description: "Research specialist",
		Prompt:      "You research things.",
		Provider:    "anthropic",
		Model:       "claude-3",
		Tools:       []string{"web_search", "web_fetch"},
		Keywords:    []string{"research", "investigate"},
	}

	if info.Name != "researcher" {
		t.Errorf("expected name 'researcher', got %q", info.Name)
	}
	if len(info.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(info.Tools))
	}
	if len(info.Keywords) != 2 {
		t.Errorf("expected 2 keywords, got %d", len(info.Keywords))
	}
}
