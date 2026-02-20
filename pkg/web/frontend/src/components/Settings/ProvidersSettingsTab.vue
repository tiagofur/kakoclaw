<template>
  <div class="space-y-6 max-w-4xl mx-auto animate-fadeIn">
    <div
      v-for="(info, name) in providers"
      :key="name"
      class="glass-panel rounded-2xl p-6 transition-all hover:bg-white/5"
    >
      <div class="flex items-center justify-between mb-6">
        <h3 class="font-bold capitalize text-sm flex items-center text-kakoclaw-text tracking-wide">
           <span class="w-9 h-9 rounded-xl bg-kakoclaw-bg border border-kakoclaw-border flex items-center justify-center mr-3 text-kakoclaw-accent text-[10px] font-black shadow-inner">{{ name.substring(0,2).toUpperCase() }}</span>
           {{ name }}
        </h3>
        <div class="flex items-center space-x-3">
          <button 
            @click="openModelsModal(name, info)"
            class="p-1.5 rounded-lg text-kakoclaw-text-secondary hover:text-kakoclaw-accent hover:bg-kakoclaw-bg transition-colors"
            title="Configure Models"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </button>
          <span
            class="px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider rounded-full"
            :class="info.configured ? 'bg-emerald-500/10 text-emerald-400' : 'bg-gray-500/10 text-gray-400'"
          >{{ info.configured ? 'Configured' : 'Not configured' }}</span>
        </div>
      </div>
      
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-end">
        <div class="lg:col-span-5">
          <label class="block text-[10px] font-bold text-kakoclaw-text-secondary mb-2 uppercase tracking-widest opacity-70">API Key</label>
          <input 
            v-model="info.api_key" 
            type="password" 
            :placeholder="info.configured ? '••••••••••••••••' : 'Enter API Key'" 
            class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none focus:border-kakoclaw-accent text-kakoclaw-text backdrop-blur-sm transition-all"
          >
        </div>
        <div class="lg:col-span-4">
          <label class="block text-[10px] font-bold text-kakoclaw-text-secondary mb-2 uppercase tracking-widest opacity-70">API Base (optional)</label>
          <input 
            v-model="info.api_base" 
            type="text" 
            placeholder="https://api..." 
            class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none focus:border-kakoclaw-accent text-kakoclaw-text backdrop-blur-sm transition-all"
          >
        </div>
        <div class="lg:col-span-3">
          <button 
            @click="$emit('save', {providers: {[name]: {api_key: info.api_key, api_base: info.api_base}}})" 
            :disabled="saving"
            class="w-full bg-kakoclaw-accent text-white h-11 rounded-xl font-bold hover:bg-kakoclaw-accent-hover transition-all shadow-lg shadow-kakoclaw-accent/20 flex items-center justify-center disabled:opacity-50 active:scale-95"
          >
            <span v-if="saving" class="w-3 h-3 border-2 border-white/20 border-t-white rounded-full animate-spin mr-2"></span>
            Save {{ name }}
          </button>
        </div>
      </div>
    </div>

    <!-- Models Config Modal -->
    <div v-if="showModelsModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="bg-kakoclaw-surface rounded-2xl shadow-2xl w-full max-w-md border border-kakoclaw-border overflow-hidden animate-in fade-in zoom-in duration-200 flex flex-col max-h-[90vh]">
        <div class="flex justify-between items-center p-6 border-b border-kakoclaw-border bg-kakoclaw-bg/20">
          <h3 class="text-lg font-bold text-kakoclaw-text flex items-center capitalize">
            Configure {{ selectedProviderName }} Models
          </h3>
          <button @click="showModelsModal = false" class="text-kakoclaw-text-secondary hover:text-kakoclaw-text flex items-center justify-center w-8 h-8 rounded-full hover:bg-kakoclaw-bg transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <div class="p-6 space-y-5 overflow-y-auto flex-1 custom-scrollbar">
           <div>
              <label class="block text-xs font-bold text-kakoclaw-text-secondary mb-1.5 uppercase">Add New Model</label>
              <div class="flex space-x-2">
                <input 
                  v-model="newModelInput" 
                  @keyup.enter="addModel"
                  type="text" 
                  placeholder="e.g. gpt-4-turbo" 
                  class="flex-1 px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-xl text-sm outline-none focus:ring-2 focus:ring-kakoclaw-accent/20 focus:border-kakoclaw-accent text-kakoclaw-text"
                >
                <button 
                  @click="addModel" 
                  :disabled="!newModelInput.trim()"
                  class="px-4 py-2 bg-kakoclaw-bg border border-kakoclaw-border hover:border-kakoclaw-accent hover:text-kakoclaw-accent text-kakoclaw-text-secondary rounded-xl text-sm font-bold transition-all disabled:opacity-50"
                >
                  Add
                </button>
              </div>
           </div>

           <div>
              <label class="text-xs font-bold text-kakoclaw-text-secondary mb-3 uppercase flex justify-between items-center">
                <span>Configured Models</span>
                <span class="text-[10px] bg-kakoclaw-bg px-2 py-0.5 rounded text-kakoclaw-text-secondary">{{ editingModels.length }} models</span>
              </label>
              
              <div v-if="editingModels.length === 0" class="text-center py-8 text-sm text-kakoclaw-text-secondary bg-kakoclaw-bg/30 rounded-xl border border-dashed border-kakoclaw-border">
                No models configured.<br>
                <span class="text-xs">Using default models.</span>
              </div>
              
              <div v-else class="space-y-2">
                <div v-for="(model, idx) in editingModels" :key="idx" class="flex items-center justify-between bg-kakoclaw-bg border border-kakoclaw-border px-3 py-2.5 rounded-xl group hover:border-kakoclaw-accent/50 transition-colors">
                  <span class="text-sm text-kakoclaw-text font-mono truncate" :title="model">{{ model }}</span>
                  <button @click="removeModel(idx)" class="text-kakoclaw-text-secondary hover:text-red-400 p-1 opacity-50 group-hover:opacity-100 transition-all rounded hover:bg-red-400/10">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>
           </div>
        </div>

        <div class="flex justify-between items-center p-6 border-t border-kakoclaw-border bg-kakoclaw-bg/20">
          <button @click="resetToDefaults" class="text-xs font-bold text-kakoclaw-text-secondary hover:text-red-400 transition-colors">
            Reset to Defaults
          </button>
          <div class="flex space-x-3">
            <button @click="showModelsModal = false" class="px-4 py-2 text-sm font-medium text-kakoclaw-text-secondary hover:text-kakoclaw-text transition-colors">Cancel</button>
            <button @click="saveModelsConfig" :disabled="saving" class="px-6 py-2 text-sm font-bold bg-kakoclaw-accent text-white rounded-xl shadow-lg shadow-kakoclaw-accent/20 hover:bg-kakoclaw-accent-hover transition-all flex items-center disabled:opacity-50">
              <span v-if="saving" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin mr-2"></span>
              Save
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  providers: {
    type: Object,
    required: true
  },
  providersList: {
    type: Array,
    default: () => []
  },
  saving: {
    type: Boolean,
    default: false
  }
})
const emit = defineEmits(['save'])

const showModelsModal = ref(false)
const selectedProviderName = ref('')
const newModelInput = ref('')
const editingModels = ref([])

const openModelsModal = (name, info) => {
  selectedProviderName.value = name
  
  if (info.models && info.models.length > 0) {
    // Custom models exist
    editingModels.value = [...info.models]
  } else {
    // Check for defaults from API
    const apiProvider = props.providersList.find(p => p.name === name)
    if (apiProvider && apiProvider.models) {
       editingModels.value = apiProvider.models.map(m => m.id)
    } else {
       editingModels.value = []
    }
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

const removeModel = (idx) => {
  editingModels.value.splice(idx, 1)
}

const resetToDefaults = () => {
  if (confirm('Are you sure you want to clear custom models? The defaults will be used instead. (Click Save to apply)')) {
    editingModels.value = []
  }
}

const saveModelsConfig = () => {
  emit('save', {
    providers: {
      [selectedProviderName.value]: {
        models: editingModels.value
      }
    }
  })
  showModelsModal.value = false
}
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: rgba(156, 163, 175, 0.2); border-radius: 10px; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background-color: rgba(156, 163, 175, 0.4); }
</style>
