# Onboarding Wizard Guide

## Overview

MakoClaw includes a comprehensive onboarding wizard that guides new users through initial setup. This wizard is automatically shown on first login after registration and helps configure:

1. **LLM Provider** (required) - OpenAI, Anthropic, Groq, Ollama, etc.
2. **Workspace Setup** (optional) - Skills and example files
3. **Communication Channels** (optional) - Telegram, Discord, Slack, etc.

## User Experience Flow

### First-Time Login

When a user registers and logs in for the first time:

1. User completes registration at `/signup`
2. Logs in at `/login`
3. Router guard detects `onboarding_completed = false`
4. Automatically redirects to `/onboarding`
5. User proceeds through 6-step wizard
6. Upon completion, `onboarding_completed` is set to `true`
7. User is redirected to dashboard

### Wizard Steps

#### Step 0: Welcome

- Introduces the setup process
- Lists what will be configured
- No configuration required

#### Step 1: AI Provider (Required)

- Choose provider: OpenAI, Anthropic, Groq, OpenRouter, Ollama, etc.
- Enter API key (or configure local endpoints for Ollama)
- Select model
- **Cannot skip** - provider is required for AI functionality

#### Step 2: Workspace (Optional)

- Select skills to install:
  - `github` - GitHub integration tools
  - `summarize` - Document summarization
  - `weather` - Weather information
  - `skill-creator` - Create new skills
  - `development` - Development tools
  - `tmux` - Terminal multiplexer integration
- Option to include example files (`WELCOME.md`, `NOTES.md`)
- **Can skip** - workspace customization is optional
- "Use Recommended Defaults" button selects: github, summarize, skill-creator + example files

#### Step 3: Channel (Optional)

- Configure communication channels
- Supports: Telegram, Discord, Slack, WhatsApp, Signal, QQ, DingTalk, Feishu, MaixCam
- Enter bot token and channel ID
- **Can skip** - channels can be configured later

#### Step 4: Preview

- Review all configurations before saving
- Edit any step by clicking on it
- Shows provider, workspace, and channel settings

#### Step 5: Success

- Confirmation screen
- Displays next steps
- Button to go to dashboard

## Degraded Mode vs Onboarding

MakoClaw has two separate setup flows:

### Onboarding Wizard (`/onboarding`)

- **Purpose**: First-time user comprehensive setup
- **Trigger**: Automatic redirect when `onboarding_completed = false`
- **Steps**: 6-step wizard (Welcome → Provider → Workspace → Channel → Preview → Success)
- **Features**:
  - Workspace customization with skills selection
  - Optional channel setup
  - Marks `onboarding_completed = true` when finished
- **View**: `OnboardingView.vue` → `SetupWizard.vue` (6 steps)

### Degraded Mode Setup (`/setup`)

- **Purpose**: Quick LLM provider configuration when missing
- **Trigger**: User clicks "Configure Now" button in degraded mode banner
- **Steps**: 2-step quick setup (Choose Provider → Configure)
- **Features**:
  - Provider configuration only
  - No workspace or channel setup
  - Does NOT mark onboarding complete
- **View**: `SetupView.vue` (2 steps)

The degraded mode banner:

- Only shows when provider is not configured AND onboarding is already complete
- Hidden during first-login onboarding (router redirects to `/onboarding` instead)

## Technical Implementation

### Database Schema

```sql
ALTER TABLE users ADD COLUMN onboarding_completed BOOLEAN NOT NULL DEFAULT 0;
```

Fields:

- `onboarding_completed` - Boolean flag tracking if user has completed onboarding

### API Endpoints

#### GET `/api/v1/auth/me`

Returns current user info including onboarding status:

```json
{
  "username": "alice",
  "role": "user",
  "user_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "onboarding_completed": false
}
```

#### POST `/api/v1/me/onboarding/complete`

Marks onboarding as completed for authenticated user.

Request: Empty body (uses JWT token for user identification)

Response:

```json
{
  "success": true,
  "message": "Onboarding completed successfully"
}
```

#### POST `/api/v1/me/workspace/init`

Installs selected skills and example files to user workspace.

Request:

```json
{
  "skills": ["github", "summarize", "skill-creator"],
  "exampleFiles": true
}
```

Response:

```json
{
  "success": true,
  "installed": {
    "skills": ["github", "summarize", "skill-creator"],
    "files": ["WELCOME.md", "NOTES.md"]
  }
}
```

### Frontend Architecture

#### Stores

- **`onboardingStore.js`**: Manages onboarding state
  - State: `currentStep`, `onboardingCompleted`, `isFirstLogin`
  - Getters: `needsOnboarding()`, `canSkipCurrentStep()`
  - Actions: `checkOnboardingStatus()`, `completeOnboarding()`, `skipStep()`

#### Components

- **`SetupWizard.vue`**: 6-step wizard orchestrator
  - Manages navigation between steps
  - Integrates child components for each step
  - Handles save and completion logic

- **`WorkspaceSetupForm.vue`**: Workspace customization component
  - Skills selection with checkboxes
  - Example files toggle
  - "Use Recommended Defaults" button
  - Live preview of selections

- **`ProviderForm.vue`**: LLM provider configuration
- **`ChannelForm.vue`**: Communication channel setup
- **`ConfigPreview.vue`**: Configuration review

#### Router Guard

```javascript
router.beforeEach(async (to, from, next) => {
  // ... auth checks ...

  if (authStore.isAuthenticated && onboardingStore.needsOnboarding) {
    if (to.name !== "onboarding") {
      next("/onboarding");
      return;
    }
  }

  next();
});
```

### Backend Implementation

#### User Struct

```go
type User struct {
    // ... existing fields ...
    OnboardingCompleted bool `json:"onboarding_completed"`
}
```

#### Storage Methods

- `CentralStorage.MarkOnboardingCompleted(userUUID string) error`
- `CentralStorage.GetUserByUUID(uuid string) (*User, error)`

#### Web Handlers

- `handleMe()` - Returns user info with onboarding status
- `handleCompleteOnboarding()` - Marks onboarding complete
- `handleWorkspaceInit()` - Installs skills and example files

## Skills Installation

Built-in skills are copied from the repository `skills/` directory to user workspace:

```
~/.MakoClaw/users/{uuid}/workspace/skills/
  ├── github/
  │   └── SKILL.md
  ├── summarize/
  │   └── SKILL.md
  └── skill-creator/
      └── SKILL.md
```

Example files created:

- `WELCOME.md` - Getting started guide
- `NOTES.md` - Template for notes and ideas

## Configuration

No additional configuration is required. The onboarding wizard is automatically enabled once the database migration adds the `onboarding_completed` column.

### Disabling Onboarding (if needed)

To manually mark a user as onboarded:

```sql
UPDATE users SET onboarding_completed = 1 WHERE username = 'alice';
```

Or via web UI:

1. Complete the onboarding wizard normally
2. Or in Settings, manually configure provider and channels

## Testing

### Manual Testing Flow

1. **Register new user**

   ```
   POST /api/v1/auth/signup
   {"username": "test_user", "password": "test_pass"}
   ```

2. **Login**

   ```
   POST /api/v1/auth/login
   {"username": "test_user", "password": "test_pass"}
   ```

3. **Verify auto-redirect**
   - Access `/dashboard` or any authenticated route
   - Should redirect to `/onboarding`

4. **Complete wizard**
   - Step 0: Click "Next"
   - Step 1: Select provider, enter API key, click "Next"
   - Step 2: Select skills (or skip)
   - Step 3: Configure channel (or skip)
   - Step 4: Review and click "Next"
   - Step 5: Click "Go to Dashboard"

5. **Verify completion**

   ```
   GET /api/v1/auth/me
   ```

   Should return `"onboarding_completed": true`

6. **Verify no more redirects**
   - Access `/dashboard` - should stay on dashboard
   - Verify degraded mode banner shows if provider configured incorrectly

### Database Verification

Check onboarding status in database:

```sql
SELECT username, onboarding_completed FROM users;
```

### Skills Verification

Check installed skills:

```bash
ls ~/.MakoClaw/users/{uuid}/workspace/skills/
```

## Troubleshooting

### Redirect Loop

**Symptom**: Browser keeps redirecting between `/onboarding` and `/dashboard`

**Cause**: Router guard logic error or onboarding store not initialized

**Fix**:

- Check browser console for errors
- Verify `onboardingStore.checkOnboardingStatus()` is called in `App.vue`
- Check that `needsOnboarding` getter returns correct value

### Skills Not Installing

**Symptom**: Workspace step completes but skills not in workspace

**Cause**: Incorrect skills path or permission issues

**Fix**:

- Check backend logs for skill installation errors
- Verify built-in skills exist in repository `/skills/` directory
- Check user workspace permissions: `~/.MakoClaw/users/{uuid}/workspace/`

### Onboarding Completed But Provider Still Missing

**Symptom**: Onboarding complete, but degraded mode banner still shown

**Cause**: This is expected behavior - onboarding completion doesn't guarantee valid provider

**Fix**:

- User should click "Configure Now" in degraded mode banner
- This opens quick `/setup` flow to fix provider configuration
- Or manually configure provider in Settings

### Cannot Skip Workspace Step

**Symptom**: Skip button not showing on Workspace step

**Cause**: Logic error in `canSkipCurrentStep` computed property

**Fix**:

- Check `SetupWizard.vue` - step 2 (Workspace) should have `canSkipCurrentStep = true`
- Verify `v-if="canSkipCurrentStep"` on Skip button

## Future Enhancements

Potential improvements for future versions:

1. **Wizard Restart**: Allow users to re-run onboarding from Settings
2. **Partial Completion**: Save progress and allow resuming later
3. **Multi-Provider**: Support configuring multiple providers in wizard
4. **Custom Skills**: Allow selecting custom skills from marketplace
5. **Team Onboarding**: Invite team members with pre-configured settings
6. **Analytics**: Track which steps users skip most often
7. **Localization**: Translate wizard to multiple languages

## Related Documentation

- [Quick Start Guide](quickstart.md)
- [Installation Guide](installation.md)
- [Skills Documentation](../development/AGENTS.md)
- [Multi-Agent Setup](../development/MULTI_AGENT_SETUP.md)
- [Configuration Guide](../development/CONFIG_ARCHITECTURE.md)
