package com.makoclaw.feature.workflows.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.Workflow
import com.makoclaw.feature.workflows.data.repository.WorkflowsRepository
import com.makoclaw.feature.workflows.presentation.state.WorkflowsEffect
import com.makoclaw.feature.workflows.presentation.state.WorkflowsEvent
import com.makoclaw.feature.workflows.presentation.state.WorkflowsUiState
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

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
            WorkflowsEvent.LoadWorkflows,
            WorkflowsEvent.Refresh -> loadWorkflows()
            is WorkflowsEvent.CreateWorkflow -> createWorkflow(event.name, event.description)
            is WorkflowsEvent.UpdateWorkflow -> updateWorkflow(event.workflow)
            is WorkflowsEvent.DeleteWorkflow -> deleteWorkflow(event.id)
            is WorkflowsEvent.ExecuteWorkflow -> executeWorkflow(event.id)
            is WorkflowsEvent.SelectWorkflow -> selectWorkflow(event.id)
            is WorkflowsEvent.ViewLogs -> viewLogs(event.workflowId)
            WorkflowsEvent.OpenEditor -> openEditor()
            WorkflowsEvent.CloseEditor -> closeEditor()
            is WorkflowsEvent.SaveEditor -> saveEditor(event.workflow)
        }
    }

    private fun loadWorkflows() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                repository.getWorkflows().collect { workflows ->
                    _uiState.update {
                        it.copy(
                            isLoading = false,
                            workflows = workflows,
                            isEmpty = workflows.isEmpty(),
                            selectedWorkflow = syncSelection(it.selectedWorkflow, workflows)
                        )
                    }
                }
            }.onFailure { error ->
                _uiState.update {
                    it.copy(isLoading = false, error = "Failed to load workflows: ${error.message}")
                }
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error loading workflows"))
            }
        }
    }

    private fun createWorkflow(name: String, description: String) {
        val workflow = Workflow(
            id = "",
            name = name,
            description = description,
            createdAt = System.currentTimeMillis(),
            updatedAt = System.currentTimeMillis()
        )
        saveWorkflow(flowWorkflow = workflow)
    }

    private fun updateWorkflow(workflow: Workflow) = saveWorkflow(workflow)

    private fun saveWorkflow(flowWorkflow: Workflow) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                val target = flowWorkflow.copy(updatedAt = System.currentTimeMillis())
                if (target.id.isBlank()) repository.createWorkflow(target) else repository.updateWorkflow(target)
            }.onFailure {
                _uiState.update { state -> state.copy(isLoading = false, error = it.message) }
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error saving workflow"))
            }?.getOrNull()?.collect {
                closeEditor()
                _effects.emit(WorkflowsEffect.ShowSnackbar(it.message.ifBlank { "Workflow saved" }))
                loadWorkflows()
            }
        }
    }

    private fun deleteWorkflow(id: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching { repository.deleteWorkflow(id).collect() }
                .onSuccess {
                    _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow deleted"))
                    loadWorkflows()
                }
                .onFailure {
                    _uiState.update { state -> state.copy(isLoading = false, error = it.message) }
                    _effects.emit(WorkflowsEffect.ShowSnackbar("Error deleting workflow"))
                }
        }
    }

    private fun executeWorkflow(id: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isExecuting = true, error = null) }
            _effects.emit(WorkflowsEffect.ExecutionStarted(id))
            runCatching {
                repository.executeWorkflow(id).collect { log ->
                    _uiState.update { state ->
                        state.copy(
                            isExecuting = false,
                            executionLogs = state.executionLogs + log
                        )
                    }
                }
            }.onSuccess {
                _effects.emit(WorkflowsEffect.ExecutionCompleted(id, true))
                _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow executed successfully"))
            }.onFailure {
                _uiState.update { state ->
                    state.copy(isExecuting = false, error = "Execution failed: ${it.message}")
                }
                _effects.emit(WorkflowsEffect.ExecutionCompleted(id, false))
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error executing workflow"))
            }
        }
    }

    private fun selectWorkflow(id: String) {
        _uiState.update { state ->
            state.copy(selectedWorkflow = state.workflows.find { it.id == id })
        }
    }

    private fun viewLogs(workflowId: String) {
        viewModelScope.launch {
            runCatching {
                repository.getExecutionLogs(workflowId).collect { logs ->
                    _uiState.update { it.copy(executionLogs = logs) }
                }
            }.onSuccess {
                _effects.emit(WorkflowsEffect.NavigateToLogs(workflowId))
            }.onFailure {
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error loading logs"))
            }
        }
    }

    private fun openEditor() {
        val selected = _uiState.value.selectedWorkflow
        _uiState.update {
            it.copy(
                isEditorMode = true,
                editorWorkflow = selected?.copy() ?: Workflow(
                    id = "",
                    name = "",
                    description = "",
                    createdAt = System.currentTimeMillis(),
                    updatedAt = System.currentTimeMillis()
                )
            )
        }
        viewModelScope.launch {
            _effects.emit(WorkflowsEffect.NavigateToEditor(selected?.id))
        }
    }

    private fun closeEditor() {
        _uiState.update { it.copy(isEditorMode = false, editorWorkflow = null) }
    }

    private fun syncSelection(selected: Workflow?, workflows: List<Workflow>): Workflow? =
        selected?.let { current -> workflows.firstOrNull { it.id == current.id } }
}
