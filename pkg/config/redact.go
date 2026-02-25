package config

import (
	"strings"
)

// SensitiveFieldSuffixes lists field name suffixes that indicate secrets.
// These are matched as suffixes (e.g., "api_key" matches "my_api_key" but not "api_keyring").
var SensitiveFieldSuffixes = []string{
	"api_key",
	"apikey",
	"_token",
	"password",
	"_secret",
	"private_key",
	"encrypt_key",
}

// SensitiveFieldExact lists field names that must match exactly (case-insensitive).
// These avoid false positives like "max_tokens" matching "token".
var SensitiveFieldExact = []string{
	"token",
	"secret",
}

// SensitiveFieldPatterns lists patterns that use substring matching.
// Be careful with these - they should be specific enough to avoid false positives.
var SensitiveFieldPatterns = []string{
	"bot_token",
	"app_token",
	"access_token",
	"refresh_token",
	"app_secret",
	"client_secret",
	"signing_secret",
	"verification_token",
}

// IsSensitiveField checks if a field name matches sensitive patterns.
// Uses a combination of exact matching, suffix matching, and pattern matching
// to avoid false positives like "max_tokens" being flagged as sensitive.
func IsSensitiveField(fieldName string) bool {
	lower := strings.ToLower(fieldName)

	// Check exact matches first (highest priority)
	for _, exact := range SensitiveFieldExact {
		if lower == exact {
			return true
		}
	}

	// Check suffix matches (e.g., "my_api_key" matches "_key" suffix)
	for _, suffix := range SensitiveFieldSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}

	// Check pattern matches (substring)
	for _, pattern := range SensitiveFieldPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// RedactValue returns a redacted representation of a sensitive value.
// Format:
//   - Empty string: returns "[NOT SET]"
//   - 1-4 characters: returns "[REDACTED]" (too short to show partial)
//   - 5+ characters: returns "sk-****{last4}" showing first 3 chars + **** + last 4 chars
func RedactValue(value string) string {
	if value == "" {
		return "[NOT SET]"
	}
	// For short values, we can't show enough to be useful without revealing too much
	if len(value) <= 8 {
		return "[REDACTED]"
	}
	// Show first 3 chars + **** + last 4 chars for longer values
	// e.g., "sk-proj-abc12345xyz" -> "sk-****5xyz"
	return value[:3] + "****" + value[len(value)-4:]
}

// RedactConfigPath extracts the field name from a dot-notation path
// and checks if it's sensitive.
// Example: "providers.openai.api_key" -> checks "api_key" -> returns true
func RedactConfigPath(path string) bool {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return false
	}
	fieldName := parts[len(parts)-1]
	return IsSensitiveField(fieldName)
}

// RedactConfig returns a copy of the Config with all sensitive fields masked.
// This is useful for logging or debugging config without exposing secrets.
func RedactConfig(cfg *Config) *Config {
	if cfg == nil {
		return nil
	}

	// Create a new Config struct (don't copy the mutex)
	redacted := &Config{
		Agents:          cfg.Agents,
		Gateway:         cfg.Gateway,
		ToolPermissions: cfg.ToolPermissions,
		Storage:         cfg.Storage,
		DegradedMode:    cfg.DegradedMode,
	}

	// Redact providers
	redacted.Providers = ProvidersConfig{
		Anthropic:  redactProvider(cfg.Providers.Anthropic),
		OpenAI:     redactProvider(cfg.Providers.OpenAI),
		OpenRouter: redactProvider(cfg.Providers.OpenRouter),
		Groq:       redactProvider(cfg.Providers.Groq),
		Zhipu:      redactProvider(cfg.Providers.Zhipu),
		VLLM:       redactProvider(cfg.Providers.VLLM),
		Gemini:     redactProvider(cfg.Providers.Gemini),
		Nvidia:     redactProvider(cfg.Providers.Nvidia),
		Moonshot:   redactProvider(cfg.Providers.Moonshot),
		Ollama:     redactProvider(cfg.Providers.Ollama),
	}

	// Redact channels
	redacted.Channels = ChannelsConfig{
		Telegram: TelegramConfig{
			Enabled:   cfg.Channels.Telegram.Enabled,
			Token:     redactIfSet(cfg.Channels.Telegram.Token),
			Proxy:     cfg.Channels.Telegram.Proxy,
			AllowFrom: cfg.Channels.Telegram.AllowFrom,
		},
		Discord: DiscordConfig{
			Enabled:   cfg.Channels.Discord.Enabled,
			Token:     redactIfSet(cfg.Channels.Discord.Token),
			AllowFrom: cfg.Channels.Discord.AllowFrom,
		},
		Slack: SlackConfig{
			Enabled:   cfg.Channels.Slack.Enabled,
			BotToken:  redactIfSet(cfg.Channels.Slack.BotToken),
			AppToken:  redactIfSet(cfg.Channels.Slack.AppToken),
			AllowFrom: cfg.Channels.Slack.AllowFrom,
		},
		WhatsApp: WhatsAppConfig{
			Enabled:   cfg.Channels.WhatsApp.Enabled,
			BridgeURL: cfg.Channels.WhatsApp.BridgeURL,
			AllowFrom: cfg.Channels.WhatsApp.AllowFrom,
		},
		Signal: SignalConfig{
			Enabled:     cfg.Channels.Signal.Enabled,
			PhoneNumber: cfg.Channels.Signal.PhoneNumber,
			AllowFrom:   cfg.Channels.Signal.AllowFrom,
		},
		Feishu: FeishuConfig{
			Enabled:           cfg.Channels.Feishu.Enabled,
			AppID:             redactIfSet(cfg.Channels.Feishu.AppID),
			AppSecret:         redactIfSet(cfg.Channels.Feishu.AppSecret),
			EncryptKey:        redactIfSet(cfg.Channels.Feishu.EncryptKey),
			VerificationToken: redactIfSet(cfg.Channels.Feishu.VerificationToken),
			AllowFrom:         cfg.Channels.Feishu.AllowFrom,
		},
		QQ: QQConfig{
			Enabled:   cfg.Channels.QQ.Enabled,
			AppID:     redactIfSet(cfg.Channels.QQ.AppID),
			AppSecret: redactIfSet(cfg.Channels.QQ.AppSecret),
			AllowFrom: cfg.Channels.QQ.AllowFrom,
		},
		DingTalk: DingTalkConfig{
			Enabled:      cfg.Channels.DingTalk.Enabled,
			ClientID:     redactIfSet(cfg.Channels.DingTalk.ClientID),
			ClientSecret: redactIfSet(cfg.Channels.DingTalk.ClientSecret),
			AllowFrom:    cfg.Channels.DingTalk.AllowFrom,
		},
		MaixCam: MaixCamConfig{
			Enabled:   cfg.Channels.MaixCam.Enabled,
			Host:      cfg.Channels.MaixCam.Host,
			Port:      cfg.Channels.MaixCam.Port,
			AllowFrom: cfg.Channels.MaixCam.AllowFrom,
		},
	}

	// Redact web config password
	redacted.Web = WebConfig{
		Enabled:   cfg.Web.Enabled,
		Host:      cfg.Web.Host,
		Port:      cfg.Web.Port,
		Username:  cfg.Web.Username,
		Password:  redactIfSet(cfg.Web.Password),
		JWTExpiry: cfg.Web.JWTExpiry,
	}

	// Redact tools config
	redacted.Tools = ToolsConfig{
		Web: WebToolsConfig{
			Search: WebSearchConfig{
				APIKey:     redactIfSet(cfg.Tools.Web.Search.APIKey),
				MaxResults: cfg.Tools.Web.Search.MaxResults,
			},
		},
		Email: EmailToolsConfig{
			Enabled:  cfg.Tools.Email.Enabled,
			Host:     cfg.Tools.Email.Host,
			Port:     cfg.Tools.Email.Port,
			Username: cfg.Tools.Email.Username,
			Password: redactIfSet(cfg.Tools.Email.Password),
			From:     cfg.Tools.Email.From,
			To:       cfg.Tools.Email.To,
		},
		MCP: cfg.Tools.MCP, // MCP config doesn't contain secrets in standard fields
	}

	return redacted
}

// redactProvider redacts a single provider config
func redactProvider(p ProviderConfig) ProviderConfig {
	return ProviderConfig{
		APIKey:     redactIfSet(p.APIKey),
		APIBase:    p.APIBase,
		Proxy:      p.Proxy,
		AuthMethod: p.AuthMethod,
		Models:     p.Models,
	}
}

// redactIfSet returns RedactValue if the string is non-empty, otherwise returns empty string
func redactIfSet(value string) string {
	if value == "" {
		return ""
	}
	return RedactValue(value)
}
