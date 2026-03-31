package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/cron"
	"github.com/sipeed/makoclaw/pkg/skills"
	"github.com/sipeed/makoclaw/pkg/storage"
)

func captureStdLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := new(bytes.Buffer)
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	log.SetOutput(buf)
	log.SetFlags(0)
	log.SetPrefix("")
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	})
	return buf
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()

	store, err := storage.New(config.StorageConfig{Path: filepath.Join(dir, "tasks.db")})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	// Create central storage for auth operations
	cs, err := storage.NewCentral(filepath.Join(dir, "central.db"))
	if err != nil {
		t.Fatalf("NewCentral failed: %v", err)
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
		t.Fatalf("newAuthManager failed: %v", err)
	}

	// Wire per-user storage manager so getUserStorage returns isolated per-user DBs
	userMgr := storage.NewUserStorageManager(cs, dir)
	s.userMgr = userMgr
	t.Cleanup(func() { userMgr.Close() })

	s.skillInstaller = skills.NewSkillInstaller(dir)

	return s
}

func getTestToken(t *testing.T, s *Server) string {
	t.Helper()
	token, err := s.authManager.login("admin", "StrongPassword123!")
	if err != nil {
		t.Fatalf("failed to get test token: %v", err)
	}
	return token
}

func withTestUserContext(t *testing.T, s *Server, req *http.Request) *http.Request {
	t.Helper()
	// Ensure admin user exists in central store (or fallback to legacy store)
	var userUUID string
	if s.centralStore != nil {
		user, err := s.centralStore.GetUserByUsername("admin")
		if err != nil {
			user, err = s.centralStore.CreateUser("admin", "StrongPassword123!", "admin")
			if err != nil && err != storage.ErrUserExists {
				t.Fatalf("failed to create test user: %v", err)
			}
			if user == nil {
				user, _ = s.centralStore.GetUserByUsername("admin")
			}
		}
		if user != nil {
			userUUID = user.UUID
		}
	} else {
		_, err := s.store.GetUserByUsername("admin")
		if err != nil {
			_, err = s.store.CreateUserWithEmail("admin", "admin@test.com", "StrongPassword123!", "admin")
			if err != nil {
				t.Fatalf("failed to create test user: %v", err)
			}
		}
	}
	claims := &jwtClaims{
		Sub:  "admin",
		UUID: userUUID,
		Role: "admin",
	}
	ctx := context.WithValue(req.Context(), userClaimsKey, claims)
	return req.WithContext(ctx)
}

func withUserContext(req *http.Request, user *storage.User) *http.Request {
	claims := &jwtClaims{
		Sub:  user.Username,
		UUID: user.UUID,
		Role: user.Role,
	}
	ctx := context.WithValue(req.Context(), userClaimsKey, claims)
	return req.WithContext(ctx)
}

func createUserForTest(t *testing.T, s *Server, username string) *storage.User {
	t.Helper()

	user, err := s.centralStore.CreateUserWithEmail(username, username+"@test.com", "StrongPassword123!", "user")
	if err != nil {
		if err != storage.ErrUserExists {
			t.Fatalf("failed to create user %s: %v", username, err)
		}
		user, err = s.centralStore.GetUserByUsername(username)
		if err != nil {
			t.Fatalf("failed to load existing user %s: %v", username, err)
		}
	}
	if user == nil {
		t.Fatalf("user %s is nil", username)
	}
	if s.userMgr != nil {
		if _, err := s.userMgr.GetOrCreate(user.UUID); err != nil {
			t.Fatalf("failed to create per-user store for %s: %v", username, err)
		}
	}
	return user
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
	createReq = withTestUserContext(t, s, createReq)
	createRR := httptest.NewRecorder()
	s.handleTasks(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createRR.Code)
	}
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	listReq = withTestUserContext(t, s, listReq)
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
	createReq = withTestUserContext(t, s, createReq)
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
	putReq = withTestUserContext(t, s, putReq)
	putRR := httptest.NewRecorder()
	s.handleTasks(putRR, putReq)
	if putRR.Code != http.StatusOK {
		t.Fatalf("expected 200 put update, got %d", putRR.Code)
	}

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+created.ID+"/status", strings.NewReader(`{"status":"in_progress"}`))
	patchReq = withTestUserContext(t, s, patchReq)
	patchRR := httptest.NewRecorder()
	s.handleTasks(patchRR, patchReq)
	if patchRR.Code != http.StatusOK {
		t.Fatalf("expected 200 status patch, got %d", patchRR.Code)
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+created.ID, nil)
	delReq = withTestUserContext(t, s, delReq)
	delRR := httptest.NewRecorder()
	s.handleTasks(delRR, delReq)
	if delRR.Code != http.StatusOK {
		t.Fatalf("expected 200 delete, got %d", delRR.Code)
	}
}

func TestHandleTaskLogsEndpoint(t *testing.T) {
	s := newTestServer(t)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{"title":"task logs"}`))
	createReq = withTestUserContext(t, s, createReq)
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
	logsReq = withTestUserContext(t, s, logsReq)
	logsRR := httptest.NewRecorder()
	s.handleTasks(logsRR, logsReq)
	if logsRR.Code != http.StatusOK {
		t.Fatalf("expected 200 logs, got %d", logsRR.Code)
	}
}

func TestTaskChatCommands(t *testing.T) {
	s := newTestServer(t)

	// Get per-user storage for the admin user
	user, err := s.centralStore.GetUserByUsername("admin")
	if err != nil {
		user, err = s.centralStore.CreateUser("admin", "StrongPassword123!", "admin")
		if err != nil && err != storage.ErrUserExists {
			t.Fatalf("failed to create admin user: %v", err)
		}
		user, err = s.centralStore.GetUserByUsername("admin")
		if err != nil {
			t.Fatalf("failed to get admin user: %v", err)
		}
	}
	if user == nil {
		t.Fatal("admin user not found")
	}
	userStore, err := s.userMgr.GetOrCreate(user.UUID)
	if err != nil {
		t.Fatalf("failed to get per-user store: %v", err)
	}

	ok, msg := s.handleTaskChatCommand(user.ID, userStore, "/task create revisar logs")
	if !ok || !strings.Contains(msg, "Tarea creada") {
		t.Fatalf("expected create command handled, got ok=%v msg=%q", ok, msg)
	}

	ok, msg = s.handleTaskChatCommand(user.ID, userStore, "/task list")
	if !ok || !strings.Contains(msg, "revisar logs") {
		t.Fatalf("expected list command output, got ok=%v msg=%q", ok, msg)
	}

	createdID, err := userStore.CreateTaskForUser(user.ID, "mover estado", "", "backlog", "")
	if err != nil {
		t.Fatalf("create task for move command failed: %v", err)
	}
	idStr := toString(createdID)
	ok, msg = s.handleTaskChatCommand(user.ID, userStore, "/task move "+idStr+" done")
	if !ok || !strings.Contains(msg, "movida a done") {
		t.Fatalf("expected move command output, got ok=%v msg=%q", ok, msg)
	}
	got, err := userStore.GetTaskForUser(user.ID, createdID)
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
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleTasks(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestProcessNextTodoTaskLegacySkipsWhenUserManagerPresent(t *testing.T) {
	s := newTestServer(t)
	logBuf := captureStdLog(t)

	legacyStore, err := storage.New(config.StorageConfig{Path: filepath.Join(t.TempDir(), "legacy.db")})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	t.Cleanup(func() { _ = legacyStore.Close() })

	const legacyUserID int64 = 1
	if _, err := legacyStore.ExecRaw(`INSERT INTO users (id, uuid, username, password_hash, role, email) VALUES (?, ?, ?, ?, ?, ?)`, legacyUserID, "legacy-user-1", "legacy-admin", "hash", "admin", "legacy@example.com"); err != nil {
		t.Fatalf("insert legacy user failed: %v", err)
	}
	if _, err := legacyStore.CreateTaskForUser(legacyUserID, "legacy todo", "should not run", "todo", ""); err != nil {
		t.Fatalf("CreateTaskForUser failed: %v", err)
	}

	s.store = legacyStore
	s.userMgr = storage.NewUserStorageManager(s.centralStore, t.TempDir())
	t.Cleanup(func() { s.userMgr.Close() })

	s.processNextTodoTaskLegacy(context.Background())

	tasks, err := legacyStore.ListAllUsersTasks(false)
	if err != nil {
		t.Fatalf("ListAllUsersTasks failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Status != "todo" {
		t.Fatalf("expected task to remain todo, got %q", tasks[0].Status)
	}

	if !strings.Contains(logBuf.String(), "processNextTodoTaskLegacy called in multi-user mode — skipping") {
		t.Fatalf("expected skip log, got %q", logBuf.String())
	}
}

func TestLogUnsafeLegacyTaskWorkerModeEmitsFatalCondition(t *testing.T) {
	s := newTestServer(t)
	logBuf := captureStdLog(t)

	legacyStore, err := storage.New(config.StorageConfig{Path: filepath.Join(t.TempDir(), "legacy-multi.db")})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	t.Cleanup(func() { _ = legacyStore.Close() })

	if _, err := legacyStore.ExecRaw(`INSERT INTO users (id, uuid, username, password_hash, role, email) VALUES (?, ?, ?, ?, ?, ?)`, 1, "legacy-user-a", "legacy-admin-a", "hash", "admin", "a@example.com"); err != nil {
		t.Fatalf("insert user A failed: %v", err)
	}
	if _, err := legacyStore.ExecRaw(`INSERT INTO users (id, uuid, username, password_hash, role, email) VALUES (?, ?, ?, ?, ?, ?)`, 2, "legacy-user-b", "legacy-admin-b", "hash", "admin", "b@example.com"); err != nil {
		t.Fatalf("insert user B failed: %v", err)
	}

	s.store = legacyStore
	s.userMgr = nil

	s.logUnsafeLegacyTaskWorkerMode()

	output := logBuf.String()
	if !strings.Contains(output, "FATAL: UNSAFE legacy mode with multiple users detected") {
		t.Fatalf("expected fatal condition log, got %q", output)
	}
	if !strings.Contains(output, "[ERROR]") {
		t.Fatalf("expected error level log, got %q", output)
	}
}

func TestHandleMetricsIsolatesPerUserLLMCounts(t *testing.T) {
	s := newTestServer(t)
	userA := createUserForTest(t, s, "metrics-a")
	userB := createUserForTest(t, s, "metrics-b")

	s.getOrCreateUserMetrics(userA.UUID).RecordLLMCall("model-a", 120*time.Millisecond, 11, 13, nil)
	s.getOrCreateUserMetrics(userB.UUID).RecordLLMCall("model-b", 80*time.Millisecond, 7, 9, nil)
	s.getOrCreateUserMetrics(userB.UUID).RecordLLMCall("model-b", 90*time.Millisecond, 5, 6, nil)

	type metricsResponse struct {
		LLMCalls     int64                       `json:"llm_calls"`
		LLMByModel   map[string]map[string]int64 `json:"llm_by_model"`
		RecentEvents []map[string]interface{}    `json:"recent_events"`
	}

	assertMetrics := func(name string, user *storage.User, expectedCalls int64, expectedModel string, unexpectedModel string) {
		t.Helper()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
		req = withUserContext(req, user)
		rr := httptest.NewRecorder()

		s.handleMetrics(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d: %s", name, rr.Code, rr.Body.String())
		}

		var out metricsResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
			t.Fatalf("%s: invalid metrics response: %v", name, err)
		}
		if out.LLMCalls != expectedCalls {
			t.Fatalf("%s: expected llm_calls=%d, got %d", name, expectedCalls, out.LLMCalls)
		}
		if _, ok := out.LLMByModel[expectedModel]; !ok {
			t.Fatalf("%s: expected model %q in response: %#v", name, expectedModel, out.LLMByModel)
		}
		if _, ok := out.LLMByModel[unexpectedModel]; ok {
			t.Fatalf("%s: unexpected model %q leaked into response: %#v", name, unexpectedModel, out.LLMByModel)
		}
		if len(out.RecentEvents) != int(expectedCalls) {
			t.Fatalf("%s: expected %d recent events, got %d", name, expectedCalls, len(out.RecentEvents))
		}
		for _, evt := range out.RecentEvents {
			if evt["model"] != expectedModel {
				t.Fatalf("%s: unexpected event model %#v", name, evt["model"])
			}
		}
	}

	assertMetrics("user-a", userA, 1, "model-a", "model-b")
	assertMetrics("user-b", userB, 2, "model-b", "model-a")
}

func TestGetCronServiceForRequestReturnsErrCronNotInitialized(t *testing.T) {
	s := newTestServer(t)
	user := createUserForTest(t, s, "cron-user")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cron", nil)
	req = withUserContext(req, user)

	cronService, userID, err := s.getCronServiceForRequest(req)
	if !errors.Is(err, cron.ErrCronNotInitialized) {
		t.Fatalf("expected ErrCronNotInitialized, got %v", err)
	}
	if cronService != nil {
		t.Fatalf("expected nil cron service, got %#v", cronService)
	}
	if userID != user.ID {
		t.Fatalf("expected userID=%d, got %d", user.ID, userID)
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
	s.skillInstaller = skills.NewSkillInstaller(s.workspace)
	content := "---\nname: demo-skill\ndescription: Test skill\n---\n\n# Demo Skill\n\n## When to use\nUse this.\n"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/create", strings.NewReader(`{"name":"demo-skill","content":`+strconvQuote(content)+`}`))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()

	s.handleSkillAction(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var user *storage.User
	if s.centralStore != nil {
		user, _ = s.centralStore.GetUserByUsername("admin")
	} else {
		user, _ = s.store.GetUserByUsername("admin")
	}
	if user == nil {
		t.Fatal("admin user not found")
	}
	// Skills are now written to the user-specific skills dir (~/.MakoClaw/users/{uuid}/skills/)
	home, _ := os.UserHomeDir()
	userSkillsDir := filepath.Join(home, ".MakoClaw", "users", user.UUID, "skills")
	skillPath := filepath.Join(userSkillsDir, "demo-skill", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("expected skill file to exist: %v", err)
	}
	// Clean up the created skill
	t.Cleanup(func() { os.RemoveAll(filepath.Join(userSkillsDir, "demo-skill")) })
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

func TestHandleChatSessionsListEmpty(t *testing.T) {
	s := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions", nil)
	req = withTestUserContext(t, s, req)
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

	// Get the per-user store for admin (same path getUserStorage would resolve)
	adminReq := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions", nil)
	adminReq = withTestUserContext(t, s, adminReq)
	userStore, _, ok := s.getUserStorage(adminReq)
	if !ok {
		t.Fatal("could not get user store")
	}
	_ = userStore.SaveMessage("web:sess1", "user", "hello")
	_ = userStore.SaveMessage("web:sess1", "assistant", "hi")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions", nil)
	req = withTestUserContext(t, s, req)
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

	// Get per-user store for admin
	setupReq := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions", nil)
	setupReq = withTestUserContext(t, s, setupReq)
	userStore, _, ok := s.getUserStorage(setupReq)
	if !ok {
		t.Fatal("could not get user store")
	}
	_ = userStore.SaveMessage("active:x", "user", "active")
	_ = userStore.SaveMessage("archived:x", "user", "archived")
	archivedTrue := true
	_, _ = userStore.UpdateSession("archived:x", nil, &archivedTrue)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions?archived=false", nil)
	req = withTestUserContext(t, s, req)
	s.handleChatSessions(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if strings.Contains(rr.Body.String(), "archived:x") {
		t.Fatalf("archived session should not appear: %s", rr.Body.String())
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions?archived=true", nil)
	req2 = withTestUserContext(t, s, req2)
	s.handleChatSessions(rr2, req2)
	if !strings.Contains(rr2.Body.String(), "archived:x") {
		t.Fatalf("expected archived session: %s", rr2.Body.String())
	}
}

func TestHandleChatSessionMessagesGet(t *testing.T) {
	s := newTestServer(t)

	setupReq := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions/msgs:test", nil)
	setupReq = withTestUserContext(t, s, setupReq)
	userStore, _, ok := s.getUserStorage(setupReq)
	if !ok {
		t.Fatal("could not get user store")
	}
	_ = userStore.SaveMessage("msgs:test", "user", "hello msg")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions/msgs:test", nil)
	req = withTestUserContext(t, s, req)
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

	setupReq := httptest.NewRequest(http.MethodPatch, "/api/v1/chat/sessions/patch:test", nil)
	setupReq = withTestUserContext(t, s, setupReq)
	userStore, _, ok := s.getUserStorage(setupReq)
	if !ok {
		t.Fatal("could not get user store")
	}
	_ = userStore.SaveMessage("patch:test", "user", "msg")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/chat/sessions/patch:test", strings.NewReader(`{"title":"My Title"}`))
	req = withTestUserContext(t, s, req)
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

	setupReq := httptest.NewRequest(http.MethodDelete, "/api/v1/chat/sessions/del:test", nil)
	setupReq = withTestUserContext(t, s, setupReq)
	userStore, _, ok := s.getUserStorage(setupReq)
	if !ok {
		t.Fatal("could not get user store")
	}
	_ = userStore.SaveMessage("del:test", "user", "bye")

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/chat/sessions/del:test", nil)
	req = withTestUserContext(t, s, req)
	s.handleChatSessionMessages(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	msgs, _ := userStore.GetMessages("del:test")
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages after delete, got %d", len(msgs))
	}
}

func TestHandleChatSessionMessagesNoID(t *testing.T) {
	s := newTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/sessions/", nil)
	req = withTestUserContext(t, s, req)
	s.handleChatSessionMessages(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// ==================== Standing Orders Tests ====================

func TestHandleStandingOrdersListEmpty(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/standing-orders", nil)
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleStandingOrders(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if _, ok := out["standing_orders"]; !ok {
		t.Fatal("expected 'standing_orders' key in response")
	}
}

func TestHandleStandingOrdersCreate(t *testing.T) {
	s := newTestServer(t)
	body := strings.NewReader(`{"content":"Always respond in English","label":"language"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/standing-orders", body)
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleStandingOrders(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var out map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if out["content"] != "Always respond in English" {
		t.Fatalf("unexpected content: %v", out["content"])
	}
}

func TestHandleStandingOrdersCreateEmptyContent(t *testing.T) {
	s := newTestServer(t)
	body := strings.NewReader(`{"content":"","label":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/standing-orders", body)
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleStandingOrders(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleStandingOrderDeleteAndNotFound(t *testing.T) {
	s := newTestServer(t)

	// Create one
	createBody := strings.NewReader(`{"content":"Be concise"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/standing-orders", createBody)
	createReq = withTestUserContext(t, s, createReq)
	createRR := httptest.NewRecorder()
	s.handleStandingOrders(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", createRR.Code, createRR.Body.String())
	}
	var created map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	id := int64(created["id"].(float64))
	idStr := strconv.FormatInt(id, 10)

	// Delete it
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/standing-orders/"+idStr, nil)
	delReq = withTestUserContext(t, s, delReq)
	delRR := httptest.NewRecorder()
	s.handleStandingOrderAction(delRR, delReq)
	if delRR.Code != http.StatusOK {
		t.Fatalf("expected 200 delete, got %d: %s", delRR.Code, delRR.Body.String())
	}

	// Delete again - should be 404
	delReq2 := httptest.NewRequest(http.MethodDelete, "/api/v1/standing-orders/"+idStr, nil)
	delReq2 = withTestUserContext(t, s, delReq2)
	delRR2 := httptest.NewRecorder()
	s.handleStandingOrderAction(delRR2, delReq2)
	if delRR2.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", delRR2.Code)
	}
}

func TestHandleStandingOrdersRequiresAuth(t *testing.T) {
	s := newTestServer(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := s.authMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/standing-orders", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", rr.Code)
	}
}
