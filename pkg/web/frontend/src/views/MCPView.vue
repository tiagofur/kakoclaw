<template>
  <div class="flex flex-col h-full bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-orange-500/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-amber-500/20 via-transparent to-transparent" />
    </div>

    <!-- Page Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <div class="flex items-center gap-3">
          <!-- Icon Container -->
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-orange-500/20 to-amber-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-orange-500/10">
            <svg
              class="w-5 h-5 sm:w-6 sm:h-6 text-orange-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M5.25 14.25h13.5m-13.5 0a3 3 0 01-3-3m3 3a3 3 0 100 6h13.5a3 3 0 100-6m-16.5-3a3 3 0 013-3h13.5a3 3 0 013 3m-19.5 0a4.5 4.5 0 01.9-2.7L5.737 5.1a3.375 3.375 0 012.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 01.9 2.7m0 0a3 3 0 01-3 3m0 3h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008zm-3 6h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008z"
              />
            </svg>
          </div>

          <!-- Title -->
          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-orange-400 bg-clip-text text-transparent">
              MCP Servers
            </h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5">
              Model Context Protocol connections
            </p>
          </div>

          <!-- Connection Status Badge -->
          <div class="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 rounded-xl">
            <span class="relative flex h-2.5 w-2.5">
              <span
                v-if="connectedCount > 0"
                class="animate-ping absolute inline-flex h-full w-full rounded-full bg-orange-400 opacity-75"
              />
              <span
                class="relative inline-flex rounded-full h-2.5 w-2.5"
                :class="connectedCount > 0 ? 'bg-orange-400' : 'bg-makoclaw-text-secondary/50'"
              />
            </span>
            <span class="text-sm font-medium text-makoclaw-text">{{ connectedCount }}/{{ servers.length }}</span>
            <span class="text-sm text-makoclaw-text-secondary">connected</span>
          </div>

          <!-- Action Buttons -->
          <div class="flex items-center gap-2">
            <button
              class="p-2.5 min-h-[40px] min-w-[40px] flex items-center justify-center rounded-xl border border-makoclaw-border/50 hover:bg-orange-500/10 hover:text-orange-400 hover:border-orange-500/30 text-makoclaw-text-secondary transition-all"
              title="Refresh"
              @click="loadServers"
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
            </button>

            <button
              :disabled="reconnectingAll || servers.length === 0"
              class="hidden sm:flex px-3 py-2 min-h-[40px] items-center gap-2 rounded-xl border border-makoclaw-border/50 hover:bg-orange-500/10 hover:text-orange-400 hover:border-orange-500/30 text-makoclaw-text-secondary text-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              @click="reconnectAll"
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
                />
              </svg>
              <span v-if="reconnectingAll">Reconnecting...</span>
              <span v-else>Reconnect All</span>
            </button>

            <button
              class="px-4 sm:px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white rounded-xl transition-all shadow-lg shadow-orange-500/25 hover:shadow-orange-500/40 text-sm font-bold flex items-center gap-2 active:scale-95 flex-shrink-0"
              @click="openAddModal"
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 4v16m8-8H4"
                />
              </svg>
              <span class="hidden sm:inline">Add Server</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto p-4 sm:p-6 custom-scrollbar">
      <!-- Loading Skeleton -->
      <div
        v-if="loading"
        class="grid grid-cols-1 lg:grid-cols-2 gap-4"
      >
        <div
          v-for="i in 3"
          :key="i"
          class="glass-panel p-5 rounded-xl"
        >
          <div class="flex items-start justify-between mb-3">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-2">
                <div class="skeleton w-2.5 h-2.5 rounded-full" />
                <div class="skeleton h-4 w-32 rounded" />
              </div>
              <div class="skeleton h-3 w-48 rounded" />
            </div>
            <div class="flex items-center gap-2">
              <div class="skeleton h-5 w-16 rounded-full" />
              <div class="skeleton h-5 w-20 rounded-full" />
            </div>
          </div>
          <div class="space-y-2 mb-3">
            <div class="skeleton h-3 w-28 rounded" />
            <div class="flex gap-1">
              <div class="skeleton h-5 w-20 rounded-full" />
              <div class="skeleton h-5 w-24 rounded-full" />
              <div class="skeleton h-5 w-16 rounded-full" />
            </div>
          </div>
          <div class="flex items-center gap-2 mt-3 pt-3 border-t border-makoclaw-border/20">
            <div class="skeleton h-7 w-20 rounded-lg" />
            <div class="skeleton h-7 w-14 rounded-lg" />
            <div class="skeleton h-7 w-12 rounded-lg" />
          </div>
        </div>
      </div>

      <template v-else>
        <!-- Empty State -->
        <div
          v-if="servers.length === 0"
          class="flex flex-col items-center justify-center py-16 text-center"
        >
          <div class="relative">
            <!-- Glow effect -->
            <div class="absolute inset-0 bg-gradient-to-br from-orange-500/30 to-amber-500/30 rounded-3xl blur-2xl opacity-50" />
            <div class="relative glass-panel p-8 rounded-2xl shadow-2xl ring-1 ring-white/10">
              <div class="w-16 h-16 mx-auto rounded-2xl bg-gradient-to-br from-orange-500/20 to-amber-500/20 flex items-center justify-center ring-1 ring-white/20">
                <svg
                  class="w-8 h-8 text-orange-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="1.5"
                    d="M5.25 14.25h13.5m-13.5 0a3 3 0 01-3-3m3 3a3 3 0 100 6h13.5a3 3 0 100-6m-16.5-3a3 3 0 013-3h13.5a3 3 0 013 3m-19.5 0a4.5 4.5 0 01.9-2.7L5.737 5.1a3.375 3.375 0 012.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 01.9 2.7m0 0a3 3 0 01-3 3m0 3h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008zm-3 6h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008z"
                  />
                </svg>
              </div>
            </div>
          </div>
          <h3 class="text-lg font-bold text-makoclaw-text mt-6">
            No MCP servers configured
          </h3>
          <p class="text-sm text-makoclaw-text-secondary/70 mt-2 max-w-xs">
            Connect external tool servers to extend your agent's capabilities.
          </p>
          <button
            class="mt-6 px-5 py-2.5 bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white rounded-xl transition-all shadow-lg shadow-orange-500/25 hover:shadow-orange-500/40 text-sm font-bold flex items-center gap-2 active:scale-95"
            @click="openAddModal"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              />
            </svg>
            Add Server
          </button>
        </div>

        <!-- Server Cards -->
        <div
          v-else
          class="grid grid-cols-1 lg:grid-cols-2 gap-4"
        >
          <div
            v-for="server in servers"
            :key="server.name"
            class="server-card group"
          >
            <!-- Server Header -->
            <div class="flex items-start justify-between mb-3">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <!-- Connection Status Indicator -->
                  <span class="relative flex-shrink-0">
                    <span
                      v-if="server.connected"
                      class="animate-ping absolute inline-flex h-2.5 w-2.5 rounded-full bg-orange-400 opacity-75"
                    />
                    <span
                      class="relative inline-flex w-2.5 h-2.5 rounded-full"
                      :class="server.connected ? 'bg-orange-400' : (server.enabled === false ? 'bg-gray-400' : 'bg-red-400')"
                    />
                  </span>
                  <h3 class="font-semibold text-makoclaw-text truncate group-hover:text-orange-400 transition-colors">
                    {{ server.name }}
                  </h3>
                  <!-- Source badge -->
                  <span
                    v-if="server.source"
                    class="px-1.5 py-0.5 text-[10px] rounded-md bg-makoclaw-bg/50 text-makoclaw-text-secondary border border-makoclaw-border/30"
                    :title="server.path || ''"
                  >
                    {{ getSourceLabel(server.source) }}
                  </span>
                </div>
                <p
                  class="text-xs text-makoclaw-text-secondary mt-1 font-mono truncate"
                  :title="getCommandDisplay(server)"
                >
                  {{ getCommandDisplay(server) }}
                </p>
              </div>
              <div class="flex items-center gap-2 ml-2 flex-shrink-0">
                <!-- Enable/Disable Toggle -->
                <button
                  :disabled="togglingEnabled === server.name"
                  class="px-2 py-0.5 text-xs rounded-full transition-all border"
                  :class="server.enabled !== false
                    ? 'bg-orange-500/10 text-orange-400 border-orange-500/30 hover:bg-orange-500/20'
                    : 'bg-gray-500/10 text-gray-400 border-gray-500/30 hover:bg-gray-500/20'"
                  :title="server.enabled !== false ? 'Click to disable' : 'Click to enable'"
                  @click="toggleEnabled(server)"
                >
                  <span v-if="togglingEnabled === server.name">...</span>
                  <span v-else>{{ server.enabled !== false ? 'Enabled' : 'Disabled' }}</span>
                </button>
                <span
                  class="px-2 py-0.5 text-xs rounded-full border"
                  :class="server.connected
                    ? 'bg-orange-500/10 text-orange-400 border-orange-500/30'
                    : 'bg-red-500/10 text-red-400 border-red-500/30'"
                >{{ server.connected ? 'Connected' : 'Disconnected' }}</span>
              </div>
            </div>

            <!-- Server Info -->
            <div
              v-if="server.connected"
              class="space-y-2 mb-3"
            >
              <div
                v-if="server.server_name"
                class="flex items-center gap-2 text-sm"
              >
                <span class="text-makoclaw-text-secondary">Server:</span>
                <span class="text-makoclaw-text">{{ server.server_name }} <span
                  v-if="server.server_version"
                  class="text-makoclaw-text-secondary"
                >v{{ server.server_version }}</span></span>
              </div>
              <div class="flex items-center gap-2 text-sm">
                <span class="text-makoclaw-text-secondary">Tools:</span>
                <span class="px-2 py-0.5 text-xs bg-orange-500/10 text-orange-400 rounded-full font-medium">{{ server.tool_count }} available</span>
              </div>
              <!-- Tool list -->
              <div
                v-if="server.tools && server.tools.length > 0"
                class="flex flex-wrap gap-1 mt-1"
              >
                <span
                  v-for="tool in server.tools"
                  :key="tool"
                  class="px-2 py-0.5 text-xs bg-makoclaw-bg/50 rounded-full text-makoclaw-text-secondary border border-makoclaw-border/30"
                >{{ tool }}</span>
              </div>
            </div>

            <!-- Command and Args Display (when not connected) -->
            <div
              v-if="!server.connected && server.args && server.args.length > 0"
              class="mb-3"
            >
              <div class="text-xs text-makoclaw-text-secondary mb-1">
                Arguments:
              </div>
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="(arg, idx) in server.args.slice(0, 5)"
                  :key="idx"
                  class="px-2 py-0.5 text-xs bg-makoclaw-bg/50 rounded-md font-mono text-makoclaw-text-secondary border border-makoclaw-border/30"
                >{{ arg }}</span>
                <span
                  v-if="server.args.length > 5"
                  class="px-2 py-0.5 text-xs text-makoclaw-text-secondary"
                >
                  +{{ server.args.length - 5 }} more
                </span>
              </div>
            </div>

            <!-- Error Info -->
            <div
              v-if="server.last_error"
              class="mb-3 p-3 bg-red-500/5 border border-red-500/20 rounded-xl"
            >
              <p class="text-xs text-red-400 break-all">
                {{ server.last_error }}
              </p>
            </div>

            <!-- Test Result -->
            <div
              v-if="testResults[server.name]"
              class="mb-3 p-3 rounded-xl border"
              :class="testResults[server.name].success
                ? 'bg-orange-500/5 border-orange-500/20'
                : 'bg-red-500/5 border-red-500/20'"
            >
              <p
                class="text-xs"
                :class="testResults[server.name].success ? 'text-orange-400' : 'text-red-400'"
              >
                {{ testResults[server.name].message }}
              </p>
            </div>

            <!-- Actions -->
            <div class="flex items-center gap-2 mt-3 pt-3 border-t border-makoclaw-border/20">
              <button
                :disabled="reconnecting === server.name"
                class="px-3 py-1.5 text-xs font-medium bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white rounded-lg shadow-sm shadow-orange-500/25 transition-all active:scale-95 disabled:opacity-50"
                @click="reconnect(server.name)"
              >
                <span v-if="reconnecting === server.name">Reconnecting...</span>
                <span v-else>{{ server.connected ? 'Reconnect' : 'Connect' }}</span>
              </button>
              <button
                :disabled="testing === server.name"
                class="px-3 py-1.5 text-xs font-medium text-makoclaw-text bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-lg hover:bg-orange-500/10 hover:text-orange-400 hover:border-orange-500/30 transition-all disabled:opacity-50"
                @click="testServer(server.name)"
              >
                <span v-if="testing === server.name">Testing...</span>
                <span v-else>Test</span>
              </button>
              <button
                class="px-3 py-1.5 text-xs font-medium text-makoclaw-text bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-lg hover:bg-orange-500/10 hover:text-orange-400 hover:border-orange-500/30 transition-all"
                @click="openEditModal(server)"
              >
                Edit
              </button>
              <button
                class="px-3 py-1.5 text-xs font-medium text-makoclaw-text bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-lg hover:bg-orange-500/10 hover:text-orange-400 hover:border-orange-500/30 transition-all"
                title="Edit as JSON"
                @click="openJSONModal(server)"
              >
                JSON
              </button>
              <button
                class="px-3 py-1.5 text-xs font-medium text-red-400 bg-red-500/10 border border-red-500/20 rounded-lg hover:bg-red-500/20 transition-all"
                @click="confirmDeleteServer(server)"
              >
                Delete
              </button>
            </div>
          </div>
        </div>

        <!-- Info Box -->
        <div
          v-if="servers.length > 0"
          class="mt-6 glass-panel p-4 rounded-xl"
        >
          <div class="flex items-start gap-3">
            <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-orange-500/20 to-amber-500/20 flex items-center justify-center flex-shrink-0">
              <svg
                class="w-4 h-4 text-orange-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <div>
              <h4 class="text-sm font-semibold text-makoclaw-text mb-1">
                About MCP
              </h4>
              <p class="text-xs text-makoclaw-text-secondary leading-relaxed">
                The Model Context Protocol (MCP) allows MakoClaw to connect to external tool servers.
                Tools discovered from MCP servers are automatically available to the AI agent during conversations.
                Tool names are prefixed with <code class="px-1.5 py-0.5 bg-makoclaw-bg/50 rounded border border-makoclaw-border/30 text-orange-400">mcp_&lt;server&gt;_</code> to avoid conflicts.
              </p>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Add / Edit Server Modal -->
    <Transition name="modal">
      <div
        v-if="showModal"
        class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-modal p-4"
        @click.self="showModal = false"
      >
        <div class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl max-w-lg w-full max-h-[90vh] overflow-y-auto shadow-2xl ring-1 ring-white/10 animate-scaleIn">
          <!-- Modal Header -->
          <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-lg bg-gradient-to-br from-orange-500/20 to-amber-500/20 flex items-center justify-center ring-1 ring-white/10">
                <svg
                  class="w-4.5 h-4.5 text-orange-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M5.25 14.25h13.5m-13.5 0a3 3 0 01-3-3m3 3a3 3 0 100 6h13.5a3 3 0 100-6m-16.5-3a3 3 0 013-3h13.5a3 3 0 013 3"
                  />
                </svg>
              </div>
              <h3 class="font-bold text-makoclaw-text">
                {{ editingServer ? 'Edit MCP Server' : 'Add MCP Server' }}
              </h3>
            </div>
          </div>

          <!-- Modal Content -->
          <div class="p-4 sm:p-6 space-y-4">
            <!-- Name -->
            <div>
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Name <span class="text-red-400">*</span></label>
              <input
                v-model="form.name"
                type="text"
                placeholder="my-server"
                :disabled="!!editingServer"
                class="w-full px-3 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm outline-none focus:ring-2 focus:ring-orange-500/30 focus:border-orange-500/50 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              >
              <p class="text-xs text-makoclaw-text-secondary mt-1.5">
                Unique identifier for this server (alphanumeric and hyphens)
              </p>
            </div>

            <!-- Command -->
            <div>
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Command <span class="text-red-400">*</span></label>
              <input
                v-model="form.command"
                type="text"
                placeholder="npx"
                class="w-full px-3 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm outline-none focus:ring-2 focus:ring-orange-500/30 focus:border-orange-500/50 transition-all font-mono"
              >
              <p class="text-xs text-makoclaw-text-secondary mt-1.5">
                Command to start the MCP server (e.g., npx, node, python)
              </p>
            </div>

            <!-- Args -->
            <div>
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Arguments</label>
              <textarea
                v-model="form.argsText"
                rows="3"
                placeholder="-y
@modelcontextprotocol/server-filesystem
/path/to/directory"
                class="w-full px-3 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm outline-none focus:ring-2 focus:ring-orange-500/30 focus:border-orange-500/50 transition-all resize-none font-mono"
              />
              <p class="text-xs text-makoclaw-text-secondary mt-1.5">
                One argument per line
              </p>
            </div>

            <!-- Env -->
            <div>
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Environment Variables</label>
              <textarea
                v-model="form.envText"
                rows="3"
                placeholder="API_KEY=your-key
DEBUG=true"
                class="w-full px-3 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm outline-none focus:ring-2 focus:ring-orange-500/30 focus:border-orange-500/50 transition-all resize-none font-mono"
              />
              <p class="text-xs text-makoclaw-text-secondary mt-1.5">
                KEY=value pairs, one per line
              </p>
            </div>

            <!-- Enabled -->
            <div class="flex items-center gap-3 p-3 bg-makoclaw-bg/30 rounded-xl border border-makoclaw-border/30">
              <input
                id="enabled"
                v-model="form.enabled"
                type="checkbox"
                class="w-4 h-4 rounded border-makoclaw-border bg-makoclaw-surface text-orange-500 focus:ring-orange-500 transition-all cursor-pointer"
              >
              <label
                for="enabled"
                class="text-sm text-makoclaw-text cursor-pointer"
              >Enable server on save</label>
            </div>

            <!-- Save Location (only for new servers) -->
            <div v-if="!editingServer">
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Save Location</label>
              <select
                v-model="form.saveTo"
                class="w-full px-3 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm outline-none focus:ring-2 focus:ring-orange-500/30 focus:border-orange-500/50 transition-all cursor-pointer"
              >
                <option value="config">
                  Config file (config.json)
                </option>
                <option value="user_folder">
                  User MCP folder (~/.MakoClaw/users/*/mcp/)
                </option>
                <option value="global_folder">
                  Global MCP folder (~/.MakoClaw/mcp/)
                </option>
              </select>
              <p class="text-xs text-makoclaw-text-secondary mt-1.5">
                Folder locations are easier to manage and share (Claude Desktop compatible format)
              </p>
            </div>

            <!-- Test Connection Button -->
            <div v-if="editingServer">
              <button
                :disabled="modalTesting"
                class="w-full px-4 py-2.5 text-sm font-medium bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl hover:bg-orange-500/10 hover:text-orange-400 hover:border-orange-500/30 transition-all disabled:opacity-50"
                @click="testServerFromModal"
              >
                <span v-if="modalTesting">Testing Connection...</span>
                <span v-else>Test Connection</span>
              </button>
              <div
                v-if="modalTestResult"
                class="mt-3 p-3 rounded-xl border"
                :class="modalTestResult.success
                  ? 'bg-orange-500/5 border-orange-500/20'
                  : 'bg-red-500/5 border-red-500/20'"
              >
                <p
                  class="text-xs"
                  :class="modalTestResult.success ? 'text-orange-400' : 'text-red-400'"
                >
                  {{ modalTestResult.message }}
                </p>
              </div>
            </div>
          </div>

          <!-- Modal Actions -->
          <div class="flex justify-end gap-3 p-4 border-t border-makoclaw-border/30">
            <button
              class="px-4 py-2 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
              @click="showModal = false"
            >
              Cancel
            </button>
            <button
              :disabled="!canSubmit || submitting"
              class="px-5 py-2 text-sm font-bold bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white rounded-xl shadow-lg shadow-orange-500/25 transition-all active:scale-95 disabled:opacity-50"
              @click="submitServer"
            >
              <span v-if="submitting">{{ editingServer ? 'Saving...' : 'Adding...' }}</span>
              <span v-else>{{ editingServer ? 'Save Changes' : 'Add Server' }}</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Delete Confirmation Modal -->
    <Transition name="modal">
      <div
        v-if="showDeleteConfirm"
        class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-modal p-4"
        @click.self="showDeleteConfirm = false"
      >
        <div class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl max-w-sm w-full shadow-2xl ring-1 ring-white/10 animate-scaleIn">
          <!-- Header -->
          <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-red-500/10 to-transparent">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-lg bg-red-500/20 flex items-center justify-center">
                <svg
                  class="w-4.5 h-4.5 text-red-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
              </div>
              <h3 class="font-bold text-makoclaw-text">
                Delete Server
              </h3>
            </div>
          </div>

          <div class="p-4">
            <p class="text-sm text-makoclaw-text-secondary mb-3">
              Are you sure you want to delete <span class="font-medium text-makoclaw-text">{{ deletingServer?.name }}</span>?
            </p>
            <div class="p-3 bg-yellow-500/5 border border-yellow-500/20 rounded-xl">
              <p class="text-xs text-yellow-400">
                Warning: All tools from this server will become unavailable to the AI agent.
              </p>
            </div>
          </div>

          <div class="flex justify-end gap-3 p-4 border-t border-makoclaw-border/30">
            <button
              class="px-4 py-2 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
              @click="showDeleteConfirm = false"
            >
              Cancel
            </button>
            <button
              :disabled="deleting"
              class="px-5 py-2 text-sm font-bold bg-red-500 hover:bg-red-600 text-white rounded-xl transition-all active:scale-95 disabled:opacity-50"
              @click="executeDeleteServer"
            >
              <span v-if="deleting">Deleting...</span>
              <span v-else>Delete</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- JSON Edit Modal -->
    <Transition name="modal">
      <div
        v-if="showJSONModal"
        class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-modal p-4"
        @click.self="showJSONModal = false"
      >
        <div class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto shadow-2xl ring-1 ring-white/10 animate-scaleIn">
          <!-- Header -->
          <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-9 h-9 rounded-lg bg-gradient-to-br from-orange-500/20 to-amber-500/20 flex items-center justify-center ring-1 ring-white/10">
                  <svg
                    class="w-4.5 h-4.5 text-orange-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
                    />
                  </svg>
                </div>
                <h3 class="font-bold text-makoclaw-text">
                  Edit as JSON
                </h3>
              </div>
              <span
                v-if="jsonEditingServer"
                class="text-sm text-makoclaw-text-secondary px-2 py-1 bg-makoclaw-bg/50 rounded-lg"
              >{{ jsonEditingServer.name }}</span>
            </div>
          </div>

          <div class="p-4 sm:p-6">
            <!-- JSON validation error -->
            <div
              v-if="jsonError"
              class="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded-xl"
            >
              <p class="text-xs text-red-400">
                {{ jsonError }}
              </p>
            </div>

            <!-- JSON textarea -->
            <textarea
              v-model="jsonContent"
              rows="15"
              class="w-full px-3 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm outline-none focus:ring-2 focus:ring-orange-500/30 focus:border-orange-500/50 transition-all resize-none font-mono"
              placeholder="{&quot;command&quot;: &quot;npx&quot;, &quot;args&quot;: [&quot;-y&quot;, &quot;@modelcontextprotocol/server-filesystem&quot;, &quot;/path&quot;], &quot;env&quot;: {}}"
              @input="validateJSON"
            />

            <p class="text-xs text-makoclaw-text-secondary mt-2">
              Claude Desktop compatible format. Required fields: <code class="px-1.5 py-0.5 bg-makoclaw-bg/50 rounded border border-makoclaw-border/30 text-orange-400">command</code>
            </p>
          </div>

          <!-- Actions -->
          <div class="flex justify-end gap-3 p-4 border-t border-makoclaw-border/30">
            <button
              class="px-4 py-2 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
              @click="showJSONModal = false"
            >
              Cancel
            </button>
            <button
              :disabled="!!jsonError || savingJSON"
              class="px-5 py-2 text-sm font-bold bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white rounded-xl shadow-lg shadow-orange-500/25 transition-all active:scale-95 disabled:opacity-50"
              @click="saveJSONConfig"
            >
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

<style scoped>
.server-card {
  @apply p-5 rounded-xl bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 hover:bg-makoclaw-surface/50 hover:border-orange-500/20 transition-all duration-200;
}
</style>
