# Configuration Permissions Matrix

This document defines which configuration sections can be modified by different user roles.

## API Endpoints

| Endpoint | Method | User Role | Admin Role | Purpose |
|----------|--------|-----------|------------|---------|
| `/api/v1/config` | GET | Read global defaults (redacted) | Read global config (redacted) | View system configuration |
| `/api/v1/config` | POST | ❌ **403 Forbidden** | ✅ **Allowed** | Modify global configuration |
| `/api/v1/me/config` | GET | ✅ **Allowed** | ✅ **Allowed** | View merged user config |
| `/api/v1/me/config/update` | POST | ✅ **Allowed** | ✅ **Allowed** | Update user-specific config sections |
| `/api/v1/me/providers` | GET | ✅ **Allowed** | ✅ **Allowed** | View user provider configurations |
| `/api/v1/me/providers/update` | PUT | ✅ **Allowed** | ✅ **Allowed** | Update user provider API keys |

## Configuration Sections

### Global Configuration (Admin-Only via `/api/v1/config`)

When an admin POSTs to `/api/v1/config`, they can modify:

- **`agents`** - Default agent behavior, workspace settings, iteration limits
- **`providers`** - Default provider API endpoints and models (keys should be per-user)
- **`channels`** - Channel adapters (Telegram, Discord, Slack, etc.) and their auth tokens
- **`tools`** - Tool-specific settings (filesystem restrictions, shell policies, etc.)
- **`gateway`** - Gateway operational parameters
- **`web`** - Web server configuration (port, auth timeouts)
- **`storage`** - Database path and settings

**Persistence**: Changes are saved to `~/.KakoClaw/config.json` (or `$KAKOCLAW_CONFIG_PATH`)

### User-Specific Configuration (All Users via `/api/v1/me/*`)

#### Per-User Providers (Database-Backed)
**Storage**: SQLite table `user_providers_config` (separate from config.json)

Users can configure their own API keys for:
- Anthropic (Claude)
- OpenAI (GPT)
- OpenRouter
- Groq
- Zhipu (GLM)
- VLLM
- Gemini
- Nvidia
- Moonshot
- Ollama

Each provider config includes:
- `api_key` - User's personal API key (redacted in responses with `****`)
- `api_base` - Custom API endpoint URL (optional)
- `proxy` - HTTP proxy URL (optional)
- `auth_method` - Authentication method (default: "bearer")
- `models` - Available models list (optional)

**Endpoint**: `PUT /api/v1/me/providers/update?provider=<name>`

#### Per-User Config Overlays (File-Backed)
**Storage**: `~/.kakoclaw/users/<user_id>/config.json`

Users can override sections of the global config:
- **`agents`** - Personal workspace preferences, model choices
- **`channels`** - Personal channel configurations
- **`tools`** - Personal tool settings

**Merging**: User config sections overlay global defaults via `config.MergeConfigs()`. User-specific settings take precedence.

**Endpoint**: `POST /api/v1/me/config/update`

## Runtime Behavior

### Agent Loops
When creating an agent loop for a user:
1. Load global config from `~/.KakoClaw/config.json`
2. Load user config from `~/.kakoclaw/users/<user_id>/config.json`
3. Merge user config over global (section-level overlay)
4. Load user providers from `user_providers_config` DB table
5. Use merged config + user providers for agent execution

### Channel Access Control
Channels have `allow_from` lists (user IDs/usernames) to restrict who can use them. Only admin can modify these lists via `/api/v1/config`.

## Security Notes

1. **API Key Isolation**: Users never see each other's API keys. Provider configs are stored per-user in DB.
2. **Global Config Protection**: Only admin role can POST to `/api/v1/config`. Regular users receive 403.
3. **API Key Redaction**: Provider API keys are always redacted in GET responses (replaced with `****` pattern).
4. **No Fallback Contamination**: New users see empty provider configs, not global defaults.

## Testing

Run smoke tests with:
```bash
/tmp/test_config_endpoints.sh
```

Expected results:
- ✅ User can GET `/api/v1/me/config` (merged config with redacted keys)
- ✅ User can GET `/api/v1/me/providers` (empty for new users)
- ✅ User can PUT to `/api/v1/me/providers/update` (save personal API key)
- ❌ User cannot POST to `/api/v1/config` (403 Forbidden)
- ✅ Admin can POST to `/api/v1/config` (save global config)

## Implementation Files

- **Auth**: `pkg/web/auth.go` (JWT claims with role)
- **Storage**: `pkg/storage/user_providers.go` (per-user provider DB)
- **Handlers**: `pkg/web/handlers_advanced.go` (global config), `pkg/web/handlers_user_config.go` (user config)
- **Config Merge**: `pkg/config/config.go` (`MergeConfigs` function)
- **Middleware**: `pkg/web/server.go` (`authMiddleware` extracts JWT claims with role)
