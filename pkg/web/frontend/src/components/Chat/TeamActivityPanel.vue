<template>
  <div
    v-if="isMultiAgent"
    class="glass-panel rounded-xl p-3 mb-3 relative overflow-hidden"
  >
    <!-- Header - clickable to expand/collapse -->
    <button
      class="flex items-center justify-between w-full text-left"
      @click="collapsed = !collapsed"
    >
      <div class="flex items-center gap-2">
        <svg
          class="w-3.5 h-3.5"
          :class="isWorking ? 'text-emerald-400' : 'text-makoclaw-text-secondary'"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"
          />
        </svg>
        <span class="text-xs font-medium text-makoclaw-text">
          Team Activity
        </span>
        <span
          v-if="isWorking"
          class="text-[9px] px-1.5 py-0.5 rounded-full bg-emerald-500/15 text-emerald-400"
        >
          LIVE
        </span>
        <!-- Compact agent badges when collapsed -->
        <div v-if="collapsed" class="flex gap-1 ml-1">
          <span
            v-for="agent in involvedAgents.slice(0, 3)"
            :key="agent"
            class="text-[9px] px-1.5 py-0.5 rounded capitalize bg-makoclaw-surface/50 text-makoclaw-text-secondary"
          >
            {{ agent }}
          </span>
          <span
            v-if="involvedAgents.length > 3"
            class="text-[9px] text-makoclaw-text-secondary"
          >
            +{{ involvedAgents.length - 3 }}
          </span>
        </div>
      </div>
      <svg
        class="w-3.5 h-3.5 text-makoclaw-text-secondary transition-transform duration-200"
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

    <transition name="expand">
      <div v-if="!collapsed" class="mt-3 space-y-2.5">
        <!-- Delegation Chain - compact inline view -->
        <div
          v-if="props.delegationChain && props.delegationChain.length > 1"
          class="flex items-center gap-1 flex-wrap"
        >
          <template
            v-for="(agent, idx) in props.delegationChain"
            :key="idx"
          >
            <span
              class="text-[10px] px-1.5 py-0.5 rounded capitalize font-medium"
              :class="idx === props.delegationChain.length - 1
                ? 'bg-emerald-500/20 text-emerald-400'
                : 'bg-makoclaw-surface/50 text-makoclaw-text-secondary'"
            >
              {{ agent }}
            </span>
            <svg
              v-if="idx < props.delegationChain.length - 1"
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
          <span
            v-if="props.activeDelegation?.elapsedMs"
            class="text-[9px] text-makoclaw-text-secondary/60 ml-1 tabular-nums"
          >
            {{ Math.round(props.activeDelegation.elapsedMs / 1000) }}s
          </span>
        </div>

        <!-- Specialist Report with Confidence - only show when complete or needs help -->
        <div
          v-if="lastReport && (lastReport.status !== 'working')"
          class="p-2 rounded-lg border text-xs"
          :class="lastReport.request_help ? 'bg-red-500/10 border-red-500/20' : 'bg-makoclaw-surface/30 border-makoclaw-border/20'"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="font-medium text-makoclaw-text capitalize">
                {{ lastReport.specialist_name }}
              </span>
              <span
                class="text-[10px] px-1.5 py-0.5 rounded"
                :class="reportStatusClass"
              >
                {{ lastReport.status }}
              </span>
            </div>
            <div v-if="lastReport.confidence" class="flex items-center gap-1">
              <div class="w-12 h-1 bg-makoclaw-bg/50 rounded-full overflow-hidden">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :class="confidenceClass"
                  :style="{ width: `${lastReport.confidence * 100}%` }"
                />
              </div>
              <span class="text-[9px] tabular-nums" :class="confidenceTextClass">
                {{ Math.round(lastReport.confidence * 100) }}%
              </span>
            </div>
          </div>
          <!-- Help Request -->
          <p
            v-if="lastReport.request_help"
            class="text-[10px] text-red-400 mt-1"
          >
            Requesting help from {{ lastReport.request_help }}
          </p>
        </div>

        <!-- Communications Log - compact -->
        <div v-if="communications.length > 0" class="space-y-1">
          <div
            v-for="(comm, index) in communications.slice(-3)"
            :key="index"
            class="flex items-center gap-1.5 text-[10px] text-makoclaw-text-secondary"
          >
            <span class="font-medium capitalize" :class="getAgentTextColor(comm.from)">
              {{ comm.from }}
            </span>
            <span class="opacity-40">&rarr;</span>
            <span class="font-medium capitalize" :class="getAgentTextColor(comm.to)">
              {{ comm.to }}
            </span>
            <span class="truncate opacity-70">{{ comm.message }}</span>
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

const collapsed = ref(true) // Start collapsed
const currentAgent = ref(null)
const communications = ref([])
const involvedAgents = ref([])
const lastReport = ref(null)

// Only show panel when there are 2+ agents involved (real multi-agent activity)
const isMultiAgent = computed(() => {
  return involvedAgents.value.length > 1 ||
         communications.value.length > 0 ||
         (props.delegationChain && props.delegationChain.length > 1)
})

watch(() => props.agentStatus, (newStatus) => {
  if (newStatus) {
    currentAgent.value = newStatus
    if (newStatus.agent && !involvedAgents.value.includes(newStatus.agent)) {
      involvedAgents.value.push(newStatus.agent)
    }
  }
}, { deep: true })

watch(() => props.teamCommunications, (newComms) => {
  if (newComms && newComms.length > 0) {
    communications.value = [...newComms]
  }
}, { deep: true })

watch(() => props.involvedAgentsList, (newList) => {
  if (newList && newList.length > 0) {
    involvedAgents.value = [...new Set([...involvedAgents.value, ...newList])]
  }
}, { deep: true })

watch(() => props.specialistReport, (newReport) => {
  if (newReport) {
    lastReport.value = newReport
    if (newReport.specialist_name && !involvedAgents.value.includes(newReport.specialist_name)) {
      involvedAgents.value.push(newReport.specialist_name)
    }
    if (newReport.request_help) {
      communications.value.push({
        from: newReport.specialist_name,
        to: newReport.request_help,
        message: newReport.suggestions?.[0] || 'assistance needed'
      })
    }
  }
}, { deep: true })

const confidenceColor = computed(() => {
  if (!lastReport.value?.confidence) return null
  const conf = lastReport.value.confidence
  if (conf >= 0.8) return 'green'
  if (conf >= 0.5) return 'yellow'
  return 'red'
})

const confidenceClass = computed(() => {
  const colors = { green: 'bg-green-500', yellow: 'bg-yellow-500', red: 'bg-red-500' }
  return colors[confidenceColor.value] || 'bg-gray-500'
})

const confidenceTextClass = computed(() => {
  const colors = { green: 'text-green-400', yellow: 'text-yellow-400', red: 'text-red-400' }
  return colors[confidenceColor.value] || 'text-gray-400'
})

const reportStatusClass = computed(() => {
  if (!lastReport.value?.status) return ''
  const classes = {
    complete: 'bg-emerald-500/20 text-emerald-400',
    partial: 'bg-amber-500/20 text-amber-400',
    needs_help: 'bg-red-500/20 text-red-400'
  }
  return classes[lastReport.value.status] || 'bg-gray-500/20 text-gray-400'
})

const isWorking = computed(() => {
  return currentAgent.value?.status === 'working' ||
         currentAgent.value?.status === 'analyzing' ||
         currentAgent.value?.status === 'delegating'
})

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

function reset() {
  currentAgent.value = null
  communications.value = []
  involvedAgents.value = []
  lastReport.value = null
}

defineExpose({ reset })
</script>

<style scoped>
.expand-enter-active,
.expand-leave-active {
  transition: all 0.2s ease;
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
  max-height: 300px;
}
</style>
