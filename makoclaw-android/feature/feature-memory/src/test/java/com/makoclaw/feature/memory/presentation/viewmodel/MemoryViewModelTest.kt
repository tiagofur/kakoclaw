package com.makoclaw.feature.memory.presentation.viewmodel

import com.makoclaw.core.model.DailyNote
import com.makoclaw.feature.memory.data.repository.MemoryRepository
import com.makoclaw.feature.memory.presentation.state.MemoryEvent
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

class MemoryViewModelTest {

    private val testDispatcher = StandardTestDispatcher()
    private lateinit var repository: MemoryRepository
    private lateinit var viewModel: MemoryViewModel

    private val sampleNotes = listOf(
        DailyNote(date = "2026-04-01", content = "Completed BATCH 4 implementation"),
        DailyNote(date = "2026-03-31", content = "Started BATCH 3C")
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

    private fun createViewModel(): MemoryViewModel {
        every { repository.getLongTermMemory() } returns flow { emit("Sample memory content") }
        coEvery { repository.getDailyNotes(any()) } returns Result.success(sampleNotes)
        viewModel = MemoryViewModel(repository, testDispatcher)
        return viewModel
    }

    @Test
    fun `initial state should be loading`() {
        every { repository.getLongTermMemory() } returns flow { emit("test") }
        coEvery { repository.getDailyNotes(any()) } returns Result.success(emptyList())
        viewModel = MemoryViewModel(repository, testDispatcher)
        assertTrue(viewModel.uiState.value.isLoading)
    }

    @Test
    fun `load memory should populate long term memory`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.isLoading)
        assertEquals("Sample memory content", state.longTermMemory)
        assertNull(state.error)
    }

    @Test
    fun `load memory should populate daily notes`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals(2, state.dailyNotes.size)
    }

    @Test
    fun `load memory with error should show error`() = runTest {
        every { repository.getLongTermMemory() } returns flow {
            throw RuntimeException("API failure")
        }
        coEvery { repository.getDailyNotes(any()) } returns Result.success(emptyList())
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.isLoading)
        assertNotNull(state.error)
    }

    @Test
    fun `update memory text should mark has changes`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(MemoryEvent.UpdateMemoryText("Modified content"))

        assertTrue(viewModel.uiState.value.hasChanges)
        assertEquals("Modified content", viewModel.uiState.value.longTermMemory)
    }

    @Test
    fun `save memory should call repository and clear has changes`() = runTest {
        coEvery { repository.saveLongTermMemory(any()) } returns Result.success(Unit)
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(MemoryEvent.UpdateMemoryText("New content"))
        viewModel.onEvent(MemoryEvent.SaveMemory)
        testDispatcher.scheduler.advanceUntilIdle()

        assertFalse(viewModel.uiState.value.hasChanges)
        assertFalse(viewModel.uiState.value.isSaving)
        coEvery { repository.saveLongTermMemory("New content") }
    }

    @Test
    fun `save memory with failure should not clear has changes`() = runTest {
        coEvery { repository.saveLongTermMemory(any()) } returns Result.failure(RuntimeException("Save error"))
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(MemoryEvent.UpdateMemoryText("New content"))
        viewModel.onEvent(MemoryEvent.SaveMemory)
        testDispatcher.scheduler.advanceUntilIdle()

        assertTrue(viewModel.uiState.value.hasChanges)
        assertFalse(viewModel.uiState.value.isSaving)
    }

    @Test
    fun `create daily note should call repository`() = runTest {
        coEvery { repository.createDailyNote(any()) } returns Result.success(
            DailyNote(date = "2026-04-01", content = "Test note")
        )
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(MemoryEvent.CreateDailyNote("Test note"))
        testDispatcher.scheduler.advanceUntilIdle()

        verify { repository.createDailyNote("Test note") }
    }

    @Test
    fun `select tab should update selected tab`() = runTest {
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        viewModel.onEvent(MemoryEvent.SelectTab(1))
        assertEquals(1, viewModel.uiState.value.selectedTab)
    }

    @Test
    fun `clear error should remove error`() = runTest {
        every { repository.getLongTermMemory() } returns flow {
            throw RuntimeException("Error")
        }
        coEvery { repository.getDailyNotes(any()) } returns Result.success(emptyList())
        createViewModel()
        testDispatcher.scheduler.advanceUntilIdle()

        assertNotNull(viewModel.uiState.value.error)

        viewModel.onEvent(MemoryEvent.ClearError)
        assertNull(viewModel.uiState.value.error)
    }
}
