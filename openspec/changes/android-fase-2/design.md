# Design: Android Fase 2 - Feature Parity

## Technical Approach

Extender la arquitectura modular existente (MVVM + Clean Architecture) para completar los 11 features esqueletales. Mantener consistencia con patrones actuales: ViewModels con StateFlow/SharedFlow, Repositories con Result<T>, Room entities, y Compose UI. Completar core modules (database, datastore, security) como base, luego implementar features en paralelo según prioridad técnica.

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|--------------|-----------|
| **Workflows Editor** | Canvas API custom con gestures | ReactNodeCompose, GraphView | Máximo control UX, menor overhead de librería externa, consistente con Compose patterns existentes |
| **Charts** | Vico 2.0 | MPAndroidChart, Compose Charts | Vico es Compose-native, mejor integración Material3, soporta lazy loading para grandes datasets |
| **Cron Selector** | Custom dial + lists UI | Cronet, Android Chronometer | UI visual más intuitiva, UX coherente con app, evita dependencias adicionales |
| **Markdown** | Markwon 4.6 | Compose Markdown, Coil + TextView | Markwon es estable, soporta syntax highlighting, ya evaluado en Fase 1 |
| **File Permissions** | Accompanist Permissions | Native permission handlers | Consistente con Compose, mejor DX, callbacks declarativos |
| **JWT Storage** | EncryptedSharedPreferences | Keystore, DataStore encrypted | Balance seguridad/simplicidad, ya en specs de core-security |
| **Image Loading** | Coil 2.4 | Glide, Fresco | Coil es Compose-first, async image loading, caching automático |
| **Auto-save** | Debounced Flow de state changes | Manual save buttons, JobScheduler | UX sin interrupciones, sincronización eficiente con backend |
| **Swarm Visualizer** | Canvas API custom | Lottie, Compose Animation | Performance superior para nodes dinámicos, interacción real-time |

## Data Flow

```
User Action (Compose UI)
    ↓
ViewModel (StateFlow + SharedFlow)
    ↓
Use Case (optional - business logic)
    ↓
Repository (Result<T>)
    ↓
Data Source (Room local / Retrofit remote)
    ↓
Backend API
    ↑
TokenManager (core-security)
    ↑
EncryptedSharedPreferences (JWT storage)
```

**Data persistence pattern:**
```
API Response → Repository → Room Entity → Flow → ViewModel → UI
                           ↑
                    Cache-first strategy
```

## File Changes

### Core Modules

| File | Action | Description |
|------|--------|-------------|
| `core/core-database/src/main/java/com/makoclaw/core/database/entity/Entities.kt` | Modify | Agregar entities: KnowledgeDocumentEntity, WorkflowEntity, MetricsEntity, CronJobEntity, SkillEntity, FileEntryEntity, MemoryEntity, McpServerEntity, ReportEntity |
| `core/core-database/src/main/java/com/makoclaw/core/database/dao/Daos.kt` | Modify | Agregar DAOs: KnowledgeDocumentDao, WorkflowDao, MetricsDao, CronJobDao, SkillDao, FileEntryDao, MemoryDao, McpServerDao, ReportDao |
| `core/core-datastore/src/main/java/com/makoclaw/core/datastore/UserPreferences.kt` | Modify | Agregar preferences: knowledgeUploadSize, metricsTimeRange, cronNotify, skillsAutoInstall, memoryRetentionDays, historyExportFormat |
| `core/core-security/src/main/java/com/makoclaw/core/security/TokenManager.kt` | Modify | Implementar saveToken(), getToken(), clearToken() con EncryptedSharedPreferences |
| `core/core-network/src/main/java/com/makoclaw/core/network/api/FeatureApis.kt` | Modify | Completar Retrofit interfaces para todos los endpoints de los 11 features (según specs) |
| `core/core-model/src/main/java/com/makoclaw/core/model/Knowledge.kt` | Modify | Completar data classes con todos los campos necesarios (según specs) |
| `core/core-model/src/main/java/com/makoclaw/core/model/Workflow.kt` | Create | WorkflowNode, WorkflowEdge, WorkflowConfig |
| `core/core-model/src/main/java/com/makoclaw/core/model/Metrics.kt` | Create | MetricsData, TimeRange, ExportFormat |
| `core/core-model/src/main/java/com/makoclaw/core/model/Cron.kt` | Create | CronSchedule, CronPreset, TestResult |
| `core/core-ui/src/main/java/com/makoclaw/core/ui/components/` | Create | EmptyState, LoadingScreen, ErrorScreen, SearchBar, FilterSheet, ConfirmDialog, ProgressIndicator |

### Feature Modules (11 features)

#### feature-knowledge
| File | Action | Description |
|------|--------|-------------|
| `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/state/KnowledgeUiState.kt` | Create | Data class con documents, loading, error, preview state |
| `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/viewmodel/KnowledgeViewModel.kt` | Create | @HiltViewModel con StateFlow, eventos LoadDocuments, UploadDocument, etc. |
| `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/data/repository/KnowledgeRepository.kt` | Create | KnowledgeRepository con getDocuments(), uploadDocument(), searchDocuments() |
| `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/screen/KnowledgeScreen.kt` | Create | Compose screen con LazyColumn, UploadFileButton, SearchBar |
| `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/components/KnowledgeDocumentCard.kt` | Create | Document card con metadata, preview, delete |
| `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/components/DocumentPreviewModal.kt` | Create | Modal con Markwon rendering para PDF/TXT/MD |
| `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/components/ChunkEditorSheet.kt` | Create | Bottom sheet para editar chunks con auto-save |

#### feature-workflows
| File | Action | Description |
|------|--------|-------------|
| `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/state/WorkflowsUiState.kt` | Create | Data class con workflows, editor state, execution logs |
| `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/viewmodel/WorkflowsViewModel.kt` | Create | @HiltViewModel con LoadWorkflows, CreateWorkflow, ExecuteWorkflow |
| `feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/repository/WorkflowsRepository.kt` | Create | WorkflowsRepository con getWorkflows(), createWorkflow(), executeWorkflow() |
| `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/screen/WorkflowsScreen.kt` | Create | Compose screen con WorkflowCanvasEditor, WorkflowCard, RunWorkflowButton |
| `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/components/WorkflowCanvasEditor.kt` | Create | Canvas con gestures (drag, pan, zoom), nodes, edges (HIGH COMPLEXITY) |
| `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/components/WorkflowNode.kt` | Create | Draggable node con content, inputs, outputs |
| `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/components/WorkflowEdge.kt` | Create | Curved Bézier connection line entre nodes |
| `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/components/ExecutionLogsModal.kt` | Create | Modal con logs de ejecución en tiempo real |

#### feature-metrics
| File | Action | Description |
|------|--------|-------------|
| `feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/state/MetricsUiState.kt` | Create | Data class con metrics data, time filter, agent filter |
| `feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/viewmodel/MetricsViewModel.kt` | Create | @HiltViewModel con LoadMetrics, SetTimeRange, SetAgentFilter, ExportMetrics |
| `feature-metrics/src/main/java/com/makoclaw/feature/metrics/data/repository/MetricsRepository.kt` | Create | MetricsRepository con getMetrics(), exportMetrics() |
| `feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/screen/MetricsScreen.kt` | Create | Compose screen con Vico charts (LineChart, BarChart, DonutChart) |
| `feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/components/MetricsLineChart.kt` | Create | Vico LineChart para LLM calls, tokens |
| `feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/components/MetricsBarChart.kt` | Create | Vico BarChart para tool usage |
| `feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/components/MetricsDonutChart.kt` | Create | Vico DonutChart para cost distribution |
| `feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/components/TimeRangeSelector.kt` | Create | Dropdown para seleccionar rango de tiempo |
| `feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/components/AgentFilterDropdown.kt` | Create | Dropdown para filtrar por agente |
| `feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/components/ExportButton.kt` | Create | Button para exportar CSV |

#### feature-cron
| File | Action | Description |
|------|--------|-------------|
| `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/state/CronUiState.kt` | Create | Data class con jobs list, schedule state, history |
| `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/viewmodel/CronViewModel.kt` | Create | @HiltViewModel con LoadJobs, CreateJob, ToggleJob, GenerateCron, TestRun |
| `feature-cron/src/main/java/com/makoclaw/feature/cron/data/repository/CronRepository.kt` | Create | CronRepository con getJobs(), createJob(), toggleJob(), generateCron() |
| `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/screen/CronScreen.kt` | Create | Compose screen con CronJobCard, CronScheduleSelector, GenerateWithAIButton |
| `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/components/CronScheduleSelector.kt` | Create | Custom UI con HourDialSelector, DayOfWeekSelector, presets list |
| `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/components/HourDialSelector.kt` | Create | Dial rotatorio para seleccionar hora |
| `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/components/DayOfWeekSelector.kt` | Create | Checkboxes para días de la semana |
| `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/components/CronExpressionDisplay.kt` | Create | Display de cron expression generada |
| `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/components/ExecutionHistoryModal.kt` | Create | Modal con historial de ejecuciones |

#### feature-skills
| File | Action | Description |
|------|--------|-------------|
| `feature-skills/src/main/java/com/makoclaw/feature/skills/presentation/state/SkillsUiState.kt` | Create | Data class con installed, marketplace, search query, filter |
| `feature-skills/src/main/java/com/makoclaw/feature/skills/presentation/viewmodel/SkillsViewModel.kt` | Create | @HiltViewModel con LoadInstalled, LoadMarketplace, InstallSkill, UninstallSkill, GenerateSkill |
| `feature-skills/src/main/java/com/makoclaw/feature/skills/data/repository/SkillsRepository.kt` | Create | SkillsRepository con getInstalled(), getMarketplace(), installSkill() |
| `feature-skills/src/main/java/com/makoclaw/feature/skills/presentation/screen/SkillsScreen.kt` | Create | Compose screen con Tabs (Installed/Marketplace/Generate) |
| `feature-skills/src/main/java/com/makoclaw/feature/skills/presentation/components/SkillsTabView.kt` | Create | TabRow con Tab para cada sección |
| `feature-skills/src/main/java/com/makoclaw/feature/skills/presentation/components/SkillCard.kt` | Create | Card con skill info, rating, InstallButton |
| `feature-skills/src/main/java/com/makoclaw/feature/skills/presentation/components/MarketplaceFilter.kt` | Create | Filter por categoría |
| `feature-skills/src/main/java/com/makoclaw/feature/skills/presentation/components/SkillSearchBar.kt` | Create | Search bar con debouncing |
| `feature-skills/src/main/java/com/makoclaw/feature/skills/presentation/components/UninstallConfirmDialog.kt` | Create | ConfirmDialog para desinstalación |
| `feature-skills/src/main/java/com/makoclaw/feature/skills/presentation/components/GenerateSkillForm.kt` | Create | Form para generación con AI |
| `feature-skills/src/main/java/com/makoclaw/feature/skills/presentation/components/RatingWidget.kt` | Create | 5-star rating interactive widget |
| `feature-skills/src/main/java/com/makoclaw/feature/skills/presentation/components/SkillDetailsModal.kt` | Create | Modal con detalles del skill |

#### feature-agents
| File | Action | Description |
|------|--------|-------------|
| `feature-agents/src/main/java/com/makoclaw/feature/agents/presentation/state/AgentsUiState.kt` | Modify | Agregar metrics, logs, edit state |
| `feature-agents/src/main/java/com/makoclaw/feature/agents/presentation/viewmodel/AgentsViewModel.kt` | Modify | Agregar EditSpecialist, GenerateSpecialist, ToggleOrchestrator |
| `feature-agents/src/main/java/com/makoclaw/feature/agents/data/repository/AgentsRepository.kt` | Create | AgentsRepository con getSpecialists(), createSpecialist(), updateSpecialist() |
| `feature-agents/src/main/java/com/makoclaw/feature/agents/presentation/screen/AgentsScreen.kt` | Modify | Agregar SwarmVisualizer, SpecialistMetricsView, SpecialistLogsModal |
| `feature-agents/src/main/java/com/makoclaw/feature/agents/presentation/components/SwarmVisualizer.kt` | Create | Canvas con agent nodes, connections, animations |
| `feature-agents/src/main/java/com/makoclaw/feature/agents/presentation/components/CreateSpecialistForm.kt` | Create | Form con name, speciality, system prompt |
| `feature-agents/src/main/java/com/makoclaw/feature/agents/presentation/components/EditSpecialistSheet.kt` | Create | Bottom sheet para editar especialista |
| `feature-agents/src/main/java/com/makoclaw/feature/agents/presentation/components/SpecialistMetricsView.kt` | Create | View con métricas específicas del especialista |
| `feature-agents/src/main/java/com/makoclaw/feature/agents/presentation/components/SpecialistLogsModal.kt` | Create | Modal con logs del especialista |
| `feature-agents/src/main/java/com/makoclaw/feature/agents/presentation/components/GenerateWithAIButton.kt` | Create | Button para generar especialista con AI |
| `feature-agents/src/main/java/com/makoclaw/feature/agents/presentation/components/OrchestratorToggle.kt` | Create | Toggle switch para activar/desactivar orchestrator |

#### feature-memory
| File | Action | Description |
|------|--------|-------------|
| `feature-memory/src/main/java/com/makoclaw/feature/memory/presentation/state/MemoryUiState.kt` | Create | Data class con long-term memory, daily notes, search query, filters |
| `feature-memory/src/main/java/com/makoclaw/feature/memory/presentation/viewmodel/MemoryViewModel.kt` | Create | @HiltViewModel con LoadLongTermMemory, LoadDailyNotes, UpdateLongTerm, SearchMemory, SetRetention |
| `feature-memory/src/main/java/com/makoclaw/feature/memory/data/repository/MemoryRepository.kt` | Create | MemoryRepository con getLongTermMemory(), getDailyNotes(), searchMemory() |
| `feature-memory/src/main/java/com/makoclaw/feature/memory/presentation/screen/MemoryScreen.kt` | Create | Compose screen con Tabs (Long-term/Daily Notes) |
| `feature-memory/src/main/java/com/makoclaw/feature/memory/presentation/components/MemoryTabView.kt` | Create | TabRow con Tab para cada sección |
| `feature-memory/src/main/java/com/makoclaw/feature/memory/presentation/components/LongTermMemoryEditor.kt` | Create | Editor con auto-save (debounced) |
| `feature-memory/src/main/java/com/makoclaw/feature/memory/presentation/components/DailyNotesTimeline.kt` | Create | Timeline vertical con notas ordenadas por fecha |
| `feature-memory/src/main/java/com/makoclaw/feature/memory/presentation/components/DailyNoteEditor.kt` | Create | Editor para crear/editar nota diaria |
| `feature-memory/src/main/java/com/makoclaw/feature/memory/presentation/components/MemorySearchBar.kt` | Create | Search bar para buscar en memoria |
| `feature-memory/src/main/java/com/makoclaw/feature/memory/presentation/components/RetentionSettingsSheet.kt` | Create | Bottom sheet para configurar retención |
| `feature-memory/src/main/java/com/makoclaw/feature/memory/presentation/components/DatePicker.kt` | Create | DatePicker para seleccionar fecha de nota |

#### feature-files
| File | Action | Description |
|------|--------|-------------|
| `feature-files/src/main/java/com/makoclaw/feature/files/presentation/state/FilesUiState.kt` | Create | Data class con current directory, files, upload progress, preview state |
| `feature-files/src/main/java/com/makoclaw/feature/files/presentation/viewmodel/FilesViewModel.kt` | Create | @HiltViewModel con BrowseDirectory, UploadFile, DownloadFile, DeleteFile, CreateFolder, RenameFile |
| `feature-files/src/main/java/com/makoclaw/feature/files/data/repository/FilesRepository.kt` | Create | FilesRepository con browseDirectory(), uploadFile(), downloadFile() |
| `feature-files/src/main/java/com/makoclaw/feature/files/presentation/screen/FilesBrowser.kt` | Create | Compose screen con FileCard, FolderCard, breadcrumbs |
| `feature-files/src/main/java/com/makoclaw/feature/files/presentation/components/FileCard.kt` | Create | Card con file info, download, delete actions |
| `feature-files/src/main/java/com/makoclaw/feature/files/presentation/components/FolderCard.kt` | Create | Card con folder info, navigate action |
| `feature-files/src/main/java/com/makoclaw/feature/files/presentation/components/UploadFileButton.kt` | Create | Button con ActivityResultContracts para seleccionar archivo |
| `feature-files/src/main/java/com/makoclaw/feature/files/presentation/components/DownloadButton.kt` | Create | Button para descargar archivo |
| `feature-files/src/main/java/com/makoclaw/feature/files/presentation/components/FilePreviewModal.kt` | Create | Modal con Coil para imágenes, PDF viewer para PDF |
| `feature-files/src/main/java/com/makoclaw/feature/files/presentation/components/CreateFolderSheet.kt` | Create | Bottom sheet para crear carpeta |
| `feature-files/src/main/java/com/makoclaw/feature/files/presentation/components/RenameFileSheet.kt` | Create | Bottom sheet para renombrar archivo |
| `feature-files/src/main/java/com/makoclaw/feature/files/presentation/components/FileTypeFilter.kt` | Create | Filter para mostrar solo ciertos tipos de archivos |

#### feature-mcp
| File | Action | Description |
|------|--------|-------------|
| `feature-mcp/src/main/java/com/makoclaw/feature/mcp/presentation/state/McpUiState.kt` | Create | Data class con servers list, selected server, logs, connection status |
| `feature-mcp/src/main/java/com/makoclaw/feature/mcp/presentation/viewmodel/McpViewModel.kt` | Create | @HiltViewModel con LoadServers, ConfigureServer, ToggleServer, ViewLogs, ReconnectServer |
| `feature-mcp/src/main/java/com/makoclaw/feature/mcp/data/repository/McpRepository.kt` | Create | McpRepository con getServers(), configureServer(), toggleServer() |
| `feature-mcp/src/main/java/com/makoclaw/feature/mcp/presentation/screen/McpScreen.kt` | Create | Compose screen con McpServerCard, McpServerDetailsModal, McpServerConfigSheet |
| `feature-mcp/src/main/java/com/makoclaw/feature/mcp/presentation/components/McpServerCard.kt` | Create | Card con server info, status indicator, toggle, actions |
| `feature-mcp/src/main/java/com/makoclaw/feature/mcp/presentation/components/McpServerDetailsModal.kt` | Create | Modal con endpoint, tools list, config |
| `feature-mcp/src/main/java/com/makoclaw/feature/mcp/presentation/components/McpServerConfigSheet.kt` | Create | Bottom sheet para configurar servidor |
| `feature-mcp/src/main/java/com/makoclaw/feature/mcp/presentation/components/ServerStatusIndicator.kt` | Create | Indicator verde/rojo para connected/disconnected |
| `feature-mcp/src/main/java/com/makoclaw/feature/mcp/presentation/components/ToolsList.kt` | Create | List de tools disponibles del servidor |
| `feature-mcp/src/main/java/com/makoclaw/feature/mcp/presentation/components/ServerLogsModal.kt` | Create | Modal con logs en tiempo real |
| `feature-mcp/src/main/java/com/makoclaw/feature/mcp/presentation/components/ReconnectButton.kt` | Create | Button para reconnect manual |

#### feature-history
| File | Action | Description |
|------|--------|-------------|
| `feature-history/src/main/java/com/makoclaw/feature/history/presentation/state/HistoryUiState.kt` | Create | Data class con sessions list, filters, search query, selected session |
| `feature-history/src/main/java/com/makoclaw/feature/history/presentation/viewmodel/HistoryViewModel.kt` | Create | @HiltViewModel con LoadSessions, SearchSessions, FilterSessions, ViewSessionDetails, DeleteSession, ArchiveSession, ExportSession |
| `feature-history/src/main/java/com/makoclaw/feature/history/data/repository/HistoryRepository.kt` | Create | HistoryRepository con getSessions(), searchSessions(), filterSessions(), exportSession() |
| `feature-history/src/main/java/com/makoclaw/feature/history/presentation/screen/HistoryList.kt` | Create | Compose screen con SessionCard, SearchBar, filters, bulk actions |
| `feature-history/src/main/java/com/makoclaw/feature/history/presentation/components/SessionCard.kt` | Create | Card con session metadata, actions (view, delete, archive, export) |
| `feature-history/src/main/java/com/makoclaw/feature/history/presentation/components/SessionDetailsModal.kt` | Create | Modal con todos los mensajes de la sesión |
| `feature-history/src/main/java/com/makoclaw/feature/history/presentation/components/SearchBar.kt` | Create | Search bar con debouncing |
| `feature-history/src/main/java/com/makoclaw/feature/history/presentation/components/DateFilter.kt` | Create | Filter por rango de fechas |
| `feature-history/src/main/java/com/makoclaw/feature/history/presentation/components/AgentFilter.kt` | Create | Filter por agente |
| `feature-history/src/main/java/com/makoclaw/feature/history/presentation/components/DeleteSessionButton.kt` | Create | Button para eliminar sesión con confirmación |
| `feature-history/src/main/java/com/makoclaw/feature/history/presentation/components/ExportButton.kt` | Create | Button para exportar (JSON/Markdown) |
| `feature-history/src/main/java/com/makoclaw/feature/history/presentation/components/ArchiveButton.kt` | Create | Button para archivar sesión |
| `feature-history/src/main/java/com/makoclaw/feature/history/presentation/components/BulkActionsSheet.kt` | Create | Bottom sheet para bulk actions (delete/archive multiple) |

#### feature-reports
| File | Action | Description |
|------|--------|-------------|
| `feature-reports/src/main/java/com/makoclaw/feature/reports/presentation/state/ReportsUiState.kt` | Create | Data class con templates, selected template, report config, history |
| `feature-reports/src/main/java/com/makoclaw/feature/reports/presentation/viewmodel/ReportsViewModel.kt` | Create | @HiltViewModel con LoadTemplates, GenerateReport, ScheduleReport, LoadHistory, CancelSchedule, ExportReport |
| `feature-reports/src/main/java/com/makoclaw/feature/reports/data/repository/ReportsRepository.kt` | Create | ReportsRepository con getTemplates(), generateReport(), scheduleReport(), getHistory() |
| `feature-reports/src/main/java/com/makoclaw/feature/reports/presentation/screen/ReportGeneratorForm.kt` | Create | Compose screen con TemplateSelector, FilterConfig, ScheduleReportSheet |
| `feature-reports/src/main/java/com/makoclaw/feature/reports/presentation/components/TemplateSelector.kt` | Create | Dropdown para seleccionar template |
| `feature-reports/src/main/java/com/makoclaw/feature/reports/presentation/components/FilterConfig.kt` | Create | Filters para fecha, agente, tipo |
| `feature-reports/src/main/java/com/makoclaw/feature/reports/presentation/components/ScheduleReportSheet.kt` | Create | Bottom sheet para programar reporte recurrente |
| `feature-reports/src/main/java/com/makoclaw/feature/reports/presentation/components/ReportHistoryList.kt` | Create | List de reportes generados con estado |
| `feature-reports/src/main/java/com/makoclaw/feature/reports/presentation/components/ExportButton.kt` | Create | Button para exportar (PDF/CSV) |
| `feature-reports/src/main/java/com/makoclaw/feature/reports/presentation/components/CancelScheduleButton.kt` | Create | Button para cancelar reporte programado |

### App Module

| File | Action | Description |
|------|--------|-------------|
| `app/src/main/java/com/makoclaw/android/navigation/Route.kt` | Modify | Agregar rutas para las 11 features nuevas |
| `app/src/main/java/com/makoclaw/android/navigation/MakoClawNavHost.kt` | Modify | Agregar composables de navegación para cada feature |

### Build Files

| File | Action | Description |
|------|--------|-------------|
| `build.gradle.kts` | Modify | Agregar librerías: Vico, Markwon, Accompanist Permissions, MPAndroidChart (fallback) |
| `gradle/libs.versions.toml` | Modify | Definir versiones de nuevas dependencias |

## Interfaces / Contracts

### ViewModel Pattern (consistente con `AgentsViewModel`)

```kotlin
data class FeatureUiState(
    val isLoading: Boolean = true,
    val data: List<Item> = emptyList(),
    val error: String? = null
)

sealed interface FeatureEffect {
    data object NavigateToDetail : FeatureEffect
    data class ShowError(val message: String) : FeatureEffect
}

@HiltViewModel
class FeatureViewModel @Inject constructor(
    private val repository: FeatureRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(FeatureUiState())
    val uiState: StateFlow<FeatureUiState> = _uiState.asStateFlow()

    private val _effects = MutableSharedFlow<FeatureEffect>()
    val effects: SharedFlow<FeatureEffect> = _effects.asSharedFlow()

    fun onLoad() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val data = repository.getData()
                _uiState.update { it.copy(isLoading = false, data = data) }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }
}
```

### Repository Pattern (consistente con `AuthRepository`)

```kotlin
@Singleton
class FeatureRepository @Inject constructor(
    private val api: FeatureApi,
    private val dao: FeatureDao
) {
    fun getData(): Flow<List<Item>> {
        return dao.getAll()
            .onEach { if (it.isEmpty()) fetchFromApi() }
    }

    private suspend fun fetchFromApi() {
        val response = api.getData()
        dao.insertAll(response.items.map { it.toEntity() })
    }

    suspend fun createItem(item: Item): Result<Unit> {
        return try {
            api.createItem(item)
            dao.insert(item.toEntity())
            Result.Success(Unit)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Create failed", e)
        }
    }
}
```

### Room Entity Pattern (consistente con `Entities.kt`)

```kotlin
@Entity(tableName = "feature_items")
data class FeatureItemEntity(
    @PrimaryKey val id: String,
    val name: String,
    val createdAt: String
)

fun FeatureItemEntity.toModel() = FeatureItem(
    id = id,
    name = name,
    createdAt = createdAt
)

fun FeatureItem.toEntity() = FeatureItemEntity(
    id = id,
    name = name,
    createdAt = createdAt
)
```

### Vico Charts Integration (feature-metrics)

```kotlin
@Composable
fun MetricsLineChart(data: List<TimeSeriesData>) {
    val chartValues = remember {
        data.map { entry ->
            Point(entry.timestamp, entry.value)
        }
    }

    CartesianChartHost(
        chart = CartesianChart(
            rememberLineCartesianLayer(
                lines = listOf(
                    rememberLine(
                        points = chartValues,
                        style = LineStyle(
                            fill = null,
                            stroke = MaterialTheme.colorScheme.primary.toPath(),
                            thickness = 2.dp
                        )
                    )
                )
            ),
            startAxis = rememberStartAxis(),
            bottomAxis = rememberBottomAxis(),
            marker = rememberMarker()
        )
    ) {
        // Zoom, pan, tooltip interactions
    }
}
```

### Custom Canvas Editor (feature-workflows - HIGH COMPLEXITY)

```kotlin
@Composable
fun WorkflowCanvasEditor(
    nodes: List<WorkflowNode>,
    edges: List<WorkflowEdge>,
    onNodeMove: (nodeId: String, offset: Offset) -> Unit,
    onConnectNodes: (from: String, to: String) -> Unit
) {
    val scale = remember { mutableStateOf(1f) }
    val offset = remember { mutableStateOf(Offset.Zero) }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .pointerInput(Unit) {
                detectTransformGestures { _, pan, zoom, _ ->
                    scale.value *= zoom
                    offset.value += pan
                }
            }
    ) {
        Canvas(modifier = Modifier.fillMaxSize()) {
            withTransform({
                scale(scale.value, Offset.Zero)
                translate(offset.value.x, offset.value.y)
            }) {
                // Draw edges (Bézier curves)
                edges.forEach { edge ->
                    drawPath(
                        path = createBezierPath(edge.start, edge.end),
                        color = Color.Gray,
                        style = Stroke(width = 2.dp.toPx())
                    )
                }

                // Draw nodes
                nodes.forEach { node ->
                    drawRoundRect(
                        color = Color.Blue,
                        topLeft = node.position,
                        size = Size(120.dp.toPx(), 80.dp.toPx()),
                        cornerRadius = CornerRadius(8.dp.toPx())
                    )
                    // Draw node content
                }
            }
        }

        // Add node button
        FloatingActionButton(
            onClick = { /* Add node */ },
            modifier = Modifier.align(Alignment.BottomEnd)
        ) {
            Icon(Icons.Default.Add, "Add node")
        }
    }
}
```

### Cron Schedule Selector (feature-cron)

```kotlin
@Composable
fun CronScheduleSelector(
    schedule: CronSchedule,
    onScheduleChange: (CronSchedule) -> Unit
) {
    Column {
        Text("Hour")
        HourDialSelector(
            selectedHour = schedule.hour,
            onHourChange = { onScheduleChange(schedule.copy(hour = it)) }
        )

        Spacer(modifier = Modifier.height(16.dp))

        Text("Days of Week")
        DayOfWeekSelector(
            selectedDays = schedule.daysOfWeek,
            onDaysChange = { onScheduleChange(schedule.copy(daysOfWeek = it)) }
        )

        Spacer(modifier = Modifier.height(16.dp))

        Text("Cron Expression")
        Text(
            text = generateCronExpression(schedule),
            style = MaterialTheme.typography.bodyMedium,
            fontFamily = FontFamily.Monospace
        )
    }
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | ViewModel state changes, Repository API/DAO calls, Use case business logic | JUnit5 + MockK, coroutine test dispatcher, `runTest` |
| Integration | Retrofit API calls, Room DAOs, DataStore preferences | `@HiltAndroidTest`, in-memory database, mock web server (OkHttp MockWebServer) |
| UI | Compose screen rendering, user interactions, navigation, state transitions | `ComposeTestRule`, `createComposeRule`, semantic assertions, `performClick`, `assertIsDisplayed` |

### Example: ViewModel Test

```kotlin
@HiltAndroidTest
class KnowledgeViewModelTest {

    @get:Rule
    val hiltRule = HiltAndroidRule(this)

    @Inject
    lateinit var viewModel: KnowledgeViewModel

    @Before
    fun setup() {
        hiltRule.inject()
    }

    @Test
    fun `load documents should update state from loading to success`() = runTest {
        // Given
        viewModel.onEvent(KnowledgeEvent.LoadDocuments)

        // Then
        viewModel.uiState.test {
            val initialState = awaitItem()
            assertTrue(initialState.isLoading)

            val loadingState = awaitItem()
            assertTrue(loadingState.isLoading)

            val successState = awaitItem()
            assertFalse(successState.isLoading)
            assertNotNull(successState.data)
        }
    }

    @Test
    fun `upload document with error should show error state`() = runTest {
        // When
        viewModel.onEvent(KnowledgeEvent.UploadDocument(mockFailureFile))

        // Then
        viewModel.uiState.test {
            val errorState = awaitItem()
            assertNotNull(errorState.error)
        }
    }
}
```

### Example: UI Test

```kotlin
class KnowledgeScreenTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun `upload button should be visible`() {
        composeTestRule.setContent {
            MakoClawTheme {
                KnowledgeScreen()
            }
        }

        composeTestRule
            .onNodeWithText("Upload Document")
            .assertIsDisplayed()
    }

    @Test
    fun `clicking upload should show file picker`() {
        composeTestRule.setContent {
            MakoClawTheme {
                KnowledgeScreen()
            }
        }

        composeTestRule
            .onNodeWithText("Upload Document")
            .performClick()

        // Verify file picker intent was triggered
    }
}
```

## Migration / Rollout

No migration requerida para datos existentes. Los nuevos features agregan tablas y columns a la base de datos Room, Room maneja migraciones automáticas con `@Database(version = N, exportSchema = true)`.

### Feature Flags (opcional para rollout gradual)

```kotlin
// core-datastore/UserPreferences.kt
data class FeatureFlags(
    val workflowsEnabled: Boolean = false,
    val metricsEnabled: Boolean = false,
    val cronEnabled: Boolean = false
)

// En ViewModel, deshabilitar features si flag = false
if (featureFlags.workflowsEnabled) {
    // Render workflows UI
}
```

### Rollout Phases (opcional)

1. **Week 1-2**: Core modules + Priority HIGH features (knowledge, workflows, metrics, cron)
2. **Week 3-4**: Priority MED features (skills, agents, memory)
3. **Week 4-5**: Priority LOW features (files, mcp, history, reports)
4. **Week 5-6**: QA, bug fixes, performance optimization

## Open Questions

- [ ] ¿Se requiere soporte offline para features (e.g., workflows ejecutar sin conexión)?
- [ ] ¿Hay límites de almacenamiento local para documentos knowledge o archivos?
- [ ] ¿Vico charts deben soportar export de imágenes (PNG/SVG)?
- [ ] ¿Cron jobs deben ejecutarse localmente (WorkManager) o solo backend?
- [ ] ¿Swarm visualizer debe mostrar historial de conexiones o solo estado actual?

---

## Dependencies

### External Libraries (new)

```kotlin
// Vico (Charts)
implementation("com.patrykandpatrick.vico:compose:2.0.0")
implementation("com.patrykandpatrick.vico:compose-m3:2.0.0")

// Markwon (Markdown)
implementation("io.noties.markwon:core:4.6.0")
implementation("io.noties.markwon:ext-strikethrough:4.6.0")

// Accompanist (Permissions)
implementation("com.google.accompanist:accompanist-permissions:0.32.0")

// Backup Charts (MPAndroidChart - fallback)
implementation("com.github.PhilJay:MPAndroidChart:v3.1.0")
```

### Existing Libraries (reused)

- Compose BOM + Material3
- Navigation Compose
- Lifecycle ViewModel Compose
- Hilt (DI)
- Room (database)
- DataStore (preferences)
- Kotlinx Serialization
- JUnit5 + MockK (testing)
