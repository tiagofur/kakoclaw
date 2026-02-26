<template>
  <div class="space-y-8 max-w-4xl mx-auto animate-fade-in-up">
    <!-- Orchestrator Section -->
    <div class="glass-panel rounded-[2rem] p-8 md:p-10 border border-makoclaw-border/50 relative overflow-hidden group">
      <div class="absolute -top-24 -right-24 w-64 h-64 bg-makoclaw-accent/10 blur-[80px] rounded-full group-hover:bg-makoclaw-accent/20 transition-all duration-1000"></div>
      
      <div class="relative z-10">
        <div class="flex items-center justify-between mb-8">
          <div class="flex items-center gap-4">
            <div class="p-3 rounded-2xl bg-gradient-to-br from-makoclaw-accent to-indigo-600 shadow-lg shadow-makoclaw-accent/20">
              <IconOrchestrator class="w-6 h-6 text-white" />
            </div>
            <div>
              <h3 class="text-[11px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70">Intelligence Matrix</h3>
              <p class="text-lg font-black text-makoclaw-text italic leading-none">Orchestrator Protocol</p>
            </div>
          </div>
          <label class="relative inline-flex items-center cursor-pointer scale-110">
            <input v-model="orchestratorConfig.enabled" type="checkbox" class="sr-only peer">
            <div class="w-14 h-7 bg-makoclaw-bg/80 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[4px] after:left-[4px] after:bg-white after:rounded-full after:h-[21px] after:w-[21px] after:transition-all peer-checked:bg-makoclaw-accent"></div>
          </label>
        </div>

        <p class="text-sm font-medium text-makoclaw-text-secondary/60 mb-10 leading-relaxed max-w-2xl">
          The Orchestrator acts as the command layer, analyzing complex objectives and delegating sub-tasks to specialized neural nodes across your fleet.
        </p>

        <Transition name="fade-slide">
          <div v-if="orchestratorConfig.enabled" class="space-y-8 pt-6 border-t border-makoclaw-border/30">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
              <div class="space-y-2">
                <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Command Provider</label>
                <select v-model="orchestratorConfig.provider" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer">
                  <option v-for="p in configuredProviders" :key="p.name" :value="p.name">{{ p.name.toUpperCase() }}</option>
                </select>
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Primary Model Core</label>
                <select v-model="orchestratorConfig.model" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-accent focus:border-makoclaw-accent outline-none transition-all cursor-pointer">
                  <optgroup v-for="p in configuredProviders" :key="p.name" :label="p.name.toUpperCase()">
                    <option v-for="m in p.models" :key="m.id" :value="m.id">{{ m.id }}</option>
                  </optgroup>
                </select>
              </div>
            </div>

            <div class="grid grid-cols-2 lg:grid-cols-4 gap-6">
              <div class="space-y-2">
                <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Cognitive Limit</label>
                <input v-model.number="orchestratorConfig.max_tokens" type="number" class="w-full bg-makoclaw-bg/20 border-2 border-makoclaw-border/30 rounded-2xl px-5 py-3 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all">
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Entropy (Temp)</label>
                <input v-model.number="orchestratorConfig.temperature" type="number" step="0.1" min="0" max="2" class="w-full bg-makoclaw-bg/20 border-2 border-makoclaw-border/30 rounded-2xl px-5 py-3 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all">
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Recursion Depth</label>
                <input v-model.number="orchestratorConfig.max_delegation_retries" type="number" class="w-full bg-makoclaw-bg/20 border-2 border-makoclaw-border/30 rounded-2xl px-5 py-3 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all">
              </div>
              <div class="flex items-center pt-6 ml-2">
                <label class="flex items-center gap-3 cursor-pointer group">
                  <input v-model="orchestratorConfig.fallback_to_default" type="checkbox" class="w-5 h-5 rounded-lg border-2 border-makoclaw-border bg-makoclaw-bg text-makoclaw-accent focus:ring-makoclaw-accent">
                  <span class="text-[10px] font-black uppercase tracking-tight text-makoclaw-text-secondary/70 group-hover:text-makoclaw-accent transition-colors">Safety Fallback</span>
                </label>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </div>

    <!-- Specialists Section -->
    <div class="glass-panel rounded-[2rem] p-8 md:p-10 border border-makoclaw-border/50 relative overflow-hidden">
      <div class="flex flex-col sm:flex-row items-center justify-between gap-6 mb-10">
        <div class="flex items-center gap-4">
          <div class="p-3 rounded-2xl bg-gradient-to-br from-blue-500 to-cyan-600 shadow-lg shadow-blue-500/20">
            <IconSpecialists class="w-6 h-6 text-white" />
          </div>
          <div>
            <h3 class="text-[11px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70 italic">Specialist Fleet</h3>
             <p class="text-xs font-bold text-blue-400 uppercase tracking-widest mt-0.5">Active Neural Nodes: {{ specialists.length }}</p>
          </div>
        </div>
        <button
          @click="openCreateSpecialist"
          class="w-full sm:w-auto px-6 py-3 bg-white text-makoclaw-bg rounded-2xl text-[11px] font-black uppercase tracking-widest flex items-center justify-center gap-2 hover:bg-white/90 transition-all active:scale-95 shadow-xl shadow-white/5"
        >
          <IconPlus class="w-4 h-4" />
          Enlist specialist
        </button>
      </div>

      <!-- Empty States -->
      <div v-if="!orchestratorConfig.enabled" class="py-20 bg-makoclaw-bg/20 rounded-[2rem] text-center border-2 border-dashed border-makoclaw-border/30 animate-pulse">
        <IconBlock class="w-16 h-16 mx-auto text-makoclaw-text-secondary/10 mb-6" />
        <p class="text-xs font-black uppercase tracking-[0.3em] text-makoclaw-text-secondary/30 italic">Matrix Disconnected</p>
        <p class="text-[10px] font-medium text-makoclaw-text-secondary/20 mt-2">Activate Orchestrator to interface with specialists</p>
      </div>

      <div v-else-if="specialists.length === 0" class="py-20 bg-makoclaw-bg/20 rounded-[2rem] text-center border-2 border-dashed border-makoclaw-border/30">
        <div class="w-20 h-20 rounded-full bg-makoclaw-surface border border-makoclaw-border flex items-center justify-center mx-auto mb-6">
           <IconSpecialists class="w-8 h-8 text-makoclaw-text-secondary/20" />
        </div>
        <p class="text-xs font-black uppercase tracking-[0.3em] text-makoclaw-text-secondary/40 italic">Hangar Empty</p>
        <p class="text-[10px] font-medium text-makoclaw-text-secondary/30 mt-4 max-w-xs mx-auto">No specialized agents detected in this sector. Enlist your first expert to enhance capabilities.</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div
          v-for="specialist in specialists"
          :key="specialist.name"
          class="group p-6 bg-makoclaw-surface/30 border-2 border-makoclaw-border/50 rounded-[2rem] hover:border-makoclaw-accent/40 transition-all duration-500 hover:shadow-2xl hover:shadow-makoclaw-accent/5 hover:-translate-y-1 relative overflow-hidden"
        >
          <!-- Background Decor -->
          <div class="absolute top-0 right-0 w-32 h-32 bg-gradient-to-bl from-white/5 to-transparent rounded-bl-full pointer-events-none transition-opacity group-hover:opacity-20 opacity-0"></div>

          <div class="flex items-start justify-between relative z-10">
            <div class="flex items-center gap-4">
              <div :class="`w-14 h-14 rounded-2xl ${getSpecialistBgColor(specialist.name)} flex items-center justify-center shadow-lg transition-transform group-hover:rotate-3 group-hover:scale-105` + ' rotate-3'">
                <div class="w-7 h-7" :class="getSpecialistColor(specialist.name)" v-html="getSpecialistIcon(specialist.name)"></div>
              </div>
              <div class="min-w-0">
                <h4 class="text-lg font-black text-makoclaw-text capitalize tracking-tight">{{ specialist.name }}</h4>
                <div class="flex items-center gap-2 mt-1">
                   <span class="text-[9px] font-black uppercase tracking-widest text-makoclaw-text-secondary opacity-50">{{ specialist.provider }}</span>
                   <span class="w-1 h-1 rounded-full bg-makoclaw-border"></span>
                   <span class="text-[9px] font-black uppercase tracking-widest text-makoclaw-accent">{{ specialist.model.split('/').pop() }}</span>
                </div>
              </div>
            </div>
            <div class="flex gap-1">
              <button @click="openEditSpecialist(specialist)" class="p-2 rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all hover:scale-110"><IconEdit class="w-4 h-4"/></button>
              <button @click="deleteSpecialist(specialist.name)" class="p-2 rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-red-400 transition-all hover:scale-110 ml-1"><IconDelete class="w-4 h-4"/></button>
            </div>
          </div>
          
          <p class="text-xs font-medium text-makoclaw-text-secondary/70 mt-4 leading-relaxed line-clamp-2 relative z-10">{{ specialist.description }}</p>
          
          <div class="flex flex-wrap gap-2 mt-6 relative z-10">
            <span
              v-for="keyword in (specialist.keywords || []).slice(0, 4)"
              :key="keyword"
              class="px-3 py-1 text-[9px] font-black uppercase tracking-widest bg-makoclaw-bg/60 border border-makoclaw-border/30 text-makoclaw-text-secondary/80 rounded-lg group-hover:border-makoclaw-accent/20 transition-colors"
            >
              {{ keyword }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Agent Defaults -->
    <div class="glass-panel rounded-[2rem] p-8 md:p-10 border border-makoclaw-border/50 relative overflow-hidden group">
      <div class="flex items-center gap-4 mb-10">
        <div class="p-3 rounded-2xl bg-makoclaw-surface border border-makoclaw-border group-hover:border-makoclaw-accent/40 transition-colors">
          <IconSettings class="w-6 h-6 text-makoclaw-text-secondary group-hover:text-makoclaw-accent" />
        </div>
        <div>
          <h3 class="text-[11px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70 italic">Default Protocols</h3>
          <p class="text-xs font-bold text-makoclaw-text-secondary uppercase tracking-widest mt-0.5">Automated Mission Parameters</p>
        </div>
      </div>

      <div class="space-y-10">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div class="space-y-2">
            <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Base Provider</label>
            <select v-model="agents.defaults.provider" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none cursor-pointer">
              <option v-for="p in configuredProviders" :key="p.name" :value="p.name">{{ p.name.toUpperCase() }}</option>
            </select>
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">Standard Core</label>
            <select v-model="agents.defaults.model" class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-accent focus:border-makoclaw-accent outline-none cursor-pointer">
              <optgroup v-for="p in configuredProviders" :key="p.name" :label="p.name.toUpperCase()">
                <option v-for="m in p.models" :key="m.id" :value="m.id">{{ m.id }}</option>
              </optgroup>
            </select>
          </div>
        </div>

        <div class="grid grid-cols-2 md:grid-cols-3 gap-8 p-6 bg-makoclaw-bg/20 rounded-[2rem] border border-makoclaw-border/30">
          <div class="space-y-2">
            <label class="text-[9px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40 ml-1">Fluidity (Entropy)</label>
            <input v-model.number="agents.defaults.temperature" type="number" step="0.1" min="0" max="2" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent outline-none">
          </div>
          <div class="space-y-2">
            <label class="text-[9px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40 ml-1">Output Cap</label>
            <input v-model.number="agents.defaults.max_tokens" type="number" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent outline-none">
          </div>
          <div class="space-y-2">
            <label class="text-[9px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40 ml-1">Tool Endurance</label>
            <input v-model.number="agents.defaults.max_tool_iterations" type="number" class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent outline-none">
          </div>
        </div>
      </div>
    </div>

    <!-- Global Sync -->
    <div class="flex justify-center md:justify-end pt-6">
      <button @click="saveSettings" :disabled="saving" class="w-full md:w-auto px-10 py-5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-[1.5rem] text-[11px] font-black uppercase tracking-[0.2em] shadow-2xl shadow-makoclaw-accent/30 transition-all flex items-center justify-center gap-3 disabled:opacity-50 active:scale-95 group">
        <span v-if="saving" class="w-5 h-5 border-4 border-white/20 border-t-white rounded-full animate-spin"></span>
        <IconSync v-else class="w-5 h-5 group-hover:rotate-180 transition-transform duration-700" />
        Commit All Protocols
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
import { ref, computed, onMounted, watch, h } from 'vue'
import { useAgentsStore } from '../../stores/agentsStore'
import { useToast } from '../../composables/useToast'
import SpecialistFormModal from './SpecialistFormModal.vue'

// Premium Icons
const IconOrchestrator = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z' })]) }
const IconSpecialists = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z' })]) }
const IconSettings = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4' })]) }
const IconPlus = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M12 6v6m0 0v6m0-6h6m-6 0H6' })]) }
const IconEdit = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z' })]) }
const IconDelete = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16' })]) }
const IconBlock = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636' })]) }
const IconSync = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15' })]) }

const toast = useToast()
const agentsStore = useAgentsStore()

const props = defineProps({
  agents: { type: Object, required: true },
  providersList: { type: Array, required: true },
  saving: { type: Boolean, default: false }
})

const emit = defineEmits(['save'])

const orchestratorConfig = ref({
  enabled: false, provider: 'anthropic', model: 'claude-opus',
  max_tokens: 12000, temperature: 0.7, max_delegation_retries: 2,
  fallback_to_default: true, description: 'Project Manager'
})

const specialists = ref([])

const configuredProviders = computed(() => {
  const providers = Array.isArray(props.providersList) ? props.providersList : []
  return providers.filter(p => p?.enabled && p.models?.length > 0)
})

const configuredProviderModels = computed(() => {
  const map = {}
  configuredProviders.value.forEach(p => map[p.name] = p.models || [])
  return map
})

const ensureValidDefaultSelection = () => {
  const p = configuredProviders.value
  if (!p.length) return
  const currentP = props.agents?.defaults?.provider
  if (!p.some(pr => pr.name === currentP)) props.agents.defaults.provider = p[0].name
  const models = configuredProviderModels.value[props.agents.defaults.provider] || []
  if (!models.some(m => m.id === props.agents.defaults.model) && models.length) props.agents.defaults.model = models[0].id
}

onMounted(async () => {
  await agentsStore.fetchAgents()
  if (agentsStore.orchestrator) orchestratorConfig.value = { ...agentsStore.orchestrator }
  specialists.value = [...agentsStore.specialists]
  ensureValidDefaultSelection()
})

watch(() => configuredProviders.value, ensureValidDefaultSelection, { deep: true })
watch(() => props.agents?.defaults?.provider, ensureValidDefaultSelection)

const getSpecialistIcon = (n) => agentsStore.getSpecialistIcon(n)
const getSpecialistColor = (n) => agentsStore.getSpecialistColor(n)
const getSpecialistBgColor = (n) => agentsStore.getSpecialistBgColor(n)

const openCreateSpecialist = () => agentsStore.openSpecialistModal('create')
const openEditSpecialist = (s) => agentsStore.openSpecialistModal('edit', s)
const closeSpecialistModal = () => agentsStore.closeSpecialistModal()
const handleSpecialistSaved = async () => { await agentsStore.fetchAgents(); specialists.value = [...agentsStore.specialists] }
const deleteSpecialist = async (n) => { if (confirm(`Sever connection to "${n}" node?`)) { await agentsStore.deleteSpecialist(n); specialists.value = [...agentsStore.specialists] } }

const saveSettings = async () => {
  const smap = {}
  specialists.value.forEach(s => smap[s.name] = { ...s })
  const updated = { ...props.agents, orchestrator: { ...orchestratorConfig.value }, specialists: smap }
  emit('save', { agents: updated })
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
.fade-slide-enter-active, .fade-slide-leave-active { transition: all 0.3s ease; }
.fade-slide-enter-from { opacity: 0; transform: translateY(-10px); }
.fade-slide-leave-to { opacity: 0; transform: translateY(-10px); }
</style>
