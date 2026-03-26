package com.makoclaw.core.database.di

import android.content.Context
import androidx.room.Room
import com.makoclaw.core.database.MakoClawDatabase
import com.makoclaw.core.database.dao.ChatMessageDao
import com.makoclaw.core.database.dao.ChatSessionDao
import com.makoclaw.core.database.dao.TaskDao
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
    ).build()

    @Provides
    fun provideChatSessionDao(db: MakoClawDatabase): ChatSessionDao = db.chatSessionDao()

    @Provides
    fun provideChatMessageDao(db: MakoClawDatabase): ChatMessageDao = db.chatMessageDao()

    @Provides
    fun provideTaskDao(db: MakoClawDatabase): TaskDao = db.taskDao()
}
