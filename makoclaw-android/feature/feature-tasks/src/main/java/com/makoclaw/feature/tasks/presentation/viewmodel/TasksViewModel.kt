package com.makoclaw.feature.tasks.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.CreateTaskRequest
import com.makoclaw.core.model.Task
import com.makoclaw.core.model.TaskStatus
import com.makoclaw.core.model.UpdateTaskStatusRequest
import com.makoclaw.core.network.api.TaskApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class TasksUiState(
    val tasks: List<Task> = emptyList(),
    val tasksByStatus: Map<TaskStatus, List<Task>> = emptyMap(),
    val selectedTask: Task? = null,
    val isLoading: Boolean = true,
    val error: String? = null,
    val searchQuery: String = ""
)

@HiltViewModel
class TasksViewModel @Inject constructor(
    private val taskApi: TaskApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(TasksUiState())
    val uiState: StateFlow<TasksUiState> = _uiState.asStateFlow()

    init {
        loadTasks()
    }

    fun loadTasks() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val response = taskApi.getTasks()
                val grouped = response.tasks.groupBy { it.status }
                _uiState.update {
                    it.copy(
                        tasks = response.tasks,
                        tasksByStatus = grouped,
                        isLoading = false
                    )
                }
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(isLoading = false, error = e.message)
                }
            }
        }
    }

    fun createTask(title: String, description: String) {
        viewModelScope.launch {
            try {
                taskApi.createTask(CreateTaskRequest(title, description))
                loadTasks()
            } catch (_: Exception) { }
        }
    }

    fun updateTaskStatus(taskId: String, status: TaskStatus) {
        viewModelScope.launch {
            try {
                taskApi.updateTaskStatus(taskId, UpdateTaskStatusRequest(status.value))
                loadTasks()
            } catch (_: Exception) { }
        }
    }

    fun deleteTask(taskId: String) {
        viewModelScope.launch {
            try {
                taskApi.deleteTask(taskId)
                loadTasks()
            } catch (_: Exception) { }
        }
    }

    fun onTaskSelected(task: Task) {
        _uiState.update { it.copy(selectedTask = task) }
    }
}
