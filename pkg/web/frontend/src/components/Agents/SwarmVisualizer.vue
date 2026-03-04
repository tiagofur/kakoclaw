<template>
  <div class="swarm-visualizer p-4">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <h4 class="text-sm font-bold text-makoclaw-text flex items-center gap-2">
        <svg class="w-4 h-4 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        {{ swarmName }}
        <span :class="[modeColor, 'text-xs px-2 py-0.5 rounded-full', modeBgColor]">{{ mode }}</span>
      </h4>
      <span v-if="status" :class="statusColor" class="text-xs font-medium">{{ status }}</span>
    </div>

    <!-- Flow Diagram -->
    <div class="flex items-center gap-2 overflow-x-auto pb-2">
      <div
        v-for="(member, idx) in members"
        :key="member"
        class="flex items-center gap-2 shrink-0"
      >
        <!-- Member Node -->
        <div
          :class="[
            'px-3 py-2 rounded-lg border text-xs font-medium transition-all duration-300',
            getMemberNodeClasses(member)
          ]"
        >
          <div class="flex items-center gap-1.5">
            <div :class="['w-2 h-2 rounded-full', getMemberDotColor(member)]" />
            <span>{{ member }}</span>
          </div>
          <div v-if="getMemberStatus(member)" class="text-[10px] mt-1 opacity-70">
            {{ getMemberStatus(member) }}
          </div>
        </div>

        <!-- Arrow between nodes -->
        <svg
          v-if="idx < members.length - 1 && mode === 'sequential'"
          class="w-6 h-4 text-makoclaw-text-secondary/40 shrink-0"
          viewBox="0 0 24 16"
          fill="none"
          stroke="currentColor"
        >
          <path d="M0 8h20M16 4l4 4-4 4" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </div>
    </div>

    <!-- Parallel/Consensus indicator -->
    <div v-if="mode === 'parallel'" class="mt-2 text-[10px] text-green-400/70 flex items-center gap-1">
      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
      </svg>
      All members run in parallel
    </div>
    <div v-else-if="mode === 'consensus'" class="mt-2 text-[10px] text-purple-400/70 flex items-center gap-1">
      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      Independent execution → consensus synthesis
    </div>

    <!-- Cost summary -->
    <div v-if="totalCost > 0" class="mt-3 text-[10px] text-makoclaw-text-secondary">
      Cost: ${{ totalCost.toFixed(4) }}
      <span v-if="duration"> · {{ duration }}</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  swarmName: { type: String, required: true },
  members: { type: Array, default: () => [] },
  mode: { type: String, default: 'sequential' },
  status: { type: String, default: '' },
  progress: { type: Object, default: () => ({}) },
  totalCost: { type: Number, default: 0 },
  duration: { type: String, default: '' }
})

const modeColor = computed(() => ({
  sequential: 'text-blue-400',
  parallel: 'text-green-400',
  consensus: 'text-purple-400'
}[props.mode] || 'text-gray-400'))

const modeBgColor = computed(() => ({
  sequential: 'bg-blue-500/10',
  parallel: 'bg-green-500/10',
  consensus: 'bg-purple-500/10'
}[props.mode] || 'bg-gray-500/10'))

const statusColor = computed(() => ({
  running: 'text-yellow-400',
  completed: 'text-green-400',
  failed: 'text-red-400'
}[props.status] || 'text-makoclaw-text-secondary'))

function getMemberStatus(member) {
  return props.progress[member] || ''
}

function getMemberNodeClasses(member) {
  const status = props.progress[member]
  if (status === 'completed' || status === 'complete') return 'border-green-500/50 bg-green-500/10 text-green-400'
  if (status === 'running' || status === 'working' || status === 'in_progress') return 'border-yellow-500/50 bg-yellow-500/10 text-yellow-400 animate-pulse'
  if (status === 'failed') return 'border-red-500/50 bg-red-500/10 text-red-400'
  return 'border-makoclaw-border/30 bg-makoclaw-surface/30 text-makoclaw-text-secondary'
}

function getMemberDotColor(member) {
  const status = props.progress[member]
  if (status === 'completed' || status === 'complete') return 'bg-green-400'
  if (status === 'running' || status === 'working' || status === 'in_progress') return 'bg-yellow-400'
  if (status === 'failed') return 'bg-red-400'
  return 'bg-gray-500'
}
</script>
