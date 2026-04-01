# Frontend Token Display Specification

## Purpose

Defines how the Dev Studio frontend surfaces token usage in the terminal header and
reacts to auto-reset events.

---

## Requirements

### Requirement: Pinia Store Token State

`devStudioStore.js` MUST expose reactive state: `sessionTokens` (integer), `sessionCostUsd`
(float), `tokenLimit` (integer), `numTurns` (integer). These MUST update on every `result`
event received over the WebSocket stream.

#### Scenario: result event updates store

- GIVEN the store has `sessionTokens=500`
- WHEN a `result` event arrives with `tokens_used=300`
- THEN `sessionTokens` becomes `800`

#### Scenario: Store initialises from stats endpoint

- GIVEN the Dev Studio panel opens
- WHEN the frontend fetches `GET /api/v1/dev/session/stats`
- THEN `sessionTokens`, `sessionCostUsd`, and `tokenLimit` are set from the response

---

### Requirement: session_reset Event Handling

When `devStudioStore.js` receives a `session_reset` event, it MUST reset
`sessionTokens`, `sessionCostUsd`, and `numTurns` to `0`, update `sessionId` to
`new_session_id`, and emit a user-visible notification in the terminal output.

#### Scenario: Reset event clears counters and notifies

- GIVEN `sessionTokens=1050`
- WHEN a `session_reset` event arrives with `new_session_id="abc-123"`
- THEN `sessionTokens` is `0`
- AND `sessionId` is `"abc-123"`
- AND a notification line appears in the terminal output

---

### Requirement: Terminal Header Token Badge

`TerminalHeader.vue` MUST render a pill badge displaying `{sessionTokens} / {tokenLimit} tokens`.
The badge MUST update reactively. When `sessionTokens >= tokenLimit * 0.9` the badge MUST
apply a warning colour; when `sessionTokens >= tokenLimit` it MUST apply a danger colour.

#### Scenario: Badge shows live token usage

- GIVEN `sessionTokens=750` and `tokenLimit=1000`
- WHEN the component renders
- THEN the badge text is `"750 / 1000 tokens"` with default styling

#### Scenario: Badge turns warning at 90% usage

- GIVEN `sessionTokens=900` and `tokenLimit=1000`
- WHEN the component renders
- THEN the badge applies the warning colour class

#### Scenario: Badge turns danger at 100% usage

- GIVEN `sessionTokens=1000` and `tokenLimit=1000`
- WHEN the component renders
- THEN the badge applies the danger colour class
