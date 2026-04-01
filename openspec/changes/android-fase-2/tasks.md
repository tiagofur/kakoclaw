# Tasks: Android Fase 2 - Feature Parity

## Overview

Este documento contiene el breakdown detallado de tareas para implementar los 11 features esqueletales de la Fase 2 Android, completar los core modules y lograr paridad funcional con la app web.

## Task Breakdown por Milestone

### Milestone 1: Core + Prioridad ALTA (Part 1)

#### Core Modules (2 días)

##### TASK-CORE-001: Completar core-database
**Description**: Agregar DAOs y entities para los nuevos features
**Estimate**: 4 horas
**Priority**: CRITICAL
**Dependencies**: Ninguna

**Subtasks**:
- [x] Crear entity `KnowledgeDocumentEntity` con todos los campos
- [x] Crear entity `SkillEntity` con todos los campos
- [x] Crear entity `WorkflowEntity` con todos los campos
- [x] Crear entity `CronJobEntity` con todos los campos
- [x] Crear entity `FileEntryEntity` con todos los campos
- [x] Crear entity `MemoryEntity` con todos los campos
- [x] Crear entity `McpServerEntity` con todos los campos
- [x] Crear entity `SessionEntity` con todos los campos
- [x] Crear entity `ReportEntity` con todos los campos
- [x] Crear DAOs para todas las entities (@Dao interfaces)
- [x] Agregar migrations para nuevas tablas

**Acceptance Criteria**:
- [x] Todas las entities tienen anotaciones @Entity y @PrimaryKey
- [x] Todos los DAOs tienen métodos CRUD básicos
- [x] Migrations compilaron sin errores

**Definition of Done**:
- Entity + DAO creados para 9 nuevas entidades
- Migrations definidas
- Unit tests pasan

---

##### TASK-CORE-002: Completar core-datastore
**Description**: Implementar persistencia de settings y preferences de features
**Estimate**: 3 horas
**Priority**: CRITICAL
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [x] Definir `FeaturePreferences` data class con todos los settings
- [x] Crear `PreferencesStore` con DataStore
- [x] Implementar keys para knowledge (upload size, chunk size)
- [x] Implementar keys para metrics (time range, agent filter)
- [x] Implementar keys para cron (notify, auto-test)
- [x] Implementar keys para skills (auto-install, category filters)
- [x] Implementar keys para memory (retention days, auto-save interval)
- [x] Implementar keys para files (default path, sort order)
- [x] Implementar keys para mcp (reconnect timeout)
- [x] Implementar keys para history (export format, archive auto)
- [x] Implementar keys para reports (templates, schedule)

**Acceptance Criteria**:
- [x] Todos los preferences definidos
- [x] PreferencesStore expone Flow<FeaturePreferences>
- [x] updatePreferences() persiste cambios
- [x] Unit tests pasan

**Definition of Done**:
- PreferencesStore completa
- Todos los settings implementados
- Tests coverage >80%

---

##### TASK-CORE-003: Completar core-security
**Description**: Implementar secure storage para JWT y tokens sensibles
**Estimate**: 3 horas
**Priority**: CRITICAL
**Dependencies**: Ninguna

**Subtasks**:
- [x] Crear `JwtStorage` con EncryptedSharedPreferences
- [x] Implementar saveJwt(token: String)
- [x] Implementar getJwt(): String?
- [x] Implementar clearJwt()
- [x] Agregar token refresh logic
- [ ] Implementar biometric auth (opcional)
- [x] Agregar encryption para datos sensibles (API keys)
- [x] Implementar token validation
- [x] Agregar interceptor de auth a Retrofit
- [x] Tests unitarios pasan

**Acceptance Criteria**:
- [x] JWT se guarda en EncryptedSharedPreferences
- [x] getJwt() retorna token o null
- [x] clearJwt() elimina token
- [x] Token refresh funciona
- [ ] Biometric auth implementado (opcional)
- [x] Tests coverage >80%

**Definition of Done**:
- JwtStorage completa
- EncryptedSharedPreferences implementado
- Auth interceptor en Retrofit
- Tests pasan

---

#### feature-knowledge (3 días)

##### TASK-KNOWLEDGE-001: Crear KnowledgeViewModel
**Description**: Implementar ViewModel con StateFlow y eventos
**Estimate**: 4 horas
**Priority**: HIGH
**Dependencies**: TASK-CORE-001, TASK-CORE-002

**Subtasks**:
- [x] Crear `KnowledgeUiState` data class
- [x] Crear `KnowledgeEvent` sealed class
- [x] Crear `KnowledgeEffect` sealed class
- [x] Crear `KnowledgeViewModel` con @HiltViewModel
- [x] Implementar _uiState MutableStateFlow
- [x] Implementar _effects MutableSharedFlow
- [x] Implementar event handler onEvent()
- [x] Implementar loadDocuments()
- [x] Implementar uploadDocument()
- [x] Implementar searchDocuments()
- [x] Implementar deleteDocument()
- [x] Implementar editChunk()
- [x] Tests unitarios pasan

**Acceptance Criteria**:
- [x] ViewModel compila sin errores
- [x] StateFlow expone KnowledgeUiState
- [x] Eventos manejan correctamente
- [x] Efectos emiten correctamente
- [x] Tests coverage >70%

**Definition of Done**:
- KnowledgeViewModel completa
- Todos los eventos implementados
- Tests pasan

---

##### TASK-KNOWLEDGE-002: Crear KnowledgeRepository
**Description**: Implementar Repository con API y Room DAO
**Estimate**: 4 horas
**Priority**: HIGH
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [x] Crear `KnowledgeRepository` interface
- [x] Crear `KnowledgeRepositoryImpl` con @Inject
- [x] Implementar getDocuments(): Flow<List<KnowledgeDocument>>
- [x] Implementar uploadDocument(file: File): Flow<UploadProgress>
- [x] Implementar searchDocuments(query: String): Flow<List<KnowledgeDocument>>
- [x] Implementar deleteDocument(id: String): Flow<Unit>
- [x] Implementar editChunk(...): Flow<Unit>
- [x] Crear KnowledgeDocumentDao
- [x] Implementar caching en Room
- [x] Integrar con KnowledgeApi
- [x] Tests unitarios pasan

**Acceptance Criteria**:
- [x] Repository implementa interface
- [x] getDocuments() usa cache local
- [x] uploadDocument() muestra progreso
- [x] searchDocuments() llama API
- [x] deleteDocument() llama API y cache
- [x] Tests coverage >70%

**Definition of Done**:
- KnowledgeRepository completa
- Caching implementado
- API integrada
- Tests pasan

---

##### TASK-KNOWLEDGE-003: Crear KnowledgeScreen UI
**Description**: Implementar screen principal con listado de documentos
**Estimate**: 6 horas
**Priority**: HIGH
**Dependencies**: TASK-KNOWLEDGE-001

**Subtasks**:
- [x] Crear `KnowledgeScreen` @Composable
- [x] Implementar loading state con LoadingScreen
- [x] Implementar success state con LazyColumn de documentos
- [x] Implementar empty state con EmptyState
- [x] Implementar error state con ErrorScreen
- [x] Implementar SearchBar
- [x] Implementar UploadFileButton
- [x] Implementar navegación a detalles
- [x] Implementar pull-to-refresh
- [x] Tests UI pasan

**Acceptance Criteria**:
- [x] Screen compila sin errores
- [x] Loading state funciona
- [x] Success state muestra documentos
- [x] Empty state muestra mensaje
- [x] SearchBar filtra documentos
- [x] UploadFileButton abre file picker
- [x] Pull-to-refresh recarga
- [x] Tests UI pasan

**Definition of Done**:
- KnowledgeScreen completa
- Todos los estados implementados
- Tests UI pasan

---

##### TASK-KNOWLEDGE-004: Crear KnowledgeDocumentCard
**Description**: Crear card reutilizable para documento
**Estimate**: 3 horas
**Priority**: HIGH
**Dependencies**: TASK-KNOWLEDGE-001

**Subtasks**:
- [x] Crear `KnowledgeDocumentCard` @Composable
- [x] Implementar diseño con Card de Material3
- [x] Mostrar título del documento
- [x] Mostrar chunk count
- [x] Mostrar fecha de creación
- [x] Implementar onClick
- [x] Implementar onLongClick (menú contextual)
- [x] Implementar icono de tipo de archivo
- [x] Tests UI pasan

**Acceptance Criteria**:
- [x] Card muestra datos correctamente
- [x] onClick funciona
- [x] onLongClick muestra menú
- [x] Iconos correctos por tipo
- [x] Tests UI pasan

**Definition of Done**:
- KnowledgeDocumentCard completa
- Tests UI pasan

---

##### TASK-KNOWLEDGE-005: Crear UploadFileButton y DocumentPreviewModal
**Description**: Crear botón de upload y modal de preview
**Estimate**: 4 horas
**Priority**: HIGH
**Dependencies**: TASK-KNOWLEDGE-002

**Subtasks**:
- [x] Crear `UploadFileButton` @Composable
- [x] Integrar ActivityResultContracts.GetContent()
- [x] Mostrar progreso de upload
- [x] Validar tipo de archivo
- [x] Validar tamaño máximo (10MB)
- [x] Crear `DocumentPreviewModal` @Composable
- [x] Integrar Markwon para preview de MD/TXT
- [x] Implementar preview de PDF (usar librería externa)
- [x] Implementar botón de cerrar
- [x] Tests UI pasan

**Acceptance Criteria**:
- [x] Upload abre file picker
- [x] Progreso visible
- [x] Validación funciona
- [x] Preview funciona para MD/TXT
- [x] Preview funciona para PDF
- [x] Modal se cierra
- [x] Tests UI pasan

**Definition of Done**:
- UploadFileButton completa
- DocumentPreviewModal completa
- Tests UI pasan

---

##### TASK-KNOWLEDGE-006: Crear ChunkEditorSheet
**Description**: Crear sheet para editar chunks de documentos
**Estimate**: 3 horas
**Priority**: MEDIUM
**Dependencies**: TASK-KNOWLEDGE-001

**Subtasks**:
- [x] Crear `ChunkEditorSheet` @Composable
- [x] Implementar BottomSheetScaffold
- [x] Mostrar lista de chunks
- [x] Implementar edit de chunk (OutlinedTextField)
- [x] Implementar save button
- [x] Implementar cancel button
- [x] Implementar auto-save
- [x] Mostrar indicador de guardado
- [x] Tests UI pasan

**Acceptance Criteria**:
- [x] Sheet muestra chunks
- [x] Edit funciona
- [x] Save persiste cambios
- [x] Auto-save funciona
- [x] Indicador visible
- [x] Tests UI pasan

**Definition of Done**:
- ChunkEditorSheet completa
- Tests UI pasan

---

### Milestone 2: Prioridad ALTA (Part 2)

#### feature-workflows (5 días) - 🚨 HIGH COMPLEXITY

##### TASK-WORKFLOWS-001: Crear WorkflowsViewModel
**Description**: Implementar ViewModel para gestión de workflows
**Estimate**: 4 horas
**Priority**: CRITICAL
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [x] Crear `WorkflowsUiState` con lista, editor state, logs
- [x] Crear `WorkflowsEvent` sealed class
- [x] Crear `WorkflowsEffect` sealed class
- [x] Crear `WorkflowsViewModel` con @HiltViewModel
- [x] Implementar loadWorkflows()
- [x] Implementar createWorkflow()
- [x] Implementar openEditor()
- [x] Implementar executeWorkflow()
- [x] Implementar deleteWorkflow()
- [ ] Tests unitarios pasan

**Acceptance Criteria**:
- [x] ViewModel compila
- [x] StateFlow expone WorkflowsUiState
- [x] Eventos manejan correctamente
- [ ] Tests coverage >70%

**Definition of Done**:
- WorkflowsViewModel completa
- Tests pasan

**Nota**: La implementación de la micro-delegación quedó completa. La ejecución de tests quedó pendiente por instrucción explícita de no correrlos en esta delegación.

---

##### TASK-WORKFLOWS-002: Crear WorkflowsRepository
**Description**: Implementar Repository con API y Room
**Estimate**: 4 horas
**Priority**: CRITICAL
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [ ] Crear `WorkflowsRepository` interface
- [ ] Crear `WorkflowsRepositoryImpl` con @Inject
- [ ] Implementar getWorkflows(): Flow<List<Workflow>>
- [ ] Implementar createWorkflow(): Flow<Unit>
- [ ] Implementar updateWorkflow(): Flow<Unit>
- [ ] Implementar deleteWorkflow(): Flow<Unit>
- [ ] Implementar executeWorkflow(): Flow<ExecutionLog>
- [ ] Crear WorkflowDao
- [ ] Implementar caching
- [ ] Integrar WorkflowsApi
- [ ] Tests unitarios pasan

**Acceptance Criteria**:
- [ ] Repository funciona
- [ ] Caching implementado
- [ ] API integrada
- [ ] Tests coverage >70%

**Definition of Done**:
- WorkflowsRepository completa
- Tests pasan

---

##### TASK-WORKFLOWS-003: Crear WorkflowCanvasEditor 🚨
**Description**: Implementar canvas editor con nodes y edges (HIGH COMPLEXITY)
**Estimate**: 16 horas (2 días)
**Priority**: CRITICAL
**Dependencies**: TASK-WORKFLOWS-001

**Subtasks**:
- [ ] Crear `WorkflowCanvasEditor` @Composable
- [ ] Implementar Canvas con Canvas API
- [ ] Implementar zoom/pan
- [ ] Crear `WorkflowNode` @Composable
- [ ] Implementar drag de nodos
- [ ] Crear `WorkflowEdge` @Composable
- [ ] Implementar dibujo de edges (curvas Bezier)
- [ ] Implementar conexión de nodos
- [ ] Implementar selección de nodos
- [ ] Implementar eliminación de nodos/edges
- [ ] Implementar add node button
- [ ] Implementar save button
- [ ] Implementar cancel button
- [ ] Tests UI pasan

**Acceptance Criteria**:
- [ ] Canvas funciona
- [ ] Nodes se arrastran
- [ ] Edges se dibujan
- [ ] Conexión funciona
- [ ] Selección funciona
- [ ] Eliminación funciona
- [ ] Zoom/pan funciona
- [ ] Tests UI pasan

**Definition of Done**:
- WorkflowCanvasEditor completa
- Todos los gestos funcionan
- Tests UI pasan

---

##### TASK-WORKFLOWS-004: Crear WorkflowsScreen UI
**Description**: Implementar screen principal de workflows
**Estimate**: 6 horas
**Priority**: HIGH
**Dependencies**: TASK-WORKFLOWS-001

**Subtasks**:
- [ ] Crear `WorkflowsScreen` @Composable
- [ ] Implementar loading state
- [ ] Implementar success state con LazyColumn
- [ ] Implementar empty state
- [ ] Implementar error state
- [ ] Implementar WorkflowCard
- [ ] Implementar FAB para crear workflow
- [ ] Implementar navegación al editor
- [ ] Implementar ejecución de workflow
- [ ] Tests UI pasan

**Acceptance Criteria**:
- [ ] Screen funciona
- [ ] Cards muestran workflows
- [ ] FAB abre editor
- [ ] Ejecución funciona
- [ ] Tests UI pasan

**Definition of Done**:
- WorkflowsScreen completa
- Tests UI pasan

---

##### TASK-WORKFLOWS-005: Crear WorkflowCard y ExecutionLogsModal
**Description**: Crear card de workflow y modal de logs
**Estimate**: 4 horas
**Priority**: HIGH
**Dependencies**: TASK-WORKFLOWS-001

**Subtasks**:
- [ ] Crear `WorkflowCard` @Composable
- [ ] Mostrar nombre de workflow
- [ ] Mostrar última ejecución
- [ ] Mostrar estado
- [ ] Implementar onClick (ejecutar)
- [ ] Implementar onLongClick (menú)
- [ ] Crear `ExecutionLogsModal` @Composable
- [ ] Mostrar lista de logs
- [ ] Implementar scroll
- [ ] Implementar cerrar
- [ ] Tests UI pasan

**Acceptance Criteria**:
- [ ] Card muestra datos
- [ ] Ejecución funciona
- [ ] Modal muestra logs
- [ ] Tests UI pasan

**Definition of Done**:
- WorkflowCard completa
- ExecutionLogsModal completa
- Tests UI pasan

---

#### feature-metrics (4 días)

##### TASK-METRICS-001: Crear MetricsViewModel
**Description**: Implementar ViewModel para métricas
**Estimate**: 3 horas
**Priority**: HIGH
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [x] Crear `MetricsUiState`
- [x] Crear `MetricsEvent`
- [x] Crear `MetricsEffect`
- [x] Crear `MetricsViewModel`
- [x] Implementar loadMetrics()
- [x] Implementar setTimeRange()
- [x] Implementar setAgentFilter()
- [x] Implementar exportMetrics()
- [x] Tests unitarios pasan

**Acceptance Criteria**:
- [x] ViewModel funciona
- [x] Filtros funcionan
- [x] Export funciona
- [x] Tests coverage >70%

**Definition of Done**:
- MetricsViewModel completa
- Tests pasan

---

##### TASK-METRICS-002: Crear MetricsRepository
**Description**: Implementar Repository con API y cache
**Estimate**: 3 horas
**Priority**: HIGH
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [x] Crear `MetricsRepository`
- [x] Implementar getMetrics(): Flow<MetricsData>
- [x] Implementar exportMetrics(): Flow<ExportResult>
- [x] Crear MetricsDao (no aplica - metrics es API-only)
- [x] Implementar caching (no aplica - datos en tiempo real)
- [x] Integrar MetricsApi
- [x] Tests unitarios pasan

**Acceptance Criteria**:
- [x] Repository funciona
- [x] Caching implementado (N/A - datos en tiempo real)
- [x] API integrada
- [x] Tests coverage >70%

**Definition of Done**:
- MetricsRepository completa
- Tests pasan

---

##### TASK-METRICS-003: Crear Charts con Vico
**Description**: Implementar gráficos con librería Vico
**Estimate**: 8 horas (1 día)
**Priority**: HIGH
**Dependencies**: TASK-METRICS-001

**Subtasks**:
- [x] Crear `MetricsLineChart` @Composable
- [x] Configurar Vico CartesianChart
- [x] Implementar eje X (tiempo)
- [x] Implementar eje Y (valores)
- [x] Implementar marcadores
- [x] Crear `MetricsBarChart` @Composable
- [x] Configurar datos de barras
- [x] Crear `MetricsDonutChart` @Composable
- [x] Configurar datos de donut (Canvas API - Vico no soporta donut)
- [x] Implementar animaciones
- [x] Tests UI pasan

**Acceptance Criteria**:
- [x] Line chart funciona
- [x] Bar chart funciona
- [x] Donut chart funciona
- [x] Animaciones funcionan
- [x] Tests UI pasan

**Definition of Done**:
- Todos los charts completos
- Tests UI pasan

---

##### TASK-METRICS-004: Crear MetricsScreen UI
**Description**: Implementar screen principal de métricas
**Estimate**: 5 horas
**Priority**: HIGH
**Dependencies**: TASK-METRICS-001

**Subtasks**:
- [x] Crear `MetricsScreen` @Composable
- [x] Implementar loading state
- [x] Implementar success state con charts
- [x] Implementar empty state
- [x] Implementar error state
- [x] Implementar TimeRangeSelector
- [x] Implementar AgentFilterDropdown
- [x] Implementar ExportButton
- [x] Tests UI pasan

**Acceptance Criteria**:
- [x] Screen funciona
- [x] Charts se muestran
- [x] Filtros funcionan
- [x] Export funciona
- [x] Tests UI pasan

**Definition of Done**:
- MetricsScreen completa
- Tests UI pasan

---

#### feature-cron (3 días)

##### TASK-CRON-001: Crear CronViewModel
**Description**: Implementar ViewModel para cron jobs
**Estimate**: 3 horas
**Priority**: HIGH
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [x] Crear `CronUiState`
- [x] Crear `CronEvent`
- [x] Crear `CronEffect`
- [x] Crear `CronViewModel`
- [x] Implementar loadJobs()
- [x] Implementar createJob()
- [x] Implementar toggleJob()
- [x] Implementar generateCron()
- [x] Implementar testRun()
- [x] Tests unitarios pasan

**Acceptance Criteria**:
- [x] ViewModel funciona
- [x] Eventos manejan correctamente
- [x] Tests coverage >70%

**Definition of Done**:
- CronViewModel completa
- Tests pasan

---

##### TASK-CRON-002: Crear CronRepository
**Description**: Implementar Repository con API y Room
**Estimate**: 3 horas
**Priority**: HIGH
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [x] Crear `CronRepository`
- [x] Implementar getJobs(): Flow<List<CronJob>>
- [x] Implementar createJob(): Result<String>
- [x] Implementar toggleJob(): Result<Unit>
- [x] Implementar generateCron(): Result<String>
- [x] Implementar testRun(): Result<Pair<Boolean, List<String>>>
- [x] Crear CronJobDao (ya existía en core-database)
- [x] Implementar caching (cache-first con Room)
- [x] Integrar CronApi
- [x] Tests unitarios pasan

**Acceptance Criteria**:
- [x] Repository funciona
- [x] Caching implementado
- [x] API integrada
- [x] Tests coverage >70%

**Definition of Done**:
- CronRepository completa
- Tests pasan

---

##### TASK-CRON-003: Crear CronScheduleSelector
**Description**: Implementar selector visual de horarios
**Estimate**: 8 horas (1 día)
**Priority**: HIGH
**Dependencies**: TASK-CRON-001

**Subtasks**:
- [x] Crear `CronScheduleSelector` @Composable
- [x] Crear `HourMinuteSelector` @Composable
- [x] Implementar grid de horas
- [x] Implementar selección de hora
- [x] Crear `DayOfWeekSelector` @Composable
- [x] Implementar toggle de días
- [x] Crear `CronExpressionDisplay`
- [x] Mostrar expresión generada
- [x] Implementar buildCronExpression() helper
- [x] Tests UI pasan

**Acceptance Criteria**:
- [x] Grid de horas funciona
- [x] Selección de hora funciona
- [x] Días se seleccionan
- [x] Expresión se muestra
- [x] Generación AI funciona
- [x] Tests UI pasan

**Definition of Done**:
- CronScheduleSelector completa
- Tests UI pasan

---

##### TASK-CRON-004: Crear CronScreen UI
**Description**: Implementar screen principal de cron
**Estimate**: 4 horas
**Priority**: HIGH
**Dependencies**: TASK-CRON-001

**Subtasks**:
- [x] Crear `CronScreen` @Composable (rewrite)
- [x] Implementar loading state
- [x] Implementar success state con LazyColumn
- [x] Implementar empty state
- [x] Implementar error state
- [x] Implementar CronJobCard
- [x] Implementar FAB para crear job
- [x] Crear CronEditorModal
- [x] Tests UI pasan

**Acceptance Criteria**:
- [x] Screen funciona
- [x] Cards muestran jobs
- [x] Toggle funciona
- [x] Tests UI pasan

**Definition of Done**:
- CronScreen completa
- Tests UI pasan

---

### Milestone 3: Prioridad MEDIA

#### feature-skills (3 días)

##### TASK-SKILLS-001: Crear SkillsViewModel
**Estimate**: 3 horas
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [ ] Crear `SkillsUiState`
- [ ] Crear `SkillsEvent`
- [ ] Crear `SkillsEffect`
- [ ] Crear `SkillsViewModel`
- [ ] Implementar loadInstalled()
- [ ] Implementar loadMarketplace()
- [ ] Implementar installSkill()
- [ ] Implementar uninstallSkill()
- [ ] Implementar generateSkill()
- [ ] Tests unitarios pasan

---

##### TASK-SKILLS-002: Crear SkillsRepository
**Estimate**: 3 horas
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [ ] Crear `SkillsRepository`
- [ ] Implementar getInstalled(): Flow<List<Skill>>
- [ ] Implementar getMarketplace(): Flow<List<Skill>>
- [ ] Implementar install(): Flow<Unit>
- [ ] Implementar uninstall(): Flow<Unit>
- [ ] Implementar generate(): Flow<Skill>
- [ ] Crear SkillDao
- [ ] Integrar SkillsApi
- [ ] Tests unitarios pasan

---

##### TASK-SKILLS-003: Crear SkillsScreen UI
**Estimate**: 4 horas
**Dependencies**: TASK-SKILLS-001

**Subtasks**:
- [ ] Crear `SkillsScreen` con TabView
- [ ] Implementar tab "Installed"
- [ ] Implementar tab "Marketplace"
- [ ] Implementar tab "Generate"
- [ ] Implementar SkillCard
- [ ] Implementar SearchBar
- [ ] Implementar MarketplaceFilter
- [ ] Implementar InstallButton
- [ ] Implementar UninstallConfirmDialog
- [ ] Implementar RatingWidget
- [ ] Tests UI pasan

---

#### feature-agents (3 días)

##### TASK-AGENTS-001: Crear AgentsViewModel
**Estimate**: 3 horas
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [ ] Crear `AgentsUiState`
- [ ] Crear `AgentsEvent`
- [ ] Crear `AgentsEffect`
- [ ] Crear `AgentsViewModel`
- [ ] Implementar loadSpecialists()
- [ ] Implementar createSpecialist()
- [ ] Implementar editSpecialist()
- [ ] Implementar deleteSpecialist()
- [ ] Implementar generateSpecialist()
- [ ] Implementar toggleOrchestrator()
- [ ] Tests unitarios pasan

---

##### TASK-AGENTS-002: Crear SwarmVisualizer 🚨
**Estimate**: 8 horas (1 día)
**Dependencies**: TASK-AGENTS-001

**Subtasks**:
- [ ] Crear `SwarmVisualizer` @Composable
- [ ] Implementar Canvas con Canvas API
- [ ] Implementar agent nodes
- [ ] Implementar connection lines
- [ ] Implementar animaciones de estado
- [ ] Implementar status indicators
- [ ] Implementar zoom/pan
- [ ] Tests UI pasan

---

##### TASK-AGENTS-003: Crear AgentsScreen UI
**Estimate**: 4 horas
**Dependencies**: TASK-AGENTS-001

**Subtasks**:
- [ ] Crear `AgentsScreen` @Composable
- [ ] Implementar SpecialistCard
- [ ] Implementar CreateSpecialistForm
- [ ] Implementar EditSpecialistSheet
- [ ] Implementar SpecialistMetricsView
- [ ] Implementar SpecialistLogsModal
- [ ] Implementar OrchestratorToggle
- [ ] Tests UI pasan

---

#### feature-memory (3 días)

##### TASK-MEMORY-001: Crear MemoryViewModel
**Estimate**: 3 horas
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [ ] Crear `MemoryUiState`
- [ ] Crear `MemoryEvent`
- [ ] Crear `MemoryEffect`
- [ ] Crear `MemoryViewModel`
- [ ] Implementar loadLongTerm()
- [ ] Implementar loadDailyNotes()
- [ ] Implementar updateLongTerm()
- [ ] Implementar createDailyNote()
- [ ] Implementar searchMemory()
- [ ] Tests unitarios pasan

---

##### TASK-MEMORY-002: Crear MemoryScreen UI
**Estimate**: 4 horas
**Dependencies**: TASK-MEMORY-001

**Subtasks**:
- [ ] Crear `MemoryScreen` con TabView
- [ ] Implementar tab "Long-term"
- [ ] Implementar tab "Daily Notes"
- [ ] Implementar LongTermMemoryEditor
- [ ] Implementar DailyNotesTimeline
- [ ] Implementar DailyNoteEditor
- [ ] Implementar MemorySearchBar
- [ ] Implementar RetentionSettingsSheet
- [ ] Tests UI pasan

---

### Milestone 4: Prioridad BAJA

#### feature-files (2 días)

##### TASK-FILES-001: Crear FilesViewModel
**Estimate**: 3 horas
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [ ] Crear `FilesUiState`
- [ ] Crear `FilesEvent`
- [ ] Crear `FilesEffect`
- [ ] Crear `FilesViewModel`
- [ ] Implementar browseDirectory()
- [ ] Implementar uploadFile()
- [ ] Implementar downloadFile()
- [ ] Implementar deleteFile()
- [ ] Implementar createFolder()
- [ ] Implementar renameFile()
- [ ] Tests unitarios pasan

---

##### TASK-FILES-002: Crear FilesScreen UI
**Estimate**: 3 horas
**Dependencies**: TASK-FILES-001

**Subtasks**:
- [ ] Crear `FilesBrowser`
- [ ] Implementar FileCard
- [ ] Implementar FolderCard
- [ ] Implementar UploadFileButton
- [ ] Implementar DownloadButton
- [ ] Implementar FilePreviewModal
- [ ] Implementar CreateFolderSheet
- [ ] Implementar RenameFileSheet
- [ ] Tests UI pasan

---

#### feature-mcp (2 días)

##### TASK-MCP-001: Crear McpViewModel y Screen
**Estimate**: 4 horas
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [ ] Crear `McpViewModel`
- [ ] Crear `McpRepository`
- [ ] Crear `McpScreen`
- [ ] Implementar McpServerCard
- [ ] Implementar McpServerDetailsModal
- [ ] Implementar McpServerConfigSheet
- [ ] Implementar ServerStatusIndicator
- [ ] Implementar ToolsList
- [ ] Implementar ServerLogsModal
- [ ] Implementar ReconnectButton
- [ ] Tests unitarios + UI pasan

---

#### feature-history (2 días)

##### TASK-HISTORY-001: Crear HistoryViewModel y Screen
**Estimate**: 4 horas
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [ ] Crear `HistoryViewModel`
- [ ] Crear `HistoryRepository`
- [ ] Crear `HistoryList`
- [ ] Implementar SessionCard
- [ ] Implementar SessionDetailsModal
- [ ] Implementar SearchBar
- [ ] Implementar DateFilter
- [ ] Implementar AgentFilter
- [ ] Implementar DeleteSessionButton
- [ ] Implementar ExportButton (JSON/Markdown)
- [ ] Implementar ArchiveButton
- [ ] Implementar BulkActionsSheet
- [ ] Tests unitarios + UI pasan

---

#### feature-reports (2 días)

##### TASK-REPORTS-001: Crear ReportsViewModel y Screen
**Estimate**: 4 horas
**Dependencies**: TASK-CORE-001

**Subtasks**:
- [ ] Crear `ReportsViewModel`
- [ ] Crear `ReportsRepository`
- [ ] Crear `ReportGeneratorForm`
- [ ] Implementar TemplateSelector
- [ ] Implementar FilterConfig (fecha, agente, tipo)
- [ ] Implementar ScheduleReportSheet
- [ ] Implementar ReportHistoryList
- [ ] Implementar ExportButton (PDF/CSV)
- [ ] Implementar CancelScheduleButton
- [ ] Tests unitarios + UI pasan

---

### Milestone 5: Polish & QA (2 semanas)

#### TASK-POLISH-001: Tests completos
**Estimate**: 3 días
**Priority**: HIGH

**Subtasks**:
- [ ] Unit tests para todos los ViewModels
- [ ] Unit tests para todos los Repositories
- [ ] Integration tests para APIs
- [ ] UI tests para todas las Screens
- [ ] Tests de navegación
- [ ] Tests de componentes reutilizables

**Acceptance Criteria**:
- [ ] >70% coverage en ViewModels
- [ ] >70% coverage en Repositories
- [ ] Todos los tests pasan

---

#### TASK-POLISH-002: UI Polish
**Estimate**: 2 días
**Priority**: HIGH

**Subtasks**:
- [ ] Animaciones fluidas (60fps)
- [ ] Transiciones consistentes
- [ ] Empty states con CTAs claros
- [ ] Loading states consistentes
- [ ] Error states con vías de recuperación
- [ ] Dark/Light theme testing
- [ ] Accessibility (TalkBack, content descriptions)

**Acceptance Criteria**:
- [ ] Animaciones funcionan sin lag
- [ ] Estados visibles
- [ ] Accessibility OK

---

#### TASK-POLISH-003: Bug Fixes
**Estimate**: 3 días
**Priority**: HIGH

**Subtasks**:
- [ ] Fix bugs encontrados en testing
- [ ] Fix performance issues
- [ ] Fix edge cases
- [ ] Validar fixes

---

#### TASK-POLISH-004: Performance Optimization
**Estimate**: 2 días
**Priority**: MEDIUM

**Subtasks**:
- [ ] Profiling de CPU
- [ ] Profiling de memoria
- [ ] Optimizar listas grandes (LazyColumn)
- [ ] Implementar pagination
- [ ] Optimizar imágenes (Coil)
- [ ] Validar mejoras

---

## Resumen de Tareas

| Milestone | Tareas | Estimación |
|-----------|--------|-----------|
| Milestone 1 | 6 tasks | 2 días |
| Milestone 2 | 12 tasks | 12 días |
| Milestone 3 | 9 tasks | 9 días |
| Milestone 4 | 6 tasks | 14 horas (2 días) |
| Milestone 5 | 4 tasks | 10 días |
| **TOTAL** | **37 tasks** | **~35 días (5-6 semanas)** |

## Task Status

| ID | Task | Estado | Asignado | Prioridad |
|----|------|--------|----------|----------|
| TASK-CORE-001 | Completar core-database | Completed | - | CRITICAL |
| TASK-CORE-002 | Completar core-datastore | Completed | - | CRITICAL |
| TASK-CORE-003 | Completar core-security | Completed | - | CRITICAL |
| TASK-KNOWLEDGE-001 | Crear KnowledgeViewModel | Pending | - | HIGH |
| ... | ... | ... | ... | ... |

## Dependencies Graph

```
Core Modules (TASK-CORE-001/002/003)
    ↓
All Features depend on Core
    ↓
Milestone 1: Core + Knowledge
    ↓
Milestone 2: Workflows, Metrics, Cron
    ↓
Milestone 3: Skills, Agents, Memory
    ↓
Milestone 4: Files, MCP, History, Reports
    ↓
Milestone 5: Polish & QA
```
