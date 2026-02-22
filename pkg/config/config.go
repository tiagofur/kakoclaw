package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
)

// FlexibleStringSlice is a []string that also accepts JSON numbers,
// so allow_from can contain both "123" and 123.
type FlexibleStringSlice []string

func (f *FlexibleStringSlice) UnmarshalJSON(data []byte) error {
	// Try []string first
	var ss []string
	if err := json.Unmarshal(data, &ss); err == nil {
		*f = ss
		return nil
	}

	// Try []interface{} to handle mixed types
	var raw []interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	result := make([]string, 0, len(raw))
	for _, v := range raw {
		switch val := v.(type) {
		case string:
			result = append(result, val)
		case float64:
			result = append(result, fmt.Sprintf("%.0f", val))
		default:
			result = append(result, fmt.Sprintf("%v", val))
		}
	}
	*f = result
	return nil
}

type Config struct {
	Agents    AgentsConfig    `json:"agents"`
	Channels  ChannelsConfig  `json:"channels"`
	Providers ProvidersConfig `json:"providers"`
	Gateway   GatewayConfig   `json:"gateway"`
	Web       WebConfig       `json:"web"`
	Tools     ToolsConfig     `json:"tools"`
	Storage   StorageConfig   `json:"storage"`
	mu        sync.RWMutex
}

type StorageConfig struct {
	Path string `json:"path" env:"KAKOCLAW_STORAGE_PATH"`
}

type AgentsConfig struct {
	Defaults     AgentDefaults               `json:"defaults"`
	Orchestrator OrchestratorConfig          `json:"orchestrator"`
	Specialists  map[string]SpecialistConfig `json:"specialists"`
}

type AgentDefaults struct {
	Workspace           string  `json:"workspace" env:"KAKOCLAW_AGENTS_DEFAULTS_WORKSPACE"`
	RestrictToWorkspace bool    `json:"restrict_to_workspace" env:"KAKOCLAW_AGENTS_DEFAULTS_RESTRICT_TO_WORKSPACE"`
	Provider            string  `json:"provider" env:"KAKOCLAW_AGENTS_DEFAULTS_PROVIDER"`
	Model               string  `json:"model" env:"KAKOCLAW_AGENTS_DEFAULTS_MODEL"`
	MaxTokens           int     `json:"max_tokens" env:"KAKOCLAW_AGENTS_DEFAULTS_MAX_TOKENS"`
	Temperature         float64 `json:"temperature" env:"KAKOCLAW_AGENTS_DEFAULTS_TEMPERATURE"`
	MaxToolIterations   int     `json:"max_tool_iterations" env:"KAKOCLAW_AGENTS_DEFAULTS_MAX_TOOL_ITERATIONS"`
}

type OrchestratorConfig struct {
	Enabled              bool    `json:"enabled"`
	Provider             string  `json:"provider"`
	Model                string  `json:"model"`
	MaxTokens            int     `json:"max_tokens"`
	Temperature          float64 `json:"temperature"`
	MaxDelegationRetries int     `json:"max_delegation_retries"`
	FallbackToDefault    bool    `json:"fallback_to_default"`
	Description          string  `json:"description"`
}

type SpecialistConfig struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Prompt            string   `json:"prompt"`
	Provider          string   `json:"provider"`
	Model             string   `json:"model"`
	MaxTokens         int      `json:"max_tokens"`
	Temperature       float64  `json:"temperature"`
	MaxToolIterations int      `json:"max_tool_iterations"`
	Tools             []string `json:"tools"`
	Keywords          []string `json:"keywords"`
}

type ChannelsConfig struct {
	WhatsApp WhatsAppConfig `json:"whatsapp"`
	Telegram TelegramConfig `json:"telegram"`
	Feishu   FeishuConfig   `json:"feishu"`
	Discord  DiscordConfig  `json:"discord"`
	MaixCam  MaixCamConfig  `json:"maixcam"`
	QQ       QQConfig       `json:"qq"`
	DingTalk DingTalkConfig `json:"dingtalk"`
	Slack    SlackConfig    `json:"slack"`
	Signal   SignalConfig   `json:"signal"`
}

type WhatsAppConfig struct {
	Enabled   bool                `json:"enabled" env:"KAKOCLAW_CHANNELS_WHATSAPP_ENABLED"`
	BridgeURL string              `json:"bridge_url" env:"KAKOCLAW_CHANNELS_WHATSAPP_BRIDGE_URL"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"KAKOCLAW_CHANNELS_WHATSAPP_ALLOW_FROM"`
}

type TelegramConfig struct {
	Enabled   bool                `json:"enabled" env:"KAKOCLAW_CHANNELS_TELEGRAM_ENABLED"`
	Token     string              `json:"token" env:"KAKOCLAW_CHANNELS_TELEGRAM_TOKEN"`
	Proxy     string              `json:"proxy" env:"KAKOCLAW_CHANNELS_TELEGRAM_PROXY"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"KAKOCLAW_CHANNELS_TELEGRAM_ALLOW_FROM"`
}

type FeishuConfig struct {
	Enabled           bool                `json:"enabled" env:"KAKOCLAW_CHANNELS_FEISHU_ENABLED"`
	AppID             string              `json:"app_id" env:"KAKOCLAW_CHANNELS_FEISHU_APP_ID"`
	AppSecret         string              `json:"app_secret" env:"KAKOCLAW_CHANNELS_FEISHU_APP_SECRET"`
	EncryptKey        string              `json:"encrypt_key" env:"KAKOCLAW_CHANNELS_FEISHU_ENCRYPT_KEY"`
	VerificationToken string              `json:"verification_token" env:"KAKOCLAW_CHANNELS_FEISHU_VERIFICATION_TOKEN"`
	AllowFrom         FlexibleStringSlice `json:"allow_from" env:"KAKOCLAW_CHANNELS_FEISHU_ALLOW_FROM"`
}

type DiscordConfig struct {
	Enabled   bool                `json:"enabled" env:"KAKOCLAW_CHANNELS_DISCORD_ENABLED"`
	Token     string              `json:"token" env:"KAKOCLAW_CHANNELS_DISCORD_TOKEN"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"KAKOCLAW_CHANNELS_DISCORD_ALLOW_FROM"`
}

type MaixCamConfig struct {
	Enabled   bool                `json:"enabled" env:"KAKOCLAW_CHANNELS_MAIXCAM_ENABLED"`
	Host      string              `json:"host" env:"KAKOCLAW_CHANNELS_MAIXCAM_HOST"`
	Port      int                 `json:"port" env:"KAKOCLAW_CHANNELS_MAIXCAM_PORT"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"KAKOCLAW_CHANNELS_MAIXCAM_ALLOW_FROM"`
}

type QQConfig struct {
	Enabled   bool                `json:"enabled" env:"KAKOCLAW_CHANNELS_QQ_ENABLED"`
	AppID     string              `json:"app_id" env:"KAKOCLAW_CHANNELS_QQ_APP_ID"`
	AppSecret string              `json:"app_secret" env:"KAKOCLAW_CHANNELS_QQ_APP_SECRET"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"KAKOCLAW_CHANNELS_QQ_ALLOW_FROM"`
}

type DingTalkConfig struct {
	Enabled      bool                `json:"enabled" env:"KAKOCLAW_CHANNELS_DINGTALK_ENABLED"`
	ClientID     string              `json:"client_id" env:"KAKOCLAW_CHANNELS_DINGTALK_CLIENT_ID"`
	ClientSecret string              `json:"client_secret" env:"KAKOCLAW_CHANNELS_DINGTALK_CLIENT_SECRET"`
	AllowFrom    FlexibleStringSlice `json:"allow_from" env:"KAKOCLAW_CHANNELS_DINGTALK_ALLOW_FROM"`
}

type SlackConfig struct {
	Enabled   bool     `json:"enabled" env:"KAKOCLAW_CHANNELS_SLACK_ENABLED"`
	BotToken  string   `json:"bot_token" env:"KAKOCLAW_CHANNELS_SLACK_BOT_TOKEN"`
	AppToken  string   `json:"app_token" env:"KAKOCLAW_CHANNELS_SLACK_APP_TOKEN"`
	AllowFrom []string `json:"allow_from" env:"KAKOCLAW_CHANNELS_SLACK_ALLOW_FROM"`
}

type SignalConfig struct {
	Enabled     bool                `json:"enabled" env:"KAKOCLAW_CHANNELS_SIGNAL_ENABLED"`
	PhoneNumber string              `json:"phone_number" env:"KAKOCLAW_CHANNELS_SIGNAL_PHONE_NUMBER"`
	AllowFrom   FlexibleStringSlice `json:"allow_from" env:"KAKOCLAW_CHANNELS_SIGNAL_ALLOW_FROM"`
}

type ProvidersConfig struct {
	Anthropic  ProviderConfig `json:"anthropic"`
	OpenAI     ProviderConfig `json:"openai"`
	OpenRouter ProviderConfig `json:"openrouter"`
	Groq       ProviderConfig `json:"groq"`
	Zhipu      ProviderConfig `json:"zhipu"`
	VLLM       ProviderConfig `json:"vllm"`
	Gemini     ProviderConfig `json:"gemini"`
	Nvidia     ProviderConfig `json:"nvidia"`
	Moonshot   ProviderConfig `json:"moonshot"`
	Ollama     ProviderConfig `json:"ollama"`
}

type ProviderConfig struct {
	APIKey     string   `json:"api_key"`
	APIBase    string   `json:"api_base"`
	Proxy      string   `json:"proxy,omitempty"`
	AuthMethod string   `json:"auth_method,omitempty"`
	Models     []string `json:"models,omitempty"`
}

type GatewayConfig struct {
	Host string `json:"host" env:"KAKOCLAW_GATEWAY_HOST"`
	Port int    `json:"port" env:"KAKOCLAW_GATEWAY_PORT"`
}

type WebConfig struct {
	Enabled   bool   `json:"enabled" env:"KAKOCLAW_WEB_ENABLED"`
	Host      string `json:"host" env:"KAKOCLAW_WEB_HOST"`
	Port      int    `json:"port" env:"KAKOCLAW_WEB_PORT"`
	Username  string `json:"username" env:"KAKOCLAW_WEB_USERNAME"`
	Password  string `json:"password" env:"KAKOCLAW_WEB_PASSWORD"`
	JWTExpiry string `json:"jwt_expiry" env:"KAKOCLAW_WEB_JWT_EXPIRY"`
}

type WebSearchConfig struct {
	APIKey     string `json:"api_key" env:"KAKOCLAW_TOOLS_WEB_SEARCH_API_KEY"`
	MaxResults int    `json:"max_results" env:"KAKOCLAW_TOOLS_WEB_SEARCH_MAX_RESULTS"`
}

type WebToolsConfig struct {
	Search WebSearchConfig `json:"search"`
}

type ToolsConfig struct {
	Web   WebToolsConfig   `json:"web"`
	Email EmailToolsConfig `json:"email"`
	MCP   MCPConfig        `json:"mcp"`
}

type MCPConfig struct {
	Servers map[string]MCPServerConfig `json:"servers"`
}

type MCPServerConfig struct {
	Enabled bool              `json:"enabled"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

type EmailToolsConfig struct {
	Enabled  bool   `json:"enabled" env:"KAKOCLAW_TOOLS_EMAIL_ENABLED"`
	Host     string `json:"host" env:"KAKOCLAW_TOOLS_EMAIL_HOST"`
	Port     int    `json:"port" env:"KAKOCLAW_TOOLS_EMAIL_PORT"`
	Username string `json:"username" env:"KAKOCLAW_TOOLS_EMAIL_USERNAME"`
	Password string `json:"password" env:"KAKOCLAW_TOOLS_EMAIL_PASSWORD"`
	From     string `json:"from" env:"KAKOCLAW_TOOLS_EMAIL_FROM"`
	To       string `json:"to" env:"KAKOCLAW_TOOLS_EMAIL_TO"`
}

func DefaultConfig() *Config {
	return &Config{
		Agents: AgentsConfig{
			Defaults: AgentDefaults{
				Workspace:           "~/.kakoclaw/workspace",
				RestrictToWorkspace: true,
				Provider:            "",
				Model:               "openrouter",
				MaxTokens:           8192,
				Temperature:         0.7,
				MaxToolIterations:   20,
			},
			Orchestrator: OrchestratorConfig{
				Enabled:              false,
				Provider:             "",
				Model:                "",
				MaxTokens:            12000,
				Temperature:          0.7,
				MaxDelegationRetries: 2,
				FallbackToDefault:    true,
				Description:          "Project Manager: Analyzes tasks and delegates to appropriate specialists",
			},
			Specialists: map[string]SpecialistConfig{},
		},
		Channels: ChannelsConfig{
			WhatsApp: WhatsAppConfig{
				Enabled:   false,
				BridgeURL: "ws://localhost:3001",
				AllowFrom: FlexibleStringSlice{},
			},
			Telegram: TelegramConfig{
				Enabled:   false,
				Token:     "",
				AllowFrom: FlexibleStringSlice{},
			},
			Feishu: FeishuConfig{
				Enabled:           false,
				AppID:             "",
				AppSecret:         "",
				EncryptKey:        "",
				VerificationToken: "",
				AllowFrom:         FlexibleStringSlice{},
			},
			Discord: DiscordConfig{
				Enabled:   false,
				Token:     "",
				AllowFrom: FlexibleStringSlice{},
			},
			MaixCam: MaixCamConfig{
				Enabled:   false,
				Host:      "0.0.0.0",
				Port:      18790,
				AllowFrom: FlexibleStringSlice{},
			},
			QQ: QQConfig{
				Enabled:   false,
				AppID:     "",
				AppSecret: "",
				AllowFrom: FlexibleStringSlice{},
			},
			DingTalk: DingTalkConfig{
				Enabled:      false,
				ClientID:     "",
				ClientSecret: "",
				AllowFrom:    FlexibleStringSlice{},
			},
			Slack: SlackConfig{
				Enabled:   false,
				BotToken:  "",
				AppToken:  "",
				AllowFrom: []string{},
			},
			Signal: SignalConfig{
				Enabled:     false,
				PhoneNumber: "",
				AllowFrom:   FlexibleStringSlice{},
			},
		},
		Providers: ProvidersConfig{
			Anthropic:  ProviderConfig{},
			OpenAI:     ProviderConfig{},
			OpenRouter: ProviderConfig{},
			Groq:       ProviderConfig{},
			Zhipu:      ProviderConfig{},
			VLLM:       ProviderConfig{},
			Gemini:     ProviderConfig{},
			Nvidia:     ProviderConfig{},
			Moonshot:   ProviderConfig{},
			Ollama:     ProviderConfig{},
		},
		Gateway: GatewayConfig{
			Host: "0.0.0.0",
			Port: 18790,
		},
		Web: WebConfig{
			Enabled:   false,
			Host:      "127.0.0.1",
			Port:      18880,
			Username:  "admin",
			Password:  "",
			JWTExpiry: "24h",
		},
		Tools: ToolsConfig{
			Web: WebToolsConfig{
				Search: WebSearchConfig{
					APIKey:     "",
					MaxResults: 5,
				},
			},
			Email: EmailToolsConfig{
				Enabled:  false,
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "",
				Password: "",
				From:     "",
				To:       "",
			},
			MCP: MCPConfig{
				Servers: map[string]MCPServerConfig{},
			},
		},
		Storage: StorageConfig{
			Path: "~/.kakoclaw/kakoclaw.db",
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	// Parse provider environment variables manually (issue #66)
	// caarlos0/env doesn't support {{.Name}} placeholders
	parseProviderEnvVars(cfg)

	if err := validateWebConfig(&cfg.Web); err != nil {
		return nil, fmt.Errorf("web config: %w", err)
	}

	return cfg, nil
}

// LoadConfigForUser loads configuration for a specific user.
// It first tries to load from ~/.kakoclaw/users/<userUUID>/config.json
// If that doesn't exist, it falls back to ~/.kakoclaw/config.json (global)
// If neither exists, it returns DefaultConfig()
func LoadConfigForUser(userUUID string) (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Try user-specific config first
	userConfigPath := filepath.Join(home, ".kakoclaw", "users", userUUID, "config.json")
	if data, err := os.ReadFile(userConfigPath); err == nil {
		cfg := DefaultConfig()
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing user config at %s: %w", userConfigPath, err)
		}
		if err := env.Parse(cfg); err != nil {
			return nil, err
		}
		parseProviderEnvVars(cfg)
		if err := validateWebConfig(&cfg.Web); err != nil {
			return nil, fmt.Errorf("web config: %w", err)
		}
		return cfg, nil
	}

	// Fall back to global config
	globalConfigPath := filepath.Join(home, ".kakoclaw", "config.json")
	cfg, err := LoadConfig(globalConfigPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateWebConfig(w *WebConfig) error {
	if w.Port < 1 || w.Port > 65535 {
		w.Port = 18880
	}
	if strings.TrimSpace(w.Host) == "" {
		w.Host = "127.0.0.1"
	}
	if strings.TrimSpace(w.Username) == "" {
		w.Username = "admin"
	}
	if w.JWTExpiry != "" {
		if _, err := time.ParseDuration(w.JWTExpiry); err != nil {
			return fmt.Errorf("invalid jwt_expiry %q: %w", w.JWTExpiry, err)
		}
	}
	return nil
}

// parseProviderEnvVars manually parses environment variables for providers
// This fixes issue #66 where {{.Name}} placeholders don't work with caarlos0/env
func parseProviderEnvVars(cfg *Config) {
	providers := map[string]*ProviderConfig{
		"anthropic":  &cfg.Providers.Anthropic,
		"openai":     &cfg.Providers.OpenAI,
		"openrouter": &cfg.Providers.OpenRouter,
		"groq":       &cfg.Providers.Groq,
		"zhipu":      &cfg.Providers.Zhipu,
		"vllm":       &cfg.Providers.VLLM,
		"gemini":     &cfg.Providers.Gemini,
		"nvidia":     &cfg.Providers.Nvidia,
		"moonshot":   &cfg.Providers.Moonshot,
		"ollama":     &cfg.Providers.Ollama,
	}

	for name, provider := range providers {
		prefix := fmt.Sprintf("KAKOCLAW_PROVIDERS_%s_", strings.ToUpper(name))

		if apiKey := os.Getenv(prefix + "API_KEY"); apiKey != "" {
			provider.APIKey = apiKey
		}
		if apiBase := os.Getenv(prefix + "API_BASE"); apiBase != "" {
			provider.APIBase = apiBase
		}
		if proxy := os.Getenv(prefix + "PROXY"); proxy != "" {
			provider.Proxy = proxy
		}
		if authMethod := os.Getenv(prefix + "AUTH_METHOD"); authMethod != "" {
			provider.AuthMethod = authMethod
		}
	}
}

func SaveConfig(path string, cfg *Config) error {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) WorkspacePath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return expandHome(c.Agents.Defaults.Workspace)
}

func (c *Config) GetAPIKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Providers.OpenRouter.APIKey != "" {
		return c.Providers.OpenRouter.APIKey
	}
	if c.Providers.Anthropic.APIKey != "" {
		return c.Providers.Anthropic.APIKey
	}
	if c.Providers.OpenAI.APIKey != "" {
		return c.Providers.OpenAI.APIKey
	}
	if c.Providers.Gemini.APIKey != "" {
		return c.Providers.Gemini.APIKey
	}
	if c.Providers.Zhipu.APIKey != "" {
		return c.Providers.Zhipu.APIKey
	}
	if c.Providers.Groq.APIKey != "" {
		return c.Providers.Groq.APIKey
	}
	if c.Providers.VLLM.APIKey != "" {
		return c.Providers.VLLM.APIKey
	}
	return ""
}

func (c *Config) GetAPIBase() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Providers.OpenRouter.APIKey != "" {
		if c.Providers.OpenRouter.APIBase != "" {
			return c.Providers.OpenRouter.APIBase
		}
		return "https://openrouter.ai/api/v1"
	}
	if c.Providers.Zhipu.APIKey != "" {
		return c.Providers.Zhipu.APIBase
	}
	if c.Providers.VLLM.APIKey != "" && c.Providers.VLLM.APIBase != "" {
		return c.Providers.VLLM.APIBase
	}
	return ""
}

func expandHome(path string) string {
	if path == "" {
		return path
	}
	if path[0] == '~' {
		home, _ := os.UserHomeDir()
		if len(path) > 1 && path[1] == '/' {
			return home + path[1:]
		}
		return home
	}
	return path
}

// GetUserConfigPath returns the path to a user's config file
func GetUserConfigPath(userUUID string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kakoclaw", "users", userUUID, "config.json")
}

// SaveConfigForUser saves a user-specific config to their config file
func SaveConfigForUser(userUUID string, cfg *Config) error {
	if userUUID == "" {
		return fmt.Errorf("user UUID is required")
	}

	path := GetUserConfigPath(userUUID)

	// Use the package-level SaveConfig function
	return SaveConfig(path, cfg)
}

// MergeConfigs merges user config over global config at section level.
// Non-empty sections in userCfg override the corresponding sections in globalCfg.
// Empty/zero-value sections in userCfg fall back to globalCfg.
func MergeConfigs(global, user *Config) *Config {
	if global == nil {
		return user
	}
	if user == nil {
		return global
	}

	merged := &Config{}

	// Merge Agents section
	if !isAgentsConfigEmpty(&user.Agents) {
		merged.Agents = user.Agents
	} else {
		merged.Agents = global.Agents
	}

	// Merge Providers section (field-by-field for each provider)
	merged.Providers = mergeProvidersConfig(&global.Providers, &user.Providers)

	// Merge Channels section
	if !isChannelsConfigEmpty(&user.Channels) {
		merged.Channels = user.Channels
	} else {
		merged.Channels = global.Channels
	}

	// Merge Tools section
	merged.Tools = mergeToolsConfig(&global.Tools, &user.Tools)

	// Merge Gateway section
	if !isGatewayConfigEmpty(&user.Gateway) {
		merged.Gateway = user.Gateway
	} else {
		merged.Gateway = global.Gateway
	}

	// Merge Web section
	if !isWebConfigEmpty(&user.Web) {
		merged.Web = user.Web
	} else {
		merged.Web = global.Web
	}

	// Merge Storage section
	if user.Storage.Path != "" {
		merged.Storage = user.Storage
	} else {
		merged.Storage = global.Storage
	}

	return merged
}

// Helper functions to check if config sections are empty

func isAgentsConfigEmpty(a *AgentsConfig) bool {
	return a == nil || (a.Defaults.Workspace == "" && a.Defaults.Provider == "" && a.Defaults.Model == "")
}

func isChannelsConfigEmpty(c *ChannelsConfig) bool {
	return c == nil || (!c.WhatsApp.Enabled && !c.Telegram.Enabled && !c.Feishu.Enabled &&
		!c.Discord.Enabled && !c.MaixCam.Enabled && !c.QQ.Enabled && !c.DingTalk.Enabled &&
		!c.Slack.Enabled && !c.Signal.Enabled)
}

func isGatewayConfigEmpty(g *GatewayConfig) bool {
	return g == nil || (g.Host == "" && g.Port == 0)
}

func isWebConfigEmpty(w *WebConfig) bool {
	return w == nil || (!w.Enabled && w.Host == "" && w.Port == 0)
}

func mergeProvidersConfig(global, user *ProvidersConfig) ProvidersConfig {
	merged := ProvidersConfig{}

	// Helper to merge individual provider
	mergeProvider := func(g, u ProviderConfig) ProviderConfig {
		if u.APIKey != "" || u.APIBase != "" {
			return u
		}
		return g
	}

	merged.Anthropic = mergeProvider(global.Anthropic, user.Anthropic)
	merged.OpenAI = mergeProvider(global.OpenAI, user.OpenAI)
	merged.OpenRouter = mergeProvider(global.OpenRouter, user.OpenRouter)
	merged.Groq = mergeProvider(global.Groq, user.Groq)
	merged.Zhipu = mergeProvider(global.Zhipu, user.Zhipu)
	merged.VLLM = mergeProvider(global.VLLM, user.VLLM)
	merged.Gemini = mergeProvider(global.Gemini, user.Gemini)
	merged.Nvidia = mergeProvider(global.Nvidia, user.Nvidia)
	merged.Moonshot = mergeProvider(global.Moonshot, user.Moonshot)
	merged.Ollama = mergeProvider(global.Ollama, user.Ollama)

	return merged
}

func mergeToolsConfig(global, user *ToolsConfig) ToolsConfig {
	merged := ToolsConfig{}

	// Merge Web tools
	if user.Web.Search.APIKey != "" {
		merged.Web.Search = user.Web.Search
	} else {
		merged.Web.Search = global.Web.Search
	}

	// Merge Email tools
	if user.Email.Enabled || user.Email.Host != "" {
		merged.Email = user.Email
	} else {
		merged.Email = global.Email
	}

	// Merge MCP tools
	if len(user.MCP.Servers) > 0 {
		merged.MCP = user.MCP
	} else {
		merged.MCP = global.MCP
	}

	return merged
}

// ValidateProviderConfig checks if a specific provider has minimum required configuration
func (c *Config) ValidateProviderConfig(providerName string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var provider ProviderConfig
	switch providerName {
	case "anthropic":
		provider = c.Providers.Anthropic
	case "openai":
		provider = c.Providers.OpenAI
	case "openrouter":
		provider = c.Providers.OpenRouter
	case "groq":
		provider = c.Providers.Groq
	case "zhipu":
		provider = c.Providers.Zhipu
	case "vllm":
		provider = c.Providers.VLLM
	case "gemini":
		provider = c.Providers.Gemini
	case "nvidia":
		provider = c.Providers.Nvidia
	case "moonshot":
		provider = c.Providers.Moonshot
	case "ollama":
		provider = c.Providers.Ollama
	default:
		return fmt.Errorf("unknown provider: %s", providerName)
	}

	// Ollama doesn't require API key
	if providerName == "ollama" {
		if provider.APIBase == "" {
			return fmt.Errorf("ollama requires api_base")
		}
		return nil
	}

	// All other providers require API key
	if provider.APIKey == "" {
		return fmt.Errorf("provider %s requires api_key", providerName)
	}

	return nil
}

// GetActiveProviders returns list of providers with valid configurations
func (c *Config) GetActiveProviders() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	active := []string{}
	providers := map[string]ProviderConfig{
		"anthropic":  c.Providers.Anthropic,
		"openai":     c.Providers.OpenAI,
		"openrouter": c.Providers.OpenRouter,
		"groq":       c.Providers.Groq,
		"zhipu":      c.Providers.Zhipu,
		"vllm":       c.Providers.VLLM,
		"gemini":     c.Providers.Gemini,
		"nvidia":     c.Providers.Nvidia,
		"moonshot":   c.Providers.Moonshot,
		"ollama":     c.Providers.Ollama,
	}

	for name := range providers {
		if c.ValidateProviderConfig(name) == nil {
			active = append(active, name)
		}
	}

	return active
}

// GetActiveChannels returns list of channels with valid configurations
func (c *Config) GetActiveChannels() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	active := []string{}

	if c.Channels.Telegram.Enabled && c.Channels.Telegram.Token != "" {
		active = append(active, "telegram")
	}
	if c.Channels.Discord.Enabled && c.Channels.Discord.Token != "" {
		active = append(active, "discord")
	}
	if c.Channels.WhatsApp.Enabled && c.Channels.WhatsApp.BridgeURL != "" {
		active = append(active, "whatsapp")
	}
	if c.Channels.Slack.Enabled && c.Channels.Slack.BotToken != "" {
		active = append(active, "slack")
	}
	if c.Channels.Feishu.Enabled && c.Channels.Feishu.AppID != "" {
		active = append(active, "feishu")
	}
	if c.Channels.QQ.Enabled && c.Channels.QQ.AppID != "" {
		active = append(active, "qq")
	}
	if c.Channels.DingTalk.Enabled && c.Channels.DingTalk.ClientID != "" {
		active = append(active, "dingtalk")
	}
	if c.Channels.Signal.Enabled && c.Channels.Signal.PhoneNumber != "" {
		active = append(active, "signal")
	}
	if c.Channels.MaixCam.Enabled {
		active = append(active, "maixcam")
	}

	return active
}
