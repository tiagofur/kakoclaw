<template>
  <div class="bg-makoclaw-bg/40 border border-makoclaw-border/40 rounded-xl overflow-hidden">
    <button
      type="button"
      class="w-full flex items-center justify-between px-3 py-2 text-xs hover:bg-makoclaw-surface/20 transition-colors"
      :aria-expanded="isExpanded"
      @click="toggleExpanded"
    >
      <div class="flex items-center gap-2 min-w-0">
        <!-- Work status icon -->
        <div
          class="w-5 h-5 rounded-md flex items-center justify-center flex-shrink-0 transition-colors"
          :class="isDone ? 'bg-emerald-500/10' : 'bg-amber-500/10'"
        >
          <svg
            v-if="!isDone"
            class="w-3 h-3 text-amber-400 animate-pulse"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          <svg
            v-else
            class="w-3 h-3 text-emerald-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <span
          class="font-semibold capitalize truncate"
          :class="isDone ? 'text-emerald-400' : 'text-amber-400'"
        >{{ specialistName }}</span>
        <span class="text-makoclaw-text-secondary">
          {{ isDone ? 'completado' : 'trabajando...' }}
        </span>
        <span v-if="isDone && durationText" class="text-[10px] text-makoclaw-text-secondary/60">
          en {{ durationText }}
        </span>
      </div>
      <div class="flex items-center gap-1.5 flex-shrink-0 ml-2">
        <span class="text-[10px] text-makoclaw-text-secondary">ver proceso</span>
        <svg
          class="w-3 h-3 text-makoclaw-text-secondary transition-transform duration-200"
          :class="{ 'rotate-180': isExpanded }"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </div>
    </button>

    <Transition name="expand">
      <div
        v-if="isExpanded"
        class="border-t border-makoclaw-border/20 bg-makoclaw-bg/20"
      >
        <div
          ref="tokenArea"
          class="px-3 py-2 max-h-48 overflow-y-auto custom-scrollbar"
        >
          <p class="text-[11px] text-makoclaw-text-secondary whitespace-pre-wrap break-words leading-relaxed font-mono">
            {{ tokens }}<span v-if="!isDone" class="streaming-cursor" />
          </p>
          <p v-if="!tokens && !isDone" class="text-[11px] text-makoclaw-text-secondary/50 italic">
            Esperando respuesta del especialista...
          </p>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, watch } from 'vue'

const props = defineProps({
  specialistName: {
    type: String,
    required: true
  },
  delegationId: {
    type: String,
    default: ''
  },
  tokens: {
    type: String,
    default: ''
  },
  isDone: {
    type: Boolean,
    default: false
  },
  durationMs: {
    type: Number,
    default: null
  }
})

const isExpanded = ref(false)
const tokenArea = ref(null)

const durationText = computed(() => {
  if (!props.durationMs) return null
  const secs = Math.round(props.durationMs / 1000)
  if (secs < 60) return `${secs}s`
  const mins = Math.floor(secs / 60)
  const rem = secs % 60
  return `${mins}m ${rem}s`
})

function toggleExpanded() {
  isExpanded.value = !isExpanded.value
}

// Auto-scroll token area when new tokens arrive and expanded
watch(() => props.tokens, async () => {
  if (!isExpanded.value || !tokenArea.value) return
  await nextTick()
  tokenArea.value.scrollTop = tokenArea.value.scrollHeight
})
</script>

<style scoped>
.expand-enter-active,
.expand-leave-active {
  transition: max-height 0.2s ease, opacity 0.2s ease;
  max-height: 300px;
  overflow: hidden;
}
.expand-enter-from,
.expand-leave-to {
  max-height: 0;
  opacity: 0;
}

.streaming-cursor::after {
  content: '\25AE';
  animation: blink 1s step-end infinite;
  margin-left: 1px;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}
</style>
