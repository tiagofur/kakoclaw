# Proposal: Phase 8.1 — Security & Trust

**Change**: `phase-8.1-security-trust`
**Status**: Draft
**Inspired by**: OpenClaw (https://github.com/openclaw/openclaw)
**Date**: 2026-03-30

---

## Intent

MakoClaw's current security model is static: allowlists are config-only, tool permissions are role-only with no runtime interception, and there are no named tool profiles for deployment contexts. This creates three gaps:

1. Unknown DM senders are silently blocked with no path to self-service onboarding.
2. There is no runtime hook layer — tools execute unconditionally once permitted by role.
3. Tool sets can't be named and reused across channels/deployments.

This phase ports the three security primitives that OpenClaw validated in production.

---

## Scope

### In Scope

- **DM Pairing Policy**: challenge-based allowlisting for unknown senders; `dmPolicy` field per channel; `/approve` command; persistent allowlists in storage
- **Security Hooks**: `before_tool_call`, `before_install`, `message_sending` interceptors with priority ordering and approval fallback
- **Tool Profiles**: named presets (`messaging`, `developer`, `minimal`) wired into `ToolPermissionsConfig`

### Out of Scope

- Group policy (group chats have separate semantics — deferred to 8.2)
- UI for pairing approval (terminal + any active channel is sufficient for now)
- Hook plugins/scripting runtime (hooks are Go-only in this phase)
- Per-session tool profile override (config-level only)

---

## Approach

### DM Pairing Policy

Extend each channel config with `dm_policy: "pairing" | "open" | "allowlist" | "disabled"` (default: `"pairing"`). `BaseChannel.HandleMessage` evaluates the policy before dispatching. Unknown senders in `pairing` mode receive a challenge code; the code is stored in `storage.Storage` (new table `pairing_store`). The `/approve <channel> <code>` command (handled by `CommandHandler`) writes the sender into the persistent allowlist. Approved senders persist across restarts.

### Security Hooks

New `pkg/hooks` package. `HookRegistry` holds ordered `HookHandler` entries (by priority). The agent loop calls hooks at three points:
- `before_tool_call` — receives tool name + args; can return `allow | block | require_approval`
- `before_install` — fires when installing a skill/plugin; can block
- `message_sending` — fires before outbound message delivery; can cancel

If a hook returns `require_approval` and no handler handles it, the request is forwarded as a message to the first available channel owner (fallback approval path).

### Tool Profiles

Add `ToolProfiles map[string]ToolProfileConfig` to `ToolPermissionsConfig`. `ToolProfileConfig` has `Allow []string`, `Deny []string`, `ExecSecurity string` (`deny|ask|full`). Built-in profiles: `messaging` (no exec, no fs-write), `developer` (exec full, full fs), `minimal` (message + query_knowledge only). Profile is applied after role filtering in `filterToolsByPermissions`.

---

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `pkg/channels/base.go` | Modified | Add `dmPolicy` evaluation in `HandleMessage`; pairing challenge send |
| `pkg/config/config.go` | Modified | Add `DMPolicy` to all channel configs; add `ToolProfiles` to `ToolPermissionsConfig` |
| `pkg/storage/sqlite.go` | Modified | Add `pairing_store` table migration |
| `pkg/agent/loop.go` | Modified | Wire hook calls at tool execution and message send points |
| `pkg/agent/permissions.go` | Modified | Apply tool profile on top of role defaults |
| `pkg/channels/command_handler.go` | Modified | Register `/approve <channel> <code>` command |
| `pkg/hooks/` | New | `HookRegistry`, `HookHandler` interface, built-in hook runner |
| `pkg/storage/pairing.go` | New | Pairing store CRUD (challenge codes + approved allowlist) |

---

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Pairing breaks existing allow_from configs | Med | `allowlist` dmPolicy mirrors current static behavior; default `pairing` only activates for senders not in allow_from |
| Hook panics block agent loop | Low | Recover from panics in hook runner; log and continue |
| Storage migration breaks existing DBs | Low | Additive `CREATE TABLE IF NOT EXISTS` migration |
| Approval fallback floods owner channel | Med | Rate-limit approval requests; deduplicate per sender per session |

---

## Rollback Plan

- DM pairing: set `dm_policy: "allowlist"` in config to restore static-only behavior. Pairing store table is additive — no data loss.
- Security hooks: `pkg/hooks` is invoked only if registered handlers exist. Remove registrations in `loop.go` to disable entirely.
- Tool profiles: `ToolProfiles` is optional map; removing entries falls back to `RoleDefaults`.

---

## Dependencies

- None external. Internal: `storage.Storage` must be available on channel startup (already is via `BaseChannel.SetStorage`).

---

## Success Criteria

- [ ] Unknown Telegram sender receives pairing challenge, completes flow with `/approve`, is added to persistent allowlist
- [ ] `before_tool_call` hook can block `exec` and the agent reports the block reason
- [ ] `messaging` tool profile prevents `exec` and `write_file` from appearing in agent's tool list
- [ ] Existing `allow_from` configs continue working unchanged with `dm_policy: "allowlist"` default override path
- [ ] All new behavior covered by unit tests in `pkg/hooks/` and `pkg/storage/`

## Next Steps

- `sdd-spec` and `sdd-design` can run in parallel (both depend only on this proposal)
