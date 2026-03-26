package com.makoclaw.feature.chat.data.repository

import com.makoclaw.core.common.Result
import com.makoclaw.core.database.dao.ChatMessageDao
import com.makoclaw.core.database.dao.ChatSessionDao
import com.makoclaw.core.database.entity.ChatMessageEntity
import com.makoclaw.core.database.entity.ChatSessionEntity
import com.makoclaw.core.model.ChatMessage
import com.makoclaw.core.model.ChatSession
import com.makoclaw.core.network.api.ChatApi
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ChatRepository @Inject constructor(
    private val chatApi: ChatApi,
    private val sessionDao: ChatSessionDao,
    private val messageDao: ChatMessageDao
) {
    suspend fun getSessions(): Result<List<ChatSession>> {
        return try {
            val response = chatApi.getSessions()
            // Cache in Room
            sessionDao.insertSessions(response.sessions.map { it.toEntity() })
            Result.Success(response.sessions)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Failed to load sessions", e)
        }
    }

    fun getCachedSessions(): Flow<List<ChatSession>> =
        sessionDao.getAllSessions().map { entities ->
            entities.map { it.toDomain() }
        }

    suspend fun getMessages(sessionId: String): Result<List<ChatMessage>> {
        return try {
            val response = chatApi.getSessionMessages(sessionId)
            // Cache in Room
            messageDao.insertMessages(response.messages.map { it.toEntity(sessionId) })
            Result.Success(response.messages)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Failed to load messages", e)
        }
    }

    fun getCachedMessages(sessionId: String): Flow<List<ChatMessage>> =
        messageDao.getMessagesForSession(sessionId).map { entities ->
            entities.map { it.toDomain() }
        }

    suspend fun deleteSession(sessionId: String): Result<Unit> {
        return try {
            chatApi.deleteSession(sessionId)
            sessionDao.deleteSession(sessionId)
            Result.Success(Unit)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Failed to delete session", e)
        }
    }

    suspend fun searchMessages(query: String): Result<List<ChatMessage>> {
        return try {
            val response = chatApi.searchMessages(query)
            Result.Success(response.results)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Search failed", e)
        }
    }

    suspend fun forkConversation(sessionId: String, messageIndex: Int): Result<Unit> {
        return try {
            chatApi.forkConversation(
                com.makoclaw.core.model.ForkRequest(sessionId, messageIndex)
            )
            Result.Success(Unit)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Fork failed", e)
        }
    }

    suspend fun cancelExecution(): Result<Unit> {
        return try {
            chatApi.cancelExecution()
            Result.Success(Unit)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Cancel failed", e)
        }
    }

    // Mappers
    private fun ChatSession.toEntity() = ChatSessionEntity(
        sessionId = sessionId,
        title = title,
        lastMessage = lastMessage,
        messageCount = messageCount,
        updatedAt = updatedAt,
        archived = archived
    )

    private fun ChatSessionEntity.toDomain() = ChatSession(
        sessionId = sessionId,
        title = title,
        lastMessage = lastMessage,
        messageCount = messageCount,
        updatedAt = updatedAt,
        archived = archived
    )

    private fun ChatMessage.toEntity(sessId: String) = ChatMessageEntity(
        sessionId = sessId,
        role = role,
        content = content,
        timestamp = timestamp
    )

    private fun ChatMessageEntity.toDomain() = ChatMessage(
        id = id.toString(),
        role = role,
        content = content,
        timestamp = timestamp
    )
}
