<template>
  <div class="bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl overflow-hidden transition-all duration-300">
    <button
      class="w-full flex items-center justify-between px-3 md:px-4 py-2 md:py-2.5 text-[10px] md:text-xs hover:bg-purple-500/10 transition-colors"
      @click="activity.expanded = !activity.expanded"
    >
      <div class="flex items-center gap-2 md:gap-3 min-w-0">
        <!-- Status dot -->
        <div
          class="w-2 h-2 rounded-full flex-shrink-0"
          :class="statusDotClass"
        />
        <!-- Agent info -->
        <div class="flex items-center gap-1.5 min-w-0">
          <span class="text-makoclaw-text-secondary">Agent:</span>
          <span
            class="font-bold capitalize truncate"
            :class="agentTextColor"
          >
            {{ activity.agent }}
          </span>
          <span
            v-if="activity.activeSkills && activity.activeSkills.length > 0"
            class="text-[9px] hidden sm:inline text-purple-400/80"
          >
            using {{ activity.activeSkills.join(', ') }}
          </span>
          <span
            v-else-if="activity.status === 'working'"
            class="text-[9px] italic hidden sm:inline text-emerald-400/80"
          >
            working...
          </span>
          <span
            v-else-if="activity.status === 'complete'"
            class="text-[9px] text-emerald-400/80 hidden sm:inline"
          >
            done
          </span>
          <span
            v-else-if="activity.status === 'delegating'"
            class="text-[9px] text-purple-400/80 italic hidden sm:inline"
          >
            delegating...
          </span>
          <span
            v-else-if="activity.status === 'fallback'"
            class="text-[9px] text-amber-400/80 italic hidden sm:inline"
          >
            fallback
          </span>
        </div>
      </div>
      <!-- Chevron -->
      <svg
        class="w-3.5 md:w-4 h-3.5 md:h-4 text-makoclaw-text-secondary transition-transform duration-300 flex-shrink-0"
        :class="{ 'rotate-180': activity.expanded }"
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

    <Transition name="expand">
      <div
        v-if="activity.expanded"
        class="px-3 md:px-4 pb-3 md:pb-4 border-t border-makoclaw-border/30 animate-fadeIn bg-makoclaw-bg/20"
      >
        <div class="mt-3 space-y-2.5">
          <!-- Reason / context -->
          <div v-if="activity.reason">
            <div class="detail-label">
              <svg
                class="w-3 h-3"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              Reason
            </div>
            <p class="detail-value italic">
              {{ activity.reason }}
            </p>
          </div>

          <!-- Delegation chain -->
          <div v-if="activity.delegationChain && activity.delegationChain.length > 1">
            <div class="detail-label">
              <svg
                class="w-3 h-3"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
                />
              </svg>
              Delegation Chain
            </div>
            <div class="flex items-center gap-1 flex-wrap mt-1">
              <template
                v-for="(ag, idx) in activity.delegationChain"
                :key="idx"
              >
                <span
                  class="text-[10px] px-1.5 py-0.5 rounded capitalize font-medium"
                  :class="idx === activity.delegationChain.length - 1
                    ? 'bg-emerald-500/20 text-emerald-400 ring-1 ring-emerald-500/30'
                    : 'bg-makoclaw-surface/50 text-makoclaw-text-secondary'"
                >
                  {{ ag }}
                </span>
                <span
                  v-if="idx < activity.delegationChain.length - 1"
                  class="text-makoclaw-text-secondary/30 text-xs"
                >
                  →
                </span>
              </template>
            </div>
          </div>

          <!-- Active Skills -->
          <div v-if="activity.activeSkills && activity.activeSkills.length > 0">
            <div class="detail-label">
              <svg
                class="w-3 h-3"
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
              Skills
            </div>
            <div class="flex flex-wrap gap-1 mt-1">
              <span
                v-for="skill in activity.activeSkills"
                :key="skill"
                class="text-[10px] px-1.5 py-0.5 rounded bg-purple-500/20 text-purple-300 ring-1 ring-purple-500/30 font-medium"
              >
                {{ skill }}
              </span>
            </div>
          </div>

          <!-- Tools used -->
          <div v-if="hasToolDetails">
            <div class="detail-label">
              <svg
                class="w-3 h-3"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                />
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                />
              </svg>
              Tools Used
            </div>
            <div class="flex flex-wrap gap-1 mt-1">
              <span
                v-for="tool in activity.toolsUsed"
                :key="tool"
                class="text-[10px] px-1.5 py-0.5 rounded bg-makoclaw-surface/60 text-makoclaw-text-secondary font-mono"
              >
                {{ tool }}
              </span>
            </div>

            <div
              v-if="agentToolCalls.length > 0"
              class="mt-2 space-y-1.5"
            >
              <div
                v-for="toolCall in agentToolCalls"
                :key="toolCall.id"
                class="rounded-lg border border-makoclaw-border/30 bg-makoclaw-bg/30 px-2.5 py-2"
              >
                <div class="flex items-center justify-between gap-2">
                  <div class="flex items-center gap-2 min-w-0">
                    <div
                      class="w-2 h-2 rounded-full flex-shrink-0"
                      :class="toolStatusDotClass(toolCall.status)"
                    />
                    <span class="text-[10px] font-mono text-makoclaw-text truncate">{{ toolCall.name }}</span>
                  </div>
                  <span
                    class="text-[9px] px-1.5 py-0.5 rounded-full flex-shrink-0"
                    :class="toolStatusBadgeClass(toolCall.status)"
                  >
                    {{ toolStatusLabel(toolCall.status) }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Confidence -->
          <div
            v-if="activity.confidence"
            class="flex items-center gap-2"
          >
            <div class="detail-label mb-0">
              <svg
                class="w-3 h-3"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              Confidence
            </div>
            <div class="flex items-center gap-1.5">
              <div class="w-16 h-1.5 bg-makoclaw-bg/50 rounded-full overflow-hidden">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :class="confidenceBarColor"
                  :style="{ width: Math.round((activity.confidence || 0) * 100) + '%' }"
                />
              </div>
              <span class="text-[10px] font-mono text-makoclaw-text-secondary">
                {{ Math.round((activity.confidence || 0) * 100) }}%
              </span>
            </div>
          </div>

          <!-- Timestamp -->
          <div
            v-if="activity.timestamp"
            class="text-[9px] text-makoclaw-text-secondary/50 mt-1"
          >
            {{ formattedTime }}
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  activity: {
    type: Object,
    required: true
  },
  msg: {
    type: Object,
    default: null
  }
})

const agentToolCalls = computed(() => {
  if (!props.msg?.toolCalls?.length) return []
  return props.msg.toolCalls.filter(tc => tc.agentName === props.activity.agent)
})

const hasToolDetails = computed(() => {
  return (props.activity.toolsUsed && props.activity.toolsUsed.length > 0) || agentToolCalls.value.length > 0
})

const statusDotClass = computed(() => {
  const map = {
    working: 'bg-emerald-400 animate-pulse',
    complete: 'bg-emerald-400',
    delegating: 'bg-purple-400 animate-pulse',
    fallback: 'bg-amber-400',
    requesting_help: 'bg-amber-400 animate-pulse',
    colleague_complete: 'bg-teal-400',
    error: 'bg-red-400',
    timeout: 'bg-red-400'
  }
  return map[props.activity.status] || 'bg-makoclaw-text-secondary'
})

const agentTextColor = computed(() => {
  const map = {
    working: 'text-emerald-400',
    complete: 'text-emerald-400',
    delegating: 'text-purple-400',
    fallback: 'text-amber-400',
    requesting_help: 'text-amber-400',
    colleague_complete: 'text-teal-400',
    error: 'text-red-400',
    timeout: 'text-red-400'
  }
  return map[props.activity.status] || 'text-makoclaw-text-secondary'
})

const confidenceBarColor = computed(() => {
  const c = props.activity.confidence || 0
  if (c >= 0.8) return 'bg-emerald-400'
  if (c >= 0.5) return 'bg-amber-400'
  return 'bg-red-400'
})

const formattedTime = computed(() => {
  if (!props.activity.timestamp) return ''
  const date = new Date(props.activity.timestamp)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
})

function toolStatusLabel(status) {
  if (status === 'started') return 'executing…'
  if (status === 'error') return 'error'
  return 'done'
}

function toolStatusBadgeClass(status) {
  if (status === 'started') return 'bg-makoclaw-warning/15 text-makoclaw-warning ring-1 ring-makoclaw-warning/30'
  if (status === 'error') return 'bg-makoclaw-error/15 text-makoclaw-error ring-1 ring-makoclaw-error/30'
  return 'bg-makoclaw-success/15 text-makoclaw-success ring-1 ring-makoclaw-success/30'
}

function toolStatusDotClass(status) {
  if (status === 'started') return 'bg-makoclaw-warning animate-pulse'
  if (status === 'error') return 'bg-makoclaw-error'
  return 'bg-makoclaw-success'
}
</script>

<style scoped>
.detail-label {
  font-size: 9px;
  color: rgb(var(--pc-text-secondary));
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-weight: 700;
  margin-bottom: 0.25rem;
  display: flex;
  align-items: center;
  gap: 0.375rem;
}
@media (min-width: 768px) {
  .detail-label {
    font-size: 10px;
  }
}
.detail-value {
  background: rgb(0 0 0 / 0.2);
  padding: 0.25rem 0.5rem;
  border-radius: 0.5rem;
  color: rgb(var(--pc-text) / 0.8);
  font-size: 10px;
  line-height: normal;
  font-family: ui-monospace, SFMono-Regular, monospace;
  border: 1px solid rgb(var(--pc-border) / 0.1);
}
@media (min-width: 768px) {
  .detail-value {
    font-size: 11px;
  }
}
</style>
