package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sipeed/kakoclaw/pkg/config"
)

// Status represents the health check result type
type Status int

const (
	StatusOK Status = iota
	StatusWarning
	StatusError
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusWarning:
		return "WARNING"
	case StatusError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (s Status) Icon() string {
	switch s {
	case StatusOK:
		return "✓"
	case StatusWarning:
		return "⚠"
	case StatusError:
		return "✗"
	default:
		return "?"
	}
}

// CheckResult represents a single health check result
type CheckResult struct {
	Name    string
	Status  Status
	Message string
	Fix     string
}

// RunChecks performs all health checks and returns the results
func RunChecks(configPath string) []CheckResult {
	var results []CheckResult

	results = append(results, checkConfigFile(configPath))
	results = append(results, checkWorkspace(configPath))
	results = append(results, checkAPIKeys(configPath))
	results = append(results, checkProviders(configPath))
	results = append(results, checkPermissions(configPath))
	results = append(results, checkDirectories())

	return results
}

func checkConfigFile(configPath string) CheckResult {
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return CheckResult{
			Name:    "Configuration File",
			Status:  StatusError,
			Message: fmt.Sprintf("Config file not found at %s", configPath),
			Fix:     "Run 'KakoClaw onboard' to initialize configuration",
		}
	}

	_, err := os.ReadFile(configPath)
	if err != nil {
		return CheckResult{
			Name:    "Configuration File",
			Status:  StatusError,
			Message: "Cannot read config file",
			Fix:     "Check file permissions with 'ls -la ~/.KakoClaw/'",
		}
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return CheckResult{
			Name:    "Configuration File",
			Status:  StatusError,
			Message: fmt.Sprintf("Invalid JSON: %v", err),
			Fix:     "Fix JSON syntax errors in config file",
		}
	}

	// Check for empty critical fields
	if cfg.Agents.Defaults.Model == "" {
		return CheckResult{
			Name:    "Configuration File",
			Status:  StatusWarning,
			Message: "No default model configured",
			Fix:     "Add 'model' to agents.defaults in config.json",
		}
	}

	return CheckResult{
		Name:    "Configuration File",
		Status:  StatusOK,
		Message: fmt.Sprintf("Valid config at %s", configPath),
	}
}

func checkWorkspace(configPath string) CheckResult {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return CheckResult{
			Name:    "Workspace",
			Status:  StatusWarning,
			Message: "Cannot load config to check workspace",
		}
	}

	workspace := cfg.WorkspacePath()
	if workspace == "" {
		return CheckResult{
			Name:    "Workspace",
			Status:  StatusError,
			Message: "No workspace configured",
			Fix:     "Set workspace path in config.json",
		}
	}

	info, err := os.Stat(workspace)
	if os.IsNotExist(err) {
		return CheckResult{
			Name:    "Workspace",
			Status:  StatusError,
			Message: fmt.Sprintf("Workspace directory does not exist: %s", workspace),
			Fix:     "Run 'KakoClaw onboard' or create the directory manually",
		}
	}

	if !info.IsDir() {
		return CheckResult{
			Name:    "Workspace",
			Status:  StatusError,
			Message: fmt.Sprintf("Workspace path is not a directory: %s", workspace),
			Fix:     "Remove the file and create a directory, or change workspace path",
		}
	}

	// Check if workspace is writable
	testFile := filepath.Join(workspace, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return CheckResult{
			Name:    "Workspace",
			Status:  StatusError,
			Message: fmt.Sprintf("Workspace is not writable: %s", workspace),
			Fix:     "Check directory permissions with 'chmod 755'",
		}
	}
	os.Remove(testFile)

	return CheckResult{
		Name:    "Workspace",
		Status:  StatusOK,
		Message: fmt.Sprintf("Writable workspace at %s", workspace),
	}
}

func checkAPIKeys(configPath string) CheckResult {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return CheckResult{
			Name:    "API Keys",
			Status:  StatusWarning,
			Message: "Cannot load config to check API keys",
		}
	}

	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return CheckResult{
			Name:    "API Keys",
			Status:  StatusError,
			Message: "No API key configured for any provider",
			Fix:     "Add API key to config.json or set environment variable (e.g., KakoClaw_PROVIDERS_OPENROUTER_API_KEY)",
		}
	}

	// Check key format (basic validation)
	if len(apiKey) < 10 {
		return CheckResult{
			Name:    "API Keys",
			Status:  StatusWarning,
			Message: "API key seems too short, might be invalid",
			Fix:     "Verify the API key is correct",
		}
	}

	return CheckResult{
		Name:    "API Keys",
		Status:  StatusOK,
		Message: "At least one API key is configured",
	}
}

func checkProviders(configPath string) CheckResult {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return CheckResult{
			Name:    "Providers",
			Status:  StatusWarning,
			Message: "Cannot load config to check providers",
		}
	}

	model := cfg.Agents.Defaults.Model
	if model == "" {
		return CheckResult{
			Name:    "Providers",
			Status:  StatusWarning,
			Message: "No model configured, cannot determine provider",
		}
	}

	// Count configured providers
	providerCount := 0
	providers := map[string]bool{
		"Anthropic":  cfg.Providers.Anthropic.APIKey != "",
		"OpenAI":     cfg.Providers.OpenAI.APIKey != "",
		"OpenRouter": cfg.Providers.OpenRouter.APIKey != "",
		"Groq":       cfg.Providers.Groq.APIKey != "",
		"Zhipu":      cfg.Providers.Zhipu.APIKey != "",
		"Gemini":     cfg.Providers.Gemini.APIKey != "",
		"Moonshot":   cfg.Providers.Moonshot.APIKey != "",
		"Nvidia":     cfg.Providers.Nvidia.APIKey != "",
	}

	for _, hasKey := range providers {
		if hasKey {
			providerCount++
		}
	}

	if providerCount == 0 {
		return CheckResult{
			Name:    "Providers",
			Status:  StatusError,
			Message: "No providers configured",
			Fix:     "Add at least one provider API key to config.json",
		}
	}

	return CheckResult{
		Name:    "Providers",
		Status:  StatusOK,
		Message: fmt.Sprintf("%d provider(s) configured, model: %s", providerCount, model),
	}
}

func checkPermissions(configPath string) CheckResult {
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	info, err := os.Stat(configPath)
	if err != nil {
		return CheckResult{
			Name:    "Permissions",
			Status:  StatusWarning,
			Message: "Cannot check config file permissions",
		}
	}

	mode := info.Mode()
	// Check if file is readable by others (not secure)
	if mode&0044 != 0 {
		return CheckResult{
			Name:    "Permissions",
			Status:  StatusWarning,
			Message: "Config file is readable by others (contains API keys)",
			Fix:     "Run 'chmod 600 ~/.KakoClaw/config.json'",
		}
	}

	return CheckResult{
		Name:    "Permissions",
		Status:  StatusOK,
		Message: "Config file permissions are secure",
	}
}

func checkDirectories() CheckResult {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return CheckResult{
			Name:    "Directories",
			Status:  StatusWarning,
			Message: "Cannot determine home directory",
		}
	}

	dirs := []string{
		filepath.Join(homeDir, ".KakoClaw"),
		filepath.Join(homeDir, ".KakoClaw", "workspace"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return CheckResult{
				Name:    "Directories",
				Status:  StatusError,
				Message: fmt.Sprintf("Required directory missing: %s", dir),
				Fix:     "Run 'KakoClaw onboard' to create required directories",
			}
		}
	}

	return CheckResult{
		Name:    "Directories",
		Status:  StatusOK,
		Message: "All required directories exist",
	}
}

func getDefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".KakoClaw", "config.json")
}

// PrintResults prints the check results in a formatted way
func PrintResults(results []CheckResult) {
	okCount := 0
	warningCount := 0
	errorCount := 0

	for _, result := range results {
		fmt.Printf("%s %s: %s\n", result.Status.Icon(), result.Name, result.Message)
		
		switch result.Status {
		case StatusOK:
			okCount++
		case StatusWarning:
			warningCount++
		case StatusError:
			errorCount++
		}

		if result.Fix != "" {
			fmt.Printf("   💡 Fix: %s\n", result.Fix)
		}
		fmt.Println()
	}

	fmt.Println("==================")
	fmt.Printf("✓ %d OK  |  ⚠ %d Warnings  |  ✗ %d Errors\n", okCount, warningCount, errorCount)
	fmt.Println()

	if errorCount > 0 {
		fmt.Println("❌ Some checks failed. Please fix the errors above.")
	} else if warningCount > 0 {
		fmt.Println("⚠️  All checks passed with warnings. Consider addressing the warnings.")
	} else {
		fmt.Println("✅ All checks passed! KakoClaw is ready to use.")
	}
}

// HasErrors returns true if any check resulted in an error
func HasErrors(results []CheckResult) bool {
	for _, result := range results {
		if result.Status == StatusError {
			return true
		}
	}
	return false
}
