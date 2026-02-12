package providers

import (
	"testing"
)

func TestGetProviderForModel(t *testing.T) {
	tests := []struct {
		model         string
		wantProvider  string
		wantModel     string
	}{
		// Issue #43: Explicit provider/model syntax
		{"openai/gpt-4", "openai", "gpt-4"},
		{"anthropic/claude-3-sonnet", "anthropic", "claude-3-sonnet"},
		{"openrouter/meta-llama/llama-3", "openrouter", "meta-llama/llama-3"},
		{"groq/llama-3.1-70b", "groq", "llama-3.1-70b"},
		{"zhipu/glm-4", "zhipu", "glm-4"},
		{"gemini/gemini-pro", "gemini", "gemini-pro"},
		{"moonshot/kimi-k2", "moonshot", "kimi-k2"},
		{"nvidia/meta/llama-3.1-405b", "nvidia", "meta/llama-3.1-405b"},
		
		// Auto-detection (no prefix)
		{"gpt-4", "openai", "gpt-4"},
		{"gpt-3.5-turbo", "openai", "gpt-3.5-turbo"},
		{"o1-preview", "openai", "o1-preview"},
		{"claude-3-opus-20240229", "anthropic", "claude-3-opus-20240229"},
		{"kimi-k2", "moonshot", "kimi-k2"},
		{"gemini-1.5-pro", "gemini", "gemini-1.5-pro"},
		{"glm-4", "zhipu", "glm-4"},
		{"llama-3-70b", "openrouter", "llama-3-70b"},
		{"mixtral-8x22b", "openrouter", "mixtral-8x22b"},
		{"custom-model", "openrouter", "custom-model"},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			gotProvider, gotModel := GetProviderForModel(tt.model)
			if gotProvider != tt.wantProvider {
				t.Errorf("GetProviderForModel(%q) provider = %q, want %q", 
					tt.model, gotProvider, tt.wantProvider)
			}
			if gotModel != tt.wantModel {
				t.Errorf("GetProviderForModel(%q) model = %q, want %q", 
					tt.model, gotModel, tt.wantModel)
			}
		})
	}
}
