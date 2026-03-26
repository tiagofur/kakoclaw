package com.makoclaw.feature.workflows.presentation.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarDuration
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.workflows.presentation.component.ExecutionLogsModal
import com.makoclaw.feature.workflows.presentation.component.WorkflowCard
import com.makoclaw.feature.workflows.presentation.state.WorkflowsEffect
import com.makoclaw.feature.workflows.presentation.state.WorkflowsEvent
import com.makoclaw.feature.workflows.presentation.viewmodel.WorkflowsViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun WorkflowsScreen(viewModel: WorkflowsViewModel = hiltViewModel()) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val snackbarHostState = remember { SnackbarHostState() }
    var showLogs by remember { mutableStateOf(false) }
    var logsWorkflowId by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(viewModel) {
        viewModel.effects.collect { effect ->
            when (effect) {
                is WorkflowsEffect.ShowSnackbar -> snackbarHostState.showSnackbar(
                    message = effect.message,
                    duration = SnackbarDuration.Short
                )
                is WorkflowsEffect.NavigateToLogs -> {
                    showLogs = true
                    logsWorkflowId = effect.workflowId
                }
                else -> Unit
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Workflows") },
                actions = {
                    IconButton(onClick = { viewModel.onEvent(WorkflowsEvent.Refresh) }) {
                        Icon(Icons.Default.Refresh, contentDescription = "Refresh")
                    }
                }
            )
        },
        snackbarHost = { SnackbarHost(snackbarHostState) },
        floatingActionButton = {
            FloatingActionButton(onClick = { viewModel.onEvent(WorkflowsEvent.OpenEditor) }) {
                Icon(Icons.Default.Add, contentDescription = "Add workflow")
            }
        }
    ) { padding ->
        when {
            uiState.isEditorMode && uiState.editorWorkflow != null -> WorkflowCanvasEditor(
                workflow = uiState.editorWorkflow,
                onSave = { viewModel.onEvent(WorkflowsEvent.SaveEditor(it)) },
                onCancel = { viewModel.onEvent(WorkflowsEvent.CloseEditor) }
            )

            uiState.isLoading -> LoadingScreen(modifier = Modifier.padding(padding))

            uiState.error != null -> Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center
            ) {
                Text(uiState.error ?: "Unknown error", color = MaterialTheme.colorScheme.error)
            }

            uiState.isEmpty -> EmptyState(
                title = "No workflows yet",
                message = "Create your first workflow to automate tasks",
                modifier = Modifier.padding(padding)
            )

            else -> LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding)
                    .padding(horizontal = 16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                item { Spacer(modifier = Modifier.height(8.dp)) }
                items(uiState.workflows, key = { it.id.ifBlank { it.name } }) { workflow ->
                    WorkflowCard(
                        workflow = workflow,
                        onClick = { viewModel.onEvent(WorkflowsEvent.SelectWorkflow(workflow.id)) },
                        onExecute = { viewModel.onEvent(WorkflowsEvent.ExecuteWorkflow(workflow.id)) },
                        onEdit = {
                            viewModel.onEvent(WorkflowsEvent.SelectWorkflow(workflow.id))
                            viewModel.onEvent(WorkflowsEvent.OpenEditor)
                        },
                        onDelete = { viewModel.onEvent(WorkflowsEvent.DeleteWorkflow(workflow.id)) },
                        onViewLogs = { viewModel.onEvent(WorkflowsEvent.ViewLogs(workflow.id)) }
                    )
                }
                item { Spacer(modifier = Modifier.height(80.dp)) }
            }
        }
    }

    if (showLogs && logsWorkflowId != null) {
        ExecutionLogsModal(
            workflowId = logsWorkflowId.orEmpty(),
            logs = uiState.executionLogs,
            onDismiss = { showLogs = false }
        )
    }
}
