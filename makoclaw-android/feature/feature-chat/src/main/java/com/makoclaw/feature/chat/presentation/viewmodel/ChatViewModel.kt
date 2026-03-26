package com.makoclaw.feature.chat.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.makoclaw.core.model.AgentEvent
import com.makoclaw.core.model.ChatMessage
import com.makoclaw.core.model.ChatSendMessage
import com.makoclaw.core.model.ToolCall
import com.makoclaw.core.network.api.ChatApi
import com.makoclaw.core.network.api.ConfigApi
import com.makoclaw.core.network.websocket.ChatWebSocketClient
import com.makoclaw.core.network.websocket.WsMessage
import com.makoclaw.feature.chat.presentation.state.ChatEffect
import com.makoclaw.feature.chat.presentation.state.ChatEvent
import com.makoclaw.feature.chat.presentation.state.ChatUiState
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import javax.inject.Inject

@HiltViewModel
class ChatViewModel @Inject constructor(
    private val chatApi: ChatApi,
    private val configApi: ConfigApi,
    private val webSocketClient: ChatWebSocketClient,
    private val json: Json
) : ViewModel() {

    private val _uiState = MutableStateFlow(ChatUiState())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()

    private val _effects = MutableSharedFlow<ChatEffect>()
    val effects: SharedFlow<ChatEffect> = _effects.asSharedFlow()

    private val streamingContent = StringBuilder()

    init {
        observeWebSocket()
        observeConnectionState()
        loadModels()
        loadSessions()
        webSocketClient.connect()
    }

    fun onEvent(event: ChatEvent) {
        when (event) {
            is ChatEvent.SendMessage -> sendMessage(event.content)
            is ChatEvent.UpdateInput -> _uiState.update { it.copy(inputText = event.text) }
            is ChatEvent.SelectModel -> _uiState.update { it.copy(selectedModel = event.model) }
            is ChatEvent.LoadSession -> loadSession(event.sessionId)
            is ChatEvent.NewSession -> newSession()
            is ChatEvent.CancelExecution -> cancelExecution()
            is ChatEvent.Reconnect -> webSocketClient.connect()
            is ChatEvent.ToggleThinking -> _uiState.update {
                it.copy(isThinkingExpanded = !it.isThinkingExpanded)
            }
            is ChatEvent.DismissError -> _uiState.update { it.copy(error = null) }
        }
    }

    private fun sendMessage(content: String) {
        if (content.isBlank()) return

        val userMessage = ChatMessage(
            role = "user",
            content = content,
            timestamp = System.currentTimeMillis().toString()
        )

        _uiState.update {
            it.copy(
                messages = it.messages + userMessage,
                inputText = "",
                agentEvents = emptyList(),
                activeToolCalls = emptyList(),
                thinkingContent = ""
            )
        }

        val wsMessage = json.encodeToString(
            ChatSendMessage(
                content = content,
                session_id = _uiState.value.currentSessionId,
                model = _uiState.value.selectedModel.ifEmpty { null }
            )
        )
        webSocketClient.send(wsMessage)

        viewModelScope.launch { _effects.emit(ChatEffect.ScrollToBottom) }
    }

    private fun observeWebSocket() {
        viewModelScope.launch {
            webSocketClient.messages.collect { msg ->
                when (msg) {
                    is WsMessage.StreamStart -> {
                        streamingContent.clear()
                        _uiState.update { it.copy(isStreaming = true) }
                    }

                    is WsMessage.StreamToken -> {
                        streamingContent.append(msg.content)
                        val streamingMsg = ChatMessage(
                            role = "assistant",
                            content = streamingContent.toString(),
                            isStreaming = true
                        )
                        _uiState.update { state ->
                            val msgs = state.messages.toMutableList()
                            val lastIdx = msgs.indexOfLast { it.isStreaming }
                            if (lastIdx >= 0) msgs[lastIdx] = streamingMsg
                            else msgs.add(streamingMsg)
                            state.copy(messages = msgs)
                        }
                        _effects.emit(ChatEffect.ScrollToBottom)
                    }

                    is WsMessage.StreamEnd -> {
                        val finalMsg = ChatMessage(
                            role = "assistant",
                            content = msg.content.ifEmpty { streamingContent.toString() },
                            agents = msg.agents,
                            isStreaming = false
                        )
                        _uiState.update { state ->
                            val msgs = state.messages.toMutableList()
                            val lastIdx = msgs.indexOfLast { it.isStreaming }
                            if (lastIdx >= 0) msgs[lastIdx] = finalMsg
                            else msgs.add(finalMsg)
                            state.copy(
                                messages = msgs,
                                isStreaming = false,
                                activeSpecialist = null,
                                delegationReason = ""
                            )
                        }
                        streamingContent.clear()
                        msg.error?.let { error ->
                            _uiState.update { it.copy(error = error) }
                        }
                    }

                    is WsMessage.AgentStatus -> {
                        val event = AgentEvent(
                            agent = msg.agent,
                            status = msg.status,
                            specialistName = msg.specialistName,
                            reason = msg.reason,
                            delegationChain = msg.delegationChain
                        )
                        _uiState.update {
                            it.copy(
                                agentEvents = it.agentEvents + event,
                                activeSpecialist = msg.specialistName.ifEmpty { null },
                                delegationReason = msg.reason
                            )
                        }
                    }

                    is WsMessage.ToolCallEvent -> {
                        val toolCall = ToolCall(
                            name = msg.name,
                            result = msg.result,
                            status = msg.status
                        )
                        _uiState.update {
                            it.copy(activeToolCalls = it.activeToolCalls + toolCall)
                        }
                    }

                    is WsMessage.ThinkingDelta -> {
                        _uiState.update {
                            it.copy(thinkingContent = it.thinkingContent + msg.content)
                        }
                    }

                    is WsMessage.DelegationUpdate -> {
                        // Delegation tracking
                    }

                    is WsMessage.SpecialistReportEvent -> {
                        // Specialist report tracking
                    }

                    is WsMessage.Ready -> {
                        _uiState.update { it.copy(isStreaming = false) }
                    }

                    is WsMessage.Error -> {
                        _uiState.update {
                            it.copy(error = msg.message, isStreaming = false)
                        }
                    }
                }
            }
        }
    }

    private fun observeConnectionState() {
        viewModelScope.launch {
            webSocketClient.isConnected.collect { connected ->
                _uiState.update { it.copy(isConnected = connected) }
            }
        }
    }

    private fun loadModels() {
        viewModelScope.launch {
            try {
                val response = configApi.getModels()
                val models = response.providers
                    .filter { it.enabled }
                    .flatMap { p -> p.models.map { "${p.name}/${it.id}" } }
                _uiState.update {
                    it.copy(
                        availableModels = models,
                        selectedModel = response.current_model
                    )
                }
            } catch (_: Exception) { }
        }
    }

    private fun loadSessions() {
        viewModelScope.launch {
            try {
                val response = chatApi.getSessions()
                _uiState.update { it.copy(sessions = response.sessions) }
            } catch (_: Exception) { }
        }
    }

    private fun loadSession(sessionId: String) {
        viewModelScope.launch {
            try {
                val response = chatApi.getSessionMessages(sessionId)
                _uiState.update {
                    it.copy(
                        messages = response.messages,
                        currentSessionId = sessionId
                    )
                }
            } catch (_: Exception) { }
        }
    }

    private fun newSession() {
        _uiState.update {
            it.copy(
                messages = emptyList(),
                currentSessionId = "web:chat-${System.currentTimeMillis()}",
                agentEvents = emptyList(),
                activeToolCalls = emptyList(),
                thinkingContent = ""
            )
        }
    }

    private fun cancelExecution() {
        webSocketClient.send("stop")
        _uiState.update { it.copy(isStreaming = false) }
    }

    override fun onCleared() {
        super.onCleared()
        webSocketClient.disconnect()
    }
}
