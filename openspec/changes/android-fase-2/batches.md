# Implementación BATCH 2: Feature Knowledge (Código Completo)

## Instrucciones

Este documento contiene TODO el código fuente para las 6 tareas del BATCH 2. Cada sección incluye:

1. **Ubicación del archivo** - Dónde guardar el archivo
2. **Código completo** - Listo para copiar/pegar
3. **Instrucciones** - Cómo usarlo

---

## TASK-KNOWLEDGE-001: Crear KnowledgeViewModel

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/state/KnowledgeUiState.kt`
**Paquete**: `com.makoclaw.feature.knowledge.presentation.state`

### Código Completo

```kotlin
package com.makoclaw.feature.knowledge.presentation.state

import com.makoclaw.core.model.KnowledgeDocument

data class KnowledgeUiState(
    val isLoading: Boolean = false,
    val documents: List<KnowledgeDocument> = emptyList(),
    val error: String? = null,
    val isEmpty: Boolean = false,
    val searchQuery: String = "",
    val selectedDocument: KnowledgeDocument? = null,
    val uploadProgress: Float? = null
)
```

**Archivo**: `makoclaw-android/feature/feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/state/KnowledgeEvent.kt`
**Paquete**: `com.makoclaw.feature.knowledge.presentation.state`

```kotlin
package com.makoclaw.feature.knowledge.presentation.state

sealed class KnowledgeEvent {
    data object LoadDocuments : KnowledgeEvent()
    data object Refresh : KnowledgeEvent()
    data class UploadDocument(val file: java.io.File) : KnowledgeEvent()
    data class SearchDocuments(val query: String) : KnowledgeEvent()
    data class DeleteDocument(val id: String) : KnowledgeEvent()
    data class EditChunk(val documentId: String, val chunkId: String, val content: String) : KnowledgeEvent()
    data class SelectDocument(val id: String) : KnowledgeEvent()
}
```

**Archivo**: `makoclaw-android/feature/feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/state/KnowledgeEffect.kt`
**Paquete**: `com.makoclaw.feature.knowledge.presentation.state`

```kotlin
package com.makoclaw.feature.knowledge.presentation.state

sealed class KnowledgeEffect {
    data class ShowSnackbar(val message: String) : KnowledgeEffect()
    data class NavigateToPreview(val id: String) : KnowledgeEffect()
    data class UploadProgress(val progress: Float) : KnowledgeEffect()
}
```

**Archivo**: `makoclaw-android/feature/feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/viewmodel/KnowledgeViewModel.kt`
**Paquete**: `com.makoclaw.feature.knowledge.presentation.viewmodel`

```kotlin
package com.makoclaw.feature.knowledge.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.KnowledgeDocument
import com.makoclaw.feature.knowledge.data.repository.KnowledgeRepository
import com.makoclaw.feature.knowledge.presentation.state.KnowledgeEffect
import com.makoclaw.feature.knowledge.presentation.state.KnowledgeEvent
import com.makoclaw.feature.knowledge.presentation.state.KnowledgeUiState
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class KnowledgeViewModel @Inject constructor(
    private val repository: KnowledgeRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(KnowledgeUiState())
    val uiState: StateFlow<KnowledgeUiState> = _uiState.asStateFlow()

    private val _effects = MutableSharedFlow<KnowledgeEffect>()
    val effects: SharedFlow<KnowledgeEffect> = _effects.asSharedFlow()

    init {
        loadDocuments()
    }

    fun onEvent(event: KnowledgeEvent) {
        when (event) {
            is KnowledgeEvent.LoadDocuments -> loadDocuments()
            is KnowledgeEvent.Refresh -> loadDocuments()
            is KnowledgeEvent.UploadDocument -> uploadDocument(event.file)
            is KnowledgeEvent.SearchDocuments -> searchDocuments(event.query)
            is KnowledgeEvent.DeleteDocument -> deleteDocument(event.id)
            is KnowledgeEvent.EditChunk -> editChunk(event.documentId, event.chunkId, event.content)
            is KnowledgeEvent.SelectDocument -> selectDocument(event.id)
        }
    }

    private fun loadDocuments() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            try {
                repository.getDocuments()
                    .collect { documents ->
                        _uiState.update {
                            it.copy(
                                isLoading = false,
                                documents = documents,
                                isEmpty = documents.isEmpty()
                            )
                        }
                    }
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        error = "Failed to load documents: ${e.message}"
                    )
                }
                _effects.emit(KnowledgeEffect.ShowSnackbar("Error loading documents"))
            }
        }
    }

    private fun uploadDocument(file: java.io.File) {
        viewModelScope.launch {
            try {
                repository.uploadDocument(file)
                    .collect { progress ->
                        _effects.emit(KnowledgeEffect.UploadProgress(progress))

                        if (progress >= 1.0f) {
                            loadDocuments()
                            _effects.emit(KnowledgeEffect.ShowSnackbar("Document uploaded successfully"))
                        }
                    }
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        uploadProgress = null,
                        error = "Upload failed: ${e.message}"
                    )
                }
                _effects.emit(KnowledgeEffect.ShowSnackbar("Error uploading document"))
            }
        }
    }

    private fun searchDocuments(query: String) {
        viewModelScope.launch {
            try {
                repository.searchDocuments(query)
                    .collect { documents ->
                        _uiState.update {
                            it.copy(
                                documents = documents,
                                isEmpty = documents.isEmpty(),
                                searchQuery = query
                            )
                        }
                    }
            } catch (e: Exception) {
                _effects.emit(KnowledgeEffect.ShowSnackbar("Search failed: ${e.message}"))
            }
        }
    }

    private fun deleteDocument(id: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }

            try {
                repository.deleteDocument(id)
                _effects.emit(KnowledgeEffect.ShowSnackbar("Document deleted"))
                loadDocuments()
            } catch (e: Exception) {
                _uiState.update { it.copy(isLoading = false, error = e.message) }
                _effects.emit(KnowledgeEffect.ShowSnackbar("Error deleting document"))
            }
        }
    }

    private fun editChunk(documentId: String, chunkId: String, content: String) {
        viewModelScope.launch {
            try {
                repository.editChunk(documentId, chunkId, content)
                _effects.emit(KnowledgeEffect.ShowSnackbar("Chunk saved"))
            } catch (e: Exception) {
                _effects.emit(KnowledgeEffect.ShowSnackbar("Error saving chunk"))
            }
        }
    }

    private fun selectDocument(id: String) {
        _uiState.update { it.copy(selectedDocument = it.documents.find { it.id == id }) }
    }
}
```

### Instrucciones

1. Crear el directorio `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/state/` si no existe
2. Crear el archivo `KnowledgeUiState.kt` y pegar el primer bloque de código
3. Crear el archivo `KnowledgeEvent.kt` y pegar el segundo bloque
4. Crear el archivo `KnowledgeEffect.kt` y pegar el tercer bloque
5. Crear el directorio `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/viewmodel/` si no existe
6. Crear el archivo `KnowledgeViewModel.kt` y pegar el último bloque
7. Verificar que no hay errores de compilación

---

## TASK-KNOWLEDGE-002: Crear KnowledgeRepository

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/data/repository/KnowledgeRepository.kt`
**Paquete**: `com.makoclaw.feature.knowledge.data.repository`

### Código Completo

**Archivo**: `makoclaw-android/feature/feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/data/repository/KnowledgeRepository.kt`

```kotlin
package com.makoclaw.feature.knowledge.data.repository

import com.makoclaw.core.model.KnowledgeDocument
import com.makoclaw.core.network.api.KnowledgeApi
import com.makoclaw.core.network.api.UploadProgress
import com.makoclaw.core.network.api.UploadRequest
import com.makoclaw.core.network.api.SearchRequest
import com.makoclaw.core.network.api.SearchResponse
import com.makoclaw.core.network.api.DocumentsResponse
import com.makoclaw.core.network.api.ChunkUpdateRequest
import com.makoclaw.core.network.api.DeleteRequest
import com.makoclaw.core.network.api.ApiResponse
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.catch
import javax.inject.Inject

interface KnowledgeRepository {
    fun getDocuments(): Flow<List<KnowledgeDocument>>
    fun uploadDocument(file: java.io.File): Flow<UploadProgress>
    fun searchDocuments(query: String): Flow<List<KnowledgeDocument>>
    fun deleteDocument(id: String): Flow<ApiResponse>
    fun editChunk(documentId: String, chunkId: String, content: String): Flow<ApiResponse>
}

@ViewModelScoped
class KnowledgeRepositoryImpl @Inject constructor(
    private val api: KnowledgeApi,
    private val dao: KnowledgeDocumentDao
) : KnowledgeRepository {

    override fun getDocuments(): Flow<List<KnowledgeDocument>> {
        return dao.getAll()
            .map { entities -> entities.map { it.toModel() } }
            .catch { e ->
                // Fallback to API if local cache fails
                api.getDocuments()
                    .map { response -> response.documents }
            }
    }

    override fun uploadDocument(file: java.io.File): Flow<UploadProgress> {
        return api.uploadDocument(
            UploadRequest(
                fileName = file.name,
                fileSize = file.length(),
                contentType = "application/octet-stream"
            )
        )
    }

    override fun searchDocuments(query: String): Flow<List<KnowledgeDocument>> {
        return api.searchDocuments(
            SearchRequest(query = query)
        ).map { response -> response.results }
    }

    override fun deleteDocument(id: String): Flow<ApiResponse> {
        return api.deleteDocument(
            DeleteRequest(documentId = id)
        )
    }

    override fun editChunk(documentId: String, chunkId: String, content: String): Flow<ApiResponse> {
        return api.updateChunk(
            ChunkUpdateRequest(
                documentId = documentId,
                chunkId = chunkId,
                content = content
            )
        )
    }
}
```

**Archivo**: `makoclaw-android/core/core-network/src/main/java/com/makoclaw/core/network/api/KnowledgeApi.kt` (crear si no existe)

```kotlin
package com.makoclaw.core.network.api

import com.makoclaw.core.network.api.UploadProgress
import okhttp3.MultipartBody
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Query

data class DocumentsResponse(
    val documents: List<KnowledgeDocument> = emptyList()
)

data class SearchResponse(
    val results: List<KnowledgeDocument> = emptyList()
)

data class UploadProgress(
    val progress: Float,
    val uploadedBytes: Long,
    val totalBytes: Long
)

data class UploadRequest(
    val fileName: String,
    val fileSize: Long,
    val contentType: String
)

data class SearchRequest(
    val query: String
)

data class DeleteRequest(
    val documentId: String
)

data class ChunkUpdateRequest(
    val documentId: String,
    val chunkId: String,
    val content: String
)

interface KnowledgeApi {
    @GET("knowledge/documents")
    suspend fun getDocuments(): DocumentsResponse

    @Multipart
    @POST("knowledge/upload")
    suspend fun uploadDocument(@Part file: MultipartBody.Part): UploadProgress

    @POST("knowledge/search")
    suspend fun searchDocuments(@Body request: SearchRequest): SearchResponse

    @DELETE("knowledge/document/{id}")
    suspend fun deleteDocument(@Path("id") id: String): ApiResponse

    @POST("knowledge/document/{id}/chunk")
    suspend fun updateChunk(@Path("id") id: String, @Body request: ChunkUpdateRequest): ApiResponse
}
```

### Instrucciones

1. Crear el directorio `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/data/repository/` si no existe
2. Crear el archivo `KnowledgeRepository.kt` (interface) y pegar el primer bloque de código
3. Crear el archivo `KnowledgeRepositoryImpl.kt` y pegar el segundo bloque de código
4. Si `KnowledgeApi.kt` no existe en `core-network/src/main/java/com/makoclaw/core/network/api/`, crearlo
5. Verificar que no hay errores de compilación

---

## TASK-KNOWLEDGE-003: Crear KnowledgeScreen UI

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/screen/KnowledgeScreen.kt`
**Paquete**: `com.makoclaw.feature.knowledge.presentation.screen`

### Código Completo

```kotlin
package com.makoclaw.feature.knowledge.presentation.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.LibraryBooks
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Upload
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarDuration
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.material3.pulltorefresh.rememberPullToRefreshState
import androidx.compose.material3.Button
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.knowledge.presentation.viewmodel.KnowledgeViewModel
import com.makoclaw.feature.knowledge.presentation.component.UploadFileButton
import com.makoclaw.feature.knowledge.presentation.component.KnowledgeDocumentCard
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun KnowledgeScreen(
    viewModel: KnowledgeViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val effects by viewModel.effects.collectAsState(initialValue = null)
    val scope = rememberCoroutineScope()

    // Snackbar host
    val snackbarHostState = SnackbarHostState()

    // Pull to refresh
    val pullRefreshState = rememberPullToRefreshState(
        onRefresh = { viewModel.onEvent(KnowledgeEvent.Refresh) }
    )

    // Handle effects
    LaunchedEffect(Unit) {
        effects?.let { effect ->
            when (effect) {
                is KnowledgeEffect.ShowSnackbar -> {
                    scope.launch {
                        snackbarHostState.showSnackbar(
                            message = effect.message,
                            duration = SnackbarDuration.Short
                        )
                    }
                }
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Knowledge Base") },
                actions = {
                    Button(onClick = { viewModel.onEvent(KnowledgeEvent.SearchDocuments(uiState.searchQuery)) }) {
                        Icon(Icons.Filled.Search, contentDescription = "Search")
                    }
                }
            )
        }
    ) { padding ->
        PullToRefreshBox(
            state = pullRefreshState,
            isRefreshing = uiState.isLoading,
            onRefresh = { viewModel.onEvent(KnowledgeEvent.Refresh) },
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            if (uiState.isLoading) {
                LoadingScreen()
            } else if (uiState.error != null) {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    horizontalAlignment = androidx.compose.foundation.layout.Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    Icon(Icons.Filled.LibraryBooks, modifier = Modifier.size(64.dp))
                    Spacer(modifier = Modifier.height(16.dp))
                    Text(
                        text = "Error: ${uiState.error}",
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.error
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    Button(onClick = { viewModel.onEvent(KnowledgeEvent.Refresh) }) {
                        Text("Retry")
                    }
                }
            } else if (uiState.isEmpty) {
                EmptyState(
                    icon = Icons.Default.LibraryBooks,
                    title = "No documents yet",
                    message = "Upload your first document to build your knowledge base",
                    actionLabel = "Upload Document"
                ) {
                    UploadFileButton(onFileSelected = { file ->
                        viewModel.onEvent(KnowledgeEvent.UploadDocument(file))
                    })
                }
            } else {
                LazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                    horizontalAlignment = androidx.compose.foundation.layout.Alignment.CenterHorizontally
                ) {
                    item { Spacer(modifier = Modifier.height(8.dp)) }

                    items(uiState.documents) { document ->
                        KnowledgeDocumentCard(
                            document = document,
                            onClick = { viewModel.onEvent(KnowledgeEvent.SelectDocument(document.id)) },
                            onLongClick = { _, offset ->
                                viewModel.onEvent(KnowledgeEvent.DeleteDocument(document.id))
                            }
                        )
                    }

                    item { Spacer(modifier = Modifier.height(80.dp)) }
                }
            }
        }
    }
}
```

### Instrucciones

1. Crear el directorio `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/screen/` si no existe
2. Crear el archivo `KnowledgeScreen.kt` y pegar todo el código
3. Verificar que no hay errores de compilación

---

## TASK-KNOWLEDGE-004: Crear KnowledgeDocumentCard

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/component/KnowledgeDocumentCard.kt`
**Paquete**: `com.makoclaw.feature.knowledge.presentation.component`

### Código Completo

```kotlin
package com.makoclaw.feature.knowledge.presentation.component

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Description
import androidx.compose.material.icons.filled.Article
import androidx.compose.material.icons.filled.PictureAsPdf
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import com.makoclaw.core.model.KnowledgeDocument
import kotlinx.datetime.Clock

@Composable
fun KnowledgeDocumentCard(
    document: KnowledgeDocument,
    onClick: () -> Unit,
    onLongClick: ((String, androidx.compose.ui.geometry.Offset) -> Unit)? = null
) {
    val isSelected = document.id == selectedDocumentId

    Card(
        onClick = onClick,
        modifier = Modifier
            .fillMaxWidth()
            .then(
                if (onLongClick != null) {
                    Modifier.combined(
                        Modifier.pointerInput(Unit) {
                            onLongClick(onLongClick)
                        }
                    )
                } else {
                    Modifier
                }
            )
            .background(
                color = if (isSelected)
                    MaterialTheme.colorScheme.primaryContainer
                else
                    MaterialTheme.colorScheme.surface
            )
            .then(
                Modifier.border(
                    width = if (isSelected) 2.dp else 1.dp,
                    color = if (isSelected)
                        MaterialTheme.colorScheme.primary
                    else
                        MaterialTheme.colorScheme.outline
                )
            ),
        colors = CardDefaults.cardColors(
            containerColor = if (isSelected)
                MaterialTheme.colorScheme.primaryContainer
            else
                MaterialTheme.colorScheme.surface
        ),
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            // Header: Icon + Title
            Row(
                horizontalArrangement = Arrangement.spacedBy(12.dp),
                verticalAlignment = androidx.compose.foundation.layout.Alignment.CenterVertically
            ) {
                // Icon based on type (simplified - would check content type from metadata)
                Icon(
                    imageVector = when {
                        document.title.endsWith(".pdf") -> Icons.Filled.PictureAsPdf
                        else -> Icons.Filled.Article
                    },
                    contentDescription = document.title,
                    modifier = Modifier.size(24.dp),
                    tint = if (isSelected)
                        MaterialTheme.colorScheme.onPrimaryContainer
                    else
                        MaterialTheme.colorScheme.onSurface
                )
                
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = document.title,
                        style = MaterialTheme.typography.titleMedium,
                        maxLines = 2,
                        overflow = androidx.compose.ui.text.TextOverflow.Ellipsis,
                        color = if (isSelected)
                            MaterialTheme.colorScheme.onPrimaryContainer
                        else
                            MaterialTheme.colorScheme.onSurface
                    )
                    
                    Spacer(modifier = Modifier.height(4.dp))
                    
                    Text(
                        text = "${document.chunks.size} chunks",
                        style = MaterialTheme.typography.bodySmall,
                        color = if (isSelected)
                            MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.7f)
                        else
                            MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                
                Spacer(modifier = Modifier.weight(1f))
                
                Text(
                    text = formatDate(document.createdAt),
                    style = MaterialTheme.typography.bodySmall,
                    color = if (isSelected)
                        MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.7f)
                        else
                        MaterialTheme.colorScheme.onSurfaceVariant
                    )
                )
            }
        }
    }
}

@Composable
private fun formatDate(timestamp: Long): String {
    val instant = Clock.System.now().fromEpochMilliseconds(timestamp)
    val formatter = java.text.SimpleDateFormat("MMM dd, yyyy", java.util.Locale.US)
    return formatter.format(instant.toJavaDate())
}
```

### Instrucciones

1. Crear el directorio `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/component/` si no existe
2. Crear el archivo `KnowledgeDocumentCard.kt` y pegar todo el código
3. Verificar que no hay errores de compilación

---

## TASK-KNOWLEDGE-005: Crear UploadFileButton y DocumentPreviewModal

### Ubicación

**Archivo 1**: `makoclaw-android/feature/feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/component/UploadFileButton.kt`
**Archivo 2**: `makoclaw-android/feature/feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/component/DocumentPreviewModal.kt`
**Paquete**: `com.makoclaw.feature.knowledge.presentation.component`

### Código Completo - UploadFileButton.kt

```kotlin
package com.makoclaw.feature.knowledge.presentation.component

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CloudUpload
import androidx.compose.material3.ExtendedFloatingActionButton
import androidx.compose.material3.FloatingActionButtonDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import java.io.File
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

@Composable
fun UploadFileButton(
    onFileSelected: (File) -> Unit,
    uploadProgress: Float? = null
) {
    val pickFile = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetContent()
    )

    // Accept PDF, TXT, and MD files
    val pickFileLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetContent(
            arrayOf("application/pdf", "text/plain", "text/markdown")
        )
    )

    Row(
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        verticalAlignment = androidx.compose.foundation.layout.Alignment.CenterVertically
    ) {
        if (uploadProgress != null && uploadProgress < 1.0f) {
            // Show progress
            Column {
                LinearProgressIndicator(
                    progress = { uploadProgress },
                    modifier = Modifier
                        .size(width = 200.dp, height = 4.dp)
                        .clip(RoundedCornerShape(2.dp)),
                    color = MaterialTheme.colorScheme.primary,
                    trackColor = MaterialTheme.colorScheme.surfaceVariant,
                )
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = "${(uploadProgress * 100).toInt()}%",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        } else {
            // Show upload button
            ExtendedFloatingActionButton(
                onClick = {
                    pickFileLauncher.launch("application/pdf")
                },
                icon = { Icon(Icons.Filled.CloudUpload, "Upload") },
                text = { Text("Upload") },
                modifier = Modifier.size(height = 56.dp),
                containerColor = MaterialTheme.colorScheme.primaryContainer,
                contentColor = MaterialTheme.colorScheme.onPrimaryContainer,
                elevation = FloatingActionButtonDefaults.elevation(6.dp),
                expanded = true
            )
        }
    }
}

@Composable
fun UploadFileButtonSmall(
    onFileSelected: (File) -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true
) {
    val pickFile = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetContent()
    )

    // Accept PDF, TXT, and MD files
    val pickFileLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetContent(
            arrayOf("application/pdf", "text/plain", "text/markdown")
        )
    )

    ExtendedFloatingActionButton(
        onClick = {
            pickFileLauncher.launch("application/pdf")
        },
        icon = { Icon(Icons.Filled.CloudUpload, "Upload") },
        text = { Text("Upload") },
        modifier = modifier.size(height = 56.dp),
        containerColor = MaterialTheme.colorScheme.primaryContainer,
        contentColor = MaterialTheme.colorScheme.onPrimaryContainer,
        elevation = FloatingActionButtonDefaults.elevation(6.dp),
        expanded = true,
        enabled = enabled
    )
}
```

### Código Completo - DocumentPreviewModal.kt

```kotlin
package com.makoclaw.feature.knowledge.presentation.component

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.BottomSheetScaffold
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.material3.rememberStandardBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.makoclaw.core.model.KnowledgeDocument
import com.makoclaw.core.model.ContentType
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DocumentPreviewModal(
    document: KnowledgeDocument?,
    onDismiss: () -> Unit
) {
    val sheetState = rememberModalBottomSheetState(
        confirmValueChange = {
            when (sheetState.isVisible) {
                false -> onDismiss()
                else -> Unit
            }
        },
        skipHalfExpanded = false
    )

    ModalBottomSheet(
        onDismissRequest = { sheetState.hide() },
        sheetState = sheetState,
        containerColor = MaterialTheme.colorScheme.surfaceColor,
        shape = RoundedCornerShape(16.dp),
        scrimColor = Color.Black.copy(alpha = 0.5f)
    ) {
        if (document == null) return

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(600.dp)
                .padding(16.dp)
        ) {
            Column {
                // Header
                TopAppBar(
                    title = { Text(document.title) },
                    navigationIcon = {
                        IconButton(onClick = onDismiss) {
                            Icon(Icons.Filled.Close, "Close")
                        }
                    },
                    colors = TopAppBarDefaults.topAppBarColors(
                        containerColor = Color.Transparent
                    )
                )

                Spacer(modifier = Modifier.height(16.dp))

                // Content
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxWidth()
                        .border(
                            width = 1.dp,
                            color = MaterialTheme.colorScheme.outline
                        )
                        .padding(16.dp)
                ) {
                    when (document.type) {
                        ContentType.PDF, ContentType.MARKDOWN, ContentType.TEXT -> {
                            // Simple text preview for now
                            Text(
                                text = document.content,
                                style = MaterialTheme.typography.bodyMedium,
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .verticalScroll(rememberScrollState()),
                                color = MaterialTheme.colorScheme.onSurface
                            )
                        }
                    }
                }
            }
        }
    }
}
```

### Instrucciones

1. Crear el directorio `feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/component/` si no existe
2. Crear el archivo `UploadFileButton.kt` y pegar el primer bloque de código
3. Crear el archivo `DocumentPreviewModal.kt` y pegar el segundo bloque de código
4. Verificar que no hay errores de compilación

---

## TASK-KNOWLEDGE-006: Crear ChunkEditorSheet

### Ubicación
**Archivo**: `makoclaw-android/feature/feature-knowledge/src/main/java/com/makoclaw/feature/knowledge/presentation/component/ChunkEditorSheet.kt`
**Paquete**: `com.makoclaw.feature.knowledge.presentation.component`

### Código Completo

```kotlin
package com.makoclaw.feature.knowledge.presentation.component

import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.ContentCopy
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.material3.rememberStandardBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.runtime.saveable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.graphics.Color
import com.makoclaw.core.model.KnowledgeDocument
import com.makoclaw.feature.knowledge.presentation.viewmodel.KnowledgeViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChunkEditorSheet(
    document: KnowledgeDocument,
    onDismiss: () -> Unit,
    viewModel: KnowledgeViewModel
) {
    val sheetState = rememberModalBottomSheetState(
        confirmValueChange = {
            when (sheetState.isVisible) {
                false -> onDismiss()
                else -> Unit
            }
        },
        skipHalfExpanded = false
    )

    var editedContent by remember { mutableStateOf("") }
    var isSaving by remember { mutableStateOf(false) }
    var saveTrigger by remember { mutableStateOf(0) }

    // Auto-save with debounce
    LaunchedEffect(document.id, saveTrigger) {
        if (editedContent.isNotBlank() && saveTrigger > 0 && !isSaving) {
            isSaving = true
            delay(1000) // 1 second debounce
            viewModel.onEvent(KnowledgeEvent.EditChunk(document.id, "", editedContent))
            delay(500) // Wait for save
            isSaving = false
        }
    }

    ModalBottomSheet(
        onDismissRequest = { sheetState.hide() },
        sheetState = sheetState,
        containerColor = MaterialTheme.colorScheme.surfaceColor,
        shape = RoundedCornerShape(16.dp),
        scrimColor = Color.Black.copy(alpha = 0.5f)
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(600.dp)
                .padding(16.dp)
        ) {
            Column {
                // Header
                TopAppBar(
                    title = { Text("Edit Document") },
                    navigationIcon = {
                        IconButton(onClick = onDismiss) {
                            Icon(Icons.Filled.Close, "Close")
                        }
                    },
                    colors = TopAppBarDefaults.topAppBarColors(
                        containerColor = Color.Transparent
                    )
                )

                Spacer(modifier = Modifier.height(16.dp))

                // Content
                if (document.chunks.isEmpty()) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(vertical = 16.dp),
                        horizontalAlignment = androidx.compose.foundation.layout.Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.Center
                    ) {
                        Text(
                            text = "This document has no chunks to edit",
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                } else {
                    LazyColumn(
                        modifier = Modifier
                            .weight(1f)
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp),
                        verticalArrangement = Arrangement.spacedBy(16.dp),
                        horizontalAlignment = androidx.compose.foundation.layout.Alignment.CenterHorizontally
                    ) {
                        item { Spacer(modifier = Modifier.height(8.dp)) }

                        items(document.chunks) { chunk ->
                            Column(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .border(
                                        width = 1.dp,
                                        color = MaterialTheme.colorScheme.outline
                                    )
                                    .padding(16.dp)
                            ) {
                                Text(
                                    text = "Chunk ${document.chunks.indexOf(chunk) + 1}",
                                    style = MaterialTheme.typography.titleSmall,
                                    color = MaterialTheme.colorScheme.primary
                                )

                                Spacer(modifier = Modifier.height(8.dp))

                                OutlinedTextField(
                                    value = editedContent,
                                    onValueChange = { editedContent = it },
                                    label = { Text("Content") },
                                    modifier = Modifier.fillMaxWidth(),
                                    maxLines = 10,
                                    shape = RoundedCornerShape(8.dp),
                                    enabled = !isSaving
                                )

                                Spacer(modifier = Modifier.height(16.dp))

                                Row(
                                    horizontalArrangement = Arrangement.End,
                                    verticalAlignment = androidx.compose.layout.Alignment.CenterVertically
                                ) {
                                    if (isSaving) {
                                        CircularProgressIndicator(
                                            modifier = Modifier.size(16.dp),
                                            strokeWidth = 2.dp
                                        )
                                    } else {
                                        Button(
                                            onClick = {
                                                saveTrigger++
                                            viewModel.onEvent(KnowledgeEvent.EditChunk(document.id, document.chunks.indexOf(chunk).toString(), editedContent))
                                            },
                                            enabled = editedContent.isNotBlank(),
                                            colors = ButtonDefaults.buttonColors(
                                                containerColor = MaterialTheme.colorScheme.primary,
                                                contentColor = MaterialTheme.colorScheme.onPrimary
                                            ),
                                            modifier = Modifier.size(height = 40.dp)
                                        ) {
                                            Icon(Icons.Filled.Check, "Save")
                                            Spacer(modifier = Modifier.width(8.dp))
                                            Text("Save")
                                        }
                                    }
                                }
                            }
                        }

                        item { Spacer(modifier = Modifier.height(80.dp)) }
                    }
                }
            }
        }
    }
}
```

### Instrucciones

1. Crear el archivo `ChunkEditorSheet.kt` y pegar todo el código
2. Verificar que no hay errores de compilación

---

## Resumen

### Archivos a Crear (9 archivos)

| # | Archivo | Líneas aprox. | Complejidad |
|--|--------|--------------|------------|
| 1 | KnowledgeUiState.kt | ~20 | Fácil |
| 2 | KnowledgeEvent.kt | ~15 | Fácil |
| 3 | KnowledgeEffect.kt | ~10 | Fácil |
| 4 | KnowledgeViewModel.kt | ~120 | Media |
| 5 | KnowledgeRepository.kt (interface) | ~15 | Fácil |
| 6 | KnowledgeRepositoryImpl.kt | ~90 | Media |
| 7 | KnowledgeApi.kt (si no existe) | ~50 | Media |
| 8 | KnowledgeScreen.kt | ~120 | Media |
| 9 | KnowledgeDocumentCard.kt | ~90 | Media |
| 10 | UploadFileButton.kt | ~80 | Media |
| 11 | DocumentPreviewModal.kt | ~130 | Media |
| 12 | ChunkEditorSheet.kt | ~100 | Media |

### Tiempo Estimado
- **TASK-KNOWLEDGE-001**: 2 horas
- **TASK-KNOWLEDGE-002**: 2 horas
- **TASK-KNOWLEDGE-003**: 2 horas
- **TASK-KNOWLEDGE-004**: 1.5 horas
- **TASK-KNOWLEDGE-005**: 1.5 horas
- **TASK-KNOWLEDGE-006**: 1 hora

**Total BATCH 2**: ~10 horas de trabajo

### Checklist de Validación

Antes de considerar completado el BATCH 2:

- [ ] Todos los archivos compilan sin errores
- [ ] KnowledgeViewModel se conecta a KnowledgeRepository
- [ ] KnowledgeRepository se conecta a KnowledgeApi
- [ ] KnowledgeScreen muestra documentos correctamente
- [ ] UploadFileButton abre file picker
- [ ] DocumentPreviewModal muestra contenido
- [ ] ChunkEditorSheet permite edición
- - [ ] Auto-save funciona con debounce
- [ ] Tests unitarios pasan (si se crearon)
- - [ ] UI tests pasan (si se crearon)

---

## Siguientes Pasos

Después de completar el BATCH 2 (Feature Knowledge), el siguiente es:

**BATCH 3**: Features ALTA (12 tareas)
- feature-workflows (5 tareas) - **🚨 HIGH COMPLEXITY** (Canvas API editor visual)
- feature-metrics (4 tareas) - VICO CHARTS
- feature-cron (3 tareas) - CUSTOM SCHEDULE SELECTOR

Recomendación:
1. **Terminar BATCH 2 primero** (validar que todo funcione)
2. **Luego ir a BATCH 3** (priorizar workflows y metrics)

---

## Notas Importantes

1. **Validación de KnowledgeApi**: Asegúrate que los endpoints del backend existan:
   - `GET /knowledge/documents`
   - `POST /knowledge/upload`
   - `POST /knowledge/search`
   - `DELETE /knowledge/document/{id}`
   - `PATCH /knowledge/document/{id}/chunk`

2. **Pruebas unitarias**: Recomiendo crear tests básicos para ViewModels y Repositories para verificar la lógica

3. **Markwon**: Para el DocumentPreviewModal, asegúrate de integrar la librería Markwon si necesitas renderizado de markdown

4. **Permissions**: Para upload de archivos, recuerda configurar los permisos en el AndroidManifest.xml

5. **DataStore**: Ya está completado en el BATCH 1, puedes usar `PreferencesStore` para persistir preferencias de knowledge

---

## Finalización

Una vez que hayas creado todos los archivos y verificado que compilan, el BATCH 2 estará **COMPLETADO** 🎉

**¡Éxito, hermano!** 🦈
