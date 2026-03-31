---
name: dev-studio
description: Embedded AI coding assistant that connects the MakoClaw Web UI to Claude Code or OpenCode via a Node.js NDJSON bridge. Supports real-time streaming, HTTP fallback, semantic memory search, and persistent project state.
when_to_use: When using the Dev Studio tab in the Web UI to run Claude Code or OpenCode, send prompts to an AI coding agent, search development memory, or manage the bridge process
emoji: 🛠️
---

# Dev Studio

Dev Studio connects the MakoClaw Web UI to a local AI coding agent (Claude Code or OpenCode) via a Node.js bridge process. It streams agent output in real time and persists project state across page reloads.

## Architecture

```
DevStudioView.vue
    ↓  WebSocket /ws/dev/terminal   (primary)
    ↓  POST /api/v1/dev/query       (HTTP fallback, after 3 WS failures)
handlers_dev_ws.go / handlers_dev.go
    ↓
pkg/bridge/bridge.go  →  exec node bundle.js
    ↓  NDJSON stdin/stdout
pkg/bridge/ts/index.ts  →  @anthropic-ai/claude-agent-sdk
```

The bridge bundles are embedded in the Go binary via `//go:embed` and extracted to disk on first use (`EnsureBridge`).

## REST API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/dev/projects` | List projects in the configured workspace |
| POST | `/api/v1/dev/bridge/start` | Start the bridge for a project directory |
| POST | `/api/v1/dev/bridge/stop` | Stop the running bridge |
| GET | `/api/v1/dev/bridge/status` | Get bridge status + active project dir |
| POST | `/api/v1/dev/query` | HTTP fallback: send a prompt, get NDJSON stream |
| POST | `/api/v1/dev/memory/search` | Semantic search over development memory |

### Start bridge request
```json
{ "project_dir": "/path/to/project" }
```

### Status response
```json
{
  "status": "running|stopped|dead",
  "project_dir": "/path/to/project"
}
```
`project_dir` is restored from `state.json` even after a page reload, so the UI can reconnect to the last active project automatically.

### HTTP fallback (NDJSON streaming)
POST `/api/v1/dev/query` with `{ "message": "..." }` returns a chunked stream of NDJSON lines.
Each line is a JSON event: `{ "type": "assistant"|"result"|"error", "text"?: "...", "content"?: "...", "error"?: "..." }`.

## WebSocket Protocol

Connect to `ws[s]://{host}/ws/dev/terminal?token={jwt}`.

**Send:**
```json
{ "type": "prompt", "message": "your prompt here" }
```

**Receive** (NDJSON events from the bridge):
```json
{ "type": "assistant", "text": "partial response..." }
{ "type": "result", "content": "final answer" }
{ "type": "error", "error": "something went wrong" }
{ "type": "ping" }
```

Ping events are filtered out and not shown in the terminal history.

## HTTP Fallback

After **3 consecutive WebSocket failures**, the frontend automatically switches to HTTP fallback mode (`usingHttpFallback = true`). In this mode:
- Prompts are sent via POST `/api/v1/dev/query`
- Response is read as a `ReadableStream` and parsed as NDJSON line-by-line
- The bridge continues running — only the transport changes
- The fallback indicator is shown in the UI

## Bridge State Persistence

The active project directory is saved to `{storePath}/bridge/{uuid}-state.json` whenever the bridge starts or stops. On page reload, `checkStatus` fetches the status endpoint and restores `currentProject` from `project_dir` if it was set during the last session.

State file schema:
```json
{
  "status": "running",
  "project_dir": "/path/to/project",
  "backend": "claude-code",
  "updated_at": "2026-03-31T12:00:00Z"
}
```

## Semantic Memory

Dev Studio can index and search development memories using ONNX embeddings.

Search via POST `/api/v1/dev/memory/search`:
```json
{ "query": "how we handle auth middleware", "limit": 5 }
```

Returns:
```json
{ "results": [{ "id": 1, "content": "...", "score": 0.91 }] }
```

Requires `DevStudio.Memory.Enabled = true` and a configured ONNX model path in config.

## Configuration

In `~/.MakoClaw/config.json` (or per-user override):

```json
{
  "dev_studio": {
    "enabled": true,
    "default_backend": "claude-code",
    "node_path": "node",
    "memory": {
      "enabled": false,
      "model_path": ""
    }
  }
}
```

- `default_backend`: `"claude-code"` (default) or `"opencode"`
- `node_path`: path to `node` binary (default: `"node"` from PATH)
- `memory.enabled`: enable semantic search over dev memories

## Prompt Injection

When memory is enabled, the backend searches for relevant past context and prepends it to the prompt as a `prompt_injection` block before sending to the bridge. The TypeScript bridge reads `options.prompt_injection` and prepends it to the actual prompt:

```
[injected memories block]

---

[user's original prompt]
```

This lets the agent use prior session knowledge without requiring the user to re-explain context.
