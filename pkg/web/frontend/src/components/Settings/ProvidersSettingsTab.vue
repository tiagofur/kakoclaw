<template>
  <div class="space-y-5 max-w-4xl mx-auto animate-fade-in-up">
    <div
      v-for="(info, name) in providers"
      :key="name"
      class="glass-panel rounded-2xl p-5 md:p-6 border border-makoclaw-border/50 relative overflow-hidden group hover:border-makoclaw-accent/30 transition-all duration-300"
    >
      <!-- Decorative background blur -->
      <div class="absolute -top-12 -right-12 w-48 h-48 bg-makoclaw-accent/5 blur-[60px] rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-1000" />

      <div class="flex items-center justify-between mb-5 relative z-10">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-makoclaw-surface border border-makoclaw-border/50 flex items-center justify-center text-makoclaw-accent shadow-md group-hover:scale-105 group-hover:rotate-3 transition-all duration-300">
            <span class="text-xs font-bold uppercase">{{ name.substring(0,2) }}</span>
          </div>
          <div>
            <h3 class="text-base font-semibold capitalize text-makoclaw-text tracking-tight">
              {{ name }}
            </h3>
            <div class="flex items-center gap-2 mt-1">
              <span
                class="px-2 py-0.5 text-[9px] font-medium tracking-wide rounded-md border flex items-center gap-1"
                :class="isProviderConfigured(info) ? 'bg-makoclaw-success/10 text-makoclaw-success border-makoclaw-success/20' : 'bg-makoclaw-text-secondary/10 text-makoclaw-text-secondary border-makoclaw-border/50'"
              >
                <span
                  class="w-1.5 h-1.5 rounded-full"
                  :class="isProviderConfigured(info) ? 'bg-makoclaw-success animate-pulse' : 'bg-makoclaw-text-secondary/40'"
                />
                {{ isProviderConfigured(info) ? 'Linked' : 'Offline' }}
              </span>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <button
            class="p-2 rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-makoclaw-accent transition-all hover:scale-105 active:scale-95 group/btn"
            title="Configure Models"
            @click="openModelsModal(name, info)"
          >
            <CpuChipIcon class="w-5 h-5 group-hover/btn:rotate-12 transition-transform" />
          </button>
        </div>
      </div>
      
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-4 items-end relative z-10">
        <div class="lg:col-span-12 xl:col-span-5 space-y-2">
          <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">
            API Key
          </label>
          <div class="relative group/input">
            <input
              v-model="info.api_key"
              type="password"
              :placeholder="isProviderConfigured(info) ? '••••••••••••••••••••••••' : 'Enter access key...'"
              class="w-full bg-makoclaw-surface border border-makoclaw-border/10 rounded-xl px-4 py-3 text-sm font-bold outline-none focus:border-makoclaw-accent/50 text-makoclaw-text transition-all"
            >
            <div class="absolute right-4 top-1/2 -translate-y-1/2 text-makoclaw-text-secondary/20 group-hover/input:text-makoclaw-accent/40 transition-colors">
              <LockClosedIcon class="w-4 h-4" />
            </div>
          </div>
        </div>
        <div class="lg:col-span-12 xl:col-span-4 space-y-2">
          <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">
            Base URL
          </label>
          <input
            v-model="info.api_base"
            type="text"
            placeholder="https://api.neural-matrix..."
            class="w-full bg-makoclaw-surface border border-makoclaw-border/10 rounded-xl px-4 py-3 text-sm font-bold outline-none focus:border-makoclaw-accent/50 text-makoclaw-text transition-all"
          >
        </div>
        <div class="lg:col-span-12 xl:col-span-3">
          <button
            class="w-full px-6 py-3 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl text-xs font-black uppercase tracking-widest shadow-xl shadow-makoclaw-accent/10 transition-all flex items-center justify-center gap-2 active:scale-95 disabled:opacity-50 group/save"
            :disabled="saving"
            @click="$emit('save', {providers: {[name]: {api_key: info.api_key, api_base: info.api_base, models: Array.isArray(info.models) ? info.models : []}}})"
          >
            <span
              v-if="saving"
              class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin"
            />
            <BoltIcon
              v-else
              class="w-4 h-4 group-hover/save:translate-y-[-2px] transition-transform"
            />
            Connect
          </button>
        </div>
      </div>
    </div>

    <!-- Image Generation Section -->
    <div class="mt-8">
      <div class="flex items-center gap-3 mb-4">
        <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-purple-500/20 to-pink-500/20 flex items-center justify-center">
          <PhotoIcon class="w-4 h-4 text-purple-400" />
        </div>
        <div class="flex-1">
          <h3 class="text-sm font-semibold text-makoclaw-text">Image Generation</h3>
          <p class="text-xs text-makoclaw-text-secondary">Configure credentials per provider · model selection in Agents tab</p>
        </div>
        <!-- Active provider selector -->
        <div class="flex items-center gap-2">
          <span class="text-[9px] font-bold uppercase tracking-widest text-purple-400">Active:</span>
          <select
            v-model="activeImageProvider"
            class="bg-purple-500/10 border border-purple-500/20 rounded-lg px-3 py-1.5 text-xs font-bold text-purple-300 outline-none cursor-pointer"
            @change="saveActiveProvider"
          >
            <option v-for="p in IMAGE_PROVIDERS" :key="p.id" :value="p.id">{{ p.label }}</option>
          </select>
        </div>
      </div>

      <!-- One card per provider (purple theme) -->
      <div
        v-for="p in IMAGE_PROVIDERS"
        :key="p.id"
        class="glass-panel rounded-2xl p-5 md:p-6 border transition-all duration-300 mb-3"
        :class="activeImageProvider === p.id
          ? 'border-purple-500/40 hover:border-purple-500/60'
          : 'border-makoclaw-border/30 hover:border-purple-500/20'"
      >
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-3">
            <div
              class="w-9 h-9 rounded-xl flex items-center justify-center text-purple-300 border shadow-md"
              :class="activeImageProvider === p.id ? 'bg-purple-500/20 border-purple-500/30' : 'bg-makoclaw-surface border-makoclaw-border/50'"
            >
              <span class="text-xs font-bold uppercase">{{ p.id.substring(0, 2) }}</span>
            </div>
            <div>
              <h4 class="text-sm font-semibold text-makoclaw-text flex items-center gap-2">
                {{ p.label }}
                <span
                  v-if="activeImageProvider === p.id"
                  class="px-1.5 py-0.5 text-[8px] font-black uppercase tracking-widest bg-purple-500/20 text-purple-300 rounded border border-purple-500/30"
                >ACTIVE</span>
              </h4>
              <div class="flex items-center gap-2 mt-0.5">
                <span
                  class="px-2 py-0.5 text-[9px] font-medium rounded-md border flex items-center gap-1"
                  :class="isImageProviderConfigured(p.id) ? 'bg-makoclaw-success/10 text-makoclaw-success border-makoclaw-success/20' : 'bg-makoclaw-text-secondary/10 text-makoclaw-text-secondary border-makoclaw-border/50'"
                >
                  <span
                    class="w-1.5 h-1.5 rounded-full"
                    :class="isImageProviderConfigured(p.id) ? 'bg-makoclaw-success animate-pulse' : 'bg-makoclaw-text-secondary/40'"
                  />
                  {{ isImageProviderConfigured(p.id) ? 'Linked' : 'Offline' }}
                </span>
                <span v-if="p.sharedKey" class="text-[9px] text-purple-400/60 font-medium">
                  uses {{ p.sharedKey }} key
                </span>
              </div>
            </div>
          </div>
          <button
            class="p-2 rounded-xl bg-makoclaw-bg/40 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-purple-400 transition-all hover:scale-105 active:scale-95 group/btn"
            title="Configure Models"
            @click="openImageProviderModelsModal(p.id)"
          >
            <CpuChipIcon class="w-5 h-5 group-hover/btn:rotate-12 transition-transform" />
          </button>
        </div>

        <div class="grid grid-cols-1 lg:grid-cols-12 gap-4 items-end">
          <!-- Shared key note + optional override -->
          <template v-if="p.sharedKey">
            <div class="lg:col-span-12 xl:col-span-9 space-y-3">
              <div class="flex items-center gap-2 px-4 py-3 rounded-xl bg-purple-500/5 border border-purple-500/15 text-purple-300 text-xs font-medium">
                <CheckCircleIcon class="w-4 h-4 flex-none" />
                <span>Shared key from <span class="font-bold">{{ p.label }}</span> LLM provider · or enter a dedicated key below</span>
              </div>
              <div class="relative group/input">
                <input
                  v-model="imageProviders[p.id].api_key"
                  type="password"
                  :placeholder="'Override ' + p.label + ' key (optional)'"
                  class="w-full bg-makoclaw-surface border border-makoclaw-border/10 rounded-xl px-4 py-3 text-sm font-bold outline-none focus:border-purple-500/40 text-makoclaw-text transition-all"
                >
                <LockClosedIcon class="absolute right-4 top-1/2 -translate-y-1/2 w-4 h-4 text-makoclaw-text-secondary/20" />
              </div>
            </div>
          </template>
          <!-- Dedicated API key -->
          <template v-else>
            <div class="lg:col-span-12 xl:col-span-5 space-y-2">
              <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1 flex items-center gap-2">
                API Key
                <a
                  v-if="p.apiKeyUrl"
                  :href="p.apiKeyUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-purple-400 hover:underline normal-case tracking-normal font-semibold"
                >Get key →</a>
              </label>
              <div class="relative group/input">
                <input
                  v-model="imageProviders[p.id].api_key"
                  type="password"
                  placeholder="••••••••••••••••••••••••"
                  class="w-full bg-makoclaw-surface border border-makoclaw-border/10 rounded-xl px-4 py-3 text-sm font-bold outline-none focus:border-purple-500/40 text-makoclaw-text transition-all"
                >
                <LockClosedIcon class="absolute right-4 top-1/2 -translate-y-1/2 w-4 h-4 text-makoclaw-text-secondary/20" />
              </div>
            </div>
          </template>

          <div class="lg:col-span-12 xl:col-span-4 space-y-2">
            <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">Base URL</label>
            <input
              v-model="imageProviders[p.id].api_base"
              type="text"
              :placeholder="p.defaultBase || 'https://api...'"
              class="w-full bg-makoclaw-surface border border-makoclaw-border/10 rounded-xl px-4 py-3 text-sm font-bold outline-none focus:border-purple-500/40 text-makoclaw-text transition-all"
            >
          </div>

          <div class="lg:col-span-12 xl:col-span-3">
            <button
              class="w-full px-6 py-3 bg-purple-600 hover:bg-purple-500 text-white rounded-xl text-xs font-black uppercase tracking-widest shadow-xl shadow-purple-500/10 transition-all flex items-center justify-center gap-2 active:scale-95 disabled:opacity-50 group/save"
              :disabled="saving"
              @click="saveImageProviderConfig(p.id)"
            >
              <span v-if="saving" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin" />
              <BoltIcon v-else class="w-4 h-4 group-hover/save:translate-y-[-2px] transition-transform" />
              Connect
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Models Config Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showModelsModal"
          class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-xl p-4"
          @click.self="showModelsModal = false"
        >
          <div
            class="glass-panel rounded-2xl p-5 max-w-xl w-full shadow-xl border border-makoclaw-border/50 relative overflow-hidden animate-zoom flex flex-col max-h-[85vh]"
            @click.stop
          >
            <div class="absolute -top-12 -left-12 w-48 h-48 bg-makoclaw-accent/10 blur-[60px] rounded-full pointer-events-none" />

            <div class="flex justify-between items-start mb-5 relative z-10">
              <div>
                <h3 class="text-2xl font-black bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent italic capitalize">
                  {{ selectedProviderName }} Cores
                </h3>
                <p class="text-[10px] font-medium text-makoclaw-text-secondary mt-1">
                  Configure available neural models for this network
                </p>
              </div>
              <button
                class="p-2.5 rounded-xl bg-makoclaw-surface border border-makoclaw-border/10 text-makoclaw-text-secondary hover:text-white transition-all shadow-lg active:scale-95"
                @click="showModelsModal = false"
              >
                <XMarkIcon class="w-5 h-5" />
              </button>
            </div>
            
            <div class="space-y-5 overflow-y-auto pr-4 custom-scrollbar relative z-10 h-full flex flex-col">
              <!-- Add Model -->
              <div class="space-y-3">
                <label class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/60 ml-1">
                  Enlist New Core Model
                </label>
                <div class="flex gap-3">
                  <input
                    v-model="newModelInput"
                    type="text"
                    placeholder="e.g. gpt-4o, claude-3-5-sonnet"
                    class="flex-1 px-6 py-4 bg-makoclaw-bg/60 border-2 border-makoclaw-border/50 rounded-2xl text-sm font-bold text-makoclaw-text focus:border-makoclaw-accent outline-none"
                    @keyup.enter="addModel"
                  >
                  <button
                    class="px-6 py-3 bg-white text-makoclaw-bg rounded-xl text-xs font-black uppercase tracking-widest transition-all hover:bg-white/90 active:scale-95 disabled:opacity-30 shadow-xl shadow-white/5"
                    :disabled="!newModelInput.trim()"
                    @click="addModel"
                  >
                    Add
                  </button>
                </div>
              </div>

              <!-- Model List -->
              <div class="flex-1 min-h-0 flex flex-col pt-4 border-t border-makoclaw-border/30">
                <div class="flex items-center justify-between mb-4">
                  <h4 class="text-[10px] font-medium tracking-wide text-makoclaw-text-secondary/60">
                    Active Core Pool
                  </h4>
                  <span class="px-2 py-0.5 bg-makoclaw-accent/10 text-makoclaw-accent text-[9px] font-black rounded-lg border border-makoclaw-accent/20">
                    {{ editingModels.length }} models
                  </span>
                </div>
                
                <div
                  v-if="editingModels.length === 0"
                  class="flex-1 flex flex-col items-center justify-center py-10 bg-makoclaw-surface/20 rounded-2xl border border-dashed border-makoclaw-border/10"
                >
                  <CpuChipIcon class="w-12 h-12 text-makoclaw-text-secondary/10 mb-4" />
                  <p class="text-sm font-bold text-makoclaw-text-secondary/40 italic">
                    No custom cores configured.
                  </p>
                  <p class="text-[9px] font-medium text-makoclaw-text-secondary/30 mt-1 uppercase tracking-widest">
                    Default parameters will be utilized
                  </p>
                </div>
                
                <div
                  v-else
                  class="grid grid-cols-1 gap-3 flex-1 overflow-y-auto pr-2 custom-scrollbar"
                >
                  <div
                    v-for="(model, idx) in editingModels"
                    :key="idx"
                    class="flex items-center justify-between bg-makoclaw-surface/40 border border-makoclaw-border/50 px-5 py-3.5 rounded-2xl group hover:border-makoclaw-accent/30 transition-all"
                  >
                    <div class="flex items-center gap-3">
                      <span class="w-2 h-2 rounded-full bg-makoclaw-accent/30 group-hover:bg-makoclaw-accent group-hover:shadow-[0_0_8px_rgba(var(--makoclaw-accent-rgb),0.5)] transition-all" />
                      <span class="text-sm font-bold text-makoclaw-text tracking-tight">
                        {{ model }}
                      </span>
                    </div>
                    <button
                      class="p-2 text-makoclaw-text-secondary/40 hover:text-red-400 hover:bg-red-400/10 rounded-xl transition-all opacity-0 group-hover:opacity-100 scale-90 group-hover:scale-100"
                      @click="removeModel(idx)"
                    >
                      <TrashIcon class="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <div class="flex items-center justify-between mt-10 pt-8 border-t border-makoclaw-border/30 relative z-10">
              <button
                class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/40 hover:text-red-400 transition-colors flex items-center gap-2 group"
                @click="resetToDefaults"
              >
                <ArrowPathIcon class="w-3.5 h-3.5 group-hover:rotate-[-45deg] transition-transform" />
                Reset Factory Defaults
              </button>
              <div class="flex gap-4">
                <button
                  class="px-4 py-2 text-xs font-bold uppercase tracking-widest text-makoclaw-text-secondary hover:text-white transition-colors"
                  @click="showModelsModal = false"
                >
                  Cancel
                </button>
                <button
                  class="px-6 py-3 bg-makoclaw-accent text-white rounded-xl text-xs font-black uppercase tracking-widest shadow-xl shadow-makoclaw-accent/10 hover:bg-makoclaw-accent-hover transition-all active:scale-95 disabled:opacity-30 flex items-center justify-center gap-2"
                  :disabled="saving"
                  @click="saveModelsConfig"
                >
                  <span
                    v-if="saving"
                    class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin"
                  />
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
import { ref, computed, watch } from 'vue'
import {
  CpuChipIcon,
  LockClosedIcon,
  BoltIcon,
  XMarkIcon,
  TrashIcon,
  ArrowPathIcon,
  PhotoIcon,
  CheckCircleIcon,
} from '@heroicons/vue/24/outline'

const IMAGE_PROVIDERS = [
  {
    id: 'together',
    label: 'Together.ai',
    sharedKey: null,
    apiKeyUrl: 'https://api.together.xyz/settings/api-keys',
    defaultBase: 'https://api.together.xyz/v1/images/generations',
    description: 'FLUX models via Together inference',
    models: [
      { id: 'black-forest-labs/FLUX.1-schnell-Free', label: 'FLUX.1 Schnell', badge: 'FREE', badgeColor: 'green' },
      { id: 'black-forest-labs/FLUX.1-schnell', label: 'FLUX.1 Schnell (paid)', badge: '$0.003/img', badgeColor: 'gray' },
      { id: 'black-forest-labs/FLUX.1-dev', label: 'FLUX.1 Dev', badge: '$0.025/img', badgeColor: 'gray' },
      { id: 'black-forest-labs/FLUX.1.1-pro', label: 'FLUX1.1 Pro', badge: '$0.04/img', badgeColor: 'gray' },
    ]
  },
  {
    id: 'openai',
    label: 'OpenAI',
    sharedKey: 'openai',
    defaultBase: 'https://api.openai.com/v1/images/generations',
    description: 'Use your existing OpenAI API key',
    models: [
      { id: 'dall-e-3', label: 'DALL·E 3', badge: '$0.04–$0.12/img', badgeColor: 'gray' },
      { id: 'gpt-image-1', label: 'GPT-Image-1', badge: '$0.01–$0.17/img', badgeColor: 'gray' },
      { id: 'dall-e-2', label: 'DALL·E 2', badge: '$0.016/img', badgeColor: 'gray' },
    ]
  },
  {
    id: 'google',
    label: 'Google (Gemini)',
    sharedKey: 'gemini',
    defaultBase: '',
    description: 'Use your existing Gemini API key',
    models: [
      { id: 'gemini-2.0-flash-exp', label: 'Gemini 2.0 Flash', badge: 'Free tier', badgeColor: 'green' },
      { id: 'imagen-3.0-fast-generate-001', label: 'Imagen 3 Fast', badge: '$0.02/img', badgeColor: 'gray' },
      { id: 'imagen-3.0-generate-001', label: 'Imagen 3', badge: '$0.04/img', badgeColor: 'gray' },
    ]
  },
  {
    id: 'fal',
    label: 'fal.ai',
    sharedKey: null,
    apiKeyUrl: 'https://fal.ai/dashboard/keys',
    defaultBase: '',
    description: 'Fast inference, great DX',
    models: [
      { id: 'fal-ai/flux/schnell', label: 'FLUX.1 Schnell', badge: '$0.003/img', badgeColor: 'gray' },
      { id: 'fal-ai/flux/dev', label: 'FLUX.1 Dev', badge: '$0.025/img', badgeColor: 'gray' },
      { id: 'fal-ai/flux-pro/v1.1', label: 'FLUX1.1 Pro', badge: '$0.04/img', badgeColor: 'gray' },
    ]
  },
  {
    id: 'replicate',
    label: 'Replicate',
    sharedKey: null,
    apiKeyUrl: 'https://replicate.com/account/api-tokens',
    defaultBase: '',
    description: 'Thousands of community models',
    models: [
      { id: 'black-forest-labs/flux-schnell', label: 'FLUX.1 Schnell', badge: '~$0.003/img', badgeColor: 'gray' },
      { id: 'black-forest-labs/flux-dev', label: 'FLUX.1 Dev', badge: '~$0.025/img', badgeColor: 'gray' },
      { id: 'stability-ai/sdxl', label: 'SDXL', badge: '~$0.003/img', badgeColor: 'gray' },
    ]
  },
  {
    id: 'zhipu',
    label: 'Zhipu AI (CogView)',
    sharedKey: 'zhipu',
    defaultBase: 'https://open.bigmodel.cn/api/paas/v4/images/generations',
    description: 'CogView image generation via Zhipu',
    models: [
      { id: 'cogview-3-flash', label: 'CogView-3 Flash', badge: 'Low cost', badgeColor: 'green' },
      { id: 'cogview-3-plus', label: 'CogView-3 Plus', badge: 'High quality', badgeColor: 'gray' },
      { id: 'cogview-3', label: 'CogView-3', badge: 'Standard', badgeColor: 'gray' },
    ]
  },
  {
    id: 'bfl',
    label: 'Black Forest Labs',
    sharedKey: null,
    apiKeyUrl: 'https://api.bfl.ml',
    defaultBase: '',
    description: 'Direct FLUX API — best quality',
    models: [
      { id: 'flux-pro-1.1', label: 'FLUX1.1 Pro', badge: '$0.04/img', badgeColor: 'gray' },
      { id: 'flux-pro', label: 'FLUX Pro', badge: '$0.05/img', badgeColor: 'gray' },
      { id: 'flux-dev', label: 'FLUX Dev', badge: '$0.025/img', badgeColor: 'gray' },
    ]
  },
]

const props = defineProps({
  providers: { type: Object, required: true },
  providersList: { type: Array, default: () => [] },
  configData: { type: Object, default: () => null },
  saving: { type: Boolean, default: false }
})

const emit = defineEmits(['save'])

// ── Image Generation config (per-provider) ───────────────────────────────────
const activeImageProvider = ref(props.configData?.tools?.image?.provider || 'openai')

const makeProviderState = () => Object.fromEntries(
  IMAGE_PROVIDERS.map(p => [p.id, { api_key: '', api_base: '' }])
)
const imageProviders = ref(makeProviderState())

// Populate from backend config when loaded
const syncImageProviders = (cfg) => {
  if (!cfg) return
  activeImageProvider.value = cfg.tools?.image?.provider || 'openai'
  IMAGE_PROVIDERS.forEach(p => {
    const backend = cfg.tools?.image_providers?.[p.id]
    if (backend) {
      imageProviders.value[p.id] = { api_key: '', api_base: backend.api_base || '' }
    }
  })
}

syncImageProviders(props.configData)

watch(() => props.configData, syncImageProviders)

const isImageProviderConfigured = (id) => {
  const backend = props.configData?.tools?.image_providers?.[id]
  return backend?.configured === true
}

const saveActiveProvider = () => {
  emit('save', { tools: { image: { provider: activeImageProvider.value } } })
}

const saveImageProviderConfig = (id) => {
  const state = imageProviders.value[id]
  const patch = { api_base: state.api_base }
  if (state.api_key) patch.api_key = state.api_key
  emit('save', { tools: { image_providers: { [id]: patch } } })
  // If no active provider yet, set this one as active
  if (!props.configData?.tools?.image?.provider) {
    activeImageProvider.value = id
    emit('save', { tools: { image: { provider: id } } })
  }
  // Clear key field after save for security
  imageProviders.value[id].api_key = ''
}
// ─────────────────────────────────────────────────────────────────────────────

const showModelsModal = ref(false)
const selectedProviderName = ref('')
const newModelInput = ref('')
const editingModels = ref([])
const isImageModal = ref(false)
const currentImageProviderId = ref('')

const isProviderConfigured = (info) => {
  if (!info || typeof info !== 'object') return false
  if (info.configured === true) return true
  const key = typeof info.api_key === 'string' ? info.api_key.trim() : ''
  return key.length > 0 && !key.includes('••••')
}

const openModelsModal = (name, info) => {
  isImageModal.value = false
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

const openImageProviderModelsModal = (id) => {
  isImageModal.value = true
  currentImageProviderId.value = id
  const pDef = IMAGE_PROVIDERS.find(p => p.id === id)
  selectedProviderName.value = pDef?.label || id
  // Load user's saved models from backend, or fall back to predefined defaults
  const backendModels = props.configData?.tools?.image_providers?.[id]?.models || []
  if (backendModels.length > 0) {
    editingModels.value = [...backendModels]
  } else {
    editingModels.value = (pDef?.models || []).map(m => m.id)
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
  if (confirm('Purge all custom core profiles? Factory defaults will be restored.')) {
    if (isImageModal.value) {
      const pDef = IMAGE_PROVIDERS.find(p => p.id === currentImageProviderId.value)
      editingModels.value = (pDef?.models || []).map(m => m.id)
    } else {
      editingModels.value = []
    }
  }
}

const saveModelsConfig = () => {
  if (isImageModal.value) {
    emit('save', { tools: { image_providers: { [currentImageProviderId.value]: { models: editingModels.value } } } })
  } else {
    emit('save', { providers: { [selectedProviderName.value]: { models: editingModels.value } } })
  }
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
