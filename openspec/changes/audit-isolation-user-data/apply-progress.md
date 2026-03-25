# Apply Progress: User Data Isolation Hardening

## Batch A

- Completed **1.1** by locating `CountUsers()` in `pkg/storage/sqlite.go` so the legacy startup guard uses a shared storage helper in the SQLite layer.
- Completed **1.2** by adding a `// SINGLE-USER ONLY` contract block and an early-return guard to `processNextTodoTaskLegacy`.
- Completed **1.3** by adding startup detection for unsafe legacy mode when more than one user exists in the shared database.

## Notes

- The logger's `FatalCF` exits the process immediately. To preserve the spec requirement that the server continue running for operator recovery, the startup guard emits a clearly marked `FATAL:` condition through `logger.ErrorCF` instead of terminating the process.

## Batch B

- Completed **1.4** by adding a legacy-worker guard test that verifies `processNextTodoTaskLegacy` leaves todo tasks untouched and emits the multi-user skip log when `userMgr` is present.
- Completed **1.5** by adding a startup guard test that captures standard logger output and verifies the legacy multi-user warning includes the `FATAL:` condition marker.
- Completed **1.6** by extending `migrateWorkflows()` with idempotent `user_id` migrations and indexes for `workflows`, `workflow_runs`, and `workflow_approvals`, while skipping those additive columns for per-user DBs.
- Completed **1.7** by updating workflow CRUD signatures to accept `userID int64` and applying `isUserDB` branching so legacy shared DB queries enforce ownership while per-user DBs keep file-based isolation behavior.
- Completed **1.8** by resolving `userID` in workflow handlers and passing it through workflow list/create/get/update/delete/run paths.

## Additional Notes

- Workflow-related tests in `pkg/storage` and `pkg/workflow` were updated to the new `userID` signatures so the affected packages continue compiling and exercising the CRUD path.
- A pre-existing `pkg/storage` sessions test failure (`missing argument with index 6`) still appears when running the broader package suite; it is unrelated to this batch and was not modified here.

## Batch C

- Completed **2.1** by adding a per-user metrics registry to `pkg/web/server.go`, including double-checked locking and first-access storage binding for each user's metrics collector.
- Completed **2.2** by verifying the old `observability.Global().SetStorage(s.store)` wiring is absent in `server.go`, leaving the global shim unused for request-scoped metrics.
- Completed **2.3** by updating `handleMetrics` to require authenticated per-user storage context, resolve the caller UUID, and return only that user's metrics snapshot.
- Completed **2.4** by adding an injectable metrics collector to `pkg/agent/loop.go`, switching agent/tool/LLM recording to the injected instance, and routing server-created per-user agent loops through a helper that attaches the correct metrics collector.

## Batch C Notes

- `pkg/web/handlers_advanced.go` now uses the shared server helper for per-user `AgentLoop` creation so web-triggered agent actions write into the same metrics instance returned by `/api/v1/metrics`.
- The metrics collector binds to the user's own storage handle when available and falls back to legacy shared storage only in non-per-user mode, matching the current design constraints for backward compatibility.

## Batch D

- Completed **2.5** by adding a two-user metrics integration test in `pkg/web/server_test.go` that records separate LLM activity for two authenticated users and verifies `/api/v1/metrics` returns only the caller's counters, models, and recent events.
- Completed **2.6** by creating `pkg/cron/errors.go` with the exported `cron.ErrCronNotInitialized` sentinel.
- Completed **2.7** by changing `getCronServiceForRequest` to return `error`, removing authenticated per-user cron auto-creation/fallback, and surfacing `ErrCronNotInitialized` whenever multi-user cron wiring is missing.
- Completed **2.8** by updating cron REST handlers to use the new error contract, preserve 401s for unresolved users, and return opaque 500 responses with structured internal logs when per-user cron is not initialized.
- Completed **2.9** by adding a unit test that verifies an authenticated multi-user request without a registered cron service returns `cron.ErrCronNotInitialized`.

## Batch D Notes

- The new cron behavior treats `userMgr != nil` as multi-user mode even if the multi-user channel manager has not been wired yet, which matches the hard-failure requirement and prevents silent fallback to shared cron state.

## Batch E

- Completed **2.12** by propagating `SetUserForAgent(userUUID, userID)` from `AgentManager.InitializeOrchestrator` into the orchestrator and every registered specialist.
- Completed **2.13** and **2.14** by extending `TeamContext` with `user_uuid` / `user_id` and populating those fields when `TaskDecompositionTool` creates or reuses team context.
- Completed **2.15** by carrying user identity into swarm execution context and reapplying it to each swarm member before execution.
- Completed **2.16** and **2.17** by extending `SpecialistSpawnTask` with `UserID` / `UserUUID`, copying those values from team context, and using them when a spawned specialist publishes completion events.
- Completed **2.18** by adding a manager test that verifies orchestrator initialization gives every specialist the caller's user identity.
- Completed **2.19** by adding an orchestrator test that verifies task decomposition preserves parent user identity in `TeamContext`.
- Completed **2.20** by adding a swarm isolation test and verifying the remaining runtime creation path in `AddOrUpdateSpecialist` also reapplies storage plus user context.
- Completed **2.21** by adding a spawner audit-trail test and documenting multi-agent isolation guarantees in the multi-agent setup guide.
- Completed **3.10** by documenting that each per-user agent loop owns its own `SessionManager` storage rooted at `{userWorkspace}/sessions/`.
- Completed **3.11** by adding a cross-user session storage integration test that proves a session written into User A's sessions directory is absent from User B's manager and filesystem path.

## Batch E Notes

- Swarm execution now inherits user identity from the active agent manager context, so REST and WebSocket swarm runs keep the same per-user workspace guarantees as direct specialist delegation.
- `SessionManager` does not expose a typed read error API today, so the new isolation test verifies the effective not-found condition through User B's empty history plus an `os.ErrNotExist` filesystem assertion on User B's sessions path.

## Batch F

- Completed **1.9** by adding a legacy shared-DB workflow isolation test that creates workflows for two users and verifies `ListWorkflows(userA)` returns only User A's rows.
- Completed **1.10** by adding a workflow migration regression test that seeds a pre-migration schema, runs the migration repeatedly, and verifies rows survive while `user_id` columns are added with the expected default.
- Completed **1.11** by documenting the manual legacy workflow backfill SQL and upgrade timing in `release-notes.md`.
- Completed **4.1** and **4.2** by updating `GetTaskLogs` to accept `userID`, applying the legacy ownership join, and passing the caller's `userID` from the task logs HTTP endpoint.
- Completed **4.3** by adding a unit test proving cross-user task log reads return an empty result set in legacy mode.
- Completed **4.4** by consolidating `EnsureUserDirectory` onto `config.EnsureUserWorkspace`.
- Completed **4.5** and **4.6** by adding workspace initialization tests for the canonical directory tree and the delegation path used by `UserStorageManager`.
- Completed **4.7** by documenting why `knowledge_documents.user_id DEFAULT 1` is intentional and harmless in per-user DB mode.
- Completed **4.8** by marking the `multiuser_manager.go` workspace `data.db` inconsistency as an explicit future TODO and capturing the same note in release notes.

## Batch F Notes

- The only runtime caller for task log reads was `pkg/web/server.go` rather than `handlers_advanced.go`; the ownership fix was applied there so the actual HTTP path now enforces the new `GetTaskLogs(taskID, userID)` contract.
- With Batch F complete, all 51 tasks in `audit-isolation-user-data` are now marked done and the change is ready for verification.
