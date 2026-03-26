package com.makoclaw.android.navigation

import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.makoclaw.android.ui.MainScaffold
import com.makoclaw.feature.agents.presentation.screen.AgentsScreen
import com.makoclaw.feature.auth.presentation.screen.LoginScreen
import com.makoclaw.feature.auth.presentation.screen.OnboardingScreen
import com.makoclaw.feature.auth.presentation.screen.SignupScreen
import com.makoclaw.feature.chat.presentation.screen.ChatScreen
import com.makoclaw.feature.cron.presentation.screen.CronScreen
import com.makoclaw.feature.dashboard.presentation.screen.DashboardScreen
import com.makoclaw.feature.files.presentation.screen.FilesScreen
import com.makoclaw.feature.history.presentation.screen.HistoryScreen
import com.makoclaw.feature.knowledge.presentation.screen.KnowledgeScreen
import com.makoclaw.feature.mcp.presentation.screen.McpScreen
import com.makoclaw.feature.memory.presentation.screen.MemoryScreen
import com.makoclaw.feature.metrics.presentation.screen.MetricsScreen
import com.makoclaw.feature.reports.presentation.screen.ReportsScreen
import com.makoclaw.feature.settings.presentation.screen.SettingsScreen
import com.makoclaw.feature.skills.presentation.screen.SkillsScreen
import com.makoclaw.feature.tasks.presentation.screen.TasksScreen
import com.makoclaw.feature.workflows.presentation.screen.WorkflowsScreen

@Composable
fun MakoClawNavHost() {
    val navController = rememberNavController()
    val authViewModel: AuthNavigationViewModel = hiltViewModel()
    val isAuthenticated by authViewModel.isAuthenticated.collectAsState()

    val startDestination: Any = if (isAuthenticated) Route.Chat else Route.Login

    NavHost(
        navController = navController,
        startDestination = startDestination
    ) {
        // Auth flow
        composable<Route.Login> {
            LoginScreen(
                onLoginSuccess = {
                    navController.navigate(Route.Chat) {
                        popUpTo(Route.Login) { inclusive = true }
                    }
                },
                onNavigateToSignup = {
                    navController.navigate(Route.Signup)
                }
            )
        }

        composable<Route.Signup> {
            SignupScreen(
                onSignupSuccess = {
                    navController.navigate(Route.Login) {
                        popUpTo(Route.Signup) { inclusive = true }
                    }
                },
                onNavigateToLogin = { navController.popBackStack() }
            )
        }

        composable<Route.Onboarding> {
            OnboardingScreen(
                onComplete = {
                    navController.navigate(Route.Chat) {
                        popUpTo(Route.Onboarding) { inclusive = true }
                    }
                }
            )
        }

        // Main app (with scaffold)
        composable<Route.Chat> {
            MainScaffold(navController = navController, currentRoute = Route.Chat) {
                ChatScreen()
            }
        }

        composable<Route.Tasks> {
            MainScaffold(navController = navController, currentRoute = Route.Tasks) {
                TasksScreen()
            }
        }

        composable<Route.Dashboard> {
            MainScaffold(navController = navController, currentRoute = Route.Dashboard) {
                DashboardScreen()
            }
        }

        composable<Route.Agents> {
            MainScaffold(navController = navController, currentRoute = Route.Agents) {
                AgentsScreen()
            }
        }

        composable<Route.History> {
            MainScaffold(navController = navController, currentRoute = Route.History) {
                HistoryScreen()
            }
        }

        composable<Route.Knowledge> {
            MainScaffold(navController = navController, currentRoute = Route.Knowledge) {
                KnowledgeScreen()
            }
        }

        composable<Route.Memory> {
            MainScaffold(navController = navController, currentRoute = Route.Memory) {
                MemoryScreen()
            }
        }

        composable<Route.Skills> {
            MainScaffold(navController = navController, currentRoute = Route.Skills) {
                SkillsScreen()
            }
        }

        composable<Route.Workflows> {
            MainScaffold(navController = navController, currentRoute = Route.Workflows) {
                WorkflowsScreen()
            }
        }

        composable<Route.Metrics> {
            MainScaffold(navController = navController, currentRoute = Route.Metrics) {
                MetricsScreen()
            }
        }

        composable<Route.Cron> {
            MainScaffold(navController = navController, currentRoute = Route.Cron) {
                CronScreen()
            }
        }

        composable<Route.Files> {
            MainScaffold(navController = navController, currentRoute = Route.Files) {
                FilesScreen()
            }
        }

        composable<Route.Mcp> {
            MainScaffold(navController = navController, currentRoute = Route.Mcp) {
                McpScreen()
            }
        }

        composable<Route.Reports> {
            MainScaffold(navController = navController, currentRoute = Route.Reports) {
                ReportsScreen()
            }
        }

        composable<Route.Settings> {
            MainScaffold(navController = navController, currentRoute = Route.Settings) {
                SettingsScreen()
            }
        }
    }
}
