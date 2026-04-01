package com.makoclaw.feature.skills.presentation.screen

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
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Download
import androidx.compose.material.icons.filled.Extension
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.MenuAnchorType
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.model.Skill
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.ErrorScreen
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.skills.presentation.state.SkillsEffect
import com.makoclaw.feature.skills.presentation.state.SkillsEvent
import com.makoclaw.feature.skills.presentation.viewmodel.SkillsViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SkillsScreen(viewModel: SkillsViewModel = hiltViewModel()) {
    val uiState by viewModel.uiState.collectAsState()
    val snackbarHostState = remember { SnackbarHostState() }
    val tabs = listOf("Installed", "Marketplace", "Generate")
    var showUninstallDialog by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(Unit) {
        viewModel.effects.collect { effect ->
            when (effect) {
                is SkillsEffect.ShowSnackbar -> snackbarHostState.showSnackbar(effect.message)
                is SkillsEffect.SkillInstalled -> { /* list refreshes via Flow */ }
                is SkillsEffect.SkillUninstalled -> showUninstallDialog = null
                is SkillsEffect.SkillGenerated -> { /* generatedSkill in state */ }
            }
        }
    }

    Scaffold(
        topBar = { TopAppBar(title = { Text("Skills") }) },
        snackbarHost = { SnackbarHost(snackbarHostState) }
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding)) {
            // Search bar
            OutlinedTextField(
                value = uiState.searchQuery,
                onValueChange = { viewModel.onEvent(SkillsEvent.UpdateSearchQuery(it)) },
                label = { Text("Search skills") },
                leadingIcon = { Icon(Icons.Filled.Search, contentDescription = null) },
                modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 8.dp),
                singleLine = true
            )

            TabRow(selectedTabIndex = uiState.selectedTab) {
                tabs.forEachIndexed { index, title ->
                    Tab(
                        selected = uiState.selectedTab == index,
                        onClick = { viewModel.onEvent(SkillsEvent.SelectTab(index)) },
                        text = { Text(title) }
                    )
                }
            }

            when {
                uiState.isLoading && uiState.selectedTab == 0 -> LoadingScreen()
                uiState.error != null && uiState.selectedTab == 0 -> ErrorScreen(
                    message = uiState.error ?: "Unknown error",
                    onRetry = { viewModel.onEvent(SkillsEvent.Refresh) }
                )
                else -> when (uiState.selectedTab) {
                    0 -> InstalledTab(
                        skills = filterSkills(uiState.installedSkills, uiState.searchQuery),
                        isInstalling = uiState.isInstalling,
                        onUninstall = { showUninstallDialog = it },
                        onGenerate = { viewModel.onEvent(SkillsEvent.GenerateSkill(it, "")) }
                    )
                    1 -> MarketplaceTab(
                        skills = filterSkills(uiState.marketplaceSkills, uiState.searchQuery),
                        onInstall = { viewModel.onEvent(SkillsEvent.InstallSkill(it)) }
                    )
                    2 -> GenerateTab(
                        isGenerating = uiState.isGenerating,
                        generatedSkill = uiState.generatedSkill,
                        onGenerate = { name, goal -> viewModel.onEvent(SkillsEvent.GenerateSkill(name, goal)) }
                    )
                }
            }
        }
    }

    // Uninstall confirmation dialog
    showUninstallDialog?.let { skillName ->
        AlertDialog(
            onDismissRequest = { showUninstallDialog = null },
            title = { Text("Uninstall Skill") },
            text = { Text("Are you sure you want to uninstall \"$skillName\"?") },
            confirmButton = {
                TextButton(
                    onClick = { viewModel.onEvent(SkillsEvent.UninstallSkill(skillName)) }
                ) { Text("Uninstall", color = MaterialTheme.colorScheme.error) }
            },
            dismissButton = {
                TextButton(onClick = { showUninstallDialog = null }) { Text("Cancel") }
            }
        )
    }
}

private fun filterSkills(skills: List<Skill>, query: String): List<Skill> {
    if (query.isBlank()) return skills
    return skills.filter {
        it.name.contains(query, ignoreCase = true) ||
        it.content.contains(query, ignoreCase = true) ||
        it.repository.contains(query, ignoreCase = true)
    }
}

@Composable
private fun InstalledTab(
    skills: List<Skill>,
    isInstalling: Boolean,
    onUninstall: (String) -> Unit,
    onGenerate: (String) -> Unit
) {
    if (skills.isEmpty()) {
        EmptyState(
            title = "No Skills Installed",
            message = "Browse the marketplace or generate one with AI.",
            icon = Icons.Filled.Extension
        )
        return
    }
    LazyColumn(
        modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        item { Spacer(modifier = Modifier.height(8.dp)) }
        items(skills, key = { it.name }) { skill ->
            SkillCard(
                skill = skill,
                isInstalled = true,
                isLoading = isInstalling,
                onAction = { onUninstall(skill.name) }
            )
        }
        item { Spacer(modifier = Modifier.height(8.dp)) }
    }
}

@Composable
private fun MarketplaceTab(
    skills: List<Skill>,
    onInstall: (String) -> Unit
) {
    if (skills.isEmpty()) {
        EmptyState(
            title = "Loading Marketplace",
            message = "Fetching available skills...",
            icon = Icons.Filled.Extension
        )
        return
    }
    LazyColumn(
        modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        item { Spacer(modifier = Modifier.height(8.dp)) }
        items(skills, key = { it.name + it.repository }) { skill ->
            SkillCard(
                skill = skill,
                isInstalled = false,
                isLoading = false,
                onAction = { onInstall(skill.repository) }
            )
        }
        item { Spacer(modifier = Modifier.height(8.dp)) }
    }
}

@Composable
private fun GenerateTab(
    isGenerating: Boolean,
    generatedSkill: Skill?,
    onGenerate: (String, String) -> Unit
) {
    var name by remember { mutableStateOf("") }
    var goal by remember { mutableStateOf("") }

    Column(modifier = Modifier.padding(16.dp)) {
        Text("AI Skill Generation", style = MaterialTheme.typography.titleMedium)
        Text(
            "Describe a skill and AI will generate it for you.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.padding(top = 8.dp, bottom = 16.dp)
        )

        OutlinedTextField(
            value = name,
            onValueChange = { name = it },
            label = { Text("Skill Name") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true
        )
        Spacer(modifier = Modifier.height(12.dp))

        OutlinedTextField(
            value = goal,
            onValueChange = { goal = it },
            label = { Text("Goal / Description") },
            modifier = Modifier.fillMaxWidth(),
            minLines = 3,
            maxLines = 6
        )
        Spacer(modifier = Modifier.height(16.dp))

        OutlinedButton(
            onClick = { onGenerate(name, goal) },
            modifier = Modifier.fillMaxWidth(),
            enabled = name.isNotBlank() && goal.isNotBlank() && !isGenerating
        ) {
            if (isGenerating) {
                CircularProgressIndicator(modifier = Modifier.size(18.dp), strokeWidth = 2.dp)
                Spacer(modifier = Modifier.width(8.dp))
            }
            Icon(Icons.Filled.AutoAwesome, contentDescription = null, modifier = Modifier.size(18.dp))
            Spacer(modifier = Modifier.width(8.dp))
            Text(if (isGenerating) "Generating..." else "Generate Skill")
        }

        generatedSkill?.let { skill ->
            Spacer(modifier = Modifier.height(16.dp))
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.primaryContainer)
            ) {
                Column(modifier = Modifier.padding(16.dp)) {
                    Text("Generated: ${skill.name}", style = MaterialTheme.typography.titleSmall)
                    Text(skill.content.take(200), style = MaterialTheme.typography.bodySmall)
                }
            }
        }
    }
}

@Composable
private fun SkillCard(
    skill: Skill,
    isInstalled: Boolean,
    isLoading: Boolean,
    onAction: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainer)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(Icons.Filled.Extension, null, tint = MaterialTheme.colorScheme.primary)
            Spacer(modifier = Modifier.width(12.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(skill.name, style = MaterialTheme.typography.titleSmall)
                if (skill.repository.isNotEmpty()) {
                    Text(
                        skill.repository,
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            if (isLoading) {
                CircularProgressIndicator(modifier = Modifier.size(24.dp), strokeWidth = 2.dp)
            } else if (isInstalled) {
                IconButton(onClick = onAction) {
                    Icon(Icons.Filled.Delete, "Uninstall", tint = MaterialTheme.colorScheme.error)
                }
            } else {
                IconButton(onClick = onAction) {
                    Icon(Icons.Filled.Download, "Install", tint = MaterialTheme.colorScheme.primary)
                }
            }
        }
    }
}
