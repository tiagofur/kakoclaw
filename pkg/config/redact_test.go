package config

import (
	"strings"
	"testing"
)

func TestIsSensitiveField(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		// Sensitive fields (should return true)
		{"api_key lowercase", "api_key", true},
		{"api_key uppercase", "API_KEY", true},
		{"api_key mixed case", "Api_Key", true},
		{"apikey no underscore", "apikey", true},
		{"token", "token", true},
		{"bot_token", "bot_token", true},
		{"app_token", "app_token", true},
		{"access_token", "access_token", true},
		{"refresh_token", "refresh_token", true},
		{"password", "password", true},
		{"PASSWORD uppercase", "PASSWORD", true},
		{"secret", "secret", true},
		{"app_secret", "app_secret", true},
		{"client_secret", "client_secret", true},
		{"signing_secret", "signing_secret", true},
		{"private_key", "private_key", true},
		{"encrypt_key", "encrypt_key", true},
		{"verification_token", "verification_token", true},
		{"prefixed api_key", "my_api_key", true},
		{"prefixed password", "user_password", true},
		{"prefixed token", "auth_token", true},
		{"oauth_token", "oauth_token", true},

		// Non-sensitive fields (should return false)
		// Note: max_tokens is NOT sensitive because it doesn't match sensitive patterns
		// The word "token" only matches exact "token" or "*_token" suffix
		{"model", "model", false},
		{"enabled", "enabled", false},
		{"api_base", "api_base", false},
		{"host", "host", false},
		{"port", "port", false},
		{"temperature", "temperature", false},
		{"max_tokens", "max_tokens", false},
		{"allow_from", "allow_from", false},
		{"username", "username", false},
		{"workspace", "workspace", false},
		{"provider", "provider", false},
		{"proxy", "proxy", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSensitiveField(tt.field)
			if result != tt.expected {
				t.Errorf("IsSensitiveField(%q) = %v, want %v", tt.field, result, tt.expected)
			}
		})
	}
}

func TestRedactValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		// Empty value
		{"empty string", "", "[NOT SET]"},

		// Short values (should be fully redacted)
		{"single char", "a", "[REDACTED]"},
		{"two chars", "ab", "[REDACTED]"},
		{"three chars", "abc", "[REDACTED]"},
		{"four chars", "abcd", "[REDACTED]"},
		{"five chars", "abcde", "[REDACTED]"},
		{"six chars", "abcdef", "[REDACTED]"},
		{"seven chars", "abcdefg", "[REDACTED]"},
		{"eight chars", "abcdefgh", "[REDACTED]"},

		// Longer values (should show first 3 + **** + last 4)
		{"nine chars", "abcdefghi", "abc****fghi"},
		{"ten chars", "abcdefghij", "abc****ghij"},
		{"typical api key", "sk-proj-abc12345xyz", "sk-****5xyz"},
		{"openai key format", "sk-1234567890abcdef", "sk-****cdef"},
		{"long secret", "very-long-secret-key-value-here", "ver****here"},
		{"uuid style", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "a1b****7890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactValue(tt.value)
			if result != tt.expected {
				t.Errorf("RedactValue(%q) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestRedactConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Sensitive paths
		{"provider api_key", "providers.openai.api_key", true},
		{"channel token", "channels.telegram.token", true},
		{"channel bot_token", "channels.slack.bot_token", true},
		{"channel app_token", "channels.slack.app_token", true},
		{"channel app_secret", "channels.feishu.app_secret", true},
		{"channel client_secret", "channels.dingtalk.client_secret", true},
		{"tool password", "tools.email.password", true},
		{"web search api_key", "tools.web.search.api_key", true},
		{"verification_token", "channels.feishu.verification_token", true},

		// Non-sensitive paths
		{"provider api_base", "providers.openai.api_base", false},
		{"channel enabled", "channels.telegram.enabled", false},
		{"channel allow_from", "channels.telegram.allow_from", false},
		{"agent model", "agents.defaults.model", false},
		{"agent temperature", "agents.defaults.temperature", false},
		{"channel host", "channels.maixcam.host", false},
		{"channel port", "channels.maixcam.port", false},
		{"tool host", "tools.email.host", false},
		{"tool username", "tools.email.username", false},

		// Edge cases
		{"empty path", "", false},
		{"single segment", "model", false},
		{"single segment token", "token", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactConfigPath(tt.path)
			if result != tt.expected {
				t.Errorf("RedactConfigPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestRedactConfig(t *testing.T) {
	// Create a config with sensitive values
	cfg := &Config{
		Providers: ProvidersConfig{
			OpenAI: ProviderConfig{
				APIKey:  "sk-test-secret-key-12345",
				APIBase: "https://api.openai.com/v1",
			},
			Anthropic: ProviderConfig{
				APIKey: "ant-test-key-67890",
			},
		},
		Channels: ChannelsConfig{
			Telegram: TelegramConfig{
				Enabled: true,
				Token:   "telegram-bot-token-abc123",
				Proxy:   "http://proxy.example.com",
			},
			Slack: SlackConfig{
				Enabled:  true,
				BotToken: "xoxb-slack-bot-token",
				AppToken: "xapp-slack-app-token",
			},
			Feishu: FeishuConfig{
				Enabled:           true,
				AppID:             "feishu-app-id-12345",
				AppSecret:         "feishu-app-secret-abc",
				EncryptKey:        "feishu-encrypt-key-xyz",
				VerificationToken: "feishu-verify-token-def",
			},
		},
		Web: WebConfig{
			Enabled:  true,
			Host:     "localhost",
			Port:     8080,
			Username: "admin",
			Password: "super-secret-password",
		},
		Tools: ToolsConfig{
			Web: WebToolsConfig{
				Search: WebSearchConfig{
					APIKey:     "brave-api-key-search",
					MaxResults: 10,
				},
			},
			Email: EmailToolsConfig{
				Enabled:  true,
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user@example.com",
				Password: "email-password-secret",
			},
		},
	}

	redacted := RedactConfig(cfg)

	// Verify sensitive values are redacted
	t.Run("provider api keys redacted", func(t *testing.T) {
		if strings.Contains(redacted.Providers.OpenAI.APIKey, "secret") {
			t.Error("OpenAI API key should be redacted")
		}
		if redacted.Providers.OpenAI.APIKey == "" {
			t.Error("OpenAI API key should show redacted form, not empty")
		}
		if strings.Contains(redacted.Providers.Anthropic.APIKey, "test") {
			t.Error("Anthropic API key should be redacted")
		}
	})

	t.Run("non-sensitive values preserved", func(t *testing.T) {
		if redacted.Providers.OpenAI.APIBase != "https://api.openai.com/v1" {
			t.Errorf("APIBase should be preserved, got %s", redacted.Providers.OpenAI.APIBase)
		}
	})

	t.Run("channel tokens redacted", func(t *testing.T) {
		if strings.Contains(redacted.Channels.Telegram.Token, "telegram") {
			t.Error("Telegram token should be redacted")
		}
		if strings.Contains(redacted.Channels.Slack.BotToken, "slack") {
			t.Error("Slack bot token should be redacted")
		}
		if strings.Contains(redacted.Channels.Slack.AppToken, "slack") {
			t.Error("Slack app token should be redacted")
		}
	})

	t.Run("channel non-sensitive preserved", func(t *testing.T) {
		if !redacted.Channels.Telegram.Enabled {
			t.Error("Telegram enabled should be preserved")
		}
		if redacted.Channels.Telegram.Proxy != "http://proxy.example.com" {
			t.Error("Telegram proxy should be preserved")
		}
	})

	t.Run("feishu secrets redacted", func(t *testing.T) {
		if strings.Contains(redacted.Channels.Feishu.AppSecret, "secret") {
			t.Error("Feishu app secret should be redacted")
		}
		if strings.Contains(redacted.Channels.Feishu.EncryptKey, "encrypt") {
			t.Error("Feishu encrypt key should be redacted")
		}
		if strings.Contains(redacted.Channels.Feishu.VerificationToken, "verify") {
			t.Error("Feishu verification token should be redacted")
		}
	})

	t.Run("web password redacted", func(t *testing.T) {
		if strings.Contains(redacted.Web.Password, "secret") {
			t.Error("Web password should be redacted")
		}
		if redacted.Web.Username != "admin" {
			t.Error("Web username should be preserved")
		}
	})

	t.Run("tool passwords redacted", func(t *testing.T) {
		if strings.Contains(redacted.Tools.Web.Search.APIKey, "api") {
			t.Error("Search API key should be redacted")
		}
		if strings.Contains(redacted.Tools.Email.Password, "password") {
			t.Error("Email password should be redacted")
		}
		if redacted.Tools.Email.Username != "user@example.com" {
			t.Error("Email username should be preserved")
		}
	})

	t.Run("original config unchanged", func(t *testing.T) {
		if cfg.Providers.OpenAI.APIKey != "sk-test-secret-key-12345" {
			t.Error("Original config should not be modified")
		}
	})
}

func TestRedactConfig_NilInput(t *testing.T) {
	result := RedactConfig(nil)
	if result != nil {
		t.Error("RedactConfig(nil) should return nil")
	}
}

func TestRedactConfig_EmptyValues(t *testing.T) {
	cfg := &Config{
		Providers: ProvidersConfig{
			OpenAI: ProviderConfig{
				APIKey:  "", // Empty
				APIBase: "https://api.openai.com/v1",
			},
		},
		Channels: ChannelsConfig{
			Telegram: TelegramConfig{
				Token: "", // Empty
			},
		},
	}

	redacted := RedactConfig(cfg)

	// Empty values should remain empty (not "[NOT SET]" since we use redactIfSet)
	if redacted.Providers.OpenAI.APIKey != "" {
		t.Errorf("Empty API key should remain empty after redaction, got %q", redacted.Providers.OpenAI.APIKey)
	}
	if redacted.Channels.Telegram.Token != "" {
		t.Errorf("Empty token should remain empty after redaction, got %q", redacted.Channels.Telegram.Token)
	}
}
