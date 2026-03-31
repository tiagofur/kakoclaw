# Bridge Specification

## Purpose

Bridge manages long-lived CLI processes (Claude Code, OpenCode) for autonomous code execution, communicating via NDJSON-multiplexed stdin/stdout.

## Requirements

### Requirement: Bridge Lifecycle

The Bridge MUST start a CLI process, maintain a persistent connection, and support graceful shutdown.

#### Scenario: Start bridge with Claude Code backend

- GIVEN a valid bridge config with backend "claude-code"
- WHEN `Start()` is called
- THEN a Node.js child process MUST be spawned running the embedded bundle
- AND the bridge MUST transition to state "running"
- AND a "ready" event MUST be received within 15 seconds

#### Scenario: Start bridge with OpenCode backend

- GIVEN a valid bridge config with backend "opencode"
- WHEN `Start()` is called
- THEN an OpenCode CLI process MUST be spawned
- AND the bridge MUST transition to state "running"

#### Scenario: Stop running bridge

- GIVEN a bridge in state "running"
- WHEN `Stop()` is called
- THEN the child process MUST receive SIGTERM
- AND the bridge MUST transition to state "stopped"
- AND all pending requests MUST receive a cancellation error

#### Scenario: Bridge process dies unexpectedly

- GIVEN a bridge in state "running"
- WHEN the child process exits unexpectedly
- THEN the bridge MUST invoke the `onDeath` callback
- AND the bridge MUST transition to state "dead"
- AND the bridge SHOULD attempt auto-recovery up to 3 times

### Requirement: Request Execution

The Bridge MUST send requests via stdin and correlate responses by `request_id`.

#### Scenario: Execute a query

- GIVEN a bridge in state "running"
- WHEN `Execute(prompt, options)` is called
- THEN a JSON request MUST be written to stdin with a unique `request_id`
- AND the method MUST return a channel of `Event` structs
- AND events MUST be emitted until a terminal event (result/error/pong)

#### Scenario: Ping bridge

- GIVEN a bridge in state "running"
- WHEN `Ping()` is called
- THEN a `{"command":"ping"}` request MUST be sent
- AND a "pong" event MUST be received within 5 seconds

#### Scenario: Execute with options

- GIVEN a bridge in state "running"
- WHEN `Execute(prompt, opts{Model: "sonnet", Cwd: "/project"})` is called
- THEN the request JSON MUST include `model`, `cwd` in the `options` object

### Requirement: NDJSON Event Parsing

The Bridge MUST parse newline-delimited JSON events from the CLI's stdout.

#### Scenario: Parse valid NDJSON stream

- GIVEN stdout emits `{"event":"tool_use","name":"Read","request_id":"abc"}\n`
- WHEN the reader processes the line
- THEN an Event with Type="tool_use", Name="Read", RequestID="abc" MUST be produced

#### Scenario: Handle malformed line

- GIVEN stdout emits a non-JSON line (e.g. stderr leak)
- WHEN the reader processes the line
- THEN the line MUST be logged as warning
- AND the reader MUST continue processing subsequent lines

### Requirement: Bridge Config

The Config MUST include a `DevStudio` section for bridge settings.

#### Scenario: Config with Dev Studio enabled

- GIVEN a config.json with `"dev_studio": {"enabled": true, "default_backend": "claude-code"}`
- WHEN config is loaded
- THEN `config.DevStudio.Enabled` MUST be true
- AND `config.DevStudio.DefaultBackend` MUST be "claude-code"

#### Scenario: Config with memory disabled

- GIVEN a config.json with `"dev_studio": {"memory": {"enabled": false}}`
- WHEN config is loaded
- THEN `config.DevStudio.Memory.Enabled` MUST be false
