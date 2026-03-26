package com.makoclaw.android.navigation

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.datastore.PreferencesDataStore
import com.makoclaw.core.security.TokenManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class AuthNavigationViewModel @Inject constructor(
    private val tokenManager: TokenManager,
    private val preferencesDataStore: PreferencesDataStore
) : ViewModel() {

    private val _isAuthenticated = MutableStateFlow(tokenManager.hasValidToken())
    val isAuthenticated: StateFlow<Boolean> = _isAuthenticated.asStateFlow()

    private val _hasServerUrl = MutableStateFlow(true) // assume true until checked
    val hasServerUrl: StateFlow<Boolean> = _hasServerUrl.asStateFlow()

    init {
        viewModelScope.launch {
            val prefs = preferencesDataStore.preferences.first()
            _hasServerUrl.value = prefs.serverUrl != "http://localhost:8080"
                    && prefs.serverUrl.isNotBlank()
        }
    }

    fun onAuthStateChanged(authenticated: Boolean) {
        _isAuthenticated.value = authenticated
    }

    fun saveServerUrl(url: String) {
        viewModelScope.launch {
            preferencesDataStore.updateServerUrl(url)
            _hasServerUrl.value = true
        }
    }
}
