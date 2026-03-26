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

sealed interface TaskWsEvent {
    data class TaskCreated(val taskId: String) : TaskWsEvent
    data class TaskUpdated(val taskId: String) : TaskWsEvent
    data class TaskDeleted(val taskId: String) : TaskWsEvent
    data class TaskStatusChanged(val taskId: String, val newStatus: String) : TaskWsEvent
}

@Singleton
class TaskWebSocketClient @Inject constructor(
    private val okHttpClient: OkHttpClient,
    private val tokenManager: TokenManager
) {
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private val json = Json { ignoreUnknownKeys = true }

    private var webSocket: WebSocket? = null
    private var reconnectAttempts = 0
    private var serverUrl: String = Constants.DEFAULT_SERVER_URL

    private val _events = MutableSharedFlow<TaskWsEvent>(extraBufferCapacity = 32)
    val events: SharedFlow<TaskWsEvent> = _events.asSharedFlow()

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
            "${Constants.WS_TASKS_PATH}?token=$token"

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
                attemptReconnect()
            }
        })
    }

    fun disconnect() {
        reconnectAttempts = Constants.WS_MAX_RECONNECT_ATTEMPTS
        webSocket?.close(1000, "Disconnect")
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
                val taskId = obj["task_id"]?.jsonPrimitive?.contentOrNull ?: ""

                val event = when (type) {
                    "task_created" -> TaskWsEvent.TaskCreated(taskId)
                    "task_updated" -> TaskWsEvent.TaskUpdated(taskId)
                    "task_deleted" -> TaskWsEvent.TaskDeleted(taskId)
                    "task_status_changed" -> TaskWsEvent.TaskStatusChanged(
                        taskId,
                        obj["status"]?.jsonPrimitive?.contentOrNull ?: ""
                    )
                    else -> null
                }
                event?.let { _events.emit(it) }
            } catch (_: Exception) { }
        }
    }
}
