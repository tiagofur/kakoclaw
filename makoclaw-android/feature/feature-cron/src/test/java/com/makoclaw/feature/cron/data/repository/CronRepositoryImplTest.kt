package com.makoclaw.feature.cron.data.repository

import com.makoclaw.core.database.dao.CronJobDao
import com.makoclaw.core.database.entity.CronJobEntity
import com.makoclaw.core.network.api.*
import com.makoclaw.core.model.CronJob
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

class CronRepositoryImplTest {

    private lateinit var api: CronApi
    private lateinit var dao: CronJobDao
    private lateinit var repository: CronRepositoryImpl

    private val sampleEntities = listOf(
        CronJobEntity(
            id = "1",
            name = "Daily Report",
            description = "Send report",
            cronExpression = "0 9 * * *",
            workflowId = "wf-1",
            enabled = true,
            createdAt = 1000L,
            updatedAt = 1000L
        )
    )

    @BeforeEach
    fun setup() {
        api = mockk()
        dao = mockk(relaxed = true)
        repository = CronRepositoryImpl(api, dao)
    }

    @Test
    fun `getJobs should return entities from dao`() = runTest {
        every { dao.getAll() } returns flowOf(sampleEntities)
        coEvery { api.getJobs() } returns CronJobsResponse(emptyList())

        val jobs = repository.getJobs().first()

        assertEquals(1, jobs.size)
        assertEquals("Daily Report", jobs[0].name)
        assertEquals("0 9 * * *", jobs[0].cronExpression)
    }

    @Test
    fun `getJobs should refresh from API`() = runTest {
        val remoteJobs = listOf(
            CronJob(id = "2", name = "Weekly", cronExpression = "0 0 * * 0")
        )
        every { dao.getAll() } returns flowOf(emptyList())
        coEvery { api.getJobs() } returns CronJobsResponse(remoteJobs)

        repository.getJobs().first()

        coVerify { api.getJobs() }
    }

    @Test
    fun `createJob should call API and refresh`() = runTest {
        coEvery { api.createJob(any()) } returns ApiResponse(success = true)
        coEvery { api.getJobs() } returns CronJobsResponse(emptyList())

        val result = repository.createJob("Test", "Desc", "0 * * * *", "wf-1")

        assertTrue(result.isSuccess)
        coVerify { api.createJob(any()) }
    }

    @Test
    fun `toggleJob should call API and update dao`() = runTest {
        coEvery { api.toggleJob(any(), any()) } returns ApiResponse(success = true)

        val result = repository.toggleJob("1", false)

        assertTrue(result.isSuccess)
        coVerify { api.toggleJob("1", false) }
        coVerify { dao.toggle("1", false) }
    }

    @Test
    fun `deleteJob should call API and delete from dao`() = runTest {
        coEvery { api.deleteJob(any()) } returns ApiResponse(success = true)

        val result = repository.deleteJob("1")

        assertTrue(result.isSuccess)
        coVerify { api.deleteJob("1") }
        coVerify { dao.deleteById("1") }
    }

    @Test
    fun `generateCron should return expression from API`() = runTest {
        coEvery { api.generateCron(any()) } returns GenerateCronResponse(expression = "0 9 * * 1")

        val result = repository.generateCron("Every Monday at 9am", "wf-1")

        assertTrue(result.isSuccess)
        assertEquals("0 9 * * 1", result.getOrDefault(""))
    }

    @Test
    fun `testRun should return validity and next runs`() = runTest {
        val testResponse = TestCronResponse(valid = true, nextRuns = listOf("2026-04-02 09:00", "2026-04-03 09:00"))
        coEvery { api.testRun(any()) } returns testResponse

        val result = repository.testRun("0 9 * * *")

        assertTrue(result.isSuccess)
        val (isValid, nextRuns) = result.getOrDefault(Pair(false, emptyList()))
        assertTrue(isValid)
        assertEquals(2, nextRuns.size)
    }

    @Test
    fun `createJob with API failure should return failure`() = runTest {
        coEvery { api.createJob(any()) } throws RuntimeException("Server error")

        val result = repository.createJob("Test", "Desc", "0 * * * *", "wf-1")

        assertTrue(result.isFailure)
    }
}
