package com.makoclaw.feature.agents.data.repository

import com.makoclaw.core.model.AgentMetrics
import com.makoclaw.core.model.CreateSpecialistRequest
import com.makoclaw.core.model.OrchestratorConfig
import com.makoclaw.core.model.Specialist
import com.makoclaw.core.model.Swarm
import com.makoclaw.core.network.api.AgentApi
import com.makoclaw.core.network.api.AgentsResponse
import com.makoclaw.core.network.api.GenerateRequest
import com.makoclaw.core.network.api.SwarmsResponse
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk

class AgentsRepositoryImplTest {

    private lateinit var api: AgentApi
    private lateinit var repository: AgentsRepositoryImpl

    private val sampleSpecialists = listOf(
        Specialist(name = "coder", speciality = "code generation"),
        Specialist(name = "reviewer", speciality = "code review")
    )

    @BeforeEach
    fun setup() {
        api = mockk()
        repository = AgentsRepositoryImpl(api)
    }

    @Test
    fun `getSpecialists should return from API`() = runTest {
        coEvery { api.getAgents() } returns AgentsResponse(sampleSpecialists, OrchestratorConfig())

        val specialists = repository.getSpecialists().first()

        assertEquals(2, specialists.size)
        assertEquals("coder", specialists[0].name)
    }

    @Test
    fun `getOrchestrator should return config from API`() = runTest {
        val config = OrchestratorConfig(enabled = true, model = "gpt-4")
        coEvery { api.getAgents() } returns AgentsResponse(emptyList(), config)

        val result = repository.getOrchestrator()

        assertTrue(result.isSuccess)
        assertEquals(true, result.getOrDefault(OrchestratorConfig()).enabled)
        assertEquals("gpt-4", result.getOrDefault(OrchestratorConfig()).model)
    }

    @Test
    fun `getSwarms should return from API`() = runTest {
        val swarms = listOf(Swarm(name = "team-a", members = listOf("a", "b")))
        coEvery { api.getSwarms() } returns SwarmsResponse(swarms)

        val result = repository.getSwarms()

        assertTrue(result.isSuccess)
        assertEquals(1, result.getOrDefault(emptyList()).size)
    }

    @Test
    fun `getMetrics should return from API`() = runTest {
        val metrics = AgentMetrics(totalCalls = 42, totalTokens = 1000, totalCost = 5.0)
        coEvery { api.getMetrics() } returns metrics

        val result = repository.getMetrics()

        assertTrue(result.isSuccess)
        assertEquals(42, result.getOrDefault(AgentMetrics()).totalCalls)
    }

    @Test
    fun `createSpecialist should call API`() = runTest {
        val specialist = Specialist(name = "new", speciality = "test")
        coEvery { api.createSpecialist(any()) } returns specialist

        val result = repository.createSpecialist("new", "test", "prompt")

        assertTrue(result.isSuccess)
        coVerify { api.createSpecialist(any()) }
    }

    @Test
    fun `deleteSpecialist should call API`() = runTest {
        coEvery { api.deleteSpecialist(any()) } returns Unit

        val result = repository.deleteSpecialist("coder")

        assertTrue(result.isSuccess)
        coVerify { api.deleteSpecialist("coder") }
    }

    @Test
    fun `generateSpecialist should call API`() = runTest {
        val specialist = Specialist(name = "ai-agent", speciality = "generated")
        coEvery { api.generateSpecialist(any()) } returns specialist

        val result = repository.generateSpecialist("A coding specialist")

        assertTrue(result.isSuccess)
        assertEquals("ai-agent", result.getOrDefault(Specialist("")).name)
    }

    @Test
    fun `toggleOrchestrator should call API`() = runTest {
        coEvery { api.updateOrchestrator(any()) } returns Unit

        val config = OrchestratorConfig(enabled = true)
        val result = repository.toggleOrchestrator(config)

        assertTrue(result.isSuccess)
        coVerify { api.updateOrchestrator(config) }
    }

    @Test
    fun `createSpecialist with API failure should return failure`() = runTest {
        coEvery { api.createSpecialist(any()) } throws RuntimeException("Server error")

        val result = repository.createSpecialist("fail", "fail", "fail")

        assertTrue(result.isFailure)
    }
}
