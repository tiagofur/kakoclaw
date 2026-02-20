<template>
  <div
    v-if="show"
    class="fixed inset-0 z-[100] flex items-end sm:items-center justify-center p-4"
    @click.self="$emit('close')"
  >
    <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="$emit('close')"></div>

    <div class="relative bg-kakoclaw-surface border border-kakoclaw-border rounded-2xl shadow-2xl w-full max-w-2xl max-h-[85vh] flex flex-col overflow-hidden">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 border-b border-kakoclaw-border bg-kakoclaw-bg/50">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 rounded-lg bg-kakoclaw-accent/10 flex items-center justify-center">
            <svg class="w-4 h-4 text-kakoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
          </div>
          <h2 class="font-bold text-kakoclaw-text">Prompt Library</h2>
        </div>
        <div class="flex items-center gap-2">
          <button
            type="button"
            @click="showForm = !showForm"
            class="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-kakoclaw-accent text-white rounded-lg hover:bg-kakoclaw-accent/90 transition-colors"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
            New Prompt
          </button>
          <button type="button" @click="$emit('close')" class="p-1.5 text-kakoclaw-text-secondary hover:text-kakoclaw-text rounded-lg transition-colors">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>
      </div>

      <!-- Create/Edit Form -->
      <div v-if="showForm" class="p-4 border-b border-kakoclaw-border bg-kakoclaw-bg/30">
        <div class="space-y-3">
          <input
            v-model="form.title"
            type="text"
            placeholder="Prompt title (e.g. Code Reviewer)"
            class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm text-kakoclaw-text outline-none focus:border-kakoclaw-accent"
          >
          <input
            v-model="form.description"
            type="text"
            placeholder="Short description (optional)"
            class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm text-kakoclaw-text outline-none focus:border-kakoclaw-accent"
          >
          <input
            v-model="form.tags"
            type="text"
            placeholder="Tags comma-separated: coding, writing, analysis"
            class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm text-kakoclaw-text outline-none focus:border-kakoclaw-accent"
          >
          <textarea
            v-model="form.content"
            placeholder="Prompt content..."
            rows="5"
            class="w-full bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-3 py-2 text-sm text-kakoclaw-text outline-none focus:border-kakoclaw-accent resize-none font-mono"
          ></textarea>
          <div class="flex gap-2 justify-end">
            <button type="button" @click="cancelForm" class="px-4 py-1.5 text-sm border border-kakoclaw-border rounded-lg text-kakoclaw-text-secondary hover:bg-kakoclaw-bg transition-colors">Cancel</button>
            <button
              type="button"
              @click="savePrompt"
              :disabled="saving || !form.title || !form.content"
              class="px-4 py-1.5 text-sm bg-kakoclaw-accent text-white rounded-lg hover:bg-kakoclaw-accent/90 disabled:opacity-50 transition-colors flex items-center gap-1.5"
            >
              <div v-if="saving" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
              {{ editingId ? 'Update' : 'Save' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Search -->
      <div class="px-4 pt-3 pb-2">
        <div class="relative">
          <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-kakoclaw-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            v-model="search"
            type="text"
            placeholder="Search prompts..."
            class="w-full pl-9 pr-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm text-kakoclaw-text outline-none focus:border-kakoclaw-accent"
          >
        </div>
      </div>

      <!-- Prompt List -->
      <div class="flex-1 overflow-auto custom-scrollbar px-4 pb-4">
        <div v-if="loading" class="flex items-center justify-center py-12">
          <div class="w-6 h-6 border-2 border-kakoclaw-accent border-t-transparent rounded-full animate-spin"></div>
        </div>

        <div v-else-if="filteredPrompts.length === 0" class="flex flex-col items-center justify-center py-12 text-kakoclaw-text-secondary">
          <svg class="w-10 h-10 mb-3 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          <p class="text-sm">{{ search ? 'No prompts matching search' : 'No prompts saved yet' }}</p>
          <p v-if="!search" class="text-xs mt-1 opacity-60">Click "New Prompt" to save your first template</p>
        </div>

        <div v-else class="space-y-2 pt-1">
          <div
            v-for="prompt in filteredPrompts"
            :key="prompt.id"
            class="group bg-kakoclaw-bg border border-kakoclaw-border rounded-xl p-4 hover:border-kakoclaw-accent/40 transition-all cursor-pointer"
            @click="usePrompt(prompt)"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 flex-wrap">
                  <h3 class="text-sm font-semibold text-kakoclaw-text">{{ prompt.title }}</h3>
                  <div v-if="prompt.tags" class="flex flex-wrap gap-1">
                    <span
                      v-for="tag in prompt.tags.split(',').filter(t => t.trim())"
                      :key="tag"
                      class="px-1.5 py-0.5 text-[10px] rounded-full bg-kakoclaw-accent/10 text-kakoclaw-accent border border-kakoclaw-accent/20"
                    >{{ tag.trim() }}</span>
                  </div>
                </div>
                <p v-if="prompt.description" class="text-xs text-kakoclaw-text-secondary mt-0.5">{{ prompt.description }}</p>
                <p class="text-xs text-kakoclaw-text-secondary/70 mt-2 line-clamp-2 font-mono">{{ prompt.content }}</p>
              </div>
              <!-- Action buttons -->
              <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0" @click.stop>
                <button
                  type="button"
                  @click="editPrompt(prompt)"
                  class="p-1.5 text-kakoclaw-text-secondary hover:text-kakoclaw-accent rounded-lg hover:bg-kakoclaw-accent/10 transition-colors"
                  title="Edit"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
                </button>
                <button
                  type="button"
                  @click="deletePrompt(prompt.id)"
                  class="p-1.5 text-kakoclaw-text-secondary hover:text-red-400 rounded-lg hover:bg-red-400/10 transition-colors"
                  title="Delete"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                </button>
              </div>
            </div>
            <!-- Use hint -->
            <div class="mt-3 opacity-0 group-hover:opacity-100 transition-opacity">
              <p class="text-[10px] text-kakoclaw-accent flex items-center gap-1">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5" /></svg>
                Click to insert into chat
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'

const props = defineProps({
  show: { type: Boolean, default: false }
})
const emit = defineEmits(['close', 'use'])
const toast = useToast()

const prompts = ref([])
const loading = ref(false)
const saving = ref(false)
const search = ref('')
const showForm = ref(false)
const editingId = ref(null)
const form = ref({ title: '', content: '', description: '', tags: '' })

const filteredPrompts = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return prompts.value
  return prompts.value.filter(p =>
    p.title.toLowerCase().includes(q) ||
    p.content.toLowerCase().includes(q) ||
    (p.description || '').toLowerCase().includes(q) ||
    (p.tags || '').toLowerCase().includes(q)
  )
})

const loadPrompts = async () => {
  loading.value = true
  try {
    const data = await advancedService.fetchPrompts()
    prompts.value = data.prompts || []
  } catch (err) {
    console.error('Failed to load prompts:', err)
  } finally {
    loading.value = false
  }
}

const savePrompt = async () => {
  saving.value = true
  try {
    if (editingId.value) {
      await advancedService.updatePrompt(editingId.value, form.value)
      toast.success('Prompt updated')
    } else {
      await advancedService.createPrompt(form.value)
      toast.success('Prompt saved')
    }
    await loadPrompts()
    cancelForm()
  } catch (err) {
    toast.error('Failed to save prompt')
  } finally {
    saving.value = false
  }
}

const editPrompt = (p) => {
  form.value = { title: p.title, content: p.content, description: p.description || '', tags: p.tags || '' }
  editingId.value = p.id
  showForm.value = true
}

const deletePrompt = async (id) => {
  try {
    await advancedService.deletePrompt(id)
    toast.success('Prompt deleted')
    await loadPrompts()
  } catch {
    toast.error('Failed to delete prompt')
  }
}

const cancelForm = () => {
  showForm.value = false
  editingId.value = null
  form.value = { title: '', content: '', description: '', tags: '' }
}

const usePrompt = (p) => {
  emit('use', p.content)
  emit('close')
}

watch(() => props.show, (v) => {
  if (v) loadPrompts()
})
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(139, 92, 246, 0.2); border-radius: 3px; }

.modal-enter-active, .modal-leave-active { transition: opacity 0.2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
