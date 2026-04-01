package com.makoclaw.feature.memory.presentation.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.NoteAlt
import androidx.compose.material.icons.filled.Save
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExtendedFloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.model.DailyNote
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.ErrorScreen
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.memory.presentation.state.MemoryEffect
import com.makoclaw.feature.memory.presentation.state.MemoryEvent
import com.makoclaw.feature.memory.presentation.viewmodel.MemoryViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MemoryScreen(viewModel: MemoryViewModel = hiltViewModel()) {
    val uiState by viewModel.uiState.collectAsState()
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(Unit) {
        viewModel.effects.collect { effect ->
            when (effect) {
                is MemoryEffect.ShowSnackbar -> snackbarHostState.showSnackbar(effect.message)
                MemoryEffect.MemorySaved -> { /* state updates via Flow */ }
                is MemoryEffect.DailyNoteCreated -> { /* list refreshes */ }
            }
        }
    }

    Scaffold(
        topBar = { TopAppBar(title = { Text("Memory") }) },
        snackbarHost = { SnackbarHost(snackbarHostState) },
        floatingActionButton = {
            if (uiState.selectedTab == 0 && uiState.hasChanges) {
                ExtendedFloatingActionButton(
                    onClick = { viewModel.onEvent(MemoryEvent.SaveMemory) },
                    icon = { Icon(Icons.Filled.Save, contentDescription = null) },
                    text = { Text(if (uiState.isSaving) "Saving..." else "Save") }
                )
            }
        }
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding)) {
            TabRow(selectedTabIndex = uiState.selectedTab) {
                Tab(selected = uiState.selectedTab == 0, onClick = { viewModel.onEvent(MemoryEvent.SelectTab(0)) }, text = { Text("Long-Term") })
                Tab(selected = uiState.selectedTab == 1, onClick = { viewModel.onEvent(MemoryEvent.SelectTab(1)) }, text = { Text("Daily Notes") })
            }

            when {
                uiState.isLoading -> LoadingScreen()
                uiState.error != null && uiState.longTermMemory.isEmpty() -> ErrorScreen(
                    message = uiState.error ?: "Unknown error",
                    onRetry = { viewModel.onEvent(MemoryEvent.Refresh) }
                )
                else -> when (uiState.selectedTab) {
                    0 -> LongTermMemoryTab(
                        memory = uiState.longTermMemory,
                        hasChanges = uiState.hasChanges,
                        isSaving = uiState.isSaving,
                        onTextChange = { viewModel.onEvent(MemoryEvent.UpdateMemoryText(it)) }
                    )
                    1 -> DailyNotesTab(
                        notes = uiState.dailyNotes,
                        onCreateNote = { viewModel.onEvent(MemoryEvent.CreateDailyNote(it)) }
                    )
                }
            }
        }
    }
}

@Composable
private fun LongTermMemoryTab(
    memory: String,
    hasChanges: Boolean,
    isSaving: Boolean,
    onTextChange: (String) -> Unit
) {
    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        OutlinedTextField(
            value = memory,
            onValueChange = onTextChange,
            label = { Text("Persistent memory context") },
            modifier = Modifier.fillMaxWidth().weight(1f),
            minLines = 10,
            enabled = !isSaving
        )
        if (hasChanges && !isSaving) {
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                "Unsaved changes",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.error
            )
        }
    }
}

@Composable
private fun DailyNotesTab(
    notes: List<DailyNote>,
    onCreateNote: (String) -> Unit
) {
    var newNoteText by remember { mutableStateOf("") }
    var showNewNoteForm by remember { mutableStateOf(false) }

    Column(modifier = Modifier.fillMaxSize()) {
        if (notes.isEmpty() && !showNewNoteForm) {
            EmptyState(
                title = "No Daily Notes",
                message = "Create observations and insights for each day.",
                icon = Icons.Filled.NoteAlt,
                modifier = Modifier.weight(1f)
            )
        } else {
            LazyColumn(modifier = Modifier.weight(1f).padding(horizontal = 16.dp)) {
                item { Spacer(modifier = Modifier.height(8.dp)) }
                items(notes, key = { it.date + it.content }) { note ->
                    Card(
                        modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
                        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainer)
                    ) {
                        Column(modifier = Modifier.padding(16.dp)) {
                            Text(note.date, style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.primary)
                            Text(note.content, style = MaterialTheme.typography.bodySmall, modifier = Modifier.padding(top = 4.dp))
                        }
                    }
                }
                item { Spacer(modifier = Modifier.height(8.dp)) }
            }
        }

        Column(modifier = Modifier.padding(16.dp)) {
            if (showNewNoteForm) {
                OutlinedTextField(
                    value = newNoteText,
                    onValueChange = { newNoteText = it },
                    label = { Text("New observation") },
                    modifier = Modifier.fillMaxWidth(),
                    minLines = 2,
                    maxLines = 4
                )
                Spacer(modifier = Modifier.height(8.dp))
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    OutlinedButton(
                        onClick = { showNewNoteForm = false; newNoteText = "" },
                        modifier = Modifier.weight(1f)
                    ) { Text("Cancel") }
                    Button(
                        onClick = {
                            onCreateNote(newNoteText)
                            newNoteText = ""
                            showNewNoteForm = false
                        },
                        modifier = Modifier.weight(1f),
                        enabled = newNoteText.isNotBlank()
                    ) { Text("Save") }
                }
            } else {
                OutlinedButton(
                    onClick = { showNewNoteForm = true },
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Icon(Icons.Filled.Add, contentDescription = null, modifier = Modifier.size(18.dp))
                    Spacer(modifier = Modifier.size(8.dp))
                    Text("Add Daily Note")
                }
            }
        }
    }
}
