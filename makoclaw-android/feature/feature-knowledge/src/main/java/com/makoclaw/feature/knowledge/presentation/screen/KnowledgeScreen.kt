package com.makoclaw.feature.knowledge.presentation.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.LibraryBooks
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.material3.pulltorefresh.rememberPullToRefreshState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.model.KnowledgeDocument
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.ErrorScreen
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.knowledge.presentation.component.ChunkEditorSheet
import com.makoclaw.feature.knowledge.presentation.component.DocumentPreviewModal
import com.makoclaw.feature.knowledge.presentation.component.KnowledgeDocumentCard
import com.makoclaw.feature.knowledge.presentation.component.UploadFileButton
import com.makoclaw.feature.knowledge.presentation.state.KnowledgeEffect
import com.makoclaw.feature.knowledge.presentation.state.KnowledgeEvent
import com.makoclaw.feature.knowledge.presentation.state.KnowledgeUiState
import com.makoclaw.feature.knowledge.presentation.viewmodel.KnowledgeViewModel
import kotlinx.coroutines.flow.SharedFlow

@Composable
fun KnowledgeScreen(
    viewModel: KnowledgeViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    KnowledgeScreen(
        uiState = uiState,
        effects = viewModel.effects,
        onRefresh = { viewModel.onEvent(KnowledgeEvent.Refresh) },
        onSearchQueryChange = { viewModel.onEvent(KnowledgeEvent.SearchDocuments(it)) },
        onClearSearch = { viewModel.onEvent(KnowledgeEvent.ClearSearch) },
        onUploadFile = { viewModel.onEvent(KnowledgeEvent.UploadDocument(it)) },
        onSelectDocument = { viewModel.onEvent(KnowledgeEvent.SelectDocument(it)) },
        onDeleteDocument = { viewModel.onEvent(KnowledgeEvent.DeleteDocument(it)) },
        onSaveChunk = { documentId, chunkId, content ->
            viewModel.onEvent(KnowledgeEvent.EditChunk(documentId, chunkId, content))
        }
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun KnowledgeScreen(
    uiState: KnowledgeUiState,
    effects: SharedFlow<KnowledgeEffect>,
    onRefresh: () -> Unit,
    onSearchQueryChange: (String) -> Unit,
    onClearSearch: () -> Unit,
    onUploadFile: (java.io.File) -> Unit,
    onSelectDocument: (String) -> Unit,
    onDeleteDocument: (String) -> Unit,
    onSaveChunk: (String, String, String) -> Unit
) {
    val snackbarHostState = remember { SnackbarHostState() }
    val pullToRefreshState = rememberPullToRefreshState()
    var previewVisible by remember { mutableStateOf(false) }
    var chunkEditorVisible by remember { mutableStateOf(false) }
    var pendingDelete by remember { mutableStateOf<KnowledgeDocument?>(null) }

    LaunchedEffect(effects) {
        effects.collect { effect ->
            when (effect) {
                is KnowledgeEffect.ShowSnackbar -> snackbarHostState.showSnackbar(effect.message)
                is KnowledgeEffect.NavigateToPreview -> previewVisible = true
                is KnowledgeEffect.UploadProgress -> Unit
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Knowledge Base") }
            )
        },
        snackbarHost = { SnackbarHost(hostState = snackbarHostState) },
        floatingActionButton = {
            if (!uiState.isLoading && !uiState.isEmpty) {
                UploadFileButton(
                    onFileSelected = onUploadFile,
                    uploadProgress = uiState.uploadProgress,
                    onValidationError = {
                        previewVisible = false
                    },
                    modifier = Modifier.testTag("knowledge_upload_fab")
                )
            }
        }
    ) { innerPadding ->
        PullToRefreshBox(
            isRefreshing = uiState.isLoading,
            onRefresh = onRefresh,
            state = pullToRefreshState,
            modifier = Modifier
                .fillMaxSize()
                .padding(innerPadding)
        ) {
            KnowledgeContent(
                uiState = uiState,
                onSearchQueryChange = onSearchQueryChange,
                onClearSearch = onClearSearch,
                onUploadFile = onUploadFile,
                onSelectDocument = onSelectDocument,
                onRequestDelete = { pendingDelete = it }
            )
        }
    }

    DocumentPreviewModal(
        document = uiState.selectedDocument,
        visible = previewVisible,
        onDismiss = { previewVisible = false },
        onEditChunks = {
            previewVisible = false
            chunkEditorVisible = true
        }
    )

    uiState.selectedDocument?.let { document ->
        ChunkEditorSheet(
            document = document,
            visible = chunkEditorVisible,
            onDismiss = { chunkEditorVisible = false },
            onSaveChunk = { chunkId, content -> onSaveChunk(document.id, chunkId, content) }
        )
    }

    pendingDelete?.let { document ->
        AlertDialog(
            onDismissRequest = { pendingDelete = null },
            title = { Text("Eliminar documento") },
            text = { Text("¿Querés eliminar ${document.title.ifBlank { document.filename }}?") },
            confirmButton = {
                TextButton(
                    onClick = {
                        onDeleteDocument(document.id)
                        pendingDelete = null
                    }
                ) {
                    Text("Eliminar")
                }
            },
            dismissButton = {
                TextButton(onClick = { pendingDelete = null }) {
                    Text("Cancelar")
                }
            }
        )
    }
}

@Composable
private fun KnowledgeContent(
    uiState: KnowledgeUiState,
    onSearchQueryChange: (String) -> Unit,
    onClearSearch: () -> Unit,
    onUploadFile: (java.io.File) -> Unit,
    onSelectDocument: (String) -> Unit,
    onRequestDelete: (KnowledgeDocument) -> Unit
) {
    when {
        uiState.isLoading && uiState.documents.isEmpty() -> LoadingScreen()
        uiState.error != null -> ErrorScreen(
            message = uiState.error,
            onRetry = onClearSearch
        )
        uiState.isEmpty -> {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(24.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center
            ) {
                EmptyState(
                    title = if (uiState.searchQuery.isBlank()) {
                        "No hay documentos"
                    } else {
                        "No encontramos resultados"
                    },
                    message = if (uiState.searchQuery.isBlank()) {
                        "Subí tu primer documento para empezar a usar la knowledge base."
                    } else {
                        "Probá con otra búsqueda o limpiá el filtro."
                    },
                    icon = Icons.Default.LibraryBooks
                )
                UploadFileButton(
                    onFileSelected = onUploadFile,
                    uploadProgress = uiState.uploadProgress,
                    modifier = Modifier.padding(top = 16.dp)
                )
            }
        }
        else -> {
            LazyColumn(
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                item {
                    OutlinedTextField(
                        value = uiState.searchQuery,
                        onValueChange = onSearchQueryChange,
                        modifier = Modifier
                            .fillMaxWidth()
                            .testTag("knowledge_search_bar"),
                        placeholder = { Text("Buscar documentos o chunks") },
                        leadingIcon = {
                            Icon(Icons.Default.Search, contentDescription = null)
                        },
                        trailingIcon = {
                            if (uiState.searchQuery.isNotBlank()) {
                                IconButton(onClick = onClearSearch) {
                                    Text("Limpiar")
                                }
                            }
                        },
                        singleLine = true
                    )
                }
                item {
                    UploadFileButton(
                        onFileSelected = onUploadFile,
                        uploadProgress = uiState.uploadProgress,
                        modifier = Modifier.fillMaxWidth()
                    )
                }
                items(uiState.documents, key = { it.id }) { document ->
                    KnowledgeDocumentCard(
                        document = document,
                        onClick = { onSelectDocument(document.id) },
                        onLongClick = { onRequestDelete(document) },
                        isSelected = uiState.selectedDocument?.id == document.id
                    )
                }
            }
        }
    }
}
