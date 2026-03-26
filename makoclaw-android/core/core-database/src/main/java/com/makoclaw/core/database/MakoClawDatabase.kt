package com.makoclaw.core.database

import androidx.room.Database
import androidx.room.RoomDatabase
import com.makoclaw.core.database.dao.ChatMessageDao
import com.makoclaw.core.database.dao.ChatSessionDao
import com.makoclaw.core.database.dao.CronJobDao
import com.makoclaw.core.database.dao.FileEntryDao
import com.makoclaw.core.database.dao.KnowledgeDocumentDao
import com.makoclaw.core.database.dao.McpServerDao
import com.makoclaw.core.database.dao.MemoryDao
import com.makoclaw.core.database.dao.ReportDao
import com.makoclaw.core.database.dao.SessionDao
import com.makoclaw.core.database.dao.SkillDao
import com.makoclaw.core.database.dao.TaskDao
import com.makoclaw.core.database.dao.WorkflowDao
import com.makoclaw.core.database.entity.AgentEntity
import com.makoclaw.core.database.entity.ChatMessageEntity
import com.makoclaw.core.database.entity.ChatSessionEntity
import com.makoclaw.core.database.entity.CronJobEntity
import com.makoclaw.core.database.entity.FileEntryEntity
import com.makoclaw.core.database.entity.KnowledgeDocumentEntity
import com.makoclaw.core.database.entity.McpServerEntity
import com.makoclaw.core.database.entity.MemoryEntity
import com.makoclaw.core.database.entity.ReportEntity
import com.makoclaw.core.database.entity.SessionEntity
import com.makoclaw.core.database.entity.SkillEntity
import com.makoclaw.core.database.entity.TaskEntity
import com.makoclaw.core.database.entity.WorkflowEntity

@Database(
    entities = [
        ChatSessionEntity::class,
        ChatMessageEntity::class,
        TaskEntity::class,
        AgentEntity::class,
        KnowledgeDocumentEntity::class,
        SkillEntity::class,
        WorkflowEntity::class,
        CronJobEntity::class,
        FileEntryEntity::class,
        MemoryEntity::class,
        McpServerEntity::class,
        SessionEntity::class,
        ReportEntity::class
    ],
    version = 2,
    exportSchema = false
)
abstract class MakoClawDatabase : RoomDatabase() {
    abstract fun chatSessionDao(): ChatSessionDao
    abstract fun chatMessageDao(): ChatMessageDao
    abstract fun taskDao(): TaskDao
    abstract fun knowledgeDocumentDao(): KnowledgeDocumentDao
    abstract fun skillDao(): SkillDao
    abstract fun workflowDao(): WorkflowDao
    abstract fun cronJobDao(): CronJobDao
    abstract fun fileEntryDao(): FileEntryDao
    abstract fun memoryDao(): MemoryDao
    abstract fun mcpServerDao(): McpServerDao
    abstract fun sessionDao(): SessionDao
    abstract fun reportDao(): ReportDao
}
