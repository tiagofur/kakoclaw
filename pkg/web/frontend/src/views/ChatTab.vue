<template>
  <div class="flex flex-col h-full bg-picoclaw-bg">
    <!-- Messages Area -->
    <div 
      ref="messagesContainer"
      class="flex-1 overflow-y-auto p-4 space-y-4"
    >
      <div v-if="messages.length === 0" class="flex items-center justify-center h-full text-picoclaw-text-secondary">
        <div class="text-center">
          <svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
          <p>Start a conversation</p>
        </div>
      </div>

      <!-- Messages -->
      <div v-for="msg in messages" :key="msg.id" class="animate-fadeIn">
        <div
          :class="[
            'flex',
            msg.role === 'user' ? 'justify-end' : 'justify-start'
          ]"
        >
          <div
            :class="[
              'max-w-xs lg:max-w-md px-4 py-2 rounded-lg',
              msg.role === 'user'
                ? 'bg-picoclaw-accent text-white rounded-br-none'
                : 'bg-picoclaw-surface border border-picoclaw-border rounded-bl-none'
            ]"
          >
            <p class="text-sm whitespace-pre-wrap break-words">{{ msg.content }}</p>
            <p class="text-xs mt-1" :class="msg.role === 'user' ? 'opacity-70' : 'text-picoclaw-text-secondary'">
              {{ formatTime(msg.timestamp) }}
            </p>
          </div>
        </div>
      </div>

      <!-- Loading indicator -->
      <div v-if="isLoading" class="flex justify-start">
        <div class="bg-picoclaw-surface border border-picoclaw-border rounded-lg rounded-bl-none px-4 py-2">
          <div class="flex gap-2">
            <div class="w-2 h-2 bg-picoclaw-accent rounded-full animate-bounce" style="animation-delay: 0s"></div>
            <div class="w-2 h-2 bg-picoclaw-accent rounded-full animate-bounce" style="animation-delay: 0.2s"></div>
            <div class="w-2 h-2 bg-picoclaw-accent rounded-full animate-bounce" style="animation-delay: 0.4s"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Input Area -->
    <div class="border-t border-picoclaw-border bg-picoclaw-surface p-4">
      <form @submit.prevent="sendMessage" class="flex gap-2">
        <input
          v-model="messageInput"
          type="text"
          placeholder="Type a message or /task list, /task run..."
          class="flex-1 px-4 py-2 bg-picoclaw-bg border border-picoclaw-border rounded focus-ring text-sm"
          :disabled="!isConnected || isLoading"
        />
        <button
          type="submit"
          :disabled="!isConnected || isLoading || !messageInput.trim()"
          class="px-4 py-2 bg-picoclaw-accent hover:bg-picoclaw-accent-hover text-white rounded transition-smooth disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
          </svg>
        </button>
      </form>

      <!-- Connection Status -->
      <div class="text-xs text-picoclaw-text-secondary mt-2">
        <span v-if="isConnected" class="text-picoclaw-success">● Connected</span>
        <span v-else class="text-picoclaw-error">● Disconnected</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import { useChatStore } from '../stores/chatStore'
import { ChatWebSocket } from '../services/websocketService'

const chatStore = useChatStore()
const messagesContainer = ref(null)
const messageInput = ref('')
const isConnected = ref(false)
const isLoading = ref(false)

const messages = chatStore.messages
const chatWs = new ChatWebSocket()

onMounted(async () => {
  try {
    await chatWs.connect()
    isConnected.value = true
    chatStore.setConnected(true)

    // Listen for messages
    chatWs.on('message', (message) => {
      if (message.type === 'message') {
        chatStore.addMessage({
          role: message.role || 'assistant',
          content: message.content
        })
      }
      if (message.type === 'ready') {
        isLoading.value = false
      }
    })

    // Listen for connection events
    chatWs.on('disconnected', () => {
      isConnected.value = false
      chatStore.setConnected(false)
    })

    chatWs.on('connected', () => {
      isConnected.value = true
      chatStore.setConnected(true)
    })

    chatStore.setWebSocket(chatWs)
  } catch (error) {
    console.error('Failed to connect to chat:', error)
  }
})

// Auto-scroll to bottom
watch(messages, async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
})

const sendMessage = async () => {
  const content = messageInput.value.trim()
  if (!content) return

  // Add user message
  chatStore.addMessage({
    role: 'user',
    content
  })

  messageInput.value = ''
  isLoading.value = true

  // Send via WebSocket
  if (chatWs.isConnected()) {
    chatWs.send({
      type: 'message',
      content
    })
  } else {
    isLoading.value = false
  }
}

const formatTime = (timestamp) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fadeIn {
  animation: fadeIn 0.3s ease-in;
}
</style>
