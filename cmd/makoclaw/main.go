// makoclaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 makoclaw contributors

package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/sipeed/makoclaw/pkg/agent"
	"github.com/sipeed/makoclaw/pkg/auth"
	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/channels"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/cron"
	"github.com/sipeed/makoclaw/pkg/doctor"
	"github.com/sipeed/makoclaw/pkg/heartbeat"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/mcp"
	"github.com/sipeed/makoclaw/pkg/migrate"
	"github.com/sipeed/makoclaw/pkg/providers"
	"github.com/sipeed/makoclaw/pkg/skills"
	"github.com/sipeed/makoclaw/pkg/storage"
	"github.com/sipeed/makoclaw/pkg/tools"
	"github.com/sipeed/makoclaw/pkg/voice"
	"github.com/sipeed/makoclaw/pkg/web"
	"github.com/sipeed/makoclaw/pkg/workflow"
)

var (
	version   = "0.2.0"
	buildTime string
	goVersion string
)

const logo = "🦈"

func printVersion() {
	fmt.Printf("%s makoclaw v%s\n", logo, version)
	if buildTime != "" {
		fmt.Printf("  Build: %s\n", buildTime)
	}
	goVer := goVersion
	if goVer == "" {
		goVer = runtime.Version()
	}
	if goVer != "" {
		fmt.Printf("  Go: %s\n", goVer)
	}
}

func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "onboard":
		onboard()
	case "agent":
		agentCmd()
	case "gateway":
		gatewayCmd()
	case "web":
		webCmd()
	case "status":
		statusCmd()
	case "migrate":
		migrateCmd()
	case "migrate-multiuser":
		migrateMultiuserCmd()
	case "auth":
		authCmd()
	case "doctor":
		doctorCmd()
	case "cron":
		cronCmd()
	case "list-users":
		listUsersCmd()
	case "reset-password":
		resetPasswordCmd()
	case "skills":
		if len(os.Args) < 3 {
			skillsHelp()
			return
		}

		subcommand := os.Args[2]

		cfg, err := loadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		workspace := cfg.WorkspacePath()
		installer := skills.NewSkillInstaller(workspace)
		// 获取全局配置目录和内置 skills 目录
		globalDir := filepath.Dir(getConfigPath())
		globalSkillsDir := filepath.Join(globalDir, "skills")
		builtinSkillsDir := resolveBuiltinSkillsDir()
		skillsLoader := skills.NewSkillsLoader(workspace, globalSkillsDir, builtinSkillsDir)

		switch subcommand {
		case "list":
			skillsListCmd(skillsLoader)
		case "install":
			skillsInstallCmd(installer)
		case "remove", "uninstall":
			if len(os.Args) < 4 {
				fmt.Println("Usage: makoclaw skills remove <skill-name>")
				return
			}
			skillsRemoveCmd(installer, os.Args[3])
		case "install-builtin":
			skillsInstallBuiltinCmd(workspace)
		case "list-builtin":
			skillsListBuiltinCmd()
		case "search":
			skillsSearchCmd(installer)
		case "show":
			if len(os.Args) < 4 {
				fmt.Println("Usage: makoclaw skills show <skill-name>")
				return
			}
			skillsShowCmd(skillsLoader, os.Args[3])
		default:
			fmt.Printf("Unknown skills command: %s\n", subcommand)
			skillsHelp()
		}
	case "version", "--version", "-v":
		printVersion()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf("%s makoclaw - Personal AI Assistant v%s\n\n", logo, version)
	fmt.Println("Usage: makoclaw <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  onboard     Initialize makoclaw configuration and workspace")
	fmt.Println("  agent       Interact with the agent directly")
	fmt.Println("  auth        Manage authentication (login, logout, status)")
	fmt.Println("  doctor      Run health checks and diagnostics")
	fmt.Println("  gateway     Start makoclaw gateway")
	fmt.Println("  web         Start web panel only")
	fmt.Println("  status      Show makoclaw status")
	fmt.Println("  cron        Manage scheduled tasks")
	fmt.Println("  migrate     Migrate from OpenClaw to makoclaw")
	fmt.Println("  migrate-multiuser  Migrate legacy data to multiuser layout")
	fmt.Println("  list-users  List all registered users")
	fmt.Println("  reset-password  Reset a user's password")
	fmt.Println("  skills      Manage skills (install, list, remove)")
	fmt.Println("  version     Show version information")
}

func onboard() {
	configPath := getConfigPath()

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config already exists at %s\n", configPath)
		fmt.Print("Overwrite? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" {
			fmt.Println("Aborted.")
			return
		}
	}

	cfg := config.DefaultConfig()
	if err := config.SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	workspace := cfg.WorkspacePath()
	os.MkdirAll(workspace, 0755)
	os.MkdirAll(filepath.Join(workspace, "memory"), 0755)
	os.MkdirAll(filepath.Join(workspace, "skills"), 0755)

	createWorkspaceTemplates(workspace)

	fmt.Printf("%s makoclaw is ready!\n", logo)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add your API key to", configPath)
	fmt.Println("     Get one at: https://openrouter.ai/keys")
	fmt.Println("  2. Chat: makoclaw agent -m \"Hello!\"")
}

func createWorkspaceTemplates(workspace string) {
	templates := map[string]string{
		"AGENTS.md": `# Agent Instructions

You are a helpful AI assistant. Be concise, accurate, and friendly.

## Guidelines

- Always explain what you're doing before taking actions
- Ask for clarification when request is ambiguous
- Use tools to help accomplish tasks
- Remember important information in your memory files
- Be proactive and helpful
- Learn from user feedback
`,
		"SOUL.md": `# Soul

I am makoclaw, a lightweight AI assistant powered by AI.

## Personality

- Helpful and friendly
- Concise and to the point
- Curious and eager to learn
- Honest and transparent

## Values

- Accuracy over speed
- User privacy and safety
- Transparency in actions
- Continuous improvement
`,
		"USER.md": `# User

Information about user goes here.

## Preferences

- Communication style: (casual/formal)
- Timezone: (your timezone)
- Language: (your preferred language)

## Personal Information

- Name: (optional)
- Location: (optional)
- Occupation: (optional)

## Learning Goals

- What the user wants to learn from AI
- Preferred interaction style
- Areas of interest
`,
		"IDENTITY.md": `# Identity

## Name
makoclaw 🦈 (The Apex AI Agent)

## Description
An ultra-efficient, Go-native personal AI assistant designed for the most demanding efficiency requirements.

## Version
0.1.0

## Purpose
- Deliver supreme AI intelligence with sub-10MB RAM footprint.
- Empower low-cost hardware ($10 boards) with high-tier agent capabilities.
- Self-bootstrapped and AI-refined architectural perfection.

## Capabilities
- 🚀 Instant startup (<1s)
- 📡 Multi-channel global communication
- 🛠️ Advanced tool orchestration (Brave Search, Files, Shell, MCP)
- 📅 Precision task scheduling and cron management
- 🧠 Long-term memory and context retention
- 🎙️ Native voice transcription

## Philosophy
- **Apex Efficiency**: Every bit is optimized for maximum impact.
- **Transparency**: Clear, auditable, and user-centric operations.
- **Independence**: Native implementation with zero heavy dependencies.

## License
MIT License - Free, Open, and Unstoppable.

## Repository
https://github.com/sipeed/makoclaw

---

"Maximum intelligence. Minimum footprint."
- makoclaw (The Apex Agent)
`,
	}

	for filename, content := range templates {
		filePath := filepath.Join(workspace, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			os.WriteFile(filePath, []byte(content), 0644)
			fmt.Printf("  Created %s\n", filename)
		}
	}

	memoryDir := filepath.Join(workspace, "memory")
	os.MkdirAll(memoryDir, 0755)
	memoryFile := filepath.Join(memoryDir, "MEMORY.md")
	if _, err := os.Stat(memoryFile); os.IsNotExist(err) {
		memoryContent := `# Long-term Memory

This file stores important information that should persist across sessions.

## User Information

(Important facts about user)

## Preferences

(User preferences learned over time)

## Important Notes

(Things to remember)

## Configuration

- Model preferences
- Channel settings
- Skills enabled
`
		os.WriteFile(memoryFile, []byte(memoryContent), 0644)
		fmt.Println("  Created memory/MEMORY.md")

		skillsDir := filepath.Join(workspace, "skills")
		if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
			os.MkdirAll(skillsDir, 0755)
			fmt.Println("  Created skills/")
		}
	}
}

func migrateCmd() {
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		migrateHelp()
		return
	}

	opts := migrate.Options{}

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			opts.DryRun = true
		case "--config-only":
			opts.ConfigOnly = true
		case "--workspace-only":
			opts.WorkspaceOnly = true
		case "--force":
			opts.Force = true
		case "--refresh":
			opts.Refresh = true
		case "--openclaw-home":
			if i+1 < len(args) {
				opts.OpenClawHome = args[i+1]
				i++
			}
		case "--makoclaw-home":
			if i+1 < len(args) {
				opts.MakoClawHome = args[i+1]
				i++
			}
		default:
			fmt.Printf("Unknown flag: %s\n", args[i])
			migrateHelp()
			os.Exit(1)
		}
	}

	result, err := migrate.Run(opts)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !opts.DryRun {
		migrate.PrintSummary(result)
	}
}

func migrateMultiuserCmd() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Storage.Path == "" {
		fmt.Println("Storage path is not configured; cannot migrate multiuser data")
		os.Exit(1)
	}

	store, err := storage.New(cfg.Storage)
	if err != nil {
		fmt.Printf("Error opening storage: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	// Initialize central database for the new architecture
	dataDir := filepath.Dir(expandHomePath(cfg.Storage.Path))
	config.InitDataDir(dataDir)
	centralDBPath := filepath.Join(dataDir, "central.db")
	centralStore, err := storage.NewCentral(centralDBPath)
	if err != nil {
		fmt.Printf("Error initializing central database: %v\n", err)
		os.Exit(1)
	}
	defer centralStore.Close()
	fmt.Println("✓ Central database initialized")

	// Seed users from legacy store into central database
	legacyUsers, err := store.ListUsers()
	if err != nil {
		fmt.Printf("Warning: Could not list legacy users: %v\n", err)
	} else {
		migratedCount := 0
		for _, u := range legacyUsers {
			// Check if user already exists in central
			if _, err := centralStore.GetUserByUsername(u.Username); err == nil {
				continue // already migrated
			}
			// Create user in central store (preserving UUID and role)
			if _, err := centralStore.CreateUser(u.Username, "temporary-migration-password", u.Role); err != nil {
				fmt.Printf("  Warning: Could not migrate user %s: %v\n", u.Username, err)
				continue
			}
			migratedCount++
		}
		if migratedCount > 0 {
			fmt.Printf("✓ Migrated %d users to central database\n", migratedCount)
		} else {
			fmt.Println("✓ All users already in central database")
		}
	}

	// Initialize per-user storage manager
	userMgr := storage.NewUserStorageManager(centralStore, dataDir)
	defer userMgr.Close()

	// Determine the target user for workspace migration
	var user *storage.User
	if strings.TrimSpace(cfg.Web.Username) != "" {
		if u, err := centralStore.GetUserByUsername(cfg.Web.Username); err == nil {
			user = u
		}
	}
	if user == nil {
		users, err := centralStore.ListUsers()
		if err != nil || len(users) == 0 {
			fmt.Println("No users found; create a user first")
			os.Exit(1)
		}
		user = users[0]
	}

	// Ensure user directory structure exists
	if err := userMgr.EnsureUserDirectory(user.UUID); err != nil {
		fmt.Printf("Warning: Failed to create user directory: %v\n", err)
	}

	// Create per-user database
	if _, err := userMgr.GetOrCreate(user.UUID); err != nil {
		fmt.Printf("Warning: Failed to create per-user database: %v\n", err)
	}

	// Run existing migration logic (workspace move, config copy, data backfill)
	res, err := migrate.MigrateToMultiuser(cfg, store, user)
	if err != nil {
		fmt.Printf("Multiuser migration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Multiuser migration completed")
	if res.WorkspaceMoved {
		fmt.Println("  - Workspace moved to user directory")
	}
	if res.ConfigCopied {
		fmt.Println("  - Config copied to user directory")
	}
	if res.DataBackfilled {
		fmt.Println("  - Database user_id backfilled")
	}
	fmt.Printf("  - User UUID: %s\n", user.UUID)
	fmt.Printf("  - User workspace: %s\n", userMgr.UserWorkspacePath(user.UUID))
	for _, w := range res.Warnings {
		fmt.Printf("  - Warning: %s\n", w)
	}
}

func migrateHelp() {
	fmt.Println("\nMigrate from OpenClaw to makoclaw")
	fmt.Println()
	fmt.Println("Usage: makoclaw migrate [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --dry-run          Show what would be migrated without making changes")
	fmt.Println("  --refresh          Re-sync workspace files from OpenClaw (repeatable)")
	fmt.Println("  --config-only      Only migrate config, skip workspace files")
	fmt.Println("  --workspace-only   Only migrate workspace files, skip config")
	fmt.Println("  --force            Skip confirmation prompts")
	fmt.Println("  --openclaw-home    Override OpenClaw home directory (default: ~/.openclaw)")
	fmt.Println("  --makoclaw-home    Override makoclaw home directory (default: ~/.MakoClaw)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  makoclaw migrate              Detect and migrate from OpenClaw")
	fmt.Println("  makoclaw migrate --dry-run    Show what would be migrated")
	fmt.Println("  makoclaw migrate --refresh    Re-sync workspace files")
	fmt.Println("  makoclaw migrate --force      Migrate without confirmation")
}

func agentCmd() {
	message := ""
	sessionKey := "cli:default"

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--debug", "-d":
			logger.SetLevel(logger.DEBUG)
			fmt.Println("🔍 Debug mode enabled")
		case "-m", "--message":
			if i+1 < len(args) {
				message = args[i+1]
				i++
			}
		case "-s", "--session":
			if i+1 < len(args) {
				sessionKey = args[i+1]
				i++
			}
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	provider, err := providers.CreateProvider(cfg)
	if err != nil {
		fmt.Printf("Error creating provider: %v\n", err)
		os.Exit(1)
	}

	msgBus := bus.NewMessageBus()
	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider)

	// Register CronTool so the CLI agent can manage cron jobs via API
	// instead of falling back to write_file on jobs.json.
	cronService := setupCronTool(agentLoop, msgBus, cfg.WorkspacePath())
	if err := cronService.Start(); err != nil {
		logger.WarnCF("agent", "Failed to start cron service", map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer cronService.Stop()

	var channelStore *storage.Storage
	if cfg.Storage.Path != "" {
		if store, err := storage.New(cfg.Storage); err == nil {
			channelStore = store
		} else {
			fmt.Printf("Warning: Failed to initialize storage for channels: %v\n", err)
		}
	}
	if channelStore != nil {
		defer channelStore.Close()
	}

	// Print agent startup info (only for interactive mode)
	startupInfo := agentLoop.GetStartupInfo()
	logger.InfoCF("agent", "Agent initialized",
		map[string]interface{}{
			"tools_count":      startupInfo["tools"].(map[string]interface{})["count"],
			"skills_total":     startupInfo["skills"].(map[string]interface{})["total"],
			"skills_available": startupInfo["skills"].(map[string]interface{})["available"],
		})

	if message != "" {
		ctx := context.Background()
		response, err := agentLoop.ProcessDirect(ctx, message, sessionKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n%s %s\n", logo, response)
	} else {
		fmt.Printf("%s Interactive mode (Ctrl+C to exit)\n\n", logo)
		interactiveMode(agentLoop, sessionKey)
	}
}

func interactiveMode(agentLoop *agent.AgentLoop, sessionKey string) {
	prompt := fmt.Sprintf("%s You: ", logo)

	rl, err := readline.NewEx(&readline.Config{
		Prompt: prompt,
		HistoryFile: func() string {
			if cacheDir, err := os.UserCacheDir(); err == nil {
				_ = os.MkdirAll(filepath.Join(cacheDir, "makoclaw"), 0700)
				return filepath.Join(cacheDir, "makoclaw", ".makoclaw_history")
			}
			if homeDir, err := os.UserHomeDir(); err == nil {
				return filepath.Join(homeDir, ".makoclaw_history")
			}
			return filepath.Join(os.TempDir(), ".makoclaw_history")
		}(),
		HistoryLimit:    100,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	if err != nil {
		fmt.Printf("Error initializing readline: %v\n", err)
		fmt.Println("Falling back to simple input mode...")
		simpleInteractiveMode(agentLoop, sessionKey)
		return
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				fmt.Println("\nGoodbye!")
				return
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			return
		}

		ctx := context.Background()
		response, err := agentLoop.ProcessDirect(ctx, input, sessionKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\n%s %s\n\n", logo, response)
	}
}

func simpleInteractiveMode(agentLoop *agent.AgentLoop, sessionKey string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s You: ", logo)
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nGoodbye!")
				return
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			return
		}

		ctx := context.Background()
		response, err := agentLoop.ProcessDirect(ctx, input, sessionKey)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\n%s %s\n\n", logo, response)
	}
}

func gatewayCmd() {
	// Check for --debug flag
	args := os.Args[2:]
	for _, arg := range args {
		if arg == "--debug" || arg == "-d" {
			logger.SetLevel(logger.DEBUG)
			fmt.Println("🔍 Debug mode enabled")
			break
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	msgBus := bus.NewMessageBus()

	// Initialize storage for channels (required for multiuser)
	var channelStore *storage.Storage
	var centralStore *storage.CentralStorage
	var userMgr *storage.UserStorageManager
	if cfg.Storage.Path != "" {
		if store, err := storage.New(cfg.Storage); err == nil {
			channelStore = store
		} else {
			fmt.Printf("Warning: Failed to initialize storage: %v\n", err)
		}
		// Initialize central database (users, settings) alongside main storage
		dataDir := filepath.Dir(expandHomePath(cfg.Storage.Path))
		config.InitDataDir(dataDir)
		centralDBPath := filepath.Join(dataDir, "central.db")
		if cs, err := storage.NewCentral(centralDBPath); err == nil {
			centralStore = cs
			fmt.Println("✓ Central database initialized")
		} else {
			fmt.Printf("Warning: Failed to initialize central storage: %v\n", err)
		}
		// Initialize per-user storage manager
		if centralStore != nil {
			userMgr = storage.NewUserStorageManager(centralStore, dataDir)
		}
		fmt.Println("✓ Per-user storage manager ready")
	}

	// Create multi-user channel manager (each user gets their own channels and agent)
	multiChannelManager := channels.NewMultiUserChannelManager(cfg, msgBus, channelStore, centralStore)

	// Initialize channels for all users from database
	if channelStore != nil {
		if err := multiChannelManager.InitializeAllUsers(); err != nil {
			logger.WarnCF("gateway", "Failed to initialize user channels", map[string]interface{}{
				"error": err.Error(),
			})
		}
		fmt.Printf("✓ Multi-user system initialized (%d users)\n", multiChannelManager.GetUserCount())
	} else {
		fmt.Println("⚠ Warning: Storage not available, multiuser features disabled")
		// Fallback: create a default agent loop for backward compatibility
		provider, err := providers.TryCreateProvider(cfg)
		if err != nil {
			// Provider configuration error - enter degraded mode instead of exiting
			fmt.Printf("Error creating provider: %v\n", err)
			fmt.Println("Starting in DEGRADED MODE. Configure provider via web panel to enable agent features.")
			provider = nil // Treat error as no provider (degraded mode)
		}
		if provider == nil {
			fmt.Println("⚠ DEGRADED MODE: No LLM provider configured")
			fmt.Println("  Agent loop disabled. Configure provider via web panel to enable.")
			cfg.DegradedMode = true
		} else {
			agentLoop := agent.NewAgentLoop(cfg, msgBus, provider)
			// Print agent startup info
			fmt.Println("\n📦 Agent Status:")
			startupInfo := agentLoop.GetStartupInfo()
			toolsInfo := startupInfo["tools"].(map[string]interface{})
			skillsInfo := startupInfo["skills"].(map[string]interface{})
			fmt.Printf("  • Tools: %d loaded\n", toolsInfo["count"])
			fmt.Printf("  • Skills: %d/%d available\n",
				skillsInfo["available"],
				skillsInfo["total"])
		}
	}

	var defaultAgentLoop *agent.AgentLoop
	defaultProvider, err := providers.TryCreateProvider(cfg)
	if err != nil {
		fmt.Printf("Error with provider configuration: %v\n", err)
		fmt.Printf("Fix the configuration or remove provider settings to start in degraded mode\n")
	} else if defaultProvider == nil {
		// Degraded mode - no provider configured
		if !cfg.DegradedMode {
			cfg.DegradedMode = true
			fmt.Println("\n⚠ DEGRADED MODE: No LLM provider configured")
			fmt.Println("  • Agent orchestration disabled")
			fmt.Println("  • Web panel available at http://localhost:" + fmt.Sprintf("%d", cfg.Web.Port))
			fmt.Println("  → Visit web panel to configure your LLM provider")
		}
	} else {
		defaultAgentLoop = agent.NewAgentLoop(cfg, msgBus, defaultProvider)
	}

	agentManager := agent.NewAgentManager(defaultAgentLoop)
	if defaultAgentLoop != nil {
		if err := agentManager.InitializeOrchestrator(cfg, msgBus, channelStore); err != nil {
			logger.WarnCF("gateway", "Failed to initialize orchestrator", map[string]interface{}{
				"error": err.Error(),
			})
		}
		agentManager.InitializeSwarms(cfg)
	}

	heartbeatService := heartbeat.NewHeartbeatService(
		cfg.WorkspacePath(),
		nil,
		30*60,
		true,
	)

	// Transcriber setup (can be shared across all users)
	var transcriber *voice.GroqTranscriber
	if cfg.Providers.Groq.APIKey != "" {
		transcriber = voice.NewGroqTranscriber(cfg.Providers.Groq.APIKey)
		logger.InfoC("voice", "Groq voice transcription enabled")
	}

	// Note: transcriber attachment to channels is handled per-user in MultiUserChannelManager
	// TODO: Implement transcriber setup in per-user channel initialization

	fmt.Printf("✓ Gateway multi-user mode ready\n")
	fmt.Println("Press Ctrl+C to stop")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := heartbeatService.Start(); err != nil {
		fmt.Printf("Error starting heartbeat service: %v\n", err)
	}
	fmt.Println("✓ Heartbeat service started")

	// Start all user channel managers
	if err := multiChannelManager.StartAll(ctx); err != nil {
		fmt.Printf("Error starting channels: %v\n", err)
	}
	fmt.Println("✓ All user channels started")

	// Agent loops are started automatically by MultiUserChannelManager
	// No need to call agentLoop.Run() here - each user has their own

	// Setup MCP manager for configured servers (global for now)
	// TODO: Make MCP servers per-user in future phase
	var mcpManager *mcp.Manager
	if len(cfg.Tools.MCP.Servers) > 0 {
		mcpManager = mcp.NewManager(cfg.Tools.MCP)
		mcpManager.Start(ctx)
		mcpStatus := mcpManager.ServerStatus()
		connCount := 0
		for _, s := range mcpStatus {
			if s.Connected {
				connCount++
			}
		}
		fmt.Printf("✓ MCP servers: %d/%d connected\n", connCount, len(mcpStatus))
	}

	var webServer *web.Server
	if cfg.Web.Enabled {
		// Web panel always starts in gateway mode.
		// Each user has their own agent loop via MultiUserChannelManager;
		// the global defaultAgentLoop may be nil when no global provider is configured
		// (per-user degraded mode). The web server supports nil agentLoop gracefully.
		if defaultAgentLoop == nil {
			fmt.Println("⚠ No global LLM provider configured — web panel starting in per-user mode")
			fmt.Println("  Users with a provider configured in their own settings will have full agent access")
		}
		webServer = web.NewServerWithWorkspace(cfg.Web, defaultAgentLoop, cfg.WorkspacePath())
		webServer.SetBus(msgBus)

		// Initialize storage for tasks, chat history, etc.
		if channelStore != nil {
			webServer.SetStorage(channelStore)
		}
		// Wire central and per-user storage for multi-user isolation
		if centralStore != nil {
			webServer.SetCentralStorage(centralStore)
		}
		if userMgr != nil {
			webServer.SetUserStorageManager(userMgr)
		}

		// Wire additional services for advanced REST endpoints
		if agentManager != nil {
			webServer.SetAgentManager(agentManager)
		}
		webServer.SetMultiUserChannelManager(multiChannelManager)
		webServer.SetFullConfig(cfg)
		if transcriber != nil {
			webServer.SetTranscriber(transcriber)
		}
		if mcpManager != nil {
			webServer.SetMCPManager(mcpManager)
		}
		home, _ := os.UserHomeDir()
		builtinDir := resolveBuiltinSkillsDir()
		skillsLoader := skills.NewSkillsLoader(
			cfg.WorkspacePath(),
			filepath.Join(home, ".MakoClaw", "skills"),
			builtinDir,
		)
		skillInstaller := skills.NewSkillInstaller(cfg.WorkspacePath())
		webServer.SetBuiltinSkillsDir(builtinDir)
		webServer.SetSkills(skillsLoader, skillInstaller)
		// Wire workflow engine only when a global agent loop is available
		if channelStore != nil && defaultAgentLoop != nil {
			wfEngine := workflow.NewEngine(defaultAgentLoop, defaultAgentLoop.ToolRegistry(), channelStore)
			webServer.SetWorkflowEngine(wfEngine)
		}
		if err := webServer.Start(ctx); err != nil {
			fmt.Printf("Error starting web server: %v\n", err)
		} else {
			fmt.Printf("✓ Web panel started on %s:%d\n", cfg.Web.Host, cfg.Web.Port)
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	fmt.Println("\nShutting down...")
	cancel()
	heartbeatService.Stop()
	// Note: Per-user cron services are stopped by MultiUserChannelManager
	if mcpManager != nil {
		mcpManager.Stop()
	}
	if webServer != nil {
		_ = webServer.Stop(context.Background())
	}
	// Stop default agent loop
	if defaultAgentLoop != nil {
		defaultAgentLoop.Stop()
	}
	// Stop all user channel managers and agents
	multiChannelManager.StopAll(ctx)

	// Close storage after all services have stopped
	if channelStore != nil {
		_ = channelStore.Close()
	}
	if centralStore != nil {
		_ = centralStore.Close()
	}
	if userMgr != nil {
		userMgr.Close()
	}
	fmt.Println("✓ Gateway stopped")
}

func webCmd() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if !cfg.Web.Enabled {
		fmt.Println("Web is disabled. Set web.enabled=true in config.")
		os.Exit(1)
	}

	provider, err := providers.TryCreateProvider(cfg)
	if err != nil {
		// Provider configuration error - enter degraded mode instead of exiting
		fmt.Printf("Error with provider configuration: %v\n", err)
		fmt.Println("Starting in DEGRADED MODE. Configure provider via web panel to enable agent features.")
		provider = nil // Treat error as no provider (degraded mode)
	}

	msgBus := bus.NewMessageBus()
	var agentLoop *agent.AgentLoop

	// Setup context for all services (needed even in degraded mode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if provider == nil {
		// Degraded mode - no provider configured
		cfg.DegradedMode = true
		fmt.Println("\n⚠ DEGRADED MODE: No LLM provider configured")
		fmt.Println("  • Agent loop disabled")
		fmt.Println("  • Cron jobs will be handled per-user via multi-user manager")
		fmt.Println("  • Web panel available for configuration")
		fmt.Println("  → Visit http://localhost:" + fmt.Sprintf("%d", cfg.Web.Port) + " to configure your LLM provider")

		// Note: Cron services are now managed per-user by the web server
		// through the multiUserChannelManager. No global cron service needed.
	} else {
		// Full mode - provider configured
		agentLoop = agent.NewAgentLoop(cfg, msgBus, provider)

		// Note: Cron services are now managed per-user by the web server
		// through the multiUserChannelManager. Each user gets their own cron service.

		go agentLoop.Run(ctx)
	}

	webServer := web.NewServerWithWorkspace(cfg.Web, agentLoop, cfg.WorkspacePath())
	webServer.SetBus(msgBus)

	// Initialize storage for tasks
	store, err := storage.New(cfg.Storage)
	if err == nil {
		webServer.SetStorage(store)
		defer store.Close()
	} else {
		fmt.Printf("Warning: Failed to initialize storage: %v\n", err)
	}

	// Initialize central database (users, settings)
	dataDir := filepath.Dir(expandHomePath(cfg.Storage.Path))
	config.InitDataDir(dataDir)
	centralDBPath := filepath.Join(dataDir, "central.db")
	centralStore, csErr := storage.NewCentral(centralDBPath)
	if csErr == nil {
		webServer.SetCentralStorage(centralStore)
		defer centralStore.Close()
		fmt.Println("✓ Central database initialized")
	} else {
		fmt.Printf("Warning: Failed to initialize central storage: %v\n", csErr)
	}

	// Initialize per-user storage manager
	userMgr := storage.NewUserStorageManager(centralStore, dataDir)
	webServer.SetUserStorageManager(userMgr)
	defer userMgr.Close()
	fmt.Println("✓ Per-user storage manager ready")

	// Create multi-user channel manager for per-user cron service support
	// In web mode, we don't start channels, but we need the manager for cron services
	multiChannelManager := channels.NewMultiUserChannelManager(cfg, msgBus, store, centralStore)
	if centralStore != nil {
		if err := multiChannelManager.InitializeAllUsers(); err != nil {
			logger.WarnCF("web", "Failed to initialize user cron services", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}
	webServer.SetMultiUserChannelManager(multiChannelManager)
	if err := multiChannelManager.StartChannelsAndCron(ctx); err != nil {
		logger.WarnCF("web", "Failed to start channels and cron", map[string]interface{}{"error": err.Error()})
	}
	fmt.Println("✓ Channels and cron services started for all users")

	// Create and wire agent manager so specialist CRUD works in web mode
	// In degraded mode, agent manager can still be created but won't have active agents
	webAgentManager := agent.NewAgentManager(agentLoop)
	if agentLoop != nil {
		if store != nil {
			if initErr := webAgentManager.InitializeOrchestrator(cfg, msgBus, store); initErr != nil {
				fmt.Printf("Warning: Failed to initialize agent orchestrator: %v\n", initErr)
			}
		}
		webAgentManager.InitializeSwarms(cfg)
	}
	webServer.SetAgentManager(webAgentManager)

	// Wire additional services for advanced REST endpoints
	webServer.SetFullConfig(cfg)
	// Note: Cron service is managed per-user via multiUserChannelManager
	// Wire voice transcriber if Groq API key is available
	if cfg.Providers.Groq.APIKey != "" {
		webTranscriber := voice.NewGroqTranscriber(cfg.Providers.Groq.APIKey)
		webServer.SetTranscriber(webTranscriber)
	}
	// Wire TTS synthesizer if configured
	if cfg.TTS.Enabled && cfg.TTS.APIKey != "" {
		ttsSynth := voice.NewTTSSynthesizer(cfg.TTS)
		webServer.SetTTSSynthesizer(ttsSynth)
	}
	// Wire MCP manager if configured
	var mcpManagerWeb *mcp.Manager
	if len(cfg.Tools.MCP.Servers) > 0 {
		mcpManagerWeb = mcp.NewManager(cfg.Tools.MCP)
		mcpManagerWeb.Start(ctx)
		webServer.SetMCPManager(mcpManagerWeb)
	}
	homeWeb, _ := os.UserHomeDir()
	builtinDirWeb := resolveBuiltinSkillsDir()
	skillsLoaderWeb := skills.NewSkillsLoader(
		cfg.WorkspacePath(),
		filepath.Join(homeWeb, ".MakoClaw", "skills"),
		builtinDirWeb,
	)
	skillInstallerWeb := skills.NewSkillInstaller(cfg.WorkspacePath())
	webServer.SetBuiltinSkillsDir(builtinDirWeb)
	webServer.SetSkills(skillsLoaderWeb, skillInstallerWeb)
	// Wire workflow engine (only if agent loop is available)
	if store != nil && agentLoop != nil {
		wfEngine := workflow.NewEngine(agentLoop, agentLoop.ToolRegistry(), store)
		webServer.SetWorkflowEngine(wfEngine)
	}

	// Wire Channel Manager
	channelManager, err := channels.NewManager(cfg, msgBus, store)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize channel manager: %v\n", err)
	} else {
		// Wire transcriber to channels if available (using same logic as gatewayCmd)
		if cfg.Providers.Groq.APIKey != "" {
			transcriber := voice.NewGroqTranscriber(cfg.Providers.Groq.APIKey)
			if telegramChannel, ok := channelManager.GetChannel("telegram"); ok {
				if tc, ok := telegramChannel.(*channels.TelegramChannel); ok {
					tc.SetTranscriber(transcriber)
				}
			}
			if discordChannel, ok := channelManager.GetChannel("discord"); ok {
				if dc, ok := discordChannel.(*channels.DiscordChannel); ok {
					dc.SetTranscriber(transcriber)
				}
			}
			if slackChannel, ok := channelManager.GetChannel("slack"); ok {
				if sc, ok := slackChannel.(*channels.SlackChannel); ok {
					sc.SetTranscriber(transcriber)
				}
			}
		}

		webServer.SetChannelManager(channelManager)
		if err := channelManager.StartAll(ctx); err != nil {
			fmt.Printf("Error starting channels: %v\n", err)
		}
	}

	if err := webServer.Start(ctx); err != nil {
		fmt.Printf("Error starting web server: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Web panel started on %s:%d\n", cfg.Web.Host, cfg.Web.Port)
	fmt.Println("Press Ctrl+C to stop")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	fmt.Println("\nShutting down...")
	cancel()
	if mcpManagerWeb != nil {
		mcpManagerWeb.Stop()
	}
	_ = webServer.Stop(context.Background())
	// Note: Per-user cron services are stopped by multiUserChannelManager
	if agentLoop != nil {
		agentLoop.Stop()
	}
	fmt.Println("✓ Web stopped")
}

func statusCmd() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	configPath := getConfigPath()

	fmt.Printf("%s makoclaw Status\n\n", logo)

	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("Config:", configPath, "✓")
	} else {
		fmt.Println("Config:", configPath, "✗")
	}

	workspace := cfg.WorkspacePath()
	if _, err := os.Stat(workspace); err == nil {
		fmt.Println("Workspace:", workspace, "✓")
	} else {
		fmt.Println("Workspace:", workspace, "✗")
	}

	if _, err := os.Stat(configPath); err == nil {
		model := cfg.Agents.Defaults.Model
		fmt.Printf("Model: %s\n", model)

		// Issue #43: Show provider mapping
		providerName, actualModel := providers.GetProviderForModel(model)
		fmt.Printf("Provider: %s (model: %s)\n", providerName, actualModel)

		hasOpenRouter := cfg.Providers.OpenRouter.APIKey != ""
		hasAnthropic := cfg.Providers.Anthropic.APIKey != ""
		hasOpenAI := cfg.Providers.OpenAI.APIKey != ""
		hasGemini := cfg.Providers.Gemini.APIKey != ""
		hasZhipu := cfg.Providers.Zhipu.APIKey != ""
		hasGroq := cfg.Providers.Groq.APIKey != ""
		hasVLLM := cfg.Providers.VLLM.APIBase != ""

		status := func(enabled bool) string {
			if enabled {
				return "✓"
			}
			return "not set"
		}
		fmt.Println("OpenRouter API:", status(hasOpenRouter))
		fmt.Println("Anthropic API:", status(hasAnthropic))
		fmt.Println("OpenAI API:", status(hasOpenAI))
		fmt.Println("Gemini API:", status(hasGemini))
		fmt.Println("Zhipu API:", status(hasZhipu))
		fmt.Println("Groq API:", status(hasGroq))
		if hasVLLM {
			fmt.Printf("vLLM/Local: ✓ %s\n", cfg.Providers.VLLM.APIBase)
		} else {
			fmt.Println("vLLM/Local: not set")
		}

		store, _ := auth.LoadStore()
		if store != nil && len(store.Credentials) > 0 {
			fmt.Println("\nOAuth/Token Auth:")
			for provider, cred := range store.Credentials {
				status := "authenticated"
				if cred.IsExpired() {
					status = "expired"
				} else if cred.NeedsRefresh() {
					status = "needs refresh"
				}
				fmt.Printf("  %s (%s): %s\n", provider, cred.AuthMethod, status)
			}
		}
	}
}

func doctorCmd() {
	configPath := getConfigPath()

	fmt.Printf("%s makoclaw Doctor\n", logo)
	fmt.Println("==================")
	fmt.Println()

	results := doctor.RunChecks(configPath)
	doctor.PrintResults(results)

	if doctor.HasErrors(results) {
		os.Exit(1)
	}
}

func authCmd() {
	if len(os.Args) < 3 {
		authHelp()
		return
	}

	switch os.Args[2] {
	case "login":
		authLoginCmd()
	case "logout":
		authLogoutCmd()
	case "status":
		authStatusCmd()
	default:
		fmt.Printf("Unknown auth command: %s\n", os.Args[2])
		authHelp()
	}
}

func authHelp() {
	fmt.Println("\nAuth commands:")
	fmt.Println("  login       Login via OAuth or paste token")
	fmt.Println("  logout      Remove stored credentials")
	fmt.Println("  status      Show current auth status")
	fmt.Println()
	fmt.Println("Login options:")
	fmt.Println("  --provider <name>    Provider to login with (openai, anthropic)")
	fmt.Println("  --device-code        Use device code flow (for headless environments)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  makoclaw auth login --provider openai")
	fmt.Println("  makoclaw auth login --provider openai --device-code")
	fmt.Println("  makoclaw auth login --provider anthropic")
	fmt.Println("  makoclaw auth logout --provider openai")
	fmt.Println("  makoclaw auth status")
}

func authLoginCmd() {
	provider := ""
	useDeviceCode := false

	args := os.Args[3:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--provider", "-p":
			if i+1 < len(args) {
				provider = args[i+1]
				i++
			}
		case "--device-code":
			useDeviceCode = true
		}
	}

	if provider == "" {
		fmt.Println("Error: --provider is required")
		fmt.Println("Supported providers: openai, anthropic")
		return
	}

	switch provider {
	case "openai":
		authLoginOpenAI(useDeviceCode)
	case "anthropic":
		authLoginPasteToken(provider)
	default:
		fmt.Printf("Unsupported provider: %s\n", provider)
		fmt.Println("Supported providers: openai, anthropic")
	}
}

func authLoginOpenAI(useDeviceCode bool) {
	cfg := auth.OpenAIOAuthConfig()

	var cred *auth.AuthCredential
	var err error

	if useDeviceCode {
		cred, err = auth.LoginDeviceCode(cfg)
	} else {
		cred, err = auth.LoginBrowser(cfg)
	}

	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		os.Exit(1)
	}

	if err := auth.SetCredential("openai", cred); err != nil {
		fmt.Printf("Failed to save credentials: %v\n", err)
		os.Exit(1)
	}

	appCfg, err := loadConfig()
	if err == nil {
		appCfg.Providers.OpenAI.AuthMethod = "oauth"
		if err := config.SaveConfig(getConfigPath(), appCfg); err != nil {
			fmt.Printf("Warning: could not update config: %v\n", err)
		}
	}

	fmt.Println("Login successful!")
	if cred.AccountID != "" {
		fmt.Printf("Account: %s\n", cred.AccountID)
	}
}

func authLoginPasteToken(provider string) {
	cred, err := auth.LoginPasteToken(provider, os.Stdin)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		os.Exit(1)
	}

	if err := auth.SetCredential(provider, cred); err != nil {
		fmt.Printf("Failed to save credentials: %v\n", err)
		os.Exit(1)
	}

	appCfg, err := loadConfig()
	if err == nil {
		switch provider {
		case "anthropic":
			appCfg.Providers.Anthropic.AuthMethod = "token"
		case "openai":
			appCfg.Providers.OpenAI.AuthMethod = "token"
		}
		if err := config.SaveConfig(getConfigPath(), appCfg); err != nil {
			fmt.Printf("Warning: could not update config: %v\n", err)
		}
	}

	fmt.Printf("Token saved for %s!\n", provider)
}

func authLogoutCmd() {
	provider := ""

	args := os.Args[3:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--provider", "-p":
			if i+1 < len(args) {
				provider = args[i+1]
				i++
			}
		}
	}

	if provider != "" {
		if err := auth.DeleteCredential(provider); err != nil {
			fmt.Printf("Failed to remove credentials: %v\n", err)
			os.Exit(1)
		}

		appCfg, err := loadConfig()
		if err == nil {
			switch provider {
			case "openai":
				appCfg.Providers.OpenAI.AuthMethod = ""
			case "anthropic":
				appCfg.Providers.Anthropic.AuthMethod = ""
			}
			config.SaveConfig(getConfigPath(), appCfg)
		}

		fmt.Printf("Logged out from %s\n", provider)
	} else {
		if err := auth.DeleteAllCredentials(); err != nil {
			fmt.Printf("Failed to remove credentials: %v\n", err)
			os.Exit(1)
		}

		appCfg, err := loadConfig()
		if err == nil {
			appCfg.Providers.OpenAI.AuthMethod = ""
			appCfg.Providers.Anthropic.AuthMethod = ""
			config.SaveConfig(getConfigPath(), appCfg)
		}

		fmt.Println("Logged out from all providers")
	}
}

func authStatusCmd() {
	store, err := auth.LoadStore()
	if err != nil {
		fmt.Printf("Error loading auth store: %v\n", err)
		return
	}

	if len(store.Credentials) == 0 {
		fmt.Println("No authenticated providers.")
		fmt.Println("Run: makoclaw auth login --provider <name>")
		return
	}

	fmt.Println("\nAuthenticated Providers:")
	fmt.Println("------------------------")
	for provider, cred := range store.Credentials {
		status := "active"
		if cred.IsExpired() {
			status = "expired"
		} else if cred.NeedsRefresh() {
			status = "needs refresh"
		}

		fmt.Printf("  %s:\n", provider)
		fmt.Printf("    Method: %s\n", cred.AuthMethod)
		fmt.Printf("    Status: %s\n", status)
		if cred.AccountID != "" {
			fmt.Printf("    Account: %s\n", cred.AccountID)
		}
		if !cred.ExpiresAt.IsZero() {
			fmt.Printf("    Expires: %s\n", cred.ExpiresAt.Format("2006-01-02 15:04"))
		}
	}
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".MakoClaw", "config.json")
}

func setupCronTool(agentLoop *agent.AgentLoop, msgBus *bus.MessageBus, workspace string) *cron.CronService {
	cronStorePath := filepath.Join(workspace, "cron", "jobs.json")

	// Create cron service
	cronService := cron.NewCronService(cronStorePath, nil)

	// Create and register CronTool
	cronTool := tools.NewCronTool(cronService, agentLoop, msgBus)
	agentLoop.RegisterTool(cronTool)

	// Set the onJob handler
	cronService.SetOnJob(func(job *cron.CronJob) (string, error) {
		result := cronTool.ExecuteJob(context.Background(), job)
		return result, nil
	})

	return cronService
}

// expandHomePath expands a leading "~" in path to the user's home directory.
// This mirrors the unexported expandHome helpers in storage and config packages.
func expandHomePath(path string) string {
	if path == "" || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if len(path) > 1 && (path[1] == '/' || path[1] == '\\') {
		return filepath.Join(home, path[2:])
	}
	return home
}

func loadDotEnvIfExists(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}
		if _, alreadySet := os.LookupEnv(key); alreadySet {
			continue
		}
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func loadLocalDotEnvDefaults() error {
	configDir := filepath.Dir(getConfigPath())
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	paths := []string{
		filepath.Join(configDir, ".env"),
		filepath.Join(cwd, ".env"),
	}
	for _, path := range paths {
		if err := loadDotEnvIfExists(path); err != nil {
			return fmt.Errorf("loading %s: %w", path, err)
		}
	}
	return nil
}

func loadConfig() (*config.Config, error) {
	if err := loadLocalDotEnvDefaults(); err != nil {
		logger.WarnCF("config", "Could not load local .env defaults", map[string]interface{}{"error": err.Error()})
	}
	return config.LoadConfig(getConfigPath())
}

func cronCmd() {
	if len(os.Args) < 3 {
		cronHelp()
		return
	}

	subcommand := os.Args[2]

	// Load config to get workspace path
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	cronStorePath := filepath.Join(cfg.WorkspacePath(), "cron", "jobs.json")

	switch subcommand {
	case "list":
		cronListCmd(cronStorePath)
	case "add":
		cronAddCmd(cronStorePath)
	case "remove":
		if len(os.Args) < 4 {
			fmt.Println("Usage: makoclaw cron remove <job_id>")
			return
		}
		cronRemoveCmd(cronStorePath, os.Args[3])
	case "enable":
		cronEnableCmd(cronStorePath, false)
	case "disable":
		cronEnableCmd(cronStorePath, true)
	default:
		fmt.Printf("Unknown cron command: %s\n", subcommand)
		cronHelp()
	}
}

func cronHelp() {
	fmt.Println("\nCron commands:")
	fmt.Println("  list              List all scheduled jobs")
	fmt.Println("  add              Add a new scheduled job")
	fmt.Println("  remove <id>       Remove a job by ID")
	fmt.Println("  enable <id>      Enable a job")
	fmt.Println("  disable <id>     Disable a job")
	fmt.Println()
	fmt.Println("Add options:")
	fmt.Println("  -n, --name       Job name")
	fmt.Println("  -m, --message    Message for agent")
	fmt.Println("  -e, --every      Run every N seconds")
	fmt.Println("  -c, --cron       Cron expression (e.g. '0 9 * * *')")
	fmt.Println("  -d, --deliver     Deliver response to channel")
	fmt.Println("  --to             Recipient for delivery")
	fmt.Println("  --channel        Channel for delivery")
}

func cronListCmd(storePath string) {
	cs := cron.NewCronService(storePath, nil)
	jobs := cs.ListJobs(true) // Show all jobs, including disabled

	if len(jobs) == 0 {
		fmt.Println("No scheduled jobs.")
		return
	}

	fmt.Println("\nScheduled Jobs:")
	fmt.Println("----------------")
	for _, job := range jobs {
		var schedule string
		if job.Schedule.Kind == "every" && job.Schedule.EveryMS != nil {
			schedule = fmt.Sprintf("every %ds", *job.Schedule.EveryMS/1000)
		} else if job.Schedule.Kind == "cron" {
			schedule = job.Schedule.Expr
		} else {
			schedule = "one-time"
		}

		nextRun := "scheduled"
		if job.State.NextRunAtMS != nil {
			nextTime := time.UnixMilli(*job.State.NextRunAtMS)
			nextRun = nextTime.Format("2006-01-02 15:04")
		}

		status := "enabled"
		if !job.Enabled {
			status = "disabled"
		}

		fmt.Printf("  %s (%s)\n", job.Name, job.ID)
		fmt.Printf("    Schedule: %s\n", schedule)
		fmt.Printf("    Status: %s\n", status)
		fmt.Printf("    Next run: %s\n", nextRun)
	}
}

func cronAddCmd(storePath string) {
	name := ""
	message := ""
	var everySec *int64
	cronExpr := ""
	jobType := "task" // Default to agent processing
	channel := ""
	to := ""

	args := os.Args[3:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-n", "--name":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		case "-m", "--message":
			if i+1 < len(args) {
				message = args[i+1]
				i++
			}
		case "-e", "--every":
			if i+1 < len(args) {
				var sec int64
				fmt.Sscanf(args[i+1], "%d", &sec)
				everySec = &sec
				i++
			}
		case "-c", "--cron":
			if i+1 < len(args) {
				cronExpr = args[i+1]
				i++
			}
		case "-t", "--type":
			if i+1 < len(args) {
				jobType = args[i+1]
				i++
			}
		case "-d", "--deliver":
			// Deprecated: for backward compatibility
			jobType = "reminder"
		case "--to":
			if i+1 < len(args) {
				to = args[i+1]
				i++
			}
		case "--channel":
			if i+1 < len(args) {
				channel = args[i+1]
				i++
			}
		}
	}

	if name == "" {
		fmt.Println("Error: --name is required")
		return
	}

	if message == "" {
		fmt.Println("Error: --message is required")
		return
	}

	if everySec == nil && cronExpr == "" {
		fmt.Println("Error: Either --every or --cron must be specified")
		return
	}

	var schedule cron.CronSchedule
	if everySec != nil {
		everyMS := *everySec * 1000
		schedule = cron.CronSchedule{
			Kind:    "every",
			EveryMS: &everyMS,
		}
	} else {
		schedule = cron.CronSchedule{
			Kind: "cron",
			Expr: cronExpr,
		}
	}

	cs := cron.NewCronService(storePath, nil)
	job, err := cs.AddJob(1, name, schedule, message, jobType, channel, to, "")
	if err != nil {
		fmt.Printf("Error adding job: %v\n", err)
		return
	}

	fmt.Printf("✓ Added job '%s' (%s)\n", job.Name, job.ID)
}

func cronRemoveCmd(storePath, jobID string) {
	cs := cron.NewCronService(storePath, nil)
	if cs.RemoveJob(jobID) {
		fmt.Printf("✓ Removed job %s\n", jobID)
	} else {
		fmt.Printf("✗ Job %s not found\n", jobID)
	}
}

func cronEnableCmd(storePath string, disable bool) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: makoclaw cron enable/disable <job_id>")
		return
	}

	jobID := os.Args[3]
	cs := cron.NewCronService(storePath, nil)
	enabled := !disable

	job := cs.EnableJob(jobID, enabled)
	if job != nil {
		status := "enabled"
		if disable {
			status = "disabled"
		}
		fmt.Printf("✓ Job '%s' %s\n", job.Name, status)
	} else {
		fmt.Printf("✗ Job %s not found\n", jobID)
	}
}

func skillsCmd() {
	if len(os.Args) < 3 {
		skillsHelp()
		return
	}

	subcommand := os.Args[2]

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	workspace := cfg.WorkspacePath()
	installer := skills.NewSkillInstaller(workspace)
	// 获取全局配置目录和内置 skills 目录
	globalDir := filepath.Dir(getConfigPath())
	globalSkillsDir := filepath.Join(globalDir, "skills")
	builtinSkillsDir := resolveBuiltinSkillsDir()
	skillsLoader := skills.NewSkillsLoader(workspace, globalSkillsDir, builtinSkillsDir)

	switch subcommand {
	case "list":
		skillsListCmd(skillsLoader)
	case "install":
		skillsInstallCmd(installer)
	case "remove", "uninstall":
		if len(os.Args) < 4 {
			fmt.Println("Usage: makoclaw skills remove <skill-name>")
			return
		}
		skillsRemoveCmd(installer, os.Args[3])
	case "search":
		skillsSearchCmd(installer)
	case "show":
		if len(os.Args) < 4 {
			fmt.Println("Usage: makoclaw skills show <skill-name>")
			return
		}
		skillsShowCmd(skillsLoader, os.Args[3])
	default:
		fmt.Printf("Unknown skills command: %s\n", subcommand)
		skillsHelp()
	}
}

func skillsHelp() {
	fmt.Println("\nSkills commands:")
	fmt.Println("  list                    List installed skills")
	fmt.Println("  install <repo>          Install skill from GitHub")
	fmt.Println("  install-builtin          Install all builtin skills to workspace")
	fmt.Println("  list-builtin             List available builtin skills")
	fmt.Println("  remove <name>           Remove installed skill")
	fmt.Println("  search                  Search available skills")
	fmt.Println("  show <name>             Show skill details")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  makoclaw skills list")
	fmt.Println("  makoclaw skills install sipeed/makoclaw-skills/weather")
	fmt.Println("  makoclaw skills install-builtin")
	fmt.Println("  makoclaw skills list-builtin")
	fmt.Println("  makoclaw skills remove weather")
}

func skillsListCmd(loader *skills.SkillsLoader) {
	allSkills := loader.ListSkills()

	if len(allSkills) == 0 {
		fmt.Println("No skills installed.")
		return
	}

	fmt.Println("\nInstalled Skills:")
	fmt.Println("------------------")
	for _, skill := range allSkills {
		fmt.Printf("  ✓ %s (%s)\n", skill.Name, skill.Source)
		if skill.Description != "" {
			fmt.Printf("    %s\n", skill.Description)
		}
	}
}

func skillsInstallCmd(installer *skills.SkillInstaller) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: makoclaw skills install <github-repo>")
		fmt.Println("Example: makoclaw skills install sipeed/makoclaw-skills/weather")
		return
	}

	repo := os.Args[3]
	fmt.Printf("Installing skill from %s...\n", repo)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := installer.InstallFromGitHub(ctx, repo); err != nil {
		fmt.Printf("✗ Failed to install skill: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Skill '%s' installed successfully!\n", filepath.Base(repo))
}

func skillsRemoveCmd(installer *skills.SkillInstaller, skillName string) {
	fmt.Printf("Removing skill '%s'...\n", skillName)

	if err := installer.Uninstall(skillName); err != nil {
		fmt.Printf("✗ Failed to remove skill: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Skill '%s' removed successfully!\n", skillName)
}

func skillsInstallBuiltinCmd(workspace string) {
	builtinSkillsDir := resolveBuiltinSkillsDir()
	workspaceSkillsDir := filepath.Join(workspace, "skills")

	fmt.Printf("Copying builtin skills to workspace...\n")

	skillsToInstall := []string{
		"weather",
		"news",
		"stock",
		"calculator",
	}

	for _, skillName := range skillsToInstall {
		builtinPath := filepath.Join(builtinSkillsDir, skillName)
		workspacePath := filepath.Join(workspaceSkillsDir, skillName)

		if _, err := os.Stat(builtinPath); err != nil {
			fmt.Printf("⊘ Builtin skill '%s' not found: %v\n", skillName, err)
			continue
		}

		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			fmt.Printf("✗ Failed to create directory for %s: %v\n", skillName, err)
			continue
		}

		if err := copyDirectory(builtinPath, workspacePath); err != nil {
			fmt.Printf("✗ Failed to copy %s: %v\n", skillName, err)
		}
	}

	fmt.Println("\n✓ All builtin skills installed!")
	fmt.Println("Now you can use them in your workspace.")
}

func skillsListBuiltinCmd() {
	builtinSkillsDir := resolveBuiltinSkillsDir()

	fmt.Println("\nAvailable Builtin Skills:")
	fmt.Println("-----------------------")

	entries, err := os.ReadDir(builtinSkillsDir)
	if err != nil {
		fmt.Printf("Error reading builtin skills: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Println("No builtin skills available.")
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			skillName := entry.Name()
			skillFile := filepath.Join(builtinSkillsDir, skillName, "SKILL.md")

			description := "No description"
			if _, err := os.Stat(skillFile); err == nil {
				data, err := os.ReadFile(skillFile)
				if err == nil {
					content := string(data)
					if idx := strings.Index(content, "\n"); idx > 0 {
						firstLine := content[:idx]
						if strings.Contains(firstLine, "description:") {
							descLine := strings.Index(content[idx:], "\n")
							if descLine > 0 {
								description = strings.TrimSpace(content[idx+descLine : idx+descLine])
							}
						}
					}
				}
			}
			status := "✓"
			fmt.Printf("  %s  %s\n", status, entry.Name())
			if description != "" {
				fmt.Printf("     %s\n", description)
			}
		}
	}
}

func skillsSearchCmd(installer *skills.SkillInstaller) {
	fmt.Println("Searching for available skills...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	availableSkills, err := installer.ListAvailableSkills(ctx)
	if err != nil {
		fmt.Printf("✗ Failed to fetch skills list: %v\n", err)
		return
	}

	if len(availableSkills) == 0 {
		fmt.Println("No skills available.")
		return
	}

	fmt.Printf("\nAvailable Skills (%d):\n", len(availableSkills))
	fmt.Println("--------------------")
	for _, skill := range availableSkills {
		fmt.Printf("  📦 %s\n", skill.Name)
		fmt.Printf("     %s\n", skill.Description)
		fmt.Printf("     Repo: %s\n", skill.Repository)
		if skill.Author != "" {
			fmt.Printf("     Author: %s\n", skill.Author)
		}
		if len(skill.Tags) > 0 {
			fmt.Printf("     Tags: %v\n", skill.Tags)
		}
		fmt.Println()
	}
}

func skillsShowCmd(loader *skills.SkillsLoader, skillName string) {
	content, ok := loader.LoadSkill(skillName)
	if !ok {
		fmt.Printf("✗ Skill '%s' not found\n", skillName)
		return
	}

	fmt.Printf("\n📦 Skill: %s\n", skillName)
	fmt.Println("----------------------")
	fmt.Println(content)
}
func listUsersCmd() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	dataDir := filepath.Dir(expandHomePath(cfg.Storage.Path))
	centralDBPath := filepath.Join(dataDir, "central.db")
	centralStore, err := storage.NewCentral(centralDBPath)
	if err != nil {
		fmt.Printf("Error opening central database: %v\n", err)
		os.Exit(1)
	}
	defer centralStore.Close()

	users, err := centralStore.ListUsers()
	if err != nil {
		fmt.Printf("Error listing users: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%-5s | %-15s | %-36s | %-10s\n", "ID", "Username", "UUID", "Role")
	fmt.Println(strings.Repeat("-", 75))
	for _, u := range users {
		fmt.Printf("%-5d | %-15s | %-36s | %-10s\n", u.ID, u.Username, u.UUID, u.Role)
	}

	// Also check legacy if central is empty
	if len(users) == 0 {
		fmt.Println("\nChecking legacy database...")
		store, err := storage.New(cfg.Storage)
		if err == nil {
			defer store.Close()
			legacyUsers, err := store.ListUsers()
			if err == nil && len(legacyUsers) > 0 {
				fmt.Printf("Found %d users in legacy database that need migration.\n", len(legacyUsers))
				for _, u := range legacyUsers {
					fmt.Printf(" - %s (%s)\n", u.Username, u.Role)
				}
			}
		}
	}
}

func resetPasswordCmd() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: makoclaw reset-password <username> <new-password>")
		return
	}
	username := os.Args[2]
	newPassword := os.Args[3]

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	dataDir := filepath.Dir(expandHomePath(cfg.Storage.Path))
	centralDBPath := filepath.Join(dataDir, "central.db")
	centralStore, err := storage.NewCentral(centralDBPath)
	if err != nil {
		fmt.Printf("Error opening central database: %v\n", err)
		os.Exit(1)
	}
	defer centralStore.Close()

	user, err := centralStore.GetUserByUsername(username)
	if err != nil {
		fmt.Printf("User %s not found in central database.\n", username)
		os.Exit(1)
	}

	if err := centralStore.UpdateUserPassword(user.ID, newPassword); err != nil {
		fmt.Printf("Error resetting password: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Password reset successfully for user %s\n", username)
}

// resolveBuiltinSkillsDir finds the built-in skills directory using a fallback chain:
// 1. {cwd}/skills/ (development mode — running from project root)
// 2. {executable_dir}/../skills/ (running from build/ directory)
// 3. ~/.MakoClaw/builtin-skills/ (production installed location)
func resolveBuiltinSkillsDir() string {
	// 1. CWD/skills (development)
	wd, _ := os.Getwd()
	candidate := filepath.Join(wd, "skills")
	if dirHasSkills(candidate) {
		return candidate
	}

	// 2. Relative to executable (build/ directory)
	if exe, err := os.Executable(); err == nil {
		candidate = filepath.Join(filepath.Dir(exe), "..", "skills")
		if abs, err := filepath.Abs(candidate); err == nil {
			if dirHasSkills(abs) {
				return abs
			}
		}
	}

	// 3. Installed location (~/.MakoClaw/builtin-skills/)
	if home, err := os.UserHomeDir(); err == nil {
		candidate = filepath.Join(home, ".MakoClaw", "builtin-skills")
		if dirHasSkills(candidate) {
			return candidate
		}
	}

	// Fallback to CWD (may be empty but won't crash)
	return filepath.Join(wd, "skills")
}

// dirHasSkills returns true if the directory contains at least one skill
// (a subdirectory with a SKILL.md file, or a nested category with one).
func dirHasSkills(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Direct skill: dir/skillName/SKILL.md
		if _, err := os.Stat(filepath.Join(dir, e.Name(), "SKILL.md")); err == nil {
			return true
		}
		// Nested category: dir/category/skillName/SKILL.md
		subEntries, err := os.ReadDir(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		for _, sub := range subEntries {
			if sub.IsDir() {
				if _, err := os.Stat(filepath.Join(dir, e.Name(), sub.Name(), "SKILL.md")); err == nil {
					return true
				}
			}
		}
	}
	return false
}
