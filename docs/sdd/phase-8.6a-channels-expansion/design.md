# Design: Phase 8.6a — Channels Expansion

## Technical Approach

Extend `ChannelsConfig` with a generic `MultiAccountConfig[T]` wrapper that accepts both the legacy single-object form and a new array form via custom `UnmarshalJSON` (mirrors the existing `FlexibleStringSlice` pattern). `initChannels()` becomes a generic loop over account slices. Nine new channel adapters follow the existing `BaseChannel` embed pattern. A `RouteResolver` struct handles outbound account selection.

---

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|--------------|-----------|
| Config migration strategy | `UnmarshalJSON` fallback on each field | Separate migration script, new top-level key | Matches `FlexibleStringSlice` precedent; zero user action required |
| Multi-account registry key | `"{channel}:{account_id}"` | Separate map per channel, numeric index | Lookup by string key already used in `dispatchOutbound`; collision-free across channels |
| Routing layer location | New file `pkg/channels/routing.go` | Inline in manager, bus middleware | Keeps manager focused on lifecycle; routing logic testable in isolation |
| Session key format | `"{channel}:{account_id}:{chatID}"` | Hash-based, numeric index | Human-readable in session files; consistent with existing `channel:chatID` shape |
| New adapters: base embed | `*BaseChannel` embed (same as Telegram, Discord, etc.) | Composition via field | All existing adapters use embed; consistent handler access |
| `account_id` propagation | Written into `InboundMessage.Metadata["account_id"]` by `HandleMessage` | New field on `InboundMessage` | `Metadata map[string]string` already exists; avoids struct API break |

---

## Data Flow

### Inbound (multi-account)

```
config.json (array) ──→ MultiAccountConfig[T].UnmarshalJSON
                              │
                    ┌─────────┴──────────┐
               account[0]          account[1]
                    │                    │
           NewXChannel(cfg)      NewXChannel(cfg)
                    │                    │
         channels["x:acc0"]   channels["x:acc1"]
                    │                    │
            BaseChannel.HandleMessage(...)
                    │  sets Metadata["account_id"]
                    │  sets SessionKey = "x:acc0:chatID"
                    └──────── bus.PublishInbound ──→ AgentManager
```

### Outbound (RouteResolver)

```
OutboundMessage{Channel:"telegram", Metadata:{...}}
        │
  RouteResolver.Resolve(msg, channels)
        │
  1. Metadata["account_id"] set? → channels["telegram:{id}"]
  2. Peer binding match?          → bound account
  3. First enabled account        → fallback
        │
  channel.Send(ctx, msg)
```

---

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/config/config.go` | Modify | Add `MultiAccountConfig[T]` generic + `UnmarshalJSON`; add 9 new config structs; update `ChannelsConfig` fields |
| `pkg/channels/base.go` | Modify | `HandleMessage` writes `account_id` to `Metadata`; session key updated to `channel:account_id:chatID` when `account_id` != "" |
| `pkg/channels/manager.go` | Modify | `initChannels()` iterates `[]T` slices; registry key = `channel:account_id`; `isChannelsConfigEmpty()` updated; `RestartChannel` updated |
| `pkg/channels/routing.go` | Create | `RouteResolver` struct with `Resolve` method |
| `pkg/channels/line.go` | Create | LINE Messaging API adapter |
| `pkg/channels/wechat.go` | Create | WeChat Official Accounts adapter |
| `pkg/channels/zalo.go` | Create | Zalo OA adapter |
| `pkg/channels/mattermost.go` | Create | Mattermost Bot adapter |
| `pkg/channels/matrix.go` | Create | Matrix Client-Server adapter |
| `pkg/channels/nostr.go` | Create | Nostr NIP-04 DM adapter |
| `pkg/channels/twitch.go` | Create | Twitch IRC adapter |
| `pkg/channels/irc.go` | Create | IRC RFC-1459 adapter |
| `pkg/channels/twilio_voice.go` | Create | Twilio Voice TwiML webhook adapter |
| `pkg/web/handlers_user_config.go` | Modify | Channel config API returns/accepts arrays |
| `go.mod` | Modify | 7 new dependencies (see below) |

---

## Interfaces / Contracts

### `MultiAccountConfig[T]`

```go
// pkg/config/config.go
type MultiAccountConfig[T any] struct {
    Accounts []T
}

func (m *MultiAccountConfig[T]) UnmarshalJSON(data []byte) error {
    // Try []T first (new form)
    var arr []T
    if err := json.Unmarshal(data, &arr); err == nil {
        m.Accounts = arr
        return nil
    }
    // Fallback: single object (legacy form) → promote to 1-element slice
    var single T
    if err := json.Unmarshal(data, &single); err != nil {
        return err
    }
    m.Accounts = []T{single}
    return nil
}
```

### Account ID convention

Each channel config struct gains:
```go
AccountID string `json:"account_id"` // defaults to "default" when empty
```

### `ChannelsConfig` updated fields (example)

```go
type ChannelsConfig struct {
    Telegram MultiAccountConfig[TelegramConfig] `json:"telegram"`
    // ... same pattern for all 9 existing + 9 new channels
}
```

### `RouteResolver`

```go
// pkg/channels/routing.go
type RouteResolver struct{}

func (r *RouteResolver) Resolve(msg bus.OutboundMessage, channels map[string]Channel) (Channel, error) {
    // 1. Explicit account_id in Metadata
    if id, ok := msg.Metadata["account_id"]; ok && id != "" {
        key := fmt.Sprintf("%s:%s", msg.Channel, id)
        if ch, ok := channels[key]; ok { return ch, nil }
    }
    // 2. Exact channel:default fallback
    key := fmt.Sprintf("%s:default", msg.Channel)
    if ch, ok := channels[key]; ok { return ch, nil }
    // 3. First key matching "channel:" prefix
    prefix := msg.Channel + ":"
    for k, ch := range channels {
        if strings.HasPrefix(k, prefix) { return ch, nil }
    }
    return nil, fmt.Errorf("no account found for channel %q", msg.Channel)
}
```

### New channel struct skeleton (identical pattern for all 9)

```go
type LINEChannel struct {
    *BaseChannel
    config config.LINEConfig
    // SDK client field
}

func NewLINEChannel(cfg config.LINEConfig, bus *bus.MessageBus) (*LINEChannel, error) { ... }
func (c *LINEChannel) Start(ctx context.Context) error  { ... }
func (c *LINEChannel) Stop(ctx context.Context) error   { ... }
func (c *LINEChannel) Send(ctx context.Context, msg bus.OutboundMessage) error { ... }
```

### New Go dependencies

| Module | Version | Channel |
|--------|---------|---------|
| `github.com/line/line-bot-sdk-go/v8` | v8.x | LINE |
| `github.com/mattermost/mattermost-server/v6/model` | v6.x | Mattermost |
| `maunium.net/go/mautrix` | v0.x | Matrix |
| `github.com/nbd-wtf/go-nostr` | v0.x | Nostr |
| `github.com/gempir/go-twitch-irc/v4` | v4.x | Twitch |
| `github.com/thoj/go-ircevent` | latest | IRC |
| `github.com/twilio/twilio-go` | v1.x | Twilio Voice |

WeChat and Zalo use standard `net/http` — no new deps.

---

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `MultiAccountConfig.UnmarshalJSON` — legacy object promoted to 1-element array | Table-driven, `encoding/json` |
| Unit | `RouteResolver.Resolve` — explicit id / default / prefix fallback | Table-driven with mock channel map |
| Unit | Session key format in `BaseChannel.HandleMessage` | Verify `InboundMessage.SessionKey` value |
| Unit | `isChannelsConfigEmpty()` with old and new config shapes | Assert correct bool result |
| Integration | Two Telegram accounts register as independent channels | Verify `channels["telegram:acc1"]` and `channels["telegram:acc2"]` both exist |
| Smoke | Each new adapter: `Start()` returns error on bad token (not panic) | Pass invalid credentials, assert `err != nil` |

---

## Migration / Rollout

Config is additive. `UnmarshalJSON` on `MultiAccountConfig[T]` accepts both old single-object and new array JSON transparently — existing `config.json` files load without changes. New adapters default to `Enabled: false`. No DB schema changes.

---

## Open Questions

- [ ] WeChat: server domain verification required by WeChat before webhook delivery — document setup step or add startup health check?
- [ ] Nostr: recommend `nsec` via env var; add log warning if key is present in config file?
- [ ] Twilio Voice: implement `X-Twilio-Signature` HMAC validation in webhook handler (security-critical, should be in scope for initial adapter)?
