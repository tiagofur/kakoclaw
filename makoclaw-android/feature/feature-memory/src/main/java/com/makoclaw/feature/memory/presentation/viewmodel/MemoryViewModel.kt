package com.makoclaw.feature.memory.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.DailyNote
import com.makoclaw.core.model.MemoryContent
import com.makoclaw.core.network.api.MemoryApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class MemoryUiState(
    val longTermMemory: String = "",
    val dailyNotes: List<DailyNote> = emptyList(),
    val isLoading: Boolean = true,
    val isSaving: Boolean = false,
    val hasChanges: Boolean = false,
    val error: String? = null
)

@HiltViewModel
class MemoryViewModel @Inject constructor(
    private val memoryApi: MemoryApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(MemoryUiState())
    val uiState: StateFlow<MemoryUiState> = _uiState.asStateFlow()

    private var originalMemory = ""

    init {
        loadMemory()
        loadDailyNotes()
    }

    fun loadMemory() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val response = memoryApi.getLongTermMemory()
                originalMemory = response.content
                _uiState.update {
                    it.copy(longTermMemory = response.content, isLoading = false)
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun loadDailyNotes() {
        viewModelScope.launch {
            try {
                val notes = memoryApi.getDailyNotes(7)
                _uiState.update { it.copy(dailyNotes = notes) }
            } catch (_: Exception) { }
        }
    }

    fun updateMemoryText(text: String) {
        _uiState.update {
            it.copy(longTermMemory = text, hasChanges = text != originalMemory)
        }
    }

    fun saveMemory() {
        viewModelScope.launch {
            _uiState.update { it.copy(isSaving = true) }
            try {
                memoryApi.updateLongTermMemory(MemoryContent(_uiState.value.longTermMemory))
                originalMemory = _uiState.value.longTermMemory
                _uiState.update { it.copy(isSaving = false, hasChanges = false) }
            } catch (e: Exception) {
                _uiState.update { it.copy(isSaving = false, error = e.message) }
            }
        }
    }
}
