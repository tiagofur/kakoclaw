# KakoClaw/KakoClaw — Roadmap de Mejoras

> Documento generado a partir del análisis competitivo contra Open WebUI, Dify, LobeChat, Flowise, n8n, Langflow, BotPress, AutoGPT y otros.

## Estado: Fase UX Completada — Fase 7 Pendiente

---

## Fase 5 — Mejoras de Alta Prioridad (COMPLETADA ✅)

### 5.1 API de Modelos Disponibles
**Estado**: COMPLETADO
**Archivos**: `pkg/web/server.go`, `pkg/providers/`

- `GET /api/v1/models` — Endpoint que retorna los proveedores configurados y sus modelos
- Lista los proveedores con API key configurada
- Para Ollama, consulta modelos locales via `/api/tags`
- Para otros, retorna modelos conocidos por proveedor
- Retorna modelo actual (`agents.defaults.model`)

### 5.2 Selector de Modelo por Conversación
**Estado**: COMPLETADO
**Archivos**: `pkg/web/frontend/src/views/ChatView.vue`, `pkg/web/server.go`, `pkg/agent/loop.go`, `chatStore.js`, `advancedService.js`

- Dropdown en el top bar del chat para seleccionar modelo/proveedor
- El WebSocket acepta campo `model` en el mensaje
- El backend usa el modelo indicado para esa petición via `ProcessDirectWithModel()`
- `processOptions.ModelOverride` propagado hasta `runLLMIteration()`
- `chatStore` mantiene `selectedModel`, `currentModel`, `allModels` computed

### 5.3 Streaming de Respuestas Token-by-Token
**Estado**: COMPLETADO
**Archivos**: `pkg/providers/types.go`, `pkg/providers/http_provider.go`, `pkg/providers/ollama_provider.go`, `pkg/agent/loop.go`, `pkg/web/server.go`, `pkg/web/frontend/src/stores/chatStore.js`, `pkg/web/frontend/src/views/ChatView.vue`

- Interfaz `StreamingLLMProvider` con método `ChatStream()` retornando `<-chan StreamChunk`
- `HTTPProvider.ChatStream()` — streaming SSE (OpenAI-compatible)
- `OllamaProvider.ChatStream()` — streaming NDJSON (Ollama native)
- `ProcessDirectWithModelStream()` y `runAgentLoopStream()` en AgentLoop
- Protocolo WebSocket: `stream_start` → N × `stream` (delta) → `stream_end` (texto completo) → `ready`
- Frontend: renderizado progresivo con cursor parpadeante, switch a MarkdownRenderer al finalizar
- Fallback automático a respuesta completa si el proveedor no implementa `StreamingLLMProvider`
- `ClaudeProvider` y `CodexProvider` usan fallback (SDKs de vendor sin streaming propio)

### 5.4 Botón Regenerar Respuesta
**Estado**: COMPLETADO
**Archivos**: `ChatView.vue`

- Botón "Regenerar" (refresh icon) visible al hover sobre el último mensaje del asistente
- Reenvía el último mensaje del usuario al LLM
- Usa el modelo seleccionado actualmente
- Reemplaza la última respuesta del asistente

### 5.5 Toggle Dark/Light Theme
**Estado**: COMPLETADO
**Archivos**: `tailwind.config.js`, `globals.css`, `index.html`, `Sidebar.vue` (ya existía), `uiStore.js` (ya existía)

- CSS custom properties (RGB triplets) para compatibilidad con Tailwind opacity
- Tema claro con paleta Slate/White definida en `:root`
- Tema oscuro en `.dark` (Slate 900/800/700)
- Toggle sol/luna en el sidebar (ya implementado en sesión anterior)
- Persistencia en localStorage via `uiStore`
- Script inline en `index.html` previene flash de tema incorrecto

---

## Fase 6 — Mejoras de Prioridad Media (COMPLETADA ✅)

### 6.1 Voice Input/Output en Panel Web (COMPLETADO ✅)
**Estado**: COMPLETADO
**Archivos**: `pkg/voice/transcriber.go`, `pkg/web/server.go`, `pkg/web/frontend/src/views/ChatView.vue`, `pkg/web/frontend/src/services/advancedService.js`

- `POST /api/v1/voice/transcribe` — Endpoint multipart para STT via Groq Whisper
- `SetTranscriber()` setter en Server para inyectar transcriber
- Botón de micrófono en ChatView con hold-to-record UX (MediaRecorder API)
- `transcribeAudio()` en advancedService para enviar audio al backend

### 6.2 MCP (Model Context Protocol) Client (COMPLETADO ✅)
**Estado**: COMPLETADO
**Archivos**: `pkg/mcp/client.go`, `pkg/mcp/manager.go`, `pkg/mcp/tool.go`, `pkg/config/config.go`, `pkg/agent/loop.go`, `pkg/web/server.go`, `pkg/web/handlers_advanced.go`, `cmd/KakoClaw/main.go`, `pkg/web/frontend/src/views/MCPView.vue`, `pkg/web/frontend/src/services/advancedService.js`

- MCP client implementado con JSON-RPC 2.0 sobre STDIO (protocolo version 2024-11-05)
- `Client`: inicia proceso, handshake initialize/initialized, tools/list discovery, tools/call execution
- `Manager`: gestiona múltiples MCP servers, Start/Stop/Reconnect, GetTools()
- `MCPTool`: wrapper que implementa `tools.Tool` interface, prefijo `mcp_<server>_<tool>` para uniqueness
- Config: `config.json > tools.mcp.servers` con enabled, command, args, env por servidor
- Registro automático de MCP tools en AgentLoop al arranque
- REST: `GET /api/v1/mcp` (lista servidores + status), `POST /api/v1/mcp/{name}/reconnect`
- Re-registro de tools en AgentLoop tras reconnect desde la UI
- Vista `MCPView.vue` con cards por servidor (status, tools, reconnect), config example para nuevos usuarios
- Wiring en `gatewayCmd()` y `webCmd()` con startup logging y graceful shutdown

### 6.3 RAG / Knowledge Base (COMPLETADO ✅)
**Estado**: COMPLETADO
**Archivos**: `pkg/storage/knowledge.go`, `pkg/tools/knowledge.go`, `pkg/web/handlers_advanced.go`, `pkg/web/server.go`, `pkg/web/frontend/src/views/KnowledgeView.vue`, `pkg/web/frontend/src/services/advancedService.js`

- SQLite FTS5 para búsqueda full-text con BM25 ranking
- Tablas: `knowledge_documents`, `knowledge_chunks`, `knowledge_fts`
- Endpoints REST: `GET/POST /api/v1/knowledge`, `DELETE /api/v1/knowledge/{id}`, `GET /api/v1/knowledge/search?q=`
- Chunking por párrafos/oraciones con overlap configurable (1000 chars, 200 overlap)
- Tool `query_knowledge` registrado en AgentLoop para RAG automático
- Vista `KnowledgeView.vue` con drag-and-drop upload, lista de documentos, búsqueda interactiva
- Sidebar link con icono de libro en sección Tools

### 6.4 Búsqueda Web Configurable en UI (COMPLETADO ✅)
**Estado**: COMPLETADO
**Archivos**: `pkg/agent/loop.go`, `pkg/web/server.go`, `pkg/web/frontend/src/stores/chatStore.js`, `pkg/web/frontend/src/views/ChatView.vue`

- Toggle en ChatView (icono lupa) para activar/desactivar búsqueda web por conversación
- `ExcludeTools` en `processOptions` para filtrar tools por request
- Variadic `excludeTools` propagado por toda la cadena: `ProcessDirectWithModel()` → `processMessageWithModel()` → `runLLMIteration()`
- WebSocket acepta campo `web_search` (bool) del cliente
- Filtrado de `web_search` tool en ambos flujos (streaming y non-streaming)

### 6.5 API Docs / Swagger (COMPLETADO ✅)
**Estado**: COMPLETADO
**Archivos**: `pkg/web/openapi.go`, `pkg/web/server.go`

- OpenAPI 3.0.3 spec completo con 37 endpoints documentados
- Swagger UI embebido en `/api/docs` (cargado desde unpkg.com CDN)
- Spec raw en `/api/v1/openapi.json`
- Auth bypass + CSP relaxation para endpoints de documentación

---

## Fase UX — Refinamiento Completo de la Experiencia de Usuario (COMPLETADA ✅)

> Pase completo de refinamiento UX dividido en 3 sub-fases: Sessions, Cron mejorado, y Quick Bugs.

### UX-A: Gestión de Sesiones de Chat (COMPLETADO ✅)
**Archivos backend**: `pkg/storage/sqlite.go`, `pkg/storage/chat.go`, `pkg/web/server.go`, `pkg/web/handlers_advanced.go`
**Archivos frontend**: `ChatView.vue`, `HistoryView.vue`, `taskService.js`

- **Nueva tabla `sessions`** con `id`, `session_id` (UNIQUE), `title`, `archived`, `created_at`, `updated_at`
- Auto-migración `migrateSessions()` — backfill desde tabla `chats` existente
- `ensureSession()` — crea registro de sesión automáticamente al guardar/importar mensajes
- CRUD completo: `GetSession`, `UpdateSession`, `DeleteSession` (cascade delete de mensajes)
- `ListSessions(archived *bool, limit, offset int)` — filtrado por archivado + paginación
- `ForkSession` ahora también llama `ensureSession` para la sesión nueva
- **API REST**:
  - `GET /api/v1/chat/sessions?archived=true&limit=50&offset=0`
  - `DELETE /api/v1/chat/sessions/{id}` — elimina sesión y todos sus mensajes
  - `PATCH /api/v1/chat/sessions/{id}` — body: `{title, archived}`
- **Frontend ChatView**: Menú contextual (Renombrar/Archivar/Eliminar) en sidebar, renombrado inline
- **Frontend HistoryView**: Botones de acción por sesión, checkbox "Archived" para filtrar
- **Botón copiar** en mensajes del asistente (ChatView + HistoryView)

### UX-B: Cron Jobs Mejorado (COMPLETADO ✅)
**Archivos backend**: `pkg/cron/service.go`, `pkg/web/handlers_advanced.go`
**Archivos frontend**: `CronView.vue`, `advancedService.js`

- **`ValidateSchedule()`** — valida los 3 tipos de schedule; usa `gronx.IsValid()` para cron expressions
- **`UpdateJob()`** — update completo de nombre, schedule, payload; recomputa next run; valida schedule
- **`PUT /api/v1/cron/{id}`** — nuevo handler para edición completa de jobs
- **Validación en `POST /api/v1/cron`** — retorna 400 (no 500) en schedule inválido
- **Fix TZ en `computeNextRun()`** — aplica `time.LoadLocation(schedule.TZ)` y convierte a UTC
- **CronView.vue reescrito completamente**:
  - 6 tipos de schedule: Daily, Weekly, Monthly, Interval, One-time, Custom (cron expression)
  - Auto-generación de expresiones cron desde la UI visual
  - Preview de próximas 3 ejecuciones
  - Modal de edición con reverse-parsing de cron expressions
  - Diálogo de confirmación para eliminar
  - Selector de timezone
  - Display amigable de schedules (ej. "Every Monday, Wednesday at 09:00")
- **`advancedService.updateCronJob(id, data)`** — nuevo método frontend

### UX-C: Quick Bugs y Polish (COMPLETADO ✅)
**Archivos**: `TaskDetailsModal.vue`, `pkg/storage/task.go`, `pkg/web/server.go`, `DashboardView.vue`, `SettingsView.vue`, `FilesView.vue`, `ChannelsView.vue`, `TasksView.vue`, `MetricsView.vue`

- **C1**: Fixed `log.action`→`log.event`, `log.details`→`log.message` en `TaskDetailsModal.vue`
- **C2**: `DeleteTask()` en storage (transaction: borra task_logs + tasks), handler DELETE actualizado, toast "Task deleted"
- **C3**: Todos los `alert()` reemplazados con toast en HistoryView y ChatView
- **C4**: `useToast` + error toast en DashboardView, SettingsView, FilesView, ChannelsView
- **C5**: Texto en español corregido: "Recientes"→"Recent", "Antiguos"→"Oldest" en TasksView
- **C6**: Auto-refresh cada 30s con cleanup `onUnmounted` en MetricsView
- **Fix build**: `import useToast from` → `import { useToast } from` en ChatView y HistoryView (named export)

---

## Fase 7 — Mejoras de Baja Prioridad (Futuro)

### 7.1 Multi-usuario con RBAC
### 7.2 Code Sandbox en Browser (Pyodide)
### 7.3 Visual Workflow Builder (drag-and-drop)
### 7.4 Import de Conversaciones
### 7.5 PWA / Desktop App (Service Worker + Manifest)
### 7.6 Observabilidad (OpenTelemetry traces para llamadas LLM)
### 7.7 Branch/Fork de Conversaciones

---

## Ventajas Competitivas Existentes

| Feature | Estado |
|---|---|
| 9 canales nativos (Telegram, WhatsApp, Discord, Slack, etc.) | Implementado |
| Binary único con frontend embebido | Implementado |
| 10 proveedores LLM | Implementado |
| Task worker con IA (auto-procesa tareas) | Implementado |
| Skills como markdown (Git-friendly) | Implementado |
| Cron nativo (at/every/cron expressions + visual builder) | Implementado |
| Subagent spawning | Implementado |
| Email reports via SMTP | Implementado |
| Panel web completo (14 vistas) | Implementado |
| Voice Input (Groq STT) | Implementado |
| Knowledge Base / RAG (FTS5) | Implementado |
| Web Search toggle per-conversation | Implementado |
| API Docs (Swagger UI) | Implementado |
| MCP Client (Model Context Protocol) | Implementado |
| Session management (rename/archive/delete/fork) | Implementado |
| Cron visual builder (6 schedule types + TZ) | Implementado |
| Toast notifications system-wide | Implementado |
| Copy assistant messages | Implementado |
| Auto-refresh metrics | Implementado |

---

## Prioridades Inmediatas (próxima sesión)

> **Fases 1-6 + UX Refinement COMPLETADAS al 100%.** Todas las mejoras de alta y media prioridad están implementadas. El pase de refinamiento UX agregó gestión de sesiones, cron visual builder, y polish general.

1. ~~**Voice I/O** — Botón micrófono + STT via Groq~~ COMPLETADO
2. ~~**MCP Client** — JSON-RPC sobre STDIO, tool discovery dinámico~~ COMPLETADO
3. ~~**RAG/Knowledge Base** — FTS5, upload, chunking, query_knowledge tool, KnowledgeView~~ COMPLETADO
4. ~~**Web Search UI** — Toggle en ChatView, ExcludeTools flow~~ COMPLETADO
5. ~~**API Docs** — OpenAPI 3.0.3 spec + Swagger UI~~ COMPLETADO
6. ~~**UX Refinement** — Sessions CRUD, Cron visual builder, toast system-wide, quick bugs~~ COMPLETADO

**Siguiente: Fase 7** — Multi-usuario RBAC, Code Sandbox, Visual Workflow Builder, etc.
