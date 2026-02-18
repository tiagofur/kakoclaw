<template>
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50 overflow-y-auto">
    <div class="bg-picoclaw-surface border border-picoclaw-border rounded-lg max-w-2xl w-full shadow-lg my-4">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 border-b border-picoclaw-border">
        <div>
          <h3 class="text-lg font-semibold">{{ task.title }}</h3>
          <p class="text-xs text-picoclaw-text-secondary">ID: {{ task.id }}</p>
        </div>
        <button
          @click="$emit('close')"
          class="p-1 hover:bg-picoclaw-border rounded transition-smooth"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Content -->
      <div class="p-4 space-y-4 max-h-96 overflow-y-auto">
        <!-- Title -->
        <div>
          <label class="block text-sm font-medium mb-2">Title</label>
          <input
            v-model="editForm.title"
            type="text"
            class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded focus-ring text-sm"
            :disabled="isLoading"
          />
        </div>

        <!-- Description -->
        <div>
          <label class="block text-sm font-medium mb-2">Description</label>
          <textarea
            v-model="editForm.description"
            class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded focus-ring text-sm resize-none"
            rows="3"
            :disabled="isLoading"
          ></textarea>
        </div>

        <!-- Status -->
        <div>
          <label class="block text-sm font-medium mb-2">Status</label>
          <select
            v-model="editForm.status"
            class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded focus-ring text-sm"
            :disabled="isLoading"
          >
            <option value="backlog">Backlog</option>
            <option value="todo">To Do</option>
            <option value="in_progress">In Progress</option>
            <option value="review">Review</option>
            <option value="done">Done</option>
          </select>
        </div>

        <!-- Result -->
        <div>
          <label class="block text-sm font-medium mb-2">Result</label>
          <textarea
            v-model="editForm.result"
            class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded focus-ring text-sm resize-none"
            rows="2"
            placeholder="Task result/output"
            :disabled="isLoading"
          ></textarea>
        </div>

        <!-- Logs -->
        <div v-if="logs.length > 0">
          <label class="block text-sm font-medium mb-2">Execution Logs</label>
          <div class="bg-picoclaw-bg border border-picoclaw-border rounded p-3 text-xs max-h-48 overflow-y-auto">
            <table class="w-full text-left">
              <tbody class="divide-y divide-picoclaw-border">
                <tr v-for="log in logs" :key="log.id" class="hover:bg-picoclaw-surface transition-smooth">
                  <td class="py-2 px-2 text-picoclaw-text-secondary whitespace-nowrap">
                    {{ formatTime(log.created_at) }}
                  </td>
                  <td class="py-2 px-2">{{ log.action }}</td>
                  <td class="py-2 px-2 text-picoclaw-text-secondary truncate">{{ log.details }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Error Message -->
        <div v-if="errorMessage" class="p-3 bg-picoclaw-error/20 border border-picoclaw-error rounded text-picoclaw-error text-sm">
          {{ errorMessage }}
        </div>
      </div>

      <!-- Actions -->
      <div class="border-t border-picoclaw-border p-4 flex gap-2 justify-end">
        <button
          @click="handleDelete"
          class="px-4 py-2 bg-picoclaw-error hover:bg-picoclaw-error/80 text-white rounded transition-smooth text-sm font-medium disabled:opacity-50"
          :disabled="isLoading"
        >
          Delete
        </button>
        <button
          @click="$emit('close')"
          class="px-4 py-2 border border-picoclaw-border rounded hover:bg-picoclaw-border transition-smooth text-sm"
          :disabled="isLoading"
        >
          Cancel
        </button>
        <button
          @click="handleUpdate"
          class="px-4 py-2 bg-picoclaw-accent hover:bg-picoclaw-accent-hover text-white rounded transition-smooth text-sm font-medium disabled:opacity-50"
          :disabled="isLoading"
        >
          {{ isLoading ? 'Saving...' : 'Save Changes' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import taskService from '../../services/taskService'

const props = defineProps({
  task: { type: Object, required: true }
})

const emit = defineEmits(['close', 'updated', 'deleted'])

const editForm = ref({
  title: props.task.title,
  description: props.task.description || '',
  status: props.task.status,
  result: props.task.result || ''
})

const logs = ref([])
const isLoading = ref(false)
const errorMessage = ref('')

onMounted(async () => {
  try {
    logs.value = await taskService.getTaskLogs(props.task.id)
  } catch (error) {
    console.error('Failed to load task logs:', error)
  }
})

const handleUpdate = async () => {
  errorMessage.value = ''
  isLoading.value = true

  try {
    await emit('updated', props.task.id, {
      title: editForm.value.title,
      description: editForm.value.description,
      status: editForm.value.status,
      result: editForm.value.result
    })
  } catch (error) {
    errorMessage.value = 'Failed to update task'
  } finally {
    isLoading.value = false
  }
}

const handleDelete = async () => {
  if (!confirm('Are you sure you want to delete this task?')) return

  errorMessage.value = ''
  isLoading.value = true

  try {
    await emit('deleted', props.task.id)
  } catch (error) {
    errorMessage.value = 'Failed to delete task'
    isLoading.value = false
  }
}

const formatTime = (timestamp) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
</script>
