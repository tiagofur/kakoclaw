<template>
  <div class="space-y-8 max-w-7xl mx-auto animate-fade-in-up">
    <!-- Filters -->
    <div class="glass-panel rounded-[2rem] p-8 border border-makoclaw-border/50 relative overflow-hidden group">
      <div class="absolute -top-12 -right-12 w-48 h-48 bg-makoclaw-accent/5 blur-[50px] rounded-full group-hover:bg-makoclaw-accent/10 transition-all duration-1000"></div>
      
      <div class="relative z-10">
        <h3 class="text-[11px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70 italic mb-8 flex items-center gap-3">
          <IconFilter class="w-4 h-4 text-makoclaw-accent" />
          Neural Trace Filters
        </h3>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <div class="space-y-2">
            <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Active Subject</label>
            <select v-model="filters.user_id" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer">
              <option value="">All Identities</option>
              <option v-for="user in users" :key="user.id" :value="user.id">{{ user.username }}</option>
            </select>
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Socket Interface</label>
            <select v-model="filters.tool" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer">
              <option value="" class="text-makoclaw-text">All Tools</option>
              <option v-for="tool in toolsList" :key="tool" :value="tool">{{ tool.toUpperCase() }}</option>
            </select>
          </div>

          <div class="space-y-2">
            <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Result Status</label>
            <select v-model="filters.success" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer">
              <option value="">All Cycles</option>
              <option value="true" class="text-makoclaw-success">Success Verified</option>
              <option value="false" class="text-red-400">Failed / Restricted</option>
            </select>
          </div>

          <div class="flex items-end">
            <button 
              @click="loadAuditLog"
              class="w-full px-6 py-3.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-2xl font-black uppercase tracking-widest shadow-xl shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 active:scale-95"
            >
              <IconSearch class="w-4 h-4" />
              Apply Trace
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Audit Log Table -->
    <div class="glass-panel rounded-[2.5rem] p-1 border border-makoclaw-border/50 overflow-hidden relative">
      <div class="p-8 pb-4 flex items-center justify-between">
        <div>
          <h3 class="text-xl font-black text-makoclaw-text italic flex items-center gap-3">
             <IconAudit class="w-6 h-6 text-makoclaw-accent" />
             Core Audit Trail
             <span v-if="auditLogs.length > 0" class="ml-2 px-2.5 py-1 text-[10px] font-black bg-makoclaw-accent/10 text-makoclaw-accent rounded-xl border border-makoclaw-accent/20">{{ auditLogs.length }} TRACES</span>
          </h3>
          <p class="text-[10px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/40 mt-1">High-fidelity mission activity logs</p>
        </div>
        <button 
          @click="refreshLog"
          :disabled="loading"
          class="p-3 rounded-2xl bg-makoclaw-surface border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all hover:scale-110 active:scale-90 disabled:opacity-30"
        >
          <IconRefresh class="w-5 h-5" :class="{ 'animate-spin': loading }" />
        </button>
      </div>

      <div v-if="loading && auditLogs.length === 0" class="flex flex-col items-center justify-center py-32 gap-6">
        <div class="w-16 h-16 border-4 border-makoclaw-accent/20 border-t-makoclaw-accent rounded-full animate-spin"></div>
        <p class="text-xs font-black uppercase tracking-widest text-makoclaw-text-secondary animate-pulse">Syncing Mission History...</p>
      </div>

      <div v-else-if="auditLogs.length === 0" class="flex flex-col items-center justify-center py-32 bg-makoclaw-bg/20">
        <IconEmpty class="w-20 h-20 text-makoclaw-text-secondary/10 mb-6" />
        <p class="text-xs font-black uppercase tracking-[0.3em] text-makoclaw-text-secondary/40 italic">No activity logs recorded</p>
        <p class="text-[10px] font-medium text-makoclaw-text-secondary/20 mt-2 uppercase tracking-widest">The matrix is quiet</p>
      </div>

      <div v-else class="overflow-x-auto custom-scrollbar">
        <table class="w-full text-left text-sm whitespace-nowrap">
          <thead class="bg-makoclaw-bg/40 text-[10px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/60">
            <tr>
              <th scope="col" class="px-8 py-5">Temporal Stamp</th>
              <th scope="col" class="px-6 py-5">Subject</th>
              <th scope="col" class="px-6 py-5 text-center">Neural Socket</th>
              <th scope="col" class="px-6 py-5 text-center">Status</th>
              <th scope="col" class="px-6 py-5 text-center">Delta</th>
              <th scope="col" class="px-8 py-5 text-right">Data</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-makoclaw-border/30 text-makoclaw-text">
            <tr v-for="log in auditLogs" :key="log.id" class="hover:bg-makoclaw-accent/[0.03] transition-colors group">
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
                <span class="font-black font-mono text-xs px-2 py-1 bg-makoclaw-bg/60 rounded-lg border border-makoclaw-border/50 text-makoclaw-text-secondary" :class="isRestrictedTool(log.tool) ? '!text-orange-400 !border-orange-500/20 bg-orange-500/5' : ''">
                  {{ log.tool.toUpperCase() }}
                </span>
              </td>
              <td class="px-6 py-5 text-center">
                <span v-if="log.success" class="inline-flex items-center gap-1.5 px-3 py-1 text-[9px] font-black uppercase rounded-full bg-makoclaw-success/10 text-makoclaw-success border border-makoclaw-success/20">
                  <span class="w-1.5 h-1.5 rounded-full bg-makoclaw-success"></span>
                  Verified
                </span>
                <span v-else class="inline-flex items-center gap-1.5 px-3 py-1 text-[9px] font-black uppercase rounded-full bg-red-500/10 text-red-400 border border-red-500/20" :title="log.error">
                  <span class="w-1.5 h-1.5 rounded-full bg-red-500"></span>
                  Failed
                </span>
              </td>
              <td class="px-6 py-5 text-center font-mono text-xs text-makoclaw-text-secondary/40 italic">
                {{ log.duration_ms ? `${log.duration_ms}ms` : '--' }}
              </td>
              <td class="px-8 py-5 text-right">
                <button 
                  @click="showDetails(log)"
                  class="px-4 py-2 bg-makoclaw-bg border border-makoclaw-border hover:border-makoclaw-accent/50 text-makoclaw-text-secondary hover:text-makoclaw-accent rounded-xl text-[10px] font-black uppercase tracking-widest transition-all hover:scale-105"
                >
                  Inspect
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Load More Section -->
        <div class="flex items-center justify-between p-8 border-t border-makoclaw-border/50 bg-makoclaw-bg/20">
          <div class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40">
            Rendered Traces: {{ auditLogs.length }}
          </div>
          <button 
            @click="loadMore"
            :disabled="loading || auditLogs.length < filters.limit"
            class="px-8 py-3 bg-makoclaw-surface border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-white rounded-2xl text-[10px] font-black uppercase tracking-widest transition-all active:scale-95 disabled:opacity-20 flex items-center gap-3 group/more"
          >
            <IconPlus class="w-3.5 h-3.5 group-hover/more:rotate-90 transition-transform" />
            Synchronize More
          </button>
        </div>
      </div>
    </div>

    <!-- Details Modal -->
    <Teleport to="body">
    <Transition name="modal">
    <div v-if="selectedLog" class="fixed inset-0 bg-black/80 backdrop-blur-xl flex items-center justify-center z-[100] p-4" @click.self="selectedLog = null">
      <div class="glass-panel border border-makoclaw-border/50 rounded-[2.5rem] shadow-2xl max-w-4xl w-full max-h-[85vh] overflow-hidden flex flex-col relative animate-zoom" @click.stop>
        <div class="absolute -top-12 -left-12 w-48 h-48 bg-makoclaw-accent/10 blur-[60px] rounded-full pointer-events-none"></div>

        <div class="p-10 pb-6 flex items-center justify-between relative z-10">
          <div>
            <h3 class="text-2xl font-black text-makoclaw-text italic tracking-tight">Trace Inspection</h3>
            <p class="text-[10px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/40 mt-1">Detailed packet payload and state</p>
          </div>
          <button @click="selectedLog = null" class="p-3 rounded-2xl bg-makoclaw-bg/60 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-white transition-all">
            <IconClose class="w-6 h-6" />
          </button>
        </div>
        
        <div class="px-10 py-6 overflow-y-auto custom-scrollbar space-y-10 relative z-10 flex-1">
          <div class="grid grid-cols-2 lg:grid-cols-4 gap-8">
            <div class="space-y-1">
              <span class="text-[9px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40">Temporal Mark</span>
              <div class="text-xs font-black text-makoclaw-text font-mono truncate">{{ formatTimestamp(selectedLog.timestamp) }}</div>
            </div>
            <div class="space-y-1">
              <span class="text-[9px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40">Active Subject</span>
              <div class="text-xs font-black text-makoclaw-accent uppercase tracking-widest truncate">{{ selectedLog.username }}</div>
            </div>
            <div class="space-y-1">
              <span class="text-[9px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40">Control Socket</span>
              <div class="text-xs font-black text-makoclaw-text uppercase tracking-widest truncate italic">{{ selectedLog.tool }}</div>
            </div>
            <div class="space-y-1">
              <span class="text-[9px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40">Protocol Result</span>
              <div class="text-xs font-black uppercase tracking-widest truncate" :class="selectedLog.success ? 'text-makoclaw-success' : 'text-red-400'">
                {{ selectedLog.success ? '✓ SUCCESS' : '✗ FAILED' }}
              </div>
            </div>
          </div>

          <div class="space-y-4">
            <h4 class="text-[10px] font-black uppercase tracking-[0.3em] text-makoclaw-text-secondary/40 flex items-center gap-2">
               <IconTerminal class="w-4 h-4" /> Payload Arguments
            </h4>
            <div class="bg-makoclaw-bg/60 border-2 border-makoclaw-border/30 rounded-3xl p-8 relative group">
               <pre class="text-xs font-black font-mono text-makoclaw-text-secondary/90 leading-relaxed overflow-x-auto text-wrap selection:bg-makoclaw-accent selection:text-white">{{ formatJSON(selectedLog.arguments) }}</pre>
               <div class="absolute top-4 right-4 text-[8px] font-black uppercase bg-makoclaw-bg px-2 py-1 rounded-lg border border-makoclaw-border/50 text-makoclaw-text-secondary/30">JSON.PROTO</div>
            </div>
          </div>

          <div v-if="selectedLog.error" class="space-y-4">
            <h4 class="text-[10px] font-black uppercase tracking-[0.3em] text-red-400/50 flex items-center gap-2">
               <IconAlert class="w-4 h-4" /> Interruption report
            </h4>
            <div class="bg-red-500/5 border-2 border-red-500/20 rounded-3xl p-8">
               <p class="text-xs font-black text-red-400 leading-relaxed italic uppercase tracking-widest">{{ selectedLog.error }}</p>
            </div>
          </div>
        </div>

        <div class="p-8 pt-4 border-t border-makoclaw-border/30 flex justify-end relative z-10 bg-makoclaw-bg/10 backdrop-blur-sm">
          <button @click="selectedLog = null" class="px-10 py-4 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-2xl text-[11px] font-black uppercase tracking-widest shadow-xl shadow-makoclaw-accent/30 transition-all active:scale-95">
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
import { ref, onMounted, h } from 'vue'
import toolsService from '../../services/toolsService'
import usersService from '../../services/usersService'
import { useToast } from '../../composables/useToast'
import { useAuthStore } from '../../stores/authStore'

// Premium Icons
const IconFilter = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z' })]) }
const IconSearch = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z' })]) }
const IconAudit = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' })]) }
const IconRefresh = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15' })]) }
const IconEmpty = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0a2 2 0 01-2 2H6a2 2 0 01-2-2m16 0l-8 4-8-4m8 4v6' })]) }
const IconPlus = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M12 6v6m0 0v6m0-6h6m-6 0H6' })]) }
const IconClose = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M6 18L18 6M6 6l12 12' })]) }
const IconTerminal = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 00-2 2z' })]) }
const IconAlert = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z' })]) }

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

const formatTimestamp = (ts) => ts ? new Date(ts).toLocaleString() : 'N/A'
const formatJSON = (j) => { try { return JSON.stringify(typeof j==='string' ? JSON.parse(j) : j, null, 2) } catch { return j } }
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
  } catch (err) { toast.error('Audit sync failed: ' + (err.response?.data?.error || err.message)) } finally { loading.value = false }
}

const refreshLog = () => { filters.value.offset = 0; loadAuditLog() }
const loadMore = () => { filters.value.offset += filters.value.limit; loadAuditLog() }
const showDetails = (l) => { selectedLog.value = l }

const loadUsers = async () => {
  if (authStore.user?.role !== 'admin') return
  try { users.value = await usersService.listUsers() } catch (err) { console.error('Subject retrieval failed:', err) }
}

onMounted(async () => { await loadUsers(); await loadAuditLog() })
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
