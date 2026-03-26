package com.makoclaw.feature.dashboard.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.network.api.AgentApi
import com.makoclaw.core.network.api.ChatApi
import com.makoclaw.core.network.api.ConfigApi
import com.makoclaw.core.network.api.MetricsApi
import com.makoclaw.core.network.api.TaskApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class DashboardUiState(
    val chatCount: Int = 0,
    val taskCount: Int = 0,
    val agentCount: Int = 0,
    val llmCalls: Int = 0,
    val isHealthy: Boolean = true,
    val currentProvider: String = "",
    val currentModel: String = "",
    val isLoading: Boolean = true,
    val username: String = ""
)

@HiltViewModel
class DashboardViewModel @Inject constructor(
    private val chatApi: ChatApi,
    private val taskApi: TaskApi,
    private val agentApi: AgentApi,
    private val configApi: ConfigApi,
    private val metricsApi: MetricsApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(DashboardUiState())
    val uiState: StateFlow<DashboardUiState> = _uiState.asStateFlow()

    init { refresh() }

    fun refresh() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val sessions = chatApi.getSessions()
                _uiState.update { it.copy(chatCount = sessions.sessions.size) }
            } catch (_: Exception) { }

            try {
                val tasks = taskApi.getTasks()
                _uiState.update { it.copy(taskCount = tasks.tasks.size) }
            } catch (_: Exception) { }

            try {
                val agents = agentApi.getAgents()
                _uiState.update { it.copy(agentCount = agents.specialists.size) }
            } catch (_: Exception) { }

            try {
                val models = configApi.getModels()
                _uiState.update {
                    it.copy(
                        currentProvider = models.current_provider,
                        currentModel = models.current_model
                    )
                }
            } catch (_: Exception) { }

            try {
                val metrics = metricsApi.getMetrics()
                _uiState.update { it.copy(llmCalls = metrics.llmCalls) }
            } catch (_: Exception) { }

            try {
                configApi.healthCheck()
                _uiState.update { it.copy(isHealthy = true) }
            } catch (_: Exception) {
                _uiState.update { it.copy(isHealthy = false) }
            }

            _uiState.update { it.copy(isLoading = false) }
        }
    }
}
