package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/providers"
	"github.com/sipeed/makoclaw/pkg/skills"
	"github.com/sipeed/makoclaw/pkg/storage"
	"github.com/sipeed/makoclaw/pkg/tools"
)

type ContextBuilder struct {
	workspace         string
	userUUID          string // User UUID for multiuser support
	userID            int64  // User ID from database
	skillsLoader      *skills.SkillsLoader
	memory            *MemoryStore
	tools             *tools.ToolRegistry // Direct reference to tool registry
	agentSystemPrompt string              // Agent-specific system prompt (orchestrator/specialist role)
	skillFilter       []string            // nil=all skills, []string{}=no skills, ["x","y"]=only those
	lightweightMode   bool                // Omit bootstrap files and memory for lean context
	userStore         *storage.Storage    // Per-user storage for skill analytics (optional)
	centralStore      *storage.CentralStorage // Central storage for marketplace usage counts (optional)
	sessionKey        string              // Session key for skill analytics events
	userEmail         string              // User's email address for system prompt context
	ordersLoader      func() []storage.StandingOrder // Loader for active standing orders (optional)
}

func getGlobalConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".MakoClaw")
}

func NewContextBuilder(workspace string) *ContextBuilder {
	// builtin skills: skills directory in current project
	// Use the skills/ directory under the current working directory
	wd, _ := os.Getwd()
	builtinSkillsDir := filepath.Join(wd, "skills")
	globalSkillsDir := filepath.Join(getGlobalConfigDir(), "skills")

	return &ContextBuilder{
		workspace:    workspace,
		userUUID:     "", // Will be set via WithUser()
		userID:       0,  // Default userID = 0 (backward compatibility)
		skillsLoader: skills.NewSkillsLoader(workspace, globalSkillsDir, builtinSkillsDir),
		memory:       NewMemoryStore(workspace),
	}
}

// WithUserEmail stores the user's email for inclusion in the system prompt.
func (cb *ContextBuilder) WithUserEmail(email string) *ContextBuilder {
	cb.userEmail = email
	return cb
}

// WithUser sets the user UUID and ID for this context builder.
// This enables multiuser support by resolving workspace paths to user-specific directories.
func (cb *ContextBuilder) WithUser(userUUID string, userID int64) *ContextBuilder {
	cb.userUUID = userUUID
	cb.userID = userID

	userWorkspace := cb.getUserWorkspacePath()
	cb.workspace = userWorkspace
	cb.memory = NewMemoryStore(userWorkspace)

	// Rebuild skills loader to point at user workspace and user skills path
	wd, _ := os.Getwd()
	builtinSkillsDir := filepath.Join(wd, "skills")
	globalSkillsDir := filepath.Join(getGlobalConfigDir(), "skills")
	cb.skillsLoader = skills.NewSkillsLoader(userWorkspace, globalSkillsDir, builtinSkillsDir)

	home, err := os.UserHomeDir()
	if err != nil {
		logger.WarnCF("agent", "Failed to get user home dir, user skills not loaded",
			map[string]interface{}{
				"user_uuid": userUUID,
				"error":     err.Error(),
			})
		return cb
	}
	userSkillsPath := filepath.Join(home, ".MakoClaw", "users", userUUID, "skills")
	cb.skillsLoader.SetUserSkillsPath(userSkillsPath)

	return cb
}

// getUserWorkspacePath returns the workspace path, either user-specific or global.
func (cb *ContextBuilder) getUserWorkspacePath() string {
	if cb.userUUID == "" {
		return cb.workspace
	}
	// For multiuser: replace workspace with user-scoped path
	home, err := os.UserHomeDir()
	if err != nil {
		return cb.workspace
	}
	return filepath.Join(home, ".MakoClaw", "users", cb.userUUID, "workspace")
}

// SetToolsRegistry sets the tools registry for dynamic tool summary generation.
func (cb *ContextBuilder) SetToolsRegistry(registry *tools.ToolRegistry) {
	cb.tools = registry
}

// SetAgentSystemPrompt sets an agent-specific system prompt that is injected
// after the identity section. Used by the orchestrator (delegation context) and
// specialists (role-specific instructions).
func (cb *ContextBuilder) SetAgentSystemPrompt(prompt string) {
	cb.agentSystemPrompt = prompt
}

// SetSkillFilter controls which skills are included in the system prompt.
//   - nil: load all skills (default behavior)
//   - empty slice: load no skills
//   - non-empty slice: load only the named skills
func (cb *ContextBuilder) SetSkillFilter(skillNames []string) {
	cb.skillFilter = skillNames
}

// SetLightweightMode enables/disables lightweight context mode.
// When enabled, bootstrap files (AGENTS.md, SOUL.md, etc.) and memory
// are omitted from the system prompt to reduce token usage.
func (cb *ContextBuilder) SetLightweightMode(enabled bool) {
	cb.lightweightMode = enabled
}

// WithAnalytics configures per-user skill usage analytics.
// When set, each call to BuildSystemPrompt will asynchronously record
// the loaded skills as usage events in the per-user storage.
func (cb *ContextBuilder) WithAnalytics(store *storage.Storage, sessionKey string) *ContextBuilder {
	cb.userStore = store
	cb.sessionKey = sessionKey
	return cb
}

// WithCentralStore sets the central storage for incrementing marketplace usage counts.
func (cb *ContextBuilder) WithCentralStore(cs *storage.CentralStorage) *ContextBuilder {
	cb.centralStore = cs
	return cb
}

// WithStandingOrders sets a loader function for active standing orders.
// When set, active standing orders are injected into the system prompt before skills.
func (cb *ContextBuilder) WithStandingOrders(loader func() []storage.StandingOrder) *ContextBuilder {
	cb.ordersLoader = loader
	return cb
}

func (cb *ContextBuilder) getIdentity() string {
	now := time.Now().Format("2006-01-02 15:04 (Monday)")
	workspacePath, _ := filepath.Abs(cb.getUserWorkspacePath())
	runtime := fmt.Sprintf("%s %s, Go %s", runtime.GOOS, runtime.GOARCH, runtime.Version())

	// Build tools section dynamically
	toolsSection := cb.buildToolsSection()

	return fmt.Sprintf(`# makoclaw 🦈

You are makoclaw, a helpful AI assistant.

## Current Time
%s

## Runtime
%s

## Workspace
Your workspace is at: %s
- Memory: %s/memory/MEMORY.md
- Daily Notes: %s/memory/YYYYMM/YYYYMMDD.md
- Skills: %s/skills/{skill-name}/SKILL.md

%s

## Important Rules

1. **ALWAYS use tools** - When you need to perform an action (schedule reminders, send messages, execute commands, etc.), you MUST call the appropriate tool. Do NOT just say you'll do it or pretend to do it.

2. **Be helpful and accurate** - When using tools, briefly explain what you're doing.

3. **Memory** - When remembering something, write to %s/memory/MEMORY.md

## Security & Permissions

**File Operations**: When workspace restriction is enabled, file operations are limited to the workspace directory.

**Shell Commands**: The exec tool has safety guards that block dangerous patterns (rm -rf, format, etc.) and enforces workspace restrictions for file paths.

**HTTP/Network**: ⚠️ HTTP connections are ALLOWED. Tools and skills can make external HTTP requests (e.g., gh CLI for GitHub API, curl, wget, API calls). There is NO safety guard blocking external URLs. You can freely use:
- GitHub CLI (gh) for api.github.com
- Web search and fetch tools
- Any external API or service
- Skills that require network access

If a command needs network access, you should execute it normally.`,
		now, runtime, workspacePath, workspacePath, workspacePath, workspacePath, toolsSection, workspacePath)
}

func (cb *ContextBuilder) buildToolsSection() string {
	if cb.tools == nil {
		return ""
	}

	summaries := cb.tools.GetSummaries()
	if len(summaries) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Available Tools\n\n")
	sb.WriteString("**CRITICAL**: You MUST use tools to perform actions. Do NOT pretend to execute commands or schedule tasks.\n\n")
	sb.WriteString("You have access to the following tools:\n\n")
	for _, s := range summaries {
		sb.WriteString(s)
		sb.WriteString("\n")
	}

	// Add contextual hints for specific capabilities
	if _, hasEmail := cb.tools.Get("send_email_report"); hasEmail {
		if cb.userEmail != "" {
			sb.WriteString(fmt.Sprintf("\n**Email**: You have email configured and ready to use. The user's email is `%s`. Use the `send_email_report` tool to send reports, summaries, data exports, and important notifications. The recipient email is pre-configured — do not ask the user for their email address.\n", cb.userEmail))
		} else {
			sb.WriteString("\n**Email**: You have email configured and ready to use. Use the `send_email_report` tool to send reports, summaries, data exports, and important notifications to the user. The recipient email is pre-configured — do not ask the user for their email address.\n")
		}
	}
	if _, hasReadEmail := cb.tools.Get("read_email"); hasReadEmail {
		sb.WriteString("\n**Email Inbox**: You can read the user's inbox using the `read_email` tool. Use it to check for new emails, read specific messages, or search by sender.\n")
	}

	return sb.String()
}

func (cb *ContextBuilder) BuildSystemPrompt() string {
	parts := []string{}

	// Core identity section
	parts = append(parts, cb.getIdentity())

	// Agent-specific system prompt (orchestrator delegation context, specialist role, etc.)
	if cb.agentSystemPrompt != "" {
		parts = append(parts, cb.agentSystemPrompt)
	}

	// Bootstrap files (skip in lightweight mode for specialists/orchestrator)
	if !cb.lightweightMode {
		bootstrapContent := cb.LoadBootstrapFiles()
		if bootstrapContent != "" {
			parts = append(parts, bootstrapContent)
		}
	} else {
		// In lightweight mode, still load personality files (SOUL.md, IDENTITY.md)
		personalityContent := cb.LoadPersonalityFiles()
		if personalityContent != "" {
			parts = append(parts, personalityContent)
		}
	}

	// Standing orders — persistent user instructions injected before skills
	if cb.ordersLoader != nil {
		orders := cb.ordersLoader()
		if len(orders) > 0 {
			var sb strings.Builder
			sb.WriteString("## Standing Orders\n\nThese are persistent instructions from the user. Follow them in every interaction.\n\n")
			for _, o := range orders {
				sb.WriteString("- ")
				sb.WriteString(o.Content)
				sb.WriteString("\n")
			}
			parts = append(parts, sb.String())
		}
	}

	// Skills - filtered by skillFilter setting
	var skillsSummary string
	if cb.skillFilter != nil {
		// Explicit filter set
		if len(cb.skillFilter) > 0 {
			skillsSummary = cb.skillsLoader.BuildSkillsSummaryForNames(cb.skillFilter)
		}
		// else: empty slice = no skills loaded
	} else {
		// nil = load all skills (default)
		skillsSummary = cb.skillsLoader.BuildSkillsSummary()
	}
	if skillsSummary != "" {
		parts = append(parts, fmt.Sprintf(`# Skills

The following skills extend your capabilities. To use a skill, read its SKILL.md file using the read_file tool.

%s`, skillsSummary))
	}

	// Record skill usage analytics asynchronously (fire-and-forget, never blocks the prompt build).
	if cb.userStore != nil && cb.skillsLoader != nil {
		loadedSkills := cb.skillsLoader.ListSkills()
		if len(loadedSkills) > 0 {
			userStore := cb.userStore
			centralStore := cb.centralStore
			sessionKey := cb.sessionKey
			go func() {
				for _, sk := range loadedSkills {
					_ = userStore.RecordSkillUsage(sk.Name, sk.Source, sessionKey)
					if centralStore != nil {
						_ = centralStore.IncrementSkillUsageCount(sk.Slug)
					}
				}
			}()
		}
	}

	// Memory context (skip in lightweight mode)
	if !cb.lightweightMode {
		memoryContext := cb.memory.GetMemoryContext()
		if memoryContext != "" {
			parts = append(parts, "# Memory\n\n"+memoryContext)
		}
	}

	// Join with "---" separator
	return strings.Join(parts, "\n\n---\n\n")
}

func (cb *ContextBuilder) LoadBootstrapFiles() string {
	bootstrapFiles := []string{
		"AGENTS.md",
		"SOUL.md",
		"USER.md",
		"IDENTITY.md",
	}

	var result string
	for _, filename := range bootstrapFiles {
		filePath := filepath.Join(cb.getUserWorkspacePath(), filename)
		if data, err := os.ReadFile(filePath); err == nil {
			result += fmt.Sprintf("## %s\n\n%s\n\n", filename, string(data))
		}
	}

	return result
}

// LoadPersonalityFiles loads only SOUL.md and IDENTITY.md from the workspace.
// Used in lightweight mode so that orchestrator and specialists retain personality
// even when the full bootstrap files are skipped.
func (cb *ContextBuilder) LoadPersonalityFiles() string {
	personalityFiles := []string{"SOUL.md", "IDENTITY.md"}
	var result string
	for _, filename := range personalityFiles {
		filePath := filepath.Join(cb.getUserWorkspacePath(), filename)
		if data, err := os.ReadFile(filePath); err == nil {
			result += fmt.Sprintf("## %s\n\n%s\n\n", filename, string(data))
		}
	}
	return result
}

func (cb *ContextBuilder) BuildMessages(history []providers.Message, summary string, currentMessage string, media []string, channel, chatID string) []providers.Message {
	messages := []providers.Message{}

	systemPrompt := cb.BuildSystemPrompt()

	// Add Current Session info if provided
	if channel != "" && chatID != "" {
		systemPrompt += fmt.Sprintf("\n\n## Current Session\nChannel: %s\nChat ID: %s", channel, chatID)
	}

	// Log system prompt summary for debugging (debug mode only)
	logger.DebugCF("agent", "System prompt built",
		map[string]interface{}{
			"total_chars":   len(systemPrompt),
			"total_lines":   strings.Count(systemPrompt, "\n") + 1,
			"section_count": strings.Count(systemPrompt, "\n\n---\n\n") + 1,
		})

	// Log preview of system prompt (avoid logging huge content)
	preview := systemPrompt
	if len(preview) > 500 {
		preview = preview[:500] + "... (truncated)"
	}
	logger.DebugCF("agent", "System prompt preview",
		map[string]interface{}{
			"preview": preview,
		})

	if summary != "" {
		systemPrompt += "\n\n## Summary of Previous Conversation\n\n" + summary
	}

	//This fix prevents the session memory from LLM failure due to elimination of toolu_IDs required from LLM
	// --- INICIO DEL FIX ---
	//Diegox-17
	for len(history) > 0 && (history[0].Role == "tool") {
		logger.DebugCF("agent", "Removing orphaned tool message from history to prevent LLM error",
			map[string]interface{}{"role": history[0].Role})
		history = history[1:]
	}
	//Diegox-17
	// --- FIN DEL FIX ---

	messages = append(messages, providers.Message{
		Role:    "system",
		Content: systemPrompt,
	})

	messages = append(messages, history...)

	messages = append(messages, providers.Message{
		Role:    "user",
		Content: currentMessage,
	})

	return messages
}

func (cb *ContextBuilder) AddToolResult(messages []providers.Message, toolCallID, toolName, result string) []providers.Message {
	messages = append(messages, providers.Message{
		Role:       "tool",
		Content:    result,
		ToolCallID: toolCallID,
	})
	return messages
}

func (cb *ContextBuilder) AddAssistantMessage(messages []providers.Message, content string, toolCalls []map[string]interface{}) []providers.Message {
	msg := providers.Message{
		Role:    "assistant",
		Content: content,
	}
	// Always add assistant message, whether or not it has tool calls
	messages = append(messages, msg)
	return messages
}

func (cb *ContextBuilder) loadSkills() string {
	allSkills := cb.skillsLoader.ListSkills()
	if len(allSkills) == 0 {
		return ""
	}

	var skillNames []string
	for _, s := range allSkills {
		skillNames = append(skillNames, s.Name)
	}

	content := cb.skillsLoader.LoadSkillsForContext(skillNames)
	if content == "" {
		return ""
	}

	return "# Skill Definitions\n\n" + content
}

// GetSkillsInfo returns information about loaded skills.
func (cb *ContextBuilder) GetSkillsInfo() map[string]interface{} {
	allSkills := cb.skillsLoader.ListSkills()
	skillNames := make([]string, 0, len(allSkills))
	for _, s := range allSkills {
		skillNames = append(skillNames, s.Name)
	}
	return map[string]interface{}{
		"total":     len(allSkills),
		"available": len(allSkills),
		"names":     skillNames,
	}
}
