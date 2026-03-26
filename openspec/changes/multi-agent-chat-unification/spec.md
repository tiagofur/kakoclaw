# Specification: Multi-Agent Chat Unification

**Change:** `multi-agent-chat-unification`
**Status:** Ready for Design
**Created:** 2026-03-25
**Based on:** [proposal.md](proposal.md), [explore.md](explore.md)

---

## Overview

This is a **full spec** (new domain — no prior multi-agent-chat-unification spec exists).

Covers two phases that eliminate ghost sessions, surface specialist reports as collapsible accordions, add orchestrator synthesis indicators, and propagate user session context to specialists.

---

## Phase 1 — Quick Wins

### Domain: `storage/session-filtering`

---

### Requirement: FR-1 — `ListSessionsForUser` MUST exclude specialist sessions

`ListSessionsForUser()` in `pkg/storage/chat.go` MUST filter out any session whose `session_id` starts with `specialist_`. This is a read-only filter — no writes or schema changes.

| Rule | Strength |
|------|----------|
| Query MUST apply `WHERE session_id NOT LIKE 'specialist_%'` (or equivalent) | MUST |
| Filter MUST apply in all modes (per-user DB and legacy shared DB) | MUST |
| Sessions created by `specialist_*` MUST still exist in DB (no deletion) | MUST |
| The API response `GET /api/v1/chat/sessions` MUST NOT include specialist sessions | MUST |

#### Scenario: User with multi-agent chat — no ghost sessions in sidebar

- GIVEN a user has session `web:chat:abc123` and the DB also contains `specialist_developer`, `specialist_researcher`
- WHEN the user's client calls `GET /api/v1/chat/sessions`
- THEN the response includes only `web:chat:abc123`
- AND `specialist_developer` and `specialist_researcher` are NOT in the list

#### Scenario: User with no specialist sessions — unaffected

- GIVEN a user has only standard sessions (`web:chat:*`)
- WHEN the client calls `GET /api/v1/chat/sessions`
- THEN all standard sessions are returned normally
- AND no filtering side effects occur

---

### Domain: `ui/specialist-segments`

---

### Requirement: FR-2 — `MessageBubble.vue` MUST render `msg.segments` as collapsible accordions

Each element in `msg.segments` MUST render as an independent collapsible accordion. The component MAY reuse the existing `ToolCallItem` collapse pattern.

| Rule | Strength |
|------|----------|
| Each segment accordion header MUST show: specialist name + status badge + content preview (first 2 lines) | MUST |
| Accordion MUST default to collapsed state on initial render | MUST |
| A single click on the header MUST toggle expanded/collapsed | MUST |
| Expanded state MUST show the full specialist content | MUST |
| Status badge values: `working` (yellow), `complete` (green), `error` (red) | MUST |
| Transition MUST use CSS `<Transition name="expand">` — no JS timers | MUST |
| When `msg.segments` is empty or undefined, no accordion section MUST render | MUST |

#### Scenario: Message with specialist segments renders collapsed by default

- GIVEN a historical message has `msg.segments = [{specialist: "developer", status: "complete", content: "..."}]`
- WHEN `MessageBubble.vue` renders the message
- THEN the segment accordion is collapsed (only header visible)
- AND the header shows "developer", a green "complete" badge, and the first 2 lines of content

#### Scenario: User expands a specialist accordion

- GIVEN a collapsed specialist accordion in `MessageBubble.vue`
- WHEN the user clicks the accordion header
- THEN the accordion expands with a CSS transition
- AND the full specialist content is visible

#### Scenario: User collapses an expanded accordion

- GIVEN an expanded specialist accordion
- WHEN the user clicks the header again
- THEN the accordion collapses with a CSS transition
- AND only the header is visible

#### Scenario: Message with no segments — no accordion section rendered

- GIVEN a message where `msg.segments` is empty (`[]`) or undefined
- WHEN `MessageBubble.vue` renders
- THEN no accordion section is rendered for segments
- AND the message renders normally with no visual anomaly

---

### Domain: `ui/synthesis-indicator`

---

### Requirement: FR-3 / FR-4 — Backend MUST emit `synthesis_start` and `synthesis_end` WS events

`pkg/web/server.go` MUST emit a `synthesis_start` event immediately before the orchestrator's final LLM call and a `synthesis_end` event upon completion.

| Rule | Strength |
|------|----------|
| `synthesis_start` payload: `{type: "synthesis_start", agent: "orchestrator", timestamp: <ISO>}` | MUST |
| `synthesis_end` payload: `{type: "synthesis_end", agent: "orchestrator", timestamp: <ISO>}` | MUST |
| Events MUST be emitted only to the session owner's WS connection | MUST |
| If the orchestrator does not delegate to specialists, events MUST NOT be emitted | SHOULD |

#### Scenario: Orchestrator emits synthesis events during multi-agent flow

- GIVEN an orchestrator delegates to at least one specialist
- WHEN the orchestrator begins its final synthesis LLM call
- THEN a `synthesis_start` WS event is emitted to the client
- WHEN the synthesis LLM call completes
- THEN a `synthesis_end` WS event is emitted to the client

---

### Requirement: FR-5 / FR-6 — Frontend MUST show and hide `SynthesisIndicator` on WS events

`chatStore.js` MUST expose a reactive `synthesizing` flag. `ChatView.vue` MUST show `SynthesisIndicator.vue` while `synthesizing === true`.

| Rule | Strength |
|------|----------|
| On `synthesis_start`: `chatStore.synthesizing = true` | MUST |
| On `synthesis_end`: `chatStore.synthesizing = false` | MUST |
| `SynthesisIndicator.vue` MUST show: icon + "Orchestrator synthesizing…" + loading animation | MUST |
| If `synthesis_end` arrives without a prior `synthesis_start`, `synthesizing` MUST be set to `false` without error | MUST |
| `SynthesisIndicator.vue` MUST NOT render when `synthesizing === false` | MUST |

#### Scenario: Frontend shows synthesis indicator

- GIVEN the client has an active streaming session
- WHEN a `synthesis_start` WS event arrives
- THEN `SynthesisIndicator.vue` mounts and shows "Orchestrator synthesizing…" with loading animation

#### Scenario: Frontend hides synthesis indicator on completion

- GIVEN `SynthesisIndicator.vue` is visible
- WHEN a `synthesis_end` WS event arrives
- THEN `SynthesisIndicator.vue` unmounts (or becomes invisible)

#### Scenario: `synthesis_end` arrives without prior `synthesis_start`

- GIVEN `chatStore.synthesizing === false`
- WHEN a `synthesis_end` event arrives
- THEN `synthesizing` remains `false`
- AND no error or crash occurs

---

### Domain: `ui/agent-history-lifecycle`

---

### Requirement: FR-7 / FR-8 — `clearAgentStatus()` MUST be session-aware

`clearAgentStatus()` in `chatStore.js` MUST NOT clear `agentHistory` during navigation within the same chat session. It MUST clear when navigating to a different session or starting a new chat.

| Rule | Strength |
|------|----------|
| Call site in `stream_end` handler MUST be removed from `ChatView.vue` | MUST |
| `clearAgentStatus()` MUST be called in `beforeRouteLeave` hook in `ChatView.vue` | MUST |
| `clearAgentStatus()` MUST be called at the start of `sendMessage()` ONLY if `currentSessionId` differs from `lastActiveSessionId` | MUST |
| `lastActiveSessionId` MUST be updated to `currentSessionId` after each `sendMessage()` | MUST |
| Navigating back to the original chat MUST show the preserved `agentHistory` | MUST |

#### Scenario: User navigates away — agentHistory clears

- GIVEN a chat with `agentHistory` populated
- WHEN the user clicks a different session in the sidebar
- THEN `beforeRouteLeave` fires and `clearAgentStatus()` is called
- AND `agentHistory` is empty in the new chat

#### Scenario: User returns to original chat — agentHistory persists

- GIVEN `agentHistory` was populated in chat session A
- AND the user navigated to chat session B (clearing history)
- WHEN the user navigates back to session A
- THEN `agentHistory` for session A is visible (it was preserved before the nav away)

#### Scenario: New chat started — agentHistory clears

- GIVEN a chat with `agentHistory` populated
- WHEN the user initiates a new chat (`sendMessage()` with a new session)
- THEN `clearAgentStatus()` is called at the start of `sendMessage()`
- AND `agentHistory` is reset before the new message is sent

#### Scenario: Same-session message — agentHistory NOT cleared

- GIVEN a streaming message completes and `stream_end` arrives
- WHEN `endStreamingMessage()` is called
- THEN `clearAgentStatus()` is NOT called
- AND `agentHistory` remains visible to the user

---

## Phase 2 — Medium Term

### Domain: `agent/session-propagation`

---

### Requirement: FR-9 — Orchestrator MUST propagate user `sessionKey` to specialists

`processSpecialistTask()` in `pkg/agent/orchestrator.go` MUST extract the user's `sessionKey` from the Go context and pass it to `specialist.ProcessWithSpeciality()`.

| Rule | Strength |
|------|----------|
| `sessionKey` MUST be stored in `context.Context` using a typed key (not a plain string key) | MUST |
| `processSpecialistTask()` MUST extract `sessionKey` from context before calling specialist | MUST |
| `specialist.ProcessWithSpeciality()` MUST accept and use the propagated `sessionKey` | MUST |
| If `sessionKey` is absent from context, specialist MUST fall back to `"specialist_{name}"` | MUST |
| This behavior MUST be guarded by a feature flag until validated in staging | SHOULD |

#### Scenario: Specialist receives user session context

- GIVEN a user has session `web:chat:abc123`
- WHEN the orchestrator delegates to the "developer" specialist
- THEN `processSpecialistTask` extracts `web:chat:abc123` from context
- AND the specialist's `ProcessWithSpeciality` call receives `sessionKey = "web:chat:abc123"`

#### Scenario: Specialist accesses user conversation history

- GIVEN the specialist received `sessionKey = "web:chat:abc123"`
- WHEN the specialist queries its context window
- THEN it has access to the prior conversation messages from the user's session

#### Scenario: Missing sessionKey falls back gracefully

- GIVEN `sessionKey` is absent from the Go context (legacy path)
- WHEN `processSpecialistTask` is called
- THEN the specialist falls back to `sessionKey = "specialist_{name}"`
- AND no error is raised; behavior matches pre-Phase-2 state

---

### Requirement: FR-10 — `aggregateReports()` MUST be called in the standard delegation flow

`pkg/agent/orchestrator.go` MUST call `aggregateReports()` after all specialist responses are received in the standard `runAgentLoop` delegation path, not only in `ProcessWithFeedbackLoop()`.

| Rule | Strength |
|------|----------|
| `aggregateReports()` MUST be invoked when 1+ specialists have responded | MUST |
| The aggregated summary MUST include contributions from all specialists that responded | MUST |
| If a specialist timed out, its response MUST be excluded; others MUST still be aggregated | MUST |
| Aggregated summary MUST be used as the orchestrator's final response content | MUST |

#### Scenario: Orchestrator aggregates two specialist responses

- GIVEN the orchestrator delegated to "developer" and "researcher" specialists
- AND both responded successfully
- WHEN the orchestrator completes the delegation round
- THEN `aggregateReports()` is called with both specialist responses
- AND the final message to the user includes insights from both specialists

#### Scenario: One specialist times out — partial aggregation

- GIVEN the orchestrator delegated to "developer" and "researcher"
- AND "researcher" times out before responding
- WHEN the orchestrator completes the delegation round
- THEN `aggregateReports()` is called with only the "developer" response
- AND the final message indicates partial results (researcher did not respond)

---

## Edge Cases

| ID | Description | Expected Behavior |
|----|-------------|-------------------|
| Edge-1 | User has 0 specialists configured | Standard chat flow unchanged; no accordions, no synthesis events |
| Edge-2 | Specialist timeout during multi-agent | Orchestrator continues with available responses; timeout specialist excluded from aggregation |
| Edge-3 | User navigates between sessions of same user | `clearAgentStatus()` fires on each navigation (different `sessionId` triggers clear) |
| Edge-4 | User starts a new chat | `clearAgentStatus()` called at start of `sendMessage()` when session changes |
| Edge-5 | `synthesis_end` before `synthesis_start` | Frontend sets `synthesizing = false`; no crash or error |
| Edge-6 | `msg.segments` has content but specialist `status` is missing | Badge SHOULD default to `working` (yellow); MUST NOT crash |
| Edge-7 | Feature flag for Phase 2 disabled | Specialist falls back to `"specialist_{name}"` session key; behavior identical to pre-Phase-2 |

---

## Data Structures

### WS Events (Backend → Frontend)

```go
// synthesis_start
{ "type": "synthesis_start", "agent": "orchestrator", "timestamp": "<ISO8601>" }

// synthesis_end
{ "type": "synthesis_end", "agent": "orchestrator", "timestamp": "<ISO8601>" }
```

### `chatStore.js` — New State

```js
// Session filtering (computed)
const visibleSessions = computed(() =>
  sessions.value.filter(s => !s.session_id.startsWith('specialist_'))
)

// Synthesis state
const synthesizing = ref(false)

// Session-aware clearAgentStatus
function clearAgentStatus() {
  agentHistory.value = []
  specialistReports.value = []
  delegationChain.value = []
  // reset other agent state fields
}
```

### `msg.segments` shape (existing, documented for reference)

```js
msg.segments = [{
  specialist: string,  // e.g. "developer"
  status: 'working' | 'complete' | 'error',
  content: string,     // full specialist response text
  confidence: number   // optional 0-1 score
}]
```

---

## Non-Functional Requirements

### Performance

| Requirement | Strength |
|-------------|----------|
| Session filter query MUST use an indexed prefix match | SHOULD |
| Accordion CSS transition MUST NOT trigger layout reflow on every keystroke | MUST |
| `clearAgentStatus()` MUST run synchronously and complete before next render cycle | MUST |

### Security

| Requirement | Strength |
|-------------|----------|
| `synthesis_start/end` events MUST only be sent to the session owner's WS connection | MUST |
| Phase 2 `sessionKey` propagation MUST NOT expose one user's session to another user's specialist | MUST |

### Rollback

| Requirement | Strength |
|-------------|----------|
| Phase 1 filter in `ListSessionsForUser` MUST be revertible by removing the WHERE clause — no data loss | MUST |
| Phase 1 frontend changes are presentation-only; reverting `MessageBubble.vue` and `ChatView.vue` restores prior behavior | MUST |
| Phase 2 sessionKey propagation MUST be gated; disabling the feature flag restores `"specialist_{name}"` fallback | MUST |

---

## Acceptance Criteria Summary

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| F1-AC-1 | `GET /api/v1/chat/sessions` excludes `specialist_*` sessions | Integration test: insert specialist session, assert absent from API response |
| F1-AC-2 | Segment accordions collapsed by default in historical message | Playwright: render message with segments, assert no expanded accordion on load |
| F1-AC-3 | Click expands accordion; click again collapses | Playwright: click header, assert expansion; click again, assert collapse |
| F1-AC-4 | Status badges use correct semantic colors | Playwright/visual: assert badge class per segment status |
| F1-AC-5 | `synthesis_start` WS event sets `synthesizing = true` in store | Unit: dispatch event, inspect store |
| F1-AC-6 | `synthesis_end` WS event sets `synthesizing = false` in store | Unit: dispatch event after start, inspect store |
| F1-AC-7 | `SynthesisIndicator` mounts/unmounts based on `synthesizing` flag | Playwright: mock WS events, assert indicator presence/absence |
| F1-AC-8 | `agentHistory` persists across same-session streaming messages | Unit: trigger `stream_end`, assert `agentHistory` not cleared |
| F1-AC-9 | `agentHistory` clears on route leave | Playwright: navigate away, assert store cleared |
| F2-AC-1 | Specialist receives user `sessionKey` from context | Unit: mock `processSpecialistTask` with context, assert specialist called with correct key |
| F2-AC-2 | Fallback to `"specialist_{name}"` if context key absent | Unit: empty context, assert fallback key used |
| F2-AC-3 | `aggregateReports()` called after standard delegation | Unit: mock delegation flow, assert `aggregateReports` invoked |
| F2-AC-4 | Partial aggregation on specialist timeout | Unit: one specialist times out, assert aggregation with remaining responses |

---

## Related Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Exploration | `openspec/changes/multi-agent-chat-unification/explore.md` | Complete |
| Proposal | `openspec/changes/multi-agent-chat-unification/proposal.md` | Complete |
| Specification | `openspec/changes/multi-agent-chat-unification/spec.md` | This document |
| Design | `openspec/changes/multi-agent-chat-unification/design.md` | Pending |
| Tasks | `openspec/changes/multi-agent-chat-unification/tasks.md` | Pending |
