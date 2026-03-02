<template>
  <div class="h-full flex flex-col bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-makoclaw-accent/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-15 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-indigo-500/20 via-transparent to-transparent" />
    </div>

    <!-- Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <div class="flex items-center gap-3">
          <!-- Icon Container -->
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-makoclaw-accent/20 to-indigo-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-makoclaw-accent/10">
            <svg
              class="w-5 h-5 sm:w-6 sm:h-6 text-makoclaw-accent"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z"
              />
            </svg>
          </div>

          <!-- Title -->
          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-makoclaw-accent bg-clip-text text-transparent">
              Dashboard
            </h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5 flex items-center gap-2">
              <span class="w-1.5 h-1.5 rounded-full bg-makoclaw-success animate-pulse" />
              System operational • {{ authStore.user?.username || 'User' }}
            </p>
          </div>

          <!-- Refresh Button -->
          <button
            class="p-2.5 min-h-[40px] min-w-[40px] rounded-xl bg-makoclaw-surface/50 border border-makoclaw-border/50 hover:bg-makoclaw-surface-hover hover:border-makoclaw-accent/30 transition-all flex items-center justify-center active:scale-95"
            @click="reloadData"
          >
            <svg
              class="w-4 h-4 text-makoclaw-text-secondary"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto custom-scrollbar p-4 sm:p-6 relative z-10">
      <!-- Loading Skeleton -->
      <div
        v-if="loading"
        class="space-y-5 animate-pulse"
      >
        <div class="h-40 bg-makoclaw-surface/30 rounded-2xl border border-makoclaw-border/30" />
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div
            v-for="i in 4"
            :key="i"
            class="h-28 bg-makoclaw-surface/30 rounded-2xl border border-makoclaw-border/30"
          />
        </div>
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div class="h-72 bg-makoclaw-surface/30 rounded-2xl border border-makoclaw-border/30" />
          <div class="h-72 bg-makoclaw-surface/30 rounded-2xl border border-makoclaw-border/30" />
        </div>
      </div>

      <template v-else>
        <div class="max-w-7xl mx-auto space-y-6">
          <!-- Welcome Banner -->
          <div class="relative overflow-hidden bg-gradient-to-br from-makoclaw-accent/20 via-blue-600/15 to-indigo-700/20 backdrop-blur-xl rounded-2xl p-5 sm:p-6 shadow-xl shadow-makoclaw-accent/10 group border border-white/10 ring-1 ring-white/5">
            <!-- Subtle mesh overlay -->
            <div class="absolute inset-0 opacity-20">
              <div class="absolute top-0 left-0 w-full h-full bg-[radial-gradient(circle_at_50%_50%,rgba(255,255,255,0.1)_0%,transparent_50%)]" />
            </div>

            <div class="absolute top-0 right-0 p-6 opacity-10 hidden md:block">
              <svg
                class="w-20 h-20"
                fill="currentColor"
                viewBox="0 0 24 24"
              ><path d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
            </div>

            <div class="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-4">
              <div>
                <span class="inline-block px-3 py-1 bg-makoclaw-accent/20 backdrop-blur-md rounded-full text-[10px] font-bold text-makoclaw-accent uppercase tracking-wider mb-3 border border-makoclaw-accent/30">Overview</span>
                <h3 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-makoclaw-accent bg-clip-text text-transparent">
                  Welcome back, {{ authStore.user?.username || 'User' }}
                </h3>
                <p class="text-makoclaw-text-secondary mt-1 text-sm max-w-lg">
                  Your workspace is ready. All systems operational across connected channels.
                </p>
              </div>

              <div class="flex flex-wrap gap-3">
                <router-link
                  to="/chat"
                  class="px-4 sm:px-5 py-2.5 min-h-[40px] bg-makoclaw-accent/20 backdrop-blur-md border border-makoclaw-accent/30 text-makoclaw-accent rounded-xl font-bold shadow-lg shadow-makoclaw-accent/10 hover:bg-makoclaw-accent/30 hover:shadow-makoclaw-accent/20 transition-all active:scale-95 text-sm flex items-center gap-2 ring-1 ring-makoclaw-accent/20"
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
                      d="M12 4v16m8-8H4"
                    />
                  </svg>
                  <span class="hidden sm:inline">New Session</span>
                  <span class="sm:hidden">New</span>
                </router-link>
                <router-link
                  to="/tasks"
                  class="px-4 sm:px-5 py-2.5 min-h-[40px] bg-makoclaw-surface/50 backdrop-blur-md border border-makoclaw-border/40 text-makoclaw-text rounded-xl font-bold hover:bg-makoclaw-surface-hover hover:border-makoclaw-accent/30 transition-all active:scale-95 text-sm ring-1 ring-white/5"
                >
                  View Tasks
                </router-link>
              </div>
            </div>
          </div>

          <!-- Stats Grid -->
          <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
            <div
              v-for="stat in statItems"
              :key="stat.label"
              class="bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 rounded-2xl p-4 sm:p-5 transition-all duration-300 hover:bg-makoclaw-surface/50 hover:border-makoclaw-accent/20 hover:shadow-lg hover:shadow-makoclaw-accent/5 group ring-1 ring-white/5 relative overflow-hidden"
            >
              <!-- Hover accent line -->
              <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-indigo-500 group-hover:w-full transition-all duration-500 opacity-50" />
              <div class="flex justify-between items-start">
                <div
                  class="w-9 h-9 sm:w-10 sm:h-10 rounded-xl bg-gradient-to-br flex items-center justify-center shadow-lg transition-transform duration-300 group-hover:scale-105 backdrop-blur-md ring-1 ring-white/20"
                  :class="stat.iconBg"
                >
                  <component
                    :is="stat.icon"
                    class="w-4 h-4 sm:w-5 sm:h-5 text-white"
                  />
                </div>
                <div
                  v-if="stat.trend"
                  class="px-2 py-0.5 rounded-full bg-makoclaw-success/10 text-[9px] font-bold text-makoclaw-success border border-makoclaw-success/20 flex items-center gap-1"
                >
                  <svg
                    class="w-2 h-2"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  ><path d="M24 22h-24l12-20z" /></svg>
                  {{ stat.trend }}%
                </div>
              </div>
              <div class="mt-3 sm:mt-4">
                <div class="text-[10px] font-bold uppercase tracking-wider text-makoclaw-text-secondary/60 mb-1">
                  {{ stat.label }}
                </div>
                <div class="text-xl sm:text-2xl font-bold text-makoclaw-text">
                  {{ stat.value }}
                </div>
              </div>
            </div>
          </div>

          <!-- Main Dashboard Grid -->
          <div class="grid grid-cols-1 xl:grid-cols-3 gap-6">
            <!-- Charts Section (Left) -->
            <div class="xl:col-span-2 space-y-6">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Model Usage Chart -->
                <div class="bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 rounded-2xl p-5 ring-1 ring-white/5">
                  <div class="flex items-center justify-between mb-5">
                    <h3 class="text-sm font-bold text-makoclaw-text">
                      Model Usage
                    </h3>
                    <div class="w-2 h-2 rounded-full bg-blue-500 animate-pulse shadow-lg shadow-blue-500/50" />
                  </div>
                  <div class="relative min-h-[280px] flex items-center justify-center">
                    <Doughnut
                      v-if="modelChartData.labels.length > 0"
                      :data="modelChartData"
                      :options="chartOptions"
                    />
                    <div
                      v-else
                      class="text-sm text-makoclaw-text-secondary/50"
                    >
                      No data available
                    </div>
                  </div>
                </div>

                <!-- Task Status Chart -->
                <div class="bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 rounded-2xl p-5 ring-1 ring-white/5">
                  <div class="flex items-center justify-between mb-5">
                    <h3 class="text-sm font-bold text-makoclaw-text">
                      Task Status
                    </h3>
                    <div class="w-2 h-2 rounded-full bg-emerald-500 shadow-lg shadow-emerald-500/50" />
                  </div>
                  <div class="relative min-h-[280px] flex items-center justify-center">
                    <Bar
                      v-if="taskChartData.labels.length > 0"
                      :data="taskChartData"
                      :options="barOptions"
                    />
                    <div
                      v-else
                      class="text-sm text-makoclaw-text-secondary/50"
                    >
                      No tasks yet
                    </div>
                  </div>
                </div>
              </div>

              <!-- Metrics Summary -->
              <div class="bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 rounded-2xl p-5 ring-1 ring-white/5">
                <h3 class="text-sm font-bold text-makoclaw-text mb-5">
                  System Metrics
                </h3>
                <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 sm:gap-6">
                  <div
                    v-for="metric in detailedMetrics"
                    :key="metric.label"
                    class="p-3 sm:p-4 rounded-xl hover:bg-makoclaw-accent/5 transition-all group border border-transparent hover:border-makoclaw-accent/10"
                  >
                    <div class="text-[10px] font-bold text-makoclaw-text-secondary/60 uppercase tracking-wider mb-1 group-hover:text-makoclaw-accent transition-colors">
                      {{ metric.label }}
                    </div>
                    <div class="text-xl sm:text-2xl font-bold text-makoclaw-text">
                      {{ metric.value }}
                    </div>
                    <div class="h-1 w-8 bg-makoclaw-accent/20 rounded-full mt-2 group-hover:w-full transition-all duration-500" />
                  </div>
                </div>
              </div>
            </div>

            <!-- Right Sidebar -->
            <div class="space-y-6">
              <!-- Quick Actions -->
              <div class="bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 rounded-2xl p-5 ring-1 ring-white/5">
                <h3 class="text-sm font-bold text-makoclaw-text mb-4">
                  Quick Actions
                </h3>
                <div class="grid grid-cols-2 gap-3">
                  <router-link
                    v-for="action in quickActions"
                    :key="action.label"
                    :to="action.to"
                    class="flex flex-col items-center justify-center p-4 rounded-xl border border-makoclaw-border/30 hover:border-makoclaw-accent/30 bg-makoclaw-bg/30 hover:bg-makoclaw-accent/5 transition-all group active:scale-95"
                  >
                    <div
                      class="w-10 h-10 rounded-xl flex items-center justify-center mb-3 transition-transform group-hover:scale-105 shadow-lg backdrop-blur-md ring-1 ring-white/20"
                      :class="action.color"
                    >
                      <component
                        :is="action.icon"
                        class="w-5 h-5 text-white"
                      />
                    </div>
                    <span class="text-[10px] font-bold uppercase tracking-wide text-makoclaw-text-secondary group-hover:text-makoclaw-accent transition-colors">{{ action.label }}</span>
                  </router-link>
                </div>
              </div>

              <!-- Recent Activity -->
              <div class="bg-makoclaw-surface/30 backdrop-blur-sm border border-makoclaw-border/30 rounded-2xl p-5 ring-1 ring-white/5 flex flex-col h-[420px]">
                <div class="flex items-center justify-between mb-4">
                  <h3 class="text-sm font-bold text-makoclaw-text">
                    Recent Activity
                  </h3>
                  <router-link
                    to="/history"
                    class="text-[10px] font-bold text-makoclaw-accent uppercase hover:underline tracking-wide"
                  >
                    View All
                  </router-link>
                </div>

                <div class="flex-1 overflow-y-auto custom-scrollbar space-y-3">
                  <div
                    v-if="recentActivity.length === 0"
                    class="flex flex-col items-center justify-center py-12 text-center"
                  >
                    <div class="w-12 h-12 rounded-xl bg-makoclaw-bg/50 border border-makoclaw-border/50 flex items-center justify-center mb-4">
                      <svg
                        class="w-5 h-5 text-makoclaw-text-secondary/50"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      ><path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                      /></svg>
                    </div>
                    <p class="text-xs text-makoclaw-text-secondary/60">
                      No recent activity
                    </p>
                  </div>

                  <div
                    v-for="item in recentActivity"
                    :key="item.id"
                    class="flex items-center gap-3 p-3 rounded-xl bg-makoclaw-bg/30 border border-transparent hover:border-makoclaw-accent/20 hover:bg-makoclaw-surface/30 transition-all group cursor-pointer"
                  >
                    <div
                      class="w-9 h-9 rounded-xl flex items-center justify-center flex-shrink-0 shadow-md"
                      :class="item.iconBg"
                    >
                      <component
                        :is="item.icon"
                        class="w-4 h-4 text-white"
                      />
                    </div>
                    <div class="flex-1 min-w-0">
                      <div class="text-xs font-bold text-makoclaw-text truncate group-hover:text-makoclaw-accent transition-colors">
                        {{ item.title }}
                      </div>
                      <div class="flex items-center gap-2 mt-1">
                        <span class="text-[10px] text-makoclaw-text-secondary/60">{{ item.time }}</span>
                        <span class="w-1 h-1 rounded-full bg-makoclaw-border" />
                        <span class="text-[10px] font-bold text-makoclaw-accent uppercase">{{ item.type }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import { useAuthStore } from '../stores/authStore'
import taskService from '../services/taskService'
import advancedService from '../services/advancedService'
import { useToast } from '../composables/useToast'
import { Doughnut, Bar } from 'vue-chartjs'
import { Chart as ChartJS, Title, Tooltip, Legend, ArcElement, CategoryScale, LinearScale, BarElement } from 'chart.js'

// Simple Functional Icons
const IconTasks = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2' })]) }
const IconChat = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z' })]) }
const IconMessage = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z' })]) }
const IconProgress = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M13 10V3L4 14h7v7l9-11h-7z' })]) }
const IconPlus = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M12 4v16m8-8H4' })]) }
const IconHistory = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z' })]) }
const IconBrain = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z' })]) }

ChartJS.register(Title, Tooltip, Legend, ArcElement, CategoryScale, LinearScale, BarElement)

const authStore = useAuthStore()
const toast = useToast()
const loading = ref(true)
const tasks = ref([])
const sessions = ref([])
const metricsData = ref(null)

const reloadData = async () => {
  loading.value = true
  await loadDashboardData()
}

const formatNumber = (num) => {
  if (!num) return '0'
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

const statItems = computed(() => [
  { label: 'Total Tasks', value: stats.value.totalTasks, icon: IconTasks, iconBg: 'from-blue-500/40 to-indigo-600/40', trend: 12 },
  { label: 'In Progress', value: stats.value.inProgress, icon: IconProgress, iconBg: 'from-makoclaw-accent/40 to-blue-400/40', trend: 5 },
  { label: 'Chat Sessions', value: stats.value.chatSessions, icon: IconChat, iconBg: 'from-cyan-500/40 to-blue-500/40', trend: 8 },
  { label: 'Messages', value: formatNumber(stats.value.totalMessages), icon: IconMessage, iconBg: 'from-indigo-500/40 to-purple-600/40', trend: 15 },
])

const detailedMetrics = computed(() => {
  if (!metricsData.value) return []
  return [
    { label: 'LLM Calls', value: formatNumber(metricsData.value.llm_calls) },
    { label: 'Tool Calls', value: formatNumber(metricsData.value.tool_calls) },
    { label: 'Agent Runs', value: formatNumber(metricsData.value.agent_runs) },
    { label: 'Tokens', value: formatNumber(metricsData.value.llm_tokens_in + metricsData.value.llm_tokens_out) },
  ]
})

const quickActions = [
  { label: 'New Chat', to: '/chat', icon: IconPlus, color: 'bg-makoclaw-accent/30 shadow-lg shadow-makoclaw-accent/20' },
  { label: 'Tasks', to: '/tasks', icon: IconTasks, color: 'bg-indigo-500/30 shadow-lg shadow-indigo-500/20' },
  { label: 'History', to: '/history', icon: IconHistory, color: 'bg-cyan-500/30 shadow-lg shadow-cyan-500/20' },
  { label: 'Memory', to: '/memory', icon: IconBrain, color: 'bg-blue-600/30 shadow-lg shadow-blue-600/20' },
]

const recentActivity = computed(() => {
  const combined = [
    ...tasks.value.slice(0, 5).map(t => ({
      id: t.id,
      title: t.title,
      time: formatDate(t.created_at),
      type: 'Task',
      icon: IconTasks,
      iconBg: 'bg-indigo-500'
    })),
    ...sessions.value.slice(0, 5).map(s => ({
      id: s.session_id,
      title: sessionLabel(s.session_id),
      time: formatDate(s.last_activity),
      type: 'Chat',
      icon: IconChat,
      iconBg: 'bg-cyan-500'
    }))
  ]
  return combined.sort((a,b) => new Date(b.time) - new Date(a.time)).slice(0, 8)
})

const statusBreakdown = computed(() => {
  const counts = { backlog: 0, todo: 0, in_progress: 0, review: 0, done: 0 }
  tasks.value.forEach(t => {
    if (counts[t.status] !== undefined) counts[t.status]++
  })
  return [
    { status: 'backlog', label: 'Backlog', count: counts.backlog },
    { status: 'todo', label: 'To Do', count: counts.todo },
    { status: 'in_progress', label: 'Active', count: counts.in_progress },
    { status: 'review', label: 'Review', count: counts.review },
    { status: 'done', label: 'Done', count: counts.done }
  ]
})

const sessionLabel = (sessionId) => {
  if (sessionId.startsWith('web:chat:')) {
    return 'Chat ' + sessionId.replace('web:chat:', '').substring(0, 6).toUpperCase()
  }
  if (sessionId.startsWith('web:task:')) {
    return 'Task #' + sessionId.replace('web:task:', '')
  }
  return sessionId.substring(0, 12)
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
      position: 'bottom',
      labels: { color: 'rgba(156, 163, 175, 0.7)', font: { size: 10, weight: 'bold', family: 'Inter' }, padding: 15, usePointStyle: true }
    }
  },
  cutout: '72%'
}

const barOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    y: {
      beginAtZero: true,
      grid: { color: 'rgba(255,255,255,0.03)', drawBorder: false },
      ticks: { color: 'rgba(156, 163, 175, 0.5)', font: { size: 10 } }
    },
    x: {
      grid: { display: false },
      ticks: { color: 'rgba(156, 163, 175, 0.6)', font: { size: 10, weight: 'bold' } }
    }
  }
}

const modelChartData = computed(() => {
  if (!metricsData.value || !metricsData.value.llm_by_model) return { labels: [], datasets: [{ data: [] }] }
  const labels = Object.keys(metricsData.value.llm_by_model)
  const data = Object.values(metricsData.value.llm_by_model).map(m => m.calls)
  return {
    labels,
    datasets: [{
      backgroundColor: ['#3b82f6', '#22c55e', '#84cc16', '#06b6d4', '#0891b2', '#14b8a6'],
      borderColor: 'transparent',
      hoverOffset: 10,
      data
    }]
  }
})

const taskChartData = computed(() => ({
  labels: statusBreakdown.value.map(s => s.label),
  datasets: [{
    backgroundColor: ['#9ca3af', '#facc15', '#60a5fa', '#fb923c', '#4ade80'],
    data: statusBreakdown.value.map(s => s.count),
    borderRadius: 6,
    barPercentage: 0.6
  }]
}))

const loadDashboardData = async () => {
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
}

onMounted(loadDashboardData)
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.15);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.25);
}
</style>
