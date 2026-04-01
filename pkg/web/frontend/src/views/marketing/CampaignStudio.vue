<template>
  <div class="flex flex-col h-full">
    <!-- ─── Header ──────────────────────────────────────── -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 pb-3">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-pink-500/10">
            <svg class="w-5 h-5 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <h1 class="text-lg font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-pink-400 bg-clip-text text-transparent">
              Campaign Studio
            </h1>
            <p class="text-xs text-makoclaw-text-secondary">
              {{ account }} / {{ campaign }}
            </p>
          </div>
          <button
            v-if="!wf"
            class="px-4 py-2.5 bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl transition-all shadow-lg shadow-pink-500/25 text-sm font-bold flex items-center gap-2 active:scale-95"
            :disabled="loading"
            @click="handleStart"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Start Studio
          </button>
          <!-- Progress bar when active -->
          <div v-if="wf && wf.progress" class="hidden sm:flex items-center gap-2 text-xs text-makoclaw-text-secondary">
            <div class="w-24 h-1.5 rounded-full bg-makoclaw-surface overflow-hidden">
              <div
                class="h-full rounded-full bg-gradient-to-r from-pink-500 to-violet-500 transition-all duration-500"
                :style="{ width: wf.progress.pct + '%' }"
              />
            </div>
            {{ wf.progress.completed }}/{{ wf.progress.total }}
          </div>
        </div>
      </div>
    </div>

    <!-- ─── Not Started ──────────────────────────────────── -->
    <div v-if="!wf" class="flex-1 overflow-y-auto custom-scrollbar p-6">
      <div class="flex flex-col items-center justify-center py-16 text-center">
        <div class="relative">
          <div class="absolute inset-0 bg-gradient-to-br from-pink-500/30 to-violet-500/30 rounded-3xl blur-2xl opacity-50" />
          <div class="relative w-20 h-20 rounded-2xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-2xl">
            <svg class="w-10 h-10 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
        </div>
        <h3 class="text-xl font-bold text-makoclaw-text mt-6">Campaign Studio</h3>
        <p class="text-sm text-makoclaw-text-secondary/70 mt-2 max-w-md">
          Create a complete marketing campaign step by step with AI assistance. Generate briefs, strategy, copy, visuals, and scheduling — with your approval at every stage.
        </p>
        <div class="flex flex-wrap justify-center gap-2 mt-6">
          <span v-for="stage in stageOrder" :key="stage" class="px-2.5 py-1 rounded-lg bg-makoclaw-surface/50 text-xs text-makoclaw-text-secondary border border-makoclaw-border/30">
            {{ stageIcons[stage] }} {{ stageLabels[stage] }}
          </span>
        </div>
      </div>
    </div>

    <!-- ─── Stepper Pipeline ─────────────────────────────── -->
    <div v-else class="flex-1 flex flex-col overflow-hidden">
      <!-- Stage tabs -->
      <div class="px-4 sm:px-6 pt-4 pb-2 border-b border-makoclaw-border/20">
        <div class="flex gap-1 overflow-x-auto custom-scrollbar pb-1">
          <button
            v-for="stage in stageOrder"
            :key="stage"
            class="flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-semibold whitespace-nowrap transition-all"
            :class="stageTabClass(stage)"
            @click="selectedStageView = stage"
          >
            <span>{{ stageIcons[stage] }}</span>
            <span>{{ stageLabels[stage] }}</span>
            <span v-if="stageState(stage) === 'approved'" class="text-emerald-400">&#10003;</span>
            <span v-else-if="stageState(stage) === 'generating'" class="animate-spin">&#9696;</span>
            <span v-else-if="stageState(stage) === 'awaiting_approval'" class="text-amber-400">&#9679;</span>
          </button>
        </div>
      </div>

      <!-- Stage content -->
      <div class="flex-1 overflow-y-auto custom-scrollbar p-4 sm:p-6">
        <!-- Stage Panel -->
        <div class="max-w-4xl mx-auto space-y-4">
          <!-- State: Pending -->
          <div v-if="currentStageState === 'pending' && isCurrentStage" class="glass-panel rounded-2xl p-6 space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center">
                <span class="text-lg">{{ stageIcons[currentStage] }}</span>
              </div>
              <div>
                <h3 class="text-sm font-bold text-makoclaw-text">{{ stageLabels[currentStage] }}</h3>
                <p class="text-xs text-makoclaw-text-secondary">Ready to generate with AI</p>
              </div>
            </div>

            <!-- Config options for this stage -->
            <div v-if="currentStage === 'brief' || currentStage === 'copy'" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div v-if="currentStage === 'brief'">
                <label class="text-xs font-semibold text-makoclaw-text-secondary block mb-1">Objective</label>
                <input
                  v-model="generateParams.objective"
                  class="w-full px-3 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 transition-all min-h-[40px]"
                  placeholder="e.g., product launch, brand awareness..."
                />
              </div>
              <div v-if="currentStage === 'brief'">
                <label class="text-xs font-semibold text-makoclaw-text-secondary block mb-1">Target Audience</label>
                <input
                  v-model="generateParams.target_audience"
                  class="w-full px-3 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 transition-all min-h-[40px]"
                  placeholder="e.g., CTOs at tech startups..."
                />
              </div>
              <div>
                <label class="text-xs font-semibold text-makoclaw-text-secondary block mb-1">Platforms</label>
                <input
                  v-model="generateParams.platforms"
                  class="w-full px-3 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 transition-all min-h-[40px]"
                  placeholder="twitter,linkedin"
                />
              </div>
              <div>
                <label class="text-xs font-semibold text-makoclaw-text-secondary block mb-1">Tone</label>
                <input
                  v-model="generateParams.tone"
                  class="w-full px-3 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 transition-all min-h-[40px]"
                  placeholder="professional, casual, technical..."
                />
              </div>
              <div v-if="currentStage === 'copy'">
                <label class="text-xs font-semibold text-makoclaw-text-secondary block mb-1">Posts per platform</label>
                <input
                  v-model.number="generateParams.post_count"
                  type="number"
                  min="1"
                  max="20"
                  class="w-full px-3 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 transition-all min-h-[40px]"
                  placeholder="5"
                />
              </div>
            </div>

            <div v-if="currentStage === 'visuals'" class="space-y-3">
              <div>
                <label class="text-xs font-semibold text-makoclaw-text-secondary block mb-1">Visual Style</label>
                <input
                  v-model="generateParams.image_style"
                  class="w-full px-3 py-2 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:ring-2 focus:ring-pink-500/30 focus:border-pink-500/50 transition-all min-h-[40px]"
                  placeholder="e.g., clean minimal tech, vibrant startup culture..."
                />
              </div>
            </div>

            <div class="flex gap-2 pt-2">
              <button
                class="px-5 py-2.5 bg-gradient-to-r from-pink-500 to-violet-500 hover:from-pink-600 hover:to-violet-600 text-white rounded-xl transition-all shadow-lg shadow-pink-500/25 text-sm font-bold flex items-center gap-2 active:scale-95"
                :disabled="generating"
                @click="handleGenerate"
              >
                <svg v-if="generating" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                {{ generating ? 'Generating...' : 'Generate with AI' }}
              </button>
              <button
                class="px-4 py-2.5 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text rounded-xl hover:bg-makoclaw-surface/50 transition-all min-h-[40px]"
                @click="handleSkip"
              >
                Skip
              </button>
            </div>
          </div>

          <!-- State: Pending (future stage, not yet reachable) -->
          <div v-if="currentStageState === 'pending' && !isCurrentStage" class="glass-panel rounded-2xl p-6 flex flex-col items-center justify-center text-center space-y-3">
            <div class="w-12 h-12 rounded-2xl bg-makoclaw-surface/50 flex items-center justify-center">
              <span class="text-2xl">{{ stageIcons[viewingStage] }}</span>
            </div>
            <h3 class="text-sm font-bold text-makoclaw-text">{{ stageLabels[viewingStage] }}</h3>
            <p class="text-xs text-makoclaw-text-secondary">
              This stage is waiting. Complete <strong>{{ stageLabels[currentStage] }}</strong> first.
            </p>
          </div>

          <!-- State: Generating -->
          <div v-if="currentStageState === 'generating'" class="glass-panel rounded-2xl p-8 flex flex-col items-center justify-center text-center space-y-4">
            <div class="relative">
              <div class="absolute inset-0 bg-gradient-to-br from-pink-500/30 to-violet-500/30 rounded-full blur-xl opacity-50 animate-subtlePulse" />
              <div class="relative w-16 h-16 rounded-2xl bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center ring-1 ring-white/10">
                <svg class="w-8 h-8 text-pink-400 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
              </div>
            </div>
            <div>
              <h3 class="text-sm font-bold text-makoclaw-text">AI is working on {{ stageLabels[currentStage] }}...</h3>
              <p class="text-xs text-makoclaw-text-secondary mt-1">This may take up to 3 minutes. The AI is reading context and generating content.</p>
            </div>
          </div>

          <!-- State: Awaiting Approval -->
          <div v-if="currentStageState === 'awaiting_approval'" class="space-y-4">
            <!-- AI Output Preview -->
            <div class="glass-panel rounded-2xl overflow-hidden">
              <div class="px-4 py-3 border-b border-makoclaw-border/20 bg-gradient-to-r from-makoclaw-surface/50 to-transparent">
                <div class="flex items-center gap-2">
                  <div class="w-7 h-7 rounded-lg bg-gradient-to-br from-pink-500/20 to-violet-500/20 flex items-center justify-center">
                    <svg class="w-3.5 h-3.5 text-pink-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                  </div>
                  <h3 class="text-sm font-bold text-makoclaw-text">AI Generated {{ stageLabels[currentStage] }}</h3>
                  <span class="ml-auto text-[10px] text-makoclaw-text-secondary">Attempt #{{ currentAttempts }}</span>
                </div>
              </div>
              <div class="p-4 max-h-[50vh] overflow-y-auto custom-scrollbar">
                <pre class="text-sm text-makoclaw-text-secondary whitespace-pre-wrap font-mono bg-makoclaw-bg/30 rounded-xl p-4 border border-makoclaw-border/20">{{ aiOutput }}</pre>
              </div>
            </div>

            <!-- Generated Files -->
            <div v-if="currentOutputFiles.length > 0" class="glass-panel rounded-2xl p-4">
              <h4 class="text-xs font-bold text-makoclaw-text mb-2">Generated Files</h4>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="file in currentOutputFiles"
                  :key="file"
                  class="px-2.5 py-1 rounded-lg bg-emerald-500/10 text-emerald-400 text-xs border border-emerald-500/20"
                >
                  {{ file }}
                </span>
              </div>
            </div>

            <!-- Action Buttons -->
            <div class="glass-panel rounded-2xl p-4 space-y-3">
              <div class="flex flex-wrap gap-2">
                <button
                  class="px-5 py-2.5 bg-gradient-to-r from-emerald-500 to-emerald-600 hover:from-emerald-600 hover:to-emerald-700 text-white rounded-xl transition-all shadow-lg shadow-emerald-500/25 text-sm font-bold flex items-center gap-2 active:scale-95"
                  @click="handleApprove"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                  Approve &amp; Continue
                </button>
                <button
                  class="px-4 py-2.5 bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 text-white rounded-xl transition-all shadow-lg shadow-amber-500/25 text-sm font-bold flex items-center gap-2 active:scale-95"
                  @click="showRejectModal = true"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Redo with Feedback
                </button>
                <button
                  class="px-4 py-2.5 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text rounded-xl hover:bg-makoclaw-surface/50 transition-all min-h-[40px]"
                  @click="handleSkip"
                >
                  Skip
                </button>
              </div>
            </div>
          </div>

          <!-- State: Approved -->
          <div v-if="currentStageState === 'approved'" class="glass-panel rounded-2xl p-6 flex items-center gap-3">
            <div class="w-8 h-8 rounded-lg bg-emerald-500/20 flex items-center justify-center">
              <svg class="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <div class="flex-1">
              <h3 class="text-sm font-bold text-emerald-400">{{ stageLabels[viewingStage] }} — Approved</h3>
              <p class="text-xs text-makoclaw-text-secondary">This stage is complete</p>
            </div>
            <button
              class="px-3 py-1.5 text-xs text-makoclaw-text-secondary hover:text-amber-400 rounded-lg hover:bg-makoclaw-surface/50 transition-all"
              @click="handleReset(viewingStage)"
            >
              Redo
            </button>
          </div>

          <!-- State: Skipped -->
          <div v-if="currentStageState === 'skipped'" class="glass-panel rounded-2xl p-6 flex items-center gap-3">
            <div class="w-8 h-8 rounded-lg bg-makoclaw-surface/50 flex items-center justify-center">
              <svg class="w-4 h-4 text-makoclaw-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
              </svg>
            </div>
            <div class="flex-1">
              <h3 class="text-sm font-bold text-makoclaw-text-secondary">{{ stageLabels[viewingStage] }} — Skipped</h3>
            </div>
            <button
              class="px-3 py-1.5 text-xs text-makoclaw-text-secondary hover:text-pink-400 rounded-lg hover:bg-makoclaw-surface/50 transition-all"
              @click="handleReset(viewingStage)"
            >
              Redo
            </button>
          </div>

          <!-- Completed banner -->
          <div v-if="wf && wf.completed_at" class="glass-panel rounded-2xl p-6 text-center space-y-3 border border-emerald-500/20">
            <div class="w-12 h-12 mx-auto rounded-2xl bg-gradient-to-br from-emerald-500/20 to-green-500/20 flex items-center justify-center">
              <svg class="w-6 h-6 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3 class="text-lg font-bold text-emerald-400">Campaign Complete!</h3>
            <p class="text-sm text-makoclaw-text-secondary">All stages have been completed. Your campaign assets are ready.</p>
          </div>
        </div>
      </div>
    </div>

    <!-- ─── Reject Feedback Modal ───────────────────────── -->
    <Teleport to="body">
      <div v-if="showRejectModal" class="fixed inset-0 z-[60] flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showRejectModal = false" />
        <div class="relative bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl ring-1 ring-white/10 w-full max-w-lg animate-scaleIn">
          <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent">
            <div class="flex items-center gap-2">
              <div class="w-7 h-7 rounded-lg bg-gradient-to-br from-amber-500/20 to-orange-500/20 flex items-center justify-center">
                <svg class="w-3.5 h-3.5 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
              </div>
              <h3 class="text-sm font-bold text-makoclaw-text">Redo with Feedback</h3>
            </div>
          </div>
          <div class="p-4 space-y-3">
            <p class="text-xs text-makoclaw-text-secondary">Tell the AI what to improve:</p>
            <textarea
              v-model="rejectFeedback"
              rows="4"
              class="w-full px-3 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:ring-2 focus:ring-amber-500/30 focus:border-amber-500/50 transition-all resize-none"
              placeholder="e.g., Make the tone more casual, focus on developer audience, add more hashtags..."
            />
          </div>
          <div class="p-4 border-t border-makoclaw-border/20 flex justify-end gap-2">
            <button
              class="px-4 py-2 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text rounded-xl hover:bg-makoclaw-surface/50 transition-all"
              @click="showRejectModal = false"
            >
              Cancel
            </button>
            <button
              class="px-4 py-2.5 bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 text-white rounded-xl text-sm font-bold transition-all active:scale-95"
              :disabled="!rejectFeedback.trim()"
              @click="handleReject"
            >
              Redo
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useMarketingStore } from '../../stores/marketingStore'

const props = defineProps({
  account: { type: String, required: true },
  campaign: { type: String, required: true }
})

const store = useMarketingStore()

// Stage metadata
const stageOrder = ['brief', 'strategy', 'copy', 'visuals', 'schedule', 'launch']
const stageLabels = {
  brief: 'Campaign Brief',
  strategy: 'Content Strategy',
  copy: 'Marketing Copy',
  visuals: 'Visual Assets',
  schedule: 'Posting Schedule',
  launch: 'Launch Report'
}
const stageIcons = {
  brief: '📋',
  strategy: '🎯',
  copy: '✍️',
  visuals: '🎨',
  schedule: '📅',
  launch: '🚀'
}

// Local state
const selectedStageView = ref(null)
const showRejectModal = ref(false)
const rejectFeedback = ref('')
const generateParams = ref({
  objective: '',
  target_audience: '',
  platforms: 'twitter,linkedin',
  tone: 'professional',
  post_count: 5,
  image_style: ''
})

// Store getters
const wf = computed(() => store.workflow)
const loading = computed(() => store.workflowLoading)
const generating = computed(() => store.workflowGenerating)

const currentStage = computed(() => wf.value?.current_stage || '')
const viewingStage = computed(() => selectedStageView.value || currentStage.value)

const isCurrentStage = computed(() => viewingStage.value === currentStage.value)

const currentStageState = computed(() => {
  if (!wf.value?.stages) return 'pending'
  return wf.value.stages[viewingStage.value]?.state || 'pending'
})

const currentAttempts = computed(() => {
  if (!wf.value?.stages) return 0
  return wf.value.stages[viewingStage.value]?.attempts || 0
})

const aiOutput = computed(() => {
  if (!wf.value?.stages) return ''
  return wf.value.stages[viewingStage.value]?.output || ''
})

const currentOutputFiles = computed(() => {
  if (!wf.value?.stages) return []
  return wf.value.stages[viewingStage.value]?.output_files || []
})

// Methods
function stageState(stage) {
  return wf.value?.stages?.[stage]?.state || 'pending'
}

function stageTabClass(stage) {
  const state = stageState(stage)
  const isViewing = stage === viewingStage.value
  const isCurrent = stage === currentStage.value

  if (isViewing) {
    return 'bg-pink-500/15 text-pink-400 border border-pink-500/20'
  }
  if (state === 'approved') {
    return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/10'
  }
  if (state === 'generating') {
    return 'bg-amber-500/10 text-amber-400 border border-amber-500/10'
  }
  if (state === 'awaiting_approval') {
    return 'bg-amber-500/10 text-amber-400 border border-amber-500/10'
  }
  if (isCurrent) {
    return 'bg-makoclaw-surface/50 text-makoclaw-text border border-makoclaw-border/30'
  }
  return 'text-makoclaw-text-secondary hover:bg-makoclaw-surface/30 border border-transparent'
}

async function handleStart() {
  await store.startWorkflow(props.account, props.campaign)
}

async function handleGenerate() {
  try {
    const result = await store.generateStage(props.account, props.campaign, {
      objective: generateParams.value.objective || undefined,
      target_audience: generateParams.value.target_audience || undefined,
      platforms: generateParams.value.platforms || undefined,
      tone: generateParams.value.tone || undefined,
      post_count: generateParams.value.post_count || undefined,
      image_style: generateParams.value.image_style || undefined
    })
  } catch (e) {
    console.error('generate failed', e)
  }
}

async function handleApprove() {
  try {
    await store.approveStage(props.account, props.campaign)
    // Sync view to the new current stage (backend already advanced it)
    selectedStageView.value = currentStage.value
  } catch (e) {
    console.error('approve failed', e)
  }
}

async function handleReject() {
  if (!rejectFeedback.value.trim()) return
  try {
    await store.rejectStage(props.account, props.campaign, rejectFeedback.value.trim())
    showRejectModal.value = false
    rejectFeedback.value = ''
  } catch (e) {
    console.error('reject failed', e)
  }
}

async function handleSkip() {
  try {
    await store.skipStage(props.account, props.campaign)
    // Sync view to the new current stage (backend already advanced it)
    selectedStageView.value = currentStage.value
  } catch (e) {
    console.error('skip failed', e)
  }
}

async function handleReset(stage) {
  try {
    await store.resetStage(props.account, props.campaign, stage)
    selectedStageView.value = stage
  } catch (e) {
    console.error('reset failed', e)
  }
}

// Lifecycle
onMounted(async () => {
  await store.fetchWorkflowStatus(props.account, props.campaign)
  if (currentStage.value) {
    selectedStageView.value = currentStage.value
  }
})

// Watch campaign changes
watch(() => [props.account, props.campaign], async () => {
  await store.fetchWorkflowStatus(props.account, props.campaign)
  if (currentStage.value) {
    selectedStageView.value = currentStage.value
  }
})
</script>
