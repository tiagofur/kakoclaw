import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useChatStore = defineStore('chat', () => {
  const messages = ref([])
  const isConnected = ref(false)
  const isLoading = ref(false)
  const ws = ref(null)

  function addMessage(message) {
    messages.value.push({
      id: Date.now(),
      ...message,
      timestamp: new Date().toISOString()
    })
  }

  function clearMessages() {
    messages.value = []
  }

  function setConnected(connected) {
    isConnected.value = connected
  }

  function setLoading(loading) {
    isLoading.value = loading
  }

  function setWebSocket(websocket) {
    ws.value = websocket
  }

  return {
    messages,
    isConnected,
    isLoading,
    ws,
    addMessage,
    clearMessages,
    setConnected,
    setLoading,
    setWebSocket
  }
})
