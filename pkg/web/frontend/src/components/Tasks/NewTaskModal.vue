<template>
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
    <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-lg max-w-md w-full shadow-lg">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 border-b border-kakoclaw-border">
        <h3 class="text-lg font-semibold">Create New Task</h3>
        <button
          @click="$emit('close')"
          class="p-1 hover:bg-kakoclaw-border rounded transition-smooth"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Form -->
      <form @submit.prevent="handleCreateTask" class="p-4 space-y-4">
        <!-- Title -->
        <div>
          <label for="title" class="block text-sm font-medium mb-2">
            Task Title
          </label>
          <input
            v-model="form.title"
            id="title"
            type="text"
            placeholder="Enter task title"
            class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded focus-ring text-sm"
            required
            :disabled="isLoading"
          />
        </div>

        <!-- Description -->
        <div>
          <label for="description" class="block text-sm font-medium mb-2">
            Description (optional)
          </label>
          <textarea
            v-model="form.description"
            id="description"
            placeholder="Enter task description"
            class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded focus-ring text-sm resize-none"
            rows="3"
            :disabled="isLoading"
          ></textarea>
        </div>

        <!-- Status -->
        <div>
          <label for="status" class="block text-sm font-medium mb-2">
            Initial Status
          </label>
          <select
            v-model="form.status"
            id="status"
            class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded focus-ring text-sm"
            :disabled="isLoading"
          >
            <option value="backlog">Backlog</option>
            <option value="todo">To Do</option>
            <option value="in_progress">In Progress</option>
            <option value="review">Review</option>
            <option value="done">Done</option>
          </select>
        </div>

        <!-- Error Message -->
        <div v-if="errorMessage" class="p-3 bg-kakoclaw-error/20 border border-kakoclaw-error rounded text-kakoclaw-error text-sm">
          {{ errorMessage }}
        </div>

        <!-- Actions -->
        <div class="flex gap-3 pt-4 border-t border-kakoclaw-border">
          <button
            type="button"
            @click="$emit('close')"
            class="flex-1 px-3 py-2 border border-kakoclaw-border rounded hover:bg-kakoclaw-border transition-smooth"
            :disabled="isLoading"
          >
            Cancel
          </button>
          <button
            type="submit"
            class="flex-1 px-3 py-2 bg-kakoclaw-accent hover:bg-kakoclaw-accent-hover text-white rounded transition-smooth disabled:opacity-50"
            :disabled="isLoading"
          >
            {{ isLoading ? 'Creating...' : 'Create Task' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const emit = defineEmits(['close', 'created'])

const form = ref({
  title: '',
  description: '',
  status: 'backlog'
})
const isLoading = ref(false)
const errorMessage = ref('')

const handleCreateTask = async () => {
  errorMessage.value = ''

  if (!form.value.title.trim()) {
    errorMessage.value = 'Task title is required'
    return
  }

  isLoading.value = true

  try {
    emit('created', {
      title: form.value.title,
      description: form.value.description,
      status: form.value.status
    })
  } catch (error) {
    console.error('Create task error:', error)
    errorMessage.value = 'Failed to create task'
  } finally {
    isLoading.value = false
  }
}
</script>
