<template>
  <div class="h-full flex flex-col">
    <!-- Header with filters -->
    <div class="flex-none p-4 border-b border-picoclaw-border bg-picoclaw-surface sticky top-0 z-10">
      <div class="flex justify-between items-center mb-3">
        <h2 class="text-xl font-bold bg-gradient-to-r from-picoclaw-accent to-purple-500 bg-clip-text text-transparent">Chat History</h2>
        <div class="flex items-center gap-2">
          <div class="relative" ref="exportDropdownRef">
            <button @click="showExportMenu = !showExportMenu" class="p-2 hover:bg-picoclaw-bg rounded-lg text-picoclaw-text-secondary transition-colors" title="Export chat history">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
            </button>
            <div v-if="showExportMenu" class="absolute right-0 top-full mt-1 w-48 bg-picoclaw-surface border border-picoclaw-border rounded-lg shadow-lg p-1 z-50">
              <button @click="handleExportAll" class="w-full text-left px-3 py-2 hover:bg-picoclaw-bg rounded text-sm transition-colors">Export All Chats</button>
              <button v-if="activeSessionId" @click="handleExportSession" class="w-full text-left px-3 py-2 hover:bg-picoclaw-bg rounded text-sm transition-colors">Export Current Session</button>
            </div>
          </div>
          <button @click="triggerImport" class="p-2 hover:bg-picoclaw-bg rounded-lg text-picoclaw-text-secondary transition-colors" title="Import conversations (ChatGPT, Claude, PicoClaw)">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" /></svg>
          </button>
          <input ref="importFileInput" type="file" accept=".json" class="hidden" @change="handleImportFile" />
          <button @click="loadSessions" class="p-2 hover:bg-picoclaw-bg rounded-lg text-picoclaw-text-secondary transition-colors" title="Refresh">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
          </button>
        </div>
      </div>

      <!-- Search bar -->
      <div class="relative mb-3">
        <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-picoclaw-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
        <input
          v-model="searchQuery"
          @input="onSearchInput"
          type="text"
          placeholder="Search message content..."
          class="w-full pl-10 pr-8 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent transition-colors"
        />
        <button
          v-if="searchQuery"
          @click="clearSearch"
          class="absolute right-2 top-1/2 -translate-y-1/2 p-1 hover:bg-picoclaw-surface rounded text-picoclaw-text-secondary"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
        </button>
      </div>

      <!-- Filter row -->
      <div class="flex items-center gap-3">
        <select v-model="filterType" class="bg-picoclaw-bg border border-picoclaw-border rounded-lg px-3 py-1.5 text-sm outline-none focus:border-picoclaw-accent">
          <option value="all">All Sessions</option>
          <option value="chat">Chats Only</option>
          <option value="task">Tasks Only</option>
        </select>
        <select v-model="filterDate" class="bg-picoclaw-bg border border-picoclaw-border rounded-lg px-3 py-1.5 text-sm outline-none focus:border-picoclaw-accent">
          <option value="all">All Time</option>
          <option value="today">Today</option>
          <option value="7d">Last 7 Days</option>
          <option value="30d">Last 30 Days</option>
        </select>
        <label class="flex items-center gap-1.5 text-sm text-picoclaw-text-secondary cursor-pointer">
          <input type="checkbox" v-model="showArchived" @change="loadSessions" class="rounded border-picoclaw-border text-picoclaw-accent focus:ring-picoclaw-accent" />
          Archived
        </label>
        <span class="text-xs text-picoclaw-text-secondary ml-auto">
          {{ isSearchMode ? `${searchResults.length} results` : `${filteredSessions.length} sessions` }}
        </span>
      </div>
    </div>

    <div class="flex-1 flex overflow-hidden">
      <!-- Session List / Search Results -->
      <div class="w-1/3 border-r border-picoclaw-border overflow-y-auto p-2 space-y-1 bg-picoclaw-surface/50">
        <div v-if="loading" class="text-center py-4 text-picoclaw-text-secondary animate-pulse">Loading sessions...</div>
        <div v-else-if="searching" class="text-center py-4 text-picoclaw-text-secondary animate-pulse">Searching...</div>

        <!-- Search results mode -->
        <template v-else-if="isSearchMode">
          <div v-if="searchResults.length === 0" class="text-center py-4 text-picoclaw-text-secondary">
            No messages matching "{{ searchQuery }}"
          </div>
          <div
            v-for="(result, idx) in searchResults"
            :key="idx"
            @click="selectSearchResult(result)"
            class="p-3 rounded-lg cursor-pointer transition-all border border-transparent hover:border-picoclaw-border hover:bg-picoclaw-bg/50"
            :class="selectedSearchResult === idx ? 'bg-picoclaw-bg border-picoclaw-accent/30 shadow-sm' : ''"
          >
            <div class="flex items-center gap-2 mb-1">
              <span class="text-[10px] font-medium px-1.5 py-0.5 rounded"
                :class="result.role === 'user' ? 'bg-picoclaw-accent/20 text-picoclaw-accent' : 'bg-purple-500/20 text-purple-400'">
                {{ result.role }}
              </span>
              <span class="text-[10px] px-1.5 py-0.5 rounded"
                :class="result.session_id.startsWith('web:task:') ? 'bg-amber-500/20 text-amber-400' : 'bg-emerald-500/20 text-emerald-400'">
                {{ result.session_id.startsWith('web:task:') ? 'Task' : 'Chat' }}
              </span>
            </div>
            <div class="text-sm truncate text-picoclaw-text" v-html="highlightMatch(result.content, searchQuery)"></div>
            <div class="text-[10px] text-picoclaw-text-secondary mt-1">{{ formatTime(result.created_at) }}</div>
          </div>
        </template>

        <!-- Normal sessions mode -->
        <template v-else>
          <div v-if="filteredSessions.length === 0" class="text-center py-4 text-picoclaw-text-secondary">No matching history found.</div>

          <div
            v-for="session in filteredSessions"
            :key="session.session_id"
            @click="selectSession(session)"
            class="p-3 rounded-lg cursor-pointer transition-all border border-transparent hover:border-picoclaw-border group"
            :class="selectedSession?.session_id === session.session_id ? 'bg-picoclaw-bg border-picoclaw-accent/30 shadow-sm' : 'hover:bg-picoclaw-bg/50'"
          >
            <!-- Inline rename -->
            <div v-if="renamingSession === session.session_id" class="flex items-center gap-1" @click.stop>
              <input
                v-model="renameInput"
                @keyup.enter="submitRenameSession(session.session_id)"
                @keyup.escape="cancelRenameSession"
                @blur="submitRenameSession(session.session_id)"
                class="flex-1 bg-picoclaw-bg border border-picoclaw-accent rounded px-2 py-1 text-xs text-picoclaw-text focus:outline-none"
                autofocus
                placeholder="Session title..."
              />
            </div>
            <template v-else>
              <div class="flex items-center gap-2 mb-1">
                <!-- Session type icon + badge -->
                <svg v-if="isTaskSession(session)" class="w-4 h-4 text-amber-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" /></svg>
                <svg v-else class="w-4 h-4 text-emerald-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
                <div class="font-medium truncate text-sm flex-1">{{ getSessionTitle(session) }}</div>
                <!-- Action buttons -->
                <div class="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity" @click.stop>
                  <button @click="startRenameSession(session)" class="p-1 hover:bg-picoclaw-border rounded text-picoclaw-text-secondary hover:text-picoclaw-accent transition-colors" title="Rename">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
                  </button>
                  <button @click="archiveSessionById(session.session_id)" class="p-1 hover:bg-picoclaw-border rounded text-picoclaw-text-secondary hover:text-amber-400 transition-colors" :title="showArchived ? 'Unarchive' : 'Archive'">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" /></svg>
                  </button>
                  <button @click="deleteSessionById(session.session_id)" class="p-1 hover:bg-picoclaw-border rounded text-picoclaw-text-secondary hover:text-picoclaw-error transition-colors" title="Delete">
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
                  </button>
                </div>
              </div>
              <div class="text-xs text-picoclaw-text-secondary flex justify-between items-center mt-1.5 pl-6">
                <span>{{ formatTime(session.updated_at) }}</span>
                <div class="flex items-center gap-2">
                  <span v-if="session.message_count" class="text-[10px] text-picoclaw-text-secondary">
                    {{ session.message_count }} msg{{ session.message_count !== 1 ? 's' : '' }}
                  </span>
                  <span class="px-1.5 py-0.5 rounded text-[10px] font-medium"
                    :class="isTaskSession(session) ? 'bg-amber-500/20 text-amber-400' : 'bg-emerald-500/20 text-emerald-400'">
                    {{ isTaskSession(session) ? 'Task' : 'Chat' }}
                  </span>
                </div>
              </div>
            </template>
          </div>
        </template>
      </div>

      <!-- Message View -->
      <div class="flex-1 overflow-y-auto p-4 bg-picoclaw-bg/30 relative">
        <div v-if="!selectedSession && !selectedSearchResultSession" class="h-full flex flex-col items-center justify-center text-picoclaw-text-secondary">
          <svg class="w-12 h-12 mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" /></svg>
          <p class="text-sm">Select a session to view the conversation</p>
          <p class="text-xs mt-1 opacity-60">Or search for specific messages above</p>
        </div>

        <div v-else class="space-y-4 max-w-3xl mx-auto pb-20">
          <div class="text-center mb-6 sticky top-0 z-10 bg-picoclaw-bg/95 backdrop-blur py-2 border-b border-picoclaw-border/50 flex justify-between items-center px-4 rounded-b-xl shadow-sm">
            <div class="flex items-center gap-2 text-xs text-picoclaw-text-secondary">
              <svg v-if="activeSessionIsTask" class="w-3.5 h-3.5 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" /></svg>
              <svg v-else class="w-3.5 h-3.5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
              <span class="font-mono">{{ activeSessionId }}</span>
            </div>
            <button
              @click="continueChat"
              class="flex items-center gap-2 px-3 py-1.5 bg-picoclaw-accent text-white text-xs rounded-lg hover:bg-picoclaw-accent-hover transition-colors shadow-sm"
            >
              <span>Continue Chat</span>
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" /></svg>
            </button>
          </div>

          <div v-if="loadingMessages" class="text-center py-8">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-picoclaw-accent mx-auto"></div>
          </div>

          <div
            v-else
            v-for="(msg, index) in messages"
            :key="index"
            class="flex flex-col gap-1 group"
            :class="msg.role === 'user' ? 'items-end' : 'items-start'"
          >
            <div class="text-xs text-picoclaw-text-secondary px-1 opacity-70 flex items-center gap-2">
              <span>{{ msg.role }}</span>
              <button
                v-if="msg.role === 'assistant'"
                @click="copyMessageContent(msg.content)"
                class="opacity-0 group-hover:opacity-100 transition-opacity p-0.5 hover:bg-picoclaw-surface rounded text-picoclaw-text-secondary hover:text-picoclaw-accent"
                title="Copy response"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" /></svg>
              </button>
              <button
                @click="forkAtMessage(msg)"
                class="opacity-0 group-hover:opacity-100 transition-opacity p-0.5 hover:bg-picoclaw-surface rounded text-picoclaw-text-secondary hover:text-picoclaw-accent"
                title="Fork conversation from this point"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>
              </button>
            </div>
            <div
              class="max-w-[85%] rounded-2xl px-4 py-3 text-sm shadow-sm leading-relaxed"
              :class="msg.role === 'user'
                ? 'bg-picoclaw-accent text-white rounded-tr-none whitespace-pre-wrap'
                : 'bg-white dark:bg-gray-800 border border-picoclaw-border rounded-tl-none'"
            >
              <span v-if="msg.role === 'user'">{{ msg.content }}</span>
              <MarkdownRenderer v-else :content="msg.content" />
            </div>
            <div class="text-[10px] text-gray-400 px-1">{{ formatTime(msg.created_at || msg.timestamp) }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, onBeforeUnmount } from 'vue'
import api from '../services/api'
import taskService from '../services/taskService'
import advancedService from '../services/advancedService'
import MarkdownRenderer from '../components/Chat/MarkdownRenderer.vue'
import { useRouter } from 'vue-router'
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()
const sessions = ref([])
const selectedSession = ref(null)
const messages = ref([])
const loading = ref(false)
const loadingMessages = ref(false)
const filterType = ref('all')
const filterDate = ref('all')
const showArchived = ref(false)

// Session management
const renamingSession = ref(null)
const renameInput = ref('')

// Search state
const searchQuery = ref('')
const searchResults = ref([])
const searching = ref(false)
const selectedSearchResult = ref(null)
const selectedSearchResultSession = ref(null)
let searchTimeout = null

// Export state
const showExportMenu = ref(false)
const exportDropdownRef = ref(null)

// Import state
const importFileInput = ref(null)
const importing = ref(false)

const handleExportAll = () => {
  advancedService.exportChat()
  showExportMenu.value = false
}

const handleExportSession = () => {
  const sid = activeSessionId.value
  if (sid) {
    advancedService.exportChat(sid)
  }
  showExportMenu.value = false
}

const triggerImport = () => {
  importFileInput.value?.click()
}

const handleImportFile = async (event) => {
  const file = event.target.files?.[0]
  if (!file) return

  importing.value = true
  try {
    const text = await file.text()
    const data = JSON.parse(text)
    const result = await advancedService.importChat(data)
    toast.success(`Imported ${result.sessions_imported} session(s) with ${result.messages_imported} message(s)`)
    await loadSessions()
  } catch (error) {
    console.error('Import failed:', error)
    toast.error('Import failed: ' + (error.response?.data || error.message || 'Unknown error'))
  } finally {
    importing.value = false
    // Reset file input so the same file can be re-selected
    if (importFileInput.value) importFileInput.value.value = ''
  }
}

const handleClickOutside = (e) => {
  if (exportDropdownRef.value && !exportDropdownRef.value.contains(e.target)) {
    showExportMenu.value = false
  }
}

const isSearchMode = computed(() => searchQuery.value.trim().length >= 2)

const isTaskSession = (session) => {
  return session.session_id && session.session_id.startsWith('web:task:')
}

const activeSessionId = computed(() => {
  if (selectedSession.value) return selectedSession.value.session_id
  if (selectedSearchResultSession.value) return selectedSearchResultSession.value
  return ''
})

const activeSessionIsTask = computed(() => {
  return activeSessionId.value.startsWith('web:task:')
})

const filteredSessions = computed(() => {
  let result = sessions.value

  // Filter by type
  if (filterType.value !== 'all') {
    result = result.filter(s => {
      const isChat = s.session_id && s.session_id.startsWith('web:chat')
      return filterType.value === 'chat' ? isChat : !isChat
    })
  }

  // Filter by date
  if (filterDate.value !== 'all') {
    const now = new Date()
    let cutoff
    if (filterDate.value === 'today') {
      cutoff = new Date(now.getFullYear(), now.getMonth(), now.getDate())
    } else if (filterDate.value === '7d') {
      cutoff = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
    } else if (filterDate.value === '30d') {
      cutoff = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
    }
    if (cutoff) {
      result = result.filter(s => new Date(s.updated_at) >= cutoff)
    }
  }

  return result
})

const continueChat = () => {
  const sid = activeSessionId.value
  if (sid) {
    router.push({ path: '/chat', query: { id: sid } })
  }
}

const forkAtMessage = async (msg) => {
  const sid = activeSessionId.value
  if (!sid) return

  if (!confirm('Fork conversation from this message? A new session will be created with all messages up to this point.')) return

  try {
    const result = await advancedService.forkChat(sid, msg.id)
    toast.success(`Forked! New session created with ${result.messages_copied} message(s)`)
    await loadSessions()
    // Navigate to the forked session in chat
    router.push({ path: '/chat', query: { id: result.new_session_id } })
  } catch (error) {
    console.error('Fork failed:', error)
    toast.error('Fork failed: ' + (error.response?.data || error.message || 'Unknown error'))
  }
}

const loadSessions = async () => {
  loading.value = true
  try {
    const data = await taskService.fetchChatSessions({ archived: showArchived.value })
    sessions.value = data.sessions || []
  } catch (error) {
    console.error('Failed to load sessions:', error)
    toast.error('Failed to load sessions')
  } finally {
    loading.value = false
  }
}

// Session actions
const deleteSessionById = async (sessionId) => {
  if (!confirm('Delete this session and all its messages? This cannot be undone.')) return
  try {
    await taskService.deleteSession(sessionId)
    sessions.value = sessions.value.filter(s => s.session_id !== sessionId)
    if (selectedSession.value?.session_id === sessionId) {
      selectedSession.value = null
      messages.value = []
    }
    toast.success('Session deleted')
  } catch (error) {
    console.error('Failed to delete session:', error)
    toast.error('Failed to delete session')
  }
}

const archiveSessionById = async (sessionId) => {
  const isArchived = showArchived.value
  try {
    await taskService.updateSession(sessionId, { archived: !isArchived })
    sessions.value = sessions.value.filter(s => s.session_id !== sessionId)
    if (selectedSession.value?.session_id === sessionId) {
      selectedSession.value = null
      messages.value = []
    }
    toast.success(isArchived ? 'Session unarchived' : 'Session archived')
  } catch (error) {
    console.error('Failed to update session:', error)
    toast.error('Failed to update session')
  }
}

const startRenameSession = (session) => {
  renamingSession.value = session.session_id
  renameInput.value = session.title || ''
}

const submitRenameSession = async (sessionId) => {
  if (renamingSession.value !== sessionId) return
  try {
    await taskService.updateSession(sessionId, { title: renameInput.value.trim() })
    const session = sessions.value.find(s => s.session_id === sessionId)
    if (session) session.title = renameInput.value.trim()
    toast.success('Session renamed')
  } catch (error) {
    console.error('Failed to rename session:', error)
    toast.error('Failed to rename session')
  }
  renamingSession.value = null
  renameInput.value = ''
}

const cancelRenameSession = () => {
  renamingSession.value = null
  renameInput.value = ''
}

const selectSession = async (session) => {
  selectedSession.value = session
  selectedSearchResult.value = null
  selectedSearchResultSession.value = null
  loadingMessages.value = true
  messages.value = []

  try {
    const response = await api.get(`/chat/sessions/${encodeURIComponent(session.session_id)}`)
    messages.value = response.data.messages || []
  } catch (error) {
    console.error('Failed to load messages:', error)
  } finally {
    loadingMessages.value = false
  }
}

const selectSearchResult = async (result) => {
  const idx = searchResults.value.indexOf(result)
  selectedSearchResult.value = idx
  selectedSession.value = null
  selectedSearchResultSession.value = result.session_id
  loadingMessages.value = true
  messages.value = []

  try {
    const response = await api.get(`/chat/sessions/${encodeURIComponent(result.session_id)}`)
    messages.value = response.data.messages || []
  } catch (error) {
    console.error('Failed to load session messages:', error)
  } finally {
    loadingMessages.value = false
  }
}

const onSearchInput = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  const q = searchQuery.value.trim()
  if (q.length < 2) {
    searchResults.value = []
    selectedSearchResult.value = null
    selectedSearchResultSession.value = null
    return
  }
  searchTimeout = setTimeout(async () => {
    searching.value = true
    try {
      const data = await taskService.searchMessages(q)
      searchResults.value = data.messages || []
    } catch (error) {
      console.error('Search failed:', error)
      searchResults.value = []
    } finally {
      searching.value = false
    }
  }, 300)
}

const clearSearch = () => {
  searchQuery.value = ''
  searchResults.value = []
  selectedSearchResult.value = null
  selectedSearchResultSession.value = null
}

const highlightMatch = (text, query) => {
  if (!query || !text) return text
  const truncated = text.length > 120 ? text.substring(0, 120) + '...' : text
  const escaped = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return truncated.replace(
    new RegExp(`(${escaped})`, 'gi'),
    '<mark class="bg-picoclaw-accent/30 text-picoclaw-text rounded px-0.5">$1</mark>'
  )
}

const getSessionTitle = (session) => {
  if (session.title) return session.title
  if (session.last_message) {
    if (session.last_message.length > 50) {
      return session.last_message.substring(0, 50) + '...'
    }
    return session.last_message
  }
  return `Session ${session.session_id}`
}

const copyMessageContent = async (content) => {
  try {
    await navigator.clipboard.writeText(content)
    toast.success('Copied to clipboard')
  } catch {
    toast.error('Failed to copy')
  }
}

const formatTime = (ts) => {
  if (!ts) return ''
  return new Date(ts).toLocaleString()
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  loadSessions()
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>
