<template>
  <transition name="slide-fade">
    <div
      v-if="isActive"
      class="glass-panel rounded-lg px-4 py-3 mb-4 flex items-center gap-3 border-l-4"
      :class="borderColorClass"
    >
      <!-- Spinner animado -->
      <div class="relative w-8 h-8">
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

      <!-- Texto de estado -->
      <div class="flex-1 min-w-0">
        <!-- Delegation chain breadcrumb -->
        <div
          v-if="displayChain.length > 0"
          class="flex items-center gap-1 flex-wrap mb-1"
        >
          <template
            v-for="(agent, idx) in displayChain"
            :key="idx"
          >
            <span
              class="text-[10px] px-1.5 py-0.5 rounded capitalize font-medium"
              :class="idx === displayChain.length - 1
                ? 'bg-emerald-500/20 text-emerald-400 ring-1 ring-emerald-500/30'
                : 'bg-makoclaw-surface/50 text-makoclaw-text-secondary'"
            >
              {{ agent }}
            </span>
            <svg
              v-if="idx < displayChain.length - 1"
              class="w-3 h-3 text-makoclaw-text-secondary/40 flex-shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 5l7 7-7 7"
              />
            </svg>
          </template>
        </div>

        <div class="flex items-center gap-2">
          <SpecialistBadge
            v-if="currentAgent && displayChain.length === 0"
            :name="currentAgent"
            class="text-xs"
          />
          <span
            class="text-sm font-medium"
            :class="textColorClass"
          >
            {{ statusText }}
          </span>
        </div>

        <!-- Razón de delegación -->
        <p
          v-if="delegationReason"
          class="text-xs text-makoclaw-text-secondary mt-1 italic truncate"
        >
          {{ delegationReason }}
        </p>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { computed } from 'vue'
import { useChatStore } from '../../stores/chatStore'
import SpecialistBadge from './SpecialistBadge.vue'

const chatStore = useChatStore()

// Show indicator when:
// 1. Orchestrator has explicit status (analyzing, delegating, working, fallback)
// 2. OR when loading/streaming (as a fallback to ensure visibility)
const hasOrchestratorStatus = computed(() => {
  return chatStore.orchestratorStatus !== 'idle' && chatStore.orchestratorStatus !== 'complete'
})

const isProcessing = computed(() => {
  return chatStore.globalIsLoading || chatStore.isStreaming
})

const isActive = computed(() => {
  // Show if orchestrator has status OR if we're processing (ensures always visible when working)
  return hasOrchestratorStatus.value || (isProcessing.value && !chatStore.isStreaming)
})

const currentAgent = computed(() => {
  return chatStore.currentAgent || chatStore.activeSpecialist
})

const displayChain = computed(() => {
  return chatStore.delegationChain || []
})

const delegationReason = computed(() => chatStore.delegationReason)

const statusText = computed(() => {
  const status = chatStore.orchestratorStatus
  const agent = currentAgent.value

  // If we have explicit orchestrator status, use it
  if (hasOrchestratorStatus.value) {
    const statusMap = {
      analyzing: 'Analyzing your request...',
      delegating: `Delegating to ${agent || 'specialist'}...`,
      working: `${agent || 'Agent'} is working on your request...`,
      fallback: `No specialist matched, using ${agent || 'general agent'}...`,
      max_delegations_reached: 'Synthesizing response...',
      timeout: `${agent || 'Agent'} timed out`,
      synthesizing: 'Synthesizing final response...'
    }
    return statusMap[status] || 'Processing...'
  }

  // Fallback: show generic processing message
  return 'Agent is processing your request...'
})

const iconColorClass = computed(() => {
  const status = chatStore.orchestratorStatus
  if (hasOrchestratorStatus.value) {
    return {
      analyzing: 'text-blue-400',
      delegating: 'text-purple-400',
      working: 'text-green-400',
      fallback: 'text-amber-400',
      max_delegations_reached: 'text-amber-400',
      timeout: 'text-red-400',
      synthesizing: 'text-blue-400'
    }[status] || 'text-makoclaw-text-secondary'
  }
  // Default color for generic processing
  return 'text-blue-400'
})

const textColorClass = computed(() => iconColorClass.value)

const borderColorClass = computed(() => {
  const status = chatStore.orchestratorStatus
  if (hasOrchestratorStatus.value) {
    return {
      analyzing: 'border-blue-400',
      delegating: 'border-purple-400',
      working: 'border-green-400',
      fallback: 'border-amber-400',
      max_delegations_reached: 'border-amber-400',
      timeout: 'border-red-400',
      synthesizing: 'border-blue-400'
    }[status] || 'border-makoclaw-text-secondary'
  }
  // Default border for generic processing
  return 'border-blue-400'
})
</script>

<style scoped>
.slide-fade-enter-active {
  transition: all 0.3s ease-out;
}

.slide-fade-leave-active {
  transition: all 0.2s cubic-bezier(1, 0.5, 0.8, 1);
}

.slide-fade-enter-from {
  transform: translateY(-10px);
  opacity: 0;
}

.slide-fade-leave-to {
  transform: translateY(10px);
  opacity: 0;
}
</style>

