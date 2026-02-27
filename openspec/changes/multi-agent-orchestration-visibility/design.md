# Technical Design: Multi-Agent Orchestration Visibility

**Change ID:** multi-agent-orchestration-visibility
**Status:** Draft
**Author:** SDD Sub-Agent
**Created:** 2026-02-27
**Based on:** [exploration.md](exploration.md), [proposal.md](proposal.md)

---

## Table of Contents

1. [Design Overview](#1-design-overview)
2. [Architecture Decision Records](#2-architecture-decision-records)
3. [Backend Design: pkg/agent/](#3-backend-design-pkgagent)
4. [Backend Design: pkg/web/](#4-backend-design-pkgweb)
5. [Frontend Design: pkg/web/frontend/](#5-frontend-design-pkgwebfrontend)
6. [Sequence Diagrams](#6-sequence-diagrams)
7. [Data Flow and Protocol Schemas](#7-data-flow-and-protocol-schemas)
8. [Component Interaction Diagram](#8-component-interaction-diagram)
9. [Concurrency Design](#9-concurrency-design)
10. [Token Budget Tracking](#10-token-budget-tracking)
11. [Testing Strategy](#11-testing-strategy)
12. [Migration and Backward Compatibility](#12-migration-and-backward-compatibility)

---

## 1. Design Overview

### Problem Recap

The orchestrator-to-specialist delegation pipeline loses all streaming callbacks at the `ProcessDirect()` boundary. Specialists run their entire LLM loop silently because `ProcessDirect` constructs a `processOptions{}` with zero callbacks. The frontend has full plumbing for agent status, content segments, and tool call attribution, but these components are starved of events.

### Solution Architecture

Thread callbacks from the WebSocket handler through the orchestrator's `DelegationTool.Execute()` into the specialist's LLM loop by:

1. Adding a `DelegationCallbacks` struct that bundles all four callback types.
2. Adding `ProcessDirectWithCallbacks` on `AgentLoop` as an additive method.
3. Adding `ProcessWithSpecialityStream` on `SpecialistAgent` that creates per-call tool registry copies (fixing the concurrency bug) and wraps callbacks with agent attribution.
4. Extracting callbacks from `ctx` in `processSpecialistTask` and forwarding them.
5. Extending `ToolEvent` with an `Agent` field for attribution.
6. Enhancing the frontend to handle attributed tool calls and streaming agent status.

All changes are **additive**. Existing methods (`ProcessDirect`, `ProcessWithSpeciality`, `runAgentLoop`) remain untouched. New code paths are only activated when callbacks are present.

---

## 2. Architecture Decision Records

### ADR-1: ProcessDirectWithCallbacks vs Modifying ProcessDirect

**Context:** We need to pass callbacks into the specialist's LLM loop. Two approaches are possible: (a) add a new method `ProcessDirectWithCallbacks`, or (b) modify the existing `ProcessDirect` to accept optional callbacks.

**Decision:** Add a new `ProcessDirectWithCallbacks` method (additive approach).

**Rationale:**
- `ProcessDirect` is called from 5+ locations: `ProcessDirectWithChannel`, `ProcessDirectWithUser`, `ProcessDirectWithChannelForUser`, `ProcessWithSpeciality`, `ProcessOrchestratorMessage`. Modifying it requires touching all call sites.
- The existing `ProcessDirect` creates a `bus.InboundMessage` with channel `"cli"` and sender `"cron"` -- appropriate for its non-streaming use case. The new method needs different message construction.
- Additive approach means zero risk to existing non-orchestrated flows (CLI, cron, gateway, spawn).
- The new method can directly call `processMessageWithModelStream` which already handles streaming, rather than duplicating logic.

**Consequences:**
- Two similar methods exist; maintain parity via shared helper functions.
- New specialists always use the streaming path when callbacks are available.

---

### ADR-2: Per-Call Tool Registry Copies vs Shared Mutable State

**Context:** `ProcessWithSpeciality` currently swaps `sa.tools` (the shared field on `SpecialistAgent`) under a mutex, executes the LLM loop, then restores the original. This has a documented race condition under concurrent delegations (see proposal.md Appendix B).

**Decision:** Eliminate the shared mutable state entirely. Create a per-call filtered `ToolRegistry` copy and pass it as a parameter to the processing method, never touching `sa.tools`.

**Rationale:**
- The race condition is real and provable (Goroutine B captures filtered-A as "original").
- A per-call copy is cheap: `ToolFilter()` already creates a `NewToolRegistry()` and iterates the tools map. Tool objects themselves are shared (only registry metadata is copied).
- Eliminates the `processMu` mutex entirely -- no lock contention under concurrent delegations.
- Follows the principle of immutable data over shared mutable state.

**Consequences:**
- Need a new internal method on `AgentLoop` that accepts a `*tools.ToolRegistry` parameter override: `runAgentLoopStreamWithTools`.
- The specialist's base `sa.tools` field is never modified after construction. The `processMu` mutex can be removed.
- Memory overhead is negligible: each `ToolRegistry` copy is a `map[string]Tool` of pointers.

---

### ADR-3: Context-Embedded Callbacks vs Explicit Parameters for Streaming

**Context:** The existing `AgentStatusCallback` and `ContentSegmentCallback` are propagated via `context.WithValue`. The `OnToken` and `OnTool` callbacks are passed as explicit function parameters. Should we unify the approach?

**Decision:** Extract all callbacks from context at the orchestrator boundary and pass them as explicit parameters to the specialist via `DelegationCallbacks`. Keep context-based propagation for the WebSocket handler layer (where it already works), but use explicit parameters for the agent-to-agent boundary.

**Rationale:**
- The very bug we are fixing stems from implicit context-based callback propagation being incomplete -- `ProcessDirect` ignores context-embedded callbacks.
- Explicit parameters provide compile-time type safety and make the dependency visible in function signatures.
- The context approach works well at the WebSocket layer (injected once, read by `emitAgentStatus`/`emitContentSegment`) but poorly at the agent-loop layer (where `processOptions` is the callback carrier).
- Hybrid approach: context injection stays at the HTTP/WS boundary, explicit threading at the agent/specialist boundary.

**Consequences:**
- `processSpecialistTask` extracts callbacks from `ctx` and bundles them into `DelegationCallbacks`.
- The specialist does not need to know about context keys -- it receives explicit callbacks.

---

### ADR-4: WebSocket Event Format for Agent Metadata During Streaming

**Context:** During specialist streaming, the frontend needs to know which agent is producing tokens. Options: (a) add `agent` field to existing `stream` events, (b) emit `agent_status` before specialist tokens begin, (c) wrap specialist output in `content_segment` events.

**Decision:** Use a hybrid approach:
1. Emit `agent_status` with `status: "streaming"` when a specialist begins its LLM response.
2. Add optional `agent` field to `tool_call` events for attribution.
3. Keep `stream` events unchanged (no `agent` field) -- the current `agent_status` state tells the frontend which agent is active.
4. Emit `content_segment` with the complete specialist result when the specialist finishes.

**Rationale:**
- Adding `agent` to every `stream` token event would increase WebSocket traffic by ~15-20 bytes per token with no visual benefit (the frontend already knows which agent is active from `agent_status`).
- The `agent_status` event with `status: "streaming"` naturally pairs with existing frontend state (`orchestratorStatus`, `currentAgent`) to attribute tokens.
- `content_segment` events serve double duty: they provide the attributed content block for message history/rendering AND trigger the frontend to associate accumulated tokens with that agent.
- `tool_call` events need the `agent` field because tool calls can be interleaved (especially with future parallel delegation).

**Consequences:**
- Frontend uses `currentAgent` from `agent_status` to visually attribute streaming tokens.
- Historical messages use `segments` array for post-hoc attribution.
- The `stream` event schema is unchanged -- full backward compatibility.

---

### ADR-5: Content Segment Protocol (Frontend Accumulation and Rendering)

**Context:** When a specialist streams tokens, the frontend accumulates them in `msg.content`. When the specialist finishes, a `content_segment` event arrives with the complete text. How should the frontend reconcile streaming content with segment attribution?

**Decision:** Implement a **streaming segment buffer** pattern:

1. When `agent_status` with `status: "streaming"` arrives for a specialist, the chatStore opens a new streaming segment buffer.
2. Incoming `stream` tokens are appended BOTH to `msg.content` (for display) AND to the active segment buffer (for attribution).
3. When `content_segment` arrives, it replaces the buffer with the authoritative content and closes the segment.
4. When `agent_status` with `status: "complete"` arrives, any open buffer is flushed.
5. The orchestrator's own tokens (synthesis/summary after specialist work) are attributed to "orchestrator" using the same mechanism.

**Rationale:**
- Users see tokens in real-time (responsive UX).
- After streaming completes, segments provide accurate attribution for message history.
- The `content_segment` event is authoritative (server-computed), so it resolves any accumulation drift.
- Works cleanly for both single-specialist and future multi-specialist parallel scenarios.

**Consequences:**
- New chatStore state: `activeStreamingSegment` (tracks which agent is currently streaming).
- `appendStreamToken` checks if a segment is active and buffers accordingly.
- Rendering uses `segments` array when available, `msg.content` as fallback (no change to MessageBubble logic).

---

### ADR-6: Parallel Delegation -- Goroutine Management, Result Merging, Error Handling

**Context:** The proposal includes parallel delegation as Phase 3. This ADR records the design decisions for when it is implemented.

**Decision:** Implement parallel delegation as a new `delegate_to_specialists` tool (plural) alongside the existing `delegate_to_specialist` tool. Design choices:

1. **Goroutine management:** Use `errgroup.Group` with a configurable concurrency limit (`MaxParallelDelegations`, default 3). Each specialist runs in its own goroutine.
2. **Result merging:** Results are collected into a `[]DelegationResult` slice. The ordering matches the request order (not completion order). Results are returned as a JSON array to the orchestrator LLM.
3. **Error handling:** Partial failures are tolerated. Each `DelegationResult` has its own `Success` and `Error` fields. The orchestrator LLM sees all results (successes and failures) and decides how to proceed.
4. **Callback multiplexing:** Each parallel specialist wraps callbacks with its agent name. A `sync.Mutex` protects the shared WebSocket writer (already in place). Token interleaving is acceptable because `agent_status` events delineate which specialist is active. For cleaner UX, the frontend can buffer parallel tokens per-agent and display them in separate segments.
5. **Context propagation:** Each goroutine gets its own child context (for independent cancellation) but shares the parent context's callbacks.

**Rationale:**
- `errgroup.Group` is the standard Go pattern for concurrent work with error collection and concurrency limiting.
- Separate tool name (`delegate_to_specialists`) means the LLM explicitly opts into parallel mode -- no accidental parallel execution.
- Partial failure tolerance matches real-world behavior: one specialist may fail while others succeed.
- Concurrency limit prevents resource exhaustion (each specialist creates its own LLM connection).

**Consequences:**
- New config field: `OrchestratorConfig.MaxParallelDelegations int` (default 3).
- New tool: `delegate_to_specialists` registered only on the orchestrator.
- Frontend needs no changes for Phase 3 -- the per-agent segment attribution from Phase 2 handles interleaved content naturally.

---

### ADR-7: Token Budget Tracking -- Where to Measure, How to Report

**Context:** Token usage per specialist call is currently invisible. The proposal calls for adding `TokensUsed` to `DelegationResult`.

**Decision:** Capture token usage from the LLM provider's response and propagate it through `DelegationResult`. Measurement points:

1. **Where to measure:** In `runAgentLoopStream`/`runAgentLoop`, the `observability.Global().RecordLLMCall()` already computes `tokensIn` and `tokensOut`. We will also capture these in the `processOptions` via a new `TokenUsage` accumulator field.
2. **How to accumulate:** A `TokenUsageAccumulator` struct with an `atomic.Int64` for thread-safe accumulation across tool-calling iterations.
3. **How to report:** `ProcessDirectWithCallbacks` returns `(string, *TokenUsage, error)`. The `DelegationResult` includes the total `TokensUsed`. The `stream_end` WebSocket message includes a `token_usage` field. The `agent_status` event with `status: "complete"` includes `tokens_used`.
4. **Provider-accurate vs estimated:** When the provider returns `UsageInfo` (on the final streaming chunk), use the provider's numbers. Otherwise, fall back to the `len(content)/4` estimation already used.

**Rationale:**
- Provider-accurate token counts are already available in `StreamChunk.Usage` for providers that support it (OpenAI, Claude).
- Accumulating across iterations handles multi-turn tool-calling conversations.
- Atomic operations allow safe accumulation without additional mutexes.
- Informational in v1 (no enforcement) avoids complex budget logic.

**Consequences:**
- New `TokenUsage` struct in `pkg/agent/loop.go`.
- `DelegationResult` gains `TokensUsed` field.
- `stream_end` WebSocket message gains optional `token_usage` field (backward-compatible).

---

## 3. Backend Design: pkg/agent/

### 3.1 New Types

```go
// File: pkg/agent/loop.go (additions)

// DelegationCallbacks bundles all callbacks for specialist delegation.
// All fields are optional -- nil callbacks are simply not invoked.
type DelegationCallbacks struct {
    OnToken          StreamCallback
    OnTool           ToolCallback
    OnAgentStatus    AgentStatusCallback
    OnContentSegment ContentSegmentCallback
    SkipHistory      bool // If true, specialist does not load session history
}

// TokenUsage tracks token consumption for a processing call.
type TokenUsage struct {
    InputTokens  int64 `json:"input_tokens"`
    OutputTokens int64 `json:"output_tokens"`
    TotalTokens  int64 `json:"total_tokens"`
}
```

### 3.2 Extended ToolEvent

```go
// File: pkg/agent/loop.go (modification)

type ToolEvent struct {
    Name   string                 `json:"name"`
    Args   map[string]interface{} `json:"arguments"`
    Result string                 `json:"result,omitempty"`
    Status string                 `json:"status"` // "started", "finished", "error"
    Agent  string                 `json:"agent,omitempty"` // NEW: agent that invoked the tool
}
```

The `Agent` field is `omitempty` -- existing code that constructs `ToolEvent` without setting `Agent` produces identical JSON output.

### 3.3 New Method: AgentLoop.ProcessDirectWithCallbacks

```go
// File: pkg/agent/loop.go (addition)

// ProcessDirectWithCallbacks processes a message with full callback support.
// This is the entry point for specialist delegation where streaming visibility
// is needed. Unlike ProcessDirect, this routes to the streaming path and
// forwards all callbacks to the inner LLM loop.
//
// If toolsOverride is non-nil, it is used instead of al.tools for this call only.
// This enables per-call tool filtering without mutating shared state.
func (al *AgentLoop) ProcessDirectWithCallbacks(
    ctx context.Context,
    content string,
    sessionKey string,
    callbacks DelegationCallbacks,
    toolsOverride *tools.ToolRegistry,
) (string, *TokenUsage, error) {
    msg := bus.InboundMessage{
        Channel:    "specialist",
        SenderID:   "orchestrator",
        ChatID:     "delegation",
        Content:    content,
        SessionKey: sessionKey,
    }

    // Build processOptions with all callbacks
    opts := processOptions{
        SessionKey:       sessionKey,
        Channel:          msg.Channel,
        ChatID:           msg.ChatID,
        UserMessage:      content,
        DefaultResponse:  "",
        EnableSummary:    false, // Specialists don't trigger summarization
        SendResponse:     false,
        OnToken:          callbacks.OnToken,
        OnTool:           callbacks.OnTool,
        OnAgentStatus:    callbacks.OnAgentStatus,
        OnContentSegment: callbacks.OnContentSegment,
    }

    // Use the streaming loop if callbacks are provided and provider supports it
    if callbacks.OnToken != nil {
        if _, canStream := al.provider.(providers.StreamingLLMProvider); canStream {
            result, err := al.runAgentLoopStreamWithTools(ctx, opts, callbacks.OnToken, toolsOverride, callbacks.SkipHistory)
            // Token usage extracted from observability (phase 3 enhancement)
            return result, nil, err
        }
    }

    // Fallback to non-streaming with tool override
    result, err := al.runAgentLoopWithTools(ctx, opts, toolsOverride, callbacks.SkipHistory)
    return result, nil, err
}
```

### 3.4 New Internal Methods: runAgentLoopStreamWithTools and runAgentLoopWithTools

These are thin wrappers around the existing `runAgentLoopStream`/`runAgentLoop` that accept a `toolsOverride` parameter. The core logic is identical -- the only difference is which tool registry is used for `GetDefinitions()` and `ExecuteWithContext()`.

```go
// File: pkg/agent/loop.go (addition)

// runAgentLoopStreamWithTools is like runAgentLoopStream but uses a custom tool registry
// and optionally skips session history loading (for lightweight specialist execution).
func (al *AgentLoop) runAgentLoopStreamWithTools(
    ctx context.Context,
    opts processOptions,
    onToken StreamCallback,
    toolsOverride *tools.ToolRegistry,
    skipHistory bool,
) (string, error) {
    // Clear/register involved agents
    al.ClearInvolvedAgents()
    agentName := al.model
    if agentName == "" {
        agentName = "main"
    }
    al.AddInvolvedAgent(agentName)

    agentStart := time.Now()

    // Determine which tool registry to use
    effectiveTools := al.tools
    if toolsOverride != nil {
        effectiveTools = toolsOverride
    }

    // Update tool contexts
    al.updateToolContextsForRegistry(effectiveTools, opts.Channel, opts.ChatID)

    // Build messages (optionally skip history)
    var history []session.Message
    var summary string
    if !skipHistory {
        history = al.sessions.GetHistoryForUser(al.userID, opts.SessionKey)
        summary = al.sessions.GetSummaryForUser(al.userID, opts.SessionKey)
    }
    messages := al.contextBuilder.BuildMessages(
        history, summary, opts.UserMessage, nil, opts.Channel, opts.ChatID,
    )

    // Run LLM iteration loop with streaming and tool override
    finalContent, iteration, err := al.runLLMIterationStreamWithTools(
        ctx, messages, opts, onToken, effectiveTools,
    )
    observability.Global().RecordAgentRun(time.Since(agentStart), iteration, err)
    if err != nil {
        return "", err
    }

    if finalContent == "" {
        finalContent = opts.DefaultResponse
    }

    // Log response
    responsePreview := utils.Truncate(finalContent, 120)
    logger.InfoCF("agent", fmt.Sprintf("Specialist streaming response: %s", responsePreview),
        map[string]interface{}{
            "session_key":  opts.SessionKey,
            "iterations":   iteration,
            "final_length": len(finalContent),
        })

    return finalContent, nil
}
```

The `runLLMIterationStreamWithTools` method is identical to `runLLMIterationStream` except it uses `effectiveTools.GetDefinitions()` and `effectiveTools.ExecuteWithContext()` instead of `al.tools`. This can be implemented either as a parameter to the existing method (refactoring `runLLMIterationStream` to accept a tools parameter) or as a separate wrapper. The refactoring approach is preferred to avoid code duplication:

```go
// Refactored signature (internal):
func (al *AgentLoop) runLLMIterationStream(
    ctx context.Context,
    messages []providers.Message,
    opts processOptions,
    onToken StreamCallback,
) (string, int, error) {
    return al.runLLMIterationStreamImpl(ctx, messages, opts, onToken, al.tools)
}

func (al *AgentLoop) runLLMIterationStreamWithTools(
    ctx context.Context,
    messages []providers.Message,
    opts processOptions,
    onToken StreamCallback,
    tools *tools.ToolRegistry,
) (string, int, error) {
    return al.runLLMIterationStreamImpl(ctx, messages, opts, onToken, tools)
}

// runLLMIterationStreamImpl contains the actual implementation.
// The existing body of runLLMIterationStream moves here, with al.tools replaced by the tools parameter.
func (al *AgentLoop) runLLMIterationStreamImpl(
    ctx context.Context,
    messages []providers.Message,
    opts processOptions,
    onToken StreamCallback,
    effectiveTools *tools.ToolRegistry,
) (string, int, error) {
    // ... existing body, but using effectiveTools.GetDefinitions()
    //     and effectiveTools.ExecuteWithContext() instead of al.tools
}
```

### 3.5 SpecialistAgent Changes

#### 3.5.1 New Method: ProcessWithSpecialityStream

```go
// File: pkg/agent/specialist.go (addition)

// ProcessWithSpecialityStream processes a message using this specialist's configuration
// with full callback support for real-time streaming visibility.
//
// Unlike ProcessWithSpeciality, this method:
// 1. Creates a per-call tool registry copy (no shared mutable state, no mutex).
// 2. Wraps callbacks with agent attribution (specialist name prefix).
// 3. Routes to ProcessDirectWithCallbacks for streaming support.
// 4. Optionally skips session history for lightweight execution.
func (sa *SpecialistAgent) ProcessWithSpecialityStream(
    ctx context.Context,
    userMessage string,
    callbacks DelegationCallbacks,
) (string, *TokenUsage, error) {
    // 1. Create per-call tool registry copy (concurrency-safe, no mutex needed)
    filteredTools := sa.ToolFilter()

    // 2. Wrap callbacks with agent attribution
    wrappedCallbacks := sa.wrapCallbacksWithAttribution(callbacks)

    // 3. Inject specialist name into context for tools
    agentCtx := kakoclawContext.WithAgentName(ctx, sa.name)

    // 4. Emit "streaming" status before LLM call
    if wrappedCallbacks.OnAgentStatus != nil {
        _ = wrappedCallbacks.OnAgentStatus(AgentStatusEvent{
            Agent:     sa.name,
            Status:    "streaming",
            Timestamp: time.Now(),
        })
    }

    // 5. Process with full callbacks and per-call tool registry
    sessionKey := fmt.Sprintf("specialist_%s", sa.name)
    wrappedCallbacks.SkipHistory = true // Specialists don't load conversation history
    result, usage, err := sa.ProcessDirectWithCallbacks(
        agentCtx, userMessage, sessionKey, wrappedCallbacks, filteredTools,
    )

    return result, usage, err
}

// wrapCallbacksWithAttribution creates callback wrappers that annotate events
// with this specialist's name for frontend attribution.
func (sa *SpecialistAgent) wrapCallbacksWithAttribution(callbacks DelegationCallbacks) DelegationCallbacks {
    wrapped := DelegationCallbacks{
        OnAgentStatus:    callbacks.OnAgentStatus,
        OnContentSegment: callbacks.OnContentSegment,
        SkipHistory:      callbacks.SkipHistory,
    }

    // Wrap OnToken: tokens pass through unchanged.
    // Attribution is handled by the agent_status "streaming" event.
    wrapped.OnToken = callbacks.OnToken

    // Wrap OnTool: add agent name to tool events.
    if callbacks.OnTool != nil {
        specialistName := sa.name
        wrapped.OnTool = func(ev ToolEvent) error {
            ev.Agent = specialistName
            return callbacks.OnTool(ev)
        }
    }

    return wrapped
}
```

#### 3.5.2 Removing the processMu Mutex

The existing `ProcessWithSpeciality` method remains unchanged (backward-compatible), but a warning comment is added noting that `ProcessWithSpecialityStream` is the preferred path for concurrent use. In a future cleanup, `ProcessWithSpeciality` can be updated to also use per-call copies.

```go
// ProcessWithSpeciality processes a message using this specialist's configuration.
// DEPRECATED for concurrent use: prefer ProcessWithSpecialityStream which creates
// per-call tool registry copies instead of swapping the shared field.
// This method is retained for backward compatibility with non-streaming paths.
```

### 3.6 Orchestrator Changes

#### 3.6.1 Modified processSpecialistTask

```go
// File: pkg/agent/orchestrator.go (modification)

func (oa *OrchestratorAgent) processSpecialistTask(ctx context.Context, specialistName, task string) (string, error) {
    specialist, err := oa.registry.GetSpecialist(specialistName)
    if err != nil {
        return "", err
    }

    timeout := 5 * time.Minute
    ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    // Extract callbacks from context for specialist forwarding
    callbacks := extractDelegationCallbacks(ctx)

    resultChan := make(chan delegationResultInternal, 1)
    errChan := make(chan error, 1)

    go func() {
        if callbacks.OnToken != nil {
            // Use streaming path with full callback support
            result, usage, err := specialist.ProcessWithSpecialityStream(ctxWithTimeout, task, callbacks)
            if err != nil {
                errChan <- err
                return
            }
            resultChan <- delegationResultInternal{Result: result, Usage: usage}
        } else {
            // Fallback to non-streaming path (backward-compatible)
            result, err := specialist.ProcessWithSpeciality(ctxWithTimeout, task)
            if err != nil {
                errChan <- err
                return
            }
            resultChan <- delegationResultInternal{Result: result}
        }
    }()

    var result string
    select {
    case res := <-resultChan:
        result = res.Result
        // Emit content segment with final result
        emitContentSegment(ctx, ContentSegment{
            Agent:     specialistName,
            Content:   result,
            SegmentID: fmt.Sprintf("seg_%s_%d", specialistName, time.Now().UnixNano()),
            Timestamp: time.Now(),
        })

    case err := <-errChan:
        // Emit error status
        emitAgentStatus(ctx, AgentStatusEvent{
            Agent:     specialistName,
            Status:    "error",
            Reason:    err.Error(),
            Timestamp: time.Now(),
        })
        return "", fmt.Errorf("specialist processing error: %w", err)

    case <-ctxWithTimeout.Done():
        emitAgentStatus(ctx, AgentStatusEvent{
            Agent:     specialistName,
            Status:    "timeout",
            Timestamp: time.Now(),
        })
        return "", fmt.Errorf("specialist %s timed out after %v", specialistName, timeout)
    }

    // Register involved agents
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

// delegationResultInternal is used internally to pass results from the goroutine.
type delegationResultInternal struct {
    Result string
    Usage  *TokenUsage
}

// extractDelegationCallbacks extracts all available callbacks from the context
// and bundles them into a DelegationCallbacks struct.
func extractDelegationCallbacks(ctx context.Context) DelegationCallbacks {
    return DelegationCallbacks{
        OnAgentStatus:    agentStatusCallbackFromCtx(ctx),
        OnContentSegment: contentSegmentCallbackFromCtx(ctx),
        // OnToken and OnTool are not in ctx -- they come from processOptions
        // which are only available inside the agent loop. We need a different
        // mechanism to forward these. See section 3.7.
    }
}
```

#### 3.6.2 Forwarding OnToken and OnTool from processOptions

The `OnToken` and `OnTool` callbacks live in `processOptions` (set by `processMessageWithModelStream`), not in `ctx`. The `DelegationTool.Execute()` method has access to `ctx` but NOT to `processOptions`. This is the core design challenge.

**Solution: Add OnToken and OnTool to context alongside AgentStatusCallback and ContentSegmentCallback.**

New context keys and helper functions:

```go
// File: pkg/agent/orchestrator.go (additions)

type streamCallbackKey struct{}
type toolCallbackKey struct{}

// ContextWithStreamCallback embeds a StreamCallback into the context
func ContextWithStreamCallback(ctx context.Context, callback StreamCallback) context.Context {
    return context.WithValue(ctx, streamCallbackKey{}, callback)
}

func streamCallbackFromCtx(ctx context.Context) StreamCallback {
    if v, ok := ctx.Value(streamCallbackKey{}).(StreamCallback); ok {
        return v
    }
    return nil
}

// ContextWithToolCallback embeds a ToolCallback into the context
func ContextWithToolCallback(ctx context.Context, callback ToolCallback) context.Context {
    return context.WithValue(ctx, toolCallbackKey{}, callback)
}

func toolCallbackFromCtx(ctx context.Context) ToolCallback {
    if v, ok := ctx.Value(toolCallbackKey{}).(ToolCallback); ok {
        return v
    }
    return nil
}
```

**Injection point:** In `runAgentLoopStream` (and the new `runAgentLoopStreamWithTools`), before executing tool calls, inject the `OnToken` and `OnTool` callbacks into the context:

```go
// In runAgentLoopStreamWithTools, at the beginning:
if opts.OnToken != nil {
    ctx = ContextWithStreamCallback(ctx, opts.OnToken)
}
if opts.OnTool != nil {
    ctx = ContextWithToolCallback(ctx, opts.OnTool)
}
```

Then `extractDelegationCallbacks` becomes:

```go
func extractDelegationCallbacks(ctx context.Context) DelegationCallbacks {
    return DelegationCallbacks{
        OnToken:          streamCallbackFromCtx(ctx),
        OnTool:           toolCallbackFromCtx(ctx),
        OnAgentStatus:    agentStatusCallbackFromCtx(ctx),
        OnContentSegment: contentSegmentCallbackFromCtx(ctx),
    }
}
```

**Why this is acceptable despite ADR-3 preferring explicit parameters:** The callbacks cross the `Tool.Execute()` boundary (which has a fixed `(ctx, args)` signature we cannot change). Context is the only propagation mechanism across the Tool interface. Once extracted by `processSpecialistTask`, they become explicit parameters via `DelegationCallbacks`.

### 3.7 Context Trimming Strategy for Specialists

Specialists should execute with minimal context to save tokens. The strategy:

1. **SkipHistory = true** (default for specialist calls): Do not load session history from the `SessionManager`. The specialist's session key (`specialist_<name>`) has no accumulated history anyway (ephemeral sessions).

2. **Lightweight mode** (already implemented): `SetLightweightMode(true)` on the specialist's `ContextBuilder` skips bootstrap files (AGENTS.md, SOUL.md, USER.md, IDENTITY.md) and memory context.

3. **Task-only message**: The specialist receives only its system prompt + the single task message from the orchestrator. No conversation history.

4. **Specialist system prompt trimming** (future enhancement): Currently the full `getIdentity()` header is included even in lightweight mode. A future optimization can add a `SetMinimalIdentity()` mode that strips the runtime/OS information from the specialist's system prompt.

---

## 4. Backend Design: pkg/web/

### 4.1 Enhanced WebSocket Streaming Protocol

The WebSocket handler in `server.go` already injects `AgentStatusCallback` and `ContentSegmentCallback` into `ctx`. With the changes in section 3.6.2, it also needs to inject `StreamCallback` and `ToolCallback`:

```go
// File: pkg/web/server.go (modification to handleChatWS streaming path)

// Existing: inject into ctx
ctx = agent.ContextWithAgentStatusCallback(ctx, wsAgentStatusCb)
ctx = agent.ContextWithContentSegmentCallback(ctx, wsContentSegmentCb)

// NEW: also inject streaming callbacks into ctx for specialist forwarding
ctx = agent.ContextWithStreamCallback(ctx, func(token string) error {
    wsMu.Lock()
    defer wsMu.Unlock()
    return conn.WriteJSON(map[string]interface{}{
        "type":    "stream",
        "content": token,
    })
})
ctx = agent.ContextWithToolCallback(ctx, func(ev agent.ToolEvent) error {
    wsMu.Lock()
    defer wsMu.Unlock()
    msg := map[string]interface{}{
        "type":   "tool_call",
        "name":   ev.Name,
        "args":   ev.Args,
        "result": ev.Result,
        "status": ev.Status,
    }
    if ev.Agent != "" {
        msg["agent"] = ev.Agent
    }
    return conn.WriteJSON(msg)
})
```

Note: The `onToken` and `onTool` callbacks passed directly to `ProcessDirectWithUserAndModelStream` remain as-is for the orchestrator's own LLM loop. The context-embedded versions are forwarded to specialists. Both write to the same WebSocket connection.

### 4.2 New WebSocket Event Types

No new event types are needed. The existing types are extended:

| Event Type | Existing Fields | New Fields | Description |
|---|---|---|---|
| `tool_call` | `name`, `args`, `result`, `status` | `agent` (optional) | Agent that invoked the tool |
| `agent_status` | `agent`, `status`, `specialist_name`, `reason`, `timestamp` | _(no change)_ | New status value `"streaming"` added |
| `stream_end` | `content`, `agents`, `error` | `token_usage` (optional, Phase 3) | Token usage summary |

New `agent_status.status` values:

| Status | Meaning | When emitted |
|---|---|---|
| `analyzing` | Orchestrator is analyzing the request | At delegation start |
| `delegating` | Orchestrator is delegating to a specialist | Before specialist call |
| `working` | Specialist is processing (pre-LLM) | Before specialist LLM call |
| `streaming` | Specialist's LLM is producing tokens | When specialist starts streaming |
| `complete` | Specialist finished successfully | After specialist returns result |
| `error` | Specialist encountered an error | On specialist failure |
| `timeout` | Specialist timed out | On specialist timeout |

### 4.3 API Changes for Agent Metadata in Chat History

The chat history API already stores agent metadata in the `metadata` JSON field. No schema changes are needed. The `metadata.agents` array continues to work as-is. The `content_segment` data is not persisted in the database (it is derived from the message content and agents list at render time). This is by design -- segments are a streaming-time optimization for real-time attribution, not a persistence concern.

---

## 5. Frontend Design: pkg/web/frontend/

### 5.1 Dynamic SpecialistBadge Component

**File:** `pkg/web/frontend/src/components/Chat/SpecialistBadge.vue`

Replace the hardcoded 6-type color/icon map with a deterministic hash-based approach that falls back to known types.

```javascript
// New helper function
function hashCode(str) {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = ((hash << 5) - hash) + char
    hash |= 0 // Convert to 32-bit integer
  }
  return Math.abs(hash)
}

// Tailwind color palette for dynamic badges (12 distinct hues)
const dynamicPalette = [
  { color: 'text-rose-400',    bg: 'bg-rose-500/10' },
  { color: 'text-orange-400',  bg: 'bg-orange-500/10' },
  { color: 'text-yellow-400',  bg: 'bg-yellow-500/10' },
  { color: 'text-lime-400',    bg: 'bg-lime-500/10' },
  { color: 'text-emerald-400', bg: 'bg-emerald-500/10' },
  { color: 'text-teal-400',    bg: 'bg-teal-500/10' },
  { color: 'text-sky-400',     bg: 'bg-sky-500/10' },
  { color: 'text-indigo-400',  bg: 'bg-indigo-500/10' },
  { color: 'text-violet-400',  bg: 'bg-violet-500/10' },
  { color: 'text-fuchsia-400', bg: 'bg-fuchsia-500/10' },
  { color: 'text-pink-400',    bg: 'bg-pink-500/10' },
  { color: 'text-red-400',     bg: 'bg-red-500/10' },
]

const colors = computed(() => {
  // Known types (backward-compatible)
  const knownMap = {
    developer:     { color: 'text-blue-400',   bg: 'bg-blue-500/10' },
    documentation: { color: 'text-blue-400',   bg: 'bg-blue-500/10' },
    testing:       { color: 'text-amber-400',  bg: 'bg-amber-500/10' },
    devops:        { color: 'text-purple-400', bg: 'bg-purple-500/10' },
    analyst:       { color: 'text-cyan-400',   bg: 'bg-cyan-500/10' },
    researcher:    { color: 'text-pink-400',   bg: 'bg-pink-500/10' },
    orchestrator:  { color: 'text-emerald-400', bg: 'bg-emerald-500/10' },
  }

  if (knownMap[props.name]) return knownMap[props.name]

  // Dynamic hash-based color for unknown specialist names
  const idx = hashCode(props.name) % dynamicPalette.length
  return dynamicPalette[idx]
})

const icon = computed(() => {
  const knownIcons = {
    developer: '<path .../>',     // existing
    documentation: '<path .../>',  // existing
    testing: '<path .../>',        // existing
    devops: '<path .../>',         // existing
    analyst: '<path .../>',        // existing
    researcher: '<path .../>',     // existing
    orchestrator: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />',
  }

  if (knownIcons[props.name]) return knownIcons[props.name]

  // Dynamic icon: first letter of name in a circle
  // Falls back to the generic user icon
  return '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />'
})
```

### 5.2 Agent Timeline Component

**New file:** `pkg/web/frontend/src/components/Chat/AgentTimeline.vue`

A collapsible timeline that shows the sequence of agent events alongside a message. Used in `MessageBubble.vue` when `msg.agents.length > 1` (multi-agent response).

```
+-------------------------------------------+
| > Orchestration Timeline                  |
|                                           |
|  [O] Analyzing request         00:00      |
|  |                                        |
|  [O] Delegating to developer   00:01      |
|  |   "requires code implementation"       |
|  |                                        |
|  [D] developer is working      00:01      |
|  |   > read_file (main.go)     00:02      |
|  |   > write_file (handler.go) 00:04      |
|  |                                        |
|  [D] developer complete        00:08      |
|  |                                        |
|  [O] Synthesizing response     00:08      |
+-------------------------------------------+
```

**Props:**
- `events: AgentHistoryEvent[]` -- timeline events from `chatStore.agentHistory`
- `collapsed: boolean` -- default `true` (user clicks to expand)

**Data structure:**

```typescript
interface AgentHistoryEvent {
  type: 'status' | 'tool_call'
  agent: string
  status?: string         // for type='status'
  tool?: string          // for type='tool_call'
  reason?: string
  timestamp: string      // ISO 8601
}
```

### 5.3 ChatStore Changes for Real-Time Agent Tracking

**File:** `pkg/web/frontend/src/stores/chatStore.js`

#### New State

```javascript
const activeStreamingAgent = ref(null)    // Which agent is currently streaming tokens
const streamingSegmentBuffer = ref('')     // Buffer for current agent's streaming tokens
```

#### Modified: setAgentStatus

```javascript
function setAgentStatus(agent, status, specialistName = null, reason = '') {
    orchestratorStatus.value = status
    currentAgent.value = agent

    if (specialistName) {
        activeSpecialist.value = specialistName
    }
    if (reason) {
        delegationReason.value = reason
    }

    // Track streaming agent for token attribution
    if (status === 'streaming') {
        activeStreamingAgent.value = agent
        streamingSegmentBuffer.value = ''
    } else if (status === 'complete' || status === 'error' || status === 'timeout') {
        // Flush any buffered segment content
        if (activeStreamingAgent.value === agent) {
            activeStreamingAgent.value = null
            streamingSegmentBuffer.value = ''
        }
    }

    agentHistory.value.push({
        type: 'status',
        agent,
        status,
        specialistName,
        reason,
        timestamp: new Date().toISOString()
    })
}
```

#### Modified: addToolCall

```javascript
function addToolCall(toolCall) {
    if (!streamingMessageId.value) return
    const msg = messages.value.find(m => m.id === streamingMessageId.value)
    if (msg) {
        if (!msg.toolCalls) msg.toolCalls = []

        const existingIdx = msg.toolCalls.findLastIndex(
            tc => tc.name === toolCall.name && tc.status === 'started'
        )
        if (existingIdx !== -1 && toolCall.status !== 'started') {
            msg.toolCalls[existingIdx] = { ...msg.toolCalls[existingIdx], ...toolCall }
        } else {
            msg.toolCalls.push({
                ...toolCall,
                id: Date.now() + Math.random(),
                timestamp: new Date().toISOString()
            })
        }

        // NEW: Track tool calls in agent history for timeline
        if (toolCall.agent) {
            agentHistory.value.push({
                type: 'tool_call',
                agent: toolCall.agent,
                tool: toolCall.name,
                status: toolCall.status,
                timestamp: new Date().toISOString()
            })
        }
    }
}
```

### 5.4 MessageBubble Enhancements for Segmented Content

The existing `MessageBubble.vue` already handles segmented content correctly. The only change is to show tool calls with agent attribution:

```vue
<!-- Tool Calls Rendering (modification) -->
<div v-if="msg.toolCalls && msg.toolCalls.length > 0" class="mb-4 space-y-2">
  <ToolCallItem
    v-for="tc in msg.toolCalls"
    :key="tc.id"
    :tc="tc"
    :show-agent="msg.agents && msg.agents.length > 1"
  />
</div>
```

The `ToolCallItem` component receives a new optional `showAgent` prop. When true, it displays the agent name before the tool call using `SpecialistBadge`.

### 5.5 ChatView.vue WebSocket Handler Changes

No changes needed. The existing handler already processes:
- `agent_status` events -> `chatStore.setAgentStatus()`
- `content_segment` events -> `chatStore.addContentSegment()`
- `tool_call` events -> `chatStore.addToolCall()`

The new `agent` field on `tool_call` events is automatically passed through. The new `streaming` status on `agent_status` events is handled by the updated `setAgentStatus`.

---

## 6. Sequence Diagrams

### 6.1 Current Flow (Broken Pipeline)

```
User          WebSocket      AgentLoop(O)     DelegationTool     SpecialistAgent     AgentLoop(S)
 |               |               |                 |                   |                  |
 |--message----->|               |                 |                   |                  |
 |               |--ProcessDirectWithUserAndModelStream(onToken,onTool)|                  |
 |               |               |                 |                   |                  |
 |               |  ctx+=AgentStatusCb             |                   |                  |
 |               |  ctx+=ContentSegmentCb          |                   |                  |
 |               |               |                 |                   |                  |
 |               |               |--runAgentLoopStream(opts{OnToken,OnTool})              |
 |               |               |   LLM decides: call delegate_to_specialist             |
 |               |               |                 |                   |                  |
 |               |               |----Execute(ctx)-->                  |                  |
 |               |               |                 |                   |                  |
 |  <--agent_status("analyzing")-|                 |  (via ctx)        |                  |
 |  <--agent_status("delegating")|                 |  (via ctx)        |                  |
 |  <--agent_status("working")---|                 |  (via ctx)        |                  |
 |               |               |                 |                   |                  |
 |               |               |                 |--processSpecialistTask(ctx)          |
 |               |               |                 |                   |                  |
 |               |               |                 |  specialist.ProcessWithSpeciality(ctx, task)
 |               |               |                 |                   |                  |
 |               |               |                 |                   |--ProcessDirect() |
 |               |               |                 |                   |  processMessage()|
 |               |               |                 |                   |  runAgentLoop(processOptions{})
 |               |               |                 |                   |                  |
 |               |               |                 |                   |  NO CALLBACKS!   |
 |               |               |                 |                   |  LLM runs silent |
 |               |               |                 |                   |  Tools run silent|
 |               |               |                 |                   |                  |
 |               |               |                 |  <--result string--|                  |
 |               |               |                 |                   |                  |
 |  <--content_segment(result)---|                 |  (via ctx, final) |                  |
 |  <--agent_status("complete")--|                 |  (via ctx)        |                  |
 |               |               |                 |                   |                  |
 |               |               |  <--tool result-|                   |                  |
 |               |               |                 |                   |                  |
 |               |               |--LLM synthesizes final response     |                  |
 |  <--stream tokens-------------|                 |                   |                  |
 |  <--stream_end(content,agents)|                 |                   |                  |
```

### 6.2 Proposed Flow (Fixed Pipeline)

```
User          WebSocket      AgentLoop(O)     DelegationTool     SpecialistAgent     AgentLoop(S)
 |               |               |                 |                   |                  |
 |--message----->|               |                 |                   |                  |
 |               |--ProcessDirectWithUserAndModelStream(onToken,onTool)|                  |
 |               |               |                 |                   |                  |
 |               |  ctx+=AgentStatusCb             |                   |                  |
 |               |  ctx+=ContentSegmentCb          |                   |                  |
 |               |  ctx+=StreamCb (NEW)            |                   |                  |
 |               |  ctx+=ToolCb (NEW)              |                   |                  |
 |               |               |                 |                   |                  |
 |               |               |--runAgentLoopStreamWithTools(opts{ALL callbacks})      |
 |               |               |   ctx+=OnToken, ctx+=OnTool (embed in ctx for tools)  |
 |               |               |   LLM decides: call delegate_to_specialist             |
 |               |               |                 |                   |                  |
 |               |               |----Execute(ctx)-->                  |                  |
 |               |               |                 |                   |                  |
 |  <--agent_status("analyzing")-|                 |  (via ctx)        |                  |
 |  <--agent_status("delegating")|                 |  (via ctx)        |                  |
 |  <--agent_status("working")---|                 |  (via ctx)        |                  |
 |               |               |                 |                   |                  |
 |               |               |                 |--processSpecialistTask(ctx)          |
 |               |               |                 |   callbacks = extractDelegationCallbacks(ctx)
 |               |               |                 |                   |                  |
 |               |               |                 |  specialist.ProcessWithSpecialityStream(ctx, task, callbacks)
 |               |               |                 |                   |                  |
 |               |               |                 |                   |--filteredTools = ToolFilter()
 |               |               |                 |                   |  wrappedCbs = wrapWithAttribution(cbs)
 |               |               |                 |                   |                  |
 |  <--agent_status("streaming")-|                 |                   |  (via wrapped cb)|
 |               |               |                 |                   |                  |
 |               |               |                 |                   |--ProcessDirectWithCallbacks(ctx,
 |               |               |                 |                   |    task, sessionKey, wrappedCbs,
 |               |               |                 |                   |    filteredTools)
 |               |               |                 |                   |                  |
 |               |               |                 |                   |  runAgentLoopStreamWithTools(
 |               |               |                 |                   |    opts{ALL CALLBACKS}, filteredTools)
 |               |               |                 |                   |                  |
 |  <--stream(token)-------------|-----------------|-------------------|--onToken(token)  |
 |  <--stream(token)-------------|-----------------|-------------------|--onToken(token)  |
 |               |               |                 |                   |                  |
 |               |               |                 |                   |  LLM calls tool  |
 |  <--tool_call(name,agent="developer")-----------|--onTool(ev)-------|                  |
 |               |               |                 |                   |  tool executes   |
 |  <--tool_call(result,agent="developer")---------|--onTool(ev)-------|                  |
 |               |               |                 |                   |                  |
 |  <--stream(token)-------------|-----------------|-------------------|--onToken(token)  |
 |               |               |                 |                   |                  |
 |               |               |                 |                   |  LLM finishes    |
 |               |               |                 |  <--result string--|                  |
 |               |               |                 |                   |                  |
 |  <--content_segment(result)---|                 |  (via ctx)        |                  |
 |  <--agent_status("complete")--|                 |  (via ctx)        |                  |
 |               |               |                 |                   |                  |
 |               |               |  <--tool result-|                   |                  |
 |               |               |                 |                   |                  |
 |               |               |--LLM synthesizes final response     |                  |
 |  <--stream tokens-------------|                 |                   |                  |
 |  <--stream_end(content,agents)|                 |                   |                  |
```

### 6.3 Parallel Delegation Flow (Phase 3)

```
User          WebSocket      AgentLoop(O)     ParallelDelegTool  Spec-A(dev)    Spec-B(test)
 |               |               |                 |                |               |
 |--message----->|               |                 |                |               |
 |               |               |--runAgentLoopStreamWithTools     |               |
 |               |               |   LLM: call delegate_to_specialists              |
 |               |               |                 |                |               |
 |               |               |----Execute(ctx)-->               |               |
 |               |               |                 |                |               |
 |  <--agent_status("analyzing")-|                 |                |               |
 |  <--agent_status("delegating" specialist="dev,test")             |               |
 |               |               |                 |                |               |
 |               |               |                 |--goroutine A---|               |
 |               |               |                 |--goroutine B---|---------------|
 |               |               |                 |                |               |
 |  <--agent_status("streaming" agent="dev")-------|                |               |
 |  <--stream(token from dev)----|                 |                |               |
 |  <--agent_status("streaming" agent="test")------|----------------|               |
 |  <--stream(token from test)---|                 |                |               |
 |  <--tool_call(agent="dev")----|                 |                |               |
 |  <--stream(token from test)---|                 |                |               |
 |  <--stream(token from dev)----|                 |                |               |
 |               |               |                 |                |               |
 |  <--content_segment(dev)------|                 |  <--result A---|               |
 |  <--agent_status("complete" agent="dev")--------|                |               |
 |  <--content_segment(test)-----|                 |  <--result B---|               |
 |  <--agent_status("complete" agent="test")-------|                |               |
 |               |               |                 |                |               |
 |               |               |  <--aggregated results           |               |
 |               |               |--LLM synthesizes                 |               |
 |  <--stream_end(content,agents)|                 |                |               |
```

---

## 7. Data Flow and Protocol Schemas

### 7.1 DelegationCallbacks Data Flow

```
WebSocket Handler (server.go)
    |
    | Injects into ctx:
    |   - AgentStatusCallback    (ctx key: agentStatusCallbackKey)
    |   - ContentSegmentCallback (ctx key: contentSegmentCallbackKey)
    |   - StreamCallback         (ctx key: streamCallbackKey)         <-- NEW
    |   - ToolCallback           (ctx key: toolCallbackKey)           <-- NEW
    |
    v
AgentLoop.ProcessDirectWithUserAndModelStream()
    |
    | Passes OnToken, OnTool directly in processOptions
    | Also stores in ctx for cross-tool-boundary propagation
    |
    v
runAgentLoopStreamWithTools()
    |
    | Embeds OnToken and OnTool into ctx:
    |   ctx = ContextWithStreamCallback(ctx, opts.OnToken)
    |   ctx = ContextWithToolCallback(ctx, opts.OnTool)
    |
    | LLM calls delegate_to_specialist tool
    |
    v
DelegationTool.Execute(ctx, args)
    |
    | ctx contains all 4 callbacks
    |
    v
processSpecialistTask(ctx, name, task)
    |
    | callbacks = extractDelegationCallbacks(ctx)
    |   -> OnToken:          streamCallbackFromCtx(ctx)
    |   -> OnTool:           toolCallbackFromCtx(ctx)
    |   -> OnAgentStatus:    agentStatusCallbackFromCtx(ctx)
    |   -> OnContentSegment: contentSegmentCallbackFromCtx(ctx)
    |
    v
specialist.ProcessWithSpecialityStream(ctx, task, callbacks)
    |
    | filteredTools = sa.ToolFilter()  // per-call copy
    | wrappedCallbacks = wrapCallbacksWithAttribution(callbacks)
    |   -> OnTool wrapper adds ev.Agent = sa.name
    |   -> OnToken passes through unchanged
    |   -> OnAgentStatus passes through unchanged
    |   -> OnContentSegment passes through unchanged
    |
    v
ProcessDirectWithCallbacks(ctx, task, sessionKey, wrappedCallbacks, filteredTools)
    |
    | Builds processOptions with all wrapped callbacks
    | Routes to runAgentLoopStreamWithTools(ctx, opts, filteredTools, skipHistory=true)
    |
    v
runAgentLoopStreamWithTools()
    |
    | Uses filteredTools for GetDefinitions() and ExecuteWithContext()
    | Calls wrappedOnToken for each streamed token -> WebSocket -> User
    | Calls wrappedOnTool for each tool call -> WebSocket -> User (with agent attribution)
    |
    | Returns result string to ProcessDirectWithCallbacks
    |   -> Returns to ProcessWithSpecialityStream
    |     -> Returns to processSpecialistTask
    |       -> emitContentSegment(ctx, final result)
    |       -> Returns to DelegationTool.Execute
    |         -> Returns tool result to orchestrator's LLM loop
```

### 7.2 WebSocket Event Schema (Complete)

#### stream

```json
{
    "type": "stream",
    "content": "Hello"
}
```
No changes. Agent attribution comes from `agent_status` state.

#### tool_call (extended)

```json
{
    "type": "tool_call",
    "name": "read_file",
    "args": {"path": "/etc/config"},
    "result": "...",
    "status": "started",
    "agent": "developer"
}
```
New optional `agent` field. Only present for specialist tool calls. Omitted for orchestrator/main agent tool calls (backward-compatible).

#### agent_status (new status values)

```json
{
    "type": "agent_status",
    "agent": "developer",
    "status": "streaming",
    "specialist_name": "",
    "reason": "",
    "timestamp": "2026-02-27T10:30:00Z"
}
```
New `status` values: `"streaming"`, `"error"`, `"timeout"`.

#### content_segment (no changes)

```json
{
    "type": "content_segment",
    "agent": "developer",
    "content": "I've implemented the feature by...",
    "segment_id": "seg_developer_1740650400000",
    "timestamp": "2026-02-27T10:30:08Z"
}
```

#### stream_end (extended, Phase 3)

```json
{
    "type": "stream_end",
    "content": "Full response text...",
    "agents": ["orchestrator", "developer"],
    "token_usage": {
        "input_tokens": 4200,
        "output_tokens": 1850,
        "total_tokens": 6050
    }
}
```
New optional `token_usage` field. Phase 3 addition.

---

## 8. Component Interaction Diagram

```
+-------------------+        +-----------------+        +------------------+
|   ChatView.vue    |        |  chatStore.js   |        | WebSocket (ws)   |
|                   |        |                 |        |                  |
| handleMessage()--->        |                 |        |                  |
|   stream_start ---------> startStreaming()   |        |                  |
|   stream ---------->      appendStreamToken()|        |                  |
|   agent_status -------->  setAgentStatus()   |        |                  |
|   tool_call --------->    addToolCall()      |        |                  |
|   content_segment ----->  addContentSegment()|        |                  |
|   stream_end --------->   endStreaming()     |        |                  |
|   ready ----------->      setGlobalLoading() |        |                  |
+-------------------+        +-----------------+        +------------------+
                                    |
                    +---------------+----------------+
                    |               |                |
              +-----v-----+  +-----v-----+  +------v------+
              |MessageBubble| | AgentStatus| | AgentTimeline|
              |    .vue     | | Indicator  | |    .vue      |
              |             | |   .vue     | |   (NEW)      |
              | - segments  | | - status   | | - events     |
              | - toolCalls | | - agent    | | - collapsed  |
              | - agents    | | - reason   | |              |
              +------+------+ +-----+------+ +------+------+
                     |              |                |
              +------v------+ +----v----+    +------v------+
              |SpecialistBadge| |Spinner| |SpecialistBadge|
              |   (dynamic)   | |       | |   (dynamic)   |
              +---------------+ +-------+ +---------------+
                     |
              +------v------+
              | ToolCallItem |
              | (with agent) |
              +--------------+
```

---

## 9. Concurrency Design

### 9.1 Per-Call Tool Registry (Eliminating Shared Mutable State)

**Before (race-prone):**

```
SpecialistAgent
  +-- tools: *ToolRegistry  (SHARED, mutated by swap)
  +-- processMu: sync.Mutex (insufficient protection)
```

**After (race-free):**

```
SpecialistAgent
  +-- tools: *ToolRegistry  (IMMUTABLE after construction -- never swapped)
  +-- processMu: (REMOVED)

ProcessWithSpecialityStream():
  filteredTools := sa.ToolFilter()  // Creates NEW ToolRegistry each call
  // filteredTools is local to this goroutine -- zero contention
  sa.ProcessDirectWithCallbacks(ctx, msg, key, cbs, filteredTools)
```

**Memory analysis:**
- `ToolRegistry` = `map[string]Tool` + `sync.RWMutex`
- Each specialist typically has 5-15 tools.
- `map[string]Tool` copy: ~5-15 map entries of pointers (< 1KB per call).
- Allocation frequency: once per delegation (not per LLM iteration).
- GC pressure: negligible for the delegation rate (~1-10/min in practice).

### 9.2 WebSocket Writer Concurrency

The existing `wsMu sync.Mutex` in `handleChatWS` protects all WebSocket writes. When specialist callbacks write to the WebSocket, they acquire the same mutex. This is correct and sufficient:

```
Orchestrator LLM streaming tokens  -> wsMu.Lock() -> conn.WriteJSON -> wsMu.Unlock()
Specialist LLM streaming tokens    -> wsMu.Lock() -> conn.WriteJSON -> wsMu.Unlock()
Specialist tool call events        -> wsMu.Lock() -> conn.WriteJSON -> wsMu.Unlock()
Agent status events (from ctx)     -> wsMu.Lock() -> conn.WriteJSON -> wsMu.Unlock()
Content segment events (from ctx)  -> wsMu.Lock() -> conn.WriteJSON -> wsMu.Unlock()
```

Under parallel delegation (Phase 3), the mutex ensures serialized writes. No additional synchronization is needed.

### 9.3 Context Cancellation

The existing `context.WithTimeout` in `processSpecialistTask` creates a child context per specialist. When the user disconnects (WebSocket close), the parent `r.Context()` is cancelled, which cascades to all specialist contexts. The specialist's LLM streaming loop checks `ctx.Done()` and aborts.

For parallel delegation, each specialist goroutine gets its own child context. If one specialist fails/times out, only its context is cancelled. The `errgroup.Group` collects results from all goroutines.

---

## 10. Token Budget Tracking

### 10.1 Measurement Points

```
runLLMIterationStreamImpl()
  |
  | For each iteration:
  |   ChatStream() returns StreamChunk with Usage on final chunk
  |   if chunk.Usage != nil:
  |     accumulator.AddInput(chunk.Usage.PromptTokens)
  |     accumulator.AddOutput(chunk.Usage.CompletionTokens)
  |   else:
  |     accumulator.AddEstimated(len(content) / 4)
  |
  v
ProcessDirectWithCallbacks() returns TokenUsage
  |
  v
processSpecialistTask() includes in DelegationResult
  |
  v
DelegationTool.Execute() returns formatted result with token info
  |
  v
stream_end includes token_usage (aggregated from all specialists)
```

### 10.2 TokenUsage Accumulator

```go
type TokenUsageAccumulator struct {
    inputTokens  atomic.Int64
    outputTokens atomic.Int64
}

func (a *TokenUsageAccumulator) Add(input, output int) {
    a.inputTokens.Add(int64(input))
    a.outputTokens.Add(int64(output))
}

func (a *TokenUsageAccumulator) Total() TokenUsage {
    in := a.inputTokens.Load()
    out := a.outputTokens.Load()
    return TokenUsage{
        InputTokens:  in,
        OutputTokens: out,
        TotalTokens:  in + out,
    }
}
```

### 10.3 Reporting

Token usage is informational in Phase 3:
1. Logged via `logger.InfoCF("agent", "Specialist token usage", fields)`.
2. Included in `DelegationResult.TokensUsed` (returned to orchestrator LLM for awareness).
3. Included in `stream_end.token_usage` (displayed in frontend, if present).
4. Recorded in `observability.Global().RecordLLMCall()` (already implemented per-iteration).

---

## 11. Testing Strategy

### 11.1 Unit Tests

| Test | File | Description |
|------|------|-------------|
| `TestProcessDirectWithCallbacks` | `loop_test.go` | Verify callbacks are invoked during streaming |
| `TestProcessDirectWithCallbacks_NoCallbacks` | `loop_test.go` | Verify nil callbacks do not panic |
| `TestProcessDirectWithCallbacks_ToolsOverride` | `loop_test.go` | Verify tool registry override is used |
| `TestProcessWithSpecialityStream` | `specialist_test.go` | Verify callbacks are wrapped with attribution |
| `TestProcessWithSpecialityStream_Concurrent` | `specialist_test.go` | Verify concurrent delegations are safe (no race) |
| `TestToolFilter_PerCall` | `specialist_test.go` | Verify each call gets independent registry |
| `TestExtractDelegationCallbacks` | `orchestrator_test.go` | Verify all callback types extracted from ctx |
| `TestProcessSpecialistTask_WithCallbacks` | `orchestrator_test.go` | Verify streaming path is used when callbacks present |
| `TestProcessSpecialistTask_WithoutCallbacks` | `orchestrator_test.go` | Verify fallback to non-streaming path |
| `TestToolEventAgent` | `loop_test.go` | Verify Agent field serializes correctly |
| `TestWrapCallbacksWithAttribution` | `specialist_test.go` | Verify tool events get agent name |

### 11.2 Race Detection

```bash
go test -race ./pkg/agent/... -run TestProcessWithSpecialityStream_Concurrent
```

The concurrent test should launch 10+ goroutines calling `ProcessWithSpecialityStream` on the same specialist simultaneously. With the per-call tool registry fix, this must pass with zero races.

### 11.3 Integration Tests

| Test | Description |
|------|-------------|
| WebSocket streaming with orchestrator | Verify full pipeline: message -> orchestrator -> specialist -> streamed tokens -> stream_end with agents |
| Agent status events during delegation | Verify frontend receives analyzing -> delegating -> working -> streaming -> complete |
| Tool call events with agent attribution | Verify `tool_call` events include `agent` field for specialist tools |
| Content segment for specialist result | Verify `content_segment` event fires with specialist result |
| Backward compatibility | Verify non-orchestrated agent loop is completely unaffected |

### 11.4 Frontend Tests

| Test | Description |
|------|-------------|
| SpecialistBadge dynamic colors | Verify custom specialist names get unique, consistent colors |
| AgentTimeline rendering | Verify timeline renders from agentHistory events |
| Tool call with agent attribution | Verify ToolCallItem shows agent badge when showAgent is true |
| Segment rendering | Verify MessageBubble renders segments when available |

---

## 12. Migration and Backward Compatibility

### 12.1 No Breaking Changes

All changes are additive:

1. **New methods** (`ProcessDirectWithCallbacks`, `ProcessWithSpecialityStream`) do not modify existing methods.
2. **New `Agent` field** on `ToolEvent` is `omitempty` -- existing JSON output is identical.
3. **New context keys** (`streamCallbackKey`, `toolCallbackKey`) -- contexts without these keys return `nil` callbacks, which are safely ignored.
4. **New WebSocket fields** (`agent` on `tool_call`, `token_usage` on `stream_end`) -- existing frontends ignore unknown fields.
5. **New `agent_status` values** (`streaming`, `error`, `timeout`) -- existing frontends show "Processing..." for unknown values (graceful degradation via `statusMap` default).

### 12.2 Config Changes

**Phase 3 only:** New optional field in `OrchestratorConfig`:

```go
type OrchestratorConfig struct {
    // ... existing fields ...
    MaxParallelDelegations int `json:"max_parallel_delegations"` // Default: 3
}
```

This field is optional with a sensible default. Existing configs continue to work without modification.

### 12.3 Rollback Plan

- **Phase 1:** Remove the new context keys from `handleChatWS`. `processSpecialistTask` falls back to `ProcessWithSpeciality` (non-streaming). Zero risk.
- **Phase 2:** Remove `agent` field from `tool_call` JSON output. Remove `SpecialistBadge` dynamic fallback (revert to hardcoded). Frontend continues to work.
- **Phase 3:** Remove `delegate_to_specialists` tool. Orchestrator uses only `delegate_to_specialist` (sequential). No configuration impact.

---

## Appendix A: File Modification Summary

| File | Phase | Changes |
|------|-------|---------|
| `pkg/agent/loop.go` | 1 | Add `DelegationCallbacks`, `TokenUsage`, `ToolEvent.Agent` field, `ProcessDirectWithCallbacks`, `runAgentLoopStreamWithTools`, `runAgentLoopWithTools`, refactor `runLLMIterationStream` -> `runLLMIterationStreamImpl` |
| `pkg/agent/specialist.go` | 1 | Add `ProcessWithSpecialityStream`, `wrapCallbacksWithAttribution`; deprecation comment on `ProcessWithSpeciality` |
| `pkg/agent/orchestrator.go` | 1 | Add context keys (`streamCallbackKey`, `toolCallbackKey`), `extractDelegationCallbacks`, modify `processSpecialistTask` to use streaming path, add `delegationResultInternal` |
| `pkg/web/server.go` | 1 | Add `ContextWithStreamCallback` and `ContextWithToolCallback` injection in `handleChatWS`; add `agent` field to `tool_call` JSON output |
| `pkg/web/frontend/src/components/Chat/SpecialistBadge.vue` | 2 | Dynamic hash-based colors/icons for custom specialist names |
| `pkg/web/frontend/src/components/Chat/AgentTimeline.vue` | 2 | NEW: collapsible orchestration timeline component |
| `pkg/web/frontend/src/components/ToolCallItem.vue` | 2 | Add optional `showAgent` prop with `SpecialistBadge` |
| `pkg/web/frontend/src/stores/chatStore.js` | 2 | Add `activeStreamingAgent`, `streamingSegmentBuffer`; modify `setAgentStatus`, `addToolCall` for timeline events |
| `pkg/web/frontend/src/components/MessageBubble.vue` | 2 | Pass `showAgent` to `ToolCallItem` when multi-agent |
| `pkg/config/config.go` | 3 | Add `MaxParallelDelegations` to `OrchestratorConfig` |
| `pkg/agent/orchestrator.go` | 3 | Add `ParallelDelegationTool`, `delegate_to_specialists` tool registration |
| `pkg/agent/loop.go` | 3 | Add `TokenUsageAccumulator`; propagate usage from `StreamChunk.Usage` |

---

## Appendix B: Glossary

| Term | Definition |
|------|------------|
| **Callback** | A function passed as a parameter and invoked during processing (e.g., `StreamCallback`, `ToolCallback`) |
| **processOptions** | Internal struct in `AgentLoop` that configures how a message is processed, including callbacks |
| **DelegationCallbacks** | New struct bundling all four callback types for specialist delegation |
| **Tool Registry** | `tools.ToolRegistry` -- a thread-safe map of tool name -> Tool implementation |
| **Per-call copy** | Creating a new `ToolRegistry` instance with filtered tools for each delegation call |
| **Agent attribution** | Annotating events (tokens, tool calls, segments) with the name of the agent that produced them |
| **Content segment** | A chunk of content attributed to a specific agent, emitted as a `content_segment` WebSocket event |
| **Streaming segment buffer** | Frontend-side buffer that accumulates tokens during specialist streaming for later segment attribution |

---

## Approval

| Role | Name | Date | Decision |
|------|------|------|----------|
| Author | SDD Sub-Agent | 2026-02-27 | Proposed |
| Reviewer | | | |
| Approver | | | |
