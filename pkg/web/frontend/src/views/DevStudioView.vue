<template>
  <div class="h-full flex bg-slate-900 text-slate-300">
    <!-- Sidebar / Projects -->
    <div class="w-64 border-r border-slate-700 p-4 shrink-0 flex flex-col">
      <h2 class="text-xl font-bold mb-4 text-white">Dev Studio</h2>

      <div class="mb-6">
        <h3 class="text-xs font-semibold text-slate-500 uppercase tracking-wide mb-2">Projects</h3>
        <ul class="space-y-1">
          <li v-for="p in devStore.projects" :key="p.path">
            <button
              @click="devStore.fetchHistory(p.name); devStore.startBridge(p.path)"
              :class="['w-full text-left px-3 py-2 rounded-md transition-colors text-sm',
                       devStore.currentProject === p.path ? 'bg-cyan-900/50 text-cyan-400' : 'hover:bg-slate-800 text-slate-300']"
            >
              <i class="fas fa-folder mr-2"></i> {{ p.name }}
            </button>
          </li>
        </ul>
        <div v-if="devStore.projects.length === 0" class="text-xs text-slate-500 mt-2 px-3">
          No projects found. Create a folder in your workspace to get started.
        </div>
        <button @click="devStore.fetchProjects()" class="mt-4 text-xs text-cyan-500 hover:text-cyan-400">
          <i class="fas fa-sync mr-1"></i> Refresh List
        </button>
      </div>

      <div class="mt-auto">
        <h3 class="text-xs font-semibold text-slate-500 uppercase tracking-wide mb-2">Bridge Status</h3>
        <div class="flex items-center text-sm px-3 py-2 bg-slate-800 rounded-md">
          <span class="w-2 h-2 rounded-full mr-2"
                :class="devStore.bridgeStatus === 'running' ? 'bg-green-500' : (devStore.bridgeError ? 'bg-red-500' : 'bg-yellow-500')"></span>
          {{ devStore.bridgeStatus }}
          <span v-if="devStore.usingHttpFallback" class="ml-2 text-xs bg-amber-800/60 text-amber-300 px-1.5 py-0.5 rounded" title="WebSocket unavailable — using HTTP fallback">HTTP</span>
        </div>
        <div v-if="devStore.bridgeError" class="mt-1 text-xs text-red-400 px-3">
          {{ devStore.bridgeError }}
        </div>
        <button
          v-if="devStore.bridgeStatus === 'running'"
          @click="devStore.stopBridge()"
          class="w-full mt-2 text-xs py-1 px-2 border border-red-900 text-red-400 hover:bg-red-900/50 rounded transition-colors"
        >
          Stop Bridge
        </button>
      </div>
    </div>

    <!-- Main Workspace -->
    <div class="flex-1 flex flex-col min-w-0">
      <div class="flex border-b border-slate-700 px-4 py-2 bg-slate-800/50">
        <button v-for="tab in ['Terminal', 'Memory']" :key="tab"
          @click="activeTab = tab"
          :class="['px-4 py-1.5 text-sm font-medium rounded-t-lg transition-colors mr-2',
                   activeTab === tab ? 'bg-slate-700 text-white' : 'text-slate-400 hover:bg-slate-700/50 hover:text-slate-200']"
        >
          {{ tab }}
        </button>
      </div>

      <!-- Terminal View -->
      <div v-show="activeTab === 'Terminal'" class="flex-1 flex flex-col min-h-0 bg-black p-4 font-mono text-sm relative">
        <div class="flex-1 overflow-y-auto pb-4 space-y-2" ref="terminalRef">
          <div v-if="devStore.terminalHistory.length === 0" class="text-slate-600 italic space-y-1">
            <p v-if="devStore.bridgeStatus === 'running'">Bridge is running. Type a prompt below to start.</p>
            <p v-else>Select a project from the sidebar to start the bridge.</p>
            <p v-if="devStore.bridgeStatus === 'idle'" class="text-slate-700 text-xs">
              If no projects appear, make sure Dev Studio is enabled in your config
              (dev_studio.enabled: true) and that your workspace has at least one subdirectory.
            </p>
          </div>
          <div v-for="(msg, idx) in devStore.terminalHistory" :key="idx" class="break-words">
            <template v-if="msg.type === 'user'">
              <span class="text-emerald-400">> </span> <span class="text-emerald-100">{{ msg.message }}</span>
            </template>
            <template v-else-if="msg.type === 'error'">
              <span class="text-red-500">[ERROR] {{ msg.message || msg.error }}</span>
            </template>
            <template v-else-if="msg.type === 'assistant'">
              <span class="text-slate-300 whitespace-pre-wrap">{{ msg.message }}</span>
            </template>
            <template v-else-if="msg.type === 'tool_use'">
              <span class="text-amber-500">[tool]</span>
              <span class="text-amber-300 ml-2">{{ msg.name }}</span>
            </template>
            <template v-else-if="msg.type === 'tool_result'">
              <span class="text-slate-500">[result]</span>
              <span class="text-slate-400 ml-2 whitespace-pre-wrap">{{ msg.message }}</span>
            </template>
            <template v-else-if="msg.type === 'system'">
              <span class="text-cyan-600">[system]</span>
              <span class="text-cyan-500 ml-2">session={{ msg.session_id }} model={{ msg.model }}</span>
            </template>
            <template v-else-if="msg.type === 'result'">
              <span class="text-green-600">[done]</span>
              <span class="text-green-500 ml-2">{{ msg.message }}</span>
              <span v-if="msg.cost_usd" class="text-slate-500 ml-2 text-xs">${{ msg.cost_usd?.toFixed(4) }} · {{ msg.duration_ms }}ms</span>
            </template>
            <template v-else>
              <span class="text-slate-400">[{{ msg.type }}]</span>
              <span class="text-slate-300 ml-2 whitespace-pre-wrap">{{ msg.message }}</span>
            </template>
          </div>
        </div>

        <div class="mt-4 pt-4 border-t border-slate-800 shrink-0">
          <form @submit.prevent="submitPrompt" class="flex relative items-center">
            <span class="absolute left-3 font-bold" :class="devStore.bridgeStatus === 'running' ? 'text-emerald-500' : 'text-slate-600'">></span>
            <input
              ref="inputRef"
              v-model="prompt"
              type="text"
              class="w-full bg-slate-900 border rounded-md py-2 pl-8 pr-4 text-emerald-100 focus:outline-none focus:border-cyan-500 transition-colors"
              :class="devStore.bridgeStatus === 'running' ? 'border-slate-700' : 'border-slate-800'"
              :placeholder="devStore.bridgeStatus === 'running' ? 'Ask Claude or OpenCode...' : 'Select a project to start...'"
            />
          </form>
        </div>
      </div>

      <!-- Memory View -->
      <div v-show="activeTab === 'Memory'" class="flex-1 p-6 overflow-y-auto">
        <h3 class="text-lg font-bold mb-4 text-white">Semantic Memory</h3>
        <div class="flex gap-2 mb-6">
          <input
            v-model="memoryQuery"
            type="text"
            class="flex-1 bg-slate-800 border border-slate-700 rounded-md px-4 py-2 text-white focus:outline-none focus:border-cyan-500"
            placeholder="Search dev memory..."
            @keyup.enter="devStore.searchMemory(memoryQuery)"
          />
          <button @click="devStore.searchMemory(memoryQuery)" class="bg-cyan-600 hover:bg-cyan-500 text-white px-4 py-2 rounded-md transition-colors">
            Search
          </button>
        </div>

        <div class="space-y-4">
          <div v-for="result in devStore.searchResults" :key="result.id" class="p-4 bg-slate-800 rounded-lg border border-slate-700">
            <div class="text-sm text-slate-400 mb-2">{{ new Date(result.created_at).toLocaleString() }}</div>
            <div class="text-slate-200 whitespace-pre-wrap">{{ result.content }}</div>
          </div>
          <div v-if="devStore.searchResults.length === 0" class="text-slate-500 text-center py-8">
            No memories found. Type a query to search semantic context.
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, watch } from 'vue'
import { useDevStudioStore } from '@/stores/devStudioStore'

const devStore = useDevStudioStore()
const activeTab = ref('Terminal')
const prompt = ref('')
const memoryQuery = ref('')
const terminalRef = ref(null)
const inputRef = ref(null)

onMounted(() => {
  devStore.fetchProjects()
  devStore.checkStatus()
  // Auto-focus the terminal input
  nextTick(() => {
    if (inputRef.value) inputRef.value.focus()
  })
})

const submitPrompt = () => {
  if (!prompt.value.trim()) return
  if (devStore.bridgeStatus !== 'running') {
    devStore.terminalHistory.push({ type: 'error', message: 'Bridge is not running. Select a project first.' })
    return
  }
  devStore.sendPrompt(prompt.value)
  prompt.value = ''
  // Re-focus input after submit
  nextTick(() => {
    if (inputRef.value) inputRef.value.focus()
  })
}

// Auto-scroll terminal
watch(() => devStore.terminalHistory.length, () => {
  nextTick(() => {
    if (terminalRef.value) {
      terminalRef.value.scrollTop = terminalRef.value.scrollHeight
    }
  })
})

// Auto-focus input when bridge starts running
watch(() => devStore.bridgeStatus, (newStatus) => {
  if (newStatus === 'running') {
    nextTick(() => {
      if (inputRef.value) inputRef.value.focus()
    })
  }
})
</script>
