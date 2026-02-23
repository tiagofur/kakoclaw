package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sipeed/makoclaw/pkg/logger"
)

// handleSetupInitialize creates a new setup session
// POST /api/v1/setup/initialize
func (s *Server) handleSetupInitialize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Channel  string            `json:"channel"`
		SenderID string            `json:"sender_id"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Channel == "" || req.SenderID == "" {
		http.Error(w, "channel and sender_id are required", http.StatusBadRequest)
		return
	}

	// Create setup session
	session, err := s.store.CreateSetupSession(req.Channel, req.SenderID, req.Metadata)
	if err != nil {
		logger.ErrorCF("web", "Failed to create setup session", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "failed to create setup session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      session.Token,
		"expires_at": session.ExpiresAt,
		"setup_url":  fmt.Sprintf("%s/onboarding?token=%s", getWebBaseURL(), session.Token),
	})
}

// handleSetupValidate validates a setup token
// GET /api/v1/setup/validate/{token}
func (s *Server) handleSetupValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/v1/setup/validate/")
	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	session, err := s.store.GetSetupSession(token)
	if err != nil {
		logger.ErrorCF("web", "Failed to get setup session", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "failed to validate token", http.StatusInternalServerError)
		return
	}

	if session == nil {
		http.Error(w, "token not found or expired", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":     true,
		"channel":   session.Channel,
		"sender_id": session.SenderID,
		"metadata":  session.Metadata,
	})
}

// handleSetupComplete finalizes a setup session
// POST /api/v1/setup/complete/{token}
func (s *Server) handleSetupComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require authentication
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/v1/setup/complete/")
	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	// Get setup session
	session, err := s.store.GetSetupSession(token)
	if err != nil {
		logger.ErrorCF("web", "Failed to get setup session", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "failed to get setup session", http.StatusInternalServerError)
		return
	}

	if session == nil {
		http.Error(w, "token not found or expired", http.StatusNotFound)
		return
	}

	// Get user ID from username
	user, err := s.resolveUserByUsername(claims.Sub)
	if err != nil {
		logger.ErrorCF("web", "Failed to get user", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Update session with user ID
	if err := s.store.UpdateSetupSession(token, user.ID); err != nil {
		logger.ErrorCF("web", "Failed to update setup session", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "failed to complete setup", http.StatusInternalServerError)
		return
	}

	// Log successful setup
	logger.InfoCF("web", "Setup session completed", map[string]interface{}{
		"token":     token,
		"user_id":   user.ID,
		"channel":   session.Channel,
		"sender_id": session.SenderID,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Setup completed successfully",
	})
}

// getWebBaseURL returns the web server base URL
func getWebBaseURL() string {
	// This could be made configurable, but for now return localhost
	return "http://localhost:8080"
}
