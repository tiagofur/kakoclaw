<template>
  <div class="glass-panel rounded-xl p-3 flex flex-col h-full min-w-[320px]">
    <!-- Column Header -->
    <div class="mb-3 pb-3 border-b border-kakoclaw-border/50 flex items-center justify-between">
      <h3 class="font-bold text-sm uppercase tracking-wider text-kakoclaw-text-secondary flex items-center gap-2">
        {{ title }}
      </h3>
      <span class="text-xs bg-kakoclaw-surface font-mono text-kakoclaw-text px-2 py-0.5 rounded-full border border-kakoclaw-border">
          {{ tasks.length }}
      </span>
    </div>

    <!-- Tasks List -->
    <div
      class="flex-1 space-y-3 overflow-y-auto px-1 -mx-1"
      @dragover.prevent
      @drop="handleDrop"
    >
      <div
        v-for="task in tasks"
        :key="task.id"
        draggable="true"
        @dragstart="dragStart($event, task)"
        @click="$emit('task-click', task)"
        class="bg-kakoclaw-surface border border-kakoclaw-border rounded-lg p-3 cursor-move hover:border-kakoclaw-accent/50 hover:shadow-lg hover:-translate-y-0.5 transition-all duration-200 group relative overflow-hidden"
      >
        <div class="absolute inset-0 bg-gradient-to-r from-transparent via-transparent to-kakoclaw-accent/5 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none"></div>
        
        <div class="flex items-start justify-between gap-2 mb-1">
           <h4 class="font-semibold text-sm leading-tight text-kakoclaw-text group-hover:text-kakoclaw-accent transition-colors">
            {{ task.title }}
           </h4>
        </div>
        
        <p v-if="task.description" class="text-xs text-kakoclaw-text-secondary line-clamp-2 mt-1 mb-2">
          {{ task.description }}
        </p>

        <div class="flex items-center gap-2 mt-2 pt-2 border-t border-kakoclaw-border/50">
          <span :class="['text-[10px] px-1.5 py-0.5 rounded font-medium uppercase tracking-wide', getStatusColor(task.status)]">
            {{ getStatusLabel(task.status) }}
          </span>
          <span class="text-[10px] text-kakoclaw-text-secondary ml-auto">
            {{ formatDate(task.created_at) }}
          </span>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="tasks.length === 0" class="flex flex-col items-center justify-center py-8 text-kakoclaw-text-secondary/50 border-2 border-dashed border-kakoclaw-border/30 rounded-lg">
        <svg class="w-8 h-8 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" /></svg>
        <p class="text-xs">No tasks</p>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  status: String,
  title: String,
  tasks: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['task-click', 'task-drop'])

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
    'backlog': 'bg-kakoclaw-border text-kakoclaw-text-secondary',
    'todo': 'bg-kakoclaw-warning/20 text-kakoclaw-warning',
    'in_progress': 'bg-kakoclaw-accent/20 text-kakoclaw-accent',
    'review': 'bg-kakoclaw-accent/20 text-kakoclaw-accent',
    'done': 'bg-kakoclaw-success/20 text-kakoclaw-success'
  }
  return colors[status] || colors['backlog']
}

const formatDate = (date) => {
  const d = new Date(date)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
</script>
