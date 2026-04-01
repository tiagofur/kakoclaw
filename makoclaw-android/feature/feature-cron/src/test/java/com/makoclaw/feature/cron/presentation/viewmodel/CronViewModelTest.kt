package com.makoclaw.feature.cron.presentation.viewmodel

import app.cash.turbine.test
import com.makoclaw.core.model.CronJob
import com.makoclaw.feature.cron.data.repository.CronRepository
import com.makoclaw.feature.cron.presentation.state.CronEvent
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

class CronViewModelTest {

    private val testDispatcher = StandardTestDispatcher()
    private lateinit var repository: CronRepository
    private lateinit var viewModel: CronViewModel

    private val sampleJobs = listOf(
        CronJob(
            id = "1",
            name = "Daily Report",
            description = "Send daily report",
            cronExpression = "0 9 * * *",
            enabled = true,
            nextRun = "2026-04-02 09:00"
        ),
        CronJob(
            id = "2",
            name = "Weekly Cleanup",
            description = "Clean old sessions",
            cronExpression = "0 2 * * 0",
            enabled = false,
            nextRun = "2026-04-06 02:00"
        )
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

    private fun createViewModel(): CronViewModel {
        every { repository.getJobs() } returns flow { emit(sampleJobs) }
        viewModel = CronViewModel(repository, testDispatcher)
        return viewModel
    }

    @Test
    fun `initial state should be loading`() {
        every { repository.getJobs() } returns flow { emit(sampleJobs) }
        viewModel = CronViewModel(repository, testDispatcher)
        assertTrue(viewModel.uiState.value.isLoading)
    }

    @Test
    fun `load jobs should populate state`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.isLoading)
        assertEquals(2, state.jobs.size)
        assertNull(state.error)
    }

    @Test
    fun `load jobs with error should show error`() = runTest {
        every { repository.getJobs() } returns flow {
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
    fun `create job should call repository`() = runTest {
        coEvery { repository.createJob(any(), any(), any(), any()) } returns Result.success("job-3")
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(
            CronEvent.CreateJob(
                name = "Test Job",
                description = "Test",
                cronExpression = "0 * * * *",
                workflowId = ""
            )
        )
        testDispatcher.scheduler.advanceUntilIdle()

        coEvery { repository.createJob("Test Job", "Test", "0 * * * *", "") }
    }

    @Test
    fun `toggle job should call repository`() = runTest {
        coEvery { repository.toggleJob(any(), any()) } returns Result.success(Unit)
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(CronEvent.ToggleJob("1", false))
        testDispatcher.scheduler.advanceUntilIdle()

        verify { repository.toggleJob("1", false) }
    }

    @Test
    fun `delete job should call repository`() = runTest {
        coEvery { repository.deleteJob(any()) } returns Result.success(Unit)
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(CronEvent.DeleteJob("1"))
        testDispatcher.scheduler.advanceUntilIdle()

        verify { repository.deleteJob("1") }
    }

    @Test
    fun `generate cron should update generated expression`() = runTest {
        coEvery { repository.generateCron(any(), any()) } returns Result.success("0 9 * * 1")
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(CronEvent.GenerateCron("Every Monday at 9am", ""))
        testDispatcher.scheduler.advanceUntilIdle()

        assertEquals("0 9 * * 1", viewModel.uiState.value.generatedExpression)
    }

    @Test
    fun `test run should update test result`() = runTest {
        coEvery { repository.testRun(any()) } returns Result.success(Pair(true, listOf("2026-04-02 09:00")))
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(CronEvent.TestRun("0 9 * * *"))
        testDispatcher.scheduler.advanceUntilIdle()

        val result = viewModel.uiState.value.testResult
        assertNotNull(result)
        assertTrue(result!!.isValid)
        assertEquals(1, result.nextRuns.size)
    }

    @Test
    fun `select job should update selected job`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(CronEvent.SelectJob("1"))

        val state = viewModel.uiState.value
        assertNotNull(state.selectedJob)
        assertEquals("1", state.selectedJob?.id)
        assertEquals("0 9 * * *", state.cronExpression)
    }

    @Test
    fun `refresh should reload from repository`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(CronEvent.Refresh)
        testDispatcher.scheduler.advanceUntilIdle()

        verify(atLeast = 2) { repository.getJobs() }
    }
}
