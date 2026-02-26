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
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
      </div>

      <!-- Texto de estado -->
      <div class="flex-1">
        <div class="flex items-center gap-2">
          <SpecialistBadge
            v-if="currentAgent"
            :name="currentAgent"
            class="text-xs"
          />
          <span class="text-sm font-medium" :class="textColorClass">
            {{ statusText }}
          </span>
        </div>

        <!-- Razón de delegación -->
        <p
          v-if="delegationReason"
          class="text-xs text-makoclaw-text-secondary mt-1 italic"
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

const isActive = computed(() => {
  return chatStore.orchestratorStatus !== 'idle' && chatStore.orchestratorStatus !== 'complete'
})

const currentAgent = computed(() => {
  return chatStore.currentAgent || chatStore.activeSpecialist
})

const delegationReason = computed(() => chatStore.delegationReason)

const statusText = computed(() => {
  const status = chatStore.orchestratorStatus
  const agent = currentAgent.value

  const statusMap = {
    analyzing: `Analyzing your request...`,
    delegating: `Delegating to ${agent || 'specialist'}...`,
    working: `${agent || 'Agent'} is working on your request...`
  }

  return statusMap[status] || 'Processing...'
})

const iconColorClass = computed(() => {
  const status = chatStore.orchestratorStatus
  return {
    analyzing: 'text-blue-400',
    delegating: 'text-purple-400',
    working: 'text-green-400'
  }[status] || 'text-makoclaw-text-secondary'
})

const textColorClass = computed(() => iconColorClass.value)

const borderColorClass = computed(() => {
  const status = chatStore.orchestratorStatus
  return {
    analyzing: 'border-blue-400',
    delegating: 'border-purple-400',
    working: 'border-green-400'
  }[status] || 'border-makoclaw-text-secondary'
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
