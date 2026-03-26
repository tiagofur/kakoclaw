package com.makoclaw.feature.agents.presentation.screen

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.SmartToy
import androidx.compose.material3.AlertDialog
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
import androidx.compose.material3.SuggestionChip
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
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
import com.makoclaw.core.model.Specialist
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.agents.presentation.viewmodel.AgentsViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AgentsScreen(
    viewModel: AgentsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    var showCreateSheet by remember { mutableStateOf(false) }
    var showGenerateDialog by remember { mutableStateOf(false) }

    Scaffold(
        topBar = { TopAppBar(title = { Text("Agents") }) },
        floatingActionButton = {
            ExtendedFloatingActionButton(
                onClick = { showCreateSheet = true },
                icon = { Icon(Icons.Filled.Add, contentDescription = null) },
                text = { Text("New Specialist") }
            )
        }
    ) { padding ->
        if (uiState.isLoading) {
            LoadingScreen(modifier = Modifier.padding(padding))
            return@Scaffold
        }

        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            // Orchestrator card
            item {
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(
                        containerColor = MaterialTheme.colorScheme.secondaryContainer
                    )
                ) {
                    Row(
                        modifier = Modifier.padding(16.dp).fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Column {
                            Text("Orchestrator", style = MaterialTheme.typography.titleMedium)
                            Text(
                                "Multi-agent coordination",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSecondaryContainer.copy(alpha = 0.7f)
                            )
                        }
                        Switch(
                            checked = uiState.orchestrator.enabled,
                            onCheckedChange = viewModel::toggleOrchestrator
                        )
                    }
                }
            }

            // AI Generate
            item {
                Card(
                    onClick = { showGenerateDialog = true },
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                        Icon(Icons.Filled.AutoAwesome, null, tint = MaterialTheme.colorScheme.primary)
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text("Generate with AI", style = MaterialTheme.typography.titleSmall)
                            Text("Describe a specialist and AI will create it",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                    }
                }
            }

            // Metrics summary
            item {
                AnimatedVisibility(visible = uiState.metrics.totalCalls > 0) {
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        colors = CardDefaults.cardColors(
                            containerColor = MaterialTheme.colorScheme.tertiaryContainer.copy(alpha = 0.5f)
                        )
                    ) {
                        Row(
                            modifier = Modifier.padding(16.dp).fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceEvenly
                        ) {
                            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                Text(uiState.metrics.totalCalls.toString(), style = MaterialTheme.typography.titleLarge)
                                Text("Calls", style = MaterialTheme.typography.labelSmall)
                            }
                            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                Text(uiState.metrics.totalTokens.toString(), style = MaterialTheme.typography.titleLarge)
                                Text("Tokens", style = MaterialTheme.typography.labelSmall)
                            }
                            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                Text("$${String.format("%.2f", uiState.metrics.totalCost)}", style = MaterialTheme.typography.titleLarge)
                                Text("Cost", style = MaterialTheme.typography.labelSmall)
                            }
                        }
                    }
                }
            }

            // Section header
            item {
                Text(
                    "Specialists (${uiState.specialists.size})",
                    style = MaterialTheme.typography.titleMedium,
                    modifier = Modifier.padding(top = 8.dp, bottom = 4.dp)
                )
            }

            // Specialists list
            if (uiState.specialists.isEmpty()) {
                item {
                    EmptyState(
                        title = "No specialists configured",
                        message = "Create a specialist or generate one with AI",
                        icon = Icons.Filled.SmartToy
                    )
                }
            } else {
                items(uiState.specialists, key = { it.name }) { specialist ->
                    SpecialistCard(
                        specialist = specialist,
                        onDelete = { viewModel.deleteSpecialist(specialist.name) }
                    )
                }
            }

            // Swarms section
            if (uiState.swarms.isNotEmpty()) {
                item {
                    Text(
                        "Swarms (${uiState.swarms.size})",
                        style = MaterialTheme.typography.titleMedium,
                        modifier = Modifier.padding(top = 16.dp, bottom = 4.dp)
                    )
                }
                items(uiState.swarms, key = { it.name }) { swarm ->
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainer)
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Text(swarm.name, style = MaterialTheme.typography.titleSmall)
                            Text("${swarm.members.size} members · ${swarm.mode}",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                    }
                }
            }

            item { Spacer(modifier = Modifier.height(80.dp)) }
        }
    }

    // Create specialist sheet
    if (showCreateSheet) {
        CreateSpecialistSheet(
            onDismiss = { showCreateSheet = false },
            onCreate = { name, speciality, prompt ->
                viewModel.createSpecialist(name, speciality, prompt)
                showCreateSheet = false
            }
        )
    }

    // AI Generate dialog
    if (showGenerateDialog) {
        var description by remember { mutableStateOf("") }
        AlertDialog(
            onDismissRequest = { showGenerateDialog = false },
            title = { Text("Generate Specialist with AI") },
            text = {
                OutlinedTextField(
                    value = description,
                    onValueChange = { description = it },
                    label = { Text("Describe the specialist") },
                    modifier = Modifier.fillMaxWidth(),
                    minLines = 3
                )
            },
            confirmButton = {
                TextButton(
                    onClick = {
                        viewModel.generateSpecialist(description)
                        showGenerateDialog = false
                    },
                    enabled = description.isNotBlank() && !uiState.isGenerating
                ) { Text(if (uiState.isGenerating) "Generating..." else "Generate") }
            },
            dismissButton = {
                TextButton(onClick = { showGenerateDialog = false }) { Text("Cancel") }
            }
        )
    }
}

@Composable
fun SpecialistCard(specialist: Specialist, onDelete: () -> Unit) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainer)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(Icons.Filled.SmartToy, null, modifier = Modifier.size(32.dp), tint = MaterialTheme.colorScheme.primary)
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(specialist.name, style = MaterialTheme.typography.titleSmall)
                if (specialist.speciality.isNotEmpty()) {
                    Text(specialist.speciality, style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
                if (specialist.tools.isNotEmpty()) {
                    Row(
                        modifier = Modifier.padding(top = 4.dp),
                        horizontalArrangement = Arrangement.spacedBy(4.dp)
                    ) {
                        specialist.tools.take(3).forEach { tool ->
                            SuggestionChip(
                                onClick = {},
                                label = { Text(tool, style = MaterialTheme.typography.labelSmall) }
                            )
                        }
                        if (specialist.tools.size > 3) {
                            Text("+${specialist.tools.size - 3}", style = MaterialTheme.typography.labelSmall,
                                modifier = Modifier.align(Alignment.CenterVertically))
                        }
                    }
                }
            }
            IconButton(onClick = onDelete) {
                Icon(Icons.Filled.Delete, "Delete", tint = MaterialTheme.colorScheme.error)
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CreateSpecialistSheet(
    onDismiss: () -> Unit,
    onCreate: (String, String, String) -> Unit
) {
    var name by remember { mutableStateOf("") }
    var speciality by remember { mutableStateOf("") }
    var systemPrompt by remember { mutableStateOf("") }

    ModalBottomSheet(onDismissRequest = onDismiss, sheetState = rememberModalBottomSheetState()) {
        Column(modifier = Modifier.fillMaxWidth().padding(24.dp)) {
            Text("New Specialist", style = MaterialTheme.typography.headlineSmall)
            Spacer(modifier = Modifier.height(16.dp))
            OutlinedTextField(value = name, onValueChange = { name = it }, label = { Text("Name") }, modifier = Modifier.fillMaxWidth(), singleLine = true)
            Spacer(modifier = Modifier.height(12.dp))
            OutlinedTextField(value = speciality, onValueChange = { speciality = it }, label = { Text("Speciality") }, modifier = Modifier.fillMaxWidth(), singleLine = true)
            Spacer(modifier = Modifier.height(12.dp))
            OutlinedTextField(value = systemPrompt, onValueChange = { systemPrompt = it }, label = { Text("System Prompt") }, modifier = Modifier.fillMaxWidth(), minLines = 3, maxLines = 6)
            Spacer(modifier = Modifier.height(24.dp))
            androidx.compose.material3.Button(
                onClick = { onCreate(name, speciality, systemPrompt) },
                modifier = Modifier.fillMaxWidth(),
                enabled = name.isNotBlank()
            ) { Text("Create Specialist") }
            Spacer(modifier = Modifier.height(16.dp))
        }
    }
}
