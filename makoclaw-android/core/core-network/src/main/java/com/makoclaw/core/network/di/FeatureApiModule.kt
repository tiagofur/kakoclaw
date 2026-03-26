package com.makoclaw.core.network.di

import com.makoclaw.core.network.api.CronApi
import com.makoclaw.core.network.api.FilesApi
import com.makoclaw.core.network.api.KnowledgeApi
import com.makoclaw.core.network.api.McpApi
import com.makoclaw.core.network.api.MemoryApi
import com.makoclaw.core.network.api.MetricsApi
import com.makoclaw.core.network.api.ReportsApi
import com.makoclaw.core.network.api.SkillsApi
import com.makoclaw.core.network.api.VoiceApi
import com.makoclaw.core.network.api.WorkflowsApi
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import retrofit2.Retrofit
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object FeatureApiModule {

    @Provides @Singleton
    fun provideKnowledgeApi(retrofit: Retrofit): KnowledgeApi =
        retrofit.create(KnowledgeApi::class.java)

    @Provides @Singleton
    fun provideSkillsApi(retrofit: Retrofit): SkillsApi =
        retrofit.create(SkillsApi::class.java)

    @Provides @Singleton
    fun provideCronApi(retrofit: Retrofit): CronApi =
        retrofit.create(CronApi::class.java)

    @Provides @Singleton
    fun provideFilesApi(retrofit: Retrofit): FilesApi =
        retrofit.create(FilesApi::class.java)

    @Provides @Singleton
    fun provideMemoryApi(retrofit: Retrofit): MemoryApi =
        retrofit.create(MemoryApi::class.java)

    @Provides @Singleton
    fun provideMcpApi(retrofit: Retrofit): McpApi =
        retrofit.create(McpApi::class.java)

    @Provides @Singleton
    fun provideWorkflowsApi(retrofit: Retrofit): WorkflowsApi =
        retrofit.create(WorkflowsApi::class.java)

    @Provides @Singleton
    fun provideMetricsApi(retrofit: Retrofit): MetricsApi =
        retrofit.create(MetricsApi::class.java)

    @Provides @Singleton
    fun provideVoiceApi(retrofit: Retrofit): VoiceApi =
        retrofit.create(VoiceApi::class.java)

    @Provides @Singleton
    fun provideReportsApi(retrofit: Retrofit): ReportsApi =
        retrofit.create(ReportsApi::class.java)
}
