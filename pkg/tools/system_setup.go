package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
)

const (
	setupDefaultTimeout = 5 * time.Minute
	setupMaxTimeout     = 10 * time.Minute
	setupMaxOutput      = 10000
)

// setupDenyPatterns blocks dangerous commands even for admins.
var setupDenyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\brm\s+-[rf]{1,2}\s+/`),
	regexp.MustCompile(`(?i)\bcurl\b.*\|\s*(sh|bash)\b`),
	regexp.MustCompile(`(?i)\bwget\b.*\|\s*(sh|bash)\b`),
	regexp.MustCompile(`(?i)\bdd\s+if=`),
	regexp.MustCompile(`(?i)\b(shutdown|reboot|poweroff)\b`),
	regexp.MustCompile(`(?i)\b(format|mkfs|diskpart)\b`),
	regexp.MustCompile(`:\(\)\s*\{.*\};\s*:`),
	regexp.MustCompile(`>\s*/dev/sd[a-z]`),
	regexp.MustCompile(`(?i)\bchmod\s+777\s+/`),
	regexp.MustCompile(`(?i)\bchown\b.*\s+/[a-z]`),
}

type SystemSetupTool struct {
	storage *storage.Storage
	userID  int64
}

func NewSystemSetupTool(s *storage.Storage) *SystemSetupTool {
	return &SystemSetupTool{storage: s}
}

func (t *SystemSetupTool) Name() string { return "system_setup" }

func (t *SystemSetupTool) Description() string {
	return `Install, manage, and track system packages and CLI tools in the runtime environment.
IMPORTANT: Before installing anything, ALWAYS confirm with the user first by listing what you plan to install and why.
Actions: install, uninstall, list, status, run_setup_script.
Supports: apt-get, pip, pip3, npm, go install, cargo, and custom setup scripts.
This tool has an extended timeout (up to 10 minutes) for long installations. Admin only.`
}

func (t *SystemSetupTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"install", "uninstall", "list", "status", "run_setup_script"},
				"description": "The action to perform",
			},
			"package_name": map[string]interface{}{
				"type":        "string",
				"description": "Package name (required for install/uninstall/status)",
			},
			"package_manager": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"apt-get", "pip", "pip3", "npm", "go", "cargo", "auto"},
				"description": "Package manager to use. 'auto' attempts detection. Default: auto",
			},
			"version": map[string]interface{}{
				"type":        "string",
				"description": "Optional version constraint (e.g. '>=1.0', '1.2.3')",
			},
			"extra_args": map[string]interface{}{
				"type":        "string",
				"description": "Additional arguments to pass to the package manager",
			},
			"script": map[string]interface{}{
				"type":        "string",
				"description": "Setup script content (for run_setup_script action only)",
			},
			"timeout_seconds": map[string]interface{}{
				"type":        "number",
				"description": "Timeout in seconds (max 600, default 300)",
			},
			"confirmed": map[string]interface{}{
				"type":        "boolean",
				"description": "Must be true to actually execute. First call without this to get a confirmation prompt.",
			},
			"status_filter": map[string]interface{}{
				"type":        "string",
				"description": "Filter for list action: installed, failed, all. Default: installed",
			},
		},
		"required": []string{"action"},
	}
}

func (t *SystemSetupTool) SetUserID(userID int64) {
	if userID > 0 {
		t.userID = userID
	}
}

func (t *SystemSetupTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	if action == "" {
		return "", fmt.Errorf("action is required")
	}

	switch action {
	case "install":
		return t.executeInstall(ctx, args)
	case "uninstall":
		return t.executeUninstall(ctx, args)
	case "list":
		return t.executeList(args)
	case "status":
		return t.executeStatus(ctx, args)
	case "run_setup_script":
		return t.executeScript(ctx, args)
	default:
		return "", fmt.Errorf("unknown action: %s (valid: install, uninstall, list, status, run_setup_script)", action)
	}
}

func (t *SystemSetupTool) executeInstall(ctx context.Context, args map[string]interface{}) (string, error) {
	pkgName, _ := args["package_name"].(string)
	if pkgName == "" {
		return "", fmt.Errorf("package_name is required for install")
	}

	pm := resolvePackageManager(args)
	version, _ := args["version"].(string)
	extraArgs, _ := args["extra_args"].(string)

	cmd := buildInstallCommand(pm, pkgName, version, extraArgs)

	if !isConfirmed(args) {
		return ConfirmationMessage("system_setup", fmt.Sprintf("will run: %s", cmd)), nil
	}

	// Check if already installed
	if existing, _ := t.storage.GetPackageByNameAndPM(pkgName, pm); existing != nil && existing.Status == "installed" {
		return fmt.Sprintf("Package '%s' is already installed (version: %s, installed at: %s). Use 'status' action for live verification.",
			pkgName, existing.Version, existing.CreatedAt.Format(time.RFC3339)), nil
	}

	if err := guardSetupCommand(cmd); err != "" {
		return fmt.Sprintf("Error: blocked dangerous command: %s", err), nil
	}

	uninstallCmd := buildUninstallCommand(pm, pkgName)
	pkgRecord := storage.InstalledPackage{
		Name:             pkgName,
		PackageManager:   pm,
		Status:           "installing",
		InstalledBy:      t.userID,
		InstallCommand:   cmd,
		UninstallCommand: uninstallCmd,
	}
	id, recordErr := t.storage.RecordPackage(pkgRecord)
	if recordErr != nil {
		logger.WarnCF("tool", "Failed to record package pre-install", map[string]interface{}{"error": recordErr.Error()})
	}

	timeout := resolveTimeout(args)
	output, execErr := runCommand(ctx, cmd, timeout)

	if execErr != "" {
		if id > 0 {
			_ = t.storage.UpdatePackageStatus(id, "failed", output, execErr, "")
		}
		return fmt.Sprintf("Installation failed:\n%s\n\nError: %s", truncateOutput(output), execErr), nil
	}

	detectedVersion := detectInstalledVersion(pm, pkgName)
	if id > 0 {
		_ = t.storage.UpdatePackageStatus(id, "installed", truncateOutput(output), "", detectedVersion)
	}

	result := fmt.Sprintf("Successfully installed '%s' via %s", pkgName, pm)
	if detectedVersion != "" {
		result += fmt.Sprintf(" (version: %s)", detectedVersion)
	}
	if output != "" {
		result += fmt.Sprintf("\n\nOutput:\n%s", truncateOutput(output))
	}
	return result, nil
}

func (t *SystemSetupTool) executeUninstall(ctx context.Context, args map[string]interface{}) (string, error) {
	pkgName, _ := args["package_name"].(string)
	if pkgName == "" {
		return "", fmt.Errorf("package_name is required for uninstall")
	}

	pm := resolvePackageManager(args)
	cmd := buildUninstallCommand(pm, pkgName)

	if !isConfirmed(args) {
		return ConfirmationMessage("system_setup", fmt.Sprintf("will run: %s", cmd)), nil
	}

	if err := guardSetupCommand(cmd); err != "" {
		return fmt.Sprintf("Error: blocked dangerous command: %s", err), nil
	}

	timeout := resolveTimeout(args)
	output, execErr := runCommand(ctx, cmd, timeout)

	if execErr != "" {
		return fmt.Sprintf("Uninstall failed:\n%s\n\nError: %s", truncateOutput(output), execErr), nil
	}

	// Update tracking record
	if existing, _ := t.storage.GetPackageByNameAndPM(pkgName, pm); existing != nil {
		_ = t.storage.UpdatePackageStatus(existing.ID, "uninstalled", truncateOutput(output), "", existing.Version)
	}

	return fmt.Sprintf("Successfully uninstalled '%s' via %s\n\nOutput:\n%s", pkgName, pm, truncateOutput(output)), nil
}

func (t *SystemSetupTool) executeList(args map[string]interface{}) (string, error) {
	filter, _ := args["status_filter"].(string)
	if filter == "" {
		filter = "installed"
	}

	packages, err := t.storage.ListPackages(filter)
	if err != nil {
		return "", fmt.Errorf("listing packages: %w", err)
	}

	if len(packages) == 0 {
		return fmt.Sprintf("No packages found with status filter '%s'.", filter), nil
	}

	return storage.PackagesToJSON(packages), nil
}

func (t *SystemSetupTool) executeStatus(ctx context.Context, args map[string]interface{}) (string, error) {
	pkgName, _ := args["package_name"].(string)
	if pkgName == "" {
		return "", fmt.Errorf("package_name is required for status")
	}

	pm := resolvePackageManager(args)

	// DB record check
	record, _ := t.storage.GetPackageByNameAndPM(pkgName, pm)

	// Live system check
	liveStatus := checkLiveStatus(ctx, pm, pkgName)

	result := map[string]interface{}{
		"package":         pkgName,
		"package_manager": pm,
		"live_check":      liveStatus,
	}
	if record != nil {
		result["tracked"] = true
		result["tracked_status"] = record.Status
		result["tracked_version"] = record.Version
		result["installed_at"] = record.CreatedAt.Format(time.RFC3339)
		result["install_command"] = record.InstallCommand
	} else {
		result["tracked"] = false
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return string(data), nil
}

func (t *SystemSetupTool) executeScript(ctx context.Context, args map[string]interface{}) (string, error) {
	script, _ := args["script"].(string)
	if script == "" {
		return "", fmt.Errorf("script is required for run_setup_script")
	}

	if !isConfirmed(args) {
		preview := script
		if len(preview) > 500 {
			preview = preview[:500] + "\n... (truncated)"
		}
		return ConfirmationMessage("system_setup", fmt.Sprintf("will execute setup script:\n---\n%s\n---", preview)), nil
	}

	if err := guardSetupCommand(script); err != "" {
		return fmt.Sprintf("Error: blocked dangerous command in script: %s", err), nil
	}

	// Write script to temp file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("makoclaw_setup_%d.sh", time.Now().UnixNano()))
	if writeErr := os.WriteFile(tmpFile, []byte("#!/bin/sh\nset -e\n"+script), 0700); writeErr != nil {
		return "", fmt.Errorf("writing temp script: %w", writeErr)
	}
	defer os.Remove(tmpFile)

	timeout := resolveTimeout(args)
	output, execErr := runCommand(ctx, "sh "+tmpFile, timeout)

	if execErr != "" {
		return fmt.Sprintf("Setup script failed:\n%s\n\nError: %s", truncateOutput(output), execErr), nil
	}

	return fmt.Sprintf("Setup script completed successfully.\n\nOutput:\n%s", truncateOutput(output)), nil
}

// --- Helper functions ---

func resolvePackageManager(args map[string]interface{}) string {
	pm, _ := args["package_manager"].(string)
	if pm == "" || pm == "auto" {
		pkgName, _ := args["package_name"].(string)
		return detectPackageManager(pkgName)
	}
	return pm
}

func detectPackageManager(pkgName string) string {
	if strings.Contains(pkgName, "github.com/") || strings.Contains(pkgName, "golang.org/") {
		return "go"
	}
	if strings.Contains(pkgName, "@") && strings.Contains(pkgName, "/") {
		return "npm"
	}
	if strings.HasPrefix(pkgName, "python-") || strings.HasPrefix(pkgName, "py-") {
		return "pip3"
	}
	return "apt-get"
}

func buildInstallCommand(pm, name, version, extraArgs string) string {
	var cmd string
	switch pm {
	case "apt-get":
		pkg := name
		if version != "" {
			pkg = name + "=" + version
		}
		cmd = fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get update -qq && apt-get install -y %s", pkg)
	case "pip", "pip3":
		pkg := name
		if version != "" {
			pkg = name + "==" + version
		}
		cmd = fmt.Sprintf("%s install %s", pm, pkg)
	case "npm":
		pkg := name
		if version != "" {
			pkg = name + "@" + version
		}
		cmd = fmt.Sprintf("npm install -g %s", pkg)
	case "go":
		ver := "latest"
		if version != "" {
			ver = version
		}
		cmd = fmt.Sprintf("go install %s@%s", name, ver)
	case "cargo":
		cmd = fmt.Sprintf("cargo install %s", name)
		if version != "" {
			cmd += " --version " + version
		}
	default:
		cmd = fmt.Sprintf("%s install %s", pm, name)
	}
	if extraArgs != "" {
		cmd += " " + extraArgs
	}
	return cmd
}

func buildUninstallCommand(pm, name string) string {
	switch pm {
	case "apt-get":
		return fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get remove -y %s", name)
	case "pip", "pip3":
		return fmt.Sprintf("%s uninstall -y %s", pm, name)
	case "npm":
		return fmt.Sprintf("npm uninstall -g %s", name)
	case "cargo":
		return fmt.Sprintf("cargo uninstall %s", name)
	default:
		return fmt.Sprintf("%s remove %s", pm, name)
	}
}

func guardSetupCommand(command string) string {
	for _, pattern := range setupDenyPatterns {
		if pattern.MatchString(command) {
			return fmt.Sprintf("matches blocked pattern: %s", pattern.String())
		}
	}
	return ""
}

func resolveTimeout(args map[string]interface{}) time.Duration {
	if ts, ok := args["timeout_seconds"].(float64); ok && ts > 0 {
		d := time.Duration(ts) * time.Second
		if d > setupMaxTimeout {
			d = setupMaxTimeout
		}
		return d
	}
	return setupDefaultTimeout
}

func runCommand(ctx context.Context, command string, timeout time.Duration) (output string, errMsg string) {
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "sh", "-c", command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output = stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return output, fmt.Sprintf("command timed out after %v", timeout)
		}
		return output, fmt.Sprintf("exit error: %v", err)
	}
	return output, ""
}

func truncateOutput(output string) string {
	if len(output) > setupMaxOutput {
		return output[:setupMaxOutput] + fmt.Sprintf("\n... (truncated, %d more chars)", len(output)-setupMaxOutput)
	}
	return output
}

func detectInstalledVersion(pm, name string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	switch pm {
	case "apt-get":
		cmd = exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Version}", name)
	case "pip", "pip3":
		cmd = exec.CommandContext(ctx, pm, "show", name)
	case "npm":
		cmd = exec.CommandContext(ctx, "npm", "list", "-g", name, "--depth=0")
	default:
		return ""
	}

	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	result := strings.TrimSpace(string(out))
	if pm == "pip" || pm == "pip3" {
		for _, line := range strings.Split(result, "\n") {
			if strings.HasPrefix(line, "Version:") {
				return strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
			}
		}
	}
	return result
}

func checkLiveStatus(ctx context.Context, pm, name string) string {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// First try 'which' for a quick binary check
	whichCmd := exec.CommandContext(checkCtx, "which", name)
	if out, err := whichCmd.Output(); err == nil {
		return fmt.Sprintf("found at %s", strings.TrimSpace(string(out)))
	}

	// PM-specific check
	var cmd *exec.Cmd
	switch pm {
	case "apt-get":
		cmd = exec.CommandContext(checkCtx, "dpkg", "-s", name)
	case "pip", "pip3":
		cmd = exec.CommandContext(checkCtx, pm, "show", name)
	case "npm":
		cmd = exec.CommandContext(checkCtx, "npm", "list", "-g", name, "--depth=0")
	default:
		return "not found (no live check available for this package manager)"
	}

	if out, err := cmd.Output(); err == nil {
		return fmt.Sprintf("installed (%s)", strings.Split(strings.TrimSpace(string(out)), "\n")[0])
	}
	return "not found"
}
