<template>
  <div
    v-if="hasActivity"
    class="glass-panel rounded-2xl p-4 mb-4 relative overflow-hidden group/panel"
  >
    <!-- Animated glow effect -->
    <div
      class="absolute -top-12 -right-12 w-32 h-32 blur-[40px] rounded-full pointer-events-none transition-all duration-700"
      :class="isWorking ? 'bg-green-500/20 animate-pulse' : 'bg-makoclaw-accent/10'"
    />

    <div class="flex items-center justify-between mb-3">
      <h3 class="text-sm font-semibold text-makoclaw-text flex items-center gap-2">
        <svg
          class="w-4 h-4"
          :class="isWorking ? 'text-green-400 animate-spin' : ''"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 10V3L4 14h7v7l9-11h-7z"
          />
        </svg>
        Team Activity
        <span
          v-if="isWorking"
          class="ml-2 text-[10px] px-2 py-0.5 rounded-full bg-green-500/20 text-green-400 animate-pulse"
        >
          LIVE
        </span>
      </h3>

      <button
        class="p-1 rounded hover:bg-makoclaw-bg/50 transition-colors"
        @click="collapsed = !collapsed"
      >
        <svg
          class="w-4 h-4 transition-transform"
          :class="{ 'rotate-180': !collapsed }"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </button>
    </div>

    <transition name="expand">
      <div v-if="!collapsed" class="space-y-3">
        <!-- Current Active Agent -->
        <div
          v-if="currentAgent"
          class="flex items-center gap-3 p-2.5 rounded-xl bg-makoclaw-surface/30 border border-makoclaw-border/20"
        >
          <div
            class="w-8 h-8 rounded-lg flex items-center justify-center"
            :class="getAgentColor(currentAgent.agent)"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="text-xs font-semibold text-makoclaw-text capitalize">
                {{ currentAgent.agent }}
              </span>
              <span
                class="text-[10px] px-1.5 py-0.5 rounded"
                :class="getStatusClass(currentAgent.status)"
              >
                {{ currentAgent.status }}
              </span>
            </div>
            <p
              v-if="currentAgent.reason"
              class="text-[11px] text-makoclaw-text-secondary truncate mt-0.5"
            >
              {{ currentAgent.reason }}
            </p>
          </div>
        </div>

        <!-- Communications Log -->
        <div v-if="communications.length > 0" class="space-y-2">
          <h4 class="text-[10px] font-medium text-makoclaw-text-secondary uppercase tracking-wider">
            Team Communications
          </h4>
          <div class="space-y-1.5 max-h-40 overflow-y-auto">
            <div
              v-for="(comm, index) in communications.slice(-5)"
              :key="index"
              class="flex items-start gap-2 p-2 rounded-lg bg-makoclaw-bg/30 text-xs"
            >
              <span
                class="font-medium shrink-0"
                :class="getAgentTextColor(comm.from)"
              >
                {{ comm.from }}
              </span>
              <svg class="w-3 h-3 text-makoclaw-text-secondary shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
              </svg>
              <span
                class="font-medium shrink-0"
                :class="getAgentTextColor(comm.to)"
              >
                {{ comm.to }}
              </span>
              <span class="text-makoclaw-text-secondary truncate">
                {{ comm.message }}
              </span>
            </div>
          </div>
        </div>

        <!-- Involved Agents Summary -->
        <div v-if="involvedAgents.length > 0" class="pt-2 border-t border-makoclaw-border/20">
          <h4 class="text-[10px] font-medium text-makoclaw-text-secondary uppercase tracking-wider mb-2">
            Involved Specialists
          </h4>
          <div class="flex flex-wrap gap-1.5">
            <span
              v-for="agent in involvedAgents"
              :key="agent"
              class="text-[10px] px-2 py-1 rounded-lg capitalize"
              :class="getAgentBadgeClass(agent)"
            >
              {{ agent }}
            </span>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  agentStatus: {
    type: Object,
    default: null
  },
  teamCommunications: {
    type: Array,
    default: () => []
  },
  involvedAgentsList: {
    type: Array,
    default: () => []
  }
})

const collapsed = ref(false)
const currentAgent = ref(null)
const communications = ref([])
const involvedAgents = ref([])

// Watch for agent status updates
watch(() => props.agentStatus, (newStatus) => {
  if (newStatus) {
    currentAgent.value = newStatus
    // Add to involved agents if not already there
    if (newStatus.agent && !involvedAgents.value.includes(newStatus.agent)) {
      involvedAgents.value.push(newStatus.agent)
    }
  }
}, { deep: true })

// Watch for communications updates
watch(() => props.teamCommunications, (newComms) => {
  if (newComms && newComms.length > 0) {
    communications.value = [...newComms]
  }
}, { deep: true })

// Watch for involved agents list
watch(() => props.involvedAgentsList, (newList) => {
  if (newList && newList.length > 0) {
    involvedAgents.value = [...new Set([...involvedAgents.value, ...newList])]
  }
}, { deep: true })

const hasActivity = computed(() => {
  return currentAgent.value || communications.value.length > 0 || involvedAgents.value.length > 0
})

const isWorking = computed(() => {
  return currentAgent.value?.status === 'working' ||
         currentAgent.value?.status === 'analyzing' ||
         currentAgent.value?.status === 'delegating' ||
         currentAgent.value?.status === 'requesting_help'
})

function getAgentColor(agent) {
  const colors = {
    orchestrator: 'bg-purple-500/20 text-purple-400',
    developer: 'bg-blue-500/20 text-blue-400',
    security: 'bg-red-500/20 text-red-400',
    documentation: 'bg-green-500/20 text-green-400',
    researcher: 'bg-amber-500/20 text-amber-400',
    tester: 'bg-cyan-500/20 text-cyan-400',
  }
  return colors[agent] || 'bg-makoclaw-surface/40 text-makoclaw-text'
}

function getAgentTextColor(agent) {
  const colors = {
    orchestrator: 'text-purple-400',
    developer: 'text-blue-400',
    security: 'text-red-400',
    documentation: 'text-green-400',
    researcher: 'text-amber-400',
    tester: 'text-cyan-400',
  }
  return colors[agent] || 'text-makoclaw-text'
}

function getAgentBadgeClass(agent) {
  const classes = {
    orchestrator: 'bg-purple-500/20 text-purple-400 border border-purple-500/30',
    developer: 'bg-blue-500/20 text-blue-400 border border-blue-500/30',
    security: 'bg-red-500/20 text-red-400 border border-red-500/30',
    documentation: 'bg-green-500/20 text-green-400 border border-green-500/30',
    researcher: 'bg-amber-500/20 text-amber-400 border border-amber-500/30',
    tester: 'bg-cyan-500/20 text-cyan-400 border border-cyan-500/30',
  }
  return classes[agent] || 'bg-makoclaw-surface/40 text-makoclaw-text border border-makoclaw-border/30'
}

function getStatusClass(status) {
  const classes = {
    working: 'bg-green-500/20 text-green-400',
    analyzing: 'bg-blue-500/20 text-blue-400',
    delegating: 'bg-purple-500/20 text-purple-400',
    requesting_help: 'bg-amber-500/20 text-amber-400',
    complete: 'bg-emerald-500/20 text-emerald-400',
    colleague_complete: 'bg-emerald-500/20 text-emerald-400',
    fallback: 'bg-orange-500/20 text-orange-400',
    failed: 'bg-red-500/20 text-red-400',
  }
  return classes[status] || 'bg-makoclaw-surface/40 text-makoclaw-text-secondary'
}

// Reset state for new conversation
function reset() {
  currentAgent.value = null
  communications.value = []
  involvedAgents.value = []
}

// Expose reset method
defineExpose({ reset })
</script>

<style scoped>
.expand-enter-active,
.expand-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
}

.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  max-height: 400px;
}
</style>
