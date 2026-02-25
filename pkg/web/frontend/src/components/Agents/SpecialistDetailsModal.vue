<template>
  <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
    <div class="bg-makoclaw-surface rounded-2xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden border border-makoclaw-border animate-in fade-in zoom-in duration-200 flex flex-col">
      <!-- Header -->
      <div class="flex justify-between items-start p-6 border-b border-makoclaw-border bg-makoclaw-bg/20">
        <div class="flex items-center gap-4">
          <div :class="`w-14 h-14 rounded-xl ${agentsStore.getSpecialistBgColor(specialist.name)} flex items-center justify-center`">
            <svg class="w-7 h-7" :class="agentsStore.getSpecialistColor(specialist.name)" fill="none" stroke="currentColor" viewBox="0 0 24 24" v-html="agentsStore.getSpecialistIcon(specialist.name)"></svg>
          </div>
          <div>
            <h3 class="text-xl font-bold text-makoclaw-text capitalize">{{ specialist.name }}</h3>
            <p class="text-sm text-makoclaw-text-secondary">{{ specialist.description }}</p>
          </div>
        </div>
        <button @click="close" class="text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">
          <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="flex-1 overflow-y-auto p-6 space-y-6 custom-scrollbar">
        <!-- Configuration -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="glass-panel rounded-xl p-4 border border-makoclaw-border/50">
            <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Provider</p>
            <p class="font-medium text-makoclaw-text">{{ specialist.provider }}</p>
          </div>
          <div class="glass-panel rounded-xl p-4 border border-makoclaw-border/50">
            <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Model</p>
            <p class="font-medium text-makoclaw-text text-sm">{{ specialist.model }}</p>
          </div>
          <div class="glass-panel rounded-xl p-4 border border-makoclaw-border/50">
            <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Max Tokens</p>
            <p class="font-medium text-makoclaw-text">{{ specialist.max_tokens }}</p>
          </div>
          <div class="glass-panel rounded-xl p-4 border border-makoclaw-border/50">
            <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Temperature</p>
            <p class="font-medium text-makoclaw-text">{{ specialist.temperature }}</p>
          </div>
        </div>

        <!-- System Prompt -->
        <div class="glass-panel rounded-xl p-5 border border-makoclaw-border/50">
          <h4 class="font-bold text-makoclaw-text mb-3 flex items-center gap-2">
            <svg class="w-4 h-4 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            System Prompt
          </h4>
          <p class="text-sm text-makoclaw-text-secondary whitespace-pre-wrap bg-makoclaw-bg/40 rounded-lg p-4">{{ specialist.prompt }}</p>
        </div>

        <!-- Tools -->
        <div class="glass-panel rounded-xl p-5 border border-makoclaw-border/50">
          <h4 class="font-bold text-makoclaw-text mb-3 flex items-center gap-2">
            <svg class="w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            </svg>
            Available Tools
          </h4>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="tool in (specialist.tools || [])"
              :key="tool"
              class="px-3 py-1.5 text-xs font-medium bg-blue-500/10 text-blue-400 rounded-full border border-blue-500/20"
            >
              {{ tool }}
            </span>
          </div>
        </div>

        <!-- Keywords -->
        <div class="glass-panel rounded-xl p-5 border border-makoclaw-border/50">
          <h4 class="font-bold text-makoclaw-text mb-3 flex items-center gap-2">
            <svg class="w-4 h-4 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
            </svg>
            Keywords
          </h4>
          <p class="text-sm text-makoclaw-text-secondary">These keywords help the Orchestrator match tasks to this specialist</p>
          <div class="flex flex-wrap gap-2 mt-3">
            <span
              v-for="keyword in (specialist.keywords || [])"
              :key="keyword"
              class="px-3 py-1.5 text-xs font-medium bg-purple-500/10 text-purple-400 rounded-full border border-purple-500/20"
            >
              {{ keyword }}
            </span>
          </div>
        </div>

        <!-- Performance Metrics (if available) -->
        <div v-if="specialistMetrics" class="glass-panel rounded-xl p-5 border border-makoclaw-border/50">
          <h4 class="font-bold text-makoclaw-text mb-4 flex items-center gap-2">
            <svg class="w-4 h-4 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
            Performance Metrics
          </h4>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div class="p-3 bg-makoclaw-bg/40 rounded-lg">
              <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-1">Total Calls</p>
              <p class="text-lg font-bold text-makoclaw-text">{{ specialistMetrics.calls || 0 }}</p>
            </div>
            <div class="p-3 bg-makoclaw-bg/40 rounded-lg">
              <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-1">Total Tokens</p>
              <p class="text-lg font-bold text-makoclaw-text">{{ formatNumber(specialistMetrics.tokens || 0) }}</p>
            </div>
            <div class="p-3 bg-makoclaw-bg/40 rounded-lg">
              <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-1">Avg Cost/Call</p>
              <p class="text-lg font-bold text-makoclaw-text">{{ formatCost(specialistMetrics.avg_cost_per_call || 0) }}</p>
            </div>
            <div class="p-3 bg-makoclaw-bg/40 rounded-lg">
              <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-1">Total Cost</p>
              <p class="text-lg font-bold text-makoclaw-text">{{ formatCost(specialistMetrics.total_cost || 0) }}</p>
            </div>
          </div>
        </div>

        <!-- Test Specialist Section -->
        <div class="glass-panel rounded-xl p-5 border border-makoclaw-border/50">
          <h4 class="font-bold text-makoclaw-text mb-3 flex items-center gap-2">
            <svg class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Test Specialist
          </h4>
          <p class="text-sm text-makoclaw-text-secondary mb-4">Send a test message to this specialist and see the response</p>

          <!-- Test Input -->
          <div class="space-y-3">
            <textarea
              v-model="testPrompt"
              placeholder="Enter a test message for the specialist..."
              rows="3"
              class="w-full px-4 py-3 bg-makoclaw-bg/60 border border-makoclaw-border/50 rounded-lg text-makoclaw-text placeholder-makoclaw-text-secondary/50 focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/50 focus:border-makoclaw-accent resize-none"
              :disabled="testing"
            ></textarea>

            <div class="flex justify-end">
              <button
                @click="runTest"
                :disabled="testing || !testPrompt.trim()"
                class="px-5 py-2.5 text-sm font-bold bg-green-600 hover:bg-green-700 text-white rounded-xl shadow-lg shadow-green-500/20 transition-all flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <svg v-if="testing" class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                {{ testing ? 'Running Test...' : 'Run Test' }}
              </button>
            </div>
          </div>

          <!-- Test Result -->
          <div v-if="testResult" class="mt-4 space-y-3">
            <!-- Response -->
            <div class="bg-makoclaw-bg/60 rounded-lg p-4 border border-green-500/30">
              <div class="flex items-center gap-2 mb-2">
                <svg class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span class="text-xs font-bold uppercase tracking-wider text-green-400">Response</span>
              </div>
              <p class="text-sm text-makoclaw-text whitespace-pre-wrap">{{ testResult.response }}</p>
            </div>

            <!-- Metadata -->
            <div class="grid grid-cols-2 gap-3">
              <!-- Tools Available -->
              <div class="bg-makoclaw-bg/40 rounded-lg p-3">
                <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Tools Available</p>
                <div class="flex flex-wrap gap-1">
                  <span
                    v-for="tool in (testResult.tools_used || [])"
                    :key="tool"
                    class="px-2 py-0.5 text-[10px] font-medium bg-blue-500/10 text-blue-400 rounded-full"
                  >
                    {{ tool }}
                  </span>
                  <span v-if="!testResult.tools_used || testResult.tools_used.length === 0" class="text-xs text-makoclaw-text-secondary">
                    No tools
                  </span>
                </div>
              </div>

              <!-- Duration -->
              <div class="bg-makoclaw-bg/40 rounded-lg p-3">
                <p class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary mb-2">Duration</p>
                <p class="text-sm font-bold text-makoclaw-text">{{ formatDuration(testResult.duration_ms) }}</p>
              </div>
            </div>
          </div>

          <!-- Test Error -->
          <div v-if="testError" class="mt-4 bg-red-500/10 rounded-lg p-4 border border-red-500/30">
            <div class="flex items-center gap-2 mb-2">
              <svg class="w-4 h-4 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span class="text-xs font-bold uppercase tracking-wider text-red-400">Error</span>
            </div>
            <p class="text-sm text-red-300">{{ testError }}</p>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex justify-end gap-3 p-6 border-t border-makoclaw-border bg-makoclaw-bg/20">
        <router-link
          to="/settings?tab=agents"
          class="px-5 py-2.5 text-sm font-medium text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
          @click="close"
        >
          Edit in Settings
        </router-link>
        <button
          @click="close"
          class="px-5 py-2.5 text-sm font-bold bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl shadow-lg shadow-makoclaw-accent/20 transition-all"
        >
          Close
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useAgentsStore } from '../../stores/agentsStore'
import { useToast } from '../../composables/useToast'

const toast = useToast()
const agentsStore = useAgentsStore()

const props = defineProps({
  show: {
    type: Boolean,
    required: true
  },
  specialist: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['close'])

const testing = ref(false)
const specialistMetrics = ref(null)
const testPrompt = ref('')
const testResult = ref(null)
const testError = ref(null)

onMounted(async () => {
  await fetchSpecialistMetrics()
})

// Reset test state when modal is opened/closed or specialist changes
watch(() => props.show, (newVal) => {
  if (newVal) {
    testPrompt.value = ''
    testResult.value = null
    testError.value = null
  }
})

watch(() => props.specialist, () => {
  testPrompt.value = ''
  testResult.value = null
  testError.value = null
})

async function fetchSpecialistMetrics() {
  const metrics = await agentsStore.fetchSpecialistDetails(props.specialist.name)
  if (metrics) {
    specialistMetrics.value = metrics.metrics
  }
}

async function runTest() {
  if (!testPrompt.value.trim()) {
    toast.warning('Please enter a test message')
    return
  }

  testing.value = true
  testResult.value = null
  testError.value = null

  try {
    const result = await agentsStore.testSpecialist(props.specialist.name, testPrompt.value.trim())
    if (result && result.success) {
      testResult.value = result
      toast.success('Test completed successfully!')
    } else if (result && result.error) {
      testError.value = result.error
      toast.error('Test failed: ' + result.error)
    } else {
      testError.value = 'Unknown error occurred'
      toast.error('Test failed')
    }
  } catch (error) {
    testError.value = error.message || 'Failed to run test'
    toast.error('Test failed: ' + (error.message || 'Unknown error'))
  } finally {
    testing.value = false
  }
}

function formatCost(cost) {
  if (!cost) return '$0.00'
  return '$' + cost.toFixed(4)
}

function formatNumber(num) {
  if (!num) return '0'
  return num.toLocaleString()
}

function formatDuration(ms) {
  if (!ms) return '0ms'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(2)}s`
  return `${(ms / 60000).toFixed(2)}m`
}

function close() {
  emit('close')
}
</script>


