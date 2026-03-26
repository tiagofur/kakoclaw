package com.makoclaw.feature.auth.data.repository

import com.makoclaw.core.common.Result
import com.makoclaw.core.model.LoginRequest
import com.makoclaw.core.model.RegisterRequest
import com.makoclaw.core.model.User
import com.makoclaw.core.network.api.AuthApi
import com.makoclaw.core.security.TokenManager
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepository @Inject constructor(
    private val authApi: AuthApi,
    private val tokenManager: TokenManager
) {
    suspend fun login(username: String, password: String): Result<User> {
        return try {
            val response = authApi.login(LoginRequest(username, password))
            val user = authApi.getCurrentUser()
            tokenManager.saveToken(
                token = response.token,
                username = user.username,
                role = user.role,
                expiresInMs = 24 * 60 * 60 * 1000L
            )
            Result.Success(user)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Login failed", e)
        }
    }

    suspend fun register(username: String, email: String, password: String): Result<Unit> {
        return try {
            authApi.register(RegisterRequest(username, email, password))
            Result.Success(Unit)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Registration failed", e)
        }
    }

    suspend fun getCurrentUser(): Result<User> {
        return try {
            Result.Success(authApi.getCurrentUser())
        } catch (e: Exception) {
            Result.Error(e.message ?: "Failed to get user", e)
        }
    }

    fun isAuthenticated(): Boolean = tokenManager.hasValidToken()

    fun logout() = tokenManager.clearToken()
}
