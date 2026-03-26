package com.makoclaw.core.network.api

import com.makoclaw.core.model.ChatSession
import com.makoclaw.core.model.ChatMessage
import com.makoclaw.core.model.ForkRequest
import kotlinx.serialization.Serializable
import okhttp3.MultipartBody
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.Multipart
import retrofit2.http.PATCH
import retrofit2.http.POST
import retrofit2.http.Part
import retrofit2.http.Path
import retrofit2.http.Query

@Serializable
data class SessionsResponse(val sessions: List<ChatSession> = emptyList())

@Serializable
data class MessagesResponse(val messages: List<ChatMessage> = emptyList())

@Serializable
data class SearchResponse(val results: List<ChatMessage> = emptyList())

@Serializable
data class ActiveExecutionsResponse(val executions: List<String> = emptyList())

interface ChatApi {

    @GET("chat/sessions")
    suspend fun getSessions(): SessionsResponse

    @GET("chat/sessions/{id}")
    suspend fun getSessionMessages(@Path("id") sessionId: String): MessagesResponse

    @DELETE("chat/sessions/{id}")
    suspend fun deleteSession(@Path("id") sessionId: String)

    @PATCH("chat/sessions/{id}")
    suspend fun updateSession(
        @Path("id") sessionId: String,
        @Body updates: Map<String, String>
    )

    @GET("chat/search")
    suspend fun searchMessages(@Query("q") query: String): SearchResponse

    @POST("chat/fork")
    suspend fun forkConversation(@Body request: ForkRequest)

    @POST("chat/cancel")
    suspend fun cancelExecution()

    @GET("chat/active")
    suspend fun getActiveExecutions(): ActiveExecutionsResponse

    @Multipart
    @POST("chat/attachments")
    suspend fun uploadAttachment(@Part file: MultipartBody.Part): Map<String, String>
}
