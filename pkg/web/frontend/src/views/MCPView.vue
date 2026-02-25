<template>
  <div class="h-full flex flex-col bg-makoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-makoclaw-border bg-makoclaw-surface flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent">MCP Servers</h2>
        <p class="text-sm text-makoclaw-text-secondary mt-1">Manage Model Context Protocol server connections</p>
      </div>
      <div class="flex items-center gap-3">
        <span class="text-sm text-makoclaw-text-secondary">
          {{ connectedCount }}/{{ servers.length }} connected
        </span>
        <button
          @click="loadServers"
          class="px-3 py-1.5 text-sm bg-makoclaw-bg border border-makoclaw-border rounded-lg hover:border-makoclaw-accent/50 transition-colors"
        >Refresh</button>
        <button
          @click="reconnectAll"
          :disabled="reconnectingAll || servers.length === 0"
          class="px-3 py-1.5 text-sm bg-makoclaw-bg border border-makoclaw-border rounded-lg hover:border-makoclaw-accent/50 transition-colors disabled:opacity-50"
        >
          <span v-if="reconnectingAll">Reconnecting...</span>
          <span v-else>Reconnect All</span>
        </button>
        <button
          @click="openAddModal"
          class="flex items-center gap-2 px-4 py-2 bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors text-sm"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
          Add Server
        </button>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6 custom-scrollbar">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-makoclaw-accent"></div>
      </div>

      <template v-else>
        <!-- No servers configured -->
        <div v-if="servers.length === 0" class="text-center py-12 text-makoclaw-text-secondary">
          <svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
          </svg>
          <p class="text-lg">No MCP servers configured</p>
          <p class="text-sm mt-2">Click "Add Server" to configure your first MCP server</p>
          <button
            @click="openAddModal"
            class="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors text-sm"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
            Add Server
          </button>
        </div>

        <!-- Server Cards -->
        <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <div
            v-for="server in servers"
            :key="server.name"
            class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5 hover:border-makoclaw-accent/50 transition-colors"
          >
            <!-- Server Header -->
            <div class="flex items-start justify-between mb-3">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span
                    class="w-2.5 h-2.5 rounded-full flex-shrink-0"
                    :class="server.connected ? 'bg-blue-400' : (server.enabled === false ? 'bg-gray-400' : 'bg-red-400')"
                  ></span>
                  <h3 class="font-semibold truncate">{{ server.name }}</h3>
                </div>
                <p class="text-xs text-makoclaw-text-secondary mt-1 font-mono truncate" :title="getCommandDisplay(server)">
                  {{ getCommandDisplay(server) }}
                </p>
              </div>
              <div class="flex items-center gap-2 ml-2 flex-shrink-0">
                <!-- Enable/Disable Toggle -->
                <button
                  @click="toggleEnabled(server)"
                  :disabled="togglingEnabled === server.name"
                  class="px-2 py-0.5 text-xs rounded-full transition-colors"
                  :class="server.enabled !== false
                    ? 'bg-blue-500/10 text-blue-400 hover:bg-blue-500/20'
                    : 'bg-gray-500/10 text-gray-400 hover:bg-gray-500/20'"
                  :title="server.enabled !== false ? 'Click to disable' : 'Click to enable'"
                >
                  <span v-if="togglingEnabled === server.name">...</span>
                  <span v-else>{{ server.enabled !== false ? 'Enabled' : 'Disabled' }}</span>
                </button>
                <span
                  class="px-2 py-0.5 text-xs rounded-full"
                  :class="server.connected
                    ? 'bg-blue-500/10 text-blue-400'
                    : 'bg-red-500/10 text-red-400'"
                >{{ server.connected ? 'Connected' : 'Disconnected' }}</span>
              </div>
            </div>

            <!-- Server Info -->
            <div v-if="server.connected" class="space-y-2 mb-3">
              <div v-if="server.server_name" class="flex items-center gap-2 text-sm">
                <span class="text-makoclaw-text-secondary">Server:</span>
                <span>{{ server.server_name }} <span v-if="server.server_version" class="text-makoclaw-text-secondary">v{{ server.server_version }}</span></span>
              </div>
              <div class="flex items-center gap-2 text-sm">
                <span class="text-makoclaw-text-secondary">Tools:</span>
                <span>{{ server.tool_count }} available</span>
              </div>
              <!-- Tool list -->
              <div v-if="server.tools && server.tools.length > 0" class="flex flex-wrap gap-1 mt-1">
                <span
                  v-for="tool in server.tools"
                  :key="tool"
                  class="px-2 py-0.5 text-xs bg-makoclaw-bg rounded-full text-makoclaw-text-secondary"
                >{{ tool }}</span>
              </div>
            </div>

            <!-- Command and Args Display (when not connected) -->
            <div v-if="!server.connected && server.args && server.args.length > 0" class="mb-3">
              <div class="text-xs text-makoclaw-text-secondary mb-1">Arguments:</div>
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="(arg, idx) in server.args.slice(0, 5)"
                  :key="idx"
                  class="px-2 py-0.5 text-xs bg-makoclaw-bg rounded font-mono text-makoclaw-text-secondary"
                >{{ arg }}</span>
                <span v-if="server.args.length > 5" class="px-2 py-0.5 text-xs text-makoclaw-text-secondary">
                  +{{ server.args.length - 5 }} more
                </span>
              </div>
            </div>

            <!-- Error Info -->
            <div v-if="server.last_error" class="mb-3 p-2 bg-red-500/5 border border-red-500/20 rounded-lg">
              <p class="text-xs text-red-400 break-all">{{ server.last_error }}</p>
            </div>

            <!-- Test Result -->
            <div v-if="testResults[server.name]" class="mb-3 p-2 rounded-lg border"
              :class="testResults[server.name].success
                ? 'bg-blue-500/5 border-blue-500/20'
                : 'bg-red-500/5 border-red-500/20'"
            >
              <p class="text-xs"
                :class="testResults[server.name].success ? 'text-blue-400' : 'text-red-400'"
              >{{ testResults[server.name].message }}</p>
            </div>

            <!-- Actions -->
            <div class="flex items-center gap-2 mt-3 pt-3 border-t border-makoclaw-border">
              <button
                @click="reconnect(server.name)"
                :disabled="reconnecting === server.name"
                class="px-3 py-1.5 text-xs bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors disabled:opacity-50"
              >
                <span v-if="reconnecting === server.name">Reconnecting...</span>
                <span v-else>{{ server.connected ? 'Reconnect' : 'Connect' }}</span>
              </button>
              <button
                @click="testServer(server.name)"
                :disabled="testing === server.name"
                class="px-3 py-1.5 text-xs text-makoclaw-text-secondary bg-makoclaw-bg border border-makoclaw-border rounded-lg hover:bg-makoclaw-border transition-colors disabled:opacity-50"
              >
                <span v-if="testing === server.name">Testing...</span>
                <span v-else>Test</span>
              </button>
              <button
                @click="openEditModal(server)"
                class="px-3 py-1.5 text-xs text-makoclaw-text-secondary bg-makoclaw-bg border border-makoclaw-border rounded-lg hover:bg-makoclaw-border transition-colors"
              >Edit</button>
              <button
                @click="confirmDeleteServer(server)"
                class="px-3 py-1.5 text-xs text-red-400 bg-red-500/10 rounded-lg hover:bg-red-500/20 transition-colors"
              >Delete</button>
            </div>
          </div>
        </div>

        <!-- Info Box -->
        <div v-if="servers.length > 0" class="mt-6 p-4 bg-makoclaw-surface border border-makoclaw-border rounded-xl">
          <h4 class="text-sm font-semibold mb-2">About MCP</h4>
          <p class="text-xs text-makoclaw-text-secondary leading-relaxed">
            The Model Context Protocol (MCP) allows makoclaw to connect to external tool servers.
            Tools discovered from MCP servers are automatically available to the AI agent during conversations.
            Tool names are prefixed with <code class="px-1 py-0.5 bg-makoclaw-bg rounded">mcp_&lt;server&gt;_</code> to avoid conflicts.
          </p>
        </div>
      </template>
    </div>

    <!-- Add / Edit Server Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="showModal = false">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-lg w-full max-h-[90vh] overflow-y-auto p-6">
        <h3 class="font-semibold text-lg mb-4">{{ editingServer ? 'Edit MCP Server' : 'Add MCP Server' }}</h3>
        <div class="space-y-4">
          <!-- Name -->
          <div>
            <label class="block text-sm font-medium mb-1">Name <span class="text-red-400">*</span></label>
            <input
              v-model="form.name"
              type="text"
              placeholder="my-server"
              :disabled="!!editingServer"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent disabled:opacity-50 disabled:cursor-not-allowed"
            />
            <p class="text-xs text-makoclaw-text-secondary mt-1">Unique identifier for this server (alphanumeric and hyphens)</p>
          </div>

          <!-- Command -->
          <div>
            <label class="block text-sm font-medium mb-1">Command <span class="text-red-400">*</span></label>
            <input
              v-model="form.command"
              type="text"
              placeholder="npx"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent font-mono"
            />
            <p class="text-xs text-makoclaw-text-secondary mt-1">Command to start the MCP server (e.g., npx, node, python)</p>
          </div>

          <!-- Args -->
          <div>
            <label class="block text-sm font-medium mb-1">Arguments</label>
            <textarea
              v-model="form.argsText"
              rows="3"
              placeholder="-y
@modelcontextprotocol/server-filesystem
/path/to/directory"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent resize-none font-mono"
            />
            <p class="text-xs text-makoclaw-text-secondary mt-1">One argument per line</p>
          </div>

          <!-- Env -->
          <div>
            <label class="block text-sm font-medium mb-1">Environment Variables</label>
            <textarea
              v-model="form.envText"
              rows="3"
              placeholder="API_KEY=your-key
DEBUG=true"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent resize-none font-mono"
            />
            <p class="text-xs text-makoclaw-text-secondary mt-1">KEY=value pairs, one per line</p>
          </div>

          <!-- Enabled -->
          <div class="flex items-center gap-2">
            <input v-model="form.enabled" type="checkbox" id="enabled" class="rounded" />
            <label for="enabled" class="text-sm">Enable server</label>
          </div>

          <!-- Test Connection Button -->
          <div v-if="editingServer" class="pt-2">
            <button
              @click="testServerFromModal"
              :disabled="modalTesting"
              class="w-full px-4 py-2 text-sm bg-makoclaw-bg border border-makoclaw-border rounded-lg hover:bg-makoclaw-border transition-colors disabled:opacity-50"
            >
              <span v-if="modalTesting">Testing Connection...</span>
              <span v-else>Test Connection</span>
            </button>
            <div v-if="modalTestResult" class="mt-2 p-2 rounded-lg border"
              :class="modalTestResult.success
                ? 'bg-blue-500/5 border-blue-500/20'
                : 'bg-red-500/5 border-red-500/20'"
            >
              <p class="text-xs"
                :class="modalTestResult.success ? 'text-blue-400' : 'text-red-400'"
              >{{ modalTestResult.message }}</p>
            </div>
          </div>
        </div>

        <!-- Modal Actions -->
        <div class="flex justify-end gap-3 mt-6">
          <button @click="showModal = false"
            class="px-4 py-2 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">Cancel</button>
          <button @click="submitServer" :disabled="!canSubmit || submitting"
            class="px-4 py-2 text-sm bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors disabled:opacity-50">
            <span v-if="submitting">{{ editingServer ? 'Saving...' : 'Adding...' }}</span>
            <span v-else>{{ editingServer ? 'Save Changes' : 'Add Server' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="showDeleteConfirm = false">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-sm w-full p-6">
        <h3 class="font-semibold text-lg mb-2">Delete Server</h3>
        <p class="text-sm text-makoclaw-text-secondary mb-2">
          Are you sure you want to delete <span class="font-medium text-makoclaw-text">{{ deletingServer?.name }}</span>?
        </p>
        <div class="p-3 bg-yellow-500/5 border border-yellow-500/20 rounded-lg mb-4">
          <p class="text-xs text-yellow-400">
            Warning: All tools from this server will become unavailable to the AI agent.
          </p>
        </div>
        <div class="flex justify-end gap-3">
          <button @click="showDeleteConfirm = false"
            class="px-4 py-2 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">Cancel</button>
          <button @click="executeDeleteServer" :disabled="deleting"
            class="px-4 py-2 text-sm bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors disabled:opacity-50">
            <span v-if="deleting">Deleting...</span>
            <span v-else>Delete</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const loading = ref(true)
const reconnecting = ref(null)
const reconnectingAll = ref(false)
const testing = ref(null)
const togglingEnabled = ref(null)
const servers = ref([])
const testResults = ref({})

// Modal state
const showModal = ref(false)
const editingServer = ref(null)
const submitting = ref(false)
const modalTesting = ref(false)
const modalTestResult = ref(null)

// Delete confirmation state
const showDeleteConfirm = ref(false)
const deletingServer = ref(null)
const deleting = ref(false)

// Form state
const defaultForm = () => ({
  name: '',
  command: '',
  argsText: '',
  envText: '',
  enabled: true
})

const form = ref(defaultForm())

const connectedCount = computed(() => servers.value.filter(s => s.connected).length)

const canSubmit = computed(() => {
  return form.value.name.trim() && form.value.command.trim()
})

const getCommandDisplay = (server) => {
  const cmd = server.command || ''
  const args = server.args || []
  if (args.length === 0) return cmd
  return `${cmd} ${args.join(' ')}`
}

// Parse form text fields to arrays/objects
const parseArgs = (text) => {
  return text.split('\n').map(line => line.trim()).filter(line => line)
}

const parseEnv = (text) => {
  const env = {}
  text.split('\n').forEach(line => {
    const trimmed = line.trim()
    if (!trimmed) return
    const eqIndex = trimmed.indexOf('=')
    if (eqIndex > 0) {
      const key = trimmed.substring(0, eqIndex).trim()
      const value = trimmed.substring(eqIndex + 1).trim()
      if (key) env[key] = value
    }
  })
  return env
}

// Convert server data to form
const serverToForm = (server) => ({
  name: server.name,
  command: server.command || '',
  argsText: (server.args || []).join('\n'),
  envText: Object.entries(server.env || {}).map(([k, v]) => `${k}=${v}`).join('\n'),
  enabled: server.enabled !== false
})

// Build API payload from form
const buildPayload = () => {
  const args = parseArgs(form.value.argsText)
  const env = parseEnv(form.value.envText)

  return {
    name: form.value.name.trim(),
    command: form.value.command.trim(),
    args: args.length > 0 ? args : undefined,
    env: Object.keys(env).length > 0 ? env : undefined,
    enabled: form.value.enabled
  }
}

// ---- Actions ----

const loadServers = async () => {
  loading.value = true
  try {
    const data = await advancedService.fetchMCPServers()
    servers.value = data.servers || []
  } catch (err) {
    console.error('Failed to load MCP servers:', err)
    toast.error('Failed to load MCP servers')
  } finally {
    loading.value = false
  }
}

const reconnect = async (name) => {
  reconnecting.value = name
  try {
    const data = await advancedService.reconnectMCPServer(name)
    if (data.ok) {
      toast.success(data.message || `Reconnected to ${name}`)
    } else {
      toast.error(data.error || 'Reconnection failed')
    }
    await loadServers()
  } catch (err) {
    console.error('Reconnect failed:', err)
    toast.error(err.response?.data?.error || 'Reconnection failed')
    await loadServers()
  } finally {
    reconnecting.value = null
  }
}

const reconnectAll = async () => {
  reconnectingAll.value = true
  try {
    const data = await advancedService.reconnectAllMCPServers()
    const reconnected = data.reconnected || []
    const failed = data.failed || []

    if (failed.length === 0 && reconnected.length > 0) {
      toast.success(`Reconnected ${reconnected.length} server(s)`)
    } else if (failed.length > 0 && reconnected.length > 0) {
      toast.warning(`Reconnected ${reconnected.length}, failed ${failed.length}: ${failed.join(', ')}`)
    } else if (failed.length > 0) {
      toast.error(`Failed to reconnect: ${failed.join(', ')}`)
    } else {
      toast.info('No servers to reconnect')
    }
    await loadServers()
  } catch (err) {
    console.error('Reconnect all failed:', err)
    toast.error(err.response?.data?.error || 'Failed to reconnect servers')
    await loadServers()
  } finally {
    reconnectingAll.value = false
  }
}

const testServer = async (name) => {
  testing.value = name
  testResults.value[name] = null
  try {
    const data = await advancedService.testMCPServer(name)
    if (data.ok) {
      testResults.value[name] = {
        success: true,
        message: `Connection successful! ${data.tool_count || 0} tools available.`
      }
      toast.success(`Test successful: ${name}`)
    } else {
      testResults.value[name] = {
        success: false,
        message: data.error || 'Connection test failed'
      }
      toast.error(data.error || 'Connection test failed')
    }
  } catch (err) {
    console.error('Test failed:', err)
    const errorMsg = err.response?.data?.error || err.message || 'Connection test failed'
    testResults.value[name] = {
      success: false,
      message: errorMsg
    }
    toast.error(errorMsg)
  } finally {
    testing.value = null
  }
}

const toggleEnabled = async (server) => {
  const newEnabled = server.enabled === false
  togglingEnabled.value = server.name
  try {
    await advancedService.updateMCPServer(server.name, {
      ...server,
      enabled: newEnabled
    })
    toast.success(`Server ${newEnabled ? 'enabled' : 'disabled'}`)
    await loadServers()
  } catch (err) {
    console.error('Toggle failed:', err)
    toast.error(err.response?.data?.error || 'Failed to toggle server')
  } finally {
    togglingEnabled.value = null
  }
}

// Modal actions
const openAddModal = () => {
  editingServer.value = null
  form.value = defaultForm()
  modalTestResult.value = null
  showModal.value = true
}

const openEditModal = (server) => {
  editingServer.value = server
  form.value = serverToForm(server)
  modalTestResult.value = null
  showModal.value = true
}

const testServerFromModal = async () => {
  if (!editingServer.value) return
  modalTesting.value = true
  modalTestResult.value = null
  try {
    const data = await advancedService.testMCPServer(editingServer.value.name)
    if (data.ok) {
      modalTestResult.value = {
        success: true,
        message: `Connection successful! ${data.tool_count || 0} tools available.`
      }
    } else {
      modalTestResult.value = {
        success: false,
        message: data.error || 'Connection test failed'
      }
    }
  } catch (err) {
    const errorMsg = err.response?.data?.error || err.message || 'Connection test failed'
    modalTestResult.value = {
      success: false,
      message: errorMsg
    }
  } finally {
    modalTesting.value = false
  }
}

const submitServer = async () => {
  if (!canSubmit.value) return
  submitting.value = true
  try {
    const payload = buildPayload()
    if (editingServer.value) {
      await advancedService.updateMCPServer(editingServer.value.name, payload)
      toast.success('Server updated successfully')
    } else {
      await advancedService.addMCPServer(payload)
      toast.success('Server added successfully')
    }
    showModal.value = false
    await loadServers()
  } catch (err) {
    console.error('Submit failed:', err)
    const errorMsg = err.response?.data?.error || err.message || 'Failed to save server'
    toast.error(errorMsg)
  } finally {
    submitting.value = false
  }
}

// Delete actions
const confirmDeleteServer = (server) => {
  deletingServer.value = server
  showDeleteConfirm.value = true
}

const executeDeleteServer = async () => {
  if (!deletingServer.value) return
  deleting.value = true
  try {
    await advancedService.deleteMCPServer(deletingServer.value.name)
    toast.success('Server deleted successfully')
    showDeleteConfirm.value = false
    deletingServer.value = null
    await loadServers()
  } catch (err) {
    console.error('Delete failed:', err)
    toast.error(err.response?.data?.error || 'Failed to delete server')
  } finally {
    deleting.value = false
  }
}

onMounted(() => loadServers())
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 8px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: rgba(156, 163, 175, 0.5); border-radius: 4px; }
</style>
