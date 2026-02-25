# Proposal: Configure Tool for Agent-Assisted Configuration

**Change ID:** configure-tool
**Status:** Draft
**Author:** SDD Sub-Agent
**Created:** 2026-02-25
**Based on:** [exploration.md](exploration.md)

---

## 1. Intent

### Problem Statement

MakoClaw's multi-user architecture isolates each user's configuration at `~/.MakoClaw/users/{uuid}/config.json`, while their workspace is sandboxed to `~/.MakoClaw/users/{uuid}/workspace/`. When `restrict_to_workspace=true` (the default), the agent's filesystem tools cannot access the config file, preventing the agent from helping users configure their channels, providers, and agent settings.

### Current State

- User config: `~/.MakoClaw/users/{uuid}/config.json` (OUTSIDE workspace sandbox)
- User workspace: `~/.MakoClaw/users/{uuid}/workspace/` (sandbox boundary)
- Agent tools with `restrict=true` cannot escape the workspace
- Users must manually edit JSON config files or use the web UI
- No agent-assisted configuration capability

### Desired State

- Agent can help users configure channels, providers, and agent settings through natural conversation
- Sensitive values (API keys, tokens, passwords) are write-only and never exposed to the LLM
- All configuration changes are audit-logged with full traceability
- Strict whitelist of allowed fields prevents misconfiguration
- Configuration changes are validated before being applied

### Value Proposition

1. **Improved UX:** Users can configure MakoClaw through conversation ("Set my OpenAI API key to sk-...")
2. **Security:** Sensitive credentials are never readable, preventing prompt injection attacks
3. **Auditability:** Full audit trail of all configuration changes
4. **Safety:** Whitelist-based access prevents destructive or invalid configurations

---

## 2. Scope

### In Scope

| Component | Description |
|-----------|-------------|
| `pkg/tools/configure.go` | New ConfigureTool implementing the Tool interface |
| `pkg/tools/base.go` | New `UserConfigTool` interface for UUID-based config access |
| `pkg/config/redact.go` | Utility functions for sensitive field redaction |
| `pkg/config/validate.go` | Field validation rules and config validation |
| `pkg/agent/loop.go` | Tool registration and user context propagation |
| Field whitelist | Explicit list of configurable fields with read/write policies |
| Audit logging | All config changes logged via existing AuditLogger |

### Out of Scope (v1)

| Item | Reason |
|------|--------|
| Specialist creation/deletion | Complex nested structure - defer to v2 |
| MCP server configuration | Complex nested structure - defer to v2 |
| Config undo/rollback | Log previous values in audit, but no active rollback |
| Live config reload | Flag changes, require user to restart/reload |
| Web UI integration | Web already has config editing - this is for agent |

### Dependencies

| Dependency | Type | Description |
|------------|------|-------------|
| `pkg/config` | Internal | Config loading, saving, merging functions |
| `pkg/tools/audit.go` | Internal | Audit logging infrastructure |
| `pkg/storage/central.go` | Internal | User UUID resolution |
| `pkg/agent/loop.go` | Internal | Tool registration, user context |

---

## 3. Approach

### Architecture Overview

```
User Request → Agent Loop → ConfigureTool → Config Package → Config File
                              ↓
                         Audit Logger
```

### Selected Option: Dedicated Tool with Whitelist (Option A)

Based on the exploration analysis, we will implement **Option A: Dedicated Configure Tool** with config helpers:

1. **`pkg/tools/configure.go`** - The tool implementation with whitelist enforcement
2. **`pkg/config/redact.go`** - Reusable sensitive field redaction utilities
3. **`pkg/config/validate.go`** - Field validation rules

### New Interface: UserConfigTool

The existing `UserAwareTool` interface only provides `userID` (int64). The configure tool needs `userUUID` (string) to locate the config file. We will add:

```go
// UserConfigTool is for tools that need access to user config files
type UserConfigTool interface {
    Tool
    SetUserContext(userID int64, userUUID string)
}
```

### Tool Actions

| Action | Description | Example |
|--------|-------------|---------|
| `get` | Read config value (sensitive fields redacted) | `get providers.openai.api_base` |
| `set` | Write a config value | `set providers.openai.api_key sk-...` |
| `enable` | Enable a channel or feature | `enable channels.telegram` |
| `disable` | Disable a channel or feature | `disable channels.discord` |
| `list_providers` | List configured providers with masked keys | Shows which have keys set |
| `list_channels` | List configured channels with status | Shows enabled/disabled |

### Security Model

1. **Whitelist-Based Access**
   - Only explicitly whitelisted paths can be read or modified
   - Paths not in whitelist return "field not configurable"

2. **Field Classification**
   | Category | Policy | Examples |
   |----------|--------|----------|
   | Sensitive | Write-only | `api_key`, `token`, `password`, `secret` |
   | Safe | Read-write | `enabled`, `api_base`, `model`, `temperature` |
   | Structural | Navigate only | `providers`, `channels`, `agents` |

3. **Response Redaction**
   - Sensitive fields shown as `[REDACTED]` or `****1234` (last 4 chars)
   - Tool responses never contain actual secret values
   - Even successful writes return generic confirmations

4. **Audit Trail**
   - Tool added to `RestrictedTools` for mandatory auditing
   - All arguments sanitized before logging
   - Log captures: user, action, path, timestamp, success/failure

5. **Validation**
   - All values validated before writing
   - Invalid config prevented (e.g., bad port ranges, invalid URLs)
   - Config file verified loadable after write

### Field Whitelist (v1)

#### Providers Section
| Path | Type | Policy |
|------|------|--------|
| `providers.{name}.api_key` | string | write-only |
| `providers.{name}.api_base` | string | read-write |
| `providers.{name}.proxy` | string | read-write |
| `providers.{name}.auth_method` | string | read-write |
| `providers.{name}.models` | []string | read-write |

Where `{name}` is one of: `anthropic`, `openai`, `openrouter`, `groq`, `zhipu`, `gemini`, `nvidia`, `moonshot`, `vllm`, `ollama`

#### Channels Section
| Path | Type | Policy |
|------|------|--------|
| `channels.{name}.enabled` | bool | read-write |
| `channels.{name}.token` | string | write-only |
| `channels.{name}.proxy` | string | read-write |
| `channels.{name}.allow_from` | []string | read-write |

Where `{name}` is one of: `telegram`, `discord`, `slack`, `whatsapp`, `signal`, `qq`, `dingtalk`, `feishu`, `maixcam`

Note: Some channels have additional sensitive fields (`app_id`, `app_secret`, `client_secret`, `bot_token`, `signing_secret`) - all write-only.

#### Agents Section
| Path | Type | Policy |
|------|------|--------|
| `agents.defaults.provider` | string | read-write |
| `agents.defaults.model` | string | read-write |
| `agents.defaults.max_tokens` | int | read-write |
| `agents.defaults.temperature` | float | read-write |
| `agents.defaults.max_tool_iterations` | int | read-write |
| `agents.orchestrator.enabled` | bool | read-write |
| `agents.orchestrator.provider` | string | read-write |
| `agents.orchestrator.model` | string | read-write |

---

## 4. Risks and Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| LLM prompt injection extracts secrets | High | Medium | Never return sensitive values; write-only policy enforced in code |
| Path traversal to other users' configs | High | Low | Validate userUUID from authenticated context only; no path construction from user input |
| Invalid config breaks system | Medium | Medium | Validate all values before writing; verify config loads after save |
| Audit log leaks secrets | Medium | Low | Sanitize arguments before logging using existing `sanitizeArguments` |
| Race condition on config writes | Low | Low | Use mutex/lock on config save operations |
| Excessive config changes | Low | Low | Existing rate limiting applies; tool calls are rate-limited |

---

## 5. Success Criteria

| Criterion | Measurement |
|-----------|-------------|
| Security | No sensitive values exposed in tool responses or audit logs |
| Functionality | All whitelisted fields can be configured via agent |
| Validation | Invalid values rejected with helpful error messages |
| Audit | All config changes appear in audit log with sanitized arguments |
| UX | Users can configure channels/providers through natural conversation |
| Testing | Unit tests cover all actions, field policies, and edge cases |

---

## 6. Alternatives Considered

### Alternative: Config API Wrapper (Option B)

Expose config modification through a controlled API layer in `pkg/config/` that the tool calls.

**Pros:**
- Centralizes config access logic
- Can be reused by web handlers

**Cons:**
- More complex architecture
- Potential for bypass if not careful
- Over-engineering for current needs

**Decision:** Rejected in favor of dedicated tool with config helpers. The tool provides cleaner security boundary while helpers enable code reuse.

### Alternative: Expand Filesystem Tool

Modify existing filesystem tools to allow config access when called for specific paths.

**Decision:** Rejected. Mixing concerns weakens security model. Config access should be explicit, not a special case of file access.

---

## 7. Open Questions

| Question | Recommendation | Status |
|----------|----------------|--------|
| Schema validation scope? | Validate changed field + verify config loads | Decided |
| Config reload behavior? | Flag change, require user restart/reload | Decided |
| Undo capability? | Out of scope v1, log previous values | Decided |
| Specialist management? | v2 feature | Deferred |
| MCP server config? | v2 feature | Deferred |

---

## 8. Implementation Notes

### Tool Registration

The ConfigureTool will be registered in `NewAgentLoop` alongside other tools:

```go
configureTool := tools.NewConfigureTool()
toolsRegistry.Register(configureTool)
```

### User Context Propagation

The agent loop needs modification to propagate `userUUID` to tools implementing `UserConfigTool`:

```go
func (al *AgentLoop) updateToolsUserConfig(userID int64, userUUID string) {
    for _, tool := range al.tools {
        if uct, ok := tool.(UserConfigTool); ok {
            uct.SetUserContext(userID, userUUID)
        }
    }
}
```

### Restricted Tool Registration

Add to `RestrictedTools` in audit configuration:

```go
RestrictedTools: []string{
    "exec", "spawn", "email", "schedule", "configure", // Added
},
```

---

## 9. Related Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Exploration | `openspec/changes/configure-tool/exploration.md` | Complete |
| Proposal | `openspec/changes/configure-tool/proposal.md` | This document |
| Specification | `openspec/changes/configure-tool/spec.md` | Pending |
| Design | `openspec/changes/configure-tool/design.md` | Pending |
| Tasks | `openspec/changes/configure-tool/tasks.md` | Pending |

---

## 10. Approval

| Role | Name | Date | Decision |
|------|------|------|----------|
| Author | SDD Sub-Agent | 2026-02-25 | Proposed |
| Reviewer | | | |
| Approver | | | |

---

## Appendix A: Example Interactions

### Setting an API Key

**User:** "Set my OpenAI API key to sk-proj-abc123..."

**Agent (internal):** `configure(action="set", path="providers.openai.api_key", value="sk-proj-abc123...")`

**Tool Response:** "Successfully updated providers.openai.api_key"

**Agent:** "I've set your OpenAI API key. The key has been securely stored and will be used for future requests."

### Checking Provider Configuration

**User:** "What providers do I have configured?"

**Agent (internal):** `configure(action="list_providers")`

**Tool Response:**
```
Configured providers:
- anthropic: api_key=[REDACTED], api_base=default
- openai: api_key=****c123, api_base=default
- groq: api_key=[NOT SET]
```

### Enabling a Channel

**User:** "Enable Telegram for me"

**Agent (internal):** `configure(action="enable", path="channels.telegram")`

**Tool Response:** "Successfully enabled channels.telegram. Note: You may need to restart the gateway for changes to take effect."

---

## Appendix B: Sensitive Field Patterns

The following field name patterns are classified as sensitive (write-only):

- `api_key`, `apikey`
- `token`, `bot_token`, `app_token`, `access_token`, `refresh_token`
- `password`
- `secret`, `app_secret`, `client_secret`, `signing_secret`
- `private_key`

These patterns are matched case-insensitively at any nesting level.
