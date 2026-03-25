# Specification: Agent Visibility Realtime

**Change:** `agent-visibility-realtime`
**Status:** Ready for Design
**Created:** 2026-03-25
**Based on:** [proposal.md](proposal.md)

---

## Overview

This is a **full spec** (new domain — no prior agent-visibility spec exists).

Covers two phases of improvement to agent observability in the MakoClaw web frontend:

- **Fase 1 (Frontend-Only):** Smart collapse of tool calls + per-agent tool-call association in multi-agent panels.
- **Fase 2 (Backend + Frontend, opt-in):** Extended Thinking stream visualization via a new `ThinkingBlock.vue` component with user-configurable opt-in.

---

## Fase 1 — Tool Call Visibility

### Domain: `ui/tool-calls`

---

### Requirement: FR-1 — Historical tool calls MUST be collapsed by default

In any message where `msg.streaming === false`, all tool call items MUST render in the collapsed state. No tool call SHALL auto-expand on initial render of a historical message.

| Rule | Strength |
|------|----------|
| `ToolCallItem` collapsed state MUST show: name + status badge | MUST |
| `ToolCallItem` collapsed state MUST NOT render `args` or `result` content | MUST |
| A click on the collapsed header MUST toggle to expanded state | MUST |
| Expanded state MUST show: name, status badge, args (JSON pretty-printed), result, timestamp | MUST |

#### Scenario: User views historical message with tool calls

- GIVEN a message with `msg.streaming === false` and 5 tool calls of varying statuses
- WHEN the message renders
- THEN all 5 `ToolCallItem` components display in collapsed state
- AND each shows only the tool name and its status badge

#### Scenario: User expands a historical tool call

- GIVEN a collapsed `ToolCallItem` in a historical message
- WHEN the user clicks the item header
- THEN the item animates open
- AND displays: tool name, status badge, args (JSON pretty-printed), result, timestamp
- AND clicking again collapses it

---

### Requirement: FR-2 — Active tool calls during streaming MUST auto-expand

During a live streaming message, a `ToolCallItem` with `status === 'started'` MUST render in expanded state automatically. When the status transitions away from `'started'`, the item MUST collapse automatically.

| Rule | Strength |
|------|----------|
| `expanded = msg.streaming && tc.status === 'started'` is the governing rule | MUST |
| A tool call with `status !== 'started'` in a streaming message MUST be collapsed | MUST |
| State transition from `started` → `finished`/`error` MUST collapse without user action | MUST |
| User MAY manually collapse an auto-expanded item; manual state persists until next status change | MAY |

#### Scenario: Active tool call auto-expands during streaming

- GIVEN an active streaming message
- WHEN a `tool_call` WS event arrives with `status: 'started'`
- THEN the corresponding `ToolCallItem` renders expanded
- AND the args content is visible in real time

#### Scenario: Completed tool call collapses automatically

- GIVEN a `ToolCallItem` with `status: 'started'` (currently expanded)
- WHEN a `tool_call` WS event updates the same tool call with `status: 'finished'`
- THEN the item collapses automatically
- AND the status badge updates to the green "done" color

---

### Requirement: FR-5 — Status badge MUST use semantic color coding

The status badge on every `ToolCallItem` MUST use a distinct, semantic color.

| Status | Badge color | Label |
|--------|-------------|-------|
| `started` | Yellow | `executing…` |
| `finished` | Green | `done` |
| `error` | Red | `error` |

#### Scenario: Error tool call displays red badge

- GIVEN a tool call where the server returned an error result
- WHEN the tool call event arrives with `status: 'error'`
- THEN the `ToolCallItem` badge renders red with label "error"
- AND the expanded view shows the error message in the result field
- AND the item remains collapsed in a historical message (user must click to see the error detail)

#### Scenario: Multiple retried tool calls with same name

- GIVEN an agent retried tool `search` twice: first with `status: 'error'`, second with `status: 'finished'`
- WHEN the historical message renders
- THEN both `ToolCallItem` entries are visible (distinct by timestamp)
- AND each shows its own status badge (red, then green)

---

### Requirement: FR-3 & FR-4 — TeamActivityPanel and AgentActivityItem MUST show per-agent tool calls

`chatStore.js` MUST attach `agentName` to each tool call object at the moment the `tool_call` WS event is received, using `currentAgent.value` at that instant.

`TeamActivityPanel` MUST derive its per-agent tool call lists via a computed from `chatStore` (not from local props), using `groupBy(streamingMsg.toolCalls, 'agentName')`.

`AgentActivityItem` MUST render a "Tool Calls" section when expanded, filtering `msg.toolCalls` by `tc.agentName === activity.agent`.

| Rule | Strength |
|------|----------|
| `currentAgent` null at event time MUST fall back to `"main"` | MUST |
| `TeamActivityPanel` MUST NOT show an empty tool-call section when an agent has 0 active tools | MUST |
| In-flight tool calls (status `started`) MUST be visually distinguishable from completed ones | MUST |
| `AgentStatusIndicator` MUST display agent name + current tool name during active execution | MUST |

#### Scenario: Multi-agent orchestration with active specialist

- GIVEN an orchestrator has delegated to a "developer" specialist
- AND the specialist is currently executing tool `code_analysis` (status: `started`)
- WHEN the `TeamActivityPanel` renders
- THEN the developer agent's row shows `code_analysis` tool as in-flight
- AND `AgentStatusIndicator` shows "developer — code_analysis"

#### Scenario: AgentActivityItem expanded with in-flight tools

- GIVEN the user expands the `AgentActivityItem` for the "developer" specialist
- WHEN the panel renders
- THEN tool calls with `tc.agentName === 'developer'` are listed
- AND the active `code_analysis` tool call is shown with yellow badge and is expanded
- AND completed tool calls (`search`) are shown collapsed with green badge

#### Scenario: Agent with no active tools — no empty section shown

- GIVEN a specialist agent has completed all its tool calls (`status: 'finished'`)
- WHEN `TeamActivityPanel` renders that agent's row
- THEN no empty "Tool Calls" section is rendered for that agent
- AND only the agent status badge is shown

#### Scenario: `currentAgent` null fallback

- GIVEN `currentAgent.value` is null when a `tool_call` WS event arrives
- WHEN `chatStore` processes the event
- THEN the tool call is assigned `agentName = 'main'`
- AND the tool call appears under the "main" agent section in `TeamActivityPanel`

---

## Fase 2 — Extended Thinking Visibility

### Domain: `ui/extended-thinking`

---

### Requirement: FR-6 — ThinkingBlock MUST be gated behind user opt-in

`ThinkingBlock.vue` MUST NOT render for any user unless `userConfig.extended_thinking === true`. Backend MUST NOT emit `thinking_delta` WS events for sessions where the user has `extended_thinking: false`.

| Rule | Strength |
|------|----------|
| Default value of `extended_thinking` in `UserConfig` MUST be `false` | MUST |
| Frontend MUST silently ignore `thinking_delta` events if `extended_thinking` is false | MUST |
| Backend MUST check `extended_thinking` flag before activating `OnThinking` callback | MUST |
| Toggle in `ProfileSettingsTab.vue` MUST persist to `/api/v1/user/config` | MUST |
| Persisted value MUST survive page reload | MUST |
| Feature MUST work only with Claude models that support thinking blocks | MUST |
| If model does not support thinking, backend MUST NOT emit `thinking_delta` regardless of flag | MUST |

#### Scenario: User with extended thinking OFF — no ThinkingBlock visible

- GIVEN a user has `extended_thinking: false` (default)
- WHEN the backend processes a streaming response, even if the underlying model emits thinking blocks
- THEN no `thinking_delta` WS event is sent to the client
- AND `ThinkingBlock.vue` is never mounted in `MessageBubble.vue`

#### Scenario: User activates extended thinking toggle in settings

- GIVEN a user navigates to `ProfileSettingsTab` and enables the toggle
- WHEN the setting is saved via `PUT /api/v1/user/config`
- THEN `userConfig.extended_thinking` is set to `true` server-side
- AND upon page reload the toggle remains ON

---

### Requirement: FR-7 & FR-8 — ThinkingBlock MUST animate during streaming and collapse on completion

`ThinkingBlock.vue` MUST auto-open and display an animated "thinking…" indicator when `thinking_delta` events are arriving. On `stream_end`, the block MUST auto-close unless the user manually opened it.

`chatStore.js` MUST accumulate `thinking_delta` content into `msg.thinkingBlocks[]` (in-session only — not persisted to SQLite).

| Rule | Strength |
|------|----------|
| `ThinkingBlock` expanded state MUST render thinking content in italic grey text | MUST |
| `ThinkingBlock` collapsed state MUST show brain icon + "Thinking" label | MUST |
| During streaming: auto-open with animation | MUST |
| On `stream_end`: auto-close UNLESS user manually opened the block | MUST |
| Thinking deltas arriving after `stream_end` MUST still be appended to `msg.thinkingBlocks[]` | MUST |
| `msg.thinkingBlocks[]` MUST NOT be persisted to SQLite | MUST |
| `msg.thinkingBlocks[]` structure: `[{ id, content, expanded }]` | MUST |

#### Scenario: Streaming complex query with extended thinking ON

- GIVEN a user with `extended_thinking: true` sends a complex prompt
- WHEN the Claude model begins streaming and emits thinking blocks
- THEN `ThinkingBlock.vue` mounts and auto-opens in `MessageBubble.vue`
- AND content appears in italic grey text as deltas arrive
- AND an animated indicator shows the agent is thinking

#### Scenario: ThinkingBlock auto-collapses on stream end

- GIVEN `ThinkingBlock.vue` is auto-open during streaming
- AND the user has NOT manually interacted with the block
- WHEN `stream_end` is received
- THEN `ThinkingBlock.vue` closes automatically
- AND the main message content follows below

#### Scenario: User manually expands ThinkingBlock — persists after stream end

- GIVEN `ThinkingBlock.vue` is auto-open during streaming
- AND the user clicks to collapse, then re-opens it manually
- WHEN `stream_end` is received
- THEN the block remains open (respects user's explicit action)

#### Scenario: Late thinking delta arrives after stream_end

- GIVEN `stream_end` has been received
- WHEN a `thinking_delta` event arrives after the stream ended
- THEN the content is appended to `msg.thinkingBlocks[]`
- AND the block does NOT re-open automatically (already closed)

---

### Requirement: FR-9 — Multi-agent thinking attribution (agent badge in ThinkingBlock)

In a multi-agent context, each `ThinkingBlock` MUST display a badge identifying which agent produced the thinking content.

| Rule | Strength |
|------|----------|
| `ThinkingBlock` MUST receive `agentName` prop | MUST |
| Badge MUST be visible in both collapsed and expanded states | MUST |
| Orchestrator thinking and specialist thinking MUST be visually distinct via the agent badge | MUST |

#### Scenario: Specialist thinking is attributed correctly

- GIVEN orchestrator delegates to a "researcher" specialist
- AND the specialist uses extended thinking
- WHEN `ThinkingBlock` renders for the specialist's message
- THEN a badge reading "researcher" is visible on the block
- AND this distinguishes it from any orchestrator `ThinkingBlock`

---

## Data Structures

### `chatStore.js` — Message Shape (additions)

```js
// Tool call object (agentName is NEW in Fase 1)
msg.toolCalls = [{
  id,          // string — unique per tool call
  name,        // string — tool function name
  args,        // object — input arguments
  result,      // string | object — tool output
  status,      // 'started' | 'finished' | 'error'
  expanded,    // boolean — UI state
  timestamp,   // ISO string
  agentName    // string — NEW: agent that issued the call (fallback: 'main')
}]

// Thinking blocks array (NEW in Fase 2 — session-only, not persisted)
msg.thinkingBlocks = [{
  id,        // string — unique per block
  content,   // string — accumulated thinking text
  expanded   // boolean — UI state
}]
```

### WS Events

```js
// Existing event (no change to shape, agentName derived client-side)
{ type: 'tool_call', name, args, result, status }

// NEW event — Fase 2, only emitted when extended_thinking: true
{ type: 'thinking_delta', content }
```

---

## Edge Cases

| ID | Description | Expected Behavior |
|----|-------------|-------------------|
| Edge-1 | `currentAgent` is null when `tool_call` arrives | Fall back to `agentName = 'main'` |
| Edge-2 | Agent has 0 active tools | `TeamActivityPanel` MUST NOT render empty tool-call section |
| Edge-3 | `thinking_delta` arrives after `stream_end` | Append to `msg.thinkingBlocks[]`; do NOT re-open block |
| Edge-4 | `extended_thinking: true` but model doesn't support thinking | Backend MUST NOT emit `thinking_delta`; no crash |
| Edge-5 | Mobile app receives `tool_call` or `thinking_delta` event | MUST NOT crash; event handling is out of scope but graceful ignore is required |

---

## Non-Functional Requirements

### Performance

| Requirement | Strength |
|-------------|----------|
| `TeamActivityPanel` computed from store MUST NOT cause unnecessary re-renders on unrelated store updates | SHOULD |
| `thinking_delta` append MUST be O(1) — append to array, do not re-sort or re-parse | MUST |
| `ThinkingBlock` animation MUST use CSS transitions, not JS timers | SHOULD |

### Security

| Requirement | Strength |
|-------------|----------|
| `thinking_delta` events MUST only be emitted to the session owner's WS connection | MUST |
| Extended thinking opt-in MUST be stored in per-user config, not shared config | MUST |

### Rollback

| Requirement | Strength |
|-------------|----------|
| Fase 1 changes are presentation-only; reverting `MessageBubble.vue` and `ToolCallItem.vue` fully restores prior behavior | MUST |
| Fase 2 is gated by `extended_thinking: false` default; disabling the UI toggle is sufficient to deactivate | MUST |

---

## Acceptance Criteria Summary

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| F1-AC-1 | Historical message: all tool calls collapsed on render | Playwright: assert no expanded items on load of a historical message |
| F1-AC-2 | Streaming: `started` tool call auto-expands | Playwright: mock WS `tool_call` event with `status: started`, assert expansion |
| F1-AC-3 | Streaming: completed tool call auto-collapses | Playwright: transition `started` → `finished`, assert collapse |
| F1-AC-4 | Status badges use correct semantic colors | Playwright/visual: assert badge class per status |
| F1-AC-5 | `TeamActivityPanel` shows tool name per agent during multi-agent | Playwright: inject multi-agent WS sequence, assert panel content |
| F1-AC-6 | `AgentActivityItem` expanded shows filtered tool calls | Playwright: expand activity item, assert only that agent's tools shown |
| F1-AC-7 | `currentAgent` null → agentName = 'main' | Unit: dispatch `tool_call` with null `currentAgent`, inspect store state |
| F2-AC-1 | `extended_thinking: false` → no `ThinkingBlock` in DOM | Playwright: default user, stream response, assert no ThinkingBlock mounted |
| F2-AC-2 | `extended_thinking: true` → ThinkingBlock animates then collapses | Playwright: activate toggle, stream response with thinking deltas, verify behavior |
| F2-AC-3 | Toggle persists across reload | E2E: toggle ON, reload, assert toggle still ON |
| F2-AC-4 | Specialist ThinkingBlock shows agent name badge | Playwright: multi-agent with specialist thinking, assert badge label |
| F2-AC-5 | Late `thinking_delta` appended without re-opening block | Unit: dispatch delta after `stream_end`, assert `thinkingBlocks` updated, block closed |

---

## Related Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Proposal | `openspec/changes/agent-visibility-realtime/proposal.md` | Complete |
| Specification | `openspec/changes/agent-visibility-realtime/spec.md` | This document |
| Design | `openspec/changes/agent-visibility-realtime/design.md` | Pending |
| Tasks | `openspec/changes/agent-visibility-realtime/tasks.md` | Pending |
