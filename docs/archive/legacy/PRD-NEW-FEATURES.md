# PRD: New Features Roadmap

**Document Version**: 1.0
**Created**: 2026-03-04
**Last Updated**: 2026-03-04
**Status**: Draft - Pending Prioritization

---

## Executive Summary

This PRD documents new features identified through competitive analysis and codebase audit. Features are categorized by impact, effort, and strategic value for MakoClaw's positioning as a lightweight, efficient AI agent framework.

**Our competitive advantages to maintain**:
- <10MB RAM footprint (vs OpenClaw's 1.52GB)
- <1s startup time (vs OpenClaw's 5.98s)
- Complete Web UI (vs NanoClaw/PicoClaw CLI-only)
- Multi-user support with isolation
- 9 communication channels

---

## Table of Contents

1. [Agent Swarms](#1-agent-swarms)
2. [PDF Tool Native](#2-pdf-tool-native)
3. [Browser Automation Tool](#3-browser-automation-tool)
4. [Voice-to-Voice Mode](#4-voice-to-voice-mode)
5. [Plugin Marketplace](#5-plugin-marketplace)
6. [Memory Timeline UI](#6-memory-timeline-ui)
7. [Workflow Visual Builder v2](#7-workflow-visual-builder-v2)
8. [Real-time Collaboration](#8-real-time-collaboration)
9. [Cost Dashboard](#9-cost-dashboard)
10. [Audit Trail UI](#10-audit-trail-ui)
11. [Multilingual Stop Keywords](#11-multilingual-stop-keywords)
12. [Offline Mode](#12-offline-mode)
13. [Mobile App](#13-mobile-app)

---

## 1. Agent Swarms — IMPLEMENTED

> **Status**: Implemented on 2026-03-04 (commit `cd5925d`)
> **Actual effort**: ~1 session (vs estimated 6-7 weeks)

### What Was Built

**Backend** (`pkg/agent/swarm.go`, ~370 lines):
- `SwarmRunner` with 3 execution modes: sequential, parallel (`errgroup`), consensus
- Shared `TeamContext` memory propagation between specialists (capped at 4000 chars)
- Per-swarm cost budget enforcement with automatic cancellation
- Context-based timeout enforcement (default 300s)
- Real-time `AgentStatusEvent` callbacks per member
- 3 built-in templates: code-review-team, research-team, full-stack-team

**Config** (`pkg/config/config.go`):
- `SwarmConfig` struct with name, description, members, mode, max_budget, timeout, shared_memory
- Persisted per-user via existing `config.json` system

**API** (`pkg/web/server.go`):
- `GET/POST /api/v1/swarms` — list/create swarms
- `GET/PUT/DELETE /api/v1/swarms/{name}` — CRUD
- `POST /api/v1/swarms/run` — execute swarm via REST
- `GET /api/v1/swarms/templates` — list built-in templates
- WebSocket `swarm_run` message type with `swarm_start`/`agent_status`/`swarm_complete` events

**Frontend**:
- Swarm section in AgentsView with create/run/delete modals
- `SwarmVisualizer.vue` — SVG flow diagram with status-colored nodes
- Pinia store integration (agentsStore + chatStore)

**Tests** (`pkg/agent/swarm_test.go`): 24 unit tests covering creation, validation, budget, timeout, aggregation, shared notes, templates, AgentManager CRUD.

**Bug fixes included**: goroutine leak in SubagentManager (detached context + 5min timeout).

### Requirements Delivered

| Req ID | Requirement | Status |
|--------|-------------|--------|
| SW-R1 | Swarm creation via UI and API | Done |
| SW-R2 | Shared memory between swarm agents | Done |
| SW-R3 | Configurable execution mode (parallel/sequential/consensus) | Done |
| SW-R4 | Swarm-level cost tracking | Done |
| SW-R5 | Visual representation of agent collaboration | Done |
| SW-R6 | Swarm templates (pre-configured teams) | Done |

---

## 2. PDF Tool Native

### Overview
Replace basic PDF text extraction with a comprehensive PDF processing tool supporting native LLM capabilities.

### Problem Statement
Current implementation only extracts text between BT/ET markers. Encrypted, scanned, or complex PDFs fail silently. OpenClaw v2026.3.2 introduced native Anthropic/Google PDF support.

### Proposed Solution

```go
type PDFTool struct {
    *BaseTool
    Provider     PDFProvider  // anthropic | google | fallback
    OCREnabled   bool
    MaxPages     int
    ExtractMeta  bool
}

type PDFProvider interface {
    ExtractText(ctx context.Context, data []byte) (*PDFResult, error)
    ExtractImages(ctx context.Context, data []byte) ([]Image, error)
    OCRPage(ctx context.Context, pageImage []byte) (string, error)
}

type PDFResult struct {
    Text      string
    Pages     []PageContent
    Metadata  PDFMetadata
    Images    []Image
    Tables    []Table
}
```

### Features
1. **Native LLM extraction**: Use Claude/Gemini vision for complex PDFs
2. **OCR fallback**: For scanned documents
3. **Table extraction**: Structured data from tables
4. **Image extraction**: Pull embedded images
5. **Metadata extraction**: Author, dates, properties

### Technical Requirements

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| PDF-R1 | Native Anthropic PDF support via vision | P0 |
| PDF-R2 | Google Gemini PDF support | P1 |
| PDF-R3 | OCR fallback for scanned PDFs | P1 |
| PDF-R4 | Table extraction to structured format | P2 |
| PDF-R5 | Batch processing for multi-file uploads | P2 |

### Effort Estimate
- Backend: 2 weeks
- Testing: 1 week
- **Total: 3 weeks**

---

## 3. Browser Automation Tool

### Overview
Add browser automation capabilities for web scraping, form filling, and visual element interaction.

### Problem Statement
Many automation tasks require browser interaction. Users currently need external tools. OpenClaw's "unbrowse" feature with visual element detection is highly demanded.

### Proposed Solution

```go
type BrowserTool struct {
    *BaseTool
    Headless    bool
    Timeout     time.Duration
    Viewport    Viewport
    UserAgent   string
}

type BrowserAction struct {
    Type        ActionType  // Navigate, Click, Type, Screenshot, Extract
    Selector    string      // CSS/XPath selector
    Value       string      // For Type actions
    WaitFor     string      // Wait condition
}

func (bt *BrowserTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    actions := parseActions(args)
    return bt.executeSequence(ctx, actions)
}
```

### Features
1. **Navigation**: Go to URLs, handle redirects
2. **Visual element detection**: Find elements by visual appearance
3. **Form interaction**: Fill forms, submit, handle CAPTCHAs (manual)
4. **Screenshot capture**: Full page or element screenshots
5. **Data extraction**: Scrape structured data
6. **Cookie/session management**: Persist sessions

### Security Considerations
- Allowlist of permitted domains
- Rate limiting
- No credential storage in browser
- Sandbox execution

### Effort Estimate
- Backend: 4 weeks (using playwright-go or rod)
- Testing: 2 weeks
- Security review: 1 week
- **Total: 7 weeks**

---

## 4. Voice-to-Voice Mode

### Overview
Enable voice responses, not just voice input transcription. Full conversational voice interaction.

### Problem Statement
Currently we support voice input via transcription (Groq). Users want voice responses for hands-free operation, accessibility, and mobile use.

### Proposed Solution

```go
type VoiceConfig struct {
    InputEnabled    bool
    OutputEnabled   bool
    Voice           string     // alloy, echo, fable, onyx, nova, shimmer
    Speed           float64    // 0.25 to 4.0
    Provider        string     // openai, elevenlabs, google
}

type VoiceResponse struct {
    Text      string
    AudioURL  string
    Duration  time.Duration
}
```

### Channels Supported
- Telegram (voice messages)
- Discord (voice channels)
- WhatsApp (voice notes)
- Web UI (audio playback)

### Technical Requirements

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| V2V-R1 | TTS integration (OpenAI/ElevenLabs) | P0 |
| V2V-R2 | Voice selection per user/channel | P1 |
| V2V-R3 | Streaming audio for long responses | P1 |
| V2V-R4 | Web UI audio player component | P1 |
| V2V-R5 | Discord voice channel integration | P2 |

### Effort Estimate
- Backend: 2 weeks
- Channel integrations: 2 weeks
- Frontend: 1 week
- **Total: 5 weeks**

---

## 5. Plugin Marketplace

### Overview
One-click skill installation with community ratings, reviews, and auto-updates.

### Current State
- We have `handlers_marketplace.go` with security scanning
- Skills can be installed manually
- No discovery, ratings, or versioning

### Proposed Enhancements

```go
type MarketplaceSkill struct {
    ID          string
    Name        string
    Version     string
    Author      Author
    Downloads   int
    Rating      float64
    Reviews     []Review
    Categories  []string
    Verified    bool       // Passed security scan
    UpdatedAt   time.Time
}

type SkillRegistry interface {
    Search(query string, filters Filters) ([]MarketplaceSkill, error)
    Install(skillID string, version string) error
    Update(skillID string) error
    Uninstall(skillID string) error
    Rate(skillID string, rating int, review string) error
}
```

### Features
1. **Skill discovery**: Search, categories, trending
2. **One-click install**: From UI
3. **Ratings & reviews**: Community feedback
4. **Versioning**: Semantic versioning, update notifications
5. **Auto-update**: Optional automatic updates
6. **Verified badge**: Security-scanned skills

### Effort Estimate
- Backend: 3 weeks
- Frontend: 2 weeks
- Infrastructure (registry server): 2 weeks
- **Total: 7 weeks**

---

## 6. Memory Timeline UI

### Overview
Visual timeline showing what the agent "remembers" across sessions, with filtering and search.

### Problem Statement
Users don't have visibility into what the agent remembers. Memory is a black box. Need transparency for trust and debugging.

### Proposed Solution

```typescript
interface MemoryEntry {
  id: string;
  timestamp: Date;
  type: 'fact' | 'preference' | 'decision' | 'context' | 'correction';
  content: string;
  source: 'user' | 'agent' | 'tool';
  confidence: number;
  relatedEntries: string[];
}

// Vue component
<MemoryTimeline
  :entries="memories"
  :filters="{ type: [], dateRange: [], source: [] }"
  @delete="handleDelete"
  @edit="handleEdit"
/>
```

### Features
1. **Timeline view**: Chronological display
2. **Type filtering**: Facts, preferences, decisions
3. **Source filtering**: User-provided vs agent-inferred
4. **Search**: Full-text search
5. **Edit/Delete**: User can correct or remove memories
6. **Export**: Download memory as JSON/markdown

### Effort Estimate
- Backend API: 1 week
- Frontend: 2 weeks
- **Total: 3 weeks**

---

## 7. Workflow Visual Builder v2

### Overview
Enhanced drag-and-drop workflow builder with conditional logic, loops, and human-in-the-loop nodes.

### Current State
We have basic workflows. Need more sophisticated flow control.

### New Node Types

```go
type ConditionalNode struct {
    BaseNode
    Condition   string      // Expression to evaluate
    TrueBranch  *Node
    FalseBranch *Node
}

type LoopNode struct {
    BaseNode
    Collection  string      // Array to iterate
    ItemVar     string      // Variable name for current item
    Body        []Node
    MaxIter     int         // Safety limit
}

type ApprovalNode struct {
    BaseNode
    Approvers   []string    // User IDs
    Timeout     time.Duration
    OnTimeout   ApprovalAction  // Approve | Reject | Escalate
}

type ErrorHandlerNode struct {
    BaseNode
    TryNodes    []Node
    CatchNodes  []Node
    FinallyNode *Node
}
```

### UI Enhancements
1. **Drag-and-drop**: Visual node placement
2. **Connection lines**: Animated flow visualization
3. **Mini-map**: Navigation for large workflows
4. **Undo/redo**: Full history
5. **Templates**: Pre-built workflow templates
6. **Version history**: Track changes

### Effort Estimate
- Backend: 2 weeks
- Frontend: 4 weeks
- **Total: 6 weeks**

---

## 8. Real-time Collaboration

### Overview
Multiple users interacting with the same agent simultaneously for team standups, brainstorming, etc.

### Use Cases
1. **Team standups**: Agent facilitates, tracks action items
2. **Brainstorming**: Multiple people contribute, agent synthesizes
3. **Pair debugging**: Two developers + agent troubleshooting
4. **Training**: Expert shows agent capabilities to team

### Technical Design

```go
type CollaborativeSession struct {
    ID          string
    AgentID     string
    Participants []Participant
    Messages    chan Message
    State       SessionState
}

type Participant struct {
    UserID      string
    DisplayName string
    Role        ParticipantRole  // Owner | Editor | Viewer
    JoinedAt    time.Time
}
```

### Features
1. **Session creation**: Start collaborative session
2. **Invite links**: Share session with others
3. **Role permissions**: Owner, editor, viewer
4. **Presence indicators**: See who's active
5. **Message attribution**: See who said what
6. **Summary generation**: Agent summarizes session

### Effort Estimate
- Backend (WebSocket enhancement): 3 weeks
- Frontend: 3 weeks
- **Total: 6 weeks**

---

## 9. Cost Dashboard — IMPLEMENTED

> **Status**: Implemented on 2026-03-04
> **Scope**: Phase 1 (real-time in-memory metrics)

### What Was Built

**Backend**:
- `GetCostTracker()` method on `AgentManager` to expose cost tracker
- `handleAgentMetrics` now returns real data from `AgentCostTracker` (was hardcoded mock)
- `POST /api/v1/agents/metrics/reset` — admin-only endpoint to reset cost metrics

**Frontend** (`MetricsView.vue`):
- 4th "Cost Estimates" card showing total cost, API calls, tokens, top agent
- "Cost by Agent" breakdown table with per-agent calls, tokens, cost, avg cost
- Admin-only reset button
- Fetches from `/agents/metrics` in parallel with existing `/metrics`

**Tests** (`cost_tracker_test.go`):
- `GetCostTracker` on nil/empty manager
- `GetCostSummary` with multiple agents
- `Reset` clears all metrics

### Remaining Work (Phase 2)
- Persist cost data to database for historical tracking
- Timeline charts (daily/weekly/monthly trends)
- Budget alerts/thresholds
- By-provider breakdown

---

## 10. Audit Trail UI — IMPLEMENTED

> **Status**: UI existed prior (`AuditLogTab.vue`). Export feature added 2026-03-04.

### What Was Built

**Pre-existing** (already in codebase):
- `AuditLogTab.vue` in Settings with filters (user, tool, success/failure)
- `handleToolAudit` handler with paginated queries
- `SQLiteAuditLogger` with full CRUD

**Added**:
- `GET /api/v1/tools/audit/export?format=csv|json` — admin-only bulk export endpoint (100K limit)
- CSV export with Content-Disposition header for download
- JSON export with timestamp and count metadata
- Export buttons (CSV/JSON) in `AuditLogTab.vue` with current filter passthrough

**Tests** (`handlers_tools_test.go`):
- JSON export, CSV export, auth check, empty results

### Remaining Work
- Suspicious activity alerts
- Configurable log retention/cleanup
- Date range filters in export UI

---

## 11. Multilingual Stop Keywords — IMPLEMENTED

> **Status**: Implemented on 2026-03-04

### What Was Built

**Backend** (`pkg/agent/stop_keywords.go`):
- `IsStopCommand(text string) bool` — O(1) lookup via pre-built map in `init()`
- 10 languages: en, es, fr, de, pt, zh, ja, ar, hi, ru
- Case-insensitive, whitespace-trimmed matching

**WebSocket integration** (`pkg/web/server.go`):
- Stop keyword check before message processing
- Cancels all active executions for session
- Sends `stream_end` + `ready` messages back to client

**Channel integration** (`pkg/channels/command_handler.go`):
- `isStopCommand` check at top of `HandleCommand` (before `/` prefix check)
- `SetCancelFunc` to wire up execution cancellation per channel
- Local keyword set (avoids circular import with agent package)

**Tests** (`pkg/agent/stop_keywords_test.go`):
- All 10 languages, case insensitivity, whitespace handling, non-stop words

---

## 12. Offline Mode

### Overview
Basic functionality without internet using local Ollama.

### Challenge
Ollama provider currently doesn't support tools.

### Required Work

1. **Add tool support to Ollama provider** (separate bug fix)
2. **Detect offline state**: Network status check
3. **Graceful degradation**: Which features work offline
4. **Sync when online**: Queue actions, sync when reconnected

### Offline-Available Features
- Chat with local LLM
- Local file operations
- Task management (local DB)
- Knowledge base (local search)

### Offline-Unavailable Features
- Web search
- External API calls
- Channel messages (queued)
- Cloud provider LLMs

### Effort Estimate
- Ollama tool support: 2 weeks
- Offline detection: 1 week
- Queue/sync: 2 weeks
- **Total: 5 weeks**

---

## 13. Mobile App

### Overview
React Native app for iOS and Android with core MakoClaw functionality.

### MVP Features
1. **Chat**: Conversation with agent
2. **Tasks**: View and manage Kanban tasks
3. **Push notifications**: Agent alerts
4. **Voice input**: Speech-to-text
5. **Quick actions**: Common commands

### Technical Stack
- React Native
- TypeScript
- Existing REST API
- Push: Firebase Cloud Messaging

### Screens
1. Login/Auth
2. Chat (main)
3. Tasks (Kanban view)
4. Settings
5. Agent selection

### Effort Estimate
- Setup & architecture: 1 week
- Chat screen: 2 weeks
- Tasks screen: 1 week
- Push notifications: 1 week
- Voice input: 1 week
- Testing & polish: 2 weeks
- **Total: 8 weeks**

---

## Priority Matrix

| Feature | Impact | Effort | Priority | Status |
|---------|--------|--------|----------|--------|
| ~~Multilingual Stop Keywords~~ | ~~Medium~~ | ~~Low~~ | ~~P0~~ | **DONE** |
| ~~Cost Dashboard~~ | ~~High~~ | ~~Medium~~ | ~~P0~~ | **DONE** (Phase 1) |
| Memory Timeline UI | Medium | Low | P1 | Pre-existing |
| ~~Audit Trail UI~~ | ~~Medium~~ | ~~Low~~ | ~~P1~~ | **DONE** (+ export) |
| PDF Tool Native | High | Medium | P1 | Pending |
| ~~Agent Swarms~~ | ~~Very High~~ | ~~High~~ | ~~P1~~ | **DONE** |
| Voice-to-Voice | High | Medium | P2 | Pending |
| Plugin Marketplace | High | High | P2 | Pending |
| Workflow Builder v2 | Medium | High | P2 | Pending |
| Browser Automation | High | High | P3 | Pending |
| Real-time Collaboration | Medium | High | P3 | Pending |
| Offline Mode | Medium | High | P3 | Pending |
| Mobile App | High | Very High | P3 | Pending |

---

## Appendix A: Competitor Feature Comparison

| Feature | MakoClaw | OpenClaw | NanoClaw | PicoClaw | ZeroClaw |
|---------|----------|----------|----------|----------|----------|
| RAM Usage | <10MB | 1.52GB | ~50MB | <10MB | ~7.8MB |
| Startup | <1s | 5.98s | ~2s | <1s | <10ms |
| Web UI | Yes | Yes | No | No | No |
| Multi-user | Yes | Yes | Yes | No | No |
| Agent Swarms | **Yes** | Yes | Yes | No | No |
| Channels | 9 | 12+ | 5 | 4 | 3 |
| Specialists | Yes | Yes | No | No | No |
| PDF Tool | Basic | Advanced | No | No | No |
| Browser | **Planned** | Yes | No | No | No |
| Voice I/O | Input only | Both | No | No | No |
| Cron Jobs | Yes | Yes | No | No | No |
| Knowledge Base | Yes | Yes | No | No | No |
| MCP Support | Yes | Yes | Yes | No | No |
| Cost Tracking | **Yes** | Yes | No | No | No |

---

## Appendix B: References

- [OpenClaw Wikipedia](https://en.wikipedia.org/wiki/OpenClaw)
- [OpenClaw v2026.3.2 Release Notes](https://www.ainvest.com/news/openclaw-v2026-3-2-release-adds-pdf-analysis-tool-150-fixes-breaking-2603/)
- [NanoClaw GitHub](https://github.com/qwibitai/nanoclaw)
- [PicoClaw GitHub](https://github.com/sipeed/picoclaw)
- [AI Agent Trends 2026](https://www.blueprism.com/resources/blog/future-ai-agents-trends/)
- [Multi-Agent Frameworks Explained](https://www.adopt.ai/blog/multi-agent-frameworks)
- [Swarms AI Enterprise Framework](https://www.swarms.ai/)

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-04 | Claude + Tiago | Initial PRD based on competitive analysis and audit |
| 1.1 | 2026-03-04 | Claude + Tiago | Agent Swarms implemented (all 6 requirements delivered) |
| 1.2 | 2026-03-04 | Claude + Tiago | Cost Dashboard (Phase 1), Audit Export, Multilingual Stop Keywords implemented |
| 1.3 | 2026-03-30 | Claude + Tiago | Phase 8: OpenClaw-Inspired Features section added (16 new features, 8 improvements) |

---

## Phase 8 — OpenClaw-Inspired Features

> **Source**: Competitive analysis of OpenClaw (https://github.com/openclaw/openclaw) vs MakoClaw.
> **Methodology**: Direct codebase comparison. Features split into two categories: (A) net-new capabilities MakoClaw lacks entirely, and (B) areas where MakoClaw has the feature but OpenClaw's implementation is materially better.
> **Strategic note**: All features must preserve MakoClaw's core differentiators — <10MB RAM, <1s startup, single binary. Nothing in this phase should compromise those constraints.

---

### Phase 8 — Summary Table

#### Category A: New Features

| # | Feature | Impact | Effort | Priority | Status |
|---|---------|--------|--------|----------|--------|
| 14 | Voice Wake Word + Talk Mode | High | High | P1 | Pending |
| 15 | Pluggable Memory Backends | High | High | P1 | Pending |
| 16 | DM Pairing Policy (Security) | High | Medium | P1 | Pending |
| 17 | Security Hooks (Plugin Hooks) | High | Medium | P1 | Pending |
| 18 | Canvas / Agent-to-UI Workspace | Medium | High | P2 | Pending |
| 19 | Device Pairing & Node Architecture | Medium | High | P2 | Pending |
| 20 | Mobile Companion Apps (Node Model) | Medium | Very High | P2 | Pending |
| 21 | Channel Action Framework | Medium | Medium | P2 | Pending |
| 22 | Thinking Level per Session | Medium | Low | P2 | Pending |
| 23 | Image Generation Tool | Medium | Medium | P2 | Pending |
| 24 | llm_task Tool (Cross-Model Subtasks) | Medium | Low | P2 | Pending |
| 25 | Multi-Account Channel Support | Medium | High | P2 | Pending |
| 26 | Additional Channels (11 new) | Low | High | P3 | Pending |
| 27 | Community Skills Registry (MakoHub) | Medium | High | P3 | Pending |
| 28 | Gmail Pub/Sub Inbound Triggers | Low | Medium | P3 | Pending |
| 29 | Lobster Workflow Integration | Low | High | P3 | Pending |

#### Category B: Improvements to Existing Features

| # | Feature | Impact | Effort | Priority | Status |
|---|---------|--------|--------|----------|--------|
| B1 | Context Compaction — Memory Flush | High | Low | P1 | Pending |
| B2 | Doctor Command — Deep Audits | High | Medium | P1 | Pending |
| B3 | Cron — Stagger + Delivery Modes | Medium | Low | P2 | Pending |
| B4 | Tool Profiles (messaging/developer/minimal) | Medium | Low | P2 | Pending |
| B5 | Session Cross-Messaging (sessions_send) | Medium | Medium | P2 | Pending |
| B6 | Onboarding — CLI Daemon Wizard | Medium | Medium | P2 | Pending |
| B7 | Web Search — Additional Providers | Low | Low | P3 | Pending |
| B8 | LLM Providers — Expanded Coverage | Low | Low | P3 | Pending |

---

### Category A: New Features

---

## 14. Voice Wake Word + Talk Mode

### Overview

Add configurable wake words that activate the agent passively on supported platforms (macOS, iOS, Android), plus a continuous "Talk Mode" that sustains a voice conversation without requiring repeated wake words — functioning like a phone call to the agent.

### Problem Statement

MakoClaw supports voice input via Groq Whisper (STT), but it is strictly push-to-record. There is no wake word detection and no hands-free continuous conversation. OpenClaw ships wake word activation and Talk Mode as first-class features, making it far better suited for ambient/hands-free environments.

### Proposed Solution

```go
type WakeWordConfig struct {
    Enabled    bool
    Words      []string   // e.g. ["hey mako", "ok mako"]
    Sensitivity float64   // 0.0–1.0
    Platform   string    // macos | ios | android
}

type TalkModeSession struct {
    ID          string
    Channel     string
    Active      bool
    STTProvider string   // deepgram | groq | whisper
    TTSProvider string   // elevenlabs | openai | system
    Voice       string
}
```

TTS providers: ElevenLabs (primary), OpenAI TTS, system TTS (macOS `say`, Android TTS) as fallback.
STT providers: Deepgram (primary), Groq Whisper (existing fallback).
Wake word engine: Picovoice Porcupine (Go bindings) or whisper.cpp local for offline.

### Functional Requirements

- Wake words configurable per-user in `config.json`
- Talk Mode toggle via `/talk on` and `/talk off` commands in any channel
- Talk Mode overlay in Web UI (visual indicator, mic status, waveform)
- TTS response delivery in Telegram (voice messages), Discord (voice channel), Web UI (audio)
- ElevenLabs voice selection per user; system TTS as zero-dependency fallback
- Wake word sensitivity tunable

### Technical Notes

- Web UI uses browser Web Speech API for wake word in browser context (no native binary needed)
- Native platform apps use Porcupine or whisper.cpp for always-on detection
- Talk Mode session is tied to a session key — no multi-session bleed

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| VW-R1 | Wake word detection (configurable words, platform-aware) | P0 |
| VW-R2 | Talk Mode continuous conversation loop | P0 |
| VW-R3 | ElevenLabs TTS with system TTS fallback | P0 |
| VW-R4 | Deepgram STT with Groq fallback | P1 |
| VW-R5 | Talk Mode overlay in Web UI | P1 |
| VW-R6 | `/talk on|off` command across channels | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/` — voice wake system, Talk Mode overlay)
**Effort estimate**: Backend 3 weeks · Channel integrations 2 weeks · Frontend 1 week · **Total: 6 weeks**

---

## 15. Pluggable Memory Backends

### Overview

Replace MakoClaw's single-backend knowledge base with a pluggable memory architecture supporting three backends: the existing builtin SQLite (improved), a local-first sidecar with reranking, and an AI-native cross-session memory service. Add hybrid keyword+vector search and multiple embedding providers.

### Problem Statement

MakoClaw's current RAG system uses SQLite FTS5 (keyword search only). There is no vector similarity search, no pluggable backend, and no cross-session user modeling. OpenClaw ships three memory backends plus hybrid search, making long-term agent memory dramatically more effective.

### Proposed Solution

```go
type MemoryBackend interface {
    Store(ctx context.Context, entry MemoryEntry) error
    Search(ctx context.Context, query string, opts SearchOptions) ([]MemoryResult, error)
    Flush(ctx context.Context) error
    Name() string
}

type SearchOptions struct {
    Hybrid     bool      // keyword + vector
    Limit      int
    MinScore   float64
    Rerank     bool
}

// Backends
type BuiltinBackend struct { /* SQLite + FTS5 + optional vector via sqlite-vec */ }
type QMDBackend struct { /* local sidecar: reranking + query expansion */ }
type HonchoBackend struct { /* AI-native: Honcho API, cross-session user modeling */ }

// Embedding providers
type EmbeddingProvider interface {
    Embed(ctx context.Context, texts []string) ([][]float32, error)
}
// Implementations: OpenAI, Gemini, Voyage, Mistral
```

Auto-flush strategy: before context compaction, the agent receives a silent internal "poke" turn to save important facts to the active memory backend before the context window is pruned.

### Functional Requirements

- Memory backend selectable per-user in config (`memory.backend: builtin|qmd|honcho`)
- Builtin backend: add sqlite-vec extension for vector search alongside existing FTS5
- Hybrid search mode: keyword score + vector similarity combined with configurable weights
- Embedding provider: OpenAI, Gemini, Voyage, Mistral — auto-detect from configured keys
- QMD backend: local sidecar process, managed lifecycle, reranking + query expansion
- Honcho backend: REST client to Honcho API, cross-session user modeling, configurable endpoint
- Auto-flush: agent loop calls `memory.Flush()` before compaction threshold is reached
- Memory backend status in doctor command and Web UI diagnostics

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| MB-R1 | Pluggable `MemoryBackend` interface replacing direct SQLite calls | P0 |
| MB-R2 | Builtin backend: hybrid search (FTS5 + sqlite-vec) | P0 |
| MB-R3 | Embedding provider abstraction (OpenAI, Gemini, Voyage, Mistral) | P0 |
| MB-R4 | Auto-flush before compaction | P1 |
| MB-R5 | QMD backend (local sidecar) | P1 |
| MB-R6 | Honcho backend (API client) | P2 |
| MB-R7 | Memory backend selector in Web UI settings | P1 |

**Inspired by**: OpenClaw (`openclaw-main/extensions/memory-*`)
**Effort estimate**: Backend 4 weeks · Frontend 1 week · Testing 2 weeks · **Total: 7 weeks**

---

## 16. DM Pairing Policy (Security)

### Overview

Replace the static allowlist approach for direct messages with a dynamic pairing challenge system. Unknown senders receive a code they must present for approval, protecting the agent from unsolicited use while keeping onboarding friction low.

### Problem Statement

MakoClaw uses static `allow_from` allowlists. Adding new users requires editing `config.json`. There is no self-service onboarding for new senders, and there is no revocation flow. OpenClaw's pairing policy solves this cleanly.

### Proposed Solution

```go
type DMPolicy string

const (
    DMPolicyPairing  DMPolicy = "pairing"   // default: unknown senders get a code, await approval
    DMPolicyOpen     DMPolicy = "open"       // explicit allowlist required (current behavior)
    DMPolicyDisabled DMPolicy = "disabled"   // no DMs accepted
)

type PairingRequest struct {
    SenderID    string
    Channel     string
    Code        string      // 6-digit alphanumeric
    RequestedAt time.Time
    ExpiresAt   time.Time   // 24h
    Approved    bool
}
```

Commands:
- `/approve <channel> <code>` — approve a pending pairing request (usable from any channel)
- `/reject <channel> <code>` — explicitly reject

Storage: pairing requests and approved senders persisted in `database.db` per user.

### Functional Requirements

- `dm_policy` field in channel config (`config.json`) with `pairing | open | disabled`
- `pairing` mode: unknown sender receives "Send this code to your admin: `XK8P2Q`" message
- Code expires after 24h; resending the first message generates a new code
- Approved senders added to persistent DB allowlist (not `config.json`) — survives restarts
- `/approve` and `/reject` commands work from any channel (admin-only)
- Web UI: pending pairing requests list in Settings > Channels with approve/reject buttons
- `disabled` mode: agent sends "DMs are disabled" and drops message

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| DM-R1 | `dm_policy` config field with three modes | P0 |
| DM-R2 | Pairing mode: code generation and delivery | P0 |
| DM-R3 | `/approve` and `/reject` commands | P0 |
| DM-R4 | Persistent approved sender DB (per channel, per user) | P0 |
| DM-R5 | Code expiry (24h) | P1 |
| DM-R6 | Web UI pending requests panel | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/` — DM policy system)
**Effort estimate**: Backend 2 weeks · Frontend 1 week · **Total: 3 weeks**

---

## 17. Security Hooks (Plugin Hooks)

### Overview

A programmable hook system that intercepts tool calls, skill installs, and outbound messages before execution. Hooks can block, require human approval, or allow operations — with approvals deliverable via any connected channel.

### Problem Statement

MakoClaw has audit logging (after the fact) and role-based tool permissions (static). There is no runtime interception of tool calls before execution. High-risk operations cannot be paused for human approval. OpenClaw's hook system enables this.

### Proposed Solution

```go
type HookPhase string

const (
    HookBeforeToolCall   HookPhase = "before_tool_call"
    HookBeforeInstall    HookPhase = "before_install"
    HookMessageSending   HookPhase = "message_sending"
)

type HookAction string

const (
    HookAllow          HookAction = "allow"
    HookBlock          HookAction = "block"
    HookRequireApproval HookAction = "require_approval"
)

type Hook struct {
    Phase      HookPhase
    Priority   int          // lower = runs first
    Match      HookMatcher  // tool name pattern, skill id, channel, etc.
    Action     HookAction
    ApprovalChannel string   // channel to forward approval request to
    TimeoutSec int
    OnTimeout  HookAction   // default: block
}

type HookResult struct {
    Action  HookAction
    Reason  string
    ApprovalID string  // if RequireApproval
}
```

Approval flow: hook fires → pending approval sent to `ApprovalChannel` → admin sees `[APPROVAL REQUIRED] exec: rm -rf /tmp/x — reply /approve <id> or /reject <id>` → agent waits up to `TimeoutSec`.

### Functional Requirements

- Hooks defined in `config.json` under `security.hooks[]`
- `before_tool_call`: match by tool name glob (e.g. `exec.*`, `write_file`)
- `before_install`: fires when skill install is initiated
- `message_sending`: fires before any outbound message is delivered to a channel
- Multiple hooks on same phase execute in priority order; first non-allow wins
- Approval forwarding: pending approval sent as message to configured channel
- `/approve <id>` and `/reject <id>` commands usable from any channel
- Timeout behavior: configurable per hook (default: block on timeout)
- Hook evaluation visible in audit log

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| SH-R1 | Hook configuration in `config.json` | P0 |
| SH-R2 | `before_tool_call` hook with tool name matching | P0 |
| SH-R3 | RequireApproval action with channel forwarding | P0 |
| SH-R4 | `/approve` and `/reject` commands | P0 |
| SH-R5 | Priority ordering for multiple hooks on same phase | P1 |
| SH-R6 | `before_install` hook | P1 |
| SH-R7 | `message_sending` hook | P2 |
| SH-R8 | Hook audit log entries | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/` — security hooks)
**Effort estimate**: Backend 2 weeks · Testing 1 week · **Total: 3 weeks**

---

## 18. Canvas / Agent-to-UI Interactive Workspace

### Overview

An HTML/CSS/JS workspace that the agent can control in real time — distinct from the static Web UI. The agent pushes content, resets, evaluates expressions, and takes snapshots. The Canvas is served by the MakoClaw gateway and displayed in an iframe or dedicated view.

### Problem Statement

MakoClaw's Web UI is a fixed interface that users control. There is no mechanism for the agent to push arbitrary rendered content to the user — data visualizations, interactive dashboards, live documents. OpenClaw's Canvas feature enables this.

### Proposed Solution

The Canvas host runs as a lightweight HTTP server (goroutine, not a separate binary) serving a minimal HTML page. The agent uses new tools to manipulate it:

```go
// Agent tools
type CanvasPushTool struct{}   // push HTML/CSS/JS content to canvas
type CanvasResetTool struct{}  // clear canvas to blank state
type CanvasEvalTool struct{}   // evaluate JS expression in canvas context, return result
type CanvasSnapshotTool struct{} // capture PNG screenshot of current canvas state
```

Canvas content is pushed via WebSocket from agent → gateway → browser. The canvas view in the Web UI displays the iframe and provides a "pop out" option.

### Functional Requirements

- Canvas host: in-process HTTP server on configurable port (default 8081)
- Agent tools: `canvas_push`, `canvas_reset`, `canvas_eval`, `canvas_snapshot`
- Web UI: Canvas tab/view with iframe + pop-out button
- Snapshot tool returns base64 PNG (uses headless chromium via `chromedp` or screenshot via browser API)
- Canvas state persisted for session — reconnecting browser restores last pushed content
- Canvas can be disabled in config (`web.canvas.enabled: false`) for resource-constrained devices

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| CV-R1 | In-process canvas HTTP server | P0 |
| CV-R2 | `canvas_push` tool (HTML/CSS/JS) | P0 |
| CV-R3 | `canvas_reset` and `canvas_eval` tools | P1 |
| CV-R4 | Canvas view in Web UI with iframe | P0 |
| CV-R5 | `canvas_snapshot` tool (PNG capture) | P2 |
| CV-R6 | Session state persistence | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/canvas-host/`)
**Effort estimate**: Backend 3 weeks · Frontend 2 weeks · **Total: 5 weeks**

---

## 19. Device Pairing & Node Architecture

### Overview

A trust model and pairing protocol for connecting external devices (phones, tablets, other machines) to a MakoClaw gateway as "nodes". Nodes expose device capabilities (camera, microphone, location, notifications) as agent tools. Pairing uses challenge signatures and Bonjour/mDNS for local discovery.

### Problem Statement

MakoClaw agents can only interact with hardware on the server where MakoClaw runs. There is no way to use the camera on a nearby phone, get the user's location, or send native device notifications. OpenClaw's node architecture solves this via device pairing.

### Proposed Solution

```go
type Node struct {
    ID           string
    DeviceType   string   // ios | android | macos | linux
    Capabilities []string // camera, screen, location, notifications, sms, contacts, calendar
    PairedAt     time.Time
    TrustScore   float64
}

type PairingChallenge struct {
    Code      string    // QR-displayable setup code
    Nonce     string
    ExpiresAt time.Time
}

type NodeCommand struct {
    NodeID  string
    Command string   // camera.capture, location.get, notifications.send, screen.record, etc.
    Params  map[string]interface{}
}
```

Pairing: gateway displays QR code → device scans → device sends challenge response signed with device key → gateway verifies and stores device identity.

Node commands become agent tools dynamically registered after pairing.

### Functional Requirements

- Gateway exposes `/api/v1/nodes/pair` for initiating pairing
- QR code displayed in Web UI Devices section
- Device identity stored in `central.db` with capability list
- Dynamically registered agent tools per paired device and capability
- Node commands: `camera.capture`, `location.get`, `notifications.send`, `screen.record`, `sms.send` (Android only), `contacts.list`, `calendar.list`
- Bonjour/mDNS announcement for local network discovery
- Node revocation: `/api/v1/nodes/{id}` DELETE
- Web UI: Devices section showing paired nodes, capabilities, last seen

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| ND-R1 | Pairing protocol with challenge signatures | P0 |
| ND-R2 | QR code display in Web UI | P0 |
| ND-R3 | Dynamic tool registration per node capability | P0 |
| ND-R4 | Core node commands (camera, location, notifications) | P0 |
| ND-R5 | Bonjour/mDNS local discovery | P1 |
| ND-R6 | Node revocation | P1 |
| ND-R7 | Extended commands (sms, contacts, calendar) | P2 |
| ND-R8 | Web UI Devices panel | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/` — node pairing system)
**Effort estimate**: Backend 4 weeks · Frontend 1 week · **Total: 5 weeks**

---

## 20. Mobile Companion Apps (Node Model)

### Overview

Native iOS and Android apps that function as "nodes" rather than simple chat clients. Each app exposes device hardware capabilities to the MakoClaw gateway: camera, microphone, location, notifications, contacts, calendar, and screen recording. Pairing is via QR code (see Feature 19).

### Problem Statement

MakoClaw has no native mobile apps. The Web UI works on mobile browsers but cannot access hardware or send native push notifications. OpenClaw ships native iOS and Android apps built on the node pairing model.

### Proposed Solution

Two separate apps, both using the same pairing protocol (Feature 19):

**iOS App**
- Camera capture (photo + video)
- Microphone (voice input for Talk Mode)
- Screen recording
- Push notifications (APNs)
- Pairing via QR scan

**Android App**
- Camera capture
- Microphone
- Location (GPS)
- Push notifications (FCM)
- SMS access (with permission)
- Contacts and Calendar read
- Motion sensors
- App usage stats
- Pairing via QR scan

Both apps maintain a persistent WebSocket connection to the MakoClaw gateway to receive node commands and deliver results.

### Functional Requirements

- iOS: Swift + SwiftUI, minimum iOS 16
- Android: Kotlin + Jetpack Compose, minimum Android 10
- Both: persistent WebSocket node connection, background reconnect
- Both: QR-based pairing with gateway (see Feature 19)
- Both: receive node commands and return results
- Both: push notifications for agent messages (APNs / FCM)
- Optional: local chat UI for conversational access without Web UI

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| MA-R1 | iOS app: pairing + camera + microphone + push | P0 |
| MA-R2 | Android app: pairing + camera + location + push | P0 |
| MA-R3 | Persistent WebSocket node connection with reconnect | P0 |
| MA-R4 | Android: SMS, contacts, calendar | P1 |
| MA-R5 | Both: local chat UI | P2 |
| MA-R6 | App Store / Play Store distribution | P2 |

**Inspired by**: OpenClaw mobile node apps
**Effort estimate**: iOS 6 weeks · Android 6 weeks · Backend (node protocol) shared with Feature 19 · **Total: 12 weeks**

---

## 21. Channel Action Framework

### Overview

Add interactive UI components to channel messages — buttons, select menus, modals (Discord/Slack), and emoji reactions as agent action approvals. Enables users to approve or reject agent actions inline in their messaging app without switching to the Web UI.

### Problem Statement

MakoClaw messages are purely text-based. Users cannot respond to agent prompts with buttons or react to messages to approve actions. OpenClaw's channel action framework adds interactive components to Discord and Slack, and uses emoji reactions as universal approval signals.

### Proposed Solution

```go
type ActionComponent interface {
    Render(channel string) interface{}  // channel-specific component format
}

type ButtonAction struct {
    Label    string
    ActionID string
    Style    ButtonStyle   // primary | danger | secondary
}

type SelectMenuAction struct {
    Placeholder string
    Options     []SelectOption
    ActionID    string
}

type EmojiReactionApproval struct {
    ApproveEmoji string  // default: 👍
    RejectEmoji  string  // default: 👎
    PendingID    string  // links to security hook approval (Feature 17)
}
```

Integration with security hooks (Feature 17): when a hook requires approval, the agent can deliver the approval request with a button component or emoji prompt in the active channel.

### Functional Requirements

- Discord: button rows, select menus, modals, slash command responses
- Slack: Block Kit buttons, slash command responses, app actions
- Universal: emoji reactions as approval signals (👍 approve / 👎 reject)
- Emoji approval integrates with DM pairing (Feature 16) and security hooks (Feature 17)
- Action results (button clicks, reactions) delivered as agent tool call responses
- Channels without component support fall back to text prompts (`Reply YES or NO`)

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| CA-R1 | Discord button rows and select menus | P0 |
| CA-R2 | Slack Block Kit buttons | P0 |
| CA-R3 | Emoji reaction approval across all channels | P0 |
| CA-R4 | Discord modals for data input | P1 |
| CA-R5 | Slash command responses (Discord + Slack) | P1 |
| CA-R6 | Text fallback for non-interactive channels | P0 |

**Inspired by**: OpenClaw (`openclaw-main/src/` — channel actions)
**Effort estimate**: Backend 2 weeks · Channel integrations 2 weeks · **Total: 4 weeks**

---

## 22. Thinking Level per Session

### Overview

A per-session command to control the extended thinking budget for models that support it (Claude extended thinking, GPT-5 extended thinking). Lets users dial up thinking for complex problems and dial down for fast responses — without changing global config.

### Problem Statement

MakoClaw can visualize extended thinking tokens in the Web UI, but the thinking budget is set globally. Users cannot adjust it per conversation. OpenClaw exposes this as a first-class session command.

### Proposed Solution

```go
type ThinkLevel string

const (
    ThinkOff     ThinkLevel = "off"
    ThinkMinimal ThinkLevel = "minimal"  // ~500 tokens
    ThinkLow     ThinkLevel = "low"      // ~2000 tokens
    ThinkMedium  ThinkLevel = "medium"   // ~5000 tokens
    ThinkHigh    ThinkLevel = "high"     // ~10000 tokens
    ThinkXHigh   ThinkLevel = "xhigh"    // ~20000 tokens
)

type SessionOptions struct {
    ThinkLevel ThinkLevel
    // other per-session options...
}
```

Command: `/think off|minimal|low|medium|high|xhigh` — sets think level for the current session only. Stored in session state, not persisted to config. Resets to global default on new session.

### Functional Requirements

- `/think <level>` command recognized in all channels and Web UI
- Think level stored in session state (not global config)
- Budget translated to provider-specific format (Claude: `budget_tokens`, GPT-5: TBD)
- Models without extended thinking: command acknowledged but has no effect
- Web UI: think level indicator in chat header, clickable to cycle through levels
- Think level resets to global default on session end

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| TL-R1 | `/think <level>` command in all channels | P0 |
| TL-R2 | Session-scoped storage (not config) | P0 |
| TL-R3 | Provider translation (Claude budget_tokens) | P0 |
| TL-R4 | Web UI indicator + click-to-change | P1 |
| TL-R5 | Graceful no-op on unsupported models | P0 |

**Inspired by**: OpenClaw (`openclaw-main/src/sessions/` — think level commands)
**Effort estimate**: Backend 1 week · Frontend 0.5 weeks · **Total: 1.5 weeks**

---

## 23. Image Generation Tool

### Overview

Add an image generation tool that supports multiple backend providers: DALL-E (OpenAI), Fal, Replicate, and optionally Midjourney. The tool is extensible via the existing provider pattern.

### Problem Statement

MakoClaw has no image generation capability. Users working with creative or visual tasks must leave MakoClaw and use external tools. OpenClaw ships an extensible image generation tool with multiple provider backends.

### Proposed Solution

```go
type ImageGenTool struct {
    *BaseTool
    Providers map[string]ImageGenProvider
    Default   string
}

type ImageGenProvider interface {
    Generate(ctx context.Context, prompt string, opts ImageGenOptions) (*GeneratedImage, error)
    Name() string
}

type ImageGenOptions struct {
    Size    string   // 1024x1024, 1792x1024, etc.
    Quality string   // standard | hd (DALL-E), fast | quality (Fal)
    Style   string   // vivid | natural (DALL-E)
    N       int      // number of images
}

type GeneratedImage struct {
    URL     string
    Base64  string
    Revised string   // revised prompt (DALL-E feature)
}
```

Tool usage: `generate_image(prompt, provider?, size?, quality?)` → returns URL or base64.
Delivery: image sent as media in Telegram/Discord/WhatsApp; as `<img>` in Web UI chat.

### Functional Requirements

- `generate_image` tool registered in AgentLoop when at least one provider is configured
- Provider selection via config key presence (OpenAI key → DALL-E enabled, Fal key → Fal enabled)
- Default provider: first configured; override per-call with `provider` param
- Response delivery: URL for cloud providers, inline base64 for local providers
- Channel delivery: Telegram sends as photo, Discord as embed, Web UI renders inline
- Usage tracked in cost dashboard (tokens equivalent or flat per-image cost)
- Config: `providers.image.default`, `providers.fal.api_key`, `providers.replicate.api_key`

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| IG-R1 | `generate_image` tool with pluggable provider interface | P0 |
| IG-R2 | DALL-E provider (uses existing OpenAI key) | P0 |
| IG-R3 | Fal provider | P1 |
| IG-R4 | Replicate provider | P1 |
| IG-R5 | Channel delivery (Telegram photo, Discord embed, Web UI inline) | P0 |
| IG-R6 | Cost tracking integration | P1 |

**Inspired by**: OpenClaw (`openclaw-main/extensions/image-generation-core`)
**Effort estimate**: Backend 2 weeks · Channel delivery 1 week · **Total: 3 weeks**

---

## 24. llm_task Tool (Cross-Model Subtasks)

### Overview

A new agent tool that executes a subtask using a different LLM than the one currently running. Enables cost and latency optimization: use cheap/fast models for summaries or classification, expensive models for reasoning.

### Problem Statement

MakoClaw's `spawn` tool creates subagent conversations. It does not allow routing a specific subtask to a different model without changing global config or spawning a full subagent session. OpenClaw's `llm_task` tool provides a lightweight single-turn cross-model execution primitive.

### Proposed Solution

```go
type LLMTaskTool struct {
    *BaseTool
    Providers map[string]providers.LLMProvider  // all configured providers
}

// Tool schema exposed to agent:
// llm_task(task: string, model: string, system?: string, max_tokens?: int) -> string
func (t *LLMTaskTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    task    := args["task"].(string)
    model   := args["model"].(string)    // e.g. "groq/llama-3-8b-instant"
    system  := args["system"].(string)   // optional system prompt
    // Select provider, execute single-turn, return text
}
```

Key difference from `spawn`: `llm_task` is a single-turn, no-history, no-tools call. It is a synchronous primitive for delegating text processing — not a conversation.

### Functional Requirements

- `llm_task(task, model, system?, max_tokens?)` tool registered in AgentLoop
- Model specified in `provider/model` syntax (same as existing model override)
- Single-turn execution: no conversation history, no tool calls in the subtask
- Cost tracked under the parent agent's session in cost dashboard
- Timeout: inherits parent context timeout
- Error handling: subtask failure returns error string (does not crash parent loop)

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| LT-R1 | `llm_task` tool with `provider/model` syntax | P0 |
| LT-R2 | Single-turn, no-history, no-tools execution | P0 |
| LT-R3 | Cost tracking under parent session | P1 |
| LT-R4 | Optional system prompt override | P1 |

**Inspired by**: OpenClaw (`openclaw-main/extensions/llm-task`)
**Effort estimate**: Backend 1 week · **Total: 1 week**

---

## 25. Multi-Account Channel Support

### Overview

Allow a single MakoClaw instance to run multiple accounts for the same channel simultaneously — for example, two Telegram bots, three WhatsApp numbers, two Discord bots. Each account is bound to a specific agent or user, with deterministic routing.

### Problem Statement

MakoClaw supports one account per channel type. Teams that manage multiple personas, brands, or bot accounts must run separate MakoClaw instances. OpenClaw supports multi-account binding with deterministic routing rules.

### Proposed Solution

```go
type ChannelAccount struct {
    ID       string
    Channel  string   // telegram | discord | whatsapp | etc.
    Token    string   // bot token or credentials
    AgentID  string   // bound specialist or default agent
    UserID   int64    // bound user (for per-user isolation)
}

// Config structure
type ChannelsConfig struct {
    // existing single-account fields preserved for backwards compatibility
    Telegram  TelegramConfig     // legacy: single account
    Accounts  []ChannelAccount   // new: multi-account list
}
```

Routing precedence (deterministic): `peer (sender) → parentPeer (group) → accountId → fallback`.

### Functional Requirements

- `channels.accounts[]` array in `config.json` for multi-account configuration
- Backwards compatible: existing single-account config still works (treated as one account)
- Channel manager: spawn one goroutine per account, independent lifecycle
- Routing: inbound message routed to bound agent based on account binding
- Each account has independent allow_from list
- Web UI Channels section: list all accounts, add/remove, show status per account
- Maximum 10 accounts per channel type (resource protection)

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| MC-R1 | `channels.accounts[]` config with multi-account support | P0 |
| MC-R2 | Backwards compatibility with single-account config | P0 |
| MC-R3 | Independent goroutine per account with lifecycle management | P0 |
| MC-R4 | Deterministic routing: sender → group → accountId → fallback | P1 |
| MC-R5 | Per-account allow_from list | P1 |
| MC-R6 | Web UI multi-account management | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/routing/` — multi-account bindings)
**Effort estimate**: Backend 3 weeks · Frontend 1 week · **Total: 4 weeks**

---

## 26. Additional Channels (11 new)

### Overview

Expand MakoClaw from 9 to 20 channels by adding the most strategically valuable channels present in OpenClaw. Prioritized by geographic reach, self-hosted affinity, and decentralized/privacy use cases.

### Proposed Channels (Prioritized)

| Channel | Protocol / SDK | Rationale | Priority |
|---------|---------------|-----------|----------|
| LINE | LINE Messaging API | 170M+ users, dominant in Japan, Thailand, Taiwan | P2 |
| Matrix | matrix-org/gomatrix | Decentralized, bridges to many other protocols | P2 |
| Twilio Voice | Twilio Go SDK | Inbound/outbound phone calls; unique capability | P2 |
| Mattermost | matterpilot/go-mattermost | Self-hosted Slack alternative, enterprise | P2 |
| Nostr | nbd-wtf/go-nostr | Decentralized, censorship-resistant, growing | P3 |
| IRC | gopherbot/go-irc | Classic, still used in developer communities | P3 |
| WeChat | Tencent iLink Bot | 1.3B users; requires business verification | P3 |
| Zalo | Zalo Bot API | 74M+ users in Vietnam | P3 |
| Nextcloud Talk | Nextcloud API | Self-hosted video+chat; pairs with Nextcloud users | P3 |
| Twitch | IRC backend | Streaming community, bot commands | P3 |
| Mixin | Mixin Messenger API | Privacy-focused, crypto community | P3 |

### Implementation Pattern (same for all)

Each channel: new file in `pkg/channels/`, embed `BaseChannel`, implement `Channel` interface, add config fields to `ChannelsConfig`, register in `initChannels()`.

### Functional Requirements

- Each channel implements the full `Channel` interface (Start, Stop, Send, IsAllowed, etc.)
- Each channel supports `dm_policy` (Feature 16) from day one
- Each channel registered in OpenAPI docs
- Twilio Voice: special handling — audio transcription before agent processing (reuse Groq STT)
- LINE: webhook-based (no long polling), LINE Messaging API v3
- Matrix: persistent sync loop, E2EE optional (libolm)

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| AC-R1 | LINE channel implementation | P2 |
| AC-R2 | Matrix channel implementation | P2 |
| AC-R3 | Twilio Voice inbound/outbound | P2 |
| AC-R4 | Mattermost channel implementation | P2 |
| AC-R5 | Nostr channel (NIP-04 DMs) | P3 |
| AC-R6 | IRC channel | P3 |
| AC-R7 | WeChat, Zalo, Nextcloud Talk, Twitch, Mixin | P3 |

**Inspired by**: OpenClaw (`openclaw-main/extensions/` — channel plugins)
**Effort estimate**: 1–2 weeks per channel · Priority P2 channels: 8 weeks · P3: 12 weeks

---

## 27. Community Skills Registry (MakoHub)

### Overview

A public registry for MakoClaw skills (analogous to OpenClaw's ClawHub). Agents can discover, rate, and install community skills on-demand from chat or the Web UI. Builds on the existing Plugin Marketplace feature (#5) and extends it with a MakoClaw-native registry server.

### Problem Statement

MakoClaw has marketplace infrastructure (`handlers_marketplace.go`) but no public registry. Skills cannot be discovered or shared without manual file distribution. The planned Plugin Marketplace (#5) needs a registry backend to be useful.

### Proposed Solution

MakoHub registry: lightweight REST API (can be a separate open-source Go service or hosted):

```
GET  /api/skills             - list all (with filters: category, rating, verified)
GET  /api/skills/{id}        - skill detail + README
POST /api/skills/{id}/rate   - submit rating + review
GET  /api/skills/trending    - top by downloads this week
GET  /api/skills/search?q=   - full-text search
```

MakoClaw client:
```go
type MakoHubClient struct {
    BaseURL string   // default: https://makohub.dev (or self-hosted)
    APIKey  string   // optional for publishing
}
```

Agent tool: `search_skills(query)`, `install_skill(id)` — install from chat.

### Functional Requirements

- MakoHub registry URL configurable (`skills.registry_url` in config)
- Self-hostable (open-source registry server shipped with MakoClaw)
- `search_skills` and `install_skill` agent tools
- Web UI Skills section: Browse tab with search, categories, trending, install button
- Verified badge: skills that have passed automated security scan
- Rating system: 1–5 stars + text review
- Install flow: download → security scan → install → confirmation

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| MH-R1 | MakoHub registry client (configurable URL) | P0 |
| MH-R2 | `search_skills` and `install_skill` agent tools | P0 |
| MH-R3 | Web UI Browse tab in Skills section | P0 |
| MH-R4 | Self-hostable registry server | P1 |
| MH-R5 | Verified badge (security scan gate) | P1 |
| MH-R6 | Ratings and reviews | P2 |

**Inspired by**: OpenClaw ClawHub registry
**Effort estimate**: Client + tools 2 weeks · Frontend 2 weeks · Registry server 3 weeks · **Total: 7 weeks**

---

## 28. Gmail Pub/Sub Inbound Triggers

### Overview

Enable real-time email-triggered agent turns via Google Cloud Pub/Sub. When a new email arrives in a configured Gmail account, MakoClaw automatically starts an agent turn with the email content — no polling.

### Problem Statement

MakoClaw can send emails via SMTP but cannot receive or react to incoming emails. Polling is inefficient. OpenClaw integrates with Gmail's Pub/Sub push notifications for real-time inbound triggers.

### Proposed Solution

```go
type GmailTriggerConfig struct {
    Enabled        bool
    ServiceAccount string   // path to GCP service account JSON
    ProjectID      string
    SubscriptionID string
    TopicID        string
    FilterLabels   []string  // only trigger on emails with these labels
}

type GmailInboundMessage struct {
    From      string
    Subject   string
    Body      string
    MessageID string
    ThreadID  string
    Labels    []string
}
```

Flow: Gmail → Pub/Sub → MakoClaw webhook (`POST /api/v1/integrations/gmail/push`) → parse → `InboundMessage` on bus → agent turn.

### Functional Requirements

- `tools.gmail_trigger` config section
- Webhook endpoint `POST /api/v1/integrations/gmail/push` (Google-signed)
- Pub/Sub subscription management: setup via `doctor --fix` or manual config
- Email parsed to `InboundMessage` with `From`, `Subject`, `Body` in content
- Duplicate suppression: track processed `MessageID` in DB (48h window)
- Filter by Gmail label (e.g. only trigger on `INBOX`, skip `SPAM`)
- Session key: `gmail:<email_address>` for conversation continuity

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| GM-R1 | Pub/Sub webhook endpoint with Google signature verification | P0 |
| GM-R2 | Gmail → `InboundMessage` parser | P0 |
| GM-R3 | Duplicate suppression | P0 |
| GM-R4 | Label filtering | P1 |
| GM-R5 | Doctor command setup check | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/` — Gmail Pub/Sub integration)
**Effort estimate**: Backend 2 weeks · **Total: 2 weeks**

---

## 29. Lobster Workflow Integration

### Overview

A bridge between MakoClaw and Lobster — a typed JSON pipeline engine with nodes, edges, and long-running human approval pauses. Extends the existing Workflow Visual Builder with Lobster-compatible pipelines and bidirectional HTTP bridging.

### Problem Statement

MakoClaw's workflow engine can execute simple pipelines, but lacks typed schemas, long-duration human approvals, and external bridge support. OpenClaw integrates with Lobster for production-grade workflow orchestration.

### Proposed Solution

```go
type LobsterConfig struct {
    Enabled  bool
    Endpoint string   // Lobster API base URL
    APIKey   string
    AllowedTools []string  // tool allowlist for Lobster-triggered calls
}

// Lobster pipeline maps to MakoClaw workflow
type LobsterNode struct {
    ID      string
    Type    string              // tool | agent | approval | webhook
    Config  map[string]interface{}
    Outputs []LobsterEdge
}

// Long-duration approval node
type LobsterApprovalNode struct {
    LobsterNode
    ApproverIDs []string
    Timeout     time.Duration
    Channel     string   // deliver approval request to this channel
}
```

Bridge: `POST /api/v1/integrations/lobster/trigger` → execute workflow with Lobster node definitions. MakoClaw tools available to Lobster via `GET /api/v1/integrations/lobster/tools`.

### Functional Requirements

- Lobster integration configurable in `config.json` under `tools.lobster`
- REST bridge: Lobster can trigger MakoClaw tools via authenticated HTTP
- MakoClaw can trigger Lobster pipelines via `run_pipeline` agent tool
- Tool allowlist: Lobster can only call tools in `allowed_tools` list
- Approval nodes: long-running pauses delivered as channel messages (Feature 21)
- Web UI: Lobster pipelines visible in Workflows section (read-only view)

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| LB-R1 | Lobster → MakoClaw tool bridge (webhook) | P0 |
| LB-R2 | `run_pipeline` agent tool for MakoClaw → Lobster | P0 |
| LB-R3 | Tool allowlist enforcement | P0 |
| LB-R4 | Approval node delivery via channel (Feature 21) | P1 |
| LB-R5 | Lobster pipelines view in Web UI Workflows section | P2 |

**Inspired by**: OpenClaw (`openclaw-main/extensions/lobster`)
**Effort estimate**: Backend 3 weeks · Frontend 1 week · **Total: 4 weeks**

---

### Category B: Improvements to Existing Features

---

## B1. Context Compaction — Memory Flush Before Pruning

### Overview

Before the agent loop compacts/prunes conversation history, execute a silent internal "memory flush" turn where the agent is prompted to identify and save important facts to the active memory backend. Prevents knowledge loss at compaction boundaries.

### Problem Statement

MakoClaw auto-summarizes sessions when history exceeds token/message thresholds. The summarization may lose specific facts, decisions, or preferences the user mentioned. OpenClaw adds a silent flush turn before compaction to preserve important context.

### Proposed Solution

In `pkg/session/` compaction logic, before pruning:

```go
// Before compaction: silent memory flush
func (s *Session) flushMemoryBeforeCompaction(ctx context.Context, agentLoop AgentLoop, memBackend MemoryBackend) error {
    flushPrompt := `[INTERNAL] You are about to lose earlier conversation history.
    Review the history and call memory tools to save any important facts,
    preferences, or decisions. Do not respond to the user. This is a silent turn.`

    // Execute agent turn with flushPrompt, memory tools only (no other tools)
    // Return without delivering response to user
}
```

### Functional Requirements

- Memory flush triggered when history reaches 80% of compaction threshold (configurable)
- Flush is a silent turn: no response delivered to user, no channel message sent
- Only memory write tools available during flush turn
- Flush prompt is not added to history (does not affect token count)
- Configurable: `session.memory_flush_enabled: true` (default)
- Flush timeout: 30s max (does not block session)

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| CF-R1 | Silent flush turn before compaction | P0 |
| CF-R2 | Flush at 80% threshold (configurable) | P0 |
| CF-R3 | Memory-tools-only during flush | P0 |
| CF-R4 | Flush configurable on/off | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/sessions/` — compaction strategy)
**Effort estimate**: Backend 1 week · **Total: 1 week**

---

## B2. Doctor Command — Deep Security Audits

### Overview

Extend the existing `doctor` command with deep audit capabilities: security checks, config drift detection, network probing, symlink escape detection, and auto-fix mode.

### Problem Statement

MakoClaw has a `doctor` command (`pkg/doctor/`) with basic diagnostics. OpenClaw's doctor runs a comprehensive security audit covering filesystem permissions, tool blast radius, config schema drift, and network health.

### Proposed Solution

New checks to add to `pkg/doctor/doctor.go`:

```go
// Security checks
type SecurityAudit struct{}
func (a *SecurityAudit) CheckFilesystemPerms() []Finding  // world-writable files, config readable
func (a *SecurityAudit) CheckToolBlastRadius() []Finding  // tools with destructive potential
func (a *SecurityAudit) CheckSymlinkEscape() []Finding    // workspace symlinks pointing outside
func (a *SecurityAudit) CheckOpenBinding() []Finding      // web server bound without auth

// Config drift
type ConfigDriftCheck struct{}
func (c *ConfigDriftCheck) CheckSchemaBaseline() []Finding  // unknown fields in config.json
func (c *ConfigDriftCheck) CheckDeprecatedFields() []Finding

// Network probe
type NetworkProbe struct{}
func (n *NetworkProbe) ProbeGateway() Finding              // HTTP health check on gateway port
func (n *NetworkProbe) ProbeChannelConnectivity() []Finding // can channels reach their APIs
```

CLI: `makoclaw doctor --deep --fix --json`
- `--deep`: runs all checks including security audit and network probes
- `--fix`: auto-fix safe issues (file permissions, deprecated config fields)
- `--json`: machine-readable output

### Functional Requirements

- `--deep` flag triggers security audit suite
- Filesystem perms: flag world-writable workspace, config.json readable by others
- Tool blast radius: identify tools with `exec`, `write_file`, `delete` that lack permission restrictions
- Symlink escape: scan workspace directory for symlinks pointing outside workspace root
- Open binding: warn if web server binds `0.0.0.0` without JWT auth enabled
- Config drift: compare config.json against known schema, report unknown/deprecated fields
- Network probe: test gateway reachability and channel API connectivity
- `--fix`: auto-correct filesystem permissions; prompt for others
- `--json`: structured output for CI integration

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| DC-R1 | `--deep` flag with security audit suite | P0 |
| DC-R2 | Filesystem permission checks | P0 |
| DC-R3 | Tool blast radius audit | P0 |
| DC-R4 | Symlink escape detection | P1 |
| DC-R5 | Open binding check | P0 |
| DC-R6 | Config drift detection | P1 |
| DC-R7 | `--fix` auto-remediation | P1 |
| DC-R8 | `--json` structured output | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/doctor/`)
**Effort estimate**: Backend 2 weeks · **Total: 2 weeks**

---

## B3. Cron — Stagger + Delivery Modes

### Overview

Add deterministic stagger per cron job (to avoid thundering herd at the top of the hour), multiple session targets, and three delivery modes: `announce` (send to channel), `webhook` (HTTP POST), and `none` (silent execution).

### Problem Statement

MakoClaw's cron scheduler runs jobs at exact scheduled times. Multiple jobs at `:00` cause a spike. There is no webhook delivery mode and no silent execution mode. OpenClaw implements stagger and delivery modes as standard cron features.

### Proposed Solution

```go
type CronDeliveryMode string

const (
    DeliveryAnnounce CronDeliveryMode = "announce"  // send result to channel (current behavior)
    DeliveryWebhook  CronDeliveryMode = "webhook"   // HTTP POST to configured URL
    DeliveryNone     CronDeliveryMode = "none"      // silent: execute, no delivery
)

type CronJob struct {
    // existing fields...
    Stagger      time.Duration    // deterministic offset: hash(jobID) % max_stagger
    SessionTarget string          // main | isolated | current | session:<id>
    Delivery     CronDeliveryMode
    WebhookURL   string           // used when Delivery == webhook
}
```

### Functional Requirements

- `stagger` field in cron job config (duration string, e.g. `"5m"`)
- Deterministic stagger: `hash(job_id) % stagger_max` — same job always offsets by same amount
- `session_target` field: `main | isolated | current | session:<custom-id>`
- `delivery` field: `announce | webhook | none`
- Webhook delivery: HTTP POST with job result as JSON body, configurable URL per job
- `none` delivery: agent turn executes but result is not sent to any channel
- Web UI Cron builder: delivery mode selector and webhook URL field

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| CR-R1 | Deterministic stagger per job | P0 |
| CR-R2 | `session_target` field with 4 modes | P1 |
| CR-R3 | `delivery` field: announce/webhook/none | P0 |
| CR-R4 | Webhook delivery: HTTP POST with result JSON | P1 |
| CR-R5 | Web UI Cron builder: delivery mode + webhook URL | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/cron/`)
**Effort estimate**: Backend 1 week · Frontend 0.5 weeks · **Total: 1.5 weeks**

---

## B4. Tool Profiles (messaging / developer / minimal)

### Overview

Add three named tool permission profiles that users can select instead of configuring individual tool permissions. Profiles are composable with `allow`, `deny`, and `alsoAllow` overrides.

### Problem Statement

MakoClaw has role-based tool permissions but no named profiles. Users must configure individual tool permissions from scratch. OpenClaw's profiles give users a safe starting point and reduce configuration complexity.

### Proposed Solution

```go
type ToolProfile string

const (
    ProfileMessaging  ToolProfile = "messaging"   // no exec, no fs write — safe for public bots
    ProfileDeveloper  ToolProfile = "developer"   // exec enabled, full fs — for trusted users
    ProfileMinimal    ToolProfile = "minimal"     // read-only fs + web search only
)

type ToolPermissionConfig struct {
    Profile   ToolProfile    // base profile
    AlsoAllow []string       // add tools to profile
    Deny      []string       // remove tools from profile
    ExecSecurity string      // deny | ask | full (applies within developer profile)
}
```

Profiles applied at agent loop initialization. `AlsoAllow` and `Deny` applied on top of profile.

### Functional Requirements

- `tool_permissions.profile` config field accepting `messaging | developer | minimal`
- `messaging` profile: no `exec`, no `write_file`, no `delete_*` tools
- `developer` profile: all tools enabled; `exec.security` sub-field: `deny | ask | full`
- `minimal` profile: `read_file`, `list_dir`, `web_search`, `web_fetch`, `query_knowledge` only
- `also_allow` and `deny` arrays for per-user overrides on top of profile
- Web UI Settings: profile selector dropdown with description of each profile
- Profiles backwards compatible with existing individual tool permissions config

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| TP-R1 | `profile` field with three named profiles | P0 |
| TP-R2 | `messaging` profile (no exec, no fs write) | P0 |
| TP-R3 | `developer` profile with `exec.security` sub-field | P0 |
| TP-R4 | `minimal` profile | P0 |
| TP-R5 | `also_allow` and `deny` composability | P1 |
| TP-R6 | Web UI profile selector | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/tools/` — tool policy)
**Effort estimate**: Backend 1 week · Frontend 0.5 weeks · **Total: 1.5 weeks**

---

## B5. Session Cross-Messaging (sessions_send tool)

### Overview

A new agent tool that sends a message to another active session and waits for its reply. Enables lightweight peer-to-peer agent coordination without spawning a new subagent — ideal for querying a specialist session that already has loaded context.

### Problem Statement

MakoClaw's `spawn` tool starts a new subagent conversation from scratch. There is no way to send a message to an already-running session and get a response. OpenClaw's `sessions_send` tool enables this ping-pong pattern.

### Proposed Solution

```go
type SessionsSendTool struct {
    *BaseTool
    SessionManager SessionManager  // access to active sessions
}

// Tool schema: sessions_send(session_id: string, message: string, timeout_sec?: int) -> string
func (t *SessionsSendTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    sessionID := args["session_id"].(string)
    message   := args["message"].(string)
    timeout   := time.Duration(args["timeout_sec"].(float64)) * time.Second

    // Route message into target session, wait for response, return text
}
```

Session targeting: `session_id` can be a session key (`cli:default`, `telegram:12345`) or a specialist name (`@researcher`).

### Functional Requirements

- `sessions_send(session_id, message, timeout_sec?)` tool registered in AgentLoop
- Target session must be active (not terminated) — returns error if not found
- Timeout default 60s; configurable per-call
- Response: text content of the target session's reply
- The target session is unaware it is being queried via cross-session (appears as normal message)
- Circular call detection: session cannot send to itself

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| SS-R1 | `sessions_send` tool with session_id targeting | P0 |
| SS-R2 | Active session lookup | P0 |
| SS-R3 | Configurable timeout | P1 |
| SS-R4 | Circular call detection | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/sessions/` — sessions_send tool)
**Effort estimate**: Backend 1.5 weeks · **Total: 1.5 weeks**

---

## B6. Onboarding — CLI Daemon Wizard

### Overview

Add a guided CLI onboarding wizard (`makoclaw onboard --install-daemon`) that walks new users through first-time setup, installs the system daemon (launchd/systemd), configures channels, selects a model, and shows security posture — all from the terminal.

### Problem Statement

MakoClaw has a web-based setup wizard but no CLI-guided installation. Users who prefer CLI or want to configure MakoClaw on a headless server must edit `config.json` manually. OpenClaw's onboarding wizard dramatically reduces time-to-first-agent.

### Proposed Solution

```go
// cmd/makoclaw/main.go — new command
func onboardCmd() {
    wizard := onboarding.NewWizard()
    wizard.Run(ctx)
}
```

Steps:
1. Welcome + prerequisite check (Go, OS, disk space)
2. Provider selection + API key input (hidden input, validated)
3. Model selection (fetched from provider)
4. Channel configuration (user picks which channels to enable)
5. Security posture review (show what is enabled, warn about exec/dm_policy)
6. Daemon install: `--install-daemon` flag triggers launchd (macOS) or systemd (Linux)
7. Health check + first agent turn

### Functional Requirements

- `makoclaw onboard` interactive CLI wizard
- `--install-daemon` flag: generate and install launchd plist (macOS) or systemd unit (Linux)
- Hidden input for API keys (no echo to terminal)
- Provider validation: test API key before proceeding
- Model list fetched live from selected provider
- Security posture summary printed before daemon install
- `--non-interactive` flag for scripted setup (read from env vars)
- Skip steps if `config.json` already exists (resume mode)

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| OB-R1 | Interactive CLI wizard with step-by-step flow | P0 |
| OB-R2 | `--install-daemon` for launchd/systemd | P0 |
| OB-R3 | Hidden API key input with validation | P0 |
| OB-R4 | Security posture summary | P1 |
| OB-R5 | `--non-interactive` scripted mode | P1 |
| OB-R6 | Resume mode when config.json exists | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/cli/onboard.ts`)
**Effort estimate**: Backend 2 weeks · **Total: 2 weeks**

---

## B7. Web Search — Additional Providers

### Overview

Extend web search from Brave (current) to include Exa (semantic), Tavily (research-optimized), Perplexity (AI-powered), and Firecrawl (AI web scraping with structured extraction).

### Problem Statement

MakoClaw uses Brave Search as its sole web search provider. Different search tasks benefit from different providers: Exa for semantic/conceptual queries, Tavily for research, Firecrawl for structured data extraction from specific pages.

### Proposed Solution

Extend the existing `web.go` tool file with additional providers under the same `WebSearchTool` interface:

```go
type WebSearchProvider interface {
    Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
    Name() string
}

// New providers
type ExaProvider struct { APIKey string }        // semantic search
type TavilyProvider struct { APIKey string }     // research-optimized
type PerplexityProvider struct { APIKey string } // AI-synthesized answers
type FirecrawlProvider struct { APIKey string }  // structured web scraping
```

Provider selected automatically based on configured API keys; user can override with `provider` param in tool call.

### Functional Requirements

- Exa provider: `exa.api_key` in config, semantic search mode
- Tavily provider: `tavily.api_key`, `search_depth: basic|advanced` option
- Perplexity provider: `perplexity.api_key`, returns synthesized answer + sources
- Firecrawl provider: `firecrawl.api_key`, structured extraction from specific URLs
- Auto-select: use first configured provider; prefer Brave if multiple configured
- `web_search(query, provider?)` — optional provider override
- Web UI Settings: search provider selection and API key input

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| WS-R1 | Exa semantic search provider | P1 |
| WS-R2 | Tavily research provider | P1 |
| WS-R3 | Perplexity AI-synthesized answers | P2 |
| WS-R4 | Firecrawl structured scraping | P2 |
| WS-R5 | Auto-select + per-call override | P1 |

**Inspired by**: OpenClaw (`openclaw-main/extensions/*-search`)
**Effort estimate**: Backend 1.5 weeks · **Total: 1.5 weeks**

---

## B8. LLM Providers — Expanded Coverage

### Overview

Add the most impactful LLM providers present in OpenClaw that MakoClaw does not yet support: DeepSeek (cost-effective reasoning), Amazon Bedrock (enterprise AWS integration), LiteLLM proxy (universal proxy for 100+ providers), and Together AI (open-source model hosting).

### Problem Statement

MakoClaw supports ~10-15 LLM providers. OpenClaw supports 25+. The gap matters for enterprise users (Bedrock), cost-conscious users (DeepSeek), and operators running private infrastructure (LiteLLM proxy).

### Proposed Solution

All use the existing `HTTPProvider` since they implement OpenAI-compatible APIs:

```go
// In CreateProvider() selection logic:
case "deepseek":
    return &HTTPProvider{
        BaseURL:  "https://api.deepseek.com/v1",
        APIKey:   cfg.Providers.DeepSeek.APIKey,
        Model:    "deepseek-reasoner",
    }

case "together":
    return &HTTPProvider{
        BaseURL:  "https://api.together.xyz/v1",
        APIKey:   cfg.Providers.Together.APIKey,
    }

case "litellm":
    return &HTTPProvider{
        BaseURL:  cfg.Providers.LiteLLM.BaseURL,  // self-hosted URL
        APIKey:   cfg.Providers.LiteLLM.APIKey,
    }
```

Amazon Bedrock requires a separate `BedrockProvider` (AWS SDK, not OpenAI-compatible natively — use Bedrock's OpenAI-compatible endpoint if available, else implement `InvokeModel`).

### Functional Requirements

- DeepSeek provider: `providers.deepseek.api_key`, models: `deepseek-chat`, `deepseek-reasoner`
- Together AI provider: `providers.together.api_key`
- LiteLLM proxy: `providers.litellm.base_url` + optional `api_key`
- Amazon Bedrock: `providers.bedrock.region`, `aws_access_key_id`, `aws_secret_access_key`
- All providers appear in `/api/v1/models` response
- Web UI provider configuration for all new providers

| Req ID | Requirement | Priority |
|--------|-------------|----------|
| LP-R1 | DeepSeek provider | P1 |
| LP-R2 | Together AI provider | P1 |
| LP-R3 | LiteLLM proxy provider (self-hosted) | P1 |
| LP-R4 | Amazon Bedrock provider | P2 |
| LP-R5 | All providers in models API + Web UI config | P1 |

**Inspired by**: OpenClaw (`openclaw-main/src/providers/`)
**Effort estimate**: Backend 1.5 weeks · Frontend 0.5 weeks · **Total: 2 weeks**

---

### Phase 8 — Priority Execution Order

Given MakoClaw's positioning as a resource-efficient framework, recommended implementation sequence:

**Sprint 1 — Security & Memory (4 weeks)**
1. B1. Context Compaction Memory Flush (1 week) — low effort, high value, no UI needed
2. 16. DM Pairing Policy (3 weeks) — security gap, foundational for other features
3. 17. Security Hooks (3 weeks) — can run parallel with DM Pairing

**Sprint 2 — Model Control & Tools (4 weeks)**
4. 22. Thinking Level per Session (1.5 weeks) — very low effort, high user demand
5. 24. llm_task Tool (1 week) — minimal scope, directly useful
6. B4. Tool Profiles (1.5 weeks) — no new concepts, builds on existing permissions
7. B3. Cron Stagger + Delivery (1.5 weeks) — small, self-contained

**Sprint 3 — Search, Providers & Onboarding (5 weeks)**
8. B7. Web Search Providers (1.5 weeks)
9. B8. LLM Providers (2 weeks)
10. B6. CLI Onboarding Wizard (2 weeks)
11. B2. Doctor Deep Audits (2 weeks)

**Sprint 4 — Voice & Canvas (8 weeks)**
12. 14. Voice Wake + Talk Mode (6 weeks)
13. 18. Canvas / A2UI (5 weeks)

**Sprint 5 — Memory & Cross-Session (9 weeks)**
14. 15. Pluggable Memory Backends (7 weeks)
15. B5. Session Cross-Messaging (1.5 weeks)
16. 23. Image Generation (3 weeks)

**Sprint 6+ — Infrastructure (ongoing)**
17. 21. Channel Action Framework
18. 25. Multi-Account Channel Support
19. 26. Additional Channels (per channel)
20. 27. MakoHub Registry
21. 19. Device Pairing & Node Architecture
22. 20. Mobile Companion Apps
23. 28. Gmail Pub/Sub
24. 29. Lobster Integration
