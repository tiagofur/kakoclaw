package com.makoclaw.core.model

import kotlinx.serialization.Serializable

@Serializable
data class User(
    val username: String,
    val email: String = "",
    val role: String = "user",
    val onboardingCompleted: Boolean = false
)

@Serializable
data class LoginRequest(
    val username: String,
    val password: String
)

@Serializable
data class LoginResponse(
    val token: String
)

@Serializable
data class RegisterRequest(
    val username: String,
    val email: String,
    val password: String
)

@Serializable
data class ChangePasswordRequest(
    val old_password: String,
    val new_password: String
)
