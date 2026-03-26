package agent

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/providers"
)

type sequenceProvider struct {
	responses []*providers.LLMResponse
	idx       int
}

func (p *sequenceProvider) Chat(ctx context.Context, messages []providers.Message, tools []providers.ToolDefinition, model string, options map[string]interface{}) (*providers.LLMResponse, error) {
	if p.idx >= len(p.responses) {
		return &providers.LLMResponse{Content: "done", FinishReason: "stop"}, nil
	}
	resp := p.responses[p.idx]
	p.idx++
	return resp, nil
}

func (p *sequenceProvider) GetDefaultModel() string {
	return "sequence-test-model"
}

type sessionRecordingTool struct {
	lastSessionKey string
}

func (t *sessionRecordingTool) Name() string { return "session_recorder" }

func (t *sessionRecordingTool) Description() string { return "records propagated session key" }

func (t *sessionRecordingTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *sessionRecordingTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	t.lastSessionKey = sessionKeyFromCtx(ctx)
	return "recorded", nil
}

func newTestSpecialistAgent(t *testing.T, name string, provider providers.LLMProvider) *SpecialistAgent {
	t.Helper()

	cfg := newAgentTestConfig(t.TempDir())
	loop := NewAgentLoop(cfg, bus.NewMessageBus(), provider)
	return &SpecialistAgent{
		AgentLoop:    loop,
		name:         name,
		description:  name + " specialist",
		allowedTools: map[string]bool{},
	}
}

// ---------------------------------------------------------------------------
// truncate
// ---------------------------------------------------------------------------

func TestTruncate_ShortString(t *testing.T) {
	result := truncate("hello", 10)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncate_ExactLength(t *testing.T) {
	result := truncate("hello", 5)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncate_LongString(t *testing.T) {
	result := truncate("hello world", 5)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncate_EmptyString(t *testing.T) {
	result := truncate("", 10)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestTruncate_ZeroMax(t *testing.T) {
	result := truncate("hello", 0)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestTruncate_SingleChar(t *testing.T) {
	result := truncate("abcdef", 1)
	if result != "a" {
		t.Errorf("expected 'a', got %q", result)
	}
}

// ---------------------------------------------------------------------------
// generateDelegationReason
// ---------------------------------------------------------------------------

func TestGenerateDelegationReason_KnownSpecialists(t *testing.T) {
	tests := []struct {
		name       string
		specialist string
		task       string
		wantPrefix string
	}{
		{"developer", "developer", "write code", "This task requires code implementation"},
		{"documentation", "documentation", "update docs", "This task requires documentation updates"},
		{"testing", "testing", "write unit tests", "This task requires test creation"},
		{"devops", "devops", "deploy to prod", "This task requires infrastructure work"},
		{"analyst", "analyst", "check data", "This task requires data analysis"},
		{"researcher", "researcher", "investigate bug", "This task requires investigation"},
		{"general", "general", "do something", "This task general-purpose"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateDelegationReason(tt.specialist, tt.task)
			if !strings.HasPrefix(result, "This task") {
				t.Errorf("expected reason to start with 'This task', got %q", result)
			}
			if result == "" {
				t.Errorf("expected non-empty reason for specialist %q", tt.specialist)
			}
		})
	}
}

func TestGenerateDelegationReason_UnknownSpecialist(t *testing.T) {
	result := generateDelegationReason("my_custom_agent", "do work")
	expected := "Delegating to my_custom_agent specialist"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGenerateDelegationReason_EmptySpecialist(t *testing.T) {
	result := generateDelegationReason("", "do work")
	expected := "Delegating to  specialist"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// ---------------------------------------------------------------------------
// parseSpecialistReport (method on OrchestratorAgent)
// ---------------------------------------------------------------------------

// newTestOrchestratorAgent creates a minimal OrchestratorAgent for testing
// methods that do not need a fully wired agent.
func newTestOrchestratorAgent() *OrchestratorAgent {
	return &OrchestratorAgent{}
}

func TestParseSpecialistReport_PlainText(t *testing.T) {
	oa := newTestOrchestratorAgent()
	report := oa.parseSpecialistReport("dev", "Here is my plain text response.")

	if report.SpecialistName != "dev" {
		t.Errorf("expected specialist name 'dev', got %q", report.SpecialistName)
	}
	if report.Status != "complete" {
		t.Errorf("expected default status 'complete', got %q", report.Status)
	}
	if report.Confidence != 0.9 {
		t.Errorf("expected default confidence 0.9, got %f", report.Confidence)
	}
	if report.Result != "Here is my plain text response." {
		t.Errorf("expected result to match input, got %q", report.Result)
	}
	if report.NeedsReview {
		t.Error("expected NeedsReview to be false for plain text")
	}
	if len(report.Suggestions) != 0 {
		t.Errorf("expected no suggestions, got %d", len(report.Suggestions))
	}
}

func TestParseSpecialistReport_ValidJSONHeader(t *testing.T) {
	oa := newTestOrchestratorAgent()
	input := `{"status":"partial","confidence":0.7,"request_help":"","suggestion":"needs more work"}
Here is the actual content of the response.`

	report := oa.parseSpecialistReport("tester", input)

	if report.SpecialistName != "tester" {
		t.Errorf("expected specialist name 'tester', got %q", report.SpecialistName)
	}
	if report.Status != "partial" {
		t.Errorf("expected status 'partial', got %q", report.Status)
	}
	if report.Confidence != 0.7 {
		t.Errorf("expected confidence 0.7, got %f", report.Confidence)
	}
	if report.NeedsReview {
		t.Error("expected NeedsReview to be false when request_help is empty")
	}
	if len(report.Suggestions) != 1 || report.Suggestions[0] != "needs more work" {
		t.Errorf("expected suggestion 'needs more work', got %v", report.Suggestions)
	}
}

func TestParseSpecialistReport_RequestHelp(t *testing.T) {
	oa := newTestOrchestratorAgent()
	input := `{"status":"needs_help","confidence":0.3,"request_help":"security","suggestion":"need security review"}
I found some issues but need a security expert.`

	report := oa.parseSpecialistReport("developer", input)

	if report.Status != "needs_help" {
		t.Errorf("expected status 'needs_help', got %q", report.Status)
	}
	if report.Confidence != 0.3 {
		t.Errorf("expected confidence 0.3, got %f", report.Confidence)
	}
	if report.RequestHelp != "security" {
		t.Errorf("expected RequestHelp 'security', got %q", report.RequestHelp)
	}
	if !report.NeedsReview {
		t.Error("expected NeedsReview to be true when request_help is set")
	}
}

func TestParseSpecialistReport_InvalidJSON(t *testing.T) {
	oa := newTestOrchestratorAgent()
	input := `{not valid json at all}
Some content follows.`

	report := oa.parseSpecialistReport("dev", input)

	// Should fall back to defaults
	if report.Status != "complete" {
		t.Errorf("expected default status 'complete', got %q", report.Status)
	}
	if report.Confidence != 0.9 {
		t.Errorf("expected default confidence 0.9, got %f", report.Confidence)
	}
}

func TestParseSpecialistReport_EmptyResult(t *testing.T) {
	oa := newTestOrchestratorAgent()
	report := oa.parseSpecialistReport("dev", "")

	if report.SpecialistName != "dev" {
		t.Errorf("expected specialist name 'dev', got %q", report.SpecialistName)
	}
	if report.Status != "complete" {
		t.Errorf("expected default status 'complete', got %q", report.Status)
	}
	if report.Result != "" {
		t.Errorf("expected empty result, got %q", report.Result)
	}
}

func TestParseSpecialistReport_JSONOnly_NoContent(t *testing.T) {
	oa := newTestOrchestratorAgent()
	input := `{"status":"complete","confidence":1.0,"request_help":"","suggestion":""}`

	report := oa.parseSpecialistReport("dev", input)

	if report.Status != "complete" {
		t.Errorf("expected status 'complete', got %q", report.Status)
	}
	if report.Confidence != 1.0 {
		t.Errorf("expected confidence 1.0, got %f", report.Confidence)
	}
}

func TestParseSpecialistReport_ZeroConfidence_KeepsDefault(t *testing.T) {
	// When confidence is 0 in JSON, the parser keeps the default 0.9
	// because the check is `if parsed.Confidence > 0`
	oa := newTestOrchestratorAgent()
	input := `{"status":"failed","confidence":0,"request_help":"","suggestion":""}
Failed task.`

	report := oa.parseSpecialistReport("dev", input)

	if report.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", report.Status)
	}
	// confidence=0 does not pass the > 0 check, so default 0.9 is kept
	if report.Confidence != 0.9 {
		t.Errorf("expected default confidence 0.9 (zero not applied), got %f", report.Confidence)
	}
}

func TestParseSpecialistReport_TimestampSet(t *testing.T) {
	oa := newTestOrchestratorAgent()
	before := time.Now()
	report := oa.parseSpecialistReport("dev", "test")
	after := time.Now()

	if report.Timestamp.Before(before) || report.Timestamp.After(after) {
		t.Error("expected timestamp to be between before and after test execution")
	}
}

// ---------------------------------------------------------------------------
// cleanSpecialistResult
// ---------------------------------------------------------------------------

func TestCleanSpecialistResult_WithJSONHeader(t *testing.T) {
	oa := newTestOrchestratorAgent()
	input := `{"status":"complete","confidence":0.9}
This is the actual content.`

	result := oa.cleanSpecialistResult(input)
	if result != "This is the actual content." {
		t.Errorf("expected cleaned result, got %q", result)
	}
}

func TestCleanSpecialistResult_WithoutJSONHeader(t *testing.T) {
	oa := newTestOrchestratorAgent()
	input := "This is just plain text without JSON."

	result := oa.cleanSpecialistResult(input)
	if result != input {
		t.Errorf("expected original text, got %q", result)
	}
}

func TestCleanSpecialistResult_EmptyString(t *testing.T) {
	oa := newTestOrchestratorAgent()
	result := oa.cleanSpecialistResult("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestCleanSpecialistResult_JSONHeaderOnlyLine(t *testing.T) {
	oa := newTestOrchestratorAgent()
	// JSON on first line, but nothing after newline (trailing whitespace only)
	input := "{\"status\":\"complete\"}\n   "

	result := oa.cleanSpecialistResult(input)
	if result != "" {
		t.Errorf("expected empty string after trim, got %q", result)
	}
}

func TestCleanSpecialistResult_MultilineContent(t *testing.T) {
	oa := newTestOrchestratorAgent()
	input := `{"status":"complete"}
Line one.
Line two.
Line three.`

	result := oa.cleanSpecialistResult(input)
	expected := "Line one.\nLine two.\nLine three."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCleanSpecialistResult_BracesNotJSON(t *testing.T) {
	oa := newTestOrchestratorAgent()
	// First line starts with { but does not look like JSON (no closing brace)
	input := "This line has no json.\nSecond line."

	result := oa.cleanSpecialistResult(input)
	// Since first line doesn't start with '{', original is returned
	if result != input {
		t.Errorf("expected original text, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// aggregateReports
// ---------------------------------------------------------------------------

func TestAggregateReports_EmptyReports(t *testing.T) {
	oa := newTestOrchestratorAgent()
	tc := &TeamContext{}

	result := oa.aggregateReports(nil, tc)
	if result != "No specialist reports to aggregate." {
		t.Errorf("expected no-reports message, got %q", result)
	}
}

func TestAggregateReports_EmptySlice(t *testing.T) {
	oa := newTestOrchestratorAgent()
	tc := &TeamContext{}

	result := oa.aggregateReports([]SpecialistReport{}, tc)
	if result != "No specialist reports to aggregate." {
		t.Errorf("expected no-reports message, got %q", result)
	}
}

func TestAggregateReports_SingleCompleteReport(t *testing.T) {
	oa := newTestOrchestratorAgent()
	tc := &TeamContext{}
	reports := []SpecialistReport{
		{
			SpecialistName: "developer",
			Status:         "complete",
			Confidence:     0.95,
			Result:         "Implemented the feature.",
		},
	}

	result := oa.aggregateReports(reports, tc)

	if !strings.Contains(result, "Task Completion Summary") {
		t.Error("expected summary header")
	}
	if !strings.Contains(result, "developer") {
		t.Error("expected developer name in result")
	}
	if !strings.Contains(result, "95%") {
		t.Error("expected confidence percentage in result")
	}
	if !strings.Contains(result, "Implemented the feature.") {
		t.Error("expected result content")
	}
}

func TestAggregateReports_MixedStatuses(t *testing.T) {
	oa := newTestOrchestratorAgent()
	tc := &TeamContext{}
	reports := []SpecialistReport{
		{
			SpecialistName: "developer",
			Status:         "complete",
			Confidence:     0.9,
			Result:         "Code written.",
		},
		{
			SpecialistName: "tester",
			Status:         "failed",
			Confidence:     0.2,
			Result:         "Tests failed.",
		},
		{
			SpecialistName: "docs",
			Status:         "partial",
			Confidence:     0.6,
			Result:         "Partial docs.",
		},
	}

	result := oa.aggregateReports(reports, tc)

	// "complete" and "partial" should be included
	if !strings.Contains(result, "developer") {
		t.Error("expected developer in result")
	}
	if !strings.Contains(result, "Code written.") {
		t.Error("expected developer result")
	}
	if !strings.Contains(result, "docs") {
		t.Error("expected docs in result")
	}
	if !strings.Contains(result, "Partial docs.") {
		t.Error("expected partial docs result")
	}
	// "failed" status should NOT appear in contributions
	if strings.Contains(result, "Tests failed.") {
		t.Error("expected failed result to be excluded from contributions")
	}
}

func TestAggregateReports_WithArtifacts(t *testing.T) {
	oa := newTestOrchestratorAgent()
	tc := &TeamContext{}
	reports := []SpecialistReport{
		{
			SpecialistName: "developer",
			Status:         "complete",
			Confidence:     0.9,
			Result:         "Done.",
			Artifacts:      []string{"main.go", "handler.go"},
		},
		{
			SpecialistName: "tester",
			Status:         "complete",
			Confidence:     0.8,
			Result:         "Tests pass.",
			Artifacts:      []string{"main_test.go"},
		},
	}

	result := oa.aggregateReports(reports, tc)

	if !strings.Contains(result, "Artifacts Created") {
		t.Error("expected artifacts section")
	}
	if !strings.Contains(result, "main.go") {
		t.Error("expected main.go in artifacts")
	}
	if !strings.Contains(result, "main_test.go") {
		t.Error("expected main_test.go in artifacts")
	}
}

func TestAggregateReports_WithCommunications(t *testing.T) {
	oa := newTestOrchestratorAgent()
	tc := &TeamContext{
		Communications: []TeamComm{
			{
				From:      "developer",
				To:        "orchestrator",
				Message:   "Completed with status: complete, confidence: 90%",
				Type:      "response",
				Timestamp: time.Now(),
			},
		},
	}
	reports := []SpecialistReport{
		{
			SpecialistName: "developer",
			Status:         "complete",
			Confidence:     0.9,
			Result:         "Done.",
		},
	}

	result := oa.aggregateReports(reports, tc)

	if !strings.Contains(result, "Team Collaboration") {
		t.Error("expected team collaboration section")
	}
}

func TestAggregateReports_NoArtifacts(t *testing.T) {
	oa := newTestOrchestratorAgent()
	tc := &TeamContext{}
	reports := []SpecialistReport{
		{
			SpecialistName: "developer",
			Status:         "complete",
			Confidence:     0.9,
			Result:         "Done.",
		},
	}

	result := oa.aggregateReports(reports, tc)

	if strings.Contains(result, "Artifacts Created") {
		t.Error("expected no artifacts section when no artifacts exist")
	}
}

// ---------------------------------------------------------------------------
// ContextWithTeamContext / TeamContextFromCtx
// ---------------------------------------------------------------------------

func TestTeamContextRoundTrip(t *testing.T) {
	ctx := context.Background()
	tc := &TeamContext{
		TaskID:       "test_123",
		OriginalTask: "do something",
		Progress:     map[string]string{"dev": "working"},
	}

	ctx = ContextWithTeamContext(ctx, tc)
	retrieved := TeamContextFromCtx(ctx)

	if retrieved == nil {
		t.Fatal("expected non-nil TeamContext from context")
	}
	if retrieved.TaskID != "test_123" {
		t.Errorf("expected TaskID 'test_123', got %q", retrieved.TaskID)
	}
	if retrieved.OriginalTask != "do something" {
		t.Errorf("expected OriginalTask 'do something', got %q", retrieved.OriginalTask)
	}
	if retrieved.Progress["dev"] != "working" {
		t.Errorf("expected Progress['dev'] = 'working', got %q", retrieved.Progress["dev"])
	}
}

func TestTeamContextFromCtx_Missing(t *testing.T) {
	ctx := context.Background()
	retrieved := TeamContextFromCtx(ctx)
	if retrieved != nil {
		t.Error("expected nil TeamContext from empty context")
	}
}

// ---------------------------------------------------------------------------
// ContextWithAgentTracker / agentTrackerFromCtx
// ---------------------------------------------------------------------------

func TestAgentTrackerRoundTrip(t *testing.T) {
	ctx := context.Background()
	tracker := &AgentLoop{}

	ctx = ContextWithAgentTracker(ctx, tracker)
	retrieved := agentTrackerFromCtx(ctx)

	if retrieved == nil {
		t.Fatal("expected non-nil tracker from context")
	}
	if retrieved != tracker {
		t.Error("expected same tracker instance")
	}
}

func TestAgentTrackerFromCtx_Missing(t *testing.T) {
	ctx := context.Background()
	retrieved := agentTrackerFromCtx(ctx)
	if retrieved != nil {
		t.Error("expected nil tracker from empty context")
	}
}

func TestContextWithSessionKeyRoundTrip(t *testing.T) {
	ctx := context.Background()
	if got := sessionKeyFromCtx(ctx); got != "" {
		t.Fatalf("expected empty session key from empty context, got %q", got)
	}

	ctx = ContextWithSessionKey(ctx, "web:chat:abc123")
	if got := sessionKeyFromCtx(ctx); got != "web:chat:abc123" {
		t.Fatalf("expected session key to round-trip, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// ContextWithAgentStatusCallback / agentStatusCallbackFromCtx / emitAgentStatus
// ---------------------------------------------------------------------------

func TestAgentStatusCallbackRoundTrip(t *testing.T) {
	ctx := context.Background()
	var received AgentStatusEvent

	callback := func(ev AgentStatusEvent) error {
		received = ev
		return nil
	}

	ctx = ContextWithAgentStatusCallback(ctx, callback)
	retrieved := agentStatusCallbackFromCtx(ctx)

	if retrieved == nil {
		t.Fatal("expected non-nil callback from context")
	}

	_ = retrieved(AgentStatusEvent{Agent: "test", Status: "working"})
	if received.Agent != "test" {
		t.Errorf("expected Agent 'test', got %q", received.Agent)
	}
}

func TestAgentStatusCallbackFromCtx_Missing(t *testing.T) {
	ctx := context.Background()
	retrieved := agentStatusCallbackFromCtx(ctx)
	if retrieved != nil {
		t.Error("expected nil callback from empty context")
	}
}

func TestEmitAgentStatus_WithCallback(t *testing.T) {
	var received AgentStatusEvent
	ctx := ContextWithAgentStatusCallback(context.Background(), func(ev AgentStatusEvent) error {
		received = ev
		return nil
	})

	emitAgentStatus(ctx, AgentStatusEvent{Agent: "orchestrator", Status: "analyzing"})

	if received.Agent != "orchestrator" {
		t.Errorf("expected Agent 'orchestrator', got %q", received.Agent)
	}
	if received.Status != "analyzing" {
		t.Errorf("expected Status 'analyzing', got %q", received.Status)
	}
}

func TestEmitAgentStatus_WithoutCallback(t *testing.T) {
	// Should not panic when no callback is set
	ctx := context.Background()
	emitAgentStatus(ctx, AgentStatusEvent{Agent: "test", Status: "working"})
}

func TestAggregateReportsWithStatus_EmitsSynthesisEvents(t *testing.T) {
	oa := newTestOrchestratorAgent()
	teamCtx := &TeamContext{}
	reports := []SpecialistReport{{SpecialistName: "developer", Status: "complete", Result: "done", Confidence: 1}}

	var events []AgentStatusEvent
	ctx := ContextWithAgentStatusCallback(context.Background(), func(ev AgentStatusEvent) error {
		events = append(events, ev)
		return nil
	})

	result := oa.aggregateReportsWithStatus(ctx, reports, teamCtx)

	if !strings.Contains(result, "developer") {
		t.Fatalf("expected aggregated result to include specialist content, got %q", result)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 synthesis events, got %d", len(events))
	}
	if events[0].Status != "synthesis_start" {
		t.Fatalf("expected first event synthesis_start, got %q", events[0].Status)
	}
	if events[1].Status != "synthesis_end" {
		t.Fatalf("expected second event synthesis_end, got %q", events[1].Status)
	}
	if events[0].Agent != "orchestrator" || events[1].Agent != "orchestrator" {
		t.Fatal("expected orchestrator synthesis events")
	}
}

func TestAggregateReportsWithStatus_SkipsSynthesisEventsWithoutReports(t *testing.T) {
	oa := newTestOrchestratorAgent()
	teamCtx := &TeamContext{}

	called := false
	ctx := ContextWithAgentStatusCallback(context.Background(), func(ev AgentStatusEvent) error {
		called = true
		return nil
	})

	result := oa.aggregateReportsWithStatus(ctx, nil, teamCtx)

	if result != "No specialist reports to aggregate." {
		t.Fatalf("unexpected aggregate result: %q", result)
	}
	if called {
		t.Fatal("expected no synthesis events when there are no specialist reports")
	}
}

func TestProcessSpecialistTask_PropagatesSessionKeyToSpecialist(t *testing.T) {
	recorder := &sessionRecordingTool{}
	provider := &sequenceProvider{responses: []*providers.LLMResponse{{
		ToolCalls: []providers.ToolCall{{
			ID:        "tool-1",
			Name:      recorder.Name(),
			Arguments: map[string]interface{}{},
		}},
		FinishReason: "tool_calls",
	}, {
		Content:      "specialist complete",
		FinishReason: "stop",
	}}}

	specialist := newTestSpecialistAgent(t, "developer", provider)
	specialist.tools.Register(recorder)

	registry := NewSpecialistRegistry()
	if err := registry.RegisterSpecialist(specialist); err != nil {
		t.Fatalf("failed to register specialist: %v", err)
	}

	oa := &OrchestratorAgent{
		SpecialistAgent: &SpecialistAgent{AgentLoop: &AgentLoop{}},
		registry:        registry,
	}
	ctx := ContextWithSessionKey(context.Background(), "web:chat:abc123")

	result, err := oa.processSpecialistTask(ctx, "developer", "review this", sessionKeyFromCtx(ctx))
	if err != nil {
		t.Fatalf("processSpecialistTask returned error: %v", err)
	}
	if !strings.Contains(result, "specialist complete") {
		t.Fatalf("expected specialist response, got %q", result)
	}
	if recorder.lastSessionKey != "web:chat:abc123" {
		t.Fatalf("expected propagated session key, got %q", recorder.lastSessionKey)
	}
}

func TestRunAgentLoop_AggregatesSpecialistReportsInStandardFlow(t *testing.T) {
	provider := &sequenceProvider{responses: []*providers.LLMResponse{{
		ToolCalls: []providers.ToolCall{{
			ID:   "del-1",
			Name: "delegate_to_specialist",
			Arguments: map[string]interface{}{
				"specialist_name": "developer",
				"task":            "implement feature",
			},
		}},
		FinishReason: "tool_calls",
	}, {
		Content:      "orchestrator raw summary",
		FinishReason: "stop",
	}}}

	specialistProvider := &staticProvider{response: "```json\n{\"status\":\"complete\",\"confidence\":0.9,\"request_help\":\"\",\"suggestion\":\"\"}\n```\nImplemented feature cleanly."}
	specialist := newTestSpecialistAgent(t, "developer", specialistProvider)

	registry := NewSpecialistRegistry()
	if err := registry.RegisterSpecialist(specialist); err != nil {
		t.Fatalf("failed to register specialist: %v", err)
	}

	cfg := newAgentTestConfig(t.TempDir())
	loop := NewAgentLoop(cfg, bus.NewMessageBus(), provider)
	orchestratorSpecialist := &SpecialistAgent{
		AgentLoop:    loop,
		name:         "orchestrator",
		description:  "test orchestrator",
		allowedTools: map[string]bool{},
	}

	oa := &OrchestratorAgent{
		SpecialistAgent: orchestratorSpecialist,
		registry:        registry,
	}
	loop.orchestratorOwner = oa
	oa.registerDelegationTool()

	result, err := loop.ProcessDirect(context.Background(), "help me", "web:chat:abc123")
	if err != nil {
		t.Fatalf("ProcessDirect returned error: %v", err)
	}
	if strings.Contains(result, "orchestrator raw summary") {
		t.Fatalf("expected aggregated specialist summary instead of raw orchestrator content, got %q", result)
	}
	if !strings.Contains(result, "Task Completion Summary") {
		t.Fatalf("expected aggregated report header, got %q", result)
	}
	if !strings.Contains(result, "developer") || !strings.Contains(result, "Implemented feature cleanly.") {
		t.Fatalf("expected aggregated specialist contribution, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// ContextWithContentSegmentCallback / contentSegmentCallbackFromCtx / emitContentSegment
// ---------------------------------------------------------------------------

func TestContentSegmentCallbackRoundTrip(t *testing.T) {
	ctx := context.Background()
	var received ContentSegment

	callback := func(seg ContentSegment) error {
		received = seg
		return nil
	}

	ctx = ContextWithContentSegmentCallback(ctx, callback)
	retrieved := contentSegmentCallbackFromCtx(ctx)

	if retrieved == nil {
		t.Fatal("expected non-nil callback from context")
	}

	_ = retrieved(ContentSegment{Agent: "dev", Content: "hello"})
	if received.Agent != "dev" {
		t.Errorf("expected Agent 'dev', got %q", received.Agent)
	}
	if received.Content != "hello" {
		t.Errorf("expected Content 'hello', got %q", received.Content)
	}
}

func TestContentSegmentCallbackFromCtx_Missing(t *testing.T) {
	ctx := context.Background()
	if contentSegmentCallbackFromCtx(ctx) != nil {
		t.Error("expected nil callback from empty context")
	}
}

func TestEmitContentSegment_WithCallback(t *testing.T) {
	var received ContentSegment
	ctx := ContextWithContentSegmentCallback(context.Background(), func(seg ContentSegment) error {
		received = seg
		return nil
	})

	emitContentSegment(ctx, ContentSegment{Agent: "tester", Content: "result"})

	if received.Agent != "tester" {
		t.Errorf("expected Agent 'tester', got %q", received.Agent)
	}
}

func TestEmitContentSegment_WithoutCallback(t *testing.T) {
	// Should not panic
	ctx := context.Background()
	emitContentSegment(ctx, ContentSegment{Agent: "test", Content: "x"})
}

// ---------------------------------------------------------------------------
// ContextWithSpecialistReportCallback / specialistReportCallbackFromCtx / emitSpecialistReport
// ---------------------------------------------------------------------------

func TestSpecialistReportCallbackRoundTrip(t *testing.T) {
	ctx := context.Background()
	var received *SpecialistReport

	callback := func(report *SpecialistReport) error {
		received = report
		return nil
	}

	ctx = ContextWithSpecialistReportCallback(ctx, callback)
	retrieved := specialistReportCallbackFromCtx(ctx)

	if retrieved == nil {
		t.Fatal("expected non-nil callback from context")
	}

	report := &SpecialistReport{SpecialistName: "dev", Status: "complete"}
	_ = retrieved(report)
	if received == nil || received.SpecialistName != "dev" {
		t.Error("expected received report with specialist name 'dev'")
	}
}

func TestEmitSpecialistReport_WithoutCallback(t *testing.T) {
	// Should not panic
	ctx := context.Background()
	emitSpecialistReport(ctx, &SpecialistReport{SpecialistName: "test"})
}

// ---------------------------------------------------------------------------
// storeLastReport / getLastSpecialistReport
// ---------------------------------------------------------------------------

func TestStoreAndGetLastReport(t *testing.T) {
	oa := newTestOrchestratorAgent()

	// Initially nil
	if oa.getLastSpecialistReport() != nil {
		t.Error("expected nil initial report")
	}

	report := &SpecialistReport{SpecialistName: "dev", Status: "complete"}
	oa.storeLastReport(report)

	retrieved := oa.getLastSpecialistReport()
	if retrieved == nil {
		t.Fatal("expected non-nil report after store")
	}
	if retrieved.SpecialistName != "dev" {
		t.Errorf("expected specialist name 'dev', got %q", retrieved.SpecialistName)
	}
}

// ---------------------------------------------------------------------------
// DelegationTool metadata
// ---------------------------------------------------------------------------

func TestDelegationTool_Name(t *testing.T) {
	dt := &DelegationTool{}
	if dt.Name() != "delegate_to_specialist" {
		t.Errorf("expected 'delegate_to_specialist', got %q", dt.Name())
	}
}

func TestDelegationTool_Description(t *testing.T) {
	dt := &DelegationTool{}
	if dt.Description() == "" {
		t.Error("expected non-empty description")
	}
}

func TestDelegationTool_Parameters(t *testing.T) {
	dt := &DelegationTool{}
	params := dt.Parameters()
	if params == nil {
		t.Fatal("expected non-nil parameters")
	}
	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected properties map")
	}
	if _, ok := props["specialist_name"]; !ok {
		t.Error("expected specialist_name property")
	}
	if _, ok := props["task"]; !ok {
		t.Error("expected task property")
	}
	if _, ok := props["context"]; !ok {
		t.Error("expected context property")
	}
}

// ---------------------------------------------------------------------------
// TaskDecompositionTool metadata
// ---------------------------------------------------------------------------

func TestTaskDecompositionTool_Name(t *testing.T) {
	tdt := &TaskDecompositionTool{}
	if tdt.Name() != "decompose_task" {
		t.Errorf("expected 'decompose_task', got %q", tdt.Name())
	}
}

func TestTaskDecompositionTool_Description(t *testing.T) {
	tdt := &TaskDecompositionTool{}
	if tdt.Description() == "" {
		t.Error("expected non-empty description")
	}
}

func TestTaskDecompositionTool_Parameters(t *testing.T) {
	tdt := &TaskDecompositionTool{}
	params := tdt.Parameters()
	if params == nil {
		t.Fatal("expected non-nil parameters")
	}
	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("expected required array")
	}
	foundTask := false
	foundSubtasks := false
	for _, r := range required {
		if r == "task" {
			foundTask = true
		}
		if r == "subtasks" {
			foundSubtasks = true
		}
	}
	if !foundTask {
		t.Error("expected 'task' in required")
	}
	if !foundSubtasks {
		t.Error("expected 'subtasks' in required")
	}
}

func TestTaskDecompositionTool_PopulatesTeamContextUserIdentity(t *testing.T) {
	teamCtx := &TeamContext{Progress: map[string]string{}}
	oa := &OrchestratorAgent{
		SpecialistAgent: &SpecialistAgent{
			AgentLoop: &AgentLoop{
				userUUID: "aaa-111",
				userID:   42,
			},
		},
	}
	tool := &TaskDecompositionTool{orchestrator: oa}

	_, err := tool.Execute(ContextWithTeamContext(context.Background(), teamCtx), map[string]interface{}{
		"task":     "break this down",
		"subtasks": []interface{}{},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if teamCtx.UserUUID != "aaa-111" {
		t.Fatalf("expected TeamContext.UserUUID aaa-111, got %q", teamCtx.UserUUID)
	}
	if teamCtx.UserID != 42 {
		t.Fatalf("expected TeamContext.UserID 42, got %d", teamCtx.UserID)
	}
}

// ---------------------------------------------------------------------------
// DelegationRequest / DelegationResult struct construction
// ---------------------------------------------------------------------------

func TestDelegationRequestFields(t *testing.T) {
	req := DelegationRequest{
		SpecialistName: "developer",
		Task:           "write code",
		Context:        "for the web module",
	}

	if req.SpecialistName != "developer" {
		t.Errorf("expected 'developer', got %q", req.SpecialistName)
	}
	if req.Task != "write code" {
		t.Errorf("expected 'write code', got %q", req.Task)
	}
	if req.Context != "for the web module" {
		t.Errorf("expected 'for the web module', got %q", req.Context)
	}
}

func TestDelegationResultFields(t *testing.T) {
	res := DelegationResult{
		SpecialistName: "tester",
		Result:         "all tests pass",
		Success:        true,
	}

	if res.SpecialistName != "tester" {
		t.Errorf("expected 'tester', got %q", res.SpecialistName)
	}
	if !res.Success {
		t.Error("expected Success to be true")
	}
}

// ---------------------------------------------------------------------------
// TeamContext, TeamNote, TeamDecision, TeamComm structs
// ---------------------------------------------------------------------------

func TestTeamContext_SubtaskDependencyTracking(t *testing.T) {
	tc := TeamContext{
		TaskID:       "task_1",
		OriginalTask: "build feature",
		Subtasks: []Subtask{
			{ID: 0, Description: "design", Specialist: "architect", Status: "pending", Priority: 1},
			{ID: 1, Description: "implement", Specialist: "developer", Dependencies: []int{0}, Status: "pending", Priority: 2},
			{ID: 2, Description: "test", Specialist: "tester", Dependencies: []int{1}, Status: "pending", Priority: 3},
		},
		Progress: map[string]string{},
	}

	if len(tc.Subtasks) != 3 {
		t.Errorf("expected 3 subtasks, got %d", len(tc.Subtasks))
	}

	// Verify dependency chain
	if len(tc.Subtasks[0].Dependencies) != 0 {
		t.Error("first subtask should have no dependencies")
	}
	if len(tc.Subtasks[1].Dependencies) != 1 || tc.Subtasks[1].Dependencies[0] != 0 {
		t.Error("second subtask should depend on first")
	}
	if len(tc.Subtasks[2].Dependencies) != 1 || tc.Subtasks[2].Dependencies[0] != 1 {
		t.Error("third subtask should depend on second")
	}
}

func TestTeamNote_Fields(t *testing.T) {
	note := TeamNote{
		Author:    "developer",
		Content:   "Found a potential issue",
		Timestamp: time.Now(),
		ForAgent:  "tester",
	}

	if note.Author != "developer" {
		t.Errorf("expected author 'developer', got %q", note.Author)
	}
	if note.ForAgent != "tester" {
		t.Errorf("expected ForAgent 'tester', got %q", note.ForAgent)
	}
}

func TestTeamDecision_Fields(t *testing.T) {
	decision := TeamDecision{
		Author:     "orchestrator",
		Decision:   "Use REST instead of GraphQL",
		Rationale:  "Simpler for MVP",
		Timestamp:  time.Now(),
		AffectedBy: []string{"developer", "frontend"},
	}

	if decision.Decision != "Use REST instead of GraphQL" {
		t.Error("unexpected decision text")
	}
	if len(decision.AffectedBy) != 2 {
		t.Errorf("expected 2 affected agents, got %d", len(decision.AffectedBy))
	}
}

func TestTeamComm_Fields(t *testing.T) {
	comm := TeamComm{
		From:      "developer",
		To:        "tester",
		Message:   "Implementation ready for testing",
		Type:      "info",
		Timestamp: time.Now(),
	}

	if comm.From != "developer" {
		t.Errorf("expected from 'developer', got %q", comm.From)
	}
	if comm.Type != "info" {
		t.Errorf("expected type 'info', got %q", comm.Type)
	}
}
