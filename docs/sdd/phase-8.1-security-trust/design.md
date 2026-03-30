# Design: Phase 8.1 — Security & Trust

## Technical Approach

Three independent security layers wired into existing extension points: (1) `pkg/hooks` package with a `HookRegistry` called from `loop.go` at `ExecuteWithContext`; (2) `DMPolicy` field per-channel config evaluated in `BaseChannel.HandleMessage` before bus dispatch; (3) `ToolProfiles` added to `ToolPermissionsConfig`, applied after role filtering in `filterToolsByPermissions`. All storage is additive (new table, no schema changes to existing tables).

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|--------------|-----------|
| Hook package location | New `pkg/hooks/` | Inline in `pkg/agent/` | Keeps hooks testable independently; `loop.go` already large |
| Hook invocation point | Before `ExecuteWithContext` at line 1600 | After; middleware wrap | Pre-call gives `block` semantics; aligns with existing `OnTool` pattern |
| `require_approval` fallback | Send `OutboundMessage` to channel owner | Separate approval queue | Reuses existing bus; no new infrastructure |
| Pairing store backend | New `pairing_store` SQLite table in per-user `storage.Storage` | Central DB | Pairing is per-user; `BaseChannel` already holds `*storage.Storage` |
| `DMPolicy` evaluation | In `BaseChannel.HandleMessage` before bus publish | In channel-specific adapters | Single enforcement point; all channels inherit |
| Tool profile application | After role filtering in `filterToolsByPermissions` | Before; separate function | Profiles narrow the already-role-filtered set; additive to existing logic |

## Data Flow

### Hook Pipeline (before_tool_call)

```
loop.go: ExecuteWithContext called
    │
    ├─ al.hooks.Run("before_tool_call", HookContext{Tool, Args})
    │       │
    │       ├─ HookResult.Action == allow  → proceed
    │       ├─ HookResult.Action == block  → return error to LLM
    │       └─ HookResult.Action == require_approval
    │               │
    │               └─ bus.PublishOutbound(approval request to channel owner)
    │                  → return blocked result to LLM (pending approval)
    │
    └─ al.tools.ExecuteWithContext(...)   ← existing path unchanged
```

### DM Pairing Flow

```
Channel.HandleMessage(senderID, ...)
    │
    ├─ IsAllowed(senderID)?  YES → existing path unchanged
    │
    └─ NO: check dmPolicy
            ├─ "open"      → pass through
            ├─ "allowlist" → block (existing behavior)
            ├─ "disabled"  → block silently
            └─ "pairing"
                    │
                    ├─ pairing_store.GetApproved(channel, senderID)? → pass through
                    ├─ pairing_store.GetPending(channel, senderID)?  → resend challenge
                    └─ new sender:
                            ├─ generate code
                            ├─ pairing_store.InsertPending(...)
                            └─ bus.PublishOutbound(challenge message to sender)
```

### /approve Command Flow

```
CommandHandler.Handle("/approve <channel> <code>")
    │
    └─ pairing_store.Approve(channel, code)
            └─ move row: pending=true → pending=false (approved)
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/hooks/hooks.go` | Create | `HookRegistry`, `HookHandler` interface, `HookResult`, built-in runner with panic recovery |
| `pkg/storage/pairing.go` | Create | `PairingStore` CRUD: insert pending, approve by code, check approved, expire old entries |
| `pkg/config/config.go` | Modify | Add `DMPolicy string` to each channel config struct; add `ToolProfiles map[string]ToolProfileConfig` + `ActiveProfile string` to `ToolPermissionsConfig` |
| `pkg/channels/base.go` | Modify | Add `dmPolicy string` field; extend `HandleMessage` with policy switch; add `SetDMPolicy`, `SetPairingStore` methods |
| `pkg/agent/loop.go` | Modify | Add `hooks *hooks.HookRegistry` field; call `hooks.Run("before_tool_call", ...)` before `ExecuteWithContext`; call `hooks.Run("message_sending", ...)` before `PublishOutbound` |
| `pkg/agent/permissions.go` | Modify | Apply active `ToolProfileConfig` (deny list + allow list) after role filtering |
| `pkg/channels/command_handler.go` | Modify | Register `/approve <channel> <code>` handler |
| `pkg/storage/sqlite.go` | Modify | Add `pairing_store` table in migration |

## Interfaces / Contracts

```go
// pkg/hooks/hooks.go

type HookAction string
const (
    HookAllow           HookAction = "allow"
    HookBlock           HookAction = "block"
    HookRequireApproval HookAction = "require_approval"
)

type HookContext struct {
    Event     string                 // "before_tool_call" | "before_install" | "message_sending"
    ToolName  string                 // populated for before_tool_call
    Args      map[string]interface{} // populated for before_tool_call
    Message   string                 // populated for message_sending
    UserID    int64
}

type HookResult struct {
    Action  HookAction
    Reason  string
}

type HookHandler interface {
    Priority() int
    Handle(ctx context.Context, hc HookContext) HookResult
}

type HookRegistry struct {
    handlers []HookHandler // sorted by Priority() ascending
}

func (r *HookRegistry) Register(h HookHandler)
func (r *HookRegistry) Run(ctx context.Context, event string, hc HookContext) HookResult
```

```go
// pkg/config/config.go additions

type ToolProfileConfig struct {
    Allow        []string `json:"allow"`         // tool names; empty = unrestricted
    Deny         []string `json:"deny"`           // tool names always removed
    ExecSecurity string   `json:"exec_security"`  // "deny" | "ask" | "full"
}

// Added to ToolPermissionsConfig:
ToolProfiles   map[string]ToolProfileConfig `json:"tool_profiles,omitempty"`
ActiveProfile  string                       `json:"active_profile,omitempty"`
```

```sql
-- pairing_store table (additive migration)
CREATE TABLE IF NOT EXISTS pairing_store (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    channel     TEXT    NOT NULL,
    sender_id   TEXT    NOT NULL,
    code        TEXT    NOT NULL,
    pending     BOOLEAN NOT NULL DEFAULT 1,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    approved_at DATETIME,
    UNIQUE(channel, sender_id)
);
CREATE INDEX IF NOT EXISTS idx_pairing_code ON pairing_store(code);
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `HookRegistry.Run`: priority order, panic recovery, block/allow/require_approval results | Table-driven tests in `pkg/hooks/hooks_test.go` |
| Unit | `PairingStore`: insert, approve, get-approved, expiry | `pkg/storage/pairing_test.go` with in-memory SQLite |
| Unit | `filterToolsByPermissions` with active profile | Extend `pkg/agent/permissions_test.go` |
| Unit | `BaseChannel.HandleMessage` with each dmPolicy value | `pkg/channels/base_test.go` with mock bus |
| Integration | `/approve` command writes to pairing store, subsequent message passes | `pkg/channels/command_handler_test.go` |

## Migration / Rollout

- `pairing_store` table: `CREATE TABLE IF NOT EXISTS` — additive, no data loss.
- Existing `allow_from` configs continue working: `BaseChannel` defaults `dmPolicy` to `"allowlist"` when the field is absent, preserving current static behavior.
- `ToolProfiles` is an optional map. When `ActiveProfile` is empty, `filterToolsByPermissions` behavior is unchanged.
- Built-in profiles (`messaging`, `developer`, `minimal`) are registered as defaults in `NewDefaults()`.

## Open Questions

- [ ] Should `require_approval` requests be deduplicated per sender per session (to avoid flooding the owner channel)? Proposal mentions rate-limiting — confirm if this goes in this phase or 8.2.
- [ ] `/approve` command: should it be owner-only (require sender to be in global `allow_from`) or any admin-role user?
