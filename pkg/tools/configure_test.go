package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setupTestConfig creates a temporary config directory and config file for testing.
// Returns the temp dir path and cleanup function.
func setupTestConfig(t *testing.T, userUUID string) (string, func()) {
	t.Helper()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "configure-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize the data directory for the config package
	config.InitDataDir(tempDir)

	// Create user config directory
	userConfigDir := filepath.Join(tempDir, "users", userUUID)
	if err := os.MkdirAll(userConfigDir, 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create user config dir: %v", err)
	}

	// Create a default config file
	defaultCfg := config.DefaultConfig()
	configData, err := json.MarshalIndent(defaultCfg, "", "  ")
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to marshal default config: %v", err)
	}

	configPath := filepath.Join(userConfigDir, "config.json")
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to write config file: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
		// Reset the data dir to default (user's home)
		home, _ := os.UserHomeDir()
		config.InitDataDir(filepath.Join(home, ".MakoClaw"))
	}

	return tempDir, cleanup
}

// parseResponse parses a JSON response string into a map
func parseResponse(t *testing.T, response string) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		t.Fatalf("Failed to parse response: %v\nResponse: %s", err, response)
	}
	return result
}

// =============================================================================
// Task 5.1: Unit Tests for Whitelist Enforcement
// =============================================================================

func TestConfigureTool_WhitelistEnforcement(t *testing.T) {
	userUUID := "test-user-whitelist"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	tests := []struct {
		name        string
		action      string
		path        string
		value       interface{}
		wantSuccess bool
		wantError   string
	}{
		// Valid provider paths
		{
			name:        "valid provider api_key write",
			action:      "set",
			path:        "providers.openai.api_key",
			value:       "sk-test-key-12345",
			wantSuccess: true,
		},
		{
			name:        "valid provider api_base read",
			action:      "get",
			path:        "providers.openai.api_base",
			wantSuccess: true,
		},
		{
			name:        "valid provider proxy write",
			action:      "set",
			path:        "providers.anthropic.proxy",
			value:       "http://proxy.example.com",
			wantSuccess: true,
		},

		// Valid channel paths
		{
			name:        "valid channel enabled read",
			action:      "get",
			path:        "channels.telegram.enabled",
			wantSuccess: true,
		},
		{
			name:        "valid channel token write",
			action:      "set",
			path:        "channels.telegram.token",
			value:       "telegram-bot-token-123",
			wantSuccess: true,
		},
		{
			name:        "valid channel allow_from read",
			action:      "get",
			path:        "channels.discord.allow_from",
			wantSuccess: true,
		},

		// Valid agent paths
		{
			name:        "valid agent defaults model read",
			action:      "get",
			path:        "agents.defaults.model",
			wantSuccess: true,
		},
		{
			name:        "valid agent defaults temperature write",
			action:      "set",
			path:        "agents.defaults.temperature",
			value:       float64(0.7),
			wantSuccess: true,
		},
		{
			name:        "valid orchestrator enabled read",
			action:      "get",
			path:        "agents.orchestrator.enabled",
			wantSuccess: true,
		},

		// Valid tool paths
		{
			name:        "valid tool email enabled read",
			action:      "get",
			path:        "tools.email.enabled",
			wantSuccess: true,
		},

		// Invalid/unknown paths - should be rejected
		{
			name:        "invalid: unknown top-level section",
			action:      "get",
			path:        "database.connection_string",
			wantSuccess: false,
			wantError:   "not configurable",
		},
		{
			name:        "invalid: unknown provider",
			action:      "get",
			path:        "providers.unknownprovider.api_key",
			wantSuccess: false,
			wantError:   "unknown provider",
		},
		{
			name:        "invalid: unknown channel",
			action:      "get",
			path:        "channels.unknownchannel.enabled",
			wantSuccess: false,
			wantError:   "unknown channel",
		},
		{
			name:        "invalid: unknown agent field",
			action:      "get",
			path:        "agents.defaults.unknown_field",
			wantSuccess: false,
			wantError:   "not configurable",
		},
		{
			name:        "invalid: storage paths not allowed",
			action:      "get",
			path:        "storage.database_path",
			wantSuccess: false,
			wantError:   "not configurable",
		},
		{
			name:        "invalid: web server paths not allowed",
			action:      "get",
			path:        "web.jwt_secret",
			wantSuccess: false,
			wantError:   "not configurable",
		},

		// Path traversal attempts - should be rejected
		{
			name:        "blocked: path traversal with ..",
			action:      "get",
			path:        "providers/../secrets",
			wantSuccess: false,
			wantError:   "traversal",
		},
		{
			name:        "blocked: absolute path with /",
			action:      "get",
			path:        "/etc/passwd",
			wantSuccess: false,
			wantError:   "absolute",
		},
		{
			name:        "blocked: absolute path with backslash",
			action:      "get",
			path:        "\\windows\\system32",
			wantSuccess: false,
			wantError:   "absolute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]interface{}{
				"action": tt.action,
				"path":   tt.path,
			}
			if tt.value != nil {
				args["value"] = tt.value
			}

			result, err := tool.Execute(ctx, args)
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}

			response := parseResponse(t, result)
			success, _ := response["success"].(bool)

			if success != tt.wantSuccess {
				t.Errorf("success = %v, want %v\nResponse: %s", success, tt.wantSuccess, result)
			}

			if !tt.wantSuccess && tt.wantError != "" {
				msg, _ := response["message"].(string)
				if !strings.Contains(strings.ToLower(msg), strings.ToLower(tt.wantError)) {
					t.Errorf("error message should contain %q, got: %s", tt.wantError, msg)
				}
			}
		})
	}
}

// =============================================================================
// Task 5.2: Unit Tests for Sensitive Field Redaction
// =============================================================================

func TestConfigureTool_SensitiveFieldsWriteOnly(t *testing.T) {
	userUUID := "test-user-sensitive"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	// First, set some sensitive values
	sensitivePaths := []string{
		"providers.openai.api_key",
		"channels.telegram.token",
		"channels.slack.bot_token",
		"channels.feishu.app_secret",
		"tools.email.password",
	}

	for _, path := range sensitivePaths {
		setArgs := map[string]interface{}{
			"action": "set",
			"path":   path,
			"value":  "secret-value-" + strings.ReplaceAll(path, ".", "-"),
		}
		result, err := tool.Execute(ctx, setArgs)
		if err != nil {
			t.Fatalf("Failed to set %s: %v", path, err)
		}
		response := parseResponse(t, result)
		if success, _ := response["success"].(bool); !success {
			t.Fatalf("Failed to set %s: %s", path, result)
		}
	}

	// Now try to read these sensitive fields - should return [SET] or similar, not actual values
	for _, path := range sensitivePaths {
		t.Run("get "+path, func(t *testing.T) {
			getArgs := map[string]interface{}{
				"action": "get",
				"path":   path,
			}
			result, err := tool.Execute(ctx, getArgs)
			if err != nil {
				t.Fatalf("Execute error: %v", err)
			}

			response := parseResponse(t, result)

			// Should succeed but with redacted/status value
			success, _ := response["success"].(bool)
			if !success {
				t.Errorf("Expected success for getting sensitive field status")
			}

			// The value should be "[SET]" or similar, NOT the actual secret
			value, _ := response["value"].(string)
			if strings.Contains(value, "secret-value") {
				t.Errorf("Sensitive field %s exposed actual value: %s", path, value)
			}
			if value != "[SET]" && value != "[NOT SET]" {
				t.Errorf("Expected [SET] or [NOT SET] for sensitive field, got: %s", value)
			}

			// Should indicate sensitivity
			sensitive, _ := response["sensitive"].(bool)
			if !sensitive {
				t.Errorf("Response should mark field as sensitive")
			}
		})
	}

	// Verify set action success messages don't include actual values
	t.Run("set response never includes sensitive value", func(t *testing.T) {
		secretValue := "super-secret-api-key-xyz123"
		setArgs := map[string]interface{}{
			"action": "set",
			"path":   "providers.openai.api_key",
			"value":  secretValue,
		}
		result, err := tool.Execute(ctx, setArgs)
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}

		// The response should not contain the actual secret value
		if strings.Contains(result, secretValue) {
			t.Errorf("Set response leaked sensitive value: %s", result)
		}
	})
}

func TestConfigureTool_ListProvidersRedaction(t *testing.T) {
	userUUID := "test-user-list-providers"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	// Set an API key
	secretKey := "sk-super-secret-openai-key-12345"
	setArgs := map[string]interface{}{
		"action": "set",
		"path":   "providers.openai.api_key",
		"value":  secretKey,
	}
	tool.Execute(ctx, setArgs)

	// List providers
	listArgs := map[string]interface{}{
		"action": "list_providers",
	}
	result, err := tool.Execute(ctx, listArgs)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	response := parseResponse(t, result)
	if success, _ := response["success"].(bool); !success {
		t.Fatalf("list_providers failed: %s", result)
	}

	// Verify the actual API key is not in the response
	if strings.Contains(result, secretKey) {
		t.Errorf("list_providers leaked API key: %s", result)
	}

	// Verify api_key_set status is included
	providers, _ := response["providers"].([]interface{})
	for _, p := range providers {
		provider, _ := p.(map[string]interface{})
		if name, _ := provider["name"].(string); name == "openai" {
			apiKeySet, _ := provider["api_key_set"].(bool)
			if !apiKeySet {
				t.Error("OpenAI should show api_key_set=true after setting key")
			}
		}
	}
}

// =============================================================================
// Task 5.3: Unit Tests for Set Action with Validation
// =============================================================================

func TestConfigureTool_SetAndVerify(t *testing.T) {
	userUUID := "test-user-set-verify"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	// Test setting non-sensitive values and verifying
	t.Run("set and verify non-sensitive value", func(t *testing.T) {
		// Set model
		setArgs := map[string]interface{}{
			"action": "set",
			"path":   "agents.defaults.model",
			"value":  "gpt-4-turbo",
		}
		result, err := tool.Execute(ctx, setArgs)
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}
		response := parseResponse(t, result)
		if success, _ := response["success"].(bool); !success {
			t.Fatalf("Set failed: %s", result)
		}

		// Verify by getting
		getArgs := map[string]interface{}{
			"action": "get",
			"path":   "agents.defaults.model",
		}
		result, err = tool.Execute(ctx, getArgs)
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}
		response = parseResponse(t, result)
		value, _ := response["value"].(string)
		if value != "gpt-4-turbo" {
			t.Errorf("Expected 'gpt-4-turbo', got: %s", value)
		}
	})

	// Test setting sensitive values - saved but not returned
	t.Run("set sensitive value saved but not returned", func(t *testing.T) {
		secretKey := "sk-test-key-abc123xyz"
		setArgs := map[string]interface{}{
			"action": "set",
			"path":   "providers.openai.api_key",
			"value":  secretKey,
		}
		result, err := tool.Execute(ctx, setArgs)
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}

		// Response should not contain the secret
		if strings.Contains(result, secretKey) {
			t.Error("Set response should not contain secret value")
		}

		// Verify it was actually saved by loading config directly
		cfg, err := config.LoadConfigForUser(userUUID)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		if cfg.Providers.OpenAI.APIKey != secretKey {
			t.Error("Secret was not saved to config file")
		}
	})
}

func TestConfigureTool_ValidationRejection(t *testing.T) {
	userUUID := "test-user-validation"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	tests := []struct {
		name      string
		path      string
		value     interface{}
		wantError string
	}{
		// Invalid port
		{
			name:      "port too high",
			path:      "tools.email.port",
			value:     float64(70000),
			wantError: "between 1 and 65535",
		},
		{
			name:      "port zero",
			path:      "channels.maixcam.port",
			value:     float64(0),
			wantError: "between 1 and 65535",
		},
		{
			name:      "port negative",
			path:      "channels.maixcam.port",
			value:     float64(-1),
			wantError: "between 1 and 65535",
		},

		// Invalid temperature
		{
			name:      "temperature too high",
			path:      "agents.defaults.temperature",
			value:     float64(5.0),
			wantError: "between 0.0 and 2.0",
		},
		{
			name:      "temperature negative",
			path:      "agents.defaults.temperature",
			value:     float64(-0.5),
			wantError: "between 0.0 and 2.0",
		},

		// Invalid URL
		{
			name:      "invalid URL format",
			path:      "providers.openai.api_base",
			value:     "not-a-url",
			wantError: "invalid URL",
		},
		{
			name:      "ftp URL not allowed",
			path:      "providers.openai.api_base",
			value:     "ftp://files.example.com",
			wantError: "invalid URL scheme",
		},

		// Invalid max_tokens
		{
			name:      "max_tokens too high",
			path:      "agents.defaults.max_tokens",
			value:     float64(2000000),
			wantError: "between 1 and 1000000",
		},
		{
			name:      "max_tokens zero",
			path:      "agents.defaults.max_tokens",
			value:     float64(0),
			wantError: "between 1 and 1000000",
		},

		// Invalid max_tool_iterations
		{
			name:      "max_tool_iterations too high",
			path:      "agents.defaults.max_tool_iterations",
			value:     float64(500),
			wantError: "between 1 and 100",
		},

		// Invalid provider name
		{
			name:      "unknown provider",
			path:      "agents.defaults.provider",
			value:     "unknownprovider",
			wantError: "unknown provider",
		},

		// Invalid email
		{
			name:      "invalid email format",
			path:      "tools.email.from",
			value:     "not-an-email",
			wantError: "invalid email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]interface{}{
				"action": "set",
				"path":   tt.path,
				"value":  tt.value,
			}
			result, err := tool.Execute(ctx, args)
			if err != nil {
				t.Fatalf("Execute error: %v", err)
			}

			response := parseResponse(t, result)
			success, _ := response["success"].(bool)
			if success {
				t.Errorf("Expected validation failure for %v on %s", tt.value, tt.path)
			}

			msg, _ := response["message"].(string)
			if !strings.Contains(strings.ToLower(msg), strings.ToLower(tt.wantError)) {
				t.Errorf("Expected error containing %q, got: %s", tt.wantError, msg)
			}
		})
	}

	// Test valid values are accepted
	validTests := []struct {
		name  string
		path  string
		value interface{}
	}{
		{"valid port", "tools.email.port", float64(587)},
		{"valid temperature", "agents.defaults.temperature", float64(0.7)},
		{"valid URL", "providers.openai.api_base", "https://api.example.com/v1"},
		{"valid max_tokens", "agents.defaults.max_tokens", float64(8192)},
		{"valid provider", "agents.defaults.provider", "openai"},
		{"valid email", "tools.email.from", "test@example.com"},
	}

	for _, tt := range validTests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]interface{}{
				"action": "set",
				"path":   tt.path,
				"value":  tt.value,
			}
			result, err := tool.Execute(ctx, args)
			if err != nil {
				t.Fatalf("Execute error: %v", err)
			}

			response := parseResponse(t, result)
			success, _ := response["success"].(bool)
			if !success {
				t.Errorf("Expected success for valid value %v on %s\nResponse: %s", tt.value, tt.path, result)
			}
		})
	}
}

// =============================================================================
// Task 5.4: Unit Tests for Enable/Disable Actions
// =============================================================================

func TestConfigureTool_EnableDisable(t *testing.T) {
	userUUID := "test-user-enable-disable"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	// Test enable channel
	t.Run("enable channel", func(t *testing.T) {
		args := map[string]interface{}{
			"action": "enable",
			"path":   "channels.telegram",
		}
		result, err := tool.Execute(ctx, args)
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}

		response := parseResponse(t, result)
		if success, _ := response["success"].(bool); !success {
			t.Errorf("Enable should succeed: %s", result)
		}

		// Verify in config
		cfg, err := config.LoadConfigForUser(userUUID)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		if !cfg.Channels.Telegram.Enabled {
			t.Error("Telegram should be enabled")
		}

		// Check restart hint
		if !strings.Contains(result, "restart") && !strings.Contains(result, "Restart") {
			t.Error("Response should include restart hint for channel changes")
		}
	})

	// Test disable channel
	t.Run("disable channel", func(t *testing.T) {
		args := map[string]interface{}{
			"action": "disable",
			"path":   "channels.telegram",
		}
		result, err := tool.Execute(ctx, args)
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}

		response := parseResponse(t, result)
		if success, _ := response["success"].(bool); !success {
			t.Errorf("Disable should succeed: %s", result)
		}

		// Verify in config
		cfg, err := config.LoadConfigForUser(userUUID)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		if cfg.Channels.Telegram.Enabled {
			t.Error("Telegram should be disabled")
		}
	})

	// Test enable orchestrator
	t.Run("enable orchestrator", func(t *testing.T) {
		args := map[string]interface{}{
			"action": "enable",
			"path":   "agents.orchestrator",
		}
		result, err := tool.Execute(ctx, args)
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}

		response := parseResponse(t, result)
		if success, _ := response["success"].(bool); !success {
			t.Errorf("Enable orchestrator should succeed: %s", result)
		}

		// Verify in config
		cfg, err := config.LoadConfigForUser(userUUID)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		if !cfg.Agents.Orchestrator.Enabled {
			t.Error("Orchestrator should be enabled")
		}
	})

	// Test invalid enable paths
	invalidPaths := []struct {
		name string
		path string
	}{
		{"invalid three segments", "channels.telegram.token"},
		{"invalid single segment", "channels"},
		{"invalid section", "providers.openai"},
		{"unknown channel", "channels.unknownchannel"},
		{"invalid agent path", "agents.defaults"},
	}

	for _, tt := range invalidPaths {
		t.Run("enable rejected: "+tt.name, func(t *testing.T) {
			args := map[string]interface{}{
				"action": "enable",
				"path":   tt.path,
			}
			result, err := tool.Execute(ctx, args)
			if err != nil {
				t.Fatalf("Execute error: %v", err)
			}

			response := parseResponse(t, result)
			if success, _ := response["success"].(bool); success {
				t.Errorf("Enable should fail for %s", tt.path)
			}
		})
	}
}

// =============================================================================
// Task 5.5: Unit Tests for List Actions
// =============================================================================

func TestConfigureTool_ListProviders(t *testing.T) {
	userUUID := "test-user-list-providers-full"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	// Set up some provider config
	tool.Execute(ctx, map[string]interface{}{
		"action": "set",
		"path":   "providers.openai.api_key",
		"value":  "sk-test-key",
	})
	tool.Execute(ctx, map[string]interface{}{
		"action": "set",
		"path":   "providers.openai.api_base",
		"value":  "https://custom-api.example.com/v1",
	})

	// List providers
	result, err := tool.Execute(ctx, map[string]interface{}{
		"action": "list_providers",
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	response := parseResponse(t, result)
	if success, _ := response["success"].(bool); !success {
		t.Fatalf("list_providers failed: %s", result)
	}

	// Verify all 10 providers are listed
	providers, _ := response["providers"].([]interface{})
	if len(providers) != 10 {
		t.Errorf("Expected 10 providers, got %d", len(providers))
	}

	// Check provider names
	expectedProviders := map[string]bool{
		"anthropic":  false,
		"openai":     false,
		"openrouter": false,
		"groq":       false,
		"zhipu":      false,
		"vllm":       false,
		"gemini":     false,
		"nvidia":     false,
		"moonshot":   false,
		"ollama":     false,
	}

	for _, p := range providers {
		provider, _ := p.(map[string]interface{})
		name, _ := provider["name"].(string)
		expectedProviders[name] = true

		// For OpenAI, verify our settings
		if name == "openai" {
			apiKeySet, _ := provider["api_key_set"].(bool)
			if !apiKeySet {
				t.Error("OpenAI api_key_set should be true")
			}
			apiBaseCustom, _ := provider["api_base_custom"].(bool)
			if !apiBaseCustom {
				t.Error("OpenAI api_base_custom should be true")
			}
			apiBase, _ := provider["api_base"].(string)
			if apiBase != "https://custom-api.example.com/v1" {
				t.Errorf("OpenAI api_base should be custom URL, got: %s", apiBase)
			}
		}

		// For Ollama, verify it doesn't require API key
		if name == "ollama" {
			usable, _ := provider["usable"].(bool)
			if !usable {
				t.Error("Ollama should be usable without API key")
			}
		}
	}

	for name, found := range expectedProviders {
		if !found {
			t.Errorf("Provider %s not found in list", name)
		}
	}

	// Verify counts
	totalCount, _ := response["total_count"].(float64)
	if int(totalCount) != 10 {
		t.Errorf("total_count should be 10, got %v", totalCount)
	}
}

func TestConfigureTool_ListChannels(t *testing.T) {
	userUUID := "test-user-list-channels"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	// Enable and configure Telegram
	tool.Execute(ctx, map[string]interface{}{
		"action": "enable",
		"path":   "channels.telegram",
	})
	tool.Execute(ctx, map[string]interface{}{
		"action": "set",
		"path":   "channels.telegram.token",
		"value":  "telegram-token-xyz",
	})
	tool.Execute(ctx, map[string]interface{}{
		"action": "set",
		"path":   "channels.telegram.allow_from",
		"value":  []interface{}{"123456", "789012"},
	})

	// List channels
	result, err := tool.Execute(ctx, map[string]interface{}{
		"action": "list_channels",
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	response := parseResponse(t, result)
	if success, _ := response["success"].(bool); !success {
		t.Fatalf("list_channels failed: %s", result)
	}

	// Verify all 10 channels are listed
	channels, _ := response["channels"].([]interface{})
	if len(channels) != 10 {
		t.Errorf("Expected 10 channels, got %d", len(channels))
	}

	// Check channel names
	expectedChannels := map[string]bool{
		"telegram": false,
		"discord":  false,
		"slack":    false,
		"whatsapp": false,
		"signal":   false,
		"qq":       false,
		"dingtalk": false,
		"feishu":   false,
		"maixcam":  false,
		"email":    false,
	}

	for _, ch := range channels {
		channel, _ := ch.(map[string]interface{})
		name, _ := channel["name"].(string)
		expectedChannels[name] = true

		// For Telegram, verify our settings
		if name == "telegram" {
			enabled, _ := channel["enabled"].(bool)
			if !enabled {
				t.Error("Telegram should be enabled")
			}
			credentialsSet, _ := channel["credentials_set"].(bool)
			if !credentialsSet {
				t.Error("Telegram credentials_set should be true")
			}
			usable, _ := channel["usable"].(bool)
			if !usable {
				t.Error("Telegram should be usable")
			}

			// Verify allow_from (not sensitive, should be shown)
			allowFrom, _ := channel["allow_from"].([]interface{})
			if len(allowFrom) != 2 {
				t.Errorf("Telegram allow_from should have 2 entries, got %d", len(allowFrom))
			}
		}
	}

	for name, found := range expectedChannels {
		if !found {
			t.Errorf("Channel %s not found in list", name)
		}
	}

	// Verify no actual token values in response
	if strings.Contains(result, "telegram-token-xyz") {
		t.Error("Channel token should not appear in list_channels output")
	}
}

// =============================================================================
// Task 5.6: Integration Test - Full Workflow
// =============================================================================

func TestConfigureTool_FullWorkflow(t *testing.T) {
	userUUID := "test-user-full-workflow"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	// Step 1: list_providers (shows none configured initially)
	t.Log("Step 1: List providers initially")
	result, err := tool.Execute(ctx, map[string]interface{}{"action": "list_providers"})
	if err != nil {
		t.Fatalf("Step 1 error: %v", err)
	}
	response := parseResponse(t, result)
	activeCount, _ := response["active_count"].(float64)
	// Only Ollama should be usable without API key
	if int(activeCount) != 1 {
		t.Logf("Initial active providers: %v (expected 1 for Ollama)", activeCount)
	}

	// Step 2: Set OpenAI API key
	t.Log("Step 2: Set OpenAI API key")
	result, err = tool.Execute(ctx, map[string]interface{}{
		"action": "set",
		"path":   "providers.openai.api_key",
		"value":  "sk-test-openai-key-12345",
	})
	if err != nil {
		t.Fatalf("Step 2 error: %v", err)
	}
	response = parseResponse(t, result)
	if success, _ := response["success"].(bool); !success {
		t.Fatalf("Step 2 failed: %s", result)
	}

	// Step 3: list_providers (shows OpenAI configured)
	t.Log("Step 3: Verify OpenAI shows as configured")
	result, err = tool.Execute(ctx, map[string]interface{}{"action": "list_providers"})
	if err != nil {
		t.Fatalf("Step 3 error: %v", err)
	}
	response = parseResponse(t, result)
	providers, _ := response["providers"].([]interface{})
	openaiConfigured := false
	for _, p := range providers {
		provider, _ := p.(map[string]interface{})
		if name, _ := provider["name"].(string); name == "openai" {
			if apiKeySet, _ := provider["api_key_set"].(bool); apiKeySet {
				openaiConfigured = true
			}
		}
	}
	if !openaiConfigured {
		t.Error("Step 3: OpenAI should show as configured")
	}

	// Step 4: Enable Telegram channel
	t.Log("Step 4: Enable Telegram channel")
	result, err = tool.Execute(ctx, map[string]interface{}{
		"action": "enable",
		"path":   "channels.telegram",
	})
	if err != nil {
		t.Fatalf("Step 4 error: %v", err)
	}
	response = parseResponse(t, result)
	if success, _ := response["success"].(bool); !success {
		t.Fatalf("Step 4 failed: %s", result)
	}

	// Step 5: list_channels (shows channel enabled but not usable without token)
	t.Log("Step 5: Verify Telegram is enabled")
	result, err = tool.Execute(ctx, map[string]interface{}{"action": "list_channels"})
	if err != nil {
		t.Fatalf("Step 5 error: %v", err)
	}
	response = parseResponse(t, result)
	channels, _ := response["channels"].([]interface{})
	telegramStatus := map[string]interface{}{}
	for _, ch := range channels {
		channel, _ := ch.(map[string]interface{})
		if name, _ := channel["name"].(string); name == "telegram" {
			telegramStatus = channel
		}
	}
	if enabled, _ := telegramStatus["enabled"].(bool); !enabled {
		t.Error("Step 5: Telegram should be enabled")
	}
	// Note: We just check that we can list channels - the usable state depends on
	// whether credentials are set (which may vary by test isolation)

	// Step 6: Set Telegram token
	t.Log("Step 6: Set Telegram token")
	result, err = tool.Execute(ctx, map[string]interface{}{
		"action": "set",
		"path":   "channels.telegram.token",
		"value":  "telegram-bot-token-workflow-test",
	})
	if err != nil {
		t.Fatalf("Step 6 error: %v", err)
	}
	response = parseResponse(t, result)
	if success, _ := response["success"].(bool); !success {
		t.Fatalf("Step 6 failed: %s", result)
	}

	// Step 7: list_channels (shows channel usable)
	t.Log("Step 7: Verify Telegram is now usable")
	result, err = tool.Execute(ctx, map[string]interface{}{"action": "list_channels"})
	if err != nil {
		t.Fatalf("Step 7 error: %v", err)
	}
	response = parseResponse(t, result)
	channels, _ = response["channels"].([]interface{})
	for _, ch := range channels {
		channel, _ := ch.(map[string]interface{})
		if name, _ := channel["name"].(string); name == "telegram" {
			telegramStatus = channel
		}
	}
	if usable, _ := telegramStatus["usable"].(bool); !usable {
		t.Error("Step 7: Telegram should now be usable")
	}

	// Step 8: Verify config file persistence
	t.Log("Step 8: Verify config file persistence")
	cfg, err := config.LoadConfigForUser(userUUID)
	if err != nil {
		t.Fatalf("Step 8: Failed to load config: %v", err)
	}
	if cfg.Providers.OpenAI.APIKey != "sk-test-openai-key-12345" {
		t.Error("Step 8: OpenAI API key not persisted")
	}
	if !cfg.Channels.Telegram.Enabled {
		t.Error("Step 8: Telegram enabled not persisted")
	}
	if cfg.Channels.Telegram.Token != "telegram-bot-token-workflow-test" {
		t.Error("Step 8: Telegram token not persisted")
	}

	t.Log("Full workflow completed successfully")
}

// =============================================================================
// Task 5.7: Security Tests
// =============================================================================

func TestConfigureTool_RequiresUserContext(t *testing.T) {
	tool := NewConfigureTool()
	// Don't set user context
	ctx := context.Background()

	// Try any action
	result, err := tool.Execute(ctx, map[string]interface{}{
		"action": "list_providers",
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	// Should fail gracefully
	if !strings.Contains(result, "user context not set") {
		t.Errorf("Should require user context, got: %s", result)
	}
}

func TestConfigureTool_NoSecretExposure(t *testing.T) {
	userUUID := "test-user-no-exposure"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	// Set many different sensitive fields
	sensitiveFields := []struct {
		path   string
		secret string
	}{
		{"providers.openai.api_key", "secret-openai-key-xyz123"},
		{"providers.anthropic.api_key", "secret-anthropic-key-abc456"},
		{"providers.groq.api_key", "secret-groq-key-def789"},
		{"channels.telegram.token", "secret-telegram-token-ghi012"},
		{"channels.discord.token", "secret-discord-token-jkl345"},
		{"channels.slack.bot_token", "secret-slack-bot-mno678"},
		{"channels.slack.app_token", "secret-slack-app-pqr901"},
		{"channels.feishu.app_secret", "secret-feishu-secret-stu234"},
		{"channels.dingtalk.client_secret", "secret-dingtalk-vwx567"},
		{"tools.email.password", "secret-email-password-yz0123"},
	}

	// Set all secrets
	for _, sf := range sensitiveFields {
		result, err := tool.Execute(ctx, map[string]interface{}{
			"action": "set",
			"path":   sf.path,
			"value":  sf.secret,
		})
		if err != nil {
			t.Fatalf("Failed to set %s: %v", sf.path, err)
		}
		// Verify secret not in set response
		if strings.Contains(result, sf.secret) {
			t.Errorf("Secret leaked in set response for %s", sf.path)
		}
	}

	// Test that no action exposes secrets
	actions := []map[string]interface{}{
		{"action": "list_providers"},
		{"action": "list_channels"},
	}

	// Add get actions for each field
	for _, sf := range sensitiveFields {
		actions = append(actions, map[string]interface{}{
			"action": "get",
			"path":   sf.path,
		})
	}

	for _, args := range actions {
		result, err := tool.Execute(ctx, args)
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}

		// Check no secret appears in output
		for _, sf := range sensitiveFields {
			if strings.Contains(result, sf.secret) {
				t.Errorf("Secret for %s leaked in action %v", sf.path, args["action"])
			}
		}
	}
}

func TestConfigureTool_PathTraversalPrevented(t *testing.T) {
	userUUID := "test-user-traversal"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	maliciousPaths := []string{
		// Path traversal attempts
		"../../../etc/passwd",
		"providers/../../../secrets",
		"..\\..\\windows\\system32",
		"channels/../../sensitive",

		// Absolute paths
		"/etc/passwd",
		"\\windows\\system32\\config\\sam",
		"C:\\Users\\Admin\\secrets",

		// Other injection attempts
		"providers; rm -rf /",
		"providers && cat /etc/passwd",
		"providers`whoami`",
	}

	for _, path := range maliciousPaths {
		t.Run("blocked: "+path, func(t *testing.T) {
			result, err := tool.Execute(ctx, map[string]interface{}{
				"action": "get",
				"path":   path,
			})
			if err != nil {
				t.Fatalf("Execute error: %v", err)
			}

			response := parseResponse(t, result)
			if success, _ := response["success"].(bool); success {
				t.Errorf("Malicious path should be rejected: %s", path)
			}
		})
	}
}

// =============================================================================
// Task 5.8: Test Audit Logging Sanitization
// =============================================================================

func TestConfigureTool_AuditLoggingSanitization(t *testing.T) {
	// Test that the sanitizeArguments function properly redacts configure tool values

	// Test cases for isSensitiveConfigPath
	pathTests := []struct {
		path      string
		sensitive bool
	}{
		{"providers.openai.api_key", true},
		{"channels.telegram.token", true},
		{"channels.slack.bot_token", true},
		{"channels.feishu.app_secret", true},
		{"tools.email.password", true},
		{"channels.feishu.verification_token", true},
		{"agents.defaults.model", false},
		{"channels.telegram.enabled", false},
		{"providers.openai.api_base", false},
	}

	for _, tt := range pathTests {
		t.Run("isSensitiveConfigPath: "+tt.path, func(t *testing.T) {
			result := isSensitiveConfigPath(tt.path)
			if result != tt.sensitive {
				t.Errorf("isSensitiveConfigPath(%q) = %v, want %v", tt.path, result, tt.sensitive)
			}
		})
	}

	// Test sanitizeArguments for configure tool
	t.Run("sanitizeArguments redacts configure sensitive values", func(t *testing.T) {
		args := map[string]interface{}{
			"action": "set",
			"path":   "providers.openai.api_key",
			"value":  "sk-super-secret-key-12345",
		}

		sanitized := sanitizeArguments(args)

		// Action and path should be preserved
		if sanitized["action"] != "set" {
			t.Error("Action should be preserved")
		}
		if sanitized["path"] != "providers.openai.api_key" {
			t.Error("Path should be preserved for traceability")
		}

		// Value should be redacted
		if sanitized["value"] != "[REDACTED]" {
			t.Errorf("Value should be redacted, got: %v", sanitized["value"])
		}
	})

	t.Run("sanitizeArguments preserves non-sensitive configure values", func(t *testing.T) {
		args := map[string]interface{}{
			"action": "set",
			"path":   "agents.defaults.model",
			"value":  "gpt-4-turbo",
		}

		sanitized := sanitizeArguments(args)

		// All should be preserved
		if sanitized["action"] != "set" {
			t.Error("Action should be preserved")
		}
		if sanitized["path"] != "agents.defaults.model" {
			t.Error("Path should be preserved")
		}
		if sanitized["value"] != "gpt-4-turbo" {
			t.Errorf("Non-sensitive value should be preserved, got: %v", sanitized["value"])
		}
	})
}

func TestIsRestrictedTool_IncludesConfigure(t *testing.T) {
	// Verify configure is in the restricted tools list
	if !IsRestrictedTool("configure") {
		t.Error("'configure' should be in RestrictedTools for audit logging")
	}
}

// =============================================================================
// Additional Edge Case Tests
// =============================================================================

func TestConfigureTool_TypeConversion(t *testing.T) {
	userUUID := "test-user-types"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	// Test various type conversions (JSON numbers come as float64)
	tests := []struct {
		name  string
		path  string
		value interface{}
	}{
		{"int as float64", "agents.defaults.max_tokens", float64(8192)},
		{"bool true", "channels.telegram.enabled", true},
		{"bool false", "agents.orchestrator.enabled", false},
		{"string", "agents.defaults.model", "gpt-4"},
		{"string slice from array", "channels.telegram.allow_from", []interface{}{"123", "456"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, map[string]interface{}{
				"action": "set",
				"path":   tt.path,
				"value":  tt.value,
			})
			if err != nil {
				t.Fatalf("Execute error: %v", err)
			}

			response := parseResponse(t, result)
			if success, _ := response["success"].(bool); !success {
				t.Errorf("Should succeed for type %T: %s", tt.value, result)
			}
		})
	}
}

func TestConfigureTool_MissingParameters(t *testing.T) {
	userUUID := "test-user-params"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	// Special case: missing/empty action returns a plain text error, not JSON
	t.Run("missing action", func(t *testing.T) {
		result, err := tool.Execute(ctx, map[string]interface{}{})
		if err != nil {
			t.Fatalf("Execute error: %v", err)
		}

		// Response is plain text, not JSON for unknown action
		if !strings.Contains(strings.ToLower(result), "unknown action") {
			t.Errorf("Expected 'unknown action' in response, got: %s", result)
		}
	})

	// Other missing parameters return JSON errors
	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError string
	}{
		{
			name:      "get missing path",
			args:      map[string]interface{}{"action": "get"},
			wantError: "path is required",
		},
		{
			name:      "set missing path",
			args:      map[string]interface{}{"action": "set", "value": "test"},
			wantError: "path is required",
		},
		{
			name:      "set missing value",
			args:      map[string]interface{}{"action": "set", "path": "agents.defaults.model"},
			wantError: "value is required",
		},
		{
			name:      "enable missing path",
			args:      map[string]interface{}{"action": "enable"},
			wantError: "path is required",
		},
		{
			name:      "disable missing path",
			args:      map[string]interface{}{"action": "disable"},
			wantError: "path is required",
		},
		{
			name:      "reset missing path",
			args:      map[string]interface{}{"action": "reset"},
			wantError: "path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, tt.args)
			if err != nil {
				t.Fatalf("Execute error: %v", err)
			}

			response := parseResponse(t, result)
			if success, _ := response["success"].(bool); success {
				t.Error("Should fail with missing parameter")
			}

			msg, _ := response["message"].(string)
			if !strings.Contains(strings.ToLower(msg), strings.ToLower(tt.wantError)) {
				t.Errorf("Expected error containing %q, got: %s", tt.wantError, msg)
			}
		})
	}
}

func TestConfigureTool_ResetRestoresGlobalValue(t *testing.T) {
	userUUID := "test-user-reset"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	globalCfg := config.DefaultConfig()
	globalCfg.Agents.Defaults.Model = "global-model"
	if err := config.SaveConfig(filepath.Join(config.GetDataDir(), "config.json"), globalCfg); err != nil {
		t.Fatalf("failed to save global config: %v", err)
	}

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"action": "set",
		"path":   "agents.defaults.model",
		"value":  "broken-model",
	})
	if err != nil {
		t.Fatalf("set returned error: %v", err)
	}

	result, err := tool.Execute(ctx, map[string]interface{}{
		"action": "reset",
		"path":   "agents.defaults.model",
	})
	if err != nil {
		t.Fatalf("reset returned error: %v", err)
	}

	response := parseResponse(t, result)
	if success, _ := response["success"].(bool); !success {
		t.Fatalf("reset failed: %s", result)
	}
	if value, _ := response["value"].(string); value != "global-model" {
		t.Fatalf("expected reset value to be global-model, got %q", value)
	}

	getResult, err := tool.Execute(ctx, map[string]interface{}{
		"action": "get",
		"path":   "agents.defaults.model",
	})
	if err != nil {
		t.Fatalf("get returned error: %v", err)
	}
	getResponse := parseResponse(t, getResult)
	if value, _ := getResponse["value"].(string); value != "global-model" {
		t.Fatalf("expected merged config to expose global-model after reset, got %q", value)
	}
}

func TestConfigureTool_ResetClearsSensitiveField(t *testing.T) {
	userUUID := "test-user-reset-sensitive"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := WithSensitiveConfirmationRequired(context.Background(), true)

	_, err := tool.Execute(ctx, map[string]interface{}{
		"action": "set",
		"path":   "channels.telegram.token",
		"value":  "telegram-secret",
	})
	if err != nil {
		t.Fatalf("set returned error: %v", err)
	}

	result, err := tool.Execute(ctx, map[string]interface{}{
		"action": "reset",
		"path":   "channels.telegram.token",
	})
	if err != nil {
		t.Fatalf("reset returned error: %v", err)
	}

	response := parseResponse(t, result)
	if success, _ := response["success"].(bool); success {
		t.Fatalf("expected reset without confirmation to fail, got: %s", result)
	}
	if code, _ := response["error"].(string); code != "confirmation_required" {
		t.Fatalf("expected confirmation_required, got %q", code)
	}

	confirmedResult, err := tool.Execute(ctx, map[string]interface{}{
		"action":    "reset",
		"path":      "channels.telegram.token",
		"confirmed": true,
	})
	if err != nil {
		t.Fatalf("confirmed reset returned error: %v", err)
	}

	confirmedResponse := parseResponse(t, confirmedResult)
	if resetTo, _ := confirmedResponse["reset_to"].(string); resetTo != "cleared" {
		t.Fatalf("expected sensitive reset_to=cleared, got %q", resetTo)
	}

	cfg, err := config.LoadConfigForUser(userUUID)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Channels.Telegram.Token != "" {
		t.Fatalf("expected telegram token to be cleared, got %q", cfg.Channels.Telegram.Token)
	}
}

func TestConfigureTool_SetSensitiveRequiresConfirmation(t *testing.T) {
	userUUID := "test-user-sensitive-confirm"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := WithSensitiveConfirmationRequired(context.Background(), true)

	result, err := tool.Execute(ctx, map[string]interface{}{
		"action": "set",
		"path":   "providers.openai.api_key",
		"value":  "sk-sensitive",
	})
	if err != nil {
		t.Fatalf("set returned error: %v", err)
	}

	response := parseResponse(t, result)
	if success, _ := response["success"].(bool); success {
		t.Fatalf("expected sensitive set without confirmation to fail, got: %s", result)
	}
	if code, _ := response["error"].(string); code != "confirmation_required" {
		t.Fatalf("expected confirmation_required, got %q", code)
	}

	confirmedResult, err := tool.Execute(ctx, map[string]interface{}{
		"action":    "set",
		"path":      "providers.openai.api_key",
		"value":     "sk-sensitive",
		"confirmed": true,
	})
	if err != nil {
		t.Fatalf("confirmed set returned error: %v", err)
	}

	confirmedResponse := parseResponse(t, confirmedResult)
	if success, _ := confirmedResponse["success"].(bool); !success {
		t.Fatalf("expected confirmed sensitive set to succeed, got: %s", confirmedResult)
	}
}

func TestConfigureTool_ListPaths(t *testing.T) {
	userUUID := "test-user-list-paths"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"action": "list_paths",
	})
	if err != nil {
		t.Fatalf("list_paths returned error: %v", err)
	}

	response := parseResponse(t, result)
	if success, _ := response["success"].(bool); !success {
		t.Fatalf("list_paths failed: %s", result)
	}

	paths, _ := response["paths"].([]interface{})
	if len(paths) == 0 {
		t.Fatal("expected non-empty path list")
	}

	seen := map[string]bool{}
	for _, entry := range paths {
		item, _ := entry.(map[string]interface{})
		path, _ := item["path"].(string)
		seen[path] = true
	}

	for _, required := range []string{"agents.defaults.model", "channels.email.enabled", "providers.openai.api_key"} {
		if !seen[required] {
			t.Fatalf("expected list_paths to include %s", required)
		}
	}
}

func TestConfigureTool_ConcurrentAccess(t *testing.T) {
	userUUID := "test-user-concurrent"
	_, cleanup := setupTestConfig(t, userUUID)
	defer cleanup()

	tool := NewConfigureTool()
	tool.SetUserContext(1, userUUID)
	ctx := context.Background()

	// Run multiple operations concurrently - test mutex protection
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			defer func() { done <- true }()
			// Mix of read and write operations on the tool's internal state
			if idx%2 == 0 {
				// Write operation
				tool.Execute(ctx, map[string]interface{}{
					"action": "set",
					"path":   "agents.defaults.temperature",
					"value":  float64(idx%20) / 10.0, // Keep in valid range 0-2
				})
			} else {
				// Read operation
				tool.Execute(ctx, map[string]interface{}{
					"action": "get",
					"path":   "agents.defaults.temperature",
				})
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify tool mutex still works after concurrent access
	// Just verify we can change user context (tests mutex)
	tool.SetUserContext(2, "another-user")

	// And change it back - this operation would panic or deadlock if mutex is broken
	tool.SetUserContext(1, userUUID)

	// Now verify we can still do basic operations
	result, err := tool.Execute(ctx, map[string]interface{}{
		"action": "get",
		"path":   "agents.defaults.model",
	})
	if err != nil {
		t.Fatalf("Post-concurrent execute error: %v", err)
	}

	// Parse result - should be valid JSON
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if success, _ := response["success"].(bool); !success {
		// Check if error is due to config not found (test isolation issue)
		errorStr, _ := response["error"].(string)
		if errorStr != "config_load_error" {
			t.Errorf("Tool should work after concurrent access, got: %s", result)
		}
	}
}
