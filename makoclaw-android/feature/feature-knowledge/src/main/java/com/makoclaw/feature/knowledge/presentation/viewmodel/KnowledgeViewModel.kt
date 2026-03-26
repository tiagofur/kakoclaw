package com.makoclaw.feature.knowledge.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.KnowledgeDocument
import com.makoclaw.core.network.api.KnowledgeApi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class KnowledgeUiState(
    val documents: List<KnowledgeDocument> = emptyList(),
    val isLoading: Boolean = true,
    val searchQuery: String = "",
    val error: String? = null
)

@HiltViewModel
class KnowledgeViewModel @Inject constructor(
    private val knowledgeApi: KnowledgeApi
) : ViewModel() {

    private val _uiState = MutableStateFlow(KnowledgeUiState())
    val uiState: StateFlow<KnowledgeUiState> = _uiState.asStateFlow()

    init { loadDocuments() }

    fun loadDocuments() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            try {
                val response = knowledgeApi.listDocuments()
                _uiState.update { it.copy(documents = response.documents, isLoading = false) }
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun search(query: String) {
        _uiState.update { it.copy(searchQuery = query) }
        if (query.isBlank()) {
            loadDocuments()
            return
        }
        viewModelScope.launch {
            try {
                knowledgeApi.search(query)
            } catch (_: Exception) { }
        }
    }

    fun deleteDocument(id: Long) {
        viewModelScope.launch {
            try {
                knowledgeApi.deleteDocument(id)
                loadDocuments()
            } catch (_: Exception) { }
        }
    }
}
