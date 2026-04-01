package com.makoclaw.feature.memory.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.feature.memory.data.repository.MemoryRepository
import com.makoclaw.feature.memory.presentation.state.MemoryEffect
import com.makoclaw.feature.memory.presentation.state.MemoryEvent
import com.makoclaw.feature.memory.presentation.state.MemoryUiState
import com.makoclaw.feature.memory.presentation.state.SearchResult
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

@HiltViewModel
class MemoryViewModel @Inject constructor(
    private val repository: MemoryRepository,
    private val dispatcher: CoroutineDispatcher = Dispatchers.IO
) : ViewModel() {

    private val _uiState = MutableStateFlow(MemoryUiState())
    private val _effects = MutableSharedFlow<MemoryEffect>()
    private var memoryJob: Job? = null
    private var originalMemory: String = ""

    val uiState: StateFlow<MemoryUiState> = _uiState.asStateFlow()
    val effects: SharedFlow<MemoryEffect> = _effects.asSharedFlow()

    init {
        loadMemory()
        loadDailyNotes()
    }

    fun onEvent(event: MemoryEvent) {
        when (event) {
            is MemoryEvent.LoadMemory -> loadMemory()
            is MemoryEvent.LoadDailyNotes -> loadDailyNotes()
            is MemoryEvent.Refresh -> { loadMemory(); loadDailyNotes() }
            is MemoryEvent.UpdateMemoryText -> updateMemoryText(event.text)
            is MemoryEvent.SaveMemory -> saveMemory()
            is MemoryEvent.CreateDailyNote -> createDailyNote(event.content)
            is MemoryEvent.SearchMemory -> searchMemory(event.query)
            is MemoryEvent.SelectTab -> selectTab(event.tab)
            is MemoryEvent.ClearError -> clearError()
        }
    }

    private fun loadMemory() {
        memoryJob?.cancel()
        memoryJob = viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            repository.getLongTermMemory()
                .flowOn(dispatcher)
                .catch { throwable ->
                    _uiState.update {
                        it.copy(isLoading = false, error = throwable.message ?: "Failed to load memory")
                    }
                }
                .collectLatest { content ->
                    originalMemory = content
                    _uiState.update { state ->
                        state.copy(
                            isLoading = false,
                            longTermMemory = content,
                            hasChanges = false,
                            error = null
                        )
                    }
                }
        }
    }

    private fun loadDailyNotes() {
        viewModelScope.launch {
            repository.getDailyNotes(7)
                .onSuccess { notes ->
                    _uiState.update { it.copy(dailyNotes = notes) }
                }
                .onFailure { throwable ->
                    _effects.emit(MemoryEffect.ShowSnackbar("Failed to load notes: " + throwable.message))
                }
        }
    }

    private fun updateMemoryText(text: String) {
        _uiState.update {
            it.copy(longTermMemory = text, hasChanges = text != originalMemory)
        }
    }

    private fun saveMemory() {
        val content = _uiState.value.longTermMemory
        viewModelScope.launch {
            _uiState.update { it.copy(isSaving = true) }
            repository.saveLongTermMemory(content)
                .onSuccess {
                    originalMemory = content
                    _uiState.update { it.copy(isSaving = false, hasChanges = false) }
                    _effects.emit(MemoryEffect.MemorySaved)
                    _effects.emit(MemoryEffect.ShowSnackbar("Memory saved"))
                }
                .onFailure { throwable ->
                    _uiState.update { it.copy(isSaving = false) }
                    _effects.emit(MemoryEffect.ShowSnackbar("Save failed: ${throwable.message}"))
                }
        }
    }

    private fun createDailyNote(content: String) {
        viewModelScope.launch {
            repository.createDailyNote(content)
                .onSuccess { note ->
                    _effects.emit(MemoryEffect.DailyNoteCreated(note.date))
                    _effects.emit(MemoryEffect.ShowSnackbar("Note created"))
                    loadDailyNotes()
                }
                .onFailure { throwable ->
                    _effects.emit(MemoryEffect.ShowSnackbar("Failed to create note: " + throwable.message))
                }
        }
    }

    private fun searchMemory(query: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(searchQuery = query) }
            if (query.isBlank()) {
                _uiState.update { it.copy(searchResults = emptyList()) }
                return@Suppress
            }
            repository.searchMemory(query)
                .onSuccess { results ->
                    _uiState.update {
                        it.copy(searchResults = results.map { r ->
                            SearchResult(
                                id = r.id,
                                content = r.content,
                                type = r.type,
                                date = r.date,
                                highlightedContent = r.highlightedContent
                            )
                        })
                    }
                }
                .onFailure { throwable ->
                    _effects.emit(MemoryEffect.ShowSnackbar("Search failed: " + throwable.message))
                }
        }
    }

    private fun selectTab(tab: Int) {
        _uiState.update { it.copy(selectedTab = tab) }
    }

    private fun clearError() {
        _uiState.update { it.copy(error = null) }
    }
}
