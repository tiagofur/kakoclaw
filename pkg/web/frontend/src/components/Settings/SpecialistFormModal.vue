<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/80 backdrop-blur-xl" @click.self="close">
        <div class="glass-panel border border-makoclaw-border/50 rounded-2xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col relative animate-zoom" @click.stop>
          <!-- Decorative Background -->
          <div class="absolute -top-24 -right-24 w-64 h-64 bg-makoclaw-accent/10 blur-[80px] rounded-full pointer-events-none group-hover:bg-makoclaw-accent/20 transition-all duration-1000"></div>
          
          <!-- Header -->
          <div class="relative z-10 flex justify-between items-center p-5 pb-3">
            <div>
              <h3 class="text-xl font-semibold text-makoclaw-text flex items-center gap-2">
                <div :class="`p-1.5 rounded-lg bg-gradient-to-br ${mode === 'create' ? 'from-makoclaw-accent to-indigo-600' : 'from-blue-500 to-cyan-500'} shadow-md text-white` ">
                  <IconPlus v-if="mode === 'create'" class="w-4 h-4" />
                  <IconEdit v-else class="w-4 h-4" />
                </div>
                {{ mode === 'create' ? 'Assemble Specialist' : 'Refine Specialist' }}
              </h3>
              <p class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/40 mt-1">
                {{ mode === 'create' ? 'Synthesize new autonomous intelligence' : 'Update existing neural parameters' }}
              </p>
            </div>
            <button @click="close" class="p-2 rounded-xl bg-makoclaw-bg/60 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-white transition-all hover:scale-105 active:scale-90">
              <IconClose class="w-5 h-5" />
            </button>
          </div>

          <div class="flex-1 overflow-y-auto p-5 pt-3 space-y-6 custom-scrollbar relative z-10 text-makoclaw-text">
            <!-- AI Generation Layer -->
            <div v-if="mode === 'create'" class="relative group/ai">
              <div class="absolute -inset-1 bg-gradient-to-r from-makoclaw-accent to-blue-600 rounded-2xl blur opacity-20 group-hover/ai:opacity-40 transition duration-1000"></div>
              <div class="relative bg-makoclaw-surface/40 backdrop-blur-md border border-makoclaw-accent/20 rounded-2xl p-5 overflow-hidden">
                <div class="flex flex-col sm:flex-row items-start gap-4">
                  <div class="w-10 h-10 rounded-xl bg-makoclaw-accent/20 flex items-center justify-center flex-shrink-0 animate-pulse border border-makoclaw-accent/30">
                    <IconBrain class="w-5 h-5 text-makoclaw-accent" />
                  </div>
                  <div class="flex-1 w-full">
                    <div class="flex items-center gap-3 mb-2">
                      <h4 class="text-xs font-semibold uppercase tracking-wide text-makoclaw-text">AI Generate</h4>
                      <span class="px-2 py-0.5 text-[8px] font-medium bg-makoclaw-accent text-white rounded-md">AI-Ready</span>
                    </div>
                    <p class="text-xs font-medium text-makoclaw-text-secondary/70 mb-4">Input specialist requirements. The matrix will auto-configure tools and prompts.</p>
                    
                    <div class="space-y-4">
                      <div class="relative">
                        <textarea
                          v-model="aiPrompt"
                          rows="3"
                          placeholder="Example: 'Create a security specialist for forensic analysis, focused on log patterns and vulnerability detection using shell and grep...'"
                          class="w-full px-4 py-3 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm font-medium focus:border-makoclaw-accent outline-none text-makoclaw-text placeholder:text-makoclaw-text-secondary/30 transition-all resize-none"
                        ></textarea>
                        <div class="absolute bottom-4 right-4 text-[8px] font-black text-makoclaw-text-secondary/20 uppercase">Natural Language Interface</div>
                      </div>
                      
                      <button
                        @click="generateWithAI"
                        :disabled="!aiPrompt.trim() || agentsStore.aiGenerating"
                        class="w-full px-5 py-2.5 bg-gradient-to-r from-makoclaw-accent to-blue-600 hover:from-makoclaw-accent-hover hover:to-blue-700 text-white rounded-xl font-semibold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 disabled:opacity-30 disabled:cursor-not-allowed group/btn"
                      >
                        <IconLoader v-if="agentsStore.aiGenerating" class="w-5 h-5 animate-spin" />
                        <IconSparkles v-else class="w-5 h-5 group-hover/btn:scale-125 transition-transform" />
                        {{ agentsStore.aiGenerating ? 'Syncing...' : 'Synthesize Configuration' }}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="flex items-center gap-4 py-2">
              <div class="h-px bg-makoclaw-border/40 flex-1"></div>
              <span class="text-[9px] font-medium tracking-wide text-makoclaw-text-secondary/30 whitespace-nowrap">Neural Parameters</span>
              <div class="h-px bg-makoclaw-border/40 flex-1"></div>
            </div>

            <!-- Manual Config -->
            <div class="space-y-6">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
                <div class="space-y-2">
                  <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-2">Identity Signature *</label>
                  <input
                    v-model="form.name"
                    type="text"
                    placeholder="e.g., protocol_alpha"
                    class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all placeholder:opacity-20"
                  />
                </div>

                <div class="space-y-2">
                  <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-2">Provider Node</label>
                  <select
                    v-model="form.provider"
                    class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer"
                  >
                    <option v-for="p in providersList" :key="p.name" :value="p.name">{{ p.name.toUpperCase() }}</option>
                  </select>
                </div>
              </div>

              <div class="space-y-2">
                <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-2">Intelligence Model</label>
                <select
                  v-model="form.model"
                  class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all cursor-pointer"
                >
                  <optgroup v-for="p in providersList" :key="p.name" :label="p.name.toUpperCase()" class="text-makoclaw-text-secondary">
                    <option v-for="m in p.models" :key="m.id" :value="m.id">{{ m.id }}</option>
                  </optgroup>
                </select>
              </div>

              <div class="space-y-2">
                <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-2">Function Description *</label>
                <input
                  v-model="form.description"
                  type="text"
                  placeholder="Primary objective of this specialist agent..."
                  class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none transition-all"
                />
              </div>

              <div class="space-y-4">
                <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-2 flex items-center gap-2">
                  <IconTerminal class="w-3.5 h-3.5" /> Static Directives (System Prompt)
                </label>
                <div class="relative group">
                  <textarea
                    v-model="form.prompt"
                    rows="5"
                    placeholder="Provide specific personality benchmarks and behavioral directives..."
                    class="w-full px-8 py-6 bg-makoclaw-bg selection:bg-makoclaw-accent selection:text-white border-2 border-makoclaw-border/40 rounded-2xl text-xs font-black font-mono text-makoclaw-text-secondary leading-relaxed focus:border-makoclaw-accent/60 outline-none transition-all resize-none shadow-xl"
                  ></textarea>
                  <div class="absolute bottom-4 right-4 text-[8px] font-medium text-makoclaw-text-secondary/10 pointer-events-none">Core Instructions</div>
                </div>
              </div>

              <!-- Params Matrix -->
              <div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
                <div class="space-y-2">
                  <label class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">Token Ceiling</label>
                  <input v-model.number="form.max_tokens" type="number" class="w-full px-5 py-3 bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-xl text-xs font-bold text-center text-makoclaw-text outline-none focus:border-makoclaw-accent" />
                </div>
                <div class="space-y-2">
                  <label class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">Fluidity (Temp)</label>
                  <input v-model.number="form.temperature" type="number" step="0.1" min="0" max="2" class="w-full px-5 py-3 bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-xl text-xs font-bold text-center text-makoclaw-text outline-none focus:border-makoclaw-accent" />
                </div>
                <div class="space-y-2">
                  <label class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-1">Recurse Limit</label>
                  <input v-model.number="form.max_tool_iterations" type="number" class="w-full px-5 py-3 bg-makoclaw-bg/40 border-2 border-makoclaw-border/50 rounded-xl text-xs font-bold text-center text-makoclaw-text outline-none focus:border-makoclaw-accent" />
                </div>
              </div>

              <!-- Tools Matrix -->
              <div class="space-y-6">
                <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-2">Accessible Neural Sockets (Tools)</label>
                <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3 p-6 bg-makoclaw-bg/30 rounded-2xl border border-makoclaw-border/50 shadow-inner">
                  <label
                    v-for="tool in availableTools"
                    :key="tool"
                    class="group/tool relative flex items-center gap-3 px-4 py-3 rounded-2xl bg-makoclaw-surface border border-makoclaw-border/50 hover:border-makoclaw-accent/50 cursor-pointer transition-all hover:bg-makoclaw-bg active:scale-95"
                    :class="form.tools.includes(tool) ? 'border-makoclaw-accent shadow-[0_4px_12px_rgba(var(--makoclaw-accent-rgb),0.2)] bg-makoclaw-accent/[0.03]' : ''"
                  >
                    <input
                      v-model="form.tools"
                      :value="tool"
                      type="checkbox"
                      class="w-4 h-4 rounded border-2 border-makoclaw-border bg-transparent text-makoclaw-accent focus:ring-makoclaw-accent checked:border-makoclaw-accent"
                    />
                    <span class="text-[11px] font-bold text-makoclaw-text-secondary group-hover/tool:text-makoclaw-text transition-colors" :class="form.tools.includes(tool) ? '!text-makoclaw-accent' : ''">{{ tool }}</span>
                  </label>
                </div>
              </div>

              <div class="space-y-2">
                <label class="text-[10px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/60 ml-2 italic">Matching Patterns (Keywords)</label>
                <div class="relative">
                  <input
                    v-model="keywordsInput"
                    type="text"
                    placeholder="forensics, logs, security, linux..."
                    class="w-full bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl px-4 py-2.5 text-sm font-medium text-makoclaw-text focus:border-makoclaw-accent outline-none placeholder:opacity-10"
                  />
                  <div class="absolute right-6 top-1/2 -translate-y-1/2 text-[8px] font-black text-makoclaw-text-secondary/20 uppercase">Comma separated</div>
                </div>
              </div>
            </div>
          </div>

          <!-- Footer Actions -->
          <div class="relative z-10 flex flex-col sm:flex-row justify-between items-center gap-4 p-8 border-t border-makoclaw-border/30 bg-makoclaw-bg/20 backdrop-blur-md">
            <button
              @click="close"
              class="w-full sm:w-auto px-5 py-2.5 text-xs font-medium text-makoclaw-text-secondary/60 hover:text-white transition-all order-2 sm:order-1"
            >
              Abort Mission
            </button>
            
            <div class="flex flex-col sm:flex-row gap-3 w-full sm:w-auto order-1 sm:order-2">
              <button
                v-if="mode === 'edit' && specialist"
                @click="testSpecialist"
                :disabled="agentsStore.loading"
                class="px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-xl font-semibold shadow-lg shadow-blue-500/20 transition-all flex items-center justify-center gap-2 group/test active:scale-95 disabled:opacity-30"
              >
                <IconLoader v-if="agentsStore.loading" class="w-4 h-4 animate-spin" />
                <IconPlay v-else class="w-4 h-4 group-hover/test:scale-125 transition-transform" />
                Run Diagnostics
              </button>
              
              <button
                @click="save"
                :disabled="!isFormValid || saving"
                class="px-5 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl font-semibold shadow-lg shadow-makoclaw-accent/20 transition-all flex items-center justify-center gap-2 active:scale-95 disabled:opacity-30 group/save"
              >
                <IconLoader v-if="saving" class="w-4 h-4 animate-spin" />
                <IconCheck v-else class="w-4 h-4 group-hover/save:scale-110 transition-transform" />
                {{ mode === 'create' ? 'Finalize Synthesis' : 'Commit Changes' }}
              </button>
            </div>
          </div>
        </div>

        <!-- Test Result Backdrop Inner Modal -->
        <Transition name="modal">
          <div v-if="showTestResultModal" class="fixed inset-0 z-[110] flex items-center justify-center p-8 bg-black/90 backdrop-blur-3xl" @click.self="showTestResultModal = false">
            <div class="glass-panel border border-blue-500/30 rounded-[3rem] shadow-[0_0_50px_rgba(59,130,246,0.2)] w-full max-w-3xl max-h-[80vh] overflow-hidden flex flex-col animate-zoom" @click.stop>
              <div class="p-10 pb-4 flex justify-between items-center">
                <div>
                  <h3 class="text-2xl font-black text-white italic tracking-tight flex items-center gap-3">
                    <IconActivity class="w-6 h-6 text-blue-400" />
                    Neural Diagnostic Output
                  </h3>
                  <p class="text-[9px] font-medium tracking-wide text-blue-400/50 mt-1">Real-time response verification</p>
                </div>
                <button @click="showTestResultModal = false" class="p-3 rounded-2xl bg-white/5 border border-white/10 text-white/40 hover:text-white transition-all">
                  <IconClose class="w-6 h-6" />
                </button>
              </div>
              <div class="flex-1 overflow-y-auto p-10 pt-4 custom-scrollbar">
                <div class="relative group">
                  <div class="absolute -inset-1 bg-blue-500/20 rounded-2xl blur group-hover:bg-blue-500/30 transition duration-1000"></div>
                  <div class="relative bg-black/40 border-2 border-white/10 rounded-2xl p-8">
                     <pre class="text-sm font-black font-mono text-blue-100/90 leading-relaxed whitespace-pre-wrap selection:bg-blue-600 selection:text-white">{{ testResult }}</pre>
                  </div>
                </div>
              </div>
              <div class="p-10 pt-4 bg-white/5 border-t border-white/10">
                <button @click="showTestResultModal = false" class="w-full px-5 py-2.5 bg-gradient-to-r from-blue-600 to-indigo-700 text-white rounded-xl font-semibold shadow-lg transition-all active:scale-95">
                  Acknowledge Report
                </button>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, h } from 'vue'
import { useAgentsStore } from '../../stores/agentsStore'
import { useToast } from '../../composables/useToast'

// Premium Icons
const IconPlus = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M12 6v6m0 0v6m0-6h6m-6 0H6' })]) }
const IconEdit = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z' })]) }
const IconClose = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M6 18L18 6M6 6l12 12' })]) }
const IconBrain = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z' })]) }
const IconSparkles = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M13 10V3L4 14h7v7l9-11h-7z' })]) }
const IconTerminal = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 00-2 2z' })]) }
const IconPlay = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z' }), h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M21 12a9 9 0 11-18 0 9 9 0 0118 0z' })]) }
const IconCheck = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M5 13l4 4L19 7' })]) }
const IconActivity = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M13 10V3L4 14h7v7l9-11h-7z' })]) }
const IconLoader = { render: () => h('svg', { fill: 'none', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('circle', { class: 'opacity-25', cx: '12', cy: '12', r: '10', stroke: 'currentColor', 'stroke-width': '4' }), h('path', { class: 'opacity-75', fill: 'currentColor', d: 'M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z' })]) }

const props = defineProps({
  show: { type: Boolean, required: true },
  mode: { type: String, default: 'create' },
  specialist: { type: Object, default: null },
  providersList: { type: Array, required: true },
  availableTools: { type: Array, default: () => ['read_file', 'write_file', 'edit_file', 'list_dir', 'exec', 'web_search', 'web_fetch'] }
})

const emit = defineEmits(['close', 'save'])
const toast = useToast()
const agentsStore = useAgentsStore()
const saving = ref(false)
const aiPrompt = ref('')
const testResult = ref(null)
const showTestResultModal = ref(false)

const form = ref({
  name: '', description: '', prompt: '', provider: 'anthropic', model: 'claude-opus',
  max_tokens: 8192, temperature: 0.7, max_tool_iterations: 20, tools: [], keywords: []
})

const keywordsInput = computed({
  get: () => form.value.keywords.join(', '),
  set: (val) => { form.value.keywords = val.split(',').map(k => k.trim()).filter(k => k) }
})

const isFormValid = computed(() => form.value.name.trim() && form.value.description.trim() && form.value.provider && form.value.model)

watch(() => props.specialist, (s) => { if (s && props.mode === 'edit') form.value = { ...s } }, { immediate: true })
watch(() => props.show, (s) => { if (s && props.mode === 'create') resetForm() })

function resetForm() {
  form.value = { 
    name: '', description: '', prompt: '', provider: 'anthropic', model: 'claude-opus',
    max_tokens: 8192, temperature: 0.7, max_tool_iterations: 20, tools: [], keywords: [] 
  }
  aiPrompt.value = ''
}

async function generateWithAI() {
  if (!aiPrompt.value.trim()) return
  const gen = await agentsStore.generateSpecialistWithAI(aiPrompt.value)
  if (gen) { form.value = { ...gen }; toast.success('Neural mapping synthesized') }
}

async function testSpecialist() {
  if (!props.specialist?.name) return
  const res = await agentsStore.testSpecialist(props.specialist.name, 'Initialize self-diagnostic. Report status.')
  if (res) { testResult.value = res; showTestResultModal.value = true; toast.success('Diagnostic complete') }
}

async function save() {
  if (!isFormValid.value) return
  saving.value = true
  try {
    const success = props.mode === 'create' ? await agentsStore.createSpecialist(form.value) : await agentsStore.updateSpecialist(props.specialist.name, form.value)
    if (success) { emit('save'); close() }
  } finally { saving.value = false }
}

function close() { emit('close'); resetForm() }
</script>

<style scoped>
@keyframes zoom { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }
.animate-zoom { animation: zoom 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards; }
.modal-enter-active, .modal-leave-active { transition: opacity 0.3s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(var(--makoclaw-accent-rgb), 0.1); border-radius: 20px; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background: rgba(var(--makoclaw-accent-rgb), 0.3); }
</style>
