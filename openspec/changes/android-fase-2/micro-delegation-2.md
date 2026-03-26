# Micro-Delegación 2: WorkflowsRepository

## Tarea: TASK-WORKFLOWS-002
Crear WorkflowsRepository con interface, implementación, DAO y API.

## Archivos a Crear (3 archivos)

### Archivo 1: WorkflowsRepository.kt
**Ubicación**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/repository/WorkflowsRepository.kt`

```kotlin
package com.makoclaw.feature.workflows.data.repository

import com.makoclaw.core.model.Workflow
import com.makoclaw.core.model.WorkflowExecutionLog
import com.makoclaw.core.network.api.WorkflowsApi
import com.makoclaw.core.network.api.ApiResponse
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.catch
import javax.inject.Inject

interface WorkflowsRepository {
    fun getWorkflows(): Flow<List<Workflow>>
    fun createWorkflow(workflow: Workflow): Flow<ApiResponse>
    fun updateWorkflow(workflow: Workflow): Flow<ApiResponse>
    fun deleteWorkflow(id: String): Flow<ApiResponse>
    fun executeWorkflow(id: String): Flow<WorkflowExecutionLog>
    fun getExecutionLogs(workflowId: String): Flow<List<WorkflowExecutionLog>>
}

@ViewModelScoped
class WorkflowsRepositoryImpl @Inject constructor(
    private val api: WorkflowsApi,
    private val dao: WorkflowDao
) : WorkflowsRepository {

    override fun getWorkflows(): Flow<List<Workflow>> {
        return dao.getAll()
            .map { entities -> entities.map { it.toModel() } }
            .catch { e ->
                // Fallback to API if local cache fails
                api.getWorkflows()
                    .map { response -> response.workflows }
            }
    }

    override fun createWorkflow(workflow: Workflow): Flow<ApiResponse> {
        return api.createWorkflow(
            com.makoclaw.core.network.api.CreateWorkflowRequest(
                name = workflow.name,
                description = workflow.description,
                nodes = workflow.nodes,
                edges = workflow.edges
            )
        )
    }

    override fun updateWorkflow(workflow: Workflow): Flow<ApiResponse> {
        return api.updateWorkflow(
            workflow.id,
            com.makoclaw.core.network.api.UpdateWorkflowRequest(
                name = workflow.name,
                description = workflow.description,
                nodes = workflow.nodes,
                edges = workflow.edges
            )
        )
    }

    override fun deleteWorkflow(id: String): Flow<ApiResponse> {
        return api.deleteWorkflow(id)
    }

    override fun executeWorkflow(id: String): Flow<WorkflowExecutionLog> {
        return api.executeWorkflow(id)
    }

    override fun getExecutionLogs(workflowId: String): Flow<List<WorkflowExecutionLog>> {
        return api.getExecutionLogs(workflowId)
            .map { response -> response.logs }
    }
}
```

### Archivo 2: WorkflowDao.kt
**Ubicación**: `makoclaw-android/feature/feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/database/WorkflowDao.kt`

```kotlin
package com.makoclaw.feature.workflows.data.database

import androidx.room.*
import kotlinx.coroutines.flow.Flow

@Entity(tableName = "workflows")
data class WorkflowEntity(
    @PrimaryKey val id: String,
    val name: String,
    val description: String,
    val nodes: String, // JSON string
    val edges: String, // JSON string
    val status: String,
    val createdAt: Long,
    val updatedAt: Long,
    val lastExecutionAt: Long?
)

@Dao
interface WorkflowDao {
    @Query("SELECT * FROM workflows ORDER BY updatedAt DESC")
    fun getAll(): Flow<List<WorkflowEntity>>

    @Query("SELECT * FROM workflows WHERE id = :id")
    suspend fun getById(id: String): WorkflowEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insert(workflow: WorkflowEntity)

    @Update
    suspend fun update(workflow: WorkflowEntity)

    @Delete
    suspend fun delete(workflow: WorkflowEntity)

    @Query("DELETE FROM workflows WHERE id = :id")
    suspend fun deleteById(id: String)
}

// Extension function to convert Entity to Model
fun WorkflowEntity.toModel(): com.makoclaw.core.model.Workflow {
    return com.makoclaw.core.model.Workflow(
        id = id,
        name = name,
        description = description,
        nodes = kotlinx.serialization.json.Json.decodeFromString(
            kotlinx.serialization.builtins.ListSerializer(com.makoclaw.core.model.WorkflowNode.serializer()),
            nodes
        ),
        edges = kotlinx.serialization.json.Json.decodeFromString(
            kotlinx.serialization.builtins.ListSerializer(com.makoclaw.core.model.WorkflowEdge.serializer()),
            edges
        ),
        status = status,
        createdAt = createdAt,
        updatedAt = updatedAt,
        lastExecutionAt = lastExecutionAt
    )
}
```

### Archivo 3: WorkflowsApi.kt
**Ubicación**: `makoclaw-android/core/core-network/src/main/java/com/makoclaw/core/network/api/WorkflowsApi.kt`

**NOTA**: Crear este archivo solo si NO existe en el directorio.

```kotlin
package com.makoclaw.core.network.api

import com.makoclaw.core.model.Workflow
import com.makoclaw.core.model.WorkflowExecutionLog
import retrofit2.http.*

data class WorkflowsResponse(
    val workflows: List<Workflow> = emptyList()
)

data class WorkflowResponse(
    val workflow: Workflow? = null
)

data class ExecutionLogsResponse(
    val logs: List<WorkflowExecutionLog> = emptyList()
)

data class CreateWorkflowRequest(
    val name: String,
    val description: String,
    val nodes: List<com.makoclaw.core.model.WorkflowNode>,
    val edges: List<com.makoclaw.core.model.WorkflowEdge>
)

data class UpdateWorkflowRequest(
    val name: String,
    val description: String,
    val nodes: List<com.makoclaw.core.model.WorkflowNode>,
    val edges: List<com.makoclaw.core.model.WorkflowEdge>
)

interface WorkflowsApi {
    @GET("workflows")
    suspend fun getWorkflows(): WorkflowsResponse

    @POST("workflows")
    suspend fun createWorkflow(@Body request: CreateWorkflowRequest): WorkflowResponse

    @PUT("workflows/{id}")
    suspend fun updateWorkflow(
        @Path("id") id: String,
        @Body request: UpdateWorkflowRequest
    ): WorkflowResponse

    @DELETE("workflows/{id}")
    suspend fun deleteWorkflow(@Path("id") id: String): ApiResponse

    @POST("workflows/{id}/execute")
    suspend fun executeWorkflow(@Path("id") id: String): WorkflowExecutionLog

    @GET("workflows/{id}/logs")
    suspend fun getExecutionLogs(@Path("id") id: String): ExecutionLogsResponse
}
```

## Instrucciones

1. Crear el directorio: `feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/repository/`
2. Crear `WorkflowsRepository.kt`
3. Crear el directorio: `feature-workflows/src/main/java/com/makoclaw/feature/workflows/data/database/`
4. Crear `WorkflowDao.kt`
5. Verificar si `WorkflowsApi.kt` existe en `core-network/src/main/java/com/makoclaw/core/network/api/`
   - Si NO existe, crearlo
   - Si YA existe, NO crearlo (no modificar)
6. Verificar que no hay errores de compilación

## Resumen

- **Archivos creados**: 3 (o 2 si WorkflowsApi ya existe)
- **Líneas de código**: ~250
- **Complejidad**: Media
- **Tiempo estimado**: ~30-45 minutos

## Finalización

Una vez completado, avísame para lanzar la Micro-Delegación 3 (Canvas Editor).

**¡Éxito, hermano!** 🦈
