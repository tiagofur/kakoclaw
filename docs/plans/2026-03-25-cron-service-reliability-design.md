# Cron Service Reliability Design

**Date:** 2026-03-25
**Status:** Approved

## Problem Statement

Two related issues prevent reliable cron job management in MakoClaw's web mode:

1. **Agent in web chat cannot manage cron jobs via API** — `newAgentLoopForUser()` creates agent loops without registering the `CronTool`. The agent falls back to `write_file` to modify `jobs.json` directly.
2. **CronService overwrites external file changes** — The service loads `jobs.json` once at startup, keeps an in-memory copy, and blindly overwrites the file on every save (including every 1-second tick). External edits are lost.
3. **Web mode doesn't initialize per-user cron services** — `webCmd()` creates a `MultiUserChannelManager` but never calls `InitializeAllUsers()`. (Already fixed.)

## Approach: Fix the Plumbing + Mtime Safety Net

### Part 1: Wire CronTool into web chat agent loop

**Location:** `pkg/web/server.go` — `handleChatWS()`

After creating the agent loop and agent manager (line ~1196), obtain the user's CronService from the `multiUserChannelManager` and register a `CronTool` on the `activeAgentLoop`.

**Flow:**
1. `handleChatWS()` creates `activeAgentLoop` via `newAgentLoopForUser()` + `AgentManager`
2. Get `cronService` from `s.multiUserChannelManager.GetCronServiceForUser(userUUID)`
3. Create `CronTool` with `cronService`, `activeAgentLoop`, `s.msgBus`
4. Register tool: `activeAgentLoop.RegisterTool(cronTool)`

**Why this works:** The `multiUserChannelManager` already creates and manages per-user CronService instances (via `InitializeAllUsers()` — now called in web mode too). We just need to give the chat agent access to it.

### Part 2: Mtime safety net in CronService

**Location:** `pkg/cron/service.go`

Add external change detection to prevent overwriting edits made outside the CronService.

**New field:**
```go
type CronService struct {
    // ... existing fields
    lastSavedMtime time.Time  // mtime after our last save
}
```

**Modified `saveStoreUnsafe()`:**
1. Before writing, `os.Stat()` the file to get current mtime
2. If mtime differs from `lastSavedMtime`, the file was modified externally
3. On external change: reload from disk, perform union merge with in-memory state
4. Save merged result
5. Update `lastSavedMtime` from the newly written file

**Union merge strategy:**
- Jobs in memory are authoritative (they reflect API operations)
- Jobs on disk that don't exist in memory are NEW (added externally) — add them
- Jobs deleted from disk but present in memory are kept (deletion should go through API)
- If same job ID exists in both, memory wins (it has the latest state)

**Why union merge:** The CronService is the source of truth for state changes (enable/disable, last run, next run). External edits are typically additions or full replacements. Union merge preserves both without data loss.

### Part 3: InitializeAllUsers in web mode (Already Applied)

**Location:** `cmd/makoclaw/main.go` — `webCmd()`

Added `multiChannelManager.InitializeAllUsers()` after creating the manager, matching the pattern used in `gatewayCmd()`.

## Files to Modify

| File | Change |
|------|--------|
| `pkg/web/server.go` | Register CronTool in `handleChatWS()` |
| `pkg/cron/service.go` | Add mtime tracking + union merge in `saveStoreUnsafe()` |
| `cmd/makoclaw/main.go` | Already done: `InitializeAllUsers()` in web mode |

## Risks

- **Merge edge case:** If someone externally replaces ALL jobs (full file rewrite), the union merge will add back in-memory jobs that were intentionally removed. Mitigation: this is acceptable — the proper way to remove jobs is via API.
- **Performance:** `os.Stat()` on every save adds negligible overhead (sub-microsecond on modern FS).
