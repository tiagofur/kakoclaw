# Design: Pipeline Middleware

## Technical Approach

Introduce `DevPipeline` in `pkg/web/dev_pipeline.go` as a thin sequential chain over function types. `Server` gains a single `devPipeline *DevPipeline` field, wired once in `NewServer`. Both `handleDevQuery` and `handleDevWS` replace the inline `devMem.Inject` block with a single `devPipeline.RunPre(ctx, &req)` call before constructing `bridge.Request`. The existing `getDevMemory` / `devMem.Inject` logic moves verbatim into `MemoryInjectionMiddleware`, which is registered as the sole pre-middleware at init time.

## Architecture Decisions

| Decision | Choice | Alternatives Rejected | Rationale |
|----------|--------|-----------------------|-----------|
| Middleware repr | `func(context.Context, *DevRequest) (*DevRequest, error)` (PreMiddlewareFunc) | Interface with single method | Functions are simpler to register inline, test, and mock without extra types |
| RunPre error strategy | First error short-circuits; handler returns HTTP 500 | Log-and-continue | Memory injection failure should block the prompt — garbage in, garbage out |
| RunPost error strategy | All post-middleware run; errors logged via `logger.WarnCF`, not returned | Short-circuit on first error | Post-processing (metrics, formatting) is best-effort; one failure must not hide response |
| Pipeline scope | One `*DevPipeline` per `*Server` (not per-user, not per-request) | Per-user pipeline map | Middleware is stateless; per-user isolation is already in `DevRequest.UserUUID`. A shared instance avoids map+mutex overhead |
| DevRequest ownership | Mutable pointer passed through chain | Immutable copy per step | Allows middleware to enrich without allocation per step; chain is sequential so no concurrent mutation |
| `devBridgesMu` reuse | `getDevMemory` already guards `devMemories` with `devBridgesMu` | New lock | Reuses existing pattern — no new synchronisation primitive needed |

## Data Flow

```
HTTP/WS handler
    │  parse body → DevRequest{UserUUID, Prompt, Options}
    ▼
devPipeline.RunPre(ctx, &req)          ← sequential, short-circuits on error
    │  [0] MemoryInjectionMiddleware
    │       getDevMemory(userUUID)
    │       devMem.Inject(ctx, prompt, 5)
    │       req.Options.PromptInjection = injected
    ▼
bridge.Request{Command, Prompt, Options}
    │
bridge.Execute(ctx, bridgeReq)         ← unchanged
    │
    ▼  <-chan bridge.Event
devPipeline.RunPost(ctx, &resp)        ← sequential, errors logged only
    │  (no post-middleware at launch; hook ready for future use)
    ▼
stream events to HTTP flusher / WS safeWrite
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `pkg/web/dev_pipeline.go` | Create | `DevRequest`, `DevResponse`, `PreMiddlewareFunc`, `PostMiddlewareFunc`, `DevPipeline` struct with `Use`, `RunPre`, `RunPost`; `MemoryInjectionMiddleware` constructor |
| `pkg/web/dev_pipeline_test.go` | Create | Unit tests: ordering, short-circuit, empty chain, `MemoryInjectionMiddleware` inject path |
| `pkg/web/server.go` | Modify | Add `devPipeline *DevPipeline` to `Server`; wire `MemoryInjectionMiddleware` in `NewServer` |
| `pkg/web/handlers_dev.go` | Modify | Replace inline `devMem.Inject` block (lines 499-503) with `devPipeline.RunPre`; build `bridge.Request` from enriched `DevRequest` |
| `pkg/web/handlers_dev_ws.go` | Modify | Replace inline `devMem.Inject` block (lines 152-156) with `devPipeline.RunPre`; same bridge construction change |

## Interfaces / Contracts

```go
// pkg/web/dev_pipeline.go

// DevRequest is the mutable envelope passed through pre-middleware.
type DevRequest struct {
    UserUUID string
    Prompt   string
    Options  bridge.RequestOptions
}

// DevResponse is the mutable envelope passed through post-middleware.
type DevResponse struct {
    UserUUID string
    Events   []bridge.Event // populated by handler after Execute
}

// PreMiddlewareFunc enriches or validates a request before bridge execution.
// Returning a non-nil error short-circuits the chain.
type PreMiddlewareFunc func(ctx context.Context, req *DevRequest) error

// PostMiddlewareFunc processes a completed response.
// Errors are logged but never short-circuit the chain.
type PostMiddlewareFunc func(ctx context.Context, resp *DevResponse) error

// DevPipeline holds the ordered middleware chains for Dev Studio queries.
type DevPipeline struct {
    pre  []PreMiddlewareFunc
    post []PostMiddlewareFunc
}

func NewDevPipeline() *DevPipeline

// Use appends a pre-middleware to the chain.
func (p *DevPipeline) Use(fn PreMiddlewareFunc)

// UsePost appends a post-middleware to the chain.
func (p *DevPipeline) UsePost(fn PostMiddlewareFunc)

// RunPre executes all pre-middleware in order; first error stops the chain.
func (p *DevPipeline) RunPre(ctx context.Context, req *DevRequest) error

// RunPost executes all post-middleware; errors are logged, never returned.
func (p *DevPipeline) RunPost(ctx context.Context, resp *DevResponse)

// MemoryInjectionMiddleware wraps devMem.Inject as a PreMiddlewareFunc.
// It is a no-op when devMem is nil (memory disabled).
func MemoryInjectionMiddleware(getDevMem func(userUUID string) (*devmemory.Store, error)) PreMiddlewareFunc
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit — chain ordering | Middleware execute in registration order | Register 3 pre-middlewares that append to a slice; assert slice order after `RunPre` |
| Unit — short-circuit | First error stops chain, subsequent funcs not called | Register error middleware at position 1; assert position 2 not called |
| Unit — empty chain | `RunPre`/`RunPost` on zero-len chains returns nil / doesn't panic | Direct call with no registered funcs |
| Unit — post log-and-continue | All post-middleware run even when one errors | Register 2 post-middlewares; first returns error; assert both called |
| Unit — `MemoryInjectionMiddleware` | Injects when devMem returns value; skips when error | Mock `getDevMem` returning store; assert `req.Options.PromptInjection` populated |
| Integration | Existing handler tests (`TestHandleDevQuery`, etc.) pass unchanged | Confirm `go test ./pkg/web/... -race` green with new pipeline wired in `NewServer` |

## Migration / Rollout

No schema changes, no feature flags, no data migration.

**Behavioural equivalence**: `MemoryInjectionMiddleware` replicates the exact conditional at `handlers_dev.go:499-503` — same `errMem == nil && errInj == nil && injected != ""` guard. The only observable difference is that the code path runs through `DevPipeline.RunPre` instead of inline.

**Rollback**: Delete `dev_pipeline.go` and `dev_pipeline_test.go`; revert the five-line change in each handler (restore inline inject blocks); revert `Server` struct field addition in `server.go`. No data to clean up.

## Open Questions

- None — design is fully determined by the proposal and existing codebase patterns.
