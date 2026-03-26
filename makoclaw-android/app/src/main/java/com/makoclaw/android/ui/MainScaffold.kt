package com.makoclaw.android.ui

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Chat
import androidx.compose.material.icons.filled.Dashboard
import androidx.compose.material.icons.filled.Groups
import androidx.compose.material.icons.filled.Menu
import androidx.compose.material.icons.filled.TaskAlt
import androidx.compose.material.icons.outlined.Chat
import androidx.compose.material.icons.outlined.Dashboard
import androidx.compose.material.icons.outlined.Groups
import androidx.compose.material.icons.outlined.TaskAlt
import androidx.compose.material3.DrawerValue
import androidx.compose.material3.Icon
import androidx.compose.material3.ModalDrawerSheet
import androidx.compose.material3.ModalNavigationDrawer
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.NavigationDrawerItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.rememberDrawerState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.navigation.NavController
import com.makoclaw.android.navigation.Route
import kotlinx.coroutines.launch

data class BottomNavItem(
    val route: Route,
    val label: String,
    val selectedIcon: ImageVector,
    val unselectedIcon: ImageVector
)

data class DrawerNavItem(
    val route: Route,
    val label: String,
    val icon: ImageVector
)

private val bottomNavItems = listOf(
    BottomNavItem(Route.Chat, "Chat", Icons.Filled.Chat, Icons.Outlined.Chat),
    BottomNavItem(Route.Tasks, "Tasks", Icons.Filled.TaskAlt, Icons.Outlined.TaskAlt),
    BottomNavItem(Route.Dashboard, "Home", Icons.Filled.Dashboard, Icons.Outlined.Dashboard),
    BottomNavItem(Route.Agents, "Agents", Icons.Filled.Groups, Icons.Outlined.Groups),
)

private val drawerNavItems = listOf(
    DrawerNavItem(Route.History, "History", Icons.Filled.Chat),
    DrawerNavItem(Route.Knowledge, "Knowledge", Icons.Filled.Chat),
    DrawerNavItem(Route.Memory, "Memory", Icons.Filled.Chat),
    DrawerNavItem(Route.Skills, "Skills", Icons.Filled.Chat),
    DrawerNavItem(Route.Workflows, "Workflows", Icons.Filled.Chat),
    DrawerNavItem(Route.Metrics, "Metrics", Icons.Filled.Chat),
    DrawerNavItem(Route.Cron, "Cron", Icons.Filled.Chat),
    DrawerNavItem(Route.Files, "Files", Icons.Filled.Chat),
    DrawerNavItem(Route.Mcp, "MCP", Icons.Filled.Chat),
    DrawerNavItem(Route.Reports, "Reports", Icons.Filled.Chat),
    DrawerNavItem(Route.Settings, "Settings", Icons.Filled.Chat),
)

@Composable
fun MainScaffold(
    navController: NavController,
    currentRoute: Route,
    content: @Composable () -> Unit
) {
    val drawerState = rememberDrawerState(initialValue = DrawerValue.Closed)
    val scope = rememberCoroutineScope()

    ModalNavigationDrawer(
        drawerState = drawerState,
        drawerContent = {
            ModalDrawerSheet {
                Text(
                    text = "MakoClaw",
                    modifier = Modifier.padding(
                        start = androidx.compose.ui.unit.dp.times(16),
                        top = androidx.compose.ui.unit.dp.times(24),
                        bottom = androidx.compose.ui.unit.dp.times(16)
                    ),
                    style = androidx.compose.material3.MaterialTheme.typography.headlineSmall
                )
                drawerNavItems.forEach { item ->
                    NavigationDrawerItem(
                        icon = { Icon(item.icon, contentDescription = item.label) },
                        label = { Text(item.label) },
                        selected = currentRoute == item.route,
                        onClick = {
                            scope.launch { drawerState.close() }
                            navController.navigate(item.route) {
                                popUpTo(Route.Chat)
                                launchSingleTop = true
                            }
                        }
                    )
                }
            }
        }
    ) {
        Scaffold(
            bottomBar = {
                NavigationBar {
                    bottomNavItems.forEach { item ->
                        NavigationBarItem(
                            icon = {
                                Icon(
                                    if (currentRoute == item.route) item.selectedIcon
                                    else item.unselectedIcon,
                                    contentDescription = item.label
                                )
                            },
                            label = { Text(item.label) },
                            selected = currentRoute == item.route,
                            onClick = {
                                navController.navigate(item.route) {
                                    popUpTo(Route.Chat)
                                    launchSingleTop = true
                                }
                            }
                        )
                    }
                    NavigationBarItem(
                        icon = { Icon(Icons.Filled.Menu, contentDescription = "More") },
                        label = { Text("More") },
                        selected = currentRoute !in bottomNavItems.map { it.route },
                        onClick = { scope.launch { drawerState.open() } }
                    )
                }
            }
        ) { innerPadding ->
            Box(modifier = Modifier.padding(innerPadding)) {
                content()
            }
        }
    }
}
