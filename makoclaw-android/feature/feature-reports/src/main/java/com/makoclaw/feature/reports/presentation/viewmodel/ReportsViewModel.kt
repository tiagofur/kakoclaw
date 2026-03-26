package com.makoclaw.feature.reports.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.network.api.EmailReportRequest
import com.makoclaw.core.network.api.ReportsApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ReportsUiState(
    val to: String = "",
    val subject: String = "",
    val content: String = "",
    val isSending: Boolean = false
)

@HiltViewModel
class ReportsViewModel @Inject constructor(
    private val reportsApi: ReportsApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(ReportsUiState())
    val uiState: StateFlow<ReportsUiState> = _uiState.asStateFlow()

    private val _effects = MutableSharedFlow<String>()
    val effects: SharedFlow<String> = _effects.asSharedFlow()

    fun updateTo(to: String) { _uiState.update { it.copy(to = to) } }
    fun updateSubject(subject: String) { _uiState.update { it.copy(subject = subject) } }
    fun updateContent(content: String) { _uiState.update { it.copy(content = content) } }

    fun sendReport() {
        val state = _uiState.value
        if (state.to.isBlank()) return

        viewModelScope.launch {
            _uiState.update { it.copy(isSending = true) }
            try {
                reportsApi.emailReport(
                    EmailReportRequest(to = state.to, subject = state.subject, content = state.content)
                )
                _effects.emit("Report sent successfully")
                _uiState.update { ReportsUiState() } // reset form
            } catch (e: Exception) {
                _effects.emit("Failed: ${e.message}")
            }
            _uiState.update { it.copy(isSending = false) }
        }
    }
}
