package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ValidateURL ensures a string is a valid URL (or empty).
// Accepts http://, https://, ws://, and wss:// schemes.
func ValidateURL(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}
	if s == "" {
		return nil // Empty is valid (means "use default")
	}

	parsed, err := url.ParseRequestURI(s)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check for valid schemes
	validSchemes := map[string]bool{
		"http":  true,
		"https": true,
		"ws":    true,
		"wss":   true,
	}
	if !validSchemes[parsed.Scheme] {
		return fmt.Errorf("invalid URL scheme: %s (must be http, https, ws, or wss)", parsed.Scheme)
	}

	// Ensure URL has a host
	if parsed.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	return nil
}

// ValidatePort ensures a value is a valid port number (1-65535).
func ValidatePort(value interface{}) error {
	var port int
	switch v := value.(type) {
	case float64:
		port = int(v)
	case int:
		port = v
	case int64:
		port = int(v)
	default:
		return fmt.Errorf("expected number, got %T", value)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	return nil
}

// ValidatePositiveInt ensures a value is a non-negative integer.
func ValidatePositiveInt(value interface{}) error {
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

	if n < 0 {
		return fmt.Errorf("value must be non-negative, got %d", n)
	}
	return nil
}

// ValidateTemperature ensures temperature is in valid range [0.0, 2.0].
func ValidateTemperature(value interface{}) error {
	var temp float64
	switch v := value.(type) {
	case float64:
		temp = v
	case int:
		temp = float64(v)
	case int64:
		temp = float64(v)
	default:
		return fmt.Errorf("expected number, got %T", value)
	}

	if temp < 0.0 || temp > 2.0 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0, got %.2f", temp)
	}
	return nil
}

// ValidateMaxTokens ensures max_tokens is in valid range [1, 1000000].
func ValidateMaxTokens(value interface{}) error {
	var tokens int
	switch v := value.(type) {
	case float64:
		tokens = int(v)
	case int:
		tokens = v
	case int64:
		tokens = int(v)
	default:
		return fmt.Errorf("expected number, got %T", value)
	}

	if tokens < 1 || tokens > 1000000 {
		return fmt.Errorf("max_tokens must be between 1 and 1000000, got %d", tokens)
	}
	return nil
}

// ValidateMaxIterations ensures max_tool_iterations is in valid range [1, 100].
func ValidateMaxIterations(value interface{}) error {
	var iterations int
	switch v := value.(type) {
	case float64:
		iterations = int(v)
	case int:
		iterations = v
	case int64:
		iterations = int(v)
	default:
		return fmt.Errorf("expected number, got %T", value)
	}

	if iterations < 1 || iterations > 100 {
		return fmt.Errorf("max_tool_iterations must be between 1 and 100, got %d", iterations)
	}
	return nil
}

// ValidateEmail provides basic email format validation.
// Requires the presence of @ and at least one . after the @.
func ValidateEmail(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}
	if s == "" {
		return nil // Empty is valid
	}

	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(s) {
		return fmt.Errorf("invalid email format: %s", s)
	}
	return nil
}

// ValidProviderNames is the list of known LLM provider names.
var ValidProviderNames = []string{
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

// ValidateProviderName ensures the provider name is known.
// Empty string is valid (means "not set").
func ValidateProviderName(name string) error {
	if name == "" {
		return nil // Empty is valid (means "use default" or "not set")
	}

	nameLower := strings.ToLower(name)
	for _, valid := range ValidProviderNames {
		if nameLower == valid {
			return nil
		}
	}
	return fmt.Errorf("unknown provider: %s (valid: %s)", name, strings.Join(ValidProviderNames, ", "))
}

// ValidChannelNames is the list of known communication channel names.
var ValidChannelNames = []string{
	"telegram",
	"discord",
	"slack",
	"whatsapp",
	"signal",
	"qq",
	"dingtalk",
	"feishu",
	"maixcam",
	"email",
}

// ValidateChannelName ensures the channel name is known.
func ValidateChannelName(name string) error {
	if name == "" {
		return fmt.Errorf("channel name cannot be empty")
	}

	nameLower := strings.ToLower(name)
	for _, valid := range ValidChannelNames {
		if nameLower == valid {
			return nil
		}
	}
	return fmt.Errorf("unknown channel: %s (valid: %s)", name, strings.Join(ValidChannelNames, ", "))
}

// ValidateSlackBotToken validates Slack bot token format.
// Valid tokens start with "xoxb-" or are empty.
func ValidateSlackBotToken(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}
	if s == "" {
		return nil // Empty is valid
	}
	if !strings.HasPrefix(s, "xoxb-") {
		return fmt.Errorf("Slack bot token must start with 'xoxb-'")
	}
	return nil
}

// ValidateSlackAppToken validates Slack app token format.
// Valid tokens start with "xapp-" or are empty.
func ValidateSlackAppToken(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}
	if s == "" {
		return nil // Empty is valid
	}
	if !strings.HasPrefix(s, "xapp-") {
		return fmt.Errorf("Slack app token must start with 'xapp-'")
	}
	return nil
}

// ValidateConfigPath validates a dot-notation config path format.
// Valid paths contain only lowercase letters, numbers, underscores, and dots.
// Path must start with a letter and each segment must start with a letter.
func ValidateConfigPath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed")
	}
	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return fmt.Errorf("absolute paths not allowed")
	}

	// Regex for valid path: lowercase letters, numbers, underscores, and dots
	// Each segment must start with a letter
	validPathPattern := regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)
	if !validPathPattern.MatchString(path) {
		return fmt.Errorf("invalid path format: must be dot-notation like 'providers.openai.api_key'")
	}

	return nil
}

// ValidateAuthMethod validates provider authentication method.
// Valid values: "", "bearer", "basic"
func ValidateAuthMethod(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}

	validMethods := map[string]bool{
		"":       true,
		"bearer": true,
		"basic":  true,
	}
	if !validMethods[s] {
		return fmt.Errorf("invalid auth_method: %s (valid: bearer, basic, or empty)", s)
	}
	return nil
}
