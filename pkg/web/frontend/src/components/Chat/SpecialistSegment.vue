<template>
  <div class="glass-panel rounded-xl border border-makoclaw-border/30 overflow-hidden shadow-lg shadow-black/5">
    <button
      type="button"
      data-testid="specialist-header"
      class="w-full px-3.5 py-3 text-left transition-all duration-200 hover:bg-makoclaw-surface/20"
      @click="toggleExpanded"
    >
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0 flex-1 space-y-2">
          <div class="flex items-center gap-2 flex-wrap">
            <div class="flex items-center gap-2 min-w-0">
              <div
                class="w-8 h-8 rounded-lg flex items-center justify-center ring-1 ring-white/10 flex-shrink-0"
                :class="agentIconClass"
              >
                <svg
                  class="w-4 h-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M9.5 3A2.5 2.5 0 007 5.5V7H5.5A2.5 2.5 0 003 9.5v5A2.5 2.5 0 005.5 17H7v1.5A2.5 2.5 0 009.5 21h5a2.5 2.5 0 002.5-2.5V17h1.5a2.5 2.5 0 002.5-2.5v-5A2.5 2.5 0 0018.5 7H17V5.5A2.5 2.5 0 0014.5 3h-5zM9 7h6V5.5a.5.5 0 00-.5-.5h-5a.5.5 0 00-.5.5V7zm3 3a1.25 1.25 0 110 2.5A1.25 1.25 0 0112 10zm-3.5 6h7"
                  />
                </svg>
              </div>
              <div class="min-w-0">
                <p class="text-sm font-semibold text-makoclaw-text capitalize truncate">
                  {{ agentName }}
                </p>
                <p
                  v-if="segmentTimestamp"
                  class="text-[11px] text-makoclaw-text-secondary"
                >
                  {{ segmentTimestamp }}
                </p>
              </div>
            </div>

            <span
              data-testid="specialist-status-badge"
              class="px-2 py-0.5 text-[10px] font-semibold rounded-full ring-1 flex-shrink-0"
              :class="statusBadgeClass"
            >
              {{ statusLabel }}
            </span>

            <span
              v-if="hasConfidence"
              class="text-[10px] px-2 py-0.5 rounded-full bg-makoclaw-accent/10 text-makoclaw-accent ring-1 ring-makoclaw-accent/20"
            >
              {{ confidenceLabel }}
            </span>
          </div>

          <p
            data-testid="specialist-preview"
            class="text-xs md:text-sm text-makoclaw-text-secondary whitespace-pre-line"
          >
            {{ contentPreview || (isStreaming ? 'Waiting for specialist response…' : 'No specialist content yet.') }}
          </p>
        </div>

        <svg
          class="w-4 h-4 mt-1 text-makoclaw-text-secondary transition-transform duration-300 flex-shrink-0"
          :class="{ 'rotate-180': isExpanded }"
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
      </div>
    </button>

    <Transition name="expand">
      <div
        v-if="isExpanded"
        data-testid="specialist-content"
        class="border-t border-makoclaw-border/20 bg-makoclaw-bg/15 px-3.5 py-3"
      >
        <div class="rounded-lg bg-black/10 border border-white/5 p-3 md:p-4">
          <MarkdownRenderer
            :content="fullContent"
            class="text-sm md:text-base"
          />
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { computed, ref, toRef } from 'vue'
import MarkdownRenderer from './MarkdownRenderer.vue'

const props = defineProps({
  segment: {
    type: Object,
    required: true
  },
  isStreaming: {
    type: Boolean,
    default: false
  }
})

const segment = toRef(props, 'segment')
const localExpanded = ref(false)

const isExpanded = computed(() => localExpanded.value)

const fullContent = computed(() => segment.value?.content ?? '')

const contentPreview = computed(() => fullContent.value.split('\n').slice(0, 2).join('\n'))

const normalizedStatus = computed(() => segment.value?.status || 'working')

const agentName = computed(() => segment.value?.agent || segment.value?.specialist || 'specialist')

const statusLabel = computed(() => normalizedStatus.value)

const statusBadgeClass = computed(() => {
  if (normalizedStatus.value === 'error') return 'bg-makoclaw-error/15 text-makoclaw-error ring-makoclaw-error/25'
  if (normalizedStatus.value === 'complete') return 'bg-makoclaw-success/15 text-makoclaw-success ring-makoclaw-success/25'
  return 'bg-makoclaw-warning/15 text-makoclaw-warning ring-makoclaw-warning/25'
})

const agentIconClass = computed(() => {
  if (normalizedStatus.value === 'error') return 'bg-makoclaw-error/10 text-makoclaw-error'
  if (normalizedStatus.value === 'complete') return 'bg-makoclaw-success/10 text-makoclaw-success'
  return 'bg-makoclaw-warning/10 text-makoclaw-warning'
})

const hasConfidence = computed(() => typeof segment.value?.confidence === 'number')

const confidenceLabel = computed(() => `${Math.round((segment.value?.confidence ?? 0) * 100)}% confidence`)

const segmentTimestamp = computed(() => {
  if (!segment.value?.timestamp) return ''

  return new Date(segment.value.timestamp).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit'
  })
})

function toggleExpanded() {
  localExpanded.value = !localExpanded.value
}
</script>
