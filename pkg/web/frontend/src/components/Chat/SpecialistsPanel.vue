<template>
  <div class="glass-panel rounded-lg p-4 mb-4">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-sm font-semibold text-makoclaw-text flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
        </svg>
        Available Specialists
      </h3>

      <button
        @click="collapsed = !collapsed"
        class="p-1 rounded hover:bg-makoclaw-bg/50 transition-colors"
      >
        <svg
          class="w-4 h-4 transition-transform"
          :class="{ 'rotate-180': !collapsed }"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>
    </div>

    <transition name="expand">
      <div v-if="!collapsed" class="space-y-2">
        <div
          v-for="specialist in specialists"
          :key="specialist.name"
          class="flex items-start gap-3 p-2 sm:p-2.5 rounded-xl bg-makoclaw-surface/20 hover:bg-makoclaw-surface/40 border border-makoclaw-border/10 hover:border-makoclaw-border/20 backdrop-blur-sm transition-all duration-200"
        >
          <SpecialistBadge :name="specialist.name" class="text-xs flex-shrink-0 mt-0.5" />

          <div class="flex-1 min-w-0">
            <p class="text-xs text-makoclaw-text-secondary">
              {{ specialist.description }}
            </p>

            <div v-if="specialist.tools && specialist.tools.length > 0" class="flex flex-wrap gap-1 mt-1">
              <span
                v-for="tool in specialist.tools.slice(0, 3)"
                :key="tool"
                class="text-[10px] px-1.5 py-0.5 rounded bg-makoclaw-bg/50 text-makoclaw-text-secondary"
              >
                {{ tool }}
              </span>
              <span
                v-if="specialist.tools.length > 3"
                class="text-[10px] px-1.5 py-0.5 text-makoclaw-text-secondary"
              >
                +{{ specialist.tools.length - 3 }} more
              </span>
            </div>
          </div>
        </div>

        <p v-if="specialists.length === 0" class="text-xs text-makoclaw-text-secondary text-center py-2">
          No specialists configured. Configure in Settings > Agents.
        </p>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import SpecialistBadge from './SpecialistBadge.vue'
import advancedService from '../../services/advancedService'

const specialists = ref([])
const collapsed = ref(true)

onMounted(async () => {
  try {
    const response = await advancedService.fetchSpecialists()
    specialists.value = response.specialists || []
  } catch (error) {
    console.error('Failed to fetch specialists:', error)
  }
})
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
  max-height: 500px;
}
</style>
