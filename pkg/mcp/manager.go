package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/logger"
	"github.com/sipeed/kakoclaw/pkg/tools"
)

// Manager manages multiple MCP server connections and their tools
type Manager struct {
	cfg     config.MCPConfig
	clients map[string]*Client
	mu      sync.RWMutex
}

// NewManager creates a new MCP manager from config
func NewManager(cfg config.MCPConfig) *Manager {
	return &Manager{
		cfg:     cfg,
		clients: make(map[string]*Client),
	}
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
		info := ServerInfo{
			Name:      name,
			Command:   client.command,
			Connected: client.Connected(),
			LastError: client.LastError(),
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
	Name          string   `json:"name"`
	Command       string   `json:"command"`
	Connected     bool     `json:"connected"`
	LastError     string   `json:"last_error,omitempty"`
	ServerName    string   `json:"server_name,omitempty"`
	ServerVersion string   `json:"server_version,omitempty"`
	ToolCount     int      `json:"tool_count"`
	Tools         []string `json:"tools,omitempty"`
}
