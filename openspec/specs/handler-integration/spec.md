# Handler Integration Specification

## Purpose

Describes how `handleDevQuery` (HTTP) and `handleDevTerminalWS` (WebSocket) MUST delegate request pre-processing to `DevPipeline`, eliminating all direct `devMem.Inject` calls from handler code.

## Requirements

### Requirement: Server Owns DevPipeline Instance

The `Server` struct MUST hold a `devPipeline *DevPipeline` field. `NewServer` (or equivalent init function) MUST construct `devPipeline` with `MemoryInjectionMiddleware` registered as the first pre-middleware.

#### Scenario: Server initialised with pipeline

- GIVEN `NewServer` is called with a valid config
- WHEN the server is ready to handle requests
- THEN `server.devPipeline` MUST be non-nil
- AND `MemoryInjectionMiddleware` MUST be the first registered pre-middleware

---

### Requirement: handleDevQuery Uses Pipeline

`handleDevQuery` MUST call `devPipeline.RunPre` before constructing `bridge.Request`. It MUST NOT contain any direct `devMem.Inject` call.

#### Scenario: Prompt enriched before bridge execution

- GIVEN an authenticated POST to `/api/v1/dev/query` with `{"message": "fix the bug"}`
- WHEN `handleDevQuery` processes the request
- THEN `devPipeline.RunPre(ctx, devReq)` MUST be called before `b.Execute`
- AND the `bridge.Request.Options.PromptInjection` MUST reflect what the pipeline set

#### Scenario: Pre-middleware error aborts the HTTP request

- GIVEN a pre-middleware that returns an error (e.g., validation failure)
- WHEN `handleDevQuery` calls `RunPre` and receives an error
- THEN `b.Execute` MUST NOT be called
- AND the handler MUST respond with HTTP 500

#### Scenario: No direct Inject call in handleDevQuery

- GIVEN the source of `handleDevQuery`
- WHEN audited for direct `devMem.Inject` calls
- THEN zero occurrences MUST be found

---

### Requirement: handleDevTerminalWS Uses Pipeline

`handleDevTerminalWS` MUST call `devPipeline.RunPre` before calling `b.Execute` for each `"prompt"` message. It MUST NOT contain any direct `devMem.Inject` call.

#### Scenario: WS prompt enriched via pipeline

- GIVEN a WebSocket connection with a running bridge
- WHEN the client sends `{"type":"prompt","message":"add tests"}`
- THEN `devPipeline.RunPre(ctx, devReq)` MUST be called before `b.Execute`
- AND the bridge request MUST carry the pipeline's `PromptInjection` output

#### Scenario: WS pre-middleware error sends error frame

- GIVEN a pre-middleware that returns an error
- WHEN `handleDevTerminalWS` calls `RunPre` and receives an error
- THEN `b.Execute` MUST NOT be called
- AND an `{"type":"error","message":"..."}` frame MUST be sent to the client

#### Scenario: No direct Inject call in handleDevTerminalWS

- GIVEN the source of `handleDevTerminalWS`
- WHEN audited for direct `devMem.Inject` calls
- THEN zero occurrences MUST be found

---

### Requirement: Both Handlers Share the Same Pipeline Instance

Both `handleDevQuery` and `handleDevTerminalWS` MUST use the same `Server.devPipeline` instance. Middleware MUST NOT be duplicated per-handler.

#### Scenario: Single pipeline instance shared

- GIVEN a `Server` instance
- WHEN `handleDevQuery` and `handleDevTerminalWS` are both called
- THEN both MUST reference the same `devPipeline` pointer
- AND middleware registration side-effects MUST be visible to both handlers

---

### Requirement: Adding New Middleware Requires No Handler Changes

Registering a new pre-middleware MUST only require calling `devPipeline.Use(fn)` during server init. Handler code MUST NOT change.

#### Scenario: New middleware added without modifying handlers

- GIVEN a new `PreMiddlewareFunc` implementing validation logic
- WHEN `devPipeline.Use(validationFn)` is called during server init
- THEN both `handleDevQuery` and `handleDevTerminalWS` MUST execute `validationFn` on every prompt
- AND no changes to either handler file MUST be required
