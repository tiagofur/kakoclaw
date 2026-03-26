package com.makoclaw.core.network.api

import com.makoclaw.core.model.ChangePasswordRequest
import com.makoclaw.core.model.LoginRequest
import com.makoclaw.core.model.LoginResponse
import com.makoclaw.core.model.RegisterRequest
import com.makoclaw.core.model.User
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST

interface AuthApi {

    @POST("auth/login")
    suspend fun login(@Body request: LoginRequest): LoginResponse

    @POST("auth/register")
    suspend fun register(@Body request: RegisterRequest)

    @GET("auth/me")
    suspend fun getCurrentUser(): User

    @GET("auth/profile")
    suspend fun getProfile(): User

    @POST("auth/change-password")
    suspend fun changePassword(@Body request: ChangePasswordRequest)
}
