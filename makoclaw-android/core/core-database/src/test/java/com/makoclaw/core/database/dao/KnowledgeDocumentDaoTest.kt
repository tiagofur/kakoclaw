package com.makoclaw.core.database.dao

import com.makoclaw.core.database.entity.KnowledgeDocumentEntity
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class KnowledgeDocumentDaoTest : BaseDaoTest() {
    @Test
    fun `insert and getAll returns knowledge documents`() = runTest {
        val entity = KnowledgeDocumentEntity("doc-1", "Doc", "content", "[]", 1L, 2L)

        database.knowledgeDocumentDao().insert(entity)

        assertEquals(listOf(entity), database.knowledgeDocumentDao().getAll().first())
    }
}
