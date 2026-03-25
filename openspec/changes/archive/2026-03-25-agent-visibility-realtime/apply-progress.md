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

## Batch 2 — Phase 2 Backend

- Extended `providers.StreamChunk` and the agent streaming loop to propagate Claude thinking deltas through a dedicated `OnThinking` callback.
- Implemented Claude streaming with Anthropic `Messages.NewStreaming()`, including text deltas, tool-call fragments, finish reasons, usage, and `thinking_delta` emission.
- Added persisted `extended_thinking` user preference support (runtime config endpoint + SQLite migrations) and gated WebSocket `thinking_delta` events behind both the user flag and Anthropic model detection.

## Completed Tasks

- [x] 2.1
- [x] 2.2
- [x] 2.3
- [x] 2.4
- [x] 2.5

## Notes

- `pkg/storage/user_storage.go` did not previously contain a `UserConfig` struct, so a lightweight runtime config type was introduced there while persistence is backed by the `users.extended_thinking` column.
- Added `PUT /api/v1/user/config` and matching `GET /api/v1/user/config` for the frontend toggle contract.
- No build was run.

## Batch 3 — Phase 2 Frontend

- Added client-side accumulation of `thinking_delta` chunks in `chatStore.js`, including auto-finalization/collapse behavior on `stream_end` and late-delta appends to the latest assistant message.
- Wired `ChatView.vue` to forward `thinking_delta` WebSocket events into the store and rendered the new `ThinkingBlock.vue` component before assistant message content.
- Added a Claude-only "Extended Thinking" toggle in `ProfileSettingsTab.vue` backed by `GET/PUT /api/v1/user/config` and hydrated model metadata when needed.

## Completed Tasks

- [x] 2.6
- [x] 2.7
- [x] 2.8
- [x] 2.9
- [x] 2.10

## Notes

- `ThinkingBlock.vue` gives manual toggle precedence over auto-expand so the user can collapse/re-open the block during streaming and preserve that choice after `stream_end`.
- Thinking blocks now carry `agentName` in addition to the minimal spec shape so the UI can satisfy the per-agent attribution badge requirement.
- No build was run.

## Batch 4 — Testing (Remaining)

- Extended `chatStore.spec.js` with Phase 2 coverage for thinking-block accumulation and `endStreamingMessage()` collapse behavior.
- Added `ThinkingBlock.spec.js` covering streaming expansion, post-stream collapse, manual toggle persistence, agent badge rendering, and key styling hooks.
- Added Playwright coverage in `e2e/tests/agent-visibility.spec.ts` using mocked API + WebSocket flows for historical tool calls, Extended Thinking persistence, multi-agent activity grouping, and the default opt-out path.

## Completed Tasks

- [x] 3.1
- [x] 3.3
- [x] 3.4

## Notes

- The `appendThinkingDelta()` assertions were aligned with the current spec/design behavior: deltas concatenate into the active block instead of creating a new block per chunk.
- Playwright scenarios are mocked at the browser boundary so the tests stay deterministic without depending on live backend orchestration.
- No build was run.

## Batch 5 — Documentation

- Added an "Agent Visibility" section to `pkg/web/frontend/README.md` covering smart tool-call collapse, Extended Thinking, and multi-agent tool visibility.
- Updated `CHANGELOG.md` under `Unreleased` with Added, Changed, Fixed, and Technical Notes entries for the realtime visibility work.

## Completed Tasks

- [x] 4.1
- [x] 4.2

## Notes

- Documentation was added to the frontend README because the feature is specific to the web chat experience and that README already documents frontend capabilities.
- No build was run.
