package canvas

import (
	"fmt"
	"net/http"
	"sync"
)

// CanvasServer manages in-memory canvas state and broadcasts updates via SSE.
type CanvasServer struct {
	devMode    bool
	content    string
	mu         sync.RWMutex
	clients    map[chan string]struct{}
	clientsMu  sync.Mutex
}

// NewCanvasServer creates a new CanvasServer.
// If devMode is true, the Eval method is enabled.
func NewCanvasServer(devMode bool) *CanvasServer {
	return &CanvasServer{
		devMode: devMode,
		clients: make(map[chan string]struct{}),
	}
}

// Push replaces the canvas content and broadcasts to all SSE clients.
func (s *CanvasServer) Push(html string) {
	s.mu.Lock()
	s.content = html
	s.mu.Unlock()

	s.broadcast(html)
}

// Eval executes JavaScript in the canvas context.
// Only allowed when dev_mode is enabled.
func (s *CanvasServer) Eval(js string) (string, error) {
	if !s.devMode {
		return "", fmt.Errorf("eval is disabled: set canvas.dev_mode = true to enable JavaScript evaluation")
	}
	// Eval is broadcast to connected clients for execution in the sandboxed iframe.
	// The actual JS execution happens client-side; server just relays.
	s.broadcast("eval:" + js)
	return "eval sent", nil
}

// Snapshot returns the current canvas content.
func (s *CanvasServer) Snapshot() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.content
}

// Reset clears the canvas content and notifies clients.
func (s *CanvasServer) Reset() {
	s.mu.Lock()
	s.content = ""
	s.mu.Unlock()

	s.broadcast("")
}

// broadcast sends a message to all connected SSE clients.
func (s *CanvasServer) broadcast(msg string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- msg:
		default:
			// Drop message if client is not keeping up
		}
	}
}

// ServeSSE is an HTTP handler that streams canvas updates via Server-Sent Events.
func (s *CanvasServer) ServeSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan string, 16)
	s.clientsMu.Lock()
	s.clients[ch] = struct{}{}
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, ch)
		s.clientsMu.Unlock()
	}()

	// Send current state on connect
	s.mu.RLock()
	current := s.content
	s.mu.RUnlock()
	if current != "" {
		fmt.Fprintf(w, "data: %s\n\n", current)
		flusher.Flush()
	}

	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
