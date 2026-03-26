# Exploration: Email as a Two-Way Communication Channel

## Problem Statement

MakoClaw currently has outbound-only email support via the `send_email_report` tool (`pkg/tools/email.go`), which sends emails through SMTP. There is no way for MakoClaw to **receive** emails and process them as inbound messages. Users want MakoClaw to monitor an IMAP inbox, treat incoming emails as chat messages, and reply via SMTP with proper email threading (Message-ID, In-Reply-To, References headers).

**Current state:**
- Outbound SMTP only via `EmailTool` in `pkg/tools/email.go`
- SMTP config lives in `ToolsConfig.Email` (`EmailToolsConfig` struct)
- No IMAP dependency in `go.mod`
- No email channel in `ChannelsConfig` (9 channels exist: Telegram, Discord, Slack, WhatsApp, Signal, QQ, DingTalk, Feishu, MaixCam)

**Desired state:**
- Email channel that periodically polls an IMAP inbox for new messages
- Incoming emails processed as `bus.InboundMessage` through the standard message bus
- Outbound replies sent via SMTP with proper threading headers
- Allow-list filtering by sender email address
- HTML-to-plaintext conversion for inbound emails
- Proper session key strategy for email conversations
- Attachment awareness (at minimum, mention of attachments; ideally, file download)

---

## Investigation Summary

### 1. Channel Interface Pattern (`pkg/channels/base.go`)

Every channel implements the `Channel` interface:

```go
type Channel interface {
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Send(ctx context.Context, msg bus.OutboundMessage) error
    IsRunning() bool
    IsAllowed(senderID string) bool
    GetUserIDForSender(senderID string) (int64, error)
    SetCommandHandler(*CommandHandler)
}
```

`BaseChannel` provides shared functionality:
- **Allow-list filtering** via `IsAllowed(senderID)` - supports compound IDs like `"123|username"` and `@`-prefixed usernames
- **Message publishing** via `HandleMessage(senderID, chatID, content, media, metadata)` - builds `InboundMessage` and publishes to bus
- **User resolution** via `SetUserResolver()` for senderID-to-userID mapping
- **Blocked user checks** via `SetStorage()`
- **Session key construction**: `channel:chatID` format (e.g., `telegram:123456`)

For email, the senderID would naturally be the sender's email address, and the chatID would be derived from the email thread (via Message-ID/References).

### 2. Reference: Telegram Channel Polling Pattern (`pkg/channels/telegram.go`)

Telegram uses a **long-polling** pattern that closely mirrors what email IMAP polling would look like:

```go
func (c *TelegramChannel) Start(ctx context.Context) error {
    c.setRunning(true)
    go func() {
        var offset int
        for {
            select {
            case <-ctx.Done(): return
            default:
            }
            result, err := c.bot.GetUpdates(ctx, &telego.GetUpdatesParams{
                Offset:  offset,
                Timeout: 30,
            })
            // ... process updates, handle errors with 5s retry
        }
    }()
    return nil
}
```

Key patterns to replicate:
- Goroutine-based polling loop with context cancellation
- Offset tracking (Telegram uses update IDs; email uses IMAP UIDs or UIDVALIDITY+UID)
- Error retry with backoff (currently hardcoded 5s)
- Allow-list check before processing media/content
- Metadata extraction (sender info, message IDs)
- Thinking indicator (not applicable for email, but acknowledgment reply could serve similar purpose)

### 3. Channel Manager Registration (`pkg/channels/manager.go`)

The `initChannels()` method follows a consistent pattern for each channel:

```go
if m.config.Channels.Email.Enabled && m.config.Channels.Email.IMAPHost != "" {
    email, err := NewEmailChannel(m.config.Channels.Email, m.bus)
    if err != nil { /* log error */ } else {
        m.applyUserResolver("email", email)
        m.channels["email"] = email
    }
}
```

Must also add to `RestartChannel()` switch statement and `GetActiveChannels()`.

The `isChannelsConfigEmpty()` function (line 1080) must include `!c.Email.Enabled`.

### 4. Message Bus Types (`pkg/bus/types.go`)

```go
type InboundMessage struct {
    UserID     int64
    Channel    string            // "email"
    SenderID   string            // sender email address
    ChatID     string            // thread ID (derived from Message-ID chain)
    Content    string            // email body (plaintext)
    Media      []string          // attachment paths (optional)
    SessionKey string            // "email:<thread-id>"
    Metadata   map[string]string // subject, message_id, in_reply_to, etc.
}

type OutboundMessage struct {
    UserID   int64
    Channel  string            // "email"
    ChatID   string            // thread ID
    Content  string            // reply body
    Metadata map[string]string // threading headers, subject
}
```

The existing structures are sufficient. Thread metadata can flow through `Metadata` for the `Send()` method to construct proper threading headers.

### 5. Existing Email Tool (`pkg/tools/email.go`)

The `EmailTool` has SMTP configuration in `EmailToolsConfig`:

```go
type EmailToolsConfig struct {
    Enabled  bool   `json:"enabled"`
    Host     string `json:"host"`      // SMTP host
    Port     int    `json:"port"`      // SMTP port
    Username string `json:"username"`  // SMTP auth username
    Password string `json:"password"`  // SMTP auth password
    From     string `json:"from"`      // Sender address
    To       string `json:"to"`        // Default recipient
}
```

Useful existing functions:
- `parseFromAddress()` - parses "Name <email>" format
- `normalizedSMTPPassword()` - handles Gmail app passwords (strips spaces)
- Header injection sanitization

### 6. Go Module Dependencies (`go.mod`)

No IMAP library currently present. Options:

| Library | Stars | Maintenance | Notes |
|---------|-------|-------------|-------|
| `github.com/emersion/go-imap/v2` | ~2k | Active | Modern, well-maintained, IDLE support |
| `github.com/emersion/go-message` | ~400 | Active (same author) | MIME parsing, pairs with go-imap |
| `golang.org/x/net/html` | stdlib | N/A | HTML-to-text conversion (already indirect dep) |

`go-imap/v2` + `go-message` from Simon Ser (emersion) is the de facto standard for Go IMAP. Both are actively maintained and used in production by projects like aerc mail client.

---

## Analysis of Key Design Decisions

### A. IMAP Polling vs IDLE Push

| Aspect | Polling | IDLE |
|--------|---------|------|
| Latency | Configurable (30s-5min) | Near-instant (~1-5s) |
| Server load | Periodic connections | Persistent connection |
| Complexity | Simple (matches Telegram pattern) | Connection keepalive, heartbeat, reconnect |
| Server support | Universal | Most modern servers (Gmail, Outlook, Fastmail) |
| Resource usage | Brief connection per poll | Long-lived TCP connection |
| Firewall/NAT | No issues | May timeout behind NAT |

**Recommendation: Start with polling, design for future IDLE support.**

Polling matches MakoClaw's existing architecture (Telegram uses polling too). The polling interval should be configurable (default: 60s). The interface should be designed so IDLE can be added later as an optimization without changing the channel contract.

For the resource-constrained hardware target (<10MB RAM), polling is also more predictable in memory usage.

### B. Email Threading Strategy

Email threading uses three headers:
- **Message-ID**: Unique ID for each email (generated on send)
- **In-Reply-To**: Message-ID of the email being replied to
- **References**: Space-separated list of all Message-IDs in the thread

**Session key strategy**: `email:<thread-root-message-id>`

When an inbound email arrives:
1. Extract `References` header (or `In-Reply-To` as fallback)
2. The first Message-ID in `References` is the thread root
3. If no references (new conversation), use the email's own `Message-ID` as the thread root
4. Session key = `email:<normalized-thread-root>`

When sending a reply:
1. Generate a new Message-ID: `<uuid@makoclaw>`
2. Set `In-Reply-To` to the most recent inbound Message-ID
3. Set `References` to the full chain from the original + our Message-ID
4. Prefix subject with `Re:` if not already present

This requires storing a mapping of `thread-root -> last-message-id` and `thread-root -> references-chain`. This can be stored in `Metadata` on the `OutboundMessage` or in a lightweight in-memory cache with periodic persistence.

### C. HTML vs Plaintext Handling

**Inbound**: Many emails arrive as `multipart/alternative` with both HTML and plaintext parts. Strategy:
1. Prefer `text/plain` part if available
2. If only HTML, strip tags and convert to readable text (using `golang.org/x/net/html` tokenizer)
3. Preserve basic structure (paragraphs, lists, links)

**Outbound**: Send as `text/plain` for simplicity. MakoClaw's responses are typically plain text or markdown. A future enhancement could add `multipart/alternative` with markdown-rendered HTML.

### D. Attachment Support

**Phase 1 (minimum viable)**: Detect attachments and include `[attachment: filename.pdf (2.3MB)]` markers in the content. No download.

**Phase 2 (future)**: Download attachments to a temp directory, pass paths via `Media []string` field on `InboundMessage`. This mirrors how Telegram handles photos/documents. Cleanup after processing using `defer` pattern from Telegram channel.

For outbound, attachment support is out of scope for the initial implementation.

### E. Allow-List Pattern for Email

The existing `BaseChannel.IsAllowed()` already handles string matching well. For email:
- SenderID = email address (e.g., `user@example.com`)
- Allow-list entries can be exact addresses or domain wildcards

Domain wildcards would be a new capability. Options:
1. **Exact match only** (simplest, consistent with other channels): `["user@example.com", "boss@company.com"]`
2. **Domain suffix matching** (useful for corporate): `["*@company.com"]`

**Recommendation**: Start with exact match only (option 1). The existing `IsAllowed()` can handle this without modification. Domain wildcards can be added later.

### F. IMAP Connection Management

**Connection pooling**: Not needed initially. A single IMAP connection per poll cycle is sufficient. Connect, fetch, disconnect. This avoids stale connection issues and is simpler for resource-constrained environments.

**Reconnection**: The polling loop naturally handles this - each poll cycle creates a fresh connection. If connection fails, log error and retry on next cycle (with the retry interval).

**TLS**: Require TLS by default (`IMAPS` on port 993). Allow optional `insecure_skip_verify` for self-signed certs in development.

### G. Config Structure: Reuse vs Separate

The email channel needs both IMAP (inbound) and SMTP (outbound) configuration. The existing `EmailToolsConfig` only has SMTP.

**Options**:

1. **Separate config entirely** (in `ChannelsConfig`):
```json
{
  "channels": {
    "email": {
      "enabled": true,
      "imap_host": "imap.gmail.com",
      "imap_port": 993,
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "username": "user@gmail.com",
      "password": "app-password",
      "from": "MakoClaw <user@gmail.com>",
      "allow_from": ["boss@company.com"],
      "poll_interval_seconds": 60,
      "mailbox": "INBOX",
      "mark_as_read": true
    }
  }
}
```

2. **Reuse SMTP from tools, add IMAP in channel**:
```json
{
  "channels": {
    "email": {
      "enabled": true,
      "imap_host": "imap.gmail.com",
      "imap_port": 993,
      "smtp_from_tools": true,
      "allow_from": ["boss@company.com"],
      "poll_interval_seconds": 60
    }
  }
}
```

**Recommendation: Option 1 (separate config).**

Reasons:
- Self-contained: channel works without tools config being set up
- Different credentials may be needed (IMAP vs SMTP auth can differ)
- Follows the pattern of every other channel having all its config in `ChannelsConfig`
- The existing `EmailTool` continues to work independently for one-off report sending
- No coupling between the tool and the channel

However, we should note that when both are configured with the same account, there's duplication. A future `smtp_from_tools: true` flag could reduce this, but it adds coupling that isn't worth the complexity initially.

---

## Affected Areas

### Must Modify
| File | Change |
|------|--------|
| `pkg/config/config.go` | Add `EmailChannelConfig` struct, add to `ChannelsConfig`, update `DefaultConfig()`, `isChannelsConfigEmpty()`, `getEmptyChannelsConfig()` |
| `pkg/channels/manager.go` | Add email channel init in `initChannels()`, `RestartChannel()` switch case |
| `go.mod` / `go.sum` | Add `github.com/emersion/go-imap/v2` and `github.com/emersion/go-message` |

### Must Create
| File | Description |
|------|-------------|
| `pkg/channels/email.go` | `EmailChannel` struct implementing `Channel` interface with IMAP polling + SMTP reply |

### Should Modify
| File | Change |
|------|--------|
| `pkg/config/config.go` | `GetActiveChannels()` - add email channel check |
| `pkg/web/frontend/src/components/Settings/` | Add email channel configuration UI (if applicable) |

### No Changes Needed
| File | Reason |
|------|--------|
| `pkg/tools/email.go` | Remains independent; the channel has its own SMTP sending |
| `pkg/bus/types.go` | Existing `InboundMessage`/`OutboundMessage` structures are sufficient |
| `pkg/channels/base.go` | `BaseChannel` already provides all needed shared functionality |
| `pkg/channels/multiuser_manager.go` | No changes needed - `NewManager()` will pick up the email channel automatically |

---

## Approaches

### Approach A: Polling-Based Email Channel (Recommended)

**Description**: Implement `EmailChannel` using periodic IMAP polling, following the Telegram channel pattern. SMTP replies include proper threading headers. Single file implementation (`email.go`), ~400-500 lines.

**Architecture**:
```
[IMAP Server] <--poll-- EmailChannel --publish--> MessageBus --> AgentLoop
                              ^
[SMTP Server] <--send--- EmailChannel <--subscribe-- MessageBus <-- AgentLoop
```

**Pros**:
- Simple, matches existing patterns perfectly
- Predictable resource usage (brief connections)
- Universal IMAP server compatibility
- Easy to test (mock IMAP server)

**Cons**:
- Latency depends on poll interval (30s-5min)
- Slightly more server load than IDLE for very active inboxes

**Complexity**: Medium (2-3 days)

### Approach B: IDLE-Based Email Channel

**Description**: Use IMAP IDLE command for push notifications of new emails. Falls back to polling if IDLE is not supported.

**Pros**:
- Near-instant email processing
- Lower server load for low-traffic inboxes

**Cons**:
- Significantly more complex (connection keepalive, heartbeat every 29 minutes per RFC 2177, reconnection logic, NAT timeout handling)
- Not all IMAP servers support IDLE
- Long-lived connections harder to manage on resource-constrained hardware
- More failure modes

**Complexity**: High (4-6 days)

### Approach C: Hybrid (Polling + Optional IDLE)

**Description**: Start with polling as default, add IDLE as opt-in feature behind a config flag.

**Pros**:
- Best of both worlds
- Incremental complexity

**Cons**:
- Two code paths to maintain
- More config surface

**Complexity**: High initially, but Approach A can be extended to C later without breaking changes.

---

## Recommendation

**Go with Approach A (Polling-Based)**, with the interface designed so IDLE support can be added later (Approach C) as a non-breaking enhancement.

Specific recommendations:
1. **New `EmailChannelConfig`** in `ChannelsConfig` with full IMAP+SMTP config (not reusing tools config)
2. **Poll interval default**: 60 seconds, configurable
3. **Session key**: `email:<thread-root-message-id>` using RFC 2822 References header chain
4. **Threading**: Full Message-ID/In-Reply-To/References support on outbound
5. **Content**: Prefer plaintext, fallback to HTML-to-text stripping
6. **Attachments**: Phase 1 = mention markers only, Phase 2 = download to temp dir
7. **Allow-list**: Exact email match initially (leveraging existing `BaseChannel.IsAllowed()`)
8. **Dependencies**: `github.com/emersion/go-imap/v2` + `github.com/emersion/go-message`
9. **UID tracking**: Store last-seen UID in a file in the user's workspace to survive restarts

---

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Gmail/Outlook require OAuth2 for IMAP | High | Medium | Support app-specific passwords initially; OAuth2 XOAUTH2 SASL is a future enhancement |
| Large emails (big attachments) could spike memory | Medium | Medium | Set a max email size limit in config (default 10MB); skip oversized emails with warning |
| Thread detection fails for forwarded/mangled emails | Medium | Low | Fall back to new session if References/In-Reply-To missing |
| IMAP server rate limiting | Low | Medium | Configurable poll interval, exponential backoff on errors |
| Email body encoding issues (charset, quoted-printable) | Medium | Low | `go-message` library handles MIME decoding; edge cases in non-standard emails |
| Duplicate processing after restart | Low | Medium | UID persistence file; mark-as-read flag (default true) |

---

## Ready for Proposal

**Yes.** The exploration is complete. All major design decisions have been analyzed with clear recommendations. The implementation is well-scoped:

- **1 new file**: `pkg/channels/email.go` (~400-500 lines)
- **3 modified files**: `config.go`, `manager.go`, `go.mod`
- **2 new dependencies**: `go-imap/v2`, `go-message`
- **Estimated effort**: 2-3 days for core implementation + 1 day for tests
- **No breaking changes**: additive only, fully backward compatible
