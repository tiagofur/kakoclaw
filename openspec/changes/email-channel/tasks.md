# Tasks: Email as Two-Way Communication Channel

## Phase 1: Foundation (Config & Dependencies)

- [ ] 1.1 Add `EmailChannelConfig` struct to `pkg/config/config.go` with all fields from design (IMAP/SMTP host/port, username, password, from, allow_from, poll_interval_seconds, mailbox, mark_as_read, max_email_size_mb, insecure_skip_verify). Add `Email EmailChannelConfig` field to `ChannelsConfig`. Set defaults in `getEmptyChannelsConfig()`. Add `!c.Email.Enabled` check in `isChannelsConfigEmpty()`. Add email to `GetActiveChannels()`.
- [ ] 1.2 Add `github.com/emersion/go-imap/v2` and `github.com/emersion/go-message` to `go.mod` via `go get`.

## Phase 2: Core Implementation (IMAP + SMTP + MIME)

- [ ] 2.1 Create `pkg/channels/email.go` — define `EmailChannel` struct (embed `*BaseChannel`), `threadState`, `uidState` types, and `NewEmailChannel()` constructor per design signatures.
- [ ] 2.2 Implement UID persistence: `loadUIDState()` reads `<workspace>/email_uid.json`, `saveUIDState()` writes it. Handle missing file (first run) and UIDValidity changes (reset lastUID).
- [ ] 2.3 Implement `Start()` — launch goroutine with `time.Ticker` at `poll_interval_seconds`. Implement `Stop()` — cancel context, set `running=false`. Implement `SetCommandHandler()`.
- [ ] 2.4 Implement `fetchNewEmails()` — IMAP TLS dial (respect `insecure_skip_verify`), login, SELECT mailbox, SEARCH UNSEEN with UID > lastUID, FETCH envelope+body, optionally STORE +FLAGS \Seen, update UID state. Include exponential backoff on consecutive failures (cap 5min), reset on success.
- [ ] 2.5 Implement `handleEmail()` — parse MIME via `go-message`, prefer `text/plain`, fallback `htmlToPlaintext()` for HTML-only. Append `[attachment: name (size)]` markers. Enforce `max_email_size_mb` skip. Build `bus.InboundMessage` with `Channel: "email"`, `SenderID: sender-address`, `SessionKey: "email:<thread-root>"`. Publish via bus. **Security**: skip if sender not in allow_from via `BaseChannel.IsAllowed()`.
- [ ] 2.6 Implement `resolveThreadRoot()` — return first entry of References header; if empty, use own Message-ID. Implement `htmlToPlaintext()` via `x/net/html` tokenizer.
- [ ] 2.7 Implement `Send()` — construct RFC 2822 message with `Message-ID: <uuid@makoclaw>`, `In-Reply-To`, `References` chain from `threadState` sync.Map lookup, `Re:` subject prefix. Send via `net/smtp.SendMail` over TLS (port from config). Update thread state after send.

## Phase 3: Integration (Channel Manager + Multi-User)

- [ ] 3.1 Modify `pkg/channels/manager.go` — add email init block in `initChannels()` following Telegram pattern: check `cfg.Channels.Email.Enabled`, call `NewEmailChannel(...)`, apply user resolver, store in `m.channels["email"]`. Add `case "email"` in `RestartChannel()`. Pass workspace path to constructor (or add `SetWorkspace` post-construction if manager pattern requires it).
- [ ] 3.2 Verify multi-user isolation: UID file stored in user workspace (`users/{uuid}/workspace/email_uid.json`), per-user config merge for `EmailChannelConfig`. No new code expected — existing `LoadConfigForUser` + workspace isolation handles this, but confirm the workspace path flows through correctly.

## Phase 4: Testing

- [ ] 4.1 Create `pkg/channels/email_test.go` — table-driven unit tests for `resolveThreadRoot()` (no References, single ref, multi-ref chain), `htmlToPlaintext()` (tags, entities, nested elements), and UID state load/save (missing file, valid file, UIDValidity mismatch). Use `t.TempDir()` for file tests.
- [ ] 4.2 Add test for `Send()` threading headers — mock SMTP to capture raw message bytes, verify `In-Reply-To`, `References`, `Subject: Re:` prefix, `Message-ID` format.
- [ ] 4.3 Add config merge test in `pkg/config/config_test.go` — verify `EmailChannelConfig` defaults, env var overrides (`MAKOCLAW_CHANNELS_EMAIL_*`), and multi-user config merge behavior.

## Phase 5: Polish (Logging, Security, Docs)

- [ ] 5.1 Add structured logging throughout `email.go`: component `"email"` — info on start/stop/poll-cycle, debug for skipped senders and poll-no-ops, warn for `insecure_skip_verify`, warn for empty `allow_from` with channel enabled, error for IMAP/SMTP failures. **Never log password field.**
- [ ] 5.2 Run `make fmt && make vet` to verify code quality. Fix any issues found.
