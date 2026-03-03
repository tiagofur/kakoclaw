package tools

import (
	"testing"
)

func TestSanitizeArguments(t *testing.T) {
	t.Run("nil input returns empty map", func(t *testing.T) {
		result := sanitizeArguments(nil)
		if result == nil {
			t.Fatal("expected non-nil map, got nil")
		}
		if len(result) != 0 {
			t.Fatalf("expected empty map, got %d entries", len(result))
		}
	})

	t.Run("safe fields pass through unchanged", func(t *testing.T) {
		args := map[string]interface{}{
			"query":   "hello world",
			"count":   float64(5),
			"path":    "/tmp/file.txt",
			"content": "some text",
		}
		result := sanitizeArguments(args)
		for key, expected := range args {
			if result[key] != expected {
				t.Errorf("key %q: expected %v, got %v", key, expected, result[key])
			}
		}
	})

	t.Run("sensitive keys are redacted", func(t *testing.T) {
		sensitiveKeys := []string{
			"password", "token", "secret", "api_key", "apikey",
			"access_token", "refresh_token", "private_key",
		}

		for _, key := range sensitiveKeys {
			args := map[string]interface{}{
				key: "super-secret-value",
			}
			result := sanitizeArguments(args)
			if result[key] != "[REDACTED]" {
				t.Errorf("key %q should be redacted, got %v", key, result[key])
			}
		}
	})

	t.Run("case insensitive redaction", func(t *testing.T) {
		tests := []struct {
			key string
		}{
			{"PASSWORD"},
			{"Token"},
			{"API_KEY"},
			{"Secret"},
			{"Access_Token"},
			{"REFRESH_TOKEN"},
			{"Private_Key"},
			{"ApiKey"},
		}

		for _, tt := range tests {
			args := map[string]interface{}{
				tt.key: "secret-value",
			}
			result := sanitizeArguments(args)
			if result[tt.key] != "[REDACTED]" {
				t.Errorf("key %q should be redacted (case-insensitive), got %v", tt.key, result[tt.key])
			}
		}
	})

	t.Run("keys containing sensitive substrings are redacted", func(t *testing.T) {
		tests := []struct {
			key string
		}{
			{"my_password_field"},
			{"user_token_value"},
			{"the_secret_sauce"},
			{"provider_api_key"},
		}

		for _, tt := range tests {
			args := map[string]interface{}{
				tt.key: "sensitive-data",
			}
			result := sanitizeArguments(args)
			if result[tt.key] != "[REDACTED]" {
				t.Errorf("key %q should be redacted (contains sensitive substring), got %v", tt.key, result[tt.key])
			}
		}
	})

	t.Run("config path sensitivity redacts value", func(t *testing.T) {
		args := map[string]interface{}{
			"path":  "providers.openai.api_key",
			"value": "sk-12345",
		}
		result := sanitizeArguments(args)
		if result["value"] != "[REDACTED]" {
			t.Errorf("value for sensitive config path should be redacted, got %v", result["value"])
		}
		// path itself is not a sensitive key name, so should pass through
		if result["path"] != "providers.openai.api_key" {
			t.Errorf("path should not be altered, got %v", result["path"])
		}
	})

	t.Run("config path safe keeps value", func(t *testing.T) {
		args := map[string]interface{}{
			"path":  "agents.defaults.model",
			"value": "gpt-4",
		}
		result := sanitizeArguments(args)
		if result["value"] != "gpt-4" {
			t.Errorf("value for safe config path should be preserved, got %v", result["value"])
		}
	})

	t.Run("config path with sensitive last segment redacts value", func(t *testing.T) {
		paths := []string{
			"channels.telegram.bot_token",
			"providers.claude.api_key",
			"channels.slack.app_secret",
			"providers.openai.access_token",
		}

		for _, path := range paths {
			args := map[string]interface{}{
				"path":  path,
				"value": "some-secret",
			}
			result := sanitizeArguments(args)
			if result["value"] != "[REDACTED]" {
				t.Errorf("value for sensitive config path %q should be redacted, got %v", path, result["value"])
			}
		}
	})

	t.Run("config path without value field does not panic", func(t *testing.T) {
		args := map[string]interface{}{
			"path": "providers.openai.api_key",
		}
		result := sanitizeArguments(args)
		if result["path"] != "providers.openai.api_key" {
			t.Errorf("path should be preserved, got %v", result["path"])
		}
	})

	t.Run("empty map returns empty map", func(t *testing.T) {
		result := sanitizeArguments(map[string]interface{}{})
		if len(result) != 0 {
			t.Fatalf("expected empty map, got %d entries", len(result))
		}
	})

	t.Run("mixed safe and sensitive keys", func(t *testing.T) {
		args := map[string]interface{}{
			"query":    "search term",
			"password": "secret123",
			"count":    float64(10),
			"token":    "tok_abc",
		}
		result := sanitizeArguments(args)
		if result["query"] != "search term" {
			t.Errorf("expected query to pass through, got %v", result["query"])
		}
		if result["password"] != "[REDACTED]" {
			t.Errorf("expected password to be redacted, got %v", result["password"])
		}
		if result["count"] != float64(10) {
			t.Errorf("expected count to pass through, got %v", result["count"])
		}
		if result["token"] != "[REDACTED]" {
			t.Errorf("expected token to be redacted, got %v", result["token"])
		}
	})
}

func TestIsSensitiveConfigPath(t *testing.T) {
	t.Run("sensitive paths", func(t *testing.T) {
		tests := []struct {
			path string
		}{
			{"providers.openai.api_key"},
			{"channels.telegram.bot_token"},
			{"channels.slack.app_secret"},
			{"providers.claude.access_token"},
			{"channels.discord.password"},
			{"providers.some.refresh_token"},
			{"channels.feishu.app_token"},
			{"auth.client_secret"},
			{"security.encrypt_key"},
			{"channels.slack.verification_token"},
			{"auth.private_key"},
		}

		for _, tt := range tests {
			t.Run(tt.path, func(t *testing.T) {
				if !isSensitiveConfigPath(tt.path) {
					t.Errorf("expected %q to be sensitive", tt.path)
				}
			})
		}
	})

	t.Run("safe paths", func(t *testing.T) {
		tests := []struct {
			path string
		}{
			{"agents.defaults.model"},
			{"storage.path"},
			{"web.port"},
			{"agents.orchestrator.enabled"},
			{"channels.telegram.enabled"},
			{"gateway.debug"},
			{"tools.exec.workspace"},
		}

		for _, tt := range tests {
			t.Run(tt.path, func(t *testing.T) {
				if isSensitiveConfigPath(tt.path) {
					t.Errorf("expected %q to be safe", tt.path)
				}
			})
		}
	})

	t.Run("empty path", func(t *testing.T) {
		if isSensitiveConfigPath("") {
			t.Error("empty path should not be sensitive")
		}
	})

	t.Run("single segment sensitive", func(t *testing.T) {
		if !isSensitiveConfigPath("api_key") {
			t.Error("single segment 'api_key' should be sensitive")
		}
	})

	t.Run("single segment safe", func(t *testing.T) {
		if isSensitiveConfigPath("model") {
			t.Error("single segment 'model' should be safe")
		}
	})

	t.Run("sensitive word in middle but safe last segment", func(t *testing.T) {
		// The function only checks the LAST segment
		if isSensitiveConfigPath("providers.api_key.enabled") {
			t.Error("path with sensitive word in middle but safe last segment should not be sensitive")
		}
	})
}

func TestIsRestrictedTool(t *testing.T) {
	t.Run("restricted tools", func(t *testing.T) {
		restrictedNames := []string{
			"exec", "spawn", "email", "write_file",
			"edit_file", "append_file", "web_fetch", "configure",
		}

		for _, name := range restrictedNames {
			t.Run(name, func(t *testing.T) {
				if !IsRestrictedTool(name) {
					t.Errorf("expected %q to be restricted", name)
				}
			})
		}
	})

	t.Run("safe tools", func(t *testing.T) {
		safeNames := []string{
			"read_file", "list_dir", "web_search",
			"task_manager", "query_knowledge", "message",
		}

		for _, name := range safeNames {
			t.Run(name, func(t *testing.T) {
				if IsRestrictedTool(name) {
					t.Errorf("expected %q to be safe (not restricted)", name)
				}
			})
		}
	})

	t.Run("empty string", func(t *testing.T) {
		if IsRestrictedTool("") {
			t.Error("empty string should not be restricted")
		}
	})

	t.Run("unknown tool", func(t *testing.T) {
		if IsRestrictedTool("nonexistent_tool") {
			t.Error("unknown tool should not be restricted")
		}
	})
}
