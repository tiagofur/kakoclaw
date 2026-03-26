package com.makoclaw.core.datastore

data class UserPreferences(
    val serverUrl: String = "http://localhost:8080",
    val theme: ThemeMode = ThemeMode.SYSTEM,
    val notificationsEnabled: Boolean = true,
    val useDynamicColor: Boolean = true,
    val useBiometricAuth: Boolean = false
)

enum class ThemeMode {
    LIGHT, DARK, SYSTEM
}
