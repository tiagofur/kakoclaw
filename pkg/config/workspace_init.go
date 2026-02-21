package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureUserWorkspace creates the user-specific workspace directory structure and bootstrap files.
func EnsureUserWorkspace(userUUID string) (string, error) {
	if userUUID == "" {
		return "", fmt.Errorf("user UUID is required")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	userRoot := filepath.Join(home, ".kakoclaw", "users", userUUID)
	workspace := filepath.Join(userRoot, "workspace")

	// Core directories
	coreDirs := []string{
		workspace,
		filepath.Join(workspace, "memory"),
		filepath.Join(workspace, "sessions"),
		filepath.Join(workspace, "skills"),
		filepath.Join(workspace, "cron"),
		filepath.Join(workspace, "tasks"),
		filepath.Join(workspace, "temp"),
	}
	for _, dir := range coreDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}

	// Bootstrap templates
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

I am KakoClaw, a lightweight AI assistant powered by AI.

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
KakoClaw (The Apex AI Agent)

## Description
The ultimate evolution of the PicoClaw lineage. KakoClaw is an ultra-efficient, Go-native personal AI assistant designed for the most demanding efficiency requirements.

## Version
0.1.0

## Purpose
- Deliver supreme AI intelligence with sub-10MB RAM footprint.
- Empower low-cost hardware (10-dollar boards) with high-tier agent capabilities.
- Self-bootstrapped and AI-refined architectural perfection.

## Capabilities
- Instant startup (less than 1s)
- Multi-channel global communication
- Advanced tool orchestration (Web, Files, Shell, MCP)
- Precision task scheduling and cron management
- Long-term memory and context retention
- Native voice transcription

## Philosophy
- Apex Efficiency: Every bit is optimized for maximum impact.
- Transparency: Clear, auditable, and user-centric operations.
- Independence: Native implementation with zero heavy dependencies.

## License
MIT License - Free, Open, and Unstoppable.

## Heritage
Proudly inspired by and evolved from PicoClaw.

## Repository
https://github.com/sipeed/kakoclaw
`,
	}

	for filename, content := range templates {
		filePath := filepath.Join(workspace, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return "", err
			}
		}
	}

	memoryFile := filepath.Join(workspace, "memory", "MEMORY.md")
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
		if err := os.WriteFile(memoryFile, []byte(memoryContent), 0644); err != nil {
			return "", err
		}
	}

	return workspace, nil
}
