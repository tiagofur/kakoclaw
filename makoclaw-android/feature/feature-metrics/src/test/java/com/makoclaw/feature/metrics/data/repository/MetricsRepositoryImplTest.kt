package com.makoclaw.feature.metrics.data.repository

import com.makoclaw.core.model.MetricsData
import com.makoclaw.core.model.TimeRange
import com.makoclaw.core.network.api.ExportRequest
import com.makoclaw.core.network.api.ExportResult
import com.makoclaw.core.network.api.MetricsApi
import com.makoclaw.core.network.api.MetricsResponse
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import io.mockk.coEvery
import io.mockk.mockk
import io.mockk.verify

class MetricsRepositoryImplTest {

    private lateinit var api: MetricsApi
    private lateinit var repository: MetricsRepositoryImpl

    private val sampleMetricsData = MetricsData(
        executionsOverTime = listOf(10f, 20f, 15f),
        agentUsage = listOf(25f, 75f),
        agentLabels = listOf("Agent A", "Agent B"),
        successRate = 0.9f,
        totalExecutions = 100,
        avgDuration = 1.5f,
        activeAgents = 2,
        costDistribution = mapOf("Agent A" to 20f, "Agent B" to 80f)
    )

    @BeforeEach
    fun setup() {
        api = mockk()
        repository = MetricsRepositoryImpl(api)
    }

    @Test
    fun `getMetrics should return data from API`() = runTest {
        coEvery { api.getMetrics(TimeRange.WEEK, null) } returns MetricsResponse(sampleMetricsData)

        val result = repository.getMetrics(TimeRange.WEEK, null).first()

        assertEquals(100, result.totalExecutions)
        assertEquals(0.9f, result.successRate, 0.01f)
        assertEquals(listOf("Agent A", "Agent B"), result.agentLabels)
    }

    @Test
    fun `getMetrics with null response metrics should return empty MetricsData`() = runTest {
        coEvery { api.getMetrics(TimeRange.MONTH, null) } returns MetricsResponse(null)

        val result = repository.getMetrics(TimeRange.MONTH, null).first()

        assertEquals(0, result.totalExecutions)
        assertTrue(result.executionsOverTime.isEmpty())
    }

    @Test
    fun `getMetrics should pass agent filter to API`() = runTest {
        coEvery { api.getMetrics(TimeRange.WEEK, "agent-123") } returns MetricsResponse(sampleMetricsData)

        repository.getMetrics(TimeRange.WEEK, "agent-123").first()

        verify { api.getMetrics(TimeRange.WEEK, "agent-123") }
    }

    @Test
    fun `exportMetrics should return export result`() = runTest {
        val exportResult = ExportResult(filePath = "/exports/metrics.csv", fileSize = 2048L)
        coEvery { api.exportMetrics(any()) } returns exportResult

        val result = repository.exportMetrics(TimeRange.WEEK, null).first()

        assertEquals("/exports/metrics.csv", result.filePath)
        assertEquals(2048L, result.fileSize)
    }

    @Test
    fun `exportMetrics should send correct format in request`() = runTest {
        val exportResult = ExportResult()
        coEvery { api.exportMetrics(any()) } returns exportResult

        repository.exportMetrics(TimeRange.DAY, "agent-1").first()

        verify { api.exportMetrics(match { req ->
            req.timeRange == "DAY" && req.agentId == "agent-1" && req.format == "csv"
        }) }
    }
}
