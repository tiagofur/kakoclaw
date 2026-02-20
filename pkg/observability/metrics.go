package observability

import (
	"encoding/json"
	"sync"
	"time"
)

// MetricsStorage is implemented by storage.Storage and wired in at server startup.
type MetricsStorage interface {
	SaveMetricCounters(counters map[string]int64) error
	LoadMetricCounters() (map[string]int64, error)
	SaveMetricBreakdowns(modelMetrics json.RawMessage, toolMetrics json.RawMessage) error
	LoadMetricBreakdowns() (json.RawMessage, json.RawMessage, error)
	AppendMetricEvent(event interface{}) error
	LoadRecentEvents() ([]json.RawMessage, error)
}

// Metrics collects in-process observability data for LLM calls, tool executions,
// and general system usage. Thread-safe.
type Metrics struct {
	mu sync.RWMutex

	// LLM call metrics
	LLMCalls     int64                    `json:"llm_calls"`
	LLMErrors    int64                    `json:"llm_errors"`
	LLMTotalMs   int64                    `json:"llm_total_ms"`
	LLMTokensIn  int64                    `json:"llm_tokens_in"`
	LLMTokensOut int64                    `json:"llm_tokens_out"`
	LLMByModel   map[string]*ModelMetrics `json:"llm_by_model"`

	// Tool execution metrics
	ToolCalls   int64                   `json:"tool_calls"`
	ToolErrors  int64                   `json:"tool_errors"`
	ToolTotalMs int64                   `json:"tool_total_ms"`
	ToolByName  map[string]*ToolMetrics `json:"tool_by_name"`

	// Agent loop metrics
	AgentRuns      int64 `json:"agent_runs"`
	AgentErrors    int64 `json:"agent_errors"`
	AgentTotalMs   int64 `json:"agent_total_ms"`
	AgentIterTotal int64 `json:"agent_iterations_total"`

	// Recent events (ring buffer of last N events)
	RecentEvents []Event `json:"recent_events"`

	// Uptime
	StartedAt time.Time `json:"started_at"`

	// Optional persistence backend (nil = in-memory only)
	store MetricsStorage
}

type ModelMetrics struct {
	Calls     int64 `json:"calls"`
	Errors    int64 `json:"errors"`
	TotalMs   int64 `json:"total_ms"`
	TokensIn  int64 `json:"tokens_in"`
	TokensOut int64 `json:"tokens_out"`
}

type ToolMetrics struct {
	Calls   int64 `json:"calls"`
	Errors  int64 `json:"errors"`
	TotalMs int64 `json:"total_ms"`
}

type Event struct {
	Type       string    `json:"type"` // "llm_call", "tool_call", "agent_run", "error"
	Model      string    `json:"model,omitempty"`
	Tool       string    `json:"tool,omitempty"`
	DurationMs int64     `json:"duration_ms"`
	Error      string    `json:"error,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

const maxRecentEvents = 100

// Global singleton
var global *Metrics

func init() {
	global = New()
}

// Global returns the global metrics instance.
func Global() *Metrics {
	return global
}

// New creates a fresh Metrics collector.
func New() *Metrics {
	return &Metrics{
		LLMByModel:   make(map[string]*ModelMetrics),
		ToolByName:   make(map[string]*ToolMetrics),
		RecentEvents: make([]Event, 0, maxRecentEvents),
		StartedAt:    time.Now(),
	}
}

// SetStorage injects a persistence backend and pre-loads existing counters from it.
func (m *Metrics) SetStorage(s MetricsStorage) {
	m.mu.Lock()
	m.store = s
	m.mu.Unlock()

	m.loadFromDB(s)
}

// loadFromDB pre-populates the in-memory counters and recent events from persistent storage.
func (m *Metrics) loadFromDB(s MetricsStorage) {
	counters, err := s.LoadMetricCounters()
	if err != nil || len(counters) == 0 {
		return
	}

	rawEvents, _ := s.LoadRecentEvents()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Restore scalar counters
	m.LLMCalls = counters["llm_calls"]
	m.LLMErrors = counters["llm_errors"]
	m.LLMTotalMs = counters["llm_total_ms"]
	m.LLMTokensIn = counters["llm_tokens_in"]
	m.LLMTokensOut = counters["llm_tokens_out"]
	m.ToolCalls = counters["tool_calls"]
	m.ToolErrors = counters["tool_errors"]
	m.ToolTotalMs = counters["tool_total_ms"]
	m.AgentRuns = counters["agent_runs"]
	m.AgentErrors = counters["agent_errors"]
	m.AgentTotalMs = counters["agent_total_ms"]
	m.AgentIterTotal = counters["agent_iterations_total"]

	// Restore recent events
	for _, raw := range rawEvents {
		var evt Event
		if json.Unmarshal(raw, &evt) == nil {
			m.RecentEvents = append(m.RecentEvents, evt)
		}
	}

	// Restore breakdown maps
	rawModels, rawTools, _ := s.LoadMetricBreakdowns()
	if len(rawModels) > 0 {
		var modelMap map[string]*ModelMetrics
		if json.Unmarshal(rawModels, &modelMap) == nil {
			m.LLMByModel = modelMap
		}
	}
	if len(rawTools) > 0 {
		var toolMap map[string]*ToolMetrics
		if json.Unmarshal(rawTools, &toolMap) == nil {
			m.ToolByName = toolMap
		}
	}
}

// flushCounters persists the current in-memory counters to DB (called without lock held).
func (m *Metrics) flushCounters() {
	m.mu.RLock()
	s := m.store
	counters := map[string]int64{
		"llm_calls":              m.LLMCalls,
		"llm_errors":             m.LLMErrors,
		"llm_total_ms":           m.LLMTotalMs,
		"llm_tokens_in":          m.LLMTokensIn,
		"llm_tokens_out":         m.LLMTokensOut,
		"tool_calls":             m.ToolCalls,
		"tool_errors":            m.ToolErrors,
		"tool_total_ms":          m.ToolTotalMs,
		"agent_runs":             m.AgentRuns,
		"agent_errors":           m.AgentErrors,
		"agent_total_ms":         m.AgentTotalMs,
		"agent_iterations_total": m.AgentIterTotal,
	}
	m.mu.RUnlock()

	if s != nil {
		_ = s.SaveMetricCounters(counters)

		// Persist breakdowns
		m.mu.RLock()
		mod, _ := json.Marshal(m.LLMByModel)
		tls, _ := json.Marshal(m.ToolByName)
		m.mu.RUnlock()
		_ = s.SaveMetricBreakdowns(mod, tls)
	}
}

// asyncFlush fires a goroutine to persist counters and a new event.
func (m *Metrics) asyncFlush(evt Event) {
	m.mu.RLock()
	s := m.store
	m.mu.RUnlock()
	if s == nil {
		return
	}
	go func() {
		m.flushCounters()
		_ = s.AppendMetricEvent(evt)
	}()
}

// RecordLLMCall records a single LLM API call.
func (m *Metrics) RecordLLMCall(model string, duration time.Duration, tokensIn, tokensOut int, err error) {
	m.mu.Lock()

	ms := duration.Milliseconds()
	m.LLMCalls++
	m.LLMTotalMs += ms
	m.LLMTokensIn += int64(tokensIn)
	m.LLMTokensOut += int64(tokensOut)

	if err != nil {
		m.LLMErrors++
	}

	mm, ok := m.LLMByModel[model]
	if !ok {
		mm = &ModelMetrics{}
		m.LLMByModel[model] = mm
	}
	mm.Calls++
	mm.TotalMs += ms
	mm.TokensIn += int64(tokensIn)
	mm.TokensOut += int64(tokensOut)
	if err != nil {
		mm.Errors++
	}

	evt := Event{
		Type:       "llm_call",
		Model:      model,
		DurationMs: ms,
		Timestamp:  time.Now(),
	}
	if err != nil {
		evt.Error = err.Error()
	}
	m.addEvent(evt)

	m.mu.Unlock()

	m.asyncFlush(evt)
}

// RecordToolCall records a single tool execution.
func (m *Metrics) RecordToolCall(toolName string, duration time.Duration, err error) {
	m.mu.Lock()

	ms := duration.Milliseconds()
	m.ToolCalls++
	m.ToolTotalMs += ms

	if err != nil {
		m.ToolErrors++
	}

	tm, ok := m.ToolByName[toolName]
	if !ok {
		tm = &ToolMetrics{}
		m.ToolByName[toolName] = tm
	}
	tm.Calls++
	tm.TotalMs += ms
	if err != nil {
		tm.Errors++
	}

	evt := Event{
		Type:       "tool_call",
		Tool:       toolName,
		DurationMs: ms,
		Timestamp:  time.Now(),
	}
	if err != nil {
		evt.Error = err.Error()
	}
	m.addEvent(evt)

	m.mu.Unlock()

	m.asyncFlush(evt)
}

// RecordAgentRun records a full agent loop execution.
func (m *Metrics) RecordAgentRun(duration time.Duration, iterations int, err error) {
	m.mu.Lock()

	ms := duration.Milliseconds()
	m.AgentRuns++
	m.AgentTotalMs += ms
	m.AgentIterTotal += int64(iterations)

	if err != nil {
		m.AgentErrors++
	}

	evt := Event{
		Type:       "agent_run",
		DurationMs: ms,
		Timestamp:  time.Now(),
	}
	if err != nil {
		evt.Error = err.Error()
	}
	m.addEvent(evt)

	m.mu.Unlock()

	m.asyncFlush(evt)
}

// Snapshot returns a read-only copy of the current metrics.
func (m *Metrics) Snapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Copy model metrics
	models := make(map[string]interface{})
	for k, v := range m.LLMByModel {
		models[k] = map[string]interface{}{
			"calls":      v.Calls,
			"errors":     v.Errors,
			"total_ms":   v.TotalMs,
			"avg_ms":     avgMs(v.TotalMs, v.Calls),
			"tokens_in":  v.TokensIn,
			"tokens_out": v.TokensOut,
		}
	}

	// Copy tool metrics
	toolsMap := make(map[string]interface{})
	for k, v := range m.ToolByName {
		toolsMap[k] = map[string]interface{}{
			"calls":    v.Calls,
			"errors":   v.Errors,
			"total_ms": v.TotalMs,
			"avg_ms":   avgMs(v.TotalMs, v.Calls),
		}
	}

	// Copy recent events
	events := make([]Event, len(m.RecentEvents))
	copy(events, m.RecentEvents)

	return map[string]interface{}{
		"uptime_seconds":         int64(time.Since(m.StartedAt).Seconds()),
		"started_at":             m.StartedAt.Format(time.RFC3339),
		"llm_calls":              m.LLMCalls,
		"llm_errors":             m.LLMErrors,
		"llm_total_ms":           m.LLMTotalMs,
		"llm_avg_ms":             avgMs(m.LLMTotalMs, m.LLMCalls),
		"llm_tokens_in":          m.LLMTokensIn,
		"llm_tokens_out":         m.LLMTokensOut,
		"llm_by_model":           models,
		"tool_calls":             m.ToolCalls,
		"tool_errors":            m.ToolErrors,
		"tool_total_ms":          m.ToolTotalMs,
		"tool_avg_ms":            avgMs(m.ToolTotalMs, m.ToolCalls),
		"tool_by_name":           toolsMap,
		"agent_runs":             m.AgentRuns,
		"agent_errors":           m.AgentErrors,
		"agent_total_ms":         m.AgentTotalMs,
		"agent_avg_ms":           avgMs(m.AgentTotalMs, m.AgentRuns),
		"agent_iterations_total": m.AgentIterTotal,
		"agent_avg_iterations":   avgFloat(m.AgentIterTotal, m.AgentRuns),
		"recent_events":          events,
	}
}

func (m *Metrics) addEvent(evt Event) {
	if len(m.RecentEvents) >= maxRecentEvents {
		m.RecentEvents = m.RecentEvents[1:]
	}
	m.RecentEvents = append(m.RecentEvents, evt)
}

func avgMs(totalMs, count int64) int64 {
	if count == 0 {
		return 0
	}
	return totalMs / count
}

func avgFloat(total, count int64) float64 {
	if count == 0 {
		return 0
	}
	return float64(total) / float64(count)
}
