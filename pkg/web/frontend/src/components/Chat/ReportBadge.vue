<template>
  <div class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-makoclaw-surface/40 border border-makoclaw-border/30 rounded-lg text-[11px] text-makoclaw-text-secondary max-w-full">
    <svg class="w-3 h-3 text-emerald-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
        d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
    </svg>
    <span class="font-medium text-makoclaw-text truncate">{{ title }}</span>
    <span class="text-makoclaw-text-secondary/50">→</span>
    <span class="font-mono text-[10px] text-makoclaw-text-secondary truncate">{{ displayPath }}</span>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  title: {
    type: String,
    required: true
  },
  filePath: {
    type: String,
    required: true
  }
})

const displayPath = computed(() => {
  // Show just the filename part for brevity
  const parts = props.filePath.replace(/\\/g, '/').split('/')
  return parts[parts.length - 1] || props.filePath
})
</script>
