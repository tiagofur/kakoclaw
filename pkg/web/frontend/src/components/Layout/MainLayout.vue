<template>
  <div class="flex h-screen bg-makoclaw-bg text-makoclaw-text overflow-hidden">
    <!-- Sidebar -->
    <Sidebar />

    <!-- Toast Notifications -->
    <ToastContainer />

    <!-- Main Content Area -->
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden relative">

      <!-- Mobile Header -->
      <header class="md:hidden h-14 bg-makoclaw-surface border-b border-makoclaw-border flex items-center justify-between px-4 flex-shrink-0 z-sticky safe-top">
         <div class="font-bold text-lg bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent">MakoClaw</div>
         <button @click="uiStore.toggleSidebar()" class="p-2 text-makoclaw-text hover:bg-makoclaw-surface-hover rounded-lg transition-colors touch-target">
           <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
             <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16m-7 6h7" />
           </svg>
         </button>
      </header>

      <!-- Degraded Mode Banner -->
      <DegradedModeBanner />

      <!-- Page Content -->
      <main class="flex-1 overflow-auto relative scroll-smooth custom-scrollbar">
        <router-view v-slot="{ Component }">
          <Transition name="page" mode="out-in">
            <component :is="Component" :key="$route.path" />
          </Transition>
        </router-view>
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
import DegradedModeBanner from '../DegradedModeBanner.vue'

const uiStore = useUIStore()
const chatStore = useChatStore()
const chatWs = getChatWebSocket()

// Flag set by ChatView when it's mounted and managing WS messages directly.
// When true, this background listener skips queueing (ChatView handles messages itself).
let chatViewActive = false

const setChatViewActive = (active) => { chatViewActive = active }

// Expose globally so ChatView can signal its mount/unmount status
window.__makoclaw_setChatViewActive = setChatViewActive

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
