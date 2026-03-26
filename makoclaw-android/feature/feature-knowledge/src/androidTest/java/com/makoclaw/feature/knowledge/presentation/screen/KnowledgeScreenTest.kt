package com.makoclaw.feature.knowledge.presentation.screen

import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.onNodeWithText
import com.makoclaw.core.model.KnowledgeChunk
import com.makoclaw.core.model.KnowledgeDocument
import kotlinx.coroutines.flow.MutableSharedFlow
import org.junit.Rule
import org.junit.Test

class KnowledgeScreenTest {

    @get:Rule
    val composeRule = createComposeRule()

    @Test
    fun emptyStateShowsUploadCallToAction() {
        composeRule.setContent {
            KnowledgeScreen(
                uiState = com.makoclaw.feature.knowledge.presentation.state.KnowledgeUiState(isEmpty = true),
                effects = MutableSharedFlow(),
                onRefresh = {},
                onSearchQueryChange = {},
                onClearSearch = {},
                onUploadFile = {},
                onSelectDocument = {},
                onDeleteDocument = {},
                onSaveChunk = { _, _, _ -> }
            )
        }

        composeRule.onNodeWithText("Subir documento").assertIsDisplayed()
    }

    @Test
    fun successStateShowsSearchBarAndDocument() {
        composeRule.setContent {
            KnowledgeScreen(
                uiState = com.makoclaw.feature.knowledge.presentation.state.KnowledgeUiState(
                    documents = listOf(sampleDocument()),
                    isEmpty = false
                ),
                effects = MutableSharedFlow(),
                onRefresh = {},
                onSearchQueryChange = {},
                onClearSearch = {},
                onUploadFile = {},
                onSelectDocument = {},
                onDeleteDocument = {},
                onSaveChunk = { _, _, _ -> }
            )
        }

        composeRule.onNodeWithTag("knowledge_search_bar").assertIsDisplayed()
        composeRule.onNodeWithText("Guide.md").assertIsDisplayed()
    }
}

private fun sampleDocument() = KnowledgeDocument(
    id = "doc-1",
    title = "Guide.md",
    filename = "Guide.md",
    type = "md",
    content = "# Guide",
    chunks = listOf(KnowledgeChunk(id = "chunk-1", documentId = "doc-1", title = "Intro", content = "hello")),
    chunkCount = 1,
    createdAt = "1000"
)
