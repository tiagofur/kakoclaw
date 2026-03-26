package com.makoclaw.feature.knowledge.data.repository

import com.makoclaw.core.database.dao.KnowledgeDocumentDao
import com.makoclaw.core.database.entity.KnowledgeDocumentEntity
import com.makoclaw.core.model.KnowledgeChunk
import com.makoclaw.core.model.KnowledgeDocument
import com.makoclaw.core.network.api.EditKnowledgeChunkRequest
import com.makoclaw.core.network.api.KnowledgeApi
import com.makoclaw.core.network.api.KnowledgeDetailResponse
import com.makoclaw.core.network.api.KnowledgeListResponse
import com.makoclaw.core.network.api.KnowledgeSearchRequest
import com.makoclaw.core.network.api.KnowledgeSearchResponse
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.every
import io.mockk.mockk
import java.io.File
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

@OptIn(ExperimentalCoroutinesApi::class)
class KnowledgeRepositoryImplTest {

    private val api = mockk<KnowledgeApi>()
    private val dao = mockk<KnowledgeDocumentDao>(relaxed = true)
    private val json = Json { encodeDefaults = true; ignoreUnknownKeys = true }

    @Test
    fun `getDocuments maps cached entities`() = runTest {
        every { dao.getAll() } returns flowOf(listOf(sampleEntity()))
        coEvery { api.getDocuments() } returns KnowledgeListResponse(emptyList())
        val repository = KnowledgeRepositoryImpl(api, dao, json)

        val result = repository.getDocuments().testAndFirst()

        assertEquals("doc-1", result.single().id)
        assertEquals("Doc.md", result.single().title)
        coVerify { api.getDocuments() }
    }

    @Test
    fun `uploadDocument emits progress and caches uploaded document`() = runTest {
        val file = kotlin.io.path.createTempFile(suffix = ".md").toFile().apply { writeText("hello") }
        coEvery { api.uploadDocument(any()) } returns sampleDocument()
        val repository = KnowledgeRepositoryImpl(api, dao, json)

        val emissions = repository.uploadDocument(file).toList()

        assertEquals(1f, emissions.last().progress)
        coVerify { dao.insert(any()) }
        file.delete()
    }

    @Test
    fun `searchDocuments groups search chunks by document`() = runTest {
        coEvery { api.searchDocuments(KnowledgeSearchRequest("rag")) } returns KnowledgeSearchResponse(
            results = listOf(
                KnowledgeChunk(id = "chunk-1", documentId = "doc-1", title = "Doc.md", content = "match")
            )
        )
        coEvery { dao.getById("doc-1") } returns sampleEntity()
        val repository = KnowledgeRepositoryImpl(api, dao, json)

        val result = repository.searchDocuments("rag").testAndFirst()

        assertEquals("doc-1", result.single().id)
        assertEquals(1, result.single().chunks.size)
    }

    @Test
    fun `deleteDocument deletes remote and local cache`() = runTest {
        coEvery { api.deleteDocument("doc-1") } returns Unit
        val repository = KnowledgeRepositoryImpl(api, dao, json)

        repository.deleteDocument("doc-1").testAndFirst()

        coVerify { api.deleteDocument("doc-1") }
        coVerify { dao.delete("doc-1") }
    }

    @Test
    fun `editChunk updates cached document from api response`() = runTest {
        val response = KnowledgeDetailResponse(
            document = sampleDocument(),
            chunks = listOf(KnowledgeChunk(id = "chunk-1", documentId = "doc-1", title = "Doc.md", content = "updated"))
        )
        coEvery {
            api.editChunk("doc-1", EditKnowledgeChunkRequest(chunkId = "chunk-1", content = "updated"))
        } returns response
        val repository = KnowledgeRepositoryImpl(api, dao, json)

        repository.editChunk("doc-1", "chunk-1", "updated").testAndFirst()

        coVerify { dao.insert(any()) }
    }
}

private suspend fun <T> kotlinx.coroutines.flow.Flow<T>.testAndFirst(): T = first()

private suspend fun <T> kotlinx.coroutines.flow.Flow<T>.toList(): List<T> {
    val values = mutableListOf<T>()
    collect { values += it }
    return values
}

private fun sampleDocument() = KnowledgeDocument(
    id = "doc-1",
    title = "Doc.md",
    filename = "Doc.md",
    type = "md",
    content = "hello",
    chunks = listOf(KnowledgeChunk(id = "chunk-1", documentId = "doc-1", title = "Doc.md", content = "hello")),
    chunkCount = 1,
    createdAt = "1000",
    updatedAt = "2000"
)

private fun sampleEntity() = KnowledgeDocumentEntity(
    id = "doc-1",
    title = "Doc.md",
    content = "hello",
    chunks = "[{\"id\":\"chunk-1\",\"documentId\":\"doc-1\",\"title\":\"Doc.md\",\"content\":\"hello\",\"metadata\":\"\",\"highlightedContent\":\"\"}]",
    createdAt = 1000L,
    updatedAt = 2000L
)
