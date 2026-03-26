package com.makoclaw.core.common

object Constants {
    const val DEFAULT_SERVER_URL = "http://localhost:8080"
    const val API_BASE_PATH = "/api/v1"
    const val WS_CHAT_PATH = "/ws/chat"
    const val WS_TASKS_PATH = "/ws/tasks"
    const val HTTP_TIMEOUT_SECONDS = 30L
    const val WS_MAX_RECONNECT_ATTEMPTS = 5
    const val WS_INITIAL_BACKOFF_MS = 1000L
    const val JWT_TOKEN_KEY = "jwt_token"
    const val AUTH_HEADER = "Authorization"
    const val BEARER_PREFIX = "Bearer "
}
