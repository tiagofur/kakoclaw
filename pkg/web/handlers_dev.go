package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/bridge"
)

type bridgeState struct {
	Status     string `json:"status"`
	ProjectDir string `json:"project_dir"`
	Backend    string `json:"backend"`
	UpdatedAt  string `json:"updated_at"`
}

func (s *Server) bridgeStateFile(userUUID string) string {
	storePath := ""
	if s.fullConfig != nil {
		storePath = s.fullConfig.Storage.Path
	}
	if storePath == "" {
		storePath = defaultWorkspace()
	}
	return filepath.Join(storePath, "bridge", userUUID+"-state.json")
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

	if s.fullConfig == nil || !s.fullConfig.DevStudio.Enabled {
		return nil, fmt.Errorf("dev studio is disabled globally")
	}

	// Resolve the actual bundle path on disk before creating the bridge.
	// EnsureBridge extracts the embedded JS bundle to storePath/bridge/ if missing or outdated.
	storePath := s.fullConfig.Storage.Path
	if storePath == "" {
		storePath = defaultWorkspace()
	}
	bridgeDir := filepath.Join(storePath, "bridge")

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	s.writeBridgeState(claims.UUID, "running", req.ProjectDir)
	writeJSONResponse(w, map[string]interface{}{"status": "started", "project_dir": req.ProjectDir})
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
		})
		return
	}

	state := b.State()
	projectDir := b.Cwd()
	// If bridge is not running, try to restore last known project from state.json
	if state != "running" {
		if saved := s.readBridgeState(claims.UUID); saved != nil && saved.ProjectDir != "" {
			projectDir = saved.ProjectDir
		}
	}
	writeJSONResponse(w, map[string]interface{}{
		"status":      state,
		"project_dir": projectDir,
	})
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
		Message string `json:"message"`
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
	sessionID := "dev_studio_" + projectName
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
		if canFlush {
			flusher.Flush()
		}
	}

	if userStorOK && userStore != nil && fullResponse.Len() > 0 {
		_ = userStore.SaveMessage(sessionID, "assistant", fullResponse.String())
	}
}
