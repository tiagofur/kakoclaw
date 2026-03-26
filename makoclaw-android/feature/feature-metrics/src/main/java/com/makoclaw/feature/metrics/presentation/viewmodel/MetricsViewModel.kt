package com.makoclaw.feature.metrics.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.network.api.MetricsApi
import com.makoclaw.core.network.api.MetricsResponse
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class MetricsUiState(
    val metrics: MetricsResponse = MetricsResponse(),
    val isLoading: Boolean = true,
    val error: String? = null
)

@HiltViewModel
class MetricsViewModel @Inject constructor(
    private val metricsApi: MetricsApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(MetricsUiState())
    val uiState: StateFlow<MetricsUiState> = _uiState.asStateFlow()

    init { loadMetrics() }

    fun loadMetrics() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val metrics = metricsApi.getMetrics()
                _uiState.update { it.copy(metrics = metrics, isLoading = false) }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }
}
