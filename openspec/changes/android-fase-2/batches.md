# Implementación BATCH 3: Features ALTA (Workflows, Metrics, Cron)

## Instrucciones

Este documento contiene TODO el código fuente para las 13 tareas del BATCH 3. Cada sección incluye:

1. **Ubicación del archivo** - Dónde guardar el archivo
2. **Código completo** - Listo para copiar/pegar
3. **Instrucciones** - Cómo usarlo

---

## FEATURE-WORKFLOWS (5 tareas)

### TASK-WORKFLOWS-001: Crear WorkflowsViewModel

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/state/WorkflowsUiState.kt`
**Paquete**: `com.makoclaw.feature.workflows.presentation.state`

### Código Completo

**Archivo**: `WorkflowsUiState.kt`

```kotlin
package com.makoclaw.feature.workflows.presentation.state

import com.makoclaw.core.model.Workflow
import com.makoclaw.core.model.WorkflowExecutionLog

data class WorkflowsUiState(
    val isLoading: Boolean = false,
    val workflows: List<Workflow> = emptyList(),
    val executionLogs: List<WorkflowExecutionLog> = emptyList(),
    val error: String? = null,
    val isEmpty: Boolean = false,
    val selectedWorkflow: Workflow? = null,
    val isExecuting: Boolean = false,
    val isEditorMode: Boolean = false,
    val editorWorkflow: Workflow? = null
)
```

**Archivo**: `WorkflowsEvent.kt`

```kotlin
package com.makoclaw.feature.workflows.presentation.state

sealed class WorkflowsEvent {
    data object LoadWorkflows : WorkflowsEvent()
    data object Refresh : WorkflowsEvent()
    data class CreateWorkflow(val name: String, val description: String) : WorkflowsEvent()
    data class UpdateWorkflow(val workflow: Workflow) : WorkflowsEvent()
    data class DeleteWorkflow(val id: String) : WorkflowsEvent()
    data class ExecuteWorkflow(val id: String) : WorkflowsEvent()
    data class SelectWorkflow(val id: String) : WorkflowsEvent()
    data class ViewLogs(val workflowId: String) : WorkflowsEvent()
    data object OpenEditor : WorkflowsEvent()
    data class CloseEditor : WorkflowsEvent()
    data class SaveEditor(val workflow: Workflow) : WorkflowsEvent()
}
```

**Archivo**: `WorkflowsEffect.kt`

```kotlin
package com.makoclaw.feature.workflows.presentation.state

sealed class WorkflowsEffect {
    data class ShowSnackbar(val message: String) : WorkflowsEffect()
    data class NavigateToEditor(val workflowId: String?) : WorkflowsEffect()
    data class NavigateToLogs(val workflowId: String) : WorkflowsEffect()
    data class ExecutionStarted(val workflowId: String) : WorkflowsEffect()
    data class ExecutionCompleted(val workflowId: String, val success: Boolean) : WorkflowsEffect()
}
```

**Archivo**: `WorkflowsViewModel.kt`

```kotlin
package com.makoclaw.feature.workflows.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.Workflow
import com.makoclaw.core.model.WorkflowExecutionLog
import com.makoclaw.feature.workflows.data.repository.WorkflowsRepository
import com.makoclaw.feature.workflows.presentation.state.WorkflowsEffect
import com.makoclaw.feature.workflows.presentation.state.WorkflowsEvent
import com.makoclaw.feature.workflows.presentation.state.WorkflowsUiState
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class WorkflowsViewModel @Inject constructor(
    private val repository: WorkflowsRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(WorkflowsUiState())
    val uiState: StateFlow<WorkflowsUiState> = _uiState.asStateFlow()

    private val _effects = MutableSharedFlow<WorkflowsEffect>()
    val effects: SharedFlow<WorkflowsEffect> = _effects.asSharedFlow()

    init {
        loadWorkflows()
    }

    fun onEvent(event: WorkflowsEvent) {
        when (event) {
            is WorkflowsEvent.LoadWorkflows -> loadWorkflows()
            is WorkflowsEvent.Refresh -> loadWorkflows()
            is WorkflowsEvent.CreateWorkflow -> createWorkflow(event.name, event.description)
            is WorkflowsEvent.UpdateWorkflow -> updateWorkflow(event.workflow)
            is WorkflowsEvent.DeleteWorkflow -> deleteWorkflow(event.id)
            is WorkflowsEvent.ExecuteWorkflow -> executeWorkflow(event.id)
            is WorkflowsEvent.SelectWorkflow -> selectWorkflow(event.id)
            is WorkflowsEvent.ViewLogs -> viewLogs(event.workflowId)
            is WorkflowsEvent.OpenEditor -> openEditor()
            is WorkflowsEvent.CloseEditor -> closeEditor()
            is WorkflowsEvent.SaveEditor -> saveEditor(event.workflow)
        }
    }

    private fun loadWorkflows() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            try {
                repository.getWorkflows()
                    .collect { workflows ->
                        _uiState.update {
                            it.copy(
                                isLoading = false,
                                workflows = workflows,
                                isEmpty = workflows.isEmpty()
                            )
                        }
                    }
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        error = "Failed to load workflows: ${e.message}"
                    )
                }
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error loading workflows"))
            }
        }
    }

    private fun createWorkflow(name: String, description: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }

            try {
                val workflow = Workflow(
                    id = "",
                    name = name,
                    description = description,
                    nodes = emptyList(),
                    edges = emptyList(),
                    status = "draft",
                    createdAt = System.currentTimeMillis(),
                    updatedAt = System.currentTimeMillis(),
                    lastExecutionAt = null
                )
                repository.createWorkflow(workflow)
                _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow created"))
                loadWorkflows()
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error creating workflow"))
            }
        }
    }

    private fun updateWorkflow(workflow: Workflow) {
        viewModelScope.launch {
            try {
                repository.updateWorkflow(workflow)
                _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow updated"))
                loadWorkflows()
            } catch (e: Exception) {
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error updating workflow"))
            }
        }
    }

    private fun deleteWorkflow(id: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }

            try {
                repository.deleteWorkflow(id)
                _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow deleted"))
                loadWorkflows()
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error deleting workflow"))
            }
        }
    }

    private fun executeWorkflow(id: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isExecuting = true) }

            try {
                _effects.emit(WorkflowsEffect.ExecutionStarted(id))

                repository.executeWorkflow(id).collect { log ->
                    _uiState.update { it.copy(executionLogs = it.executionLogs + log) }
                }

                _effects.emit(WorkflowsEffect.ExecutionCompleted(id, true))
                _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow executed successfully"))
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(isExecuting = false, error = "Execution failed: ${e.message}")
                }
                _effects.emit(WorkflowsEffect.ExecutionCompleted(id, false))
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error executing workflow"))
            }
        }
    }

    private fun selectWorkflow(id: String) {
        _uiState.update {
            it.copy(selectedWorkflow = it.workflows.find { w -> w.id == id })
        }
    }

    private fun viewLogs(workflowId: String) {
        viewModelScope.launch {
            try {
                val logs = repository.getExecutionLogs(workflowId)
                _uiState.update { it.copy(executionLogs = logs) }
                _effects.emit(WorkflowsEffect.NavigateToLogs(workflowId))
            } catch (e: Exception) {
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error loading logs"))
            }
        }
    }

    private fun openEditor() {
        _uiState.update {
            it.copy(
                isEditorMode = true,
                editorWorkflow = it.selectedWorkflow?.copy() ?: Workflow(
                    id = "",
                    name = "",
                    description = "",
                    nodes = emptyList(),
                    edges = emptyList(),
                    status = "draft",
                    createdAt = System.currentTimeMillis(),
                    updatedAt = System.currentTimeMillis(),
                    lastExecutionAt = null
                )
            )
        }
        _effects.emit(WorkflowsEffect.NavigateToEditor(it.selectedWorkflow?.id))
    }

    private fun closeEditor() {
        _uiState.update {
            it.copy(
                isEditorMode = false,
                editorWorkflow = null
            )
        }
    }

    private fun saveEditor(workflow: Workflow) {
        viewModelScope.launch {
            try {
                if (workflow.id.isBlank()) {
                    repository.createWorkflow(workflow)
                    _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow created"))
                } else {
                    repository.updateWorkflow(workflow)
                    _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow updated"))
                }
                closeEditor()
                loadWorkflows()
            } catch (e: Exception) {
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error saving workflow"))
            }
        }
    }
}
```

### Instrucciones

1. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/state/` si no existe
2. Crear los 4 archivos de state/event/effect/viewmodel
3. Verificar que no hay errores de compilación

---

### TASK-WORKFLOWS-002: Crear WorkflowsRepository

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/repository/WorkflowsRepository.kt`
**Paquete**: `com.makoclaw.feature.workflows.data.repository`

### Código Completo

**Archivo**: `WorkflowsRepository.kt`

```kotlin
package com.makoclaw.feature.workflows.data.repository

import com.makoclaw.core.model.Workflow
import com.makoclaw.core.model.WorkflowExecutionLog
import com.makoclaw.core.network.api.WorkflowsApi
import com.makoclaw.core.network.api.ApiResponse
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.catch
import javax.inject.Inject

interface WorkflowsRepository {
    fun getWorkflows(): Flow<List<Workflow>>
    fun createWorkflow(workflow: Workflow): Flow<ApiResponse>
    fun updateWorkflow(workflow: Workflow): Flow<ApiResponse>
    fun deleteWorkflow(id: String): Flow<ApiResponse>
    fun executeWorkflow(id: String): Flow<WorkflowExecutionLog>
    fun getExecutionLogs(workflowId: String): Flow<List<WorkflowExecutionLog>>
}

@ViewModelScoped
class WorkflowsRepositoryImpl @Inject constructor(
    private val api: WorkflowsApi,
    private val dao: WorkflowDao
) : WorkflowsRepository {

    override fun getWorkflows(): Flow<List<Workflow>> {
        return dao.getAll()
            .map { entities -> entities.map { it.toModel() } }
            .catch { e ->
                // Fallback to API if local cache fails
                api.getWorkflows()
                    .map { response -> response.workflows }
            }
    }

    override fun createWorkflow(workflow: Workflow): Flow<ApiResponse> {
        // Create in API then cache locally
        return api.createWorkflow(
            com.makoclaw.core.network.api.CreateWorkflowRequest(
                name = workflow.name,
                description = workflow.description,
                nodes = workflow.nodes,
                edges = workflow.edges
            )
        )
    }

    override fun updateWorkflow(workflow: Workflow): Flow<ApiResponse> {
        return api.updateWorkflow(
            workflow.id,
            com.makoclaw.core.network.api.UpdateWorkflowRequest(
                name = workflow.name,
                description = workflow.description,
                nodes = workflow.nodes,
                edges = workflow.edges
            )
        )
    }

    override fun deleteWorkflow(id: String): Flow<ApiResponse> {
        return api.deleteWorkflow(id)
    }

    override fun executeWorkflow(id: String): Flow<WorkflowExecutionLog> {
        return api.executeWorkflow(id)
    }

    override fun getExecutionLogs(workflowId: String): Flow<List<WorkflowExecutionLog>> {
        return api.getExecutionLogs(workflowId)
            .map { response -> response.logs }
    }
}
```

**Archivo**: `WorkflowsDao.kt` (crear en `feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/database/`)

```kotlin
package com.makoclaw.feature.workflows.data.database

import androidx.room.*
import kotlinx.coroutines.flow.Flow

@Entity(tableName = "workflows")
data class WorkflowEntity(
    @PrimaryKey val id: String,
    val name: String,
    val description: String,
    val nodes: String, // JSON string
    val edges: String, // JSON string
    val status: String,
    val createdAt: Long,
    val updatedAt: Long,
    val lastExecutionAt: Long?
)

@Dao
interface WorkflowDao {
    @Query("SELECT * FROM workflows ORDER BY updatedAt DESC")
    fun getAll(): Flow<List<WorkflowEntity>>

    @Query("SELECT * FROM workflows WHERE id = :id")
    suspend fun getById(id: String): WorkflowEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insert(workflow: WorkflowEntity)

    @Update
    suspend fun update(workflow: WorkflowEntity)

    @Delete
    suspend fun delete(workflow: WorkflowEntity)

    @Query("DELETE FROM workflows WHERE id = :id")
    suspend fun deleteById(id: String)
}

// Extension function to convert Entity to Model
fun WorkflowEntity.toModel(): com.makoclaw.core.model.Workflow {
    return com.makoclaw.core.model.Workflow(
        id = id,
        name = name,
        description = description,
        nodes = kotlinx.serialization.json.Json.decodeFromString(
            kotlinx.serialization.builtins.ListSerializer(com.makoclaw.core.model.WorkflowNode.serializer()),
            nodes
        ),
        edges = kotlinx.serialization.json.Json.decodeFromString(
            kotlinx.serialization.builtins.ListSerializer(com.makoclaw.core.model.WorkflowEdge.serializer()),
            edges
        ),
        status = status,
        createdAt = createdAt,
        updatedAt = updatedAt,
        lastExecutionAt = lastExecutionAt
    )
}
```

**Archivo**: `WorkflowsApi.kt` (crear en `core-network/src/main/java/com/makoclaw/core/network/api/` si no existe)

```kotlin
package com.makoclaw.core.network.api

import com.makoclaw.core.model.Workflow
import com.makoclaw.core.model.WorkflowExecutionLog
import retrofit2.http.*

data class WorkflowsResponse(
    val workflows: List<Workflow> = emptyList()
)

data class WorkflowResponse(
    val workflow: Workflow? = null
)

data class ExecutionLogsResponse(
    val logs: List<WorkflowExecutionLog> = emptyList()
)

data class CreateWorkflowRequest(
    val name: String,
    val description: String,
    val nodes: List<com.makoclaw.core.model.WorkflowNode>,
    val edges: List<com.makoclaw.core.model.WorkflowEdge>
)

data class UpdateWorkflowRequest(
    val name: String,
    val description: String,
    val nodes: List<com.makoclaw.core.model.WorkflowNode>,
    val edges: List<com.makoclaw.core.model.WorkflowEdge>
)

interface WorkflowsApi {
    @GET("workflows")
    suspend fun getWorkflows(): WorkflowsResponse

    @POST("workflows")
    suspend fun createWorkflow(@Body request: CreateWorkflowRequest): WorkflowResponse

    @PUT("workflows/{id}")
    suspend fun updateWorkflow(
        @Path("id") id: String,
        @Body request: UpdateWorkflowRequest
    ): WorkflowResponse

    @DELETE("workflows/{id}")
    suspend fun deleteWorkflow(@Path("id") id: String): ApiResponse

    @POST("workflows/{id}/execute")
    suspend fun executeWorkflow(@Path("id") id: String): WorkflowExecutionLog

    @GET("workflows/{id}/logs")
    suspend fun getExecutionLogs(@Path("id") id: String): ExecutionLogsResponse
}
```

### Instrucciones

1. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/repository/` si no existe
2. Crear el archivo `WorkflowsRepository.kt`
3. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/database/` si no existe
4. Crear el archivo `WorkflowsDao.kt`
5. Si `WorkflowsApi.kt` no existe en `core-network`, crearlo
6. Verificar que no hay errores de compilación

---

### TASK-WORKFLOWS-003: Crear WorkflowCanvasEditor 🚨

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/screen/WorkflowCanvasEditor.kt`
**Paquete**: `com.makoclaw.feature.workflows.presentation.screen`

### Código Completo

```kotlin
package com.makoclaw.feature.workflows.presentation.screen

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.graphics.nativeCanvas
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.unit.dp
import com.makoclaw.core.model.Workflow
import com.makoclaw.core.model.WorkflowNode
import com.makoclaw.core.model.WorkflowEdge
import com.makoclaw.core.model.NodeType
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlin.math.abs

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun WorkflowCanvasEditor(
    workflow: Workflow,
    onSave: (Workflow) -> Unit,
    onCancel: () -> Unit
) {
    var nodes by remember { mutableStateOf(workflow.nodes) }
    var edges by remember { mutableStateOf(workflow.edges) }
    var scale by remember { mutableStateOf(1f) }
    var offset by remember { mutableStateOf(Offset.Zero) }
    var selectedNodeId by remember { mutableStateOf<String?>(null) }
    var connectingFrom by remember { mutableStateOf<String?>(null) }
    var draggingNodeId by remember { mutableStateOf<String?>(null) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(workflow.name.ifBlank { "New Workflow" }) },
                navigationIcon = {
                    IconButton(onClick = onCancel) {
                        Icon(Icons.Default.Close, "Cancel")
                    }
                },
                actions = {
                    IconButton(
                        onClick = {
                            val updatedWorkflow = workflow.copy(
                                nodes = nodes,
                                edges = edges,
                                updatedAt = System.currentTimeMillis()
                            )
                            onSave(updatedWorkflow)
                        }
                    ) {
                        Icon(Icons.Default.Check, "Save")
                    }
                }
            )
        }
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            // Canvas for edges (drawn behind nodes)
            Canvas(
                modifier = Modifier
                    .fillMaxSize()
                    .pointerInput(Unit) {
                        // Zoom with two fingers (simplified - pinch to zoom)
                        detectDragGestures { change, dragAmount ->
                            if (draggingNodeId == null) {
                                offset += dragAmount
                            }
                            change.consume()
                        }
                    }
            ) {
                val canvasWidth = size.width
                val canvasHeight = size.height

                // Apply scale and offset
                drawContext.canvas.save()
                drawContext.canvas.translate(offset.x, offset.y)
                drawContext.canvas.scale(scale, scale)

                // Draw edges
                edges.forEach { edge ->
                    val sourceNode = nodes.find { it.id == edge.sourceNodeId }
                    val targetNode = nodes.find { it.id == edge.targetNodeId }

                    if (sourceNode != null && targetNode != null) {
                        val sourceOffset = Offset(
                            sourceNode.x + sourceNode.width / 2,
                            sourceNode.y + sourceNode.height / 2
                        )
                        val targetOffset = Offset(
                            targetNode.x + targetNode.width / 2,
                            targetNode.y + targetNode.height / 2
                        )

                        // Draw bezier curve
                        val path = Path().apply {
                            moveTo(sourceOffset.x, sourceOffset.y)
                            val controlPoint1 = Offset(
                                sourceOffset.x + (targetOffset.x - sourceOffset.x) * 0.5f,
                                sourceOffset.y
                            )
                            val controlPoint2 = Offset(
                                sourceOffset.x + (targetOffset.x - sourceOffset.x) * 0.5f,
                                targetOffset.y
                            )
                            cubicTo(
                                controlPoint1.x, controlPoint1.y,
                                controlPoint2.x, controlPoint2.y,
                                targetOffset.x, targetOffset.y
                            )
                        }

                        drawPath(
                            path = path,
                            color = Color.Gray,
                            style = Stroke(width = 2f)
                        )
                    }
                }

                // Draw connecting line if dragging connection
                if (connectingFrom != null) {
                    val sourceNode = nodes.find { it.id == connectingFrom }
                    if (sourceNode != null) {
                        val sourceOffset = Offset(
                            sourceNode.x + sourceNode.width / 2,
                            sourceNode.y + sourceNode.height / 2
                        )
                        // Draw line from source to current mouse position
                        // (simplified - would need pointer position tracking)
                    }
                }

                drawContext.canvas.restore()
            }

            // Nodes overlay
            nodes.forEach { node ->
                WorkflowNode(
                    node = node,
                    isSelected = node.id == selectedNodeId,
                    isConnecting = connectingFrom == node.id,
                    onTap = { selectedNodeId = node.id },
                    onLongPress = {
                        connectingFrom = if (connectingFrom == node.id) null else node.id
                    },
                    onDrag = { dragAmount ->
                        if (draggingNodeId == node.id) {
                            nodes = nodes.map {
                                if (it.id == node.id) {
                                    it.copy(
                                        x = it.x + dragAmount.x / scale,
                                        y = it.y + dragAmount.y / scale
                                    )
                                } else {
                                    it
                                }
                            }
                        }
                    },
                    onDragStart = {
                        draggingNodeId = node.id
                        selectedNodeId = node.id
                    },
                    onDragEnd = {
                        draggingNodeId = null
                    },
                    onDelete = {
                        nodes = nodes.filter { it.id != node.id }
                        edges = edges.filter {
                            it.sourceNodeId != node.id && it.targetNodeId != node.id
                        }
                        selectedNodeId = null
                    },
                    modifier = Modifier
                        .offset {
                            androidx.compose.ui.unit.IntOffset(
                                (node.x * scale + offset.x).toInt(),
                                (node.y * scale + offset.y).toInt()
                            )
                        }
                )
            }

            // Floating Action Button to add node
            FloatingActionButton(
                onClick = {
                    val newNode = WorkflowNode(
                        id = "node-${System.currentTimeMillis()}",
                        type = NodeType.TASK,
                        name = "New Task",
                        x = (-offset.x + 100.dp.toPx()) / scale,
                        y = (-offset.y + 100.dp.toPx()) / scale,
                        width = 150f,
                        height = 80f,
                        config = emptyMap()
                    )
                    nodes = nodes + newNode
                },
                modifier = Modifier
                    .align(Alignment.BottomEnd)
                    .padding(16.dp)
            ) {
                Icon(Icons.Default.Add, "Add Node")
            }
        }
    }
}

@Composable
fun WorkflowNode(
    node: WorkflowNode,
    isSelected: Boolean,
    isConnecting: Boolean,
    onTap: () -> Unit,
    onLongPress: () -> Unit,
    onDrag: (Offset) -> Unit,
    onDragStart: () -> Unit,
    onDragEnd: () -> Unit,
    onDelete: () -> Unit,
    modifier: Modifier = Modifier
) {
    var isDragging by remember { mutableStateOf(false) }

    Card(
        modifier = modifier
            .width(node.width.dp)
            .height(node.height.dp)
            .pointerInput(Unit) {
                detectTapGestures(
                    onTap = { onTap() },
                    onLongPress = { onLongPress() }
                )
            }
            .pointerInput(Unit) {
                detectDragGestures(
                    onDragStart = { onDragStart() },
                    onDrag = { change, dragAmount ->
                        onDrag(dragAmount)
                        change.consume()
                    },
                    onDragEnd = { onDragEnd() }
                )
            },
        colors = CardDefaults.cardColors(
            containerColor = when {
                isSelected -> MaterialTheme.colorScheme.primaryContainer
                isConnecting -> MaterialTheme.colorScheme.secondaryContainer
                else -> MaterialTheme.colorScheme.surface
            }
        ),
        border = if (isSelected) {
            androidx.compose.foundation.BorderStroke(
                2.dp,
                MaterialTheme.colorScheme.primary
            )
        } else {
            null
        }
    ) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(8.dp),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = node.name,
                style = MaterialTheme.typography.bodyMedium,
                color = when {
                    isSelected -> MaterialTheme.colorScheme.onPrimaryContainer
                    isConnecting -> MaterialTheme.colorScheme.onSecondaryContainer
                    else -> MaterialTheme.colorScheme.onSurface
                }
            )
        }

        // Delete button (small, visible when selected)
        if (isSelected) {
            IconButton(
                onClick = onDelete,
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .size(24.dp)
                    .padding(2.dp)
            ) {
                Icon(
                    Icons.Default.Delete,
                    "Delete",
                    tint = MaterialTheme.colorScheme.error,
                    modifier = Modifier.size(16.dp)
                )
            }
        }
    }
}
```

### Instrucciones

1. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/screen/` si no existe
2. Crear el archivo `WorkflowCanvasEditor.kt` y pegar todo el código
3. Verificar que no hay errores de compilación

**NOTA IMPORTANTE**: Este es un canvas editor simplificado. Para una implementación completa de producción, considera:
- Usar una librería de canvas (ej. PanZoomImage)
- Implementar pinch-to-zoom con proper gestures
- Guardar/restaurar el estado del zoom/pan
- Mejorar la selección de nodos (hit detection)
- Implementar proper edge connection UI

---

### TASK-WORKFLOWS-004: Crear WorkflowsScreen UI

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/screen/WorkflowsScreen.kt`
**Paquete**: `com.makoclaw.feature.workflows.presentation.screen`

### Código Completo

```kotlin
package com.makoclaw.feature.workflows.presentation.screen

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.workflows.presentation.viewmodel.WorkflowsViewModel
import com.makoclaw.feature.workflows.presentation.state.WorkflowsEvent
import com.makoclaw.feature.workflows.presentation.component.WorkflowCard
import com.makoclaw.feature.workflows.presentation.component.ExecutionLogsModal

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun WorkflowsScreen(
    viewModel: WorkflowsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val effects by viewModel.effects.collectAsState(initialValue = null)
    val scope = rememberCoroutineScope()

    var showLogs by remember { mutableStateOf(false) }
    var showEditor by remember { mutableStateOf(false) }
    var selectedWorkflowId by remember { mutableStateOf<String?>(null) }

    // Snackbar host
    val snackbarHostState = SnackbarHostState()

    // Handle effects
    LaunchedEffect(Unit) {
        effects?.let { effect ->
            when (effect) {
                is WorkflowsEffect.ShowSnackbar -> {
                    scope.launch {
                        snackbarHostState.showSnackbar(
                            message = effect.message,
                            duration = SnackbarDuration.Short
                        )
                    }
                }
                is WorkflowsEffect.NavigateToEditor -> {
                    showEditor = true
                    selectedWorkflowId = effect.workflowId
                }
                is WorkflowsEffect.NavigateToLogs -> {
                    showLogs = true
                    selectedWorkflowId = effect.workflowId
                }
                is WorkflowsEffect.ExecutionStarted -> {
                    // Show execution indicator
                }
                is WorkflowsEffect.ExecutionCompleted -> {
                    // Show result
                }
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Workflows") },
                actions = {
                    IconButton(onClick = { viewModel.onEvent(WorkflowsEvent.Refresh) }) {
                        Icon(Icons.Default.Refresh, "Refresh")
                    }
                }
            )
        },
        snackbarHost = { SnackbarHost(snackbarHostState) },
        floatingActionButton = {
            FloatingActionButton(
                onClick = { viewModel.onEvent(WorkflowsEvent.OpenEditor) }
            ) {
                Icon(Icons.Default.Add, "Add Workflow")
            }
        }
    ) { padding ->
        if (showEditor && uiState.editorWorkflow != null) {
            WorkflowCanvasEditor(
                workflow = uiState.editorWorkflow!!,
                onSave = { viewModel.onEvent(WorkflowsEvent.SaveEditor(it)) },
                onCancel = { viewModel.onEvent(WorkflowsEvent.CloseEditor) }
            )
        } else {
            if (uiState.isLoading) {
                LoadingScreen()
            } else if (uiState.error != null) {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    Text(
                        text = "Error: ${uiState.error}",
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.error
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    Button(onClick = { viewModel.onEvent(WorkflowsEvent.Refresh) }) {
                        Text("Retry")
                    }
                }
            } else if (uiState.isEmpty) {
                EmptyState(
                    title = "No workflows yet",
                    message = "Create your first workflow to automate tasks",
                    actionLabel = "Create Workflow"
                ) {
                    viewModel.onEvent(WorkflowsEvent.OpenEditor)
                }
            } else {
                LazyColumn(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    item { Spacer(modifier = Modifier.height(8.dp)) }

                    items(uiState.workflows) { workflow ->
                        WorkflowCard(
                            workflow = workflow,
                            onClick = { viewModel.onEvent(WorkflowsEvent.SelectWorkflow(workflow.id)) },
                            onExecute = { viewModel.onEvent(WorkflowsEvent.ExecuteWorkflow(workflow.id)) },
                            onEdit = { viewModel.onEvent(WorkflowsEvent.OpenEditor) },
                            onDelete = { viewModel.onEvent(WorkflowsEvent.DeleteWorkflow(workflow.id)) },
                            onViewLogs = { viewModel.onEvent(WorkflowsEvent.ViewLogs(workflow.id)) }
                        )
                    }

                    item { Spacer(modifier = Modifier.height(80.dp)) }
                }
            }
        }
    }

    // Execution logs modal
    if (showLogs && selectedWorkflowId != null) {
        ExecutionLogsModal(
            workflowId = selectedWorkflowId!!,
            logs = uiState.executionLogs,
            onDismiss = { showLogs = false }
        )
    }
}
```

### Instrucciones

1. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/screen/` si no existe
2. Crear el archivo `WorkflowsScreen.kt` y pegar todo el código
3. Verificar que no hay errores de compilación

---

### TASK-WORKFLOWS-005: Crear WorkflowCard y ExecutionLogsModal

### Ubicación

**Archivo 1**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/component/WorkflowCard.kt`
**Archivo 2**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/component/ExecutionLogsModal.kt`
**Paquete**: `com.makoclaw.feature.workflows.presentation.component`

### Código Completo - WorkflowCard.kt

```kotlin
package com.makoclaw.feature.workflows.presentation.component

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Visibility
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.makoclaw.core.model.Workflow
import java.text.SimpleDateFormat
import java.util.*

@Composable
fun WorkflowCard(
    workflow: Workflow,
    onClick: () -> Unit,
    onExecute: () -> Unit,
    onEdit: () -> Unit,
    onDelete: () -> Unit,
    onViewLogs: () -> Unit
) {
    var showMenu by remember { mutableStateOf(false) }

    Card(
        onClick = onClick,
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            // Header: Name + Menu
            Row(
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = workflow.name,
                        style = MaterialTheme.typography.titleMedium
                    )
                    Text(
                        text = workflow.status.uppercase(),
                        style = MaterialTheme.typography.labelSmall,
                        color = when (workflow.status.lowercase()) {
                            "active" -> MaterialTheme.colorScheme.primary
                            "draft" -> MaterialTheme.colorScheme.secondary
                            else -> MaterialTheme.colorScheme.onSurfaceVariant
                        }
                    )
                }

                Box {
                    IconButton(onClick = { showMenu = true }) {
                        Icon(Icons.Default.Edit, "Edit")
                    }
                    DropdownMenu(
                        expanded = showMenu,
                        onDismissRequest = { showMenu = false }
                    ) {
                        DropdownMenuItem(
                            text = { Text("Edit") },
                            leadingIcon = { Icon(Icons.Default.Edit, "Edit") },
                            onClick = {
                                showMenu = false
                                onEdit()
                            }
                        )
                        DropdownMenuItem(
                            text = { Text("Execute") },
                            leadingIcon = { Icon(Icons.Default.PlayArrow, "Execute") },
                            onClick = {
                                showMenu = false
                                onExecute()
                            }
                        )
                        DropdownMenuItem(
                            text = { Text("View Logs") },
                            leadingIcon = { Icon(Icons.Default.Visibility, "View Logs") },
                            onClick = {
                                showMenu = false
                                onViewLogs()
                            }
                        )
                        HorizontalDivider()
                        DropdownMenuItem(
                            text = { Text("Delete", color = MaterialTheme.colorScheme.error) },
                            leadingIcon = {
                                Icon(
                                    Icons.Default.Delete,
                                    "Delete",
                                    tint = MaterialTheme.colorScheme.error
                                )
                            },
                            onClick = {
                                showMenu = false
                                onDelete()
                            }
                        )
                    }
                }
            }

            // Description
            if (workflow.description.isNotBlank()) {
                Text(
                    text = workflow.description,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 2
                )
            }

            // Footer: Metadata
            Row(
                horizontalArrangement = Arrangement.spacedBy(16.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "${workflow.nodes.size} nodes",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Text(
                    text = formatDate(workflow.createdAt),
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                workflow.lastExecutionAt?.let {
                    Text(
                        text = "Last run: ${formatDate(it)}",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }
    }
}

private fun formatDate(timestamp: Long): String {
    val sdf = SimpleDateFormat("MMM dd, yyyy", Locale.US)
    return sdf.format(Date(timestamp))
}
```

### Código Completo - ExecutionLogsModal.kt

```kotlin
package com.makoclaw.feature.workflows.presentation.component

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.makoclaw.core.model.WorkflowExecutionLog

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ExecutionLogsModal(
    workflowId: String,
    logs: List<WorkflowExecutionLog>,
    onDismiss: () -> Unit
) {
    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = rememberModalBottomSheetState()
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .height(500.dp)
                .padding(16.dp)
        ) {
            // Header
            Row(
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Execution Logs",
                    style = MaterialTheme.typography.titleLarge
                )
                IconButton(onClick = onDismiss) {
                    Icon(Icons.Default.Close, "Close")
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Logs list
            if (logs.isEmpty()) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = androidx.compose.ui.Alignment.Center
                ) {
                    Text(
                        text = "No execution logs yet",
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            } else {
                LazyColumn(
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    items(logs) { log ->
                        LogEntryItem(log = log)
                    }
                }
            }
        }
    }
}

@Composable
fun LogEntryItem(log: WorkflowExecutionLog) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = when (log.status.lowercase()) {
                "success" -> MaterialTheme.colorScheme.primaryContainer
                "error" -> MaterialTheme.colorScheme.errorContainer
                else -> MaterialTheme.colorScheme.surfaceVariant
            }
        )
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp)
        ) {
            Text(
                text = log.status.uppercase(),
                style = MaterialTheme.typography.labelSmall,
                color = when (log.status.lowercase()) {
                    "success" -> MaterialTheme.colorScheme.onPrimaryContainer
                    "error" -> MaterialTheme.colorScheme.onErrorContainer
                    else -> MaterialTheme.colorScheme.onSurfaceVariant
                }
            )
            Text(
                text = log.message,
                style = MaterialTheme.typography.bodyMedium
            )
            Text(
                text = formatDate(log.timestamp),
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

private fun formatDate(timestamp: Long): String {
    val sdf = SimpleDateFormat("MMM dd, HH:mm:ss", java.util.Locale.US)
    return sdf.format(java.util.Date(timestamp))
}
```

### Instrucciones

1. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/component/` si no existe
2. Crear el archivo `WorkflowCard.kt` y pegar el primer bloque de código
3. Crear el archivo `ExecutionLogsModal.kt` y pegar el segundo bloque de código
4. Verificar que no hay errores de compilación

---

## FEATURE-METRICS (4 tareas)

### TASK-METRICS-001: Crear MetricsViewModel

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/state/MetricsUiState.kt`
**Paquete**: `com.makoclaw.feature.metrics.presentation.state`

### Código Completo

**Archivo**: `MetricsUiState.kt`

```kotlin
package com.makoclaw.feature.metrics.presentation.state

import com.makoclaw.core.model.MetricsData
import com.makoclaw.core.model.TimeRange

data class MetricsUiState(
    val isLoading: Boolean = false,
    val metrics: MetricsData? = null,
    val error: String? = null,
    val timeRange: TimeRange = TimeRange.WEEK,
    val selectedAgent: String? = null,
    val availableAgents: List<String> = emptyList(),
    val isExporting: Boolean = false
)
```

**Archivo**: `MetricsEvent.kt`

```kotlin
package com.makoclaw.feature.metrics.presentation.state

sealed class MetricsEvent {
    data object LoadMetrics : MetricsEvent()
    data object Refresh : MetricsEvent()
    data class SetTimeRange(val range: TimeRange) : MetricsEvent()
    data class SetAgentFilter(val agentId: String?) : MetricsEvent()
    data object ExportMetrics : MetricsEvent()
}
```

**Archivo**: `MetricsEffect.kt`

```kotlin
package com.makoclaw.feature.metrics.presentation.state

sealed class MetricsEffect {
    data class ShowSnackbar(val message: String) : MetricsEffect()
    data class ExportComplete(val filePath: String) : MetricsEffect()
}
```

**Archivo**: `MetricsViewModel.kt`

```kotlin
package com.makoclaw.feature.metrics.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.MetricsData
import com.makoclaw.core.model.TimeRange
import com.makoclaw.feature.metrics.data.repository.MetricsRepository
import com.makoclaw.feature.metrics.presentation.state.MetricsEffect
import com.makoclaw.feature.metrics.presentation.state.MetricsEvent
import com.makoclaw.feature.metrics.presentation.state.MetricsUiState
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class MetricsViewModel @Inject constructor(
    private val repository: MetricsRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(MetricsUiState())
    val uiState: StateFlow<MetricsUiState> = _uiState.asStateFlow()

    private val _effects = MutableSharedFlow<MetricsEffect>()
    val effects: SharedFlow<MetricsEffect> = _effects.asSharedFlow()

    init {
        loadMetrics()
    }

    fun onEvent(event: MetricsEvent) {
        when (event) {
            is MetricsEvent.LoadMetrics -> loadMetrics()
            is MetricsEvent.Refresh -> loadMetrics()
            is MetricsEvent.SetTimeRange -> setTimeRange(event.range)
            is MetricsEvent.SetAgentFilter -> setAgentFilter(event.agentId)
            is MetricsEvent.ExportMetrics -> exportMetrics()
        }
    }

    private fun loadMetrics() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            try {
                repository.getMetrics(_uiState.value.timeRange, _uiState.value.selectedAgent)
                    .collect { metrics ->
                        _uiState.update {
                            it.copy(
                                isLoading = false,
                                metrics = metrics
                            )
                        }
                    }
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        error = "Failed to load metrics: ${e.message}"
                    )
                }
                _effects.emit(MetricsEffect.ShowSnackbar("Error loading metrics"))
            }
        }
    }

    private fun setTimeRange(range: TimeRange) {
        _uiState.update { it.copy(timeRange = range) }
        loadMetrics()
    }

    private fun setAgentFilter(agentId: String?) {
        _uiState.update { it.copy(selectedAgent = agentId) }
        loadMetrics()
    }

    private fun exportMetrics() {
        viewModelScope.launch {
            _uiState.update { it.copy(isExporting = true) }

            try {
                val result = repository.exportMetrics(
                    _uiState.value.timeRange,
                    _uiState.value.selectedAgent
                )
                _effects.emit(MetricsEffect.ExportComplete(result.filePath))
                _effects.emit(MetricsEffect.ShowSnackbar("Metrics exported to ${result.filePath}"))
            } catch (e: Exception) {
                _effects.emit(MetricsEffect.ShowSnackbar("Error exporting metrics"))
            } finally {
                _uiState.update { it.copy(isExporting = false) }
            }
        }
    }
}
```

### Instrucciones

1. Crear el directorio `feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/state/` si no existe
2. Crear los 4 archivos de state/event/effect/viewmodel
3. Verificar que no hay errores de compilación

---

### TASK-METRICS-002: Crear MetricsRepository

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-metrics/src/main/java/com/makoclaw/feature/metrics/data/repository/MetricsRepository.kt`
**Paquete**: `com.makoclaw.feature.metrics.data.repository`

### Código Completo

```kotlin
package com.makoclaw.feature.metrics.data.repository

import com.makoclaw.core.model.MetricsData
import com.makoclaw.core.model.TimeRange
import com.makoclaw.core.network.api.MetricsApi
import com.makoclaw.core.network.api.ExportResult
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.flowOf
import javax.inject.Inject

interface MetricsRepository {
    fun getMetrics(timeRange: TimeRange, agentId: String?): Flow<MetricsData>
    fun exportMetrics(timeRange: TimeRange, agentId: String?): Flow<ExportResult>
}

@ViewModelScoped
class MetricsRepositoryImpl @Inject constructor(
    private val api: MetricsApi,
    private val dao: MetricsDao
) : MetricsRepository {

    override fun getMetrics(timeRange: TimeRange, agentId: String?): Flow<MetricsData> {
        return api.getMetrics(timeRange, agentId)
            .catch { e ->
                // Fallback to cache if API fails
                dao.getByTimeRange(timeRange)
                    ?: throw e
            }
    }

    override fun exportMetrics(timeRange: TimeRange, agentId: String?): Flow<ExportResult> {
        return api.exportMetrics(timeRange, agentId)
    }
}
```

**Archivo**: `MetricsDao.kt` (crear en `feature-metrics/src/main/java/com/makoclaw/feature/metrics/data/database/`)

```kotlin
package com.makoclaw.feature.metrics.data.database

import androidx.room.*
import kotlinx.coroutines.flow.Flow

@Entity(tableName = "metrics")
data class MetricsEntity(
    @PrimaryKey val id: String, // Format: "{timeRange}_{agentId}"
    val timeRange: String,
    val agentId: String?,
    val data: String, // JSON string of MetricsData
    val cachedAt: Long
)

@Dao
interface MetricsDao {
    @Query("SELECT * FROM metrics WHERE timeRange = :timeRange ORDER BY cachedAt DESC LIMIT 1")
    suspend fun getByTimeRange(timeRange: String): MetricsEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insert(metrics: MetricsEntity)

    @Delete
    suspend fun delete(metrics: MetricsEntity)
}
```

**Archivo**: `MetricsApi.kt` (crear en `core-network/src/main/java/com/makoclaw/core/network/api/` si no existe)

```kotlin
package com.makoclaw.core.network.api

import com.makoclaw.core.model.MetricsData
import com.makoclaw.core.model.TimeRange
import retrofit2.http.*

data class MetricsResponse(
    val metrics: MetricsData? = null
)

data class ExportResult(
    val filePath: String,
    val fileSize: Long,
    val format: String
)

data class ExportRequest(
    val timeRange: String,
    val agentId: String?,
    val format: String = "csv"
)

interface MetricsApi {
    @GET("metrics")
    suspend fun getMetrics(
        @Query("timeRange") timeRange: TimeRange,
        @Query("agentId") agentId: String?
    ): MetricsResponse

    @POST("metrics/export")
    suspend fun exportMetrics(
        @Body request: ExportRequest
    ): ExportResult
}
```

### Instrucciones

1. Crear el directorio `feature-metrics/src/main/java/com/makoclaw/feature/metrics/data/repository/` si no existe
2. Crear el archivo `MetricsRepository.kt`
3. Crear el directorio `feature-metrics/src/main/java/com/makoclaw/feature/metrics/data/database/` si no existe
4. Crear el archivo `MetricsDao.kt`
5. Si `MetricsApi.kt` no existe en `core-network`, crearlo
6. Verificar que no hay errores de compilación

---

### TASK-METRICS-003: Crear Charts con Vico

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/component/MetricsCharts.kt`
**Paquete**: `com.makoclaw.feature.metrics.presentation.component`

### Código Completo

```kotlin
package com.makoclaw.feature.metrics.presentation.component

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import com.patrykandpatrick.vico.compose.axis.horizontal.rememberBottomAxis
import com.patrykandpatrick.vico.compose.axis.vertical.rememberStartAxis
import com.patrykandpatrick.vico.compose.chart.CartesianChartHost
import com.patrykandpatrick.vico.compose.chart.column.columnChart
import com.patrykandpatrick.vico.compose.chart.line.lineChart
import com.patrykandpatrick.vico.compose.component.shapeComponent
import com.patrykandpatrick.vico.core.chart.column.ColumnChart
import com.patrykandpatrick.vico.core.chart.line.LineChart
import com.patrykandpatrick.vico.core.component.shape.DashedShape
import com.patrykandpatrick.vico.core.component.shape.ShapeComponent
import com.patrykandpatrick.vico.core.component.shape Shapes
import com.patrykandpatrick.vico.core.data.DataPoint
import com.patrykandpatrick.vico.core.data.column.ColumnData
import com.patrykandpatrick.vico.core.data.line.LineData
import com.patrykandpatrick.vico.core.entry.entriesOf
import com.patrykandpatrick.vico.core.entry.composed.ModelEntryCollection
import com.patrykandpatrick.vico.core.entry.FloatEntry

@Composable
fun MetricsLineChart(
    title: String,
    data: List<FloatEntry>,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier.fillMaxWidth()
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Text(
                text = title,
                style = MaterialTheme.typography.titleMedium
            )

            CartesianChartHost(
                chart = lineChart(),
                model = LineData.entriesOf(data),
                startAxis = rememberStartAxis(),
                bottomAxis = rememberBottomAxis(),
                modifier = Modifier
                    .fillMaxWidth()
                    .height(200.dp)
            )
        }
    }
}

@Composable
fun MetricsBarChart(
    title: String,
    data: List<FloatEntry>,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier.fillMaxWidth()
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Text(
                text = title,
                style = MaterialTheme.typography.titleMedium
            )

            CartesianChartHost(
                chart = columnChart(),
                model = ColumnData.entriesOf(data),
                startAxis = rememberStartAxis(),
                bottomAxis = rememberBottomAxis(),
                modifier = Modifier
                    .fillMaxWidth()
                    .height(200.dp)
            )
        }
    }
}

@Composable
fun MetricsDonutChart(
    title: String,
    data: Map<String, Float>,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier.fillMaxWidth()
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Text(
                text = title,
                style = MaterialTheme.typography.titleMedium
            )

            // Simplified donut chart using CircularProgressIndicator
            Row(
                horizontalArrangement = Arrangement.SpaceAround,
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.fillMaxWidth()
            ) {
                data.entries.forEachIndexed { index, (label, value) ->
                    Column(
                        horizontalAlignment = androidx.compose.foundation.layout.Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(4.dp)
                    ) {
                        CircularProgressIndicator(
                            progress = { value / 100f },
                            modifier = Modifier.size(60.dp),
                            color = MaterialTheme.colorScheme.tertiaryContainer,
                            strokeWidth = 8.dp
                        )
                        Text(
                            text = "${value.toInt()}%",
                            style = MaterialTheme.typography.bodySmall
                        )
                        Text(
                            text = label,
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }
            }
        }
    }
}
```

### Instrucciones

1. Crear el directorio `feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/component/` si no existe
2. Crear el archivo `MetricsCharts.kt` y pegar todo el código
3. **IMPORTANTE**: Agregar la dependencia de Vico en `build.gradle.kts`:
```kotlin
implementation("com.patrykandpatrick.vico:compose:1.14.0")
implementation("com.patrykandpatrick.vico:compose-m3:1.14.0")
```
4. Verificar que no hay errores de compilación

---

### TASK-METRICS-004: Crear MetricsScreen UI

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/screen/MetricsScreen.kt`
**Paquete**: `com.makoclaw.feature.metrics.presentation.screen`

### Código Completo

```kotlin
package com.makoclaw.feature.metrics.presentation.screen

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Download
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.model.TimeRange
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.metrics.presentation.viewmodel.MetricsViewModel
import com.makoclaw.feature.metrics.presentation.state.MetricsEvent
import com.makoclaw.feature.metrics.presentation.component.MetricsLineChart
import com.makoclaw.feature.metrics.presentation.component.MetricsBarChart
import com.makoclaw.feature.metrics.presentation.component.MetricsDonutChart

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MetricsScreen(
    viewModel: MetricsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val effects by viewModel.effects.collectAsState(initialValue = null)
    val scope = rememberCoroutineScope()

    // Snackbar host
    val snackbarHostState = SnackbarHostState()

    // Handle effects
    LaunchedEffect(Unit) {
        effects?.let { effect ->
            when (effect) {
                is MetricsEffect.ShowSnackbar -> {
                    scope.launch {
                        snackbarHostState.showSnackbar(
                            message = effect.message,
                            duration = SnackbarDuration.Short
                        )
                    }
                }
                is MetricsEffect.ExportComplete -> {
                    scope.launch {
                        snackbarHostState.showSnackbar(
                            message = "Exported to ${effect.filePath}",
                            duration = SnackbarDuration.Long
                        )
                    }
                }
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Metrics") },
                actions = {
                    IconButton(onClick = { viewModel.onEvent(MetricsEvent.Refresh) }) {
                        Icon(Icons.Default.Refresh, "Refresh")
                    }
                    IconButton(onClick = { viewModel.onEvent(MetricsEvent.ExportMetrics) }) {
                        Icon(Icons.Default.Download, "Export")
                    }
                }
            )
        },
        snackbarHost = { SnackbarHost(snackbarHostState) }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            // Filters
            FilterSection(
                timeRange = uiState.timeRange,
                selectedAgent = uiState.selectedAgent,
                availableAgents = uiState.availableAgents,
                onTimeRangeChange = { viewModel.onEvent(MetricsEvent.SetTimeRange(it)) },
                onAgentChange = { viewModel.onEvent(MetricsEvent.SetAgentFilter(it)) }
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Content
            if (uiState.isLoading) {
                LoadingScreen()
            } else if (uiState.error != null) {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    horizontalAlignment = androidx.compose.foundation.layout.Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    Text(
                        text = "Error: ${uiState.error}",
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.error
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    Button(onClick = { viewModel.onEvent(MetricsEvent.Refresh) }) {
                        Text("Retry")
                    }
                }
            } else if (uiState.metrics == null) {
                EmptyState(
                    title = "No metrics available",
                    message = "Start using agents to generate metrics",
                    actionLabel = "Refresh"
                ) {
                    viewModel.onEvent(MetricsEvent.Refresh)
                }
            } else {
                MetricsContent(
                    metrics = uiState.metrics!!,
                    modifier = Modifier.fillMaxSize()
                )
            }
        }
    }
}

@Composable
fun FilterSection(
    timeRange: TimeRange,
    selectedAgent: String?,
    availableAgents: List<String>,
    onTimeRangeChange: (TimeRange) -> Unit,
    onAgentChange: (String?) -> Unit
) {
    Row(
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp)
    ) {
        // Time range filter
        FilterChip(
            selected = timeRange == TimeRange.DAY,
            onClick = { onTimeRangeChange(TimeRange.DAY) },
            label = { Text("Day") }
        )
        FilterChip(
            selected = timeRange == TimeRange.WEEK,
            onClick = { onTimeRangeChange(TimeRange.WEEK) },
            label = { Text("Week") }
        )
        FilterChip(
            selected = timeRange == TimeRange.MONTH,
            onClick = { onTimeRangeChange(TimeRange.MONTH) },
            label = { Text("Month") }
        )

        Spacer(modifier = Modifier.weight(1f))

        // Agent filter (simplified as dropdown)
        if (availableAgents.isNotEmpty()) {
            var expanded by remember { mutableStateOf(false) }

            Box {
                Button(
                    onClick = { expanded = true },
                    colors = ButtonDefaults.buttonColors(
                        containerColor = MaterialTheme.colorScheme.secondaryContainer
                    )
                ) {
                    Text(selectedAgent ?: "All Agents")
                }

                DropdownMenu(
                    expanded = expanded,
                    onDismissRequest = { expanded = false }
                ) {
                    DropdownMenuItem(
                        text = { Text("All Agents") },
                        onClick = {
                            onAgentChange(null)
                            expanded = false
                        }
                    )
                    availableAgents.forEach { agent ->
                        DropdownMenuItem(
                            text = { Text(agent) },
                            onClick = {
                                onAgentChange(agent)
                                expanded = false
                            }
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun MetricsContent(
    metrics: com.makoclaw.core.model.MetricsData,
    modifier: Modifier = Modifier
) {
    LazyColumn(
        modifier = modifier,
        verticalArrangement = Arrangement.spacedBy(16.dp),
        contentPadding = PaddingValues(16.dp)
    ) {
        // Line chart: Execution count over time
        item {
            MetricsLineChart(
                title = "Executions Over Time",
                data = metrics.executionsOverTime.mapIndexed { index, value ->
                    com.patrykandpatrick.vico.core.entry.FloatEntry(index.toFloat(), value)
                }
            )
        }

        // Bar chart: Agent usage
        item {
            MetricsBarChart(
                title = "Agent Usage",
                data = metrics.agentUsage.mapIndexed { index, value ->
                    com.patrykandpatrick.vico.core.entry.FloatEntry(index.toFloat(), value)
                }
            )
        }

        // Donut chart: Status distribution
        item {
            MetricsDonutChart(
                title = "Status Distribution",
                data = mapOf(
                    "Success" to metrics.successRate,
                    "Failed" to (100 - metrics.successRate)
                )
            )
        }

        // Summary cards
        item {
            LazyVerticalGrid(
                columns = GridCells.Fixed(2),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                item {
                    SummaryCard(title = "Total Executions", value = metrics.totalExecutions.toString())
                }
                item {
                    SummaryCard(title = "Avg Duration", value = "${metrics.avgDuration}s")
                }
                item {
                    SummaryCard(title = "Success Rate", value = "${metrics.successRate}%")
                }
                item {
                    SummaryCard(title = "Active Agents", value = metrics.activeAgents.toString())
                }
            }
        }

        item { Spacer(modifier = Modifier.height(16.dp)) }
    }
}

@Composable
fun SummaryCard(title: String, value: String) {
    Card(
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp)
        ) {
            Text(
                text = title,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Text(
                text = value,
                style = MaterialTheme.typography.headlineSmall
            )
        }
    }
}
```

### Instrucciones

1. Crear el directorio `feature-metrics/src/main/java/com/makoclaw/feature/metrics/presentation/screen/` si no existe
2. Crear el archivo `MetricsScreen.kt` y pegar todo el código
3. Verificar que no hay errores de compilación

---

## FEATURE-CRON (4 tareas)

### TASK-CRON-001: Crear CronViewModel

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/state/CronUiState.kt`
**Paquete**: `com.makoclaw.feature.cron.presentation.state`

### Código Completo

**Archivo**: `CronUiState.kt`

```kotlin
package com.makoclaw.feature.cron.presentation.state

import com.makoclaw.core.model.CronJob

data class CronUiState(
    val isLoading: Boolean = false,
    val jobs: List<CronJob> = emptyList(),
    val error: String? = null,
    val isEmpty: Boolean = false,
    val selectedJob: CronJob? = null,
    val isEditorMode: Boolean = false,
    val editorJob: CronJob? = null,
    val cronExpression: String = "",
    val testResult: CronTestResult? = null
)

data class CronTestResult(
    val expression: String,
    val nextRuns: List<String>,
    val isValid: Boolean,
    val error: String? = null
)
```

**Archivo**: `CronEvent.kt`

```kotlin
package com.makoclaw.feature.cron.presentation.state

sealed class CronEvent {
    data object LoadJobs : CronEvent()
    data object Refresh : CronEvent()
    data class CreateJob(val job: CronJob) : CronEvent()
    data class UpdateJob(val job: CronJob) : CronEvent()
    data class DeleteJob(val id: String) : CronEvent()
    data class ToggleJob(val id: String) : CronEvent()
    data class SelectJob(val id: String) : CronEvent()
    data object OpenEditor : CronEvent()
    data class CloseEditor : CronEvent()
    data class SaveEditor(val job: CronJob) : CronEvent()
    data class GenerateCron(val job: CronJob) : CronEvent()
    data class TestRun(val expression: String) : CronEvent()
}
```

**Archivo**: `CronEffect.kt`

```kotlin
package com.makoclaw.feature.cron.presentation.state

sealed class CronEffect {
    data class ShowSnackbar(val message: String) : CronEffect()
    data class NavigateToEditor(val jobId: String?) : CronEffect()
    data class CronGenerated(val expression: String) : CronEffect()
}
```

**Archivo**: `CronViewModel.kt`

```kotlin
package com.makoclaw.feature.cron.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.CronJob
import com.makoclaw.feature.cron.data.repository.CronRepository
import com.makoclaw.feature.cron.presentation.state.CronEffect
import com.makoclaw.feature.cron.presentation.state.CronEvent
import com.makoclaw.feature.cron.presentation.state.CronTestResult
import com.makoclaw.feature.cron.presentation.state.CronUiState
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class CronViewModel @Inject constructor(
    private val repository: CronRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(CronUiState())
    val uiState: StateFlow<CronUiState> = _uiState.asStateFlow()

    private val _effects = MutableSharedFlow<CronEffect>()
    val effects: SharedFlow<CronEffect> = _effects.asSharedFlow()

    init {
        loadJobs()
    }

    fun onEvent(event: CronEvent) {
        when (event) {
            is CronEvent.LoadJobs -> loadJobs()
            is CronEvent.Refresh -> loadJobs()
            is CronEvent.CreateJob -> createJob(event.job)
            is CronEvent.UpdateJob -> updateJob(event.job)
            is CronEvent.DeleteJob -> deleteJob(event.id)
            is CronEvent.ToggleJob -> toggleJob(event.id)
            is CronEvent.SelectJob -> selectJob(event.id)
            is CronEvent.OpenEditor -> openEditor()
            is CronEvent.CloseEditor -> closeEditor()
            is CronEvent.SaveEditor -> saveEditor(event.job)
            is CronEvent.GenerateCron -> generateCron(event.job)
            is CronEvent.TestRun -> testRun(event.expression)
        }
    }

    private fun loadJobs() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            try {
                repository.getJobs()
                    .collect { jobs ->
                        _uiState.update {
                            it.copy(
                                isLoading = false,
                                jobs = jobs,
                                isEmpty = jobs.isEmpty()
                            )
                        }
                    }
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        error = "Failed to load jobs: ${e.message}"
                    )
                }
                _effects.emit(CronEffect.ShowSnackbar("Error loading jobs"))
            }
        }
    }

    private fun createJob(job: CronJob) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }

            try {
                repository.createJob(job)
                _effects.emit(CronEffect.ShowSnackbar("Job created"))
                loadJobs()
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
                _effects.emit(CronEffect.ShowSnackbar("Error creating job"))
            }
        }
    }

    private fun updateJob(job: CronJob) {
        viewModelScope.launch {
            try {
                repository.updateJob(job)
                _effects.emit(CronEffect.ShowSnackbar("Job updated"))
                loadJobs()
            } catch (e: Exception) {
                _effects.emit(CronEffect.ShowSnackbar("Error updating job"))
            }
        }
    }

    private fun deleteJob(id: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }

            try {
                repository.deleteJob(id)
                _effects.emit(CronEffect.ShowSnackbar("Job deleted"))
                loadJobs()
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
                _effects.emit(CronEffect.ShowSnackbar("Error deleting job"))
            }
        }
    }

    private fun toggleJob(id: String) {
        viewModelScope.launch {
            try {
                val job = _uiState.value.jobs.find { it.id == id }
                job?.let {
                    repository.toggleJob(id, !it.enabled)
                    _effects.emit(CronEffect.ShowSnackbar("Job ${if (!it.enabled) "enabled" else "disabled"}"))
                    loadJobs()
                }
            } catch (e: Exception) {
                _effects.emit(CronEffect.ShowSnackbar("Error toggling job"))
            }
        }
    }

    private fun selectJob(id: String) {
        _uiState.update {
            it.copy(selectedJob = it.jobs.find { j -> j.id == id })
        }
    }

    private fun openEditor() {
        _uiState.update {
            it.copy(
                isEditorMode = true,
                editorJob = it.selectedJob?.copy() ?: CronJob(
                    id = "",
                    name = "",
                    description = "",
                    cronExpression = "",
                    workflowId = "",
                    enabled = true,
                    createdAt = System.currentTimeMillis(),
                    updatedAt = System.currentTimeMillis()
                )
            )
        }
        _effects.emit(CronEffect.NavigateToEditor(it.selectedJob?.id))
    }

    private fun closeEditor() {
        _uiState.update {
            it.copy(
                isEditorMode = false,
                editorJob = null
            )
        }
    }

    private fun saveEditor(job: CronJob) {
        viewModelScope.launch {
            try {
                if (job.id.isBlank()) {
                    repository.createJob(job)
                    _effects.emit(CronEffect.ShowSnackbar("Job created"))
                } else {
                    repository.updateJob(job)
                    _effects.emit(CronEffect.ShowSnackbar("Job updated"))
                }
                closeEditor()
                loadJobs()
            } catch (e: Exception) {
                _effects.emit(CronEffect.ShowSnackbar("Error saving job"))
            }
        }
    }

    private fun generateCron(job: CronJob) {
        viewModelScope.launch {
            try {
                val expression = repository.generateCron(job)
                _effects.emit(CronEffect.CronGenerated(expression))
                _effects.emit(CronEffect.ShowSnackbar("Cron expression generated: $expression"))
            } catch (e: Exception) {
                _effects.emit(CronEffect.ShowSnackbar("Error generating cron"))
            }
        }
    }

    private fun testRun(expression: String) {
        viewModelScope.launch {
            try {
                val result = repository.testRun(expression)
                _uiState.update { it.copy(testResult = result) }
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        testResult = CronTestResult(
                            expression = expression,
                            nextRuns = emptyList(),
                            isValid = false,
                            error = e.message
                        )
                    )
                }
            }
        }
    }
}
```

### Instrucciones

1. Crear el directorio `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/state/` si no existe
2. Crear los 4 archivos de state/event/effect/viewmodel
3. Verificar que no hay errores de compilación

---

### TASK-CRON-002: Crear CronRepository

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-cron/src/main/java/com/makoclaw/feature/cron/data/repository/CronRepository.kt`
**Paquete**: `com.makoclaw.feature.cron.data.repository`

### Código Completo

```kotlin
package com.makoclaw.feature.cron.data.repository

import com.makoclaw.core.model.CronJob
import com.makoclaw.core.network.api.CronApi
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.catch
import javax.inject.Inject

interface CronRepository {
    fun getJobs(): Flow<List<CronJob>>
    fun createJob(job: CronJob): Flow<com.makoclaw.core.network.api.ApiResponse>
    fun updateJob(job: CronJob): Flow<com.makoclaw.core.network.api.ApiResponse>
    fun deleteJob(id: String): Flow<com.makoclaw.core.network.api.ApiResponse>
    fun toggleJob(id: String, enabled: Boolean): Flow<com.makoclaw.core.network.api.ApiResponse>
    fun generateCron(job: CronJob): Flow<String>
    fun testRun(expression: String): Flow<com.makoclaw.feature.cron.presentation.state.CronTestResult>
}

@ViewModelScoped
class CronRepositoryImpl @Inject constructor(
    private val api: CronApi,
    private val dao: CronJobDao
) : CronRepository {

    override fun getJobs(): Flow<List<CronJob>> {
        return dao.getAll()
            .map { entities -> entities.map { it.toModel() } }
            .catch { e ->
                // Fallback to API if local cache fails
                api.getJobs()
                    .map { response -> response.jobs }
            }
    }

    override fun createJob(job: CronJob): Flow<com.makoclaw.core.network.api.ApiResponse> {
        return api.createJob(
            com.makoclaw.core.network.api.CreateCronJobRequest(
                name = job.name,
                description = job.description,
                cronExpression = job.cronExpression,
                workflowId = job.workflowId
            )
        )
    }

    override fun updateJob(job: CronJob): Flow<com.makoclaw.core.network.api.ApiResponse> {
        return api.updateJob(
            job.id,
            com.makoclaw.core.network.api.UpdateCronJobRequest(
                name = job.name,
                description = job.description,
                cronExpression = job.cronExpression,
                workflowId = job.workflowId
            )
        )
    }

    override fun deleteJob(id: String): Flow<com.makoclaw.core.network.api.ApiResponse> {
        return api.deleteJob(id)
    }

    override fun toggleJob(id: String, enabled: Boolean): Flow<com.makoclaw.core.network.api.ApiResponse> {
        return api.toggleJob(id, enabled)
    }

    override fun generateCron(job: CronJob): Flow<String> {
        return kotlinx.coroutines.flow.flow {
            val result = api.generateCron(
                com.makoclaw.core.network.api.GenerateCronRequest(
                    schedule = job.description,
                    workflowId = job.workflowId
                )
            )
            emit(result.expression)
        }
    }

    override fun testRun(expression: String): Flow<com.makoclaw.feature.cron.presentation.state.CronTestResult> {
        return kotlinx.coroutines.flow.flow {
            val result = api.testRun(expression)
            emit(
                com.makoclaw.feature.cron.presentation.state.CronTestResult(
                    expression = expression,
                    nextRuns = result.nextRuns,
                    isValid = result.valid,
                    error = if (result.valid) null else "Invalid cron expression"
                )
            )
        }
    }
}
```

**Archivo**: `CronJobDao.kt` (crear en `feature-cron/src/main/java/com/makoclaw/feature/cron/data/database/`)

```kotlin
package com.makoclaw.feature.cron.data.database

import androidx.room.*
import kotlinx.coroutines.flow.Flow

@Entity(tableName = "cron_jobs")
data class CronJobEntity(
    @PrimaryKey val id: String,
    val name: String,
    val description: String,
    val cronExpression: String,
    val workflowId: String,
    val enabled: Boolean,
    val createdAt: Long,
    val updatedAt: Long
)

@Dao
interface CronJobDao {
    @Query("SELECT * FROM cron_jobs ORDER BY createdAt DESC")
    fun getAll(): Flow<List<CronJobEntity>>

    @Query("SELECT * FROM cron_jobs WHERE id = :id")
    suspend fun getById(id: String): CronJobEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insert(job: CronJobEntity)

    @Update
    suspend fun update(job: CronJobEntity)

    @Delete
    suspend fun delete(job: CronJobEntity)

    @Query("DELETE FROM cron_jobs WHERE id = :id")
    suspend fun deleteById(id: String)
}

// Extension function to convert Entity to Model
fun CronJobEntity.toModel(): com.makoclaw.core.model.CronJob {
    return com.makoclaw.core.model.CronJob(
        id = id,
        name = name,
        description = description,
        cronExpression = cronExpression,
        workflowId = workflowId,
        enabled = enabled,
        createdAt = createdAt,
        updatedAt = updatedAt
    )
}
```

**Archivo**: `CronApi.kt` (crear en `core-network/src/main/java/com/makoclaw/core/network/api/` si no existe)

```kotlin
package com.makoclaw.core.network.api

import com.makoclaw.core.model.CronJob
import retrofit2.http.*

data class CronJobsResponse(
    val jobs: List<CronJob> = emptyList()
)

data class CreateCronJobRequest(
    val name: String,
    val description: String,
    val cronExpression: String,
    val workflowId: String
)

data class UpdateCronJobRequest(
    val name: String,
    val description: String,
    val cronExpression: String,
    val workflowId: String
)

data class GenerateCronRequest(
    val schedule: String,
    val workflowId: String
)

data class GenerateCronResponse(
    val expression: String
)

data class TestCronResponse(
    val valid: Boolean,
    val nextRuns: List<String> = emptyList(),
    val error: String? = null
)

interface CronApi {
    @GET("cron/jobs")
    suspend fun getJobs(): CronJobsResponse

    @POST("cron/jobs")
    suspend fun createJob(@Body request: CreateCronJobRequest): ApiResponse

    @PUT("cron/jobs/{id}")
    suspend fun updateJob(
        @Path("id") id: String,
        @Body request: UpdateCronJobRequest
    ): ApiResponse

    @DELETE("cron/jobs/{id}")
    suspend fun deleteJob(@Path("id") id: String): ApiResponse

    @POST("cron/jobs/{id}/toggle")
    suspend fun toggleJob(@Path("id") id: String, @Body enabled: Boolean): ApiResponse

    @POST("cron/generate")
    suspend fun generateCron(@Body request: GenerateCronRequest): GenerateCronResponse

    @POST("cron/test")
    suspend fun testRun(@Body expression: String): TestCronResponse
}
```

### Instrucciones

1. Crear el directorio `feature-cron/src/main/java/com/makoclaw/feature/cron/data/repository/` si no existe
2. Crear el archivo `CronRepository.kt`
3. Crear el directorio `feature-cron/src/main/java/com/makoclaw/feature/cron/data/database/` si no existe
4. Crear el archivo `CronJobDao.kt`
5. Si `CronApi.kt` no existe en `core-network`, crearlo
6. Verificar que no hay errores de compilación

---

### TASK-CRON-003: Crear CronScheduleSelector

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/component/CronScheduleSelector.kt`
**Paquete**: `com.makoclaw.feature.cron.presentation.component`

### Código Completo

```kotlin
package com.makoclaw.feature.cron.presentation.component

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.makoclaw.core.model.CronJob

@Composable
fun CronScheduleSelector(
    job: CronJob,
    onExpressionChange: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    var selectedHour by remember { mutableStateOf(0) }
    var selectedMinute by remember { mutableStateOf(0) }
    var selectedDays by remember { mutableStateOf(setOf<Int>()) }

    Column(
        modifier = modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        Text(
            text = "Schedule",
            style = MaterialTheme.typography.titleMedium
        )

        // Time selector (hour:minute)
        HourMinuteSelector(
            hour = selectedHour,
            minute = selectedMinute,
            onHourChange = { selectedHour = it },
            onMinuteChange = { selectedMinute = it }
        )

        Spacer(modifier = Modifier.height(8.dp))

        // Day of week selector
        Text(
            text = "Days of Week",
            style = MaterialTheme.typography.bodyMedium
        )
        DayOfWeekSelector(
            selectedDays = selectedDays,
            onDayToggle = { day ->
                selectedDays = if (selectedDays.contains(day)) {
                    selectedDays - day
                } else {
                    selectedDays + day
                }
            }
        )

        Spacer(modifier = Modifier.height(16.dp))

        // Generated expression
        CronExpressionDisplay(
            expression = generateExpression(selectedHour, selectedMinute, selectedDays),
            onChange = { onExpressionChange(it) }
        )
    }
}

@Composable
fun HourMinuteSelector(
    hour: Int,
    minute: Int,
    onHourChange: (Int) -> Unit,
    onMinuteChange: (Int) -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth()
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.SpaceEvenly,
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Hour selector
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Text("Hour", style = MaterialTheme.typography.labelSmall)
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    IconButton(
                        onClick = { onHourChange((hour + 1) % 24) },
                        modifier = Modifier.size(32.dp)
                    ) {
                        Text("+", style = MaterialTheme.typography.titleLarge)
                    }
                    Text(
                        text = hour.toString().padStart(2, '0'),
                        style = MaterialTheme.typography.headlineSmall
                    )
                    IconButton(
                        onClick = { onHourChange((hour - 1 + 24) % 24) },
                        modifier = Modifier.size(32.dp)
                    ) {
                        Text("-", style = MaterialTheme.typography.titleLarge)
                    }
                }
            }

            Text(":", style = MaterialTheme.typography.headlineSmall)

            // Minute selector
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Text("Minute", style = MaterialTheme.typography.labelSmall)
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    IconButton(
                        onClick = { onMinuteChange((minute + 15) % 60) },
                        modifier = Modifier.size(32.dp)
                    ) {
                        Text("+", style = MaterialTheme.typography.titleLarge)
                    }
                    Text(
                        text = minute.toString().padStart(2, '0'),
                        style = MaterialTheme.typography.headlineSmall
                    )
                    IconButton(
                        onClick = { onMinuteChange((minute - 15 + 60) % 60) },
                        modifier = Modifier.size(32.dp)
                    ) {
                        Text("-", style = MaterialTheme.typography.titleLarge)
                    }
                }
            }
        }
    }
}

@Composable
fun DayOfWeekSelector(
    selectedDays: Set<Int>,
    onDayToggle: (Int) -> Unit
) {
    val days = listOf("Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun")

    LazyVerticalGrid(
        columns = GridCells.Fixed(7),
        horizontalArrangement = Arrangement.spacedBy(4.dp),
        verticalArrangement = Arrangement.spacedBy(4.dp),
        modifier = Modifier.fillMaxWidth()
    ) {
        items(days.size) { index ->
            val isSelected = selectedDays.contains(index)
            DayChip(
                label = days[index],
                isSelected = isSelected,
                onClick = { onDayToggle(index) }
            )
        }
    }
}

@Composable
fun DayChip(
    label: String,
    isSelected: Boolean,
    onClick: () -> Unit
) {
    Surface(
        onClick = onClick,
        modifier = Modifier
            .fillMaxWidth()
            .aspectRatio(1f)
            .clip(CircleShape),
        color = if (isSelected) {
            MaterialTheme.colorScheme.primaryContainer
        } else {
            MaterialTheme.colorScheme.surfaceVariant
        },
        border = if (isSelected) {
            androidx.compose.foundation.BorderStroke(
                2.dp,
                MaterialTheme.colorScheme.primary
            )
        } else {
            null
        }
    ) {
        Box(
            modifier = Modifier.fillMaxSize(),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = label,
                style = MaterialTheme.typography.bodySmall,
                color = if (isSelected) {
                    MaterialTheme.colorScheme.onPrimaryContainer
                } else {
                    MaterialTheme.colorScheme.onSurfaceVariant
                }
            )
        }
    }
}

@Composable
fun CronExpressionDisplay(
    expression: String,
    onChange: (String) -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Text(
                text = "Cron Expression",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            OutlinedTextField(
                value = expression,
                onValueChange = onChange,
                label = { Text("Expression") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                shape = RoundedCornerShape(8.dp)
            )
        }
    }
}

private fun generateExpression(hour: Int, minute: Int, days: Set<Int>): String {
    val dayPart = if (days.isEmpty()) "*" else days.joinToString(",")
    return "$minute $hour * * $dayPart"
}
```

### Instrucciones

1. Crear el directorio `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/component/` si no existe
2. Crear el archivo `CronScheduleSelector.kt` y pegar todo el código
3. Verificar que no hay errores de compilación

---

### TASK-CRON-004: Crear CronScreen UI

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/screen/CronScreen.kt`
**Paquete**: `com.makoclaw.feature.cron.presentation.screen`

### Código Completo

```kotlin
package com.makoclaw.feature.cron.presentation.screen

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.cron.presentation.viewmodel.CronViewModel
import com.makoclaw.feature.cron.presentation.state.CronEvent
import com.makoclaw.feature.cron.presentation.component.CronJobCard
import com.makoclaw.feature.cron.presentation.component.CronEditorModal

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CronScreen(
    viewModel: CronViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val effects by viewModel.effects.collectAsState(initialValue = null)
    val scope = rememberCoroutineScope()

    var showEditor by remember { mutableStateOf(false) }

    // Snackbar host
    val snackbarHostState = SnackbarHostState()

    // Handle effects
    LaunchedEffect(Unit) {
        effects?.let { effect ->
            when (effect) {
                is CronEffect.ShowSnackbar -> {
                    scope.launch {
                        snackbarHostState.showSnackbar(
                            message = effect.message,
                            duration = SnackbarDuration.Short
                        )
                    }
                }
                is CronEffect.NavigateToEditor -> {
                    showEditor = true
                }
                is CronEffect.CronGenerated -> {
                    // Update expression in editor
                }
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Cron Jobs") },
                actions = {
                    IconButton(onClick = { viewModel.onEvent(CronEvent.Refresh) }) {
                        Icon(Icons.Default.Refresh, "Refresh")
                    }
                }
            )
        },
        snackbarHost = { SnackbarHost(snackbarHostState) },
        floatingActionButton = {
            FloatingActionButton(
                onClick = { viewModel.onEvent(CronEvent.OpenEditor) }
            ) {
                Icon(Icons.Default.Add, "Add Job")
            }
        }
    ) { padding ->
        if (showEditor && uiState.editorJob != null) {
            CronEditorModal(
                job = uiState.editorJob!!,
                testResult = uiState.testResult,
                onDismiss = {
                    viewModel.onEvent(CronEvent.CloseEditor)
                    showEditor = false
                },
                onSave = { viewModel.onEvent(CronEvent.SaveEditor(it)) },
                onGenerateCron = { viewModel.onEvent(CronEvent.GenerateCron(it)) },
                onTestRun = { viewModel.onEvent(CronEvent.TestRun(it)) }
            )
        } else {
            if (uiState.isLoading) {
                LoadingScreen()
            } else if (uiState.error != null) {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    Text(
                        text = "Error: ${uiState.error}",
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.error
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    Button(onClick = { viewModel.onEvent(CronEvent.Refresh) }) {
                        Text("Retry")
                    }
                }
            } else if (uiState.isEmpty) {
                EmptyState(
                    title = "No cron jobs yet",
                    message = "Create your first scheduled task",
                    actionLabel = "Create Job"
                ) {
                    viewModel.onEvent(CronEvent.OpenEditor)
                }
            } else {
                LazyColumn(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    item { Spacer(modifier = Modifier.height(8.dp)) }

                    items(uiState.jobs) { job ->
                        CronJobCard(
                            job = job,
                            onClick = { viewModel.onEvent(CronEvent.SelectJob(job.id)) },
                            onEdit = { viewModel.onEvent(CronEvent.OpenEditor) },
                            onDelete = { viewModel.onEvent(CronEvent.DeleteJob(job.id)) },
                            onToggle = { viewModel.onEvent(CronEvent.ToggleJob(job.id)) }
                        )
                    }

                    item { Spacer(modifier = Modifier.height(80.dp)) }
                }
            }
        }
    }
}
```

**Archivo**: `CronJobCard.kt` (crear en `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/component/`)

```kotlin
package com.makoclaw.feature.cron.presentation.component

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Error
import androidx.compose.material.icons.filled.Schedule
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.makoclaw.core.model.CronJob

@Composable
fun CronJobCard(
    job: CronJob,
    onClick: () -> Unit,
    onEdit: () -> Unit,
    onDelete: () -> Unit,
    onToggle: () -> Unit
) {
    Card(
        onClick = onClick,
        modifier = Modifier.fillMaxWidth()
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            // Header: Name + Toggle
            Row(
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = job.name,
                        style = MaterialTheme.typography.titleMedium
                    )
                    Text(
                        text = job.description,
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }

                Switch(
                    checked = job.enabled,
                    onCheckedChange = { onToggle() }
                )
            }

            // Cron expression
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(
                    Icons.Default.Schedule,
                    contentDescription = "Schedule",
                    tint = MaterialTheme.colorScheme.primary,
                    modifier = Modifier.size(16.dp)
                )
                Text(
                    text = job.cronExpression,
                    style = MaterialTheme.typography.bodyMedium,
                    fontFamily = androidx.compose.ui.text.font.FontFamily.Monospace
                )
            }

            // Footer: Actions + Status
            Row(
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    TextButton(onClick = onEdit) {
                        Text("Edit")
                    }
                    TextButton(
                        onClick = onDelete,
                        colors = ButtonDefaults.textButtonColors(
                            contentColor = MaterialTheme.colorScheme.error
                        )
                    ) {
                        Text("Delete")
                    }
                }

                if (job.enabled) {
                    Icon(
                        Icons.Default.CheckCircle,
                        contentDescription = "Active",
                        tint = MaterialTheme.colorScheme.primary,
                        modifier = Modifier.size(20.dp)
                    )
                } else {
                    Icon(
                        Icons.Default.Error,
                        contentDescription = "Inactive",
                        tint = MaterialTheme.colorScheme.error,
                        modifier = Modifier.size(20.dp)
                    )
                }
            }
        }
    }
}
```

**Archivo**: `CronEditorModal.kt` (crear en `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/component/`)

```kotlin
package com.makoclaw.feature.cron.presentation.component

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.makoclaw.core.model.CronJob
import com.makoclaw.feature.cron.presentation.state.CronTestResult

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CronEditorModal(
    job: CronJob,
    testResult: CronTestResult?,
    onDismiss: () -> Unit,
    onSave: (CronJob) -> Unit,
    onGenerateCron: (CronJob) -> Unit,
    onTestRun: (String) -> Unit
) {
    var name by remember { mutableStateOf(job.name) }
    var description by remember { mutableStateOf(job.description) }
    var expression by remember { mutableStateOf(job.cronExpression) }

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = rememberModalBottomSheetState()
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .height(600.dp)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Header
            Row(
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = if (job.id.isBlank()) "New Cron Job" else "Edit Cron Job",
                    style = MaterialTheme.typography.titleLarge
                )
                IconButton(onClick = onDismiss) {
                    Icon(Icons.Default.Close, "Close")
                }
            }

            // Form
            OutlinedTextField(
                value = name,
                onValueChange = { name = it },
                label = { Text("Name") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true
            )

            OutlinedTextField(
                value = description,
                onValueChange = { description = it },
                label = { Text("Description") },
                modifier = Modifier.fillMaxWidth(),
                minLines = 2
            )

            // Schedule selector
            CronScheduleSelector(
                job = job,
                onExpressionChange = { expression = it }
            )

            // Manual expression override
            OutlinedTextField(
                value = expression,
                onValueChange = { expression = it },
                label = { Text("Cron Expression (Manual)") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true
            )

            // Test result
            testResult?.let { result ->
                if (result.isValid) {
                    Card(
                        colors = CardDefaults.cardColors(
                            containerColor = MaterialTheme.colorScheme.primaryContainer
                        )
                    ) {
                        Column(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(12.dp),
                            verticalArrangement = Arrangement.spacedBy(4.dp)
                        ) {
                            Text(
                                text = "Valid Expression ✓",
                                style = MaterialTheme.typography.titleSmall,
                                color = MaterialTheme.colorScheme.onPrimaryContainer
                            )
                            result.nextRuns.forEach { run ->
                                Text(
                                    text = run,
                                    style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.8f)
                                )
                            }
                        }
                    }
                } else {
                    Card(
                        colors = CardDefaults.cardColors(
                            containerColor = MaterialTheme.colorScheme.errorContainer
                        )
                    ) {
                        Column(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(12.dp),
                            verticalArrangement = Arrangement.spacedBy(4.dp)
                        ) {
                            Text(
                                text = "Invalid Expression ✗",
                                style = MaterialTheme.typography.titleSmall,
                                color = MaterialTheme.colorScheme.onErrorContainer
                            )
                            Text(
                                text = result.error ?: "Unknown error",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onErrorContainer.copy(alpha = 0.8f)
                            )
                        }
                    }
                }
            }

            // Actions
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                modifier = Modifier.fillMaxWidth()
            ) {
                OutlinedButton(
                    onClick = { onTestRun(expression) },
                    modifier = Modifier.weight(1f)
                ) {
                    Text("Test")
                }
                OutlinedButton(
                    onClick = {
                        onGenerateCron(
                            job.copy(
                                name = name,
                                description = description,
                                cronExpression = expression
                            )
                        )
                    },
                    modifier = Modifier.weight(1f)
                ) {
                    Text("AI Generate")
                }
                Button(
                    onClick = {
                        onSave(
                            job.copy(
                                name = name,
                                description = description,
                                cronExpression = expression,
                                updatedAt = System.currentTimeMillis()
                            )
                        )
                    },
                    modifier = Modifier.weight(1f),
                    enabled = name.isNotBlank()
                ) {
                    Text("Save")
                }
            }
        }
    }
}
```

### Instrucciones

1. Crear el directorio `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/screen/` si no existe
2. Crear el archivo `CronScreen.kt` y pegar todo el código
3. Crear los archivos `CronJobCard.kt` y `CronEditorModal.kt` en `feature-cron/src/main/java/com/makoclaw/feature/cron/presentation/component/`
4. Verificar que no hay errores de compilación

---

## Resumen BATCH 3

### Archivos a Crear (~50 archivos)

**feature-workflows (13 archivos)**:
- 4 state/event/effect/viewmodel
- 1 Repository + 1 DAO
- 1 API
- 1 WorkflowCanvasEditor (ALTA complejidad)
- 1 WorkflowsScreen
- 1 WorkflowCard
- 1 ExecutionLogsModal
- DI + Tests

**feature-metrics (12 archivos)**:
- 4 state/event/effect/viewmodel
- 1 Repository + 1 DAO
- 1 API
- 1 MetricsCharts (Vico)
- 1 MetricsScreen
- DI + Tests

**feature-cron (14 archivos)**:
- 4 state/event/effect/viewmodel
- 1 Repository + 1 DAO
- 1 API
- 1 CronScheduleSelector
- 1 CronScreen
- 1 CronJobCard
- 1 CronEditorModal
- DI + Tests

### Tiempo Estimado
- **feature-workflows**: ~3 días (incluyendo canvas editor complejo)
- **feature-metrics**: ~1 día
- **feature-cron**: ~1 día

**Total BATCH 3**: ~5 días de trabajo

### Checklist de Validación

Antes de considerar completado el BATCH 3:

**feature-workflows**:
- [ ] WorkflowsViewModel maneja eventos
- [ ] WorkflowsRepository conecta a API
- [ ] WorkflowCanvasEditor permite crear/editar/eliminar nodos y edges
- [ ] Zoom/pan funciona
- [ ] WorkflowsScreen muestra lista
- [ ] Ejecución de workflows funciona

**feature-metrics**:
- [ ] MetricsViewModel filtra por timeRange y agent
- [ ] MetricsRepository cachea datos
- [ ] Vico charts renderizan correctamente
- [ ] MetricsScreen muestra todos los charts
- [ ] Export funciona

**feature-cron**:
- [ ] CronViewModel maneja jobs
- [ ] CronScheduleSelector genera expresiones
- [ ] CronScreen muestra lista
- [ ] Toggle de jobs funciona
- [ ] Test de expresiones funciona

---

## Siguientes Pasos

Después de completar el BATCH 3 (Features ALTA), el siguiente es:

**BATCH 4**: Features MEDIA (12 tareas)
- feature-skills (3 tareas)
- feature-agents (3 tareas)
- feature-memory (3 tareas)
- feature-history (3 tareas)

Recomendación:
1. **Terminar BATCH 3 primero** (validar que todo funcione)
2. **Luego ir a BATCH 4**

---

## Notas Importantes

1. **Vico Library**: Asegúrate de agregar las dependencias de Vico en `build.gradle.kts`:
```kotlin
implementation("com.patrykandpatrick.vico:compose:1.14.0")
implementation("com.patrykandpatrick.vico:compose-m3:1.14.0")
```

2. **Workflow Canvas**: El editor de canvas es una implementación simplificada. Para producción, considera:
   - Usar librerías especializadas (ej. PanZoomImage)
   - Implementar proper hit detection
   - Guardar/restaurar estado de zoom/pan
   - Mejorar la UX de conexión de nodos

3. **Cron Expressions**: El generador de cron es básico. Para producción, considera:
   - Integrar librería de parseo de cron (ej. cron-utils)
   - Validar expresiones con backend
   - Mostrar preview de próximas ejecuciones

4. **Tests**: Recomiendo crear tests básicos para ViewModels y Repositories

---

## Finalización

Una vez que hayas creado todos los archivos y verificado que compilan, el BATCH 3 estará **COMPLETADO** 🎉

**¡Éxito, hermano!** 🦈
