package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DockerTool provides container management operations for development workflows.
// Semi-autonomous: local operations run freely, compose_up in production requires confirmation.
type DockerTool struct {
	workspace string
}

func NewDockerTool(workspace string) *DockerTool {
	return &DockerTool{workspace: workspace}
}

func (t *DockerTool) SetWorkspace(workspace string) {
	t.workspace = workspace
}

func (t *DockerTool) Name() string { return "docker" }

func (t *DockerTool) Description() string {
	return `Docker container management for development and deployment.
Actions:
  build          - Build a Docker image from Dockerfile
  run            - Run a container from an image
  stop           - Stop a running container
  rm             - Remove a stopped container
  logs           - View container logs
  exec           - Execute a command inside a running container
  ps             - List running containers
  images         - List local images
  compose_up     - Start services with docker compose
  compose_down   - Stop services with docker compose
  compose_logs   - View compose service logs
  compose_build  - Build compose services
  inspect        - Inspect a container or image
  prune          - Remove unused containers/images/volumes (requires confirmation)
  status         - Check Docker daemon status and version`
}

func (t *DockerTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type": "string",
				"enum": []string{
					"build", "run", "stop", "rm", "logs", "exec", "ps", "images",
					"compose_up", "compose_down", "compose_logs", "compose_build",
					"inspect", "prune", "status",
				},
				"description": "The Docker action to perform",
			},
			"args": map[string]interface{}{
				"type":        "string",
				"description": "Arguments (image name, container ID, service name, etc.)",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Working directory for Docker commands (defaults to workspace)",
			},
		},
		"required": []string{"action"},
	}
}

func (t *DockerTool) RequiresConfirmation(action string, args map[string]interface{}) (bool, string) {
	switch action {
	case "prune":
		return true, "Remove unused Docker resources (containers, images, volumes). Confirm?"
	case "rm":
		return true, "Remove Docker container. Confirm?"
	}
	return false, ""
}

func (t *DockerTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action is required")
	}

	// Check Docker availability
	if action != "status" {
		if _, err := exec.LookPath("docker"); err != nil {
			return "Docker is not installed or not in PATH. Install from https://docs.docker.com/get-docker/", nil
		}
	}

	cwd := t.workspace
	if wd, ok := args["working_dir"].(string); ok && wd != "" {
		cwd = wd
	}

	extraArgs, _ := args["args"].(string)

	switch action {
	case "status":
		return t.dockerStatus(ctx, cwd)
	case "build":
		return t.dockerBuild(ctx, cwd, extraArgs)
	case "run":
		return t.dockerRun(ctx, cwd, extraArgs)
	case "stop":
		return t.runDocker(ctx, cwd, "stop "+extraArgs, 30*time.Second)
	case "rm":
		return t.runDocker(ctx, cwd, "rm "+extraArgs, 15*time.Second)
	case "logs":
		return t.dockerLogs(ctx, cwd, extraArgs)
	case "exec":
		return t.dockerExec(ctx, cwd, extraArgs)
	case "ps":
		return t.runDocker(ctx, cwd, "ps -a --format 'table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Names}}\t{{.Ports}}'", 10*time.Second)
	case "images":
		return t.runDocker(ctx, cwd, "images --format 'table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}'", 10*time.Second)
	case "compose_up":
		return t.composeUp(ctx, cwd, extraArgs)
	case "compose_down":
		return t.runCompose(ctx, cwd, "down "+extraArgs, 30*time.Second)
	case "compose_logs":
		return t.composeLogs(ctx, cwd, extraArgs)
	case "compose_build":
		return t.runCompose(ctx, cwd, "build "+extraArgs, 10*time.Minute)
	case "inspect":
		if extraArgs == "" {
			return "", fmt.Errorf("inspect requires a container or image name in args")
		}
		return t.runDocker(ctx, cwd, "inspect "+extraArgs, 10*time.Second)
	case "prune":
		return t.runDocker(ctx, cwd, "system prune -f", 60*time.Second)
	default:
		return "", fmt.Errorf("unknown docker action: %s", action)
	}
}

func (t *DockerTool) dockerStatus(ctx context.Context, cwd string) (string, error) {
	if _, err := exec.LookPath("docker"); err != nil {
		return "Docker: Not installed", nil
	}

	version, _ := t.runDocker(ctx, cwd, "version --format '{{.Server.Version}}'", 5*time.Second)
	info, _ := t.runDocker(ctx, cwd, "info --format '{{.ContainersRunning}} running, {{.ContainersStopped}} stopped, {{.Images}} images'", 5*time.Second)

	composeStatus := "Not available"
	if _, err := t.runDocker(ctx, cwd, "compose version", 5*time.Second); err == nil {
		composeStatus = "Available"
	}

	return fmt.Sprintf("Docker: %s\nContainers: %s\nCompose: %s", strings.TrimSpace(version), strings.TrimSpace(info), composeStatus), nil
}

func (t *DockerTool) dockerBuild(ctx context.Context, cwd, args string) (string, error) {
	if args == "" {
		args = "-t app:latest ."
	}
	return t.runDocker(ctx, cwd, "build "+args, 10*time.Minute)
}

func (t *DockerTool) dockerRun(ctx context.Context, cwd, args string) (string, error) {
	if args == "" {
		return "", fmt.Errorf("run requires image name and optional flags in args (e.g., '-d -p 8080:80 nginx')")
	}
	return t.runDocker(ctx, cwd, "run "+args, 60*time.Second)
}

func (t *DockerTool) dockerLogs(ctx context.Context, cwd, args string) (string, error) {
	if args == "" {
		return "", fmt.Errorf("logs requires a container name/ID in args")
	}
	// Default to last 100 lines
	if !strings.Contains(args, "--tail") && !strings.Contains(args, "-n") {
		args = "--tail 100 " + args
	}
	return t.runDocker(ctx, cwd, "logs "+args, 15*time.Second)
}

func (t *DockerTool) dockerExec(ctx context.Context, cwd, args string) (string, error) {
	if args == "" {
		return "", fmt.Errorf("exec requires container name and command (e.g., 'mycontainer ls -la')")
	}
	return t.runDocker(ctx, cwd, "exec "+args, 60*time.Second)
}

func (t *DockerTool) composeUp(ctx context.Context, cwd, args string) (string, error) {
	cmd := "up -d"
	if args != "" {
		cmd = "up -d " + args
	}
	return t.runCompose(ctx, cwd, cmd, 5*time.Minute)
}

func (t *DockerTool) composeLogs(ctx context.Context, cwd, args string) (string, error) {
	cmd := "logs --tail 100"
	if args != "" {
		cmd = "logs --tail 100 " + args
	}
	return t.runCompose(ctx, cwd, cmd, 15*time.Second)
}

func (t *DockerTool) runDocker(ctx context.Context, cwd, dockerArgs string, timeout time.Duration) (string, error) {
	return t.runCmd(ctx, cwd, "docker "+dockerArgs, timeout)
}

func (t *DockerTool) runCompose(ctx context.Context, cwd, composeArgs string, timeout time.Duration) (string, error) {
	return t.runCmd(ctx, cwd, "docker compose "+composeArgs, timeout)
}

func (t *DockerTool) runCmd(ctx context.Context, cwd, command string, timeout time.Duration) (string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "sh", "-c", command)
	if cwd != "" {
		cmd.Dir = cwd
	}

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
		output += fmt.Sprintf("\nExit code: %v", err)
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

var (
	_ Tool            = (*DockerTool)(nil)
	_ WorkspaceTool   = (*DockerTool)(nil)
	_ ConfirmableTool = (*DockerTool)(nil)
)
