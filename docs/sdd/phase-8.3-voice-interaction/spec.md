# Phase 8.3 — Voice & Interaction Specification

**Change**: `phase-8.3-voice-interaction`
**Status**: Draft
**Date**: 2026-03-30

---

## Domain 1: Voice Wake

### Requirement: Wake Word Detection

The system MUST detect a configured wake word and trigger an agent response cycle. Detection MUST occur within 1 second on macOS. Sensitivity threshold MUST be configurable via `~/.MakoClaw/settings/voicewake.json`.

#### Scenario: Wake word detected — ElevenLabs available

- GIVEN `voice.enabled = true` and ElevenLabs API key is configured
- WHEN the configured wake word is spoken and detected above the sensitivity threshold
- THEN the system plays an activation tone, begins STT capture, and routes the utterance to the agent
- AND the agent's text response is synthesized via ElevenLabs TTS and played back

#### Scenario: Wake word detected — ElevenLabs absent

- GIVEN `voice.enabled = true` and no ElevenLabs API key is present
- WHEN the configured wake word is detected
- THEN the system falls back to the OS system TTS engine for all speech synthesis
- AND the agent response is still spoken aloud without error

#### Scenario: Voice subsystem disabled

- GIVEN `voice.enabled = false`
- WHEN any audio input is received
- THEN wake word detection does not activate and no agent cycle is triggered

---

## Domain 2: Talk Mode

### Requirement: Continuous STT→Agent→TTS Loop

The system MUST sustain a continuous voice conversation loop when Talk Mode is active. The loop MUST process each utterance end-to-end and restart listening automatically until terminated.

#### Scenario: Talk Mode activated — successful conversation turn

- GIVEN Talk Mode is active
- WHEN the user speaks a complete utterance
- THEN STT transcribes it, the agent processes it, and TTS plays the response
- AND the system immediately begins listening for the next utterance

#### Scenario: Stop command terminates Talk Mode

- GIVEN Talk Mode is active and the STT loop is running
- WHEN the user says "stop"
- THEN the system exits Talk Mode, stops the STT loop, and returns to idle
- AND no further audio capture occurs until Talk Mode is re-activated

#### Scenario: Talk Mode — STT provider fallback

- GIVEN Talk Mode is active and Deepgram API key is absent
- WHEN an utterance is captured
- THEN the system falls back to Groq Whisper for transcription
- AND the conversation loop continues without interruption

---

## Domain 3: Canvas / A2UI

### Requirement: Agent-Controlled Visual Canvas

The system MUST serve a canvas workspace at `/__makoclaw__/canvas/` and provide the agent with `push`, `eval`, `snapshot`, and `reset` commands via the `canvas` tool.

#### Scenario: Agent pushes HTML content

- GIVEN `canvas.enabled = true` and the canvas route is mounted
- WHEN the agent calls `canvas push` with valid HTML
- THEN the content renders in the frontend Canvas view within 500ms

#### Scenario: Agent evaluates JS in sandboxed iframe

- GIVEN `canvas.enabled = true` and `canvas.dev_mode = true`
- WHEN the agent calls `canvas eval` with a JavaScript expression
- THEN the expression executes inside a sandboxed iframe with restrictive CSP
- AND the result is returned to the agent without affecting the host page

#### Scenario: Agent takes a snapshot

- GIVEN the canvas has rendered content
- WHEN the agent calls `canvas snapshot`
- THEN the system returns a base64-encoded PNG of the current canvas state

#### Scenario: Canvas disabled — tool call rejected

- GIVEN `canvas.enabled = false`
- WHEN the agent calls any `canvas` command
- THEN the tool returns a clear error: `"Canvas is disabled. Set canvas.enabled = true to use this feature."`
- AND no route is mounted at `/__makoclaw__/canvas/`

---

## Domain 4: Channel Actions

### Requirement: Interactive Message Approval

The system MUST support agent-initiated interactive messages (buttons, reactions) that gate execution until a user approves or rejects via channel UI.

#### Scenario: Discord approval button — user approves

- GIVEN `channel_actions.enabled = true` and Discord interactions endpoint is configured
- WHEN the agent sends an interactive message with an "Approve" button via `interactive_message` tool
- THEN Discord renders the button in the message
- AND when the user clicks "Approve", the agent receives an approval event and proceeds

#### Scenario: Slack emoji reaction approval

- GIVEN `channel_actions.enabled = true` and a Slack message is awaiting approval
- WHEN the user reacts with 👍 to the Slack message
- THEN the emoji reaction polling loop detects the reaction and signals approval to the waiting agent

#### Scenario: No public HTTPS endpoint — actions disabled

- GIVEN `channel_actions.enabled = true` but no public HTTPS interactions URL is configured
- WHEN the system starts
- THEN interactive channel actions are disabled with a logged warning: `"Channel actions require a public HTTPS interactions_endpoint_url. Actions disabled."`
- AND the agent's `interactive_message` tool returns an error on invocation

#### Scenario: Channel actions subsystem disabled

- GIVEN `channel_actions.enabled = false`
- WHEN the agent calls `interactive_message`
- THEN the tool returns a clear error and no interactive message is sent to any channel

---

## Config Flags (Feature Gates)

| Flag | Type | Default | Effect when false |
|------|------|---------|-------------------|
| `voice.enabled` | bool | false | Wake word + Talk Mode inactive |
| `canvas.enabled` | bool | false | Canvas route not mounted; canvas tool errors |
| `canvas.dev_mode` | bool | false | `eval` command disabled |
| `channel_actions.enabled` | bool | false | interactive_message tool errors; no webhook handlers registered |
