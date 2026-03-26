package com.makoclaw.core.security

import com.makoclaw.core.security.storage.JwtStorage
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TokenManager @Inject constructor(
    private val jwtStorage: JwtStorage
) {
    private val prefs get() = jwtStorage.sharedPreferences

    companion object {
        private const val KEY_TOKEN = "jwt_token"
        private const val KEY_USERNAME = "username"
        private const val KEY_ROLE = "user_role"
        private const val KEY_EXPIRY = "token_expiry"
    }

    fun saveToken(token: String, username: String, role: String, expiresInMs: Long) {
        jwtStorage.saveJwt(token)
        prefs.edit()
            .putString(KEY_USERNAME, username)
            .putString(KEY_ROLE, role)
            .putLong(KEY_EXPIRY, System.currentTimeMillis() + expiresInMs)
            .apply()
    }

    fun getToken(): String? = jwtStorage.getJwt()

    fun getUsername(): String? = prefs.getString(KEY_USERNAME, null)

    fun getRole(): String? = prefs.getString(KEY_ROLE, null)

    fun hasValidToken(): Boolean {
        val token = getToken() ?: return false
        val expiry = prefs.getLong(KEY_EXPIRY, 0)
        return token.isNotEmpty() && System.currentTimeMillis() < expiry
    }

    fun isTokenExpired(): Boolean {
        val expiry = prefs.getLong(KEY_EXPIRY, 0)
        return System.currentTimeMillis() >= expiry
    }

    fun clearToken() {
        jwtStorage.clearTokens()
        prefs.edit()
            .remove(KEY_TOKEN)
            .remove(KEY_USERNAME)
            .remove(KEY_ROLE)
            .remove(KEY_EXPIRY)
            .apply()
    }
}
