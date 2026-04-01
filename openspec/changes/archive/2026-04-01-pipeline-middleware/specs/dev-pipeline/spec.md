# Dev Pipeline Specification

## Purpose

Composable pre/post middleware chain for Dev Studio query processing. Replaces hardcoded inline logic in handlers with a reusable `DevPipeline` type that runs middleware sequentially and in a predictable order.

## Requirements

### Requirement: Middleware Function Types

The system MUST define two composable function types:

| Type | Signature | Role |
|------|-----------|------|
| `PreMiddlewareFunc` | `func(context.Context, *DevRequest) (*DevRequest, error)` | Mutates or validates the request before bridge execution |
| `PostMiddlewareFunc` | `func(context.Context, *DevResponse) error` | Reacts to completed responses; errors are non-fatal |

#### Scenario: Pre-middleware signature is correct

- GIVEN a function matching `func(context.Context, *DevRequest) (*DevRequest, error)`
- WHEN registered via `DevPipeline.Use(fn)`
- THEN it MUST be accepted without compile error

#### Scenario: Post-middleware signature is correct

- GIVEN a function matching `func(context.Context, *DevResponse) error`
- WHEN registered via `DevPipeline.UsePost(fn)`
- THEN it MUST be accepted without compile error

---

### Requirement: Pre-Middleware Sequential Execution

`DevPipeline.RunPre` MUST execute pre-middleware functions in registration order. Each function receives the `*DevRequest` returned by the previous one.

#### Scenario: Two pre-middlewares run in order

- GIVEN a pipeline with middlewares A then B registered
- WHEN `RunPre(ctx, req)` is called
- THEN A MUST run first, then B receives A's output `*DevRequest`
- AND the final `*DevRequest` returned by B MUST be returned to the caller

#### Scenario: Empty pre-middleware chain

- GIVEN a pipeline with no pre-middlewares registered
- WHEN `RunPre(ctx, req)` is called
- THEN the original `*DevRequest` MUST be returned unchanged
- AND no error MUST be returned

---

### Requirement: Pre-Middleware Short-Circuit on Error

If any pre-middleware returns a non-nil error, `RunPre` MUST stop execution immediately and return that error. Subsequent pre-middlewares MUST NOT run.

#### Scenario: First middleware fails, second skipped

- GIVEN a pipeline with middleware A (returns error) and middleware B registered
- WHEN `RunPre(ctx, req)` is called
- THEN A's error MUST be returned
- AND B MUST NOT be called

#### Scenario: Last middleware fails

- GIVEN a pipeline with middleware A (succeeds) and middleware B (returns error)
- WHEN `RunPre(ctx, req)` is called
- THEN B's error MUST be returned
- AND the modified `*DevRequest` from A MUST NOT be returned

---

### Requirement: Post-Middleware Run-All Execution

`DevPipeline.RunPost` MUST call all registered post-middlewares. A non-nil error from one MUST NOT stop the remaining ones from running.

#### Scenario: All post-middlewares run even if one errors

- GIVEN a pipeline with post-middlewares X (returns error) and Y (succeeds)
- WHEN `RunPost(ctx, resp)` is called
- THEN both X and Y MUST be called
- AND X's error MUST be logged (not returned)
- AND `RunPost` MUST return nil

#### Scenario: Empty post-middleware chain

- GIVEN a pipeline with no post-middlewares registered
- WHEN `RunPost(ctx, resp)` is called
- THEN nil MUST be returned immediately

---

### Requirement: MemoryInjectionMiddleware

The system MUST provide `MemoryInjectionMiddleware`, a `PreMiddlewareFunc` that wraps `devmemory.Store.Inject`. It MUST set `DevRequest.PromptInjection` when a non-empty injection is returned.

#### Scenario: Memory available and returns context

- GIVEN a `devmemory.Store` that returns non-empty injected text for the prompt
- WHEN `MemoryInjectionMiddleware` runs
- THEN `DevRequest.PromptInjection` MUST be set to the returned text
- AND the modified `*DevRequest` MUST be returned without error

#### Scenario: Memory inject returns empty string

- GIVEN a `devmemory.Store` that returns an empty string
- WHEN `MemoryInjectionMiddleware` runs
- THEN `DevRequest.PromptInjection` MUST remain empty
- AND no error MUST be returned

#### Scenario: Memory inject returns an error

- GIVEN a `devmemory.Store` that returns an error
- WHEN `MemoryInjectionMiddleware` runs
- THEN the error MUST be swallowed (non-fatal)
- AND `DevRequest.PromptInjection` MUST remain empty
- AND the unmodified `*DevRequest` MUST be returned without error

---

### Requirement: DevRequest and DevResponse Types

`DevRequest` MUST carry at minimum: `UserUUID string`, `Prompt string`, `PromptInjection string`. `DevResponse` MUST carry at minimum: `UserUUID string`, `SessionID string`.

#### Scenario: DevRequest fields accessible in middleware

- GIVEN a `PreMiddlewareFunc`
- WHEN it receives a `*DevRequest`
- THEN it MUST be able to read and set `Prompt`, `PromptInjection`, and `UserUUID`
