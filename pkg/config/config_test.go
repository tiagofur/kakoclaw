package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseProviderEnvVars(t *testing.T) {
	// Set test environment variables
	os.Setenv("MAKOCLAW_PROVIDERS_ANTHROPIC_API_KEY", "test-antropic-key")
	os.Setenv("MAKOCLAW_PROVIDERS_OPENAI_API_KEY", "test-openai-key")
	os.Setenv("MAKOCLAW_PROVIDERS_OPENROUTER_API_BASE", "https://custom.openrouter.ai")
	os.Setenv("MAKOCLAW_PROVIDERS_GROQ_PROXY", "http://proxy:8080")

	defer func() {
		os.Unsetenv("MAKOCLAW_PROVIDERS_ANTHROPIC_API_KEY")
		os.Unsetenv("MAKOCLAW_PROVIDERS_OPENAI_API_KEY")
		os.Unsetenv("MAKOCLAW_PROVIDERS_OPENROUTER_API_BASE")
		os.Unsetenv("MAKOCLAW_PROVIDERS_GROQ_PROXY")
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
	os.Setenv("MAKOCLAW_PROVIDERS_ANTHROPIC_API_KEY", "env-key")
	defer os.Unsetenv("MAKOCLAW_PROVIDERS_ANTHROPIC_API_KEY")

	// Parse env vars
	parseProviderEnvVars(cfg)

	// Environment should override config
	if cfg.Providers.Anthropic.APIKey != "env-key" {
		t.Errorf("Environment variable should override config, got: %s", cfg.Providers.Anthropic.APIKey)
	}
}
func TestInitDataDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "makoclaw-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	InitDataDir(tempDir)
	defer InitDataDir("") // Reset after test

	if getDataDir() != tempDir {
		t.Errorf("getDataDir() should return InitDataDir path, got: %s", getDataDir())
	}

	userUUID := "test-uuid"
	expectedPath := filepath.Join(tempDir, "users", userUUID, "config.json")
	if GetUserConfigPath(userUUID) != expectedPath {
		t.Errorf("GetUserConfigPath() incorrect, got: %s, want: %s", GetUserConfigPath(userUUID), expectedPath)
	}
}

func TestGetUserConfigPathDefault(t *testing.T) {
	InitDataDir("") // Reset to default

	home, _ := os.UserHomeDir()
	userUUID := "test-uuid"
	expectedPath := filepath.Join(home, ".makoclaw", "users", userUUID, "config.json")

	if GetUserConfigPath(userUUID) != expectedPath {
		t.Errorf("GetUserConfigPath() default incorrect, got: %s, want: %s", GetUserConfigPath(userUUID), expectedPath)
	}
}

func TestGetUserConfigTemplate(t *testing.T) {
	// Create a global config with some values
	globalCfg := &Config{
		Agents: AgentsConfig{
			Defaults: AgentDefaults{
				Model:               "claude-3-5-sonnet-20241022",
				MaxTokens:           16384,
				Temperature:         0.8,
				MaxToolIterations:   30,
				RestrictToWorkspace: false,
			},
			Orchestrator: OrchestratorConfig{
				MaxTokens:            24000,
				Temperature:          0.5,
				MaxDelegationRetries: 3,
				FallbackToDefault:    false,
			},
		},
		Providers: ProvidersConfig{
			Anthropic: ProviderConfig{
				APIKey:  "sk-ant-global-test-key",
				APIBase: "https://api.anthropic.com",
			},
			OpenAI: ProviderConfig{
				APIKey:  "sk-openai-global-test-key",
				APIBase: "https://api.openai.com/v1",
			},
			OpenRouter: ProviderConfig{
				APIKey:  "sk-or-global-test-key",
				APIBase: "https://openrouter.ai/api/v1",
			},
			Ollama: ProviderConfig{
				APIKey:  "",
				APIBase: "http://localhost:11434/v1",
			},
		},
		Channels: ChannelsConfig{
			Telegram: TelegramConfig{
				Enabled: true,
				Token:   "global-telegram-token",
			},
			Discord: DiscordConfig{
				Enabled: true,
				Token:   "global-discord-token",
			},
		},
		Gateway: GatewayConfig{
			Host: "0.0.0.0",
			Port: 18790,
		},
		Tools: ToolsConfig{
			Web: WebToolsConfig{
				Search: WebSearchConfig{
					APIKey:     "global-brave-api-key",
					MaxResults: 10,
				},
			},
			Email: EmailToolsConfig{
				Enabled:  true,
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "admin@example.com",
				Password: "secret-password",
				From:     "admin@example.com",
			},
		},
	}

	// Generate user config template
	userCfg := GetUserConfigTemplate(globalCfg)

	// Verify structure exists
	if userCfg == nil {
		t.Fatal("GetUserConfigTemplate() returned nil")
	}
	if userCfg.Providers.Anthropic.APIBase == "" {
		t.Error("Providers config should be initialized")
	}
	if userCfg.Channels.Telegram.Token == "" && userCfg.Channels.Telegram.Token != "" {
		// This is a bit of a confusing check - let's fix it below
	}

	// Verify NO sensitive data inherited
	if userCfg.Providers.Anthropic.APIKey != "" {
		t.Errorf("Should not inherit Anthropic API key, got: %s", userCfg.Providers.Anthropic.APIKey)
	}
	if userCfg.Providers.OpenAI.APIKey != "" {
		t.Errorf("Should not inherit OpenAI API key, got: %s", userCfg.Providers.OpenAI.APIKey)
	}
	if userCfg.Providers.OpenRouter.APIKey != "" {
		t.Errorf("Should not inherit OpenRouter API key, got: %s", userCfg.Providers.OpenRouter.APIKey)
	}
	if userCfg.Channels.Telegram.Token != "" {
		t.Errorf("Should not inherit Telegram token, got: %s", userCfg.Channels.Telegram.Token)
	}
	if userCfg.Channels.Discord.Token != "" {
		t.Errorf("Should not inherit Discord token, got: %s", userCfg.Channels.Discord.Token)
	}
	if userCfg.Tools.Web.Search.APIKey != "" {
		t.Errorf("Should not inherit web search API key, got: %s", userCfg.Tools.Web.Search.APIKey)
	}
	if userCfg.Tools.Email.Password != "" {
		t.Errorf("Should not inherit email password, got: %s", userCfg.Tools.Email.Password)
	}

	// Verify safe defaults ARE inherited
	if userCfg.Agents.Defaults.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Should inherit model, got: %s", userCfg.Agents.Defaults.Model)
	}
	if userCfg.Agents.Defaults.MaxTokens != 16384 {
		t.Errorf("Should inherit max_tokens, got: %d", userCfg.Agents.Defaults.MaxTokens)
	}
	if userCfg.Agents.Defaults.Temperature != 0.8 {
		t.Errorf("Should inherit temperature, got: %f", userCfg.Agents.Defaults.Temperature)
	}
	if userCfg.Agents.Defaults.MaxToolIterations != 30 {
		t.Errorf("Should inherit max_tool_iterations, got: %d", userCfg.Agents.Defaults.MaxToolIterations)
	}
	if userCfg.Agents.Defaults.RestrictToWorkspace != false {
		t.Errorf("Should inherit restrict_to_workspace, got: %v", userCfg.Agents.Defaults.RestrictToWorkspace)
	}
	if userCfg.Agents.Orchestrator.MaxTokens != 24000 {
		t.Errorf("Should inherit orchestrator max_tokens, got: %d", userCfg.Agents.Orchestrator.MaxTokens)
	}
	if userCfg.Agents.Orchestrator.Temperature != 0.5 {
		t.Errorf("Should inherit orchestrator temperature, got: %f", userCfg.Agents.Orchestrator.Temperature)
	}
	if userCfg.Tools.Web.Search.MaxResults != 10 {
		t.Errorf("Should inherit max_results, got: %d", userCfg.Tools.Web.Search.MaxResults)
	}

	// Verify base URLs ARE inherited
	if userCfg.Providers.Anthropic.APIBase != "https://api.anthropic.com" {
		t.Errorf("Should inherit Anthropic base URL, got: %s", userCfg.Providers.Anthropic.APIBase)
	}
	if userCfg.Providers.OpenAI.APIBase != "https://api.openai.com/v1" {
		t.Errorf("Should inherit OpenAI base URL, got: %s", userCfg.Providers.OpenAI.APIBase)
	}
	if userCfg.Providers.OpenRouter.APIBase != "https://openrouter.ai/api/v1" {
		t.Errorf("Should inherit OpenRouter base URL, got: %s", userCfg.Providers.OpenRouter.APIBase)
	}
	if userCfg.Providers.Ollama.APIBase != "http://localhost:11434/v1" {
		t.Errorf("Should inherit Ollama base URL, got: %s", userCfg.Providers.Ollama.APIBase)
	}

	// Verify all channels are disabled
	if userCfg.Channels.Telegram.Enabled {
		t.Error("Telegram should be disabled")
	}
	if userCfg.Channels.Discord.Enabled {
		t.Error("Discord should be disabled")
	}
	if userCfg.Channels.Slack.Enabled {
		t.Error("Slack should be disabled")
	}
	if userCfg.Channels.WhatsApp.Enabled {
		t.Error("WhatsApp should be disabled")
	}
	if userCfg.Channels.Signal.Enabled {
		t.Error("Signal should be disabled")
	}
	if userCfg.Channels.Feishu.Enabled {
		t.Error("Feishu should be disabled")
	}
	if userCfg.Channels.QQ.Enabled {
		t.Error("QQ should be disabled")
	}
	if userCfg.Channels.DingTalk.Enabled {
		t.Error("DingTalk should be disabled")
	}
	if userCfg.Channels.MaixCam.Enabled {
		t.Error("MaixCam should be disabled")
	}

	// Verify provider field is empty (user must choose)
	if userCfg.Agents.Defaults.Provider != "" {
		t.Errorf("Provider should be empty, got: %s", userCfg.Agents.Defaults.Provider)
	}

	// Verify orchestrator is disabled
	if userCfg.Agents.Orchestrator.Enabled {
		t.Error("Orchestrator should be disabled")
	}

	// Verify orchestrator provider/model are empty
	if userCfg.Agents.Orchestrator.Provider != "" {
		t.Errorf("Orchestrator provider should be empty, got: %s", userCfg.Agents.Orchestrator.Provider)
	}
	if userCfg.Agents.Orchestrator.Model != "" {
		t.Errorf("Orchestrator model should be empty, got: %s", userCfg.Agents.Orchestrator.Model)
	}

	// Verify web interface is disabled
	if userCfg.Web.Enabled {
		t.Error("Web interface should be disabled")
	}

	// Verify email is disabled and has no sensitive config
	if userCfg.Tools.Email.Enabled {
		t.Error("Email should be disabled")
	}
	if userCfg.Tools.Email.Host != "" {
		t.Errorf("Email host should be empty, got: %s", userCfg.Tools.Email.Host)
	}
	if userCfg.Tools.Email.Username != "" {
		t.Errorf("Email username should be empty, got: %s", userCfg.Tools.Email.Username)
	}

	// Verify gateway config was inherited
	if userCfg.Gateway.Host != "0.0.0.0" {
		t.Errorf("Should inherit gateway host, got: %s", userCfg.Gateway.Host)
	}
	if userCfg.Gateway.Port != 18790 {
		t.Errorf("Should inherit gateway port, got: %d", userCfg.Gateway.Port)
	}
}

func TestGetUserConfigTemplateWithNilGlobal(t *testing.T) {
	// Test that nil global config doesn't panic and uses defaults
	userCfg := GetUserConfigTemplate(nil)

	if userCfg == nil {
		t.Fatal("GetUserConfigTemplate(nil) returned nil")
	}

	// Should have default values
	if userCfg.Agents.Defaults.Model == "" {
		t.Error("Model should have a default value")
	}
	if userCfg.Agents.Defaults.MaxTokens == 0 {
		t.Error("MaxTokens should have a default value")
	}
}

func TestIsAgentsConfigEmpty(t *testing.T) {
	// Empty config
	emptyCfg := &AgentsConfig{}
	if !isAgentsConfigEmpty(emptyCfg) {
		t.Error("Empty agents config should be considered empty")
	}

	// Config with orchestrator enabled
	withOrchestrator := &AgentsConfig{
		Orchestrator: OrchestratorConfig{Enabled: true},
	}
	if isAgentsConfigEmpty(withOrchestrator) {
		t.Error("Config with enabled orchestrator should NOT be empty")
	}

	// Config with specialists
	withSpecialists := &AgentsConfig{
		Specialists: map[string]SpecialistConfig{
			"test": {Name: "Test Specialist"},
		},
	}
	if isAgentsConfigEmpty(withSpecialists) {
		t.Error("Config with specialists should NOT be empty")
	}

	// Config with removed specialists
	withRemoved := &AgentsConfig{
		RemovedSpecialists: []string{"default-agent"},
	}
	if isAgentsConfigEmpty(withRemoved) {
		t.Error("Config with removed specialists should NOT be empty")
	}

	// Config with custom defaults
	withDefaults := &AgentsConfig{
		Defaults: AgentDefaults{
			Model: "custom-model",
		},
	}
	if isAgentsConfigEmpty(withDefaults) {
		t.Error("Config with custom defaults should NOT be empty")
	}

	// Config with custom workspace
	withWorkspace := &AgentsConfig{
		Defaults: AgentDefaults{
			Workspace: "/custom/workspace",
		},
	}
	if isAgentsConfigEmpty(withWorkspace) {
		t.Error("Config with custom workspace should NOT be empty")
	}

	// Config with custom provider
	withProvider := &AgentsConfig{
		Defaults: AgentDefaults{
			Provider: "anthropic",
		},
	}
	if isAgentsConfigEmpty(withProvider) {
		t.Error("Config with custom provider should NOT be empty")
	}
}
