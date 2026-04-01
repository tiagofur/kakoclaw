<template>
  <div class="flex flex-col h-full bg-makoclaw-bg relative overflow-hidden">
    <!-- Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 py-3">
        <div class="flex items-center gap-3">
          <!-- Sidebar Toggle -->
          <button
            class="p-2 min-h-[40px] min-w-[40px] rounded-xl bg-makoclaw-surface/50 border border-makoclaw-border/50 hover:bg-makoclaw-surface-hover hover:border-cyan-400/30 transition-all flex items-center justify-center active:scale-95"
            title="Toggle Sidebar"
            @click="toggleSidebar"
          >
            <svg
              class="w-4 h-4 text-makoclaw-text-secondary transition-transform duration-500"
              :class="{ 'rotate-180': !showSidebar }"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M11 19l-7-7 7-7m8 14l-7-7 7-7"
              />
            </svg>
          </button>

          <!-- Icon -->
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-cyan-500/20 to-blue-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-cyan-500/10">
            <svg
              class="w-5 h-5 text-cyan-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
              />
            </svg>
          </div>

          <div class="flex-1 min-w-0">
            <h1 class="text-xl font-bold bg-gradient-to-r from-makoclaw-text to-cyan-400 bg-clip-text text-transparent">
              Dev Studio
            </h1>
            <p class="text-[10px] text-makoclaw-text-secondary uppercase tracking-widest font-bold">
              Engineering Workspace
            </p>
          </div>

          <!-- Bridge Status Indicator (Header) -->
          <div
            class="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-xl border transition-all"
            :class="devStore.bridgeStatus === 'running' ? 'bg-makoclaw-success/10 border-makoclaw-success/30' : 'bg-makoclaw-surface/50 border-makoclaw-border/50'"
          >
            <div
              class="w-2 h-2 rounded-full animate-pulse"
              :class="devStore.bridgeStatus === 'running' ? 'bg-makoclaw-success shadow-[0_0_8px_rgba(var(--pc-success),0.5)]' : 'bg-makoclaw-text-secondary/40'"
            />
            <span
              class="text-[10px] font-bold uppercase tracking-wider"
              :class="devStore.bridgeStatus === 'running' ? 'text-makoclaw-success' : 'text-makoclaw-text-secondary'"
            >
              Bridge: {{ devStore.bridgeStatus }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Content: Sidebar + Main -->
    <div class="flex-1 flex overflow-hidden relative z-10">
      <!-- Sidebar (Projects) -->
      <div
        :class="[
          'flex-shrink-0 border-r border-makoclaw-border/20 bg-makoclaw-surface/30 backdrop-blur-sm transition-all duration-500 ease-[cubic-bezier(0.4,0,0.2,1)] flex flex-col ring-1 ring-white/5',
          showSidebar ? 'w-72 opacity-100' : 'w-0 opacity-0 border-none overflow-hidden scale-95 origin-left'
        ]"
      >
        <div class="p-4 border-b border-makoclaw-border/30 bg-makoclaw-surface/30">
          <div class="flex justify-between items-center mb-4">
            <div class="flex items-center gap-2 capitalize">
              <i
                class="fas fa-project-diagram text-cyan-400 text-xs"
              />
              <span class="text-xs font-bold text-makoclaw-text tracking-wider">Projects</span>
            </div>
            <button
              class="p-1.5 hover:bg-cyan-500/10 rounded-lg text-makoclaw-text-secondary hover:text-cyan-400 transition-all active:scale-95 ring-1 ring-white/5"
              title="New Project"
              @click="showNewProjectForm = !showNewProjectForm"
            >
              <i
                class="fas"
                :class="showNewProjectForm ? 'fa-times' : 'fa-plus'"
              />
            </button>
          </div>

          <!-- Project Search -->
          <div class="relative group/search mb-4">
            <i class="fas fa-search absolute left-3 top-1/2 -translate-y-1/2 text-[10px] text-makoclaw-text-secondary group-focus-within/search:text-cyan-400 transition-colors" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search projects..."
              class="w-full pl-9 pr-4 py-1.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-xs outline-none focus:border-cyan-400/50 focus:ring-2 focus:ring-cyan-400/10 transition-all font-medium"
            >
          </div>

          <!-- New Project Form (Branded) -->
          <div
            v-if="showNewProjectForm"
            class="mb-4 glass-panel p-3 rounded-xl animate-fadeIn animate-slideUp"
          >
            <input
              v-model="newProjectName"
              type="text"
              class="input-field mb-2"
              placeholder="New project name..."
              @keyup.enter="handleCreateProject"
            >
            <div class="flex items-center gap-2 mb-3">
              <input
                id="gitInit"
                v-model="newProjectGitInit"
                type="checkbox"
                class="rounded border-makoclaw-border bg-makoclaw-bg text-cyan-500 focus:ring-cyan-500/30"
              >
              <label
                for="gitInit"
                class="text-[10px] text-makoclaw-text-secondary font-medium cursor-pointer"
              >Initialize git repository</label>
            </div>
            <button
              class="btn-primary w-full py-1.5 text-xs bg-cyan-600 hover:bg-cyan-500"
              :disabled="!newProjectName.trim() || creatingProject"
              @click="handleCreateProject"
            >
              {{ creatingProject ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </div>

        <!-- Project List -->
        <div class="flex-1 overflow-y-auto p-2 space-y-1 custom-scrollbar">
          <div
            v-for="p in filteredProjects"
            :key="p.path"
            class="relative group/proj list-item-interactive overflow-hidden mb-0.5"
            :class="devStore.currentProject === p.path ? 'bg-makoclaw-surface border-makoclaw-border shadow-sm' : ''"
          >
            <!-- Hover effects -->
            <div class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-0 bg-gradient-to-b from-transparent via-cyan-400/40 to-transparent group-hover/proj:h-2/3 transition-all duration-300 pointer-events-none" />
            <div class="absolute bottom-0 left-0 h-[1.5px] w-0 bg-gradient-to-r from-cyan-400 to-blue-500 group-hover/proj:w-full transition-all duration-500 opacity-30 pointer-events-none" />

            <button
              class="w-full text-left px-3 py-2.5 rounded-xl text-xs transition-all duration-200"
              @click="devStore.fetchHistory(p.name); devStore.startBridge(p.path)"
            >
              <div class="flex items-center gap-3">
                <div
                  class="w-6 h-6 rounded-lg flex items-center justify-center flex-shrink-0"
                  :class="devStore.currentProject === p.path ? 'bg-cyan-500/20' : 'bg-makoclaw-surface/50 border border-makoclaw-border/50'"
                >
                  <i
                    class="fas fa-folder text-[10px]"
                    :class="devStore.currentProject === p.path ? 'text-cyan-400' : 'text-makoclaw-text-secondary'"
                  />
                </div>
                <div class="flex-1 min-w-0">
                  <div
                    class="truncate font-bold tracking-tight"
                    :class="devStore.currentProject === p.path ? 'text-makoclaw-text' : 'text-makoclaw-text-secondary group-hover/proj:text-makoclaw-text'"
                  >
                    {{ p.name }}
                  </div>
                  <div
                    v-if="p.modTime"
                    class="text-[9px] opacity-40 mt-0.5"
                  >
                    Last modified: {{ new Date(p.modTime).toLocaleDateString() }}
                  </div>
                </div>
              </div>
            </button>
          </div>

          <div
            v-if="filteredProjects.length === 0"
            class="py-8 text-center"
          >
            <i class="fas fa-search text-makoclaw-text-secondary/20 text-2xl mb-2" />
            <p class="text-[10px] text-makoclaw-text-secondary/60">
              {{ devStore.projects.length === 0 ? 'No repos found' : 'No matches found' }}
            </p>
          </div>
        </div>

        <div class="p-3 border-t border-makoclaw-border/20 bg-makoclaw-surface/20">
          <div
            v-if="devStore.projectsRoot"
            class="mb-3"
          >
            <div class="text-[9px] text-makoclaw-text-secondary/50 font-bold tracking-widest uppercase mb-1.5 flex items-center gap-1.5">
              <i class="fas fa-map-marker-alt text-[8px]" /> Root Path
            </div>
            <div class="text-[9px] text-makoclaw-text-secondary/60 break-all font-mono bg-makoclaw-bg/50 px-2 py-1.5 rounded-lg border border-makoclaw-border/30">
              {{ devStore.projectsRoot }}
            </div>
          </div>
          <button
            class="w-full py-1.5 rounded-lg text-[10px] font-bold text-makoclaw-text-secondary hover:text-cyan-400 hover:bg-cyan-500/5 transition-all flex items-center justify-center gap-1.5 bg-makoclaw-surface/30 border border-makoclaw-border/30"
            @click="devStore.fetchProjects()"
          >
            <i class="fas fa-sync" /> REFRESH LIST
          </button>
        </div>
      </div>

      <!-- Main Workspace -->
      <div class="flex-1 flex flex-col min-w-0 relative">
        <div class="glass-sticky top-0 z-10 border-b border-makoclaw-border/20 px-4 py-2 flex items-center gap-4">
          <div class="flex gap-1">
            <button
              v-for="tab in ['Terminal', 'Memory']"
              :key="tab"
              class="tab-button"
              :class="activeTab === tab ? 'tab-button-active' : 'tab-button-inactive'"
              @click="activeTab = tab"
            >
              <i
                class="fas mr-1.5 text-[10px]"
                :class="tab === 'Terminal' ? 'fa-terminal' : 'fa-brain'"
              />
              {{ tab }}
            </button>
          </div>

          <div
            v-if="devStore.currentProject"
            class="hidden md:flex ml-auto items-center gap-2"
          >
            <i class="fas fa-chevron-right text-[10px] text-makoclaw-text-secondary/30" />
            <span class="text-[10px] font-bold text-cyan-400/70 tracking-wider font-mono">
              {{ devStore.currentProject.split(/[\\/]/).pop() }}
            </span>
          </div>
        </div>

        <div class="flex-1 min-h-0 flex flex-col">
          <!-- Terminal View -->
          <div
            v-show="activeTab === 'Terminal'"
            class="flex-1 flex flex-col min-h-0 bg-black/90 p-4 font-mono text-sm relative"
          >
            <div
              ref="terminalRef"
              class="flex-1 overflow-y-auto pb-4 space-y-2 custom-scrollbar"
            >
              <div
                v-if="devStore.terminalHistory.length === 0"
                class="text-makoclaw-text-secondary/30 italic space-y-1.5"
              >
                <p v-if="devStore.bridgeStatus === 'running'">
                  > Bridge established. Ready for instructions.
                </p>
                <p v-else>
                  > No active bridge. Select a repository to begin.
                </p>
              </div>

              <div
                v-for="(msg, idx) in devStore.terminalHistory"
                :key="idx"
                class="break-words"
              >
                <template v-if="msg.type === 'user'">
                  <span class="text-cyan-400 font-bold">$ </span>
                  <span class="text-white">{{ msg.message }}</span>
                </template>
                <template v-else-if="msg.type === 'error'">
                  <span class="text-makoclaw-error font-bold">[!] ERROR</span>
                  <div class="ml-4 text-makoclaw-error opacity-80 mt-1">
                    {{ msg.message || msg.error }}
                  </div>
                </template>
                <template v-else-if="msg.type === 'assistant'">
                  <div class="text-slate-300 whitespace-pre-wrap pl-4 border-l-2 border-makoclaw-border/30">
                    {{ msg.message }}
                  </div>
                </template>
                <template v-else-if="msg.type === 'tool_use'">
                  <div class="flex items-center gap-2 text-amber-500 py-1">
                    <i class="fas fa-cog animate-spin-slow" />
                    <span class="text-[10px] font-bold uppercase tracking-wider">Executing: {{ msg.name }}</span>
                  </div>
                </template>
                <template v-else-if="msg.type === 'tool_result'">
                  <div class="text-makoclaw-text-secondary/60 whitespace-pre-wrap pl-6 text-xs bg-white/5 py-2 px-3 rounded-lg border border-white/5 my-1">
                    {{ msg.message }}
                  </div>
                </template>
                <template v-else-if="msg.type === 'result'">
                  <div class="flex items-center gap-2 text-makoclaw-success py-2">
                    <i class="fas fa-check-circle" />
                    <span class="font-bold">Execution complete</span>
                    <span
                      v-if="msg.cost_usd"
                      class="text-[9px] text-makoclaw-text-secondary opacity-50 font-mono"
                    >
                      ({{ (msg.duration_ms / 1000).toFixed(1) }}s)
                    </span>
                  </div>
                </template>
              </div>
            </div>

            <!-- Terminal Input -->
            <div class="mt-4 pt-4 border-t border-makoclaw-border/10 shrink-0">
              <form
                class="flex relative items-center"
                @submit.prevent="submitPrompt"
              >
                <span
                  class="absolute left-3 font-bold text-xs"
                  :class="devStore.bridgeStatus === 'running' ? 'text-cyan-500' : 'text-makoclaw-text-secondary/40'"
                >$</span>
                <input
                  ref="inputRef"
                  v-model="prompt"
                  type="text"
                  class="w-full bg-makoclaw-bg/30 border rounded-xl py-2.5 pl-8 pr-4 text-cyan-50 focus:outline-none focus:border-cyan-500/50 focus:ring-2 focus:ring-cyan-500/10 transition-all font-mono text-xs"
                  :class="devStore.bridgeStatus === 'running' ? 'border-makoclaw-border/50' : 'border-makoclaw-border/20'"
                  :placeholder="devStore.bridgeStatus === 'running' ? 'Enter command or ask assistant...' : 'Select a project to enable input...'"
                >
              </form>
            </div>
          </div>

          <!-- Memory View -->
          <div
            v-show="activeTab === 'Memory'"
            class="flex-1 p-6 overflow-y-auto animate-fadeIn"
          >
            <div class="max-w-4xl mx-auto">
              <div class="flex items-center gap-3 mb-8">
                <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-purple-500/20 to-blue-500/20 flex items-center justify-center ring-1 ring-white/10">
                  <i class="fas fa-brain text-xl text-purple-400" />
                </div>
                <div>
                  <h3 class="text-xl font-bold text-makoclaw-text tracking-tight">
                    Semantic Memory
                  </h3>
                  <p class="text-xs text-makoclaw-text-secondary">
                    Retrieve historical context and architecture decisions
                  </p>
                </div>
              </div>

              <div class="flex gap-2 mb-8 glass-panel p-2 rounded-2xl shadow-lg ring-1 ring-white/5">
                <input
                  v-model="memoryQuery"
                  type="text"
                  class="flex-1 bg-transparent border-none px-4 py-2 text-sm text-makoclaw-text focus:ring-0 outline-none placeholder:text-makoclaw-text-secondary/40"
                  placeholder="Query engram or project context..."
                  @keyup.enter="devStore.searchMemory(memoryQuery)"
                >
                <button
                  class="btn-primary flex items-center gap-2 bg-gradient-to-r from-purple-600 to-blue-600 border-none px-6"
                  @click="devStore.searchMemory(memoryQuery)"
                >
                  <i class="fas fa-search text-xs" /> SEARCH
                </button>
              </div>

              <div class="space-y-4">
                <div
                  v-for="result in devStore.searchResults"
                  :key="result.id"
                  class="glass-panel p-5 rounded-2xl border-makoclaw-border/10 hover:border-purple-500/30 transition-colors group"
                >
                  <div class="flex justify-between items-center mb-3">
                    <div class="text-[10px] font-bold text-purple-400 uppercase tracking-widest flex items-center gap-2">
                      <i class="far fa-calendar-alt" />
                      {{ new Date(result.created_at).toLocaleString() }}
                    </div>
                  </div>
                  <div class="text-sm text-makoclaw-text-secondary leading-relaxed whitespace-pre-wrap group-hover:text-makoclaw-text transition-colors">
                    {{ result.content }}
                  </div>
                </div>

                <div
                  v-if="devStore.searchResults.length === 0"
                  class="empty-state py-16"
                >
                  <div class="w-16 h-16 rounded-3xl bg-makoclaw-surface flex items-center justify-center mb-4 ring-1 ring-white/5 shadow-inner">
                    <i class="fas fa-fingerprint text-2xl text-makoclaw-text-secondary/20" />
                  </div>
                  <div class="empty-state-title">
                    No memories found
                  </div>
                  <div class="empty-state-text">
                    Query the semantic archive to surface historical project context.
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, watch, computed } from 'vue'
import { useDevStudioStore } from '@/stores/devStudioStore'

const devStore = useDevStudioStore()
const activeTab = ref('Terminal')
const prompt = ref('')
const memoryQuery = ref('')
const terminalRef = ref(null)
const inputRef = ref(null)
const searchQuery = ref('')
const showSidebar = ref(localStorage.getItem('dev_studio_show_sidebar') !== 'false')

const filteredProjects = computed(() => {
  if (!searchQuery.value) return devStore.projects
  const q = searchQuery.value.toLowerCase()
  return devStore.projects.filter(p => p.name.toLowerCase().includes(q))
})

const toggleSidebar = () => {
  showSidebar.value = !showSidebar.value
  localStorage.setItem('dev_studio_show_sidebar', showSidebar.value)
}

const showNewProjectForm = ref(false)
const newProjectName = ref('')
const newProjectGitInit = ref(true)
const creatingProject = ref(false)

onMounted(() => {
  devStore.fetchProjects()
  devStore.checkStatus()
  // Auto-focus the terminal input
  nextTick(() => {
    if (inputRef.value) inputRef.value.focus()
  })
})

const handleCreateProject = async () => {
  if (!newProjectName.value.trim() || creatingProject.value) return
  creatingProject.value = true
  try {
    await devStore.createProject(newProjectName.value, newProjectGitInit.value)
    newProjectName.value = ''
    showNewProjectForm.value = false
  } catch (e) {
    devStore.terminalHistory.push({ type: 'error', message: `Failed to create project: ${e.message}` })
  } finally {
    creatingProject.value = false
  }
}

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
