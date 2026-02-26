<template>
  <div class="space-y-6 max-w-7xl mx-auto animate-fadeIn">
    <!-- Filters -->
    <div class="glass-panel rounded-2xl p-6">
      <h3 class="text-sm font-bold uppercase tracking-widest text-makoclaw-text-secondary opacity-60 mb-4 flex items-center">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-makoclaw-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
        </svg>
        Filters
      </h3>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div>
          <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">User</label>
          <select v-model="filters.user_id" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm focus:ring-2 focus:ring-makoclaw-accent/30 outline-none text-makoclaw-text backdrop-blur-sm transition-all cursor-pointer">
            <option value="">All Users</option>
            <option v-for="user in users" :key="user.id" :value="user.id">{{ user.username }}</option>
          </select>
        </div>

        <div>
          <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">Tool</label>
          <select v-model="filters.tool" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm focus:ring-2 focus:ring-makoclaw-accent/30 outline-none text-makoclaw-text backdrop-blur-sm transition-all cursor-pointer">
            <option value="">All Tools</option>
            <option v-for="tool in toolsList" :key="tool" :value="tool">{{ tool }}</option>
          </select>
        </div>

        <div>
          <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">Status</label>
          <select v-model="filters.success" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm focus:ring-2 focus:ring-makoclaw-accent/30 outline-none text-makoclaw-text backdrop-blur-sm transition-all cursor-pointer">
            <option value="">All</option>
            <option value="true">Success</option>
            <option value="false">Failed</option>
          </select>
        </div>

        <div class="flex items-end">
          <button 
            @click="loadAuditLog"
            class="w-full px-4 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl transition-all text-sm font-bold flex items-center justify-center gap-2 shadow-lg shadow-makoclaw-accent/20 active:scale-95"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            Apply Filters
          </button>
        </div>
      </div>
    </div>

    <!-- Audit Log Table -->
    <div class="glass-panel rounded-2xl p-6">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h3 class="text-sm font-bold uppercase tracking-widest text-makoclaw-text-secondary opacity-60 flex items-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-makoclaw-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            Audit Log
            <span v-if="auditLogs.length > 0" class="ml-2 px-2 py-0.5 text-[10px] font-bold bg-makoclaw-accent/20 text-makoclaw-accent rounded-full">{{ auditLogs.length }}</span>
          </h3>
          <p class="text-xs text-makoclaw-text-secondary mt-1">Tool execution history and security audit trail</p>
        </div>
        <button 
          @click="refreshLog"
          :disabled="loading"
          class="px-3 py-2 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-colors disabled:opacity-50"
          title="Refresh"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>

      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-makoclaw-accent"></div>
      </div>

      <div v-else-if="auditLogs.length === 0" class="text-center py-12">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-makoclaw-text-secondary opacity-50 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <p class="text-makoclaw-text-secondary">No audit logs found</p>
        <p class="text-xs text-makoclaw-text-secondary mt-1">Tool executions will appear here</p>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="uppercase tracking-wider border-b border-makoclaw-border text-[10px] text-makoclaw-text-secondary font-bold">
            <tr>
              <th scope="col" class="px-4 py-3">Timestamp</th>
              <th scope="col" class="px-4 py-3">User</th>
              <th scope="col" class="px-4 py-3">Tool</th>
              <th scope="col" class="px-4 py-3">Status</th>
              <th scope="col" class="px-4 py-3">Duration</th>
              <th scope="col" class="px-4 py-3">Details</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-makoclaw-border text-makoclaw-text">
            <tr v-for="log in auditLogs" :key="log.id" class="hover:bg-makoclaw-bg/30 transition-colors">
              <td class="px-4 py-3 text-xs text-makoclaw-text-secondary whitespace-nowrap">
                {{ formatTimestamp(log.timestamp) }}
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <span class="font-medium">{{ log.username }}</span>
                  <span class="text-[10px] px-1.5 py-0.5 bg-makoclaw-bg/60 rounded font-mono text-makoclaw-text-secondary">ID:{{ log.user_id }}</span>
                </div>
              </td>
              <td class="px-4 py-3 font-mono text-sm">
                <span :class="isRestrictedTool(log.tool) ? 'text-orange-400 font-bold' : ''">
                  {{ log.tool }}
                </span>
              </td>
              <td class="px-4 py-3">
                <span v-if="log.success" class="px-2 py-0.5 text-[10px] font-bold uppercase rounded-full bg-green-500/10 text-green-400">
                  ✓ Success
                </span>
                <span v-else class="px-2 py-0.5 text-[10px] font-bold uppercase rounded-full bg-red-500/10 text-red-400 cursor-help" :title="log.error">
                  ✗ Failed
                </span>
              </td>
              <td class="px-4 py-3 text-xs text-makoclaw-text-secondary">
                <span v-if="log.duration_ms !== null && log.duration_ms !== undefined">{{ log.duration_ms }}ms</span>
                <span v-else>-</span>
              </td>
              <td class="px-4 py-3">
                <button 
                  @click="showDetails(log)"
                  class="text-makoclaw-accent hover:text-makoclaw-accent-hover text-xs font-medium transition-colors"
                >
                  View Details
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Pagination -->
        <div class="flex items-center justify-between mt-4 pt-4 border-t border-makoclaw-border">
          <div class="text-xs text-makoclaw-text-secondary">
            Showing {{ auditLogs.length }} records
          </div>
          <div class="flex gap-2">
            <button 
              @click="loadMore"
              :disabled="loading || auditLogs.length < filters.limit"
              class="px-3 py-1.5 text-xs font-medium text-makoclaw-text-secondary hover:text-makoclaw-accent disabled:opacity-30 transition-colors"
            >
              Load More
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Details Modal -->
    <Transition name="modal">
    <div v-if="selectedLog" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-modal p-4" @click.self="selectedLog = null">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-2xl shadow-2xl max-w-3xl w-full max-h-[80vh] overflow-hidden flex flex-col">
        <div class="p-6 border-b border-makoclaw-border flex items-center justify-between">
          <h3 class="text-lg font-bold text-makoclaw-text">Audit Log Details</h3>
          <button @click="selectedLog = null" class="text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <div class="p-6 overflow-y-auto custom-scrollbar space-y-4">
          <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span class="text-makoclaw-text-secondary block mb-1 text-xs uppercase tracking-wide">Timestamp</span>
              <span class="text-makoclaw-text font-mono">{{ formatTimestamp(selectedLog.timestamp) }}</span>
            </div>
            <div>
              <span class="text-makoclaw-text-secondary block mb-1 text-xs uppercase tracking-wide">User</span>
              <span class="text-makoclaw-text">{{ selectedLog.username }} (ID: {{ selectedLog.user_id }})</span>
            </div>
            <div>
              <span class="text-makoclaw-text-secondary block mb-1 text-xs uppercase tracking-wide">Tool</span>
              <span class="text-makoclaw-text font-mono font-bold">{{ selectedLog.tool }}</span>
            </div>
            <div>
              <span class="text-makoclaw-text-secondary block mb-1 text-xs uppercase tracking-wide">Status</span>
              <span v-if="selectedLog.success" class="text-green-400 font-semibold">✓ Success</span>
              <span v-else class="text-red-400 font-semibold">✗ Failed</span>
            </div>
            <div v-if="selectedLog.duration_ms !== null && selectedLog.duration_ms !== undefined">
              <span class="text-makoclaw-text-secondary block mb-1 text-xs uppercase tracking-wide">Duration</span>
              <span class="text-makoclaw-text">{{ selectedLog.duration_ms }}ms</span>
            </div>
          </div>

          <div v-if="selectedLog.arguments">
            <span class="text-makoclaw-text-secondary block mb-2 text-xs uppercase tracking-wide">Arguments</span>
            <pre class="bg-makoclaw-bg/60 border border-makoclaw-border rounded-lg p-4 text-xs font-mono text-makoclaw-text overflow-x-auto">{{ formatJSON(selectedLog.arguments) }}</pre>
          </div>

          <div v-if="selectedLog.error">
            <span class="text-makoclaw-text-secondary block mb-2 text-xs uppercase tracking-wide">Error</span>
            <div class="bg-red-500/10 border border-red-500/20 rounded-lg p-4 text-sm text-red-400">
              {{ selectedLog.error }}
            </div>
          </div>
        </div>

        <div class="p-4 border-t border-makoclaw-border flex justify-end">
          <button @click="selectedLog = null" class="px-4 py-2 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-lg text-sm font-bold transition-all active:scale-95">
            Close
          </button>
        </div>
      </div>
    </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import toolsService from '../../services/toolsService'
import usersService from '../../services/usersService'
import { useToast } from '../../composables/useToast'
import { useAuthStore } from '../../stores/authStore'

const toast = useToast()
const authStore = useAuthStore()
const loading = ref(true)
const auditLogs = ref([])
const users = ref([])
const selectedLog = ref(null)

const toolsList = [
  'exec', 'spawn', 'email', 'write_file', 'edit_file', 'append_file', 'web_fetch',
  'read_file', 'list_dir', 'web_search', 'message', 'memory', 'query_knowledge', 
  'task_manager', 'schedule'
]

const restrictedTools = ['exec', 'spawn', 'email', 'write_file', 'edit_file', 'append_file', 'web_fetch']

const filters = ref({
  user_id: '',
  tool: '',
  success: '',
  limit: 100,
  offset: 0
})

const formatTimestamp = (timestamp) => {
  if (!timestamp) return 'N/A'
  const date = new Date(timestamp)
  return date.toLocaleString()
}

const formatJSON = (jsonString) => {
  try {
    const obj = typeof jsonString === 'string' ? JSON.parse(jsonString) : jsonString
    return JSON.stringify(obj, null, 2)
  } catch {
    return jsonString
  }
}

const isRestrictedTool = (toolName) => {
  return restrictedTools.includes(toolName)
}

const loadAuditLog = async () => {
  loading.value = true
  try {
    const filterParams = {
      limit: filters.value.limit,
      offset: filters.value.offset
    }
    if (filters.value.user_id) filterParams.user_id = filters.value.user_id
    if (filters.value.tool) filterParams.tool = filters.value.tool
    if (filters.value.success !== '') filterParams.success = filters.value.success === 'true'

    const data = await toolsService.fetchAuditLog(filterParams)
    auditLogs.value = data.logs || []
  } catch (err) {
    console.error(err)
    toast.error('Failed to load audit log: ' + (err.response?.data?.error || err.message))
  } finally {
    loading.value = false
  }
}

const refreshLog = () => {
  filters.value.offset = 0
  loadAuditLog()
}

const loadMore = () => {
  filters.value.offset += filters.value.limit
  loadAuditLog()
}

const showDetails = (log) => {
  selectedLog.value = log
}

const loadUsers = async () => {
  if (authStore.user?.role !== 'admin') return
  try {
    users.value = await usersService.listUsers()
  } catch (err) {
    console.error('Failed to load users:', err)
  }
}

onMounted(async () => {
  await loadUsers()
  await loadAuditLog()
})
</script>


