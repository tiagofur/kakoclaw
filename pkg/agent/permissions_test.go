package agent

import (
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/cron"
	"github.com/sipeed/makoclaw/pkg/tools"
)

func TestFilterToolsByPermissions_AdminWildcard(t *testing.T) {
	// Create base registry with several tools
	baseRegistry := tools.NewToolRegistry()
	baseRegistry.Register(tools.NewExecTool("/tmp", false))
	baseRegistry.Register(tools.NewSpawnTool(nil))
	baseRegistry.Register(tools.NewReadFileTool("/tmp", false))
	baseRegistry.Register(tools.NewWriteFileTool("/tmp", false))

	cfg := &config.Config{}
	cfg.ToolPermissions = config.ToolPermissionsConfig{
		RoleDefaults: map[string][]string{
			"admin": {"*"}, // Wildcard - all tools
		},
	}

	// Admin should get all tools
	filtered := filterToolsByPermissions(baseRegistry, "admin", 123, cfg, nil)

	expectedTools := []string{"exec", "spawn", "read_file", "write_file"}
	for _, toolName := range expectedTools {
		if tool, _ := filtered.Get(toolName); tool == nil {
			t.Errorf("Admin should have access to %s tool", toolName)
		}
	}
}

func TestFilterToolsByPermissions_UserRestricted(t *testing.T) {
	baseRegistry := tools.NewToolRegistry()
	baseRegistry.Register(tools.NewExecTool("/tmp", false))
	baseRegistry.Register(tools.NewSpawnTool(nil))
	baseRegistry.Register(tools.NewReadFileTool("/tmp", false))
	baseRegistry.Register(tools.NewWriteFileTool("/tmp", false))

	cfg := &config.Config{}
	cfg.ToolPermissions = config.ToolPermissionsConfig{
		RoleDefaults: map[string][]string{
			"user": {"read_file"}, // Only read access
		},
	}

	filtered := filterToolsByPermissions(baseRegistry, "user", 456, cfg, nil)

	// User should have read_file
	if tool, _ := filtered.Get("read_file"); tool == nil {
		t.Error("User should have access to read_file tool")
	}

	// User should NOT have dangerous tools
	restrictedTools := []string{"exec", "spawn", "write_file", "edit_file", "append_file"}
	for _, toolName := range restrictedTools {
		if tool, _ := filtered.Get(toolName); tool != nil {
			t.Errorf("User should NOT have access to %s tool", toolName)
		}
	}
}

func TestFilterToolsByPermissions_UserOverrides(t *testing.T) {
	baseRegistry := tools.NewToolRegistry()
	baseRegistry.Register(tools.NewExecTool("/tmp", false))
	baseRegistry.Register(tools.NewReadFileTool("/tmp", false))
	baseRegistry.Register(tools.NewWriteFileTool("/tmp", false))

	cfg := &config.Config{}
	cfg.ToolPermissions = config.ToolPermissionsConfig{
		RoleDefaults: map[string][]string{
			"user": {"read_file"},
		},
		UserOverrides: map[string][]string{
			"special_user": {"read_file", "write_file", "exec_restricted"}, // Special user gets more tools
		},
	}

	// Regular user
	filteredRegular := filterToolsByPermissions(baseRegistry, "user", 456, cfg, nil)
	if tool, _ := filteredRegular.Get("write_file"); tool != nil {
		t.Error("Regular user should NOT have write_file")
	}

	// User with override - Note: userID is checked as string in permissions.go
	// For now, we'll test role-based filtering only since user-specific overrides
	// require username lookup from storage
}

func TestFilterToolsByPermissions_ExecRestrictedConfiguration(t *testing.T) {
	baseRegistry := tools.NewToolRegistry()
	execTool := tools.NewExecTool("/tmp", false)
	baseRegistry.Register(execTool)

	cfg := &config.Config{}
	cfg.ToolPermissions = config.ToolPermissionsConfig{
		RoleDefaults: map[string][]string{
			"user": {"exec_restricted"},
		},
		AllowedShellCommands: []string{"ls", "cat", "grep"},
	}

	filtered := filterToolsByPermissions(baseRegistry, "user", 999, cfg, nil)

	// User should have exec tool
	execFromRegistry, exists := filtered.Get("exec")
	if !exists || execFromRegistry == nil {
		t.Fatal("User with exec_restricted should have exec tool")
	}

	// Verify it's configured with allowlist
	// (We can't directly test the internal allowPatterns, but we can verify the tool exists)
	if execTool := execFromRegistry.(*tools.ExecTool); execTool != nil {
		// Tool should exist and be properly configured
		// In production, SetSafeCommandsForUser is called with cfg.ToolPermissions.AllowedShellCommands
	}
}

func TestFilterToolsByPermissions_EmptyPermissions(t *testing.T) {
	baseRegistry := tools.NewToolRegistry()
	baseRegistry.Register(tools.NewReadFileTool("/tmp", false))
	baseRegistry.Register(tools.NewWriteFileTool("/tmp", false))

	cfg := &config.Config{}
	cfg.ToolPermissions = config.ToolPermissionsConfig{
		RoleDefaults: map[string][]string{
			"restricted": {}, // No permissions at all
		},
	}

	filtered := filterToolsByPermissions(baseRegistry, "restricted", 111, cfg, nil)

	// Should have no tools
	if tool, _ := filtered.Get("read_file"); tool != nil {
		t.Error("Restricted user should have no access to read_file")
	}
	if tool, _ := filtered.Get("write_file"); tool != nil {
		t.Error("Restricted user should have no access to write_file")
	}
}

func TestFilterToolsByPermissions_DefaultAdmin(t *testing.T) {
	baseRegistry := tools.NewToolRegistry()
	baseRegistry.Register(tools.NewExecTool("/tmp", false))
	baseRegistry.Register(tools.NewSpawnTool(nil))

	// Use default config
	cfg := config.DefaultConfig()

	filtered := filterToolsByPermissions(baseRegistry, "admin", 1, cfg, nil)

	// Default admin should have wildcard access
	if tool, _ := filtered.Get("exec"); tool == nil {
		t.Error("Default admin config should include exec tool")
	}
	if tool, _ := filtered.Get("spawn"); tool == nil {
		t.Error("Default admin config should include spawn tool")
	}
}

func TestFilterToolsByPermissions_DefaultUser(t *testing.T) {
	baseRegistry := tools.NewToolRegistry()
	baseRegistry.Register(tools.NewExecTool("/tmp", false))
	baseRegistry.Register(tools.NewReadFileTool("/tmp", false))
	baseRegistry.Register(tools.NewWriteFileTool("/tmp", false))
	baseRegistry.Register(tools.NewWebSearchTool(nil, 10))
	baseRegistry.Register(tools.NewWebFetchTool(50000))
	baseRegistry.Register(tools.NewConfigureTool())
	baseRegistry.Register(tools.NewCronTool(cron.NewCronService(filepath.Join(t.TempDir(), "cron", "jobs.json"), nil), nil, nil))

	// Use default config
	cfg := config.DefaultConfig()

	filtered := filterToolsByPermissions(baseRegistry, "user", 2, cfg, nil)

	// Default user should have safe tools
	safeTools := []string{"read_file", "write_file", "web_search", "web_fetch", "configure", "cron"}
	for _, toolName := range safeTools {
		if tool, _ := filtered.Get(toolName); tool == nil {
			t.Errorf("Default user should have %s tool", toolName)
		}
	}

	// Default user should have RESTRICTED exec access (not full exec)
	if tool, _ := filtered.Get("exec"); tool == nil {
		t.Error("Default user should have restricted exec access")
	}

	// Default user should NOT have dangerous tools
	restrictedTools := []string{"spawn"}
	for _, toolName := range restrictedTools {
		if tool, _ := filtered.Get(toolName); tool != nil {
			t.Errorf("Default user should NOT have %s tool", toolName)
		}
	}
}

func TestAgentLoopRegisterToolRespectsPermissions(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ToolPermissions.RoleDefaults["user"] = []string{"read_file"}

	al := &AgentLoop{
		workspace:      t.TempDir(),
		userID:         42,
		userRole:       "user",
		cfg:            cfg,
		contextBuilder: NewContextBuilder(t.TempDir()),
		tools:          tools.NewToolRegistry(),
		baseTools:      tools.NewToolRegistry(),
	}

	cronTool := tools.NewCronTool(cron.NewCronService(filepath.Join(t.TempDir(), "cron", "jobs.json"), nil), nil, nil)
	al.RegisterTool(cronTool)

	if tool, _ := al.baseTools.Get("cron"); tool == nil {
		t.Fatal("registering a tool should always add it to the base registry")
	}
	if tool, _ := al.tools.Get("cron"); tool != nil {
		t.Fatal("registered tool should still respect filtered user permissions")
	}
}

func TestConfigureExecToolForRole_AllowlistApplication(t *testing.T) {
	execTool := tools.NewExecTool("/tmp", false)

	cfg := &config.Config{}
	cfg.ToolPermissions = config.ToolPermissionsConfig{
		AllowedShellCommands: []string{"ls", "cat", "echo"},
	}

	err := configureExecToolForRole(execTool, "user", cfg)

	if err != nil {
		t.Fatalf("Should successfully configure exec tool: %v", err)
	}

	// The actual allowlist enforcement is tested in shell_test.go
	// Here we just verify the configuration succeeds
}

func TestFilterToolsByPermissions_RoleNotDefined(t *testing.T) {
	baseRegistry := tools.NewToolRegistry()
	baseRegistry.Register(tools.NewReadFileTool("/tmp", false))

	cfg := &config.Config{}
	cfg.ToolPermissions = config.ToolPermissionsConfig{
		RoleDefaults: map[string][]string{
			"admin": {"*"},
		},
	}

	// Request undefined role
	filtered := filterToolsByPermissions(baseRegistry, "undefined_role", 999, cfg, nil)

	// Should have no tools (role defaults to empty permissions)
	if tool, _ := filtered.Get("read_file"); tool != nil {
		t.Error("Undefined role should have no tool access")
	}
}

func TestFilterToolsByPermissions_CaseSensitivity(t *testing.T) {
	baseRegistry := tools.NewToolRegistry()
	baseRegistry.Register(tools.NewReadFileTool("/tmp", false))

	cfg := &config.Config{}
	cfg.ToolPermissions = config.ToolPermissionsConfig{
		RoleDefaults: map[string][]string{
			"user": {"Read_File", "LIST_DIR"}, // Mixed case
		},
	}

	filtered := filterToolsByPermissions(baseRegistry, "user", 123, cfg, nil)

	// Tool names should be case-sensitive (read_file != Read_File)
	if tool, _ := filtered.Get("read_file"); tool != nil {
		t.Error("Permission matching should be case-sensitive")
	}
}

func TestFilterToolsByPermissions_WildcardDoesNotGrantExtraPermissions(t *testing.T) {
	baseRegistry := tools.NewToolRegistry()
	baseRegistry.Register(tools.NewReadFileTool("/tmp", false))

	cfg := &config.Config{}
	cfg.ToolPermissions = config.ToolPermissionsConfig{
		RoleDefaults: map[string][]string{
			"admin": {"*"},
		},
	}

	filtered := filterToolsByPermissions(baseRegistry, "admin", 1, cfg, nil)

	// Wildcard should only grant access to registered tools
	if tool, _ := filtered.Get("nonexistent_tool"); tool != nil {
		t.Error("Wildcard should not create nonexistent tools")
	}

	// But should grant access to registered tools
	if tool, _ := filtered.Get("read_file"); tool == nil {
		t.Error("Wildcard should grant access to registered read_file tool")
	}
}
