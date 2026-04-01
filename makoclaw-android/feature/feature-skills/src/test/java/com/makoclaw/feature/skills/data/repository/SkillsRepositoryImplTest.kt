package com.makoclaw.feature.skills.data.repository

import com.makoclaw.core.database.dao.SkillDao
import com.makoclaw.core.database.entity.SkillEntity
import com.makoclaw.core.model.Skill
import com.makoclaw.core.network.api.GenerateSkillRequest
import com.makoclaw.core.network.api.InstallSkillRequest
import com.makoclaw.core.network.api.SkillsApi
import com.makoclaw.core.network.api.SkillsListResponse
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.every
import io.mockk.mockk

class SkillsRepositoryImplTest {

    private lateinit var api: SkillsApi
    private lateinit var dao: SkillDao
    private lateinit var repository: SkillsRepositoryImpl

    private val sampleEntities = listOf(
        SkillEntity(
            id = "web-search",
            name = "web-search",
            description = "Search the web for information",
            author = "team",
            rating = 4.5,
            category = "search",
            installed = true,
            installedAt = 1000L
        ),
        SkillEntity(
            id = "data-analyzer",
            name = "data-analyzer",
            description = "Analyze data sets",
            author = "market",
            rating = 3.8,
            category = "data",
            installed = false,
            installedAt = null
        )
    )

    @BeforeEach
    fun setup() {
        api = mockk()
        dao = mockk(relaxed = true)
        repository = SkillsRepositoryImpl(api, dao)
    }

    @Test
    fun `getInstalled should return only installed entities from dao`() = runTest {
        every { dao.getAll() } returns flowOf(sampleEntities)
        coEvery { api.listSkills() } returns SkillsListResponse(emptyList())

        val skills = repository.getInstalled().first()

        assertEquals(1, skills.size)
        assertEquals("web-search", skills[0].name)
    }

    @Test
    fun `getInstalled should refresh from API`() = runTest {
        val remoteSkills = listOf(Skill(name = "new-skill", content = "New skill"))
        every { dao.getAll() } returns flowOf(emptyList())
        coEvery { api.listSkills() } returns SkillsListResponse(remoteSkills)

        repository.getInstalled().first()

        coVerify { api.listSkills() }
    }

    @Test
    fun `getMarketplace should return skills from API`() = runTest {
        val marketplaceSkills = listOf(
            Skill(name = "market-skill", content = "From marketplace", repository = "github.com/test")
        )
        coEvery { api.getMarketplaceSkills() } returns SkillsListResponse(marketplaceSkills)

        val result = repository.getMarketplace()

        assertTrue(result.isSuccess)
        assertEquals(1, result.getOrDefault(emptyList()).size)
        assertEquals("market-skill", result.getOrDefault(emptyList())[0].name)
    }

    @Test
    fun `getMarketplace with API failure should return failure`() = runTest {
        coEvery { api.getMarketplaceSkills() } throws RuntimeException("Server error")

        val result = repository.getMarketplace()

        assertTrue(result.isFailure)
    }

    @Test
    fun `installSkill should call API and refresh`() = runTest {
        coEvery { api.installSkill(any()) } returns Unit
        coEvery { api.listSkills() } returns SkillsListResponse(emptyList())

        val result = repository.installSkill("github.com/test/skill")

        assertTrue(result.isSuccess)
        coVerify { api.installSkill(any()) }
    }

    @Test
    fun `uninstallSkill should call API and refresh`() = runTest {
        coEvery { api.uninstallSkill(any()) } returns Unit
        coEvery { api.listSkills() } returns SkillsListResponse(emptyList())

        val result = repository.uninstallSkill("web-search")

        assertTrue(result.isSuccess)
        coVerify { api.uninstallSkill("web-search") }
    }

    @Test
    fun `generateSkill should return generated skill from API`() = runTest {
        val generated = Skill(name = "ai-skill", content = "AI generated")
        coEvery { api.generateSkill(any()) } returns generated
        coEvery { api.listSkills() } returns SkillsListResponse(emptyList())

        val result = repository.generateSkill("ai-skill", "Do AI things")

        assertTrue(result.isSuccess)
        assertEquals("ai-skill", result.getOrDefault(Skill("")).name)
    }

    @Test
    fun `installSkill with API failure should return failure`() = runTest {
        coEvery { api.installSkill(any()) } throws RuntimeException("Install failed")

        val result = repository.installSkill("github.com/test/fail")

        assertTrue(result.isFailure)
    }
}
