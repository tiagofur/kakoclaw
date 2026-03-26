package com.makoclaw.core.datastore

import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import app.cash.turbine.test
import com.makoclaw.core.datastore.model.ExportFormat
import com.makoclaw.core.datastore.model.FeaturePreferences
import com.makoclaw.core.datastore.model.SortOrder
import com.makoclaw.core.datastore.model.TimeRange
import java.nio.file.Files
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test

class PreferencesStoreTest {
    private fun createStore(): PreferencesStore {
        val file = Files.createTempFile("feature-prefs", ".preferences_pb").toFile()
        val dataStore = PreferenceDataStoreFactory.create(produceFile = { file })
        return PreferencesStore(dataStore)
    }

    @Test
    fun `preferences exposes default values`() = runTest {
        val store = createStore()

        assertEquals(FeaturePreferences(), store.preferences.first())
    }

    @Test
    fun `updatePreferences persists all values`() = runTest {
        val store = createStore()
        val updated = FeaturePreferences(
            knowledgeUploadSize = 20,
            metricsTimeRange = TimeRange.MONTH,
            cronNotify = false,
            skillsAutoInstall = true,
            memoryRetentionDays = 90,
            historyExportFormat = ExportFormat.MARKDOWN,
            filesDefaultPath = "/tmp/makoclaw",
            filesSortOrder = SortOrder.SIZE,
            mcpReconnectTimeout = 45,
            reportTemplate = "executive",
            workflowAutoSave = false
        )

        store.updatePreferences(updated)

        assertEquals(updated, store.preferences.first())
    }

    @Test
    fun `clearAll resets to defaults`() = runTest {
        val store = createStore()
        store.updatePreferences(FeaturePreferences(reportTemplate = "custom", cronNotify = false))

        store.clearAll()

        assertEquals(FeaturePreferences(), store.preferences.first())
    }

    @Test
    fun `preferences flow emits mapped values`() = runTest {
        val store = createStore()

        store.preferences.test {
            assertEquals(FeaturePreferences(), awaitItem())

            val updated = FeaturePreferences(metricsTimeRange = TimeRange.TODAY, filesSortOrder = SortOrder.TYPE)
            store.updatePreferences(updated)

            assertEquals(updated, awaitItem())
            cancelAndIgnoreRemainingEvents()
        }
    }
}
