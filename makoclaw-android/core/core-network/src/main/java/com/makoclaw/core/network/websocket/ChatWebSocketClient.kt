package com.makoclaw.core.network.websocket

import com.makoclaw.core.common.Constants
import com.makoclaw.core.security.TokenManager
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonPrimitive
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import javax.inject.Inject
import javax.inject.Singleton

sealed interface WsMessage {
    data class StreamStart(val id: String = "") : WsMessage
    data class StreamToken(val content: String) : WsMessage
    data class StreamEnd(
        val content: String,
        val agents: List<String> = emptyList(),
        val error: String? = null,
        val canceled: Int = 0
    ) : WsMessage
    data class AgentStatus(
        val agent: String,
        val status: String,
        val specialistName: String = "",
        val reason: String = "",
        val delegationChain: List<String> = emptyList()
    ) : WsMessage
    data class ToolCallEvent(
        val name: String,
        val args: String = "",
        val result: String? = null,
        val status: String = "pending"
    ) : WsMessage
    data class ThinkingDelta(val content: String) : WsMessage
    data class DelegationUpdate(
        val from: String,
        val to: String,
        val status: String
    ) : WsMessage
    data class SpecialistReportEvent(
        val specialistName: String,
        val status: String,
        val confidence: Double = 0.0
    ) : WsMessage
    data class Ready(val id: String = "") : WsMessage
    data class Error(val message: String) : WsMessage
}

@Singleton
class ChatWebSocketClient @Inject constructor(
    private val okHttpClient: OkHttpClient,
    private val tokenManager: TokenManager
) {
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private val json = Json { ignoreUnknownKeys = true }

    private var webSocket: WebSocket? = null
    private var reconnectAttempts = 0
    private var serverUrl: String = Constants.DEFAULT_SERVER_URL

    private val _messages = MutableSharedFlow<WsMessage>(extraBufferCapacity = 64)
    val messages: SharedFlow<WsMessage> = _messages.asSharedFlow()

    private val _isConnected = MutableStateFlow(false)
    val isConnected: StateFlow<Boolean> = _isConnected.asStateFlow()

    fun setServerUrl(url: String) {
        serverUrl = url
    }

    fun connect() {
        val token = tokenManager.getToken() ?: return
        val wsUrl = serverUrl
            .replace("http://", "ws://")
            .replace("https://", "wss://") +
            "${Constants.WS_CHAT_PATH}?token=$token"

        val request = Request.Builder().url(wsUrl).build()

        webSocket = okHttpClient.newWebSocket(request, object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                _isConnected.value = true
                reconnectAttempts = 0
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                parseAndEmit(text)
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                _isConnected.value = false
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                _isConnected.value = false
                scope.launch { _messages.emit(WsMessage.Error(t.message ?: "Connection failed")) }
                attemptReconnect()
            }
        })
    }

    fun send(message: String) {
        webSocket?.send(message)
    }

    fun disconnect() {
        reconnectAttempts = Constants.WS_MAX_RECONNECT_ATTEMPTS // prevent reconnection
        webSocket?.close(1000, "User disconnect")
        webSocket = null
        _isConnected.value = false
    }

    private fun attemptReconnect() {
        if (reconnectAttempts >= Constants.WS_MAX_RECONNECT_ATTEMPTS) return
        reconnectAttempts++
        val backoff = Constants.WS_INITIAL_BACKOFF_MS * (1L shl (reconnectAttempts - 1))
        scope.launch {
            delay(backoff)
            if (!_isConnected.value) connect()
        }
    }

    private fun parseAndEmit(text: String) {
        scope.launch {
            try {
                val obj = json.parseToJsonElement(text) as? JsonObject ?: return@launch
                val type = obj["type"]?.jsonPrimitive?.contentOrNull ?: return@launch

                val msg = when (type) {
                    "stream_start" -> WsMessage.StreamStart()
                    "stream" -> WsMessage.StreamToken(
                        content = obj["content"]?.jsonPrimitive?.contentOrNull ?: ""
                    )
                    "stream_end" -> WsMessage.StreamEnd(
                        content = obj["content"]?.jsonPrimitive?.contentOrNull ?: "",
                        error = obj["error"]?.jsonPrimitive?.contentOrNull
                    )
                    "agent_status" -> WsMessage.AgentStatus(
                        agent = obj["agent"]?.jsonPrimitive?.contentOrNull ?: "",
                        status = obj["status"]?.jsonPrimitive?.contentOrNull ?: "",
                        specialistName = obj["specialist_name"]?.jsonPrimitive?.contentOrNull ?: "",
                        reason = obj["reason"]?.jsonPrimitive?.contentOrNull ?: ""
                    )
                    "tool_call" -> WsMessage.ToolCallEvent(
                        name = obj["name"]?.jsonPrimitive?.contentOrNull ?: "",
                        result = obj["result"]?.jsonPrimitive?.contentOrNull,
                        status = obj["status"]?.jsonPrimitive?.contentOrNull ?: "pending"
                    )
                    "thinking_delta" -> WsMessage.ThinkingDelta(
                        content = obj["content"]?.jsonPrimitive?.contentOrNull ?: ""
                    )
                    "delegation_update" -> WsMessage.DelegationUpdate(
                        from = obj["from"]?.jsonPrimitive?.contentOrNull ?: "",
                        to = obj["to"]?.jsonPrimitive?.contentOrNull ?: "",
                        status = obj["status"]?.jsonPrimitive?.contentOrNull ?: ""
                    )
                    "specialist_report" -> WsMessage.SpecialistReportEvent(
                        specialistName = obj["specialist_name"]?.jsonPrimitive?.contentOrNull ?: "",
                        status = obj["status"]?.jsonPrimitive?.contentOrNull ?: ""
                    )
                    "ready" -> WsMessage.Ready()
                    "error" -> WsMessage.Error(
                        message = obj["content"]?.jsonPrimitive?.contentOrNull ?: "Unknown error"
                    )
                    else -> null
                }

                msg?.let { _messages.emit(it) }
            } catch (_: Exception) {
                // Silently ignore malformed messages
            }
        }
    }
}
