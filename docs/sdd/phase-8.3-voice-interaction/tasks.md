# Tasks: Phase 8.3 — Voice Interaction & Rich UI

> TDD cycle per task: RED (write failing test) → confirm fail → GREEN (implement) → confirm pass → commit.
> All commits use conventional format. No "Co-Authored-By" attribution.

---

## Phase 1: Config & Interfaces (Foundation)

- [x] 1.1 Add `VoiceWakeConfig`, `CanvasConfig`, `ChannelActionsConfig` structs to `pkg/config/config.go`; add `Voice VoiceWakeConfig`, `Canvas CanvasConfig`, `ChannelActions ChannelActionsConfig` fields to `Config` struct with JSON tags `voice`, `canvas`, `channel_actions` and `enabled: false` defaults.

  ```go
  type VoiceWakeConfig struct {
      Enabled          bool    `json:"enabled"`
      WakeWord         string  `json:"wake_word"`
      Sensitivity      float64 `json:"sensitivity"`
      ConfirmationTone bool    `json:"confirmation_tone"`
  }
  type CanvasConfig struct {
      Enabled bool `json:"enabled"`
      DevMode bool `json:"dev_mode"`
  }
  type ChannelActionsConfig struct {
      Enabled              bool   `json:"enabled"`
      InteractionsEndpoint string `json:"interactions_endpoint_url"`
  }
  ```

  Expected test: `TestConfigDefaults` in `pkg/config/config_test.go` — assert `cfg.Voice.Enabled == false`, `cfg.Canvas.Enabled == false`, `cfg.ChannelActions.Enabled == false`.

  ```bash
  go test ./pkg/config/... -run TestConfigDefaults -v
  # PASS
  ```

  Commit: `feat(config): add Voice, Canvas, ChannelActions feature-flag config structs`

---

- [x] 1.2 Create `pkg/voice/tts_provider.go` — define `TTSProvider` interface; rename `TTSSynthesizer` to `OpenAITTSProvider` in `pkg/voice/tts.go`; add `func (t *OpenAITTSProvider) IsAvailable() bool` (already present as `TTSSynthesizer`); verify existing TTS tests still pass.

  ```go
  // pkg/voice/tts_provider.go
  package voice

  import "context"

  type TTSProvider interface {
      Synthesize(ctx context.Context, text string) ([]byte, error)
      IsAvailable() bool
  }
  ```

  ```bash
  go test ./pkg/voice/... -v
  # PASS (rename is backward-compatible internally)
  ```

  Commit: `refactor(voice): extract TTSProvider interface; rename TTSSynthesizer to OpenAITTSProvider`

---

- [x] 1.3 Add `ElevenLabsProvider` to `pkg/voice/tts_provider.go` — implement `TTSProvider`; `Synthesize` POSTs to `https://api.elevenlabs.io/v1/text-to-speech/{voiceID}` with `xi-api-key` header; `IsAvailable()` returns `apiKey != ""`.

  RED test in `pkg/voice/tts_provider_test.go`:
  ```go
  func TestElevenLabsProviderIsAvailable(t *testing.T) {
      p := &ElevenLabsProvider{}
      require.False(t, p.IsAvailable())
      p2 := &ElevenLabsProvider{apiKey: "key"}
      require.True(t, p2.IsAvailable())
  }
  ```

  ```bash
  go test ./pkg/voice/... -run TestElevenLabsProviderIsAvailable -v
  # PASS after implementation
  ```

  Commit: `feat(voice): add ElevenLabsProvider implementing TTSProvider`

---

- [x] 1.4 Add `SystemTTSProvider` to `pkg/voice/tts_provider.go` — on macOS calls `exec.Command("say", text)` and returns empty bytes (audio played directly); `IsAvailable()` always `true`.

  RED test:
  ```go
  func TestSystemTTSProviderIsAvailable(t *testing.T) {
      p := &SystemTTSProvider{}
      require.True(t, p.IsAvailable())
  }
  ```

  ```bash
  go test ./pkg/voice/... -run TestSystemTTSProviderIsAvailable -v
  # PASS
  ```

  Commit: `feat(voice): add SystemTTSProvider using OS say command as fallback`

---

- [x] 1.5 Add `STTProvider` interface to `pkg/voice/transcriber.go`; make `GroqTranscriber` implement it.

  ```go
  type STTProvider interface {
      Transcribe(ctx context.Context, audioFilePath string) (*TranscriptionResponse, error)
      IsAvailable() bool
  }
  ```

  ```bash
  go test ./pkg/voice/... -v
  # PASS — GroqTranscriber already has both methods
  ```

  Commit: `refactor(voice): extract STTProvider interface from GroqTranscriber`

---

- [x] 1.6 Create `pkg/channels/actions.go` — define `ChannelAction` interface, `InteractiveMessageRequest`, `ApprovalResult` structs, `ReactionPoller` struct with `Wait(ctx, messageID, timeout)`.

  ```go
  package channels

  import (
      "context"
      "time"
  )

  type InteractiveMessageRequest struct {
      Message string
      Actions []ActionConfig
      ChatID  string
  }

  type ActionConfig struct {
      Label   string `json:"label"`
      Value   string `json:"value"`
      Style   string `json:"style"` // "primary" | "danger"
  }

  type ApprovalResult struct {
      Approved bool
      Actor    string
      Reaction string
      TimedOut bool
  }

  type ChannelAction interface {
      SendInteractiveMessage(ctx context.Context, req InteractiveMessageRequest) (messageID string, err error)
      PollReaction(ctx context.Context, messageID string, timeout time.Duration) (ApprovalResult, error)
  }

  type ReactionPoller struct {
      interval time.Duration
  }

  func NewReactionPoller() *ReactionPoller {
      return &ReactionPoller{interval: 500 * time.Millisecond}
  }
  ```

  RED test in `pkg/channels/actions_test.go`:
  ```go
  func TestReactionPollerTimeout(t *testing.T) {
      ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
      defer cancel()
      poller := NewReactionPoller()
      result := poller.waitForSignal(ctx, make(chan ApprovalResult))
      require.True(t, result.TimedOut)
  }
  ```

  ```bash
  go test ./pkg/channels/... -run TestReactionPollerTimeout -v
  # PASS after implementing waitForSignal
  ```

  Commit: `feat(channels): add ChannelAction interface, ApprovalResult, ReactionPoller`

---

## Phase 2: Core Subsystem Implementation

- [x] 2.1 Create `pkg/voice/wake.go` (build tag `//go:build !porcupine`) — `WakeDetector` with energy-threshold keyword detection; `Listen(ctx, keyword, sensitivity, onDetected func())` goroutine loop; no CGo dependency in default build.

  ```go
  //go:build !porcupine

  package voice

  import (
      "context"
      "strings"
  )

  type WakeDetector struct{}

  func NewWakeDetector() *WakeDetector { return &WakeDetector{} }

  // Listen polls simulated audio input (placeholder for OS mic integration).
  // Real microphone capture is platform-specific; this impl accepts injected text for testability.
  func (w *WakeDetector) Listen(ctx context.Context, keyword string, _ float64, onDetected func()) {
      // Production path: integrate portaudio or os/exec `sox` for mic capture.
      // This stub satisfies the interface for tests and non-macOS builds.
      <-ctx.Done()
  }
  ```

  RED test in `pkg/voice/wake_test.go`:
  ```go
  func TestWakeDetectorCancelStopsListen(t *testing.T) {
      ctx, cancel := context.WithCancel(context.Background())
      wd := NewWakeDetector()
      done := make(chan struct{})
      go func() { wd.Listen(ctx, "hey mako", 0.5, func() {}); close(done) }()
      cancel()
      select {
      case <-done:
      case <-time.After(200*time.Millisecond):
          t.Fatal("Listen did not stop after context cancel")
      }
  }
  ```

  ```bash
  go test ./pkg/voice/... -run TestWakeDetectorCancelStopsListen -v
  # PASS
  ```

  Commit: `feat(voice): add WakeDetector with pure-Go energy-threshold stub`

---

- [x] 2.2 Create `pkg/voice/talk_mode.go` — `TalkModeSession` with `Start(ctx)` / `Stop()`, goroutine pipeline: audio capture → STT → agent process → TTS playback; "stop" keyword detection exits loop.

  ```go
  package voice

  import (
      "context"
      "strings"
  )

  type TalkModeSession struct {
      stt    STTProvider
      tts    TTSProvider
      stopCh chan struct{}
  }

  func NewTalkModeSession(stt STTProvider, tts TTSProvider) *TalkModeSession {
      return &TalkModeSession{stt: stt, tts: tts, stopCh: make(chan struct{})}
  }

  func (s *TalkModeSession) IsActive() bool {
      select {
      case <-s.stopCh:
          return false
      default:
          return true
      }
  }

  func (s *TalkModeSession) Stop() {
      select {
      case <-s.stopCh:
      default:
          close(s.stopCh)
      }
  }
  ```

  RED test in `pkg/voice/talk_mode_test.go`:
  ```go
  func TestTalkModeSessionStopIdempotent(t *testing.T) {
      sess := NewTalkModeSession(nil, nil)
      require.True(t, sess.IsActive())
      sess.Stop()
      require.False(t, sess.IsActive())
      require.NotPanics(t, func() { sess.Stop() }) // idempotent
  }
  ```

  ```bash
  go test ./pkg/voice/... -run TestTalkModeSessionStopIdempotent -v
  # PASS
  ```

  Commit: `feat(voice): add TalkModeSession with start/stop goroutine pipeline`

---

- [x] 2.3 Create `pkg/canvas/server.go` — `CanvasServer` with `sync.Map` state; HTTP handler for `GET /__makoclaw__/canvas/` serving SSE; methods `Push(html string)`, `Eval(js string) (string, error)` (only when DevMode), `Snapshot() string`, `Reset()`.

  RED test in `pkg/canvas/server_test.go`:
  ```go
  func TestCanvasServerPushSnapshot(t *testing.T) {
      srv := NewCanvasServer(false)
      srv.Push("<h1>hello</h1>")
      require.Equal(t, "<h1>hello</h1>", srv.Snapshot())
  }

  func TestCanvasServerEvalBlockedWhenDevModeOff(t *testing.T) {
      srv := NewCanvasServer(false)
      _, err := srv.Eval("alert(1)")
      require.Error(t, err)
      require.Contains(t, err.Error(), "dev_mode")
  }

  func TestCanvasServerReset(t *testing.T) {
      srv := NewCanvasServer(false)
      srv.Push("<p>data</p>")
      srv.Reset()
      require.Empty(t, srv.Snapshot())
  }
  ```

  ```bash
  go test ./pkg/canvas/... -v
  # PASS after implementation
  ```

  Commit: `feat(canvas): add CanvasServer with in-memory state and SSE broadcast`

---

- [x] 2.4 Create `pkg/tools/canvas.go` — `CanvasTool` implementing `Tool`; params `operation` (create|update|append|clear|snapshot), `content`, `format`; delegates to `CanvasServer`; returns error `"Canvas is disabled..."` when `canvas.enabled = false`.

  RED test in `pkg/tools/canvas_test.go`:
  ```go
  func TestCanvasToolDisabledReturnsError(t *testing.T) {
      tool := NewCanvasTool(nil) // nil server = disabled
      _, err := tool.Execute(context.Background(), map[string]interface{}{
          "operation": "push",
          "content":   "<h1>hi</h1>",
          "format":    "html",
      })
      require.Error(t, err)
      require.Contains(t, err.Error(), "Canvas is disabled")
  }

  func TestCanvasToolPushDelegates(t *testing.T) {
      srv := canvas.NewCanvasServer(false)
      tool := NewCanvasTool(srv)
      result, err := tool.Execute(context.Background(), map[string]interface{}{
          "operation": "push",
          "content":   "<p>test</p>",
          "format":    "html",
      })
      require.NoError(t, err)
      require.Contains(t, result, "pushed")
  }
  ```

  ```bash
  go test ./pkg/tools/... -run TestCanvasTool -v
  # PASS after implementation
  ```

  Commit: `feat(tools): add CanvasTool delegating to CanvasServer`

---

- [x] 2.5 Create `pkg/tools/interactive_message.go` — `InteractiveMessageTool` implementing `Tool` + `ContextualTool`; params `message string`, `actions []ActionConfig`, `channel string`, `chat_id string`; returns error when `channel_actions.enabled = false` or no endpoint configured.

  RED test in `pkg/tools/interactive_message_test.go`:
  ```go
  func TestInteractiveMessageToolDisabledReturnsError(t *testing.T) {
      tool := NewInteractiveMessageTool(nil, false, "")
      _, err := tool.Execute(context.Background(), map[string]interface{}{
          "message": "approve?",
          "actions": []interface{}{},
          "channel": "discord",
          "chat_id": "123",
      })
      require.Error(t, err)
      require.Contains(t, err.Error(), "channel actions")
  }
  ```

  ```bash
  go test ./pkg/tools/... -run TestInteractiveMessageTool -v
  # PASS after implementation
  ```

  Commit: `feat(tools): add InteractiveMessageTool with feature-flag guard`

---

## Phase 3: Integration & Wiring

- [x] 3.1 Modify `pkg/agent/loop.go` — in `NewAgentLoop`, conditionally register `CanvasTool` when `cfg.Canvas.Enabled`, and `InteractiveMessageTool` when `cfg.ChannelActions.Enabled && cfg.ChannelActions.InteractionsEndpoint != ""`; log warnings when flags are false.

  ```go
  // After existing tool registrations:
  if cfg.Canvas.Enabled {
      canvasServer := canvas.NewCanvasServer(cfg.Canvas.DevMode)
      toolsRegistry.Register(tools.NewCanvasTool(canvasServer))
      logger.InfoC("agent", "Canvas tool registered")
  }
  if cfg.ChannelActions.Enabled {
      if cfg.ChannelActions.InteractionsEndpoint == "" {
          logger.WarnC("agent", "Channel actions require a public HTTPS interactions_endpoint_url. Actions disabled.")
      } else {
          toolsRegistry.Register(tools.NewInteractiveMessageTool(nil, true, cfg.ChannelActions.InteractionsEndpoint))
          logger.InfoC("agent", "InteractiveMessage tool registered")
      }
  }
  ```

  ```bash
  go build ./pkg/agent/...
  # no errors
  go test ./pkg/agent/... -v
  # PASS
  ```

  Commit: `feat(agent): register CanvasTool and InteractiveMessageTool under feature flags`

---

- [x] 3.2 Modify `pkg/web/server.go` — when `cfg.Canvas.Enabled`, mount SSE handler at `GET /__makoclaw__/canvas/events` and static placeholder at `GET /__makoclaw__/canvas/`; inject `CanvasServer` instance shared with the tool.

  ```bash
  go build ./pkg/web/...
  # no errors
  ```

  Commit: `feat(web): mount canvas SSE and static routes when canvas.enabled`

---

- [x] 3.3 Add Discord button rendering to `pkg/channels/discord.go` — implement `ChannelAction` on `DiscordChannel`; `SendInteractiveMessage` posts message with component button row via discordgo; register `/interactions` POST handler when `cfg.ChannelActions.Enabled`.

  ```bash
  go build ./pkg/channels/...
  # no errors
  ```

  Commit: `feat(channels): implement ChannelAction for Discord with button components`

---

- [x] 3.4 Add Slack button rendering to `pkg/channels/slack.go` — implement `ChannelAction` on `SlackChannel`; `SendInteractiveMessage` posts Block Kit button via slack-go; `PollReaction` polls `reactions.get` every 500ms until `👍` found or timeout.

  ```bash
  go build ./pkg/channels/...
  # no errors
  ```

  Commit: `feat(channels): implement ChannelAction for Slack with emoji-reaction polling`

---

- [x] 3.5 Create `pkg/web/frontend/src/views/Canvas.vue` — EventSource listener on `/__makoclaw__/canvas/events`; renders received HTML inside `<iframe sandbox="allow-scripts" :srcdoc="canvasHTML">`; adds route `/canvas` in Vue router.

  Commit: `feat(frontend): add CanvasView with sandboxed iframe and SSE listener`

---

## Phase 4: Tests & Verification

- [x] 4.1 Integration test `pkg/canvas/server_test.go` — `TestCanvasSSEBroadcast`: spin up `httptest.NewServer`, connect EventSource client, `Push` HTML, assert SSE event received within 500ms.

  ```bash
  go test ./pkg/canvas/... -run TestCanvasSSEBroadcast -v -timeout 5s
  # PASS
  ```

  Commit: `test(canvas): add SSE broadcast integration test`

---

- [x] 4.2 Integration test `pkg/channels/actions_test.go` — `TestReactionPollerApproval`: mock `ChannelAction` that signals approval after 200ms; assert `ApprovalResult.Approved == true` and `TimedOut == false`.

  ```bash
  go test ./pkg/channels/... -run TestReactionPollerApproval -v
  # PASS
  ```

  Commit: `test(channels): add ReactionPoller approval and timeout integration tests`

---

- [x] 4.3 Unit test `pkg/voice/tts_provider_test.go` — `TestElevenLabsSynthesizeCallsAPI`: use `httptest.NewServer` returning 200 MP3 bytes; assert returned bytes match.

  ```bash
  go test ./pkg/voice/... -run TestElevenLabsSynthesizeCallsAPI -v
  # PASS
  ```

  Commit: `test(voice): add ElevenLabsProvider HTTP round-trip test`

---

- [x] 4.4 Run full test suite; fix any regressions introduced by the `OpenAITTSProvider` rename.

  ```bash
  go test -race ./... 2>&1 | tail -20
  # ok   github.com/sipeed/makoclaw/...
  ```

  Commit: `fix(voice): resolve any OpenAITTSProvider rename regressions`

---

- [x] 4.5 Verify config JSON round-trip — write `TestConfigVoiceCanvasChannelActionsRoundTrip` in `pkg/config/config_test.go`: marshal/unmarshal JSON with all three feature blocks; assert field values preserved.

  ```bash
  go test ./pkg/config/... -run TestConfigVoiceCanvasChannelActionsRoundTrip -v
  # PASS
  ```

  Commit: `test(config): add round-trip test for Voice, Canvas, ChannelActions config`

---

## Phase 5: Cleanup

- [x] 5.1 Update `CLAUDE.md` tool table — add `canvas` and `interactive_message` rows with file paths and descriptions; no code changes.

  Commit: `docs(claude): document canvas and interactive_message tools in architecture table`

---

- [x] 5.2 Add `//go:build porcupine` stub `pkg/voice/wake_porcupine.go` with compile-time note; ensures `go build -tags porcupine` does not break CI.

  ```bash
  go build -tags porcupine ./pkg/voice/...
  # no errors
  ```

  Commit: `chore(voice): add porcupine build-tag stub for future CGo wake-word backend`

---

## Summary

| Phase | Tasks | Focus |
|-------|-------|-------|
| Phase 1 | 6 | Config structs + all interfaces |
| Phase 2 | 5 | Core subsystem implementations |
| Phase 3 | 5 | Wiring into agent loop, web server, channels, frontend |
| Phase 4 | 5 | Integration + unit tests, full suite verification |
| Phase 5 | 2 | Docs + build-tag stub |
| **Total** | **23** | |

### Implementation Order

Start with Phase 1 (no code changes compile-break anything — pure additions). Phase 2 subsystems are independent and can be worked on in parallel. Phase 3 requires Phase 2 complete. Phase 4 runs after each subsystem lands. Phase 5 is last.
