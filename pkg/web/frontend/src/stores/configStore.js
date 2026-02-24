import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../services/api'

export const useConfigStore = defineStore('config', () => {
  const configured = ref(false)
  const degradedMode = ref(false)
  const reason = ref('')
  const activeProviders = ref([])
  const loading = ref(false)
  const checkError = ref(null)

  // Check configuration status
  async function checkStatus() {
    loading.value = true
    checkError.value = null
    try {
      const response = await api.get('/config/status')
      configured.value = response.data.configured
      degradedMode.value = response.data.degradedMode
      reason.value = response.data.reason || ''
      activeProviders.value = response.data.activeProviders || []
      return response.data
    } catch (error) {
      console.error('Failed to check config status:', error)
      checkError.value = error.message
      // Assume degraded mode if we can't check
      degradedMode.value = true
      configured.value = false
      throw error
    } finally {
      loading.value = false
    }
  }

  // Update provider configuration
  async function updateProvider(providerData) {
    try {
      const response = await api.post('/config/provider', providerData)
      // Refresh status after update
      await checkStatus()
      return response.data
    } catch (error) {
      console.error('Failed to update provider:', error)
      throw error
    }
  }

  // Validate provider credentials
  async function validateProvider(providerData) {
    try {
      const response = await api.post('/config/validate', providerData)
      return response.data
    } catch (error) {
      console.error('Failed to validate provider:', error)
      throw error
    }
  }

  // Check if the system needs configuration
  function needsConfiguration() {
    return degradedMode.value || !configured.value
  }

  return {
    configured,
    degradedMode,
    reason,
    activeProviders,
    loading,
    checkError,
    checkStatus,
    updateProvider,
    validateProvider,
    needsConfiguration
  }
})
