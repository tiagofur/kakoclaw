# 🐸 KakoClaw — Revisión Completa & Changelog

> Este documento es el historial vivo de mejoras, correcciones y estado de la app.
> Se actualiza conforme avanzamos.

**Última actualización:** 2026-02-20

---

## 📊 Resumen Ejecutivo

| Métrica                                  | Valor                                        |
| ---------------------------------------- | -------------------------------------------- |
| **Total endpoints API**                  | 45+                                          |
| **Vistas frontend**                      | 15                                           |
| **Paquetes backend**                     | 22                                           |
| **Funciones completadas**                | ~50                                          |
| **Bugs encontrados**                     | 3 (1 medio, 2 menores)                       |
| **Code smells**                          | 3 (archivos muy largos, confirm nativo)      |
| **Features faltantes (alta prioridad)**  | 4                                            |
| **Features faltantes (media prioridad)** | 6                                            |
| **Features faltantes (baja prioridad)**  | 5                                            |
| **Fases completadas**                    | 5, 6, UX ✅                                  |
| **Fase pendiente**                       | 7 (Multi-usuario, Sandbox, Visual Workflows) |

---

## 1. ✅ Funciones Completadas

### Core / Chat

| Función                              | Estado | Notas                                |
| ------------------------------------ | ------ | ------------------------------------ |
| Chat con IA vía WebSocket            | ✅     | Streaming token-by-token             |
| Selector de modelo por conversación  | ✅     | Dropdown con todos los proveedores   |
| Regenerar última respuesta           | ✅     | Botón hover en mensaje del asistente |
| Búsqueda web toggle por conversación | ✅     | Icono lupa en ChatView               |
| Cancelar ejecución del agente        | ✅     | `POST /api/v1/chat/cancel`           |
| Copiar mensajes del asistente        | ✅     | Botón copy en ChatView + HistoryView |
| Voice Input (STT)                    | ✅     | Groq Whisper via micrófono           |
| Fork/branch de conversaciones        | ✅     | `POST /api/v1/chat/fork`             |

### Sesiones & Historial

| Función                         | Estado | Notas                                |
| ------------------------------- | ------ | ------------------------------------ |
| CRUD completo de sesiones       | ✅     | Crear, renombrar, archivar, eliminar |
| Filtrar historial por archivado | ✅     | Checkbox "Archived" en HistoryView   |
| Búsqueda en mensajes            | ✅     | `GET /api/v1/chat/search?q=`         |
| Continuar conversación pasada   | ✅     | Seleccionar sesión desde historial   |

### Tareas (Task Board)

| Función                               | Estado | Notas                                |
| ------------------------------------- | ------ | ------------------------------------ |
| CRUD de tareas                        | ✅     | Backlog/Todo/In Progress/Review/Done |
| Drag & Drop de tareas                 | ✅     | Cambio de status via drag            |
| Detalle de tarea con logs y resultado | ✅     | Modal con logs del agente            |
| Task worker (auto-procesa tareas)     | ✅     | Worker en background                 |
| Archivar/Desarchivar tareas           | ✅     | Endpoints `/archive` y `/unarchive`  |
| Eliminar tareas                       | ✅     | Con cascade delete de logs           |
| Búsqueda de tareas                    | ✅     | `GET /api/v1/tasks/search?q=`        |
| Comandos `/task` en chat              | ✅     | list, run, move, etc.                |

### Skills

| Función                     | Estado | Notas                           |
| --------------------------- | ------ | ------------------------------- |
| Listar skills instalados    | ✅     | SkillsView                      |
| Instalar/Desinstalar skills | ✅     | Desde marketplace o repositorio |
| Crear skills con IA         | ✅     | Generación de draft + creación  |
| Editar skills existentes    | ✅     | Implementado en sesión reciente |

### Cron / Tareas Programadas

| Función                            | Estado | Notas                                              |
| ---------------------------------- | ------ | -------------------------------------------------- |
| 6 tipos de schedule                | ✅     | Daily, Weekly, Monthly, Interval, One-time, Custom |
| Visual builder de cron expressions | ✅     | UI con preview de próximas 3 ejecuciones           |
| CRUD completo de cron jobs         | ✅     | Crear, editar, eliminar, toggle                    |
| Selector de timezone               | ✅     | En CronView                                        |
| Ejecutar job manualmente           | ✅     | Botón "Run now"                                    |

### Knowledge Base / RAG

| Función                          | Estado | Notas                          |
| -------------------------------- | ------ | ------------------------------ |
| Upload de documentos             | ✅     | Drag-and-drop en KnowledgeView |
| Búsqueda full-text (FTS5)        | ✅     | BM25 ranking                   |
| Tool `query_knowledge` en agente | ✅     | RAG automático                 |
| Eliminar documentos              | ✅     | Delete por ID                  |

### Workflows

| Función                  | Estado | Notas                             |
| ------------------------ | ------ | --------------------------------- |
| Crear/Editar workflows   | ✅     | Pipeline visual con drag-and-drop |
| 3 tipos de paso          | ✅     | Prompt, Tool, Condition           |
| Ejecutar workflows       | ✅     | Con resultados inline             |
| Historial de ejecuciones | ✅     | Botón "History" por workflow      |
| Test Run desde editor    | ✅     | Auto-save + run                   |

### Settings / Config

| Función                       | Estado | Notas                                        |
| ----------------------------- | ------ | -------------------------------------------- |
| Configuración de agentes      | ✅     | Modelo, temperatura, max tokens, iteraciones |
| Configuración de proveedores  | ✅     | API key, API base por proveedor              |
| Configuración de canales      | ✅     | Telegram, Discord, QQ, DingTalk, etc.        |
| Configuración de herramientas | ✅     | Web search, MCP                              |
| Cambio de contraseña          | ✅     | Modal dedicado                               |
| Dark/Light theme              | ✅     | Toggle con persistencia                      |

### Otros

| Función                                 | Estado | Notas                                 |
| --------------------------------------- | ------ | ------------------------------------- |
| File browser del workspace              | ✅     | Con breadcrumbs y viewer              |
| Descarga de archivos (individual + zip) | ✅     | Implementado recientemente            |
| Dashboard con métricas                  | ✅     | Resumen general                       |
| MCP Server management                   | ✅     | Lista, status, reconnect              |
| Metrics/Observability                   | ✅     | Auto-refresh cada 30s                 |
| Export tasks (JSON/CSV)                 | ✅     | Descarga directa                      |
| Export chat (JSON)                      | ✅     | Por sesión o todas                    |
| Import conversaciones                   | ✅     | ChatGPT, Claude, KakoClaw formats     |
| Backup/Restore completo                 | ✅     | DB + workspace + config + env         |
| API Docs (Swagger UI)                   | ✅     | OpenAPI 3.0.3 en `/api/docs`          |
| 9 canales de mensajería                 | ✅     | Telegram, Discord, QQ, DingTalk, etc. |
| Memoria long-term + daily notes         | ✅     | MemoryView                            |
| Reports view                            | ✅     | ReportsView                           |

---

## 2. 🔧 Funciones que Podemos Mejorar

| #   | Área           | Descripción                                                                     | Prioridad   |
| --- | -------------- | ------------------------------------------------------------------------------- | ----------- |
| M1  | File Browser   | Sin funcionalidad de upload de archivos al workspace                            | ✅ Resuelto |
| M2  | Knowledge Base | Sin edición/actualización de documentos ni preview de chunks                    | ✅ Resuelto |
| M3  | Dashboard      | Falta gráficas de tendencia, estadísticas de uso por modelo, actividad reciente | ✅ Resuelto |
| M4  | Memory View    | Muy simple — falta búsqueda, edición inline, timeline de daily notes            | ✅ Resuelto |
| M5  | History View   | Sin paginación real, sin búsqueda full-text desde HistoryView                   | ✅ Resuelto |
| M6  | WorkflowView   | Usa `confirm()` nativo + toast propio en vez de `useToast` composable           | ✅ Resuelto |
| M7  | ChatView       | 949 líneas — demasiado largo, extraer componentes                               | ✅ Resuelto |
| M8  | SettingsView   | 773 líneas — cada tab debería ser componente separado                           | ✅ Resuelto |
| M9  | Backup         | Sin backup programado (integrar con cron)                                       | Baja        |
| M10 | Streaming UX   | Sin feedback visual de tools durante streaming                                  | ✅ Resuelto |

---

## 3. 🚨 Bugs y Errores de Código

| #   | Severidad | Archivo                    | Descripción                                                                                                                                                                                          | Estado      |
| --- | --------- | -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- |
| B1  | 🔴 Media  | `server.go:543-573`        | Task Archive no hace broadcast WebSocket — otros clientes no reciben la actualización en real-time. Hay un TODO inline. Unarchive tiene el mismo problema.                                           | ✅ Resuelto |
| B2  | 🟡 Menor  | `advancedService.js:94-97` | `downloadFile` usa `window.open()` sin JWT token en header. Exportaciones (`exportTasks`, `exportChat`) tienen el mismo patrón. Puede fallar con reverse proxy estricto.                             | ✅ Resuelto |
| B3  | 🟡 Menor  | `server.go:766-771`        | `defer` dentro del `for` loop del WebSocket — la limpieza de `activeExecution` se acumula y solo se ejecuta al cerrar la conexión, no después de cada mensaje. Causa memory leak en sesiones largas. | ✅ Resuelto |
| B4  | 🟢 Info   | `server.go:285-306`        | Content type detection manual con `if/else` en vez de `mime.TypeByExtension()`. Frágil pero funcional.                                                                                               | ✅ Resuelto |
| B5  | 🟢 Info   | `server.go:328`            | SPA `index.html` con `Cache-Control: public, max-age=3600` — debería ser `no-cache` para que usuarios obtengan última versión.                                                                       | ✅ Resuelto |

---

## 4. 📋 Features Faltantes

### Alta Prioridad

| #   | Función                       | Justificación                                                               | Referencia  |
| --- | ----------------------------- | --------------------------------------------------------------------------- | ----------- |
| F1  | Multi-usuario con RBAC        | Esencial para despliegues en equipo. Open WebUI ya lo tiene.                | ✅ Resuelto |
| F2  | PWA / Installable App         | Uso offline + notificaciones push. Service Worker parcialmente configurado. | ✅ Resuelto |
| F3  | Chat Toggles per-conversation | Deshabilitar tools específicos per-chat (no solo web_search).               | ✅ Resuelto |
| F4  | Visualización de Tool Calls   | Mostrar tools en uso durante streaming como tarjetas expandibles.           | ✅ Resuelto |

### Media Prioridad

| #   | Función                    | Justificación                                                          | Referencia     |
| --- | -------------------------- | ---------------------------------------------------------------------- | -------------- |
| F5  | Human-in-the-Loop          | Nodos "Human Input" que pausan workflows para revisión humana.         | Dify 1.13      |
| F6  | Analytics Dashboard        | Métricas de uso, costos y tendencias.                                  | Open WebUI 0.8 |
| F7  | Prompt Templates / Library | Guardar y reutilizar prompts con versionado.                           | Open WebUI     |
| F8  | Model Compare Mode         | Comparar respuestas de diferentes modelos side-by-side.                | LobeChat       |
| F9  | File Upload en Chat        | Adjuntar archivos/imágenes al chat para modelos multimodales.          | Dify 1.13      |
| F10 | Code Sandbox               | Runtime Pyodide en browser. `pyodideRunner.js` ya existe parcialmente. | ROADMAP 7.2    |

### Baja Prioridad

| #   | Función                 | Justificación                               | Referencia          |
| --- | ----------------------- | ------------------------------------------- | ------------------- |
| F11 | Nested Sub-agents       | Sub-agents con profundidad configurable.    | OpenClaw v2026.2.15 |
| F12 | Plugin Ecosystem        | Sistema de plugins extensible.              | LobeChat            |
| F13 | Agent Collaboration     | Workflows de múltiples agentes colaborando. | Dify Roadmap        |
| F14 | Chat Export PDF         | Exportar conversaciones a PDF.              | LobeChat            |
| F15 | Visual Workflow Builder | Drag-and-drop tipo nodos (actual es lista). | ROADMAP 7.3         |

---

## 5. 📰 Análisis Competitivo (Feb 2026)

### PicoClaw (proyecto base)

- 12,000+ GitHub stars (Feb 16)
- Buscando maintainers de comunidad
- RAM aumentó a 10-20MB en updates recientes
- Advertencia: no producción antes de v1.0

### OpenClaw (v2026.2.15)

- 200,000+ GitHub stars
- Nested sub-agents con profundidad configurable
- Discord Components v2 (botones, selects, modals)
- Hooks `llm_input`/`llm_output` para plugins
- Seguridad: SHA-256 sandbox, redacción de tokens, sanitización de paths
- Fundador contratado por OpenAI

### Open WebUI (v0.8.0—0.8.3)

- Skills experimentales con inyección en chat
- Chat toggles para deshabilitar tools por conversación
- Analytics dashboard
- Access control UI rediseñada
- Prompt version control y tags

### Dify (v1.13—1.14rc)

- Human-in-the-Loop con nodo "Human Input"
- Multimodal nativo en Agent App
- Agent Skills con runtime sandboxed (beta)
- OpenTelemetry para observabilidad

### LobeChat

- Knowledge Base mejorada con pgvector
- Plugin ecosystem extensible
- Compare Mode entre modelos
- Rich text editor con math y task lists
- PDF export de conversaciones

---

## 6. 📝 Changelog de Correcciones y Mejoras

> Aquí se registra cada fix y feature que implementamos, en orden cronológico.

### 2026-02-20

| Tipo | ID  | Descripción                                                                                                                         | Archivos                                      | Estado        |
| ---- | --- | ----------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------- | ------------- |
| 📋   | —   | Revisión completa de la app y generación de este documento                                                                          | `docs/APP_REVIEW_AND_CHANGELOG.md`            | ✅ Completado |
| 🐛   | B1  | Fix: WebSocket task archive no hacía broadcast a otros clientes                                                                     | `server.go`                                   | ✅ Resuelto   |
| 🐛   | B2  | Fix: `downloadFile` y endpoints de exportación ahora incluyen `?token=` JWT por query string para resolver acceso Auth en descargas | `advancedService.js`, `server.go`             | ✅ Resuelto   |
| 🐛   | B3  | Fix: `defer` function encapsulada de los listeners sockets para prevenir que memory leak en sesiones extensas                       | `server.go`                                   | ✅ Resuelto   |
| 🐛   | B5  | Fix: Cache headers de `index.html` fallback ajustadas a `no-cache, no-store`                                                        | `server.go`                                   | ✅ Resuelto   |
| ✨   | M1  | Feature: Upload de archivos al workspace con Drag & Drop y selector                                                                 | `views/FilesView.vue`, `handlers_advanced.go` | ✅ Resuelto   |
| ✨   | F4  | Feature: Visualización de Tool Calls interactiva en streaming                                                                       | `ChatView.vue`, `chatStore.js`, `server.go`   | ✅ Resuelto   |
| ✨   | F3  | Feature: Gestión granular de tools por conversación (Toggles de herramientas AI)                                                    | `ChatView.vue`, `chatStore.js`, `server.go`   | ✅ Resuelto   |
| 🧹   | M7  | Refactor: Extracción de `MessageBubble` y `ToolCallItem` de `ChatView` para mejorar mantenibilidad                                  | `ChatView.vue`, `components/...`              | ✅ Resuelto   |
| ✨   | M10 | Feature: Streaming UX mejorada con estado de ejecución visible                                                                      | `ChatView.vue`, `ToolCallItem.vue`            | ✅ Resuelto   |
| ✨   | F1  | Feature: Sistema Multi-usuario (RBAC) con SQLite. Gestión de usuarios vía `SettingsView` (solo Admin)                               | `auth.go`, `handlers_users.go`, `storage/...` | ✅ Resuelto   |
| ✨   | F2  | Feature: Capacidades PWA, notificaciones push de tareas terminadas e interfaz "Install App" en `Sidebar`                            | `App.vue`, `taskStore.js`, `Sidebar.vue`      | ✅ Resuelto   |
| ✨   | M2  | Feature: Edición y visualización de chunks en Knowledge Base                                                                        | `KnowledgeView.vue`, `advancedService.js`     | ✅ Resuelto   |
| ✨   | M3  | Feature: Dashboard avanzado con gráficas de Chart.js y métricas de observabilidad                                                   | `DashboardView.vue`, `advancedService.js`     | ✅ Resuelto   |
| ✨   | M4  | Feature: Memory View con búsqueda live y timeline de notas diarias                                                                  | `MemoryView.vue`, `memoryService.js`          | ✅ Resuelto   |
| 🧹   | M8  | Refactor: Descomposición de `SettingsView` en componentes de pestañas (`Agent`, `Providers`, `Channels`)                            | `SettingsView.vue`, `components/Settings/...` | ✅ Resuelto   |
| ✨   | M10 | Feature: Visualización de herramientas (Tool Calls) mejorada durante el streaming                                                   | `ChatView.vue`, `ToolCallItem.vue`            | ✅ Resuelto   |

---

_Este documento se actualiza conforme avanzamos. Cada sesión de trabajo agrega entradas al changelog._
