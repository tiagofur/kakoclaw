package com.makoclaw.android.navigation

import androidx.lifecycle.ViewModel
import com.makoclaw.core.security.TokenManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject

@HiltViewModel
class AuthNavigationViewModel @Inject constructor(
    private val tokenManager: TokenManager
) : ViewModel() {

    private val _isAuthenticated = MutableStateFlow(tokenManager.hasValidToken())
    val isAuthenticated: StateFlow<Boolean> = _isAuthenticated.asStateFlow()

    fun onAuthStateChanged(authenticated: Boolean) {
        _isAuthenticated.value = authenticated
    }
}
