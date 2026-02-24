# Tool Permissions System

## Overview

MakoClaw implements a granular tool permission system that restricts access to potentially dangerous tools based on user roles. This prevents security issues like unauthorized file modifications, arbitrary command execution, or resource abuse.

## Architecture

### Permission Levels

The system implements **role-based access control (RBAC)** with individual user overrides:

1. **Role Defaults**: Base permissions for each role (`admin`, `user`)
2. **User Overrides**: Custom tool lists that override role defaults for specific users
3. **Shell Command Allowlist**: Restricted set of safe commands for non-admin users

### Restricted Tools

The following tools are considered **restricted** and are logged in the audit trail:

- `exec` - Shell command execution
- `spawn` - Subagent creation (resource intensive)
- `email` - Send emails (spam/phishing risk)  
- `write_file` - Write files to disk
- `edit_file` - Modify existing files
- `append_file` - Append content to files
- `web_fetch` - Download web content (malware risk)

### Safe Tools (Available to All Users)

Regular users retain access to read-only and low-risk tools:

- `read_file` - Read file contents
- `list_dir` - List directory contents
- `task_manager` - Manage Kanban tasks
- `query_knowledge` - Search knowledge base  
- `memory` - Access long-term memory
- `web_search` - Brave Search API (read-only)
- `message` - Send messages across channels

## Configuration

### JSON Configuration Structure

Add this to your `~/.MakoClaw/config.json`:

```json
{
  "tool_permissions": {
    "role_defaults": {
      "admin": ["*"],
      "user": [
        "read_file",
        "list_dir",
        "task_manager",
        "query_knowledge",
        "memory",
        "web_search",
        "message",
        "exec_restricted"
      ]
    },
    "allowed_shell_commands": [
      "ls", "dir",
      "cat", "type", 
      "head", "tail",
      "grep", "findstr",
      "find", "where",
      "pwd", "cd",
      "echo",
      "date",
      "whoami",
      "which",
      "wc",
      "sort",
      "uniq",
      "diff",
      "tree",
      "file",
      "stat"
    ],
    "user_overrides": {
      "trusted_user": ["*"],
      "developer": ["read_file", "list_dir", "write_file", "exec_restricted"]
    }
  }
}
```

### Special Permission Values

- **`"*"`** - Wildcard grants access to ALL tools (admins only)
- **`"exec_restricted"`** - Allows `exec` tool but ONLY with allowlist commands

### Default Safe Shell Commands

When `exec_restricted` is enabled, users can execute these read-only commands:

| Command | Purpose |
|---------|---------|
| `ls`, `dir` | List directory contents |
| `cat`, `type` | View file contents |
| `head`, `tail` | View partial file contents |
| `grep`, `findstr` | Search in files |
| `find`, `where` | Locate files |
| `pwd`, `cd` | Working directory |
| `echo` | Print text |
| `date` | Show current date/time |
| `whoami` | Show current user |
| `which` | Locate command path |
| `wc` | Count lines/words |
| `sort` | Sort output |
| `uniq` | Remove duplicates |
| `diff` | Compare files |
| `tree` | Directory tree view |
| `file` | Identify file type |
| `stat` | File statistics |

**Security Note**: Even with allowlist, the shell command blacklist (blocking `rm -rf`, `shutdown`, etc.) remains active for all users.

## Admin API Endpoints

### Get Tool Permissions Matrix

```bash
GET /api/v1/tools/permissions
Authorization: Bearer <admin_token>
```

**Response:**
```json
{
  "role_defaults": {
    "admin": ["*"],
    "user": ["read_file", "list_dir", "exec_restricted"]
  },
  "allowed_shell_commands": ["ls", "cat", "grep"],
  "user_overrides": {}
}
```

### Update Role Permissions

```bash
PUT /api/v1/tools/permissions
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "role_defaults": {
    "admin": ["*"],
    "user": ["read_file", "list_dir"]
  },
  "allowed_shell_commands": ["ls", "cat", "grep", "find"]
}
```

**Important**: Admin role MUST always have `["*"]` permission.

### Get User's Effective Tool Permissions

```bash
GET /api/v1/users/:id/tools
Authorization: Bearer <admin_token|own_token>
```

**Response:**
```json
{
  "user_id": 42,
  "username": "alice",
  "role": "user",
  "role_defaults": ["read_file", "list_dir"],
  "user_overrides": null,
  "effective_tools": ["read_file", "list_dir"]
}
```

### Set User-Specific Tool Overrides

```bash
PUT /api/v1/users/:id/tools
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "allowed_tools": ["read_file", "write_file", "exec_restricted"]
}
```

To **reset to role defaults**, send:
```json
{
  "allowed_tools": null
}
```

### Query Audit Logs

```bash
GET /api/v1/tools/audit?user_id=42&tool=exec&limit=100
Authorization: Bearer <admin_token>
```

**Query Parameters:**
- `user_id` (optional) - Filter by user ID
- `tool` (optional) - Filter by tool name
- `start` (optional) - Start time (RFC3339)
- `end` (optional) - End time (RFC3339)
- `limit` (optional) - Max results (default: 100, max: 1000)
- `offset` (optional) - Pagination offset

**Response:**
```json
{
  "logs": [
    {
      "id": 1,
      "timestamp": "2026-02-23T10:30:00Z",
      "user_id": 42,
      "username": "alice",
      "tool": "exec",
      "arguments": {
        "command": "ls -la"
      },
      "success": true,
      "error": "",
      "duration_ms": 45
    }
  ],
  "count": 1
}
```

**Note**: Sensitive arguments (passwords, tokens) are automatically redacted in logs.

## Security Considerations

### Shell Command Allowlist

The `exec` tool with `exec_restricted` permission enforces a **prefix-based whitelist**:

- Commands are matched at the start of the string
- Only whitelisted commands can be the first token
- Blocks command injection like `ls; rm -rf /`
- Blacklist (fork bombs, `rm -rf`, `shutdown`) is always active

### Workspace Boundary

When `agents.defaults.restrict_to_workspace: true` in config:

- All filesystem tools validate paths against user's workspace
- Blocks path traversal attempts (`../../../etc/passwd`)
- Each user has isolated workspace: `~/.MakoClaw/users/{uuid}/workspace/`

### Audit Trail

Every execution of a restricted tool is logged with:

- User ID and username
- Tool name and sanitized arguments
- Success/failure status
- Execution duration
- Timestamp

Logs are stored in `tool_audit_log` table in user's database.

## Use Cases

### Scenario 1: Internal Users

Company wants employees to interact with the agent but prevent accidental system damage:

```json
{
  "role_defaults": {
    "user": ["read_file", "list_dir", "task_manager", "web_search", "exec_restricted"]
  },
  "allowed_shell_commands": ["ls", "cat", "grep", "find", "pwd"]
}
```

### Scenario 2: External API Users

Public API where `exec` access should be completely blocked:

```json
{
  "role_defaults": {
    "user": ["read_file", "list_dir", "query_knowledge", "web_search"]
  }
}
```
No `exec` or `exec_restricted` - completely denied.

### Scenario 3: Trusted Power User

Developer who needs write access for specific automation:

```json
{
  "user_overrides": {
    "john_dev": ["*"]
  }
}
```
Temporary elevation to admin-level tools.

## Troubleshooting

### "Command blocked by safety guard (not in allowlist)"

**Cause**: User tried to execute a shell command not in `allowed_shell_commands`.

**Solutions**:
1. Add command to the allowlist (if safe)
2. Grant user `write_file` + external tool instead of scripts
3. Have admin execute the command manually

### "Forbidden: admin role required"

**Cause**: Non-admin user tried to modify tool permissions.

**Solution**: Only admins can change tool permissions. Contact your administrator.

### Tool permission changes not taking effect

**Cause**: Agent loop caches tools at initialization.

**Solution**: Restart the gateway or web service to reload configuration:
```bash
systemctl restart makoclaw-gateway
```

## Development Notes

### Adding New Restricted Tools

1. Add tool name to `RestrictedTools` map in `pkg/tools/audit.go`
2. Tool will automatically be logged in audit trail
3. Add to appropriate role default lists

### Extending Audit Logger

To add custom audit fields, modify:
- `ToolExecutionLog` struct in `pkg/tools/audit.go`
- `CREATE TABLE tool_audit_log` schema
- `LogToolExecution()` method

### Testing Permissions

```go
// Test that user cannot access restricted tool
user, _ := store.GetUserByID(2)
tools := user.GetEffectiveToolPermissions(roleDefaults)
assert.NotContains(t, tools, "spawn")
```

## Related Documentation

- [CONFIG_PERMISSIONS.md](CONFIG_PERMISSIONS.md) - Web API permission matrix
- [MULTI_AGENT_SETUP.md](MULTI_AGENT_SETUP.md) - Specialist agent tool restrictions
- [http-safety-guard-error.md](../troubleshooting/http-safety-guard-error.md) - Shell command safety
