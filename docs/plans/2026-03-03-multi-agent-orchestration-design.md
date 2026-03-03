# Multi-Agent Orchestration: Event-Driven Pipeline Design

**Date:** 2026-03-03
**Approach:** Event-Driven Pipeline (Approach A)
**Scope:** Backend response flow + WebSocket protocol + Frontend UI visibility

---

## Problem Statement

The current multi-agent system has three critical issues:

1. **Synchronous blocking delegation** - `processSpecialistTask()` blocks the entire orchestrator loop while waiting for specialist completion. No progress feedback reaches the user.
2. **Lost/incomplete responses** - Specialist responses don't reliably reach the orchestrator. Empty responses cause re-delegation loops. No structured result format.
3. **Poor UI visibility** - The frontend has agent status components but they don't show delegation chains, inter-specialist collaboration, or real-time progress with iteration tracking.

## Design Decisions

- **Orchestrator unified model** - Orchestrator always receives specialist response and synthesizes before presenting to user
- **Visible delegation chain** - User sees full chain: orchestrator → specialist A → specialist B as inline events
- **Result-only display** - User sees progress indicators while working, but only the final attributed result (no specialist streaming)
- **End-to-end solution** - Backend + protocol + frontend designed together

---

## Section 1: Backend - Callback Propagation

### Current Problem
When `processSpecialistTask()` (orchestrator.go:694-750) executes a specialist, the specialist's callbacks (`AgentStatusEvent`, `ContentSegment`) are NOT propagated to the parent context. The WebSocket callbacks are in the orchestrator's context only.

### Solution: Callback Chain Inheritance

When a specialist is executed, extract parent callbacks and inject wrapped versions into the specialist context. The wrapper adds delegation metadata (chain, depth, parent agent).

```
Orchestrator Context (has WebSocket callbacks)
    ↓ inherits callbacks (wrapped with depth=1)
Specialist Context (executes with propagated callbacks)
    ↓ inherits callbacks (wrapped with depth=2)
Colleague Context (if request_colleague, also propagates)
```

### Changes to `orchestrator.go`

In `processSpecialistTask()`, before specialist execution:

1. Extract `AgentStatusCallback` from parent context
2. Create wrapped callback that adds `DelegationChain`, `DelegationDepth`, `ParentAgent`
3. Inject into specialist context via `ContextWithAgentStatusCallback()`
4. Same for `SpecialistReportCallback` and `ContentSegmentCallback`

### Enriched `AgentStatusEvent`

New fields added to existing struct:
- `DelegationChain []string` - e.g., `["orchestrator", "developer", "security_analyst"]`
- `DelegationDepth int` - 0=orchestrator, 1=specialist, 2=colleague
- `ParentAgent string` - who delegated to this agent
- `Timestamp time.Time` - when event occurred

### Enriched `SpecialistReport`

New fields:
- `DelegationChain []string`
- `DelegationDepth int`
- `ToolsUsed []string`
- `IterationsUsed int`

### Changes to `specialist.go`

In `RequestColleagueTool.Execute()`: Same callback wrapping pattern with depth+1.

### Event Flow Example

```
→ agent_status {agent:"orchestrator", status:"delegating", specialist:"developer", depth:0}
→ agent_status {agent:"developer", status:"working", chain:["orchestrator","developer"], depth:1}
→ agent_status {agent:"developer", status:"delegating", specialist:"security", depth:1}
→ agent_status {agent:"security", status:"working", chain:["orchestrator","developer","security"], depth:2}
→ specialist_report {agent:"security", confidence:85, depth:2}
→ agent_status {agent:"security", status:"complete", depth:2}
→ specialist_report {agent:"developer", confidence:90, depth:1}
→ agent_status {agent:"developer", status:"complete", depth:1}
→ agent_status {agent:"orchestrator", status:"synthesizing", depth:0}
```

---

## Section 2: Backend - Structured Response Flow

### Current Problem
Specialist returns plain string. Empty/truncated responses cause orchestrator to loop. No quality metadata.

### Solution: DelegationResult

Replace plain string return with structured JSON:

```go
type DelegationResult struct {
    Success        bool     `json:"success"`
    SpecialistName string   `json:"specialist_name"`
    Response       string   `json:"response"`
    Confidence     int      `json:"confidence"`
    ToolsUsed      []string `json:"tools_used"`
    HelpRequested  *string  `json:"help_requested"`
    Iterations     int      `json:"iterations_used"`
    Chain          []string `json:"delegation_chain"`
}
```

### Fail-Safe: Empty Response Detection

In `processSpecialistTask()`:
```go
if result == "" || len(strings.TrimSpace(result)) < 10 {
    return DelegationResult{Success: false, Response: "empty", Confidence: 0}
}
```

### Fail-Safe: Re-delegation Limit

Track delegation count per user message. If >= 3 delegations without satisfactory result, force response with accumulated partial results:

```go
if oa.delegationCount >= 3 {
    emitAgentStatus(ctx, "max_delegations_reached")
    // Compile all partial responses and return best effort
}
```

### Fail-Safe: Iteration Progress Emission

During specialist execution, emit `delegation_update` events periodically (on each LLM iteration) so frontend can show progress:

```go
// In specialist's runLLMIteration, emit progress
emitDelegationUpdate(ctx, DelegationUpdate{
    DelegationID: delegationID,
    From: parentAgent,
    To: specialistName,
    Iteration: currentIteration,
    MaxIterations: maxIterations,
    ElapsedMs: elapsed,
})
```

---

## Section 3: WebSocket Protocol Enrichment

### Existing events (unchanged, backward-compatible)
- `stream_start`, `stream`, `stream_end`
- `agent_status`, `specialist_report`, `content_segment`, `tool_call`

### New fields on existing events

**`agent_status`** - adds: `delegation_chain`, `delegation_depth`, `parent_agent`, `timestamp`

**`specialist_report`** - adds: `delegation_chain`, `delegation_depth`, `tools_used`, `iterations_used`

**`stream_end`** - adds: `delegation_summary` array with per-delegation stats

### New event: `delegation_update`

Real-time progress of active delegation:
```json
{
  "type": "delegation_update",
  "delegation_id": "del_abc123",
  "from": "orchestrator",
  "to": "developer",
  "status": "in_progress",
  "started_at": "2026-03-03T10:15:30Z",
  "elapsed_ms": 5200,
  "iteration": 3,
  "max_iterations": 20
}
```

---

## Section 4: Frontend - UI Visibility

### Principle
"Result-only with attributed progress" - User sees who is working and the delegation chain in real time, but only the final synthesized result.

### Component Changes

#### 1. `AgentStatusIndicator.vue` - Delegation Chain Display

Show nested chain instead of single agent:

```
┌─────────────────────────────────────────────┐
│ 🟣 orchestrator → 🟢 developer (working)    │
│   "Implementing authentication module"       │
│   Iteration 3/20 • 5.2s elapsed             │
└─────────────────────────────────────────────┘
```

With sub-delegation:
```
┌─────────────────────────────────────────────┐
│ 🟣 orchestrator → 🟢 developer              │
│   → 🟡 security_analyst (working)           │
│   "Reviewing code for vulnerabilities"       │
│   Iteration 2/20 • 3.1s elapsed             │
└─────────────────────────────────────────────┘
```

#### 2. `AgentEventBubble.vue` - Chain Events

Enriched inline event bubbles with chain context:

```
────── orchestrator delegated to developer ──────
────── developer requested help from security_analyst ──────
────── security_analyst completed (85% confidence) ──────
────── developer completed (90% confidence) ──────
```

#### 3. `TeamActivityPanel.vue` - Tree View

Replace flat structure with delegation tree:

```
Team Activity
├─ orchestrator (analyzing)
│  └─ developer (working) - 90% confidence
│     └─ security_analyst (complete) - 85% confidence
│
Communications:
  developer → security_analyst: "Check auth code"
  security_analyst → developer: "Found SQL injection risk"
```

#### 4. `MessageBubble.vue` - Delegation Summary

Collapsible section at end of message:

```
🤖 developer + security_analyst contributed
▸ View delegation details
  ├─ developer: 90% confidence, used: edit_file, read_file
  │  └─ security_analyst: 85% confidence, used: read_file
  └─ 3 tool calls, 2 delegations, 8.3s total
```

### Store Changes (`chatStore.js`)

New state:
```javascript
delegationChain: [],           // Active chain of {agent, status, depth, startedAt}
activeDelegation: null,        // Currently executing delegation
delegationHistory: [],         // Completed delegations for current message
```

New actions:
```javascript
updateDelegationChain(event)   // Process agent_status with chain info
completeDelegation(report)     // When specialist_report arrives
buildDelegationSummary()       // For stream_end, compile summary
```

### WebSocket Handler Changes (`ChatView.vue`)

`handleMessage()` enhanced to:
- Process `delegation_chain` and `delegation_depth` from agent_status events
- Process new `delegation_update` events for progress tracking
- Build delegation tree from events for TeamActivityPanel
- Compile delegation summary on `stream_end`

---

## Architecture Summary

```
User Message → WebSocket → Server
                              ↓
                    AgentLoop (Orchestrator)
                    ├─ LLM decides to delegate
                    ├─ DelegationTool.Execute()
                    │   ├─ Emit: agent_status(delegating)
                    │   ├─ Wrap parent callbacks → specialist context
                    │   ├─ specialist.ProcessWithSpeciality()
                    │   │   ├─ Emit: agent_status(working, depth=1)
                    │   │   ├─ [optional] request_colleague
                    │   │   │   ├─ Emit: agent_status(working, depth=2)
                    │   │   │   └─ Return structured result
                    │   │   ├─ Emit: specialist_report(depth=1)
                    │   │   └─ Return DelegationResult{}
                    │   ├─ Emit: agent_status(complete)
                    │   └─ Return JSON result to LLM
                    ├─ LLM synthesizes response
                    └─ stream_end with delegation_summary
                              ↓
                    WebSocket → Frontend
                    ├─ AgentStatusIndicator (chain display)
                    ├─ AgentEventBubble (inline events)
                    ├─ TeamActivityPanel (tree view)
                    └─ MessageBubble (delegation summary)
```

---

## Files to Modify

### Backend (Go)
| File | Changes |
|------|---------|
| `pkg/agent/loop.go` | Add DelegationChain/Depth fields to AgentStatusEvent, add DelegationUpdate event type and callback |
| `pkg/agent/orchestrator.go` | Callback propagation in processSpecialistTask(), DelegationResult struct, re-delegation limit, delegation_update emission |
| `pkg/agent/specialist.go` | Callback propagation in RequestColleagueTool, ToolsUsed tracking |
| `pkg/web/server.go` | Handle new event types in WebSocket handler, emit delegation_update |

### Frontend (Vue)
| File | Changes |
|------|---------|
| `pkg/web/frontend/src/stores/chatStore.js` | delegationChain state, updateDelegationChain/completeDelegation/buildDelegationSummary actions |
| `pkg/web/frontend/src/views/ChatView.vue` | Handle delegation_update in handleMessage(), build delegation tree |
| `pkg/web/frontend/src/components/Chat/AgentStatusIndicator.vue` | Chain display with depth, iteration progress |
| `pkg/web/frontend/src/components/Chat/AgentEventBubble.vue` | Chain context in events, confidence display |
| `pkg/web/frontend/src/components/Chat/TeamActivityPanel.vue` | Tree view instead of flat list |
| `pkg/web/frontend/src/components/MessageBubble.vue` | Delegation summary section |

---

## Testing Strategy

1. **Unit tests**: DelegationResult serialization, callback wrapping, chain building
2. **Integration test**: Send message that triggers delegation → verify events arrive in order
3. **Frontend**: Verify delegation chain renders correctly with mock WebSocket events
4. **Edge cases**: Empty specialist response, max re-delegation limit, timeout, colleague recursion depth
