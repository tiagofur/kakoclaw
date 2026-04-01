# Dev Studio — User Guide

Dev Studio embeds a coding agent (Claude Code or OpenCode) inside MakoClaw's web UI. You select a project, and get an interactive terminal where you give instructions in natural language — the agent reads files, writes code, executes commands, and reports back in real time.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Enable Dev Studio](#enable-dev-studio)
3. [Authentication](#authentication)
4. [Adding Projects](#adding-projects)
5. [Using the Terminal](#using-the-terminal)
6. [Configuration Reference](#configuration-reference)
7. [Backends: Claude Code vs OpenCode](#backends-claude-code-vs-opencode)
8. [Semantic Memory](#semantic-memory)
9. [Troubleshooting](#troubleshooting)
10. [Architecture Overview](#architecture-overview)

---

## Prerequisites

- MakoClaw running (web mode)
- **Node.js + npm** installed in the runtime environment
- **Claude Code CLI** installed globally (`npm install -g @anthropic-ai/claude-code`) — the Agent SDK spawns it as a subprocess
- An **Anthropic API key** or Claude OAuth login

### Docker

The Dockerfile copies Node.js from the build stage and installs Claude Code CLI globally. If you're running an older image, install manually:

```bash
docker exec -u root -it <container> bash
# If Node/npm are missing:
apt-get update && apt-get install -y nodejs npm
# Install Claude Code CLI (required — the SDK spawns it as a subprocess):
npm install -g @anthropic-ai/claude-code
```

---

## Enable Dev Studio

### Option A: From the Web UI

1. Go to **Settings > Tools**
2. Toggle **Dev Studio** on
3. Save

### Option B: Edit config directly

In `~/.MakoClaw/users/<your-uuid>/config.json` (or `~/.MakoClaw/config.json` for global):

```json
{
  "dev_studio": {
    "enabled": true
  }
}
```

Restart MakoClaw after editing the file manually.

---

## Authentication

The bridge process needs credentials to talk to the LLM provider. How you provide them depends on the backend.

### Claude Code backend (default)

The bridge uses `@anthropic-ai/claude-agent-sdk`, which authenticates in this order:

#### 1. API Key (recommended for Docker / VPS)

Set the `ANTHROPIC_API_KEY` environment variable:

```bash
# docker run
docker run -e ANTHROPIC_API_KEY=sk-ant-api03-... makoclaw

# docker-compose.yml
services:
  makoclaw:
    environment:
      - ANTHROPIC_API_KEY=sk-ant-api03-...
```

This is the simplest method. No interactive login required.

#### 2. OAuth Login (interactive)

If you prefer OAuth (e.g., to use Claude Pro/Max subscription credits):

```bash
docker exec -it <container> bash
npx @anthropic-ai/claude-code login
```

This opens an OAuth flow:
1. It prints a URL — open it in your local browser
2. Authorize the application
3. Copy the code back into the terminal

Credentials are saved to `~/.claude/.credentials.json` inside the container. They persist as long as the container's filesystem does (use a volume to survive rebuilds).

> **Tip**: OAuth also enables Cloud MCP servers from claude.ai, which are automatically merged into Dev Studio sessions.

### OpenCode backend

OpenCode handles its own authentication. Refer to the [OpenCode documentation](https://github.com/opencode-ai/opencode) for setup instructions. Make sure the `opencode` binary is in the container's PATH.

---

## Adding Projects

Dev Studio looks for projects in `{workspace}/repos/` by default (configurable via `projects_dir`).

### From the Web UI

1. Open **Dev Studio** in the sidebar
2. Click **New Project** in the left panel
3. Enter a name and optionally check "Initialize git"
4. The project appears in the list immediately

### Manually (CLI / Docker)

Any directory inside the projects folder appears automatically:

```bash
# Enter the container
docker exec -it <container> bash

# Navigate to projects directory
cd ~/.MakoClaw/workspace/repos

# Clone a repo
git clone https://github.com/your-org/your-project.git

# Or create an empty project
mkdir my-new-project && cd my-new-project && git init
```

Refresh the project list in the UI (click the refresh icon) and it will appear.

### Using a volume for projects

To persist projects across container rebuilds, mount a volume:

```yaml
# docker-compose.yml
services:
  makoclaw:
    volumes:
      - ./my-projects:/home/makoclaw/.MakoClaw/workspace/repos
```

---

## Using the Terminal

### Starting a session

1. Open **Dev Studio** from the sidebar
2. Click a project in the left panel
3. Wait for the bridge status indicator to turn **green** ("running")
4. Type a prompt in the input box at the bottom

### What you'll see

The terminal shows a real-time stream of events:

| Event | What it means |
|-------|---------------|
| `$ your prompt` | Your input (echoed back) |
| Assistant text | The agent's reasoning and explanations |
| `Executing: bash` | The agent is running a shell command |
| `Reading: src/main.go` | The agent is reading a file |
| Tool result | Output from a tool execution |
| `Result` (with cost/duration) | The query completed successfully |
| `Error` | Something went wrong |

### Example prompts

```
list all files in this project
```

```
explain the architecture of this codebase
```

```
add unit tests for the auth middleware
```

```
find and fix the bug in the login handler — users report getting 500 errors
```

```
refactor the database queries to use transactions
```

### Switching projects

Click a different project in the sidebar. The current bridge stops and restarts with the new project directory. Session history is saved per-project — you can come back later and see what you did.

### Connection resilience

Dev Studio uses WebSocket for real-time streaming. If the WebSocket drops 3 times in a row, it automatically switches to **HTTP streaming** (NDJSON). The experience is identical — just slightly higher latency. No action needed from you.

---

## Configuration Reference

Full configuration with defaults:

```json
{
  "dev_studio": {
    "enabled": false,
    "default_backend": "claude-code",
    "node_path": "node",
    "projects_dir": "repos",
    "max_session_tokens": 200000,
    "memory": {
      "enabled": false,
      "model": "all-MiniLM-L6-v2"
    }
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Master switch. All Dev Studio endpoints return 503 when disabled. |
| `default_backend` | string | `"claude-code"` | Which coding agent to use: `"claude-code"` or `"opencode"`. |
| `node_path` | string | `"node"` | Path to the Node.js binary. Change if `node` isn't in your PATH. |
| `projects_dir` | string | `"repos"` | Subdirectory inside workspace where projects are stored. |
| `max_session_tokens` | int | `200000` | Maximum tokens per conversation session. |
| `memory.enabled` | bool | `false` | Enable semantic memory search (see [Semantic Memory](#semantic-memory)). |
| `memory.model` | string | `"all-MiniLM-L6-v2"` | ONNX embedding model for memory vectors. |

### Environment variables

These environment variables affect the bridge process:

| Variable | Purpose |
|----------|---------|
| `ANTHROPIC_API_KEY` | API key for Claude Code backend |
| `ANTHROPIC_MODEL` | Override the default model (e.g., `claude-sonnet-4-20250514`) |

---

## Backends: Claude Code vs OpenCode

### Claude Code (default)

- Uses `@anthropic-ai/claude-agent-sdk` (official Anthropic SDK)
- Requires **Node.js** in the runtime
- Authenticates via `ANTHROPIC_API_KEY` or OAuth (`~/.claude/.credentials.json`)
- Supports Claude's full tool suite (file read/write, bash, search, etc.)
- Supports MCP servers (both local and Cloud MCP from claude.ai)
- Shows cost and duration per query
- 10-minute timeout per query

### OpenCode

- Runs the `opencode` binary directly (no Node.js required)
- Authenticates via OpenCode's own configuration
- Spawns a new subprocess per query (`opencode -p "prompt" --non-interactive`)
- Simpler integration, but no streaming of individual tool events
- No MCP server support through Dev Studio
- No cost tracking

### Switching backends

In your config:

```json
{
  "dev_studio": {
    "default_backend": "opencode"
  }
}
```

Make sure the `opencode` binary is installed and in your PATH.

---

## Semantic Memory

When enabled, Dev Studio maintains a local vector database of past conversations. Before each prompt, it searches for relevant context and injects it automatically.

### How it works

1. Conversations are saved to `dev_memory.db` (SQLite with vector embeddings)
2. When you send a prompt, relevant past context is retrieved by cosine similarity
3. Retrieved memories are prepended to your prompt (invisible to you, visible to the agent)
4. The agent benefits from context like "last time we worked on auth, we used JWT with RS256"

### Enable it

```json
{
  "dev_studio": {
    "memory": {
      "enabled": true,
      "model": "all-MiniLM-L6-v2"
    }
  }
}
```

The embedding model is downloaded automatically on first use. The default `all-MiniLM-L6-v2` is lightweight (~80MB) and works well for code context.

### Search memory manually

Switch to the **Memory** tab in Dev Studio and type a search query. Results show matching memories with timestamps and content previews.

---

## Troubleshooting

### Bridge won't start (503 error)

| Error message | Cause | Fix |
|---------------|-------|-----|
| `dev studio is not enabled for this user` | Config not saved or `enabled: false` | Enable in Settings > Tools, or set `"enabled": true` in config |
| `failed to setup bridge bundle: create bridge dir: not a directory` | `Storage.Path` points to a `.db` file instead of a directory | Update to latest version (this bug was fixed) |
| `exec: "node": executable file not found` | Node.js not installed in runtime | Install Node.js in the container (see [Prerequisites](#prerequisites)) |
| `server configuration not available` | Server started without config | Check that config.json exists and is valid JSON |

### Bridge starts but queries fail

| Error message | Cause | Fix |
|---------------|-------|-----|
| `authentication_error` | No API key or invalid key | Set `ANTHROPIC_API_KEY` env var or run OAuth login |
| `query timeout: no result after 10 minutes` | Prompt too complex or API unresponsive | Try a simpler prompt. Check API status at status.anthropic.com |
| `rate_limit_error` | Too many requests | Wait and retry. Consider upgrading your API plan |

### WebSocket connection issues

If you see frequent disconnections:
- Check that your reverse proxy (nginx, Caddy, etc.) supports WebSocket upgrades
- Ensure timeout settings are generous (at least 60s)
- After 3 failures, Dev Studio automatically falls back to HTTP — no manual action needed

### Checking logs

Bridge logs go to stderr of the MakoClaw process:

```bash
# Docker
docker logs <container> 2>&1 | grep "\[bridge\]"

# Systemd
journalctl -u makoclaw | grep "\[bridge\]"
```

---

## Architecture Overview

```
Browser (DevStudioView.vue)
    |
    |-- WebSocket: ws://host/ws/dev/terminal (primary)
    |-- HTTP POST: /api/v1/dev/query (fallback after 3 WS failures)
    |
Web Server (handlers_dev.go, handlers_dev_ws.go)
    |
    |-- Manages Bridge instances per user (one bridge per user)
    |-- Saves session history to storage
    |-- Injects semantic memory if enabled
    |
Bridge Process (pkg/bridge/bridge.go)
    |
    |-- Long-lived Node.js process
    |-- Communicates via NDJSON over stdin/stdout
    |-- One process per user, reused across queries
    |
TypeScript Bridge (pkg/bridge/ts/index.ts)
    |
    |-- Receives JSON requests on stdin
    |-- Calls @anthropic-ai/claude-agent-sdk query()
    |-- Streams events back on stdout
    |
Claude API (api.anthropic.com)
```

### Key paths

| Path | Purpose |
|------|---------|
| `~/.MakoClaw/workspace/repos/` | Default projects directory |
| `~/.MakoClaw/workspace/bridge/` | Extracted bridge bundles + state files |
| `~/.MakoClaw/users/{uuid}/config.json` | Per-user configuration |
| `~/.claude/.credentials.json` | OAuth credentials (if using login) |

### REST Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/dev/projects` | List projects |
| POST | `/api/v1/dev/projects` | Create new project |
| POST | `/api/v1/dev/bridge/start` | Start bridge for a project |
| POST | `/api/v1/dev/bridge/stop` | Stop the bridge |
| GET | `/api/v1/dev/bridge/status` | Get bridge status |
| POST | `/api/v1/dev/query` | HTTP fallback for terminal |
| POST | `/api/v1/dev/memory/search` | Search semantic memory |
| WS | `/ws/dev/terminal` | WebSocket terminal |
