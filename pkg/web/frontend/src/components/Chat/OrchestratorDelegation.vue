<template>
  <div class="bg-makoclaw-bg/40 border border-makoclaw-border/40 rounded-xl overflow-hidden">
    <button
      type="button"
      class="w-full flex items-center justify-between px-3 py-2 text-xs hover:bg-makoclaw-surface/20 transition-colors"
      :aria-expanded="isExpanded"
      @click="isExpanded = !isExpanded"
    >
      <div class="flex items-center gap-2 min-w-0">
        <!-- Brain icon -->
        <div class="w-5 h-5 rounded-md bg-violet-500/10 flex items-center justify-center flex-shrink-0">
          <svg class="w-3 h-3 text-violet-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
        </div>
        <span class="text-violet-400 font-medium">Delegando a</span>
        <span class="font-semibold text-makoclaw-text capitalize truncate">{{ specialistName }}</span>
      </div>
      <div class="flex items-center gap-1.5 flex-shrink-0 ml-2">
        <span class="text-[10px] text-makoclaw-text-secondary">ver instrucciones</span>
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
        class="border-t border-makoclaw-border/20 px-3 py-2 bg-makoclaw-bg/20"
      >
        <p class="text-[11px] text-makoclaw-text-secondary whitespace-pre-wrap break-words leading-relaxed font-mono">{{ task }}</p>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref } from 'vue'

defineProps({
  specialistName: {
    type: String,
    required: true
  },
  task: {
    type: String,
    required: true
  },
  delegationId: {
    type: String,
    default: ''
  }
})

const isExpanded = ref(false)
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
</style>
