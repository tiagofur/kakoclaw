# Micro-Delegación 1: WorkflowsViewModel

## Tarea: TASK-WORKFLOWS-001
Crear WorkflowsViewModel con state, events y effects.

## Archivos a Crear (4 archivos)

### Archivo 1: WorkflowsUiState.kt
**Ubicación**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/state/WorkflowsUiState.kt`

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

### Archivo 2: WorkflowsEvent.kt
**Ubicación**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/state/WorkflowsEvent.kt`

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

### Archivo 3: WorkflowsEffect.kt
**Ubicación**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/state/WorkflowsEffect.kt`

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

### Archivo 4: WorkflowsViewModel.kt
**Ubicación**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/viewmodel/WorkflowsViewModel.kt`

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

## Instrucciones

1. Crear el directorio: `feature-workflows/src/main/java/com/makoclaw/feature/workflows/presentation/state/`
2. Crear los 4 archivos en las ubicaciones indicadas
3. Verificar que no hay errores de compilación
4. Avanzar a la siguiente micro-delegación

## Resumen

- **Archivos creados**: 4
- **Líneas de código**: ~300
- **Complejidad**: Media
- **Tiempo estimado**: ~30-45 minutos

## Finalización

Una vez completado, avísame para lanzar la Micro-Delegación 2 (Repository).

**¡Éxito, hermano!** 🦈
