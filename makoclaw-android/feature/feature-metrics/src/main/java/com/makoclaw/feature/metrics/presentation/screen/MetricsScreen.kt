package com.makoclaw.feature.metrics.presentation.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
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
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.core.ui.component.EmptyState
import com.makoclaw.core.ui.component.ErrorScreen
import com.makoclaw.core.ui.component.LoadingScreen
import com.makoclaw.feature.metrics.presentation.component.AgentFilterDropdown
import com.makoclaw.feature.metrics.presentation.component.ExportButton
import com.makoclaw.feature.metrics.presentation.component.MetricsBarChart
import com.makoclaw.feature.metrics.presentation.component.MetricsDonutChart
import com.makoclaw.feature.metrics.presentation.component.MetricsLineChart
import com.makoclaw.feature.metrics.presentation.component.TimeRangeSelector
import com.makoclaw.feature.metrics.presentation.state.MetricsEffect
import com.makoclaw.feature.metrics.presentation.state.MetricsEvent
import com.makoclaw.feature.metrics.presentation.viewmodel.MetricsViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MetricsScreen(viewModel: MetricsViewModel = hiltViewModel()) {
    val uiState by viewModel.uiState.collectAsState()
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(Unit) {
        viewModel.effects.collect { effect ->
            when (effect) {
                is MetricsEffect.ShowSnackbar ->
                    snackbarHostState.showSnackbar(effect.message)
                is MetricsEffect.ExportComplete ->
                    snackbarHostState.showSnackbar("Exported to ${effect.filePath}")
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Metrics") },
                actions = {
                    IconButton(onClick = { viewModel.onEvent(MetricsEvent.Refresh) }) {
                        Icon(Icons.Filled.Refresh, contentDescription = "Refresh")
                    }
                }
            )
        },
        snackbarHost = { SnackbarHost(snackbarHostState) }
    ) { padding ->
        when {
            uiState.isLoading && uiState.metricsData.totalExecutions == 0 -> {
                LoadingScreen(modifier = Modifier.padding(padding))
            }
            uiState.error != null && uiState.metricsData.totalExecutions == 0 -> {
                ErrorScreen(
                    message = uiState.error,
                    modifier = Modifier.padding(padding),
                    onRetry = { viewModel.onEvent(MetricsEvent.Refresh) }
                )
            }
            uiState.metricsData.totalExecutions == 0 -> {
                EmptyState(
                    title = "No Metrics Data",
                    message = "Metrics will appear once you start using agents.",
                    modifier = Modifier.padding(padding)
                )
            }
            else -> {
                MetricsContent(
                    state = uiState,
                    onEvent = viewModel::onEvent,
                    modifier = Modifier.padding(padding)
                )
            }
        }
    }
}

@Composable
private fun MetricsContent(
    state: com.makoclaw.feature.metrics.presentation.state.MetricsUiState,
    onEvent: (MetricsEvent) -> Unit,
    modifier: Modifier = Modifier
) {
    val metrics = state.metricsData
    val configuration = LocalConfiguration.current
    val isWideScreen = configuration.screenWidthDp > 600

    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // ── Filters row ──
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            TimeRangeSelector(
                selectedTimeRange = state.selectedTimeRange,
                onTimeRangeSelected = { onEvent(MetricsEvent.SetTimeRange(it)) },
                modifier = Modifier.weight(1f)
            )
            AgentFilterDropdown(
                agents = state.availableAgents,
                selectedAgentId = state.selectedAgentId,
                onAgentSelected = { onEvent(MetricsEvent.SetAgentFilter(it)) },
                modifier = Modifier.weight(1f)
            )
        }

        // ── Summary cards ──
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            MetricCard("Executions", metrics.totalExecutions.toString(), Modifier.weight(1f))
            MetricCard(
                "Success Rate",
                String.format("%.1f%%", metrics.successRate * 100),
                Modifier.weight(1f)
            )
        }
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            MetricCard("Avg Duration", String.format("%.1fs", metrics.avgDuration), Modifier.weight(1f))
            MetricCard("Active Agents", metrics.activeAgents.toString(), Modifier.weight(1f))
        }

        // ── Line chart: Executions over time ──
        if (metrics.executionsOverTime.isNotEmpty()) {
            ChartSection(title = "Executions Over Time") {
                MetricsLineChart(
                    data = metrics.executionsOverTime,
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(if (isWideScreen) 280.dp else 220.dp)
                )
            }
        }

        // ── Bar chart: Agent usage ──
        if (metrics.agentUsage.isNotEmpty()) {
            ChartSection(title = "Agent Usage") {
                MetricsBarChart(
                    data = metrics.agentUsage,
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(if (isWideScreen) 280.dp else 220.dp)
                )
            }
        }

        // ── Donut chart: Cost distribution ──
        if (metrics.costDistribution.isNotEmpty()) {
            ChartSection(title = "Cost Distribution") {
                MetricsDonutChart(
                    segments = metrics.costDistribution,
                    modifier = Modifier.fillMaxWidth(),
                    size = if (isWideScreen) 220.dp else 180.dp,
                    strokeWidth = if (isWideScreen) 32.dp else 28.dp
                )
            }
        }

        // ── Export button ──
        ExportButton(
            onClick = { onEvent(MetricsEvent.ExportMetrics) },
            isExporting = state.isExporting,
            modifier = Modifier.fillMaxWidth()
        )

        Spacer(modifier = Modifier.height(16.dp))
    }
}

@Composable
private fun ChartSection(
    title: String,
    content: @Composable () -> Unit
) {
    Column {
        Text(
            text = title,
            style = MaterialTheme.typography.titleMedium,
            color = MaterialTheme.colorScheme.onSurface,
            modifier = Modifier.padding(bottom = 8.dp)
        )
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.surfaceContainer
            )
        ) {
            Column(modifier = Modifier.padding(12.dp)) {
                content()
            }
        }
    }
}

@Composable
private fun MetricCard(label: String, value: String, modifier: Modifier = Modifier) {
    Card(
        modifier = modifier,
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceContainer
        )
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(value, style = MaterialTheme.typography.headlineMedium)
            Text(
                label,
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}
