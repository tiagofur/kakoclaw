# Proposal: Multi-Agent Orchestration Visibility

**Change ID:** multi-agent-orchestration-visibility
**Status:** Draft
**Author:** SDD Sub-Agent
**Created:** 2026-02-27
**Based on:** [exploration.md](exploration.md)

---

## 1. Intent

### Problem Statement

MakoClaw's multi-agent orchestration system is architecturally well-built -- the orchestrator, specialist registry, delegation tool, agent tracking, and frontend components all exist and are wired together. However, the system delivers a **silent experience** when specialists are working because the callback propagation chain is broken at the orchestrator-to-specialist boundary.

Specifically: when a user sends a message via the WebSocket chat, the web handler injects `OnToken`, `OnTool`, `OnAgentStatus`, and `OnContentSegment` callbacks into a `processOptions` struct. When the orchestrator delegates to a specialist, `DelegationTool.Execute()` calls `processSpecialistTask()`, which calls `specialist.ProcessWithSpeciality()`, which calls `ProcessDirect()` -- and `ProcessDirect` ultimately calls `processMessageWithModel()`, which creates a **fresh** `processOptions{}` with **no callbacks**. The specialist's entire LLM loop runs silently: no streamed tokens, no tool call events, no status updates reach the user.

The frontend has all the plumbing ready -- `AgentStatusIndicator`, `SpecialistBadge`, `MessageBubble` with segment rendering, `chatStore` with agent tracking state -- but these components are starved of events because the backend never sends them during specialist execution.

### Current State

- Orchestrator delegates work to specialists via `delegate_to_specialist` tool -- this works correctly
- Specialist runs its full LLM loop via `ProcessDirect()` which creates `processOptions{}` with zero callbacks
- User sees **nothing** during specialist execution (no streaming, no tool calls, no status)
- Agent attribution (`involvedAgents`) works but only appears at `stream_end`, not during streaming
- `AgentStatusIndicator` only shows orchestrator-level events ("analyzing", "delegating"), not specialist internals
- `content_segment` events fire once when the specialist returns its final result, not during streaming
- Concurrency bug: `ProcessWithSpeciality` swaps `sa.tools` under mutex but the defer-restore can race with concurrent delegations
- `SpecialistBadge` is hardcoded to 6 specialist types (developer, documentation, testing, devops, analyst, researcher)

### Desired State

- **Real-time streaming** from specialist LLM responses reaches the user as tokens are generated
- **Tool call visibility**: specialist tool invocations appear in the frontend as they happen
- **Granular agent status events** during specialist processing (not just orchestrator-level events)
- **Per-agent content attribution** during streaming, not just in `stream_end`
- **Dynamic specialist badges** that work for any custom-named specialist
- **Thread-safe specialist execution** without the tool-swap concurrency bug
- **Token-efficient context** for specialists (trimmed history, budget awareness)

### Value Proposition

1. **User trust and transparency**: Users can see exactly which agent is working and what it is doing in real time, building confidence in multi-agent responses
2. **Debugging and observability**: Developers and operators can diagnose delegation issues, slow specialists, and tool failures through visible events
3. **Token savings**: Proper context trimming and parallel delegation reduce total token consumption for multi-agent conversations
4. **Correctness**: Fixing the concurrency bug prevents intermittent tool registry corruption under concurrent delegations

---

## 2. Scope

### In Scope

| Component | Description |
|-----------|-------------|
| `pkg/agent/loop.go` | New `ProcessDirectWithCallbacks` method that accepts full callback set |
| `pkg/agent/specialist.go` | Thread callbacks through `ProcessWithSpeciality`; fix concurrency bug with per-call tool registry copies |
| `pkg/agent/orchestrator.go` | Extract and forward callbacks from `ctx` to specialist; parallel delegation support |
| `pkg/web/server.go` | Pass `OnAgentStatus` and `OnContentSegment` into `processOptions` for streaming path |
| `pkg/web/frontend/src/components/Chat/SpecialistBadge.vue` | Dynamic badge generation from specialist name hash |
| `pkg/web/frontend/src/stores/chatStore.js` | Handle real-time specialist tool call events with agent attribution |
| `pkg/web/frontend/src/views/ChatView.vue` | Display specialist tool calls and streaming tokens with attribution |
| Testing | Unit tests for callback propagation, concurrency safety, and parallel delegation |

### Out of Scope (v1)

| Item | Reason |
|------|--------|
| spawn tool streaming | Separate system; lower priority and simpler architecture |
| Specialist session persistence | Ephemeral specialist sessions are by-design for token efficiency |
| Response caching across specialists | Complex cache invalidation logic; defer to v2 |
| Delegation cost estimation | Requires token counting integration; defer to v2 |
| Inter-specialist communication | Not needed for current use cases |

### Dependencies

| Dependency | Type | Description |
|------------|------|-------------|
| `pkg/agent/loop.go` | Internal | Core agent loop with processOptions, streaming infrastructure |
| `pkg/agent/orchestrator.go` | Internal | Delegation tool, agent status events, content segments |
| `pkg/agent/specialist.go` | Internal | Specialist execution, tool filtering |
| `pkg/providers/` | Internal | StreamingLLMProvider interface for specialist streaming |
| `pkg/web/server.go` | Internal | WebSocket handler, callback injection |
| `pkg/web/frontend/` | Internal | Vue components for agent visibility |

---

## 3. Approach

### Architecture Overview

**Current flow (broken):**
```
WebSocket handler (callbacks injected into ctx)
  -> activeAgentLoop.ProcessDirectWithUserAndModelStream(callbacks)
    -> runAgentLoopStream(processOptions{OnToken, OnTool, ...})
      -> LLM decides to call delegate_to_specialist
        -> DelegationTool.Execute(ctx)  // ctx has callbacks
          -> processSpecialistTask(ctx)
            -> specialist.ProcessWithSpeciality(ctx)
              -> ProcessDirect()         // <-- CALLBACKS LOST HERE
                -> processMessageWithModel()
                  -> runAgentLoop(processOptions{})  // EMPTY callbacks
```

**Proposed flow (fixed):**
```
WebSocket handler (callbacks injected into ctx)
  -> activeAgentLoop.ProcessDirectWithUserAndModelStream(callbacks)
    -> runAgentLoopStream(processOptions{OnToken, OnTool, ...})
      -> LLM decides to call delegate_to_specialist
        -> DelegationTool.Execute(ctx)  // ctx has callbacks
          -> processSpecialistTask(ctx)
            -> extract callbacks from ctx
            -> specialist.ProcessWithSpecialityStream(ctx, callbacks)
              -> ProcessDirectWithCallbacks()     // <-- CALLBACKS FORWARDED
                -> processMessageWithModelStream()
                  -> runAgentLoopStream(processOptions{OnToken*, OnTool*, OnAgentStatus, OnContentSegment})
```

\* `OnToken` and `OnTool` are wrapped to prefix events with specialist agent name for attribution.

### Phase 1: Fix the Pipeline (Backend)

**Goal:** Thread callbacks through the entire orchestrator -> specialist delegation chain so specialist work is visible.

#### 1.1 New `ProcessDirectWithCallbacks` method on `AgentLoop`

Add a method that accepts the full callback set and passes them through to `runAgentLoopStream`:

```go
type DelegationCallbacks struct {
    OnToken          StreamCallback
    OnTool           ToolCallback
    OnAgentStatus    AgentStatusCallback
    OnContentSegment ContentSegmentCallback
}

func (al *AgentLoop) ProcessDirectWithCallbacks(
    ctx context.Context,
    content, sessionKey string,
    callbacks DelegationCallbacks,
) (string, error)
```

This method creates a `processOptions` with all callbacks populated and routes to `runAgentLoopStream` instead of `runAgentLoop`.

#### 1.2 New `ProcessWithSpecialityStream` on `SpecialistAgent`

Modify the specialist to accept and forward callbacks:

```go
func (sa *SpecialistAgent) ProcessWithSpecialityStream(
    ctx context.Context,
    userMessage string,
    callbacks DelegationCallbacks,
) (string, error)
```

This method wraps the callbacks to add agent attribution (prefixing specialist name) before forwarding to `ProcessDirectWithCallbacks`. Falls back to `ProcessWithSpeciality` if no callbacks are provided.

#### 1.3 Extract callbacks from context in `processSpecialistTask`

Modify `processSpecialistTask` to extract `AgentStatusCallback` and `ContentSegmentCallback` from `ctx`, plus the `OnToken` and `OnTool` callbacks from the orchestrator's own `processOptions`. These are then bundled into `DelegationCallbacks` and passed to `ProcessWithSpecialityStream`.

#### 1.4 Fix the concurrency bug in tool swap

**Current bug:** `ProcessWithSpeciality` swaps `sa.tools` under a mutex, but the mutex only protects the swap operation -- not the entire execution. If two concurrent delegations arrive:

```
Goroutine A: Lock -> swap tools -> Unlock -> runs with filtered tools
Goroutine B: Lock -> swap tools -> Unlock -> runs with filtered tools
Goroutine A: Lock -> restore original -> Unlock  // Now B is broken
```

**Fix:** Instead of swapping the shared `sa.tools` field, create a per-call copy of the AgentLoop's tool registry. The specialist's `ProcessWithSpecialityStream` / `ProcessWithSpeciality` should pass the filtered registry as a parameter rather than mutating the shared state:

```go
func (sa *SpecialistAgent) ProcessWithSpecialityStream(ctx context.Context, userMessage string, callbacks DelegationCallbacks) (string, error) {
    filteredTools := sa.ToolFilter()  // Creates a new registry each time
    // Pass filteredTools to a new method that accepts a tool registry override
    return sa.processDirectWithToolsAndCallbacks(ctx, userMessage, filteredTools, callbacks)
}
```

This eliminates the mutex entirely and makes concurrent delegations safe.

### Phase 2: Enrich Visibility (Backend + Frontend)

**Goal:** Deliver rich, real-time agent events to the frontend during specialist execution.

#### 2.1 Callback wrappers with agent attribution

Create wrapper functions that annotate each callback event with the specialist name:

```go
func wrapTokenCallback(agentName string, onToken StreamCallback) StreamCallback {
    return func(token string) error {
        // Token is forwarded as-is; attribution is tracked separately
        return onToken(token)
    }
}

func wrapToolCallback(agentName string, onTool ToolCallback) ToolCallback {
    return func(ev ToolEvent) error {
        // Prefix tool event with agent name for frontend attribution
        ev.Agent = agentName  // New field on ToolEvent
        return onTool(ev)
    }
}
```

#### 2.2 Extend `ToolEvent` with agent attribution

Add an `Agent` field to `ToolEvent` so the frontend knows which agent invoked a tool:

```go
type ToolEvent struct {
    Name   string                 `json:"name"`
    Args   map[string]interface{} `json:"arguments"`
    Result string                 `json:"result,omitempty"`
    Status string                 `json:"status"`
    Agent  string                 `json:"agent,omitempty"`  // NEW
}
```

#### 2.3 WebSocket protocol enhancement for `tool_call`

Extend the `tool_call` WebSocket message to include the `agent` field:

```json
{
    "type": "tool_call",
    "name": "read_file",
    "args": {"path": "/etc/config"},
    "status": "started",
    "agent": "developer"
}
```

This is backward-compatible -- existing frontends will simply ignore the new field.

#### 2.4 Dynamic SpecialistBadge

Replace the hardcoded 6-type badge map with a deterministic hash-based approach:

```javascript
// Generate consistent color/icon from specialist name
function getBadgeConfig(name) {
    const knownTypes = { /* existing 6 types */ };
    if (knownTypes[name]) return knownTypes[name];

    // Hash-based fallback for custom specialists
    const hash = hashCode(name);
    const hue = hash % 360;
    return {
        color: `hsl(${hue}, 60%, 45%)`,
        bgColor: `hsl(${hue}, 60%, 95%)`,
        icon: getIconFromName(name),  // First letter or category-matched icon
        label: name.charAt(0).toUpperCase() + name.slice(1)
    };
}
```

#### 2.5 Agent timeline in chat store

Enhance `agentHistory` tracking to include tool call events with timestamps, so the frontend can render a timeline view showing the orchestration sequence:

```javascript
// chatStore additions
addAgentToolEvent(agent, toolName, status, timestamp) {
    this.agentHistory.push({
        type: 'tool_call',
        agent,
        tool: toolName,
        status,
        timestamp
    });
}
```

### Phase 3: Token Optimization

**Goal:** Reduce token consumption in multi-agent conversations.

#### 3.1 Context trimming for specialists

Specialists currently receive the full conversation history via `ProcessDirect`. Since specialists handle isolated tasks, they should receive:
- Their specialist system prompt (already done via `SetAgentSystemPrompt`)
- Only the current task message (not the full user conversation history)
- A brief summary of relevant context if provided by the orchestrator

Modify `ProcessDirectWithCallbacks` to accept an option to skip session history loading:

```go
type DelegationCallbacks struct {
    // ... existing fields ...
    SkipHistory bool  // Don't load session history for this call
}
```

#### 3.2 Parallel delegation support

Allow the orchestrator to delegate to multiple specialists simultaneously. This requires:

1. A new `delegate_to_specialists` tool (plural) that accepts an array of delegation requests
2. Concurrent execution using goroutines with a `sync.WaitGroup`
3. Result aggregation with per-specialist error handling
4. Proper callback multiplexing (specialist tokens are interleaved but each segment is attributed)

```go
func (dt *ParallelDelegationTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    // Parse multiple delegation requests
    // Launch goroutines per specialist
    // Collect results with WaitGroup
    // Return aggregated result
}
```

#### 3.3 Token budget tracking

Add per-specialist token counting to track and optionally limit token usage:

```go
type DelegationResult struct {
    SpecialistName string `json:"specialist_name"`
    Result         string `json:"result"`
    Success        bool   `json:"success"`
    Error          string `json:"error,omitempty"`
    TokensUsed     int    `json:"tokens_used,omitempty"`  // NEW
}
```

This is informational in v1 (logged and reported) but can be extended to enforce budgets in v2.

---

## 4. Risks and Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Callback wrapper overhead slows streaming | Low | Low | Wrappers are thin function closures; benchmark to confirm negligible overhead |
| Interleaved tokens from parallel specialists confuse the frontend | Medium | Medium | Use content segments with agent attribution; frontend renders segments separately |
| Specialist streaming creates backpressure on WebSocket | Medium | Low | WebSocket writer already buffers; add write timeout and drop tokens under pressure |
| Breaking change to `ToolEvent` struct | Low | Low | New `Agent` field is `omitempty`; existing code unaffected |
| Removing tool-swap mutex introduces subtle regression | Medium | Low | The fix *eliminates* shared mutable state; each call gets its own registry copy |
| Phase 3 parallel delegation introduces complex error handling | Medium | Medium | Start with sequential delegation in Phase 1/2; parallel is additive in Phase 3 |
| Specialist streaming conflicts with orchestrator's tool-result processing | Medium | Medium | Specialist tokens are streamed in real-time but the final result is still returned as a string to the orchestrator's tool loop |

### Rollback Plan

Each phase is independently deployable and rollbackable:

- **Phase 1 rollback:** If `ProcessDirectWithCallbacks` causes issues, `ProcessWithSpeciality` still calls `ProcessDirect` as today. The new methods are additive.
- **Phase 2 rollback:** Frontend changes are purely additive (new `agent` field on `tool_call`). Removing the new field returns to current behavior.
- **Phase 3 rollback:** Parallel delegation is a new tool alongside the existing `delegate_to_specialist`. Removing it has no effect on the sequential path.

---

## 5. Success Criteria

| Criterion | Measurement |
|-----------|-------------|
| **Specialist streaming** | User sees streamed tokens from specialist LLM responses in real time |
| **Tool call visibility** | Specialist tool invocations appear in the chat UI with agent attribution |
| **Agent status granularity** | Frontend shows at least 4 distinct statuses during a delegation (analyzing, delegating, working, complete) |
| **Concurrency safety** | No test failures or data races under `go test -race ./pkg/agent/...` with concurrent delegations |
| **Dynamic badges** | Custom-named specialists (e.g., "security-auditor") get unique colored badges without code changes |
| **Token reduction** | Specialists consume fewer tokens than a full agent loop (measurable via `DelegationResult.TokensUsed`) |
| **Backward compatibility** | Existing non-orchestrated agent loops are completely unaffected |
| **No breaking changes** | All existing WebSocket message types continue to work; new fields are additive only |

---

## 6. Alternatives Considered

### Alternative A: Inject callbacks via context.Context instead of explicit parameters

Pass `OnToken`, `OnTool`, etc. through `context.WithValue()` keys rather than adding new method signatures.

**Pros:**
- No new method signatures needed
- Callbacks travel through any function that passes `ctx`

**Cons:**
- Type-unsafe: context values lose compile-time type checking
- Hidden dependencies: hard to know which functions require which context values
- Debugging difficulty: callbacks disappear silently if key is wrong

**Decision:** Rejected. Explicit parameters are clearer and safer. The `ctx` approach is already used for `AgentStatusCallback` and `ContentSegmentCallback`, which has proven error-prone (the very bug we are fixing stems from callback propagation via context being incomplete). For the core streaming path, explicit parameters are preferred.

### Alternative B: Make specialists share the orchestrator's agent loop

Instead of specialists having their own `AgentLoop`, have them run as a "mode" of the orchestrator's loop with swapped tools and system prompt.

**Pros:**
- Callbacks automatically propagate (same loop)
- Simpler architecture

**Cons:**
- Destroys specialist isolation (shared state, shared session)
- Prevents concurrent delegation (single loop)
- Major refactor of existing architecture

**Decision:** Rejected. The current architecture of independent specialist loops is sound. We just need to thread the callbacks through.

### Alternative C: Event bus for agent events instead of direct callbacks

Create an internal event bus (channel-based or pub/sub) for agent events, decoupling producers from consumers.

**Pros:**
- Clean separation of concerns
- Easy to add new event consumers (logging, metrics)

**Cons:**
- Adds significant complexity for a problem that has a direct fix
- Introduces async event ordering challenges
- Over-engineering for the current scale

**Decision:** Deferred. May be valuable in v2 when more event consumers exist. For now, direct callback threading is simpler and sufficient.

---

## 7. Open Questions

| Question | Recommendation | Status |
|----------|----------------|--------|
| Should specialist tokens be visually distinguished from orchestrator tokens? | Yes -- use subtle background color or prefix in the streaming UI | Decided |
| Should parallel delegation be limited to N concurrent specialists? | Yes -- default to 3; configurable via `orchestrator.max_parallel_delegations` | Decided |
| Should the orchestrator see specialist streaming tokens? | No -- orchestrator only sees the final result as a tool response | Decided |
| Should `content_segment` events include token-level granularity? | No -- segments are emitted per-specialist-response; individual tokens go via `stream` events | Decided |
| How to handle specialist errors during streaming? | Emit `agent_status` with `status: "error"`, then let orchestrator decide next step | Decided |

---

## 8. Implementation Notes

### Affected Packages

| Package | Files Modified | Description |
|---------|---------------|-------------|
| `pkg/agent` | `loop.go` | Add `ProcessDirectWithCallbacks`, `DelegationCallbacks` struct |
| `pkg/agent` | `specialist.go` | Add `ProcessWithSpecialityStream`, fix concurrency bug |
| `pkg/agent` | `orchestrator.go` | Extract callbacks from ctx, forward to specialist, parallel delegation tool |
| `pkg/web` | `server.go` | Pass `OnAgentStatus`/`OnContentSegment` to processOptions |
| `pkg/web/frontend` | `SpecialistBadge.vue` | Dynamic hash-based badge generation |
| `pkg/web/frontend` | `chatStore.js` | Agent tool call tracking, timeline events |
| `pkg/web/frontend` | `ChatView.vue` | Render specialist tool calls with attribution |
| `pkg/web/frontend` | `MessageBubble.vue` | Potentially minor updates for real-time segments |

### Multi-User Impact

This change is transparent to multi-user deployments. The callback propagation is per-request and does not affect user isolation, session management, or workspace boundaries. Per-user specialist configurations continue to work as-is.

### Breaking Changes

None. All changes are additive:
- New methods (`ProcessDirectWithCallbacks`, `ProcessWithSpecialityStream`) do not modify existing methods
- New `Agent` field on `ToolEvent` is `omitempty`
- New WebSocket fields are backward-compatible
- Existing non-orchestrated agent loops are completely unaffected

### Phasing Strategy

| Phase | Effort | Impact | Dependencies |
|-------|--------|--------|--------------|
| Phase 1: Fix the Pipeline | ~2-3 sessions | Critical -- unblocks all visibility | None |
| Phase 2: Enrich Visibility | ~2-3 sessions | High -- delivers the user-facing value | Phase 1 |
| Phase 3: Token Optimization | ~2 sessions | Medium -- efficiency gains | Phase 1 (Phase 2 optional) |

Phase 1 should be implemented first as it unblocks both Phase 2 and Phase 3. Phase 2 and Phase 3 can then proceed in parallel.

---

## 9. Related Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Exploration | `openspec/changes/multi-agent-orchestration-visibility/exploration.md` | Complete |
| Proposal | `openspec/changes/multi-agent-orchestration-visibility/proposal.md` | This document |
| Specification | `openspec/changes/multi-agent-orchestration-visibility/spec.md` | Pending |
| Design | `openspec/changes/multi-agent-orchestration-visibility/design.md` | Pending |
| Tasks | `openspec/changes/multi-agent-orchestration-visibility/tasks.md` | Pending |

---

## 10. Approval

| Role | Name | Date | Decision |
|------|------|------|----------|
| Author | SDD Sub-Agent | 2026-02-27 | Proposed |
| Reviewer | | | |
| Approver | | | |

---

## Appendix A: Callback Propagation Chain (Before/After)

### Before (Current)

```
WebSocket handleChatWS()
  |-- ctx = ContextWithAgentStatusCallback(ctx, wsAgentStatusCb)
  |-- ctx = ContextWithContentSegmentCallback(ctx, wsContentSegmentCb)
  |-- activeAgentLoop.ProcessDirectWithUserAndModelStream(ctx, ..., onToken, onTool)
      |-- processMessageWithModelStream(ctx, ..., onToken, onTool)
          |-- runAgentLoopStream(ctx, processOptions{OnToken, OnTool})
              |-- LLM calls delegate_to_specialist
                  |-- DelegationTool.Execute(ctx)       // ctx has status+segment callbacks
                      |-- emitAgentStatus(ctx, ...)     // WORKS (uses ctx callbacks)
                      |-- processSpecialistTask(ctx, ...)
                          |-- specialist.ProcessWithSpeciality(ctx, task)
                              |-- ProcessDirect(ctx, ...)
                                  |-- processMessageWithModel(ctx, ...)
                                      |-- runAgentLoop(ctx, processOptions{})  // NO CALLBACKS
                                          |-- LLM runs silently
                                          |-- Tools run silently
                          |-- emitContentSegment(ctx, result)  // WORKS (final result only)
```

### After (Proposed)

```
WebSocket handleChatWS()
  |-- ctx = ContextWithAgentStatusCallback(ctx, wsAgentStatusCb)
  |-- ctx = ContextWithContentSegmentCallback(ctx, wsContentSegmentCb)
  |-- activeAgentLoop.ProcessDirectWithUserAndModelStream(ctx, ..., onToken, onTool)
      |-- processMessageWithModelStream(ctx, ..., onToken, onTool)
          |-- runAgentLoopStream(ctx, processOptions{OnToken, OnTool, OnAgentStatus, OnContentSegment})
              |-- LLM calls delegate_to_specialist
                  |-- DelegationTool.Execute(ctx)
                      |-- emitAgentStatus(ctx, "delegating")
                      |-- processSpecialistTask(ctx, ...)
                          |-- callbacks = extractCallbacksFromCtx(ctx)
                          |-- specialist.ProcessWithSpecialityStream(ctx, task, callbacks)
                              |-- filteredTools = sa.ToolFilter()  // Per-call copy, no mutex
                              |-- wrappedCallbacks = wrapWithAttribution(sa.name, callbacks)
                              |-- ProcessDirectWithCallbacks(ctx, ..., wrappedCallbacks)
                                  |-- processMessageWithModelStream(ctx, ..., wrappedOnToken, wrappedOnTool)
                                      |-- runAgentLoopStream(ctx, processOptions{ALL CALLBACKS})
                                          |-- LLM tokens stream to user (attributed)
                                          |-- Tool calls visible to user (attributed)
                          |-- emitContentSegment(ctx, result)  // Final result segment
```

## Appendix B: Concurrency Bug Detail

### Current Code (specialist.go lines 252-274)

```go
func (sa *SpecialistAgent) ProcessWithSpeciality(ctx context.Context, userMessage string) (string, error) {
    sa.processMu.Lock()
    originalTools := sa.tools
    sa.tools = sa.ToolFilter()
    sa.processMu.Unlock()

    defer func() {
        sa.processMu.Lock()
        sa.tools = originalTools
        sa.processMu.Unlock()
    }()

    return sa.ProcessDirect(agentCtx, fullMessage, ...)
}
```

### Race Scenario

```
Time  Goroutine A                          Goroutine B
----  -----------                          -----------
T1    Lock; originalA = sa.tools(full)
T2    sa.tools = filtered_A; Unlock
T3    Running with filtered_A...           Lock; originalB = sa.tools(filtered_A!) // WRONG
T4    Running...                           sa.tools = filtered_B; Unlock
T5    Running...                           Running with filtered_B...
T6    Lock; sa.tools = originalA(full); Unlock
T7                                         Running with filtered_B...
T8                                         Lock; sa.tools = originalB(filtered_A!); Unlock // CORRUPTED
```

At T8, `sa.tools` is set to `filtered_A` (what Goroutine B captured as "original"), which is wrong -- it should be the full tool set.

### Fix

Eliminate shared mutable state. Each call creates its own filtered registry copy and passes it as a parameter, never touching `sa.tools`.
