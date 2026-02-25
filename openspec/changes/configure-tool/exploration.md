# Exploration: Configure Tool for Agent-Assisted Configuration

## Problem Statement

MakoClaw's multi-user architecture isolates each user's configuration at `~/.MakoClaw/users/{uuid}/config.json`, while their workspace is sandboxed to `~/.MakoClaw/users/{uuid}/workspace/`. When `restrict_to_workspace=true` (the default), the agent's filesystem tools cannot access the config file, preventing the agent from helping users configure their channels and providers.

**Current state:**
- User config: `~/.MakoClaw/users/{uuid}/config.json` (OUTSIDE workspace)
- User workspace: `~/.MakoClaw/users/{uuid}/workspace/` (sandbox boundary)
- Agent tools with `restrict=true` cannot escape the workspace
- Users must manually edit JSON config files or use the web UI

**Desired state:**
- Agent can help users configure channels, providers, and agent settings
- Sensitive values (API keys, tokens, passwords) are write-only (never exposed)
- All configuration changes are audit-logged
- Strict whitelist of allowed fields prevents misconfiguration

---

## Investigation Summary

### 1. Tool Implementation Pattern (`pkg/tools/`)

All tools implement the `Tool` interface from `pkg/tools/base.go`:

```go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}
```

Optional extension interfaces:
- **ContextualTool** - `SetContext(channel, chatID string)` - for tools needing message context
- **WorkspaceTool** - `SetWorkspace(workspace string)` - for tools operating on workspace paths
- **UserAwareTool** - `SetUserID(userID int64)` - for tools that filter data by user

The `UserAwareTool` interface is most relevant for the configure tool, as it provides the user context needed to locate the correct config file.

**Example: TaskTool (`pkg/tools/tasks.go`):**
```go
type TaskTool struct {
    storage *storage.Storage
    userID  int64
}

func (t *TaskTool) SetUserID(userID int64) {
    if userID > 0 {
        t.userID = userID
    }
}
```

### 2. Config Loading/Saving (`pkg/config/config.go`)

Key functions:
- `LoadConfigForUser(userUUID string)` - Loads user-specific config, falls back to global
- `SaveConfigForUser(userUUID string, cfg *Config)` - Saves user config
- `GetUserConfigPath(userUUID string)` - Returns `~/.MakoClaw/users/{uuid}/config.json`
- `MergeConfigs(global, user *Config)` - Merges user config over global

**Config structure (simplified):**
```go
type Config struct {
    Agents    AgentsConfig    // defaults, orchestrator, specialists
    Channels  ChannelsConfig  // telegram, discord, slack, etc.
    Providers ProvidersConfig // anthropic, openai, groq, etc.
    Tools     ToolsConfig     // web search, email, MCP
    // ...
}
```

### 3. User Context in Tools (`pkg/agent/loop.go`)

The agent loop provides user context through:

1. **SetUserForAgent(userUUID string, userID int64)** - Sets user context on the agent loop
2. **updateToolsUser(userID int64)** - Calls `SetUserID` on all `UserAwareTool` implementations
3. **applyMessageUserContext(msg)** - Resolves user from message and applies context

**Key insight:** The configure tool needs the `userUUID` (not just `userID`) to locate the config file. The current `UserAwareTool` interface only provides `userID`. We'll need a new interface.

### 4. Audit Logging (`pkg/tools/audit.go`)

Existing audit infrastructure:

```go
type ToolExecutionLog struct {
    Timestamp time.Time
    UserID    int64
    Username  string
    Tool      string
    Arguments map[string]interface{}  // Sanitized
    Success   bool
    Error     string
    Duration  int64
}

type AuditLogger interface {
    LogToolExecution(ctx context.Context, log ToolExecutionLog) error
    QueryLogs(ctx context.Context, filters AuditQueryFilters) ([]ToolExecutionLog, error)
}
```

The `sanitizeArguments` function already redacts sensitive fields:
```go
sensitiveKeys := []string{
    "password", "token", "secret", "api_key", "apikey",
    "access_token", "refresh_token", "private_key",
}
```

**Key insight:** The configure tool should:
1. Always be added to `RestrictedTools` for mandatory auditing
2. Use `sanitizeArguments` before logging
3. NEVER include sensitive values in return strings

### 5. Tool Registration (`pkg/agent/loop.go`)

Tools are registered in `NewAgentLoop`:

```go
toolsRegistry := tools.NewToolRegistry()
toolsRegistry.Register(tools.NewReadFileTool(workspace, restrict))
toolsRegistry.Register(tools.NewWriteFileTool(workspace, restrict))
// ...
```

The configure tool would be registered similarly, with access to the config package.

### 6. Config Structure Analysis

**Providers Config (10 providers):**
```go
type ProvidersConfig struct {
    Anthropic  ProviderConfig
    OpenAI     ProviderConfig
    OpenRouter ProviderConfig
    Groq       ProviderConfig
    // ... 6 more
}

type ProviderConfig struct {
    APIKey     string   // SENSITIVE - write-only
    APIBase    string   // Safe to read/write
    Proxy      string   // Safe to read/write
    AuthMethod string   // Safe to read/write
    Models     []string // Safe to read/write
}
```

**Channels Config (9 channels):**
```go
type ChannelsConfig struct {
    Telegram TelegramConfig
    Discord  DiscordConfig
    Slack    SlackConfig
    // ... 6 more
}

// Example: TelegramConfig
type TelegramConfig struct {
    Enabled   bool                // Safe
    Token     string              // SENSITIVE
    Proxy     string              // Safe
    AllowFrom FlexibleStringSlice // Safe
}
```

**Agents Config:**
```go
type AgentsConfig struct {
    Defaults     AgentDefaults           // model, provider, temp, etc.
    Orchestrator OrchestratorConfig      // enabled, provider, model
    Specialists  map[string]SpecialistConfig
}
```

---

## Design Considerations

### Field Classification

| Category | Fields | Policy |
|----------|--------|--------|
| **Sensitive (Write-only)** | `api_key`, `token`, `password`, `secret`, `app_secret`, `client_secret`, `bot_token`, `app_token` | Write only, never read, always redacted in logs |
| **Safe (Read/Write)** | `enabled`, `api_base`, `proxy`, `model`, `temperature`, `max_tokens`, `allow_from`, `host`, `port` | Full read/write access |
| **Structural** | `providers.*`, `channels.*`, `agents.*` | Navigate structure, no direct modification |

### Tool Actions

Proposed action set:

1. **get** - Read current config (sensitive fields shown as `[REDACTED]`)
2. **set** - Write a specific field value
3. **enable** - Enable a channel or feature
4. **disable** - Disable a channel or feature
5. **list_providers** - Show configured providers with masked keys
6. **list_channels** - Show configured channels with masked tokens

### Security Model

1. **Whitelist-based access:** Only explicitly allowed paths can be modified
2. **Sensitive field redaction:** API keys shown as `sk-****1234` (last 4 chars) or `[REDACTED]`
3. **Validation:** All values validated before writing (e.g., port ranges, URL formats)
4. **Audit trail:** Every config change logged with user, field, timestamp
5. **No destructive operations:** Cannot delete entire sections, only update fields

### User Context Resolution

New interface needed:

```go
// UserConfigTool is for tools that need access to user config files
type UserConfigTool interface {
    Tool
    SetUserContext(userID int64, userUUID string)
}
```

---

## Implementation Approach

### Option A: Dedicated Configure Tool (Recommended)

Create `pkg/tools/configure.go` with:

```go
type ConfigureTool struct {
    userID   int64
    userUUID string
}

func (t *ConfigureTool) SetUserContext(userID int64, userUUID string) {
    t.userID = userID
    t.userUUID = userUUID
}

func (t *ConfigureTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    action := args["action"].(string)
    switch action {
    case "get":
        return t.handleGet(args)
    case "set":
        return t.handleSet(args)
    case "enable", "disable":
        return t.handleToggle(args, action == "enable")
    case "list_providers", "list_channels":
        return t.handleList(args)
    }
}
```

**Pros:**
- Clean separation of concerns
- Explicit security boundary
- Easy to audit and test
- Can be individually disabled via permissions

**Cons:**
- New code to maintain
- Need to update loop.go to pass userUUID

### Option B: Config API Wrapper

Expose config modification through a controlled API layer in `pkg/config/` that the tool calls.

**Pros:**
- Centralizes config access logic
- Can be reused by web handlers

**Cons:**
- More complex architecture
- Potential for bypass if not careful

### Recommendation: Option A with Config Helpers

Implement as a dedicated tool, but extract reusable config helpers:

1. `pkg/config/redact.go` - Sensitive field redaction utilities
2. `pkg/config/validate.go` - Field validation rules
3. `pkg/tools/configure.go` - The actual tool implementation

---

## Field Whitelist (Draft)

### Providers Section

| Path | Type | Policy | Description |
|------|------|--------|-------------|
| `providers.{name}.api_key` | string | write-only | API key (sensitive) |
| `providers.{name}.api_base` | string | read-write | API base URL |
| `providers.{name}.proxy` | string | read-write | Proxy URL |
| `providers.{name}.auth_method` | string | read-write | Auth method |
| `providers.{name}.models` | []string | read-write | Available models |

### Channels Section

| Path | Type | Policy | Description |
|------|------|--------|-------------|
| `channels.telegram.enabled` | bool | read-write | Enable/disable |
| `channels.telegram.token` | string | write-only | Bot token (sensitive) |
| `channels.telegram.proxy` | string | read-write | Proxy URL |
| `channels.telegram.allow_from` | []string | read-write | Allowed sender IDs |
| `channels.discord.enabled` | bool | read-write | Enable/disable |
| `channels.discord.token` | string | write-only | Bot token (sensitive) |
| `channels.discord.allow_from` | []string | read-write | Allowed sender IDs |
| ... (similar for other channels) |

### Agents Section

| Path | Type | Policy | Description |
|------|------|--------|-------------|
| `agents.defaults.provider` | string | read-write | Default provider |
| `agents.defaults.model` | string | read-write | Default model |
| `agents.defaults.max_tokens` | int | read-write | Max tokens |
| `agents.defaults.temperature` | float | read-write | Temperature |
| `agents.defaults.max_tool_iterations` | int | read-write | Max iterations |
| `agents.orchestrator.enabled` | bool | read-write | Enable orchestrator |
| `agents.orchestrator.provider` | string | read-write | Orchestrator provider |
| `agents.orchestrator.model` | string | read-write | Orchestrator model |

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| LLM prompt injection to extract secrets | High | Never return sensitive values, only confirmations |
| Path traversal to modify other users' configs | High | Validate userUUID from authenticated context |
| Invalid config breaks system | Medium | Validate all values before writing |
| Audit log leaks secrets | Medium | Always sanitize arguments before logging |
| Race condition on config writes | Low | Use mutex/lock on config save |
| Excessive config changes | Low | Rate limit tool calls |

---

## Open Questions

1. **Schema validation:** Should we validate the entire config after modification, or just the changed field?
   - Recommendation: Validate field + basic config validity (e.g., can still load)

2. **Config reload:** Should changes take effect immediately or require restart?
   - Recommendation: Flag that config changed, let user decide when to restart/reload

3. **Undo capability:** Should we support reverting config changes?
   - Recommendation: Out of scope for v1, but log previous values in audit

4. **Specialist management:** Should the tool allow creating/deleting specialists?
   - Recommendation: v2 feature - start with simple field updates

5. **MCP server config:** Should MCP servers be configurable via this tool?
   - Recommendation: v2 feature - complex nested structure

---

## Next Steps

1. **Spec Phase:** Define exact tool parameters, actions, and field whitelist
2. **Design Phase:** Detail the implementation architecture and interfaces
3. **Tasks Phase:** Break down into implementable tasks
4. **Implementation:** Build the tool incrementally with tests

---

## Related Files

- `pkg/tools/base.go` - Tool interface definitions
- `pkg/tools/audit.go` - Audit logging infrastructure
- `pkg/tools/tasks.go` - Example UserAwareTool implementation
- `pkg/config/config.go` - Config loading/saving/merging
- `pkg/config/workspace_init.go` - User workspace initialization
- `pkg/agent/loop.go` - Tool registration and user context
- `pkg/agent/permissions.go` - Tool permission filtering
- `pkg/storage/central.go` - User management and central DB
- `pkg/storage/user.go` - User model and tool permissions
