package com.makoclaw.core.database.migration

import androidx.room.Room
import androidx.sqlite.db.SupportSQLiteOpenHelper
import androidx.sqlite.db.framework.FrameworkSQLiteOpenHelperFactory
import androidx.test.core.app.ApplicationProvider
import com.makoclaw.core.database.MakoClawDatabase
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test

class DatabaseMigration1to2Test {
    @Test
    fun `migration from version 1 to 2 creates new feature tables`() {
        val context = ApplicationProvider.getApplicationContext<android.content.Context>()
        val databaseName = "migration-1-2-test.db"
        context.deleteDatabase(databaseName)

        val callback = object : SupportSQLiteOpenHelper.Callback(1) {
            override fun onCreate(db: androidx.sqlite.db.SupportSQLiteDatabase) {
                db.execSQL("CREATE TABLE IF NOT EXISTS chat_sessions (sessionId TEXT NOT NULL, title TEXT NOT NULL, lastMessage TEXT NOT NULL, messageCount INTEGER NOT NULL, updatedAt TEXT NOT NULL, archived INTEGER NOT NULL, PRIMARY KEY(sessionId))")
                db.execSQL("CREATE TABLE IF NOT EXISTS chat_messages (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, sessionId TEXT NOT NULL, role TEXT NOT NULL, content TEXT NOT NULL, timestamp TEXT NOT NULL, agentsJson TEXT, toolCallsJson TEXT)")
                db.execSQL("CREATE TABLE IF NOT EXISTS tasks (id TEXT NOT NULL, title TEXT NOT NULL, description TEXT NOT NULL, status TEXT NOT NULL, result TEXT NOT NULL, createdAt TEXT NOT NULL, updatedAt TEXT NOT NULL, archived INTEGER NOT NULL, PRIMARY KEY(id))")
                db.execSQL("CREATE TABLE IF NOT EXISTS agents (name TEXT NOT NULL, configJson TEXT NOT NULL, isOrchestrator INTEGER NOT NULL, PRIMARY KEY(name))")
            }

            override fun onUpgrade(
                db: androidx.sqlite.db.SupportSQLiteDatabase,
                oldVersion: Int,
                newVersion: Int
            ) = Unit
        }

        FrameworkSQLiteOpenHelperFactory()
            .create(
                SupportSQLiteOpenHelper.Configuration.builder(context)
                    .name(databaseName)
                    .callback(callback)
                    .build()
            )
            .apply {
                writableDatabase.close()
                close()
            }

        val migratedDatabase = Room.databaseBuilder(context, MakoClawDatabase::class.java, databaseName)
            .addMigrations(DatabaseMigration1to2)
            .build()

        val tables = buildSet {
            migratedDatabase.openHelper.writableDatabase
                .query("SELECT name FROM sqlite_master WHERE type='table'")
                .use { cursor ->
                    while (cursor.moveToNext()) {
                        add(cursor.getString(0))
                    }
                }
        }

        assertTrue(tables.containsAll(listOf(
            "knowledge_documents",
            "skills",
            "workflows",
            "cron_jobs",
            "file_entries",
            "memories",
            "mcp_servers",
            "sessions",
            "reports"
        )))

        migratedDatabase.close()
        context.deleteDatabase(databaseName)
    }
}
