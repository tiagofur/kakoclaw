# Token Tracking & Auto-Reset Specification

## Purpose

Defines how Dev Studio accumulates token usage per user+session from bridge result events,
enforces the configured limit via automatic session reset, and persists state across restarts.

---

## Requirements

### Requirement: Token Accumulation

The system MUST accumulate `tokens_used`, `cost_usd`, and `num_turns` from every `result`
bridge event, keyed by `userUUID + sessionID`. Both the HTTP and WebSocket query paths
MUST record to the same tracker.

#### Scenario: First result event recorded

- GIVEN no prior token state exists for `userUUID + sessionID`
- WHEN a `result` event arrives with `tokens_used=500, cost_usd=0.01`
- THEN the tracker stores `{tokens: 500, cost: 0.01}` for that key

#### Scenario: Consecutive result events accumulate

- GIVEN the tracker holds `{tokens: 500}` for a session
- WHEN a second `result` event arrives with `tokens_used=300`
- THEN the tracker stores `{tokens: 800}` for that session

#### Scenario: Different sessions are isolated

- GIVEN two sessions `S1` and `S2` for the same user
- WHEN each receives a `result` event
- THEN each session's counters remain independent

---

### Requirement: Auto-Reset on Limit Exceeded

The system MUST trigger an automatic session reset when accumulated tokens for a
`userUUID + sessionID` exceed `DevStudioConfig.MaxSessionTokens`. The reset MUST:
1. Inject a synthetic `session_reset` event into the current stream before closing it
2. Generate a new `sessionID`
3. Call `bridge.Reset()` to clear bridge state
4. Zero the token counters for the new session

#### Scenario: Limit exceeded triggers reset

- GIVEN `MaxSessionTokens=1000` and a session at `950 tokens`
- WHEN a `result` event arrives with `tokens_used=100`
- THEN a `session_reset` event is injected into the stream
- AND a new `sessionID` is assigned
- AND the token counter resets to `0`

#### Scenario: Exactly at limit does NOT reset

- GIVEN `MaxSessionTokens=1000` and a session at `900 tokens`
- WHEN a `result` event arrives with `tokens_used=100`
- THEN NO `session_reset` event is emitted
- AND the counter is `1000`

#### Scenario: Reset during WebSocket stream

- GIVEN an active WebSocket session nearing the limit
- WHEN the limit is exceeded mid-stream
- THEN the `session_reset` event is sent before the stream closes
- AND subsequent queries use the new `sessionID`

---

### Requirement: State Persistence

The system MUST persist token state to `{workspace}/bridge/{userUUID}-session-tokens.json`
using an atomic write (write temp file → rename). File write failures MUST be logged as
warnings but MUST NOT interrupt query processing.

#### Scenario: State survives bridge restart

- GIVEN token state `{tokens: 800}` was persisted before restart
- WHEN the bridge process restarts
- THEN `SessionTracker` reloads state from the JSON file
- AND subsequent `result` events accumulate from `800`

#### Scenario: Missing file on first boot

- GIVEN no token state file exists
- WHEN `SessionTracker` initializes
- THEN it starts with an empty map (zero counters for all sessions)

#### Scenario: File write failure is non-fatal

- GIVEN the JSON file path is not writable
- WHEN a `result` event is recorded
- THEN a warning is logged
- AND the in-memory state is still updated
- AND the query response is returned normally

---

### Requirement: Concurrency Safety

The `SessionTracker` MUST be protected by a mutex so concurrent HTTP and WebSocket
handlers do not cause data races. The implementation MUST pass `go test -race ./pkg/web/...`.

#### Scenario: Concurrent requests do not race

- GIVEN two goroutines simultaneously calling `tracker.Record()` for different sessions
- WHEN both complete
- THEN each session's counter reflects exactly its own events (no data loss, no panic)
