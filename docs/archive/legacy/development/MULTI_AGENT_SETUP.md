# Multi-Agent Specialist System

## Overview

MakoClaw now supports a **multi-agent specialist system** where an **Orchestrator Agent** analyzes incoming tasks and intelligently delegates them to specialized agents. This allows for:

- 🎯 **Task-specific optimization**: Each specialist has its own model, provider, and capability set
- 💰 **Cost efficiency**: Use cheaper models for simple tasks, premium models for complex ones
- 🚀 **Better performance**: Specialists excel at their domains
- 📊 **Transparent tracking**: Monitor which specialist handled each task and track costs

## Architecture

```
User Request
    ↓
[Orchestrator Agent] ← Analyzes task, picks specialist
    ↓
[Specialist: Developer | Testing | Documentation | DevOps | Analyst | Researcher]
    ↓
Result returned to user with specialist attribution
```

## Isolation guarantees

In multi-user deployments, the orchestrator, every specialist, spawned specialist tasks, and swarm members now inherit the active user's identity before they run.

- Each agent resolves its workspace under `~/.MakoClaw/users/<uuid>/workspace`.
- Session files stay under that user's `sessions/` directory and are not shared across users.
- Swarm and spawned specialist audit records now keep both `UserUUID` and `UserID` so delegation history stays attributable to the originating user.
- Runtime-added specialists are rehydrated with the same storage and user context as the parent agent manager.

## Configuration

### Basic Orchestrator Setup

```json
{
  "agents": {
    "orchestrator": {
      "enabled": true,
      "provider": "anthropic",
      "model": "claude-opus",
      "max_tokens": 12000,
      "temperature": 0.7,
      "max_delegation_retries": 2,
      "fallback_to_default": true,
      "description": "Project Manager: Analyzes tasks and delegates to specialists"
    },
    "specialists": {
      "developer": { ... },
      "testing": { ... }
    }
  }
}
```

### Specialist Configuration

Each specialist has:

```json
{
  "name": "developer",
  "description": "Software Developer: Handles coding tasks, refactoring, debugging",
  "prompt": "You are an expert software developer...",
  "provider": "anthropic",
  "model": "claude-opus",
  "max_tokens": 12000,
  "temperature": 0.3,
  "max_tool_iterations": 20,
  "tools": ["file_read", "file_write", "execute_shell", "list_dir"],
  "keywords": ["code", "programming", "debug", "refactor"]
}
```

**Key fields:**

| Field | Purpose |
|-------|---------|
| `name` | Unique identifier for the specialist |
| `description` | What this specialist does (used by orchestrator) |
| `prompt` | System prompt to prime the specialist |
| `provider` | LLM provider (anthropic, openai, openrouter, etc.) |
| `model` | Specific model to use |
| `tools` | Allowed tools (restricted for security/cost) |
| `keywords` | Topics specialist handles (for orchestrator matching) |

## Pre-configured Specialists

### 1. Developer
- **Best for**: Coding, refactoring, debugging, architecture
- **Model**: Claude Opus (strongest coder)
- **Tools**: file operations, shell execution
- **Cost optimization**: Premium model for complex code tasks

### 2. Documentation
- **Best for**: Writing READMEs, API docs, guides
- **Model**: Open source (cheaper)
- **Tools**: File operations only
- **Cost optimization**: No exec needed for docs

### 3. Testing
- **Best for**: Writing tests, validating code
- **Model**: Claude Sonnet (good balance)
- **Tools**: File operations, shell execution
- **Cost optimization**: Medium-tier model for balanced cost/quality

### 4. DevOps
- **Best for**: Deployment, CI/CD, Docker, Kubernetes
- **Model**: Open source (cost-effective)
- **Tools**: File operations, shell execution
- **Cost optimization**: Good for infrastructure tasks

### 5. Analyst
- **Best for**: Data analysis, reports, insights
- **Model**: Open source (flexible)
- **Tools**: File operations only
- **Cost optimization**: Analysis doesn't need inference overhead

### 6. Researcher
- **Best for**: Web search, information gathering, synthesis
- **Model**: Open source (cost-effective)
- **Tools**: Web search, file reading
- **Cost optimization**: Stream-oriented, doesn't need complex reasoning

## Cost Optimization Tips

### 1. **Assign models strategically**
```json
{
  "developer": {
    "model": "claude-opus",  // Premium for complex code
    "temperature": 0.3        // Lower = more focused, cheaper
  },
  "documentation": {
    "model": "llama-2-70b",   // Open source, sufficient quality
    "temperature": 0.5        // Higher = more creative
  }
}
```

### 2. **Restrict tools per specialist**
Only give tools each specialist needs:
```json
{
  "documentation": {
    "tools": ["file_read", "file_write"]  // No shell exec
  },
  "developer": {
    "tools": ["file_read", "file_write", "execute_shell"]  // Needs execution
  }
}
```

### 3. **Monitor per-specialist costs**
Access cost analytics via:
```bash
GET /api/v1/agents
GET /api/v1/agents/{specialist_name}
GET /api/v1/agents/{specialist_name}/analytics
```

### 4. **Use OpenRouter for cost efficiency**
OpenRouter offers:
- Best-price models
- Routing across providers
- One API key for 100+ models

```json
{
  "researcher": {
    "provider": "openrouter",
    "model": "meta-llama/llama-2-70b"  // Much cheaper
  }
}
```

## Usage Examples

### Enable Multi-Agent System

1. **Set up in config.json**:
```bash
cp config.example.json ~/.MakoClaw/config.json
# Edit to enable orchestrator and add your API keys
```

2. **Start MakoClaw**:
```bash
./makoclaw gateway
```

3. **Access via web or client**:
- Chat normally - orchestrator will analyze and delegate
- Responses show which specialist handled it

### Custom Specialist Example

Create a Python specialist:

```json
{
  "python_expert": {
    "name": "python_expert",
    "description": "Python expert: Advanced Python, async, libraries",
    "prompt": "You are a Python expert with deep knowledge of async, type hints, testing frameworks...",
    "provider": "anthropic",
    "model": "claude-sonnet",
    "max_tokens": 8192,
    "temperature": 0.4,
    "tools": ["file_read", "file_write", "execute_shell"],
    "keywords": ["python", "async", "type hints", "fastapi", "numpy"]
  }
}
```

## API Endpoints

### List Specialists
```bash
GET /api/v1/agents
```

Response:
```json
{
  "orchestrated": true,
  "specialists": [
    {
      "name": "developer",
      "description": "Software Developer...",
      "tools": ["file_read", "file_write", "execute_shell", "list_dir"]
    }
  ]
}
```

### Get Specialist Details
```bash
GET /api/v1/agents/{name}
```

### Get Specialist Sessions
```bash
GET /api/v1/agents/{name}/sessions
```

### Update Specialist Config (Runtime)
```bash
POST /api/v1/agents/{name}/config/update
Content-Type: application/json

{
  "provider": "openai",
  "model": "gpt-4"
}
```

## Session Tracking

Each session now includes:
- **agent_profile**: Which specialist handled it
- **parent_session_key**: Link to orchestrator session
- **specialist_metadata**: Stats about specialist execution

Access via:
```bash
GET /api/v1/chat/sessions/{id}
```

## Cost Tracking

The system tracks:
- **API calls per specialist**
- **Tokens used** (input + output)
- **Estimated costs** based on provider pricing
- **Efficiency metrics** (tokens per call, cost per token)

View with:
```bash
GET /api/v1/agents  # Shows all specialist info
```

## Failover Behavior

If a specialist fails:

1. **Orchestrator detects failure** (via tool error)
2. **Retries with alternative specialist** (if configured)
3. **Falls back to default agent** (if fallback enabled)
4. **Logs the incident** for debugging

## Advanced: Custom Specialist Prompt

Inject domain knowledge:

```json
{
  "ml_expert": {
    "prompt": "You are an ML engineer specializing in:\n- PyTorch and TensorFlow\n- Model optimization\n- Distributed training\n\nAlways explain trade-offs and provide production-ready code.",
    ...
  }
}
```

## Troubleshooting

### Orchestrator not delegating
- ✅ Check `"enabled": true` in config
- ✅ Ensure `specialists` section has entries
- ✅ Verify provider API keys are set

### Specialist failing
- ✅ Check specialist has required `tools`
- ✅ Verify provider config is correct
- ✅ Check logs: `tail -f ~/.MakoClaw/makoclaw.log`

### High costs
- ✅ Use cheaper models for non-expert tasks
- ✅ Reduce `max_tokens` per specialist
- ✅ Restrict `tools` to necessary ones
- ✅ Use OpenRouter for better pricing

## Performance Metrics

Monitor performance:
```
Tokens per call (avg) = Total tokens / API calls
Cost per call = Total cost / API calls  
Specialist efficiency = Tokens used / model premium
```

Lower cost/call = Better cost optimization ✓

## Next Steps

1. **Copy config.example.json**: `cp config.example.json ~/.MakoClaw/config.json`
2. **Enable orchestrator**: Set `"enabled": true`
3. **Add API keys**: Fill in provider credentials
4. **Start gateway**: `./makoclaw gateway`
5. **Monitor costs**: Check `/api/v1/agents` for usage
6. **Optimize**: Adjust models and tools based on metrics

## Questions?

- Check logs: `~/.MakoClaw/makoclaw.log`
- Test API: `curl http://localhost:18880/api/v1/agents`
- Run doctor: `./makoclaw doctor`

---

**Built for efficiency and extensibility** 🚀
