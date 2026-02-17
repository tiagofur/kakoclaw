# PicoClaw Copilot Instructions

## Build, test, and lint commands
- Build local binary: `make build`
- CI-style cross-platform build: `make build-all` (used in `.github/workflows/build.yml`)
- Run all tests: `go test ./...`
- Run a single test: `go test ./pkg/config -run TestParseProviderEnvVars -v`
- Format code: `make fmt` (runs `go fmt ./...`)
- Lint: no repository `make lint` target exists; docs reference `golangci-lint run` when installed locally

## High-level architecture
- `cmd/picoclaw/main.go` is a manual command dispatcher (no Cobra): `agent`, `gateway`, `web`, `cron`, `skills`, `auth`, `doctor`, `migrate`, `onboard`.
- Runtime message flow is channel-centric:
  1) channel adapters normalize incoming events to `bus.InboundMessage`,
  2) `pkg/bus` queues inbound/outbound messages,
  3) `pkg/agent/loop.go` runs the tool-calling LLM loop,
  4) outbound responses are published back to bus and delivered by `pkg/channels/manager.go`.
- `AgentLoop` wires tools in `NewAgentLoop` (filesystem, shell, web search/fetch, messaging, spawn, edit/append, tasks; cron tool is added separately in `setupCronTool`).
- Context construction (`pkg/agent/context.go`) composes:
  - dynamic system identity + runtime/workspace info,
  - workspace bootstrap docs (`AGENTS.md`, `SOUL.md`, `USER.md`, `IDENTITY.md`),
  - skills summary/content and memory context,
  - session summary/history + current message.
- Gateway mode (`gatewayCmd`) orchestrates agent loop + channel manager + cron service + heartbeat service, and optionally the web server.

## Key repository conventions
- Default persistent paths are user-home based (e.g., config `~/.picoclaw/config.json`, workspace `~/.picoclaw/workspace`), and workspace path comes from `cfg.WorkspacePath()`.
- Session identity convention is `channel:chat` (for example `cli:default`), and agent history is summarized automatically once thresholds are hit (message-count/token based).
- Skills loading precedence is strict: workspace skills override global (`~/.picoclaw/skills`), which override built-in repo skills (`skills/`).
- Channel access lists intentionally accept either strings or numbers in JSON via `FlexibleStringSlice`; preserve this behavior when editing config parsing.
- Logging convention is structured component logging (`logger.*C` / `logger.*CF`) with map fields; keep new logs component-scoped and structured.
- Tool implementations should be registered via `ToolRegistry` and exposed through schema metadata (`Name/Description/Parameters/Execute`) so providers can call them as functions.
