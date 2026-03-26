package com.makoclaw.feature.workflows.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.Workflow
import com.makoclaw.feature.workflows.data.repository.WorkflowsRepository
import com.makoclaw.feature.workflows.presentation.state.WorkflowsEffect
import com.makoclaw.feature.workflows.presentation.state.WorkflowsEvent
import com.makoclaw.feature.workflows.presentation.state.WorkflowsUiState
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
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
                repository.getWorkflows().collect { workflows ->
                    _uiState.update { currentState ->
                        currentState.copy(
                            isLoading = false,
                            workflows = workflows,
                            isEmpty = workflows.isEmpty(),
                            selectedWorkflow = syncSelection(currentState.selectedWorkflow, workflows)
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
        saveWorkflow(
            Workflow(
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
        )
    }

    private fun updateWorkflow(workflow: Workflow) = saveWorkflow(workflow)

    private fun saveWorkflow(workflow: Workflow) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            try {
                val targetWorkflow = workflow.copy(updatedAt = System.currentTimeMillis())

                if (targetWorkflow.id.isBlank()) {
                    repository.createWorkflow(targetWorkflow).collect()
                    _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow created"))
                } else {
                    repository.updateWorkflow(targetWorkflow).collect()
                    _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow updated"))
                }

                closeEditor()
                loadWorkflows()
            } catch (e: Exception) {
                _uiState.update { currentState ->
                    currentState.copy(isLoading = false, error = e.message)
                }
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error saving workflow"))
            }
        }
    }

    private fun deleteWorkflow(id: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            try {
                repository.deleteWorkflow(id).collect()
                _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow deleted"))
                loadWorkflows()
            } catch (e: Exception) {
                _uiState.update { currentState ->
                    currentState.copy(isLoading = false, error = e.message)
                }
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error deleting workflow"))
            }
        }
    }

    private fun executeWorkflow(id: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isExecuting = true, error = null) }

            try {
                _effects.emit(WorkflowsEffect.ExecutionStarted(id))
                repository.executeWorkflow(id).collect { log ->
                    _uiState.update { currentState ->
                        currentState.copy(executionLogs = currentState.executionLogs + log)
                    }
                }

                _uiState.update { it.copy(isExecuting = false) }
                _effects.emit(WorkflowsEffect.ExecutionCompleted(id, true))
                _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow executed successfully"))
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        isExecuting = false,
                        error = "Execution failed: ${e.message}"
                    )
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

            try {
                repository.getExecutionLogs(workflowId).collect { logs ->
                    _uiState.update { it.copy(executionLogs = logs) }
                }

                _effects.emit(WorkflowsEffect.NavigateToLogs(workflowId))
            } catch (e: Exception) {
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error loading logs"))
            }
        }
    }

    private fun openEditor() {
        val selectedWorkflow = _uiState.value.selectedWorkflow

        _uiState.update {
            it.copy(
                isEditorMode = true,
                editorWorkflow = selectedWorkflow?.copy() ?: Workflow(
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

        viewModelScope.launch {
            _effects.emit(WorkflowsEffect.NavigateToEditor(selectedWorkflow?.id))
        }
    }

    private fun closeEditor() {
        _uiState.update { it.copy(isEditorMode = false, editorWorkflow = null) }
    }

    private fun saveEditor(workflow: Workflow) {
        viewModelScope.launch {
            try {
                if (workflow.id.isBlank()) {
                    repository.createWorkflow(workflow).collect()
                    _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow created"))
                } else {
                    repository.updateWorkflow(workflow).collect()
                    _effects.emit(WorkflowsEffect.ShowSnackbar("Workflow updated"))
                }

                closeEditor()
                loadWorkflows()
            } catch (e: Exception) {
                _effects.emit(WorkflowsEffect.ShowSnackbar("Error saving workflow"))
            }
        }
    }

    private fun syncSelection(selected: Workflow?, workflows: List<Workflow>): Workflow? =
        selected?.let { current -> workflows.firstOrNull { it.id == current.id } }
}
