<template>
  <div class="space-y-8 max-w-4xl mx-auto animate-fadeIn">
    <!-- Orchestrator Section -->
    <div class="glass-panel rounded-2xl p-6 md:p-8">
      <div class="flex items-center justify-between mb-6">
        <h3 class="text-sm font-bold uppercase tracking-widest text-makoclaw-text-secondary opacity-60 flex items-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-makoclaw-accent" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
          </svg>
          Orchestrator
        </h3>
        <label class="relative inline-flex items-center cursor-pointer">
          <input v-model="orchestratorConfig.enabled" type="checkbox" class="sr-only peer">
          <div class="w-11 h-6 bg-makoclaw-bg/60 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-makoclaw-accent/50 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-makoclaw-accent"></div>
        </label>
      </div>

      <p class="text-sm text-makoclaw-text-secondary mb-6">
        Enable multi-agent system where the Orchestrator analyzes tasks and delegates to specialized agents.
      </p>

      <div v-if="orchestratorConfig.enabled" class="space-y-5 animate-fadeIn">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">Provider</label>
            <select v-model="orchestratorConfig.provider" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm focus:ring-2 focus:ring-makoclaw-accent/30 outline-none text-makoclaw-text backdrop-blur-sm transition-all cursor-pointer">
              <option v-for="p in providersList" :key="p.name" :value="p.name">{{ p.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">Model</label>
            <select v-model="orchestratorConfig.model" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm focus:ring-2 focus:ring-makoclaw-accent/30 outline-none text-makoclaw-text backdrop-blur-sm transition-all cursor-pointer">
              <optgroup v-for="p in providersList" :key="p.name" :label="p.name">
                <option v-for="m in p.models" :key="m.id" :value="m.id">{{ m.id }}</option>
              </optgroup>
            </select>
          </div>
        </div>

        <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">Max Tokens</label>
            <input v-model.number="orchestratorConfig.max_tokens" type="number" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-makoclaw-text focus:border-makoclaw-accent backdrop-blur-sm transition-all">
          </div>
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">Temperature</label>
            <input v-model.number="orchestratorConfig.temperature" type="number" step="0.1" min="0" max="2" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-makoclaw-text focus:border-makoclaw-accent backdrop-blur-sm transition-all">
          </div>
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">Max Retries</label>
            <input v-model.number="orchestratorConfig.max_delegation_retries" type="number" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-makoclaw-text focus:border-makoclaw-accent backdrop-blur-sm transition-all">
          </div>
        </div>

        <div>
          <label class="flex items-center space-x-3 cursor-pointer">
            <input v-model="orchestratorConfig.fallback_to_default" type="checkbox" class="rounded border-makoclaw-border bg-makoclaw-surface text-makoclaw-accent focus:ring-makoclaw-accent">
            <span class="text-sm text-makoclaw-text font-medium">Fallback to default agent on failure</span>
          </label>
        </div>
      </div>
    </div>

    <!-- Specialists Section -->
    <div class="glass-panel rounded-2xl p-6 md:p-8">
      <div class="flex items-center justify-between mb-6">
        <h3 class="text-sm font-bold uppercase tracking-widest text-makoclaw-text-secondary opacity-60 flex items-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
          Specialists
          <span v-if="specialists.length > 0" class="ml-2 px-2 py-0.5 text-[10px] font-bold bg-makoclaw-accent/20 text-makoclaw-accent rounded-full">{{ specialists.length }}</span>
        </h3>
        <button
          @click="openCreateSpecialist"
          class="px-4 py-2 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-lg text-sm font-bold flex items-center gap-2 shadow-lg shadow-makoclaw-accent/20 transition-all active:scale-95"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          Add Specialist
        </button>
      </div>

      <div v-if="!orchestratorConfig.enabled" class="p-8 bg-makoclaw-bg/40 rounded-xl text-center border border-dashed border-makoclaw-border">
        <svg class="w-12 h-12 mx-auto text-makoclaw-text-secondary/50 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
        </svg>
        <p class="text-sm text-makoclaw-text-secondary">Enable the Orchestrator to manage specialists</p>
      </div>

      <div v-else-if="specialists.length === 0" class="p-8 bg-makoclaw-bg/40 rounded-xl text-center border border-dashed border-makoclaw-border">
        <svg class="w-12 h-12 mx-auto text-makoclaw-text-secondary/50 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
        </svg>
        <p class="text-sm text-makoclaw-text-secondary mb-2">No specialists configured yet</p>
        <p class="text-xs text-makoclaw-text-secondary/70">Add your first specialist to start using the multi-agent system</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div
          v-for="specialist in specialists"
          :key="specialist.name"
          class="group p-4 bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl hover:border-makoclaw-accent/50 hover:shadow-lg hover:shadow-makoclaw-accent/5 transition-all"
        >
          <div class="flex items-start justify-between mb-3">
            <div class="flex items-center gap-3">
              <div :class="`w-10 h-10 rounded-lg ${getSpecialistBgColor(specialist.name)} flex items-center justify-center`">
                <svg class="w-5 h-5" :class="getSpecialistColor(specialist.name)" fill="none" stroke="currentColor" viewBox="0 0 24 24" v-html="getSpecialistIcon(specialist.name)"></svg>
              </div>
              <div>
                <h4 class="font-bold text-makoclaw-text capitalize">{{ specialist.name }}</h4>
                <p class="text-xs text-makoclaw-text-secondary">{{ specialist.provider }} / {{ specialist.model }}</p>
              </div>
            </div>
            <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <button @click="openEditSpecialist(specialist)" class="p-1.5 hover:bg-makoclaw-bg rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-accent transition-colors">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
              </button>
              <button @click="deleteSpecialist(specialist.name)" class="p-1.5 hover:bg-red-500/10 rounded-lg text-makoclaw-text-secondary hover:text-red-400 transition-colors">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </div>
          <p class="text-sm text-makoclaw-text-secondary mb-3 line-clamp-2">{{ specialist.description }}</p>
          <div class="flex flex-wrap gap-1">
            <span
              v-for="keyword in (specialist.keywords || []).slice(0, 3)"
              :key="keyword"
              class="px-2 py-0.5 text-[10px] font-medium bg-makoclaw-bg/60 text-makoclaw-text-secondary rounded-full"
            >
              {{ keyword }}
            </span>
            <span v-if="(specialist.keywords || []).length > 3" class="px-2 py-0.5 text-[10px] font-medium bg-makoclaw-bg/60 text-makoclaw-text-secondary rounded-full">
              +{{ specialist.keywords.length - 3 }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Agent Defaults -->
    <div class="glass-panel rounded-2xl p-6 md:p-8">
      <h3 class="text-sm font-bold uppercase tracking-widest text-makoclaw-text-secondary opacity-60 mb-6 flex items-center">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2 text-makoclaw-text-secondary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
        </svg>
        Agent Defaults
      </h3>

      <div class="space-y-5">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">Default Provider</label>
            <select v-model="agents.defaults.provider" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm focus:ring-2 focus:ring-makoclaw-accent/30 outline-none text-makoclaw-text backdrop-blur-sm transition-all cursor-pointer">
              <option v-for="p in providersList" :key="p.name" :value="p.name">{{ p.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">Default Model</label>
            <select v-model="agents.defaults.model" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm focus:ring-2 focus:ring-makoclaw-accent/30 outline-none text-makoclaw-text backdrop-blur-sm transition-all cursor-pointer">
              <optgroup v-for="p in providersList" :key="p.name" :label="p.name">
                <option v-for="m in p.models" :key="m.id" :value="m.id">{{ m.id }}</option>
              </optgroup>
            </select>
          </div>
        </div>

        <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">Temperature</label>
            <input v-model.number="agents.defaults.temperature" type="number" step="0.1" min="0" max="2" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-makoclaw-text focus:border-makoclaw-accent backdrop-blur-sm transition-all">
          </div>
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">Max Tokens</label>
            <input v-model.number="agents.defaults.max_tokens" type="number" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-makoclaw-text focus:border-makoclaw-accent backdrop-blur-sm transition-all">
          </div>
          <div>
            <label class="block text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-2 opacity-70">Max Tool Iterations</label>
            <input v-model.number="agents.defaults.max_tool_iterations" type="number" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl px-4 py-2.5 text-sm outline-none text-makoclaw-text focus:border-makoclaw-accent backdrop-blur-sm transition-all">
          </div>
        </div>
      </div>
    </div>

    <!-- Save Button -->
    <div class="flex justify-end">
      <button @click="saveSettings" :disabled="saving" class="px-8 py-3 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl font-bold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center gap-2 disabled:opacity-50 active:scale-95">
        <span v-if="saving" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
        Save All Settings
      </button>
    </div>

    <!-- Specialist Form Modal -->
    <SpecialistFormModal
      :show="agentsStore.showSpecialistModal"
      :mode="agentsStore.specialistFormMode"
      :specialist="agentsStore.selectedSpecialist"
      :providersList="providersList"
      @close="closeSpecialistModal"
      @save="handleSpecialistSaved"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAgentsStore } from '../../stores/agentsStore'
import { useToast } from '../../composables/useToast'
import SpecialistFormModal from './SpecialistFormModal.vue'

const toast = useToast()
const agentsStore = useAgentsStore()

const props = defineProps({
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

const emit = defineEmits(['save'])

const orchestratorConfig = ref({
  enabled: false,
  provider: 'anthropic',
  model: 'claude-opus',
  max_tokens: 12000,
  temperature: 0.7,
  max_delegation_retries: 2,
  fallback_to_default: true,
  description: 'Project Manager: Analyzes tasks and delegates to appropriate specialists'
})

const specialists = ref([])

onMounted(async () => {
  await agentsStore.fetchAgents()
  if (agentsStore.orchestrator) {
    orchestratorConfig.value = { ...agentsStore.orchestrator }
  }
  specialists.value = [...agentsStore.specialists]
})

function getSpecialistIcon(name) {
  return agentsStore.getSpecialistIcon(name)
}

function getSpecialistColor(name) {
  return agentsStore.getSpecialistColor(name)
}

function getSpecialistBgColor(name) {
  return agentsStore.getSpecialistBgColor(name)
}

function openCreateSpecialist() {
  agentsStore.openSpecialistModal('create')
}

function openEditSpecialist(specialist) {
  agentsStore.openSpecialistModal('edit', specialist)
}

function closeSpecialistModal() {
  agentsStore.closeSpecialistModal()
}

async function handleSpecialistSaved() {
  await agentsStore.fetchAgents()
  specialists.value = [...agentsStore.specialists]
}

async function deleteSpecialist(name) {
  if (!confirm(`Are you sure you want to delete the "${name}" specialist?`)) return
  const success = await agentsStore.deleteSpecialist(name)
  if (success) {
    specialists.value = [...agentsStore.specialists]
  }
}

async function saveSettings() {
  // Specialists are managed locally in this component's specialists ref,
  // while props.agents might be stale or incomplete.
  // We reconstruct the specialist map to match the backend config structure.
  const specialistMap = {}
  specialists.value.forEach(s => {
    specialistMap[s.name] = { ...s }
  })

  const updatedAgents = {
    ...props.agents,
    orchestrator: { ...orchestratorConfig.value },
    specialists: specialistMap
  }

  if (updatedAgents.orchestrator.enabled && specialists.value.length === 0) {
    toast.warning('Please add at least one specialist before enabling the orchestrator')
    return
  }

  emit('save', { agents: updatedAgents })
}
</script>


