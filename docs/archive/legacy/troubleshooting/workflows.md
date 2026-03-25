# Workflows Troubleshooting

Common issues when working with MakoClaw workflows and their solutions.

---

## Workflow Times Out

### Symptom
Workflow fails after approximately 5 minutes with timeout error.

### Cause
Server enforces a 5-minute maximum execution time per workflow run.

### Solutions

**1. Split into smaller workflows**
```json
// Instead of one long workflow
"Big Workflow" → [20 steps]

// Split into multiple
"Part 1" → [7 steps]
"Part 2" → [7 steps]
"Part 3" → [6 steps]
```

**2. Optimize expensive operations**
- Reduce `max_results` in web searches
- Use more specific prompts (faster LLM responses)
- Minimize file I/O operations
- Avoid redundant steps

**3. Use tool steps instead of prompts when possible**
- Tool steps are faster than LLM prompts
- Direct tool execution bypasses function calling overhead

**Example:**
```json
// Slower (goes through LLM)
{
  "type": "prompt",
  "config": {
    "message": "List files in current directory"
  }
}

// Faster (direct tool call)
{
  "type": "tool",
  "config": {
    "tool_name": "list_directory",
    "args": {"path": "."}
  }
}
```

---

## "Tool Not Found" Error

### Symptom
Step fails with error: `tool not found: <tool_name>`

### Cause
Tool name doesn't match registered tools in the system.

### Solutions

**1. Verify tool name**
```bash
# List available tools via API
curl http://localhost:18880/api/v1/tools

# Or check in web panel
# Navigate to: Tools → Available Tools
```

**2. Common tool name mistakes**
```
❌ Wrong          ✅ Correct
read-file        read_file
writeFile        write_file
shell            shell_exec
search           web_search
```

**3. Check tool is enabled**
Some tools require configuration in `config.json`:
```json
{
  "tools": {
    "web": {
      "search": {
        "api_key": "your-brave-api-key"
      }
    }
  }
}
```

**4. Verify tool exists in your version**
```bash
# Check MakoClaw version
./build/makoclaw version

# Update to latest
git pull && make build
```

---

## Template Variable is Empty

### Symptom
`{{step.N.output}}` resolves to empty string in workflow execution.

### Cause
- Step N failed silently
- Step number is incorrect (off-by-one error)
- Referenced step hasn't executed yet
- Step output was actually empty

### Solutions

**1. Verify step numbering (1-based, not 0-based)**
```json
// Steps are numbered starting from 1
Step 1: id="step_1" → Reference: {{step.1.output}}
Step 2: id="step_2" → Reference: {{step.2.output}}
Step 3: id="step_3" → Reference: {{step.3.output}}
```

**2. Check previous step succeeded**
```json
// Add on_error: "stop" to critical steps
{
  "id": "step_1",
  "type": "tool",
  "on_error": "stop",  // ← Workflow stops if this fails
  "config": {...}
}
```

**3. Inspect test results**
- Run workflow via "Test Run" button
- Expand each step in results panel
- Check "output" and "error" fields
- Verify step N actually has output

**4. Handle empty outputs explicitly**
```json
// Check if output exists before using it
{
  "type": "condition",
  "label": "Check step 1 has output",
  "config": {
    "operator": "not_empty",
    "reference": "{{step.1.output}}",
    "value": ""
  }
}
```

---

## Condition Not Working as Expected

### Symptom
Condition step doesn't skip steps when you expect it to.

### Cause
- Operator doesn't match your use case
- Case sensitivity issues
- Whitespace in comparison
- Misunderstanding of condition behavior

### Solutions

**1. Understand condition behavior**
```
condition = TRUE  → Continue to next step
condition = FALSE → Skip all following non-condition steps
```

**2. Use correct operator**

```json
// Case-insensitive substring match
{
  "operator": "contains",
  "reference": "{{step.1.output}}",
  "value": "error"
}

// Exact match (trimmed, case-sensitive)
{
  "operator": "equals",
  "reference": "{{step.1.output}}",
  "value": "SUCCESS"
}

// Check if not empty
{
  "operator": "not_empty",
  "reference": "{{step.1.output}}",
  "value": ""
}

// Regex pattern match
{
  "operator": "regex",
  "reference": "{{step.1.output}}",
  "value": "\\d{3}-\\d{4}"
}
```

**3. Test with simple values first**
```json
// Hardcode values to test logic
{
  "operator": "contains",
  "reference": "test string with error",  // ← Hardcoded for testing
  "value": "error"
}
// Then switch back to template variable
{
  "operator": "contains",
  "reference": "{{step.1.output}}",
  "value": "error"
}
```

**4. Use `contains` for partial matching**
```json
// Checking if output mentions "fail"
{
  "operator": "contains",  // ← More flexible than "equals"
  "reference": "{{step.1.output}}",
  "value": "fail"
}
```

**5. Debug condition output**
Add a prompt step after condition to see what it evaluated:
```json
[
  {"type": "condition", "id": "step_2", ...},
  {
    "type": "prompt",
    "id": "step_3",
    "label": "Debug condition",
    "config": {
      "message": "Previous step output was: {{step.1.output}}"
    }
  }
]
```

---

## Steps Execute After False Condition

### Symptom
Steps continue running even though condition evaluated to false.

### Cause
**Condition steps always execute** — they don't skip themselves.

### Behavior

```json
[
  {"id": "step_1", "type": "prompt", ...},     // ✓ Executes
  {"id": "step_2", "type": "condition", ...},  // ✓ Executes (checks condition)
  // If step_2 = false:
  {"id": "step_3", "type": "tool", ...},       // ✗ SKIPPED
  {"id": "step_4", "type": "prompt", ...},     // ✗ SKIPPED
  {"id": "step_5", "type": "condition", ...},  // ✓ Executes (resets skip state)
  {"id": "step_6", "type": "tool", ...}        // ✓ Executes (if step_5 = true)
]
```

### Solution
Place condition steps **immediately before** the steps you want to conditionally execute:

```json
// Good: Condition right before tool
[
  {"type": "prompt", "label": "Analyze code"},
  {"type": "condition", "label": "Check if bugs found"},
  {"type": "tool", "label": "Create task"}  // ← Only runs if previous condition = true
]

// Bad: Condition far from dependent step
[
  {"type": "condition", "label": "Check if bugs found"},
  {"type": "prompt", "label": "Unrelated step"},  // ← Still executes if condition = false
  {"type": "tool", "label": "Create task"}
]
```

---

## Workflow Stuck "Running"

### Symptom
Workflow shows status "running" indefinitely with no progress.

### Cause
Backend crash, database lock, or network timeout.

### Solutions

**1. Check backend is running**
```bash
# Health check
curl http://localhost:18880/api/v1/health

# Expected: {"status":"ok"}
```

**2. Check backend logs**
```bash
# If running in terminal
# Look for errors in output

# If running as daemon
tail -f /tmp/makoclaw.log

# Look for workflow-related errors
grep -i "workflow\|panic\|error" /tmp/makoclaw.log
```

**3. Restart MakoClaw**
```bash
# Gateway mode
pkill makoclaw
./build/makoclaw gateway

# Web mode
pkill makoclaw
./build/makoclaw web
```

**4. Check database**
```bash
# SQLite database location
ls -lh ~/.MakoClaw/MakoClaw.db

# If corrupted, backup and recreate
cp ~/.MakoClaw/MakoClaw.db ~/.MakoClaw/MakoClaw.db.backup
rm ~/.MakoClaw/MakoClaw.db
# Restart MakoClaw (auto-creates new DB)
```

**5. Report bug**
If issue persists:
1. Capture backend logs
2. Export problematic workflow JSON
3. Create issue at: https://github.com/sipeed/makoclaw/issues

---

## Invalid JSON in Step Config

### Symptom
Error when saving workflow: `invalid JSON: <error message>`

### Cause
Malformed JSON in tool arguments or step configuration.

### Solutions

**1. Validate JSON syntax**
```bash
# Use online validator: https://jsonlint.com
# Or command line:
echo '{"key": "value"}' | jq
```

**2. Common JSON mistakes**

```json
// ❌ Wrong: Trailing comma
{
  "arg1": "value",
  "arg2": "value",  ← Remove this comma
}

// ✅ Correct
{
  "arg1": "value",
  "arg2": "value"
}

// ❌ Wrong: Single quotes
{'key': 'value'}

// ✅ Correct: Double quotes
{"key": "value"}

// ❌ Wrong: Unescaped quotes in string
{"message": "Say "hello""}

// ✅ Correct: Escaped quotes
{"message": "Say \"hello\""}

// ❌ Wrong: Unquoted keys
{command: "ls"}

// ✅ Correct: Quoted keys
{"command": "ls"}
```

**3. Use JSON formatter**
Copy/paste your config into formatter before saving:
- VSCode: `Shift+Alt+F`
- Online: https://jsonformatter.org

---

## Multi-user Context Conflicts

### Symptom
Workflows from different users interfere with each other or see each other's data.

### Cause
**Known Limitation**: Current implementation uses shared agent context (`defaultAgentLoop`).

### Status
⚠️ **Not yet fixed.** See TODO comment in [main.go](../../cmd/makoclaw/main.go#L849)

### Workarounds

**1. Single-user deployment (recommended)**
- Use MakoClaw in single-user mode
- Each user runs their own instance

**2. Separate workspaces**
```json
// User 1 config
{
  "workspace": "/home/user1/.makoclaw/workspace"
}

// User 2 config
{
  "workspace": "/home/user2/.makoclaw/workspace"
}
```

**3. Use different MakoClaw instances**
```bash
# User 1: Port 18880
./makoclaw web --port 18880 --config /home/user1/.makoclaw/config.json

# User 2: Port 18881
./makoclaw web --port 18881 --config /home/user2/.makoclaw/config.json
```

### Tracking
This limitation is documented and planned for resolution in future updates.

---

## Workflow Import/Export Not Working

### Symptom
No UI button to export or import workflows.

### Cause
Feature not yet implemented in frontend.

### Workaround

**Export via API:**
```bash
# Get workflow JSON
curl http://localhost:18880/api/v1/workflows/1 > workflow-export.json
```

**Import via API:**
```bash
# Create workflow from JSON
curl -X POST http://localhost:18880/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d @workflow-export.json
```

**Template library:**
See [workflow-templates.json](../examples/workflow-templates.json) for pre-built workflows you can copy.

---

## LLM Returns Malformed Response

### Symptom
Prompt step succeeds but output is unusable (incomplete, wrong format, etc.)

### Cause
LLM didn't follow instructions in prompt.

### Solutions

**1. Make prompts more explicit**
```json
// ❌ Vague
{"message": "Analyze this code: {{step.1.output}}"}

// ✅ Explicit
{
  "message": "Analyze this code and respond with ONLY 'GOOD' or 'BAD':\n\n{{step.1.output}}"
}
```

**2. Request structured output**
```json
{
  "message": "Generate 3 search queries for: AI tools\n\nRespond in this exact format:\nquery1\nquery2\nquery3"
}
```

**3. Add examples (few-shot)**
```json
{
  "message": "Classify sentiment.\n\nExample:\nInput: 'Great product!'\nOutput: POSITIVE\n\nInput: 'Terrible service'\nOutput: NEGATIVE\n\nNow classify:\n{{step.1.output}}"
}
```

**4. Use condition to validate**
```json
[
  {"type": "prompt", "id": "step_1", "label": "Generate"},
  {
    "type": "condition",
    "id": "step_2",
    "label": "Validate format",
    "config": {
      "operator": "regex",
      "reference": "{{step.1.output}}",
      "value": "^(GOOD|BAD)$"
    }
  },
  {"type": "tool", "id": "step_3", "label": "Process result"}
]
```

---

## Web Panel Not Accessible

### Symptom
Cannot access http://localhost:18880

### Cause
- Service not running
- Wrong port
- Firewall blocking

### Solutions

**1. Verify service is running**
```bash
# Check health endpoint
curl http://localhost:18880/api/v1/health

# Expected: {"status":"ok"}
```

**2. Check port in config**
```json
{
  "web": {
    "listen_addr": ":18880"  // ← Verify this matches your URL
  }
}
```

**3. Check process**
```bash
# macOS/Linux
ps aux | grep makoclaw

# Check if gateway or web mode is running
```

**4. Check firewall**
```bash
# macOS: Allow in System Preferences → Security → Firewall
# Linux: Check iptables
sudo iptables -L | grep 18880
```

**5. Try different port**
```bash
# Start on different port
./makoclaw web --port 8080

# Access: http://localhost:8080
```

---

## For More Help

- **Documentation**: [Workflows Guide](../examples/workflows.md)
- **API Reference**: [Workflows API](../api-reference/workflows.md)
- **GitHub Issues**: https://github.com/sipeed/makoclaw/issues
- **Discord**: https://discord.gg/V4sAZ9XWpN

---

**Document Version**: 1.0.0  
**Last Updated**: February 2026
