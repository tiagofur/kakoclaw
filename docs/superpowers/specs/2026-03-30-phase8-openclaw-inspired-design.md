---
title: Phase 8 — OpenClaw-Inspired Features Design
date: 2026-03-30
status: approved
---

# Phase 8 — OpenClaw-Inspired Features

## Context

MakoClaw reviewed the OpenClaw open-source project (https://github.com/openclaw/openclaw — Node.js AI agent framework) and identified 24 features and improvements to incorporate. These are organized into 7 sub-phases documented as full SDD artifact sets.

Source analysis available in: `docs/archive/legacy/PRD-NEW-FEATURES.md` (Phase 8 section)
Reference codebase: `openclaw-main/` (local copy)

## Documentation Strategy

Each sub-phase produces two documentation layers:

### Layer 1 — Enriched PRD
For each feature within the sub-phase, added to `docs/archive/legacy/PRD-NEW-FEATURES.md`:
- User Stories
- Acceptance Criteria
- Out of Scope
- Open Questions

### Layer 2 — SDD Artifacts
One complete set per sub-phase stored in engram with topic keys `sdd/phase-8.X-<slug>/*`:
- `proposal` — what changes, why, scope
- `spec` — functional requirements with Given/When/Then scenarios
- `design` — Go architecture, interfaces, design decisions
- `tasks` — ordered implementation checklist

## Sub-Phases

### 8.1 — Security & Trust
**Slug**: `phase-8.1-security-trust`
**Features**:
- DM Pairing Policy: challenge-based allowlisting for unknown senders (`dmPolicy: pairing|open|disabled`)
- Security Hooks: programmable hooks `before_tool_call`, `before_install`, `message_sending`
- Tool Profiles: named presets `messaging`, `developer`, `minimal` with exec security levels
**Source**: `openclaw-main/src/` — DM policy, security hooks

### 8.2 — AI Intelligence
**Slug**: `phase-8.2-ai-intelligence`
**Features**:
- Thinking Levels per Session: `/think off|minimal|low|medium|high|xhigh` chat command, controls extended thinking budget
- `llm_task` Tool: agent delegates subtasks to a different (cheaper/faster) model
- Context Compaction Memory Flush: silent agent turn before compaction to persist important context
**Source**: `openclaw-main/extensions/llm-task`, `openclaw-main/src/sessions/`

### 8.3 — Voice & Interaction
**Slug**: `phase-8.3-voice-interaction`
**Features**:
- Voice Wake + Talk Mode: global wake words, continuous voice conversation overlay, TTS (ElevenLabs + system fallback), STT (Deepgram)
- Canvas / A2UI: agent-controlled HTML/CSS/JS workspace, push/reset/eval/snapshot
- Channel Action Framework: Discord buttons/modals, Slack action blocks, emoji-based approvals (`👍 approve`)
**Source**: `openclaw-main/src/canvas-host/`, `openclaw-main/src/voice/`

### 8.4 — Memory & Knowledge
**Slug**: `phase-8.4-memory-knowledge`
**Features**:
- Pluggable Memory Backends: builtin (SQLite + vector), QMD (local sidecar with reranking), Honcho (AI-native cross-session)
- Session Cross-Messaging: `sessions_send` tool — agent sends message to another existing session and receives reply
**Source**: `openclaw-main/extensions/memory-*`

### 8.5 — Media & Generation
**Slug**: `phase-8.5-media-generation`
**Features**:
- Image Generation Tool: DALL-E (OpenAI), Fal, Replicate; extensible provider pattern
- PDF Tool (enriched): native Anthropic/Google PDF vision support, OCR fallback, table extraction, batch processing (enriched from existing backlog item with OpenClaw patterns)
**Source**: `openclaw-main/extensions/image-generation-core`

### 8.6a — Channels Expansion
**Slug**: `phase-8.6a-channels-expansion`
**Features**:
- Multi-Account per Channel: multiple WhatsApp numbers, Telegram bots, Discord bots in one instance with deterministic routing
- New Channels: LINE (Japan/Asia), WeChat (Tencent iLink), Zalo (Vietnam), Mattermost (self-hosted), Matrix (decentralized), Nostr (NIP-04 DMs), Twitch (IRC backend), IRC (classic), Twilio Voice (telephony)
**Source**: `openclaw-main/extensions/` (channel plugins)

### 8.6b — Developer Experience
**Slug**: `phase-8.6b-developer-experience`
**Features**:
- CLI Onboarding Wizard: `makoclaw onboard --install-daemon`, step-by-step channel pairing, model selection, daemon install (launchd/systemd)
- Doctor Deep Diagnostics: `doctor --deep --fix --json`, security audit, config drift detection, filesystem permission checks, symlink escape detection
- Cron Stagger + Delivery Modes: deterministic per-job stagger, session targets (`main|isolated|current|session:id`), delivery modes (`announce|webhook|none`)
- Web Search Providers: add Exa (semantic), Tavily (research), Perplexity, Firecrawl (AI scraping)
- LLM Providers Expansion: DeepSeek, Amazon Bedrock, LiteLLM proxy (self-hosted), Together AI
**Source**: `openclaw-main/src/cli/`, `openclaw-main/src/doctor/`

## Artifact Store

Backend: **engram** (active in project)

Topic key format: `sdd/{change-name}/{artifact}`

| Sub-phase | Change name |
|-----------|-------------|
| 8.1 | `phase-8.1-security-trust` |
| 8.2 | `phase-8.2-ai-intelligence` |
| 8.3 | `phase-8.3-voice-interaction` |
| 8.4 | `phase-8.4-memory-knowledge` |
| 8.5 | `phase-8.5-media-generation` |
| 8.6a | `phase-8.6a-channels-expansion` |
| 8.6b | `phase-8.6b-developer-experience` |

## Implementation Order

Sub-phases are independent and can be executed in parallel. Recommended sequence by value/effort ratio:

1. **8.2** (AI Intelligence) — highest ROI, low effort
2. **8.1** (Security & Trust) — foundational security gaps
3. **8.6b** (DX) — quick wins (providers, cron, doctor)
4. **8.4** (Memory) — foundational for AI quality
5. **8.5** (Media) — user-visible features
6. **8.3** (Voice & Interaction) — high effort, high impact
7. **8.6a** (Channels) — largest scope, parallel execution

## Out of Scope for Phase 8

- Real-time collaboration (separate initiative)
- Mobile app development (requires native dev resources beyond Go)
- Offline mode / local-only deployment (separate initiative)
- Backward compatibility breaks to existing channel configurations
