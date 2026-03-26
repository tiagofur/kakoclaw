# Android Fase 2 - Feature Parity Specification

## Purpose

Especificaciones técnicas completas para completar los 11 features esqueletales de la app Android MakoClaw, logrando paridad funcional con la app web.

## Requirements

### feature-knowledge

#### Requirement: Listar documentos knowledge base

The system MUST display a list of knowledge base documents with metadata (name, type, size, date).

- GIVEN user has knowledge documents
- WHEN user navigates to knowledge screen
- THEN system shows list of documents with metadata
- AND documents are sorted by date descending

#### Requirement: Subir documentos

The system MUST allow uploading documents (PDF, TXT, MD) with progress indication.

- GIVEN user is on knowledge screen
- WHEN user taps upload button and selects file
- THEN system shows upload progress
- AND upon completion, document appears in list

#### Requirement: Búsqueda RAG

The system MUST support Retrieval-Augmented Generation search across documents.

- GIVEN user has indexed documents
- WHEN user enters search query
- THEN system returns relevant chunks with source context
- AND results highlight matching content

#### Requirement: Previsualización de documentos

The system MUST provide document preview for supported formats (PDF, TXT, MD).

- GIVEN user selects a document
- WHEN user taps preview
- THEN system displays document content in modal

#### Requirement: Editar chunks de documentos

The system MUST allow editing individual document chunks.

- GIVEN user is previewing a document
- WHEN user edits a chunk
- THEN system saves changes on auto-save
- AND changes persist after app restart

#### Requirement: Eliminar documentos

The system MUST support document deletion with confirmation.

- GIVEN user selects a document
- WHEN user taps delete and confirms
- THEN system removes document from list
- AND document is deleted from backend

### feature-workflows

#### Requirement: Listar workflows

The system MUST display list of existing workflows with metadata (name, status, last run).

- GIVEN user has created workflows
- WHEN user navigates to workflows screen
- THEN system shows workflow cards with metadata

#### Requirement: Crear workflow con editor visual

The system MUST provide visual canvas editor for creating workflows with draggable nodes.

- GIVEN user taps create workflow
- WHEN user adds nodes to canvas
- THEN nodes are draggable and position persists
- AND nodes can be connected with curved edges

#### Requirement: Ejecutar workflow

The system MUST support workflow execution with real-time status updates.

- GIVEN user has valid workflow
- WHEN user taps run workflow
- THEN system executes workflow and shows progress
- AND results display in logs modal

#### Requirement: Validar workflows

The system MUST detect invalid workflows (cycles, disconnected nodes).

- GIVEN user creates workflow with disconnected nodes
- WHEN user attempts to run
- THEN system shows validation error
- AND highlights disconnected nodes

#### Requirement: Usar templates de workflow

The system MUST provide workflow templates for quick start.

- GIVEN user taps use template
- WHEN user selects a template
- THEN system loads template into editor
- AND user can modify template

### feature-metrics

#### Requirement: Mostrar gráficos de métricas

The system MUST display metric charts using Vico library (LLM calls, tokens, cost, tool usage).

- GIVEN user navigates to metrics screen
- WHEN system loads metrics data
- THEN system renders line/bar/donut charts
- AND charts are interactive (zoom, pan, tooltip)

#### Requirement: Filtrar por rango de tiempo

The system MUST support time range filtering (last hour, today, week, month, custom).

- GIVEN user selects time range
- WHEN user applies filter
- THEN system refreshes charts with filtered data

#### Requirement: Desglosar por agente

The system MUST allow filtering metrics by agent.

- GIVEN user has multiple agents
- WHEN user selects agent filter
- THEN system shows metrics breakdown by agent

#### Requirement: Exportar métricas

The system MUST support exporting metrics to CSV format.

- GIVEN user has metrics data
- WHEN user taps export
- THEN system downloads CSV file
- AND CSV contains all visible metrics

### feature-cron

#### Requirement: Listar cron jobs

The system MUST display list of cron jobs with schedule, status, next run.

- GIVEN user has created cron jobs
- WHEN user navigates to cron screen
- THEN system shows job cards with schedule

#### Requirement: Crear cron job con selector visual

The system MUST provide visual schedule selector with hour dial and day lists.

- GIVEN user taps create cron job
- WHEN user uses dial to select hour
- THEN hour selection updates cron expression preview
- AND days of week can be toggled

#### Requirement: Generar cron expression con AI

The system MUST support AI-assisted cron expression generation.

- GIVEN user taps generate with AI
- WHEN user describes schedule in natural language
- THEN system generates valid cron expression
- AND preview shows calculated times

#### Requirement: Toggle activar/desactivar cron job

The system MUST support enabling/disabling cron jobs.

- GIVEN user has cron job
- WHEN user toggles job status
- THEN system updates job status immediately
- AND backend reflects status change

#### Requirement: Test run de cron job

The system MUST support one-time test execution of cron jobs.

- GIVEN user selects cron job
- WHEN user taps test run
- THEN system executes job immediately
- AND results display in modal

### feature-skills

#### Requirement: Listar skills instaladas

The system MUST display installed skills with rating, category, description.

- GIVEN user has installed skills
- WHEN user navigates to skills tab
- THEN system shows installed skills cards

#### Requirement: Navegar marketplace de skills

The system MUST provide marketplace UI to browse available skills.

- GIVEN user taps marketplace tab
- WHEN system loads marketplace data
- THEN system shows available skills with install button

#### Requirement: Instalar skill

The system MUST support skill installation with confirmation.

- GIVEN user selects skill from marketplace
- WHEN user taps install
- THEN system shows installation progress
- AND skill moves to installed tab

#### Requirement: Desinstalar skill

The system MUST support skill uninstallation with confirmation dialog.

- GIVEN user has installed skill
- WHEN user taps uninstall and confirms
- THEN system removes skill from installed list

#### Requirement: Generar skill con AI

The system MUST support AI-assisted skill creation.

- GIVEN user taps generate skill
- WHEN user provides skill description
- THEN system generates skill code
- AND skill is installed automatically

#### Requirement: Calificar skill

The system MUST provide 5-star rating widget for skills.

- GIVEN user has used skill
- WHEN user submits rating
- THEN system updates skill rating
- AND rating persists across sessions

### feature-agents

#### Requirement: Listar especialistas (agents)

The system MUST display list of specialist agents with status, metrics, last activity.

- GIVEN user has specialist agents
- WHEN user navigates to agents screen
- THEN system shows agent cards with status indicators

#### Requirement: Crear especialista

The system MUST support creating new specialist agents with configuration form.

- GIVEN user taps create specialist
- WHEN user fills form and submits
- THEN system creates agent
- AND agent appears in list

#### Requirement: Editar especialista

The system MUST support editing specialist agent configuration.

- GIVEN user selects specialist
- WHEN user modifies configuration
- THEN system saves changes
- AND configuration updates persist

#### Requirement: Ver swarm visualizer

The system MUST provide swarm visualization showing agent coordination.

- GIVEN user taps swarm view
- WHEN system loads agent activity
- THEN system displays nodes and edges on canvas
- AND animations show active connections

#### Requirement: Ver métricas de especialista

The system MUST display specialist-specific metrics (LLM calls, tokens, response time).

- GIVEN user selects specialist
- WHEN user navigates to metrics view
- THEN system shows specialist metrics charts

#### Requirement: Toggle orchestrator

The system MUST support enabling/disabling orchestrator mode.

- GIVEN user toggles orchestrator
- WHEN system processes toggle
- THEN orchestrator status updates
- AND all agents reflect orchestrator state

#### Requirement: Generar especialista con AI

The system MUST support AI-assisted specialist creation.

- GIVEN user taps generate specialist
- WHEN user provides specialist description
- THEN system generates specialist configuration
- AND specialist is created automatically

### feature-memory

#### Requirement: Ver memoria a largo plazo

The system MUST display long-term memory content with edit capability.

- GIVEN user navigates to memory tab
- WHEN system loads memory
- THEN system shows memory content in editor

#### Requirement: Editar memoria a largo plazo

The system MUST support editing long-term memory with auto-save.

- GIVEN user modifies memory content
- WHEN system detects changes
- THEN system auto-saves every 30s
- AND changes persist after app restart

#### Requirement: Ver notas diarias

The system MUST display daily notes in timeline view.

- GIVEN user has daily notes
- WHEN user navigates to daily notes tab
- THEN system shows timeline with notes sorted by date

#### Requirement: Crear nota diaria

The system MUST support creating new daily notes.

- GIVEN user taps create note
- WHEN user enters note content
- THEN system creates note for current date
- AND note appears in timeline

#### Requirement: Buscar en memoria

The system MUST support full-text search across memory and notes.

- GIVEN user enters search query
- WHEN system processes search
- THEN system shows matching results
- AND results highlight search terms

#### Requirement: Configurar retención de memoria

The system MUST allow configuring memory retention period.

- GIVEN user opens retention settings
- WHEN user selects retention period
- THEN system updates retention configuration
- AND old memories are automatically cleaned

### feature-files

#### Requirement: Navegar estructura de archivos

The system MUST provide file browser with folder navigation.

- GIVEN user navigates to files screen
- WHEN system loads directory structure
- THEN system shows folders and files
- AND navigation breadcrumbs display path

#### Requirement: Subir archivos

The system MUST support file upload with progress indication and permissions.

- GIVEN user taps upload button
- WHEN user selects file and grants permissions
- THEN system shows upload progress
- AND file appears in current directory

#### Requirement: Descargar archivos

The system MUST support file download to device storage.

- GIVEN user selects file
- WHEN user taps download
- THEN system downloads file
- AND file saves to device downloads

#### Requirement: Eliminar archivos

The system MUST support file and folder deletion with confirmation.

- GIVEN user selects file/folder
- WHEN user taps delete and confirms
- THEN system removes item from list
- AND item is deleted from backend

#### Requirement: Previsualizar archivos

The system MUST provide preview for images, PDF, and code files.

- GIVEN user selects supported file
- WHEN user taps preview
- THEN system displays file content in modal

#### Requirement: Crear carpetas

The system MUST support folder creation with name input.

- GIVEN user taps create folder
- WHEN user enters folder name
- THEN system creates folder in current directory

#### Requirement: Renombrar archivos/carpetas

The system MUST support renaming files and folders.

- GIVEN user selects item
- WHEN user enters new name
- THEN system updates item name
- AND change persists after app restart

### feature-mcp

#### Requirement: Listar servidores MCP

The system MUST display list of MCP servers with connection status.

- GIVEN user has configured MCP servers
- WHEN user navigates to MCP screen
- THEN system shows server cards with status indicators

#### Requirement: Ver detalles de servidor MCP

The system MUST display server details (endpoint, tools, config).

- GIVEN user selects MCP server
- WHEN user taps details
- THEN system shows server configuration
- AND available tools list is displayed

#### Requirement: Configurar servidor MCP

The system MUST support MCP server configuration form.

- GIVEN user taps configure server
- WHEN user saves configuration
- THEN system updates server settings
- AND server attempts connection

#### Requirement: Habilitar/deshabilitar servidor MCP

The system MUST support enabling/disabling MCP servers.

- GIVEN user toggles server status
- WHEN system processes toggle
- THEN server status updates immediately
- AND connected servers show active indicator

#### Requirement: Ver logs de servidor MCP

The system MUST display server logs for debugging.

- GIVEN user selects server
- WHEN user taps view logs
- THEN system shows connection logs
- AND logs update in real-time

#### Requirement: Reconnectar servidor MCP

The system MUST support manual reconnect attempt.

- GIVEN server is disconnected
- WHEN user taps reconnect
- THEN system attempts connection
- AND status indicator updates

### feature-history

#### Requirement: Listar sesiones de chat

The system MUST display chat history with metadata (date, agent, message count).

- GIVEN user has chat sessions
- WHEN user navigates to history screen
- THEN system shows session cards with metadata

#### Requirement: Buscar sesiones

The system MUST support searching sessions by content, agent, or date.

- GIVEN user enters search query
- WHEN system processes search
- THEN system shows matching sessions
- AND results highlight search terms

#### Requirement: Filtrar sesiones

The system MUST support filtering by date range, agent, and status.

- GIVEN user applies filters
- WHEN system processes filters
- THEN session list updates
- AND active filters are displayed

#### Requirement: Ver detalles de sesión

The system MUST display full session messages in modal.

- GIVEN user selects session
- WHEN user taps view details
- THEN system shows all messages
- AND messages are formatted correctly

#### Requirement: Eliminar sesión

The system MUST support session deletion with confirmation.

- GIVEN user selects session
- WHEN user taps delete and confirms
- THEN system removes session from list

#### Requirement: Archivar sesión

The system MUST support session archiving.

- GIVEN user selects session
- WHEN user taps archive
- THEN system moves session to archive
- AND session no longer appears in main list

#### Requirement: Exportar sesión

The system MUST support session export to JSON or Markdown.

- GIVEN user selects session
- WHEN user taps export and selects format
- THEN system downloads file in chosen format

#### Requirement: Acciones bulk

The system MUST support bulk actions (delete/archive multiple sessions).

- GIVEN user selects multiple sessions
- WHEN user taps bulk action
- THEN system applies action to all selected

### feature-reports

#### Requirement: Generar reporte por email

The system MUST support generating and emailing reports.

- GIVEN user selects report type
- WHEN user configures filters and sends
- THEN system generates report
- AND report is emailed to user

#### Requirement: Seleccionar template de reporte

The system MUST provide report template selection.

- GIVEN user taps generate report
- WHEN system loads templates
- THEN user can select from available templates

#### Requirement: Configurar filtros de reporte

The system MUST support report filters (date range, agent, type).

- GIVEN user creates report
- WHEN user applies filters
- THEN system shows filtered data in preview

#### Requirement: Programar reporte recurrente

The system MUST support scheduling recurring reports.

- GIVEN user taps schedule report
- WHEN user configures frequency
- THEN system creates schedule
- AND report generates automatically

#### Requirement: Ver historial de reportes

The system MUST display history of generated reports.

- GIVEN user has generated reports
- WHEN user navigates to report history
- THEN system shows report list with status

#### Requirement: Cancelar reporte programado

The system MUST support cancelling scheduled reports.

- GIVEN user selects scheduled report
- WHEN user taps cancel
- THEN system removes schedule
- AND report stops generating

#### Requirement: Exportar reporte

The system MUST support exporting reports to PDF or CSV.

- GIVEN user has generated report
- WHEN user taps export
- THEN system downloads file in chosen format

## Common Requirements

### Requirement: Dark/Light theme support

The system MUST support both dark and light themes with consistent styling across all features.

- GIVEN user changes theme setting
- WHEN system applies theme
- THEN all screens update immediately
- AND theme persists after app restart

### Requirement: Loading states consistentes

The system MUST show consistent loading indicators (skeleton, progress) across all features.

- GIVEN user navigates to any feature
- WHEN data is loading
- THEN system shows loading indicator
- AND indicator is consistent across app

### Requirement: Empty states con CTAs claros

The system MUST display empty states with clear call-to-action buttons.

- GIVEN feature has no data
- WHEN user views empty screen
- THEN system shows empty state message
- AND CTA button is prominent and clear

### Requirement: Error states con vías de recuperación

The system MUST display error messages with retry or recovery actions.

- GIVEN API request fails
- WHEN system detects error
- THEN system shows error message
- AND retry or alternative action is available

### Requirement: Responsive design

The system MUST support mobile and tablet screen sizes with adaptive layouts.

- GIVEN user changes device orientation
- WHEN system adjusts layout
- THEN UI adapts to new size
- AND all functionality remains accessible

### Requirement: Accessibility (TalkBack)

The system MUST support TalkBack with content descriptions for all interactive elements.

- GIVEN user enables TalkBack
- WHEN user navigates app
- THEN all elements have content descriptions
- AND focus order is logical

### Requirement: Animaciones fluidas (60fps)

The system MUST maintain 60fps animations across all transitions and interactions.

- GIVEN user performs action
- WHEN animation plays
- THEN animation runs at 60fps
- AND no jank or stuttering occurs

## Backend APIs

| Feature | Endpoint | Method | Description |
|---------|----------|--------|-------------|
| knowledge | /knowledge/documents | GET | Listar documentos |
| knowledge | /knowledge/upload | POST | Subir documento |
| knowledge | /knowledge/document/{id} | GET | Obtener documento |
| knowledge | /knowledge/document/{id} | DELETE | Eliminar documento |
| knowledge | /knowledge/search | POST | Buscar documentos (RAG) |
| knowledge | /knowledge/document/{id}/chunk | PATCH | Editar chunk |
| workflows | /workflows | GET | Listar workflows |
| workflows | /workflows/create | POST | Crear workflow |
| workflows | /workflow/{id} | GET | Obtener workflow |
| workflows | /workflow/{id} | DELETE | Eliminar workflow |
| workflows | /workflow/{id}/run | POST | Ejecutar workflow |
| workflows | /workflow/{id}/logs | GET | Ver logs de ejecución |
| metrics | /metrics | GET | Obtener métricas |
| metrics | /metrics/export | GET | Exportar métricas (CSV) |
| cron | /cron/jobs | GET | Listar cron jobs |
| cron | /cron/create | POST | Crear cron job |
| cron | /cron/{id} | PATCH | Editar cron job |
| cron | /cron/{id} | DELETE | Eliminar cron job |
| cron | /cron/{id}/toggle | POST | Toggle activar/desactivar |
| cron | /cron/generate | POST | Generar cron expression con AI |
| cron | /cron/{id}/test | POST | Test run de cron job |
| skills | /skills/installed | GET | Listar skills instaladas |
| skills | /skills/marketplace | GET | Navegar marketplace |
| skills | /skills/install | POST | Instalar skill |
| skills | /skills/uninstall | POST | Desinstalar skill |
| skills | /skills/generate | POST | Generar skill con AI |
| agents | /agents/specialists | GET | Listar especialistas |
| agents | /agents/specialist/create | POST | Crear especialista |
| agents | /agents/specialist/{id} | PATCH | Editar especialista |
| agents | /agents/specialist/{id} | DELETE | Eliminar especialista |
| agents | /agents/specialist/generate | POST | Generar especialista con AI |
| agents | /agents/specialist/{id}/metrics | GET | Métricas de especialista |
| agents | /agents/specialist/{id}/logs | GET | Logs de especialista |
| agents | /agents/orchestrator/toggle | POST | Toggle orchestrator |
| memory | /memory/long-term | GET | Obtener memoria a largo plazo |
| memory | /memory/long-term | PATCH | Editar memoria a largo plazo |
| memory | /memory/daily-notes | GET | Listar notas diarias |
| memory | /memory/daily-note | POST | Crear nota diaria |
| memory | /memory/daily-note/{id} | PATCH | Editar nota diaria |
| memory | /memory/search | GET | Buscar en memoria |
| memory | /memory/retention-settings | PATCH | Configurar retención |
| files | /files/browse | GET | Navegar archivos |
| files | /files/upload | POST | Subir archivo |
| files | /files/download/{id} | GET | Descargar archivo |
| files | /files/{id} | DELETE | Eliminar archivo |
| files | /files/folder | POST | Crear carpeta |
| files | /files/{id}/rename | PATCH | Renombrar archivo |
| mcp | /mcp/servers | GET | Listar servidores MCP |
| mcp | /mcp/server/{id} | GET | Ver detalles de servidor |
| mcp | /mcp/server/configure | POST | Configurar servidor |
| mcp | /mcp/server/{id}/toggle | POST | Toggle habilitar/deshabilitar |
| mcp | /mcp/server/{id}/logs | GET | Ver logs del servidor |
| mcp | /mcp/server/{id}/reconnect | POST | Reconnectar servidor |
| history | /chat/sessions | GET | Listar sesiones |
| history | /chat/session/{id} | GET | Ver detalles de sesión |
| history | /chat/session/{id} | DELETE | Eliminar sesión |
| history | /chat/session/{id}/archive | POST | Archivar sesión |
| history | /chat/search | GET | Buscar sesiones |
| history | /chat/session/{id}/export | POST | Exportar sesión |
| reports | /reports/email | POST | Generar reporte por email |
| reports | /reports/templates | GET | Listar templates |
| reports | /reports/schedule | POST | Programar reporte |
| reports | /reports/history | GET | Ver historial de reportes |
| reports | /reports/schedule/{id} | DELETE | Cancelar reporte |
| reports | /reports/export | POST | Exportar reporte (PDF/CSV) |

## UI States

### Loading State

The system MUST show skeleton loader or progress indicator during data fetching.

- GIVEN user navigates to any feature
- WHEN data is being loaded
- THEN system shows loading indicator
- AND user cannot interact with UI

### Success State

The system MUST display loaded content in interactive state.

- GIVEN data has loaded successfully
- WHEN screen renders
- THEN content is visible and interactive
- AND all actions are available

### Empty State

The system MUST show empty state message with CTA button when no data exists.

- GIVEN feature has no data
- WHEN user views screen
- THEN empty state message is displayed
- AND CTA button encourages user action

### Error State

The system MUST display error message with recovery action.

- GIVEN error occurs (API, network, etc.)
- WHEN system detects error
- THEN error message is shown
- AND retry or alternative action is available

### Editing State

The system MUST show form or editor in active editing mode.

- GIVEN user initiates edit action
- WHEN editor/form opens
- THEN editing UI is displayed
- AND save/cancel actions are available

### Saving State

The system MUST show save progress indicator during data persistence.

- GIVEN user initiates save action
- WHEN data is being saved
- THEN save indicator is shown
- AND interruption is prevented

## Shared UI Components

### EmptyState

The system MUST provide reusable empty state component with icon, message, and CTA.

- GIVEN component is used
- WHEN screen has no data
- THEN component displays with provided icon, message, and CTA

### LoadingScreen

The system MUST provide reusable loading screen component with indicator.

- GIVEN component is used
- WHEN loading occurs
- THEN component shows loading indicator
- AND style is consistent across app

### ErrorScreen

The system MUST provide reusable error screen component with message and retry button.

- GIVEN component is used
- WHEN error occurs
- THEN component shows error message
- AND retry button triggers refresh action

### SearchBar

The system MUST provide reusable search bar with debounced input.

- GIVEN component is used
- WHEN user types search query
- THEN search triggers after 300ms debounce
- AND clear button is available

### FilterSheet

The system MUST provide reusable bottom sheet for filters (date, type, agent).

- GIVEN component is used
- WHEN user opens filters
- THEN sheet slides from bottom
- AND user can select multiple filters

### ConfirmDialog

The system MUST provide reusable confirmation dialog for destructive actions.

- GIVEN component is used
- WHEN destructive action is triggered
- THEN dialog shows confirmation message
- AND action only proceeds on confirm

### ProgressIndicator

The system MUST provide reusable progress indicator for async operations.

- GIVEN component is used
- WHEN async operation runs
- THEN indicator shows percentage or spinner
- AND operation can be cancelled if supported

## General Acceptance Criteria

### Functional Requirements

- [ ] All 11 features implemented and tested
- [ ] CRUD operations work correctly for all features
- [ ] All backend API endpoints integrated
- [ ] Error handling covers all edge cases
- [ ] Auto-save works where applicable (memory, notes)

### UI/UX Requirements

- [ ] All screens follow Material Design 3 guidelines
- [ ] All states (loading, success, empty, error) implemented
- [ ] Responsive design works on mobile and tablet
- [ ] Dark/Light theme supported across all features
- [ ] Animations run at 60fps without jank
- [ ] Transitions are smooth and consistent

### Accessibility Requirements

- [ ] All interactive elements have content descriptions
- [ ] TalkBack support verified for all screens
- [ ] Focus order is logical and predictable
- [ ] Touch targets meet minimum size (48dp)

### Performance Requirements

- [ ] No performance regressions vs Fase 1 baseline
- [ ] Lists use lazy loading/virtualization
- [ ] Charts render efficiently (Vico optimization)
- [ ] Image loading uses Coil with caching

### Testing Requirements

- [ ] Unit tests written for all features (target: >70% coverage)
- [ ] UI tests cover happy paths for all features
- [ ] Edge cases tested (upload failures, empty states, API errors)
- [ ] Integration tests verify API communication

### Documentation Requirements

- [ ] Complex components documented (workflows editor, cron selector)
- [ ] API integration documented
- [ ] Code comments added for complex logic
- [ ] Component usage examples provided

## Edge Cases

### Network Errors

- GIVEN network connection is lost
- WHEN API request is made
- THEN system shows error message
- AND retry button is available

### Large File Upload

- GIVEN user uploads file >10MB
- WHEN upload starts
- THEN system shows progress indicator
- AND upload completes successfully

### Empty Results

- GIVEN search returns no results
- WHEN results display
- THEN empty state message is shown
- AND CTA button encourages different search

### Concurrent Operations

- GIVEN user performs multiple actions rapidly
- WHEN operations execute
- THEN system handles conflicts gracefully
- AND data remains consistent

### Memory Limit

- GIVEN app uses significant memory
- WHEN memory pressure increases
- THEN system frees unused resources
- AND app does not crash
