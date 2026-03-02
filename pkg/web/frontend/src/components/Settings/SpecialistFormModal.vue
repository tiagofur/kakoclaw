<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="show"
        class="fixed inset-0 z-modal flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
        @click.self="close"
      >
        <div
          class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl w-full max-w-3xl max-h-[90vh] overflow-hidden flex flex-col ring-1 ring-white/10 animate-scaleIn"
          @click.stop
        >
          <!-- Header -->
          <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div
                  :class="[
                    'w-10 h-10 rounded-xl flex items-center justify-center shadow-lg',
                    mode === 'create'
                      ? 'bg-gradient-to-br from-lime-500/20 to-green-500/20 shadow-lime-500/10'
                      : 'bg-gradient-to-br from-blue-500/20 to-cyan-500/20 shadow-blue-500/10'
                  ]"
                >
                  <svg
                    v-if="mode === 'create'"
                    class="w-5 h-5 text-lime-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                    />
                  </svg>
                  <svg
                    v-else
                    class="w-5 h-5 text-blue-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                    />
                  </svg>
                </div>
                <div>
                  <h3 class="text-sm font-bold text-makoclaw-text">
                    {{ mode === 'create' ? 'Create Specialist' : 'Edit Specialist' }}
                  </h3>
                  <p class="text-xs text-makoclaw-text-secondary mt-0.5">
                    {{ mode === 'create' ? 'Configure a new autonomous agent' : 'Update agent parameters' }}
                  </p>
                </div>
              </div>
              <button
                class="p-2 rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text hover:bg-makoclaw-surface-hover transition-colors"
                @click="close"
              >
                <svg
                  class="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>
          </div>

          <!-- Scrollable Content -->
          <div class="flex-1 overflow-y-auto p-5 space-y-6 custom-scrollbar">

            <!-- AI Generation Section (create mode only) -->
            <div
              v-if="mode === 'create'"
              class="relative"
            >
              <div class="bg-makoclaw-surface/30 border border-makoclaw-border/50 rounded-xl p-5 hover:border-makoclaw-accent/20 transition-all">
                <div class="flex items-start gap-4">
                  <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-makoclaw-accent/20 to-indigo-500/20 flex items-center justify-center flex-shrink-0 ring-1 ring-white/10">
                    <svg
                      class="w-5 h-5 text-makoclaw-accent"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                      />
                    </svg>
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2 mb-1">
                      <h4 class="text-sm font-bold text-makoclaw-text">
                        AI Generation
                      </h4>
                      <span class="badge-info">AI-Ready</span>
                    </div>
                    <p class="text-xs text-makoclaw-text-secondary/70 mb-4">
                      Describe the specialist's purpose and the AI will configure the model, prompt, and tools automatically.
                    </p>

                    <div class="space-y-3">
                      <textarea
                        v-model="aiPrompt"
                        rows="3"
                        placeholder="e.g. 'Create a security specialist focused on log analysis and vulnerability detection...'"
                        class="w-full px-4 py-3 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 resize-none backdrop-blur-sm"
                      />

                      <button
                        :disabled="!aiPrompt.trim() || agentsStore.aiGenerating"
                        class="px-5 py-2.5 bg-gradient-to-r from-makoclaw-accent to-indigo-600 hover:from-makoclaw-accent-hover hover:to-indigo-700 text-white rounded-xl text-sm font-bold shadow-lg shadow-makoclaw-accent/25 transition-all flex items-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed active:scale-95"
                        @click="generateWithAI"
                      >
                        <svg
                          v-if="agentsStore.aiGenerating"
                          class="w-4 h-4 animate-spin"
                          fill="none"
                          viewBox="0 0 24 24"
                        >
                          <circle
                            class="opacity-25"
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            stroke-width="4"
                          />
                          <path
                            class="opacity-75"
                            fill="currentColor"
                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                          />
                        </svg>
                        <svg
                          v-else
                          class="w-4 h-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M13 10V3L4 14h7v7l9-11h-7z"
                          />
                        </svg>
                        {{ agentsStore.aiGenerating ? 'Generating...' : 'Generate with AI' }}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Section Divider -->
            <div class="flex items-center gap-3">
              <div class="h-px bg-makoclaw-border/30 flex-1" />
              <span class="text-xs font-medium text-makoclaw-text-secondary/50">Manual Configuration</span>
              <div class="h-px bg-makoclaw-border/30 flex-1" />
            </div>

            <!-- Manual Configuration -->
            <div class="space-y-5">

              <!-- Row 1: Name + Provider -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
                <div class="space-y-1.5">
                  <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">
                    Name *
                  </label>
                  <input
                    v-model="form.name"
                    type="text"
                    placeholder="e.g., security_analyst"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
                  >
                </div>

                <div class="space-y-1.5">
                  <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">
                    Provider
                  </label>
                  <div class="relative">
                    <select
                      v-model="form.provider"
                      class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none appearance-none cursor-pointer backdrop-blur-sm"
                    >
                      <option
                        v-for="p in providersList"
                        :key="p.name"
                        :value="p.name"
                      >
                        {{ p.name.toUpperCase() }}
                      </option>
                    </select>
                    <div class="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-makoclaw-text-secondary/40">
                      <svg
                        class="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M19 9l-7 7-7-7"
                        />
                      </svg>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Row 2: Model + Description -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
                <div class="space-y-1.5">
                  <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">
                    Model
                  </label>
                  <div class="relative">
                    <select
                      v-model="form.model"
                      class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-accent focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none appearance-none cursor-pointer backdrop-blur-sm"
                    >
                      <optgroup
                        v-for="p in providersList"
                        :key="p.name"
                        :label="p.name.toUpperCase()"
                        class="text-makoclaw-text-secondary bg-makoclaw-surface"
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
                    <div class="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-makoclaw-text-secondary/40">
                      <svg
                        class="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M19 9l-7 7-7-7"
                        />
                      </svg>
                    </div>
                  </div>
                </div>

                <div class="space-y-1.5">
                  <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">
                    Description *
                  </label>
                  <input
                    v-model="form.description"
                    type="text"
                    placeholder="Primary objective of this specialist..."
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
                  >
                </div>
              </div>

              <!-- System Prompt -->
              <div class="space-y-1.5">
                <label class="text-xs font-medium text-makoclaw-text-secondary ml-1 flex items-center gap-2">
                  <svg
                    class="w-3.5 h-3.5 text-makoclaw-accent"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                    />
                  </svg>
                  System Prompt
                </label>
                <textarea
                  v-model="form.prompt"
                  rows="5"
                  placeholder="Define the specialist's behavior, personality, and constraints..."
                  class="w-full px-4 py-3 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm font-mono text-makoclaw-text-secondary leading-relaxed focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none resize-none placeholder:text-makoclaw-text-secondary/30 backdrop-blur-sm"
                />
              </div>

              <!-- Parameters Row -->
              <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div class="space-y-1.5">
                  <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">
                    Max Tokens
                  </label>
                  <input
                    v-model.number="form.max_tokens"
                    type="number"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none backdrop-blur-sm"
                  >
                </div>
                <div class="space-y-1.5">
                  <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">
                    Temperature
                  </label>
                  <input
                    v-model.number="form.temperature"
                    type="number"
                    step="0.1"
                    min="0"
                    max="2"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none backdrop-blur-sm"
                  >
                </div>
                <div class="space-y-1.5">
                  <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">
                    Max Tool Iterations
                  </label>
                  <input
                    v-model.number="form.max_tool_iterations"
                    type="number"
                    class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none backdrop-blur-sm"
                  >
                </div>
              </div>

              <!-- Tools -->
              <div class="space-y-2">
                <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">
                  Tools
                </label>
                <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
                  <label
                    v-for="tool in availableTools"
                    :key="tool"
                    :class="[
                      'flex items-center gap-2.5 px-3 py-2.5 rounded-xl border cursor-pointer transition-all duration-200 text-sm',
                      form.tools.includes(tool)
                        ? 'bg-makoclaw-accent/5 border-makoclaw-accent/40 text-makoclaw-accent'
                        : 'bg-makoclaw-bg/30 border-makoclaw-border/50 text-makoclaw-text-secondary hover:border-makoclaw-accent/20 hover:bg-makoclaw-surface-hover'
                    ]"
                  >
                    <input
                      v-model="form.tools"
                      :value="tool"
                      type="checkbox"
                      class="w-3.5 h-3.5 rounded border-makoclaw-border bg-transparent text-makoclaw-accent focus:ring-makoclaw-accent/30 cursor-pointer"
                    >
                    <span class="text-xs font-medium truncate">{{ tool }}</span>
                  </label>
                </div>
              </div>

              <!-- Keywords -->
              <div class="space-y-1.5">
                <label class="text-xs font-medium text-makoclaw-text-secondary ml-1">
                  Keywords
                </label>
                <input
                  v-model="keywordsInput"
                  type="text"
                  placeholder="forensics, logs, security, linux..."
                  class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm text-makoclaw-text focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all outline-none placeholder:text-makoclaw-text-secondary/40 backdrop-blur-sm"
                >
                <p class="text-[11px] text-makoclaw-text-secondary/40 ml-1">
                  Comma-separated keywords for routing
                </p>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="p-4 border-t border-makoclaw-border/30 bg-makoclaw-surface/30 flex flex-col sm:flex-row justify-between items-center gap-3">
            <button
              class="w-full sm:w-auto btn-ghost text-makoclaw-text-secondary/60 hover:text-makoclaw-error"
              @click="close"
            >
              Cancel
            </button>

            <div class="flex gap-3 w-full sm:w-auto">
              <button
                v-if="mode === 'edit' && specialist"
                :disabled="agentsStore.loading"
                class="btn-secondary flex items-center gap-2 disabled:opacity-40"
                @click="testSpecialist"
              >
                <svg
                  v-if="agentsStore.loading"
                  class="w-4 h-4 animate-spin"
                  fill="none"
                  viewBox="0 0 24 24"
                >
                  <circle
                    class="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    stroke-width="4"
                  />
                  <path
                    class="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
                <svg
                  v-else
                  class="w-4 h-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
                  />
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                Test
              </button>

              <button
                :disabled="!isFormValid || saving"
                class="flex-1 sm:flex-none px-5 py-2.5 bg-gradient-to-r from-makoclaw-accent to-makoclaw-accent-hover text-white rounded-xl text-sm font-bold shadow-lg shadow-makoclaw-accent/25 transition-all flex items-center justify-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed active:scale-95"
                @click="save"
              >
                <svg
                  v-if="saving"
                  class="w-4 h-4 animate-spin"
                  fill="none"
                  viewBox="0 0 24 24"
                >
                  <circle
                    class="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    stroke-width="4"
                  />
                  <path
                    class="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
                <svg
                  v-else
                  class="w-4 h-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                {{ mode === 'create' ? 'Create Specialist' : 'Save Changes' }}
              </button>
            </div>
          </div>
        </div>

        <!-- Test Result Modal (Nested) -->
        <Transition name="modal">
          <div
            v-if="showTestResultModal"
            class="fixed inset-0 z-modal-nested flex items-center justify-center p-6 bg-black/70 backdrop-blur-sm"
            @click.self="showTestResultModal = false"
          >
            <div
              class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col ring-1 ring-white/10 animate-scaleIn"
              @click.stop
            >
              <!-- Test Modal Header -->
              <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-3">
                    <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center">
                      <svg
                        class="w-4 h-4 text-blue-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          stroke-width="2"
                          d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                        />
                      </svg>
                    </div>
                    <div>
                      <h3 class="text-sm font-bold text-makoclaw-text">
                        Test Result
                      </h3>
                      <p class="text-xs text-makoclaw-text-secondary">
                        Diagnostic output
                      </p>
                    </div>
                  </div>
                  <button
                    class="p-2 rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text hover:bg-makoclaw-surface-hover transition-colors"
                    @click="showTestResultModal = false"
                  >
                    <svg
                      class="w-5 h-5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M6 18L18 6M6 6l12 12"
                      />
                    </svg>
                  </button>
                </div>
              </div>

              <!-- Test Output -->
              <div class="flex-1 overflow-y-auto p-5 custom-scrollbar">
                <div class="bg-makoclaw-bg/60 border border-makoclaw-border/50 rounded-xl p-5 overflow-hidden">
                  <pre class="text-sm font-mono text-makoclaw-text-secondary leading-relaxed whitespace-pre-wrap selection:bg-makoclaw-accent/30">{{ testResult }}</pre>
                </div>
              </div>

              <!-- Test Modal Footer -->
              <div class="p-4 border-t border-makoclaw-border/30">
                <button
                  class="w-full btn-primary py-2.5"
                  @click="showTestResultModal = false"
                >
                  Close
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
import { ref, computed, watch } from 'vue'
import { useAgentsStore } from '../../stores/agentsStore'
import { useToast } from '../../composables/useToast'

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
  if (gen) { form.value = { ...gen }; toast.success('Configuration generated successfully') }
}

async function testSpecialist() {
  if (!props.specialist?.name) return
  const res = await agentsStore.testSpecialist(props.specialist.name, 'Initialize self-diagnostic. Report status.')
  if (res) { testResult.value = res; showTestResultModal.value = true; toast.success('Test complete') }
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
