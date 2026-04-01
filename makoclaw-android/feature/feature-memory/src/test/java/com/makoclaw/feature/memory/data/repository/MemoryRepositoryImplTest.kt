package com.makoclaw.feature.memory.data.repository

import com.makoclaw.core.database.dao.MemoryDao
import com.makoclaw.core.database.entity.MemoryEntity
import com.makoclaw.core.model.DailyNote
import com.makoclaw.core.model.MemoryContent
import com.makoclaw.core.network.api.MemoryApi
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

class MemoryRepositoryImplTest {

    private lateinit var api: MemoryApi
    private lateinit var dao: MemoryDao
    private lateinit var repository: MemoryRepositoryImpl

    private val sampleEntities = listOf(
        MemoryEntity(id = "long-term", type = "long-term", content = "User prefers dark mode", timestamp = 1000L),
        MemoryEntity(id = "daily-1", type = "daily-note", content = "Worked on skills feature", timestamp = 2000L)
    )

    @BeforeEach
    fun setup() {
        api = mockk()
        dao = mockk(relaxed = true)
        repository = MemoryRepositoryImpl(api, dao)
    }

    @Test
    fun `getLongTermMemory should return content from dao`() = runTest {
        every { dao.getAll() } returns flowOf(sampleEntities)
        coEvery { api.getLongTermMemory() } returns MemoryContent("API content")

        val content = repository.getLongTermMemory().first()

        assertEquals("User prefers dark mode", content)
    }

    @Test
    fun `getLongTermMemory should refresh from API`() = runTest {
        every { dao.getAll() } returns flowOf(emptyList())
        coEvery { api.getLongTermMemory() } returns MemoryContent("API memory content")

        repository.getLongTermMemory().first()

        coVerify { api.getLongTermMemory() }
    }

    @Test
    fun `saveLongTermMemory should call API and update dao`() = runTest {
        coEvery { api.updateLongTermMemory(any()) } returns Unit

        val result = repository.saveLongTermMemory("New memory content")

        assertTrue(result.isSuccess)
        coVerify { api.updateLongTermMemory(any()) }
        coVerify { dao.insert(any()) }
    }

    @Test
    fun `getDailyNotes should return from API`() = runTest {
        val notes = listOf(
            DailyNote(date = "2026-04-01", content = "Test note")
        )
        coEvery { api.getDailyNotes(any()) } returns notes

        val result = repository.getDailyNotes(7)

        assertTrue(result.isSuccess)
        assertEquals(1, result.getOrDefault(emptyList()).size)
    }

    @Test
    fun `getDailyNotes with API failure should return failure`() = runTest {
        coEvery { api.getDailyNotes(any()) } throws RuntimeException("Server error")

        val result = repository.getDailyNotes(7)

        assertTrue(result.isFailure)
    }

    @Test
    fun `createDailyNote should insert into dao`() = runTest {
        val result = repository.createDailyNote("New observation")

        assertTrue(result.isSuccess)
        coVerify { dao.insert(any()) }
    }

    @Test
    fun `saveLongTermMemory with API failure should return failure`() = runTest {
        coEvery { api.updateLongTermMemory(any()) } throws RuntimeException("Save failed")

        val result = repository.saveLongTermMemory("fail")

        assertTrue(result.isFailure)
    }
}
