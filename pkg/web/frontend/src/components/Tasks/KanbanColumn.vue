<template>
  <div class="glass-panel rounded-2xl p-2.5 sm:p-3 md:p-4 flex flex-col h-full min-w-[256px] sm:min-w-[288px] md:min-w-[320px] shadow-sm transition-all duration-500 hover:shadow-[0_8px_30px_rgb(0,0,0,0.04)]">
    <!-- Column Header -->
    <div class="mb-4 pb-0 flex flex-col">
      <div class="pb-3 flex items-center justify-between">
        <h3 class="font-bold text-xs uppercase tracking-[0.2em] text-makoclaw-text-secondary flex items-center gap-2 opacity-80">
          <span
            class="w-2 h-2 rounded-full"
            :class="statusDotColor"
          />
          {{ title }}
        </h3>
        <span class="text-[10px] bg-makoclaw-bg/50 font-bold text-makoclaw-accent px-2.5 py-1 rounded-full border border-makoclaw-accent/10 shadow-sm backdrop-blur-md">
          {{ tasks.length }}
        </span>
      </div>
      <div class="h-[1px] bg-gradient-to-r from-transparent via-makoclaw-accent/10 to-transparent" />
    </div>

    <!-- Tasks List -->
    <div
      class="flex-1 space-y-2 sm:space-y-3 overflow-y-auto px-0.5 sm:px-1 -mx-0.5 sm:-mx-1 custom-scrollbar"
      @dragover.prevent
      @drop="handleDrop"
    >
      <div
        v-for="task in tasks"
        :key="task.id"
        draggable="true"
        class="bg-makoclaw-surface/30 border border-makoclaw-border/20 rounded-xl p-3 sm:p-3.5 md:p-4 cursor-grab active:cursor-grabbing hover:border-makoclaw-accent/30 hover:shadow-xl hover:-translate-y-[2px] transition-all duration-300 group relative overflow-hidden backdrop-blur-md ring-1 ring-white/[0.03] hover:ring-white/[0.08]"
        @dragstart="dragStart($event, task)"
        @click="$emit('task-click', task)"
      >
        <div class="absolute inset-x-0 top-0 h-1 bg-gradient-to-r from-makoclaw-accent to-blue-500 opacity-0 group-hover:opacity-100 transition-opacity" />
        <!-- Animated bottom-line on hover -->
        <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover:w-full transition-all duration-500 opacity-60" />
        <!-- Soft gradient glow (top-right corner) -->
        <div class="absolute -top-8 -right-8 w-20 h-20 bg-gradient-to-br from-makoclaw-accent to-blue-500 rounded-full opacity-0 blur-[20px] group-hover:opacity-15 transition-all duration-500" />

        <div class="flex items-start justify-between gap-2 mb-1">
          <h4 class="font-semibold text-sm leading-tight text-makoclaw-text group-hover:text-makoclaw-accent transition-colors flex-1 min-w-0">
            {{ task.title }}
          </h4>
          <div class="flex items-center gap-1 flex-shrink-0">
            <SpecialistBadge
              v-if="task.agent"
              :name="task.agent"
            />
            <!-- Action buttons - hover visibility, touch-friendly -->
            <div class="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity touch-visible" @click.stop>
              <button
                v-if="!task.archived"
                class="p-1 hover:bg-yellow-500/20 rounded text-makoclaw-text-secondary hover:text-yellow-500 transition-colors"
                title="Archive"
                @click="handleArchive(task)"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                </svg>
              </button>
              <button
                class="p-1 hover:bg-makoclaw-error/20 rounded text-makoclaw-text-secondary hover:text-makoclaw-error transition-colors"
                title="Delete"
                @click="handleDelete(task)"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <p
          v-if="task.description"
          class="text-xs text-makoclaw-text-secondary line-clamp-2 mt-1 mb-2"
        >
          {{ task.description }}
        </p>

        <div class="flex items-center gap-2 mt-2 pt-2 border-t border-makoclaw-border/20">
          <span :class="['text-[10px] px-1.5 py-0.5 rounded font-medium uppercase tracking-wide', getStatusColor(task.status)]">
            {{ getStatusLabel(task.status) }}
          </span>
          <span class="text-[10px] text-makoclaw-text-secondary ml-auto">
            {{ formatDate(task.created_at) }}
          </span>
        </div>
      </div>

      <!-- Empty State -->
      <div
        v-if="tasks.length === 0"
        class="flex flex-col items-center justify-center py-6 sm:py-8 text-makoclaw-text-secondary/50 border-2 border-dashed border-makoclaw-border/20 rounded-xl"
      >
        <svg
          class="w-8 h-8 mb-2"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        ><path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"
        /></svg>
        <p class="text-xs">
          No tasks
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import SpecialistBadge from '../Chat/SpecialistBadge.vue'

const props = defineProps({
  status: {
    type: String,
    default: ''
  },
  title: {
    type: String,
    default: ''
  },
  tasks: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['task-click', 'task-drop', 'task-archive', 'task-delete'])

const dragStart = (event, task) => {
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.setData('taskId', task.id)
  event.dataTransfer.setData('sourceStatus', task.status)
}

const handleDrop = (event) => {
  const taskId = event.dataTransfer?.getData('taskId')
  const sourceStatus = event.dataTransfer?.getData('sourceStatus')
  if (!taskId) {
    return
  }
  emit('task-drop', taskId, props.status, sourceStatus)
}

const handleArchive = (task) => {
  emit('task-archive', task)
}

const handleDelete = (task) => {
  emit('task-delete', task)
}

const getStatusLabel = (status) => {
  const labels = {
    'backlog': 'Backlog',
    'todo': 'To Do',
    'in_progress': 'In Progress',
    'review': 'Review',
    'done': 'Done'
  }
  return labels[status] || status
}

const getStatusColor = (status) => {
  const colors = {
    'backlog': 'bg-makoclaw-border text-makoclaw-text-secondary',
    'todo': 'bg-makoclaw-warning/20 text-makoclaw-warning',
    'in_progress': 'bg-makoclaw-accent/20 text-makoclaw-accent',
    'review': 'bg-makoclaw-accent/20 text-makoclaw-accent',
    'done': 'bg-makoclaw-success/20 text-makoclaw-success'
  }
  return colors[status] || colors['backlog']
}

const statusDotColor = computed(() => {
  const colors = {
    'backlog': 'bg-makoclaw-text-secondary/40',
    'todo': 'bg-makoclaw-warning',
    'in_progress': 'bg-makoclaw-accent animate-pulse',
    'review': 'bg-amber-500 animate-pulse',
    'done': 'bg-makoclaw-success'
  }
  return colors[props.status] || 'bg-makoclaw-text-secondary/40'
})

const formatDate = (date) => {
  const d = new Date(date)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
</script>

<style scoped>
/* Touch-friendly action buttons - visible on touch devices */
@media (hover: none) {
  .touch-visible {
    opacity: 0.6 !important;
  }
}
</style>

