<template>
  <Transition name="fade">
    <div
      v-if="isSynthesizing"
      data-testid="synthesis-indicator"
      class="glass-panel rounded-xl px-3.5 py-3 mb-3 border border-emerald-500/15 shadow-lg shadow-emerald-500/10"
    >
      <div class="flex items-center gap-3">
        <div class="relative w-9 h-9 rounded-xl bg-emerald-500/15 text-emerald-400 ring-1 ring-emerald-500/20 flex items-center justify-center flex-shrink-0">
          <svg
            class="w-5 h-5 animate-pulse"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.8"
              d="M9.5 3A2.5 2.5 0 007 5.5V7H5.5A2.5 2.5 0 003 9.5v5A2.5 2.5 0 005.5 17H7v1.5A2.5 2.5 0 009.5 21h5a2.5 2.5 0 002.5-2.5V17h1.5a2.5 2.5 0 002.5-2.5v-5A2.5 2.5 0 0018.5 7H17V5.5A2.5 2.5 0 0014.5 3h-5zM9 7h6V5.5a.5.5 0 00-.5-.5h-5a.5.5 0 00-.5.5V7zm1.5 7h3m-4.5 3h6"
            />
          </svg>
        </div>

        <div class="min-w-0 flex-1">
          <p class="text-sm font-semibold text-emerald-400">
            {{ agentLabel }} synthesizing...
          </p>
          <div class="flex items-center gap-1.5 mt-1 text-[11px] text-makoclaw-text-secondary">
            <span>Combining specialist outputs</span>
            <div class="flex items-center gap-1">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse [animation-delay:150ms]" />
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse [animation-delay:300ms]" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  isSynthesizing: {
    type: Boolean,
    default: false
  },
  agent: {
    type: String,
    default: 'orchestrator'
  }
})

const agentLabel = computed(() => {
  const value = props.agent || 'orchestrator'
  return value.charAt(0).toUpperCase() + value.slice(1)
})
</script>
