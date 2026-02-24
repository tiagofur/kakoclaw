<template>
  <div class="space-y-6 max-w-6xl mx-auto animate-fadeIn">
    <!-- Role Permissions Matrix -->
    <div class="glass-panel rounded-2xl p-6 md:p-8">
      <div class="flex items-center justify-between mb-6">
        <div>
          <h3 class="text-sm font-bold uppercase tracking-widest text-makoclaw-text-secondary opacity-60 flex items-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-makoclaw-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
            Tool Permissions by Role
          </h3>
          <p class="text-xs text-makoclaw-text-secondary mt-1">Configure which tools each role can access</p>
        </div>
        <button 
          @click="savePermissions"
          :disabled="saving || !hasChanges"
          class="px-4 py-2 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl transition-all shadow-lg shadow-makoclaw-accent/20 text-sm font-bold flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed active:scale-95"
        >
          <span v-if="saving" class="w-3 h-3 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
          {{ saving ? 'Saving...' : 'Save Changes' }}
        </button>
      </div>

      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-makoclaw-accent"></div>
      </div>

      <div v-else class="space-y-6">
        <!-- Info Banner -->
        <div class="bg-blue-500/10 border border-blue-500/20 rounded-xl p-4 flex items-start gap-3">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-blue-400 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <div class="text-sm text-makoclaw-text">
            <p class="font-semibold mb-1">Permission System:</p>
            <ul class="list-disc list-inside space-y-0.5 text-makoclaw-text-secondary">
              <li>Use <code class="px-1.5 py-0.5 bg-makoclaw-bg/60 rounded text-xs">*</code> wildcard to grant all tools</li>
              <li>Use <code class="px-1.5 py-0.5 bg-makoclaw-bg/60 rounded text-xs">exec_restricted</code> token to allow safe shell commands only</li>
              <li>User-specific overrides can be configured per user below</li>
            </ul>
          </div>
        </div>

        <!-- Permissions Table -->
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="uppercase tracking-wider border-b border-makoclaw-border text-[10px] text-makoclaw-text-secondary font-bold">
              <tr>
                <th scope="col" class="px-4 py-3 sticky left-0 bg-makoclaw-surface z-10">Tool</th>
                <th scope="col" class="px-4 py-3 text-center">Admin</th>
                <th scope="col" class="px-4 py-3 text-center">User</th>
                <th scope="col" class="px-4 py-3">Description</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-makoclaw-border text-makoclaw-text">
              <!-- Wildcard Row -->
              <tr class="bg-makoclaw-accent/5">
                <td class="px-4 py-3 font-mono text-sm font-bold sticky left-0 bg-makoclaw-surface z-10">
                  <code class="px-2 py-1 bg-makoclaw-bg/60 rounded">*</code>
                </td>
                <td class="px-4 py-3 text-center">
                  <input 
                    type="checkbox" 
                    :checked="rolePermissions.admin && rolePermissions.admin.includes('*')"
                    @change="toggleWildcard('admin', $event.target.checked)"
                    class="rounded border-makoclaw-border text-makoclaw-accent focus:ring-makoclaw-accent cursor-pointer"
                  >
                </td>
                <td class="px-4 py-3 text-center">
                  <input 
                    type="checkbox"
                    :checked="rolePermissions.user && rolePermissions.user.includes('*')"
                    @change="toggleWildcard('user', $event.target.checked)"
                    class="rounded border-makoclaw-border text-makoclaw-accent focus:ring-makoclaw-accent cursor-pointer"
                  >
                </td>
                <td class="px-4 py-3 text-makoclaw-text-secondary italic">Grant access to all tools</td>
              </tr>

              <!-- Special Token Row -->
              <tr class="bg-yellow-500/5">
                <td class="px-4 py-3 font-mono text-sm font-bold sticky left-0 bg-makoclaw-surface z-10">
                  <code class="px-2 py-1 bg-makoclaw-bg/60 rounded text-yellow-400">exec_restricted</code>
                </td>
                <td class="px-4 py-3 text-center">
                  <input 
                    type="checkbox"
                    :checked="rolePermissions.admin && rolePermissions.admin.includes('exec_restricted')"
                    @change="toggleTool('admin', 'exec_restricted', $event.target.checked)"
                    :disabled="rolePermissions.admin && rolePermissions.admin.includes('*')"
                    class="rounded border-makoclaw-border text-makoclaw-accent focus:ring-makoclaw-accent cursor-pointer disabled:opacity-30"
                  >
                </td>
                <td class="px-4 py-3 text-center">
                  <input 
                    type="checkbox"
                    :checked="rolePermissions.user && rolePermissions.user.includes('exec_restricted')"
                    @change="toggleTool('user', 'exec_restricted', $event.target.checked)"
                    :disabled="rolePermissions.user && rolePermissions.user.includes('*')"
                    class="rounded border-makoclaw-border text-makoclaw-accent focus:ring-makoclaw-accent cursor-pointer disabled:opacity-30"
                  >
                </td>
                <td class="px-4 py-3 text-makoclaw-text-secondary">
                  Shell with safe commands only <span class="text-xs">(ls, cat, grep, pwd, etc.)</span>
                </td>
              </tr>

              <!-- Tool Rows -->
              <tr v-for="tool in availableTools" :key="tool.name" class="hover:bg-makoclaw-bg/30 transition-colors">
                <td class="px-4 py-3 font-mono text-sm sticky left-0 bg-makoclaw-surface z-10">
                  <div class="flex items-center gap-2">
                    <span :class="isRestrictedTool(tool.name) ? 'text-orange-400' : ''">{{ tool.name }}</span>
                    <span v-if="isRestrictedTool(tool.name)" class="text-[10px] px-1.5 py-0.5 bg-orange-500/10 text-orange-400 rounded-full font-bold uppercase" title="Audited tool">🔒</span>
                  </div>
                </td>
                <td class="px-4 py-3 text-center">
                  <input 
                    type="checkbox"
                    :checked="hasPermission('admin', tool.name)"
                    @change="toggleTool('admin', tool.name, $event.target.checked)"
                    :disabled="rolePermissions.admin && rolePermissions.admin.includes('*')"
                    class="rounded border-makoclaw-border text-makoclaw-accent focus:ring-makoclaw-accent cursor-pointer disabled:opacity-30"
                  >
                </td>
                <td class="px-4 py-3 text-center">
                  <input 
                    type="checkbox"
                    :checked="hasPermission('user', tool.name)"
                    @change="toggleTool('user', tool.name, $event.target.checked)"
                    :disabled="rolePermissions.user && rolePermissions.user.includes('*')"
                    class="rounded border-makoclaw-border text-makoclaw-accent focus:ring-makoclaw-accent cursor-pointer disabled:opacity-30"
                  >
                </td>
                <td class="px-4 py-3 text-makoclaw-text-secondary text-xs">{{ tool.description }}</td>
              </tr>

              <tr v-if="availableTools.length === 0">
                <td colspan="4" class="px-4 py-8 text-center text-makoclaw-text-secondary">No tools found</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Shell Commands Reference -->
        <details class="mt-4 bg-makoclaw-bg/30 rounded-xl overflow-hidden">
          <summary class="px-4 py-3 cursor-pointer text-sm font-medium text-makoclaw-text hover:bg-makoclaw-bg/50 transition-colors">
            📋 Safe Shell Commands Reference (23 commands)
          </summary>
          <div class="p-4 border-t border-makoclaw-border">
            <div class="grid grid-cols-2 md:grid-cols-4 gap-2 font-mono text-xs text-makoclaw-text-secondary">
              <code v-for="cmd in safeCommands" :key="cmd" class="px-2 py-1 bg-makoclaw-bg/60 rounded">{{ cmd }}</code>
            </div>
          </div>
        </details>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import toolsService from '../../services/toolsService'
import { useToast } from '../../composables/useToast'

const toast = useToast()
const loading = ref(true)
const saving = ref(false)
const rolePermissions = ref({
  admin: [],
  user: []
})
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

const safeCommands = [
  'ls', 'cat', 'head', 'tail', 'grep', 'find', 'pwd', 'echo', 'date', 'whoami',
  'which', 'wc', 'sort', 'uniq', 'diff', 'tree', 'file', 'stat', 'du', 'df',
  'env', 'printenv', 'type'
]

const restrictedTools = ['exec', 'spawn', 'email', 'write_file', 'edit_file', 'append_file', 'web_fetch']

const hasChanges = computed(() => {
  if (!originalPermissions.value) return false
  return JSON.stringify(rolePermissions.value) !== JSON.stringify(originalPermissions.value)
})

const hasPermission = (role, toolName) => {
  if (!rolePermissions.value[role]) return false
  if (rolePermissions.value[role].includes('*')) return true
  return rolePermissions.value[role].includes(toolName)
}

const isRestrictedTool = (toolName) => {
  return restrictedTools.includes(toolName)
}

const toggleWildcard = (role, checked) => {
  if (checked) {
    rolePermissions.value[role] = ['*']
  } else {
    rolePermissions.value[role] = []
  }
}

const toggleTool = (role, toolName, checked) => {
  if (!rolePermissions.value[role]) {
    rolePermissions.value[role] = []
  }

  // Remove wildcard if present
  const wildcardIndex = rolePermissions.value[role].indexOf('*')
  if (wildcardIndex !== -1) {
    rolePermissions.value[role].splice(wildcardIndex, 1)
  }

  const index = rolePermissions.value[role].indexOf(toolName)
  if (checked && index === -1) {
    rolePermissions.value[role].push(toolName)
  } else if (!checked && index !== -1) {
    rolePermissions.value[role].splice(index, 1)
  }
}

const loadPermissions = async () => {
  loading.value = true
  try {
    const data = await toolsService.fetchToolPermissions()
    rolePermissions.value = data.role_defaults || { admin: ['*'], user: [] }
    originalPermissions.value = JSON.parse(JSON.stringify(rolePermissions.value))
  } catch (err) {
    console.error(err)
    toast.error('Failed to load permissions: ' + (err.response?.data?.error || err.message))
  } finally {
    loading.value = false
  }
}

const savePermissions = async () => {
  saving.value = true
  try {
    await toolsService.updateToolPermissions({ role_defaults: rolePermissions.value })
    originalPermissions.value = JSON.parse(JSON.stringify(rolePermissions.value))
    toast.success('✅ Permissions saved successfully')
  } catch (err) {
    console.error(err)
    toast.error('Failed to save permissions: ' + (err.response?.data?.error || err.message))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadPermissions()
})
</script>
