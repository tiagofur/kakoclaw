# Exploration: Audit User Data Isolation

**Change:** `audit-isolation-user-data`
**Date:** 2026-03-25
**Status:** Complete

---

## Goal

Audit ALL user-specific data access in MakoClaw to ensure it is properly isolated per user. Each user's data must be confined to their personal workspace (`~/.MakoClaw/users/<uuid>/`) and their personal database (`user.db`). Only auth/login data should remain in the central/shared database.

---

## Architecture Overview

The system uses a **dual-database design**:

### 1. Central DB (`~/.MakoClaw/central.db`)
Managed by `CentralStorage` (`pkg/storage/central.go`).

**Correct tables (auth/global only):**
- `users` — UUID, role, blocked status, token_version, auth credentials
- `settings` — global JWT secret
- `channel_users` — channel-to-user whitelists
- `setup_sessions` — onboarding tokens
- `skill_submissions`, `skill_ratings`, `skill_bundles`, `skill_categories` — marketplace (global, intentionally shared)

### 2. Per-user DB (`~/.MakoClaw/users/<uuid>/user.db`)
Managed by `UserStorageManager` (`pkg/storage/user_storage.go`).

**Tables (isolation by file — no `user_id` column needed):**
- `chats`, `sessions` — conversations
- `tasks`, `task_logs` — background tasks
- `prompts` — prompt templates
- `metrics_counters`, `metrics_events`, `metrics_breakdowns` — usage metrics
- `knowledge_documents`, `knowledge_chunks`, `knowledge_fts` — knowledge base
- `workflows`, `workflow_runs`, `workflow_approvals` — automations
- `user_providers` — per-user AI provider configs
- `skill_usage_events` — skill analytics

### 3. Legacy Shared DB (BACKWARD COMPAT)
The `Storage.migrate()` method in `pkg/storage/sqlite.go` creates a full single-DB schema with `user_id` columns for isolation (old approach). Still used in single-user/legacy deployments via `s.store` on the `Server` struct.

---

## Per-user Workspace Structure

`UserStorageManager.EnsureUserDirectory()` and `config.EnsureUserWorkspace()` both create:

```
~/.MakoClaw/users/<uuid>/
  user.db                    ← per-user SQLite DB
  config.json                ← per-user config overrides
  workspace/
    AGENTS.md
    SOUL.md
    USER.md
    IDENTITY.md
    memory/
      MEMORY.md
    sessions/                ← SessionManager JSON files
    skills/
    cron/
      jobs.json              ← per-user CronService store
    tasks/
    temp/
```

---

## Issues Found

### ISSUE-1 — `observability.Global()` is a process-wide singleton wired to the legacy shared store

**File:** `pkg/web/server.go:253`
```go
observability.Global().SetStorage(s.store)
```

The `Metrics` singleton aggregates LLM calls, tool calls, and agent runs for the **entire process** without user attribution. It persists to `s.store` (the legacy shared DB), not to each user's `user.db`.

**Impact:** Metrics data in `handleMetrics` is not isolated — any authenticated user can read the aggregate metrics of all users.

**Fix direction:** Either (a) use per-user `Metrics` instances stored in each user's DB, or (b) add `user_id` attribution to metric events and filter by requester.

---

### ISSUE-2 — `GetTaskLogs` has no user ownership check

**File:** `pkg/web/server.go:1023`
```go
logs, err := store.GetTaskLogs(logsID)
```

`store` here IS the per-user storage (resolved via `getUserStorage(r)` at line 800), so isolation by DB file is correct. **However**, `GetTaskLogs` itself (in `pkg/storage/task_logs.go`) does not verify that the `task_id` belongs to this user's DB. Since the task was created in the user's own DB, this is safe in practice — but there is no explicit ownership guard.

**Impact:** Low risk in per-user DB mode (isolation is by DB file). Medium risk in legacy shared-DB mode where task IDs from other users could be probed.

**Fix direction:** In legacy mode, add a `JOIN tasks ON tasks.id = task_logs.task_id AND tasks.user_id = ?` guard.

---

### ISSUE-3 — `processNextTodoTaskLegacy` calls `s.store.ListAllUsersTasks` (cross-user scan)

**File:** `pkg/web/server.go:2706-2718`
```go
func (s *Server) processNextTodoTaskLegacy(ctx context.Context) {
    tasks, err := s.store.ListAllUsersTasks(false)
    ...
    s.processTodoTaskWithLoop(ctx, s.store, s.agentLoop, t.UserID, t)
```

This is the legacy path: it reads tasks for ALL users from the shared store and processes them with a single shared `agentLoop`. There is no per-user sandboxing of the agent loop here.

**Impact:** In legacy mode, one user's task could be executed using another user's agent context (if `agentLoop` does not re-scope to the task's `UserID`). This is an EXISTING design flaw in legacy mode.

**Mitigation in place:** The new per-user path (`processNextTodoTaskPerUser`) at line 2721 correctly opens per-user stores and creates per-user agent loops. It takes precedence when `s.userMgr != nil && s.centralStore != nil`.

**Fix direction:** The legacy path should be removed or further guarded. Document that `s.store` + `s.agentLoop` only work safely in single-user deployments.

---

### ISSUE-4 — `workflows` table has NO `user_id` column

**File:** `pkg/storage/workflow.go`

The `workflows`, `workflow_runs`, and `workflow_approvals` tables have no `user_id` column and no user-scoping in any CRUD method. `ListWorkflows()` returns ALL workflows with no filter.

In per-user DB mode (each user has their own `user.db`), this is fine — isolation is by file.

**But:** If `s.store` (legacy shared DB) is used to serve workflow requests, ALL users see ALL workflows.

**Verification needed:** Check which store is used in the workflow handlers.

---

### ISSUE-5 — `knowledge_documents` migration creates `user_id NOT NULL DEFAULT 1`

**File:** `pkg/storage/knowledge.go:43`
```sql
user_id INTEGER NOT NULL DEFAULT 1,
```

In the per-user DB, `user_id` always defaults to `1` (meaningless — there's only one user per DB). This is harmless in per-user mode but is inconsistent: if the per-user DB is ever queried via the legacy code path, records could appear to belong to user ID `1` rather than the actual user.

**Fix direction:** In per-user DB mode, `user_id` column is redundant. The `SaveKnowledgeDocument` call always passes the real `userID`, so the default is just a safety net. No action needed, but worth documenting.

---

### ISSUE-6 — Duplicate workspace initialization code

**Files:**
- `pkg/config/workspace_init.go` → `EnsureUserWorkspace(userUUID)`
- `pkg/storage/user_storage.go` → `UserStorageManager.EnsureUserDirectory(userUUID)`

Both functions create identical directory trees under `~/.MakoClaw/users/<uuid>/workspace/`. They diverged and may fall out of sync if new workspace subdirectories are added to one but not the other.

**Fix direction:** Consolidate into a single canonical function (likely in `pkg/config`), called from both places.

---

### ISSUE-7 — CronService fallback to shared instance

**File:** `pkg/web/server.go` → `getCronServiceForRequest()`

When a per-user `CronService` cannot be resolved, the server falls back to the shared `s.cronService`. The shared service's `jobs.json` is NOT scoped to a user's workspace — it lives at a global path.

Jobs in the shared cron service DO have a `UserID` field and all list/remove/update operations filter by it — so logical isolation exists. But the jobs are stored in a single flat file accessible to the process without per-user filesystem isolation.

**Fix direction:** If the user is authenticated, fail hard instead of falling back to shared cron. The fallback hides misconfiguration.

---

### ISSUE-8 — `SessionManager.SetStorage()` must be called per-user

**File:** `pkg/session/manager.go`

`SessionManager` persists sessions to a `storage` directory set via `SetStorage()`. For per-user isolation, this must point to `~/.MakoClaw/users/<uuid>/workspace/sessions/`. If the manager is shared (not per-user), sessions from different users land in the same directory.

**Verification needed:** Confirm that `MultiUserChannelManager` creates one `SessionManager` per user with the correct storage path.

---

### ISSUE-9 — Multi-Agent User Context Propagation

**Finding:** The multi-agent system (specialists, swarms, orchestrators) may not consistently propagate user identity, risking cross-user workspace access and incomplete audit trails.

**File:** `pkg/agent/manager.go:InitializeOrchestrator`
```go
specialist := agent.NewAgentLoop(globalCfg)
// No SetUserForAgent() call — specialist lacks user context!
```

**Impact:** HIGH — specialists created by the manager operate without user identity, potentially reading from wrong workspace or misattributing actions in audit logs.

**Fix direction:** Add `SetUserForAgent(userUUID, userID)` call for all specialists in `InitializeOrchestrator`.

---

**File:** `pkg/agent/orchestrator.go` — `TeamContext` struct
```go
type TeamContext struct {
    ParentAgentID string
    TaskID        string
    TaskType      string
    // No UserUUID or UserID fields!
}
```

**Impact:** MEDIUM — when delegating to specialists, the team context lacks user identity for workspace isolation.

**Fix direction:** Add `UserUUID string` and `UserID int64` to `TeamContext`. Update `TaskDecompositionTool` to populate these from parent agent.

---

**File:** `pkg/agent/swarm.go` — `SwarmRunner`
```go
func (sr *SwarmRunner) Execute(ctx context.Context, swarm *SwarmConfig) error {
    for _, specialistCfg := range swarm.Specialists {
        specialist := agent.NewAgentLoop(globalCfg)
        // Missing: SetUserForAgent(userUUID, userID)
    }
}
```

**Impact:** MEDIUM — swarm executions may bypass per-user workspace isolation if user context is not propagated to child specialists.

**Fix direction:** Ensure `SwarmRunner` receives and passes user context to all spawned specialists.

---

**File:** `pkg/agent/spawner.go` — `SpecialistSpawnTask`
```go
type SpecialistSpawnTask struct {
    TaskID      string
    AgentID     string
    Specialist  string
    // No UserID or UserUUID for audit trail!
}
```

**Impact:** LOW — audit trail cannot trace which user spawned which specialist during multi-agent workflows.

**Fix direction:** Add `UserID int64` and `UserUUID string` fields to `SpecialistSpawnTask` for audit tracking.

---

## What Is Working Correctly

- ✅ All HTTP handlers use `getUserStorage(r)` to get the per-user DB
- ✅ `handleTasks`, `handleChatSessions`, `handleKnowledge`, `handleCron`, `handleBackup`, `handleWorkflows` all resolve per-user store first
- ✅ `processNextTodoTaskPerUser` correctly opens per-user stores and creates isolated agent loops
- ✅ CronJob model has `UserID` field; all CronService filter methods scope by user
- ✅ Knowledge CRUD (`SaveKnowledgeDocument`, `ListKnowledgeDocuments`, `DeleteKnowledgeDocument`, `SearchKnowledge`) always filters by `userID`
- ✅ Task CRUD uses `isUserDB` flag to skip redundant `user_id` filtering when in per-user DB mode
- ✅ Config system: `LoadConfigForUser` / `SaveConfigForUser` load/save to per-user `config.json`
- ✅ `GetUserConfigTemplate` does NOT copy API keys/passwords to new users
- ✅ Central store only handles auth/marketplace — no user content data
- ✅ Workspace directory structure is consistently bootstrapped per user

---

## Summary of Risks (Severity)

| # | Issue | Severity | Mode Affected |
|---|-------|----------|---------------|
| 1 | Metrics singleton not user-scoped; any user sees aggregate | Medium | Both modes |
| 2 | `GetTaskLogs` no ownership check in legacy mode | Low | Legacy only |
| 3 | Legacy task worker uses cross-user shared agent loop | High | Legacy only |
| 4 | `workflows` table has no `user_id` in legacy shared DB | High | Legacy only |
| 5 | `knowledge_documents.user_id DEFAULT 1` in per-user DB | Trivial | Per-user mode |
| 6 | Duplicate workspace init code (drift risk) | Low | Both modes |
| 7 | CronService falls back to shared instance silently | Medium | Both modes |
| 8 | SessionManager storage path must be set per-user | Medium | Per-user mode |
| 9 | Multi-agent user context not propagated (specialists, swarms) | HIGH | Both modes |

---

## Recommended Next Steps

1. **Workflows user isolation** — Add `user_id` column to `workflows` table and scope all CRUD methods (mirrors the `tasks` pattern). This is the highest-risk gap in the data model itself.
2. **Multi-agent user context propagation** — Ensure all specialists, swarms, and orchestrators carry user identity via `SetUserForAgent(userUUID, userID)` calls and `TeamContext` fields.
3. **Metrics per-user** — Replace `observability.Global()` with per-user metrics instances, or add `user_id` attribution + filtering to the metrics endpoint.
4. **CronService fallback hardening** — Remove or gate the fallback to shared cron; log an error instead of silently degrading.
5. **SessionManager audit** — Confirm `MultiUserChannelManager` sets per-user storage paths on session managers.
6. **Legacy task worker deprecation** — Add a log warning that `processNextTodoTaskLegacy` is unsafe in multi-user deployments; document that it is only for single-user mode.
7. **Workspace init consolidation** — Merge `config.EnsureUserWorkspace` and `UserStorageManager.EnsureUserDirectory` into one canonical function.
