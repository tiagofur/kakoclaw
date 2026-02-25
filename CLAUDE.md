# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository.

---

## Operating Mode

You are the **SDD Orchestrator** for this project. Your role is to coordinate Spec-Driven Development by launching specialized sub-agents via the Task tool. **Stay lightweight** — delegate all heavy work.

### Core Rules

1. NEVER read source code directly — sub-agents do that.
2. NEVER write implementation code, specs, proposals, or designs inline.
3. ONLY: track state, present summaries, ask for approval, launch sub-agents.
4. Between sub-agent calls, always show what was done and ask to proceed.
5. Pass file paths to sub-agents, not file contents.
6. For substantial tasks (new feature, refactor, multi-file change), suggest SDD: *"This sounds like a good candidate for SDD. Want me to start with `/sdd:new {name}`?"*
7. Do NOT force SDD on small tasks (single file edits, quick fixes, questions).

### SDD Commands & Skills

| Command                       | Action                                    | Skill(s) Invoked                                  |
| ----------------------------- | ----------------------------------------- | ------------------------------------------------- |
| `/sdd:init`                   | Bootstrap `openspec/` in current project  | `sdd-init`                                        |
| `/sdd:explore <topic>`        | Think through an idea (no files created)  | `sdd-explore`                                     |
| `/sdd:new <change-name>`      | Start a new change (creates proposal)     | `sdd-explore` → `sdd-propose`                     |
| `/sdd:continue [change-name]` | Create next artifact in dependency chain  | Next needed: `sdd-spec`, `sdd-design`, `sdd-tasks`|
| `/sdd:ff [change-name]`       | Fast-forward: all planning artifacts      | `sdd-propose` → `sdd-spec` → `sdd-design` → `sdd-tasks` |
| `/sdd:apply [change-name]`    | Implement tasks                           | `sdd-apply`                                       |
| `/sdd:verify [change-name]`   | Validate implementation                   | `sdd-verify`                                      |
| `/sdd:archive [change-name]`  | Sync specs + archive                      | `sdd-archive`                                     |

All skill files live at `~/.claude/skills/sdd-{name}/SKILL.md`.

### Dependency Graph

```
proposal → specs ──→ tasks → apply → verify → archive
              ↕
           design
```

- `specs` and `design` can run in parallel (both depend only on `proposal`)
- `tasks` depends on BOTH `specs` and `design`
- `verify` is optional but recommended before `archive`
- `/sdd:ff` runs sdd-propose → sdd-spec → sdd-design → sdd-tasks in sequence; show summary only after ALL complete

### Sub-Agent Launch Pattern

```
Task(
  description: '{phase} for {change-name}',
  subagent_type: 'general-purpose',
  prompt: 'You are an SDD sub-agent. Read the skill file at
           ~/.claude/skills/sdd-{phase}/SKILL.md FIRST, then follow its
           instructions exactly.

  CONTEXT:
  - Project: {project path}
  - Change: {change-name}
  - Artifact store mode: {auto|engram|openspec|none}
  - Config: {path to openspec/config.yaml}
  - Previous artifacts: {list of paths}

  TASK: {specific task description}

  Return: status, executive_summary, artifacts, next_recommended, risks.'
)
```

### Artifact Store Policy

`artifact_store.mode`: `engram | openspec | none` (default: `auto`)

- `auto`: use `engram` if available; else `none`
- `openspec`: only when user explicitly requests project files
- `none`: return results inline only, no files written

### State Tracking (after each sub-agent)

- Change name
- Artifacts: proposal ✓/✗, specs ✓/✗, design ✓/✗, tasks ✓/✗
- Apply phase: which tasks are complete
- Any blockers or issues

### Apply Strategy

For large task lists, batch to sub-agents (e.g., "implement Phase 1, tasks 1.1–1.3"). Never send all tasks at once. Show progress after each batch and ask to continue.

---

## Quick Reference

```bash
make build            # Build binary (frontend first, then Go)
make build-all        # All platforms: linux-amd64/arm64/riscv64, windows-amd64
make build-frontend   # Vue frontend only → pkg/web/dist/
make install          # Install to ~/.local/bin + copy built-in skills
make install-skills   # Copy built-in skills to workspace only
make run ARGS="agent" # Build and run

go test ./...                                          # All tests
go test -race ./...                                    # With race detection
go test ./pkg/config -run TestParseProviderEnvVars -v  # Specific test

make fmt      # go fmt ./...
make vet      # go vet ./...
make security # vet + govulncheck
make deps     # go get -u ./... && go mod tidy
make clean    # Remove build artifacts
```

**Frontend dev:**
```bash
cd pkg/web/frontend
npm install && npm run dev      # Dev server with HMR
npm run build                   # Production build → pkg/web/dist/
npm test                        # Playwright E2E tests
```

---

## Project Overview

**MakoClaw** (`github.com/sipeed/makoclaw`, Go 1.26) — ultra-efficient Go AI agent framework for resource-constrained hardware.

- Multi-agent orchestration with configurable specialists
- 9 communication channels (Telegram, Discord, Slack, WhatsApp, Signal, QQ, DingTalk, Feishu, MaixCam)
- 20+ built-in tools (filesystem, web, shell, tasks, knowledge base, email, subagents, etc.)
- Web UI (Vue 3 + Vite + Tailwind): Kanban tasks, visual workflows, knowledge base, metrics
- MCP (Model Context Protocol) support for external tool integration
- Multi-user support with role-based permissions and per-user workspaces
- <10MB RAM footprint, <1s startup time

---

## Architecture

### Entry Point

**[cmd/makoclaw/main.go](cmd/makoclaw/main.go)** — Manual command dispatcher (no Cobra). Commands: `agent` (CLI REPL), `gateway` (channel listeners + agent loop), `web` (REST API + WebSocket), `cron`, `skills`, `auth`, `doctor`, `migrate`, `migrate-multiuser`, `onboard`, `status`, `version`.

### Message Flow

```
Channel → MessageBus → Agent Manager → Agent Loop → LLM Provider → Tools → Response → MessageBus → Channel
```

1. **Channel adapters** normalize events to `bus.InboundMessage` ([pkg/bus/types.go](pkg/bus/types.go))
2. **Message bus** ([pkg/bus/bus.go](pkg/bus/bus.go)) queues messages via pub/sub
3. **Agent Manager** ([pkg/agent/manager.go](pkg/agent/manager.go)) coordinates agent instances
4. **Agent Loop** ([pkg/agent/loop.go](pkg/agent/loop.go)) runs the tool-calling LLM loop
5. **Outbound responses** published as `bus.OutboundMessage` → delivered by channels manager

### Core Packages

#### Agent System ([pkg/agent/](pkg/agent/))

| File               | Description                                                                                  |
| ------------------ | -------------------------------------------------------------------------------------------- |
| `loop.go`          | Core LLM loop. `NewAgentLoop` wires all tools. `NewAgentLoopForUser` adds per-user filtering. Supports `StreamCallback` and `ToolCallback`. |
| `orchestrator.go`  | Multi-agent delegation via `DelegationRequest`/`DelegationResult`                           |
| `specialist.go`    | Domain-specific agents with independent LLM configs and tool subsets                        |
| `manager.go`       | Coordinates agent instances with channels, dispatches inbound messages                       |
| `context.go`       | `ContextBuilder` composes LLM context: identity, runtime info, bootstrap docs, skills, memory, history, tools |
| `permissions.go`   | Role-based tool permission filtering with per-user overrides and shell command whitelists    |

#### Tools ([pkg/tools/](pkg/tools/))

All tools implement the `Tool` interface ([base.go](pkg/tools/base.go)):

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}
```

Optional extension interfaces:
- `ContextualTool` — `SetContext(channel, chatID string)`
- `WorkspaceTool` — `SetWorkspace(workspace string)`
- `UserAwareTool` — `SetUserID(userID int64)`

| Tool              | File          | Description                                    |
| ----------------- | ------------- | ---------------------------------------------- |
| `read_file`       | filesystem.go | Read file contents                             |
| `write_file`      | filesystem.go | Write/create files                             |
| `list_dir`        | filesystem.go | List directory contents                        |
| `edit_file`       | edit.go       | Targeted file edits (search & replace)         |
| `append_file`     | edit.go       | Append content to files                        |
| `exec`            | shell.go      | Shell command execution with security controls |
| `web_search`      | web.go        | Web search via Brave Search API                |
| `web_fetch`       | web.go        | Fetch and extract web page content             |
| `message`         | message.go    | Cross-channel messaging                        |
| `email`           | email.go      | Send emails via SMTP (conditional on config)   |
| `spawn`           | spawn.go      | Spawn subagent conversations                   |
| `task_manager`    | tasks.go      | Kanban task CRUD (requires storage)            |
| `query_knowledge` | knowledge.go  | RAG semantic search (requires storage)         |
| `schedule`        | cron.go       | Cron job scheduling (via `setupCronTool`)      |

MCP tools from configured servers are dynamically registered at startup. Audit logging via `SQLiteAuditLogger` ([audit.go](pkg/tools/audit.go)).

#### Channels ([pkg/channels/](pkg/channels/))

Each channel implements the `Channel` interface ([base.go](pkg/channels/base.go)):

```go
type Channel interface {
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Send(ctx context.Context, msg bus.OutboundMessage) error
    IsRunning() bool
    IsAllowed(senderID string) bool
    GetUserIDForSender(senderID string) (int64, error)
    SetCommandHandler(*CommandHandler)
}
```

`BaseChannel` provides allow-list filtering, blocked-user checking, and session key construction.

| Channel     | File        | SDK/Protocol                      |
| ----------- | ----------- | --------------------------------- |
| Telegram    | telegram.go | telego (with voice transcription) |
| Discord     | discord.go  | discordgo                         |
| Slack       | slack.go    | slack-go                          |
| WhatsApp    | whatsapp.go | Bridge URL                        |
| Signal      | signal.go   | Signal CLI                        |
| QQ          | qq.go       | tencent-connect/botgo             |
| DingTalk    | dingtalk.go | dingtalk-stream-sdk               |
| Feishu/Lark | feishu.go   | larksuite/oapi-sdk-go             |
| MaixCam     | maixcam.go  | Custom TCP protocol               |

Supporting: [manager.go](pkg/channels/manager.go) (lifecycle), [multiuser_manager.go](pkg/channels/multiuser_manager.go) (multi-user routing), [command_handler.go](pkg/channels/command_handler.go) (command processing).

#### Providers ([pkg/providers/](pkg/providers/))

```go
type LLMProvider interface {
    Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error)
    GetDefaultModel() string
}
```

Optional `StreamingLLMProvider` adds `ChatStream` for token-by-token streaming.

| Provider       | File               | Supports                                                         |
| -------------- | ------------------ | ---------------------------------------------------------------- |
| HTTPProvider   | http_provider.go   | OpenRouter, OpenAI, Groq, Zhipu, Gemini, Nvidia, Moonshot, VLLM |
| ClaudeProvider | claude_provider.go | Anthropic Claude (official SDK, OAuth token refresh)            |
| CodexProvider  | codex_provider.go  | OpenAI Codex (Responses API)                                     |
| OllamaProvider | ollama_provider.go | Ollama (local, direct HTTP)                                      |
| MockProvider   | mock_provider.go   | Testing (canned responses)                                       |

`CreateProvider(cfg)` auto-selects. Supports `provider/model` syntax (e.g., `openai/gpt-4`).

#### Configuration ([pkg/config/](pkg/config/))

Config: `~/.MakoClaw/config.json` with env var overrides (`MAKOCLAW_<SECTION>_<SUBSECTION>_<FIELD>`).

- `Config` — top-level: `Agents`, `Channels`, `Providers`, `Gateway`, `Web`, `Tools`, `ToolPermissions`, `Storage`
- `AgentsConfig` — defaults + orchestrator config + specialists map
- `ChannelsConfig` — per-channel enable/token/allow_from for all 9 channels
- `ProvidersConfig` — API keys/bases for 10 providers
- `ToolPermissionsConfig` — role-based access with user overrides
- `FlexibleStringSlice` — accepts strings and numbers in JSON allow_from lists (preserve this behavior)

Multi-user: `LoadConfigForUser(uuid)` + `MergeConfigs(global, user)`.

#### Storage ([pkg/storage/](pkg/storage/))

SQLite-backed with two DB types:

- **Per-user** (`storage.Storage`): tasks, knowledge, chat history, metrics, workflows, prompts, channel mappings, task logs, setup sessions
- **Central** (`storage.CentralStorage` in [central.go](pkg/storage/central.go)): user accounts, authentication, per-user provider configs

Key files: sqlite.go, task.go, knowledge.go, workflow.go, metrics.go, user.go, user_providers.go, user_storage.go, multiuser.go, backup.go, chat.go, prompts.go, channel_mapping.go, task_logs.go, setup_session.go

#### Session Management ([pkg/session/](pkg/session/))

File-based history in `<workspace>/sessions/<session_key>.json`. Convention: `channel:chat` (e.g., `cli:default`, `telegram:123456`). Auto-summarizes when history exceeds token/message thresholds. Multi-user isolated.

#### Skills System ([pkg/skills/](pkg/skills/))

Markdown-based extension system. Each skill is a directory with `SKILL.md` (YAML frontmatter: name, description, emoji, required binaries, install instructions).

Loading precedence: workspace skills → global `~/.makoclaw/skills/` → built-in `skills/`

Built-in skills: `development` (code-review, go-best-practices, test-strategy), `github`, `skill-creator`, `summarize`, `tmux`, `weather`

#### Other Packages

| Package                | Description                                                        |
| ---------------------- | ------------------------------------------------------------------ |
| pkg/bus/               | Message bus pub/sub for `InboundMessage`/`OutboundMessage`         |
| pkg/auth/              | OAuth 2.0 + PKCE authentication, token store, JWT                  |
| pkg/cron/              | Cron scheduler (gronx library)                                     |
| pkg/mcp/               | MCP client for external tool servers (JSON-RPC 2.0)                |
| pkg/web/               | REST API + JWT auth + WebSocket streaming + static serving         |
| pkg/voice/             | Voice transcription (Groq speech-to-text)                          |
| pkg/workflow/          | Visual workflow engine for pipeline execution                      |
| pkg/heartbeat/         | Periodic health check service                                      |
| pkg/doctor/            | System diagnostics (config, provider connectivity, DB checks)      |
| pkg/migrate/           | DB and workspace migrations (legacy → multiuser layout)            |
| pkg/observability/     | Metrics collection and tracking                                    |
| pkg/ratelimit/         | Token-bucket rate limiting (login attempts)                        |
| pkg/logger/            | Structured component-based logging                                 |
| pkg/utils/             | String and media utility helpers                                   |

### Web Server ([pkg/web/](pkg/web/))

| File                    | Description                                     |
| ----------------------- | ----------------------------------------------- |
| server.go               | Main setup, routes, middleware                  |
| auth.go                 | Login/signup, JWT token handling                |
| handlers_features.go    | Chat, tasks, knowledge, memory, skills, cron    |
| handlers_advanced.go    | Metrics, workflows                              |
| handlers_tools.go       | Tool listing/execution                          |
| handlers_users.go       | User management (admin)                         |
| handlers_user_config.go | Per-user configuration                          |
| handlers_backup.go      | Backup/restore                                  |
| handlers_setup.go       | Initial setup wizard                            |
| providers_handler.go    | Provider configuration                          |
| openapi.go              | OpenAPI spec generation                         |

### Frontend ([pkg/web/frontend/](pkg/web/frontend/))

Vue 3 + Vite + Tailwind CSS PWA.

**Views (20):** Chat, Dashboard, Tasks (Kanban), Knowledge, Skills, Workflows, Agents, Settings, Metrics, Memory, History, Cron, MCP, Files, Reports, Login, Signup, Setup, Onboarding, Landing

**Stores (Pinia):** auth, chat, config, tasks, agents, onboarding, ui

---

## Key Repository Conventions

### Persistent Paths

| Path                                        | Purpose                     |
| ------------------------------------------- | --------------------------- |
| `~/.MakoClaw/config.json`                   | Config                      |
| `~/.MakoClaw/workspace/`                    | Default workspace           |
| `~/.MakoClaw/workspace/sessions/<key>.json` | Session history             |
| `~/.MakoClaw/workspace/database.db`         | SQLite DB                   |
| `~/.MakoClaw/users/<uuid>/workspace/`       | Per-user workspaces         |
| `~/.makoclaw/skills/`                       | Global skills               |
| `<workspace>/skills/`                       | Workspace skills (override) |

### Logging

```go
logger.InfoCF("component", "message", map[string]interface{}{"key": "value"})
logger.InfoC("component", "message") // no fields
```

Components: `channel`, `channels`, `agent`, `llm`, `tool`, `web`, `workflow`, `mcp`, `cron`, `auth`, `storage`, `migrate`, etc.

### Tool Implementation Pattern

1. Create struct implementing `Tool` interface in [pkg/tools/](pkg/tools/)
2. Optionally implement `ContextualTool`, `WorkspaceTool`, or `UserAwareTool`
3. Register in `NewAgentLoop` in [pkg/agent/loop.go](pkg/agent/loop.go)
4. Schema auto-generated via `ToolToSchema()` for LLM function calling

### Workspace Bootstrap Files

Loaded into the system prompt via `ContextBuilder` ([pkg/agent/context.go](pkg/agent/context.go)):

- `AGENTS.md` — Instructions for AI agents
- `IDENTITY.md` — System identity/purpose
- `SOUL.md` — Personality/behavior guidelines
- `USER.md` — User-specific context

---

## Common Tasks

### Adding a New Tool
1. Create struct in [pkg/tools/](pkg/tools/) implementing `Tool` interface
2. Register in `NewAgentLoop` ([pkg/agent/loop.go](pkg/agent/loop.go))
3. Implement `ContextualTool` if needs channel context
4. Implement `WorkspaceTool` if needs workspace path
5. Implement `UserAwareTool` if needs user filtering

### Adding a New Channel
1. Create file in [pkg/channels/](pkg/channels/), embed `BaseChannel`, implement `Channel` interface
2. Add config struct to [pkg/config/config.go](pkg/config/config.go) under `ChannelsConfig`
3. Register in `initChannels()` in [pkg/channels/manager.go](pkg/channels/manager.go)

### Adding a New LLM Provider
1. Implement `LLMProvider` interface in [pkg/providers/](pkg/providers/)
2. Optionally implement `StreamingLLMProvider` for streaming
3. Add config fields to `ProvidersConfig` in [pkg/config/config.go](pkg/config/config.go)
4. Add selection case in `CreateProvider` ([pkg/providers/http_provider.go](pkg/providers/http_provider.go))

### Modifying Context Construction

`ContextBuilder.BuildContext()` in [pkg/agent/context.go](pkg/agent/context.go):
1. System identity (`getSystemIdentity()`)
2. Runtime/workspace info (OS, arch, Go version, time)
3. Bootstrap docs (AGENTS.md, SOUL.md, USER.md, IDENTITY.md)
4. Skills summary/content
5. Memory context
6. Tool summaries from registry
7. Session summary/history
8. Current message

---

## Testing

| Package        | Test files                                                             | Coverage area                       |
| -------------- | ---------------------------------------------------------------------- | ----------------------------------- |
| pkg/agent/     | permissions_test.go                                                    | Role-based permission filtering     |
| pkg/auth/      | oauth_test.go, pkce_test.go, store_test.go                             | OAuth, PKCE, token store            |
| pkg/channels/  | base_test.go, slack_test.go                                            | Allow-list filtering, Slack adapter |
| pkg/config/    | config_test.go                                                         | Config parsing, env var overrides   |
| pkg/cron/      | service_test.go                                                        | Cron scheduling                     |
| pkg/doctor/    | doctor_test.go                                                         | Diagnostics                         |
| pkg/logger/    | logger_test.go                                                         | Structured logging                  |
| pkg/migrate/   | migrate_test.go                                                        | Migration utilities                 |
| pkg/providers/ | claude_provider_test.go, codex_provider_test.go, http_provider_test.go | LLM providers                       |
| pkg/ratelimit/ | ratelimit_test.go                                                      | Rate limiting                       |
| pkg/storage/   | storage_test.go                                                        | SQLite operations                   |
| pkg/tools/     | email_test.go, shell_test.go, tasks_test.go                            | Email, shell exec, task management  |
| pkg/web/       | auth_test.go, handlers_backup_test.go, server_test.go                  | Auth, backup, server                |

Frontend E2E: `cd pkg/web/frontend && npm test` (Playwright)

---

## CI/CD

GitHub Actions ([.github/workflows/build.yml](.github/workflows/build.yml)):
- Triggers on push to `main` and all PRs
- Sets up Go from go.mod version
- Runs `make build-all` (frontend + all platform binaries)

---

## Project Naming

PicoClaw → KakoClaw → **MakoClaw** (current)

- Go module: `github.com/sipeed/makoclaw`
- Binary: `makoclaw`
- Config dir: `~/.MakoClaw/`
- Repo dir may still be named `kakoclaw`
- Some internal references may use older naming
