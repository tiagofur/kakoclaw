# Specifications: Multi-Agent Orchestration Visibility

**Change ID:** multi-agent-orchestration-visibility
**Status:** Draft
**Author:** SDD Sub-Agent (spec phase)
**Created:** 2026-02-27
**Based on:** [proposal.md](proposal.md), [exploration.md](exploration.md)
**RFC 2119 Keywords:** MUST, MUST NOT, SHALL, SHALL NOT, SHOULD, SHOULD NOT, MAY, REQUIRED, OPTIONAL

---

## Table of Contents

1. [Domain: agent-orchestration](#1-domain-agent-orchestration)
   - [1.1 Callback Propagation](#11-callback-propagation)
   - [1.2 Specialist Streaming](#12-specialist-streaming)
   - [1.3 Concurrency Fix for Tool Registry](#13-concurrency-fix-for-tool-registry)
   - [1.4 Real-Time Agent Status Events](#14-real-time-agent-status-events)
   - [1.5 Content Segments with Per-Agent Attribution](#15-content-segments-with-per-agent-attribution)
2. [Domain: agent-visibility](#2-domain-agent-visibility)
   - [2.1 Real-Time Agent Status Indicator](#21-real-time-agent-status-indicator)
   - [2.2 Content Segments Rendered with Agent Attribution](#22-content-segments-rendered-with-agent-attribution)
   - [2.3 Dynamic Specialist Badges](#23-dynamic-specialist-badges)
   - [2.4 Agent Timeline / History View](#24-agent-timeline--history-view)
   - [2.5 WebSocket Events with Agent Metadata During Streaming](#25-websocket-events-with-agent-metadata-during-streaming)
3. [Domain: agent-token-optimization](#3-domain-agent-token-optimization)
   - [3.1 Specialist Context Trimming](#31-specialist-context-trimming)
   - [3.2 Parallel Delegation Support](#32-parallel-delegation-support)
   - [3.3 Per-Specialist Token Budget Tracking](#33-per-specialist-token-budget-tracking)
4. [Cross-Cutting Concerns](#4-cross-cutting-concerns)
5. [Glossary](#5-glossary)

---

## 1. Domain: agent-orchestration

This domain covers all backend changes to the agent loop, orchestrator, and specialist execution pipeline that enable callback propagation, streaming, and concurrency safety.

### 1.1 Callback Propagation

#### REQ-ORCH-CB-001: DelegationCallbacks struct

The system MUST define a `DelegationCallbacks` struct in `pkg/agent/loop.go` that bundles all callback types needed for specialist delegation:

- `OnToken StreamCallback` -- streamed text tokens
- `OnTool ToolCallback` -- tool call start/finish/error events
- `OnAgentStatus AgentStatusCallback` -- agent status changes
- `OnContentSegment ContentSegmentCallback` -- attributed content segments
- `SkipHistory bool` -- whether to skip session history loading for this call

**Rationale:** Currently, callbacks are split between explicit parameters (`OnToken`, `OnTool` in `processOptions`) and context values (`AgentStatusCallback`, `ContentSegmentCallback` via `ContextWith*`). Bundling them into a single struct creates a clear contract for specialist delegation.

#### Scenario: DelegationCallbacks struct is well-formed

```
Given the DelegationCallbacks struct exists in pkg/agent/loop.go
When a caller constructs a DelegationCallbacks with all fields populated
Then all fields MUST be accessible and typed correctly
And the struct MUST be exported for use in orchestrator.go and specialist.go
```

#### Scenario: DelegationCallbacks struct with nil callbacks

```
Given a DelegationCallbacks where all callback fields are nil
When passed to ProcessDirectWithCallbacks
Then the system MUST NOT panic
And MUST fall back to non-streaming execution (equivalent to ProcessDirect)
```

---

#### REQ-ORCH-CB-002: ProcessDirectWithCallbacks method on AgentLoop

The system MUST add a new method `ProcessDirectWithCallbacks` to `AgentLoop` with the following signature:

```go
func (al *AgentLoop) ProcessDirectWithCallbacks(
    ctx context.Context,
    content, sessionKey string,
    callbacks DelegationCallbacks,
) (string, error)
```

This method MUST:
1. Construct a `bus.InboundMessage` with `Channel: "cli"`, `ChatID: "direct"`.
2. Populate a `processOptions` struct with `OnToken`, `OnTool`, `OnAgentStatus`, `OnContentSegment` from the `callbacks` parameter.
3. If `callbacks.SkipHistory` is true, the method MUST set a flag on `processOptions` that causes session history loading to be bypassed.
4. If `callbacks.OnToken` is non-nil, route to `runAgentLoopStream`. Otherwise, route to `runAgentLoop`.
5. Return the full accumulated response string and any error.

The method MUST NOT modify the existing `ProcessDirect`, `ProcessDirectWithModelStream`, or any other existing `Process*` methods.

#### Scenario: Specialist processes with all callbacks

```
Given an AgentLoop for a specialist with a streaming-capable provider
And a DelegationCallbacks with OnToken, OnTool, OnAgentStatus, OnContentSegment all set
When ProcessDirectWithCallbacks is called with a user message
Then the LLM response MUST be streamed token-by-token via OnToken
And any tool calls MUST emit events via OnTool (status "started" then "finished" or "error")
And the full response text MUST be returned as the string result
```

#### Scenario: ProcessDirectWithCallbacks with nil OnToken falls back to non-streaming

```
Given an AgentLoop for a specialist
And a DelegationCallbacks where OnToken is nil but OnTool is set
When ProcessDirectWithCallbacks is called
Then the system MUST use runAgentLoop (non-streaming)
And tool call events MUST still be emitted via OnTool
And the full response text MUST be returned
```

#### Scenario: ProcessDirectWithCallbacks with SkipHistory=true

```
Given a specialist AgentLoop with session history containing 20 prior messages
And a DelegationCallbacks with SkipHistory set to true
When ProcessDirectWithCallbacks is called with sessionKey "specialist_developer"
Then the LLM request MUST NOT include any prior session history messages
And MUST only include the system prompt and the current user message
And MUST still return the full response
```

#### Scenario: ProcessDirectWithCallbacks propagates context cancellation

```
Given a specialist AgentLoop processing via ProcessDirectWithCallbacks
And the ctx is canceled (e.g., user disconnects WebSocket)
When the cancellation signal reaches the LLM call
Then the method MUST return with context.Canceled error
And MUST NOT leak goroutines
And any in-progress tool execution SHOULD be aborted
```

---

#### REQ-ORCH-CB-003: processSpecialistTask extracts and forwards callbacks

The `processSpecialistTask` method in `pkg/agent/orchestrator.go` MUST be modified to:

1. Extract `AgentStatusCallback` from `ctx` using `agentStatusCallbackFromCtx(ctx)`.
2. Extract `ContentSegmentCallback` from `ctx` using `contentSegmentCallbackFromCtx(ctx)`.
3. Retrieve `OnToken` and `OnTool` callbacks from the orchestrator's active `processOptions` (these MUST be threaded through from the WebSocket handler's streaming path). The mechanism for retrieving these from the orchestrator's own execution context SHOULD use additional context keys (see REQ-ORCH-CB-004).
4. Bundle all extracted callbacks into a `DelegationCallbacks` struct.
5. Call `specialist.ProcessWithSpecialityStream(ctx, task, callbacks)` instead of `specialist.ProcessWithSpeciality(ctx, task)`.
6. If no streaming callbacks are available (all nil), MUST fall back to calling `specialist.ProcessWithSpeciality(ctx, task)` to preserve backward compatibility.

#### Scenario: WebSocket streaming path threads all callbacks to specialist

```
Given a user sends a message via WebSocket chat
And the orchestrator LLM decides to call delegate_to_specialist for "developer"
And the WebSocket handler injected OnToken, OnTool, OnAgentStatus, and OnContentSegment callbacks
When processSpecialistTask executes
Then it MUST extract all four callback types
And MUST pass them as DelegationCallbacks to ProcessWithSpecialityStream
And the specialist's streamed tokens MUST reach the WebSocket client
```

#### Scenario: CLI/non-streaming path has no callbacks to extract

```
Given a user sends a message via the CLI (ProcessDirect, no callbacks)
And the orchestrator delegates to a specialist
When processSpecialistTask tries to extract callbacks from ctx
Then all extracted callbacks MUST be nil
And the method MUST fall back to ProcessWithSpeciality (non-streaming)
And the result MUST be returned to the orchestrator as today
```

#### Scenario: Timeout during specialist streaming

```
Given processSpecialistTask is executing with a 5-minute timeout
And the specialist is streaming tokens via ProcessWithSpecialityStream
When the timeout expires
Then processSpecialistTask MUST emit an agent_status event with status "timeout"
And MUST return an error indicating timeout
And MUST cancel the specialist's context to stop streaming
And MUST NOT leave dangling goroutines
```

---

#### REQ-ORCH-CB-004: Streaming callback context keys for orchestrator forwarding

The system MUST define context keys to pass `OnToken` and `OnTool` callbacks through `context.Context`, similar to the existing pattern for `AgentStatusCallback` and `ContentSegmentCallback`:

```go
func ContextWithStreamCallbacks(ctx context.Context, onToken StreamCallback, onTool ToolCallback) context.Context
func streamCallbacksFromCtx(ctx context.Context) (StreamCallback, ToolCallback)
```

The `processMessageWithModelStream` method MUST embed these callbacks into the context before calling `runAgentLoopStream`, so that nested code (such as `DelegationTool.Execute`) can retrieve them.

#### Scenario: Stream callbacks are available in DelegationTool.Execute

```
Given a user sends a message via WebSocket with streaming enabled
And the WebSocket handler provides OnToken and OnTool callbacks
When the orchestrator's runAgentLoopStream calls a tool (delegate_to_specialist)
And DelegationTool.Execute retrieves callbacks via streamCallbacksFromCtx(ctx)
Then OnToken MUST be non-nil
And OnTool MUST be non-nil
```

#### Scenario: Non-streaming paths do not inject stream callbacks

```
Given a user sends a message via ProcessDirect (no streaming)
When DelegationTool.Execute calls streamCallbacksFromCtx(ctx)
Then both OnToken and OnTool MUST be nil
And the delegation MUST proceed non-streaming
```

---

### 1.2 Specialist Streaming

#### REQ-ORCH-SS-001: ProcessWithSpecialityStream method on SpecialistAgent

The system MUST add a new method `ProcessWithSpecialityStream` to `SpecialistAgent` in `pkg/agent/specialist.go`:

```go
func (sa *SpecialistAgent) ProcessWithSpecialityStream(
    ctx context.Context,
    userMessage string,
    callbacks DelegationCallbacks,
) (string, error)
```

This method MUST:
1. Create a per-call filtered tool registry via `sa.ToolFilter()` (see REQ-ORCH-CC-001).
2. Wrap the provided callbacks with agent attribution (see REQ-ORCH-CS-001).
3. Call `ProcessDirectWithCallbacks` on a temporary execution context that uses the filtered tool registry.
4. Return the full accumulated response and any error.
5. NOT modify `sa.tools` (the shared tool registry field on `SpecialistAgent`).

The existing `ProcessWithSpeciality` method MUST remain unchanged as a non-streaming fallback.

#### Scenario: Specialist streams tokens to the user via wrapped callback

```
Given a specialist "developer" with a streaming-capable provider
And DelegationCallbacks with all callbacks populated
When ProcessWithSpecialityStream is called with task "implement login form"
Then the specialist's LLM response tokens MUST be streamed via the wrapped OnToken callback
And the wrapped OnToken callback MUST NOT modify the token content
And the full response MUST also be returned as the string result
```

#### Scenario: Specialist without streaming-capable provider

```
Given a specialist "analyst" using an OllamaProvider that does not implement StreamingLLMProvider
And DelegationCallbacks with OnToken set
When ProcessWithSpecialityStream is called
Then ProcessDirectWithCallbacks MUST detect the non-streaming provider
And MUST fall back to non-streaming execution
And the full response MUST still be returned
And OnToken MAY be called once with the complete response, or not at all
```

#### Scenario: Specialist tool calls emit events with attribution

```
Given a specialist "developer" executes a tool "read_file" during its LLM loop
And the OnTool callback is wrapped with agent attribution
When the tool starts execution
Then an OnTool event MUST be emitted with status "started" and Agent "developer"
When the tool completes
Then an OnTool event MUST be emitted with status "finished" and Agent "developer"
```

#### Scenario: Specialist error during streaming

```
Given a specialist "testing" is streaming a response via ProcessWithSpecialityStream
And the LLM provider returns an error mid-stream
When the error occurs
Then the accumulated partial response MUST NOT be returned as a successful result
And the error MUST be propagated to processSpecialistTask
And the orchestrator MUST see the error as a tool result
```

---

### 1.3 Concurrency Fix for Tool Registry

#### REQ-ORCH-CC-001: Per-call tool registry copy (eliminate shared mutable state)

The system MUST fix the concurrency bug in `ProcessWithSpeciality` (and the new `ProcessWithSpecialityStream`) by eliminating the pattern of swapping `sa.tools`:

1. `ProcessWithSpecialityStream` MUST call `sa.ToolFilter()` to create a **new** `*tools.ToolRegistry` instance for each call.
2. The filtered registry MUST be passed to `ProcessDirectWithCallbacks` as a tool registry override rather than modifying the shared `sa.tools` field.
3. The `processMu` mutex in `SpecialistAgent` MAY be removed or deprecated once the tool-swap pattern is eliminated.
4. The existing `ProcessWithSpeciality` SHOULD also be updated to use the same per-call copy pattern for consistency, even though it is the fallback path.

The mechanism for passing a tool registry override MUST be one of:
- (a) A new field `ToolsOverride *tools.ToolRegistry` on `processOptions`, OR
- (b) A new method `ProcessDirectWithCallbacksAndTools` that accepts a registry parameter, OR
- (c) An `AgentLoop` method `WithToolsOverride(registry)` that returns a lightweight wrapper.

**Design decision:** The chosen mechanism SHALL be documented in the design artifact. The spec does not prescribe which approach to use, only that shared mutable state (`sa.tools` swap) MUST be eliminated.

#### Scenario: Two concurrent delegations to the same specialist

```
Given specialist "developer" is registered in the registry
And two goroutines concurrently call ProcessWithSpecialityStream on the same SpecialistAgent
When goroutine A creates filtered registry A with tools [read_file, write_file, exec]
And goroutine B creates filtered registry B with tools [read_file, write_file, exec]
Then each goroutine MUST use its own independent registry copy
And changes to registry A (e.g., mid-execution tool registration) MUST NOT affect registry B
And the shared sa.tools MUST remain unchanged throughout
```

#### Scenario: Specialist concurrent calls with go test -race

```
Given the test suite runs with `go test -race ./pkg/agent/...`
And a test exercises two concurrent ProcessWithSpecialityStream calls on the same specialist
When the test completes
Then there MUST be zero data race warnings
And both calls MUST return correct results
```

#### Scenario: Original sa.tools is never modified during execution

```
Given specialist "developer" has sa.tools containing 20 registered tools
When ProcessWithSpecialityStream is called with allowedTools = [read_file, write_file]
Then sa.tools MUST still contain all 20 tools after the call completes
And sa.tools MUST still contain all 20 tools during the call
And the specialist's LLM MUST only see read_file and write_file in its available tools
```

---

### 1.4 Real-Time Agent Status Events

#### REQ-ORCH-AS-001: Granular status events during specialist execution

The system MUST emit `AgentStatusEvent` events at the following points during specialist execution:

| Point | Agent | Status | Required Fields |
|-------|-------|--------|-----------------|
| Orchestrator begins analysis | `"orchestrator"` | `"analyzing"` | Timestamp |
| Delegation decision made | `"orchestrator"` | `"delegating"` | SpecialistName, Reason, Timestamp |
| Specialist begins work | `{specialist_name}` | `"working"` | Timestamp |
| Specialist tool execution starts | `{specialist_name}` | `"tool_call"` | Timestamp (NEW) |
| Specialist completes successfully | `{specialist_name}` | `"complete"` | Timestamp |
| Specialist times out | `{specialist_name}` | `"timeout"` | Timestamp |
| Specialist errors | `{specialist_name}` | `"error"` | Timestamp (NEW) |

Events marked (NEW) represent additions to the current set. The existing events ("analyzing", "delegating", "working", "complete", "timeout") MUST continue to be emitted as today.

#### Scenario: Full lifecycle of a successful delegation

```
Given the orchestrator is enabled with specialist "developer" configured
And the WebSocket handler has injected AgentStatusCallback into ctx
When the user asks "write a hello world function"
And the orchestrator delegates to "developer"
Then the following agent_status events MUST be emitted in order:
  1. agent="orchestrator", status="analyzing"
  2. agent="orchestrator", status="delegating", specialist_name="developer"
  3. agent="developer", status="working"
  4. agent="developer", status="complete"
And each event MUST have a non-zero Timestamp
```

#### Scenario: Specialist emits error status on failure

```
Given specialist "developer" encounters an LLM provider error during execution
When the error is caught in ProcessWithSpecialityStream
Then an agent_status event MUST be emitted with agent="developer", status="error"
And the error MUST be propagated back to processSpecialistTask
```

#### Scenario: No agent status callback available (non-WebSocket path)

```
Given a message is processed via CLI (ProcessDirect) with no AgentStatusCallback in ctx
When the orchestrator delegates to a specialist
Then emitAgentStatus calls MUST be no-ops (as today)
And the delegation MUST still succeed
And no error MUST be raised from missing callback
```

---

#### REQ-ORCH-AS-002: AgentStatusEvent extended with optional metadata

The `AgentStatusEvent` struct MAY be extended with an optional `Metadata` field:

```go
type AgentStatusEvent struct {
    Agent          string                 `json:"agent"`
    Status         string                 `json:"status"`
    SpecialistName string                 `json:"specialist_name,omitempty"`
    Reason         string                 `json:"reason,omitempty"`
    Timestamp      time.Time              `json:"timestamp"`
    Metadata       map[string]interface{} `json:"metadata,omitempty"` // NEW, OPTIONAL
}
```

The `Metadata` field MAY carry additional context such as tool name for `"tool_call"` status events. This field MUST be `omitempty` to maintain backward compatibility.

---

### 1.5 Content Segments with Per-Agent Attribution

#### REQ-ORCH-CS-001: Callback wrappers add agent attribution

The system MUST implement callback wrapper functions that annotate events with the specialist's name:

1. **wrapTokenCallback(agentName string, onToken StreamCallback) StreamCallback**: The wrapper MUST forward the token as-is without modification. Token attribution is handled separately via agent status events and content segments.

2. **wrapToolCallback(agentName string, onTool ToolCallback) ToolCallback**: The wrapper MUST set the `Agent` field on `ToolEvent` before forwarding.

3. **wrapAgentStatusCallback(agentName string, onStatus AgentStatusCallback) AgentStatusCallback**: The wrapper MUST set the `Agent` field to `agentName` if the event's Agent is empty or is the specialist's internal model name.

These wrappers MUST be applied in `ProcessWithSpecialityStream` before passing callbacks to `ProcessDirectWithCallbacks`.

#### Scenario: Tool events from specialist include agent attribution

```
Given specialist "developer" executes tool "write_file"
And OnTool is wrapped with wrapToolCallback("developer", originalOnTool)
When the tool emits a ToolEvent with Name="write_file", Status="started"
Then the wrapped callback MUST set ToolEvent.Agent to "developer"
And MUST forward the event to the original OnTool callback
```

#### Scenario: Token callback does not modify content

```
Given specialist "developer" streams token "Hello"
And OnToken is wrapped with wrapTokenCallback("developer", originalOnToken)
When the wrapped callback is invoked with "Hello"
Then it MUST call originalOnToken with exactly "Hello"
And MUST NOT prepend, append, or modify the token string
```

---

#### REQ-ORCH-CS-002: ToolEvent struct extended with Agent field

The `ToolEvent` struct in `pkg/agent/loop.go` MUST be extended:

```go
type ToolEvent struct {
    Name   string                 `json:"name"`
    Args   map[string]interface{} `json:"arguments"`
    Result string                 `json:"result,omitempty"`
    Status string                 `json:"status"` // "started", "finished", "error"
    Agent  string                 `json:"agent,omitempty"` // NEW
}
```

The `Agent` field MUST be `omitempty` to ensure backward compatibility. Existing code that constructs `ToolEvent` without the `Agent` field MUST continue to compile and work.

#### Scenario: ToolEvent without Agent field (backward compatibility)

```
Given existing code constructs ToolEvent{Name: "exec", Status: "started"}
When this event is serialized to JSON
Then the JSON MUST NOT contain an "agent" key
And existing WebSocket handlers MUST process the event without error
```

#### Scenario: ToolEvent with Agent field set

```
Given specialist "developer" constructs ToolEvent{Name: "exec", Status: "started", Agent: "developer"}
When this event is serialized to JSON
Then the JSON MUST contain "agent": "developer"
```

---

#### REQ-ORCH-CS-003: Content segment emitted at specialist completion

The existing behavior where `processSpecialistTask` emits a `ContentSegment` with the specialist's final result MUST be preserved:

```
Given specialist "developer" completes and returns result "Here is the function..."
When processSpecialistTask receives the result
Then it MUST emit a ContentSegment with:
  - Agent: "developer"
  - Content: "Here is the function..."
  - SegmentID: "seg_developer_{timestamp_nanos}"
  - Timestamp: current time
```

This behavior MUST NOT be affected by whether streaming was used. The final content segment serves as the authoritative attributed result.

---

## 2. Domain: agent-visibility

This domain covers all frontend and API changes that render real-time agent information to the user.

### 2.1 Real-Time Agent Status Indicator

#### REQ-VIS-SI-001: AgentStatusIndicator handles all specialist statuses

The `AgentStatusIndicator.vue` component MUST handle the following statuses:

| Status | Display | Border Color | Icon Color |
|--------|---------|-------------|------------|
| `analyzing` | "Analyzing your request..." | Blue | Blue |
| `delegating` | "Delegating to {agent}..." | Purple | Purple |
| `working` | "{agent} is working on your request..." | Green | Green |
| `tool_call` | "{agent} is executing a tool..." | Amber | Amber (NEW) |
| `error` | "{agent} encountered an error" | Red | Red (NEW) |
| `complete` | Hidden (transitions to idle) | -- | -- |
| `timeout` | "{agent} timed out" | Red | Red (NEW) |

New statuses (`tool_call`, `error`, `timeout`) MUST be added to the component's `statusMap`, `iconColorClass`, and `borderColorClass` computed properties.

#### Scenario: Status indicator shows "working" with specialist badge

```
Given the WebSocket receives an agent_status event:
  { agent: "developer", status: "working" }
When chatStore.setAgentStatus is called
Then AgentStatusIndicator MUST be visible (isActive=true)
And MUST display a SpecialistBadge for "developer"
And MUST show text "developer is working on your request..."
And MUST have a green left border
```

#### Scenario: Status indicator shows "error" in red

```
Given the WebSocket receives an agent_status event:
  { agent: "developer", status: "error" }
When chatStore.setAgentStatus is called
Then AgentStatusIndicator MUST be visible
And MUST show text "developer encountered an error"
And MUST have a red left border and red icon
```

#### Scenario: Status indicator hides on "complete"

```
Given AgentStatusIndicator is visible showing "working"
When the WebSocket receives agent_status { status: "complete" }
And chatStore.setAgentStatus sets orchestratorStatus to "complete"
Then AgentStatusIndicator MUST transition to hidden (isActive=false)
And the transition MUST use the existing slide-fade animation
```

#### Scenario: Status indicator hides on stream_end

```
Given AgentStatusIndicator is visible showing "working"
When the WebSocket receives a stream_end message
And chatStore.clearAgentStatus is called
Then AgentStatusIndicator MUST transition to hidden
And orchestratorStatus MUST be reset to "idle"
```

---

### 2.2 Content Segments Rendered with Agent Attribution

#### REQ-VIS-CS-001: MessageBubble renders segments with specialist badges

The `MessageBubble.vue` component MUST continue to render segmented content when `msg.segments` has entries:

1. Each segment MUST display a `SpecialistBadge` with the segment's `agent` name.
2. Each segment's content MUST be rendered via `MarkdownRenderer`.
3. Segments MUST be separated by a subtle border (`border-makoclaw-border/30`).
4. The last segment MUST NOT have a bottom border.

This behavior already exists and MUST NOT regress.

#### Scenario: Multi-agent response with two segments

```
Given a message has segments:
  [
    { agent: "developer", content: "Here is the code...", segmentId: "seg_1" },
    { agent: "testing", content: "Here are the tests...", segmentId: "seg_2" }
  ]
When MessageBubble renders this message
Then it MUST show two segments
And each MUST have a SpecialistBadge header ("developer", "testing")
And the first segment MUST have a bottom border separator
And the second segment MUST NOT have a bottom border separator
```

#### Scenario: Message with no segments falls back to unified content

```
Given a message has empty segments array and content "Hello, how can I help?"
When MessageBubble renders this message
Then it MUST render the unified content via MarkdownRenderer (or streaming cursor if streaming)
And MUST NOT display any segment headers or specialist badges in the content area
```

#### Scenario: Streaming message receives content_segment mid-stream

```
Given a message is currently streaming (msg.streaming=true)
And the WebSocket receives a content_segment event:
  { agent: "developer", content: "Partial result...", segment_id: "seg_1" }
When chatStore.addContentSegment is called
Then the streaming message's segments array MUST contain the new segment
And "developer" MUST be added to the message's agents array (if not already present)
```

---

### 2.3 Dynamic Specialist Badges

#### REQ-VIS-DB-001: Hash-based badge generation for custom specialists

The `SpecialistBadge.vue` component MUST be updated to support arbitrary specialist names, not just the current hardcoded six types (developer, documentation, testing, devops, analyst, researcher).

The badge system MUST:
1. Retain the existing six hardcoded specialist configurations (icons and colors) for backward compatibility.
2. For any specialist name NOT in the hardcoded map, generate a deterministic badge using a hash of the name.
3. The hash-based badge MUST produce:
   - A consistent HSL color derived from the name hash (hue = hash % 360, saturation ~60%, lightness ~45% for text, ~95% for background in dark mode compatible format).
   - A generic icon (first letter of the name, or a generic person icon).
   - A capitalized label derived from the specialist name.
4. The same specialist name MUST always produce the same color across sessions and page reloads.

#### Scenario: Known specialist type uses hardcoded badge

```
Given a SpecialistBadge with name="developer"
When the component renders
Then it MUST use the hardcoded blue color (text-blue-400, bg-blue-500/10)
And MUST use the code icon (<path ... d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />)
```

#### Scenario: Custom specialist type uses hash-based badge

```
Given a SpecialistBadge with name="security-auditor"
When the component renders
Then it MUST generate a color from the hash of "security-auditor"
And the color MUST be consistent (same hash always produces same hue)
And MUST display the label "Security-auditor" (or "Security Auditor" if name splitting is supported)
And MUST NOT fall back to generic gray (text-makoclaw-text-secondary)
```

#### Scenario: Two different custom specialist names produce different colors

```
Given SpecialistBadge rendered for "security-auditor" and "performance-optimizer"
When both badges are visible
Then they MUST have different hue values (unless hash collision, which is acceptable but unlikely)
And both MUST be visually distinguishable
```

#### Scenario: Empty or null specialist name

```
Given a SpecialistBadge with name=null or name=""
When the component renders
Then it MUST NOT render (v-if="specialist" guard)
And MUST NOT throw a JavaScript error
```

---

### 2.4 Agent Timeline / History View

#### REQ-VIS-TL-001: Agent history tracks tool call events

The `chatStore.agentHistory` array MUST track not only agent status changes but also tool call events with agent attribution. The store MUST support the following event types:

| Type | Fields |
|------|--------|
| `status` | agent, status, specialistName, reason, timestamp |
| `tool_call` | agent, tool, status, args (optional), timestamp (NEW) |

A new method MUST be added to chatStore:

```javascript
function addAgentToolEvent(agent, toolName, status, timestamp) {
    agentHistory.value.push({
        type: 'tool_call',
        agent,
        tool: toolName,
        status,
        timestamp: timestamp || new Date().toISOString()
    })
}
```

#### Scenario: Agent history captures delegation lifecycle

```
Given the user sends a message that triggers orchestrator delegation
And the orchestrator delegates to "developer"
And "developer" calls tools "read_file" and "write_file"
When the response completes
Then agentHistory MUST contain (in order):
  1. { type: "status", agent: "orchestrator", status: "analyzing" }
  2. { type: "status", agent: "orchestrator", status: "delegating", specialistName: "developer" }
  3. { type: "status", agent: "developer", status: "working" }
  4. { type: "tool_call", agent: "developer", tool: "read_file", status: "started" }
  5. { type: "tool_call", agent: "developer", tool: "read_file", status: "finished" }
  6. { type: "tool_call", agent: "developer", tool: "write_file", status: "started" }
  7. { type: "tool_call", agent: "developer", tool: "write_file", status: "finished" }
  8. { type: "status", agent: "developer", status: "complete" }
```

#### Scenario: Agent history is cleared between messages

```
Given agentHistory has events from a previous message
When clearAgentStatus is called (at stream_end or new message start)
Then agentHistory MUST be reset to an empty array
```

---

#### REQ-VIS-TL-002: Collapsible agent timeline component

The frontend SHOULD include a collapsible timeline component that visualizes the agent orchestration flow. This component:

1. SHOULD render each `agentHistory` entry as a timeline node.
2. SHOULD visually distinguish status events from tool_call events.
3. SHOULD use `SpecialistBadge` to identify agents in each node.
4. SHOULD be collapsed by default to avoid cluttering the chat interface.
5. SHOULD be expandable via user interaction (click/toggle).
6. MAY be placed in the `MessageBubble` component (after the content, before the timestamp) or as a sidebar panel.

**Note:** This requirement is SHOULD-level. The exact component design is deferred to the design artifact.

#### Scenario: Timeline is collapsed by default

```
Given a multi-agent message has agentHistory with 5 events
When the message is rendered
Then the timeline component MUST be collapsed
And SHOULD show a summary indicator (e.g., "3 agents involved" or a small badge row)
```

#### Scenario: Timeline expands on user click

```
Given the timeline component is collapsed
When the user clicks the expand toggle
Then all agentHistory events MUST be visible in chronological order
And each event MUST show its timestamp, agent badge, and event description
```

---

### 2.5 WebSocket Events with Agent Metadata During Streaming

#### REQ-VIS-WS-001: tool_call WebSocket messages include agent field

The WebSocket `tool_call` message type MUST be extended to include an optional `agent` field:

```json
{
    "type": "tool_call",
    "name": "read_file",
    "args": {"path": "/etc/config"},
    "result": "",
    "status": "started",
    "agent": "developer"
}
```

The `agent` field MUST be:
- Present when the tool call originates from a specialist agent.
- Absent (omitted from JSON) when the tool call originates from the primary agent (non-orchestrated).

This is backward-compatible: existing frontends that do not handle the `agent` field MUST continue to work.

#### Scenario: Specialist tool call includes agent in WebSocket message

```
Given specialist "developer" calls tool "exec" with args {"command": "go test"}
And the OnTool callback wraps the event with Agent="developer"
When the WebSocket handler serializes the tool_call event
Then the JSON message MUST include "agent": "developer"
```

#### Scenario: Primary agent tool call omits agent field

```
Given the primary agent (non-orchestrated) calls tool "web_search"
And no agent attribution wrapper is applied
When the WebSocket handler serializes the tool_call event
Then the JSON message MUST NOT include an "agent" key
And the message MUST look identical to the current format
```

---

#### REQ-VIS-WS-002: stream events include agent context during specialist streaming

During specialist streaming, the system SHOULD provide agent context so the frontend knows which agent is producing tokens. This MAY be achieved by:

(a) Emitting an `agent_status` event with `status: "working"` before specialist streaming begins (already specified in REQ-ORCH-AS-001), AND/OR
(b) Adding an optional `agent` field to `stream` WebSocket messages.

If approach (b) is used, the `stream` message format becomes:

```json
{
    "type": "stream",
    "content": "token",
    "agent": "developer"
}
```

**Decision:** Approach (a) is REQUIRED (specialist "working" status before tokens). Approach (b) is OPTIONAL and MAY be implemented if the frontend needs per-token attribution. The design artifact SHALL specify which approach is used.

#### Scenario: Frontend knows which agent is streaming via status event

```
Given the orchestrator delegates to "developer"
And an agent_status event { agent: "developer", status: "working" } is emitted
When stream tokens begin arriving
Then the frontend MUST know that tokens belong to "developer"
Because chatStore.currentAgent is set to "developer" from the prior status event
```

---

#### REQ-VIS-WS-003: stream_end includes agents array (existing, preserved)

The existing behavior where `stream_end` includes an `agents` array MUST be preserved:

```json
{
    "type": "stream_end",
    "content": "final response",
    "agents": ["orchestrator", "developer"]
}
```

#### Scenario: stream_end with orchestrator and specialist

```
Given the orchestrator delegated to "developer" and "testing"
When stream_end is sent
Then the agents array MUST include ["orchestrator", "developer", "testing"]
And the order SHOULD reflect the order agents were involved
```

#### Scenario: stream_end without orchestration

```
Given a simple non-orchestrated response
When stream_end is sent
Then the agents array MUST either be absent or contain only the primary agent identifier
```

---

## 3. Domain: agent-token-optimization

This domain covers changes to reduce token consumption in multi-agent conversations.

### 3.1 Specialist Context Trimming

#### REQ-OPT-CT-001: Specialists skip session history by default

When a specialist processes a delegated task via `ProcessDirectWithCallbacks` with `SkipHistory=true`:

1. The system MUST NOT load prior session history for the specialist's session key.
2. The LLM request MUST contain only:
   - The specialist's system prompt (identity + agent-specific prompt from `SetAgentSystemPrompt`)
   - The current task message
3. Bootstrap files (AGENTS.md, SOUL.md, USER.md, IDENTITY.md) MUST already be skipped due to `lightweightMode=true` (existing behavior, preserved).
4. Memory context MUST already be skipped due to `lightweightMode=true` (existing behavior, preserved).

#### Scenario: Specialist receives minimal context

```
Given specialist "developer" with lightweightMode=true and SkipHistory=true
And the specialist's session "specialist_developer" has 15 prior messages
When ProcessDirectWithCallbacks processes the task
Then the LLM request messages MUST contain exactly:
  - 1 system message (specialist system prompt with identity + custom prompt)
  - 1 user message (the delegated task)
And MUST NOT contain any of the 15 prior session messages
```

#### Scenario: Specialist without SkipHistory loads full session

```
Given specialist "developer" with SkipHistory=false (or DelegationCallbacks without SkipHistory)
When ProcessDirectWithCallbacks processes the task
Then session history MUST be loaded normally
And the LLM request MUST include prior messages from the specialist's session
```

#### Scenario: Specialist identity header is trimmed in lightweight mode

```
Given specialist "developer" in lightweightMode=true
When ContextBuilder.BuildSystemPrompt() is called
Then the system prompt MUST NOT include bootstrap files (AGENTS.md, SOUL.md, USER.md, IDENTITY.md)
And MUST NOT include memory context
And MUST include the specialist's custom prompt (from SetAgentSystemPrompt)
And SHOULD include minimal runtime info (OS, time) for tool context
```

---

### 3.2 Parallel Delegation Support

#### REQ-OPT-PD-001: Parallel delegation tool

The system MUST add a new tool `delegate_to_specialists` (plural) alongside the existing `delegate_to_specialist` (singular):

```go
type ParallelDelegationTool struct {
    orchestrator *OrchestratorAgent
}

func (pdt *ParallelDelegationTool) Name() string { return "delegate_to_specialists" }
```

The tool MUST accept an array of delegation requests:

```json
{
    "delegations": [
        { "specialist_name": "developer", "task": "implement the API", "context": "..." },
        { "specialist_name": "testing", "task": "write tests for the API", "context": "..." }
    ]
}
```

The tool MUST:
1. Validate all specialist names before starting any delegation.
2. Launch each delegation in a separate goroutine.
3. Use `sync.WaitGroup` to wait for all delegations to complete.
4. Collect results (or errors) from each specialist.
5. Return a JSON-formatted aggregated result containing all specialist results.
6. Respect a configurable maximum concurrent delegations limit (`max_parallel_delegations`, default 3).
7. If more delegations are requested than the limit, excess delegations MUST be queued and executed as slots free up.

#### Scenario: Parallel delegation to two specialists

```
Given specialist "developer" and "testing" are both registered
And the orchestrator calls delegate_to_specialists with both
When both specialists execute concurrently
Then both MUST run in separate goroutines
And the tool MUST wait for both to complete
And the returned result MUST contain results from both:
  { "results": [
      { "specialist_name": "developer", "result": "...", "success": true },
      { "specialist_name": "testing", "result": "...", "success": true }
  ]}
```

#### Scenario: One specialist fails during parallel delegation

```
Given delegate_to_specialists is called for "developer" and "nonexistent"
When "nonexistent" is not found in the registry
Then the tool MUST still execute "developer" delegation
And the returned result MUST contain:
  { "results": [
      { "specialist_name": "developer", "result": "...", "success": true },
      { "specialist_name": "nonexistent", "result": "", "success": false, "error": "specialist 'nonexistent' not found" }
  ]}
And the tool MUST NOT return an error (partial success is valid)
```

#### Scenario: Parallel delegation exceeds max_parallel_delegations

```
Given max_parallel_delegations is set to 2
And delegate_to_specialists is called with 4 delegations
When execution begins
Then at most 2 specialists MUST run concurrently at any time
And the remaining 2 MUST wait until a slot is available
And all 4 results MUST be returned in the final aggregated result
```

#### Scenario: Parallel delegation with streaming callbacks

```
Given delegate_to_specialists is called with 2 delegations
And both specialists stream tokens via ProcessWithSpecialityStream
When tokens from both specialists arrive concurrently
Then all tokens MUST be forwarded to the WebSocket client
And agent_status events MUST correctly attribute which specialist is working
And content_segment events MUST correctly attribute each specialist's result
And the WebSocket write mutex (wsMu) MUST prevent interleaved JSON writes
```

#### Scenario: Parallel delegation with context cancellation

```
Given delegate_to_specialists is executing 3 specialists concurrently
When the parent context is canceled (e.g., user disconnects)
Then all 3 specialist goroutines MUST be notified via context cancellation
And the tool MUST return promptly (not wait for all specialists to time out)
And MUST NOT leak goroutines
```

#### Scenario: Empty delegations array

```
Given delegate_to_specialists is called with an empty delegations array
When the tool validates the input
Then it MUST return an error: "at least one delegation is required"
And MUST NOT launch any goroutines
```

---

#### REQ-OPT-PD-002: Parallel delegation configuration

The `OrchestratorConfig` in `pkg/config/config.go` MUST be extended with:

```go
type OrchestratorConfig struct {
    // ... existing fields ...
    MaxParallelDelegations int `json:"max_parallel_delegations,omitempty"` // Default: 3
}
```

The default value MUST be 3 if not specified in config. The value MUST be at least 1.

#### Scenario: Config with max_parallel_delegations=5

```
Given config.json contains: { "agents": { "orchestrator": { "max_parallel_delegations": 5 } } }
When the orchestrator is initialized
Then delegate_to_specialists MUST allow up to 5 concurrent delegations
```

#### Scenario: Config without max_parallel_delegations uses default

```
Given config.json does not specify max_parallel_delegations
When the orchestrator is initialized
Then the default limit MUST be 3
```

---

### 3.3 Per-Specialist Token Budget Tracking

#### REQ-OPT-TB-001: DelegationResult includes token usage

The `DelegationResult` struct MUST be extended with token usage information:

```go
type DelegationResult struct {
    SpecialistName string `json:"specialist_name"`
    Result         string `json:"result"`
    Success        bool   `json:"success"`
    Error          string `json:"error,omitempty"`
    TokensUsed     int    `json:"tokens_used,omitempty"`     // NEW: Total tokens (prompt + completion)
    PromptTokens   int    `json:"prompt_tokens,omitempty"`   // NEW: Prompt/input tokens
    CompletionTokens int  `json:"completion_tokens,omitempty"` // NEW: Completion/output tokens
    DurationMs     int64  `json:"duration_ms,omitempty"`     // NEW: Execution duration in milliseconds
}
```

Token counts MUST be sourced from the LLM provider's response metadata (`LLMResponse.Usage` or equivalent). If the provider does not report token usage, the fields MUST be zero (not estimated).

#### Scenario: Specialist delegation result includes token count

```
Given specialist "developer" processes a task
And the LLM provider reports usage: prompt_tokens=500, completion_tokens=200
When the DelegationResult is constructed
Then TokensUsed MUST be 700
And PromptTokens MUST be 500
And CompletionTokens MUST be 200
```

#### Scenario: Provider does not report token usage

```
Given specialist "analyst" uses an Ollama provider that does not report usage
When the DelegationResult is constructed
Then TokensUsed MUST be 0
And PromptTokens MUST be 0
And CompletionTokens MUST be 0
And the result MUST still be returned successfully
```

#### Scenario: DelegationResult includes duration

```
Given specialist "developer" takes 3.5 seconds to complete
When the DelegationResult is constructed
Then DurationMs MUST be approximately 3500 (within reasonable clock precision)
```

---

#### REQ-OPT-TB-002: Token usage logging for specialists

The system MUST log token usage for each specialist delegation at INFO level:

```go
logger.InfoCF("agent", "Specialist delegation completed", map[string]interface{}{
    "specialist":       specialistName,
    "tokens_used":      result.TokensUsed,
    "prompt_tokens":    result.PromptTokens,
    "completion_tokens": result.CompletionTokens,
    "duration_ms":      result.DurationMs,
    "success":          result.Success,
})
```

#### Scenario: Token usage is logged on successful delegation

```
Given specialist "developer" completes with 700 tokens used
When the delegation result is returned to processSpecialistTask
Then an INFO log MUST be written with component "agent"
And the log MUST include specialist name, token counts, and duration
```

---

#### REQ-OPT-TB-003: Token usage in WebSocket response (informational)

The `content_segment` WebSocket message MAY be extended with token usage information:

```json
{
    "type": "content_segment",
    "agent": "developer",
    "content": "...",
    "segment_id": "seg_developer_1234",
    "timestamp": "2026-02-27T10:00:00Z",
    "tokens_used": 700,
    "duration_ms": 3500
}
```

This is OPTIONAL in v1. The fields MUST be `omitempty` if included.

#### Scenario: Content segment includes token usage when available

```
Given specialist "developer" reports 700 tokens used
When the content_segment event is emitted
Then the WebSocket message MAY include "tokens_used": 700
And the frontend MAY display this information in the timeline view
```

---

## 4. Cross-Cutting Concerns

### Security

#### REQ-CC-SEC-001: No security boundary changes

This change MUST NOT alter any security boundaries:
1. Tool permission filtering MUST still be applied per-specialist based on their `allowedTools` configuration.
2. Audit logging (`SQLiteAuditLogger`) MUST continue to log tool executions, including those by specialists.
3. Workspace isolation MUST be preserved -- specialist execution MUST NOT access files outside the user's workspace.
4. The new `Agent` field on `ToolEvent` and WebSocket messages is informational only and MUST NOT be used for access control decisions.

#### Scenario: Specialist tool calls are audit-logged

```
Given specialist "developer" executes tool "exec" with command "ls"
And audit logging is enabled
When the tool execution completes
Then the audit log MUST contain an entry for the tool execution
And the entry SHOULD include the specialist name as metadata
```

---

### Backward Compatibility

#### REQ-CC-BC-001: Existing API contracts preserved

1. All existing `Process*` methods on `AgentLoop` MUST remain unchanged in signature and behavior.
2. All existing WebSocket message types MUST retain their current fields.
3. New fields added to WebSocket messages MUST be optional (`omitempty`).
4. The existing `ProcessWithSpeciality` method MUST remain as a non-streaming fallback.
5. Non-orchestrated agent loops (no orchestrator configured) MUST be completely unaffected.
6. The `DelegationTool` (`delegate_to_specialist`) MUST continue to work as today for the non-streaming path.

#### Scenario: Non-orchestrated agent loop is unaffected

```
Given an agent configuration with orchestrator.enabled=false
When the user sends a message via WebSocket
Then the message MUST be processed by the primary agent loop as today
And no agent_status events MUST be emitted (except the fallback "working" event)
And no content_segment events MUST be emitted
And stream_end MUST NOT include an agents array (or it may include only the primary agent)
```

---

### Multi-User

#### REQ-CC-MU-001: Per-request callback isolation

All callback propagation MUST be per-request:
1. Callbacks are injected into the request context by the WebSocket handler.
2. Each WebSocket connection has its own callbacks writing to its own connection.
3. Specialist delegation in one user's request MUST NOT affect another user's request.
4. The per-call tool registry copy (REQ-ORCH-CC-001) ensures that concurrent delegations from different users do not interfere.

#### Scenario: Two users concurrently trigger orchestrator delegations

```
Given User A and User B each send a message via separate WebSocket connections
And both messages trigger orchestrator delegation to specialist "developer"
When both delegations execute concurrently
Then User A MUST only receive events from their delegation
And User B MUST only receive events from their delegation
And the specialist MUST handle both calls safely (per-call tool registry copy)
And no data races MUST occur
```

---

### Performance

#### REQ-CC-PERF-001: Callback wrapper overhead

The callback wrapper functions (wrapTokenCallback, wrapToolCallback, etc.) MUST NOT introduce measurable latency to the streaming path. Specifically:
1. Wrapper overhead MUST be less than 1 microsecond per callback invocation.
2. Wrappers MUST NOT allocate memory on the heap for each token (use closure captures, not map lookups).
3. The WebSocket write path (`wsMu.Lock()` -> `conn.WriteJSON()` -> `wsMu.Unlock()`) is the dominant latency factor and MUST NOT be made worse.

---

## 5. Glossary

| Term | Definition |
|------|-----------|
| **AgentLoop** | The core message processing loop (`pkg/agent/loop.go`) that handles LLM calls, tool execution, and response generation. |
| **Callback** | A function passed as a parameter to be invoked when a specific event occurs (token streamed, tool called, status changed). |
| **Content Segment** | A piece of content attributed to a specific agent, emitted as a `content_segment` WebSocket event. |
| **Delegation** | The act of the orchestrator assigning a task to a specialist agent via the `delegate_to_specialist` tool. |
| **DelegationCallbacks** | A struct bundling all callback types needed for specialist delegation (OnToken, OnTool, OnAgentStatus, OnContentSegment, SkipHistory). |
| **Orchestrator** | A special agent that analyzes tasks and delegates to specialists rather than executing them directly. |
| **processOptions** | Internal struct configuring how a message is processed (session, channel, callbacks, model override). |
| **Specialist** | A domain-specific agent with filtered tools, custom prompt, and potentially a different LLM provider/model. |
| **SpecialistRegistry** | A thread-safe registry managing all configured specialist agents. |
| **StreamCallback** | A function `func(token string) error` called for each streamed token from the LLM. |
| **ToolCallback** | A function `func(ev ToolEvent) error` called when a tool starts, finishes, or errors. |
| **ToolEvent** | A struct representing a tool execution event (name, args, result, status, agent). |
| **ToolRegistry** | A collection of registered tools (`*tools.ToolRegistry`) available to an agent loop. |

---

## Appendix A: Requirement Traceability Matrix

| Requirement ID | Phase | Domain | Priority | Depends On |
|---------------|-------|--------|----------|------------|
| REQ-ORCH-CB-001 | 1 | agent-orchestration | MUST | -- |
| REQ-ORCH-CB-002 | 1 | agent-orchestration | MUST | REQ-ORCH-CB-001 |
| REQ-ORCH-CB-003 | 1 | agent-orchestration | MUST | REQ-ORCH-CB-002, REQ-ORCH-CB-004, REQ-ORCH-SS-001 |
| REQ-ORCH-CB-004 | 1 | agent-orchestration | MUST | -- |
| REQ-ORCH-SS-001 | 1 | agent-orchestration | MUST | REQ-ORCH-CB-002, REQ-ORCH-CC-001 |
| REQ-ORCH-CC-001 | 1 | agent-orchestration | MUST | -- |
| REQ-ORCH-AS-001 | 1 | agent-orchestration | MUST | REQ-ORCH-CB-003 |
| REQ-ORCH-AS-002 | 1 | agent-orchestration | MAY | REQ-ORCH-AS-001 |
| REQ-ORCH-CS-001 | 1 | agent-orchestration | MUST | REQ-ORCH-SS-001, REQ-ORCH-CS-002 |
| REQ-ORCH-CS-002 | 1 | agent-orchestration | MUST | -- |
| REQ-ORCH-CS-003 | 1 | agent-orchestration | MUST | -- (existing, preserved) |
| REQ-VIS-SI-001 | 2 | agent-visibility | MUST | REQ-ORCH-AS-001 |
| REQ-VIS-CS-001 | 2 | agent-visibility | MUST | REQ-ORCH-CS-003 |
| REQ-VIS-DB-001 | 2 | agent-visibility | MUST | -- |
| REQ-VIS-TL-001 | 2 | agent-visibility | MUST | REQ-VIS-WS-001 |
| REQ-VIS-TL-002 | 2 | agent-visibility | SHOULD | REQ-VIS-TL-001 |
| REQ-VIS-WS-001 | 2 | agent-visibility | MUST | REQ-ORCH-CS-002 |
| REQ-VIS-WS-002 | 2 | agent-visibility | SHOULD | REQ-ORCH-AS-001 |
| REQ-VIS-WS-003 | 2 | agent-visibility | MUST | -- (existing, preserved) |
| REQ-OPT-CT-001 | 3 | agent-token-optimization | MUST | REQ-ORCH-CB-001, REQ-ORCH-CB-002 |
| REQ-OPT-PD-001 | 3 | agent-token-optimization | MUST | REQ-ORCH-SS-001, REQ-ORCH-CC-001 |
| REQ-OPT-PD-002 | 3 | agent-token-optimization | MUST | REQ-OPT-PD-001 |
| REQ-OPT-TB-001 | 3 | agent-token-optimization | MUST | -- |
| REQ-OPT-TB-002 | 3 | agent-token-optimization | MUST | REQ-OPT-TB-001 |
| REQ-OPT-TB-003 | 3 | agent-token-optimization | MAY | REQ-OPT-TB-001 |
| REQ-CC-SEC-001 | All | cross-cutting | MUST | -- |
| REQ-CC-BC-001 | All | cross-cutting | MUST | -- |
| REQ-CC-MU-001 | All | cross-cutting | MUST | REQ-ORCH-CC-001 |
| REQ-CC-PERF-001 | 1 | cross-cutting | MUST | REQ-ORCH-CS-001 |

---

## Appendix B: File Impact Summary

| File | Domain | Changes |
|------|--------|---------|
| `pkg/agent/loop.go` | agent-orchestration | DelegationCallbacks struct, ProcessDirectWithCallbacks, ToolEvent.Agent field, StreamCallback/ToolCallback context keys |
| `pkg/agent/specialist.go` | agent-orchestration | ProcessWithSpecialityStream, per-call tool registry (remove mutex swap), ProcessWithSpeciality updated for consistency |
| `pkg/agent/orchestrator.go` | agent-orchestration, agent-token-optimization | processSpecialistTask callback extraction, ParallelDelegationTool, DelegationResult token fields, token logging |
| `pkg/config/config.go` | agent-token-optimization | OrchestratorConfig.MaxParallelDelegations |
| `pkg/web/server.go` | agent-visibility | tool_call WebSocket message with Agent field, optional stream agent field |
| `pkg/web/frontend/src/components/Chat/SpecialistBadge.vue` | agent-visibility | Hash-based badge generation, fallback for unknown types |
| `pkg/web/frontend/src/components/Chat/AgentStatusIndicator.vue` | agent-visibility | New statuses (tool_call, error, timeout) |
| `pkg/web/frontend/src/stores/chatStore.js` | agent-visibility | addAgentToolEvent, enhanced agentHistory tracking |
| `pkg/web/frontend/src/views/ChatView.vue` | agent-visibility | Handle tool_call agent field, agent timeline integration |
| `pkg/web/frontend/src/components/MessageBubble.vue` | agent-visibility | Optional timeline component integration |
