# Update Complete

**Telegram Bot:**

- Added debug logging and timeouts to the connection. The "Thinking..." issue should now be diagnosable if it persists, but the restart might have fixed it.

**Chat History Filtering:**

- Updated the backend (`server.go`) to automatically filter out sessions starting with `task:` from the API response.
- Chat history sidebar should now only show real conversations.

**Deployment:**

- Rebuilt and restarted the Docker container.

Please reload the frontend to see the changes.
