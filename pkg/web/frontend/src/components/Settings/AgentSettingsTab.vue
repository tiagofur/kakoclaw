<template>
  <div class="space-y-6 max-w-2xl mx-auto animate-fadeIn">
    <div class="glass-panel rounded-2xl p-8">
      <h3 class="text-sm font-bold uppercase tracking-widest text-kakoclaw-text-secondary opacity-60 mb-8 flex items-center">
         <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-kakoclaw-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
         </svg>
         Agent Defaults
      </h3>
      
      <div class="space-y-6">
        <div>
          <label class="block text-[10px] font-bold uppercase tracking-widest text-kakoclaw-text-secondary mb-2 opacity-70">Default Provider</label>
          <select v-model="agents.defaults.provider" class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm focus:ring-2 focus:ring-kakoclaw-accent/30 outline-none text-kakoclaw-text backdrop-blur-sm transition-all cursor-pointer">
            <option v-for="p in providersList" :key="p.name" :value="p.name">{{ p.name }}</option>
          </select>
        </div>

        <div>
          <label class="block text-[10px] font-bold uppercase tracking-widest text-kakoclaw-text-secondary mb-2 opacity-70">Default Model</label>
          <select v-model="agents.defaults.model" class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm focus:ring-2 focus:ring-kakoclaw-accent/30 outline-none text-kakoclaw-text backdrop-blur-sm transition-all cursor-pointer">
            <optgroup v-for="p in providersList" :key="p.name" :label="p.name">
              <option v-for="m in p.models" :key="m.id" :value="m.id">{{ m.id }}</option>
            </optgroup>
          </select>
        </div>

        <div class="grid grid-cols-2 gap-6">
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-kakoclaw-text-secondary mb-2 opacity-70">Temperature</label>
            <input v-model.number="agents.defaults.temperature" type="number" step="0.1" min="0" max="2" class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-kakoclaw-text focus:border-kakoclaw-accent backdrop-blur-sm transition-all">
          </div>
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-kakoclaw-text-secondary mb-2 opacity-70">Max Tokens</label>
            <input v-model.number="agents.defaults.max_tokens" type="number" class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-kakoclaw-text focus:border-kakoclaw-accent backdrop-blur-sm transition-all">
          </div>
        </div>

        <div>
          <label class="block text-[10px] font-bold uppercase tracking-widest text-kakoclaw-text-secondary mb-2 opacity-70">Max Tool Iterations</label>
          <input v-model.number="agents.defaults.max_tool_iterations" type="number" class="w-full bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-kakoclaw-text focus:border-kakoclaw-accent backdrop-blur-sm transition-all">
        </div>

        <div class="pt-6">
          <button @click="$emit('save', {agents})" :disabled="saving" class="w-full bg-kakoclaw-accent text-white py-3 rounded-xl font-bold hover:bg-kakoclaw-accent-hover transition-all shadow-lg shadow-kakoclaw-accent/20 flex items-center justify-center disabled:opacity-50 active:scale-95">
            <span v-if="saving" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin mr-2"></span>
            Save Agent Settings
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  agents: {
    type: Object,
    required: true
  },
  providersList: {
    type: Array,
    required: true
  },
  saving: {
    type: Boolean,
    default: false
  }
})
defineEmits(['save'])
</script>
