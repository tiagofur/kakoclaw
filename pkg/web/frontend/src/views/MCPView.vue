<template>
  <div class="h-full flex flex-col bg-kakoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-kakoclaw-border bg-kakoclaw-surface flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold bg-gradient-to-r from-kakoclaw-accent to-emerald-500 bg-clip-text text-transparent">MCP Servers</h2>
        <p class="text-sm text-kakoclaw-text-secondary mt-1">Manage Model Context Protocol server connections</p>
      </div>
      <div class="flex items-center gap-3">
        <span class="text-sm text-kakoclaw-text-secondary">
          {{ connectedCount }}/{{ servers.length }} connected
        </span>
        <button
          @click="loadServers"
          class="px-3 py-1.5 text-sm bg-kakoclaw-bg border border-kakoclaw-border rounded-lg hover:border-kakoclaw-accent/50 transition-colors"
        >Refresh</button>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6 custom-scrollbar">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
      </div>

      <template v-else>
        <!-- No servers configured -->
        <div v-if="servers.length === 0" class="text-center py-12 text-kakoclaw-text-secondary">
          <svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
          </svg>
          <p class="text-lg">No MCP servers configured</p>
          <p class="text-sm mt-2">Add MCP servers to <code class="px-1.5 py-0.5 bg-kakoclaw-surface rounded text-xs">config.json</code> under <code class="px-1.5 py-0.5 bg-kakoclaw-surface rounded text-xs">tools.mcp.servers</code></p>
          <div class="mt-4 max-w-lg mx-auto text-left">
            <pre class="bg-kakoclaw-surface border border-kakoclaw-border rounded-lg p-4 text-xs overflow-x-auto"><code>{
  "tools": {
    "mcp": {
      "servers": {
        "my-server": {
          "enabled": true,
          "command": "npx",
          "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path"],
          "env": {}
        }
      }
    }
  }
}</code></pre>
          </div>
        </div>

        <!-- Server Cards -->
        <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <div
            v-for="server in servers"
            :key="server.name"
            class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5 hover:border-kakoclaw-accent/50 transition-colors"
          >
            <!-- Server Header -->
            <div class="flex items-start justify-between mb-3">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span
                    class="w-2.5 h-2.5 rounded-full flex-shrink-0"
                    :class="server.connected ? 'bg-emerald-400' : 'bg-red-400'"
                  ></span>
                  <h3 class="font-semibold truncate">{{ server.name }}</h3>
                </div>
                <p class="text-xs text-kakoclaw-text-secondary mt-1 font-mono truncate" :title="server.command">{{ server.command }}</p>
              </div>
              <span
                class="ml-2 px-2 py-0.5 text-xs rounded-full flex-shrink-0"
                :class="server.connected
                  ? 'bg-emerald-500/10 text-emerald-400'
                  : 'bg-red-500/10 text-red-400'"
              >{{ server.connected ? 'Connected' : 'Disconnected' }}</span>
            </div>

            <!-- Server Info -->
            <div v-if="server.connected" class="space-y-2 mb-3">
              <div v-if="server.server_name" class="flex items-center gap-2 text-sm">
                <span class="text-kakoclaw-text-secondary">Server:</span>
                <span>{{ server.server_name }} <span v-if="server.server_version" class="text-kakoclaw-text-secondary">v{{ server.server_version }}</span></span>
              </div>
              <div class="flex items-center gap-2 text-sm">
                <span class="text-kakoclaw-text-secondary">Tools:</span>
                <span>{{ server.tool_count }} available</span>
              </div>
              <!-- Tool list -->
              <div v-if="server.tools && server.tools.length > 0" class="flex flex-wrap gap-1 mt-1">
                <span
                  v-for="tool in server.tools"
                  :key="tool"
                  class="px-2 py-0.5 text-xs bg-kakoclaw-bg rounded-full text-kakoclaw-text-secondary"
                >{{ tool }}</span>
              </div>
            </div>

            <!-- Error Info -->
            <div v-if="server.last_error" class="mb-3 p-2 bg-red-500/5 border border-red-500/20 rounded-lg">
              <p class="text-xs text-red-400 break-all">{{ server.last_error }}</p>
            </div>

            <!-- Actions -->
            <div class="flex items-center gap-2 mt-3 pt-3 border-t border-kakoclaw-border">
              <button
                @click="reconnect(server.name)"
                :disabled="reconnecting === server.name"
                class="px-3 py-1.5 text-xs bg-kakoclaw-accent text-white rounded-lg hover:bg-kakoclaw-accent/90 transition-colors disabled:opacity-50"
              >
                <span v-if="reconnecting === server.name">Reconnecting...</span>
                <span v-else>{{ server.connected ? 'Reconnect' : 'Connect' }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Info Box -->
        <div v-if="servers.length > 0" class="mt-6 p-4 bg-kakoclaw-surface border border-kakoclaw-border rounded-xl">
          <h4 class="text-sm font-semibold mb-2">About MCP</h4>
          <p class="text-xs text-kakoclaw-text-secondary leading-relaxed">
            The Model Context Protocol (MCP) allows KakoClaw to connect to external tool servers.
            Tools discovered from MCP servers are automatically available to the AI agent during conversations.
            Tool names are prefixed with <code class="px-1 py-0.5 bg-kakoclaw-bg rounded">mcp_&lt;server&gt;_</code> to avoid conflicts.
          </p>
        </div>
      </template>
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
const servers = ref([])

const connectedCount = computed(() => servers.value.filter(s => s.connected).length)

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

onMounted(() => loadServers())
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 8px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: rgba(156, 163, 175, 0.5); border-radius: 4px; }
</style>
