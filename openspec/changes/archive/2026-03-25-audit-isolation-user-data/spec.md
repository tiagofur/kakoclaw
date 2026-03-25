# Specification: User Data Isolation Hardening

**Change:** `audit-isolation-user-data`
**Status:** Complete
**Created:** 2026-03-25
**Based on:** [proposal.md](proposal.md), [explore.md](explore.md)

---

## Overview

This is a **full spec** (new domain — no prior isolation-hardening spec exists).

Covers 8 isolation gaps found in the audit, organized into 3 priority phases. The system operates in two modes throughout:

- **Per-user DB mode**: each user has `~/.MakoClaw/users/<uuid>/user.db` — isolation by file
- **Legacy shared-DB mode**: single `s.store` shared across all users — isolation by `user_id` column

All requirements apply to **both modes** unless explicitly scoped.

---

## Phase 1 — HIGH Priority

### ISSUE-3: Legacy Task Worker Isolation

#### Requirement: Legacy worker MUST NOT run in multi-user deployments

The `processNextTodoTaskLegacy` function MUST NOT process tasks when more than one user account exists. It MUST only operate in single-user deployments where `s.userMgr == nil`.

| Rule | Strength |
|------|----------|
| If `s.userMgr != nil`, legacy path MUST return immediately with error log | MUST |
| On startup, if `s.userMgr == nil` AND user count > 1, MUST emit a FATAL-level structured log | MUST |
| Source file MUST contain a `// SINGLE-USER ONLY` comment block at function declaration | MUST |
| Log entry MUST use component-based format: `logger.ErrorCF("task-worker", ...)` | MUST |

#### Scenario: Legacy path called in multi-user deployment

- GIVEN `s.userMgr != nil` (per-user mode is active)
- WHEN `processNextTodoTaskLegacy` is invoked by the task scheduler
- THEN the function returns immediately without processing any task
- AND a structured error log is emitted: `"processNextTodoTaskLegacy called in multi-user mode — skipping"`

#### Scenario: Server starts in legacy mode with multiple users

- GIVEN `s.userMgr == nil` (legacy mode)
- AND the `users` table in `s.store` contains more than one row
- WHEN the server completes startup
- THEN a FATAL-level log is emitted warning of unsafe multi-user legacy operation
- AND the server SHOULD continue (not crash) to allow operator recovery

#### Scenario: Legacy path operates correctly in single-user mode

- GIVEN `s.userMgr == nil`
- AND the `users` table contains exactly one row
- WHEN `processNextTodoTaskLegacy` is invoked
- THEN it processes the next pending task normally
- AND no warning log is emitted

---

### ISSUE-4: Workflows Table User Isolation

#### Requirement: Workflows MUST be user-scoped in all storage modes

The `workflows`, `workflow_runs`, and `workflow_approvals` tables in the legacy shared DB MUST include a `user_id` column. All CRUD operations MUST scope queries by `user_id`.

| Rule | Strength |
|------|----------|
| `workflows` table MUST have `user_id INTEGER NOT NULL DEFAULT 0` (additive migration) | MUST |
| `workflow_runs` table MUST have `user_id INTEGER NOT NULL DEFAULT 0` | MUST |
| `workflow_approvals` table MUST have `user_id INTEGER NOT NULL DEFAULT 0` | MUST |
| `ListWorkflows()` MUST accept `userID int64` and apply `WHERE user_id = ?` in legacy mode | MUST |
| `GetWorkflow()`, `UpdateWorkflow()`, `DeleteWorkflow()` MUST verify `user_id` ownership | MUST |
| In per-user DB mode (`isUserDB = true`), `user_id` filter MUST be skipped (mirrors tasks pattern) | MUST |
| `handleWorkflows` MUST pass `userID` from `getUserStorage(r)` context to all storage calls | MUST |
| Migration MUST use `ALTER TABLE … ADD COLUMN` — no table drop or data loss | MUST |

#### Data Requirements

| Column | Table | Type | Default | Notes |
|--------|-------|------|---------|-------|
| `user_id` | `workflows` | `INTEGER NOT NULL` | `0` | Backfill from context before first multi-user query |
| `user_id` | `workflow_runs` | `INTEGER NOT NULL` | `0` | Same |
| `user_id` | `workflow_approvals` | `INTEGER NOT NULL` | `0` | Same |

#### Scenario: User A cannot read User B's workflows (legacy mode)

- GIVEN two users exist in the legacy shared DB: User A (id=1) and User B (id=2)
- AND User B has workflows with `user_id = 2`
- WHEN User A calls `GET /api/workflows`
- THEN the response contains only User A's workflows (zero rows if none)
- AND User B's workflows are NOT included

#### Scenario: Workflow creation assigns correct owner (legacy mode)

- GIVEN User A (id=1) is authenticated
- WHEN User A calls `POST /api/workflows` with a valid workflow payload
- THEN the created workflow row has `user_id = 1`
- AND `ListWorkflows(userID=2)` does NOT return this workflow

#### Scenario: Per-user DB mode — no user_id filter applied

- GIVEN the system is running in per-user DB mode (`isUserDB = true`)
- WHEN `ListWorkflows` is called
- THEN the query executes WITHOUT a `WHERE user_id = ?` clause
- AND all workflows in the file-isolated DB are returned

#### Scenario: Workflow delete rejects cross-user attempt (legacy mode)

- GIVEN User A owns workflow id=42 (`user_id = 1`)
- WHEN a request attempts to delete workflow id=42 with `userID = 2`
- THEN the delete returns 0 affected rows
- AND no error is raised (silent ownership guard)

---

## Phase 2 — HIGH Priority (Added from Multi-Agent Exploration)

### ISSUE-9: Multi-Agent User Context Propagation

#### Requirement: All agent components MUST propagate user identity

Specialists, swarms, and orchestrators created during multi-agent workflows MUST carry user identity to ensure per-user workspace isolation and complete audit trails. The agent system MUST NOT create components that operate without user context.

| Rule | Strength |
|------|----------|
| `manager.go:InitializeOrchestrator` MUST call `SetUserForAgent(userUUID, userID)` for all specialists | MUST |
| `TeamContext` struct MUST include `userUUID` and `userID` fields | MUST |
| `TaskDecompositionTool` MUST create `TeamContext` with user identity from parent agent | MUST |
| `SwarmRunner` MUST propagate user context to all spawned specialists | MUST |
| `SpecialistSpawnTask` struct MUST include `UserID int64` field for audit trail | MUST |
| Agent loop workspace resolution MUST fail gracefully if user context is missing | SHOULD |

#### Data Requirements

| Field | Struct | Type | Purpose |
|-------|--------|------|---------|
| `userUUID` | `TeamContext` | `string` | User's UUID for workspace path resolution |
| `userID` | `TeamContext` | `int64` | User's numeric ID for database operations |
| `UserID` | `SpecialistSpawnTask` | `int64` | Audit trail for who spawned this specialist |
| `userUUID` | `SpecialistSpawnTask` | `string` | Optional: UUID for workspace operations |

#### Scenario: Specialist created with user context in manager

- GIVEN an orchestrator is initialized in `manager.go:InitializeOrchestrator` with userUUID and userID
- WHEN specialists are created via `NewAgentLoop(globalCfg)`
- THEN `SetUserForAgent(userUUID, userID)` is called for each specialist
- AND the specialist's workspace is scoped to the user's workspace

#### Scenario: TeamContext carries user identity to specialists

- GIVEN a parent agent (userUUID=`aaa-111`, userID=`1`) delegates a task to a specialist
- WHEN `TaskDecompositionTool` creates the specialist task
- THEN the resulting `TeamContext` includes `userUUID=aaa-111` and `userID=1`
- AND the specialist uses these values for workspace and database operations

#### Scenario: SwarmRunner propagates user context

- GIVEN a user (userUUID=`aaa-111`) initiates a swarm execution
- WHEN `SwarmRunner` spawns child specialists for the swarm
- THEN each child specialist receives the parent's user context
- AND all swarm operations use the user's per-user workspace

#### Scenario: SpecialistSpawnTask includes userID for audit

- GIVEN a specialist task is spawned during a multi-agent workflow
- WHEN the `SpecialistSpawnTask` is created
- THEN the task includes `UserID` field set to the spawning user's ID
- AND audit logs can trace which user spawned which specialist

#### Scenario: Graceful fallback for missing user context

- GIVEN a specialist is created without explicit user context (legacy path)
- WHEN the agent attempts workspace operations
- THEN the system fails gracefully with a structured error log
- AND no cross-user data access occurs

---

## Phase 3 — MEDIUM Priority

### ISSUE-1: Metrics Per-User Isolation

#### Requirement: Metrics MUST NOT expose cross-user aggregate data

Each authenticated user MUST see only their own metrics. `observability.Global()` MUST NOT be wired to the legacy shared store for multi-user deployments.

| Rule | Strength |
|------|----------|
| `handleMetrics` MUST return only the requesting user's metrics | MUST |
| `observability.Global().SetStorage(s.store)` MUST be removed from `server.go` | MUST |
| Per-user metrics MUST be stored in / read from the user's own DB or scoped by `user_id` | MUST |
| `Global()` shim MAY remain as a no-op for backward compatibility in single-user mode | MAY |
| Metrics instances MUST be injected via the same context path as per-user storage | SHOULD |

#### Scenario: User A cannot see User B's metrics

- GIVEN User A and User B have each made LLM calls recorded in metrics
- WHEN User A calls `GET /api/metrics`
- THEN the response contains only User A's LLM call counts and events
- AND User B's metrics are NOT included in any counter or event list

#### Scenario: Metrics persist to user-scoped store

- GIVEN User A is authenticated
- WHEN an LLM call completes for User A's session
- THEN the metrics event is written to User A's scoped store (not the global/shared store)

#### Scenario: Global() no-op in single-user legacy mode

- GIVEN `s.userMgr == nil` (single-user legacy mode)
- WHEN the server starts
- THEN `observability.Global()` operates as a no-op or is wired to the single user's store
- AND no panic or error occurs

---

### ISSUE-7: CronService Hard Failure on Missing Per-User Instance

#### Requirement: CronService MUST fail explicitly for authenticated users

`getCronServiceForRequest()` MUST NOT silently fall back to the shared `s.cronService` when a per-user `CronService` cannot be resolved for an authenticated request.

| Rule | Strength |
|------|----------|
| If user is authenticated AND per-user cron cannot be resolved, MUST return `nil, ErrCronNotInitialized` | MUST |
| The shared `s.cronService` fallback MUST be removed for authenticated requests | MUST |
| All HTTP handlers calling `getCronServiceForRequest()` already check error returns — no additional changes needed | SHOULD (verify) |
| `ErrCronNotInitialized` MUST be a named sentinel error in the cron package | MUST |
| Unauthenticated or startup paths (where no user context exists) MAY still use the shared service | MAY |

#### Scenario: Authenticated request — per-user cron unavailable

- GIVEN a user is authenticated (UUID resolved from request context)
- AND the per-user `CronService` has not been initialized for this user
- WHEN `getCronServiceForRequest()` is called
- THEN it returns `nil, ErrCronNotInitialized`
- AND the HTTP handler returns a 500 error to the client
- AND a structured error log is emitted identifying the user UUID

#### Scenario: Authenticated request — per-user cron available

- GIVEN a user is authenticated
- AND the per-user `CronService` is initialized and registered for this user
- WHEN `getCronServiceForRequest()` is called
- THEN it returns the user's `CronService` instance and `nil` error

#### Scenario: Unauthenticated / startup path — shared cron used

- GIVEN no user is authenticated in the request context
- WHEN `getCronServiceForRequest()` is called
- THEN it MAY return the shared `s.cronService`
- AND no error is returned

---

### ISSUE-8: SessionManager Per-User Storage Path

#### Requirement: Each user's SessionManager MUST write to their own workspace

For every authenticated channel, `SessionManager.SetStorage()` MUST be called with the path `~/.MakoClaw/users/<uuid>/workspace/sessions/`. A shared session manager MUST NOT be used across multiple users.

| Rule | Strength |
|------|----------|
| `MultiUserChannelManager` MUST call `SetStorage` with per-user path for every user | MUST |
| Path format MUST be `{userWorkspaceRoot}/sessions/` | MUST |
| Sessions written for User A MUST NOT be readable via User B's session manager | MUST |
| An integration test MUST verify cross-user session isolation | MUST |

#### Scenario: Session stored in correct per-user path

- GIVEN User A (uuid=`aaa-111`) is authenticated on a channel
- WHEN User A's agent writes a session
- THEN the session file is created under `~/.MakoClaw/users/aaa-111/workspace/sessions/`
- AND the file does NOT appear under any other user's workspace

#### Scenario: Cross-user session read returns not found

- GIVEN User A has session id=`sess-99` stored in their workspace
- AND User B's `SessionManager` is initialized with `~/.MakoClaw/users/bbb-222/workspace/sessions/`
- WHEN User B's `SessionManager` attempts to read `sess-99`
- THEN the read returns a not-found error (file does not exist in User B's path)

#### Scenario: SetStorage not called — initialization fails loudly

- GIVEN `MultiUserChannelManager` creates a `SessionManager` for a new user
- AND `SetStorage` is NOT called before the first session write
- THEN the write MUST return an error
- AND MUST NOT silently write to a default or shared path

---

## Phase 3 — LOW / Trivial

### ISSUE-2: Task Logs Ownership Guard (Legacy Mode)

#### Requirement: GetTaskLogs MUST enforce task ownership in legacy shared-DB mode

In legacy mode (`isUserDB = false`), `GetTaskLogs` MUST verify the requested `task_id` belongs to the requesting user before returning logs.

| Rule | Strength |
|------|----------|
| In legacy mode, query MUST `JOIN tasks ON tasks.id = task_logs.task_id AND tasks.user_id = ?` | MUST |
| In per-user DB mode (`isUserDB = true`), the JOIN MUST be skipped (isolation by file is sufficient) | MUST |
| No schema changes are required | — |
| If `task_id` is not owned by the user in legacy mode, return empty result set (not an error) | SHOULD |

#### Scenario: User A cannot read User B's task logs (legacy mode)

- GIVEN User A (id=1) requests logs for task_id belonging to User B (id=2) in legacy shared DB
- WHEN `GetTaskLogs(taskID, userID=1)` is called in legacy mode
- THEN the result is an empty list
- AND no error is raised

#### Scenario: Per-user DB mode — no ownership join applied

- GIVEN the system is in per-user DB mode
- WHEN `GetTaskLogs(taskID)` is called
- THEN the query runs WITHOUT the ownership JOIN
- AND all logs for the task_id in the file-isolated DB are returned

---

### ISSUE-6: Workspace Init Consolidation

#### Requirement: A single canonical function MUST govern workspace creation

`pkg/config/workspace_init.go:EnsureUserWorkspace` MUST be the sole canonical implementation. `UserStorageManager.EnsureUserDirectory` MUST delegate to it.

| Rule | Strength |
|------|----------|
| `EnsureUserWorkspace(userUUID string)` in `pkg/config` is the canonical function | MUST |
| `UserStorageManager.EnsureUserDirectory` MUST call `config.EnsureUserWorkspace` internally | MUST |
| Both functions MUST remain idempotent (`os.MkdirAll` semantics — no error on existing dirs) | MUST |
| A unit test MUST assert all expected subdirectories exist after one call | MUST |
| The canonical function MUST create all subdirs listed in the Architecture Overview of explore.md | MUST |

#### Expected Subdirectories (canonical list)

```
~/.MakoClaw/users/<uuid>/
  user.db                  (created by storage layer, not this function)
  config.json              (created by config layer, not this function)
  workspace/
    AGENTS.md
    SOUL.md
    USER.md
    IDENTITY.md
    memory/
      MEMORY.md
    sessions/
    skills/
    cron/
      jobs.json
    tasks/
    temp/
```

#### Scenario: First-time workspace creation

- GIVEN a new user UUID that has no existing workspace
- WHEN `EnsureUserWorkspace(uuid)` is called
- THEN all expected subdirectories are created
- AND the function returns without error

#### Scenario: Idempotent re-initialization

- GIVEN a workspace already exists for the user
- WHEN `EnsureUserWorkspace(uuid)` is called again
- THEN no error is returned
- AND no existing files are modified or deleted

#### Scenario: Delegated call from UserStorageManager

- GIVEN `UserStorageManager.EnsureUserDirectory(uuid)` is called
- THEN it calls `config.EnsureUserWorkspace(uuid)` internally
- AND the result is identical to calling the canonical function directly

---

### ISSUE-5: knowledge_documents.user_id DEFAULT 1

#### Requirement: Document DEFAULT 1 as intentional no-op (documentation only)

No code change is required. The `DEFAULT 1` on `knowledge_documents.user_id` is harmless in per-user DB mode (one user per file). The column is always populated with the real `userID` by `SaveKnowledgeDocument`.

| Rule | Strength |
|------|----------|
| An inline comment MUST be added to `pkg/storage/knowledge.go` at the migration line | MUST |
| Comment MUST explain: (1) why `DEFAULT 1` exists, (2) that it is safe in per-user mode, (3) that `userID` is always passed explicitly at write time | MUST |

#### Scenario: Comment is present and accurate

- GIVEN a developer reads `pkg/storage/knowledge.go`
- WHEN they encounter `user_id INTEGER NOT NULL DEFAULT 1`
- THEN an inline comment immediately adjacent explains the safety rationale

---

## Non-Functional Requirements

### Multi-Agent Context Propagation

| Requirement | Strength |
|-------------|----------|
| All agent components MUST have user identity set before execution | MUST |
| User context MUST be immutable after initialization | MUST |
| Audit trail MUST capture which user spawned each specialist | MUST |
| Missing user context MUST result in structured error, not silent fallback | MUST |

### Security

### Security

| Requirement | Strength |
|-------------|----------|
| No authenticated user MUST be able to read another user's workflows, metrics, sessions, or task logs | MUST |
| All cross-user access attempts in legacy mode MUST silently return empty results (no information disclosure via error messages) | MUST |
| `ErrCronNotInitialized` error MUST NOT include user-identifying data in its message string | MUST |
| Legacy path guard log MUST use component logging and MUST NOT expose other users' data | MUST |

### Performance

| Requirement | Strength |
|-------------|----------|
| Workflow `user_id` column MUST be indexed (`CREATE INDEX IF NOT EXISTS`) to avoid full-table scans | MUST |
| The ownership JOIN in `GetTaskLogs` legacy mode MUST use the existing `tasks` primary key index | SHOULD |
| Per-user metrics instances MUST NOT introduce per-request DB connections — reuse the established per-user store handle | MUST |

### Maintainability

| Requirement | Strength |
|-------------|----------|
| `processNextTodoTaskLegacy` MUST have a `// DEPRECATED` or `// SINGLE-USER ONLY` header comment | MUST |
| The workspace canonical function MUST be the single source of truth — no other function may create workspace subdirectories independently | MUST |
| All new sentinel errors MUST be defined as package-level `var Err… = errors.New(…)` constants | MUST |

---

## Acceptance Criteria Summary

| Issue | Criterion | How to Verify |
|-------|-----------|---------------|
| ISSUE-3 | Legacy path exits immediately if `userMgr != nil` | Unit test: call with mocked `userMgr != nil`, assert no task processed |
| ISSUE-3 | FATAL log emitted on startup with legacy mode + multiple users | Unit test: mock user count > 1, assert log output |
| ISSUE-4 | `ListWorkflows` returns only requesting user's rows in legacy mode | Integration test: insert rows for two users, query as user A |
| ISSUE-4 | Migration uses `ALTER TABLE ADD COLUMN` — no data loss | Test: pre-populate DB, run migration, assert existing rows survive |
| ISSUE-1 | `handleMetrics` returns only requesting user's data | Integration test: two users generate metrics, each sees only own data |
| ISSUE-7 | `getCronServiceForRequest` returns `ErrCronNotInitialized` for authenticated user with no cron | Unit test: authenticated context, no registered cron, assert error |
| ISSUE-8 | Sessions for User A not readable by User B's SessionManager | Integration test: write session as A, read via B's manager, assert not-found |
| ISSUE-2 | `GetTaskLogs` in legacy mode returns empty for cross-user task_id | Unit test: mock legacy mode, request logs for another user's task |
| ISSUE-6 | Single call to `EnsureUserWorkspace` creates all expected subdirs | Unit test: fresh temp dir, call function, assert all paths exist |
| ISSUE-6 | `EnsureUserDirectory` delegates to canonical function | Unit test: mock/spy canonical function, verify call forwarded |
| ISSUE-5 | Inline comment present at `DEFAULT 1` in knowledge.go | Code review / grep |
| ISSUE-9 | Specialists in manager have `SetUserForAgent()` called | Unit test: mock `InitializeOrchestrator`, assert `SetUserForAgent` invoked |
| ISSUE-9 | `TeamContext` includes user identity for specialists | Unit test: create specialist via `TaskDecompositionTool`, assert user fields populated |
| ISSUE-9 | `SwarmRunner` propagates user context to children | Integration test: spawn swarm, verify children have parent's user context |
| ISSUE-9 | `SpecialistSpawnTask` includes `UserID` | Unit test: inspect `SpecialistSpawnTask` struct, assert `UserID` field present |

---

## Related Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Exploration | `openspec/changes/audit-isolation-user-data/explore.md` | Complete |
| Proposal | `openspec/changes/audit-isolation-user-data/proposal.md` | Complete |
| Specification | `openspec/changes/audit-isolation-user-data/spec.md` | This document |
| Design | `openspec/changes/audit-isolation-user-data/design.md` | Complete |
| Tasks | `openspec/changes/audit-isolation-user-data/tasks.md` | Complete |
| Verify | `openspec/changes/audit-isolation-user-data/verify-report.md` | Archived |
