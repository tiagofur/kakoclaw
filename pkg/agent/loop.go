// KakoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 KakoClaw contributors

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sipeed/kakoclaw/pkg/bus"
	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/logger"
	"github.com/sipeed/kakoclaw/pkg/mcp"
	"github.com/sipeed/kakoclaw/pkg/observability"
	"github.com/sipeed/kakoclaw/pkg/providers"
	"github.com/sipeed/kakoclaw/pkg/ratelimit"
	"github.com/sipeed/kakoclaw/pkg/session"
	"github.com/sipeed/kakoclaw/pkg/storage"
	"github.com/sipeed/kakoclaw/pkg/tools"
	"github.com/sipeed/kakoclaw/pkg/utils"
)

type AgentLoop struct {
	bus              *bus.MessageBus
	provider         providers.LLMProvider
	workspace        string
	defaultWorkspace string // Base workspace when no user is set
	userUUID         string // User UUID for multiuser support
	userID           int64  // User ID for multiuser support
	model            string
	contextWindow    int // Maximum context window size in tokens
	maxIterations    int
	sessions         *session.SessionManager
	contextBuilder   *ContextBuilder
	tools            *tools.ToolRegistry
	running          atomic.Bool
	summarizing      sync.Map // Tracks which sessions are currently being summarized
	storage          *storage.Storage
}

// ToolRegistry returns the agent loop's tool registry so external
// components (e.g. the workflow engine) can invoke tools directly.
func (al *AgentLoop) ToolRegistry() *tools.ToolRegistry {
	return al.tools
}

// processOptions configures how a message is processed
type processOptions struct {
	SessionKey      string         // Session identifier for history/context
	Channel         string         // Target channel for tool execution
	ChatID          string         // Target chat ID for tool execution
	UserMessage     string         // User message content (may include prefix)
	DefaultResponse string         // Response when LLM returns empty
	EnableSummary   bool           // Whether to trigger summarization
	SendResponse    bool           // Whether to send response via bus
	ModelOverride   string         // If set, use this model instead of the default for LLM calls
	ExcludeTools    []string       // Tool names to exclude from this request (e.g., "web_search")
	OnToken         StreamCallback // Optional callback for text tokens
	OnTool          ToolCallback   // Optional callback for tool call updates
}

// ToolEvent represents a tool call update during agent execution.
type ToolEvent struct {
	Name   string                 `json:"name"`
	Args   map[string]interface{} `json:"arguments"`
	Result string                 `json:"result,omitempty"`
	Status string                 `json:"status"` // "started", "finished", "error"
}

// ToolCallback is called when a tool is about to be executed or starts/finishes.
type ToolCallback func(ev ToolEvent) error

func NewAgentLoop(cfg *config.Config, msgBus *bus.MessageBus, provider providers.LLMProvider) *AgentLoop {
	workspace := cfg.WorkspacePath()
	os.MkdirAll(workspace, 0755)

	restrict := cfg.Agents.Defaults.RestrictToWorkspace

	// Initialize Storage
	var store *storage.Storage
	if cfg.Storage.Path != "" {
		var err error
		store, err = storage.New(cfg.Storage)
		if err != nil {
			logger.ErrorCF("agent", "Failed to initialize storage", map[string]interface{}{"error": err.Error()})
		}
	}

	toolsRegistry := tools.NewToolRegistry()
	toolsRegistry.Register(tools.NewReadFileTool(workspace, restrict))
	toolsRegistry.Register(tools.NewWriteFileTool(workspace, restrict))
	toolsRegistry.Register(tools.NewListDirTool(workspace, restrict))
	toolsRegistry.Register(tools.NewExecTool(workspace, restrict))

	braveAPIKey := cfg.Tools.Web.Search.APIKey
	toolsRegistry.Register(tools.NewWebSearchTool(braveAPIKey, cfg.Tools.Web.Search.MaxResults))
	toolsRegistry.Register(tools.NewWebFetchTool(50000))

	if cfg.Tools.Email.Enabled {
		if strings.TrimSpace(cfg.Tools.Email.Host) == "" || cfg.Tools.Email.Port <= 0 {
			logger.WarnC("agent", "Email tool enabled but SMTP host/port are missing")
		} else {
			toolsRegistry.Register(tools.NewEmailTool(cfg.Tools.Email))
		}
	}

	// Register message tool
	messageTool := tools.NewMessageTool()
	messageTool.SetSendCallback(func(channel, chatID, content string) error {
		msgBus.PublishOutbound(bus.OutboundMessage{
			Channel: channel,
			ChatID:  chatID,
			Content: content,
		})
		return nil
	})
	toolsRegistry.Register(messageTool)

	// Register spawn tool
	subagentManager := tools.NewSubagentManager(provider, workspace, msgBus)
	spawnTool := tools.NewSpawnTool(subagentManager)
	toolsRegistry.Register(spawnTool)

	// Register edit file tool
	editFileTool := tools.NewEditFileTool(workspace, restrict)
	toolsRegistry.Register(editFileTool)
	toolsRegistry.Register(tools.NewAppendFileTool(workspace, restrict))

	// Register task manager tool (shared web tasks DB)
	if store != nil {
		if taskTool, err := tools.NewTaskTool(store); err == nil {
			toolsRegistry.Register(taskTool)
		} else {
			logger.WarnCF("agent", "Task manager tool unavailable", map[string]interface{}{"error": err.Error()})
		}
		// Register knowledge base search tool (RAG)
		toolsRegistry.Register(tools.NewKnowledgeTool(store))
	}

	// Register MCP tools from configured servers
	if len(cfg.Tools.MCP.Servers) > 0 {
		mcpMgr := mcp.NewManager(cfg.Tools.MCP)
		mcpMgr.Start(context.Background())
		for _, mcpTool := range mcpMgr.GetTools() {
			toolsRegistry.Register(mcpTool)
			logger.InfoCF("agent", "Registered MCP tool", map[string]interface{}{"name": mcpTool.Name()})
		}
	}

	sessionsManager := session.NewSessionManager(filepath.Join(workspace, "sessions"))

	// Create context builder and set tools registry
	contextBuilder := NewContextBuilder(workspace)
	contextBuilder.SetToolsRegistry(toolsRegistry)

	return &AgentLoop{
		bus:              msgBus,
		provider:         provider,
		workspace:        workspace,
		defaultWorkspace: workspace,
		userUUID:         "", // Will be set via SetUserForAgent if needed
		userID:           0,  // Default for backward compatibility
		model:            cfg.Agents.Defaults.Model,
		contextWindow:    cfg.Agents.Defaults.MaxTokens, // Restore context window for summarization
		maxIterations:    cfg.Agents.Defaults.MaxToolIterations,
		sessions:         sessionsManager,
		contextBuilder:   contextBuilder,
		tools:            toolsRegistry,
		summarizing:      sync.Map{},
		storage:          store,
	}
}

// SetUserForAgent configures the agent loop for a specific user (multiuser support).
func (al *AgentLoop) SetUserForAgent(userUUID string, userID int64) {
	al.userUUID = userUUID
	al.userID = userID

	if userUUID == "" {
		al.workspace = al.defaultWorkspace
		al.sessions.SetStorage(filepath.Join(al.workspace, "sessions"))
		al.updateToolsWorkspace(al.workspace)
		al.contextBuilder.WithUser(userUUID, userID)
		return
	}

	if userUUID != "" {
		workspace, err := config.EnsureUserWorkspace(userUUID)
		if err != nil {
			logger.WarnCF("agent", "Failed to ensure user workspace", map[string]interface{}{"error": err.Error()})
		} else {
			al.workspace = workspace
			al.sessions.SetStorage(filepath.Join(workspace, "sessions"))
			al.updateToolsWorkspace(workspace)
		}
	}

	al.contextBuilder.WithUser(userUUID, userID)
}

// updateToolsWorkspace updates workspace paths for tools that depend on a workspace directory.
func (al *AgentLoop) updateToolsWorkspace(workspace string) {
	if al.tools == nil {
		return
	}
	al.tools.ForEach(func(t tools.Tool) {
		if wt, ok := t.(tools.WorkspaceTool); ok {
			wt.SetWorkspace(workspace)
		}
	})
}

// applyMessageUserContext configures user context for the current inbound message.
func (al *AgentLoop) applyMessageUserContext(msg bus.InboundMessage) {
	if msg.UserID == 0 {
		al.SetUserForAgent("", 0)
		return
	}
	if al.storage == nil {
		al.SetUserForAgent("", msg.UserID)
		return
	}

	user, err := al.storage.GetUserByID(msg.UserID)
	if err != nil {
		logger.WarnCF("agent", "Failed to resolve user UUID", map[string]interface{}{"error": err.Error()})
		al.SetUserForAgent("", msg.UserID)
		return
	}

	al.SetUserForAgent(user.UUID, msg.UserID)
}

func (al *AgentLoop) Run(ctx context.Context) error {
	al.running.Store(true)

	for al.running.Load() {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, ok := al.bus.ConsumeInbound(ctx)
			if !ok {
				continue
			}

			al.applyMessageUserContext(msg)

			response, err := al.processMessage(ctx, msg)
			if err != nil {
				response = fmt.Sprintf("Error processing message: %v", err)
			}

			if response != "" {
				al.bus.PublishOutbound(bus.OutboundMessage{
					UserID:  msg.UserID,
					Channel: msg.Channel,
					ChatID:  msg.ChatID,
					Content: response,
				})
			}
		}
	}

	return nil
}

func (al *AgentLoop) Stop() {
	al.running.Store(false)
}

func (al *AgentLoop) RegisterTool(tool tools.Tool) {
	al.tools.Register(tool)
}

func (al *AgentLoop) ProcessDirect(ctx context.Context, content, sessionKey string) (string, error) {
	return al.ProcessDirectWithChannel(ctx, content, sessionKey, "cli", "direct")
}

func (al *AgentLoop) ProcessDirectWithChannel(ctx context.Context, content, sessionKey, channel, chatID string) (string, error) {
	msg := bus.InboundMessage{
		Channel:    channel,
		SenderID:   "cron",
		ChatID:     chatID,
		Content:    content,
		SessionKey: sessionKey,
	}

	return al.processMessage(ctx, msg)
}

// ProcessDirectWithUser processes a message on behalf of a specific user.
func (al *AgentLoop) ProcessDirectWithUser(ctx context.Context, userID int64, content, sessionKey string) (string, error) {
	msg := bus.InboundMessage{
		Channel:    "cli",
		SenderID:   "task_worker",
		ChatID:     "direct",
		Content:    content,
		SessionKey: sessionKey,
		UserID:     userID,
	}

	return al.processMessage(ctx, msg)
}

// ProcessDirectWithModel processes a message using a specific model override.
// If modelOverride is empty, uses the default configured model.
// excludeTools optionally specifies tool names to exclude from this request.
func (al *AgentLoop) ProcessDirectWithModel(ctx context.Context, content, sessionKey, modelOverride string, excludeTools ...string) (string, error) {
	msg := bus.InboundMessage{
		Channel:    "cli",
		SenderID:   "cron",
		ChatID:     "direct",
		Content:    content,
		SessionKey: sessionKey,
	}

	return al.processMessageWithModel(ctx, msg, modelOverride, excludeTools...)
}

// StreamCallback is called for each streamed token. Return an error to abort streaming.
type StreamCallback func(token string) error

// ProcessDirectWithModelStream processes a message and streams the final response token-by-token.
// If the provider doesn't support streaming, falls back to sending the full response at once.
// The onToken callback is called for each token; the full accumulated response is still returned.
// excludeTools optionally specifies tool names to exclude from this request.
func (al *AgentLoop) ProcessDirectWithModelStream(ctx context.Context, content, sessionKey, modelOverride string, onToken StreamCallback, onTool ToolCallback, excludeTools ...string) (string, error) {
	msg := bus.InboundMessage{
		Channel:    "cli",
		SenderID:   "cron",
		ChatID:     "direct",
		Content:    content,
		SessionKey: sessionKey,
	}

	return al.processMessageWithModelStream(ctx, msg, modelOverride, onToken, onTool, excludeTools...)
}

// SupportsStreaming returns true if the current provider supports streaming.
func (al *AgentLoop) SupportsStreaming() bool {
	_, ok := al.provider.(providers.StreamingLLMProvider)
	return ok
}

func (al *AgentLoop) processMessage(ctx context.Context, msg bus.InboundMessage) (string, error) {
	return al.processMessageWithModel(ctx, msg, "")
}

func (al *AgentLoop) processMessageWithModel(ctx context.Context, msg bus.InboundMessage, modelOverride string, excludeTools ...string) (string, error) {
	al.applyMessageUserContext(msg)

	// Issue #9: Rate limiting
	// Check user rate limit
	userKey := fmt.Sprintf("user:%s", msg.SenderID)
	if msg.UserID > 0 {
		userKey = fmt.Sprintf("user:%d", msg.UserID)
	}
	if !ratelimit.GetGlobalLimiter().Allow(userKey) {
		logger.WarnCF("agent", "Rate limit exceeded for user", map[string]interface{}{
			"sender_id": msg.SenderID,
		})
		return "Rate limit exceeded. Please wait a moment before sending more messages.", nil
	}

	// Add message preview to log
	preview := utils.Truncate(msg.Content, 80)
	logger.InfoCF("agent", fmt.Sprintf("Processing message from %s:%s: %s", msg.Channel, msg.SenderID, preview),
		map[string]interface{}{
			"channel":     msg.Channel,
			"chat_id":     msg.ChatID,
			"sender_id":   msg.SenderID,
			"session_key": msg.SessionKey,
		})

	// Route system messages to processSystemMessage
	if msg.Channel == "system" {
		return al.processSystemMessage(ctx, msg)
	}

	// Process as user message
	return al.runAgentLoop(ctx, processOptions{
		SessionKey:      msg.SessionKey,
		Channel:         msg.Channel,
		ChatID:          msg.ChatID,
		UserMessage:     msg.Content,
		DefaultResponse: "I've completed processing but have no response to give.",
		EnableSummary:   true,
		SendResponse:    false,
		ModelOverride:   modelOverride,
		ExcludeTools:    excludeTools,
	})
}

func (al *AgentLoop) processMessageWithModelStream(ctx context.Context, msg bus.InboundMessage, modelOverride string, onToken StreamCallback, onTool ToolCallback, excludeTools ...string) (string, error) {
	// Rate limiting
	userKey := fmt.Sprintf("user:%s", msg.SenderID)
	if !ratelimit.GetGlobalLimiter().Allow(userKey) {
		return "Rate limit exceeded. Please wait a moment before sending more messages.", nil
	}

	preview := utils.Truncate(msg.Content, 80)
	logger.InfoCF("agent", fmt.Sprintf("Processing streaming message from %s:%s: %s", msg.Channel, msg.SenderID, preview),
		map[string]interface{}{
			"channel":     msg.Channel,
			"sender_id":   msg.SenderID,
			"session_key": msg.SessionKey,
		})

	if msg.Channel == "system" {
		return al.processSystemMessage(ctx, msg)
	}

	return al.runAgentLoopStream(ctx, processOptions{
		SessionKey:      msg.SessionKey,
		Channel:         msg.Channel,
		ChatID:          msg.ChatID,
		UserMessage:     msg.Content,
		DefaultResponse: "I've completed processing but have no response to give.",
		EnableSummary:   true,
		SendResponse:    false,
		ModelOverride:   modelOverride,
		ExcludeTools:    excludeTools,
		OnToken:         onToken,
		OnTool:          onTool,
	}, onToken)
}

func (al *AgentLoop) processSystemMessage(ctx context.Context, msg bus.InboundMessage) (string, error) {
	// Verify this is a system message
	if msg.Channel != "system" {
		return "", fmt.Errorf("processSystemMessage called with non-system message channel: %s", msg.Channel)
	}

	logger.InfoCF("agent", "Processing system message",
		map[string]interface{}{
			"sender_id": msg.SenderID,
			"chat_id":   msg.ChatID,
		})

	// Parse origin from chat_id (format: "channel:chat_id")
	var originChannel, originChatID string
	if idx := strings.Index(msg.ChatID, ":"); idx > 0 {
		originChannel = msg.ChatID[:idx]
		originChatID = msg.ChatID[idx+1:]
	} else {
		// Fallback
		originChannel = "cli"
		originChatID = msg.ChatID
	}

	// Use the origin session for context
	sessionKey := fmt.Sprintf("%s:%s", originChannel, originChatID)

	// Process as system message with routing back to origin
	return al.runAgentLoop(ctx, processOptions{
		SessionKey:      sessionKey,
		Channel:         originChannel,
		ChatID:          originChatID,
		UserMessage:     fmt.Sprintf("[System: %s] %s", msg.SenderID, msg.Content),
		DefaultResponse: "Background task completed.",
		EnableSummary:   false,
		SendResponse:    true, // Send response back to original channel
	})
}

// runAgentLoop is the core message processing logic.
// It handles context building, LLM calls, tool execution, and response handling.
func (al *AgentLoop) runAgentLoop(ctx context.Context, opts processOptions) (string, error) {
	agentStart := time.Now()

	// 1. Update tool contexts
	al.updateToolContexts(opts.Channel, opts.ChatID)

	// 2. Build messages
	history := al.sessions.GetHistoryForUser(al.userID, opts.SessionKey)
	summary := al.sessions.GetSummaryForUser(al.userID, opts.SessionKey)
	messages := al.contextBuilder.BuildMessages(
		history,
		summary,
		opts.UserMessage,
		nil,
		opts.Channel,
		opts.ChatID,
	)

	// 3. Save user message to session
	al.sessions.AddMessageForUser(al.userID, opts.SessionKey, "user", opts.UserMessage)
	if al.storage != nil {
		if err := al.storage.SaveMessageForUser(al.userID, opts.SessionKey, "user", opts.UserMessage); err != nil {
			logger.ErrorCF("agent", "Failed to save user message to storage", map[string]interface{}{"error": err.Error()})
		}
	}

	// 4. Run LLM iteration loop
	finalContent, iteration, err := al.runLLMIteration(ctx, messages, opts)
	observability.Global().RecordAgentRun(time.Since(agentStart), iteration, err)
	if err != nil {
		return "", err
	}

	// 5. Handle empty response
	if finalContent == "" {
		finalContent = opts.DefaultResponse
	}

	// 6. Save final assistant message to session
	al.sessions.AddMessageForUser(al.userID, opts.SessionKey, "assistant", finalContent)
	al.sessions.SaveForUser(al.userID, al.sessions.GetOrCreateForUser(al.userID, opts.SessionKey))
	if al.storage != nil {
		if err := al.storage.SaveMessageForUser(al.userID, opts.SessionKey, "assistant", finalContent); err != nil {
			logger.ErrorCF("agent", "Failed to save assistant message to storage", map[string]interface{}{"error": err.Error()})
		}
	}

	// 7. Optional: summarization
	if opts.EnableSummary {
		al.maybeSummarize(opts.SessionKey)
	}

	// 8. Optional: send response via bus
	if opts.SendResponse {
		al.bus.PublishOutbound(bus.OutboundMessage{
			UserID:  al.userID,
			Channel: opts.Channel,
			ChatID:  opts.ChatID,
			Content: finalContent,
		})
	}

	// 9. Log response
	responsePreview := utils.Truncate(finalContent, 120)
	logger.InfoCF("agent", fmt.Sprintf("Response: %s", responsePreview),
		map[string]interface{}{
			"session_key":  opts.SessionKey,
			"iterations":   iteration,
			"final_length": len(finalContent),
		})

	return finalContent, nil
}

// runAgentLoopStream is like runAgentLoop but streams the final text response token-by-token.
// Tool call iterations are handled non-streaming. Only the final text answer is streamed.
func (al *AgentLoop) runAgentLoopStream(ctx context.Context, opts processOptions, onToken StreamCallback) (string, error) {
	agentStart := time.Now()

	// 1. Update tool contexts
	al.updateToolContexts(opts.Channel, opts.ChatID)

	// 2. Build messages
	history := al.sessions.GetHistoryForUser(al.userID, opts.SessionKey)
	summary := al.sessions.GetSummaryForUser(al.userID, opts.SessionKey)
	messages := al.contextBuilder.BuildMessages(
		history,
		summary,
		opts.UserMessage,
		nil,
		opts.Channel,
		opts.ChatID,
	)

	// 3. Save user message to session
	al.sessions.AddMessageForUser(al.userID, opts.SessionKey, "user", opts.UserMessage)
	if al.storage != nil {
		if err := al.storage.SaveMessageForUser(al.userID, opts.SessionKey, "user", opts.UserMessage); err != nil {
			logger.ErrorCF("agent", "Failed to save user message to storage", map[string]interface{}{"error": err.Error()})
		}
	}

	// 4. Run LLM iteration loop with streaming on the final response
	finalContent, iteration, err := al.runLLMIterationStream(ctx, messages, opts, onToken)
	observability.Global().RecordAgentRun(time.Since(agentStart), iteration, err)
	if err != nil {
		return "", err
	}

	// 5. Handle empty response
	if finalContent == "" {
		finalContent = opts.DefaultResponse
	}

	// 6. Save final assistant message to session
	al.sessions.AddMessageForUser(al.userID, opts.SessionKey, "assistant", finalContent)
	al.sessions.SaveForUser(al.userID, al.sessions.GetOrCreateForUser(al.userID, opts.SessionKey))
	if al.storage != nil {
		if err := al.storage.SaveMessageForUser(al.userID, opts.SessionKey, "assistant", finalContent); err != nil {
			logger.ErrorCF("agent", "Failed to save assistant message to storage", map[string]interface{}{"error": err.Error()})
		}
	}

	// 7. Optional: summarization
	if opts.EnableSummary {
		al.maybeSummarize(opts.SessionKey)
	}

	// 8. Optional: send response via bus
	if opts.SendResponse {
		al.bus.PublishOutbound(bus.OutboundMessage{
			UserID:  al.userID,
			Channel: opts.Channel,
			ChatID:  opts.ChatID,
			Content: finalContent,
		})
	}

	// 9. Log response
	responsePreview := utils.Truncate(finalContent, 120)
	logger.InfoCF("agent", fmt.Sprintf("Streaming response: %s", responsePreview),
		map[string]interface{}{
			"session_key":  opts.SessionKey,
			"iterations":   iteration,
			"final_length": len(finalContent),
		})

	return finalContent, nil
}

// runLLMIteration executes the LLM call loop with tool handling.
// Returns the final content, iteration count, and any error.
func (al *AgentLoop) runLLMIteration(ctx context.Context, messages []providers.Message, opts processOptions) (string, int, error) {
	iteration := 0
	var finalContent string

	// Determine which model to use (override or default)
	model := al.model
	if opts.ModelOverride != "" {
		model = opts.ModelOverride
	}

	for iteration < al.maxIterations {
		iteration++

		logger.DebugCF("agent", "LLM iteration",
			map[string]interface{}{
				"iteration": iteration,
				"max":       al.maxIterations,
			})

		// Build tool definitions
		toolDefs := al.tools.GetDefinitions()
		providerToolDefs := make([]providers.ToolDefinition, 0, len(toolDefs))
		for _, td := range toolDefs {
			toolName := td["function"].(map[string]interface{})["name"].(string)
			// Skip excluded tools (e.g., web_search when user toggles it off)
			if len(opts.ExcludeTools) > 0 {
				excluded := false
				for _, ex := range opts.ExcludeTools {
					if ex == toolName {
						excluded = true
						break
					}
				}
				if excluded {
					continue
				}
			}
			providerToolDefs = append(providerToolDefs, providers.ToolDefinition{
				Type: td["type"].(string),
				Function: providers.ToolFunctionDefinition{
					Name:        toolName,
					Description: td["function"].(map[string]interface{})["description"].(string),
					Parameters:  td["function"].(map[string]interface{})["parameters"].(map[string]interface{}),
				},
			})
		}

		// Log LLM request details
		logger.DebugCF("agent", "LLM request",
			map[string]interface{}{
				"iteration":         iteration,
				"model":             model,
				"messages_count":    len(messages),
				"tools_count":       len(providerToolDefs),
				"max_tokens":        8192,
				"temperature":       0.7,
				"system_prompt_len": len(messages[0].Content),
			})

		// Log full messages (detailed)
		logger.DebugCF("agent", "Full LLM request",
			map[string]interface{}{
				"iteration":     iteration,
				"messages_json": formatMessagesForLog(messages),
				"tools_json":    formatToolsForLog(providerToolDefs),
			})

		// Call LLM
		llmStart := time.Now()
		response, err := al.provider.Chat(ctx, messages, providerToolDefs, model, map[string]interface{}{
			"max_tokens":  8192,
			"temperature": 0.7,
		})
		llmDur := time.Since(llmStart)
		tokensIn := al.estimateTokens(messages)
		tokensOut := 0
		if err == nil {
			tokensOut = len(response.Content) / 4 // rough estimate
		}
		observability.Global().RecordLLMCall(model, llmDur, tokensIn, tokensOut, err)

		if err != nil {
			logger.ErrorCF("agent", "LLM call failed",
				map[string]interface{}{
					"iteration": iteration,
					"error":     err.Error(),
				})
			return "", iteration, fmt.Errorf("LLM call failed: %w", err)
		}

		// Check if no tool calls - we're done
		if len(response.ToolCalls) == 0 {
			finalContent = response.Content
			logger.InfoCF("agent", "LLM response without tool calls (direct answer)",
				map[string]interface{}{
					"iteration":     iteration,
					"content_chars": len(finalContent),
				})
			break
		}

		// Log tool calls
		toolNames := make([]string, 0, len(response.ToolCalls))
		for _, tc := range response.ToolCalls {
			toolNames = append(toolNames, tc.Name)
		}
		logger.InfoCF("agent", "LLM requested tool calls",
			map[string]interface{}{
				"tools":     toolNames,
				"count":     len(toolNames),
				"iteration": iteration,
			})

		// Build assistant message with tool calls
		assistantMsg := providers.Message{
			Role:    "assistant",
			Content: response.Content,
		}
		for _, tc := range response.ToolCalls {
			argumentsJSON, _ := json.Marshal(tc.Arguments)
			assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, providers.ToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: &providers.FunctionCall{
					Name:      tc.Name,
					Arguments: string(argumentsJSON),
				},
			})
		}
		messages = append(messages, assistantMsg)

		// Save assistant message with tool calls to session
		al.sessions.AddFullMessageForUser(al.userID, opts.SessionKey, assistantMsg)

		// Execute tool calls
		for _, tc := range response.ToolCalls {
			// Log tool call with arguments preview
			argsJSON, _ := json.Marshal(tc.Arguments)
			argsPreview := utils.Truncate(string(argsJSON), 200)
			logger.InfoCF("agent", fmt.Sprintf("Tool call: %s(%s)", tc.Name, argsPreview),
				map[string]interface{}{
					"tool":      tc.Name,
					"iteration": iteration,
				})

			// Notify start of tool call
			if opts.OnTool != nil {
				_ = opts.OnTool(ToolEvent{Name: tc.Name, Args: tc.Arguments, Status: "started"})
			}

			toolStart := time.Now()
			result, err := al.tools.ExecuteWithContext(ctx, tc.Name, tc.Arguments, opts.Channel, opts.ChatID)
			toolDur := time.Since(toolStart)
			observability.Global().RecordToolCall(tc.Name, toolDur, err)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			// Notify end of tool call
			if opts.OnTool != nil {
				status := "finished"
				if err != nil {
					status = "error"
				}
				_ = opts.OnTool(ToolEvent{Name: tc.Name, Args: tc.Arguments, Result: result, Status: status})
			}

			toolResultMsg := providers.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			}
			messages = append(messages, toolResultMsg)

			// Save tool result message to session
			al.sessions.AddFullMessageForUser(al.userID, opts.SessionKey, toolResultMsg)
		}
	}

	return finalContent, iteration, nil
}

// runLLMIterationStream is like runLLMIteration but streams the final text response.
// Tool call iterations use non-streaming Chat(). Only the last iteration (no tool calls)
// uses ChatStream() if the provider supports it.
func (al *AgentLoop) runLLMIterationStream(ctx context.Context, messages []providers.Message, opts processOptions, onToken StreamCallback) (string, int, error) {
	iteration := 0
	var finalContent string

	model := al.model
	if opts.ModelOverride != "" {
		model = opts.ModelOverride
	}

	streamingProvider, canStream := al.provider.(providers.StreamingLLMProvider)

	for iteration < al.maxIterations {
		iteration++

		logger.DebugCF("agent", "LLM streaming iteration",
			map[string]interface{}{
				"iteration":  iteration,
				"max":        al.maxIterations,
				"can_stream": canStream,
			})

		// Build tool definitions
		toolDefs := al.tools.GetDefinitions()
		providerToolDefs := make([]providers.ToolDefinition, 0, len(toolDefs))
		for _, td := range toolDefs {
			toolName := td["function"].(map[string]interface{})["name"].(string)
			// Skip excluded tools (e.g., web_search when user toggles it off)
			if len(opts.ExcludeTools) > 0 {
				excluded := false
				for _, ex := range opts.ExcludeTools {
					if ex == toolName {
						excluded = true
						break
					}
				}
				if excluded {
					continue
				}
			}
			providerToolDefs = append(providerToolDefs, providers.ToolDefinition{
				Type: td["type"].(string),
				Function: providers.ToolFunctionDefinition{
					Name:        toolName,
					Description: td["function"].(map[string]interface{})["description"].(string),
					Parameters:  td["function"].(map[string]interface{})["parameters"].(map[string]interface{}),
				},
			})
		}

		llmOpts := map[string]interface{}{
			"max_tokens":  8192,
			"temperature": 0.7,
		}

		// Try streaming for this iteration
		if canStream {
			llmStart := time.Now()
			ch, err := streamingProvider.ChatStream(ctx, messages, providerToolDefs, model, llmOpts)
			if err != nil {
				observability.Global().RecordLLMCall(model, time.Since(llmStart), al.estimateTokens(messages), 0, err)
				logger.ErrorCF("agent", "Streaming LLM call failed",
					map[string]interface{}{
						"iteration": iteration,
						"error":     err.Error(),
					})
				return "", iteration, fmt.Errorf("streaming LLM call failed: %w", err)
			}

			// Accumulate the response from the stream
			var contentBuilder strings.Builder
			var toolCalls []providers.ToolCall
			var finishReason string
			// Map to accumulate fragmented tool call arguments
			toolCallArgs := make(map[int]strings.Builder)
			toolCallMeta := make(map[int]providers.ToolCall)

			for chunk := range ch {
				// Stream text tokens to client
				if chunk.Content != "" {
					contentBuilder.WriteString(chunk.Content)
					if onToken != nil {
						if err := onToken(chunk.Content); err != nil {
							// Client disconnected or error — stop processing
							return contentBuilder.String(), iteration, err
						}
					}
				}

				// Accumulate tool call fragments
				for _, tc := range chunk.ToolCalls {
					idx := 0 // Default index for tool calls
					if tc.ID != "" {
						// New tool call — store metadata
						toolCallMeta[idx] = tc
					}
					if tc.Function != nil && tc.Function.Arguments != "" {
						b := toolCallArgs[idx]
						b.WriteString(tc.Function.Arguments)
						toolCallArgs[idx] = b
					}
				}

				if chunk.FinishReason != "" {
					finishReason = chunk.FinishReason
				}
			}

			// Assemble completed tool calls
			for idx, meta := range toolCallMeta {
				args := make(map[string]interface{})
				if argsBuilder, ok := toolCallArgs[idx]; ok {
					if err := json.Unmarshal([]byte(argsBuilder.String()), &args); err != nil {
						args["raw"] = argsBuilder.String()
					}
				}
				toolCalls = append(toolCalls, providers.ToolCall{
					ID:        meta.ID,
					Name:      meta.Name,
					Type:      meta.Type,
					Arguments: args,
				})
			}

			// Record streaming LLM call metrics
			streamContent := contentBuilder.String()
			streamTokensOut := len(streamContent) / 4
			observability.Global().RecordLLMCall(model, time.Since(llmStart), al.estimateTokens(messages), streamTokensOut, nil)

			// If no tool calls — we're done
			if len(toolCalls) == 0 {
				finalContent = streamContent
				break
			}

			// Tool calls — handle them (non-streaming for tool execution)
			_ = finishReason
			logger.InfoCF("agent", "Streaming iteration got tool calls",
				map[string]interface{}{
					"iteration": iteration,
					"count":     len(toolCalls),
				})

			// Build assistant message with tool calls
			assistantMsg := providers.Message{
				Role:    "assistant",
				Content: contentBuilder.String(),
			}
			for _, tc := range toolCalls {
				argumentsJSON, _ := json.Marshal(tc.Arguments)
				assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, providers.ToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: &providers.FunctionCall{
						Name:      tc.Name,
						Arguments: string(argumentsJSON),
					},
				})
			}
			messages = append(messages, assistantMsg)
			al.sessions.AddFullMessageForUser(al.userID, opts.SessionKey, assistantMsg)

			// Execute tool calls
			for _, tc := range toolCalls {
				argsJSON, _ := json.Marshal(tc.Arguments)
				argsPreview := utils.Truncate(string(argsJSON), 200)
				logger.InfoCF("agent", fmt.Sprintf("Tool call: %s(%s)", tc.Name, argsPreview),
					map[string]interface{}{
						"tool":      tc.Name,
						"iteration": iteration,
					})

				// Notify start of tool call
				if opts.OnTool != nil {
					_ = opts.OnTool(ToolEvent{Name: tc.Name, Args: tc.Arguments, Status: "started"})
				}

				toolStart := time.Now()
				result, err := al.tools.ExecuteWithContext(ctx, tc.Name, tc.Arguments, opts.Channel, opts.ChatID)
				toolDur := time.Since(toolStart)
				observability.Global().RecordToolCall(tc.Name, toolDur, err)
				if err != nil {
					result = fmt.Sprintf("Error: %v", err)
				}

				// Notify end of tool call
				if opts.OnTool != nil {
					status := "finished"
					if err != nil {
						status = "error"
					}
					_ = opts.OnTool(ToolEvent{Name: tc.Name, Args: tc.Arguments, Result: result, Status: status})
				}

				toolResultMsg := providers.Message{
					Role:       "tool",
					Content:    result,
					ToolCallID: tc.ID,
				}
				messages = append(messages, toolResultMsg)
				al.sessions.AddFullMessageForUser(al.userID, opts.SessionKey, toolResultMsg)
			}

			continue // Next iteration
		}

		// Fallback: non-streaming Chat()
		fallbackStart := time.Now()
		response, err := al.provider.Chat(ctx, messages, providerToolDefs, model, llmOpts)
		fallbackDur := time.Since(fallbackStart)
		fallbackTokensIn := al.estimateTokens(messages)
		fallbackTokensOut := 0
		if err == nil {
			fallbackTokensOut = len(response.Content) / 4
		}
		observability.Global().RecordLLMCall(model, fallbackDur, fallbackTokensIn, fallbackTokensOut, err)
		if err != nil {
			return "", iteration, fmt.Errorf("LLM call failed: %w", err)
		}

		if len(response.ToolCalls) == 0 {
			finalContent = response.Content
			// Send the full content as a single token for non-streaming providers
			if onToken != nil && finalContent != "" {
				_ = onToken(finalContent)
			}
			break
		}

		// Tool calls — same handling as runLLMIteration
		assistantMsg := providers.Message{
			Role:    "assistant",
			Content: response.Content,
		}
		for _, tc := range response.ToolCalls {
			argumentsJSON, _ := json.Marshal(tc.Arguments)
			assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, providers.ToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: &providers.FunctionCall{
					Name:      tc.Name,
					Arguments: string(argumentsJSON),
				},
			})
		}
		messages = append(messages, assistantMsg)
		al.sessions.AddFullMessageForUser(al.userID, opts.SessionKey, assistantMsg)

		for _, tc := range response.ToolCalls {
			argsJSON, _ := json.Marshal(tc.Arguments)
			argsPreview := utils.Truncate(string(argsJSON), 200)
			logger.InfoCF("agent", fmt.Sprintf("Tool call: %s(%s)", tc.Name, argsPreview),
				map[string]interface{}{
					"tool":      tc.Name,
					"iteration": iteration,
				})

			toolStart := time.Now()
			result, err := al.tools.ExecuteWithContext(ctx, tc.Name, tc.Arguments, opts.Channel, opts.ChatID)
			toolDur := time.Since(toolStart)
			observability.Global().RecordToolCall(tc.Name, toolDur, err)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			toolResultMsg := providers.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			}
			messages = append(messages, toolResultMsg)
			al.sessions.AddFullMessageForUser(al.userID, opts.SessionKey, toolResultMsg)
		}
	}

	return finalContent, iteration, nil
}

// updateToolContexts updates the context for tools that need channel/chatID info.
func (al *AgentLoop) updateToolContexts(channel, chatID string) {
	if tool, ok := al.tools.Get("message"); ok {
		if mt, ok := tool.(*tools.MessageTool); ok {
			mt.SetContext(channel, chatID)
		}
	}
	if tool, ok := al.tools.Get("spawn"); ok {
		if st, ok := tool.(*tools.SpawnTool); ok {
			st.SetContext(channel, chatID)
		}
	}
}

// maybeSummarize triggers summarization if the session history exceeds thresholds.
func (al *AgentLoop) maybeSummarize(sessionKey string) {
	newHistory := al.sessions.GetHistoryForUser(al.userID, sessionKey)
	tokenEstimate := al.estimateTokens(newHistory)
	threshold := al.contextWindow * 75 / 100

	if len(newHistory) > 20 || tokenEstimate > threshold {
		if _, loading := al.summarizing.LoadOrStore(sessionKey, true); !loading {
			go func() {
				defer al.summarizing.Delete(sessionKey)
				al.summarizeSession(sessionKey)
			}()
		}
	}
}

// GetStartupInfo returns information about loaded tools and skills for logging.
func (al *AgentLoop) GetStartupInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// Tools info
	tools := al.tools.List()
	info["tools"] = map[string]interface{}{
		"count": len(tools),
		"names": tools,
	}

	// Skills info
	info["skills"] = al.contextBuilder.GetSkillsInfo()

	return info
}

// formatMessagesForLog formats messages for logging
func formatMessagesForLog(messages []providers.Message) string {
	if len(messages) == 0 {
		return "[]"
	}

	var result string
	result += "[\n"
	for i, msg := range messages {
		result += fmt.Sprintf("  [%d] Role: %s\n", i, msg.Role)
		if msg.ToolCalls != nil && len(msg.ToolCalls) > 0 {
			result += "  ToolCalls:\n"
			for _, tc := range msg.ToolCalls {
				result += fmt.Sprintf("    - ID: %s, Type: %s, Name: %s\n", tc.ID, tc.Type, tc.Name)
				if tc.Function != nil {
					result += fmt.Sprintf("      Arguments: %s\n", utils.Truncate(tc.Function.Arguments, 200))
				}
			}
		}
		if msg.Content != "" {
			content := utils.Truncate(msg.Content, 200)
			result += fmt.Sprintf("  Content: %s\n", content)
		}
		if msg.ToolCallID != "" {
			result += fmt.Sprintf("  ToolCallID: %s\n", msg.ToolCallID)
		}
		result += "\n"
	}
	result += "]"
	return result
}

// formatToolsForLog formats tool definitions for logging
func formatToolsForLog(tools []providers.ToolDefinition) string {
	if len(tools) == 0 {
		return "[]"
	}

	var result string
	result += "[\n"
	for i, tool := range tools {
		result += fmt.Sprintf("  [%d] Type: %s, Name: %s\n", i, tool.Type, tool.Function.Name)
		result += fmt.Sprintf("      Description: %s\n", tool.Function.Description)
		if len(tool.Function.Parameters) > 0 {
			result += fmt.Sprintf("      Parameters: %s\n", utils.Truncate(fmt.Sprintf("%v", tool.Function.Parameters), 200))
		}
	}
	result += "]"
	return result
}

// summarizeSession summarizes the conversation history for a session.
func (al *AgentLoop) summarizeSession(sessionKey string) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	history := al.sessions.GetHistoryForUser(al.userID, sessionKey)
	summary := al.sessions.GetSummaryForUser(al.userID, sessionKey)

	// Keep last 4 messages for continuity
	if len(history) <= 4 {
		return
	}

	toSummarize := history[:len(history)-4]

	// Oversized Message Guard
	// Skip messages larger than 50% of context window to prevent summarizer overflow
	maxMessageTokens := al.contextWindow / 2
	validMessages := make([]providers.Message, 0)
	omitted := false

	for _, m := range toSummarize {
		if m.Role != "user" && m.Role != "assistant" {
			continue
		}
		// Estimate tokens for this message
		msgTokens := len(m.Content) / 4
		if msgTokens > maxMessageTokens {
			omitted = true
			continue
		}
		validMessages = append(validMessages, m)
	}

	if len(validMessages) == 0 {
		return
	}

	// Multi-Part Summarization
	// Split into two parts if history is significant
	var finalSummary string
	if len(validMessages) > 10 {
		mid := len(validMessages) / 2
		part1 := validMessages[:mid]
		part2 := validMessages[mid:]

		s1, _ := al.summarizeBatch(ctx, part1, "")
		s2, _ := al.summarizeBatch(ctx, part2, "")

		// Merge them
		mergePrompt := fmt.Sprintf("Merge these two conversation summaries into one cohesive summary:\n\n1: %s\n\n2: %s", s1, s2)
		resp, err := al.provider.Chat(ctx, []providers.Message{{Role: "user", Content: mergePrompt}}, nil, al.model, map[string]interface{}{
			"max_tokens":  1024,
			"temperature": 0.3,
		})
		if err == nil {
			finalSummary = resp.Content
		} else {
			finalSummary = s1 + " " + s2
		}
	} else {
		finalSummary, _ = al.summarizeBatch(ctx, validMessages, summary)
	}

	if omitted && finalSummary != "" {
		finalSummary += "\n[Note: Some oversized messages were omitted from this summary for efficiency.]"
	}

	if finalSummary != "" {
		al.sessions.SetSummaryForUser(al.userID, sessionKey, finalSummary)
		al.sessions.TruncateHistoryForUser(al.userID, sessionKey, 4)
		al.sessions.SaveForUser(al.userID, al.sessions.GetOrCreateForUser(al.userID, sessionKey))
	}
}

// summarizeBatch summarizes a batch of messages.
func (al *AgentLoop) summarizeBatch(ctx context.Context, batch []providers.Message, existingSummary string) (string, error) {
	prompt := "Provide a concise summary of this conversation segment, preserving core context and key points.\n"
	if existingSummary != "" {
		prompt += "Existing context: " + existingSummary + "\n"
	}
	prompt += "\nCONVERSATION:\n"
	for _, m := range batch {
		prompt += fmt.Sprintf("%s: %s\n", m.Role, m.Content)
	}

	response, err := al.provider.Chat(ctx, []providers.Message{{Role: "user", Content: prompt}}, nil, al.model, map[string]interface{}{
		"max_tokens":  1024,
		"temperature": 0.3,
	})
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// estimateTokens estimates the number of tokens in a message list.
func (al *AgentLoop) estimateTokens(messages []providers.Message) int {
	total := 0
	for _, m := range messages {
		total += len(m.Content) / 4 // Simple heuristic: 4 chars per token
	}
	return total
}
