<template>
  <div class="flex justify-center py-1.5">
    <div
      class="inline-flex items-center gap-2 px-3 py-1 bg-makoclaw-surface/30 border border-makoclaw-border/20 rounded-full text-xs text-makoclaw-text-secondary"
    >
      <!-- Status dot -->
      <span
        class="w-2 h-2 rounded-full flex-shrink-0"
        :class="dotColor"
      />

      <!-- Event label -->
      <span class="truncate max-w-xs">{{ label }}</span>

      <!-- Timestamp -->
      <span class="opacity-40 text-[10px] flex-shrink-0">{{ formattedTime }}</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  event: {
    type: Object,
    required: true
  }
})

const dotColor = computed(() => {
  const colorMap = {
    delegating: 'bg-purple-400',
    working: 'bg-emerald-400 animate-pulse',
    complete: 'bg-emerald-400',
    fallback: 'bg-amber-400',
    requesting_help: 'bg-amber-400 animate-pulse',
    colleague_complete: 'bg-teal-400',
    max_delegations_reached: 'bg-amber-500',
    timeout: 'bg-red-400',
    synthesizing: 'bg-blue-400 animate-pulse'
  }
  return colorMap[props.event.status] || 'bg-makoclaw-text-secondary'
})

const label = computed(() => {
  const { agent, status, specialistName, parentAgent } = props.event
  const agentLabel = agent || 'agent'
  const targetLabel = specialistName || ''
  const viaLabel = parentAgent ? ` (via ${parentAgent})` : ''

  switch (status) {
    case 'delegating':
      return `${agentLabel} delegated to ${targetLabel}`
    case 'working':
      return `${agentLabel} is working...${viaLabel}`
    case 'complete':
      return `${agentLabel} finished${viaLabel}`
    case 'fallback':
      return `Falling back to ${targetLabel}`
    case 'requesting_help':
      return `${agentLabel} requested help from ${targetLabel}`
    case 'colleague_complete':
      return `${agentLabel} completed assistance${viaLabel}`
    case 'max_delegations_reached':
      return 'Maximum delegations reached — synthesizing results'
    case 'timeout':
      return `${agentLabel} timed out`
    case 'synthesizing':
      return 'Synthesizing final response...'
    default:
      return `${agentLabel}: ${status}`
  }
})

const formattedTime = computed(() => {
  if (!props.event.timestamp) return ''
  const date = new Date(props.event.timestamp)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
})
</script>

