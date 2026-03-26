package com.makoclaw.feature.cron.presentation.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Schedule
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExtendedFloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.cron.presentation.viewmodel.CronViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CronScreen(viewModel: CronViewModel = hiltViewModel()) {
    val uiState by viewModel.uiState.collectAsState()
    var showCreateSheet by remember { mutableStateOf(false) }

    Scaffold(
        topBar = { TopAppBar(title = { Text("Scheduled Tasks") }) },
        floatingActionButton = {
            ExtendedFloatingActionButton(
                onClick = { showCreateSheet = true },
                icon = { Icon(Icons.Filled.Add, contentDescription = null) },
                text = { Text("New Job") }
            )
        }
    ) { padding ->
        when {
            uiState.isLoading -> LoadingScreen(modifier = Modifier.padding(padding))
            uiState.jobs.isEmpty() -> Column(modifier = Modifier.fillMaxSize().padding(padding)) {
                EmptyState(title = "No scheduled tasks", message = "Create cron jobs to run tasks automatically", icon = Icons.Filled.Schedule)
            }
            else -> LazyColumn(
                modifier = Modifier.fillMaxSize().padding(padding).padding(horizontal = 16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                item { Spacer(modifier = Modifier.height(4.dp)) }
                items(uiState.jobs, key = { it.id }) { job ->
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainer)
                    ) {
                        Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                            Column(modifier = Modifier.weight(1f)) {
                                Text(job.name.ifEmpty { job.message.take(40) }, style = MaterialTheme.typography.titleSmall)
                                Text(job.schedule, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.primary)
                                if (job.nextRun.isNotEmpty()) {
                                    Text("Next: ${job.nextRun}", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                                }
                            }
                            Switch(checked = job.enabled, onCheckedChange = { viewModel.toggleJob(job.id, it) })
                            IconButton(onClick = { viewModel.deleteJob(job.id) }) {
                                Icon(Icons.Filled.Delete, "Delete", tint = MaterialTheme.colorScheme.error)
                            }
                        }
                    }
                }
                item { Spacer(modifier = Modifier.height(80.dp)) }
            }
        }
    }

    if (showCreateSheet) {
        var name by remember { mutableStateOf("") }
        var schedule by remember { mutableStateOf("") }
        var message by remember { mutableStateOf("") }
        ModalBottomSheet(onDismissRequest = { showCreateSheet = false }, sheetState = rememberModalBottomSheetState()) {
            Column(modifier = Modifier.fillMaxWidth().padding(24.dp)) {
                Text("New Cron Job", style = MaterialTheme.typography.headlineSmall)
                Spacer(modifier = Modifier.height(16.dp))
                OutlinedTextField(value = name, onValueChange = { name = it }, label = { Text("Name") }, modifier = Modifier.fillMaxWidth(), singleLine = true)
                Spacer(modifier = Modifier.height(12.dp))
                OutlinedTextField(value = schedule, onValueChange = { schedule = it }, label = { Text("Schedule (cron expression)") }, modifier = Modifier.fillMaxWidth(), singleLine = true, placeholder = { Text("0 * * * *") })
                Spacer(modifier = Modifier.height(12.dp))
                OutlinedTextField(value = message, onValueChange = { message = it }, label = { Text("Message / Command") }, modifier = Modifier.fillMaxWidth(), minLines = 2)
                Spacer(modifier = Modifier.height(24.dp))
                androidx.compose.material3.Button(
                    onClick = { viewModel.createJob(name, schedule, message); showCreateSheet = false },
                    modifier = Modifier.fillMaxWidth(),
                    enabled = schedule.isNotBlank() && message.isNotBlank()
                ) { Text("Create Job") }
                Spacer(modifier = Modifier.height(16.dp))
            }
        }
    }
}
