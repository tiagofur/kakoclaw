package com.makoclaw.feature.cron.presentation.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Schedule
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExtendedFloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.ErrorScreen
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.cron.presentation.component.CronEditorModal
import com.makoclaw.feature.cron.presentation.component.CronJobCard
import com.makoclaw.feature.cron.presentation.state.CronEffect
import com.makoclaw.feature.cron.presentation.state.CronEvent
import com.makoclaw.feature.cron.presentation.viewmodel.CronViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CronScreen(viewModel: CronViewModel = hiltViewModel()) {
    val uiState by viewModel.uiState.collectAsState()
    val snackbarHostState = remember { SnackbarHostState() }
    var showEditor by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        viewModel.effects.collect { effect ->
            when (effect) {
                is CronEffect.ShowSnackbar ->
                    snackbarHostState.showSnackbar(effect.message)
                is CronEffect.JobCreated -> {
                    showEditor = false
                }
                is CronEffect.JobDeleted -> { /* already removed from list */ }
                is CronEffect.ShowTestResult -> {
                    val msg = if (effect.isValid) {
                        "Valid! Next: ${effect.nextRuns.firstOrNull() ?: "N/A"}"
                    } else {
                        "Invalid expression"
                    }
                    snackbarHostState.showSnackbar(msg)
                }
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(title = { Text("Scheduled Tasks") })
        },
        snackbarHost = { SnackbarHost(snackbarHostState) },
        floatingActionButton = {
            ExtendedFloatingActionButton(
                onClick = { showEditor = true },
                icon = { Icon(Icons.Filled.Add, contentDescription = null) },
                text = { Text("New Job") }
            )
        }
    ) { padding ->
        when {
            uiState.isLoading -> {
                LoadingScreen(modifier = Modifier.padding(padding))
            }
            uiState.error != null && uiState.jobs.isEmpty() -> {
                ErrorScreen(
                    message = uiState.error ?: "Unknown error",
                    modifier = Modifier.padding(padding),
                    onRetry = { viewModel.onEvent(CronEvent.Refresh) }
                )
            }
            uiState.isEmpty -> {
                EmptyState(
                    title = "No Scheduled Tasks",
                    message = "Create cron jobs to run tasks automatically.",
                    icon = Icons.Filled.Schedule,
                    modifier = Modifier.padding(padding)
                )
            }
            else -> {
                LazyColumn(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding)
                        .padding(horizontal = 16.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    item { Spacer(modifier = Modifier.height(4.dp)) }

                    items(uiState.jobs, key = { it.id }) { job ->
                        CronJobCard(
                            job = job,
                            onToggle = { enabled ->
                                viewModel.onEvent(CronEvent.ToggleJob(job.id, enabled))
                            },
                            onDelete = {
                                viewModel.onEvent(CronEvent.DeleteJob(job.id))
                            },
                            onTestRun = {
                                viewModel.onEvent(CronEvent.TestRun(job.cronExpression))
                            },
                            onClick = {
                                viewModel.onEvent(CronEvent.SelectJob(job.id))
                            }
                        )
                    }

                    item { Spacer(modifier = Modifier.height(80.dp)) }
                }
            }
        }
    }

    if (showEditor) {
        CronEditorModal(
            onDismiss = {
                showEditor = false
                viewModel.onEvent(CronEvent.ClearSelection)
            },
            onCreateJob = { name, description, cronExpression ->
                viewModel.onEvent(
                    CronEvent.CreateJob(
                        name = name,
                        description = description,
                        cronExpression = cronExpression,
                        workflowId = ""
                    )
                )
            },
            onGenerateCron = { prompt ->
                viewModel.onEvent(CronEvent.GenerateCron(prompt, ""))
            },
            generatedExpression = uiState.generatedExpression,
            isValid = uiState.testResult?.isValid,
            nextRuns = uiState.testResult?.nextRuns ?: emptyList(),
            onTestRun = { expression ->
                viewModel.onEvent(CronEvent.TestRun(expression))
            },
            isCreating = uiState.isCreating
        )
    }
}
