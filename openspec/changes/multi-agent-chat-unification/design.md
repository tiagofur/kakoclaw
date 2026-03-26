# Design: Multi-Agent Chat Unification

**Change:** `multi-agent-chat-unification`
**Status:** Ready for Tasks
**Created:** 2026-03-25
**Based on:** [proposal.md](proposal.md), [spec.md](spec.md)

---

## Technical Approach

Two-phase delivery. Phase 1 makes targeted, low-risk changes: one SQL WHERE clause on the read path, two new Vue components, re-routing of a single `clearAgentStatus()` call, and two new WS event types emitted from the existing `AgentStatusCallback` pipeline. Phase 2 threads the user `sessionKey` through Go's `context.Context` using a typed key — the same pattern already used for `agentTrackerKey`, `contentSegmentCallbackKey`, etc.

---

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|-------------|-----------|
| Where to filter specialist sessions | `ListSessionsForUser()` SQL `WHERE session_id NOT LIKE 'specialist_%'` (+ legacy fallback branch) | Frontend computed, separate API endpoint | Single enforcement point; filters both primary and legacy query paths; zero frontend cost |
| Synthesis event delivery | New `AgentStatusEvent` statuses (`synthesis_start`, `synthesis_end`) reusing the existing `ContextWithAgentStatusCallback` pipe | New WS message type; new callback type | Reuses all existing serialization code in `server.go`; frontend already dispatches on `message.type === 'agent_status'`; zero new backend contracts |
| `SpecialistSegment` component | New `SpecialistSegment.vue` alongside `ToolCallItem.vue` pattern | Inline in `MessageBubble.vue`, extend `AgentActivityItem` | `ToolCallItem` already proven; extraction keeps `MessageBubble` readable; segment-specific props (status badge, preview) don't belong in tool-call item |
| `clearAgentStatus()` lifecycle | Move call from `stream_end` → `beforeRouteLeave` + session-change guard in `sendMessage()` | Store watcher on route, `onUnmounted` hook | Matches Vue Router lifecycle exactly; `beforeRouteLeave` fires on every navigation; avoids accidental unmount-based clearing |
| Phase 2 sessionKey propagation | Typed `context.Context` key (`sessionKeyCtxKey{}`) | Add param to `processSpecialistTask` signature, struct field | Consistent with existing `agentTrackerKey`, `delegationChainKey` patterns in `orchestrator.go`; non-breaking signature change |
| `aggregateReports()` in standard flow | Call after `DelegationTool.Execute` accumulates responses within the `runAgentLoop` cycle | Separate method, new orchestrator state machine | `aggregateReports()` already exists; only call site addition needed; no new structs |

---

## Data Flow

### Phase 1 — Session Filtering

```
GET /api/v1/chat/sessions
    │
    ▼
ListSessionsForUser()  ←─ adds: WHERE session_id NOT LIKE 'specialist_%'
    │                      (applied to BOTH isUserDB and legacy branches)
    ▼
[]SessionSummary (no specialist_* entries)
    │
    ▼
fetchSessions() in ChatView.vue
    │
    ▼
sessions.value  (existing filter keeps task: exclusion;
                 specialist_* never arrive → no frontend filter needed)
```

### Phase 1 — Synthesis Indicator

```
orchestrator.go: DelegationTool.Execute() receives all specialist results
    │
    ▼  emitAgentStatus(ctx, {Status: "synthesis_start"})
    │
    ▼
ContextWithAgentStatusCallback (embedded in ctx by server.go)
    │
    ▼
conn.WriteJSON({type: "agent_status", status: "synthesis_start", agent: "orchestrator"})
    │
    ▼
ChatView.vue handleMessage("agent_status")
    │  status === "synthesis_start" → chatStore.synthesizing = true
    ▼
<SynthesisIndicator v-if="chatStore.synthesizing" />  →  "Orchestrator synthesizing…"

    orchestrator finalizes LLM call
    │
    ▼  emitAgentStatus(ctx, {Status: "synthesis_end"})  →  synthesizing = false
    │
    ▼
SynthesisIndicator unmounts
```

### Phase 1 — clearAgentStatus Lifecycle

```
Current (broken):
  stream_end → clearAgentStatus()  // wipes history while user is still reading

New:
  stream_start   → clearAgentStatus()  (only if session changed from lastActiveSessionId)
  beforeRouteLeave → clearAgentStatus()  (unconditional)
  stream_end     → [NO clearAgentStatus call]
```

### Phase 2 — SessionKey Propagation

```
server.go injects sessionKey into context:
  ctx = context.WithValue(ctx, sessionKeyCtxKey{}, "web:chat:abc123")
    │
    ▼
orchestrator.go processSpecialistTask():
  sessionKey := sessionKeyFromCtx(ctx)  // "web:chat:abc123" or fallback ""
    │
    ▼
specialist.ProcessWithSpeciality(ctx, task, sessionKey)
    │  sessionKey != "" → use it as sessionID for ProcessDirect()
    │  sessionKey == "" → fallback: fmt.Sprintf("specialist_%s", sa.name)
    ▼
specialist stores messages under user's session
```

---

## File Changes

### Phase 1

| File | Action | Description |
|------|--------|-------------|
| `pkg/storage/chat.go` | Modify | Add `AND session_id NOT LIKE 'specialist_%'` to `WHERE` clause in both branches of `ListSessionsForUser()`; same filter in legacy fallback query |
| `pkg/agent/loop.go` | Modify | Add `"synthesis_start"` and `"synthesis_end"` to `AgentStatusEvent.Status` doc comment (no struct change needed — `Status` is already `string`) |
| `pkg/agent/orchestrator.go` | Modify | In `DelegationTool.Execute()`, emit `synthesis_start` before the orchestrator's final LLM call and `synthesis_end` after; emit only when `len(specialistResponses) > 0` |
| `pkg/web/frontend/src/stores/chatStore.js` | Modify | Add `synthesizing: ref(false)`; handle `synthesis_start`/`synthesis_end` in a new `setSynthesizing(bool)` action; add `lastActiveSessionId: ref(null)` for session-aware clear guard |
| `pkg/web/frontend/src/views/ChatView.vue` | Modify | (1) Remove `clearAgentStatus()` from `stream_end` handler; (2) Move call to `beforeRouteLeave`; (3) Add session-change guard at start of `sendMessage()`; (4) Handle `synthesis_start`/`synthesis_end` in `handleMessage()` via `chatStore.setSynthesizing()`; (5) Add `<SynthesisIndicator>` above `<TeamActivityPanel>` |
| `pkg/web/frontend/src/components/MessageBubble.vue` | Modify | Replace current `showSegments` toggle block with `<SpecialistSegment>` per segment |
| `pkg/web/frontend/src/components/Chat/SpecialistSegment.vue` | Create | Accordion component for a single `msg.segments` entry; collapsed by default; CSS transition |
| `pkg/web/frontend/src/components/Chat/SynthesisIndicator.vue` | Create | Inline indicator: icon + "Orchestrator synthesizing…" + pulsing dots; renders only when `chatStore.synthesizing` |

### Phase 2

| File | Action | Description |
|------|--------|-------------|
| `pkg/agent/orchestrator.go` | Modify | Add `sessionKeyCtxKey{}` typed key; add `sessionKeyFromCtx()` helper; modify `processSpecialistTask()` to extract key and pass to specialist |
| `pkg/agent/specialist.go` | Modify | Add optional `sessionKey string` to `ProcessWithSpeciality()` or introduce `ProcessWithSpecialityAndSession()` overload; use sessionKey when calling `ProcessDirect()` |
| `pkg/web/server.go` | Modify | Inject user's `sessionKey` into context before calling orchestrator (same location as `ContextWithAgentStatusCallback`) |
| `pkg/agent/orchestrator.go` | Modify | Add `aggregateReports()` call in standard `runAgentLoop` delegation path (after all `DelegationTool.Execute()` complete, `len(specialistResponses) > 0`) |

---

## Interfaces / Contracts

### New WS events (Backend → Frontend)

These reuse the existing `agent_status` WS message type — no new field contracts:

```json
// synthesis_start
{ "type": "agent_status", "agent": "orchestrator", "status": "synthesis_start", "timestamp": "2026-03-25T..." }

// synthesis_end
{ "type": "agent_status", "agent": "orchestrator", "status": "synthesis_end", "timestamp": "2026-03-25T..." }
```

The frontend's existing `handleMessage('agent_status')` dispatch already routes by `message.status` — adding two new status strings requires only two new `if` branches.

### chatStore new state

```js
const synthesizing = ref(false)
const lastActiveSessionId = ref(null)

function setSynthesizing(val) { synthesizing.value = !!val }
```

### SpecialistSegment.vue props

```ts
props: {
  segment: {
    type: Object,  // { specialist, status, content, confidence? }
    required: true
  },
  isStreaming: { type: Boolean, default: false }
}
```

### Phase 2 — Go context key

```go
type sessionKeyCtxKey struct{}

func contextWithSessionKey(ctx context.Context, key string) context.Context {
  return context.WithValue(ctx, sessionKeyCtxKey{}, key)
}

func sessionKeyFromCtx(ctx context.Context) string {
  if v, ok := ctx.Value(sessionKeyCtxKey{}).(string); ok { return v }
  return ""
}
```

---

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit — Go | `ListSessionsForUser` excludes `specialist_*`; legacy branch also excluded | `storage_test.go`: insert specialist session, assert absent from result |
| Unit — Go | `synthesis_start/end` events emitted by `DelegationTool.Execute` | `orchestrator_test.go`: mock `AgentStatusCallback`, assert statuses emitted in order |
| Unit — Go (Phase 2) | `sessionKeyFromCtx` fallback when key absent | `orchestrator_test.go`: empty context → returns `""` |
| Unit — JS | `chatStore.setSynthesizing(true/false)` | Vitest: dispatch, inspect `synthesizing` ref |
| Unit — JS | `clearAgentStatus` NOT called on `stream_end` | Vitest: spy `clearAgentStatus`; fire `stream_end`; assert not called |
| Unit — JS | `clearAgentStatus` called on session change in `sendMessage` | Vitest: set different `lastActiveSessionId`; call `sendMessage`; assert called |
| Component | `SpecialistSegment` collapsed by default; click toggles | Vue Test Utils: mount with segment, assert no expanded class; simulate click |
| Component | `SynthesisIndicator` mounts/unmounts on `synthesizing` flag | Vue Test Utils: set store flag, assert element present/absent |
| E2E | Sidebar shows no `specialist_*` entries | Playwright: trigger multi-agent chat, assert sidebar list |
| E2E | Accordion expand/collapse | Playwright: click segment header, assert content visible/hidden |
| E2E | Synthesis indicator appears then disappears | Playwright: mock WS events, assert indicator lifecycle |

---

## Migration / Rollout

No migration required. All changes are additive:

- SQL filter is a read-only `WHERE` addition — revert = remove the clause.
- New WS statuses are ignored by old frontend versions (unknown `status` → no-op).
- Phase 2 guarded by feature flag `FEATURE_SESSION_PROPAGATION` (env var / config key); when disabled, specialist falls back to `"specialist_{name}"` (current behavior).

---

## Open Questions

- [ ] **Phase 2 feature flag mechanism**: Does the project have an existing feature-flag pattern (config key, env var, runtime toggle)? Needs team confirmation before implementing the guard in `processSpecialistTask`.
- [ ] **`aggregateReports()` call site in standard flow**: `aggregateReports()` currently takes `(reports []SpecialistReport, teamCtx *TeamContext)`. The standard `DelegationTool.Execute` path accumulates results as `string` returns, not `SpecialistReport` slices. The orchestrator's `lastReport` field only holds the most recent report. Clarify: should we accumulate all reports in a slice (new field on `OrchestratorAgent`) or call `aggregateReports()` with only the last report per delegation round?

---

## Related Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Exploration | `openspec/changes/multi-agent-chat-unification/explore.md` | Complete |
| Proposal | `openspec/changes/multi-agent-chat-unification/proposal.md` | Complete |
| Specification | `openspec/changes/multi-agent-chat-unification/spec.md` | Complete |
| Design | `openspec/changes/multi-agent-chat-unification/design.md` | **This document** |
| Tasks | `openspec/changes/multi-agent-chat-unification/tasks.md` | Pending |
