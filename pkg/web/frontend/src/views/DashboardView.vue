<template>
  <div class="h-full flex flex-col bg-makoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-makoclaw-border bg-makoclaw-surface">
      <h2 class="text-xl font-bold bg-gradient-to-r from-makoclaw-accent to-blue-500 bg-clip-text text-transparent">Dashboard</h2>
      <p class="text-sm text-makoclaw-text-secondary mt-1">Overview of your makoclaw workspace</p>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-4 md:p-6 space-y-4 md:space-y-6 custom-scrollbar">

      <!-- Loading Skeleton -->
      <div v-if="loading" class="space-y-6">
        <!-- Stats Grid Skeleton -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div v-for="i in 4" :key="i" class="glass-panel rounded-2xl p-5">
            <div class="skeleton h-3 w-20 mb-3 rounded"></div>
            <div class="skeleton h-8 w-16 mb-2 rounded"></div>
          </div>
        </div>
        <!-- Recent Tasks Skeleton -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5">
            <div class="skeleton h-4 w-28 mb-4 rounded"></div>
            <div class="space-y-3">
              <div v-for="i in 3" :key="'task-'+i" class="flex items-center gap-3 p-3 rounded-lg bg-makoclaw-bg">
                <div class="skeleton w-2 h-2 rounded-full"></div>
                <div class="flex-1">
                  <div class="skeleton h-3.5 w-3/4 mb-2 rounded"></div>
                  <div class="skeleton h-2.5 w-16 rounded"></div>
                </div>
                <div class="skeleton h-2.5 w-12 rounded"></div>
              </div>
            </div>
          </div>
          <!-- Recent Sessions Skeleton -->
          <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5">
            <div class="skeleton h-4 w-36 mb-4 rounded"></div>
            <div class="space-y-3">
              <div v-for="i in 3" :key="'session-'+i" class="flex items-center gap-3 p-3 rounded-lg bg-makoclaw-bg">
                <div class="skeleton w-5 h-5 rounded"></div>
                <div class="flex-1">
                  <div class="skeleton h-3.5 w-1/2 mb-2 rounded"></div>
                  <div class="skeleton h-2.5 w-20 rounded"></div>
                </div>
                <div class="skeleton h-2.5 w-12 rounded"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <template v-else>
        <!-- Welcome banner -->
        <div class="relative overflow-hidden bg-gradient-to-r from-makoclaw-accent to-blue-600 rounded-2xl p-8 mb-8 shadow-lg shadow-makoclaw-accent/10 group">
          <div class="absolute top-0 right-0 p-4 opacity-10 transition-transform group-hover:scale-110 duration-700">
            <svg class="w-32 h-32" fill="currentColor" viewBox="0 0 24 24"><path d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>
          </div>
          <div class="relative z-10">
            <h3 class="text-2xl md:text-3xl font-bold text-white">Welcome back, {{ authStore.user?.username || 'Commander' }}</h3>
            <p class="text-white/80 mt-2 text-sm md:text-base max-w-lg">Your makoclaw workspace is active. Here's a quick summary of your agent's activity and system performance.</p>
          </div>
        </div>
        <!-- Stats Grid -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div class="glass-panel rounded-2xl p-5 transition-all duration-300 hover:shadow-makoclaw-accent/5 hover:-translate-y-[1px] animate-fadeUp opacity-0" style="animation-delay: 0ms">
            <div class="text-xs font-semibold uppercase tracking-wider text-makoclaw-text-secondary opacity-70">Total Tasks</div>
            <div class="text-3xl font-bold mt-2">{{ stats.totalTasks }}</div>
          </div>
          <div class="glass-panel rounded-2xl p-5 transition-all duration-300 hover:shadow-makoclaw-accent/5 hover:-translate-y-[1px] animate-fadeUp opacity-0" style="animation-delay: 50ms">
            <div class="text-xs font-semibold uppercase tracking-wider text-makoclaw-text-secondary opacity-70">In Progress</div>
            <div class="text-3xl font-bold mt-2 text-makoclaw-accent">{{ stats.inProgress }}</div>
          </div>
          <div class="glass-panel rounded-2xl p-5 transition-all duration-300 hover:shadow-makoclaw-accent/5 hover:-translate-y-[1px] animate-fadeUp opacity-0" style="animation-delay: 100ms">
            <div class="text-xs font-semibold uppercase tracking-wider text-makoclaw-text-secondary opacity-70">Chat Sessions</div>
            <div class="text-3xl font-bold mt-2 text-makoclaw-accent">{{ stats.chatSessions }}</div>
          </div>
          <div class="glass-panel rounded-2xl p-5 transition-all duration-300 hover:shadow-makoclaw-accent/5 hover:-translate-y-[1px] animate-fadeUp opacity-0" style="animation-delay: 150ms">
            <div class="text-xs font-semibold uppercase tracking-wider text-makoclaw-text-secondary opacity-70">Total Messages</div>
            <div class="text-3xl font-bold mt-2 text-makoclaw-accent">{{ stats.totalMessages }}</div>
          </div>
        </div>

        <!-- Charts Section -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6" v-if="metricsData">
          <!-- Model Activity Ratio -->
          <div class="glass-panel rounded-2xl p-6 flex flex-col items-center">
            <h3 class="text-sm font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-6 self-start">Model Activity</h3>
            <div class="w-full relative flex-1 min-h-[250px] flex items-center justify-center">
               <Doughnut v-if="modelChartData.labels.length > 0" :data="modelChartData" :options="chartOptions" />
               <div v-else class="text-sm text-makoclaw-text-secondary opacity-50 italic">No model usage data available yet.</div>
            </div>
          </div>
          
          <!-- Tasks by Status -->
          <div class="glass-panel rounded-2xl p-6 flex flex-col items-center">
             <h3 class="text-sm font-bold uppercase tracking-widest text-makoclaw-text-secondary mb-6 self-start">Tasks by Status</h3>
             <div class="w-full relative flex-1 min-h-[250px] flex items-center justify-center">
                <Bar v-if="taskChartData.labels.length > 0" :data="taskChartData" :options="barOptions" />
                <div v-else class="text-sm text-makoclaw-text-secondary opacity-50 italic">No task data available yet.</div>
             </div>
          </div>
        </div>

        <!-- Observability Stats Grid -->
        <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5" v-if="metricsData">
          <h3 class="font-semibold mb-4">Total System Activity</h3>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 text-center">
            <div>
              <p class="text-xl font-bold">{{ metricsData.llm_calls }}</p>
              <p class="text-xs text-makoclaw-text-secondary mt-1">LLM Calls</p>
            </div>
            <div>
              <p class="text-xl font-bold">{{ metricsData.tool_calls }}</p>
              <p class="text-xs text-makoclaw-text-secondary mt-1">Tool Calls</p>
            </div>
            <div>
              <p class="text-xl font-bold">{{ metricsData.agent_runs }}</p>
              <p class="text-xs text-makoclaw-text-secondary mt-1">Agent Runs</p>
            </div>
            <div>
              <p class="text-xl font-bold">{{ formatNumber(metricsData.llm_tokens_in + metricsData.llm_tokens_out) }}</p>
              <p class="text-xs text-makoclaw-text-secondary mt-1">Tokens Processed</p>
            </div>
          </div>
        </div>

        <!-- Two Column Layout -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">

          <!-- Recent Tasks -->
          <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5">
            <div class="flex items-center justify-between mb-4">
              <h3 class="font-semibold">Recent Tasks</h3>
              <router-link to="/tasks" class="text-sm text-makoclaw-accent hover:underline">View all</router-link>
            </div>
            <div v-if="recentTasks.length === 0" class="text-sm text-makoclaw-text-secondary py-4 text-center">No tasks yet</div>
            <div v-else class="space-y-2">
              <div
                v-for="task in recentTasks"
                :key="task.id"
                class="flex items-center gap-3 p-3 rounded-lg bg-makoclaw-bg list-item-interactive"
              >
                <span
                  class="w-2 h-2 rounded-full flex-shrink-0"
                  :class="{
                    'bg-makoclaw-text-secondary/50': task.status === 'backlog',
                    'bg-makoclaw-warning': task.status === 'todo',
                    'bg-makoclaw-accent': task.status === 'in_progress',
                    'bg-makoclaw-warning': task.status === 'review',
                    'bg-makoclaw-success': task.status === 'done'
                  }"
                ></span>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <div class="text-sm font-medium truncate">{{ task.title }}</div>
                    <SpecialistBadge v-if="task.agent" :name="task.agent" class="scale-90" />
                  </div>
                  <div class="text-xs text-makoclaw-text-secondary capitalize">{{ task.status.replace('_', ' ') }}</div>
                </div>
                <div class="text-xs text-makoclaw-text-secondary flex-shrink-0">
                  {{ formatDate(task.created_at) }}
                </div>
              </div>
            </div>
          </div>

          <!-- Recent Sessions -->
          <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5">
            <div class="flex items-center justify-between mb-4">
              <h3 class="font-semibold">Recent Conversations</h3>
              <router-link to="/history" class="text-sm text-makoclaw-accent hover:underline">View all</router-link>
            </div>
            <div v-if="recentSessions.length === 0" class="text-sm text-makoclaw-text-secondary py-4 text-center">No conversations yet</div>
            <div v-else class="space-y-2">
              <div
                v-for="session in recentSessions"
                :key="session.session_id"
                class="flex items-center gap-3 p-3 rounded-lg bg-makoclaw-bg list-item-interactive"
              >
                <svg v-if="session.session_id.startsWith('web:chat:')" class="w-5 h-5 flex-shrink-0 text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
                <svg v-else class="w-5 h-5 flex-shrink-0 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
                <div class="flex-1 min-w-0">
                  <div class="text-sm font-medium truncate">{{ sessionLabel(session.session_id) }}</div>
                  <div class="text-xs text-makoclaw-text-secondary">{{ session.message_count }} messages</div>
                </div>
                <div class="text-xs text-makoclaw-text-secondary flex-shrink-0">
                  {{ formatDate(session.last_activity) }}
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Quick Actions -->
        <div class="bg-makoclaw-surface border border-makoclaw-border rounded-xl p-5">
          <h3 class="font-semibold mb-4">Quick Actions</h3>
          <div class="flex flex-wrap gap-3">
            <router-link
              to="/chat"
              class="flex items-center gap-2 px-4 py-2 bg-makoclaw-accent text-white rounded-lg hover:bg-makoclaw-accent/90 transition-colors text-sm"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
              New Chat
            </router-link>
            <router-link
              to="/tasks"
              class="flex items-center gap-2 px-4 py-2 bg-makoclaw-bg border border-makoclaw-border text-makoclaw-text rounded-lg hover:bg-makoclaw-border/50 transition-colors text-sm"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
              New Task
            </router-link>
            <router-link
              to="/history"
              class="flex items-center gap-2 px-4 py-2 bg-makoclaw-bg border border-makoclaw-border text-makoclaw-text rounded-lg hover:bg-makoclaw-border/50 transition-colors text-sm"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
              Browse History
            </router-link>
            <router-link
              to="/memory"
              class="flex items-center gap-2 px-4 py-2 bg-makoclaw-bg border border-makoclaw-border text-makoclaw-text rounded-lg hover:bg-makoclaw-border/50 transition-colors text-sm"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" /></svg>
              Edit Memory
            </router-link>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/authStore'
import taskService from '../services/taskService'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'
import SpecialistBadge from '../components/Chat/SpecialistBadge.vue'
import { Doughnut, Bar } from 'vue-chartjs'
import { Chart as ChartJS, Title, Tooltip, Legend, ArcElement, CategoryScale, LinearScale, BarElement } from 'chart.js'

ChartJS.register(Title, Tooltip, Legend, ArcElement, CategoryScale, LinearScale, BarElement)

const authStore = useAuthStore()
const toast = useToast()
const loading = ref(true)
const tasks = ref([])
const sessions = ref([])
const metricsData = ref(null)

const formatNumber = (num) => {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'k'
  return num
}

const stats = computed(() => {
  const totalTasks = tasks.value.length
  const inProgress = tasks.value.filter(t => t.status === 'in_progress').length
  const chatSessions = sessions.value.filter(s => s.session_id.startsWith('web:chat:')).length
  const totalMessages = sessions.value.reduce((sum, s) => sum + (s.message_count || 0), 0)
  return { totalTasks, inProgress, chatSessions, totalMessages }
})

const statusBreakdown = computed(() => {
  const counts = { backlog: 0, todo: 0, in_progress: 0, review: 0, done: 0 }
  tasks.value.forEach(t => {
    if (counts[t.status] !== undefined) counts[t.status]++
  })
  return [
    { status: 'backlog', label: 'Backlog', count: counts.backlog, color: 'text-makoclaw-text-secondary' },
    { status: 'todo', label: 'To Do', count: counts.todo, color: 'text-makoclaw-warning' },
    { status: 'in_progress', label: 'In Progress', count: counts.in_progress, color: 'text-makoclaw-accent' },
    { status: 'review', label: 'Review', count: counts.review, color: 'text-makoclaw-warning' },
    { status: 'done', label: 'Done', count: counts.done, color: 'text-makoclaw-success' }
  ]
})

const recentTasks = computed(() => tasks.value.slice(0, 5))

const recentSessions = computed(() => sessions.value.slice(0, 5))

const sessionLabel = (sessionId) => {
  if (sessionId.startsWith('web:chat:')) {
    return 'Chat ' + sessionId.replace('web:chat:', '').substring(0, 8)
  }
  if (sessionId.startsWith('web:task:')) {
    return 'Task #' + sessionId.replace('web:task:', '')
  }
  return sessionId.substring(0, 20)
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diff = now - d
  if (diff < 60000) return 'just now'
  if (diff < 3600000) return Math.floor(diff / 60000) + 'm ago'
  if (diff < 86400000) return Math.floor(diff / 3600000) + 'h ago'
  if (diff < 604800000) return Math.floor(diff / 86400000) + 'd ago'
  return d.toLocaleDateString()
}

// Chart Options
const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'right',
      labels: { color: '#9CA3AF', font: { size: 12 } }
    }
  },
  cutout: '65%'
}

const barOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false }
  },
  scales: {
    y: {
      beginAtZero: true,
      grid: { color: 'rgba(255,255,255,0.05)' },
      ticks: { color: '#9CA3AF' }
    },
    x: {
      grid: { display: false },
      ticks: { color: '#9CA3AF' }
    }
  }
}

// Data Processing for Charts
const modelChartData = computed(() => {
  if (!metricsData.value || !metricsData.value.llm_by_model) return { labels: [], datasets: [{ data: [] }] }
  
  const labels = Object.keys(metricsData.value.llm_by_model)
  const data = Object.values(metricsData.value.llm_by_model).map(m => m.calls)
  
  return {
    labels,
    datasets: [{
      backgroundColor: ['#3b82f6', '#22c55e', '#84cc16', '#06b6d4', '#0891b2', '#14b8a6'],
      borderColor: '#1e1e2e',
      borderWidth: 2,
      data
    }]
  }
})

const taskChartData = computed(() => {
  return {
    labels: statusBreakdown.value.map(s => s.label),
    datasets: [{
      backgroundColor: ['#9ca3af', '#facc15', '#60a5fa', '#fb923c', '#4ade80'],
      data: statusBreakdown.value.map(s => s.count),
      borderRadius: 4
    }]
  }
})

onMounted(async () => {
  try {
    const [tasksData, sessionsData, metrics] = await Promise.all([
      taskService.fetchTasks(false),
      taskService.fetchChatSessions(),
      advancedService.fetchMetrics()
    ])
    tasks.value = tasksData.tasks || []
    sessions.value = sessionsData.sessions || []
    metricsData.value = metrics || null
  } catch (err) {
    console.error('Failed to load dashboard data:', err)
    toast.error('Failed to load dashboard data')
  } finally {
    loading.value = false
  }
})
</script>

