package com.makoclaw.feature.settings.presentation.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Logout
import androidx.compose.material.icons.filled.Person
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.ScrollableTabRow
import androidx.compose.material3.Tab
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.settings.presentation.viewmodel.SettingsViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(
    viewModel: SettingsViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val tabs = listOf("Profile", "Providers", "Agents", "Channels", "Tools", "Audit", "System", "Soul")

    Scaffold(
        topBar = { TopAppBar(title = { Text("Settings") }) }
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding)) {
            ScrollableTabRow(selectedTabIndex = uiState.selectedTab, edgePadding = 16.dp) {
                tabs.forEachIndexed { index, title ->
                    Tab(
                        selected = uiState.selectedTab == index,
                        onClick = { viewModel.selectTab(index) },
                        text = { Text(title) }
                    )
                }
            }

            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f)
                    .verticalScroll(rememberScrollState())
                    .padding(16.dp)
            ) {
                when (uiState.selectedTab) {
                    0 -> ProfileTab(viewModel, uiState.user?.username ?: "", uiState.user?.email ?: "")
                    1 -> ProvidersTab(uiState.models)
                    2 -> PlaceholderTab("Agent configuration — manage specialist agents")
                    3 -> PlaceholderTab("Channel configuration — Telegram, Discord, Slack, etc.")
                    4 -> ToolsTab(uiState.tools)
                    5 -> PlaceholderTab("Audit log — tool usage and activity history")
                    6 -> PlaceholderTab("System settings — core configuration")
                    7 -> PlaceholderTab("Soul settings — AI personality and identity")
                }
            }
        }
    }
}

@Composable
private fun ProfileTab(viewModel: SettingsViewModel, username: String, email: String) {
    var currentPassword by remember { mutableStateOf("") }
    var newPassword by remember { mutableStateOf("") }

    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainer)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(Icons.Filled.Person, null, tint = MaterialTheme.colorScheme.primary)
                Column(modifier = Modifier.padding(start = 12.dp)) {
                    Text(username.ifEmpty { "User" }, style = MaterialTheme.typography.titleMedium)
                    if (email.isNotEmpty()) {
                        Text(email, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            }
        }
    }

    Spacer(modifier = Modifier.height(24.dp))
    Text("Change Password", style = MaterialTheme.typography.titleSmall)
    Spacer(modifier = Modifier.height(8.dp))
    OutlinedTextField(
        value = currentPassword, onValueChange = { currentPassword = it },
        label = { Text("Current Password") }, modifier = Modifier.fillMaxWidth(),
        singleLine = true, visualTransformation = PasswordVisualTransformation()
    )
    Spacer(modifier = Modifier.height(8.dp))
    OutlinedTextField(
        value = newPassword, onValueChange = { newPassword = it },
        label = { Text("New Password") }, modifier = Modifier.fillMaxWidth(),
        singleLine = true, visualTransformation = PasswordVisualTransformation()
    )
    Spacer(modifier = Modifier.height(12.dp))
    Button(
        onClick = {
            viewModel.changePassword(currentPassword, newPassword)
            currentPassword = ""; newPassword = ""
        },
        enabled = currentPassword.isNotBlank() && newPassword.length >= 8
    ) { Text("Change Password") }

    Spacer(modifier = Modifier.height(32.dp))
    Button(
        onClick = viewModel::logout,
        colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.error),
        modifier = Modifier.fillMaxWidth()
    ) {
        Icon(Icons.AutoMirrored.Filled.Logout, null)
        Text("  Sign Out", color = MaterialTheme.colorScheme.onError)
    }
}

@Composable
private fun ProvidersTab(models: com.makoclaw.core.model.ModelsResponse) {
    if (models.providers.isEmpty()) {
        Text("No providers configured", style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
        return
    }
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        models.providers.forEach { provider ->
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainer)
            ) {
                Row(modifier = Modifier.padding(16.dp), verticalAlignment = Alignment.CenterVertically) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text(provider.name, style = MaterialTheme.typography.titleSmall)
                        Text(
                            "${provider.models.size} models",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                    androidx.compose.material3.SuggestionChip(
                        onClick = {},
                        label = { Text(if (provider.enabled) "Active" else "Disabled") }
                    )
                }
            }
        }
    }
}

@Composable
private fun ToolsTab(tools: List<String>) {
    if (tools.isEmpty()) {
        Text("Loading tools...", style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
        return
    }
    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        tools.forEach { tool ->
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainer)
            ) {
                Text(tool, modifier = Modifier.padding(12.dp), style = MaterialTheme.typography.bodyMedium)
            }
        }
    }
}

@Composable
private fun PlaceholderTab(description: String) {
    Text(description, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
}
