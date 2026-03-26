package com.makoclaw.feature.skills.presentation.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Download
import androidx.compose.material.icons.filled.Extension
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.model.Skill
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.skills.presentation.viewmodel.SkillsViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SkillsScreen(viewModel: SkillsViewModel = hiltViewModel()) {
    val uiState by viewModel.uiState.collectAsState()
    val tabs = listOf("Installed", "Marketplace", "Generate")

    Scaffold(
        topBar = { TopAppBar(title = { Text("Skills") }) }
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding)) {
            TabRow(selectedTabIndex = uiState.selectedTab) {
                tabs.forEachIndexed { index, title ->
                    Tab(selected = uiState.selectedTab == index, onClick = { viewModel.selectTab(index) }, text = { Text(title) })
                }
            }

            when {
                uiState.isLoading && uiState.selectedTab == 0 -> LoadingScreen()
                else -> when (uiState.selectedTab) {
                    0 -> SkillList(uiState.installedSkills, installed = true, onUninstall = viewModel::uninstallSkill)
                    1 -> SkillList(uiState.marketplaceSkills, installed = false, onInstall = { viewModel.installSkill(it) })
                    2 -> Column(modifier = Modifier.padding(16.dp)) {
                        Text("AI Skill Generation", style = MaterialTheme.typography.titleMedium)
                        Text("Describe a skill and AI will generate it for you.",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            modifier = Modifier.padding(top = 8.dp))
                    }
                }
            }
        }
    }
}

@Composable
private fun SkillList(
    skills: List<Skill>,
    installed: Boolean,
    onUninstall: ((String) -> Unit)? = null,
    onInstall: ((String) -> Unit)? = null
) {
    if (skills.isEmpty()) {
        EmptyState(title = if (installed) "No skills installed" else "Marketplace loading...", icon = Icons.Filled.Extension)
        return
    }
    LazyColumn(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
        item { Spacer(modifier = Modifier.height(8.dp)) }
        items(skills, key = { it.name }) { skill ->
            Card(modifier = Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainer)) {
                Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                    Icon(Icons.Filled.Extension, null, tint = MaterialTheme.colorScheme.primary)
                    Spacer(modifier = Modifier.width(12.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text(skill.name, style = MaterialTheme.typography.titleSmall)
                        if (skill.repository.isNotEmpty()) Text(skill.repository, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                    if (installed && onUninstall != null) {
                        IconButton(onClick = { onUninstall(skill.name) }) { Icon(Icons.Filled.Delete, "Uninstall", tint = MaterialTheme.colorScheme.error) }
                    }
                    if (!installed && onInstall != null) {
                        IconButton(onClick = { onInstall(skill.repository) }) { Icon(Icons.Filled.Download, "Install", tint = MaterialTheme.colorScheme.primary) }
                    }
                }
            }
        }
    }
}
