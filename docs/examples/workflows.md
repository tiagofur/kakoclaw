# Workflows - Automation Pipelines

> **Status**: ✅ **Fully Functional** | **Feature Phase**: MVP Complete | **UI**: Web Panel Only | **Platform**: 🦈 MakoClaw

**Workflows** enable you to create multi-step automation pipelines that combine LLM prompts, tool executions, and conditional logic. Perfect for repetitive tasks, complex research, code analysis, content generation, and more.

---

## Table of Contents

- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Step Types Reference](#step-types-reference)
- [Template Variables](#template-variables)
- [Error Handling](#error-handling)
- [Examples for Programmers](#examples-for-programmers)
- [Examples for Technical Writers](#examples-for-technical-writers)
- [Best Practices](#best-practices)
- [Limitations](#limitations)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

### Access Workflows

1. Start MakoClaw web panel:
    ```bash
    MakoClaw web
    # or
    MakoClaw gateway
    ```

2. Navigate to **http://localhost:18880** (or your configured port)

3. Log in with your credentials

4. Click **"Workflows"** in the left sidebar

5. Click **"New Workflow"** button

### Your First Workflow: Simple Research Assistant

Let's create a workflow that searches the web and summarizes results:

**Step 1: Access the workflow editor**
1. Make sure you're on the Workflows page
2. Click the **"New Workflow"** button (green button, top-right corner)
3. You'll enter the workflow editor view

**Step 2: Name your workflow**
- **Name**: "Quick Research"
- **Description**: "Search and summarize a topic"

**Step 3: Add your first step**

Now you'll see three buttons below the workflow name:
- `+ Prompt` (blue)
- `+ Tool` (amber)
- `+ Condition` (green)

Click **"+ Tool"** button:
- **Label**: "Search web"
- **Tool Name**: `web_search`
- **Arguments (JSON)**:
  ```json
  {
    "query": "latest advances in Go programming language 2026"
  }
  ```
- **On Error**: Stop

Click **"+ Prompt"** button to add a second step:
- **Label**: "Summarize results"
- **Message**: "Please summarize these search results in 3 key points:\n\n{{step.1.output}}"
- **Model**: (leave empty for default)
- **On Error**: Stop

**Step 4: Save and test**

1. Click **"Save"** button (top-right corner)
2. Once saved, click **"Test Run"** button (in the Test Run section below the steps)

You'll see execution results appear in the "Test Results" panel, showing:
- Each step's output
- Execution time
- Any errors

🎉 **Congratulations!** You've created your first workflow.

---

## Core Concepts

### What is a Workflow?

A **workflow** is a sequence of steps executed in order. Each step can:
- Send prompts to the LLM
- Execute tools (filesystem, shell, web search, etc.)
- Evaluate conditions to control flow
- Use outputs from previous steps

### Workflow Anatomy

```
Workflow: "Code Review Bot"
├─ Step 1 (tool): Read file
├─ Step 2 (prompt): Analyze code
├─ Step 3 (condition): Check if issues found
└─ Step 4 (tool): Create task (runs only if step 3 = true)
```

### Execution Model

- **Sequential**: Steps run one after another
- **Blocking**: Each step waits for the previous to complete
- **Stateful**: Later steps can reference earlier outputs
- **Timeout**: Maximum 5 minutes per workflow execution
- **Session**: Each run creates an isolated session (`workflow:ID:run:RUN_ID`)

---

## Step Types Reference

### 1. Prompt Step

Sends a message to the LLM and captures the response.

**Fields:**
- **Label**: Human-readable step name (e.g., "Analyze code")
- **Message**: The prompt to send to the LLM
- **Model**: (Optional) Override default model (e.g., `claude-3-5-sonnet-20241022`)
- **On Error**: `stop` or `continue`

**Example:**
```json
{
  "id": "step_1234567890_1",
  "type": "prompt",
  "label": "Generate ideas",
  "on_error": "stop",
  "config": {
    "message": "Generate 5 creative project ideas for a Go developer",
    "model": ""
  }
}
```

**Use Cases:**
- Content generation
- Code analysis
- Summarization
- Question answering
- Translation

---

### 2. Tool Step

Executes a registered tool directly (bypasses LLM function calling).

**Fields:**
- **Label**: Human-readable step name
- **Tool Name**: Exact tool identifier (see available tools with `/api/v1/tools`)
- **Arguments**: JSON object with tool-specific parameters
- **On Error**: `stop` or `continue`

**Available Tools:**
- `read_file` - Read file contents
- `write_file` - Write file (overwrites)
- `append_file` - Append to file
- `delete_file` - Delete file
- `list_directory` - List directory contents
- `shell_exec` - Execute shell command
- `web_search` - Search the web
- `web_fetch` - Fetch webpage content
- `send_message` - Send message to channel
- `create_task` - Create task in task board
- `update_task` - Update existing task
- `knowledge_add` - Add entry to knowledge base
- `knowledge_search` - Search knowledge base

**Example:**
```json
{
  "id": "step_1234567890_2",
  "type": "tool",
  "label": "Run tests",
  "on_error": "continue",
  "config": {
    "tool_name": "shell_exec",
    "args": {
      "command": "go test ./... -v",
      "timeout": 60
    }
  }
}
```

**Use Cases:**
- File operations
- Command execution
- Web scraping
- Database operations (via shell)
- API calls (via web_fetch)
- Task automation

---

### 3. Condition Step

Evaluates a condition and skips the next non-condition steps if false.

**Fields:**
- **Label**: Human-readable step name
- **Operator**: `contains`, `equals`, `not_empty`, `regex`
- **Reference**: Template variable to evaluate (e.g., `{{step.1.output}}`)
- **Value**: Comparison value
- **On Error**: `stop` or `continue`

**Operators:**
- **contains**: Case-insensitive substring match
- **equals**: Exact match (trimmed whitespace)
- **not_empty**: Checks if reference is non-empty after trimming
- **regex**: Matches against regular expression

**Example:**
```json
{
  "id": "step_1234567890_3",
  "type": "condition",
  "label": "Check if tests failed",
  "on_error": "stop",
  "config": {
    "operator": "contains",
    "reference": "{{step.2.output}}",
    "value": "FAIL"
  }
}
```

**Behavior:**
- If condition is **TRUE**: Continue to next step
- If condition is **FALSE**: Skip all following steps until next condition step or end of workflow
- The step result shows: `"condition: true — continuing"` or `"condition: false — skipping next steps"`

**Use Cases:**
- Conditional notifications (only alert on errors)
- Branching logic (different actions based on tool output)
- Validation gates (stop if check fails)
- Error detection (check for specific patterns)

---

## Template Variables

### Syntax

Use `{{step.N.output}}` or `{{step.N.error}}` to reference previous step results.

- **N**: Step number (1-based, first step = 1)
- **output**: The successful output of the step
- **error**: The error message (if step failed)

### Examples

```
Message: "Summarize this: {{step.1.output}}"
→ Replaces with: "Summarize this: [actual output from step 1]"

Tool arg: { "filename": "result_{{step.2.output}}.txt" }
→ Replaces with: { "filename": "result_success.txt" }

Condition: "{{step.3.error}}"
→ Replaces with error message (or empty if no error)
```

### Interpolation Rules

1. **Works in**: Prompt messages, tool argument strings, condition references/values
2. **Does NOT work in**: Tool names, step IDs, on_error policies
3. **Steps are 1-indexed**: First step is `{{step.1.output}}`
4. **Empty if missing**: References to non-existent steps return empty string
5. **Recursive**: Cannot nest template variables

### Advanced Example

```json
{
  "type": "prompt",
  "config": {
    "message": "Compare these two outputs:\n\nFirst: {{step.1.output}}\n\nSecond: {{step.3.output}}\n\nErrors: {{step.2.error}}"
  }
}
```

---

## Error Handling

### Per-Step Policy

Each step has an **on_error** field:

**`stop` (default)**: 
- Stops workflow immediately on error
- Sets workflow status to `failed`
- No further steps execute
- Useful for critical steps

**`continue`**:
- Logs error but continues to next step
- Step result includes `"error": "message"`
- Workflow completes (status: `completed_with_errors`)
- Useful for optional steps

### Workflow Status

After execution, workflows have one of these statuses:

- **`running`**: Currently executing
- **`completed`**: All steps succeeded
- **`completed_with_errors`**: Some steps failed but continued
- **`failed`**: Stopped due to error or timeout

### Example Strategy

```json
[
  {
    "label": "Critical: Read config",
    "on_error": "stop"
  },
  {
    "label": "Optional: Check for updates",
    "on_error": "continue"
  },
  {
    "label": "Critical: Process data",
    "on_error": "stop"
  }
]
```

---

## Examples for Programmers

### 1. Automated Code Review

**Use Case**: Analyze a file for code quality issues and create a task if problems found.

**Steps:**
1. Read source file
2. Send to LLM for review
3. Check if "critical" or "bug" in response
4. Create high-priority task if issues found

**JSON:**
```json
{
  "name": "Code Review Bot",
  "description": "Analyzes code files and flags issues",
  "enabled": true,
  "steps": [
    {
      "id": "step_1",
      "type": "tool",
      "label": "Read source file",
      "on_error": "stop",
      "config": {
        "tool_name": "read_file",
        "args": {
          "path": "pkg/agent/loop.go"
        }
      }
    },
    {
      "id": "step_2",
      "type": "prompt",
      "label": "Analyze code quality",
      "on_error": "stop",
      "config": {
        "message": "Review this Go code for:\n- Potential bugs\n- Performance issues\n- Security vulnerabilities\n- Code smells\n\nCode:\n```go\n{{step.1.output}}\n```\n\nRespond with 'CRITICAL' if bugs/security issues found, otherwise 'OK'.",
        "model": ""
      }
    },
    {
      "id": "step_3",
      "type": "condition",
      "label": "Check if issues found",
      "on_error": "stop",
      "config": {
        "operator": "contains",
        "reference": "{{step.2.output}}",
        "value": "CRITICAL"
      }
    },
    {
      "id": "step_4",
      "type": "tool",
      "label": "Create task for issues",
      "on_error": "continue",
      "config": {
        "tool_name": "create_task",
        "args": {
          "title": "Code Review - Critical Issues Found",
          "description": "{{step.2.output}}",
          "status": "todo",
          "priority": "high"
        }
      }
    }
  ]
}
```

**Expected Result:**
- If critical issues: Task created in todo board
- If no issues: Workflow completes, no task created

---

### 2. Test Failure Analyzer

**Use Case**: Run tests, analyze failures with LLM, save report.

```json
{
  "name": "Test Failure Analyzer",
  "description": "Runs tests and analyzes failures",
  "enabled": true,
  "steps": [
    {
      "id": "step_1",
      "type": "tool",
      "label": "Run test suite",
      "on_error": "continue",
      "config": {
        "tool_name": "shell_exec",
        "args": {
          "command": "go test ./... -v",
          "timeout": 120
        }
      }
    },
    {
      "id": "step_2",
      "type": "condition",
      "label": "Check if tests failed",
      "on_error": "stop",
      "config": {
        "operator": "contains",
        "reference": "{{step.1.output}}",
        "value": "FAIL"
      }
    },
    {
      "id": "step_3",
      "type": "prompt",
      "label": "Analyze failures",
      "on_error": "stop",
      "config": {
        "message": "Analyze these test failures and suggest fixes:\n\n{{step.1.output}}\n\nProvide:\n1. Root cause\n2. Suggested fix\n3. Priority level",
        "model": ""
      }
    },
    {
      "id": "step_4",
      "type": "tool",
      "label": "Save analysis report",
      "on_error": "continue",
      "config": {
        "tool_name": "write_file",
        "args": {
          "path": "test-analysis-report.md",
          "content": "# Test Failure Analysis\n\n## Test Output\n\n{{step.1.output}}\n\n## AI Analysis\n\n{{step.3.output}}"
        }
      }
    }
  ]
}
```

---

### 3. Dependency Vulnerability Scanner

**Use Case**: Check for outdated/vulnerable dependencies.

```json
{
  "name": "Dependency Security Check",
  "description": "Scans for vulnerable dependencies",
  "enabled": true,
  "steps": [
    {
      "id": "step_1",
      "type": "tool",
      "label": "List Go dependencies",
      "on_error": "stop",
      "config": {
        "tool_name": "shell_exec",
        "args": {
          "command": "go list -m all",
          "timeout": 30
        }
      }
    },
    {
      "id": "step_2",
      "type": "tool",
      "label": "Check for vulnerabilities",
      "on_error": "continue",
      "config": {
        "tool_name": "shell_exec",
        "args": {
          "command": "govulncheck ./...",
          "timeout": 60
        }
      }
    },
    {
      "id": "step_3",
      "type": "prompt",
      "label": "Generate security report",
      "on_error": "stop",
      "config": {
        "message": "Create a security report from this vulnerability scan:\n\n{{step.2.output}}\n\nInclude:\n- Critical vulnerabilities (if any)\n- Recommended actions\n- Priority order",
        "model": ""
      }
    },
    {
      "id": "step_4",
      "type": "condition",
      "label": "Check if vulnerabilities found",
      "on_error": "stop",
      "config": {
        "operator": "contains",
        "reference": "{{step.2.output}}",
        "value": "vulnerability"
      }
    },
    {
      "id": "step_5",
      "type": "tool",
      "label": "Alert team",
      "on_error": "continue",
      "config": {
        "tool_name": "send_message",
        "args": {
          "channel": "telegram:default",
          "message": "⚠️ Security vulnerabilities detected!\n\n{{step.3.output}}"
        }
      }
    }
  ]
}
```

---

### 4. Documentation Generator

**Use Case**: Read code file, generate documentation, save to docs folder.

```json
{
  "name": "Auto Documentation",
  "description": "Generates markdown docs from source code",
  "enabled": true,
  "steps": [
    {
      "id": "step_1",
      "type": "tool",
      "label": "Read source file",
      "on_error": "stop",
      "config": {
        "tool_name": "read_file",
        "args": {
          "path": "pkg/workflow/engine.go"
        }
      }
    },
    {
      "id": "step_2",
      "type": "prompt",
      "label": "Generate documentation",
      "on_error": "stop",
      "config": {
        "message": "Create comprehensive documentation for this Go code:\n\n```go\n{{step.1.output}}\n```\n\nInclude:\n- Overview\n- Public API (functions/types)\n- Usage examples\n- Notes/caveats\n\nFormat as markdown.",
        "model": ""
      }
    },
    {
      "id": "step_3",
      "type": "tool",
      "label": "Save documentation",
      "on_error": "stop",
      "config": {
        "tool_name": "write_file",
        "args": {
          "path": "docs/api-reference/workflow-engine.md",
          "content": "{{step.2.output}}"
        }
      }
    }
  ]
}
```

---

### 5. PR Description Generator

**Use Case**: Generate detailed PR description from git diff.

```json
{
  "name": "PR Description Generator",
  "description": "Creates PR descriptions from git changes",
  "enabled": true,
  "steps": [
    {
      "id": "step_1",
      "type": "tool",
      "label": "Get git diff",
      "on_error": "stop",
      "config": {
        "tool_name": "shell_exec",
        "args": {
          "command": "git diff main..HEAD",
          "timeout": 10
        }
      }
    },
    {
      "id": "step_2",
      "type": "prompt",
      "label": "Generate PR description",
      "on_error": "stop",
      "config": {
        "message": "Create a PR description from this diff:\n\n{{step.1.output}}\n\nFormat:\n## Summary\n\n## Changes\n\n## Testing\n\n## Notes",
        "model": ""
      }
    },
    {
      "id": "step_3",
      "type": "tool",
      "label": "Save to file",
      "on_error": "continue",
      "config": {
        "tool_name": "write_file",
        "args": {
          "path": "PR_DESCRIPTION.md",
          "content": "{{step.2.output}}"
        }
      }
    }
  ]
}
```

---

## Examples for Technical Writers

### 1. Research Assistant

**Use Case**: Multi-source research on a topic with synthesis.

```json
{
  "name": "Research Assistant",
  "description": "Comprehensive research workflow",
  "enabled": true,
  "steps": [
    {
      "id": "step_1",
      "type": "prompt",
      "label": "Generate search queries",
      "on_error": "stop",
      "config": {
        "message": "Generate 3 targeted search queries to research: 'AI-powered code review tools 2026'\n\nReturn only the queries, one per line.",
        "model": ""
      }
    },
    {
      "id": "step_2",
      "type": "tool",
      "label": "Search query 1",
      "on_error": "continue",
      "config": {
        "tool_name": "web_search",
        "args": {
          "query": "{{step.1.output}}",
          "max_results": 5
        }
      }
    },
    {
      "id": "step_3",
      "type": "prompt",
      "label": "Synthesize findings",
      "on_error": "stop",
      "config": {
        "message": "Create a research summary from these sources:\n\n{{step.2.output}}\n\nFormat:\n## Key Findings\n## Trends\n## Sources",
        "model": ""
      }
    },
    {
      "id": "step_4",
      "type": "tool",
      "label": "Save research notes",
      "on_error": "continue",
      "config": {
        "tool_name": "append_file",
        "args": {
          "path": "research-notes.md",
          "content": "\n\n---\n\n# Research: AI Code Review Tools\n\nDate: {{step.3.output}}\n\n{{step.3.output}}"
        }
      }
    }
  ]
}
```

---

### 2. Article Fact Checker

**Use Case**: Verify claims in article draft against web sources.

```json
{
  "name": "Fact Checker",
  "description": "Verifies claims in articles",
  "enabled": true,
  "steps": [
    {
      "id": "step_1",
      "type": "tool",
      "label": "Read article draft",
      "on_error": "stop",
      "config": {
        "tool_name": "read_file",
        "args": {
          "path": "drafts/my-article.md"
        }
      }
    },
    {
      "id": "step_2",
      "type": "prompt",
      "label": "Extract factual claims",
      "on_error": "stop",
      "config": {
        "message": "Extract all factual claims from this article that should be verified:\n\n{{step.1.output}}\n\nReturn as numbered list.",
        "model": ""
      }
    },
    {
      "id": "step_3",
      "type": "tool",
      "label": "Search for verification",
      "on_error": "continue",
      "config": {
        "tool_name": "web_search",
        "args": {
          "query": "verify: {{step.2.output}}",
          "max_results": 10
        }
      }
    },
    {
      "id": "step_4",
      "type": "prompt",
      "label": "Generate verification report",
      "on_error": "stop",
      "config": {
        "message": "Compare claims with sources:\n\nClaims:\n{{step.2.output}}\n\nSources:\n{{step.3.output}}\n\nMark each claim as:\n✓ Verified\n? Needs review\n✗ Disputed",
        "model": ""
      }
    },
    {
      "id": "step_5",
      "type": "tool",
      "label": "Save verification report",
      "on_error": "continue",
      "config": {
        "tool_name": "write_file",
        "args": {
          "path": "fact-check-report.md",
          "content": "# Fact Check Report\n\n{{step.4.output}}"
        }
      }
    }
  ]
}
```

---

### 3. SEO Optimizer

**Use Case**: Analyze article for SEO, suggest improvements.

```json
{
  "name": "SEO Optimizer",
  "description": "Optimizes articles for search engines",
  "enabled": true,
  "steps": [
    {
      "id": "step_1",
      "type": "tool",
      "label": "Read article",
      "on_error": "stop",
      "config": {
        "tool_name": "read_file",
        "args": {
          "path": "blog/my-post.md"
        }
      }
    },
    {
      "id": "step_2",
      "type": "prompt",
      "label": "Analyze SEO",
      "on_error": "stop",
      "config": {
        "message": "Analyze this article for SEO:\n\n{{step.1.output}}\n\nProvide:\n1. Primary keyword suggestions\n2. Title improvements\n3. Meta description (150-160 chars)\n4. Header structure review\n5. Internal linking opportunities",
        "model": ""
      }
    },
    {
      "id": "step_3",
      "type": "tool",
      "label": "Search competitor content",
      "on_error": "continue",
      "config": {
        "tool_name": "web_search",
        "args": {
          "query": "top ranking articles about {{step.1.output}}",
          "max_results": 5
        }
      }
    },
    {
      "id": "step_4",
      "type": "prompt",
      "label": "Generate recommendations",
      "on_error": "stop",
      "config": {
        "message": "Create SEO improvement plan:\n\nCurrent: {{step.2.output}}\n\nCompetitors: {{step.3.output}}\n\nProvide prioritized action items.",
        "model": ""
      }
    },
    {
      "id": "step_5",
      "type": "tool",
      "label": "Save SEO report",
      "on_error": "continue",
      "config": {
        "tool_name": "write_file",
        "args": {
          "path": "seo-recommendations.md",
          "content": "{{step.4.output}}"
        }
      }
    }
  ]
}
```

---

### 4. Content Outline Generator

**Use Case**: Generate article outline from topic with research.

```json
{
  "name": "Outline Generator",
  "description": "Creates article outlines with research",
  "enabled": true,
  "steps": [
    {
      "id": "step_1",
      "type": "tool",
      "label": "Research topic",
      "on_error": "continue",
      "config": {
        "tool_name": "web_search",
        "args": {
          "query": "comprehensive guide to Go concurrency patterns",
          "max_results": 10
        }
      }
    },
    {
      "id": "step_2",
      "type": "prompt",
      "label": "Generate outline",
      "on_error": "stop",
      "config": {
        "message": "Create a detailed article outline for 'Go Concurrency Patterns Guide' using this research:\n\n{{step.1.output}}\n\nInclude:\n- Introduction\n- 5-7 main sections\n- Subsections\n- Key points to cover\n- Estimated word count per section",
        "model": ""
      }
    },
    {
      "id": "step_3",
      "type": "prompt",
      "label": "Generate title suggestions",
      "on_error": "continue",
      "config": {
        "message": "Generate 5 compelling titles for this outline:\n\n{{step.2.output}}\n\nMake them:\n- SEO-friendly\n- Engaging\n- Clear about value",
        "model": ""
      }
    },
    {
      "id": "step_4",
      "type": "tool",
      "label": "Save outline",
      "on_error": "stop",
      "config": {
        "tool_name": "write_file",
        "args": {
          "path": "outlines/go-concurrency-guide.md",
          "content": "# Article Outline\n\n## Title Options\n\n{{step.3.output}}\n\n## Outline\n\n{{step.2.output}}"
        }
      }
    }
  ]
}
```

---

### 5. Multi-language Publisher

**Use Case**: Translate article and adapt for different regions.

```json
{
  "name": "Multi-language Publisher",
  "description": "Translates and localizes content",
  "enabled": true,
  "steps": [
    {
      "id": "step_1",
      "type": "tool",
      "label": "Read English article",
      "on_error": "stop",
      "config": {
        "tool_name": "read_file",
        "args": {
          "path": "blog/en/my-article.md"
        }
      }
    },
    {
      "id": "step_2",
      "type": "prompt",
      "label": "Translate to Spanish",
      "on_error": "stop",
      "config": {
        "message": "Translate this article to Spanish, adapting cultural references:\n\n{{step.1.output}}",
        "model": ""
      }
    },
    {
      "id": "step_3",
      "type": "tool",
      "label": "Save Spanish version",
      "on_error": "continue",
      "config": {
        "tool_name": "write_file",
        "args": {
          "path": "blog/es/my-article.md",
          "content": "{{step.2.output}}"
        }
      }
    },
    {
      "id": "step_4",
      "type": "prompt",
      "label": "Translate to French",
      "on_error": "stop",
      "config": {
        "message": "Translate this article to French, adapting cultural references:\n\n{{step.1.output}}",
        "model": ""
      }
    },
    {
      "id": "step_5",
      "type": "tool",
      "label": "Save French version",
      "on_error": "continue",
      "config": {
        "tool_name": "write_file",
        "args": {
          "path": "blog/fr/my-article.md",
          "content": "{{step.4.output}}"
        }
      }
    }
  ]
}
```

---

## Best Practices

### 1. Design for Failure

- Set `on_error: "continue"` for optional steps (logging, notifications)
- Set `on_error: "stop"` for critical steps (config reading, data processing)
- Always test error scenarios

### 2. Keep Steps Focused

- One clear action per step
- Label steps descriptively
- Avoid monolithic prompts

### 3. Optimize for Speed

- Minimize file I/O
- Use specific prompts (faster responses)
- Avoid redundant searches

### 4. Use Conditions Wisely

- Place conditions before expensive operations
- Test condition logic thoroughly
- Remember: False conditions skip ALL following steps

### 5. Template Variable Hygiene

- Verify step numbers are correct
- Test with empty/error outputs
- Escape special characters if needed

### 6. Workflow Documentation

- Use descriptive names
- Write detailed descriptions
- Document expected inputs/outputs

### 7. Iterative Development

1. Start with 2-3 steps
2. Test thoroughly
3. Add more steps incrementally
4. Use "Test Run" frequently

---

## Limitations

### Current Limitations

⚠️ **Single-User Context**: Workflows execute in a shared agent context. Not yet isolated per-user.

⚠️ **Sequential Only**: No parallel execution of steps.

⚠️ **5-Minute Timeout**: Workflows automatically fail after 5 minutes.

⚠️ **No Manual Resume**: Failed workflows must restart from beginning.

⚠️ **No Scheduling**: Manual trigger only (cron integration planned).

⚠️ **No Versioning**: Edits overwrite previous versions.

⚠️ **No Import/Export**: No built-in way to share workflows (yet).

### Workarounds

**Long-running tasks**: Split into multiple smaller workflows

**Scheduled execution**: Use system cron to call API endpoint

**Workflow sharing**: Copy/paste JSON via REST API

---

## Troubleshooting

### Workflow Times Out

**Symptom**: Workflow fails after ~5 minutes

**Solutions**:
- Split into smaller workflows
- Reduce prompt complexity
- Optimize tool operations
- Check for infinite loops in conditions

### "Tool Not Found" Error

**Symptom**: Step fails with "tool not found"

**Solutions**:
- Verify tool name exactly matches registry (check `/api/v1/tools`)
- Check spelling and capitalization
- Confirm tool is enabled in your configuration

### Template Variable Empty

**Symptom**: `{{step.N.output}}` resolves to empty string

**Solutions**:
- Verify step N completed successfully
- Check step number is correct (1-based)
- Review step N's output in test results
- Confirm step N didn't fail silently

### Condition Not Working

**Symptom**: Condition doesn't skip as expected

**Solutions**:
- Test operator with simple values first
- Check string trimming/casing
- Verify reference points to correct step
- Use `contains` instead of `equals` for partial matching

### Steps Execute Even After False Condition

**Symptom**: Steps run despite condition being false

**Solutions**:
- Remember: Only non-condition steps are skipped
- Condition steps themselves always execute
- Check condition operator and value

### Workflow Stuck "Running"

**Symptom**: Workflow shows "running" indefinitely

**Solutions**:
- Likely a backend issue
- Check backend logs: `/tmp/makoclaw.log`
- Restart MakoClaw service
- Report as bug if persists

---

## API Reference

For programmatic workflow management, see [API Reference](../api-reference/workflows.md).

---

## Pre-built Templates

See [workflow-templates.json](./workflow-templates.json) for ready-to-use workflow examples you can import.

---

## Feedback & Support

- **Issue Tracker**: https://github.com/sipeed/makoclaw/issues
- **Documentation**: https://github.com/sipeed/makoclaw/tree/main/docs
- **Discussions**: https://github.com/sipeed/makoclaw/discussions

---

**Version**: 1.0.0  
**Last Updated**: February 2026  
**Status**: Production Ready (MVP)
