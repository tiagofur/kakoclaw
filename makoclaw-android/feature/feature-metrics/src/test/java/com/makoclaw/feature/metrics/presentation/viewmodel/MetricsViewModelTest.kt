package com.makoclaw.feature.metrics.presentation.viewmodel

import app.cash.turbine.test
import com.makoclaw.core.model.MetricsData
import com.makoclaw.core.model.TimeRange
import com.makoclaw.core.network.api.ExportResult
import com.makoclaw.feature.metrics.data.repository.MetricsRepository
import com.makoclaw.feature.metrics.presentation.state.MetricsEvent
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
import org.junit.jupiter.api.extension.ExtendWith
import io.mockk.coEvery
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify

@ExtendWith()
class MetricsViewModelTest {

    private val testDispatcher = StandardTestDispatcher()
    private lateinit var repository: MetricsRepository
    private lateinit var viewModel: MetricsViewModel

    private val sampleMetrics = MetricsData(
        executionsOverTime = listOf(10f, 20f, 15f, 30f),
        agentUsage = listOf(25f, 50f, 25f),
        agentLabels = listOf("Agent A", "Agent B", "Agent C"),
        successRate = 0.85f,
        totalExecutions = 120,
        avgDuration = 2.5f,
        activeAgents = 3,
        costDistribution = mapOf("Agent A" to 15f, "Agent B" to 30f)
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

    private fun createViewModel(): MetricsViewModel {
        every { repository.getMetrics(any(), any()) } returns flow { emit(sampleMetrics) }
        viewModel = MetricsViewModel(repository, testDispatcher)
        return viewModel
    }

    @Test
    fun `initial state should be loading`() {
        every { repository.getMetrics(any(), any()) } returns flow { emit(sampleMetrics) }
        viewModel = MetricsViewModel(repository, testDispatcher)

        // Before dispatcher advances, initial load is pending
        val state = viewModel.uiState.value
        assertTrue(state.isLoading)
    }

    @Test
    fun `load metrics should update state with data`() = runTest {
        every { repository.getMetrics(TimeRange.WEEK, null) } returns flow { emit(sampleMetrics) }
        viewModel = MetricsViewModel(repository, testDispatcher)

        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.isLoading)
        assertEquals(120, state.metricsData.totalExecutions)
        assertEquals(0.85f, state.metricsData.successRate, 0.01f)
        assertNull(state.error)
    }

    @Test
    fun `load metrics with error should show error state`() = runTest {
        every { repository.getMetrics(any(), any()) } returns flow {
            throw RuntimeException("Network error")
        }

        viewModel = MetricsViewModel(repository, testDispatcher)
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.isLoading)
        assertNotNull(state.error)
        assertEquals("Network error", state.error)
    }

    @Test
    fun `set time range should update state and reload`() = runTest {
        val weeklyData = sampleMetrics
        val dailyData = sampleMetrics.copy(totalExecutions = 50)

        every { repository.getMetrics(TimeRange.WEEK, null) } returns flow { emit(weeklyData) }
        every { repository.getMetrics(TimeRange.DAY, null) } returns flow { emit(dailyData) }

        viewModel = MetricsViewModel(repository, testDispatcher)
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(MetricsEvent.SetTimeRange(TimeRange.DAY))
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals(TimeRange.DAY, state.selectedTimeRange)
        assertEquals(50, state.metricsData.totalExecutions)
    }

    @Test
    fun `set agent filter should update state and reload`() = runTest {
        val allData = sampleMetrics
        val filteredData = sampleMetrics.copy(totalExecutions = 40)

        every { repository.getMetrics(TimeRange.WEEK, null) } returns flow { emit(allData) }
        every { repository.getMetrics(TimeRange.WEEK, "agent-a") } returns flow { emit(filteredData) }

        viewModel = MetricsViewModel(repository, testDispatcher)
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(MetricsEvent.SetAgentFilter("agent-a"))
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals("agent-a", state.selectedAgentId)
        assertEquals(40, state.metricsData.totalExecutions)
    }

    @Test
    fun `export metrics should emit ExportComplete effect`() = runTest {
        val exportResult = ExportResult(filePath = "/data/metrics.csv", fileSize = 1024L)
        every { repository.getMetrics(any(), any()) } returns flow { emit(sampleMetrics) }
        every { repository.exportMetrics(any(), any()) } returns flow { emit(exportResult) }

        viewModel = MetricsViewModel(repository, testDispatcher)
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(MetricsEvent.ExportMetrics)
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.isExporting)
        assertEquals("/data/metrics.csv", state.exportFilePath)
    }

    @Test
    fun `available agents should be extracted from metrics data`() = runTest {
        every { repository.getMetrics(any(), any()) } returns flow { emit(sampleMetrics) }

        viewModel = MetricsViewModel(repository, testDispatcher)
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals(listOf("Agent A", "Agent B", "Agent C"), state.availableAgents)
    }

    @Test
    fun `refresh should reload metrics`() = runTest {
        every { repository.getMetrics(any(), any()) } returns flow { emit(sampleMetrics) }

        viewModel = MetricsViewModel(repository, testDispatcher)
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(MetricsEvent.Refresh)
        testDispatcher.scheduler.advanceUntilIdle()

        verify(atLeast = 2) { repository.getMetrics(any(), any()) }
    }
}
