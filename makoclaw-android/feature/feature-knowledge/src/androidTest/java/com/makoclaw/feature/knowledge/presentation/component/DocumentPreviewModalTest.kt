package com.makoclaw.feature.knowledge.presentation.component

import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.onNodeWithText
import com.makoclaw.core.model.KnowledgeDocument
import org.junit.Rule
import org.junit.Test

class DocumentPreviewModalTest {

    @get:Rule
    val composeRule = createComposeRule()

    @Test
    fun previewModalDisplaysContent() {
        composeRule.setContent {
            DocumentPreviewModal(
                document = KnowledgeDocument(
                    id = "doc-1",
                    title = "Guide.md",
                    filename = "Guide.md",
                    type = "md",
                    content = "# Guide"
                ),
                onDismiss = {},
                visible = true
            )
        }

        composeRule.onNodeWithTag("knowledge_preview_modal").assertIsDisplayed()
        composeRule.onNodeWithText("Guide.md").assertIsDisplayed()
    }
}
