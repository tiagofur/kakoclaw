<template>
  <div class="flex h-screen bg-kakoclaw-bg text-kakoclaw-text overflow-hidden">
    <!-- Sidebar -->
    <Sidebar />

    <!-- Toast Notifications -->
    <ToastContainer />

    <!-- Main Content Area -->
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden relative">
      
      <!-- Mobile Header -->
      <header class="md:hidden h-14 bg-kakoclaw-surface border-b border-kakoclaw-border flex items-center justify-between px-4 flex-shrink-0 z-20">
         <div class="font-bold text-lg text-kakoclaw-accent">KakoClaw</div>
         <button @click="uiStore.toggleSidebar()" class="p-2 text-kakoclaw-text hover:bg-kakoclaw-border rounded">
           <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
             <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16m-7 6h7" />
           </svg>
         </button>
      </header>

      <!-- Page Content -->
      <main class="flex-1 overflow-auto relative scroll-smooth p-4 md:p-6">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onBeforeUnmount } from 'vue'
import { useUIStore } from '../../stores/uiStore'
import { useChatStore } from '../../stores/chatStore'
import { getChatWebSocket } from '../../services/websocketService'
import Sidebar from './Sidebar.vue'
import ToastContainer from './ToastContainer.vue'

const uiStore = useUIStore()
const chatStore = useChatStore()
const chatWs = getChatWebSocket()

// Flag set by ChatView when it's mounted and managing WS messages directly.
// When true, this background listener skips queueing (ChatView handles messages itself).
let chatViewActive = false

const setChatViewActive = (active) => { chatViewActive = active }

// Expose globally so ChatView can signal its mount/unmount status
window.__kakoclaw_setChatViewActive = setChatViewActive

// Background handler: runs when ChatView is NOT mounted.
// Captures WS messages into the chatStore so they aren't lost.
const backgroundMessageHandler = (message) => {
  if (chatViewActive) return // ChatView is handling messages itself

  if (message.type === 'stream_start') {
    chatStore.setIsWorking(true)
    chatStore.enqueuePendingMessage(message)
  } else if (message.type === 'stream') {
    chatStore.enqueuePendingMessage(message)
  } else if (message.type === 'stream_end') {
    chatStore.enqueuePendingMessage(message)
  } else if (message.type === 'message') {
    chatStore.enqueuePendingMessage(message)
  } else if (message.type === 'tool_call') {
    chatStore.enqueuePendingMessage(message)
  } else if (message.type === 'ready') {
    chatStore.setIsWorking(false)
    chatStore.setGlobalLoading(false)
    chatStore.enqueuePendingMessage(message)
  }
}

onMounted(() => {
  chatWs.on('message', backgroundMessageHandler)
})

onBeforeUnmount(() => {
  chatWs.off('message', backgroundMessageHandler)
})
</script>

<style scoped>
/* Custom Scrollbar */
main::-webkit-scrollbar {
  width: 8px;
}
main::-webkit-scrollbar-track {
  background: transparent;
}
main::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5); /* gray-400 */
  border-radius: 4px;
}
main::-webkit-scrollbar-thumb:hover {
  background-color: rgba(156, 163, 175, 0.8);
}
</style>
