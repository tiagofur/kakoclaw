package com.makoclaw.core.model

import kotlinx.serialization.Serializable

@Serializable
data class Task(
    val id: String = "",
    val title: String,
    val description: String = "",
    val status: TaskStatus = TaskStatus.BACKLOG,
    val result: String = "",
    val createdAt: String = "",
    val updatedAt: String = "",
    val archived: Boolean = false
)

@Serializable
enum class TaskStatus(val value: String) {
    BACKLOG("backlog"),
    TODO("todo"),
    IN_PROGRESS("in_progress"),
    REVIEW("review"),
    DONE("done");

    companion object {
        fun fromValue(value: String): TaskStatus =
            entries.find { it.value == value } ?: BACKLOG
    }
}

@Serializable
data class CreateTaskRequest(
    val title: String,
    val description: String = "",
    val status: String = "backlog"
)

@Serializable
data class UpdateTaskStatusRequest(
    val status: String
)
