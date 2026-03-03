package mcp

import (
	"context"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
)

// ---------------------------------------------------------------------------
// NewManager
// ---------------------------------------------------------------------------

func TestNewManager_EmptyConfig(t *testing.T) {
	cfg := config.MCPConfig{}
	mgr := NewManager(cfg)
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestNewManager_WithServers(t *testing.T) {
	cfg := config.MCPConfig{
		Servers: map[string]config.MCPServerConfig{
			"test-server": {
				Enabled: true,
				Command: "echo",
				Args:    []string{"hello"},
			},
		},
	}
	mgr := NewManager(cfg)
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
}

// ---------------------------------------------------------------------------
// HasServer
// ---------------------------------------------------------------------------

func TestManager_HasServer_BeforeStart(t *testing.T) {
	cfg := config.MCPConfig{
		Servers: map[string]config.MCPServerConfig{
			"myserver": {Enabled: true, Command: "echo"},
		},
	}
	mgr := NewManager(cfg)

	// HasServer checks the clients map, which is populated by Start(), not NewManager()
	if mgr.HasServer("myserver") {
		t.Error("HasServer should return false before Start()")
	}
}

// ---------------------------------------------------------------------------
// AddServer (without auto-connect)
// ---------------------------------------------------------------------------

func TestManager_AddServer_NoAutoConnect(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})

	serverCfg := config.MCPServerConfig{
		Enabled: true,
		Command: "echo",
		Args:    []string{"test"},
		Env:     map[string]string{"KEY": "value"},
	}

	err := mgr.AddServer(context.Background(), "new-server", serverCfg, false)
	if err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	// Verify server is now registered
	if !mgr.HasServer("new-server") {
		t.Error("HasServer should return true after AddServer")
	}

	// Verify config is stored
	got, ok := mgr.GetServerConfig("new-server")
	if !ok {
		t.Fatal("GetServerConfig should return true for added server")
	}
	if got.Command != "echo" {
		t.Errorf("Command = %q; want %q", got.Command, "echo")
	}
	if len(got.Args) != 1 || got.Args[0] != "test" {
		t.Errorf("Args = %v; want [test]", got.Args)
	}
	if got.Env["KEY"] != "value" {
		t.Errorf("Env[KEY] = %q; want %q", got.Env["KEY"], "value")
	}
}

func TestManager_AddServer_Duplicate(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})

	serverCfg := config.MCPServerConfig{
		Enabled: true,
		Command: "echo",
	}

	err := mgr.AddServer(context.Background(), "dup", serverCfg, false)
	if err != nil {
		t.Fatalf("first AddServer failed: %v", err)
	}

	err = mgr.AddServer(context.Background(), "dup", serverCfg, false)
	if err == nil {
		t.Error("adding duplicate server should return error")
	}
}

// ---------------------------------------------------------------------------
// AddServerWithSource
// ---------------------------------------------------------------------------

func TestManager_AddServerWithSource(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})

	serverCfg := config.MCPServerConfig{
		Enabled: true,
		Command: "node",
		Args:    []string{"server.js"},
	}

	err := mgr.AddServerWithSource(
		context.Background(),
		"sourced-server",
		serverCfg,
		false,
		SourceGlobalFolder,
		"/path/to/mcp",
	)
	if err != nil {
		t.Fatalf("AddServerWithSource failed: %v", err)
	}

	if !mgr.HasServer("sourced-server") {
		t.Error("HasServer should return true after AddServerWithSource")
	}

	// Verify source tracking via ServerStatus
	statuses := mgr.ServerStatus()
	if len(statuses) != 1 {
		t.Fatalf("ServerStatus returned %d entries; want 1", len(statuses))
	}
	if statuses[0].Source != SourceGlobalFolder {
		t.Errorf("Source = %q; want %q", statuses[0].Source, SourceGlobalFolder)
	}
	if statuses[0].Path != "/path/to/mcp" {
		t.Errorf("Path = %q; want %q", statuses[0].Path, "/path/to/mcp")
	}
}

// ---------------------------------------------------------------------------
// RemoveServer
// ---------------------------------------------------------------------------

func TestManager_RemoveServer(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})

	serverCfg := config.MCPServerConfig{
		Enabled: true,
		Command: "echo",
	}

	err := mgr.AddServer(context.Background(), "to-remove", serverCfg, false)
	if err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	err = mgr.RemoveServer("to-remove")
	if err != nil {
		t.Fatalf("RemoveServer failed: %v", err)
	}

	if mgr.HasServer("to-remove") {
		t.Error("HasServer should return false after RemoveServer")
	}
}

func TestManager_RemoveServer_NotFound(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})

	err := mgr.RemoveServer("nonexistent")
	if err == nil {
		t.Error("RemoveServer on nonexistent server should return error")
	}
}

// ---------------------------------------------------------------------------
// GetServerConfig
// ---------------------------------------------------------------------------

func TestManager_GetServerConfig_NotFound(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})

	_, ok := mgr.GetServerConfig("nonexistent")
	if ok {
		t.Error("GetServerConfig should return false for nonexistent server")
	}
}

func TestManager_GetServerConfig_FromInitialConfig(t *testing.T) {
	cfg := config.MCPConfig{
		Servers: map[string]config.MCPServerConfig{
			"initial": {
				Enabled: true,
				Command: "python",
				Args:    []string{"serve.py"},
			},
		},
	}
	mgr := NewManager(cfg)

	got, ok := mgr.GetServerConfig("initial")
	if !ok {
		t.Fatal("GetServerConfig should return true for config-provided server")
	}
	if got.Command != "python" {
		t.Errorf("Command = %q; want %q", got.Command, "python")
	}
}

// ---------------------------------------------------------------------------
// GetTools on empty/disconnected manager
// ---------------------------------------------------------------------------

func TestManager_GetTools_Empty(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})

	tools := mgr.GetTools()
	if len(tools) != 0 {
		t.Errorf("GetTools on empty manager returned %d tools; want 0", len(tools))
	}
}

func TestManager_GetTools_DisconnectedServers(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})

	// Add server without connecting
	serverCfg := config.MCPServerConfig{
		Enabled: true,
		Command: "echo",
	}
	_ = mgr.AddServer(context.Background(), "unconnected", serverCfg, false)

	tools := mgr.GetTools()
	if len(tools) != 0 {
		t.Errorf("GetTools with disconnected server returned %d tools; want 0", len(tools))
	}
}

// ---------------------------------------------------------------------------
// ServerStatus
// ---------------------------------------------------------------------------

func TestManager_ServerStatus_Empty(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})

	statuses := mgr.ServerStatus()
	if len(statuses) != 0 {
		t.Errorf("ServerStatus on empty manager returned %d; want 0", len(statuses))
	}
}

func TestManager_ServerStatus_WithServers(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})

	cfg1 := config.MCPServerConfig{Enabled: true, Command: "echo", Args: []string{"-n"}}
	cfg2 := config.MCPServerConfig{Enabled: false, Command: "node"}

	_ = mgr.AddServer(context.Background(), "server1", cfg1, false)
	_ = mgr.AddServer(context.Background(), "server2", cfg2, false)

	statuses := mgr.ServerStatus()
	if len(statuses) != 2 {
		t.Fatalf("ServerStatus returned %d entries; want 2", len(statuses))
	}

	// Build a map for easier lookup (order is non-deterministic)
	statusMap := make(map[string]ServerInfo)
	for _, s := range statuses {
		statusMap[s.Name] = s
	}

	s1, ok := statusMap["server1"]
	if !ok {
		t.Fatal("server1 not found in ServerStatus")
	}
	if s1.Command != "echo" {
		t.Errorf("server1 Command = %q; want %q", s1.Command, "echo")
	}
	if s1.Connected {
		t.Error("server1 should not be connected (never called Start)")
	}
	if !s1.Enabled {
		t.Error("server1 Enabled should be true")
	}

	s2, ok := statusMap["server2"]
	if !ok {
		t.Fatal("server2 not found in ServerStatus")
	}
	if s2.Enabled {
		t.Error("server2 Enabled should be false")
	}
}

// ---------------------------------------------------------------------------
// sliceEqual (internal helper)
// ---------------------------------------------------------------------------

func TestSliceEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
		want bool
	}{
		{"both empty", []string{}, []string{}, true},
		{"both nil", nil, nil, true},
		{"equal", []string{"a", "b"}, []string{"a", "b"}, true},
		{"different length", []string{"a"}, []string{"a", "b"}, false},
		{"different content", []string{"a", "b"}, []string{"a", "c"}, false},
		{"nil vs empty", nil, []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sliceEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("sliceEqual(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Reconnect on nonexistent server
// ---------------------------------------------------------------------------

func TestManager_Reconnect_NotFound(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})

	err := mgr.Reconnect(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Reconnect on nonexistent server should return error")
	}
}

// ---------------------------------------------------------------------------
// Stop on empty manager
// ---------------------------------------------------------------------------

func TestManager_Stop_Empty(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})
	// Should not panic
	mgr.Stop()
}

// ---------------------------------------------------------------------------
// UpdateServer on nonexistent server
// ---------------------------------------------------------------------------

func TestManager_UpdateServer_NotFound(t *testing.T) {
	mgr := NewManager(config.MCPConfig{})

	err := mgr.UpdateServer(context.Background(), "nonexistent", config.MCPServerConfig{})
	if err == nil {
		t.Error("UpdateServer on nonexistent server should return error")
	}
}
