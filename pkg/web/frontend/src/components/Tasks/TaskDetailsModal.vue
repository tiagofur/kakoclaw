<template>
  <Transition name="modal">
  <div class="fixed inset-0 bg-black/40 backdrop-blur-md flex items-center justify-center p-3 sm:p-4 z-modal overflow-y-auto">
    <div class="glass-panel rounded-2xl max-w-2xl md:max-w-3xl lg:max-w-4xl w-full shadow-2xl my-4">
      <!-- Header -->
      <div class="flex items-center justify-between p-3 sm:p-4 border-b border-makoclaw-border/30">
        <div>
          <h3 class="text-lg font-semibold">{{ task.title }}</h3>
          <div class="flex items-center gap-3 mt-1">
            <p class="text-xs text-makoclaw-text-secondary">ID: {{ task.id }}</p>
            <div v-if="task.agent" class="flex items-center gap-1.5 border-l border-makoclaw-border pl-3">
              <span class="text-xs text-makoclaw-text-secondary">Agent:</span>
              <SpecialistBadge :name="task.agent" class="scale-90 origin-left" />
            </div>
          </div>
        </div>
        <button
          @click="$emit('close')"
          class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center hover:bg-makoclaw-accent/10 rounded-xl text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Content -->
      <div class="p-3 sm:p-4 md:p-6 space-y-3 sm:space-y-4 max-h-[60vh] sm:max-h-[70vh] overflow-y-auto custom-scrollbar">
        <!-- Title -->
        <div>
          <label class="block text-sm font-medium mb-2">Title</label>
          <input
            v-model="editForm.title"
            type="text"
            class="w-full px-2.5 sm:px-3 md:px-4 py-1.5 sm:py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent text-sm transition-all backdrop-blur-sm"
            :disabled="isLoading"
          />
        </div>

        <!-- Description -->
        <div>
          <label class="block text-sm font-medium mb-2">Description</label>
          <textarea
            v-model="editForm.description"
            class="w-full px-2.5 sm:px-3 md:px-4 py-1.5 sm:py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent text-sm transition-all backdrop-blur-sm resize-y"
            rows="5"
            :disabled="isLoading"
          ></textarea>
        </div>

        <!-- Status -->
        <div>
          <label class="block text-sm font-medium mb-2">Status</label>
          <select
            v-model="editForm.status"
            class="w-full px-2.5 sm:px-3 md:px-4 py-1.5 sm:py-2 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent text-sm transition-all backdrop-blur-sm"
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
          <div class="flex items-center justify-between mb-2">
            <label class="block text-sm font-medium">Result / Output</label>
            <button 
              v-if="editForm.result"
              @click="copyResult" 
              class="text-xs text-makoclaw-accent hover:underline flex items-center gap-1"
            >
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
              </svg>
              {{ copied ? 'Copied!' : 'Copy' }}
            </button>
          </div>
          <div class="relative group">
            <textarea
              v-model="editForm.result"
              class="w-full px-3 sm:px-4 py-2 sm:py-3 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent text-sm sm:text-base font-mono resize-y leading-relaxed backdrop-blur-sm transition-all"
              rows="12"
              placeholder="Task result/output"
              :disabled="isLoading"
            ></textarea>
          </div>
        </div>

        <!-- Logs -->
        <div v-if="logs.length > 0">
          <label class="block text-sm font-medium mb-2">Execution Logs</label>
          <div class="bg-makoclaw-bg/50 border border-makoclaw-border/30 rounded-xl p-2 sm:p-3 text-xs max-h-36 sm:max-h-48 overflow-y-auto backdrop-blur-sm custom-scrollbar">
            <table class="w-full text-left">
              <tbody class="divide-y divide-makoclaw-border">
                <tr v-for="log in logs" :key="log.id" class="hover:bg-makoclaw-surface transition-smooth">
                  <td class="py-2 px-2 text-makoclaw-text-secondary whitespace-nowrap">
                    {{ formatTime(log.created_at) }}
                  </td>
                  <td class="py-2 px-2">{{ log.event }}</td>
                  <td class="py-2 px-2 text-makoclaw-text-secondary truncate">{{ log.message }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Error Message -->
        <div v-if="errorMessage" class="p-3 bg-makoclaw-error/20 border border-makoclaw-error rounded text-makoclaw-error text-sm">
          {{ errorMessage }}
        </div>
      </div>

      <!-- Actions -->
      <!-- Actions -->
      <div class="border-t border-makoclaw-border/30 p-3 sm:p-4 flex gap-1.5 sm:gap-2 justify-end flex-wrap">
        <button
          v-if="!task.archived"
          @click="handleArchive"
          class="px-3 sm:px-4 py-2 min-h-[36px] bg-yellow-500/20 text-yellow-500 hover:bg-yellow-500/30 border border-yellow-500/50 rounded-xl transition-all text-sm font-medium disabled:opacity-50"
          :disabled="isLoading"
        >
          Archive
        </button>
        <button
          v-if="task.archived"
          @click="handleUnarchive"
          class="px-3 sm:px-4 py-2 min-h-[36px] bg-blue-500/20 text-blue-500 hover:bg-blue-500/30 border border-blue-500/50 rounded-xl transition-all text-sm font-medium disabled:opacity-50"
          :disabled="isLoading"
        >
          Unarchive
        </button>
        <button
          @click="handleDelete"
          class="px-3 sm:px-4 py-2 min-h-[36px] bg-makoclaw-error hover:bg-makoclaw-error/80 text-white rounded-xl transition-all text-sm font-medium disabled:opacity-50"
          :disabled="isLoading"
        >
          Delete
        </button>
        <button
          @click="$emit('close')"
          class="px-3 sm:px-4 py-2 min-h-[36px] border border-makoclaw-border/50 rounded-xl hover:bg-makoclaw-bg transition-all text-sm"
          :disabled="isLoading"
        >
          Cancel
        </button>
        <button
          @click="handleUpdate"
          class="px-3 sm:px-4 py-2 min-h-[36px] bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl transition-all shadow-md shadow-makoclaw-accent/20 hover:shadow-makoclaw-accent/40 text-sm font-medium disabled:opacity-50"
          :disabled="isLoading"
        >
          {{ isLoading ? 'Saving...' : 'Save Changes' }}
        </button>
      </div>
    </div>
  </div>
  </Transition>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import taskService from '../../services/taskService'
import SpecialistBadge from '../Chat/SpecialistBadge.vue'

const props = defineProps({
  task: { type: Object, required: true }
})

const emit = defineEmits(['close', 'updated', 'deleted', 'archived', 'unarchived'])

const editForm = ref({
  title: props.task.title,
  description: props.task.description || '',
  status: props.task.status,
  result: props.task.result || ''
})

const logs = ref([])
const isLoading = ref(false)
const errorMessage = ref('')
const copied = ref(false)

const copyResult = () => {
  if (!editForm.value.result) return
  navigator.clipboard.writeText(editForm.value.result)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}

onMounted(async () => {
  try {
    // Check if task is archived, logs might not exist or be relevant
    if (!props.task.archived) {
         logs.value = await taskService.getTaskLogs(props.task.id)
    }
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

const handleArchive = async () => {
  if (!confirm('Are you sure you want to archive this task?')) return
  isLoading.value = true
  try {
    await emit('archived', props.task.id)
    emit('close')
  } catch (error) {
    errorMessage.value = 'Failed to archive task'
    isLoading.value = false
  }
}

const handleUnarchive = async () => {
  isLoading.value = true
  try {
    await emit('unarchived', props.task.id)
    emit('close')
  } catch (error) {
    errorMessage.value = 'Failed to unarchive task'
    isLoading.value = false
  }
}

const formatTime = (timestamp) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
</script>

