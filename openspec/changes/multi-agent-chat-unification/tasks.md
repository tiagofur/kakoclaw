# Tasks: Multi-Agent Chat Unification

## Phase 1: Foundation — Backend Filtering & Events

- [x] 1.1 `pkg/storage/chat.go` — Add `AND session_id NOT LIKE 'specialist_%'` to `WHERE` clause in both branches of `ListSessionsForUser()` (primary and legacy fallback)
- [x] 1.2 `pkg/agent/loop.go` — Document `synthesis_start` / `synthesis_end` as valid `AgentStatusEvent.Status` values (doc comment only, no struct change)
- [x] 1.3 `pkg/agent/orchestrator.go` — In `DelegationTool.Execute()`, emit `emitAgentStatus(ctx, {Status: "synthesis_start"})` before the final orchestrator LLM call and `synthesis_end` after; guard with `len(specialistResponses) > 0`

## Phase 2: Core Implementation — Frontend Components

- [x] 2.1 Create `pkg/web/frontend/src/components/Chat/SpecialistSegment.vue` — accordion collapsed by default, props: `segment` + `isStreaming`; header shows specialist name + status badge (working=yellow/complete=green/error=red) + 2-line preview; `<Transition name="expand">` for toggle
- [x] 2.2 Create `pkg/web/frontend/src/components/Chat/SynthesisIndicator.vue` — renders only when `chatStore.synthesizing === true`; shows icon + "Orchestrator synthesizing…" + pulsing-dots animation
- [ ] 2.3 `pkg/web/frontend/src/stores/chatStore.js` — Add `synthesizing: ref(false)`, `lastActiveSessionId: ref(null)`, and `setSynthesizing(val)` action; handle `synthesis_start/end` status values in the existing `agent_status` dispatch

## Phase 3: Integration — Wiring Components into Views

- [x] 3.1 `pkg/web/frontend/src/components/MessageBubble.vue` — Replace current `showSegments` toggle block with `<SpecialistSegment v-for="segment in msg.segments">` (render before main content; skip when `msg.segments` is empty/undefined)
- [x] 3.2 `pkg/web/frontend/src/views/ChatView.vue` — (a) Remove `clearAgentStatus()` from `stream_end` handler; (b) Add `beforeRouteLeave` guard calling `clearAgentStatus()`; (c) Add session-change guard (`currentSessionId !== lastActiveSessionId`) at start of `sendMessage()`; update `lastActiveSessionId` after each send
- [x] 3.3 `pkg/web/frontend/src/views/ChatView.vue` — Import and render `<SynthesisIndicator v-if="chatStore.synthesizing" />` above `<TeamActivityPanel>`; wire `synthesis_start/end` to `chatStore.setSynthesizing()` in `handleMessage()`

## Phase 4: Phase 2 — Session Propagation & Aggregation

- [x] 4.1 `pkg/agent/orchestrator.go` — Add typed key `sessionKeyCtxKey{}` + helpers `contextWithSessionKey()` / `sessionKeyFromCtx()`; modify `processSpecialistTask()` to extract key from context and pass to specialist; fallback to `""` when absent ⚑ security-sensitive (cross-user session isolation)
- [ ] 4.2 `pkg/agent/specialist.go` — Add `sessionKey string` param to `ProcessWithSpeciality()` (or add `ProcessWithSpecialityAndSession()` overload); use sessionKey in `ProcessDirect()` call; fallback to `fmt.Sprintf("specialist_%s", sa.name)` when empty
- [ ] 4.3 `pkg/web/server.go` — Inject user `sessionKey` into context via `contextWithSessionKey()` before orchestrator call; guard with feature flag `FEATURE_SESSION_PROPAGATION` ⚑ security-sensitive
- [x] 4.4 `pkg/agent/orchestrator.go` — Call `aggregateReports()` in standard `runAgentLoop` delegation path after all `DelegationTool.Execute()` complete; guard with `len(specialistResponses) > 0`

## Phase 5: Testing

- [ ] 5.1 `pkg/storage/chat_storage_test.go` — Insert `specialist_developer` session, assert absent from `ListSessionsForUser()` result; assert `web:chat:*` sessions still returned
- [ ] 5.2 `pkg/agent/orchestrator_test.go` — Mock `AgentStatusCallback`; assert `synthesis_start` emitted before and `synthesis_end` after final LLM call; assert neither emitted when no specialists delegated
- [ ] 5.3 `pkg/agent/orchestrator_test.go` (Phase 2) — Empty context → `sessionKeyFromCtx` returns `""`; populated context → returns correct key
- [ ] 5.4 Vitest — `chatStore`: `setSynthesizing(true/false)` toggles flag; `stream_end` does NOT call `clearAgentStatus()`; different `lastActiveSessionId` triggers clear on `sendMessage()`
- [ ] 5.5 Vue Test Utils — `SpecialistSegment`: collapsed by default; click toggles; badge color per status; 2-line preview shown in header
- [ ] 5.6 Vue Test Utils — `SynthesisIndicator`: mounts when `synthesizing=true`; unmounts when `false`
- [ ] 5.7 Playwright `e2e/multi-agent-chat.spec.js` — Sidebar shows no `specialist_*` entries after multi-agent chat; accordion expand/collapse; synthesis indicator lifecycle (mock WS events); `agentHistory` preserved within same session

## Phase 6: Cleanup & Documentation

- [ ] 6.1 `docs/README.md` — Add "Multi-Agent Chat" section: specialist segments as collapsible accordions, synthesis indicator, no ghost sessions in sidebar
- [ ] 6.2 `CHANGELOG.md` — Added: specialist segments (collapsed by default), synthesis indicator; Fixed: specialist sessions excluded from sidebar, agent history preserved within session
- [ ] 6.3 Run `make fmt && make vet` to verify no lint regressions across all modified Go files

## Dependencies & Order

**Batch 1** (no deps): 1.1 → 1.2 → 1.3
**Batch 2** (components before views): 2.1 → 2.2 → 2.3 → 3.1 → 3.2 → 3.3
**Batch 3** (Phase 2, after Fase 1 stable): 4.1 → 4.2 → 4.3 → 4.4
**Batch 4** (parallel after impl): 5.1–5.7
**Batch 5** (last): 6.1 → 6.2 → 6.3

> ⚑ Tasks 4.1 and 4.3 touch session isolation — review for cross-user data leakage before merging.
