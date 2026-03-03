# Multi-Agent Orchestration: Event-Driven Pipeline Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix multi-agent response flow (specialist → orchestrator), add delegation chain tracking with depth, enrich WebSocket protocol, and update frontend UI for full agent visibility.

**Architecture:** Event-Driven Pipeline — propagate specialist callbacks to parent context with chain metadata, return structured DelegationResult from specialist execution, add re-delegation limits, and enrich all existing UI components with chain/depth/progress info.

**Tech Stack:** Go 1.26 (backend), Vue 3 + Pinia (frontend), WebSocket JSON protocol

---

## Task 1: Enrich AgentStatusEvent with Delegation Chain Fields

**Files:**
- Modify: `pkg/agent/loop.go:122-129`

**Step 1: Add delegation chain fields to AgentStatusEvent**

In `pkg/agent/loop.go`, update the `AgentStatusEvent` struct (line 122) to add chain tracking fields:

```go
// AgentStatusEvent represents agent status changes during execution
type AgentStatusEvent struct {
	Agent           string    `json:"agent"`
	Status          string    `json:"status"` // "analyzing", "delegating", "working", "complete", "synthesizing", "timeout"
	SpecialistName  string    `json:"specialist_name,omitempty"`
	Reason          string    `json:"reason,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
	DelegationChain []string  `json:"delegation_chain,omitempty"` // e.g. ["orchestrator", "developer", "security"]
	DelegationDepth int       `json:"delegation_depth,omitempty"` // 0=orchestrator, 1=specialist, 2=colleague
	ParentAgent     string    `json:"parent_agent,omitempty"`     // who delegated to this agent
}
```

**Step 2: Add delegation chain fields to SpecialistReport**

In `pkg/agent/orchestrator.go`, update `SpecialistReport` struct (line 127) — add after `Iteration` field (line 140):

```go
	DelegationChain []string `json:"delegation_chain,omitempty"` // delegation path
	DelegationDepth int      `json:"delegation_depth,omitempty"` // depth in chain
	ToolsUsed       []string `json:"tools_used,omitempty"`       // tools the specialist called
	IterationsUsed  int      `json:"iterations_used,omitempty"`  // LLM iterations consumed
```

**Step 3: Add DelegationUpdate event type and callback**

In `pkg/agent/loop.go`, after `ContentSegmentCallback` (line 143), add:

```go
// DelegationUpdate represents real-time progress of an active delegation
type DelegationUpdate struct {
	DelegationID  string    `json:"delegation_id"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	Status        string    `json:"status"` // "started", "in_progress", "complete", "error"
	Iteration     int       `json:"iteration"`
	MaxIterations int       `json:"max_iterations"`
	ElapsedMs     int64     `json:"elapsed_ms"`
	Timestamp     time.Time `json:"timestamp"`
}

// DelegationUpdateCallback is called for delegation progress updates
type DelegationUpdateCallback func(update DelegationUpdate) error
```

**Step 4: Add context helpers for DelegationUpdateCallback**

In `pkg/agent/orchestrator.go`, after the specialist report context helpers (line 98), add:

```go
type delegationUpdateCallbackKey struct{}

func ContextWithDelegationUpdateCallback(ctx context.Context, callback DelegationUpdateCallback) context.Context {
	return context.WithValue(ctx, delegationUpdateCallbackKey{}, callback)
}

func delegationUpdateCallbackFromCtx(ctx context.Context) DelegationUpdateCallback {
	if v, ok := ctx.Value(delegationUpdateCallbackKey{}).(DelegationUpdateCallback); ok {
		return v
	}
	return nil
}

func emitDelegationUpdate(ctx context.Context, update DelegationUpdate) {
	if callback := delegationUpdateCallbackFromCtx(ctx); callback != nil {
		_ = callback(update)
	}
}
```

**Step 5: Verify compilation**

Run: `cd /c/Users/tfurt/source/repos/kakoclaw && go build ./pkg/agent/...`
Expected: Build succeeds with no errors.

**Step 6: Commit**

```bash
git add pkg/agent/loop.go pkg/agent/orchestrator.go
git commit -m "feat(agent): add delegation chain fields to AgentStatusEvent and SpecialistReport"
```

---

## Task 2: Propagate Callbacks from Orchestrator to Specialist Context

**Files:**
- Modify: `pkg/agent/orchestrator.go:666-764` (`processSpecialistTask`)

**Step 1: Write test for callback propagation**

Create `pkg/agent/orchestrator_callback_test.go`:

```go
package agent

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestCallbackPropagationToSpecialist(t *testing.T) {
	// Verify that AgentStatusCallback from parent context is invoked
	// when specialist emits status events
	var mu sync.Mutex
	var events []AgentStatusEvent

	ctx := context.Background()
	ctx = ContextWithAgentStatusCallback(ctx, func(ev AgentStatusEvent) error {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, ev)
		return nil
	})

	// Emit a test event
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:           "developer",
		Status:          "working",
		DelegationChain: []string{"orchestrator", "developer"},
		DelegationDepth: 1,
		ParentAgent:     "orchestrator",
		Timestamp:       time.Now(),
	})

	mu.Lock()
	defer mu.Unlock()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].DelegationDepth != 1 {
		t.Errorf("expected depth 1, got %d", events[0].DelegationDepth)
	}
	if len(events[0].DelegationChain) != 2 {
		t.Errorf("expected chain length 2, got %d", len(events[0].DelegationChain))
	}
	if events[0].ParentAgent != "orchestrator" {
		t.Errorf("expected parent 'orchestrator', got '%s'", events[0].ParentAgent)
	}
}

func TestDelegationUpdateEmission(t *testing.T) {
	var received DelegationUpdate
	ctx := context.Background()
	ctx = ContextWithDelegationUpdateCallback(ctx, func(update DelegationUpdate) error {
		received = update
		return nil
	})

	emitDelegationUpdate(ctx, DelegationUpdate{
		DelegationID:  "del_test",
		From:          "orchestrator",
		To:            "developer",
		Status:        "in_progress",
		Iteration:     3,
		MaxIterations: 20,
		ElapsedMs:     5200,
		Timestamp:     time.Now(),
	})

	if received.DelegationID != "del_test" {
		t.Errorf("expected delegation_id 'del_test', got '%s'", received.DelegationID)
	}
	if received.Iteration != 3 {
		t.Errorf("expected iteration 3, got %d", received.Iteration)
	}
}
```

**Step 2: Run test to verify it passes**

Run: `cd /c/Users/tfurt/source/repos/kakoclaw && go test ./pkg/agent/ -run TestCallbackPropagation -v && go test ./pkg/agent/ -run TestDelegationUpdateEmission -v`
Expected: PASS

**Step 3: Update processSpecialistTask to propagate callbacks and emit chain events**

Replace the `processSpecialistTask` method in `pkg/agent/orchestrator.go` (lines 666-764). The key changes:
1. Extract parent callbacks and wrap them with chain metadata
2. Emit `delegation_start` event with chain info before specialist runs
3. Emit `delegation_update` during specialist execution (via the specialist's iteration callback in context)
4. Enrich the `SpecialistReport` with chain/depth/tools info
5. Emit `delegation_end` after specialist completes

Replace `processSpecialistTask` (line 666 to line 764) with:

```go
// processSpecialistTask executes a task through a specialist agent
func (oa *OrchestratorAgent) processSpecialistTask(ctx context.Context, specialistName, task string) (string, error) {
	specialist, err := oa.registry.GetSpecialist(specialistName)
	if err != nil {
		return "", err
	}

	// Build delegation chain from parent context
	parentChain := delegationChainFromCtx(ctx)
	currentChain := append(parentChain, specialistName)
	currentDepth := len(parentChain) // orchestrator=0, first specialist=1, colleague=2
	parentAgent := "orchestrator"
	if len(parentChain) > 0 {
		parentAgent = parentChain[len(parentChain)-1]
	}

	// Generate a unique delegation ID
	delegationID := fmt.Sprintf("del_%s_%d", specialistName, time.Now().UnixNano())
	startTime := time.Now()

	// Store chain in context for nested delegations
	specialistCtx := contextWithDelegationChain(ctx, currentChain)

	// Inject response format instruction for structured feedback
	taskWithFormat := task + `

--- RESPONSE FORMAT ---
Start your response with a JSON header on a single line:
{"status":"complete","confidence":0.9,"request_help":"","suggestion":""}

Status values: "complete" (task done well), "partial" (task partly done), "needs_help" (need another specialist)
Confidence: 0.0-1.0 (how confident are you in the result)
request_help: specialist name if you need help (e.g., "security", "documentation")
suggestion: brief note about what you did or need

Then provide your full response below the JSON line.
---`

	// Timeout configurable (default: 5 minutes)
	timeout := 5 * time.Minute
	ctxWithTimeout, cancel := context.WithTimeout(specialistCtx, timeout)
	defer cancel()

	// Emit delegation start with chain info
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:           specialistName,
		Status:          "working",
		DelegationChain: currentChain,
		DelegationDepth: currentDepth,
		ParentAgent:     parentAgent,
		Timestamp:       time.Now(),
	})

	emitDelegationUpdate(ctx, DelegationUpdate{
		DelegationID:  delegationID,
		From:          parentAgent,
		To:            specialistName,
		Status:        "started",
		Iteration:     0,
		MaxIterations: oa.maxIterations,
		ElapsedMs:     0,
		Timestamp:     time.Now(),
	})

	// Execute in goroutine to detect timeout
	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := specialist.ProcessWithSpeciality(ctxWithTimeout, taskWithFormat)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for result or timeout
	var result string
	select {
	case result = <-resultChan:
		elapsed := time.Since(startTime).Milliseconds()

		// Parse structured report from specialist response
		report := oa.parseSpecialistReport(specialistName, result)
		// Enrich report with chain metadata
		report.DelegationChain = currentChain
		report.DelegationDepth = currentDepth
		oa.storeLastReport(report)

		// Emit specialist report event for frontend visibility
		emitSpecialistReport(ctx, report)

		// Log report details
		logger.InfoCF("agent", "Specialist report received", map[string]interface{}{
			"specialist": specialistName,
			"status":     report.Status,
			"confidence": report.Confidence,
			"help":       report.RequestHelp,
			"chain":      currentChain,
			"depth":      currentDepth,
		})

		// Clean result (remove JSON header) for content segment
		cleanResult := oa.cleanSpecialistResult(result)

		// Emit content segment with clean result
		emitContentSegment(ctx, ContentSegment{
			Agent:     specialistName,
			Content:   cleanResult,
			SegmentID: fmt.Sprintf("seg_%s_%d", specialistName, time.Now().UnixNano()),
			Timestamp: time.Now(),
		})

		// Emit delegation complete
		emitDelegationUpdate(ctx, DelegationUpdate{
			DelegationID:  delegationID,
			From:          parentAgent,
			To:            specialistName,
			Status:        "complete",
			Iteration:     report.Iteration,
			MaxIterations: oa.maxIterations,
			ElapsedMs:     elapsed,
			Timestamp:     time.Now(),
		})

		// Handle empty/insufficient response
		if len(strings.TrimSpace(cleanResult)) < 10 {
			logger.WarnCF("agent", "Specialist returned insufficient response", map[string]interface{}{
				"specialist": specialistName,
				"length":     len(cleanResult),
			})
			// Return structured error so orchestrator LLM can decide what to do
			cleanResult = fmt.Sprintf("[Specialist %s returned an insufficient response. Confidence: %.0f%%. Consider delegating to another specialist or providing more context.]",
				specialistName, report.Confidence*100)
		}

		// Update result to clean version for return
		result = cleanResult

	case err := <-errChan:
		emitDelegationUpdate(ctx, DelegationUpdate{
			DelegationID:  delegationID,
			From:          parentAgent,
			To:            specialistName,
			Status:        "error",
			ElapsedMs:     time.Since(startTime).Milliseconds(),
			Timestamp:     time.Now(),
		})
		return "", fmt.Errorf("specialist processing error: %w", err)

	case <-ctxWithTimeout.Done():
		// Emit timeout status with chain info
		emitAgentStatus(ctx, AgentStatusEvent{
			Agent:           specialistName,
			Status:          "timeout",
			DelegationChain: currentChain,
			DelegationDepth: currentDepth,
			ParentAgent:     parentAgent,
			Timestamp:       time.Now(),
		})
		emitDelegationUpdate(ctx, DelegationUpdate{
			DelegationID:  delegationID,
			From:          parentAgent,
			To:            specialistName,
			Status:        "error",
			ElapsedMs:     time.Since(startTime).Milliseconds(),
			Timestamp:     time.Now(),
		})
		return "", fmt.Errorf("specialist %s timed out after %v", specialistName, timeout)
	}

	// Use tracker from context if available
	tracker := agentTrackerFromCtx(ctx)
	if tracker == nil {
		tracker = oa.AgentLoop
	}
	if len(tracker.GetInvolvedAgents()) == 0 {
		tracker.AddInvolvedAgent("orchestrator")
	}
	tracker.AddInvolvedAgent(specialistName)

	return result, nil
}
```

**Step 4: Add delegation chain context helpers**

In `pkg/agent/orchestrator.go`, after the delegation update context helpers, add:

```go
type delegationChainKey struct{}

func contextWithDelegationChain(ctx context.Context, chain []string) context.Context {
	return context.WithValue(ctx, delegationChainKey{}, chain)
}

func delegationChainFromCtx(ctx context.Context) []string {
	if v, ok := ctx.Value(delegationChainKey{}).([]string); ok {
		result := make([]string, len(v))
		copy(result, v)
		return result
	}
	return []string{"orchestrator"}
}
```

**Step 5: Update DelegationTool.Execute to include chain in its status emissions**

In `pkg/agent/orchestrator.go`, update `DelegationTool.Execute()` (lines 328-427). The status events emitted at lines 377-403 should include chain info. Update the "analyzing" event at line 377:

```go
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:           "orchestrator",
		Status:          "analyzing",
		DelegationChain: delegationChainFromCtx(ctx),
		DelegationDepth: 0,
		Timestamp:       time.Now(),
	})
```

Update the "delegating" event at line 384:

```go
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:           "orchestrator",
		Status:          "delegating",
		SpecialistName:  specialistName,
		Reason:          generateDelegationReason(specialistName, task),
		DelegationChain: delegationChainFromCtx(ctx),
		DelegationDepth: 0,
		Timestamp:       time.Now(),
	})
```

Remove the "working" event at lines 398-403 (it's now emitted inside `processSpecialistTask` with chain info).

Update the "complete" event at line 414:

```go
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:           specialistName,
		Status:          "complete",
		DelegationChain: append(delegationChainFromCtx(ctx), specialistName),
		DelegationDepth: 1,
		ParentAgent:     "orchestrator",
		Timestamp:       time.Now(),
	})
```

**Step 6: Update RequestColleagueTool.Execute to include chain in events**

In `pkg/agent/specialist.go`, update the `requesting_help` event at line 363:

```go
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:          t.currentAgent.name,
		Status:         "requesting_help",
		SpecialistName: colleagueName,
		Reason:         fmt.Sprintf("Consulting %s: %s", colleagueName, truncateString(question, 100)),
		DelegationChain: delegationChainFromCtx(ctx),
		DelegationDepth: t.currentDepth + 1,
		ParentAgent:     t.currentAgent.name,
		Timestamp:       time.Now(),
	})
```

And the `colleague_complete` event at line 410:

```go
	emitAgentStatus(ctx, AgentStatusEvent{
		Agent:           colleagueName,
		Status:          "colleague_complete",
		DelegationChain: append(delegationChainFromCtx(ctx), colleagueName),
		DelegationDepth: t.currentDepth + 1,
		ParentAgent:     t.currentAgent.name,
		Timestamp:       time.Now(),
	})
```

**Step 7: Verify compilation and run tests**

Run: `cd /c/Users/tfurt/source/repos/kakoclaw && go build ./pkg/agent/... && go test ./pkg/agent/ -run TestCallback -v`
Expected: Build succeeds, tests pass.

**Step 8: Commit**

```bash
git add pkg/agent/orchestrator.go pkg/agent/specialist.go pkg/agent/orchestrator_callback_test.go
git commit -m "feat(agent): propagate callbacks to specialist context with delegation chain tracking"
```

---

## Task 3: Add Re-delegation Limit to Orchestrator

**Files:**
- Modify: `pkg/agent/orchestrator.go:100-110` (OrchestratorAgent struct)
- Modify: `pkg/agent/orchestrator.go:328-427` (DelegationTool.Execute)

**Step 1: Add delegation counter to OrchestratorAgent**

In `pkg/agent/orchestrator.go`, add to the `OrchestratorAgent` struct (line 100):

```go
type OrchestratorAgent struct {
	*SpecialistAgent
	registry          *SpecialistRegistry
	delegationRetries int
	fallbackToDefault bool

	// Specialist report tracking for feedback loop
	lastReport   *SpecialistReport
	lastReportMu sync.RWMutex

	// Delegation limit tracking (per message)
	delegationCount   int
	delegationCountMu sync.Mutex
	maxDelegations    int // default: 3
}
```

**Step 2: Add delegation limit check in DelegationTool.Execute**

At the beginning of `DelegationTool.Execute()` (line 328), after parsing args, add:

```go
	// Check delegation limit
	dt.orchestrator.delegationCountMu.Lock()
	dt.orchestrator.delegationCount++
	count := dt.orchestrator.delegationCount
	dt.orchestrator.delegationCountMu.Unlock()

	maxDel := dt.orchestrator.maxDelegations
	if maxDel <= 0 {
		maxDel = 3
	}

	if count > maxDel {
		emitAgentStatus(ctx, AgentStatusEvent{
			Agent:     "orchestrator",
			Status:    "max_delegations_reached",
			Reason:    fmt.Sprintf("Reached maximum %d delegations for this message. Synthesizing response with available results.", maxDel),
			Timestamp: time.Now(),
		})
		return fmt.Sprintf("[Maximum delegation limit (%d) reached. Please synthesize a response using the information gathered so far.]", maxDel), nil
	}
```

**Step 3: Reset delegation count on new message**

In `pkg/agent/loop.go`, in the `ClearInvolvedAgents()` method (line 88), this is already called before each new message. We need to also clear delegation count.

Add a method to `OrchestratorAgent` to reset per-message state, and call it in the orchestrator's `ProcessWithSpeciality` or wherever the main loop starts processing a new user message. The simplest approach: reset in `DelegationTool.Execute` when the context has a fresh message marker, OR expose a `ResetDelegationCount()` method and call it from the web handler.

Add to `orchestrator.go`:

```go
func (oa *OrchestratorAgent) ResetDelegationCount() {
	oa.delegationCountMu.Lock()
	defer oa.delegationCountMu.Unlock()
	oa.delegationCount = 0
}
```

In `pkg/web/server.go`, before calling `ProcessDirectWithUserAndModelStream` (line 1323), add:

```go
			// Reset per-message state for orchestrator
			if orch, ok := activeAgentLoop.(*OrchestratorAgent); ok {
				orch.ResetDelegationCount()
			}
```

Note: The `activeAgentLoop` may be wrapped — check how it's accessed. If it's accessed via the `AgentManager`, you may need to expose this via the manager. Alternatively, store the reset in context and check in DelegationTool.

**Step 4: Verify compilation**

Run: `cd /c/Users/tfurt/source/repos/kakoclaw && go build ./pkg/...`
Expected: Build succeeds.

**Step 5: Commit**

```bash
git add pkg/agent/orchestrator.go pkg/web/server.go
git commit -m "feat(agent): add re-delegation limit (max 3 per message) to prevent infinite loops"
```

---

## Task 4: Enrich WebSocket Protocol in Server Handler

**Files:**
- Modify: `pkg/web/server.go:1279-1375`

**Step 1: Update agent_status WebSocket emission to include chain fields**

In `pkg/web/server.go`, update the agent status callback (lines 1280-1291) to include new fields:

```go
			ctx = agent.ContextWithAgentStatusCallback(ctx, func(ev agent.AgentStatusEvent) error {
				wsMu.Lock()
				defer wsMu.Unlock()
				msg := map[string]interface{}{
					"type":            "agent_status",
					"agent":           ev.Agent,
					"status":          ev.Status,
					"specialist_name": ev.SpecialistName,
					"reason":          ev.Reason,
					"timestamp":       ev.Timestamp.Format(time.RFC3339),
				}
				if len(ev.DelegationChain) > 0 {
					msg["delegation_chain"] = ev.DelegationChain
				}
				if ev.DelegationDepth > 0 {
					msg["delegation_depth"] = ev.DelegationDepth
				}
				if ev.ParentAgent != "" {
					msg["parent_agent"] = ev.ParentAgent
				}
				return conn.WriteJSON(msg)
			})
```

**Step 2: Update specialist_report emission to include chain/tools/iterations**

Update the specialist report callback (lines 1306-1321):

```go
			ctx = agent.ContextWithSpecialistReportCallback(ctx, func(report *agent.SpecialistReport) error {
				wsMu.Lock()
				defer wsMu.Unlock()
				msg := map[string]interface{}{
					"type":            "specialist_report",
					"specialist_name": report.SpecialistName,
					"status":          report.Status,
					"confidence":      report.Confidence,
					"request_help":    report.RequestHelp,
					"help_context":    report.HelpContext,
					"suggestions":     report.Suggestions,
					"needs_review":    report.NeedsReview,
					"timestamp":       report.Timestamp.Format(time.RFC3339),
				}
				if len(report.DelegationChain) > 0 {
					msg["delegation_chain"] = report.DelegationChain
				}
				if report.DelegationDepth > 0 {
					msg["delegation_depth"] = report.DelegationDepth
				}
				if len(report.ToolsUsed) > 0 {
					msg["tools_used"] = report.ToolsUsed
				}
				if report.IterationsUsed > 0 {
					msg["iterations_used"] = report.IterationsUsed
				}
				return conn.WriteJSON(msg)
			})
```

**Step 3: Add delegation_update callback to context**

After the specialist report callback, add:

```go
			ctx = agent.ContextWithDelegationUpdateCallback(ctx, func(update agent.DelegationUpdate) error {
				wsMu.Lock()
				defer wsMu.Unlock()
				return conn.WriteJSON(map[string]interface{}{
					"type":           "delegation_update",
					"delegation_id":  update.DelegationID,
					"from":           update.From,
					"to":             update.To,
					"status":         update.Status,
					"iteration":      update.Iteration,
					"max_iterations": update.MaxIterations,
					"elapsed_ms":     update.ElapsedMs,
					"timestamp":      update.Timestamp.Format(time.RFC3339),
				})
			})
```

**Step 4: Enrich stream_end with delegation_summary**

Update the stream_end emission (lines 1363-1375). We need to collect delegation data during the request. Add a slice to track delegations before the streaming call, and populate it via the callbacks.

Before the `ContextWithAgentStatusCallback` block (around line 1279), add:

```go
			var delegationsMu sync.Mutex
			var delegationSummary []map[string]interface{}
```

In the specialist report callback, also record to summary:

```go
				// Record for stream_end summary
				delegationsMu.Lock()
				delegationSummary = append(delegationSummary, map[string]interface{}{
					"from":       report.DelegationChain[0],
					"to":         report.SpecialistName,
					"confidence": report.Confidence,
					"tools_used": report.ToolsUsed,
				})
				delegationsMu.Unlock()
```

Update stream_end (line 1366):

```go
			agents := activeAgentLoop.GetInvolvedAgents()
			wsMu.Lock()
			streamEndMsg := map[string]interface{}{
				"type":    "stream_end",
				"content": response,
			}
			if len(agents) > 0 {
				streamEndMsg["agents"] = agents
			}
			delegationsMu.Lock()
			if len(delegationSummary) > 0 {
				streamEndMsg["delegation_summary"] = delegationSummary
			}
			delegationsMu.Unlock()
			_ = conn.WriteJSON(streamEndMsg)
			_ = conn.WriteJSON(map[string]interface{}{"type": "ready"})
			wsMu.Unlock()
```

**Step 5: Verify compilation**

Run: `cd /c/Users/tfurt/source/repos/kakoclaw && go build ./pkg/web/...`
Expected: Build succeeds.

**Step 6: Commit**

```bash
git add pkg/web/server.go
git commit -m "feat(web): enrich WebSocket protocol with delegation chain, depth, and delegation_update events"
```

---

## Task 5: Update chatStore with Delegation Chain State

**Files:**
- Modify: `pkg/web/frontend/src/stores/chatStore.js:22-31` (state), `131-179` (actions)

**Step 1: Add delegation chain state**

In `chatStore.js`, after `lastSpecialistReport` (line 31), add:

```javascript
  const delegationChain = ref([])        // Active chain: [{agent, status, depth, startedAt, parentAgent}]
  const activeDelegation = ref(null)     // Currently executing delegation progress
  const delegationHistory = ref([])      // Completed delegations for current message
```

**Step 2: Add updateDelegationChain action**

After `addSpecialistReport` function (line 179), add:

```javascript
  // Update delegation chain from enriched agent_status events
  function updateDelegationChain(event) {
    const { agent, status, delegation_chain, delegation_depth, parent_agent } = event

    if (delegation_chain && delegation_chain.length > 0) {
      delegationChain.value = delegation_chain.map((a, i) => ({
        agent: a,
        status: i === delegation_chain.length - 1 ? status : 'delegated',
        depth: i,
        startedAt: new Date().toISOString(),
        parentAgent: i > 0 ? delegation_chain[i - 1] : null
      }))
    }

    // Update chain entry for this agent's status
    const existing = delegationChain.value.find(e => e.agent === agent)
    if (existing) {
      existing.status = status
    }
  }

  // Handle delegation_update events (progress tracking)
  function updateDelegationProgress(update) {
    activeDelegation.value = {
      delegationId: update.delegation_id,
      from: update.from,
      to: update.to,
      status: update.status,
      iteration: update.iteration,
      maxIterations: update.max_iterations,
      elapsedMs: update.elapsed_ms,
      timestamp: update.timestamp
    }

    // If complete or error, move to history and clear active
    if (update.status === 'complete' || update.status === 'error') {
      delegationHistory.value.push({ ...activeDelegation.value })
      activeDelegation.value = null
    }
  }

  // Build delegation summary from stream_end data
  function setDelegationSummary(summary) {
    if (summary && summary.length > 0) {
      delegationHistory.value = summary.map(d => ({
        from: d.from,
        to: d.to,
        confidence: d.confidence,
        toolsUsed: d.tools_used || []
      }))
    }
  }
```

**Step 3: Update clearAgentStatus to also clear delegation state**

Update `clearAgentStatus` (line 153):

```javascript
  function clearAgentStatus() {
    orchestratorStatus.value = 'idle'
    currentAgent.value = null
    activeSpecialist.value = null
    delegationReason.value = ''
    agentHistory.value = []
    specialistReports.value = []
    lastSpecialistReport.value = null
    delegationChain.value = []
    activeDelegation.value = null
    delegationHistory.value = []
  }
```

**Step 4: Update endStreamingMessage to snapshot delegation data**

Update `endStreamingMessage` (line 67). In the agentActivity snapshot (line 82), add delegation data:

```javascript
        if (agentHistory.value.length > 0 || specialistReports.value.length > 0 || delegationHistory.value.length > 0) {
          msg.agentActivity = {
            hadMultiAgent: agentHistory.value.length > 1,
            history: [...agentHistory.value],
            reports: [...specialistReports.value],
            delegations: [...delegationHistory.value]
          }
        }
```

**Step 5: Export new state and actions**

Add to the return statement of the store:

```javascript
    delegationChain,
    activeDelegation,
    delegationHistory,
    updateDelegationChain,
    updateDelegationProgress,
    setDelegationSummary,
```

**Step 6: Commit**

```bash
git add pkg/web/frontend/src/stores/chatStore.js
git commit -m "feat(frontend): add delegation chain state and actions to chatStore"
```

---

## Task 6: Update ChatView WebSocket Handler for New Events

**Files:**
- Modify: `pkg/web/frontend/src/views/ChatView.vue:1207-1320` (`handleMessage`)

**Step 1: Update agent_status handler to use chain data**

In `handleMessage()`, replace the `agent_status` block (lines 1238-1281) with:

```javascript
  if (message.type === 'agent_status') {
    chatStore.setAgentStatus(
      message.agent,
      message.status,
      message.specialist_name,
      message.reason
    )
    // Update delegation chain if chain data present
    if (message.delegation_chain) {
      chatStore.updateDelegationChain(message)
    }
    // Insert inline agent event bubble with chain info
    chatStore.addAgentEvent({
      agent: message.agent,
      status: message.status,
      specialistName: message.specialist_name,
      reason: message.reason,
      delegationChain: message.delegation_chain || [],
      delegationDepth: message.delegation_depth || 0,
      parentAgent: message.parent_agent || '',
      timestamp: new Date().toISOString()
    })
    // Update team activity panel
    teamAgentStatus.value = {
      agent: message.agent,
      status: message.status,
      specialistName: message.specialist_name,
      reason: message.reason,
      delegationChain: message.delegation_chain || [],
      delegationDepth: message.delegation_depth || 0,
      parentAgent: message.parent_agent || ''
    }
    // Track involved agents
    if (message.agent && !involvedAgentsList.value.includes(message.agent)) {
      involvedAgentsList.value.push(message.agent)
    }
    // Track inter-specialist communications
    if (message.status === 'requesting_help' && message.specialist_name) {
      teamCommunications.value.push({
        from: message.agent,
        to: message.specialist_name,
        message: message.reason || 'Requesting assistance',
        timestamp: new Date().toISOString()
      })
    }
    if (message.status === 'colleague_complete') {
      teamCommunications.value.push({
        from: message.agent,
        to: message.parent_agent || 'orchestrator',
        message: 'Completed assistance',
        timestamp: new Date().toISOString()
      })
    }
  }
```

**Step 2: Add delegation_update handler**

After the `tool_call` handler (line 1286), add:

```javascript
  if (message.type === 'delegation_update') {
    chatStore.updateDelegationProgress(message)
  }
```

**Step 3: Update stream_end handler to capture delegation_summary**

In the `stream_end` handler (line 1229), after `endStreamingMessage`, add:

```javascript
  if (message.type === 'stream_end') {
    if (message.error) {
      chatStore.endStreamingMessage(`Error: ${message.error}`, [])
    } else {
      chatStore.endStreamingMessage(message.content || '', message.agents || [])
    }
    // Capture delegation summary if present
    if (message.delegation_summary) {
      chatStore.setDelegationSummary(message.delegation_summary)
    }
    chatStore.clearAgentStatus()
    fetchSessions()
  }
```

**Step 4: Update stream_start handler to reset delegation state**

In the `stream_start` handler (line 1218), the `clearAgentStatus()` call already resets delegation state (from Task 5 changes). No additional change needed.

**Step 5: Commit**

```bash
git add pkg/web/frontend/src/views/ChatView.vue
git commit -m "feat(frontend): handle delegation_update events and chain data in ChatView"
```

---

## Task 7: Update AgentStatusIndicator with Delegation Chain Display

**Files:**
- Modify: `pkg/web/frontend/src/components/Chat/AgentStatusIndicator.vue`

**Step 1: Rewrite AgentStatusIndicator to show delegation chain**

Replace the entire template section (lines 1-57) with:

```vue
<template>
  <transition name="slide-fade">
    <div
      v-if="isActive"
      class="glass-panel rounded-lg px-4 py-3 mb-4 border-l-4"
      :class="borderColorClass"
    >
      <div class="flex items-center gap-3">
        <!-- Spinner -->
        <div class="relative w-8 h-8 flex-shrink-0">
          <svg
            class="animate-spin"
            :class="iconColorClass"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
        </div>

        <!-- Chain + Status -->
        <div class="flex-1 min-w-0">
          <!-- Delegation Chain -->
          <div class="flex items-center gap-1.5 flex-wrap">
            <template v-for="(agent, i) in displayChain" :key="i">
              <SpecialistBadge :name="agent.agent" class="text-xs" />
              <svg v-if="i < displayChain.length - 1" class="w-3 h-3 text-makoclaw-text-secondary flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </template>
            <span class="text-sm font-medium" :class="textColorClass">
              {{ leafStatusText }}
            </span>
          </div>

          <!-- Delegation reason -->
          <p v-if="delegationReason" class="text-xs text-makoclaw-text-secondary mt-1 italic truncate">
            {{ delegationReason }}
          </p>

          <!-- Progress bar (iteration tracking) -->
          <div v-if="activeDelegation" class="flex items-center gap-2 mt-1.5">
            <div class="flex-1 h-1 bg-makoclaw-bg/50 rounded-full overflow-hidden">
              <div
                class="h-full bg-makoclaw-accent/60 rounded-full transition-all duration-300"
                :style="{ width: progressPercent + '%' }"
              />
            </div>
            <span class="text-[10px] text-makoclaw-text-secondary whitespace-nowrap">
              {{ activeDelegation.iteration }}/{{ activeDelegation.maxIterations }}
              <span v-if="elapsedSeconds > 0" class="ml-1">{{ elapsedSeconds }}s</span>
            </span>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>
```

**Step 2: Update script section**

Replace the script (lines 60-136) with:

```vue
<script setup>
import { computed } from 'vue'
import { useChatStore } from '../../stores/chatStore'
import SpecialistBadge from './SpecialistBadge.vue'

const chatStore = useChatStore()

const hasOrchestratorStatus = computed(() => {
  return chatStore.orchestratorStatus !== 'idle' && chatStore.orchestratorStatus !== 'complete'
})

const isProcessing = computed(() => {
  return chatStore.globalIsLoading || chatStore.isStreaming
})

const isActive = computed(() => {
  return hasOrchestratorStatus.value || (isProcessing.value && !chatStore.isStreaming)
})

// Build display chain from delegation chain or fallback to current agent
const displayChain = computed(() => {
  if (chatStore.delegationChain.length > 0) {
    return chatStore.delegationChain
  }
  const agent = chatStore.currentAgent || chatStore.activeSpecialist
  if (agent) {
    return [{ agent, status: chatStore.orchestratorStatus, depth: 0 }]
  }
  return []
})

const activeDelegation = computed(() => chatStore.activeDelegation)

const progressPercent = computed(() => {
  if (!activeDelegation.value) return 0
  const { iteration, maxIterations } = activeDelegation.value
  return maxIterations > 0 ? Math.min((iteration / maxIterations) * 100, 100) : 0
})

const elapsedSeconds = computed(() => {
  if (!activeDelegation.value?.elapsedMs) return 0
  return Math.round(activeDelegation.value.elapsedMs / 1000)
})

const delegationReason = computed(() => chatStore.delegationReason)

const leafStatusText = computed(() => {
  const status = chatStore.orchestratorStatus
  const statusMap = {
    analyzing: 'analyzing...',
    delegating: 'delegating...',
    working: 'working...',
    synthesizing: 'synthesizing response...',
    fallback: 'fallback...',
    max_delegations_reached: 'compiling results...'
  }
  return statusMap[status] || 'processing...'
})

const leafStatus = computed(() => {
  return chatStore.orchestratorStatus || 'working'
})

const iconColorClass = computed(() => {
  const colorMap = {
    analyzing: 'text-blue-400',
    delegating: 'text-purple-400',
    working: 'text-green-400',
    synthesizing: 'text-cyan-400',
    fallback: 'text-amber-400'
  }
  return colorMap[leafStatus.value] || 'text-blue-400'
})

const textColorClass = computed(() => iconColorClass.value)

const borderColorClass = computed(() => {
  const colorMap = {
    analyzing: 'border-blue-400',
    delegating: 'border-purple-400',
    working: 'border-green-400',
    synthesizing: 'border-cyan-400',
    fallback: 'border-amber-400'
  }
  return colorMap[leafStatus.value] || 'border-blue-400'
})
</script>
```

Keep the existing `<style scoped>` block unchanged.

**Step 3: Commit**

```bash
git add pkg/web/frontend/src/components/Chat/AgentStatusIndicator.vue
git commit -m "feat(frontend): show delegation chain with progress bar in AgentStatusIndicator"
```

---

## Task 8: Update AgentEventBubble with Chain Context

**Files:**
- Modify: `pkg/web/frontend/src/components/Chat/AgentEventBubble.vue`

**Step 1: Update props and label computed to use chain data**

Update `addAgentEvent` in chatStore.js already passes chain data. Now update the bubble.

In `AgentEventBubble.vue`, update the label computed (line 42) to show confidence for complete events and chain info:

```javascript
const label = computed(() => {
  const { agent, status, specialistName, delegationDepth, parentAgent } = props.event

  const agentLabel = agent || 'agent'
  const targetLabel = specialistName || ''
  const fromLabel = parentAgent || 'orchestrator'

  switch (status) {
    case 'delegating':
      return `${agentLabel} delegated to ${targetLabel}`
    case 'working':
      return delegationDepth > 0
        ? `${agentLabel} is working (via ${fromLabel})...`
        : `${agentLabel} is working...`
    case 'complete':
      return `${agentLabel} finished`
    case 'fallback':
      return `Falling back to ${targetLabel}`
    case 'requesting_help':
      return `${agentLabel} requested help from ${targetLabel}`
    case 'colleague_complete':
      return `${agentLabel} completed assistance for ${fromLabel}`
    case 'synthesizing':
      return `${agentLabel} synthesizing response...`
    case 'max_delegations_reached':
      return `Maximum delegations reached, compiling results...`
    default:
      return `${agentLabel}: ${status}`
  }
})
```

Update the dotColor to handle new statuses:

```javascript
const dotColor = computed(() => {
  const colorMap = {
    delegating: 'bg-purple-400',
    working: 'bg-emerald-400 animate-pulse',
    complete: 'bg-emerald-400',
    fallback: 'bg-amber-400',
    requesting_help: 'bg-amber-400 animate-pulse',
    colleague_complete: 'bg-emerald-400',
    synthesizing: 'bg-cyan-400 animate-pulse',
    max_delegations_reached: 'bg-amber-400',
    timeout: 'bg-red-400'
  }
  return colorMap[props.event.status] || 'bg-makoclaw-text-secondary'
})
```

**Step 2: Commit**

```bash
git add pkg/web/frontend/src/components/Chat/AgentEventBubble.vue
git commit -m "feat(frontend): enrich AgentEventBubble with chain context and new statuses"
```

---

## Task 9: Update TeamActivityPanel with Tree View

**Files:**
- Modify: `pkg/web/frontend/src/components/Chat/TeamActivityPanel.vue`

**Step 1: Update props to accept delegation chain**

Add to props (line 210):

```javascript
  delegationChain: {
    type: Array,
    default: () => []
  },
  activeDelegation: {
    type: Object,
    default: null
  }
```

**Step 2: Replace the "Current Active Agent" section (lines 61-93) with tree view**

Replace lines 61-93 with:

```vue
        <!-- Delegation Chain Tree -->
        <div v-if="chainTree.length > 0" class="space-y-1">
          <div
            v-for="(node, i) in chainTree"
            :key="node.agent"
            class="flex items-center gap-2 p-2 rounded-lg"
            :class="i === chainTree.length - 1 ? 'bg-makoclaw-surface/40 border border-makoclaw-border/30' : 'bg-makoclaw-bg/20'"
            :style="{ marginLeft: `${node.depth * 20}px` }"
          >
            <!-- Tree connector -->
            <span v-if="node.depth > 0" class="text-makoclaw-text-secondary text-xs">└</span>
            <div
              class="w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0"
              :class="getAgentColor(node.agent)"
            >
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
              </svg>
            </div>
            <span class="text-xs font-semibold text-makoclaw-text capitalize">{{ node.agent }}</span>
            <span
              class="text-[10px] px-1.5 py-0.5 rounded"
              :class="getStatusClass(node.status)"
            >
              {{ node.status }}
            </span>
            <!-- Confidence if available -->
            <span v-if="node.confidence" class="text-[10px] ml-auto" :class="getConfidenceColor(node.confidence)">
              {{ Math.round(node.confidence * 100) }}%
            </span>
          </div>
        </div>

        <!-- Active Delegation Progress -->
        <div v-if="delegation" class="p-2 rounded-lg bg-makoclaw-surface/20 border border-makoclaw-border/10">
          <div class="flex items-center justify-between text-[10px] text-makoclaw-text-secondary mb-1">
            <span>{{ delegation.from }} → {{ delegation.to }}</span>
            <span>{{ delegation.iteration }}/{{ delegation.maxIterations }}</span>
          </div>
          <div class="h-1 bg-makoclaw-bg/50 rounded-full overflow-hidden">
            <div
              class="h-full bg-makoclaw-accent/50 rounded-full transition-all duration-300"
              :style="{ width: delegationProgress + '%' }"
            />
          </div>
        </div>

        <!-- Fallback: current agent (when no chain data) -->
        <div
          v-else-if="currentAgent && chainTree.length === 0"
          class="flex items-center gap-3 p-2.5 rounded-xl bg-makoclaw-surface/30 border border-makoclaw-border/20"
        >
          <div
            class="w-8 h-8 rounded-lg flex items-center justify-center"
            :class="getAgentColor(currentAgent.agent)"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="text-xs font-semibold text-makoclaw-text capitalize">{{ currentAgent.agent }}</span>
              <span class="text-[10px] px-1.5 py-0.5 rounded" :class="getStatusClass(currentAgent.status)">
                {{ currentAgent.status }}
              </span>
            </div>
            <p v-if="currentAgent.reason" class="text-[11px] text-makoclaw-text-secondary truncate mt-0.5">
              {{ currentAgent.reason }}
            </p>
          </div>
        </div>
```

**Step 3: Add computed properties for tree and delegation**

In the script section, add:

```javascript
// Build chain tree from delegation chain prop
const chainTree = computed(() => {
  if (props.delegationChain && props.delegationChain.length > 0) {
    return props.delegationChain
  }
  return []
})

const delegation = computed(() => props.activeDelegation)

const delegationProgress = computed(() => {
  if (!delegation.value) return 0
  const { iteration, maxIterations } = delegation.value
  return maxIterations > 0 ? Math.min((iteration / maxIterations) * 100, 100) : 0
})

function getConfidenceColor(confidence) {
  if (confidence >= 0.8) return 'text-green-400'
  if (confidence >= 0.5) return 'text-yellow-400'
  return 'text-red-400'
}
```

**Step 4: Update ChatView to pass new props to TeamActivityPanel**

In `ChatView.vue`, find where `TeamActivityPanel` is rendered (around line 449) and add:

```vue
<TeamActivityPanel
  ref="teamActivityRef"
  :agent-status="teamAgentStatus"
  :team-communications="teamCommunications"
  :involved-agents-list="involvedAgentsList"
  :specialist-report="chatStore.lastSpecialistReport"
  :delegation-chain="chatStore.delegationChain"
  :active-delegation="chatStore.activeDelegation"
/>
```

**Step 5: Commit**

```bash
git add pkg/web/frontend/src/components/Chat/TeamActivityPanel.vue pkg/web/frontend/src/views/ChatView.vue
git commit -m "feat(frontend): add delegation tree view and progress tracking to TeamActivityPanel"
```

---

## Task 10: Update MessageBubble with Delegation Summary

**Files:**
- Modify: `pkg/web/frontend/src/components/MessageBubble.vue:117-165`

**Step 1: Replace the agent workflow details section**

Replace lines 117-165 (the existing `agentActivity` section) with a delegation-aware version:

```vue
        <!-- Delegation Summary (when multi-agent) -->
        <div v-if="msg.agentActivity?.hadMultiAgent">
          <button
            class="flex items-center gap-1.5 mt-3 text-xs text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
            @click="showActivity = !showActivity"
          >
            <svg
              class="w-3 h-3 transition-transform duration-200"
              :class="{ 'rotate-90': showActivity }"
              fill="none" stroke="currentColor" viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
            View delegation details
          </button>
          <div
            v-if="showActivity"
            class="mt-2 pl-3 border-l-2 border-makoclaw-border/30 space-y-2"
          >
            <!-- Delegation tree -->
            <div v-if="msg.agentActivity.delegations?.length > 0" class="space-y-1.5">
              <div
                v-for="(del, i) in msg.agentActivity.delegations"
                :key="i"
                class="flex items-center gap-2 text-xs text-makoclaw-text-secondary"
              >
                <SpecialistBadge :name="del.to || del.specialist" class="text-[10px]" />
                <span v-if="del.confidence" :class="confidenceColor(del.confidence)">
                  {{ Math.round((del.confidence || 0) * 100) }}%
                </span>
                <span v-if="del.toolsUsed?.length" class="opacity-50">
                  used: {{ del.toolsUsed.join(', ') }}
                </span>
              </div>
            </div>

            <!-- Agent timeline (fallback for messages without delegation data) -->
            <div v-else-if="msg.agentActivity.history?.length > 0" class="space-y-1">
              <div
                v-for="(entry, i) in msg.agentActivity.history"
                :key="i"
                class="flex items-center gap-2 text-xs text-makoclaw-text-secondary"
              >
                <span
                  class="w-1.5 h-1.5 rounded-full flex-shrink-0"
                  :class="activityDotColor(entry.status)"
                />
                <span class="capitalize">{{ entry.agent }}</span>
                <span class="opacity-60">{{ entry.status }}</span>
                <span v-if="entry.specialistName" class="opacity-60">{{ entry.specialistName }}</span>
              </div>
            </div>

            <!-- Reports -->
            <div
              v-for="report in msg.agentActivity.reports"
              :key="report.specialist_name || report.specialistName"
              class="flex items-center gap-2 text-xs text-makoclaw-text-secondary mt-1"
            >
              <span class="w-1.5 h-1.5 rounded-full bg-blue-400 flex-shrink-0" />
              <span class="capitalize">{{ report.specialist_name || report.specialistName }}</span>
              <span v-if="report.confidence" class="opacity-60">
                confidence: {{ Math.round((report.confidence || 0) * 100) }}%
              </span>
            </div>
          </div>
        </div>
```

**Step 2: Add confidenceColor helper**

In the script section, add:

```javascript
function confidenceColor(confidence) {
  if (confidence >= 0.8) return 'text-green-400'
  if (confidence >= 0.5) return 'text-yellow-400'
  return 'text-red-400'
}
```

**Step 3: Commit**

```bash
git add pkg/web/frontend/src/components/MessageBubble.vue
git commit -m "feat(frontend): show delegation summary with confidence and tools in MessageBubble"
```

---

## Task 11: Build Frontend and Integration Test

**Files:**
- Build: `pkg/web/frontend/`
- Build: Go project root

**Step 1: Build frontend**

Run: `cd /c/Users/tfurt/source/repos/kakoclaw && make build-frontend`
Expected: Build succeeds with no errors.

**Step 2: Fix any frontend build errors**

If there are TypeScript/compilation errors, fix them based on error output.

**Step 3: Build Go backend**

Run: `cd /c/Users/tfurt/source/repos/kakoclaw && go build ./...`
Expected: Build succeeds.

**Step 4: Run existing tests**

Run: `cd /c/Users/tfurt/source/repos/kakoclaw && go test ./pkg/agent/... -v -count=1`
Expected: All tests pass (existing + new callback tests).

**Step 5: Commit the built frontend assets**

```bash
git add pkg/web/dist/
git commit -m "build: rebuild frontend with multi-agent delegation chain UI"
```

---

## Task 12: Final Integration Verification

**Step 1: Start the web server**

Run: `cd /c/Users/tfurt/source/repos/kakoclaw && make run ARGS="web"`

**Step 2: Manual verification checklist**

Verify in the browser:
- [ ] Chat with orchestrator enabled — specialist delegation shows chain in AgentStatusIndicator
- [ ] AgentEventBubble shows chain context (e.g., "developer is working (via orchestrator)")
- [ ] TeamActivityPanel shows delegation tree with depth indentation
- [ ] MessageBubble shows "View delegation details" with confidence and tools
- [ ] Multiple delegations work (orchestrator delegates, specialist requests colleague)
- [ ] Max delegation limit (3) triggers graceful response compilation
- [ ] Empty specialist response shows informative fallback message
- [ ] stream_end includes delegation_summary

**Step 3: Final commit with all changes**

If any fixes were needed:
```bash
git add -A
git commit -m "fix: address integration issues in multi-agent delegation flow"
```
