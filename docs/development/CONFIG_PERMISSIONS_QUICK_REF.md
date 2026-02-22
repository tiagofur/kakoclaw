# Config Permissions Quick Reference

## API Endpoints at a Glance

### Regular User Capabilities
```bash
# ✅ View personal config (merged with global defaults)
GET /api/v1/me/config

# ✅ View personal provider configs (empty for new users)
GET /api/v1/me/providers

# ✅ Update personal provider API keys
PUT /api/v1/me/providers/update?provider=openai
{
  "api_key": "sk-proj-...",
  "api_base": "https://api.openai.com/v1"
}

# ✅ Update personal config sections
POST /api/v1/me/config/update
{
  "agents": {
    "defaults": {
      "model": "gpt-4"
    }
  }
}

# ❌ FORBIDDEN: Cannot modify global config
POST /api/v1/config  → 403 Forbidden
```

### Admin Capabilities
```bash
# ✅ All user capabilities PLUS:

# ✅ Modify global configuration
POST /api/v1/config
{
  "agents": {
    "defaults": {
      "max_tool_iterations": 30
    }
  }
}

# ✅ Manage users
GET  /api/v1/users               # List all users
PUT  /api/v1/users/{id}          # Update user role/password
DELETE /api/v1/users/{id}        # Delete user
```

## Configuration Sections

### Global Config (Admin-Only via `/api/v1/config`)
```json
{
  "agents": { /* Default agent behavior */ },
  "providers": { /* Provider endpoints (keys per-user) */ },
  "channels": { /* Telegram, Discord, etc. */ },
  "tools": { /* Filesystem, shell, web restrictions */ },
  "gateway": { /* Gateway settings */ },
  "web": { /* Web server config */ },
  "storage": { /* Database config */ }
}
```
**Storage**: `~/.KakoClaw/config.json`

### Per-User Config

**Providers (DB)**: `user_providers_config` table
- Anthropic, OpenAI, OpenRouter, Groq, Zhipu, VLLM, Gemini, Nvidia, Moonshot, Ollama
- API keys are **redacted** in responses: `"sk-t****-123"`

**Overlays (File)**: `~/.kakoclaw/users/<uuid>/config.json`
- Personal `agents`, `channels`, `tools` overrides
- Merged over global defaults at runtime

## Quick Tests

```bash
# Test basic permissions
/tmp/test_config_endpoints.sh

# Test admin permissions
/tmp/test_complete_permissions.sh
```

## Common Tasks

### Promote User to Admin
```bash
# Via database (requires backend restart or re-login)
sqlite3 ~/.KakoClaw/KakoClaw.db \
  "UPDATE users SET role='admin' WHERE username='alice';"

# Via API (requires existing admin token)
curl -X PUT "http://localhost:18880/api/v1/users/{id}" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role":"admin"}'
```

### Reset User Password
```bash
# Via API (admin only)
curl -X PUT "http://localhost:18880/api/v1/users/{id}" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"password":"NewPass123!"}'
```

### Set Personal API Key
```bash
# Via web UI: Settings → Providers → OpenAI → Enter key

# Via API:
curl -X PUT "http://localhost:18880/api/v1/me/providers/update?provider=openai" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"api_key":"sk-proj-..."}'
```

## Security Notes

✅ **Isolation**: Each user's provider API keys are stored separately  
✅ **Redaction**: API keys never appear in full in GET responses  
✅ **No Fallback**: New users see empty configs (no global contamination)  
✅ **Role-Based**: Only admin can modify global config  
✅ **JWT Claims**: Role included in auth token (`"admin"` | `"user"`)

## Documentation

- **Full Guide**: [docs/development/CONFIG_PERMISSIONS.md](./CONFIG_PERMISSIONS.md)
- **Implementation**: [docs/development/IMPLEMENTATION_CONFIG_PERMISSIONS.md](./IMPLEMENTATION_CONFIG_PERMISSIONS.md)
