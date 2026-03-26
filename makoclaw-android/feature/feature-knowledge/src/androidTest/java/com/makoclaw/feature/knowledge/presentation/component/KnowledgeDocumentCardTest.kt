package com.makoclaw.feature.knowledge.presentation.component

import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithText
import com.makoclaw.core.model.KnowledgeChunk
import com.makoclaw.core.model.KnowledgeDocument
import org.junit.Rule
import org.junit.Test

class KnowledgeDocumentCardTest {

    @get:Rule
    val composeRule = createComposeRule()

    @Test
    fun cardShowsDocumentMetadata() {
        composeRule.setContent {
            KnowledgeDocumentCard(
                document = sampleDocument(),
                onClick = {},
                onLongClick = {}
            )
        }

        composeRule.onNodeWithText("Doc.pdf").assertIsDisplayed()
        composeRule.onNodeWithText("1 chunks").assertIsDisplayed()
    }
}

private fun sampleDocument() = KnowledgeDocument(
    id = "doc-1",
    title = "Doc.pdf",
    filename = "Doc.pdf",
    type = "pdf",
    chunks = listOf(KnowledgeChunk(id = "chunk-1", documentId = "doc-1", title = "Chunk", content = "text")),
    chunkCount = 1,
    createdAt = "1000"
)
