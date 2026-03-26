package com.makoclaw.feature.history.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.ChatSession
import com.makoclaw.core.network.api.ChatApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class HistoryUiState(
    val sessions: List<ChatSession> = emptyList(),
    val filteredSessions: List<ChatSession> = emptyList(),
    val searchQuery: String = "",
    val isLoading: Boolean = true,
    val error: String? = null
)

@HiltViewModel
class HistoryViewModel @Inject constructor(
    private val chatApi: ChatApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(HistoryUiState())
    val uiState: StateFlow<HistoryUiState> = _uiState.asStateFlow()

    init { loadSessions() }

    fun loadSessions() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val response = chatApi.getSessions()
                _uiState.update {
                    it.copy(
                        sessions = response.sessions,
                        filteredSessions = response.sessions,
                        isLoading = false
                    )
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun search(query: String) {
        _uiState.update { state ->
            val filtered = if (query.isBlank()) state.sessions
            else state.sessions.filter {
                it.title.contains(query, ignoreCase = true) ||
                it.lastMessage.contains(query, ignoreCase = true)
            }
            state.copy(searchQuery = query, filteredSessions = filtered)
        }
    }

    fun deleteSession(sessionId: String) {
        viewModelScope.launch {
            try {
                chatApi.deleteSession(sessionId)
                loadSessions()
            } catch (_: Exception) { }
        }
    }
}
