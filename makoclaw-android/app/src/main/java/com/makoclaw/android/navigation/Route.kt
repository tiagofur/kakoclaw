package com.makoclaw.android.navigation

import kotlinx.serialization.Serializable

sealed interface Route {
    // Auth graph
    @Serializable data object Landing : Route
    @Serializable data object Login : Route
    @Serializable data object Signup : Route
    @Serializable data object Onboarding : Route
    @Serializable data object Setup : Route

    // Main graph
    @Serializable data object Chat : Route
    @Serializable data object Dashboard : Route
    @Serializable data object Tasks : Route
    @Serializable data object Agents : Route
    @Serializable data object History : Route
    @Serializable data object Knowledge : Route
    @Serializable data object Memory : Route
    @Serializable data object Skills : Route
    @Serializable data object Workflows : Route
    @Serializable data object Metrics : Route
    @Serializable data object Cron : Route
    @Serializable data object Files : Route
    @Serializable data object Mcp : Route
    @Serializable data object Reports : Route
    @Serializable data object Settings : Route
}
