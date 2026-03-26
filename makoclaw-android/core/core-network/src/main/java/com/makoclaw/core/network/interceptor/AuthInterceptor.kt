package com.makoclaw.core.network.interceptor

import com.makoclaw.core.security.interceptor.JwtInterceptor
import okhttp3.Interceptor
import okhttp3.Response
import javax.inject.Inject

class AuthInterceptor @Inject constructor(
    private val jwtInterceptor: JwtInterceptor
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response = jwtInterceptor.intercept(chain)
}
