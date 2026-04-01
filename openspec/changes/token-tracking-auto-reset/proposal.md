# Proposal: Token Tracking + Auto-Reset (Dev Studio)

## Intent

`MaxSessionTokens` exists in `DevStudioConfig` but is never read. Bridge result events already carry `cost_usd`, `duration_ms`, and `num_turns`. This change wires those together: accumulate tokens per user/session, enforce the configured limit with an automatic session reset, and surface usage in the frontend terminal.

## Scope

### In Scope
- **Token + cost accumulation** from `bridge.Event` result events per `userUUID+sessionID`
- **Auto-reset** when accumulated tokens exceed `MaxSessionTokens` (new session ID, user notification event)
- **Persistence** of token counters via JSON file (`{workspace}/bridge/{userUUID}-session-tokens.json`) — survives bridge restarts, no new DB dependency
- **REST endpoint** `GET /api/v1/dev/session/stats` returning current token count, cost, and limit
- **Frontend badge/progress bar** in the Dev Studio terminal header showing `tokens used / max`
- **Pinia store** update: `sessionTokens`, `sessionCostUsd`, `tokenLimit` reactive state

### Out of Scope
- Per-project token budgets (all projects share the user limit today)
- Token history / audit log across sessions
- UI settings screen for editing `MaxSessionTokens` (already exists in user config API)
- Backend LLM provider token counting (separate from Dev Studio bridge)

## Approach

New `pkg/web/session_tracker.go` holds a `SessionTracker` struct (mutex-protected map, file-backed). `handleDevQuery` and the WebSocket handler both call `tracker.Record(userUUID, sessionID, tokens, costUSD)` after each result event. If the new total exceeds the limit, `tracker.ShouldReset()` returns true — the handler injects a synthetic `session_reset` event into the stream, generates a new `sessionID`, and calls `b.Reset()` on the bridge.

File persistence uses atomic write (write temp → rename). No SQLite migration needed.

Frontend: `devStudioStore.js` receives `session_reset` and `result` events; updates reactive counters; terminal header shows a pill badge.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/web/session_tracker.go` | **New** | Token accumulator, file persistence, reset logic |
| `pkg/web/handlers_dev.go` | Modified | Call tracker after result event; inject reset notification |
| `pkg/web/handlers_dev_ws.go` | Modified | Same tracker call for WebSocket path |
| `pkg/web/server.go` | Modified | Register `GET /api/v1/dev/session/stats` |
| `pkg/bridge/events.go` | Modified | Add `session_reset` synthetic event type constant |
| `pkg/web/frontend/src/stores/devStudioStore.js` | Modified | `sessionTokens`, `sessionCostUsd`, `tokenLimit`; handle `session_reset` |
| `pkg/web/frontend/src/components/DevStudio/TerminalHeader.vue` | Modified | Token usage badge |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Race condition on concurrent WS + HTTP queries | Medium | Mutex in `SessionTracker`; tested with `-race` |
| File write fails on constrained hardware | Low | Log warning, continue without persistence (non-fatal) |
| Auto-reset mid-stream confuses frontend | Medium | Synthetic event sent before stream closes; frontend handles gracefully |

## Rollback Plan

`session_tracker.go` is a new file — delete it. Revert the four modified Go files and two frontend files. No schema migration. Token state files in `workspace/bridge/` can be left or deleted safely.

## Dependencies

- No new Go modules required
- `MaxSessionTokens` already in `DevStudioConfig` (default 200000)

## Success Criteria

- [ ] Tokens accumulate correctly from consecutive `result` events
- [ ] Session auto-resets when limit exceeded; frontend shows notification
- [ ] Token state survives a bridge restart (persisted to file)
- [ ] `GET /api/v1/dev/session/stats` returns correct counts
- [ ] Frontend badge updates in real time
- [ ] `go test -race ./pkg/web/...` passes
- [ ] Change is non-breaking for single-user and multi-user deployments
