package com.makoclaw.core.database.dao

import com.makoclaw.core.database.entity.FileEntryEntity
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class FileEntryDaoTest : BaseDaoTest() {
    @Test
    fun `insert and getAll returns file entries`() = runTest {
        val entity = FileEntryEntity("file-1", "report.csv", "/reports/report.csv", "csv", 100L, 9L)

        database.fileEntryDao().insert(entity)

        assertEquals(listOf(entity), database.fileEntryDao().getAll().first())
    }
}
