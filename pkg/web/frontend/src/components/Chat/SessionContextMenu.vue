<template>
  <Teleport to="body">
    <div
      v-if="show"
      class="fixed inset-0 z-dropdown"
      @click="$emit('close')"
    >
      <div
        class="absolute bg-makoclaw-surface/95 backdrop-blur-xl border border-makoclaw-border/50 rounded-xl shadow-2xl py-1.5 min-w-[160px] animate-scaleIn ring-1 ring-white/10"
        :style="menuStyle"
        @click.stop
      >
        <button
          class="w-full text-left px-3 py-2 text-sm hover:bg-makoclaw-accent/10 hover:text-makoclaw-accent transition-all flex items-center gap-2.5"
          @click="$emit('rename', sessionId)"
        >
          <svg
            class="w-4 h-4 opacity-70"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
          /></svg>
          Rename
        </button>
        <button
          class="w-full text-left px-3 py-2 text-sm hover:bg-makoclaw-accent/10 hover:text-makoclaw-accent transition-all flex items-center gap-2.5"
          @click="$emit('archive', sessionId)"
        >
          <svg
            class="w-4 h-4 opacity-70"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"
          /></svg>
          {{ showArchived ? 'Unarchive' : 'Archive' }}
        </button>
        <div class="border-t border-makoclaw-border/50 my-1.5 mx-2" />
        <button
          class="w-full text-left px-3 py-2 text-sm hover:bg-makoclaw-error/10 text-makoclaw-error transition-all flex items-center gap-2.5"
          @click="$emit('delete', sessionId)"
        >
          <svg
            class="w-4 h-4 opacity-70"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
          /></svg>
          Delete
        </button>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  sessionId: { type: String, default: null },
  x: { type: Number, default: 0 },
  y: { type: Number, default: 0 },
  showArchived: { type: Boolean, default: false }
})

defineEmits(['close', 'rename', 'archive', 'delete'])

const menuStyle = computed(() => {
  const menuWidth = 180
  const menuHeight = 150
  const clampedX = Math.min(props.x, window.innerWidth - menuWidth - 8)
  const clampedY = Math.min(props.y, window.innerHeight - menuHeight - 8)
  return {
    left: Math.max(8, clampedX) + 'px',
    top: Math.max(8, clampedY) + 'px'
  }
})
</script>
