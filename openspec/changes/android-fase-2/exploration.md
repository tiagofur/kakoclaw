# Exploration: Android Fase 2 - Feature Parity

## Current State

**Android Architecture:** La app sigue Clean Architecture con MVVM, modulación clara (app/, core/, feature/), y usa Jetpack Compose + Material3. La estructura está bien definida pero la mayoría de features de Fase 2 son esqueletales (40-60% completitud).

**Web App Reference:** La app web Vue tiene paridad completa en todas las features con UI rica (glass morphism, gradientes), componentes avanzados (visual workflow editor, gráficos, AI generation), y funcionalidad profunda.

## Affected Areas

**Features esqueletales (11):**
- `feature-agents/` — Lista y CRUD básicos, falta Swarm visualizer
- `feature-history/` — Lista de sesiones con search, falta filtros avanzados y exportación
- `feature-knowledge/` — Estructura creada, falta upload de archivos y RAG search UI
- `feature-memory/` — Estructura creada, falta timeline y tabs de long-term/daily
- `feature-skills/` — Tabs de installed/marketplace/generate, falta rating, categorías
- `feature-workflows/` — Lista básica, falta editor visual de nodos/connections
- `feature-metrics/` — Skeleton, falta gráficos de charts (LLM calls, tokens, cost)
- `feature-cron/` — CRUD básico, falta selector visual de horarios
- `feature-files/` — Lista de archivos, falta upload/preview/delete
- `feature-mcp/` — Lista de servidores, falta reconnect y configuración
- `feature-reports/` — Formulario básico, falta templates y exportación

**Core modules:**
- `core-database/` — DAOs para chat, tasks, agents (solo entidades básicas)
- `core-datastore/` — UserPreferences simple, falta persistencia de settings
- `core-network/` — APIs completas definidas (AgentApi, KnowledgeApi, SkillsApi, etc.)
- `core-security/` — Mínima implementación, falta secure storage para JWT

## Approaches

**Approach A: Parallel Implementation (Recommended)**
- Implementar features en paralelo por equipo
- Priorizar: agents → knowledge → workflows → metrics → cron → skills → memory → files → mcp → reports → history
- Reutilizar patrones de web: states, modales, CRUD flows
- **Pros:** Velocidad, consistencia, learning paralelo
- **Cons:** Requiere coordinación, riesgo de divergencia
- **Effort:** High (4 semanas estimado)

**Approach B: Sequential Implementation**
- Implementar feature por feature completamente antes de avanzar
- Orden por dependencia: knowledge (base) → agents → workflows → cron
- **Pros:** Menor riesgo, mejor calidad, pruebas integrales
- **Cons:** Más lento, no permite paralelismo
- **Effort:** High (6+ semanas)

**Approach C: Hybrid - Core First, UI Later**
- Implementar primero APIs + ViewModels de todos los features
- Luego enfocar en UI/UX en paralelo
- **Pros:** Backend listo, UI puede iterar rápido
- **Cons:** Riesgo de desalineación, UI sin feedback temprano
- **Effort:** Medium (5 semanas)

## APIs de Backend Necesarias

**APIs ya definidas (FeatureApis.kt):** ✅ Todas las APIs necesarias están definidas
- KnowledgeApi: listDocuments, upload, search, get, delete
- SkillsApi: list, install, uninstall, generate, marketplace
- CronApi: list, create, delete, toggle, AI generate
- FilesApi: browse, upload, delete
- MemoryApi: getLongTerm, updateLongTerm, getDailyNotes
- McpApi: list, get, reconnect
- WorkflowsApi: list, create, get, delete, run
- MetricsApi: get metrics
- ReportsApi: email report

**APIs faltantes (posibles):**
- History export API (exportar sesiones como JSON/Markdown)
- Knowledge chunk edit/update (ya existe, pero puede necesitar más)
- Files preview API (previsualización de archivos)

## Core Modules por Completar

**core-database (30% → 80%):**
- ✅ ChatSessionDao, ChatMessageDao, TaskDao
- ⚠️ Faltan DAOs para: KnowledgeDocument, Skill, Workflow, CronJob, FileEntry, Memory
- ⚠️ Faltan entities para todas las features adicionales
- ⚠️ Migrations no implementadas

**core-datastore (30% → 70%):**
- ✅ UserPreferences básico (serverUrl, theme, notifications)
- ⚠️ Faltan preferences para: feature settings, filters, UI preferences
- ⚠️ No hay implementación de caching de settings

**core-security (40% → 80%):**
- ⚠️ Falta secure storage para JWT (usar EncryptedSharedPreferences)
- ⚠️ Falta biometric auth
- ⚠️ Falta encryption de datos sensibles

**core-model (95%):**
- ✅ Models completos para: Agent, Chat, Task, Knowledge, Skill, Workflow, Cron, File, Memory, MCP
- ✅ Response types definidos

## Tecnologías/Librerías Recomendadas

**Workflows (Node/Graph Editor):**
- **Recomendado:** Implementar custom con Canvas API de Compose
- Alternativa: `composable-graph` (librería experimental)
- Patrones: Draggable nodes, curved connections, zoom/pan

**Metrics (Charts/Graphs):**
- **Recomendado:** `Vico` (Compose charts library)
- Alternativas: `Compose Charts`, `Victory Compose` wrapper
- Tipos: Line charts (LLM calls), Bar charts (tool usage), Donut charts (cost distribution)

**Cron (Schedule Selector):**
- **Recomendado:** Implementar custom selector (dial + lists)
- Patrones: Clock selector visual, preset schedules list, cron expression builder

**Knowledge (File Upload):**
- **Recomendado:** `Accompanist Permissions` + `ActivityResultContracts`
- Alternativa: `Coil` para preview de imágenes
- Features: Drag & drop, progress indicators, file type icons

**Markdown Rendering:**
- Ya usando `Markwon` en chat, reutilizar para knowledge/memory
- Alternativa: `Compose Markdown` (más lightweight)

## Matriz de Estado por Feature

| Feature | % Completitud | Qué existe | Qué falta |
|---------|---------------|------------|-----------|
| feature-agents | 60% | Lista specialists, CRUD básico, metrics, toggle orchestrator, generate with AI | Swarm visualizer, edit specialist, specialist details modal |
| feature-history | 60% | Lista sesiones, search básico, delete session | Filtros (date, agent, status), export (JSON/Markdown), archive, bulk actions |
| feature-knowledge | 50% | Screen+ViewModel structure, list API | File upload, document cards, RAG search UI, chunk editing, document preview |
| feature-memory | 50% | Structure, API calls | Long-term editor, daily notes timeline, search, tabs, auto-save |
| feature-skills | 50% | Tabs (installed/marketplace), install/uninstall, generate | Marketplace categories, skill rating, skill details, skill search, uninstall confirmation |
| feature-workflows | 40% | Lista workflows, run workflow, delete | **Editor visual** (nodes, connections), create/edit workflow, workflow templates, run logs |
| feature-metrics | 50% | Structure, API calls | **Gráficos** (line, bar, pie charts), time filters, agent breakdown, export CSV |
| feature-cron | 50% | Lista jobs, CRUD básico, toggle, AI generate | **Schedule selector** visual, job history, test run, cron expression editor |
| feature-files | 50% | Lista archivos, browse API | File upload, file preview (images, PDF, code), delete, download, file type icons |
| feature-mcp | 50% | Lista servers, reconnect | Server configuration, logs viewer, status indicators, server templates |
| feature-reports | 50% | Email report form | Report templates, schedule reports, report history, export (PDF, CSV) |

## Prioridades de Implementación

**ALTA Prioridad (Core features):**
1. **feature-knowledge** — Base para RAG, usado por chat
2. **feature-workflows** — Automation compleja, editor visual needed
3. **feature-metrics** — Observability crítica, gráficos needed
4. **feature-cron** — Scheduled tasks, selector visual needed

**MEDIA Prioridad:**
5. **feature-skills** — Marketplace, extensibilidad
6. **feature-agents** — Swarm visualizer para coordinación
7. **feature-memory** — Long-term context importante

**BAJA Prioridad:**
8. **feature-files** — File management, útil pero no crítico
9. **feature-mcp** — Protocol servers, avanzado
10. **feature-history** — Historial con filtros
11. **feature-reports** — Report generation, nice-to-have

## Riesgos y Blockers

**Técnicos:**
- 🚨 **Workflows visual editor** — Complejidad alta, requiere Canvas API custom
- 🚨 **Charts library** — Vico u otra puede tener bugs, requiere evaluación
- ⚠️ **Knowledge file upload** — Manejo de archivos grandes, progress indicators
- ⚠️ **Cron selector** — UX compleja para expresiones cron

**Dependencias:**
- ⚠️ **Backend APIs** — Requieren endpoints específicos (export, preview)
- ⚠️ **Core modules** — Database/DataStore necesitan implementación completa

**Integración:**
- ⚠️ **Paridad con web** — Mantener consistencia de UI/UX
- ⚠️ **Testing** — UI tests complejos para workflows/editor visual

**Estimación:**
- Total effort: **4-6 semanas** (2-3 developers)
- Riesgo medio: Workflows editor puede extender timeline
- Blockers mínimos: APIs ya definidas, libs disponibles

## Conclusiones

**Ready for Proposal:** ✅ Sí — El análisis completo está listo para crear la propuesta de implementación.

**Recommendations:**
1. Crear propuesta con Approach A (Parallel Implementation)
2. Priorizar features por criticidad
3. Definir tecnologías específicas para componentes complejos
4. Incluir milestone de core modules (database, datastore, security)

**Next steps:**
- Crear proposal con timeline y milestones
- Definir task breakdown por feature
- Especificar dependencies y blockers
- Incluir technical decisions para componentes complejos
