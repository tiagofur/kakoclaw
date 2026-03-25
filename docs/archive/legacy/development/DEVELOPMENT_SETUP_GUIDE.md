# MakoClaw Development Setup Guide

Complete guide to set up MakoClaw for efficient, clean development with comprehensive testing, good UX/UI, and high-quality code.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Configuration Files](#configuration-files)
3. [Agent Setup](#agent-setup)
4. [Skills Installation](#skills-installation)
5. [MCP Integration](#mcp-integration)
6. [Testing Infrastructure](#testing-infrastructure)
7. [Development Workflow](#development-workflow)
8. [Quality Gates](#quality-gates)

## Quick Start

### Prerequisites

```bash
# Go (1.21+)
go version

# Node.js (20+) for frontend
node --version
npm --version

# Git
git --version
```

### Initial Setup

```bash
# Clone repository
git clone https://github.com/your-org/makoclaw.git
cd makoclaw

# Install Go dependencies
go mod download

# Install frontend dependencies
cd pkg/web/frontend
npm install
cd ../..

# Build the binary
make build

# Run MakoClaw
./makoclaw web
```

## Configuration Files

### 1. Development Configuration

Use the optimized development configuration:

```bash
# Copy development config
cp config.development.json ~/.MakoClaw/config.json

# Edit with your settings
nano ~/.MakoClaw/config.json
```

**Key Settings:**

```json
{
  "agents": {
    "defaults": {
      "provider": "openrouter",
      "model": "anthropic/claude-3.5-sonnet"
    },
    "orchestrator": {
      "enabled": true
    }
  },
  "web": {
    "enabled": true,
    "port": 18880
  }
}
```

### 2. Environment Variables

Create `.env` file:

```bash
# Providers
OPENROUTER_API_KEY=sk-or-v1-xxxxx
ANTHROPIC_API_KEY=sk-ant-xxxxx

# Web Interface
MAKOCLAW_WEB_PASSWORD=admin123

# MCP Services
GITHUB_PERSONAL_ACCESS_TOKEN=ghp_xxxxx
BRAVE_API_KEY=your_brave_api_key

# Email (Optional)
MAKOCLAW_TOOLS_EMAIL_ENABLED=true
MAKOCLAW_TOOLS_EMAIL_HOST=smtp.gmail.com
MAKOCLAW_TOOLS_EMAIL_PORT=587
MAKOCLAW_TOOLS_EMAIL_USERNAME=your_email@gmail.com
MAKOCLAW_TOOLS_EMAIL_PASSWORD=your_app_password
```

## Agent Setup

### Configured Specialists

The development config includes 14 pre-configured specialists:

| Specialist | Purpose | Model |
|-----------|---------|-------|
| **code-reviewer** | Code quality, security, performance reviews | claude-3.5-sonnet |
| **bug-fixer** | Diagnose and fix bugs systematically | claude-3.5-sonnet |
| **feature-developer** | Implement new features with best practices | claude-3.5-sonnet |
| **test-writer** | Write comprehensive tests | claude-3.5-sonnet |
| **documentation-writer** | Create clear documentation | claude-3.5-sonnet:beta |
| **refactoring-specialist** | Improve code structure | claude-3.5-sonnet |
| **performance-optimizer** | Identify and fix bottlenecks | claude-3.5-sonnet |
| **security-auditor** | Find security vulnerabilities | claude-3.5-sonnet |
| **backend-specialist** | Server-side architecture and APIs | claude-3.5-sonnet |
| **frontend-specialist** | UI/UX implementation | claude-3.5-sonnet |
| **architect-specialist** | System design and patterns | claude-3.5-sonnet |
| **devops-specialist** | CI/CD and infrastructure | claude-3.5-sonnet |
| **database-specialist** | Schema and query optimization | claude-3.5-sonnet |
| **mobile-specialist** | iOS/Android development | claude-3.5-sonnet |

### Adding Custom Specialists

```json
{
  "agents": {
    "specialists": {
      "my-specialist": {
        "name": "my-specialist",
        "description": "Does something specific",
        "prompt": "You are an expert in...",
        "provider": "openrouter",
        "model": "anthropic/claude-3.5-sonnet",
        "max_tokens": 12000,
        "temperature": 0.5,
        "max_tool_iterations": 25,
        "tools": ["file_read", "file_write", "execute_shell"],
        "keywords": ["custom", "specific"]
      }
    }
  }
}
```

### Using Specialists

**Via CLI:**
```bash
# Let orchestrator decide
makoclaw agent "Review this code for security issues"

# Specify specialist
makoclaw agent --specialist=security-auditor "Check auth.go for vulnerabilities"
```

**Via Web UI:**
1. Open http://localhost:18880
2. Go to Agents page
3. Select specialist or use auto mode

## Skills Installation

### Built-in Skills

Copy development skills to your workspace:

```bash
# Create skills directory
mkdir -p ~/.MakoClaw/skills/development

# Copy skills
cp -r skills/development/* ~/.MakoClaw/skills/development/
```

### Available Development Skills

#### 1. Go Best Practices

**File:** `skills/development/go-best-practices/SKILL.md`

**Covers:**
- Project structure and organization
- Naming conventions
- Error handling patterns
- Concurrency best practices
- Interface design
- Testing patterns
- Performance tips

**Usage:**
```
@go-best-practices How should I structure this Go package?
```

#### 2. Testing Strategy

**File:** `skills/development/test-strategy/SKILL.md`

**Covers:**
- Testing pyramid (unit, integration, E2E)
- TDD workflow
- Test patterns (table-driven, mocking, fixtures)
- Coverage requirements
- Performance testing
- Anti-patterns

**Usage:**
```
@test-strategy Help me write tests for this user service
```

#### 3. Code Review Checklist

**File:** `skills/development/code-review-checklist/SKILL.md`

**Covers:**
- Functionality and requirements
- Code quality and style
- Architecture and design
- Performance and scalability
- Security best practices
- Testing coverage
- Documentation

**Usage:**
```
@code-review-checklist Review this pull request
```

### Creating Custom Skills

```markdown
---
name: my-custom-skill
title: My Custom Skill
description: What this skill does
version: 1.0.0
author: Your Name
category: development
tags: [custom, specific]
---

# My Custom Skill

You are an expert in...

## Your Expertise

1. Topic 1
2. Topic 2
3. Topic 3

## Guidelines

- Rule 1
- Rule 2

## Examples

Example of how to use this skill...
```

## MCP Integration

### Recommended MCP Servers

#### 1. Filesystem MCP

Advanced file operations:

```json
{
  "mcp": {
    "servers": {
      "filesystem": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/projects"],
        "enabled": true
      }
    }
  }
}
```

#### 2. GitHub MCP

Repository automation:

```json
{
  "mcp": {
    "servers": {
      "github": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-github"],
        "env": {
          "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_PERSONAL_ACCESS_TOKEN}"
        },
        "enabled": true
      }
    }
  }
}
```

#### 3. Brave Search MCP

Web search capabilities:

```json
{
  "mcp": {
    "servers": {
      "brave-search": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-brave-search"],
        "env": {
          "BRAVE_API_KEY": "${BRAVE_API_KEY}"
        },
        "enabled": true
      }
    }
  }
}
```

### Installing MCP Servers

```bash
# Install Node.js packages
npm install -g @modelcontextprotocol/server-filesystem
npm install -g @modelcontextprotocol/server-github
npm install -g @modelcontextprotocol/server-brave-search

# Add to config.json
nano ~/.MakoClaw/config.json

# Restart MakoClaw
makoclaw web
```

### Testing MCP

```bash
# Check MCP server status
makoclaw doctor --mcp

# List available MCP tools
makoclaw tools --mcp

# Test specific tool
makoclaw agent "Search the web for Go best practices"
```

## Testing Infrastructure

### Running Tests

**Go Tests:**
```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run with race detector
go test ./... -race

# Run specific test
go test ./pkg/config -run TestParseProviderEnvVars -v

# Run benchmarks
go test ./... -bench=. -benchmem
```

**Frontend Tests:**
```bash
cd pkg/web/frontend

# Run tests
npm test

# Run with coverage
npm test -- --coverage

# Run E2E tests
npm run test:e2e
```

### Coverage Requirements

| Component | Minimum Coverage |
|-----------|------------------|
| Core Logic | 90% |
| API Handlers | 85% |
| Utilities | 90% |
| UI Components | 70% |

### Test Organization

```
pkg/
├── config/
│   ├── config.go
│   └── config_test.go           # Unit tests
├── web/
│   ├── handlers.go
│   └── handlers_test.go
└── ...

test/
├── integration/                 # Integration tests
│   ├── api_test.go
│   └── database_test.go
├── e2e/                         # E2E tests
│   └── auth_flow_test.go
├── fixtures/                    # Test data
│   └── users.json
└── testutil/                    # Test helpers
    └── db.go
```

## Development Workflow

### 1. Feature Development

```bash
# Create feature branch
git checkout -b feature/new-feature

# Write tests first (TDD)
# Create test file
vim pkg/myfeature/myfeature_test.go

# Run tests (should fail)
go test ./pkg/myfeature

# Implement feature
vim pkg/myfeature/myfeature.go

# Run tests again (should pass)
go test ./pkg/myfeature

# Run all tests
go test ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run

# Commit changes
git add .
git commit -m "feat: add new feature"

# Push to remote
git push origin feature/new-feature
```

### 2. Code Review

Using the code-review-checklist skill:

```bash
# Review your own code
makoclaw agent --specialist=code-reviewer "Review pkg/myfeature/"

# Create pull request
gh pr create --title "Add new feature" --body "Implements #123"

# Get AI review
makoclaw agent "Review this PR: https://github.com/.../pull/123"
```

### 3. Bug Fixing

```bash
# Use bug-fixer specialist
makoclaw agent --specialist=bug-fixer "Fix the authentication bug described in issue #456"

# Or use TDD approach
# 1. Write failing test
vim bug_test.go
go test ./...  # Fails

# 2. Write fix
vim bug.go
go test ./...  # Passes

# 3. Verify no regressions
go test ./... -race
```

### 4. Performance Optimization

```bash
# Run profiler
go test ./... -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Use performance-optimizer specialist
makoclaw agent --specialist=performance-optimizer "Optimize the database query in user.go"

# Run benchmarks
go test ./... -bench=. -benchmem
```

## Quality Gates

### Pre-commit Checks

Add `.git/hooks/pre-commit`:

```bash
#!/bin/bash

echo "🔍 Running pre-commit checks..."

# Format code
echo "📝 Formatting code..."
go fmt ./...

# Run linter
echo "🔎 Running linter..."
golangci-lint run || exit 1

# Run tests
echo "🧪 Running tests..."
go test ./... -race -cover || exit 1

# Check coverage
echo "📊 Checking coverage..."
go test ./... -coverprofile=coverage.out
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Coverage: $COVERAGE%"
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "❌ Coverage below 80%"
    exit 1
fi

echo "✅ All checks passed!"
```

Make executable:
```bash
chmod +x .git/hooks/pre-commit
```

### CI/CD Pipeline

See `.github/workflows/test.yml` for complete CI configuration.

Key checks:
- ✅ All tests pass
- ✅ Coverage ≥80%
- ✅ No race conditions
- ✅ Linter passes
- ✅ Security scan passes

## IDE Setup

### VS Code

Install extensions:
- Go (golang.go)
- Vue - Official (Vue.volar)
- ESLint
- Prettier

**Recommended Settings:**
```json
{
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.testFlags": ["-race", "-cover"],
  "editor.formatOnSave": true,
  "eslint.validate": ["javascript", "typescript", "vue"]
}
```

### Neovim/Vim

Install plugins:
```lua
-- Go
use 'ray-x/go.nvim'
use 'ray-x/guihua.lua'

-- Vue
use 'leafOfTree/vim-vue-plugin'

-- LSP
use 'neovim/nvim-lspconfig'
```

## Troubleshooting

### Common Issues

**Issue:** Tests fail with "connection refused"
```bash
# Solution: Start required services
docker-compose up -d postgres redis
```

**Issue:** MCP server not connecting
```bash
# Solution: Check logs
tail -f ~/.MakoClaw/workspace/logs/mcp.log

# Test MCP server directly
npx -y @modelcontextprotocol/server-github
```

**Issue:** Frontend build fails
```bash
# Solution: Clean and reinstall
cd pkg/web/frontend
rm -rf node_modules package-lock.json
npm install
npm run build
```

## Best Practices Summary

### Code Quality
- ✅ Write tests before/with code (TDD)
- ✅ Follow language idioms and style guides
- ✅ Keep functions small and focused
- ✅ Handle errors properly
- ✅ Document complex logic

### Testing
- ✅ Aim for >80% coverage
- ✅ Test edge cases and errors
- ✅ Use table-driven tests
- ✅ Mock external dependencies
- ✅ Keep tests fast

### Git
- ✅ Write clear commit messages
- ✅ Create feature branches
- ✅ Use pull requests for review
- ✅ Keep commits atomic

### Documentation
- ✅ Document public APIs
- ✅ Keep README up to date
- ✅ Comment complex logic
- ✅ Maintain CHANGELOG

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Vue 3 Docs](https://vuejs.org/)
- [Testing Best Practices](.context/docs/testing-strategy.md)
- [MCP Guide](docs/development/MCP_SETUP_GUIDE.md)
- [Development Setup](docs/development/setup.md)

## Support

- 📖 [Documentation](docs/)
- 💬 [Discussions](https://github.com/your-org/makoclaw/discussions)
- 🐛 [Issues](https://github.com/your-org/makoclaw/issues)
- ✉️ [Email](mailto:support@example.com)

---

**Happy coding! 🚀**

Remember: Quality code is not about writing perfect code the first time. It's about having the tools, processes, and mindset to continuously improve your codebase.
