# Apply Progress: agent-visibility-realtime

## Batch 1 — Phase 1 Frontend Only

- Implemented `currentAgent` tracking and `streamingMessage` access in `chatStore.js`.
- Added `agent_status` synchronization in `ChatView.vue` and normalized historical tool call args.
- Refactored tool call UI to computed-driven expansion and propagated tool visibility into message, team, and agent activity panels.
- Enhanced `AgentStatusIndicator.vue` to display the active tool beneath the current agent.

## Completed Tasks

- [x] 1.1
- [x] 1.2
- [x] 1.3
- [x] 1.4
- [x] 1.5
- [x] 1.6
- [x] 1.7

## Notes

- Historical tool calls now stay collapsed by default because expansion is derived from `msg.streaming` and the current tool call status.
- Team activity tool lists are derived from store state instead of watcher-managed local copies.
- No build was run.

## Batch 4 — Testing (Batch 1)

- Added Vitest + Vue Test Utils setup in `pkg/web/frontend` for store and component unit tests.
- Created `chatStore.spec.js` covering `agentName` attribution, `'main'` fallback, streaming expansion, and `updateCurrentAgent()` state updates.
- Created `ToolCallItem.spec.js` covering streaming auto-expand, completed collapsed state, manual toggle behavior, and semantic status badges.
- Left task 3.1 unchecked in `tasks.md` because the full task still includes `appendThinkingDelta()` and `endStreamingMessage()` assertions from Phase 2.
