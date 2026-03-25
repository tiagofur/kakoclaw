# Design: User Data Isolation Hardening

**Change:** `audit-isolation-user-data`
**Date:** 2026-03-25
**Based on:** proposal.md, spec.md

---

## Technical Approach

Close 8 isolation gaps through targeted, additive changes. No schema drops, no data migration, no behavioral changes for per-user DB mode (already correct). Legacy shared-DB mode gets `user_id` scoping where missing. All changes follow the `isUserDB` pattern already established in `pkg/storage/sqlite.go`.

---

## Architecture Decisions

| Decision | Options | Choice | Rationale |
|---|---|---|---|
| ISSUE-1: Metrics | A) Per-user `Metrics` instances via `getUserStorage` context; B) `user_id` column on existing tables | **Option A** | Zero schema change; mirrors exact pattern of per-user storage injection; `Metrics.New()` already exists; lower DB overhead |
| ISSUE-1: Global() fate | Keep as-is; make no-op; remove | **Keep as no-op** | Single-user legacy mode still needs a non-nil target; `Global().SetStorage(s.store)` line removed but `Global()` function survives for backward compat |
| ISSUE-4: Migration strategy | Full table recreation; `ALTER TABLE ADD COLUMN` | **ALTER TABLE ADD COLUMN** | Additive — existing rows survive with `DEFAULT 0`; safe rollback (revert code, column stays inert) |
| ISSUE-8: SessionManager | Fix path in `MultiUserChannelManager`; no-op (already correct) | **No-op — already correct** | `agent.NewAgentLoopForUser` calls `session.NewSessionManager(filepath.Join(workspace, "sessions"))` with the per-user workspace path; no fix needed, only documentation |
| ISSUE-7: Cron fallback | Remove shared fallback; hard error; sentinel error | **Sentinel error `ErrCronNotInitialized`** | All callers already check `ok`; clean API contract with no silent degradation for authenticated users |
| ISSUE-6: Consolidation | Move logic to `pkg/config`; move to `pkg/storage`; shared utility pkg | **Canonical in `pkg/config`** — `EnsureUserWorkspace` is already the more complete impl (includes templates + memory file); `user_storage.go` delegates |

---

## Current State vs Target State

```
CURRENT
═══════
server.go:253 ── observability.Global().SetStorage(s.store)  ← shared store, all users
handleMetrics   ── observability.Global().Snapshot()          ← all-user aggregate
processNextTodoTaskLegacy ── no guard, runs in multi-user mode
getCronServiceForRequest  ── falls back to s.cronService silently
workflows table ── no user_id column (legacy shared DB)
GetTaskLogs     ── no ownership JOIN (legacy shared DB)
EnsureUserDirectory ── duplicate of EnsureUserWorkspace (diverged templates)
SessionManager  ── ✅ already per-user via NewAgentLoopForUser

TARGET
══════
server.go       ── observability.Global().SetStorage() removed
handleMetrics   ── resolves per-user Metrics via getUserStorage context
processNextTodoTaskLegacy ── guard: return if s.userMgr != nil
getCronServiceForRequest  ── return nil, ErrCronNotInitialized for authenticated users
workflows table ── user_id column added via ALTER TABLE; CRUD scoped by isUserDB
GetTaskLogs     ── ownership JOIN in legacy mode (isUserDB == false)
EnsureUserDirectory ── delegates to config.EnsureUserWorkspace
```

---

## Phase 1 — HIGH

### ISSUE-3: Legacy Task Worker Guard

**File:** `pkg/web/server.go` — `processNextTodoTaskLegacy`

```
processNextTodoTask()
  ├─ if userMgr != nil && centralStore != nil → processNextTodoTaskPerUser()   [existing]
  └─ else → processNextTodoTaskLegacy()   ← ADD: early return + error log if userMgr != nil
```

Changes:
1. Add early-return guard at top of `processNextTodoTaskLegacy`:
   ```go
   // SINGLE-USER ONLY: This path is unsafe for multi-user deployments.
   // Use processNextTodoTaskPerUser when s.userMgr != nil.
   if s.userMgr != nil {
       logger.ErrorCF("task-worker", "processNextTodoTaskLegacy called in multi-user mode — skipping",
           map[string]interface{}{"user_mgr_present": true})
       return
   }
   ```
2. Add startup check in `server.go` (after store is wired):
   ```go
   if s.userMgr == nil && s.store != nil {
       if count, err := s.store.CountUsers(); err == nil && count > 1 {
           logger.ErrorCF("task-worker", "UNSAFE: legacy mode with multiple users detected",
               map[string]interface{}{"user_count": count})
       }
   }
   ```
3. `CountUsers()` needs to be added to `Storage` interface (simple `SELECT COUNT(*) FROM users`).

**Rollback:** Revert guard lines — no data changed.

---

### ISSUE-4: Workflows Table Migration

**File:** `pkg/storage/workflow.go` — `migrateWorkflows()`

**SQL DDL (additive migration appended to `migrateWorkflows`):**
```sql
-- Additive migration: add user_id to all workflow tables
-- Uses the same pragma_table_info pattern already used for 'parameters' column.
ALTER TABLE workflows         ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE workflow_runs     ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE workflow_approvals ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_workflows_user_id        ON workflows(user_id);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_user_id    ON workflow_runs(user_id);
CREATE INDEX IF NOT EXISTS idx_workflow_approvals_user_id ON workflow_approvals(user_id);
```

Applied using the same `pragma_table_info` column-existence check:
```go
var colExists int
s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('workflows') WHERE name='user_id'`).Scan(&colExists)
if colExists == 0 {
    s.db.Exec(`ALTER TABLE workflows ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0`)
    s.db.Exec(`ALTER TABLE workflow_runs ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0`)
    s.db.Exec(`ALTER TABLE workflow_approvals ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0`)
    s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_workflows_user_id ON workflows(user_id)`)
    // ... indexes for other tables
}
```

**CRUD changes — mirrors `tasks` pattern (`isUserDB` flag):**

| Method | Legacy mode (`isUserDB=false`) | Per-user mode (`isUserDB=true`) |
|--------|-------------------------------|----------------------------------|
| `ListWorkflows(userID int64)` | `WHERE user_id = ?` | no filter |
| `GetWorkflow(id, userID int64)` | `WHERE id=? AND user_id=?` | `WHERE id=?` |
| `CreateWorkflow(userID int64, ...)` | INSERT includes `user_id` | INSERT omits `user_id` |
| `UpdateWorkflow(id, userID int64, ...)` | `WHERE id=? AND user_id=?` | `WHERE id=?` |
| `DeleteWorkflow(id, userID int64)` | `WHERE id=? AND user_id=?` | `WHERE id=?` |

`handleWorkflows` currently uses `store, _, ok := s.getUserStorage(r)` — change to capture `userUUID` and resolve `userID` from central store. Pass `userID` to all workflow CRUD calls.

**Backfill:** Existing rows get `user_id=0`. In legacy deployments with one user, the operator runs:
```sql
UPDATE workflows SET user_id = (SELECT id FROM users LIMIT 1);
```
This is not automated — documented in release notes.

**Rollback:** Revert CRUD code to old signatures. The `user_id` column stays in DB but is ignored — no data loss, no errors.

---

## Phase 2 — HIGH (ISSUE-9: Multi-Agent User Context Propagation)

### ISSUE-9: Multi-Agent User Context Propagation

**Files:**
- `pkg/agent/manager.go` — `InitializeOrchestrator` method
- `pkg/agent/orchestrator.go` — `TeamContext` struct and task delegation
- `pkg/agent/swarm.go` — `SwarmRunner` user context propagation
- `pkg/agent/spawner.go` — `SpecialistSpawnTask` struct
- `pkg/agent/loop.go` — `SetUserForAgent` method

#### Issue 9.1: Specialist User Context in Manager

**Finding:** Specialists are created in `manager.go:InitializeOrchestrator` via `NewAgentLoop(globalCfg)` but `SetUserForAgent(userUUID, userID)` is not called. This means specialists lack user context for workspace isolation.

**Fix:** Add `SetUserForAgent` call after each specialist creation.

```go
func (m *Manager) InitializeOrchestrator(userUUID string, userID int64, globalCfg *config.Config) error {
    // ... existing code ...
    
    specialist := agent.NewAgentLoop(globalCfg)
    specialist.SetUserForAgent(userUUID, userID) // ← ADD THIS
    
    m.orchestrator = orchestrator.New(specialist)
    // ...
}
```

#### Issue 9.2: TeamContext User Identity

**Finding:** `TeamContext` struct in `orchestrator.go` may not include user identity fields, causing specialists to spawn without user context.

**Fix:** Add `userUUID` and `userID` to `TeamContext` struct.

```go
type TeamContext struct {
    ParentAgentID string
    TaskID        string
    TaskType      string
    
    // User context for isolation
    UserUUID string // ← ADD
    UserID   int64  // ← ADD
    
    // ... existing fields ...
}
```

Update `TaskDecompositionTool` to populate these fields from the parent agent's context:

```go
func (t *TaskDecompositionTool) Execute(ctx context.Context, input string) (string, error) {
    userUUID := t.agent.GetUserUUID()
    userID := t.agent.GetUserID()
    
    teamCtx := TeamContext{
        ParentAgentID: t.agent.ID(),
        UserUUID:      userUUID, // ← ADD
        UserID:        userID,   // ← ADD
        // ...
    }
    // ...
}
```

#### Issue 9.3: SwarmRunner User Context Propagation

**Finding:** `SwarmRunner` may bypass per-user workspace isolation by not propagating user context to spawned specialists.

**Fix:** Ensure swarm executor receives and passes user context.

```go
func (sr *SwarmRunner) Execute(ctx context.Context, userUUID string, userID int64, swarm *SwarmConfig) error {
    for _, specialistCfg := range swarm.Specialists {
        specialist := agent.NewAgentLoop(globalCfg)
        specialist.SetUserForAgent(userUUID, userID) // ← PROPAGATE USER CONTEXT
        
        // ... execute specialist ...
    }
}
```

#### Issue 9.4: SpecialistSpawnTask Audit Trail

**Finding:** `SpecialistSpawnTask` struct in `spawner.go` lacks `userID` field for audit tracking.

**Fix:** Add `UserID` field.

```go
type SpecialistSpawnTask struct {
    TaskID      string
    AgentID     string
    Specialist  string
    
    // Audit trail
    UserID int64  // ← ADD
    UserUUID string // ← ADD (optional for workspace)
    
    Status      string
    CreatedAt   time.Time
    CompletedAt *time.Time
}
```

Update spawn logic to set `UserID` from parent agent context:

```go
func (s *SpecialistSpawner) Spawn(ctx context.Context, parentAgent *AgentLoop, specialistType string) (*Specialist, error) {
    task := SpecialistSpawnTask{
        TaskID:     generateTaskID(),
        UserID:     parentAgent.GetUserID(),   // ← ADD
        UserUUID:   parentAgent.GetUserUUID(), // ← ADD
        Specialist: specialistType,
        Status:     "pending",
        CreatedAt:  time.Now(),
    }
    // ...
}
```

**Rollback:** All changes are additive. Remove `SetUserForAgent` calls and struct fields. No data migration required.

---

## Phase 3 — MEDIUM

### ISSUE-1: Per-User Metrics

**Architecture decision: Option A — per-user `Metrics` instances.**

**Data flow:**
```
HTTP request
  └─ handleMetrics(w, r)
       └─ getUserStorage(r) → (userStore, userUUID, ok)
            └─ getOrCreateUserMetrics(userUUID) → *observability.Metrics
                 └─ metrics.Snapshot() → response
```

**New component:** `pkg/web/server.go` — `userMetrics map[string]*observability.Metrics` + `userMetricsMu sync.RWMutex`

```go
func (s *Server) getOrCreateUserMetrics(userUUID string) *observability.Metrics {
    s.userMetricsMu.RLock()
    if m, ok := s.userMetrics[userUUID]; ok {
        s.userMetricsMu.RUnlock()
        return m
    }
    s.userMetricsMu.RUnlock()
    s.userMetricsMu.Lock()
    defer s.userMetricsMu.Unlock()
    if m, ok := s.userMetrics[userUUID]; ok { return m }
    m := observability.New()
    s.userMetrics[userUUID] = m
    return m
}
```

**`handleMetrics` change:**
```go
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
    _, userUUID, ok := s.getUserStorage(r)
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    snapshot := s.getOrCreateUserMetrics(userUUID).Snapshot()
    writeJSONResponse(w, snapshot)
}
```

**`observability.Global().SetStorage(s.store)` removal** (server.go:253): Remove this line. `Global()` becomes an unused no-op singleton — intentionally kept for backward compat but not wired.

**Metrics recording:** The agent loop's `RecordLLMCall` / `RecordToolCall` / `RecordAgentRun` currently call `observability.Global()`. These calls must be redirected to the per-user instance. The `AgentLoop` already carries `userUUID` — add `metrics *observability.Metrics` field and inject it at construction in `NewAgentLoopForUser`. The `Server` passes the per-user metrics instance when creating the loop.

**Rollback:** Remove `userMetrics` map, restore `Global().SetStorage(s.store)` and `Global().Snapshot()` calls.

---

### ISSUE-7: CronService Hard Failure

**File:** `pkg/web/server.go` — `getCronServiceForRequest()`

**New sentinel error** in `pkg/cron/service.go` (or a new `pkg/cron/errors.go`):
```go
var ErrCronNotInitialized = errors.New("per-user cron service not initialized")
```

**Signature change:**
```go
// Before: func (s *Server) getCronServiceForRequest(r *http.Request) (*cron.CronService, int64, bool)
// After:  func (s *Server) getCronServiceForRequest(r *http.Request) (*cron.CronService, int64, error)
```

**Logic change** (replace lines 723-727):
```go
// Before:
if s.cronService != nil {
    return s.cronService, userID, true
}
return nil, userID, true

// After (authenticated user, multi-user mode):
if s.multiUserChannelManager != nil {
    return nil, userID, cron.ErrCronNotInitialized
}
// Legacy single-user mode: shared cron is acceptable
if s.cronService != nil {
    return s.cronService, userID, nil
}
return nil, userID, cron.ErrCronNotInitialized
```

All callers in `handlers_advanced.go` use `cronService, userID, ok := s.getCronServiceForRequest(r)` — update to `err` and return 500 with `ErrCronNotInitialized` message (opaque to client, logged internally).

**Rollback:** Revert signature + fallback logic. Callers revert to `ok bool`.

---

### ISSUE-8: SessionManager Audit — ALREADY CORRECT

**Finding:** `pkg/agent/loop.go:315` calls `session.NewSessionManager(filepath.Join(workspace, "sessions"))` where `workspace` is set to `~/.MakoClaw/users/<uuid>/workspace` for every user via `NewAgentLoopForUser`. The `multiuser_manager.go` calls `agent.NewAgentLoopForUser(userUUID, ...)` per user.

**Action:** No code change required. Add comment in `multiuser_manager.go` near line 114:
```go
// Each user's AgentLoop gets its own SessionManager pointed at
// {userWorkspace}/sessions/ — cross-user session access is structurally impossible.
```

**Note:** `multiuser_manager.go:108-109` opens a `data.db` under workspace via `storage.New()` (NOT `isUserDB=true`). This is a separate legacy path and pre-existing — not in scope for this change.

---

## Phase 3 — LOW/TRIVIAL

### ISSUE-2: Task Logs Ownership Guard

**File:** `pkg/storage/task_logs.go` — `GetTaskLogs`

```go
func (s *Storage) GetTaskLogs(taskID int64, userID int64) ([]TaskLog, error) {
    var query string
    var args []interface{}
    if s.isUserDB {
        // Per-user DB: isolation by file, no ownership filter needed
        query = `SELECT id, task_id, event, message, created_at FROM task_logs WHERE task_id = ? ORDER BY created_at ASC`
        args = []interface{}{taskID}
    } else {
        // Legacy shared DB: verify task ownership to prevent cross-user access
        query = `SELECT tl.id, tl.task_id, tl.event, tl.message, tl.created_at
                 FROM task_logs tl
                 JOIN tasks t ON t.id = tl.task_id AND t.user_id = ?
                 WHERE tl.task_id = ?
                 ORDER BY tl.created_at ASC`
        args = []interface{}{userID, taskID}
    }
    // ... rows scan unchanged
}
```

All callers updated to pass `userID` (extractable from `getUserStorage(r)` context).

**Rollback:** Revert method signature, remove JOIN. No schema change.

---

### ISSUE-6: Workspace Init Consolidation

**Files:**
- `pkg/config/workspace_init.go` — already canonical (more complete templates); keep as-is
- `pkg/storage/user_storage.go:EnsureUserDirectory` — replace body with single delegate call

```go
func (m *UserStorageManager) EnsureUserDirectory(userUUID string) error {
    _, err := config.EnsureUserWorkspace(userUUID)
    return err
}
```

The existing `EnsureUserWorkspace` returns `(string, error)` — the workspace path. The delegate discards the path (callers of `EnsureUserDirectory` don't use it).

**Rollback:** Restore original body. No data impact (both are idempotent).

---

### ISSUE-5: Documentation Only

**File:** `pkg/storage/knowledge.go:43`

Add inline comment above the migration line:
```go
// user_id DEFAULT 1: In per-user DB mode, this column is always populated
// explicitly by SaveKnowledgeDocument(userID, ...). The DEFAULT 1 is a
// safety net that is never triggered in practice — it would only matter if
// this schema were used in legacy shared-DB mode, where it would be incorrect.
// Acceptable: per-user DB is the production path; legacy mode does not use
// knowledge storage with multiple users in practice.
`user_id INTEGER NOT NULL DEFAULT 1,`
```

---

## Phase 3 — MEDIUM

### ISSUE-1: Per-User Metrics

| File | Action | Description |
|------|--------|-------------|
| `pkg/storage/workflow.go` | Modify | Add migration for `user_id` columns + indexes; update all CRUD signatures with `userID int64` + `isUserDB` branching |
| `pkg/storage/task_logs.go` | Modify | Add `userID int64` param to `GetTaskLogs`; ownership JOIN in legacy mode |
| `pkg/storage/sqlite.go` | Modify | `CountUsers()` helper (for ISSUE-3 startup check) |
| `pkg/web/server.go` | Modify | Remove `Global().SetStorage`; add `userMetrics` map + `getOrCreateUserMetrics`; update `handleMetrics`; add startup multi-user guard; update `processNextTodoTaskLegacy` guard; update `getCronServiceForRequest` signature |
| `pkg/web/handlers_advanced.go` | Modify | Update `getCronServiceForRequest` callers; update workflow handlers to pass `userID` |
| `pkg/cron/errors.go` | Create | `var ErrCronNotInitialized = errors.New(...)` |
| `pkg/storage/user_storage.go` | Modify | `EnsureUserDirectory` delegates to `config.EnsureUserWorkspace` |
| `pkg/storage/knowledge.go` | Modify | Add inline comment at `DEFAULT 1` |
| `pkg/channels/multiuser_manager.go` | Modify | Add comment confirming per-user SessionManager isolation |
| `pkg/observability/metrics.go` | Modify | Optionally expose `New()` for per-user construction — already exported |
| `pkg/agent/manager.go` | Modify | Add `SetUserForAgent(userUUID, userID)` calls for all specialists in `InitializeOrchestrator` |
| `pkg/agent/orchestrator.go` | Modify | Add `userUUID` and `userID` to `TeamContext` struct; populate from parent agent |
| `pkg/agent/swarm.go` | Modify | Ensure `SwarmRunner` propagates user context to spawned specialists |
| `pkg/agent/spawner.go` | Modify | Add `UserID` and `UserUUID` fields to `SpecialistSpawnTask` struct |
| `pkg/agent/loop.go` | Verify | Ensure `SetUserForAgent` method exists and handles user context correctly |

---

## Sequence Diagrams

### Metrics Injection Path (after fix)

```
HTTP GET /api/v1/metrics
  │
  ▼
handleMetrics(r)
  │
  ├─ getUserStorage(r) ──→ (_, userUUID, ok)
  │       │
  │       └─ jwtClaims.UUID → userMgr.GetOrCreate(uuid)
  │
  ├─ getOrCreateUserMetrics(userUUID)
  │       │
  │       └─ userMetrics map[userUUID] → *Metrics (created on first access)
  │
  └─ metrics.Snapshot() ──→ JSON response (user-scoped only)
```

### getCronServiceForRequest Control Flow (after fix)

```
getCronServiceForRequest(r)
  │
  ├─ resolve userID from JWT claims
  │
  ├─ if multiUserChannelManager != nil
  │     ├─ GetCronServiceForUser(uuid)  → found? return it ✓
  │     ├─ GetOrCreateCronServiceForUser(uuid) → ok? return it ✓
  │     └─ else → return nil, ErrCronNotInitialized ✗
  │
  └─ else (legacy single-user)
        ├─ s.cronService != nil → return it ✓
        └─ else → return nil, ErrCronNotInitialized ✗
```

### Workflow Migration Execution (on DB open)

```
NewUserStorage(dbPath) / New(cfg)
  │
  └─ migrateWorkflows()
        │
        ├─ CREATE TABLE IF NOT EXISTS workflows (no user_id yet)
        ├─ CREATE TABLE IF NOT EXISTS workflow_runs
        ├─ CREATE TABLE IF NOT EXISTS workflow_approvals
        │
        └─ CHECK pragma_table_info('workflows') WHERE name='user_id'
              │
              ├─ columnExists = 0 →
              │     ALTER TABLE workflows ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0
              │     ALTER TABLE workflow_runs ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0
              │     ALTER TABLE workflow_approvals ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0
              │     CREATE INDEX idx_workflows_user_id
              │     CREATE INDEX idx_workflow_runs_user_id
              │     CREATE INDEX idx_workflow_approvals_user_id
              │
              └─ columnExists = 1 → skip (idempotent)
```

---

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `processNextTodoTaskLegacy` exits with `userMgr != nil` | Mock `userMgr`, assert no task processed, assert log emitted |
| Unit | `GetTaskLogs` in legacy mode returns empty for wrong `user_id` | `isUserDB=false`, insert task for user 1, query as user 2 |
| Unit | `GetTaskLogs` in per-user mode ignores `user_id` | `isUserDB=true`, assert no JOIN in query (via results) |
| Unit | `EnsureUserDirectory` delegates to `config.EnsureUserWorkspace` | Spy on calls, assert same subdirs |
| Unit | `getCronServiceForRequest` returns `ErrCronNotInitialized` | Authenticated context, no cron registered |
| Integration | `ListWorkflows` in legacy mode returns only requesting user's rows | Insert 2 users' rows, query as user A, assert user B's rows absent |
| Integration | Workflow migration is idempotent and non-destructive | Pre-populate DB, run migration twice, assert row count unchanged |
| Integration | Per-user metrics: User A and B see only own data | Two users make LLM calls, each `GET /metrics` returns isolated counts |
| Integration | Session files land in correct per-user path | Write session as user A, assert file under `users/aaa/workspace/sessions/`, absent from `users/bbb/` |
| Unit | `InitializeOrchestrator` calls `SetUserForAgent` for specialists | Mock agent creation, assert `SetUserForAgent` invoked with correct params |
| Unit | `TeamContext` includes user identity | Create `TeamContext` via `TaskDecompositionTool`, assert `UserUUID` and `UserID` populated |
| Unit | `SwarmRunner` propagates user context | Mock swarm execution, verify child specialists receive parent's user context |
| Unit | `SpecialistSpawnTask` includes `UserID` | Create spawn task, assert `UserID` field present and correct |

---

## Open Questions

- [ ] `multiuser_manager.go:108-109` opens `data.db` under workspace via `storage.New()` (legacy mode, `isUserDB=false`) instead of `user.db` via `userMgr`. This is a pre-existing inconsistency not captured in the 8 issues — confirm it is explicitly out of scope for this change.
- [ ] Backfill of existing `workflows.user_id = 0` rows in production legacy deployments: confirm that a manual SQL snippet in release notes is sufficient, or if an automated migration is needed.
- [ ] Verify that `agent.NewAgentLoop(globalCfg)` signature accepts `SetUserForAgent` call without errors. Check if `globalCfg` needs to include user context fields.
