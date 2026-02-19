<template>
  <div class="h-full flex flex-col bg-kakoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-kakoclaw-border bg-kakoclaw-surface flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold bg-gradient-to-r from-kakoclaw-accent to-purple-500 bg-clip-text text-transparent">Files</h2>
        <p class="text-sm text-kakoclaw-text-secondary mt-1">Browse workspace files</p>
      </div>
      <div class="text-sm text-kakoclaw-text-secondary font-mono">
        /{{ currentPath || '' }}
      </div>
    </div>

    <!-- Breadcrumb -->
    <div class="flex-none px-4 py-2 border-b border-kakoclaw-border bg-kakoclaw-surface flex items-center gap-1 text-sm overflow-x-auto">
      <button @click="navigateTo('')" class="text-kakoclaw-accent hover:underline flex-shrink-0">workspace</button>
      <template v-for="(part, i) in breadcrumbs" :key="i">
        <span class="text-kakoclaw-text-secondary">/</span>
        <button @click="navigateTo(breadcrumbs.slice(0, i + 1).join('/'))" class="text-kakoclaw-accent hover:underline flex-shrink-0">{{ part }}</button>
      </template>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto custom-scrollbar">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
      </div>

      <template v-else>
        <!-- Directory listing -->
        <div v-if="entries !== null" class="divide-y divide-kakoclaw-border">
          <!-- Parent directory -->
          <button
            v-if="currentPath"
            @click="navigateUp()"
            class="w-full flex items-center gap-3 px-6 py-3 hover:bg-kakoclaw-surface transition-colors text-left"
          >
            <svg class="w-5 h-5 text-kakoclaw-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 17l-5-5m0 0l5-5m-5 5h12" />
            </svg>
            <span class="text-kakoclaw-text-secondary">..</span>
          </button>

          <button
            v-for="entry in entries"
            :key="entry.path"
            @click="entry.is_dir ? navigateTo(entry.path) : viewFile(entry.path)"
            class="w-full flex items-center gap-3 px-6 py-3 hover:bg-kakoclaw-surface transition-colors text-left"
          >
            <!-- Folder icon -->
            <svg v-if="entry.is_dir" class="w-5 h-5 text-yellow-400 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
              <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
            </svg>
            <!-- File icon -->
            <svg v-else class="w-5 h-5 text-kakoclaw-text-secondary flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <div class="flex-1 min-w-0">
              <span class="text-sm font-medium truncate block">{{ entry.name }}</span>
            </div>
            <span v-if="!entry.is_dir" class="text-xs text-kakoclaw-text-secondary flex-shrink-0">{{ formatSize(entry.size) }}</span>
            <span class="text-xs text-kakoclaw-text-secondary flex-shrink-0">{{ formatDate(entry.mod_time) }}</span>
          </button>

          <div v-if="entries.length === 0" class="text-center py-12 text-kakoclaw-text-secondary">
            <p>Empty directory</p>
          </div>
        </div>

        <!-- File viewer -->
        <div v-if="fileContent !== null" class="p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="font-semibold">{{ fileName }}</h3>
            <span class="text-sm text-kakoclaw-text-secondary">{{ formatSize(fileSize) }}</span>
          </div>
          <div v-if="fileError" class="text-yellow-400 text-sm mb-4">{{ fileError }}</div>
          <pre v-else class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-4 text-sm font-mono overflow-auto max-h-[70vh] whitespace-pre-wrap">{{ fileContent }}</pre>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const loading = ref(true)
const currentPath = ref('')
const entries = ref(null)
const fileContent = ref(null)
const fileName = ref('')
const fileSize = ref(0)
const fileError = ref(null)

const breadcrumbs = computed(() => {
  if (!currentPath.value) return []
  return currentPath.value.split('/').filter(Boolean)
})

const navigateTo = async (path) => {
  loading.value = true
  fileContent.value = null
  currentPath.value = path
  try {
    const data = await advancedService.fetchFiles(path)
    if (data.entries !== undefined) {
      entries.value = data.entries || []
    } else if (data.content !== undefined) {
      // It's a file
      entries.value = null
      fileContent.value = data.content
      fileName.value = data.name
      fileSize.value = data.size
      fileError.value = null
    } else if (data.error) {
      entries.value = null
      fileContent.value = null
      fileError.value = data.error
      fileName.value = data.name
      fileSize.value = data.size
    }
  } catch (err) {
    console.error('Failed to browse files:', err)
    toast.error('Failed to load files')
  } finally {
    loading.value = false
  }
}

const viewFile = async (path) => {
  await navigateTo(path)
}

const navigateUp = () => {
  const parts = currentPath.value.split('/').filter(Boolean)
  parts.pop()
  navigateTo(parts.join('/'))
}

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024
    i++
  }
  return `${bytes.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString()
}

onMounted(() => navigateTo(''))
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 8px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: rgba(156, 163, 175, 0.5); border-radius: 4px; }
</style>
