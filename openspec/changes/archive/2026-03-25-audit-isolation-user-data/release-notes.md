# Release Notes: audit-isolation-user-data

## Workflow backfill for legacy shared-DB upgrades

If you are upgrading a legacy shared-database deployment that already has workflow data, run the following SQL **after the upgrade and before the first server start** so pre-existing rows are assigned to the first user instead of remaining at `user_id = 0`:

```sql
UPDATE workflows SET user_id = (SELECT id FROM users LIMIT 1) WHERE user_id = 0;
UPDATE workflow_runs SET user_id = (SELECT id FROM users LIMIT 1) WHERE user_id = 0;
UPDATE workflow_approvals SET user_id = (SELECT id FROM users LIMIT 1) WHERE user_id = 0;
```

This backfill is only needed for legacy shared-DB installs with existing workflow rows. New rows written after this change already persist the authenticated owner's `user_id`.

## Explicitly out of scope

`pkg/channels/multiuser_manager.go` still creates a workspace-local `data.db` via `storage.New()` instead of routing through `UserStorageManager` and `user.db`. That inconsistency predates this change and is documented as a future follow-up, not part of the isolation fixes in this release.
