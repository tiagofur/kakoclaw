<template>
  <div class="h-full flex flex-col bg-kakoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-kakoclaw-border bg-kakoclaw-surface flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold bg-gradient-to-r from-kakoclaw-accent to-emerald-500 bg-clip-text text-transparent">Knowledge Base</h2>
        <p class="text-sm text-kakoclaw-text-secondary mt-1">Upload documents to give the AI context for better answers</p>
      </div>
      <div class="flex items-center gap-2">
        <span class="text-sm text-kakoclaw-text-secondary">{{ documents.length }} document{{ documents.length !== 1 ? 's' : '' }}</span>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6 custom-scrollbar">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
      </div>

      <template v-else>
        <!-- Upload Area -->
        <div
          class="border-2 border-dashed rounded-xl p-8 text-center transition-colors mb-6 cursor-pointer"
          :class="dragOver ? 'border-kakoclaw-accent bg-kakoclaw-accent/5' : 'border-kakoclaw-border hover:border-kakoclaw-accent/50'"
          @dragover.prevent="dragOver = true"
          @dragleave.prevent="dragOver = false"
          @drop.prevent="handleDrop"
          @click="$refs.fileInput.click()"
        >
          <input ref="fileInput" type="file" class="hidden" accept=".txt,.md,.pdf,.json,.csv,.html,.xml,.yaml,.yml,.log" multiple @change="handleFileSelect" />
          <svg class="w-10 h-10 mx-auto text-kakoclaw-text-secondary mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          <p class="text-kakoclaw-text-secondary">
            <span v-if="uploading" class="text-kakoclaw-accent">Uploading...</span>
            <span v-else>Drop files here or <span class="text-kakoclaw-accent font-medium">click to browse</span></span>
          </p>
          <p class="text-xs text-kakoclaw-text-secondary mt-1">Supports TXT, MD, PDF, JSON, CSV, HTML, XML, YAML, LOG</p>
        </div>

        <!-- Search -->
        <div class="mb-6">
          <div class="flex gap-2">
            <div class="flex-1 relative">
              <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-kakoclaw-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              <input
                v-model="searchQuery"
                type="text"
                placeholder="Search knowledge base..."
                class="w-full pl-10 pr-4 py-2.5 bg-kakoclaw-surface border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent transition-colors"
                @keyup.enter="runSearch"
              />
            </div>
            <button
              @click="runSearch"
              :disabled="!searchQuery.trim()"
              class="px-4 py-2.5 bg-kakoclaw-accent text-white rounded-lg text-sm font-medium hover:bg-kakoclaw-accent/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >Search</button>
          </div>

          <!-- Search Results -->
          <div v-if="searchResults.length > 0" class="mt-4 space-y-3">
            <h3 class="text-sm font-semibold text-kakoclaw-text-secondary">Search Results ({{ searchResults.length }})</h3>
            <div
              v-for="(result, idx) in searchResults"
              :key="idx"
              class="bg-kakoclaw-surface border border-kakoclaw-border rounded-lg p-4"
            >
              <div class="flex items-center gap-2 mb-2">
                <span class="text-xs font-medium text-kakoclaw-accent">{{ result.document_name }}</span>
                <span class="text-xs text-kakoclaw-text-secondary">chunk #{{ result.position }}</span>
                <span class="text-xs text-kakoclaw-text-secondary ml-auto">score: {{ result.rank?.toFixed(2) }}</span>
              </div>
              <p class="text-sm text-kakoclaw-text whitespace-pre-wrap line-clamp-4">{{ result.content }}</p>
            </div>
          </div>
          <div v-else-if="searchPerformed && searchResults.length === 0" class="mt-4 text-center py-4 text-kakoclaw-text-secondary text-sm">
            No results found for "{{ lastSearchQuery }}"
          </div>
        </div>

        <!-- Documents List -->
        <div>
          <h3 class="text-sm font-semibold text-kakoclaw-text-secondary mb-3">Documents</h3>
          <div v-if="documents.length === 0" class="text-center py-12 text-kakoclaw-text-secondary">
            <svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
            </svg>
            <p class="text-lg">No documents yet</p>
            <p class="text-sm mt-1">Upload documents to build your knowledge base</p>
          </div>
          <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div
              v-for="doc in documents"
              :key="doc.id"
              class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5 hover:border-kakoclaw-accent/50 transition-colors"
            >
              <div class="flex items-start justify-between">
                <div class="flex-1 min-w-0">
                  <h4 class="font-semibold truncate" :title="doc.name">{{ doc.name }}</h4>
                  <div class="flex flex-wrap gap-2 mt-2 text-xs text-kakoclaw-text-secondary">
                    <span class="px-2 py-0.5 bg-kakoclaw-bg rounded-full">{{ doc.mime_type || 'text/plain' }}</span>
                    <span class="px-2 py-0.5 bg-kakoclaw-bg rounded-full">{{ formatSize(doc.size) }}</span>
                    <span class="px-2 py-0.5 bg-kakoclaw-bg rounded-full">{{ doc.chunk_count }} chunk{{ doc.chunk_count !== 1 ? 's' : '' }}</span>
                  </div>
                </div>
              </div>
              <div class="flex items-center justify-between mt-4">
                <span class="text-xs text-kakoclaw-text-secondary">{{ formatDate(doc.created_at) }}</span>
                <div class="flex gap-2">
                  <button
                    @click="openDocViewer(doc)"
                    class="px-3 py-1.5 text-xs text-kakoclaw-text bg-kakoclaw-bg border border-kakoclaw-border rounded-lg hover:bg-kakoclaw-surface hover:text-kakoclaw-accent transition-colors"
                  >
                    View
                  </button>
                  <button
                    @click="deleteDoc(doc.id, doc.name)"
                    :disabled="deleting === doc.id"
                    class="px-3 py-1.5 text-xs text-red-400 bg-red-500/10 rounded-lg hover:bg-red-500/20 transition-colors disabled:opacity-50"
                  >
                    <span v-if="deleting === doc.id">Deleting...</span>
                    <span v-else>Delete</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Document Viewer Modal -->
    <div v-if="selectedDoc" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div class="bg-kakoclaw-bg border border-kakoclaw-border rounded-xl w-full max-w-4xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden text-kakoclaw-text">
        <div class="p-4 border-b border-kakoclaw-border bg-kakoclaw-surface flex items-center justify-between flex-none">
          <div>
            <h3 class="font-bold text-lg text-kakoclaw-text">{{ selectedDoc.name }}</h3>
            <p class="text-xs text-kakoclaw-text-secondary mt-1">
              {{ selectedDoc.chunk_count }} chunks • {{ formatSize(selectedDoc.size) }}
            </p>
          </div>
          <button @click="closeDocViewer" class="p-2 hover:bg-kakoclaw-border rounded-lg text-kakoclaw-text-secondary transition-colors">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="flex-1 overflow-auto p-6 bg-kakoclaw-bg">
          <div v-if="loadingChunks" class="flex justify-center items-center h-40">
             <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
          </div>
          <div v-else class="space-y-6">
             <div v-for="chunk in docChunks" :key="chunk.id" class="bg-kakoclaw-surface border border-kakoclaw-border rounded-lg p-4 hover:border-kakoclaw-accent/30 transition-colors">
                <div class="flex items-center justify-between mb-3 pb-2 border-b border-kakoclaw-border">
                   <h5 class="text-xs font-bold text-kakoclaw-text-secondary uppercase tracking-wider">Chunk #{{ chunk.position + 1 }}</h5>
                   <button v-if="editingChunkId !== chunk.id" @click="startEditingChunk(chunk)" class="text-xs text-kakoclaw-accent hover:underline flex items-center gap-1">
                     <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg>
                     Edit
                   </button>
                   <div v-else class="flex gap-2">
                     <button @click="cancelEditingChunk" class="text-xs text-kakoclaw-text-secondary hover:underline">Cancel</button>
                     <button @click="saveChunk(chunk.id)" :disabled="savingChunk" class="text-xs bg-kakoclaw-accent text-white px-2 py-1 rounded hover:bg-kakoclaw-accent-hover disabled:opacity-50">
                       {{ savingChunk ? 'Saving...' : 'Save' }}
                     </button>
                   </div>
                </div>

                <textarea
                   v-if="editingChunkId === chunk.id"
                   v-model="editChunkContent"
                   class="w-full h-40 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg p-3 text-sm font-mono focus:border-kakoclaw-accent outline-none text-kakoclaw-text resize-y"
                ></textarea>
                <div v-else class="text-sm text-kakoclaw-text whitespace-pre-wrap font-mono leading-relaxed bg-kakoclaw-bg/50 p-3 rounded-lg border border-kakoclaw-border/30">
                  {{ chunk.content }}
                </div>
             </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const loading = ref(true)
const uploading = ref(false)
const dragOver = ref(false)
const deleting = ref(null)
const documents = ref([])
const searchQuery = ref('')
const lastSearchQuery = ref('')
const searchResults = ref([])
const searchPerformed = ref(false)

// Chunk Viewer/Editor state
const selectedDoc = ref(null)
const docChunks = ref([])
const loadingChunks = ref(false)
const editingChunkId = ref(null)
const editChunkContent = ref('')
const savingChunk = ref(false)

const loadDocuments = async () => {
  loading.value = true
  try {
    const data = await advancedService.fetchKnowledgeDocs()
    documents.value = data.documents || []
  } catch (err) {
    console.error('Failed to load knowledge documents:', err)
    toast.error('Failed to load documents')
  } finally {
    loading.value = false
  }
}

const uploadFiles = async (files) => {
  if (!files || files.length === 0) return
  uploading.value = true
  let successCount = 0
  let failCount = 0
  for (const file of files) {
    try {
      await advancedService.uploadKnowledgeDoc(file)
      successCount++
    } catch (err) {
      console.error(`Failed to upload ${file.name}:`, err)
      failCount++
    }
  }
  uploading.value = false
  if (successCount > 0) {
    toast.success(`Uploaded ${successCount} document${successCount !== 1 ? 's' : ''}`)
    await loadDocuments()
  }
  if (failCount > 0) {
    toast.error(`Failed to upload ${failCount} file${failCount !== 1 ? 's' : ''}`)
  }
}

const handleDrop = (e) => {
  dragOver.value = false
  const files = e.dataTransfer?.files
  if (files) uploadFiles(Array.from(files))
}

const handleFileSelect = (e) => {
  const files = e.target?.files
  if (files) uploadFiles(Array.from(files))
  e.target.value = '' // reset so same file can be re-uploaded
}

const deleteDoc = async (id, name) => {
  if (!confirm(`Delete "${name}"? This cannot be undone.`)) return
  deleting.value = id
  try {
    await advancedService.deleteKnowledgeDoc(id)
    toast.success('Document deleted')
    await loadDocuments()
  } catch (err) {
    console.error('Failed to delete document:', err)
    toast.error('Failed to delete document')
  } finally {
    deleting.value = null
  }
}

// Viewer and Chunk Edit Logic
const openDocViewer = async (doc) => {
  selectedDoc.value = doc
  loadingChunks.value = true
  editingChunkId.value = null
  try {
    const data = await advancedService.fetchKnowledgeChunks(doc.id)
    docChunks.value = data.chunks || []
  } catch (err) {
    console.error('Failed to load doc chunks:', err)
    toast.error('Failed to load document chunks')
    selectedDoc.value = null
  } finally {
    loadingChunks.value = false
  }
}

const closeDocViewer = () => {
  selectedDoc.value = null
  docChunks.value = []
  editingChunkId.value = null
}

const startEditingChunk = (chunk) => {
  editingChunkId.value = chunk.id
  editChunkContent.value = chunk.content
}

const cancelEditingChunk = () => {
  editingChunkId.value = null
  editChunkContent.value = ''
}

const saveChunk = async (chunkId) => {
  const newContent = editChunkContent.value.trim()
  if (!newContent) {
    toast.error('Chunk content cannot be empty')
    return
  }

  savingChunk.value = true
  try {
    await advancedService.updateKnowledgeChunk(chunkId, newContent)
    
    // update local state
    const chunkIndex = docChunks.value.findIndex(c => c.id === chunkId)
    if (chunkIndex !== -1) {
      docChunks.value[chunkIndex].content = newContent
    }
    
    toast.success('Chunk updated')
    editingChunkId.value = null
  } catch (err) {
    console.error('Failed to update chunk', err)
    toast.error('Failed to update chunk')
  } finally {
    savingChunk.value = false
  }
}

const runSearch = async () => {
  const q = searchQuery.value.trim()
  if (!q) return
  lastSearchQuery.value = q
  searchPerformed.value = true
  try {
    const data = await advancedService.searchKnowledge(q)
    searchResults.value = data.results || []
  } catch (err) {
    console.error('Search failed:', err)
    toast.error('Search failed')
    searchResults.value = []
  }
}

const formatSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let size = bytes
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024
    i++
  }
  return `${size.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

onMounted(() => loadDocuments())
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 8px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: rgba(156, 163, 175, 0.5); border-radius: 4px; }
.line-clamp-4 { display: -webkit-box; -webkit-line-clamp: 4; -webkit-box-orient: vertical; overflow: hidden; }
</style>
