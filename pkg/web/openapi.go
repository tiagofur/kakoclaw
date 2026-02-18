package web

import (
	"net/http"
)

// handleAPIDocsUI serves the Swagger UI page that loads the OpenAPI spec.
func (s *Server) handleAPIDocsUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(swaggerUIHTML))
}

// handleOpenAPISpec serves the raw OpenAPI 3.0 JSON spec.
func (s *Server) handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(openAPISpec))
}

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>PicoClaw API Documentation</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    .swagger-ui .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: '/api/v1/openapi.json',
      dom_id: '#swagger-ui',
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIBundle.SwaggerUIStandalonePreset
      ],
      layout: 'BaseLayout',
      deepLinking: true,
      defaultModelsExpandDepth: 1,
      defaultModelExpandDepth: 1,
      docExpansion: 'list',
      persistAuthorization: true
    });
  </script>
</body>
</html>`

const openAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "PicoClaw API",
    "description": "REST API for the PicoClaw AI agent platform. Includes chat, tasks, skills, cron jobs, channels, voice, memory, files, and export.",
    "version": "1.0.0",
    "license": { "name": "MIT" }
  },
  "servers": [
    { "url": "/", "description": "Current server" }
  ],
  "components": {
    "securitySchemes": {
      "bearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    },
    "schemas": {
      "Error": {
        "type": "object",
        "properties": {
          "error": { "type": "string" }
        }
      },
      "Task": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "title": { "type": "string" },
          "description": { "type": "string" },
          "status": { "type": "string", "enum": ["backlog", "todo", "in_progress", "review", "done"] },
          "result": { "type": "string" },
          "created_at": { "type": "string", "format": "date-time" },
          "updated_at": { "type": "string", "format": "date-time" },
          "archived": { "type": "boolean" }
        }
      },
      "ChatMessage": {
        "type": "object",
        "properties": {
          "role": { "type": "string", "enum": ["user", "assistant", "system"] },
          "content": { "type": "string" },
          "timestamp": { "type": "string", "format": "date-time" }
        }
      },
      "ChatSession": {
        "type": "object",
        "properties": {
          "session_id": { "type": "string" },
          "last_message": { "type": "string" },
          "message_count": { "type": "integer" },
          "updated_at": { "type": "string", "format": "date-time" }
        }
      },
      "CronJob": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "schedule": { "type": "object" },
          "message": { "type": "string" },
          "deliver": { "type": "boolean" },
          "channel": { "type": "string" },
          "to": { "type": "string" },
          "enabled": { "type": "boolean" },
          "next_run": { "type": "string", "format": "date-time" }
        }
      },
      "Skill": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "content": { "type": "string" },
          "repository": { "type": "string" }
        }
      },
      "FileEntry": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "path": { "type": "string" },
          "is_dir": { "type": "boolean" },
          "size": { "type": "integer" },
          "mod_time": { "type": "string", "format": "date-time" }
        }
      },
      "Provider": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "enabled": { "type": "boolean" },
          "is_active": { "type": "boolean" },
          "models": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "id": { "type": "string" },
                "provider": { "type": "string" }
              }
            }
          }
        }
      }
    }
  },
  "security": [{ "bearerAuth": [] }],
  "paths": {
    "/api/v1/health": {
      "get": {
        "tags": ["Health"],
        "summary": "Health check",
        "security": [],
        "responses": {
          "200": {
            "description": "Server is healthy",
            "content": { "application/json": { "schema": { "type": "object", "properties": { "status": { "type": "string", "example": "ok" } } } } }
          }
        }
      }
    },
    "/api/v1/auth/login": {
      "post": {
        "tags": ["Authentication"],
        "summary": "Login with username and password",
        "description": "Rate-limited to 5 attempts per minute per IP.",
        "security": [],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["username", "password"],
                "properties": {
                  "username": { "type": "string" },
                  "password": { "type": "string" }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "JWT token",
            "content": { "application/json": { "schema": { "type": "object", "properties": { "token": { "type": "string" } } } } }
          },
          "401": { "description": "Invalid credentials" },
          "429": { "description": "Rate limit exceeded" }
        }
      }
    },
    "/api/v1/auth/change-password": {
      "post": {
        "tags": ["Authentication"],
        "summary": "Change the current user's password",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["old_password", "new_password"],
                "properties": {
                  "old_password": { "type": "string" },
                  "new_password": { "type": "string" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Password changed", "content": { "application/json": { "schema": { "type": "object", "properties": { "status": { "type": "string", "example": "ok" } } } } } },
          "400": { "description": "Invalid request" },
          "401": { "description": "Wrong old password" }
        }
      }
    },
    "/api/v1/auth/me": {
      "get": {
        "tags": ["Authentication"],
        "summary": "Get current authenticated user",
        "responses": {
          "200": { "description": "Current user", "content": { "application/json": { "schema": { "type": "object", "properties": { "username": { "type": "string" } } } } } }
        }
      }
    },
    "/api/v1/tasks": {
      "get": {
        "tags": ["Tasks"],
        "summary": "List all tasks",
        "parameters": [
          { "name": "include_archived", "in": "query", "schema": { "type": "boolean", "default": false }, "description": "Include archived tasks" }
        ],
        "responses": {
          "200": { "description": "Task list", "content": { "application/json": { "schema": { "type": "object", "properties": { "tasks": { "type": "array", "items": { "$ref": "#/components/schemas/Task" } } } } } } }
        }
      },
      "post": {
        "tags": ["Tasks"],
        "summary": "Create a new task",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["title"],
                "properties": {
                  "title": { "type": "string" },
                  "description": { "type": "string" },
                  "status": { "type": "string", "enum": ["backlog", "todo", "in_progress", "review", "done"], "default": "backlog" }
                }
              }
            }
          }
        },
        "responses": {
          "201": { "description": "Task created", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Task" } } } },
          "400": { "description": "Invalid request" }
        }
      }
    },
    "/api/v1/tasks/{id}": {
      "put": {
        "tags": ["Tasks"],
        "summary": "Update a task",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "title": { "type": "string" },
                  "description": { "type": "string" },
                  "status": { "type": "string" },
                  "result": { "type": "string" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Task updated", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Task" } } } },
          "404": { "description": "Task not found" }
        }
      },
      "delete": {
        "tags": ["Tasks"],
        "summary": "Delete (archive) a task",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Task deleted", "content": { "application/json": { "schema": { "type": "object", "properties": { "status": { "type": "string", "example": "ok" } } } } } },
          "404": { "description": "Task not found" }
        }
      }
    },
    "/api/v1/tasks/{id}/status": {
      "patch": {
        "tags": ["Tasks"],
        "summary": "Change task status",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["status"],
                "properties": {
                  "status": { "type": "string", "enum": ["backlog", "todo", "in_progress", "review", "done"] }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Status changed", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Task" } } } },
          "404": { "description": "Task not found" }
        }
      }
    },
    "/api/v1/tasks/{id}/archive": {
      "post": {
        "tags": ["Tasks"],
        "summary": "Archive a task",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Task archived" },
          "404": { "description": "Task not found" }
        }
      }
    },
    "/api/v1/tasks/{id}/unarchive": {
      "post": {
        "tags": ["Tasks"],
        "summary": "Unarchive a task",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Task unarchived" },
          "404": { "description": "Task not found" }
        }
      }
    },
    "/api/v1/tasks/{id}/logs": {
      "get": {
        "tags": ["Tasks"],
        "summary": "Get execution logs for a task",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Task logs", "content": { "application/json": { "schema": { "type": "object", "properties": { "logs": { "type": "array", "items": { "type": "object" } } } } } } }
        }
      }
    },
    "/api/v1/tasks/search": {
      "get": {
        "tags": ["Tasks"],
        "summary": "Search tasks by keyword",
        "parameters": [{ "name": "q", "in": "query", "required": true, "schema": { "type": "string" }, "description": "Search query" }],
        "responses": {
          "200": { "description": "Matching tasks", "content": { "application/json": { "schema": { "type": "object", "properties": { "tasks": { "type": "array", "items": { "$ref": "#/components/schemas/Task" } } } } } } },
          "400": { "description": "Missing query parameter" }
        }
      }
    },
    "/api/v1/chat/sessions": {
      "get": {
        "tags": ["Chat"],
        "summary": "List all chat sessions",
        "responses": {
          "200": { "description": "Session list", "content": { "application/json": { "schema": { "type": "object", "properties": { "sessions": { "type": "array", "items": { "$ref": "#/components/schemas/ChatSession" } } } } } } }
        }
      }
    },
    "/api/v1/chat/sessions/{id}": {
      "get": {
        "tags": ["Chat"],
        "summary": "Get messages for a chat session",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Session messages", "content": { "application/json": { "schema": { "type": "object", "properties": { "messages": { "type": "array", "items": { "$ref": "#/components/schemas/ChatMessage" } } } } } } },
          "404": { "description": "Session not found" }
        }
      }
    },
    "/api/v1/chat/search": {
      "get": {
        "tags": ["Chat"],
        "summary": "Search chat messages",
        "parameters": [{ "name": "q", "in": "query", "required": true, "schema": { "type": "string" }, "description": "Search query" }],
        "responses": {
          "200": { "description": "Matching messages", "content": { "application/json": { "schema": { "type": "object", "properties": { "messages": { "type": "array", "items": { "$ref": "#/components/schemas/ChatMessage" } } } } } } },
          "400": { "description": "Missing query parameter" }
        }
      }
    },
    "/api/v1/memory/longterm": {
      "get": {
        "tags": ["Memory"],
        "summary": "Read long-term memory",
        "responses": {
          "200": { "description": "Memory content", "content": { "application/json": { "schema": { "type": "object", "properties": { "content": { "type": "string" } } } } } }
        }
      },
      "post": {
        "tags": ["Memory"],
        "summary": "Write long-term memory",
        "requestBody": {
          "required": true,
          "content": { "application/json": { "schema": { "type": "object", "required": ["content"], "properties": { "content": { "type": "string" } } } } }
        },
        "responses": {
          "200": { "description": "Memory saved", "content": { "application/json": { "schema": { "type": "object", "properties": { "status": { "type": "string", "example": "ok" } } } } } }
        }
      }
    },
    "/api/v1/memory/daily": {
      "get": {
        "tags": ["Memory"],
        "summary": "Read recent daily notes",
        "parameters": [{ "name": "days", "in": "query", "schema": { "type": "integer", "default": 7 }, "description": "Number of days to include" }],
        "responses": {
          "200": { "description": "Daily notes", "content": { "application/json": { "schema": { "type": "object", "properties": { "content": { "type": "string" } } } } } }
        }
      }
    },
    "/api/v1/skills": {
      "get": {
        "tags": ["Skills"],
        "summary": "List installed skills or marketplace",
        "parameters": [{ "name": "type", "in": "query", "schema": { "type": "string", "enum": ["installed", "available"] }, "description": "Use 'available' to list marketplace skills" }],
        "responses": {
          "200": { "description": "Skills list", "content": { "application/json": { "schema": { "type": "object", "properties": { "skills": { "type": "array", "items": { "$ref": "#/components/schemas/Skill" } } } } } } }
        }
      }
    },
    "/api/v1/skills/{name}": {
      "get": {
        "tags": ["Skills"],
        "summary": "View a skill's content",
        "parameters": [{ "name": "name", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Skill content", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Skill" } } } },
          "404": { "description": "Skill not found" }
        }
      },
      "delete": {
        "tags": ["Skills"],
        "summary": "Uninstall a skill",
        "parameters": [{ "name": "name", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Skill uninstalled", "content": { "application/json": { "schema": { "type": "object", "properties": { "status": { "type": "string" }, "name": { "type": "string" } } } } } },
          "404": { "description": "Skill not found" }
        }
      }
    },
    "/api/v1/skills/install": {
      "post": {
        "tags": ["Skills"],
        "summary": "Install a skill from GitHub",
        "requestBody": {
          "required": true,
          "content": { "application/json": { "schema": { "type": "object", "required": ["repository"], "properties": { "repository": { "type": "string", "example": "owner/repo" } } } } }
        },
        "responses": {
          "200": { "description": "Skill installed", "content": { "application/json": { "schema": { "type": "object", "properties": { "status": { "type": "string" }, "repository": { "type": "string" } } } } } },
          "400": { "description": "Invalid request" }
        }
      }
    },
    "/api/v1/cron": {
      "get": {
        "tags": ["Cron"],
        "summary": "List cron jobs",
        "parameters": [{ "name": "include_disabled", "in": "query", "schema": { "type": "boolean", "default": false }, "description": "Include disabled jobs" }],
        "responses": {
          "200": { "description": "Cron jobs", "content": { "application/json": { "schema": { "type": "object", "properties": { "jobs": { "type": "array", "items": { "$ref": "#/components/schemas/CronJob" } }, "status": { "type": "object" } } } } } }
        }
      },
      "post": {
        "tags": ["Cron"],
        "summary": "Create a cron job",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["name", "schedule", "message"],
                "properties": {
                  "name": { "type": "string" },
                  "schedule": {
                    "type": "object",
                    "properties": {
                      "kind": { "type": "string", "enum": ["at", "every", "cron"] },
                      "atMs": { "type": "integer" },
                      "everyMs": { "type": "integer" },
                      "expr": { "type": "string" },
                      "tz": { "type": "string" }
                    }
                  },
                  "message": { "type": "string" },
                  "deliver": { "type": "boolean" },
                  "channel": { "type": "string" },
                  "to": { "type": "string" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Cron job created", "content": { "application/json": { "schema": { "type": "object", "properties": { "job": { "$ref": "#/components/schemas/CronJob" } } } } } },
          "400": { "description": "Invalid request" }
        }
      }
    },
    "/api/v1/cron/{id}": {
      "delete": {
        "tags": ["Cron"],
        "summary": "Remove a cron job",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "responses": {
          "200": { "description": "Cron job removed" },
          "404": { "description": "Job not found" }
        }
      },
      "patch": {
        "tags": ["Cron"],
        "summary": "Enable or disable a cron job",
        "parameters": [{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }],
        "requestBody": {
          "required": true,
          "content": { "application/json": { "schema": { "type": "object", "required": ["enabled"], "properties": { "enabled": { "type": "boolean" } } } } }
        },
        "responses": {
          "200": { "description": "Job updated", "content": { "application/json": { "schema": { "type": "object", "properties": { "job": { "$ref": "#/components/schemas/CronJob" } } } } } },
          "404": { "description": "Job not found" }
        }
      }
    },
    "/api/v1/channels": {
      "get": {
        "tags": ["Channels"],
        "summary": "List messaging channels and their status",
        "responses": {
          "200": { "description": "Channel status", "content": { "application/json": { "schema": { "type": "object", "properties": { "channels": { "type": "object" }, "enabled": { "type": "array", "items": { "type": "string" } } } } } } }
        }
      }
    },
    "/api/v1/config": {
      "get": {
        "tags": ["Config"],
        "summary": "Get redacted server configuration",
        "description": "API keys are masked for security. Read-only.",
        "responses": {
          "200": { "description": "Server configuration", "content": { "application/json": { "schema": { "type": "object", "properties": { "config": { "type": "object" } } } } } }
        }
      }
    },
    "/api/v1/models": {
      "get": {
        "tags": ["Models"],
        "summary": "List available AI providers and models",
        "responses": {
          "200": {
            "description": "Provider and model list",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "current_model": { "type": "string" },
                    "current_provider": { "type": "string" },
                    "providers": { "type": "array", "items": { "$ref": "#/components/schemas/Provider" } }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/v1/files": {
      "get": {
        "tags": ["Files"],
        "summary": "Browse workspace root directory",
        "responses": {
          "200": { "description": "Directory listing", "content": { "application/json": { "schema": { "type": "object", "properties": { "path": { "type": "string" }, "entries": { "type": "array", "items": { "$ref": "#/components/schemas/FileEntry" } } } } } } }
        }
      }
    },
    "/api/v1/files/{path}": {
      "get": {
        "tags": ["Files"],
        "summary": "Browse directory or read file content",
        "description": "Returns directory listing for directories, or file content (up to 1 MB) for files. Path traversal outside workspace is blocked.",
        "parameters": [{ "name": "path", "in": "path", "required": true, "schema": { "type": "string" }, "description": "Relative path within workspace" }],
        "responses": {
          "200": { "description": "File or directory content" },
          "400": { "description": "Path traversal attempt blocked" },
          "404": { "description": "Path not found" }
        }
      }
    },
    "/api/v1/export/tasks": {
      "get": {
        "tags": ["Export"],
        "summary": "Export all tasks",
        "parameters": [{ "name": "format", "in": "query", "schema": { "type": "string", "enum": ["json", "csv"], "default": "json" }, "description": "Export format" }],
        "responses": {
          "200": { "description": "Exported tasks (JSON or CSV download)" }
        }
      }
    },
    "/api/v1/export/chat": {
      "get": {
        "tags": ["Export"],
        "summary": "Export chat history",
        "parameters": [{ "name": "session_id", "in": "query", "schema": { "type": "string" }, "description": "Export a specific session; omit for all sessions summary" }],
        "responses": {
          "200": { "description": "Exported chat data" }
        }
      }
    },
    "/api/v1/voice/transcribe": {
      "post": {
        "tags": ["Voice"],
        "summary": "Transcribe audio to text (Groq Whisper)",
        "description": "Accepts audio files up to 25 MB via multipart form upload. Uses Groq Whisper large-v3 model.",
        "requestBody": {
          "required": true,
          "content": {
            "multipart/form-data": {
              "schema": {
                "type": "object",
                "required": ["audio"],
                "properties": {
                  "audio": { "type": "string", "format": "binary", "description": "Audio file (.webm, .wav, .mp3, etc.)" }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Transcription result",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "text": { "type": "string" },
                    "language": { "type": "string" },
                    "duration": { "type": "number" }
                  }
                }
              }
            }
          },
          "400": { "description": "No audio file or invalid upload" },
          "500": { "description": "Transcription failed" },
          "503": { "description": "Transcriber not configured" }
        }
      }
    }
  }
}`
