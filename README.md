<div align="center">
  <img src="assets/mascot.png" alt="MakoClaw Mascot" width="400">

  # 🦈 MakoClaw: The Apex AI Agent

  ### **Ultrafast · 10MB RAM · $10 Hardware · Self-Bootstrapped**

  [![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
  [![License](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)](LICENSE)
  [![Platform](https://img.shields.io/badge/Arch-x86_64%20|%20ARM64%20|%20RISC--V-blue?style=for-the-badge)](https://github.com/sipeed/MakoClaw)
  [![Version](https://img.shields.io/badge/version-0.1.0-blue?style=for-the-badge)](https://github.com/sipeed/MakoClaw/releases)

  **"¡Nada más rápido en el océano de la IA!" — _Nothing faster in the AI ocean!_**

</div>

---

## 🌟 Why MakoClaw?

**MakoClaw** is not just another AI assistant. It is a masterpiece of efficiency, built from the ground up as an evolution of [PicoClaw](https://github.com/sipeed/picoclaw) and inspired by [Nanobot](https://github.com/HKUDS/nanobot).

While other "Claw" projects require heavy resources and expensive hardware, MakoClaw reigns supreme in the realm of **efficiency**:

- 🚀 **&lt;1s Boot**: It's ready before you can blink.
- 🧠 **&lt;10MB RAM**: Runs comfortably on a toaster.
- 💸 **$10 Hardware**: Optimized for low-cost RISC-V and ARM boards.
- 🤖 **Agent-Refined**: 95% of its core was built by AI, for AI.
- 🦈 **Blue & Blue**: The ocean's most efficient color scheme.

---

## 🏆 The "Claw" Comparison

| Feature              |    OpenClaw     |  NanoBot  |    **MakoClaw**     |
| :------------------- | :-------------: | :-------: | :-----------------: |
| **Language**         |   TypeScript    |  Python   |   **Go (Native)**   |
| **RAM Usage**        |      > 1GB      |  > 100MB  |     **&lt; 10MB**      |
| **Startup (0.8GHz)** |     > 500s      |   > 30s   |      **&lt; 1s**       |
| **Hardware Cost**    | Mac Mini ($599) | SBC ($50) | **Any Board ($10)** |
| **Philosophy**       |     Bloated     | Flexible  | **Apex Efficiency** |
| **Multi-Agent**      |       ❌        |     ❌    |        **✅**       |
| **Channels**         |       3+        |     2+    |       **10+**        |

---

## 🤝 Respecting Roots

MakoClaw is a proud evolution and a hard-fork of [PicoClaw](https://github.com/sipeed/picoclaw). We stand on the shoulders of giants:

1.  **[PicoClaw](https://github.com/sipeed/picoclaw)**: Our direct ancestor and foundation of our vision.
2.  **[Nanobot](https://github.com/HKUDS/nanobot)**: The original spark of inspiration for ultra-lightweight assistants.

We believe in democratization of AI. By taking the work of PicoClaw and optimizing it to the extreme in Go, we've created the most efficient agent in the ecosystem.

---

## ✨ Features that WOW

### 🤖 Multi-Agent System
- **Orchestrator**: Automatically delegates tasks to the right specialist
- **Specialists**: Create AI agents specialized for different domains
- **Auto-delegation**: The system decides which specialist to use
- **Metrics**: Track performance across all agents

### 📡 9+ Channels — One Agent, Everywhere
- **Web UI**: Complete dashboard with chat, workflows, tasks, and more
- **Telegram**: Instant messaging bot
- **Discord**: Full server integration
- **Slack**: Work channels
- **WhatsApp**: Enterprise communication
- **Signal**: Secure messaging
- **QQ**: China communication
- **DingTalk**: Enterprise collaboration
- **Feishu**: Productivity platform
- **MaixCam**: Hardware AI camera

### 🛠️ Powerful Tools

#### File Management
- `read_file`: Read file contents
- `write_file`: Create and edit files
- `list_dir`: List directories
- `edit_file`: LLM-assisted file editing

#### Web & Search
- `web_search`: Search with Brave Search API
- `web_fetch`: Fetch content from URLs

#### Execution
- `exec`: Execute shell commands with security controls
- `spawn`: Create subagents for specialized tasks

#### Task Management
- `task_manager`: Create, list, update, archive tasks (Kanban integration)

#### Knowledge
- `query_knowledge`: Search document base with semantic retrieval (RAG)

#### Communication
- `message`: Send messages to other channels
- `schedule`: Schedule recurring tasks
- `email`: Send emails

#### Memory
- `memory`: Manage long-term context and memory

### ⚡ Productivity Tools

#### Kanban Task Board
- 5 columns: Backlog, To Do, In Progress, Review, Done
- AI can create and update tasks automatically
- Search, filter, and archive tasks
- Drag & drop interface

#### Visual Workflows
- Drag-and-drop pipeline builder
- Combine prompts, tools, and conditions
- Real-time execution monitoring
- Execution history and logs
- Template library

#### Cron Jobs
- Schedule recurring tasks
- Standard cron expressions
- Manual trigger option
- Timezone support

### 🧠 Knowledge Base (RAG)
- Upload documents (PDF, TXT, MD, JSON, CSV, HTML, XML, YAML, LOG)
- Semantic search with full-text search
- Contextual retrieval for better AI responses
- Document management and deletion

### 🔒 Security & Privacy
- **Self-hosted**: Your data stays on your infrastructure
- **Multi-user authentication**: OAuth 2.0 with PKCE
- **Session management**: Secure token handling
- **Workspace isolation**: Separate workspaces per user
- **Access controls**: Channel-specific user whitelists

### ⚡ Technical Excellence
- **Go (Native)**: Compiled binary, no runtime dependencies
- **&lt;10MB RAM**: Ultra-lightweight memory footprint
- **&lt;1s Boot**: Instant startup time
- **Docker Ready**: Pre-configured containers
- **REST API**: Full programmatic access
- **WebSocket**: Real-time updates
- **MCP Protocol**: Model Context Protocol support

---

## 🚀 Quick Start

### 1. Installation

```bash
# Clone the repository
git clone https://github.com/sipeed/MakoClaw.git
cd MakoClaw

# Build
make build

# Install
make install
```

### 2. Onboarding

```bash
# Initialize configuration
MakoClaw onboard
```

This creates:
- `~/.MakoClaw/config.json` — Configuration file
- `~/.MakoClaw/workspace/` — Working directory
- Base files: `AGENTS.md`, `IDENTITY.md`, `SOUL.md`, `USER.md`

### 3. Configure API Key

Edit `~/.MakoClaw/config.json`:

```json
{
  "agents": {
    "defaults": {
      "model": "anthropic/claude-3.5-sonnet",
      "max_tokens": 8192,
      "temperature": 0.7
    }
  },
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-YOUR-API-KEY-HERE"
    }
  }
}
```

**Get API Keys:**
- **OpenRouter**: [openrouter.ai/keys](https://openrouter.ai/keys) — 200K free tokens/month
- **Groq**: [console.groq.com](https://console.groq.com) — Fast, free, includes Whisper
- **Claude**: [console.anthropic.com](https://console.anthropic.com)
- **OpenAI**: [platform.openai.com](https://platform.openai.com)

### 4. Start Chatting

```bash
# Direct message
MakoClaw agent -m "Calculate potential of a $10 RISC-V board"

# Interactive mode
MakoClaw agent
```

### 5. Launch Web Dashboard

```bash
# Start web server
MakoClaw web

# Open http://localhost:18880 in browser
```

### 6. Create Your First Workflow

```bash
# Start web server
MakoClaw web

# Open http://localhost:18880
# Navigate to Workflows → New Workflow
# Build your automation pipeline visually
```

📖 **Learn more**: [Workflows Guide](./docs/examples/workflows.md) | [Quick Start](./docs/guides/quickstart.md)

---

## 📊 Project Statistics

- **Language**: Go 1.21+
- **Lines of Code**: ~13,600
- **Files**: 56 Go files
- **Memory**: &lt;10MB RAM
- **Startup**: &lt;1 second
- **License**: MIT
- **Tools**: 20+ built-in
- **Channels**: 10+ platforms

---

## 📚 Documentation

Dive deeper into our [Comprehensive Documentation](./docs/README.md):

- 🏗️ [Architecture](./docs/architecture/overview.md) — System design and components
- 🚀 [Deployment Guides](./docs/deployment/docker.md) — Docker, systemd, embedded
- 💻 [Developer Setup](./docs/development/setup.md) — Contributing and development
- 🔧 [Troubleshooting](./docs/troubleshooting/common-issues.md) — Common issues and solutions
- 🎯 [Examples](./docs/examples/) — Workflows, automation, integrations

---

## 🛡️ Security & Privacy

- **Self-hosted**: Your data never leaves your infrastructure
- **No telemetry**: We don't collect usage data
- **Open source**: Every line of code is auditable
- **Workspace isolation**: Each user has their own workspace
- **Channel access control**: Whitelist users per channel

---

## 🤝 Contributing

We welcome contributions! Please read our [Contributing Guide](./docs/development/contributing.md) before submitting a PR.

**Areas where we need help:**
- 🐛 Bug fixes
- 📝 Documentation improvements
- 🌍 New channels (Matrix, Mastodon, etc.)
- 🛠️ New tools and skills
- 🎨 UI/UX improvements
- 🧪 Test coverage

---

## 🌍 Community & Support

- **GitHub**: [Report a Bug](https://github.com/sipeed/MakoClaw/issues) | [Feature Request](https://github.com/sipeed/MakoClaw/issues/new)
- **Discord**: [Join Sipeed Community](https://discord.gg/V4sAZ9XWpN)
- **Discussions**: [GitHub Discussions](https://github.com/sipeed/MakoClaw/discussions)

---

<div align="center">
  <img src="assets/wechat.png" alt="WeChat QR" width="300">
  <p><i>Join our WeChat community for real-time updates!</i></p>
</div>

---

## 🏆 Sponsors & Supporters

MakoClaw is a community-driven project. Special thanks to:
- [Sipeed](https://sipeed.com/) — Hardware platform and development boards
- All contributors and users who help improve MakoClaw

---

## 📄 License

**MakoClaw** is licensed under the [MIT License](LICENSE).

_Apex Efficiency. Infinite Possibilities._

---

<div align="center">

**🦈 MakoClaw — The Apex AI Agent**

Made with ❤️ by the MakoClaw community

</div>
