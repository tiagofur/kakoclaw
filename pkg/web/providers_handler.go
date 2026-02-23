package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sipeed/makoclaw/pkg/storage"
)

// ProvidersHandler handles provider configuration endpoints
type ProvidersHandler struct {
	store *storage.Storage
}

// ProviderUpdateRequest represents a request to update a provider configuration
type ProviderUpdateRequest struct {
	APIKey     string   `json:"api_key"`
	APIBase    string   `json:"api_base"`
	Proxy      string   `json:"proxy,omitempty"`
	AuthMethod string   `json:"auth_method,omitempty"`
	Models     []string `json:"models,omitempty"`
}

// ProviderResponse represents a provider configuration in responses
type ProviderResponse struct {
	APIKey     string   `json:"api_key"`
	APIBase    string   `json:"api_base"`
	Proxy      string   `json:"proxy,omitempty"`
	AuthMethod string   `json:"auth_method,omitempty"`
	Models     []string `json:"models,omitempty"`
}

// ProvidersConfigResponse represents all provider configurations
type ProvidersConfigResponse struct {
	Anthropic  ProviderResponse `json:"anthropic"`
	OpenAI     ProviderResponse `json:"openai"`
	OpenRouter ProviderResponse `json:"openrouter"`
	Groq       ProviderResponse `json:"groq"`
	Zhipu      ProviderResponse `json:"zhipu"`
	VLLM       ProviderResponse `json:"vllm"`
	Gemini     ProviderResponse `json:"gemini"`
	Nvidia     ProviderResponse `json:"nvidia"`
	Moonshot   ProviderResponse `json:"moonshot"`
	Ollama     ProviderResponse `json:"ollama"`
}

// NewProvidersHandler creates a new handlers for provider configuration
func NewProvidersHandler(store *storage.Storage) *ProvidersHandler {
	return &ProvidersHandler{
		store: store,
	}
}

// GetProvidersConfig returns the current user's provider configuration
func (h *ProvidersHandler) GetProvidersConfig(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	config, err := h.store.GetUserProvidersConfig(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get provider config: %v", err), http.StatusInternalServerError)
		return
	}

	response := ProvidersConfigResponse{
		Anthropic:  convertProviderConfig(config.Anthropic),
		OpenAI:     convertProviderConfig(config.OpenAI),
		OpenRouter: convertProviderConfig(config.OpenRouter),
		Groq:       convertProviderConfig(config.Groq),
		Zhipu:      convertProviderConfig(config.Zhipu),
		VLLM:       convertProviderConfig(config.VLLM),
		Gemini:     convertProviderConfig(config.Gemini),
		Nvidia:     convertProviderConfig(config.Nvidia),
		Moonshot:   convertProviderConfig(config.Moonshot),
		Ollama:     convertProviderConfig(config.Ollama),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProviderConfig updates a specific provider configuration
func (h *ProvidersHandler) UpdateProviderConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("user_id").(int64)
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

	var req ProviderUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	providerConfig := storage.ProviderConfig{
		APIKey:     req.APIKey,
		APIBase:    req.APIBase,
		Proxy:      req.Proxy,
		AuthMethod: req.AuthMethod,
		Models:     req.Models,
	}

	if err := h.store.UpdateUserProviderConfig(userID, providerName, providerConfig); err != nil {
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

// Helper function to convert storage ProviderConfig to response format
func convertProviderConfig(cfg storage.ProviderConfig) ProviderResponse {
	return ProviderResponse{
		APIKey:     cfg.APIKey,
		APIBase:    cfg.APIBase,
		Proxy:      cfg.Proxy,
		AuthMethod: cfg.AuthMethod,
		Models:     cfg.Models,
	}
}
