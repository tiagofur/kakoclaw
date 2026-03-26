package com.makoclaw.feature.knowledge.presentation.screen

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.LibraryBooks
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExtendedFloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.makoclaw.core.ui.component.EmptyState

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun KnowledgeScreen() {
    Scaffold(
        topBar = { TopAppBar(title = { Text("Knowledge Base") }) },
        floatingActionButton = {
            ExtendedFloatingActionButton(
                onClick = { /* TODO: upload document */ },
                icon = { Icon(Icons.Filled.Add, contentDescription = null) },
                text = { Text("Upload") }
            )
        }
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding)) {
            EmptyState(
                title = "No documents yet",
                message = "Upload documents for RAG-powered search",
                icon = Icons.Filled.LibraryBooks
            )
        }
    }
}
