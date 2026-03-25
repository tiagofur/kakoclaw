# Tasks: User Data Isolation Hardening

**Change:** `audit-isolation-user-data`
**Date:** 2026-03-25
**Based on:** proposal.md, spec.md, design.md

---

## Phase 1 — HIGH Priority: Critical Data-Leak Fixes

### ISSUE-3: Legacy Task Worker Guard

- [x] **1.1** Add `CountUsers() (int, error)` to `Storage` interface and implement in `pkg/storage/sqlite.go` (`SELECT COUNT(*) FROM users`)
  - **Files:** `pkg/storage/sqlite.go`
  - **Dependencies:** none
  - **Acceptance:** `CountUsers()` returns correct count; compiles with interface satisfied
  - **Effort:** Small

- [x] **1.2** Add early-return guard at top of `processNextTodoTaskLegacy` in `pkg/web/server.go`; add `// SINGLE-USER ONLY` header comment
  - **Files:** `pkg/web/server.go`
  - **Dependencies:** none
  - **Acceptance:** Function returns immediately + emits `logger.ErrorCF("task-worker", ...)` when `s.userMgr != nil`
  - **Effort:** Small

- [x] **1.3** Add startup multi-user detection check in `server.go` (after store is wired): if `userMgr == nil` AND `CountUsers() > 1`, emit FATAL-level structured log
  - **Files:** `pkg/web/server.go`
  - **Dependencies:** 1.1
  - **Acceptance:** Log emitted on startup when legacy mode + multiple users detected; server continues without crashing
  - **Effort:** Small

- [x] **1.4** Unit test: `processNextTodoTaskLegacy` with `userMgr != nil` — assert no task processed + error log emitted
  - **Files:** `pkg/web/server_test.go` (or existing test file)
  - **Dependencies:** 1.2
  - **Acceptance:** Test passes; scenario from spec confirmed
  - **Effort:** Small

- [x] **1.5** Unit test: startup guard emits FATAL log when legacy mode with user count > 1
  - **Files:** `pkg/web/server_test.go`
  - **Dependencies:** 1.1, 1.3
  - **Acceptance:** Mock `CountUsers() = 2`, assert log output matches spec
  - **Effort:** Small

---

### ISSUE-4: Workflows Table Migration & CRUD Isolation

- [x] **1.6** Append `user_id` migration to `migrateWorkflows()` in `pkg/storage/workflow.go`: `ALTER TABLE ADD COLUMN` for `workflows`, `workflow_runs`, `workflow_approvals` + 3 indexes; guard with `pragma_table_info` column-existence check
  - **Files:** `pkg/storage/workflow.go`
  - **Dependencies:** none
  - **Acceptance:** Migration is idempotent; existing rows survive with `user_id=0`; all 3 indexes created
  - **Effort:** Small

- [x] **1.7** Update all workflow CRUD method signatures to accept `userID int64`; add `isUserDB` branch — legacy mode applies `WHERE user_id = ?` / includes `user_id` in INSERT; per-user mode skips filter (mirrors tasks pattern)
  - **Files:** `pkg/storage/workflow.go`
  - **Dependencies:** 1.6
  - **Acceptance:** `ListWorkflows(userID)`, `GetWorkflow(id, userID)`, `CreateWorkflow(userID, ...)`, `UpdateWorkflow(id, userID, ...)`, `DeleteWorkflow(id, userID)` all scoped; compiles
  - **Effort:** Medium

- [x] **1.8** Update `handleWorkflows` in `pkg/web/handlers_advanced.go` to capture `userID` from `getUserStorage(r)` context and pass it to all workflow CRUD calls
  - **Files:** `pkg/web/handlers_advanced.go`
  - **Dependencies:** 1.7
  - **Acceptance:** Handler resolves `userID`; all CRUD calls receive correct `userID`; no compile errors
  - **Effort:** Small

- [ ] **1.9** Integration test: `ListWorkflows` in legacy mode returns only requesting user's rows
  - **Files:** `pkg/storage/workflow_test.go`
  - **Dependencies:** 1.7
  - **Acceptance:** Insert rows for user A and user B; query as user A returns only user A rows
  - **Effort:** Small

- [ ] **1.10** Integration test: workflow migration is idempotent and non-destructive
  - **Files:** `pkg/storage/workflow_test.go`
  - **Dependencies:** 1.6
  - **Acceptance:** Pre-populate DB; run migration twice; assert row count unchanged; all `user_id` columns present
  - **Effort:** Small

- [ ] **1.11** Document backfill SQL snippet for existing `user_id=0` rows in `openspec/changes/audit-isolation-user-data/release-notes.md`
  - **Files:** `openspec/changes/audit-isolation-user-data/release-notes.md` (new)
  - **Dependencies:** 1.6
  - **Acceptance:** Release notes contain `UPDATE workflows SET user_id = (SELECT id FROM users LIMIT 1)` snippet with explanation; open question #2 resolved as manual/documented
  - **Effort:** Small

---

## Phase 2 — MEDIUM Priority: Information Leakage / Silent Misconfiguration

### ISSUE-1: Per-User Metrics Instances

- [ ] **2.1** Add `userMetrics map[string]*observability.Metrics` field and `userMetricsMu sync.RWMutex` to `Server` struct; initialize map in `NewServer` / constructor; implement `getOrCreateUserMetrics(userUUID string) *observability.Metrics` method
  - **Files:** `pkg/web/server.go`
  - **Dependencies:** none
  - **Acceptance:** Double-checked locking pattern correct; map initialized; `New()` called for first access per UUID
  - **Effort:** Small

- [ ] **2.2** Remove `observability.Global().SetStorage(s.store)` from `server.go:253`; `Global()` shim survives as no-op
  - **Files:** `pkg/web/server.go`
  - **Dependencies:** 2.1
  - **Acceptance:** Line removed; legacy single-user mode does not panic; `Global()` function still exists
  - **Effort:** Small

- [ ] **2.3** Update `handleMetrics` to call `getUserStorage(r)` → resolve `userUUID` → call `getOrCreateUserMetrics(userUUID).Snapshot()`; return 401 if `!ok`
  - **Files:** `pkg/web/server.go`
  - **Dependencies:** 2.1
  - **Acceptance:** Handler only returns per-user metrics; returns 401 for unauthenticated requests
  - **Effort:** Small

- [ ] **2.4** Add `metrics *observability.Metrics` field to `AgentLoop`; inject per-user instance at construction in `NewAgentLoopForUser`; redirect `RecordLLMCall` / `RecordToolCall` / `RecordAgentRun` from `Global()` to the injected instance
  - **Files:** `pkg/agent/loop.go`, `pkg/web/server.go` (construction site)
  - **Dependencies:** 2.1
  - **Acceptance:** Agent metrics recorded to per-user instance, not global; server passes correct instance at `NewAgentLoopForUser`
  - **Effort:** Medium

- [ ] **2.5** Integration test: two users generate LLM call metrics; each `GET /api/metrics` returns only own data
  - **Files:** `pkg/web/server_test.go` or `pkg/web/metrics_test.go`
  - **Dependencies:** 2.3, 2.4
  - **Acceptance:** User A's metrics contain only User A's counts; no cross-contamination
  - **Effort:** Small

---

### ISSUE-7: CronService Hard Failure

- [ ] **2.6** Create `pkg/cron/errors.go` with `var ErrCronNotInitialized = errors.New("per-user cron service not initialized")`
  - **Files:** `pkg/cron/errors.go` (new)
  - **Dependencies:** none
  - **Acceptance:** File exists; exported sentinel error compiles
  - **Effort:** Small

- [ ] **2.7** Update `getCronServiceForRequest` signature in `pkg/web/server.go`: change return type from `(*cron.CronService, int64, bool)` to `(*cron.CronService, int64, error)`; replace silent shared-cron fallback with `return nil, userID, cron.ErrCronNotInitialized` for authenticated multi-user requests; legacy single-user path returns `nil` error
  - **Files:** `pkg/web/server.go`
  - **Dependencies:** 2.6
  - **Acceptance:** Authenticated user + no per-user cron → returns `ErrCronNotInitialized`; legacy path unaffected
  - **Effort:** Small

- [ ] **2.8** Update all callers of `getCronServiceForRequest` in `pkg/web/handlers_advanced.go`: change `ok bool` to `err error`; return HTTP 500 on `ErrCronNotInitialized` with opaque message + internal structured log
  - **Files:** `pkg/web/handlers_advanced.go`
  - **Dependencies:** 2.7
  - **Acceptance:** All callers updated; no `ok bool` usages remain; 500 returned to client; error logged with user UUID
  - **Effort:** Small

- [ ] **2.9** Unit test: authenticated context + no registered cron → `getCronServiceForRequest` returns `ErrCronNotInitialized`
  - **Files:** `pkg/web/server_test.go`
  - **Dependencies:** 2.7
  - **Acceptance:** Test passes; error matches sentinel
  - **Effort:** Small

---

### ISSUE-8: SessionManager Audit (Comment Only)

- [ ] **2.10** Add comment in `pkg/channels/multiuser_manager.go` near line 114 confirming per-user `SessionManager` isolation (path is `{userWorkspace}/sessions/`, cross-user access is structurally impossible)
  - **Files:** `pkg/channels/multiuser_manager.go`
  - **Dependencies:** none
  - **Acceptance:** Comment present; no code change; ISSUE-8 closed as confirmed-correct
  - **Effort:** Small

- [ ] **2.11** Integration test: write session as User A, attempt read via User B's `SessionManager`, assert not-found error
  - **Files:** `pkg/session/manager_test.go` or `pkg/channels/multiuser_manager_test.go`
  - **Dependencies:** 2.10
  - **Acceptance:** Session file under `users/aaa/.../sessions/` not visible via `users/bbb/` manager
  - **Effort:** Small

---

## Phase 3 — LOW Priority: Code Quality & Documentation

### ISSUE-2: Task Logs Ownership Guard

- [ ] **3.1** Update `GetTaskLogs` signature in `pkg/storage/task_logs.go` to accept `userID int64`; add `isUserDB` branch — legacy mode executes ownership `JOIN tasks ON tasks.id = task_logs.task_id AND tasks.user_id = ?`; per-user mode skips JOIN
  - **Files:** `pkg/storage/task_logs.go`
  - **Dependencies:** none
  - **Acceptance:** Legacy mode returns empty set for cross-user `task_id`; per-user mode unaffected
  - **Effort:** Small

- [ ] **3.2** Update all callers of `GetTaskLogs` (handlers) to pass `userID` extracted from `getUserStorage(r)` context
  - **Files:** `pkg/web/handlers_advanced.go` (or wherever `GetTaskLogs` is called)
  - **Dependencies:** 3.1
  - **Acceptance:** No callers use old zero-arg signature; compiles
  - **Effort:** Small

- [ ] **3.3** Unit test: `GetTaskLogs` in legacy mode returns empty list when `userID` does not own the `task_id`
  - **Files:** `pkg/storage/task_logs_test.go`
  - **Dependencies:** 3.1
  - **Acceptance:** Insert task for user 1; query as user 2 (`isUserDB=false`); assert empty result, no error
  - **Effort:** Small

---

### ISSUE-6: Workspace Init Consolidation

- [ ] **3.4** Replace body of `UserStorageManager.EnsureUserDirectory` in `pkg/storage/user_storage.go` with single delegate call: `_, err := config.EnsureUserWorkspace(userUUID); return err`
  - **Files:** `pkg/storage/user_storage.go`
  - **Dependencies:** none (canonical function already exists in `pkg/config/workspace_init.go`)
  - **Acceptance:** Function body replaced; callers receive same behavior; idempotent
  - **Effort:** Small

- [ ] **3.5** Unit test: `EnsureUserWorkspace` on a fresh temp dir creates all expected subdirectories (workspace, memory/, sessions/, skills/, cron/, tasks/, temp/)
  - **Files:** `pkg/config/workspace_init_test.go`
  - **Dependencies:** none
  - **Acceptance:** All paths from spec §ISSUE-6 exist after one call
  - **Effort:** Small

- [ ] **3.6** Unit test: `EnsureUserDirectory` delegates to `config.EnsureUserWorkspace` (spy/mock or verify identical subdirectory outcome)
  - **Files:** `pkg/storage/user_storage_test.go`
  - **Dependencies:** 3.4, 3.5
  - **Acceptance:** Delegating call confirmed; output identical to canonical function
  - **Effort:** Small

---

### ISSUE-5: Documentation

- [ ] **3.7** Add inline comment to `pkg/storage/knowledge.go` at `user_id INTEGER NOT NULL DEFAULT 1` explaining: why DEFAULT 1 exists, that it is safe in per-user mode, and that `userID` is always passed explicitly at write time
  - **Files:** `pkg/storage/knowledge.go`
  - **Dependencies:** none
  - **Acceptance:** Comment present immediately adjacent to the schema line; matches spec wording
  - **Effort:** Small

---

### Open Questions Resolution

- [ ] **3.8** Confirm `multiuser_manager.go:108-109` (`storage.New()` with `isUserDB=false`) is explicitly out of scope: add a `// TODO(future): ...` comment and note in `openspec/changes/audit-isolation-user-data/explore.md` or release notes
  - **Files:** `pkg/channels/multiuser_manager.go`, `openspec/changes/audit-isolation-user-data/release-notes.md`
  - **Dependencies:** none
  - **Acceptance:** Open question #1 documented and closed; no code change to this path in this change
  - **Effort:** Small

---

## Summary

| Phase | Tasks | Focus |
|-------|-------|-------|
| Phase 1 | 1.1 – 1.11 (11 tasks) | HIGH: task worker guard + workflows isolation |
| Phase 2 | 2.1 – 2.11 (11 tasks) | MEDIUM: metrics, cron, session audit |
| Phase 3 | 3.1 – 3.8 (8 tasks) | LOW: task logs, workspace consolidation, docs |
| **Total** | **30 tasks** | |

## Implementation Order Notes

- **Start with 1.1** (CountUsers) as it unblocks 1.3.
- **1.6 before 1.7** — migration must exist before CRUD changes reference the column.
- **1.7 before 1.8** — handler depends on updated storage signatures.
- **2.1 before 2.2, 2.3, 2.4** — map must exist before any method touches it.
- **2.6 before 2.7 before 2.8** — error type → signature → callers.
- **3.4 can run anytime** — canonical function already exists; delegation is a one-liner.
- **All Phase 3 tasks are independent** — can be parallelized across files.
