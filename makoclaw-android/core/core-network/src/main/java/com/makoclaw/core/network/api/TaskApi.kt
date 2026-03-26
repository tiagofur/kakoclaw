package com.makoclaw.core.network.api

import com.makoclaw.core.model.CreateTaskRequest
import com.makoclaw.core.model.Task
import com.makoclaw.core.model.UpdateTaskStatusRequest
import kotlinx.serialization.Serializable
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.PATCH
import retrofit2.http.POST
import retrofit2.http.PUT
import retrofit2.http.Path
import retrofit2.http.Query

@Serializable
data class TasksResponse(val tasks: List<Task> = emptyList())

interface TaskApi {

    @GET("tasks")
    suspend fun getTasks(
        @Query("include_archived") includeArchived: Boolean = false
    ): TasksResponse

    @POST("tasks")
    suspend fun createTask(@Body request: CreateTaskRequest): Task

    @PUT("tasks/{id}")
    suspend fun updateTask(@Path("id") id: String, @Body task: Task): Task

    @DELETE("tasks/{id}")
    suspend fun deleteTask(@Path("id") id: String)

    @PATCH("tasks/{id}/status")
    suspend fun updateTaskStatus(
        @Path("id") id: String,
        @Body request: UpdateTaskStatusRequest
    )

    @POST("tasks/{id}/archive")
    suspend fun archiveTask(@Path("id") id: String)

    @POST("tasks/{id}/unarchive")
    suspend fun unarchiveTask(@Path("id") id: String)

    @GET("tasks/search")
    suspend fun searchTasks(@Query("q") query: String): TasksResponse
}
