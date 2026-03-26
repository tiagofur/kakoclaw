package com.makoclaw.feature.cron.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.CronJob
import com.makoclaw.core.network.api.AiCronRequest
import com.makoclaw.core.network.api.CreateCronRequest
import com.makoclaw.core.network.api.CronApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class CronUiState(
    val jobs: List<CronJob> = emptyList(),
    val isLoading: Boolean = true,
    val error: String? = null,
    val aiSchedule: String = "",
    val aiDescription: String = ""
)

@HiltViewModel
class CronViewModel @Inject constructor(
    private val cronApi: CronApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(CronUiState())
    val uiState: StateFlow<CronUiState> = _uiState.asStateFlow()

    init { loadJobs() }

    fun loadJobs() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val response = cronApi.listJobs()
                _uiState.update { it.copy(jobs = response.jobs, isLoading = false) }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun createJob(name: String, schedule: String, message: String) {
        viewModelScope.launch {
            try {
                cronApi.createJob(CreateCronRequest(name, schedule, message))
                loadJobs()
            } catch (_: Exception) { }
        }
    }

    fun deleteJob(id: String) {
        viewModelScope.launch {
            try {
                cronApi.deleteJob(id)
                loadJobs()
            } catch (_: Exception) { }
        }
    }

    fun toggleJob(id: String, enabled: Boolean) {
        viewModelScope.launch {
            try {
                cronApi.toggleJob(id, mapOf("enabled" to enabled))
                loadJobs()
            } catch (_: Exception) { }
        }
    }

    fun generateCron(prompt: String) {
        viewModelScope.launch {
            try {
                val response = cronApi.generateCron(AiCronRequest(prompt))
                _uiState.update {
                    it.copy(aiSchedule = response.schedule, aiDescription = response.description)
                }
            } catch (_: Exception) { }
        }
    }
}
