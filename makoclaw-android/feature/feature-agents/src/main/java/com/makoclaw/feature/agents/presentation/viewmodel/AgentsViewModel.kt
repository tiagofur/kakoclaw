package com.makoclaw.feature.agents.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.OrchestratorConfig
import com.makoclaw.core.model.Specialist
import com.makoclaw.feature.agents.data.repository.AgentsRepository
import com.makoclaw.feature.agents.presentation.state.AgentsEffect
import com.makoclaw.feature.agents.presentation.state.AgentsEvent
import com.makoclaw.feature.agents.presentation.state.AgentsUiState
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Coordinates state, CRUD, AI generation, and orchestrator for agents. */
@HiltViewModel
class AgentsViewModel @Inject constructor(
    private val repository: AgentsRepository,
    private val dispatcher: CoroutineDispatcher = Dispatchers.IO
) : ViewModel() {

    private val _uiState = MutableStateFlow(AgentsUiState())
    private val _effects = MutableSharedFlow<AgentsEffect>()

    val uiState: StateFlow<AgentsUiState> = _uiState.asStateFlow()
    val effects: SharedFlow<AgentsEffect> = _effects.asSharedFlow()

    init {
        loadAgents()
        loadMetrics()
        loadSwarms()
    }

    fun onEvent(event: AgentsEvent) {
        when (event) {
            AgentsEvent.LoadAgents -> loadAgents()
            AgentsEvent.LoadMetrics -> loadMetrics()
            AgentsEvent.LoadSwarms -> loadSwarms()
            AgentsEvent.Refresh -> {
                loadAgents()
                loadMetrics()
                loadSwarms()
            }
            is AgentsEvent.CreateSpecialist -> createSpecialist(event)
            is AgentsEvent.EditSpecialist -> editSpecialist(event)
            is AgentsEvent.DeleteSpecialist -> deleteSpecialist(event.name)
            is AgentsEvent.GenerateSpecialist -> generateSpecialist(event.description)
            is AgentsEvent.ToggleOrchestrator -> toggleOrchestrator(event.enabled)
            is AgentsEvent.SelectSpecialist -> selectSpecialist(event.name)
            AgentsEvent.ClearSelection -> clearSelection()
            AgentsEvent.ClearError -> clearError()
        }
    }

    private fun loadAgents() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            try {
                repository.getOrchestrator()
                    .onSuccess { orchestrator ->
                        _uiState.update { it.copy(orchestrator = orchestrator) }
                    }
            } catch (_: Exception) { }

            repository.getSpecialists().collect { specialists ->
                _uiState.update { state ->
                    state.copy(
                        isLoading = false,
                        specialists = specialists,
                        selectedSpecialist = syncSelection(state.selectedSpecialist, specialists),
                        error = null
                    )
                }
            }
        }
    }

    private fun loadMetrics() {
        viewModelScope.launch(dispatcher) {
            repository.getMetrics()
                .onSuccess { metrics ->
                    _uiState.update { it.copy(metrics = metrics) }
                }
        }
    }

    private fun loadSwarms() {
        viewModelScope.launch(dispatcher) {
            repository.getSwarms()
                .onSuccess { swarms ->
                    _uiState.update { it.copy(swarms = swarms) }
                }
        }
    }

    private fun createSpecialist(event: AgentsEvent.CreateSpecialist) {
        viewModelScope.launch {
            repository.createSpecialist(event.name, event.speciality, event.systemPrompt)
                .onSuccess { specialist ->
                    _effects.emit(AgentsEffect.SpecialistCreated(specialist.name))
                    _effects.emit(AgentsEffect.ShowSnackbar("Specialist created: ${specialist.name}"))
                    loadAgents()
                }
                .onFailure { throwable ->
                    _effects.emit(AgentsEffect.ShowSnackbar("Create failed: ${throwable.message}"))
                }
        }
    }

    private fun editSpecialist(event: AgentsEvent.EditSpecialist) {
        viewModelScope.launch {
            repository.updateSpecialist(event.name, event.speciality, event.systemPrompt)
                .onSuccess {
                    _uiState.update { it.copy(isEditing = false) }
                    _effects.emit(AgentsEffect.ShowSnackbar("Specialist updated"))
                    loadAgents()
                }
                .onFailure { throwable ->
                    _effects.emit(AgentsEffect.ShowSnackbar("Update failed: ${throwable.message}"))
                }
        }
    }

    private fun deleteSpecialist(name: String) {
        viewModelScope.launch {
            repository.deleteSpecialist(name)
                .onSuccess {
                    _uiState.update { state ->
                        state.copy(
                            selectedSpecialist = state.selectedSpecialist?.takeUnless { it.name == name }
                        )
                    }
                    _effects.emit(AgentsEffect.SpecialistDeleted(name))
                    _effects.emit(AgentsEffect.ShowSnackbar("Specialist deleted"))
                }
                .onFailure { throwable ->
                    _effects.emit(AgentsEffect.ShowSnackbar("Delete failed: ${throwable.message}"))
                }
        }
    }

    private fun generateSpecialist(description: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isGenerating = true) }
            repository.generateSpecialist(description)
                .onSuccess { specialist ->
                    _uiState.update { it.copy(isGenerating = false) }
                    _effects.emit(AgentsEffect.SpecialistGenerated(specialist.name))
                    _effects.emit(AgentsEffect.ShowSnackbar("Generated: ${specialist.name}"))
                    loadAgents()
                }
                .onFailure { throwable ->
                    _uiState.update { it.copy(isGenerating = false) }
                    _effects.emit(AgentsEffect.ShowSnackbar("Generation failed: ${throwable.message}"))
                }
        }
    }

    private fun toggleOrchestrator(enabled: Boolean) {
        viewModelScope.launch {
            val config = _uiState.value.orchestrator.copy(enabled = enabled)
            repository.toggleOrchestrator(config)
                .onSuccess {
                    _uiState.update { it.copy(orchestrator = config) }
                    _effects.emit(AgentsEffect.OrchestratorToggled(enabled))
                    _effects.emit(AgentsEffect.ShowSnackbar(if (enabled) "Orchestrator enabled" else "Orchestrator disabled"))
                }
                .onFailure { throwable ->
                    _effects.emit(AgentsEffect.ShowSnackbar("Toggle failed: ${throwable.message}"))
                }
        }
    }

    private fun selectSpecialist(name: String) {
        val selected = _uiState.value.specialists.firstOrNull { it.name == name }
        _uiState.update { it.copy(selectedSpecialist = selected) }
    }

    private fun clearSelection() {
        _uiState.update { it.copy(selectedSpecialist = null, isEditing = false) }
    }

    private fun clearError() {
        _uiState.update { it.copy(error = null) }
    }

    private fun syncSelection(
        selected: Specialist?,
        specialists: List<Specialist>
    ): Specialist? = selected?.let { current ->
        specialists.firstOrNull { it.name == current.name }
    }
}
