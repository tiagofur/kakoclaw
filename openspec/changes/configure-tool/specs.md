# Specification: Configure Tool

**Change ID:** configure-tool
**Status:** Draft
**Author:** SDD Sub-Agent
**Created:** 2026-02-25
**Based on:** [proposal.md](proposal.md)

---

## 1. Tool Interface

### 1.1 Name and Description

| Property | Value |
|----------|-------|
| **Name** | `configure` |
| **Description** | "Manage user configuration for providers, channels, and agent settings. Supports reading non-sensitive values, writing any whitelisted field, and enabling/disabling services. All operations are audited and sensitive values (API keys, tokens) are write-only." |

### 1.2 Parameters Schema (JSON Schema)

```json
{
  "type": "object",
  "properties": {
    "action": {
      "type": "string",
      "enum": ["get", "set", "enable", "disable", "list_providers", "list_channels"],
      "description": "The configuration action to perform"
    },
    "path": {
      "type": "string",
      "description": "Dot-notation path to the config field (e.g., 'providers.openai.api_key', 'channels.telegram.enabled'). Required for get, set, enable, disable actions."
    },
    "value": {
      "type": ["string", "number", "boolean", "array"],
      "description": "The value to set. Required for 'set' action. Type must match the target field."
    }
  },
  "required": ["action"],
  "additionalProperties": false
}
```

### 1.3 Go Parameters Method

```go
func (t *ConfigureTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "action": map[string]interface{}{
                "type":        "string",
                "enum":        []string{"get", "set", "enable", "disable", "list_providers", "list_channels"},
                "description": "The configuration action to perform",
            },
            "path": map[string]interface{}{
                "type":        "string",
                "description": "Dot-notation path to the config field (e.g., 'providers.openai.api_key'). Required for get/set/enable/disable.",
            },
            "value": map[string]interface{}{
                "type":        []string{"string", "number", "boolean", "array"},
                "description": "The value to set. Required for 'set' action.",
            },
        },
        "required": []string{"action"},
    }
}
```

---

## 2. Actions Specification

### 2.1 Action: `get`

**Purpose:** Read a configuration value. Sensitive fields are redacted.

**Required Parameters:**
- `action`: `"get"`
- `path`: Config path (e.g., `"providers.openai.api_base"`)

**Behavior:**
1. Validate path is in whitelist
2. Check if field is sensitive (write-only)
3. If sensitive: return `[SET]` or `[NOT SET]` indicator
4. If non-sensitive: return actual value
5. Log operation to audit trail

**Response Format:**
```json
{
  "success": true,
  "path": "providers.openai.api_base",
  "value": "https://api.openai.com/v1",
  "type": "string"
}
```

For sensitive fields:
```json
{
  "success": true,
  "path": "providers.openai.api_key",
  "value": "[SET]",
  "sensitive": true,
  "hint": "Value is set but cannot be displayed"
}
```

For unset sensitive fields:
```json
{
  "success": true,
  "path": "providers.openai.api_key",
  "value": "[NOT SET]",
  "sensitive": true
}
```

### 2.2 Action: `set`

**Purpose:** Write a configuration value. All whitelisted fields are writable.

**Required Parameters:**
- `action`: `"set"`
- `path`: Config path
- `value`: New value (type must match field)

**Behavior:**
1. Validate path is in whitelist
2. Validate value type matches field type
3. Validate value against field-specific rules (see Section 3.3)
4. Load current config
5. Apply change
6. Validate config still loads correctly
7. Save config
8. Log operation to audit trail (with value redacted for sensitive fields)

**Response Format:**
```json
{
  "success": true,
  "path": "providers.openai.api_key",
  "message": "Configuration updated successfully",
  "restart_required": false
}
```

For channel changes requiring restart:
```json
{
  "success": true,
  "path": "channels.telegram.token",
  "message": "Configuration updated successfully",
  "restart_required": true,
  "hint": "Restart the gateway for changes to take effect"
}
```

### 2.3 Action: `enable`

**Purpose:** Enable a channel or feature by setting its `enabled` field to `true`.

**Required Parameters:**
- `action`: `"enable"`
- `path`: Path to the section (e.g., `"channels.telegram"`, `"agents.orchestrator"`)

**Behavior:**
1. Validate path points to an enableable section
2. Set `{path}.enabled = true`
3. Save config
4. Log operation to audit trail

**Valid Enable Paths:**
- `channels.{channel_name}` - Any channel
- `agents.orchestrator` - Orchestrator

**Response Format:**
```json
{
  "success": true,
  "path": "channels.telegram",
  "enabled": true,
  "message": "Telegram channel enabled",
  "restart_required": true,
  "hint": "Restart the gateway for changes to take effect"
}
```

### 2.4 Action: `disable`

**Purpose:** Disable a channel or feature by setting its `enabled` field to `false`.

**Required Parameters:**
- `action`: `"disable"`
- `path`: Path to the section (e.g., `"channels.telegram"`)

**Behavior:**
1. Validate path points to a disableable section
2. Set `{path}.enabled = false`
3. Save config
4. Log operation to audit trail

**Response Format:**
```json
{
  "success": true,
  "path": "channels.telegram",
  "enabled": false,
  "message": "Telegram channel disabled",
  "restart_required": true,
  "hint": "Restart the gateway for changes to take effect"
}
```

### 2.5 Action: `list_providers`

**Purpose:** List all LLM providers with their configuration status.

**Required Parameters:**
- `action`: `"list_providers"`

**Behavior:**
1. Iterate through all provider configs
2. For each provider, report:
   - Name
   - Whether API key is set (without revealing value)
   - API base URL (if customized)
   - Proxy (if set)
   - Whether provider is usable (has required config)
3. Log operation to audit trail

**Response Format:**
```json
{
  "success": true,
  "providers": [
    {
      "name": "anthropic",
      "api_key_set": true,
      "api_base": "https://api.anthropic.com",
      "api_base_custom": false,
      "proxy": null,
      "usable": true
    },
    {
      "name": "openai",
      "api_key_set": true,
      "api_base": "https://custom.openai.example.com/v1",
      "api_base_custom": true,
      "proxy": null,
      "usable": true
    },
    {
      "name": "groq",
      "api_key_set": false,
      "api_base": "https://api.groq.com/openai/v1",
      "api_base_custom": false,
      "proxy": null,
      "usable": false
    },
    {
      "name": "ollama",
      "api_key_set": false,
      "api_base": "http://localhost:11434/v1",
      "api_base_custom": false,
      "proxy": null,
      "usable": true,
      "note": "Ollama does not require API key"
    }
  ],
  "active_count": 3,
  "total_count": 10
}
```

### 2.6 Action: `list_channels`

**Purpose:** List all communication channels with their configuration status.

**Required Parameters:**
- `action`: `"list_channels"`

**Behavior:**
1. Iterate through all channel configs
2. For each channel, report:
   - Name
   - Enabled status
   - Whether credentials are set (without revealing values)
   - Allow list (IDs can be shown)
   - Whether channel is usable (enabled + has required credentials)
3. Log operation to audit trail

**Response Format:**
```json
{
  "success": true,
  "channels": [
    {
      "name": "telegram",
      "enabled": true,
      "credentials_set": true,
      "credential_fields": ["token"],
      "allow_from": ["123456789", "987654321"],
      "usable": true
    },
    {
      "name": "discord",
      "enabled": false,
      "credentials_set": true,
      "credential_fields": ["token"],
      "allow_from": [],
      "usable": false,
      "note": "Channel is disabled"
    },
    {
      "name": "slack",
      "enabled": true,
      "credentials_set": false,
      "credential_fields": ["bot_token", "app_token"],
      "allow_from": [],
      "usable": false,
      "note": "Missing required credentials: bot_token, app_token"
    },
    {
      "name": "feishu",
      "enabled": false,
      "credentials_set": false,
      "credential_fields": ["app_id", "app_secret"],
      "allow_from": [],
      "usable": false
    }
  ],
  "active_count": 1,
  "total_count": 9
}
```

---

## 3. Field Whitelist

### 3.1 Providers Section

All 10 providers follow the same field pattern:

| Provider Name | Config Key |
|---------------|------------|
| Anthropic | `anthropic` |
| OpenAI | `openai` |
| OpenRouter | `openrouter` |
| Groq | `groq` |
| Zhipu | `zhipu` |
| VLLM | `vllm` |
| Gemini | `gemini` |
| Nvidia | `nvidia` |
| Moonshot | `moonshot` |
| Ollama | `ollama` |

**Fields per provider:**

| Path Pattern | Type | Policy | Validation | Description |
|--------------|------|--------|------------|-------------|
| `providers.{name}.api_key` | string | **write-only** | Non-empty when set | API key (sensitive) |
| `providers.{name}.api_base` | string | read-write | Valid URL or empty | API base URL |
| `providers.{name}.proxy` | string | read-write | Valid URL or empty | Proxy URL |
| `providers.{name}.auth_method` | string | read-write | `"" | "bearer" | "basic"` | Authentication method |
| `providers.{name}.models` | []string | read-write | Array of strings | Available models list |

### 3.2 Channels Section

Each channel has unique credential fields but common structural fields:

#### 3.2.1 Telegram

| Path | Type | Policy | Validation |
|------|------|--------|------------|
| `channels.telegram.enabled` | bool | read-write | true/false |
| `channels.telegram.token` | string | **write-only** | Non-empty when set |
| `channels.telegram.proxy` | string | read-write | Valid URL or empty |
| `channels.telegram.allow_from` | []string | read-write | Array of user IDs |

#### 3.2.2 Discord

| Path | Type | Policy | Validation |
|------|------|--------|------------|
| `channels.discord.enabled` | bool | read-write | true/false |
| `channels.discord.token` | string | **write-only** | Non-empty when set |
| `channels.discord.allow_from` | []string | read-write | Array of user IDs |

#### 3.2.3 Slack

| Path | Type | Policy | Validation |
|------|------|--------|------------|
| `channels.slack.enabled` | bool | read-write | true/false |
| `channels.slack.bot_token` | string | **write-only** | Starts with `xoxb-` when set |
| `channels.slack.app_token` | string | **write-only** | Starts with `xapp-` when set |
| `channels.slack.allow_from` | []string | read-write | Array of user IDs |

#### 3.2.4 WhatsApp

| Path | Type | Policy | Validation |
|------|------|--------|------------|
| `channels.whatsapp.enabled` | bool | read-write | true/false |
| `channels.whatsapp.bridge_url` | string | read-write | Valid WebSocket URL |
| `channels.whatsapp.allow_from` | []string | read-write | Array of phone numbers |

#### 3.2.5 Signal

| Path | Type | Policy | Validation |
|------|------|--------|------------|
| `channels.signal.enabled` | bool | read-write | true/false |
| `channels.signal.phone_number` | string | read-write | Phone number format |
| `channels.signal.allow_from` | []string | read-write | Array of phone numbers |

#### 3.2.6 Feishu

| Path | Type | Policy | Validation |
|------|------|--------|------------|
| `channels.feishu.enabled` | bool | read-write | true/false |
| `channels.feishu.app_id` | string | **write-only** | Non-empty when set |
| `channels.feishu.app_secret` | string | **write-only** | Non-empty when set |
| `channels.feishu.encrypt_key` | string | **write-only** | Non-empty when set |
| `channels.feishu.verification_token` | string | **write-only** | Non-empty when set |
| `channels.feishu.allow_from` | []string | read-write | Array of user IDs |

#### 3.2.7 QQ

| Path | Type | Policy | Validation |
|------|------|--------|------------|
| `channels.qq.enabled` | bool | read-write | true/false |
| `channels.qq.app_id` | string | **write-only** | Non-empty when set |
| `channels.qq.app_secret` | string | **write-only** | Non-empty when set |
| `channels.qq.allow_from` | []string | read-write | Array of user IDs |

#### 3.2.8 DingTalk

| Path | Type | Policy | Validation |
|------|------|--------|------------|
| `channels.dingtalk.enabled` | bool | read-write | true/false |
| `channels.dingtalk.client_id` | string | **write-only** | Non-empty when set |
| `channels.dingtalk.client_secret` | string | **write-only** | Non-empty when set |
| `channels.dingtalk.allow_from` | []string | read-write | Array of user IDs |

#### 3.2.9 MaixCam

| Path | Type | Policy | Validation |
|------|------|--------|------------|
| `channels.maixcam.enabled` | bool | read-write | true/false |
| `channels.maixcam.host` | string | read-write | Valid IP or hostname |
| `channels.maixcam.port` | int | read-write | 1-65535 |
| `channels.maixcam.allow_from` | []string | read-write | Array of identifiers |

### 3.3 Agents Section

| Path | Type | Policy | Validation |
|------|------|--------|------------|
| `agents.defaults.provider` | string | read-write | Valid provider name or empty |
| `agents.defaults.model` | string | read-write | Non-empty string |
| `agents.defaults.max_tokens` | int | read-write | 1-1000000 |
| `agents.defaults.temperature` | float | read-write | 0.0-2.0 |
| `agents.defaults.max_tool_iterations` | int | read-write | 1-100 |
| `agents.orchestrator.enabled` | bool | read-write | true/false |
| `agents.orchestrator.provider` | string | read-write | Valid provider name or empty |
| `agents.orchestrator.model` | string | read-write | Non-empty string when enabled |
| `agents.orchestrator.max_tokens` | int | read-write | 1-1000000 |
| `agents.orchestrator.temperature` | float | read-write | 0.0-2.0 |
| `agents.orchestrator.max_delegation_retries` | int | read-write | 0-10 |
| `agents.orchestrator.fallback_to_default` | bool | read-write | true/false |

### 3.4 Tools Section (Limited)

| Path | Type | Policy | Validation |
|------|------|--------|------------|
| `tools.web.search.api_key` | string | **write-only** | Non-empty when set |
| `tools.web.search.max_results` | int | read-write | 1-100 |
| `tools.email.enabled` | bool | read-write | true/false |
| `tools.email.host` | string | read-write | Valid hostname |
| `tools.email.port` | int | read-write | 1-65535 |
| `tools.email.username` | string | read-write | Valid email or username |
| `tools.email.password` | string | **write-only** | Non-empty when set |
| `tools.email.from` | string | read-write | Valid email address |
| `tools.email.to` | string | read-write | Valid email address |

---

## 4. Response Formats

### 4.1 Success Response Structure

All successful responses follow this base structure:

```go
type ConfigureResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
    // Action-specific fields added below
}
```

### 4.2 Error Response Structure

```json
{
  "success": false,
  "error": "error_code",
  "message": "Human-readable error description",
  "path": "providers.invalid.field",
  "hint": "Optional suggestion for how to fix the error"
}
```

### 4.3 Response Serialization

Responses are serialized to JSON strings for return from `Execute()`:

```go
func (t *ConfigureTool) formatResponse(data interface{}) (string, error) {
    bytes, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}
```

---

## 5. Error Scenarios

### 5.1 Error Codes and Messages

| Error Code | HTTP Analog | Message | Cause |
|------------|-------------|---------|-------|
| `invalid_action` | 400 | "Unknown action: {action}" | Action not in enum |
| `missing_path` | 400 | "Path is required for {action} action" | No path provided for get/set/enable/disable |
| `missing_value` | 400 | "Value is required for set action" | No value provided for set |
| `path_not_allowed` | 403 | "Field '{path}' is not configurable" | Path not in whitelist |
| `sensitive_read` | 403 | "Field '{path}' is write-only and cannot be read" | Attempted to read sensitive field directly (note: get action returns [SET]/[NOT SET] instead of this error) |
| `invalid_value_type` | 400 | "Expected {expected_type} for '{path}', got {actual_type}" | Type mismatch |
| `validation_failed` | 400 | "Validation failed: {details}" | Value fails field-specific validation |
| `config_load_error` | 500 | "Failed to load configuration: {details}" | Config file read/parse error |
| `config_save_error` | 500 | "Failed to save configuration: {details}" | Config file write error |
| `config_invalid` | 500 | "Configuration invalid after change: {details}" | Modified config fails to load |
| `user_context_missing` | 500 | "User context not set" | SetUserContext not called |
| `path_not_enableable` | 400 | "Path '{path}' does not support enable/disable" | Enable/disable on non-toggleable field |

### 5.2 Validation Error Details

**URL Validation:**
```json
{
  "success": false,
  "error": "validation_failed",
  "message": "Validation failed: invalid URL format",
  "path": "providers.openai.api_base",
  "value": "not-a-url",
  "hint": "URL must start with http:// or https://"
}
```

**Range Validation:**
```json
{
  "success": false,
  "error": "validation_failed",
  "message": "Validation failed: value out of range",
  "path": "agents.defaults.temperature",
  "value": 5.0,
  "hint": "Temperature must be between 0.0 and 2.0"
}
```

**Port Validation:**
```json
{
  "success": false,
  "error": "validation_failed",
  "message": "Validation failed: invalid port number",
  "path": "channels.maixcam.port",
  "value": 99999,
  "hint": "Port must be between 1 and 65535"
}
```

---

## 6. Security Requirements

### 6.1 Sensitive Field Patterns

Fields matching these patterns are classified as **sensitive** and are **write-only**:

```go
var sensitiveFieldPatterns = []string{
    "api_key",
    "apikey",
    "token",
    "bot_token",
    "app_token",
    "access_token",
    "refresh_token",
    "password",
    "secret",
    "app_secret",
    "client_secret",
    "signing_secret",
    "private_key",
    "encrypt_key",
    "verification_token",
}
```

**Matching Logic:**
- Case-insensitive substring match against the final field name
- Example: `providers.openai.api_key` -> matches "api_key"
- Example: `channels.slack.bot_token` -> matches "bot_token"

### 6.2 Redaction Requirements

**MUST redact:**
1. Actual sensitive field values in all responses
2. Sensitive values in audit log arguments
3. Sensitive values in error messages

**MUST NOT redact:**
1. Field names/paths (these are always visible)
2. Non-sensitive field values
3. `allow_from` lists (user IDs are not credentials)
4. Enable/disable status

**Redaction Format:**
- `[SET]` - Sensitive field has a non-empty value
- `[NOT SET]` - Sensitive field is empty
- `[REDACTED]` - Used in audit logs for argument sanitization

### 6.3 Audit Requirements

**All configure tool calls MUST be audited:**

1. Add `configure` to `RestrictedTools` in `pkg/tools/audit.go`:
   ```go
   var RestrictedTools = map[string]bool{
       "exec":        true,
       "spawn":       true,
       "email":       true,
       "write_file":  true,
       "edit_file":   true,
       "append_file": true,
       "web_fetch":   true,
       "configure":   true, // NEW
   }
   ```

2. **Audit Log Entry Fields:**
   | Field | Source | Example |
   |-------|--------|---------|
   | `timestamp` | Current time | `2026-02-25T10:30:00Z` |
   | `user_id` | From user context | `42` |
   | `username` | From user context | `john@example.com` |
   | `tool` | Tool name | `"configure"` |
   | `arguments` | Sanitized args | `{"action":"set","path":"providers.openai.api_key","value":"[REDACTED]"}` |
   | `success` | Operation result | `true` |
   | `error` | Error message if failed | `""` |
   | `duration_ms` | Execution time | `45` |

3. **Argument Sanitization:**
   The existing `sanitizeArguments` function in `audit.go` handles this:
   ```go
   // Before logging:
   args := map[string]interface{}{
       "action": "set",
       "path":   "providers.openai.api_key",
       "value":  "sk-actual-secret-key", // WILL BE REDACTED
   }
   // After sanitization:
   sanitized := map[string]interface{}{
       "action": "set",
       "path":   "providers.openai.api_key",
       "value":  "[REDACTED]",
   }
   ```

### 6.4 Path Validation Requirements

**MUST validate:**
1. Path is not empty
2. Path matches a whitelisted pattern (no wildcards in user input)
3. No path traversal attempts (`..`, leading `/`)
4. Path segments only contain alphanumeric, underscore, dot

**Regex Pattern for Path Validation:**
```go
var validPathPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)
```

### 6.5 User Isolation Requirements

1. **Config isolation:** Each user's config is at `~/.MakoClaw/users/{uuid}/config.json`
2. **UUID validation:** Only accept UUIDs from authenticated context, never from user input
3. **No cross-user access:** Tool cannot access other users' configs
4. **Global config read-only:** Tool only modifies user-specific config, never global

---

## 7. New Interface: UserConfigTool

### 7.1 Interface Definition

```go
// UserConfigTool is for tools that need access to user config files.
// Unlike UserAwareTool (which only provides userID for data filtering),
// this interface provides the userUUID needed to locate config files.
type UserConfigTool interface {
    Tool
    SetUserContext(userID int64, userUUID string)
}
```

### 7.2 Implementation in ConfigureTool

```go
type ConfigureTool struct {
    userID   int64
    userUUID string
}

func (t *ConfigureTool) SetUserContext(userID int64, userUUID string) {
    t.userID = userID
    t.userUUID = userUUID
}
```

### 7.3 Agent Loop Integration

The agent loop needs to propagate `userUUID` to tools implementing `UserConfigTool`:

```go
// In pkg/agent/loop.go
func (al *AgentLoop) updateToolsUserContext(userID int64, userUUID string) {
    for _, tool := range al.tools {
        // Existing UserAwareTool support
        if uat, ok := tool.(UserAwareTool); ok {
            uat.SetUserID(userID)
        }
        // New UserConfigTool support
        if uct, ok := tool.(UserConfigTool); ok {
            uct.SetUserContext(userID, userUUID)
        }
    }
}
```

---

## 8. Validation Rules

### 8.1 Type Validation

| Target Type | Go Type | Validation |
|-------------|---------|------------|
| `string` | `string` | Accept string values |
| `bool` | `bool` | Accept `true`/`false`, or strings `"true"`/`"false"` |
| `int` | `int` | Accept integers, reject floats with fractional parts |
| `float` | `float64` | Accept any numeric value |
| `[]string` | `[]string` | Accept arrays of strings, or comma-separated string |

### 8.2 Field-Specific Validation

**URL Fields:**
- Must be empty OR start with `http://` or `https://`
- WebSocket URLs must start with `ws://` or `wss://`

**Port Fields:**
- Integer between 1 and 65535

**Temperature:**
- Float between 0.0 and 2.0

**Max Tokens:**
- Integer between 1 and 1,000,000

**Max Tool Iterations:**
- Integer between 1 and 100

**Provider Names:**
- Must be one of: `anthropic`, `openai`, `openrouter`, `groq`, `zhipu`, `vllm`, `gemini`, `nvidia`, `moonshot`, `ollama`
- Or empty string (to unset)

**Slack Tokens:**
- `bot_token` must start with `xoxb-` or be empty
- `app_token` must start with `xapp-` or be empty

**Email Address:**
- Basic format validation: contains `@` and `.`

---

## 9. Implementation Checklist

### 9.1 New Files

| File | Description |
|------|-------------|
| `pkg/tools/configure.go` | Main tool implementation |
| `pkg/config/redact.go` | Sensitive field redaction utilities |
| `pkg/config/validate.go` | Field validation rules and functions |

### 9.2 Modified Files

| File | Changes |
|------|---------|
| `pkg/tools/base.go` | Add `UserConfigTool` interface |
| `pkg/tools/audit.go` | Add `"configure"` to `RestrictedTools` |
| `pkg/agent/loop.go` | Register tool, propagate user context |

### 9.3 Test Coverage Requirements

| Test Category | Required Tests |
|---------------|----------------|
| Actions | All 6 actions with valid inputs |
| Whitelist | Allowed paths pass, disallowed paths fail |
| Sensitive fields | Write succeeds, read returns status indicator |
| Validation | All field types, edge cases, invalid values |
| Error handling | All error codes with appropriate messages |
| User isolation | Cannot access other users' configs |
| Audit logging | All operations logged with sanitized args |

---

## 10. Related Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Exploration | `openspec/changes/configure-tool/exploration.md` | Complete |
| Proposal | `openspec/changes/configure-tool/proposal.md` | Complete |
| Specification | `openspec/changes/configure-tool/specs.md` | This document |
| Design | `openspec/changes/configure-tool/design.md` | Pending |
| Tasks | `openspec/changes/configure-tool/tasks.md` | Pending |

---

## Appendix A: Complete Whitelist Reference

```go
// FieldPolicy defines whether a field is readable, writable, or both
type FieldPolicy int

const (
    PolicyReadWrite FieldPolicy = iota // Can read and write
    PolicyWriteOnly                    // Can write, but read returns [SET]/[NOT SET]
)

// FieldSpec defines a whitelisted configuration field
type FieldSpec struct {
    Path       string
    Type       string      // "string", "bool", "int", "float", "[]string"
    Policy     FieldPolicy
    Validation func(value interface{}) error
}

// Complete whitelist - see Section 3 for full details
var configWhitelist = map[string]FieldSpec{
    // Providers (10 providers x 5 fields = 50 entries)
    "providers.anthropic.api_key":      {Type: "string", Policy: PolicyWriteOnly},
    "providers.anthropic.api_base":     {Type: "string", Policy: PolicyReadWrite, Validation: validateURL},
    "providers.anthropic.proxy":        {Type: "string", Policy: PolicyReadWrite, Validation: validateURL},
    "providers.anthropic.auth_method":  {Type: "string", Policy: PolicyReadWrite, Validation: validateAuthMethod},
    "providers.anthropic.models":       {Type: "[]string", Policy: PolicyReadWrite},
    // ... (repeat for all 10 providers)

    // Channels (9 channels with varying fields)
    "channels.telegram.enabled":        {Type: "bool", Policy: PolicyReadWrite},
    "channels.telegram.token":          {Type: "string", Policy: PolicyWriteOnly},
    "channels.telegram.proxy":          {Type: "string", Policy: PolicyReadWrite, Validation: validateURL},
    "channels.telegram.allow_from":     {Type: "[]string", Policy: PolicyReadWrite},
    // ... (see Section 3.2 for all channel fields)

    // Agents
    "agents.defaults.provider":         {Type: "string", Policy: PolicyReadWrite, Validation: validateProviderName},
    "agents.defaults.model":            {Type: "string", Policy: PolicyReadWrite},
    "agents.defaults.max_tokens":       {Type: "int", Policy: PolicyReadWrite, Validation: validateMaxTokens},
    "agents.defaults.temperature":      {Type: "float", Policy: PolicyReadWrite, Validation: validateTemperature},
    "agents.defaults.max_tool_iterations": {Type: "int", Policy: PolicyReadWrite, Validation: validateMaxIterations},
    "agents.orchestrator.enabled":      {Type: "bool", Policy: PolicyReadWrite},
    // ... (see Section 3.3 for all agent fields)

    // Tools
    "tools.web.search.api_key":         {Type: "string", Policy: PolicyWriteOnly},
    "tools.web.search.max_results":     {Type: "int", Policy: PolicyReadWrite, Validation: validateMaxResults},
    "tools.email.enabled":              {Type: "bool", Policy: PolicyReadWrite},
    // ... (see Section 3.4 for all tool fields)
}
```

---

## Appendix B: Example Tool Invocations

### B.1 Setting an API Key

**Agent Prompt:**
```
User: "Set my OpenAI API key to sk-proj-abc123def456"
```

**Tool Call:**
```json
{
  "name": "configure",
  "arguments": {
    "action": "set",
    "path": "providers.openai.api_key",
    "value": "sk-proj-abc123def456"
  }
}
```

**Tool Response:**
```json
{
  "success": true,
  "path": "providers.openai.api_key",
  "message": "Configuration updated successfully",
  "restart_required": false
}
```

### B.2 Checking Provider Status

**Agent Prompt:**
```
User: "What API providers do I have set up?"
```

**Tool Call:**
```json
{
  "name": "configure",
  "arguments": {
    "action": "list_providers"
  }
}
```

### B.3 Enabling a Channel

**Agent Prompt:**
```
User: "Turn on Telegram"
```

**Tool Call:**
```json
{
  "name": "configure",
  "arguments": {
    "action": "enable",
    "path": "channels.telegram"
  }
}
```

### B.4 Reading a Non-Sensitive Value

**Agent Prompt:**
```
User: "What's my default LLM model?"
```

**Tool Call:**
```json
{
  "name": "configure",
  "arguments": {
    "action": "get",
    "path": "agents.defaults.model"
  }
}
```

**Tool Response:**
```json
{
  "success": true,
  "path": "agents.defaults.model",
  "value": "claude-3-opus-20240229",
  "type": "string"
}
```
