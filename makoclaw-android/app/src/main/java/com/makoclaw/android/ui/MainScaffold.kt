package com.makoclaw.android.ui

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AccountTree
import androidx.compose.material.icons.filled.Analytics
import androidx.compose.material.icons.filled.Assessment
import androidx.compose.material.icons.filled.Chat
import androidx.compose.material.icons.filled.Dashboard
import androidx.compose.material.icons.filled.Extension
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.Groups
import androidx.compose.material.icons.filled.History
import androidx.compose.material.icons.filled.Hub
import androidx.compose.material.icons.filled.LibraryBooks
import androidx.compose.material.icons.filled.Menu
import androidx.compose.material.icons.filled.Memory
import androidx.compose.material.icons.filled.Schedule
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.SmartToy
import androidx.compose.material.icons.filled.TaskAlt
import androidx.compose.material.icons.outlined.Chat
import androidx.compose.material.icons.outlined.Dashboard
import androidx.compose.material.icons.outlined.Groups
import androidx.compose.material.icons.outlined.TaskAlt
import androidx.compose.material3.DrawerValue
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalDrawerSheet
import androidx.compose.material3.ModalNavigationDrawer
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.NavigationDrawerItem
import androidx.compose.material3.NavigationDrawerItemDefaults
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.rememberDrawerState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
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
    val icon: ImageVector,
    val section: String = ""
)

private val bottomNavItems = listOf(
    BottomNavItem(Route.Chat, "Chat", Icons.Filled.Chat, Icons.Outlined.Chat),
    BottomNavItem(Route.Tasks, "Tasks", Icons.Filled.TaskAlt, Icons.Outlined.TaskAlt),
    BottomNavItem(Route.Dashboard, "Home", Icons.Filled.Dashboard, Icons.Outlined.Dashboard),
    BottomNavItem(Route.Agents, "Agents", Icons.Filled.Groups, Icons.Outlined.Groups),
)

private val drawerNavItems = listOf(
    // Conversations
    DrawerNavItem(Route.History, "History", Icons.Filled.History, "Conversations"),
    DrawerNavItem(Route.Memory, "Memory", Icons.Filled.Memory, "Conversations"),
    // Knowledge
    DrawerNavItem(Route.Knowledge, "Knowledge Base", Icons.Filled.LibraryBooks, "Knowledge"),
    DrawerNavItem(Route.Skills, "Skills & Marketplace", Icons.Filled.Extension, "Knowledge"),
    // Automation
    DrawerNavItem(Route.Workflows, "Workflows", Icons.Filled.AccountTree, "Automation"),
    DrawerNavItem(Route.Cron, "Scheduled Tasks", Icons.Filled.Schedule, "Automation"),
    // System
    DrawerNavItem(Route.Metrics, "Metrics", Icons.Filled.Analytics, "System"),
    DrawerNavItem(Route.Files, "Files", Icons.Filled.Folder, "System"),
    DrawerNavItem(Route.Mcp, "MCP Servers", Icons.Filled.Hub, "System"),
    DrawerNavItem(Route.Reports, "Reports", Icons.Filled.Assessment, "System"),
    // Config
    DrawerNavItem(Route.Settings, "Settings", Icons.Filled.Settings, ""),
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
                // App header
                Spacer(modifier = Modifier.height(16.dp))
                Icon(
                    imageVector = Icons.Filled.SmartToy,
                    contentDescription = null,
                    modifier = Modifier
                        .padding(start = 28.dp)
                        .size(32.dp),
                    tint = MaterialTheme.colorScheme.primary
                )
                Text(
                    text = "MakoClaw",
                    modifier = Modifier.padding(start = 28.dp, top = 8.dp),
                    style = MaterialTheme.typography.titleLarge,
                    color = MaterialTheme.colorScheme.primary
                )
                Text(
                    text = "AI Agent Framework",
                    modifier = Modifier.padding(start = 28.dp, bottom = 16.dp),
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                HorizontalDivider(modifier = Modifier.padding(horizontal = 28.dp))
                Spacer(modifier = Modifier.height(8.dp))

                var lastSection = ""
                drawerNavItems.forEach { item ->
                    if (item.section.isNotEmpty() && item.section != lastSection) {
                        if (lastSection.isNotEmpty()) {
                            Spacer(modifier = Modifier.height(4.dp))
                        }
                        Text(
                            text = item.section,
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            modifier = Modifier.padding(
                                start = 28.dp,
                                top = 12.dp,
                                bottom = 4.dp
                            )
                        )
                        lastSection = item.section
                    }

                    NavigationDrawerItem(
                        icon = {
                            Icon(
                                item.icon,
                                contentDescription = item.label,
                                modifier = Modifier.size(22.dp)
                            )
                        },
                        label = { Text(item.label) },
                        selected = currentRoute == item.route,
                        onClick = {
                            scope.launch { drawerState.close() }
                            navController.navigate(item.route) {
                                popUpTo(Route.Chat)
                                launchSingleTop = true
                            }
                        },
                        modifier = Modifier.padding(NavigationDrawerItemDefaults.ItemPadding)
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
