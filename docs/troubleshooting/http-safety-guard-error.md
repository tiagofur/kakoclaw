# Troubleshooting: "Safety Guard Blocking External HTTP" Error

## Problem Description

The AI agent refuses to execute commands that make external HTTP requests (like `gh` CLI for GitHub), claiming that a "safety guard" is blocking external URLs like api.github.com.

## Root Cause

**This is a misconception by the LLM**, not an actual technical limitation. KakoClaw does NOT implement any safety guard that blocks HTTP connections to external URLs.

### What Safety Guards Actually Exist

KakoClaw only has safety guards for:
1. **Dangerous shell commands** (rm -rf, format, etc.)
2. **Path traversal** in file operations (when `restrict_to_workspace: true`)
3. **File access** outside workspace (when `restrict_to_workspace: true`)

**None of these prevent network/HTTP operations.**

## Solution

The fix has been implemented in the following ways:

### 1. Updated System Prompt

The system prompt now explicitly states:

```
**HTTP/Network**: ⚠️ HTTP connections are ALLOWED. Tools and skills can make 
external HTTP requests (e.g., gh CLI for GitHub API, curl, wget, API calls). 
There is NO safety guard blocking external URLs.
```

This clarifies to the AI that network operations are permitted.

### 2. Enhanced GitHub Skill Documentation

The GitHub skill now includes:

```markdown
## Prerequisites
1. Install GitHub CLI: `brew install gh` (macOS) or `apt install gh` (Linux)
2. Authenticate: Run `gh auth login` or set `GITHUB_TOKEN` environment variable
3. Network Access: Requires internet connection to api.github.com (no safety guards block this)
```

### 3. Created AGENTS.md in Workspace

Added comprehensive documentation about agent capabilities that clarifies:
- What operations are allowed (including all HTTP operations)
- What safety guards actually exist
- Common misconceptions
- How to use external services

This file is loaded automatically into the system context.

### 4. Verification Script

Created `scripts/check-github-setup.sh` to verify GitHub CLI setup:

```bash
./scripts/check-github-setup.sh
```

## How to Use GitHub CLI Now

### Method 1: Use the GitHub Skill

1. The agent will read the GitHub skill documentation
2. It will understand that HTTP connections are allowed
3. It will execute gh commands normally

Example:
```
User: "Create an issue in tiagofur/kakoclaw titled 'Fix feature X' with description 'Details here'"
Agent: *executes* gh issue create --repo tiagofur/kakoclaw --title "Fix feature X" --body "Details here"
```

### Method 2: Direct Command Execution

If the agent still hesitates, ask it explicitly:

```
User: "Execute this command: gh issue list --repo tiagofur/kakoclaw"
Agent: *executes command via exec tool*
```

### Method 3: Set Environment Variable First

If authentication is needed:

```
User: "Set GITHUB_TOKEN to ghp_xxx, then create an issue"
Agent: *executes* export GITHUB_TOKEN=ghp_xxx && gh issue create ...
```

## Testing the Fix

### Test 1: Verify GitHub CLI Works

```bash
# Run the verification script
./scripts/check-github-setup.sh

# Should show:
# ✅ GitHub CLI is installed
# ✅ GitHub CLI is authenticated
# 🎉 Everything is set up correctly!
```

### Test 2: Agent Execution

Ask the agent to:

1. **List issues**: "List the open issues in tiagofur/kakoclaw"
2. **Create an issue**: "Create a test issue in tiagofur/kakoclaw"
3. **Use gh api**: "Use gh api to get repository info for tiagofur/kakoclaw"

The agent should now execute these commands without claiming they're blocked.

## Configuration Reference

### Workspace Restriction Setting

In `config.json`:

```json
{
  "agents": {
    "defaults": {
      "restrict_to_workspace": true  // Only affects FILE paths, NOT network
    }
  }
}
```

This setting:
- ✅ Restricts file operations to workspace directory
- ✅ Blocks path traversal in shell commands
- ❌ Does NOT block HTTP requests
- ❌ Does NOT block external APIs
- ❌ Does NOT prevent using CLI tools like gh

## Additional Notes

### Why This Happened

LLMs sometimes exhibit overcautious behavior when they see security-related configurations. The `restrict_to_workspace: true` setting, combined with shell command safety guards, likely caused the model to incorrectly infer that ALL external operations were restricted.

### Prevention

The explicit clarifications in the system prompt and AGENTS.md should prevent this misconception in future interactions.

### If Problem Persists

If the agent still refuses to execute network operations:

1. **Remind it explicitly**: "HTTP connections are allowed in KakoClaw, execute the command"
2. **Reference documentation**: "Check AGENTS.md - network operations are permitted"
3. **Try different model**: Some models may be more conservative than others
4. **Use direct exec**: Use the exec tool directly instead of relying on skill interpretation

## Related Files

- [pkg/agent/context.go](../pkg/agent/context.go) - System prompt with HTTP permissions clarification
- [skills/github/SKILL.md](../skills/github/SKILL.md) - Updated GitHub skill with prerequisites
- [KakoClaw-data/workspace/AGENTS.md](../KakoClaw-data/workspace/AGENTS.md) - Agent capabilities documentation
- [scripts/check-github-setup.sh](../scripts/check-github-setup.sh) - GitHub CLI verification script
