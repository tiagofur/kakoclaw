# Proposal: Pipeline Middleware

## Intent

`handleDevQuery` and `handleDevWS` hardcode a fixed processing order: memory inject → bridge execute → stream. Adding any new pre/post processing (validation, context enrichment, metrics) requires editing these handlers directly. This change introduces a composable middleware chain so new behaviours can be added without touching handler logic.

## Scope

### In Scope
- `DevPipeline` type in `pkg/web/dev_pipeline.go` with pre/post middleware interfaces and chain execution
- `PreMiddleware` interface: `func(context.Context, *DevRequest) (*DevRequest, error)`
- `PostMiddleware` interface: `func(context.Context, *DevResponse) error`
- Refactor `handleDevQuery` to run prompt through the pipeline before calling `b.Execute`
- Refactor `handleDevWS` (`handlers_dev_ws.go`) with same pipeline
- Move `devMem.Inject` call into a `MemoryInjectionMiddleware` (first pre-middleware)
- Unit tests for pipeline: ordering, short-circuit on error, empty chain

### Out of Scope
- Per-project middleware configuration (config.json schema changes — future)
- Frontend changes (middleware is backend-only)
- Post-processing middleware implementations beyond the interface (metrics, formatting — future)
- Changes to `pkg/bridge/protocol.go` or `RequestOptions`

## Approach

Add `pkg/web/dev_pipeline.go` with:
```
type DevPipeline struct {
    pre  []PreMiddlewareFunc
    post []PostMiddlewareFunc
}
```
`DevPipeline.RunPre` iterates pre-middleware sequentially; first error short-circuits. `RunPost` runs all post-middleware; errors are logged, not fatal. `MemoryInjectionMiddleware` wraps the existing `devMem.Inject` call. `Server` gains a `devPipeline *DevPipeline` field wired up in `NewServer`. Handlers delegate to `devPipeline.RunPre` before constructing `bridge.Request`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/web/dev_pipeline.go` | New | Pipeline type, interfaces, `MemoryInjectionMiddleware` |
| `pkg/web/handlers_dev.go` | Modified | Replace inline `devMem.Inject` with `devPipeline.RunPre` |
| `pkg/web/handlers_dev_ws.go` | Modified | Same pipeline wiring for WebSocket handler |
| `pkg/web/server.go` | Modified | Add `devPipeline` field; wire `MemoryInjectionMiddleware` on init |
| `pkg/web/dev_pipeline_test.go` | New | Unit tests for chain ordering and error short-circuit |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Behaviour regression in memory injection | Low | Existing integration tests + new unit tests covering inject path |
| WS and HTTP handlers diverge | Med | Both share the same `DevPipeline` instance — single source of truth |
| Breaking multi-user isolation | Low | Pipeline is per-`Server`, not global; `DevRequest` carries `userUUID` |

## Rollback Plan

Revert `handlers_dev.go` and `handlers_dev_ws.go` to inline `devMem.Inject` calls (one-line change each). Delete `dev_pipeline.go` and its test. No schema migrations, no data changes.

## Dependencies

None — no new external packages required.

## Success Criteria

- [ ] `go test ./pkg/web/... -race` passes with pipeline tests included
- [ ] `handleDevQuery` contains no direct `devMem.Inject` call
- [ ] `handleDevWS` contains no direct `devMem.Inject` call
- [ ] Adding a new pre-middleware requires only: implement `PreMiddlewareFunc`, call `devPipeline.Use(myMiddleware)`
- [ ] `make fmt && make vet` passes with zero warnings
