package com.makoclaw.feature.agents.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.AgentMetrics
import com.makoclaw.core.model.CreateSpecialistRequest
import com.makoclaw.core.model.OrchestratorConfig
import com.makoclaw.core.model.Specialist
import com.makoclaw.core.model.Swarm
import com.makoclaw.core.network.api.AgentApi
import com.makoclaw.core.network.api.GenerateRequest
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class AgentsUiState(
    val specialists: List<Specialist> = emptyList(),
    val orchestrator: OrchestratorConfig = OrchestratorConfig(),
    val swarms: List<Swarm> = emptyList(),
    val metrics: AgentMetrics = AgentMetrics(),
    val isLoading: Boolean = true,
    val isGenerating: Boolean = false,
    val error: String? = null
)

@HiltViewModel
class AgentsViewModel @Inject constructor(
    private val agentApi: AgentApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(AgentsUiState())
    val uiState: StateFlow<AgentsUiState> = _uiState.asStateFlow()

    init {
        loadAgents()
        loadMetrics()
        loadSwarms()
    }

    fun loadAgents() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val response = agentApi.getAgents()
                _uiState.update {
                    it.copy(
                        specialists = response.specialists,
                        orchestrator = response.orchestrator,
                        isLoading = false
                    )
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun loadMetrics() {
        viewModelScope.launch {
            try {
                val metrics = agentApi.getMetrics()
                _uiState.update { it.copy(metrics = metrics) }
            } catch (_: Exception) { }
        }
    }

    fun loadSwarms() {
        viewModelScope.launch {
            try {
                val response = agentApi.getSwarms()
                _uiState.update { it.copy(swarms = response.swarms) }
            } catch (_: Exception) { }
        }
    }

    fun createSpecialist(name: String, speciality: String, systemPrompt: String) {
        viewModelScope.launch {
            try {
                agentApi.createSpecialist(CreateSpecialistRequest(name, speciality, systemPrompt))
                loadAgents()
            } catch (_: Exception) { }
        }
    }

    fun deleteSpecialist(name: String) {
        viewModelScope.launch {
            try {
                agentApi.deleteSpecialist(name)
                loadAgents()
            } catch (_: Exception) { }
        }
    }

    fun generateSpecialist(description: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isGenerating = true) }
            try {
                agentApi.generateSpecialist(GenerateRequest(description))
                loadAgents()
            } catch (_: Exception) { }
            _uiState.update { it.copy(isGenerating = false) }
        }
    }

    fun toggleOrchestrator(enabled: Boolean) {
        viewModelScope.launch {
            try {
                val config = _uiState.value.orchestrator.copy(enabled = enabled)
                agentApi.updateOrchestrator(config)
                _uiState.update { it.copy(orchestrator = config) }
            } catch (_: Exception) { }
        }
    }
}
