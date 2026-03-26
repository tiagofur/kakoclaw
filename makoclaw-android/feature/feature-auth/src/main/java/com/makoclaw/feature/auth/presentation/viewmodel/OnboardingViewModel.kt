package com.makoclaw.feature.auth.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.network.api.ConfigApi
import com.makoclaw.core.network.api.ValidateProviderRequest
import com.makoclaw.feature.auth.presentation.state.OnboardingUiState
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

sealed interface OnboardingEffect {
    data object Complete : OnboardingEffect
    data class ShowError(val message: String) : OnboardingEffect
}

@HiltViewModel
class OnboardingViewModel @Inject constructor(
    private val configApi: ConfigApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(OnboardingUiState())
    val uiState: StateFlow<OnboardingUiState> = _uiState.asStateFlow()

    private val _effects = MutableSharedFlow<OnboardingEffect>()
    val effects: SharedFlow<OnboardingEffect> = _effects.asSharedFlow()

    fun onProviderNameChanged(value: String) { _uiState.update { it.copy(providerName = value) } }
    fun onApiKeyChanged(value: String) { _uiState.update { it.copy(apiKey = value) } }

    fun nextStep() {
        val state = _uiState.value
        if (state.currentStep < state.totalSteps - 1) {
            _uiState.update { it.copy(currentStep = state.currentStep + 1) }
        } else {
            completeOnboarding()
        }
    }

    fun previousStep() {
        val state = _uiState.value
        if (state.currentStep > 0) {
            _uiState.update { it.copy(currentStep = state.currentStep - 1) }
        }
    }

    fun validateProvider() {
        val state = _uiState.value
        if (state.providerName.isBlank() || state.apiKey.isBlank()) return

        _uiState.update { it.copy(isValidating = true, error = null) }
        viewModelScope.launch {
            try {
                configApi.validateProvider(
                    ValidateProviderRequest(state.providerName, state.apiKey)
                )
                _uiState.update { it.copy(isValidating = false) }
                nextStep()
            } catch (e: Exception) {
                _uiState.update { it.copy(isValidating = false, error = e.message) }
                _effects.emit(OnboardingEffect.ShowError(e.message ?: "Validation failed"))
            }
        }
    }

    private fun completeOnboarding() {
        viewModelScope.launch {
            try {
                configApi.completeOnboarding()
                _uiState.update { it.copy(isComplete = true) }
                _effects.emit(OnboardingEffect.Complete)
            } catch (e: Exception) {
                _effects.emit(OnboardingEffect.ShowError(e.message ?: "Failed to complete"))
            }
        }
    }
}
