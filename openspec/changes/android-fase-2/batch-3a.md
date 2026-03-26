# Implementación BATCH 3A: Feature Workflows

## Instrucciones

Este documento contiene TODO el código fuente para las 5 tareas del feature-workflows.

---

## TASK-WORKFLOWS-001: Crear WorkflowsViewModel

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
1. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/state/`
2. Crear los 4 archivos: `WorkflowsUiState.kt`, `WorkflowsEvent.kt`, `WorkflowsEffect.kt`, `WorkflowsViewModel.kt`

---

## TASK-WORKFLOWS-002: Crear WorkflowsRepository

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/repository/WorkflowsRepository.kt`

### Código Completo

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
                api.getWorkflows()
                    .map { response -> response.workflows }
            }
    }

    override fun createWorkflow(workflow: Workflow): Flow<ApiResponse> {
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

**Archivo**: `WorkflowDao.kt` en `feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/database/`

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

**Archivo**: `WorkflowsApi.kt` en `core-network/src/main/java/com/makoclaw/core/network/api/` (crear si no existe)

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
1. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/repository/`
2. Crear `WorkflowsRepository.kt`
3. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/database/`
4. Crear `WorkflowDao.kt`
5. Si `WorkflowsApi.kt` no existe en `core-network`, crearlo

---

## TASK-WORKFLOWS-003: Crear WorkflowCanvasEditor

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/screen/WorkflowCanvasEditor.kt`

### Código Completo

```kotlin
package com.makoclaw.feature.workflows.presentation.screen

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectDragGestures
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
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.unit.dp
import com.makoclaw.core.model.Workflow
import com.makoclaw.core.model.WorkflowNode
import com.makoclaw.core.model.WorkflowEdge
import com.makoclaw.core.model.NodeType

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
            // Canvas for edges
            Canvas(
                modifier = Modifier
                    .fillMaxSize()
                    .pointerInput(Unit) {
                        detectDragGestures { change, dragAmount ->
                            if (draggingNodeId == null) {
                                offset += dragAmount
                            }
                            change.consume()
                        }
                    }
            ) {
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

                drawContext.canvas.restore()
            }

            // Nodes overlay
            nodes.forEach { node ->
                WorkflowNode(
                    node = node,
                    isSelected = node.id == selectedNodeId,
                    onTap = { selectedNodeId = node.id },
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

            // FAB to add node
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
    onTap: () -> Unit,
    onDrag: (Offset) -> Unit,
    onDragStart: () -> Unit,
    onDragEnd: () -> Unit,
    onDelete: () -> Unit,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier
            .width(node.width.dp)
            .height(node.height.dp)
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
            containerColor = if (isSelected)
                MaterialTheme.colorScheme.primaryContainer
            else
                MaterialTheme.colorScheme.surface
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
                .clickable { onTap() }
                .padding(8.dp),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = node.name,
                style = MaterialTheme.typography.bodyMedium
            )
        }

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
1. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/screen/`
2. Crear `WorkflowCanvasEditor.kt`
3. Verificar que no hay errores de compilación

**NOTA**: Este es un canvas editor simplificado. Para producción, considera:
- Implementar proper pinch-to-zoom
- Guardar/restaurar estado de zoom/pan
- Mejorar hit detection

---

## TASK-WORKFLOWS-004: Crear WorkflowsScreen UI

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/screen/WorkflowsScreen.kt`

### Código Completo

```kotlin
package com.makoclaw.feature.workflows.presentation.screen

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
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

    val snackbarHostState = SnackbarHostState()

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
                is WorkflowsEffect.ExecutionStarted -> {}
                is WorkflowsEffect.ExecutionCompleted -> {}
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
1. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/screen/`
2. Crear `WorkflowsScreen.kt`

---

## TASK-WORKFLOWS-005: Crear WorkflowCard y ExecutionLogsModal

### Ubicación
**Archivo 1**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/component/WorkflowCard.kt`
**Archivo 2**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/component/ExecutionLogsModal.kt`

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

            if (workflow.description.isNotBlank()) {
                Text(
                    text = workflow.description,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 2
                )
            }

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
1. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/component/`
2. Crear `WorkflowCard.kt`
3. Crear `ExecutionLogsModal.kt`

---

## DI Module: WorkflowsModule

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/di/WorkflowsModule.kt`

### Código Completo

```kotlin
package com.makoclaw.feature.workflows.data.di

import com.makoclaw.feature.workflows.data.repository.WorkflowsRepository
import com.makoclaw.feature.workflows.data.repository.WorkflowsRepositoryImpl
import com.makoclaw.feature.workflows.data.database.WorkflowDao
import dagger.Binds
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.components.ViewModelComponent
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object WorkflowsDatabaseModule {

    @Provides
    @Singleton
    fun provideWorkflowDao(
        database: com.makoclaw.core.database.MakoClawDatabase
    ): WorkflowDao {
        return database.workflowDao()
    }
}

@Module
@InstallIn(ViewModelComponent::class)
abstract class WorkflowsRepositoryModule {

    @Binds
    abstract fun bindWorkflowsRepository(
        impl: WorkflowsRepositoryImpl
    ): WorkflowsRepository
}
```

### Instrucciones
1. Crear el directorio `feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/di/`
2. Crear `WorkflowsModule.kt`
3. Actualizar `MakoClawDatabase.kt` para incluir `WorkflowDao`:

```kotlin
@Dao
abstract fun workflowDao(): WorkflowDao
```

---

## Resumen

### Archivos a Crear (~13 archivos)

| # | Archivo | Líneas aprox. |
|--|--------|--------------|
| 1 | WorkflowsUiState.kt | ~15 |
| 2 | WorkflowsEvent.kt | ~15 |
| 3 | WorkflowsEffect.kt | ~10 |
| 4 | WorkflowsViewModel.kt | ~200 |
| 5 | WorkflowsRepository.kt | ~60 |
| 6 | WorkflowDao.kt | ~50 |
| 7 | WorkflowsApi.kt | ~60 |
| 8 | WorkflowCanvasEditor.kt | ~200 |
| 9 | WorkflowsScreen.kt | ~150 |
| 10 | WorkflowCard.kt | ~100 |
| 11 | ExecutionLogsModal.kt | ~100 |
| 12 | WorkflowsModule.kt | ~30 |

**Total BATCH 3A**: ~13 archivos, ~1000 líneas

---

## Finalización

Una vez creados todos los archivos y verificado que compilen, el **BATCH 3A estará COMPLETADO** 🎉

**¡Éxito, hermano!** 🦈
