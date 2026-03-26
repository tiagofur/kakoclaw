package com.makoclaw.core.database.dao

import com.makoclaw.core.database.entity.WorkflowEntity
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class WorkflowDaoTest : BaseDaoTest() {
    @Test
    fun `insert and getAll returns workflows`() = runTest {
        val entity = WorkflowEntity("wf-1", "Workflow", "desc", "[]", "[]", "draft", 1L, 2L, null)

        database.workflowDao().insert(entity)

        assertEquals(listOf(entity), database.workflowDao().getAll().first())
    }
}
