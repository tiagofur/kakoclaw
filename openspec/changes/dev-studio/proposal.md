# Proposal: Dev Studio

## Intent

Add a dedicated programming workspace to MakoClaw that orchestrates local AI coding CLIs (Claude Code, OpenCode) via a Bridge pattern. The existing LLM providers and agent system are UNTOUCHED — Dev Studio is a self-contained feature with its own execution path, isolated memory, and purpose-built UI.

## Scope

### In Scope
- Bridge process manager (Go↔CLI via NDJSON multiplexed stdin/stdout)
- Claude Code SDK TypeScript wrapper (embedded via `go:embed`)
- OpenCode CLI adapter (subprocess communication)
- Dev Studio isolated semantic memory (SQLite + local ONNX embeddings, opt-in)
- Session management with token-based auto-reset for Dev Studio sessions
- REST API handlers for Dev Studio (projects, sessions, bridge control, memory)
- Vue 3 "Dev Studio" view (sidebar + detail + tabs: Terminal, Files, Memory, Sessions, Settings)
- Tests after each component (Go unit tests, handler tests)

### Out of Scope
- Modifying existing LLM providers (ClaudeProvider, HTTPProvider, etc.)
- Replacing existing agent loop or tool system
- Multi-agent orchestration within Dev Studio (single CLI per session)
- Markdown agent registry (deferred — can add later as enhancement)
- Cron-triggered dev tasks (deferred)
- E2E Playwright tests (deferred to polish phase)

## Approach

**Bridge as Autonomous Executor**: Dev Studio sends prompts to a long-lived CLI process (Claude Code or OpenCode). The CLI handles everything (LLM + tools). Go only orchestrates lifecycle and streams events to the frontend via WebSocket.

```
User → Vue Dev Studio → REST API → Bridge Manager → CLI Process (Claude Code / OpenCode)
                 ↑                                          ↓
            WebSocket ←──── Event Stream (NDJSON) ──────────┘
```

**Memory isolation**: Dev Studio gets its own SQLite DB (`dev_memory.db`) separate from the main `database.db`. Embeddings are opt-in via config flag.

**Backend first**: Phases 1-3 (Go + Bridge + API) before Phase 4 (Vue frontend).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/bridge/` | **New** | Bridge process manager, protocol, events |
| `pkg/devmemory/` | **New** | Isolated semantic memory store |
| `bridge/` | **New** | TypeScript Bridge (Claude Code wrapper) |
| `pkg/web/handlers_dev.go` | **New** | REST API for Dev Studio |
| `pkg/web/handlers_dev_memory.go` | **New** | Memory management API |
| `pkg/web/server.go` | Modified | Route registration |
| `pkg/config/config.go` | Modified | `DevStudio` config section |
| `pkg/session/manager.go` | Modified | Token-based auto-reset |
| `src/views/DevStudioView.vue` | **New** | Frontend main view |
| `src/stores/devStudioStore.js` | **New** | Pinia store |
| `src/router/index.js` | Modified | Route entry |
| `src/components/Layout/Sidebar.vue` | Modified | Nav entry |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| OpenCode has no stable API | Medium | Fallback: parse CLI stdout, version-pin |
| Windows stdin/stdout pipe buffering | Medium | Use `cmd.StdinPipe()` with explicit flush |
| Hugot ONNX model bloats binary | High | Opt-in config flag, lazy download on first use |
| Bridge process crash mid-session | Medium | Auto-recovery with session resume (Aurelia pattern) |
| Scope creep across phases | Low | Strict phase gates, test after each task |

## Rollback Plan

All new code is in NEW packages/files (`pkg/bridge/`, `pkg/devmemory/`, `bridge/`, `handlers_dev*.go`, `DevStudioView.vue`). Rollback = delete these files and revert the 4 modified files (`server.go`, `config.go`, `session/manager.go`, `router/index.js`, `Sidebar.vue`). No existing functionality is altered.

## Dependencies

- `knights-analytics/hugot` v0.6+ (for local ONNX embeddings — opt-in)
- `@anthropic-ai/claude-agent-sdk` (npm, for Bridge TypeScript)
- Node.js runtime on PATH (for Bridge process)
- OpenCode binary on PATH (for OpenCode mode)

## Success Criteria

- [ ] Bridge can start, ping, execute queries, and stream events for Claude Code
- [ ] Bridge can start and execute queries for OpenCode
- [ ] Dev Studio memory is isolated from main database
- [ ] Semantic search returns relevant results from Dev Studio memory
- [ ] Session auto-resets when token threshold exceeded
- [ ] REST API serves projects, sessions, bridge status, memory
- [ ] Vue Dev Studio view renders with functional Terminal tab
- [ ] WebSocket streams bridge events in real-time to frontend
- [ ] All Go packages have passing tests
- [ ] Bridge process recovers from crashes automatically
