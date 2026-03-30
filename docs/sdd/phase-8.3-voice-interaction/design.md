# Design: Phase 8.3 — Voice & Interaction

## Technical Approach

Three independent subsystems (Voice, Canvas, Channel Actions) are added as opt-in packages gated by feature flags in `Config`. Each follows the existing pattern: new struct in relevant package, registered as `Tool` in `pkg/agent/loop.go`, config fields added to `pkg/config/config.go`. No existing APIs are broken.

## Architecture Decisions

| Decision | Choice | Rejected | Rationale |
|----------|--------|----------|-----------|
| TTS abstraction | New `TTSProvider` interface wrapping existing `TTSSynthesizer` | Extend existing struct | Existing `TTSSynthesizer` is OpenAI-only; ElevenLabs needs a separate HTTP client |
| Wake word detection | Pure-Go energy threshold + keyword match via `porcupine` CGo binding (macOS only via build tag) | Bundled native binary | Build tags isolate platform dep; no subprocess needed |
| Canvas state | In-memory `sync.Map` in `pkg/canvas/server.go` | SQLite | Proposal explicitly excludes persistence; in-memory is sufficient and simpler |
| Emoji approval | Polling loop in `pkg/channels/actions.go` (500ms interval, configurable timeout) | Webhook callback | Discord/Slack webhook for reactions requires public endpoint; polling works without it |
| Canvas `eval` security | Sandboxed `<iframe sandbox="allow-scripts">` + CSP; eval blocked unless `canvas.dev_mode=true` | Open iframe | Proposal identifies this as High risk; sandbox is the minimal safe default |

## Data Flow

### Voice / Talk Mode

```
Microphone (OS) ──► WakeDetector.Listen()
                         │ wake word detected
                         ▼
                   TalkModeSession.Start()
                         │
              ┌──────────┴──────────────┐
              │  audio capture goroutine │
              └──────────┬──────────────┘
                         │ audio bytes
                         ▼
              STTProvider.Transcribe()   (Groq or Deepgram)
                         │ text
                         ▼
              AgentLoop.ProcessMessage()
                         │ response text
                         ▼
              TTSProvider.Synthesize()   (ElevenLabs or system)
                         │ audio bytes
                         ▼
                   OS audio playback
```

### Canvas

```
Agent ──► CanvasTool.Execute(push|eval|snapshot|reset)
                │
                ▼
          CanvasServer (pkg/canvas/)
          sync.Map state  ◄──► SSE broadcast
                                    │
                                    ▼
                          GET /__makoclaw__/canvas/
                          (Vue frontend EventSource listener)
```

### Channel Actions

```
Agent ──► InteractiveMessageTool.Execute()
                │ renders button block / modal
                ▼
          ChannelAction.Send()   (Discord or Slack impl)
                │
                ▼
          ReactionPoller.Wait()  ←── 500ms tick
                │ ✓ / ✗
                ▼
          returns approval result to agent
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/config/config.go` | Modify | Add `VoiceWakeConfig`, `CanvasConfig`; extend `Config` struct with `Voice`, `Canvas`, `ChannelActions` fields |
| `pkg/voice/wake.go` | Create | `WakeDetector` struct + `VoiceWakeConfig`; macOS build tag |
| `pkg/voice/tts_provider.go` | Create | `TTSProvider` interface; `ElevenLabsProvider` and `SystemTTSProvider` impls |
| `pkg/voice/tts.go` | Modify | Existing `TTSSynthesizer` implements `TTSProvider` (rename to `OpenAITTSProvider`) |
| `pkg/voice/talk_mode.go` | Create | `TalkModeSession` struct, goroutine pipeline |
| `pkg/canvas/server.go` | Create | `CanvasServer` — HTTP handler, SSE broadcaster, in-memory state |
| `pkg/tools/canvas.go` | Create | `CanvasTool` — implements `Tool`; delegates to `CanvasServer` |
| `pkg/tools/interactive_message.go` | Create | `InteractiveMessageTool` — implements `Tool` + `ContextualTool` |
| `pkg/channels/actions.go` | Create | `ChannelAction` interface, `ReactionPoller`, `ApprovalRequest`/`ApprovalResult` |
| `pkg/channels/discord.go` | Modify | Add interaction webhook handler; implement `ChannelAction` for buttons/modals |
| `pkg/channels/slack.go` | Modify | Add button block rendering; implement `ChannelAction` for Slack |
| `pkg/agent/loop.go` | Modify | Register `CanvasTool` and `InteractiveMessageTool` when flags enabled |
| `pkg/web/server.go` | Modify | Mount `/__makoclaw__/canvas/` route; inject `CanvasServer` |
| `pkg/web/frontend/src/views/Canvas.vue` | Create | Vue view — EventSource listener, renders canvas HTML in sandboxed iframe |

## Interfaces / Contracts

```go
// pkg/voice/tts_provider.go
type TTSProvider interface {
    Synthesize(ctx context.Context, text string) ([]byte, error)
    IsAvailable() bool
}

// pkg/voice/wake.go
type VoiceWakeConfig struct {
    Enabled        bool    `json:"enabled"`
    WakeWord       string  `json:"wake_word"`        // e.g. "hey mako"
    Sensitivity    float64 `json:"sensitivity"`      // 0.0–1.0
    ConfirmationTone bool  `json:"confirmation_tone"`
}

type TalkModeSession struct {
    stt      STTProvider        // GroqTranscriber or DeepgramTranscriber
    tts      TTSProvider
    loop     *agent.AgentLoop
    stopCh   chan struct{}
    audioOut chan []byte
}

// pkg/channels/actions.go
type ChannelAction interface {
    SendInteractiveMessage(ctx context.Context, req InteractiveMessageRequest) (messageID string, err error)
    PollReaction(ctx context.Context, messageID string, timeout time.Duration) (ApprovalResult, error)
}

type ApprovalResult struct {
    Approved  bool
    Actor     string
    Reaction  string
    TimedOut  bool
}

// pkg/canvas/server.go
// Commands: push (replace HTML), eval (inject JS), snapshot (return HTML), reset (clear)
type CanvasCommand struct {
    Op      string `json:"op"`      // push|eval|snapshot|reset
    Payload string `json:"payload"` // HTML or JS
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `TTSProvider` impls, `CanvasTool.Execute()`, `VoiceWakeConfig` parsing | Table-driven tests in `_test.go` files |
| Integration | `CanvasServer` SSE broadcast, `ReactionPoller` timeout behavior | `httptest.NewServer`, mock `ChannelAction` |
| E2E | Canvas renders in frontend, approval gate blocks `exec` | Playwright; mock Discord/Slack webhook |

## Migration / Rollout

Feature flags in `config.json`:
```json
"voice":           { "enabled": false },
"canvas":          { "enabled": false },
"channel_actions": { "enabled": false }
```

All three default to `false`. No schema migrations. Canvas state is ephemeral (in-memory). Subsystems can be enabled independently and disabled without redeployment.

## Open Questions

- [ ] `porcupine` CGo dependency — confirm license is acceptable for MIT project or choose pure-Go alternative (e.g. energy-threshold-only detection)
- [ ] Deepgram STT integration path: extend `STTProvider` interface alongside existing `GroqTranscriber`, or make Deepgram a config-selected swap?
- [ ] System TTS on macOS: use `say` command via `exec.Command`; confirm this is acceptable given `ExecTool` deny-pattern constraints
