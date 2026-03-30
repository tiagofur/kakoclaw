// makoclaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 makoclaw contributors

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

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/hooks"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/mcp"
	"github.com/sipeed/makoclaw/pkg/observability"
	"github.com/sipeed/makoclaw/pkg/providers"
	"github.com/sipeed/makoclaw/pkg/ratelimit"
	"github.com/sipeed/makoclaw/pkg/session"
	"github.com/sipeed/makoclaw/pkg/storage"
	"github.com/sipeed/makoclaw/pkg/tools"
	"github.com/sipeed/makoclaw/pkg/utils"
)

type AgentLoop struct {
	bus                *bus.MessageBus
	provider           providers.LLMProvider
	workspace          string
	defaultWorkspace   string // Base workspace when no user is set
	userUUID           string // User UUID for multiuser support
	userID             int64  // User ID for multiuser support
	userRole           string // User role for permission checks
	model              string
	contextWindow      int // Maximum context window size in tokens
	maxIterations      int
	sessions           *session.SessionManager
	contextBuilder     *ContextBuilder
	tools              *tools.ToolRegistry
	baseTools          *tools.ToolRegistry // Unfiltered tools (before permission filtering)
	running            atomic.Bool
	summarizing        sync.Map                  // Tracks which sessions are currently being summarized
	storage            *storage.Storage
	autoCreatedStorage bool                      // True if storage was created by NewAgentLoop (not injected)
	centralStorage     *storage.CentralStorage   // Central DB for user identity and permissions lookups
	mcpManager         *mcp.Manager              // MCP manager for cleanup on shutdown
	auditLogger        *tools.SQLiteAuditLogger  // Audit logger for restricted tools
	cfg                *config.Config            // Config for permission checks
	metrics            *observability.Metrics
	userMu             sync.Mutex        // Mutex for user context writes (SetUserForAgent, applyMessageUserContext)
	involvedAgentsMu   sync.Mutex        // Mutex for thread-safe agent tracking
	involvedAgents     []string          // Agents involved in current/last response
	summarizeWg        sync.WaitGroup    // Tracks active summarization goroutines
	costTracker        *AgentCostTracker // Tracks token usage and estimated costs
	orchestratorOwner  *OrchestratorAgent
	hooks              *hooks.HookRegistry
}

// ToolRegistry returns the agent loop's tool registry so external
// components (e.g. the workflow engine) can invoke tools directly.
func (al *AgentLoop) ToolRegistry() *tools.ToolRegistry {
	return al.tools
}

// CostTracker returns the agent's cost tracker for metrics access.
func (al *AgentLoop) CostTracker() *AgentCostTracker {
	return al.costTracker
}

// AddInvolvedAgent adds an agent/specialist name to the list of agents involved in the current response.
func (al *AgentLoop) AddInvolvedAgent(name string) {
	al.involvedAgentsMu.Lock()
	defer al.involvedAgentsMu.Unlock()
	// Avoid duplicates
	for _, existing := range al.involvedAgents {
		if existing == name {
			return
		}
	}
	al.involvedAgents = append(al.involvedAgents, name)
}

// GetInvolvedAgents returns the list of agents involved in the current/last response.
func (al *AgentLoop) GetInvolvedAgents() []string {
	al.involvedAgentsMu.Lock()
	defer al.involvedAgentsMu.Unlock()
	result := make([]string, len(al.involvedAgents))
	copy(result, al.involvedAgents)
	return result
}

// ClearInvolvedAgents clears the list of involved agents (call before processing a new message).
func (al *AgentLoop) ClearInvolvedAgents() {
	al.involvedAgentsMu.Lock()
	defer al.involvedAgentsMu.Unlock()
	al.involvedAgents = al.involvedAgents[:0]
}

// processOptions configures how a message is processed
type processOptions struct {
	SessionKey       string                 // Session identifier for history/context
	Channel          string                 // Target channel for tool execution
	ChatID           string                 // Target chat ID for tool execution
	UserMessage      string                 // User message content (may include prefix)
	DefaultResponse  string                 // Response when LLM returns empty
	EnableSummary    bool                   // Whether to trigger summarization
	SendResponse     bool                   // Whether to send response via bus
	ModelOverride    string                 // If set, use this model instead of the default for LLM calls
	ExcludeTools     []string               // Tool names to exclude from this request (e.g., "web_search")
	OnToken          StreamCallback         // Optional callback for text tokens
	OnThinking       func(string)           // Optional callback for thinking deltas
	OnTool           ToolCallback           // Optional callback for tool call updates
	OnAgentStatus    AgentStatusCallback    // Optional callback for agent status updates
	OnContentSegment ContentSegmentCallback // Optional callback for content segments
}

// ToolEvent represents a tool call update during agent execution.
type ToolEvent struct {
	Name      string                 `json:"name"`
	Args      map[string]interface{} `json:"arguments"`
	Result    string                 `json:"result,omitempty"`
	Status    string                 `json:"status"`    // "started", "finished", "error"
	AgentName string                 `json:"agent_name,omitempty"` // specialist name, empty for orchestrator
}

// ToolCallback is called when a tool is about to be executed or starts/finishes.
type ToolCallback func(ev ToolEvent) error

// AgentStatusEvent represents agent status changes during execution.
//
// Synthesis lifecycle statuses are emitted around the orchestrator's final
// summary pass after specialist responses have been collected:
//   - synthesis_start: emitted before runLLMIteration() or
//     runLLMIterationStream() begins the final synthesis step for specialist
//     responses.
//   - synthesis_end: emitted after that synthesis step completes and the
//     unified orchestrator summary is ready.
type AgentStatusEvent struct {
	Agent           string    `json:"agent"`
	Status          string    `json:"status"` // "analyzing", "delegating", "working", "complete", "synthesizing", "synthesis_start", "synthesis_end", "timeout"
	SpecialistName  string    `json:"specialist_name,omitempty"`
	Reason          string    `json:"reason,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
	DelegationChain []string  `json:"delegation_chain,omitempty"` // e.g. ["orchestrator", "developer", "security"]
	DelegationDepth int       `json:"delegation_depth,omitempty"` // 0=orchestrator, 1=specialist, 2=colleague
	ParentAgent     string    `json:"parent_agent,omitempty"`     // who delegated to this agent
	ActiveSkills    []string  `json:"active_skills,omitempty"`    // skills the agent is using
	MaxIterations   int       `json:"max_iterations,omitempty"`   // iteration limit for visibility
}

// AgentStatusCallback is called when agent status changes
type AgentStatusCallback func(ev AgentStatusEvent) error

// ContentSegment represents a piece of content attributed to an agent
type ContentSegment struct {
	Agent     string    `json:"agent"`
	Content   string    `json:"content"`
	SegmentID string    `json:"segment_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ContentSegmentCallback is called when content is produced by an agent
type ContentSegmentCallback func(segment ContentSegment) error

// DelegationUpdate represents real-time progress of an active delegation
type DelegationUpdate struct {
	DelegationID  string    `json:"delegation_id"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	Status        string    `json:"status"` // "started", "in_progress", "complete", "error"
	Iteration     int       `json:"iteration"`
	MaxIterations int       `json:"max_iterations"`
	ElapsedMs     int64     `json:"elapsed_ms"`
	Timestamp     time.Time `json:"timestamp"`
}

// DelegationUpdateCallback is called for delegation progress updates
type DelegationUpdateCallback func(update DelegationUpdate) error

// NewAgentLoopForUser creates an agent loop for a specific user with their merged configuration.
// It loads the user's config and merges it with the global config, then initializes the agent loop.
func NewAgentLoopForUser(userUUID string, globalCfg *config.Config, msgBus *bus.MessageBus, centralStore *storage.CentralStorage, userStore ...*storage.Storage) (*AgentLoop, error) {
	if userUUID == "" {
		return nil, fmt.Errorf("userUUID is required")
	}

	// Load user config and merge with global
	userCfg, err := config.LoadConfigForUser(userUUID)
	if err != nil {
		logger.WarnCF("agent", "Failed to load user config, using global", map[string]interface{}{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		userCfg = globalCfg
	}

	// Merge user config over global config
	mergedCfg := config.MergeConfigs(globalCfg, userCfg)

	// Create provider with merged config
	provider, err := providers.CreateProvider(mergedCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider for user %s: %w", userUUID, err)
	}

	// Create agent loop with merged config
	al := NewAgentLoop(mergedCfg, msgBus, provider)
	al.SetCentralStorage(centralStore)
	if len(userStore) > 0 && userStore[0] != nil {
		al.SetStorage(userStore[0])
	}

	// Set user context
	if centralStore != nil {
		user, err := centralStore.GetUserByUUID(userUUID)
		if err == nil {
			al.SetUserForAgent(userUUID, user.ID)
		} else {
			logger.WarnCF("agent", "Failed to get user from storage", map[string]interface{}{
				"user_uuid": userUUID,
				"error":     err.Error(),
			})
		}
	}

	return al, nil
}

// Config returns the agent loop's configuration (which may be merged with user specific config)
func (al *AgentLoop) Config() *config.Config {
	return al.cfg
}

func NewAgentLoop(cfg *config.Config, msgBus *bus.MessageBus, provider providers.LLMProvider) *AgentLoop {
	workspace := cfg.WorkspacePath()
	// Only create the global workspace directory in single-user/CLI mode.
	// In multi-user web mode, each user has their own isolated workspace created
	// by EnsureUserWorkspace/EnsureUserDirectory. The global workspace is not used.
	if cfg.Web.Enabled {
		// Multi-user web mode: suppress global workspace creation.
		_ = workspace
	} else {
		os.MkdirAll(workspace, 0755)
	}

	restrict := cfg.Agents.Defaults.RestrictToWorkspace

	// Initialize Storage
	var store *storage.Storage
	var autoCreatedStorage bool
	if cfg.Storage.Path != "" {
		var err error
		store, err = storage.New(cfg.Storage)
		if err != nil {
			logger.ErrorCF("agent", "Failed to initialize storage", map[string]interface{}{"error": err.Error()})
		} else {
			autoCreatedStorage = true
		}
	}

	toolsRegistry := tools.NewToolRegistry()
	toolsRegistry.Register(tools.NewReadFileTool(workspace, restrict))
	toolsRegistry.Register(tools.NewWriteFileTool(workspace, restrict))
	toolsRegistry.Register(tools.NewListDirTool(workspace, restrict))
	toolsRegistry.Register(tools.NewExecTool(workspace, restrict))
	toolsRegistry.Register(tools.NewPDFTool(workspace, restrict))

	// Search provider selection: SearXNG (self-hosted) > Brave (API key)
	var searchProvider tools.SearchProvider
	if searxngURL := cfg.Tools.Web.Search.SearXNGURL; searxngURL != "" {
		searchProvider = tools.NewSearXNGSearchProvider(searxngURL)
		logger.InfoCF("agent", "Search provider: SearXNG", map[string]interface{}{"url": searxngURL})
	} else if braveKey := cfg.Tools.Web.Search.APIKey; braveKey != "" {
		searchProvider = tools.NewBraveSearchProvider(braveKey)
		logger.InfoC("agent", "Search provider: Brave Search API")
	} else {
		logger.WarnC("agent", "No search provider configured — web_search will return an error")
	}
	toolsRegistry.Register(tools.NewWebSearchTool(searchProvider, cfg.Tools.Web.Search.MaxResults))
	toolsRegistry.Register(tools.NewWebFetchTool(50000))

	if cfg.Tools.Email.Enabled {
		if strings.TrimSpace(cfg.Tools.Email.Host) == "" || cfg.Tools.Email.Port <= 0 {
			logger.WarnCF("agent", "Email tool enabled but SMTP host/port are missing — tool will NOT be available", map[string]interface{}{
				"host": cfg.Tools.Email.Host,
				"port": cfg.Tools.Email.Port,
			})
		} else {
			toolsRegistry.Register(tools.NewEmailTool(cfg.Tools.Email))
			logger.InfoCF("agent", "Email tool registered", map[string]interface{}{
				"host": cfg.Tools.Email.Host,
				"port": cfg.Tools.Email.Port,
				"to":   cfg.Tools.Email.To,
			})
		}
	}

	// Register email reader tool (IMAP inbox access) if email channel is configured
	if cfg.Channels.Email.Enabled && cfg.Channels.Email.IMAPHost != "" && cfg.Channels.Email.Password != "" {
		toolsRegistry.Register(tools.NewReadEmailTool(cfg.Channels.Email))
		logger.InfoCF("agent", "Email reader tool registered", map[string]interface{}{
			"imap_host": cfg.Channels.Email.IMAPHost,
			"mailbox":   cfg.Channels.Email.Mailbox,
		})
	}

	// Register image generation tool
	var imageProvider tools.ImageProvider
	{
		provider := strings.ToLower(cfg.Tools.Image.Provider)
		// Lookup per-provider config, then apply key fallbacks
		var pc config.ImageProviderConfig
		switch provider {
		case "together":
			pc = cfg.Tools.ImageProviders.Together
		case "google", "imagen":
			pc = cfg.Tools.ImageProviders.Google
			if pc.APIKey == "" {
				pc.APIKey = cfg.Providers.Gemini.APIKey
			}
		case "fal", "fal.ai":
			pc = cfg.Tools.ImageProviders.Fal
		case "replicate":
			pc = cfg.Tools.ImageProviders.Replicate
		case "zhipu", "cogview":
			pc = cfg.Tools.ImageProviders.Zhipu
			if pc.APIKey == "" {
				pc.APIKey = cfg.Providers.Zhipu.APIKey
			}
			if pc.APIBase != "" && !strings.HasPrefix(pc.APIBase, "http") {
				logger.WarnCF("agent", "zhipu image api_base is not a valid URL, using default", map[string]any{"stored": pc.APIBase})
				pc.APIBase = ""
			}
			if pc.APIBase == "" {
				pc.APIBase = "https://open.bigmodel.cn/api/paas/v4/images/generations"
			}
		case "bfl":
			pc = cfg.Tools.ImageProviders.BFL
		case "openrouter":
			pc = cfg.Tools.ImageProviders.OpenRouter
			if pc.APIKey == "" {
				pc.APIKey = cfg.Providers.OpenRouter.APIKey
			}
			if pc.APIBase != "" && !strings.HasPrefix(pc.APIBase, "http") {
				logger.WarnCF("agent", "openrouter image api_base is not a valid URL, using default", map[string]any{"stored": pc.APIBase})
				pc.APIBase = ""
			}
			if pc.APIBase == "" {
				pc.APIBase = "https://openrouter.ai/api/v1/images/generations"
			}
		default: // openai / openai-compatible
			pc = cfg.Tools.ImageProviders.OpenAI
			if pc.APIKey == "" {
				pc.APIKey = cfg.Providers.OpenAI.APIKey
			}
			if pc.APIBase != "" && !strings.HasPrefix(pc.APIBase, "http") {
				logger.WarnCF("agent", "openai image api_base is not a valid URL, using default", map[string]any{"stored": pc.APIBase})
				pc.APIBase = ""
			}
			if pc.APIBase == "" {
				pc.APIBase = "https://api.openai.com/v1/images/generations"
			}
		}
		// Legacy global fallback (backward compat with old single-key config)
		if pc.APIKey == "" {
			pc.APIKey = cfg.Tools.Image.APIKey
		}
		if pc.APIBase == "" {
			pc.APIBase = cfg.Tools.Image.APIBase
		}

		imageAPIKey := pc.APIKey
		apiBase := pc.APIBase
		// Image model: agent defaults take priority over tools.image.model
		model := cfg.Agents.Defaults.ImageModel
		if model == "" {
			model = cfg.Tools.Image.Model
		}

		switch provider {
		case "fal", "fal.ai":
			if model == "" {
				model = "fal-ai/flux/dev"
			}
			if imageAPIKey != "" {
				imageProvider = tools.NewFalImageProvider(imageAPIKey, model)
			}
		case "replicate":
			if model == "" {
				model = "black-forest-labs/flux-schnell"
			}
			if imageAPIKey != "" {
				imageProvider = tools.NewReplicateImageProvider(imageAPIKey, model)
			}
		case "together":
			if model == "" {
				model = "black-forest-labs/FLUX.1-schnell-Free"
			}
			if imageAPIKey != "" {
				imageProvider = tools.NewTogetherImageProvider(imageAPIKey, model)
			}
		case "google", "imagen":
			if model == "" {
				model = "gemini-2.0-flash-exp"
			}
			if imageAPIKey != "" {
				imageProvider = tools.NewGoogleImageProvider(imageAPIKey, model)
			}
		case "zhipu", "cogview":
			if model == "" {
				model = "cogview-3-flash"
			}
			if imageAPIKey != "" {
				imageProvider = tools.NewOpenAIImageProvider(imageAPIKey, apiBase, model)
			}
		case "bfl":
			if model == "" {
				model = "flux-pro-1.1"
			}
			if imageAPIKey != "" {
				imageProvider = tools.NewBFLImageProvider(imageAPIKey, model)
			}
		case "openrouter":
			if model == "" {
				model = "openai/dall-e-3"
			}
			if imageAPIKey != "" {
				imageProvider = tools.NewOpenAIImageProvider(imageAPIKey, apiBase, model)
			}
		default: // openai / openai-compatible
			if model == "" {
				model = "gpt-image-1"
			}
			if imageAPIKey != "" {
				imageProvider = tools.NewOpenAIImageProvider(imageAPIKey, apiBase, model)
			}
		}
		if imageProvider != nil {
			logger.InfoCF("agent", "Image generation provider configured", map[string]any{
				"provider": cfg.Tools.Image.Provider,
				"model":    model,
				"url":      apiBase,
			})
		} else {
			logger.WarnC("agent", "No image provider configured — image_generate will return an error")
		}
	}
	toolsRegistry.Register(tools.NewImageGenerateTool(imageProvider, workspace, restrict))
	toolsRegistry.Register(tools.NewShowImageTool(workspace))

	// Register social media tools (native platform providers — zero cost)
	socialProvider := tools.NewMultiPlatformProvider(cfg.Tools.SocialMedia)
	if socialProvider.PlatformCount() > 0 {
		logger.InfoCF("agent", "Social media platforms configured", map[string]any{"platforms": socialProvider.ConfiguredPlatforms()})
	} else {
		logger.WarnC("agent", "No social media platforms configured — social_post will return an error")
	}
	toolsRegistry.Register(tools.NewSocialPostTool(socialProvider))
	toolsRegistry.Register(tools.NewSocialAnalyticsTool(socialProvider))
	toolsRegistry.Register(tools.NewMarketingCampaignTool(workspace))

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
		// Register system setup tool for admin package management
		toolsRegistry.Register(tools.NewSystemSetupTool(store))
	}

	// Register MCP tools from configured servers
	var mcpMgr *mcp.Manager
	if len(cfg.Tools.MCP.Servers) > 0 {
		mcpMgr = mcp.NewManager(cfg.Tools.MCP)
		mcpMgr.Start(context.Background())
		for _, mcpTool := range mcpMgr.GetTools() {
			toolsRegistry.Register(mcpTool)
			logger.InfoCF("agent", "Registered MCP tool", map[string]interface{}{"name": mcpTool.Name()})
		}
	}

	// Register configure tool for runtime config management
	configureTool := tools.NewConfigureTool()
	toolsRegistry.Register(configureTool)

	sessionsManager := session.NewSessionManager(filepath.Join(workspace, "sessions"))

	// Create context builder and set tools registry
	contextBuilder := NewContextBuilder(workspace)
	contextBuilder.SetToolsRegistry(toolsRegistry)

	// Expose user's email to system prompt so the agent knows who it's emailing
	if cfg.Tools.Email.To != "" {
		contextBuilder.WithUserEmail(cfg.Tools.Email.To)
	}

	// Initialize audit logger if storage is available
	var auditLogger *tools.SQLiteAuditLogger
	if store != nil {
		var err error
		auditLogger, err = tools.NewSQLiteAuditLogger(store)
		if err != nil {
			logger.WarnCF("agent", "Failed to initialize audit logger", map[string]interface{}{"error": err.Error()})
		}
	}

	return &AgentLoop{
		bus:                msgBus,
		provider:           provider,
		workspace:          workspace,
		defaultWorkspace:   workspace,
		userUUID:           "",      // Will be set via SetUserForAgent if needed
		userID:             0,       // Default for backward compatibility
		userRole:           "admin", // Default to admin for backward compatibility
		model:              cfg.Agents.Defaults.Model,
		contextWindow:      cfg.Agents.Defaults.MaxTokens, // Restore context window for summarization
		maxIterations:      cfg.Agents.Defaults.MaxToolIterations,
		sessions:           sessionsManager,
		contextBuilder:     contextBuilder,
		tools:              toolsRegistry,
		baseTools:          toolsRegistry, // Keep reference to unfiltered tools
		summarizing:        sync.Map{},
		storage:            store,
		autoCreatedStorage: autoCreatedStorage,
		centralStorage:     nil, // Set via SetCentralStorage in multi-user mode
		mcpManager:         mcpMgr,
		auditLogger:        auditLogger,
		cfg:                cfg,
		metrics:            observability.Global(),
		costTracker:        NewAgentCostTracker(),
		hooks:              hooks.NewHookRegistry(),
	}
}

// SetCentralStorage sets the central database storage for user identity lookups.
func (al *AgentLoop) SetCentralStorage(cs *storage.CentralStorage) {
	al.centralStorage = cs
}

// SetMetrics replaces the active metrics collector for this agent loop.
func (al *AgentLoop) SetMetrics(m *observability.Metrics) {
	if m == nil {
		return
	}
	al.metrics = m
}

// SetUserForAgent configures the agent loop for a specific user (multiuser support).
// Also applies tool permission filtering based on user role.
func (al *AgentLoop) SetUserForAgent(userUUID string, userID int64) {
	al.userMu.Lock()
	defer al.userMu.Unlock()

	al.userUUID = userUUID
	al.userID = userID

	// Get user from central storage to determine role
	if userID > 0 && al.centralStorage != nil {
		user, err := al.centralStorage.GetUserByID(userID)
		if err == nil {
			al.userRole = user.Role

			// Apply permission filtering
			filteredTools := filterToolsByPermissions(al.baseTools, user.Role, userID, al.cfg, al.centralStorage)
			al.tools = filteredTools
			al.contextBuilder.SetToolsRegistry(filteredTools)

			logger.InfoCF("agent", "User configured with role-based tool permissions", map[string]interface{}{
				"user_id":    userID,
				"user_uuid":  userUUID,
				"role":       user.Role,
				"tool_count": len(filteredTools.List()),
			})
		} else {
			logger.WarnCF("agent", "Failed to load user, using default admin permissions", map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			})
			al.userRole = "admin"
		}
	} else {
		al.userRole = "admin" // Default to admin for backwards compatibility
	}

	if userUUID == "" {
		al.workspace = al.defaultWorkspace
		al.sessions.SetStorage(filepath.Join(al.workspace, "sessions"))
		al.updateToolsWorkspace(al.workspace)
		al.updateToolsUser(userID)
		al.updateToolsUserConfig(userID, userUUID)
		al.contextBuilder.WithUser(userUUID, userID)
		return
	}

	workspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		logger.WarnCF("agent", "Failed to ensure user workspace", map[string]interface{}{"error": err.Error()})
	} else {
		al.workspace = workspace
		al.sessions.SetStorage(filepath.Join(workspace, "sessions"))
		al.updateToolsWorkspace(workspace)
	}

	al.contextBuilder.WithUser(userUUID, userID)
	al.updateToolsUser(userID)
	al.updateToolsUserConfig(userID, userUUID)
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

// updateToolsUser updates user ID for tools that need to filter data by user.
func (al *AgentLoop) updateToolsUser(userID int64) {
	if al.tools == nil {
		return
	}
	al.tools.ForEach(func(t tools.Tool) {
		if ut, ok := t.(tools.UserAwareTool); ok {
			ut.SetUserID(userID)
		}
	})
}

// updateToolsUserConfig propagates user context (userID + userUUID) to UserConfigTool implementations.
// This is separate from updateToolsUser because it requires the userUUID for config file access.
func (al *AgentLoop) updateToolsUserConfig(userID int64, userUUID string) {
	if al.tools == nil {
		return
	}
	al.tools.ForEach(func(t tools.Tool) {
		if uct, ok := t.(tools.UserConfigTool); ok {
			uct.SetUserContext(userID, userUUID)
		}
	})
}

// applyMessageUserContext configures user context for the current inbound message.
func (al *AgentLoop) applyMessageUserContext(msg bus.InboundMessage) {
	if msg.UserID == 0 {
		al.SetUserForAgent("", 0)
		return
	}
	// If UUID is already set for this user (e.g., by swarm SetUserForAgent), preserve it
	al.userMu.Lock()
	alreadySet := al.userUUID != "" && al.userID == msg.UserID
	al.userMu.Unlock()
	if alreadySet {
		return
	}
	if al.centralStorage == nil {
		al.SetUserForAgent("", msg.UserID)
		return
	}

	user, err := al.centralStorage.GetUserByID(msg.UserID)
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

			// Wrap processMessage in a closure with panic recovery so that
			// a crash never kills the goroutine silently — the channel's
			// thinking animation would run forever without a response.
			var response string
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.ErrorCF("agent", "Panic in processMessage", map[string]interface{}{
							"panic":       fmt.Sprintf("%v", r),
							"session_key": msg.SessionKey,
							"channel":     msg.Channel,
							"chat_id":     msg.ChatID,
						})
						response = "An internal error occurred. Please try again."
					}
				}()
				var err error
				response, err = al.processMessage(ctx, msg)
				if err != nil {
					response = fmt.Sprintf("Error processing message: %v", err)
				}
			}()

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
	al.summarizeWg.Wait()
	if al.mcpManager != nil {
		al.mcpManager.Stop()
	}
}

func (al *AgentLoop) RegisterTool(tool tools.Tool) {
	if al.baseTools != nil {
		al.baseTools.Register(tool)
	} else if al.tools != nil {
		al.tools.Register(tool)
	}

	if al.baseTools == nil {
		return
	}

	al.tools = filterToolsByPermissions(al.baseTools, al.userRole, al.userID, al.cfg, al.centralStorage)
	if al.contextBuilder != nil {
		al.contextBuilder.SetToolsRegistry(al.tools)
	}
	al.updateToolsWorkspace(al.workspace)
	al.updateToolsUser(al.userID)
	al.updateToolsUserConfig(al.userID, al.userUUID)
}

// SetStorage replaces the active data storage used by stateful tools and message persistence.
// The central storage used for identity/permission lookups remains unchanged.
func (al *AgentLoop) SetStorage(store *storage.Storage) {
	if store == nil {
		return
	}

	// Close the auto-created storage to avoid leaking the DB connection
	if al.storage != nil && al.autoCreatedStorage {
		al.storage.Close()
		al.autoCreatedStorage = false
	}

	al.storage = store

	if taskTool, err := tools.NewTaskTool(store); err == nil {
		al.baseTools.Register(taskTool)
	} else {
		logger.WarnCF("agent", "Task manager tool unavailable after storage switch", map[string]interface{}{"error": err.Error()})
	}
	al.baseTools.Register(tools.NewKnowledgeTool(store))

	al.tools = filterToolsByPermissions(al.baseTools, al.userRole, al.userID, al.cfg, al.centralStorage)
	al.contextBuilder.SetToolsRegistry(al.tools)
	al.updateToolsWorkspace(al.workspace)
	al.updateToolsUser(al.userID)
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

// ProcessDirectWithChannelForUser processes a message for a specific user with a channel override.
func (al *AgentLoop) ProcessDirectWithChannelForUser(ctx context.Context, userID int64, content, sessionKey, channel, chatID string) (string, error) {
	msg := bus.InboundMessage{
		Channel:    channel,
		SenderID:   "cron",
		ChatID:     chatID,
		Content:    content,
		SessionKey: sessionKey,
		UserID:     userID,
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

// ProcessDirectWithUserAndModel processes a message for a specific user using a model override.
// excludeTools optionally specifies tool names to exclude from this request.
func (al *AgentLoop) ProcessDirectWithUserAndModel(ctx context.Context, userID int64, content, sessionKey, modelOverride string, excludeTools ...string) (string, error) {
	msg := bus.InboundMessage{
		Channel:    "cli",
		SenderID:   fmt.Sprintf("user:%d", userID),
		ChatID:     "direct",
		Content:    content,
		SessionKey: sessionKey,
		UserID:     userID,
	}

	return al.processMessageWithModel(ctx, msg, modelOverride, excludeTools...)
}

// StreamCallback is called for each streamed token. Return an error to abort streaming.
type StreamCallback func(token string) error

// ProcessDirectWithModelStream processes a message and streams the final response token-by-token.
// If the provider doesn't support streaming, falls back to sending the full response at once.
// The onToken callback is called for each token; the full accumulated response is still returned.
// excludeTools optionally specifies tool names to exclude from this request.
func (al *AgentLoop) ProcessDirectWithModelStream(ctx context.Context, content, sessionKey, modelOverride string, onToken StreamCallback, onThinking func(string), onTool ToolCallback, excludeTools ...string) (string, error) {
	msg := bus.InboundMessage{
		Channel:    "cli",
		SenderID:   "cron",
		ChatID:     "direct",
		Content:    content,
		SessionKey: sessionKey,
	}

	return al.processMessageWithModelStream(ctx, msg, modelOverride, onToken, onThinking, onTool, excludeTools...)
}

// ProcessDirectWithUserAndModelStream processes a message for a specific user with streaming.
// excludeTools optionally specifies tool names to exclude from this request.
func (al *AgentLoop) ProcessDirectWithUserAndModelStream(ctx context.Context, userID int64, content, sessionKey, modelOverride string, onToken StreamCallback, onThinking func(string), onTool ToolCallback, excludeTools ...string) (string, error) {
	msg := bus.InboundMessage{
		Channel:    "cli",
		SenderID:   fmt.Sprintf("user:%d", userID),
		ChatID:     "direct",
		Content:    content,
		SessionKey: sessionKey,
		UserID:     userID,
	}

	return al.processMessageWithModelStream(ctx, msg, modelOverride, onToken, onThinking, onTool, excludeTools...)
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
	if strings.TrimSpace(modelOverride) != "" {
		ctx = ContextWithModelOverride(ctx, strings.TrimSpace(modelOverride))
	}

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

	// Extract callbacks from context (set by server.go)
	onAgentStatus := agentStatusCallbackFromCtx(ctx)
	onContentSegment := contentSegmentCallbackFromCtx(ctx)

	// Process as user message
	return al.runAgentLoop(ctx, processOptions{
		SessionKey:       msg.SessionKey,
		Channel:          msg.Channel,
		ChatID:           msg.ChatID,
		UserMessage:      msg.Content,
		DefaultResponse:  "I've completed processing but have no response to give.",
		EnableSummary:    true,
		SendResponse:     false,
		ModelOverride:    modelOverride,
		ExcludeTools:     excludeTools,
		OnAgentStatus:    onAgentStatus,
		OnContentSegment: onContentSegment,
	})
}

func (al *AgentLoop) processMessageWithModelStream(ctx context.Context, msg bus.InboundMessage, modelOverride string, onToken StreamCallback, onThinking func(string), onTool ToolCallback, excludeTools ...string) (string, error) {
	if strings.TrimSpace(modelOverride) != "" {
		ctx = ContextWithModelOverride(ctx, strings.TrimSpace(modelOverride))
	}

	al.applyMessageUserContext(msg)

	// Rate limiting
	userKey := fmt.Sprintf("user:%s", msg.SenderID)
	if msg.UserID > 0 {
		userKey = fmt.Sprintf("user:%d", msg.UserID)
	}
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

	// Extract callbacks from context (set by server.go)
	onAgentStatus := agentStatusCallbackFromCtx(ctx)
	onContentSegment := contentSegmentCallbackFromCtx(ctx)

	return al.runAgentLoopStream(ctx, processOptions{
		SessionKey:       msg.SessionKey,
		Channel:          msg.Channel,
		ChatID:           msg.ChatID,
		UserMessage:      msg.Content,
		DefaultResponse:  "I've completed processing but have no response to give.",
		EnableSummary:    true,
		SendResponse:     false,
		ModelOverride:    modelOverride,
		ExcludeTools:     excludeTools,
		OnToken:          onToken,
		OnThinking:       onThinking,
		OnTool:           onTool,
		OnAgentStatus:    onAgentStatus,
		OnContentSegment: onContentSegment,
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
	var teamCtx *TeamContext
	if al.orchestratorOwner != nil {
		teamCtx = al.orchestratorOwner.resetCurrentExecutionState(opts.UserMessage, al.userUUID, al.userID)
		ctx = ContextWithTeamContext(ctx, teamCtx)
	}

	// Clear previous involved agents before processing
	al.ClearInvolvedAgents()
	// Register this agent loop as primary responder
	agentName := al.model
	if agentName == "" {
		agentName = "main"
	}
	al.AddInvolvedAgent(agentName)

	// Emit simple status if orchestrator not available (fallback)
	if opts.OnAgentStatus != nil && (al.cfg == nil || !al.cfg.Agents.Orchestrator.Enabled) {
		_ = opts.OnAgentStatus(AgentStatusEvent{
			Agent:     "agent",
			Status:    "working",
			Timestamp: time.Now(),
		})
	}

	agentStart := time.Now()
	shouldPersist := opts.SessionKey != ""

	// 1. Update tool contexts
	al.updateToolContexts(opts.Channel, opts.ChatID)

	// 2. Build messages
	// Wire analytics for this request's session key (fire-and-forget inside BuildSystemPrompt).
	if al.storage != nil {
		al.contextBuilder.WithAnalytics(al.storage, opts.SessionKey)
		al.contextBuilder.WithCentralStore(al.centralStorage)
	}

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
	if shouldPersist && al.storage != nil {
		if err := al.storage.SaveMessageForUser(al.userID, opts.SessionKey, "user", opts.UserMessage, ""); err != nil {
			logger.ErrorCF("agent", "Failed to save user message to storage", map[string]interface{}{"error": err.Error()})
		}
	}

	// 4. Run LLM iteration loop
	finalContent, iteration, err := al.runLLMIteration(ctx, messages, opts)
	al.metrics.RecordAgentRun(time.Since(agentStart), iteration, err)
	if err != nil {
		return "", err
	}

	// 5. Handle empty response
	if finalContent == "" {
		finalContent = opts.DefaultResponse
	}

	var specialistReports []SpecialistReport
	if al.orchestratorOwner != nil {
		var aggregatedContent string
		specialistReports, aggregatedContent = al.orchestratorOwner.aggregateCurrentExecutionReports(teamCtx)
		if aggregatedContent != "" {
			finalContent = aggregatedContent
		}
	}

	// 6. Save final assistant message to session
	al.sessions.AddMessageForUser(al.userID, opts.SessionKey, "assistant", finalContent)
	if shouldPersist {
		al.sessions.SaveForUser(al.userID, al.sessions.GetOrCreateForUser(al.userID, opts.SessionKey))
	}
	if shouldPersist && al.storage != nil {
		// Prepare metadata with involved agents and tool calls
		agents := al.GetInvolvedAgents()
		if len(agents) == 0 {
			agents = []string{"default"}
		}

		// Collect recent tool calls and tool results
		var trailing []providers.Message
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				// Don't go past the last user message
				break
			}
			if len(messages[i].ToolCalls) > 0 || messages[i].Role == "tool" {
				trailing = append([]providers.Message{messages[i]}, trailing...)
			}
		}

		metadataObj := map[string]interface{}{
			"agents": agents,
		}

		if len(trailing) > 0 {
			metadataObj["messages"] = trailing
		}
		if len(specialistReports) > 0 {
			metadataObj["segments"] = specialistSegmentsFromReports(specialistReports)
		}

		var metadata string
		if metaJSON, err := json.Marshal(metadataObj); err == nil {
			metadata = string(metaJSON)
		} else {
			logger.WarnCF("agent", "Failed to marshal agent metadata", map[string]interface{}{"error": err.Error()})
		}
		if err := al.storage.SaveMessageForUser(al.userID, opts.SessionKey, "assistant", finalContent, metadata); err != nil {
			logger.ErrorCF("agent", "Failed to save assistant message to storage", map[string]interface{}{"error": err.Error()})
		}
	}

	// 7. Optional: summarization
	if opts.EnableSummary {
		al.maybeSummarize(opts.SessionKey)
	}

	// 8. Optional: send response via bus
	if opts.SendResponse {
		// Security hook: message_sending
		if al.hooks != nil {
			hookResult := al.hooks.Run(ctx, "message_sending", hooks.HookContext{
				Message: finalContent,
				UserID:  al.userID,
			})
			if hookResult.Action == hooks.HookBlock {
				logger.InfoCF("agent", "message_sending hook blocked delivery", map[string]interface{}{
					"reason": hookResult.Reason,
				})
			} else {
				al.bus.PublishOutbound(bus.OutboundMessage{
					UserID:  al.userID,
					Channel: opts.Channel,
					ChatID:  opts.ChatID,
					Content: finalContent,
				})
			}
		} else {
			al.bus.PublishOutbound(bus.OutboundMessage{
				UserID:  al.userID,
				Channel: opts.Channel,
				ChatID:  opts.ChatID,
				Content: finalContent,
			})
		}
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
	var teamCtx *TeamContext
	if al.orchestratorOwner != nil {
		teamCtx = al.orchestratorOwner.resetCurrentExecutionState(opts.UserMessage, al.userUUID, al.userID)
		ctx = ContextWithTeamContext(ctx, teamCtx)
	}

	// Clear previous involved agents before processing
	al.ClearInvolvedAgents()
	// Register this agent loop as primary responder
	agentName := al.model
	if agentName == "" {
		agentName = "main"
	}
	al.AddInvolvedAgent(agentName)

	// Emit simple status if orchestrator not available (fallback)
	if opts.OnAgentStatus != nil && (al.cfg == nil || !al.cfg.Agents.Orchestrator.Enabled) {
		_ = opts.OnAgentStatus(AgentStatusEvent{
			Agent:     "agent",
			Status:    "working",
			Timestamp: time.Now(),
		})
	}

	agentStart := time.Now()
	shouldPersist := opts.SessionKey != ""

	// 1. Update tool contexts
	al.updateToolContexts(opts.Channel, opts.ChatID)

	// 2. Build messages
	// Wire analytics for this request's session key (fire-and-forget inside BuildSystemPrompt).
	if al.storage != nil {
		al.contextBuilder.WithAnalytics(al.storage, opts.SessionKey)
		al.contextBuilder.WithCentralStore(al.centralStorage)
	}

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
	if shouldPersist && al.storage != nil {
		if err := al.storage.SaveMessageForUser(al.userID, opts.SessionKey, "user", opts.UserMessage, ""); err != nil {
			logger.ErrorCF("agent", "Failed to save user message to storage", map[string]interface{}{"error": err.Error()})
		}
	}

	// 4. Run LLM iteration loop with streaming on the final response
	finalContent, iteration, err := al.runLLMIterationStream(ctx, messages, opts, onToken)
	al.metrics.RecordAgentRun(time.Since(agentStart), iteration, err)
	if err != nil {
		return "", err
	}

	// 5. Handle empty response
	if finalContent == "" {
		finalContent = opts.DefaultResponse
	}

	var specialistReports []SpecialistReport
	if al.orchestratorOwner != nil {
		var aggregatedContent string
		specialistReports, aggregatedContent = al.orchestratorOwner.aggregateCurrentExecutionReports(teamCtx)
		if aggregatedContent != "" {
			finalContent = aggregatedContent
		}
	}

	// 6. Save final assistant message to session
	al.sessions.AddMessageForUser(al.userID, opts.SessionKey, "assistant", finalContent)
	if shouldPersist {
		al.sessions.SaveForUser(al.userID, al.sessions.GetOrCreateForUser(al.userID, opts.SessionKey))
	}
	if shouldPersist && al.storage != nil {
		// Prepare metadata with involved agents and tool calls
		agents := al.GetInvolvedAgents()
		if len(agents) == 0 {
			agents = []string{"default"}
		}

		// Collect recent tool calls and tool results
		var trailing []providers.Message
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				// Don't go past the last user message
				break
			}
			if len(messages[i].ToolCalls) > 0 || messages[i].Role == "tool" {
				trailing = append([]providers.Message{messages[i]}, trailing...)
			}
		}

		metadataObj := map[string]interface{}{
			"agents": agents,
		}

		if len(trailing) > 0 {
			metadataObj["messages"] = trailing
		}
		if len(specialistReports) > 0 {
			metadataObj["segments"] = specialistSegmentsFromReports(specialistReports)
		}

		var metadata string
		if metaJSON, err := json.Marshal(metadataObj); err == nil {
			metadata = string(metaJSON)
		} else {
			logger.WarnCF("agent", "Failed to marshal agent metadata", map[string]interface{}{"error": err.Error()})
		}
		if err := al.storage.SaveMessageForUser(al.userID, opts.SessionKey, "assistant", finalContent, metadata); err != nil {
			logger.ErrorCF("agent", "Failed to save assistant message to storage", map[string]interface{}{"error": err.Error()})
		}
	}

	// 7. Optional: summarization
	if opts.EnableSummary {
		al.maybeSummarize(opts.SessionKey)
	}

	// 8. Optional: send response via bus
	if opts.SendResponse {
		// Security hook: message_sending
		if al.hooks != nil {
			hookResult := al.hooks.Run(ctx, "message_sending", hooks.HookContext{
				Message: finalContent,
				UserID:  al.userID,
			})
			if hookResult.Action == hooks.HookBlock {
				logger.InfoCF("agent", "message_sending hook blocked delivery", map[string]interface{}{
					"reason": hookResult.Reason,
				})
			} else {
				al.bus.PublishOutbound(bus.OutboundMessage{
					UserID:  al.userID,
					Channel: opts.Channel,
					ChatID:  opts.ChatID,
					Content: finalContent,
				})
			}
		} else {
			al.bus.PublishOutbound(bus.OutboundMessage{
				UserID:  al.userID,
				Channel: opts.Channel,
				ChatID:  opts.ChatID,
				Content: finalContent,
			})
		}
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
	pendingSynthesis := false
	synthesisActive := false

	// Determine which model to use (override or default)
	model := al.model
	if opts.ModelOverride != "" {
		model = opts.ModelOverride
	}

	for iteration < al.maxIterations {
		iteration++

		// synthesis_start is emitted before the orchestrator begins the final
		// synthesis step that consolidates specialist responses into one summary.
		if pendingSynthesis && !synthesisActive {
			emitAgentStatus(ctx, AgentStatusEvent{
				Agent:     "orchestrator",
				Status:    "synthesis_start",
				Timestamp: time.Now(),
			})
			synthesisActive = true
			pendingSynthesis = false
		}

		logger.DebugCF("agent", "LLM iteration",
			map[string]interface{}{
				"iteration": iteration,
				"max":       al.maxIterations,
			})

		// Build tool definitions (exclude all when "*" is in ExcludeTools)
		var providerToolDefs []providers.ToolDefinition
		excludeAll := false
		for _, ex := range opts.ExcludeTools {
			if ex == "*" {
				excludeAll = true
				break
			}
		}
		if !excludeAll {
			toolDefs := al.tools.GetDefinitions()
			providerToolDefs = make([]providers.ToolDefinition, 0, len(toolDefs))
			for _, td := range toolDefs {
				fnMap, ok := td["function"].(map[string]interface{})
				if !ok {
					continue
				}
				toolName, _ := fnMap["name"].(string)
				if toolName == "" {
					continue
				}
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
				tdType, _ := td["type"].(string)
				desc, _ := fnMap["description"].(string)
				params, _ := fnMap["parameters"].(map[string]interface{})
				providerToolDefs = append(providerToolDefs, providers.ToolDefinition{
					Type: tdType,
					Function: providers.ToolFunctionDefinition{
						Name:        toolName,
						Description: desc,
						Parameters:  params,
					},
				})
			}
		}

		// Log LLM request details
		systemPromptLen := 0
		if len(messages) > 0 {
			systemPromptLen = len(messages[0].Content)
		}
		logger.DebugCF("agent", "LLM request",
			map[string]interface{}{
				"iteration":         iteration,
				"model":             model,
				"messages_count":    len(messages),
				"tools_count":       len(providerToolDefs),
				"max_tokens":        8192,
				"temperature":       0.7,
				"system_prompt_len": systemPromptLen,
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
			tokensOut = (len(response.Content) * 10) / 35 // ~3.5 chars/token estimate
			// Use actual usage if available from provider
			if response.Usage != nil {
				tokensIn = response.Usage.PromptTokens
				tokensOut = response.Usage.CompletionTokens
			}
		}
		al.metrics.RecordLLMCall(model, llmDur, tokensIn, tokensOut, err)

		// Record API call for cost tracking
		if err == nil && al.costTracker != nil {
			al.costTracker.RecordAPICall("main", int64(tokensIn), int64(tokensOut), model)
		}

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
			// synthesis_end is emitted after the final synthesis step completes and
			// the unified orchestrator summary is ready to return.
			if synthesisActive {
				emitAgentStatus(ctx, AgentStatusEvent{
					Agent:     "orchestrator",
					Status:    "synthesis_end",
					Timestamp: time.Now(),
				})
				synthesisActive = false
			}
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
			argumentsJSON, err := json.Marshal(tc.Arguments)
			if err != nil {
				logger.WarnCF("agent", "Failed to marshal tool arguments", map[string]interface{}{"tool": tc.Name, "error": err.Error()})
				continue
			}
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

		specialistResponses := 0
		// Execute tool calls
		for _, tc := range response.ToolCalls {
			// Log tool call with arguments preview
			argsJSON, err := json.Marshal(tc.Arguments)
			if err != nil {
				argsJSON = []byte(fmt.Sprintf("%v", tc.Arguments))
			}
			argsPreview := utils.Truncate(string(argsJSON), 200)
			logger.InfoCF("agent", fmt.Sprintf("Tool call: %s(%s)", tc.Name, argsPreview),
				map[string]interface{}{
					"tool":      tc.Name,
					"iteration": iteration,
				})

			// Security hook: before_tool_call
			if al.hooks != nil {
				hookResult := al.hooks.Run(ctx, "before_tool_call", hooks.HookContext{
					ToolName: tc.Name,
					Args:     tc.Arguments,
					UserID:   al.userID,
				})
				if hookResult.Action == hooks.HookBlock {
					messages = append(messages, providers.Message{
						Role:       "tool",
						Content:    fmt.Sprintf("[blocked by security hook: %s]", hookResult.Reason),
						ToolCallID: tc.ID,
					})
					continue
				}
				if hookResult.Action == hooks.HookRequireApproval {
					messages = append(messages, providers.Message{
						Role:       "tool",
						Content:    fmt.Sprintf("[tool requires approval: %s]", hookResult.Reason),
						ToolCallID: tc.ID,
					})
					logger.WarnCF("agent", "require_approval with no owner channel, treating as block", map[string]interface{}{
						"tool": tc.Name,
					})
					continue
				}
			}

			// Notify start of tool call
			agentName := utils.AgentNameFrom(ctx)
			if opts.OnTool != nil {
				_ = opts.OnTool(ToolEvent{Name: tc.Name, Args: tc.Arguments, Status: "started", AgentName: agentName})
			}

			toolStart := time.Now()
			result, err := al.tools.ExecuteWithContext(ctx, tc.Name, tc.Arguments, opts.Channel, opts.ChatID)
			toolDur := time.Since(toolStart)
			al.metrics.RecordToolCall(tc.Name, toolDur, err)

			// Audit restricted tool executions
			if tools.IsRestrictedTool(tc.Name) && al.auditLogger != nil && al.userID > 0 {
				username := fmt.Sprintf("user_%d", al.userID)
				if al.storage != nil {
					if user, userErr := al.storage.GetUserByID(al.userID); userErr == nil {
						username = user.Username
					}
				}

				auditLog := tools.ToolExecutionLog{
					Timestamp: toolStart,
					UserID:    al.userID,
					Username:  username,
					Tool:      tc.Name,
					Arguments: tc.Arguments,
					Success:   err == nil,
					Duration:  toolDur.Milliseconds(),
				}
				if err != nil {
					auditLog.Error = err.Error()
				}

				if auditErr := al.auditLogger.LogToolExecution(ctx, auditLog); auditErr != nil {
					logger.WarnCF("agent", "Failed to log tool execution to audit", map[string]interface{}{
						"tool":  tc.Name,
						"error": auditErr.Error(),
					})
				}
			}

			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}
			if tc.Name == "delegate_to_specialist" && err == nil && strings.TrimSpace(result) != "" {
				specialistResponses++
			}

			// Notify end of tool call
			if opts.OnTool != nil {
				status := "finished"
				if err != nil {
					status = "error"
				}
				_ = opts.OnTool(ToolEvent{Name: tc.Name, Args: tc.Arguments, Result: result, Status: status, AgentName: agentName})
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
		if specialistResponses > 0 {
			pendingSynthesis = true
		}
	}

	// If we exhausted iterations without a text response, force a concluding call
	// with no tools so the LLM must produce a final answer instead of looping.
	if finalContent == "" && iteration >= al.maxIterations {
		logger.WarnCF("agent", "Iteration limit reached without text response, forcing concluding call",
			map[string]interface{}{
				"iterations": iteration,
				"max":        al.maxIterations,
			})

		// Append an instruction to wrap up
		messages = append(messages, providers.Message{
			Role:    "user",
			Content: "[System: You have reached the maximum number of tool iterations. You MUST now provide your final response as text. Summarize what you accomplished and any remaining items.]",
		})

		// Call LLM with NO tools so it must produce text
		concludeResp, err := al.provider.Chat(ctx, messages, nil, model, map[string]interface{}{
			"max_tokens":  4096,
			"temperature": 0.7,
		})
		if err != nil {
			logger.ErrorCF("agent", "Concluding LLM call failed", map[string]interface{}{"error": err.Error()})
			return "", iteration, fmt.Errorf("concluding LLM call failed: %w", err)
		}
		// synthesis_end is emitted after the forced concluding synthesis step
		// completes and the unified orchestrator summary is ready.
		if synthesisActive {
			emitAgentStatus(ctx, AgentStatusEvent{
				Agent:     "orchestrator",
				Status:    "synthesis_end",
				Timestamp: time.Now(),
			})
			synthesisActive = false
		}
		finalContent = concludeResp.Content
		iteration++
	}

	return finalContent, iteration, nil
}

// runLLMIterationStream is like runLLMIteration but streams the final text response.
// Tool call iterations use non-streaming Chat(). Only the last iteration (no tool calls)
// uses ChatStream() if the provider supports it.
func (al *AgentLoop) runLLMIterationStream(ctx context.Context, messages []providers.Message, opts processOptions, onToken StreamCallback) (string, int, error) {
	iteration := 0
	var finalContent string
	pendingSynthesis := false
	synthesisActive := false

	model := al.model
	if opts.ModelOverride != "" {
		model = opts.ModelOverride
	}

	streamingProvider, canStream := al.provider.(providers.StreamingLLMProvider)

	for iteration < al.maxIterations {
		iteration++

		// synthesis_start is emitted before the orchestrator begins the final
		// synthesis step that consolidates specialist responses into one summary.
		if pendingSynthesis && !synthesisActive {
			emitAgentStatus(ctx, AgentStatusEvent{
				Agent:     "orchestrator",
				Status:    "synthesis_start",
				Timestamp: time.Now(),
			})
			synthesisActive = true
			pendingSynthesis = false
		}

		logger.DebugCF("agent", "LLM streaming iteration",
			map[string]interface{}{
				"iteration":  iteration,
				"max":        al.maxIterations,
				"can_stream": canStream,
			})

		// Build tool definitions (exclude all when "*" is in ExcludeTools)
		var providerToolDefs []providers.ToolDefinition
		excludeAll := false
		for _, ex := range opts.ExcludeTools {
			if ex == "*" {
				excludeAll = true
				break
			}
		}
		if !excludeAll {
			toolDefs := al.tools.GetDefinitions()
			providerToolDefs = make([]providers.ToolDefinition, 0, len(toolDefs))
			for _, td := range toolDefs {
				fnMap, ok := td["function"].(map[string]interface{})
				if !ok {
					continue
				}
				toolName, _ := fnMap["name"].(string)
				if toolName == "" {
					continue
				}
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
				tdType, _ := td["type"].(string)
				desc, _ := fnMap["description"].(string)
				params, _ := fnMap["parameters"].(map[string]interface{})
				providerToolDefs = append(providerToolDefs, providers.ToolDefinition{
					Type: tdType,
					Function: providers.ToolFunctionDefinition{
						Name:        toolName,
						Description: desc,
						Parameters:  params,
					},
				})
			}
		}

		llmOpts := map[string]interface{}{
			"max_tokens":  8192,
			"temperature": 0.7,
		}
		if opts.OnThinking != nil {
			llmOpts["extended_thinking"] = true
			llmOpts["thinking_budget_tokens"] = 1024
		}

		// Try streaming for this iteration
		if canStream {
			llmStart := time.Now()
			ch, err := streamingProvider.ChatStream(ctx, messages, providerToolDefs, model, llmOpts)
			if err != nil {
				al.metrics.RecordLLMCall(model, time.Since(llmStart), al.estimateTokens(messages), 0, err)
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
				if chunk.ThinkingDelta != "" && opts.OnThinking != nil {
					opts.OnThinking(chunk.ThinkingDelta)
				}

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
				for idx, tc := range chunk.ToolCalls {
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
			streamTokensOut := (len(streamContent) * 10) / 35 // ~3.5 chars/token estimate
			al.metrics.RecordLLMCall(model, time.Since(llmStart), al.estimateTokens(messages), streamTokensOut, nil)

			// If no tool calls — we're done
			if len(toolCalls) == 0 {
				// synthesis_end is emitted after the final synthesis step completes
				// and the unified orchestrator summary is ready to return.
				if synthesisActive {
					emitAgentStatus(ctx, AgentStatusEvent{
						Agent:     "orchestrator",
						Status:    "synthesis_end",
						Timestamp: time.Now(),
					})
					synthesisActive = false
				}
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
				argumentsJSON, err := json.Marshal(tc.Arguments)
				if err != nil {
					logger.WarnCF("agent", "Failed to marshal tool arguments", map[string]interface{}{"tool": tc.Name, "error": err.Error()})
					continue
				}
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

			specialistResponses := 0
			// Execute tool calls
			for _, tc := range toolCalls {
				argsJSON, err := json.Marshal(tc.Arguments)
				if err != nil {
					argsJSON = []byte(fmt.Sprintf("%v", tc.Arguments))
				}
				argsPreview := utils.Truncate(string(argsJSON), 200)
				logger.InfoCF("agent", fmt.Sprintf("Tool call: %s(%s)", tc.Name, argsPreview),
					map[string]interface{}{
						"tool":      tc.Name,
						"iteration": iteration,
					})

				// Security hook: before_tool_call
				if al.hooks != nil {
					hookResult := al.hooks.Run(ctx, "before_tool_call", hooks.HookContext{
						ToolName: tc.Name,
						Args:     tc.Arguments,
						UserID:   al.userID,
					})
					if hookResult.Action == hooks.HookBlock {
						messages = append(messages, providers.Message{
							Role:       "tool",
							Content:    fmt.Sprintf("[blocked by security hook: %s]", hookResult.Reason),
							ToolCallID: tc.ID,
						})
						continue
					}
					if hookResult.Action == hooks.HookRequireApproval {
						messages = append(messages, providers.Message{
							Role:       "tool",
							Content:    fmt.Sprintf("[tool requires approval: %s]", hookResult.Reason),
							ToolCallID: tc.ID,
						})
						logger.WarnCF("agent", "require_approval with no owner channel, treating as block", map[string]interface{}{
							"tool": tc.Name,
						})
						continue
					}
				}

				// Notify start of tool call
				if opts.OnTool != nil {
					_ = opts.OnTool(ToolEvent{Name: tc.Name, Args: tc.Arguments, Status: "started"})
				}

				toolStart := time.Now()
				result, err := al.tools.ExecuteWithContext(ctx, tc.Name, tc.Arguments, opts.Channel, opts.ChatID)
				toolDur := time.Since(toolStart)
				al.metrics.RecordToolCall(tc.Name, toolDur, err)
				if err != nil {
					result = fmt.Sprintf("Error: %v", err)
				}
				if tc.Name == "delegate_to_specialist" && err == nil && strings.TrimSpace(result) != "" {
					specialistResponses++
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
			if specialistResponses > 0 {
				pendingSynthesis = true
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
			fallbackTokensOut = (len(response.Content) * 10) / 35 // ~3.5 chars/token estimate
			// Use actual usage if available
			if response.Usage != nil {
				fallbackTokensIn = response.Usage.PromptTokens
				fallbackTokensOut = response.Usage.CompletionTokens
			}
		}
		al.metrics.RecordLLMCall(model, fallbackDur, fallbackTokensIn, fallbackTokensOut, err)

		// Record API call for cost tracking
		if err == nil && al.costTracker != nil {
			al.costTracker.RecordAPICall("main", int64(fallbackTokensIn), int64(fallbackTokensOut), model)
		}

		if err != nil {
			return "", iteration, fmt.Errorf("LLM call failed: %w", err)
		}

		if len(response.ToolCalls) == 0 {
			// synthesis_end is emitted after the fallback synthesis step completes
			// and the unified orchestrator summary is ready to return.
			if synthesisActive {
				emitAgentStatus(ctx, AgentStatusEvent{
					Agent:     "orchestrator",
					Status:    "synthesis_end",
					Timestamp: time.Now(),
				})
				synthesisActive = false
			}
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
		specialistResponses := 0
		for _, tc := range response.ToolCalls {
			argumentsJSON, err := json.Marshal(tc.Arguments)
			if err != nil {
				logger.WarnCF("agent", "Failed to marshal tool arguments", map[string]interface{}{"tool": tc.Name, "error": err.Error()})
				continue
			}
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
			argsJSON, err := json.Marshal(tc.Arguments)
			if err != nil {
				argsJSON = []byte(fmt.Sprintf("%v", tc.Arguments))
			}
			argsPreview := utils.Truncate(string(argsJSON), 200)
			logger.InfoCF("agent", fmt.Sprintf("Tool call: %s(%s)", tc.Name, argsPreview),
				map[string]interface{}{
					"tool":      tc.Name,
					"iteration": iteration,
				})

			// Security hook: before_tool_call
			if al.hooks != nil {
				hookResult := al.hooks.Run(ctx, "before_tool_call", hooks.HookContext{
					ToolName: tc.Name,
					Args:     tc.Arguments,
					UserID:   al.userID,
				})
				if hookResult.Action == hooks.HookBlock {
					messages = append(messages, providers.Message{
						Role:       "tool",
						Content:    fmt.Sprintf("[blocked by security hook: %s]", hookResult.Reason),
						ToolCallID: tc.ID,
					})
					continue
				}
				if hookResult.Action == hooks.HookRequireApproval {
					messages = append(messages, providers.Message{
						Role:       "tool",
						Content:    fmt.Sprintf("[tool requires approval: %s]", hookResult.Reason),
						ToolCallID: tc.ID,
					})
					logger.WarnCF("agent", "require_approval with no owner channel, treating as block", map[string]interface{}{
						"tool": tc.Name,
					})
					continue
				}
			}

			toolStart := time.Now()
			result, err := al.tools.ExecuteWithContext(ctx, tc.Name, tc.Arguments, opts.Channel, opts.ChatID)
			toolDur := time.Since(toolStart)
			al.metrics.RecordToolCall(tc.Name, toolDur, err)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}
			if tc.Name == "delegate_to_specialist" && err == nil && strings.TrimSpace(result) != "" {
				specialistResponses++
			}

			toolResultMsg := providers.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			}
			messages = append(messages, toolResultMsg)
			al.sessions.AddFullMessageForUser(al.userID, opts.SessionKey, toolResultMsg)
		}
		if specialistResponses > 0 {
			pendingSynthesis = true
		}
	}

	// If we exhausted iterations without a text response, force a concluding call
	// with no tools so the LLM must produce a final answer instead of looping.
	if finalContent == "" && iteration >= al.maxIterations {
		logger.WarnCF("agent", "Streaming iteration limit reached without text response, forcing concluding call",
			map[string]interface{}{
				"iterations": iteration,
				"max":        al.maxIterations,
			})

		// Append an instruction to wrap up
		messages = append(messages, providers.Message{
			Role:    "user",
			Content: "[System: You have reached the maximum number of tool iterations. You MUST now provide your final response as text. Summarize what you accomplished and any remaining items.]",
		})

		// Call LLM with NO tools so it must produce text
		concludeResp, err := al.provider.Chat(ctx, messages, nil, model, map[string]interface{}{
			"max_tokens":  4096,
			"temperature": 0.7,
		})
		if err != nil {
			logger.ErrorCF("agent", "Concluding streaming LLM call failed", map[string]interface{}{"error": err.Error()})
			return "", iteration, fmt.Errorf("concluding LLM call failed: %w", err)
		}
		// synthesis_end is emitted after the forced concluding synthesis step
		// completes and the unified orchestrator summary is ready.
		if synthesisActive {
			emitAgentStatus(ctx, AgentStatusEvent{
				Agent:     "orchestrator",
				Status:    "synthesis_end",
				Timestamp: time.Now(),
			})
			synthesisActive = false
		}
		finalContent = concludeResp.Content
		// Stream the concluding content to the client
		if onToken != nil && finalContent != "" {
			_ = onToken(finalContent)
		}
		iteration++
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
			al.summarizeWg.Add(1)
			go func() {
				defer al.summarizeWg.Done()
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
		if len(msg.ToolCalls) > 0 {
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
		// Estimate tokens for this message (~3.5 chars/token)
		msgTokens := (len(m.Content) * 10) / 35
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
// This is a rough heuristic (~3.5 chars/token for English, varies by language/code).
// When possible, actual token counts from provider Usage data should be preferred.
// Used as fallback for providers that don't report usage (e.g., Ollama).
func (al *AgentLoop) estimateTokens(messages []providers.Message) int {
	total := 0
	for _, m := range messages {
		// Use ~3.5 chars per token (conservative estimate for mixed content)
		// This errs on the side of triggering summarization earlier, which is safer
		// than hitting context limits. English averages ~4, code ~3, non-Latin ~1-2.
		total += (len(m.Content) * 10) / 35 // equivalent to / 3.5
	}
	return total
}
