package com.makoclaw.feature.knowledge.presentation.component

import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithTag
import org.junit.Rule
import org.junit.Test

class UploadFileButtonTest {

    @get:Rule
    val composeRule = createComposeRule()

    @Test
    fun uploadButtonIsVisibleWhenNotUploading() {
        composeRule.setContent {
            UploadFileButton(onFileSelected = {})
        }

        composeRule.onNodeWithTag("knowledge_upload_button").assertIsDisplayed()
    }

    @Test
    fun progressIndicatorIsVisibleDuringUpload() {
        composeRule.setContent {
            UploadFileButton(onFileSelected = {}, uploadProgress = 0.5f)
        }

        composeRule.onNodeWithTag("knowledge_upload_progress").assertIsDisplayed()
    }
}
