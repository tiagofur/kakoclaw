<template>
  <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
    <div class="bg-makoclaw-surface rounded-2xl shadow-2xl w-full max-w-3xl max-h-[90vh] overflow-hidden border border-makoclaw-border animate-in fade-in zoom-in duration-200 flex flex-col">
      <!-- Header -->
      <div class="flex justify-between items-center p-5 border-b border-makoclaw-border bg-makoclaw-bg/20">
        <h3 class="text-lg font-bold text-makoclaw-text flex items-center gap-2">
          <svg v-if="mode === 'create'" class="w-5 h-5 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          <svg v-else class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
          {{ mode === 'create' ? 'Create New Specialist' : 'Edit Specialist' }}
        </h3>
        <button @click="close" class="text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">
          <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="flex-1 overflow-y-auto p-5 space-y-6 custom-scrollbar">
        <!-- AI Assistant Section (Create Mode Only) -->
        <div v-if="mode === 'create'" class="bg-gradient-to-r from-makoclaw-accent/10 to-blue-500/10 border border-makoclaw-accent/30 rounded-xl p-5">
          <div class="flex items-start gap-3">
            <div class="w-10 h-10 rounded-lg bg-makoclaw-accent/20 flex items-center justify-center flex-shrink-0">
              <svg class="w-5 h-5 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div class="flex-1">
              <h4 class="font-bold text-makoclaw-text mb-1 flex items-center gap-2">
                Create Agent with AI
                <span class="px-2 py-0.5 text-[10px] font-bold uppercase bg-makoclaw-accent text-white rounded-full">New</span>
              </h4>
              <p class="text-sm text-makoclaw-text-secondary mb-3">Describe your specialist and AI will automatically configure its settings, tools, and parameters.</p>
              <div class="space-y-3">
                <textarea
                  v-model="aiPrompt"
                  rows="3"
                  placeholder="e.g., 'Create a Python specialist for data analysis and visualization using pandas and matplotlib. Should be able to execute code and generate charts.'"
                  class="w-full px-4 py-3 bg-makoclaw-bg/60 border border-makoclaw-border rounded-xl text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none text-makoclaw-text resize-none backdrop-blur-sm"
                ></textarea>
                <button
                  @click="generateWithAI"
                  :disabled="!aiPrompt.trim() || agentsStore.aiGenerating"
                  class="w-full px-4 py-2.5 bg-gradient-to-r from-makoclaw-accent to-blue-600 hover:from-makoclaw-accent-hover hover:to-blue-700 text-white rounded-xl font-bold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98]"
                >
                  <svg v-if="agentsStore.aiGenerating" class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  {{ agentsStore.aiGenerating ? 'Generating Configuration...' : 'Generate with AI' }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="h-px bg-makoclaw-border my-2"></div>

        <!-- Manual Configuration Form -->
        <div class="space-y-5">
          <h4 class="font-bold text-makoclaw-text flex items-center gap-2">
            <svg class="w-4 h-4 text-makoclaw-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            </svg>
            Specialist Configuration
          </h4>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
            <!-- Name -->
            <div>
              <label class="block text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Name *</label>
              <input
                v-model="form.name"
                type="text"
                placeholder="e.g., python_expert"
                class="w-full px-4 py-2.5 bg-makoclaw-bg/60 border border-makoclaw-border rounded-xl text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none text-makoclaw-text backdrop-blur-sm"
              />
            </div>

            <!-- Provider -->
            <div>
              <label class="block text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Provider</label>
              <select
                v-model="form.provider"
                class="w-full px-4 py-2.5 bg-makoclaw-bg/60 border border-makoclaw-border rounded-xl text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none text-makoclaw-text cursor-pointer backdrop-blur-sm"
              >
                <option v-for="p in providersList" :key="p.name" :value="p.name">{{ p.name }}</option>
              </select>
            </div>
          </div>

          <!-- Model -->
          <div>
            <label class="block text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Model</label>
            <select
              v-model="form.model"
              class="w-full px-4 py-2.5 bg-makoclaw-bg/60 border border-makoclaw-border rounded-xl text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none text-makoclaw-text cursor-pointer backdrop-blur-sm"
            >
              <optgroup v-for="p in providersList" :key="p.name" :label="p.name">
                <option v-for="m in p.models" :key="m.id" :value="m.id">{{ m.id }}</option>
              </optgroup>
            </select>
          </div>

          <!-- Description -->
          <div>
            <label class="block text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Description *</label>
            <input
              v-model="form.description"
              type="text"
              placeholder="e.g., Python expert for data analysis and visualization"
              class="w-full px-4 py-2.5 bg-makoclaw-bg/60 border border-makoclaw-border rounded-xl text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none text-makoclaw-text backdrop-blur-sm"
            />
          </div>

          <!-- Custom Prompt -->
          <div>
            <label class="block text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">System Prompt</label>
            <textarea
              v-model="form.prompt"
              rows="4"
              placeholder="You are an expert Python developer specializing in data analysis..."
              class="w-full px-4 py-3 bg-makoclaw-bg/60 border border-makoclaw-border rounded-xl text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none text-makoclaw-text resize-none backdrop-blur-sm"
            ></textarea>
          </div>

          <!-- Parameters Grid -->
          <div class="grid grid-cols-3 gap-4">
            <div>
              <label class="block text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Max Tokens</label>
              <input
                v-model.number="form.max_tokens"
                type="number"
                class="w-full px-4 py-2.5 bg-makoclaw-bg/60 border border-makoclaw-border rounded-xl text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none text-makoclaw-text backdrop-blur-sm"
              />
            </div>
            <div>
              <label class="block text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Temperature</label>
              <input
                v-model.number="form.temperature"
                type="number"
                step="0.1"
                min="0"
                max="2"
                class="w-full px-4 py-2.5 bg-makoclaw-bg/60 border border-makoclaw-border rounded-xl text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none text-makoclaw-text backdrop-blur-sm"
              />
            </div>
            <div>
              <label class="block text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Max Tool Iterations</label>
              <input
                v-model.number="form.max_tool_iterations"
                type="number"
                class="w-full px-4 py-2.5 bg-makoclaw-bg/60 border border-makoclaw-border rounded-xl text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none text-makoclaw-text backdrop-blur-sm"
              />
            </div>
          </div>

          <!-- Tools Selection -->
          <div>
            <label class="block text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-3">Allowed Tools</label>
            <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2 p-4 bg-makoclaw-bg/40 rounded-xl border border-makoclaw-border">
              <label
                v-for="tool in availableTools"
                :key="tool"
                class="flex items-center gap-2 px-3 py-2 rounded-lg bg-makoclaw-bg/60 hover:bg-makoclaw-bg transition-colors cursor-pointer"
              >
                <input
                  v-model="form.tools"
                  :value="tool"
                  type="checkbox"
                  class="rounded border-makoclaw-border bg-makoclaw-surface text-makoclaw-accent focus:ring-makoclaw-accent"
                />
                <span class="text-sm text-makoclaw-text">{{ tool }}</span>
              </label>
            </div>
          </div>

          <!-- Keywords -->
          <div>
            <label class="block text-xs font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Keywords (comma separated)</label>
            <input
              v-model="keywordsInput"
              type="text"
              placeholder="python, pandas, data, analysis, charts"
              class="w-full px-4 py-2.5 bg-makoclaw-bg/60 border border-makoclaw-border rounded-xl text-sm focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent outline-none text-makoclaw-text backdrop-blur-sm"
            />
            <p class="text-[10px] text-makoclaw-text-secondary mt-1">These keywords help the orchestrator match tasks to this specialist</p>
          </div>
        </div>
      </div>

      <!-- Footer Actions -->
      <div class="flex justify-end gap-3 p-5 border-t border-makoclaw-border bg-makoclaw-bg/20">
        <button
          @click="close"
          class="px-5 py-2.5 text-sm font-medium text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
        >
          Cancel
        </button>
        <button
          v-if="mode === 'edit' && specialist"
          @click="testSpecialist"
          :disabled="agentsStore.loading"
          class="px-5 py-2.5 text-sm font-bold bg-blue-600 hover:bg-blue-700 text-white rounded-xl shadow-lg shadow-blue-500/20 transition-all flex items-center gap-2 disabled:opacity-50"
        >
          <svg v-if="agentsStore.loading" class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          Test
        </button>
        <button
          @click="save"
          :disabled="!isFormValid || saving"
          class="px-6 py-2.5 text-sm font-bold bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed active:scale-[0.98]"
        >
          <span v-if="saving" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
          {{ mode === 'create' ? 'Create Specialist' : 'Save Changes' }}
        </button>
      </div>
    </div>

    <!-- Test Result Modal -->
    <div v-if="showTestResultModal" class="fixed inset-0 z-[60] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="bg-makoclaw-surface rounded-2xl shadow-2xl w-full max-w-2xl max-h-[80vh] overflow-hidden border border-makoclaw-border animate-in fade-in zoom-in duration-200 flex flex-col">
        <div class="flex justify-between items-center p-5 border-b border-makoclaw-border bg-makoclaw-bg/20">
          <h3 class="text-lg font-bold text-makoclaw-text flex items-center gap-2">
            <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Test Result
          </h3>
          <button @click="showTestResultModal = false" class="text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">
            <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="flex-1 overflow-y-auto p-5">
          <pre class="bg-makoclaw-bg rounded-xl p-4 text-sm text-makoclaw-text whitespace-pre-wrap custom-scrollbar">{{ testResult }}</pre>
        </div>
        <div class="p-5 border-t border-makoclaw-border bg-makoclaw-bg/20">
          <button @click="showTestResultModal = false" class="w-full px-4 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl font-bold shadow-lg shadow-makoclaw-accent/20 transition-all">
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useAgentsStore } from '../../stores/agentsStore'
import { useToast } from '../../composables/useToast'

const toast = useToast()
const agentsStore = useAgentsStore()

const props = defineProps({
  show: {
    type: Boolean,
    required: true
  },
  mode: {
    type: String,
    default: 'create'
  },
  specialist: {
    type: Object,
    default: null
  },
  providersList: {
    type: Array,
    required: true
  },
  availableTools: {
    type: Array,
    default: () => ['file_read', 'file_write', 'execute_shell', 'list_dir', 'web_search', 'git', 'database']
  }
})

const emit = defineEmits(['close', 'save'])

const saving = ref(false)
const aiPrompt = ref('')
const testResult = ref(null)
const showTestResultModal = ref(false)

const form = ref({
  name: '',
  description: '',
  prompt: '',
  provider: 'anthropic',
  model: 'claude-opus',
  max_tokens: 8192,
  temperature: 0.7,
  max_tool_iterations: 20,
  tools: [],
  keywords: []
})

const keywordsInput = computed({
  get: () => form.value.keywords.join(', '),
  set: (val) => {
    form.value.keywords = val.split(',').map(k => k.trim()).filter(k => k)
  }
})

const isFormValid = computed(() => {
  return form.value.name.trim() &&
         form.value.description.trim() &&
         form.value.provider &&
         form.value.model
})

watch(() => props.specialist, (newSpecialist) => {
  if (newSpecialist && props.mode === 'edit') {
    form.value = { ...newSpecialist }
  }
}, { immediate: true })

watch(() => props.show, (newVal) => {
  if (newVal && props.mode === 'create') {
    resetForm()
  }
})

function resetForm() {
  form.value = {
    name: '',
    description: '',
    prompt: '',
    provider: 'anthropic',
    model: 'claude-opus',
    max_tokens: 8192,
    temperature: 0.7,
    max_tool_iterations: 20,
    tools: [],
    keywords: []
  }
  aiPrompt.value = ''
}

async function generateWithAI() {
  if (!aiPrompt.value.trim()) return

  const generated = await agentsStore.generateSpecialistWithAI(aiPrompt.value)
  if (generated) {
    form.value = { ...generated }
    toast.success('Specialist configuration generated!')
  }
}

async function testSpecialist() {
  if (!props.specialist?.name) return

  const result = await agentsStore.testSpecialist(props.specialist.name, 'Hello, please introduce yourself and describe your capabilities.')
  if (result) {
    testResult.value = result
    showTestResultModal.value = true
    toast.success('Test completed!')
  }
}

async function save() {
  if (!isFormValid.value) return

  saving.value = true

  try {
    let success
    if (props.mode === 'create') {
      success = await agentsStore.createSpecialist(form.value)
    } else {
      success = await agentsStore.updateSpecialist(props.specialist.name, form.value)
    }

    if (success) {
      emit('save')
      close()
    }
  } finally {
    saving.value = false
  }
}

function close() {
  emit('close')
  resetForm()
}
</script>


