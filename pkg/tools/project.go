package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ProjectTool provides project management operations: detect language/framework,
// run builds, tests, linting, and dependency management across multiple languages.
type ProjectTool struct {
	workspace string
}

func NewProjectTool(workspace string) *ProjectTool {
	return &ProjectTool{workspace: workspace}
}

func (t *ProjectTool) SetWorkspace(workspace string) {
	t.workspace = workspace
}

func (t *ProjectTool) Name() string { return "project" }

func (t *ProjectTool) Description() string {
	return `Project management tool for multi-language development.
Actions:
  detect    - Auto-detect project type, language, package manager, and available commands
  deps      - Install or update dependencies (npm install, go mod tidy, pip install, etc.)
  build     - Build the project using the detected or specified build system
  test      - Run tests with the appropriate test runner
  lint      - Run linter/formatter for the project language
  format    - Auto-format code
  run       - Run the project (dev server, main binary, etc.)
  script    - Run a named script (e.g., npm run dev, make deploy)
  clean     - Clean build artifacts
Supports: Go, Node.js/TypeScript, Python, Rust, Swift/iOS, Kotlin/Android, Ruby, Java/Gradle/Maven.
Use detect first to understand the project before running other actions.`
}

func (t *ProjectTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"detect", "deps", "build", "test", "lint", "format", "run", "script", "clean"},
				"description": "The project action to perform",
			},
			"args": map[string]interface{}{
				"type":        "string",
				"description": "Additional arguments (e.g., script name for 'script', test pattern for 'test')",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Project directory (defaults to workspace)",
			},
		},
		"required": []string{"action"},
	}
}

// projectInfo holds detected project metadata
type projectInfo struct {
	Language    string
	Framework   string
	PackageFile string
	PackageMgr  string
	BuildCmd    string
	TestCmd     string
	LintCmd     string
	FormatCmd   string
	RunCmd      string
	CleanCmd    string
	DepsCmd     string
}

func (t *ProjectTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
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

	info := t.detectProject(cwd)

	switch action {
	case "detect":
		return t.formatDetection(info, cwd), nil
	case "deps":
		return t.runProjectCmd(ctx, cwd, info.DepsCmd, args)
	case "build":
		return t.runProjectCmd(ctx, cwd, info.BuildCmd, args)
	case "test":
		return t.runProjectCmd(ctx, cwd, info.TestCmd, args)
	case "lint":
		return t.runProjectCmd(ctx, cwd, info.LintCmd, args)
	case "format":
		return t.runProjectCmd(ctx, cwd, info.FormatCmd, args)
	case "run":
		return t.runProjectCmd(ctx, cwd, info.RunCmd, args)
	case "script":
		return t.runScript(ctx, cwd, info, args)
	case "clean":
		return t.runProjectCmd(ctx, cwd, info.CleanCmd, args)
	default:
		return "", fmt.Errorf("unknown project action: %s", action)
	}
}

func (t *ProjectTool) detectProject(dir string) projectInfo {
	info := projectInfo{}

	// Check files in priority order
	checks := []struct {
		file string
		fn   func(*projectInfo, string)
	}{
		{"go.mod", func(i *projectInfo, d string) {
			i.Language = "Go"
			i.PackageFile = "go.mod"
			i.PackageMgr = "go"
			i.DepsCmd = "go mod tidy"
			i.BuildCmd = "go build ./..."
			i.TestCmd = "go test ./..."
			i.LintCmd = "go vet ./..."
			i.FormatCmd = "go fmt ./..."
			i.RunCmd = "go run ."
			i.CleanCmd = "go clean -cache"
			// Check for Makefile
			if fileExists(filepath.Join(d, "Makefile")) {
				i.BuildCmd = "make build"
				i.TestCmd = "make test"
			}
		}},
		{"package.json", func(i *projectInfo, d string) {
			i.Language = "JavaScript/TypeScript"
			i.PackageFile = "package.json"
			// Detect package manager
			if fileExists(filepath.Join(d, "pnpm-lock.yaml")) {
				i.PackageMgr = "pnpm"
				i.DepsCmd = "pnpm install"
			} else if fileExists(filepath.Join(d, "yarn.lock")) {
				i.PackageMgr = "yarn"
				i.DepsCmd = "yarn install"
			} else if fileExists(filepath.Join(d, "bun.lockb")) || fileExists(filepath.Join(d, "bun.lock")) {
				i.PackageMgr = "bun"
				i.DepsCmd = "bun install"
			} else {
				i.PackageMgr = "npm"
				i.DepsCmd = "npm install"
			}
			i.BuildCmd = i.PackageMgr + " run build"
			i.TestCmd = i.PackageMgr + " test"
			i.LintCmd = i.PackageMgr + " run lint"
			i.FormatCmd = i.PackageMgr + " run format"
			i.RunCmd = i.PackageMgr + " run dev"
			i.CleanCmd = "rm -rf node_modules dist .next .nuxt build"
			// Detect framework
			if fileExists(filepath.Join(d, "next.config.js")) || fileExists(filepath.Join(d, "next.config.mjs")) || fileExists(filepath.Join(d, "next.config.ts")) {
				i.Framework = "Next.js"
			} else if fileExists(filepath.Join(d, "nuxt.config.ts")) || fileExists(filepath.Join(d, "nuxt.config.js")) {
				i.Framework = "Nuxt"
			} else if fileExists(filepath.Join(d, "vite.config.ts")) || fileExists(filepath.Join(d, "vite.config.js")) {
				i.Framework = "Vite"
			} else if fileExists(filepath.Join(d, "angular.json")) {
				i.Framework = "Angular"
			}
			// Check for TypeScript
			if fileExists(filepath.Join(d, "tsconfig.json")) {
				i.Language = "TypeScript"
			}
		}},
		{"Cargo.toml", func(i *projectInfo, d string) {
			i.Language = "Rust"
			i.PackageFile = "Cargo.toml"
			i.PackageMgr = "cargo"
			i.DepsCmd = "cargo fetch"
			i.BuildCmd = "cargo build"
			i.TestCmd = "cargo test"
			i.LintCmd = "cargo clippy"
			i.FormatCmd = "cargo fmt"
			i.RunCmd = "cargo run"
			i.CleanCmd = "cargo clean"
		}},
		{"pyproject.toml", func(i *projectInfo, d string) {
			i.Language = "Python"
			i.PackageFile = "pyproject.toml"
			if fileExists(filepath.Join(d, "poetry.lock")) {
				i.PackageMgr = "poetry"
				i.DepsCmd = "poetry install"
				i.RunCmd = "poetry run python -m " + filepath.Base(d)
			} else if fileExists(filepath.Join(d, "uv.lock")) {
				i.PackageMgr = "uv"
				i.DepsCmd = "uv sync"
				i.RunCmd = "uv run python -m " + filepath.Base(d)
			} else {
				i.PackageMgr = "pip"
				i.DepsCmd = "pip install -e ."
				i.RunCmd = "python -m " + filepath.Base(d)
			}
			i.TestCmd = "pytest"
			i.LintCmd = "ruff check ."
			i.FormatCmd = "ruff format ."
			i.BuildCmd = "python -m build"
			i.CleanCmd = "rm -rf dist build *.egg-info __pycache__"
		}},
		{"requirements.txt", func(i *projectInfo, d string) {
			if i.Language != "" {
				return // Already detected via pyproject.toml
			}
			i.Language = "Python"
			i.PackageFile = "requirements.txt"
			i.PackageMgr = "pip"
			i.DepsCmd = "pip install -r requirements.txt"
			i.TestCmd = "pytest"
			i.LintCmd = "ruff check ."
			i.FormatCmd = "ruff format ."
			i.RunCmd = "python main.py"
			i.CleanCmd = "rm -rf __pycache__ .pytest_cache"
		}},
		{"Podfile", func(i *projectInfo, d string) {
			i.Language = "Swift"
			i.Framework = "iOS (CocoaPods)"
			i.PackageFile = "Podfile"
			i.PackageMgr = "cocoapods"
			i.DepsCmd = "pod install"
			i.BuildCmd = "xcodebuild -workspace *.xcworkspace -scheme * -sdk iphonesimulator build"
			i.TestCmd = "xcodebuild test -workspace *.xcworkspace -scheme * -sdk iphonesimulator -destination 'platform=iOS Simulator,name=iPhone 15'"
			i.CleanCmd = "xcodebuild clean"
		}},
		{"Package.swift", func(i *projectInfo, d string) {
			i.Language = "Swift"
			i.Framework = "Swift Package Manager"
			i.PackageFile = "Package.swift"
			i.PackageMgr = "swift"
			i.DepsCmd = "swift package resolve"
			i.BuildCmd = "swift build"
			i.TestCmd = "swift test"
			i.RunCmd = "swift run"
			i.CleanCmd = "swift package clean"
		}},
		{"build.gradle", func(i *projectInfo, d string) {
			i.Language = "Kotlin/Java"
			i.Framework = "Android (Gradle)"
			i.PackageFile = "build.gradle"
			i.PackageMgr = "gradle"
			gradle := "./gradlew"
			if _, err := os.Stat(filepath.Join(d, "gradlew")); os.IsNotExist(err) {
				gradle = "gradle"
			}
			i.DepsCmd = gradle + " dependencies"
			i.BuildCmd = gradle + " assembleDebug"
			i.TestCmd = gradle + " test"
			i.LintCmd = gradle + " lint"
			i.RunCmd = gradle + " installDebug"
			i.CleanCmd = gradle + " clean"
		}},
		{"build.gradle.kts", func(i *projectInfo, d string) {
			if i.Language != "" {
				return
			}
			i.Language = "Kotlin"
			i.Framework = "Android (Gradle KTS)"
			i.PackageFile = "build.gradle.kts"
			i.PackageMgr = "gradle"
			gradle := "./gradlew"
			if _, err := os.Stat(filepath.Join(d, "gradlew")); os.IsNotExist(err) {
				gradle = "gradle"
			}
			i.DepsCmd = gradle + " dependencies"
			i.BuildCmd = gradle + " assembleDebug"
			i.TestCmd = gradle + " test"
			i.LintCmd = gradle + " lint"
			i.RunCmd = gradle + " installDebug"
			i.CleanCmd = gradle + " clean"
		}},
		{"Gemfile", func(i *projectInfo, d string) {
			i.Language = "Ruby"
			i.PackageFile = "Gemfile"
			i.PackageMgr = "bundler"
			i.DepsCmd = "bundle install"
			i.TestCmd = "bundle exec rspec"
			i.LintCmd = "bundle exec rubocop"
			i.FormatCmd = "bundle exec rubocop -a"
			// Check for Rails
			if fileExists(filepath.Join(d, "Rakefile")) && fileExists(filepath.Join(d, "config", "application.rb")) {
				i.Framework = "Rails"
				i.RunCmd = "bundle exec rails server"
				i.BuildCmd = "bundle exec rails assets:precompile"
			}
			i.CleanCmd = "rm -rf tmp/cache"
		}},
		{"pom.xml", func(i *projectInfo, d string) {
			if i.Language != "" {
				return
			}
			i.Language = "Java"
			i.Framework = "Maven"
			i.PackageFile = "pom.xml"
			i.PackageMgr = "maven"
			i.DepsCmd = "mvn dependency:resolve"
			i.BuildCmd = "mvn package"
			i.TestCmd = "mvn test"
			i.RunCmd = "mvn exec:java"
			i.CleanCmd = "mvn clean"
		}},
		{"Makefile", func(i *projectInfo, d string) {
			if i.Language != "" {
				return // Don't override already-detected language
			}
			i.Language = "Unknown (Makefile detected)"
			i.PackageFile = "Makefile"
			i.PackageMgr = "make"
			i.BuildCmd = "make"
			i.TestCmd = "make test"
			i.CleanCmd = "make clean"
		}},
	}

	for _, check := range checks {
		if fileExists(filepath.Join(dir, check.file)) {
			check.fn(&info, dir)
		}
	}

	if info.Language == "" {
		info.Language = "Unknown"
	}

	return info
}

func (t *ProjectTool) formatDetection(info projectInfo, dir string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Project: %s\n", dir))
	sb.WriteString(fmt.Sprintf("Language: %s\n", info.Language))
	if info.Framework != "" {
		sb.WriteString(fmt.Sprintf("Framework: %s\n", info.Framework))
	}
	if info.PackageMgr != "" {
		sb.WriteString(fmt.Sprintf("Package Manager: %s\n", info.PackageMgr))
	}
	if info.PackageFile != "" {
		sb.WriteString(fmt.Sprintf("Package File: %s\n", info.PackageFile))
	}
	sb.WriteString("\nAvailable commands:\n")
	cmds := map[string]string{
		"deps":   info.DepsCmd,
		"build":  info.BuildCmd,
		"test":   info.TestCmd,
		"lint":   info.LintCmd,
		"format": info.FormatCmd,
		"run":    info.RunCmd,
		"clean":  info.CleanCmd,
	}
	for name, cmd := range cmds {
		if cmd != "" {
			sb.WriteString(fmt.Sprintf("  %-8s → %s\n", name, cmd))
		}
	}
	return sb.String()
}

func (t *ProjectTool) runProjectCmd(ctx context.Context, cwd, cmd string, args map[string]interface{}) (string, error) {
	if cmd == "" {
		return "No command configured for this action. Use 'detect' to check project type.", nil
	}

	extraArgs, _ := args["args"].(string)
	if extraArgs != "" {
		cmd = cmd + " " + extraArgs
	}

	return t.runShell(ctx, cwd, cmd, 5*time.Minute)
}

func (t *ProjectTool) runScript(ctx context.Context, cwd string, info projectInfo, args map[string]interface{}) (string, error) {
	scriptName, _ := args["args"].(string)
	if scriptName == "" {
		return "", fmt.Errorf("script action requires a script name in args (e.g., 'dev', 'deploy')")
	}

	var cmd string
	switch info.PackageMgr {
	case "npm":
		cmd = "npm run " + scriptName
	case "yarn":
		cmd = "yarn " + scriptName
	case "pnpm":
		cmd = "pnpm " + scriptName
	case "bun":
		cmd = "bun run " + scriptName
	case "make":
		cmd = "make " + scriptName
	case "cargo":
		cmd = "cargo " + scriptName
	case "gradle":
		gradle := "./gradlew"
		if _, err := os.Stat(filepath.Join(cwd, "gradlew")); os.IsNotExist(err) {
			gradle = "gradle"
		}
		cmd = gradle + " " + scriptName
	default:
		cmd = scriptName // Run as raw command
	}

	return t.runShell(ctx, cwd, cmd, 5*time.Minute)
}

func (t *ProjectTool) runShell(ctx context.Context, cwd, command string, timeout time.Duration) (string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd
	cmd = exec.CommandContext(cmdCtx, "sh", "-c", command)
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

	maxLen := 50000
	if len(output) > maxLen {
		output = output[:maxLen] + fmt.Sprintf("\n... (truncated, %d more chars)", len(output)-maxLen)
	}

	return output, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Ensure ProjectTool implements required interfaces
var (
	_ Tool          = (*ProjectTool)(nil)
	_ WorkspaceTool = (*ProjectTool)(nil)
)
