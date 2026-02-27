# Implementation Tasks: Multi-Agent Orchestration Visibility

**Change ID:** multi-agent-orchestration-visibility
**Status:** Ready for Implementation
**Author:** SDD Sub-Agent (tasks phase)
**Created:** 2026-02-27
**Based on:** [exploration.md](exploration.md), [proposal.md](proposal.md), [specs.md](specs.md), [design.md](design.md)

---

## Task Overview

| Phase | Tasks | Effort | Priority | Dependencies |
|-------|-------|--------|----------|--------------|
| **Phase 1: Fix the Pipeline** | 1.1 -- 1.9 | ~2-3 sessions | Critical | None |
| **Phase 2: Enrich Visibility** | 2.1 -- 2.7 | ~2-3 sessions | High | Phase 1 |
| **Phase 3: Token Optimization** | 3.1 -- 3.6 | ~2 sessions | Medium | Phase 1 |
| **Cross-Cutting** | 4.1 -- 4.4 | Continuous | High | All phases |

**Dependency graph:**

```
Phase 1: 1.1 ─> 1.2 ─> 1.3 ─> 1.5 ─> 1.6 ─> 1.7 ─> 1.8 ─> 1.9
               1.4 ──────────/         \──── 1.7
                                        \─── 1.8 ─> 1.9

Phase 2: 2.1 (independent)
         2.2 ─> 2.3
         2.4 ─> 2.5
         2.6 (depends on 1.7)
         2.7 (depends on 1.8)

Phase 3: 3.1 (depends on 1.2)
         3.2 ─> 3.3 (depend on 1.3, 1.5)
         3.4 ─> 3.5 ─> 3.6
```

---

## Phase 1: Fix the Pipeline (Critical, Backend)

The core objective of Phase 1 is to thread streaming callbacks from the WebSocket handler through the entire orchestrator-to-specialist delegation chain so specialist work is visible to the user in real time.

---

### Task 1.1: Define DelegationCallbacks struct and ToolEvent.Agent field

**Priority:** Critical
**Depends on:** None
**Blocks:** 1.2, 1.5, 1.6

**Description:**

Add the `DelegationCallbacks` struct and extend `ToolEvent` with an `Agent` field in `pkg/agent/loop.go`. These are the foundational types used by all subsequent tasks.

**Requirements:** REQ-ORCH-CB-001, REQ-ORCH-CS-002

**Changes:**

1. Add `DelegationCallbacks` struct with fields:
   - `OnToken StreamCallback`
   - `OnTool ToolCallback`
   - `OnAgentStatus AgentStatusCallback`
   - `OnContentSegment ContentSegmentCallback`
   - `SkipHistory bool`
2. Add `Agent string \`json:"agent,omitempty"\`` field to the existing `ToolEvent` struct.
3. Add `TokenUsage` struct with `InputTokens`, `OutputTokens`, `TotalTokens int64` fields (needed by Phase 3 but cheap to add now).

**Files touched:**
- `pkg/agent/loop.go` -- add structs and extend ToolEvent

**Acceptance criteria:**
- `DelegationCallbacks` is exported and usable from `orchestrator.go` and `specialist.go`.
- Existing code constructing `ToolEvent` without `Agent` compiles without changes.
- `ToolEvent` JSON serialization omits `agent` when empty (verified by a quick test).

**Estimated effort:** 15 minutes

---

### Task 1.2: Add ProcessDirectWithCallbacks method on AgentLoop

**Priority:** Critical
**Depends on:** 1.1
**Blocks:** 1.5, 1.9, 3.1

**Description:**

Add a new `ProcessDirectWithCallbacks` method to `AgentLoop` that accepts a `DelegationCallbacks` struct and an optional `*tools.ToolRegistry` override. This is the entry point for specialist delegation with streaming visibility.

**Requirements:** REQ-ORCH-CB-002

**Changes:**

1. Add `ProcessDirectWithCallbacks(ctx, content, sessionKey string, callbacks DelegationCallbacks, toolsOverride *tools.ToolRegistry) (string, *TokenUsage, error)` method.
2. Method constructs `bus.InboundMessage` with `Channel: "specialist"`, `SenderID: "orchestrator"`, `ChatID: "delegation"`.
3. Builds `processOptions` with all callback fields populated from `DelegationCallbacks`.
4. If `callbacks.OnToken != nil` and provider implements `StreamingLLMProvider`, route to new `runAgentLoopStreamWithTools` (Task 1.3). Otherwise route to new `runAgentLoopWithTools` (Task 1.3).
5. Return accumulated response string, nil TokenUsage (Phase 3 populates this), and error.

**Files touched:**
- `pkg/agent/loop.go` -- add method

**Acceptance criteria:**
- Method compiles and is callable from specialist.go.
- When all callbacks are nil, behaves equivalently to `ProcessDirect` (fallback to non-streaming).
- Does NOT modify any existing `Process*` methods.
- Context cancellation propagates correctly (method returns `context.Canceled`).

**Estimated effort:** 30 minutes

---

### Task 1.3: Refactor runAgentLoopStream to accept tools override and skipHistory

**Priority:** Critical
**Depends on:** 1.2
**Blocks:** 1.5, 3.2

**Description:**

Refactor the existing `runAgentLoopStream` (and `runAgentLoop`) to support a tool registry override and a `skipHistory` flag, without breaking existing callers. This is done by extracting the body into an `Impl` method and adding thin wrappers.

**Requirements:** ADR-2 (design.md section 3.4)

**Changes:**

1. Extract body of `runLLMIterationStream` into `runLLMIterationStreamImpl(ctx, messages, opts, onToken, effectiveTools *tools.ToolRegistry) (string, int, error)`. The original `runLLMIterationStream` becomes a thin wrapper passing `al.tools`.
2. Add `runLLMIterationStreamWithTools` that passes the provided tools parameter.
3. Add `runAgentLoopStreamWithTools(ctx, opts, onToken, toolsOverride, skipHistory) (string, error)`. When `skipHistory` is true, skip session history loading. When `toolsOverride` is non-nil, use it instead of `al.tools`.
4. Add `runAgentLoopWithTools(ctx, opts, toolsOverride, skipHistory) (string, error)` for the non-streaming fallback path.
5. Ensure existing callers of `runAgentLoopStream` and `runAgentLoop` are unaffected.

**Files touched:**
- `pkg/agent/loop.go` -- refactor iteration methods, add new wrapper methods

**Acceptance criteria:**
- Existing `runAgentLoopStream` and `runAgentLoop` continue to work identically (no call-site changes).
- `runAgentLoopStreamWithTools` correctly uses the provided tool registry for `GetDefinitions()` and `ExecuteWithContext()`.
- When `skipHistory=true`, the LLM request contains only the system prompt and current user message.
- `go test ./pkg/agent/...` passes with no regressions.

**Estimated effort:** 45 minutes

---

### Task 1.4: Fix concurrency bug -- per-call tool registry in ProcessWithSpecialityStream

**Priority:** Critical
**Depends on:** None (can proceed in parallel with 1.2/1.3)
**Blocks:** 1.5

**Description:**

Implement the new `ProcessWithSpecialityStream` method on `SpecialistAgent` that creates a per-call tool registry copy instead of swapping the shared `sa.tools` field. This eliminates the concurrency bug documented in the exploration (Appendix B).

**Requirements:** REQ-ORCH-CC-001, REQ-ORCH-SS-001

**Changes:**

1. Add `ProcessWithSpecialityStream(ctx, userMessage string, callbacks DelegationCallbacks) (string, *TokenUsage, error)` to `SpecialistAgent`.
2. Inside, call `sa.ToolFilter()` to create a **new** per-call `*tools.ToolRegistry`.
3. Call `sa.wrapCallbacksWithAttribution(callbacks)` (Task 1.6) -- if not yet available, pass callbacks through unwrapped and add wrapping in Task 1.6.
4. Set `callbacks.SkipHistory = true` (specialists do not load conversation history).
5. Call `sa.ProcessDirectWithCallbacks(ctx, fullMessage, sessionKey, wrappedCallbacks, filteredTools)`.
6. Do NOT touch `sa.tools` -- the shared field remains immutable.
7. Add deprecation comment on existing `ProcessWithSpeciality` noting that `ProcessWithSpecialityStream` is preferred for concurrent use.

**Files touched:**
- `pkg/agent/specialist.go` -- add method, deprecation comment

**Acceptance criteria:**
- `ProcessWithSpecialityStream` never modifies `sa.tools`.
- Two concurrent calls to `ProcessWithSpecialityStream` on the same specialist each get independent tool registries.
- `go test -race ./pkg/agent/...` passes with concurrent specialist test.
- The `processMu` mutex is no longer needed for the new path (can be left for backward-compat on old `ProcessWithSpeciality`).

**Estimated effort:** 30 minutes

---

### Task 1.5: Implement extractDelegationCallbacks in orchestrator

**Priority:** Critical
**Depends on:** 1.1, 1.3, 1.4
**Blocks:** 1.7, 1.8

**Description:**

Modify `processSpecialistTask` in the orchestrator to extract all available callbacks from `ctx`, bundle them into `DelegationCallbacks`, and route to `ProcessWithSpecialityStream` when callbacks are present.

**Requirements:** REQ-ORCH-CB-003, REQ-ORCH-CB-004

**Changes:**

1. Add context keys and helper functions for `StreamCallback` and `ToolCallback`:
   - `type streamCallbackKey struct{}`
   - `type toolCallbackKey struct{}`
   - `ContextWithStreamCallback(ctx, callback) context.Context`
   - `streamCallbackFromCtx(ctx) StreamCallback`
   - `ContextWithToolCallback(ctx, callback) context.Context`
   - `toolCallbackFromCtx(ctx) ToolCallback`
2. Add `extractDelegationCallbacks(ctx) DelegationCallbacks` that retrieves all four callback types from ctx.
3. Modify `processSpecialistTask` to:
   - Call `extractDelegationCallbacks(ctx)`.
   - If `callbacks.OnToken != nil`, call `specialist.ProcessWithSpecialityStream(ctx, task, callbacks)`.
   - Otherwise, fall back to `specialist.ProcessWithSpeciality(ctx, task)` (backward-compatible).
4. Add `delegationResultInternal` struct for goroutine result passing.
5. In `runAgentLoopStreamWithTools` (from Task 1.3), inject `OnToken` and `OnTool` into ctx before the tool-execution loop:
   ```go
   if opts.OnToken != nil { ctx = ContextWithStreamCallback(ctx, opts.OnToken) }
   if opts.OnTool != nil { ctx = ContextWithToolCallback(ctx, opts.OnTool) }
   ```

**Files touched:**
- `pkg/agent/orchestrator.go` -- add context keys, `extractDelegationCallbacks`, modify `processSpecialistTask`
- `pkg/agent/loop.go` -- inject stream/tool callbacks into ctx in `runAgentLoopStreamWithTools`

**Acceptance criteria:**
- WebSocket streaming path: all four callbacks are extracted and forwarded to specialist.
- CLI/non-streaming path: all callbacks are nil, specialist falls back to `ProcessWithSpeciality`.
- Existing `emitAgentStatus` and `emitContentSegment` calls continue to work via ctx.
- Content segment is still emitted with specialist result at the end of `processSpecialistTask`.

**Estimated effort:** 45 minutes

---

### Task 1.6: Implement callback wrappers with agent attribution

**Priority:** Critical
**Depends on:** 1.1
**Blocks:** 1.7

**Description:**

Implement the `wrapCallbacksWithAttribution` method on `SpecialistAgent` that wraps `OnTool` to set the `Agent` field on `ToolEvent` events, and leaves `OnToken`, `OnAgentStatus`, and `OnContentSegment` as pass-through.

**Requirements:** REQ-ORCH-CS-001

**Changes:**

1. Add `wrapCallbacksWithAttribution(callbacks DelegationCallbacks) DelegationCallbacks` on `SpecialistAgent`.
2. `OnToken` is passed through unchanged (token attribution is handled by `agent_status` events).
3. `OnTool` wrapper sets `ev.Agent = sa.name` before forwarding.
4. `OnAgentStatus` is passed through unchanged.
5. `OnContentSegment` is passed through unchanged.
6. Wire this into `ProcessWithSpecialityStream` (Task 1.4).

**Files touched:**
- `pkg/agent/specialist.go` -- add method

**Acceptance criteria:**
- Tool events from specialist include `Agent` field set to specialist name.
- Token content is never modified by the wrapper.
- Nil callbacks are handled gracefully (no panic).
- Method is unit-testable in isolation.

**Estimated effort:** 15 minutes

---

### Task 1.7: Inject StreamCallback and ToolCallback into ctx in WebSocket handler

**Priority:** Critical
**Depends on:** 1.5, 1.6
**Blocks:** 1.9

**Description:**

Modify the WebSocket handler in `server.go` to inject `StreamCallback` and `ToolCallback` into the context alongside the existing `AgentStatusCallback` and `ContentSegmentCallback`. This completes the callback chain from WebSocket to specialist.

**Requirements:** REQ-ORCH-CB-004, design.md section 4.1

**Changes:**

1. In `handleChatWS` (streaming path), after the existing `ctx = agent.ContextWithAgentStatusCallback(...)` and `ctx = agent.ContextWithContentSegmentCallback(...)` lines, add:
   ```go
   ctx = agent.ContextWithStreamCallback(ctx, func(token string) error {
       wsMu.Lock()
       defer wsMu.Unlock()
       return conn.WriteJSON(map[string]interface{}{"type": "stream", "content": token})
   })
   ctx = agent.ContextWithToolCallback(ctx, func(ev agent.ToolEvent) error {
       wsMu.Lock()
       defer wsMu.Unlock()
       msg := map[string]interface{}{
           "type": "tool_call", "name": ev.Name, "args": ev.Args,
           "result": ev.Result, "status": ev.Status,
       }
       if ev.Agent != "" { msg["agent"] = ev.Agent }
       return conn.WriteJSON(msg)
   })
   ```
2. The direct `onToken` and `onTool` lambdas passed to `ProcessDirectWithUserAndModelStream` remain unchanged (they handle the orchestrator's own LLM tokens).

**Files touched:**
- `pkg/web/server.go` -- modify `handleChatWS` streaming path

**Acceptance criteria:**
- WebSocket `tool_call` messages include `agent` field when set.
- WebSocket `tool_call` messages omit `agent` field when empty (backward-compatible).
- Specialist streaming tokens reach the WebSocket client.
- Specialist tool call events reach the WebSocket client with agent attribution.
- The `wsMu` mutex is acquired for all writes (concurrency-safe).

**Estimated effort:** 20 minutes

---

### Task 1.8: Emit granular agent status events during specialist work

**Priority:** High
**Depends on:** 1.5
**Blocks:** 1.9

**Description:**

Ensure that specialist execution emits granular `agent_status` events at the correct points: `"working"` before LLM call, `"streaming"` when tokens begin, `"complete"` when done, `"error"` on failure, and `"timeout"` on timeout.

**Requirements:** REQ-ORCH-AS-001, REQ-ORCH-AS-002

**Changes:**

1. In `ProcessWithSpecialityStream`, emit `agent_status` with `status: "streaming"` and `agent: sa.name` before calling `ProcessDirectWithCallbacks`.
2. Verify that `processSpecialistTask` already emits:
   - `"working"` before specialist call (existing behavior -- confirm).
   - `"complete"` on success (existing behavior -- confirm).
   - `"timeout"` on context deadline exceeded (existing behavior -- confirm).
3. Add `"error"` status emission in `processSpecialistTask` when the specialist returns an error (may already exist -- verify and add if missing).
4. Optionally extend `AgentStatusEvent` with `Metadata map[string]interface{} \`json:"metadata,omitempty"\`` for carrying tool names in `"tool_call"` status events (OPTIONAL per spec, implement if straightforward).

**Files touched:**
- `pkg/agent/specialist.go` -- emit "streaming" status
- `pkg/agent/orchestrator.go` -- verify/add "error" status emission

**Acceptance criteria:**
- Full delegation lifecycle emits: analyzing -> delegating -> working -> streaming -> [tool_call(s)] -> complete.
- Error path emits: analyzing -> delegating -> working -> error.
- Timeout path emits: analyzing -> delegating -> working -> timeout.
- All events have non-zero `Timestamp`.
- Non-WebSocket paths (CLI) emit no events (graceful no-op).

**Estimated effort:** 20 minutes

---

### Task 1.9: Preserve content segment emission and verify end-to-end pipeline

**Priority:** Critical
**Depends on:** 1.7, 1.8
**Blocks:** Phase 2, Phase 3

**Description:**

Verify the complete end-to-end pipeline works: WebSocket message -> orchestrator -> specialist (streaming) -> tokens/tool calls/status events -> content segment -> stream_end with agents. Write integration-level tests.

**Requirements:** REQ-ORCH-CS-003

**Changes:**

1. Verify that `processSpecialistTask` still emits `ContentSegment` with the specialist's final result after `ProcessWithSpecialityStream` completes. This should already be in place -- confirm it is not skipped in the streaming path.
2. Verify `stream_end` includes the correct `agents` array (orchestrator + specialist names).
3. Add unit tests:
   - `TestProcessDirectWithCallbacks` -- verify callbacks are invoked.
   - `TestProcessDirectWithCallbacks_NoCallbacks` -- verify nil callbacks do not panic.
   - `TestProcessDirectWithCallbacks_ToolsOverride` -- verify tool registry override is used.
   - `TestProcessWithSpecialityStream` -- verify callbacks wrapped with attribution.
   - `TestProcessWithSpecialityStream_Concurrent` -- verify no data races with concurrent calls.
   - `TestExtractDelegationCallbacks` -- verify all callbacks extracted from ctx.
   - `TestToolEventAgent` -- verify Agent field serialization.
   - `TestWrapCallbacksWithAttribution` -- verify tool events get agent name.
4. Run `go test -race ./pkg/agent/...` to confirm no data races.

**Files touched:**
- `pkg/agent/loop_test.go` -- add tests
- `pkg/agent/specialist_test.go` -- add tests (create if needed)
- `pkg/agent/orchestrator_test.go` -- add tests (create if needed)

**Acceptance criteria:**
- All tests pass, including under `-race`.
- Content segment is emitted at specialist completion (not lost in streaming path).
- `stream_end` includes correct agents array.
- Non-orchestrated agent loop is completely unaffected (verified by existing tests passing).

**Estimated effort:** 60 minutes

---

## Phase 2: Enrich Visibility (Backend + Frontend)

Phase 2 delivers the user-facing value: real-time agent information rendered in the chat UI. All tasks in Phase 2 depend on Phase 1 being complete.

---

### Task 2.1: Dynamic SpecialistBadge with hash-based colors

**Priority:** High
**Depends on:** None (frontend-only, can start in parallel with Phase 1)
**Blocks:** 2.3

**Description:**

Update `SpecialistBadge.vue` to support arbitrary specialist names via deterministic hash-based color generation, falling back to the existing hardcoded 6-type map for known specialists.

**Requirements:** REQ-VIS-DB-001, design.md section 5.1

**Changes:**

1. Add `hashCode(str)` utility function that produces a deterministic integer from a string.
2. Define `dynamicPalette` array with 12 distinct Tailwind color pairs (text + background).
3. Update the `colors` computed property:
   - Check `knownMap` first (existing 6 types + add "orchestrator").
   - Fall back to `dynamicPalette[hashCode(name) % dynamicPalette.length]`.
4. Update the `icon` computed property:
   - Check `knownIcons` first (existing 6 types + add orchestrator icon).
   - Fall back to a generic person icon SVG path.
5. Ensure empty/null name renders nothing (v-if guard).

**Files touched:**
- `pkg/web/frontend/src/components/Chat/SpecialistBadge.vue`

**Acceptance criteria:**
- Known specialists (developer, documentation, testing, devops, analyst, researcher) render with their existing colors/icons.
- Custom names (e.g., "security-auditor", "performance-optimizer") get unique, consistent colors.
- Same name always produces same color across page reloads.
- Two different custom names produce visually distinguishable badges.
- Empty/null name does not render and does not throw errors.

**Estimated effort:** 25 minutes

---

### Task 2.2: Create AgentTimeline component

**Priority:** Medium
**Depends on:** None (can start in parallel)
**Blocks:** 2.3

**Description:**

Create a new `AgentTimeline.vue` component that displays the orchestration event history as a collapsible vertical timeline. Shows agent status changes and tool calls with timestamps.

**Requirements:** REQ-VIS-TL-002, design.md section 5.2

**Changes:**

1. Create `pkg/web/frontend/src/components/Chat/AgentTimeline.vue`.
2. Props:
   - `events: Array` -- the `agentHistory` array from chatStore.
   - `collapsed: Boolean` -- default `true`.
3. Template:
   - Collapsed state: shows summary "N agents involved" with small badge row.
   - Expanded state: vertical timeline with nodes for each event.
   - Each node shows: `SpecialistBadge`, event description, relative timestamp.
   - Status events: "Analyzing request", "Delegating to {name}", "{name} is working", etc.
   - Tool call events: "{name} > {tool}" with started/finished status.
4. Styling: glass morphism consistent with the project's UI polish (subtle backdrop-blur, border, rounded).
5. Transition: smooth expand/collapse animation.

**Files touched:**
- `pkg/web/frontend/src/components/Chat/AgentTimeline.vue` -- NEW

**Acceptance criteria:**
- Component renders correctly with empty events array.
- Component shows collapsed summary by default.
- Clicking expand shows full timeline.
- Each event shows the correct agent badge and description.
- Tool call events are visually distinct from status events.

**Estimated effort:** 45 minutes

---

### Task 2.3: Integrate AgentTimeline into MessageBubble

**Priority:** Medium
**Depends on:** 2.1, 2.2
**Blocks:** None

**Description:**

Add the `AgentTimeline` component to `MessageBubble.vue` for multi-agent responses (when `msg.agents.length > 1`).

**Requirements:** REQ-VIS-CS-001, REQ-VIS-TL-002

**Changes:**

1. Import `AgentTimeline` in `MessageBubble.vue`.
2. Add timeline after message content, before timestamp, when `msg.agents && msg.agents.length > 1`.
3. Pass `msg.agentHistory` (or derive from chatStore) as the `events` prop.
4. Ensure tool calls displayed in the message show agent attribution via `SpecialistBadge` when `showAgent` is true (multi-agent).

**Files touched:**
- `pkg/web/frontend/src/components/MessageBubble.vue`

**Acceptance criteria:**
- Single-agent messages do not show the timeline.
- Multi-agent messages show a collapsed timeline.
- Timeline integrates visually with the existing message layout.

**Estimated effort:** 20 minutes

---

### Task 2.4: Enhance chatStore for real-time agent tracking and segment buffering

**Priority:** High
**Depends on:** None (can start after Phase 1 concepts are understood)
**Blocks:** 2.5

**Description:**

Enhance the chatStore with streaming segment buffering and agent tool call tracking for the timeline view.

**Requirements:** REQ-VIS-TL-001, design.md section 5.3

**Changes:**

1. Add new state:
   - `activeStreamingAgent = ref(null)` -- which agent is currently streaming tokens.
   - `streamingSegmentBuffer = ref('')` -- buffer for current agent's streaming tokens.
2. Modify `setAgentStatus`:
   - When `status === 'streaming'`: set `activeStreamingAgent` to the agent, clear segment buffer.
   - When `status === 'complete'` or `'error'` or `'timeout'`: if the agent matches `activeStreamingAgent`, clear it.
   - Push to `agentHistory` with all event fields.
3. Modify `addToolCall`:
   - When `toolCall.agent` is set, push a `tool_call` event to `agentHistory` with agent, tool name, status, and timestamp.
4. Add `addAgentToolEvent(agent, toolName, status, timestamp)` method for explicit timeline event tracking.
5. Ensure `agentHistory` is cleared when a new streaming message starts (`clearAgentStatus` or `startStreamingMessage`).

**Files touched:**
- `pkg/web/frontend/src/stores/chatStore.js`

**Acceptance criteria:**
- `activeStreamingAgent` is set when `streaming` status arrives.
- `agentHistory` captures the full delegation lifecycle (status changes + tool calls).
- `agentHistory` is cleared between messages.
- Tool calls with `agent` field are tracked in `agentHistory`.

**Estimated effort:** 30 minutes

---

### Task 2.5: Handle agent-attributed tool_call events in ChatView

**Priority:** High
**Depends on:** 2.4
**Blocks:** None

**Description:**

Ensure `ChatView.vue` correctly handles the new `agent` field on `tool_call` WebSocket events and passes it to the chatStore.

**Requirements:** REQ-VIS-WS-001

**Changes:**

1. In the WebSocket `tool_call` handler in `ChatView.vue`, ensure the `agent` field from the event data is included when calling `chatStore.addToolCall(toolCall)`.
2. This should already work if the event data is passed through as-is -- verify and fix if needed.
3. Handle the new `agent_status` values (`streaming`, `error`, `timeout`) in the status handler -- ensure they are passed to `chatStore.setAgentStatus` without errors.

**Files touched:**
- `pkg/web/frontend/src/views/ChatView.vue` -- verify/modify WebSocket handlers

**Acceptance criteria:**
- Tool call events with `agent` field are passed to chatStore correctly.
- New agent status values (`streaming`, `error`, `timeout`) do not cause errors.
- Existing status values continue to work.

**Estimated effort:** 15 minutes

---

### Task 2.6: Update AgentStatusIndicator for new status values

**Priority:** Medium
**Depends on:** Phase 1 (specifically 1.8)
**Blocks:** None

**Description:**

Update the `AgentStatusIndicator.vue` component to handle the new agent status values emitted by the backend.

**Requirements:** REQ-VIS-SI-001

**Changes:**

1. Add new entries to `statusMap`:
   - `streaming`: "{agent} is generating a response..."
   - `tool_call`: "{agent} is executing a tool..."
   - `error`: "{agent} encountered an error"
   - `timeout`: "{agent} timed out"
2. Add corresponding color mappings:
   - `streaming`: blue border, blue icon (like `analyzing`)
   - `tool_call`: amber border, amber icon
   - `error`: red border, red icon
   - `timeout`: red border, red icon
3. The `complete` status should transition to hidden (existing behavior -- verify).

**Files touched:**
- `pkg/web/frontend/src/components/Chat/AgentStatusIndicator.vue`

**Acceptance criteria:**
- All new statuses display with correct text and colors.
- `complete` status hides the indicator.
- Unknown statuses show a generic "Processing..." (graceful degradation).

**Estimated effort:** 15 minutes

---

### Task 2.7: WebSocket tool_call events include agent field in JSON output

**Priority:** High
**Depends on:** Phase 1 (specifically 1.7)
**Blocks:** None

**Description:**

Verify and ensure that the WebSocket handler correctly serializes the `agent` field in `tool_call` events. This was partially done in Task 1.7 but needs explicit verification.

**Requirements:** REQ-VIS-WS-001

**Changes:**

1. Verify the `onTool` callback in `handleChatWS` includes the `agent` field when present.
2. Verify the existing (non-specialist) tool call path does NOT include `agent` field (backward-compatible).
3. Verify `stream_end` message still includes `agents` array correctly.

**Files touched:**
- `pkg/web/server.go` -- verify (may already be done by Task 1.7)

**Acceptance criteria:**
- Specialist tool calls: `{"type":"tool_call","name":"...","agent":"developer"}`.
- Non-specialist tool calls: `{"type":"tool_call","name":"..."}` (no agent field).
- Backward-compatible with existing frontends.

**Estimated effort:** 10 minutes

---

## Phase 3: Token Optimization (Backend)

Phase 3 focuses on reducing token consumption in multi-agent conversations and adding parallel delegation support.

---

### Task 3.1: Specialist context trimming with SkipHistory

**Priority:** High
**Depends on:** 1.2
**Blocks:** None

**Description:**

Ensure `ProcessDirectWithCallbacks` correctly implements the `SkipHistory` flag from `DelegationCallbacks`. When true, the specialist's LLM request should contain only the system prompt and the current task message -- no session history.

**Requirements:** REQ-OPT-CT-001

**Changes:**

1. In `runAgentLoopStreamWithTools` (and `runAgentLoopWithTools`), when `skipHistory=true`:
   - Do not call `al.sessions.GetHistoryForUser(...)`.
   - Do not call `al.sessions.GetSummaryForUser(...)`.
   - Pass empty history and summary to `al.contextBuilder.BuildMessages(...)`.
2. Verify that `ProcessWithSpecialityStream` sets `callbacks.SkipHistory = true` (done in Task 1.4).
3. Verify that `lightweightMode=true` already skips bootstrap files and memory (existing behavior).

**Files touched:**
- `pkg/agent/loop.go` -- verify `skipHistory` in `runAgentLoopStreamWithTools`

**Acceptance criteria:**
- Specialist with `SkipHistory=true` sends only system prompt + task message to LLM.
- Specialist with `SkipHistory=false` loads full session history (backward-compatible).
- Measured by logging the message count sent to the LLM.

**Estimated effort:** 15 minutes

---

### Task 3.2: Parallel delegation tool -- delegate_to_specialists

**Priority:** Medium
**Depends on:** 1.3, 1.5
**Blocks:** 3.3

**Description:**

Implement a new `ParallelDelegationTool` that allows the orchestrator to delegate to multiple specialists concurrently.

**Requirements:** REQ-OPT-PD-001

**Changes:**

1. Add `ParallelDelegationTool` struct in `pkg/agent/orchestrator.go`:
   - `Name() string` returns `"delegate_to_specialists"`.
   - `Description()` explains it accepts an array of delegation requests.
   - `Parameters()` defines `delegations` as an array of `{specialist_name, task, context}` objects.
2. `Execute(ctx, args)`:
   - Parse `delegations` array from args.
   - Validate all specialist names exist before launching any goroutine.
   - Use `errgroup.Group` with concurrency limit from `MaxParallelDelegations` config.
   - Each goroutine calls `processSpecialistTask(ctx, name, task)` (reusing existing function).
   - Collect results into `[]DelegationResult` preserving request order.
   - Return JSON-formatted aggregated results.
   - Handle partial failures: individual specialist errors do not fail the whole tool call.
3. Register the tool on the orchestrator's agent loop alongside the existing `delegate_to_specialist`.

**Files touched:**
- `pkg/agent/orchestrator.go` -- add `ParallelDelegationTool`

**Acceptance criteria:**
- Two concurrent delegations run in separate goroutines.
- One specialist failure does not prevent others from completing.
- Concurrency limit is respected.
- Empty delegations array returns an error.
- Context cancellation cascades to all goroutines (no leaks).

**Estimated effort:** 60 minutes

---

### Task 3.3: MaxParallelDelegations configuration

**Priority:** Medium
**Depends on:** 3.2
**Blocks:** None

**Description:**

Add `MaxParallelDelegations` field to `OrchestratorConfig` with a default of 3.

**Requirements:** REQ-OPT-PD-002

**Changes:**

1. Add `MaxParallelDelegations int \`json:"max_parallel_delegations,omitempty"\`` to `OrchestratorConfig`.
2. In `InitializeOrchestrator`, apply default of 3 if not set.
3. Wire the config value to `ParallelDelegationTool`.

**Files touched:**
- `pkg/config/config.go` -- add field to `OrchestratorConfig`
- `pkg/agent/orchestrator.go` -- read config and pass to `ParallelDelegationTool`

**Acceptance criteria:**
- Config with `max_parallel_delegations: 5` allows 5 concurrent delegations.
- Config without the field defaults to 3.
- Value of 0 defaults to 3 (defensive).

**Estimated effort:** 15 minutes

---

### Task 3.4: Add token usage fields to DelegationResult

**Priority:** Medium
**Depends on:** None
**Blocks:** 3.5

**Description:**

Extend `DelegationResult` with token usage and duration fields for per-specialist cost tracking.

**Requirements:** REQ-OPT-TB-001

**Changes:**

1. Add fields to `DelegationResult`:
   - `TokensUsed int \`json:"tokens_used,omitempty"\``
   - `PromptTokens int \`json:"prompt_tokens,omitempty"\``
   - `CompletionTokens int \`json:"completion_tokens,omitempty"\``
   - `DurationMs int64 \`json:"duration_ms,omitempty"\``
2. In `processSpecialistTask`, measure duration with `time.Since(start)` and populate `DurationMs`.
3. Populate token fields from `TokenUsage` returned by `ProcessDirectWithCallbacks` (when implemented).

**Files touched:**
- `pkg/agent/orchestrator.go` -- extend `DelegationResult`, populate fields

**Acceptance criteria:**
- `DelegationResult` JSON output includes token fields when non-zero.
- `DelegationResult` JSON omits token fields when zero (backward-compatible).
- Duration is measured accurately.

**Estimated effort:** 20 minutes

---

### Task 3.5: Token usage logging for specialist delegations

**Priority:** Low
**Depends on:** 3.4
**Blocks:** 3.6

**Description:**

Log token usage at INFO level after each specialist delegation completes.

**Requirements:** REQ-OPT-TB-002

**Changes:**

1. In `processSpecialistTask`, after receiving the delegation result, log:
   ```go
   logger.InfoCF("agent", "Specialist delegation completed", map[string]interface{}{
       "specialist": specialistName,
       "tokens_used": result.TokensUsed,
       "prompt_tokens": result.PromptTokens,
       "completion_tokens": result.CompletionTokens,
       "duration_ms": result.DurationMs,
       "success": result.Success,
   })
   ```

**Files touched:**
- `pkg/agent/orchestrator.go` -- add logging

**Acceptance criteria:**
- INFO log is emitted with component "agent" after each delegation.
- Log includes specialist name, token counts, duration, and success status.

**Estimated effort:** 10 minutes

---

### Task 3.6: Optional token usage in WebSocket stream_end and content_segment events

**Priority:** Low
**Depends on:** 3.5
**Blocks:** None

**Description:**

Optionally include token usage in `stream_end` and `content_segment` WebSocket events for frontend display.

**Requirements:** REQ-OPT-TB-003, design.md section 7.2

**Changes:**

1. In the `stream_end` WebSocket message construction in `handleChatWS`, include optional `token_usage` field when available.
2. In `emitContentSegment`, optionally include `tokens_used` and `duration_ms` fields.
3. Both fields are optional and omitted when zero/nil.

**Files touched:**
- `pkg/web/server.go` -- extend `stream_end` and `content_segment` messages

**Acceptance criteria:**
- `stream_end` includes `token_usage` when specialist token data is available.
- `stream_end` omits `token_usage` when not available (backward-compatible).
- Existing frontends that do not handle `token_usage` are unaffected.

**Estimated effort:** 20 minutes

---

## Cross-Cutting Tasks

These tasks apply across all phases and should be verified continuously.

---

### Task 4.1: Security -- no privilege escalation through delegation

**Priority:** High
**Depends on:** All phases
**Blocks:** None

**Description:**

Verify that the callback propagation and parallel delegation changes do not create security holes.

**Requirements:** REQ-CC-SEC-001

**Verification checklist:**

1. Tool permission filtering is still applied per-specialist via `ToolFilter()` with `allowedTools`.
2. Audit logging (`SQLiteAuditLogger`) continues to log specialist tool executions.
3. Workspace isolation is preserved -- specialist tools operate within the user's workspace.
4. The `Agent` field on `ToolEvent` is informational only and not used for access control.
5. Per-call tool registry copies contain only the specialist's allowed tools (not the full registry).

**Files to review:**
- `pkg/agent/specialist.go` -- `ToolFilter()` logic
- `pkg/tools/audit.go` -- audit logger still invoked
- `pkg/agent/permissions.go` -- permission filtering unaffected

**Estimated effort:** 15 minutes (review only)

---

### Task 4.2: Backward compatibility -- non-orchestrated loops unaffected

**Priority:** Critical
**Depends on:** All phases
**Blocks:** None

**Description:**

Verify that all changes are purely additive and non-orchestrated agent loops work identically to before.

**Requirements:** REQ-CC-BC-001

**Verification checklist:**

1. `ProcessDirect`, `ProcessDirectWithModelStream`, `ProcessDirectWithUser`, `ProcessDirectWithChannel`, `ProcessDirectWithChannelForUser` -- all unchanged.
2. `runAgentLoop` and `runAgentLoopStream` -- unchanged (thin wrappers now call `Impl` methods but behavior is identical).
3. Non-orchestrated WebSocket chat: no `agent_status`, no `content_segment`, `stream_end` with no agents array or single-agent array.
4. CLI chat: completely unaffected.
5. Gateway channels: completely unaffected.
6. Run full test suite: `go test ./...`.

**Files to review:**
- All modified files in Phase 1

**Estimated effort:** 20 minutes (test run + review)

---

### Task 4.3: Multi-user isolation verification

**Priority:** High
**Depends on:** Phase 1
**Blocks:** None

**Description:**

Verify that callback propagation is per-request and does not leak between users.

**Requirements:** REQ-CC-MU-001

**Verification checklist:**

1. Callbacks are injected per-WebSocket-connection in `handleChatWS`.
2. Each connection has its own `wsMu` mutex and `conn` reference.
3. Per-call tool registry copies prevent cross-user tool state contamination.
4. Specialist session keys are per-user when multi-user is enabled.
5. `context.Context` is request-scoped and does not persist between requests.

**Files to review:**
- `pkg/web/server.go` -- `handleChatWS` callback injection
- `pkg/agent/specialist.go` -- per-call tool registry
- `pkg/agent/orchestrator.go` -- context-based callback extraction

**Estimated effort:** 15 minutes (review only)

---

### Task 4.4: Performance -- streaming latency budget

**Priority:** Medium
**Depends on:** Phase 1
**Blocks:** None

**Description:**

Verify that the callback wrappers and context injection do not measurably impact streaming latency.

**Requirements:** REQ-CC-PERF-001

**Verification checklist:**

1. Callback wrappers are thin closures with no allocations per invocation.
2. `wrapToolCallback` only sets one field on the `ToolEvent` struct.
3. Context injection (`ContextWithStreamCallback`, `ContextWithToolCallback`) adds two `context.WithValue` calls -- negligible overhead.
4. WebSocket write path (`wsMu.Lock -> WriteJSON -> Unlock`) is the dominant latency factor and is unchanged.
5. Optional: benchmark `wrapToolCallback` in isolation to confirm <1us per call.

**Files to review:**
- `pkg/agent/specialist.go` -- `wrapCallbacksWithAttribution`
- `pkg/agent/loop.go` -- context injection in `runAgentLoopStreamWithTools`

**Estimated effort:** 10 minutes (review, optional benchmark)

---

## Implementation Order (Recommended)

### Session 1: Foundation (Tasks 1.1, 1.2, 1.3, 1.4, 1.6)

These tasks establish the core types, methods, and concurrency fix. They can be done in a single session:

1. **1.1** -- Define types (15 min)
2. **1.6** -- Callback wrappers (15 min, depends on 1.1)
3. **1.4** -- ProcessWithSpecialityStream + concurrency fix (30 min)
4. **1.2** -- ProcessDirectWithCallbacks (30 min)
5. **1.3** -- Refactor run methods (45 min)

**Total: ~2.5 hours**

### Session 2: Pipeline Completion (Tasks 1.5, 1.7, 1.8, 1.9)

Thread callbacks through the full pipeline and verify end-to-end:

1. **1.5** -- extractDelegationCallbacks + orchestrator wiring (45 min)
2. **1.7** -- WebSocket handler injection (20 min)
3. **1.8** -- Granular status events (20 min)
4. **1.9** -- End-to-end verification + tests (60 min)

**Total: ~2.5 hours**

### Session 3: Frontend Visibility (Tasks 2.1, 2.2, 2.4, 2.5, 2.6, 2.3, 2.7)

Frontend enhancements for real-time agent visibility:

1. **2.1** -- Dynamic SpecialistBadge (25 min)
2. **2.2** -- AgentTimeline component (45 min)
3. **2.4** -- ChatStore enhancements (30 min)
4. **2.5** -- ChatView handler updates (15 min)
5. **2.6** -- AgentStatusIndicator updates (15 min)
6. **2.3** -- Integrate timeline into MessageBubble (20 min)
7. **2.7** -- Verify WebSocket agent field (10 min)

**Total: ~2.5 hours**

### Session 4: Token Optimization (Tasks 3.1, 3.2, 3.3, 3.4, 3.5, 3.6)

Token savings and parallel delegation:

1. **3.1** -- Context trimming verification (15 min)
2. **3.4** -- DelegationResult token fields (20 min)
3. **3.5** -- Token usage logging (10 min)
4. **3.2** -- Parallel delegation tool (60 min)
5. **3.3** -- MaxParallelDelegations config (15 min)
6. **3.6** -- WebSocket token usage (20 min)

**Total: ~2.5 hours**

### Session 5: Cross-Cutting Verification (Tasks 4.1, 4.2, 4.3, 4.4)

Security review, backward compatibility, multi-user isolation, and performance:

1. **4.1** -- Security review (15 min)
2. **4.2** -- Backward compatibility verification (20 min)
3. **4.3** -- Multi-user isolation (15 min)
4. **4.4** -- Performance review (10 min)
5. Build frontend dist: `cd pkg/web/frontend && npm run build`
6. Full test suite: `go test ./...`

**Total: ~1 hour**

---

## File Impact Summary

| File | Phase | Tasks | Type of Change |
|------|-------|-------|----------------|
| `pkg/agent/loop.go` | 1, 3 | 1.1, 1.2, 1.3, 1.5, 3.1 | Add types, methods, refactor iterations |
| `pkg/agent/specialist.go` | 1 | 1.4, 1.6, 1.8 | Add methods, concurrency fix |
| `pkg/agent/orchestrator.go` | 1, 3 | 1.5, 1.8, 3.2, 3.3, 3.4, 3.5 | Context keys, callback extraction, parallel tool, token tracking |
| `pkg/config/config.go` | 3 | 3.3 | Add MaxParallelDelegations field |
| `pkg/web/server.go` | 1, 2, 3 | 1.7, 2.7, 3.6 | WebSocket callback injection, agent field, token usage |
| `pkg/web/frontend/src/components/Chat/SpecialistBadge.vue` | 2 | 2.1 | Dynamic hash-based colors |
| `pkg/web/frontend/src/components/Chat/AgentTimeline.vue` | 2 | 2.2 | NEW component |
| `pkg/web/frontend/src/components/Chat/AgentStatusIndicator.vue` | 2 | 2.6 | New status values |
| `pkg/web/frontend/src/components/MessageBubble.vue` | 2 | 2.3 | Timeline integration |
| `pkg/web/frontend/src/stores/chatStore.js` | 2 | 2.4 | Agent tracking, segment buffering |
| `pkg/web/frontend/src/views/ChatView.vue` | 2 | 2.5 | Verify WebSocket handlers |
| `pkg/agent/loop_test.go` | 1 | 1.9 | New tests |
| `pkg/agent/specialist_test.go` | 1 | 1.9 | New tests |
| `pkg/agent/orchestrator_test.go` | 1 | 1.9 | New tests |

---

## Risk Register

| Risk | Phase | Impact | Mitigation |
|------|-------|--------|------------|
| Refactoring `runLLMIterationStream` into `Impl` breaks existing behavior | 1 | High | Task 1.3 preserves exact signatures; run full test suite after |
| Specialist streaming conflicts with orchestrator tool-result processing | 1 | Medium | Specialist tokens stream in real-time; final result still returned as string to orchestrator |
| Interleaved tokens from parallel specialists confuse frontend | 3 | Medium | `agent_status` events delineate active agent; content segments provide authoritative attribution |
| Memory leak from per-call tool registry copies | 1 | Low | ToolRegistry is ~1KB per copy; GC handles it; delegation rate is ~1-10/min |
| WebSocket write contention under parallel delegation | 3 | Medium | Existing `wsMu` mutex serializes all writes; tested under concurrent load |
| `ProcessWithSpeciality` (old path) still has concurrency bug | 1 | Low | Deprecation comment added; new code uses `ProcessWithSpecialityStream`; old path is fallback only |

---

## Approval

| Role | Name | Date | Decision |
|------|------|------|----------|
| Author | SDD Sub-Agent | 2026-02-27 | Proposed |
| Reviewer | | | |
| Approver | | | |
