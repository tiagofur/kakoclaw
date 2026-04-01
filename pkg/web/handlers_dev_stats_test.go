package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleSessionStatsReturnsCurrentStatsForAuthenticatedUser(t *testing.T) {
	s := newTestServer(t)
	s.sessionTracker = NewSessionTracker(t.TempDir(), 1000)
	admin, err := s.centralStore.GetUserByUsername("admin")
	if err != nil {
		t.Fatalf("load admin user: %v", err)
	}
	s.sessionTracker.Record(admin.UUID, "session-a", 750, 0.015)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dev/session/stats", nil)
	req.Header.Set("Authorization", "Bearer "+getTestToken(t, s))
	rr := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/dev/session/stats", s.handleSessionStats)
	s.authMiddleware(mux).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var out SessionStats
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out.SessionID != "session-a" {
		t.Fatalf("expected session-a, got %q", out.SessionID)
	}
	if out.TokensUsed != 750 {
		t.Fatalf("expected 750 tokens, got %d", out.TokensUsed)
	}
	if out.CostUSD != 0.015 {
		t.Fatalf("expected 0.015 cost, got %v", out.CostUSD)
	}
	if out.TokenLimit != 1000 {
		t.Fatalf("expected token limit 1000, got %d", out.TokenLimit)
	}
	if out.NumTurns != 750 {
		t.Fatalf("expected 750 turns, got %d", out.NumTurns)
	}
}

func TestHandleSessionStatsRejectsUnauthenticatedRequests(t *testing.T) {
	s := newTestServer(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/dev/session/stats", s.handleSessionStats)
	handler := s.authMiddleware(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dev/session/stats", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
