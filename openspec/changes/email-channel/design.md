# Design: Email as Two-Way Communication Channel

## Technical Approach

Polling-based IMAP channel following the exact Telegram pattern: goroutine poll loop with `time.Ticker`, `BaseChannel` embedding, `context.Context` cancellation for graceful shutdown. SMTP sending reuses the `net/smtp` + header-construction patterns from `pkg/tools/email.go`. Thread state tracked in-memory with a `sync.Map` keyed by thread-root Message-ID.

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|-------------|-----------|
| Poll vs IDLE | Ticker-based polling (default 60s) | IMAP IDLE, hybrid | Matches Telegram pattern; predictable memory; universal server support; IDLE adds connection-keepalive complexity unsuitable for <10MB target |
| Config location | Separate `EmailChannelConfig` in `ChannelsConfig` | Reuse `EmailToolsConfig` SMTP fields | Every channel is self-contained; IMAP/SMTP creds may differ; no coupling to tools config |
| IMAP library | `emersion/go-imap/v2` + `emersion/go-message` | stdlib `net/mail` only | go-imap is de facto Go IMAP standard; handles MIME, charset, multipart; active maintenance |
| Session key | `email:<thread-root-message-id>` | `email:<sender>`, `email:<subject-hash>` | RFC 2822 References chain gives deterministic thread grouping; survives subject changes |
| UID tracking | File-based (`<workspace>/email_uid.json`) | SQLite table, in-memory only | Survives restarts without schema migration; simple JSON `{mailbox, uidvalidity, lastUID}` |
| HTML handling | Prefer `text/plain`; fallback HTML-to-text via `go-message` + `x/net/html` tokenizer | Always strip HTML; render markdown to HTML outbound | Simplest correct approach; outbound stays `text/plain` |
| Attachments | Marker-only: `[attachment: name (size)]` | Download to temp dir | Phase 1 scope; download is Phase 2 (mirrors Telegram media pattern) |
| Connection strategy | Connect-per-poll (fresh dial each cycle) | Connection pool, persistent conn | Avoids stale-connection bugs; brief connections fit resource-constrained target |

## Data Flow

```
IMAP Server                          MakoClaw                              SMTP Server
    │                                    │                                      │
    │◄── pollLoop connects (TLS) ───────│                                      │
    │    SEARCH UNSEEN SINCE lastUID    │                                      │
    │── returns new UIDs ──────────────►│                                      │
    │    FETCH uid (ENVELOPE, BODY)     │                                      │
    │── email data ───────────────────►│                                      │
    │    STORE +FLAGS \Seen             │                                      │
    │                                   │──► handleEmail()                     │
    │                                   │    parse MIME (go-message)           │
    │                                   │    extract thread-root from          │
    │                                   │      References / In-Reply-To       │
    │                                   │    build InboundMessage              │
    │                                   │──► BaseChannel.HandleMessage()       │
    │                                   │      ──► bus.PublishInbound()        │
    │                                   │            ──► AgentLoop             │
    │                                   │                  ──► LLM+Tools      │
    │                                   │◄── bus.OutboundMessage ─────────────│
    │                                   │    Send() constructs RFC 2822 msg   │
    │                                   │    In-Reply-To + References headers │
    │                                   │──────────────── smtp.SendMail() ──►│
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/channels/email.go` | Create | `EmailChannel` struct: IMAP poll loop, SMTP send, threading, MIME parsing (~450 LOC) |
| `pkg/config/config.go` | Modify | Add `EmailChannelConfig` struct + field in `ChannelsConfig` + defaults in `getEmptyChannelsConfig()` + `isChannelsConfigEmpty()` + `GetActiveChannels()` |
| `pkg/channels/manager.go` | Modify | Add email init block in `initChannels()` + `case "email"` in `RestartChannel()` |
| `go.mod` / `go.sum` | Modify | Add `github.com/emersion/go-imap/v2`, `github.com/emersion/go-message` |

## Interfaces / Contracts

```go
// pkg/config/config.go
type EmailChannelConfig struct {
    Enabled           bool                `json:"enabled" env:"MAKOCLAW_CHANNELS_EMAIL_ENABLED"`
    IMAPHost          string              `json:"imap_host" env:"MAKOCLAW_CHANNELS_EMAIL_IMAP_HOST"`
    IMAPPort          int                 `json:"imap_port" env:"MAKOCLAW_CHANNELS_EMAIL_IMAP_PORT"`
    SMTPHost          string              `json:"smtp_host" env:"MAKOCLAW_CHANNELS_EMAIL_SMTP_HOST"`
    SMTPPort          int                 `json:"smtp_port" env:"MAKOCLAW_CHANNELS_EMAIL_SMTP_PORT"`
    Username          string              `json:"username" env:"MAKOCLAW_CHANNELS_EMAIL_USERNAME"`
    Password          string              `json:"password" env:"MAKOCLAW_CHANNELS_EMAIL_PASSWORD"`
    From              string              `json:"from" env:"MAKOCLAW_CHANNELS_EMAIL_FROM"`
    AllowFrom         FlexibleStringSlice `json:"allow_from" env:"MAKOCLAW_CHANNELS_EMAIL_ALLOW_FROM"`
    Mailbox           string              `json:"mailbox" env:"MAKOCLAW_CHANNELS_EMAIL_MAILBOX"`           // default "INBOX"
    PollIntervalSecs  int                 `json:"poll_interval_seconds" env:"MAKOCLAW_CHANNELS_EMAIL_POLL_INTERVAL_SECONDS"` // default 60
    MarkAsRead        bool                `json:"mark_as_read" env:"MAKOCLAW_CHANNELS_EMAIL_MARK_AS_READ"` // default true
    MaxEmailSizeMB    int                 `json:"max_email_size_mb" env:"MAKOCLAW_CHANNELS_EMAIL_MAX_SIZE_MB"` // default 10
    InsecureSkipVerify bool               `json:"insecure_skip_verify" env:"MAKOCLAW_CHANNELS_EMAIL_INSECURE_SKIP_VERIFY"`
}

// pkg/channels/email.go
type EmailChannel struct {
    *BaseChannel
    config         config.EmailChannelConfig
    cancel         context.CancelFunc       // stops poll goroutine
    threads        sync.Map                 // threadRoot -> *threadState
    workspace      string                   // for UID persistence file
    commandHandler *CommandHandler
}

type threadState struct {
    LastMessageID string   // most recent Message-ID in thread
    References    []string // full References chain
    Subject       string   // original subject (for Re: prefix)
}

// UID persistence (JSON file at <workspace>/email_uid.json)
type uidState struct {
    Mailbox     string `json:"mailbox"`
    UIDValidity uint32 `json:"uid_validity"`
    LastUID     uint32 `json:"last_uid"`
}
```

### Key method signatures

```go
func NewEmailChannel(cfg config.EmailChannelConfig, bus *bus.MessageBus, workspace string) (*EmailChannel, error)
func (c *EmailChannel) Start(ctx context.Context) error      // launches poll goroutine
func (c *EmailChannel) Stop(ctx context.Context) error       // cancels context, sets running=false
func (c *EmailChannel) Send(ctx context.Context, msg bus.OutboundMessage) error // SMTP with threading
func (c *EmailChannel) SetCommandHandler(handler *CommandHandler)

// internal
func (c *EmailChannel) pollLoop(ctx context.Context)         // ticker loop, connect-fetch-disconnect
func (c *EmailChannel) fetchNewEmails(ctx context.Context) error // single IMAP poll cycle
func (c *EmailChannel) handleEmail(envelope, body) error     // parse MIME → InboundMessage
func (c *EmailChannel) resolveThreadRoot(refs []string, inReplyTo, messageID string) string
func (c *EmailChannel) loadUIDState() (*uidState, error)
func (c *EmailChannel) saveUIDState(state *uidState) error
func (c *EmailChannel) htmlToPlaintext(html string) string   // x/net/html tokenizer
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `resolveThreadRoot` — various References/In-Reply-To combos | Table-driven tests, no IMAP needed |
| Unit | `htmlToPlaintext` — HTML stripping edge cases | Table-driven with sample HTML |
| Unit | `handleEmail` — MIME parsing, attachment markers, metadata extraction | Mock `go-message` entities |
| Unit | `Send` threading — correct In-Reply-To/References/Subject headers | Capture raw SMTP bytes via mock |
| Unit | UID state load/save — persistence and UIDValidity change handling | Temp file in `t.TempDir()` |
| Integration | Full poll cycle — connect, fetch, parse, publish to bus | Use `MailHog` or mock IMAP server (`go-imap/v2` includes `imapserver` package) |
| Integration | Config merge — `EmailChannelConfig` merges correctly in multi-user | Existing `config_test.go` pattern |

## Migration / Rollout

No migration required. Fully additive:
- New config field `channels.email` defaults to `enabled: false`
- No database schema changes
- UID state file created on first poll cycle
- Existing `EmailTool` (SMTP-only) continues independently

## Open Questions

- [ ] Should `NewEmailChannel` receive `workspace` as a parameter (like above) or resolve it from config? Telegram doesn't need workspace, but we need it for UID file. Manager currently passes config only — may need a `SetWorkspace` call post-construction.
- [ ] Should failed IMAP auth surface a user-visible error in web UI, or only log? Other channels only log.
