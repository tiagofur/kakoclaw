package com.makoclaw.feature.metrics.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.MetricsData
import com.makoclaw.core.model.TimeRange
import com.makoclaw.feature.metrics.data.repository.MetricsRepository
import com.makoclaw.feature.metrics.presentation.state.MetricsEffect
import com.makoclaw.feature.metrics.presentation.state.MetricsEvent
import com.makoclaw.feature.metrics.presentation.state.MetricsUiState
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.flow.flowOn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Coordinates state, filtering, and export for the metrics screen. */
@HiltViewModel
class MetricsViewModel @Inject constructor(
    private val repository: MetricsRepository,
    private val dispatcher: CoroutineDispatcher = Dispatchers.IO
) : ViewModel() {

    private val _uiState = MutableStateFlow(MetricsUiState())
    private val _effects = MutableSharedFlow<MetricsEffect>()

    val uiState: StateFlow<MetricsUiState> = _uiState.asStateFlow()
    val effects: SharedFlow<MetricsEffect> = _effects.asSharedFlow()

    init {
        loadMetrics()
    }

    fun onEvent(event: MetricsEvent) {
        when (event) {
            MetricsEvent.LoadMetrics,
            MetricsEvent.Refresh -> loadMetrics()
            is MetricsEvent.SetTimeRange -> setTimeRange(event.timeRange)
            is MetricsEvent.SetAgentFilter -> setAgentFilter(event.agentId)
            MetricsEvent.ExportMetrics -> exportMetrics()
            MetricsEvent.ClearExport -> clearExport()
        }
    }

    private fun loadMetrics() {
        val state = _uiState.value
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            repository.getMetrics(state.selectedTimeRange, state.selectedAgentId)
                .flowOn(dispatcher)
                .catch { throwable ->
                    _uiState.update {
                        it.copy(isLoading = false, error = throwable.message ?: "Failed to load metrics")
                    }
                }
                .collectLatest { data ->
                    _uiState.update {
                        it.copy(
                            isLoading = false,
                            metricsData = data,
                            error = null,
                            availableAgents = extractAgents(data)
                        )
                    }
                }
        }
    }

    private fun setTimeRange(timeRange: TimeRange) {
        _uiState.update { it.copy(selectedTimeRange = timeRange) }
        loadMetrics()
    }

    private fun setAgentFilter(agentId: String?) {
        _uiState.update { it.copy(selectedAgentId = agentId) }
        loadMetrics()
    }

    private fun exportMetrics() {
        val state = _uiState.value
        viewModelScope.launch {
            _uiState.update { it.copy(isExporting = true) }
            repository.exportMetrics(state.selectedTimeRange, state.selectedAgentId)
                .flowOn(dispatcher)
                .catch { throwable ->
                    _uiState.update { it.copy(isExporting = false) }
                    _effects.emit(MetricsEffect.ShowSnackbar("Export failed: ${throwable.message}"))
                }
                .collectLatest { result ->
                    _uiState.update {
                        it.copy(isExporting = false, exportFilePath = result.filePath)
                    }
                    _effects.emit(MetricsEffect.ExportComplete(result.filePath))
                }
        }
    }

    private fun clearExport() {
        _uiState.update { it.copy(exportFilePath = null) }
    }

    private fun extractAgents(data: MetricsData): List<String> =
        data.agentLabels.ifEmpty {
            data.costDistribution.keys.toList()
        }
}
