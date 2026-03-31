package config_test

import (
	"encoding/json"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
)

func TestDevStudioConfig_ZeroValueIsDisabled(t *testing.T) {
	var c config.Config

	if c.DevStudio.Enabled {
		t.Error("Expected DevStudio to be disabled by default on zero-value Config")
	}
	if c.DevStudio.Memory.Enabled {
		t.Error("Expected DevStudio Memory to be disabled by default on zero-value Config")
	}
}

func TestDevStudioConfig_DefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.DevStudio.Enabled {
		t.Error("DefaultConfig() should have DevStudio disabled by default")
	}
	if cfg.DevStudio.DefaultBackend != "claude-code" {
		t.Errorf("Default backend should be 'claude-code', got %q", cfg.DevStudio.DefaultBackend)
	}
	if cfg.DevStudio.MaxSessionTokens != 200000 {
		t.Errorf("Default session tokens should be 200000, got %d", cfg.DevStudio.MaxSessionTokens)
	}
	if cfg.DevStudio.Memory.Enabled {
		t.Error("DefaultConfig() should have DevStudio Memory disabled by default")
	}
}

func TestDevStudioConfig_JSONSerialization(t *testing.T) {
	data := []byte(`{
		"dev_studio": {
			"enabled": true,
			"default_backend": "opencode",
			"memory": {
				"enabled": true,
				"model": "custom-model"
			}
		}
	}`)

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if !cfg.DevStudio.Enabled {
		t.Error("Expected DevStudio to be enabled via JSON")
	}
	if cfg.DevStudio.DefaultBackend != "opencode" {
		t.Errorf("Expected backend 'opencode', got %q", cfg.DevStudio.DefaultBackend)
	}
	if !cfg.DevStudio.Memory.Enabled {
		t.Error("Expected memory to be enabled via JSON")
	}
	if cfg.DevStudio.Memory.Model != "custom-model" {
		t.Errorf("Expected memory model 'custom-model', got %q", cfg.DevStudio.Memory.Model)
	}
}
