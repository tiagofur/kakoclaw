<template>
  <div class="glass-panel rounded-2xl overflow-hidden border border-makoclaw-border/40 border-l-4 border-l-blue-500/50 shadow-lg shadow-blue-500/5">
    <button
      class="w-full flex items-center justify-between gap-3 px-3 md:px-4 py-2.5 text-left hover:bg-makoclaw-surface/30 transition-all duration-200"
      type="button"
      @click="toggleExpanded"
    >
      <div class="flex items-center gap-2 min-w-0">
        <div class="flex items-center justify-center w-7 h-7 rounded-xl bg-gradient-to-br from-blue-500/15 to-purple-500/15 ring-1 ring-white/10 flex-shrink-0">
          <span class="text-sm">🧠</span>
        </div>
        <div class="min-w-0">
          <div class="flex items-center gap-2 flex-wrap">
            <span class="text-xs font-semibold uppercase tracking-wide text-makoclaw-text-secondary">Thinking</span>
            <span
              v-if="thinkingBlock.agentName"
              class="px-1.5 py-0.5 rounded-full text-[10px] font-medium capitalize bg-blue-500/10 text-blue-400 ring-1 ring-blue-500/20"
            >
              {{ thinkingBlock.agentName }}
            </span>
            <span
              v-if="showStreamingPulse"
              class="inline-flex items-center gap-1 text-[10px] text-blue-400/80"
            >
              <span class="w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse" />
              live
            </span>
          </div>
        </div>
      </div>

      <svg
        class="w-4 h-4 text-makoclaw-text-secondary transition-transform duration-300 flex-shrink-0"
        :class="{ 'rotate-180': isExpanded }"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <Transition name="expand">
      <div
        v-if="isExpanded"
        class="px-3 md:px-4 pb-3 md:pb-4 border-t border-makoclaw-border/30 bg-makoclaw-bg/15"
      >
        <div class="mt-3 rounded-xl bg-makoclaw-bg/30 border border-makoclaw-border/20 px-3 py-2.5 shadow-inner">
          <pre class="whitespace-pre-wrap break-words text-[11px] md:text-xs leading-relaxed italic text-makoclaw-text-secondary font-mono">{{ thinkingBlock.content }}</pre>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  thinkingBlock: {
    type: Object,
    required: true
  },
  isStreaming: {
    type: Boolean,
    default: false
  }
})

const isExpanded = computed(() => {
  if (props.thinkingBlock.manuallyToggled) {
    return !!props.thinkingBlock.expanded
  }

  return !!props.thinkingBlock.expanded || props.isStreaming
})

const showStreamingPulse = computed(() => props.isStreaming && !props.thinkingBlock.finalized)

function toggleExpanded() {
  props.thinkingBlock.manuallyToggled = true
  props.thinkingBlock.expanded = !isExpanded.value
}
</script>

<style scoped>
.expand-enter-active,
.expand-leave-active {
  transition: all 0.25s ease;
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
