package com.makoclaw.feature.auth.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.LoginRequest
import com.makoclaw.core.network.api.AuthApi
import com.makoclaw.core.security.TokenManager
import com.makoclaw.feature.auth.presentation.state.LoginUiState
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

sealed interface LoginEffect {
    data object NavigateToMain : LoginEffect
    data class ShowError(val message: String) : LoginEffect
}

@HiltViewModel
class LoginViewModel @Inject constructor(
    private val authApi: AuthApi,
    private val tokenManager: TokenManager
) : ViewModel() {

    private val _uiState = MutableStateFlow(LoginUiState())
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()

    private val _effects = MutableSharedFlow<LoginEffect>()
    val effects: SharedFlow<LoginEffect> = _effects.asSharedFlow()

    fun onUsernameChanged(username: String) {
        _uiState.update { it.copy(username = username) }
    }

    fun onPasswordChanged(password: String) {
        _uiState.update { it.copy(password = password) }
    }

    fun onServerUrlChanged(url: String) {
        _uiState.update { it.copy(serverUrl = url) }
    }

    fun onLogin() {
        val state = _uiState.value
        if (state.username.isBlank() || state.password.isBlank()) {
            viewModelScope.launch {
                _effects.emit(LoginEffect.ShowError("Please fill in all fields"))
            }
            return
        }

        _uiState.update { it.copy(isLoading = true, error = null) }

        viewModelScope.launch {
            try {
                val response = authApi.login(
                    LoginRequest(state.username, state.password)
                )
                val user = authApi.getCurrentUser()
                tokenManager.saveToken(
                    token = response.token,
                    username = user.username,
                    role = user.role,
                    expiresInMs = 24 * 60 * 60 * 1000L // 24h default
                )
                _uiState.update { it.copy(isLoading = false) }
                _effects.emit(LoginEffect.NavigateToMain)
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        error = e.message ?: "Login failed"
                    )
                }
                _effects.emit(LoginEffect.ShowError(e.message ?: "Login failed"))
            }
        }
    }
}
