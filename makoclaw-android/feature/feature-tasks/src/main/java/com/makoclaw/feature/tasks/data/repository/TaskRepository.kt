package com.makoclaw.feature.tasks.data.repository

import com.makoclaw.core.common.Result
import com.makoclaw.core.database.dao.TaskDao
import com.makoclaw.core.database.entity.TaskEntity
import com.makoclaw.core.model.CreateTaskRequest
import com.makoclaw.core.model.Task
import com.makoclaw.core.model.TaskStatus
import com.makoclaw.core.model.UpdateTaskStatusRequest
import com.makoclaw.core.network.api.TaskApi
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TaskRepository @Inject constructor(
    private val taskApi: TaskApi,
    private val taskDao: TaskDao
) {
    suspend fun getTasks(): Result<List<Task>> {
        return try {
            val response = taskApi.getTasks()
            taskDao.insertTasks(response.tasks.map { it.toEntity() })
            Result.Success(response.tasks)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Failed to load tasks", e)
        }
    }

    fun getCachedTasks(): Flow<List<Task>> =
        taskDao.getAllTasks().map { entities -> entities.map { it.toDomain() } }

    fun getCachedTasksByStatus(status: TaskStatus): Flow<List<Task>> =
        taskDao.getTasksByStatus(status.value).map { entities ->
            entities.map { it.toDomain() }
        }

    suspend fun createTask(title: String, description: String): Result<Task> {
        return try {
            val task = taskApi.createTask(CreateTaskRequest(title, description))
            taskDao.insertTask(task.toEntity())
            Result.Success(task)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Failed to create task", e)
        }
    }

    suspend fun updateTaskStatus(taskId: String, status: TaskStatus): Result<Unit> {
        return try {
            // Optimistic update
            taskDao.updateTaskStatus(taskId, status.value)
            taskApi.updateTaskStatus(taskId, UpdateTaskStatusRequest(status.value))
            Result.Success(Unit)
        } catch (e: Exception) {
            // Revert on failure
            getTasks()
            Result.Error(e.message ?: "Failed to update status", e)
        }
    }

    suspend fun deleteTask(taskId: String): Result<Unit> {
        return try {
            taskApi.deleteTask(taskId)
            taskDao.deleteTask(taskId)
            Result.Success(Unit)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Failed to delete task", e)
        }
    }

    suspend fun archiveTask(taskId: String): Result<Unit> {
        return try {
            taskApi.archiveTask(taskId)
            taskDao.deleteTask(taskId)
            Result.Success(Unit)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Failed to archive task", e)
        }
    }

    suspend fun searchTasks(query: String): Result<List<Task>> {
        return try {
            val response = taskApi.searchTasks(query)
            Result.Success(response.tasks)
        } catch (e: Exception) {
            Result.Error(e.message ?: "Search failed", e)
        }
    }

    private fun Task.toEntity() = TaskEntity(
        id = id, title = title, description = description,
        status = status.value, result = result,
        createdAt = createdAt, updatedAt = updatedAt, archived = archived
    )

    private fun TaskEntity.toDomain() = Task(
        id = id, title = title, description = description,
        status = TaskStatus.fromValue(status), result = result,
        createdAt = createdAt, updatedAt = updatedAt, archived = archived
    )
}
