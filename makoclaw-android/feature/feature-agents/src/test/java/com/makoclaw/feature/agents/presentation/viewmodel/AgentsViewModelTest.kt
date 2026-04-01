package com.makoclaw.feature.agents.presentation.viewmodel

import com.makoclaw.core.model.AgentMetrics
import com.makoclaw.core.model.OrchestratorConfig
import com.makoclaw.core.model.Specialist
import com.makoclaw.core.model.Swarm
import com.makoclaw.feature.agents.data.repository.AgentsRepository
import com.makoclaw.feature.agents.presentation.state.AgentsEvent
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
import io.mockk.coVerify
import io.mockk.every
import io.mockk.mockk

class AgentsViewModelTest {

    private val testDispatcher = StandardTestDispatcher()
    private lateinit var repository: AgentsRepository
    private lateinit var viewModel: AgentsViewModel

    private val sampleSpecialists = listOf(
        Specialist(name = "coder", speciality = "code generation", systemPrompt = "You write code"),
        Specialist(name = "reviewer", speciality = "code review", systemPrompt = "You review code")
    )

    private val sampleSwarms = listOf(
        Swarm(name = "dev-team", members = listOf("coder", "reviewer"), mode = "sequential")
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

    private fun createViewModel(): AgentsViewModel {
        every { repository.getSpecialists() } returns flow { emit(sampleSpecialists) }
        coEvery { repository.getOrchestrator() } returns Result.success(OrchestratorConfig(enabled = false))
        coEvery { repository.getMetrics() } returns Result.success(AgentMetrics())
        coEvery { repository.getSwarms() } returns Result.success(sampleSwarms)
        viewModel = AgentsViewModel(repository, testDispatcher)
        return viewModel
    }

    @Test
    fun `initial state should be loading`() {
        every { repository.getSpecialists() } returns flow { emit(emptyList()) }
        coEvery { repository.getOrchestrator() } returns Result.success(OrchestratorConfig())
        viewModel = AgentsViewModel(repository, testDispatcher)
        assertTrue(viewModel.uiState.value.isLoading)
    }

    @Test
    fun `load agents should populate specialists`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.isLoading)
        assertEquals(2, state.specialists.size)
        assertNull(state.error)
    }

    @Test
    fun `load agents should populate swarms`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals(1, state.swarms.size)
        assertEquals("dev-team", state.swarms[0].name)
    }

    @Test
    fun `create specialist should call repository`() = runTest {
        val newSpecialist = Specialist(name = "tester", speciality = "testing", systemPrompt = "Test code")
        coEvery { repository.createSpecialist(any(), any(), any()) } returns Result.success(newSpecialist)
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(AgentsEvent.CreateSpecialist("tester", "testing", "Test code"))
        testDispatcher.scheduler.advanceUntilIdle()

        coVerify { repository.createSpecialist("tester", "testing", "Test code") }
    }

    @Test
    fun `delete specialist should call repository`() = runTest {
        coEvery { repository.deleteSpecialist(any()) } returns Result.success(Unit)
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(AgentsEvent.DeleteSpecialist("coder"))
        testDispatcher.scheduler.advanceUntilIdle()

        coVerify { repository.deleteSpecialist("coder") }
    }

    @Test
    fun `generate specialist should call repository`() = runTest {
        val generated = Specialist(name = "ai-agent", speciality = "auto-generated")
        coEvery { repository.generateSpecialist(any()) } returns Result.success(generated)
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(AgentsEvent.GenerateSpecialist("A testing specialist"))
        testDispatcher.scheduler.advanceUntilIdle()

        assertFalse(viewModel.uiState.value.isGenerating)
        coVerify { repository.generateSpecialist("A testing specialist") }
    }

    @Test
    fun `toggle orchestrator should update state`() = runTest {
        coEvery { repository.toggleOrchestrator(any()) } returns Result.success(Unit)
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(AgentsEvent.ToggleOrchestrator(true))
        testDispatcher.scheduler.advanceUntilIdle()

        assertTrue(viewModel.uiState.value.orchestrator.enabled)
    }

    @Test
    fun `select specialist should update selected`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(AgentsEvent.SelectSpecialist("coder"))

        assertNotNull(viewModel.uiState.value.selectedSpecialist)
        assertEquals("coder", viewModel.uiState.value.selectedSpecialist?.name)
    }

    @Test
    fun `clear selection should remove selected`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(AgentsEvent.SelectSpecialist("coder"))
        viewModel.onEvent(AgentsEvent.ClearSelection)

        assertNull(viewModel.uiState.value.selectedSpecialist)
    }

    @Test
    fun `generate specialist with failure should not crash`() = runTest {
        coEvery { repository.generateSpecialist(any()) } returns Result.failure(RuntimeException("AI error"))
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(AgentsEvent.GenerateSpecialist("fail"))
        testDispatcher.scheduler.advanceUntilIdle()

        assertFalse(viewModel.uiState.value.isGenerating)
    }

    @Test
    fun `refresh should reload all data`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(AgentsEvent.Refresh)
        testDispatcher.scheduler.advanceUntilIdle()

        coVerify(atLeast = 2) { repository.getOrchestrator() }
    }
}
