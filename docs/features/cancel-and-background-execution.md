# Cancel and Background Execution Feature

## Overview

KakoClaw now supports canceling agent executions mid-task and running agents in the background while navigating through the web panel. This provides better control and a more flexible user experience.

## Features

### 1. Cancel Execution

**Description:** Stop an agent mid-execution without waiting for completion.

**How it works:**
- Each WebSocket chat session creates a cancelable context (`context.WithCancel`)
- Active executions are tracked in a thread-safe map with their session ID and cancel function
- When the cancel button is clicked, the API endpoint triggers `context.Cancel()`
- The agent loop respects the context and stops processing

**Frontend:**
- Cancel button appears next to the send button when `isLoading` is true
- Clicking cancel calls `/api/v1/chat/cancel` with the current session ID
- Toast notification confirms cancellation

**Backend:**
- `activeExecution` struct tracks session ID, start time, and cancel function
- `Server.activeExecs` map stores all running executions
- `handleChatCancel` endpoint cancels all executions for a given session
- Context cancellation propagates through the agent loop

### 2. Background Execution

**Description:** Navigate away from the chat view while the agent continues working.

**How it works:**
- Agent execution is decoupled from WebSocket connection lifecycle
- The `onBeforeUnmount` hook in ChatView removes event listeners but does NOT disconnect WebSocket
- Active executions continue processing in the background
- Users can navigate to Tasks, Settings, or other views without interruption

**Implementation:**
```javascript
onBeforeUnmount(() => {
  // Remove listeners to prevent duplicates, but DON'T disconnect
  // This allows the agent to continue working even when navigating away from chat
  chatWs.off('message', handleMessage)
  chatWs.off('disconnected', handleDisconnected)
  chatWs.off('connected', handleConnected)
})
```

### 3. Active Executions Monitoring (Optional)

**Composable:** `useActiveExecutions`

**Description:** Provides reactive state for monitoring all active agent executions across sessions.

**Usage:**
```javascript
import { useActiveExecutions } from '@/composables/useActiveExecutions'

const { activeExecutions, isPolling, refresh } = useActiveExecutions({
  pollInterval: 3000 // Optional, default 3s
})

// activeExecutions.value = [
//   { id: "web:chat:1234567890", session_id: "web:chat", started_at: "...", duration: "..." }
// ]
```

**Features:**
- Auto-starts polling on mount
- Auto-stops polling on unmount
- Provides `refresh()` method for manual updates
- Configurable poll interval

## API Endpoints

### POST /api/v1/chat/cancel

Cancel a running agent execution.

**Request:**
```json
{
  "session_id": "web:chat"
}
```

**Response:**
```json
{
  "canceled": 1,
  "message": "Canceled 1 execution(s)"
}
```

**Authentication:** Required (Bearer token)

### GET /api/v1/chat/active

Get all active agent executions.

**Response:**
```json
[
  {
    "id": "web:chat:1708472881234567890",
    "session_id": "web:chat",
    "started_at": "2026-02-20T04:08:01Z",
    "duration": "45.3s"
  }
]
```

**Authentication:** Required (Bearer token)

## Code Changes

### Backend (pkg/web/server.go)

1. **New struct `activeExecution`:**
   - Tracks session ID, start time, and cancel function
   - Stored in `Server.activeExecs` map

2. **Updated `handleChatWS`:**
   - Creates cancelable context with `context.WithCancel()`
   - Tracks execution in `activeExecs` map
   - Cleans up on completion
   - Returns "Execution canceled by user" message on cancellation

3. **New endpoints:**
   - `handleChatCancel`: Cancels executions by session ID
   - `handleChatActive`: Lists all active executions

### Frontend (pkg/web/frontend/src/views/ChatView.vue)

1. **Cancel button UI:**
   - Conditionally rendered when `isLoading` is true
   - Red button with X icon
   - Positioned between mic and send buttons

2. **`cancelExecution` method:**
   - Calls `/api/v1/chat/cancel` API
   - Shows toast notification on success/failure
   - Resets loading state

3. **Background execution:**
   - `onBeforeUnmount` removes listeners but keeps WebSocket connected
   - Allows navigation without interrupting agent

### Composable (pkg/web/frontend/src/composables/useActiveExecutions.js)

New composable for monitoring active executions:
- Auto-polling with configurable interval
- Lifecycle management (auto-start/stop)
- Reactive state for active executions list

## User Experience

### Before
- No way to stop agent mid-execution (had to wait or refresh page)
- Navigating away from chat interrupted agent work
- No visibility into background tasks

### After
- Click cancel button to stop agent immediately
- Navigate freely while agent works in background
- Optional indicator for active background tasks
- Better control and flexibility

## Testing

### Manual Testing Steps

1. **Test Cancel:**
   - Start a long-running task (e.g., "search the web for 10 different topics")
   - Click cancel button while agent is working
   - Verify: Loading indicator disappears, toast shows "Execution canceled"
   - Check agent stopped processing

2. **Test Background Execution:**
   - Start a task in chat
   - Navigate to Tasks or Settings view
   - Wait for agent to complete
   - Return to chat view
   - Verify: Response appears in chat history

3. **Test Active Executions API:**
   ```bash
   # Login to get token
   TOKEN=$(curl -s http://localhost:18880/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"your-password"}' | jq -r .token)
   
   # Start a task (in web UI)
   
   # Check active executions
   curl -s http://localhost:18880/api/v1/chat/active \
     -H "Authorization: Bearer $TOKEN" | jq .
   ```

## Future Enhancements

1. **Active Tasks Indicator:**
   - Add badge to sidebar showing count of active executions
   - Visual indicator when tasks complete in background

2. **Execution History:**
   - Track completed executions with duration and outcome
   - Allow viewing past execution details

3. **Multi-execution Management:**
   - Bulk cancel multiple executions
   - Priority management for concurrent tasks

4. **Notifications:**
   - Browser notifications when background tasks complete
   - Sound alerts for important events

## Troubleshooting

### Cancel button doesn't appear
- Check: Is `isLoading` state being set correctly?
- Check: Is WebSocket connected?
- Check: Browser console for errors

### Cancellation doesn't stop agent
- Check: Is context propagating through agent loop?
- Check: Docker logs for context cancellation errors
- Verify: `/api/v1/chat/cancel` returns `canceled: 1`

### Background execution interrupted
- Check: WebSocket connection state
- Verify: `onBeforeUnmount` is NOT disconnecting WebSocket
- Check: Router navigation isn't forcing reload

## Related Files

- Backend: `pkg/web/server.go`
- Frontend: `pkg/web/frontend/src/views/ChatView.vue`
- Composable: `pkg/web/frontend/src/composables/useActiveExecutions.js`
- Documentation: `docs/features/cancel-and-background-execution.md`

## Version

- **Introduced:** v0.9.0 (2026-02-19)
- **Last Updated:** 2026-02-20
