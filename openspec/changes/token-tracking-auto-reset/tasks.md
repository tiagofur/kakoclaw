# Tasks: Token Tracking + Auto-Reset

## Phase 1: Foundation

- [ ] 1.1 **[RED]** Write failing tests for `SessionTracker` in `pkg/web/session_tracker_test.go`: `Record` accumulates `NumTurns`/`CostUSD`, `ShouldReset` returns `false` at exactly limit and `true` when over, `ZeroSession` resets counters, file persistence round-trips, missing file starts empty, concurrent `Record` passes `-race`
- [ ] 1.2 **[GREEN]** Create `pkg/web/session_tracker.go`: `SessionTracker` struct (mutex + map + `stateDir` + `maxTokens`), `NewSessionTracker`, `Record`, `ShouldReset` (strict `>`), `ZeroSession`, `Stats`, atomic JSON persistence (`write tmp → os.Rename`); log warning on write failure, never abort
- [ ] 1.3 **[REFACTOR]** Verify all `session_tracker_test.go` scenarios pass under `go test -race ./pkg/web/...`
- [ ] 1.4 Add `EventSessionReset = "session_reset"` constant and `TokensUsed int` field to `pkg/bridge/events.go`

## Phase 2: Backend Integration

- [ ] 2.1 Add `sessionTracker *SessionTracker` field to `Server` in `pkg/web/server.go`; initialise with `NewSessionTracker(workspaceDir+"/bridge", cfg.MaxSessionTokens)` in `NewServer`; register route `GET /api/v1/dev/session/stats` → `handleSessionStats` handler (auth required via existing `extractClaims`; return `SessionStats` JSON; `200 OK` with zero values when no session)
- [ ] 2.2 **[RED]** Write failing integration test for `GET /api/v1/dev/session/stats` in `pkg/web/server_test.go` (or new `handlers_dev_stats_test.go`): authenticated returns `200` with correct JSON; unauthenticated returns `401`
- [ ] 2.3 **[GREEN]** Implement `handleSessionStats` in `pkg/web/handlers_dev.go` or a new `handlers_dev_stats.go`; wire to `s.sessionTracker.Stats(userUUID, sessionID)`
- [ ] 2.4 Modify `handleDevQuery` in `pkg/web/handlers_dev.go`: after each `result` event, call `s.sessionTracker.Record(..., ev.NumTurns, ev.CostUSD)`; if `ShouldReset` → write `session_reset` synthetic event to stream, `newID := uuid.NewString()`, call `b.Stop()`/`b.Start(ctx)`, call `tracker.ZeroSession(userUUID, newID)`
- [ ] 2.5 Modify `handleDevTerminalWS` in `pkg/web/handlers_dev_ws.go`: same `Record`/`ShouldReset`/inject logic as 2.4 for the WebSocket streaming loop

## Phase 3: Frontend

- [ ] 3.1 **[RED]** Write Vitest tests for `devStudioStore.js` (`pkg/web/frontend/src/stores/devStudioStore.test.js`): `result` event increments `sessionTokens`; `session_reset` event zeros counters and updates `sessionId`; panel-open fetch populates state from stats endpoint mock
- [ ] 3.2 **[GREEN]** Modify `pkg/web/frontend/src/stores/devStudioStore.js`: add `sessionTokens`, `sessionCostUsd`, `tokenLimit`, `numTurns` refs; handle `result` (accumulate) and `session_reset` (zero + update `sessionId` + push notification line); fetch `/api/v1/dev/session/stats` on panel open
- [ ] 3.3 **[RED]** Write Vitest + `@vue/test-utils` tests for `TerminalHeader.vue`: badge text `"{tokens} / {limit} tokens"`; warning class at ≥90%; danger class at ≥100%
- [ ] 3.4 **[GREEN]** Create `pkg/web/frontend/src/components/DevStudio/TerminalHeader.vue`: pill badge bound to `sessionTokens`/`tokenLimit` from store; computed colour class (default/warning/danger)

## Phase 4: Verification

- [ ] 4.1 Run `go test -race ./pkg/web/... ./pkg/bridge/...` — all tests green, zero race conditions
- [ ] 4.2 Run `npm run test` (Vitest) under `pkg/web/frontend` — all store and component tests green
- [ ] 4.3 Manual smoke test: send queries until limit exceeded; confirm `session_reset` event, badge resets, new `sessionID` on subsequent query
