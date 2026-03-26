package com.makoclaw.core.database.dao

import com.makoclaw.core.database.entity.McpServerEntity
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class McpServerDaoTest : BaseDaoTest() {
    @Test
    fun `insert and getAll returns mcp servers`() = runTest {
        val entity = McpServerEntity("mcp-1", "Docs", "https://mcp.example.com", true, "connected", 5L)

        database.mcpServerDao().insert(entity)

        assertEquals(listOf(entity), database.mcpServerDao().getAll().first())
    }
}
