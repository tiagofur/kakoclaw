package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
)

func TestHandleDevProjects_Unauthorized(t *testing.T) {
	s := NewServer(config.WebConfig{}, nil, nil)
	// Do not add auth token

	req := httptest.NewRequest("GET", "/api/v1/dev/projects", nil)
	w := httptest.NewRecorder()

	s.server = &http.Server{Handler: s.authMiddleware(http.HandlerFunc(s.handleDevProjects))}
	s.server.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 Unauthorized, got %d", w.Code)
	}
}

func TestHandleDevBridgeStatus_Disabled(t *testing.T) {
	s := NewServer(config.WebConfig{}, nil, nil)
	s.SetFullConfig(&config.Config{}) // DevStudio globally disabled by default

	req := httptest.NewRequest("GET", "/api/v1/dev/bridge/status", nil)
	// Inject mock claims
	req = req.WithContext(injectMockClaims(req.Context(), "test-uuid"))
	w := httptest.NewRecorder()

	s.handleDevBridgeStatus(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected 500 when disabled, got %d", w.Code)
	}
}

func injectMockClaims(ctx context.Context, uuid string) context.Context {
	return context.WithValue(ctx, userClaimsKey, &jwtClaims{UUID: uuid})
}
