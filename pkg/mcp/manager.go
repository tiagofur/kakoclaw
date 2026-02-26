package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/tools"
)

// Manager manages multiple MCP server connections and their tools
type Manager struct {
	cfg     config.MCPConfig
	clients map[string]*Client
	sources map[string]MCPServerInfo // Tracks where each server config came from
	mu      sync.RWMutex
}

// NewManager creates a new MCP manager from config
func NewManager(cfg config.MCPConfig) *Manager {
	return &Manager{
		cfg:     cfg,
		clients: make(map[string]*Client),
		sources: make(map[string]MCPServerInfo),
	}
}

// NewManagerFromLoader creates a new MCP manager using the loader
func NewManagerFromLoader(loader *Loader, userUUID string) (*Manager, error) {
	servers, err := loader.LoadAll(userUUID)
	if err != nil {
		return nil, err
	}

	cfg := config.MCPConfig{
		Servers: make(map[string]config.MCPServerConfig),
	}
	sources := make(map[string]MCPServerInfo)

	for _, s := range servers {
		cfg.Servers[s.Name] = s.Config
		sources[s.Name] = s
	}

	return &Manager{
		cfg:     cfg,
		clients: make(map[string]*Client),
		sources: sources,
	}, nil
}

// Start connects to all configured MCP servers and discovers their tools.
// Non-fatal: servers that fail to connect are logged and skipped.
func (m *Manager) Start(ctx context.Context) {
	for name, serverCfg := range m.cfg.Servers {
		if !serverCfg.Enabled {
			logger.InfoCF("mcp", "MCP server disabled, skipping", map[string]interface{}{"name": name})
			continue
		}

		env := make([]string, 0, len(serverCfg.Env))
		for k, v := range serverCfg.Env {
			env = append(env, k+"="+v)
		}

		client := NewClient(name, serverCfg.Command, serverCfg.Args, env)

		m.mu.Lock()
		m.clients[name] = client
		m.mu.Unlock()

		if err := client.Connect(ctx); err != nil {
			logger.WarnCF("mcp", "Failed to connect to MCP server (will be available for manual reconnect)", map[string]interface{}{
				"name":  name,
				"error": err.Error(),
			})
		}
	}
}

// Stop disconnects from all MCP servers
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, client := range m.clients {
		client.Close()
	}
}

// Reconnect tries to reconnect a specific MCP server
func (m *Manager) Reconnect(ctx context.Context, name string) error {
	m.mu.RLock()
	client, ok := m.clients[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("MCP server %q not found in config", name)
	}

	// Close existing connection if any
	client.Close()

	// Reconnect
	return client.Connect(ctx)
}

// GetTools returns all tools from all connected MCP servers, wrapped as tools.Tool
func (m *Manager) GetTools() []tools.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []tools.Tool
	for _, client := range m.clients {
		if !client.Connected() {
			continue
		}
		for _, toolInfo := range client.Tools() {
			result = append(result, NewMCPTool(client, toolInfo))
		}
	}
	return result
}

// ServerStatus returns the status of all configured MCP servers
func (m *Manager) ServerStatus() []ServerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var servers []ServerInfo
	for name, client := range m.clients {
		// Get config for this server
		serverCfg, hasCfg := m.cfg.Servers[name]

		info := ServerInfo{
			Name:      name,
			Command:   client.command,
			Connected: client.Connected(),
			LastError: client.LastError(),
		}

		// Add config details if available
		if hasCfg {
			info.Args = serverCfg.Args
			info.Env = serverCfg.Env
			info.Enabled = serverCfg.Enabled
		}

		// Add source information if tracked
		if sourceInfo, hasSource := m.sources[name]; hasSource {
			info.Source = sourceInfo.Source
			info.Path = sourceInfo.Path
		}

		if client.Connected() {
			si := client.ServerInfo()
			info.ServerName = si.ServerInfo.Name
			info.ServerVersion = si.ServerInfo.Version
			info.ToolCount = len(client.Tools())
			toolNames := make([]string, 0, len(client.Tools()))
			for _, t := range client.Tools() {
				toolNames = append(toolNames, t.Name)
			}
			info.Tools = toolNames
		}
		servers = append(servers, info)
	}
	return servers
}

// ServerInfo represents the status of an MCP server for API responses
type ServerInfo struct {
	Name          string          `json:"name"`
	Command       string          `json:"command"`
	Args          []string        `json:"args,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Enabled       bool            `json:"enabled"`
	Connected     bool            `json:"connected"`
	Source        MCPServerSource `json:"source,omitempty"`
	Path          string          `json:"path,omitempty"`
	LastError     string          `json:"last_error,omitempty"`
	ServerName    string          `json:"server_name,omitempty"`
	ServerVersion string          `json:"server_version,omitempty"`
	ToolCount     int             `json:"tool_count"`
	Tools         []string        `json:"tools,omitempty"`
}

// AddServer adds a new MCP server to the manager and optionally connects to it.
// This does NOT persist the configuration - caller should save config separately.
func (m *Manager) AddServer(ctx context.Context, name string, serverCfg config.MCPServerConfig, autoConnect bool) error {
	return m.AddServerWithSource(ctx, name, serverCfg, autoConnect, MCPServerSource(""), "")
}

// AddServerWithSource adds a new MCP server with source tracking
func (m *Manager) AddServerWithSource(ctx context.Context, name string, serverCfg config.MCPServerConfig, autoConnect bool, source MCPServerSource, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if server already exists
	if _, exists := m.clients[name]; exists {
		return fmt.Errorf("MCP server %q already exists", name)
	}

	// Update internal config
	if m.cfg.Servers == nil {
		m.cfg.Servers = make(map[string]config.MCPServerConfig)
	}
	m.cfg.Servers[name] = serverCfg

	// Track source if provided
	if source != "" {
		if m.sources == nil {
			m.sources = make(map[string]MCPServerInfo)
		}
		m.sources[name] = MCPServerInfo{
			Name:   name,
			Config: serverCfg,
			Source: source,
			Path:   path,
		}
	}

	// Create client
	env := make([]string, 0, len(serverCfg.Env))
	for k, v := range serverCfg.Env {
		env = append(env, k+"="+v)
	}
	client := NewClient(name, serverCfg.Command, serverCfg.Args, env)
	m.clients[name] = client

	// Optionally connect
	if autoConnect && serverCfg.Enabled {
		// Unlock while connecting to avoid holding lock during I/O
		m.mu.Unlock()
		err := client.Connect(ctx)
		m.mu.Lock()
		if err != nil {
			logger.WarnCF("mcp", "Failed to connect to new MCP server", map[string]interface{}{
				"name":  name,
				"error": err.Error(),
			})
			// Don't return error - server is added but not connected
		}
	}

	return nil
}

// RemoveServer stops and removes an MCP server from the manager.
// This does NOT persist the configuration - caller should save config separately.
func (m *Manager) RemoveServer(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[name]
	if !exists {
		return fmt.Errorf("MCP server %q not found", name)
	}

	// Close the client
	client.Close()

	// Remove from internal state
	delete(m.clients, name)
	delete(m.cfg.Servers, name)

	return nil
}

// UpdateServer updates an existing MCP server's configuration.
// If the command or args changed and the server is running, it will be reconnected.
// This does NOT persist the configuration - caller should save config separately.
func (m *Manager) UpdateServer(ctx context.Context, name string, serverCfg config.MCPServerConfig) error {
	m.mu.Lock()
	client, exists := m.clients[name]
	oldCfg, hadCfg := m.cfg.Servers[name]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("MCP server %q not found", name)
	}

	// Check if command/args changed
	commandChanged := !hadCfg || oldCfg.Command != serverCfg.Command || !sliceEqual(oldCfg.Args, serverCfg.Args)
	wasConnected := client.Connected()

	// Update config
	m.mu.Lock()
	m.cfg.Servers[name] = serverCfg
	m.mu.Unlock()

	// If command/args changed and server was running, we need to restart it
	if commandChanged && wasConnected {
		// Close old client
		client.Close()

		// Create new client with updated config
		env := make([]string, 0, len(serverCfg.Env))
		for k, v := range serverCfg.Env {
			env = append(env, k+"="+v)
		}
		newClient := NewClient(name, serverCfg.Command, serverCfg.Args, env)

		m.mu.Lock()
		m.clients[name] = newClient
		m.mu.Unlock()

		// Connect if enabled
		if serverCfg.Enabled {
			if err := newClient.Connect(ctx); err != nil {
				return fmt.Errorf("failed to reconnect after update: %w", err)
			}
		}
	} else if serverCfg.Enabled && !wasConnected {
		// Server wasn't connected but is now enabled - try to connect
		if err := client.Connect(ctx); err != nil {
			logger.WarnCF("mcp", "Failed to connect MCP server after enabling", map[string]interface{}{
				"name":  name,
				"error": err.Error(),
			})
		}
	} else if !serverCfg.Enabled && wasConnected {
		// Server is being disabled - close connection
		client.Close()
	}

	return nil
}

// GetServerConfig returns the configuration for a specific server
func (m *Manager) GetServerConfig(name string) (config.MCPServerConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cfg, ok := m.cfg.Servers[name]
	return cfg, ok
}

// HasServer checks if a server with the given name exists
func (m *Manager) HasServer(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.clients[name]
	return exists
}

// sliceEqual compares two string slices for equality
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
