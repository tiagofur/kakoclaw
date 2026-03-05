import { useAuthStore } from '../stores/authStore'
import api from './api'

export default {
    login: async (username: string, password: string) => {
        const response = await api.post('/auth/login', { username, password })
        const token = response.data.token
        const expiresIn = response.data.expires_in ? Math.floor(response.data.expires_in / 60) : 1440

        // Fetch user profile with the new token
        const meResponse = await api.get('/auth/me', {
            headers: { Authorization: `Bearer ${token}` }
        })

        const authStore = useAuthStore()
        await authStore.setCredentials(
            meResponse.data, // { username, role }
            token,
            expiresIn
        )
        return response.data
    },

    logout: async () => {
        const authStore = useAuthStore()
        await authStore.logout()
    },

    changePassword: async (currentPassword: string, newPassword: string) => {
        const response = await api.post('/auth/change-password', {
            current_password: currentPassword,
            new_password: newPassword
        })
        return response.data
    },

    getCurrentUser: async () => {
        const response = await api.get('/auth/me')
        return response.data
    }
}
