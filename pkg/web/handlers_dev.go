package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sipeed/makoclaw/pkg/bridge"
)

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
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := s.extractClaims(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userWorkspace := filepath.Join(defaultWorkspace(), claims.UUID)
	if s.userMgr != nil {
		if storePath := s.fullConfig.Storage.Path; storePath != "" {
			userWorkspace = filepath.Join(storePath, "users", claims.UUID, "workspace")
		}
	}
	
	// Ensure workspace exists
	_ = os.MkdirAll(userWorkspace, 0755)

	entries, err := os.ReadDir(userWorkspace)
	if err != nil {
		http.Error(w, "Failed to read projects directory", http.StatusInternalServerError)
		return
	}

	var projects []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() {
			info, _ := entry.Info()
			projects = append(projects, map[string]interface{}{
				"name":    entry.Name(),
				"path":    filepath.Join(userWorkspace, entry.Name()),
				"modTime": info.ModTime(),
			})
		}
	}

	writeJSONResponse(w, map[string]interface{}{
		"projects": projects,
		"root":     userWorkspace,
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

	writeJSONResponse(w, map[string]interface{}{"status": "started"})
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]interface{}{
		"status": b.State(),
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
		_ = enc.Encode(ev)
		if canFlush {
			flusher.Flush()
		}
	}

	if userStorOK && userStore != nil && fullResponse.Len() > 0 {
		_ = userStore.SaveMessage(sessionID, "assistant", fullResponse.String())
	}
}
