package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// SafeShellCommands is the default allowlist of read-only/safe commands for restricted users
var SafeShellCommands = []string{
	"ls", "dir", // List directory
	"cat", "type", // View file contents
	"head", "tail", // View partial file contents
	"grep", "findstr", // Search in files
	"find", "where", // Find files
	"pwd", "cd", // Working directory
	"echo",   // Print text
	"date",   // Current date/time
	"whoami", // Current user
	"which",  // Locate command
	"wc",     // Count lines/words
	"sort",   // Sort lines
	"uniq",   // Remove duplicates
	"diff",   // Compare files
	"tree",   // Directory tree
	"file",   // Identify file type
	"stat",   // File statistics
}

type ExecTool struct {
	workingDir          string
	timeout             time.Duration
	denyPatterns        []*regexp.Regexp
	allowPatterns       []*regexp.Regexp
	restrictToWorkspace bool
	developerMode       bool // Disables workspace restriction, extends output/timeout for dev workflows
	maxOutput           int  // Max output chars (default 10000, developer mode 50000)
}

func NewExecTool(workingDir string, restrict bool) *ExecTool {
	denyPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\brm\s+-[rf]{1,2}\b`),
		regexp.MustCompile(`\bdel\s+/[fq]\b`),
		regexp.MustCompile(`\brmdir\s+/s\b`),
		regexp.MustCompile(`\b(format|mkfs|diskpart)\b\s`), // Match disk wiping commands (must be followed by space/args)
		regexp.MustCompile(`\bdd\s+if=`),
		regexp.MustCompile(`>\s*/dev/sd[a-z]\b`), // Block writes to disk devices (but allow /dev/null)
		regexp.MustCompile(`\b(shutdown|reboot|poweroff)\b`),
		regexp.MustCompile(`:\(\)\s*\{.*\};\s*:`),
	}

	return &ExecTool{
		workingDir:          workingDir,
		timeout:             60 * time.Second,
		denyPatterns:        denyPatterns,
		allowPatterns:       nil,
		restrictToWorkspace: restrict,
		maxOutput:           10000,
	}
}

func (t *ExecTool) SetWorkspace(workspace string) {
	t.workingDir = workspace
}

func (t *ExecTool) Name() string {
	return "exec"
}

func (t *ExecTool) Description() string {
	return "Execute a shell command and return its output. Use with caution."
}

func (t *ExecTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to execute",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Optional working directory for the command",
			},
		},
		"required": []string{"command"},
	}
}

func (t *ExecTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	command, ok := args["command"].(string)
	if !ok {
		return "", fmt.Errorf("command is required")
	}

	cwd := t.workingDir
	if wd, ok := args["working_dir"].(string); ok && wd != "" {
		// Validate working_dir is within workspace when restricted
		if t.restrictToWorkspace && t.workingDir != "" {
			absWd, err := filepath.Abs(wd)
			if err == nil {
				absWorkspace, err2 := filepath.Abs(t.workingDir)
				if err2 == nil {
					rel, err3 := filepath.Rel(absWorkspace, absWd)
					if err3 != nil || strings.HasPrefix(rel, "..") {
						return "Error: working_dir is outside the workspace", nil
					}
				}
			}
		}
		cwd = wd
	}

	if cwd == "" {
		wd, err := os.Getwd()
		if err == nil {
			cwd = wd
		}
	}

	if guardError := t.guardCommand(command, cwd); guardError != "" {
		return fmt.Sprintf("Error: %s", guardError), nil
	}

	cmdCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(cmdCtx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(cmdCtx, "sh", "-c", command)
	}
	if cwd != "" {
		cmd.Dir = cwd
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nSTDERR:\n" + stderr.String()
	}

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return fmt.Sprintf("Error: Command timed out after %v", t.timeout), nil
		}
		output += fmt.Sprintf("\nExit code: %v", err)
	}

	if output == "" {
		output = "(no output)"
	}

	maxLen := t.maxOutput
	if maxLen <= 0 {
		maxLen = 10000
	}
	if len(output) > maxLen {
		output = output[:maxLen] + fmt.Sprintf("\n... (truncated, %d more chars)", len(output)-maxLen)
	}

	return output, nil
}

func (t *ExecTool) guardCommand(command, cwd string) string {
	cmd := strings.TrimSpace(command)
	lower := strings.ToLower(cmd)

	for _, pattern := range t.denyPatterns {
		if pattern.MatchString(lower) {
			return "Command blocked by safety guard (dangerous pattern detected)"
		}
	}

	if len(t.allowPatterns) > 0 {
		allowed := false
		for _, pattern := range t.allowPatterns {
			if pattern.MatchString(lower) {
				allowed = true
				break
			}
		}
		if !allowed {
			return "Command blocked by safety guard (not in allowlist)"
		}

		if strings.ContainsAny(cmd, "|;&`") || strings.Contains(cmd, "$(") {
			return "Command blocked by safety guard (chaining operators not allowed)"
		}
	}

	if t.restrictToWorkspace {
		if strings.Contains(cmd, "..\\") || strings.Contains(cmd, "../") {
			return "Command blocked by safety guard (path traversal detected)"
		}

		cwdPath, err := filepath.Abs(cwd)
		if err != nil {
			return "Command blocked by safety guard (failed to resolve workspace path)"
		}

		pathPattern := regexp.MustCompile(`[A-Za-z]:\\[^\\\"']+|/[^\s\"']+`)
		matches := pathPattern.FindAllString(cmd, -1)

		filtered := make([]string, 0, len(matches))
		for _, m := range matches {
			idx := strings.Index(cmd, m)
			if idx > 0 {
				prefix := cmd[:idx]
				lastScheme := strings.LastIndex(prefix, "://")
				if lastScheme != -1 && !strings.ContainsAny(prefix[lastScheme+3:], " \t\n\"'") {
					continue
				}
			}
			filtered = append(filtered, m)
		}
		matches = filtered

		for _, raw := range matches {
			p, err := filepath.Abs(raw)
			if err != nil {
				continue
			}

			rel, err := filepath.Rel(cwdPath, p)
			if err != nil {
				continue
			}

			if strings.HasPrefix(rel, "..") {
				return "Command blocked by safety guard (path outside working dir)"
			}
		}
	}

	return ""
}

func (t *ExecTool) SetTimeout(timeout time.Duration) {
	t.timeout = timeout
}

func (t *ExecTool) SetRestrictToWorkspace(restrict bool) {
	t.restrictToWorkspace = restrict
}

// SetDeveloperMode enables developer mode: extends timeout to 5 min,
// increases output limit to 50K chars, and disables workspace restriction.
func (t *ExecTool) SetDeveloperMode(enabled bool) {
	t.developerMode = enabled
	if enabled {
		t.timeout = 5 * time.Minute
		t.maxOutput = 50000
		t.restrictToWorkspace = false
	}
}

// SetAdminMode enables full unrestricted access for admin users:
// developer mode + clears the deny patterns blacklist so no commands are blocked.
func (t *ExecTool) SetAdminMode(enabled bool) {
	t.SetDeveloperMode(enabled)
	if enabled {
		t.denyPatterns = nil
	}
}

func (t *ExecTool) SetAllowPatterns(patterns []string) error {
	t.allowPatterns = make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("invalid allow pattern %q: %w", p, err)
		}
		t.allowPatterns = append(t.allowPatterns, re)
	}
	return nil
}

// SetSafeCommandsForUser configures the exec tool to only allow safe commands
// Uses the SafeShellCommands list by default, or a custom list if provided
func (t *ExecTool) SetSafeCommandsForUser(customCommands []string) error {
	commands := SafeShellCommands
	if len(customCommands) > 0 {
		commands = customCommands
	}

	// Build regex patterns that match command at start of line
	patterns := make([]string, 0, len(commands))
	for _, cmd := range commands {
		// Match command at start or after whitespace/pipe/semicolon
		patterns = append(patterns, fmt.Sprintf(`^\s*%s\b`, regexp.QuoteMeta(cmd)))
	}

	return t.SetAllowPatterns(patterns)
}
