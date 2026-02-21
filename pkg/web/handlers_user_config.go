package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/logger"
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

	// Build redacted view using existing functions
	redacted := map[string]interface{}{
		"agents":    mergedCfg.Agents,
		"providers": redactProviders(mergedCfg),
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

	response := ConfigWithSource{
		Config:      redacted,
		Sources:     buildConfigSources(globalCfg, userCfg),
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

	// Restart user's channels if multi-user channel manager is available
	if s.multiUserChannelManager != nil {
		if err := s.multiUserChannelManager.RestartUserChannels(r.Context(), user.UUID); err != nil {
			logger.WarnCF("web", "Failed to restart user channels after config update", map[string]interface{}{
				"user_uuid": user.UUID,
				"error":     err.Error(),
			})
			// Don't fail the request, config was saved successfully
		} else {
			logger.InfoCF("web", "Restarted user channels after config update", map[string]interface{}{
				"user_uuid": user.UUID,
			})
		}
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

// handleGetUserProviders returns list of active providers for the user
func (s *Server) handleGetUserProviders(w http.ResponseWriter, r *http.Request) {
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

	// Load user's merged config
	userCfg, _ := config.LoadConfigForUser(user.UUID)
	mergedCfg := config.MergeConfigs(s.fullConfig, userCfg)

	active := mergedCfg.GetActiveProviders()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"active_providers": active,
	})
}

// handleGetUserChannels returns list of active channels for the user
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

	// Load user's merged config
	userCfg, _ := config.LoadConfigForUser(user.UUID)
	mergedCfg := config.MergeConfigs(s.fullConfig, userCfg)

	active := mergedCfg.GetActiveChannels()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"active_channels": active,
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
