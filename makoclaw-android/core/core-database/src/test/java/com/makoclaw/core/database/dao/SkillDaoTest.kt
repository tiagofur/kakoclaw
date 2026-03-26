package com.makoclaw.core.database.dao

import com.makoclaw.core.database.entity.SkillEntity
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class SkillDaoTest : BaseDaoTest() {
    @Test
    fun `insert and getAll returns skills`() = runTest {
        val entity = SkillEntity("skill-1", "Skill", "desc", "author", 4.5, "productivity", true, 5L)

        database.skillDao().insert(entity)

        assertEquals(listOf(entity), database.skillDao().getAll().first())
    }
}
