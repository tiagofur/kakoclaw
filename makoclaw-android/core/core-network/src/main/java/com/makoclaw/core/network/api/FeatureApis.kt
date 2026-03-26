package com.makoclaw.core.network.api

import com.makoclaw.core.model.CronJob
import com.makoclaw.core.model.DailyNote
import com.makoclaw.core.model.FileEntry
import com.makoclaw.core.model.KnowledgeChunk
import com.makoclaw.core.model.KnowledgeDocument
import com.makoclaw.core.model.McpServer
import com.makoclaw.core.model.MemoryContent
import com.makoclaw.core.model.MetricsData
import com.makoclaw.core.model.Skill
import com.makoclaw.core.model.Workflow
import com.makoclaw.core.model.WorkflowEdge
import com.makoclaw.core.model.WorkflowExecutionLog
import com.makoclaw.core.model.WorkflowNode
import com.makoclaw.core.model.TimeRange
import kotlinx.serialization.Serializable
import okhttp3.MultipartBody
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part
import retrofit2.http.Path
import retrofit2.http.PUT
import retrofit2.http.Query

@Serializable
data class ApiResponse(
    val success: Boolean = true,
    val message: String = ""
)

// ─── Knowledge ───

@Serializable
data class KnowledgeListResponse(val documents: List<KnowledgeDocument> = emptyList())

@Serializable
data class KnowledgeSearchResponse(val results: List<KnowledgeChunk> = emptyList())

@Serializable
data class KnowledgeSearchRequest(val query: String)

@Serializable
data class KnowledgeDetailResponse(
    val document: KnowledgeDocument = KnowledgeDocument(),
    val chunks: List<KnowledgeChunk> = emptyList()
)

@Serializable
data class EditKnowledgeChunkRequest(
    val chunkId: String,
    val content: String
)

interface KnowledgeApi {
    @GET("knowledge/documents")
    suspend fun getDocuments(): KnowledgeListResponse

    @Multipart
    @POST("knowledge/upload")
    suspend fun uploadDocument(@Part file: MultipartBody.Part): KnowledgeDocument

    @POST("knowledge/search")
    suspend fun searchDocuments(@Body request: KnowledgeSearchRequest): KnowledgeSearchResponse

    @GET("knowledge/document/{id}")
    suspend fun getDocument(@Path("id") id: String): KnowledgeDetailResponse

    @DELETE("knowledge/document/{id}")
    suspend fun deleteDocument(@Path("id") id: String)

    @retrofit2.http.PATCH("knowledge/document/{id}/chunk")
    suspend fun editChunk(
        @Path("id") id: String,
        @Body request: EditKnowledgeChunkRequest
    ): KnowledgeDetailResponse
}

// ─── Skills ───

@Serializable
data class SkillsListResponse(val skills: List<Skill> = emptyList())

@Serializable
data class InstallSkillRequest(val repository: String)

@Serializable
data class GenerateSkillRequest(
    val name: String,
    val goal: String,
    val capabilities: String = "",
    val constraints: String = "",
    val tools: String = "",
    val examples: String = ""
)

interface SkillsApi {
    @GET("skills")
    suspend fun listSkills(): SkillsListResponse

    @GET("skills?type=available")
    suspend fun listAvailableSkills(): SkillsListResponse

    @GET("skills/{name}")
    suspend fun getSkill(@Path("name") name: String): Skill

    @POST("skills/install")
    suspend fun installSkill(@Body request: InstallSkillRequest)

    @DELETE("skills/{name}")
    suspend fun uninstallSkill(@Path("name") name: String)

    @POST("skills/generate")
    suspend fun generateSkill(@Body request: GenerateSkillRequest): Skill

    @GET("skills/analytics")
    suspend fun getAnalytics(): Map<String, Any>

    @GET("marketplace/skills")
    suspend fun getMarketplaceSkills(): SkillsListResponse

    @GET("marketplace/categories")
    suspend fun getCategories(): Map<String, List<String>>
}

// ─── Cron ───

@Serializable
data class CronListResponse(val jobs: List<CronJob> = emptyList())

@Serializable
data class CronJobsResponse(val jobs: List<CronJob> = emptyList())

@Serializable
data class CreateCronJobRequest(
    val name: String,
    val description: String,
    val cronExpression: String,
    val workflowId: String
)

@Serializable
data class UpdateCronJobRequest(
    val name: String,
    val description: String,
    val cronExpression: String,
    val workflowId: String
)

@Serializable
data class GenerateCronRequest(
    val schedule: String,
    val workflowId: String
)

@Serializable
data class GenerateCronResponse(val expression: String = "")

@Serializable
data class TestCronResponse(
    val valid: Boolean = true,
    val nextRuns: List<String> = emptyList(),
    val error: String? = null
)

interface CronApi {
    @GET("cron/jobs")
    suspend fun getJobs(): CronJobsResponse

    @POST("cron/jobs")
    suspend fun createJob(@Body request: CreateCronJobRequest): ApiResponse

    @PUT("cron/jobs/{id}")
    suspend fun updateJob(
        @Path("id") id: String,
        @Body request: UpdateCronJobRequest
    ): ApiResponse

    @DELETE("cron/jobs/{id}")
    suspend fun deleteJob(@Path("id") id: String): ApiResponse

    @POST("cron/jobs/{id}/toggle")
    suspend fun toggleJob(@Path("id") id: String, @Body enabled: Boolean): ApiResponse

    @POST("cron/generate")
    suspend fun generateCron(@Body request: GenerateCronRequest): GenerateCronResponse

    @POST("cron/test")
    suspend fun testRun(@Body expression: String): TestCronResponse
}

// ─── Files ───

@Serializable
data class FilesListResponse(
    val path: String = "",
    val entries: List<FileEntry> = emptyList()
)

interface FilesApi {
    @GET("files")
    suspend fun listRoot(): FilesListResponse

    @GET("files/{path}")
    suspend fun browse(@Path("path", encoded = true) path: String): FilesListResponse

    @Multipart
    @POST("files/upload")
    suspend fun uploadFile(@Part file: MultipartBody.Part): Map<String, String>

    @DELETE("files/{path}")
    suspend fun deleteFile(@Path("path", encoded = true) path: String)
}

// ─── Memory ───

interface MemoryApi {
    @GET("memory/longterm")
    suspend fun getLongTermMemory(): MemoryContent

    @POST("memory/longterm")
    suspend fun updateLongTermMemory(@Body content: MemoryContent)

    @GET("memory/daily")
    suspend fun getDailyNotes(@Query("days") days: Int = 7): List<DailyNote>
}

// ─── MCP ───

@Serializable
data class McpListResponse(val servers: List<McpServer> = emptyList())

interface McpApi {
    @GET("mcp")
    suspend fun listServers(): McpListResponse

    @GET("mcp/{name}")
    suspend fun getServer(@Path("name") name: String): McpServer

    @POST("mcp/{name}/reconnect")
    suspend fun reconnectServer(@Path("name") name: String)

    @POST("mcp/reconnect-all")
    suspend fun reconnectAll()
}

// ─── Workflows ───

@Serializable
data class WorkflowsResponse(val workflows: List<Workflow> = emptyList())

@Serializable
data class WorkflowResponse(val workflow: Workflow? = null)

@Serializable
data class ExecutionLogsResponse(val logs: List<WorkflowExecutionLog> = emptyList())

@Serializable
data class CreateWorkflowRequest(
    val name: String,
    val description: String,
    val nodes: List<WorkflowNode>,
    val edges: List<WorkflowEdge>
)

@Serializable
data class UpdateWorkflowRequest(
    val name: String,
    val description: String,
    val nodes: List<WorkflowNode>,
    val edges: List<WorkflowEdge>
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

// ─── Metrics ───

@Serializable
data class MetricsResponse(val metrics: MetricsData? = null)

@Serializable
data class ExportRequest(
    val timeRange: String,
    val agentId: String?,
    val format: String = "csv"
)

@Serializable
data class ExportResult(
    val filePath: String = "",
    val fileSize: Long = 0L,
    val format: String = "csv"
)

interface MetricsApi {
    @GET("metrics")
    suspend fun getMetrics(
        @Query("timeRange") timeRange: TimeRange,
        @Query("agentId") agentId: String?
    ): MetricsResponse

    @POST("metrics/export")
    suspend fun exportMetrics(@Body request: ExportRequest): ExportResult
}

// ─── Voice ───

@Serializable
data class TranscriptionResponse(
    val text: String = "",
    val language: String = "",
    val duration: Double = 0.0
)

interface VoiceApi {
    @Multipart
    @POST("voice/transcribe")
    suspend fun transcribe(@Part audio: MultipartBody.Part): TranscriptionResponse

    @POST("voice/synthesize")
    suspend fun synthesize(@Body request: Map<String, String>): okhttp3.ResponseBody
}

// ─── Reports ───

@Serializable
data class EmailReportRequest(
    val to: String,
    val subject: String = "",
    val content: String = ""
)

interface ReportsApi {
    @POST("reports/email")
    suspend fun emailReport(@Body request: EmailReportRequest)
}
