package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
	"github.com/sipeed/makoclaw/pkg/tools"
)

// handleToolPermissions returns the tool permission matrix for roles
func (s *Server) handleToolPermissions(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil || claims.Role != "admin" {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet {
		// Return current tool permissions configuration
		s.fullConfig.RLock()
		permissions := s.fullConfig.ToolPermissions
		s.fullConfig.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"role_defaults":          permissions.RoleDefaults,
			"allowed_shell_commands": permissions.AllowedShellCommands,
			"user_overrides":         permissions.UserOverrides,
		})
		return
	}

	if r.Method == http.MethodPut {
		// Update tool permissions configuration
		var in struct {
			RoleDefaults         map[string][]string `json:"role_defaults"`
			AllowedShellCommands []string            `json:"allowed_shell_commands"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		// Validate that admin role always has full access
		if adminPerms, ok := in.RoleDefaults["admin"]; ok {
			hasWildcard := false
			for _, perm := range adminPerms {
				if perm == "*" {
					hasWildcard = true
					break
				}
			}
			if !hasWildcard {
				http.Error(w, "admin role must have wildcard (*) permission", http.StatusBadRequest)
				return
			}
		}

		// Update configuration
		s.fullConfig.Lock()
		if in.RoleDefaults != nil {
			s.fullConfig.ToolPermissions.RoleDefaults = in.RoleDefaults
		}
		if in.AllowedShellCommands != nil {
			s.fullConfig.ToolPermissions.AllowedShellCommands = in.AllowedShellCommands
		}
		s.fullConfig.Unlock()

		// TODO: Save configuration to disk (needs Save() method implementation)
		logger.InfoCF("web", "Tool permissions updated (config not persisted yet)", map[string]interface{}{
			"admin": claims.Sub,
		})

		logger.InfoCF("web", "Tool permissions updated", map[string]interface{}{
			"admin": claims.Sub,
		})

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// handleUserTools returns or updates the tool permissions for a specific user
func (s *Server) handleUserTools(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract user ID from path
	parts := splitPath(r.URL.Path)
	if len(parts) < 4 || parts[3] == "" {
		http.Error(w, "user ID required", http.StatusBadRequest)
		return
	}

	targetUserID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	// Determine which storage to use for user lookup
	var store *storage.Storage
	if s.centralStore != nil {
		// In central storage mode, we need to query via GetUserByID
		// CentralStorage doesn't implement Storage interface, so we handle separately
		targetUser, err := s.centralStore.GetUserByID(targetUserID)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		// Only admin or self can view, only admin can modify
		if claims.Role != "admin" && claims.Sub != targetUser.Username {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if r.Method == http.MethodGet {
			// Return effective tool permissions for this user
			s.fullConfig.RLock()
			roleDefaults, ok := s.fullConfig.ToolPermissions.RoleDefaults[targetUser.Role]
			s.fullConfig.RUnlock()

			if !ok {
				roleDefaults = []string{} // No permissions if role not configured
			}

			effectiveTools := targetUser.GetEffectiveToolPermissions(roleDefaults)

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"user_id":         targetUser.ID,
				"username":        targetUser.Username,
				"role":            targetUser.Role,
				"role_defaults":   roleDefaults,
				"user_overrides":  targetUser.AllowedTools,
				"effective_tools": effectiveTools,
			})
			return
		}

		if r.Method == http.MethodPut {
			// Only admin can modify tool permissions
			if claims.Role != "admin" {
				http.Error(w, "forbidden: admin role required", http.StatusForbidden)
				return
			}

			var in struct {
				AllowedTools *[]string `json:"allowed_tools"` // null = reset to defaults
			}
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}

			// Update user's tool overrides
			var toolsList []string
			if in.AllowedTools != nil {
				toolsList = *in.AllowedTools
			}
			_ = toolsList // TODO: implement when UpdateUserTools is added to CentralStorage

			// TODO: CentralStorage doesn't have UpdateUserTools yet - needs implementation
			// if err := s.centralStore.UpdateUserTools(targetUserID, toolsList); err != nil {
			logger.WarnCF("web", "UpdateUserTools not implemented for CentralStorage", map[string]interface{}{
				"target_user_id": targetUserID,
			})
			http.Error(w, "tool permissions update not yet implemented for multi-user mode", http.StatusNotImplemented)
			return
			/*
				logger.InfoCF("web", "User tool permissions updated", map[string]interface{}{
					"admin":          claims.Sub,
					"target_user_id": targetUserID,
					"tools":          tools,
				})

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success": true}`))
				return
			*/
		}
	} else {
		store = s.store
	}

	// Get target user
	if store == nil {
		http.Error(w, "storage not configured", http.StatusServiceUnavailable)
		return
	}

	targetUser, err := store.GetUserByID(targetUserID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Only admin or self can view, only admin can modify
	if claims.Role != "admin" && claims.Sub != targetUser.Username {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet {
		// Return effective tool permissions for this user
		s.fullConfig.RLock()
		roleDefaults, ok := s.fullConfig.ToolPermissions.RoleDefaults[targetUser.Role]
		s.fullConfig.RUnlock()

		if !ok {
			roleDefaults = []string{} // No permissions if role not configured
		}

		effectiveTools := targetUser.GetEffectiveToolPermissions(roleDefaults)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"user_id":         targetUser.ID,
			"username":        targetUser.Username,
			"role":            targetUser.Role,
			"role_defaults":   roleDefaults,
			"user_overrides":  targetUser.AllowedTools,
			"effective_tools": effectiveTools,
		})
		return
	}

	if r.Method == http.MethodPut {
		// Only admin can modify tool permissions
		if claims.Role != "admin" {
			http.Error(w, "forbidden: admin role required", http.StatusForbidden)
			return
		}

		var in struct {
			AllowedTools *[]string `json:"allowed_tools"` // null = reset to defaults
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		// Update user's tool overrides
		var tools []string
		if in.AllowedTools != nil {
			tools = *in.AllowedTools
		}

		if err := store.UpdateUserTools(targetUserID, tools); err != nil {
			logger.ErrorCF("web", "Failed to update user tools", map[string]interface{}{
				"target_user_id": targetUserID,
				"error":          err.Error(),
			})
			http.Error(w, "failed to update user tools", http.StatusInternalServerError)
			return
		}

		logger.InfoCF("web", "User tool permissions updated", map[string]interface{}{
			"admin":          claims.Sub,
			"target_user_id": targetUserID,
			"tools":          tools,
		})

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// handleToolAudit returns audit logs for tool executions
func (s *Server) handleToolAudit(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil || claims.Role != "admin" {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	filters := tools.AuditQueryFilters{
		Tool:   query.Get("tool"),
		Limit:  100, // Default limit
		Offset: 0,
	}

	if userIDStr := query.Get("user_id"); userIDStr != "" {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			filters.UserID = &userID
		}
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			filters.Limit = limit
		}
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	if startStr := query.Get("start"); startStr != "" {
		if start, err := time.Parse(time.RFC3339, startStr); err == nil {
			filters.StartTime = &start
		}
	}

	if endStr := query.Get("end"); endStr != "" {
		if end, err := time.Parse(time.RFC3339, endStr); err == nil {
			filters.EndTime = &end
		}
	}

	// Get audit logger from agent manager or create one
	var store *storage.Storage
	if s.centralStore != nil {
		// For centralStore, cannot use directly with audit logger since types don't match
		// We'd need to extend audit logger to work with CentralStorage
		// For now, fall back to regular store if available
		if s.store != nil {
			store = s.store
		} else {
			http.Error(w, "audit logging not available in central mode without legacy store", http.StatusServiceUnavailable)
			return
		}
	} else {
		store = s.store
	}

	if store == nil {
		http.Error(w, "audit logging not available", http.StatusServiceUnavailable)
		return
	}

	auditLogger, err := tools.NewSQLiteAuditLogger(store)
	if err != nil {
		logger.ErrorCF("web", "Failed to create audit logger", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "failed to initialize audit logger", http.StatusInternalServerError)
		return
	}

	logs, err := auditLogger.QueryLogs(r.Context(), filters)
	if err != nil {
		logger.ErrorCF("web", "Failed to query audit logs", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "failed to query audit logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
		"filters": map[string]interface{}{
			"user_id": filters.UserID,
			"tool":    filters.Tool,
			"limit":   filters.Limit,
			"offset":  filters.Offset,
		},
	})
}

// splitPath splits a URL path into segments
func splitPath(path string) []string {
	parts := make([]string, 0)
	for _, p := range strings.Split(path, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}
