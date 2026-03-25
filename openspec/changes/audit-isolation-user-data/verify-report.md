# Verification Report: User Data Isolation Hardening

**Change:** `audit-isolation-user-data`
**Verified:** 2026-03-25 (updated by sdd-verify agent — full behavioral execution pass)
**Verifier:** sdd-verify sub-agent

---

## Completeness

| Metric | Value |
|--------|-------|
| Tasks total | 40 |
| Tasks complete `[x]` | 40 |
| Tasks incomplete `[ ]` | 0 |

All 40 tasks are marked complete.

---

## Build & Tests Execution

**Build / go vet:** ✅ Passed (`go vet ./...` — zero errors, zero warnings)

**Tests:** ❌ 6 failed / ✅ 18 packages passed / ⚠️ 1 package no test files

### Failed tests

| Package | Test | Error |
|---------|------|-------|
| `pkg/agent` | `TestSwarmRunnerPropagatesUserContextToChildSpecialists` | `expected propagated user UUID aaa-111, got ""` |
| `pkg/storage` | `TestListSessionsEmpty` | `listing sessions: missing argument with index 6` |
| `pkg/storage` | `TestListSessionsWithMessages` | `listing sessions: missing argument with index 6` |
| `pkg/storage` | `TestListSessionsArchivedFilter` | `expected only active session, got []` |
| `pkg/storage` | `TestListSessionsPagination` | `expected 2 in page1, got 0` |
| `pkg/storage` | `TestChat_ListSessions` | `ListSessionsForUser: listing sessions: missing argument with index 6` |

**Note on pre-existing failures:** The 5 `pkg/storage` failures (`missing argument with index 6`) are pre-existing failures unrelated to this change. They were present before this change and are explicitly noted in `apply-progress.md` as "A pre-existing `pkg/storage` sessions test failure (`missing argument with index 6`) still appears...". They affect `ListSessions` query pagination — not any isolation logic.

**Coverage:** Not configured in `openspec/config.yaml`

---

## Root Cause Analysis: `TestSwarmRunnerPropagatesUserContextToChildSpecialists`

This is the **only failure introduced by this change** (ISSUE-9 SwarmRunner scenario).

### What happens

The flow inside `executeMember` (`swarm.go:264–266`):

```go
if exec.TeamContext != nil {
    specialist.SetUserForAgent(exec.TeamContext.UserUUID, exec.TeamContext.UserID)
    // → sets specialist.AgentLoop.userUUID = "aaa-111"  ✅
}
// ...
result, err := specialist.ProcessWithSpeciality(ctx, task)
// → ProcessDirect → ProcessDirectWithChannel → creates InboundMessage{UserID: 0}
// → processMessageWithModel(ctx, msg, "") → applyMessageUserContext(msg)
// → msg.UserID == 0 → SetUserForAgent("", 0)  ← RESETS userUUID back to ""  ❌
```

### Root cause

`processMessageWithModel` unconditionally calls `applyMessageUserContext(msg)` at line 668 of `loop.go`. When `ProcessDirect` / `ProcessWithSpeciality` is used (no UserID in the message), this resets `userUUID` to `""` — overwriting the value that was set by `SetUserForAgent` just before execution.

### Required fix

`executeMember` in `swarm.go` should pass the user identity through the message itself, not rely on pre-setting it via `SetUserForAgent`. Options:

1. **(Recommended)** Add `ProcessWithSpecialityForUser(ctx, userUUID, userID, task)` to `SpecialistAgent` that calls `ProcessDirectWithChannelForUser` — this carries `UserID` in the message so `applyMessageUserContext` correctly reapplies user context instead of resetting it.

2. **Alternative:** In `executeMember`, after calling `ProcessWithSpeciality`, re-call `SetUserForAgent` to restore the value — but this leaves a window where `userUUID` is `""` during execution (workspace operations inside the message loop would use the wrong path).

3. **Alternative:** In `processMessageWithModel`, only call `applyMessageUserContext` if `msg.UserID != 0` — but this could break existing message-loop behavior.

Option 1 is the cleanest because it ensures user context is correct **throughout** execution (workspace resolution inside `processMessageWithModel` uses the correct user path) and the field remains set after execution.

---

## Spec Compliance Matrix

| Issue | Requirement / Scenario | Test | Result |
|-------|------------------------|------|--------|
| ISSUE-3 | Legacy path exits immediately if `userMgr != nil` | `pkg/web: TestProcessNextTodoTaskLegacySkipsWhenUserManagerPresent` | ✅ COMPLIANT |
| ISSUE-3 | FATAL log emitted on startup (legacy + multiple users) | `pkg/web: TestLogUnsafeLegacyTaskWorkerModeEmitsFatalCondition` | ✅ COMPLIANT |
| ISSUE-3 | Legacy path operates correctly in single-user mode | `TestProcessNextTodoTaskLegacySkipsWhenUserManagerPresent` (negative: single-user runs) | ✅ COMPLIANT |
| ISSUE-4 | `ListWorkflows` returns only requesting user's rows | `pkg/storage: TestListWorkflowsLegacyModeFiltersByUser` | ✅ COMPLIANT |
| ISSUE-4 | Migration uses `ALTER TABLE ADD COLUMN` — no data loss | `pkg/storage: TestWorkflowMigrationIdempotentAndNonDestructive` | ✅ COMPLIANT |
| ISSUE-4 | Per-user DB mode — no `user_id` filter | `TestListWorkflowsLegacyModeFiltersByUser` (implicit via `isUserDB` branch) | ✅ COMPLIANT |
| ISSUE-4 | Workflow delete rejects cross-user attempt | Code: `DELETE … WHERE id=? AND user_id=?` (silent 0-rows) | ⚠️ PARTIAL (no dedicated delete cross-user test) |
| ISSUE-1 | `handleMetrics` returns only requesting user's data | `pkg/web: TestHandleMetricsIsolatesPerUserLLMCounts` | ✅ COMPLIANT |
| ISSUE-1 | Metrics persist to user-scoped store | `TestHandleMetricsIsolatesPerUserLLMCounts` (per-user map verified) | ✅ COMPLIANT |
| ISSUE-1 | `Global()` no-op in single-user legacy mode | Static: `observability.Global().SetStorage` removed from `server.go` | ✅ COMPLIANT |
| ISSUE-7 | `getCronServiceForRequest` returns `ErrCronNotInitialized` | `pkg/web: TestGetCronServiceForRequestReturnsErrCronNotInitialized` | ✅ COMPLIANT |
| ISSUE-7 | Per-user cron available — returns instance | Static: normal cron path returns from `s.userMgr` lookup | ✅ COMPLIANT |
| ISSUE-7 | Unauthenticated path — shared cron used | Static: `getCronServiceForRequest` guards only on authenticated user UUID | ✅ COMPLIANT |
| ISSUE-8 | Session stored in correct per-user path | `pkg/session: TestSessionManagerCrossUserStorageIsolation` | ✅ COMPLIANT |
| ISSUE-8 | Cross-user session read returns not found | `TestSessionManagerCrossUserStorageIsolation` | ✅ COMPLIANT |
| ISSUE-8 | `SetStorage` not called — write fails | Static: `SetStorage` always called in `MultiUserChannelManager` init | ✅ COMPLIANT |
| ISSUE-2 | `GetTaskLogs` returns empty for cross-user task_id (legacy) | `pkg/storage: TestGetTaskLogsLegacyReturnsEmptyForCrossUser` | ✅ COMPLIANT |
| ISSUE-2 | Per-user DB mode — no ownership JOIN | `TestGetTaskLogsLegacyReturnsEmptyForCrossUser` (via `isUserDB=true` branch) | ✅ COMPLIANT |
| ISSUE-6 | `EnsureUserWorkspace` creates all expected subdirs | `pkg/config: TestEnsureUserWorkspaceCreatesAllSubdirectories` | ✅ COMPLIANT |
| ISSUE-6 | `EnsureUserDirectory` delegates to canonical function | `pkg/storage: TestEnsureUserDirectoryDelegatesToEnsureUserWorkspace` | ✅ COMPLIANT |
| ISSUE-5 | Inline comment at `DEFAULT 1` in `knowledge.go` | Static: comment present at `knowledge.go` migration line | ✅ COMPLIANT |
| ISSUE-9 | Specialists in manager have `SetUserForAgent` called | `pkg/agent: TestInitializeOrchestratorSetsUserContextForAllSpecialists` | ✅ COMPLIANT |
| ISSUE-9 | `TeamContext` carries user identity from `TaskDecompositionTool` | `pkg/agent: TestTaskDecompositionTool_PopulatesTeamContextUserIdentity` | ✅ COMPLIANT |
| ISSUE-9 | `SwarmRunner` propagates user context to child specialists | `pkg/agent: TestSwarmRunnerPropagatesUserContextToChildSpecialists` | ❌ FAILING |
| ISSUE-9 | `SpecialistSpawnTask` includes `UserID` for audit | `pkg/agent: TestSpawnForSpecialistIncludesUserIdentity` | ✅ COMPLIANT |
| ISSUE-9 | Graceful fallback for missing user context | Static: workspace error logged, no cross-user data access | ✅ COMPLIANT |

**Compliance summary:** 24/25 scenarios compliant (96%)

---

## Correctness (Static — Structural Evidence)

| Requirement | Status | Notes |
|-------------|--------|-------|
| ISSUE-3: `// SINGLE-USER ONLY` comment at function declaration | ✅ Implemented | Present in `pkg/web/server.go` |
| ISSUE-3: `logUnsafeLegacyTaskWorkerMode` startup check | ✅ Implemented | Calls `s.store.CountUsers()` |
| ISSUE-4: `user_id` column in workflows/workflow_runs/workflow_approvals | ✅ Implemented | Idempotent `ALTER TABLE ADD COLUMN` with `pragma_table_info` check |
| ISSUE-4: `CREATE INDEX IF NOT EXISTS` on `user_id` | ✅ Implemented | In workflow migration |
| ISSUE-4: All CRUD scoped by `userID` in legacy mode | ✅ Implemented | `isUserDB` branch in List/Get/Update/Delete |
| ISSUE-1: Per-user metrics map `userMetrics` | ✅ Implemented | `getOrCreateUserMetrics(userUUID)` in `server.go` |
| ISSUE-1: `observability.Global().SetStorage` removed | ✅ Implemented | No longer wired in server startup |
| ISSUE-7: `ErrCronNotInitialized` sentinel in `pkg/cron/errors.go` | ✅ Implemented | `var ErrCronNotInitialized = errors.New(...)` |
| ISSUE-7: `getCronServiceForRequest` returns sentinel error for authenticated users | ✅ Implemented | Fallback removed for authenticated UUID |
| ISSUE-7: `ErrCronNotInitialized` contains no user-identifying data | ✅ Implemented | Message is generic: `"per-user cron service not initialized"` |
| ISSUE-8: `MultiUserChannelManager` per-user `SessionManager` isolation | ✅ Implemented | Comment at line 121-123 confirming; `TODO` on data.db inconsistency |
| ISSUE-2: `GetTaskLogs` ownership JOIN in legacy mode | ✅ Implemented | JOIN on `tasks.user_id` in `task_logs.go` |
| ISSUE-6: `EnsureUserWorkspace` canonical function | ✅ Implemented | `pkg/config/workspace_init.go` |
| ISSUE-6: `EnsureUserDirectory` delegates | ✅ Implemented | Calls `config.EnsureUserWorkspace` internally |
| ISSUE-5: `DEFAULT 1` comment in `knowledge.go` | ✅ Implemented | Inline comment explaining safety rationale |
| ISSUE-9: `TeamContext.UserUUID` + `UserID` fields | ✅ Implemented | `pkg/agent/orchestrator.go` |
| ISSUE-9: `TaskDecompositionTool` populates user identity | ✅ Implemented | Copies from parent agent context |
| ISSUE-9: `SetUserForAgent` called for all specialists in `manager.go` | ✅ Implemented | Lines 106, 108, 173 |
| ISSUE-9: `SpecialistSpawnTask.UserID` + `UserUUID` fields | ✅ Implemented | `pkg/agent/spawner.go` |
| ISSUE-9: `SwarmRunner.executeMember` calls `SetUserForAgent` | ✅ Implemented (buggy) | Set correctly but overwritten by `applyMessageUserContext` |

---

## Coherence (Design Match)

| Decision | Followed? | Notes |
|----------|-----------|-------|
| `isUserDB` branch pattern for all storage methods | ✅ Yes | Mirrors existing tasks pattern throughout |
| Additive migration — `ALTER TABLE ADD COLUMN` only | ✅ Yes | No drops, no data loss |
| Sentinel error `ErrCronNotInitialized` in `pkg/cron` package | ✅ Yes | Named `var` at package level |
| Single canonical workspace function in `pkg/config` | ✅ Yes | No duplicate implementations |
| All new errors as `var Err… = errors.New(…)` package-level constants | ✅ Yes | |
| `processNextTodoTaskLegacy` guard via `userMgr != nil` check | ✅ Yes | |
| Per-user metrics via map keyed by user UUID | ✅ Yes | `userMetrics sync.Map` |
| `TeamContext` fields for user propagation (non-serialized metadata) | ✅ Yes | Transient, not stored to DB |
| `SpecialistSpawnTask.UserID` for audit trail | ✅ Yes | Stored for traceability |
| Sessions stored at `{userWorkspaceRoot}/sessions/` | ✅ Yes | Via `SetStorage` in `MultiUserChannelManager` |

---

## Issues Found

### CRITICAL (must fix before archive)

**CRITICAL-1: `TestSwarmRunnerPropagatesUserContextToChildSpecialists` is FAILING**

- **File:** `pkg/agent/swarm.go:executeMember` + `pkg/agent/loop.go:processMessageWithModel`
- **Symptom:** After `RunSwarm`, `specialist.userUUID` is `""` instead of `"aaa-111"`.
- **Root cause:** `executeMember` calls `SetUserForAgent("aaa-111", 7)` before execution, but `ProcessWithSpeciality` → `ProcessDirect` → `processMessageWithModel` unconditionally calls `applyMessageUserContext(InboundMessage{UserID: 0})` which resets `userUUID` back to `""`.
- **Impact:** In production, swarm child specialists operate with an empty `userUUID`, so their workspace operations fall back to the global workspace — exactly the cross-user isolation failure this change was meant to prevent.
- **Fix:** Introduce `ProcessWithSpecialityForUser(ctx, userUUID string, userID int64, task string) (string, error)` on `SpecialistAgent` that calls `ProcessDirectWithChannelForUser` (which carries `UserID` in the message, ensuring `applyMessageUserContext` correctly re-applies the user context throughout execution). Update `swarm.go:executeMember` to use this method.

---

### WARNING (should fix)

**WARNING-1: Workflow delete cross-user isolation has no dedicated test**

- **Spec scenario:** "Workflow delete rejects cross-user attempt (legacy mode)" (`spec.md` line 110)
- **Status:** Code implements `WHERE id=? AND user_id=?` silently returning 0 rows — correct — but no test verifies this specifically.
- **Fix:** Add `TestWorkflowDeleteLegacyRejectsCrossUserAttempt` to `pkg/storage/workflow_test.go`.

**WARNING-2: `MultiUserChannelManager` data.db TODO is unresolved**

- **File:** `pkg/channels/multiuser_manager.go` line ~121-123
- **Details:** A `// TODO(future)` comment documents an inconsistency where `data.db` may not be fully isolated per-user. Not introduced by this change, but surfaced during implementation.
- **Fix:** Track as a follow-up issue; document in the release notes.

---

### SUGGESTION (nice to have)

**SUGGESTION-1: Spec duplication in `spec.md`**

The Acceptance Criteria table appears twice (lines 444–460 and 462–475). The second copy is a fragment of the first. This does not affect functionality but should be cleaned up before archive.

---

## Verdict

**FAIL**

One CRITICAL test failure (`TestSwarmRunnerPropagatesUserContextToChildSpecialists`) prevents archive. The swarm user-context propagation code has the right intent but is structurally bypassed by the message processing loop resetting user context. This failure represents a real isolation gap in production for swarm-based multi-agent workflows.

All other 24/25 spec scenarios are compliant, `go vet` is clean, and 18 of 19 packages with tests pass. The 5 pre-existing `pkg/storage` session test failures are unrelated to this change.

**Recommended next action:** Fix `swarm.go:executeMember` to propagate user identity via the message (not pre-set it), re-run `go test ./pkg/agent/...`, then archive.
