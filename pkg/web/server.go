package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"

	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sipeed/kakoclaw/pkg/agent"
	"github.com/sipeed/kakoclaw/pkg/channels"
	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/cron"
	"github.com/sipeed/kakoclaw/pkg/logger"
	"github.com/sipeed/kakoclaw/pkg/mcp"
	"github.com/sipeed/kakoclaw/pkg/observability"
	"github.com/sipeed/kakoclaw/pkg/providers"
	"github.com/sipeed/kakoclaw/pkg/ratelimit"
	"github.com/sipeed/kakoclaw/pkg/skills"
	"github.com/sipeed/kakoclaw/pkg/storage"
	"github.com/sipeed/kakoclaw/pkg/voice"
	"github.com/sipeed/kakoclaw/pkg/workflow"
)

//go:embed dist/*
var staticFS embed.FS

// activeExecution tracks a running agent execution with cancellation
type activeExecution struct {
	SessionID string
	StartedAt time.Time
	Cancel    context.CancelFunc
}

type Server struct {
	cfg            config.WebConfig
	fullConfig     *config.Config
	workspace      string
	agentLoop      *agent.AgentLoop
	server         *http.Server
	authManager    *authManager
	store          *storage.Storage
	loginLimit     *ratelimit.RateLimiter
	tasksMu        sync.RWMutex
	tasksClients   map[*websocket.Conn]struct{}
	connMu         map[*websocket.Conn]*sync.Mutex
	memory         *agent.MemoryStore
	cronService    *cron.CronService
	skillsLoader   *skills.SkillsLoader
	skillInstaller *skills.SkillInstaller
	channelManager *channels.Manager
	transcriber    *voice.GroqTranscriber
	mcpManager     *mcp.Manager
	workflowEngine *workflow.Engine
	execMu         sync.RWMutex
	activeExecs    map[string]*activeExecution
}

type taskItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	Result      string    `json:"result,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Archived    bool      `json:"archived"` // Added Archived field
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: checkWebSocketOrigin,
}

func NewServer(cfg config.WebConfig, agentLoop *agent.AgentLoop, store *storage.Storage) *Server {
	workspace := defaultWorkspace()
	return &Server{
		cfg:          cfg,
		workspace:    workspace,
		agentLoop:    agentLoop,
		store:        store,
		loginLimit:   ratelimit.NewRateLimiter(),
		tasksClients: make(map[*websocket.Conn]struct{}),
		connMu:       make(map[*websocket.Conn]*sync.Mutex),
		memory:       agent.NewMemoryStore(workspace),
		activeExecs:  make(map[string]*activeExecution),
	}
}

func NewServerWithWorkspace(cfg config.WebConfig, agentLoop *agent.AgentLoop, workspace string) *Server {
	if strings.TrimSpace(workspace) == "" {
		workspace = defaultWorkspace()
	}
	s := NewServer(cfg, agentLoop, nil)
	s.workspace = workspace
	return s
}

// SetStorage allows setting storage after initialization (helper for legacy calls)
func (s *Server) SetStorage(store *storage.Storage) {
	s.store = store
}

// SetCronService injects the cron service for REST exposure
func (s *Server) SetCronService(cs *cron.CronService) {
	s.cronService = cs
}

// SetSkills injects skills loader and installer for REST exposure
func (s *Server) SetSkills(loader *skills.SkillsLoader, installer *skills.SkillInstaller) {
	s.skillsLoader = loader
	s.skillInstaller = installer
}

// SetChannelManager injects the channel manager for REST exposure
func (s *Server) SetChannelManager(cm *channels.Manager) {
	s.channelManager = cm
}

// SetTranscriber injects the Groq voice transcriber for REST exposure
func (s *Server) SetTranscriber(t *voice.GroqTranscriber) {
	s.transcriber = t
}

// SetFullConfig injects the full config for read-only settings endpoint
func (s *Server) SetFullConfig(cfg *config.Config) {
	s.fullConfig = cfg
}

// SetMCPManager injects the MCP manager for REST exposure
func (s *Server) SetMCPManager(m *mcp.Manager) {
	s.mcpManager = m
}

// SetWorkflowEngine injects the workflow engine for REST exposure
func (s *Server) SetWorkflowEngine(e *workflow.Engine) {
	s.workflowEngine = e
}

func defaultWorkspace() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return filepath.Join(home, ".KakoClaw", "workspace")
}

func (s *Server) Start(ctx context.Context) error {
	authManager, err := newAuthManager(s.store, s.cfg.Username, s.cfg.Password, s.cfg.JWTExpiry)
	if err != nil {
		return err
	}
	s.authManager = authManager

	// Ensure storage is available
	if s.store == nil {
		// Fallback to internal storage if not provided (should be provided by main)
		// For now, we assume it's passed or set. If not, we log warning.
		logger.WarnC("web", "Storage not provided to web server, some features may be disabled")
	} else {
		// Wire persistent storage into the metrics singleton so that counters survive restarts.
		observability.Global().SetStorage(s.store)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", s.handleHealth)
	mux.HandleFunc("/api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("/api/v1/auth/change-password", s.handleChangePassword)
	mux.HandleFunc("/api/v1/auth/me", s.handleMe)
	mux.HandleFunc("/api/v1/users", s.handleUsers)       // User management (Admin only)
	mux.HandleFunc("/api/v1/users/", s.handleUserAction) // User actions
	mux.HandleFunc("/api/v1/tasks", s.handleTasks)
	mux.HandleFunc("/api/v1/tasks/search", s.handleTaskSearch) // Search tasks
	mux.HandleFunc("/api/v1/tasks/", s.handleTasks)
	mux.HandleFunc("/api/v1/chat/sessions", s.handleChatSessions)             // New endpoint
	mux.HandleFunc("/api/v1/chat/sessions/", s.handleChatSessionMessages)     // New endpoint
	mux.HandleFunc("/api/v1/chat/search", s.handleChatSearch)                 // Search messages
	mux.HandleFunc("/api/v1/chat/fork", s.handleChatFork)                     // Fork conversation
	mux.HandleFunc("/api/v1/chat/cancel", s.handleChatCancel)                 // Cancel execution
	mux.HandleFunc("/api/v1/chat/active", s.handleChatActive)                 // Active executions
	mux.HandleFunc("/api/v1/memory/longterm", s.handleLongTermMemory)         // New endpoint
	mux.HandleFunc("/api/v1/memory/daily", s.handleDailyNotes)                // New endpoint
	mux.HandleFunc("/api/v1/skills", s.handleSkills)                          // Skills list + marketplace
	mux.HandleFunc("/api/v1/skills/", s.handleSkillAction)                    // Install/uninstall/view
	mux.HandleFunc("/api/v1/cron", s.handleCron)                              // Cron jobs list + create
	mux.HandleFunc("/api/v1/cron/", s.handleCronAction)                       // Cron job actions
	mux.HandleFunc("/api/v1/channels", s.handleChannels)                      // Channels status
	mux.HandleFunc("/api/v1/config", s.handleConfig)                          // Config (read-only, redacted)
	mux.HandleFunc("/api/v1/files", s.handleFiles)                            // File browser
	mux.HandleFunc("/api/v1/files/", s.handleFiles)                           // File browser subpaths
	mux.HandleFunc("/api/v1/export/tasks", s.handleExportTasks)               // Export tasks
	mux.HandleFunc("/api/v1/export/chat", s.handleExportChat)                 // Export chat history
	mux.HandleFunc("/api/v1/import/chat", s.handleImportChat)                 // Import conversations
	mux.HandleFunc("/api/v1/models", s.handleModels)                          // Available models/providers
	mux.HandleFunc("/api/v1/voice/transcribe", s.handleVoiceTranscribe)       // Voice-to-text (Groq STT)
	mux.HandleFunc("/api/v1/knowledge", s.handleKnowledge)                    // Knowledge base: list + upload
	mux.HandleFunc("/api/v1/knowledge/search", s.handleKnowledgeSearch)       // Knowledge base: FTS5 search
	mux.HandleFunc("/api/v1/knowledge/chunks/", s.handleKnowledgeChunkAction) // Knowledge base: update chunks
	mux.HandleFunc("/api/v1/knowledge/", s.handleKnowledgeAction)             // Knowledge base: view chunks or delete by ID
	mux.HandleFunc("/api/v1/openapi.json", s.handleOpenAPISpec)               // OpenAPI 3.0 spec (JSON)
	mux.HandleFunc("/api/docs", s.handleAPIDocsUI)                            // Swagger UI
	mux.HandleFunc("/api/v1/mcp", s.handleMCPServers)                         // MCP servers: list + status
	mux.HandleFunc("/api/v1/mcp/", s.handleMCPServerAction)                   // MCP server actions: reconnect
	mux.HandleFunc("/api/v1/metrics", s.handleMetrics)                        // Observability metrics
	mux.HandleFunc("/api/v1/tools", s.handleToolsList)                        // Available tools list
	mux.HandleFunc("/api/v1/prompts", s.handlePrompts)                        // Prompt templates: list + create
	mux.HandleFunc("/api/v1/prompts/", s.handlePromptAction)                  // Prompt templates: update/delete
	mux.HandleFunc("/api/v1/chat/attachments", s.handleChatAttachment)        // Chat file upload/extract
	mux.HandleFunc("/api/v1/workflows", s.handleWorkflows)                    // Workflows: list + create
	mux.HandleFunc("/api/v1/workflows/", s.handleWorkflowAction)              // Workflow actions: get/update/delete/run
	mux.HandleFunc("/api/v1/backup/export", s.handleBackupExport)             // Export backup
	mux.HandleFunc("/api/v1/backup/import", s.handleBackupImport)             // Import backup
	mux.HandleFunc("/api/v1/backup/validate", s.handleBackupValidate)         // Validate backup
	mux.HandleFunc("/ws/chat", s.handleChatWS)
	mux.HandleFunc("/ws/tasks", s.handleTasksWS)
	mux.Handle("/", s.staticHandler())

	s.server = &http.Server{
		Addr:    s.cfg.Host + ":" + toString(int64(s.cfg.Port)),
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

	if s.agentLoop != nil && s.store != nil {
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

	// Storage close is handled by agent loop or main
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func (s *Server) staticHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the requested file
		sub, err := fs.Sub(staticFS, "dist")
		if err != nil {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" || path == "/" {
			path = "index.html"
		}

		f, err := sub.Open(path)
		if err == nil {
			defer f.Close()
			stat, _ := f.Stat()

			// Get content type
			contentType := mime.TypeByExtension(filepath.Ext(path))
			if contentType == "" {
				// Fallbacks for things that might not be in standard Windows/Linux mime db
				switch filepath.Ext(path) {
				case ".woff2":
					contentType = "font/woff2"
				case ".webmanifest":
					contentType = "application/manifest+json"
				case ".svg":
					contentType = "image/svg+xml"
				default:
					contentType = "application/octet-stream"
				}
			}

			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
			// Service worker must not be cached; manifest should be short-lived
			if strings.HasSuffix(path, "sw.js") || strings.HasSuffix(path, ".webmanifest") {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			} else {
				w.Header().Set("Cache-Control", "public, max-age=3600")
			}
			w.WriteHeader(http.StatusOK)
			io.Copy(w, f)
			return
		}

		// If not found and it's not an API route, serve index.html for SPA routing
		if !strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/ws/") {
			f, err := sub.Open("index.html")
			if err == nil {
				defer f.Close()
				stat, _ := f.Stat()
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
				w.WriteHeader(http.StatusOK)
				io.Copy(w, f)
				return
			}
		}

		http.NotFound(w, r)
	})
}

const maxWebSocketClients = 100

type contextKey string

const userClaimsKey contextKey = "userClaims"

func (s *Server) authMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net blob:; style-src 'self' 'unsafe-inline'; connect-src 'self' ws: wss: https://cdn.jsdelivr.net; worker-src 'self' blob:")

		if strings.HasPrefix(r.URL.Path, "/api/v1/auth/login") {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/v1/health") {
			next.ServeHTTP(w, r)
			return
		}
		// API docs are public (Swagger UI needs to load without auth)
		if r.URL.Path == "/api/docs" || r.URL.Path == "/api/v1/openapi.json" {
			// Relax CSP for Swagger UI page to load external scripts/styles
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://unpkg.com; style-src 'self' 'unsafe-inline' https://unpkg.com; connect-src 'self' ws: wss:; img-src 'self' data: https://unpkg.com")
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws/") {
			claims, ok := s.extractClaims(r)
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "unauthorized",
				})
				return
			}
			// Attach claims to context
			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	}
}

func (s *Server) isAuthorized(r *http.Request) bool {
	_, ok := s.extractClaims(r)
	return ok
}

func (s *Server) extractClaims(r *http.Request) (*jwtClaims, bool) {
	if s.authManager == nil {
		return nil, false
	}

	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token := strings.TrimSpace(authHeader[7:])
		if claims, err := s.authManager.verifyToken(token); err == nil {
			return claims, true
		}
	}
	if strings.HasPrefix(r.URL.Path, "/ws/") || strings.HasPrefix(r.URL.Path, "/api/") {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if claims, err := s.authManager.verifyToken(token); err == nil && token != "" {
			return claims, true
		}
	}

	return nil, false
}

func checkWebSocketOrigin(r *http.Request) bool {
	// Allow all origins to support reverse proxies (Nginx/Caddy) where Host header
	// might differ from Origin. We already validate the JWT token in authMiddleware.
	return true
}

// getUserIDFromClaims extracts the user_id from the request's JWT claims.
// Returns user_id and true if found; otherwise 0 and false.
func (s *Server) getUserIDFromClaims(r *http.Request) (int64, bool) {
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil {
		return 0, false
	}
	if s.store == nil {
		return 0, false
	}
	user, err := s.store.GetUserByUsername(claims.Sub)
	if err != nil {
		return 0, false
	}
	return user.ID, true
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "tasks store unavailable", http.StatusServiceUnavailable)
		return
	}
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			includeArchived := r.URL.Query().Get("include_archived") == "true"
			tasks, err := s.store.ListTasksForUser(userID, includeArchived)
			if err != nil {
				http.Error(w, "failed to list tasks", http.StatusInternalServerError)
				return
			}

			items := make([]taskItem, 0, len(tasks))
			for _, t := range tasks {
				items = append(items, taskItem{
					ID:          toString(t.ID),
					Title:       t.Title,
					Description: t.Description,
					Status:      t.Status,
					Result:      t.Result,
					CreatedAt:   t.CreatedAt,
					Archived:    t.Archived,
				})
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"tasks": items})
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
			id, err := s.store.CreateTaskForUser(userID, strings.TrimSpace(in.Title), strings.TrimSpace(in.Description), status)
			if err != nil {
				http.Error(w, "failed to create task", http.StatusInternalServerError)
				return
			}

			// Fetch the created task to get full details (like created_at)
			created, err := s.store.GetTaskForUser(userID, id)
			if err != nil {
				http.Error(w, "failed to get created task", http.StatusInternalServerError)
				return
			}

			item := taskItem{
				ID:          toString(created.ID),
				Title:       created.Title,
				Description: created.Description,
				Status:      created.Status,
				Result:      created.Result,
				CreatedAt:   created.CreatedAt,
				Archived:    created.Archived,
			}

			s.broadcastTaskEvent("created", item)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(item)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if strings.HasSuffix(path, "/status") {
		idStr := strings.TrimSuffix(path, "/status")
		idStr = strings.TrimSuffix(idStr, "/")
		if idStr == "" || r.Method != http.MethodPatch {
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

		id, err := parseID(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		updated, err := s.store.UpdateTaskStatusForUser(userID, id, status)
		if err != nil {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		item := taskItem{
			ID:          toString(updated.ID),
			Title:       updated.Title,
			Description: updated.Description,
			Status:      updated.Status,
			Result:      updated.Result,
			CreatedAt:   updated.CreatedAt,
			Archived:    updated.Archived,
		}

		s.broadcastTaskEvent("status_changed", item)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(item)
		return
	}

	if strings.HasSuffix(path, "/archive") {
		idStr := strings.TrimSuffix(path, "/archive")
		idStr = strings.TrimSuffix(idStr, "/")
		if idStr == "" || r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id, err := parseID(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		if err := s.store.ArchiveTaskForUser(userID, id); err != nil {
			http.Error(w, "failed to archive task", http.StatusInternalServerError)
			return
		}

		task, err := s.store.GetTaskForUser(userID, id)
		if err == nil {
			item := taskItem{
				ID:          toString(task.ID),
				Title:       task.Title,
				Description: task.Description,
				Status:      task.Status,
				Result:      task.Result,
				CreatedAt:   task.CreatedAt,
				Archived:    task.Archived,
			}
			s.broadcastTaskEvent("updated", item)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	if strings.HasSuffix(path, "/unarchive") {
		idStr := strings.TrimSuffix(path, "/unarchive")
		idStr = strings.TrimSuffix(idStr, "/")
		if idStr == "" || r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id, err := parseID(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		if err := s.store.UnarchiveTaskForUser(userID, id); err != nil {
			http.Error(w, "failed to unarchive task", http.StatusInternalServerError)
			return
		}

		task, err := s.store.GetTaskForUser(userID, id)
		if err == nil {
			item := taskItem{
				ID:          toString(task.ID),
				Title:       task.Title,
				Description: task.Description,
				Status:      task.Status,
				Result:      task.Result,
				CreatedAt:   task.CreatedAt,
				Archived:    task.Archived,
			}
			s.broadcastTaskEvent("updated", item)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	// Task logs endpoint
	if strings.HasSuffix(path, "/logs") {
		logsIDStr := strings.TrimSuffix(path, "/logs")
		logsIDStr = strings.TrimSpace(logsIDStr)
		logsID, err := parseID(logsIDStr)
		if err != nil {
			http.Error(w, "invalid task id for logs", http.StatusBadRequest)
			return
		}
		logs, err := s.store.GetTaskLogs(logsID)
		if err != nil {
			http.Error(w, "failed to get task logs", http.StatusInternalServerError)
			return
		}
		if logs == nil {
			logs = []storage.TaskLog{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"logs": logs})
		return
	}

	idStr := strings.TrimSpace(path)
	if idStr == "" {
		http.Error(w, "task id required", http.StatusBadRequest)
		return
	}

	id, err := parseID(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
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

		updated, err := s.store.UpdateTaskForUser(
			userID,
			id,
			strings.TrimSpace(in.Title),
			strings.TrimSpace(in.Description),
			status,
			strings.TrimSpace(in.Result),
		)
		if err != nil {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		item := taskItem{
			ID:          toString(updated.ID),
			Title:       updated.Title,
			Description: updated.Description,
			Status:      updated.Status,
			Result:      updated.Result,
			CreatedAt:   updated.CreatedAt,
			Archived:    updated.Archived,
		}

		s.broadcastTaskEvent("updated", item)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(item)
	case http.MethodDelete:
		if err := s.store.DeleteTaskForUser(userID, id); err != nil {
			http.Error(w, "failed to delete task", http.StatusInternalServerError)
			return
		}
		s.broadcastTaskEvent("deleted", taskItem{ID: idStr})
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

	// Extract user_id once per connection
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Mutex for thread-safe WebSocket writes (streaming callback writes from same goroutine,
	// but we protect against concurrent request handling edge cases)
	var wsMu sync.Mutex

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var req struct {
			Type         string   `json:"type"`
			Content      string   `json:"content"`
			SessionID    string   `json:"session_id"`
			Model        string   `json:"model"`
			WebSearch    *bool    `json:"web_search"`    // legacy: nil = default (enabled), false = exclude web_search tool
			ExcludeTools []string `json:"exclude_tools"` // new: granular control
		}

		// Try to decode as JSON
		if err := json.Unmarshal(raw, &req); err != nil {
			// If not JSON, treat as raw string (legacy/simple support)
			req.Content = string(raw)
			req.SessionID = "web:chat" // Default session
		}

		input := strings.TrimSpace(req.Content)
		if input == "" {
			continue
		}

		sessionID := strings.TrimSpace(req.SessionID)
		if sessionID == "" {
			sessionID = "web:chat"
		}

		if handled, response := s.handleTaskChatCommand(userID, input); handled {
			wsMu.Lock()
			_ = conn.WriteJSON(chatResponse{Role: "assistant", Content: response})
			wsMu.Unlock()
			continue
		}

		// Build exclude-tools list based on client toggles
		var excludeTools []string
		if req.WebSearch != nil && !*req.WebSearch {
			excludeTools = append(excludeTools, "web_search")
		}
		if len(req.ExcludeTools) > 0 {
			excludeTools = append(excludeTools, req.ExcludeTools...)
		}

		func() {
			// Create cancelable context for this execution
			ctx, cancel := context.WithCancel(r.Context())
			execID := fmt.Sprintf("%s:%d", sessionID, time.Now().UnixNano())

			// Track active execution
			s.execMu.Lock()
			s.activeExecs[execID] = &activeExecution{
				SessionID: sessionID,
				StartedAt: time.Now(),
				Cancel:    cancel,
			}
			s.execMu.Unlock()

			// Cleanup after execution completes
			defer func(id string) {
				s.execMu.Lock()
				delete(s.activeExecs, id)
				s.execMu.Unlock()
				cancel()
			}(execID)

			// Use streaming if supported
			if s.agentLoop.SupportsStreaming() {
				// Send stream_start
				wsMu.Lock()
				_ = conn.WriteJSON(map[string]interface{}{"type": "stream_start"})
				wsMu.Unlock()

				response, err := s.agentLoop.ProcessDirectWithModelStream(
					ctx, input, sessionID, req.Model,
					func(token string) error {
						wsMu.Lock()
						defer wsMu.Unlock()
						return conn.WriteJSON(map[string]interface{}{
							"type":    "stream",
							"content": token,
						})
					},
					func(ev agent.ToolEvent) error {
						wsMu.Lock()
						defer wsMu.Unlock()
						return conn.WriteJSON(map[string]interface{}{
							"type":   "tool_call",
							"name":   ev.Name,
							"args":   ev.Args,
							"result": ev.Result,
							"status": ev.Status,
						})
					},
					excludeTools...,
				)

				if err != nil {
					errMsg := err.Error()
					if ctx.Err() == context.Canceled {
						errMsg = "Execution canceled by user"
					}
					wsMu.Lock()
					_ = conn.WriteJSON(map[string]interface{}{
						"type":    "stream_end",
						"content": "",
						"error":   errMsg,
					})
					_ = conn.WriteJSON(map[string]interface{}{"type": "ready"})
					wsMu.Unlock()
					return
				}

				wsMu.Lock()
				_ = conn.WriteJSON(map[string]interface{}{
					"type":    "stream_end",
					"content": response,
				})
				_ = conn.WriteJSON(map[string]interface{}{"type": "ready"})
				wsMu.Unlock()
			} else {
				// Non-streaming fallback
				response, err := s.agentLoop.ProcessDirectWithModel(ctx, input, sessionID, req.Model, excludeTools...)
				if err != nil {
					errMsg := err.Error()
					if ctx.Err() == context.Canceled {
						errMsg = "Execution canceled by user"
					}
					wsMu.Lock()
					_ = conn.WriteJSON(map[string]interface{}{
						"type":    "message",
						"role":    "system",
						"content": "error: " + errMsg,
					})
					_ = conn.WriteJSON(map[string]interface{}{"type": "ready"})
					wsMu.Unlock()
					return
				}
				wsMu.Lock()
				_ = conn.WriteJSON(map[string]interface{}{
					"type":    "message",
					"role":    "assistant",
					"content": response,
				})
				_ = conn.WriteJSON(map[string]interface{}{"type": "ready"})
				wsMu.Unlock()
			}
		}()
	}
}

func (s *Server) handleTaskChatCommand(userID int64, input string) (bool, string) {
	if s.store == nil {
		return false, ""
	}
	lower := strings.ToLower(strings.TrimSpace(input))
	if strings.HasPrefix(lower, "/task create ") {
		title := strings.TrimSpace(input[len("/task create "):])
		if title == "" {
			return true, "Uso: /task create <titulo>"
		}
		id, err := s.store.CreateTaskForUser(userID, title, "", "todo")
		if err != nil {
			return true, "No pude crear la tarea."
		}

		created, err := s.store.GetTaskForUser(userID, id)
		if err != nil {
			return true, "Tarea creada pero fallé al recuperarla via ID."
		}

		item := taskItem{
			ID:          toString(created.ID),
			Title:       created.Title,
			Description: created.Description,
			Status:      created.Status,
			Result:      created.Result,
			CreatedAt:   created.CreatedAt,
			Archived:    created.Archived,
		}

		s.broadcastTaskEvent("created", item)
		return true, fmt.Sprintf("Tarea creada en todo: %s (%s)", created.Title, toString(created.ID))
	}

	if lower == "/task list" || lower == "/list" {
		tasks, err := s.store.ListTasksForUser(userID, false)
		if err != nil {
			return true, "No pude listar tareas."
		}
		if len(tasks) == 0 {
			return true, "No hay tareas activas."
		}
		lines := make([]string, 0, len(tasks))
		for i, t := range tasks {
			if i >= 10 {
				break
			}
			lines = append(lines, fmt.Sprintf("- [%s] %s (%s)", t.Status, t.Title, toString(t.ID)))
		}
		return true, strings.Join(lines, "\n")
	}

	if strings.HasPrefix(lower, "/task run ") {
		idStr := strings.TrimSpace(input[len("/task run "):])
		if idStr == "" {
			return true, "Uso: /task run <id>"
		}
		id, err := parseID(idStr)
		if err != nil {
			return true, "ID inválido"
		}

		updated, err := s.store.UpdateTaskStatusForUser(userID, id, "todo")
		if err != nil {
			return true, "No encontré esa tarea."
		}

		item := taskItem{
			ID:          toString(updated.ID),
			Title:       updated.Title,
			Description: updated.Description,
			Status:      updated.Status,
			Result:      updated.Result,
			CreatedAt:   updated.CreatedAt,
			Archived:    updated.Archived,
		}

		s.broadcastTaskEvent("status_changed", item)
		return true, "Tarea en cola para ejecución."
	}

	if strings.HasPrefix(lower, "/task move ") {
		fields := strings.Fields(strings.TrimSpace(input))
		if len(fields) != 4 {
			return true, "Uso: /task move <id> <status>"
		}
		idStr := strings.TrimSpace(fields[2])
		id, err := parseID(idStr)
		if err != nil {
			return true, "ID inválido"
		}
		status, ok := normalizeTaskStatus(fields[3])
		if !ok {
			return true, "Estado inválido. Usa: backlog|todo|in_progress|review|done"
		}
		updated, err := s.store.UpdateTaskStatusForUser(userID, id, status)
		if err != nil {
			return true, "No encontré esa tarea."
		}

		item := taskItem{
			ID:          toString(updated.ID),
			Title:       updated.Title,
			Description: updated.Description,
			Status:      updated.Status,
			Result:      updated.Result,
			CreatedAt:   updated.CreatedAt,
			Archived:    updated.Archived,
		}

		s.broadcastTaskEvent("status_changed", item)
		return true, "Tarea movida a " + status + "."
	}

	if strings.HasPrefix(lower, "/archive ") {
		idStr := strings.TrimSpace(input[len("/archive "):])
		if idStr == "" {
			return true, "Uso: /archive <id>"
		}
		id, err := parseID(idStr)
		if err != nil {
			return true, "ID inválido"
		}
		if err := s.store.ArchiveTaskForUser(userID, id); err != nil {
			return true, "Error al archivar la tarea."
		}
		// Broadcast delete to frontend
		s.broadcastTaskEvent("deleted", taskItem{ID: idStr})
		return true, "Tarea archivada."
	}

	return false, ""
}

// handleChatCancel cancels a running agent execution
func (s *Server) handleChatCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	// Find and cancel all executions for this session
	s.execMu.Lock()
	canceled := 0
	for id, exec := range s.activeExecs {
		if exec.SessionID == req.SessionID {
			exec.Cancel()
			delete(s.activeExecs, id)
			canceled++
		}
	}
	s.execMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"canceled": canceled,
		"message":  fmt.Sprintf("Canceled %d execution(s)", canceled),
	})
}

// handleChatActive returns the list of active executions
func (s *Server) handleChatActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.execMu.RLock()
	active := make([]map[string]interface{}, 0, len(s.activeExecs))
	for id, exec := range s.activeExecs {
		active = append(active, map[string]interface{}{
			"id":         id,
			"session_id": exec.SessionID,
			"started_at": exec.StartedAt.Format(time.RFC3339),
			"duration":   time.Since(exec.StartedAt).String(),
		})
	}
	s.execMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(active)
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
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := s.authManager.changePassword(claims.Sub, in.OldPassword, in.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleTaskSearch(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}
	results, err := s.store.SearchTasksForUser(userID, q)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"tasks": results})
}

func (s *Server) handleChatSessions(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse optional query params: ?archived=true&limit=50&offset=0
	var archivedFilter *bool
	if archivedStr := r.URL.Query().Get("archived"); archivedStr != "" {
		v := archivedStr == "true" || archivedStr == "1"
		archivedFilter = &v
	}
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	sessions, err := s.store.ListSessionsForUser(userID, archivedFilter, limit, offset)
	if err != nil {
		http.Error(w, "failed to list sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"sessions": sessions})
}

func (s *Server) handleChatSessionMessages(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// /api/v1/chat/sessions/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/chat/sessions")
	id := strings.TrimPrefix(path, "/")
	id = strings.TrimSpace(id)

	if id == "" {
		http.Error(w, "session id required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		messages, err := s.store.GetMessagesForUser(userID, id)
		if err != nil {
			http.Error(w, "failed to get messages", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"messages": messages})

	case http.MethodDelete:
		if err := s.store.DeleteSessionForUser(userID, id); err != nil {
			http.Error(w, "failed to delete session", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	case http.MethodPatch:
		var payload struct {
			Title    *string `json:"title"`
			Archived *bool   `json:"archived"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if payload.Title == nil && payload.Archived == nil {
			http.Error(w, "nothing to update", http.StatusBadRequest)
			return
		}
		sess, err := s.store.UpdateSessionForUser(userID, id, payload.Title, payload.Archived)
		if err != nil {
			http.Error(w, "failed to update session", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sess)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleChatSearch(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}
	results, err := s.store.SearchMessagesForUser(userID, q)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"messages": results})
}

func (s *Server) handleChatFork(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.store == nil {
		http.Error(w, "store unavailable", http.StatusServiceUnavailable)
		return
	}
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		SessionID string `json:"session_id"`
		MessageID int64  `json:"message_id"` // 0 = fork all messages
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.SessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	newSessionID := fmt.Sprintf("web:chat:fork:%d", time.Now().UnixMilli())

	count, err := s.store.ForkSessionForUser(userID, payload.SessionID, newSessionID, payload.MessageID)
	if err != nil {
		http.Error(w, "fork failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":              true,
		"new_session_id":  newSessionID,
		"messages_copied": count,
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if s.authManager == nil {
		http.Error(w, "auth unavailable", http.StatusServiceUnavailable)
		return
	}
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"username": claims.Sub,
		"role":     claims.Role,
	})
}

func toString(v int64) string {
	return fmt.Sprintf("%d", v)
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
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
	// Map backend event to frontend type
	frontendType := event
	switch event {
	case "created":
		frontendType = "task_created"
	case "updated", "status_changed":
		frontendType = "task_updated"
	case "deleted":
		frontendType = "task_deleted"
	}

	payload, err := json.Marshal(map[string]interface{}{
		"type":    frontendType,
		"task":    task,
		"task_id": task.ID,
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
	tasks, err := s.store.ListAllUsersTasks(false)
	if err != nil {
		logger.WarnCF("web", "task worker: failed to list tasks", map[string]interface{}{"error": err.Error()})
		return
	}
	for _, t := range tasks {
		if t.Status != "todo" {
			continue
		}

		// Log: task picked up
		_ = s.store.AddTaskLog(t.ID, "started", "Task picked up by worker")

		// Move to in_progress
		inProgress, err := s.store.UpdateTaskStatusForUser(t.UserID, t.ID, "in_progress")
		if err != nil {
			_ = s.store.AddTaskLog(t.ID, "error", "Failed to move to in_progress: "+err.Error())
			logger.WarnCF("web", "task worker: failed to update status", map[string]interface{}{"task_id": t.ID, "error": err.Error()})
			continue
		}

		_ = s.store.AddTaskLog(t.ID, "status_changed", "Status changed to in_progress")

		itemInProgress := taskItem{
			ID:          toString(inProgress.ID),
			Title:       inProgress.Title,
			Description: inProgress.Description,
			Status:      inProgress.Status,
			Result:      inProgress.Result,
			CreatedAt:   inProgress.CreatedAt,
			Archived:    inProgress.Archived,
		}

		s.broadcastTaskEvent("status_changed", itemInProgress)

		// Execute task
		_ = s.store.AddTaskLog(t.ID, "executing", "Sending task to AI agent")
		prompt := "Ejecuta esta tarea y devuelve un resumen breve.\nTitulo: " + t.Title + "\nDescripcion: " + t.Description
		// Use a dedicated session for task worker to not mix with user chat
		taskSessionKey := "web:task:" + toString(t.ID)
		result, err := s.agentLoop.ProcessDirectWithUser(ctx, t.UserID, prompt, taskSessionKey)

		var finalStatus string
		var finalResult string

		if err != nil {
			finalStatus = "review"
			finalResult = "error: " + err.Error()
			_ = s.store.AddTaskLog(t.ID, "error", "Agent execution failed: "+err.Error())
		} else {
			finalStatus = "review"
			finalResult = result
			_ = s.store.AddTaskLog(t.ID, "completed", "Agent returned result successfully")
		}

		// Update result and status
		updated, updateErr := s.store.UpdateTaskForUser(t.UserID, t.ID, t.Title, t.Description, finalStatus, finalResult)
		if updateErr != nil {
			_ = s.store.AddTaskLog(t.ID, "error", "Failed to save result: "+updateErr.Error())
			logger.WarnCF("web", "task worker: failed to save result", map[string]interface{}{"task_id": t.ID, "error": updateErr.Error()})
			continue
		}

		_ = s.store.AddTaskLog(t.ID, "status_changed", "Status changed to review")

		itemUpdated := taskItem{
			ID:          toString(updated.ID),
			Title:       updated.Title,
			Description: updated.Description,
			Status:      updated.Status,
			Result:      updated.Result,
			CreatedAt:   updated.CreatedAt,
			Archived:    updated.Archived,
		}

		s.broadcastTaskEvent("updated", itemUpdated)
		s.broadcastTaskEvent("status_changed", itemUpdated)
		return
	}
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type modelInfo struct {
		ID       string `json:"id"`
		Provider string `json:"provider"`
	}

	type providerInfo struct {
		Name     string      `json:"name"`
		Enabled  bool        `json:"enabled"`
		Models   []modelInfo `json:"models"`
		IsActive bool        `json:"is_active"`
	}

	var providersList []providerInfo
	var currentModel string
	var currentProvider string

	getModels := func(providerName string, defaultModels []modelInfo, configuredModels []string) []modelInfo {
		if len(configuredModels) > 0 {
			custom := make([]modelInfo, 0, len(configuredModels))
			for _, m := range configuredModels {
				custom = append(custom, modelInfo{ID: m, Provider: providerName})
			}
			return custom
		}
		return defaultModels
	}

	if s.fullConfig != nil {
		currentModel = s.fullConfig.Agents.Defaults.Model
		currentProvider, _ = providers.GetProviderForModel(currentModel)

		// Anthropic
		if s.fullConfig.Providers.Anthropic.APIKey != "" || s.fullConfig.Providers.Anthropic.AuthMethod != "" {
			providersList = append(providersList, providerInfo{
				Name: "anthropic", Enabled: true,
				IsActive: currentProvider == "anthropic",
				Models: getModels("anthropic", []modelInfo{
					{ID: "claude-sonnet-4-20250514", Provider: "anthropic"},
					{ID: "claude-3-5-haiku-20241022", Provider: "anthropic"},
					{ID: "claude-3-5-sonnet-20241022", Provider: "anthropic"},
					{ID: "claude-3-haiku-20240307", Provider: "anthropic"},
				}, s.fullConfig.Providers.Anthropic.Models),
			})
		}

		// OpenAI
		if s.fullConfig.Providers.OpenAI.APIKey != "" || s.fullConfig.Providers.OpenAI.AuthMethod != "" {
			providersList = append(providersList, providerInfo{
				Name: "openai", Enabled: true,
				IsActive: currentProvider == "openai",
				Models: getModels("openai", []modelInfo{
					{ID: "gpt-4o", Provider: "openai"},
					{ID: "gpt-4o-mini", Provider: "openai"},
					{ID: "gpt-4-turbo", Provider: "openai"},
					{ID: "o1", Provider: "openai"},
					{ID: "o1-mini", Provider: "openai"},
				}, s.fullConfig.Providers.OpenAI.Models),
			})
		}

		// OpenRouter
		if s.fullConfig.Providers.OpenRouter.APIKey != "" {
			providersList = append(providersList, providerInfo{
				Name: "openrouter", Enabled: true,
				IsActive: currentProvider == "openrouter",
				Models: getModels("openrouter", []modelInfo{
					{ID: "anthropic/claude-sonnet-4-20250514", Provider: "openrouter"},
					{ID: "openai/gpt-4o", Provider: "openrouter"},
					{ID: "google/gemini-2.5-pro-preview", Provider: "openrouter"},
					{ID: "deepseek/deepseek-r1", Provider: "openrouter"},
					{ID: "meta-llama/llama-4-maverick", Provider: "openrouter"},
				}, s.fullConfig.Providers.OpenRouter.Models),
			})
		}

		// Groq
		if s.fullConfig.Providers.Groq.APIKey != "" {
			providersList = append(providersList, providerInfo{
				Name: "groq", Enabled: true,
				IsActive: currentProvider == "groq",
				Models: getModels("groq", []modelInfo{
					{ID: "llama-3.3-70b-versatile", Provider: "groq"},
					{ID: "llama-3.1-8b-instant", Provider: "groq"},
					{ID: "mixtral-8x7b-32768", Provider: "groq"},
				}, s.fullConfig.Providers.Groq.Models),
			})
		}

		// Gemini
		if s.fullConfig.Providers.Gemini.APIKey != "" {
			providersList = append(providersList, providerInfo{
				Name: "gemini", Enabled: true,
				IsActive: currentProvider == "gemini",
				Models: getModels("gemini", []modelInfo{
					{ID: "gemini-2.5-pro-preview-05-06", Provider: "gemini"},
					{ID: "gemini-2.5-flash-preview-05-20", Provider: "gemini"},
					{ID: "gemini-2.0-flash", Provider: "gemini"},
				}, s.fullConfig.Providers.Gemini.Models),
			})
		}

		// Zhipu
		if s.fullConfig.Providers.Zhipu.APIKey != "" {
			providersList = append(providersList, providerInfo{
				Name: "zhipu", Enabled: true,
				IsActive: currentProvider == "zhipu",
				Models: getModels("zhipu", []modelInfo{
					{ID: "glm-4.7", Provider: "zhipu"},
					{ID: "glm-4-flash", Provider: "zhipu"},
				}, s.fullConfig.Providers.Zhipu.Models),
			})
		}

		// Moonshot
		if s.fullConfig.Providers.Moonshot.APIKey != "" {
			providersList = append(providersList, providerInfo{
				Name: "moonshot", Enabled: true,
				IsActive: currentProvider == "moonshot",
				Models: getModels("moonshot", []modelInfo{
					{ID: "moonshot/kimi-k2.5", Provider: "moonshot"},
				}, s.fullConfig.Providers.Moonshot.Models),
			})
		}

		// Nvidia
		if s.fullConfig.Providers.Nvidia.APIKey != "" {
			providersList = append(providersList, providerInfo{
				Name: "nvidia", Enabled: true,
				IsActive: currentProvider == "nvidia",
				Models: getModels("nvidia", []modelInfo{
					{ID: "nvidia/llama-3.1-nemotron-70b-instruct", Provider: "nvidia"},
				}, s.fullConfig.Providers.Nvidia.Models),
			})
		}

		// Ollama
		if s.fullConfig.Providers.Ollama.APIBase != "" {
			providersList = append(providersList, providerInfo{
				Name: "ollama", Enabled: true,
				IsActive: currentProvider == "ollama",
				Models: getModels("ollama", []modelInfo{
					{ID: "llama3.2", Provider: "ollama"},
				}, s.fullConfig.Providers.Ollama.Models),
			})
		}

		// VLLM
		if s.fullConfig.Providers.VLLM.APIBase != "" {
			providersList = append(providersList, providerInfo{
				Name: "vllm", Enabled: true,
				IsActive: currentProvider == "vllm",
				Models:   []modelInfo{},
			})
		}
	}

	if providersList == nil {
		providersList = []providerInfo{}
	}
	for i := range providersList {
		if providersList[i].Models == nil {
			providersList[i].Models = []modelInfo{}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"current_model":    currentModel,
		"current_provider": currentProvider,
		"providers":        providersList,
	})
}

func (s *Server) handleLongTermMemory(w http.ResponseWriter, r *http.Request) {
	if s.memory == nil {
		http.Error(w, "memory store unavailable", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		content := s.memory.ReadLongTerm()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"content": content})
	case http.MethodPost:
		var in struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if err := s.memory.WriteLongTerm(in.Content); err != nil {
			http.Error(w, "failed to write memory", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleDailyNotes(w http.ResponseWriter, r *http.Request) {
	if s.memory == nil {
		http.Error(w, "memory store unavailable", http.StatusServiceUnavailable)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 7
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	content := s.memory.GetRecentDailyNotes(days)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"content": content})
}

// handleVoiceTranscribe handles POST /api/v1/voice/transcribe
// Accepts multipart/form-data with an "audio" file field.
// Returns JSON { "text": "...", "language": "...", "duration": 0.0 }
func (s *Server) handleVoiceTranscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.transcriber == nil || !s.transcriber.IsAvailable() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Voice transcription not available. Configure Groq API key.",
		})
		return
	}

	// Limit upload size to 25MB (Groq Whisper API limit)
	const maxUploadSize = 25 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to parse form. File may be too large (max 25MB).",
		})
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Missing 'audio' file in request.",
		})
		return
	}
	defer file.Close()

	// Save to temp file (transcriber requires file path)
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".webm" // Default extension for browser MediaRecorder
	}
	tmpFile, err := os.CreateTemp("", "KakoClaw-voice-*"+ext)
	if err != nil {
		logger.ErrorCF("web", "Failed to create temp file for voice", map[string]interface{}{"error": err.Error()})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, file); err != nil {
		logger.ErrorCF("web", "Failed to write voice temp file", map[string]interface{}{"error": err.Error()})
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = tmpFile.Close() // Close before transcribing so file is flushed

	result, err := s.transcriber.Transcribe(r.Context(), tmpFile.Name())
	if err != nil {
		logger.ErrorCF("web", "Voice transcription failed", map[string]interface{}{"error": err.Error()})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Transcription failed: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// handleMetrics returns in-process observability metrics (LLM calls, tool calls, agent runs).
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	snapshot := observability.Global().Snapshot()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snapshot)
}
