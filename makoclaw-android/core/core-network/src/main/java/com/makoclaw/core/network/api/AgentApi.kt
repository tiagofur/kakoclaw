package com.makoclaw.core.network.api

import com.makoclaw.core.model.AgentMetrics
import com.makoclaw.core.model.CreateSpecialistRequest
import com.makoclaw.core.model.OrchestratorConfig
import com.makoclaw.core.model.Specialist
import com.makoclaw.core.model.Swarm
import kotlinx.serialization.Serializable
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.PUT
import retrofit2.http.Path

@Serializable
data class AgentsResponse(
    val specialists: List<Specialist> = emptyList(),
    val orchestrator: OrchestratorConfig = OrchestratorConfig()
)

@Serializable
data class SwarmsResponse(val swarms: List<Swarm> = emptyList())

@Serializable
data class SwarmTemplatesResponse(val templates: List<Swarm> = emptyList())

@Serializable
data class GenerateRequest(val description: String)

@Serializable
data class RunSwarmRequest(
    val name: String,
    val input: String,
    val session_id: String? = null
)

interface AgentApi {

    @GET("agents")
    suspend fun getAgents(): AgentsResponse

    @GET("agents/{name}")
    suspend fun getAgentDetails(@Path("name") name: String): Specialist

    @POST("agents/specialist")
    suspend fun createSpecialist(@Body request: CreateSpecialistRequest): Specialist

    @PUT("agents/specialist/{name}")
    suspend fun updateSpecialist(
        @Path("name") name: String,
        @Body request: CreateSpecialistRequest
    ): Specialist

    @DELETE("agents/specialist/{name}")
    suspend fun deleteSpecialist(@Path("name") name: String)

    @POST("agents/generate")
    suspend fun generateSpecialist(@Body request: GenerateRequest): Specialist

    @POST("agents/orchestrator")
    suspend fun updateOrchestrator(@Body config: OrchestratorConfig)

    @GET("agents/metrics")
    suspend fun getMetrics(): AgentMetrics

    @POST("agents/metrics/reset")
    suspend fun resetMetrics()

    @GET("swarms")
    suspend fun getSwarms(): SwarmsResponse

    @POST("swarms")
    suspend fun createSwarm(@Body swarm: Swarm): Swarm

    @DELETE("swarms/{name}")
    suspend fun deleteSwarm(@Path("name") name: String)

    @GET("swarms/templates")
    suspend fun getSwarmTemplates(): SwarmTemplatesResponse

    @POST("swarms/run")
    suspend fun runSwarm(@Body request: RunSwarmRequest)
}
