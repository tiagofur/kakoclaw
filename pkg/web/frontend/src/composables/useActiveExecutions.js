import { ref, onMounted, onBeforeUnmount } from 'vue'

/**
 * Composable to monitor active agent executions
 * Polls the /api/v1/chat/active endpoint and provides reactive state
 */
export function useActiveExecutions(options = {}) {
  const { pollInterval = 3000 } = options
  
  const activeExecutions = ref([])
  const isPolling = ref(false)
  let pollTimer = null

  const fetchActiveExecutions = async () => {
    try {
      const token = localStorage.getItem('auth.token')
      const response = await fetch('/api/v1/chat/active', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      
      if (response.ok) {
        const data = await response.json()
        activeExecutions.value = data || []
      }
    } catch (error) {
      console.error('Failed to fetch active executions:', error)
    }
  }

  const startPolling = () => {
    if (isPolling.value) return
    isPolling.value = true
    
    // Initial fetch
    fetchActiveExecutions()
    
    // Poll at intervals
    pollTimer = setInterval(fetchActiveExecutions, pollInterval)
  }

  const stopPolling = () => {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
    isPolling.value = false
  }

  // Auto-start polling when composable is used
  onMounted(() => {
    startPolling()
  })

  // Cleanup on unmount
  onBeforeUnmount(() => {
    stopPolling()
  })

  return {
    activeExecutions,
    isPolling,
    startPolling,
    stopPolling,
    refresh: fetchActiveExecutions
  }
}
