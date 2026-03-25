# Proposal: User Data Isolation Hardening

**Change:** `audit-isolation-user-data`
**Date:** 2026-03-25
**Status:** Complete

---

## Intent

Eliminate all confirmed and potential data-leak paths between users in MakoClaw. The audit identified 8 isolation gaps; 4 affect both deployment modes (MEDIUM+) and 2 represent real data-leak vectors in legacy shared-DB mode (HIGH). This proposal closes every gap in priority order, with emphasis on the HIGH-severity issues first.

---

## Scope

### In Scope

- **ISSUE-4**: Add `user_id` column + scoped CRUD to `workflows`, `workflow_runs`, `workflow_approvals` in the legacy shared DB
- **ISSUE-3**: Deprecate / guard `processNextTodoTaskLegacy` — document single-user-only contract; add a hard panic/log if called with multiple users present
- **ISSUE-1**: Replace `observability.Global()` singleton with per-user `Metrics` instances (or add `user_id` attribution + filter to the metrics endpoint)
- **ISSUE-7**: Remove silent fallback to shared `CronService`; fail hard with a structured error if per-user cron cannot be resolved for an authenticated request
- **ISSUE-8**: Audit `MultiUserChannelManager` to confirm per-user `SessionManager` storage paths; fix if absent
- **ISSUE-2**: Add `JOIN tasks` ownership guard to `GetTaskLogs` in legacy shared-DB mode
- **ISSUE-6**: Consolidate `config.EnsureUserWorkspace` and `UserStorageManager.EnsureUserDirectory` into one canonical function in `pkg/config`
- **ISSUE-5**: Document `knowledge_documents.user_id DEFAULT 1` as intentional no-op in per-user DB (no code change)
- **ISSUE-9**: Multi-Agent User Context Propagation — ensure all agent components (specialists, swarms, orchestrators) properly propagate user identity for workspace isolation and audit trails

### Out of Scope

- Migrating existing legacy shared-DB deployments to per-user DB layout
- Marketplace / skill data isolation (already intentionally shared)
- Auth / central store schema changes
- New user-facing features or UI changes

---

## Approach

Fix issues in strict priority order:

### Phase 1 — HIGH (data-leak risk)

**ISSUE-4 — Workflows schema + CRUD**
- Add `user_id INTEGER NOT NULL` to `workflows`, `workflow_runs`, `workflow_approvals` via a migration in `pkg/storage/workflow.go`
- Add `user_id` parameter to all CRUD methods; scope every `SELECT` with `WHERE user_id = ?`
- Mirror the pattern already used by `tasks` (`isUserDB` flag to skip filter in per-user mode)
- Verify `handleWorkflows` passes `userID` from `getUserStorage(r)` context

**ISSUE-3 — Legacy task worker**
- Add a startup check: if `s.userMgr == nil` AND more than one user row exists in `s.store`, log a FATAL warning
- Add `// SINGLE-USER ONLY` comment block to `processNextTodoTaskLegacy`
- Gate the function: return early with error log if called with `s.userMgr != nil`

### Phase 2 — MEDIUM (information leakage / silent misconfiguration)

**ISSUE-1 — Metrics isolation**
- Preferred: inject per-user `Metrics` via `getUserStorage` context (same pattern as DB)
- Fallback: add `user_id` column to `metrics_events`/`metrics_counters`; filter in `handleMetrics` by requesting user
- Remove `observability.Global().SetStorage(s.store)` from `server.go:253`

**ISSUE-7 — CronService fallback**
- In `getCronServiceForRequest()`, replace the silent fallback with `return nil, ErrCronNotInitialized`
- All callers already check error returns — no handler changes needed beyond removing fallback logic

**ISSUE-8 — SessionManager path audit**
- Grep `MultiUserChannelManager` for `SessionManager` instantiation
- Assert `SetStorage` is called with `~/.MakoClaw/users/<uuid>/workspace/sessions/`
- Add integration test: create two users, write a session for each, confirm cross-user read returns 404

### Phase 3 — LOW / Trivial (code quality)

**ISSUE-9 — Multi-Agent User Context Propagation**
- Add `SetUserForAgent(userUUID, userID)` call for all specialists in `manager.go:InitializeOrchestrator`
- Ensure swarm executions propagate user context via `TeamContext`
- Add `userID` field to `SpecialistSpawnTask` for audit trail
- Verify `SwarmRunner` respects per-user workspace isolation
- Add user identity to `TaskDecompositionTool` team context creation

**ISSUE-2 — GetTaskLogs ownership guard**
- In legacy mode (`!isUserDB`): add `JOIN tasks ON tasks.id = task_logs.task_id AND tasks.user_id = ?`
- No schema changes needed

**ISSUE-6 — Workspace init consolidation**
- Move canonical implementation to `pkg/config/workspace_init.go:EnsureUserWorkspace`
- Replace body of `UserStorageManager.EnsureUserDirectory` with a call to the canonical function
- Add a single unit test covering all expected subdirectories

**ISSUE-5 — Documentation only**
- Add inline comment in `pkg/storage/knowledge.go` explaining why `DEFAULT 1` is harmless in per-user mode

**ISSUE-9 — Multi-Agent User Context Propagation**
- In `pkg/agent/manager.go:InitializeOrchestrator`, call `SetUserForAgent(userUUID, userID)` for all specialists created via `NewAgentLoop(globalCfg)`
- In `pkg/agent/orchestrator.go`, add `userUUID` and `userID` to `TeamContext` when delegating to specialists or spawning tasks
- In `pkg/agent/swarm.go`, ensure swarm executor propagates user context from parent agent loop to child specialists
- In `pkg/agent/spawner.go`, add `UserID int64` field to `SpecialistSpawnTask` struct for audit tracking
- Verify that `SwarmRunner` uses per-user workspace paths when executing swarm tasks

---

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/storage/workflow.go` | Modified | Add `user_id` column + migration; scope all CRUD |
| `pkg/web/server.go` | Modified | Guard `processNextTodoTaskLegacy`; remove Global metrics wiring; harden cron fallback |
| `pkg/observability/` | Modified | Support per-user metrics instances |
| `pkg/storage/task_logs.go` | Modified | Add ownership JOIN in legacy mode |
| `pkg/config/workspace_init.go` | Modified | Canonical workspace init (consolidation) |
| `pkg/storage/user_storage.go` | Modified | Delegate to canonical workspace init |
| `pkg/session/manager.go` | Audited | Confirm per-user storage path; fix if needed |
| `pkg/storage/knowledge.go` | Comment | Document `DEFAULT 1` intent |
| `pkg/agent/manager.go` | Modified | Add `SetUserForAgent()` calls for specialists |
| `pkg/agent/orchestrator.go` | Modified | Add user context to `TeamContext` |
| `pkg/agent/swarm.go` | Modified | Ensure user context propagation in swarms |
| `pkg/agent/spawner.go` | Modified | Add `userID` to `SpecialistSpawnTask` |

---

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Workflows migration breaks existing legacy deployments | Medium | Use `ALTER TABLE ... ADD COLUMN` with `DEFAULT 0`; backfill `user_id` from context before any query |
| Removing Global metrics breaks legacy single-user mode | Low | Keep `Global()` as a no-op shim; route per-user metrics through the new path |
| CronService hard-fail breaks bot startup | Low | Guard only triggers if user IS authenticated — anonymous / startup paths unaffected |
| Workspace init consolidation causes double-init | Low | Function is idempotent (`MkdirAll`); no risk of data loss |
| Multi-agent user context breaks existing specialist spawning | Low | `SetUserForAgent()` is optional; graceful fallback to no-op if user context not set |

---

## Rollback Plan

- All schema changes use additive migrations (`ADD COLUMN`) — safe to roll back by reverting application code without touching DB files
- No data is deleted or moved
- Git revert of the PR restores previous behavior; per-user DBs are unaffected (no column added there)

---

## Dependencies

- Exploration complete (`explore.md`)
- No external library changes required

---

## Success Criteria

- [ ] `workflows` CRUD returns only the requesting user's data in both legacy and per-user modes
- [ ] `processNextTodoTaskLegacy` is explicitly documented and guarded against multi-user use
- [ ] `handleMetrics` returns only the requesting user's metrics
- [ ] `getCronServiceForRequest()` returns a structured error (not a silent fallback) when per-user cron is unavailable
- [ ] `SessionManager` storage path verified to be per-user for all authenticated channels
- [ ] `GetTaskLogs` in legacy mode cannot return logs for a task belonging to another user
- [ ] Single canonical workspace init function with 100% subdirectory coverage in tests
- [ ] All 8 issues from the audit are resolved or explicitly documented as accepted risk
- [ ] Specialists created by manager have `SetUserForAgent(userUUID, userID)` called
- [ ] `TeamContext` includes user identity when spawning specialists or tasks
- [ ] `SwarmRunner` executions propagate user context to child specialists
- [ ] `SpecialistSpawnTask` includes `userID` for audit trail
- [ ] All 9 issues are resolved (8 audit + 1 multi-agent)
