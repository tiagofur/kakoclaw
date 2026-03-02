<template>
  <Transition name="modal">
    <div
      v-show="true"
      class="fixed inset-0 bg-black/40 backdrop-blur-md flex items-center justify-center p-3 sm:p-4 z-modal"
    >
      <div class="glass-panel rounded-2xl max-w-md w-full shadow-2xl relative overflow-hidden group/modal">
        <!-- Background decorative glow -->
        <div class="absolute -top-24 -left-24 w-48 h-48 bg-makoclaw-accent/10 blur-[60px] rounded-full pointer-events-none group-hover/modal:bg-makoclaw-accent/15 transition-all duration-1000" />

        <!-- Header -->
        <div class="flex items-center justify-between p-3 sm:p-4 border-b border-makoclaw-border/30">
          <h3 class="text-lg font-semibold">
            Create New Task
          </h3>
          <button
            class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center hover:bg-makoclaw-accent/10 rounded-xl text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all"
            @click="$emit('close')"
          >
            <svg
              class="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        <!-- Form -->
        <form
          class="p-3 sm:p-4 md:p-5 space-y-3 sm:space-y-4"
          @submit.prevent="handleCreateTask"
        >
          <!-- Title -->
          <div>
            <label
              for="title"
              class="block text-sm font-medium mb-2"
            >
              Task Title
            </label>
            <input
              id="title"
              v-model="form.title"
              type="text"
              placeholder="Enter task title"
              class="w-full px-2.5 sm:px-3 md:px-4 py-1.5 sm:py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent text-sm transition-all backdrop-blur-sm"
              required
              :disabled="isLoading"
            >
          </div>

          <!-- Description -->
          <div>
            <label
              for="description"
              class="block text-sm font-medium mb-2"
            >
              Description (optional)
            </label>
            <textarea
              id="description"
              v-model="form.description"
              placeholder="Enter task description"
              class="w-full px-2.5 sm:px-3 md:px-4 py-1.5 sm:py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent text-sm transition-all backdrop-blur-sm resize-none"
              rows="3"
              :disabled="isLoading"
            />
          </div>

          <!-- Status -->
          <div>
            <label
              for="status"
              class="block text-sm font-medium mb-2"
            >
              Initial Status
            </label>
            <select
              id="status"
              v-model="form.status"
              class="w-full px-2.5 sm:px-3 md:px-4 py-1.5 sm:py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent text-sm transition-all backdrop-blur-sm"
              :disabled="isLoading"
            >
              <option value="backlog">
                Backlog
              </option>
              <option value="todo">
                To Do
              </option>
              <option value="in_progress">
                In Progress
              </option>
              <option value="review">
                Review
              </option>
              <option value="done">
                Done
              </option>
            </select>
          </div>

          <!-- Error Message -->
          <div
            v-if="errorMessage"
            class="p-3 bg-makoclaw-error/20 border border-makoclaw-error rounded text-makoclaw-error text-sm"
          >
            {{ errorMessage }}
          </div>

          <!-- Actions -->
          <div class="flex gap-2 sm:gap-3 pt-3 sm:pt-4 border-t border-makoclaw-border/30">
            <button
              type="button"
              class="flex-1 px-3 py-2 min-h-[36px] border border-makoclaw-border/50 rounded-xl hover:bg-makoclaw-bg transition-all text-sm"
              :disabled="isLoading"
              @click="$emit('close')"
            >
              Cancel
            </button>
            <button
              type="submit"
              class="flex-1 px-3 py-2 min-h-[36px] bg-gradient-to-r from-makoclaw-accent to-blue-600 hover:from-makoclaw-accent-hover hover:to-blue-700 text-white rounded-xl transition-all shadow-md shadow-makoclaw-accent/20 hover:shadow-makoclaw-accent/40 hover:-translate-y-0.5 text-sm font-medium disabled:opacity-50 active:scale-95"
              :disabled="isLoading"
            >
              {{ isLoading ? 'Creating...' : 'Create Task' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </Transition>
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
