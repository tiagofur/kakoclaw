package com.makoclaw.feature.knowledge.presentation.viewmodel

import app.cash.turbine.test
import com.makoclaw.core.model.KnowledgeChunk
import com.makoclaw.core.model.KnowledgeDocument
import com.makoclaw.core.model.UploadProgress
import com.makoclaw.feature.knowledge.data.repository.KnowledgeRepository
import com.makoclaw.feature.knowledge.presentation.state.KnowledgeEffect
import com.makoclaw.feature.knowledge.presentation.state.KnowledgeEvent
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertFalse
import org.junit.jupiter.api.Assertions.assertNull
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import java.io.File

@OptIn(ExperimentalCoroutinesApi::class)
class KnowledgeViewModelTest {

    private val dispatcher = StandardTestDispatcher()
    private lateinit var repository: KnowledgeRepository

    @BeforeEach
    fun setUp() {
        Dispatchers.setMain(dispatcher)
        repository = mockk()
        every { repository.getDocuments() } returns flowOf(listOf(sampleDocument()))
        every { repository.searchDocuments(any()) } returns flowOf(listOf(sampleDocument(id = "search-result")))
        every { repository.uploadDocument(any()) } returns flowOf(
            UploadProgress(progress = 0.4f),
            UploadProgress(progress = 1f, document = sampleDocument())
        )
        every { repository.deleteDocument(any()) } returns flowOf(Unit)
        every { repository.editChunk(any(), any(), any()) } returns flowOf(Unit)
    }

    @AfterEach
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `init loads documents into success state`() = runTest(dispatcher) {
        val viewModel = KnowledgeViewModel(repository, dispatcher)

        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.isLoading)
        assertEquals(1, state.documents.size)
        assertFalse(state.isEmpty)
        assertNull(state.error)
    }

    @Test
    fun `search documents updates query after debounce`() = runTest(dispatcher) {
        val viewModel = KnowledgeViewModel(repository, dispatcher)

        viewModel.onEvent(KnowledgeEvent.SearchDocuments("rag"))
        advanceTimeBy(300)
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals("rag", state.searchQuery)
        assertEquals("search-result", state.documents.single().id)
    }

    @Test
    fun `upload document emits progress and success snackbar`() = runTest(dispatcher) {
        val viewModel = KnowledgeViewModel(repository, dispatcher)

        viewModel.effects.test {
            viewModel.onEvent(KnowledgeEvent.UploadDocument(File("notes.md")))
            advanceUntilIdle()

            assertTrue(awaitItem() is KnowledgeEffect.UploadProgress)
            assertTrue(awaitItem() is KnowledgeEffect.UploadProgress)
            assertEquals(
                KnowledgeEffect.ShowSnackbar("Documento subido correctamente"),
                awaitItem()
            )
            cancelAndIgnoreRemainingEvents()
        }

        assertNull(viewModel.uiState.value.uploadProgress)
    }

    @Test
    fun `select document updates selection and emits preview navigation`() = runTest(dispatcher) {
        val viewModel = KnowledgeViewModel(repository, dispatcher)
        advanceUntilIdle()

        viewModel.effects.test {
            viewModel.onEvent(KnowledgeEvent.SelectDocument("doc-1"))
            advanceUntilIdle()

            assertEquals("doc-1", viewModel.uiState.value.selectedDocument?.id)
            assertEquals(KnowledgeEffect.NavigateToPreview("doc-1"), awaitItem())
            cancelAndIgnoreRemainingEvents()
        }
    }

    @Test
    fun `delete document delegates to repository`() = runTest(dispatcher) {
        val viewModel = KnowledgeViewModel(repository, dispatcher)

        viewModel.onEvent(KnowledgeEvent.DeleteDocument("doc-1"))
        advanceUntilIdle()

        verify { repository.deleteDocument("doc-1") }
    }
}

private fun sampleDocument(id: String = "doc-1") = KnowledgeDocument(
    id = id,
    title = "Guide.md",
    filename = "Guide.md",
    type = "md",
    content = "# Heading",
    chunks = listOf(KnowledgeChunk(id = "chunk-1", documentId = id, title = "Intro", content = "Hello")),
    chunkCount = 1,
    createdAt = "1000",
    updatedAt = "2000"
)
