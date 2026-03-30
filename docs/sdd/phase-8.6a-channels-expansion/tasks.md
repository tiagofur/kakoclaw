# Tasks: Phase 8.6a — Channels Expansion

## Phase 1: Foundation — Config Schema & Dependencies

- [ ] 1.1 Add `AccountID string \`json:"account_id"\`` field to all 10 existing channel config structs in `pkg/config/config.go` (`TelegramConfig`, `WhatsAppConfig`, `FeishuConfig`, `DiscordConfig`, `MaixCamConfig`, `QQConfig`, `DingTalkConfig`, `SlackConfig`, `SignalConfig`, `EmailChannelConfig`).
- [ ] 1.2 Add `MultiAccountConfig[T any]` generic struct with `Accounts []T` and `UnmarshalJSON` that tries `[]T` first, falls back to single `T` promoted to `[]T{single}` with `AccountID` defaulted to `"default"` when empty — in `pkg/config/config.go` after `FlexibleStringSlice`.
- [ ] 1.3 Update `ChannelsConfig` in `pkg/config/config.go`: change all 10 existing fields from `XConfig` to `MultiAccountConfig[XConfig]`; add 9 new fields (`LINE`, `WeChat`, `Zalo`, `Mattermost`, `Matrix`, `Nostr`, `Twitch`, `IRC`, `TwilioVoice`) as `MultiAccountConfig[XConfig]` with new config structs (`LINEConfig`, `WeChatConfig`, `ZaloConfig`, `MattermostConfig`, `MatrixConfig`, `NostrConfig`, `TwitchConfig`, `IRCConfig`, `TwilioVoiceConfig`).
- [ ] 1.4 Write failing test `TestMultiAccountConfig_UnmarshalJSON` in `pkg/config/config_test.go` — table-driven: single-object Telegram JSON → 1-element `Accounts`; array JSON → 2-element `Accounts`; missing `account_id` → defaults to `"default"`. Run: `go test ./pkg/config/... -run TestMultiAccountConfig -v` → expect compile/test failure.
- [ ] 1.5 Implement `MultiAccountConfig.UnmarshalJSON` (task 1.2) and `AccountID` fields (task 1.1) until `go test ./pkg/config/... -run TestMultiAccountConfig -v` passes.
- [ ] 1.6 Update `isChannelsConfigEmpty` in `pkg/config/config.go`: replace per-field `.Enabled` checks with `anyEnabled(c.Telegram.Accounts) || anyEnabled(c.WhatsApp.Accounts) || ...` for all 19 channels using a helper `func anyEnabled[T interface{ isEnabled() bool }](accounts []T) bool`. Run: `go test ./pkg/config/... -v` → all pass.
- [ ] 1.7 Add 7 new `require` entries to `go.mod` and run `go mod tidy`:
  - `github.com/line/line-bot-sdk-go/v8`
  - `github.com/mattermost/mattermost-server/v6/model`
  - `maunium.net/go/mautrix`
  - `github.com/nbd-wtf/go-nostr`
  - `github.com/gempir/go-twitch-irc/v4`
  - `github.com/thoj/go-ircevent`
  - `github.com/twilio/twilio-go`
- [ ] 1.8 Commit: `feat(config): add MultiAccountConfig generic + 9 new channel config structs`

## Phase 2: Core Infrastructure — BaseChannel & Routing

- [ ] 2.1 Write failing test `TestBaseChannel_HandleMessage_SessionKey_WithAccountID` in `pkg/channels/base_test.go`: construct `BaseChannel` with name `"whatsapp"` and `accountID "personal"`; call `HandleMessage("sender1","chat1","hi",nil,map[string]string{"account_id":"personal"})`; capture published `InboundMessage` via a mock bus; assert `SessionKey == "whatsapp:personal:chat1"` and `Metadata["account_id"] == "personal"`. Run → fail.
- [ ] 2.2 Modify `BaseChannel` in `pkg/channels/base.go`: add `accountID string` field; update `NewBaseChannel` signature to accept `accountID string`; update `HandleMessage` to read `account_id` from incoming `metadata`, write it to `msg.Metadata["account_id"]`, and build `sessionKey = fmt.Sprintf("%s:%s:%s", c.name, c.accountID, chatID)` when `c.accountID != ""`. Run `go test ./pkg/channels/... -run TestBaseChannel_HandleMessage_SessionKey -v` → pass.
- [ ] 2.3 Write failing tests `TestRouteResolver_Resolve` in `pkg/channels/routing_test.go` (new file): 3 cases — explicit `account_id` in metadata hits correct channel; no `account_id` falls back to `"default"` key; no `account_id` and no `"default"` uses first prefix match. Run → compile error (file missing).
- [ ] 2.4 Create `pkg/channels/routing.go` with `RouteResolver` struct and `Resolve(msg bus.OutboundMessage, channels map[string]Channel) (Channel, error)` implementing the 3-step priority from design. Run `go test ./pkg/channels/... -run TestRouteResolver -v` → pass.
- [ ] 2.5 Update `pkg/channels/manager.go`: `initChannels()` iterates each `cfg.Channels.X.Accounts` slice; for each account, derives key as `fmt.Sprintf("%s:%s", channelName, account.AccountID)`; registers under that key. Update `dispatchOutbound` to use `RouteResolver.Resolve` instead of direct map lookup. Update `RestartChannel` to accept the `channel:account_id` key format.
- [ ] 2.6 Run `go test ./pkg/channels/... -v` → all pass. Commit: `feat(channels): multi-account BaseChannel, RouteResolver, manager iteration`

## Phase 3: New Channel Adapters — Batch A (LINE, WeChat, Zalo, Mattermost)

- [ ] 3.1 Create `pkg/channels/line.go`: `LINEConfig` struct (fields: `Enabled`, `AccountID`, `ChannelSecret`, `ChannelAccessToken`, `AllowFrom FlexibleStringSlice`); `LINEChannel` struct embedding `*BaseChannel` with `linebot.Client`; `NewLINEChannel(cfg LINEConfig, bus) (*LINEChannel, error)` returns error if `ChannelAccessToken == ""`; `Start` registers HTTP webhook at `/webhook/line/{AccountID}` using `linebot.ParseRequest`; `Send` calls `linebot.NewTextMessage` + `ReplyMessage` or `PushMessage`; `Stop` deregisters the webhook. Smoke test: `NewLINEChannel(LINEConfig{ChannelAccessToken:""}, nil)` returns non-nil error.
- [ ] 3.2 Create `pkg/channels/wechat.go`: `WeChatConfig` (fields: `Enabled`, `AccountID`, `AppID`, `AppSecret`, `Token`, `EncodingAESKey`, `AllowFrom`); HTTP webhook at `/webhook/wechat/{AccountID}` that validates signature (`sha1(token+timestamp+nonce)`) and processes XML `<xml>` messages; `Send` calls WeChat Customer Service API via `net/http`; returns error on bad `Token` at startup.
- [ ] 3.3 Create `pkg/channels/zalo.go`: `ZaloConfig` (fields: `Enabled`, `AccountID`, `OAAccessToken`, `AppSecretKey`, `AllowFrom`); HTTP webhook at `/webhook/zalo/{AccountID}` validating HMAC-SHA256 `mac` param; processes `user_send_text` events; `Send` calls Zalo OA API `sendmessage` endpoint; returns error if `OAAccessToken == ""`.
- [ ] 3.4 Create `pkg/channels/mattermost.go`: `MattermostConfig` (fields: `Enabled`, `AccountID`, `ServerURL`, `BotToken`, `TeamName`, `Channels []string`, `AllowFrom`); uses `mattermost-server/v6/model.NewAPIv4Client`; `Start` connects via WebSocket driver, joins configured channels; `Send` calls `CreatePost`; returns error if `BotToken == ""` or `ServerURL == ""`.
- [ ] 3.5 Register LINE, WeChat, Zalo, Mattermost in `initChannels()` in `pkg/channels/manager.go` (iterate `cfg.Channels.LINE.Accounts`, etc.).
- [ ] 3.6 Run `go build ./pkg/channels/...` → no errors. Commit: `feat(channels): add LINE, WeChat, Zalo, Mattermost adapters`

## Phase 4: New Channel Adapters — Batch B (Matrix, Nostr, Twitch, IRC, Twilio Voice)

- [ ] 4.1 Create `pkg/channels/matrix.go`: `MatrixConfig` (fields: `Enabled`, `AccountID`, `HomeserverURL`, `AccessToken`, `UserID`, `Rooms []string`, `AllowFrom`); uses `mautrix-go` client; `Start` calls `client.Sync()` in goroutine with exponential backoff (3 retries, doubles from 2s); handles `m.room.message` events; `Send` calls `client.SendText`; returns error if `AccessToken == ""`.
- [ ] 4.2 Create `pkg/channels/nostr.go`: `NostrConfig` (fields: `Enabled`, `AccountID`, `Nsec`, `Relays []string`, `AllowFrom`); `Start` parses `nsec` with `nostr.GetPublicKey`/`nostr.GeneratePrivateKey` — returns error if parse fails; subscribes to NIP-04 DM events on configured relays; decrypts with `nip04.Decrypt`; `Send` encrypts with `nip04.Encrypt`, publishes kind-4 event.
- [ ] 4.3 Create `pkg/channels/twitch.go`: `TwitchConfig` (fields: `Enabled`, `AccountID`, `Username`, `OAuthToken`, `Channels []string`, `AllowFrom`); uses `go-twitch-irc/v4`; `Start` calls `client.Connect()` in goroutine; handles `OnPrivateMessage` filtering `@BotName` mentions or direct whispers; `Send` calls `client.Say`; returns error if `OAuthToken == ""`.
- [ ] 4.4 Create `pkg/channels/irc.go`: `IRCConfig` (fields: `Enabled`, `AccountID`, `Server`, `Port`, `Nick`, `Password`, `Channels []string`, `TLS bool`, `AllowFrom`); uses `go-ircevent`; `Start` calls `conn.Connect()`; handles `PRIVMSG` for direct messages and nick-prefixed messages; handles `KICK` event with rejoin backoff; handles 474 error (banned) with log + no-retry; `Send` calls `conn.Privmsg`; returns error if `Server == ""`.
- [ ] 4.5 Create `pkg/channels/twilio_voice.go`: `TwilioVoiceConfig` (fields: `Enabled`, `AccountID`, `AccountSID`, `AuthToken`, `WebhookURL`, `AllowFrom`); exposes HTTP handler at `/webhook/twilio/{AccountID}`; validates `X-Twilio-Signature` using `twilio-go`'s `client.ValidateRequest`; on valid request, passes `CallSid` + `SpeechResult` (or `Digits`) to agent loop; returns TwiML `<Response><Say>{response}</Say></Response>`; HTTP 403 on invalid signature. Returns error if `AccountSID == ""` or `AuthToken == ""`.
- [ ] 4.6 Register Matrix, Nostr, Twitch, IRC, TwilioVoice in `initChannels()` in `pkg/channels/manager.go`.
- [ ] 4.7 Run `go build ./pkg/channels/...` → no errors. Commit: `feat(channels): add Matrix, Nostr, Twitch, IRC, TwilioVoice adapters`

## Phase 5: Web API & Integration Tests

- [ ] 5.1 Update `pkg/web/handlers_user_config.go`: channel GET/PUT handlers that previously returned/accepted single objects must now return/accept arrays (or `MultiAccountConfig[T]`). Ensure JSON marshaling of `MultiAccountConfig` emits the `Accounts` array correctly — add `MarshalJSON` to emit `[]T` directly (not wrapped in `{"accounts":[...]}`).
- [ ] 5.2 Write integration test `TestManager_MultiAccountRegistration` in `pkg/channels/manager_test.go` (new file): build a `Config` with two Telegram accounts (`account_id: "acc1"` and `account_id: "acc2"`, both `Enabled: true`, valid mock tokens); create `Manager`; assert `m.channels["telegram:acc1"]` and `m.channels["telegram:acc2"]` both exist and are non-nil. Run → fail, then implement until pass.
- [ ] 5.3 Write test `TestIsChannelsConfigEmpty_MultiAccount` in `pkg/config/config_test.go`: assert `isChannelsConfigEmpty` returns `false` for a config with one Telegram account enabled, `true` for config with all accounts disabled. Run → pass.
- [ ] 5.4 Write smoke tests `TestNewXChannel_InvalidToken` for LINE, Matrix, Nostr, Twilio in `pkg/channels/adapters_smoke_test.go` (new file): pass empty/invalid credentials, assert `err != nil` and no panic.
- [ ] 5.5 Write test `TestBaseChannel_HandleMessage_SessionKey_NoAccountID` in `pkg/channels/base_test.go`: when `accountID == ""`, session key MUST be `"channel:chatID"` (legacy format preserved). Run → pass.
- [ ] 5.6 Run `go test ./... -race` → all pass. Commit: `test(channels): integration + smoke tests for multi-account and new adapters`

## Phase 6: Cleanup

- [ ] 6.1 Remove any `TODO` / placeholder comments left in adapter files; ensure each adapter logs startup success with `logger.InfoC("channel", "{Name} channel enabled")` consistent with existing adapters.
- [ ] 6.2 Update `pkg/config/config.go` `DefaultConfig()` / `getEmptyChannelsConfig()` to include default (zero-value) entries for all 9 new `MultiAccountConfig` fields so JSON marshaling produces clean output.
- [ ] 6.3 Verify `go vet ./...` and `go build ./...` clean. Commit: `chore(channels): cleanup adapters and defaults`
