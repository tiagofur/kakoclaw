<template>
  <div class="flex flex-col h-full bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-purple-500/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-indigo-500/20 via-transparent to-transparent" />
    </div>

    <!-- Page Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <div class="flex items-center gap-3">
          <!-- Icon Container -->
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-purple-500/20 to-indigo-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-purple-500/10">
            <svg
              class="w-5 h-5 sm:w-6 sm:h-6 text-purple-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9.53 16.122a3 3 0 00-5.78 1.128 2.25 2.25 0 01-2.4 2.245 4.5 4.5 0 008.4-2.245c0-.399-.078-.78-.22-1.128zm0 0a15.998 15.998 0 003.388-1.62m-5.043-.025a15.994 15.994 0 011.622-3.395m3.42 3.42a15.995 15.995 0 004.764-4.648l3.876-5.814a1.151 1.151 0 00-1.597-1.597L14.146 6.32a15.996 15.996 0 00-4.649 4.763m3.42 3.42a6.776 6.776 0 00-3.42-3.42"
              />
            </svg>
          </div>

          <!-- Title -->
          <div>
            <h1 class="text-lg sm:text-xl font-bold text-white">Canvas</h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary">Agent-controlled visual workspace</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Canvas Content -->
    <div class="flex-1 relative z-10 p-4 sm:p-6">
      <div v-if="!connected" class="flex items-center justify-center h-full">
        <div class="text-center glass-card p-8 rounded-2xl">
          <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-purple-500/20 flex items-center justify-center">
            <svg class="w-8 h-8 text-purple-400 animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <p class="text-makoclaw-text-secondary">Connecting to canvas...</p>
        </div>
      </div>

      <div v-else-if="!canvasHTML" class="flex items-center justify-center h-full">
        <div class="text-center glass-card p-8 rounded-2xl">
          <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-gray-500/20 flex items-center justify-center">
            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
          <p class="text-makoclaw-text-secondary">Canvas is empty. The agent will push content here.</p>
        </div>
      </div>

      <iframe
        v-else
        sandbox="allow-scripts"
        :srcdoc="canvasHTML"
        class="w-full h-full rounded-xl border border-makoclaw-border/20 bg-white"
        title="MakoClaw Canvas"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const canvasHTML = ref('')
const connected = ref(false)
let eventSource = null

onMounted(() => {
  eventSource = new EventSource('/__makoclaw__/canvas/events')

  eventSource.onopen = () => {
    connected.value = true
  }

  eventSource.onmessage = (event) => {
    connected.value = true
    canvasHTML.value = event.data
  }

  eventSource.onerror = () => {
    connected.value = false
  }
})

onUnmounted(() => {
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
})
</script>
