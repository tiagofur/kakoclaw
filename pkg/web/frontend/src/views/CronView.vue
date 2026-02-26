<template>
  <div class="h-full flex flex-col bg-makoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-makoclaw-border bg-makoclaw-surface flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent">Cron Jobs</h2>
        <p class="text-sm text-makoclaw-text-secondary mt-1">Scheduled tasks and recurring automations</p>
      </div>
      <div class="flex items-center gap-2">
        <button
          @click="openAiCronModal"
          class="flex items-center gap-2 px-4 py-2 bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent-hover transition-colors text-sm"
          title="Generate cron job from natural language"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
          Create with AI
        </button>
        <button
          @click="openCreateModal"
          class="flex items-center gap-2 px-4 py-2 bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors text-sm"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
          New Job
        </button>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-4 md:p-6 custom-scrollbar">
      <!-- Loading Skeleton -->
      <div v-if="loading" class="space-y-3">
        <div v-for="i in 3" :key="i" class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-2">
                <div class="skeleton h-4 w-36 rounded"></div>
                <div class="skeleton h-5 w-14 rounded-full"></div>
                <div class="skeleton h-5 w-20 rounded-full"></div>
              </div>
              <div class="skeleton h-3 w-3/4 rounded"></div>
            </div>
          </div>
          <div class="flex items-center gap-4 mt-3">
            <div class="skeleton h-3 w-40 rounded"></div>
            <div class="skeleton h-3 w-24 rounded"></div>
            <div class="skeleton h-3 w-28 rounded"></div>
          </div>
        </div>
      </div>

      <template v-else>
        <!-- Status Banner -->
        <div class="mb-4 px-4 py-3 rounded-lg border"
          :class="status.enabled ? 'bg-makoclaw-accent/10 border-makoclaw-accent/20 text-makoclaw-accent' : 'bg-makoclaw-warning/10 border-makoclaw-warning/20 text-makoclaw-warning'"
        >
          <span class="font-medium">Cron service: {{ statusLabel }}</span>
          <span v-if="status.jobs !== undefined" class="ml-2 text-sm opacity-75">({{ status.jobs }} active jobs)</span>
        </div>

        <div v-if="jobs.length === 0" class="flex flex-col items-center justify-center py-16 text-center">
          <div class="w-16 h-16 rounded-2xl bg-makoclaw-accent/10 flex items-center justify-center mb-4">
            <svg class="w-8 h-8 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h3 class="font-semibold text-makoclaw-text mb-1">No scheduled jobs</h3>
          <p class="text-sm text-makoclaw-text-secondary max-w-xs mb-4">Create cron jobs to automate tasks on a schedule.</p>
          <button class="btn-primary" @click="openCreateModal">New Job</button>
        </div>

        <div class="space-y-3">
          <div
            v-for="job in jobs"
            :key="job.id"
            class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <h3 class="font-semibold">{{ job.name }}</h3>
                  <span
                    class="px-2 py-0.5 text-xs rounded-full"
                    :class="job.enabled ? 'bg-makoclaw-accent/10 text-makoclaw-accent' : 'bg-makoclaw-text-secondary/10 text-makoclaw-text-secondary'"
                  >{{ job.enabled ? 'Active' : 'Disabled' }}</span>
                  <span class="px-2 py-0.5 text-xs rounded-full bg-makoclaw-text-secondary/10 text-makoclaw-text-secondary">
                    {{ getJobTypeDisplay(job.payload) }}
                  </span>
                </div>
                <p class="text-sm text-makoclaw-text-secondary mt-1">{{ job.payload.message }}</p>
              </div>
            </div>

            <div class="flex items-center gap-4 mt-3 text-xs text-makoclaw-text-secondary">
              <span>Schedule: <span class="text-makoclaw-text font-mono">{{ formatSchedule(job.schedule) }}</span></span>
              <span v-if="job.schedule.tz" class="font-mono">TZ: {{ job.schedule.tz }}</span>
              <span v-if="job.state.lastStatus">Last: {{ job.state.lastStatus }}</span>
              <span v-if="job.state.nextRunAtMs">Next: {{ formatTimestamp(job.state.nextRunAtMs) }}</span>
            </div>

            <div class="flex items-center gap-2 mt-3">
              <button
                @click="runJob(job)"
                class="px-3 py-1.5 text-xs text-makoclaw-accent bg-makoclaw-accent/10 rounded-lg hover:bg-makoclaw-accent/20 transition-colors"
              >Run Now</button>
              <button
                @click="openEditModal(job)"
                class="px-3 py-1.5 text-xs text-makoclaw-text-secondary bg-makoclaw-bg border border-makoclaw-border rounded-lg hover:bg-makoclaw-border transition-colors"
              >Edit</button>
              <button
                @click="toggleJob(job.id, !job.enabled)"
                class="px-3 py-1.5 text-xs rounded-lg transition-colors"
                :class="job.enabled ? 'bg-makoclaw-warning/10 text-makoclaw-warning hover:bg-makoclaw-warning/20' : 'bg-makoclaw-accent/10 text-makoclaw-accent hover:bg-makoclaw-accent/20'"
              >{{ job.enabled ? 'Disable' : 'Enable' }}</button>
              <button
                @click="confirmDeleteJob(job)"
                class="px-3 py-1.5 text-xs text-makoclaw-error bg-makoclaw-error/10 rounded-lg hover:bg-makoclaw-error/20 transition-colors"
              >Delete</button>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Create / Edit Job Modal -->
    <Transition name="modal">
    <div v-if="showModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-modal p-4" @click.self="showModal = false">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-lg w-full max-h-[90vh] overflow-y-auto p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="font-semibold text-lg">{{ editingJobId ? 'Edit Cron Job' : 'Create Cron Job' }}</h3>
          <button
            v-if="editingJobId"
            @click="openJsonEditor"
            class="flex items-center gap-2 px-3 py-1.5 text-xs text-makoclaw-accent bg-makoclaw-accent/10 rounded-lg hover:bg-makoclaw-accent/20 transition-colors"
            title="Edit as JSON"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
            Edit JSON
          </button>
        </div>
        <div class="space-y-4">
          <!-- Name -->
          <div>
            <label class="block text-sm font-medium mb-1">Name</label>
            <input v-model="form.name" type="text" placeholder="My scheduled task"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent" />
          </div>

          <!-- Message -->
          <div>
            <label class="block text-sm font-medium mb-1">Message (what the agent should do)</label>
            <textarea v-model="form.message" rows="3" placeholder="Summarize today's tasks and send a report..."
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent resize-none" />
          </div>

          <!-- Schedule Type -->
          <div>
            <label class="block text-sm font-medium mb-1">Schedule Type</label>
            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="opt in scheduleOptions"
                :key="opt.value"
                @click="form.scheduleType = opt.value"
                class="px-3 py-2 text-xs rounded-lg border transition-colors text-center"
                :class="form.scheduleType === opt.value
                  ? 'border-makoclaw-accent bg-makoclaw-accent/10 text-makoclaw-accent'
                  : 'border-makoclaw-border bg-makoclaw-bg text-makoclaw-text-secondary hover:border-makoclaw-accent/50'"
              >{{ opt.label }}</button>
            </div>
          </div>

          <!-- Daily: time picker -->
          <div v-if="form.scheduleType === 'daily'" class="space-y-2">
            <label class="block text-sm font-medium">Run at</label>
            <input v-model="form.time" type="time"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent" />
          </div>

          <!-- Weekly: day-of-week + time -->
          <div v-if="form.scheduleType === 'weekly'" class="space-y-3">
            <div>
              <label class="block text-sm font-medium mb-2">Days</label>
              <div class="flex gap-1.5">
                <button
                  v-for="(day, idx) in weekDays"
                  :key="idx"
                  @click="toggleWeekDay(idx)"
                  class="w-9 h-9 text-xs rounded-lg border transition-colors flex items-center justify-center"
                  :class="form.weekDays.includes(idx)
                    ? 'border-makoclaw-accent bg-makoclaw-accent/10 text-makoclaw-accent'
                    : 'border-makoclaw-border bg-makoclaw-bg text-makoclaw-text-secondary hover:border-makoclaw-accent/50'"
                >{{ day }}</button>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Time</label>
              <input v-model="form.time" type="time"
                class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent" />
            </div>
          </div>

          <!-- Monthly: day-of-month + time -->
          <div v-if="form.scheduleType === 'monthly'" class="space-y-3">
            <div>
              <label class="block text-sm font-medium mb-2">Day of month</label>
              <select v-model.number="form.monthDay"
                class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent">
                <option v-for="d in 31" :key="d" :value="d">{{ d }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Time</label>
              <input v-model="form.time" type="time"
                class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent" />
            </div>
          </div>

          <!-- Interval: every N minutes/hours -->
          <div v-if="form.scheduleType === 'interval'" class="space-y-2">
            <label class="block text-sm font-medium">Repeat every</label>
            <div class="flex gap-2">
              <input v-model.number="form.intervalValue" type="number" min="1" placeholder="30"
                class="flex-1 px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent" />
              <select v-model="form.intervalUnit"
                class="px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent">
                <option value="minutes">Minutes</option>
                <option value="hours">Hours</option>
              </select>
            </div>
          </div>

          <!-- One-time: date + time picker -->
          <div v-if="form.scheduleType === 'onetime'" class="space-y-2">
            <label class="block text-sm font-medium">Run at</label>
            <input v-model="form.oneTimeDateTime" type="datetime-local"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent" />
          </div>

          <!-- Custom: raw cron expression -->
          <div v-if="form.scheduleType === 'custom'" class="space-y-2">
            <label class="block text-sm font-medium">Cron Expression</label>
            <input v-model="form.cronExpr" type="text" placeholder="0 9 * * 1-5"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent font-mono" />
            <p class="text-xs text-makoclaw-text-secondary">Standard 5-field cron: minute hour day-of-month month day-of-week</p>
          </div>

          <!-- Timezone (for cron-based schedules) -->
          <div v-if="['daily', 'weekly', 'monthly', 'custom'].includes(form.scheduleType)" class="space-y-2">
            <label class="block text-sm font-medium">Timezone</label>
            <select v-model="form.timezone"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent">
              <option value="">UTC (default)</option>
              <option v-for="tz in commonTimezones" :key="tz" :value="tz">{{ tz }}</option>
            </select>
          </div>

          <!-- Generated expression preview -->
          <div v-if="generatedExpr" class="px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg">
            <p class="text-xs text-makoclaw-text-secondary mb-1">Generated expression</p>
            <code class="text-sm font-mono text-makoclaw-accent">{{ generatedExpr }}</code>
          </div>

          <!-- Next 3 runs preview -->
          <div v-if="nextRuns.length > 0" class="px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg">
            <p class="text-xs text-makoclaw-text-secondary mb-1">Next runs</p>
            <ul class="space-y-0.5">
              <li v-for="(run, i) in nextRuns" :key="i" class="text-sm text-makoclaw-text font-mono">{{ run }}</li>
            </ul>
          </div>

          <!-- Job Type -->
          <div>
            <label class="block text-sm font-medium mb-2">Job Type</label>
            <select v-model="form.job_type"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent">
              <option value="task">🤖 Task (Process through agent)</option>
              <option value="reminder">🔔 Reminder (Direct message)</option>
            </select>
            <p class="mt-1.5 text-xs text-makoclaw-text-secondary">
              <strong>Task:</strong> Agent processes the prompt and sends the result (e.g., "Check weather" → "Today: Sunny, 22°C")<br>
              <strong>Reminder:</strong> Sends the message directly without processing (e.g., "Meeting in 10 min!")
            </p>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm font-medium mb-1">Channel</label>
              <input v-model="form.channel" type="text" placeholder="telegram"
                class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">To (Chat ID)</label>
              <input v-model="form.to" type="text" placeholder=""
                class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent" />
            </div>
          </div>
        </div>

        <!-- Modal Actions -->
        <div class="flex justify-end gap-3 mt-6">
          <button @click="showModal = false"
            class="px-4 py-2 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">Cancel</button>
          <button @click="submitJob" :disabled="!canSubmit"
            class="px-4 py-2 text-sm bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors disabled:opacity-50">
            {{ editingJobId ? 'Save' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
    </Transition>

    <!-- Delete Confirmation Modal -->
    <Transition name="modal">
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-modal p-4" @click.self="showDeleteConfirm = false">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-sm w-full p-6">
        <h3 class="font-semibold text-lg mb-2">Delete Job</h3>
        <p class="text-sm text-makoclaw-text-secondary mb-4">
          Are you sure you want to delete <span class="font-medium text-makoclaw-text">{{ deletingJob?.name }}</span>? This action cannot be undone.
        </p>
        <div class="flex justify-end gap-3">
          <button @click="showDeleteConfirm = false"
            class="px-4 py-2 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors">Cancel</button>
          <button @click="executeDeleteJob"
            class="px-4 py-2 text-sm bg-makoclaw-error text-white rounded-lg hover:bg-makoclaw-error/80 transition-colors">Delete</button>
        </div>
      </div>
    </div>
    </Transition>

    <!-- JSON Editor Modal -->
    <Transition name="modal">
    <div v-if="showJsonModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-modal p-4" @click.self="showJsonModal = false">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-3xl w-full max-h-[90vh] flex flex-col">
        <div class="p-4 border-b border-makoclaw-border flex justify-between items-center">
          <h3 class="font-semibold text-lg">Edit Job as JSON</h3>
          <button
            @click="requestAiJsonFix"
            :disabled="savingJson"
            class="flex items-center gap-2 px-3 py-1.5 text-xs text-makoclaw-accent bg-makoclaw-accent/10 rounded-lg hover:bg-makoclaw-accent/20 transition-colors disabled:opacity-50"
            title="Use AI to fix JSON syntax and validation errors"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
            AI Fix JSON
          </button>
        </div>
        <div class="p-4 flex-1 overflow-hidden flex flex-col">
          <textarea
            v-model="jsonEditContent"
            class="w-full h-full font-mono text-sm px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg resize-none outline-none focus:border-makoclaw-accent min-h-[400px]"
            placeholder="JSON content..."
            spellcheck="false"
          />
          <div v-if="jsonEditError" class="mt-2 p-2 bg-makoclaw-error/10 text-makoclaw-error text-xs rounded border border-makoclaw-error/20">
            {{ jsonEditError }}
          </div>
        </div>
        <div class="p-4 border-t border-makoclaw-border flex justify-end gap-3">
          <button
            @click="showJsonModal = false"
            class="px-4 py-2 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
          >Cancel</button>
          <button
            @click="saveFromJson"
            :disabled="savingJson"
            class="px-4 py-2 text-sm bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/80 transition-colors disabled:opacity-50"
          >
            {{ savingJson ? 'Saving...' : 'Save Changes' }}
          </button>
        </div>
      </div>
    </div>
    </Transition>

    <!-- AI Cron Creator Modal -->
    <Transition name="modal">
    <div v-if="showAiModal" class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-modal p-4" @click.self="closeAiModal">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="p-4 border-b border-makoclaw-border">
          <h3 class="font-semibold text-lg">Create Cron Job with AI</h3>
          <p class="text-xs text-makoclaw-text-secondary mt-1">Describe what you want in plain language</p>
        </div>
        <div class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium mb-2">What should this cron job do?</label>
            <textarea
              v-model="aiPrompt"
              rows="3"
              placeholder="e.g., Send me weather every day at 8am, or Backup my tasks every 2 hours"
              class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-sm outline-none focus:border-makoclaw-accent resize-none"
              :disabled="aiGenerating"
            />
          </div>
          <button
            @click="generateCronWithAI"
            :disabled="!aiPrompt.trim() || aiGenerating"
            class="w-full flex items-center justify-center gap-2 px-4 py-2 bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <svg v-if="aiGenerating" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
            {{ aiGenerating ? 'Generating...' : 'Generate Job' }}
          </button>

          <!-- AI Result Preview -->
          <div v-if="aiResult" class="space-y-3 pt-4 border-t border-makoclaw-border">
            <div class="p-3 bg-makoclaw-accent/10 border border-makoclaw-accent/20 rounded-lg">
              <p class="text-sm text-makoclaw-accent">{{ aiExplanation }}</p>
            </div>
            <div class="p-4 bg-makoclaw-bg border border-makoclaw-border rounded-lg space-y-2">
              <div><span class="text-sm font-medium">Name:</span> <span class="text-sm text-makoclaw-text-secondary">{{ aiResult.name }}</span></div>
              <div><span class="text-sm font-medium">Message:</span> <span class="text-sm text-makoclaw-text-secondary">{{ aiResult.message || aiResult.payload?.message }}</span></div>
              <div><span class="text-sm font-medium">Schedule:</span> <span class="text-sm text-makoclaw-text-secondary">{{ formatSchedule(aiResult.schedule) }}</span></div>
            </div>
            <div class="flex gap-2">
              <button
                @click="editAiResult"
                class="flex-1 flex items-center justify-center gap-2 px-4 py-2 border border-makoclaw-border rounded-lg hover:bg-makoclaw-bg transition-colors text-sm"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                </svg>
                Edit Before Saving
              </button>
              <button
                @click="saveAiResult"
                class="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/80 transition-colors text-sm"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
                Create Job
              </button>
            </div>
          </div>
        </div>
        <div class="p-4 border-t border-makoclaw-border flex justify-end">
          <button
            @click="closeAiModal"
            class="px-4 py-2 text-sm text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
          >
            {{ aiResult ? 'Cancel' : 'Close' }}
          </button>
        </div>
      </div>
    </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const loading = ref(true)
const jobs = ref([])
const status = ref({ enabled: false })
const showModal = ref(false)
const editingJobId = ref(null)
const showDeleteConfirm = ref(false)
const deletingJob = ref(null)

// JSON Editor state
const showJsonModal = ref(false)
const jsonEditContent = ref('')
const jsonEditError = ref(null)
const savingJson = ref(false)

// AI Cron Creator state
const showAiModal = ref(false)
const aiPrompt = ref('')
const aiGenerating = ref(false)
const aiResult = ref(null)
const aiExplanation = ref('')

const weekDays = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

const scheduleOptions = [
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
  { value: 'interval', label: 'Interval' },
  { value: 'onetime', label: 'One-time' },
  { value: 'custom', label: 'Custom' }
]

const commonTimezones = [
  'America/New_York', 'America/Chicago', 'America/Denver', 'America/Los_Angeles',
  'America/Sao_Paulo', 'America/Mexico_City', 'America/Argentina/Buenos_Aires',
  'Europe/London', 'Europe/Paris', 'Europe/Berlin', 'Europe/Madrid', 'Europe/Moscow',
  'Asia/Shanghai', 'Asia/Tokyo', 'Asia/Seoul', 'Asia/Kolkata', 'Asia/Singapore',
  'Australia/Sydney', 'Pacific/Auckland'
]

const statusLabel = computed(() => {
  if (status.value.enabled) return 'Running'
  if (status.value.reason === 'not_initialized') return 'Not available for this user'
  return 'Not available'
})

const defaultForm = () => ({
  name: '',
  message: '',
  scheduleType: 'daily',
  time: '09:00',
  weekDays: [0], // Monday
  monthDay: 1,
  intervalValue: 30,
  intervalUnit: 'minutes',
  oneTimeDateTime: '',
  cronExpr: '',
  timezone: '',
  job_type: 'task', // Default to agent processing
  channel: '',
  to: ''
})

const form = ref(defaultForm())

const toggleWeekDay = (idx) => {
  const arr = form.value.weekDays
  const pos = arr.indexOf(idx)
  if (pos >= 0) {
    if (arr.length > 1) arr.splice(pos, 1) // keep at least one selected
  } else {
    arr.push(idx)
    arr.sort()
  }
}

// Build cron expression from the visual form
const generatedExpr = computed(() => {
  const f = form.value
  if (f.scheduleType === 'interval' || f.scheduleType === 'onetime') return ''

  const [hh, mm] = (f.time || '09:00').split(':').map(Number)

  if (f.scheduleType === 'daily') {
    return `${mm} ${hh} * * *`
  }
  if (f.scheduleType === 'weekly') {
    // Cron days: 0=Sun,1=Mon..6=Sat  Our array: 0=Mon..6=Sun
    const cronDays = f.weekDays.map(d => (d + 1) % 7).sort().join(',')
    return `${mm} ${hh} * * ${cronDays}`
  }
  if (f.scheduleType === 'monthly') {
    return `${mm} ${hh} ${f.monthDay} * *`
  }
  if (f.scheduleType === 'custom') {
    return f.cronExpr || ''
  }
  return ''
})

// Compute next 3 runs (client-side approximation)
const nextRuns = computed(() => {
  const f = form.value

  if (f.scheduleType === 'onetime') {
    if (!f.oneTimeDateTime) return []
    const d = new Date(f.oneTimeDateTime)
    if (isNaN(d.getTime())) return []
    return [d.toLocaleString()]
  }

  if (f.scheduleType === 'interval') {
    const ms = f.intervalUnit === 'hours' ? f.intervalValue * 3600000 : f.intervalValue * 60000
    if (!ms || ms <= 0) return []
    const now = Date.now()
    return [1, 2, 3].map(i => new Date(now + ms * i).toLocaleString())
  }

  // For cron-based schedules, compute from the expression
  const expr = generatedExpr.value
  if (!expr || expr.trim().split(/\s+/).length !== 5) return []

  try {
    const runs = getNextCronRuns(expr, 3, f.timezone)
    return runs.map(d => d.toLocaleString())
  } catch {
    return []
  }
})

const canSubmit = computed(() => {
  const f = form.value
  if (!f.name.trim() || !f.message.trim()) return false
  if (f.scheduleType === 'interval') return f.intervalValue > 0
  if (f.scheduleType === 'onetime') return !!f.oneTimeDateTime
  if (f.scheduleType === 'custom') return !!f.cronExpr.trim()
  return true
})

// Build the API payload from the form
function buildPayload() {
  const f = form.value
  const payload = {
    name: f.name.trim(),
    message: f.message.trim(),
    job_type: f.job_type || 'task', // Default to task
    channel: f.channel,
    to: f.to,
    schedule: {}
  }

  if (f.scheduleType === 'interval') {
    const ms = f.intervalUnit === 'hours' ? f.intervalValue * 3600000 : f.intervalValue * 60000
    payload.schedule = { kind: 'every', everyMs: ms }
  } else if (f.scheduleType === 'onetime') {
    const ts = new Date(f.oneTimeDateTime).getTime()
    payload.schedule = { kind: 'at', atMs: ts }
  } else {
    // daily, weekly, monthly, custom → all produce a cron expression
    payload.schedule = {
      kind: 'cron',
      expr: generatedExpr.value,
      tz: f.timezone || undefined
    }
  }

  return payload
}

// Reverse-parse a job into form fields for editing
function jobToForm(job) {
  const f = defaultForm()
  f.name = job.name
  f.message = job.payload.message
  // Migrate old format if needed
  if (job.payload.job_type) {
    f.job_type = job.payload.job_type
  } else if (job.payload.deliver !== undefined) {
    f.job_type = job.payload.deliver ? 'reminder' : 'task'
  } else {
    f.job_type = 'task' // Default
  }
  f.channel = job.payload.channel || ''
  f.to = job.payload.to || ''
  f.timezone = job.schedule.tz || ''

  if (job.schedule.kind === 'every') {
    f.scheduleType = 'interval'
    const totalMs = job.schedule.everyMs || 0
    if (totalMs >= 3600000 && totalMs % 3600000 === 0) {
      f.intervalValue = totalMs / 3600000
      f.intervalUnit = 'hours'
    } else {
      f.intervalValue = Math.round(totalMs / 60000) || 1
      f.intervalUnit = 'minutes'
    }
  } else if (job.schedule.kind === 'at') {
    f.scheduleType = 'onetime'
    if (job.schedule.atMs) {
      const d = new Date(job.schedule.atMs)
      // Format as YYYY-MM-DDTHH:MM for datetime-local input
      const pad = n => String(n).padStart(2, '0')
      f.oneTimeDateTime = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
    }
  } else if (job.schedule.kind === 'cron' && job.schedule.expr) {
    const parsed = parseCronExpr(job.schedule.expr)
    if (parsed) {
      f.scheduleType = parsed.type
      f.time = parsed.time
      if (parsed.type === 'weekly') f.weekDays = parsed.weekDays
      if (parsed.type === 'monthly') f.monthDay = parsed.monthDay
    } else {
      f.scheduleType = 'custom'
      f.cronExpr = job.schedule.expr
    }
  }

  return f
}

// Try to detect if a cron expression matches a simple daily/weekly/monthly pattern
function parseCronExpr(expr) {
  const parts = expr.trim().split(/\s+/)
  if (parts.length !== 5) return null

  const [min, hr, dom, mon, dow] = parts
  const mm = parseInt(min, 10)
  const hh = parseInt(hr, 10)
  if (isNaN(mm) || isNaN(hh)) return null

  const pad = n => String(n).padStart(2, '0')
  const time = `${pad(hh)}:${pad(mm)}`

  // Daily: M H * * *
  if (dom === '*' && mon === '*' && dow === '*') {
    return { type: 'daily', time }
  }

  // Weekly: M H * * 0,1,5  (comma-separated days)
  if (dom === '*' && mon === '*' && dow !== '*') {
    const cronDays = dow.split(',').map(Number).filter(n => !isNaN(n))
    if (cronDays.length > 0) {
      // Convert from cron days (0=Sun..6=Sat) to our index (0=Mon..6=Sun)
      const weekDays = cronDays.map(d => d === 0 ? 6 : d - 1).sort()
      return { type: 'weekly', time, weekDays }
    }
  }

  // Monthly: M H D * *
  if (mon === '*' && dow === '*') {
    const d = parseInt(dom, 10)
    if (!isNaN(d) && d >= 1 && d <= 31) {
      return { type: 'monthly', time, monthDay: d }
    }
  }

  return null
}

// Simple cron next-run calculator (client-side approximation for preview)
function getNextCronRuns(expr, count, tz) {
  const parts = expr.trim().split(/\s+/)
  if (parts.length !== 5) return []

  const runs = []
  const now = new Date()
  let candidate = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours(), now.getMinutes() + 1, 0, 0)

  const maxIterations = 525960 // ~1 year of minutes
  for (let i = 0; i < maxIterations && runs.length < count; i++) {
    if (matchesCron(parts, candidate)) {
      runs.push(new Date(candidate))
    }
    candidate = new Date(candidate.getTime() + 60000)
  }
  return runs
}

function matchesCron(parts, date) {
  const [minP, hrP, domP, monP, dowP] = parts
  const min = date.getMinutes()
  const hr = date.getHours()
  const dom = date.getDate()
  const mon = date.getMonth() + 1
  const dow = date.getDay() // 0=Sun

  return matchField(minP, min, 0, 59)
    && matchField(hrP, hr, 0, 23)
    && matchField(domP, dom, 1, 31)
    && matchField(monP, mon, 1, 12)
    && matchField(dowP, dow, 0, 6)
}

function matchField(field, value, min, max) {
  if (field === '*') return true
  // Handle */n
  if (field.startsWith('*/')) {
    const step = parseInt(field.substring(2), 10)
    return !isNaN(step) && step > 0 && value % step === 0
  }
  // Handle ranges like 1-5
  if (field.includes('-') && !field.includes(',')) {
    const [lo, hi] = field.split('-').map(Number)
    return value >= lo && value <= hi
  }
  // Handle comma-separated values (possibly with ranges)
  const vals = new Set()
  for (const part of field.split(',')) {
    if (part.includes('-')) {
      const [lo, hi] = part.split('-').map(Number)
      for (let v = lo; v <= hi; v++) vals.add(v)
    } else {
      vals.add(parseInt(part, 10))
    }
  }
  return vals.has(value)
}

function getJobTypeDisplay(payload) {
  // Migrate old format if needed
  let jobType = payload.job_type
  if (!jobType && payload.deliver !== undefined) {
    jobType = payload.deliver ? 'reminder' : 'task'
  }
  if (!jobType) {
    jobType = 'task' // Default
  }
  return jobType === 'reminder' ? '🔔 Reminder' : '🤖 Task'
}

// ---- Actions ----

const loadJobs = async () => {
  loading.value = true
  try {
    const data = await advancedService.fetchCronJobs()
    jobs.value = data.jobs || []
    status.value = data.status || { enabled: false }
  } catch (err) {
    toast.error('Failed to load cron jobs')
    console.error('Failed to load cron jobs:', err)
  } finally {
    loading.value = false
  }
}

const openCreateModal = () => {
  editingJobId.value = null
  form.value = defaultForm()
  showModal.value = true
}

const openEditModal = (job) => {
  editingJobId.value = job.id
  form.value = jobToForm(job)
  showModal.value = true
}

const submitJob = async () => {
  const payload = buildPayload()
  try {
    if (editingJobId.value) {
      await advancedService.updateCronJob(editingJobId.value, payload)
      toast.success('Job updated')
    } else {
      await advancedService.createCronJob(payload)
      toast.success('Job created')
    }
    showModal.value = false
    await loadJobs()
  } catch (err) {
    const msg = err?.response?.data || err.message || 'Unknown error'
    toast.error(`Failed to ${editingJobId.value ? 'update' : 'create'} job: ${msg}`)
  }
}

const toggleJob = async (id, enabled) => {
  try {
    await advancedService.toggleCronJob(id, enabled)
    toast.success(enabled ? 'Job enabled' : 'Job disabled')
    await loadJobs()
  } catch (err) {
    toast.error('Failed to toggle job')
  }
}

const runJob = async (job) => {
  try {
    await advancedService.runCronJob(job.id)
    toast.success(`Job '${job.name}' triggered`)
    await loadJobs()
  } catch (err) {
    const msg = err?.response?.data || err.message || 'Unknown error'
    toast.error(`Failed to run job: ${msg}`)
  }
}

const confirmDeleteJob = (job) => {
  deletingJob.value = job
  showDeleteConfirm.value = true
}

const executeDeleteJob = async () => {
  if (!deletingJob.value) return
  try {
    await advancedService.deleteCronJob(deletingJob.value.id)
    toast.success('Job deleted')
    showDeleteConfirm.value = false
    deletingJob.value = null
    await loadJobs()
  } catch (err) {
    toast.error('Failed to delete job')
  }
}

// JSON Editor methods
const openJsonEditor = () => {
  const job = jobs.value.find(j => j.id === editingJobId.value)
  if (!job) {
    toast.error('Job not found')
    return
  }
  jsonEditContent.value = JSON.stringify(job, null, 2)
  jsonEditError.value = null
  showModal.value = false
  showJsonModal.value = true
}

const saveFromJson = async () => {
  try {
    const parsed = JSON.parse(jsonEditContent.value)

    // Validate required fields
    if (!parsed.name || !parsed.schedule) {
      jsonEditError.value = 'Invalid job structure: name and schedule are required'
      return
    }

    // Validate payload structure
    const payload = parsed.payload || {}
    if (!payload.message && !parsed.message) {
      jsonEditError.value = 'Invalid job structure: message is required'
      return
    }

    savingJson.value = true
    jsonEditError.value = null

    // Construct API payload
    const apiPayload = {
      name: parsed.name,
      message: payload.message || parsed.message,
      job_type: payload.job_type || parsed.job_type || 'task',
      channel: payload.channel || parsed.channel || '',
      to: payload.to || parsed.to || '',
      schedule: parsed.schedule
    }

    await advancedService.updateCronJob(editingJobId.value, apiPayload)
    toast.success('Job updated from JSON')
    showJsonModal.value = false
    await loadJobs()
  } catch (err) {
    if (err instanceof SyntaxError) {
      jsonEditError.value = `JSON Parse Error: ${err.message}`
    } else {
      jsonEditError.value = `Failed to save: ${err.response?.data?.error || err.message}`
    }
  } finally {
    savingJson.value = false
  }
}

const requestAiJsonFix = async () => {
  if (!jsonEditContent.value.trim()) {
    toast.error('No JSON content to fix')
    return
  }

  savingJson.value = true
  try {
    const result = await advancedService.fixJsonWithAI(jsonEditContent.value, 'cron_job')
    jsonEditContent.value = result.fixed_json
    jsonEditError.value = null
    toast.success(result.changes || 'AI validated JSON - no changes needed')
  } catch (err) {
    toast.error(err?.response?.data?.error || 'AI fix failed')
  } finally {
    savingJson.value = false
  }
}

// AI Cron Creator methods
const openAiCronModal = () => {
  aiPrompt.value = ''
  aiResult.value = null
  aiExplanation.value = ''
  aiGenerating.value = false
  showAiModal.value = true
}

const closeAiModal = () => {
  showAiModal.value = false
  aiPrompt.value = ''
  aiResult.value = null
  aiExplanation.value = ''
}

const generateCronWithAI = async () => {
  if (!aiPrompt.value.trim()) {
    toast.error('Please enter a description')
    return
  }

  aiGenerating.value = true
  try {
    const result = await advancedService.createCronWithAI(aiPrompt.value.trim())
    aiResult.value = result.job
    aiExplanation.value = result.explanation
    toast.success('Cron job generated successfully')
  } catch (err) {
    toast.error(err?.response?.data?.error || 'AI generation failed')
  } finally {
    aiGenerating.value = false
  }
}

const editAiResult = () => {
  if (!aiResult.value) return

  // Populate the form with AI-generated values
  form.name = aiResult.value.name || ''
  form.message = aiResult.value.message || aiResult.value.payload?.message || ''
  form.job_type = aiResult.value.job_type || 'task'
  form.channel = aiResult.value.channel || ''
  form.to = aiResult.value.to || ''

  // Parse schedule
  const schedule = aiResult.value.schedule
  if (schedule) {
    if (schedule.kind === 'cron' && schedule.expr) {
      const parsed = parseCronExpr(schedule.expr)
      if (parsed) {
        if (parsed.type === 'daily') {
          form.scheduleType = 'daily'
          form.time = parsed.time
        } else if (parsed.type === 'weekly') {
          form.scheduleType = 'weekly'
          form.weekDays = parsed.weekDays
          form.time = parsed.time
        } else {
          form.scheduleType = 'custom'
          form.cronExpr = schedule.expr
        }
      }
    } else if (schedule.kind === 'every' && schedule.everyMs) {
      form.scheduleType = 'interval'
      form.intervalMinutes = Math.round(schedule.everyMs / 60000)
    }

    if (schedule.tz) {
      form.timezone = schedule.tz
    }
  }

  // Close AI modal and open edit modal
  showAiModal.value = false
  editingJobId.value = null
  showModal.value = true
}

const saveAiResult = async () => {
  if (!aiResult.value) return

  try {
    const payload = {
      name: aiResult.value.name,
      message: aiResult.value.message || aiResult.value.payload?.message,
      job_type: aiResult.value.job_type || 'task',
      channel: aiResult.value.channel || '',
      to: aiResult.value.to || '',
      schedule: aiResult.value.schedule
    }

    await advancedService.createCronJob(payload)
    toast.success('AI cron job created')
    showAiModal.value = false
    await loadJobs()
  } catch (err) {
    toast.error(err?.response?.data?.error || 'Failed to create job')
  }
}

const formatSchedule = (schedule) => {
  if (schedule.kind === 'every' && schedule.everyMs) {
    const mins = Math.round(schedule.everyMs / 60000)
    if (mins < 60) return `every ${mins}m`
    const hrs = Math.round(mins / 60)
    return hrs === 1 ? 'every hour' : `every ${hrs}h`
  }
  if (schedule.kind === 'cron' && schedule.expr) {
    const parsed = parseCronExpr(schedule.expr)
    if (parsed) {
      if (parsed.type === 'daily') return `daily at ${parsed.time}`
      if (parsed.type === 'weekly') {
        const days = parsed.weekDays.map(d => weekDays[d]).join(', ')
        return `${days} at ${parsed.time}`
      }
      if (parsed.type === 'monthly') return `monthly on day ${parsed.monthDay} at ${parsed.time}`
    }
    return schedule.expr
  }
  if (schedule.kind === 'at' && schedule.atMs) return `once at ${new Date(schedule.atMs).toLocaleString()}`
  return schedule.kind
}

const formatTimestamp = (ms) => {
  if (!ms) return ''
  return new Date(ms).toLocaleString()
}

onMounted(() => loadJobs())
</script>



