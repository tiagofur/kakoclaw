package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/skills"
	"github.com/sipeed/kakoclaw/pkg/storage"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	
	store, err := storage.New(config.StorageConfig{Path: filepath.Join(dir, "tasks.db")})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	s := NewServerWithWorkspace(config.WebConfig{
		Username:  "admin",
		Password:  "StrongPassword123!",
		JWTExpiry: "24h",
	}, nil, dir)
	s.store = store

	s.authManager, err = newAuthManager(s.store, s.cfg.Username, s.cfg.Password, s.cfg.JWTExpiry)
	if err != nil {
		t.Fatalf("newAuthManager failed: %v", err)
	}

	return s
}

func TestAuthMiddlewareBlocksUnauthorizedAPI(t *testing.T) {
	s := newTestServer(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := s.authMiddleware(next)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestLoginAndBearerAuthorization(t *testing.T) {
	s := newTestServer(t)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"StrongPassword123!"}`))
	loginRR := httptest.NewRecorder()
	s.handleLogin(loginRR, loginReq)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("expected 200 login, got %d", loginRR.Code)
	}
	var out map[string]string
	if err := json.Unmarshal(loginRR.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid login response: %v", err)
	}
	token := out["token"]
	if token == "" {
		t.Fatal("expected token")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if !s.isAuthorized(req) {
		t.Fatal("expected jwt bearer auth to pass")
	}
}

func TestHandleTasksCreateAndListSQLite(t *testing.T) {
	s := newTestServer(t)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{"title":"task a"}`))
	createRR := httptest.NewRecorder()
	s.handleTasks(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createRR.Code)
	}
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	listRR := httptest.NewRecorder()
	s.handleTasks(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listRR.Code)
	}
	if !strings.Contains(listRR.Body.String(), `"title":"task a"`) {
		t.Fatalf("expected task in list, got: %s", listRR.Body.String())
	}
}

func TestHandleTasksUpdateStatusAndDelete(t *testing.T) {
	s := newTestServer(t)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{"title":"task b","status":"todo"}`))
	createRR := httptest.NewRecorder()
	s.handleTasks(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createRR.Code)
	}
	var created taskItem
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("invalid create response: %v", err)
	}

	putReq := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/"+created.ID, strings.NewReader(`{"title":"task b updated","description":"updated","status":"review","result":"done soon"}`))
	putRR := httptest.NewRecorder()
	s.handleTasks(putRR, putReq)
	if putRR.Code != http.StatusOK {
		t.Fatalf("expected 200 put update, got %d", putRR.Code)
	}

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+created.ID+"/status", strings.NewReader(`{"status":"in_progress"}`))
	patchRR := httptest.NewRecorder()
	s.handleTasks(patchRR, patchReq)
	if patchRR.Code != http.StatusOK {
		t.Fatalf("expected 200 status patch, got %d", patchRR.Code)
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+created.ID, nil)
	delRR := httptest.NewRecorder()
	s.handleTasks(delRR, delReq)
	if delRR.Code != http.StatusOK {
		t.Fatalf("expected 200 delete, got %d", delRR.Code)
	}
}

func TestHandleTaskLogsEndpoint(t *testing.T) {
	s := newTestServer(t)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{"title":"task logs"}`))
	createRR := httptest.NewRecorder()
	s.handleTasks(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createRR.Code)
	}
	var created taskItem
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("invalid create response: %v", err)
	}
	logsReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+created.ID+"/logs", nil)
	logsRR := httptest.NewRecorder()
	s.handleTasks(logsRR, logsReq)
	if logsRR.Code != http.StatusOK {
		t.Fatalf("expected 200 logs, got %d", logsRR.Code)
	}
}

func TestTaskChatCommands(t *testing.T) {
	s := newTestServer(t)
	ok, msg := s.handleTaskChatCommand("/task create revisar logs")
	if !ok || !strings.Contains(msg, "Tarea creada") {
		t.Fatalf("expected create command handled, got ok=%v msg=%q", ok, msg)
	}

	ok, msg = s.handleTaskChatCommand("/task list")
	if !ok || !strings.Contains(msg, "revisar logs") {
		t.Fatalf("expected list command output, got ok=%v msg=%q", ok, msg)
	}

	createdID, err := s.store.CreateTask("mover estado", "", "backlog")
	if err != nil {
		t.Fatalf("create task for move command failed: %v", err)
	}
	idStr := toString(createdID)
	ok, msg = s.handleTaskChatCommand("/task move " + idStr + " done")
	if !ok || !strings.Contains(msg, "movida a done") {
		t.Fatalf("expected move command output, got ok=%v msg=%q", ok, msg)
	}
	got, err := s.store.GetTask(createdID)
	if err != nil {
		t.Fatalf("get moved task failed: %v", err)
	}
	if got.Status != "done" {
		t.Fatalf("expected status done, got %s", got.Status)
	}
}

func TestHandleTasksRejectsEmptyTitle(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{"title":"  "}`))
	rr := httptest.NewRecorder()
	s.handleTasks(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleSkillsAvailableReturnsEmptyOnFetcherError(t *testing.T) {
	s := newTestServer(t)
	s.skillInstaller = skills.NewSkillInstaller(t.TempDir())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/skills?type=available", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	s.handleSkills(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"skills":[]`) {
		t.Fatalf("expected empty skills list, got: %s", rr.Body.String())
	}
}

func TestHandleModelsReturnsStableShape(t *testing.T) {
	s := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/models", nil)

	s.handleModels(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var out struct {
		Providers []struct {
			Models []map[string]interface{} `json:"models"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid response json: %v", err)
	}
	if out.Providers == nil {
		t.Fatal("expected providers to be an empty array, not null")
	}
}

func TestHandleSkillGenerateRequiresAgentLoop(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/generate", strings.NewReader(`{"name":"demo-skill","goal":"help with docs"}`))
	rr := httptest.NewRecorder()

	s.handleSkillAction(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}

func TestHandleSkillCreateWritesSkillFile(t *testing.T) {
	s := newTestServer(t)
	content := "---\nname: demo-skill\ndescription: Test skill\n---\n\n# Demo Skill\n\n## When to use\nUse this.\n"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/create", strings.NewReader(`{"name":"demo-skill","content":`+strconvQuote(content)+`}`))
	rr := httptest.NewRecorder()

	s.handleSkillAction(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	skillPath := filepath.Join(s.workspace, "skills", "demo-skill", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("expected skill file to exist: %v", err)
	}
}

func strconvQuote(v string) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestWebSocketOriginCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws/chat", nil)
	req.Host = "example.com"
	req.Header.Set("Origin", "https://example.com")
	if !checkWebSocketOrigin(req) {
		t.Fatal("expected same-host origin to pass")
	}
	req2 := httptest.NewRequest(http.MethodGet, "/ws/chat", nil)
	req2.Host = "example.com"
	req2.Header.Set("Origin", "https://other.com")
	if !checkWebSocketOrigin(req2) {
		t.Fatal("expected cross-host origin to also pass as per current implementation")
	}
}

func TestIsAuthorizedAllowsJWTInWebSocketQuery(t *testing.T) {
	s := newTestServer(t)
	token, err := s.authManager.login("admin", "StrongPassword123!")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/ws/chat?token="+token, nil)
	if !s.isAuthorized(req) {
		t.Fatal("expected websocket query jwt to authorize")
	}
}

// --- Chat session endpoints ---

func TestHandleChatSessionsListEmpty(t *testing.T) {
	s := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions", nil)
	s.handleChatSessions(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var out struct {
		Sessions []interface{} `json:"sessions"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestHandleChatSessionsListWithData(t *testing.T) {
	s := newTestServer(t)
	_ = s.store.SaveMessage("web:sess1", "user", "hello")
	_ = s.store.SaveMessage("web:sess1", "assistant", "hi")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions", nil)
	s.handleChatSessions(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "web:sess1") {
		t.Fatalf("expected session in list, got: %s", rr.Body.String())
	}
}

func TestHandleChatSessionsArchivedFilter(t *testing.T) {
	s := newTestServer(t)
	_ = s.store.SaveMessage("active:x", "user", "active")
	_ = s.store.SaveMessage("archived:x", "user", "archived")
	archivedTrue := true
	_, _ = s.store.UpdateSession("archived:x", nil, &archivedTrue)

	// Non-archived (default)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions?archived=false", nil)
	s.handleChatSessions(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if strings.Contains(rr.Body.String(), "archived:x") {
		t.Fatalf("archived session should not appear: %s", rr.Body.String())
	}

	// Archived only
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions?archived=true", nil)
	s.handleChatSessions(rr2, req2)
	if !strings.Contains(rr2.Body.String(), "archived:x") {
		t.Fatalf("expected archived session: %s", rr2.Body.String())
	}
}

func TestHandleChatSessionMessagesGet(t *testing.T) {
	s := newTestServer(t)
	_ = s.store.SaveMessage("msgs:test", "user", "hello msg")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions/msgs:test", nil)
	s.handleChatSessionMessages(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "hello msg") {
		t.Fatalf("expected message content, got: %s", rr.Body.String())
	}
}

func TestHandleChatSessionMessagesPatch(t *testing.T) {
	s := newTestServer(t)
	_ = s.store.SaveMessage("patch:test", "user", "msg")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/chat/sessions/patch:test", strings.NewReader(`{"title":"My Title"}`))
	s.handleChatSessionMessages(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "My Title") {
		t.Fatalf("expected updated title, got: %s", rr.Body.String())
	}
}

func TestHandleChatSessionMessagesDelete(t *testing.T) {
	s := newTestServer(t)
	_ = s.store.SaveMessage("del:test", "user", "bye")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/chat/sessions/del:test", nil)
	s.handleChatSessionMessages(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify deleted
	msgs, _ := s.store.GetMessages("del:test")
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages after delete, got %d", len(msgs))
	}
}

func TestHandleChatSessionMessagesNoID(t *testing.T) {
	s := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions/", nil)
	s.handleChatSessionMessages(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
