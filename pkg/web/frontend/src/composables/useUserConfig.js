import { ref, computed } from 'vue'
import { useAuthStore } from '../stores/authStore'

export function useUserConfig() {
  const authStore = useAuthStore()
  const isLoading = ref(false)
  const error = ref('')
  const config = ref(null)

  const fetchConfig = async () => {
    try {
      isLoading.value = true
      error.value = ''

      const response = await fetch('/api/v1/me/config', {
        headers: {
          'Authorization': `Bearer ${authStore.token}`
        }
      })

      if (!response.ok) {
        throw new Error('Failed to fetch config')
      }

      config.value = await response.json()
      return config.value
    } catch (err) {
      error.value = err.message
      console.error('Error fetching config:', err)
    } finally {
      isLoading.value = false
    }
  }

  const hasProvider = computed(() => {
    return config.value?.config?.providers && Object.keys(config.value.config.providers).length > 0
  })

  const hasChannel = computed(() => {
    return config.value?.config?.channels && Object.keys(config.value.config.channels).length > 0
  })

  const isConfigured = computed(() => {
    return hasProvider.value && hasChannel.value
  })

  return {
    config,
    isLoading,
    error,
    fetchConfig,
    hasProvider,
    hasChannel,
    isConfigured
  }
}
