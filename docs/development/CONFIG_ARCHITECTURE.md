# Configuration Permission Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        KakoClaw Config System                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌───────────────────────────┐  ┌──────────────────────────┐   │
│  │   Global Config (Admin)   │  │   Per-User Config (All)  │   │
│  │                           │  │                          │   │
│  │  ~/.KakoClaw/config.json  │  │  Database + User Files   │   │
│  │                           │  │                          │   │
│  │  • agents.defaults        │  │  • user_providers_config │   │
│  │  • providers.defaults     │  │    (SQLite table)        │   │
│  │  • channels.all           │  │                          │   │
│  │  • tools.settings         │  │  • ~/.kakoclaw/users/    │   │
│  │  • gateway.config         │  │    {uuid}/config.json    │   │
│  │  • web.settings           │  │                          │   │
│  │  • storage.config         │  │  Overlays: agents,       │   │
│  │                           │  │  channels, tools         │   │
│  │  ⚠️  Admin POST only       │  │                          │   │
│  └───────────────────────────┘  └──────────────────────────┘   │
│           │                                    │                 │
│           └────────────┬───────────────────────┘                │
│                        ↓                                         │
│              ┌──────────────────┐                               │
│              │  config.MergeConfigs()  │                        │
│              │  Section-level overlay  │                        │
│              └──────────────────┘                               │
│                        ↓                                         │
│              ┌──────────────────┐                               │
│              │  Runtime Config  │                               │
│              │  for Agent Loop  │                               │
│              └──────────────────┘                               │
└─────────────────────────────────────────────────────────────────┘
```

## API Permission Matrix

```
                           ┌──────────┬────────┐
                           │   User   │ Admin  │
┌──────────────────────────┼──────────┼────────┤
│ GET /api/v1/config       │    ✅    │   ✅   │  View global (redacted)
│ POST /api/v1/config      │    ❌    │   ✅   │  Modify global
├──────────────────────────┼──────────┼────────┤
│ GET /api/v1/me/config    │    ✅    │   ✅   │  View merged config
│ POST /api/v1/me/config   │    ✅    │   ✅   │  Update user config
├──────────────────────────┼──────────┼────────┤
│ GET /api/v1/me/providers │    ✅    │   ✅   │  View user providers
│ PUT /api/v1/me/providers │    ✅    │   ✅   │  Update provider keys
├──────────────────────────┼──────────┼────────┤
│ GET /api/v1/users        │    ❌    │   ✅   │  List all users
│ PUT /api/v1/users/{id}   │    ❌    │   ✅   │  Update user role
│ DELETE /api/v1/users/{id}│    ❌    │   ✅   │  Delete user
└──────────────────────────┴──────────┴────────┘
```

## Request Flow

### Regular User Request
```
┌─────────┐
│ Browser │ Authorization: Bearer eyJ... (role="user")
└────┬────┘
     │ GET /api/v1/me/config
     ↓
┌────────────────┐
│ authMiddleware │ Extract JWT → claims.Role = "user"
└────┬───────────┘
     │ Context: userClaimsKey → claims
     ↓
┌────────────────────┐
│ handleGetUserConfig│
└────┬───────────────┘
     │ 1. Load global: ~/.KakoClaw/config.json
     │ 2. Load user DB: user_providers_config (user_id=X)
     │ 3. Load user file: ~/.kakoclaw/users/X/config.json
     │ 4. config.MergeConfigs(global, user)
     │ 5. Redact API keys
     ↓
┌────────┐
│ JSON   │ {"config": {...redacted...}}
└────────┘
```

### Admin Global Config Update
```
┌─────────┐
│ Browser │ Authorization: Bearer eyJ... (role="admin")
└────┬────┘
     │ POST /api/v1/config {"agents":{"defaults":{...}}}
     ↓
┌────────────────┐
│ authMiddleware │ Extract JWT → claims.Role = "admin"
└────┬───────────┘
     │ Context: userClaimsKey → claims
     ↓
┌──────────────┐
│ handleConfig │
└────┬─────────┘
     │ if claims.Role != "admin" → 403 Forbidden
     │ ✅ Role = "admin", proceed
     │
     │ 1. Decode JSON body
     │ 2. updateAgentsConfig(fullConfig, body["agents"])
     │ 3. config.SaveConfig(~/.KakoClaw/config.json)
     │ 4. restart affected channels
     ↓
┌────────┐
│ 200 OK │ {"config": {...}}
└────────┘
```

### User Attempts Global Update (Blocked)
```
┌─────────┐
│ Browser │ Authorization: Bearer eyJ... (role="user")
└────┬────┘
     │ POST /api/v1/config {"agents":{"defaults":{...}}}
     ↓
┌────────────────┐
│ authMiddleware │ Extract JWT → claims.Role = "user"
└────┬───────────┘
     │ Context: userClaimsKey → claims
     ↓
┌──────────────┐
│ handleConfig │
└────┬─────────┘
     │ if claims.Role != "admin" → ✋ STOP
     ↓
┌────────────────┐
│ 403 Forbidden  │ "forbidden: admin role required"
└────────────────┘
```

## Database Schema

### `users` Table
```sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uuid TEXT NOT NULL UNIQUE,
  username TEXT NOT NULL UNIQUE,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'user',  -- 'admin' | 'user'
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### `user_providers_config` Table
```sql
CREATE TABLE user_providers_config (
  user_id INTEGER PRIMARY KEY,
  config TEXT NOT NULL,  -- JSON: {anthropic:{api_key:"..."}, openai:{...}, ...}
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

## File System Layout

```
~/.KakoClaw/
├── config.json                        # Global config (admin-managed)
├── KakoClaw.db                        # SQLite (users, providers, sessions)
├── workspace/                         # Shared workspace files
└── users/
    ├── {user1-uuid}/
    │   ├── config.json                # User1 config overlay
    │   └── workspace/                 # User1 workspace (optional)
    └── {user2-uuid}/
        ├── config.json                # User2 config overlay
        └── workspace/                 # User2 workspace (optional)
```

## Security Boundaries

```
┌─────────────────────────────────────────────────┐
│              Security Boundaries                 │
├─────────────────────────────────────────────────┤
│                                                  │
│  1. JWT Role Claim                              │
│     ✅ Embedded in signed token                 │
│     ✅ Verified via HMAC signature              │
│     ✅ Extracted by authMiddleware              │
│                                                  │
│  2. Database Isolation                          │
│     ✅ user_providers_config.user_id foreign key│
│     ✅ Each row = one user's credentials        │
│     ✅ No cross-user reads                      │
│                                                  │
│  3. API Key Redaction                           │
│     ✅ redactUserProviders(config)              │
│     ✅ Pattern: "sk-t****-123"                  │
│     ✅ Applied before all GET responses         │
│                                                  │
│  4. File Permissions                            │
│     ✅ ~/.kakoclaw/users/{uuid}/ readable by    │
│        process owner only                       │
│     ✅ config.json mode 0600                     │
│                                                  │
│  5. Admin Role Check                            │
│     ✅ if claims.Role != "admin" → 403          │
│     ✅ Applied before any global config write   │
│     ✅ Cannot be bypassed via JWT forgery       │
│                                                  │
└─────────────────────────────────────────────────┘
```

## Testing Strategy

```
┌──────────────────────────────────────────┐
│          Test Coverage                    │
├──────────────────────────────────────────┤
│                                           │
│  Unit Tests:                              │
│    • JWT generation/verification          │
│    • Config merge logic                   │
│    • API key redaction                    │
│    • Payload normalization                │
│                                           │
│  Integration Tests:                       │
│    ✅ /tmp/test_config_endpoints.sh       │
│       - User config CRUD                  │
│       - Provider isolation                │
│       - 403 on global config              │
│                                           │
│    ✅ /tmp/test_complete_permissions.sh   │
│       - User vs admin role checks         │
│       - Global config modification        │
│       - Change persistence                │
│       - Cross-user visibility             │
│                                           │
│  E2E Tests:                               │
│    • Setup Wizard flow                    │
│    • Runtime agent execution              │
│    • Multi-user concurrent access         │
│                                           │
└──────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Database for Providers, File for Config Overlays
**Why**: Provider configs change frequently (API key rotation), config overlays are rare (personal preferences). Database enables atomic updates, files enable version control.

### 2. Section-Level Merge (Not Field-Level)
**Why**: Simpler logic, predictable behavior. If user specifies `agents.defaults`, entire section overrides global.

### 3. Redaction Always On (Even for Token Owner)
**Why**: Prevents accidental exposure in logs, screenshots, support requests. Original keys remain in DB.

### 4. JWT Role in Claims (Not Database Lookup)
**Why**: Performance - no DB query per request. Role changes require re-login (acceptable UX).

### 5. Admin-Only Global Config (No "Power User" Tier)
**Why**: Binary permission model is easier to reason about and secure. Can extend with granular permissions later.

## Migration Path

```
Old System (Pre-RBAC):              New System (Post-RBAC):
┌─────────────────────┐             ┌─────────────────────┐
│ config.json         │             │ Global config.json  │
│   - Used by all     │    ───►     │   - Admin only      │
│   - Shared keys     │             │   - No API keys     │
│   - No isolation    │             └─────────────────────┘
└─────────────────────┘                        │
                                               │ Fork per user
                                               ↓
                                    ┌─────────────────────┐
                                    │ Per-User DB/Files   │
                                    │   - Own API keys    │
                                    │   - Personal prefs  │
                                    │   - Isolated        │
                                    └─────────────────────┘
```

### Migration Steps
1. ✅ Create `user_providers_config` table
2. ✅ Remove placeholder keys from global config
3. ✅ Add role checks to POST `/api/v1/config`
4. ✅ Update frontend to use `/api/v1/me/*` endpoints
5. ✅ Users re-enter API keys via Setup Wizard
6. ✅ Test isolation with multiple users
7. ✅ Document new permission model

## Future Extensions

- [ ] Granular permissions (per-section admin)
- [ ] Audit log for config changes
- [ ] Config versioning / rollback
- [ ] Provider usage quotas per user
- [ ] Team/organization support
- [ ] LDAP/SAML integration

---

**Version**: 1.0 (2026-02-21)  
**Status**: ✅ Production Ready  
**Tests**: All Passing
