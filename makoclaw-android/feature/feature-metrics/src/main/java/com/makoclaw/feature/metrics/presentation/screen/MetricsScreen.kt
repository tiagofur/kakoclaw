package com.makoclaw.feature.metrics.presentation.screen

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
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.metrics.presentation.viewmodel.MetricsViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MetricsScreen(viewModel: MetricsViewModel = hiltViewModel()) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Metrics") },
                actions = { IconButton(onClick = viewModel::loadMetrics) { Icon(Icons.Filled.Refresh, "Refresh") } }
            )
        }
    ) { padding ->
        if (uiState.isLoading) { LoadingScreen(modifier = Modifier.padding(padding)); return@Scaffold }

        val m = uiState.metrics
        Column(
            modifier = Modifier.fillMaxSize().padding(padding).verticalScroll(rememberScrollState()).padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Summary cards
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                MetricCard("LLM Calls", m.llmCalls.toString(), Modifier.weight(1f))
                MetricCard("Tokens", m.totalTokens.toString(), Modifier.weight(1f))
            }
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                MetricCard("Cost", "$${String.format("%.2f", m.totalCost)}", Modifier.weight(1f))
                MetricCard("Tool Calls", m.toolInvocations.toString(), Modifier.weight(1f))
            }

            // Agent breakdown
            if (m.agentBreakdown.isNotEmpty()) {
                Text("Agent Breakdown", style = MaterialTheme.typography.titleMedium)
                m.agentBreakdown.entries.sortedByDescending { it.value }.forEach { (agent, count) ->
                    Card(modifier = Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainer)) {
                        Row(modifier = Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                            Text(agent, style = MaterialTheme.typography.bodyMedium, modifier = Modifier.weight(1f))
                            Text(count.toString(), style = MaterialTheme.typography.titleSmall, color = MaterialTheme.colorScheme.primary)
                        }
                    }
                }
            }

            // Tool breakdown
            if (m.toolBreakdown.isNotEmpty()) {
                Text("Tool Usage", style = MaterialTheme.typography.titleMedium)
                m.toolBreakdown.entries.sortedByDescending { it.value }.forEach { (tool, count) ->
                    Card(modifier = Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainer)) {
                        Row(modifier = Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                            Text(tool, style = MaterialTheme.typography.bodyMedium, modifier = Modifier.weight(1f))
                            Text(count.toString(), style = MaterialTheme.typography.titleSmall, color = MaterialTheme.colorScheme.secondary)
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))
        }
    }
}

@Composable
private fun MetricCard(label: String, value: String, modifier: Modifier = Modifier) {
    Card(modifier = modifier, colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceContainer)) {
        Column(modifier = Modifier.padding(16.dp), horizontalAlignment = Alignment.CenterHorizontally) {
            Text(value, style = MaterialTheme.typography.headlineMedium)
            Text(label, style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}
