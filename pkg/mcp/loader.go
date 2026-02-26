package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
)

// MCPServerSource represents where an MCP server config came from
type MCPServerSource string

const (
	SourceUserFolder   MCPServerSource = "user_folder"   // ~/.MakoClaw/users/<uuid>/mcp/
	SourceGlobalFolder MCPServerSource = "global_folder" // ~/.MakoClaw/mcp/
	SourceUserConfig   MCPServerSource = "user_config"   // ~/.MakoClaw/users/<uuid>/config.json
	SourceGlobalConfig MCPServerSource = "global_config" // ~/.MakoClaw/config.json
)

// MCPServerInfo represents a loaded MCP server configuration with its source
type MCPServerInfo struct {
	Name    string                  `json:"name"`
	Config  config.MCPServerConfig  `json:"config"`
	Source  MCPServerSource         `json:"source"`
	Path    string                  `json:"path"` // Path to the mcp.json file (for folder sources)
}

// Loader loads MCP server configurations from multiple sources
type Loader struct {
	userMCPPath   string // ~/.MakoClaw/users/<uuid>/mcp/
	globalMCPPath string // ~/.MakoClaw/mcp/
	dataDir       string // ~/.MakoClaw/
}

// NewLoader creates a new MCP loader
func NewLoader(dataDir string) *Loader {
	return &Loader{
		dataDir:       dataDir,
		globalMCPPath: filepath.Join(dataDir, "mcp"),
	}
}

// SetUserMCPPath sets the user-specific MCP directory
// Should be called with ~/.MakoClaw/users/<userUUID>/mcp
func (l *Loader) SetUserMCPPath(path string) {
	l.userMCPPath = path
}

// LoadAll loads MCP server configs from all sources with priority:
// 1. User folder (~/.MakoClaw/users/<uuid>/mcp/)
// 2. Global folder (~/.MakoClaw/mcp/)
// 3. User config.json
// 4. Global config.json
// Later sources don't override earlier ones (by name)
func (l *Loader) LoadAll(userUUID string) ([]MCPServerInfo, error) {
	servers := make([]MCPServerInfo, 0)
	seen := make(map[string]bool)

	// 1. User folder MCPs (highest priority)
	if l.userMCPPath != "" {
		folderServers := l.loadFromFolder(l.userMCPPath, SourceUserFolder)
		for _, s := range folderServers {
			if !seen[s.Name] {
				servers = append(servers, s)
				seen[s.Name] = true
			}
		}
	}

	// 2. Global folder MCPs
	if l.globalMCPPath != "" {
		folderServers := l.loadFromFolder(l.globalMCPPath, SourceGlobalFolder)
		for _, s := range folderServers {
			if !seen[s.Name] {
				servers = append(servers, s)
				seen[s.Name] = true
			}
		}
	}

	// 3. User config.json
	if userUUID != "" {
		configServers := l.loadFromConfig(userUUID, SourceUserConfig)
		for _, s := range configServers {
			if !seen[s.Name] {
				servers = append(servers, s)
				seen[s.Name] = true
			}
		}
	}

	// 4. Global config.json
	globalServers := l.loadFromGlobalConfig()
	for _, s := range globalServers {
		if !seen[s.Name] {
			servers = append(servers, s)
			seen[s.Name] = true
		}
	}

	return servers, nil
}

// loadFromFolder loads MCP configs from a folder structure:
// <basePath>/<server-name>/mcp.json
func (l *Loader) loadFromFolder(basePath string, source MCPServerSource) []MCPServerInfo {
	servers := make([]MCPServerInfo, 0)

	dirs, err := os.ReadDir(basePath)
	if err != nil {
		// Folder doesn't exist or can't be read - not an error
		return servers
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		serverName := dir.Name()
		mcpFile := filepath.Join(basePath, serverName, "mcp.json")

		cfg, err := l.loadMCPFile(mcpFile)
		if err != nil {
			logger.WarnCF("mcp", "Failed to load MCP config from folder", map[string]interface{}{
				"path":  mcpFile,
				"error": err.Error(),
			})
			continue
		}

		servers = append(servers, MCPServerInfo{
			Name:   serverName,
			Config: *cfg,
			Source: source,
			Path:   mcpFile,
		})

		logger.InfoCF("mcp", "Loaded MCP server from folder", map[string]interface{}{
			"name":   serverName,
			"source": source,
			"path":   mcpFile,
		})
	}

	return servers
}

// loadMCPFile loads a single mcp.json file
// Format compatible with Claude Desktop:
// {
//   "command": "npx",
//   "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path"],
//   "env": {}
// }
func (l *Loader) loadMCPFile(path string) (*config.MCPServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse the JSON - support both our format and Claude Desktop format
	var rawConfig struct {
		Enabled *bool             `json:"enabled"`
		Command string            `json:"command"`
		Args    []string          `json:"args"`
		Env     map[string]string `json:"env"`
	}

	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return nil, err
	}

	cfg := &config.MCPServerConfig{
		Enabled: true, // Default to enabled
		Command: rawConfig.Command,
		Args:    rawConfig.Args,
		Env:     rawConfig.Env,
	}

	// Override enabled if explicitly set
	if rawConfig.Enabled != nil {
		cfg.Enabled = *rawConfig.Enabled
	}

	return cfg, nil
}

// loadFromConfig loads MCP configs from user's config.json
func (l *Loader) loadFromConfig(userUUID string, source MCPServerSource) []MCPServerInfo {
	servers := make([]MCPServerInfo, 0)

	cfg, err := config.LoadConfigForUser(userUUID)
	if err != nil {
		return servers
	}

	if cfg.Tools.MCP.Servers == nil {
		return servers
	}

	for name, serverCfg := range cfg.Tools.MCP.Servers {
		servers = append(servers, MCPServerInfo{
			Name:   name,
			Config: serverCfg,
			Source: source,
			Path:   config.GetUserConfigPath(userUUID),
		})
	}

	return servers
}

// loadFromGlobalConfig loads MCP configs from global config.json
func (l *Loader) loadFromGlobalConfig() []MCPServerInfo {
	servers := make([]MCPServerInfo, 0)

	globalPath := config.GetGlobalConfigPath()
	cfg, err := config.LoadConfig(globalPath)
	if err != nil {
		return servers
	}

	if cfg.Tools.MCP.Servers == nil {
		return servers
	}

	for name, serverCfg := range cfg.Tools.MCP.Servers {
		servers = append(servers, MCPServerInfo{
			Name:   name,
			Config: serverCfg,
			Source: SourceGlobalConfig,
			Path:   globalPath,
		})
	}

	return servers
}

// SaveToFolder saves an MCP server config to a folder
// Creates <basePath>/<serverName>/mcp.json
func (l *Loader) SaveToFolder(basePath, serverName string, cfg config.MCPServerConfig) error {
	serverDir := filepath.Join(basePath, serverName)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return err
	}

	mcpFile := filepath.Join(serverDir, "mcp.json")

	// Prepare JSON content - use Claude Desktop compatible format
	content := map[string]interface{}{
		"command": cfg.Command,
		"args":    cfg.Args,
		"env":     cfg.Env,
	}

	// Only include enabled if false (true is default)
	if !cfg.Enabled {
		content["enabled"] = cfg.Enabled
	}

	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(mcpFile, data, 0644)
}

// DeleteFromFolder removes an MCP server folder
func (l *Loader) DeleteFromFolder(basePath, serverName string) error {
	serverDir := filepath.Join(basePath, serverName)
	return os.RemoveAll(serverDir)
}

// GetGlobalMCPPath returns the global MCP folder path
func (l *Loader) GetGlobalMCPPath() string {
	return l.globalMCPPath
}

// GetUserMCPPath returns the user MCP folder path
func (l *Loader) GetUserMCPPath() string {
	return l.userMCPPath
}

// EnsureMCPFolders creates the MCP folder structure if it doesn't exist
func (l *Loader) EnsureMCPFolders() error {
	// Create global MCP folder
	if l.globalMCPPath != "" {
		if err := os.MkdirAll(l.globalMCPPath, 0755); err != nil {
			return err
		}
	}

	// Create user MCP folder if set
	if l.userMCPPath != "" {
		if err := os.MkdirAll(l.userMCPPath, 0755); err != nil {
			return err
		}
	}

	return nil
}
