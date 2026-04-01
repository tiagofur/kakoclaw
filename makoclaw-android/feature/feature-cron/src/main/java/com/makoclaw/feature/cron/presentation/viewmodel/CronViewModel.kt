package com.makoclaw.feature.cron.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.feature.cron.data.repository.CronRepository
import com.makoclaw.feature.cron.presentation.state.CronEffect
import com.makoclaw.feature.cron.presentation.state.CronEvent
import com.makoclaw.feature.cron.presentation.state.CronUiState
import com.makoclaw.feature.cron.presentation.state.TestRunState
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
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

/** Coordinates state, CRUD, AI generation, and test runs for cron jobs. */
@HiltViewModel
class CronViewModel @Inject constructor(
    private val repository: CronRepository,
    private val dispatcher: CoroutineDispatcher = Dispatchers.IO
) : ViewModel() {

    private val _uiState = MutableStateFlow(CronUiState())
    private val _effects = MutableSharedFlow<CronEffect>()

    private var jobsJob: Job? = null

    val uiState: StateFlow<CronUiState> = _uiState.asStateFlow()
    val effects: SharedFlow<CronEffect> = _effects.asSharedFlow()

    init {
        loadJobs()
    }

    fun onEvent(event: CronEvent) {
        when (event) {
            CronEvent.LoadJobs,
            CronEvent.Refresh -> loadJobs()
            is CronEvent.CreateJob -> createJob(event)
            is CronEvent.ToggleJob -> toggleJob(event.id, event.enabled)
            is CronEvent.DeleteJob -> deleteJob(event.id)
            is CronEvent.SelectJob -> selectJob(event.id)
            is CronEvent.UpdateCronExpression -> updateCronExpression(event.expression)
            is CronEvent.TestRun -> testRun(event.expression)
            is CronEvent.GenerateCron -> generateCron(event.prompt, event.workflowId)
            CronEvent.ClearSelection -> clearSelection()
            CronEvent.ClearTestResult -> clearTestResult()
        }
    }

    private fun loadJobs() {
        jobsJob?.cancel()
        jobsJob = viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            repository.getJobs()
                .flowOn(dispatcher)
                .catch { throwable ->
                    _uiState.update {
                        it.copy(isLoading = false, error = throwable.message ?: "Failed to load jobs")
                    }
                }
                .collectLatest { jobs ->
                    _uiState.update { state ->
                        state.copy(
                            isLoading = false,
                            jobs = jobs,
                            isEmpty = jobs.isEmpty(),
                            error = null,
                            selectedJob = syncSelection(state.selectedJob, jobs)
                        )
                    }
                }
        }
    }

    private fun createJob(event: CronEvent.CreateJob) {
        viewModelScope.launch {
            _uiState.update { it.copy(isCreating = true) }
            repository.createJob(
                name = event.name,
                description = event.description,
                cronExpression = event.cronExpression,
                workflowId = event.workflowId
            )
                .onSuccess { jobId ->
                    _uiState.update { it.copy(isCreating = false) }
                    _effects.emit(CronEffect.JobCreated(jobId))
                    _effects.emit(CronEffect.ShowSnackbar("Cron job created"))
                }
                .onFailure { throwable ->
                    _uiState.update { it.copy(isCreating = false) }
                    _effects.emit(CronEffect.ShowSnackbar("Failed: ${throwable.message}"))
                }
        }
    }

    private fun toggleJob(id: String, enabled: Boolean) {
        viewModelScope.launch {
            repository.toggleJob(id, enabled)
                .onSuccess {
                    _effects.emit(CronEffect.ShowSnackbar(if (enabled) "Job enabled" else "Job disabled"))
                }
                .onFailure { throwable ->
                    _effects.emit(CronEffect.ShowSnackbar("Toggle failed: ${throwable.message}"))
                }
        }
    }

    private fun deleteJob(id: String) {
        viewModelScope.launch {
            repository.deleteJob(id)
                .onSuccess {
                    _uiState.update { state ->
                        state.copy(
                            selectedJob = state.selectedJob?.takeUnless { it.id == id }
                        )
                    }
                    _effects.emit(CronEffect.JobDeleted(id))
                    _effects.emit(CronEffect.ShowSnackbar("Job deleted"))
                }
                .onFailure { throwable ->
                    _effects.emit(CronEffect.ShowSnackbar("Delete failed: ${throwable.message}"))
                }
        }
    }

    private fun selectJob(id: String) {
        val selected = _uiState.value.jobs.firstOrNull { it.id == id }
        _uiState.update { it.copy(selectedJob = selected, cronExpression = selected?.cronExpression ?: "") }
    }

    private fun updateCronExpression(expression: String) {
        _uiState.update { it.copy(cronExpression = expression) }
    }

    private fun testRun(expression: String) {
        viewModelScope.launch {
            repository.testRun(expression)
                .onSuccess { (isValid, nextRuns) ->
                    _uiState.update {
                        it.copy(testResult = TestRunState(isValid = isValid, nextRuns = nextRuns))
                    }
                    _effects.emit(CronEffect.ShowTestResult(isValid, nextRuns))
                }
                .onFailure { throwable ->
                    _uiState.update {
                        it.copy(testResult = TestRunState(isValid = false, error = throwable.message))
                    }
                }
        }
    }

    private fun generateCron(prompt: String, workflowId: String) {
        viewModelScope.launch {
            repository.generateCron(prompt, workflowId)
                .onSuccess { expression ->
                    _uiState.update {
                        it.copy(generatedExpression = expression, cronExpression = expression)
                    }
                    _effects.emit(CronEffect.ShowSnackbar("Expression generated"))
                }
                .onFailure { throwable ->
                    _effects.emit(CronEffect.ShowSnackbar("AI generation failed: ${throwable.message}"))
                }
        }
    }

    private fun clearSelection() {
        _uiState.update { it.copy(selectedJob = null, cronExpression = "", testResult = null) }
    }

    private fun clearTestResult() {
        _uiState.update { it.copy(testResult = null) }
    }

    private fun syncSelection(
        selectedJob: com.makoclaw.core.model.CronJob?,
        jobs: List<com.makoclaw.core.model.CronJob>
    ): com.makoclaw.core.model.CronJob? = selectedJob?.let { current ->
        jobs.firstOrNull { it.id == current.id }
    }
}
