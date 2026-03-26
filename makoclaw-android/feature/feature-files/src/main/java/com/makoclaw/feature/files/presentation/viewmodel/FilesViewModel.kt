package com.makoclaw.feature.files.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.FileEntry
import com.makoclaw.core.network.api.FilesApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class FilesUiState(
    val currentPath: String = "",
    val entries: List<FileEntry> = emptyList(),
    val pathHistory: List<String> = emptyList(),
    val isLoading: Boolean = true,
    val error: String? = null
)

@HiltViewModel
class FilesViewModel @Inject constructor(
    private val filesApi: FilesApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(FilesUiState())
    val uiState: StateFlow<FilesUiState> = _uiState.asStateFlow()

    init { loadRoot() }

    fun loadRoot() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val response = filesApi.listRoot()
                _uiState.update {
                    it.copy(
                        currentPath = response.path,
                        entries = response.entries,
                        pathHistory = listOf(response.path),
                        isLoading = false
                    )
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun navigateTo(path: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val response = filesApi.browse(path)
                _uiState.update {
                    it.copy(
                        currentPath = path,
                        entries = response.entries,
                        pathHistory = it.pathHistory + path,
                        isLoading = false
                    )
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun navigateBack(): Boolean {
        val history = _uiState.value.pathHistory
        if (history.size <= 1) return false
        val previousPath = history[history.size - 2]
        _uiState.update { it.copy(pathHistory = history.dropLast(1)) }
        navigateTo(previousPath)
        return true
    }

    fun deleteFile(path: String) {
        viewModelScope.launch {
            try {
                filesApi.deleteFile(path)
                navigateTo(_uiState.value.currentPath)
            } catch (_: Exception) { }
        }
    }
}
