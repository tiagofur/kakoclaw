package com.makoclaw.feature.cron.presentation.screen

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Schedule
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
fun CronScreen() {
    Scaffold(
        topBar = { TopAppBar(title = { Text("Scheduled Tasks") }) },
        floatingActionButton = {
            ExtendedFloatingActionButton(
                onClick = { /* TODO */ },
                icon = { Icon(Icons.Filled.Add, contentDescription = null) },
                text = { Text("New Job") }
            )
        }
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding)) {
            EmptyState(
                title = "No scheduled tasks",
                message = "Create cron jobs to run tasks automatically",
                icon = Icons.Filled.Schedule
            )
        }
    }
}
