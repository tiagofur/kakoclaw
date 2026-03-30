# Phase 8 — SDD Artifacts Generation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Generate complete SDD artifact sets (proposal → spec → design → tasks) for all 7 Phase 8 sub-phases inspired by OpenClaw.

**Architecture:** Each sub-phase runs the full SDD pipeline independently. Proposals are generated first (they have no dependencies). Spec and Design run in parallel after the proposal. Tasks run last, depending on both spec and design. Sub-phases are independent of each other and can be parallelized.

**Tech Stack:** SDD skills (`sdd-propose`, `sdd-spec`, `sdd-design`, `sdd-tasks`), engram as artifact store, MakoClaw Go codebase.

**Reference:** Design doc at `docs/superpowers/specs/2026-03-30-phase8-openclaw-inspired-design.md`

**Artifact store:** engram, topic keys format: `sdd/{change-name}/{artifact}`

---

## Dependency Graph

```
[8.1 proposal] → [8.1 spec] + [8.1 design] → [8.1 tasks]
[8.2 proposal] → [8.2 spec] + [8.2 design] → [8.2 tasks]
[8.3 proposal] → [8.3 spec] + [8.3 design] → [8.3 tasks]
[8.4 proposal] → [8.4 spec] + [8.4 design] → [8.4 tasks]
[8.5 proposal] → [8.5 spec] + [8.5 design] → [8.5 tasks]
[8.6a proposal] → [8.6a spec] + [8.6a design] → [8.6a tasks]
[8.6b proposal] → [8.6b spec] + [8.6b design] → [8.6b tasks]
```

All 7 sub-phases are independent of each other.

---

## Sub-Phase Reference

| Sub-phase | Change name (engram key) | Features |
|-----------|--------------------------|----------|
| 8.1 Security & Trust | `phase-8.1-security-trust` | DM Pairing Policy, Security Hooks, Tool Profiles |
| 8.2 AI Intelligence | `phase-8.2-ai-intelligence` | Thinking Levels, llm_task Tool, Memory Flush |
| 8.3 Voice & Interaction | `phase-8.3-voice-interaction` | Voice Wake/Talk Mode, Canvas A2UI, Channel Actions |
| 8.4 Memory & Knowledge | `phase-8.4-memory-knowledge` | Pluggable Memory Backends, Session Cross-Messaging |
| 8.5 Media & Generation | `phase-8.5-media-generation` | Image Generation Tool, PDF Tool (enriched) |
| 8.6a Channels Expansion | `phase-8.6a-channels-expansion` | Multi-Account, 9 new channels |
| 8.6b Developer Experience | `phase-8.6b-developer-experience` | CLI Onboarding, Doctor, Cron, Search providers, LLM providers |

---

## Wave 1 — Proposals (all 7 in parallel)

### Task 1.1: Proposal — 8.1 Security & Trust

**Skill:** `~/.claude/skills/sdd-propose/SKILL.md`
**Change name:** `phase-8.1-security-trust`
**Engram key:** `sdd/phase-8.1-security-trust/proposal`

- [ ] Read skill file at `~/.claude/skills/sdd-propose/SKILL.md`
- [ ] Read design doc at `docs/superpowers/specs/2026-03-30-phase8-openclaw-inspired-design.md` (section 8.1)
- [ ] Read PRD section for Phase 8.1 in `docs/archive/legacy/PRD-NEW-FEATURES.md`
- [ ] Write proposal to engram with topic_key `sdd/phase-8.1-security-trust/proposal`
- [ ] Proposal must include: intent, scope (in/out), approach summary, risks, success criteria
- [ ] Verify proposal saved: search engram for `phase-8.1-security-trust`

### Task 1.2: Proposal — 8.2 AI Intelligence

**Skill:** `~/.claude/skills/sdd-propose/SKILL.md`
**Change name:** `phase-8.2-ai-intelligence`
**Engram key:** `sdd/phase-8.2-ai-intelligence/proposal`

- [ ] Read skill file at `~/.claude/skills/sdd-propose/SKILL.md`
- [ ] Read design doc at `docs/superpowers/specs/2026-03-30-phase8-openclaw-inspired-design.md` (section 8.2)
- [ ] Read PRD section for Phase 8.2 in `docs/archive/legacy/PRD-NEW-FEATURES.md`
- [ ] Write proposal to engram with topic_key `sdd/phase-8.2-ai-intelligence/proposal`
- [ ] Proposal must include: intent, scope (in/out), approach summary, risks, success criteria
- [ ] Verify proposal saved

### Task 1.3: Proposal — 8.3 Voice & Interaction

**Skill:** `~/.claude/skills/sdd-propose/SKILL.md`
**Change name:** `phase-8.3-voice-interaction`
**Engram key:** `sdd/phase-8.3-voice-interaction/proposal`

- [ ] Read skill file at `~/.claude/skills/sdd-propose/SKILL.md`
- [ ] Read design doc (section 8.3)
- [ ] Read PRD section for Phase 8.3
- [ ] Write proposal to engram with topic_key `sdd/phase-8.3-voice-interaction/proposal`
- [ ] Proposal must include: intent, scope (in/out), approach summary, risks, success criteria
- [ ] Verify proposal saved

### Task 1.4: Proposal — 8.4 Memory & Knowledge

**Skill:** `~/.claude/skills/sdd-propose/SKILL.md`
**Change name:** `phase-8.4-memory-knowledge`
**Engram key:** `sdd/phase-8.4-memory-knowledge/proposal`

- [ ] Read skill file at `~/.claude/skills/sdd-propose/SKILL.md`
- [ ] Read design doc (section 8.4)
- [ ] Read PRD section for Phase 8.4
- [ ] Write proposal to engram with topic_key `sdd/phase-8.4-memory-knowledge/proposal`
- [ ] Proposal must include: intent, scope (in/out), approach summary, risks, success criteria
- [ ] Verify proposal saved

### Task 1.5: Proposal — 8.5 Media & Generation

**Skill:** `~/.claude/skills/sdd-propose/SKILL.md`
**Change name:** `phase-8.5-media-generation`
**Engram key:** `sdd/phase-8.5-media-generation/proposal`

- [ ] Read skill file at `~/.claude/skills/sdd-propose/SKILL.md`
- [ ] Read design doc (section 8.5)
- [ ] Read PRD section for Phase 8.5
- [ ] Write proposal to engram with topic_key `sdd/phase-8.5-media-generation/proposal`
- [ ] Proposal must include: intent, scope (in/out), approach summary, risks, success criteria
- [ ] Verify proposal saved

### Task 1.6: Proposal — 8.6a Channels Expansion

**Skill:** `~/.claude/skills/sdd-propose/SKILL.md`
**Change name:** `phase-8.6a-channels-expansion`
**Engram key:** `sdd/phase-8.6a-channels-expansion/proposal`

- [ ] Read skill file at `~/.claude/skills/sdd-propose/SKILL.md`
- [ ] Read design doc (section 8.6a)
- [ ] Read PRD section for Phase 8.6a
- [ ] Write proposal to engram with topic_key `sdd/phase-8.6a-channels-expansion/proposal`
- [ ] Proposal must include: intent, scope (in/out), approach summary, risks, success criteria
- [ ] Verify proposal saved

### Task 1.7: Proposal — 8.6b Developer Experience

**Skill:** `~/.claude/skills/sdd-propose/SKILL.md`
**Change name:** `phase-8.6b-developer-experience`
**Engram key:** `sdd/phase-8.6b-developer-experience/proposal`

- [ ] Read skill file at `~/.claude/skills/sdd-propose/SKILL.md`
- [ ] Read design doc (section 8.6b)
- [ ] Read PRD section for Phase 8.6b
- [ ] Write proposal to engram with topic_key `sdd/phase-8.6b-developer-experience/proposal`
- [ ] Proposal must include: intent, scope (in/out), approach summary, risks, success criteria
- [ ] Verify proposal saved

---

## Wave 2 — Specs + Designs (parallel per sub-phase, after Wave 1)

> Run spec and design for each sub-phase concurrently. Spec and design within the same sub-phase can also run in parallel.

### Task 2.1a: Spec — 8.1 Security & Trust

**Skill:** `~/.claude/skills/sdd-spec/SKILL.md`
**Depends on:** Task 1.1 (proposal must exist in engram)

- [ ] Read skill file at `~/.claude/skills/sdd-spec/SKILL.md`
- [ ] Retrieve proposal from engram: `mem_search("sdd/phase-8.1-security-trust/proposal")` → `mem_get_observation(id)`
- [ ] Write spec to engram: topic_key `sdd/phase-8.1-security-trust/spec`
- [ ] Spec must include: functional requirements with Given/When/Then scenarios for each feature
- [ ] Scenarios must cover: DM Pairing (pairing flow, allowlist persistence, /approve command), Security Hooks (before_tool_call blocking, before_install approval, message_sending cancel), Tool Profiles (messaging/developer/minimal presets, exec security levels)
- [ ] Verify spec saved

### Task 2.1b: Design — 8.1 Security & Trust

**Skill:** `~/.claude/skills/sdd-design/SKILL.md`
**Depends on:** Task 1.1 (proposal must exist in engram)
**Can run in parallel with:** Task 2.1a

- [ ] Read skill file at `~/.claude/skills/sdd-design/SKILL.md`
- [ ] Retrieve proposal from engram
- [ ] Read relevant Go files: `pkg/agent/permissions.go`, `pkg/channels/base.go`, `pkg/config/config.go`, `pkg/tools/base.go`
- [ ] Write design to engram: topic_key `sdd/phase-8.1-security-trust/design`
- [ ] Design must include: Go struct definitions, interface changes, config schema additions, hook registration pattern, storage schema for allowlists
- [ ] Verify design saved

### Task 2.2a: Spec — 8.2 AI Intelligence

**Skill:** `~/.claude/skills/sdd-spec/SKILL.md`
**Depends on:** Task 1.2

- [ ] Read skill file at `~/.claude/skills/sdd-spec/SKILL.md`
- [ ] Retrieve proposal from engram: `sdd/phase-8.2-ai-intelligence/proposal`
- [ ] Write spec: topic_key `sdd/phase-8.2-ai-intelligence/spec`
- [ ] Scenarios must cover: `/think` command parsing and level storage, llm_task tool execution with alternate provider, memory flush trigger before compaction
- [ ] Verify spec saved

### Task 2.2b: Design — 8.2 AI Intelligence

**Skill:** `~/.claude/skills/sdd-design/SKILL.md`
**Depends on:** Task 1.2

- [ ] Read skill file at `~/.claude/skills/sdd-design/SKILL.md`
- [ ] Retrieve proposal from engram
- [ ] Read relevant Go files: `pkg/agent/loop.go`, `pkg/providers/`, `pkg/session/`
- [ ] Write design: topic_key `sdd/phase-8.2-ai-intelligence/design`
- [ ] Design must include: ThinkLevel type, session storage for think level, LLMTaskTool struct, memory flush hook interface
- [ ] Verify design saved

### Task 2.3a: Spec — 8.3 Voice & Interaction

**Skill:** `~/.claude/skills/sdd-spec/SKILL.md`
**Depends on:** Task 1.3

- [ ] Read skill file at `~/.claude/skills/sdd-spec/SKILL.md`
- [ ] Retrieve proposal from engram: `sdd/phase-8.3-voice-interaction/proposal`
- [ ] Write spec: topic_key `sdd/phase-8.3-voice-interaction/spec`
- [ ] Scenarios must cover: wake word detection flow, Talk Mode session lifecycle, TTS provider selection, Canvas push/eval/snapshot, Discord button interactions, Slack action blocks, emoji approval flow
- [ ] Verify spec saved

### Task 2.3b: Design — 8.3 Voice & Interaction

**Skill:** `~/.claude/skills/sdd-design/SKILL.md`
**Depends on:** Task 1.3

- [ ] Read skill file at `~/.claude/skills/sdd-design/SKILL.md`
- [ ] Retrieve proposal from engram
- [ ] Read relevant Go files: `pkg/voice/`, `pkg/channels/discord.go`, `pkg/channels/slack.go`, `pkg/web/`
- [ ] Write design: topic_key `sdd/phase-8.3-voice-interaction/design`
- [ ] Design must include: VoiceWakeConfig, TalkModeSession, CanvasTool struct, ChannelAction interface, EmojiApproval handler
- [ ] Verify design saved

### Task 2.4a: Spec — 8.4 Memory & Knowledge

**Skill:** `~/.claude/skills/sdd-spec/SKILL.md`
**Depends on:** Task 1.4

- [ ] Read skill file at `~/.claude/skills/sdd-spec/SKILL.md`
- [ ] Retrieve proposal from engram: `sdd/phase-8.4-memory-knowledge/proposal`
- [ ] Write spec: topic_key `sdd/phase-8.4-memory-knowledge/spec`
- [ ] Scenarios must cover: switching memory backends via config, vector search returning ranked results, sessions_send delivering message and awaiting reply, fallback when target session is inactive
- [ ] Verify spec saved

### Task 2.4b: Design — 8.4 Memory & Knowledge

**Skill:** `~/.claude/skills/sdd-design/SKILL.md`
**Depends on:** Task 1.4

- [ ] Read skill file at `~/.claude/skills/sdd-design/SKILL.md`
- [ ] Retrieve proposal from engram
- [ ] Read relevant Go files: `pkg/storage/knowledge.go`, `pkg/tools/knowledge.go`, `pkg/session/`
- [ ] Write design: topic_key `sdd/phase-8.4-memory-knowledge/design`
- [ ] Design must include: MemoryBackend interface, BuiltinBackend, QMDBackend, HonchoBackend structs, SessionsSendTool, embedding provider abstraction
- [ ] Verify design saved

### Task 2.5a: Spec — 8.5 Media & Generation

**Skill:** `~/.claude/skills/sdd-spec/SKILL.md`
**Depends on:** Task 1.5

- [ ] Read skill file at `~/.claude/skills/sdd-spec/SKILL.md`
- [ ] Retrieve proposal from engram: `sdd/phase-8.5-media-generation/proposal`
- [ ] Write spec: topic_key `sdd/phase-8.5-media-generation/spec`
- [ ] Scenarios must cover: generate_image with DALL-E provider, generate_image with Fal provider, PDF text extraction, PDF OCR fallback, PDF table extraction, batch PDF processing
- [ ] Verify spec saved

### Task 2.5b: Design — 8.5 Media & Generation

**Skill:** `~/.claude/skills/sdd-design/SKILL.md`
**Depends on:** Task 1.5

- [ ] Read skill file at `~/.claude/skills/sdd-design/SKILL.md`
- [ ] Retrieve proposal from engram
- [ ] Read relevant Go files: `pkg/tools/`, `pkg/config/config.go`
- [ ] Write design: topic_key `sdd/phase-8.5-media-generation/design`
- [ ] Design must include: ImageGenerationTool, ImageProvider interface, DALLEProvider, FalProvider, PDFTool struct, OCR integration
- [ ] Verify design saved

### Task 2.6a: Spec — 8.6a Channels Expansion

**Skill:** `~/.claude/skills/sdd-spec/SKILL.md`
**Depends on:** Task 1.6

- [ ] Read skill file at `~/.claude/skills/sdd-spec/SKILL.md`
- [ ] Retrieve proposal from engram: `sdd/phase-8.6a-channels-expansion/proposal`
- [ ] Write spec: topic_key `sdd/phase-8.6a-channels-expansion/spec`
- [ ] Scenarios must cover: multi-account routing (3 WhatsApp accounts → 3 agents), LINE message receive/send, each new channel basic send/receive, channel binding deterministic routing
- [ ] Verify spec saved

### Task 2.6b-spec: Spec — 8.6b Developer Experience

**Skill:** `~/.claude/skills/sdd-spec/SKILL.md`
**Depends on:** Task 1.7

- [ ] Read skill file at `~/.claude/skills/sdd-spec/SKILL.md`
- [ ] Retrieve proposal from engram: `sdd/phase-8.6b-developer-experience/proposal`
- [ ] Write spec: topic_key `sdd/phase-8.6b-developer-experience/spec`
- [ ] Scenarios must cover: onboard wizard step sequence, doctor --deep output format, cron job with stagger, cron delivery via webhook, Exa search returning results, DeepSeek provider call
- [ ] Verify spec saved

### Task 2.6a-design: Design — 8.6a Channels Expansion

**Skill:** `~/.claude/skills/sdd-design/SKILL.md`
**Depends on:** Task 1.6

- [ ] Read skill file at `~/.claude/skills/sdd-design/SKILL.md`
- [ ] Retrieve proposal from engram
- [ ] Read relevant Go files: `pkg/channels/base.go`, `pkg/channels/manager.go`, `pkg/config/config.go`
- [ ] Write design: topic_key `sdd/phase-8.6a-channels-expansion/design`
- [ ] Design must include: MultiAccountBinding struct, account routing table, each new channel struct skeleton implementing Channel interface
- [ ] Verify design saved

### Task 2.6b-design: Design — 8.6b Developer Experience

**Skill:** `~/.claude/skills/sdd-design/SKILL.md`
**Depends on:** Task 1.7

- [ ] Read skill file at `~/.claude/skills/sdd-design/SKILL.md`
- [ ] Retrieve proposal from engram
- [ ] Read relevant Go files: `pkg/cron/`, `cmd/makoclaw/main.go`, `pkg/doctor/`, `pkg/tools/web.go`
- [ ] Write design: topic_key `sdd/phase-8.6b-developer-experience/design`
- [ ] Design must include: OnboardWizard struct, DoctorDeepCheck interface, CronStagger algorithm, CronDeliveryMode enum, SearchProvider interface additions
- [ ] Verify design saved

---

## Wave 3 — Tasks (after Wave 2, per sub-phase)

### Task 3.1: Tasks — 8.1 Security & Trust

**Skill:** `~/.claude/skills/sdd-tasks/SKILL.md`
**Depends on:** Tasks 2.1a + 2.1b

- [ ] Read skill file at `~/.claude/skills/sdd-tasks/SKILL.md`
- [ ] Retrieve spec from engram: `sdd/phase-8.1-security-trust/spec`
- [ ] Retrieve design from engram: `sdd/phase-8.1-security-trust/design`
- [ ] Write task checklist to engram: topic_key `sdd/phase-8.1-security-trust/tasks`
- [ ] Tasks must be ordered, actionable, and reference exact Go files
- [ ] Each task must have: description, files to modify, test to write, commit message
- [ ] Verify tasks saved

### Task 3.2: Tasks — 8.2 AI Intelligence

**Skill:** `~/.claude/skills/sdd-tasks/SKILL.md`
**Depends on:** Tasks 2.2a + 2.2b

- [ ] Read skill file at `~/.claude/skills/sdd-tasks/SKILL.md`
- [ ] Retrieve spec + design from engram for `phase-8.2-ai-intelligence`
- [ ] Write task checklist: topic_key `sdd/phase-8.2-ai-intelligence/tasks`
- [ ] Verify tasks saved

### Task 3.3: Tasks — 8.3 Voice & Interaction

**Skill:** `~/.claude/skills/sdd-tasks/SKILL.md`
**Depends on:** Tasks 2.3a + 2.3b

- [ ] Read skill file at `~/.claude/skills/sdd-tasks/SKILL.md`
- [ ] Retrieve spec + design from engram for `phase-8.3-voice-interaction`
- [ ] Write task checklist: topic_key `sdd/phase-8.3-voice-interaction/tasks`
- [ ] Verify tasks saved

### Task 3.4: Tasks — 8.4 Memory & Knowledge

**Skill:** `~/.claude/skills/sdd-tasks/SKILL.md`
**Depends on:** Tasks 2.4a + 2.4b

- [ ] Read skill file at `~/.claude/skills/sdd-tasks/SKILL.md`
- [ ] Retrieve spec + design from engram for `phase-8.4-memory-knowledge`
- [ ] Write task checklist: topic_key `sdd/phase-8.4-memory-knowledge/tasks`
- [ ] Verify tasks saved

### Task 3.5: Tasks — 8.5 Media & Generation

**Skill:** `~/.claude/skills/sdd-tasks/SKILL.md`
**Depends on:** Tasks 2.5a + 2.5b

- [ ] Read skill file at `~/.claude/skills/sdd-tasks/SKILL.md`
- [ ] Retrieve spec + design from engram for `phase-8.5-media-generation`
- [ ] Write task checklist: topic_key `sdd/phase-8.5-media-generation/tasks`
- [ ] Verify tasks saved

### Task 3.6a: Tasks — 8.6a Channels Expansion

**Skill:** `~/.claude/skills/sdd-tasks/SKILL.md`
**Depends on:** Tasks 2.6a (spec) + 2.6a-design

- [ ] Read skill file at `~/.claude/skills/sdd-tasks/SKILL.md`
- [ ] Retrieve spec + design from engram for `phase-8.6a-channels-expansion`
- [ ] Write task checklist: topic_key `sdd/phase-8.6a-channels-expansion/tasks`
- [ ] Verify tasks saved

### Task 3.7: Tasks — 8.6b Developer Experience

**Skill:** `~/.claude/skills/sdd-tasks/SKILL.md`
**Depends on:** Tasks 2.6b-spec + 2.6b-design

- [ ] Read skill file at `~/.claude/skills/sdd-tasks/SKILL.md`
- [ ] Retrieve spec + design from engram for `phase-8.6b-developer-experience`
- [ ] Write task checklist: topic_key `sdd/phase-8.6b-developer-experience/tasks`
- [ ] Verify tasks saved

---

## Completion Checklist

- [ ] All 7 proposals saved to engram
- [ ] All 7 specs saved to engram
- [ ] All 7 designs saved to engram
- [ ] All 7 task checklists saved to engram
- [ ] Total: 28 engram artifacts created
- [ ] Design doc at `docs/superpowers/specs/2026-03-30-phase8-openclaw-inspired-design.md` still intact
- [ ] PRD at `docs/archive/legacy/PRD-NEW-FEATURES.md` Phase 8 section intact
