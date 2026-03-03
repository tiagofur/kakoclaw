package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/skills"
	"github.com/sipeed/makoclaw/pkg/storage"
)

// newTestServerWithAI creates a test server wired with a mock provider
// so that NewAgentLoopForUser succeeds and AI generation handlers can run.
func newTestServerWithAI(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()

	store, err := storage.New(config.StorageConfig{Path: filepath.Join(dir, "tasks.db")})
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	cs, err := storage.NewCentral(filepath.Join(dir, "central.db"))
	if err != nil {
		t.Fatalf("NewCentral: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	s := NewServerWithWorkspace(config.WebConfig{
		Username:  "admin",
		Password:  "StrongPassword123!",
		JWTExpiry: "24h",
	}, nil, dir)
	s.store = store
	s.centralStore = cs

	s.authManager, err = newAuthManager(cs, s.cfg.Username, s.cfg.Password, s.cfg.JWTExpiry)
	if err != nil {
		t.Fatalf("newAuthManager: %v", err)
	}

	userMgr := storage.NewUserStorageManager(cs, dir)
	s.userMgr = userMgr
	t.Cleanup(func() { userMgr.Close() })

	s.skillInstaller = skills.NewSkillInstaller(dir)

	// Wire fullConfig with mock provider so NewAgentLoopForUser succeeds
	s.fullConfig = &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Provider:  "mock",
				Model:     "mock-model",
				MaxTokens: 1024,
			},
		},
		Storage: config.StorageConfig{Path: filepath.Join(dir, "agent.db")},
	}
	s.msgBus = bus.NewMessageBus()

	return s
}

// TestHandleSkillGenerateConfig_ModelNotSystemPrompt verifies that the fix
// for passing systemPrompt as modelOverride works. Before the fix, the LLM
// call itself failed ("AI generation failed") because the entire system prompt
// was used as a model name. After the fix, the LLM call succeeds but the mock
// provider returns non-JSON, so we get "Failed to parse AI response" instead.
func TestHandleSkillGenerateConfig_ModelNotSystemPrompt(t *testing.T) {
	s := newTestServerWithAI(t)

	body := `{"prompt":"a skill that monitors system health"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/generate-config", strings.NewReader(body))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()

	s.handleSkillGenerateConfig(rr, req)

	// The mock provider doesn't return JSON, so we expect either:
	// - 200 if mock somehow returns valid JSON
	// - 500 with "Failed to parse AI response" (LLM succeeded, just not JSON)
	// But NOT 500 with "AI generation failed" (which would mean the model name was wrong)
	if rr.Code == http.StatusInternalServerError {
		var errResp map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err == nil {
			errMsg := errResp["error"]
			if strings.Contains(errMsg, "AI generation failed") {
				t.Fatalf("LLM call failed (model name likely corrupted): %s", errMsg)
			}
			// "Failed to parse AI response as JSON" is expected with mock provider
			t.Logf("Expected mock provider behavior: %s", errMsg)
		}
	}
}

func TestHandleSkillGenerateConfig_EmptyPrompt(t *testing.T) {
	s := newTestServerWithAI(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/generate-config", strings.NewReader(`{"prompt":""}`))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()

	s.handleSkillGenerateConfig(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty prompt, got %d", rr.Code)
	}
}

func TestHandleSkillGenerateConfig_MethodNotAllowed(t *testing.T) {
	s := newTestServerWithAI(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/skills/generate-config", nil)
	rr := httptest.NewRecorder()

	s.handleSkillGenerateConfig(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

// TestHandleSpecialistGenerate_LLMCallSucceeds verifies the LLM call itself
// works (model name is correct). Mock provider doesn't return valid JSON for
// the specialist schema, so we expect a JSON parse error, not an LLM failure.
func TestHandleSpecialistGenerate_LLMCallSucceeds(t *testing.T) {
	s := newTestServerWithAI(t)

	body := `{"description":"an agent that reviews code quality"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/generate", strings.NewReader(body))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()

	s.handleSpecialistGenerate(rr, req)

	if rr.Code == http.StatusInternalServerError {
		var errResp map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err == nil {
			errMsg := errResp["error"]
			if strings.Contains(errMsg, "AI processing failed") {
				t.Fatalf("LLM call failed unexpectedly: %s", errMsg)
			}
			// "AI produced invalid JSON" is expected with mock provider
			t.Logf("Expected mock provider behavior: %s", errMsg)
		}
	}
}

func TestHandleSpecialistGenerate_EmptyDescription(t *testing.T) {
	s := newTestServerWithAI(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/generate", strings.NewReader(`{"description":""}`))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()

	s.handleSpecialistGenerate(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty description, got %d", rr.Code)
	}
}

func TestHandleSpecialistGenerate_MethodNotAllowed(t *testing.T) {
	s := newTestServerWithAI(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/generate", nil)
	rr := httptest.NewRecorder()

	s.handleSpecialistGenerate(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

// TestHandleSkillGenerateConfig_NoAIConfig verifies graceful degradation
// when the server lacks AI configuration.
func TestHandleSkillGenerateConfig_NoAIConfig(t *testing.T) {
	s := newTestServer(t) // no fullConfig or msgBus

	body := `{"prompt":"test skill"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/generate-config", strings.NewReader(body))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()

	s.handleSkillGenerateConfig(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when AI not configured, got %d: %s", rr.Code, rr.Body.String())
	}
}

// TestHandleSpecialistGenerate_NoAIConfig verifies graceful degradation
// when the server lacks AI configuration.
func TestHandleSpecialistGenerate_NoAIConfig(t *testing.T) {
	s := newTestServer(t) // no fullConfig or msgBus

	body := `{"description":"test agent"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/generate", strings.NewReader(body))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()

	s.handleSpecialistGenerate(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when AI not configured, got %d: %s", rr.Code, rr.Body.String())
	}
}
