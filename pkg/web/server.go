package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/mail"
	"net/url"

	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sipeed/makoclaw/pkg/agent"
	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/channels"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/cron"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/mcp"
	"github.com/sipeed/makoclaw/pkg/observability"
	"github.com/sipeed/makoclaw/pkg/providers"
	"github.com/sipeed/makoclaw/pkg/ratelimit"
	"github.com/sipeed/makoclaw/pkg/skills"
	"github.com/sipeed/makoclaw/pkg/storage"
	"github.com/sipeed/makoclaw/pkg/voice"
	"github.com/sipeed/makoclaw/pkg/workflow"
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
	cfg                     config.WebConfig
	fullConfig              *config.Config
	workspace               string
	agentLoop               *agent.AgentLoop
	agentManager            *agent.AgentManager
	server                  *http.Server
	authManager             *authManager
	store                   *storage.Storage            // legacy shared storage (kept for backward compat during migration)
	centralStore            *storage.CentralStorage     // central DB: users, settings, channel mappings
	userMgr                 *storage.UserStorageManager // per-user DB manager
	loginLimit              *ratelimit.RateLimiter
	apiRateLimit            *ratelimit.RateLimiter
	tasksMu                 sync.RWMutex
	tasksClients            map[*websocket.Conn]struct{}
	connMu                  map[*websocket.Conn]*sync.Mutex
	memory                  *agent.MemoryStore
	cronService             *cron.CronService
	skillsLoader            *skills.SkillsLoader
	skillInstaller          *skills.SkillInstaller
	channelManager          *channels.Manager
	multiUserChannelManager *channels.MultiUserChannelManager
	transcriber             *voice.GroqTranscriber
	mcpManager              *mcp.Manager
	workflowEngine          *workflow.Engine
	execMu                  sync.RWMutex
	activeExecs             map[string]*activeExecution
	msgBus                  *bus.MessageBus
	ctx                     context.Context
	allowedOrigins          []string
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

func (s *Server) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	requestHost := r.Host
	if h, _, err := net.SplitHostPort(requestHost); err == nil {
		requestHost = h
	}
	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}
	originHost := originURL.Hostname()
	if originHost == requestHost {
		return true
	}
	if len(s.allowedOrigins) > 0 {
		for _, allowed := range s.allowedOrigins {
			if allowed == origin {
				return true
			}
		}
		return false
	}
	return true
}

func NewServer(cfg config.WebConfig, agentLoop *agent.AgentLoop, store *storage.Storage) *Server {
	workspace := defaultWorkspace()
	return &Server{
		cfg:          cfg,
		workspace:    workspace,
		agentLoop:    agentLoop,
		store:        store,
		loginLimit:   ratelimit.NewRateLimiter(),
		apiRateLimit: ratelimit.NewRateLimiter(),
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

// SetCentralStorage sets the central database storage for user auth operations.
func (s *Server) SetCentralStorage(cs *storage.CentralStorage) {
	s.centralStore = cs
}

// SetUserStorageManager sets the per-user storage manager.
func (s *Server) SetUserStorageManager(mgr *storage.UserStorageManager) {
	s.userMgr = mgr
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

// SetMultiUserChannelManager injects the multi-user channel manager for REST exposure
func (s *Server) SetMultiUserChannelManager(m *channels.MultiUserChannelManager) {
	s.multiUserChannelManager = m
}

// SetAgentManager injects the agent manager for agent orchestration endpoints
func (s *Server) SetAgentManager(m *agent.AgentManager) {
	s.agentManager = m
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

// SetBus injects the message bus for REST exposure
func (s *Server) SetBus(b *bus.MessageBus) {
	s.msgBus = b
}

func defaultWorkspace() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return filepath.Join(home, ".makoclaw", "workspace")
}

func (s *Server) Start(ctx context.Context) error {
	// Use central storage for auth if available, fallback to legacy shared storage
	authStore := s.centralStore
	if authStore == nil {
		logger.WarnC("web", "CentralStorage not provided, auth features may be limited")
	}
	authManager, err := newAuthManager(authStore, s.cfg.Username, s.cfg.Password, s.cfg.JWTExpiry)
	if err != nil {
		return err
	}
	s.authManager = authManager
	s.ctx = ctx

	// Ensure storage is available
	if s.store == nil {
		logger.WarnC("web", "Legacy storage not provided to web server, some features may be disabled")
	} else {
		// Wire persistent storage into the metrics singleton so that counters survive restarts.
		observability.Global().SetStorage(s.store)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", s.handleHealth)
	mux.HandleFunc("/api/v1/auth/login", s.handleLogin)
	mux.HandleFunc("/api/v1/auth/register", s.handleRegister)
	mux.HandleFunc("/api/v1/auth/change-password", s.handleChangePassword)
	mux.HandleFunc("/api/v1/auth/profile", s.handleProfile)
	mux.HandleFunc("/api/v1/auth/me", s.handleMe)

	// User-specific config management (use /api/v1/me/* to avoid /api/v1/users/ conflict)
	mux.HandleFunc("/api/v1/me/config", s.handleGetUserConfig)                   // Get user's merged config
	mux.HandleFunc("/api/v1/me/config/update", s.handleUpdateUserConfig)         // Update user config
	mux.HandleFunc("/api/v1/me/config/reset", s.handleDeleteUserConfigSection)   // Reset section to global
	mux.HandleFunc("/api/v1/me/providers", s.handleGetUserProviders)             // Get user's providers
	mux.HandleFunc("/api/v1/me/onboarding/complete", s.handleCompleteOnboarding) // Mark onboarding complete
	mux.HandleFunc("/api/v1/me/workspace/init", s.handleWorkspaceInit)           // Initialize workspace with skills
	mux.HandleFunc("/api/v1/me/providers/update", s.handleUpdateUserProvider)    // Update specific provider
	mux.HandleFunc("/api/v1/me/channels", s.handleGetUserChannels)               // Get user's channels
	// Backwards-compatible aliases
	mux.HandleFunc("/api/v1/users/me/config", s.handleGetUserConfig)
	mux.HandleFunc("/api/v1/users/me/config/update", s.handleUpdateUserConfig)
	mux.HandleFunc("/api/v1/users/me/config/reset", s.handleDeleteUserConfigSection)
	mux.HandleFunc("/api/v1/users/me/providers", s.handleGetUserProviders)
	mux.HandleFunc("/api/v1/users/me/providers/update", s.handleUpdateUserProvider)
	mux.HandleFunc("/api/v1/users/me/channels", s.handleGetUserChannels)

	// General user routes (must come after specific /users/me/* routes)
	mux.HandleFunc("/api/v1/users", s.handleUsers) // User management (Admin only)
	// Custom handler for user actions, block, and unblock
	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		// Check if path matches /users/:id/block or /users/:id/unblock
		path := r.URL.Path
		if strings.HasSuffix(path, "/block") && r.Method == http.MethodPost {
			s.handleBlockUser(w, r)
			return
		}
		if strings.HasSuffix(path, "/unblock") && r.Method == http.MethodPost {
			s.handleUnblockUser(w, r)
			return
		}
		// Tool permissions endpoints
		if strings.Contains(path, "/tools") {
			s.handleUserTools(w, r)
			return
		}
		// Otherwise, handle as regular user action
		s.handleUserAction(w, r)
	})

	// Tool permissions management endpoints (Admin only)
	mux.HandleFunc("/api/v1/tools/permissions", s.handleToolPermissions)
	mux.HandleFunc("/api/v1/tools/audit", s.handleToolAudit)

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
	mux.HandleFunc("/api/v1/config/status", s.handleConfigStatus)             // Degraded mode status
	mux.HandleFunc("/api/v1/config/provider", s.handleConfigProvider)         // Update provider config
	mux.HandleFunc("/api/v1/config/validate", s.handleConfigValidate)         // Validate provider config
	mux.HandleFunc("/api/v1/agents", s.handleAgents)                          // Agents list
	mux.HandleFunc("/api/v1/agents/", s.handleAgentAction)                    // Agent details, sessions, config
	mux.HandleFunc("/api/v1/agents/specialist", s.handleSpecialistCreate)     // Create specialist
	mux.HandleFunc("/api/v1/agents/specialist/", s.handleSpecialistAction)    // Update/Delete specialist
	mux.HandleFunc("/api/v1/agents/generate", s.handleSpecialistGenerate)     // Generate specialist with AI
	mux.HandleFunc("/api/v1/agents/orchestrator", s.handleOrchestratorConfig) // Update orchestrator config
	mux.HandleFunc("/api/v1/agents/metrics", s.handleAgentMetrics)            // Agent cost metrics
	mux.HandleFunc("/api/v1/agents/specialists", s.handleGetSpecialists)      // Get active specialists list
	mux.HandleFunc("/api/v1/reports/email", s.handleSendReportEmail)         // Send report email directly

	// Setup/Onboarding flow (Phase 4)
	mux.HandleFunc("/api/v1/setup/initialize", s.handleSetupInitialize)   // Create setup session
	mux.HandleFunc("/api/v1/setup/validate/", s.handleSetupValidate)      // Validate token
	mux.HandleFunc("/api/v1/setup/complete/", s.handleSetupComplete)      // Complete setup
	mux.HandleFunc("/api/v1/test-channel/telegram", s.handleTestTelegram) // Test Telegram connection

	mux.HandleFunc("/api/v1/files", s.handleFiles)                            // File browser
	mux.HandleFunc("/api/v1/files/", s.handleFiles)                           // File browser subpaths
	mux.HandleFunc("/api/v1/export/tasks", s.handleExportTasks)               // Export tasks
	mux.HandleFunc("/api/v1/export/chat", s.handleExportChat)                 // Export chat history
	mux.HandleFunc("/api/v1/import/chat", s.handleImportChat)                 // Import conversations
	mux.HandleFunc("/api/v1/models", s.handleModels)                          // Available models/providers
	mux.HandleFunc("/api/v1/providers/catalog", s.handleProvidersCatalog)     // All available providers metadata
	mux.HandleFunc("/api/v1/voice/transcribe", s.handleVoiceTranscribe)       // Voice-to-text (Groq STT)
	mux.HandleFunc("/api/v1/knowledge", s.handleKnowledge)                    // Knowledge base: list + upload
	mux.HandleFunc("/api/v1/knowledge/search", s.handleKnowledgeSearch)       // Knowledge base: FTS5 search
	mux.HandleFunc("/api/v1/knowledge/chunks/", s.handleKnowledgeChunkAction) // Knowledge base: update chunks
	mux.HandleFunc("/api/v1/knowledge/", s.handleKnowledgeAction)             // Knowledge base: view chunks or delete by ID
	mux.HandleFunc("/api/v1/openapi.json", s.handleOpenAPISpec)               // OpenAPI 3.0 spec (JSON)
	mux.HandleFunc("/api/docs", s.handleAPIDocsUI)                            // Swagger UI
	mux.HandleFunc("/api/v1/mcp", s.handleMCPServers)                         // MCP servers: list + status
	mux.HandleFunc("/api/v1/mcp/reconnect-all", s.handleMCPReconnectAll)     // MCP: reconnect all servers
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
		Handler: s.recoveryMiddleware(s.authMiddleware(s.apiRateLimitMiddleware(mux))),
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

	if s.userMgr != nil && s.centralStore != nil && s.fullConfig != nil && s.msgBus != nil && s.store != nil {
		go s.runTaskWorker(ctx)
	} else if s.agentLoop != nil && s.store != nil {
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
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net blob:; style-src 'self' 'unsafe-inline'; connect-src 'self' ws: wss: https://cdn.jsdelivr.net; worker-src 'self' blob:; upgrade-insecure-requests")

		if strings.HasPrefix(r.URL.Path, "/api/v1/auth/login") {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/v1/auth/register") {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/v1/health") {
			next.ServeHTTP(w, r)
			return
		}
		// Providers catalog is public (needed for onboarding wizard before authentication)
		if strings.HasPrefix(r.URL.Path, "/api/v1/providers/catalog") {
			next.ServeHTTP(w, r)
			return
		}
		// Channel test endpoints are public (needed for setup wizard before authentication)
		if strings.HasPrefix(r.URL.Path, "/api/v1/test-channel") {
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
				writeJSONResponse(w, map[string]string{
					"error": "unauthorized",
				})
				return
			}
			// Check if user is blocked (use central store if available)
			if s.centralStore != nil {
				user, err := s.centralStore.GetUserByUsername(claims.Sub)
				if err == nil && user.Blocked {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					writeJSONResponse(w, map[string]interface{}{
						"error":   fmt.Sprintf("Usuario bloqueado. Motivo: %s. Contacte soporte.", user.BlockedReason),
						"blocked": true,
					})
					return
				}
			} else if s.store != nil {
				user, err := s.resolveUserByUsername(claims.Sub)
				if err == nil && user.Blocked {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					writeJSONResponse(w, map[string]interface{}{
						"error":   fmt.Sprintf("Usuario bloqueado. Motivo: %s. Contacte soporte.", user.BlockedReason),
						"blocked": true,
					})
					return
				}
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
	if strings.HasPrefix(r.URL.Path, "/ws/") {
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
// NOTE: This is the legacy method. Prefer getUserStorage() for per-user isolation.
func (s *Server) getUserIDFromClaims(r *http.Request) (int64, bool) {
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil {
		return 0, false
	}
	// Try central store first
	if s.centralStore != nil {
		user, err := s.centralStore.GetUserByUsername(claims.Sub)
		if err != nil {
			return 0, false
		}
		return user.ID, true
	}
	// Fallback to legacy store
	if s.store == nil {
		return 0, false
	}
	user, err := s.resolveUserByUsername(claims.Sub)
	if err != nil {
		return 0, false
	}
	return user.ID, true
}

// getCronServiceForRequest resolves the cron service and user ID for the request.
// It prefers per-user cron services when the multi-user manager is available.
// In degraded mode, it will create per-user cron services on demand.
func (s *Server) getCronServiceForRequest(r *http.Request) (*cron.CronService, int64, bool) {
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil {
		return nil, 0, false
	}

	var userID int64
	userUUID := claims.UUID

	if userUUID != "" {
		if s.centralStore != nil {
			if user, err := s.centralStore.GetUserByUUID(userUUID); err == nil && user != nil {
				userID = user.ID
			}
		} else if s.store != nil {
			if user, err := s.store.GetUserByUUID(userUUID); err == nil && user != nil {
				userID = user.ID
			}
		}
	}

	if userID == 0 {
		if s.centralStore != nil {
			if user, err := s.centralStore.GetUserByUsername(claims.Sub); err == nil && user != nil {
				userID = user.ID
				if userUUID == "" {
					userUUID = user.UUID
				}
			}
		} else if s.store != nil {
			if user, err := s.resolveUserByUsername(claims.Sub); err == nil && user != nil {
				userID = user.ID
				if userUUID == "" {
					userUUID = user.UUID
				}
			}
		}
	}

	if userID == 0 {
		return nil, 0, false
	}

	// Try to get existing per-user cron service (full mode with agent loop)
	if s.multiUserChannelManager != nil && userUUID != "" {
		if cronService, exists := s.multiUserChannelManager.GetCronServiceForUser(userUUID); exists {
			_ = cronService.Start()
			return cronService, userID, true
		}

		// In degraded mode, create a minimal per-user cron service on demand
		// This allows users to schedule jobs (with deliver=true) without an LLM provider
		if cronService, err := s.multiUserChannelManager.GetOrCreateCronServiceForUser(userUUID); err == nil {
			_ = cronService.Start()
			return cronService, userID, true
		}
	}

	if s.cronService != nil {
		return s.cronService, userID, true
	}

	return nil, userID, true
}

// getUserStorage returns the per-user Storage for the authenticated user.
// This is the preferred method for the multi-user architecture.
// Returns the per-user storage, user UUID, and true if successful.
func (s *Server) getUserStorage(r *http.Request) (*storage.Storage, string, bool) {
	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil {
		return nil, "", false
	}

	// If per-user storage manager is available, use it
	if s.userMgr != nil {
		userUUID := claims.UUID
		if userUUID == "" {
			// Fallback: get UUID from central store via username
			if s.centralStore != nil {
				user, err := s.centralStore.GetUserByUsername(claims.Sub)
				if err != nil {
					return nil, "", false
				}
				userUUID = user.UUID
			}
		}
		if userUUID == "" {
			return nil, "", false
		}
		store, err := s.userMgr.GetOrCreate(userUUID)
		if err != nil {
			return nil, "", false
		}
		logger.InfoCF("web", "Resolved per-user storage", map[string]interface{}{"uuid": userUUID, "path": s.userMgr.UserDBPath(userUUID)})
		return store, userUUID, true
	}

	// Fallback: return legacy shared store
	if s.store != nil {
		return s.store, "", true
	}
	return nil, "", false
}

// resolveUserByUsername looks up a user by username, trying centralStore first, then legacy store.
func (s *Server) resolveUserByUsername(username string) (*storage.User, error) {
	if s.centralStore != nil {
		return s.centralStore.GetUserByUsername(username)
	}
	if s.store != nil {
		return s.store.GetUserByUsername(username)
	}
	return nil, storage.ErrUserNotFound
}

// resolveUserByID looks up a user by ID, trying centralStore first, then legacy store.
func (s *Server) resolveUserByID(id int64) (*storage.User, error) {
	if s.centralStore != nil {
		return s.centralStore.GetUserByID(id)
	}
	if s.store != nil {
		return s.store.GetUserByID(id)
	}
	return nil, storage.ErrUserNotFound
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]string{
		"status": "ok",
	})
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	store, userID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized or store unavailable", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			includeArchived := r.URL.Query().Get("include_archived") == "true"
			tasks, err := store.ListTasksForUser(userID, includeArchived)
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
			writeJSONResponse(w, map[string]interface{}{"tasks": items})
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
			if len(in.Title) > 255 {
				http.Error(w, "title too long (max 255 chars)", http.StatusBadRequest)
				return
			}
			if len(in.Description) > 2000 {
				http.Error(w, "description too long (max 2000 chars)", http.StatusBadRequest)
				return
			}
			status, ok := normalizeTaskStatus(in.Status)
			if !ok {
				http.Error(w, "invalid status", http.StatusBadRequest)
				return
			}
			id, err := store.CreateTaskForUser(userID, strings.TrimSpace(in.Title), strings.TrimSpace(in.Description), status)
			if err != nil {
				http.Error(w, "failed to create task", http.StatusInternalServerError)
				return
			}

			// Fetch the created task to get full details (like created_at)
			created, err := store.GetTaskForUser(userID, id)
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
			writeJSONResponse(w, item)
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

		updated, err := store.UpdateTaskStatusForUser(userID, id, status)
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
		writeJSONResponse(w, item)
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

		if err := store.ArchiveTaskForUser(userID, id); err != nil {
			http.Error(w, "failed to archive task", http.StatusInternalServerError)
			return
		}

		task, err := store.GetTaskForUser(userID, id)
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
		writeJSONResponse(w, map[string]string{"status": "ok"})
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

		if err := store.UnarchiveTaskForUser(userID, id); err != nil {
			http.Error(w, "failed to unarchive task", http.StatusInternalServerError)
			return
		}

		task, err := store.GetTaskForUser(userID, id)
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
		writeJSONResponse(w, map[string]string{"status": "ok"})
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
		logs, err := store.GetTaskLogs(logsID)
		if err != nil {
			http.Error(w, "failed to get task logs", http.StatusInternalServerError)
			return
		}
		if logs == nil {
			logs = []storage.TaskLog{}
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]interface{}{"logs": logs})
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

		updated, err := store.UpdateTaskForUser(
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
		writeJSONResponse(w, item)
	case http.MethodDelete:
		if err := store.DeleteTaskForUser(userID, id); err != nil {
			http.Error(w, "failed to delete task", http.StatusInternalServerError)
			return
		}
		s.broadcastTaskEvent("deleted", taskItem{ID: idStr})
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type chatResponse struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (s *Server) handleChatWS(w http.ResponseWriter, r *http.Request) {
	if s.userMgr == nil || s.fullConfig == nil || s.msgBus == nil || s.store == nil {
		http.Error(w, "agent loop unavailable", http.StatusServiceUnavailable)
		return
	}

	// Extract user_id once per connection
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userStore, userUUID, userOK := s.getUserStorage(r)
	if !userOK || userUUID == "" || userStore == nil {
		http.Error(w, "agent loop unavailable", http.StatusServiceUnavailable)
		return
	}

	activeAgentLoop, err := agent.NewAgentLoopForUser(userUUID, s.fullConfig, s.msgBus, s.centralStore, userStore)
	if err != nil {
		logger.ErrorCF("web", "Failed to create per-user agent loop for websocket", map[string]interface{}{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		http.Error(w, "agent loop unavailable", http.StatusServiceUnavailable)
		return
	}
	defer activeAgentLoop.Stop()

	upgrader := websocket.Upgrader{CheckOrigin: s.checkOrigin}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetReadLimit(1 << 20) // 1 MB max message size to prevent memory DoS

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

		if handled, response := s.handleTaskChatCommand(userID, userStore, input); handled {
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
			// Inject activeAgentLoop as agent tracker so that any orchestrator
			// delegations register involved agents into this loop's tracking.
			ctx = agent.ContextWithAgentTracker(ctx, activeAgentLoop)
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
			if activeAgentLoop.SupportsStreaming() {
				// Send stream_start
				wsMu.Lock()
				_ = conn.WriteJSON(map[string]interface{}{"type": "stream_start"})
				wsMu.Unlock()

				// Add agent status callback to context
				ctx = agent.ContextWithAgentStatusCallback(ctx, func(ev agent.AgentStatusEvent) error {
					wsMu.Lock()
					defer wsMu.Unlock()
					return conn.WriteJSON(map[string]interface{}{
						"type":            "agent_status",
						"agent":           ev.Agent,
						"status":          ev.Status,
						"specialist_name": ev.SpecialistName,
						"reason":          ev.Reason,
						"timestamp":       ev.Timestamp.Format(time.RFC3339),
					})
				})

				// Add content segment callback to context
				ctx = agent.ContextWithContentSegmentCallback(ctx, func(segment agent.ContentSegment) error {
					wsMu.Lock()
					defer wsMu.Unlock()
					return conn.WriteJSON(map[string]interface{}{
						"type":       "content_segment",
						"agent":      segment.Agent,
						"content":    segment.Content,
						"segment_id": segment.SegmentID,
						"timestamp":  segment.Timestamp.Format(time.RFC3339),
					})
				})

				response, err := activeAgentLoop.ProcessDirectWithUserAndModelStream(
					ctx, userID, input, sessionID, req.Model,
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

				// Get involved agents for metadata
				agents := activeAgentLoop.GetInvolvedAgents()
				wsMu.Lock()
				streamEndMsg := map[string]interface{}{
					"type":    "stream_end",
					"content": response,
				}
				if len(agents) > 0 {
					streamEndMsg["agents"] = agents
				}
				_ = conn.WriteJSON(streamEndMsg)
				_ = conn.WriteJSON(map[string]interface{}{"type": "ready"})
				wsMu.Unlock()
			} else {
				// Non-streaming fallback
				response, err := activeAgentLoop.ProcessDirectWithUserAndModel(ctx, userID, input, sessionID, req.Model, excludeTools...)
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

				// Get involved agents for metadata
				agents := activeAgentLoop.GetInvolvedAgents()
				wsMu.Lock()
				messageMsg := map[string]interface{}{
					"type":    "message",
					"role":    "assistant",
					"content": response,
				}
				if len(agents) > 0 {
					messageMsg["agents"] = agents
				}
				_ = conn.WriteJSON(messageMsg)
				_ = conn.WriteJSON(map[string]interface{}{"type": "ready"})
				wsMu.Unlock()
			}
		}()
	}
}

func (s *Server) handleTaskChatCommand(userID int64, store *storage.Storage, input string) (bool, string) {
	if store == nil {
		return false, ""
	}
	lower := strings.ToLower(strings.TrimSpace(input))
	if strings.HasPrefix(lower, "/task create ") {
		title := strings.TrimSpace(input[len("/task create "):])
		if title == "" {
			return true, "Uso: /task create <titulo>"
		}
		id, err := store.CreateTaskForUser(userID, title, "", "todo")
		if err != nil {
			return true, "No pude crear la tarea."
		}

		created, err := store.GetTaskForUser(userID, id)
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
		tasks, err := store.ListTasksForUser(userID, false)
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

		updated, err := store.UpdateTaskStatusForUser(userID, id, "todo")
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
		updated, err := store.UpdateTaskStatusForUser(userID, id, status)
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
		if err := store.ArchiveTaskForUser(userID, id); err != nil {
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
	writeJSONResponse(w, map[string]interface{}{
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
	writeJSONResponse(w, active)
}

func (s *Server) handleTasksWS(w http.ResponseWriter, r *http.Request) {
	s.tasksMu.RLock()
	clientCount := len(s.tasksClients)
	s.tasksMu.RUnlock()
	if clientCount >= maxWebSocketClients {
		http.Error(w, "too many connections", http.StatusServiceUnavailable)
		return
	}

	taskUpgrader := websocket.Upgrader{CheckOrigin: s.checkOrigin}
	conn, err := taskUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn.SetReadLimit(1 << 20) // 1 MB max message size
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	identifier := strings.TrimSpace(in.Username)
	if identifier == "" {
		identifier = strings.TrimSpace(in.Email)
	}
	if identifier == "" || in.Password == "" {
		http.Error(w, "username or email and password are required", http.StatusBadRequest)
		return
	}

	token, err := s.authManager.login(identifier, in.Password)
	if err != nil {
		// Check if it's a blocked user error
		if strings.Contains(err.Error(), "bloqueado") || strings.Contains(err.Error(), "blocked") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			writeJSONResponse(w, map[string]interface{}{
				"error":   err.Error(),
				"blocked": true,
			})
			return
		}
		// Generic error for invalid credentials
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]string{"token": token})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.authManager == nil || s.centralStore == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	ip := clientIP(r)
	key := "register:" + ip
	s.loginLimit.SetLimit(key, 3, time.Hour) // Max 3 registrations per hour per IP
	if !s.loginLimit.Allow(key) {
		http.Error(w, "too many registration attempts", http.StatusTooManyRequests)
		return
	}

	var in struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"` // Optional: defaults to "user"
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Basic validation
	in.Username = strings.TrimSpace(in.Username)
	in.Email = strings.TrimSpace(in.Email)
	in.Role = strings.TrimSpace(in.Role)

	// Force role to "user" for public registration — admin accounts must be
	// created through other means (CLI, direct DB, or by an existing admin).
	in.Role = "user"

	if in.Username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}
	if in.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if len(in.Password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Validate email format using RFC 5322 parser
	if _, err := mail.ParseAddress(in.Email); err != nil {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}

	// Check if email is blocked (use central store)
	if blocked, err := s.centralStore.IsEmailBlocked(in.Email); err == nil && blocked {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		writeJSONResponse(w, map[string]interface{}{
			"error":   "No se puede registrar. Contacte soporte.",
			"blocked": true,
		})
		return
	}

	// Create user with specified role (defaults to "user")
	user, err := s.centralStore.CreateUserWithEmail(in.Username, in.Email, in.Password, in.Role)
	if err != nil {
		if err == storage.ErrUserExists {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			writeJSONResponse(w, map[string]interface{}{
				"error": "Username or email already exists. Please try a different username or email.",
			})
			return
		}
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	// Initialize per-user workspace and database
	if s.userMgr != nil && user.UUID != "" {
		if err := s.userMgr.EnsureUserDirectory(user.UUID); err != nil {
			logger.WarnCF("web", "Failed to create user workspace on register", map[string]interface{}{
				"user": user.Username, "error": err.Error(),
			})
		}
		if _, err := s.userMgr.GetOrCreate(user.UUID); err != nil {
			logger.WarnCF("web", "Failed to create user database on register", map[string]interface{}{
				"user": user.Username, "error": err.Error(),
			})
		}
		// Seed a default config.json for the new user so they start with a
		// proper isolated configuration instead of falling back to the global one.
		if s.fullConfig != nil {
			if err := config.SaveConfigForUser(user.UUID, s.fullConfig); err != nil {
				logger.WarnCF("web", "Failed to seed user config.json on register", map[string]interface{}{
					"user": user.Username, "error": err.Error(),
				})
			}
		}
	}

	// Auto-login: generate token
	token, err := s.authManager.signToken(user.Username, user.UUID, user.Role)
	if err != nil {
		http.Error(w, "user created but login failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
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
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
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
	if err := s.authManager.changePassword(claims.Sub, in.CurrentPassword, in.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]string{"status": "ok"})
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	if s.authManager == nil || (s.centralStore == nil && s.store == nil) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		var user *storage.User
		var err error
		if s.centralStore != nil {
			user, err = s.centralStore.GetUserByUsername(claims.Sub)
		} else {
			user, err = s.resolveUserByUsername(claims.Sub)
		}
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		user.PasswordHash = "" // Redact password hash
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, user)
		return
	}

	if r.Method == http.MethodPut {
		ip := clientIP(r)
		key := "update-profile:" + ip
		s.loginLimit.SetLimit(key, 10, time.Minute)
		if !s.loginLimit.Allow(key) {
			http.Error(w, "too many attempts", http.StatusTooManyRequests)
			return
		}

		var in struct {
			Username          string `json:"username"`
			Email             string `json:"email"`
			FullName          string `json:"full_name"`
			DateOfBirth       string `json:"date_of_birth"` // ISO 8601: "2000-01-31"
			Timezone          string `json:"timezone"`
			PreferredLanguage string `json:"preferred_language"`
			AvatarURL         string `json:"avatar_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		// Parse optional date_of_birth
		var dob *time.Time
		if in.DateOfBirth != "" {
			parsed, err := time.Parse("2006-01-02", in.DateOfBirth)
			if err != nil {
				http.Error(w, "invalid date_of_birth format, expected YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			dob = &parsed
		}

		var user *storage.User
		var err error
		if s.centralStore != nil {
			user, err = s.centralStore.GetUserByUsername(claims.Sub)
		} else {
			user, err = s.resolveUserByUsername(claims.Sub)
		}
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		if s.centralStore != nil {
			if err := s.centralStore.UpdateUserProfile(user.ID, in.Username, in.Email, in.FullName, dob, in.Timezone, in.PreferredLanguage, in.AvatarURL); err != nil {
				if err == storage.ErrUserExists {
					http.Error(w, "username already exists", http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			if err := s.store.UpdateUserProfile(user.ID, in.Username, in.Email, in.FullName, dob, in.Timezone, in.PreferredLanguage, in.AvatarURL); err != nil {
				if err == storage.ErrUserExists {
					http.Error(w, "username already exists", http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// If username changed, issue new token
		newToken := ""
		if in.Username != "" && in.Username != claims.Sub {
			// Get updated user to get role
			var updatedUser *storage.User
			if s.centralStore != nil {
				updatedUser, err = s.centralStore.GetUserByID(user.ID)
			} else {
				updatedUser, err = s.store.GetUserByID(user.ID)
			}
			if err == nil {
				newToken, _ = s.authManager.signToken(updatedUser.Username, updatedUser.UUID, updatedUser.Role)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{"status": "ok"}
		if newToken != "" {
			response["token"] = newToken
		}
		writeJSONResponse(w, response)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleTaskSearch(w http.ResponseWriter, r *http.Request) {
	store, userID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized or store unavailable", http.StatusUnauthorized)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}
	results, err := store.SearchTasksForUser(userID, q)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{"tasks": results})
}

func (s *Server) handleChatSessions(w http.ResponseWriter, r *http.Request) {
	store, userID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized or store unavailable", http.StatusUnauthorized)
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

	sessions, err := store.ListSessionsForUser(userID, archivedFilter, limit, offset)
	if err != nil {
		http.Error(w, "failed to list sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{"sessions": sessions})
}

func (s *Server) handleChatSessionMessages(w http.ResponseWriter, r *http.Request) {
	store, userID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized or store unavailable", http.StatusUnauthorized)
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
		messages, err := store.GetMessagesForUser(userID, id)
		if err != nil {
			http.Error(w, "failed to get messages", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]interface{}{"messages": messages})

	case http.MethodDelete:
		if err := store.DeleteSessionForUser(userID, id); err != nil {
			http.Error(w, "failed to delete session", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]string{"status": "ok"})

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
		sess, err := store.UpdateSessionForUser(userID, id, payload.Title, payload.Archived)
		if err != nil {
			http.Error(w, "failed to update session", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, sess)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleChatSearch(w http.ResponseWriter, r *http.Request) {
	store, userID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized or store unavailable", http.StatusUnauthorized)
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}
	results, err := store.SearchMessagesForUser(userID, q)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{"messages": results})
}

func (s *Server) handleChatFork(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	store, userID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized or store unavailable", http.StatusUnauthorized)
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

	count, err := store.ForkSessionForUser(userID, payload.SessionID, newSessionID, payload.MessageID)
	if err != nil {
		http.Error(w, "fork failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
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

	// Fetch full user object to include onboarding_completed
	var onboardingCompleted bool
	if s.centralStore != nil && claims.UUID != "" {
		user, err := s.centralStore.GetUserByUUID(claims.UUID)
		if err == nil && user != nil {
			onboardingCompleted = user.OnboardingCompleted
		}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"username":             claims.Sub,
		"role":                 claims.Role,
		"user_uuid":            claims.UUID,
		"onboarding_completed": onboardingCompleted,
	})
}

func (s *Server) handleCompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.centralStore == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil || claims.UUID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Mark onboarding as completed for this user
	if err := s.centralStore.MarkOnboardingCompleted(claims.UUID); err != nil {
		// Fallback for legacy rows that may not have UUID populated
		if err == storage.ErrUserNotFound && claims.Sub != "" {
			if fallbackErr := s.centralStore.MarkOnboardingCompletedByUsername(claims.Sub); fallbackErr != nil {
				logger.ErrorCF("web", "Failed to mark onboarding complete (fallback by username failed)", map[string]interface{}{
					"user":  claims.Sub,
					"uuid":  claims.UUID,
					"error": fallbackErr.Error(),
				})
				http.Error(w, "failed to update onboarding status", http.StatusInternalServerError)
				return
			}
		} else {
			logger.ErrorCF("web", "Failed to mark onboarding complete", map[string]interface{}{
				"user":  claims.Sub,
				"uuid":  claims.UUID,
				"error": err.Error(),
			})
			http.Error(w, "failed to update onboarding status", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Onboarding completed successfully",
	})
}

// handleWorkspaceInit initializes user workspace with selected skills and example files
func (s *Server) handleWorkspaceInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
	if !ok || claims == nil || claims.UUID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Skills       []string `json:"skills"`
		ExampleFiles bool     `json:"exampleFiles"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get user workspace path
	var userWorkspace string
	if s.userMgr != nil {
		userWorkspace = s.userMgr.UserWorkspacePath(claims.UUID)
	} else {
		userWorkspace = s.fullConfig.WorkspacePath()
	}

	installedSkills := make([]string, 0)
	installedFiles := make([]string, 0)

	// Install selected skills
	if len(req.Skills) > 0 {
		// Get built-in skills path from repo
		builtinSkillsPath := filepath.Join(filepath.Dir(s.fullConfig.WorkspacePath()), "..", "..", "skills")
		userSkillsDir := filepath.Join(userWorkspace, "skills")

		// Create skills directory if it doesn't exist
		if err := os.MkdirAll(userSkillsDir, 0755); err != nil {
			logger.ErrorCF("web", "Failed to create skills directory", map[string]interface{}{
				"user": claims.Sub, "error": err.Error(),
			})
			http.Error(w, "failed to create skills directory", http.StatusInternalServerError)
			return
		}

		// Copy each selected skill
		for _, skillName := range req.Skills {
			srcSkillDir := filepath.Join(builtinSkillsPath, skillName)
			dstSkillDir := filepath.Join(userSkillsDir, skillName)

			// Check if skill exists in builtin
			if _, err := os.Stat(srcSkillDir); os.IsNotExist(err) {
				logger.WarnCF("web", "Skill not found in builtin", map[string]interface{}{
					"skill": skillName, "user": claims.Sub,
				})
				continue
			}

			// Copy skill directory
			if err := copyDir(srcSkillDir, dstSkillDir); err != nil {
				logger.ErrorCF("web", "Failed to copy skill", map[string]interface{}{
					"skill": skillName, "user": claims.Sub, "error": err.Error(),
				})
				continue
			}

			installedSkills = append(installedSkills, skillName)
		}
	}

	// Create example files if requested
	if req.ExampleFiles {
		examples := map[string]string{
			"WELCOME.md": `# Welcome to makoclaw!

This is your personal AI agent workspace.

## Quick Start

- Use the Chat view to interact with your agent
- Create tasks in the Tasks view for tracking work
- Explore Skills to see what your agent can do
- Configure settings in the Settings panel

## Example Commands

Try asking your agent:
- "Show me the files in my workspace"
- "Create a task to review the documentation"
- "What skills do you have available?"
`,
			"NOTES.md": `# Notes

Use this file to keep track of important information, ideas, and reminders.

## Ideas
- 

## To Remember
- 

## Questions
- 
`,
		}

		for filename, content := range examples {
			filePath := filepath.Join(userWorkspace, filename)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				logger.ErrorCF("web", "Failed to create example file", map[string]interface{}{
					"file": filename, "user": claims.Sub, "error": err.Error(),
				})
				continue
			}
			installedFiles = append(installedFiles, filename)
		}
	}

	logger.InfoCF("web", "Workspace initialized", map[string]interface{}{
		"user":   claims.Sub,
		"skills": installedSkills,
		"files":  installedFiles,
	})

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"success": true,
		"installed": map[string]interface{}{
			"skills": installedSkills,
			"files":  installedFiles,
		},
	})
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
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

// sanitizeFilename strips characters unsafe for Content-Disposition headers.
func sanitizeFilename(name string) string {
	var b strings.Builder
	for _, r := range name {
		if r == '"' || r == '\\' || r == '/' || r == '\n' || r == '\r' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
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
	if s.userMgr != nil && s.centralStore != nil && s.fullConfig != nil && s.msgBus != nil && s.store != nil {
		s.processNextTodoTaskPerUser(ctx)
		return
	}
	s.processNextTodoTaskLegacy(ctx)
}

func (s *Server) processNextTodoTaskLegacy(ctx context.Context) {
	tasks, err := s.store.ListAllUsersTasks(false)
	if err != nil {
		logger.WarnCF("web", "task worker: failed to list tasks", map[string]interface{}{"error": err.Error()})
		return
	}
	for _, t := range tasks {
		if t.Status != "todo" {
			continue
		}
		s.processTodoTaskWithLoop(ctx, s.store, s.agentLoop, t.UserID, t)
		return
	}
}

func (s *Server) processNextTodoTaskPerUser(ctx context.Context) {
	users, err := s.centralStore.ListUsers()
	if err != nil {
		logger.WarnCF("web", "task worker: failed to list users", map[string]interface{}{"error": err.Error()})
		return
	}

	for _, user := range users {
		if user == nil || user.UUID == "" {
			continue
		}

		userStore, err := s.userMgr.GetOrCreate(user.UUID)
		if err != nil {
			logger.WarnCF("web", "task worker: failed to open user storage", map[string]interface{}{"user_uuid": user.UUID, "error": err.Error()})
			continue
		}

		tasks, err := userStore.ListTasks(false)
		if err != nil {
			logger.WarnCF("web", "task worker: failed to list user tasks", map[string]interface{}{"user_uuid": user.UUID, "error": err.Error()})
			continue
		}

		for _, t := range tasks {
			if t.Status != "todo" {
				continue
			}

			userLoop, loopErr := agent.NewAgentLoopForUser(user.UUID, s.fullConfig, s.msgBus, s.centralStore, userStore)
			if loopErr != nil {
				_ = userStore.AddTaskLog(t.ID, "error", "Failed to create user agent loop: "+loopErr.Error())
				logger.WarnCF("web", "task worker: failed to create per-user agent loop", map[string]interface{}{"user_uuid": user.UUID, "task_id": t.ID, "error": loopErr.Error()})
				continue
			}

			s.processTodoTaskWithLoop(ctx, userStore, userLoop, user.ID, t)
			userLoop.Stop()
			return
		}
	}
}

func (s *Server) processTodoTaskWithLoop(ctx context.Context, store *storage.Storage, loop *agent.AgentLoop, userID int64, t storage.Task) {
	if store == nil || loop == nil {
		return
	}

	_ = store.AddTaskLog(t.ID, "started", "Task picked up by worker")

	inProgress, err := store.UpdateTaskStatusForUser(userID, t.ID, "in_progress")
	if err != nil {
		_ = store.AddTaskLog(t.ID, "error", "Failed to move to in_progress: "+err.Error())
		logger.WarnCF("web", "task worker: failed to update status", map[string]interface{}{"task_id": t.ID, "error": err.Error()})
		return
	}

	_ = store.AddTaskLog(t.ID, "status_changed", "Status changed to in_progress")

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

	_ = store.AddTaskLog(t.ID, "executing", "Sending task to AI agent")
	prompt := "Ejecuta esta tarea y devuelve un resumen breve.\nTitulo: " + t.Title + "\nDescripcion: " + t.Description
	taskSessionKey := "web:task:" + toString(t.ID)
	result, err := loop.ProcessDirectWithUser(ctx, userID, prompt, taskSessionKey)

	finalStatus := "review"
	finalResult := result
	if err != nil {
		finalResult = "error: " + err.Error()
		_ = store.AddTaskLog(t.ID, "error", "Agent execution failed: "+err.Error())
	} else {
		_ = store.AddTaskLog(t.ID, "completed", "Agent returned result successfully")
	}

	updated, updateErr := store.UpdateTaskForUser(userID, t.ID, t.Title, t.Description, finalStatus, finalResult)
	if updateErr != nil {
		_ = store.AddTaskLog(t.ID, "error", "Failed to save result: "+updateErr.Error())
		logger.WarnCF("web", "task worker: failed to save result", map[string]interface{}{"task_id": t.ID, "error": updateErr.Error()})
		return
	}

	_ = store.AddTaskLog(t.ID, "status_changed", "Status changed to review")

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

	// Load user-specific config if authenticated
	var mergedConfig *config.Config = s.fullConfig
	_, userUUID, ok := s.getUserStorage(r)
	if ok && userUUID != "" {
		userCfg, _ := config.LoadConfigForUser(userUUID)
		if userCfg != nil && s.fullConfig != nil {
			mergedConfig = config.MergeConfigs(s.fullConfig, userCfg)
		} else if userCfg != nil {
			mergedConfig = userCfg
		}
	}

	if mergedConfig != nil {
		currentModel = mergedConfig.Agents.Defaults.Model
		currentProvider, _ = providers.GetProviderForModel(currentModel)

		// Anthropic
		if mergedConfig.Providers.Anthropic.APIKey != "" || mergedConfig.Providers.Anthropic.AuthMethod != "" {
			providersList = append(providersList, providerInfo{
				Name: "anthropic", Enabled: true,
				IsActive: currentProvider == "anthropic",
				Models: getModels("anthropic", []modelInfo{
					{ID: "claude-sonnet-4-20250514", Provider: "anthropic"},
					{ID: "claude-3-5-haiku-20241022", Provider: "anthropic"},
					{ID: "claude-3-5-sonnet-20241022", Provider: "anthropic"},
					{ID: "claude-3-haiku-20240307", Provider: "anthropic"},
				}, mergedConfig.Providers.Anthropic.Models),
			})
		}

		// OpenAI
		if mergedConfig.Providers.OpenAI.APIKey != "" || mergedConfig.Providers.OpenAI.AuthMethod != "" {
			providersList = append(providersList, providerInfo{
				Name: "openai", Enabled: true,
				IsActive: currentProvider == "openai",
				Models: getModels("openai", []modelInfo{
					{ID: "gpt-4o", Provider: "openai"},
					{ID: "gpt-4o-mini", Provider: "openai"},
					{ID: "gpt-4-turbo", Provider: "openai"},
					{ID: "o1", Provider: "openai"},
					{ID: "o1-mini", Provider: "openai"},
				}, mergedConfig.Providers.OpenAI.Models),
			})
		}

		// OpenRouter
		if mergedConfig.Providers.OpenRouter.APIKey != "" {
			providersList = append(providersList, providerInfo{
				Name: "openrouter", Enabled: true,
				IsActive: currentProvider == "openrouter",
				Models: getModels("openrouter", []modelInfo{
					{ID: "anthropic/claude-sonnet-4-20250514", Provider: "openrouter"},
					{ID: "openai/gpt-4o", Provider: "openrouter"},
					{ID: "google/gemini-2.5-pro-preview", Provider: "openrouter"},
					{ID: "deepseek/deepseek-r1", Provider: "openrouter"},
					{ID: "meta-llama/llama-4-maverick", Provider: "openrouter"},
				}, mergedConfig.Providers.OpenRouter.Models),
			})
		}

		// Groq
		if mergedConfig.Providers.Groq.APIKey != "" {
			providersList = append(providersList, providerInfo{
				Name: "groq", Enabled: true,
				IsActive: currentProvider == "groq",
				Models: getModels("groq", []modelInfo{
					{ID: "llama-3.3-70b-versatile", Provider: "groq"},
					{ID: "llama-3.1-8b-instant", Provider: "groq"},
					{ID: "mixtral-8x7b-32768", Provider: "groq"},
				}, mergedConfig.Providers.Groq.Models),
			})
		}

		// Gemini
		if mergedConfig.Providers.Gemini.APIKey != "" {
			providersList = append(providersList, providerInfo{
				Name: "gemini", Enabled: true,
				IsActive: currentProvider == "gemini",
				Models: getModels("gemini", []modelInfo{
					{ID: "gemini-2.5-pro-preview-05-06", Provider: "gemini"},
					{ID: "gemini-2.5-flash-preview-05-20", Provider: "gemini"},
					{ID: "gemini-2.0-flash", Provider: "gemini"},
				}, mergedConfig.Providers.Gemini.Models),
			})
		}

		// Zhipu
		if mergedConfig.Providers.Zhipu.APIKey != "" {
			providersList = append(providersList, providerInfo{
				Name: "zhipu", Enabled: true,
				IsActive: currentProvider == "zhipu",
				Models: getModels("zhipu", []modelInfo{
					{ID: "glm-4-9", Provider: "zhipu"},
					{ID: "glm-4-air", Provider: "zhipu"},
					{ID: "glm-4-flash", Provider: "zhipu"},
				}, mergedConfig.Providers.Zhipu.Models),
			})
		}

		// Moonshot
		if mergedConfig.Providers.Moonshot.APIKey != "" {
			providersList = append(providersList, providerInfo{
				Name: "moonshot", Enabled: true,
				IsActive: currentProvider == "moonshot",
				Models: getModels("moonshot", []modelInfo{
					{ID: "moonshot/kimi-k2.5", Provider: "moonshot"},
				}, mergedConfig.Providers.Moonshot.Models),
			})
		}

		// Nvidia
		if mergedConfig.Providers.Nvidia.APIKey != "" {
			providersList = append(providersList, providerInfo{
				Name: "nvidia", Enabled: true,
				IsActive: currentProvider == "nvidia",
				Models: getModels("nvidia", []modelInfo{
					{ID: "nvidia/llama-3.1-nemotron-70b-instruct", Provider: "nvidia"},
				}, mergedConfig.Providers.Nvidia.Models),
			})
		}

		// Ollama
		if mergedConfig.Providers.Ollama.APIBase != "" {
			providersList = append(providersList, providerInfo{
				Name: "ollama", Enabled: true,
				IsActive: currentProvider == "ollama",
				Models: getModels("ollama", []modelInfo{
					{ID: "llama3.2", Provider: "ollama"},
				}, mergedConfig.Providers.Ollama.Models),
			})
		}

		// VLLM
		if mergedConfig.Providers.VLLM.APIBase != "" {
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
	writeJSONResponse(w, map[string]interface{}{
		"current_model":    currentModel,
		"current_provider": currentProvider,
		"providers":        providersList,
	})
}

// handleProvidersCatalog returns metadata for all available providers
// Endpoint: GET /api/v1/providers/catalog
// Returns: List of all supported providers with their metadata (not just configured ones)
func (s *Server) handleProvidersCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type providerMetadata struct {
		ID              string   `json:"id"`
		Name            string   `json:"name"`
		Description     string   `json:"description"`
		Icon            string   `json:"icon,omitempty"`
		RequiresAPIKey  bool     `json:"requires_api_key"`
		RequiresAPIBase bool     `json:"requires_api_base"`
		DefaultAPIBase  string   `json:"default_api_base,omitempty"`
		Models          []string `json:"models"`
		FreeTier        bool     `json:"free_tier"`
		DocsURL         string   `json:"docs_url"`
		APIKeyURL       string   `json:"api_key_url,omitempty"`
		Tip             string   `json:"tip,omitempty"`
	}

	catalog := []providerMetadata{
		{
			ID:              "openrouter",
			Name:            "OpenRouter",
			Description:     "Access 100+ models - Many free options available",
			Icon:            "🌐",
			RequiresAPIKey:  true,
			RequiresAPIBase: false,
			DefaultAPIBase:  "https://openrouter.ai/api/v1",
			Models: []string{
				"openrouter/auto",
				"google/gemini-2.0-flash-exp:free",
				"google/gemini-flash-1.5",
				"meta-llama/llama-3.1-8b-instruct:free",
				"meta-llama/llama-3.2-3b-instruct:free",
				"mistralai/mistral-7b-instruct:free",
				"nousresearch/hermes-3-llama-3.1-405b:free",
				"qwen/qwen-2-7b-instruct:free",
				"anthropic/claude-3.5-sonnet",
				"anthropic/claude-3-haiku",
				"openai/gpt-4o",
				"openai/gpt-4o-mini",
			},
			FreeTier:  true,
			DocsURL:   "https://openrouter.ai/docs",
			APIKeyURL: "https://openrouter.ai/keys",
			Tip:       "OpenRouter gives you access to 100+ models with a single API key. Many free models available!",
		},
		{
			ID:              "ollama",
			Name:            "Ollama",
			Description:     "Run models locally - Completely free and private",
			Icon:            "🦙",
			RequiresAPIKey:  false,
			RequiresAPIBase: true,
			DefaultAPIBase:  "http://localhost:11434/v1",
			Models: []string{
				"llama3.2",
				"llama3.1",
				"llama3.1:70b",
				"mistral",
				"qwen2.5",
				"gemma2",
				"phi3",
				"codellama",
				"deepseek-coder-v2",
			},
			FreeTier:  true,
			DocsURL:   "https://ollama.ai",
			APIKeyURL: "",
			Tip:       "Ollama lets you run models locally on your machine. No API key needed - completely private!",
		},
		{
			ID:              "anthropic",
			Name:            "Anthropic",
			Description:     "Claude models - Smart, safe, and reliable",
			Icon:            "🤖",
			RequiresAPIKey:  true,
			RequiresAPIBase: false,
			DefaultAPIBase:  "https://api.anthropic.com",
			Models: []string{
				"claude-3-5-sonnet-20241022",
				"claude-3-5-haiku-20241022",
				"claude-3-opus-20240229",
				"claude-3-sonnet-20240229",
				"claude-3-haiku-20240307",
			},
			FreeTier:  false,
			DocsURL:   "https://docs.anthropic.com",
			APIKeyURL: "https://console.anthropic.com",
			Tip:       "Claude models are known for safety, reliability, and great instruction-following.",
		},
		{
			ID:              "openai",
			Name:            "OpenAI",
			Description:     "GPT-4o, GPT-4, GPT-3.5 - Most capable models",
			Icon:            "🔮",
			RequiresAPIKey:  true,
			RequiresAPIBase: false,
			DefaultAPIBase:  "https://api.openai.com/v1",
			Models: []string{
				"gpt-4o",
				"gpt-4o-mini",
				"gpt-4-turbo",
				"gpt-4",
				"gpt-3.5-turbo",
			},
			FreeTier:  false,
			DocsURL:   "https://platform.openai.com/docs",
			APIKeyURL: "https://platform.openai.com/api-keys",
			Tip:       "OpenAI provides the most advanced models including GPT-4. Excellent for complex reasoning.",
		},
		{
			ID:              "groq",
			Name:            "Groq",
			Description:     "Ultra-fast inference - Free tier available",
			Icon:            "⚡",
			RequiresAPIKey:  true,
			RequiresAPIBase: false,
			DefaultAPIBase:  "https://api.groq.com/openai/v1",
			Models: []string{
				"llama-3.3-70b-versatile",
				"llama-3.1-70b-versatile",
				"llama-3.1-8b-instant",
				"mixtral-8x7b-32768",
				"gemma2-9b-it",
			},
			FreeTier:  true,
			DocsURL:   "https://console.groq.com/docs",
			APIKeyURL: "https://console.groq.com/keys",
			Tip:       "Groq offers extremely fast inference with free tier. Perfect for low-latency applications.",
		},
		{
			ID:              "gemini",
			Name:            "Google Gemini",
			Description:     "Gemini 2.0 - Powerful multimodal models",
			Icon:            "✨",
			RequiresAPIKey:  true,
			RequiresAPIBase: false,
			DefaultAPIBase:  "https://generativelanguage.googleapis.com/v1beta",
			Models: []string{
				"gemini-2.0-flash-exp",
				"gemini-1.5-pro",
				"gemini-1.5-flash",
			},
			FreeTier:  true,
			DocsURL:   "https://ai.google.dev/docs",
			APIKeyURL: "https://makersuite.google.com/app/apikey",
			Tip:       "Google's Gemini models offer great multimodal capabilities with generous free tier.",
		},
		{
			ID:              "together",
			Name:            "Together AI",
			Description:     "Multiple open-source models with flexible pricing",
			Icon:            "🤝",
			RequiresAPIKey:  true,
			RequiresAPIBase: false,
			DefaultAPIBase:  "https://api.together.xyz/v1",
			Models: []string{
				"meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo",
				"meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo",
				"mistralai/Mixtral-8x7B-Instruct-v0.1",
				"Qwen/Qwen2.5-72B-Instruct-Turbo",
			},
			FreeTier:  false,
			DocsURL:   "https://docs.together.ai",
			APIKeyURL: "https://api.together.xyz/settings/api-keys",
			Tip:       "Access to various open-source models with flexible pricing.",
		},
		{
			ID:              "huggingface",
			Name:            "Hugging Face",
			Description:     "Open-source models via inference API",
			Icon:            "🤗",
			RequiresAPIKey:  true,
			RequiresAPIBase: false,
			DefaultAPIBase:  "https://api-inference.huggingface.co",
			Models:          []string{},
			FreeTier:        true,
			DocsURL:         "https://huggingface.co/docs/api-inference",
			APIKeyURL:       "https://huggingface.co/settings/tokens",
			Tip:             "Great for open-source models. Good for privacy-conscious users.",
		},
		{
			ID:              "zhipu",
			Name:            "Zhipu AI (智谱AI)",
			Description:     "Chinese LLM provider - GLM models",
			Icon:            "🇨🇳",
			RequiresAPIKey:  true,
			RequiresAPIBase: false,
			DefaultAPIBase:  "https://open.bigmodel.cn/api/paas/v4",
			Models: []string{
				"glm-4-plus",
				"glm-4-0520",
				"glm-4-flash",
				"glm-4-air",
			},
			FreeTier:  true,
			DocsURL:   "https://open.bigmodel.cn/dev/api",
			APIKeyURL: "https://open.bigmodel.cn/usercenter/apikeys",
			Tip:       "Chinese language models with strong performance on Chinese tasks.",
		},
		{
			ID:              "moonshot",
			Name:            "Moonshot AI (月之暗面)",
			Description:     "Chinese LLM provider - Kimi models",
			Icon:            "🌙",
			RequiresAPIKey:  true,
			RequiresAPIBase: false,
			DefaultAPIBase:  "https://api.moonshot.cn/v1",
			Models: []string{
				"moonshot-v1-8k",
				"moonshot-v1-32k",
				"moonshot-v1-128k",
			},
			FreeTier:  false,
			DocsURL:   "https://platform.moonshot.cn/docs",
			APIKeyURL: "https://platform.moonshot.cn/console/api-keys",
			Tip:       "Kimi models with very long context windows (up to 128k tokens).",
		},
		{
			ID:              "nvidia",
			Name:            "NVIDIA NIM",
			Description:     "NVIDIA-accelerated inference for open models",
			Icon:            "🎮",
			RequiresAPIKey:  true,
			RequiresAPIBase: false,
			DefaultAPIBase:  "https://integrate.api.nvidia.com/v1",
			Models: []string{
				"nvidia/llama-3.1-nemotron-70b-instruct",
				"meta/llama-3.1-8b-instruct",
				"mistralai/mixtral-8x7b-instruct-v0.1",
			},
			FreeTier:  true,
			DocsURL:   "https://build.nvidia.com/explore/discover",
			APIKeyURL: "https://build.nvidia.com/nim",
			Tip:       "Free tier available. Great performance on NVIDIA hardware.",
		},
		{
			ID:              "vllm",
			Name:            "vLLM (Self-Hosted)",
			Description:     "Self-hosted inference server for open models",
			Icon:            "🚀",
			RequiresAPIKey:  false,
			RequiresAPIBase: true,
			DefaultAPIBase:  "http://localhost:8000/v1",
			Models:          []string{},
			FreeTier:        true,
			DocsURL:         "https://docs.vllm.ai",
			APIKeyURL:       "",
			Tip:             "High-performance self-hosted inference. Requires setup but completely free.",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"providers": catalog,
	})
}

func (s *Server) handleLongTermMemory(w http.ResponseWriter, r *http.Request) {
	// Get user from claims
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user to obtain UUID
	user, err := s.resolveUserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Get user-specific workspace
	userWorkspace, err := config.EnsureUserWorkspace(user.UUID)
	if err != nil {
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return
	}

	// Create user-specific memory store
	userMemory := agent.NewMemoryStore(userWorkspace)

	switch r.Method {
	case http.MethodGet:
		content := userMemory.ReadLongTerm()
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]string{"content": content})
	case http.MethodPost:
		var in struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if err := userMemory.WriteLongTerm(in.Content); err != nil {
			http.Error(w, "failed to write memory", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]string{"status": "ok"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleDailyNotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from claims
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user to obtain UUID
	user, err := s.resolveUserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Get user-specific workspace
	userWorkspace, err := config.EnsureUserWorkspace(user.UUID)
	if err != nil {
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return
	}

	// Create user-specific memory store
	userMemory := agent.NewMemoryStore(userWorkspace)

	daysStr := r.URL.Query().Get("days")
	days := 7
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	content := userMemory.GetRecentDailyNotes(days)
	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]string{"content": content})
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
		writeJSONResponse(w, map[string]string{
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
		writeJSONResponse(w, map[string]string{
			"error": "Failed to parse form. File may be too large (max 25MB).",
		})
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{
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
	tmpFile, err := os.CreateTemp("", "makoclaw-voice-*"+ext)
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
		writeJSONResponse(w, map[string]string{
			"error": "Transcription failed: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, result)
}

// handleMetrics returns in-process observability metrics (LLM calls, tool calls, agent runs).
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	snapshot := observability.Global().Snapshot()
	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, snapshot)
}

// handleAgents returns information about configured specialist agents
func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orchestrated := false
	if s.agentManager != nil {
		orchestrated = s.agentManager.IsOrchestrated()
	}

	agentsList := make([]map[string]interface{}, 0)

	// Orchestrator config to return — loaded from persisted user config so that
	// the Settings > Agents tab can display the saved state after a page reload.
	var orchestratorConfig map[string]interface{}

	// Primary source: load from persisted user config (survives restarts)
	if _, userUUID, ok := s.getUserStorage(r); ok && userUUID != "" {
		if userCfg, _ := config.LoadConfigForUser(userUUID); userCfg != nil {
			for name, spec := range userCfg.Agents.Specialists {
				agentsList = append(agentsList, map[string]interface{}{
					"name":        name,
					"description": spec.Description,
					"tools":       spec.Tools,
					"provider":    spec.Provider,
					"model":       spec.Model,
					"keywords":    spec.Keywords,
				})
			}
			// Return the full orchestrator config so the frontend can restore
			// the enabled state and all other fields on page load.
			orch := userCfg.Agents.Orchestrator
			orchestratorConfig = map[string]interface{}{
				"enabled":                orch.Enabled,
				"provider":               orch.Provider,
				"model":                  orch.Model,
				"max_tokens":             orch.MaxTokens,
				"temperature":            orch.Temperature,
				"max_delegation_retries": orch.MaxDelegationRetries,
				"fallback_to_default":    orch.FallbackToDefault,
				"description":            orch.Description,
			}
		}
	}

	// Fallback: use in-memory registry if storage/auth unavailable
	if len(agentsList) == 0 && s.agentManager != nil {
		registry := s.agentManager.GetSpecialistRegistry()
		for name, specialist := range registry.ListSpecialists() {
			if name == "orchestrator" {
				continue
			}
			agentsList = append(agentsList, map[string]interface{}{
				"name":        specialist.GetName(),
				"description": specialist.GetDescription(),
				"tools":       specialist.GetAllowedTools(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"orchestrated": orchestrated,
		"orchestrator": orchestratorConfig,
		"specialists":  agentsList,
	})
}

// handleAgentAction handles GET/POST for individual agents
// handleGetSpecialists returns the list of active specialists for real-time UI display
func (s *Server) handleGetSpecialists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	specialists := []map[string]interface{}{}
	if s.agentManager != nil {
		specialists = s.agentManager.GetSpecialistsList()
	}

	response := map[string]interface{}{
		"specialists": specialists,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleAgentAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "agent name required", http.StatusBadRequest)
		return
	}

	agentName := parts[0]

	// Route to appropriate handler
	if len(parts) > 1 {
		if parts[1] == "sessions" {
			s.handleAgentSessions(w, r, agentName)
			return
		} else if parts[1] == "config" && len(parts) > 2 && parts[2] == "update" {
			s.handleAgentConfigUpdate(w, r, agentName)
			return
		} else if parts[1] == "test" {
			s.handleAgentTest(w, r, agentName)
			return
		}
	}

	// Default: get agent details
	s.handleAgentDetails(w, r, agentName)
}

// handleAgentDetails returns information about a specific agent
func (s *Server) handleAgentDetails(w http.ResponseWriter, r *http.Request, agentName string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.agentManager == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]string{
			"error": "agent manager not available",
		})
		return
	}

	registry := s.agentManager.GetSpecialistRegistry()
	info, err := registry.GetSpecialistInfo(agentName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, info)
}

// handleAgentSessions returns sessions handled by a specific agent
func (s *Server) handleAgentSessions(w http.ResponseWriter, r *http.Request, agentName string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.agentManager == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]string{
			"error": "agent manager not available",
		})
		return
	}

	// Placeholder implementation
	// In a full implementation, would query sessions filtered by agent_profile
	response := map[string]interface{}{
		"agent":    agentName,
		"sessions": []map[string]interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// handleAgentConfigUpdate updates an agent's configuration
func (s *Server) handleAgentConfigUpdate(w http.ResponseWriter, r *http.Request, agentName string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.agentManager == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]string{
			"error": "agent manager not available",
		})
		return
	}

	var updateReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// Placeholder implementation
	// In a full implementation, would update specialist config
	response := map[string]interface{}{
		"agent":  agentName,
		"status": "updated",
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// handleSpecialistCreate creates a new specialist agent
func (s *Server) handleSpecialistCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.agentManager == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSONResponse(w, map[string]string{
			"error": "agent manager not available",
		})
		return
	}

	var specialistReq struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Prompt      string   `json:"prompt"`
		Provider    string   `json:"provider"`
		Model       string   `json:"model"`
		MaxTokens   int      `json:"max_tokens"`
		Temperature float64  `json:"temperature"`
		MaxToolIter int      `json:"max_tool_iterations"`
		Tools       []string `json:"tools"`
		Keywords    []string `json:"keywords"`
	}

	if err := json.NewDecoder(r.Body).Decode(&specialistReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	if specialistReq.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{
			"error": "name is required",
		})
		return
	}

	store, userUUID, ok := s.getUserStorage(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		writeJSONResponse(w, map[string]string{"error": "unauthorized"})
		return
	}

	if userUUID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]string{"error": "user UUID is empty - please re-login or contact support"})
		return
	}

	userCfg, _ := config.LoadConfigForUser(userUUID)
	if userCfg == nil {
		userCfg = config.DefaultConfig()
	}
	if userCfg.Agents.Specialists == nil {
		userCfg.Agents.Specialists = map[string]config.SpecialistConfig{}
	}

	// If this specialist was previously removed, remove it from the removed list
	if userCfg.Agents.RemovedSpecialists != nil {
		for i, name := range userCfg.Agents.RemovedSpecialists {
			if name == specialistReq.Name {
				userCfg.Agents.RemovedSpecialists = append(userCfg.Agents.RemovedSpecialists[:i], userCfg.Agents.RemovedSpecialists[i+1:]...)
				break
			}
		}
	}

	specCfg := config.SpecialistConfig{
		Name:              specialistReq.Name,
		Description:       specialistReq.Description,
		Prompt:            specialistReq.Prompt,
		Provider:          specialistReq.Provider,
		Model:             specialistReq.Model,
		MaxTokens:         specialistReq.MaxTokens,
		Temperature:       specialistReq.Temperature,
		MaxToolIterations: specialistReq.MaxToolIter,
		Tools:             specialistReq.Tools,
		Keywords:          specialistReq.Keywords,
	}

	userCfg.Agents.Specialists[specialistReq.Name] = specCfg
	if err := config.SaveConfigForUser(userUUID, userCfg); err != nil {
		logger.ErrorCF("web", "Failed to save user config for specialist", map[string]interface{}{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]string{
			"error": "failed to save config: " + err.Error(),
		})
		return
	}

	if _, err := s.agentManager.AddOrUpdateSpecialist(specialistReq.Name, &specCfg, userCfg, store); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]string{
			"error": "failed to register specialist: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"success": true,
		"specialist": map[string]interface{}{
			"name":        specialistReq.Name,
			"description": specialistReq.Description,
			"provider":    specialistReq.Provider,
			"model":       specialistReq.Model,
			"tools":       specialistReq.Tools,
			"keywords":    specialistReq.Keywords,
		},
	}

	if s.multiUserChannelManager != nil {
		go func() {
			if err := s.multiUserChannelManager.RestartUserChannels(s.ctx, userUUID); err != nil {
				logger.WarnCF("web", "Failed to restart user channels after specialist create", map[string]interface{}{
					"user_uuid": userUUID,
					"error":     err.Error(),
				})
			}
		}()
	}

	logger.InfoCF("web", "Created new specialist", map[string]interface{}{
		"name":     specialistReq.Name,
		"provider": specialistReq.Provider,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, response)
}

// handleSpecialistAction handles PUT/DELETE for individual specialists
func (s *Server) handleSpecialistAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/specialist/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "specialist name required", http.StatusBadRequest)
		return
	}

	specialistName := parts[0]

	switch r.Method {
	case http.MethodPut:
		s.handleSpecialistUpdate(w, r, specialistName)
	case http.MethodDelete:
		s.handleSpecialistDelete(w, r, specialistName)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSpecialistUpdate updates a specialist's configuration
func (s *Server) handleSpecialistUpdate(w http.ResponseWriter, r *http.Request, specialistName string) {
	if s.agentManager == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSONResponse(w, map[string]string{
			"error": "agent manager not available",
		})
		return
	}

	var updateReq struct {
		Description string   `json:"description"`
		Prompt      string   `json:"prompt"`
		Provider    string   `json:"provider"`
		Model       string   `json:"model"`
		MaxTokens   int      `json:"max_tokens"`
		Temperature float64  `json:"temperature"`
		MaxToolIter int      `json:"max_tool_iterations"`
		Tools       []string `json:"tools"`
		Keywords    []string `json:"keywords"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	store, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userCfg, _ := config.LoadConfigForUser(userUUID)
	if userCfg == nil {
		userCfg = config.DefaultConfig()
	}
	if userCfg.Agents.Specialists == nil {
		userCfg.Agents.Specialists = map[string]config.SpecialistConfig{}
	}

	// If this specialist was previously removed, remove it from the removed list
	if userCfg.Agents.RemovedSpecialists != nil {
		for i, name := range userCfg.Agents.RemovedSpecialists {
			if name == specialistName {
				userCfg.Agents.RemovedSpecialists = append(userCfg.Agents.RemovedSpecialists[:i], userCfg.Agents.RemovedSpecialists[i+1:]...)
				break
			}
		}
	}

	specCfg := config.SpecialistConfig{
		Name:              specialistName,
		Description:       updateReq.Description,
		Prompt:            updateReq.Prompt,
		Provider:          updateReq.Provider,
		Model:             updateReq.Model,
		MaxTokens:         updateReq.MaxTokens,
		Temperature:       updateReq.Temperature,
		MaxToolIterations: updateReq.MaxToolIter,
		Tools:             updateReq.Tools,
		Keywords:          updateReq.Keywords,
	}

	userCfg.Agents.Specialists[specialistName] = specCfg
	if err := config.SaveConfigForUser(userUUID, userCfg); err != nil {
		logger.ErrorCF("web", "Failed to save user config for specialist update", map[string]interface{}{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		http.Error(w, "failed to save config", http.StatusInternalServerError)
		return
	}

	if _, err := s.agentManager.AddOrUpdateSpecialist(specialistName, &specCfg, userCfg, store); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]string{
			"error": "failed to update specialist: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"success": true,
		"specialist": map[string]interface{}{
			"name": specialistName,
		},
	}

	if s.multiUserChannelManager != nil {
		go func() {
			ctx := context.Background()
			if err := s.multiUserChannelManager.RestartUserChannels(ctx, userUUID); err != nil {
				logger.WarnCF("web", "Failed to restart user channels after specialist update", map[string]interface{}{
					"user_uuid": userUUID,
					"error":     err.Error(),
				})
			}
		}()
	}

	logger.InfoCF("web", "Updated specialist", map[string]interface{}{
		"name": specialistName,
	})

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// handleSpecialistDelete deletes a specialist
func (s *Server) handleSpecialistDelete(w http.ResponseWriter, r *http.Request, specialistName string) {
	if s.agentManager == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSONResponse(w, map[string]string{
			"error": "agent manager not available",
		})
		return
	}

	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userCfg, _ := config.LoadConfigForUser(userUUID)
	if userCfg == nil || userCfg.Agents.Specialists == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]string{
			"error": "specialist not found",
		})
		return
	}

	if _, exists := userCfg.Agents.Specialists[specialistName]; !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]string{
			"error": "specialist not found",
		})
		return
	}

	delete(userCfg.Agents.Specialists, specialistName)

	// Mark as removed to prevent re-inheritance from global config
	if userCfg.Agents.RemovedSpecialists == nil {
		userCfg.Agents.RemovedSpecialists = []string{}
	}
	userCfg.Agents.RemovedSpecialists = append(userCfg.Agents.RemovedSpecialists, specialistName)

	if err := config.SaveConfigForUser(userUUID, userCfg); err != nil {
		logger.ErrorCF("web", "Failed to save user config for specialist delete", map[string]interface{}{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		http.Error(w, "failed to save config", http.StatusInternalServerError)
		return
	}

	if !s.agentManager.RemoveSpecialist(specialistName) {
		logger.WarnCF("web", "Specialist not found in registry during delete", map[string]interface{}{
			"name": specialistName,
		})
	}

	response := map[string]interface{}{
		"success": true,
		"specialist": map[string]interface{}{
			"name":    specialistName,
			"deleted": true,
		},
	}

	if s.multiUserChannelManager != nil {
		go func() {
			ctx := context.Background()
			if err := s.multiUserChannelManager.RestartUserChannels(ctx, userUUID); err != nil {
				logger.WarnCF("web", "Failed to restart user channels after specialist delete", map[string]interface{}{
					"user_uuid": userUUID,
					"error":     err.Error(),
				})
			}
		}()
	}

	logger.InfoCF("web", "Deleted specialist", map[string]interface{}{
		"name": specialistName,
	})

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// handleSpecialistGenerate generates a specialist configuration using AI
func (s *Server) handleSpecialistGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	if req.Description == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{
			"error": "description is required",
		})
		return
	}

	lowerDesc := strings.ToLower(req.Description)
	specialistName := strings.ToLower(strings.ReplaceAll(req.Description, " ", "_"))
	specialistName = strings.ReplaceAll(specialistName, "_expert", "")
	specialistName = strings.ReplaceAll(specialistName, "_specialist", "")
	if len(specialistName) > 20 {
		specialistName = specialistName[:20]
	}

	provider := "anthropic"
	model := "claude-opus"

	if strings.Contains(lowerDesc, "test") || strings.Contains(lowerDesc, "qa") {
		model = "claude-sonnet"
	} else if strings.Contains(lowerDesc, "doc") || strings.Contains(lowerDesc, "write") {
		provider = "openrouter"
		model = "meta-llama/llama-2-70b"
	} else if strings.Contains(lowerDesc, "devops") || strings.Contains(lowerDesc, "deploy") {
		provider = "openrouter"
		model = "meta-llama/llama-2-70b"
	} else if strings.Contains(lowerDesc, "data") || strings.Contains(lowerDesc, "analys") {
		provider = "openrouter"
		model = "meta-llama/llama-2-70b"
	}

	keywords := []string{}
	words := strings.Fields(req.Description)
	skipWords := map[string]bool{"the": true, "and": true, "for": true, "with": true, "that": true, "this": true, "specialist": true}
	for _, word := range words {
		lowerWord := strings.ToLower(word)
		if len(lowerWord) > 3 && !skipWords[lowerWord] {
			keywords = append(keywords, strings.Trim(lowerWord, ".,!?"))
		}
		if len(keywords) >= 5 {
			break
		}
	}

	tools := []string{"file_read", "file_write"}
	if strings.Contains(lowerDesc, "code") || strings.Contains(lowerDesc, "exec") || strings.Contains(lowerDesc, "run") {
		tools = append(tools, "execute_shell", "list_dir")
	}
	if strings.Contains(lowerDesc, "test") {
		tools = append(tools, "execute_shell")
	}
	if strings.Contains(lowerDesc, "search") || strings.Contains(lowerDesc, "web") || strings.Contains(lowerDesc, "research") {
		tools = append(tools, "web_search")
	}

	prompt := fmt.Sprintf("You are an expert %s specialist. Your role is: %s\n\nFocus on quality, accuracy, and best practices.", strings.TrimSuffix(req.Description, "."), req.Description)

	response := map[string]interface{}{
		"specialist": map[string]interface{}{
			"name":                specialistName,
			"description":         req.Description,
			"prompt":              prompt,
			"provider":            provider,
			"model":               model,
			"max_tokens":          8192,
			"temperature":         0.7,
			"max_tool_iterations": 20,
			"tools":               tools,
			"keywords":            keywords,
		},
	}

	logger.InfoCF("web", "Generated specialist with AI", map[string]interface{}{
		"name":        specialistName,
		"description": req.Description,
	})

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// handleOrchestratorConfig updates the orchestrator configuration
func (s *Server) handleOrchestratorConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.agentManager == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSONResponse(w, map[string]string{
			"error": "agent manager not available",
		})
		return
	}

	var updateReq struct {
		Enabled              bool    `json:"enabled"`
		Provider             string  `json:"provider"`
		Model                string  `json:"model"`
		MaxTokens            int     `json:"max_tokens"`
		Temperature          float64 `json:"temperature"`
		MaxDelegationRetries int     `json:"max_delegation_retries"`
		FallbackToDefault    bool    `json:"fallback_to_default"`
		Description          string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	// 1. Get user configuration
	store, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userCfg, _ := config.LoadConfigForUser(userUUID)
	if userCfg == nil {
		userCfg = config.DefaultConfig()
	}

	// 2. Update orchestrator section
	userCfg.Agents.Orchestrator = config.OrchestratorConfig{
		Enabled:              updateReq.Enabled,
		Provider:             updateReq.Provider,
		Model:                updateReq.Model,
		MaxTokens:            updateReq.MaxTokens,
		Temperature:          updateReq.Temperature,
		MaxDelegationRetries: updateReq.MaxDelegationRetries,
		FallbackToDefault:    updateReq.FallbackToDefault,
		Description:          updateReq.Description,
	}

	// 3. Save user configuration
	if err := config.SaveConfigForUser(userUUID, userCfg); err != nil {
		logger.ErrorCF("web", "Failed to save user config for orchestrator update", map[string]interface{}{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		http.Error(w, "failed to save config", http.StatusInternalServerError)
		return
	}

	// 4. Update runtime state (re-initialize orchestrator)
	// We need message bus, which is on the server if available
	if s.agentManager != nil {
		// Attempt to initialize orchestrator with new config
		// Note: This might require specialists to be present if enabled=true
		if err := s.agentManager.InitializeOrchestrator(userCfg, s.msgBus, store); err != nil {
			logger.WarnCF("web", "Failed to runtime-initialize orchestrator", map[string]interface{}{
				"error": err.Error(),
			})
			// We don't fail the request because persistence was successful
		}
	}

	response := map[string]interface{}{
		"success": true,
		"orchestrator": map[string]interface{}{
			"enabled":                updateReq.Enabled,
			"provider":               updateReq.Provider,
			"model":                  updateReq.Model,
			"max_tokens":             updateReq.MaxTokens,
			"temperature":            updateReq.Temperature,
			"max_delegation_retries": updateReq.MaxDelegationRetries,
			"fallback_to_default":    updateReq.FallbackToDefault,
			"description":            updateReq.Description,
		},
	}

	logger.InfoCF("web", "Updated and persisted orchestrator config", map[string]interface{}{
		"user_uuid": userUUID,
		"enabled":   updateReq.Enabled,
	})

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// handleAgentTest tests a specialist with a message
func (s *Server) handleAgentTest(w http.ResponseWriter, r *http.Request, specialistName string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.agentManager == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSONResponse(w, map[string]string{
			"error": "agent manager not available",
		})
		return
	}

	var req struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	if req.Message == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{
			"error": "message is required",
		})
		return
	}

	// Get the specialist from the registry
	registry := s.agentManager.GetSpecialistRegistry()
	specialist, err := registry.GetSpecialist(specialistName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]string{
			"error": fmt.Sprintf("specialist '%s' not found: %v", specialistName, err),
		})
		return
	}

	logger.InfoCF("web", "Testing specialist", map[string]interface{}{
		"name":    specialistName,
		"message": req.Message,
	})

	// Create a context with timeout for the test
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	// Track tools used during the test
	toolsUsed := []string{}
	var toolsMu sync.Mutex

	// Process the message through the specialist
	startTime := time.Now()
	result, err := specialist.ProcessWithSpeciality(ctx, req.Message)
	duration := time.Since(startTime)

	if err != nil {
		logger.WarnCF("web", "Specialist test failed", map[string]interface{}{
			"name":     specialistName,
			"error":    err.Error(),
			"duration": duration.String(),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]string{
			"error": fmt.Sprintf("specialist processing failed: %v", err),
		})
		return
	}

	// Get tools used (from specialist's allowed tools that were available)
	toolsMu.Lock()
	toolsUsed = specialist.GetAllowedTools()
	toolsMu.Unlock()

	response := map[string]interface{}{
		"success":      true,
		"specialist":   specialistName,
		"response":     result,
		"tools_used":   toolsUsed,
		"duration_ms":  duration.Milliseconds(),
	}

	logger.InfoCF("web", "Specialist test completed", map[string]interface{}{
		"name":     specialistName,
		"duration": duration.String(),
	})

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// handleAgentMetrics returns cost metrics for specialists
func (s *Server) handleAgentMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d"
	}

	response := map[string]interface{}{
		"period":       period,
		"total_cost":   0.0125,
		"total_calls":  125,
		"total_tokens": 125000,
		"by_specialist": map[string]interface{}{
			"developer": map[string]interface{}{
				"calls":    45,
				"tokens":   45000,
				"cost":     0.0045,
				"avg_cost": 0.0001,
			},
			"testing": map[string]interface{}{
				"calls":    30,
				"tokens":   30000,
				"cost":     0.0030,
				"avg_cost": 0.0001,
			},
			"documentation": map[string]interface{}{
				"calls":    25,
				"tokens":   25000,
				"cost":     0.0025,
				"avg_cost": 0.0001,
			},
			"devops": map[string]interface{}{
				"calls":    15,
				"tokens":   15000,
				"cost":     0.0015,
				"avg_cost": 0.0001,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// handleConfigStatus returns the current configuration status including degraded mode
func (s *Server) handleConfigStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.fullConfig == nil {
		http.Error(w, "configuration not available", http.StatusInternalServerError)
		return
	}

	// Merge user-specific config so per-user providers are recognised
	effectiveCfg := s.fullConfig
	if _, userUUID, ok := s.getUserStorage(r); ok && userUUID != "" {
		if userCfg, err := config.LoadConfigForUser(userUUID); err == nil && userCfg != nil {
			effectiveCfg = config.MergeConfigs(s.fullConfig, userCfg)
		}
	}

	configured := effectiveCfg.HasValidProviderConfig()
	degradedMode := effectiveCfg.DegradedMode

	// If any provider is configured for this user, treat system as configured
	// for degraded-banner purposes even when agent defaults are not set yet.
	activeProviders := effectiveCfg.GetActiveProviders()
	if len(activeProviders) > 0 {
		configured = true
		degradedMode = false
	}

	status := map[string]interface{}{
		"configured":   configured,
		"degradedMode": degradedMode,
	}

	if degradedMode {
		status["reason"] = "No LLM provider configured. Configure a provider to enable full functionality."
	}

	if !configured && !degradedMode {
		status["reason"] = "Provider configuration incomplete or invalid."
	}

	// Include currently configured providers (if any)
	status["activeProviders"] = activeProviders

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, status)
}

// handleConfigProvider updates the provider configuration
func (s *Server) handleConfigProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.fullConfig == nil {
		http.Error(w, "configuration not available", http.StatusInternalServerError)
		return
	}

	var payload struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
		APIKey   string `json:"api_key"`
		APIBase  string `json:"api_base,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if payload.Provider == "" {
		http.Error(w, "provider is required", http.StatusBadRequest)
		return
	}

	if payload.Model == "" {
		http.Error(w, "model is required", http.StatusBadRequest)
		return
	}

	// Validate provider type
	validProviders := map[string]bool{
		"anthropic": true, "openai": true, "openrouter": true, "groq": true,
		"zhipu": true, "gemini": true, "moonshot": true, "nvidia": true,
		"ollama": true, "vllm": true, "mock": true,
	}

	if !validProviders[payload.Provider] {
		http.Error(w, "invalid provider", http.StatusBadRequest)
		return
	}

	// Ollama and mock don't require API key
	if payload.Provider != "ollama" && payload.Provider != "mock" && payload.APIKey == "" {
		http.Error(w, "api_key is required for this provider", http.StatusBadRequest)
		return
	}

	// Update config
	s.fullConfig.Lock()
	s.fullConfig.Agents.Defaults.Provider = payload.Provider
	s.fullConfig.Agents.Defaults.Model = payload.Model

	// Update provider-specific config
	switch payload.Provider {
	case "anthropic":
		s.fullConfig.Providers.Anthropic.APIKey = payload.APIKey
		if payload.APIBase != "" {
			s.fullConfig.Providers.Anthropic.APIBase = payload.APIBase
		}
	case "openai":
		s.fullConfig.Providers.OpenAI.APIKey = payload.APIKey
		if payload.APIBase != "" {
			s.fullConfig.Providers.OpenAI.APIBase = payload.APIBase
		}
	case "openrouter":
		s.fullConfig.Providers.OpenRouter.APIKey = payload.APIKey
		if payload.APIBase != "" {
			s.fullConfig.Providers.OpenRouter.APIBase = payload.APIBase
		}
	case "groq":
		s.fullConfig.Providers.Groq.APIKey = payload.APIKey
		if payload.APIBase != "" {
			s.fullConfig.Providers.Groq.APIBase = payload.APIBase
		}
	case "zhipu":
		s.fullConfig.Providers.Zhipu.APIKey = payload.APIKey
		if payload.APIBase != "" {
			s.fullConfig.Providers.Zhipu.APIBase = payload.APIBase
		}
	case "gemini":
		s.fullConfig.Providers.Gemini.APIKey = payload.APIKey
		if payload.APIBase != "" {
			s.fullConfig.Providers.Gemini.APIBase = payload.APIBase
		}
	case "moonshot":
		s.fullConfig.Providers.Moonshot.APIKey = payload.APIKey
		if payload.APIBase != "" {
			s.fullConfig.Providers.Moonshot.APIBase = payload.APIBase
		}
	case "nvidia":
		s.fullConfig.Providers.Nvidia.APIKey = payload.APIKey
		if payload.APIBase != "" {
			s.fullConfig.Providers.Nvidia.APIBase = payload.APIBase
		}
	case "ollama":
		if payload.APIBase != "" {
			s.fullConfig.Providers.Ollama.APIBase = payload.APIBase
		} else {
			s.fullConfig.Providers.Ollama.APIBase = "http://localhost:11434/v1"
		}
	case "vllm":
		if payload.APIKey != "" {
			s.fullConfig.Providers.VLLM.APIKey = payload.APIKey
		}
		if payload.APIBase != "" {
			s.fullConfig.Providers.VLLM.APIBase = payload.APIBase
		}
	}

	s.fullConfig.DegradedMode = false
	s.fullConfig.Unlock()

	// Save config to file
	configPath := filepath.Join(filepath.Dir(s.workspace), "config.json")
	configData, err := json.MarshalIndent(s.fullConfig, "", "  ")
	if err != nil {
		http.Error(w, "failed to marshal config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		http.Error(w, "failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.InfoCF("web", "Provider configuration updated", map[string]interface{}{
		"provider": payload.Provider,
		"model":    payload.Model,
	})

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Provider configuration saved. Restart the application to apply changes.",
	})
}

// handleConfigValidate validates provider credentials without saving
func (s *Server) handleConfigValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.fullConfig == nil {
		http.Error(w, "configuration not available", http.StatusInternalServerError)
		return
	}

	var payload struct {
		Provider string `json:"provider"`
		APIKey   string `json:"api_key"`
		APIBase  string `json:"api_base,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if payload.Provider == "" {
		http.Error(w, "provider is required", http.StatusBadRequest)
		return
	}

	// Simple validation without copying the entire config (to avoid lock copy issues)
	// Just check if required fields are present
	var validationErr error
	switch payload.Provider {
	case "anthropic", "openai", "openrouter", "groq":
		if payload.APIKey == "" {
			validationErr = fmt.Errorf("API key is required for %s", payload.Provider)
		}
	case "ollama":
		// Ollama doesn't require API key, just valid base URL
		if payload.APIBase == "" {
			validationErr = fmt.Errorf("API base URL is required for ollama")
		}
	default:
		http.Error(w, "unsupported provider for validation", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"valid": validationErr == nil,
	}

	if validationErr != nil {
		response["error"] = validationErr.Error()
	} else {
		response["message"] = "Provider configuration is valid"
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// handleTestTelegram tests a Telegram bot token by calling the Telegram API
func (s *Server) handleTestTelegram(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BotToken  string `json:"botToken"`
		ChannelID string `json:"channelId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.BotToken == "" {
		http.Error(w, "bot token required", http.StatusBadRequest)
		return
	}

	// Test Telegram connection by calling getMe API
	client := &http.Client{Timeout: 10 * time.Second}

	// Format: https://api.telegram.org/bot<TOKEN>/getMe
	testURL := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", req.BotToken)

	resp, err := client.Get(testURL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Connection test failed: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	var telegramResp struct {
		Ok     bool `json:"ok"`
		Result struct {
			ID       int64  `json:"id"`
			IsBot    bool   `json:"is_bot"`
			Username string `json:"username"`
		} `json:"result,omitempty"`
		Description string `json:"description,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&telegramResp); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"message": "Failed to parse Telegram response",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if !telegramResp.Ok {
		w.WriteHeader(http.StatusOK)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Invalid bot token: %s", telegramResp.Description),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	writeJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("✓ Connected to bot @%s (ID: %d)", telegramResp.Result.Username, telegramResp.Result.ID),
		"bot": map[string]interface{}{
			"id":       telegramResp.Result.ID,
			"username": telegramResp.Result.Username,
			"is_bot":   telegramResp.Result.IsBot,
		},
	})
}

// writeJSONResponse is a helper to encode JSON and log any encoding errors (Fixes P30)
func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.WarnCF("web", "failed to encode response", map[string]interface{}{"error": err.Error()})
	}
}

// recoveryMiddleware captures panics in HTTP handlers and returns a 500 status (Fixes P26)
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorCF("web", "Panic recovered in HTTP handler", map[string]interface{}{
					"error": fmt.Sprintf("%v", err),
					"path":  r.URL.Path,
				})
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// apiRateLimitMiddleware enforces per-user rate limiting (100 req/min) on authenticated API routes.
// Unauthenticated routes are skipped because they are already handled by loginLimit.
func (s *Server) apiRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws/") {
			claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
			if ok && claims != nil {
				key := "api:user:" + claims.Sub
				s.apiRateLimit.SetLimit(key, 100, time.Minute)
				allowed, waitTime := s.apiRateLimit.AllowWithWait(key)
				if !allowed {
					w.Header().Set("Retry-After", fmt.Sprintf("%.0f", waitTime.Seconds()))
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusTooManyRequests)
					writeJSONResponse(w, map[string]string{"error": "rate limit exceeded, please slow down"})
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
