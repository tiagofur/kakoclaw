package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sipeed/makoclaw/pkg/bridge"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
)

type bridgeState struct {
	Status     string `json:"status"`
	ProjectDir string `json:"project_dir"`
	Backend    string `json:"backend"`
	UpdatedAt  string `json:"updated_at"`
}

func (s *Server) bridgeStateFile(userUUID string) string {
	return filepath.Join(s.devWorkspaceDir(), "bridge", userUUID+"-state.json")
}

// devWorkspaceDir returns the workspace directory for dev studio artifacts.
func (s *Server) devWorkspaceDir() string {
	if s.fullConfig != nil {
		if ws := s.fullConfig.WorkspacePath(); ws != "" {
			return ws
		}
	}
	return defaultWorkspace()
}

func (s *Server) writeBridgeState(userUUID, status, projectDir string) {
	st := bridgeState{
		Status:     status,
		ProjectDir: projectDir,
		Backend:    "claude-code",
		UpdatedAt:  time.Now().Format(time.RFC3339),
	}
	if s.fullConfig != nil && s.fullConfig.DevStudio.DefaultBackend != "" {
		st.Backend = s.fullConfig.DevStudio.DefaultBackend
	}
	path := s.bridgeStateFile(userUUID)
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	if b, err := json.MarshalIndent(st, "", "  "); err == nil {
		_ = os.WriteFile(path, b, 0644)
	}
}

func (s *Server) readBridgeState(userUUID string) *bridgeState {
	data, err := os.ReadFile(s.bridgeStateFile(userUUID))
	if err != nil {
		return nil
	}
	var st bridgeState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil
	}
	return &st
}

// getDevBridge retrieves or creates a bridge instance for the given user.
func (s *Server) getDevBridge(userUUID string) (*bridge.Bridge, error) {
	s.devBridgesMu.RLock()
	if b, ok := s.devBridges[userUUID]; ok {
		s.devBridgesMu.RUnlock()
		return b.(*bridge.Bridge), nil
	}
	s.devBridgesMu.RUnlock()

	s.devBridgesMu.Lock()
	defer s.devBridgesMu.Unlock()

	// Double check
	if b, ok := s.devBridges[userUUID]; ok {
		return b.(*bridge.Bridge), nil
	}

	if s.fullConfig == nil {
		return nil, fmt.Errorf("server configuration not available")
	}
	// Check per-user config — DevStudio.Enabled is a user-level setting
	if userCfg, err := config.LoadConfigForUser(userUUID); err != nil || !userCfg.DevStudio.Enabled {
		return nil, fmt.Errorf("dev studio is not enabled for this user")
	}

	// Resolve the actual bundle path on disk before creating the bridge.
	// EnsureBridge extracts the embedded JS bundle to workspace/bridge/ if missing or outdated.
	bridgeDir := filepath.Join(s.devWorkspaceDir(), "bridge")

	backendName := s.fullConfig.DevStudio.DefaultBackend
	if backendName == "" {
		backendName = "claude-code"
	}

	var bundlePath string
	var bundleErr error
	if backendName == "opencode" {
		bundlePath, bundleErr = bridge.EnsureBridge(bridgeDir, bridge.EmbeddedOpenCodeJS, "opencode")
	} else {
		bundlePath, bundleErr = bridge.EnsureBridge(bridgeDir, bridge.EmbeddedBundleJS, "claude-code")
	}
	if bundleErr != nil {
		return nil, fmt.Errorf("failed to setup bridge bundle: %w", bundleErr)
	}

	nodePath := s.fullConfig.DevStudio.NodePath
	if nodePath == "" {
		nodePath = "node"
	}

	// Create new bridge
	cfg := bridge.BridgeConfig{
		Backend:    bundlePath,
		NodePath:   nodePath,
		MaxRetries: 3,
	}
	b := bridge.New(cfg)
	if b == nil {
		return nil, fmt.Errorf("failed to create bridge")
	}

	s.devBridges[userUUID] = b
	return b, nil
}

func (s *Server) handleDevProjects(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.extractClaims(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userWorkspace string
	if s.userMgr != nil {
		userWorkspace = s.userMgr.UserWorkspacePath(claims.UUID)
	} else {
		userWorkspace = s.fullConfig.WorkspacePath()
	}

	projectsSubdir := "repos"
	if s.fullConfig != nil && s.fullConfig.DevStudio.ProjectsDir != "" {
		projectsSubdir = s.fullConfig.DevStudio.ProjectsDir
	}
	projectsDir := filepath.Join(userWorkspace, projectsSubdir)

	// Ensure projects directory exists
	_ = os.MkdirAll(projectsDir, 0755)

	if r.Method == http.MethodPost {
		var req struct {
			Name    string `json:"name"`
			GitInit bool   `json:"git_init"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			http.Error(w, "Invalid body: 'name' required", http.StatusBadRequest)
			return
		}

		// Prevent path traversal
		projectName := filepath.Base(req.Name)
		newProjectPath := filepath.Join(projectsDir, projectName)

		if err := os.MkdirAll(newProjectPath, 0755); err != nil {
			http.Error(w, "Failed to create project directory", http.StatusInternalServerError)
			return
		}

		if req.GitInit {
			// Try to run git init
			cmd := exec.CommandContext(r.Context(), "git", "init")
			cmd.Dir = newProjectPath
			_ = cmd.Run()
		}

		writeJSONResponse(w, map[string]interface{}{
			"status": "created",
			"name":   projectName,
			"path":   newProjectPath,
		})
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		http.Error(w, "Failed to read projects directory", http.StatusInternalServerError)
		return
	}

	// System directories to ignore
	systemDirs := map[string]bool{
		"memory":   true,
		"sessions": true,
		"skills":   true,
		"cron":     true,
		"tasks":    true,
		"temp":     true,
	}

	var projects []map[string]interface{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || systemDirs[name] {
			continue
		}

		info, _ := entry.Info()
		projects = append(projects, map[string]interface{}{
			"name":    name,
			"path":    filepath.Join(projectsDir, name),
			"modTime": info.ModTime(),
		})
	}

	writeJSONResponse(w, map[string]interface{}{
		"projects":     projects,
		"projects_dir": projectsDir,
		"root":         userWorkspace,
	})
}

func (s *Server) handleDevBridgeStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := s.extractClaims(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ProjectDir string `json:"project_dir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	b, err := s.getDevBridge(claims.UUID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if b.State() == "running" {
		if b.Cwd() == req.ProjectDir {
			writeJSONResponse(w, map[string]interface{}{"status": "already running"})
			return
		}
		// Switch project: stop and restart with new dir
		_ = b.Stop()
	}

	// Update Cwd before starting
	b.SetCwd(req.ProjectDir)

	if err := b.Start(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Verify the bridge process is actually responsive before reporting success.
	pingCtx, pingCancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer pingCancel()
	if err := b.Ping(pingCtx); err != nil {
		_ = b.Stop()
		s.writeBridgeState(claims.UUID, "stopped", "")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Bridge started but is not responding (is the backend CLI installed?): %v", err),
		})
		return
	}

	backendName := s.resolveBackendName()
	s.writeBridgeState(claims.UUID, "running", req.ProjectDir)
	writeJSONResponse(w, map[string]interface{}{"status": "started", "project_dir": req.ProjectDir, "backend": backendName})
}

func (s *Server) handleDevBridgeStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := s.extractClaims(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	b, err := s.getDevBridge(claims.UUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := b.Stop(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeBridgeState(claims.UUID, "stopped", "")
	writeJSONResponse(w, map[string]interface{}{"status": "stopped"})
}

func (s *Server) resolveBackendName() string {
	if s.fullConfig != nil && s.fullConfig.DevStudio.DefaultBackend != "" {
		return s.fullConfig.DevStudio.DefaultBackend
	}
	return "claude-code"
}

func (s *Server) handleDevBridgeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := s.extractClaims(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	b, err := s.getDevBridge(claims.UUID)
	if err != nil {
		// Dev Studio disabled or bridge unavailable — return idle status gracefully
		writeJSONResponse(w, map[string]interface{}{
			"status":      "idle",
			"project_dir": "",
			"backend":     "",
		})
		return
	}

	state := b.State()
	projectDir := b.Cwd()
	backendName := s.resolveBackendName()
	// If bridge is not running, try to restore last known project from state.json
	if state != "running" {
		if saved := s.readBridgeState(claims.UUID); saved != nil {
			if saved.ProjectDir != "" {
				projectDir = saved.ProjectDir
			}
			if saved.Backend != "" {
				backendName = saved.Backend
			}
		}
	}
	writeJSONResponse(w, map[string]interface{}{
		"status":      state,
		"project_dir": projectDir,
		"backend":     backendName,
	})
}

func (s *Server) handleSessionStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := s.extractClaims(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := SessionStats{TokenLimit: 0}
	if s.sessionTracker != nil {
		stats = s.sessionTracker.Stats(claims.UUID, strings.TrimSpace(r.URL.Query().Get("session_id")))
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, stats)
}

func (s *Server) resolveDevSessionID(userUUID, requestedSessionID, projectName string) string {
	requestedSessionID = strings.TrimSpace(requestedSessionID)
	if requestedSessionID != "" {
		return requestedSessionID
	}

	defaultSessionID := "dev_studio_" + projectName
	if s.sessionTracker == nil {
		return defaultSessionID
	}

	activeSessionID := s.sessionTracker.Stats(userUUID, "").SessionID
	if activeSessionID == "" {
		return defaultSessionID
	}
	if strings.HasPrefix(activeSessionID, "dev_studio_") && activeSessionID != defaultSessionID {
		return defaultSessionID
	}
	return activeSessionID
}

func makeSessionResetEvent(newSessionID string) map[string]interface{} {
	return map[string]interface{}{
		"type":           bridge.EventSessionReset,
		"new_session_id": newSessionID,
		"message":        "Session auto-reset: token limit reached",
	}
}

func (s *Server) trackDevSessionEvent(ctx context.Context, userUUID string, sessionID *string, ev bridge.Event, b *bridge.Bridge, emit func(map[string]interface{}) error) {
	if s.sessionTracker == nil || sessionID == nil || strings.TrimSpace(*sessionID) == "" || ev.Type != "result" {
		return
	}

	s.sessionTracker.Record(userUUID, *sessionID, ev.NumTurns, ev.CostUSD)
	if !s.sessionTracker.ShouldReset(userUUID, *sessionID) {
		return
	}

	newSessionID := uuid.NewString()
	if err := emit(makeSessionResetEvent(newSessionID)); err != nil {
		logger.WarnCF("web", "failed to emit session reset event", map[string]interface{}{"user_uuid": userUUID, "error": err.Error()})
		return
	}

	if b != nil {
		if err := b.Stop(); err != nil {
			logger.WarnCF("web", "failed to stop bridge during session reset", map[string]interface{}{"user_uuid": userUUID, "error": err.Error()})
		}
		if err := b.Start(ctx); err != nil {
			logger.WarnCF("web", "failed to restart bridge during session reset", map[string]interface{}{"user_uuid": userUUID, "error": err.Error()})
		}
	}

	s.sessionTracker.ZeroSession(userUUID, newSessionID)
	*sessionID = newSessionID
}

// handleDevQuery is the HTTP fallback for the WebSocket terminal.
// POST /api/v1/dev/query — accepts {"message":"..."}, streams bridge events as NDJSON.
// Used when the client cannot maintain a WebSocket connection after 3 consecutive failures.
func (s *Server) handleDevQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := s.extractClaims(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Message   string `json:"message"`
		SessionID string `json:"session_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
		http.Error(w, "Invalid body: 'message' required", http.StatusBadRequest)
		return
	}

	b, err := s.getDevBridge(claims.UUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userStore, _, userStorOK := s.getUserStorage(r)
	projectName := filepath.Base(b.Cwd())
	sessionID := s.resolveDevSessionID(claims.UUID, req.SessionID, projectName)
	if userStorOK && userStore != nil {
		_ = userStore.SaveMessage(sessionID, "user", req.Message)
	}

	metadata := bridge.RequestOptions{}
	if devMem, errMem := s.getDevMemory(claims.UUID); errMem == nil {
		if injected, errInj := devMem.Inject(r.Context(), req.Message, 5); errInj == nil && injected != "" {
			metadata.PromptInjection = injected
		}
	}

	bridgeReq := bridge.Request{
		Command: "query",
		Prompt:  req.Message,
		Options: metadata,
	}

	ch, errExec := b.Execute(r.Context(), bridgeReq)
	if errExec != nil {
		http.Error(w, errExec.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, canFlush := w.(http.Flusher)

	enc := json.NewEncoder(w)
	var fullResponse strings.Builder
	for ev := range ch {
		if ev.Text != "" {
			fullResponse.WriteString(ev.Text)
		}
		if ev.Content != "" {
			fullResponse.WriteString(ev.Content)
		}
		_ = enc.Encode(eventToFrontend(ev))
		s.trackDevSessionEvent(r.Context(), claims.UUID, &sessionID, ev, b, func(msg map[string]interface{}) error {
			return enc.Encode(msg)
		})
		if canFlush {
			flusher.Flush()
		}
	}

	if userStorOK && userStore != nil && fullResponse.Len() > 0 {
		_ = userStore.SaveMessage(sessionID, "assistant", fullResponse.String())
	}
}
