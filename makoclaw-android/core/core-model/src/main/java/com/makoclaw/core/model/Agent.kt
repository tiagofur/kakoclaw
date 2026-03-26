package com.makoclaw.core.model

import kotlinx.serialization.Serializable

@Serializable
data class Specialist(
    val name: String,
    val speciality: String = "",
    val systemPrompt: String = "",
    val tools: List<String> = emptyList(),
    val model: String = "",
    val keywords: List<String> = emptyList(),
    val icon: String = "",
    val color: String = ""
)

@Serializable
data class OrchestratorConfig(
    val enabled: Boolean = false,
    val model: String = "",
    val maxDelegationDepth: Int = 3
)

@Serializable
data class AgentMetrics(
    val totalCalls: Int = 0,
    val totalTokens: Long = 0,
    val totalCost: Double = 0.0,
    val agentBreakdown: Map<String, Int> = emptyMap()
)

@Serializable
data class Swarm(
    val name: String,
    val description: String = "",
    val members: List<String> = emptyList(),
    val mode: String = "sequential",  // sequential, parallel, router
    val maxBudget: Int = 100,
    val timeout: Int = 3600,
    val sharedMemory: Boolean = true
)

@Serializable
data class CreateSpecialistRequest(
    val name: String,
    val speciality: String,
    val system_prompt: String,
    val tools: List<String> = emptyList(),
    val model: String? = null
)
