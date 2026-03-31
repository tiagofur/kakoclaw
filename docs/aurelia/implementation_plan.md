# SDD Exploration: Development Workspace

## Change Name: `dev-workspace`

---

## Current State

MakoClaw is a mature AI agent framework with:
- **Go backend** (24 packages in `pkg/`): Agent loop, providers, tools, sessions, channels, config, storage, cron, MCP
- **Vue 3 frontend** (22 views): Chat, Dashboard, Tasks, Marketing, Skills, Agents, etc.
- **Marketing as reference pattern**: Full-stack feature with sidebar list + detail panel + tabs — the exact pattern we need to replicate for Development

### Marketing Pattern (our blueprint)

```
Backend:
├── pkg/web/handlers_marketing.go         ← CRUD campaigns, briefs, strategy
├── pkg/web/handlers_marketing_analytics.go ← Analytics/charts data
├── pkg/web/handlers_marketing_audience.go  ← Contacts/lists/segments
├── pkg/web/handlers_marketing_templates.go ← Reusable templates
├── pkg/web/server.go                       ← Route registration
└── pkg/tools/                              ← social_post, marketing_campaign tools

Frontend:
├── src/views/Marketing.vue                 ← 1,946 lines, sidebar + detail + tabs
├── src/stores/marketingStore.js            ← Pinia store (API calls)
├── src/router/index.js                     ← Route entry under MainLayout
└── src/components/Layout/Sidebar.vue       ← Nav entry
```

### What we're building: The "Development" equivalent

Instead of campaigns/briefs/strategy/templates, we'll have:
- **Projects** (equivalent to campaigns) — workspaces where code lives
- **Sessions** — active coding sessions with an AI backend (OpenCode or Claude Code)
- **Bridge** — the Go↔CLI communication layer
- **Memory** — semantic memory with local embeddings
- **Agent definitions** — markdown-defined coding agents

---

## Affected Areas

### Go Backend (new packages)

| Package | Purpose | New/Modified |
|---------|---------|-------------|
| `pkg/bridge/` | Bridge process manager (Go↔CLI) | **NEW** |
| `pkg/bridge/protocol.go` | Request/Response/Event types | **NEW** |
| `pkg/memory/` | Semantic memory with embeddings | **NEW** |
| `pkg/web/handlers_dev.go` | REST API for dev workspace | **NEW** |
| `pkg/web/server.go` | Route registration | Modified |
| `pkg/agent/loop.go` | Wire bridge provider | Modified |
| `pkg/config/config.go` | Dev workspace config section | Modified |
| `pkg/session/manager.go` | Token-based auto-reset | Modified |

### Vue Frontend (new files)

| File | Purpose | New/Modified |
|------|---------|-------------|
| `src/views/DevelopmentView.vue` | Main dev workspace view | **NEW** |
| `src/stores/developmentStore.js` | Pinia store for dev API | **NEW** |
| `src/router/index.js` | Route entry | Modified |
| `src/components/Layout/Sidebar.vue` | Nav entry | Modified |

### TypeScript Bridge

| File | Purpose | New/Modified |
|------|---------|-------------|
| `bridge/index.ts` | Claude Code SDK wrapper | **NEW** (project root) |
| `bridge/package.json` | Bridge dependencies | **NEW** |

---

## Approaches

### Approach A — Bridge as LLM Provider (API-level integration)

The Bridge implements `providers.LLMProvider` — MakoClaw controls tools, the Bridge only handles LLM calls.

| Aspect | Assessment |
|--------|-----------|
| Pros | Clean separation, uses existing tool chain, consistent with other providers |
| Cons | Loses Claude Code/OpenCode's native tool execution (filesystem, terminal) |
| Effort | Medium |

### Approach B — Bridge as Autonomous Executor (Aurelia-style) ⭐ RECOMMENDED

The Bridge manages the full execution (LLM + tools). MakoClaw sends a prompt, the Bridge runs Claude Code/OpenCode end-to-end, and streams events back (tool_use, assistant, result).

| Aspect | Assessment |
|--------|-----------|
| Pros | Full CLI power, native tool execution, session resume, proven by Aurelia |
| Cons | Parallel tool system (Bridge tools vs MakoClaw tools), needs clear boundary |
| Effort | Medium-High |

### Approach C — Hybrid (Provider + Bridge modes)

Config-driven: user chooses per-project whether to use Bridge mode (autonomous CLI) or Provider mode (MakoClaw-controlled).

| Aspect | Assessment |
|--------|-----------|
| Pros | Maximum flexibility, accommodates both use cases |
| Cons | More complexity, two code paths to maintain |
| Effort | High |

### Recommendation: **Approach B** (Autonomous Executor)

> [!IMPORTANT]
> For local programming, you want the CLI (opencode/claude-code) to have direct filesystem and terminal access. MakoClaw shouldn't duplicate that — it should ORCHESTRATE and PRESENT the results.

The Dev Workspace becomes an **orchestration + visualization layer** over the CLI backends, not a replacement for them.

---

## Proposed Implementation Phases

### Phase 1 — Foundation (Go Backend Infrastructure)
1. `pkg/bridge/` — Bridge process manager, protocol, events
2. `pkg/memory/` — Semantic memory store with embeddings
3. `pkg/config/` — Dev workspace config section
4. `pkg/session/` — Token-based auto-reset enhancement
5. Tests for each component

### Phase 2 — Bridge TypeScript Layer
1. `bridge/index.ts` — Claude Code SDK wrapper (port from Aurelia)
2. `bridge/package.json` — Dependencies
3. `go:embed` integration for single-binary distribution
4. Bridge setup auto-bootstrap (npm install)
5. Tests: Bridge ping, query, event parsing

### Phase 3 — REST API Handlers
1. `pkg/web/handlers_dev.go` — Projects CRUD, sessions, bridge control
2. `pkg/web/handlers_dev_memory.go` — Memory search, inject, manage
3. Route registration in `server.go`
4. Tests for each handler

### Phase 4 — Vue Frontend: Development View
1. `DevelopmentView.vue` — Main view with sidebar + detail + tabs
2. `developmentStore.js` — Pinia store
3. Router + Sidebar integration
4. Tabs: Terminal, Files, Memory, Sessions, Settings

### Phase 5 — OpenCode Integration
1. Bridge adapter for OpenCode (investigate SDK/CLI protocol)
2. Config: provider selector (claude-code vs opencode)
3. Tests for OpenCode bridge

### Phase 6 — Polish & Integration
1. Agent markdown registry for dev-specific agents
2. Cron integration for scheduled dev tasks
3. E2E tests (Playwright)
4. Documentation updates

---

## View Design: Development Workspace

```
┌──────────────┬───────────────────────────────────────────────┐
│  PROJECTS    │  Project: my-app                              │
│              │  ┌─────┬────────┬────────┬─────────┬────────┐ │
│ ▸ my-app     │  │Term │ Files  │ Memory │Sessions │Settings│ │
│   kakoclaw   │  ├─────┴────────┴────────┴─────────┴────────┤ │
│   website    │  │                                           │ │
│              │  │  Terminal Tab:                             │ │
│  ──────────  │  │  ┌─────────────────────────────────────┐  │ │
│  + New       │  │  │ [Bridge Output Stream]              │  │ │
│              │  │  │ > User: fix the auth bug             │  │ │
│  BACKEND     │  │  │ < Tool: Read auth/handler.go        │  │ │
│  ⚡ Claude    │  │  │ < Tool: Write auth/handler.go       │  │ │
│  ⚙ OpenCode  │  │  │ < Assistant: Fixed the bug...       │  │ │
│              │  │  └─────────────────────────────────────┘  │ │
│  STATUS      │  │  ┌─────────────────────────────────────┐  │ │
│  🟢 Bridge OK │  │  │ [Input]  Type a message...    [Send]│  │ │
│  📊 42K toks  │  │  └─────────────────────────────────────┘  │ │
│              │  │                                           │ │
└──────────────┴───────────────────────────────────────────────┘
```

### Tab breakdown:

| Tab | Content |
|-----|---------|
| **Terminal** | Live bridge event stream + chat input (primary interaction) |
| **Files** | Project file browser (read from bridge's cwd) |
| **Memory** | Semantic memory search + management |
| **Sessions** | Session list, token usage, resume/reset controls |
| **Settings** | Project config: backend (claude-code/opencode), model, cwd, agents |

### Color theme

Marketing uses **pink→violet gradient**. Development will use **cyan→blue gradient** (programming vibes, clearly distinct):

```css
/* Development palette */
--dev-primary: from-cyan-500 to-blue-500;
--dev-accent: cyan-400;
--dev-glow: cyan-500/25;
```

---

## Risks

1. **OpenCode SDK uncertainty** — Need to investigate if OpenCode has a programmatic API. If CLI-only, we'll need to parse stdout which is fragile.
2. **Binary size** — Hugot ONNX model (~23MB) significantly increases binary. Must be opt-in.
3. **Windows pipe edge cases** — stdin/stdout communication with child processes on Windows has known quirks with buffering.
4. **Bridge process lifecycle** — Must handle crashes, restarts, and cleanup gracefully.
5. **Scope creep** — 6 phases is substantial. Must stay disciplined with incremental delivery.

---

## Open Questions

> [!IMPORTANT]
> 1. **OpenCode**: ¿Tiene SDK programático o solo CLI? Esto define Phase 5 completamente.
> 2. **Memory embeddings**: ¿Hacemos opt-in (config flag) dado el constraint de RAM/binary size?
> 3. **View naming**: ¿"Development", "Programming", "Code", "Dev Studio"? Necesito tu preferencia.
> 4. **Bridge first or Frontend first?** Phase 1-2 (backend) before Phase 4 (frontend), or do frontend skeleton first to see the UI?

---

## Ready for Proposal

**Yes** — pending answers to the open questions above. The exploration covered:
- ✅ Marketing pattern fully mapped as blueprint
- ✅ Bridge architecture from Aurelia analyzed
- ✅ Memory, Session, Cron gaps identified
- ✅ 6 implementation phases defined
- ✅ UI design wireframed
- ✅ Risks documented
