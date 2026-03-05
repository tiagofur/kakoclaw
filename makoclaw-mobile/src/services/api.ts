import axios from 'axios'
import { useAuthStore } from '../stores/authStore'
import { useConfigStore } from '../stores/configStore'

function createApiClient() {
    const client = axios.create({
        timeout: 30000,
        headers: {
            'Content-Type': 'application/json'
        }
    })

    // Request interceptor — inject base URL + auth token
    client.interceptors.request.use(
        config => {
            const configStore = useConfigStore()
            const authStore = useAuthStore()

            // Dynamic base URL from config store (user-configurable server address)
            config.baseURL = `${configStore.serverUrl}/api/v1`

            if (authStore.token) {
                config.headers.Authorization = `Bearer ${authStore.token}`
            }
            return config
        },
        error => Promise.reject(error)
    )

    // Response interceptor — handle 401 globally
    client.interceptors.response.use(
        response => response,
        error => {
            const isAuthEndpoint = error.config?.url?.includes('/auth/login') || error.config?.url?.includes('/auth/register')
            if (error.response?.status === 401 && !isAuthEndpoint) {
                const authStore = useAuthStore()
                authStore.logout()
            }
            return Promise.reject(error)
        }
    )

    return client
}

// Lazy singleton — created on first use (after Pinia is ready)
let _client: ReturnType<typeof createApiClient> | null = null

function getClient() {
    if (!_client) {
        _client = createApiClient()
    }
    return _client
}

export default {
    get: (...args: Parameters<ReturnType<typeof createApiClient>['get']>) => getClient().get(...args),
    post: (...args: Parameters<ReturnType<typeof createApiClient>['post']>) => getClient().post(...args),
    put: (...args: Parameters<ReturnType<typeof createApiClient>['put']>) => getClient().put(...args),
    patch: (...args: Parameters<ReturnType<typeof createApiClient>['patch']>) => getClient().patch(...args),
    delete: (...args: Parameters<ReturnType<typeof createApiClient>['delete']>) => getClient().delete(...args),
}
