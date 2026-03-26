package com.makoclaw.core.network.interceptor

import com.makoclaw.core.common.Constants
import com.makoclaw.core.security.TokenManager
import okhttp3.Interceptor
import okhttp3.Response
import javax.inject.Inject

class AuthInterceptor @Inject constructor(
    private val tokenManager: TokenManager
) : Interceptor {

    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        val token = tokenManager.getToken()

        val authenticatedRequest = if (token != null) {
            request.newBuilder()
                .header(Constants.AUTH_HEADER, "${Constants.BEARER_PREFIX}$token")
                .build()
        } else {
            request
        }

        val response = chain.proceed(authenticatedRequest)

        if (response.code == 401) {
            tokenManager.clearToken()
        }

        return response
    }
}
