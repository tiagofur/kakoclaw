# Design: Token Tracking + Auto-Reset (Dev Studio)

## Technical Approach

Introduce `pkg/web/session_tracker.go` — a mutex-protected, file-backed accumulator that both `handleDevQuery` (HTTP NDJSON) and `handleDevTerminalWS` (WebSocket) call after every `result` event. When accumulated tokens exceed `DevStudioConfig.MaxSessionTokens > 0`, the handler injects a synthetic `session_reset` event before closing the stream, generates a new `sessionID` via `uuid.NewString()`, and calls `b.Stop()`/`b.Start()` to reset bridge state. Frontend reads token state on panel open via `GET /api/v1/dev/session/stats` and updates reactively from streamed events.

## Architecture Decisions

| Decision | Choice | Alternatives Rejected | Rationale |
|---|---|---|---|
| Persistence | JSON file, atomic write (tmp → rename) | SQLite table | No new schema migration; consistent with existing `{userUUID}-state.json` pattern; atomic rename is OS-guaranteed |
| Tracker location | `pkg/web/` as `session_tracker.go` | `pkg/bridge/` | Bridge knows nothing about users/sessions; tracker is a web-layer concern |
| Reset mechanism | `b.Stop()` + `b.Start()` (existing public methods) | New `b.Reset()` method on Bridge | Both public methods already exist; adding a `Reset()` alias adds no value |
| Event injection | Synthesise `bridge.Event{Type: "session_reset"}` in handler before stream closes | Bridge-side reset signal | Bridge process is stateless re: token limits; server owns the limit logic |
| Limit guard | `tokens > limit` triggers reset (`== limit` does NOT) | `>= limit` | Spec scenario "Exactly at limit does NOT reset" mandates strict `>` |
| SessionID source | `sessionID` embedded in WS message from frontend | Derive from project name | HTTP path already uses `"dev_studio_"+projectName`; tracker key = `userUUID + sessionID` using same pattern |

## Data Flow

```
Browser (WS/HTTP)
      │ prompt
      ▼
handleDevTerminalWS / handleDevQuery
      │ b.Execute(ctx, req)
      ▼
bridge.Bridge ──NDJSON──► event channel
      │
      │  for ev := range ch
      ├─── ev.Type == "result"?
      │         │  YES
      │         ▼
      │    tracker.Record(userUUID, sessionID, ev.NumTurns, ev.CostUSD)
      │         │
      │         ├── ShouldReset()? (accumulated > MaxSessionTokens)
      │         │         │  YES
      │         │         ▼
      │         │   inject session_reset event → frontend
      │         │   newID = uuid.NewString()
      │         │   b.Stop() → b.Start(ctx)
      │         │   tracker.ZeroSession(userUUID, newID)
      │         │
      │         └── NO → continue streaming
      │
      ▼
writeJSONResponse / safeWrite(eventToFrontend(ev))
```

**Persistence path** (after every `Record`):
```
tracker.mu.Lock()
  state[key].Tokens += ev.NumTurns   (tokens proxy; Event carries NumTurns not raw tokens)
  state[key].CostUSD += ev.CostUSD
  save: write tmp file → os.Rename (atomic)
tracker.mu.Unlock()
```

> **Note:** `bridge.Event` carries `num_turns` and `cost_usd` but no raw `tokens_used` field. The spec uses `tokens_used` as a logical counter — implementation maps it to `NumTurns` (the only integer counter on the result event). If the bridge protocol later exposes a dedicated token field, only `tracker.Record()` needs updating.

## File Changes

| File | Action | Description |
|---|---|---|
| `pkg/web/session_tracker.go` | **Create** | `SessionTracker` struct: `Record()`, `ShouldReset()`, `ZeroSession()`, `Stats()`, file persistence |
| `pkg/bridge/events.go` | **Modify** | Add `TokensUsed int` field to `Event`; add `EventSessionReset = "session_reset"` constant |
| `pkg/web/handlers_dev.go` | **Modify** | Add `s.sessionTracker` field call in `handleDevQuery` event loop after `result` event; inject reset event; generate new sessionID |
| `pkg/web/handlers_dev_ws.go` | **Modify** | Same tracker call in WS goroutine streaming loop |
| `pkg/web/server.go` | **Modify** | Add `sessionTracker *SessionTracker` field to `Server`; init in `NewServer`; register `GET /api/v1/dev/session/stats` |
| `pkg/web/frontend/src/stores/devStudioStore.js` | **Modify** | Add `sessionTokens`, `sessionCostUsd`, `tokenLimit`, `numTurns` refs; update on `result`/`session_reset` events; fetch stats on panel open |
| `pkg/web/frontend/src/components/DevStudio/TerminalHeader.vue` | **Create** | Token pill badge; warning (≥90%) and danger (≥100%) colour classes; reactive from store |

## Interfaces / Contracts

```go
// pkg/web/session_tracker.go
type SessionStats struct {
    SessionID  string  `json:"session_id"`
    TokensUsed int     `json:"tokens_used"`   // maps to NumTurns
    CostUSD    float64 `json:"cost_usd"`
    TokenLimit int     `json:"token_limit"`
    NumTurns   int     `json:"num_turns"`
}

type SessionTracker struct {
    mu        sync.Mutex
    sessions  map[string]*SessionStats  // key = userUUID + ":" + sessionID
    stateDir  string
    maxTokens int
}

func NewSessionTracker(stateDir string, maxTokens int) *SessionTracker
func (t *SessionTracker) Record(userUUID, sessionID string, turns int, costUSD float64)
func (t *SessionTracker) ShouldReset(userUUID, sessionID string) bool
func (t *SessionTracker) ZeroSession(userUUID, newSessionID string)
func (t *SessionTracker) Stats(userUUID, sessionID string) SessionStats
```

```go
// pkg/bridge/events.go — additions
const EventSessionReset = "session_reset"

// Event struct — new field
TokensUsed int `json:"tokens_used,omitempty"`
```

**REST endpoint** — `GET /api/v1/dev/session/stats` → auth required (existing `extractClaims`) → `200 OK` with `SessionStats` JSON; zero values when no session exists.

**Synthetic event payload** sent to frontend:
```json
{ "type": "session_reset", "new_session_id": "<uuid>", "message": "Session auto-reset: token limit reached" }
```

## Testing Strategy

| Layer | What | Approach |
|---|---|---|
| Unit | `SessionTracker.Record`, `ShouldReset`, `ZeroSession`, file persistence, concurrent safety | `go test -race ./pkg/web/...`; tmp dir for state files; parallel goroutine test |
| Integration | `GET /api/v1/dev/session/stats` returns correct JSON; auth rejection returns 401 | `net/http/httptest` with a stub `Server` |
| Unit (frontend) | Store updates `sessionTokens` on `result` event; resets on `session_reset`; badge colour classes | Vitest + `@vue/test-utils` |

## Migration / Rollout

- **No DB migration.** Feature activates when `DevStudioConfig.MaxSessionTokens > 0` (existing field, default 200000 per proposal).
- When `MaxSessionTokens == 0`, `ShouldReset` always returns `false` — tracker still records stats for the `/stats` endpoint but never triggers reset.
- Token state files (`{workspace}/bridge/{userUUID}-session-tokens.json`) are created on first `result` event. Existing deployments without the file start from zero counters.
- Rollback: delete `session_tracker.go`, revert 4 Go files and 2 frontend files. JSON state files are harmless leftovers.

## Open Questions

- None blocking. Design is complete based on existing codebase patterns.
