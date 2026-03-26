package com.makoclaw.core.database.entity

import androidx.room.Entity
import androidx.room.ForeignKey
import androidx.room.Index
import androidx.room.PrimaryKey

@Entity(tableName = "chat_sessions")
data class ChatSessionEntity(
    @PrimaryKey val sessionId: String,
    val title: String = "",
    val lastMessage: String = "",
    val messageCount: Int = 0,
    val updatedAt: String = "",
    val archived: Boolean = false
)

@Entity(
    tableName = "chat_messages",
    foreignKeys = [
        ForeignKey(
            entity = ChatSessionEntity::class,
            parentColumns = ["sessionId"],
            childColumns = ["sessionId"],
            onDelete = ForeignKey.CASCADE
        )
    ],
    indices = [Index("sessionId")]
)
data class ChatMessageEntity(
    @PrimaryKey(autoGenerate = true) val id: Long = 0,
    val sessionId: String,
    val role: String,
    val content: String,
    val timestamp: String = "",
    val agentsJson: String? = null,
    val toolCallsJson: String? = null
)

@Entity(tableName = "tasks")
data class TaskEntity(
    @PrimaryKey val id: String,
    val title: String,
    val description: String = "",
    val status: String = "backlog",
    val result: String = "",
    val createdAt: String = "",
    val updatedAt: String = "",
    val archived: Boolean = false
)

@Entity(tableName = "agents")
data class AgentEntity(
    @PrimaryKey val name: String,
    val configJson: String = "",
    val isOrchestrator: Boolean = false
)
