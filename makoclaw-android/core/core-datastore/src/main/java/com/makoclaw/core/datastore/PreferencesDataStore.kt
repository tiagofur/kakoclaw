package com.makoclaw.core.datastore

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "makoclaw_prefs")

@Singleton
class PreferencesDataStore @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private object Keys {
        val SERVER_URL = stringPreferencesKey("server_url")
        val THEME = stringPreferencesKey("theme")
        val NOTIFICATIONS = booleanPreferencesKey("notifications_enabled")
        val DYNAMIC_COLOR = booleanPreferencesKey("dynamic_color")
        val BIOMETRIC_AUTH = booleanPreferencesKey("biometric_auth")
    }

    val preferences: Flow<UserPreferences> = context.dataStore.data.map { prefs ->
        UserPreferences(
            serverUrl = prefs[Keys.SERVER_URL] ?: "http://localhost:8080",
            theme = ThemeMode.valueOf(prefs[Keys.THEME] ?: "SYSTEM"),
            notificationsEnabled = prefs[Keys.NOTIFICATIONS] ?: true,
            useDynamicColor = prefs[Keys.DYNAMIC_COLOR] ?: true,
            useBiometricAuth = prefs[Keys.BIOMETRIC_AUTH] ?: false
        )
    }

    suspend fun updateServerUrl(url: String) {
        context.dataStore.edit { it[Keys.SERVER_URL] = url }
    }

    suspend fun updateTheme(theme: ThemeMode) {
        context.dataStore.edit { it[Keys.THEME] = theme.name }
    }

    suspend fun updateNotifications(enabled: Boolean) {
        context.dataStore.edit { it[Keys.NOTIFICATIONS] = enabled }
    }

    suspend fun updateDynamicColor(enabled: Boolean) {
        context.dataStore.edit { it[Keys.DYNAMIC_COLOR] = enabled }
    }

    suspend fun updateBiometricAuth(enabled: Boolean) {
        context.dataStore.edit { it[Keys.BIOMETRIC_AUTH] = enabled }
    }

    fun getServerUrl(): Flow<String> = context.dataStore.data.map { prefs ->
        prefs[Keys.SERVER_URL] ?: "http://localhost:8080"
    }
}
