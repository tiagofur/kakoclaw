package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// UserProvidersConfig stores per-user LLM provider configurations
type UserProvidersConfig struct {
	UserID     int64          `json:"user_id"`
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

// ProviderConfig stores API key, base URL, and proxy for a provider
type ProviderConfig struct {
	APIKey     string   `json:"api_key"`
	APIBase    string   `json:"api_base"`
	Proxy      string   `json:"proxy,omitempty"`
	AuthMethod string   `json:"auth_method,omitempty"`
	Models     []string `json:"models,omitempty"`
}

// migrateUserProviders creates the user_providers_config table
func (s *Storage) migrateUserProviders() error {
	query := `CREATE TABLE IF NOT EXISTS user_providers_config (
		user_id INTEGER PRIMARY KEY,
		config TEXT NOT NULL
	);`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("user_providers_config migration: %w", err)
	}
	return nil
}

// SaveUserProvidersConfig saves the providers configuration for a specific user
func (s *Storage) SaveUserProvidersConfig(userID int64, providers *UserProvidersConfig) error {
	configJSON, err := json.Marshal(providers)
	if err != nil {
		return fmt.Errorf("marshal providers config: %w", err)
	}

	query := `INSERT INTO user_providers_config (user_id, config) VALUES (?, ?)
	ON CONFLICT(user_id) DO UPDATE SET config = excluded.config`

	_, err = s.db.Exec(query, userID, string(configJSON))
	if err != nil {
		return fmt.Errorf("save user providers config: %w", err)
	}
	return nil
}

// GetUserProvidersConfig retrieves the providers configuration for a specific user
func (s *Storage) GetUserProvidersConfig(userID int64) (*UserProvidersConfig, error) {
	query := `SELECT config FROM user_providers_config WHERE user_id = ?`

	var configJSON string
	err := s.db.QueryRow(query, userID).Scan(&configJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty config if user hasn't set any providers yet
			return &UserProvidersConfig{UserID: userID}, nil
		}
		return nil, fmt.Errorf("get user providers config: %w", err)
	}

	var providers UserProvidersConfig
	err = json.Unmarshal([]byte(configJSON), &providers)
	if err != nil {
		return nil, fmt.Errorf("unmarshal providers config: %w", err)
	}

	providers.UserID = userID
	return &providers, nil
}

// UpdateUserProviderConfig updates a single provider configuration for a user
func (s *Storage) UpdateUserProviderConfig(userID int64, providerName string, config ProviderConfig) error {
	// First get existing config
	providers, err := s.GetUserProvidersConfig(userID)
	if err != nil {
		return err
	}

	// Update the specific provider
	switch providerName {
	case "anthropic":
		providers.Anthropic = config
	case "openai":
		providers.OpenAI = config
	case "openrouter":
		providers.OpenRouter = config
	case "groq":
		providers.Groq = config
	case "zhipu":
		providers.Zhipu = config
	case "vllm":
		providers.VLLM = config
	case "gemini":
		providers.Gemini = config
	case "nvidia":
		providers.Nvidia = config
	case "moonshot":
		providers.Moonshot = config
	case "ollama":
		providers.Ollama = config
	default:
		return fmt.Errorf("unknown provider: %s", providerName)
	}

	return s.SaveUserProvidersConfig(userID, providers)
}

// DeleteUserProvidersConfig deletes all provider configurations for a user
func (s *Storage) DeleteUserProvidersConfig(userID int64) error {
	query := `DELETE FROM user_providers_config WHERE user_id = ?`
	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("delete user providers config: %w", err)
	}
	return nil
}
