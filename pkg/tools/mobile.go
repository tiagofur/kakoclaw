package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// MobileTool provides iOS and Android build, test, and deployment operations.
// Semi-autonomous: debug builds and tests run freely, release/archive/install require confirmation.
type MobileTool struct {
	workspace string
}

func NewMobileTool(workspace string) *MobileTool {
	return &MobileTool{workspace: workspace}
}

func (t *MobileTool) SetWorkspace(workspace string) {
	t.workspace = workspace
}

func (t *MobileTool) Name() string { return "mobile" }

func (t *MobileTool) Description() string {
	return `iOS and Android mobile development tool.
Actions:
  detect          - Detect mobile platform (iOS/Android) and available tooling
  build           - Build the project (debug by default)
  test            - Run unit/UI tests
  run_simulator   - Launch app in iOS Simulator
  run_emulator    - Launch app in Android Emulator
  list_devices    - List connected devices and available simulators/emulators
  install         - Install app on connected device (requires confirmation)
  archive         - Create release archive/AAB (requires confirmation)
  lint            - Run mobile-specific linting
  deps            - Install dependencies (CocoaPods, SPM, Gradle sync)
  clean           - Clean build artifacts
  signing_status  - Check code signing configuration (iOS)
iOS: Uses xcodebuild, xcrun simctl, fastlane, pod.
Android: Uses ./gradlew, adb, sdkmanager.`
}

func (t *MobileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type": "string",
				"enum": []string{
					"detect", "build", "test", "run_simulator", "run_emulator",
					"list_devices", "install", "archive", "lint", "deps", "clean",
					"signing_status",
				},
				"description": "The mobile action to perform",
			},
			"platform": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"ios", "android", "auto"},
				"description": "Target platform (auto-detected if omitted)",
			},
			"args": map[string]interface{}{
				"type":        "string",
				"description": "Additional arguments (e.g., scheme name, build variant, device ID)",
			},
			"configuration": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"debug", "release"},
				"description": "Build configuration (default: debug)",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Project directory (defaults to workspace)",
			},
		},
		"required": []string{"action"},
	}
}

func (t *MobileTool) RequiresConfirmation(action string, args map[string]interface{}) (bool, string) {
	switch action {
	case "install":
		return true, "Install app on connected device. Confirm?"
	case "archive":
		return true, "Create release archive (this may trigger code signing). Confirm?"
	}
	return false, ""
}

type mobileInfo struct {
	Platform string // "ios", "android", "both", "unknown"
	// iOS
	HasXcode      bool
	HasFastlane   bool
	HasCocoapods  bool
	Workspace     string // .xcworkspace path
	XcodeProject  string // .xcodeproj path
	Scheme        string
	HasSPM        bool
	// Android
	HasGradle    bool
	HasADB       bool
	GradleCmd    string
	HasKotlin    bool
}

func (t *MobileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action is required")
	}

	cwd := t.workspace
	if wd, ok := args["working_dir"].(string); ok && wd != "" {
		cwd = wd
	}
	if cwd == "" {
		wd, _ := os.Getwd()
		cwd = wd
	}

	platform, _ := args["platform"].(string)
	if platform == "" || platform == "auto" {
		platform = t.detectPlatform(cwd)
	}

	config, _ := args["configuration"].(string)
	if config == "" {
		config = "debug"
	}

	info := t.detectMobile(cwd)

	switch action {
	case "detect":
		return t.formatDetection(info, cwd), nil
	case "build":
		return t.build(ctx, cwd, platform, config, info, args)
	case "test":
		return t.test(ctx, cwd, platform, info, args)
	case "run_simulator":
		return t.runSimulator(ctx, cwd, info, args)
	case "run_emulator":
		return t.runEmulator(ctx, cwd, info, args)
	case "list_devices":
		return t.listDevices(ctx, cwd, platform)
	case "install":
		return t.install(ctx, cwd, platform, info, args)
	case "archive":
		return t.archive(ctx, cwd, platform, config, info, args)
	case "lint":
		return t.lint(ctx, cwd, platform, info, args)
	case "deps":
		return t.installDeps(ctx, cwd, platform, info)
	case "clean":
		return t.clean(ctx, cwd, platform, info)
	case "signing_status":
		return t.signingStatus(ctx, cwd, info)
	default:
		return "", fmt.Errorf("unknown mobile action: %s", action)
	}
}

func (t *MobileTool) detectPlatform(dir string) string {
	hasIOS := fileExists(filepath.Join(dir, "Podfile")) ||
		fileExists(filepath.Join(dir, "Package.swift")) ||
		t.hasXcodeProject(dir)
	hasAndroid := fileExists(filepath.Join(dir, "build.gradle")) ||
		fileExists(filepath.Join(dir, "build.gradle.kts")) ||
		fileExists(filepath.Join(dir, "app", "build.gradle")) ||
		fileExists(filepath.Join(dir, "app", "build.gradle.kts"))

	if hasIOS && hasAndroid {
		return "both"
	}
	if hasIOS {
		return "ios"
	}
	if hasAndroid {
		return "android"
	}
	return "unknown"
}

func (t *MobileTool) hasXcodeProject(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".xcodeproj") || strings.HasSuffix(e.Name(), ".xcworkspace") {
			return true
		}
	}
	return false
}

func (t *MobileTool) detectMobile(dir string) mobileInfo {
	info := mobileInfo{}

	// iOS detection
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("xcodebuild"); err == nil {
			info.HasXcode = true
		}
	}
	if _, err := exec.LookPath("fastlane"); err == nil {
		info.HasFastlane = true
	}
	if _, err := exec.LookPath("pod"); err == nil {
		info.HasCocoapods = true
	}
	info.HasSPM = fileExists(filepath.Join(dir, "Package.swift"))

	// Find .xcworkspace and .xcodeproj
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".xcworkspace") && !strings.Contains(e.Name(), "Pods") {
			info.Workspace = e.Name()
		}
		if strings.HasSuffix(e.Name(), ".xcodeproj") {
			info.XcodeProject = e.Name()
		}
	}

	// Android detection
	if _, err := exec.LookPath("adb"); err == nil {
		info.HasADB = true
	}
	if fileExists(filepath.Join(dir, "gradlew")) {
		info.HasGradle = true
		info.GradleCmd = "./gradlew"
	} else if fileExists(filepath.Join(dir, "app", "build.gradle")) || fileExists(filepath.Join(dir, "app", "build.gradle.kts")) {
		if _, err := exec.LookPath("gradle"); err == nil {
			info.HasGradle = true
			info.GradleCmd = "gradle"
		}
	}
	info.HasKotlin = fileExists(filepath.Join(dir, "app", "src", "main", "kotlin")) ||
		fileExists(filepath.Join(dir, "build.gradle.kts"))

	// Determine platform
	info.Platform = t.detectPlatform(dir)

	return info
}

func (t *MobileTool) formatDetection(info mobileInfo, dir string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Project: %s\n", dir))
	sb.WriteString(fmt.Sprintf("Platform: %s\n\n", info.Platform))

	sb.WriteString("iOS Tooling:\n")
	sb.WriteString(fmt.Sprintf("  Xcode:      %s\n", boolStatus(info.HasXcode)))
	sb.WriteString(fmt.Sprintf("  Fastlane:   %s\n", boolStatus(info.HasFastlane)))
	sb.WriteString(fmt.Sprintf("  CocoaPods:  %s\n", boolStatus(info.HasCocoapods)))
	sb.WriteString(fmt.Sprintf("  SPM:        %s\n", boolStatus(info.HasSPM)))
	if info.Workspace != "" {
		sb.WriteString(fmt.Sprintf("  Workspace:  %s\n", info.Workspace))
	}
	if info.XcodeProject != "" {
		sb.WriteString(fmt.Sprintf("  Project:    %s\n", info.XcodeProject))
	}

	sb.WriteString("\nAndroid Tooling:\n")
	sb.WriteString(fmt.Sprintf("  Gradle:     %s\n", boolStatus(info.HasGradle)))
	sb.WriteString(fmt.Sprintf("  ADB:        %s\n", boolStatus(info.HasADB)))
	sb.WriteString(fmt.Sprintf("  Kotlin:     %s\n", boolStatus(info.HasKotlin)))
	if info.GradleCmd != "" {
		sb.WriteString(fmt.Sprintf("  Gradle Cmd: %s\n", info.GradleCmd))
	}

	return sb.String()
}

func boolStatus(b bool) string {
	if b {
		return "Available"
	}
	return "Not found"
}

func (t *MobileTool) build(ctx context.Context, cwd, platform, config string, info mobileInfo, args map[string]interface{}) (string, error) {
	extraArgs, _ := args["args"].(string)

	switch platform {
	case "ios":
		if !info.HasXcode {
			return "Xcode is required for iOS builds. This tool only works on macOS.", nil
		}
		cmd := t.iosBuildCmd(info, config)
		if extraArgs != "" {
			cmd += " " + extraArgs
		}
		return t.runMobileCmd(ctx, cwd, cmd, 10*time.Minute)

	case "android":
		if !info.HasGradle {
			return "Gradle is required for Android builds. Ensure gradlew exists or gradle is installed.", nil
		}
		task := "assembleDebug"
		if config == "release" {
			task = "assembleRelease"
		}
		cmd := info.GradleCmd + " " + task
		if extraArgs != "" {
			cmd += " " + extraArgs
		}
		return t.runMobileCmd(ctx, cwd, cmd, 10*time.Minute)

	case "both":
		// Build iOS first, then Android
		results := []string{}
		if info.HasXcode {
			cmd := t.iosBuildCmd(info, config)
			r, _ := t.runMobileCmd(ctx, cwd, cmd, 10*time.Minute)
			results = append(results, "=== iOS Build ===\n"+r)
		}
		if info.HasGradle {
			task := "assembleDebug"
			if config == "release" {
				task = "assembleRelease"
			}
			r, _ := t.runMobileCmd(ctx, cwd, info.GradleCmd+" "+task, 10*time.Minute)
			results = append(results, "=== Android Build ===\n"+r)
		}
		return strings.Join(results, "\n\n"), nil

	default:
		return "Could not detect mobile platform. Ensure Podfile, .xcodeproj, or build.gradle exists.", nil
	}
}

func (t *MobileTool) iosBuildCmd(info mobileInfo, config string) string {
	xcConfig := "Debug"
	if config == "release" {
		xcConfig = "Release"
	}

	if info.Workspace != "" {
		return fmt.Sprintf("xcodebuild -workspace %s -scheme %s -configuration %s -sdk iphonesimulator build",
			info.Workspace, t.guessScheme(info), xcConfig)
	}
	if info.XcodeProject != "" {
		return fmt.Sprintf("xcodebuild -project %s -scheme %s -configuration %s -sdk iphonesimulator build",
			info.XcodeProject, t.guessScheme(info), xcConfig)
	}
	return "xcodebuild build"
}

func (t *MobileTool) guessScheme(info mobileInfo) string {
	if info.Scheme != "" {
		return info.Scheme
	}
	// Strip extension from project name
	name := info.XcodeProject
	if info.Workspace != "" {
		name = info.Workspace
	}
	name = strings.TrimSuffix(name, ".xcworkspace")
	name = strings.TrimSuffix(name, ".xcodeproj")
	if name == "" {
		return "App"
	}
	return name
}

func (t *MobileTool) test(ctx context.Context, cwd, platform string, info mobileInfo, args map[string]interface{}) (string, error) {
	extraArgs, _ := args["args"].(string)

	switch platform {
	case "ios":
		cmd := fmt.Sprintf("xcodebuild test -scheme %s -sdk iphonesimulator -destination 'platform=iOS Simulator,name=iPhone 16'",
			t.guessScheme(info))
		if info.Workspace != "" {
			cmd = fmt.Sprintf("xcodebuild test -workspace %s -scheme %s -sdk iphonesimulator -destination 'platform=iOS Simulator,name=iPhone 16'",
				info.Workspace, t.guessScheme(info))
		}
		if extraArgs != "" {
			cmd += " " + extraArgs
		}
		return t.runMobileCmd(ctx, cwd, cmd, 10*time.Minute)

	case "android":
		cmd := info.GradleCmd + " test"
		if extraArgs != "" {
			cmd += " " + extraArgs
		}
		return t.runMobileCmd(ctx, cwd, cmd, 10*time.Minute)

	default:
		return "Could not detect platform for testing.", nil
	}
}

func (t *MobileTool) runSimulator(ctx context.Context, cwd string, info mobileInfo, args map[string]interface{}) (string, error) {
	if !info.HasXcode {
		return "Xcode is required to run iOS Simulator. This only works on macOS.", nil
	}
	extraArgs, _ := args["args"].(string)

	// Boot simulator and run app
	cmds := []string{
		"xcrun simctl boot 'iPhone 16' 2>/dev/null || true",
		"open -a Simulator",
	}

	results := []string{}
	for _, cmd := range cmds {
		r, _ := t.runMobileCmd(ctx, cwd, cmd, 30*time.Second)
		results = append(results, r)
	}

	// Build and install
	buildCmd := t.iosBuildCmd(info, "debug")
	if extraArgs != "" {
		buildCmd += " " + extraArgs
	}
	r, err := t.runMobileCmd(ctx, cwd, buildCmd, 5*time.Minute)
	results = append(results, r)

	return strings.Join(results, "\n"), err
}

func (t *MobileTool) runEmulator(ctx context.Context, cwd string, info mobileInfo, args map[string]interface{}) (string, error) {
	if !info.HasADB {
		return "ADB is required to run Android Emulator. Install Android SDK platform-tools.", nil
	}

	// List available AVDs and start one
	r, err := t.runMobileCmd(ctx, cwd, "emulator -list-avds", 10*time.Second)
	if err != nil {
		return "Could not list emulators: " + r, nil
	}

	avds := strings.Split(strings.TrimSpace(r), "\n")
	if len(avds) == 0 || avds[0] == "" {
		return "No Android Virtual Devices found. Create one with: avdmanager create avd", nil
	}

	// Start first available AVD in background
	return t.runMobileCmd(ctx, cwd, fmt.Sprintf("emulator -avd %s &", avds[0]), 10*time.Second)
}

func (t *MobileTool) listDevices(ctx context.Context, cwd, platform string) (string, error) {
	results := []string{}

	if platform == "ios" || platform == "both" || platform == "auto" {
		r, _ := t.runMobileCmd(ctx, cwd, "xcrun simctl list devices available --json 2>/dev/null || xcrun simctl list devices available", 10*time.Second)
		results = append(results, "=== iOS Devices/Simulators ===\n"+r)
	}

	if platform == "android" || platform == "both" || platform == "auto" {
		r, _ := t.runMobileCmd(ctx, cwd, "adb devices -l 2>/dev/null || echo 'ADB not available'", 10*time.Second)
		results = append(results, "=== Android Devices ===\n"+r)
	}

	return strings.Join(results, "\n\n"), nil
}

func (t *MobileTool) install(ctx context.Context, cwd, platform string, info mobileInfo, args map[string]interface{}) (string, error) {
	extraArgs, _ := args["args"].(string)

	switch platform {
	case "ios":
		if extraArgs == "" {
			return "Specify the .app path in args for iOS device installation.", nil
		}
		return t.runMobileCmd(ctx, cwd, "xcrun simctl install booted "+extraArgs, 60*time.Second)
	case "android":
		apkPath := extraArgs
		if apkPath == "" {
			apkPath = "app/build/outputs/apk/debug/app-debug.apk"
		}
		return t.runMobileCmd(ctx, cwd, "adb install -r "+apkPath, 60*time.Second)
	default:
		return "Specify platform (ios/android) for install.", nil
	}
}

func (t *MobileTool) archive(ctx context.Context, cwd, platform, config string, info mobileInfo, args map[string]interface{}) (string, error) {
	switch platform {
	case "ios":
		if info.HasFastlane {
			return t.runMobileCmd(ctx, cwd, "fastlane build", 15*time.Minute)
		}
		scheme := t.guessScheme(info)
		cmd := fmt.Sprintf("xcodebuild -scheme %s -configuration Release -archivePath build/%s.xcarchive archive", scheme, scheme)
		if info.Workspace != "" {
			cmd = fmt.Sprintf("xcodebuild -workspace %s -scheme %s -configuration Release -archivePath build/%s.xcarchive archive",
				info.Workspace, scheme, scheme)
		}
		return t.runMobileCmd(ctx, cwd, cmd, 15*time.Minute)

	case "android":
		return t.runMobileCmd(ctx, cwd, info.GradleCmd+" bundleRelease", 15*time.Minute)

	default:
		return "Specify platform (ios/android) for archive.", nil
	}
}

func (t *MobileTool) lint(ctx context.Context, cwd, platform string, info mobileInfo, args map[string]interface{}) (string, error) {
	switch platform {
	case "ios":
		if _, err := exec.LookPath("swiftlint"); err == nil {
			return t.runMobileCmd(ctx, cwd, "swiftlint", 60*time.Second)
		}
		return "SwiftLint not installed. Install with: brew install swiftlint", nil
	case "android":
		if info.HasGradle {
			return t.runMobileCmd(ctx, cwd, info.GradleCmd+" lint", 5*time.Minute)
		}
		return "Gradle not available for linting.", nil
	default:
		return "Specify platform (ios/android) for lint.", nil
	}
}

func (t *MobileTool) installDeps(ctx context.Context, cwd, platform string, info mobileInfo) (string, error) {
	results := []string{}

	if platform == "ios" || platform == "both" {
		if info.HasCocoapods && fileExists(filepath.Join(cwd, "Podfile")) {
			r, _ := t.runMobileCmd(ctx, cwd, "pod install", 5*time.Minute)
			results = append(results, "=== CocoaPods ===\n"+r)
		}
		if info.HasSPM {
			r, _ := t.runMobileCmd(ctx, cwd, "swift package resolve", 3*time.Minute)
			results = append(results, "=== Swift Package Manager ===\n"+r)
		}
	}

	if platform == "android" || platform == "both" {
		if info.HasGradle {
			r, _ := t.runMobileCmd(ctx, cwd, info.GradleCmd+" dependencies", 5*time.Minute)
			results = append(results, "=== Gradle Dependencies ===\n"+r)
		}
	}

	if len(results) == 0 {
		return "No dependency managers detected.", nil
	}
	return strings.Join(results, "\n\n"), nil
}

func (t *MobileTool) clean(ctx context.Context, cwd, platform string, info mobileInfo) (string, error) {
	results := []string{}

	if platform == "ios" || platform == "both" {
		r, _ := t.runMobileCmd(ctx, cwd, "xcodebuild clean 2>/dev/null; rm -rf build DerivedData", 30*time.Second)
		results = append(results, "iOS clean: "+r)
	}

	if platform == "android" || platform == "both" {
		if info.HasGradle {
			r, _ := t.runMobileCmd(ctx, cwd, info.GradleCmd+" clean", 60*time.Second)
			results = append(results, "Android clean: "+r)
		}
	}

	if len(results) == 0 {
		return "No platform detected for cleaning.", nil
	}
	return strings.Join(results, "\n"), nil
}

func (t *MobileTool) signingStatus(ctx context.Context, cwd string, info mobileInfo) (string, error) {
	if !info.HasXcode {
		return "Code signing status is only available on macOS with Xcode.", nil
	}

	results := []string{}

	// List signing identities
	r, _ := t.runMobileCmd(ctx, cwd, "security find-identity -v -p codesigning", 10*time.Second)
	results = append(results, "=== Signing Identities ===\n"+r)

	// List provisioning profiles
	r, _ = t.runMobileCmd(ctx, cwd, "ls ~/Library/MobileDevice/Provisioning\\ Profiles/ 2>/dev/null || echo 'No profiles found'", 5*time.Second)
	results = append(results, "=== Provisioning Profiles ===\n"+r)

	return strings.Join(results, "\n\n"), nil
}

func (t *MobileTool) runMobileCmd(ctx context.Context, cwd, command string, timeout time.Duration) (string, error) {
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

	maxLen := 50000
	if len(output) > maxLen {
		output = output[:maxLen] + fmt.Sprintf("\n... (truncated, %d more chars)", len(output)-maxLen)
	}

	return output, nil
}

var (
	_ Tool            = (*MobileTool)(nil)
	_ WorkspaceTool   = (*MobileTool)(nil)
	_ ConfirmableTool = (*MobileTool)(nil)
)
