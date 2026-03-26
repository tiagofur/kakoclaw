package com.makoclaw.feature.mcp.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.McpServer
import com.makoclaw.core.network.api.McpApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class McpUiState(
    val servers: List<McpServer> = emptyList(),
    val isLoading: Boolean = true,
    val isReconnecting: Boolean = false,
    val error: String? = null
)

@HiltViewModel
class McpViewModel @Inject constructor(
    private val mcpApi: McpApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(McpUiState())
    val uiState: StateFlow<McpUiState> = _uiState.asStateFlow()

    init { loadServers() }

    fun loadServers() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val response = mcpApi.listServers()
                _uiState.update { it.copy(servers = response.servers, isLoading = false) }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun reconnectServer(name: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isReconnecting = true) }
            try {
                mcpApi.reconnectServer(name)
                loadServers()
            } catch (_: Exception) { }
            _uiState.update { it.copy(isReconnecting = false) }
        }
    }

    fun reconnectAll() {
        viewModelScope.launch {
            _uiState.update { it.copy(isReconnecting = true) }
            try {
                mcpApi.reconnectAll()
                loadServers()
            } catch (_: Exception) { }
            _uiState.update { it.copy(isReconnecting = false) }
        }
    }
}
