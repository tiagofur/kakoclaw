<template>
  <div class="h-full flex flex-col bg-makoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-makoclaw-border bg-makoclaw-surface flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent">Files</h2>
        <p class="text-sm text-makoclaw-text-secondary mt-1">Browse and manage workspace files</p>
      </div>
      <div class="flex items-center gap-3">
        <input type="file" ref="fileInput" class="hidden" @change="handleFileUpload" multiple />
        <button
          @click="showNewFolderModal = true"
          class="flex items-center gap-2 px-4 py-2 bg-makoclaw-surface border border-makoclaw-border text-makoclaw-text rounded-lg hover:bg-makoclaw-border transition-all text-sm"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
          </svg>
          New Folder
        </button>
        <button
          @click="$refs.fileInput.click()"
          :disabled="uploading"
          class="flex items-center gap-2 px-4 py-2 bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-all text-sm shadow-lg shadow-makoclaw-accent/20 disabled:opacity-50"
        >
          <svg v-if="uploading" class="animate-spin h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0l-4 4m4-4v12" />
          </svg>
          {{ uploading ? 'Uploading...' : 'Upload' }}
        </button>
        <div class="text-sm text-makoclaw-text-secondary font-mono bg-makoclaw-bg px-2 py-1 rounded border border-makoclaw-border">
          /{{ currentPath || '' }}
        </div>
      </div>
    </div>

    <!-- Search Bar -->
    <div class="flex-none px-4 py-3 border-b border-makoclaw-border bg-makoclaw-surface">
      <div class="relative max-w-md">
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-makoclaw-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search files..."
          class="w-full pl-10 pr-8 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm focus:outline-none focus:border-makoclaw-accent transition-colors"
          @input="debouncedSearch"
        />
        <button
          v-if="searchQuery"
          @click="clearSearch"
          class="absolute right-3 top-1/2 -translate-y-1/2 text-makoclaw-text-secondary hover:text-makoclaw-text"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <p v-if="isSearching" class="text-xs text-makoclaw-text-secondary mt-1">Searching...</p>
      <p v-else-if="searchQuery && entries !== null" class="text-xs text-makoclaw-text-secondary mt-1">
        {{ entries.length }} result{{ entries.length !== 1 ? 's' : '' }} for "{{ searchQuery }}"
      </p>
    </div>

    <!-- Breadcrumb -->
    <div class="flex-none px-4 py-2 border-b border-makoclaw-border bg-makoclaw-surface flex items-center gap-1 text-sm overflow-x-auto">
      <button @click="navigateTo('')" class="text-makoclaw-accent hover:underline flex-shrink-0">workspace</button>
      <template v-for="(part, i) in breadcrumbs" :key="i">
        <span class="text-makoclaw-text-secondary">/</span>
        <button @click="navigateTo(breadcrumbs.slice(0, i + 1).join('/'))" class="text-makoclaw-accent hover:underline flex-shrink-0">{{ part }}</button>
      </template>
    </div>

    <!-- Content -->
    <div
      class="flex-1 overflow-auto custom-scrollbar relative"
      :class="{ 'bg-makoclaw-accent/5': isDragging }"
      @dragover.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="onDrop"
    >
      <div v-if="isDragging" class="absolute inset-0 z-50 flex items-center justify-center pointer-events-none">
        <div class="bg-makoclaw-accent text-white px-6 py-3 rounded-full shadow-xl animate-bounce flex items-center gap-2">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          Drop files to upload to /{{ currentPath }}
        </div>
      </div>
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-makoclaw-accent"></div>
      </div>

      <template v-else>
        <!-- Directory listing -->
        <div v-if="entries !== null" class="divide-y divide-makoclaw-border">
          <!-- Parent directory -->
          <button
            v-if="currentPath && !searchQuery"
            @click="navigateUp()"
            class="w-full flex items-center gap-3 px-6 py-3 hover:bg-makoclaw-surface transition-colors text-left"
          >
            <svg class="w-5 h-5 text-makoclaw-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 17l-5-5m0 0l5-5m-5 5h12" />
            </svg>
            <span class="text-makoclaw-text-secondary">..</span>
          </button>

          <div
            v-for="entry in entries"
            :key="entry.path"
            class="w-full flex items-center gap-3 px-6 py-3 hover:bg-makoclaw-surface transition-colors text-left cursor-pointer group"
          >
            <!-- Entry content (clickable for navigation) -->
            <div
              class="flex-1 flex items-center gap-3 min-w-0"
              @click="entry.is_dir ? navigateTo(entry.path) : viewFile(entry.path)"
            >
              <!-- Folder icon -->
              <svg v-if="entry.is_dir" class="w-5 h-5 text-yellow-400 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
              </svg>
              <!-- File icon -->
              <svg v-else class="w-5 h-5 text-makoclaw-text-secondary flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <div class="flex-1 min-w-0">
                <!-- Inline rename -->
                <input
                  v-if="renamingEntry === entry.path"
                  v-model="renameValue"
                  type="text"
                  class="w-full text-sm font-medium bg-makoclaw-bg border border-makoclaw-accent rounded px-2 py-0.5 focus:outline-none"
                  @keyup.enter="confirmRename(entry)"
                  @keyup.esc="cancelRename"
                  @blur="cancelRename"
                  @click.stop
                  ref="renameInput"
                />
                <span
                  v-else
                  class="text-sm font-medium truncate block"
                  @dblclick.stop="startRename(entry)"
                  :title="'Double-click to rename'"
                >{{ entry.name }}</span>
              </div>
            </div>
            <span v-if="!entry.is_dir" class="text-xs text-makoclaw-text-secondary flex-shrink-0">{{ formatSize(entry.size) }}</span>
            <span class="text-xs text-makoclaw-text-secondary flex-shrink-0">{{ formatDate(entry.mod_time) }}</span>

            <!-- Action buttons -->
            <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <!-- Rename button -->
              <button
                @click.stop="startRename(entry)"
                class="p-1.5 text-makoclaw-text-secondary hover:text-makoclaw-accent hover:bg-makoclaw-bg rounded-lg transition-all"
                title="Rename"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                </svg>
              </button>
              <!-- Download button -->
              <button
                @click.stop="downloadEntry(entry.path)"
                class="p-1.5 text-makoclaw-text-secondary hover:text-makoclaw-accent hover:bg-makoclaw-bg rounded-lg transition-all"
                title="Download"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
              </button>
              <!-- Delete button -->
              <button
                @click.stop="confirmDelete(entry)"
                :disabled="deletingPath === entry.path"
                class="p-1.5 text-makoclaw-text-secondary hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all disabled:opacity-50"
                title="Delete"
              >
                <svg v-if="deletingPath === entry.path" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </div>

          <div v-if="entries.length === 0" class="text-center py-12 text-makoclaw-text-secondary">
            <p v-if="searchQuery">No files found matching "{{ searchQuery }}"</p>
            <p v-else>Empty directory</p>
          </div>
        </div>

        <!-- File viewer -->
        <div v-if="fileContent !== null" class="p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="font-semibold">{{ fileName }}</h3>
            <div class="flex items-center gap-4">
              <!-- Edit button for text files -->
              <button
                v-if="isTextFile && !isEditing"
                @click="startEditing"
                class="flex items-center gap-2 px-3 py-1.5 text-sm bg-makoclaw-surface border border-makoclaw-border rounded-lg hover:bg-makoclaw-border transition-colors text-makoclaw-accent"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                </svg>
                Edit
              </button>
              <!-- Save/Cancel buttons when editing -->
              <template v-if="isEditing">
                <button
                  @click="cancelEditing"
                  class="flex items-center gap-2 px-3 py-1.5 text-sm bg-makoclaw-surface border border-makoclaw-border rounded-lg hover:bg-makoclaw-border transition-colors text-makoclaw-text-secondary"
                >
                  Cancel
                </button>
                <button
                  @click="saveFileContent"
                  :disabled="savingFile"
                  class="flex items-center gap-2 px-3 py-1.5 text-sm bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors disabled:opacity-50"
                >
                  <svg v-if="savingFile" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  {{ savingFile ? 'Saving...' : 'Save' }}
                </button>
              </template>
              <button
                @click="downloadEntry(currentPath)"
                class="flex items-center gap-2 px-3 py-1.5 text-sm bg-makoclaw-surface border border-makoclaw-border rounded-lg hover:bg-makoclaw-border transition-colors text-makoclaw-accent"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
                Download
              </button>
              <span class="text-sm text-makoclaw-text-secondary">{{ formatSize(fileSize) }}</span>
            </div>
          </div>
          <div v-if="fileError" class="text-yellow-400 text-sm mb-4">{{ fileError }}</div>
          <!-- Edit mode -->
          <textarea
            v-else-if="isEditing"
            v-model="editContent"
            class="w-full bg-makoclaw-surface border border-makoclaw-border rounded-xl p-4 text-sm font-mono overflow-auto min-h-[50vh] max-h-[70vh] whitespace-pre-wrap focus:outline-none focus:border-makoclaw-accent resize-y"
          ></textarea>
          <!-- View mode -->
          <pre v-else class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-4 text-sm font-mono overflow-auto max-h-[70vh] whitespace-pre-wrap">{{ fileContent }}</pre>
        </div>
      </template>
    </div>

    <!-- New Folder Modal -->
    <div v-if="showNewFolderModal" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div class="bg-makoclaw-bg border border-makoclaw-border rounded-xl w-full max-w-md shadow-2xl">
        <div class="p-4 border-b border-makoclaw-border">
          <h3 class="font-bold text-lg">Create New Folder</h3>
        </div>
        <div class="p-4">
          <label class="block text-sm text-makoclaw-text-secondary mb-2">Folder name</label>
          <input
            v-model="newFolderName"
            type="text"
            placeholder="Enter folder name"
            class="w-full px-4 py-2 bg-makoclaw-surface border border-makoclaw-border rounded-lg text-sm focus:outline-none focus:border-makoclaw-accent"
            @keyup.enter="createFolder"
            ref="folderNameInput"
          />
          <p class="text-xs text-makoclaw-text-secondary mt-2">
            Will be created in: /{{ currentPath || 'workspace' }}
          </p>
        </div>
        <div class="p-4 border-t border-makoclaw-border flex justify-end gap-3">
          <button
            @click="showNewFolderModal = false; newFolderName = ''"
            class="px-4 py-2 text-sm bg-makoclaw-surface border border-makoclaw-border rounded-lg hover:bg-makoclaw-border transition-colors"
          >
            Cancel
          </button>
          <button
            @click="createFolder"
            :disabled="!newFolderName.trim() || creatingFolder"
            class="px-4 py-2 text-sm bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors disabled:opacity-50"
          >
            {{ creatingFolder ? 'Creating...' : 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="deleteConfirmEntry" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div class="bg-makoclaw-bg border border-makoclaw-border rounded-xl w-full max-w-md shadow-2xl">
        <div class="p-4 border-b border-makoclaw-border">
          <h3 class="font-bold text-lg text-red-400">Confirm Delete</h3>
        </div>
        <div class="p-4">
          <p class="text-sm">
            Are you sure you want to delete
            <span class="font-semibold">"{{ deleteConfirmEntry.name }}"</span>?
          </p>
          <p v-if="deleteConfirmEntry.is_dir" class="text-xs text-yellow-400 mt-2">
            Warning: This will delete the folder and all its contents.
          </p>
          <p class="text-xs text-makoclaw-text-secondary mt-2">
            This action cannot be undone.
          </p>
        </div>
        <div class="p-4 border-t border-makoclaw-border flex justify-end gap-3">
          <button
            @click="deleteConfirmEntry = null"
            class="px-4 py-2 text-sm bg-makoclaw-surface border border-makoclaw-border rounded-lg hover:bg-makoclaw-border transition-colors"
          >
            Cancel
          </button>
          <button
            @click="deleteEntry"
            :disabled="deletingPath"
            class="px-4 py-2 text-sm bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors disabled:opacity-50"
          >
            {{ deletingPath ? 'Deleting...' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const loading = ref(true)
const uploading = ref(false)
const isDragging = ref(false)
const currentPath = ref('')
const entries = ref(null)
const fileContent = ref(null)
const fileName = ref('')
const fileSize = ref(0)
const fileError = ref(null)

// Search state
const searchQuery = ref('')
const isSearching = ref(false)
let searchTimeout = null

// New folder state
const showNewFolderModal = ref(false)
const newFolderName = ref('')
const creatingFolder = ref(false)
const folderNameInput = ref(null)

// Delete state
const deleteConfirmEntry = ref(null)
const deletingPath = ref(null)

// Rename state
const renamingEntry = ref(null)
const renameValue = ref('')
const renameInput = ref(null)

// File editing state
const isEditing = ref(false)
const editContent = ref('')
const savingFile = ref(false)

// Text file extensions that can be edited
const textExtensions = ['.txt', '.md', '.json', '.js', '.ts', '.jsx', '.tsx', '.vue', '.css', '.scss', '.html', '.xml', '.yaml', '.yml', '.toml', '.ini', '.cfg', '.conf', '.log', '.sh', '.bash', '.zsh', '.py', '.go', '.rs', '.java', '.c', '.cpp', '.h', '.hpp', '.rb', '.php', '.sql', '.env', '.gitignore', '.dockerfile', '.makefile']

const isTextFile = computed(() => {
  if (!fileName.value) return false
  const name = fileName.value.toLowerCase()
  return textExtensions.some(ext => name.endsWith(ext)) || !name.includes('.')
})

// Watch for modal open to focus input
watch(showNewFolderModal, async (isOpen) => {
  if (isOpen) {
    await nextTick()
    folderNameInput.value?.focus()
  }
})

const handleFileUpload = async (event) => {
  const files = event.target.files
  if (!files || files.length === 0) return
  await uploadFiles(files)
  event.target.value = '' // Clear input
}

const onDrop = async (event) => {
  isDragging.value = false
  const files = event.dataTransfer.files
  if (!files || files.length === 0) return
  await uploadFiles(files)
}

const uploadFiles = async (files) => {
  uploading.value = true
  let successCount = 0

  for (const file of Array.from(files)) {
    try {
      await advancedService.uploadFile(currentPath.value, file)
      successCount++
    } catch (err) {
      console.error(`Failed to upload ${file.name}:`, err)
      toast.error(`Failed to upload ${file.name}`)
    }
  }

  if (successCount > 0) {
    toast.success(`Successfully uploaded ${successCount} file(s)`)
    await navigateTo(currentPath.value) // Refresh list
  }
  uploading.value = false
}

const breadcrumbs = computed(() => {
  if (!currentPath.value) return []
  return currentPath.value.split('/').filter(Boolean)
})

const navigateTo = async (path) => {
  loading.value = true
  fileContent.value = null
  isEditing.value = false
  editContent.value = ''
  currentPath.value = path
  searchQuery.value = '' // Clear search when navigating
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

const downloadEntry = (path) => {
  advancedService.downloadFile(path)
}

const navigateUp = () => {
  const parts = currentPath.value.split('/').filter(Boolean)
  parts.pop()
  navigateTo(parts.join('/'))
}

// Search functionality with debounce
const debouncedSearch = () => {
  if (searchTimeout) clearTimeout(searchTimeout)

  if (!searchQuery.value.trim()) {
    // If search is cleared, reload the current directory
    navigateTo(currentPath.value)
    return
  }

  isSearching.value = true
  searchTimeout = setTimeout(async () => {
    try {
      const data = await advancedService.searchFiles(currentPath.value, searchQuery.value)
      entries.value = data.entries || []
    } catch (err) {
      console.error('Search failed:', err)
      toast.error('Search failed')
    } finally {
      isSearching.value = false
    }
  }, 300)
}

const clearSearch = () => {
  searchQuery.value = ''
  navigateTo(currentPath.value)
}

// Create folder
const createFolder = async () => {
  const folderName = newFolderName.value.trim()
  if (!folderName) return

  creatingFolder.value = true
  try {
    const path = currentPath.value ? `${currentPath.value}/${folderName}` : folderName
    await advancedService.createFolder(path)
    toast.success(`Folder "${folderName}" created`)
    showNewFolderModal.value = false
    newFolderName.value = ''
    await navigateTo(currentPath.value)
  } catch (err) {
    console.error('Failed to create folder:', err)
    toast.error('Failed to create folder')
  } finally {
    creatingFolder.value = false
  }
}

// Delete functionality
const confirmDelete = (entry) => {
  deleteConfirmEntry.value = entry
}

const deleteEntry = async () => {
  if (!deleteConfirmEntry.value) return

  const entry = deleteConfirmEntry.value
  deletingPath.value = entry.path

  try {
    await advancedService.deleteFile(entry.path)
    toast.success(`"${entry.name}" deleted`)
    deleteConfirmEntry.value = null
    await navigateTo(currentPath.value)
  } catch (err) {
    console.error('Failed to delete:', err)
    toast.error('Failed to delete')
  } finally {
    deletingPath.value = null
  }
}

// Rename functionality
const startRename = async (entry) => {
  renamingEntry.value = entry.path
  renameValue.value = entry.name
  await nextTick()
  // Focus and select the name (without extension for files)
  const input = document.querySelector('input[type="text"][class*="border-makoclaw-accent"]')
  if (input) {
    input.focus()
    const dotIndex = entry.is_dir ? -1 : entry.name.lastIndexOf('.')
    if (dotIndex > 0) {
      input.setSelectionRange(0, dotIndex)
    } else {
      input.select()
    }
  }
}

const cancelRename = () => {
  renamingEntry.value = null
  renameValue.value = ''
}

const confirmRename = async (entry) => {
  const newName = renameValue.value.trim()
  if (!newName || newName === entry.name) {
    cancelRename()
    return
  }

  try {
    await advancedService.renameFile(entry.path, newName)
    toast.success(`Renamed to "${newName}"`)
    cancelRename()
    await navigateTo(currentPath.value)
  } catch (err) {
    console.error('Failed to rename:', err)
    toast.error('Failed to rename')
    cancelRename()
  }
}

// File editing functionality
const startEditing = () => {
  editContent.value = fileContent.value
  isEditing.value = true
}

const cancelEditing = () => {
  isEditing.value = false
  editContent.value = ''
}

const saveFileContent = async () => {
  savingFile.value = true
  try {
    await advancedService.updateFileContent(currentPath.value, editContent.value)
    fileContent.value = editContent.value
    isEditing.value = false
    toast.success('File saved')
  } catch (err) {
    console.error('Failed to save file:', err)
    toast.error('Failed to save file')
  } finally {
    savingFile.value = false
  }
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
