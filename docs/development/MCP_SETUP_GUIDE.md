# MCP (Model Context Protocol) Setup Guide

This guide explains how to set up and use MCP servers to enhance MakoClaw's capabilities with external tools and services.

## What is MCP?

MCP (Model Context Protocol) is a standardized protocol that allows AI assistants to integrate with external tools, APIs, and services. MCP servers provide additional tools that MakoClaw can use to perform tasks beyond its built-in capabilities.

## Recommended MCP Servers for Development

### 1. Filesystem MCP Server

**Purpose:** Advanced file operations beyond basic read/write

**Repository:** [modelcontextprotocol/servers](https://github.com/modelcontextprotocol/servers)

**Setup:**
```bash
# Install the server
npm install -g @modelcontextprotocol/server-filesystem

# Add to MakoClaw config
# Edit ~/.MakoClaw/config.json
```

**Configuration:**
```json
{
  "mcp": {
    "servers": {
      "filesystem": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/paths"]
      }
    }
  }
}
```

**Tools Provided:**
- `read_file` - Read file contents
- `write_file` - Write to files
- `create_directory` - Create directories
- `list_directory` - List directory contents
- `move_file` - Move/rename files
- `search_files` - Search for files by name or content

---

### 2. GitHub MCP Server

**Purpose:** Interact with GitHub repositories

**Repository:** [modelcontextprotocol/servers](https://github.com/modelcontextprotocol/servers)

**Prerequisites:** GitHub Personal Access Token

**Setup:**
```bash
# Set GitHub token in environment
export GITHUB_PERSONAL_ACCESS_TOKEN="ghp_xxxxxxxxxxxx"

# Add to .env
echo "GITHUB_PERSONAL_ACCESS_TOKEN=ghp_xxxxxxxxxxxx" >> ~/.MakoClaw/.env
```

**Configuration:**
```json
{
  "mcp": {
    "servers": {
      "github": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-github"],
        "env": {
          "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_PERSONAL_ACCESS_TOKEN}"
        }
      }
    }
  }
}
```

**Tools Provided:**
- `create_or_update_file` - Create/update files in a repository
- `create_pull_request` - Create a pull request
- `fork_repository` - Fork a repository
- `push_files` - Push files to a repository
- `create_issue` - Create a GitHub issue
- `create_repository` - Create a new repository
- `get_file_contents` - Get file contents from a repository
- `search_repositories` - Search for repositories
- `search_issues_and_prs` - Search issues and pull requests

**Use Cases:**
- Automate repository management
- Create pull requests from code changes
- Search and analyze repositories
- Automate issue creation

---

### 3. Brave Search MCP Server

**Purpose:** Web search capabilities using Brave Search API

**Repository:** [modelcontextprotocol/servers](https://github.com/modelcontextprotocol/servers)

**Prerequisites:** Brave Search API key

**Setup:**
```bash
# Get API key from https://api.search.brave.com/app/keys
export BRAVE_API_KEY="your_brave_api_key"

# Add to .env
echo "BRAVE_API_KEY=your_brave_api_key" >> ~/.MakoClaw/.env
```

**Configuration:**
```json
{
  "mcp": {
    "servers": {
      "brave-search": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-brave-search"],
        "env": {
          "BRAVE_API_KEY": "${BRAVE_API_KEY}"
        }
      }
    }
  }
}
```

**Tools Provided:**
- `brave_web_search` - Search the web using Brave Search

**Use Cases:**
- Research current information
- Find documentation and examples
- Look up best practices
- Verify facts and figures

---

### 4. Postgres MCP Server

**Purpose:** Interact with PostgreSQL databases

**Repository:** [modelcontextprotocol/servers](https://github.com/modelcontextprotocol/servers)

**Prerequisites:** PostgreSQL database connection details

**Configuration:**
```json
{
  "mcp": {
    "servers": {
      "postgres": {
        "command": "npx",
        "args": [
          "-y",
          "@modelcontextprotocol/server-postgres",
          "postgresql://user:password@localhost:5432/dbname"
        ]
      }
    }
  }
}
```

**Tools Provided:**
- `execute_query` - Execute SQL queries
- `list_tables` - List all tables
- `describe_table` - Get table schema
- `analyze_query` - Analyze query performance

**Use Cases:**
- Query databases for information
- Analyze data
- Generate reports
- Optimize queries

---

### 5. Puppeteer MCP Server

**Purpose:** Browser automation and web scraping

**Repository:** [modelcontextprotocol/servers](https://github.com/modelcontextprotocol/servers)

**Setup:**
```bash
npm install -g @modelcontextprotocol/server-puppeteer
```

**Configuration:**
```json
{
  "mcp": {
    "servers": {
      "puppeteer": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-puppeteer"]
      }
    }
  }
}
```

**Tools Provided:**
- `puppeteer_navigate` - Navigate to a URL
- `puppeteer_screenshot` - Take a screenshot
- `puppeteer_click` - Click an element
- `puppeteer_fill` - Fill a form field
- `puppeteer_select` - Select from a dropdown
- `puppeteer_hover` - Hover over an element
- `puppeteer_evaluate` - Execute JavaScript in the browser

**Use Cases:**
- Automated testing
- Web scraping
- Screenshot generation
- Form automation

---

### 6. Google Drive MCP Server

**Purpose:** Interact with Google Drive files

**Prerequisites:** Google Cloud project with Drive API enabled

**Configuration:**
```json
{
  "mcp": {
    "servers": {
      "google-drive": {
        "command": "npx",
        "args": [
          "-y",
          "@modelcontextprotocol/server-gdrive"
        ],
        "env": {
          "GOOGLE_CREDENTIALS": "/path/to/service-account-key.json"
        }
      }
    }
  }
}
```

**Tools Provided:**
- `create_file` - Create a new file
- `read_file` - Read file contents
- `update_file` - Update an existing file
- `list_files` - List files in a folder
- `search_files` - Search for files
- `share_file` - Share a file

**Use Cases:**
- Document management
- Backup files to Drive
- Collaborative editing
- File organization

---

## Complete MCP Configuration Example

Here's a comprehensive MCP configuration for development:

```json
{
  "mcp": {
    "servers": {
      "filesystem": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/projects"],
        "enabled": true
      },
      "github": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-github"],
        "env": {
          "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_PERSONAL_ACCESS_TOKEN}"
        },
        "enabled": true
      },
      "brave-search": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-brave-search"],
        "env": {
          "BRAVE_API_KEY": "${BRAVE_API_KEY}"
        },
        "enabled": true
      },
      "postgres": {
        "command": "npx",
        "args": [
          "-y",
          "@modelcontextprotocol/server-postgres",
          "postgresql://makoclaw:makoclaw@localhost:5432/makoclaw_dev"
        ],
        "enabled": true
      },
      "puppeteer": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-puppeteer"],
        "enabled": false
      },
      "google-drive": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-gdrive"],
        "env": {
          "GOOGLE_CREDENTIALS": "/path/to/credentials.json"
        },
        "enabled": false
      }
    }
  }
}
```

## Testing MCP Servers

After configuring an MCP server, test it:

```bash
# Start MakoClaw with MCP logging
makoclaw agent --verbose

# Or test specific MCP server
makocraw doctor --mcp
```

Check available MCP tools:

```bash
# List all MCP tools
makoclaw tools --mcp

# Get details about a specific MCP tool
makoclaw tools --mcp --name=github_create_pull_request
```

## Troubleshooting

### MCP Server Not Starting

**Check the logs:**
```bash
tail -f ~/.MakoClaw/workspace/logs/mcp.log
```

**Common Issues:**
1. **Node.js not installed**
   ```bash
   # Install Node.js
   curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
   sudo apt-get install -y nodejs
   ```

2. **Missing environment variables**
   - Verify variables in `~/.MakoClaw/.env`
   - Restart MakoClaw after changing `.env`

3. **Incorrect permissions**
   - Check file/directory permissions
   - Ensure MCP server can access required paths

### MCP Server Connection Issues

**Test MCP server directly:**
```bash
# Test filesystem server
npx -y @modelcontextprotocol/server-filesystem /path/to/dir

# Test GitHub server
npx -y @modelcontextprotocol/server-github
```

## Creating Custom MCP Servers

You can create custom MCP servers for your specific needs:

**Template:**
```javascript
// my-mcp-server/index.js
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';

const server = new Server({
  name: 'my-custom-server',
  version: '1.0.0',
}, {
  capabilities: {
    tools: {},
  },
});

// Register a tool
server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: [{
    name: 'my_custom_tool',
    description: 'Does something custom',
    inputSchema: {
      type: 'object',
      properties: {
        param1: {
          type: 'string',
          description: 'First parameter'
        }
      },
      required: ['param1']
    }
  }]
}));

// Handle tool execution
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  if (request.params.name === 'my_custom_tool') {
    const param1 = request.params.arguments?.param1;
    // Do something with param1
    return {
      content: [{
        type: 'text',
        text: `Result: ${param1}`
      }]
    };
  }
});

const transport = new StdioServerTransport();
await server.connect(transport);
```

**Configuration:**
```json
{
  "mcp": {
    "servers": {
      "my-custom": {
        "command": "node",
        "args": ["/path/to/my-mcp-server/index.js"]
      }
    }
  }
}
```

## Best Practices

### Security

1. **Limit access scopes:**
   - Use minimal required permissions
   - Create separate API keys for MCP
   - Rotate credentials regularly

2. **Validate inputs:**
   - MCP servers should validate all inputs
   - Sanitize file paths to prevent directory traversal
   - Limit resource usage

3. **Sandbox:**
   - Run MCP servers in separate processes
   - Use containerization for isolation
   - Monitor for unusual activity

### Performance

1. **Enable/disable servers as needed:**
   - Disable unused MCP servers
   - Use lightweight servers for development
   - Cache results when appropriate

2. **Monitor usage:**
   - Track MCP server call counts
   - Monitor response times
   - Set timeouts for long-running operations

### Reliability

1. **Error handling:**
   - Implement proper error handling in MCP servers
   - Provide meaningful error messages
   - Graceful degradation when servers fail

2. **Health checks:**
   - Implement health check endpoints
   - Monitor server status
   - Automatic restart on failure

## Resources

- [MCP Specification](https://modelcontextprotocol.io/)
- [Official MCP Servers](https://github.com/modelcontextprotocol/servers)
- [MCP SDK](https://github.com/modelcontextprotocol/sdk)
- [Community MCP Servers](https://github.com/topics/mcp-server)

## Quick Start Checklist

- [ ] Install Node.js and npm
- [ ] Create GitHub personal access token
- [ ] Get Brave Search API key
- [ ] Update `~/.MakoClaw/.env` with credentials
- [ ] Configure MCP servers in `~/.MakoClaw/config.json`
- [ ] Restart MakoClaw
- [ ] Test MCP servers with `makoclaw doctor --mcp`
- [ ] Verify tools are available with `makoclaw tools --mcp`
