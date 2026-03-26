package com.makoclaw.feature.auth.presentation.state

data class LoginUiState(
    val username: String = "",
    val password: String = "",
    val serverUrl: String = "http://localhost:8080",
    val isLoading: Boolean = false,
    val error: String? = null,
    val showServerConfig: Boolean = false
)

data class SignupUiState(
    val username: String = "",
    val email: String = "",
    val password: String = "",
    val confirmPassword: String = "",
    val isLoading: Boolean = false,
    val error: String? = null
)

data class OnboardingUiState(
    val currentStep: Int = 0,
    val totalSteps: Int = 6,
    val providerName: String = "",
    val apiKey: String = "",
    val isValidating: Boolean = false,
    val isComplete: Boolean = false,
    val error: String? = null
)
