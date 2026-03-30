# Proposal: Phase 8.3 — Voice & Interaction

**Change**: `phase-8.3-voice-interaction`
**Status**: Draft
**Inspired by**: OpenClaw (https://github.com/openclaw/openclaw)
**Date**: 2026-03-30

---

## Intent

MakoClaw currently supports only text-based interaction across channels. This phase adds three interaction modalities: voice-driven conversation (wake word + Talk Mode), an agent-controlled visual canvas (A2UI), and rich interactive channel actions (buttons, modals, reaction approvals). These close the gap between MakoClaw and modern agent UX patterns seen in consumer AI products.

## Scope

### In Scope
- **Voice Wake + Talk Mode**: wake word detection, continuous voice conversation, ElevenLabs TTS + system TTS fallback, Deepgram STT + existing Groq Whisper extension, macOS platform first
- **Canvas / A2UI**: HTTP-served HTML/CSS/JS workspace at `/__makoclaw__/canvas/`, agent-controllable via `push`/`eval`/`snapshot`/`reset` commands
- **Channel Action Framework**: Discord buttons/select menus/modals, Slack button blocks/slash command responses, emoji reactions as approval mechanism, exec-approval overlays

### Out of Scope
- iOS/Android native voice nodes (deferred post-macOS)
- Canvas persistence/versioning across sessions
- QQ, DingTalk, Feishu interactive components
- Voice synthesis for channel messages (text channels remain text-only)

## Approach

Three parallel subsystems added to existing infrastructure:

1. **Voice**: New `pkg/voice/wake.go` + `pkg/voice/tts.go` extending existing Groq Whisper (`pkg/voice/`). Wake word config at `~/.MakoClaw/settings/voicewake.json`. Talk Mode runs as a dedicated goroutine with STT→agent loop→TTS pipeline.

2. **Canvas**: New `pkg/canvas/` package. Gateway mounts `/__makoclaw__/canvas/` route. Agent gets a `canvas` tool (`push`, `eval`, `snapshot`, `reset`). Vue frontend adds a Canvas view for local display.

3. **Channel Actions**: Extend `pkg/channels/discord.go` and `pkg/channels/slack.go` with interaction webhooks. New `pkg/channels/actions.go` defines `ChannelAction` interface. Agent gets an `interactive_message` tool. Emoji reaction polling added to applicable channels.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/voice/` | Modified + New | Add `wake.go`, `tts.go`; extend existing transcription |
| `pkg/canvas/` | New | Canvas host package: serve, push, eval, snapshot, reset |
| `pkg/channels/discord.go` | Modified | Add interaction webhook handler, button/modal support |
| `pkg/channels/slack.go` | Modified | Add button blocks, slash command interactive responses |
| `pkg/channels/actions.go` | New | `ChannelAction` interface + emoji reaction approval loop |
| `pkg/tools/canvas.go` | New | `canvas` tool registered in agent loop |
| `pkg/tools/interactive_message.go` | New | `interactive_message` tool for channel actions |
| `pkg/agent/loop.go` | Modified | Register canvas + interactive_message tools |
| `pkg/web/server.go` | Modified | Mount `/__makoclaw__/canvas/` route |
| `pkg/web/frontend/` | Modified | Add Canvas view |
| `pkg/config/config.go` | Modified | Add `VoiceConfig`, `CanvasConfig` config structs |
| `~/.MakoClaw/settings/voicewake.json` | New (runtime) | Wake word configuration file |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Wake word false positives trigger agent unintentionally | Med | Configurable sensitivity threshold + activation confirmation tone |
| ElevenLabs API latency degrades Talk Mode UX | Med | System TTS fallback; async streaming where supported |
| Canvas `eval` arbitrary JS execution is a security surface | High | Sandbox iframe with CSP; eval only allowed in explicit "dev mode" |
| Discord/Slack interaction webhooks require public HTTPS endpoint | Med | Document ngrok/tunnel requirement; skip gracefully if not configured |
| Scope creep across 3 parallel subsystems | Med | Each subsystem has independent feature flag; can ship incrementally |

## Rollback Plan

Each subsystem is gated by config flags (`voice.enabled`, `canvas.enabled`, `channel_actions.enabled`). Setting these to `false` disables the subsystem entirely without code removal. No schema migrations required for initial implementation — canvas state is in-memory only.

## Dependencies

- ElevenLabs API key (optional; falls back to system TTS)
- Deepgram API key (optional; falls back to existing Groq Whisper)
- Discord/Slack apps must have `interactions_endpoint_url` configured for channel actions
- macOS only for voice wake (initial); requires `pkg/voice` platform build tags

## Success Criteria

- [ ] Wake word triggers agent response within 1s on macOS
- [ ] Talk Mode sustains 5-turn voice conversation without manual intervention
- [ ] Agent can push HTML to canvas and it renders in frontend within 500ms
- [ ] Discord message with approval button correctly gates `exec` tool execution
- [ ] Emoji reaction on Slack message correctly signals approval to waiting agent
- [ ] All three subsystems independently disableable via config with no side effects

## Next Steps

- `sdd-spec` and `sdd-design` can run in parallel (both depend only on this proposal)
