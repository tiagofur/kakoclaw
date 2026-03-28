<template>
  <div class="space-y-5 max-w-4xl mx-auto animate-fade-in-up">
    <!-- Orchestrator Section -->
    <div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
      <div class="absolute -top-24 -right-24 w-64 h-64 bg-makoclaw-accent/10 blur-[80px] rounded-full group-hover:bg-makoclaw-accent/20 transition-all duration-1000" />
      
      <div class="relative z-10">
        <div class="flex items-center justify-between mb-6">
          <div class="flex items-center gap-3">
            <div class="p-2.5 rounded-xl bg-gradient-to-br from-makoclaw-accent to-indigo-600 shadow-lg shadow-makoclaw-accent/10">
              <CpuChipIcon class="w-5 h-5 text-white" />
            </div>
            <div>
              <h3 class="text-[10px] font-bold uppercase tracking-[0.15em] text-makoclaw-accent leading-none mb-1">
                Intelligence Matrix
              </h3>
              <p class="text-lg font-black text-makoclaw-text uppercase italic tracking-tight">
                Orchestrator Protocol
              </p>
            </div>
          </div>
          <label class="relative inline-flex items-center cursor-pointer">
            <input
              v-model="orchestratorConfig.enabled"
              type="checkbox"
              class="sr-only peer"
            >
            <div class="w-11 h-6 bg-makoclaw-bg/80 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-makoclaw-accent" />
          </label>
        </div>

        <p class="text-sm font-medium text-makoclaw-text-secondary/60 mb-8 leading-relaxed max-w-2xl">
          The Orchestrator acts as the command layer, analyzing complex objectives and delegating sub-tasks to specialized neural nodes across your fleet.
        </p>

        <Transition name="fade-slide">
          <div
            v-if="orchestratorConfig.enabled"
            class="space-y-5 pt-5 border-t border-makoclaw-border/30"
          >
            <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
              <div class="space-y-2">
                <label class="text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">
                  Command Provider
                </label>
                <select
                  v-model="orchestratorConfig.provider"
                  class="w-full bg-makoclaw-surface border border-makoclaw-border/10 rounded-xl px-4 py-3 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent/50 outline-none transition-all cursor-pointer"
                >
                  <option
                    v-for="p in configuredProviders"
                    :key="p.name"
                    :value="p.name"
                  >
                    {{ p.name.toUpperCase() }}
                  </option>
                </select>
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-bold uppercase tracking-widest text-makoclaw-text-secondary/60 ml-1">
                  Primary Model Core
                </label>
                <select
                  v-model="orchestratorConfig.model"
                  class="w-full bg-makoclaw-surface border border-makoclaw-border/10 rounded-xl px-4 py-3 text-sm font-bold text-makoclaw-accent focus:border-makoclaw-accent/50 outline-none transition-all cursor-pointer"
                >
                  <optgroup
                    v-for="p in configuredProviders"
                    :key="p.name"
                    :label="p.name.toUpperCase()"
                  >
                    <option
                      v-for="m in p.models"
                      :key="m.id"
                      :value="m.id"
                    >
                      {{ m.id }}
                    </option>
                  </optgroup>
                </select>
              </div>
            </div>

            <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
              <div class="space-y-2">
                <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">
                  Cognitive Limit
                </label>
                <input
                  v-model.number="orchestratorConfig.max_tokens"
                  type="number"
                  class="w-full bg-makoclaw-bg/20 border border-makoclaw-border/30 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all"
                >
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">
                  Entropy (Temp)
                </label>
                <input
                  v-model.number="orchestratorConfig.temperature"
                  type="number"
                  step="0.1"
                  min="0"
                  max="2"
                  class="w-full bg-makoclaw-bg/20 border border-makoclaw-border/30 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all"
                >
              </div>
              <div class="space-y-2">
                <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">
                  Recursion Depth
                </label>
                <input
                  v-model.number="orchestratorConfig.max_delegation_retries"
                  type="number"
                  class="w-full bg-makoclaw-bg/20 border border-makoclaw-border/30 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all"
                >
              </div>
              <div class="flex items-center pt-6 ml-2">
                <label class="flex items-center gap-3 cursor-pointer group">
                  <input
                    v-model="orchestratorConfig.fallback_to_default"
                    type="checkbox"
                    class="w-5 h-5 rounded-lg border-2 border-makoclaw-border bg-makoclaw-bg text-makoclaw-accent focus:ring-makoclaw-accent"
                  >
                  <span class="text-[10px] font-medium uppercase tracking-tight text-makoclaw-text-secondary/70 group-hover:text-makoclaw-accent transition-colors">
                    Safety Fallback
                  </span>
                </label>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </div>

    <!-- Specialists Section -->
    <div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden">
      <div class="flex flex-col sm:flex-row items-center justify-between gap-4 mb-6">
        <div class="flex items-center gap-3">
          <div class="p-2.5 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-600 shadow-lg shadow-blue-500/10">
            <UserGroupIcon class="w-5 h-5 text-white" />
          </div>
          <div>
            <h3 class="text-[10px] font-bold uppercase tracking-[0.15em] text-blue-400 leading-none mb-1">
              Specialist Fleet
            </h3>
            <p class="text-xs font-bold text-makoclaw-text-secondary/60">
              Active Neural Nodes: {{ specialists.length }}
            </p>
          </div>
        </div>
        <button
          class="w-full sm:w-auto px-6 py-3 bg-white text-makoclaw-bg rounded-xl text-xs font-black uppercase tracking-widest flex items-center justify-center gap-2 hover:bg-white/90 transition-all active:scale-95 shadow-xl shadow-white/5"
          @click="openCreateSpecialist"
        >
          <PlusIcon class="w-4 h-4" />
          Enlist specialist
        </button>
      </div>

      <!-- Manual Mode Notice -->
      <div
        v-if="!orchestratorConfig.enabled && specialists.length > 0"
        class="mb-4 px-4 py-3 bg-amber-500/10 border border-amber-500/30 rounded-xl flex items-center gap-3"
      >
        <ExclamationTriangleIcon class="w-5 h-5 text-amber-400 flex-shrink-0" />
        <div>
          <p class="text-xs font-medium text-amber-300">Manual Mode Active</p>
          <p class="text-[10px] text-amber-300/70 mt-0.5">
            Orchestrator disabled. Use <code class="px-1.5 py-0.5 bg-amber-500/20 rounded text-amber-200">@specialist_name</code> in chat to invoke directly.
          </p>
        </div>
      </div>

      <!-- Empty State -->
      <div
        v-if="specialists.length === 0"
        class="py-16 bg-makoclaw-bg/20 rounded-2xl text-center border border-dashed border-makoclaw-border/30"
      >
        <div class="w-20 h-20 rounded-full bg-makoclaw-surface border border-makoclaw-border flex items-center justify-center mx-auto mb-6">
          <UserGroupIcon class="w-8 h-8 text-makoclaw-text-secondary/20" />
        </div>
        <p class="text-xs font-medium tracking-wide text-makoclaw-text-secondary/40">
          Hangar Empty
        </p>
        <p class="text-[10px] font-medium text-makoclaw-text-secondary/30 mt-4 max-w-xs mx-auto">
          No specialized agents detected in this sector. Enlist your first expert to enhance capabilities.
        </p>
      </div>

      <div
        v-if="specialists.length > 0"
        class="grid grid-cols-1 md:grid-cols-2 gap-6"
      >
        <div
          v-for="specialist in specialists"
          :key="specialist.name"
          class="group p-4 bg-makoclaw-surface/30 border border-makoclaw-border/50 rounded-2xl hover:border-makoclaw-accent/40 transition-all duration-300 hover:shadow-xl hover:shadow-makoclaw-accent/5 hover:-translate-y-0.5 relative overflow-hidden"
        >
          <!-- Background Decor -->
          <div class="absolute top-0 right-0 w-32 h-32 bg-gradient-to-bl from-white/5 to-transparent rounded-bl-full pointer-events-none transition-opacity group-hover:opacity-20 opacity-0" />

          <div class="flex items-start justify-between relative z-10">
            <div class="flex items-center gap-4">
              <div
                class="w-14 h-14 rounded-2xl flex items-center justify-center shadow-lg transition-transform group-hover:rotate-3 group-hover:scale-105 rotate-3"
                :class="getSpecialistBgColor(specialist.name)"
              >
                <div
                  class="w-7 h-7"
                  :class="getSpecialistColor(specialist.name)"
                  v-html="getSpecialistIcon(specialist.name)"
                />
              </div>
              <div class="min-w-0">
                <h4 class="text-lg font-black text-makoclaw-text capitalize tracking-tight">
                  {{ specialist.name }}
                </h4>
                <div class="flex items-center gap-2 mt-1">
                  <span class="text-[9px] font-medium tracking-wide text-makoclaw-text-secondary opacity-50">
                    {{ specialist.provider }}
                  </span>
                  <span class="w-1 h-1 rounded-full bg-makoclaw-border" />
                  <span class="text-[9px] font-medium tracking-wide text-makoclaw-accent">
                    {{ specialist.model.split('/').pop() }}
                  </span>
                </div>
              </div>
            </div>
            <div class="flex gap-1">
              <button
                class="p-2 rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all hover:scale-110"
                @click="openEditSpecialist(specialist)"
              >
                <PencilSquareIcon class="w-4 h-4" />
              </button>
              <button
                class="p-2 rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-red-400 transition-all hover:scale-110 ml-1"
                @click="deleteSpecialist(specialist.name)"
              >
                <TrashIcon class="w-4 h-4" />
              </button>
            </div>
          </div>
          
          <p class="text-xs font-medium text-makoclaw-text-secondary/70 mt-4 leading-relaxed line-clamp-2 relative z-10">
            {{ specialist.description }}
          </p>
          
          <div class="flex flex-wrap gap-2 mt-6 relative z-10">
            <span
              v-for="keyword in (specialist.keywords || []).slice(0, 4)"
              :key="keyword"
              class="px-3 py-1 text-[9px] font-medium tracking-wide bg-makoclaw-bg/60 border border-makoclaw-border/30 text-makoclaw-text-secondary/80 rounded-lg group-hover:border-makoclaw-accent/20 transition-colors"
            >
              {{ keyword }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Agent Defaults -->
    <div class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group">
      <div class="flex items-center gap-4 mb-6">
        <div class="p-3 rounded-2xl bg-makoclaw-surface border border-makoclaw-border group-hover:border-makoclaw-accent/40 transition-colors">
          <Cog6ToothIcon class="w-6 h-6 text-makoclaw-text-secondary group-hover:text-makoclaw-accent" />
        </div>
        <div>
          <h3 class="text-[11px] font-medium tracking-wide text-makoclaw-text-secondary/70">
            Default Protocols
          </h3>
          <p class="text-xs font-medium text-makoclaw-text-secondary mt-0.5">
            Automated Mission Parameters
          </p>
        </div>
      </div>

      <div class="space-y-10">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
          <div class="space-y-2">
            <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">
              Base Provider
            </label>
            <select
              v-model="localDefaults.provider"
              class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none cursor-pointer"
            >
              <option
                v-for="p in configuredProviders"
                :key="p.name"
                :value="p.name"
              >
                {{ p.name.toUpperCase() }}
              </option>
            </select>
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">
              Standard Core
            </label>
            <select
              v-model="localDefaults.model"
              class="w-full bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-2xl px-5 py-3.5 text-sm font-bold text-makoclaw-accent focus:border-makoclaw-accent outline-none cursor-pointer"
            >
              <optgroup
                v-for="p in configuredProviders"
                :key="p.name"
                :label="p.name.toUpperCase()"
              >
                <option
                  v-for="m in p.models"
                  :key="m.id"
                  :value="m.id"
                >
                  {{ m.id }}
                </option>
              </optgroup>
            </select>
          </div>
        </div>

        <!-- Image provider + model row -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
          <div class="space-y-2">
            <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1 flex items-center gap-1.5">
              <span class="text-purple-400">✦</span> Vision Provider
            </label>
            <select
              v-model="localImageProvider"
              class="w-full bg-makoclaw-bg/40 border-2 border-purple-500/30 rounded-2xl px-5 py-3.5 text-sm font-bold text-purple-400 focus:border-purple-400/50 outline-none cursor-pointer"
              @change="localDefaults.image_model = ''"
            >
              <option v-for="p in IMAGE_PROVIDERS" :key="p.id" :value="p.id">{{ p.label }}</option>
            </select>
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1 flex items-center gap-1.5">
              <span class="text-purple-400">✦</span> Vision Core
              <span class="text-[8px] font-normal normal-case tracking-normal text-makoclaw-text-secondary/30">(image_generate model)</span>
            </label>
            <select
              v-if="imageModels.length"
              v-model="localDefaults.image_model"
              class="w-full bg-makoclaw-bg/40 border-2 border-purple-500/30 rounded-2xl px-5 py-3.5 text-sm font-bold text-purple-400 focus:border-purple-400/50 outline-none cursor-pointer"
            >
              <option value="">— provider default —</option>
              <option v-for="m in imageModels" :key="m" :value="m">{{ m }}</option>
            </select>
            <p
              v-else
              class="px-5 py-3.5 bg-makoclaw-bg/20 border border-dashed border-purple-500/20 rounded-2xl text-[10px] text-makoclaw-text-secondary/40"
            >
              Configure models in <span class="font-bold text-purple-400/60">Settings → Providers → {{ localImageProvider }}</span>
            </p>
          </div>
        </div>

        <div class="grid grid-cols-2 md:grid-cols-3 gap-5 p-4 bg-makoclaw-bg/20 rounded-2xl border border-makoclaw-border/30">
          <div class="space-y-2">
            <label class="text-[9px] font-medium tracking-wide text-makoclaw-text-secondary/40 ml-1">
              Fluidity (Entropy)
            </label>
            <input
              v-model.number="localDefaults.temperature"
              type="number"
              step="0.1"
              min="0"
              max="2"
              class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent outline-none"
            >
          </div>
          <div class="space-y-2">
            <label class="text-[9px] font-medium tracking-wide text-makoclaw-text-secondary/40 ml-1">
              Output Cap
            </label>
            <input
              v-model.number="localDefaults.max_tokens"
              type="number"
              class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent outline-none"
            >
          </div>
          <div class="space-y-2">
            <label class="text-[9px] font-medium tracking-wide text-makoclaw-text-secondary/40 ml-1">
              Tool Endurance
            </label>
            <input
              v-model.number="localDefaults.max_tool_iterations"
              type="number"
              class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2 text-sm font-black text-makoclaw-text focus:border-makoclaw-accent outline-none"
            >
          </div>
        </div>
      </div>
    </div>

    <!-- Global Sync -->
    <div class="flex justify-center md:justify-end pt-6">
      <button
        class="w-full md:w-auto px-5 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl text-[11px] font-semibold shadow-lg shadow-makoclaw-accent/30 transition-all flex items-center justify-center gap-2 disabled:opacity-50 active:scale-95 group"
        :disabled="saving"
        @click="saveSettings"
      >
        <span
          v-if="saving"
          class="w-5 h-5 border-4 border-white/20 border-t-white rounded-full animate-spin"
        />
        <ArrowPathIcon
          v-else
          class="w-5 h-5 group-hover:rotate-180 transition-transform duration-700"
        />
        Commit All Protocols
      </button>
    </div>

    <!-- Specialist Form Modal -->
    <SpecialistFormModal
      :show="agentsStore.showSpecialistModal"
      :mode="agentsStore.specialistFormMode"
      :specialist="agentsStore.selectedSpecialist"
      :providers-list="providersList"
      @close="agentsStore.closeSpecialistModal"
      @save="handleSpecialistSaved"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useAgentsStore } from '../../stores/agentsStore'
import { useToast } from '../../composables/useToast'
import SpecialistFormModal from './SpecialistFormModal.vue'

import {
  CpuChipIcon,
  UserGroupIcon,
  Cog6ToothIcon,
  PlusIcon,
  PencilSquareIcon,
  TrashIcon,
  ExclamationTriangleIcon,
  ArrowPathIcon
} from '@heroicons/vue/24/outline'

const toast = useToast()
const agentsStore = useAgentsStore()

const props = defineProps({
  agents: { type: Object, required: true },
  providersList: { type: Array, required: true },
  configData: { type: Object, default: () => null },
  saving: { type: Boolean, default: false }
})

const emit = defineEmits(['save', 'refresh-config'])

const orchestratorConfig = ref({
  enabled: false,
  provider: 'anthropic',
  model: 'claude-opus',
  max_tokens: 12000,
  temperature: 0.7,
  max_delegation_retries: 2,
  fallback_to_default: true,
  description: 'Project Manager'
})

const localDefaults = ref({
  provider: '',
  model: '',
  image_model: '',
  temperature: 0.7,
  max_tokens: 4096,
  max_tool_iterations: 10
})

// Image providers list (mirrors ProvidersSettingsTab definitions)
const IMAGE_PROVIDERS = [
  { id: 'together',    label: 'Together.ai' },
  { id: 'openai',      label: 'OpenAI' },
  { id: 'openrouter',  label: 'OpenRouter' },
  { id: 'google',      label: 'Google (Gemini)' },
  { id: 'fal',       label: 'fal.ai' },
  { id: 'replicate', label: 'Replicate' },
  { id: 'zhipu',     label: 'Zhipu AI (CogView)' },
  { id: 'bfl',       label: 'Black Forest Labs' },
]

const IMAGE_PROVIDER_DEFAULTS = {
  together:    ['black-forest-labs/FLUX.1-schnell-Free', 'black-forest-labs/FLUX.1-schnell', 'black-forest-labs/FLUX.1-dev'],
  openai:      ['gpt-image-1', 'dall-e-3', 'dall-e-2'],
  openrouter:  ['openai/dall-e-3', 'black-forest-labs/FLUX.1-schnell', 'black-forest-labs/FLUX.1-pro'],
  google:      ['gemini-2.0-flash-exp', 'imagen-3.0-fast-generate-001', 'imagen-3.0-generate-001'],
  zhipu:     ['cogview-3-flash', 'cogview-3-plus', 'cogview-3'],
  fal:       ['fal-ai/flux/schnell', 'fal-ai/flux/dev', 'fal-ai/flux-pro/v1.1'],
  replicate: ['black-forest-labs/flux-schnell', 'black-forest-labs/flux-dev'],
  bfl:       ['flux-pro-1.1', 'flux-pro', 'flux-dev'],
}

const localImageProvider = ref(props.configData?.tools?.image?.provider || 'openai')

watch(() => props.configData?.tools?.image?.provider, (val) => {
  if (val) localImageProvider.value = val
})

const imageModels = computed(() => {
  const provider = localImageProvider.value
  const saved = props.configData?.tools?.image_providers?.[provider]?.models
  if (saved?.length) return saved
  return IMAGE_PROVIDER_DEFAULTS[provider] || []
})

const specialists = ref([])

const configuredProviders = computed(() => {
  const providers = Array.isArray(props.providersList) ? props.providersList : []
  return providers.filter(p => p?.enabled && p.models?.length > 0)
})

const configuredProviderModels = computed(() => {
  const map = {}
  configuredProviders.value.forEach(p => {
    map[p.name] = p.models || []
  })
  return map
})

const ensureValidDefaultSelection = () => {
  const p = configuredProviders.value
  if (!p.length) return
  
  if (!p.some(pr => pr.name === localDefaults.value.provider)) {
    localDefaults.value.provider = p[0].name
  }
  
  const models = configuredProviderModels.value[localDefaults.value.provider] || []
  if (!models.some(m => m.id === localDefaults.value.model) && models.length) {
    localDefaults.value.model = models[0].id
  }
}

onMounted(async () => {
  await agentsStore.fetchAgents()
  
  if (agentsStore.orchestrator) {
    orchestratorConfig.value = { ...agentsStore.orchestrator }
  }
  
  if (props.agents?.defaults) {
    localDefaults.value = { ...props.agents.defaults }
  }
  
  specialists.value = [...agentsStore.specialists]
  ensureValidDefaultSelection()
})

watch(() => props.agents, (newVal) => {
  if (newVal?.defaults) {
    localDefaults.value = { ...newVal.defaults }
  }
}, { deep: true })

watch(() => configuredProviders.value, ensureValidDefaultSelection, { deep: true })
watch(() => localDefaults.value.provider, ensureValidDefaultSelection)

const getSpecialistIcon = (n) => agentsStore.getSpecialistIcon(n)
const getSpecialistColor = (n) => agentsStore.getSpecialistColor(n)
const getSpecialistBgColor = (n) => agentsStore.getSpecialistBgColor(n)

const openCreateSpecialist = () => agentsStore.openSpecialistModal('create')
const openEditSpecialist = (s) => agentsStore.openSpecialistModal('edit', s)

const handleSpecialistSaved = async () => {
  await agentsStore.fetchAgents()
  specialists.value = [...agentsStore.specialists]
  emit('refresh-config')
}

const deleteSpecialist = async (n) => {
  if (confirm(`Sever connection to "${n}" node?`)) {
    await agentsStore.deleteSpecialist(n)
    specialists.value = [...agentsStore.specialists]
    emit('refresh-config')
    toast.success('Specialist severed')
  }
}

const saveSettings = async () => {
  const smap = {}
  specialists.value.forEach(s => {
    smap[s.name] = { ...s }
  })
  
  const updated = { 
    ...props.agents, 
    defaults: { ...localDefaults.value },
    orchestrator: { ...orchestratorConfig.value }, 
    specialists: smap 
  }
  emit('save', { agents: updated })
  // Sync active image provider to tools config
  const currentProvider = props.configData?.tools?.image?.provider
  if (localImageProvider.value !== currentProvider) {
    emit('save', { tools: { image: { provider: localImageProvider.value } } })
  }
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
