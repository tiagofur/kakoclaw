package web

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sipeed/makoclaw/pkg/bridge"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/storage"
)

func TestServer_InitializesDevPipeline(t *testing.T) {
	s := NewServer(config.WebConfig{}, nil, nil)
	if s.devPipeline == nil {
		t.Fatalf("expected devPipeline to be initialized")
	}
	if len(s.devPipeline.pre) == 0 {
		t.Fatalf("expected MemoryInjectionMiddleware to be registered as first pre middleware")
	}
}

func TestHandleDevQuery_UsesPipelineBeforeExecute(t *testing.T) {
	t.Run("pipeline output is passed to bridge request", func(t *testing.T) {
		s, token, userUUID := newAuthenticatedDevServer(t)
		bridgeExecCalled := false

		s.devPipeline = NewDevPipeline()
		s.devPipeline.Use(func(ctx context.Context, req *DevRequest) (*DevRequest, error) {
			if req.UserUUID != userUUID {
				t.Fatalf("expected user uuid %q, got %q", userUUID, req.UserUUID)
			}
			req.PromptInjection = "pipeline-context"
			return req, nil
		})

		s.devBridgeExecs[userUUID] = func(ctx context.Context, req bridge.Request) (<-chan bridge.Event, error) {
			bridgeExecCalled = true
			if req.Options.PromptInjection != "pipeline-context" {
				t.Fatalf("expected PromptInjection from pipeline, got %q", req.Options.PromptInjection)
			}
			ch := make(chan bridge.Event, 1)
			ch <- bridge.Event{Type: "result", Content: "done"}
			close(ch)
			return ch, nil
		}

		body := bytes.NewBufferString(`{"message":"fix the bug"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/dev/query", body)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.handleDevQuery(w, req)

		if !bridgeExecCalled {
			t.Fatalf("expected bridge execute to be called")
		}
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d with body %s", w.Code, w.Body.String())
		}
	})

	t.Run("pipeline error aborts request before execute", func(t *testing.T) {
		s, token, userUUID := newAuthenticatedDevServer(t)
		bridgeExecCalled := false
		boom := errors.New("pipeline failed")

		s.devPipeline = NewDevPipeline()
		s.devPipeline.Use(func(ctx context.Context, req *DevRequest) (*DevRequest, error) {
			return nil, boom
		})

		s.devBridgeExecs[userUUID] = func(ctx context.Context, req bridge.Request) (<-chan bridge.Event, error) {
			bridgeExecCalled = true
			ch := make(chan bridge.Event)
			close(ch)
			return ch, nil
		}

		body := bytes.NewBufferString(`{"message":"fix the bug"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/dev/query", body)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		s.handleDevQuery(w, req)

		if bridgeExecCalled {
			t.Fatalf("expected bridge execute not to be called when pipeline fails")
		}
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 status, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), boom.Error()) {
			t.Fatalf("expected error body to contain %q, got %q", boom.Error(), w.Body.String())
		}
	})
}

func newAuthenticatedDevServer(t *testing.T) (*Server, string, string) {
	t.Helper()

	central, err := storage.NewCentral(filepath.Join(t.TempDir(), "central.db"))
	if err != nil {
		t.Fatalf("NewCentral: %v", err)
	}
	t.Cleanup(func() { _ = central.Close() })

	user, err := central.CreateUser("devuser", "password123", "user")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	authManager, err := newAuthManager(central, "admin", "password123", "24h")
	if err != nil {
		t.Fatalf("newAuthManager: %v", err)
	}

	token, err := authManager.signToken(user.Username, user.UUID, user.Role, user.TokenVersion)
	if err != nil {
		t.Fatalf("signToken: %v", err)
	}

	s := NewServer(config.WebConfig{}, nil, nil)
	s.SetCentralStorage(central)
	s.authManager = authManager
	s.SetFullConfig(&config.Config{DevStudio: config.DevStudioConfig{Enabled: true}})

	b := bridge.New(bridge.BridgeConfig{Backend: "test-backend", NodePath: "node"})
	b.SetCwd(filepath.Join(t.TempDir(), "project"))
	s.devBridges[user.UUID] = b

	return s, token, user.UUID
}

func TestHandleDevTerminalWS_UsesPipelineBeforeExecute(t *testing.T) {
	t.Run("pipeline output is passed to bridge request", func(t *testing.T) {
		s, token, userUUID := newAuthenticatedDevServer(t)
		bridgeExecCalled := make(chan bridge.Request, 1)

		s.devPipeline = NewDevPipeline()
		s.devPipeline.Use(func(ctx context.Context, req *DevRequest) (*DevRequest, error) {
			req.PromptInjection = "ws-pipeline-context"
			return req, nil
		})

		s.devBridgeExecs[userUUID] = func(ctx context.Context, req bridge.Request) (<-chan bridge.Event, error) {
			bridgeExecCalled <- req
			ch := make(chan bridge.Event, 1)
			ch <- bridge.Event{Type: "result", Content: "done"}
			close(ch)
			return ch, nil
		}

		wsURL, cleanup := startDevTerminalWSServer(t, s)
		defer cleanup()

		conn := dialDevTerminalWS(t, wsURL, token)
		defer conn.Close()

		writeWSJSON(t, conn, map[string]string{"type": "prompt", "message": "add tests"})

		select {
		case req := <-bridgeExecCalled:
			if req.Options.PromptInjection != "ws-pipeline-context" {
				t.Fatalf("expected PromptInjection from pipeline, got %q", req.Options.PromptInjection)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("timed out waiting for bridge execute")
		}
	})

	t.Run("pipeline error sends error frame and skips execute", func(t *testing.T) {
		s, token, userUUID := newAuthenticatedDevServer(t)
		boom := errors.New("ws pipeline failed")
		bridgeExecCalled := false

		s.devPipeline = NewDevPipeline()
		s.devPipeline.Use(func(ctx context.Context, req *DevRequest) (*DevRequest, error) {
			return nil, boom
		})

		s.devBridgeExecs[userUUID] = func(ctx context.Context, req bridge.Request) (<-chan bridge.Event, error) {
			bridgeExecCalled = true
			ch := make(chan bridge.Event)
			close(ch)
			return ch, nil
		}

		wsURL, cleanup := startDevTerminalWSServer(t, s)
		defer cleanup()

		conn := dialDevTerminalWS(t, wsURL, token)
		defer conn.Close()

		writeWSJSON(t, conn, map[string]string{"type": "prompt", "message": "add tests"})

		var msg map[string]interface{}
		readWSJSON(t, conn, &msg)

		if bridgeExecCalled {
			t.Fatalf("expected bridge execute to be skipped on pipeline error")
		}
		if got := msg["type"]; got != "error" {
			t.Fatalf("expected error frame, got %#v", msg)
		}
		if !strings.Contains(msg["message"].(string), boom.Error()) {
			t.Fatalf("expected message to contain %q, got %#v", boom.Error(), msg)
		}
	})
}

func startDevTerminalWSServer(t *testing.T, s *Server) (string, func()) {
	t.Helper()
	server := httptest.NewServer(s.authMiddleware(http.HandlerFunc(s.handleDevTerminalWS)))
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/dev/terminal"
	return wsURL, server.Close
}

func dialDevTerminalWS(t *testing.T, wsURL, token string) *websocket.Conn {
	t.Helper()
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("Dial websocket: %v", err)
	}
	return conn
}

func writeWSJSON(t *testing.T, conn *websocket.Conn, payload any) {
	t.Helper()
	if err := conn.WriteJSON(payload); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
}

func readWSJSON(t *testing.T, conn *websocket.Conn, target any) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	if err := conn.ReadJSON(target); err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
}
