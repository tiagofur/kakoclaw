package config

import (
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		// Valid URLs
		{"empty string", "", false},
		{"http url", "http://example.com", false},
		{"https url", "https://api.openai.com/v1", false},
		{"https with port", "https://api.example.com:8080/v1", false},
		{"ws url", "ws://localhost:3001", false},
		{"wss url", "wss://secure.example.com/socket", false},
		{"http localhost", "http://localhost:8080", false},
		{"http with path", "http://api.example.com/v1/chat", false},

		// Invalid URLs
		{"no scheme", "example.com", true},
		{"ftp scheme", "ftp://files.example.com", true},
		{"file scheme", "file:///etc/passwd", true},
		{"just text", "not-a-url", true},
		{"missing host", "http://", true},

		// Type errors
		{"integer", 123, true},
		{"boolean", true, true},
		{"nil", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateURL(%v) error = %v, wantError %v", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		// Valid ports
		{"port 1", 1, false},
		{"port 80", 80, false},
		{"port 443", 443, false},
		{"port 8080", 8080, false},
		{"port 65535", 65535, false},
		{"port as float64", float64(8080), false},
		{"port as int64", int64(8080), false},

		// Invalid ports
		{"port 0", 0, true},
		{"port -1", -1, true},
		{"port 65536", 65536, true},
		{"port 70000", 70000, true},
		{"port negative float", float64(-100), true},

		// Type errors
		{"string", "8080", true},
		{"boolean", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePort(%v) error = %v, wantError %v", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestValidatePositiveInt(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		// Valid values
		{"zero", 0, false},
		{"positive int", 100, false},
		{"large int", 1000000, false},
		{"float64 positive", float64(50), false},
		{"int64 positive", int64(50), false},

		// Invalid values
		{"negative int", -1, true},
		{"negative float", float64(-10), true},

		// Type errors
		{"string", "100", true},
		{"boolean", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveInt(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePositiveInt(%v) error = %v, wantError %v", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestValidateTemperature(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		// Valid temperatures
		{"zero", float64(0), false},
		{"half", float64(0.5), false},
		{"one", float64(1.0), false},
		{"one point five", float64(1.5), false},
		{"two", float64(2.0), false},
		{"int zero", 0, false},
		{"int one", 1, false},
		{"int two", 2, false},

		// Invalid temperatures
		{"negative", float64(-0.1), true},
		{"too high", float64(2.1), true},
		{"way too high", float64(5.0), true},
		{"negative int", -1, true},

		// Type errors
		{"string", "0.5", true},
		{"boolean", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemperature(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateTemperature(%v) error = %v, wantError %v", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestValidateMaxTokens(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		// Valid values
		{"min valid", 1, false},
		{"typical", 8192, false},
		{"large", 100000, false},
		{"max valid", 1000000, false},
		{"float64", float64(4096), false},

		// Invalid values
		{"zero", 0, true},
		{"negative", -1, true},
		{"too large", 1000001, true},

		// Type errors
		{"string", "8192", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxTokens(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMaxTokens(%v) error = %v, wantError %v", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestValidateMaxIterations(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		// Valid values
		{"min valid", 1, false},
		{"typical", 20, false},
		{"max valid", 100, false},
		{"float64", float64(50), false},

		// Invalid values
		{"zero", 0, true},
		{"negative", -1, true},
		{"too large", 101, true},

		// Type errors
		{"string", "20", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxIterations(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateMaxIterations(%v) error = %v, wantError %v", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		// Valid emails
		{"empty string", "", false},
		{"simple email", "user@example.com", false},
		{"email with dots", "first.last@example.com", false},
		{"email with plus", "user+tag@example.com", false},
		{"email with subdomain", "user@mail.example.com", false},
		{"email with numbers", "user123@example456.com", false},

		// Invalid emails
		{"no at sign", "userexample.com", true},
		{"no domain", "user@", true},
		{"no tld", "user@example", true},
		{"double at", "user@@example.com", true},
		{"spaces", "user @example.com", true},
		{"just text", "not-an-email", true},

		// Type errors
		{"integer", 123, true},
		{"boolean", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateEmail(%v) error = %v, wantError %v", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestValidateProviderName(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		wantError bool
	}{
		// Valid providers
		{"empty (unset)", "", false},
		{"anthropic", "anthropic", false},
		{"openai", "openai", false},
		{"openrouter", "openrouter", false},
		{"groq", "groq", false},
		{"zhipu", "zhipu", false},
		{"vllm", "vllm", false},
		{"gemini", "gemini", false},
		{"nvidia", "nvidia", false},
		{"moonshot", "moonshot", false},
		{"ollama", "ollama", false},
		{"uppercase", "OPENAI", false},
		{"mixed case", "OpenAI", false},

		// Invalid providers
		{"unknown", "unknown", true},
		{"gpt", "gpt", true},
		{"claude", "claude", true},
		{"bedrock", "bedrock", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProviderName(tt.provider)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateProviderName(%q) error = %v, wantError %v", tt.provider, err, tt.wantError)
			}
		})
	}
}

func TestValidateChannelName(t *testing.T) {
	tests := []struct {
		name      string
		channel   string
		wantError bool
	}{
		// Valid channels
		{"telegram", "telegram", false},
		{"discord", "discord", false},
		{"slack", "slack", false},
		{"whatsapp", "whatsapp", false},
		{"signal", "signal", false},
		{"qq", "qq", false},
		{"dingtalk", "dingtalk", false},
		{"feishu", "feishu", false},
		{"maixcam", "maixcam", false},
		{"uppercase", "TELEGRAM", false},
		{"mixed case", "Telegram", false},

		// Invalid channels
		{"empty", "", true},
		{"unknown", "unknown", true},
		{"email", "email", false},
		{"sms", "sms", true},
		{"teams", "teams", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChannelName(tt.channel)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateChannelName(%q) error = %v, wantError %v", tt.channel, err, tt.wantError)
			}
		})
	}
}

func TestValidateSlackBotToken(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		// Valid tokens
		{"empty", "", false},
		{"valid bot token", "xoxb-TESTTOKEN-EXAMPLE", false},
		{"valid long token", "xoxb-TESTTOKEN123-EXAMPLEABCDEF", false},

		// Invalid tokens
		{"wrong prefix", "xoxa-TESTTOKEN", true},
		{"no prefix", "123456789-abcdef", true},
		{"app token format", "xapp-123456789", true},

		// Type errors
		{"integer", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSlackBotToken(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateSlackBotToken(%v) error = %v, wantError %v", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestValidateSlackAppToken(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		// Valid tokens
		{"empty", "", false},
		{"valid app token", "xapp-TESTTOKEN-EXAMPLE", false},
		{"valid long token", "xapp-TESTTOKEN123-EXAMPLEABCDEF", false},

		// Invalid tokens
		{"wrong prefix", "xoxa-TESTTOKEN", true},
		{"no prefix", "TESTTOKEN-EXAMPLE", true},
		{"bot token format", "xoxb-TESTTOKEN", true},

		// Type errors
		{"integer", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSlackAppToken(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateSlackAppToken(%v) error = %v, wantError %v", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestValidateConfigPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		// Valid paths
		{"simple", "providers", false},
		{"two segments", "providers.openai", false},
		{"three segments", "providers.openai.api_key", false},
		{"with underscore", "agents.defaults.max_tokens", false},
		{"with numbers", "channels.telegram2.enabled", false},

		// Invalid paths
		{"empty", "", true},
		{"path traversal", "providers/../secrets", true},
		{"absolute path", "/providers/openai", true},
		{"windows absolute", "\\providers\\openai", true},
		{"starts with number", "1providers.openai", true},
		{"segment starts with number", "providers.1openai", true},
		{"uppercase", "Providers.OpenAI", true},
		{"special chars", "providers.open-ai.api_key", true},
		{"spaces", "providers. openai", true},
		{"double dots", "providers..openai", true},
		{"trailing dot", "providers.openai.", true},
		{"leading dot", ".providers.openai", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfigPath(tt.path)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateConfigPath(%q) error = %v, wantError %v", tt.path, err, tt.wantError)
			}
		})
	}
}

func TestValidateAuthMethod(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		// Valid methods
		{"empty", "", false},
		{"bearer", "bearer", false},
		{"basic", "basic", false},

		// Invalid methods
		{"unknown", "unknown", true},
		{"digest", "digest", true},
		{"oauth", "oauth", true},

		// Type errors
		{"integer", 123, true},
		{"boolean", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuthMethod(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateAuthMethod(%v) error = %v, wantError %v", tt.value, err, tt.wantError)
			}
		})
	}
}
