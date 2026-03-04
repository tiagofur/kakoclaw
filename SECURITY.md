# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.9.x   | :white_check_mark: |
| 0.8.x   | :white_check_mark: |
| < 0.8   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

1. **DO NOT** create a public GitHub issue for security vulnerabilities
2. Email security concerns to: [security@makoclaw.dev](mailto:security@makoclaw.dev)
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial Assessment**: Within 1 week
- **Resolution Target**: Within 30 days for critical issues

### What to Expect

1. We will acknowledge receipt of your report
2. We will investigate and validate the issue
3. We will work on a fix and coordinate disclosure
4. We will credit you in the security advisory (unless you prefer anonymity)

---

## Security Best Practices for Users

### API Keys and Credentials

```yaml
# DO NOT commit API keys to version control
# Use environment variables instead:
export MAKOCLAW_PROVIDERS_OPENAI_APIKEY="sk-..."
export MAKOCLAW_PROVIDERS_ANTHROPIC_APIKEY="sk-ant-..."
```

### Shell Command Security

The `exec` tool has built-in security controls:

1. **Allowlist**: Only permitted commands can run
2. **Path validation**: Prevents path traversal attacks
3. **Denylist**: Dangerous commands are blocked by default

Configure allowed commands in `config.json`:

```json
{
  "tool_permissions": {
    "allowed_shell_commands": ["git", "npm", "go", "make"]
  }
}
```

### Multi-User Deployment

When deploying for multiple users:

1. **Enable authentication**: Always use JWT auth in production
2. **Use HTTPS**: Never expose HTTP in production
3. **Isolate workspaces**: Each user gets isolated storage
4. **Review permissions**: Configure role-based tool access

### Channel Security

For messaging channels:

1. **Use allowlists**: Configure `allow_from` for each channel
2. **Validate tokens**: Keep bot tokens secure
3. **Monitor usage**: Review audit logs regularly

---

## Known Security Considerations

### Current Limitations

1. **Ollama Provider**: Does not support tool restrictions (tools parameter ignored)
2. **PDF Extraction**: Basic implementation may not sanitize all content
3. **WebSocket**: Ensure proper origin validation in production

### Audit Logging

All tool executions are logged to SQLite:

```sql
-- View recent tool executions
SELECT * FROM audit_log ORDER BY created_at DESC LIMIT 100;
```

### Rate Limiting

Built-in rate limiting protects against:
- Login brute force (5 attempts per minute)
- API abuse (configurable per endpoint)

---

## Security Architecture

### Authentication Flow

```
User → Login → JWT Token → API Requests → Validation → Response
                ↓
         Refresh Token (httpOnly cookie)
```

### Tool Execution Flow

```
User Request → Permission Check → Allowlist Check → Sandbox → Execute → Audit Log
```

### Multi-User Isolation

```
~/.MakoClaw/
├── central.db          # Shared: users, auth only
└── users/
    └── {uuid}/
        ├── config.json   # Isolated: user settings
        ├── workspace/
        │   ├── database.db  # Isolated: user data
        │   └── sessions/    # Isolated: chat history
        └── skills/          # Isolated: user skills
```

---

## Vulnerability Disclosure History

| Date | Issue | Severity | Status |
|------|-------|----------|--------|
| 2026-03-04 | Debug logging exposure | High | Documented |
| 2026-03-04 | Path traversal in backups | Medium | Documented |
| 2026-03-04 | Goroutine leak in orchestrator | Medium | Documented |

See [BUGS-KNOWN-ISSUES.md](BUGS-KNOWN-ISSUES.md) for full details.

---

## Security Checklist for Contributors

Before submitting PRs:

- [ ] No hardcoded credentials or API keys
- [ ] No `fmt.Printf` with sensitive data
- [ ] SQL queries use parameterized statements
- [ ] File paths are validated against traversal
- [ ] Shell commands go through allowlist
- [ ] Error messages don't leak internal details
- [ ] Tests don't require real credentials

---

## Third-Party Dependencies

We monitor dependencies for known vulnerabilities using:

```bash
# Run security scan
make security  # Runs go vet + govulncheck
```

Report dependency vulnerabilities the same way as code vulnerabilities.

---

## Contact

- Security issues: security@makoclaw.dev
- General questions: [GitHub Discussions](https://github.com/sipeed/makoclaw/discussions)
- Bug reports: [GitHub Issues](https://github.com/sipeed/makoclaw/issues)
