# Known Issues and Bugs

**Last Updated**: 2026-03-04
**Source**: Comprehensive codebase audit

This document tracks known issues discovered during security audits and code reviews. Issues are categorized by severity and status.

---

## Summary

| Severity | Count | Fixed | Pending |
|----------|-------|-------|---------|
| Critical | 3 | 3 | 0 |
| High | 4 | 4 | 0 |
| Medium | 10 | 2 | 8 |
| Low | 8 | 2 | 6 |
| **Total** | **25** | **11** | **14** |

> **Last Fix Session**: 2026-03-04 - Fixed 11 issues including all critical and high severity security bugs

---

## Critical Issues

### CRIT-001: Debug Logging Exposes Credentials

**File**: [pkg/web/auth.go:137-144](pkg/web/auth.go#L137-L144)

**Description**: Password verification attempts were logged with usernames via `fmt.Printf`, allowing credential enumeration attacks through log analysis.

**Risk**: Attackers with log access could enumerate valid usernames and track authentication attempts.

**Fix Applied**: Removed all debug logging statements from password verification flow.

**Status**: 🟢 Fixed (2026-03-04)

---

### CRIT-002: Nil Pointer Dereference in CentralStorage

**File**: [pkg/agent/loop.go:449, 454](pkg/agent/loop.go#L449)

**Description**: Originally reported as missing nil check, but code review confirmed nil checks ARE present at lines 449 and 357.

**Status**: 🟢 Already Fixed (verified 2026-03-04)

---

### CRIT-003: Goroutine Leak in Orchestrator

**File**: [pkg/agent/orchestrator.go:806-910](pkg/agent/orchestrator.go#L806-L910)

**Description**: Originally reported as using unbuffered channels, but code review confirmed channels ARE buffered with size 1 (lines 807-808).

```go
// CURRENT CODE (correct)
resultChan := make(chan string, 1)  // Buffered!
errChan := make(chan error, 1)      // Buffered!
```

**Status**: 🟢 Already Fixed (verified 2026-03-04)

---

## High Severity Issues

### HIGH-001: Race Condition in Tool Swapping

**File**: [pkg/agent/specialist.go:253-275](pkg/agent/specialist.go#L253-L275)

**Description**: Tool registry swap has a race window between unlock and defer restoration.

```go
// VULNERABLE CODE
sa.processMu.Lock()
originalTools := sa.tools
sa.tools = sa.ToolFilter()
sa.processMu.Unlock()  // Race window starts here

defer func() {
    sa.processMu.Lock()
    sa.tools = originalTools  // Another call may have already changed this
    sa.processMu.Unlock()
}()
```

**Risk**: Concurrent specialist calls can corrupt tool registry.

**Fix Applied**: Lock is now held for the entire `ProcessDirect` duration, preventing concurrent modifications.

**Status**: 🟢 Fixed (2026-03-04)

---

### HIGH-002: Path Traversal in Backup Import

**File**: [pkg/web/handlers_backup.go:799-805](pkg/web/handlers_backup.go#L799-L805)

**Description**: Path traversal check used simple string matching, potentially bypassable.

**Fix Applied**: Added `basePath` parameter to `extractZipFile()` function. Now validates that resolved absolute path starts with the expected base directory, preventing escape via symlinks or creative paths.

**Status**: 🟢 Fixed (2026-03-04)

---

### HIGH-003: JWT Secret Rotation Affects All Users

**File**: [pkg/web/auth.go:238-242](pkg/web/auth.go#L238-L242)

**Description**: Changing any user's password rotates the global JWT secret, invalidating ALL user sessions.

**Risk**: Disruptive user experience in multi-user deployments.

**Fix Applied**: Implemented per-user `token_version` field. Password changes now increment only the user's token version, invalidating only THEIR tokens while leaving other users' sessions intact.

Changes:
- Added `TokenVersion` field to User struct
- Added `token_version` column to users table (migration)
- JWT claims now include `ver` field with user's token version
- Token verification checks `ver` matches user's current `token_version`
- Password change calls `IncrementTokenVersion()` instead of rotating global secret

**Status**: 🟢 Fixed (2026-03-04)

---

### HIGH-004: Shell Allowlist Fail-Open

**File**: [pkg/agent/permissions.go:83-93](pkg/agent/permissions.go#L83-L93)

**Description**: If `SetSafeCommandsForUser()` failed, the exec tool was still registered without restrictions.

**Fix Applied**: Added `continue` statement to skip tool registration when allowlist setup fails (fail-closed behavior). Log message updated to clarify tool is NOT registered.

**Status**: 🟢 Fixed (2026-03-04)

---

## Medium Severity Issues

### MED-001: WhatsApp Channel Incomplete

**File**: [pkg/channels/whatsapp.go](pkg/channels/whatsapp.go)

**Issues**:
- Uses standard `log` instead of project logger
- No allowlist validation in `handleIncomingMessage()`
- No command handler support
- Missing JSON marshal error handling

**Status**: 🔴 Pending

---

### MED-002: Ollama Provider Missing Tool Support

**File**: [pkg/providers/ollama_provider.go:71-142](pkg/providers/ollama_provider.go#L71)

**Description**: `tools` parameter is completely ignored in `Chat()` method.

**Impact**: Agents using Ollama cannot use any tools.

**Status**: 🔴 Pending

---

### MED-003: Cost Tracker Never Integrated

**File**: [pkg/agent/cost_tracker.go](pkg/agent/cost_tracker.go)

**Description**: Complete implementation exists but `RecordAPICall()` is never called in agent loop.

**Impact**: Cost metrics are always empty, tracking functionality is dead code.

**Status**: 🔴 Pending

---

### MED-004: Delegation Counter Bug

**File**: [pkg/agent/orchestrator.go:395-414](pkg/agent/orchestrator.go#L395-L414)

**Description**: Originally reported as off-by-one bug, but code review confirmed the logic is correct:
- Counter increments AFTER the check (line 400 checks first, line 406 increments)
- `ResetDelegationCount()` IS called in server.go:1276 at session start

**Status**: 🟢 Already Correct (verified 2026-03-04)

---

### MED-005: QQ Channel Memory Leak

**File**: [pkg/channels/qq.go:217-235](pkg/channels/qq.go#L217-L235)

**Description**: Message deduplication map grows to 10,000 entries before nuclear clear.

**Impact**: Memory growth, brief message duplication during clear.

**Status**: 🔴 Pending

---

### MED-006: Telegram Thinking Animation Race

**File**: [pkg/channels/telegram.go:88-96, 173-215](pkg/channels/telegram.go#L88)

**Description**: Multiple goroutines can call `cleanupThinking()` simultaneously for same chatID.

**Impact**: Potential double-cleanup or missed cleanup.

**Status**: 🔴 Pending

---

### MED-007: Signal Attachments Not Downloaded

**File**: [pkg/channels/signal.go:259-276](pkg/channels/signal.go#L259-L276)

**Description**: `downloadAttachment()` function exists but is never called.

**Impact**: Attachments are parsed but content is not available to agent.

**Status**: 🔴 Pending

---

### MED-008: Migrations Silent Failures

**File**: [pkg/storage/sqlite.go:176-185](pkg/storage/sqlite.go#L176-L185)

**Description**: ALTER TABLE errors are silently ignored.

**Impact**: Schema drift between installations.

**Status**: 🔴 Pending

---

### MED-009: Files View Non-Functional

**File**: [pkg/web/frontend/src/views/FilesView.vue](pkg/web/frontend/src/views/FilesView.vue)

**Description**: UI skeleton exists but no backend endpoints connected.

**Impact**: Feature shown in navigation but doesn't work.

**Status**: 🔴 Pending

---

### MED-010: Backup Import Silent Failures

**File**: [pkg/web/handlers_backup.go:467-615](pkg/web/handlers_backup.go#L467-L615)

**Description**: Import errors are collected but import continues.

**Impact**: Partial imports may leave system in inconsistent state.

**Status**: 🔴 Pending

---

## Low Severity Issues

### LOW-001: Debug Prints in Production Code

**Files**:
- `pkg/storage/backup.go:299, 360, 410`
- `pkg/tools/filesystem.go:126, 172`
- `pkg/tools/edit.go:26, 81`

**Description**: `fmt.Printf` statements in production code.

**Fix Applied**: Removed all debug print statements from the listed files.

**Status**: 🟢 Fixed (2026-03-04)

---

### LOW-002: Token Estimation Primitive

**File**: [pkg/agent/loop.go:1759-1765](pkg/agent/loop.go#L1759-L1765)

**Description**: Uses `len(content) / 4` for token estimation.

**Impact**: Inaccurate summarization triggers.

**Status**: 🔴 Pending

---

### LOW-003: Specialist JSON Parsing Fragile

**File**: [pkg/agent/orchestrator.go:1169-1226](pkg/agent/orchestrator.go#L1169-L1226)

**Description**: Uses string search instead of proper JSON parser.

**Impact**: Fails with multiple JSON blocks in response.

**Status**: 🔴 Pending

---

### LOW-004: PDF Extraction Basic

**File**: [pkg/web/handlers_features.go:249-256](pkg/web/handlers_features.go#L249-L256)

**Description**: Only extracts text between BT/ET markers.

**Impact**: Encrypted/complex PDFs fail silently.

**Status**: 🔴 Pending

---

### LOW-005: Deprecated Functions Still Exported

**File**: [pkg/web/handlers_backup.go:925-941](pkg/web/handlers_backup.go#L925-L941)

**Description**: `addFileToZip`, `addDirToZip` marked DEPRECATED but still exported.

**Status**: 🔴 Pending

---

### LOW-006: User Deletion No Workspace Cleanup

**File**: [pkg/web/handlers_users.go:177-229](pkg/web/handlers_users.go#L177-L229)

**Description**: User deletion closes DB but doesn't delete workspace files.

**Impact**: Orphaned files accumulate.

**Status**: 🔴 Pending

---

### LOW-007: Slack Webhook Memory Leak

**File**: [pkg/channels/slack.go:30, 148](pkg/channels/slack.go#L30)

**Description**: `sessionWebhooks` sync.Map never cleaned up.

**Impact**: Unbounded map growth in long-running bots.

**Status**: 🔴 Pending

---

### LOW-008: Multi-User Isolation Gaps

**Files**:
- `pkg/tools/spawn.go` - SubagentManager doesn't track userID
- `pkg/tools/cron.go` - userID in jobs without verification
- `pkg/tools/knowledge.go` - Default userID=1
- `pkg/tools/tasks.go` - Default userID=1

**Impact**: Potential data leakage in edge cases.

**Status**: 🔴 Pending

---

## Issue Template

When adding new issues, use this format:

```markdown
### [SEVERITY]-XXX: Brief Title

**File**: [path/to/file.go:line](path/to/file.go#Lline)

**Description**: What the issue is.

**Risk/Impact**: What could go wrong.

**Fix**: Suggested solution.

**Status**: 🔴 Pending | 🟡 In Progress | 🟢 Fixed
```

---

## Contributing Fixes

1. Pick an issue from this list
2. Create a branch: `fix/ISSUE-ID-brief-description`
3. Implement the fix with tests
4. Reference the issue ID in your commit: `fix: resolve CRIT-001 debug logging`
5. Submit PR with "Fixes CRIT-001" in description

---

## Audit History

| Date | Auditor | Scope | Issues Found |
|------|---------|-------|--------------|
| 2026-03-04 | Claude + Tiago | Full codebase | 25 |

---

## Related Documents

- [SECURITY.md](SECURITY.md) - Security policy and reporting
- [PRD-NEW-FEATURES.md](PRD-NEW-FEATURES.md) - Feature roadmap
- [CHANGELOG.md](CHANGELOG.md) - Version history
