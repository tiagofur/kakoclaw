<template>
  <div
    v-if="isMultiAgent"
    class="glass-panel rounded-xl p-3 mb-3 relative overflow-hidden"
  >
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
          <p
            v-if="lastReport.request_help"
            class="text-[10px] text-red-400 mt-1"
          >
            Requesting help from {{ lastReport.request_help }}
          </p>
        </div>

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

        <div
          v-if="involvedAgents.length > 0"
          class="space-y-2"
        >
          <div
            v-for="agent in involvedAgents"
            :key="agent"
            class="rounded-lg border border-makoclaw-border/20 bg-makoclaw-surface/20 px-2.5 py-2"
          >
            <div class="flex items-center gap-2 flex-wrap">
              <span
                class="text-[10px] px-1.5 py-0.5 rounded capitalize font-medium"
                :class="agentBadgeClass(agent)"
              >
                {{ agent }}
              </span>
              <span
                v-if="hasActiveTools(agent)"
                class="text-[9px] text-makoclaw-text-secondary"
              >
                {{ activeToolCallsFor(agent).length }} active tool{{ activeToolCallsFor(agent).length > 1 ? 's' : '' }}
              </span>
            </div>

            <div
              v-if="hasActiveTools(agent)"
              class="mt-2 space-y-1.5"
            >
              <div
                v-for="toolCall in activeToolCallsFor(agent)"
                :key="toolCall.id"
                class="rounded-lg border border-makoclaw-border/25 bg-makoclaw-bg/30 px-2 py-1.5"
              >
                <div class="flex items-center justify-between gap-2">
                  <div class="flex items-center gap-2 min-w-0">
                    <div class="w-2 h-2 rounded-full bg-makoclaw-warning animate-pulse flex-shrink-0" />
                    <span class="text-[10px] font-mono text-makoclaw-text truncate">{{ toolCall.name }}</span>
                  </div>
                  <span class="text-[9px] px-1.5 py-0.5 rounded-full bg-makoclaw-warning/15 text-makoclaw-warning ring-1 ring-makoclaw-warning/30 flex-shrink-0">
                    executing…
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useChatStore } from '../../stores/chatStore'

const chatStore = useChatStore()

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

const collapsed = ref(true)

const communications = computed(() => props.teamCommunications || [])
const lastReport = computed(() => props.specialistReport)

function groupBy(items, key) {
  return items.reduce((groups, item) => {
    const groupKey = item?.[key] || 'main'
    if (!groups[groupKey]) groups[groupKey] = []
    groups[groupKey].push(item)
    return groups
  }, {})
}

const streamingMsg = computed(() => chatStore.streamingMessage)

const agentToolCalls = computed(() => groupBy(streamingMsg.value?.toolCalls || [], 'agentName'))

const activeToolCallsByAgent = computed(() => {
  return Object.entries(agentToolCalls.value).reduce((groups, [agent, toolCalls]) => {
    const activeTools = toolCalls.filter(toolCall => toolCall.status === 'started')
    if (activeTools.length > 0) {
      groups[agent] = activeTools
    }
    return groups
  }, {})
})

const involvedAgents = computed(() => {
  const activityAgents = (streamingMsg.value?.agentActivities || []).map(activity => activity.agent)
  return [...new Set([
    ...props.involvedAgentsList,
    ...props.delegationChain,
    props.agentStatus?.agent,
    props.agentStatus?.specialistName,
    lastReport.value?.specialist_name,
    ...activityAgents,
    ...Object.keys(agentToolCalls.value)
  ].filter(Boolean))]
})

const isMultiAgent = computed(() => {
  return involvedAgents.value.length > 1 ||
         communications.value.length > 0 ||
         (props.delegationChain && props.delegationChain.length > 1)
})

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
  return props.agentStatus?.status === 'working' ||
         props.agentStatus?.status === 'analyzing' ||
         props.agentStatus?.status === 'delegating'
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

function agentBadgeClass(agent) {
  return `${getAgentTextColor(agent)} bg-makoclaw-surface/50`
}

function activeToolCallsFor(agent) {
  return activeToolCallsByAgent.value[agent] || []
}

function hasActiveTools(agent) {
  return activeToolCallsFor(agent).length > 0
}
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
  max-height: 500px;
}
</style>
