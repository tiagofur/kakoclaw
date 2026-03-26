package com.makoclaw.core.model

import kotlinx.serialization.Serializable

@Serializable
data class KnowledgeDocument(
    val id: Long = 0,
    val filename: String,
    val chunkCount: Int = 0,
    val createdAt: String = ""
)

@Serializable
data class KnowledgeChunk(
    val id: Long = 0,
    val documentId: Long = 0,
    val content: String = "",
    val metadata: String = ""
)

@Serializable
data class Skill(
    val name: String,
    val content: String = "",
    val repository: String = ""
)

@Serializable
data class CronJob(
    val id: String = "",
    val name: String = "",
    val schedule: String = "",
    val message: String = "",
    val enabled: Boolean = true,
    val nextRun: String = ""
)

@Serializable
data class Workflow(
    val name: String,
    val description: String = "",
    val nodes: List<String> = emptyList(),
    val createdAt: String = ""
)

@Serializable
data class FileEntry(
    val name: String,
    val path: String = "",
    val isDir: Boolean = false,
    val size: Long = 0,
    val modTime: String = ""
)

@Serializable
data class MemoryContent(
    val content: String = ""
)

@Serializable
data class DailyNote(
    val date: String,
    val content: String
)

@Serializable
data class McpServer(
    val name: String,
    val status: String = "disconnected",
    val tools: List<String> = emptyList()
)
