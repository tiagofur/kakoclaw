<template>
  <div class="h-full flex flex-col bg-picoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-picoclaw-border bg-picoclaw-surface flex items-center justify-between">
      <div>
        <h2 class="text-xl font-bold bg-gradient-to-r from-picoclaw-accent to-purple-500 bg-clip-text text-transparent">Cron Jobs</h2>
        <p class="text-sm text-picoclaw-text-secondary mt-1">Scheduled tasks and recurring automations</p>
      </div>
      <button
        @click="openCreateModal"
        class="flex items-center gap-2 px-4 py-2 bg-picoclaw-accent text-white rounded-lg hover:bg-picoclaw-accent/90 transition-colors text-sm"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
        New Job
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6 custom-scrollbar">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-picoclaw-accent"></div>
      </div>

      <template v-else>
        <!-- Status Banner -->
        <div class="mb-4 px-4 py-3 rounded-lg border"
          :class="status.enabled ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-400' : 'bg-yellow-500/10 border-yellow-500/20 text-yellow-400'"
        >
          <span class="font-medium">Cron service: {{ status.enabled ? 'Running' : 'Not available' }}</span>
          <span v-if="status.jobs !== undefined" class="ml-2 text-sm opacity-75">({{ status.jobs }} active jobs)</span>
        </div>

        <div v-if="jobs.length === 0" class="text-center py-12 text-picoclaw-text-secondary">
          <p class="text-lg">No cron jobs configured</p>
          <p class="text-sm mt-2">Create a scheduled job to automate tasks</p>
        </div>

        <div class="space-y-3">
          <div
            v-for="job in jobs"
            :key="job.id"
            class="bg-picoclaw-surface border border-picoclaw-border rounded-xl p-5"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <h3 class="font-semibold">{{ job.name }}</h3>
                  <span
                    class="px-2 py-0.5 text-xs rounded-full"
                    :class="job.enabled ? 'bg-emerald-500/10 text-emerald-400' : 'bg-gray-500/10 text-gray-400'"
                  >{{ job.enabled ? 'Active' : 'Disabled' }}</span>
                </div>
                <p class="text-sm text-picoclaw-text-secondary mt-1">{{ job.payload.message }}</p>
              </div>
            </div>

            <div class="flex items-center gap-4 mt-3 text-xs text-picoclaw-text-secondary">
              <span>Schedule: <span class="text-picoclaw-text font-mono">{{ formatSchedule(job.schedule) }}</span></span>
              <span v-if="job.schedule.tz" class="font-mono">TZ: {{ job.schedule.tz }}</span>
              <span v-if="job.state.lastStatus">Last: {{ job.state.lastStatus }}</span>
              <span v-if="job.state.nextRunAtMs">Next: {{ formatTimestamp(job.state.nextRunAtMs) }}</span>
            </div>

            <div class="flex items-center gap-2 mt-3">
              <button
                @click="openEditModal(job)"
                class="px-3 py-1.5 text-xs text-picoclaw-accent bg-picoclaw-accent/10 rounded-lg hover:bg-picoclaw-accent/20 transition-colors"
              >Edit</button>
              <button
                @click="toggleJob(job.id, !job.enabled)"
                class="px-3 py-1.5 text-xs rounded-lg transition-colors"
                :class="job.enabled ? 'bg-yellow-500/10 text-yellow-400 hover:bg-yellow-500/20' : 'bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20'"
              >{{ job.enabled ? 'Disable' : 'Enable' }}</button>
              <button
                @click="confirmDeleteJob(job)"
                class="px-3 py-1.5 text-xs text-red-400 bg-red-500/10 rounded-lg hover:bg-red-500/20 transition-colors"
              >Delete</button>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Create / Edit Job Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="showModal = false">
      <div class="bg-picoclaw-surface border border-picoclaw-border rounded-xl max-w-lg w-full max-h-[90vh] overflow-y-auto p-6">
        <h3 class="font-semibold text-lg mb-4">{{ editingJobId ? 'Edit Cron Job' : 'Create Cron Job' }}</h3>
        <div class="space-y-4">
          <!-- Name -->
          <div>
            <label class="block text-sm font-medium mb-1">Name</label>
            <input v-model="form.name" type="text" placeholder="My scheduled task"
              class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent" />
          </div>

          <!-- Message -->
          <div>
            <label class="block text-sm font-medium mb-1">Message (what the agent should do)</label>
            <textarea v-model="form.message" rows="3" placeholder="Summarize today's tasks and send a report..."
              class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent resize-none" />
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
                  ? 'border-picoclaw-accent bg-picoclaw-accent/10 text-picoclaw-accent'
                  : 'border-picoclaw-border bg-picoclaw-bg text-picoclaw-text-secondary hover:border-picoclaw-accent/50'"
              >{{ opt.label }}</button>
            </div>
          </div>

          <!-- Daily: time picker -->
          <div v-if="form.scheduleType === 'daily'" class="space-y-2">
            <label class="block text-sm font-medium">Run at</label>
            <input v-model="form.time" type="time"
              class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent" />
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
                    ? 'border-picoclaw-accent bg-picoclaw-accent/10 text-picoclaw-accent'
                    : 'border-picoclaw-border bg-picoclaw-bg text-picoclaw-text-secondary hover:border-picoclaw-accent/50'"
                >{{ day }}</button>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Time</label>
              <input v-model="form.time" type="time"
                class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent" />
            </div>
          </div>

          <!-- Monthly: day-of-month + time -->
          <div v-if="form.scheduleType === 'monthly'" class="space-y-3">
            <div>
              <label class="block text-sm font-medium mb-2">Day of month</label>
              <select v-model.number="form.monthDay"
                class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent">
                <option v-for="d in 31" :key="d" :value="d">{{ d }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Time</label>
              <input v-model="form.time" type="time"
                class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent" />
            </div>
          </div>

          <!-- Interval: every N minutes/hours -->
          <div v-if="form.scheduleType === 'interval'" class="space-y-2">
            <label class="block text-sm font-medium">Repeat every</label>
            <div class="flex gap-2">
              <input v-model.number="form.intervalValue" type="number" min="1" placeholder="30"
                class="flex-1 px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent" />
              <select v-model="form.intervalUnit"
                class="px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent">
                <option value="minutes">Minutes</option>
                <option value="hours">Hours</option>
              </select>
            </div>
          </div>

          <!-- One-time: date + time picker -->
          <div v-if="form.scheduleType === 'onetime'" class="space-y-2">
            <label class="block text-sm font-medium">Run at</label>
            <input v-model="form.oneTimeDateTime" type="datetime-local"
              class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent" />
          </div>

          <!-- Custom: raw cron expression -->
          <div v-if="form.scheduleType === 'custom'" class="space-y-2">
            <label class="block text-sm font-medium">Cron Expression</label>
            <input v-model="form.cronExpr" type="text" placeholder="0 9 * * 1-5"
              class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent font-mono" />
            <p class="text-xs text-picoclaw-text-secondary">Standard 5-field cron: minute hour day-of-month month day-of-week</p>
          </div>

          <!-- Timezone (for cron-based schedules) -->
          <div v-if="['daily', 'weekly', 'monthly', 'custom'].includes(form.scheduleType)" class="space-y-2">
            <label class="block text-sm font-medium">Timezone</label>
            <select v-model="form.timezone"
              class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent">
              <option value="">UTC (default)</option>
              <option v-for="tz in commonTimezones" :key="tz" :value="tz">{{ tz }}</option>
            </select>
          </div>

          <!-- Generated expression preview -->
          <div v-if="generatedExpr" class="px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg">
            <p class="text-xs text-picoclaw-text-secondary mb-1">Generated expression</p>
            <code class="text-sm font-mono text-picoclaw-accent">{{ generatedExpr }}</code>
          </div>

          <!-- Next 3 runs preview -->
          <div v-if="nextRuns.length > 0" class="px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg">
            <p class="text-xs text-picoclaw-text-secondary mb-1">Next runs</p>
            <ul class="space-y-0.5">
              <li v-for="(run, i) in nextRuns" :key="i" class="text-sm text-picoclaw-text font-mono">{{ run }}</li>
            </ul>
          </div>

          <!-- Deliver to channel -->
          <div class="flex items-center gap-2">
            <input v-model="form.deliver" type="checkbox" id="deliver" class="rounded" />
            <label for="deliver" class="text-sm">Deliver result to channel</label>
          </div>
          <div v-if="form.deliver" class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-sm font-medium mb-1">Channel</label>
              <input v-model="form.channel" type="text" placeholder="telegram"
                class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">To (Chat ID)</label>
              <input v-model="form.to" type="text" placeholder=""
                class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded-lg text-sm outline-none focus:border-picoclaw-accent" />
            </div>
          </div>
        </div>

        <!-- Modal Actions -->
        <div class="flex justify-end gap-3 mt-6">
          <button @click="showModal = false"
            class="px-4 py-2 text-sm text-picoclaw-text-secondary hover:text-picoclaw-text transition-colors">Cancel</button>
          <button @click="submitJob" :disabled="!canSubmit"
            class="px-4 py-2 text-sm bg-picoclaw-accent text-white rounded-lg hover:bg-picoclaw-accent/90 transition-colors disabled:opacity-50">
            {{ editingJobId ? 'Save' : 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" @click.self="showDeleteConfirm = false">
      <div class="bg-picoclaw-surface border border-picoclaw-border rounded-xl max-w-sm w-full p-6">
        <h3 class="font-semibold text-lg mb-2">Delete Job</h3>
        <p class="text-sm text-picoclaw-text-secondary mb-4">
          Are you sure you want to delete <span class="font-medium text-picoclaw-text">{{ deletingJob?.name }}</span>? This action cannot be undone.
        </p>
        <div class="flex justify-end gap-3">
          <button @click="showDeleteConfirm = false"
            class="px-4 py-2 text-sm text-picoclaw-text-secondary hover:text-picoclaw-text transition-colors">Cancel</button>
          <button @click="executeDeleteJob"
            class="px-4 py-2 text-sm bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors">Delete</button>
        </div>
      </div>
    </div>
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
  deliver: false,
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
    deliver: f.deliver,
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
  f.deliver = job.payload.deliver || false
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
.custom-scrollbar::-webkit-scrollbar { width: 8px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background-color: rgba(156, 163, 175, 0.5); border-radius: 4px; }
</style>
