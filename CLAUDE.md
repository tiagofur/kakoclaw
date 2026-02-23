# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build, Test, and Lint Commands

```bash
# Build the binary (builds frontend first)
make build

# Build for all platforms (used in CI)
make build-all

# Install to system (includes built-in skills installation)
make install

# Run all tests
go test ./...

# Run a specific test
go test ./pkg/config -run TestParseProviderEnvVars -v

# Format code
make fmt   # runs `go fmt ./...`

# Run static analysis
make vet   # runs `go vet ./...`
make security  # runs vet + vulncheck

# Clean build artifacts
make clean
```

## Project Overview

**MakoClaw** (formerly PicoClaw) is an ultra-efficient Go-based AI agent framework designed for resource-constrained hardware. It features:
- Multi-agent orchestration with specialists
- 10+ communication channels (Telegram, Discord, Slack, WhatsApp, Signal, QQ, DingTalk, Feishu, MaixCam)
- 20+ built-in tools (filesystem, web, shell, tasks, knowledge base, etc.)
- Web UI with Kanban tasks, visual workflows, and knowledge base
- <10MB RAM footprint and <1s startup time

## Architecture

### Entry Point

- **[cmd/makoclaw/main.go](cmd/makoclaw/main.go)**: Manual command dispatcher (no Cobra framework). Commands: `agent`, `gateway`, `web`, `cron`, `skills`, `auth`, `doctor`, `migrate`, `onboard`.

### Message Flow Architecture

The system uses a **channel-centric message bus architecture**:

```
Channel → MessageBus → Agent Manager → Agent Loop → LLM Provider → Tools → Response → MessageBus → Channel
```

**Key Flow:**
1. **Channel adapters** normalize incoming events to `bus.InboundMessage`
2. **Message bus** (`pkg/bus`) queues inbound/outbound messages
3. **Agent Loop** ([pkg/agent/loop.go](pkg/agent/loop.go)) runs the tool-calling LLM loop
4. **Outbound responses** are published back to bus and delivered by channels manager

### Core Components

#### Agent System ([pkg/agent/](pkg/agent/))

- **[loop.go](pkg/agent/loop.go)**: Core LLM interaction loop with tool calling
- **[orchestrator.go](pkg/agent/orchestrator.go)**: Multi-agent delegation - analyzes tasks and routes to appropriate specialists
- **[specialist.go](pkg/agent/specialist.go)**: Domain-specific agents with independent LLM configurations
- **[manager.go](pkg/agent/manager.go)**: Coordinates agent instances with channels
- **[context.go](pkg/agent/context.go)**: Composes full LLM context from system identity, skills, memory, history, and current message

#### Tools ([pkg/tools/](pkg/tools/))

All tools implement the `Tool` interface and are registered via `ToolRegistry`. Tools wired in `NewAgentLoop`:
- **Filesystem**: `read_file`, `write_file`, `list_dir`, `edit_file`
- **Web**: `web_search` (Brave Search API), `web_fetch`
- **Execution**: `exec` (shell with security controls), `spawn` (subagents)
- **Tasks**: `task_manager` (Kanban CRUD), `schedule` (cron jobs)
- **Knowledge**: `query_knowledge` (RAG semantic search)
- **Communication**: `message` (cross-channel), `email`
- **Memory**: `memory` (long-term context)

The `cron` tool is added separately in `setupCronTool`.

#### Channels ([pkg/channels/](pkg/channels/))

Each channel implements the `Channel` interface, normalizing external messages to `InboundMessage`. Supports 10+ platforms with user resolution and access control whitelists.

#### Providers ([pkg/providers/](pkg/providers/))

LLM provider abstraction supporting OpenRouter, Claude, OpenAI/Codex, Ollama, and Groq via HTTP provider pattern.

#### Session Management ([pkg/session/](pkg/session/))

SQLite-backed conversation history with:
- Automatic summarization when history exceeds token/message thresholds
- Session identity convention: `channel:chat` (e.g., `cli:default`, `telegram:123456`)
- Multi-user session isolation

#### Skills System ([pkg/skills/](pkg/skills/))

Markdown-based extension system. Loading precedence:
1. Workspace skills override
2. Global skills (`~/.MakoClaw/skills`)
3. Built-in repo skills (`skills/`)

Skills contain YAML frontmatter with metadata and instructions for the agent.

#### MCP Support ([pkg/mcp/](pkg/mcp/))

Model Context Protocol support for external tool integration via client-server pattern.

### Web Server ([pkg/web/](pkg/web/))

REST API with WebSocket for real-time updates. Frontend is Vue 3 + Vite at [pkg/web/frontend/](pkg/web/frontend/).

## Key Repository Conventions

### Persistent Paths

- **Config**: `~/.MakoClaw/config.json` (from `cfg.WorkspacePath()` parent)
- **Workspace**: `~/.MakoClaw/workspace/` (from `cfg.WorkspacePath()`)
- **Sessions**: `~/.MakoClaw/workspace/sessions/<session_key>.json`
- **Database**: `~/.MakoClaw/workspace/database.db`

### Configuration Notes

- Channel access lists accept both strings and numbers via `FlexibleStringSlice` - preserve this when editing config parsing
- Provider settings support multiple providers (openrouter, anthropic, openai, groq, ollama)
- Each specialist can have independent model configuration

### Logging Convention

Use structured component logging:
```go
logger.InfoCF("component", "message", map[string]interface{}{
    "key": "value",
})
```
Components include: `channel`, `agent`, `llm`, `tool`, `web`, `workflow`, etc.

### Tool Implementation Pattern

Tools must implement:
```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}
```

Register via `toolRegistry.Register(tool)` and expose through schema for LLM function calling.

### Workspace Bootstrap Files

When initializing, MakoClaw creates:
- `AGENTS.md`: Instructions for AI agents
- `IDENTITY.md`: System identity/purpose
- `SOUL.md`: Personality/behavior guidelines
- `USER.md`: User-specific context

These are loaded into the system prompt via context builder.

## Common Tasks

### Adding a New Tool

1. Implement the `Tool` interface in [pkg/tools/](pkg/tools/)
2. Register in `NewAgentLoop` in [pkg/agent/loop.go](pkg/agent/loop.go)
3. Add schema metadata for LLM consumption

### Adding a New Channel

1. Create new file in [pkg/channels/](pkg/channels/)
2. Implement `Channel` interface
3. Register in channel factory
4. Add configuration schema to [pkg/config/config.go](pkg/config/config.go)

### Modifying Context Construction

Context is built in [pkg/agent/context.go](pkg/agent/context.go):
- System identity from `getSystemIdentity()`
- Runtime/workspace info
- Bootstrap docs (AGENTS.md, SOUL.md, USER.md, IDENTITY.md)
- Skills summary/content
- Memory context
- Session summary/history
- Current message

## Frontend Development

The Vue 3 frontend is in [pkg/web/frontend/](pkg/web/frontend/):
```bash
cd pkg/web/frontend
npm install
npm run build    # Outputs to pkg/web/dist/
```

## Testing Notes

- Tests use Go's standard `testing` package
- Some packages have `_test.go` files (e.g., [pkg/config/config_test.go](pkg/config/config_test.go))
- Use `-v` flag for verbose output
- Use `-race` for race detection

## Migration Notes

This project appears to be in transition from "kakoclaw" to "makoclaw" branding:
- Binary name: `makoclaw`
- Config directory: `~/.MakoClaw/`
- Some files still reference old naming
