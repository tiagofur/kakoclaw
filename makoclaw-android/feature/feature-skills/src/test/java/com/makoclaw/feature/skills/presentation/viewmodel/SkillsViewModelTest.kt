package com.makoclaw.feature.skills.presentation.viewmodel

import com.makoclaw.core.model.Skill
import com.makoclaw.feature.skills.data.repository.SkillsRepository
import com.makoclaw.feature.skills.presentation.state.SkillsEvent
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertFalse
import org.junit.jupiter.api.Assertions.assertNotNull
import org.junit.jupiter.api.Assertions.assertNull
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import io.mockk.coEvery
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify

class SkillsViewModelTest {

    private val testDispatcher = StandardTestDispatcher()
    private lateinit var repository: SkillsRepository
    private lateinit var viewModel: SkillsViewModel

    private val sampleInstalledSkills = listOf(
        Skill(name = "web-search", content = "Search the web", repository = "github.com/example/web-search"),
        Skill(name = "code-review", content = "Review code quality", repository = "github.com/example/code-review")
    )

    private val sampleMarketplaceSkills = listOf(
        Skill(name = "data-analyzer", content = "Analyze data sets", repository = "github.com/market/data-analyzer"),
        Skill(name = "doc-generator", content = "Generate documentation", repository = "github.com/market/doc-generator")
    )

    @BeforeEach
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        repository = mockk(relaxed = true)
    }

    @AfterEach
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun createViewModel(): SkillsViewModel {
        every { repository.getInstalled() } returns flow { emit(sampleInstalledSkills) }
        viewModel = SkillsViewModel(repository, testDispatcher)
        return viewModel
    }

    @Test
    fun `initial state should be loading`() {
        every { repository.getInstalled() } returns flow { emit(sampleInstalledSkills) }
        viewModel = SkillsViewModel(repository, testDispatcher)
        assertTrue(viewModel.uiState.value.isLoading)
    }

    @Test
    fun `load installed should populate state`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.isLoading)
        assertEquals(2, state.installedSkills.size)
        assertNull(state.error)
    }

    @Test
    fun `load installed with error should show error`() = runTest {
        every { repository.getInstalled() } returns flow {
            throw RuntimeException("API failure")
        }
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.isLoading)
        assertNotNull(state.error)
        assertEquals("API failure", state.error)
    }

    @Test
    fun `select tab should update selected tab`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(SkillsEvent.SelectTab(1))
        assertEquals(1, viewModel.uiState.value.selectedTab)
    }

    @Test
    fun `select marketplace tab should trigger load marketplace`() = runTest {
        coEvery { repository.getMarketplace() } returns Result.success(sampleMarketplaceSkills)
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(SkillsEvent.SelectTab(1))
        testDispatcher.scheduler.advanceUntilIdle()

        coEvery { repository.getMarketplace() }
    }

    @Test
    fun `install skill should call repository`() = runTest {
        coEvery { repository.installSkill(any()) } returns Result.success(Unit)
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(SkillsEvent.InstallSkill("github.com/test/skill"))
        testDispatcher.scheduler.advanceUntilIdle()

        coEvery { repository.installSkill("github.com/test/skill") }
    }

    @Test
    fun `uninstall skill should call repository`() = runTest {
        coEvery { repository.uninstallSkill(any()) } returns Result.success(Unit)
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(SkillsEvent.UninstallSkill("web-search"))
        testDispatcher.scheduler.advanceUntilIdle()

        verify { repository.uninstallSkill("web-search") }
    }

    @Test
    fun `generate skill should call repository and update state`() = runTest {
        val generatedSkill = Skill(name = "custom-skill", content = "Generated content")
        coEvery { repository.generateSkill(any(), any()) } returns Result.success(generatedSkill)
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(SkillsEvent.GenerateSkill("custom-skill", "Do custom things"))
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.isGenerating)
        assertNotNull(state.generatedSkill)
        assertEquals("custom-skill", state.generatedSkill?.name)
    }

    @Test
    fun `generate skill with failure should not crash`() = runTest {
        coEvery { repository.generateSkill(any(), any()) } returns Result.failure(RuntimeException("AI error"))
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(SkillsEvent.GenerateSkill("fail", "fail"))
        testDispatcher.scheduler.advanceUntilIdle()

        assertFalse(viewModel.uiState.value.isGenerating)
    }

    @Test
    fun `update search query should update state`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(SkillsEvent.UpdateSearchQuery("web"))

        assertEquals("web", viewModel.uiState.value.searchQuery)
    }

    @Test
    fun `set category filter should update state`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(SkillsEvent.SetCategoryFilter("search"))

        assertEquals("search", viewModel.uiState.value.categoryFilter)
    }

    @Test
    fun `clear error should remove error`() = runTest {
        every { repository.getInstalled() } returns flow {
            throw RuntimeException("Error")
        }
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        assertNotNull(viewModel.uiState.value.error)

        viewModel.onEvent(SkillsEvent.ClearError)
        assertNull(viewModel.uiState.value.error)
    }
}
