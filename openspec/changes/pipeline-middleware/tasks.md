# Tasks: Pipeline Middleware

## Phase 1: Foundation — Types & Pipeline Core (TDD RED)

- [ ] 1.1 Create `pkg/web/dev_pipeline_test.go` — failing tests for `DevRequest`/`DevResponse` structs (fields: `UserUUID`, `Prompt`, `PromptInjection`, `SessionID`)
- [ ] 1.2 Create `pkg/web/dev_pipeline_test.go` — failing tests for `RunPre`: ordering (3 middlewares append to slice), empty chain returns nil, short-circuit on first error skips subsequent funcs
- [ ] 1.3 Add failing tests for `RunPost`: all funcs run even when one errors, empty chain no-panics
- [ ] 1.4 Add failing test for `MemoryInjectionMiddleware`: sets `PromptInjection` when inject returns non-empty; no-op when returns empty; swallows error and returns nil

## Phase 2: Foundation — Types & Pipeline Core (TDD GREEN)

- [ ] 2.1 Create `pkg/web/dev_pipeline.go` — define `DevRequest`, `DevResponse`, `PreMiddlewareFunc`, `PostMiddlewareFunc`
- [ ] 2.2 Implement `DevPipeline` struct with `pre []PreMiddlewareFunc`, `post []PostMiddlewareFunc`; add `NewDevPipeline()`, `Use()`, `UsePost()`
- [ ] 2.3 Implement `RunPre`: iterate `pre`, return first error (short-circuit); pass same `*DevRequest` pointer through
- [ ] 2.4 Implement `RunPost`: iterate all `post`, log errors via `logger.WarnCF`, never return error
- [ ] 2.5 Implement `MemoryInjectionMiddleware(getDevMem func(string) (*devmemory.Store, error)) PreMiddlewareFunc` — mirrors guard at `handlers_dev.go:499-503`
- [ ] 2.6 Run `go test ./pkg/web/... -run TestDevPipeline -race` — all Phase 1 tests must pass (GREEN)

## Phase 3: Server Wiring (TDD RED → GREEN)

- [ ] 3.1 Write failing test (or compile-check): `Server` struct has `devPipeline *DevPipeline` field; `NewServer` returns non-nil `devPipeline` with `MemoryInjectionMiddleware` as first pre-middleware
- [ ] 3.2 Modify `pkg/web/server.go` — add `devPipeline *DevPipeline` to `Server`; in `NewServer` call `NewDevPipeline()` and `Use(MemoryInjectionMiddleware(s.getDevMemory))`
- [ ] 3.3 Verify `go build ./pkg/web/...` succeeds after server change

## Phase 4: Handler Integration (TDD RED → GREEN)

- [ ] 4.1 Write failing test for `handleDevQuery`: pipeline `RunPre` called before `b.Execute`; `RunPre` error → HTTP 500, `b.Execute` not called
- [ ] 4.2 Modify `pkg/web/handlers_dev.go` — remove inline `devMem.Inject` block (lines 499-503); construct `DevRequest{UserUUID, Prompt, Options}`; call `devPipeline.RunPre(ctx, &devReq)`; build `bridge.Request` from enriched `devReq`
- [ ] 4.3 Write failing test for `handleDevTerminalWS`: pipeline `RunPre` called before `b.Execute`; `RunPre` error → WS error frame sent, `b.Execute` not called
- [ ] 4.4 Modify `pkg/web/handlers_dev_ws.go` — remove inline `devMem.Inject` block (lines 152-156); same `DevRequest` + `RunPre` pattern; on error write `{"type":"error","message":"..."}` frame
- [ ] 4.5 Run `go test ./pkg/web/... -race` — all handler tests pass (GREEN)

## Phase 5: Refactor & Verification

- [ ] 5.1 Audit `handlers_dev.go` and `handlers_dev_ws.go` — assert zero `devMem.Inject` occurrences (run `grep -rn "devMem.Inject" pkg/web/`)
- [ ] 5.2 Run `make fmt && make vet` — zero warnings
- [ ] 5.3 Run full suite `go test ./pkg/web/... -race` — confirm green including pre-existing handler tests
