package com.makoclaw.core.security.storage

import android.content.Context
import androidx.test.core.app.ApplicationProvider
import com.makoclaw.core.security.interceptor.JwtInterceptor
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import okhttp3.Interceptor
import okhttp3.Protocol
import okhttp3.Request
import okhttp3.Response
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertNull
import org.junit.jupiter.api.Test

class JwtStorageTest {
    private fun createStorage(name: String = "jwt-storage-test"): JwtStorage {
        val context = ApplicationProvider.getApplicationContext<Context>()
        val prefs = context.getSharedPreferences(name, Context.MODE_PRIVATE)
        prefs.edit().clear().commit()
        return JwtStorage(prefs)
    }

    @Test
    fun `save and get jwt works`() {
        val storage = createStorage()

        storage.saveJwt("jwt-token")

        assertEquals("jwt-token", storage.getJwt())
    }

    @Test
    fun `save and get refresh token works`() {
        val storage = createStorage()

        storage.saveRefreshToken("refresh-token")

        assertEquals("refresh-token", storage.getRefreshToken())
    }

    @Test
    fun `clearTokens removes jwt and refresh token`() {
        val storage = createStorage()
        storage.saveJwt("jwt-token")
        storage.saveRefreshToken("refresh-token")

        storage.clearTokens()

        assertNull(storage.getJwt())
        assertNull(storage.getRefreshToken())
    }

    @Test
    fun `save and get api key works`() {
        val storage = createStorage()

        storage.saveApiKey("api-key")

        assertEquals("api-key", storage.getApiKey())
    }

    @Test
    fun `clearAll removes all secure values`() {
        val storage = createStorage()
        storage.saveJwt("jwt-token")
        storage.saveRefreshToken("refresh-token")
        storage.saveApiKey("api-key")

        storage.clearAll()

        assertNull(storage.getJwt())
        assertNull(storage.getRefreshToken())
        assertNull(storage.getApiKey())
    }

    @Test
    fun `interceptor refreshes jwt and retries once`() {
        val storage = createStorage("jwt-storage-interceptor")
        storage.saveJwt("expired-token")
        storage.saveRefreshToken("refresh-token")

        val requests = mutableListOf<Request>()
        val chain = mockk<Interceptor.Chain>()
        val request = Request.Builder().url("https://example.com/api").build()
        every { chain.request() } returns request
        every { chain.proceed(capture(requests)) } returnsMany listOf(
            responseFor(request, 401),
            responseFor(request, 200)
        )

        val interceptor = object : JwtInterceptor(storage) {
            override fun refreshJwt(refreshToken: String): String? =
                if (refreshToken == "refresh-token") "new-token" else null
        }

        val response = interceptor.intercept(chain)

        assertEquals(200, response.code)
        assertEquals("Bearer expired-token", requests.first().header("Authorization"))
        assertEquals("Bearer new-token", requests.last().header("Authorization"))
        verify(exactly = 2) { chain.proceed(any()) }
    }

    private fun responseFor(request: Request, code: Int): Response = Response.Builder()
        .request(request)
        .protocol(Protocol.HTTP_1_1)
        .code(code)
        .message(if (code == 200) "OK" else "Unauthorized")
        .build()
}
