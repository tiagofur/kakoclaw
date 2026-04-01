package com.makoclaw.feature.skills.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.Skill
import com.makoclaw.feature.skills.data.repository.SkillsRepository
import com.makoclaw.feature.skills.presentation.state.SkillsEffect
import com.makoclaw.feature.skills.presentation.state.SkillsEvent
import com.makoclaw.feature.skills.presentation.state.SkillsUiState
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.flow.flowOn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Coordinates state, CRUD, marketplace, and AI generation for skills. */
@HiltViewModel
class SkillsViewModel @Inject constructor(
    private val repository: SkillsRepository,
    private val dispatcher: CoroutineDispatcher = Dispatchers.IO
) : ViewModel() {

    private val _uiState = MutableStateFlow(SkillsUiState())
    private val _effects = MutableSharedFlow<SkillsEffect>()

    private var installedJob: Job? = null

    val uiState: StateFlow<SkillsUiState> = _uiState.asStateFlow()
    val effects: SharedFlow<SkillsEffect> = _effects.asSharedFlow()

    init {
        loadInstalled()
    }

    fun onEvent(event: SkillsEvent) {
        when (event) {
            SkillsEvent.LoadInstalled -> loadInstalled()
            SkillsEvent.LoadMarketplace -> loadMarketplace()
            SkillsEvent.Refresh -> {
                loadInstalled()
                loadMarketplace()
            }
            is SkillsEvent.SelectTab -> selectTab(event.tab)
            is SkillsEvent.InstallSkill -> installSkill(event.repository)
            is SkillsEvent.UninstallSkill -> uninstallSkill(event.name)
            is SkillsEvent.GenerateSkill -> generateSkill(event.name, event.goal)
            is SkillsEvent.UpdateSearchQuery -> updateSearchQuery(event.query)
            is SkillsEvent.SetCategoryFilter -> setCategoryFilter(event.category)
            SkillsEvent.ClearError -> clearError()
        }
    }

    private fun loadInstalled() {
        installedJob?.cancel()
        installedJob = viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            repository.getInstalled()
                .flowOn(dispatcher)
                .catch { throwable ->
                    _uiState.update {
                        it.copy(isLoading = false, error = throwable.message ?: "Failed to load skills")
                    }
                }
                .collectLatest { skills ->
                    _uiState.update { state ->
                        state.copy(
                            isLoading = false,
                            installedSkills = skills,
                            isEmpty = state.selectedTab == 0 && skills.isEmpty(),
                            error = null
                        )
                    }
                }
        }
    }

    private fun loadMarketplace() {
        viewModelScope.launch {
            repository.getMarketplace()
                .onSuccess { marketplaceSkills ->
                    _uiState.update { it.copy(marketplaceSkills = marketplaceSkills) }
                }
                .onFailure { throwable ->
                    _effects.emit(SkillsEffect.ShowSnackbar("Failed to load marketplace: ${throwable.message}"))
                }
        }
    }

    private fun selectTab(tab: Int) {
        _uiState.update { it.copy(selectedTab = tab) }
        if (tab == 1 && _uiState.value.marketplaceSkills.isEmpty()) {
            loadMarketplace()
        }
    }

    private fun installSkill(repository: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isInstalling = true) }
            repository.installSkill(repository)
                .onSuccess {
                    _uiState.update { it.copy(isInstalling = false) }
                    _effects.emit(SkillsEffect.SkillInstalled(repository))
                    _effects.emit(SkillsEffect.ShowSnackbar("Skill installed"))
                }
                .onFailure { throwable ->
                    _uiState.update { it.copy(isInstalling = false) }
                    _effects.emit(SkillsEffect.ShowSnackbar("Install failed: ${throwable.message}"))
                }
        }
    }

    private fun uninstallSkill(name: String) {
        viewModelScope.launch {
            repository.uninstallSkill(name)
                .onSuccess {
                    _effects.emit(SkillsEffect.SkillUninstalled(name))
                    _effects.emit(SkillsEffect.ShowSnackbar("Skill uninstalled"))
                }
                .onFailure { throwable ->
                    _effects.emit(SkillsEffect.ShowSnackbar("Uninstall failed: ${throwable.message}"))
                }
        }
    }

    private fun generateSkill(name: String, goal: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isGenerating = true) }
            repository.generateSkill(name, goal)
                .onSuccess { skill ->
                    _uiState.update { it.copy(isGenerating = false, generatedSkill = skill) }
                    _effects.emit(SkillsEffect.SkillGenerated(skill))
                    _effects.emit(SkillsEffect.ShowSnackbar("Skill generated: ${skill.name}"))
                }
                .onFailure { throwable ->
                    _uiState.update { it.copy(isGenerating = false) }
                    _effects.emit(SkillsEffect.ShowSnackbar("Generation failed: ${throwable.message}"))
                }
        }
    }

    private fun updateSearchQuery(query: String) {
        _uiState.update { it.copy(searchQuery = query) }
    }

    private fun setCategoryFilter(category: String?) {
        _uiState.update { it.copy(categoryFilter = category) }
    }

    private fun clearError() {
        _uiState.update { it.copy(error = null) }
    }
}
