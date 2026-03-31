package web

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sipeed/makoclaw/pkg/bridge"
	"github.com/sipeed/makoclaw/pkg/logger"
)

type devTerminalWSMessage struct {
	Type    string `json:"type"` // "prompt", "ping", "interrupt"
	Message string `json:"message,omitempty"`
}

func (s *Server) handleDevTerminalWS(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.extractClaims(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.ErrorCF("web", "dev-studio ws upgrade failed", map[string]interface{}{"err": err.Error()})
		return
	}
	defer conn.Close()

	b, err := s.getDevBridge(claims.UUID)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","error":"`+err.Error()+`"}`))
		return
	}

	var writeMu sync.Mutex
	safeWrite := func(msg []byte) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteMessage(websocket.TextMessage, msg)
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Keep alive
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := safeWrite([]byte(`{"type":"ping"}`)); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var in devTerminalWSMessage
		if err := json.Unmarshal(msgBytes, &in); err != nil {
			continue
		}

		if in.Type == "prompt" {
			if b.State() != "running" {
				_ = safeWrite([]byte(`{"type":"error","error":"bridge not running"}`))
				continue
			}

			// Persistence: get user storage
			userStore, _, userStorOK := s.getUserStorage(r)
			projectName := filepath.Base(b.Cwd())
			sessionID := "dev_studio_" + projectName

			if userStorOK && userStore != nil {
				_ = userStore.SaveMessage(sessionID, "user", in.Message)
			}

			// Add injected context from devmemory if enabled
			metadata := bridge.RequestOptions{}
			if devMem, errMap := s.getDevMemory(claims.UUID); errMap == nil {
				if injected, errInject := devMem.Inject(ctx, in.Message, 5); errInject == nil && injected != "" {
					metadata.PromptInjection = injected
				}
			}

			req := bridge.Request{
				Command: "query",
				Prompt:  in.Message,
				Options: metadata,
			}
			ch, errExec := b.Execute(r.Context(), req)
			if errExec != nil {
				out, _ := json.Marshal(map[string]interface{}{"type": "error", "error": errExec.Error()})
				_ = safeWrite(out)
				continue
			}

			// Stream events to websocket
			go func() {
				var fullResponse strings.Builder
				for ev := range ch {
					if ev.Message != "" {
						fullResponse.WriteString(ev.Message)
					}
					outBytes, err := json.Marshal(ev)
					if err == nil {
						if writeErr := safeWrite(outBytes); writeErr != nil {
							break
						}
					}
				}
				// Save assistant response once finished
				if userStorOK && userStore != nil && fullResponse.Len() > 0 {
					_ = userStore.SaveMessage(sessionID, "assistant", fullResponse.String())
				}
			}()
		}
	}
}
