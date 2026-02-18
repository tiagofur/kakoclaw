import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const token = ref(null)
  const sessionExpiry = ref(null)
  
  const isAuthenticated = computed(() => !!token.value)
  
  const isSessionExpired = computed(() => {
    if (!sessionExpiry.value) return false
    return new Date() > new Date(sessionExpiry.value)
  })

  function setCredentials(newUser, newToken, expiryMinutes = 1440) {
    user.value = newUser
    token.value = newToken
    const expiry = new Date()
    expiry.setMinutes(expiry.getMinutes() + expiryMinutes)
    sessionExpiry.value = expiry.toISOString()
    
    // Persist to localStorage
    localStorage.setItem('auth.user', JSON.stringify(newUser))
    localStorage.setItem('auth.token', newToken)
    localStorage.setItem('auth.sessionExpiry', sessionExpiry.value)
  }

  function restoreSession() {
    const savedUser = localStorage.getItem('auth.user')
    const savedToken = localStorage.getItem('auth.token')
    const savedExpiry = localStorage.getItem('auth.sessionExpiry')
    
    if (savedUser && savedToken && savedExpiry) {
      user.value = JSON.parse(savedUser)
      token.value = savedToken
      sessionExpiry.value = savedExpiry
    }
  }

  function logout() {
    user.value = null
    token.value = null
    sessionExpiry.value = null
    localStorage.removeItem('auth.user')
    localStorage.removeItem('auth.token')
    localStorage.removeItem('auth.sessionExpiry')
  }

  return {
    user,
    token,
    sessionExpiry,
    isAuthenticated,
    isSessionExpired,
    setCredentials,
    restoreSession,
    logout
  }
})
