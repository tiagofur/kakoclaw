<template>
  <div class="flex flex-col h-full bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-cyan-500/30 via-transparent to-transparent"></div>
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-blue-500/20 via-transparent to-transparent"></div>
    </div>

    <!-- Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <!-- Title Row -->
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-cyan-500/20 to-blue-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-cyan-500/10">
            <svg class="w-5 h-5 sm:w-6 sm:h-6 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-cyan-400 bg-clip-text text-transparent">Cron Jobs</h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5 hidden sm:block">Schedule automated tasks and reminders</p>
          </div>
          <div class="flex items-center gap-2">
            <button
              @click="openAiCronModal"
              class="px-3 sm:px-4 py-2.5 min-h-[40px] bg-makoclaw-surface/50 border border-makoclaw-border/50 text-makoclaw-text rounded-xl hover:bg-makoclaw-surface-hover hover:border-cyan-500/30 transition-all text-sm font-medium flex items-center gap-2 backdrop-blur-sm"
            >
              <svg class="w-4 h-4 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
              </svg>
              <span class="hidden sm:inline">Create with AI</span>
            </button>
            <button
              @click="openCreateModal"
              class="px-4 sm:px-5 py-2.5 min-h-[40px] bg-gradient-to-r from-cyan-500 to-cyan-600 hover:from-cyan-600 hover:to-cyan-700 text-white rounded-xl transition-all shadow-lg shadow-cyan-500/25 hover:shadow-cyan-500/40 text-sm font-bold flex items-center gap-2 active:scale-95 flex-shrink-0"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              <span class="hidden sm:inline">New Job</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Status Banner -->
      <div class="px-4 sm:px-6 pb-3 sm:pb-4">
        <div
          v-if="!loading"
          class="flex items-center gap-3 px-4 py-2.5 rounded-xl border transition-all backdrop-blur-sm"
          :class="status.enabled
            ? 'bg-cyan-500/5 border-cyan-500/20 text-cyan-400'
            : 'bg-amber-500/5 border-amber-500/20 text-amber-400'"
        >
          <div class="flex items-center gap-2">
            <div class="w-2 h-2 rounded-full animate-pulse" :class="status.enabled ? 'bg-cyan-400' : 'bg-amber-400'"></div>
            <span class="text-sm font-medium">{{ statusLabel }}</span>
          </div>
          <span v-if="status.jobs !== undefined" class="text-xs opacity-75">{{ status.jobs }} active job{{ status.jobs !== 1 ? 's' : '' }}</span>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-4 sm:p-6 custom-scrollbar relative">
      <!-- Loading State -->
      <div v-if="loading" class="space-y-3">
        <div v-for="i in 3" :key="i" class="cron-card-skeleton">
          <div class="flex items-start gap-3">
            <div class="skeleton w-10 h-10 rounded-xl"></div>
            <div class="flex-1 space-y-2">
              <div class="skeleton h-4 w-40 rounded"></div>
              <div class="skeleton h-3 w-full rounded"></div>
            </div>
          </div>
          <div class="flex gap-3 mt-4">
            <div class="skeleton h-6 w-24 rounded-lg"></div>
            <div class="skeleton h-6 w-20 rounded-lg"></div>
          </div>
        </div>
      </div>

      <template v-else>
        <!-- Empty State -->
        <div v-if="jobs.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
          <div class="relative">
            <div class="absolute inset-0 bg-gradient-to-br from-cyan-500/30 to-blue-500/30 rounded-3xl blur-2xl opacity-50"></div>
            <div class="relative glass-panel p-8 rounded-2xl shadow-2xl ring-1 ring-white/10">
              <div class="w-16 h-16 mx-auto rounded-2xl bg-gradient-to-br from-cyan-500/20 to-blue-500/20 flex items-center justify-center ring-1 ring-white/20">
                <svg class="w-8 h-8 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
            </div>
          </div>
          <h3 class="text-lg font-bold text-makoclaw-text mt-6">No scheduled jobs</h3>
          <p class="text-sm text-makoclaw-text-secondary/70 mt-2 max-w-xs">Create cron jobs to automate tasks and send scheduled reminders.</p>
          <button @click="openCreateModal" class="mt-6 px-5 py-2.5 bg-gradient-to-r from-cyan-500 to-cyan-600 hover:from-cyan-600 hover:to-cyan-700 text-white rounded-xl font-bold shadow-lg shadow-cyan-500/25 transition-all active:scale-95 flex items-center gap-2">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            Create Your First Job
          </button>
        </div>

        <!-- Jobs List -->
        <div v-else class="space-y-3">
          <div
            v-for="job in jobs"
            :key="job.id"
            class="cron-card group"
          >
            <!-- Card Header -->
            <div class="flex items-start gap-3 mb-3">
              <div class="cron-icon" :class="job.enabled ? 'cron-icon-active' : 'cron-icon-disabled'">
                <svg v-if="getJobTypeDisplay(job.payload) === 'task'" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
                <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                </svg>
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-1 flex-wrap">
                  <h3 class="font-semibold text-makoclaw-text">{{ job.name }}</h3>
                  <span
                    class="px-2 py-0.5 text-[10px] font-medium rounded-full"
                    :class="job.enabled ? 'bg-cyan-500/10 text-cyan-400' : 'bg-makoclaw-text-secondary/10 text-makoclaw-text-secondary'"
                  >
                    {{ job.enabled ? 'Active' : 'Disabled' }}
                  </span>
                  <span class="px-2 py-0.5 text-[10px] font-medium rounded-full bg-makoclaw-bg/50 text-makoclaw-text-secondary">
                    {{ getJobTypeDisplay(job.payload) === 'task' ? 'Task' : 'Reminder' }}
                  </span>
                </div>
                <p class="text-sm text-makoclaw-text-secondary line-clamp-2">{{ job.payload.message }}</p>
              </div>
            </div>

            <!-- Schedule Info -->
            <div class="flex flex-wrap items-center gap-3 text-xs text-makoclaw-text-secondary mb-4">
              <div class="flex items-center gap-1.5 px-2.5 py-1 bg-makoclaw-bg/50 rounded-lg">
                <svg class="w-3.5 h-3.5 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span class="font-mono">{{ formatSchedule(job.schedule) }}</span>
              </div>
              <div v-if="job.schedule.tz" class="flex items-center gap-1.5 px-2.5 py-1 bg-makoclaw-bg/50 rounded-lg">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span class="font-mono">{{ job.schedule.tz }}</span>
              </div>
              <div v-if="job.state.lastStatus" class="flex items-center gap-1.5">
                <span class="text-makoclaw-text-secondary/60">Last:</span>
                <span :class="job.state.lastStatus === 'ok' ? 'text-emerald-400' : 'text-amber-400'">{{ job.state.lastStatus }}</span>
              </div>
              <div v-if="job.state.nextRunAtMs" class="flex items-center gap-1.5">
                <span class="text-makoclaw-text-secondary/60">Next:</span>
                <span>{{ formatTimestamp(job.state.nextRunAtMs) }}</span>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex items-center gap-2 pt-3 border-t border-makoclaw-border/30">
              <button
                @click="runJob(job)"
                class="cron-action-btn text-cyan-400"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Run Now
              </button>
              <button
                @click="openEditModal(job)"
                class="cron-action-btn"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
                Edit
              </button>
              <button
                @click="toggleJob(job.id, !job.enabled)"
                class="cron-action-btn"
                :class="job.enabled ? 'text-amber-400' : 'text-emerald-400'"
              >
                <svg v-if="job.enabled" class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <svg v-else class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                {{ job.enabled ? 'Disable' : 'Enable' }}
              </button>
              <button
                @click="confirmDeleteJob(job)"
                class="cron-action-btn text-makoclaw-error ml-auto opacity-0 group-hover:opacity-100"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Create/Edit Modal -->
    <Transition name="modal">
      <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
        <div class="modal-content max-w-lg max-h-[90vh]">
          <div class="modal-header">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-cyan-500/20 to-makoclaw-accent/20 flex items-center justify-center">
                <svg class="w-5 h-5 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <div>
                <h2 class="font-bold text-lg text-makoclaw-text">{{ editingJobId ? 'Edit Job' : 'Create Job' }}</h2>
                <p class="text-xs text-makoclaw-text-secondary">Schedule automated tasks</p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <button
                v-if="editingJobId"
                @click="openJsonEditor"
                class="text-xs px-3 py-1.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-lg hover:bg-makoclaw-surface transition-colors flex items-center gap-1.5"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                </svg>
                JSON
              </button>
              <button @click="showModal = false" class="modal-close-btn">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>

          <div class="modal-body custom-scrollbar space-y-5">
            <!-- Name -->
            <div>
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Name</label>
              <input v-model="form.name" type="text" placeholder="My scheduled task" class="input-field" />
            </div>

            <!-- Message -->
            <div>
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Message</label>
              <textarea v-model="form.message" rows="3" placeholder="What should the agent do..." class="input-field resize-none" />
            </div>

            <!-- Job Type -->
            <div>
              <label class="block text-sm font-medium text-makoclaw-text mb-2">Job Type</label>
              <div class="grid grid-cols-2 gap-2">
                <label
                  class="job-type-option"
                  :class="form.job_type === 'task' ? 'job-type-option-active' : ''"
                >
                  <input type="radio" v-model="form.job_type" value="task" class="sr-only" />
                  <svg class="w-5 h-5 mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                  <span class="text-sm font-medium">Task</span>
                  <span class="text-[10px] text-makoclaw-text-secondary">Agent processes it</span>
                </label>
                <label
                  class="job-type-option"
                  :class="form.job_type === 'reminder' ? 'job-type-option-active' : ''"
                >
                  <input type="radio" v-model="form.job_type" value="reminder" class="sr-only" />
                  <svg class="w-5 h-5 mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                  </svg>
                  <span class="text-sm font-medium">Reminder</span>
                  <span class="text-[10px] text-makoclaw-text-secondary">Direct message</span>
                </label>
              </div>
            </div>

            <!-- Schedule Type -->
            <div>
              <label class="block text-sm font-medium text-makoclaw-text mb-2">Schedule</label>
              <div class="grid grid-cols-3 sm:grid-cols-6 gap-2">
                <button
                  v-for="opt in scheduleOptions"
                  :key="opt.value"
                  @click="form.scheduleType = opt.value"
                  class="schedule-type-btn"
                  :class="form.scheduleType === opt.value ? 'schedule-type-btn-active' : ''"
                >{{ opt.label }}</button>
              </div>
            </div>

            <!-- Schedule-specific inputs -->
            <div v-if="form.scheduleType === 'daily'" class="space-y-3">
              <div>
                <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Time</label>
                <input v-model="form.time" type="time" class="input-field" />
              </div>
            </div>

            <div v-if="form.scheduleType === 'weekly'" class="space-y-3">
              <div>
                <label class="block text-sm font-medium text-makoclaw-text mb-2">Days</label>
                <div class="flex gap-1.5">
                  <button
                    v-for="(day, idx) in weekDays"
                    :key="idx"
                    @click="toggleWeekDay(idx)"
                    class="day-btn"
                    :class="form.weekDays.includes(idx) ? 'day-btn-active' : ''"
                  >{{ day }}</button>
                </div>
              </div>
              <div>
                <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Time</label>
                <input v-model="form.time" type="time" class="input-field" />
              </div>
            </div>

            <div v-if="form.scheduleType === 'monthly'" class="space-y-3">
              <div>
                <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Day of month</label>
                <select v-model.number="form.monthDay" class="input-field">
                  <option v-for="d in 31" :key="d" :value="d">{{ d }}</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Time</label>
                <input v-model="form.time" type="time" class="input-field" />
              </div>
            </div>

            <div v-if="form.scheduleType === 'interval'" class="space-y-3">
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Repeat every</label>
              <div class="flex gap-2">
                <input v-model.number="form.intervalValue" type="number" min="1" placeholder="30" class="input-field flex-1" />
                <select v-model="form.intervalUnit" class="input-field w-32">
                  <option value="minutes">Minutes</option>
                  <option value="hours">Hours</option>
                </select>
              </div>
            </div>

            <div v-if="form.scheduleType === 'onetime'" class="space-y-3">
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Run at</label>
              <input v-model="form.oneTimeDateTime" type="datetime-local" class="input-field" />
            </div>

            <div v-if="form.scheduleType === 'custom'" class="space-y-3">
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Cron Expression</label>
              <input v-model="form.cronExpr" type="text" placeholder="0 9 * * 1-5" class="input-field font-mono" />
              <p class="text-xs text-makoclaw-text-secondary">minute hour day-of-month month day-of-week</p>
            </div>

            <!-- Timezone -->
            <div v-if="['daily', 'weekly', 'monthly', 'custom'].includes(form.scheduleType)">
              <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Timezone</label>
              <select v-model="form.timezone" class="input-field">
                <option value="">UTC (default)</option>
                <option v-for="tz in commonTimezones" :key="tz" :value="tz">{{ tz }}</option>
              </select>
            </div>

            <!-- Preview -->
            <div v-if="generatedExpr" class="p-3 bg-makoclaw-bg/50 rounded-xl border border-makoclaw-border/30">
              <p class="text-xs text-makoclaw-text-secondary mb-1">Expression</p>
              <code class="text-sm font-mono text-cyan-400">{{ generatedExpr }}</code>
            </div>

            <div v-if="nextRuns.length > 0" class="p-3 bg-makoclaw-bg/50 rounded-xl border border-makoclaw-border/30">
              <p class="text-xs text-makoclaw-text-secondary mb-2">Next runs</p>
              <ul class="space-y-1">
                <li v-for="(run, i) in nextRuns" :key="i" class="text-xs font-mono text-makoclaw-text">{{ run }}</li>
              </ul>
            </div>

            <!-- Channel -->
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-sm font-medium text-makoclaw-text mb-1.5">Channel</label>
                <input v-model="form.channel" type="text" placeholder="telegram" class="input-field" />
              </div>
              <div>
                <label class="block text-sm font-medium text-makoclaw-text mb-1.5">To (Chat ID)</label>
                <input v-model="form.to" type="text" placeholder="" class="input-field" />
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="showModal = false" class="btn-ghost">Cancel</button>
            <button @click="submitJob" :disabled="!canSubmit" class="btn-primary">
              {{ editingJobId ? 'Save' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Delete Confirmation -->
    <Transition name="modal">
      <div v-if="showDeleteConfirm" class="modal-backdrop" @click.self="showDeleteConfirm = false">
        <div class="modal-content max-w-sm">
          <div class="p-6 text-center">
            <div class="w-16 h-16 rounded-2xl bg-makoclaw-error/10 flex items-center justify-center mx-auto mb-4">
              <svg class="w-8 h-8 text-makoclaw-error" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </div>
            <h3 class="font-bold text-lg text-makoclaw-text mb-2">Delete Job</h3>
            <p class="text-sm text-makoclaw-text-secondary mb-6">
              Are you sure you want to delete <span class="font-medium text-makoclaw-text">{{ deletingJob?.name }}</span>?
            </p>
            <div class="flex justify-center gap-3">
              <button @click="showDeleteConfirm = false" class="btn-ghost">Cancel</button>
              <button @click="executeDeleteJob" class="btn-danger">Delete</button>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- JSON Editor Modal -->
    <Transition name="modal">
      <div v-if="showJsonModal" class="modal-backdrop" @click.self="showJsonModal = false">
        <div class="modal-content max-w-3xl h-[80vh]">
          <div class="modal-header">
            <div class="flex items-center gap-3">
              <h2 class="font-bold text-lg text-makoclaw-text">Edit as JSON</h2>
            </div>
            <div class="flex items-center gap-2">
              <button
                @click="requestAiJsonFix"
                :disabled="savingJson"
                class="text-xs px-3 py-1.5 bg-cyan-500/10 text-cyan-400 rounded-lg hover:bg-cyan-500/20 transition-colors flex items-center gap-1.5 disabled:opacity-50"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                </svg>
                AI Fix
              </button>
              <button @click="showJsonModal = false" class="modal-close-btn">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
          <div class="flex-1 overflow-hidden p-4">
            <textarea
              v-model="jsonEditContent"
              class="w-full h-full input-field font-mono text-sm resize-none"
              placeholder="JSON content..."
              spellcheck="false"
            ></textarea>
            <div v-if="jsonEditError" class="mt-2 p-3 bg-makoclaw-error/10 border border-makoclaw-error/20 rounded-lg text-xs text-makoclaw-error">
              {{ jsonEditError }}
            </div>
          </div>
          <div class="modal-footer">
            <button @click="showJsonModal = false" class="btn-ghost">Cancel</button>
            <button @click="saveFromJson" :disabled="savingJson" class="btn-primary">
              {{ savingJson ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- AI Cron Creator Modal -->
    <Transition name="modal">
      <div v-if="showAiModal" class="modal-backdrop" @click.self="closeAiModal">
        <div class="modal-content max-w-2xl">
          <div class="modal-header">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-cyan-500/20 to-makoclaw-accent/20 flex items-center justify-center">
                <svg class="w-5 h-5 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                </svg>
              </div>
              <div>
                <h2 class="font-bold text-lg text-makoclaw-text">Create with AI</h2>
                <p class="text-xs text-makoclaw-text-secondary">Describe what you want in plain language</p>
              </div>
            </div>
            <button @click="closeAiModal" class="modal-close-btn">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div class="modal-body space-y-4">
            <div>
              <label class="block text-sm font-medium text-makoclaw-text mb-2">What should this job do?</label>
              <textarea
                v-model="aiPrompt"
                rows="3"
                placeholder="e.g., Send me weather every day at 8am, or Backup tasks every 2 hours"
                class="input-field resize-none"
                :disabled="aiGenerating"
              ></textarea>
            </div>
            <button
              @click="generateCronWithAI"
              :disabled="!aiPrompt.trim() || aiGenerating"
              class="w-full py-3 bg-gradient-to-r from-cyan-500 to-makoclaw-accent text-white rounded-xl font-semibold shadow-lg shadow-cyan-500/20 hover:shadow-cyan-500/30 transition-all disabled:opacity-50 flex items-center justify-center gap-2"
            >
              <svg v-if="aiGenerating" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
              </svg>
              <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              {{ aiGenerating ? 'Generating...' : 'Generate Job' }}
            </button>

            <!-- AI Result -->
            <div v-if="aiResult" class="space-y-4 pt-4 border-t border-makoclaw-border/30">
              <div class="p-3 bg-cyan-500/10 border border-cyan-500/20 rounded-xl">
                <p class="text-sm text-cyan-400">{{ aiExplanation }}</p>
              </div>
              <div class="p-4 bg-makoclaw-bg/50 border border-makoclaw-border/30 rounded-xl space-y-2">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-makoclaw-text-secondary w-20">Name:</span>
                  <span class="text-sm text-makoclaw-text">{{ aiResult.name }}</span>
                </div>
                <div class="flex items-start gap-2">
                  <span class="text-sm font-medium text-makoclaw-text-secondary w-20">Message:</span>
                  <span class="text-sm text-makoclaw-text">{{ aiResult.message || aiResult.payload?.message }}</span>
                </div>
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-makoclaw-text-secondary w-20">Schedule:</span>
                  <span class="text-sm text-makoclaw-text font-mono">{{ formatSchedule(aiResult.schedule) }}</span>
                </div>
              </div>
              <div class="flex gap-2">
                <button @click="editAiResult" class="flex-1 btn-secondary flex items-center justify-center gap-2">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                  Edit
                </button>
                <button @click="saveAiResult" class="flex-1 btn-primary flex items-center justify-center gap-2">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                  Create
                </button>
              </div>
            </div>
          </div>

          <div class="modal-footer" v-if="!aiResult">
            <button @click="closeAiModal" class="btn-ghost">Close</button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
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

const showJsonModal = ref(false)
const jsonEditContent = ref('')
const jsonEditError = ref(null)
const savingJson = ref(false)

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
  { value: 'onetime', label: 'Once' },
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
  if (status.value.enabled) return 'Cron service running'
  if (status.value.reason === 'not_initialized') return 'Not available for this user'
  return 'Service not available'
})

const defaultForm = () => ({
  name: '',
  message: '',
  scheduleType: 'daily',
  time: '09:00',
  weekDays: [0],
  monthDay: 1,
  intervalValue: 30,
  intervalUnit: 'minutes',
  oneTimeDateTime: '',
  cronExpr: '',
  timezone: '',
  job_type: 'task',
  channel: '',
  to: ''
})

const form = ref(defaultForm())

const toggleWeekDay = (idx) => {
  const arr = form.value.weekDays
  const pos = arr.indexOf(idx)
  if (pos >= 0) {
    if (arr.length > 1) arr.splice(pos, 1)
  } else {
    arr.push(idx)
    arr.sort()
  }
}

const generatedExpr = computed(() => {
  const f = form.value
  if (f.scheduleType === 'interval' || f.scheduleType === 'onetime') return ''
  const [hh, mm] = (f.time || '09:00').split(':').map(Number)
  if (f.scheduleType === 'daily') return `${mm} ${hh} * * *`
  if (f.scheduleType === 'weekly') {
    const cronDays = f.weekDays.map(d => (d + 1) % 7).sort().join(',')
    return `${mm} ${hh} * * ${cronDays}`
  }
  if (f.scheduleType === 'monthly') return `${mm} ${hh} ${f.monthDay} * *`
  if (f.scheduleType === 'custom') return f.cronExpr || ''
  return ''
})

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
  const expr = generatedExpr.value
  if (!expr || expr.trim().split(/\s+/).length !== 5) return []
  try {
    const runs = getNextCronRuns(expr, 3)
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

function buildPayload() {
  const f = form.value
  const payload = {
    name: f.name.trim(),
    message: f.message.trim(),
    job_type: f.job_type || 'task',
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
    payload.schedule = {
      kind: 'cron',
      expr: generatedExpr.value,
      tz: f.timezone || undefined
    }
  }
  return payload
}

function jobToForm(job) {
  const f = defaultForm()
  f.name = job.name
  f.message = job.payload.message
  if (job.payload.job_type) {
    f.job_type = job.payload.job_type
  } else if (job.payload.deliver !== undefined) {
    f.job_type = job.payload.deliver ? 'reminder' : 'task'
  } else {
    f.job_type = 'task'
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

function parseCronExpr(expr) {
  const parts = expr.trim().split(/\s+/)
  if (parts.length !== 5) return null
  const [min, hr, dom, mon, dow] = parts
  const mm = parseInt(min, 10)
  const hh = parseInt(hr, 10)
  if (isNaN(mm) || isNaN(hh)) return null
  const pad = n => String(n).padStart(2, '0')
  const time = `${pad(hh)}:${pad(mm)}`
  if (dom === '*' && mon === '*' && dow === '*') return { type: 'daily', time }
  if (dom === '*' && mon === '*' && dow !== '*') {
    const cronDays = dow.split(',').map(Number).filter(n => !isNaN(n))
    if (cronDays.length > 0) {
      const weekDays = cronDays.map(d => d === 0 ? 6 : d - 1).sort()
      return { type: 'weekly', time, weekDays }
    }
  }
  if (mon === '*' && dow === '*') {
    const d = parseInt(dom, 10)
    if (!isNaN(d) && d >= 1 && d <= 31) return { type: 'monthly', time, monthDay: d }
  }
  return null
}

function getNextCronRuns(expr, count) {
  const parts = expr.trim().split(/\s+/)
  if (parts.length !== 5) return []
  const runs = []
  const now = new Date()
  let candidate = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours(), now.getMinutes() + 1, 0, 0)
  const maxIterations = 525960
  for (let i = 0; i < maxIterations && runs.length < count; i++) {
    if (matchesCron(parts, candidate)) runs.push(new Date(candidate))
    candidate = new Date(candidate.getTime() + 60000)
  }
  return runs
}

function matchesCron(parts, date) {
  const [minP, hrP, domP, monP, dowP] = parts
  return matchField(minP, date.getMinutes(), 0, 59)
    && matchField(hrP, date.getHours(), 0, 23)
    && matchField(domP, date.getDate(), 1, 31)
    && matchField(monP, date.getMonth() + 1, 1, 12)
    && matchField(dowP, date.getDay(), 0, 6)
}

function matchField(field, value) {
  if (field === '*') return true
  if (field.startsWith('*/')) {
    const step = parseInt(field.substring(2), 10)
    return !isNaN(step) && step > 0 && value % step === 0
  }
  if (field.includes('-') && !field.includes(',')) {
    const [lo, hi] = field.split('-').map(Number)
    return value >= lo && value <= hi
  }
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
  let jobType = payload.job_type
  if (!jobType && payload.deliver !== undefined) jobType = payload.deliver ? 'reminder' : 'task'
  return jobType || 'task'
}

const loadJobs = async () => {
  loading.value = true
  try {
    const data = await advancedService.fetchCronJobs()
    jobs.value = data.jobs || []
    status.value = data.status || { enabled: false }
  } catch {
    toast.error('Failed to load cron jobs')
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
    toast.error(`Failed: ${msg}`)
  }
}

const toggleJob = async (id, enabled) => {
  try {
    await advancedService.toggleCronJob(id, enabled)
    toast.success(enabled ? 'Job enabled' : 'Job disabled')
    await loadJobs()
  } catch {
    toast.error('Failed to toggle job')
  }
}

const runJob = async (job) => {
  try {
    await advancedService.runCronJob(job.id)
    toast.success(`Job '${job.name}' triggered`)
    await loadJobs()
  } catch (err) {
    const msg = err?.response?.data || err.message
    toast.error(`Failed: ${msg}`)
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
  } catch {
    toast.error('Failed to delete job')
  }
}

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
    if (!parsed.name || !parsed.schedule) {
      jsonEditError.value = 'Invalid structure: name and schedule required'
      return
    }
    const payload = parsed.payload || {}
    if (!payload.message && !parsed.message) {
      jsonEditError.value = 'Invalid structure: message required'
      return
    }
    savingJson.value = true
    jsonEditError.value = null
    const apiPayload = {
      name: parsed.name,
      message: payload.message || parsed.message,
      job_type: payload.job_type || parsed.job_type || 'task',
      channel: payload.channel || parsed.channel || '',
      to: payload.to || parsed.to || '',
      schedule: parsed.schedule
    }
    await advancedService.updateCronJob(editingJobId.value, apiPayload)
    toast.success('Job updated')
    showJsonModal.value = false
    await loadJobs()
  } catch (err) {
    if (err instanceof SyntaxError) {
      jsonEditError.value = `JSON Parse Error: ${err.message}`
    } else {
      jsonEditError.value = `Failed: ${err.response?.data?.error || err.message}`
    }
  } finally {
    savingJson.value = false
  }
}

const requestAiJsonFix = async () => {
  if (!jsonEditContent.value.trim()) {
    toast.error('No JSON content')
    return
  }
  savingJson.value = true
  try {
    const result = await advancedService.fixJsonWithAI(jsonEditContent.value, 'cron_job')
    jsonEditContent.value = result.fixed_json
    jsonEditError.value = null
    toast.success(result.changes || 'JSON validated')
  } catch (err) {
    toast.error(err?.response?.data?.error || 'AI fix failed')
  } finally {
    savingJson.value = false
  }
}

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
    toast.error('Enter a description')
    return
  }
  aiGenerating.value = true
  try {
    const result = await advancedService.createCronWithAI(aiPrompt.value.trim())
    aiResult.value = result.job
    aiExplanation.value = result.explanation
    toast.success('Job generated')
  } catch (err) {
    toast.error(err?.response?.data?.error || 'AI generation failed')
  } finally {
    aiGenerating.value = false
  }
}

const editAiResult = () => {
  if (!aiResult.value) return
  form.value = defaultForm()
  form.value.name = aiResult.value.name || ''
  form.value.message = aiResult.value.message || aiResult.value.payload?.message || ''
  form.value.job_type = aiResult.value.job_type || 'task'
  form.value.channel = aiResult.value.channel || ''
  form.value.to = aiResult.value.to || ''

  const schedule = aiResult.value.schedule
  if (schedule) {
    if (schedule.kind === 'cron' && schedule.expr) {
      const parsed = parseCronExpr(schedule.expr)
      if (parsed) {
        form.value.scheduleType = parsed.type
        form.value.time = parsed.time
        if (parsed.type === 'weekly') form.value.weekDays = parsed.weekDays
      } else {
        form.value.scheduleType = 'custom'
        form.value.cronExpr = schedule.expr
      }
    } else if (schedule.kind === 'every' && schedule.everyMs) {
      form.value.scheduleType = 'interval'
      form.value.intervalValue = Math.round(schedule.everyMs / 60000)
      form.value.intervalUnit = 'minutes'
    }
    if (schedule.tz) form.value.timezone = schedule.tz
  }

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
    toast.success('Job created')
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

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* Cron Cards */
.cron-card {
  @apply bg-makoclaw-surface/80 backdrop-blur-sm border border-makoclaw-border/50 rounded-xl p-4
         transition-all duration-200 hover:border-makoclaw-accent/20 hover:shadow-lg hover:shadow-cyan-500/5;
}

.cron-card-skeleton {
  @apply bg-makoclaw-surface/50 border border-makoclaw-border/30 rounded-xl p-4;
}

.cron-icon {
  @apply w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0 transition-colors;
}

.cron-icon-active {
  @apply bg-gradient-to-br from-cyan-500/20 to-makoclaw-accent/20 text-cyan-400;
}

.cron-icon-disabled {
  @apply bg-makoclaw-text-secondary/10 text-makoclaw-text-secondary;
}

.cron-action-btn {
  @apply px-3 py-1.5 text-xs font-medium text-makoclaw-text-secondary bg-makoclaw-bg/50 rounded-lg
         hover:bg-makoclaw-bg transition-all flex items-center gap-1.5;
}

/* Schedule Type Buttons */
.schedule-type-btn {
  @apply px-2 py-2 text-xs font-medium rounded-lg border border-makoclaw-border/50 bg-makoclaw-bg/30
         text-makoclaw-text-secondary hover:text-makoclaw-text hover:border-makoclaw-accent/30 transition-all;
}

.schedule-type-btn-active {
  @apply border-cyan-500/50 bg-cyan-500/10 text-cyan-400;
}

/* Day Buttons */
.day-btn {
  @apply w-9 h-9 text-xs font-medium rounded-lg border border-makoclaw-border/50 bg-makoclaw-bg/30
         text-makoclaw-text-secondary hover:border-makoclaw-accent/30 transition-all;
}

.day-btn-active {
  @apply border-cyan-500/50 bg-cyan-500/10 text-cyan-400;
}

/* Job Type Options */
.job-type-option {
  @apply flex flex-col items-center justify-center p-4 rounded-xl border border-makoclaw-border/50 bg-makoclaw-bg/30
         cursor-pointer transition-all hover:border-makoclaw-accent/30 text-center;
}

.job-type-option-active {
  @apply border-cyan-500/50 bg-cyan-500/10 text-cyan-400;
}

/* Modal Styles */
.modal-backdrop {
  @apply fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-modal p-4;
}

.modal-content {
  @apply bg-makoclaw-surface border border-makoclaw-border/50 rounded-2xl w-full shadow-2xl flex flex-col overflow-hidden;
}

.modal-header {
  @apply flex items-center justify-between p-4 border-b border-makoclaw-border/30;
}

.modal-body {
  @apply flex-1 overflow-auto p-4;
}

.modal-footer {
  @apply flex justify-end gap-3 p-4 border-t border-makoclaw-border/30;
}

.modal-close-btn {
  @apply p-2 hover:bg-makoclaw-bg rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors;
}
</style>
