<template>
  <div class="space-y-5 animate-fade-in-up">
    <!-- Web & Gateway -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
      <div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50 transition-all duration-300 hover:border-makoclaw-accent/30 hover:shadow-lg">
        <div class="flex items-center gap-3 mb-5">
          <div class="p-2 rounded-xl bg-blue-500/10 text-blue-400">
            <GlobeAltIcon class="w-5 h-5" />
          </div>
          <h3 class="text-xs font-medium tracking-wide text-makoclaw-text-secondary/70 uppercase">
            Web Infrastructure
          </h3>
        </div>
        <div class="space-y-4">
          <div
            v-for="(val, key) in configData?.web"
            :key="key"
            class="flex justify-between items-center px-4 py-3 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30"
          >
            <span class="text-xs font-bold text-makoclaw-text-secondary/80">{{ formatKey(key) }}</span>
            <span class="text-xs font-black font-mono text-makoclaw-accent">{{ String(val) }}</span>
          </div>
        </div>
      </div>
      
      <div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50 transition-all duration-300 hover:border-makoclaw-accent/30 hover:shadow-lg">
        <div class="flex items-center gap-3 mb-5">
          <div class="p-2 rounded-xl bg-makoclaw-accent/10 text-makoclaw-accent">
            <ServerIcon class="w-5 h-5" />
          </div>
          <h3 class="text-xs font-medium tracking-wide text-makoclaw-text-secondary/70 uppercase">
            Gateway Layer
          </h3>
        </div>
        <div class="space-y-4">
          <div
            v-for="(val, key) in configData?.gateway"
            :key="key"
            class="flex justify-between items-center px-4 py-3 rounded-2xl bg-makoclaw-bg/30 border border-makoclaw-border/30"
          >
            <span class="text-xs font-bold text-makoclaw-text-secondary/80">{{ formatKey(key) }}</span>
            <span class="text-xs font-black font-mono text-makoclaw-accent">{{ String(val) }}</span>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Utilities -->
    <div class="glass-panel rounded-2xl p-5 border border-makoclaw-border/50 transition-all duration-300 hover:shadow-lg">
      <div class="flex items-center justify-between mb-5">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-xl bg-orange-500/10 text-orange-400">
            <WrenchScrewdriverIcon class="w-5 h-5" />
          </div>
          <h3 class="text-xs font-medium tracking-wide text-makoclaw-text-secondary/70 uppercase">
            Utility Parameters
          </h3>
        </div>
        <button
          :disabled="saving"
          class="px-4 py-1.5 rounded-xl bg-makoclaw-accent/10 text-makoclaw-accent border border-makoclaw-accent/20 text-[10px] font-medium tracking-wide hover:bg-makoclaw-accent hover:text-white transition-all active:scale-95"
          @click="$emit('save', { tools: configData?.tools })"
        >
          {{ saving ? 'Syncing...' : 'Sync Parameters' }}
        </button>
      </div>
      <div
        v-if="configData?.tools?.web?.search"
        class="grid grid-cols-1 sm:grid-cols-2 gap-5"
      >
        <div class="space-y-2">
          <label class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/60">Search Access Key</label>
          <input
            v-model="configData.tools.web.search.api_key"
            type="password"
            placeholder="••••••••••••••••"
            class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent transition-all outline-none"
          >
        </div>
        <div class="space-y-2">
          <label class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/60">Search Capacity</label>
          <input
            v-model.number="configData.tools.web.search.max_results"
            type="number"
            class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent transition-all outline-none"
          >
        </div>
      </div>
    </div>

    <!-- Backup & Recovery SECTION OVERHAUL -->
    <div class="glass-panel rounded-2xl p-5 border border-makoclaw-accent/20 bg-makoclaw-accent/5">
      <div class="flex items-center gap-3 mb-5">
        <div class="p-3 rounded-2xl bg-makoclaw-accent text-white shadow-lg shadow-makoclaw-accent/20">
          <ArchiveBoxArrowDownIcon class="w-6 h-6" />
        </div>
        <div>
          <h3 class="text-lg font-black text-makoclaw-text uppercase tracking-tight italic">
            Snapshot Core
          </h3>
          <p class="text-xs font-medium text-makoclaw-text-secondary opacity-60">
            Complete system backup and restoration
          </p>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-10">
        <!-- Export -->
        <div class="space-y-6">
          <h4 class="text-[11px] font-medium tracking-wide text-makoclaw-text flex items-center gap-2">
            <span class="w-1.5 h-1.5 rounded-full bg-blue-500 animate-pulse" />
            Generate Snapshot
          </h4>
          <div class="p-6 bg-makoclaw-bg/40 rounded-3xl border border-makoclaw-border/50 space-y-3 shadow-inner">
            <label
              v-for="(val, opt) in exportOptions"
              :key="opt"
              class="flex items-center justify-between p-3 rounded-xl hover:bg-white/5 cursor-pointer transition-colors group"
            >
              <span class="text-xs font-black text-makoclaw-text-secondary/80 uppercase tracking-tighter group-hover:text-makoclaw-text">{{ opt.replace('include_', '').replace('_', ' ') }}</span>
              <input
                v-model="exportOptions[opt]"
                type="checkbox"
                class="w-5 h-5 rounded-lg border-2 border-makoclaw-border bg-makoclaw-bg text-makoclaw-accent focus:ring-makoclaw-accent transition-colors"
              >
            </label>
          </div>
          <button
            :disabled="exporting || Object.values(exportOptions).every(v => !v)"
            class="w-full bg-blue-600 hover:bg-blue-500 text-white py-2.5 rounded-xl font-semibold shadow-lg shadow-blue-500/20 transition-all flex items-center justify-center disabled:opacity-30 active:scale-95 group"
            @click="exportBackup"
          >
            <ArrowDownTrayIcon
              v-if="!exporting"
              class="w-5 h-5 mr-3 transition-transform group-hover:-translate-y-1"
            />
            <span
              v-else
              class="w-5 h-5 border-4 border-white/20 border-t-white rounded-full animate-spin mr-3"
            />
            {{ exporting ? 'Packing System...' : 'Download (.makoclaw)' }}
          </button>
        </div>

        <!-- Import -->
        <div class="space-y-6">
          <h4 class="text-[11px] font-medium tracking-wide text-makoclaw-text flex items-center gap-2">
            <span class="w-1.5 h-1.5 rounded-full bg-orange-500 animate-pulse" />
            Inject Snapshot
          </h4>
          <div class="border-2 border-dashed border-makoclaw-border/50 rounded-3xl p-6 hover:border-makoclaw-accent/50 transition-all group relative bg-makoclaw-bg/20">
            <input
              ref="fileInput"
              type="file"
              accept=".makoclaw"
              class="hidden"
              @change="handleFileSelect"
            >
            <div
              v-if="!selectedFile"
              class="flex flex-col items-center justify-center py-4 cursor-pointer"
              @click="$refs.fileInput.click()"
            >
              <CloudArrowUpIcon class="w-12 h-12 text-makoclaw-text-secondary/30 group-hover:text-makoclaw-accent transition-colors duration-500 group-hover:-translate-y-2" />
              <p class="text-xs font-medium mt-3 text-makoclaw-text-secondary/50 group-hover:text-makoclaw-text-secondary">
                Load .makoclaw Module
              </p>
            </div>
            <div
              v-else
              class="space-y-4"
            >
              <div class="flex items-center gap-4 p-4 bg-makoclaw-bg/60 rounded-2xl border border-makoclaw-border/50 shadow-inner">
                <div class="p-2 bg-makoclaw-accent/10 rounded-xl text-makoclaw-accent">
                  <DocumentIcon class="w-6 h-6" />
                </div>
                <div class="flex-1 min-w-0">
                  <div class="text-[11px] font-black text-makoclaw-text truncate">
                    {{ selectedFile.name }}
                  </div>
                  <div class="text-[9px] font-bold text-makoclaw-text-secondary opacity-50">
                    {{ formatBytes(selectedFile.size) }}
                  </div>
                </div>
                <button
                  class="p-2 text-makoclaw-text-secondary hover:text-red-400 transition-colors"
                  @click="clearSelectedFile"
                >
                  <XMarkIcon class="w-4 h-4" />
                </button>
              </div>
              <button
                :disabled="importing || !validationResult?.has_any_content"
                class="w-full bg-orange-600 hover:bg-orange-500 text-white py-2.5 rounded-xl font-semibold shadow-lg shadow-orange-500/20 transition-all flex items-center justify-center disabled:opacity-30 active:scale-95 group"
                @click="importBackup"
              >
                <BoltIcon
                  v-if="!importing"
                  class="w-5 h-5 mr-3 transition-transform group-hover:scale-110"
                />
                <span
                  v-else
                  class="w-5 h-5 border-4 border-white/20 border-t-white rounded-full animate-spin mr-3"
                />
                {{ importing ? 'Injecting Systems...' : 'Execute Injection' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../../stores/authStore'
import { useToast } from '../../composables/useToast'
import { 
  GlobeAltIcon, ServerIcon, WrenchScrewdriverIcon, ArchiveBoxArrowDownIcon,
  ArrowDownTrayIcon, CloudArrowUpIcon, DocumentIcon, XMarkIcon, BoltIcon
} from '@heroicons/vue/24/outline'

defineProps({
  configData: { type: Object, default: () => ({}) },
  saving: { type: Boolean, default: false }
})

defineEmits(['save'])
const authStore = useAuthStore()
const toast = useToast()

const formatKey = (k) => k.replace(/_/g, ' ').toUpperCase()
const formatBytes = (b) => (b / 1024 / 1024).toFixed(2) + ' MB'

// Backup logic
const exporting = ref(false)
const exportOptions = ref({ include_database: true, include_workspace: true })
const selectedFile = ref(null)
const importing = ref(false)
const validationResult = ref(null)
const fileInput = ref(null)

const exportBackup = async () => {
   exporting.value = true
   try {
     const params = new URLSearchParams(exportOptions.value)
     const res = await fetch(`/api/v1/backup/export?${params}`, { 
       headers: { 'Authorization': `Bearer ${authStore.token}` } 
     })
     if(!res.ok) throw new Error('Export failed')
     const blob = await res.blob()
     const url = URL.createObjectURL(blob)
     const a = document.createElement('a')
     a.href = url; a.download = 'makoclaw_backup.makoclaw'; a.click()
     toast.success('Backup sequence completed')
   } catch(e) { 
     toast.error('Backup aborted: ' + e.message) 
   } finally { 
     exporting.value = false 
   }
}

const handleFileSelect = async (e) => {
  const file = e.target.files[0]
  if (!file) return
  selectedFile.value = file
  try {
    const fd = new FormData(); fd.append('file', file)
    const res = await fetch('/api/v1/backup/validate', { 
      method: 'POST', 
      headers: { 'Authorization': `Bearer ${authStore.token}` }, 
      body: fd 
    })
    if(!res.ok) throw new Error('Validation failed')
    validationResult.value = await res.json()
  } catch(e) { 
    toast.error('Validation failed: ' + e.message) 
  }
}

const importBackup = async () => {
  importing.value = true
  try {
    const fd = new FormData(); 
    fd.append('file', selectedFile.value); 
    fd.append('options', JSON.stringify({ replace_database: true, replace_workspace: true }))
    
    const res = await fetch('/api/v1/backup/import', { 
      method: 'POST', 
      headers: { 'Authorization': `Bearer ${authStore.token}` }, 
      body: fd 
    })
    if(!res.ok) throw new Error('Import failed')
    
    toast.success('Injection successful')
    setTimeout(() => {
      window.location.reload()
    }, 1000)
  } catch(e) { 
    toast.error('Injection failed: ' + e.message) 
  } finally { 
    importing.value = false 
  }
}

const clearSelectedFile = () => { 
  selectedFile.value = null
  validationResult.value = null 
}
</script>

<style scoped>
@keyframes fade-in-up {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
.animate-fade-in-up {
  animation: fade-in-up 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>
