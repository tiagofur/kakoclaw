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
