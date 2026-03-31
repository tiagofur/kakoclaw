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
        <button @click="devStore.fetchProjects()" class="mt-4 text-xs text-cyan-500 hover:text-cyan-400">
          <i class="fas fa-sync mr-1"></i> Refresh List
        </button>
      </div>

      <div class="mt-auto">
        <h3 class="text-xs font-semibold text-slate-500 uppercase tracking-wide mb-2">Bridge Status</h3>
        <div class="flex items-center text-sm px-3 py-2 bg-slate-800 rounded-md">
          <span class="w-2 h-2 rounded-full mr-2"
                :class="devStore.bridgeStatus === 'running' ? 'bg-green-500' : 'bg-red-500'"></span>
          {{ devStore.bridgeStatus }}
          <span v-if="devStore.usingHttpFallback" class="ml-2 text-xs bg-amber-800/60 text-amber-300 px-1.5 py-0.5 rounded" title="WebSocket unavailable — using HTTP fallback">HTTP</span>
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
          <div v-if="devStore.terminalHistory.length === 0" class="text-slate-600 italic">
            Waiting for connection... Select a project to start coding.
          </div>
          <div v-for="(msg, idx) in devStore.terminalHistory" :key="idx" class="break-words">
            <template v-if="msg.type === 'user'">
              <span class="text-emerald-400">> </span> <span class="text-emerald-100">{{ msg.message }}</span>
            </template>
            <template v-else-if="msg.type === 'error'">
              <span class="text-red-500">[ERROR] {{ msg.error || msg.message }}</span>
            </template>
            <template v-else>
              <span class="text-slate-400">[{{ msg.type }}]</span> 
              <span class="text-slate-300 ml-2 whitespace-pre-wrap">{{ msg.message }}</span>
            </template>
          </div>
        </div>
        
        <div class="mt-4 pt-4 border-t border-slate-800 shrink-0">
          <form @submit.prevent="submitPrompt" class="flex relative items-center">
            <span class="absolute left-3 text-emerald-500 font-bold">></span>
            <input 
              v-model="prompt" 
              type="text" 
              class="w-full bg-slate-900 border border-slate-700 rounded-md py-2 pl-8 pr-4 text-emerald-100 focus:outline-none focus:border-cyan-500"
              placeholder="Ask Claude or OpenCode..."
              :disabled="devStore.bridgeStatus !== 'running'"
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

onMounted(() => {
  devStore.fetchProjects()
  devStore.checkStatus()
})

const submitPrompt = () => {
  if (!prompt.value.trim()) return
  devStore.sendPrompt(prompt.value)
  prompt.value = ''
}

// Auto-scroll terminal
watch(() => devStore.terminalHistory.length, () => {
  nextTick(() => {
    if (terminalRef.value) {
      terminalRef.value.scrollTop = terminalRef.value.scrollHeight
    }
  })
})
</script>
