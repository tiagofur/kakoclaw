package com.makoclaw.feature.knowledge.presentation.component

import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.onNodeWithText
import com.makoclaw.core.model.KnowledgeChunk
import com.makoclaw.core.model.KnowledgeDocument
import org.junit.Rule
import org.junit.Test

class ChunkEditorSheetTest {

    @get:Rule
    val composeRule = createComposeRule()

    @Test
    fun chunkEditorShowsChunkFields() {
        composeRule.setContent {
            ChunkEditorSheet(
                document = KnowledgeDocument(
                    id = "doc-1",
                    title = "Guide.md",
                    chunks = listOf(
                        KnowledgeChunk(
                            id = "chunk-1",
                            documentId = "doc-1",
                            title = "Intro",
                            content = "hello"
                        )
                    )
                ),
                visible = true,
                onDismiss = {},
                onSaveChunk = { _, _ -> }
            )
        }

        composeRule.onNodeWithTag("knowledge_chunk_editor").assertIsDisplayed()
        composeRule.onNodeWithText("Intro").assertIsDisplayed()
    }
}
