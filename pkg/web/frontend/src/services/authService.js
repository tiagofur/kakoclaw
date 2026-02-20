import { useAuthStore } from '../stores/authStore'
import { useChatStore } from '../stores/chatStore'
import client from './api'

export default {
  login: async (username, password) => {
    const response = await client.post('/auth/login', { username, password })
    const token = response.data.token
    const expiresIn = response.data.expires_in ? Math.floor(response.data.expires_in / 60) : 1440

    // Temporarily set token in client to fetch user details
    client.defaults.headers.common['Authorization'] = `Bearer ${token}`
    const meResponse = await client.get('/auth/me')
    
    const authStore = useAuthStore()
    authStore.setCredentials(
      meResponse.data, // now includes { username, role }
      token,
      expiresIn
    )
    return response.data
  },

  logout: () => {
    const authStore = useAuthStore()
    authStore.logout()
  },

  changePassword: async (currentPassword, newPassword) => {
    const response = await client.post('/auth/change-password', {
      current_password: currentPassword,
      new_password: newPassword
    })
    return response.data
  },

  getCurrentUser: async () => {
    const response = await client.get('/auth/me')
    return response.data
  }
}
