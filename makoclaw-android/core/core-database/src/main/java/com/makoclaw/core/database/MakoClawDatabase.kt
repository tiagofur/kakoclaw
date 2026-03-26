package com.makoclaw.core.database

import androidx.room.Database
import androidx.room.RoomDatabase
import com.makoclaw.core.database.dao.ChatMessageDao
import com.makoclaw.core.database.dao.ChatSessionDao
import com.makoclaw.core.database.dao.TaskDao
import com.makoclaw.core.database.entity.AgentEntity
import com.makoclaw.core.database.entity.ChatMessageEntity
import com.makoclaw.core.database.entity.ChatSessionEntity
import com.makoclaw.core.database.entity.TaskEntity

@Database(
    entities = [
        ChatSessionEntity::class,
        ChatMessageEntity::class,
        TaskEntity::class,
        AgentEntity::class
    ],
    version = 1,
    exportSchema = false
)
abstract class MakoClawDatabase : RoomDatabase() {
    abstract fun chatSessionDao(): ChatSessionDao
    abstract fun chatMessageDao(): ChatMessageDao
    abstract fun taskDao(): TaskDao
}
