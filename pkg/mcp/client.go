package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sipeed/picoclaw/pkg/logger"
)

// JSON-RPC 2.0 types for MCP protocol

type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP protocol types

type MCPToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type MCPToolListResult struct {
	Tools []MCPToolInfo `json:"tools"`
}

type MCPCallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type MCPToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type MCPCallToolResult struct {
	Content []MCPToolContent `json:"content"`
	IsError bool             `json:"isError,omitempty"`
}

type MCPInitializeParams struct {
	ProtocolVersion string            `json:"protocolVersion"`
	Capabilities    MCPCapabilities   `json:"capabilities"`
	ClientInfo      MCPImplementation `json:"clientInfo"`
}

type MCPCapabilities struct{}

type MCPImplementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type MCPInitializeResult struct {
	ProtocolVersion string            `json:"protocolVersion"`
	Capabilities    json.RawMessage   `json:"capabilities"`
	ServerInfo      MCPImplementation `json:"serverInfo"`
}

// Client represents a connection to a single MCP server via STDIO
type Client struct {
	name    string
	command string
	args    []string
	env     []string

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader

	nextID  atomic.Int64
	pending map[int64]chan *jsonRPCResponse
	mu      sync.Mutex
	writeMu sync.Mutex

	serverInfo MCPInitializeResult
	tools      []MCPToolInfo
	connected  bool
	lastError  string
}

// NewClient creates a new MCP client for the given server config
func NewClient(name, command string, args, env []string) *Client {
	return &Client{
		name:    name,
		command: command,
		args:    args,
		env:     env,
		pending: make(map[int64]chan *jsonRPCResponse),
	}
}

// Name returns the server name
func (c *Client) Name() string { return c.name }

// Connected returns whether the client is connected
func (c *Client) Connected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// LastError returns the last error message
func (c *Client) LastError() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastError
}

// ServerInfo returns the MCP server info from initialization
func (c *Client) ServerInfo() MCPInitializeResult {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.serverInfo
}

// Tools returns the discovered tools
func (c *Client) Tools() []MCPToolInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tools
}

// Connect starts the MCP server process, initializes the protocol, and discovers tools
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	c.connected = false
	c.lastError = ""
	c.mu.Unlock()

	// Start the process
	c.cmd = exec.CommandContext(ctx, c.command, c.args...)
	if len(c.env) > 0 {
		c.cmd.Env = append(c.cmd.Environ(), c.env...)
	}

	var err error
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		c.setError(fmt.Sprintf("failed to create stdin pipe: %v", err))
		return err
	}

	stdoutPipe, err := c.cmd.StdoutPipe()
	if err != nil {
		c.setError(fmt.Sprintf("failed to create stdout pipe: %v", err))
		return err
	}
	c.stdout = bufio.NewReader(stdoutPipe)

	if err := c.cmd.Start(); err != nil {
		c.setError(fmt.Sprintf("failed to start process: %v", err))
		return err
	}

	// Start reading responses in background
	go c.readLoop()

	// Initialize the MCP protocol
	initCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	initResult, err := c.initialize(initCtx)
	if err != nil {
		c.setError(fmt.Sprintf("initialization failed: %v", err))
		c.Close()
		return err
	}

	c.mu.Lock()
	c.serverInfo = *initResult
	c.mu.Unlock()

	// Send initialized notification
	c.sendNotification("notifications/initialized", nil)

	// Discover tools
	toolsResult, err := c.listTools(initCtx)
	if err != nil {
		c.setError(fmt.Sprintf("tools/list failed: %v", err))
		c.Close()
		return err
	}

	c.mu.Lock()
	c.tools = toolsResult.Tools
	c.connected = true
	c.mu.Unlock()

	logger.InfoCF("mcp", "Connected to MCP server", map[string]interface{}{
		"name":    c.name,
		"server":  initResult.ServerInfo.Name,
		"version": initResult.ServerInfo.Version,
		"tools":   len(toolsResult.Tools),
	})

	return nil
}

// CallTool invokes a tool on the MCP server
func (c *Client) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*MCPCallToolResult, error) {
	if !c.Connected() {
		return nil, fmt.Errorf("MCP server %q is not connected", c.name)
	}

	resp, err := c.sendRequest(ctx, "tools/call", MCPCallToolParams{
		Name:      toolName,
		Arguments: arguments,
	})
	if err != nil {
		return nil, err
	}

	var result MCPCallToolResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to decode tools/call result: %w", err)
	}
	return &result, nil
}

// Close shuts down the MCP server process
func (c *Client) Close() {
	c.mu.Lock()
	c.connected = false
	c.mu.Unlock()

	if c.stdin != nil {
		_ = c.stdin.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_ = c.cmd.Wait()
	}

	// Drain all pending requests
	c.mu.Lock()
	for id, ch := range c.pending {
		close(ch)
		delete(c.pending, id)
	}
	c.mu.Unlock()
}

// --- internal methods ---

func (c *Client) setError(msg string) {
	c.mu.Lock()
	c.lastError = msg
	c.mu.Unlock()
	logger.ErrorCF("mcp", "MCP client error", map[string]interface{}{
		"server": c.name,
		"error":  msg,
	})
}

func (c *Client) initialize(ctx context.Context) (*MCPInitializeResult, error) {
	resp, err := c.sendRequest(ctx, "initialize", MCPInitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    MCPCapabilities{},
		ClientInfo: MCPImplementation{
			Name:    "picoclaw",
			Version: "1.0.0",
		},
	})
	if err != nil {
		return nil, err
	}

	var result MCPInitializeResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to decode initialize result: %w", err)
	}
	return &result, nil
}

func (c *Client) listTools(ctx context.Context) (*MCPToolListResult, error) {
	resp, err := c.sendRequest(ctx, "tools/list", nil)
	if err != nil {
		return nil, err
	}

	var result MCPToolListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to decode tools/list result: %w", err)
	}
	return &result, nil
}

func (c *Client) sendRequest(ctx context.Context, method string, params interface{}) (*jsonRPCResponse, error) {
	id := c.nextID.Add(1)

	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// Create response channel
	respCh := make(chan *jsonRPCResponse, 1)
	c.mu.Lock()
	c.pending[id] = respCh
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	// Encode and send
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	c.writeMu.Lock()
	_, err = fmt.Fprintf(c.stdin, "%s\n", data)
	c.writeMu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Wait for response
	select {
	case resp, ok := <-respCh:
		if !ok {
			return nil, fmt.Errorf("connection closed while waiting for response")
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("MCP error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Client) sendNotification(method string, params interface{}) {
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      0, // notifications don't have an ID in MCP, but we use 0 as sentinel
		Method:  method,
		Params:  params,
	}
	// For notifications, we omit the id field by using a special struct
	type notification struct {
		JSONRPC string      `json:"jsonrpc"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params,omitempty"`
	}
	n := notification{JSONRPC: req.JSONRPC, Method: req.Method, Params: req.Params}

	data, err := json.Marshal(n)
	if err != nil {
		return
	}
	c.writeMu.Lock()
	_, _ = fmt.Fprintf(c.stdin, "%s\n", data)
	c.writeMu.Unlock()
}

func (c *Client) readLoop() {
	for {
		line, err := c.stdout.ReadBytes('\n')
		if err != nil {
			if c.Connected() {
				c.setError(fmt.Sprintf("read error: %v", err))
				c.mu.Lock()
				c.connected = false
				c.mu.Unlock()
			}
			return
		}

		var resp jsonRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			// Could be a notification or malformed data; skip
			continue
		}

		// Only route responses with an ID (not notifications from server)
		if resp.ID == nil {
			continue
		}

		c.mu.Lock()
		ch, ok := c.pending[*resp.ID]
		c.mu.Unlock()
		if ok {
			ch <- &resp
		}
	}
}
