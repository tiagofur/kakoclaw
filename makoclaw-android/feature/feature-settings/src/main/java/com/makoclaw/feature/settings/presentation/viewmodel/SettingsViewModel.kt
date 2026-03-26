package com.makoclaw.feature.settings.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.ChangePasswordRequest
import com.makoclaw.core.model.User
import com.makoclaw.core.network.api.AuthApi
import com.makoclaw.core.network.api.ConfigApi
import com.makoclaw.core.network.api.ModelsResponse
import com.makoclaw.core.security.TokenManager
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

data class SettingsUiState(
    val user: User? = null,
    val models: ModelsResponse = ModelsResponse(),
    val tools: List<String> = emptyList(),
    val selectedTab: Int = 0,
    val isLoading: Boolean = true,
    val error: String? = null
)

sealed interface SettingsEffect {
    data class ShowMessage(val message: String) : SettingsEffect
    data object Logout : SettingsEffect
}

@HiltViewModel
class SettingsViewModel @Inject constructor(
    private val authApi: AuthApi,
    private val configApi: ConfigApi,
    private val tokenManager: TokenManager
) : ViewModel() {

    private val _uiState = MutableStateFlow(SettingsUiState())
    val uiState: StateFlow<SettingsUiState> = _uiState.asStateFlow()

    private val _effects = MutableSharedFlow<SettingsEffect>()
    val effects: SharedFlow<SettingsEffect> = _effects.asSharedFlow()

    init { loadProfile() }

    fun loadProfile() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val user = authApi.getCurrentUser()
                _uiState.update { it.copy(user = user, isLoading = false) }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun loadProviders() {
        viewModelScope.launch {
            try {
                val models = configApi.getModels()
                _uiState.update { it.copy(models = models) }
            } catch (_: Exception) { }
        }
    }

    fun loadTools() {
        viewModelScope.launch {
            try {
                val tools = configApi.getTools()
                _uiState.update { it.copy(tools = tools["tools"] ?: emptyList()) }
            } catch (_: Exception) { }
        }
    }

    fun changePassword(currentPassword: String, newPassword: String) {
        viewModelScope.launch {
            try {
                authApi.changePassword(ChangePasswordRequest(currentPassword, newPassword))
                _effects.emit(SettingsEffect.ShowMessage("Password changed successfully"))
            } catch (e: Exception) {
                _effects.emit(SettingsEffect.ShowMessage("Failed: ${e.message}"))
            }
        }
    }

    fun logout() {
        tokenManager.clearToken()
        viewModelScope.launch {
            _effects.emit(SettingsEffect.Logout)
        }
    }

    fun selectTab(tab: Int) {
        _uiState.update { it.copy(selectedTab = tab) }
        when (tab) {
            1 -> loadProviders()
            4 -> loadTools()
        }
    }
}
