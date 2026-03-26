# Proposal: Email as Two-Way Communication Channel

## Intent

MakoClaw has outbound-only email (SMTP tool). Users need MakoClaw to **receive** emails via IMAP, process them through the agent loop, and reply with proper threading. This adds email as the 10th channel, matching the existing Channel interface pattern.

## Scope

### In Scope
- `EmailChannel` implementing `Channel` interface with IMAP polling + SMTP reply
- `EmailChannelConfig` in `ChannelsConfig` (separate from `EmailToolsConfig`)
- Channel manager registration (init, restart, active-channels check)
- Email threading via Message-ID/In-Reply-To/References headers
- HTML-to-plaintext fallback for inbound emails
- Attachment detection markers (`[attachment: file.pdf (2MB)]`)
- UID persistence across restarts
- Allow-list filtering by sender email address
- Dependencies: `go-imap/v2`, `go-message`

### Out of Scope
- IMAP IDLE push (future optimization)
- OAuth2/XOAUTH2 authentication
- Attachment download/upload
- Rich HTML outbound rendering
- Domain wildcard allow-list (`*@company.com`)
- Frontend settings UI for email channel (deferred)

## Approach

Polling-based IMAP channel following the Telegram long-poll pattern. A goroutine polls IMAP on a configurable interval (default 60s), publishes `InboundMessage` to the bus. Outbound replies sent via SMTP with full RFC 2822 threading headers. Session key: `email:<thread-root-message-id>`. Single file implementation (~400-500 LOC).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/channels/email.go` | **New** | EmailChannel struct, IMAP polling loop, SMTP send, threading |
| `pkg/config/config.go` | Modified | Add `EmailChannelConfig`, wire into `ChannelsConfig`, defaults |
| `pkg/channels/manager.go` | Modified | Register email in `initChannels()`, `RestartChannel()`, `isChannelsConfigEmpty()` |
| `go.mod` / `go.sum` | Modified | Add `go-imap/v2` + `go-message` dependencies |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Gmail/Outlook require OAuth2 | High | Support app-specific passwords; document OAuth2 as Phase 2 |
| Large emails spike memory | Medium | Config max email size (default 10MB), skip oversized |
| Thread detection fails on forwarded emails | Medium | Fall back to new session when References missing |
| IMAP server rate limiting | Low | Configurable poll interval + exponential backoff on errors |

## Rollback Plan

Fully additive change. Rollback = remove `pkg/channels/email.go`, revert config struct additions and manager registration. No schema migrations, no breaking changes to existing channels or the email tool.

## Dependencies

- `github.com/emersion/go-imap/v2` — IMAP client library
- `github.com/emersion/go-message` — MIME/email message parsing

## Success Criteria

- [ ] Agent receives an email sent to the configured IMAP inbox and processes it as an inbound message
- [ ] Agent replies via SMTP with correct In-Reply-To and References headers (thread preserved in email clients)
- [ ] Emails from senders not in the allow-list are ignored
- [ ] Poll loop survives IMAP connection errors with retry
- [ ] UID tracking persists across restarts (no duplicate processing)
- [ ] Non-breaking: existing channels and email tool unaffected
- [ ] Works in both single-user and multi-user deployments
