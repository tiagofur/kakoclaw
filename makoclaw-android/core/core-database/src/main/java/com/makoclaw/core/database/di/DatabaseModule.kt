package com.makoclaw.core.database.di

import android.content.Context
import androidx.room.Room
import com.makoclaw.core.database.MakoClawDatabase
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
import com.makoclaw.core.database.migration.DatabaseMigration1to2
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object DatabaseModule {

    @Provides
    @Singleton
    fun provideDatabase(
        @ApplicationContext context: Context
    ): MakoClawDatabase = Room.databaseBuilder(
        context,
        MakoClawDatabase::class.java,
        "makoclaw.db"
    )
        .addMigrations(DatabaseMigration1to2)
        .build()

    @Provides
    fun provideChatSessionDao(db: MakoClawDatabase): ChatSessionDao = db.chatSessionDao()

    @Provides
    fun provideChatMessageDao(db: MakoClawDatabase): ChatMessageDao = db.chatMessageDao()

    @Provides
    fun provideTaskDao(db: MakoClawDatabase): TaskDao = db.taskDao()

    @Provides
    fun provideKnowledgeDocumentDao(db: MakoClawDatabase): KnowledgeDocumentDao = db.knowledgeDocumentDao()

    @Provides
    fun provideSkillDao(db: MakoClawDatabase): SkillDao = db.skillDao()

    @Provides
    fun provideWorkflowDao(db: MakoClawDatabase): WorkflowDao = db.workflowDao()

    @Provides
    fun provideCronJobDao(db: MakoClawDatabase): CronJobDao = db.cronJobDao()

    @Provides
    fun provideFileEntryDao(db: MakoClawDatabase): FileEntryDao = db.fileEntryDao()

    @Provides
    fun provideMemoryDao(db: MakoClawDatabase): MemoryDao = db.memoryDao()

    @Provides
    fun provideMcpServerDao(db: MakoClawDatabase): McpServerDao = db.mcpServerDao()

    @Provides
    fun provideSessionDao(db: MakoClawDatabase): SessionDao = db.sessionDao()

    @Provides
    fun provideReportDao(db: MakoClawDatabase): ReportDao = db.reportDao()
}
