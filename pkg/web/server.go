package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/ratelimit"
)

//go:embed static/*
var staticFS embed.FS

type Server struct {
	cfg          config.WebConfig
	workspace    string
	agentLoop    *agent.AgentLoop
	server       *http.Server
	authManager  *authManager
	tasks        *taskStore
	loginLimit   *ratelimit.RateLimiter
	tasksMu      sync.RWMutex
	tasksClients map[*websocket.Conn]struct{}
	connMu       map[*websocket.Conn]*sync.Mutex
}

type taskItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	Result      string    `json:"result,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: checkWebSocketOrigin,
}

func NewServer(cfg config.WebConfig, agentLoop *agent.AgentLoop) *Server {
	workspace := defaultWorkspace()
	return &Server{
		cfg:          cfg,
		workspace:    workspace,
		agentLoop:    agentLoop,
		loginLimit:   ratelimit.NewRateLimiter(),
		tasksClients: make(map[*websocket.Conn]struct{}),
		connMu:       make(map[*websocket.Conn]*sync.Mutex),
	}
}

func NewServerWithWorkspace(cfg config.WebConfig, agentLoop *agent.AgentLoop, workspace string) *Server {
	if strings.TrimSpace(workspace) == "" {
		workspace = defaultWorkspace()
	}
	s := NewServer(cfg, agentLoop)
	s.workspace = workspace
	return s
}

func defaultWorkspace() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return filepath.Join(home, ".picoclaw", "workspace")
}

func (s *Server) Start(ctx context.Context) error {
	dataDir := filepath.Join(s.workspace, "web")
	authManager, err := newAuthManager(dataDir, s.cfg.Username, s.cfg.Password, s.cfg.JWTExpiry)
	if err != nil {
		return err
	}
	s.authManager = authManager
	taskStore, err := newTaskStore(dataDir)
	if err != nil {
		return err
	}
	s.tasks = taskStore

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", s.handleHealth)
	mux.HandleFunc("/api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("/api/v1/auth/change-password", s.handleChangePassword)
	mux.HandleFunc("/api/v1/auth/me", s.handleMe)
	mux.HandleFunc("/api/v1/tasks", s.handleTasks)
	mux.HandleFunc("/api/v1/tasks/", s.handleTasks)
	mux.HandleFunc("/ws/chat", s.handleChatWS)
	mux.HandleFunc("/ws/tasks", s.handleTasksWS)
	mux.Handle("/", s.staticHandler())

	s.server = &http.Server{
		Addr:    s.cfg.Host + ":" + toString(s.cfg.Port),
		Handler: s.authMiddleware(mux),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			logger.ErrorCF("web", "Web server shutdown error", map[string]interface{}{"error": err.Error()})
		}
	}()

	go func() {
		logger.InfoCF("web", "Web server starting", map[string]interface{}{
			"addr":    s.server.Addr,
			"enabled": s.cfg.Enabled,
		})
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorCF("web", "Web server error", map[string]interface{}{"error": err.Error()})
		}
	}()

	if s.agentLoop != nil {
		go s.runTaskWorker(ctx)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.tasksMu.Lock()
	for conn := range s.tasksClients {
		_ = conn.Close()
		delete(s.tasksClients, conn)
		delete(s.connMu, conn)
	}
	s.tasksMu.Unlock()

	if s.tasks != nil {
		_ = s.tasks.close()
	}
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func (s *Server) staticHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return http.NotFoundHandler()
	}
	return http.FileServer(http.FS(sub))
}

const maxWebSocketClients = 100

func (s *Server) authMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; connect-src 'self' ws: wss:")

		if strings.HasPrefix(r.URL.Path, "/api/v1/auth/login") {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/v1/health") {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws/") {
			if !s.isAuthorized(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "unauthorized",
				})
				return
			}
		}
		next.ServeHTTP(w, r)
	}
}

func (s *Server) isAuthorized(r *http.Request) bool {
	if s.authManager == nil {
		return false
	}

	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token := strings.TrimSpace(authHeader[7:])
		if _, err := s.authManager.verifyToken(token); err == nil {
			return true
		}
	}
	if strings.HasPrefix(r.URL.Path, "/ws/") {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if _, err := s.authManager.verifyToken(token); err == nil {
			return true
		}
	}

	return false
}

func checkWebSocketOrigin(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, r.Host)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if s.tasks == nil {
		http.Error(w, "tasks store unavailable", http.StatusServiceUnavailable)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			tasks, err := s.tasks.list()
			if err != nil {
				http.Error(w, "failed to list tasks", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
		case http.MethodPost:
			var in taskItem
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "invalid json", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(in.Title) == "" {
				http.Error(w, "title is required", http.StatusBadRequest)
				return
			}
			status, ok := normalizeTaskStatus(in.Status)
			if !ok {
				http.Error(w, "invalid status", http.StatusBadRequest)
				return
			}
			created, err := s.tasks.create(strings.TrimSpace(in.Title), strings.TrimSpace(in.Description), status)
			if err != nil {
				http.Error(w, "failed to create task", http.StatusInternalServerError)
				return
			}
			s.broadcastTaskEvent("created", created)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if strings.HasSuffix(path, "/status") {
		id := strings.TrimSuffix(path, "/status")
		id = strings.TrimSuffix(id, "/")
		if id == "" || r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		status, ok := normalizeTaskStatus(in.Status)
		if !ok {
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}
		updated, err := s.tasks.updateStatus(id, status)
		if err != nil {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		s.broadcastTaskEvent("status_changed", updated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
		return
	}
	if strings.HasSuffix(path, "/logs") {
		id := strings.TrimSuffix(path, "/logs")
		id = strings.TrimSuffix(id, "/")
		if id == "" || r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		logs, err := s.tasks.listLogs(id)
		if err != nil {
			http.Error(w, "failed to list logs", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"logs": logs})
		return
	}

	id := strings.TrimSpace(path)
	if id == "" {
		http.Error(w, "task id required", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPut:
		var in taskItem
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		title := strings.TrimSpace(in.Title)
		if title == "" {
			http.Error(w, "title is required", http.StatusBadRequest)
			return
		}
		status, ok := normalizeTaskStatus(in.Status)
		if !ok {
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}
		updated, err := s.tasks.update(id, title, strings.TrimSpace(in.Description), status, strings.TrimSpace(in.Result))
		if err != nil {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		s.broadcastTaskEvent("updated", updated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
	case http.MethodDelete:
		if err := s.tasks.delete(id); err != nil {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		s.broadcastTaskEvent("deleted", taskItem{ID: id})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type chatResponse struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (s *Server) handleChatWS(w http.ResponseWriter, r *http.Request) {
	if s.agentLoop == nil {
		http.Error(w, "agent loop unavailable", http.StatusServiceUnavailable)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}

		input := strings.TrimSpace(string(raw))
		if input == "" {
			continue
		}

		if handled, response := s.handleTaskChatCommand(input); handled {
			_ = conn.WriteJSON(chatResponse{Role: "assistant", Content: response})
			continue
		}

		response, err := s.agentLoop.ProcessDirect(r.Context(), input, "web:chat")
		if err != nil {
			_ = conn.WriteJSON(chatResponse{Role: "system", Content: "error: " + err.Error()})
			continue
		}
		_ = conn.WriteJSON(chatResponse{Role: "assistant", Content: response})
	}
}

func (s *Server) handleTaskChatCommand(input string) (bool, string) {
	if s.tasks == nil {
		return false, ""
	}
	lower := strings.ToLower(strings.TrimSpace(input))
	if strings.HasPrefix(lower, "/task create ") {
		title := strings.TrimSpace(input[len("/task create "):])
		if title == "" {
			return true, "Uso: /task create <titulo>"
		}
		created, err := s.tasks.create(title, "", "todo")
		if err != nil {
			return true, "No pude crear la tarea."
		}
		s.broadcastTaskEvent("created", created)
		return true, fmt.Sprintf("Tarea creada en todo: %s (%s)", created.Title, created.ID)
	}
	if lower == "/task list" {
		tasks, err := s.tasks.list()
		if err != nil {
			return true, "No pude listar tareas."
		}
		if len(tasks) == 0 {
			return true, "No hay tareas."
		}
		lines := make([]string, 0, len(tasks))
		for i, t := range tasks {
			if i >= 10 {
				break
			}
			lines = append(lines, fmt.Sprintf("- [%s] %s (%s)", t.Status, t.Title, t.ID))
		}
		return true, strings.Join(lines, "\n")
	}
	if strings.HasPrefix(lower, "/task run ") {
		id := strings.TrimSpace(input[len("/task run "):])
		if id == "" {
			return true, "Uso: /task run <id>"
		}
		updated, err := s.tasks.updateStatus(id, "todo")
		if err != nil {
			return true, "No encontré esa tarea."
		}
		s.broadcastTaskEvent("status_changed", updated)
		return true, "Tarea en cola para ejecución."
	}
	if strings.HasPrefix(lower, "/task move ") {
		fields := strings.Fields(strings.TrimSpace(input))
		if len(fields) != 4 {
			return true, "Uso: /task move <id> <status>"
		}
		id := strings.TrimSpace(fields[2])
		status, ok := normalizeTaskStatus(fields[3])
		if !ok {
			return true, "Estado inválido. Usa: backlog|todo|in_progress|review|done"
		}
		updated, err := s.tasks.updateStatus(id, status)
		if err != nil {
			return true, "No encontré esa tarea."
		}
		s.broadcastTaskEvent("status_changed", updated)
		return true, "Tarea movida a " + status + "."
	}
	return false, ""
}

func (s *Server) handleTasksWS(w http.ResponseWriter, r *http.Request) {
	s.tasksMu.RLock()
	clientCount := len(s.tasksClients)
	s.tasksMu.RUnlock()
	if clientCount >= maxWebSocketClients {
		http.Error(w, "too many connections", http.StatusServiceUnavailable)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	mu := &sync.Mutex{}
	s.tasksMu.Lock()
	s.tasksClients[conn] = struct{}{}
	s.connMu[conn] = mu
	s.tasksMu.Unlock()

	defer func() {
		s.tasksMu.Lock()
		delete(s.tasksClients, conn)
		delete(s.connMu, conn)
		s.tasksMu.Unlock()
		_ = conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.authManager == nil {
		http.Error(w, "auth unavailable", http.StatusServiceUnavailable)
		return
	}
	ip := clientIP(r)
	key := "login:" + ip
	s.loginLimit.SetLimit(key, 5, time.Minute)
	if !s.loginLimit.Allow(key) {
		http.Error(w, "too many login attempts", http.StatusTooManyRequests)
		return
	}
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	token, err := s.authManager.login(in.Username, in.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.authManager == nil {
		http.Error(w, "auth unavailable", http.StatusServiceUnavailable)
		return
	}
	ip := clientIP(r)
	cpKey := "change-password:" + ip
	s.loginLimit.SetLimit(cpKey, 5, time.Minute)
	if !s.loginLimit.Allow(cpKey) {
		http.Error(w, "too many attempts", http.StatusTooManyRequests)
		return
	}
	var in struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := s.authManager.changePassword(in.OldPassword, in.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleMe(w http.ResponseWriter, _ *http.Request) {
	if s.authManager == nil {
		http.Error(w, "auth unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"username": s.authManager.state.Username})
}

func toString(v int) string {
	return fmt.Sprintf("%d", v)
}

func clientIP(r *http.Request) string {
	forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	hostPort := strings.TrimSpace(r.RemoteAddr)
	if hostPort == "" {
		return "unknown"
	}
	return hostPort
}

func normalizeTaskStatus(status string) (string, bool) {
	s := strings.TrimSpace(strings.ToLower(status))
	if s == "" {
		return "backlog", true
	}
	switch s {
	case "backlog", "todo", "in_progress", "review", "done":
		return s, true
	default:
		return "", false
	}
}

func (s *Server) broadcastTaskEvent(event string, task taskItem) {
	payload, err := json.Marshal(map[string]interface{}{
		"event": event,
		"task":  task,
	})
	if err != nil {
		return
	}

	s.tasksMu.RLock()
	type connWithMu struct {
		conn *websocket.Conn
		mu   *sync.Mutex
	}
	conns := make([]connWithMu, 0, len(s.tasksClients))
	for c := range s.tasksClients {
		if mu, ok := s.connMu[c]; ok {
			conns = append(conns, connWithMu{conn: c, mu: mu})
		}
	}
	s.tasksMu.RUnlock()

	for _, cm := range conns {
		cm.mu.Lock()
		err := cm.conn.WriteMessage(websocket.TextMessage, payload)
		cm.mu.Unlock()
		if err != nil {
			s.tasksMu.Lock()
			delete(s.tasksClients, cm.conn)
			delete(s.connMu, cm.conn)
			s.tasksMu.Unlock()
			_ = cm.conn.Close()
		}
	}
}

func (s *Server) runTaskWorker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processNextTodoTask(ctx)
		}
	}
}

func (s *Server) processNextTodoTask(ctx context.Context) {
	tasks, err := s.tasks.list()
	if err != nil {
		logger.WarnCF("web", "task worker: failed to list tasks", map[string]interface{}{"error": err.Error()})
		return
	}
	for _, t := range tasks {
		if t.Status != "todo" {
			continue
		}
		inProgress, err := s.tasks.updateStatus(t.ID, "in_progress")
		if err != nil {
			logger.WarnCF("web", "task worker: failed to update status", map[string]interface{}{"task_id": t.ID, "error": err.Error()})
			continue
		}
		s.broadcastTaskEvent("status_changed", inProgress)

		prompt := "Ejecuta esta tarea y devuelve un resumen breve.\nTitulo: " + t.Title + "\nDescripcion: " + t.Description
		result, err := s.agentLoop.ProcessDirect(ctx, prompt, "web:task:"+t.ID)
		if err != nil {
			updated, updateErr := s.tasks.update(t.ID, t.Title, t.Description, "review", "error: "+err.Error())
			if updateErr != nil {
				logger.WarnCF("web", "task worker: failed to save error result", map[string]interface{}{"task_id": t.ID, "error": updateErr.Error()})
			}
			if logErr := s.tasks.addLog(t.ID, "worker_error", err.Error()); logErr != nil {
				logger.WarnCF("web", "task worker: failed to add error log", map[string]interface{}{"task_id": t.ID, "error": logErr.Error()})
			}
			s.broadcastTaskEvent("updated", updated)
			s.broadcastTaskEvent("status_changed", updated)
			continue
		}
		updated, err := s.tasks.update(t.ID, t.Title, t.Description, "review", result)
		if err != nil {
			logger.WarnCF("web", "task worker: failed to save result", map[string]interface{}{"task_id": t.ID, "error": err.Error()})
			continue
		}
		if logErr := s.tasks.addLog(t.ID, "worker_completed", "moved to review"); logErr != nil {
			logger.WarnCF("web", "task worker: failed to add log", map[string]interface{}{"task_id": t.ID, "error": logErr.Error()})
		}
		s.broadcastTaskEvent("updated", updated)
		s.broadcastTaskEvent("status_changed", updated)
		return
	}
}
