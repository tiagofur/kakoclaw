package com.makoclaw.feature.skills.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.Skill
import com.makoclaw.core.network.api.GenerateSkillRequest
import com.makoclaw.core.network.api.InstallSkillRequest
import com.makoclaw.core.network.api.SkillsApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class SkillsUiState(
    val installedSkills: List<Skill> = emptyList(),
    val marketplaceSkills: List<Skill> = emptyList(),
    val selectedTab: Int = 0, // 0=Installed, 1=Marketplace, 2=Generate
    val isLoading: Boolean = true,
    val isGenerating: Boolean = false,
    val generatedSkill: Skill? = null,
    val error: String? = null
)

@HiltViewModel
class SkillsViewModel @Inject constructor(
    private val skillsApi: SkillsApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(SkillsUiState())
    val uiState: StateFlow<SkillsUiState> = _uiState.asStateFlow()

    init { loadSkills() }

    fun loadSkills() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val installed = skillsApi.listSkills()
                _uiState.update {
                    it.copy(installedSkills = installed.skills, isLoading = false)
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun loadMarketplace() {
        viewModelScope.launch {
            try {
                val marketplace = skillsApi.getMarketplaceSkills()
                _uiState.update { it.copy(marketplaceSkills = marketplace.skills) }
            } catch (_: Exception) { }
        }
    }

    fun installSkill(repository: String) {
        viewModelScope.launch {
            try {
                skillsApi.installSkill(InstallSkillRequest(repository))
                loadSkills()
            } catch (_: Exception) { }
        }
    }

    fun uninstallSkill(name: String) {
        viewModelScope.launch {
            try {
                skillsApi.uninstallSkill(name)
                loadSkills()
            } catch (_: Exception) { }
        }
    }

    fun generateSkill(name: String, goal: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isGenerating = true) }
            try {
                val skill = skillsApi.generateSkill(GenerateSkillRequest(name = name, goal = goal))
                _uiState.update { it.copy(isGenerating = false, generatedSkill = skill) }
            } catch (e: Exception) {
                _uiState.update { it.copy(isGenerating = false, error = e.message) }
            }
        }
    }

    fun selectTab(tab: Int) {
        _uiState.update { it.copy(selectedTab = tab) }
        if (tab == 1) loadMarketplace()
    }
}
