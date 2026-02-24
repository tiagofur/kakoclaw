# Degraded Mode Guide

## Overview

**Degraded Mode** allows MakoClaw to start and run **without any LLM provider configured**. This is perfect for:

- **First-time setup**: Configure your provider via the web UI instead of editing JSON files
- **Docker deployments**: Start containers without pre-configuring API keys
- **Testing**: Verify your deployment works before adding credentials
- **Development**: Work on frontend/backend features without LLM dependencies

## What Works in Degraded Mode?

### ✅ Available Features

- **Web Panel**: Full access to the web interface
- **Authentication**: User login and session management
- **Setup Wizard**: Interactive provider configuration
- **Static Pages**: All UI components and navigation
- **Settings**: View and modify configuration (non-LLM settings)
- **Storage**: Database and file operations
- **REST API**: Most endpoints (except those requiring agent)

### ❌ Disabled Features

- **Agent Loop**: No LLM-powered chat or assistance
- **Cron Jobs**: Scheduled tasks (require agent loop)
- **AI Tools**: Any tool that requires LLM interaction
  - `spawn` (subagents)
  - AI-powered features in workflows
- **Orchestrator**: Multi-agent delegation
- **Channel Bots**: Telegram, Discord, etc. (require agent)

## Starting in Degraded Mode

### Docker Compose

```bash
# Start without environment variables
docker-compose up -d
```

The container will start successfully even without:

- `MAKOCLAW_PROVIDERS_OPENROUTER_API_KEY`
- `MAKOCLAW_AGENTS_DEFAULTS_PROVIDER`
- `MAKOCLAW_AGENTS_DEFAULTS_MODEL`

### Binary/Native

```bash
# Start web mode
makoclaw web

# Or gateway mode
makoclaw gateway
```

If no provider is configured in `~/.MakoClaw/config.json`, you'll see:

```
⚠ DEGRADED MODE: No LLM provider configured
  • Agent loop disabled
  • Cron service disabled
  • Web panel available for configuration
  → Visit http://localhost:18880 to configure your LLM provider
```

## Configuring via Web UI

### Step 1: Access Web Panel

Navigate to `http://localhost:18880` (or your configured port).

You'll see a warning banner:

```
⚠️ Degraded Mode Active
No LLM provider configured. Agent features are disabled.
[Configure Now]
```

### Step 2: Click "Configure Now"

This opens the **Setup Wizard** with a friendly interface to configure your provider.

### Step 3: Select Provider

Choose from supported providers:

- **Ollama** (Self-hosted, free)
- **Anthropic** (Claude models)
- **OpenAI** (GPT models)
- **OpenRouter** (Multiple models, free tier)
- **Groq** (Fast inference, free tier)

### Step 4: Enter Configuration

For each provider, you'll need different information:

#### Ollama

```
Base URL: http://localhost:11434
Model: llama2
```

#### Anthropic

```
API Key: sk-ant-xxxxx
Model: claude-3-5-sonnet-20241022
```

#### OpenAI

```
API Key: sk-xxxxx
Model: gpt-4
```

#### OpenRouter

```
API Key: sk-or-v1-xxxxx
Model: anthropic/claude-3.5-sonnet
```

#### Groq

```
API Key: gsk_xxxxx
Model: mixtral-8x7b-32768
```

### Step 5: Validate (Optional)

Click **"Test Connection"** to verify your credentials work.

### Step 6: Save

Click **"Save Configuration"** to persist your settings to disk.

### Step 7: Restart

Restart MakoClaw to enable full features:

```bash
# Docker
docker-compose restart

# Native
# Press Ctrl+C to stop, then start again
makoclaw web
```

## Configuration File Format

When you save via the web UI, your config is written to:

- **Docker**: `/root/.MakoClaw/config.json`
- **Native**: `~/.MakoClaw/config.json`

Example generated config:

```json
{
  "agents": {
    "defaults": {
      "provider": "openrouter",
      "model": "anthropic/claude-3.5-sonnet",
      "max_tokens": 8192,
      "temperature": 0.7
    }
  },
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-xxxxx",
      "base_url": "https://openrouter.ai/api/v1"
    }
  }
}
```

## Manual Configuration (Alternative)

If you prefer to edit the config file manually:

1. Stop MakoClaw
2. Edit `~/.MakoClaw/config.json`
3. Add provider configuration (see examples above)
4. Add agent defaults:
   ```json
   {
     "agents": {
       "defaults": {
         "provider": "openrouter",
         "model": "anthropic/claude-3.5-sonnet"
       }
     }
   }
   ```
5. Start MakoClaw

## Checking Current Status

### Via Web UI

The banner at the top shows current status:

- **Green border**: Provider configured and working
- **Orange banner**: Degraded Mode (no provider)
- **Red banner**: Provider misconfigured (validation failed)

### Via API

```bash
curl http://localhost:18880/api/v1/config/status
```

Response:

```json
{
  "hasProvider": false,
  "degradedMode": true,
  "configuredProvider": "",
  "configuredModel": ""
}
```

## Docker Environment Variables

### Required (Minimum)

```yaml
services:
  makoclaw:
    environment:
      # Only web password is required
      - MAKOCLAW_WEB_PASSWORD=your-secure-password
```

### Optional (LLM Configuration)

```yaml
# OpenRouter example
- MAKOCLAW_AGENTS_DEFAULTS_PROVIDER=openrouter
- MAKOCLAW_AGENTS_DEFAULTS_MODEL=anthropic/claude-3.5-sonnet
- MAKOCLAW_PROVIDERS_OPENROUTER_API_KEY=sk-or-v1-xxxxx

# Ollama example (self-hosted)
- MAKOCLAW_AGENTS_DEFAULTS_PROVIDER=ollama
- MAKOCLAW_AGENTS_DEFAULTS_MODEL=llama2
- MAKOCLAW_PROVIDERS_OLLAMA_BASE_URL=http://host.docker.internal:11434

# Anthropic example
- MAKOCLAW_AGENTS_DEFAULTS_PROVIDER=anthropic
- MAKOCLAW_AGENTS_DEFAULTS_MODEL=claude-3-5-sonnet-20241022
- MAKOCLAW_PROVIDERS_ANTHROPIC_API_KEY=sk-ant-xxxxx
```

See [Docker Deployment Guide](../deployment/DOCKER_DEPLOYMENT.md) for complete examples.

## Troubleshooting

### Web panel shows "Degraded Mode" even after configuring

1. Verify config was saved: `cat ~/.MakoClaw/config.json`
2. Check `hasValidProviderConfig()` returns true
3. Restart MakoClaw completely
4. Check logs for provider creation errors

### "Invalid API Key" error after configuration

1. Double-check your API key is correct
2. Verify the provider name matches (e.g., `anthropic` not `claude`)
3. Test the key directly with the provider's API
4. Check for typos in model name

### Docker container won't start

1. Verify `MAKOCLAW_WEB_PASSWORD` is set (required)
2. Check docker-compose logs: `docker-compose logs makoclaw`
3. Ensure no port conflicts on 18880
4. Verify volume mounts are correct

### Can't access web panel

1. Check container is running: `docker ps`
2. Verify port mapping: `docker port makoclaw-makoclaw-1`
3. Try `http://localhost:18880` directly (not 127.0.0.1)
4. Check firewall rules

## Best Practices

### Security

- **Never commit API keys** to version control
- **Use environment variables** for Docker deployments
- **Rotate keys regularly** via the web UI
- **Use Ollama** for fully offline/air-gapped deployments

### Development

- **Start in degraded mode** for frontend work
- **Use local Ollama** to avoid API costs during testing
- **Configure via UI** unless you have specific config needs
- **Test with free tiers** (OpenRouter, Groq) before paid providers

### Production

- **Pre-configure providers** in deployment pipelines
- **Health check** the `/api/v1/config/status` endpoint
- **Monitor logs** for provider errors
- **Set up Ollama** as a fallback provider

## Related Guides

- [Docker Deployment](../deployment/DOCKER_DEPLOYMENT.md)
- [Quick Start Guide](quickstart.md)
- [Configuration Reference](../development/CONFIG_ARCHITECTURE.md)
- [Multi-User Setup](../development/MULTI_AGENT_SETUP.md)
