package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// GitTool provides native Git operations with structured output and safety controls.
// Semi-autonomous: local operations (commit, branch, checkout) proceed freely,
// remote operations (push, delete remote branch) require confirmation.
type GitTool struct {
	workspace string
}

func NewGitTool(workspace string) *GitTool {
	return &GitTool{workspace: workspace}
}

func (t *GitTool) SetWorkspace(workspace string) {
	t.workspace = workspace
}

func (t *GitTool) Name() string { return "git" }

func (t *GitTool) Description() string {
	return `Native Git version control operations with safety controls.
Actions: clone, init, status, add, commit, branch, checkout, merge, pull, push,
         log, diff, stash, tag, remote, rebase, cherry_pick, reset, create_pr.
Local operations (commit, branch, checkout, stash) run without confirmation.
Remote operations (push, force_push, delete remote branch) require confirmation.
Use action="status" to see current repo state before making changes.`
}

func (t *GitTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type": "string",
				"enum": []string{
					"clone", "init", "status", "add", "commit", "branch", "checkout",
					"merge", "pull", "push", "log", "diff", "stash", "tag", "remote",
					"rebase", "cherry_pick", "reset", "create_pr",
				},
				"description": "The git action to perform",
			},
			"args": map[string]interface{}{
				"type":        "string",
				"description": "Arguments for the git command (e.g., branch name, commit message, remote URL)",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Commit/tag message (for commit, tag actions)",
			},
			"files": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Files to add/stage (for add action). Use [\".\"] for all files.",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Working directory for git commands (defaults to workspace)",
			},
		},
		"required": []string{"action"},
	}
}

func (t *GitTool) RequiresConfirmation(action string, args map[string]interface{}) (bool, string) {
	switch action {
	case "push":
		argsStr, _ := args["args"].(string)
		if strings.Contains(argsStr, "--force") || strings.Contains(argsStr, "-f") {
			return true, "Force push will overwrite remote history. Confirm?"
		}
		return true, "Push commits to remote repository. Confirm?"
	case "reset":
		argsStr, _ := args["args"].(string)
		if strings.Contains(argsStr, "--hard") {
			return true, "Hard reset will discard all uncommitted changes. Confirm?"
		}
	case "branch":
		argsStr, _ := args["args"].(string)
		if strings.Contains(argsStr, "-D") || strings.Contains(argsStr, "--delete") {
			return true, "Delete branch. Confirm?"
		}
	case "rebase":
		return true, "Rebase will rewrite commit history. Confirm?"
	case "create_pr":
		return true, "Create a pull request on the remote. Confirm?"
	}
	return false, ""
}

func (t *GitTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action is required")
	}

	cwd := t.workspace
	if wd, ok := args["working_dir"].(string); ok && wd != "" {
		cwd = wd
	}
	if cwd == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("could not determine working directory: %w", err)
		}
		cwd = wd
	}

	switch action {
	case "clone":
		return t.gitClone(ctx, args, cwd)
	case "init":
		return t.runGit(ctx, cwd, "init")
	case "status":
		return t.gitStatus(ctx, cwd)
	case "add":
		return t.gitAdd(ctx, args, cwd)
	case "commit":
		return t.gitCommit(ctx, args, cwd)
	case "branch":
		return t.gitBranch(ctx, args, cwd)
	case "checkout":
		return t.gitCheckout(ctx, args, cwd)
	case "merge":
		return t.gitMerge(ctx, args, cwd)
	case "pull":
		return t.gitPull(ctx, args, cwd)
	case "push":
		return t.gitPush(ctx, args, cwd)
	case "log":
		return t.gitLog(ctx, args, cwd)
	case "diff":
		return t.gitDiff(ctx, args, cwd)
	case "stash":
		return t.gitStash(ctx, args, cwd)
	case "tag":
		return t.gitTag(ctx, args, cwd)
	case "remote":
		return t.gitRemote(ctx, args, cwd)
	case "rebase":
		return t.gitRebase(ctx, args, cwd)
	case "cherry_pick":
		return t.gitCherryPick(ctx, args, cwd)
	case "reset":
		return t.gitReset(ctx, args, cwd)
	case "create_pr":
		return t.gitCreatePR(ctx, args, cwd)
	default:
		return "", fmt.Errorf("unknown git action: %s", action)
	}
}

func (t *GitTool) gitClone(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	url, _ := args["args"].(string)
	if url == "" {
		return "", fmt.Errorf("clone requires a repository URL in args")
	}
	parts := strings.Fields(url)
	gitArgs := append([]string{"clone"}, parts...)
	return t.runGit(ctx, cwd, gitArgs...)
}

func (t *GitTool) gitStatus(ctx context.Context, cwd string) (string, error) {
	// Get both porcelain (parseable) and human-readable output
	porcelain, err := t.runGit(ctx, cwd, "status", "--porcelain")
	if err != nil {
		return "", err
	}
	branch, _ := t.runGit(ctx, cwd, "branch", "--show-current")
	branch = strings.TrimSpace(branch)

	if porcelain == "" || strings.TrimSpace(porcelain) == "" {
		return fmt.Sprintf("Branch: %s\nWorking tree clean", branch), nil
	}

	return fmt.Sprintf("Branch: %s\n\nChanges:\n%s", branch, porcelain), nil
}

func (t *GitTool) gitAdd(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	files, ok := args["files"].([]interface{})
	if ok && len(files) > 0 {
		gitArgs := []string{"add"}
		for _, f := range files {
			if s, ok := f.(string); ok {
				gitArgs = append(gitArgs, s)
			}
		}
		return t.runGit(ctx, cwd, gitArgs...)
	}
	// Fallback to args string
	argsStr, _ := args["args"].(string)
	if argsStr == "" {
		argsStr = "."
	}
	parts := append([]string{"add"}, strings.Fields(argsStr)...)
	return t.runGit(ctx, cwd, parts...)
}

func (t *GitTool) gitCommit(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	message, _ := args["message"].(string)
	if message == "" {
		return "", fmt.Errorf("commit requires a message")
	}
	argsStr, _ := args["args"].(string)
	gitArgs := []string{"commit", "-m", message}
	if argsStr != "" {
		gitArgs = append(gitArgs, strings.Fields(argsStr)...)
	}
	return t.runGit(ctx, cwd, gitArgs...)
}

func (t *GitTool) gitBranch(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)
	if argsStr == "" {
		return t.runGit(ctx, cwd, "branch", "-a")
	}
	parts := append([]string{"branch"}, strings.Fields(argsStr)...)
	return t.runGit(ctx, cwd, parts...)
}

func (t *GitTool) gitCheckout(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)
	if argsStr == "" {
		return "", fmt.Errorf("checkout requires a branch/commit name in args")
	}
	parts := append([]string{"checkout"}, strings.Fields(argsStr)...)
	return t.runGit(ctx, cwd, parts...)
}

func (t *GitTool) gitMerge(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)
	if argsStr == "" {
		return "", fmt.Errorf("merge requires a branch name in args")
	}
	parts := append([]string{"merge"}, strings.Fields(argsStr)...)
	return t.runGit(ctx, cwd, parts...)
}

func (t *GitTool) gitPull(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)
	gitArgs := []string{"pull"}
	if argsStr != "" {
		gitArgs = append(gitArgs, strings.Fields(argsStr)...)
	}
	return t.runGit(ctx, cwd, gitArgs...)
}

func (t *GitTool) gitPush(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)

	// Safety: block force push to main/master
	if strings.Contains(argsStr, "--force") || strings.Contains(argsStr, "-f") {
		branch, _ := t.runGit(ctx, cwd, "branch", "--show-current")
		branch = strings.TrimSpace(branch)
		if branch == "main" || branch == "master" {
			return "BLOCKED: Force push to main/master is not allowed. Use a feature branch.", nil
		}
	}

	gitArgs := []string{"push"}
	if argsStr != "" {
		gitArgs = append(gitArgs, strings.Fields(argsStr)...)
	}
	return t.runGitWithTimeout(ctx, cwd, 120*time.Second, gitArgs...)
}

func (t *GitTool) gitLog(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)
	gitArgs := []string{"log", "--oneline", "-20"}
	if argsStr != "" {
		gitArgs = []string{"log"}
		gitArgs = append(gitArgs, strings.Fields(argsStr)...)
	}
	return t.runGit(ctx, cwd, gitArgs...)
}

func (t *GitTool) gitDiff(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)
	gitArgs := []string{"diff"}
	if argsStr != "" {
		gitArgs = append(gitArgs, strings.Fields(argsStr)...)
	}
	output, err := t.runGit(ctx, cwd, gitArgs...)
	if err != nil {
		return "", err
	}
	// Truncate large diffs
	if len(output) > 20000 {
		output = output[:20000] + "\n... (diff truncated, showing first 20000 chars)"
	}
	return output, nil
}

func (t *GitTool) gitStash(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)
	gitArgs := []string{"stash"}
	if argsStr != "" {
		gitArgs = append(gitArgs, strings.Fields(argsStr)...)
	}
	return t.runGit(ctx, cwd, gitArgs...)
}

func (t *GitTool) gitTag(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)
	message, _ := args["message"].(string)
	if argsStr == "" {
		return t.runGit(ctx, cwd, "tag", "-l")
	}
	gitArgs := []string{"tag"}
	if message != "" {
		gitArgs = append(gitArgs, "-a", argsStr, "-m", message)
	} else {
		gitArgs = append(gitArgs, strings.Fields(argsStr)...)
	}
	return t.runGit(ctx, cwd, gitArgs...)
}

func (t *GitTool) gitRemote(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)
	if argsStr == "" {
		return t.runGit(ctx, cwd, "remote", "-v")
	}
	parts := append([]string{"remote"}, strings.Fields(argsStr)...)
	return t.runGit(ctx, cwd, parts...)
}

func (t *GitTool) gitRebase(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)
	if argsStr == "" {
		return "", fmt.Errorf("rebase requires a base branch in args")
	}
	// Block interactive rebase (requires TTY)
	if strings.Contains(argsStr, "-i") || strings.Contains(argsStr, "--interactive") {
		return "Interactive rebase is not supported (requires TTY). Use non-interactive rebase.", nil
	}
	parts := append([]string{"rebase"}, strings.Fields(argsStr)...)
	return t.runGit(ctx, cwd, parts...)
}

func (t *GitTool) gitCherryPick(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)
	if argsStr == "" {
		return "", fmt.Errorf("cherry_pick requires a commit hash in args")
	}
	parts := append([]string{"cherry-pick"}, strings.Fields(argsStr)...)
	return t.runGit(ctx, cwd, parts...)
}

func (t *GitTool) gitReset(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	argsStr, _ := args["args"].(string)
	if argsStr == "" {
		return t.runGit(ctx, cwd, "reset")
	}
	parts := append([]string{"reset"}, strings.Fields(argsStr)...)
	return t.runGit(ctx, cwd, parts...)
}

func (t *GitTool) gitCreatePR(ctx context.Context, args map[string]interface{}, cwd string) (string, error) {
	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		return "GitHub CLI (gh) is not installed. Install it with: brew install gh (macOS) or see https://cli.github.com", nil
	}

	message, _ := args["message"].(string)
	argsStr, _ := args["args"].(string)

	ghArgs := []string{"pr", "create"}
	if message != "" {
		// Split message: first line = title, rest = body
		lines := strings.SplitN(message, "\n", 2)
		ghArgs = append(ghArgs, "--title", lines[0])
		if len(lines) > 1 {
			ghArgs = append(ghArgs, "--body", lines[1])
		}
	}
	if argsStr != "" {
		ghArgs = append(ghArgs, strings.Fields(argsStr)...)
	}

	return t.runCmdWithTimeout(ctx, cwd, 60*time.Second, "gh", ghArgs...)
}

// runGit executes a git command with default timeout
func (t *GitTool) runGit(ctx context.Context, cwd string, gitArgs ...string) (string, error) {
	return t.runGitWithTimeout(ctx, cwd, 60*time.Second, gitArgs...)
}

// runGitWithTimeout executes a git command with a specified timeout
func (t *GitTool) runGitWithTimeout(ctx context.Context, cwd string, timeout time.Duration, gitArgs ...string) (string, error) {
	return t.runCmdWithTimeout(ctx, cwd, timeout, "git", gitArgs...)
}

// runCmdWithTimeout executes any command with a specified timeout
func (t *GitTool) runCmdWithTimeout(ctx context.Context, cwd string, timeout time.Duration, command string, cmdArgs ...string) (string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, command, cmdArgs...)
	if cwd != "" {
		cmd.Dir = cwd
	}

	// Sanitize dangerous env vars for git
	cmd.Env = filterGitEnv(os.Environ())

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return fmt.Sprintf("Error: Command timed out after %v", timeout), nil
		}
		if output == "" {
			output = err.Error()
		}
		return output, nil // Return output with error info, not a Go error
	}

	if output == "" {
		output = "(no output)"
	}

	maxLen := 30000
	if len(output) > maxLen {
		output = output[:maxLen] + fmt.Sprintf("\n... (truncated, %d more chars)", len(output)-maxLen)
	}

	return output, nil
}

// filterGitEnv removes potentially dangerous git environment variables
func filterGitEnv(env []string) []string {
	// Block env vars that could execute arbitrary code via git hooks
	dangerousGitVars := regexp.MustCompile(`^(GIT_EXEC_PATH|GIT_TEMPLATE_DIR|GIT_EXTERNAL_DIFF)=`)

	filtered := make([]string, 0, len(env))
	for _, e := range env {
		if !dangerousGitVars.MatchString(e) {
			filtered = append(filtered, e)
		}
	}

	// Ensure git doesn't open editors (non-interactive)
	filtered = append(filtered, "GIT_EDITOR=true")
	filtered = append(filtered, "GIT_TERMINAL_PROMPT=0")

	// Set platform-specific merge strategy
	if runtime.GOOS != "windows" {
		filtered = append(filtered, "GIT_MERGE_AUTOEDIT=no")
	}

	return filtered
}

// Ensure GitTool implements required interfaces
var (
	_ Tool           = (*GitTool)(nil)
	_ WorkspaceTool  = (*GitTool)(nil)
	_ ConfirmableTool = (*GitTool)(nil)
)
