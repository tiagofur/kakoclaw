<template>
  <div class="h-full flex flex-col bg-kakoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-kakoclaw-border bg-kakoclaw-surface flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold bg-gradient-to-r from-kakoclaw-accent to-purple-500 bg-clip-text text-transparent">
          {{ editing ? (editingWorkflow.id ? 'Edit Workflow' : 'New Workflow') : 'Workflows' }}
        </h2>
        <p class="text-sm text-kakoclaw-text-secondary mt-1">
          {{ editing ? 'Visual pipeline builder' : 'Create and manage automation pipelines' }}
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button v-if="editing" @click="cancelEdit"
          class="flex items-center gap-2 px-4 py-2 text-kakoclaw-text-secondary hover:text-kakoclaw-text border border-kakoclaw-border rounded-lg transition-colors text-sm">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
          Cancel
        </button>
        <button v-if="editing" @click="saveWorkflow" :disabled="saving"
          class="flex items-center gap-2 px-4 py-2 bg-kakoclaw-accent text-white rounded-lg hover:bg-kakoclaw-accent/90 transition-colors text-sm disabled:opacity-50">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
          {{ saving ? 'Saving...' : 'Save' }}
        </button>
        <button v-if="!editing" @click="startCreate"
          class="flex items-center gap-2 px-4 py-2 bg-kakoclaw-accent text-white rounded-lg hover:bg-kakoclaw-accent/90 transition-colors text-sm">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
          New Workflow
        </button>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6 custom-scrollbar">
      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
      </div>

      <!-- ===== LIST VIEW ===== -->
      <template v-else-if="!editing">
        <div v-if="workflows.length === 0" class="text-center py-12 text-kakoclaw-text-secondary">
          <svg class="w-16 h-16 mx-auto mb-4 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          <p class="text-lg">No workflows yet</p>
          <p class="text-sm mt-2">Create a workflow to automate multi-step pipelines</p>
        </div>

        <div class="space-y-3">
          <div v-for="wf in workflows" :key="wf.id"
            class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5 hover:border-kakoclaw-accent/30 transition-colors">
            <div class="flex items-start justify-between">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <h3 class="font-semibold">{{ wf.name }}</h3>
                  <span class="px-2 py-0.5 text-xs rounded-full"
                    :class="wf.enabled ? 'bg-emerald-500/10 text-emerald-400' : 'bg-gray-500/10 text-gray-400'">
                    {{ wf.enabled ? 'Enabled' : 'Disabled' }}
                  </span>
                </div>
                <p v-if="wf.description" class="text-sm text-kakoclaw-text-secondary mt-1">{{ wf.description }}</p>
              </div>
            </div>

            <div class="flex items-center gap-4 mt-3 text-xs text-kakoclaw-text-secondary">
              <span>{{ countSteps(wf) }} steps</span>
              <span>Created {{ formatDate(wf.created_at) }}</span>
            </div>

            <div class="flex items-center gap-2 mt-3">
              <button @click="startEdit(wf)"
                class="px-3 py-1.5 text-xs bg-kakoclaw-accent/10 text-kakoclaw-accent rounded-lg hover:bg-kakoclaw-accent/20 transition-colors">
                Edit
              </button>
              <button @click="runWorkflow(wf)" :disabled="runningId === wf.id"
                class="px-3 py-1.5 text-xs bg-emerald-500/10 text-emerald-400 rounded-lg hover:bg-emerald-500/20 transition-colors disabled:opacity-50">
                {{ runningId === wf.id ? 'Running...' : 'Run' }}
              </button>
              <button @click="showRuns(wf)"
                class="px-3 py-1.5 text-xs bg-blue-500/10 text-blue-400 rounded-lg hover:bg-blue-500/20 transition-colors">
                History
              </button>
              <button @click="deleteWorkflow(wf)"
                class="px-3 py-1.5 text-xs text-red-400 bg-red-500/10 rounded-lg hover:bg-red-500/20 transition-colors">
                Delete
              </button>
            </div>

            <!-- Inline Run Results -->
            <div v-if="wf._lastResults" class="mt-4 border-t border-kakoclaw-border pt-3">
              <h4 class="text-xs font-semibold text-kakoclaw-text-secondary mb-2">Last Run Results</h4>
              <div class="space-y-2">
                <div v-for="(res, i) in wf._lastResults" :key="i"
                  class="text-xs p-2 rounded-lg bg-kakoclaw-bg border border-kakoclaw-border">
                  <div class="flex items-center gap-2 mb-1">
                    <span class="font-mono font-semibold">{{ res.label || res.step_id }}</span>
                    <span class="px-1.5 py-0.5 rounded text-[10px]"
                      :class="res.error ? 'bg-red-500/10 text-red-400' : res.skipped ? 'bg-gray-500/10 text-gray-400' : 'bg-emerald-500/10 text-emerald-400'">
                      {{ res.error ? 'Error' : res.skipped ? 'Skipped' : 'OK' }}
                    </span>
                    <span v-if="res.duration_ms" class="text-kakoclaw-text-secondary">{{ res.duration_ms }}ms</span>
                  </div>
                  <pre v-if="res.output" class="whitespace-pre-wrap text-kakoclaw-text-secondary max-h-24 overflow-auto">{{ res.output }}</pre>
                  <pre v-if="res.error" class="whitespace-pre-wrap text-red-400 max-h-24 overflow-auto">{{ res.error }}</pre>
                </div>
              </div>
            </div>

            <!-- Inline Runs History -->
            <div v-if="wf._runs" class="mt-4 border-t border-kakoclaw-border pt-3">
              <div class="flex items-center justify-between mb-2">
                <h4 class="text-xs font-semibold text-kakoclaw-text-secondary">Recent Runs</h4>
                <button @click="wf._runs = null" class="text-xs text-kakoclaw-text-secondary hover:text-kakoclaw-text">Hide</button>
              </div>
              <div v-if="wf._runs.length === 0" class="text-xs text-kakoclaw-text-secondary">No runs yet</div>
              <div class="space-y-1">
                <div v-for="run in wf._runs" :key="run.id"
                  class="text-xs flex items-center gap-3 p-2 rounded bg-kakoclaw-bg">
                  <span class="px-1.5 py-0.5 rounded"
                    :class="run.status === 'completed' ? 'bg-emerald-500/10 text-emerald-400' : run.status === 'running' ? 'bg-blue-500/10 text-blue-400' : 'bg-red-500/10 text-red-400'">
                    {{ run.status }}
                  </span>
                  <span class="text-kakoclaw-text-secondary">{{ formatDate(run.started_at) }}</span>
                  <span v-if="run.finished_at" class="text-kakoclaw-text-secondary">
                    ({{ runDuration(run) }})
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- ===== EDITOR VIEW ===== -->
      <template v-else>
        <!-- Workflow Meta -->
        <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5 mb-6">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1">Name</label>
              <input v-model="editingWorkflow.name" type="text" placeholder="My Workflow"
                class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent" />
            </div>
            <div>
              <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1">Description</label>
              <input v-model="editingWorkflow.description" type="text" placeholder="What does this workflow do?"
                class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent" />
            </div>
          </div>
        </div>

        <!-- Pipeline Steps -->
        <div class="mb-4">
          <h3 class="text-sm font-semibold text-kakoclaw-text-secondary mb-3">Pipeline Steps</h3>

          <VueDraggable v-model="editingWorkflow.steps" :animation="200" handle=".drag-handle"
            class="space-y-3">
            <template #item="{ element: step, index }">
              <div :key="step.id" class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl overflow-hidden"
                :class="expandedStep === step.id ? 'ring-1 ring-kakoclaw-accent/50' : ''">
                <!-- Step Header -->
                <div class="flex items-center gap-3 px-4 py-3 cursor-pointer" @click="toggleStep(step.id)">
                  <div class="drag-handle cursor-grab text-kakoclaw-text-secondary hover:text-kakoclaw-text">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8h16M4 16h16" />
                    </svg>
                  </div>
                  <div class="flex items-center gap-2 flex-1 min-w-0">
                    <span class="w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-bold"
                      :class="stepTypeClass(step.type)">
                      {{ index + 1 }}
                    </span>
                    <span class="px-2 py-0.5 text-[10px] rounded font-semibold uppercase tracking-wider"
                      :class="stepTypeClass(step.type)">
                      {{ step.type }}
                    </span>
                    <span class="text-sm truncate">{{ step.label || 'Untitled step' }}</span>
                  </div>
                  <button @click.stop="removeStep(index, step.id)"
                    class="p-1 text-kakoclaw-text-secondary hover:text-red-400 transition-colors">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                  <svg class="w-4 h-4 text-kakoclaw-text-secondary transition-transform" :class="expandedStep === step.id ? 'rotate-180' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                  </svg>
                </div>

                <!-- Step Config (expanded) -->
                <div v-if="expandedStep === step.id" class="px-4 pb-4 border-t border-kakoclaw-border pt-3 space-y-3">
                  <div>
                    <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1">Label</label>
                    <input v-model="step.label" type="text" placeholder="Step label"
                      class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent" />
                  </div>

                  <div>
                    <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1">On Error</label>
                    <select v-model="step.on_error"
                      class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent">
                      <option value="stop">Stop workflow</option>
                      <option value="continue">Continue to next step</option>
                    </select>
                  </div>

                  <!-- Prompt Config -->
                  <template v-if="step.type === 'prompt'">
                    <div>
                      <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1">
                        Message
                        <span class="font-normal opacity-60">(supports {{ templateHint }} templates)</span>
                      </label>
                      <textarea v-model="step._config.message" rows="4" placeholder="Enter prompt message..."
                        class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent font-mono resize-y"></textarea>
                    </div>
                    <div>
                      <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1">Model Override (optional)</label>
                      <input v-model="step._config.model" type="text" placeholder="Leave empty for default"
                        class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent" />
                    </div>
                  </template>

                  <!-- Tool Config -->
                  <template v-if="step.type === 'tool'">
                    <div>
                      <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1">Tool Name</label>
                      <select v-if="availableTools.length > 0" v-model="step._config.tool_name"
                        class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent">
                        <option value="">Select a tool...</option>
                        <option v-for="t in availableTools" :key="t" :value="t">{{ t }}</option>
                      </select>
                      <input v-else v-model="step._config.tool_name" type="text" placeholder="tool_name"
                        class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent" />
                    </div>
                    <div>
                      <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1">
                        Arguments (JSON)
                        <span class="font-normal opacity-60">(string values support {{ templateHint }} templates)</span>
                      </label>
                      <textarea v-model="step._config._argsJson" rows="4" placeholder='{"key": "value"}'
                        class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent font-mono resize-y"
                        :class="step._config._argsError ? 'border-red-500' : ''"></textarea>
                      <p v-if="step._config._argsError" class="text-xs text-red-400 mt-1">{{ step._config._argsError }}</p>
                    </div>
                  </template>

                  <!-- Condition Config -->
                  <template v-if="step.type === 'condition'">
                    <div>
                      <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1">
                        Reference
                        <span class="font-normal opacity-60">(e.g. {{ templateExample }})</span>
                      </label>
                      <input v-model="step._config.reference" type="text" placeholder="{{step.1.output}}"
                        class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent font-mono" />
                    </div>
                    <div>
                      <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1">Operator</label>
                      <select v-model="step._config.operator"
                        class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent">
                        <option value="contains">Contains</option>
                        <option value="equals">Equals</option>
                        <option value="not_empty">Not Empty</option>
                        <option value="regex">Regex Match</option>
                      </select>
                    </div>
                    <div v-if="step._config.operator !== 'not_empty'">
                      <label class="block text-xs font-semibold text-kakoclaw-text-secondary mb-1">Value</label>
                      <input v-model="step._config.value" type="text" placeholder="Compare value"
                        class="w-full px-3 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg text-sm focus:outline-none focus:border-kakoclaw-accent" />
                    </div>
                  </template>
                </div>
              </div>
            </template>
          </VueDraggable>
        </div>

        <!-- Add Step Buttons -->
        <div class="flex items-center gap-2 mb-6">
          <button @click="addStep('prompt')"
            class="flex items-center gap-2 px-3 py-2 text-xs bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded-lg hover:bg-blue-500/20 transition-colors">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
            Prompt
          </button>
          <button @click="addStep('tool')"
            class="flex items-center gap-2 px-3 py-2 text-xs bg-amber-500/10 text-amber-400 border border-amber-500/20 rounded-lg hover:bg-amber-500/20 transition-colors">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
            Tool
          </button>
          <button @click="addStep('condition')"
            class="flex items-center gap-2 px-3 py-2 text-xs bg-purple-500/10 text-purple-400 border border-purple-500/20 rounded-lg hover:bg-purple-500/20 transition-colors">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
            Condition
          </button>
        </div>

        <!-- Test Run -->
        <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5">
          <div class="flex items-center justify-between mb-3">
            <h3 class="text-sm font-semibold">Test Run</h3>
            <button v-if="editingWorkflow.id" @click="testRun" :disabled="testRunning"
              class="flex items-center gap-2 px-3 py-1.5 text-xs bg-emerald-500/10 text-emerald-400 rounded-lg hover:bg-emerald-500/20 transition-colors disabled:opacity-50">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
              {{ testRunning ? 'Running...' : 'Run Now' }}
            </button>
            <span v-else class="text-xs text-kakoclaw-text-secondary">Save the workflow first to test it</span>
          </div>

          <div v-if="testResults" class="space-y-2">
            <div v-for="(res, i) in testResults" :key="i"
              class="text-xs p-3 rounded-lg bg-kakoclaw-bg border border-kakoclaw-border">
              <div class="flex items-center gap-2 mb-1">
                <span class="font-mono font-semibold">{{ res.label || res.step_id }}</span>
                <span class="px-1.5 py-0.5 rounded text-[10px]"
                  :class="res.error ? 'bg-red-500/10 text-red-400' : res.skipped ? 'bg-gray-500/10 text-gray-400' : 'bg-emerald-500/10 text-emerald-400'">
                  {{ res.error ? 'Error' : res.skipped ? 'Skipped' : 'OK' }}
                </span>
                <span v-if="res.duration_ms" class="text-kakoclaw-text-secondary">{{ res.duration_ms }}ms</span>
              </div>
              <pre v-if="res.output" class="whitespace-pre-wrap text-kakoclaw-text-secondary max-h-32 overflow-auto">{{ res.output }}</pre>
              <pre v-if="res.error" class="whitespace-pre-wrap text-red-400 max-h-32 overflow-auto">{{ res.error }}</pre>
            </div>
          </div>
          <div v-else class="text-xs text-kakoclaw-text-secondary">
            No test results yet. Save and run the workflow to see output here.
          </div>
        </div>
      </template>
    </div>

    <!-- Error Toast -->
    <div v-if="error" class="fixed bottom-4 right-4 max-w-sm bg-red-500/90 text-white px-4 py-3 rounded-lg shadow-lg z-50 text-sm">
      {{ error }}
      <button @click="error = ''" class="ml-2 opacity-75 hover:opacity-100">&times;</button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import workflowService from '../services/workflowService'

const loading = ref(true)
const saving = ref(false)
const editing = ref(false)
const error = ref('')
const workflows = ref([])
const availableTools = ref([])
const expandedStep = ref(null)
const runningId = ref(null)
const testRunning = ref(false)
const testResults = ref(null)

// Template syntax hint strings (cannot use literal {{ }} in Vue templates)
const templateHint = '{{step.N.output}}'
const templateExample = '{{step.1.output}}'

function extractError(e) {
  if (typeof e.response?.data === 'string') return e.response.data
  return e.response?.data?.error || e.message
}

const editingWorkflow = reactive({
  id: null,
  name: '',
  description: '',
  enabled: true,
  steps: []
})

onMounted(async () => {
  await Promise.all([loadWorkflows(), loadTools()])
})

async function loadWorkflows() {
  loading.value = true
  try {
    const data = await workflowService.fetchWorkflows()
    workflows.value = (data.workflows || []).map(w => ({ ...w, _lastResults: null, _runs: null }))
  } catch (e) {
    error.value = 'Failed to load workflows: ' + extractError(e)
  } finally {
    loading.value = false
  }
}

async function loadTools() {
  try {
    const data = await workflowService.fetchTools()
    availableTools.value = data.tools || []
  } catch {
    // Tools list is optional — user can type tool names manually
  }
}

function startCreate() {
  Object.assign(editingWorkflow, { id: null, name: '', description: '', enabled: true, steps: [] })
  testResults.value = null
  expandedStep.value = null
  editing.value = true
}

function startEdit(wf) {
  let steps = []
  try {
    const raw = typeof wf.steps === 'string' ? JSON.parse(wf.steps) : wf.steps
    steps = (raw || []).map(s => deserializeStep(s))
  } catch { steps = [] }

  Object.assign(editingWorkflow, {
    id: wf.id,
    name: wf.name,
    description: wf.description || '',
    enabled: wf.enabled,
    steps
  })
  testResults.value = null
  expandedStep.value = steps.length > 0 ? steps[0].id : null
  editing.value = true
}

function cancelEdit() {
  editing.value = false
  testResults.value = null
  expandedStep.value = null
}

function toggleStep(stepId) {
  expandedStep.value = expandedStep.value === stepId ? null : stepId
}

// Step helpers
let stepCounter = 0

function makeStepId() {
  return 'step_' + Date.now() + '_' + (++stepCounter)
}

function addStep(type) {
  const step = {
    id: makeStepId(),
    type,
    label: '',
    on_error: 'stop',
    _config: defaultConfig(type)
  }
  editingWorkflow.steps.push(step)
  expandedStep.value = step.id
}

function removeStep(idx, stepId) {
  editingWorkflow.steps.splice(idx, 1)
  if (expandedStep.value === stepId) expandedStep.value = null
}

function defaultConfig(type) {
  if (type === 'prompt') return { message: '', model: '' }
  if (type === 'tool') return { tool_name: '', _argsJson: '{}', _argsError: '' }
  if (type === 'condition') return { operator: 'contains', value: '', reference: '' }
  return {}
}

function deserializeStep(raw) {
  const step = { id: raw.id || makeStepId(), type: raw.type, label: raw.label || '', on_error: raw.on_error || 'stop' }
  let cfg = {}
  try { cfg = typeof raw.config === 'string' ? JSON.parse(raw.config) : (raw.config || {}) } catch { cfg = {} }

  if (raw.type === 'prompt') {
    step._config = { message: cfg.message || '', model: cfg.model || '' }
  } else if (raw.type === 'tool') {
    let argsJson = '{}'
    try { argsJson = JSON.stringify(cfg.args || {}, null, 2) } catch { argsJson = '{}' }
    step._config = { tool_name: cfg.tool_name || '', _argsJson: argsJson, _argsError: '' }
  } else if (raw.type === 'condition') {
    step._config = { operator: cfg.operator || 'contains', value: cfg.value || '', reference: cfg.reference || '' }
  } else {
    step._config = cfg
  }
  return step
}

function serializeSteps(steps) {
  return steps.map(s => {
    const out = { id: s.id, type: s.type, label: s.label, on_error: s.on_error || 'stop' }
    if (s.type === 'prompt') {
      out.config = { message: s._config.message || '' }
      if (s._config.model) out.config.model = s._config.model
    } else if (s.type === 'tool') {
      let args = {}
      try { args = JSON.parse(s._config._argsJson || '{}') } catch { /* keep empty */ }
      out.config = { tool_name: s._config.tool_name || '', args }
    } else if (s.type === 'condition') {
      out.config = { operator: s._config.operator, value: s._config.value, reference: s._config.reference }
    }
    return out
  })
}

async function saveWorkflow() {
  if (!editingWorkflow.name.trim()) {
    error.value = 'Workflow name is required'
    return
  }
  if (editingWorkflow.steps.length === 0) {
    error.value = 'Workflow must have at least one step'
    return
  }
  // Validate tool step JSON
  for (const step of editingWorkflow.steps) {
    if (step.type === 'tool') {
      try {
        JSON.parse(step._config._argsJson || '{}')
        step._config._argsError = ''
      } catch (e) {
        step._config._argsError = 'Invalid JSON: ' + e.message
        error.value = 'Fix JSON errors in tool step arguments before saving'
        return
      }
    }
  }

  saving.value = true
  try {
    const payload = {
      name: editingWorkflow.name,
      description: editingWorkflow.description,
      enabled: editingWorkflow.enabled,
      steps: serializeSteps(editingWorkflow.steps)
    }

    if (editingWorkflow.id) {
      await workflowService.updateWorkflow(editingWorkflow.id, payload)
    } else {
      const created = await workflowService.createWorkflow(payload)
      editingWorkflow.id = created.id
    }
    await loadWorkflows()
    error.value = ''
  } catch (e) {
    error.value = 'Failed to save: ' + extractError(e)
  } finally {
    saving.value = false
  }
}

async function deleteWorkflow(wf) {
  if (!confirm(`Delete workflow "${wf.name}"?`)) return
  try {
    await workflowService.deleteWorkflow(wf.id)
    await loadWorkflows()
  } catch (e) {
    error.value = 'Failed to delete: ' + extractError(e)
  }
}

async function runWorkflow(wf) {
  runningId.value = wf.id
  try {
    const data = await workflowService.runWorkflow(wf.id)
    wf._lastResults = data.results || []
  } catch (e) {
    error.value = 'Run failed: ' + extractError(e)
  } finally {
    runningId.value = null
  }
}

async function testRun() {
  if (!editingWorkflow.id) return
  // Auto-save before running
  await saveWorkflow()
  if (saving.value || error.value) return
  testRunning.value = true
  try {
    const data = await workflowService.runWorkflow(editingWorkflow.id)
    testResults.value = data.results || []
  } catch (e) {
    error.value = 'Test run failed: ' + extractError(e)
  } finally {
    testRunning.value = false
  }
}

async function showRuns(wf) {
  try {
    const data = await workflowService.getWorkflowRuns(wf.id)
    wf._runs = data.runs || []
  } catch (e) {
    error.value = 'Failed to load runs: ' + extractError(e)
  }
}

// Display helpers
function countSteps(wf) {
  try {
    const raw = typeof wf.steps === 'string' ? JSON.parse(wf.steps) : wf.steps
    return Array.isArray(raw) ? raw.length : 0
  } catch { return 0 }
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function runDuration(run) {
  if (!run.started_at || !run.finished_at) return ''
  const ms = new Date(run.finished_at) - new Date(run.started_at)
  if (ms < 1000) return ms + 'ms'
  return (ms / 1000).toFixed(1) + 's'
}

function stepTypeClass(type) {
  if (type === 'prompt') return 'bg-blue-500/10 text-blue-400'
  if (type === 'tool') return 'bg-amber-500/10 text-amber-400'
  if (type === 'condition') return 'bg-purple-500/10 text-purple-400'
  return 'bg-gray-500/10 text-gray-400'
}
</script>
