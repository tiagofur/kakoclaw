package com.makoclaw.feature.workflows.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.Workflow
import com.makoclaw.core.network.api.RunWorkflowRequest
import com.makoclaw.core.network.api.WorkflowsApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class WorkflowsUiState(
    val workflows: List<Workflow> = emptyList(),
    val isLoading: Boolean = true,
    val isRunning: Boolean = false,
    val error: String? = null
)

@HiltViewModel
class WorkflowsViewModel @Inject constructor(
    private val workflowsApi: WorkflowsApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(WorkflowsUiState())
    val uiState: StateFlow<WorkflowsUiState> = _uiState.asStateFlow()

    init { loadWorkflows() }

    fun loadWorkflows() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val response = workflowsApi.listWorkflows()
                _uiState.update { it.copy(workflows = response.workflows, isLoading = false) }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun runWorkflow(name: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isRunning = true) }
            try {
                workflowsApi.runWorkflow(name, RunWorkflowRequest())
            } catch (_: Exception) { }
            _uiState.update { it.copy(isRunning = false) }
        }
    }

    fun deleteWorkflow(name: String) {
        viewModelScope.launch {
            try {
                workflowsApi.deleteWorkflow(name)
                loadWorkflows()
            } catch (_: Exception) { }
        }
    }
}
