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
