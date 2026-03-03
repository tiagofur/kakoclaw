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

        <!-- Specialist Report with Confidence -->
        <div
          v-if="lastReport"
          class="p-2.5 rounded-xl border"
          :class="lastReport.request_help ? 'bg-red-500/10 border-red-500/30' : 'bg-makoclaw-surface/30 border-makoclaw-border/20'"
        >
          <div class="flex items-center justify-between mb-2">
            <div class="flex items-center gap-2">
              <span class="text-xs font-semibold text-makoclaw-text capitalize">
                {{ lastReport.specialist_name }}
              </span>
              <span
                class="text-[10px] px-1.5 py-0.5 rounded border"
                :class="reportStatusClass"
              >
                {{ lastReport.status }}
              </span>
            </div>
            <!-- Confidence Indicator -->
            <div v-if="lastReport.confidence" class="flex items-center gap-1.5">
              <span class="text-[10px] text-makoclaw-text-secondary">Confidence</span>
              <div class="flex items-center gap-1">
                <div class="w-16 h-1.5 bg-makoclaw-bg/50 rounded-full overflow-hidden">
                  <div
                    class="h-full rounded-full transition-all duration-500"
                    :class="confidenceClass"
                    :style="{ width: `${lastReport.confidence * 100}%` }"
                  />
                </div>
                <span
                  class="text-[10px] font-medium"
                  :class="confidenceTextClass"
                >
                  {{ Math.round(lastReport.confidence * 100) }}%
                </span>
              </div>
            </div>
          </div>
          <!-- Help Request Alert -->
          <div
            v-if="lastReport.request_help"
            class="flex items-center gap-2 p-2 rounded-lg bg-red-500/10 border border-red-500/20 mt-2"
          >
            <svg class="w-4 h-4 text-red-400 animate-pulse shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div class="flex-1 min-w-0">
              <span class="text-[11px] text-red-400 font-medium">
                Requesting help from {{ lastReport.request_help }}
              </span>
              <p v-if="lastReport.suggestions?.[0]" class="text-[10px] text-red-300/70 truncate">
                {{ lastReport.suggestions[0] }}
              </p>
            </div>
          </div>
        </div>

        <!-- Delegation Chain Tree -->
        <div
          v-if="props.delegationChain && props.delegationChain.length > 1"
          class="p-2.5 rounded-xl bg-makoclaw-surface/30 border border-makoclaw-border/20"
        >
          <h4 class="text-[10px] font-medium text-makoclaw-text-secondary uppercase tracking-wider mb-2">
            Delegation Chain
          </h4>
          <div class="space-y-1">
            <div
              v-for="(agent, idx) in props.delegationChain"
              :key="idx"
              class="flex items-center gap-2"
              :style="{ paddingLeft: `${idx * 12}px` }"
            >
              <span
                v-if="idx > 0"
                class="text-makoclaw-text-secondary/30"
              >
                └
              </span>
              <span
                class="text-[10px] px-1.5 py-0.5 rounded capitalize font-medium"
                :class="idx === props.delegationChain.length - 1
                  ? 'bg-emerald-500/20 text-emerald-400 ring-1 ring-emerald-500/30'
                  : 'bg-makoclaw-surface/50 text-makoclaw-text-secondary'"
              >
                {{ agent }}
              </span>
              <span
                v-if="idx === props.delegationChain.length - 1 && props.activeDelegation"
                class="text-[9px] text-green-400 animate-pulse"
              >
                ● active
              </span>
            </div>
          </div>
          <!-- Active delegation progress -->
          <div
            v-if="props.activeDelegation"
            class="mt-2 flex items-center gap-2 text-[10px] text-makoclaw-text-secondary"
          >
            <span>{{ props.activeDelegation.from }} → {{ props.activeDelegation.to }}</span>
            <span v-if="props.activeDelegation.elapsedMs" class="opacity-60">
              {{ Math.round(props.activeDelegation.elapsedMs / 1000) }}s
            </span>
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
  },
  specialistReport: {
    type: Object,
    default: null
  },
  delegationChain: {
    type: Array,
    default: () => []
  },
  activeDelegation: {
    type: Object,
    default: null
  }
})

const collapsed = ref(false)
const currentAgent = ref(null)
const communications = ref([])
const involvedAgents = ref([])
const lastReport = ref(null)

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

// Watch for specialist report
watch(() => props.specialistReport, (newReport) => {
  if (newReport) {
    lastReport.value = newReport
    // Add specialist to involved agents
    if (newReport.specialist_name && !involvedAgents.value.includes(newReport.specialist_name)) {
      involvedAgents.value.push(newReport.specialist_name)
    }
    // If requesting help, add to communications
    if (newReport.request_help) {
      communications.value.push({
        from: newReport.specialist_name,
        to: newReport.request_help,
        message: `Requesting help: ${newReport.suggestions?.[0] || 'assistance needed'}`
      })
    }
  }
}, { deep: true })

const hasActivity = computed(() => {
  return currentAgent.value || communications.value.length > 0 || involvedAgents.value.length > 0 || lastReport.value
})

// Confidence level indicator (green >0.8, yellow 0.5-0.8, red <0.5)
const confidenceColor = computed(() => {
  if (!lastReport.value?.confidence) return null
  const conf = lastReport.value.confidence
  if (conf >= 0.8) return 'green'
  if (conf >= 0.5) return 'yellow'
  return 'red'
})

const confidenceClass = computed(() => {
  const colors = {
    green: 'bg-green-500',
    yellow: 'bg-yellow-500',
    red: 'bg-red-500'
  }
  return colors[confidenceColor.value] || 'bg-gray-500'
})

const confidenceTextClass = computed(() => {
  const colors = {
    green: 'text-green-400',
    yellow: 'text-yellow-400',
    red: 'text-red-400'
  }
  return colors[confidenceColor.value] || 'text-gray-400'
})

const reportStatusClass = computed(() => {
  if (!lastReport.value?.status) return ''
  const classes = {
    complete: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    partial: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
    needs_help: 'bg-red-500/20 text-red-400 border-red-500/30 animate-pulse'
  }
  return classes[lastReport.value.status] || 'bg-gray-500/20 text-gray-400'
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
    max_delegations_reached: 'bg-amber-500/20 text-amber-400',
    timeout: 'bg-red-500/20 text-red-400',
    synthesizing: 'bg-blue-500/20 text-blue-400',
  }
  return classes[status] || 'bg-makoclaw-surface/40 text-makoclaw-text-secondary'
}

// Reset state for new conversation
function reset() {
  currentAgent.value = null
  communications.value = []
  involvedAgents.value = []
  lastReport.value = null
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
