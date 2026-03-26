package com.makoclaw.feature.knowledge.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.KnowledgeDocument
import com.makoclaw.feature.knowledge.data.di.IoDispatcher
import com.makoclaw.feature.knowledge.data.repository.KnowledgeRepository
import com.makoclaw.feature.knowledge.presentation.state.KnowledgeEffect
import com.makoclaw.feature.knowledge.presentation.state.KnowledgeEvent
import com.makoclaw.feature.knowledge.presentation.state.KnowledgeUiState
import dagger.hilt.android.lifecycle.HiltViewModel
import java.io.File
import javax.inject.Inject
import kotlinx.coroutines.CoroutineDispatcher
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
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

/** Coordinates state, search, upload and editing for knowledge documents. */
@HiltViewModel
class KnowledgeViewModel @Inject constructor(
    private val repository: KnowledgeRepository,
    @IoDispatcher private val dispatcher: CoroutineDispatcher
) : ViewModel() {

    private val _uiState = MutableStateFlow(KnowledgeUiState())
    private val _effects = MutableSharedFlow<KnowledgeEffect>()

    private var documentsJob: Job? = null
    private var searchJob: Job? = null

    val uiState: StateFlow<KnowledgeUiState> = _uiState.asStateFlow()
    val effects: SharedFlow<KnowledgeEffect> = _effects.asSharedFlow()

    init {
        loadDocuments()
    }

    fun onEvent(event: KnowledgeEvent) {
        when (event) {
            KnowledgeEvent.LoadDocuments,
            KnowledgeEvent.Refresh -> loadDocuments()
            is KnowledgeEvent.UploadDocument -> uploadDocument(event.file)
            is KnowledgeEvent.SearchDocuments -> searchDocuments(event.query)
            is KnowledgeEvent.DeleteDocument -> deleteDocument(event.id)
            is KnowledgeEvent.EditChunk -> editChunk(event.documentId, event.chunkId, event.content)
            is KnowledgeEvent.SelectDocument -> selectDocument(event.id)
            KnowledgeEvent.ClearSearch -> clearSearch()
        }
    }

    private fun loadDocuments() {
        searchJob?.cancel()
        documentsJob?.cancel()
        documentsJob = viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            repository.getDocuments()
                .flowOn(dispatcher)
                .catch { throwable ->
                    handleError(throwable.message ?: "Failed to load documents")
                }
                .collectLatest { documents ->
                    val sortedDocuments = documents.sortedByDescending(::createdAtMillis)
                    _uiState.update { state ->
                        state.copy(
                            isLoading = false,
                            documents = sortedDocuments,
                            error = null,
                            isEmpty = sortedDocuments.isEmpty(),
                            selectedDocument = syncSelection(state.selectedDocument, sortedDocuments)
                        )
                    }
                }
        }
    }

    private fun uploadDocument(file: File) {
        viewModelScope.launch {
            repository.uploadDocument(file)
                .flowOn(dispatcher)
                .catch { throwable ->
                    handleError(throwable.message ?: "Upload failed")
                }
                .collectLatest { progress ->
                    _uiState.update {
                        it.copy(uploadProgress = progress.progress, error = null)
                    }
                    _effects.emit(KnowledgeEffect.UploadProgress(progress.progress))

                    if (progress.progress >= 1f) {
                        _uiState.update { state -> state.copy(uploadProgress = null) }
                        _effects.emit(KnowledgeEffect.ShowSnackbar("Documento subido correctamente"))
                    }
                }
        }
    }

    private fun searchDocuments(query: String) {
        searchJob?.cancel()
        _uiState.update {
            it.copy(searchQuery = query, error = null, isLoading = query.isNotBlank())
        }

        if (query.isBlank()) {
            loadDocuments()
            return
        }

        documentsJob?.cancel()
        searchJob = viewModelScope.launch {
            delay(300)
            repository.searchDocuments(query)
                .flowOn(dispatcher)
                .catch { throwable ->
                    handleError(throwable.message ?: "Search failed")
                }
                .collectLatest { documents ->
                    _uiState.update { state ->
                        state.copy(
                            isLoading = false,
                            documents = documents,
                            isEmpty = documents.isEmpty(),
                            selectedDocument = syncSelection(state.selectedDocument, documents)
                        )
                    }
                }
        }
    }

    private fun deleteDocument(id: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            repository.deleteDocument(id)
                .flowOn(dispatcher)
                .catch { throwable ->
                    handleError(throwable.message ?: "Delete failed")
                }
                .collectLatest {
                    _uiState.update {
                        it.copy(
                            isLoading = false,
                            selectedDocument = it.selectedDocument?.takeUnless { document -> document.id == id }
                        )
                    }
                    _effects.emit(KnowledgeEffect.ShowSnackbar("Documento eliminado"))
                }
        }
    }

    private fun editChunk(documentId: String, chunkId: String, content: String) {
        viewModelScope.launch {
            repository.editChunk(documentId, chunkId, content)
                .flowOn(dispatcher)
                .catch { throwable ->
                    handleError(throwable.message ?: "Chunk update failed")
                }
                .collectLatest {
                    _effects.emit(KnowledgeEffect.ShowSnackbar("Chunk guardado"))
                }
        }
    }

    private fun selectDocument(id: String) {
        val selected = _uiState.value.documents.firstOrNull { it.id == id }
        _uiState.update { it.copy(selectedDocument = selected) }

        if (selected != null) {
            viewModelScope.launch {
                _effects.emit(KnowledgeEffect.NavigateToPreview(selected.id))
            }
        }
    }

    private fun clearSearch() {
        _uiState.update { it.copy(searchQuery = "") }
        loadDocuments()
    }

    private suspend fun handleError(message: String) {
        _uiState.update { it.copy(isLoading = false, uploadProgress = null, error = message) }
        _effects.emit(KnowledgeEffect.ShowSnackbar(message))
    }

    private fun syncSelection(
        selectedDocument: KnowledgeDocument?,
        documents: List<KnowledgeDocument>
    ): KnowledgeDocument? = selectedDocument?.let { current ->
        documents.firstOrNull { it.id == current.id }
    }

    private fun createdAtMillis(document: KnowledgeDocument): Long =
        document.createdAt.toLongOrNull() ?: 0L
}
