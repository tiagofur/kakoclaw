package com.makoclaw.feature.auth.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.RegisterRequest
import com.makoclaw.core.network.api.AuthApi
import com.makoclaw.feature.auth.presentation.state.SignupUiState
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

sealed interface SignupEffect {
    data object NavigateToLogin : SignupEffect
    data class ShowError(val message: String) : SignupEffect
}

@HiltViewModel
class SignupViewModel @Inject constructor(
    private val authApi: AuthApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(SignupUiState())
    val uiState: StateFlow<SignupUiState> = _uiState.asStateFlow()

    private val _effects = MutableSharedFlow<SignupEffect>()
    val effects: SharedFlow<SignupEffect> = _effects.asSharedFlow()

    fun onUsernameChanged(value: String) { _uiState.update { it.copy(username = value) } }
    fun onEmailChanged(value: String) { _uiState.update { it.copy(email = value) } }
    fun onPasswordChanged(value: String) { _uiState.update { it.copy(password = value) } }
    fun onConfirmPasswordChanged(value: String) { _uiState.update { it.copy(confirmPassword = value) } }

    fun onSignup() {
        val state = _uiState.value
        if (state.username.isBlank() || state.email.isBlank() || state.password.isBlank()) {
            viewModelScope.launch { _effects.emit(SignupEffect.ShowError("Please fill in all fields")) }
            return
        }
        if (state.password != state.confirmPassword) {
            viewModelScope.launch { _effects.emit(SignupEffect.ShowError("Passwords do not match")) }
            return
        }
        if (state.password.length < 8) {
            viewModelScope.launch { _effects.emit(SignupEffect.ShowError("Password must be at least 8 characters")) }
            return
        }

        _uiState.update { it.copy(isLoading = true, error = null) }

        viewModelScope.launch {
            try {
                authApi.register(RegisterRequest(state.username, state.email, state.password))
                _uiState.update { it.copy(isLoading = false) }
                _effects.emit(SignupEffect.NavigateToLogin)
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
                _effects.emit(SignupEffect.ShowError(e.message ?: "Registration failed"))
            }
        }
    }
}
