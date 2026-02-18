<template>
  <div class="bg-picoclaw-surface border border-picoclaw-border rounded-lg p-3 flex flex-col h-full">
    <!-- Column Header -->
    <div class="mb-3 pb-3 border-b border-picoclaw-border">
      <h3 class="font-semibold flex items-center gap-2">
        {{ title }}
        <span class="text-xs bg-picoclaw-border text-picoclaw-text-secondary px-2 py-1 rounded">
          {{ tasks.length }}
        </span>
      </h3>
    </div>

    <!-- Tasks List -->
    <div
      class="flex-1 space-y-2 overflow-y-auto"
      @dragover.prevent
      @drop="$emit('task-drop', $event)"
    >
      <div
        v-for="task in tasks"
        :key="task.id"
        draggable="true"
        @dragstart="dragStart($event, task)"
        @click="$emit('task-click', task)"
        class="bg-picoclaw-bg border border-picoclaw-border rounded p-2 cursor-move hover:border-picoclaw-accent transition-smooth group"
      >
        <h4 class="font-medium text-sm truncate group-hover:text-picoclaw-accent">
          {{ task.title }}
        </h4>
        <p v-if="task.description" class="text-xs text-picoclaw-text-secondary truncate mt-1">
          {{ task.description }}
        </p>
        <div class="flex items-center gap-2 mt-2 pt-2 border-t border-picoclaw-border">
          <span :class="['text-xs px-2 py-1 rounded', getStatusColor(task.status)]">
            {{ getStatusLabel(task.status) }}
          </span>
          <span class="text-xs text-picoclaw-text-secondary ml-auto">
            {{ formatDate(task.created_at) }}
          </span>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="tasks.length === 0" class="text-center py-8 text-picoclaw-text-secondary">
        <p class="text-sm">No tasks</p>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  status: String,
  title: String,
  tasks: {
    type: Array,
    default: () => []
  }
})

defineEmits(['task-click', 'task-drop'])

const dragStart = (event, task) => {
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.setData('taskId', task.id)
  event.dataTransfer.setData('sourceStatus', task.status)
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
    'backlog': 'bg-picoclaw-border text-picoclaw-text-secondary',
    'todo': 'bg-picoclaw-warning/20 text-picoclaw-warning',
    'in_progress': 'bg-picoclaw-accent/20 text-picoclaw-accent',
    'review': 'bg-picoclaw-accent/20 text-picoclaw-accent',
    'done': 'bg-picoclaw-success/20 text-picoclaw-success'
  }
  return colors[status] || colors['backlog']
}

const formatDate = (date) => {
  const d = new Date(date)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
</script>
