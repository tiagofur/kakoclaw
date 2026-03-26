package com.makoclaw.core.database.dao

import androidx.room.Room
import androidx.test.core.app.ApplicationProvider
import com.makoclaw.core.database.MakoClawDatabase
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach

abstract class BaseDaoTest {
    protected lateinit var database: MakoClawDatabase

    @BeforeEach
    fun setUpDatabase() {
        database = Room.inMemoryDatabaseBuilder(
            ApplicationProvider.getApplicationContext(),
            MakoClawDatabase::class.java
        )
            .allowMainThreadQueries()
            .build()
    }

    @AfterEach
    fun tearDownDatabase() {
        database.close()
    }
}
