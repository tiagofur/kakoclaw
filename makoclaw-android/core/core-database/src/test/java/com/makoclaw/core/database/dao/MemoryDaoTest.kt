package com.makoclaw.core.database.dao

import com.makoclaw.core.database.entity.MemoryEntity
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class MemoryDaoTest : BaseDaoTest() {
    @Test
    fun `insert and getAll returns memories`() = runTest {
        val entity = MemoryEntity("mem-1", "long-term", "remember this", 11L)

        database.memoryDao().insert(entity)

        assertEquals(listOf(entity), database.memoryDao().getAll().first())
    }
}
