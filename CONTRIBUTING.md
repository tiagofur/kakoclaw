# Contributing to MakoClaw

Thank you for your interest in contributing! This document provides guidelines for contributing to MakoClaw.

---

## Quick Links

- [Bug Reports](#reporting-bugs)
- [Feature Requests](#suggesting-features)
- [Pull Requests](#pull-requests)
- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Security Issues](#security)

---

## Ways to Contribute

- 🐛 Report bugs
- 💡 Suggest features
- 📝 Improve documentation
- 🔧 Submit pull requests
- 🧪 Write tests
- 📢 Share the project

---

## Getting Started

### 1. Fork and Clone

```bash
git clone https://github.com/YOUR_USERNAME/makoclaw.git
cd makoclaw
git remote add upstream https://github.com/sipeed/makoclaw.git
```

### 2. Development Setup

```bash
# Install dependencies
make deps

# Build
make build

# Run tests
go test ./...

# Verify installation
./makoclaw version
```

See [docs/development/setup.md](docs/development/setup.md) for detailed instructions.

### 3. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-description
```

---

## Reporting Bugs

Before reporting, please:
1. Search [existing issues](https://github.com/sipeed/makoclaw/issues)
2. Verify you're using the latest version
3. Try to reproduce in a clean environment

### Bug Report Template

```markdown
**Description**
Clear description of the bug.

**To Reproduce**
1. Run command '...'
2. Send message '...'
3. See error

**Expected Behavior**
What should happen.

**Environment**
- OS: [e.g., Ubuntu 22.04, Windows 11]
- Version: [e.g., 0.9.0]
- Go Version: [e.g., 1.26]

**Logs**
```
Relevant log output (remove API keys!)
```

**Config (without secrets)**
```json
{
  "agents": { ... }
}
```
```

---

## Suggesting Features

Check [PRD-NEW-FEATURES.md](PRD-NEW-FEATURES.md) for planned features.

### Feature Request Template

```markdown
**Problem**
What problem does this solve?

**Proposed Solution**
Clear description of the feature.

**Alternatives Considered**
Other approaches you've thought about.

**Additional Context**
Screenshots, mockups, etc.
```

---

## Pull Requests

### Before Submitting

1. **Tests pass**
   ```bash
   go test ./...
   go test -race ./...
   ```

2. **Code is formatted**
   ```bash
   go fmt ./...
   go vet ./...
   ```

3. **No security issues**
   ```bash
   make security
   ```

4. **Documentation updated** (if applicable)

### PR Process

1. Create branch from `main`
2. Make changes with atomic commits
3. Push to your fork
4. Create PR with template below
5. Wait for review (24-48h)

### PR Template

```markdown
## Description
Brief description of changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation

## Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
- [ ] Security considerations addressed
- [ ] Follows code style guide

## Related Issues
Fixes #123
```

---

## Development Setup

### Prerequisites

- Go 1.26+
- Node.js 18+ (for frontend)
- Make

### Build Commands

```bash
make build            # Build binary
make build-frontend   # Build Vue frontend
make build-all        # All platforms
make test             # Run tests
make fmt              # Format code
make vet              # Static analysis
make security         # Security scan
```

### Frontend Development

```bash
cd pkg/web/frontend
npm install
npm run dev    # Dev server with HMR
npm run build  # Production build
npm test       # E2E tests
```

---

## Code Style

### Go Conventions

```go
// Package tools provides agent tool implementations.
package tools

// ReadFileTool reads files from the workspace.
// It supports both relative and absolute paths.
type ReadFileTool struct {
    workspace string
    restrict  bool
}

// NewReadFileTool creates a new ReadFileTool instance.
// If restrict is true, paths are limited to workspace.
func NewReadFileTool(workspace string, restrict bool) *ReadFileTool {
    return &ReadFileTool{
        workspace: workspace,
        restrict:  restrict,
    }
}

// Execute reads the file and returns its content.
func (t *ReadFileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    // Implementation
}
```

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Packages | lowercase | `tools`, `agent` |
| Exported | PascalCase | `ReadFile`, `NewAgent` |
| Private | camelCase | `readInternal` |
| Constants | PascalCase/UPPER | `MaxRetries`, `DEFAULT_TIMEOUT` |
| Interfaces | -er suffix | `Reader`, `Writer` |

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting (no code change)
- `refactor`: Code restructuring
- `perf`: Performance improvement
- `test`: Adding tests
- `chore`: Build, deps, etc.

**Examples:**

```bash
feat(tools): add PDF extraction tool

fix(agent): resolve goroutine leak in orchestrator

docs(readme): update installation instructions

feat(channels)!: change WhatsApp message format

BREAKING CHANGE: WhatsApp messages now require bridge v2
```

---

## Testing

### Unit Tests

```go
func TestReadFileTool_Execute(t *testing.T) {
    tests := []struct {
        name    string
        path    string
        want    string
        wantErr bool
    }{
        {
            name:    "valid file",
            path:    "test.txt",
            want:    "content",
            wantErr: false,
        },
        {
            name:    "missing file",
            path:    "nonexistent.txt",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tool := NewReadFileTool("/tmp", true)
            got, err := tool.Execute(context.Background(), map[string]interface{}{
                "path": tt.path,
            })
            if (err != nil) != tt.wantErr {
                t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("Execute() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Security

### Reporting Vulnerabilities

**DO NOT** open public issues for security vulnerabilities.

1. Email: security@makoclaw.dev
2. Include detailed description
3. Include reproduction steps
4. Suggest fix if possible

See [SECURITY.md](SECURITY.md) for full policy.

### Security Checklist

Before submitting PRs:

- [ ] No hardcoded credentials
- [ ] No `fmt.Printf` with sensitive data
- [ ] SQL uses parameterized queries
- [ ] File paths validated against traversal
- [ ] Shell commands go through allowlist
- [ ] Error messages don't leak internals

---

## Priority Areas

### High Priority
1. Tests - Increase coverage
2. Documentation - API docs, guides
3. Channels - New platform support
4. Providers - More LLM providers
5. Performance - Memory/CPU optimization

### Known Issues
See [BUGS-KNOWN-ISSUES.md](BUGS-KNOWN-ISSUES.md) for issues to fix.

### Planned Features
See [PRD-NEW-FEATURES.md](PRD-NEW-FEATURES.md) for roadmap.

---

## Recognition

Contributors are recognized in:
- [CHANGELOG.md](CHANGELOG.md)
- Release notes
- Documentation

---

## Questions?

- **GitHub Issues**: Bugs and features
- **GitHub Discussions**: Questions
- **Discord**: [Join our server](https://discord.gg/V4sAZ9XWpN)

---

## Code of Conduct

- Be respectful and constructive
- Accept constructive criticism
- Focus on what's best for the community
- Show empathy toward others

---

Thank you for contributing! 🦈
