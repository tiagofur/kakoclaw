package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/logger"
	"github.com/sipeed/kakoclaw/pkg/storage"
)

// ConfigSource indicates where a config value comes from
type ConfigSource string

const (
	ConfigSourceUser   ConfigSource = "user"   // User-specific override
	ConfigSourceGlobal ConfigSource = "global" // Global default
)

// ConfigWithSource wraps config sections with source indicators
type ConfigWithSource struct {
	Config      interface{}             `json:"config"`
	Sources     map[string]ConfigSource `json:"sources"` // field path -> source
	HasOverride bool                    `json:"has_override"`
}

// handleGetUserConfig returns the user's merged config with source indicators
func (s *Server) handleGetUserConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := s.store.GetUserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Load user config (may not exist, that's ok)
	userCfg, _ := config.LoadConfigForUser(user.UUID)

	// Load global config
	globalCfg := s.fullConfig
	if globalCfg == nil {
		http.Error(w, "global config not available", http.StatusInternalServerError)
		return
	}

	// Merge configs
	mergedCfg := config.MergeConfigs(globalCfg, userCfg)

	userProviders, err := s.store.GetUserProvidersConfig(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get provider config: %v", err), http.StatusInternalServerError)
		return
	}

	// Build redacted view using existing functions
	redacted := map[string]interface{}{
		"agents":    mergedCfg.Agents,
		"providers": redactUserProviders(userProviders),
		"channels":  redactChannels(mergedCfg),
		"tools": map[string]interface{}{
			"web": map[string]interface{}{
				"search": map[string]interface{}{
					"api_key":     redactKey(mergedCfg.Tools.Web.Search.APIKey),
					"max_results": mergedCfg.Tools.Web.Search.MaxResults,
				},
			},
			"email": map[string]interface{}{
				"enabled":  mergedCfg.Tools.Email.Enabled,
				"host":     mergedCfg.Tools.Email.Host,
				"port":     mergedCfg.Tools.Email.Port,
				"username": mergedCfg.Tools.Email.Username,
				"from":     mergedCfg.Tools.Email.From,
				"to":       mergedCfg.Tools.Email.To,
			},
		},
	}

	// Determine if user has custom overrides
	hasOverride := userCfg != nil && !isConfigEmpty(userCfg)

	sources := buildConfigSources(globalCfg, userCfg)
	if isUserProvidersEmpty(userProviders) {
		removeProviderSources(sources)
	}

	response := ConfigWithSource{
		Config:      redacted,
		Sources:     sources,
		HasOverride: hasOverride,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUpdateUserConfig updates a user's config section
func (s *Server) handleUpdateUserConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := s.store.GetUserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	var update map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	update = normalizeUpdateKeys(update)

	// Load existing user config or create new one
	userCfg, _ := config.LoadConfigForUser(user.UUID)
	if userCfg == nil {
		userCfg = config.DefaultConfig()
	}

	// Apply updates to user config
	if err := applyConfigUpdates(userCfg, update); err != nil {
		http.Error(w, fmt.Sprintf("failed to apply updates: %v", err), http.StatusBadRequest)
		return
	}

	// Save user config
	if err := config.SaveConfigForUser(user.UUID, userCfg); err != nil {
		logger.ErrorCF("web", "Failed to save user config", map[string]interface{}{
			"user_uuid": user.UUID,
			"error":     err.Error(),
		})
		http.Error(w, "failed to save config", http.StatusInternalServerError)
		return
	}

	logger.InfoCF("web", "User config updated", map[string]interface{}{
		"user_uuid": user.UUID,
		"username":  user.Username,
	})

	// Restart user's channels asynchronously to avoid blocking the response
	if s.multiUserChannelManager != nil {
		go func() {
			ctx := context.Background()
			if err := s.multiUserChannelManager.RestartUserChannels(ctx, user.UUID); err != nil {
				logger.WarnCF("web", "Failed to restart user channels after config update", map[string]interface{}{
					"user_uuid": user.UUID,
					"error":     err.Error(),
				})
			} else {
				logger.InfoCF("web", "Restarted user channels after config update", map[string]interface{}{
					"user_uuid": user.UUID,
				})
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Configuration updated successfully",
	})
}

// handleDeleteUserConfigSection deletes a config section, reverting to global default
func (s *Server) handleDeleteUserConfigSection(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := s.store.GetUserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	section := r.URL.Query().Get("section")
	if section == "" {
		http.Error(w, "section parameter required", http.StatusBadRequest)
		return
	}

	// Load user config
	userCfg, _ := config.LoadConfigForUser(user.UUID)
	if userCfg == nil {
		// Nothing to delete
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "No override to remove",
		})
		return
	}

	// Reset the specified section to zero values
	resetConfigSection(userCfg, section)

	// Save updated config
	if err := config.SaveConfigForUser(user.UUID, userCfg); err != nil {
		http.Error(w, "failed to save config", http.StatusInternalServerError)
		return
	}

	logger.InfoCF("web", "User config section reset", map[string]interface{}{
		"user_uuid": user.UUID,
		"section":   section,
	})

	// Restart user's channels if multi-user channel manager is available
	if s.multiUserChannelManager != nil {
		if err := s.multiUserChannelManager.RestartUserChannels(r.Context(), user.UUID); err != nil {
			logger.WarnCF("web", "Failed to restart user channels after config reset", map[string]interface{}{
				"user_uuid": user.UUID,
				"section":   section,
				"error":     err.Error(),
			})
			// Don't fail the request, config was saved successfully
		} else {
			logger.InfoCF("web", "Restarted user channels after config reset", map[string]interface{}{
				"user_uuid": user.UUID,
				"section":   section,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Section '%s' reset to global default", section),
	})
}

// Helper functions

func buildConfigSources(global, user *config.Config) map[string]ConfigSource {
	sources := make(map[string]ConfigSource)

	// This is simplified - in a real implementation, you'd check each field
	if user != nil {
		// Check providers
		if user.Providers.Anthropic.APIKey != "" {
			sources["providers.anthropic"] = ConfigSourceUser
		} else if global != nil && global.Providers.Anthropic.APIKey != "" {
			sources["providers.anthropic"] = ConfigSourceGlobal
		}

		if user.Providers.OpenAI.APIKey != "" {
			sources["providers.openai"] = ConfigSourceUser
		} else if global != nil && global.Providers.OpenAI.APIKey != "" {
			sources["providers.openai"] = ConfigSourceGlobal
		}

		// Check channels
		if user.Channels.Telegram.Enabled {
			sources["channels.telegram"] = ConfigSourceUser
		} else if global != nil && global.Channels.Telegram.Enabled {
			sources["channels.telegram"] = ConfigSourceGlobal
		}
	}

	return sources
}

func redactUserProviders(cfg *storage.UserProvidersConfig) map[string]interface{} {
	return map[string]interface{}{
		"anthropic": map[string]string{
			"api_key":  redactKey(cfg.Anthropic.APIKey),
			"api_base": cfg.Anthropic.APIBase,
		},
		"openai": map[string]string{
			"api_key":  redactKey(cfg.OpenAI.APIKey),
			"api_base": cfg.OpenAI.APIBase,
		},
		"openrouter": map[string]string{
			"api_key":  redactKey(cfg.OpenRouter.APIKey),
			"api_base": cfg.OpenRouter.APIBase,
		},
		"groq": map[string]string{
			"api_key":  redactKey(cfg.Groq.APIKey),
			"api_base": cfg.Groq.APIBase,
		},
		"zhipu": map[string]string{
			"api_key":  redactKey(cfg.Zhipu.APIKey),
			"api_base": cfg.Zhipu.APIBase,
		},
		"vllm": map[string]string{
			"api_key":  redactKey(cfg.VLLM.APIKey),
			"api_base": cfg.VLLM.APIBase,
		},
		"gemini": map[string]string{
			"api_key":  redactKey(cfg.Gemini.APIKey),
			"api_base": cfg.Gemini.APIBase,
		},
		"nvidia": map[string]string{
			"api_key":  redactKey(cfg.Nvidia.APIKey),
			"api_base": cfg.Nvidia.APIBase,
		},
		"moonshot": map[string]string{
			"api_key":  redactKey(cfg.Moonshot.APIKey),
			"api_base": cfg.Moonshot.APIBase,
		},
		"ollama": map[string]string{
			"api_key":  redactKey(cfg.Ollama.APIKey),
			"api_base": cfg.Ollama.APIBase,
		},
	}
}

func isProviderConfigEmpty(cfg storage.ProviderConfig) bool {
	return cfg.APIKey == "" && cfg.APIBase == "" && cfg.Proxy == "" && cfg.AuthMethod == "" && len(cfg.Models) == 0
}

func isUserProvidersEmpty(cfg *storage.UserProvidersConfig) bool {
	if cfg == nil {
		return true
	}

	return isProviderConfigEmpty(cfg.Anthropic) &&
		isProviderConfigEmpty(cfg.OpenAI) &&
		isProviderConfigEmpty(cfg.OpenRouter) &&
		isProviderConfigEmpty(cfg.Groq) &&
		isProviderConfigEmpty(cfg.Zhipu) &&
		isProviderConfigEmpty(cfg.VLLM) &&
		isProviderConfigEmpty(cfg.Gemini) &&
		isProviderConfigEmpty(cfg.Nvidia) &&
		isProviderConfigEmpty(cfg.Moonshot) &&
		isProviderConfigEmpty(cfg.Ollama)
}

func removeProviderSources(sources map[string]ConfigSource) {
	for key := range sources {
		if strings.HasPrefix(key, "providers.") {
			delete(sources, key)
		}
	}
}

func isConfigEmpty(cfg *config.Config) bool {
	if cfg == nil {
		return true
	}
	// Check if any non-default values are set
	return cfg.Providers.Anthropic.APIKey == "" &&
		cfg.Providers.OpenAI.APIKey == "" &&
		cfg.Channels.Telegram.Token == "" &&
		cfg.Tools.Email.Host == ""
}

func applyConfigUpdates(cfg *config.Config, updates map[string]interface{}) error {
	// Marshal current config to intermediate format
	currentJSON, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	var intermediate map[string]interface{}
	if err := json.Unmarshal(currentJSON, &intermediate); err != nil {
		return err
	}

	// Apply updates
	for key, value := range updates {
		intermediate[key] = value
	}

	// Unmarshal back to Config
	updatedJSON, err := json.Marshal(intermediate)
	if err != nil {
		return err
	}

	return json.Unmarshal(updatedJSON, cfg)
}

func normalizeUpdateKeys(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return map[string]interface{}{}
	}

	output := make(map[string]interface{}, len(input))
	for key, value := range input {
		output[toSnakeCase(key)] = normalizeUpdateValue(value)
	}
	return output
}

func normalizeUpdateValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return normalizeUpdateKeys(typed)
	case []interface{}:
		normalized := make([]interface{}, 0, len(typed))
		for _, item := range typed {
			normalized = append(normalized, normalizeUpdateValue(item))
		}
		return normalized
	default:
		return value
	}
}

func toSnakeCase(input string) string {
	if input == "" {
		return input
	}

	var builder strings.Builder
	builder.Grow(len(input) + 4)
	for i, r := range input {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				builder.WriteByte('_')
			}
			builder.WriteRune(r + ('a' - 'A'))
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func resetConfigSection(cfg *config.Config, section string) {
	switch section {
	case "providers":
		cfg.Providers = config.ProvidersConfig{}
	case "channels":
		cfg.Channels = config.ChannelsConfig{}
	case "tools":
		cfg.Tools = config.ToolsConfig{}
	case "agents":
		cfg.Agents = config.AgentsConfig{}
	}
}

// handleGetUserProviders returns the current user's provider configuration
func (s *Server) handleGetUserProviders(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if s.store == nil {
		http.Error(w, "storage not available", http.StatusInternalServerError)
		return
	}

	config, err := s.store.GetUserProvidersConfig(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get provider config: %v", err), http.StatusInternalServerError)
		return
	}

	response := redactUserProviders(config)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUpdateUserProvider updates a specific provider configuration for the user
func (s *Server) handleUpdateUserProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if s.store == nil {
		http.Error(w, "storage not available", http.StatusInternalServerError)
		return
	}

	// Get provider name from query parameter
	providerName := r.URL.Query().Get("provider")
	if providerName == "" {
		http.Error(w, "provider parameter is required", http.StatusBadRequest)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	apiKey, _ := req["api_key"].(string)
	apiBase, _ := req["api_base"].(string)
	proxy, _ := req["proxy"].(string)
	authMethod, _ := req["auth_method"].(string)

	providerConfig := storage.ProviderConfig{
		APIKey:     apiKey,
		APIBase:    apiBase,
		Proxy:      proxy,
		AuthMethod: authMethod,
	}

	if err := s.store.UpdateUserProviderConfig(userID, providerName, providerConfig); err != nil {
		http.Error(w, fmt.Sprintf("failed to update provider config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  fmt.Sprintf("Provider '%s' configuration updated", providerName),
		"provider": providerName,
	})
}

// handleGetUserChannels returns the current user's channel configuration
func (s *Server) handleGetUserChannels(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := s.store.GetUserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Load user config (may not exist, that's ok)
	userCfg, _ := config.LoadConfigForUser(user.UUID)

	// Load global config
	globalCfg := s.fullConfig
	if globalCfg == nil {
		http.Error(w, "global config not available", http.StatusInternalServerError)
		return
	}

	// Merge configs
	mergedCfg := config.MergeConfigs(globalCfg, userCfg)

	// Return channels configuration
	response := map[string]interface{}{
		"channels": map[string]interface{}{
			"telegram": map[string]interface{}{
				"enabled":    mergedCfg.Channels.Telegram.Enabled,
				"token":      redactKey(mergedCfg.Channels.Telegram.Token),
				"proxy":      mergedCfg.Channels.Telegram.Proxy,
				"allow_from": mergedCfg.Channels.Telegram.AllowFrom,
			},
			"discord": map[string]interface{}{
				"enabled":    mergedCfg.Channels.Discord.Enabled,
				"token":      redactKey(mergedCfg.Channels.Discord.Token),
				"allow_from": mergedCfg.Channels.Discord.AllowFrom,
			},
			"whatsapp": map[string]interface{}{
				"enabled":    mergedCfg.Channels.WhatsApp.Enabled,
				"bridge_url": mergedCfg.Channels.WhatsApp.BridgeURL,
				"allow_from": mergedCfg.Channels.WhatsApp.AllowFrom,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
