package com.makoclaw.core.network.api

import com.makoclaw.core.model.CronJob
import com.makoclaw.core.model.DailyNote
import com.makoclaw.core.model.FileEntry
import com.makoclaw.core.model.KnowledgeChunk
import com.makoclaw.core.model.KnowledgeDocument
import com.makoclaw.core.model.McpServer
import com.makoclaw.core.model.MemoryContent
import com.makoclaw.core.model.Skill
import com.makoclaw.core.model.Workflow
import kotlinx.serialization.Serializable
import okhttp3.MultipartBody
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part
import retrofit2.http.Path
import retrofit2.http.Query

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
data class CreateCronRequest(
    val name: String,
    val schedule: String,
    val message: String,
    val enabled: Boolean = true
)

@Serializable
data class AiCronRequest(val prompt: String)

@Serializable
data class AiCronResponse(val schedule: String = "", val description: String = "")

interface CronApi {
    @GET("cron")
    suspend fun listJobs(@Query("include_disabled") includeDisabled: Boolean = true): CronListResponse

    @POST("cron")
    suspend fun createJob(@Body request: CreateCronRequest): CronJob

    @DELETE("cron/{id}")
    suspend fun deleteJob(@Path("id") id: String)

    @POST("cron/{id}")
    suspend fun toggleJob(@Path("id") id: String, @Body enabled: Map<String, Boolean>)

    @POST("ai/create-cron")
    suspend fun generateCron(@Body request: AiCronRequest): AiCronResponse
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
data class WorkflowsListResponse(val workflows: List<Workflow> = emptyList())

@Serializable
data class RunWorkflowRequest(val parameters: Map<String, String> = emptyMap())

interface WorkflowsApi {
    @GET("workflows")
    suspend fun listWorkflows(): WorkflowsListResponse

    @POST("workflows")
    suspend fun createWorkflow(@Body workflow: Workflow): Workflow

    @GET("workflows/{name}")
    suspend fun getWorkflow(@Path("name") name: String): Workflow

    @DELETE("workflows/{name}")
    suspend fun deleteWorkflow(@Path("name") name: String)

    @POST("workflows/{name}/run")
    suspend fun runWorkflow(
        @Path("name") name: String,
        @Body request: RunWorkflowRequest = RunWorkflowRequest()
    )
}

// ─── Metrics ───

@Serializable
data class MetricsResponse(
    val llmCalls: Int = 0,
    val totalTokens: Long = 0,
    val totalCost: Double = 0.0,
    val toolInvocations: Int = 0,
    val agentBreakdown: Map<String, Int> = emptyMap(),
    val toolBreakdown: Map<String, Int> = emptyMap()
)

interface MetricsApi {
    @GET("metrics")
    suspend fun getMetrics(): MetricsResponse
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
