# Dev Studio API Specification

## Purpose

REST API and WebSocket endpoints for the Dev Studio frontend. Manages projects, bridge sessions, memory, and real-time event streaming.

## Requirements

### Requirement: Project Management

The API MUST support CRUD operations for Dev Studio projects.

#### Scenario: Create project

- GIVEN an authenticated user
- WHEN `POST /api/dev/projects` with `{"name": "my-app", "path": "/home/user/my-app", "backend": "claude-code"}`
- THEN a project record MUST be persisted
- AND status 201 with the project JSON MUST be returned

#### Scenario: List projects

- GIVEN 2 existing projects for the user
- WHEN `GET /api/dev/projects`
- THEN both projects MUST be returned as JSON array

### Requirement: Bridge Control

The API MUST expose endpoints to start, stop, and query bridge status per project.

#### Scenario: Start bridge for project

- GIVEN a project "my-app" with no running bridge
- WHEN `POST /api/dev/projects/{id}/bridge/start`
- THEN a bridge process MUST be started for the project
- AND status 200 with `{"status": "running"}` MUST be returned

#### Scenario: Stop bridge

- GIVEN a project with running bridge
- WHEN `POST /api/dev/projects/{id}/bridge/stop`
- THEN the bridge MUST be stopped
- AND status 200 MUST be returned

#### Scenario: Get bridge status

- GIVEN a project with running bridge
- WHEN `GET /api/dev/projects/{id}/bridge/status`
- THEN `{"status": "running", "pid": N, "uptime_ms": N}` MUST be returned

### Requirement: Bridge Execution via WebSocket

The API MUST stream bridge events to the frontend via WebSocket.

#### Scenario: Send prompt and receive events

- GIVEN a WebSocket connection to `/api/dev/projects/{id}/ws`
- WHEN client sends `{"type": "execute", "prompt": "fix the auth bug"}`
- THEN bridge events MUST be streamed as WebSocket text frames
- AND each frame MUST be a JSON Event object
- AND the stream MUST end with a terminal event (result/error)

#### Scenario: WebSocket reconnect

- GIVEN a previously connected WebSocket that was disconnected
- WHEN a new WebSocket connection is established
- THEN the client MUST be able to resume receiving events from active requests

### Requirement: Memory API

The API MUST expose memory management endpoints (only when memory is enabled).

#### Scenario: Search memory

- GIVEN memory enabled and populated
- WHEN `POST /api/dev/memory/search` with `{"query": "error handling", "limit": 5}`
- THEN matching memories MUST be returned sorted by similarity

#### Scenario: Memory disabled

- GIVEN `dev_studio.memory.enabled: false` in config
- WHEN any memory endpoint is called
- THEN status 501 MUST be returned with `{"error": "dev studio memory is disabled"}`

### Requirement: Session Management

The API MUST support session listing, resume, and token-based auto-reset.

#### Scenario: List sessions

- GIVEN 3 bridge sessions for a project
- WHEN `GET /api/dev/projects/{id}/sessions`
- THEN all 3 sessions with token usage MUST be returned

#### Scenario: Auto-reset on token threshold

- GIVEN a session with 180K tokens used and threshold at 200K
- WHEN the next execute pushes tokens above 200K
- THEN the session MUST auto-reset
- AND a new session MUST start with the warm context summary
