package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
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
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Load user config (may not exist, that's ok)
	userCfg, _ := config.LoadConfigForUser(userUUID)

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
		"agents": mergedCfg.Agents,
		"providers": map[string]interface{}{
			"anthropic": map[string]interface{}{
				"api_key":    redactKey(mergedCfg.Providers.Anthropic.APIKey),
				"api_base":   mergedCfg.Providers.Anthropic.APIBase,
				"models":     mergedCfg.Providers.Anthropic.Models,
				"configured": mergedCfg.ValidateProviderConfig("anthropic") == nil,
			},
			"openai": map[string]interface{}{
				"api_key":    redactKey(mergedCfg.Providers.OpenAI.APIKey),
				"api_base":   mergedCfg.Providers.OpenAI.APIBase,
				"models":     mergedCfg.Providers.OpenAI.Models,
				"configured": mergedCfg.ValidateProviderConfig("openai") == nil,
			},
			"openrouter": map[string]interface{}{
				"api_key":    redactKey(mergedCfg.Providers.OpenRouter.APIKey),
				"api_base":   mergedCfg.Providers.OpenRouter.APIBase,
				"models":     mergedCfg.Providers.OpenRouter.Models,
				"configured": mergedCfg.ValidateProviderConfig("openrouter") == nil,
			},
			"groq": map[string]interface{}{
				"api_key":    redactKey(mergedCfg.Providers.Groq.APIKey),
				"api_base":   mergedCfg.Providers.Groq.APIBase,
				"models":     mergedCfg.Providers.Groq.Models,
				"configured": mergedCfg.ValidateProviderConfig("groq") == nil,
			},
			"zhipu": map[string]interface{}{
				"api_key":    redactKey(mergedCfg.Providers.Zhipu.APIKey),
				"api_base":   mergedCfg.Providers.Zhipu.APIBase,
				"models":     mergedCfg.Providers.Zhipu.Models,
				"configured": mergedCfg.ValidateProviderConfig("zhipu") == nil,
			},
			"vllm": map[string]interface{}{
				"api_key":    redactKey(mergedCfg.Providers.VLLM.APIKey),
				"api_base":   mergedCfg.Providers.VLLM.APIBase,
				"models":     mergedCfg.Providers.VLLM.Models,
				"configured": mergedCfg.ValidateProviderConfig("vllm") == nil,
			},
			"gemini": map[string]interface{}{
				"api_key":    redactKey(mergedCfg.Providers.Gemini.APIKey),
				"api_base":   mergedCfg.Providers.Gemini.APIBase,
				"models":     mergedCfg.Providers.Gemini.Models,
				"configured": mergedCfg.ValidateProviderConfig("gemini") == nil,
			},
			"nvidia": map[string]interface{}{
				"api_key":    redactKey(mergedCfg.Providers.Nvidia.APIKey),
				"api_base":   mergedCfg.Providers.Nvidia.APIBase,
				"models":     mergedCfg.Providers.Nvidia.Models,
				"configured": mergedCfg.ValidateProviderConfig("nvidia") == nil,
			},
			"moonshot": map[string]interface{}{
				"api_key":    redactKey(mergedCfg.Providers.Moonshot.APIKey),
				"api_base":   mergedCfg.Providers.Moonshot.APIBase,
				"models":     mergedCfg.Providers.Moonshot.Models,
				"configured": mergedCfg.ValidateProviderConfig("moonshot") == nil,
			},
			"ollama": map[string]interface{}{
				"api_key":    redactKey(mergedCfg.Providers.Ollama.APIKey),
				"api_base":   mergedCfg.Providers.Ollama.APIBase,
				"models":     mergedCfg.Providers.Ollama.Models,
				"configured": mergedCfg.ValidateProviderConfig("ollama") == nil,
			},
		},
		"channels": redactChannels(mergedCfg),
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
				"password": map[string]interface{}{"has_password": mergedCfg.Tools.Email.Password != ""}, // Indicador sin valor real
				"from":     mergedCfg.Tools.Email.From,
				"to":       mergedCfg.Tools.Email.To,
			},
		},
	}

	// Determine if user has custom overrides
	hasOverride := userCfg != nil && !isConfigEmpty(userCfg)

	sources := buildConfigSources(globalCfg, userCfg)
	if len(mergedCfg.GetActiveProviders()) == 0 {
		removeProviderSources(sources)
	}

	response := ConfigWithSource{
		Config:      redacted,
		Sources:     sources,
		HasOverride: hasOverride,
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// handleUpdateUserConfig updates a user's config section
func (s *Server) handleUpdateUserConfig(w http.ResponseWriter, r *http.Request) {
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var update map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	update = normalizeUpdateKeys(update)

	// Log the update operation
	logger.InfoCF("web", "Updating user config", map[string]interface{}{
		"user_uuid":        userUUID,
		"updated_sections": getMapKeysFromMap(update),
	})

	// Load existing user config or create new one
	userCfg, _ := config.LoadConfigForUser(userUUID)
	if userCfg == nil {
		userCfg = config.DefaultConfig()
	}

	// Apply updates to user config
	if err := applyConfigUpdates(userCfg, update); err != nil {
		http.Error(w, fmt.Sprintf("failed to apply updates: %v", err), http.StatusBadRequest)
		return
	}

	// Save user config
	if err := config.SaveConfigForUser(userUUID, userCfg); err != nil {
		logger.ErrorCF("web", "Failed to save user config", map[string]interface{}{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		http.Error(w, "failed to save config", http.StatusInternalServerError)
		return
	}

	logger.InfoCF("web", "User config updated", map[string]interface{}{
		"user_uuid": userUUID,
	})

	// Restart user's channels asynchronously to avoid blocking the response.
	// Use context.Background() because the HTTP request context will be cancelled
	// after the response is sent, which would kill the restarted channels/agent loop.
	if s.multiUserChannelManager != nil {
		go func() {
			if err := s.multiUserChannelManager.RestartUserChannels(context.Background(), userUUID); err != nil {
				logger.WarnCF("web", "Failed to restart user channels after config update", map[string]interface{}{
					"user_uuid": userUUID,
					"error":     err.Error(),
				})
			} else {
				logger.InfoCF("web", "Restarted user channels after config update", map[string]interface{}{
					"user_uuid": userUUID,
				})
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Configuration updated successfully",
	})
}

// handleDeleteUserConfigSection deletes a config section, reverting to global default
func (s *Server) handleDeleteUserConfigSection(w http.ResponseWriter, r *http.Request) {
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	section := r.URL.Query().Get("section")
	if section == "" {
		http.Error(w, "section parameter required", http.StatusBadRequest)
		return
	}

	// Load user config
	userCfg, _ := config.LoadConfigForUser(userUUID)
	if userCfg == nil {
		// Nothing to delete
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]interface{}{
			"success": true,
			"message": "No override to remove",
		})
		return
	}

	// Reset the specified section to zero values
	resetConfigSection(userCfg, section)

	// Save updated config
	if err := config.SaveConfigForUser(userUUID, userCfg); err != nil {
		http.Error(w, "failed to save config", http.StatusInternalServerError)
		return
	}

	logger.InfoCF("web", "User config section reset", map[string]interface{}{
		"user_uuid": userUUID,
		"section":   section,
	})

	// Restart user's channels if multi-user channel manager is available
	if s.multiUserChannelManager != nil {
		if err := s.multiUserChannelManager.RestartUserChannels(r.Context(), userUUID); err != nil {
			logger.WarnCF("web", "Failed to restart user channels after config reset", map[string]interface{}{
				"user_uuid": userUUID,
				"section":   section,
				"error":     err.Error(),
			})
			// Don't fail the request, config was saved successfully
		} else {
			logger.InfoCF("web", "Restarted user channels after config reset", map[string]interface{}{
				"user_uuid": userUUID,
				"section":   section,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
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
	// SPECIAL CASE: Deep merge for agents and tools to avoid overwriting entire top-level sections
	if agentsUpdate, ok := updates["agents"].(map[string]interface{}); ok {
		// Handle defaults
		if defaultsUpdate, ok := agentsUpdate["defaults"].(map[string]interface{}); ok {
			if workspace, ok := defaultsUpdate["workspace"].(string); ok {
				cfg.Agents.Defaults.Workspace = workspace
			}
			if restrict, ok := defaultsUpdate["restrict_to_workspace"].(bool); ok {
				cfg.Agents.Defaults.RestrictToWorkspace = restrict
			}
			if provider, ok := defaultsUpdate["provider"].(string); ok {
				cfg.Agents.Defaults.Provider = provider
			}
			if model, ok := defaultsUpdate["model"].(string); ok {
				cfg.Agents.Defaults.Model = model
			}
			if tokens, ok := defaultsUpdate["max_tokens"].(float64); ok {
				cfg.Agents.Defaults.MaxTokens = int(tokens)
			}
			if temperature, ok := defaultsUpdate["temperature"].(float64); ok {
				cfg.Agents.Defaults.Temperature = temperature
			}
			if iterations, ok := defaultsUpdate["max_tool_iterations"].(float64); ok {
				cfg.Agents.Defaults.MaxToolIterations = int(iterations)
			}
		}
		// Handle orchestrator
		if orchUpdate, ok := agentsUpdate["orchestrator"].(map[string]interface{}); ok {
			if enabled, ok := orchUpdate["enabled"].(bool); ok {
				cfg.Agents.Orchestrator.Enabled = enabled
			}
			if provider, ok := orchUpdate["provider"].(string); ok {
				cfg.Agents.Orchestrator.Provider = provider
			}
			if model, ok := orchUpdate["model"].(string); ok {
				cfg.Agents.Orchestrator.Model = model
			}
			if tokens, ok := orchUpdate["max_tokens"].(float64); ok {
				cfg.Agents.Orchestrator.MaxTokens = int(tokens)
			}
			if temperature, ok := orchUpdate["temperature"].(float64); ok {
				cfg.Agents.Orchestrator.Temperature = temperature
			}
			if retries, ok := orchUpdate["max_delegation_retries"].(float64); ok {
				cfg.Agents.Orchestrator.MaxDelegationRetries = int(retries)
			}
			if fallback, ok := orchUpdate["fallback_to_default"].(bool); ok {
				cfg.Agents.Orchestrator.FallbackToDefault = fallback
			}
			if description, ok := orchUpdate["description"].(string); ok {
				cfg.Agents.Orchestrator.Description = description
			}
		}
		// Handle specialists (partial map merge)
		if specUpdate, ok := agentsUpdate["specialists"].(map[string]interface{}); ok {
			if cfg.Agents.Specialists == nil {
				cfg.Agents.Specialists = make(map[string]config.SpecialistConfig)
			}
			for name, specRaw := range specUpdate {
				specJSON, _ := json.Marshal(specRaw)
				var spec config.SpecialistConfig
				if err := json.Unmarshal(specJSON, &spec); err == nil {
					cfg.Agents.Specialists[name] = spec
				}
			}
		}
		// Remove "agents" from bulk update map to avoid shallow overwrite
		delete(updates, "agents")
	}

	if channelsUpdate, ok := updates["channels"].(map[string]interface{}); ok {
		for channelName, rawUpdate := range channelsUpdate {
			channelPatch, ok := rawUpdate.(map[string]interface{})
			if !ok {
				continue
			}

			switch channelName {
			case "whatsapp":
				if err := mergeStructFromMap(&cfg.Channels.WhatsApp, channelPatch); err != nil {
					return err
				}
			case "telegram":
				if t, ok := channelPatch["token"].(string); ok && t == "" {
					delete(channelPatch, "token")
				}
				if err := mergeStructFromMap(&cfg.Channels.Telegram, channelPatch); err != nil {
					return err
				}
			case "feishu":
				for _, secretKey := range []string{"app_secret", "encrypt_key", "verification_token"} {
					if s, ok := channelPatch[secretKey].(string); ok && s == "" {
						delete(channelPatch, secretKey)
					}
				}
				if err := mergeStructFromMap(&cfg.Channels.Feishu, channelPatch); err != nil {
					return err
				}
			case "discord":
				if t, ok := channelPatch["token"].(string); ok && t == "" {
					delete(channelPatch, "token")
				}
				if err := mergeStructFromMap(&cfg.Channels.Discord, channelPatch); err != nil {
					return err
				}
			case "maixcam":
				if err := mergeStructFromMap(&cfg.Channels.MaixCam, channelPatch); err != nil {
					return err
				}
			case "qq":
				if s, ok := channelPatch["app_secret"].(string); ok && s == "" {
					delete(channelPatch, "app_secret")
				}
				if err := mergeStructFromMap(&cfg.Channels.QQ, channelPatch); err != nil {
					return err
				}
			case "dingtalk":
				if s, ok := channelPatch["client_secret"].(string); ok && s == "" {
					delete(channelPatch, "client_secret")
				}
				if err := mergeStructFromMap(&cfg.Channels.DingTalk, channelPatch); err != nil {
					return err
				}
			case "slack":
				if t, ok := channelPatch["bot_token"].(string); ok && t == "" {
					delete(channelPatch, "bot_token")
				}
				if t, ok := channelPatch["app_token"].(string); ok && t == "" {
					delete(channelPatch, "app_token")
				}
				if err := mergeStructFromMap(&cfg.Channels.Slack, channelPatch); err != nil {
					return err
				}
			case "signal":
				if err := mergeStructFromMap(&cfg.Channels.Signal, channelPatch); err != nil {
					return err
				}
			case "email":
				if p, ok := channelPatch["password"].(string); ok && p == "" {
					delete(channelPatch, "password")
				}
				if err := mergeStructFromMap(&cfg.Channels.Email, channelPatch); err != nil {
					return err
				}
			}
		}

		delete(updates, "channels")
	}

	if providersUpdate, ok := updates["providers"].(map[string]interface{}); ok {
		for providerName, rawUpdate := range providersUpdate {
			providerPatch, ok := rawUpdate.(map[string]interface{})
			if !ok {
				continue
			}

			var target *config.ProviderConfig
			switch providerName {
			case "anthropic":
				target = &cfg.Providers.Anthropic
			case "openai":
				target = &cfg.Providers.OpenAI
			case "openrouter":
				target = &cfg.Providers.OpenRouter
			case "groq":
				target = &cfg.Providers.Groq
			case "zhipu":
				target = &cfg.Providers.Zhipu
			case "vllm":
				target = &cfg.Providers.VLLM
			case "gemini":
				target = &cfg.Providers.Gemini
			case "nvidia":
				target = &cfg.Providers.Nvidia
			case "moonshot":
				target = &cfg.Providers.Moonshot
			case "ollama":
				target = &cfg.Providers.Ollama
			default:
				continue
			}

			if apiKey, ok := providerPatch["api_key"].(string); ok {
				if apiKey != "" && !strings.Contains(apiKey, "****") {
					target.APIKey = apiKey
				}
			}
			if apiBase, ok := providerPatch["api_base"].(string); ok {
				target.APIBase = apiBase
			}
			if proxy, ok := providerPatch["proxy"].(string); ok {
				target.Proxy = proxy
			}
			if authMethod, ok := providerPatch["auth_method"].(string); ok {
				target.AuthMethod = authMethod
			}
			if modelsRaw, ok := providerPatch["models"].([]interface{}); ok {
				models := make([]string, 0, len(modelsRaw))
				for _, modelRaw := range modelsRaw {
					if model, ok := modelRaw.(string); ok {
						models = append(models, model)
					}
				}
				target.Models = models
			}
		}

		delete(updates, "providers")
	}

	if toolsUpdate, ok := updates["tools"].(map[string]interface{}); ok {
		// Handle web
		if webUpdate, ok := toolsUpdate["web"].(map[string]interface{}); ok {
			if searchUpdate, ok := webUpdate["search"].(map[string]interface{}); ok {
				searchJSON, _ := json.Marshal(searchUpdate)
				json.Unmarshal(searchJSON, &cfg.Tools.Web.Search)
			}
		}
		// Handle email - special case for password
		if emailUpdate, ok := toolsUpdate["email"].(map[string]interface{}); ok {
			// Check if password is being updated
			passwordValue := emailUpdate["password"]
			passwordChanged := false

			// If password is a map with has_password field, it's from GET response, not an update
			if passwordMap, ok := passwordValue.(map[string]interface{}); ok {
				// This is just the indicator from GET, remove it and don't update password
				_ = passwordMap // Use the variable
				delete(emailUpdate, "password")
			} else if passwordStr, ok := passwordValue.(string); ok {
				// Actual password string provided
				if passwordStr != "" {
					passwordChanged = true
				} else {
					// Empty password means don't update existing password
					delete(emailUpdate, "password")
				}
			} else {
				// Password field missing or null, don't update
				delete(emailUpdate, "password")
			}

			// Only update email config if there are actual changes
			emailJSON, _ := json.Marshal(emailUpdate)
			var tempCfg config.EmailToolsConfig
			json.Unmarshal(emailJSON, &tempCfg)

			// Preserve existing password if not being changed
			if !passwordChanged && cfg.Tools.Email.Password != "" {
				tempCfg.Password = cfg.Tools.Email.Password
			}

			cfg.Tools.Email = tempCfg
		}
		// Remove "tools" from bulk update map
		delete(updates, "tools")
	}

	// Marshaling/Unmarshaling approach for the rest of simpler fields
	currentJSON, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	var intermediate map[string]interface{}
	if err := json.Unmarshal(currentJSON, &intermediate); err != nil {
		return err
	}

	// Apply remaining updates
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

func mergeStructFromMap[T any](target *T, patch map[string]interface{}) error {
	currentJSON, err := json.Marshal(target)
	if err != nil {
		return err
	}

	var currentMap map[string]interface{}
	if err := json.Unmarshal(currentJSON, &currentMap); err != nil {
		return err
	}

	for key, value := range patch {
		currentMap[key] = value
	}

	mergedJSON, err := json.Marshal(currentMap)
	if err != nil {
		return err
	}

	return json.Unmarshal(mergedJSON, target)
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
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the authenticated user's ID for provider config lookup
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	cfg, err := store.GetUserProvidersConfig(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get provider config: %v", err), http.StatusInternalServerError)
		return
	}

	response := redactUserProviders(cfg)

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// handleUpdateUserProvider updates a specific provider configuration for the user
func (s *Server) handleUpdateUserProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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

	if err := store.UpdateUserProviderConfig(userID, providerName, providerConfig); err != nil {
		http.Error(w, fmt.Sprintf("failed to update provider config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSONResponse(w, map[string]string{
		"message":  fmt.Sprintf("Provider '%s' configuration updated", providerName),
		"provider": providerName,
	})
}

// handleGetUserChannels returns the current user's channel configuration
func (s *Server) handleGetUserChannels(w http.ResponseWriter, r *http.Request) {
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Load user config (may not exist, that's ok)
	userCfg, _ := config.LoadConfigForUser(userUUID)

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
	writeJSONResponse(w, response)
}

// getMapKeysFromMap returns the keys of a map as a string slice
func getMapKeysFromMap(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
