package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	
	"github.com/sipeed/makoclaw/pkg/devmemory"
	"github.com/sipeed/makoclaw/pkg/logger"
)

func (s *Server) getDevMemory(userUUID string) (*devmemory.Store, error) {
	s.devBridgesMu.RLock()
	if store, ok := s.devMemories[userUUID]; ok {
		s.devBridgesMu.RUnlock()
		return store.(*devmemory.Store), nil
	}
	s.devBridgesMu.RUnlock()

	s.devBridgesMu.Lock()
	defer s.devBridgesMu.Unlock()

	if store, ok := s.devMemories[userUUID]; ok {
		return store.(*devmemory.Store), nil
	}

	if s.fullConfig == nil || !s.fullConfig.DevStudio.Enabled {
		return nil, fmt.Errorf("dev studio is disabled globally")
	}

	if !s.fullConfig.DevStudio.Memory.Enabled {
		return nil, fmt.Errorf("dev memory is disabled globally")
	}

	userWorkspace := filepath.Join(defaultWorkspace(), userUUID)
	if s.userMgr != nil {
		if storePath := s.fullConfig.Storage.Path; storePath != "" {
			userWorkspace = filepath.Join(storePath, "users", userUUID, "workspace")
		}
	}
	_ = os.MkdirAll(userWorkspace, 0755)

	dbPath := filepath.Join(userWorkspace, "dev_memory.db")
	modelDir := filepath.Join(userWorkspace, "models", "huggingface", s.fullConfig.DevStudio.Memory.ModelDir)

	var embedder devmemory.Embedder
	// Check if Hugot is available/enabled
	// If Provider is not defined in Memory config, we check Model
	if s.fullConfig.DevStudio.Memory.Model != "" {
		hugotInst, err := devmemory.NewHugotEmbedder(modelDir)
		if err != nil {
			logger.ErrorCF("web", "Failed to init devmemory hugot embedder", map[string]interface{}{"err": err.Error(), "model": modelDir})
		} else {
			embedder = hugotInst
		}
	}

	store, err := devmemory.NewStore(dbPath, embedder)
	if err != nil {
		return nil, fmt.Errorf("failed to open dev memory db: %w", err)
	}

	s.devMemories[userUUID] = store
	return store, nil
}

func (s *Server) handleDevMemorySearch(w http.ResponseWriter, r *http.Request) {
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
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	store, err := s.getDevMemory(claims.UUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	results, err := store.Search(r.Context(), req.Query, req.Limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]interface{}{"results": results})
}

func (s *Server) handleDevMemoryStore(w http.ResponseWriter, r *http.Request) {
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
		Content  string `json:"content"`
		Metadata string `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	store, err := s.getDevMemory(claims.UUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	err = store.Add(r.Context(), req.Content, req.Metadata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]interface{}{"status": "stored"})
}

func (s *Server) handleDevMemoryDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := s.extractClaims(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	store, err := s.getDevMemory(claims.UUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	err = store.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]interface{}{"status": "deleted"})
}
