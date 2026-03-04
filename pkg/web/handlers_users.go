package web

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
)

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	// Require admin privileges
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil || claims.Role != "admin" {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet {
		var users []*storage.User
		var err error
		if s.centralStore != nil {
			users, err = s.centralStore.ListUsers()
		} else {
			users, err = s.store.ListUsers()
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Redact password hashes
		for _, u := range users {
			u.PasswordHash = ""
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, users)
		return
	}

	if r.Method == http.MethodPost {
		var in struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if in.Role == "" {
			in.Role = "user"
		}
		if len(in.Username) > 64 {
			http.Error(w, "username too long (max 64 chars)", http.StatusBadRequest)
			return
		}

		var user *storage.User
		var err error
		if s.centralStore != nil {
			user, err = s.centralStore.CreateUser(in.Username, in.Password, in.Role)
		} else {
			user, err = s.store.CreateUser(in.Username, in.Password, in.Role)
		}
		if err != nil {
			if err == storage.ErrUserExists {
				http.Error(w, "user already exists", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Initialize per-user workspace and database
		if s.userMgr != nil && user.UUID != "" {
			if err := s.userMgr.EnsureUserDirectory(user.UUID); err != nil {
				logger.WarnCF("web", "Failed to create user workspace", map[string]interface{}{
					"user": user.Username, "error": err.Error(),
				})
			}
			// Open/create per-user database
			if _, err := s.userMgr.GetOrCreate(user.UUID); err != nil {
				logger.WarnCF("web", "Failed to create user database", map[string]interface{}{
					"user": user.Username, "error": err.Error(),
				})
			}
			// Seed a default config.json in the user's data directory so the
			// user-specific config is present from the start (avoids falling back
			// to the global config on first load).
			if s.fullConfig != nil {
				if err := config.SaveConfigForUser(user.UUID, s.fullConfig); err != nil {
					logger.WarnCF("web", "Failed to seed user config.json", map[string]interface{}{
						"user": user.Username, "error": err.Error(),
					})
				}
			}
		}

		user.PasswordHash = ""
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		writeJSONResponse(w, user)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleUserAction(w http.ResponseWriter, r *http.Request) {
	// Require admin privileges
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil || claims.Role != "admin" {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/users/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	idStr := pathParts[0]
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPut {
		var in struct {
			Password string `json:"password,omitempty"`
			Role     string `json:"role,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if in.Password != "" {
			var err error
			if s.centralStore != nil {
				err = s.centralStore.UpdateUserPassword(userID, in.Password)
			} else {
				err = s.store.UpdateUserPassword(userID, in.Password)
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if in.Role != "" {
			if in.Role != "admin" && in.Role != "user" {
				http.Error(w, "invalid role", http.StatusBadRequest)
				return
			}
			var err error
			if s.centralStore != nil {
				err = s.centralStore.UpdateUserRole(userID, in.Role)
			} else {
				err = s.store.UpdateUserRole(userID, in.Role)
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]string{"status": "ok"})
		return
	}

	if r.Method == http.MethodDelete {
		var user *storage.User
		var err error
		if s.centralStore != nil {
			user, err = s.centralStore.GetUserByID(userID)
		} else {
			user, err = s.store.GetUserByID(userID)
		}
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		// Prevent deleting the last admin
		if user.Role == "admin" {
			var users []*storage.User
			if s.centralStore != nil {
				users, err = s.centralStore.ListUsers()
			} else {
				users, err = s.store.ListUsers()
			}
			if err == nil {
				adminCount := 0
				for _, u := range users {
					if u.Role == "admin" {
						adminCount++
					}
				}
				if adminCount <= 1 {
					http.Error(w, "cannot delete the last admin user", http.StatusConflict)
					return
				}
			}
		}

		if s.centralStore != nil {
			err = s.centralStore.DeleteUser(userID)
		} else {
			err = s.store.DeleteUser(userID)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Archive per-user data (close connection if open)
		if s.userMgr != nil && user.UUID != "" {
			_ = s.userMgr.CloseUser(user.UUID)
		}

		// Clean up user's workspace directory to prevent orphaned files
		if user.UUID != "" && s.fullConfig != nil && s.fullConfig.Storage.Path != "" {
			userWorkspace := filepath.Join(s.fullConfig.Storage.Path, "users", user.UUID)
			if err := os.RemoveAll(userWorkspace); err != nil {
				logger.WarnCF("web", "Failed to cleanup user workspace", map[string]interface{}{
					"user_uuid": user.UUID,
					"path":      userWorkspace,
					"error":     err.Error(),
				})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]string{"status": "ok"})
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleBlockUser(w http.ResponseWriter, r *http.Request) {
	// Require admin privileges
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil || claims.Role != "admin" {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/users/"), "/")
	if len(pathParts) < 2 || pathParts[0] == "" {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var in struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Validate reason
	in.Reason = strings.TrimSpace(in.Reason)
	if len(in.Reason) < 10 {
		http.Error(w, "block reason must be at least 10 characters", http.StatusBadRequest)
		return
	}

	// Get admin user ID
	var adminUser *storage.User
	if s.centralStore != nil {
		adminUser, err = s.centralStore.GetUserByUsername(claims.Sub)
	} else {
		adminUser, err = s.store.GetUserByUsername(claims.Sub)
	}
	if err != nil {
		http.Error(w, "admin user not found", http.StatusInternalServerError)
		return
	}

	// Block the user
	if s.centralStore != nil {
		err = s.centralStore.BlockUser(userID, adminUser.ID, in.Reason)
	} else {
		err = s.store.BlockUser(userID, adminUser.ID, in.Reason)
	}
	if err != nil {
		if err == storage.ErrCannotBlockSelf {
			http.Error(w, "cannot block yourself", http.StatusConflict)
			return
		}
		if err == storage.ErrCannotBlockLastAdmin {
			http.Error(w, "cannot block the last admin user", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated user
	var user *storage.User
	if s.centralStore != nil {
		user, err = s.centralStore.GetUserByID(userID)
	} else {
		user, err = s.store.GetUserByID(userID)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.PasswordHash = ""

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, user)
}

func (s *Server) handleUnblockUser(w http.ResponseWriter, r *http.Request) {
	// Require admin privileges
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil || claims.Role != "admin" {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/users/"), "/")
	if len(pathParts) < 2 || pathParts[0] == "" {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// Unblock the user
	if s.centralStore != nil {
		err = s.centralStore.UnblockUser(userID)
	} else {
		err = s.store.UnblockUser(userID)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated user
	var user *storage.User
	if s.centralStore != nil {
		user, err = s.centralStore.GetUserByID(userID)
	} else {
		user, err = s.store.GetUserByID(userID)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.PasswordHash = ""

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, user)
}
