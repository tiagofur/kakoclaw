# Proposal: Phase 8.6a — Channels Expansion

**Change**: `phase-8.6a-channels-expansion`
**Status**: Draft
**Inspired by**: OpenClaw (https://github.com/openclaw/openclaw)
**Date**: 2026-03-30

---

## Intent

MakoClaw currently supports one account per channel per instance. This blocks real-world deployments that need multiple WhatsApp numbers, Telegram bots, or Discord bots serving different users/agents from a single binary. Additionally, 9 high-demand messaging platforms (LINE, WeChat, Zalo, Mattermost, Matrix, Nostr, Twitch, IRC, Twilio Voice) have no adapter, limiting MakoClaw's reach in Asia, self-hosted environments, and decentralized ecosystems.

## Scope

### In Scope
- Multi-account support for all existing channels (array config replacing single-object config)
- Deterministic routing: peer → parentPeer → guildId+roles → accountId → fallback
- 9 new channel adapters: LINE, WeChat, Zalo, Mattermost, Matrix, Nostr, Twitch, IRC, Twilio Voice
- Config schema evolution: `ChannelsConfig` fields become `[]T` slices with per-entry `account_id`
- `initChannels()` updated to iterate account slices, registering one `Channel` instance per account
- `isChannelsConfigEmpty()` updated for new schema
- Web UI: multi-account listing and per-account enable/token fields

### Out of Scope
- OAuth flows or credential management UIs for new channels (basic token/key input only)
- Voice transcription for Twilio (audio bridge deferred)
- WeChat personal account API (Official Accounts only)
- End-to-end encryption key management for Matrix/Nostr (basic DM only)
- Full IRC network federation or bouncer support

## Approach

**Phase A — Config schema migration (non-breaking):**
Extend `ChannelsConfig` to support both the old single-object form and a new `[]AccountConfig` array form via custom `UnmarshalJSON` (same pattern as `FlexibleStringSlice`). Each account entry gains an `account_id` string. Existing single-object configs are transparently promoted to a one-element array.

**Phase B — Multi-account channel manager:**
`initChannels()` iterates account arrays. Each account creates one `Channel` instance registered under `"{channel}:{account_id}"` in the channel registry. `HandleMessage` embeds `account_id` in `InboundMessage.Metadata` for downstream routing. Session key becomes `{channel}:{account_id}:{chatID}`.

**Phase C — Routing layer:**
Implement `RouteResolver` (inspired by OpenClaw `src/routing/`) that selects which channel+account to use for outbound messages:
1. Exact peer match → bound account
2. Parent peer / guild roles match → account with role binding
3. Explicit `account_id` in outbound message
4. Default (first enabled) account fallback

**Phase D — New channel adapters (9):**

| Channel | Protocol/SDK | Auth |
|---------|-------------|------|
| LINE | Messaging API REST (line-bot-sdk-go) | Channel Access Token |
| WeChat | Official Accounts API (HTTP) | AppID + AppSecret |
| Zalo | Zalo Bot API (HTTP) | App ID + OA Access Token |
| Mattermost | Bot API + WebSocket (mattermost/mattermost-server/model) | Bot token |
| Matrix | Client-Server API (mautrix-go) | Access token + homeserver |
| Nostr | NIP-04 DMs (nbd-wtf/go-nostr) | Private key (nsec) |
| Twitch | IRC over WebSocket (gempir/go-twitch-irc) | OAuth token |
| IRC | RFC 1459 (thoj/go-ircevent) | Nick + optional SASL |
| Twilio Voice | REST API + TwiML webhooks (twilio-go) | Account SID + Auth Token |

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/config/config.go` | Modified | `ChannelsConfig` fields become flexible single/array + 9 new config structs |
| `pkg/channels/manager.go` | Modified | `initChannels()` iterates account arrays; `isChannelsConfigEmpty()` updated |
| `pkg/channels/base.go` | Modified | `InboundMessage` metadata gets `account_id`; session key includes account |
| `pkg/channels/` | New (×9) | line.go, wechat.go, zalo.go, mattermost.go, matrix.go, nostr.go, twitch.go, irc.go, twilio_voice.go |
| `pkg/channels/routing.go` | New | `RouteResolver` for multi-account outbound routing |
| `pkg/web/handlers_user_config.go` | Modified | Multi-account channel config API |
| `pkg/web/frontend/` | Modified | Multi-account UI in Settings > Channels |
| `go.mod` | Modified | 7 new Go module dependencies |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Config schema migration breaks existing `config.json` | Med | Custom UnmarshalJSON promotes old single-object to 1-element array transparently |
| Session key collision across accounts | Med | Account ID embedded in session key |
| WeChat requires server domain verification | High | Document setup requirement; adapter fails gracefully if webhook not reachable |
| Matrix E2E encryption state management | High | Deliver unencrypted rooms first; E2E deferred |
| Nostr private key exposure in config | High | Warn in logs + docs; recommend environment variable for nsec |
| Twilio webhook authentication | Med | Implement `X-Twilio-Signature` validation |

## Rollback Plan

1. Config schema is additive — old single-object form still parses via `UnmarshalJSON` fallback.
2. New channel adapters are gated by `Enabled: false` default.
3. `go.mod` additions: `go mod tidy` after rollback removes unused deps.
4. DB: no schema changes in this phase.

## Dependencies

- `github.com/line/line-bot-sdk-go/v8`
- `github.com/mattermost/mattermost-server/v6/model`
- `maunium.net/go/mautrix`
- `github.com/nbd-wtf/go-nostr`
- `github.com/gempir/go-twitch-irc/v4`
- `github.com/thoj/go-ircevent`
- `github.com/twilio/twilio-go`

## Success Criteria

- [ ] Existing single-account configs load without changes (backward compat test)
- [ ] Two Telegram accounts can be configured and each receives/sends messages independently
- [ ] All 9 new channel adapters compile and connect in happy-path smoke tests
- [ ] `isChannelsConfigEmpty()` returns correct results for both old and new config forms
- [ ] Routing resolves correct account for bound peers and falls back to default account
- [ ] `go test ./...` passes with no regressions
- [ ] Web UI shows multi-account lists and allows adding/removing accounts per channel

## Next Steps

- `sdd-spec` and `sdd-design` can run in parallel (both depend only on this proposal)
