<template>
  <div class="h-full flex flex-col bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-rose-500/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-fuchsia-500/20 via-transparent to-transparent" />
    </div>

    <!-- Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <div class="flex items-center gap-3">
          <!-- Icon Container -->
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-rose-500/20 to-fuchsia-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-rose-500/10">
            <svg
              class="w-5 h-5 sm:w-6 sm:h-6 text-rose-400"
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
          </div>

          <!-- Title -->
          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-rose-400 bg-clip-text text-transparent">
              {{ editing ? (editingWorkflow.id ? 'Edit Workflow' : 'New Workflow') : 'Workflows' }}
            </h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5">
              {{ editing ? 'Visual pipeline builder' : 'Create and manage automation pipelines' }}
            </p>
          </div>

          <!-- Actions -->
          <div class="flex items-center gap-2">
            <button
              v-if="editing"
              class="flex items-center gap-2 px-3 sm:px-4 py-2.5 min-h-[40px] text-makoclaw-text-secondary hover:text-makoclaw-text border border-makoclaw-border/50 rounded-xl transition-all text-sm backdrop-blur-sm hover:bg-makoclaw-surface/30"
              @click="cancelEdit"
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              /></svg>
              <span class="hidden sm:inline">Cancel</span>
            </button>
            <button
              v-if="editing"
              :disabled="saving"
              class="flex items-center gap-2 px-4 sm:px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-rose-500 to-rose-600 hover:from-rose-600 hover:to-rose-700 text-white rounded-xl transition-all shadow-lg shadow-rose-500/25 text-sm font-bold active:scale-95 disabled:opacity-50"
              @click="saveWorkflow"
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M5 13l4 4L19 7"
              /></svg>
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
            <button
              v-if="!editing"
              class="flex items-center gap-2 px-4 sm:px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-rose-500 to-rose-600 hover:from-rose-600 hover:to-rose-700 text-white rounded-xl transition-all shadow-lg shadow-rose-500/25 text-sm font-bold active:scale-95 flex-shrink-0"
              @click="startCreate"
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              /></svg>
              <span class="hidden sm:inline">New Workflow</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-4 md:p-6 custom-scrollbar relative">
      <!-- Loading -->
      <div
        v-if="loading"
        class="flex items-center justify-center py-12"
      >
        <div class="relative">
          <div class="absolute inset-0 bg-gradient-to-br from-rose-500/20 to-fuchsia-500/20 rounded-full blur-xl animate-pulse" />
          <div class="relative animate-spin rounded-full h-10 w-10 border-2 border-transparent border-t-rose-400 border-r-rose-400" />
        </div>
      </div>

      <!-- ===== LIST VIEW ===== -->
      <template v-else-if="!editing">
        <!-- Empty State -->
        <div
          v-if="workflows.length === 0"
          class="flex flex-col items-center justify-center py-12 text-center"
        >
          <div class="relative">
            <div class="absolute inset-0 bg-gradient-to-br from-rose-500/30 to-fuchsia-500/30 rounded-3xl blur-2xl opacity-50" />
            <div class="relative glass-panel p-8 rounded-2xl shadow-2xl ring-1 ring-white/10">
              <div class="w-16 h-16 mx-auto rounded-2xl bg-gradient-to-br from-rose-500/20 to-fuchsia-500/20 flex items-center justify-center ring-1 ring-white/20">
                <svg
                  class="w-8 h-8 text-rose-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="1.5"
                    d="M13 10V3L4 14h7v7l9-11h-7z"
                  />
                </svg>
              </div>
            </div>
          </div>
          <h3 class="text-lg font-bold text-makoclaw-text mt-6">
            No workflows yet
          </h3>
          <p class="text-sm text-makoclaw-text-secondary/70 mt-2 max-w-xs">
            Create a workflow to automate multi-step pipelines
          </p>
          <button
            class="mt-6 px-5 py-2.5 bg-gradient-to-r from-rose-500 to-rose-600 hover:from-rose-600 hover:to-rose-700 text-white rounded-xl font-bold shadow-lg shadow-rose-500/25 transition-all active:scale-95 flex items-center gap-2"
            @click="startCreate"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 4v16m8-8H4"
            /></svg>
            Create Your First Workflow
          </button>
        </div>

        <!-- Workflow Cards -->
        <div class="space-y-3">
          <div
            v-for="wf in workflows"
            :key="wf.id"
            class="workflow-card p-5 rounded-xl"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-rose-500/20 to-fuchsia-500/20 flex items-center justify-center ring-1 ring-white/10">
                    <svg
                      class="w-4 h-4 text-rose-400"
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
                  </div>
                  <h3 class="font-bold text-makoclaw-text">
                    {{ wf.name }}
                  </h3>
                  <span
                    class="px-2.5 py-0.5 text-xs font-semibold rounded-full"
                    :class="wf.enabled ? 'bg-rose-500/15 text-rose-400 ring-1 ring-rose-500/30' : 'bg-gray-500/10 text-gray-400 ring-1 ring-gray-500/20'"
                  >
                    {{ wf.enabled ? 'Enabled' : 'Disabled' }}
                  </span>
                </div>
                <p
                  v-if="wf.description"
                  class="text-sm text-makoclaw-text-secondary mt-2 ml-10"
                >
                  {{ wf.description }}
                </p>
              </div>
            </div>

            <div class="flex items-center gap-4 mt-3 ml-10 text-xs text-makoclaw-text-secondary">
              <span class="flex items-center gap-1.5">
                <svg
                  class="w-3.5 h-3.5 text-rose-400/60"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M4 6h16M4 12h16M4 18h7"
                  />
                </svg>
                {{ countSteps(wf) }} steps
              </span>
              <span class="flex items-center gap-1.5">
                <svg
                  class="w-3.5 h-3.5 text-fuchsia-400/60"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                  />
                </svg>
                Created {{ formatDate(wf.created_at) }}
              </span>
            </div>

            <div class="flex items-center gap-2 mt-4 ml-10">
              <button
                class="px-3 py-1.5 text-xs font-medium bg-rose-500/10 text-rose-400 rounded-lg hover:bg-rose-500/20 transition-colors ring-1 ring-rose-500/20"
                @click="startEdit(wf)"
              >
                Edit
              </button>
              <button
                :disabled="runningId === wf.id"
                class="px-3 py-1.5 text-xs font-medium bg-fuchsia-500/10 text-fuchsia-400 rounded-lg hover:bg-fuchsia-500/20 transition-colors ring-1 ring-fuchsia-500/20 disabled:opacity-50"
                @click="runWorkflow(wf)"
              >
                {{ runningId === wf.id ? 'Running...' : 'Run' }}
              </button>
              <button
                class="px-3 py-1.5 text-xs font-medium bg-makoclaw-accent/10 text-makoclaw-accent rounded-lg hover:bg-makoclaw-accent/20 transition-colors ring-1 ring-makoclaw-accent/20"
                @click="showRuns(wf)"
              >
                History
              </button>
              <button
                class="px-3 py-1.5 text-xs font-medium text-red-400 bg-red-500/10 rounded-lg hover:bg-red-500/20 transition-colors ring-1 ring-red-500/20"
                @click="deleteWorkflow(wf)"
              >
                Delete
              </button>
            </div>

            <!-- Inline Run Results -->
            <div
              v-if="wf._lastResults"
              class="mt-4 border-t border-makoclaw-border/30 pt-4 ml-10"
            >
              <h4 class="text-xs font-bold text-makoclaw-text-secondary mb-3 uppercase tracking-wider">
                Last Run Results
              </h4>
              <div class="space-y-2">
                <div
                  v-for="(res, i) in wf._lastResults"
                  :key="i"
                  class="text-xs p-3 rounded-lg bg-makoclaw-bg/60 border border-makoclaw-border/30 backdrop-blur-sm animate-fadeUp"
                  :style="{ animationDelay: `${i * 50}ms` }"
                >
                  <div class="flex items-center gap-2 mb-1">
                    <span class="font-mono font-bold text-makoclaw-text">{{ res.label || res.step_id }}</span>
                    <span
                      class="px-2 py-0.5 rounded-full text-[10px] font-semibold"
                      :class="res.error ? 'bg-red-500/15 text-red-400' : res.skipped ? 'bg-gray-500/10 text-gray-400' : 'bg-rose-500/15 text-rose-400'"
                    >
                      {{ res.error ? 'Error' : res.skipped ? 'Skipped' : 'OK' }}
                    </span>
                    <span
                      v-if="res.duration_ms"
                      class="text-makoclaw-text-secondary"
                    >{{ res.duration_ms }}ms</span>
                  </div>
                  <pre
                    v-if="res.output"
                    class="whitespace-pre-wrap text-makoclaw-text-secondary max-h-24 overflow-auto custom-scrollbar"
                  >{{ res.output }}</pre>
                  <pre
                    v-if="res.error"
                    class="whitespace-pre-wrap text-red-400 max-h-24 overflow-auto custom-scrollbar"
                  >{{ res.error }}</pre>
                </div>
              </div>
            </div>

            <!-- Inline Runs History -->
            <div
              v-if="wf._runs"
              class="mt-4 border-t border-makoclaw-border/30 pt-4 ml-10"
            >
              <div class="flex items-center justify-between mb-3">
                <h4 class="text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider">
                  Recent Runs
                </h4>
                <button
                  class="text-xs text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
                  @click="wf._runs = null"
                >
                  Hide
                </button>
              </div>
              <div
                v-if="wf._runs.length === 0"
                class="text-xs text-makoclaw-text-secondary/60"
              >
                No runs yet
              </div>
              <div class="space-y-1.5">
                <div
                  v-for="run in wf._runs"
                  :key="run.id"
                  class="text-xs flex items-center gap-3 p-2.5 rounded-lg bg-makoclaw-bg/60 border border-makoclaw-border/30"
                >
                  <span
                    class="px-2 py-0.5 rounded-full font-semibold"
                    :class="run.status === 'completed' ? 'bg-rose-500/15 text-rose-400' : run.status === 'running' ? 'bg-fuchsia-500/15 text-fuchsia-400' : 'bg-red-500/15 text-red-400'"
                  >
                    {{ run.status }}
                  </span>
                  <span class="text-makoclaw-text-secondary">{{ formatDate(run.started_at) }}</span>
                  <span
                    v-if="run.finished_at"
                    class="text-makoclaw-text-secondary/60"
                  >
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
        <div class="glass-panel rounded-2xl p-5 sm:p-6 mb-6 ring-1 ring-white/10">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div>
              <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Name</label>
              <input
                v-model="editingWorkflow.name"
                type="text"
                placeholder="My Workflow"
                class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 transition-all backdrop-blur-sm min-h-[40px]"
              >
            </div>
            <div>
              <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Description</label>
              <input
                v-model="editingWorkflow.description"
                type="text"
                placeholder="What does this workflow do?"
                class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 transition-all backdrop-blur-sm min-h-[40px]"
              >
            </div>
          </div>

          <!-- Parameters Section -->
          <div class="border-t border-makoclaw-border/30 pt-4">
            <div class="flex items-center justify-between mb-3">
              <label class="block text-xs font-bold text-makoclaw-text-secondary uppercase tracking-wider">Input Parameters</label>
              <button
                class="text-xs px-3 py-1.5 bg-gradient-to-r from-rose-500/10 to-fuchsia-500/10 text-rose-400 rounded-lg hover:from-rose-500/20 hover:to-fuchsia-500/20 transition-colors ring-1 ring-rose-500/20 font-medium"
                @click="addParameter"
              >
                + Add Parameter
              </button>
            </div>
            <div
              v-if="editingWorkflow.parameters.length === 0"
              class="text-xs text-makoclaw-text-secondary/60 italic p-3 bg-makoclaw-bg/30 rounded-lg border border-dashed border-makoclaw-border/30"
            >
              No parameters defined yet. Add parameters to make your workflow reusable.
            </div>
            <div
              v-else
              class="space-y-2"
            >
              <div
                v-for="(param, idx) in editingWorkflow.parameters"
                :key="idx"
                class="flex items-center gap-2 p-2 bg-makoclaw-bg/30 rounded-lg border border-makoclaw-border/30"
              >
                <input
                  v-model="param.name"
                  type="text"
                  placeholder="parameter_name"
                  class="w-32 px-2.5 py-1.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-lg text-xs focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 font-mono"
                >
                <input
                  v-model="param.label"
                  type="text"
                  placeholder="Display label"
                  class="flex-1 px-2.5 py-1.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-lg text-xs focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50"
                >
                <input
                  v-model="param.default_value"
                  type="text"
                  placeholder="Default value (optional)"
                  class="flex-1 px-2.5 py-1.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-lg text-xs focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50"
                >
                <button
                  class="p-1.5 text-makoclaw-text-secondary hover:text-red-400 transition-colors rounded-lg hover:bg-red-500/10"
                  @click="removeParameter(idx)"
                >
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
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Pipeline Steps -->
        <div class="mb-4">
          <h3 class="text-sm font-bold text-makoclaw-text mb-3 flex items-center gap-2">
            <svg
              class="w-4 h-4 text-rose-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 6h16M4 12h16M4 18h7"
              />
            </svg>
            Pipeline Steps
          </h3>

          <!-- Empty state when no steps -->
          <div
            v-show="editingWorkflow.steps.length === 0"
            class="text-center py-8 border-2 border-dashed border-rose-500/20 rounded-xl text-makoclaw-text-secondary mb-4 bg-rose-500/5"
          >
            <svg
              class="w-10 h-10 mx-auto mb-3 text-rose-400/40"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M12 6v6m0 0v6m0-6h6m-6 0H6"
              />
            </svg>
            <p class="text-sm">
              No steps yet. Add a step below to start building your pipeline.
            </p>
          </div>

          <!-- Steps list with VueDraggable -->
          <VueDraggable
            v-model="editingWorkflow.steps"
            :animation="200"
            handle=".drag-handle"
            class="space-y-3"
          >
            <div
              v-for="(step, index) in editingWorkflow.steps"
              :key="step.id"
              class="step-card rounded-xl overflow-hidden"
              :class="expandedStep === step.id ? 'ring-2 ring-rose-500/30' : ''"
            >
              <!-- Step Header -->
              <div
                class="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-makoclaw-surface/30 transition-colors"
                @click="toggleStep(step.id)"
              >
                <div class="drag-handle cursor-grab text-makoclaw-text-secondary hover:text-makoclaw-text p-1 -m-1">
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
                      d="M4 8h16M4 16h16"
                    />
                  </svg>
                </div>
                <div class="flex items-center gap-2 flex-1 min-w-0">
                  <span
                    class="w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-bold"
                    :class="stepTypeClass(step.type)"
                  >
                    {{ index + 1 }}
                  </span>
                  <span
                    class="px-2 py-0.5 text-[10px] rounded-full font-bold uppercase tracking-wider"
                    :class="stepTypeClass(step.type)"
                  >
                    {{ step.type }}
                  </span>
                  <span class="text-sm text-makoclaw-text truncate">{{ step.label || 'Untitled step' }}</span>
                </div>
                <button
                  class="p-1.5 text-makoclaw-text-secondary hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all"
                  @click.stop="removeStep(index, step.id)"
                >
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
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                </button>
                <svg
                  class="w-4 h-4 text-makoclaw-text-secondary transition-transform duration-200"
                  :class="expandedStep === step.id ? 'rotate-180' : ''"
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

              <!-- Step Config (expanded) -->
              <Transition name="expand">
                <div
                  v-if="expandedStep === step.id"
                  class="px-4 pb-4 border-t border-makoclaw-border/30 pt-4 space-y-4 bg-makoclaw-bg/30"
                >
                  <div>
                    <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Label</label>
                    <input
                      v-model="step.label"
                      type="text"
                      placeholder="Step label"
                      class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 transition-all min-h-[40px]"
                    >
                  </div>

                  <div>
                    <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">On Error</label>
                    <select
                      v-model="step.on_error"
                      class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 transition-all min-h-[40px] cursor-pointer"
                    >
                      <option value="stop">
                        Stop workflow
                      </option>
                      <option value="continue">
                        Continue to next step
                      </option>
                    </select>
                  </div>

                  <!-- Prompt Config -->
                  <template v-if="step.type === 'prompt'">
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                        Message
                        <span class="font-normal normal-case opacity-60 ml-1">(supports {{ templateHint }} templates)</span>
                      </label>
                      <textarea
                        v-model="step._config.message"
                        rows="4"
                        placeholder="Enter prompt message..."
                        class="w-full px-4 py-3 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 font-mono resize-y transition-all"
                      />
                    </div>
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Model Override (optional)</label>
                      <input
                        v-model="step._config.model"
                        type="text"
                        placeholder="Leave empty for default"
                        class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 transition-all min-h-[40px]"
                      >
                    </div>
                  </template>

                  <!-- Tool Config -->
                  <template v-if="step.type === 'tool'">
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Tool Name</label>
                      <select
                        v-if="availableTools.length > 0"
                        v-model="step._config.tool_name"
                        class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 transition-all min-h-[40px] cursor-pointer"
                      >
                        <option value="">
                          Select a tool...
                        </option>
                        <option
                          v-for="t in availableTools"
                          :key="t"
                          :value="t"
                        >
                          {{ t }}
                        </option>
                      </select>
                      <input
                        v-else
                        v-model="step._config.tool_name"
                        type="text"
                        placeholder="tool_name"
                        class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 transition-all min-h-[40px]"
                      >
                    </div>
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                        Arguments (JSON)
                        <span class="font-normal normal-case opacity-60 ml-1">(string values support {{ templateHint }} templates)</span>
                      </label>
                      <textarea
                        v-model="step._config._argsJson"
                        rows="4"
                        placeholder="{&quot;key&quot;: &quot;value&quot;}"
                        class="w-full px-4 py-3 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 font-mono resize-y transition-all"
                        :class="step._config._argsError ? 'border-red-500 ring-1 ring-red-500/30' : ''"
                      />
                      <p
                        v-if="step._config._argsError"
                        class="text-xs text-red-400 mt-2 flex items-center gap-1"
                      >
                        <svg
                          class="w-3.5 h-3.5"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                          />
                        </svg>
                        {{ step._config._argsError }}
                      </p>
                    </div>
                  </template>

                  <!-- Condition Config -->
                  <template v-if="step.type === 'condition'">
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                        Reference
                        <span class="font-normal normal-case opacity-60 ml-1">(e.g. {{ templateExample }})</span>
                      </label>
                      <input
                        v-model="step._config.reference"
                        type="text"
                        placeholder="{{step.1.output}}"
                        class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 font-mono transition-all min-h-[40px]"
                      >
                    </div>
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Operator</label>
                      <select
                        v-model="step._config.operator"
                        class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 transition-all min-h-[40px] cursor-pointer"
                      >
                        <option value="contains">
                          Contains
                        </option>
                        <option value="equals">
                          Equals
                        </option>
                        <option value="not_empty">
                          Not Empty
                        </option>
                        <option value="regex">
                          Regex Match
                        </option>
                      </select>
                    </div>
                    <div v-if="step._config.operator !== 'not_empty'">
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Value</label>
                      <input
                        v-model="step._config.value"
                        type="text"
                        placeholder="Compare value"
                        class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 transition-all min-h-[40px]"
                      >
                    </div>
                  </template>

                  <!-- Loop Config -->
                  <template v-if="step.type === 'loop'">
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                        Items
                        <span class="font-normal normal-case opacity-60 ml-1">(comma-separated or template)</span>
                      </label>
                      <input
                        v-model="step._config.items"
                        type="text"
                        placeholder="item1, item2, item3"
                        class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500/30 focus:border-cyan-500/50 font-mono transition-all min-h-[40px]"
                      >
                    </div>
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Variable Name</label>
                      <input
                        v-model="step._config.variable"
                        type="text"
                        placeholder="item"
                        class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500/30 focus:border-cyan-500/50 transition-all min-h-[40px]"
                      >
                    </div>
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Nested Steps (JSON)</label>
                      <textarea
                        v-model="step._config._stepsJson"
                        rows="4"
                        placeholder='[{"id":"s1","type":"prompt","label":"Process","config":{"message":"Process item"}}]'
                        class="w-full px-4 py-3 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500/30 focus:border-cyan-500/50 font-mono resize-y transition-all"
                      />
                    </div>
                  </template>

                  <!-- Parallel Config -->
                  <template v-if="step.type === 'parallel'">
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Max Concurrency</label>
                      <input
                        v-model.number="step._config.max_concurrency"
                        type="number"
                        min="1"
                        max="10"
                        placeholder="0 = unlimited"
                        class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/30 focus:border-emerald-500/50 transition-all min-h-[40px]"
                      >
                    </div>
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Parallel Steps (JSON)</label>
                      <textarea
                        v-model="step._config._stepsJson"
                        rows="4"
                        placeholder='[{"id":"s1","type":"tool","label":"Task A","config":{"tool_name":"web_search","args":{}}}]'
                        class="w-full px-4 py-3 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/30 focus:border-emerald-500/50 font-mono resize-y transition-all"
                      />
                    </div>
                  </template>

                  <!-- Approval Config -->
                  <template v-if="step.type === 'approval'">
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">Approval Message</label>
                      <textarea
                        v-model="step._config.message"
                        rows="3"
                        placeholder="Describe what needs to be approved..."
                        class="w-full px-4 py-3 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-orange-500/30 focus:border-orange-500/50 resize-y transition-all"
                      />
                    </div>
                    <div>
                      <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                        Approvers
                        <span class="font-normal normal-case opacity-60 ml-1">(comma-separated, optional)</span>
                      </label>
                      <input
                        v-model="step._config.approvers"
                        type="text"
                        placeholder="admin, manager"
                        class="w-full px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-orange-500/30 focus:border-orange-500/50 transition-all min-h-[40px]"
                      >
                    </div>
                  </template>
                </div>
              </Transition>
            </div>
          </VueDraggable>
        </div>

        <!-- Add Step Buttons -->
        <div class="flex items-center gap-2 mb-6">
          <button
            class="flex items-center gap-2 px-4 py-2.5 text-xs font-bold bg-rose-500/10 text-rose-400 border border-rose-500/20 rounded-xl hover:bg-rose-500/20 transition-all active:scale-95 min-h-[40px]"
            @click="addStep('prompt')"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 4v16m8-8H4"
            /></svg>
            Prompt
          </button>
          <button
            class="flex items-center gap-2 px-4 py-2.5 text-xs font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20 rounded-xl hover:bg-amber-500/20 transition-all active:scale-95 min-h-[40px]"
            @click="addStep('tool')"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 4v16m8-8H4"
            /></svg>
            Tool
          </button>
          <button
            class="flex items-center gap-2 px-4 py-2.5 text-xs font-bold bg-fuchsia-500/10 text-fuchsia-400 border border-fuchsia-500/20 rounded-xl hover:bg-fuchsia-500/20 transition-all active:scale-95 min-h-[40px]"
            @click="addStep('condition')"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 4v16m8-8H4"
            /></svg>
            Condition
          </button>
          <button
            class="flex items-center gap-2 px-4 py-2.5 text-xs font-bold bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 rounded-xl hover:bg-cyan-500/20 transition-all active:scale-95 min-h-[40px]"
            @click="addStep('loop')"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
            Loop
          </button>
          <button
            class="flex items-center gap-2 px-4 py-2.5 text-xs font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-xl hover:bg-emerald-500/20 transition-all active:scale-95 min-h-[40px]"
            @click="addStep('parallel')"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" /></svg>
            Parallel
          </button>
          <button
            class="flex items-center gap-2 px-4 py-2.5 text-xs font-bold bg-orange-500/10 text-orange-400 border border-orange-500/20 rounded-xl hover:bg-orange-500/20 transition-all active:scale-95 min-h-[40px]"
            @click="addStep('approval')"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            Approval
          </button>
        </div>

        <!-- Test Run -->
        <div class="glass-panel rounded-2xl p-5 sm:p-6 ring-1 ring-white/10">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-sm font-bold text-makoclaw-text flex items-center gap-2">
              <svg
                class="w-4 h-4 text-fuchsia-400"
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
              Test Run
            </h3>
            <button
              v-if="editingWorkflow.id"
              :disabled="testRunning"
              class="flex items-center gap-2 px-4 py-2 text-xs font-bold bg-gradient-to-r from-fuchsia-500/10 to-rose-500/10 text-fuchsia-400 rounded-xl hover:from-fuchsia-500/20 hover:to-rose-500/20 transition-all ring-1 ring-fuchsia-500/20 disabled:opacity-50 min-h-[36px]"
              @click="testRun"
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
              /><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              /></svg>
              {{ testRunning ? 'Running...' : 'Run Now' }}
            </button>
            <span
              v-else
              class="text-xs text-makoclaw-text-secondary/60 italic"
            >Save the workflow first to test it</span>
          </div>

          <div
            v-if="testResults"
            class="space-y-2"
          >
            <div
              v-for="(res, i) in testResults"
              :key="i"
              class="text-xs p-3 rounded-lg bg-makoclaw-bg/60 border border-makoclaw-border/30 animate-fadeUp"
              :style="{ animationDelay: `${i * 50}ms` }"
            >
              <div class="flex items-center gap-2 mb-2">
                <span class="font-mono font-bold text-makoclaw-text">{{ res.label || res.step_id }}</span>
                <span
                  class="px-2 py-0.5 rounded-full text-[10px] font-semibold"
                  :class="res.error ? 'bg-red-500/15 text-red-400' : res.skipped ? 'bg-gray-500/10 text-gray-400' : 'bg-rose-500/15 text-rose-400'"
                >
                  {{ res.error ? 'Error' : res.skipped ? 'Skipped' : 'OK' }}
                </span>
                <span
                  v-if="res.duration_ms"
                  class="text-makoclaw-text-secondary"
                >{{ res.duration_ms }}ms</span>
              </div>
              <pre
                v-if="res.output"
                class="whitespace-pre-wrap text-makoclaw-text-secondary max-h-32 overflow-auto custom-scrollbar p-2 bg-makoclaw-bg/50 rounded-lg"
              >{{ res.output }}</pre>
              <pre
                v-if="res.error"
                class="whitespace-pre-wrap text-red-400 max-h-32 overflow-auto custom-scrollbar p-2 bg-red-500/5 rounded-lg"
              >{{ res.error }}</pre>
            </div>
          </div>
          <div
            v-else
            class="text-xs text-makoclaw-text-secondary/60 p-4 bg-makoclaw-bg/30 rounded-lg border border-dashed border-makoclaw-border/30 text-center"
          >
            No test results yet. Save and run the workflow to see output here.
          </div>
        </div>
      </template>
    </div>

    <!-- Parameter Input Modal -->
    <Transition name="modal">
      <div
        v-if="showParamModal"
        class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-modal p-4"
        @click.self="cancelParameters"
      >
        <div class="bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl ring-1 ring-white/10 w-full max-w-md max-h-[80vh] overflow-hidden animate-scaleIn">
          <!-- Modal Header -->
          <div class="p-5 border-b border-makoclaw-border/30 bg-gradient-to-r from-rose-500/10 to-fuchsia-500/10">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-rose-500/20 to-fuchsia-500/20 flex items-center justify-center ring-1 ring-white/10">
                <svg
                  class="w-5 h-5 text-rose-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"
                  />
                </svg>
              </div>
              <div>
                <h2 class="text-lg font-bold text-makoclaw-text">
                  Workflow Parameters
                </h2>
                <p class="text-xs text-makoclaw-text-secondary">
                  Provide values for the workflow parameters
                </p>
              </div>
            </div>
          </div>

          <!-- Modal Content -->
          <div class="p-5 overflow-y-auto max-h-[50vh] custom-scrollbar">
            <div class="space-y-4">
              <div
                v-for="param in (paramModalWorkflow?.parameters || [])"
                :key="param.name"
              >
                <label class="block text-xs font-bold text-makoclaw-text-secondary mb-2 uppercase tracking-wider">
                  {{ param.label || param.name }}
                </label>
                <input
                  v-model="paramInputs[param.name]"
                  type="text"
                  :placeholder="param.default_value || 'Enter value...'"
                  class="w-full px-4 py-2.5 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-rose-500/30 focus:border-rose-500/50 transition-all min-h-[40px] backdrop-blur-sm"
                >
              </div>
            </div>
          </div>

          <!-- Modal Footer -->
          <div class="p-5 border-t border-makoclaw-border/30 flex items-center gap-3 bg-makoclaw-bg/30">
            <button
              class="flex-1 px-4 py-2.5 bg-gradient-to-r from-rose-500 to-rose-600 hover:from-rose-600 hover:to-rose-700 text-white rounded-xl transition-all font-bold text-sm shadow-lg shadow-rose-500/25 active:scale-95 min-h-[40px]"
              @click="submitParameters"
            >
              Run Workflow
            </button>
            <button
              class="px-4 py-2.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl hover:bg-makoclaw-surface/50 transition-all text-sm min-h-[40px]"
              @click="cancelParameters"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import workflowService from '../services/workflowService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const loading = ref(true)
const saving = ref(false)
const editing = ref(false)
const workflows = ref([])
const availableTools = ref([])
const expandedStep = ref(null)
const runningId = ref(null)
const testRunning = ref(false)
const testResults = ref(null)

// Parameter modal state
const showParamModal = ref(false)
const paramModalWorkflow = ref(null)
const paramModalIsTest = ref(false)
const paramInputs = reactive({})

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
  steps: [],
  parameters: []
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
    toast.error('Failed to load workflows: ' + extractError(e))
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
  Object.assign(editingWorkflow, { id: null, name: '', description: '', enabled: true, steps: [], parameters: [] })
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

  let parameters = []
  try {
    const raw = typeof wf.parameters === 'string' ? JSON.parse(wf.parameters) : wf.parameters
    parameters = raw || []
  } catch { parameters = [] }

  Object.assign(editingWorkflow, {
    id: wf.id,
    name: wf.name,
    description: wf.description || '',
    enabled: wf.enabled,
    steps,
    parameters
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

// Parameter helpers
function addParameter() {
  editingWorkflow.parameters.push({
    name: '',
    label: '',
    default_value: ''
  })
}

function removeParameter(idx) {
  editingWorkflow.parameters.splice(idx, 1)
}

function defaultConfig(type) {
  if (type === 'prompt') return { message: '', model: '' }
  if (type === 'tool') return { tool_name: '', _argsJson: '{}', _argsError: '' }
  if (type === 'condition') return { operator: 'contains', value: '', reference: '' }
  if (type === 'loop') return { items: '', variable: 'item', _stepsJson: '[]' }
  if (type === 'parallel') return { max_concurrency: 0, _stepsJson: '[]' }
  if (type === 'approval') return { message: '', approvers: '' }
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
  } else if (raw.type === 'loop') {
    let stepsJson = '[]'
    try { stepsJson = JSON.stringify(cfg.steps || [], null, 2) } catch { stepsJson = '[]' }
    step._config = { items: cfg.items || '', variable: cfg.variable || 'item', _stepsJson: stepsJson }
  } else if (raw.type === 'parallel') {
    let stepsJson = '[]'
    try { stepsJson = JSON.stringify(cfg.steps || [], null, 2) } catch { stepsJson = '[]' }
    step._config = { max_concurrency: cfg.max_concurrency || 0, _stepsJson: stepsJson }
  } else if (raw.type === 'approval') {
    step._config = { message: cfg.message || '', approvers: (cfg.approvers || []).join(', ') }
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
    } else if (s.type === 'loop') {
      let steps = []
      try { steps = JSON.parse(s._config._stepsJson || '[]') } catch { /* keep empty */ }
      out.config = { items: s._config.items || '', variable: s._config.variable || 'item', steps }
    } else if (s.type === 'parallel') {
      let steps = []
      try { steps = JSON.parse(s._config._stepsJson || '[]') } catch { /* keep empty */ }
      out.config = { steps, max_concurrency: s._config.max_concurrency || 0 }
    } else if (s.type === 'approval') {
      const approvers = (s._config.approvers || '').split(',').map(a => a.trim()).filter(Boolean)
      out.config = { message: s._config.message || '', approvers }
    }
    return out
  })
}

async function saveWorkflow() {
  if (!editingWorkflow.name.trim()) {
    toast.error('Workflow name is required')
    return
  }
  if (editingWorkflow.steps.length === 0) {
    toast.error('Workflow must have at least one step')
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
        toast.error('Fix JSON errors in tool step arguments before saving')
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
      steps: serializeSteps(editingWorkflow.steps),
      parameters: editingWorkflow.parameters || []
    }

    if (editingWorkflow.id) {
      await workflowService.updateWorkflow(editingWorkflow.id, payload)
      toast.success('Workflow updated successfully')
    } else {
      const created = await workflowService.createWorkflow(payload)
      editingWorkflow.id = created.id
      toast.success('Workflow created successfully')
    }
    await loadWorkflows()
  } catch (e) {
    toast.error('Failed to save: ' + extractError(e))
  } finally {
    saving.value = false
  }
}

async function deleteWorkflow(wf) {
  if (!confirm(`Delete workflow "${wf.name}"?`)) return
  try {
    await workflowService.deleteWorkflow(wf.id)
    toast.success('Workflow deleted successfully')
    await loadWorkflows()
  } catch (e) {
    toast.error('Failed to delete: ' + extractError(e))
  }
}

async function runWorkflow(wf) {
  // Check if workflow has parameters
  let params = []
  try {
    const raw = typeof wf.parameters === 'string' ? JSON.parse(wf.parameters) : wf.parameters
    params = raw || []
  } catch { params = [] }

  if (params.length > 0) {
    // Show parameter input modal
    paramModalWorkflow.value = wf
    paramModalIsTest.value = false
    // Initialize param inputs with default values
    params.forEach(p => {
      paramInputs[p.name] = p.default_value || ''
    })
    showParamModal.value = true
    return
  }

  // No parameters, run directly
  await executeWorkflow(wf, {})
}

async function executeWorkflow(wf, parameters) {
  runningId.value = wf.id
  try {
    const data = await workflowService.runWorkflow(wf.id, parameters)
    wf._lastResults = data.results || []
    toast.success('Workflow executed successfully')
  } catch (e) {
    toast.error('Run failed: ' + extractError(e))
  } finally {
    runningId.value = null
  }
}

async function testRun() {
  if (!editingWorkflow.id) return
  // Auto-save before running
  await saveWorkflow()
  if (saving.value) return

  // Check if workflow has parameters
  const params = editingWorkflow.parameters || []
  if (params.length > 0) {
    // Show parameter input modal
    paramModalWorkflow.value = { id: editingWorkflow.id, parameters: params }
    paramModalIsTest.value = true
    // Initialize param inputs with default values
    params.forEach(p => {
      paramInputs[p.name] = p.default_value || ''
    })
    showParamModal.value = true
    return
  }

  // No parameters, run directly
  await executeTestRun({})
}

async function executeTestRun(parameters) {
  testRunning.value = true
  try {
    const data = await workflowService.runWorkflow(editingWorkflow.id, parameters)
    testResults.value = data.results || []
    toast.success('Test run completed')
  } catch (e) {
    toast.error('Test run failed: ' + extractError(e))
  } finally {
    testRunning.value = false
  }
}

function submitParameters() {
  // Collect parameter values
  const params = {}
  const wf = paramModalWorkflow.value
  let paramDefs = []
  try {
    const raw = typeof wf.parameters === 'string' ? JSON.parse(wf.parameters) : wf.parameters
    paramDefs = raw || []
  } catch { paramDefs = [] }

  paramDefs.forEach(p => {
    params[p.name] = paramInputs[p.name] || ''
  })

  showParamModal.value = false

  // Execute workflow with parameters
  if (paramModalIsTest.value) {
    executeTestRun(params)
  } else {
    const fullWf = workflows.value.find(w => w.id === wf.id)
    if (fullWf) {
      executeWorkflow(fullWf, params)
    }
  }

  // Clear modal state
  paramModalWorkflow.value = null
  Object.keys(paramInputs).forEach(k => delete paramInputs[k])
}

function cancelParameters() {
  showParamModal.value = false
  paramModalWorkflow.value = null
  Object.keys(paramInputs).forEach(k => delete paramInputs[k])
}

async function showRuns(wf) {
  try {
    const data = await workflowService.getWorkflowRuns(wf.id)
    wf._runs = data.runs || []
  } catch (e) {
    toast.error('Failed to load runs: ' + extractError(e))
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
  if (type === 'prompt') return 'bg-rose-500/15 text-rose-400'
  if (type === 'tool') return 'bg-amber-500/15 text-amber-400'
  if (type === 'condition') return 'bg-fuchsia-500/15 text-fuchsia-400'
  if (type === 'loop') return 'bg-cyan-500/15 text-cyan-400'
  if (type === 'parallel') return 'bg-emerald-500/15 text-emerald-400'
  if (type === 'approval') return 'bg-orange-500/15 text-orange-400'
  return 'bg-gray-500/10 text-gray-400'
}
</script>

<style scoped>
.workflow-card {
  @apply bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 hover:bg-makoclaw-surface/50 hover:border-rose-500/20 transition-all duration-200;
}

.step-card {
  @apply bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 transition-all duration-200;
}
</style>
