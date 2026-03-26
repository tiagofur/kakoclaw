package com.makoclaw.feature.memory.presentation.screen

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Save
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExtendedFloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.memory.presentation.viewmodel.MemoryViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MemoryScreen(
    viewModel: MemoryViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    var selectedTab by remember { mutableIntStateOf(0) }

    Scaffold(
        topBar = { TopAppBar(title = { Text("Memory") }) },
        floatingActionButton = {
            if (selectedTab == 0 && uiState.hasChanges) {
                ExtendedFloatingActionButton(
                    onClick = viewModel::saveMemory,
                    icon = { Icon(Icons.Filled.Save, contentDescription = null) },
                    text = { Text(if (uiState.isSaving) "Saving..." else "Save") }
                )
            }
        }
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding)) {
            TabRow(selectedTabIndex = selectedTab) {
                Tab(selected = selectedTab == 0, onClick = { selectedTab = 0 }, text = { Text("Long-Term") })
                Tab(selected = selectedTab == 1, onClick = { selectedTab = 1 }, text = { Text("Daily Notes") })
            }

            if (uiState.isLoading) {
                LoadingScreen()
            } else when (selectedTab) {
                0 -> OutlinedTextField(
                    value = uiState.longTermMemory,
                    onValueChange = viewModel::updateMemoryText,
                    label = { Text("Persistent memory context") },
                    modifier = Modifier.fillMaxWidth().weight(1f).padding(16.dp),
                    minLines = 10
                )
                1 -> {
                    if (uiState.dailyNotes.isEmpty()) {
                        Text(
                            "No daily observations yet",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            modifier = Modifier.padding(16.dp)
                        )
                    } else {
                        LazyColumn(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
                            item { Spacer(modifier = Modifier.height(8.dp)) }
                            items(uiState.dailyNotes) { note ->
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
                        }
                    }
                }
            }
        }
    }
}
