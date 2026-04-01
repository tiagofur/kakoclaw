<template>
  <div class="flex items-center justify-between gap-3 px-4 py-3 border-b border-makoclaw-border/10 bg-makoclaw-surface/20 backdrop-blur-sm">
    <div class="flex items-center gap-2 min-w-0">
      <div class="w-8 h-8 rounded-xl bg-gradient-to-br from-orange-500/20 to-amber-500/20 flex items-center justify-center ring-1 ring-white/10">
        <i class="fas fa-terminal text-orange-400 text-xs" />
      </div>
      <div class="min-w-0">
        <div class="text-xs font-bold uppercase tracking-[0.2em] text-makoclaw-text-secondary">Terminal</div>
        <div class="text-[10px] text-makoclaw-text-secondary/70 truncate">Live Dev Studio session usage</div>
      </div>
    </div>

    <div
      data-testid="token-badge"
      class="inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-[11px] font-bold border transition-all"
      :class="badgeClass"
    >
      <i class="fas fa-coins text-[10px]" />
      <span>{{ sessionTokens }} / {{ tokenLimit }} tokens</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useDevStudioStore } from '@/stores/devStudioStore'

const devStore = useDevStudioStore()

const sessionTokens = computed(() => devStore.sessionTokens)
const tokenLimit = computed(() => devStore.tokenLimit)

const badgeClass = computed(() => {
  if (tokenLimit.value > 0 && sessionTokens.value >= tokenLimit.value) {
    return 'token-badge-danger bg-makoclaw-error/10 border-makoclaw-error/30 text-makoclaw-error'
  }
  if (tokenLimit.value > 0 && sessionTokens.value >= tokenLimit.value * 0.9) {
    return 'token-badge-warning bg-makoclaw-warning/10 border-makoclaw-warning/30 text-makoclaw-warning'
  }
  return 'token-badge-default bg-makoclaw-surface/60 border-makoclaw-border/50 text-makoclaw-text-secondary'
})
</script>
