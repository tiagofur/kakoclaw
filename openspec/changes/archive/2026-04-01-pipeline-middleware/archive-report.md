# Archive Report: pipeline-middleware

**Archived**: 2026-04-01  
**Change**: pipeline-middleware  
**Archive Path**: `openspec/changes/archive/2026-04-01-pipeline-middleware/`

## Specs Synced

| Domain | Action | Main Spec Path |
|--------|--------|----------------|
| `dev-pipeline` | Created (new — no prior main spec) | `openspec/specs/dev-pipeline/spec.md` |
| `handler-integration` | Created (new — no prior main spec) | `openspec/specs/handler-integration/spec.md` |

Both delta specs were full specs (not partial deltas), so they were copied directly into `openspec/specs/`.

## Archive Contents

- `proposal.md` ✅
- `specs/` ✅ (2 domains: dev-pipeline, handler-integration)
- `design.md` ✅
- `tasks.md` ✅ (21/21 tasks complete)
- `verify-report.md` — not present (change was verified externally; 21/21 tasks marked complete)

## Implementation Summary

Introduced a composable middleware chain (`DevPipeline`) in `pkg/web/dev_pipeline.go` for Dev Studio query processing. The `Server` struct gained a `devPipeline *DevPipeline` field wired at init time with `MemoryInjectionMiddleware` as the first pre-middleware. Both `handleDevQuery` (HTTP) and `handleDevTerminalWS` (WebSocket) were refactored to delegate pre-processing to the pipeline, eliminating all direct `devMem.Inject` calls from handler code.

### Files Changed

| File | Action |
|------|--------|
| `pkg/web/dev_pipeline.go` | Created — `DevRequest`, `DevResponse`, `PreMiddlewareFunc`, `PostMiddlewareFunc`, `DevPipeline`, `MemoryInjectionMiddleware` |
| `pkg/web/dev_pipeline_test.go` | Created — unit tests for chain ordering, short-circuit, empty chain, MemoryInjectionMiddleware |
| `pkg/web/server.go` | Modified — `devPipeline *DevPipeline` field + `NewServer` wiring |
| `pkg/web/handlers_dev.go` | Modified — replaced inline `devMem.Inject` with `devPipeline.RunPre` |
| `pkg/web/handlers_dev_ws.go` | Modified — replaced inline `devMem.Inject` with `devPipeline.RunPre` |

## Operator Notes

- No `config.json` schema changes — no operator action required.
- No data migrations.
- Rollback: delete `dev_pipeline.go` + `dev_pipeline_test.go`, revert handler inline inject blocks, revert `Server` struct field.

## SDD Cycle Status

All 5 phases complete: propose → spec → design → tasks → apply → archive.
