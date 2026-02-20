# Quick Reference: Cancel & Background Execution

## ✅ Implementation Complete

**Date:** February 20, 2026  
**Version:** v0.9.0  
**Status:** Deployed and verified

## What's New

### 🛑 Cancel Execution
Stop the agent mid-task with a single click.

**How to use:**
1. Start a task in chat
2. While agent is working, look for the red **X button** next to the send button
3. Click the X to cancel
4. Agent stops immediately

**Visual Cue:** Cancel button only appears when agent is actively working (loading indicator visible)

### 🔄 Background Execution
Navigate freely while agent works.

**How it works:**
1. Start a task in chat
2. Navigate to Tasks, Settings, or any other view
3. Agent continues working in background
4. Return to chat to see results

**No action required** - this happens automatically!

## Testing Instructions

### Test 1: Cancel a Task
```
1. Open: http://127.0.0.1:18880
2. Login with your credentials
3. In chat, type: "search the web for 10 different topics about AI"
4. Press Send
5. Wait 2-3 seconds (loading indicator appears)
6. Click the red X button
7. ✓ Verify: Loading stops, toast shows "Execution canceled"
```

### Test 2: Background Execution
```
1. In chat, type: "summarize the last 5 sessions"
2. Press Send (agent starts working)
3. Click "Tasks" in the sidebar
4. Wait 10 seconds
5. Return to Chat view
6. ✓ Verify: Agent's response appears in chat history
```

### Test 3: API Endpoints (Developer)
```bash
# Get auth token
TOKEN=$(curl -s http://localhost:18880/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}' | jq -r .token)

# Check active executions
curl -s http://localhost:18880/api/v1/chat/active \
  -H "Authorization: Bearer $TOKEN" | jq .

# Cancel execution (while task is running)
curl -X POST http://localhost:18880/api/v1/chat/cancel \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"session_id":"web:chat"}' | jq .
```

## Files Changed

### Backend
- `pkg/web/server.go` - Added cancel/active endpoints and context tracking

### Frontend
- `pkg/web/frontend/src/views/ChatView.vue` - Added cancel button and background support
- `pkg/web/frontend/src/composables/useActiveExecutions.js` - New composable (optional)

### Documentation
- `docs/features/cancel-and-background-execution.md` - Full feature docs
- `IMPLEMENTATION_SUMMARY_CANCEL_FEATURE.md` - Implementation details
- `scripts/verify-cancel-feature.sh` - Automated verification

## Verification Results

```bash
./scripts/verify-cancel-feature.sh
```

**All checks passed:**
- ✅ Web server running (port 18880)
- ✅ Cancel endpoint exists
- ✅ Active endpoint exists
- ✅ Frontend code deployed
- ✅ No critical errors in logs

## Technical Architecture

### Cancel Flow
```
User clicks Cancel
  → POST /api/v1/chat/cancel
    → Backend finds execution in activeExecs map
      → Calls context.Cancel()
        → Agent loop receives ctx.Err() == Canceled
          → Stops processing
            → Returns "Execution canceled by user"
```

### Background Execution
```
User navigates away from Chat
  → onBeforeUnmount() called
    → Removes event listeners
      → WebSocket stays connected
        → Agent continues processing
          → Messages queued
            → User returns to Chat
              → onMounted() re-attaches listeners
                → Queued messages displayed
```

## UI Elements

### Cancel Button
- **Location:** Between microphone and send button
- **Color:** Red background
- **Icon:** X (cross)
- **Visibility:** Only when `isLoading === true`
- **Action:** Calls `/api/v1/chat/cancel`

### Global Loading Indicator
- **Location:** Top bar, next to model selector
- **Text:** "Agent working..."
- **Spinner:** Animated lightning bolt
- **Purpose:** Shows background tasks even when not in chat view

## API Reference

### POST /api/v1/chat/cancel
Cancel execution for a session.

**Auth:** Required (Bearer token)  
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

### GET /api/v1/chat/active
List active executions.

**Auth:** Required (Bearer token)  
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

## Troubleshooting

### Cancel button doesn't appear
**Check:** Is WebSocket connected? Look for "Connected" indicator in bottom-right of chat.

### Cancel doesn't stop agent
**Check:** Docker logs: `docker logs kakoclaw-kakoclaw-1 --tail 50`

### Background tasks interrupted
**Check:** Browser console for WebSocket errors: `F12 → Console`

## Known Limitations

1. **Multiple instances:** Running multiple Docker containers with same Telegram bot causes conflicts
2. **WebSocket reconnection:** If connection drops during navigation, background task may fail
3. **Browser tab close:** Closing browser tab disconnects WebSocket and stops agent

## Future Enhancements

- [ ] Visual badge showing count of active background tasks
- [ ] Browser notifications when background tasks complete
- [ ] Task history with execution time and status
- [ ] Bulk cancel (stop all active executions)
- [ ] Priority queue for concurrent tasks

## Support

**Logs:**
```bash
docker logs kakoclaw-kakoclaw-1 -f
```

**Status:**
```bash
curl http://localhost:18880/api/v1/health
```

**Container:**
```bash
docker ps | grep kakoclaw
```

## Summary

| Feature | Status | How to Use |
|---------|--------|-----------|
| Cancel execution | ✅ Live | Click red X button while agent working |
| Background tasks | ✅ Live | Navigate away from chat - agent continues |
| Active executions API | ✅ Live | GET /api/v1/chat/active |
| Cancel API | ✅ Live | POST /api/v1/chat/cancel |

**Ready to use!** Open http://127.0.0.1:18880 and test the features.
