package agent

import (
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
	"github.com/sipeed/makoclaw/pkg/tools"
)

// filterToolsByPermissions creates a filtered ToolRegistry based on user role and permissions.
// Returns a new registry containing only tools the user is allowed to access.
func filterToolsByPermissions(
	baseRegistry *tools.ToolRegistry,
	userRole string,
	userID int64,
	cfg *config.Config,
	centralStore *storage.CentralStorage,
) *tools.ToolRegistry {
	// Get role default permissions
	rolePermissions, ok := cfg.ToolPermissions.RoleDefaults[userRole]
	if !ok {
		logger.WarnCF("agent", "Unknown role, denying all tools", map[string]interface{}{
			"role":    userRole,
			"user_id": userID,
		})
		return tools.NewToolRegistry() // Empty registry
	}

	// Get effective permissions (user overrides or role defaults)
	effectivePermissions := rolePermissions
	if userID > 0 && centralStore != nil {
		user, err := centralStore.GetUserByID(userID)
		if err == nil {
			effectivePermissions = user.GetEffectiveToolPermissions(rolePermissions)
		}
	}

	// Check if user has wildcard access (admin)
	hasWildcard := false
	for _, perm := range effectivePermissions {
		if perm == "*" {
			hasWildcard = true
			break
		}
	}

	// If wildcard, return base registry unchanged (but still configure exec allowlist for non-admins)
	if hasWildcard {
		logger.InfoCF("agent", "User has full tool access", map[string]interface{}{
			"role":    userRole,
			"user_id": userID,
		})
		// Admin gets full exec access - no allowlist
		return baseRegistry
	}

	// Build permission map for quick lookup
	allowedTools := make(map[string]bool)
	hasRestrictedExec := false
	for _, tool := range effectivePermissions {
		if tool == "exec_restricted" {
			hasRestrictedExec = true
			allowedTools["exec"] = true // Allow exec but will configure allowlist
		} else {
			allowedTools[tool] = true
		}
	}

	// Create filtered registry
	filtered := tools.NewToolRegistry()
	allTools := baseRegistry.List()

	for _, toolName := range allTools {
		if allowedTools[toolName] {
			if tool, ok := baseRegistry.Get(toolName); ok {
				// Special handling for exec tool with restricted access
				if toolName == "exec" && hasRestrictedExec {
					if execTool, ok := tool.(*tools.ExecTool); ok {
						// Configure safe command allowlist
						commands := cfg.ToolPermissions.AllowedShellCommands
						if len(commands) == 0 {
							commands = tools.SafeShellCommands
						}
						if err := execTool.SetSafeCommandsForUser(commands); err != nil {
							// FAIL-CLOSED: Do not register exec tool if allowlist setup fails
							logger.ErrorCF("agent", "Failed to set shell allowlist - exec tool NOT registered (fail-closed)", map[string]interface{}{
								"error":   err.Error(),
								"user_id": userID,
							})
							continue // Skip registering this tool
						}
						logger.InfoCF("agent", "Configured restricted shell access", map[string]interface{}{
							"user_id":          userID,
							"allowed_commands": len(commands),
						})
					}
				}
				filtered.Register(tool)
			}
		}
	}

	allowedList := filtered.List()
	logger.InfoCF("agent", "Tools filtered by permissions", map[string]interface{}{
		"role":          userRole,
		"user_id":       userID,
		"allowed_count": len(allowedList),
		"allowed_tools": allowedList,
	})

	return filtered
}

// configureExecToolForRole applies shell command restrictions to an exec tool based on user role
func configureExecToolForRole(execTool *tools.ExecTool, userRole string, cfg *config.Config) error {
	// Admins get unrestricted access (only blacklist applies)
	if userRole == "admin" {
		return nil
	}

	// Non-admins get allowlist of safe commands
	commands := cfg.ToolPermissions.AllowedShellCommands
	if len(commands) == 0 {
		commands = tools.SafeShellCommands
	}

	return execTool.SetSafeCommandsForUser(commands)
}
