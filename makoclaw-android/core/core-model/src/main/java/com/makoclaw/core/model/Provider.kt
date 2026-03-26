package com.makoclaw.core.model

import kotlinx.serialization.Serializable

@Serializable
data class ProviderInfo(
    val name: String,
    val enabled: Boolean = false,
    val isActive: Boolean = false,
    val models: List<ModelInfo> = emptyList()
)

@Serializable
data class ModelInfo(
    val id: String,
    val provider: String
)

@Serializable
data class ModelsResponse(
    val current_model: String = "",
    val current_provider: String = "",
    val providers: List<ProviderInfo> = emptyList()
)

@Serializable
data class ConfigStatus(
    val configured: Boolean = false,
    val degraded: Boolean = false,
    val reason: String = "",
    val activeProviders: List<String> = emptyList()
)
