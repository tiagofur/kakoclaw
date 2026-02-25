# Implementation Tasks: Configure Tool

**Change ID:** configure-tool
**Status:** Draft
**Author:** SDD Sub-Agent
**Created:** 2026-02-25
**Based on:** [specs.md](specs.md), [design.md](design.md)

---

## Overview

This document breaks down the ConfigureTool implementation into small, focused tasks organized by phase. Each task is designed to be completable in a single focused session.

**Estimated Total Effort:** 8-12 hours across 5 phases

---

## Phase 1: Foundation (Core Infrastructure)

**Goal:** Establish the base structure and interfaces needed for the configure tool.

### Task 1.1: Add UserConfigTool Interface to base.go

**File:** `pkg/tools/base.go`

**Description:** Add the new `UserConfigTool` interface that extends `Tool` with user context capabilities for config file access.

**Changes:**
```go
// UserConfigTool is for tools that need access to user config files.
// Unlike UserAwareTool (which only provides userID for data filtering),
// this interface provides the userUUID needed to locate config files.
type UserConfigTool interface {
    Tool
    SetUserContext(userID int64, userUUID string)
}
```

**Acceptance Criteria:**
- [x] Interface added after existing `UserAwareTool` interface
- [x] Documentation comment explains difference from `UserAwareTool`
- [x] File compiles without errors

**Estimated Time:** 10 minutes

---

### Task 1.2: Create ConfigureTool Skeleton

**File:** `pkg/tools/configure.go` (NEW)

**Description:** Create the basic ConfigureTool struct implementing `Tool` and `UserConfigTool` interfaces with stub methods.

**Changes:**
- Create new file with package declaration and imports
- Define `ConfigureTool` struct with `userID`, `userUUID`, `mu sync.RWMutex`
- Implement `Name()`, `Description()`, `Parameters()`
- Implement `SetUserContext()`
- Implement `Execute()` with action routing skeleton (return errors for unimplemented actions)
- Add interface compliance check: `var _ UserConfigTool = (*ConfigureTool)(nil)`

**Acceptance Criteria:**
- [x] File compiles without errors
- [x] `NewConfigureTool()` constructor works
- [x] All interface methods have signatures (can return placeholders)
- [x] Execute() routes to action handlers (can be stubs)

**Estimated Time:** 30 minutes

---

### Task 1.3: Define Field Whitelist Structure

**File:** `pkg/tools/configure.go`

**Description:** Add the field whitelist data structures and initial field definitions.

**Changes:**
```go
// FieldPolicy defines read/write permissions for a config field
type FieldPolicy struct {
    Type      FieldType
    Readable  bool
    Writable  bool
    Validator ValidatorFn
}

type FieldType int

const (
    TypeString FieldType = iota
    TypeInt
    TypeFloat
    TypeBool
    TypeStringSlice
)

type ValidatorFn func(value interface{}) error

// fieldWhitelist defines all configurable fields and their policies
var fieldWhitelist = map[string]FieldPolicy{
    // Provider fields...
    // Channel fields...
    // Agent fields...
    // Tool fields...
}

var providerNames = []string{...}
var channelNames = []string{...}
```

**Acceptance Criteria:**
- [x] All types defined (FieldPolicy, FieldType, ValidatorFn)
- [x] fieldWhitelist contains all paths from specs.md Section 3
- [x] providerNames contains all 10 providers
- [x] channelNames contains all 9 channels
- [x] Wildcard patterns used for providers.* and channels.*

**Estimated Time:** 45 minutes

---

### Task 1.4: Define Sensitive Field Patterns

**File:** `pkg/tools/configure.go`

**Description:** Add the sensitive field detection patterns and helper function.

**Changes:**
```go
// sensitiveFieldPatterns lists field name patterns that contain secrets
var sensitiveFieldPatterns = []string{
    "api_key", "apikey",
    "token", "bot_token", "app_token", "access_token", "refresh_token",
    "password",
    "secret", "app_secret", "client_secret", "signing_secret",
    "private_key", "encrypt_key", "verification_token",
}

// isSensitiveField checks if a field name matches sensitive patterns
func isSensitiveField(fieldName string) bool {
    lower := strings.ToLower(fieldName)
    for _, pattern := range sensitiveFieldPatterns {
        if strings.Contains(lower, pattern) {
            return true
        }
    }
    return false
}
```

**Acceptance Criteria:**
- [x] All patterns from specs.md Section 6.1 included
- [x] `isSensitiveField()` returns true for sensitive fields
- [x] Case-insensitive matching works

**Estimated Time:** 15 minutes

---

## Phase 2: Config Helpers

**Goal:** Create reusable utility functions for redaction and validation in the config package.

### Task 2.1: Create Redaction Helpers

**File:** `pkg/config/redact.go` (NEW)

**Description:** Create utility functions for sensitive field redaction that can be reused across the codebase.

**Changes:**
```go
package config

// SensitiveFieldPatterns lists field name patterns that contain secrets
var SensitiveFieldPatterns = []string{...}

// IsSensitiveField checks if a field name matches sensitive patterns
func IsSensitiveField(fieldName string) bool

// RedactValue returns a redacted representation of a sensitive value
// Returns "[NOT SET]" for empty, "[REDACTED]" for short values, "****{last4}" for longer
func RedactValue(value string) string

// RedactConfigPath extracts the field name from a path and checks sensitivity
func RedactConfigPath(path string) bool
```

**Acceptance Criteria:**
- [x] `IsSensitiveField("api_key")` returns true
- [x] `IsSensitiveField("model")` returns false
- [x] `RedactValue("")` returns "[NOT SET]"
- [x] `RedactValue("abc")` returns "[REDACTED]"
- [x] `RedactValue("sk-proj-abc12345")` returns "sk-****2345" (first 3 + **** + last 4)
- [x] `RedactConfigPath("providers.openai.api_key")` returns true

**Estimated Time:** 30 minutes

---

### Task 2.2: Create Validation Helpers

**File:** `pkg/config/validate.go` (NEW)

**Description:** Create field validation functions for common types.

**Changes:**
```go
package config

// ValidateURL ensures a string is a valid URL (or empty)
func ValidateURL(value interface{}) error

// ValidatePort ensures a value is a valid port number (1-65535)
func ValidatePort(value interface{}) error

// ValidatePositiveInt ensures a value is a non-negative integer
func ValidatePositiveInt(value interface{}) error

// ValidateTemperature ensures temperature is in valid range [0.0, 2.0]
func ValidateTemperature(value interface{}) error

// ValidateEmail provides basic email format validation
func ValidateEmail(value interface{}) error

// ValidateProviderName ensures the provider name is known
func ValidateProviderName(name string) error

// ValidateChannelName ensures the channel name is known
func ValidateChannelName(name string) error

// ValidateMaxTokens ensures max_tokens is in range [1, 1000000]
func ValidateMaxTokens(value interface{}) error

// ValidateMaxIterations ensures max_tool_iterations is in range [1, 100]
func ValidateMaxIterations(value interface{}) error
```

**Acceptance Criteria:**
- [x] URL validation accepts http://, https://, empty
- [x] URL validation accepts ws://, wss:// for WebSocket
- [x] Port validation rejects <1 and >65535
- [x] Temperature validation rejects <0 and >2.0
- [x] Email validation requires @ and .
- [x] Provider/channel name validation rejects unknown names

**Estimated Time:** 45 minutes

---

### Task 2.3: Write Tests for Redaction Helpers

**File:** `pkg/config/redact_test.go` (NEW)

**Description:** Unit tests for redaction helper functions.

**Test Cases:**
- `TestIsSensitiveField` - Test various field names
- `TestRedactValue` - Test empty, short, and long values
- `TestRedactConfigPath` - Test full config paths

**Acceptance Criteria:**
- [x] All edge cases covered
- [x] Tests pass with `go test ./pkg/config -v -run Redact`

**Estimated Time:** 20 minutes

---

### Task 2.4: Write Tests for Validation Helpers

**File:** `pkg/config/validate_test.go` (NEW)

**Description:** Unit tests for validation helper functions.

**Test Cases:**
- `TestValidateURL` - Valid URLs, invalid URLs, empty
- `TestValidatePort` - Valid ports, edge cases, invalid
- `TestValidateTemperature` - Valid range, edge cases
- `TestValidateEmail` - Valid emails, invalid formats
- `TestValidateProviderName` - Known and unknown providers
- `TestValidateChannelName` - Known and unknown channels

**Acceptance Criteria:**
- [x] All validators have test coverage
- [x] Tests pass with `go test ./pkg/config -v -run Validate`

**Estimated Time:** 30 minutes

---

## Phase 3: Tool Actions

**Goal:** Implement all six action handlers for the ConfigureTool.

### Task 3.1: Implement Path Resolution Helper

**File:** `pkg/tools/configure.go`

**Description:** Implement the helper function to resolve config paths against the whitelist, handling wildcards.

**Changes:**
```go
// resolveFieldPolicy looks up the policy for a given path, resolving wildcards
func (t *ConfigureTool) resolveFieldPolicy(path string) (*FieldPolicy, error)
```

**Logic:**
1. Direct lookup in whitelist
2. If not found, try wildcard: replace second segment with `*`
3. Validate the specific provider/channel name
4. Return error if path not in whitelist

**Acceptance Criteria:**
- [x] "providers.openai.api_key" resolves to providers.*.api_key policy
- [x] "channels.telegram.enabled" resolves to channels.*.enabled policy
- [x] "agents.defaults.model" resolves directly
- [x] "some.invalid.path" returns error

**Estimated Time:** 30 minutes

---

### Task 3.2: Implement Field Value Accessors

**File:** `pkg/tools/configure.go`

**Description:** Implement helpers to get and set config field values using reflection.

**Changes:**
```go
// getFieldValue retrieves a field value from config using reflection
func (t *ConfigureTool) getFieldValue(cfg *config.Config, path string) (interface{}, error)

// setFieldValue sets a field value in config using reflection
func (t *ConfigureTool) setFieldValue(cfg *config.Config, path string, value interface{}) error

// validateValueType ensures the value type matches the expected field type
func (t *ConfigureTool) validateValueType(value interface{}, expectedType FieldType) error
```

**Acceptance Criteria:**
- [x] Can read nested struct fields via path
- [x] Can write nested struct fields via path
- [x] Type conversions handle JSON number -> int/float
- [x] FlexibleStringSlice handled for allow_from fields

**Estimated Time:** 60 minutes

---

### Task 3.3: Implement `get` Action

**File:** `pkg/tools/configure.go`

**Description:** Implement the get action to read configuration values with redaction for sensitive fields.

**Changes:**
```go
func (t *ConfigureTool) handleGet(ctx context.Context, args map[string]interface{}, userUUID string) (string, error)
```

**Logic:**
1. Extract and validate path parameter
2. Resolve field policy from whitelist
3. Check if field is readable (not write-only)
4. Load user config via `config.LoadConfigForUser(userUUID)`
5. Get field value via reflection
6. If sensitive, return "[SET]" or "[NOT SET]"
7. Format and return JSON response

**Acceptance Criteria:**
- [x] Non-sensitive fields return actual values
- [x] Sensitive fields return "[SET]" or "[NOT SET]"
- [x] Non-whitelisted paths return error
- [x] Missing path parameter returns error

**Estimated Time:** 45 minutes

---

### Task 3.4: Implement `set` Action

**File:** `pkg/tools/configure.go`

**Description:** Implement the set action to write configuration values with validation.

**Changes:**
```go
func (t *ConfigureTool) handleSet(ctx context.Context, args map[string]interface{}, userUUID string, userID int64) (string, error)
```

**Logic:**
1. Extract and validate path and value parameters
2. Resolve field policy from whitelist
3. Check if field is writable
4. Validate value type matches expected
5. Run field-specific validator if present
6. Load current config
7. Apply change via reflection
8. Save config via `config.SaveConfigForUser(userUUID, cfg)`
9. Verify config loads successfully after save
10. Return success message (never include actual value for sensitive)

**Acceptance Criteria:**
- [x] Valid values are saved to config file
- [x] Invalid values rejected with helpful error
- [x] Non-whitelisted paths rejected
- [x] Success message never includes sensitive values
- [x] Config verified loadable after save

**Estimated Time:** 60 minutes

---

### Task 3.5: Implement `enable` Action

**File:** `pkg/tools/configure.go`

**Description:** Implement the enable action as a convenience wrapper around set.

**Changes:**
```go
func (t *ConfigureTool) handleEnable(ctx context.Context, args map[string]interface{}, userUUID string, userID int64) (string, error)
```

**Logic:**
1. Validate path format: `channels.{name}` or `agents.orchestrator`
2. Validate channel/agent name
3. Append `.enabled` to path
4. Delegate to handleSet with value=true
5. Return success with restart hint for channels

**Acceptance Criteria:**
- [x] "channels.telegram" enables Telegram
- [x] "agents.orchestrator" enables orchestrator
- [x] Invalid path format returns error
- [x] Response includes restart hint for channels

**Estimated Time:** 20 minutes

---

### Task 3.6: Implement `disable` Action

**File:** `pkg/tools/configure.go`

**Description:** Implement the disable action as a convenience wrapper around set.

**Changes:**
```go
func (t *ConfigureTool) handleDisable(ctx context.Context, args map[string]interface{}, userUUID string, userID int64) (string, error)
```

**Logic:**
1. Same as enable, but set value=false

**Acceptance Criteria:**
- [x] "channels.telegram" disables Telegram
- [x] Response includes restart hint

**Estimated Time:** 10 minutes

---

### Task 3.7: Implement `list_providers` Action

**File:** `pkg/tools/configure.go`

**Description:** Implement the list_providers action to show all providers with their configuration status.

**Changes:**
```go
func (t *ConfigureTool) handleListProviders(ctx context.Context, userUUID string) (string, error)
```

**Logic:**
1. Load user config
2. Iterate through all provider configs
3. For each provider, build status object:
   - name
   - api_key_set (bool from RedactValue check)
   - api_base (actual value or "default")
   - api_base_custom (bool)
   - proxy (if set)
   - usable (has required config)
4. Format as JSON response

**Acceptance Criteria:**
- [x] All 10 providers listed
- [x] API key status shown without actual value
- [x] Custom API base indicated
- [x] Active/total counts included

**Estimated Time:** 45 minutes

---

### Task 3.8: Implement `list_channels` Action

**File:** `pkg/tools/configure.go`

**Description:** Implement the list_channels action to show all channels with their configuration status.

**Changes:**
```go
func (t *ConfigureTool) handleListChannels(ctx context.Context, userUUID string) (string, error)
```

**Logic:**
1. Load user config
2. Iterate through all channel configs
3. For each channel, build status object:
   - name
   - enabled
   - credentials_set (check required credential fields)
   - credential_fields (list of required fields)
   - allow_from (actual values - these are not sensitive)
   - usable (enabled + credentials set)
   - note (if unusable, explain why)
4. Format as JSON response

**Acceptance Criteria:**
- [x] All 9 channels listed
- [x] Enabled status shown
- [x] Credential status shown without actual values
- [x] Allow list shown (not sensitive)
- [x] Active/total counts included

**Estimated Time:** 45 minutes

---

## Phase 4: Integration

**Goal:** Wire the ConfigureTool into the agent loop and audit system.

### Task 4.1: Add Config User Save/Load Functions

**File:** `pkg/config/config.go` (or appropriate file)

**Description:** Ensure `LoadConfigForUser` and `SaveConfigForUser` functions exist and work correctly.

**Check/Add:**
```go
// LoadConfigForUser loads configuration for a specific user
func LoadConfigForUser(userUUID string) (*Config, error)

// SaveConfigForUser saves configuration for a specific user
func SaveConfigForUser(userUUID string, cfg *Config) error

// GetUserConfigPath returns the path to a user's config file
func GetUserConfigPath(userUUID string) string
```

**Acceptance Criteria:**
- [x] Functions exist or are created
- [x] User config path is `~/.MakoClaw/users/{uuid}/config.json`
- [x] Save creates parent directories if needed
- [x] Load falls back to global config if user config doesn't exist

**Note:** These functions already exist in `pkg/config/config.go`.

**Estimated Time:** 30 minutes

---

### Task 4.2: Update Agent Loop for UserConfigTool

**File:** `pkg/agent/loop.go`

**Description:** Add support for propagating user context to UserConfigTool implementations.

**Changes:**
```go
// updateToolsUserConfig propagates user context to UserConfigTool implementations
func (al *AgentLoop) updateToolsUserConfig(userID int64, userUUID string) {
    if al.tools == nil {
        return
    }
    al.tools.ForEach(func(t tools.Tool) {
        if uct, ok := t.(tools.UserConfigTool); ok {
            uct.SetUserContext(userID, userUUID)
        }
    })
}
```

Also update `SetUserForAgent` to call `updateToolsUserConfig`:
```go
func (al *AgentLoop) SetUserForAgent(userUUID string, userID int64) {
    // ... existing code ...
    al.updateToolsUser(userID)
    al.updateToolsUserConfig(userID, userUUID) // NEW
}
```

**Acceptance Criteria:**
- [x] `updateToolsUserConfig` method added
- [x] Called in `SetUserForAgent`
- [x] Called in `InitToolsWithStorage` (if applicable)

**Estimated Time:** 20 minutes

---

### Task 4.3: Register ConfigureTool in Agent Loop

**File:** `pkg/agent/loop.go`

**Description:** Register the ConfigureTool in the tool registry during agent loop initialization.

**Changes:**
Add to `NewAgentLoop` or `InitToolsWithStorage`:
```go
// Register configure tool
configureTool := tools.NewConfigureTool()
toolsRegistry.Register(configureTool)
```

**Acceptance Criteria:**
- [x] Tool registered alongside other tools
- [x] Tool appears in tool list when agent runs
- [x] `go build` succeeds

**Estimated Time:** 15 minutes

---

### Task 4.4: Add to RestrictedTools for Audit Logging

**File:** `pkg/tools/audit.go`

**Description:** Add "configure" to the RestrictedTools map so all config changes are audited.

**Changes:**
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

**Acceptance Criteria:**
- [x] "configure" added to RestrictedTools
- [x] `IsRestrictedTool("configure")` returns true

**Estimated Time:** 5 minutes

---

### Task 4.5: Update Audit Sanitization for Configure Values

**File:** `pkg/tools/audit.go`

**Description:** Ensure the `sanitizeArguments` function handles the "value" key in configure tool arguments when the path indicates a sensitive field.

**Check/Update:**
The existing `sanitizeArguments` already redacts based on key names like "password", "token", etc. However, for the configure tool, the sensitive value is in the "value" key, and sensitivity depends on the "path" key.

**Changes:**
Add configure-specific handling:
```go
func sanitizeArguments(args map[string]interface{}) map[string]interface{} {
    // ... existing code ...

    // Special handling for configure tool - check if path indicates sensitive field
    if path, ok := args["path"].(string); ok {
        if value, exists := args["value"]; exists {
            // Check if the path ends with a sensitive field name
            parts := strings.Split(path, ".")
            if len(parts) > 0 {
                fieldName := parts[len(parts)-1]
                if isSensitiveFieldName(fieldName) {
                    sanitized["value"] = "[REDACTED]"
                }
            }
        }
    }

    return sanitized
}

func isSensitiveFieldName(name string) bool {
    sensitivePatterns := []string{"api_key", "apikey", "token", "password", "secret", "private_key"}
    nameLower := strings.ToLower(name)
    for _, pattern := range sensitivePatterns {
        if strings.Contains(nameLower, pattern) {
            return true
        }
    }
    return false
}
```

**Acceptance Criteria:**
- [x] `configure` args with sensitive paths have value redacted
- [x] Non-sensitive configure values are logged normally
- [x] Existing sanitization behavior preserved

**Estimated Time:** 30 minutes

---

## Phase 5: Testing

**Goal:** Comprehensive test coverage for the ConfigureTool.

### Task 5.1: Unit Tests for Whitelist Enforcement

**File:** `pkg/tools/configure_test.go` (NEW)

**Description:** Test that only whitelisted paths can be accessed.

**Test Cases:**
- `TestConfigureTool_WhitelistEnforcement`
  - Valid provider paths allowed
  - Valid channel paths allowed
  - Valid agent paths allowed
  - Invalid/unknown paths rejected
  - Path traversal attempts rejected

**Acceptance Criteria:**
- [x] All test cases pass
- [x] Tests run in isolation (use temp directories)

**Estimated Time:** 30 minutes

---

### Task 5.2: Unit Tests for Sensitive Field Redaction

**File:** `pkg/tools/configure_test.go`

**Description:** Test that sensitive fields are never exposed.

**Test Cases:**
- `TestConfigureTool_SensitiveFieldsWriteOnly`
  - Get action on api_key returns write-only error
  - Get action on token returns write-only error
  - Set action on api_key succeeds (doesn't return value)
  - list_providers shows masked values

**Acceptance Criteria:**
- [x] No actual sensitive values in any responses
- [x] Write operations succeed without exposing values

**Estimated Time:** 30 minutes

---

### Task 5.3: Unit Tests for Set and Validation

**File:** `pkg/tools/configure_test.go`

**Description:** Test set action with validation.

**Test Cases:**
- `TestConfigureTool_SetAndVerify`
  - Set non-sensitive value, verify in config file
  - Set sensitive value, verify saved but not returned
- `TestConfigureTool_ValidationRejection`
  - Invalid port (70000) rejected
  - Invalid temperature (5.0) rejected
  - Invalid URL rejected
  - Valid values accepted

**Acceptance Criteria:**
- [x] Valid values persisted correctly
- [x] Invalid values rejected with helpful messages
- [x] Config file integrity maintained

**Estimated Time:** 45 minutes

---

### Task 5.4: Unit Tests for Enable/Disable

**File:** `pkg/tools/configure_test.go`

**Description:** Test enable and disable actions.

**Test Cases:**
- `TestConfigureTool_EnableDisable`
  - Enable channel, verify enabled=true in config
  - Disable channel, verify enabled=false in config
  - Enable with invalid path rejected
  - Response includes restart hint

**Acceptance Criteria:**
- [x] Enable/disable work for valid paths
- [x] Invalid paths rejected

**Estimated Time:** 20 minutes

---

### Task 5.5: Unit Tests for List Actions

**File:** `pkg/tools/configure_test.go`

**Description:** Test list_providers and list_channels actions.

**Test Cases:**
- `TestConfigureTool_ListProviders`
  - All 10 providers in output
  - API key status shown correctly
  - Custom API base indicated
- `TestConfigureTool_ListChannels`
  - All 9 channels in output
  - Enabled status correct
  - Credential status without actual values

**Acceptance Criteria:**
- [x] All providers/channels listed
- [x] No sensitive values in output

**Estimated Time:** 30 minutes

---

### Task 5.6: Integration Test - Full Workflow

**File:** `pkg/tools/configure_test.go`

**Description:** End-to-end test of typical user workflow.

**Test Case:**
- `TestConfigureTool_FullWorkflow`
  1. list_providers (shows none configured)
  2. set API key
  3. list_providers (shows key set)
  4. enable channel
  5. list_channels (shows channel enabled)
  6. set channel token
  7. list_channels (shows usable)
  8. Verify config file persisted correctly

**Acceptance Criteria:**
- [x] Full workflow completes without errors
- [x] Config persisted correctly

**Estimated Time:** 30 minutes

---

### Task 5.7: Security Tests

**File:** `pkg/tools/configure_test.go`

**Description:** Security-focused tests.

**Test Cases:**
- `TestConfigureTool_RequiresUserContext`
  - Tool fails gracefully when userUUID not set
- `TestConfigureTool_NoSecretExposure`
  - After setting 10+ different sensitive fields
  - No output contains actual secret values
  - Audit log args are sanitized
- `TestConfigureTool_PathTraversalPrevented`
  - Paths like `../../other/config` rejected
  - Paths with leading `/` rejected
  - Only alphanumeric and dots allowed

**Acceptance Criteria:**
- [x] All security tests pass
- [x] No way to extract secrets through tool

**Estimated Time:** 30 minutes

---

### Task 5.8: Test Audit Logging

**File:** `pkg/tools/configure_test.go` or `pkg/tools/audit_test.go`

**Description:** Verify audit logging works correctly for configure tool.

**Test Cases:**
- `TestConfigureTool_AuditLogging`
  - Configure tool calls are logged
  - Sensitive values are redacted in logs
  - Path is logged for traceability

**Acceptance Criteria:**
- [x] Audit entries created for configure calls
- [x] No sensitive values in audit log

**Estimated Time:** 20 minutes

---

## Summary

### Phase Breakdown

| Phase | Tasks | Est. Time | Key Deliverables |
|-------|-------|-----------|------------------|
| 1. Foundation | 4 tasks | ~1.5 hours | Interfaces, skeleton, whitelist |
| 2. Config Helpers | 4 tasks | ~2 hours | Redaction, validation, tests |
| 3. Tool Actions | 8 tasks | ~4.5 hours | All 6 actions implemented |
| 4. Integration | 5 tasks | ~1.5 hours | Agent loop wiring, audit |
| 5. Testing | 8 tasks | ~3.5 hours | Full test coverage |

### Files Modified

| File | Change Type |
|------|-------------|
| `pkg/tools/base.go` | Modified - Add interface |
| `pkg/tools/configure.go` | **NEW** - Main implementation |
| `pkg/tools/configure_test.go` | **NEW** - Tests |
| `pkg/tools/audit.go` | Modified - Add to RestrictedTools, update sanitization |
| `pkg/config/redact.go` | **NEW** - Redaction utilities |
| `pkg/config/redact_test.go` | **NEW** - Redaction tests |
| `pkg/config/validate.go` | **NEW** - Validation utilities |
| `pkg/config/validate_test.go` | **NEW** - Validation tests |
| `pkg/agent/loop.go` | Modified - Register tool, propagate context |

### Dependencies Between Tasks

```
Phase 1: Foundation
  1.1 → 1.2 → 1.3 → 1.4

Phase 2: Config Helpers (can run parallel to Phase 1 after 1.2)
  2.1 → 2.3
  2.2 → 2.4

Phase 3: Tool Actions (requires Phase 1 + 2)
  3.1 → 3.2 → 3.3 → 3.4 → 3.5 → 3.6
                  ↘ 3.7
                  ↘ 3.8

Phase 4: Integration (requires Phase 3)
  4.1 → 4.2 → 4.3 → 4.4 → 4.5

Phase 5: Testing (requires Phase 4)
  5.1 through 5.8 (can run somewhat parallel)
```

---

## Verification Checklist

After all tasks complete, verify:

- [ ] `go build ./...` succeeds
- [ ] `go test ./pkg/tools -v -run Configure` passes
- [ ] `go test ./pkg/config -v -run Redact` passes
- [ ] `go test ./pkg/config -v -run Validate` passes
- [ ] `go vet ./...` reports no issues
- [ ] Manual test: Run agent and execute configure tool commands
- [ ] Security audit: Confirm no sensitive values exposed

---

## Related Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Exploration | `openspec/changes/configure-tool/exploration.md` | Complete |
| Proposal | `openspec/changes/configure-tool/proposal.md` | Complete |
| Specification | `openspec/changes/configure-tool/specs.md` | Complete |
| Design | `openspec/changes/configure-tool/design.md` | Complete |
| Tasks | `openspec/changes/configure-tool/tasks.md` | This document |
