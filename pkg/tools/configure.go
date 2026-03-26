package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
)

// FieldType defines the type of a configuration field
type FieldType int

const (
	TypeString FieldType = iota
	TypeInt
	TypeFloat
	TypeBool
	TypeStringSlice
)

// ValidatorFn is a function that validates a field value
type ValidatorFn func(value interface{}) error

// FieldPolicy defines read/write permissions for a config field
type FieldPolicy struct {
	Type      FieldType
	Readable  bool
	Writable  bool
	Validator ValidatorFn
}

// fieldWhitelist defines all configurable fields and their policies
// Wildcard patterns: providers.* and channels.* are resolved at runtime
var fieldWhitelist = map[string]FieldPolicy{
	// ============================================================
	// PROVIDERS (10 providers x 5 fields = 50 patterns)
	// All providers follow the same field pattern
	// ============================================================
	"providers.*.api_key":     {Type: TypeString, Readable: false, Writable: true},
	"providers.*.api_base":    {Type: TypeString, Readable: true, Writable: true, Validator: validateURL},
	"providers.*.proxy":       {Type: TypeString, Readable: true, Writable: true, Validator: validateURL},
	"providers.*.auth_method": {Type: TypeString, Readable: true, Writable: true},
	"providers.*.models":      {Type: TypeStringSlice, Readable: true, Writable: true},

	// ============================================================
	// CHANNELS (9 channels with varying credential fields)
	// Common fields across all channels
	// ============================================================
	"channels.*.enabled":    {Type: TypeBool, Readable: true, Writable: true},
	"channels.*.allow_from": {Type: TypeStringSlice, Readable: true, Writable: true},

	// Channel-specific credential fields (write-only)
	"channels.*.token":              {Type: TypeString, Readable: false, Writable: true}, // telegram, discord
	"channels.*.bot_token":          {Type: TypeString, Readable: false, Writable: true}, // slack
	"channels.*.app_token":          {Type: TypeString, Readable: false, Writable: true}, // slack
	"channels.*.app_id":             {Type: TypeString, Readable: false, Writable: true}, // feishu, qq
	"channels.*.app_secret":         {Type: TypeString, Readable: false, Writable: true}, // feishu, qq
	"channels.*.client_id":          {Type: TypeString, Readable: false, Writable: true}, // dingtalk
	"channels.*.client_secret":      {Type: TypeString, Readable: false, Writable: true}, // dingtalk
	"channels.*.encrypt_key":        {Type: TypeString, Readable: false, Writable: true}, // feishu
	"channels.*.verification_token": {Type: TypeString, Readable: false, Writable: true}, // feishu

	// Channel-specific non-sensitive fields
	"channels.*.proxy":        {Type: TypeString, Readable: true, Writable: true, Validator: validateURL},  // telegram
	"channels.*.bridge_url":   {Type: TypeString, Readable: true, Writable: true, Validator: validateURL},  // whatsapp
	"channels.*.host":         {Type: TypeString, Readable: true, Writable: true},                          // maixcam
	"channels.*.port":         {Type: TypeInt, Readable: true, Writable: true, Validator: validatePort},    // maixcam
	"channels.*.phone_number": {Type: TypeString, Readable: true, Writable: true},                          // signal

	// ============================================================
	// AGENTS
	// Default agent configuration and orchestrator settings
	// ============================================================
	"agents.defaults.provider":            {Type: TypeString, Readable: true, Writable: true, Validator: validateProviderName},
	"agents.defaults.model":               {Type: TypeString, Readable: true, Writable: true},
	"agents.defaults.max_tokens":          {Type: TypeInt, Readable: true, Writable: true, Validator: validateMaxTokens},
	"agents.defaults.temperature":         {Type: TypeFloat, Readable: true, Writable: true, Validator: validateTemperature},
	"agents.defaults.max_tool_iterations": {Type: TypeInt, Readable: true, Writable: true, Validator: validateMaxIterations},

	"agents.orchestrator.enabled":                {Type: TypeBool, Readable: true, Writable: true},
	"agents.orchestrator.provider":               {Type: TypeString, Readable: true, Writable: true, Validator: validateProviderName},
	"agents.orchestrator.model":                  {Type: TypeString, Readable: true, Writable: true},
	"agents.orchestrator.max_tokens":             {Type: TypeInt, Readable: true, Writable: true, Validator: validateMaxTokens},
	"agents.orchestrator.temperature":            {Type: TypeFloat, Readable: true, Writable: true, Validator: validateTemperature},
	"agents.orchestrator.max_delegation_retries": {Type: TypeInt, Readable: true, Writable: true, Validator: validateMaxIterations},
	"agents.orchestrator.fallback_to_default":    {Type: TypeBool, Readable: true, Writable: true},

	// ============================================================
	// TOOLS
	// Web search and email tool configuration
	// ============================================================
	"tools.web.search.api_key":     {Type: TypeString, Readable: false, Writable: true},
	"tools.web.search.max_results": {Type: TypeInt, Readable: true, Writable: true, Validator: validateMaxResults},
	"tools.email.enabled":          {Type: TypeBool, Readable: true, Writable: true},
	"tools.email.host":             {Type: TypeString, Readable: true, Writable: true},
	"tools.email.port":             {Type: TypeInt, Readable: true, Writable: true, Validator: validatePort},
	"tools.email.username":         {Type: TypeString, Readable: true, Writable: true},
	"tools.email.password":         {Type: TypeString, Readable: false, Writable: true},
	"tools.email.from":             {Type: TypeString, Readable: true, Writable: true, Validator: validateEmail},
	"tools.email.to":               {Type: TypeString, Readable: true, Writable: true, Validator: validateEmail},

	// ============================================================
	// EMAIL CHANNEL (IMAP polling + SMTP reply)
	// ============================================================
	"channels.email.enabled":              {Type: TypeBool, Readable: true, Writable: true},
	"channels.email.imap_host":            {Type: TypeString, Readable: true, Writable: true},
	"channels.email.imap_port":            {Type: TypeInt, Readable: true, Writable: true, Validator: validatePort},
	"channels.email.smtp_host":            {Type: TypeString, Readable: true, Writable: true},
	"channels.email.smtp_port":            {Type: TypeInt, Readable: true, Writable: true, Validator: validatePort},
	"channels.email.username":             {Type: TypeString, Readable: true, Writable: true},
	"channels.email.password":             {Type: TypeString, Readable: false, Writable: true},
	"channels.email.from":                 {Type: TypeString, Readable: true, Writable: true},
	"channels.email.mailbox":              {Type: TypeString, Readable: true, Writable: true},
	"channels.email.poll_interval_seconds": {Type: TypeInt, Readable: true, Writable: true},
	"channels.email.mark_as_read":         {Type: TypeBool, Readable: true, Writable: true},
}

// providerNames lists all known LLM provider names
var providerNames = []string{
	"anthropic",
	"openai",
	"openrouter",
	"groq",
	"zhipu",
	"vllm",
	"gemini",
	"nvidia",
	"moonshot",
	"ollama",
}

// channelNames lists all known communication channel names
var channelNames = []string{
	"telegram",
	"discord",
	"slack",
	"whatsapp",
	"signal",
	"qq",
	"dingtalk",
	"feishu",
	"maixcam",
}

// sensitiveFieldPatterns lists field name patterns that contain secrets
// Fields matching these patterns are write-only (cannot be read)
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

// Name returns the tool name
func (t *ConfigureTool) Name() string {
	return "configure"
}

// Description returns the tool description
func (t *ConfigureTool) Description() string {
	return `Manage user configuration for providers, channels, and agent settings.
Actions: get, set, enable, disable, list_providers, list_channels.
Sensitive fields (API keys, tokens) are write-only and never displayed.

IMPORTANT: Paths use dot-notation with the full section prefix. Examples:
  Providers: providers.openai.api_key, providers.anthropic.api_base, providers.groq.api_key
  Channels:  channels.telegram.token, channels.telegram.enabled, channels.telegram.allow_from
             channels.discord.token, channels.slack.bot_token, channels.slack.app_token
             channels.whatsapp.bridge_url, channels.signal.phone_number
             channels.feishu.app_id, channels.feishu.app_secret
             channels.qq.app_id, channels.dingtalk.client_id
             channels.email.enabled, channels.email.imap_host, channels.email.smtp_host
             channels.email.username, channels.email.from, channels.email.poll_interval_seconds
  Agents:    agents.defaults.provider, agents.defaults.model, agents.defaults.temperature
             agents.orchestrator.enabled, agents.orchestrator.model
  Tools:     tools.web.search.api_key, tools.email.enabled, tools.email.host

Note: Telegram uses 'token' (not 'bot_token'). Slack uses 'bot_token' and 'app_token'.
Use 'list_channels' or 'list_providers' to see current status before modifying.
For enable/disable, use path like 'channels.telegram' or 'agents.orchestrator'.`
}

// Parameters returns the tool parameter schema
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
				"description": "The value to set. Required for 'set' action. Type depends on field.",
			},
		},
		"required": []string{"action"},
	}
}

// SetUserContext sets the user context for config file access
func (t *ConfigureTool) SetUserContext(userID int64, userUUID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.userID = userID
	t.userUUID = userUUID
}

// Execute performs the configuration action
func (t *ConfigureTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	t.mu.RLock()
	userUUID := t.userUUID
	userID := t.userID
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
		return "Error: unknown action '" + action + "'. Valid actions: get, set, enable, disable, list_providers, list_channels", nil
	}
}

// =============================================================================
// Path Resolution and Field Access Helpers (Task 3.1-3.2)
// =============================================================================

// normalizePath attempts to fix common path mistakes:
// - "telegram.token" → "channels.telegram.token"
// - "openai.api_key" → "providers.openai.api_key"
// - "telegram.bot_token" → "channels.telegram.token" (Telegram uses "token", not "bot_token")
func (t *ConfigureTool) normalizePath(path string) string {
	parts := strings.Split(path, ".")

	// Single segment: check if it's a known channel or provider name
	if len(parts) == 1 {
		for _, ch := range channelNames {
			if parts[0] == ch {
				return "channels." + path
			}
		}
		for _, p := range providerNames {
			if parts[0] == p {
				return "providers." + path
			}
		}
		return path
	}

	section := parts[0]

	// Already has a valid section prefix
	if section == "providers" || section == "channels" || section == "agents" || section == "tools" {
		// Fix Telegram-specific: channels.telegram.bot_token → channels.telegram.token
		if section == "channels" && len(parts) >= 3 && parts[1] == "telegram" && parts[2] == "bot_token" {
			parts[2] = "token"
			return strings.Join(parts, ".")
		}
		return path
	}

	// Check if first segment is a known channel name (e.g., "telegram.token")
	for _, ch := range channelNames {
		if section == ch {
			normalized := "channels." + path
			// Fix Telegram-specific: telegram.bot_token → channels.telegram.token
			if ch == "telegram" && len(parts) >= 2 && parts[1] == "bot_token" {
				normalized = "channels.telegram.token"
			}
			return normalized
		}
	}

	// Check if first segment is a known provider name (e.g., "openai.api_key")
	for _, p := range providerNames {
		if section == p {
			return "providers." + path
		}
	}

	return path
}

// resolveFieldPolicy looks up the policy for a given path, resolving wildcards.
// Paths like "providers.openai.api_key" are matched against "providers.*.api_key"
func (t *ConfigureTool) resolveFieldPolicy(path string) (*FieldPolicy, error) {
	// Normalize path to fix common mistakes (missing prefix, wrong field names)
	path = t.normalizePath(path)

	// Validate path format first
	if err := config.ValidateConfigPath(path); err != nil {
		return nil, err
	}

	// Direct lookup first
	if policy, ok := fieldWhitelist[path]; ok {
		return &policy, nil
	}

	// Try wildcard patterns for providers.* and channels.*
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("path '%s' is not configurable", path)
	}

	section := parts[0]
	name := parts[1]

	// Build wildcard path: replace second segment with *
	var wildcardPath string
	if len(parts) >= 3 {
		// providers.openai.api_key -> providers.*.api_key
		wildcardPath = section + ".*." + strings.Join(parts[2:], ".")
	} else {
		// providers.openai -> providers.*
		wildcardPath = section + ".*"
	}

	if policy, ok := fieldWhitelist[wildcardPath]; ok {
		// Validate the specific name based on section
		switch section {
		case "providers":
			if err := config.ValidateProviderName(name); err != nil {
				return nil, err
			}
		case "channels":
			if err := config.ValidateChannelName(name); err != nil {
				return nil, err
			}
		}
		return &policy, nil
	}

	return nil, fmt.Errorf("path '%s' is not configurable", path)
}

// getFieldValue retrieves a field value from config using reflection.
// Handles nested paths like "providers.openai.api_key" or "agents.defaults.model"
func (t *ConfigureTool) getFieldValue(cfg *config.Config, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	val := reflect.ValueOf(cfg).Elem()

	for _, part := range parts {
		if !val.IsValid() {
			return nil, fmt.Errorf("invalid path segment '%s'", part)
		}

		if val.Kind() == reflect.Struct {
			// Convert path segment to struct field name (snake_case to CamelCase)
			fieldName := t.toFieldName(part)
			field := val.FieldByName(fieldName)
			if !field.IsValid() {
				// Try alternative naming conventions
				field = val.FieldByNameFunc(func(name string) bool {
					return strings.EqualFold(name, fieldName) ||
						strings.EqualFold(name, part) ||
						strings.EqualFold(name, strings.ReplaceAll(part, "_", ""))
				})
			}
			if !field.IsValid() {
				return nil, fmt.Errorf("field '%s' not found in struct", part)
			}
			val = field
		} else if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return nil, fmt.Errorf("nil pointer at '%s'", part)
			}
			val = val.Elem()
		} else {
			return nil, fmt.Errorf("cannot navigate into %s at '%s'", val.Kind(), part)
		}
	}

	return val.Interface(), nil
}

// setFieldValue sets a field value in config using reflection.
// Handles type conversions for JSON-style values (float64 -> int, etc.)
func (t *ConfigureTool) setFieldValue(cfg *config.Config, path string, value interface{}) error {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	val := reflect.ValueOf(cfg).Elem()

	// Navigate to the parent of the target field
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return fmt.Errorf("nil pointer at '%s'", part)
			}
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			return fmt.Errorf("expected struct at '%s', got %s", part, val.Kind())
		}

		fieldName := t.toFieldName(part)
		field := val.FieldByName(fieldName)
		if !field.IsValid() {
			field = val.FieldByNameFunc(func(name string) bool {
				return strings.EqualFold(name, fieldName) ||
					strings.EqualFold(name, part) ||
					strings.EqualFold(name, strings.ReplaceAll(part, "_", ""))
			})
		}
		if !field.IsValid() {
			return fmt.Errorf("field '%s' not found", part)
		}
		val = field
	}

	// Set the final field
	finalPart := parts[len(parts)-1]
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return fmt.Errorf("nil pointer at '%s'", finalPart)
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", val.Kind())
	}

	fieldName := t.toFieldName(finalPart)
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		field = val.FieldByNameFunc(func(name string) bool {
			return strings.EqualFold(name, fieldName) ||
				strings.EqualFold(name, finalPart) ||
				strings.EqualFold(name, strings.ReplaceAll(finalPart, "_", ""))
		})
	}
	if !field.IsValid() {
		return fmt.Errorf("field '%s' not found", finalPart)
	}
	if !field.CanSet() {
		return fmt.Errorf("field '%s' cannot be set", finalPart)
	}

	// Type conversion and assignment
	return t.setReflectValue(field, value, path)
}

// setReflectValue handles type conversion and sets the reflect value
func (t *ConfigureTool) setReflectValue(field reflect.Value, value interface{}, path string) error {
	switch field.Kind() {
	case reflect.String:
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for '%s', got %T", path, value)
		}
		field.SetString(s)

	case reflect.Int, reflect.Int64:
		switch v := value.(type) {
		case float64:
			field.SetInt(int64(v))
		case int:
			field.SetInt(int64(v))
		case int64:
			field.SetInt(v)
		default:
			return fmt.Errorf("expected number for '%s', got %T", path, value)
		}

	case reflect.Float64:
		switch v := value.(type) {
		case float64:
			field.SetFloat(v)
		case int:
			field.SetFloat(float64(v))
		case int64:
			field.SetFloat(float64(v))
		default:
			return fmt.Errorf("expected number for '%s', got %T", path, value)
		}

	case reflect.Bool:
		switch v := value.(type) {
		case bool:
			field.SetBool(v)
		case string:
			// Accept string "true"/"false"
			b := strings.EqualFold(v, "true")
			field.SetBool(b)
		default:
			return fmt.Errorf("expected boolean for '%s', got %T", path, value)
		}

	case reflect.Slice:
		// Handle []string and FlexibleStringSlice
		return t.setSliceValue(field, value, path)

	default:
		return fmt.Errorf("unsupported field type %s for '%s'", field.Kind(), path)
	}

	return nil
}

// setSliceValue handles slice type assignment ([]string, FlexibleStringSlice)
func (t *ConfigureTool) setSliceValue(field reflect.Value, value interface{}, path string) error {
	var strSlice []string

	switch v := value.(type) {
	case []interface{}:
		strSlice = make([]string, len(v))
		for i, item := range v {
			strSlice[i] = fmt.Sprint(item)
		}
	case []string:
		strSlice = v
	case string:
		// Accept comma-separated string
		strSlice = strings.Split(v, ",")
		for i := range strSlice {
			strSlice[i] = strings.TrimSpace(strSlice[i])
		}
	default:
		return fmt.Errorf("expected array for '%s', got %T", path, value)
	}

	// Check if it's FlexibleStringSlice or regular []string
	if field.Type() == reflect.TypeOf(config.FlexibleStringSlice{}) {
		field.Set(reflect.ValueOf(config.FlexibleStringSlice(strSlice)))
	} else {
		field.Set(reflect.ValueOf(strSlice))
	}

	return nil
}

// toFieldName converts snake_case path segment to Go struct field name (CamelCase)
func (t *ConfigureTool) toFieldName(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// validateValueType checks if a value matches the expected FieldType
func (t *ConfigureTool) validateValueType(value interface{}, expectedType FieldType) error {
	switch expectedType {
	case TypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case TypeInt:
		switch value.(type) {
		case int, int64, float64:
			// Accept numeric types (JSON numbers come as float64)
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	case TypeFloat:
		switch value.(type) {
		case int, int64, float64:
			// Accept numeric types
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case TypeBool:
		switch value.(type) {
		case bool:
			// OK
		case string:
			// Accept "true"/"false" strings
		default:
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case TypeStringSlice:
		switch value.(type) {
		case []interface{}, []string, string:
			// Accept arrays or comma-separated string
		default:
			return fmt.Errorf("expected array, got %T", value)
		}
	}
	return nil
}

// =============================================================================
// Action Handlers (Tasks 3.3-3.8)
// =============================================================================

// handleGet reads a configuration value with redaction for sensitive fields
func (t *ConfigureTool) handleGet(ctx context.Context, args map[string]interface{}, userUUID string) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return t.errorResponse("missing_path", "path is required for get action", "")
	}

	// Normalize path to fix common mistakes
	path = t.normalizePath(path)

	// Resolve and validate field policy
	policy, err := t.resolveFieldPolicy(path)
	if err != nil {
		return t.errorResponse("path_not_allowed", err.Error(), path)
	}

	// Check if field is readable (sensitive fields are write-only)
	if !policy.Readable {
		// For write-only fields, return [SET] or [NOT SET] status
		cfg, err := config.LoadMergedConfigForUser(userUUID)
		if err != nil {
			return t.errorResponse("config_load_error", fmt.Sprintf("failed to load config: %v", err), path)
		}

		value, err := t.getFieldValue(cfg, path)
		if err != nil {
			return t.errorResponse("field_access_error", err.Error(), path)
		}

		// Check if value is set
		strVal := fmt.Sprint(value)
		var status string
		if strVal == "" || strVal == "<nil>" {
			status = "[NOT SET]"
		} else {
			status = "[SET]"
		}

		return t.successResponse(map[string]interface{}{
			"path":      path,
			"value":     status,
			"sensitive": true,
			"hint":      "This field is write-only. Value cannot be displayed.",
		})
	}

	// Load merged config (user + global) for reading
	cfg, err := config.LoadMergedConfigForUser(userUUID)
	if err != nil {
		return t.errorResponse("config_load_error", fmt.Sprintf("failed to load config: %v", err), path)
	}

	value, err := t.getFieldValue(cfg, path)
	if err != nil {
		return t.errorResponse("field_access_error", err.Error(), path)
	}

	// Double-check sensitivity via path (extra safety)
	if config.RedactConfigPath(path) {
		strVal := fmt.Sprint(value)
		var status string
		if strVal == "" || strVal == "<nil>" {
			status = "[NOT SET]"
		} else {
			status = "[SET]"
		}
		return t.successResponse(map[string]interface{}{
			"path":      path,
			"value":     status,
			"sensitive": true,
		})
	}

	// Return the actual value
	return t.successResponse(map[string]interface{}{
		"path":  path,
		"value": value,
		"type":  t.fieldTypeName(policy.Type),
	})
}

// handleSet writes a configuration value with validation
func (t *ConfigureTool) handleSet(ctx context.Context, args map[string]interface{}, userUUID string, userID int64) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return t.errorResponse("missing_path", "path is required for set action", "")
	}

	// Normalize path to fix common mistakes
	path = t.normalizePath(path)

	value, hasValue := args["value"]
	if !hasValue {
		return t.errorResponse("missing_value", "value is required for set action", path)
	}

	// Resolve and validate field policy
	policy, err := t.resolveFieldPolicy(path)
	if err != nil {
		return t.errorResponse("path_not_allowed", err.Error(), path)
	}

	if !policy.Writable {
		return t.errorResponse("field_read_only", fmt.Sprintf("field '%s' is read-only", path), path)
	}

	// Validate value type
	if err := t.validateValueType(value, policy.Type); err != nil {
		return t.errorResponse("invalid_value_type", err.Error(), path)
	}

	// Run field-specific validator if present
	if policy.Validator != nil {
		if err := policy.Validator(value); err != nil {
			return t.errorResponse("validation_failed", err.Error(), path)
		}
	}

	// Load current config
	cfg, err := config.LoadConfigForUser(userUUID)
	if err != nil {
		return t.errorResponse("config_load_error", fmt.Sprintf("failed to load config: %v", err), path)
	}

	// Apply change
	if err := t.setFieldValue(cfg, path, value); err != nil {
		return t.errorResponse("field_set_error", err.Error(), path)
	}

	// Save config
	if err := config.SaveConfigForUser(userUUID, cfg); err != nil {
		return t.errorResponse("config_save_error", fmt.Sprintf("failed to save config: %v", err), path)
	}

	// Verify config loads successfully after save
	if _, err := config.LoadConfigForUser(userUUID); err != nil {
		return t.errorResponse("config_invalid", fmt.Sprintf("config validation failed after save: %v", err), path)
	}

	// Log the change (sensitive values are not logged)
	isSensitive := config.RedactConfigPath(path)
	logger.InfoCF("configure", "Configuration updated", map[string]interface{}{
		"user_id":   userID,
		"user_uuid": userUUID,
		"path":      path,
		"sensitive": isSensitive,
	})

	// Determine if restart is required (channel changes)
	restartRequired := strings.HasPrefix(path, "channels.")

	response := map[string]interface{}{
		"path":             path,
		"message":          fmt.Sprintf("Successfully updated %s", path),
		"restart_required": restartRequired,
	}
	if restartRequired {
		response["hint"] = "Restart the gateway for changes to take effect"
	}

	return t.successResponse(response)
}

// handleEnable enables a channel or feature
func (t *ConfigureTool) handleEnable(ctx context.Context, args map[string]interface{}, userUUID string, userID int64) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return t.errorResponse("missing_path", "path is required for enable action", "")
	}

	// Normalize path: e.g., "telegram" → "channels.telegram"
	path = t.normalizePath(path)

	// Validate path format: must be channels.{name} or agents.orchestrator
	parts := strings.Split(path, ".")
	if len(parts) != 2 {
		return t.errorResponse("invalid_path", "enable action requires path like 'channels.telegram' or 'agents.orchestrator'", path)
	}

	section := parts[0]
	name := parts[1]

	// Validate section and name
	switch section {
	case "channels":
		if err := config.ValidateChannelName(name); err != nil {
			return t.errorResponse("invalid_channel", err.Error(), path)
		}
	case "agents":
		if name != "orchestrator" {
			return t.errorResponse("invalid_path", "only 'agents.orchestrator' can be enabled/disabled", path)
		}
	default:
		return t.errorResponse("invalid_path", "enable action only works for channels and agents.orchestrator", path)
	}

	// Delegate to handleSet with enabled=true
	setArgs := map[string]interface{}{
		"path":  path + ".enabled",
		"value": true,
	}

	result, err := t.handleSet(ctx, setArgs, userUUID, userID)
	if err != nil {
		return result, err
	}

	// Parse and modify response for better messaging
	var response map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(result), &response); jsonErr == nil {
		if success, ok := response["success"].(bool); ok && success {
			response["message"] = fmt.Sprintf("%s enabled successfully", name)
			response["enabled"] = true
			if section == "channels" {
				response["hint"] = "Restart the gateway for changes to take effect"
			}
			return t.successResponse(response)
		}
	}

	return result, err
}

// handleDisable disables a channel or feature
func (t *ConfigureTool) handleDisable(ctx context.Context, args map[string]interface{}, userUUID string, userID int64) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return t.errorResponse("missing_path", "path is required for disable action", "")
	}

	// Normalize path: e.g., "telegram" → "channels.telegram"
	path = t.normalizePath(path)

	// Validate path format: must be channels.{name} or agents.orchestrator
	parts := strings.Split(path, ".")
	if len(parts) != 2 {
		return t.errorResponse("invalid_path", "disable action requires path like 'channels.telegram' or 'agents.orchestrator'", path)
	}

	section := parts[0]
	name := parts[1]

	// Validate section and name
	switch section {
	case "channels":
		if err := config.ValidateChannelName(name); err != nil {
			return t.errorResponse("invalid_channel", err.Error(), path)
		}
	case "agents":
		if name != "orchestrator" {
			return t.errorResponse("invalid_path", "only 'agents.orchestrator' can be enabled/disabled", path)
		}
	default:
		return t.errorResponse("invalid_path", "disable action only works for channels and agents.orchestrator", path)
	}

	// Delegate to handleSet with enabled=false
	setArgs := map[string]interface{}{
		"path":  path + ".enabled",
		"value": false,
	}

	result, err := t.handleSet(ctx, setArgs, userUUID, userID)
	if err != nil {
		return result, err
	}

	// Parse and modify response for better messaging
	var response map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(result), &response); jsonErr == nil {
		if success, ok := response["success"].(bool); ok && success {
			response["message"] = fmt.Sprintf("%s disabled successfully", name)
			response["enabled"] = false
			if section == "channels" {
				response["hint"] = "Restart the gateway for changes to take effect"
			}
			return t.successResponse(response)
		}
	}

	return result, err
}

// handleListProviders lists all providers with their configuration status
func (t *ConfigureTool) handleListProviders(ctx context.Context, userUUID string) (string, error) {
	cfg, err := config.LoadMergedConfigForUser(userUUID)
	if err != nil {
		return t.errorResponse("config_load_error", fmt.Sprintf("failed to load config: %v", err), "")
	}

	// Build provider status list
	providers := []map[string]interface{}{}

	// Helper to check provider status
	checkProvider := func(name string, p config.ProviderConfig, defaultBase string, requiresKey bool) map[string]interface{} {
		apiKeySet := p.APIKey != ""
		apiBase := p.APIBase
		if apiBase == "" {
			apiBase = defaultBase
		}
		apiBaseCustom := p.APIBase != "" && p.APIBase != defaultBase

		usable := true
		var note string
		if requiresKey && !apiKeySet {
			usable = false
			note = "Missing API key"
		}

		status := map[string]interface{}{
			"name":            name,
			"api_key_set":     apiKeySet,
			"api_base":        apiBase,
			"api_base_custom": apiBaseCustom,
			"usable":          usable,
		}
		if p.Proxy != "" {
			status["proxy"] = p.Proxy
		}
		if note != "" {
			status["note"] = note
		}
		return status
	}

	// Check all 10 providers
	providers = append(providers, checkProvider("anthropic", cfg.Providers.Anthropic, "https://api.anthropic.com", true))
	providers = append(providers, checkProvider("openai", cfg.Providers.OpenAI, "https://api.openai.com/v1", true))
	providers = append(providers, checkProvider("openrouter", cfg.Providers.OpenRouter, "https://openrouter.ai/api/v1", true))
	providers = append(providers, checkProvider("groq", cfg.Providers.Groq, "https://api.groq.com/openai/v1", true))
	providers = append(providers, checkProvider("zhipu", cfg.Providers.Zhipu, "", true))
	providers = append(providers, checkProvider("vllm", cfg.Providers.VLLM, "", true))
	providers = append(providers, checkProvider("gemini", cfg.Providers.Gemini, "", true))
	providers = append(providers, checkProvider("nvidia", cfg.Providers.Nvidia, "", true))
	providers = append(providers, checkProvider("moonshot", cfg.Providers.Moonshot, "", true))

	// Ollama doesn't require API key
	ollamaStatus := checkProvider("ollama", cfg.Providers.Ollama, "http://localhost:11434/v1", false)
	ollamaStatus["note"] = "Ollama does not require API key"
	providers = append(providers, ollamaStatus)

	// Count active providers
	activeCount := 0
	for _, p := range providers {
		if usable, ok := p["usable"].(bool); ok && usable {
			activeCount++
		}
	}

	return t.successResponse(map[string]interface{}{
		"providers":    providers,
		"active_count": activeCount,
		"total_count":  len(providers),
	})
}

// handleListChannels lists all channels with their configuration status
func (t *ConfigureTool) handleListChannels(ctx context.Context, userUUID string) (string, error) {
	cfg, err := config.LoadMergedConfigForUser(userUUID)
	if err != nil {
		return t.errorResponse("config_load_error", fmt.Sprintf("failed to load config: %v", err), "")
	}

	// Build channel status list
	channels := []map[string]interface{}{}

	// Telegram
	channels = append(channels, t.buildChannelStatus("telegram",
		cfg.Channels.Telegram.Enabled,
		[]credentialField{{"token", cfg.Channels.Telegram.Token}},
		cfg.Channels.Telegram.AllowFrom))

	// Discord
	channels = append(channels, t.buildChannelStatus("discord",
		cfg.Channels.Discord.Enabled,
		[]credentialField{{"token", cfg.Channels.Discord.Token}},
		cfg.Channels.Discord.AllowFrom))

	// Slack
	channels = append(channels, t.buildChannelStatus("slack",
		cfg.Channels.Slack.Enabled,
		[]credentialField{{"bot_token", cfg.Channels.Slack.BotToken}, {"app_token", cfg.Channels.Slack.AppToken}},
		cfg.Channels.Slack.AllowFrom))

	// WhatsApp (uses bridge_url, not sensitive)
	whatsappStatus := map[string]interface{}{
		"name":             "whatsapp",
		"enabled":          cfg.Channels.WhatsApp.Enabled,
		"bridge_url":       cfg.Channels.WhatsApp.BridgeURL,
		"credentials_set":  cfg.Channels.WhatsApp.BridgeURL != "",
		"credential_fields": []string{"bridge_url"},
		"allow_from":       []string(cfg.Channels.WhatsApp.AllowFrom),
		"usable":           cfg.Channels.WhatsApp.Enabled && cfg.Channels.WhatsApp.BridgeURL != "",
	}
	if !cfg.Channels.WhatsApp.Enabled {
		whatsappStatus["note"] = "Channel is disabled"
	} else if cfg.Channels.WhatsApp.BridgeURL == "" {
		whatsappStatus["note"] = "Missing bridge_url"
	}
	channels = append(channels, whatsappStatus)

	// Signal (uses phone_number, not sensitive)
	signalStatus := map[string]interface{}{
		"name":             "signal",
		"enabled":          cfg.Channels.Signal.Enabled,
		"phone_number":     cfg.Channels.Signal.PhoneNumber,
		"credentials_set":  cfg.Channels.Signal.PhoneNumber != "",
		"credential_fields": []string{"phone_number"},
		"allow_from":       []string(cfg.Channels.Signal.AllowFrom),
		"usable":           cfg.Channels.Signal.Enabled && cfg.Channels.Signal.PhoneNumber != "",
	}
	if !cfg.Channels.Signal.Enabled {
		signalStatus["note"] = "Channel is disabled"
	} else if cfg.Channels.Signal.PhoneNumber == "" {
		signalStatus["note"] = "Missing phone_number"
	}
	channels = append(channels, signalStatus)

	// Feishu
	channels = append(channels, t.buildChannelStatus("feishu",
		cfg.Channels.Feishu.Enabled,
		[]credentialField{
			{"app_id", cfg.Channels.Feishu.AppID},
			{"app_secret", cfg.Channels.Feishu.AppSecret},
		},
		cfg.Channels.Feishu.AllowFrom))

	// QQ
	channels = append(channels, t.buildChannelStatus("qq",
		cfg.Channels.QQ.Enabled,
		[]credentialField{
			{"app_id", cfg.Channels.QQ.AppID},
			{"app_secret", cfg.Channels.QQ.AppSecret},
		},
		cfg.Channels.QQ.AllowFrom))

	// DingTalk
	channels = append(channels, t.buildChannelStatus("dingtalk",
		cfg.Channels.DingTalk.Enabled,
		[]credentialField{
			{"client_id", cfg.Channels.DingTalk.ClientID},
			{"client_secret", cfg.Channels.DingTalk.ClientSecret},
		},
		cfg.Channels.DingTalk.AllowFrom))

	// MaixCam (no credentials, uses host/port)
	maixcamStatus := map[string]interface{}{
		"name":            "maixcam",
		"enabled":         cfg.Channels.MaixCam.Enabled,
		"host":            cfg.Channels.MaixCam.Host,
		"port":            cfg.Channels.MaixCam.Port,
		"credentials_set": true, // No credentials needed
		"credential_fields": []string{},
		"allow_from":      []string(cfg.Channels.MaixCam.AllowFrom),
		"usable":          cfg.Channels.MaixCam.Enabled,
	}
	if !cfg.Channels.MaixCam.Enabled {
		maixcamStatus["note"] = "Channel is disabled"
	}
	channels = append(channels, maixcamStatus)

	// Email (IMAP polling + SMTP reply)
	emailStatus := map[string]interface{}{
		"name":               "email",
		"enabled":            cfg.Channels.Email.Enabled,
		"imap_host":          cfg.Channels.Email.IMAPHost,
		"smtp_host":          cfg.Channels.Email.SMTPHost,
		"username":           cfg.Channels.Email.Username,
		"from":               cfg.Channels.Email.From,
		"mailbox":            cfg.Channels.Email.Mailbox,
		"poll_interval_secs": cfg.Channels.Email.PollIntervalSecs,
		"mark_as_read":       cfg.Channels.Email.MarkAsRead,
		"credentials_set":    cfg.Channels.Email.Username != "" && cfg.Channels.Email.Password != "",
		"credential_fields":  []string{"username", "password"},
		"allow_from":         []string(cfg.Channels.Email.AllowFrom),
		"usable":             cfg.Channels.Email.Enabled && cfg.Channels.Email.IMAPHost != "" && cfg.Channels.Email.Password != "",
	}
	if !cfg.Channels.Email.Enabled {
		emailStatus["note"] = "Channel is disabled"
	} else if cfg.Channels.Email.IMAPHost == "" {
		emailStatus["note"] = "Missing imap_host"
	} else if cfg.Channels.Email.Password == "" {
		emailStatus["note"] = "Missing password"
	}
	channels = append(channels, emailStatus)

	// Count active channels
	activeCount := 0
	for _, ch := range channels {
		if usable, ok := ch["usable"].(bool); ok && usable {
			activeCount++
		}
	}

	return t.successResponse(map[string]interface{}{
		"channels":     channels,
		"active_count": activeCount,
		"total_count":  len(channels),
	})
}

// credentialField represents a credential field name and value
type credentialField struct {
	name  string
	value string
}

// buildChannelStatus creates a channel status entry for channels with sensitive credentials
func (t *ConfigureTool) buildChannelStatus(name string, enabled bool, creds []credentialField, allowFrom config.FlexibleStringSlice) map[string]interface{} {
	// Check if all credentials are set
	allSet := true
	missingFields := []string{}
	credFieldNames := []string{}

	for _, cred := range creds {
		credFieldNames = append(credFieldNames, cred.name)
		if cred.value == "" {
			allSet = false
			missingFields = append(missingFields, cred.name)
		}
	}

	usable := enabled && allSet

	status := map[string]interface{}{
		"name":              name,
		"enabled":           enabled,
		"credentials_set":   allSet,
		"credential_fields": credFieldNames,
		"allow_from":        []string(allowFrom),
		"usable":            usable,
	}

	// Add note if not usable
	if !enabled {
		status["note"] = "Channel is disabled"
	} else if !allSet {
		status["note"] = fmt.Sprintf("Missing required credentials: %s", strings.Join(missingFields, ", "))
	}

	return status
}

// =============================================================================
// Response Helpers
// =============================================================================

// successResponse formats a successful response as JSON
func (t *ConfigureTool) successResponse(data map[string]interface{}) (string, error) {
	data["success"] = true
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// errorResponse formats an error response as JSON
func (t *ConfigureTool) errorResponse(errorCode, message, path string) (string, error) {
	data := map[string]interface{}{
		"success": false,
		"error":   errorCode,
		"message": message,
	}
	if path != "" {
		data["path"] = path
	}
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// fieldTypeName returns a human-readable name for a FieldType
func (t *ConfigureTool) fieldTypeName(ft FieldType) string {
	switch ft {
	case TypeString:
		return "string"
	case TypeInt:
		return "integer"
	case TypeFloat:
		return "number"
	case TypeBool:
		return "boolean"
	case TypeStringSlice:
		return "array"
	default:
		return "unknown"
	}
}

// =============================================================================
// Validators (wrappers around config.Validate* functions)
// =============================================================================

func validateURL(value interface{}) error {
	return config.ValidateURL(value)
}

func validatePort(value interface{}) error {
	return config.ValidatePort(value)
}

func validateProviderName(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}
	return config.ValidateProviderName(s)
}

func validateMaxTokens(value interface{}) error {
	return config.ValidateMaxTokens(value)
}

func validateTemperature(value interface{}) error {
	return config.ValidateTemperature(value)
}

func validateMaxIterations(value interface{}) error {
	return config.ValidateMaxIterations(value)
}

func validateMaxResults(value interface{}) error {
	// Similar to validateMaxIterations but with different range
	var n int
	switch v := value.(type) {
	case float64:
		n = int(v)
	case int:
		n = v
	case int64:
		n = int(v)
	default:
		return fmt.Errorf("expected number, got %T", value)
	}
	if n < 1 || n > 100 {
		return fmt.Errorf("max_results must be between 1 and 100, got %d", n)
	}
	return nil
}

func validateEmail(value interface{}) error {
	return config.ValidateEmail(value)
}
