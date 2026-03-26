package com.makoclaw.feature.workflows.presentation.viewmodel

import app.cash.turbine.test
import com.makoclaw.core.model.Workflow
import com.makoclaw.core.model.WorkflowExecutionLog
import com.makoclaw.core.network.api.ApiResponse
import com.makoclaw.feature.workflows.data.repository.WorkflowsRepository
import com.makoclaw.feature.workflows.presentation.state.WorkflowsEffect
import com.makoclaw.feature.workflows.presentation.state.WorkflowsEvent
import io.mockk.every
import io.mockk.mockk
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertFalse
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

@OptIn(ExperimentalCoroutinesApi::class)
class WorkflowsViewModelTest {
    private val dispatcher = StandardTestDispatcher()
    private lateinit var repository: WorkflowsRepository

    @BeforeEach
    fun setUp() {
        Dispatchers.setMain(dispatcher)
        repository = mockk(relaxed = true)
        every { repository.getWorkflows() } returns flowOf(listOf(sampleWorkflow()))
        every { repository.deleteWorkflow(any()) } returns flowOf(ApiResponse(message = "deleted"))
        every { repository.executeWorkflow(any()) } returns flowOf(sampleLog())
        every { repository.getExecutionLogs(any()) } returns flowOf(listOf(sampleLog()))
        every { repository.createWorkflow(any()) } returns flowOf(ApiResponse(message = "created"))
        every { repository.updateWorkflow(any()) } returns flowOf(ApiResponse(message = "updated"))
    }

    @AfterEach
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `init loads workflows`() = runTest(dispatcher) {
        val viewModel = WorkflowsViewModel(repository)
        advanceUntilIdle()
        assertFalse(viewModel.uiState.value.isLoading)
        assertEquals(1, viewModel.uiState.value.workflows.size)
    }

    @Test
    fun `view logs emits navigation effect`() = runTest(dispatcher) {
        val viewModel = WorkflowsViewModel(repository)
        advanceUntilIdle()

        viewModel.effects.test {
            viewModel.onEvent(WorkflowsEvent.ViewLogs("wf-1"))
            advanceUntilIdle()
            assertEquals(WorkflowsEffect.NavigateToLogs("wf-1"), awaitItem())
        }
    }
}

private fun sampleWorkflow() = Workflow(
    id = "wf-1",
    name = "Daily sync",
    description = "Sync data",
    createdAt = 1L,
    updatedAt = 2L
)

private fun sampleLog() = WorkflowExecutionLog(
    id = "log-1",
    workflowId = "wf-1",
    status = "success",
    message = "done",
    timestamp = 10L
)
