<template>
  <transition name="slide-fade">
    <div
      v-if="isActive"
      class="rounded-lg px-3 py-2 mb-3 flex items-center gap-2.5 bg-makoclaw-surface/40 border border-makoclaw-border/20"
    >
      <!-- Spinner -->
      <div class="relative w-5 h-5 flex-shrink-0">
        <svg
          class="animate-spin"
          :class="iconColorClass"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          />
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      </div>

      <div class="min-w-0 flex-1">
        <div
          class="text-sm font-medium truncate"
          :class="iconColorClass"
        >
          {{ statusText }}
        </div>
        <div
          v-if="agentLine"
          class="text-[11px] text-makoclaw-text-secondary truncate mt-0.5"
        >
          {{ agentLine }}
        </div>
      </div>

      <!-- Elapsed time -->
      <span
        v-if="elapsedSeconds > 0"
        class="text-[10px] text-makoclaw-text-secondary ml-auto flex-shrink-0 tabular-nums"
      >
        {{ formatElapsed(elapsedSeconds) }}
      </span>
    </div>
  </transition>
</template>

<script setup>
import { computed, ref, watch, onUnmounted } from 'vue'
import { useChatStore } from '../../stores/chatStore'

const chatStore = useChatStore()

const hasOrchestratorStatus = computed(() => {
  return chatStore.orchestratorStatus !== 'idle' && chatStore.orchestratorStatus !== 'complete'
})

const isProcessing = computed(() => {
  return chatStore.globalIsLoading || chatStore.isStreaming
})

const isActive = computed(() => {
  return hasOrchestratorStatus.value || (isProcessing.value && !chatStore.isStreaming)
})

const currentAgent = computed(() => {
  return chatStore.currentAgent || chatStore.activeSpecialist
})

const activeToolCall = computed(() => {
  const toolCalls = chatStore.streamingMessage?.toolCalls || []
  const exactMatch = toolCalls.find(tc => tc.status === 'started' && tc.agentName === currentAgent.value)
  return exactMatch || toolCalls.find(tc => tc.status === 'started') || null
})

const agentLine = computed(() => {
  const agentName = activeToolCall.value?.agentName || currentAgent.value
  const toolName = activeToolCall.value?.name

  if (agentName && toolName) return `${agentName} — ${toolName}`
  return agentName || toolName || ''
})

const statusText = computed(() => {
  const status = chatStore.orchestratorStatus
  const agent = currentAgent.value

  if (hasOrchestratorStatus.value) {
    const statusMap = {
      analyzing: 'Analyzing request...',
      delegating: agent ? `Routing to ${agent}` : 'Routing to specialist...',
      working: agent ? `${agent} working...` : 'Working...',
      fallback: 'Processing with default agent...',
      max_delegations_reached: 'Synthesizing...',
      timeout: agent ? `${agent} timed out` : 'Timed out',
      synthesizing: 'Synthesizing response...'
    }
    return statusMap[status] || 'Processing...'
  }

  return 'Processing...'
})

const iconColorClass = computed(() => {
  const status = chatStore.orchestratorStatus
  if (status === 'timeout') return 'text-red-400'
  if (status === 'working') return 'text-emerald-400'
  if (status === 'delegating') return 'text-purple-400'
  return 'text-blue-400'
})

// Elapsed time tracking
const elapsedSeconds = ref(0)
let timer = null

watch(isActive, (active) => {
  if (active) {
    elapsedSeconds.value = 0
    timer = setInterval(() => {
      elapsedSeconds.value++
    }, 1000)
  } else {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
    elapsedSeconds.value = 0
  }
}, { immediate: true })

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

function formatElapsed(seconds) {
  if (seconds < 60) return `${seconds}s`
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}m ${secs}s`
}
</script>

<style scoped>
.slide-fade-enter-active {
  transition: all 0.2s ease-out;
}

.slide-fade-leave-active {
  transition: all 0.15s ease-in;
}

.slide-fade-enter-from {
  transform: translateY(-6px);
  opacity: 0;
}

.slide-fade-leave-to {
  transform: translateY(6px);
  opacity: 0;
}
</style>
