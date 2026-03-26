package com.makoclaw.core.database.dao

import com.makoclaw.core.database.entity.ReportEntity
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class ReportDaoTest : BaseDaoTest() {
    @Test
    fun `insert and getAll returns reports`() = runTest {
        val entity = ReportEntity("report-1", "Weekly", "metrics", "pdf", 10L, 20L)

        database.reportDao().insert(entity)

        assertEquals(listOf(entity), database.reportDao().getAll().first())
    }
}
