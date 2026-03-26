package com.makoclaw.core.database.dao

import com.makoclaw.core.database.entity.CronJobEntity
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertFalse
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test

class CronJobDaoTest : BaseDaoTest() {
    @Test
    fun `toggle updates cron job enabled state`() = runTest {
        val entity = CronJobEntity("cron-1", "Nightly", "desc", "0 0 * * *", "workflow-1", true, 1L, 2L)
        val dao = database.cronJobDao()

        dao.insert(entity)
        dao.toggle(entity.id, false)

        assertFalse(dao.getAll().first().single().enabled)
        dao.toggle(entity.id, true)
        assertTrue(dao.getAll().first().single().enabled)
    }
}
