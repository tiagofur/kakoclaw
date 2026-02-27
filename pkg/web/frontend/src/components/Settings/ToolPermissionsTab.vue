<template>
  <div class="space-y-5 max-w-6xl mx-auto animate-fade-in-up">
    <!-- Role Permissions Matrix -->
    <div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
      <div class="absolute -top-12 -right-12 w-48 h-48 bg-makoclaw-accent/10 blur-[60px] rounded-full opacity-50 group-hover:opacity-100 transition-opacity"></div>
      
      <div class="relative z-10">
        <div class="flex flex-col sm:flex-row items-center justify-between gap-4 mb-5">
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-xl bg-gradient-to-br from-makoclaw-accent to-indigo-600 shadow-md shadow-makoclaw-accent/20">
              <IconShield class="w-4 h-4 text-white" />
            </div>
            <div>
              <h3 class="text-[11px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">Permissions</h3>
              <p class="text-xs font-medium text-makoclaw-accent uppercase tracking-wide mt-0.5">Role-Based Access Control</p>
            </div>
          </div>
          <button 
            @click="savePermissions"
            :disabled="saving || !hasChanges"
            class="w-full sm:w-auto px-5 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl font-semibold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 active:scale-95 disabled:opacity-30 group/save"
          >
            <span v-if="saving" class="w-4 h-4 border-3 border-white/20 border-t-white rounded-full animate-spin"></span>
            <IconSave v-else class="w-4 h-4 group-hover/save:scale-110 transition-transform" />
            Commit Changes
          </button>
        </div>

        <div v-if="loading" class="flex flex-col items-center justify-center py-24 gap-4">
          <div class="w-12 h-12 border-4 border-makoclaw-accent/20 border-t-makoclaw-accent rounded-full animate-spin"></div>
          <p class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/50">Recalculating Permissions...</p>
        </div>

        <div v-else class="space-y-6">
          <!-- Info Banner -->
          <div class="bg-blue-500/5 border border-blue-500/10 rounded-xl p-4 flex items-start gap-3">
            <div class="p-2 rounded-xl bg-blue-500/10 text-blue-400 mt-1"><IconInfo class="w-5 h-5" /></div>
            <div>
              <p class="text-[11px] font-medium tracking-wide text-makoclaw-text mb-2">Protocol Architecture</p>
              <ul class="space-y-2">
                <li class="flex items-center gap-2 text-xs font-medium text-makoclaw-text-secondary/70">
                   <span class="w-1 h-1 rounded-full bg-makoclaw-accent"></span>
                   Use <code class="px-2 py-0.5 bg-makoclaw-bg/60 rounded-lg text-makoclaw-accent font-black select-all">*</code> as a master wildcard.
                </li>
                <li class="flex items-center gap-2 text-xs font-medium text-makoclaw-text-secondary/70">
                   <span class="w-1 h-1 rounded-full bg-makoclaw-accent"></span>
                   <code class="px-2 py-0.5 bg-makoclaw-bg/60 rounded-lg text-amber-500 font-black">exec_restricted</code> enables safe-shell paradigms only.
                </li>
              </ul>
            </div>
          </div>

          <!-- Permissions Table Overhaul -->
          <div class="overflow-hidden rounded-xl border border-makoclaw-border/50 bg-makoclaw-bg/20">
            <table class="w-full text-left text-sm border-collapse">
              <thead>
                <tr class="bg-makoclaw-bg/40 border-b border-makoclaw-border/50">
                  <th scope="col" class="px-5 py-3 text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/50">Intelligence Socket</th>
                  <th scope="col" class="px-4 py-3 text-center text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/50">Admins</th>
                  <th scope="col" class="px-4 py-3 text-center text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/50">Agents</th>
                  <th scope="col" class="px-5 py-3 text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/50 hidden md:table-cell">Function Definition</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-makoclaw-border/30 text-makoclaw-text">
                <!-- Master Wildcard -->
                <tr class="bg-makoclaw-accent/[0.03] group/row">
                  <td class="px-5 py-4 font-semibold text-makoclaw-accent">
                    <span class="flex items-center gap-3">
                       <span class="w-2 h-2 rounded-full bg-makoclaw-accent animate-pulse"></span>
                       Master Override (*)
                    </span>
                  </td>
                  <td class="px-6 py-4 text-center">
                    <label class="inline-flex items-center cursor-pointer">
                      <input type="checkbox" :checked="rolePermissions.admin?.includes('*')" @change="toggleWildcard('admin', $event.target.checked)" class="sr-only peer">
                      <div class="w-10 h-5 bg-makoclaw-bg/80 rounded-full peer peer-checked:bg-makoclaw-accent after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:after:translate-x-5 relative transition-all"></div>
                    </label>
                  </td>
                  <td class="px-6 py-4 text-center">
                    <label class="inline-flex items-center cursor-pointer">
                      <input type="checkbox" :checked="rolePermissions.user?.includes('*')" @change="toggleWildcard('user', $event.target.checked)" class="sr-only peer">
                      <div class="w-10 h-5 bg-makoclaw-bg/80 rounded-full peer peer-checked:bg-makoclaw-accent after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:after:translate-x-5 relative transition-all"></div>
                    </label>
                  </td>
                  <td class="px-5 py-4 text-[10px] font-medium text-makoclaw-text-secondary/40 uppercase tracking-wide hidden md:table-cell">Total Neural Integration</td>
                </tr>

                <!-- Restricted Exec -->
                <tr class="bg-amber-500/[0.03] group/row">
                  <td class="px-5 py-4 font-semibold text-amber-500">
                    <span class="flex items-center gap-3">
                       <IconLock class="w-4 h-4 opacity-40" />
                       Exec Restricted
                    </span>
                  </td>
                  <td class="px-6 py-4 text-center">
                    <input type="checkbox" :checked="rolePermissions.admin?.includes('exec_restricted')" @change="toggleTool('admin', 'exec_restricted', $event.target.checked)" :disabled="rolePermissions.admin?.includes('*')" class="w-5 h-5 rounded-lg border-2 border-makoclaw-border bg-makoclaw-bg text-makoclaw-accent focus:ring-makoclaw-accent disabled:opacity-20 cursor-pointer">
                  </td>
                  <td class="px-6 py-4 text-center">
                    <input type="checkbox" :checked="rolePermissions.user?.includes('exec_restricted')" @change="toggleTool('user', 'exec_restricted', $event.target.checked)" :disabled="rolePermissions.user?.includes('*')" class="w-5 h-5 rounded-lg border-2 border-makoclaw-border bg-makoclaw-bg text-makoclaw-accent focus:ring-makoclaw-accent disabled:opacity-20 cursor-pointer">
                  </td>
                  <td class="px-5 py-4 text-[10px] font-medium text-makoclaw-text-secondary/40 uppercase tracking-wide hidden md:table-cell">Sandboxed OS Interaction</td>
                </tr>

                <!-- Tools -->
                <tr v-for="tool in availableTools" :key="tool.name" class="hover:bg-white/[0.02] transition-colors group/tool">
                  <td class="px-8 py-5">
                    <div class="flex items-center gap-3">
                      <div :class="`w-2 h-2 rounded-full ${isRestrictedTool(tool.name) ? 'bg-orange-400' : 'bg-makoclaw-text-secondary/20 group-hover/tool:bg-makoclaw-accent'} transition-colors`"></div>
                      <span class="font-bold tracking-tight text-makoclaw-text group-hover/tool:text-makoclaw-accent transition-colors">{{ tool.name }}</span>
                      <span v-if="isRestrictedTool(tool.name)" class="text-[8px] font-medium uppercase text-orange-400 bg-orange-400/10 px-1.5 py-0.5 rounded">Audited</span>
                    </div>
                  </td>
                  <td class="px-6 py-5 text-center">
                    <input type="checkbox" :checked="hasPermission('admin', tool.name)" @change="toggleTool('admin', tool.name, $event.target.checked)" :disabled="rolePermissions.admin?.includes('*')" class="w-5 h-5 rounded-lg border-2 border-makoclaw-border bg-makoclaw-bg text-makoclaw-accent focus:ring-makoclaw-accent disabled:opacity-20 cursor-pointer transition-all">
                  </td>
                  <td class="px-6 py-5 text-center">
                    <input type="checkbox" :checked="hasPermission('user', tool.name)" @change="toggleTool('user', tool.name, $event.target.checked)" :disabled="rolePermissions.user?.includes('*')" class="w-5 h-5 rounded-lg border-2 border-makoclaw-border bg-makoclaw-bg text-makoclaw-accent focus:ring-makoclaw-accent disabled:opacity-20 cursor-pointer transition-all">
                  </td>
                  <td class="px-8 py-5 text-[11px] font-medium text-makoclaw-text-secondary/60 italic hidden md:table-cell group-hover/tool:text-makoclaw-text-secondary transition-colors">{{ tool.description }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Safe Shell Commands Matrix -->
          <div class="glass-panel p-5 rounded-2xl bg-makoclaw-surface/30 border border-makoclaw-border/50">
             <div class="flex items-center gap-3 mb-5">
                <IconTerminal class="w-5 h-5 text-makoclaw-text-secondary/50" />
                <h4 class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/70">Restricted Paradigm Set</h4>
             </div>
             <div class="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-6 gap-3">
                <div v-for="cmd in safeCommands" :key="cmd" class="px-4 py-2.5 bg-makoclaw-bg/60 border border-makoclaw-border/50 rounded-xl text-[10px] font-black font-mono text-makoclaw-text-secondary/80 text-center hover:border-makoclaw-accent/30 hover:text-makoclaw-accent transition-all cursor-default">
                  {{ cmd }}
                </div>
             </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import toolsService from '../../services/toolsService'
import { useToast } from '../../composables/useToast'

// Premium Icons
const IconShield = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z' })]) }
const IconSave = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M13 10V3L4 14h7v7l9-11h-7z' })]) }
const IconInfo = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z' })]) }
const IconLock = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 00-2 2zm10-10V7a4 4 0 00-8 0v4h8z' })]) }
const IconTerminal = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 00-2 2z' })]) }

const toast = useToast()
const loading = ref(true)
const saving = ref(false)
const rolePermissions = ref({ admin: [], user: [] })
const originalPermissions = ref(null)

const availableTools = ref([
  { name: 'read_file', description: 'Read file contents from workspace' },
  { name: 'write_file', description: 'Create or overwrite files' },
  { name: 'edit_file', description: 'Line-based file editing' },
  { name: 'append_file', description: 'Append content to files' },
  { name: 'list_dir', description: 'List directory contents' },
  { name: 'exec', description: 'Execute shell commands (full access if granted)' },
  { name: 'web_search', description: 'Search the web via Brave API' },
  { name: 'web_fetch', description: 'Fetch and parse web pages' },
  { name: 'message', description: 'Send messages to other channels/users' },
  { name: 'email', description: 'Send emails via SMTP' },
  { name: 'spawn', description: 'Spawn subagent for delegated tasks' },
  { name: 'memory', description: 'Store/retrieve long-term context' },
  { name: 'query_knowledge', description: 'Semantic search in knowledge base' },
  { name: 'task_manager', description: 'Create/update/query Kanban tasks' },
  { name: 'schedule', description: 'Schedule cron jobs' }
])

const safeCommands = ['ls', 'cat', 'head', 'tail', 'grep', 'find', 'pwd', 'echo', 'date', 'whoami', 'which', 'wc', 'sort', 'uniq', 'diff', 'tree', 'file', 'stat', 'du', 'df', 'env', 'printenv', 'type']
const restrictedTools = ['exec', 'spawn', 'email', 'write_file', 'edit_file', 'append_file', 'web_fetch']

const hasChanges = computed(() => originalPermissions.value && JSON.stringify(rolePermissions.value) !== JSON.stringify(originalPermissions.value))
const hasPermission = (r, t) => rolePermissions.value[r]?.includes('*') || rolePermissions.value[r]?.includes(t)
const isRestrictedTool = (t) => restrictedTools.includes(t)

const toggleWildcard = (r, c) => { rolePermissions.value[r] = c ? ['*'] : [] }
const toggleTool = (r, t, c) => {
  if (!rolePermissions.value[r]) rolePermissions.value[r] = []
  const w = rolePermissions.value[r].indexOf('*'); if (w !== -1) rolePermissions.value[r].splice(w, 1)
  const idx = rolePermissions.value[r].indexOf(t)
  if (c && idx === -1) rolePermissions.value[r].push(t)
  else if (!c && idx !== -1) rolePermissions.value[r].splice(idx, 1)
}

const loadPermissions = async () => {
  loading.value = true
  try {
    const data = await toolsService.fetchToolPermissions()
    rolePermissions.value = data.role_defaults || { admin: ['*'], user: [] }
    originalPermissions.value = JSON.parse(JSON.stringify(rolePermissions.value))
  } catch (err) { toast.error('Failed to sync permissions: ' + (err.response?.data?.error || err.message)) } finally { loading.value = false }
}

const savePermissions = async () => {
  saving.value = true
  try {
    await toolsService.updateToolPermissions({ role_defaults: rolePermissions.value })
    originalPermissions.value = JSON.parse(JSON.stringify(rolePermissions.value))
    toast.success('Matrix Permissions Committed')
  } catch (err) { toast.error('Failed to commit permissions: ' + (err.response?.data?.error || err.message)) } finally { saving.value = false }
}

onMounted(() => { loadPermissions() })
</script>

<style scoped>
@keyframes fade-in-up {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
.animate-fade-in-up { animation: fade-in-up 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
</style>
