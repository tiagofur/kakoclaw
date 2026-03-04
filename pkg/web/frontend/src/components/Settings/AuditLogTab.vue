<template>
  <div class="space-y-5 max-w-7xl mx-auto animate-fade-in-up">
    <!-- Filters -->
    <div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50 relative overflow-hidden group">
      <div class="absolute -top-12 -right-12 w-48 h-48 bg-makoclaw-accent/5 blur-[50px] rounded-full group-hover:bg-makoclaw-accent/10 transition-all duration-1000" />
      
      <div class="relative z-10">
        <h3 class="text-[10px] font-bold uppercase tracking-[0.15em] text-makoclaw-accent leading-none mb-5 flex items-center gap-2">
          <AdjustmentsHorizontalIcon class="w-4 h-4" />
          Neural Trace Filters
        </h3>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div class="space-y-2">
            <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">
              Active Subject
            </label>
            <select
              v-model="filters.user_id"
              class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer"
            >
              <option value="">
                All Identities
              </option>
              <option
                v-for="user in users"
                :key="user.id"
                :value="user.id"
              >
                {{ user.username }}
              </option>
            </select>
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">
              Socket Interface
            </label>
            <select
              v-model="filters.tool"
              class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer"
            >
              <option value="">
                All Tools
              </option>
              <option
                v-for="tool in toolsList"
                :key="tool"
                :value="tool"
              >
                {{ tool.toUpperCase() }}
              </option>
            </select>
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">
              Result Status
            </label>
            <select
              v-model="filters.success"
              class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer"
            >
              <option value="">
                All Cycles
              </option>
              <option value="true">
                Success Verified
              </option>
              <option value="false">
                Failed / Restricted
              </option>
            </select>
          </div>

          <div class="flex items-end">
            <button
              class="w-full px-5 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl text-xs font-black uppercase tracking-widest shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 active:scale-95"
              @click="loadAuditLog"
            >
              <MagnifyingGlassIcon class="w-4 h-4" />
              Apply Trace
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Audit Log Table -->
    <div class="glass-panel rounded-2xl border border-makoclaw-border/50 overflow-hidden relative">
      <div class="p-5 pb-3 flex items-center justify-between">
        <div>
          <h3 class="text-base font-black text-makoclaw-text flex items-center gap-2 uppercase tracking-tight italic">
            <DocumentMagnifyingGlassIcon class="w-5 h-5 text-makoclaw-accent" />
            Core Audit Trail
            <span
              v-if="auditLogs.length > 0"
              class="ml-2 px-3 py-1 text-[10px] font-black bg-makoclaw-accent/10 text-makoclaw-accent rounded-xl border border-makoclaw-accent/20 shadow-sm"
            >
              {{ auditLogs.length }} TRACES
            </span>
          </h3>
          <p class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/40 mt-1 uppercase">
            High-fidelity mission activity logs
          </p>
        </div>
        <div class="flex items-center gap-2">
          <button
            class="px-3 py-2 rounded-xl bg-makoclaw-surface border border-makoclaw-border/10 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all active:scale-95 text-[10px] font-bold uppercase tracking-wider flex items-center gap-1.5"
            @click="exportAudit('csv')"
          >
            <ArrowDownTrayIcon class="w-4 h-4" />
            CSV
          </button>
          <button
            class="px-3 py-2 rounded-xl bg-makoclaw-surface border border-makoclaw-border/10 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all active:scale-95 text-[10px] font-bold uppercase tracking-wider flex items-center gap-1.5"
            @click="exportAudit('json')"
          >
            <ArrowDownTrayIcon class="w-4 h-4" />
            JSON
          </button>
          <button
            class="p-3 rounded-2xl bg-makoclaw-surface border border-makoclaw-border/10 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all hover:scale-110 active:scale-90 disabled:opacity-30 shadow-lg"
            :disabled="loading"
            @click="refreshLog"
          >
            <ArrowPathIcon
              class="w-5 h-5"
              :class="{ 'animate-spin': loading }"
            />
          </button>
        </div>
      </div>

      <div
        v-if="loading && auditLogs.length === 0"
        class="flex flex-col items-center justify-center py-32 gap-6"
      >
        <div class="w-16 h-16 border-4 border-makoclaw-accent/20 border-t-makoclaw-accent rounded-full animate-spin" />
        <p class="text-xs font-medium tracking-wide text-makoclaw-text-secondary animate-pulse">
          Syncing Mission History...
        </p>
      </div>

      <div
        v-else-if="auditLogs.length === 0"
        class="flex flex-col items-center justify-center py-32 bg-makoclaw-bg/20"
      >
        <ServerIcon class="w-20 h-20 text-makoclaw-text-secondary/10 mb-6" />
        <p class="text-xs font-medium tracking-wide text-makoclaw-text-secondary/40 uppercase">
          No activity logs recorded
        </p>
        <p class="text-[10px] font-black text-makoclaw-accent/30 mt-2 uppercase tracking-[0.2em]">
          The matrix is quiet
        </p>
      </div>

      <div
        v-else
        class="overflow-x-auto custom-scrollbar"
      >
        <table class="w-full text-left text-sm whitespace-nowrap">
          <thead class="bg-makoclaw-bg/40 text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60">
            <tr>
              <th
                scope="col"
                class="px-8 py-5"
              >
                Temporal Stamp
              </th>
              <th
                scope="col"
                class="px-6 py-5"
              >
                Subject
              </th>
              <th
                scope="col"
                class="px-6 py-5 text-center"
              >
                Neural Socket
              </th>
              <th
                scope="col"
                class="px-6 py-5 text-center"
              >
                Status
              </th>
              <th
                scope="col"
                class="px-6 py-5 text-center"
              >
                Delta
              </th>
              <th
                scope="col"
                class="px-8 py-5 text-right"
              >
                Data
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-makoclaw-border/30 text-makoclaw-text">
            <tr
              v-for="log in auditLogs"
              :key="log.id"
              class="hover:bg-makoclaw-accent/[0.03] transition-colors group"
            >
              <td class="px-8 py-5 font-mono text-[11px] text-makoclaw-text-secondary/60 italic">
                {{ formatTimestamp(log.timestamp) }}
              </td>
              <td class="px-6 py-5">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-makoclaw-surface border border-makoclaw-border flex items-center justify-center text-[10px] font-black text-makoclaw-text-secondary group-hover:text-makoclaw-accent transition-colors">
                    {{ log.username[0].toUpperCase() }}
                  </div>
                  <div>
                    <span class="font-black tracking-tight text-makoclaw-text">{{ log.username }}</span>
                  </div>
                </div>
              </td>
              <td class="px-6 py-5 text-center">
                <span
                  class="font-black font-mono text-xs px-2 py-1 bg-makoclaw-bg/60 rounded-lg border border-makoclaw-border/50 text-makoclaw-text-secondary"
                  :class="isRestrictedTool(log.tool) ? '!text-orange-400 !border-orange-500/20 bg-orange-500/5' : ''"
                >
                  {{ log.tool.toUpperCase() }}
                </span>
              </td>
              <td class="px-6 py-5 text-center">
                <span
                  v-if="log.success"
                  class="inline-flex items-center gap-1.5 px-2 py-0.5 text-[9px] font-medium uppercase rounded-full bg-makoclaw-success/10 text-makoclaw-success border border-makoclaw-success/20"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-makoclaw-success" />
                  Verified
                </span>
                <span
                  v-else
                  class="inline-flex items-center gap-1.5 px-2 py-0.5 text-[9px] font-medium uppercase rounded-full bg-red-500/10 text-red-400 border border-red-500/20"
                  :title="log.error"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-red-500" />
                  Failed
                </span>
              </td>
              <td class="px-6 py-5 text-center font-mono text-xs text-makoclaw-text-secondary/40 italic">
                {{ log.duration_ms ? `${log.duration_ms}ms` : '--' }}
              </td>
              <td class="px-8 py-5 text-right">
                <button
                  class="px-4 py-1.5 bg-makoclaw-bg border border-makoclaw-border hover:border-makoclaw-accent/50 text-makoclaw-text-secondary hover:text-makoclaw-accent rounded-xl text-[10px] font-medium tracking-wide transition-all hover:scale-105"
                  @click="showDetails(log)"
                >
                  Inspect
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Load More Section -->
        <div class="flex items-center justify-between p-8 border-t border-makoclaw-border/50 bg-makoclaw-bg/20">
          <div class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/40">
            Rendered Traces: {{ auditLogs.length }}
          </div>
          <button
            class="px-5 py-2.5 bg-makoclaw-bg border border-makoclaw-border/10 text-makoclaw-text-secondary hover:text-makoclaw-accent hover:border-makoclaw-accent/30 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all active:scale-95 disabled:opacity-20 flex items-center gap-2 group/more shadow-lg"
            :disabled="loading || auditLogs.length < filters.limit"
            @click="loadMore"
          >
            <PlusIcon class="w-4 h-4 group-hover/more:rotate-90 transition-transform" />
            Synchronize More
          </button>
        </div>
      </div>
    </div>

    <!-- Details Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="selectedLog"
          class="fixed inset-0 bg-black/80 backdrop-blur-xl flex items-center justify-center z-[100] p-4"
          @click.self="selectedLog = null"
        >
          <div
            class="glass-panel border border-makoclaw-border/50 rounded-2xl shadow-xl max-w-4xl w-full max-h-[85vh] overflow-hidden flex flex-col relative animate-zoom"
            @click.stop
          >
            <div class="absolute -top-12 -left-12 w-48 h-48 bg-makoclaw-accent/10 blur-[60px] rounded-full pointer-events-none" />

            <div class="p-10 pb-6 flex items-center justify-between relative z-10">
              <div>
                <h3 class="text-2xl font-black text-makoclaw-text italic tracking-tight">
                  Trace Inspection
                </h3>
                <p class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/40 mt-1">
                  Detailed packet payload and state
                </p>
              </div>
              <button
                class="p-3 rounded-2xl bg-makoclaw-bg/60 border border-makoclaw-border/10 text-makoclaw-text-secondary hover:text-white transition-all shadow-xl"
                @click="selectedLog = null"
              >
                <XMarkIcon class="w-6 h-6" />
              </button>
            </div>
            
            <div class="px-10 py-6 overflow-y-auto custom-scrollbar space-y-10 relative z-10 flex-1">
              <div class="grid grid-cols-2 lg:grid-cols-4 gap-5">
                <div class="space-y-1">
                  <span class="text-[9px] font-medium tracking-wide text-makoclaw-text-secondary/40">Temporal Mark</span>
                  <div class="text-xs font-black text-makoclaw-text font-mono truncate">
                    {{ formatTimestamp(selectedLog.timestamp) }}
                  </div>
                </div>
                <div class="space-y-1">
                  <span class="text-[9px] font-medium tracking-wide text-makoclaw-text-secondary/40">Active Subject</span>
                  <div class="text-xs font-medium text-makoclaw-accent uppercase truncate">
                    {{ selectedLog.username }}
                  </div>
                </div>
                <div class="space-y-1">
                  <span class="text-[9px] font-medium tracking-wide text-makoclaw-text-secondary/40">Control Socket</span>
                  <div class="text-xs font-medium text-makoclaw-text truncate">
                    {{ selectedLog.tool }}
                  </div>
                </div>
                <div class="space-y-1">
                  <span class="text-[9px] font-medium tracking-wide text-makoclaw-text-secondary/40">Protocol Result</span>
                  <div
                    class="text-xs font-medium uppercase truncate"
                    :class="selectedLog.success ? 'text-makoclaw-success' : 'text-red-400'"
                  >
                    {{ selectedLog.success ? '✓ SUCCESS' : '✗ FAILED' }}
                  </div>
                </div>
              </div>

              <div class="space-y-4">
                <h4 class="text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary/50 flex items-center gap-2">
                  <CommandLineIcon class="w-4 h-4" /> Payload Arguments
                </h4>
                <div class="bg-makoclaw-bg/60 border-2 border-makoclaw-border/10 rounded-3xl p-8 relative group shadow-inner">
                  <pre class="text-xs font-black font-mono text-makoclaw-text-secondary/90 leading-relaxed overflow-x-auto text-wrap selection:bg-makoclaw-accent selection:text-white">{{ formatJSON(selectedLog.arguments) }}</pre>
                  <div class="absolute top-4 right-4 text-[8px] font-black bg-makoclaw-bg px-2 py-0.5 rounded-md border border-makoclaw-border/10 text-makoclaw-accent/40 uppercase tracking-tighter">
                    JSON.PROTO
                  </div>
                </div>
              </div>

              <div
                v-if="selectedLog.error"
                class="space-y-4"
              >
                <h4 class="text-[10px] font-bold uppercase tracking-widest text-red-500/50 flex items-center gap-2">
                  <ExclamationTriangleIcon class="w-4 h-4" /> Interruption report
                </h4>
                <div class="bg-red-500/5 border-2 border-red-500/10 rounded-3xl p-8 shadow-inner">
                  <p class="text-xs font-medium text-red-400 leading-relaxed">
                    {{ selectedLog.error }}
                  </p>
                </div>
              </div>
            </div>

            <div class="p-8 pt-4 border-t border-makoclaw-border/30 flex justify-end relative z-10 bg-makoclaw-bg/10 backdrop-blur-sm">
              <button
                class="px-5 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl text-[11px] font-semibold shadow-lg shadow-makoclaw-accent/30 transition-all active:scale-95"
                @click="selectedLog = null"
              >
                Dismiss inspection
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import {
  AdjustmentsHorizontalIcon,
  MagnifyingGlassIcon,
  DocumentMagnifyingGlassIcon,
  ArrowPathIcon,
  ArrowDownTrayIcon,
  ServerIcon,
  PlusIcon,
  XMarkIcon,
  CommandLineIcon,
  ExclamationTriangleIcon
} from '@heroicons/vue/24/outline'
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

const filters = ref({ user_id: '', tool: '', success: '', limit: 100, offset: 0 })

const formatTimestamp = (ts) => {
  return ts ? new Date(ts).toLocaleString() : 'N/A'
}

const formatJSON = (j) => {
  try {
    return JSON.stringify(typeof j === 'string' ? JSON.parse(j) : j, null, 2)
  } catch {
    return j
  }
}

const isRestrictedTool = (t) => restrictedTools.includes(t)

const loadAuditLog = async () => {
  loading.value = true
  try {
    const p = { limit: filters.value.limit, offset: filters.value.offset }
    if (filters.value.user_id) p.user_id = filters.value.user_id
    if (filters.value.tool) p.tool = filters.value.tool
    if (filters.value.success !== '') p.success = filters.value.success === 'true'
    const data = await toolsService.fetchAuditLog(p)
    auditLogs.value = data.logs || []
  } catch (err) {
    toast.error('Audit sync failed: ' + (err.response?.data?.error || err.message))
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

const showDetails = (l) => {
  selectedLog.value = l
}

const exportAudit = (format) => {
  const params = new URLSearchParams()
  params.set('format', format)
  if (filters.value.user_id) params.set('user_id', filters.value.user_id)
  if (filters.value.tool) params.set('tool', filters.value.tool)

  const token = localStorage.getItem('token')
  const baseUrl = import.meta.env.VITE_API_URL || '/api/v1'
  const url = `${baseUrl}/tools/audit/export?${params.toString()}`

  // Open in new window for download
  const link = document.createElement('a')
  link.href = url
  link.target = '_blank'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

const loadUsers = async () => {
  if (authStore.user?.role !== 'admin') return
  try {
    users.value = await usersService.listUsers()
  } catch (err) {
    console.error('Subject retrieval failed:', err)
  }
}

onMounted(async () => {
  await loadUsers()
  await loadAuditLog()
})
</script>

<style scoped>
@keyframes fade-in-up {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
.animate-fade-in-up { animation: fade-in-up 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
.animate-zoom { animation: zoom 0.3s cubic-bezier(0.16, 1, 0.3, 1); }
@keyframes zoom { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }
.modal-enter-active, .modal-leave-active { transition: opacity 0.3s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.custom-scrollbar::-webkit-scrollbar { width: 5px; height: 5px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(var(--makoclaw-accent-rgb), 0.2); border-radius: 10px; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background: rgba(var(--makoclaw-accent-rgb), 0.4); }
</style>
