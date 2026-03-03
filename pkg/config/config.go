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
	"github.com/sipeed/makoclaw/pkg/logger"
)

var (
	dataDir   string
	dataDirMu sync.RWMutex
)

// InitDataDir initializes the base data directory for KakoClaw.
// This should be the directory that contains the 'users' folder and/or the central.db.
func InitDataDir(path string) {
	dataDirMu.Lock()
	defer dataDirMu.Unlock()
	dataDir = path
}

func getDataDir() string {
	dataDirMu.RLock()
	defer dataDirMu.RUnlock()
	if dataDir != "" {
		return dataDir
	}

	// Fallback to user home directory (standard location for Docker volumes)
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".MakoClaw")
}

// GetDataDir returns the MakoClaw data directory (exported version)
func GetDataDir() string {
	return getDataDir()
}

// FlexibleStringSlice is a []string that also accepts JSON numbers,
// so allow_from can contain both "123" and 123.
type FlexibleStringSlice []string

func (f *FlexibleStringSlice) UnmarshalJSON(data []byte) error {
	// Try plain string first (e.g. "123,456" from form inputs)
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			*f = FlexibleStringSlice{}
		} else {
			parts := strings.Split(s, ",")
			result := make([]string, 0, len(parts))
			for _, p := range parts {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					result = append(result, trimmed)
				}
			}
			*f = result
		}
		return nil
	}

	// Try []string
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
	Agents          AgentsConfig          `json:"agents"`
	Channels        ChannelsConfig        `json:"channels"`
	Providers       ProvidersConfig       `json:"providers"`
	Gateway         GatewayConfig         `json:"gateway"`
	Web             WebConfig             `json:"web"`
	Tools           ToolsConfig           `json:"tools"`
	ToolPermissions ToolPermissionsConfig `json:"tool_permissions"`
	Storage         StorageConfig         `json:"storage"`
	DegradedMode    bool                  `json:"-"` // Runtime flag: true when no valid LLM provider is configured
	mu              sync.RWMutex
}

type StorageConfig struct {
	Path string `json:"path" env:"MAKOCLAW_STORAGE_PATH"`
}

type AgentsConfig struct {
	Defaults           AgentDefaults               `json:"defaults"`
	Orchestrator       OrchestratorConfig          `json:"orchestrator"`
	Specialists        map[string]SpecialistConfig `json:"specialists"`
	RemovedSpecialists []string                    `json:"removed_specialists,omitempty"` // Agents explicitly removed by user
}

type AgentDefaults struct {
	Workspace           string  `json:"workspace" env:"MAKOCLAW_AGENTS_DEFAULTS_WORKSPACE"`
	RestrictToWorkspace bool    `json:"restrict_to_workspace" env:"MAKOCLAW_AGENTS_DEFAULTS_RESTRICT_TO_WORKSPACE"`
	Provider            string  `json:"provider" env:"MAKOCLAW_AGENTS_DEFAULTS_PROVIDER"`
	Model               string  `json:"model" env:"MAKOCLAW_AGENTS_DEFAULTS_MODEL"`
	MaxTokens           int     `json:"max_tokens" env:"MAKOCLAW_AGENTS_DEFAULTS_MAX_TOKENS"`
	Temperature         float64 `json:"temperature" env:"MAKOCLAW_AGENTS_DEFAULTS_TEMPERATURE"`
	MaxToolIterations   int     `json:"max_tool_iterations" env:"MAKOCLAW_AGENTS_DEFAULTS_MAX_TOOL_ITERATIONS"`
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
	Skills            []string `json:"skills,omitempty"` // Skill names to load (omitted=all, empty=none)
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
	Enabled   bool                `json:"enabled" env:"MAKOCLAW_CHANNELS_WHATSAPP_ENABLED"`
	BridgeURL string              `json:"bridge_url" env:"MAKOCLAW_CHANNELS_WHATSAPP_BRIDGE_URL"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"MAKOCLAW_CHANNELS_WHATSAPP_ALLOW_FROM"`
}

type TelegramConfig struct {
	Enabled   bool                `json:"enabled" env:"MAKOCLAW_CHANNELS_TELEGRAM_ENABLED"`
	Token     string              `json:"token" env:"MAKOCLAW_CHANNELS_TELEGRAM_TOKEN"`
	Proxy     string              `json:"proxy" env:"MAKOCLAW_CHANNELS_TELEGRAM_PROXY"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"MAKOCLAW_CHANNELS_TELEGRAM_ALLOW_FROM"`
}

type FeishuConfig struct {
	Enabled           bool                `json:"enabled" env:"MAKOCLAW_CHANNELS_FEISHU_ENABLED"`
	AppID             string              `json:"app_id" env:"MAKOCLAW_CHANNELS_FEISHU_APP_ID"`
	AppSecret         string              `json:"app_secret" env:"MAKOCLAW_CHANNELS_FEISHU_APP_SECRET"`
	EncryptKey        string              `json:"encrypt_key" env:"MAKOCLAW_CHANNELS_FEISHU_ENCRYPT_KEY"`
	VerificationToken string              `json:"verification_token" env:"MAKOCLAW_CHANNELS_FEISHU_VERIFICATION_TOKEN"`
	AllowFrom         FlexibleStringSlice `json:"allow_from" env:"MAKOCLAW_CHANNELS_FEISHU_ALLOW_FROM"`
}

type DiscordConfig struct {
	Enabled   bool                `json:"enabled" env:"MAKOCLAW_CHANNELS_DISCORD_ENABLED"`
	Token     string              `json:"token" env:"MAKOCLAW_CHANNELS_DISCORD_TOKEN"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"MAKOCLAW_CHANNELS_DISCORD_ALLOW_FROM"`
}

type MaixCamConfig struct {
	Enabled   bool                `json:"enabled" env:"MAKOCLAW_CHANNELS_MAIXCAM_ENABLED"`
	Host      string              `json:"host" env:"MAKOCLAW_CHANNELS_MAIXCAM_HOST"`
	Port      int                 `json:"port" env:"MAKOCLAW_CHANNELS_MAIXCAM_PORT"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"MAKOCLAW_CHANNELS_MAIXCAM_ALLOW_FROM"`
}

type QQConfig struct {
	Enabled   bool                `json:"enabled" env:"MAKOCLAW_CHANNELS_QQ_ENABLED"`
	AppID     string              `json:"app_id" env:"MAKOCLAW_CHANNELS_QQ_APP_ID"`
	AppSecret string              `json:"app_secret" env:"MAKOCLAW_CHANNELS_QQ_APP_SECRET"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"MAKOCLAW_CHANNELS_QQ_ALLOW_FROM"`
}

type DingTalkConfig struct {
	Enabled      bool                `json:"enabled" env:"MAKOCLAW_CHANNELS_DINGTALK_ENABLED"`
	ClientID     string              `json:"client_id" env:"MAKOCLAW_CHANNELS_DINGTALK_CLIENT_ID"`
	ClientSecret string              `json:"client_secret" env:"MAKOCLAW_CHANNELS_DINGTALK_CLIENT_SECRET"`
	AllowFrom    FlexibleStringSlice `json:"allow_from" env:"MAKOCLAW_CHANNELS_DINGTALK_ALLOW_FROM"`
}

type SlackConfig struct {
	Enabled   bool                `json:"enabled" env:"MAKOCLAW_CHANNELS_SLACK_ENABLED"`
	BotToken  string              `json:"bot_token" env:"MAKOCLAW_CHANNELS_SLACK_BOT_TOKEN"`
	AppToken  string              `json:"app_token" env:"MAKOCLAW_CHANNELS_SLACK_APP_TOKEN"`
	AllowFrom FlexibleStringSlice `json:"allow_from" env:"MAKOCLAW_CHANNELS_SLACK_ALLOW_FROM"`
}

type SignalConfig struct {
	Enabled     bool                `json:"enabled" env:"MAKOCLAW_CHANNELS_SIGNAL_ENABLED"`
	PhoneNumber string              `json:"phone_number" env:"MAKOCLAW_CHANNELS_SIGNAL_PHONE_NUMBER"`
	AllowFrom   FlexibleStringSlice `json:"allow_from" env:"MAKOCLAW_CHANNELS_SIGNAL_ALLOW_FROM"`
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
	Host string `json:"host" env:"MAKOCLAW_GATEWAY_HOST"`
	Port int    `json:"port" env:"MAKOCLAW_GATEWAY_PORT"`
}

type WebConfig struct {
	Enabled   bool   `json:"enabled" env:"MAKOCLAW_WEB_ENABLED"`
	Host      string `json:"host" env:"MAKOCLAW_WEB_HOST"`
	Port      int    `json:"port" env:"MAKOCLAW_WEB_PORT"`
	Username  string `json:"username" env:"MAKOCLAW_WEB_USERNAME"`
	Password  string `json:"password" env:"MAKOCLAW_WEB_PASSWORD"`
	JWTExpiry string `json:"jwt_expiry" env:"MAKOCLAW_WEB_JWT_EXPIRY"`
}

type WebSearchConfig struct {
	APIKey     string `json:"api_key" env:"MAKOCLAW_TOOLS_WEB_SEARCH_API_KEY"`
	MaxResults int    `json:"max_results" env:"MAKOCLAW_TOOLS_WEB_SEARCH_MAX_RESULTS"`
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
	Enabled  bool   `json:"enabled" env:"MAKOCLAW_TOOLS_EMAIL_ENABLED"`
	Host     string `json:"host" env:"MAKOCLAW_TOOLS_EMAIL_HOST"`
	Port     int    `json:"port" env:"MAKOCLAW_TOOLS_EMAIL_PORT"`
	Username string `json:"username" env:"MAKOCLAW_TOOLS_EMAIL_USERNAME"`
	Password string `json:"password" env:"MAKOCLAW_TOOLS_EMAIL_PASSWORD"`
	From     string `json:"from" env:"MAKOCLAW_TOOLS_EMAIL_FROM"`
	To       string `json:"to" env:"MAKOCLAW_TOOLS_EMAIL_TO"`
}

// ToolPermissionsConfig defines tool access control by role and user overrides
type ToolPermissionsConfig struct {
	// RoleDefaults maps role names ("admin", "user") to lists of allowed tools
	// Special value "*" means all tools
	// Special value "exec_restricted" means exec with allowlist only
	RoleDefaults map[string][]string `json:"role_defaults"`

	// AllowedShellCommands is the list of safe shell commands that users can execute
	// Only applies when "exec_restricted" is in user's tool list
	AllowedShellCommands []string `json:"allowed_shell_commands"`

	// UserOverrides maps usernames to custom tool lists, overriding role defaults
	// Use null/empty to reset to role defaults
	UserOverrides map[string][]string `json:"user_overrides,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Agents: AgentsConfig{
			Defaults: AgentDefaults{
				Workspace:           "~/.MakoClaw/workspace",
				RestrictToWorkspace: true,
				Provider:            "",
				Model:               "",
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
				AllowFrom: FlexibleStringSlice{},
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
		ToolPermissions: ToolPermissionsConfig{
			RoleDefaults: map[string][]string{
				"admin": {"*"}, // Admins get all tools
				"user": {
					// File operations (restricted to user workspace)
					"read_file",
					"write_file",
					"edit_file",
					"append_file",
					"list_dir",
					// Task and knowledge management
					"task_manager",
					"query_knowledge",
					"memory",
					// Web access
					"web_search",
					"web_fetch",
					// Communication
					"message",
					// Limited shell access via allowlist
					"exec_restricted",
				},
			},
			AllowedShellCommands: []string{
				"ls", "dir", "cat", "type", "head", "tail", "grep", "findstr",
				"find", "where", "pwd", "cd", "echo", "date", "whoami",
				"which", "wc", "sort", "uniq", "diff", "tree", "file", "stat",
			},
			UserOverrides: map[string][]string{},
		},
		Storage: StorageConfig{
			Path: "~/.MakoClaw/makoclaw.db",
		},
	}
}

// getEmptyChannelsConfig returns a ChannelsConfig with all channels disabled
// and no sensitive tokens/credentials
func getEmptyChannelsConfig() ChannelsConfig {
	return ChannelsConfig{
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
			AllowFrom: FlexibleStringSlice{},
		},
		Signal: SignalConfig{
			Enabled:     false,
			PhoneNumber: "",
			AllowFrom:   FlexibleStringSlice{},
		},
	}
}

// getEmptyProvidersConfigWithBaseURLs returns a ProvidersConfig with no API keys
// but inherits base URLs from global config if available
func getEmptyProvidersConfigWithBaseURLs(globalConfig *Config) ProvidersConfig {
	// Helper to get base URL safely
	getBaseURL := func(provider *ProviderConfig, defaultURL string) string {
		if provider != nil && provider.APIBase != "" {
			return provider.APIBase
		}
		return defaultURL
	}

	return ProvidersConfig{
		Anthropic: ProviderConfig{
			APIKey:  "",
			APIBase: getBaseURL(&globalConfig.Providers.Anthropic, "https://api.anthropic.com"),
		},
		OpenAI: ProviderConfig{
			APIKey:  "",
			APIBase: getBaseURL(&globalConfig.Providers.OpenAI, "https://api.openai.com/v1"),
		},
		OpenRouter: ProviderConfig{
			APIKey:  "",
			APIBase: getBaseURL(&globalConfig.Providers.OpenRouter, "https://openrouter.ai/api/v1"),
		},
		Groq: ProviderConfig{
			APIKey:  "",
			APIBase: getBaseURL(&globalConfig.Providers.Groq, "https://api.groq.com/openai/v1"),
		},
		Zhipu: ProviderConfig{
			APIKey:  "",
			APIBase: getBaseURL(&globalConfig.Providers.Zhipu, ""),
		},
		VLLM: ProviderConfig{
			APIKey:  "",
			APIBase: getBaseURL(&globalConfig.Providers.VLLM, ""),
		},
		Gemini: ProviderConfig{
			APIKey:  "",
			APIBase: getBaseURL(&globalConfig.Providers.Gemini, ""),
		},
		Nvidia: ProviderConfig{
			APIKey:  "",
			APIBase: getBaseURL(&globalConfig.Providers.Nvidia, ""),
		},
		Moonshot: ProviderConfig{
			APIKey:  "",
			APIBase: getBaseURL(&globalConfig.Providers.Moonshot, ""),
		},
		Ollama: ProviderConfig{
			APIKey:  "",
			APIBase: getBaseURL(&globalConfig.Providers.Ollama, "http://localhost:11434/v1"),
		},
	}
}

// GetUserConfigTemplate creates a clean config template for a new user.
// Inherits safe defaults from global config but NOT sensitive data (API keys, tokens, passwords).
// This ensures each new user starts with a clean configuration while inhererring
// non-sensitive settings like model defaults, max_tokens, temperature, and provider base URLs.
func GetUserConfigTemplate(globalConfig *Config) *Config {
	if globalConfig == nil {
		globalConfig = DefaultConfig()
	}

	return &Config{
		Agents: AgentsConfig{
			Defaults: AgentDefaults{
				// Inherit non-sensitive defaults
				Workspace:           filepath.Join("~", ".MakoClaw", "users", "{uuid}", "workspace"),
				RestrictToWorkspace: globalConfig.Agents.Defaults.RestrictToWorkspace,
				Provider:            "",                                             // Empty: user must choose
				Model:               globalConfig.Agents.Defaults.Model,             // Inherit
				MaxTokens:           globalConfig.Agents.Defaults.MaxTokens,         // Inherit
				Temperature:         globalConfig.Agents.Defaults.Temperature,       // Inherit
				MaxToolIterations:   globalConfig.Agents.Defaults.MaxToolIterations, // Inherit
			},
			Orchestrator: OrchestratorConfig{
				Enabled:              false, // Always start disabled
				Provider:             "",
				Model:                "",
				MaxTokens:            globalConfig.Agents.Orchestrator.MaxTokens,
				Temperature:          globalConfig.Agents.Orchestrator.Temperature,
				MaxDelegationRetries: globalConfig.Agents.Orchestrator.MaxDelegationRetries,
				FallbackToDefault:    globalConfig.Agents.Orchestrator.FallbackToDefault,
				Description:          "Project Manager: Analyzes tasks and delegates to appropriate specialists",
			},
			Specialists:        make(map[string]SpecialistConfig),
			RemovedSpecialists: []string{},
		},
		Channels:  getEmptyChannelsConfig(),
		Providers: getEmptyProvidersConfigWithBaseURLs(globalConfig),
		Gateway: GatewayConfig{
			Host: globalConfig.Gateway.Host,
			Port: globalConfig.Gateway.Port,
		},
		Web: WebConfig{
			Enabled:   false, // User enables if needed
			Host:      "127.0.0.1",
			Port:      18880,
			Username:  "admin",
			Password:  "",
			JWTExpiry: "24h",
		},
		Tools: ToolsConfig{
			Web: WebToolsConfig{
				Search: WebSearchConfig{
					APIKey:     "",                                       // User must provide
					MaxResults: globalConfig.Tools.Web.Search.MaxResults, // Inherit
				},
			},
			Email: EmailToolsConfig{
				Enabled:  false,
				Host:     "", // Empty: user configures their own
				Port:     587,
				Username: "",
				Password: "",
				From:     "",
				To:       "",
			},
			MCP: MCPConfig{
				Servers: make(map[string]MCPServerConfig),
			},
		},
		Storage: StorageConfig{
			Path: filepath.Join("~", ".MakoClaw", "users", "{uuid}", "user.db"),
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
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
// It first tries to load from ~/.MakoClaw/users/<userUUID>/config.json
// If that doesn't exist, it falls back to ~/.MakoClaw/config.json (global)
// If neither exists, it returns DefaultConfig()
func LoadConfigForUser(userUUID string) (*Config, error) {
	baseDir := getDataDir()

	// Try user-specific config first
	userConfigPath := filepath.Join(baseDir, "users", userUUID, "config.json")

	if data, err := os.ReadFile(userConfigPath); err == nil {
		userCfg := DefaultConfig()
		if err := json.Unmarshal(data, userCfg); err != nil {
			return nil, fmt.Errorf("parsing user config at %s: %w", userConfigPath, err)
		}

		// Merge global config for default agents
		// This ensures that default agents from global config appear even after user creates their own
		globalConfigPath := filepath.Join(baseDir, "config.json")
		if globalCfg, err := LoadConfig(globalConfigPath); err == nil {
			// Initialize specialists map if nil
			if userCfg.Agents.Specialists == nil {
				userCfg.Agents.Specialists = make(map[string]SpecialistConfig)
			}

			// Initialize removed specialists list if nil
			if userCfg.Agents.RemovedSpecialists == nil {
				userCfg.Agents.RemovedSpecialists = []string{}
			}

			// Create a map of removed specialists for quick lookup
			removedMap := make(map[string]bool)
			for _, name := range userCfg.Agents.RemovedSpecialists {
				removedMap[name] = true
			}

			// Merge specialists: user specialists override global, but keep global defaults
			// Skip specialists that the user explicitly removed
			for name, spec := range globalCfg.Agents.Specialists {
				if _, exists := userCfg.Agents.Specialists[name]; !exists {
					// Only inherit if user hasn't explicitly removed it
					if !removedMap[name] {
						userCfg.Agents.Specialists[name] = spec
					}
				}
			}

			// Merge orchestrator config if not set in user config
			if userCfg.Agents.Orchestrator.Provider == "" && globalCfg.Agents.Orchestrator.Provider != "" {
				userCfg.Agents.Orchestrator = globalCfg.Agents.Orchestrator
			}
		}

		// User config must be loaded as persisted. Do not apply generic env parsing
		// or provider env var overrides here, otherwise process-level env values can
		// unintentionally override user-saved settings (e.g. tools.email.enabled).
		if err := validateWebConfig(&userCfg.Web); err != nil {
			return nil, fmt.Errorf("web config: %w", err)
		}

		logger.DebugCF("config", "Loaded user config", map[string]interface{}{
			"user_uuid":          userUUID,
			"user_config_exists": true,
			"has_email_config":   userCfg.Tools.Email.Host != "",
			"has_search_api_key": userCfg.Tools.Web.Search.APIKey != "",
		})

		return userCfg, nil
	}

	// Fall back to global config
	globalConfigPath := filepath.Join(baseDir, "config.json")
	cfg, err := LoadConfig(globalConfigPath)
	if err != nil {
		return nil, err
	}

	logger.DebugCF("config", "Loading user config - using global fallback", map[string]interface{}{
		"user_uuid":             userUUID,
		"user_config_exists":    false,
		"using_global_fallback": true,
	})

	// Lazily seed the user config.json so subsequent loads find a user-specific file.
	// We only do this if the user's directory already exists (i.e. the user has been
	// fully provisioned). If the directory doesn't exist yet we skip silently.
	//
	// New users get a clean config template that inherits safe defaults (model, max_tokens, etc.)
	// but NOT sensitive data (API keys, tokens, passwords).
	userDir := filepath.Join(baseDir, "users", userUUID)
	if info, statErr := os.Stat(userDir); statErr == nil && info.IsDir() {
		// Create clean config template for new user (inherits safe defaults, NOT sensitive data)
		userConfigTemplate := GetUserConfigTemplate(cfg)

		// Replace {uuid} placeholder with actual UUID
		userConfigTemplate.Agents.Defaults.Workspace = filepath.Join(userDir, "workspace")
		userConfigTemplate.Storage.Path = filepath.Join(userDir, "user.db")

		if saveErr := SaveConfig(userConfigPath, userConfigTemplate); saveErr != nil {
			logger.WarnCF("config", "Failed to seed user config.json", map[string]interface{}{
				"user_uuid": userUUID,
				"error":     saveErr.Error(),
			})
		} else {
			logger.InfoCF("config", "Created clean config template for new user", map[string]interface{}{
				"user_uuid": userUUID,
				"path":      userConfigPath,
			})
		}
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
		prefix := fmt.Sprintf("MAKOCLAW_PROVIDERS_%s_", strings.ToUpper(name))

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

// GetGlobalConfigPath returns the path to the global config file
func GetGlobalConfigPath() string {
	return filepath.Join(getDataDir(), "config.json")
}

// GetUserConfigPath returns the path to a user's config file
func GetUserConfigPath(userUUID string) string {
	return filepath.Join(getDataDir(), "users", userUUID, "config.json")
}

// SaveConfigForUser saves a user-specific config to their config file
func SaveConfigForUser(userUUID string, cfg *Config) error {
	if userUUID == "" {
		return fmt.Errorf("user UUID is required")
	}

	path := GetUserConfigPath(userUUID)

	logger.InfoCF("config", "Saving user config", map[string]interface{}{
		"user_uuid":          userUUID,
		"path":               path,
		"has_email_config":   cfg.Tools.Email.Host != "",
		"has_search_api_key": cfg.Tools.Web.Search.APIKey != "",
	})

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

	// Merge ToolPermissions section
	if len(user.ToolPermissions.RoleDefaults) > 0 {
		merged.ToolPermissions = user.ToolPermissions
	} else {
		merged.ToolPermissions = global.ToolPermissions
	}

	// Preserve DegradedMode flag
	merged.DegradedMode = global.DegradedMode || user.DegradedMode

	return merged
}

// Helper functions to check if config sections are empty

func isAgentsConfigEmpty(a *AgentsConfig) bool {
	if a == nil {
		return true
	}

	// Check if user has meaningful agent configuration
	// Consider it non-empty if:
	// 1. Orchestrator is explicitly enabled
	// 2. User has custom specialists defined
	// 3. User has removed default specialists
	// 4. Default agent fields are customized

	hasOrchestratorEnabled := a.Orchestrator.Enabled
	hasSpecialists := len(a.Specialists) > 0
	hasRemovedSpecialists := len(a.RemovedSpecialists) > 0
	hasCustomDefaults := a.Defaults.Workspace != "" ||
		a.Defaults.Provider != "" ||
		a.Defaults.Model != ""

	return !(hasOrchestratorEnabled || hasSpecialists || hasRemovedSpecialists || hasCustomDefaults)
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
	return c.validateProviderConfigLocked(providerName)
}

// validateProviderConfigLocked is the lock-free inner implementation.
// Caller must hold at least c.mu.RLock().
func (c *Config) validateProviderConfigLocked(providerName string) error {
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

	// All other providers require API key or auth method (e.g., OAuth)
	if provider.APIKey == "" && provider.AuthMethod == "" {
		return fmt.Errorf("provider %s requires api_key or auth_method", providerName)
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
		if c.validateProviderConfigLocked(name) == nil {
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

// RLock locks the config for reading
func (c *Config) RLock() {
	c.mu.RLock()
}

// RUnlock unlocks the config after reading
func (c *Config) RUnlock() {
	c.mu.RUnlock()
}

// Lock locks the config for writing
func (c *Config) Lock() {
	c.mu.Lock()
}

// Unlock unlocks the config after writing
func (c *Config) Unlock() {
	c.mu.Unlock()
}

// HasValidProviderConfig checks if there is a valid LLM provider configuration.
// Returns true if default agent has a valid provider+model, or if orchestrator is enabled with valid config.
// This is used to determine if the system can run in full mode vs degraded mode.
func (c *Config) HasValidProviderConfig() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if default agent has valid provider and model
	if c.Agents.Defaults.Provider != "" && c.Agents.Defaults.Model != "" {
		// Validate that the provider is actually configured
		if c.validateProviderConfigLocked(c.Agents.Defaults.Provider) == nil {
			return true
		}
	}

	// Check if orchestrator is enabled and has valid config
	if c.Agents.Orchestrator.Enabled {
		if c.Agents.Orchestrator.Provider != "" && c.Agents.Orchestrator.Model != "" {
			if c.validateProviderConfigLocked(c.Agents.Orchestrator.Provider) == nil {
				return true
			}
		}
	}

	// Check if any specialist has valid config
	for _, spec := range c.Agents.Specialists {
		if spec.Provider != "" && spec.Model != "" {
			if c.validateProviderConfigLocked(spec.Provider) == nil {
				return true
			}
		}
	}

	return false
}
