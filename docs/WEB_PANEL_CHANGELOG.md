# KakoClaw Web Panel — Changelog Completo

Documentacion de todas las mejoras realizadas al panel web de KakoClaw (frontend Vue 3 + backend Go).

---

## Fase 1 — Bug Fixes Criticos

**6 bugs corregidos:**

| # | Archivo | Bug | Fix |
|---|---------|-----|-----|
| 1 | `taskService.js` | Metodos faltantes para sesiones y archivado | Agregados: `fetchChatSessions`, `fetchSessionMessages`, `archiveTask`, `unarchiveTask` |
| 2 | `chatStore.js` | Faltaban acciones `setMessages()` y `sendMessage()` | Implementadas ambas acciones en el store |
| 3 | `ReportsView.vue` | `sendReport` no manejaba fallback cuando WS no esta conectado | Agregado fallback a REST API cuando WebSocket no esta disponible |
| 4 | `TasksView.vue` | Doble extraccion de datos (`data.tasks.tasks`) | Corregido a extraccion simple `data.tasks` |
| 5 | `HistoryView.vue` | No mostraba sesiones correctamente | Reescrito completamente en Fase 2 |
| 6 | `TasksView.vue` | WebSocket no procesaba eventos correctamente | Corregido manejo de `task_updated`, `task_created`, `task_deleted` |

---

## Fase 2 — Historia y Filtrado Mejorado

### Backend
- **`pkg/storage/chat.go`**: `ListSessions()` reescrito con subconsulta para obtener conteo de mensajes por sesion (`SessionSummary.MessageCount`)
- **`pkg/web/server.go`**: Nuevo endpoint `GET /api/v1/chat/search` para busqueda full-text en mensajes

### Frontend
- **`HistoryView.vue`**: Reescritura completa con:
  - Barra de busqueda con debounce (300ms)
  - Filtro por tipo (Chat/Task)
  - Filtro por fecha (Today/7d/30d/All)
  - Badges de conteo de mensajes
  - Resaltado de coincidencias en resultados de busqueda
- **`ChatView.vue`**: Sidebar mejorado con iconos Chat/Task y conteo de mensajes

---

## Fase 3 — Superpoderes del Frontend

### 3.1–3.2 Markdown Rendering
- **`MarkdownRenderer.vue`** (NUEVO): Componente reutilizable con `markdown-it` + `highlight.js`
- Aplicado en `ChatView.vue` y `HistoryView.vue` para mensajes del asistente
- Dependencias agregadas: `markdown-it`, `highlight.js`

### 3.3–3.4 Sistema de Notificaciones Toast
- **`ToastContainer.vue`** (NUEVO): Contenedor de notificaciones con animaciones
- **`useToast.js`** (NUEVO): Composable reactivo (`toast.success()`, `toast.error()`, `toast.info()`)
- Integrado en: `MemoryView`, `ReportsView`, `TasksView`

### 3.5 Autocompletado de Comandos Slash
- **`ChatView.vue`**: Textarea con popup de autocompletado
- 9 comandos: `/task`, `/memory`, `/report`, `/help`, `/status`, `/clear`, `/search`, `/skills`, `/cron`
- Navegacion con teclado (flechas + Tab/Enter)

### 3.6 Busqueda de Tareas
- **Backend**: `GET /api/v1/tasks/search?q=...`
- **Frontend**: `searchTasks()` en `taskService.js`

### 3.7 Sistema de Task Logs
- **`pkg/storage/task_logs.go`** (NUEVO): Tabla `task_logs` con `AddTaskLog()` y `GetTaskLogs()`
- **`pkg/storage/sqlite.go`**: Migracion para crear tabla + indice
- **`pkg/web/server.go`**: Task worker ahora registra logs (started, completed, failed)
- **Endpoint**: `GET /api/v1/tasks/{id}/logs` (corregido, antes era stub)

### 3.8 Dashboard
- **`DashboardView.vue`** (NUEVO): Pagina de inicio con:
  - Tarjetas de estadisticas (total tasks, active, done, sessions)
  - Desglose por estado de tareas
  - Tareas y sesiones recientes
  - Acciones rapidas (New Chat, New Task, View History)
- Configurado como ruta por defecto (`/` -> `/dashboard`)

### 3.9 Build y Test
- Frontend: `npm run build` limpio
- Backend: `go build ./...` + `go test ./...` pasando

---

## Fase 4 — Funcionalidades Avanzadas

### 4.0 Infraestructura del Server

**`pkg/web/server.go`** — Campos opcionales agregados al struct `Server`:

```go
type Server struct {
    // ... campos existentes ...
    fullConfig      *config.Config
    cronService     *cron.CronService
    skillsLoader    *skills.SkillsLoader
    skillInstaller  *skills.SkillInstaller
    channelManager  *channels.Manager
}
```

**5 setter methods** para inyeccion de dependencias:
- `SetFullConfig(cfg *config.Config)`
- `SetCronService(cs *cron.CronService)`
- `SetChannelManager(cm *channels.Manager)`
- `SetSkills(loader *skills.SkillsLoader, installer *skills.SkillInstaller)`
- `SetStorage(store *storage.Storage)`

**`cmd/KakoClaw/main.go`** — Wiring en ambos entry points:
- `gatewayCmd()`: Conecta storage, cron, channels, config, skills
- `webCmd()`: Conecta storage, config, skills (sin cron/channels en modo web-only)

### 4.1 Skills Panel

**Backend** (`handlers_advanced.go`):
- `GET /api/v1/skills` — Lista skills instaladas
- `GET /api/v1/skills?type=available` — Lista skills disponibles en marketplace
- `GET /api/v1/skills/{name}` — Ver contenido de un skill
- `POST /api/v1/skills/install` — Instalar skill desde GitHub (`{"repository": "user/repo"}`)
- `DELETE /api/v1/skills/{name}` — Desinstalar skill

**Frontend** (`SkillsView.vue`):
- Tabs: Installed / Marketplace
- Tarjetas de skills con badges de source (local/github/builtin)
- Acciones: Install, Uninstall, View
- Modal viewer para ver contenido del skill

### 4.2 Cron Jobs Panel

**Backend** (`handlers_advanced.go`):
- `GET /api/v1/cron` — Lista todos los jobs (con `?include_disabled=true/false`)
- `POST /api/v1/cron` — Crear job (soporta schedule types: `at`, `every`, `cron`)
- `PATCH /api/v1/cron/{id}` — Enable/disable job
- `DELETE /api/v1/cron/{id}` — Eliminar job

**Frontend** (`CronView.vue`):
- Lista de jobs con enable/disable toggle y delete
- Modal de creacion con 3 tipos de schedule:
  - **Interval**: cada N minutos/horas
  - **Cron Expression**: formato cron estandar
  - **One-time**: fecha y hora especifica
- Entrega opcional a canal (Telegram, Discord, etc.)

### 4.3 Channels Panel

**Backend** (`handlers_advanced.go`):
- `GET /api/v1/channels` — Status de todos los canales

**Frontend** (`ChannelsView.vue`):
- Tarjetas por canal con indicadores de color:
  - Verde: Connected (enabled + running)
  - Amarillo: Enabled (enabled, not running)
  - Gris: Disabled
- Iconos por canal (Telegram, Discord, Slack, WhatsApp, etc.)
- 9 canales soportados: telegram, discord, slack, whatsapp, feishu, dingtalk, qq, maixcam, signal

### 4.4 Settings Panel

**Backend** (`handlers_advanced.go`):
- `GET /api/v1/config` — Configuracion completa (con API keys redactadas)
- Redaccion automatica de campos sensibles: `api_key`, `token`, `secret`, `password`, `webhook`

**Frontend** (`SettingsView.vue`):
- Vista con tabs: Agent, Providers, Channels, System
- Configuracion read-only redactada (seguridad)
- Formato JSON con highlighting

### 4.5 File Browser

**Backend** (`handlers_advanced.go`):
- `GET /api/v1/files/` — Listar directorio
- `GET /api/v1/files/{path}` — Listar subdirectorio o leer archivo (<1MB)
- **Proteccion path traversal**: Validacion con `filepath.Abs()` + `strings.HasPrefix()`

**Frontend** (`FilesView.vue`):
- Navegacion por directorios con breadcrumbs
- Iconos de carpeta/archivo
- Viewer para archivos <1MB
- Formateo de tamanos (KB/MB/GB)
- Ordenamiento: directorios primero, luego por nombre

### 4.6 Export

**Backend** (`handlers_advanced.go`):
- `GET /api/v1/export/tasks?format=json|csv` — Exportar todas las tareas
- `GET /api/v1/export/chat?session_id=...` — Exportar chat (todas las sesiones o una especifica)

**Frontend** (`advancedService.js`):
- `exportTasks(format)` — Abre descarga en nueva ventana
- `exportChat(sessionId)` — Abre descarga en nueva ventana

**UI Integration**:
- `TasksView.vue`: Dropdown "Export" con opciones JSON/CSV
- `HistoryView.vue`: Dropdown "Export" con "All Chats" y "Current Session"

### 4.7 Wiring Final

**`router/index.js`**: 11 rutas hijo registradas:
- `/dashboard`, `/chat`, `/tasks`, `/history`, `/memory`, `/reports`
- `/skills`, `/cron`, `/channels`, `/settings`, `/files` (NUEVAS)

**`Sidebar.vue`**: 3 secciones de navegacion:
1. **Primary**: Dashboard, Chat, Tasks
2. **Tools**: Skills, Cron Jobs, Channels, Files
3. **Secondary**: History, Memory, Reports, Settings

---

## Auditoria y Fixes de Calidad

Despues de completar las 4 fases, se realizo una auditoria completa del codigo (frontend + backend).

### Bugs Criticos Encontrados y Corregidos

| # | Severidad | Archivo | Problema | Fix |
|---|-----------|---------|----------|-----|
| C1 | **ALTA** | `cmd/KakoClaw/main.go` | `SetStorage()` no se llamaba en `gatewayCmd()` — storage era nil, todos los endpoints de tasks/chat devolvian 503 | Agregada inicializacion de storage antes de `webServer.Start()` |
| C2 | **ALTA** | `handlers_advanced.go` | Proteccion path traversal insuficiente (`strings.ReplaceAll("..", "")` es bypassable) | Reemplazado con `filepath.Abs()` + validacion `strings.HasPrefix()` |

### Bugs Medios Encontrados y Corregidos

| # | Severidad | Archivo | Problema | Fix |
|---|-----------|---------|----------|-----|
| M1 | **MEDIA** | `task.go`, `chat.go`, `task_logs.go` | Faltaba `rows.Err()` check despues de `for rows.Next()` — 6 funciones afectadas | Agregado `rows.Err()` check en las 6 funciones |
| M2 | **MEDIA** | `sqlite.go` | `query[:11]` podia causar panic si `len(query)` era entre 6-10 | Cambiado a `strings.HasPrefix(query, "ALTER TABLE")` |

### Fixes de Calidad (Low)

| # | Archivo | Fix |
|---|---------|-----|
| L1 | `cmd/KakoClaw/main.go` | Agregado `defer store.Close()` en `webCmd()` para cerrar SQLite correctamente |
| L2 | `MainLayout.vue` | Eliminados imports muertos (`onUnmounted`, `useAuthStore`) |

---

## Referencia de API REST

Todos los endpoints requieren autenticacion JWT (header `Authorization: Bearer <token>`).

### Core
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | Login, devuelve JWT |
| POST | `/api/v1/auth/change-password` | Cambiar password |
| GET | `/api/v1/auth/me` | Info del usuario actual |
| GET | `/api/v1/health` | Health check |

### Tasks
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| GET | `/api/v1/tasks` | Listar tareas (`?include_archived=true`) |
| POST | `/api/v1/tasks` | Crear tarea |
| GET | `/api/v1/tasks/{id}` | Obtener tarea |
| PUT | `/api/v1/tasks/{id}` | Actualizar tarea |
| PATCH | `/api/v1/tasks/{id}` | Actualizar status |
| DELETE | `/api/v1/tasks/{id}` | Eliminar tarea |
| POST | `/api/v1/tasks/{id}/archive` | Archivar tarea |
| POST | `/api/v1/tasks/{id}/unarchive` | Desarchivar tarea |
| GET | `/api/v1/tasks/{id}/logs` | Logs de ejecucion |
| GET | `/api/v1/tasks/search?q=...` | Buscar tareas |

### Chat
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| GET | `/api/v1/chat/sessions` | Listar sesiones con conteo de mensajes |
| GET | `/api/v1/chat/sessions/{id}` | Mensajes de una sesion |
| GET | `/api/v1/chat/search?q=...` | Buscar en mensajes |

### Memory
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| GET | `/api/v1/memory/longterm` | Leer memoria largo plazo |
| PUT | `/api/v1/memory/longterm` | Guardar memoria largo plazo |
| GET | `/api/v1/memory/daily?days=7` | Leer notas diarias |

### Skills (Fase 4)
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| GET | `/api/v1/skills` | Lista skills instaladas |
| GET | `/api/v1/skills?type=available` | Lista skills del marketplace |
| GET | `/api/v1/skills/{name}` | Contenido de un skill |
| POST | `/api/v1/skills/install` | Instalar skill (`{"repository": "user/repo"}`) |
| DELETE | `/api/v1/skills/{name}` | Desinstalar skill |

### Cron Jobs (Fase 4)
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| GET | `/api/v1/cron` | Lista jobs (`?include_disabled=true`) |
| POST | `/api/v1/cron` | Crear job |
| PATCH | `/api/v1/cron/{id}` | Enable/disable job (`{"enabled": true}`) |
| DELETE | `/api/v1/cron/{id}` | Eliminar job |

**Formato de creacion de Cron Job:**
```json
{
  "name": "Daily Report",
  "schedule": {
    "kind": "every",
    "every_ms": 86400000
  },
  "payload": {
    "kind": "message",
    "message": "Generate daily report",
    "deliver": "channel",
    "channel": "telegram"
  }
}
```

Schedule types:
- `"kind": "every"` + `"every_ms": 3600000` — Cada N milisegundos
- `"kind": "cron"` + `"expr": "0 9 * * *"` — Expresion cron
- `"kind": "at"` + `"at_ms": 1708300800000` — Fecha/hora especifica (Unix ms)

### Channels (Fase 4)
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| GET | `/api/v1/channels` | Status de todos los canales |

### Config (Fase 4)
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| GET | `/api/v1/config` | Configuracion completa (API keys redactadas) |

### Files (Fase 4)
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| GET | `/api/v1/files/` | Listar directorio raiz (workspace) |
| GET | `/api/v1/files/{path}` | Listar subdirectorio o leer archivo |

### Export (Fase 4)
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| GET | `/api/v1/export/tasks?format=json` | Exportar tareas (json o csv) |
| GET | `/api/v1/export/chat` | Exportar todos los chats |
| GET | `/api/v1/export/chat?session_id=...` | Exportar sesion especifica |

### WebSockets
| Endpoint | Descripcion |
|----------|-------------|
| `/ws/chat` | Chat en tiempo real con el agente AI |
| `/ws/tasks` | Actualizaciones en tiempo real de tareas |

---

## Configuracion Necesaria

### `config.json` — Seccion Web
```json
{
  "web": {
    "enabled": true,
    "host": "0.0.0.0",
    "port": 8080,
    "auth": {
      "username": "admin",
      "password_hash": "$2a$10$..."
    },
    "jwt_secret": "your-secret-key",
    "session_timeout": "24h"
  }
}
```

### Dependencias Frontend
Instaladas automaticamente con `npm install` en `pkg/web/frontend/`:
- `vue` 3.x, `vue-router`, `pinia` — Framework
- `axios` — HTTP client
- `markdown-it` — Markdown rendering (Fase 3)
- `highlight.js` — Code syntax highlighting (Fase 3)
- `tailwindcss` — CSS utility framework

### Build
```bash
# Frontend
cd pkg/web/frontend && npm install && npm run build

# Go binary (incluye frontend embebido)
go build -o KakoClaw ./cmd/KakoClaw

# O con Make
make build
```

### Modos de Ejecucion

**Gateway mode** (completo — channels + cron + web):
```bash
./KakoClaw gateway
```
- Inicializa: storage, agent loop, cron, channels, heartbeat, web server
- Todos los endpoints disponibles incluyendo cron y channels

**Web-only mode**:
```bash
./KakoClaw web
```
- Inicializa: storage, agent loop, web server
- Sin cron ni channels (endpoints devuelven arrays vacios o 503)
- Skills y config si disponibles

---

## Archivos Nuevos Creados

### Backend (Go)
| Archivo | Descripcion | Lineas |
|---------|-------------|--------|
| `pkg/web/handlers_advanced.go` | Handlers REST: skills, cron, channels, config, files, export | ~530 |
| `pkg/storage/task_logs.go` | TaskLog struct + CRUD para logs de tareas | ~45 |

### Frontend (Vue 3)
| Archivo | Descripcion |
|---------|-------------|
| `views/DashboardView.vue` | Dashboard con stats, desglose, recientes, quick actions |
| `views/SkillsView.vue` | Panel de skills (installed + marketplace) |
| `views/CronView.vue` | Panel de cron jobs (list + create modal) |
| `views/ChannelsView.vue` | Panel de canales con status cards |
| `views/SettingsView.vue` | Settings con tabs (read-only, redacted) |
| `views/FilesView.vue` | File browser con breadcrumbs y viewer |
| `components/Chat/MarkdownRenderer.vue` | Markdown + code highlighting |
| `components/Layout/ToastContainer.vue` | Sistema de notificaciones toast |
| `composables/useToast.js` | Composable reactivo para toasts |
| `services/advancedService.js` | API client para endpoints avanzados |

### Archivos Modificados
| Archivo | Cambios |
|---------|---------|
| `pkg/web/server.go` | +5 campos struct, +5 setters, +12 rutas, task worker logging, fix logs endpoint |
| `pkg/storage/sqlite.go` | +tabla task_logs, +indice, fix migrate bounds check, +import strings |
| `pkg/storage/chat.go` | +SessionSummary.MessageCount, rewrite ListSessions, +rows.Err() checks |
| `pkg/storage/task.go` | +rows.Err() checks en ListTasks y SearchTasks |
| `cmd/KakoClaw/main.go` | +SetStorage en gatewayCmd, +defer store.Close en webCmd, +skills/config wiring |
| `router/index.js` | +11 rutas hijo (6 existentes + 5 nuevas) |
| `Sidebar.vue` | +seccion Tools con 4 items, +Settings en secondary |
| `MainLayout.vue` | +ToastContainer, cleanup dead imports |
| `ChatView.vue` | +MarkdownRenderer, +autocomplete, +sidebar mejorado |
| `HistoryView.vue` | Reescritura completa + export buttons |
| `TasksView.vue` | +export dropdown, +advancedService import, bug fixes |
| `ReportsView.vue` | +toast, +WS fallback |
| `MemoryView.vue` | +toast |
| `taskService.js` | +fetchChatSessions, +fetchSessionMessages, +searchMessages, +searchTasks, +archive/unarchive |
| `chatStore.js` | +setMessages(), +sendMessage() |
| `package.json` | +markdown-it, +highlight.js |

---

## Fase 5 — Mejoras Competitivas (Analisis vs Open WebUI, Dify, LobeChat, etc.)

> Basado en analisis competitivo contra: Open WebUI, Dify, LobeChat, Flowise, n8n, Langflow, BotPress, AutoGPT.

### 5.0 Fix Critico: Import Faltante
- **`pkg/web/server.go`**: Faltaba import `"github.com/sipeed/KakoClaw/pkg/providers"` — el handler `handleModels` (escrito en sesion anterior) referenciaba `providers.GetProviderForModel()` pero el import no existia. Bloqueaba compilacion.

### 5.1 API de Modelos Disponibles
**Backend** (`pkg/web/server.go:1201-1363`):
- `GET /api/v1/models` — Retorna proveedores configurados con sus modelos
- Detecta proveedores activos por API key o auth_method configurados
- Soporta 10 proveedores: Anthropic, OpenAI, OpenRouter, Groq, Gemini, Zhipu, Moonshot, Nvidia, Ollama, VLLM
- Respuesta incluye `current_model`, `current_provider`, y array de `providers` con sus modelos

**Frontend** (`advancedService.js`):
- `fetchModels()` — Nuevo metodo que consume `GET /api/v1/models`

### 5.2 Selector de Modelo por Conversacion

**Backend** — Model override pipeline:
| Archivo | Cambio |
|---------|--------|
| `pkg/agent/loop.go` | `processOptions.ModelOverride` — nuevo campo |
| `pkg/agent/loop.go` | `ProcessDirectWithModel(ctx, content, sessionKey, modelOverride)` — nuevo metodo publico |
| `pkg/agent/loop.go` | `processMessageWithModel(ctx, msg, modelOverride)` — nuevo metodo privado |
| `pkg/agent/loop.go` | `runLLMIteration()` — usa variable local `model` (override o default) en vez de `al.model` directo |
| `pkg/web/server.go` | `handleChatWS` — parsea campo `model` del JSON del WebSocket, lo pasa a `ProcessDirectWithModel()` |

**Frontend** — Model selection UI:
| Archivo | Cambio |
|---------|--------|
| `chatStore.js` | Nuevos: `selectedModel`, `currentModel`, `availableProviders`, `allModels` (computed), `setModelsData()`, `setSelectedModel()` |
| `ChatView.vue` | Top bar con dropdown `<select>` de modelos, muestra `provider/model (default)` |
| `ChatView.vue` | `onMounted` llama `advancedService.fetchModels()` y carga en store |
| `ChatView.vue` | `sendMessage()` incluye `model: chatStore.selectedModel` en el payload WebSocket |

**Protocolo WebSocket actualizado:**
```json
// Client -> Server (antes)
{ "type": "message", "content": "...", "session_id": "web:chat:..." }

// Client -> Server (ahora)
{ "type": "message", "content": "...", "session_id": "web:chat:...", "model": "claude-sonnet-4-20250514" }
```
El campo `model` es opcional — si esta vacio o ausente, se usa el default de `config.json`.

### 5.3 Boton Regenerar Respuesta

**Frontend** (`ChatView.vue`):
- Icono refresh (SVG) aparece al hover sobre el ultimo mensaje del asistente
- `isLastAssistantMessage(msg)` — determina si el mensaje es el ultimo del asistente
- `regenerateResponse()`:
  1. Encuentra el ultimo mensaje del usuario
  2. Elimina el ultimo mensaje del asistente del array
  3. Reenvia el mensaje del usuario via WebSocket con el modelo seleccionado actualmente
- Disabled durante `isLoading`

### 5.4 Toggle Dark/Light Theme

**Sistema de CSS Custom Properties:**

| Archivo | Cambio |
|---------|--------|
| `tailwind.config.js` | Colores cambiados de hex hardcoded a `rgb(var(--pc-*) / <alpha-value>)` — soporta opacity modifiers de Tailwind |
| `globals.css` | Variables definidas como RGB triplets: `:root` (light) y `.dark` (dark) |
| `globals.css` | `body` usa `rgb(var(--pc-bg))` y `rgb(var(--pc-text))` con transicion de 0.2s |
| `globals.css` | Scrollbar usa variables `--pc-scrollbar-*` |
| `index.html` | Script inline previene flash: lee `localStorage('ui.theme')` antes de que Vue monte |

**Paleta Light Theme:**
| Variable | Color | Hex |
|----------|-------|-----|
| `--pc-bg` | Slate 50 | `#f8fafc` |
| `--pc-surface` | White | `#ffffff` |
| `--pc-surface-hover` | Slate 100 | `#f1f5f9` |
| `--pc-border` | Slate 200 | `#e2e8f0` |
| `--pc-text` | Slate 900 | `#0f172a` |
| `--pc-text-secondary` | Slate 500 | `#64748b` |

**Paleta Dark Theme** (misma que antes, ahora via variables):
| Variable | Color | Hex |
|----------|-------|-----|
| `--pc-bg` | Slate 900 | `#0f172a` |
| `--pc-surface` | Slate 800 | `#1e293b` |
| `--pc-surface-hover` | Slate 700 | `#334155` |
| `--pc-border` | Slate 700 | `#334155` |
| `--pc-text` | Slate 50 | `#f8fafc` |
| `--pc-text-secondary` | Slate 400 | `#94a3b8` |

**Nota**: El toggle en sidebar y la persistencia en localStorage ya existian de sesiones anteriores (`Sidebar.vue`, `uiStore.js`). Esta sesion agrego el sistema de variables CSS que hace que el toggle funcione visualmente.

### 5.5 Streaming Token-by-Token
**Estado**: COMPLETADO

**Backend — Provider Streaming:**
| Archivo | Cambio |
|---------|--------|
| `pkg/providers/types.go` | `StreamChunk` struct — representa un token/fragmento del stream |
| `pkg/providers/types.go` | `StreamingLLMProvider` interface — extiende `LLMProvider` con `ChatStream()` |
| `pkg/providers/http_provider.go` | `ChatStream()` — SSE streaming para OpenAI-compatible APIs (OpenAI, OpenRouter, Groq, Zhipu, Moonshot, Nvidia, VLLM) |
| `pkg/providers/ollama_provider.go` | `ChatStream()` — Streaming nativo para Ollama (NDJSON line-by-line) |

**Backend — Agent Loop Streaming:**
| Archivo | Cambio |
|---------|--------|
| `pkg/agent/loop.go` | `StreamCallback` type — `func(token string) error` |
| `pkg/agent/loop.go` | `ProcessDirectWithModelStream()` — nuevo metodo publico para streaming |
| `pkg/agent/loop.go` | `SupportsStreaming()` — verifica si el provider implementa `StreamingLLMProvider` |
| `pkg/agent/loop.go` | `runAgentLoopStream()` — como `runAgentLoop()` pero usa streaming en la respuesta final |
| `pkg/agent/loop.go` | `runLLMIterationStream()` — loop de iteraciones LLM con streaming; tool calls se ejecutan sin streaming, solo la respuesta texto final se streamea |

**Backend — WebSocket Handler:**
| Archivo | Cambio |
|---------|--------|
| `pkg/web/server.go` | `handleChatWS()` — detecta si el provider soporta streaming; si lo soporta, envia tokens progresivamente |

**Protocolo WebSocket Streaming:**
```json
// Server -> Client: Stream start
{ "type": "stream_start" }

// Server -> Client: Token (repetido N veces)
{ "type": "stream", "content": "Hel" }
{ "type": "stream", "content": "lo " }
{ "type": "stream", "content": "world" }

// Server -> Client: Stream end (contiene respuesta completa como authoritative)
{ "type": "stream_end", "content": "Hello world" }

// Server -> Client: Ready
{ "type": "ready" }
```

**Fallback**: Si el provider no implementa `StreamingLLMProvider` (ClaudeProvider con SDK Anthropic, CodexProvider con SDK OpenAI), el handler cae al flujo no-streaming existente (`type: "message"` con respuesta completa).

**Frontend — Streaming UI:**
| Archivo | Cambio |
|---------|--------|
| `chatStore.js` | `isStreaming`, `streamingMessageId` — estado de streaming |
| `chatStore.js` | `startStreamingMessage()` — crea mensaje assistant vacio con `streaming: true` |
| `chatStore.js` | `appendStreamToken(token)` — concatena token al mensaje en streaming |
| `chatStore.js` | `endStreamingMessage(finalContent)` — finaliza el streaming, marca `streaming: false` |
| `ChatView.vue` | Handler para `stream_start`, `stream`, `stream_end` message types |
| `ChatView.vue` | Rendering condicional: texto plano + cursor parpadeante durante streaming, Markdown despues |
| `ChatView.vue` | Loading indicator oculto durante streaming (el mensaje parcial ya es visible) |
| `ChatView.vue` | CSS: `.streaming-cursor` con animacion `blink` de 0.8s |

---

## Referencia de API REST (Actualizada Fase 5)

### Models (Fase 5 — NUEVO)
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| GET | `/api/v1/models` | Proveedores disponibles, sus modelos, y modelo actual |

**Respuesta ejemplo:**
```json
{
  "current_model": "claude-sonnet-4-20250514",
  "current_provider": "anthropic",
  "providers": [
    {
      "name": "anthropic",
      "enabled": true,
      "is_active": true,
      "models": [
        { "id": "claude-sonnet-4-20250514", "provider": "anthropic" },
        { "id": "claude-3-5-haiku-20241022", "provider": "anthropic" }
      ]
    },
    {
      "name": "openai",
      "enabled": true,
      "is_active": false,
      "models": [
        { "id": "gpt-4o", "provider": "openai" }
      ]
    }
  ]
}
```

---

## Notas Conocidas

1. **`memoryService.js`** devuelve respuestas Axios crudas (con `.data` extra) a diferencia de los otros services que desenvuelven `response.data`. No es un bug pero es una inconsistencia.
2. **`server_test.go`** tiene errores LSP pre-existentes (`s.tasks undefined`, `newTaskStore undefined`) — no causados por nuestros cambios.
3. **Config es read-only** en el panel web. Para editar configuracion, modificar `~/.KakoClaw/config.json` directamente.
4. **File browser** esta limitado al directorio workspace. Archivos >1MB no se pueden ver en el viewer (solo listar).

---

## Fase UX — Refinamiento Completo de la Experiencia de Usuario

> Pase completo de UX polish dividido en 3 sub-fases (A/B/C). Todas completadas.

### UX-A: Gestión de Sesiones de Chat

**Backend — Nueva tabla `sessions`:**
| Archivo | Cambio |
|---------|--------|
| `pkg/storage/sqlite.go` | Nueva tabla `sessions` en migración + `migrateSessions()` backfill |
| `pkg/storage/chat.go` | Reescritura completa: nuevos tipos `Session`, `ensureSession()`, CRUD completo |
| `pkg/web/server.go` | Handlers GET/DELETE/PATCH para `/api/v1/chat/sessions/{id}`, query params en GET |
| `pkg/web/handlers_advanced.go` | `ListSessions` call actualizado con nuevos parámetros en export handler |

**Esquema `sessions`:**
```sql
CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    archived INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**API REST (nuevos/actualizados):**
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| GET | `/api/v1/chat/sessions?archived=true&limit=50&offset=0` | Listar sesiones con filtro y paginación |
| DELETE | `/api/v1/chat/sessions/{id}` | Eliminar sesión + cascade delete mensajes |
| PATCH | `/api/v1/chat/sessions/{id}` | Actualizar título y/o archived status |

**Frontend — ChatView.vue:**
- Menú contextual (click derecho / botón ⋯) en sesiones del sidebar: Rename, Archive, Delete
- Renombrado inline con input text
- `taskService.deleteSession(id)`, `taskService.updateSession(id, data)` — nuevos métodos
- `taskService.fetchChatSessions({archived, limit, offset})` — actualizado con query params
- Toast notifications en vez de `alert()` para todas las acciones

**Frontend — HistoryView.vue:**
- Botones de acción (Rename/Archive/Delete) por sesión
- Checkbox "Archived" para filtrar sesiones archivadas
- Usa `taskService` actualizado para todas las operaciones

**Frontend — Copiar mensajes:**
- Botón copiar (clipboard icon) en mensajes del asistente en ChatView y HistoryView
- `navigator.clipboard.writeText()` con toast de confirmación

### UX-B: Cron Jobs — Visual Builder

**Backend — Validación y edición:**
| Archivo | Cambio |
|---------|--------|
| `pkg/cron/service.go` | `ValidateSchedule()` — valida at/every/cron; usa `gronx.IsValid()` para cron expressions |
| `pkg/cron/service.go` | `UpdateJob()` — update completo con recompute de next run y validación |
| `pkg/cron/service.go` | `computeNextRun()` — fix TZ: aplica `time.LoadLocation()` y convierte a UTC |
| `pkg/web/handlers_advanced.go` | `PUT /api/v1/cron/{id}` — nuevo handler para edición completa |
| `pkg/web/handlers_advanced.go` | `POST /api/v1/cron` — ahora valida schedule (retorna 400, no 500) |

**API REST (nuevo):**
| Metodo | Endpoint | Descripcion |
|--------|----------|-------------|
| PUT | `/api/v1/cron/{id}` | Editar job completo (nombre, schedule, payload) |

**Frontend — CronView.vue (reescritura completa):**
- 6 tipos de schedule via UI visual:
  - **Daily**: hora específica cada día
  - **Weekly**: días de la semana + hora
  - **Monthly**: día del mes + hora
  - **Interval**: cada N minutos/horas
  - **One-time**: fecha y hora específica
  - **Custom**: expresión cron libre (5 campos)
- Auto-generación de cron expressions desde los campos visuales
- Preview de próximas 3 ejecuciones calculadas en frontend
- Modal de edición con reverse-parsing (cron expression → campos visuales)
- Diálogo de confirmación para eliminar jobs
- Selector de timezone con autodetección del timezone del navegador
- Display amigable de schedules: ej. "Every Mon, Wed at 09:00" en vez de "0 9 * * 1,3"

**Frontend — advancedService.js:**
- `updateCronJob(id, data)` — `PUT /api/v1/cron/${id}` — nuevo método

### UX-C: Quick Bugs y Polish

| # | Archivo | Cambio |
|---|---------|--------|
| C1 | `TaskDetailsModal.vue` | Fixed `log.action`→`log.event`, `log.details`→`log.message` (campos correctos de la tabla) |
| C2 | `pkg/storage/task.go` + `server.go` | `DeleteTask(id)` con transaction (borra task_logs + tasks), handler DELETE actualizado |
| C3 | `ChatView.vue`, `HistoryView.vue` | Todos los `alert()` reemplazados con `toast.success()`/`toast.error()` |
| C4 | `DashboardView.vue`, `SettingsView.vue`, `FilesView.vue`, `ChannelsView.vue` | Agregado `useToast` + error toast en catch de fetch failures |
| C5 | `TasksView.vue` | Texto en español corregido: "Recientes"→"Recent", "Antiguos"→"Oldest" |
| C6 | `MetricsView.vue` | Auto-refresh cada 30s con `setInterval` + cleanup en `onUnmounted` |

**Build fix:**
- `ChatView.vue` y `HistoryView.vue` tenían `import useToast from` (default import) pero `useToast.js` exporta named: `export function useToast()`. Corregido a `import { useToast } from`.

### Archivos Nuevos (Fase UX)

Ninguno — todos los cambios fueron sobre archivos existentes.

### Archivos Modificados (Fase UX)

| Archivo | Cambios |
|---------|---------|
| `pkg/storage/sqlite.go` | +tabla sessions, +migrateSessions() |
| `pkg/storage/chat.go` | Reescritura: Session type, ensureSession, CRUD, ListSessions con filtros |
| `pkg/storage/task.go` | +DeleteTask() con transaction |
| `pkg/cron/service.go` | +ValidateSchedule(), +UpdateJob(), fix TZ en computeNextRun() |
| `pkg/web/server.go` | Handlers actualizados: chat sessions GET/DELETE/PATCH, task DELETE |
| `pkg/web/handlers_advanced.go` | +PUT /api/v1/cron/{id}, validación en POST cron, ListSessions actualizado |
| `ChatView.vue` | Context menu, inline rename, copy button, toast, fixed import |
| `HistoryView.vue` | Session actions, archived filter, copy button, toast, fixed import |
| `CronView.vue` | Reescritura completa: visual builder 6 types, edit modal, TZ selector |
| `TasksView.vue` | Fixed Spanish text, toast "Task deleted" |
| `TaskDetailsModal.vue` | Fixed log field names |
| `DashboardView.vue` | +useToast, error toast |
| `SettingsView.vue` | +useToast, error toast |
| `FilesView.vue` | +useToast, error toast |
| `ChannelsView.vue` | +useToast, error toast |
| `MetricsView.vue` | +30s auto-refresh with onUnmounted cleanup |
| `taskService.js` | +deleteSession, +updateSession, updated fetchChatSessions |
| `advancedService.js` | +updateCronJob(id, data) |
