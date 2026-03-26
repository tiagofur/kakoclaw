package com.makoclaw.core.network.api

import com.makoclaw.core.model.ConfigStatus
import com.makoclaw.core.model.ModelsResponse
import kotlinx.serialization.Serializable
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST

@Serializable
data class HealthResponse(val status: String = "ok")

@Serializable
data class ProviderCatalog(val providers: List<ProviderCatalogEntry> = emptyList())

@Serializable
data class ProviderCatalogEntry(
    val name: String,
    val displayName: String = "",
    val requiresApiKey: Boolean = true
)

@Serializable
data class ValidateProviderRequest(
    val provider: String,
    val api_key: String
)

@Serializable
data class UpdateProviderRequest(
    val provider: String,
    val api_key: String,
    val config: Map<String, String> = emptyMap()
)

interface ConfigApi {

    @GET("health")
    suspend fun healthCheck(): HealthResponse

    @GET("config/status")
    suspend fun getConfigStatus(): ConfigStatus

    @GET("models")
    suspend fun getModels(): ModelsResponse

    @GET("providers/catalog")
    suspend fun getProvidersCatalog(): ProviderCatalog

    @POST("config/validate")
    suspend fun validateProvider(@Body request: ValidateProviderRequest)

    @POST("config/provider")
    suspend fun updateProvider(@Body request: UpdateProviderRequest)

    @GET("me/config")
    suspend fun getUserConfig(): Map<String, String>

    @POST("me/onboarding/complete")
    suspend fun completeOnboarding()

    @GET("tools")
    suspend fun getTools(): Map<String, List<String>>
}
