package config

import (
	"os"
	"testing"
)

func TestParseProviderEnvVars(t *testing.T) {
	// Set test environment variables
	os.Setenv("PICOCLAW_PROVIDERS_ANTHROPIC_API_KEY", "test-antropic-key")
	os.Setenv("PICOCLAW_PROVIDERS_OPENAI_API_KEY", "test-openai-key")
	os.Setenv("PICOCLAW_PROVIDERS_OPENROUTER_API_BASE", "https://custom.openrouter.ai")
	os.Setenv("PICOCLAW_PROVIDERS_GROQ_PROXY", "http://proxy:8080")
	
	defer func() {
		os.Unsetenv("PICOCLAW_PROVIDERS_ANTHROPIC_API_KEY")
		os.Unsetenv("PICOCLAW_PROVIDERS_OPENAI_API_KEY")
		os.Unsetenv("PICOCLAW_PROVIDERS_OPENROUTER_API_BASE")
		os.Unsetenv("PICOCLAW_PROVIDERS_GROQ_PROXY")
	}()
	
	cfg := DefaultConfig()
	parseProviderEnvVars(cfg)
	
	// Verify Anthropic API key was set
	if cfg.Providers.Anthropic.APIKey != "test-antropic-key" {
		t.Errorf("Anthropic API Key not set correctly, got: %s", cfg.Providers.Anthropic.APIKey)
	}
	
	// Verify OpenAI API key was set
	if cfg.Providers.OpenAI.APIKey != "test-openai-key" {
		t.Errorf("OpenAI API Key not set correctly, got: %s", cfg.Providers.OpenAI.APIKey)
	}
	
	// Verify OpenRouter API base was set
	if cfg.Providers.OpenRouter.APIBase != "https://custom.openrouter.ai" {
		t.Errorf("OpenRouter API Base not set correctly, got: %s", cfg.Providers.OpenRouter.APIBase)
	}
	
	// Verify Groq proxy was set
	if cfg.Providers.Groq.Proxy != "http://proxy:8080" {
		t.Errorf("Groq Proxy not set correctly, got: %s", cfg.Providers.Groq.Proxy)
	}
}

func TestProviderEnvVarsOverrideConfig(t *testing.T) {
	// Create a config with existing values
	cfg := DefaultConfig()
	cfg.Providers.Anthropic.APIKey = "config-key"
	
	// Set environment variable
	os.Setenv("PICOCLAW_PROVIDERS_ANTHROPIC_API_KEY", "env-key")
	defer os.Unsetenv("PICOCLAW_PROVIDERS_ANTHROPIC_API_KEY")
	
	// Parse env vars
	parseProviderEnvVars(cfg)
	
	// Environment should override config
	if cfg.Providers.Anthropic.APIKey != "env-key" {
		t.Errorf("Environment variable should override config, got: %s", cfg.Providers.Anthropic.APIKey)
	}
}
