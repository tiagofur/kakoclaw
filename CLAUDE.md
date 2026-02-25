# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build, Test, and Lint Commands

```bash
# Build the binary (builds frontend first, then Go binary)
make build

# Build for all platforms (linux-amd64, linux-arm64, linux-riscv64, windows-amd64)
make build-all

# Build only the Vue frontend (outputs to pkg/web/dist/)
make build-frontend

# Install to system (~/.local/bin) and copy built-in skills to workspace
make install

# Install only built-in skills to workspace
make install-skills

# Run all tests
go test ./...

# Run a specific test
go test ./pkg/config -run TestParseProviderEnvVars -v

# Run tests with race detection
go test -race ./...

# Format code
make fmt   # runs `go fmt ./...`

# Run static analysis
make vet   # runs `go vet ./...`

# Run security checks (vet + govulncheck)
make security

# Update dependencies
make deps  # runs `go get -u ./... && go mod tidy`

# Clean build artifacts
make clean

# Build and run
make run ARGS="agent"
```

## Project Overview

**MakoClaw** (module: `github.com/sipeed/makoclaw`, Go 1.26) is an ultra-efficient Go-based AI agent framework designed for resource-constrained hardware. It features:
- Multi-agent orchestration with configurable specialists
- 9 communication channels (Telegram, Discord, Slack, WhatsApp, Signal, QQ, DingTalk, Feishu, MaixCam)
- 20+ built-in tools (filesystem, web, shell, tasks, knowledge base, email, subagents, etc.)
- Web UI (Vue 3 + Vite + Tailwind) with Kanban tasks, visual workflows, knowledge base, and metrics dashboard
- MCP (Model Context Protocol) support for external tool integration
- Multi-user support with role-based tool permissions and per-user workspaces
- <10MB RAM footprint and <1s startup time

## Architecture

### Entry Point

- **[cmd/makoclaw/main.go](cmd/makoclaw/main.go)**: Manual command dispatcher (no Cobra framework). Commands: `agent` (interactive CLI REPL), `gateway` (channel listeners + agent loop), `web` (full web server with REST API and WebSocket), `cron`, `skills`, `auth`, `doctor`, `migrate`, `migrate-multiuser`, `onboard`, `status`, `version`.

### Message Flow Architecture

The system uses a **channel-centric message bus architecture**:

```
Channel → MessageBus → Agent Manager → Agent Loop → LLM Provider → Tools → Response → MessageBus → Channel
```

**Key Flow:**
1. **Channel adapters** normalize incoming events to `bus.InboundMessage` (defined in [pkg/bus/types.go](pkg/bus/types.go))
2. **Message bus** ([pkg/bus/bus.go](pkg/bus/bus.go)) queues inbound/outbound messages via publish/subscribe
3. **Agent Manager** ([pkg/agent/manager.go](pkg/agent/manager.go)) coordinates agent instances with channels
4. **Agent Loop** ([pkg/agent/loop.go](pkg/agent/loop.go)) runs the tool-calling LLM loop with iterative tool execution
5. **Outbound responses** are published back to bus as `bus.OutboundMessage` and delivered by channels manager

### Core Packages

#### Agent System ([pkg/agent/](pkg/agent/))

- **[loop.go](pkg/agent/loop.go)**: Core LLM interaction loop. `NewAgentLoop` wires all tools. `NewAgentLoopForUser` creates per-user loops with merged configs and permission filtering. Supports streaming via `StreamCallback` and tool events via `ToolCallback`.
- **[orchestrator.go](pkg/agent/orchestrator.go)**: Multi-agent delegation — `OrchestratorAgent` analyzes tasks and routes to appropriate `SpecialistAgent`s via `DelegationRequest`/`DelegationResult`.
- **[specialist.go](pkg/agent/specialist.go)**: Domain-specific agents with independent LLM configurations, prompt overrides, and tool subsets.
- **[manager.go](pkg/agent/manager.go)**: Coordinates agent instances with channels, dispatches inbound messages.
- **[context.go](pkg/agent/context.go)**: `ContextBuilder` composes the full LLM context: system identity, runtime/workspace info, bootstrap docs (AGENTS.md, SOUL.md, USER.md, IDENTITY.md), skills summary, memory context, session summary/history, tool summaries, and current message.
- **[permissions.go](pkg/agent/permissions.go)**: Role-based tool permission filtering with per-user overrides and shell command whitelists.

#### Tools ([pkg/tools/](pkg/tools/))

All tools implement the `Tool` interface (in [base.go](pkg/tools/base.go)):
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}
```

Optional extension interfaces:
- `ContextualTool` — receives channel/chatID context via `SetContext(channel, chatID string)`
- `WorkspaceTool` — receives workspace path via `SetWorkspace(workspace string)`
- `UserAwareTool` — receives user ID via `SetUserID(userID int64)`

Tools are registered via `ToolRegistry` ([registry.go](pkg/tools/registry.go)) and wired in `NewAgentLoop`. Full tool list:

| Tool | File | Description |
|------|------|-------------|
| `read_file` | filesystem.go | Read file contents |
| `write_file` | filesystem.go | Write/create files |
| `list_dir` | filesystem.go | List directory contents |
| `edit_file` | edit.go | Targeted file edits (search & replace) |
| `append_file` | edit.go | Append content to files |
| `exec` | shell.go | Shell command execution with security controls |
| `web_search` | web.go | Web search via Brave Search API |
| `web_fetch` | web.go | Fetch and extract web page content |
| `message` | message.go | Cross-channel messaging |
| `email` | email.go | Send emails via SMTP (conditional on config) |
| `spawn` | spawn.go | Spawn subagent conversations |
| `task_manager` | tasks.go | Kanban task CRUD (requires storage) |
| `query_knowledge` | knowledge.go | RAG semantic search (requires storage) |
| `schedule` | cron.go | Cron job scheduling (added via `setupCronTool`) |

Additionally, MCP tools from configured servers are dynamically registered at startup.

Audit logging for restricted tools is handled by `SQLiteAuditLogger` in [audit.go](pkg/tools/audit.go).

#### Channels ([pkg/channels/](pkg/channels/))

Each channel implements the `Channel` interface (in [base.go](pkg/channels/base.go)):
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

`BaseChannel` provides common allow-list filtering, blocked-user checking, and session key construction. Channel implementations:

| Channel | File | SDK/Protocol |
|---------|------|-------------|
| Telegram | telegram.go | telego (with voice transcription) |
| Discord | discord.go | discordgo |
| Slack | slack.go | slack-go |
| WhatsApp | whatsapp.go | Bridge URL |
| Signal | signal.go | Signal CLI |
| QQ | qq.go | tencent-connect/botgo |
| DingTalk | dingtalk.go | dingtalk-stream-sdk |
| Feishu/Lark | feishu.go | larksuite/oapi-sdk-go |
| MaixCam | maixcam.go | Custom TCP protocol |

Supporting files: [manager.go](pkg/channels/manager.go) (lifecycle management), [multiuser_manager.go](pkg/channels/multiuser_manager.go) (multi-user routing), [command_handler.go](pkg/channels/command_handler.go) (command processing).

#### Providers ([pkg/providers/](pkg/providers/))

LLM provider abstraction defined in [types.go](pkg/providers/types.go):
```go
type LLMProvider interface {
    Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error)
    GetDefaultModel() string
}
```

Optional `StreamingLLMProvider` interface adds `ChatStream` for token-by-token streaming.

| Provider | File | Supports | Notes |
|----------|------|----------|-------|
| HTTPProvider | http_provider.go | OpenRouter, OpenAI, Groq, Zhipu, Gemini, Nvidia, Moonshot, VLLM | Generic OpenAI-compatible HTTP provider with streaming |
| ClaudeProvider | claude_provider.go | Anthropic Claude | Official Anthropic SDK with OAuth token refresh |
| CodexProvider | codex_provider.go | OpenAI Codex | Official OpenAI SDK with Responses API |
| OllamaProvider | ollama_provider.go | Ollama (local) | Direct HTTP to local Ollama instance |
| MockProvider | mock_provider.go | Testing | Canned responses for tests |

`CreateProvider(cfg)` in http_provider.go auto-selects the provider. Supports explicit `provider/model` syntax (e.g., `openai/gpt-4`).

#### Configuration ([pkg/config/](pkg/config/))

Config loaded from `~/.MakoClaw/config.json` with env var overrides (prefix `MAKOCLAW_`). Key structs in [config.go](pkg/config/config.go):

- `Config` — top-level: `Agents`, `Channels`, `Providers`, `Gateway`, `Web`, `Tools`, `ToolPermissions`, `Storage`
- `AgentsConfig` — defaults (provider, model, max_tokens, temperature, max_tool_iterations), orchestrator config, specialists map
- `ChannelsConfig` — per-channel enable/token/allow_from settings for all 9 channels
- `ProvidersConfig` — API keys/bases for: anthropic, openai, openrouter, groq, zhipu, vllm, gemini, nvidia, moonshot, ollama
- `ToolPermissionsConfig` — role-based tool access control with user overrides
- `FlexibleStringSlice` — accepts both strings and numbers in JSON for allow_from lists

Environment variable convention: `MAKOCLAW_<SECTION>_<SUBSECTION>_<FIELD>` (e.g., `MAKOCLAW_CHANNELS_TELEGRAM_TOKEN`).

Multi-user config: `LoadConfigForUser(uuid)` + `MergeConfigs(global, user)` for per-user overrides.

#### Session Management ([pkg/session/](pkg/session/))

File-based conversation history in `<workspace>/sessions/<session_key>.json`:
- Session identity convention: `channel:chat` (e.g., `cli:default`, `telegram:123456`)
- Automatic summarization when history exceeds token/message thresholds
- Multi-user session isolation

#### Storage ([pkg/storage/](pkg/storage/))

SQLite-backed persistent storage with two database types:
- **Per-user storage** (`storage.Storage`): tasks, knowledge base, chat history, metrics, workflows, prompts, channel mappings, task logs, setup sessions
- **Central storage** (`storage.CentralStorage` in [central.go](pkg/storage/central.go)): user accounts, authentication, per-user provider configs

Key files: sqlite.go (core DB), task.go, knowledge.go, workflow.go, metrics.go, user.go, user_providers.go, user_storage.go, multiuser.go, backup.go, chat.go, prompts.go, channel_mapping.go, task_logs.go, setup_session.go

#### Skills System ([pkg/skills/](pkg/skills/))

Markdown-based extension system. Skills are directories containing `SKILL.md` with YAML frontmatter (name, description, metadata including emoji, required binaries, install instructions). Loading precedence:
1. Workspace skills (user overrides)
2. Global skills (`~/.makoclaw/skills/`)
3. Built-in repo skills (`skills/`)

Built-in skills: `development` (code-review, go-best-practices, test-strategy), `github`, `skill-creator`, `summarize`, `tmux`, `weather`

#### Other Packages

| Package | Description |
|---------|-------------|
| [pkg/bus/](pkg/bus/) | Message bus with pub/sub for `InboundMessage`/`OutboundMessage` |
| [pkg/auth/](pkg/auth/) | OAuth 2.0 + PKCE authentication, token store, JWT |
| [pkg/cron/](pkg/cron/) | Cron job scheduler using gronx library |
| [pkg/mcp/](pkg/mcp/) | Model Context Protocol client for external tool servers (JSON-RPC 2.0) |
| [pkg/web/](pkg/web/) | REST API server with JWT auth, WebSocket streaming, static file serving |
| [pkg/voice/](pkg/voice/) | Voice transcription (Groq speech-to-text) |
| [pkg/workflow/](pkg/workflow/) | Visual workflow engine for pipeline execution |
| [pkg/heartbeat/](pkg/heartbeat/) | Periodic heartbeat/health check service |
| [pkg/doctor/](pkg/doctor/) | System diagnostics (config validation, provider connectivity, DB checks) |
| [pkg/migrate/](pkg/migrate/) | Database and workspace migrations (legacy → multiuser layout) |
| [pkg/observability/](pkg/observability/) | Metrics collection and tracking |
| [pkg/ratelimit/](pkg/ratelimit/) | Token-bucket rate limiting (used for login attempts) |
| [pkg/logger/](pkg/logger/) | Structured component-based logging |
| [pkg/utils/](pkg/utils/) | String and media utility helpers |

### Web Server ([pkg/web/](pkg/web/))

REST API + WebSocket server with JWT authentication. Handler files:
- `server.go` — main server setup, routes, middleware
- `auth.go` — login/signup, JWT token handling
- `handlers_features.go` — chat, tasks, knowledge, memory, skills, cron endpoints
- `handlers_advanced.go` — metrics, workflows
- `handlers_tools.go` — tool listing/execution endpoints
- `handlers_users.go` — user management (admin)
- `handlers_user_config.go` — per-user configuration
- `handlers_backup.go` — backup/restore
- `handlers_setup.go` — initial setup wizard
- `providers_handler.go` — provider configuration
- `openapi.go` — OpenAPI spec generation

### Frontend ([pkg/web/frontend/](pkg/web/frontend/))

Vue 3 + Vite + Tailwind CSS PWA application.

**Views (20):** Chat, Dashboard, Tasks (Kanban), Knowledge, Skills, Workflows, Agents, Settings, Metrics, Memory, History, Cron, MCP, Files, Reports, Login, Signup, Setup, Onboarding, Landing

**Stores (Pinia):** auth, chat, config, tasks, agents, onboarding, ui

**Frontend development:**
```bash
cd pkg/web/frontend
npm install
npm run dev      # Dev server with HMR
npm run build    # Production build → pkg/web/dist/
npm run preview  # Preview production build
npm test         # Playwright E2E tests
```

## Key Repository Conventions

### Persistent Paths

- **Config**: `~/.MakoClaw/config.json`
- **Workspace**: `~/.MakoClaw/workspace/`
- **Sessions**: `~/.MakoClaw/workspace/sessions/<session_key>.json`
- **Database**: `~/.MakoClaw/workspace/database.db`
- **User workspaces** (multi-user): `~/.MakoClaw/users/<uuid>/workspace/`
- **Global skills**: `~/.makoclaw/skills/`
- **Workspace skills**: `<workspace>/skills/`

### Configuration Notes

- Channel access lists accept both strings and numbers via `FlexibleStringSlice` — preserve this when editing config parsing
- Provider settings support 10 providers (anthropic, openai, openrouter, groq, zhipu, vllm, gemini, nvidia, moonshot, ollama)
- Each specialist can have independent model/provider configuration
- All config fields support env var overrides with `MAKOCLAW_` prefix

### Logging Convention

Use structured component logging:
```go
logger.InfoCF("component", "message", map[string]interface{}{
    "key": "value",
})
```
Shorthand (no fields): `logger.InfoC("component", "message")`

Components include: `channel`, `channels`, `agent`, `llm`, `tool`, `web`, `workflow`, `mcp`, `cron`, `auth`, `storage`, `migrate`, etc.

### Tool Implementation Pattern

1. Create a struct implementing the `Tool` interface in [pkg/tools/](pkg/tools/)
2. Optionally implement `ContextualTool`, `WorkspaceTool`, or `UserAwareTool`
3. Register in `NewAgentLoop` in [pkg/agent/loop.go](pkg/agent/loop.go)
4. Schema is auto-generated via `ToolToSchema()` for LLM function calling

### Workspace Bootstrap Files

When initializing, MakoClaw creates markdown files loaded into the system prompt:
- `AGENTS.md`: Instructions for AI agents
- `IDENTITY.md`: System identity/purpose
- `SOUL.md`: Personality/behavior guidelines
- `USER.md`: User-specific context

These are loaded via `ContextBuilder` in [pkg/agent/context.go](pkg/agent/context.go).

## Common Tasks

### Adding a New Tool

1. Create tool struct in [pkg/tools/](pkg/tools/) implementing `Tool` interface
2. Register in `NewAgentLoop` in [pkg/agent/loop.go](pkg/agent/loop.go)
3. If tool needs channel context, implement `ContextualTool`
4. If tool needs workspace path, implement `WorkspaceTool`
5. If tool needs user filtering, implement `UserAwareTool`

### Adding a New Channel

1. Create new file in [pkg/channels/](pkg/channels/)
2. Embed `BaseChannel` and implement `Channel` interface (especially `Start`, `Stop`, `Send`)
3. Add config struct to [pkg/config/config.go](pkg/config/config.go) under `ChannelsConfig`
4. Register in `initChannels()` in [pkg/channels/manager.go](pkg/channels/manager.go)

### Adding a New LLM Provider

1. Implement `LLMProvider` interface in [pkg/providers/](pkg/providers/)
2. Optionally implement `StreamingLLMProvider` for streaming support
3. Add config fields to `ProvidersConfig` in [pkg/config/config.go](pkg/config/config.go)
4. Add provider selection case in `CreateProvider` in [pkg/providers/http_provider.go](pkg/providers/http_provider.go)

### Modifying Context Construction

Context is built in [pkg/agent/context.go](pkg/agent/context.go) via `ContextBuilder.BuildContext()`:
1. System identity from `getSystemIdentity()`
2. Runtime/workspace info (OS, arch, Go version, time)
3. Bootstrap docs (AGENTS.md, SOUL.md, USER.md, IDENTITY.md)
4. Skills summary/content
5. Memory context
6. Tool summaries from registry
7. Session summary/history
8. Current message

## Testing

Tests exist across these packages:

| Package | Test files | Coverage area |
|---------|-----------|---------------|
| pkg/agent/ | permissions_test.go | Role-based permission filtering |
| pkg/auth/ | oauth_test.go, pkce_test.go, store_test.go | OAuth, PKCE, token store |
| pkg/channels/ | base_test.go, slack_test.go | Allow-list filtering, Slack adapter |
| pkg/config/ | config_test.go | Config parsing, env var overrides |
| pkg/cron/ | service_test.go | Cron scheduling |
| pkg/doctor/ | doctor_test.go | Diagnostics |
| pkg/logger/ | logger_test.go | Structured logging |
| pkg/migrate/ | migrate_test.go | Migration utilities |
| pkg/providers/ | claude_provider_test.go, codex_provider_test.go, http_provider_test.go | LLM providers |
| pkg/ratelimit/ | ratelimit_test.go | Rate limiting |
| pkg/storage/ | storage_test.go | SQLite operations |
| pkg/tools/ | email_test.go, shell_test.go, tasks_test.go | Email, shell exec, task management |
| pkg/web/ | auth_test.go, handlers_backup_test.go, server_test.go | Auth, backup, server |

Frontend E2E tests: `cd pkg/web/frontend && npm test` (Playwright)

## CI/CD

GitHub Actions workflow in [.github/workflows/build.yml](.github/workflows/build.yml):
- Triggers on push to `main` and all pull requests
- Sets up Go from go.mod version
- Runs `make build-all` (builds frontend + all platform binaries)

## Project Naming

The project has gone through naming transitions: PicoClaw → KakoClaw → MakoClaw
- Go module: `github.com/sipeed/makoclaw`
- Binary name: `makoclaw`
- Config directory: `~/.MakoClaw/`
- Repository directory may still be named `kakoclaw`
- Some internal references may use older naming conventions
