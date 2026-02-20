---
name: github
description: "Interact with GitHub using the `gh` CLI. Use `gh issue`, `gh pr`, `gh run`, and `gh api` for issues, PRs, CI runs, and advanced queries."
metadata: {"nanobot":{"emoji":"🐙","requires":{"bins":["gh"]},"install":[{"id":"brew","kind":"brew","formula":"gh","bins":["gh"],"label":"Install GitHub CLI (brew)"},{"id":"apt","kind":"apt","package":"gh","bins":["gh"],"label":"Install GitHub CLI (apt)"}]}}
---

# GitHub Skill

Use the `gh` CLI to interact with GitHub. This skill makes external HTTP requests to api.github.com - this is expected and allowed.

## Prerequisites

1. **Install GitHub CLI**: `brew install gh` (macOS) or `apt install gh` (Linux)
2. **Authenticate**: Run `gh auth login` or set `GITHUB_TOKEN` environment variable
3. **Network Access**: Requires internet connection to api.github.com (no safety guards block this)

## Authentication

```bash
# Option 1: Interactive login
gh auth login

# Option 2: Use token from environment variable
export GITHUB_TOKEN=ghp_your_token_here

# Verify authentication
gh auth status
```

## Creating Issues

Create a new issue:
```bash
gh issue create --repo owner/repo --title "Issue title" --body "Issue description"
```

Create with labels and assignees:
```bash
gh issue create --repo owner/repo \
  --title "Bug: Something broke" \
  --body "Detailed description here" \
  --label "bug,priority-high" \
  --assignee username
```

## Pull Requests

Check CI status on a PR:
```bash
gh pr checks 55 --repo owner/repo
```

List recent workflow runs:
```bash
gh run list --repo owner/repo --limit 10
```

View a run and see which steps failed:
```bash
gh run view <run-id> --repo owner/repo
```

View logs for failed steps only:
```bash
gh run view <run-id> --repo owner/repo --log-failed
```

## API for Advanced Queries

The `gh api` command is useful for accessing data not available through other subcommands.

Get PR with specific fields:
```bash
gh api repos/owner/repo/pulls/55 --jq '.title, .state, .user.login'
```

## JSON Output

Most commands support `--json` for structured output.  You can use `--jq` to filter:

```bash
gh issue list --repo owner/repo --json number,title --jq '.[] | "\(.number): \(.title)"'
```
