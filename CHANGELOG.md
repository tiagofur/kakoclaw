# Changelog

All notable changes to MakoClaw are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- PRD-NEW-FEATURES.md with 13 new feature proposals
- Comprehensive security audit documentation
- Competitor analysis (OpenClaw, NanoClaw, PicoClaw, ZeroClaw)
- Real-time tool call visibility with smart collapse (expanded during streaming, collapsed in historical messages)
- Extended Thinking visualization (Claude only) with opt-in toggle in Settings
- Per-agent tool call visibility in TeamActivityPanel for multi-agent scenarios
- AgentStatusIndicator now shows active tool name beneath the agent

### Changed
- ToolCallItem now uses computed state for expansion/collapse instead of direct mutation
- TeamActivityPanel refactored from watcher-based to computed from store (eliminates sync issues)
- AgentActivityItem displays in-flight tool calls per agent

### Fixed
- Historical tool calls now collapsed by default (reduces visual noise)

### Technical Notes
- Added `ThinkingDelta` field to `StreamChunk` for extended thinking support
- Implemented `ChatStream()` in ClaudeProvider to emit thinking blocks
- Added `OnThinking` callback to agent streaming loop
- Added `ExtendedThinking` user preference with persistence to SQLite
- Added WS event `thinking_delta` gated by user preference
- Added API endpoints `GET/PUT /api/v1/user/config` for user preferences

### Security
- Identified 7 critical/high security issues (see BUGS-KNOWN-ISSUES.md)

---

## [0.9.0] - 2026-03-04

### Added
- **Orchestrator Agent**: Multi-agent system with specialist delegation, reporting, and team collaboration
- **Skill Analytics**: Usage tracking, trending marketplace, and installed badges
- **Slack Integration**: App Token and Allow From fields in channel config
- **MessageBubble Component**: Chat messages with user/assistant roles, agent activities, tool calls, markdown rendering

### Fixed
- Normalize incomplete skill frontmatter to prevent 400 on save
- Slugified skill name when incrementing usage_count in analytics

---

## [0.8.0] - 2026-03-01

### Added
- **Multi-Agent Orchestration**: Task decomposition, specialist delegation, and collaboration
- **Advanced Skill Management API**: Skill generation, refinement, and installation
- **PWA Support**: Progressive Web App with offline capabilities
- **Chat Interface**: Streaming capabilities and activity tracking

### Changed
- Redesigned chat UI with agent activities visualization
- Enhanced session management with multi-user isolation

---

## [0.7.0] - 2026-02-27

### Added
- **UX Polish Phase**: Complete refinement of user experience
  - Dashboard glass morphism and responsive layout
  - Chat page fluid width and compact toolbar
  - Tasks page collapsible filters and glass effects
- **Frontend Design System**: Comprehensive component patterns

### Changed
- All views updated with consistent glass morphism styling
- Improved mobile responsiveness across all pages

---

## [0.6.0] - 2026-02-26

### Added
- **Session Management** (UX-A)
  - New `sessions` table with CRUD operations
  - Archive/restore sessions
  - Rename sessions inline
  - Fork sessions with full history
  - `DELETE /api/v1/chat/sessions/{id}` endpoint
  - `PATCH /api/v1/chat/sessions/{id}` endpoint

- **Cron Visual Builder** (UX-B)
  - 6 schedule types: Daily, Weekly, Monthly, Interval, One-time, Custom
  - Auto-generation of cron expressions from visual UI
  - Preview of next 3 executions
  - Timezone selector
  - Human-readable schedule display

- **UX Quick Fixes** (UX-C)
  - Toast notifications system-wide (replaced all alerts)
  - Copy button on assistant messages
  - Auto-refresh metrics every 30s
  - Task deletion with cascade
  - Fixed log field naming in TaskDetailsModal

---

## [0.5.0] - 2026-02-24

### Added
- **API Models Endpoint**: `GET /api/v1/models` returns configured providers and models
- **Model Selector**: Dropdown in chat top bar to select model per conversation
- **Token-by-Token Streaming**: Real-time response rendering
  - `StreamingLLMProvider` interface with `ChatStream()`
  - WebSocket protocol: `stream_start` → `stream` → `stream_end` → `ready`
  - Cursor animation during streaming
- **Regenerate Button**: Refresh last assistant response
- **Dark/Light Theme Toggle**: Persistent theme preference

### Changed
- Streaming fallback for providers without native streaming (Claude, Codex)

---

## [0.4.0] - 2026-02-22

### Added
- **Voice Input**: Hold-to-record with Groq Whisper STT
  - `POST /api/v1/voice/transcribe` endpoint
  - Microphone button in ChatView
- **MCP Client**: Model Context Protocol support
  - JSON-RPC 2.0 over STDIO
  - Dynamic tool discovery
  - Multi-server management
  - MCPView.vue for server status and reconnect
- **Knowledge Base / RAG**: SQLite FTS5 full-text search
  - Document upload with drag-and-drop
  - Automatic chunking with overlap
  - `query_knowledge` tool for agents
  - KnowledgeView.vue with search interface
- **Web Search Toggle**: Enable/disable web search per conversation
- **API Documentation**: OpenAPI 3.0.3 spec with Swagger UI at `/api/docs`

---

## [0.3.0] - 2026-02-18

### Added
- **Specialist Agents**: Domain-specific agents with filtered tools
- **Per-User Configuration**: User-specific API keys and settings
- **Settings View**: Frontend configuration management
- **User Storage Isolation**: Separate workspaces per user

### Changed
- Agent loop refactored for multi-user support
- Config merge behavior (user overrides global)

---

## [0.2.0] - 2026-02-14

### Added
- **Web Frontend**: Vue 3 + Vite + Tailwind CSS
  - Chat view with markdown rendering
  - Tasks view with Kanban board
  - Dashboard with metrics
  - Settings page
  - History view
- **WebSocket Support**: Real-time chat communication
- **JWT Authentication**: Login/signup with token refresh
- **Multi-User Support**: User accounts with isolated storage

### Changed
- Migrated from CLI-only to web + CLI hybrid

---

## [0.1.0] - 2026-02-10

### Added
- **Core Agent Loop**: Tool-calling LLM loop with iteration limits
- **9 Communication Channels**:
  - Telegram (with voice transcription)
  - Discord
  - Slack
  - WhatsApp (bridge)
  - Signal (CLI)
  - QQ
  - DingTalk
  - Feishu/Lark
  - MaixCam
- **10 LLM Providers**:
  - Anthropic Claude
  - OpenAI
  - OpenRouter
  - Groq
  - Ollama (local)
  - Zhipu
  - Gemini
  - Nvidia
  - Moonshot
  - VLLM
- **20+ Built-in Tools**:
  - File operations (read, write, edit, list)
  - Shell execution with security controls
  - Web search (Brave)
  - Web fetch
  - Task management
  - Email (SMTP)
  - Subagent spawning
  - Cron scheduling
- **Skills System**: Markdown-based extensions
- **Message Bus**: Pub/sub for channel communication
- **SQLite Storage**: Tasks, knowledge, history, metrics

### Technical
- Go 1.26 with <10MB RAM footprint
- <1s startup time
- Single binary deployment

---

## Project History

- **2026-02**: MakoClaw development begins (fork of PicoClaw concepts)
- **2026-01**: KakoClaw naming phase
- **2025-11**: Original PicoClaw inspiration

---

## Migration Notes

### From PicoClaw
MakoClaw maintains compatibility with PicoClaw configurations. Key differences:
- Config directory: `~/.MakoClaw/` (was `~/.picoclaw/`)
- Binary name: `makoclaw` (was `picoclaw`)
- Multi-user support (new)
- Web UI (new)

### From KakoClaw
Direct upgrade path. Rename config directory if needed.

---

## Links

- [GitHub Repository](https://github.com/sipeed/makoclaw)
- [Documentation](docs/README.md)
- [PRD New Features](PRD-NEW-FEATURES.md)
- [Roadmap](docs/ROADMAP.md)
- [Known Issues](BUGS-KNOWN-ISSUES.md)
