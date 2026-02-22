# Phase 4: Channel Auto-Onboarding Implementation

**Status**: ✅ COMPLETE & TESTED  
**Commit**: `d777c96`  
**Code Lines**: 847 total  
**Build Time**: 2.87s  
**Binary Size**: 31MB  

## Overview

Phase 4 enables users to initiate the setup/onboarding flow directly from their messaging platforms (Telegram, Discord, Slack) using the `/setup` command. This creates a seamless, mobile-friendly onboarding experience via QR codes and deep linking.

## Backend Implementation

### 1. **Storage Layer** (`storage/setup_session.go` - 118 lines)

New data model for temporary setup sessions:

```go
type SetupSession struct {
    ID        int64
    Token     string           // Unique 32-byte random token
    UserID    int64            // Associated user after completion
    Channel   string           // "telegram", "discord", "slack"
    SenderID  string           // External channel user ID
    Metadata  map[string]string
    ExpiresAt time.Time        // 1 hour TTL
    CreatedAt time.Time
    UsedAt    *time.Time       // When setup completed
}
```

**Methods**:
- `CreateSetupSession()` - Generate new token with 1-hour expiry
- `GetSetupSession()` - Retrieve valid (non-expired) session
- `UpdateSetupSession()` - Mark complete and associate with user
- `CleanupExpiredSetupSessions()` - Maintenance task

**Database**:
- Added `setup_sessions` table with indices on `token` and `expires_at`
- Part of standard migration in `sqlite.go`

### 2. **API Handlers** (`web/handlers_setup.go` - 158 lines)

Three new REST endpoints:

#### `POST /api/v1/setup/initialize`
Generates a new setup token and returns setup URL.

```json
Request:
{
  "channel": "telegram",
  "sender_id": "123456789",
  "metadata": {"source": "setup_wizard"}
}

Response:
{
  "token": "abc123...",
  "expires_at": "2026-02-21T15:49:00Z",
  "setup_url": "http://localhost:8080/onboarding?token=abc123..."
}
```

#### `GET /api/v1/setup/validate/{token}`
Validates token and returns associated metadata (no auth required).

```json
Response:
{
  "valid": true,
  "channel": "telegram",
  "sender_id": "123456789",
  "metadata": {...}
}
```

#### `POST /api/v1/setup/complete/{token}`
Finalizes setup and associates token with authenticated user (auth required).

```json
Response:
{
  "success": true,
  "message": "Setup completed successfully"
}
```

### 3. **Command Handler** (`channels/command_handler.go` - 89 lines)

Generic command processor for all channels:

```go
type CommandHandler struct {
    store *storage.Storage
}

// HandleCommand(ctx, channel, senderID, content) -> (handled, response, error)
```

**Supported Commands**:
- `/setup` - Generates setup token, displays URL with QR code
- `/status` - Returns bot status message

**Features**:
- Automatic token generation on `/setup` command
- Configurable command parsing (handles edge cases)
- Error logging and user-friendly responses

### 4. **Channel Integration**

#### Telegram (`channels/telegram.go`)
```go
// Added fields to TelegramChannel
type TelegramChannel struct {
    ...
    commandHandler *CommandHandler
}

// New method
func (c *TelegramChannel) SetCommandHandler(handler *CommandHandler) {
    c.commandHandler = handler
}

// In handleMessage():
if c.commandHandler != nil && IsCommand(content) {
    handled, response, err := c.commandHandler.HandleCommand(ctx, "telegram", senderID, content)
    if handled {
        _, _ = c.bot.SendMessage(ctx, tu.Message(tu.ID(chatID), response))
        return  // Don't process as regular message
    }
}
```

#### Discord (`channels/discord.go`)
Similar integration with Discord's MessageCreate event handler.

#### Slack (`channels/slack.go`)
Similar integration with Slack's message event handler, with reaction acknowledgment.

### 5. **Channel Interface Update** (`channels/base.go`)

Updated the Channel interface to support command handlers:

```go
type Channel interface {
    ...
    SetCommandHandler(*CommandHandler) // New method
}

// Base implementation: no-op
func (c *BaseChannel) SetCommandHandler(handler *CommandHandler) {
    // Default: no-op - subclasses override
}
```

## Frontend Implementation

### 1. **QR Code Component** (`components/QRCode.vue` - 118 lines)

Standalone Vue 3 component for QR code generation:

```vue
<QRCode url="https://kakoclaw.app/onboarding?token=abc123..." />
```

**Features**:
- Uses `qrcode` npm package for generation
- Canvas-based rendering
- Displays both QR code and clickable link
- Responsive dark theme (Tailwind CSS)
- Auto-regenerates on URL change

**Styling**:
- Dark slate background: `bg-slate-700/30`
- Blue accent for QR: `border-blue-500`
- Mobile-friendly layout

### 2. **Channel Form Enhancement** (`components/Setup/ChannelForm.vue`)

Added "Quick Setup" section with QR code generation:

```vue
<!-- Quick Setup via Channel -->
<div v-if="selectedChannel" class="bg-slate-700/30 rounded-lg p-6">
  <section>
    <button @click="generateSetupToken">
      📱 Generate Setup Link for {{ selectedChannel.name }}
    </button>
    <QRCode v-if="setupToken" :url="setupUrl" />
    <button @click="copySetupUrl">📋 Copy Link</button>
  </section>
</div>
```

**Functions**:
- `generateSetupToken()` - POST to `/api/v1/setup/initialize`
- `copySetupUrl()` - Copy token URL to clipboard
- `resetSetupToken()` - Clear and regenerate

**User Experience**:
- Single click to generate setup token
- QR code appears immediately
- Copy-to-clipboard feedback
- Regenerate button for new tokens
- Still supports manual token input below

### 3. **NPM Dependency**
```bash
npm install qrcode --save
```
Adds `qrcode@^1.x` to project dependencies.

## Workflow

### For End Users

1. **User opens app in Telegram/Discord/Slack**
   ```
   User: /setup
   ```

2. **Bot responds with setup link + QR**
   ```
   Bot: 🚀 Welcome to KakoClaw Setup!
        Please visit this link: https://localhost:8080/onboarding?token=abc123...
        ⏱️ This link expires in 1 hour
   ```

3. **User scans QR or clicks link** → Redirected to `/onboarding?token=abc123...`

4. **Frontend validates token** → Pre-fills channel selection with detected platform

5. **User completes wizard** → Setup token is finalized with POST `/api/v1/setup/complete/{token}`

6. **Session complete** → Token marked as used, associated with user account

### For Developers

```go
// Example: Integrate command handler into channel setup
commandHandler := channels.NewCommandHandler(storage)
telegramChannel.SetCommandHandler(commandHandler)
```

## Security Considerations

1. **Token Expiry**: 1-hour TTL prevents long-lived setup links
2. **One-time use**: Tokens marked as `used_at` after completion
3. **Channel binding**: Tokens track which channel initiated setup
4. **User association**: Token linked to authenticated user on completion
5. **Input validation**: Command parsing handles malformed input gracefully

## Database Schema

```sql
CREATE TABLE setup_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT NOT NULL UNIQUE,
    user_id INTEGER,
    channel TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    metadata TEXT NOT NULL DEFAULT '{}',
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    used_at DATETIME
);

CREATE INDEX idx_setup_sessions_token ON setup_sessions(token);
CREATE INDEX idx_setup_sessions_expires_at ON setup_sessions(expires_at);
```

## Testing Checklist

- [x] Backend compiles without errors
- [x] Frontend compiles with Vite (2.87s)
- [x] `/api/v1/setup/initialize` returns valid token URLs
- [x] QR code renders correctly in browser
- [x] Copy-to-clipboard functionality works
- [x] Token validation logic handles expired tokens
- [x] Command parsing correctly identifies `/setup`
- [x] All three channels (Telegram, Discord, Slack) can use `/setup`
- [x] Database migration creates `setup_sessions` table
- [x] Binary builds successfully (31MB)

## What's Next (Phase 5 Ideas)

1. **Auto-linking**: Use `/setup` responses to automatically generate channel-specific callback handlers
2. **QR deepening**: Redirect QR code to mobile app instead of web
3. **Setup analytics**: Track setup completion rates by channel
4. **Retry logic**: Allow expired tokens to be regenerated
5. **Multi-user setup flow**: Different setup flows for admin vs. regular users

## Files Changed

### Backend
- `pkg/storage/setup_session.go` (NEW)
- `pkg/storage/sqlite.go` (MODIFIED: added migration)
- `pkg/web/handlers_setup.go` (NEW)
- `pkg/web/server.go` (MODIFIED: registered routes)
- `pkg/channels/command_handler.go` (NEW)
- `pkg/channels/base.go` (MODIFIED: added interface method)
- `pkg/channels/telegram.go` (MODIFIED: integrated handler)
- `pkg/channels/discord.go` (MODIFIED: integrated handler)
- `pkg/channels/slack.go` (MODIFIED: integrated handler)

### Frontend
- `pkg/web/frontend/src/components/QRCode.vue` (NEW)
- `pkg/web/frontend/src/components/Setup/ChannelForm.vue` (MODIFIED)
- `package.json` (MODIFIED: added qrcode)
- `package-lock.json` (MODIFIED)

## Build Information

```
Frontend Build:
- Bundle: 775KB gzipped
- Modules: 245
- Time: 2.87s
- Status: ✅ Success

Backend Build:
- Binary: 31MB (darwin/arm64)
- Version: d777c96-dirty
- Go: go1.26.0
- Status: ✅ Success
```

## Commit

```
d777c96 Phase 4: Channel Auto-Onboarding with /setup commands - 847 lines
```

Pushed to: `origin/multiuser`

---

**Implementation Date**: February 21, 2026  
**Author**: GitHub Copilot  
**Status**: ✅ Production Ready
