<template>
  <div class="space-y-8 max-w-4xl mx-auto animate-fade-in-up">
    <div
      v-for="(info, name) in providers"
      :key="name"
      class="glass-panel rounded-[2rem] p-8 md:p-10 border border-makoclaw-border/50 relative overflow-hidden group hover:border-makoclaw-accent/30 transition-all duration-500"
    >
      <!-- Decorative background blur -->
      <div class="absolute -top-12 -right-12 w-48 h-48 bg-makoclaw-accent/5 blur-[60px] rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-1000"></div>

      <div class="flex items-center justify-between mb-8 relative z-10">
        <div class="flex items-center gap-4">
          <div class="w-14 h-14 rounded-2xl bg-makoclaw-surface border-2 border-makoclaw-border/50 flex items-center justify-center text-makoclaw-accent shadow-xl group-hover:scale-110 group-hover:rotate-3 transition-all duration-500">
             <span class="text-xs font-black uppercase tracking-tighter">{{ name.substring(0,2) }}</span>
          </div>
          <div>
            <h3 class="text-xl font-black capitalize text-makoclaw-text tracking-tight italic">{{ name }}</h3>
            <div class="flex items-center gap-2 mt-1">
               <span 
                class="px-2 py-0.5 text-[9px] font-black uppercase tracking-widest rounded-lg border flex items-center gap-1.5"
                :class="isProviderConfigured(info) ? 'bg-makoclaw-success/10 text-makoclaw-success border-makoclaw-success/20' : 'bg-makoclaw-text-secondary/10 text-makoclaw-text-secondary border-makoclaw-border/50'"
               >
                 <span :class="`w-1.5 h-1.5 rounded-full ${isProviderConfigured(info) ? 'bg-makoclaw-success animate-pulse' : 'bg-makoclaw-text-secondary/40'}`"></span>
                 {{ isProviderConfigured(info) ? 'Linked' : 'Offline' }}
               </span>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <button 
            @click="openModelsModal(name, info)"
            class="p-3 rounded-2xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all hover:scale-110 active:scale-95 group/btn"
            title="Configure Models"
          >
            <IconModels class="w-5 h-5 group-hover/btn:rotate-12 transition-transform" />
          </button>
        </div>
      </div>
      
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-8 items-end relative z-10">
        <div class="lg:col-span-12 xl:col-span-5 space-y-2">
          <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Access Protocol (API Key)</label>
          <div class="relative group/input">
            <input 
              v-model="info.api_key" 
              type="password" 
              :placeholder="isProviderConfigured(info) ? '••••••••••••••••••••••••' : 'Enter access key...'" 
              class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-6 py-4 text-sm font-bold outline-none focus:border-makoclaw-accent text-makoclaw-text backdrop-blur-sm transition-all"
            >
            <div class="absolute right-4 top-1/2 -translate-y-1/2 text-makoclaw-text-secondary/20 group-hover/input:text-makoclaw-accent/40 transition-colors">
              <IconLock class="w-4 h-4" />
            </div>
          </div>
        </div>
        <div class="lg:col-span-12 xl:col-span-4 space-y-2">
          <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Gateway Entry (Base URL)</label>
          <input 
            v-model="info.api_base" 
            type="text" 
            placeholder="https://api.neural-matrix..." 
            class="w-full bg-makoclaw-bg/20 border-2 border-makoclaw-border/30 rounded-2xl px-6 py-4 text-sm font-bold outline-none focus:border-makoclaw-accent text-makoclaw-text transition-all"
          >
        </div>
        <div class="lg:col-span-12 xl:col-span-3">
          <button 
            @click="$emit('save', {providers: {[name]: {api_key: info.api_key, api_base: info.api_base, models: Array.isArray(info.models) ? info.models : []}}})" 
            :disabled="saving"
            class="w-full px-6 py-4 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-2xl font-black uppercase tracking-widest shadow-xl shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-3 active:scale-95 disabled:opacity-50 group/save"
          >
            <span v-if="saving" class="w-4 h-4 border-3 border-white/20 border-t-white rounded-full animate-spin"></span>
            <IconSave v-else class="w-4 h-4 group-hover/save:translate-y-[-2px] transition-transform" />
            Connect
          </button>
        </div>
      </div>
    </div>

    <!-- Models Config Modal -->
    <Teleport to="body">
      <Transition name="modal">
      <div v-if="showModelsModal" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-xl p-4" @click.self="showModelsModal = false">
        <div class="glass-panel rounded-[2.5rem] p-10 max-w-xl w-full shadow-2xl border border-makoclaw-border/50 relative overflow-hidden animate-zoom flex flex-col max-h-[85vh]" @click.stop>
          <div class="absolute -top-12 -left-12 w-48 h-48 bg-makoclaw-accent/10 blur-[60px] rounded-full pointer-events-none"></div>

          <div class="flex justify-between items-start mb-10 relative z-10">
            <div>
              <h3 class="text-2xl font-black bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent italic capitalize">
                {{ selectedProviderName }} Cores
              </h3>
              <p class="text-[10px] font-bold text-makoclaw-text-secondary uppercase tracking-[0.2em] mt-1">Configure available neural models for this network</p>
            </div>
            <button @click="showModelsModal = false" class="p-2 rounded-xl bg-makoclaw-bg/60 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-white transition-all">
              <IconClose class="w-5 h-5" />
            </button>
          </div>
          
          <div class="space-y-8 overflow-y-auto pr-4 custom-scrollbar relative z-10 h-full flex flex-col">
             <!-- Add Model -->
             <div class="space-y-3">
                <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Enlist New Core Model</label>
                <div class="flex gap-3">
                  <input 
                    v-model="newModelInput" 
                    @keyup.enter="addModel"
                    type="text" 
                    placeholder="e.g. gpt-4o, claude-3-5-sonnet" 
                    class="flex-1 px-6 py-4 bg-makoclaw-bg/60 border-2 border-makoclaw-border/50 rounded-2xl text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none"
                  >
                  <button 
                    @click="addModel" 
                    :disabled="!newModelInput.trim()"
                    class="px-8 py-4 bg-white text-makoclaw-bg rounded-2xl font-black uppercase tracking-widest text-[11px] transition-all hover:bg-white/90 active:scale-95 disabled:opacity-30"
                  >
                    Add
                  </button>
                </div>
             </div>

             <!-- Model List -->
             <div class="flex-1 min-h-0 flex flex-col pt-4 border-t border-makoclaw-border/30">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="text-[10px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/60">Active Core Pool</h4>
                  <span class="px-2 py-0.5 bg-makoclaw-accent/10 text-makoclaw-accent text-[9px] font-black rounded-lg border border-makoclaw-accent/20">{{ editingModels.length }} models</span>
                </div>
                
                <div v-if="editingModels.length === 0" class="flex-1 flex flex-col items-center justify-center py-12 bg-makoclaw-bg/20 rounded-[2rem] border-2 border-dashed border-makoclaw-border/30">
                  <IconModels class="w-12 h-12 text-makoclaw-text-secondary/10 mb-4" />
                  <p class="text-sm font-bold text-makoclaw-text-secondary/40 italic">No custom cores configured.</p>
                  <p class="text-[9px] font-medium text-makoclaw-text-secondary/30 mt-1 uppercase tracking-widest italic">Default parameters will be utilized</p>
                </div>
                
                <div v-else class="grid grid-cols-1 gap-3 flex-1 overflow-y-auto pr-2 custom-scrollbar">
                  <div v-for="(model, idx) in editingModels" :key="idx" class="flex items-center justify-between bg-makoclaw-surface/40 border border-makoclaw-border/50 px-5 py-3.5 rounded-2xl group hover:border-makoclaw-accent/30 transition-all">
                    <div class="flex items-center gap-3">
                       <span class="w-2 h-2 rounded-full bg-makoclaw-accent/30 group-hover:bg-makoclaw-accent group-hover:shadow-[0_0_8px_rgba(var(--makoclaw-accent-rgb),0.5)] transition-all"></span>
                       <span class="text-sm font-bold text-makoclaw-text tracking-tight">{{ model }}</span>
                    </div>
                    <button @click="removeModel(idx)" class="p-2 text-makoclaw-text-secondary/40 hover:text-red-400 hover:bg-red-400/10 rounded-xl transition-all opacity-0 group-hover:opacity-100 scale-90 group-hover:scale-100">
                      <IconDelete class="w-4 h-4" />
                    </button>
                  </div>
                </div>
             </div>
          </div>

          <div class="flex items-center justify-between mt-10 pt-8 border-t border-makoclaw-border/30 relative z-10">
            <button @click="resetToDefaults" class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40 hover:text-red-400 transition-colors flex items-center gap-2 group">
              <IconReset class="w-3.5 h-3.5 group-hover:rotate-[-45deg] transition-transform" />
              Reset Factory Defaults
            </button>
            <div class="flex gap-4">
              <button @click="showModelsModal = false" class="px-6 py-3 text-[11px] font-black uppercase tracking-widest text-makoclaw-text-secondary hover:text-white transition-colors">Cancel</button>
              <button @click="saveModelsConfig" :disabled="saving" class="px-10 py-4 bg-makoclaw-accent text-white rounded-2xl font-black uppercase tracking-widest shadow-xl shadow-makoclaw-accent/20 hover:bg-makoclaw-accent-hover transition-all active:scale-95 disabled:opacity-30 flex items-center">
                <span v-if="saving" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin mr-3"></span>
                Commit Pool
              </button>
            </div>
          </div>
        </div>
      </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, h } from 'vue'

const props = defineProps({
  providers: { type: Object, required: true },
  providersList: { type: Array, default: () => [] },
  saving: { type: Boolean, default: false }
})
const emit = defineEmits(['save'])

// Premium Icons
const IconModels = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z' })]) }
const IconLock = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 00-2 2zm10-10V7a4 4 0 00-8 0v4h8z' })]) }
const IconSave = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M13 10V3L4 14h7v7l9-11h-7z' })]) }
const IconClose = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M6 18L18 6M6 6l12 12' })]) }
const IconDelete = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16' })]) }
const IconReset = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15' })]) }

const showModelsModal = ref(false)
const selectedProviderName = ref('')
const newModelInput = ref('')
const editingModels = ref([])

const isProviderConfigured = (info) => {
  if (!info || typeof info !== 'object') return false
  if (info.configured === true) return true
  const key = typeof info.api_key === 'string' ? info.api_key.trim() : ''
  return key.length > 0 && !key.includes('••••')
}

const openModelsModal = (name, info) => {
  selectedProviderName.value = name
  if (info.models && info.models.length > 0) {
    editingModels.value = [...info.models]
  } else {
    const pList = Array.isArray(props.providersList) ? props.providersList : []
    const apiP = pList.find(p => p.name === name)
    editingModels.value = apiP?.models ? apiP.models.map(m => m.id) : []
  }
  newModelInput.value = ''
  showModelsModal.value = true
}

const addModel = () => {
  const m = newModelInput.value.trim()
  if (m && !editingModels.value.includes(m)) {
    editingModels.value.push(m)
    newModelInput.value = ''
  }
}

const removeModel = (idx) => editingModels.value.splice(idx, 1)
const resetToDefaults = () => { if (confirm('Purge all custom core profiles? Factory defaults will be restored.')) editingModels.value = [] }

const saveModelsConfig = () => {
  emit('save', { providers: { [selectedProviderName.value]: { models: editingModels.value } } })
  showModelsModal.value = false
}
</script>

<style scoped>
@keyframes fade-in-up {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
.animate-fade-in-up { animation: fade-in-up 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
.animate-zoom { animation: zoom 0.3s cubic-bezier(0.16, 1, 0.3, 1); }
@keyframes zoom { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }
.modal-enter-active, .modal-leave-active { transition: opacity 0.3s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(var(--makoclaw-accent-rgb), 0.2); border-radius: 10px; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background: rgba(var(--makoclaw-accent-rgb), 0.4); }
</style>
