package com.makoclaw.core.security.di

import android.content.Context
import com.makoclaw.core.security.TokenManager
import com.makoclaw.core.security.storage.JwtStorage
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object SecurityModule {

    @Provides
    @Singleton
    fun provideJwtStorage(
        @ApplicationContext context: Context
    ): JwtStorage = JwtStorage(context)

    @Provides
    @Singleton
    fun provideTokenManager(jwtStorage: JwtStorage): TokenManager = TokenManager(jwtStorage)
}
