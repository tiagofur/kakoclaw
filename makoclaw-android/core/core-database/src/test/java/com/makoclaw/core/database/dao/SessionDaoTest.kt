package com.makoclaw.core.database.dao

import com.makoclaw.core.database.entity.SessionEntity
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test

class SessionDaoTest : BaseDaoTest() {
    @Test
    fun `archive updates archived flag`() = runTest {
        val entity = SessionEntity("session-1", "Chat", 3, "hola", 7L, false)
        val dao = database.sessionDao()

        dao.insert(entity)
        dao.archive(entity.id, true)

        assertTrue(dao.getAll().first().single().archived)
    }
}
