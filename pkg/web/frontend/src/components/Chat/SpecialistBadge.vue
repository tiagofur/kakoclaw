<template>
  <span
    v-if="name"
    :class="`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-medium ${bgColor} ${textColor} border border-current/20`"
    :title="tooltip"
  >
    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" v-html="icon"></svg>
    <span class="capitalize">{{ name }}</span>
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  name: {
    type: String,
    default: null
  }
})

const name = computed(() => props.name)
const icon = computed(() => {
  const icons = {
    developer: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />',
    documentation: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />',
    testing: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />',
    devops: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />',
    analyst: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 8v8m-4-5v5m-4-2v2m-2 4h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />',
    researcher: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />',
    general: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />',
    orchestrator: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />'
  }
  return icons[props.name] || '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />'
})

const colors = computed(() => {
  const colorMap = {
    developer: { color: 'text-blue-400', bg: 'bg-blue-500/10' },
    documentation: { color: 'text-blue-400', bg: 'bg-blue-500/10' },
    testing: { color: 'text-amber-400', bg: 'bg-amber-500/10' },
    devops: { color: 'text-purple-400', bg: 'bg-purple-500/10' },
    analyst: { color: 'text-cyan-400', bg: 'bg-cyan-500/10' },
    researcher: { color: 'text-pink-400', bg: 'bg-pink-500/10' },
    general: { color: 'text-emerald-400', bg: 'bg-emerald-500/10' },
    orchestrator: { color: 'text-indigo-400', bg: 'bg-indigo-500/10' }
  }
  return colorMap[props.name] || { color: 'text-makoclaw-text-secondary', bg: 'bg-makoclaw-text-secondary/10' }
})

const textColor = computed(() => colors.value.color)
const bgColor = computed(() => colors.value.bg)
const tooltip = computed(() => `Specialist: ${props.name}`)
</script>

