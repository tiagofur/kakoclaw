package observability

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Mock MetricsStorage
// ---------------------------------------------------------------------------

type mockMetricsStorage struct {
	counters        map[string]int64
	breakdownsModel json.RawMessage
	breakdownsTool  json.RawMessage
	events          []json.RawMessage
	saveCalled      bool

	loadCountersCalled    bool
	loadBreakdownsCalled  bool
	loadEventsCalled      bool
	appendEventCalled     bool
	saveCountersCalled    bool
	saveBreakdownsCalled  bool
}

func newMockStorage() *mockMetricsStorage {
	return &mockMetricsStorage{
		counters: make(map[string]int64),
		events:   make([]json.RawMessage, 0),
	}
}

func (m *mockMetricsStorage) SaveMetricCounters(counters map[string]int64) error {
	m.saveCountersCalled = true
	m.saveCalled = true
	m.counters = counters
	return nil
}

func (m *mockMetricsStorage) LoadMetricCounters() (map[string]int64, error) {
	m.loadCountersCalled = true
	return m.counters, nil
}

func (m *mockMetricsStorage) SaveMetricBreakdowns(modelMetrics json.RawMessage, toolMetrics json.RawMessage) error {
	m.saveBreakdownsCalled = true
	m.breakdownsModel = modelMetrics
	m.breakdownsTool = toolMetrics
	return nil
}

func (m *mockMetricsStorage) LoadMetricBreakdowns() (json.RawMessage, json.RawMessage, error) {
	m.loadBreakdownsCalled = true
	return m.breakdownsModel, m.breakdownsTool, nil
}

func (m *mockMetricsStorage) AppendMetricEvent(event interface{}) error {
	m.appendEventCalled = true
	raw, _ := json.Marshal(event)
	m.events = append(m.events, raw)
	return nil
}

func (m *mockMetricsStorage) LoadRecentEvents() ([]json.RawMessage, error) {
	m.loadEventsCalled = true
	return m.events, nil
}

// ---------------------------------------------------------------------------
// TestNew_CreatesEmptyMetrics
// ---------------------------------------------------------------------------

func TestNew_CreatesEmptyMetrics(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("New() returned nil")
	}

	// All scalar counters should be zero
	if m.LLMCalls != 0 {
		t.Errorf("LLMCalls = %d; want 0", m.LLMCalls)
	}
	if m.LLMErrors != 0 {
		t.Errorf("LLMErrors = %d; want 0", m.LLMErrors)
	}
	if m.LLMTotalMs != 0 {
		t.Errorf("LLMTotalMs = %d; want 0", m.LLMTotalMs)
	}
	if m.LLMTokensIn != 0 {
		t.Errorf("LLMTokensIn = %d; want 0", m.LLMTokensIn)
	}
	if m.LLMTokensOut != 0 {
		t.Errorf("LLMTokensOut = %d; want 0", m.LLMTokensOut)
	}
	if m.ToolCalls != 0 {
		t.Errorf("ToolCalls = %d; want 0", m.ToolCalls)
	}
	if m.ToolErrors != 0 {
		t.Errorf("ToolErrors = %d; want 0", m.ToolErrors)
	}
	if m.ToolTotalMs != 0 {
		t.Errorf("ToolTotalMs = %d; want 0", m.ToolTotalMs)
	}
	if m.AgentRuns != 0 {
		t.Errorf("AgentRuns = %d; want 0", m.AgentRuns)
	}
	if m.AgentErrors != 0 {
		t.Errorf("AgentErrors = %d; want 0", m.AgentErrors)
	}
	if m.AgentTotalMs != 0 {
		t.Errorf("AgentTotalMs = %d; want 0", m.AgentTotalMs)
	}
	if m.AgentIterTotal != 0 {
		t.Errorf("AgentIterTotal = %d; want 0", m.AgentIterTotal)
	}

	// Maps should be initialized but empty
	if m.LLMByModel == nil {
		t.Error("LLMByModel is nil; want initialized empty map")
	}
	if len(m.LLMByModel) != 0 {
		t.Errorf("len(LLMByModel) = %d; want 0", len(m.LLMByModel))
	}
	if m.ToolByName == nil {
		t.Error("ToolByName is nil; want initialized empty map")
	}
	if len(m.ToolByName) != 0 {
		t.Errorf("len(ToolByName) = %d; want 0", len(m.ToolByName))
	}

	// RecentEvents should be initialized but empty
	if m.RecentEvents == nil {
		t.Error("RecentEvents is nil; want initialized empty slice")
	}
	if len(m.RecentEvents) != 0 {
		t.Errorf("len(RecentEvents) = %d; want 0", len(m.RecentEvents))
	}

	// StartedAt should be set to approximately now
	if time.Since(m.StartedAt) > 5*time.Second {
		t.Errorf("StartedAt seems too old: %v", m.StartedAt)
	}
}

// ---------------------------------------------------------------------------
// TestRecordLLMCall_IncrementsCounters
// ---------------------------------------------------------------------------

func TestRecordLLMCall_IncrementsCounters(t *testing.T) {
	m := New()

	// Record a successful call
	m.RecordLLMCall("gpt-4", 100*time.Millisecond, 50, 200, nil)

	if m.LLMCalls != 1 {
		t.Errorf("LLMCalls = %d; want 1", m.LLMCalls)
	}
	if m.LLMErrors != 0 {
		t.Errorf("LLMErrors = %d; want 0", m.LLMErrors)
	}
	if m.LLMTokensIn != 50 {
		t.Errorf("LLMTokensIn = %d; want 50", m.LLMTokensIn)
	}
	if m.LLMTokensOut != 200 {
		t.Errorf("LLMTokensOut = %d; want 200", m.LLMTokensOut)
	}
	if m.LLMTotalMs != 100 {
		t.Errorf("LLMTotalMs = %d; want 100", m.LLMTotalMs)
	}

	// Record a call with error
	m.RecordLLMCall("gpt-4", 200*time.Millisecond, 30, 0, errors.New("timeout"))

	if m.LLMCalls != 2 {
		t.Errorf("LLMCalls = %d; want 2", m.LLMCalls)
	}
	if m.LLMErrors != 1 {
		t.Errorf("LLMErrors = %d; want 1", m.LLMErrors)
	}
	if m.LLMTokensIn != 80 {
		t.Errorf("LLMTokensIn = %d; want 80 (50+30)", m.LLMTokensIn)
	}
	if m.LLMTokensOut != 200 {
		t.Errorf("LLMTokensOut = %d; want 200 (200+0)", m.LLMTokensOut)
	}
	if m.LLMTotalMs != 300 {
		t.Errorf("LLMTotalMs = %d; want 300 (100+200)", m.LLMTotalMs)
	}

	// Record a call with a different model
	m.RecordLLMCall("claude-3", 150*time.Millisecond, 100, 500, nil)

	if m.LLMCalls != 3 {
		t.Errorf("LLMCalls = %d; want 3", m.LLMCalls)
	}

	// Verify LLMByModel map
	gpt4, ok := m.LLMByModel["gpt-4"]
	if !ok {
		t.Fatal("LLMByModel[\"gpt-4\"] not found")
	}
	if gpt4.Calls != 2 {
		t.Errorf("gpt-4 Calls = %d; want 2", gpt4.Calls)
	}
	if gpt4.Errors != 1 {
		t.Errorf("gpt-4 Errors = %d; want 1", gpt4.Errors)
	}
	if gpt4.TokensIn != 80 {
		t.Errorf("gpt-4 TokensIn = %d; want 80", gpt4.TokensIn)
	}
	if gpt4.TokensOut != 200 {
		t.Errorf("gpt-4 TokensOut = %d; want 200", gpt4.TokensOut)
	}
	if gpt4.TotalMs != 300 {
		t.Errorf("gpt-4 TotalMs = %d; want 300", gpt4.TotalMs)
	}

	claude, ok := m.LLMByModel["claude-3"]
	if !ok {
		t.Fatal("LLMByModel[\"claude-3\"] not found")
	}
	if claude.Calls != 1 {
		t.Errorf("claude-3 Calls = %d; want 1", claude.Calls)
	}
	if claude.Errors != 0 {
		t.Errorf("claude-3 Errors = %d; want 0", claude.Errors)
	}
	if claude.TokensIn != 100 {
		t.Errorf("claude-3 TokensIn = %d; want 100", claude.TokensIn)
	}
	if claude.TokensOut != 500 {
		t.Errorf("claude-3 TokensOut = %d; want 500", claude.TokensOut)
	}
}

// ---------------------------------------------------------------------------
// TestRecordToolCall_IncrementsCounters
// ---------------------------------------------------------------------------

func TestRecordToolCall_IncrementsCounters(t *testing.T) {
	m := New()

	// Record successful tool calls
	m.RecordToolCall("read_file", 50*time.Millisecond, nil)
	m.RecordToolCall("read_file", 30*time.Millisecond, nil)
	m.RecordToolCall("exec", 500*time.Millisecond, nil)

	// Record a tool call with error
	m.RecordToolCall("exec", 100*time.Millisecond, errors.New("permission denied"))

	if m.ToolCalls != 4 {
		t.Errorf("ToolCalls = %d; want 4", m.ToolCalls)
	}
	if m.ToolErrors != 1 {
		t.Errorf("ToolErrors = %d; want 1", m.ToolErrors)
	}
	if m.ToolTotalMs != 680 {
		t.Errorf("ToolTotalMs = %d; want 680 (50+30+500+100)", m.ToolTotalMs)
	}

	// Verify ToolByName map
	readFile, ok := m.ToolByName["read_file"]
	if !ok {
		t.Fatal("ToolByName[\"read_file\"] not found")
	}
	if readFile.Calls != 2 {
		t.Errorf("read_file Calls = %d; want 2", readFile.Calls)
	}
	if readFile.Errors != 0 {
		t.Errorf("read_file Errors = %d; want 0", readFile.Errors)
	}
	if readFile.TotalMs != 80 {
		t.Errorf("read_file TotalMs = %d; want 80 (50+30)", readFile.TotalMs)
	}

	exec, ok := m.ToolByName["exec"]
	if !ok {
		t.Fatal("ToolByName[\"exec\"] not found")
	}
	if exec.Calls != 2 {
		t.Errorf("exec Calls = %d; want 2", exec.Calls)
	}
	if exec.Errors != 1 {
		t.Errorf("exec Errors = %d; want 1", exec.Errors)
	}
	if exec.TotalMs != 600 {
		t.Errorf("exec TotalMs = %d; want 600 (500+100)", exec.TotalMs)
	}
}

// ---------------------------------------------------------------------------
// TestRecordAgentRun_TracksDuration
// ---------------------------------------------------------------------------

func TestRecordAgentRun_TracksDuration(t *testing.T) {
	m := New()

	// Record successful agent run
	m.RecordAgentRun(2*time.Second, 5, nil)

	if m.AgentRuns != 1 {
		t.Errorf("AgentRuns = %d; want 1", m.AgentRuns)
	}
	if m.AgentErrors != 0 {
		t.Errorf("AgentErrors = %d; want 0", m.AgentErrors)
	}
	if m.AgentTotalMs != 2000 {
		t.Errorf("AgentTotalMs = %d; want 2000", m.AgentTotalMs)
	}
	if m.AgentIterTotal != 5 {
		t.Errorf("AgentIterTotal = %d; want 5", m.AgentIterTotal)
	}

	// Record agent run with error
	m.RecordAgentRun(1*time.Second, 3, errors.New("context cancelled"))

	if m.AgentRuns != 2 {
		t.Errorf("AgentRuns = %d; want 2", m.AgentRuns)
	}
	if m.AgentErrors != 1 {
		t.Errorf("AgentErrors = %d; want 1", m.AgentErrors)
	}
	if m.AgentTotalMs != 3000 {
		t.Errorf("AgentTotalMs = %d; want 3000 (2000+1000)", m.AgentTotalMs)
	}
	if m.AgentIterTotal != 8 {
		t.Errorf("AgentIterTotal = %d; want 8 (5+3)", m.AgentIterTotal)
	}

	// Record another successful run
	m.RecordAgentRun(500*time.Millisecond, 2, nil)

	if m.AgentRuns != 3 {
		t.Errorf("AgentRuns = %d; want 3", m.AgentRuns)
	}
	if m.AgentErrors != 1 {
		t.Errorf("AgentErrors = %d; want 1 (unchanged)", m.AgentErrors)
	}
	if m.AgentIterTotal != 10 {
		t.Errorf("AgentIterTotal = %d; want 10 (5+3+2)", m.AgentIterTotal)
	}
}

// ---------------------------------------------------------------------------
// TestSnapshot_ReturnsCorrectData
// ---------------------------------------------------------------------------

func TestSnapshot_ReturnsCorrectData(t *testing.T) {
	m := New()

	// Record some LLM calls
	m.RecordLLMCall("gpt-4", 100*time.Millisecond, 50, 200, nil)
	m.RecordLLMCall("gpt-4", 200*time.Millisecond, 30, 100, nil)

	// Record some tool calls
	m.RecordToolCall("read_file", 50*time.Millisecond, nil)

	// Record an agent run
	m.RecordAgentRun(1*time.Second, 4, nil)

	snap := m.Snapshot()

	// Verify LLM counters
	if got, ok := snap["llm_calls"].(int64); !ok || got != 2 {
		t.Errorf("snap[llm_calls] = %v; want 2", snap["llm_calls"])
	}
	if got, ok := snap["llm_errors"].(int64); !ok || got != 0 {
		t.Errorf("snap[llm_errors] = %v; want 0", snap["llm_errors"])
	}
	if got, ok := snap["llm_total_ms"].(int64); !ok || got != 300 {
		t.Errorf("snap[llm_total_ms] = %v; want 300", snap["llm_total_ms"])
	}
	if got, ok := snap["llm_avg_ms"].(int64); !ok || got != 150 {
		t.Errorf("snap[llm_avg_ms] = %v; want 150 (300/2)", snap["llm_avg_ms"])
	}
	if got, ok := snap["llm_tokens_in"].(int64); !ok || got != 80 {
		t.Errorf("snap[llm_tokens_in] = %v; want 80", snap["llm_tokens_in"])
	}
	if got, ok := snap["llm_tokens_out"].(int64); !ok || got != 300 {
		t.Errorf("snap[llm_tokens_out] = %v; want 300", snap["llm_tokens_out"])
	}

	// Verify LLM by model breakdown
	models, ok := snap["llm_by_model"].(map[string]interface{})
	if !ok {
		t.Fatal("snap[llm_by_model] is not map[string]interface{}")
	}
	gpt4, ok := models["gpt-4"].(map[string]interface{})
	if !ok {
		t.Fatal("models[gpt-4] not found or wrong type")
	}
	if got, ok := gpt4["calls"].(int64); !ok || got != 2 {
		t.Errorf("gpt-4 calls = %v; want 2", gpt4["calls"])
	}
	if got, ok := gpt4["avg_ms"].(int64); !ok || got != 150 {
		t.Errorf("gpt-4 avg_ms = %v; want 150", gpt4["avg_ms"])
	}

	// Verify tool counters
	if got, ok := snap["tool_calls"].(int64); !ok || got != 1 {
		t.Errorf("snap[tool_calls] = %v; want 1", snap["tool_calls"])
	}
	if got, ok := snap["tool_avg_ms"].(int64); !ok || got != 50 {
		t.Errorf("snap[tool_avg_ms] = %v; want 50", snap["tool_avg_ms"])
	}

	// Verify tool by name breakdown
	tools, ok := snap["tool_by_name"].(map[string]interface{})
	if !ok {
		t.Fatal("snap[tool_by_name] is not map[string]interface{}")
	}
	readFile, ok := tools["read_file"].(map[string]interface{})
	if !ok {
		t.Fatal("tools[read_file] not found or wrong type")
	}
	if got, ok := readFile["calls"].(int64); !ok || got != 1 {
		t.Errorf("read_file calls = %v; want 1", readFile["calls"])
	}

	// Verify agent counters
	if got, ok := snap["agent_runs"].(int64); !ok || got != 1 {
		t.Errorf("snap[agent_runs] = %v; want 1", snap["agent_runs"])
	}
	if got, ok := snap["agent_avg_ms"].(int64); !ok || got != 1000 {
		t.Errorf("snap[agent_avg_ms] = %v; want 1000", snap["agent_avg_ms"])
	}
	if got, ok := snap["agent_iterations_total"].(int64); !ok || got != 4 {
		t.Errorf("snap[agent_iterations_total] = %v; want 4", snap["agent_iterations_total"])
	}
	if got, ok := snap["agent_avg_iterations"].(float64); !ok || got != 4.0 {
		t.Errorf("snap[agent_avg_iterations] = %v; want 4.0", snap["agent_avg_iterations"])
	}

	// Verify recent events are present
	events, ok := snap["recent_events"].([]Event)
	if !ok {
		t.Fatal("snap[recent_events] is not []Event")
	}
	// 2 LLM calls + 1 tool call + 1 agent run = 4 events
	if len(events) != 4 {
		t.Errorf("len(recent_events) = %d; want 4", len(events))
	}

	// Verify uptime is present and non-negative
	if _, ok := snap["uptime_seconds"].(int64); !ok {
		t.Error("snap[uptime_seconds] missing or wrong type")
	}
	if _, ok := snap["started_at"].(string); !ok {
		t.Error("snap[started_at] missing or wrong type")
	}
}

// ---------------------------------------------------------------------------
// TestAddEvent_RingBuffer
// ---------------------------------------------------------------------------

func TestAddEvent_RingBuffer(t *testing.T) {
	m := New()

	// Fill to capacity: addEvent is private, so we use RecordToolCall which
	// calls addEvent internally. Each RecordToolCall adds exactly one event.
	for i := 0; i < maxRecentEvents; i++ {
		m.RecordToolCall("original_tool", 1*time.Millisecond, nil)
	}

	if len(m.RecentEvents) != maxRecentEvents {
		t.Fatalf("len(RecentEvents) = %d; want %d after filling", len(m.RecentEvents), maxRecentEvents)
	}

	// Verify all events are the original tool
	for i, evt := range m.RecentEvents {
		if evt.Tool != "original_tool" {
			t.Errorf("RecentEvents[%d].Tool = %q; want \"original_tool\"", i, evt.Tool)
		}
	}

	// Add 5 more events beyond capacity with a different tool name
	for i := 0; i < 5; i++ {
		m.RecordToolCall("overflow_tool", 2*time.Millisecond, nil)
	}

	// Size should remain at max
	if len(m.RecentEvents) != maxRecentEvents {
		t.Errorf("len(RecentEvents) = %d; want %d after overflow", len(m.RecentEvents), maxRecentEvents)
	}

	// The first 5 original events should have been evicted.
	// The first event should still be "original_tool" (the 6th original event
	// is now at index 0).
	if m.RecentEvents[0].Tool != "original_tool" {
		t.Errorf("RecentEvents[0].Tool = %q; want \"original_tool\"", m.RecentEvents[0].Tool)
	}

	// The last 5 events should be the overflow_tool events
	for i := maxRecentEvents - 5; i < maxRecentEvents; i++ {
		if m.RecentEvents[i].Tool != "overflow_tool" {
			t.Errorf("RecentEvents[%d].Tool = %q; want \"overflow_tool\"", i, m.RecentEvents[i].Tool)
		}
	}

	// The events right before the overflow ones should still be original
	for i := maxRecentEvents - 6; i >= 0; i-- {
		if m.RecentEvents[i].Tool != "original_tool" {
			t.Errorf("RecentEvents[%d].Tool = %q; want \"original_tool\"", i, m.RecentEvents[i].Tool)
			break
		}
	}
}

// ---------------------------------------------------------------------------
// TestSetStorage_LoadsAndFlushes
// ---------------------------------------------------------------------------

func TestSetStorage_LoadsAndFlushes(t *testing.T) {
	m := New()

	// Pre-populate the mock storage with counter data so we can verify loading
	store := newMockStorage()
	store.counters = map[string]int64{
		"llm_calls":              10,
		"llm_errors":             2,
		"llm_total_ms":           5000,
		"llm_tokens_in":          1000,
		"llm_tokens_out":         3000,
		"tool_calls":             20,
		"tool_errors":            3,
		"tool_total_ms":          2000,
		"agent_runs":             5,
		"agent_errors":           1,
		"agent_total_ms":         10000,
		"agent_iterations_total": 25,
	}

	// Pre-populate model breakdown
	modelMap := map[string]*ModelMetrics{
		"gpt-4": {Calls: 10, Errors: 2, TotalMs: 5000, TokensIn: 1000, TokensOut: 3000},
	}
	store.breakdownsModel, _ = json.Marshal(modelMap)

	// Pre-populate tool breakdown
	toolMap := map[string]*ToolMetrics{
		"read_file": {Calls: 20, Errors: 3, TotalMs: 2000},
	}
	store.breakdownsTool, _ = json.Marshal(toolMap)

	// Pre-populate events
	evt := Event{Type: "llm_call", Model: "gpt-4", DurationMs: 500, Timestamp: time.Now()}
	rawEvt, _ := json.Marshal(evt)
	store.events = []json.RawMessage{rawEvt}

	// Call SetStorage which should load from DB
	m.SetStorage(store)

	// Verify LoadMetricCounters was called
	if !store.loadCountersCalled {
		t.Error("LoadMetricCounters was not called during SetStorage")
	}

	// Verify LoadMetricBreakdowns was called
	if !store.loadBreakdownsCalled {
		t.Error("LoadMetricBreakdowns was not called during SetStorage")
	}

	// Verify LoadRecentEvents was called
	if !store.loadEventsCalled {
		t.Error("LoadRecentEvents was not called during SetStorage")
	}

	// Verify counters were loaded
	if m.LLMCalls != 10 {
		t.Errorf("LLMCalls = %d; want 10 (loaded from storage)", m.LLMCalls)
	}
	if m.LLMErrors != 2 {
		t.Errorf("LLMErrors = %d; want 2", m.LLMErrors)
	}
	if m.ToolCalls != 20 {
		t.Errorf("ToolCalls = %d; want 20", m.ToolCalls)
	}
	if m.AgentRuns != 5 {
		t.Errorf("AgentRuns = %d; want 5", m.AgentRuns)
	}
	if m.AgentIterTotal != 25 {
		t.Errorf("AgentIterTotal = %d; want 25", m.AgentIterTotal)
	}

	// Verify model breakdown was loaded
	gpt4, ok := m.LLMByModel["gpt-4"]
	if !ok {
		t.Fatal("LLMByModel[\"gpt-4\"] not found after loading from storage")
	}
	if gpt4.Calls != 10 {
		t.Errorf("gpt-4 Calls = %d; want 10", gpt4.Calls)
	}

	// Verify tool breakdown was loaded
	rf, ok := m.ToolByName["read_file"]
	if !ok {
		t.Fatal("ToolByName[\"read_file\"] not found after loading from storage")
	}
	if rf.Calls != 20 {
		t.Errorf("read_file Calls = %d; want 20", rf.Calls)
	}

	// Verify events were loaded
	if len(m.RecentEvents) != 1 {
		t.Errorf("len(RecentEvents) = %d; want 1", len(m.RecentEvents))
	}

	// Now record something new and verify it flushes to storage
	m.RecordLLMCall("gpt-4", 100*time.Millisecond, 10, 20, nil)

	// Give async flush a moment to complete
	time.Sleep(50 * time.Millisecond)

	if !store.saveCountersCalled {
		t.Error("SaveMetricCounters was not called after recording a call (async flush)")
	}
	if !store.appendEventCalled {
		t.Error("AppendMetricEvent was not called after recording a call (async flush)")
	}
}

// ---------------------------------------------------------------------------
// TestGlobal_ReturnsSingleton
// ---------------------------------------------------------------------------

func TestGlobal_ReturnsSingleton(t *testing.T) {
	g := Global()
	if g == nil {
		t.Fatal("Global() returned nil")
	}

	// Calling Global() again should return the same instance
	g2 := Global()
	if g != g2 {
		t.Error("Global() did not return the same instance on consecutive calls")
	}
}
