# Tasks: Phase 8.1 — Security & Trust

**Change**: `phase-8.1-security-trust`
**Date**: 2026-03-30
**TDD Mode**: RED → GREEN → COMMIT per task group

---

## Phase 1: Foundation — Config Types & Storage Schema

### 1.1 Add `DMPolicy` field to each channel config struct

**File**: `pkg/config/config.go`

For each of the 9 channel config structs (`TelegramConfig`, `DiscordConfig`, `SlackConfig`, `WhatsAppConfig`, `FeishuConfig`, `QQConfig`, `DingTalkConfig`, `SignalConfig`, `MaixCamConfig`), add:

```go
DMPolicy string `json:"dm_policy,omitempty" env:"MAKOCLAW_CHANNELS_<NAME>_DM_POLICY"`
```

Example for `TelegramConfig`:
```go
type TelegramConfig struct {
    Enabled   bool                `json:"enabled" env:"MAKOCLAW_CHANNELS_TELEGRAM_ENABLED"`
    Token     string              `json:"token" env:"MAKOCLAW_CHANNELS_TELEGRAM_TOKEN"`
    Proxy     string              `json:"proxy" env:"MAKOCLAW_CHANNELS_TELEGRAM_PROXY"`
    AllowFrom FlexibleStringSlice `json:"allow_from" env:"MAKOCLAW_CHANNELS_TELEGRAM_ALLOW_FROM"`
    DMPolicy  string              `json:"dm_policy,omitempty" env:"MAKOCLAW_CHANNELS_TELEGRAM_DM_POLICY"`
}
```

Apply to all 9 channel config structs. Do NOT add to `EmailConfig` (server push, not DM).

**Verify**: `go vet ./pkg/config/...` passes with no errors.

---

### 1.2 Add `ToolProfileConfig` and profile fields to `ToolPermissionsConfig`

**File**: `pkg/config/config.go`

Add the new type after `ToolPermissionsConfig`:

```go
// ToolProfileConfig defines a named tool-access preset.
type ToolProfileConfig struct {
    Allow        []string `json:"allow"`        // tool names; empty = unrestricted
    Deny         []string `json:"deny"`          // tool names always removed
    ExecSecurity string   `json:"exec_security"` // "deny" | "ask" | "full"
}
```

Extend `ToolPermissionsConfig`:

```go
type ToolPermissionsConfig struct {
    RoleDefaults  map[string][]string            `json:"role_defaults"`
    ToolProfiles  map[string]ToolProfileConfig   `json:"tool_profiles,omitempty"`
    ActiveProfile string                         `json:"active_profile,omitempty"`
}
```

**Verify**: `go vet ./pkg/config/...` passes.

---

### 1.3 Register built-in tool profiles in `NewDefaults()`

**File**: `pkg/config/config.go` — inside the `NewDefaults()` function, within `ToolPermissions`:

```go
ToolPermissions: ToolPermissionsConfig{
    RoleDefaults: map[string][]string{ /* existing content */ },
    ToolProfiles: map[string]ToolProfileConfig{
        "messaging": {
            Deny:         []string{"exec", "write_file"},
            ExecSecurity: "deny",
        },
        "developer": {
            Allow:        []string{},
            ExecSecurity: "full",
        },
        "minimal": {
            Allow:        []string{"message", "query_knowledge"},
            ExecSecurity: "deny",
        },
    },
    ActiveProfile: "",
},
```

**Verify**: `go test ./pkg/config/... -run TestNewDefaults -v` — test exists and passes (or write a quick one in 1.4).

---

### 1.4 Write failing test for built-in profiles

**File**: `pkg/config/config_test.go`

```go
func TestNewDefaults_BuiltInToolProfiles(t *testing.T) {
    cfg := config.NewDefaults()
    require.Contains(t, cfg.ToolPermissions.ToolProfiles, "messaging")
    require.Contains(t, cfg.ToolPermissions.ToolProfiles, "developer")
    require.Contains(t, cfg.ToolPermissions.ToolProfiles, "minimal")

    messaging := cfg.ToolPermissions.ToolProfiles["messaging"]
    require.Contains(t, messaging.Deny, "exec")
    require.Contains(t, messaging.Deny, "write_file")

    minimal := cfg.ToolPermissions.ToolProfiles["minimal"]
    require.Equal(t, []string{"message", "query_knowledge"}, minimal.Allow)
}
```

Run: `go test ./pkg/config/... -run TestNewDefaults_BuiltInToolProfiles -v`
Expected: **PASS** (after 1.3 is done).

Commit: `feat(config): add DMPolicy and ToolProfileConfig types with built-in profiles`

---

### 1.5 Add `pairing_store` migration to `migrateUserDB()`

**File**: `pkg/storage/sqlite.go` — append inside `migrateUserDB()` queries slice:

```go
`CREATE TABLE IF NOT EXISTS pairing_store (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    channel     TEXT    NOT NULL,
    sender_id   TEXT    NOT NULL,
    code        TEXT    NOT NULL,
    pending     BOOLEAN NOT NULL DEFAULT 1,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    approved_at DATETIME,
    UNIQUE(channel, sender_id)
);`,
`CREATE INDEX IF NOT EXISTS idx_pairing_code ON pairing_store(code);`,
```

**Verify**: `go vet ./pkg/storage/...` passes.

---

### 1.6 Create `PairingStore` with full CRUD

**File**: `pkg/storage/pairing.go` (new file)

```go
package storage

import (
    "crypto/rand"
    "database/sql"
    "encoding/hex"
    "errors"
    "fmt"
    "time"
)

// ErrCodeNotFound is returned when an approve code has no matching pending entry.
var ErrCodeNotFound = errors.New("pairing: code not found or already approved")

// PairingStore provides challenge-based allowlist CRUD backed by the user SQLite DB.
type PairingStore struct {
    db *sql.DB
}

// NewPairingStore wraps an existing *sql.DB (from Storage).
func NewPairingStore(db *sql.DB) *PairingStore {
    return &PairingStore{db: db}
}

// GenerateCode returns a 6-character hex challenge code.
func GenerateCode() (string, error) {
    b := make([]byte, 3)
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("generating pairing code: %w", err)
    }
    return hex.EncodeToString(b), nil
}

// InsertPending upserts a pending challenge for (channel, senderID).
// If a pending row already exists it is a no-op (deduplication).
func (ps *PairingStore) InsertPending(channel, senderID, code string) error {
    _, err := ps.db.Exec(`
        INSERT INTO pairing_store (channel, sender_id, code, pending, created_at)
        VALUES (?, ?, ?, 1, ?)
        ON CONFLICT(channel, sender_id) DO UPDATE SET
            code       = CASE WHEN pending = 1 THEN excluded.code ELSE code END,
            created_at = CASE WHEN pending = 1 THEN excluded.created_at ELSE created_at END
        WHERE pending = 1
    `, channel, senderID, code, time.Now().UTC())
    return err
}

// HasPending returns true when a pending (not yet approved) entry exists.
func (ps *PairingStore) HasPending(channel, senderID string) (bool, error) {
    var count int
    err := ps.db.QueryRow(`
        SELECT COUNT(*) FROM pairing_store
        WHERE channel = ? AND sender_id = ? AND pending = 1
    `, channel, senderID).Scan(&count)
    return count > 0, err
}

// IsApproved returns true when the sender has an approved entry.
func (ps *PairingStore) IsApproved(channel, senderID string) (bool, error) {
    var count int
    err := ps.db.QueryRow(`
        SELECT COUNT(*) FROM pairing_store
        WHERE channel = ? AND sender_id = ? AND pending = 0
    `, channel, senderID).Scan(&count)
    return count > 0, err
}

// Approve finds a pending row by code and marks it approved.
// Returns ErrCodeNotFound if the code is unknown or already used.
func (ps *PairingStore) Approve(channel, code string) (senderID string, err error) {
    tx, err := ps.db.Begin()
    if err != nil {
        return "", err
    }
    defer func() {
        if err != nil {
            _ = tx.Rollback()
        }
    }()

    err = tx.QueryRow(`
        SELECT sender_id FROM pairing_store
        WHERE channel = ? AND code = ? AND pending = 1
    `, channel, code).Scan(&senderID)
    if errors.Is(err, sql.ErrNoRows) {
        return "", ErrCodeNotFound
    }
    if err != nil {
        return "", err
    }

    _, err = tx.Exec(`
        UPDATE pairing_store SET pending = 0, approved_at = ?
        WHERE channel = ? AND code = ? AND pending = 1
    `, time.Now().UTC(), channel, code)
    if err != nil {
        return "", err
    }

    return senderID, tx.Commit()
}
```

**Verify**: `go build ./pkg/storage/...` compiles cleanly.

---

### 1.7 Write PairingStore unit tests (RED → GREEN)

**File**: `pkg/storage/pairing_test.go` (new file)

```go
package storage_test

import (
    "path/filepath"
    "testing"

    "github.com/sipeed/makoclaw/pkg/storage"
    "github.com/stretchr/testify/require"
)

func newTestStorage(t *testing.T) *storage.Storage {
    t.Helper()
    path := filepath.Join(t.TempDir(), "test.db")
    s, err := storage.NewUserStorage(path)
    require.NoError(t, err)
    t.Cleanup(func() { _ = s.Close() })
    return s
}

func TestPairingStore_InsertAndApprove(t *testing.T) {
    s := newTestStorage(t)
    ps := s.Pairing()

    ok, err := ps.IsApproved("telegram", "u1")
    require.NoError(t, err)
    require.False(t, ok)

    require.NoError(t, ps.InsertPending("telegram", "u1", "abc123"))

    pending, err := ps.HasPending("telegram", "u1")
    require.NoError(t, err)
    require.True(t, pending)

    senderID, err := ps.Approve("telegram", "abc123")
    require.NoError(t, err)
    require.Equal(t, "u1", senderID)

    approved, err := ps.IsApproved("telegram", "u1")
    require.NoError(t, err)
    require.True(t, approved)
}

func TestPairingStore_DeduplicatePending(t *testing.T) {
    s := newTestStorage(t)
    ps := s.Pairing()

    require.NoError(t, ps.InsertPending("telegram", "u2", "code1"))
    // Second insert for same sender must not error and must not overwrite
    require.NoError(t, ps.InsertPending("telegram", "u2", "code2"))

    // Original code still works
    _, err := ps.Approve("telegram", "code1")
    require.NoError(t, err)
}

func TestPairingStore_UnknownCodeReturnsError(t *testing.T) {
    s := newTestStorage(t)
    ps := s.Pairing()

    _, err := ps.Approve("telegram", "nope")
    require.ErrorIs(t, err, storage.ErrCodeNotFound)
}
```

This test requires `s.Pairing()` — add that accessor in task 1.8.

Run: `go test ./pkg/storage/... -run TestPairingStore -v` → **RED** (method missing).

---

### 1.8 Expose `Pairing()` accessor on `Storage`

**File**: `pkg/storage/sqlite.go` — add method after `Close()`:

```go
// Pairing returns a PairingStore backed by this storage's DB.
func (s *Storage) Pairing() *PairingStore {
    return NewPairingStore(s.db)
}
```

Run: `go test ./pkg/storage/... -run TestPairingStore -v` → **GREEN**.

Commit: `feat(storage): add pairing_store table and PairingStore CRUD`

---

## Phase 2: Hooks Package

### 2.1 Write failing hook registry tests

**File**: `pkg/hooks/hooks_test.go` (new file — `pkg/hooks/` dir is new)

```go
package hooks_test

import (
    "context"
    "testing"

    "github.com/sipeed/makoclaw/pkg/hooks"
    "github.com/stretchr/testify/require"
)

type mockHandler struct {
    priority int
    action   hooks.HookAction
    reason   string
    called   *bool
}

func (m *mockHandler) Priority() int { return m.priority }
func (m *mockHandler) Handle(_ context.Context, _ hooks.HookContext) hooks.HookResult {
    *m.called = true
    return hooks.HookResult{Action: m.action, Reason: m.reason}
}

func TestHookRegistry_AllowPassthrough(t *testing.T) {
    r := hooks.NewHookRegistry()
    called := false
    r.Register(&mockHandler{priority: 10, action: hooks.HookAllow, called: &called})
    result := r.Run(context.Background(), "before_tool_call", hooks.HookContext{ToolName: "exec"})
    require.Equal(t, hooks.HookAllow, result.Action)
    require.True(t, called)
}

func TestHookRegistry_BlockShortCircuits(t *testing.T) {
    r := hooks.NewHookRegistry()
    secondCalled := false
    r.Register(&mockHandler{priority: 10, action: hooks.HookBlock, reason: "denied", called: new(bool)})
    r.Register(&mockHandler{priority: 20, action: hooks.HookAllow, called: &secondCalled})
    result := r.Run(context.Background(), "before_tool_call", hooks.HookContext{ToolName: "exec"})
    require.Equal(t, hooks.HookBlock, result.Action)
    require.Equal(t, "denied", result.Reason)
    require.False(t, secondCalled, "second handler must not run after block")
}

func TestHookRegistry_PriorityOrder(t *testing.T) {
    r := hooks.NewHookRegistry()
    order := []int{}
    r.Register(&struct {
        hooks.HookHandler
    }{&mockOrderHandler{p: 30, order: &order}})
    r.Register(&struct{ hooks.HookHandler }{&mockOrderHandler{p: 10, order: &order}})
    r.Register(&struct{ hooks.HookHandler }{&mockOrderHandler{p: 20, order: &order}})
    r.Run(context.Background(), "before_tool_call", hooks.HookContext{})
    require.Equal(t, []int{10, 20, 30}, order)
}

type mockOrderHandler struct {
    p     int
    order *[]int
}

func (m *mockOrderHandler) Priority() int { return m.p }
func (m *mockOrderHandler) Handle(_ context.Context, _ hooks.HookContext) hooks.HookResult {
    *m.order = append(*m.order, m.p)
    return hooks.HookResult{Action: hooks.HookAllow}
}

func TestHookRegistry_PanicRecovered(t *testing.T) {
    r := hooks.NewHookRegistry()
    r.Register(&panicHandler{})
    result := r.Run(context.Background(), "before_tool_call", hooks.HookContext{ToolName: "exec"})
    // After panic recovery, default is allow
    require.Equal(t, hooks.HookAllow, result.Action)
}

type panicHandler struct{}

func (p *panicHandler) Priority() int { return 10 }
func (p *panicHandler) Handle(_ context.Context, _ hooks.HookContext) hooks.HookResult {
    panic("simulated hook panic")
}

func TestHookRegistry_EmptyRegistryAllows(t *testing.T) {
    r := hooks.NewHookRegistry()
    result := r.Run(context.Background(), "before_tool_call", hooks.HookContext{})
    require.Equal(t, hooks.HookAllow, result.Action)
}
```

Run: `go test ./pkg/hooks/... -v` → **RED** (package doesn't exist).

---

### 2.2 Create `pkg/hooks/hooks.go`

**File**: `pkg/hooks/hooks.go` (new file)

```go
package hooks

import (
    "context"
    "sort"

    "github.com/sipeed/makoclaw/pkg/logger"
)

// HookAction is the outcome of a HookHandler invocation.
type HookAction string

const (
    HookAllow           HookAction = "allow"
    HookBlock           HookAction = "block"
    HookRequireApproval HookAction = "require_approval"
)

// HookContext carries event metadata to each handler.
type HookContext struct {
    Event    string                 // "before_tool_call" | "before_install" | "message_sending"
    ToolName string                 // populated for before_tool_call
    Args     map[string]interface{} // populated for before_tool_call
    Message  string                 // populated for message_sending
    UserID   int64
}

// HookResult is the outcome returned by a handler.
type HookResult struct {
    Action HookAction
    Reason string
}

// HookHandler is the interface each security hook must implement.
type HookHandler interface {
    Priority() int
    Handle(ctx context.Context, hc HookContext) HookResult
}

// HookRegistry holds ordered HookHandler entries.
type HookRegistry struct {
    handlers []HookHandler
}

// NewHookRegistry creates an empty registry.
func NewHookRegistry() *HookRegistry {
    return &HookRegistry{}
}

// Register adds a handler and re-sorts by ascending priority.
func (r *HookRegistry) Register(h HookHandler) {
    r.handlers = append(r.handlers, h)
    sort.Slice(r.handlers, func(i, j int) bool {
        return r.handlers[i].Priority() < r.handlers[j].Priority()
    })
}

// Run executes all handlers for the given event in priority order.
// Stops on the first block or require_approval result.
// Panics inside handlers are recovered and logged; execution continues with allow.
func (r *HookRegistry) Run(ctx context.Context, event string, hc HookContext) (result HookResult) {
    result = HookResult{Action: HookAllow}
    hc.Event = event

    for _, h := range r.handlers {
        func() {
            defer func() {
                if rec := recover(); rec != nil {
                    logger.WarnCF("hooks", "handler panicked, defaulting to allow", map[string]interface{}{
                        "event":   event,
                        "recover": rec,
                    })
                }
            }()
            result = h.Handle(ctx, hc)
        }()

        if result.Action != HookAllow {
            return result
        }
    }
    return result
}
```

Run: `go test ./pkg/hooks/... -v` → **GREEN**.

Commit: `feat(hooks): add HookRegistry with priority ordering and panic recovery`

---

## Phase 3: BaseChannel DM Policy

### 3.1 Write failing BaseChannel DM policy tests

**File**: `pkg/channels/base_test.go` — add new test cases (file already exists):

```go
func TestBaseChannel_DMPolicy_Open(t *testing.T) {
    b := channels.NewBaseChannel("test", nil, nil, nil)
    b.SetDMPolicy("open", nil)
    // open passes even unknown senders
    require.True(t, b.ShouldDispatch("unknown-sender"))
}

func TestBaseChannel_DMPolicy_Disabled(t *testing.T) {
    b := channels.NewBaseChannel("test", nil, nil, nil)
    b.SetDMPolicy("disabled", nil)
    require.False(t, b.ShouldDispatch("unknown-sender"))
}

func TestBaseChannel_DMPolicy_Allowlist(t *testing.T) {
    b := channels.NewBaseChannel("test", nil, nil, []string{"known"})
    b.SetDMPolicy("allowlist", nil)
    require.False(t, b.ShouldDispatch("unknown"))
    require.True(t, b.ShouldDispatch("known"))
}
```

Run: `go test ./pkg/channels/... -run TestBaseChannel_DMPolicy -v` → **RED**.

---

### 3.2 Add `dmPolicy` field and `SetDMPolicy` / `ShouldDispatch` to `BaseChannel`

**File**: `pkg/channels/base.go`

Add fields to `BaseChannel` struct:
```go
dmPolicy    string
pairingStore *storage.PairingStore
```

Add methods:

```go
// SetDMPolicy configures the DM policy and optional pairing store.
func (c *BaseChannel) SetDMPolicy(policy string, ps *storage.PairingStore) {
    c.dmPolicy = policy
    c.pairingStore = ps
}

// ShouldDispatch returns true if the sender is allowed to have their message processed.
// It encapsulates the dm_policy logic independently of the full HandleMessage flow.
func (c *BaseChannel) ShouldDispatch(senderID string) bool {
    // Static allowlist always passes
    if c.IsAllowed(senderID) {
        return true
    }
    switch c.dmPolicy {
    case "open":
        return true
    case "disabled":
        return false
    case "allowlist":
        return false // already checked above
    case "pairing":
        if c.pairingStore == nil {
            return false
        }
        approved, err := c.pairingStore.IsApproved(c.name, senderID)
        if err != nil {
            logger.WarnCF("channel", "pairing store error", map[string]interface{}{"err": err})
            return false
        }
        return approved
    default:
        // No policy set — preserve legacy allowlist behavior
        return false
    }
}
```

Run: `go test ./pkg/channels/... -run TestBaseChannel_DMPolicy -v` → **GREEN**.

---

### 3.3 Wire DM policy into `HandleMessage`

**File**: `pkg/channels/base.go` — modify `HandleMessage` at the top of the function, replacing the `!c.IsAllowed` check:

Before (existing):
```go
func (c *BaseChannel) HandleMessage(senderID, chatID, content string, media []string, metadata map[string]string) error {
    if !c.IsAllowed(senderID) {
        return nil
    }
```

After:
```go
func (c *BaseChannel) HandleMessage(senderID, chatID, content string, media []string, metadata map[string]string) error {
    if c.dmPolicy != "" {
        if !c.ShouldDispatch(senderID) {
            if c.dmPolicy == "pairing" {
                c.issuePairingChallenge(senderID, chatID)
            }
            return nil
        }
    } else if !c.IsAllowed(senderID) {
        return nil
    }
```

---

### 3.4 Implement `issuePairingChallenge` on `BaseChannel`

**File**: `pkg/channels/base.go`

```go
// issuePairingChallenge sends or resends a challenge code to an unknown sender.
func (c *BaseChannel) issuePairingChallenge(senderID, chatID string) {
    if c.pairingStore == nil || c.bus == nil {
        return
    }
    pending, err := c.pairingStore.HasPending(c.name, senderID)
    if err != nil {
        logger.WarnCF("channel", "pairing store HasPending error", map[string]interface{}{"err": err})
        return
    }
    if pending {
        // Resend existing challenge; don't generate new code
        return
    }
    code, err := storage.GenerateCode()
    if err != nil {
        logger.WarnCF("channel", "failed to generate pairing code", map[string]interface{}{"err": err})
        return
    }
    if err := c.pairingStore.InsertPending(c.name, senderID, code); err != nil {
        logger.WarnCF("channel", "failed to insert pending pairing", map[string]interface{}{"err": err})
        return
    }
    challenge := fmt.Sprintf("👋 This agent requires pairing. Ask the owner to run: /approve %s %s", c.name, code)
    c.bus.PublishOutbound(bus.OutboundMessage{
        Channel: c.name,
        ChatID:  chatID,
        Content: challenge,
    })
}
```

Run: `go test ./pkg/channels/... -v` → all channel tests pass.

Commit: `feat(channels): implement DMPolicy and challenge-based pairing flow in BaseChannel`

---

## Phase 4: /approve Command

### 4.1 Write failing /approve command test

**File**: `pkg/channels/command_handler_test.go` — add test (file may already exist):

```go
func TestCommandHandler_Approve_ValidCode(t *testing.T) {
    path := filepath.Join(t.TempDir(), "test.db")
    s, err := storage.NewUserStorage(path)
    require.NoError(t, err)
    ps := s.Pairing()
    require.NoError(t, ps.InsertPending("telegram", "user1", "abc123"))

    ch := channels.NewCommandHandler(s)
    ch.SetPairingStore(ps)

    handled, resp, err := ch.HandleCommand(context.Background(), "telegram", "owner", "/approve telegram abc123")
    require.NoError(t, err)
    require.True(t, handled)
    require.Contains(t, resp, "approved")

    approved, err := ps.IsApproved("telegram", "user1")
    require.NoError(t, err)
    require.True(t, approved)
}

func TestCommandHandler_Approve_UnknownCode(t *testing.T) {
    path := filepath.Join(t.TempDir(), "test.db")
    s, err := storage.NewUserStorage(path)
    require.NoError(t, err)
    ps := s.Pairing()
    ch := channels.NewCommandHandler(s)
    ch.SetPairingStore(ps)

    handled, resp, err := ch.HandleCommand(context.Background(), "telegram", "owner", "/approve telegram nope")
    require.NoError(t, err)
    require.True(t, handled)
    require.Contains(t, resp, "not found")
}
```

Run: `go test ./pkg/channels/... -run TestCommandHandler_Approve -v` → **RED**.

---

### 4.2 Add `PairingStore` to `CommandHandler` and implement `/approve`

**File**: `pkg/channels/command_handler.go`

Add field to struct:
```go
type CommandHandler struct {
    store        *storage.Storage
    cancelFunc   func(channel, senderID string) int
    pairingStore *storage.PairingStore
}
```

Add setter:
```go
func (ch *CommandHandler) SetPairingStore(ps *storage.PairingStore) {
    ch.pairingStore = ps
}
```

Add case in `HandleCommand` switch:
```go
case "approve":
    return ch.handleApproveCommand(ctx, channel, senderID, args)
```

Add handler:
```go
func (ch *CommandHandler) handleApproveCommand(_ context.Context, _, _, args string) (bool, string, error) {
    if ch.pairingStore == nil {
        return true, "Pairing store not configured.", nil
    }
    parts := strings.Fields(args)
    if len(parts) < 2 {
        return true, "Usage: /approve <channel> <code>", nil
    }
    targetChannel, code := parts[0], parts[1]
    senderID, err := ch.pairingStore.Approve(targetChannel, code)
    if errors.Is(err, storage.ErrCodeNotFound) {
        return true, fmt.Sprintf("Code %q not found or already used.", code), nil
    }
    if err != nil {
        return true, fmt.Sprintf("Approve failed: %v", err), nil
    }
    return true, fmt.Sprintf("✅ Sender %q approved on channel %q.", senderID, targetChannel), nil
}
```

Run: `go test ./pkg/channels/... -run TestCommandHandler_Approve -v` → **GREEN**.

Commit: `feat(channels): add /approve command for pairing allowlist management`

---

## Phase 5: Tool Profiles in permissions.go

### 5.1 Write failing tool profile tests

**File**: `pkg/agent/permissions_test.go` — add:

```go
func TestFilterToolsByPermissions_MessagingProfileDeniesExec(t *testing.T) {
    cfg := config.NewDefaults()
    cfg.ToolPermissions.ActiveProfile = "messaging"

    base := tools.NewToolRegistry()
    base.Register(&fakeNamedTool{name: "exec"})
    base.Register(&fakeNamedTool{name: "write_file"})
    base.Register(&fakeNamedTool{name: "message"})

    filtered := filterToolsByPermissions(base, "admin", 0, cfg, nil)
    names := toolNames(filtered)
    require.NotContains(t, names, "exec")
    require.NotContains(t, names, "write_file")
    require.Contains(t, names, "message")
}

func TestFilterToolsByPermissions_MinimalProfileAllowsOnlyDeclared(t *testing.T) {
    cfg := config.NewDefaults()
    cfg.ToolPermissions.ActiveProfile = "minimal"

    base := tools.NewToolRegistry()
    base.Register(&fakeNamedTool{name: "exec"})
    base.Register(&fakeNamedTool{name: "message"})
    base.Register(&fakeNamedTool{name: "query_knowledge"})
    base.Register(&fakeNamedTool{name: "read_file"})

    filtered := filterToolsByPermissions(base, "admin", 0, cfg, nil)
    names := toolNames(filtered)
    require.ElementsMatch(t, []string{"message", "query_knowledge"}, names)
}

func TestFilterToolsByPermissions_UnknownProfileErrors(t *testing.T) {
    cfg := config.NewDefaults()
    cfg.ToolPermissions.ActiveProfile = "nonexistent"

    base := tools.NewToolRegistry()
    base.Register(&fakeNamedTool{name: "exec"})

    // Unknown profile must return empty registry
    filtered := filterToolsByPermissions(base, "admin", 0, cfg, nil)
    require.Equal(t, 0, len(toolNames(filtered)))
}

func TestFilterToolsByPermissions_NoProfilePreservesRoleDefaults(t *testing.T) {
    cfg := config.NewDefaults()
    cfg.ToolPermissions.ActiveProfile = "" // no profile

    base := tools.NewToolRegistry()
    base.Register(&fakeNamedTool{name: "exec"})
    base.Register(&fakeNamedTool{name: "message"})

    filtered := filterToolsByPermissions(base, "admin", 0, cfg, nil)
    names := toolNames(filtered)
    // admin has wildcard, should get all tools
    require.Contains(t, names, "exec")
    require.Contains(t, names, "message")
}
```

Add helper if not already present:
```go
func toolNames(r *tools.ToolRegistry) []string {
    var names []string
    for _, t := range r.All() {
        names = append(names, t.Name())
    }
    return names
}
```

Run: `go test ./pkg/agent/... -run TestFilterToolsByPermissions -v` → **RED** (profile logic missing).

---

### 5.2 Apply active `ToolProfileConfig` in `filterToolsByPermissions`

**File**: `pkg/agent/permissions.go` — add at the END of the function, before returning `filteredRegistry`:

```go
// Apply active tool profile (narrows the already role-filtered set).
if profile := cfg.ToolPermissions.ActiveProfile; profile != "" {
    profileCfg, ok := cfg.ToolPermissions.ToolProfiles[profile]
    if !ok {
        logger.WarnCF("agent", "Unknown tool profile, denying all tools", map[string]interface{}{
            "profile": profile,
        })
        return tools.NewToolRegistry()
    }
    filteredRegistry = applyToolProfile(filteredRegistry, profileCfg)
}
return filteredRegistry
```

Add helper function in `permissions.go`:

```go
// applyToolProfile applies an allow/deny list from a ToolProfileConfig.
func applyToolProfile(registry *tools.ToolRegistry, profile config.ToolProfileConfig) *tools.ToolRegistry {
    result := tools.NewToolRegistry()

    // Build deny set
    denySet := make(map[string]bool, len(profile.Deny))
    for _, name := range profile.Deny {
        denySet[name] = true
    }

    // Build allow set (empty = all allowed)
    allowSet := make(map[string]bool, len(profile.Allow))
    for _, name := range profile.Allow {
        allowSet[name] = true
    }

    for _, tool := range registry.All() {
        name := tool.Name()
        if denySet[name] {
            continue
        }
        if len(allowSet) > 0 && !allowSet[name] {
            continue
        }
        result.Register(tool)
    }
    return result
}
```

Run: `go test ./pkg/agent/... -run TestFilterToolsByPermissions -v` → **GREEN**.

Commit: `feat(agent): apply ToolProfileConfig after role filtering in filterToolsByPermissions`

---

## Phase 6: Hook Call Sites in loop.go

### 6.1 Add `hooks` field to `AgentLoop` and wire in `NewAgentLoop`

**File**: `pkg/agent/loop.go`

Add field to `AgentLoop` struct:
```go
hooks *hooks.HookRegistry
```

In `NewAgentLoop`, after current initialization:
```go
al.hooks = hooks.NewHookRegistry()
```

Update import: `"github.com/sipeed/makoclaw/pkg/hooks"`

**Verify**: `go build ./pkg/agent/...` compiles.

---

### 6.2 Call `hooks.Run("before_tool_call", ...)` before `ExecuteWithContext`

**File**: `pkg/agent/loop.go` — find the call site where tools are executed via `ExecuteWithContext`. Add before it:

```go
// Security hook: before_tool_call
if al.hooks != nil {
    hookResult := al.hooks.Run(ctx, "before_tool_call", hooks.HookContext{
        ToolName: toolCall.Name,
        Args:     toolCall.Args,
        UserID:   al.userID,
    })
    if hookResult.Action == hooks.HookBlock {
        toolResults = append(toolResults, providers.ToolResult{
            ToolCallID: toolCall.ID,
            Content:    fmt.Sprintf("[blocked by security hook: %s]", hookResult.Reason),
        })
        continue
    }
    if hookResult.Action == hooks.HookRequireApproval {
        // No owner channel configured: treat as block
        toolResults = append(toolResults, providers.ToolResult{
            ToolCallID: toolCall.ID,
            Content:    fmt.Sprintf("[tool requires approval: %s]", hookResult.Reason),
        })
        logger.WarnCF("agent", "require_approval with no owner channel, treating as block", map[string]interface{}{
            "tool": toolCall.Name,
        })
        continue
    }
}
```

---

### 6.3 Call `hooks.Run("message_sending", ...)` before `PublishOutbound`

**File**: `pkg/agent/loop.go` — find the primary `al.bus.PublishOutbound` call site (line ~744) and guard it:

```go
// Security hook: message_sending
if al.hooks != nil {
    hookResult := al.hooks.Run(ctx, "message_sending", hooks.HookContext{
        Message: response,
        UserID:  al.userID,
    })
    if hookResult.Action == hooks.HookBlock {
        logger.InfoCF("agent", "message_sending hook blocked delivery", map[string]interface{}{
            "reason": hookResult.Reason,
        })
        // Do not deliver
    } else {
        al.bus.PublishOutbound(bus.OutboundMessage{ /* existing */ })
    }
} else {
    al.bus.PublishOutbound(bus.OutboundMessage{ /* existing */ })
}
```

**Verify**: `go build ./pkg/agent/...` and `go test ./pkg/agent/... -v` pass.

Commit: `feat(agent): wire HookRegistry into tool execution and message delivery in AgentLoop`

---

## Phase 7: Integration Tests

### 7.1 Verify shell_test.go still passes (no regressions)

```bash
go test ./pkg/tools/... -run TestShell -v
```

Expected: all existing shell tests pass unchanged.

---

### 7.2 Run all tests and confirm no regressions

```bash
go test ./... -count=1
```

Expected output: all packages pass, no new failures.

---

### 7.3 Verify `go vet` is clean

```bash
go vet ./...
```

Expected: no output (exit 0).

Commit: `test(security): confirm no regressions across full test suite`

---

## Phase 8: Cleanup & Config Validation

### 8.1 Add startup validation for `ActiveProfile`

**File**: `pkg/agent/loop.go` — in `NewAgentLoop` or `NewAgentLoopForUser`, after loading config:

```go
if profile := cfg.ToolPermissions.ActiveProfile; profile != "" {
    if _, ok := cfg.ToolPermissions.ToolProfiles[profile]; !ok {
        return nil, fmt.Errorf("config error: unknown tool_profile %q — available: %v",
            profile, availableProfileNames(cfg.ToolPermissions.ToolProfiles))
    }
}
```

Add helper:
```go
func availableProfileNames(profiles map[string]config.ToolProfileConfig) []string {
    names := make([]string, 0, len(profiles))
    for k := range profiles {
        names = append(names, k)
    }
    sort.Strings(names)
    return names
}
```

---

### 8.2 Write startup validation test

**File**: `pkg/agent/permissions_test.go`:

```go
func TestFilterToolsByPermissions_UnknownProfileReturnsEmpty(t *testing.T) {
    cfg := config.NewDefaults()
    cfg.ToolPermissions.ActiveProfile = "does-not-exist"
    base := tools.NewToolRegistry()
    base.Register(&fakeNamedTool{name: "exec"})
    filtered := filterToolsByPermissions(base, "admin", 0, cfg, nil)
    require.Equal(t, 0, len(toolNames(filtered)), "unknown profile must return empty registry")
}
```

Run: `go test ./pkg/agent/... -run TestFilterToolsByPermissions_UnknownProfileReturnsEmpty -v` → **GREEN** (already handled in 5.2).

---

### 8.3 Run full test suite one final time

```bash
go test -race ./... -count=1
```

Expected: all tests pass, no race conditions detected.

Commit: `feat(agent): add startup validation for unknown ActiveProfile`

---

## Implementation Order Summary

| Phase | Tasks | Focus |
|-------|-------|-------|
| Phase 1 | 1.1–1.8 | Config types + PairingStore |
| Phase 2 | 2.1–2.2 | HookRegistry package |
| Phase 3 | 3.1–3.4 | BaseChannel DM policy |
| Phase 4 | 4.1–4.2 | /approve command |
| Phase 5 | 5.1–5.2 | Tool profiles in permissions |
| Phase 6 | 6.1–6.3 | Hook call sites in loop.go |
| Phase 7 | 7.1–7.3 | Integration verification |
| Phase 8 | 8.1–8.3 | Startup validation + cleanup |
| **Total** | **23** | |

**Dependency order**: Phases 1 and 2 are independent and can proceed in parallel. Phase 3 depends on Phase 1 (PairingStore). Phase 4 depends on Phase 1 (PairingStore) and Phase 3 (CommandHandler). Phase 5 depends on Phase 1 (config types). Phase 6 depends on Phase 2 (hooks package). Phases 7–8 depend on all prior phases.
