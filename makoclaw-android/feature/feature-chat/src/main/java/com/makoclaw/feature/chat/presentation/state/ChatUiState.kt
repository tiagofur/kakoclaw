package com.makoclaw.feature.chat.presentation.state

import com.makoclaw.core.model.AgentEvent
import com.makoclaw.core.model.ChatMessage
import com.makoclaw.core.model.ChatSession
import com.makoclaw.core.model.ToolCall

data class ChatUiState(
    val messages: List<ChatMessage> = emptyList(),
    val sessions: List<ChatSession> = emptyList(),
    val currentSessionId: String = "web:chat",
    val isConnected: Boolean = false,
    val isStreaming: Boolean = false,
    val inputText: String = "",
    val selectedModel: String = "",
    val availableModels: List<String> = emptyList(),
    val activeSpecialist: String? = null,
    val delegationReason: String = "",
    val agentEvents: List<AgentEvent> = emptyList(),
    val activeToolCalls: List<ToolCall> = emptyList(),
    val thinkingContent: String = "",
    val isThinkingExpanded: Boolean = false,
    val error: String? = null
)

sealed interface ChatEvent {
    data class SendMessage(val content: String) : ChatEvent
    data class UpdateInput(val text: String) : ChatEvent
    data class SelectModel(val model: String) : ChatEvent
    data class LoadSession(val sessionId: String) : ChatEvent
    data object NewSession : ChatEvent
    data object CancelExecution : ChatEvent
    data object Reconnect : ChatEvent
    data object ToggleThinking : ChatEvent
    data object DismissError : ChatEvent
}

sealed interface ChatEffect {
    data class ShowSnackbar(val message: String) : ChatEffect
    data object ScrollToBottom : ChatEffect
}
