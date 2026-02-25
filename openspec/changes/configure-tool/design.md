# Technical Design: Configure Tool for Agent-Assisted Configuration

**Change ID:** configure-tool
**Status:** Draft
**Author:** SDD Sub-Agent
**Created:** 2026-02-25
**Based on:** [proposal.md](proposal.md), [exploration.md](exploration.md)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Architecture](#2-architecture)
3. [Data Flow](#3-data-flow)
4. [Key Components](#4-key-components)
5. [Security Design](#5-security-design)
6. [Code Examples](#6-code-examples)
7. [Testing Strategy](#7-testing-strategy)
8. [Implementation Notes](#8-implementation-notes)

---

## 1. Overview

### Purpose

This design document details the technical implementation of the `ConfigureTool` - a new tool that enables the MakoClaw agent to help users configure their channels, providers, and agent settings through natural conversation while maintaining strict security controls.

### Design Goals

1. **Secure by Default**: Sensitive values are write-only; never exposed to LLM
2. **Whitelist-Only Access**: Explicit field allowlist prevents misconfiguration
3. **Full Auditability**: All changes logged with user context
4. **Minimal Footprint**: Single tool file plus helper modules
5. **Consistent Patterns**: Follows existing tool implementation conventions

### Scope

| Component | Path | Purpose |
|-----------|------|---------|
| ConfigureTool | `pkg/tools/configure.go` | Main tool implementation |
| UserConfigTool interface | `pkg/tools/base.go` | New interface for UUID-based config access |
| Redaction helpers | `pkg/config/redact.go` | Sensitive field redaction utilities |
| Validation helpers | `pkg/config/validate.go` | Field validation rules |
| Loop updates | `pkg/agent/loop.go` | Tool registration, UUID propagation |
| Audit integration | `pkg/tools/audit.go` | Add to RestrictedTools |

---

## 2. Architecture

### Component Diagram

```
                                    ┌─────────────────────────────────────┐
                                    │           Agent Loop                │
                                    │  (pkg/agent/loop.go)                │
                                    │                                     │
                                    │  - SetUserForAgent(uuid, id)        │
                                    │  - updateToolsUserConfig(uuid, id)  │
                                    │  - Tool registration                │
                                    └──────────────┬──────────────────────┘
                                                   │
                                                   │ SetUserContext(userID, userUUID)
                                                   ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                              ConfigureTool                                           │
│                          (pkg/tools/configure.go)                                    │
│                                                                                      │
│  ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐                    │
│  │   Field         │   │   Action        │   │   Response      │                    │
│  │   Whitelist     │   │   Handlers      │   │   Builder       │                    │
│  │                 │   │                 │   │                 │                    │
│  │ - providers.*   │   │ - handleGet()   │   │ - redactValue() │                    │
│  │ - channels.*    │   │ - handleSet()   │   │ - formatList()  │                    │
│  │ - agents.*      │   │ - handleEnable()│   │ - buildError()  │                    │
│  │                 │   │ - handleDisable()   │                 │                    │
│  └────────┬────────┘   │ - handleList*() │   └────────┬────────┘                    │
│           │            └────────┬────────┘            │                             │
│           │                     │                     │                             │
│           └──────────────────┬──┴─────────────────────┘                             │
│                              │                                                      │
└──────────────────────────────┼──────────────────────────────────────────────────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
        ▼                      ▼                      ▼
┌───────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Config       │    │  Validation     │    │  Audit          │
│  Package      │    │  Helpers        │    │  Logger         │
│               │    │                 │    │                 │
│ LoadForUser() │    │ ValidateField() │    │ LogToolExec()   │
│ SaveForUser() │    │ ValidateURL()   │    │ sanitizeArgs()  │
│ GetUserPath() │    │ ValidatePort()  │    │                 │
└───────────────┘    └─────────────────┘    └─────────────────┘
```

### Interface Hierarchy

```go
// Existing interfaces in pkg/tools/base.go
type Tool interface {
    Name() string
    Description() string
    Parameters() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}

type UserAwareTool interface {
    Tool
    SetUserID(userID int64)
}

// NEW: For tools needing user UUID (config file access)
type UserConfigTool interface {
    Tool
    SetUserContext(userID int64, userUUID string)
}
```

### Integration Points

| Component | Integration Method | Purpose |
|-----------|-------------------|---------|
| Agent Loop | `SetUserContext()` call in `SetUserForAgent()` | Propagate UUID to tool |
| Audit System | Add to `RestrictedTools` map | Mandatory logging |
| Config Package | `LoadConfigForUser()`, `SaveConfigForUser()` | Config I/O |
| Tool Registry | `toolsRegistry.Register()` | Registration |

---

## 3. Data Flow

### Request Processing Flow

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              Request Flow                                           │
└─────────────────────────────────────────────────────────────────────────────────────┘

User Request: "Set my OpenAI API key to sk-proj-abc123"
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│  1. PARSE & VALIDATE                                                                │
│                                                                                      │
│  args = {                                                                           │
│    "action": "set",                                                                 │
│    "path": "providers.openai.api_key",                                              │
│    "value": "sk-proj-abc123..."                                                     │
│  }                                                                                  │
│                                                                                      │
│  ├── Validate action is known: ✓                                                    │
│  ├── Validate path is in whitelist: ✓                                               │
│  ├── Validate value type matches schema: ✓                                          │
│  └── Validate field-specific rules: ✓                                               │
└──────────────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│  2. LOAD CONFIG                                                                     │
│                                                                                      │
│  configPath := ~/.MakoClaw/users/{userUUID}/config.json                             │
│  cfg, err := config.LoadConfigForUser(userUUID)                                     │
│                                                                                      │
│  ├── Try user config first                                                          │
│  ├── Fall back to global if not exists                                              │
│  └── Return merged config                                                           │
└──────────────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│  3. APPLY CHANGE                                                                    │
│                                                                                      │
│  path: "providers.openai.api_key"                                                   │
│                                                                                      │
│  ├── Navigate to: cfg.Providers.OpenAI                                              │
│  ├── Set field: APIKey = "sk-proj-abc123..."                                        │
│  └── Validate config still loads: ✓                                                 │
└──────────────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│  4. SAVE CONFIG                                                                     │
│                                                                                      │
│  err := config.SaveConfigForUser(userUUID, cfg)                                     │
│                                                                                      │
│  ├── Marshal to JSON                                                                │
│  ├── Write to user config path                                                      │
│  └── Verify file written successfully                                               │
└──────────────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│  5. AUDIT LOG                                                                       │
│                                                                                      │
│  auditLog := ToolExecutionLog{                                                      │
│    UserID:    userID,                                                               │
│    Tool:      "configure",                                                          │
│    Arguments: {                                                                     │
│      "action": "set",                                                               │
│      "path": "providers.openai.api_key",                                            │
│      "value": "[REDACTED]"           // <-- Sanitized!                              │
│    },                                                                               │
│    Success:   true,                                                                 │
│  }                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│  6. RESPONSE                                                                        │
│                                                                                      │
│  "Successfully updated providers.openai.api_key"                                    │
│                                                                                      │
│  NOTE: Response NEVER contains the actual value for sensitive fields                │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

### State Transitions

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Receive    │────▶│   Validate   │────▶│   Execute    │────▶│   Respond    │
│   Request    │     │   Request    │     │   Action     │     │   + Audit    │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
                            │                    │
                            ▼                    ▼
                     ┌──────────────┐     ┌──────────────┐
                     │ Return Error │     │ Return Error │
                     │ (validation) │     │ (execution)  │
                     └──────────────┘     └──────────────┘
```

---

## 4. Key Components

### 4.1 ConfigureTool Structure (`pkg/tools/configure.go`)

```go
package tools

import (
    "context"
    "fmt"
    "reflect"
    "strings"
    "sync"

    "github.com/sipeed/makoclaw/pkg/config"
    "github.com/sipeed/makoclaw/pkg/logger"
)

// ConfigureTool allows the agent to modify user configuration
type ConfigureTool struct {
    userID   int64
    userUUID string
    mu       sync.RWMutex
}

// Ensure ConfigureTool implements UserConfigTool
var _ UserConfigTool = (*ConfigureTool)(nil)

// NewConfigureTool creates a new configure tool instance
func NewConfigureTool() *ConfigureTool {
    return &ConfigureTool{}
}

func (t *ConfigureTool) Name() string {
    return "configure"
}

func (t *ConfigureTool) Description() string {
    return `Manage user configuration for providers, channels, and agent settings.
Actions: get, set, enable, disable, list_providers, list_channels.
Sensitive fields (API keys, tokens) are write-only and never displayed.`
}

func (t *ConfigureTool) SetUserContext(userID int64, userUUID string) {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.userID = userID
    t.userUUID = userUUID
}
```

### 4.2 Field Whitelist Definition

```go
// FieldPolicy defines read/write permissions for a config field
type FieldPolicy struct {
    Path      string       // Dot-notation path (supports wildcards)
    Type      FieldType    // string, int, float, bool, stringSlice
    Readable  bool         // Can be read via 'get'
    Writable  bool         // Can be written via 'set'
    Validator ValidatorFn  // Optional validation function
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
    // Providers - API keys are write-only
    "providers.*.api_key":     {Type: TypeString, Readable: false, Writable: true},
    "providers.*.api_base":    {Type: TypeString, Readable: true, Writable: true, Validator: validateURL},
    "providers.*.proxy":       {Type: TypeString, Readable: true, Writable: true, Validator: validateURL},
    "providers.*.auth_method": {Type: TypeString, Readable: true, Writable: true},
    "providers.*.models":      {Type: TypeStringSlice, Readable: true, Writable: true},

    // Channels - tokens are write-only
    "channels.*.enabled":      {Type: TypeBool, Readable: true, Writable: true},
    "channels.*.token":        {Type: TypeString, Readable: false, Writable: true},
    "channels.*.bot_token":    {Type: TypeString, Readable: false, Writable: true},
    "channels.*.app_token":    {Type: TypeString, Readable: false, Writable: true},
    "channels.*.app_id":       {Type: TypeString, Readable: false, Writable: true},
    "channels.*.app_secret":   {Type: TypeString, Readable: false, Writable: true},
    "channels.*.client_id":    {Type: TypeString, Readable: false, Writable: true},
    "channels.*.client_secret":{Type: TypeString, Readable: false, Writable: true},
    "channels.*.proxy":        {Type: TypeString, Readable: true, Writable: true, Validator: validateURL},
    "channels.*.allow_from":   {Type: TypeStringSlice, Readable: true, Writable: true},
    "channels.*.host":         {Type: TypeString, Readable: true, Writable: true},
    "channels.*.port":         {Type: TypeInt, Readable: true, Writable: true, Validator: validatePort},
    "channels.*.bridge_url":   {Type: TypeString, Readable: true, Writable: true, Validator: validateURL},
    "channels.*.phone_number": {Type: TypeString, Readable: true, Writable: true},

    // Agents - all readable/writable
    "agents.defaults.provider":           {Type: TypeString, Readable: true, Writable: true},
    "agents.defaults.model":              {Type: TypeString, Readable: true, Writable: true},
    "agents.defaults.max_tokens":         {Type: TypeInt, Readable: true, Writable: true, Validator: validatePositiveInt},
    "agents.defaults.temperature":        {Type: TypeFloat, Readable: true, Writable: true, Validator: validateTemperature},
    "agents.defaults.max_tool_iterations":{Type: TypeInt, Readable: true, Writable: true, Validator: validatePositiveInt},
    "agents.orchestrator.enabled":        {Type: TypeBool, Readable: true, Writable: true},
    "agents.orchestrator.provider":       {Type: TypeString, Readable: true, Writable: true},
    "agents.orchestrator.model":          {Type: TypeString, Readable: true, Writable: true},

    // Tools - web search
    "tools.web.search.api_key":    {Type: TypeString, Readable: false, Writable: true},
    "tools.web.search.max_results":{Type: TypeInt, Readable: true, Writable: true, Validator: validatePositiveInt},

    // Tools - email
    "tools.email.enabled":  {Type: TypeBool, Readable: true, Writable: true},
    "tools.email.host":     {Type: TypeString, Readable: true, Writable: true},
    "tools.email.port":     {Type: TypeInt, Readable: true, Writable: true, Validator: validatePort},
    "tools.email.username": {Type: TypeString, Readable: true, Writable: true},
    "tools.email.password": {Type: TypeString, Readable: false, Writable: true},
    "tools.email.from":     {Type: TypeString, Readable: true, Writable: true, Validator: validateEmail},
    "tools.email.to":       {Type: TypeString, Readable: true, Writable: true, Validator: validateEmail},
}

// Provider names for wildcard resolution
var providerNames = []string{
    "anthropic", "openai", "openrouter", "groq", "zhipu",
    "vllm", "gemini", "nvidia", "moonshot", "ollama",
}

// Channel names for wildcard resolution
var channelNames = []string{
    "telegram", "discord", "slack", "whatsapp", "signal",
    "qq", "dingtalk", "feishu", "maixcam",
}
```

### 4.3 Tool Parameters Schema

```go
func (t *ConfigureTool) Parameters() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "action": map[string]interface{}{
                "type": "string",
                "enum": []string{"get", "set", "enable", "disable", "list_providers", "list_channels"},
                "description": "Action to perform",
            },
            "path": map[string]interface{}{
                "type": "string",
                "description": "Config path in dot notation (e.g., 'providers.openai.api_key')",
            },
            "value": map[string]interface{}{
                "description": "Value to set (for 'set' action). Type depends on field.",
            },
        },
        "required": []string{"action"},
    }
}
```

### 4.4 Execute Method

```go
func (t *ConfigureTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    t.mu.RLock()
    userID := t.userID
    userUUID := t.userUUID
    t.mu.RUnlock()

    if userUUID == "" {
        return "Error: user context not set. Configuration requires authenticated user.", nil
    }

    action, _ := args["action"].(string)

    switch action {
    case "get":
        return t.handleGet(ctx, args, userUUID)
    case "set":
        return t.handleSet(ctx, args, userUUID, userID)
    case "enable":
        return t.handleEnable(ctx, args, userUUID, userID)
    case "disable":
        return t.handleDisable(ctx, args, userUUID, userID)
    case "list_providers":
        return t.handleListProviders(ctx, userUUID)
    case "list_channels":
        return t.handleListChannels(ctx, userUUID)
    default:
        return fmt.Sprintf("Error: unknown action '%s'. Valid actions: get, set, enable, disable, list_providers, list_channels", action), nil
    }
}
```

### 4.5 Redaction Helpers (`pkg/config/redact.go`)

```go
package config

import (
    "strings"
)

// SensitiveFieldPatterns lists field name patterns that contain secrets
var SensitiveFieldPatterns = []string{
    "api_key", "apikey",
    "token", "bot_token", "app_token", "access_token", "refresh_token",
    "password",
    "secret", "app_secret", "client_secret", "signing_secret",
    "private_key", "encrypt_key",
}

// IsSensitiveField checks if a field name matches sensitive patterns
func IsSensitiveField(fieldName string) bool {
    lower := strings.ToLower(fieldName)
    for _, pattern := range SensitiveFieldPatterns {
        if strings.Contains(lower, pattern) {
            return true
        }
    }
    return false
}

// RedactValue returns a redacted representation of a sensitive value
// Shows "[REDACTED]" for empty/unset, "****{last4}" for set values
func RedactValue(value string) string {
    if value == "" {
        return "[NOT SET]"
    }
    if len(value) <= 4 {
        return "[REDACTED]"
    }
    return "****" + value[len(value)-4:]
}

// RedactConfigPath extracts the field name from a path and checks sensitivity
func RedactConfigPath(path string) bool {
    parts := strings.Split(path, ".")
    if len(parts) == 0 {
        return false
    }
    fieldName := parts[len(parts)-1]
    return IsSensitiveField(fieldName)
}
```

### 4.6 Validation Helpers (`pkg/config/validate.go`)

```go
package config

import (
    "fmt"
    "net/url"
    "regexp"
)

// ValidateURL ensures a string is a valid URL
func ValidateURL(value interface{}) error {
    s, ok := value.(string)
    if !ok {
        return fmt.Errorf("expected string, got %T", value)
    }
    if s == "" {
        return nil // Empty is valid (means "use default")
    }
    _, err := url.ParseRequestURI(s)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }
    return nil
}

// ValidatePort ensures a value is a valid port number (1-65535)
func ValidatePort(value interface{}) error {
    var port int
    switch v := value.(type) {
    case float64:
        port = int(v)
    case int:
        port = v
    default:
        return fmt.Errorf("expected number, got %T", value)
    }
    if port < 1 || port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535, got %d", port)
    }
    return nil
}

// ValidatePositiveInt ensures a value is a positive integer
func ValidatePositiveInt(value interface{}) error {
    var n int
    switch v := value.(type) {
    case float64:
        n = int(v)
    case int:
        n = v
    default:
        return fmt.Errorf("expected number, got %T", value)
    }
    if n < 0 {
        return fmt.Errorf("value must be positive, got %d", n)
    }
    return nil
}

// ValidateTemperature ensures temperature is in valid range [0.0, 2.0]
func ValidateTemperature(value interface{}) error {
    var temp float64
    switch v := value.(type) {
    case float64:
        temp = v
    case int:
        temp = float64(v)
    default:
        return fmt.Errorf("expected number, got %T", value)
    }
    if temp < 0.0 || temp > 2.0 {
        return fmt.Errorf("temperature must be between 0.0 and 2.0, got %f", temp)
    }
    return nil
}

// ValidateEmail provides basic email format validation
func ValidateEmail(value interface{}) error {
    s, ok := value.(string)
    if !ok {
        return fmt.Errorf("expected string, got %T", value)
    }
    if s == "" {
        return nil // Empty is valid
    }
    // Basic email regex
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(s) {
        return fmt.Errorf("invalid email format: %s", s)
    }
    return nil
}

// ValidateProviderName ensures the provider name is known
func ValidateProviderName(name string) error {
    validProviders := map[string]bool{
        "anthropic": true, "openai": true, "openrouter": true,
        "groq": true, "zhipu": true, "vllm": true,
        "gemini": true, "nvidia": true, "moonshot": true, "ollama": true,
    }
    if !validProviders[name] {
        return fmt.Errorf("unknown provider: %s", name)
    }
    return nil
}

// ValidateChannelName ensures the channel name is known
func ValidateChannelName(name string) error {
    validChannels := map[string]bool{
        "telegram": true, "discord": true, "slack": true,
        "whatsapp": true, "signal": true, "qq": true,
        "dingtalk": true, "feishu": true, "maixcam": true,
    }
    if !validChannels[name] {
        return fmt.Errorf("unknown channel: %s", name)
    }
    return nil
}
```

### 4.7 Agent Loop Updates (`pkg/agent/loop.go`)

```go
// Add to SetUserForAgent method, after existing updateToolsUser call:

func (al *AgentLoop) SetUserForAgent(userUUID string, userID int64) {
    // ... existing code ...

    al.contextBuilder.WithUser(userUUID, userID)
    al.updateToolsUser(userID)
    al.updateToolsUserConfig(userID, userUUID)  // NEW
}

// NEW: updateToolsUserConfig propagates user context to UserConfigTool implementations
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

// Add to NewAgentLoop tool registration section:
func NewAgentLoop(cfg *config.Config, msgBus *bus.MessageBus, provider providers.LLMProvider) *AgentLoop {
    // ... existing tool registrations ...

    // Register configure tool
    configureTool := tools.NewConfigureTool()
    toolsRegistry.Register(configureTool)

    // ... rest of function ...
}
```

### 4.8 Audit Integration (`pkg/tools/audit.go`)

```go
// Update RestrictedTools map to include configure
var RestrictedTools = map[string]bool{
    "exec":        true,
    "spawn":       true,
    "email":       true,
    "write_file":  true,
    "edit_file":   true,
    "append_file": true,
    "web_fetch":   true,
    "configure":   true,  // NEW: Config changes must be audited
}
```

---

## 5. Security Design

### 5.1 Threat Model

| Threat | Attack Vector | Mitigation |
|--------|---------------|------------|
| Secret Extraction | LLM prompt injection asking to read API keys | Write-only policy: sensitive fields never returned |
| Path Traversal | Malicious path like `../../other_user/config` | UUID from authenticated context only; no path construction from args |
| Config Corruption | Invalid values breaking system | Validate before write; verify config loads after save |
| Audit Bypass | Secrets in audit logs | `sanitizeArguments()` applied before logging |
| Race Conditions | Concurrent config writes | Mutex lock during config save operations |
| Privilege Escalation | User modifying admin-only fields | Whitelist-only access; no structural modifications |

### 5.2 Security Layers

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Layer 1: Authentication                                                   │
│                                                                             │
│  - userUUID MUST be set by authenticated agent loop                        │
│  - Execute() returns error if userUUID is empty                            │
│  - No user input influences config file path                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  Layer 2: Authorization (Whitelist)                                        │
│                                                                             │
│  - Only paths in fieldWhitelist can be accessed                            │
│  - Each field has explicit Readable/Writable flags                         │
│  - Wildcard resolution only for known providers/channels                   │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  Layer 3: Validation                                                       │
│                                                                             │
│  - Type checking before assignment                                         │
│  - Field-specific validators (URL, port, email, etc.)                     │
│  - Config verified loadable after modification                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  Layer 4: Response Sanitization                                            │
│                                                                             │
│  - Sensitive fields NEVER appear in tool output                            │
│  - Success messages are generic: "Successfully updated X"                  │
│  - Error messages don't leak sensitive values                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  Layer 5: Audit Trail                                                      │
│                                                                             │
│  - All configure tool calls logged                                         │
│  - Arguments sanitized before storage                                      │
│  - Includes: user, action, path, timestamp, success/failure                │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.3 Sensitive Field Handling

```go
// Example: handleGet for a sensitive field
func (t *ConfigureTool) handleGet(ctx context.Context, args map[string]interface{}, userUUID string) (string, error) {
    path, _ := args["path"].(string)

    // Look up field policy
    policy, err := t.resolveFieldPolicy(path)
    if err != nil {
        return fmt.Sprintf("Error: %v", err), nil
    }

    // SECURITY: Check if field is readable
    if !policy.Readable {
        return fmt.Sprintf("Error: field '%s' is write-only and cannot be read", path), nil
    }

    // Load config and get value...
    cfg, err := config.LoadConfigForUser(userUUID)
    if err != nil {
        return fmt.Sprintf("Error loading config: %v", err), nil
    }

    value, err := t.getFieldValue(cfg, path)
    if err != nil {
        return fmt.Sprintf("Error: %v", err), nil
    }

    // SECURITY: Even for readable fields, double-check sensitivity
    if config.RedactConfigPath(path) {
        return fmt.Sprintf("%s = %s", path, config.RedactValue(fmt.Sprint(value))), nil
    }

    return fmt.Sprintf("%s = %v", path, value), nil
}
```

---

## 6. Code Examples

### 6.1 Complete Action Handlers

```go
// handleSet updates a configuration field value
func (t *ConfigureTool) handleSet(ctx context.Context, args map[string]interface{}, userUUID string, userID int64) (string, error) {
    path, _ := args["path"].(string)
    value := args["value"]

    if path == "" {
        return "Error: 'path' is required for set action", nil
    }
    if value == nil {
        return "Error: 'value' is required for set action", nil
    }

    // Resolve and validate field policy
    policy, err := t.resolveFieldPolicy(path)
    if err != nil {
        return fmt.Sprintf("Error: field '%s' is not configurable", path), nil
    }

    if !policy.Writable {
        return fmt.Sprintf("Error: field '%s' is read-only", path), nil
    }

    // Type validation
    if err := t.validateValueType(value, policy.Type); err != nil {
        return fmt.Sprintf("Error: %v", err), nil
    }

    // Custom field validation
    if policy.Validator != nil {
        if err := policy.Validator(value); err != nil {
            return fmt.Sprintf("Error: invalid value for '%s': %v", path, err), nil
        }
    }

    // Load current config
    cfg, err := config.LoadConfigForUser(userUUID)
    if err != nil {
        return fmt.Sprintf("Error loading config: %v", err), nil
    }

    // Apply change
    if err := t.setFieldValue(cfg, path, value); err != nil {
        return fmt.Sprintf("Error setting field: %v", err), nil
    }

    // Save config
    if err := config.SaveConfigForUser(userUUID, cfg); err != nil {
        return fmt.Sprintf("Error saving config: %v", err), nil
    }

    // Verify config still loads
    if _, err := config.LoadConfigForUser(userUUID); err != nil {
        // Rollback would happen here in a more robust implementation
        return fmt.Sprintf("Error: config validation failed after save: %v", err), nil
    }

    logger.InfoCF("configure", "Configuration updated", map[string]interface{}{
        "user_id":   userID,
        "user_uuid": userUUID,
        "path":      path,
        "sensitive": config.RedactConfigPath(path),
    })

    // Return success (never include the actual value for sensitive fields)
    return fmt.Sprintf("Successfully updated %s", path), nil
}

// handleEnable enables a channel
func (t *ConfigureTool) handleEnable(ctx context.Context, args map[string]interface{}, userUUID string, userID int64) (string, error) {
    path, _ := args["path"].(string)

    // Validate path format: must be channels.{name}
    parts := strings.Split(path, ".")
    if len(parts) != 2 || parts[0] != "channels" {
        return "Error: enable action requires path like 'channels.telegram'", nil
    }

    channelName := parts[1]
    if err := config.ValidateChannelName(channelName); err != nil {
        return fmt.Sprintf("Error: %v", err), nil
    }

    // Delegate to handleSet with enabled=true
    args["path"] = path + ".enabled"
    args["value"] = true

    result, err := t.handleSet(ctx, args, userUUID, userID)
    if err != nil {
        return result, err
    }

    return fmt.Sprintf("Successfully enabled %s. Note: You may need to restart the gateway for changes to take effect.", channelName), nil
}

// handleListProviders lists all providers with their configuration status
func (t *ConfigureTool) handleListProviders(ctx context.Context, userUUID string) (string, error) {
    cfg, err := config.LoadConfigForUser(userUUID)
    if err != nil {
        return fmt.Sprintf("Error loading config: %v", err), nil
    }

    var sb strings.Builder
    sb.WriteString("Configured providers:\n")

    providers := map[string]config.ProviderConfig{
        "anthropic":  cfg.Providers.Anthropic,
        "openai":     cfg.Providers.OpenAI,
        "openrouter": cfg.Providers.OpenRouter,
        "groq":       cfg.Providers.Groq,
        "zhipu":      cfg.Providers.Zhipu,
        "vllm":       cfg.Providers.VLLM,
        "gemini":     cfg.Providers.Gemini,
        "nvidia":     cfg.Providers.Nvidia,
        "moonshot":   cfg.Providers.Moonshot,
        "ollama":     cfg.Providers.Ollama,
    }

    for name, p := range providers {
        apiKeyStatus := config.RedactValue(p.APIKey)
        apiBase := p.APIBase
        if apiBase == "" {
            apiBase = "default"
        }
        sb.WriteString(fmt.Sprintf("- %s: api_key=%s, api_base=%s\n", name, apiKeyStatus, apiBase))
    }

    return sb.String(), nil
}
```

### 6.2 Field Path Resolution

```go
// resolveFieldPolicy looks up the policy for a given path, resolving wildcards
func (t *ConfigureTool) resolveFieldPolicy(path string) (*FieldPolicy, error) {
    // Direct lookup first
    if policy, ok := fieldWhitelist[path]; ok {
        return &policy, nil
    }

    // Try wildcard patterns
    parts := strings.Split(path, ".")
    if len(parts) < 2 {
        return nil, fmt.Errorf("path '%s' not in whitelist", path)
    }

    // Build wildcard path: replace second segment with *
    wildcardPath := parts[0] + ".*"
    if len(parts) > 2 {
        wildcardPath += "." + strings.Join(parts[2:], ".")
    }

    if policy, ok := fieldWhitelist[wildcardPath]; ok {
        // Validate the specific name
        switch parts[0] {
        case "providers":
            if err := config.ValidateProviderName(parts[1]); err != nil {
                return nil, err
            }
        case "channels":
            if err := config.ValidateChannelName(parts[1]); err != nil {
                return nil, err
            }
        }
        return &policy, nil
    }

    return nil, fmt.Errorf("path '%s' not in whitelist", path)
}

// getFieldValue retrieves a field value from config using reflection
func (t *ConfigureTool) getFieldValue(cfg *config.Config, path string) (interface{}, error) {
    parts := strings.Split(path, ".")

    val := reflect.ValueOf(cfg).Elem()

    for _, part := range parts {
        // Handle map access for providers/channels
        if val.Kind() == reflect.Struct {
            // Convert part to title case for struct field lookup
            fieldName := strings.Title(part)
            val = val.FieldByName(fieldName)
            if !val.IsValid() {
                return nil, fmt.Errorf("field '%s' not found", part)
            }
        } else {
            return nil, fmt.Errorf("cannot navigate into %s", val.Kind())
        }
    }

    return val.Interface(), nil
}

// setFieldValue sets a field value in config using reflection
func (t *ConfigureTool) setFieldValue(cfg *config.Config, path string, value interface{}) error {
    parts := strings.Split(path, ".")

    val := reflect.ValueOf(cfg).Elem()

    // Navigate to parent
    for i := 0; i < len(parts)-1; i++ {
        fieldName := strings.Title(parts[i])
        val = val.FieldByName(fieldName)
        if !val.IsValid() {
            return fmt.Errorf("field '%s' not found", parts[i])
        }
        if val.Kind() == reflect.Ptr {
            val = val.Elem()
        }
    }

    // Set the final field
    fieldName := strings.Title(parts[len(parts)-1])
    field := val.FieldByName(fieldName)
    if !field.IsValid() || !field.CanSet() {
        return fmt.Errorf("field '%s' not found or not settable", fieldName)
    }

    // Type conversion and assignment
    switch field.Kind() {
    case reflect.String:
        s, ok := value.(string)
        if !ok {
            return fmt.Errorf("expected string for %s", path)
        }
        field.SetString(s)
    case reflect.Int, reflect.Int64:
        switch v := value.(type) {
        case float64:
            field.SetInt(int64(v))
        case int:
            field.SetInt(int64(v))
        default:
            return fmt.Errorf("expected number for %s", path)
        }
    case reflect.Float64:
        switch v := value.(type) {
        case float64:
            field.SetFloat(v)
        case int:
            field.SetFloat(float64(v))
        default:
            return fmt.Errorf("expected number for %s", path)
        }
    case reflect.Bool:
        b, ok := value.(bool)
        if !ok {
            return fmt.Errorf("expected boolean for %s", path)
        }
        field.SetBool(b)
    case reflect.Slice:
        // Handle []string (FlexibleStringSlice)
        if slice, ok := value.([]interface{}); ok {
            strSlice := make([]string, len(slice))
            for i, v := range slice {
                strSlice[i] = fmt.Sprint(v)
            }
            field.Set(reflect.ValueOf(config.FlexibleStringSlice(strSlice)))
        } else {
            return fmt.Errorf("expected array for %s", path)
        }
    default:
        return fmt.Errorf("unsupported field type: %s", field.Kind())
    }

    return nil
}
```

---

## 7. Testing Strategy

### 7.1 Test Categories

| Category | Purpose | Location |
|----------|---------|----------|
| Unit Tests | Individual function testing | `pkg/tools/configure_test.go` |
| Integration Tests | Config load/save cycles | `pkg/tools/configure_test.go` |
| Security Tests | Whitelist enforcement, redaction | `pkg/tools/configure_test.go` |
| Validation Tests | Field validators | `pkg/config/validate_test.go` |
| Redaction Tests | Sensitive field handling | `pkg/config/redact_test.go` |

### 7.2 Unit Test Examples

```go
// pkg/tools/configure_test.go

package tools

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/sipeed/makoclaw/pkg/config"
)

func TestConfigureTool_WhitelistEnforcement(t *testing.T) {
    tool := NewConfigureTool()
    tool.SetUserContext(1, "test-uuid-123")

    // Create temp config
    tempDir := t.TempDir()
    config.InitDataDir(tempDir)
    userDir := filepath.Join(tempDir, "users", "test-uuid-123")
    os.MkdirAll(userDir, 0755)

    tests := []struct {
        name      string
        path      string
        wantError bool
    }{
        {"valid provider path", "providers.openai.api_key", false},
        {"valid channel path", "channels.telegram.enabled", false},
        {"valid agent path", "agents.defaults.model", false},
        {"invalid path", "some.random.path", true},
        {"dangerous path", "web.password", true},
        {"traversal attempt", "../../other_user/config", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, _ := tool.Execute(context.Background(), map[string]interface{}{
                "action": "get",
                "path":   tt.path,
            })

            hasError := strings.Contains(result, "Error")
            if hasError != tt.wantError {
                t.Errorf("path %q: got error=%v, want error=%v, result=%s",
                    tt.path, hasError, tt.wantError, result)
            }
        })
    }
}

func TestConfigureTool_SensitiveFieldsWriteOnly(t *testing.T) {
    tool := NewConfigureTool()
    tool.SetUserContext(1, "test-uuid-123")

    // Setup temp config directory
    tempDir := t.TempDir()
    config.InitDataDir(tempDir)
    userDir := filepath.Join(tempDir, "users", "test-uuid-123")
    os.MkdirAll(userDir, 0755)

    // Create a config with an API key
    cfg := config.DefaultConfig()
    cfg.Providers.OpenAI.APIKey = "sk-test-secret-key-12345"
    config.SaveConfigForUser("test-uuid-123", cfg)

    // Try to read the API key
    result, _ := tool.Execute(context.Background(), map[string]interface{}{
        "action": "get",
        "path":   "providers.openai.api_key",
    })

    // Should NOT contain the actual key
    if strings.Contains(result, "sk-test-secret-key-12345") {
        t.Error("API key was exposed in get response")
    }

    // Should indicate it's write-only
    if !strings.Contains(result, "write-only") && !strings.Contains(result, "Error") {
        t.Error("Should indicate field is write-only")
    }
}

func TestConfigureTool_SetAndVerify(t *testing.T) {
    tool := NewConfigureTool()
    tool.SetUserContext(1, "test-uuid-123")

    tempDir := t.TempDir()
    config.InitDataDir(tempDir)
    userDir := filepath.Join(tempDir, "users", "test-uuid-123")
    os.MkdirAll(userDir, 0755)

    // Set a non-sensitive value
    result, err := tool.Execute(context.Background(), map[string]interface{}{
        "action": "set",
        "path":   "agents.defaults.model",
        "value":  "gpt-4",
    })

    if err != nil {
        t.Fatalf("Execute failed: %v", err)
    }

    if !strings.Contains(result, "Successfully") {
        t.Errorf("Expected success message, got: %s", result)
    }

    // Verify the value was saved
    cfg, _ := config.LoadConfigForUser("test-uuid-123")
    if cfg.Agents.Defaults.Model != "gpt-4" {
        t.Errorf("Model not saved, got: %s", cfg.Agents.Defaults.Model)
    }
}

func TestConfigureTool_ValidationRejection(t *testing.T) {
    tool := NewConfigureTool()
    tool.SetUserContext(1, "test-uuid-123")

    tempDir := t.TempDir()
    config.InitDataDir(tempDir)
    userDir := filepath.Join(tempDir, "users", "test-uuid-123")
    os.MkdirAll(userDir, 0755)

    tests := []struct {
        name  string
        path  string
        value interface{}
    }{
        {"invalid port too high", "channels.maixcam.port", 70000},
        {"invalid port negative", "channels.maixcam.port", -1},
        {"invalid temperature", "agents.defaults.temperature", 5.0},
        {"invalid URL", "providers.openai.api_base", "not-a-url"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, _ := tool.Execute(context.Background(), map[string]interface{}{
                "action": "set",
                "path":   tt.path,
                "value":  tt.value,
            })

            if !strings.Contains(result, "Error") && !strings.Contains(result, "invalid") {
                t.Errorf("Expected validation error for %s=%v, got: %s",
                    tt.path, tt.value, result)
            }
        })
    }
}

func TestConfigureTool_RequiresUserContext(t *testing.T) {
    tool := NewConfigureTool()
    // Intentionally NOT calling SetUserContext

    result, _ := tool.Execute(context.Background(), map[string]interface{}{
        "action": "list_providers",
    })

    if !strings.Contains(result, "Error") || !strings.Contains(result, "user context") {
        t.Errorf("Should require user context, got: %s", result)
    }
}
```

### 7.3 Redaction Tests

```go
// pkg/config/redact_test.go

package config

import "testing"

func TestIsSensitiveField(t *testing.T) {
    tests := []struct {
        field    string
        expected bool
    }{
        {"api_key", true},
        {"APIKey", true},
        {"api_base", false},
        {"token", true},
        {"bot_token", true},
        {"password", true},
        {"model", false},
        {"enabled", false},
        {"app_secret", true},
        {"client_secret", true},
        {"temperature", false},
    }

    for _, tt := range tests {
        t.Run(tt.field, func(t *testing.T) {
            result := IsSensitiveField(tt.field)
            if result != tt.expected {
                t.Errorf("IsSensitiveField(%q) = %v, want %v",
                    tt.field, result, tt.expected)
            }
        })
    }
}

func TestRedactValue(t *testing.T) {
    tests := []struct {
        value    string
        expected string
    }{
        {"", "[NOT SET]"},
        {"abc", "[REDACTED]"},
        {"abcd", "[REDACTED]"},
        {"sk-proj-abc12345", "****2345"},
        {"very-long-secret-key-value", "****alue"},
    }

    for _, tt := range tests {
        t.Run(tt.value, func(t *testing.T) {
            result := RedactValue(tt.value)
            if result != tt.expected {
                t.Errorf("RedactValue(%q) = %q, want %q",
                    tt.value, result, tt.expected)
            }
        })
    }
}
```

### 7.4 Integration Tests

```go
func TestConfigureTool_FullWorkflow(t *testing.T) {
    tool := NewConfigureTool()
    tool.SetUserContext(1, "test-uuid-workflow")

    tempDir := t.TempDir()
    config.InitDataDir(tempDir)
    userDir := filepath.Join(tempDir, "users", "test-uuid-workflow")
    os.MkdirAll(userDir, 0755)

    ctx := context.Background()

    // 1. List providers (should show defaults)
    result, _ := tool.Execute(ctx, map[string]interface{}{
        "action": "list_providers",
    })
    if !strings.Contains(result, "openai") {
        t.Error("list_providers should include openai")
    }

    // 2. Set API key
    result, _ = tool.Execute(ctx, map[string]interface{}{
        "action": "set",
        "path":   "providers.openai.api_key",
        "value":  "sk-test-key-12345",
    })
    if !strings.Contains(result, "Successfully") {
        t.Errorf("Failed to set API key: %s", result)
    }

    // 3. List providers (should show key is set)
    result, _ = tool.Execute(ctx, map[string]interface{}{
        "action": "list_providers",
    })
    if !strings.Contains(result, "****2345") {
        t.Error("API key should show as masked ****2345")
    }

    // 4. Enable a channel
    result, _ = tool.Execute(ctx, map[string]interface{}{
        "action": "enable",
        "path":   "channels.telegram",
    })
    if !strings.Contains(result, "enabled") {
        t.Errorf("Failed to enable channel: %s", result)
    }

    // 5. Verify config persisted
    cfg, _ := config.LoadConfigForUser("test-uuid-workflow")
    if cfg.Providers.OpenAI.APIKey != "sk-test-key-12345" {
        t.Error("API key not persisted")
    }
    if !cfg.Channels.Telegram.Enabled {
        t.Error("Channel not enabled")
    }
}
```

---

## 8. Implementation Notes

### 8.1 File Structure

```
pkg/
├── tools/
│   ├── base.go           # Add UserConfigTool interface
│   ├── configure.go      # NEW: Main tool implementation
│   ├── configure_test.go # NEW: Tool tests
│   └── audit.go          # Update RestrictedTools
│
├── config/
│   ├── config.go         # Existing (no changes needed)
│   ├── redact.go         # NEW: Redaction helpers
│   ├── redact_test.go    # NEW: Redaction tests
│   ├── validate.go       # NEW: Validation helpers
│   └── validate_test.go  # NEW: Validation tests
│
└── agent/
    └── loop.go           # Add updateToolsUserConfig, register tool
```

### 8.2 Dependencies

No new external dependencies required. Uses only:
- Standard library: `context`, `fmt`, `reflect`, `strings`, `sync`
- Internal packages: `pkg/config`, `pkg/logger`

### 8.3 Backward Compatibility

- Existing config files unchanged
- No migration required
- Tool is additive; can be disabled via tool permissions if needed

### 8.4 Known Limitations (v1)

1. **No Specialist Management**: Creating/deleting specialists is deferred to v2
2. **No MCP Server Config**: Complex nested structure deferred to v2
3. **No Rollback**: Previous values logged but not automatically recoverable
4. **No Live Reload**: Config changes require gateway restart to take effect

### 8.5 Future Enhancements (v2)

- Specialist CRUD operations
- MCP server configuration
- Config validation with provider connectivity checks
- Automatic backup before changes
- Undo/rollback capability
- Config diff/history view

---

## Appendix A: Complete Field Whitelist

| Path | Type | Readable | Writable | Validator |
|------|------|----------|----------|-----------|
| `providers.*.api_key` | string | No | Yes | - |
| `providers.*.api_base` | string | Yes | Yes | validateURL |
| `providers.*.proxy` | string | Yes | Yes | validateURL |
| `providers.*.auth_method` | string | Yes | Yes | - |
| `providers.*.models` | []string | Yes | Yes | - |
| `channels.*.enabled` | bool | Yes | Yes | - |
| `channels.*.token` | string | No | Yes | - |
| `channels.*.bot_token` | string | No | Yes | - |
| `channels.*.app_token` | string | No | Yes | - |
| `channels.*.app_id` | string | No | Yes | - |
| `channels.*.app_secret` | string | No | Yes | - |
| `channels.*.client_id` | string | No | Yes | - |
| `channels.*.client_secret` | string | No | Yes | - |
| `channels.*.proxy` | string | Yes | Yes | validateURL |
| `channels.*.allow_from` | []string | Yes | Yes | - |
| `channels.*.host` | string | Yes | Yes | - |
| `channels.*.port` | int | Yes | Yes | validatePort |
| `channels.*.bridge_url` | string | Yes | Yes | validateURL |
| `channels.*.phone_number` | string | Yes | Yes | - |
| `agents.defaults.provider` | string | Yes | Yes | - |
| `agents.defaults.model` | string | Yes | Yes | - |
| `agents.defaults.max_tokens` | int | Yes | Yes | validatePositiveInt |
| `agents.defaults.temperature` | float | Yes | Yes | validateTemperature |
| `agents.defaults.max_tool_iterations` | int | Yes | Yes | validatePositiveInt |
| `agents.orchestrator.enabled` | bool | Yes | Yes | - |
| `agents.orchestrator.provider` | string | Yes | Yes | - |
| `agents.orchestrator.model` | string | Yes | Yes | - |
| `tools.web.search.api_key` | string | No | Yes | - |
| `tools.web.search.max_results` | int | Yes | Yes | validatePositiveInt |
| `tools.email.enabled` | bool | Yes | Yes | - |
| `tools.email.host` | string | Yes | Yes | - |
| `tools.email.port` | int | Yes | Yes | validatePort |
| `tools.email.username` | string | Yes | Yes | - |
| `tools.email.password` | string | No | Yes | - |
| `tools.email.from` | string | Yes | Yes | validateEmail |
| `tools.email.to` | string | Yes | Yes | validateEmail |

---

## Appendix B: Example Tool Responses

### Successful Set (Sensitive Field)
```
Successfully updated providers.openai.api_key
```

### Successful Set (Safe Field)
```
Successfully updated agents.defaults.model
```

### List Providers
```
Configured providers:
- anthropic: api_key=[NOT SET], api_base=default
- openai: api_key=****1234, api_base=default
- openrouter: api_key=[NOT SET], api_base=default
- groq: api_key=****5678, api_base=default
- zhipu: api_key=[NOT SET], api_base=default
- vllm: api_key=[NOT SET], api_base=default
- gemini: api_key=[NOT SET], api_base=default
- nvidia: api_key=[NOT SET], api_base=default
- moonshot: api_key=[NOT SET], api_base=default
- ollama: api_key=[NOT SET], api_base=http://localhost:11434/v1
```

### List Channels
```
Configured channels:
- telegram: enabled=false, token=[NOT SET]
- discord: enabled=true, token=****4567
- slack: enabled=false, bot_token=[NOT SET]
- whatsapp: enabled=false
- signal: enabled=false
- qq: enabled=false
- dingtalk: enabled=false
- feishu: enabled=false
- maixcam: enabled=false, host=0.0.0.0, port=18790
```

### Validation Error
```
Error: invalid value for 'agents.defaults.temperature': temperature must be between 0.0 and 2.0, got 5.000000
```

### Whitelist Error
```
Error: field 'some.random.path' is not configurable
```

### Write-Only Field Error
```
Error: field 'providers.openai.api_key' is write-only and cannot be read
```

---

## Approval

| Role | Name | Date | Decision |
|------|------|------|----------|
| Author | SDD Sub-Agent | 2026-02-25 | Proposed |
| Reviewer | | | |
| Approver | | | |
