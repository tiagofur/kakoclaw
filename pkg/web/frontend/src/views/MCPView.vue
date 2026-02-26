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
    <div class="flex-1 overflow-auto p-4 md:p-6 custom-scrollbar">
      <!-- Loading Skeleton -->
      <div v-if="loading" class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div v-for="i in 3" :key="i" class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5">
          <div class="flex items-start justify-between mb-3">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-2">
                <div class="skeleton w-2.5 h-2.5 rounded-full"></div>
                <div class="skeleton h-4 w-32 rounded"></div>
              </div>
              <div class="skeleton h-3 w-48 rounded"></div>
            </div>
            <div class="flex items-center gap-2">
              <div class="skeleton h-5 w-16 rounded-full"></div>
              <div class="skeleton h-5 w-20 rounded-full"></div>
            </div>
          </div>
          <div class="space-y-2 mb-3">
            <div class="skeleton h-3 w-28 rounded"></div>
            <div class="flex gap-1">
              <div class="skeleton h-5 w-20 rounded-full"></div>
              <div class="skeleton h-5 w-24 rounded-full"></div>
              <div class="skeleton h-5 w-16 rounded-full"></div>
            </div>
          </div>
          <div class="flex items-center gap-2 mt-3 pt-3 border-t border-makoclaw-border">
            <div class="skeleton h-7 w-20 rounded-lg"></div>
            <div class="skeleton h-7 w-14 rounded-lg"></div>
            <div class="skeleton h-7 w-12 rounded-lg"></div>
          </div>
        </div>
      </div>

      <template v-else>
        <!-- No servers configured -->
        <div v-if="servers.length === 0" class="flex flex-col items-center justify-center py-16 text-center">
          <div class="w-16 h-16 rounded-2xl bg-makoclaw-accent/10 flex items-center justify-center mb-4">
            <svg class="w-8 h-8 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5.25 14.25h13.5m-13.5 0a3 3 0 01-3-3m3 3a3 3 0 100 6h13.5a3 3 0 100-6m-16.5-3a3 3 0 013-3h13.5a3 3 0 013 3m-19.5 0a4.5 4.5 0 01.9-2.7L5.737 5.1a3.375 3.375 0 012.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 01.9 2.7m0 0a3 3 0 01-3 3m0 3h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008zm-3 6h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008z" />
            </svg>
          </div>
          <h3 class="font-semibold text-makoclaw-text mb-1">No MCP servers configured</h3>
          <p class="text-sm text-makoclaw-text-secondary max-w-xs mb-4">Connect external tool servers to extend your agent's capabilities.</p>
          <button class="btn-primary" @click="openAddModal">Add Server</button>
        </div>

        <!-- Server Cards -->
        <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <div
            v-for="server in servers"
            :key="server.name"
            class="card-interactive p-5"
          >
            <!-- Server Header -->
            <div class="flex items-start justify-between mb-3">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span
                    class="w-2.5 h-2.5 rounded-full flex-shrink-0"
                    :class="server.connected ? 'bg-blue-400' : (server.enabled === false ? 'bg-gray-400' : 'bg-red-400 animate-subtlePulse')"
                  ></span>
                  <h3 class="font-semibold truncate">{{ server.name }}</h3>
                  <!-- Source badge -->
                  <span v-if="server.source" class="px-1.5 py-0.5 text-[10px] rounded bg-makoclaw-bg text-makoclaw-text-secondary"
                    :title="server.path || ''">
                    {{ getSourceLabel(server.source) }}
                  </span>
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
                @click="openJSONModal(server)"
                class="px-3 py-1.5 text-xs text-makoclaw-text-secondary bg-makoclaw-bg border border-makoclaw-border rounded-lg hover:bg-makoclaw-border transition-colors"
                title="Edit as JSON"
              >JSON</button>
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
    <Transition name="modal">
    <div v-if="showModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-modal p-4" @click.self="showModal = false">
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

          <!-- Save Location (only for new servers) -->
          <div v-if="!editingServer">
            <label class="block text-sm font-medium mb-1">Save Location</label>
            <select
              v-model="form.saveTo"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent"
            >
              <option value="config">Config file (config.json)</option>
              <option value="user_folder">User MCP folder (~/.MakoClaw/users/*/mcp/)</option>
              <option value="global_folder">Global MCP folder (~/.MakoClaw/mcp/)</option>
            </select>
            <p class="text-xs text-makoclaw-text-secondary mt-1">
              Folder locations are easier to manage and share (Claude Desktop compatible format)
            </p>
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
    </Transition>

    <!-- Delete Confirmation Modal -->
    <Transition name="modal">
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-modal p-4" @click.self="showDeleteConfirm = false">
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
    </Transition>

    <!-- JSON Edit Modal -->
    <Transition name="modal">
    <div v-if="showJSONModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-modal p-4" @click.self="showJSONModal = false">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="font-semibold text-lg">Edit as JSON</h3>
          <span v-if="jsonEditingServer" class="text-sm text-makoclaw-text-secondary">{{ jsonEditingServer.name }}</span>
        </div>

        <!-- JSON validation error -->
        <div v-if="jsonError" class="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg">
          <p class="text-xs text-red-400">{{ jsonError }}</p>
        </div>

        <!-- JSON textarea -->
        <textarea
          v-model="jsonContent"
          rows="15"
          class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent resize-none font-mono"
          placeholder='{"command": "npx", "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path"], "env": {}}'
          @input="validateJSON"
        />

        <p class="text-xs text-makoclaw-text-secondary mt-2">
          Claude Desktop compatible format. Required fields: <code class="px-1 py-0.5 bg-makoclaw-bg rounded">command</code>
        </p>

        <!-- Modal Actions -->
        <div class="flex justify-end gap-3 mt-6">
          <button @click="showJSONModal = false"
            class="px-4 py-2 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">Cancel</button>
          <button @click="saveJSONConfig" :disabled="!!jsonError || savingJSON"
            class="px-4 py-2 text-sm bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors disabled:opacity-50">
            <span v-if="savingJSON">Saving...</span>
            <span v-else>Save Changes</span>
          </button>
        </div>
      </div>
    </div>
    </Transition>
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

// JSON edit modal state
const showJSONModal = ref(false)
const jsonEditingServer = ref(null)
const jsonContent = ref('')
const jsonError = ref('')
const savingJSON = ref(false)

// Form state
const defaultForm = () => ({
  name: '',
  command: '',
  argsText: '',
  envText: '',
  enabled: true,
  saveTo: 'config' // 'config', 'user_folder', 'global_folder'
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

// Get human-readable source label
const getSourceLabel = (source) => {
  const labels = {
    'user_folder': 'user folder',
    'global_folder': 'folder',
    'user_config': 'config',
    'global_config': 'global'
  }
  return labels[source] || source || 'config'
}

// Convert server data to JSON for editing
const serverToJSON = (server) => {
  const obj = {
    command: server.command || '',
    args: server.args || [],
    env: server.env || {}
  }
  if (server.enabled === false) {
    obj.enabled = false
  }
  return JSON.stringify(obj, null, 2)
}

// Open JSON edit modal
const openJSONModal = (server) => {
  jsonEditingServer.value = server
  jsonContent.value = serverToJSON(server)
  jsonError.value = ''
  showJSONModal.value = true
}

// Validate JSON content
const validateJSON = () => {
  if (!jsonContent.value.trim()) {
    jsonError.value = 'JSON content is required'
    return false
  }
  try {
    const parsed = JSON.parse(jsonContent.value)
    if (!parsed.command || typeof parsed.command !== 'string') {
      jsonError.value = 'Missing or invalid "command" field'
      return false
    }
    jsonError.value = ''
    return true
  } catch (e) {
    jsonError.value = `Invalid JSON: ${e.message}`
    return false
  }
}

// Save JSON config
const saveJSONConfig = async () => {
  if (!validateJSON() || !jsonEditingServer.value) return

  savingJSON.value = true
  try {
    const parsed = JSON.parse(jsonContent.value)
    const payload = {
      enabled: parsed.enabled !== false,
      command: parsed.command,
      args: parsed.args || [],
      env: parsed.env || {}
    }
    await advancedService.updateMCPServer(jsonEditingServer.value.name, payload)
    toast.success('Server updated successfully')
    showJSONModal.value = false
    await loadServers()
  } catch (err) {
    console.error('JSON save failed:', err)
    toast.error(err.response?.data?.error || 'Failed to save configuration')
  } finally {
    savingJSON.value = false
  }
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
  enabled: server.enabled !== false,
  saveTo: 'config' // When editing, always save to config (original location)
})

// Build API payload from form
const buildPayload = (isNew = false) => {
  const args = parseArgs(form.value.argsText)
  const env = parseEnv(form.value.envText)

  const payload = {
    name: form.value.name.trim(),
    command: form.value.command.trim(),
    args: args.length > 0 ? args : undefined,
    env: Object.keys(env).length > 0 ? env : undefined,
    enabled: form.value.enabled
  }

  // Include save_to for new servers
  if (isNew && form.value.saveTo) {
    payload.save_to = form.value.saveTo
  }

  return payload
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
    const isNew = !editingServer.value
    const payload = buildPayload(isNew)
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


