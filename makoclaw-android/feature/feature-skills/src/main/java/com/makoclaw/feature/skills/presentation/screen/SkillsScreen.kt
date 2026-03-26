package com.makoclaw.feature.skills.presentation.screen

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Extension
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.makoclaw.core.ui.component.EmptyState

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SkillsScreen() {
    Scaffold(
        topBar = { TopAppBar(title = { Text("Skills & Marketplace") }) }
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding)) {
            EmptyState(
                title = "Skills Marketplace",
                message = "Browse, install, and manage skills for your agents",
                icon = Icons.Filled.Extension
            )
        }
    }
}
