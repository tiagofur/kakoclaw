package com.makoclaw.feature.auth.presentation.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.makoclaw.feature.auth.presentation.viewmodel.OnboardingEffect
import com.makoclaw.feature.auth.presentation.viewmodel.OnboardingViewModel

@Composable
fun OnboardingScreen(
    onComplete: () -> Unit,
    viewModel: OnboardingViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val snackbar = remember { SnackbarHostState() }

    val stepTitles = listOf(
        "Welcome to MakoClaw",
        "Configure AI Provider",
        "Set Up Workspace",
        "Connect Channels",
        "Review Configuration",
        "You're All Set!"
    )
    val stepDescriptions = listOf(
        "Your AI agent framework is ready to be configured.",
        "Connect an AI provider to power your agents.",
        "Set up your workspace and install initial skills.",
        "Connect messaging channels (optional).",
        "Review your configuration before getting started.",
        "Your MakoClaw instance is ready to use!"
    )

    LaunchedEffect(Unit) {
        viewModel.effects.collect { effect ->
            when (effect) {
                is OnboardingEffect.Complete -> onComplete()
                is OnboardingEffect.ShowError -> snackbar.showSnackbar(effect.message)
            }
        }
    }

    Scaffold(snackbarHost = { SnackbarHost(snackbar) }) { padding ->
        Column(
            modifier = Modifier.fillMaxSize().padding(padding).padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            LinearProgressIndicator(
                progress = { (uiState.currentStep + 1).toFloat() / uiState.totalSteps },
                modifier = Modifier.fillMaxWidth()
            )
            Text(
                "Step ${uiState.currentStep + 1} of ${uiState.totalSteps}",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 8.dp)
            )

            Spacer(modifier = Modifier.height(48.dp))

            Text(stepTitles[uiState.currentStep], style = MaterialTheme.typography.headlineMedium)
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                stepDescriptions[uiState.currentStep],
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            // Provider config on step 1
            if (uiState.currentStep == 1) {
                Spacer(modifier = Modifier.height(24.dp))
                OutlinedTextField(
                    value = uiState.providerName,
                    onValueChange = viewModel::onProviderNameChanged,
                    label = { Text("Provider (e.g. openai, anthropic)") },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true
                )
                Spacer(modifier = Modifier.height(12.dp))
                OutlinedTextField(
                    value = uiState.apiKey,
                    onValueChange = viewModel::onApiKeyChanged,
                    label = { Text("API Key") },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true
                )
            }

            Spacer(modifier = Modifier.weight(1f))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                if (uiState.currentStep > 0) {
                    OutlinedButton(onClick = viewModel::previousStep) { Text("Back") }
                } else {
                    Spacer(modifier = Modifier)
                }

                Button(
                    onClick = {
                        if (uiState.currentStep == 1 && uiState.providerName.isNotBlank()) {
                            viewModel.validateProvider()
                        } else {
                            viewModel.nextStep()
                        }
                    },
                    enabled = !uiState.isValidating
                ) {
                    Text(
                        when {
                            uiState.isValidating -> "Validating..."
                            uiState.currentStep == 1 -> "Validate & Continue"
                            uiState.currentStep < uiState.totalSteps - 1 -> "Next"
                            else -> "Get Started"
                        }
                    )
                }
            }
        }
    }
}
