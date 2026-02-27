# Exploration: Multi-Agent Orchestration Visibility

**Change:** multi-agent-orchestration-visibility
**Date:** 2026-02-27
**Status:** exploration_complete

---

## Executive Summary

The multi-agent orchestration system is **architecturally well-built** but has several critical gaps that prevent it from delivering on its promise. The backend infrastructure (orchestrator, specialists, delegation tool, status callbacks, agent tracking) is already substantial and mostly functional. However, the system suffers from:

1. **A broken delegation pipeline**: The orchestrator uses its own separate `AgentLoop` with `ProcessDirect`, which does NOT propagate streaming callbacks or the `ctx`-embedded status/segment callbacks to the specialist's inner LLM loop. The callbacks are set on the WebSocket handler's `ctx`, and the specialist's `ProcessWithSpeciality` calls `ProcessDirect` which creates a brand new `processOptions` without any callbacks.

2. **No streaming from specialists**: When the orchestrator delegates to a specialist, the specialist runs its entire LLM loop non-streaming (via `ProcessDirect`), then returns the full result. The user sees nothing during specialist execution.

3. **Frontend has the plumbing but it is mostly unused**: `AgentStatusIndicator`, `SpecialistBadge`, `ContentSegment` rendering, and `agentHistory` are all implemented in the frontend, but they rarely get triggered because the backend callback chain is broken.

4. **Skills per specialist work but are rarely configured**: The `SpecialistConfig.Skills` field exists and `SetSkillFilter` is properly called, but specialists typically get `lightweightMode=true` which skips bootstrap files and memory, and most specialists are configured with empty skills lists.

5. **Token efficiency has good foundations**: `lightweightMode` for specialists/orchestrator strips bootstrap files and memory. Skill filtering works. But specialists still get the full identity system prompt.

---

## Detailed Findings

### 1. Delegation Flow (Backend)

**Path:** User message → WebSocket handler → `activeAgentLoop.ProcessDirectWithUserAndModelStream()` → `processMessageWithModelStream()` → `runAgentLoopStream()` → LLM calls orchestrator → orchestrator calls `delegate_to_specialist` tool → `DelegationTool.Execute()` → `processSpecialistTask()` → specialist's `ProcessWithSpeciality()` → `ProcessDirect()` → `runAgentLoop()` (non-streaming)

**Critical Issue:** The callbacks break at the orchestrator-to-specialist boundary.

- The WebSocket handler injects `AgentStatusCallback` and `ContentSegmentCallback` into `ctx` (lines 1224-1248 in `server.go`).
- The `DelegationTool.Execute()` receives this `ctx` and correctly calls `emitAgentStatus()` / `emitContentSegment()` to emit events.
- BUT the specialist's `ProcessWithSpeciality()` calls `ProcessDirect()`, which calls `processMessage()` → `processMessageWithModel()` → `runAgentLoop()`.
- `runAgentLoop()` creates a fresh `processOptions{}` with NO `OnToken`, `OnTool`, `OnAgentStatus`, or `OnContentSegment` callbacks.
- Therefore, the specialist's tool calls are invisible, the specialist's response is not streamed, and the specialist's internal processing emits no status events.

**What does work:**
- `emitAgentStatus()` works for the orchestrator-level events ("analyzing", "delegating", "working", "complete") because those are called directly in `DelegationTool.Execute()` using the original `ctx`.
- `emitContentSegment()` works when the specialist returns its result back to the orchestrator, because `processSpecialistTask()` calls it directly (line 309-314).
- `agentTrackerFromCtx(ctx)` correctly propagates involved agent names via the `ContextWithAgentTracker` mechanism.
- The `stream_end` message correctly includes `agents` array.

**What does NOT work:**
- No streaming tokens from specialist's LLM response.
- No tool call events from specialist's tool executions.
- No granular status updates during specialist's internal processing.
- The specialist's LLM loop runs completely silently from the user's perspective.

### 2. Agent Tracking and Attribution

**File:** `pkg/agent/loop.go` lines 54-92

The `involvedAgents` tracking is well-implemented:
- `ClearInvolvedAgents()` is called at the start of each `runAgentLoop` / `runAgentLoopStream`.
- `AddInvolvedAgent(name)` is called to register the primary agent (model name or "main").
- When delegation occurs, `processSpecialistTask()` adds "orchestrator" and the specialist name.
- `ContextWithAgentTracker` allows the web chat's `activeAgentLoop` to be the tracker (so specialist names are registered on the correct loop instance).

**What works:** Agent names ARE tracked and sent in `stream_end.agents` and persisted in `chats.metadata` as `{"agents": ["orchestrator", "developer"]}`.

**What does NOT work:**
- The agents are only revealed at the END of the response (in `stream_end`).
- There is no per-token attribution during streaming.
- When loading session history, agents are parsed from `metadata` and displayed as badges, but there is no way to attribute parts of the response to different agents in historical messages.

### 3. Frontend Rendering

**Components already built:**
- `AgentStatusIndicator.vue` - Shows a status bar with spinner during orchestrator operations (analyzing/delegating/working). Works correctly when events arrive.
- `SpecialistBadge.vue` - Shows colored badge with icon for each specialist type (developer, documentation, testing, devops, analyst, researcher). Has hardcoded icons/colors for 6 specialist types.
- `SpecialistsPanel.vue` - Collapsible panel showing available specialists. Fetches from API.
- `MessageBubble.vue` - Has full segment rendering: if `msg.segments` has entries, renders each segment with a `SpecialistBadge` header and attributed content. If no segments, renders unified content. Also shows agent badges at the bottom of messages.

**chatStore.js state:**
- `orchestratorStatus`, `currentAgent`, `activeSpecialist`, `delegationReason`, `agentHistory` - All tracked.
- `setAgentStatus()`, `clearAgentStatus()`, `addContentSegment()` - All implemented.
- Streaming messages have `agents: []` and `segments: []` initialized.

**ChatView.vue handlers:**
- `agent_status` events → `chatStore.setAgentStatus()`
- `content_segment` events → `chatStore.addContentSegment()`
- `stream_end` → `chatStore.endStreamingMessage(content, agents)`
- Session history load parses `metadata.agents` from stored messages.

**Gap:** The frontend is fully ready but rarely receives the events because:
1. `agent_status` events only fire for orchestrator-level events, not specialist internals.
2. `content_segment` events only fire once when the specialist returns its final result (not during streaming).
3. No specialist tool call events reach the frontend.

### 4. Specialist Configuration

**Config structure** (`config.SpecialistConfig`):
```go
type SpecialistConfig struct {
    Name              string   `json:"name"`
    Description       string   `json:"description"`
    Prompt            string   `json:"prompt"`
    Provider          string   `json:"provider"`
    Model             string   `json:"model"`
    MaxTokens         int      `json:"max_tokens"`
    Temperature       float64  `json:"temperature"`
    MaxToolIterations int      `json:"max_tool_iterations"`
    Tools             []string `json:"tools"`
    Keywords          []string `json:"keywords"`
    Skills            []string `json:"skills,omitempty"`
}
```

**What works:**
- Each specialist can have a different provider/model (token savings by using cheaper models for simpler tasks).
- Tool filtering works: `ToolFilter()` creates a filtered registry with only allowed tools.
- Skill filtering works: `SetSkillFilter(cfg.Skills)` limits which skills are in the specialist's system prompt.
- `lightweightMode=true` for both orchestrator and specialists skips bootstrap files (AGENTS.md, SOUL.md, USER.md, IDENTITY.md) and memory context.
- Specialist prompt is injected via `SetAgentSystemPrompt(cfg.Prompt)`.

**Gaps:**
- Skills field is `omitempty` and defaults to `nil` which means ALL skills are loaded for specialists that don't specify it. The `SetSkillFilter` is only called if `cfg.Skills != nil`.
- Specialists still get the full identity/runtime header from `getIdentity()` even in lightweight mode.
- No way to configure specialist-specific memory or context.

### 5. spawn tool vs orchestrator delegation

Two separate mechanisms exist for multi-agent work:

| Aspect | Orchestrator Delegation | Spawn Tool |
|--------|------------------------|------------|
| Mechanism | `delegate_to_specialist` tool called by orchestrator LLM | `spawn` tool called by any agent |
| Execution | Synchronous (blocks orchestrator) | Asynchronous (runs in background goroutine) |
| Result delivery | Returns result directly to orchestrator | Sends result back via message bus as system message |
| Context | Has full specialist config, tool filtering, skill filtering | Minimal: just system prompt "You are a subagent" |
| Streaming | None (ProcessDirect) | None |
| Tracking | Tracked via involvedAgents | Not tracked in involvedAgents |
| Status events | Emits agent_status events | No status events |

The spawn tool is much simpler and less powerful. It uses the base provider with no specialist configuration.

### 6. Token Saving Strategies

**Currently implemented:**
- `lightweightMode` for orchestrator/specialists: Saves ~1000-5000 tokens per specialist call by skipping bootstrap files and memory.
- Orchestrator skill filter set to `[]string{}` (empty): No skills loaded for orchestrator.
- Specialist skill filter: Configurable per-specialist.
- Orchestrator only has `delegate_to_specialist` tool: Minimal tool definitions sent to LLM.
- Tool filtering for specialists: Only allowed tools are sent as definitions.

**Missing:**
- No context window awareness: Specialists get full identity prompt even when not needed.
- No response caching: Same specialist may be called repeatedly with similar tasks.
- No parallel delegation: Orchestrator delegates sequentially, not in parallel.
- No delegation cost estimation: No mechanism to estimate token cost before delegating.
- Orchestrator still receives full previous conversation history (via ProcessDirect).

### 7. WebSocket Protocol

Messages the frontend can receive:

| Type | Fields | When |
|------|--------|------|
| `stream_start` | - | Before streaming begins |
| `stream` | `content` (token) | Each streamed token |
| `stream_end` | `content`, `agents[]`, `error` | Streaming complete |
| `tool_call` | `name`, `args`, `result`, `status` | Tool execution start/finish |
| `agent_status` | `agent`, `status`, `specialist_name`, `reason`, `timestamp` | Orchestrator status change |
| `content_segment` | `agent`, `content`, `segment_id`, `timestamp` | Content from a specialist |
| `message` | `role`, `content`, `agents[]` | Non-streaming response |
| `ready` | - | Agent is ready for next message |

---

## Key Questions Answered

### Q1: Does the orchestrator actually call specialists?
**Yes**, but only when the orchestrator is enabled, specialists are configured, and the LLM decides to use the `delegate_to_specialist` tool. The flow works end-to-end but runs silently from the user's perspective.

### Q2: How are specialist responses merged/returned?
The specialist's response string is returned as the tool result of `delegate_to_specialist`. The orchestrator LLM then processes this result and formulates its own final response. There is no automatic merging; the orchestrator LLM decides how to present the specialist's work.

### Q3: Is there metadata about which agent produced which part?
**Partially.** `involvedAgents` tracks which agents participated, and `ContentSegment` events can attribute content to specific agents. But the orchestrator's final response is not attributed -- only the raw specialist result is emitted as a `content_segment`. The orchestrator's synthesis/summary is not attributed.

### Q4: How are skills assigned to specialists vs orchestrator?
- Orchestrator: Skills are explicitly set to empty (`SetSkillFilter([]string{})`).
- Specialists: If `cfg.Skills` is non-nil, only those skills are loaded. If nil (default), ALL skills are loaded.
- Default agent (non-orchestrated): ALL skills are loaded.

### Q5: What does the WebSocket stream include?
Text tokens, tool call events, agent status events, content segments, and the final `stream_end` with agents list. The protocol is comprehensive but the events are rarely emitted due to the broken callback chain.

### Q6: How does spawn tool work vs orchestrator delegation?
Spawn is asynchronous and simple (no specialist config). Orchestrator delegation is synchronous and uses full specialist configuration. They are independent systems. See comparison table above.

### Q7: What token-saving strategies exist?
`lightweightMode`, skill filtering, and tool filtering. See section 6 for full analysis.

---

## Risks & Concerns

1. **Concurrency bug in specialist tool swap**: `ProcessWithSpeciality` swaps `sa.tools` with a filtered copy under a mutex, but if two concurrent delegations hit the same specialist, the restore in `defer` could race with a new swap. The mutex only protects the swap operation, not the entire execution.

2. **Orchestrator uses ProcessDirect which bypasses streaming**: This is by design (tool call results are not streamed) but means the user sees no progress during potentially long specialist executions.

3. **SpecialistBadge hardcoded to 6 types**: Users can create custom-named specialists but the badge component only recognizes developer, documentation, testing, devops, analyst, researcher. Custom names get a generic gray badge.

4. **No error recovery in delegation**: If a specialist fails, the orchestrator LLM sees the error as a tool result. There is `delegationRetries` config but it is only applied as `maxToolIterations` on the orchestrator's agent loop, not as retry logic for individual specialist calls.

5. **Session isolation**: The orchestrator creates specialist sessions with keys like `specialist_developer`, which are isolated from the user's web chat session. This means specialist context does not persist across user conversations.

---

## Recommended Next Steps

1. **Fix the callback propagation chain**: Pass `OnToken`, `OnTool`, `OnAgentStatus`, `OnContentSegment` from the WebSocket handler through the orchestrator to the specialist's `runAgentLoopStream`. This is the highest-impact fix.

2. **Enable specialist streaming**: Modify `ProcessWithSpeciality` to optionally use `ProcessDirectWithModelStream` when callbacks are available in the context.

3. **Enrich agent status events**: Add more granular events (specialist tool calls, specialist thinking, specialist streaming tokens) so the frontend can show real-time progress.

4. **Dynamic SpecialistBadge**: Generate badge colors/icons from specialist name hash instead of hardcoded map, so custom specialists get unique visual identity.

5. **Parallel delegation support**: Allow the orchestrator to delegate to multiple specialists simultaneously for tasks that can be parallelized.

6. **Token budget system**: Track token usage per specialist call and implement budgets/limits per-specialist.

7. **Frontend: timeline view**: Use `agentHistory` to show a timeline of agent events alongside the response, giving users visibility into the orchestration process.

---

## File Reference

| File | Lines | Key Content |
|------|-------|-------------|
| `pkg/agent/orchestrator.go` | 1-445 | OrchestratorAgent, DelegationTool, AgentStatusEvent callbacks, BuildOrchestratorContext |
| `pkg/agent/specialist.go` | 1-398 | SpecialistAgent, SpecialistRegistry, ToolFilter, ProcessWithSpeciality, LoadSpecialistsFromConfig |
| `pkg/agent/manager.go` | 1-189 | AgentManager, InitializeOrchestrator, GetActiveAgent |
| `pkg/agent/loop.go` | 1-1400+ | AgentLoop, involvedAgents tracking, processOptions, runAgentLoop/runAgentLoopStream, callback types |
| `pkg/agent/context.go` | 1-412 | ContextBuilder, SetSkillFilter, SetLightweightMode, SetAgentSystemPrompt, BuildSystemPrompt |
| `pkg/config/config.go` | 96-136 | AgentsConfig, OrchestratorConfig, SpecialistConfig |
| `pkg/tools/spawn.go` | 1-79 | SpawnTool (async subagent) |
| `pkg/tools/subagent.go` | 1-132 | SubagentManager (background task runner) |
| `pkg/web/server.go` | 1088-1339 | handleChatWS (WebSocket handler, callback injection, agent tracker) |
| `pkg/web/frontend/src/stores/chatStore.js` | 1-400 | Chat store (orchestratorStatus, agentHistory, segments, tool calls) |
| `pkg/web/frontend/src/stores/agentsStore.js` | 1-232 | Agents store (specialists, orchestrator config, badge helpers) |
| `pkg/web/frontend/src/views/ChatView.vue` | 720-755 | WebSocket message handlers (agent_status, content_segment) |
| `pkg/web/frontend/src/components/MessageBubble.vue` | 1-100 | Segmented content rendering, agent badges |
| `pkg/web/frontend/src/components/Chat/AgentStatusIndicator.vue` | 1-127 | Real-time agent status display |
| `pkg/web/frontend/src/components/Chat/SpecialistBadge.vue` | 1-51 | Specialist visual badge (hardcoded 6 types) |
| `pkg/web/frontend/src/components/Chat/SpecialistsPanel.vue` | 1-104 | Available specialists panel |
