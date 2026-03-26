package com.makoclaw.core.model

import kotlinx.serialization.Serializable

@Serializable
data class ChatMessage(
    val id: String = "",
    val role: String,           // user, assistant, system
    val content: String,
    val timestamp: String = "",
    val agents: List<String> = emptyList(),
    val toolCalls: List<ToolCall> = emptyList(),
    val thinkingContent: String = "",
    val specialistName: String? = null,
    val isStreaming: Boolean = false
)

@Serializable
data class ChatSession(
    val sessionId: String,
    val title: String = "",
    val lastMessage: String = "",
    val messageCount: Int = 0,
    val updatedAt: String = "",
    val archived: Boolean = false
)

@Serializable
data class ToolCall(
    val name: String,
    val args: Map<String, String> = emptyMap(),
    val result: String? = null,
    val status: String = "pending" // pending, success, error
)

@Serializable
data class AgentEvent(
    val agent: String,
    val status: String,           // working, complete, timeout, failed
    val specialistName: String = "",
    val reason: String = "",
    val delegationChain: List<String> = emptyList(),
    val delegationDepth: Int = 0,
    val activeSkills: List<String> = emptyList(),
    val timestamp: String = ""
)

@Serializable
data class SpecialistReport(
    val specialistName: String,
    val status: String,           // success, needs_help, failed
    val confidence: Double = 0.0,
    val suggestions: List<String> = emptyList(),
    val toolsUsed: List<String> = emptyList(),
    val iterationsUsed: Int = 0,
    val timestamp: String = ""
)

@Serializable
data class ChatSendMessage(
    val type: String = "message",
    val content: String,
    val session_id: String = "web:chat",
    val model: String? = null,
    val web_search: Boolean? = null,
    val exclude_tools: List<String> = emptyList(),
    val target_specialist: String? = null
)

@Serializable
data class ForkRequest(
    val session_id: String,
    val message_index: Int
)
