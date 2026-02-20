package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/sipeed/kakoclaw/pkg/storage"
)

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	// Require admin privileges
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil || claims.Role != "admin" {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet {
		users, err := s.store.ListUsers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Redact password hashes
		for _, u := range users {
			u.PasswordHash = ""
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(users)
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

		user, err := s.store.CreateUser(in.Username, in.Password, in.Role)
		if err != nil {
			if err == storage.ErrUserExists {
				http.Error(w, "user already exists", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user.PasswordHash = ""
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(user)
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
			if err := s.store.UpdateUserPassword(userID, in.Password); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if in.Role != "" {
			if in.Role != "admin" && in.Role != "user" {
				http.Error(w, "invalid role", http.StatusBadRequest)
				return
			}
			if err := s.store.UpdateUserRole(userID, in.Role); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	if r.Method == http.MethodDelete {
		user, err := s.store.GetUserByID(userID)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		// Prevent deleting the last admin
		if user.Role == "admin" {
			users, err := s.store.ListUsers()
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

		if err := s.store.DeleteUser(userID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}
