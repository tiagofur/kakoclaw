# Workflows API Reference

REST API endpoints for programmatic workflow management.

**Base URL**: `http://localhost:18880/api/v1` (or your configured `web.listen_addr`)

**Authentication**: Required (session cookie or API key)

---

## Table of Contents

- [Endpoints](#endpoints)
  - [List Workflows](#list-workflows)
  - [Create Workflow](#create-workflow)
  - [Get Workflow](#get-workflow)
  - [Update Workflow](#update-workflow)
  - [Delete Workflow](#delete-workflow)
  - [Run Workflow](#run-workflow)
  - [List Workflow Runs](#list-workflow-runs)
  - [List Available Tools](#list-available-tools)
- [Data Structures](#data-structures)
- [Error Responses](#error-responses)
- [Examples](#examples)

---

## Endpoints

### List Workflows

Get all workflows for the current user.

**Request:**
```http
GET /api/v1/workflows
```

**Response:** `200 OK`
```json
{
  "workflows": [
    {
      "id": 1,
      "name": "Code Review Bot",
      "description": "Analyzes code files and flags issues",
      "enabled": true,
      "steps": "[{\"id\":\"step_1\",\"type\":\"tool\",\"label\":\"Read file\",\"on_error\":\"stop\",\"config\":{\"tool_name\":\"read_file\",\"args\":{\"path\":\"main.go\"}}}]",
      "schedule": null,
      "created_at": "2026-02-20T10:30:00Z",
      "updated_at": "2026-02-20T10:30:00Z"
    }
  ]
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:18880/api/v1/workflows \
  -H "Cookie: session=your-session-token"
```

---

### Create Workflow

Create a new workflow.

**Request:**
```http
POST /api/v1/workflows
Content-Type: application/json
```

**Body:**
```json
{
  "name": "My Workflow",
  "description": "Does something useful",
  "steps": [
    {
      "id": "step_1",
      "type": "prompt",
      "label": "Generate ideas",
      "on_error": "stop",
      "config": {
        "message": "Generate 5 project ideas",
        "model": ""
      }
    }
  ],
  "schedule": null
}
```

**Response:** `201 Created`
```json
{
  "id": 2,
  "name": "My Workflow",
  "description": "Does something useful",
  "enabled": true,
  "steps": "[{\"id\":\"step_1\",\"type\":\"prompt\",\"label\":\"Generate ideas\",\"on_error\":\"stop\",\"config\":{\"message\":\"Generate 5 project ideas\",\"model\":\"\"}}]",
  "schedule": null,
  "created_at": "2026-02-21T14:20:00Z",
  "updated_at": "2026-02-21T14:20:00Z"
}
```

**Validation:**
- `name` is required (string, non-empty)
- `steps` must be valid JSON array
- Each step must have: `id`, `type`, `label`, `on_error`, `config`

**cURL Example:**
```bash
curl -X POST http://localhost:18880/api/v1/workflows \
  -H "Cookie: session=your-session-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Quick Test",
    "description": "Test workflow",
    "steps": [
      {
        "id": "step_1",
        "type": "prompt",
        "label": "Say hello",
        "on_error": "stop",
        "config": {
          "message": "Say hello world",
          "model": ""
        }
      }
    ]
  }'
```

---

### Get Workflow

Retrieve a single workflow by ID.

**Request:**
```http
GET /api/v1/workflows/{id}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Code Review Bot",
  "description": "Analyzes code files",
  "enabled": true,
  "steps": "[...]",
  "schedule": null,
  "created_at": "2026-02-20T10:30:00Z",
  "updated_at": "2026-02-20T10:30:00Z"
}
```

**Error:** `404 Not Found`
```json
{
  "error": "workflow not found"
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:18880/api/v1/workflows/1 \
  -H "Cookie: session=your-session-token"
```

---

### Update Workflow

Update an existing workflow.

**Request:**
```http
PUT /api/v1/workflows/{id}
Content-Type: application/json
```

**Body:**
```json
{
  "name": "Updated Name",
  "description": "Updated description",
  "enabled": false,
  "steps": [
    {
      "id": "step_1",
      "type": "prompt",
      "label": "Updated step",
      "on_error": "continue",
      "config": {
        "message": "New prompt",
        "model": ""
      }
    }
  ],
  "schedule": null
}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Updated Name",
  "description": "Updated description",
  "enabled": false,
  "steps": "[...]",
  "schedule": null,
  "created_at": "2026-02-20T10:30:00Z",
  "updated_at": "2026-02-21T15:45:00Z"
}
```

**Notes:**
- All fields are optional except `name`
- If `enabled` is omitted, existing value is preserved
- `steps` fully replaces existing steps (no partial updates)

**cURL Example:**
```bash
curl -X PUT http://localhost:18880/api/v1/workflows/1 \
  -H "Cookie: session=your-session-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Code Review Bot v2",
    "description": "Improved analyzer",
    "enabled": true,
    "steps": [...]
  }'
```

---

### Delete Workflow

Delete a workflow permanently.

**Request:**
```http
DELETE /api/v1/workflows/{id}
```

**Response:** `200 OK`
```json
{
  "status": "deleted"
}
```

**Error:** `404 Not Found`
```json
{
  "error": "failed to delete: <reason>"
}
```

**Notes:**
- Cascades to `workflow_runs` (all run history deleted)
- Cannot be undone

**cURL Example:**
```bash
curl -X DELETE http://localhost:18880/api/v1/workflows/1 \
  -H "Cookie: session=your-session-token"
```

---

### Run Workflow

Execute a workflow and return results.

**Request:**
```http
POST /api/v1/workflows/{id}/run
```

**Response:** `200 OK`
```json
{
  "ok": true,
  "results": [
    {
      "step_id": "step_1",
      "step_type": "prompt",
      "label": "Generate ideas",
      "output": "1. Build a CLI tool\n2. Create a web service\n3. ...",
      "error": "",
      "duration_ms": 1234,
      "skipped": false
    },
    {
      "step_id": "step_2",
      "step_type": "tool",
      "label": "Save to file",
      "output": "File written successfully",
      "error": "",
      "duration_ms": 56,
      "skipped": false
    }
  ]
}
```

**Error Response:** `500 Internal Server Error`
```json
{
  "error": "execution failed: context deadline exceeded"
}
```

**Notes:**
- **Timeout**: 5 minutes (300 seconds) max
- Creates a `workflow_run` record in database
- Execution is synchronous (blocking)
- Returns results even if some steps failed (check `error` field per step)

**cURL Example:**
```bash
curl -X POST http://localhost:18880/api/v1/workflows/1/run \
  -H "Cookie: session=your-session-token" \
  -H "Content-Type: application/json"
```

**Timeout Example:**
```bash
# Set custom timeout with curl (must be less than server's 5min limit)
curl -X POST http://localhost:18880/api/v1/workflows/1/run \
  -H "Cookie: session=your-session-token" \
  --max-time 120
```

---

### List Workflow Runs

Get execution history for a workflow.

**Request:**
```http
GET /api/v1/workflows/{id}/runs
```

**Response:** `200 OK`
```json
{
  "runs": [
    {
      "id": 42,
      "workflow_id": 1,
      "status": "completed",
      "results": "[{\"step_id\":\"step_1\",\"step_type\":\"prompt\",\"label\":\"Generate ideas\",\"output\":\"...\",\"duration_ms\":1234}]",
      "started_at": "2026-02-21T16:30:00Z",
      "finished_at": "2026-02-21T16:30:05Z"
    },
    {
      "id": 41,
      "workflow_id": 1,
      "status": "failed",
      "results": "[{\"step_id\":\"step_1\",\"step_type\":\"prompt\",\"label\":\"Generate ideas\",\"output\":\"\",\"error\":\"timeout\",\"duration_ms\":300000}]",
      "started_at": "2026-02-21T16:25:00Z",
      "finished_at": "2026-02-21T16:30:00Z"
    }
  ]
}
```

**Query Parameters:**
- None (returns last 20 runs by default)

**Status Values:**
- `running`: Currently executing
- `completed`: All steps succeeded
- `completed_with_errors`: Some steps failed but workflow continued
- `failed`: Stopped due to error or timeout

**cURL Example:**
```bash
curl -X GET http://localhost:18880/api/v1/workflows/1/runs \
  -H "Cookie: session=your-session-token"
```

---

### List Available Tools

Get list of registered tools for use in workflow steps.

**Request:**
```http
GET /api/v1/tools
```

**Response:** `200 OK`
```json
{
  "tools": [
    "read_file",
    "write_file",
    "append_file",
    "delete_file",
    "list_directory",
    "shell_exec",
    "web_search",
    "web_fetch",
    "send_message",
    "create_task",
    "update_task",
    "knowledge_add",
    "knowledge_search"
  ]
}
```

**Use Case:**
- Validate tool names before creating workflow
- Populate tool name dropdowns in UI
- Generate documentation

**cURL Example:**
```bash
curl -X GET http://localhost:18880/api/v1/tools \
  -H "Cookie: session=your-session-token"
```

---

## Data Structures

### Workflow Object

```typescript
{
  id: number,              // Auto-generated
  name: string,            // Required, unique identifier
  description: string,     // Optional, human-readable description
  enabled: boolean,        // Default: true
  steps: string,           // JSON-encoded Step[] array
  schedule: string | null, // Reserved for future cron integration
  created_at: string,      // ISO 8601 timestamp
  updated_at: string       // ISO 8601 timestamp
}
```

### Step Object

```typescript
{
  id: string,              // Unique within workflow (e.g., "step_1234567890_1")
  type: "prompt" | "tool" | "condition",
  label: string,           // Human-readable step name
  on_error: "stop" | "continue",
  config: StepConfig       // Type-specific configuration
}
```

### PromptConfig

```typescript
{
  message: string,         // Prompt to send to LLM
  model: string            // Optional model override (e.g., "claude-3-5-sonnet-20241022")
}
```

### ToolConfig

```typescript
{
  tool_name: string,       // Exact tool name from registry
  args: {                  // Tool-specific arguments
    [key: string]: any
  }
}
```

### ConditionConfig

```typescript
{
  operator: "contains" | "equals" | "not_empty" | "regex",
  reference: string,       // Template variable (e.g., "{{step.1.output}}")
  value: string            // Comparison value
}
```

### StepResult Object

```typescript
{
  step_id: string,         // Matches Step.id
  step_type: string,       // "prompt", "tool", or "condition"
  label: string,           // Human-readable name
  output: string,          // Successful result
  error: string,           // Error message (empty if success)
  duration_ms: number,     // Execution time in milliseconds
  skipped: boolean         // True if skipped by condition
}
```

### WorkflowRun Object

```typescript
{
  id: number,              // Auto-generated
  workflow_id: number,     // Foreign key to workflows table
  status: "running" | "completed" | "completed_with_errors" | "failed",
  results: string,         // JSON-encoded StepResult[] array
  started_at: string,      // ISO 8601 timestamp
  finished_at: string | null  // ISO 8601 timestamp (null if still running)
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "Human-readable error message"
}
```

### Common Status Codes

- `200 OK`: Success
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid input (malformed JSON, missing required fields)
- `404 Not Found`: Workflow or resource not found
- `500 Internal Server Error`: Server-side error (timeout, database error, etc.)
- `503 Service Unavailable`: Workflow engine or storage not available

### Example Error Responses

**Invalid JSON:**
```json
{
  "error": "invalid JSON: unexpected end of JSON input"
}
```

**Missing Required Field:**
```json
{
  "error": "name is required"
}
```

**Workflow Not Found:**
```json
{
  "error": "workflow not found"
}
```

**Execution Timeout:**
```json
{
  "error": "execution failed: context deadline exceeded"
}
```

---

## Examples

### Complete Workflow Lifecycle

```bash
#!/bin/bash

BASE_URL="http://localhost:18880/api/v1"
SESSION="your-session-cookie"

# 1. Create workflow
WORKFLOW_ID=$(curl -s -X POST "$BASE_URL/workflows" \
  -H "Cookie: session=$SESSION" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Workflow",
    "description": "API test",
    "steps": [
      {
        "id": "step_1",
        "type": "prompt",
        "label": "Generate text",
        "on_error": "stop",
        "config": {
          "message": "Write a haiku about programming",
          "model": ""
        }
      }
    ]
  }' | jq -r '.id')

echo "Created workflow ID: $WORKFLOW_ID"

# 2. Run workflow
curl -s -X POST "$BASE_URL/workflows/$WORKFLOW_ID/run" \
  -H "Cookie: session=$SESSION" | jq '.results'

# 3. Check run history
curl -s -X GET "$BASE_URL/workflows/$WORKFLOW_ID/runs" \
  -H "Cookie: session=$SESSION" | jq '.runs[0]'

# 4. Cleanup
curl -s -X DELETE "$BASE_URL/workflows/$WORKFLOW_ID" \
  -H "Cookie: session=$SESSION"
```

---

### Python Integration Example

```python
import requests
import json

class WorkflowClient:
    def __init__(self, base_url="http://localhost:18880", session_cookie=None):
        self.base_url = base_url
        self.session = requests.Session()
        if session_cookie:
            self.session.cookies.set('session', session_cookie)
    
    def create_workflow(self, name, description, steps):
        """Create a new workflow."""
        url = f"{self.base_url}/api/v1/workflows"
        payload = {
            "name": name,
            "description": description,
            "steps": steps
        }
        resp = self.session.post(url, json=payload)
        resp.raise_for_status()
        return resp.json()
    
    def run_workflow(self, workflow_id):
        """Execute a workflow and return results."""
        url = f"{self.base_url}/api/v1/workflows/{workflow_id}/run"
        resp = self.session.post(url, timeout=310)  # Slightly more than server's 5min
        resp.raise_for_status()
        return resp.json()
    
    def get_runs(self, workflow_id):
        """Get run history for a workflow."""
        url = f"{self.base_url}/api/v1/workflows/{workflow_id}/runs"
        resp = self.session.get(url)
        resp.raise_for_status()
        return resp.json()['runs']

# Usage
client = WorkflowClient(session_cookie="your-session-token")

# Create
wf = client.create_workflow(
    name="Python Test",
    description="Created from Python",
    steps=[
        {
            "id": "step_1",
            "type": "prompt",
            "label": "Generate code",
            "on_error": "stop",
            "config": {
                "message": "Write a Python function to reverse a string",
                "model": ""
            }
        }
    ]
)

print(f"Created workflow {wf['id']}")

# Run
results = client.run_workflow(wf['id'])
print(f"Execution completed: {len(results['results'])} steps")
for step in results['results']:
    print(f"- {step['label']}: {step['output'][:50]}...")

# History
runs = client.get_runs(wf['id'])
print(f"Total runs: {len(runs)}")
```

---

### JavaScript (Node.js) Example

```javascript
const axios = require('axios');

class WorkflowClient {
  constructor(baseURL = 'http://localhost:18880', sessionCookie) {
    this.client = axios.create({
      baseURL: baseURL + '/api/v1',
      headers: {
        'Cookie': `session=${sessionCookie}`
      },
      timeout: 310000 // 5min + buffer
    });
  }

  async createWorkflow(name, description, steps) {
    const { data } = await this.client.post('/workflows', {
      name,
      description,
      steps
    });
    return data;
  }

  async runWorkflow(workflowId) {
    const { data } = await this.client.post(`/workflows/${workflowId}/run`);
    return data;
  }

  async getRuns(workflowId) {
    const { data } = await this.client.get(`/workflows/${workflowId}/runs`);
    return data.runs;
  }

  async deleteWorkflow(workflowId) {
    const { data } = await this.client.delete(`/workflows/${workflowId}`);
    return data;
  }
}

// Usage
(async () => {
  const client = new WorkflowClient('http://localhost:18880', 'your-session-token');

  // Create workflow
  const wf = await client.createWorkflow(
    'Node.js Test',
    'Created from Node.js',
    [
      {
        id: 'step_1',
        type: 'tool',
        label: 'Check directory',
        on_error: 'stop',
        config: {
          tool_name: 'list_directory',
          args: { path: '.' }
        }
      }
    ]
  );

  console.log(`Created workflow ${wf.id}`);

  // Run it
  const results = await client.runWorkflow(wf.id);
  console.log('Results:', results.results);

  // Cleanup
  await client.deleteWorkflow(wf.id);
})();
```

---

## Rate Limiting

⚠️ **Note**: No rate limiting currently implemented. Workflow execution time is limited to 5 minutes per run.

**Best Practices:**
- Don't run workflows in tight loops
- Consider execution time when polling for results
- Use run history endpoint sparingly

---

## Changelog

### v1.0.0 (February 2026)
- Initial API release
- CRUD operations for workflows
- Synchronous workflow execution with 5-minute timeout
- Run history tracking
- Tool listing endpoint

### Planned Features
- Asynchronous workflow execution
- Webhook triggers
- Workflow export/import
- Per-user workflow isolation
- Scheduled execution (cron integration)

---

**Version**: 1.0.0  
**Last Updated**: February 2026  
**Stability**: Stable (MVP)
