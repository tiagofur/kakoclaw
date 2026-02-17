package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sipeed/picoclaw/pkg/config"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	s := NewServerWithWorkspace(config.WebConfig{
		Username:  "admin",
		Password:  "StrongPassword123!",
		JWTExpiry: "24h",
	}, nil, dir)
	var err error
	s.authManager, err = newAuthManager(dir, s.cfg.Username, s.cfg.Password, s.cfg.JWTExpiry)
	if err != nil {
		t.Fatalf("newAuthManager failed: %v", err)
	}
	s.tasks, err = newTaskStore(dir)
	if err != nil {
		t.Fatalf("newTaskStore failed: %v", err)
	}
	t.Cleanup(func() { _ = s.tasks.close() })
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

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+created.ID+"/status", strings.NewReader(`{"status":"in_progress"}`))
	patchRR := httptest.NewRecorder()
	s.handleTasks(patchRR, patchReq)
	if patchRR.Code != http.StatusOK {
		t.Fatalf("expected 200 status patch, got %d", patchRR.Code)
	}

	putReq := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/"+created.ID, strings.NewReader(`{"title":"task b2","description":"x","status":"review","result":"done"}`))
	putRR := httptest.NewRecorder()
	s.handleTasks(putRR, putReq)
	if putRR.Code != http.StatusOK {
		t.Fatalf("expected 200 put, got %d", putRR.Code)
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
	if !strings.Contains(logsRR.Body.String(), "created") {
		t.Fatalf("expected created log, got: %s", logsRR.Body.String())
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

	created, err := s.tasks.create("mover estado", "", "backlog")
	if err != nil {
		t.Fatalf("create task for move command failed: %v", err)
	}
	ok, msg = s.handleTaskChatCommand("/task move " + created.ID + " done")
	if !ok || !strings.Contains(msg, "movida a done") {
		t.Fatalf("expected move command output, got ok=%v msg=%q", ok, msg)
	}
	got, err := s.tasks.get(created.ID)
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

func TestWebSocketOriginCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws/chat", nil)
	req.Host = "example.com"
	req.Header.Set("Origin", "https://example.com")
	if !checkWebSocketOrigin(req) {
		t.Fatal("expected same-host origin to pass")
	}
	req2 := httptest.NewRequest(http.MethodGet, "/ws/chat", nil)
	req2.Host = "example.com"
	req2.Header.Set("Origin", "https://evil.com")
	if checkWebSocketOrigin(req2) {
		t.Fatal("expected cross-host origin to fail")
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
