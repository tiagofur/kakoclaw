package bridge

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
)

const eventChannelBuffer = 16

const maxStderrBytes = 4096

// Bridge manages a long-lived CLI process and communicates via stdin/stdout using NDJSON.
type Bridge struct {
	cfg BridgeConfig

	mu      sync.Mutex
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	scanner *bufio.Scanner

	pending   map[string]chan Event
	pendingMu sync.Mutex

	state   string
	stateMu sync.RWMutex

	stderrBuf bytes.Buffer

	reqCounter atomic.Uint64
	done       chan struct{}
}

// New creates a new Bridge process manager.
func New(cfg BridgeConfig) *Bridge {
	if cfg.NodePath == "" {
		cfg.NodePath = "node"
	}
	done := make(chan struct{})
	close(done)

	return &Bridge{
		cfg:     cfg,
		pending: make(map[string]chan Event),
		state:   "idle",
		done:    done,
	}
}

// State returns the current bridge state ("idle", "running", "dead", "stopped").
func (b *Bridge) State() string {
	b.stateMu.RLock()
	defer b.stateMu.RUnlock()
	return b.state
}

func (b *Bridge) setState(s string) {
	b.stateMu.Lock()
	defer b.stateMu.Unlock()
	b.state = s
}

// BackendName returns the logical backend name ("claude-code" or "opencode").
func (b *Bridge) BackendName() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.cfg.BackendName
}

// Cwd returns the working directory configured for this bridge.
func (b *Bridge) Cwd() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.cfg.Cwd
}

// SetCwd updates the configuration working directory.
func (b *Bridge) SetCwd(cwd string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cfg.Cwd = cwd
}

// Start launches the bridge process. Safe to call multiple times.
func (b *Bridge) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.startLocked(ctx)
}

// LastStderr returns the last captured stderr output from the bridge process
// (up to maxStderrBytes). Useful for diagnosing process crashes.
func (b *Bridge) LastStderr() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := b.stderrBuf.String()
	if len(s) > maxStderrBytes {
		s = s[len(s)-maxStderrBytes:]
	}
	return s
}

func (b *Bridge) startLocked(ctx context.Context) error {
	if b.State() == "running" {
		return nil
	}

	// Reset stderr buffer for this launch
	b.stderrBuf.Reset()

	// Both backends run via Node.js on their respective bundle scripts.
	// The cfg.Backend field holds the path to the extracted .mjs bundle.
	cmd := exec.CommandContext(ctx, b.cfg.NodePath, b.cfg.Backend)
	cmd.Dir = b.cfg.Cwd

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("bridge: stdin pipe: %w", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("bridge: stdout pipe: %w", err)
	}
	// Capture stderr for error reporting while also printing to server console.
	cmd.Stderr = io.MultiWriter(os.Stderr, &b.stderrBuf)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("bridge: start process: %w", err)
	}

	b.cmd = cmd
	b.stdin = stdinPipe
	b.scanner = bufio.NewScanner(stdoutPipe)
	b.scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	b.setState("running")
	b.done = make(chan struct{})

	go b.readLoop()

	return nil
}

// Stop kills the bridge process gracefully.
func (b *Bridge) Stop() error {
	b.mu.Lock()
	state := b.State()
	if state == "stopped" || state == "idle" {
		b.mu.Unlock()
		return nil
	}
	b.setState("stopped")
	stdin, cmd, done := b.stdin, b.cmd, b.done
	b.mu.Unlock()

	if stdin != nil {
		_ = stdin.Close()
	}

	if done != nil {
		<-done // Wait for readloop to exit
	}

	if cmd != nil {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	}

	b.mu.Lock()
	b.cmd = nil
	b.stdin = nil
	b.scanner = nil
	b.pendingMu.Lock()
	b.pending = make(map[string]chan Event)
	b.pendingMu.Unlock()
	b.mu.Unlock()

	return nil
}

// ExecuteSync sends a request and blocks until a terminal event is received.
func (b *Bridge) ExecuteSync(ctx context.Context, req Request) (*Event, error) {
	ch, err := b.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	var last *Event
	for ev := range ch {
		ev := ev
		last = &ev
		if ev.IsTerminal() {
			go func() {
				for range ch {
				}
			}()
			return last, nil
		}
	}

	if last != nil {
		return last, nil
	}
	return nil, fmt.Errorf("bridge: process exited without producing any events")
}

// Ping verifies the bridge process is responsive.
func (b *Bridge) Ping(ctx context.Context) error {
	req := Request{Command: "ping"}
	ev, err := b.ExecuteSync(ctx, req)
	if err != nil {
		return fmt.Errorf("bridge: ping failed: %w", err)
	}

	if ev.Type == "error" {
		return fmt.Errorf("bridge: ping returned error: %s", ev.Message)
	}
	if ev.Type != "pong" {
		return fmt.Errorf("bridge: ping got unexpected event: %s", ev.Type)
	}

	return nil
}

// Execute sends a request to the Bridge process and returns a channel of events.
func (b *Bridge) Execute(ctx context.Context, req Request) (<-chan Event, error) {
	b.mu.Lock()
	if b.State() != "running" {
		if err := b.startLocked(context.Background()); err != nil {
			b.mu.Unlock()
			return nil, err
		}
	}
	b.mu.Unlock()

	if req.RequestID == "" {
		req.RequestID = fmt.Sprintf("req-%d", b.reqCounter.Add(1))
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("bridge: marshal request: %w", err)
	}

	ch := make(chan Event, eventChannelBuffer)
	b.pendingMu.Lock()
	b.pending[req.RequestID] = ch
	b.pendingMu.Unlock()

	b.mu.Lock()
	if b.State() != "running" {
		b.mu.Unlock()
		b.pendingMu.Lock()
		delete(b.pending, req.RequestID)
		b.pendingMu.Unlock()
		safeClose(ch)
		return nil, fmt.Errorf("bridge: process died before write")
	}

	_, err = b.stdin.Write(append(payload, '\n'))
	b.mu.Unlock()

	if err != nil {
		b.pendingMu.Lock()
		delete(b.pending, req.RequestID)
		b.pendingMu.Unlock()
		safeClose(ch)
		return nil, fmt.Errorf("bridge: write request: %w", err)
	}

	out := make(chan Event, eventChannelBuffer)
	go func() {
		defer close(out)
		for {
			select {
			case ev, ok := <-ch:
				if !ok {
					return
				}
				select {
				case out <- ev:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}
