<template>
  <div class="h-full flex flex-col bg-makoclaw-bg relative overflow-hidden">
    <!-- Decorative background blobs -->
    <div class="absolute -top-24 -right-24 w-96 h-96 bg-makoclaw-accent/10 blur-[100px] rounded-full pointer-events-none animate-pulse"></div>
    <div class="absolute top-1/2 -left-24 w-72 h-72 bg-blue-500/5 blur-[80px] rounded-full pointer-events-none" style="animation: float 20s infinite ease-in-out"></div>

    <!-- Header -->
    <div class="flex-none p-6 glass-sticky z-20">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-2xl font-black tracking-tight bg-gradient-to-r from-makoclaw-accent via-blue-400 to-cyan-400 bg-clip-text text-transparent">
            Workspace Overview
          </h2>
          <p class="text-xs font-medium text-makoclaw-text-secondary/70 mt-1 flex items-center gap-2">
            <span class="w-1.5 h-1.5 rounded-full bg-makoclaw-success animate-pulse"></span>
            System operational • {{ authStore.user?.username || 'Commander' }}
          </p>
        </div>
        <div class="flex gap-2">
           <button @click="reloadData" class="p-2 rounded-xl bg-makoclaw-surface border border-makoclaw-border hover:border-makoclaw-accent/50 transition-all active:scale-90 group">
             <svg class="w-5 h-5 text-makoclaw-text-secondary group-hover:text-makoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
               <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
             </svg>
           </button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-4 md:p-8 space-y-8 custom-scrollbar relative z-10">

      <!-- Loading Skeleton -->
      <div v-if="loading" class="space-y-8 animate-pulse">
        <div class="h-48 bg-makoclaw-surface/50 rounded-3xl border border-makoclaw-border"></div>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div v-for="i in 4" :key="i" class="h-32 bg-makoclaw-surface/50 rounded-2xl border border-makoclaw-border"></div>
        </div>
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div class="h-80 bg-makoclaw-surface/50 rounded-3xl border border-makoclaw-border"></div>
          <div class="h-80 bg-makoclaw-surface/50 rounded-3xl border border-makoclaw-border"></div>
        </div>
      </div>

      <template v-else>
        <!-- Welcome banner -->
        <div class="relative overflow-hidden bg-gradient-to-br from-makoclaw-accent via-blue-600 to-indigo-700 rounded-[2rem] p-8 md:p-12 shadow-2xl shadow-makoclaw-accent/20 group animate-fade-in-up">
          <!-- Animated Background mesh -->
          <div class="absolute inset-0 opacity-20 group-hover:opacity-30 transition-opacity duration-1000">
            <div class="absolute top-0 left-0 w-full h-full bg-[radial-gradient(circle_at_50%_50%,rgba(255,255,255,0.1)_0%,transparent_50%)]"></div>
          </div>
          
          <div class="absolute top-0 right-0 p-8 transform rotate-12 opacity-10 transition-transform group-hover:rotate-45 group-hover:scale-125 duration-1000 hidden md:block">
            <svg class="w-64 h-64" fill="currentColor" viewBox="0 0 24 24"><path d="M13 10V3L4 14h7v7l9-11h-7z"/></svg>
          </div>

          <div class="relative z-10">
            <span class="inline-block px-3 py-1 bg-white/10 backdrop-blur-md rounded-full text-[10px] font-bold text-white uppercase tracking-widest mb-4">Command Center</span>
            <h3 class="text-3xl md:text-5xl font-black text-white leading-tight">Welcome back,<br/>{{ authStore.user?.username || 'Commander' }}</h3>
            <p class="text-white/70 mt-4 text-sm md:text-lg max-w-xl leading-relaxed">Your AI fleet is standing by. All systems are nominal across 10+ channels. What are we building today?</p>
            
            <div class="mt-8 flex flex-wrap gap-4">
               <router-link to="/chat" class="px-6 py-3 bg-white text-makoclaw-accent rounded-xl font-bold shadow-xl hover:shadow-white/20 transition-all hover:-translate-y-1 active:scale-95 text-sm uppercase tracking-wider">Launch New Session</router-link>
               <router-link to="/tasks" class="px-6 py-3 bg-white/10 backdrop-blur-md border border-white/20 text-white rounded-xl font-bold hover:bg-white/20 transition-all active:scale-95 text-sm uppercase tracking-wider">Tasks Dashboard</router-link>
            </div>
          </div>
        </div>

        <!-- Stats Grid -->
        <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6">
          <div v-for="(stat, idx) in statItems" :key="stat.label" 
               class="glass-panel rounded-3xl p-6 transition-all duration-500 hover:shadow-2xl hover:shadow-makoclaw-accent/10 hover:-translate-y-2 group animate-fade-in-up"
               :style="{ 'animation-delay': (200 + idx * 100) + 'ms' }">
            <div class="flex justify-between items-start">
              <div class="p-3 rounded-2xl bg-gradient-to-br transition-colors duration-500" :class="stat.iconBg">
                <component :is="stat.icon" class="w-6 h-6 text-white" />
              </div>
              <div v-if="stat.trend" class="px-2 py-1 rounded-full bg-makoclaw-success/10 text-[10px] font-bold text-makoclaw-success border border-makoclaw-success/20">
                +{{ stat.trend }}%
              </div>
            </div>
            <div class="mt-6">
              <div class="text-[10px] font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/60">{{ stat.label }}</div>
              <div class="text-4xl font-black mt-1 bg-gradient-to-br from-makoclaw-text to-makoclaw-text-secondary bg-clip-text text-transparent">{{ stat.value }}</div>
            </div>
          </div>
        </div>

        <!-- Main Dashboard Section -->
        <div class="grid grid-cols-1 xl:grid-cols-3 gap-8">
          
          <!-- Activity Charts (Left Area) -->
          <div class="xl:col-span-2 space-y-8">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
               <div class="glass-panel rounded-[2rem] p-8 flex flex-col animate-fade-in-up" style="animation-delay: 600ms">
                <div class="flex items-center justify-between mb-8">
                  <h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70">Model Intelligence</h3>
                  <div class="w-2 h-2 rounded-full bg-blue-500 animate-pulse"></div>
                </div>
                <div class="w-full relative flex-1 min-h-[300px] flex items-center justify-center">
                   <Doughnut v-if="modelChartData.labels.length > 0" :data="modelChartData" :options="chartOptions" />
                   <div v-else class="text-sm font-medium text-makoclaw-text-secondary/50 italic">A waiting data stream...</div>
                </div>
              </div>
              
               <div class="glass-panel rounded-[2rem] p-8 flex flex-col animate-fade-in-up" style="animation-delay: 750ms">
                 <div class="flex items-center justify-between mb-8">
                  <h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70">Operations Status</h3>
                  <div class="w-2 h-2 rounded-full bg-emerald-500"></div>
                </div>
                 <div class="w-full relative flex-1 min-h-[300px] flex items-center justify-center">
                    <Bar v-if="taskChartData.labels.length > 0" :data="taskChartData" :options="barOptions" />
                    <div v-else class="text-sm font-medium text-makoclaw-text-secondary/50 italic">A waiting data stream...</div>
                 </div>
              </div>
            </div>

            <!-- Detailed Stats Table-like grid -->
            <div class="glass-panel rounded-[2rem] p-8 animate-fade-in-up" style="animation-delay: 900ms">
              <h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70 mb-8 px-2">Live System Metrics</h3>
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-8">
                <div v-for="metric in detailedMetrics" :key="metric.label" class="p-4 rounded-3xl hover:bg-makoclaw-surface/50 transition-colors group">
                  <div class="text-[10px] font-bold text-makoclaw-text-secondary/50 uppercase tracking-widest mb-1 group-hover:text-makoclaw-accent transition-colors">{{ metric.label }}</div>
                  <div class="text-2xl font-black text-makoclaw-text">{{ metric.value }}</div>
                  <div class="h-1 w-8 bg-makoclaw-accent/20 rounded-full mt-2 group-hover:w-full transition-all duration-500"></div>
                </div>
              </div>
            </div>
          </div>

          <!-- Feed & Actions (Right Sidebar Area) -->
          <div class="space-y-8">
            <!-- Launchpad (Replaces Quick Actions) -->
            <div class="glass-panel rounded-[2rem] p-8 animate-fade-in-up" style="animation-delay: 400ms">
              <h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70 mb-6 px-2">Action Launchpad</h3>
              <div class="grid grid-cols-2 gap-3">
                 <router-link v-for="action in launchpadActions" :key="action.label" :to="action.to"
                   class="flex flex-col items-center justify-center p-4 rounded-[1.5rem] border border-makoclaw-border hover:border-makoclaw-accent/40 bg-makoclaw-surface/30 hover:bg-makoclaw-accent/5 transition-all group active:scale-95"
                 >
                   <div class="w-10 h-10 rounded-xl flex items-center justify-center mb-3 transition-transform group-hover:scale-110 group-hover:rotate-3" :class="action.color">
                     <component :is="action.icon" class="w-5 h-5 text-white" />
                   </div>
                   <span class="text-[11px] font-black uppercase text-makoclaw-text-secondary/80 group-hover:text-makoclaw-accent">{{ action.label }}</span>
                 </router-link>
              </div>
            </div>

            <!-- Recent Activity Feed -->
            <div class="glass-panel rounded-[2rem] p-8 animate-fade-in-up flex flex-col h-[500px]" style="animation-delay: 550ms">
              <div class="flex items-center justify-between mb-8 px-2">
                <h3 class="text-xs font-black uppercase tracking-[0.2em] text-makoclaw-text-secondary/70">Recent Pulse</h3>
                <router-link to="/history" class="text-[10px] font-bold text-makoclaw-accent uppercase hover:underline tracking-widest">All History</router-link>
              </div>
              
              <div class="flex-1 overflow-auto custom-scrollbar space-y-3 px-1">
                <div v-if="recentActivity.length === 0" class="flex flex-col items-center justify-center py-20 text-center">
                  <div class="w-12 h-12 rounded-full border-2 border-dashed border-makoclaw-border flex items-center justify-center mb-4">
                    <svg class="w-6 h-6 text-makoclaw-text-secondary/30" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                  </div>
                  <p class="text-xs font-bold text-makoclaw-text-secondary/40 uppercase tracking-widest leading-relaxed">No data detected<br/>in current cycle</p>
                </div>
                
                <div v-for="item in recentActivity" :key="item.id" 
                     class="flex items-center gap-4 p-4 rounded-2xl bg-makoclaw-surface/40 border border-makoclaw-border/30 hover:border-makoclaw-accent/30 transition-all group cursor-pointer">
                  <div class="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0 transition-colors" :class="item.iconBg">
                    <component :is="item.icon" class="w-5 h-5 text-white" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="text-[11px] font-bold text-makoclaw-text truncate group-hover:text-makoclaw-accent transition-colors">{{ item.title }}</div>
                    <div class="flex items-center gap-2 mt-1">
                      <span class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text-secondary/50">{{ item.time }}</span>
                      <span class="w-1 h-1 rounded-full bg-makoclaw-border"></span>
                      <span class="text-[10px] font-bold text-makoclaw-accent">{{ item.type }}</span>
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

// Simple Functional Icons for modern look
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
  { label: 'Intelligence Fleet', value: stats.value.totalTasks, icon: IconTasks, iconBg: 'from-blue-500 to-indigo-600', trend: 12 },
  { label: 'Active Missions', value: stats.value.inProgress, icon: IconProgress, iconBg: 'from-makoclaw-accent to-blue-400', trend: 5 },
  { label: 'Neural Links', value: stats.value.chatSessions, icon: IconChat, iconBg: 'from-cyan-500 to-blue-500', trend: 8 },
  { label: 'Total Cognition', value: formatNumber(stats.value.totalMessages), icon: IconMessage, iconBg: 'from-indigo-500 to-purple-600', trend: 15 },
])

const detailedMetrics = computed(() => {
  if (!metricsData.value) return []
  return [
    { label: 'LLM Calls', value: formatNumber(metricsData.value.llm_calls) },
    { label: 'Tool Triggers', value: formatNumber(metricsData.value.tool_calls) },
    { label: 'Agent Pulses', value: formatNumber(metricsData.value.agent_runs) },
    { label: 'Bits Processed', value: formatNumber(metricsData.value.llm_tokens_in + metricsData.value.llm_tokens_out) },
  ]
})

const launchpadActions = [
  { label: 'Initiate Link', to: '/chat', icon: IconPlus, color: 'bg-makoclaw-accent shadow-lg shadow-makoclaw-accent/20' },
  { label: 'New Prototype', to: '/tasks', icon: IconTasks, color: 'bg-indigo-500 shadow-lg shadow-indigo-500/20' },
  { label: 'Sync Records', to: '/history', icon: IconHistory, color: 'bg-cyan-500 shadow-lg shadow-cyan-500/20' },
  { label: 'Refine Core', to: '/memory', icon: IconBrain, color: 'bg-blue-600 shadow-lg shadow-blue-600/20' },
]

const recentActivity = computed(() => {
  const combined = [
    ...tasks.value.slice(0, 5).map(t => ({
      id: t.id,
      title: t.title,
      time: formatDate(t.created_at),
      type: 'TASK',
      icon: IconTasks,
      iconBg: 'bg-indigo-500'
    })),
    ...sessions.value.slice(0, 5).map(s => ({
      id: s.session_id,
      title: sessionLabel(s.session_id),
      time: formatDate(s.last_activity),
      type: 'CHAT',
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
    { status: 'backlog', label: 'Backlog', count: counts.backlog, color: 'text-makoclaw-text-secondary' },
    { status: 'todo', label: 'To Do', count: counts.todo, color: 'text-makoclaw-warning' },
    { status: 'in_progress', label: 'In Progress', count: counts.in_progress, color: 'text-makoclaw-accent' },
    { status: 'review', label: 'Review', count: counts.review, color: 'text-makoclaw-warning' },
    { status: 'done', label: 'Done', count: counts.done, color: 'text-makoclaw-success' }
  ]
})

const sessionLabel = (sessionId) => {
  if (sessionId.startsWith('web:chat:')) {
    return 'Hyper-Chat ' + sessionId.replace('web:chat:', '').substring(0, 6).toUpperCase()
  }
  if (sessionId.startsWith('web:task:')) {
    return 'Unit #' + sessionId.replace('web:task:', '')
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
      labels: { color: 'rgba(156, 163, 175, 0.7)', font: { size: 10, weight: 'bold', family: 'Inter' }, padding: 20, usePointStyle: true }
    }
  },
  cutout: '75%'
}

const barOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    y: {
      beginAtZero: true,
      grid: { color: 'rgba(255,255,255,0.02)', drawBorder: false },
      ticks: { color: 'rgba(156, 163, 175, 0.4)', font: { size: 10 } }
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
      hoverOffset: 15,
      data
    }]
  }
})

const taskChartData = computed(() => ({
  labels: statusBreakdown.value.map(s => s.label),
  datasets: [{
    backgroundColor: ['#9ca3af', '#facc15', '#60a5fa', '#fb923c', '#4ade80'],
    data: statusBreakdown.value.map(s => s.count),
    borderRadius: 8,
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
@keyframes float {
  0%, 100% { transform: translateY(0) translateX(0); }
  50% { transform: translateY(-20px) translateX(10px); }
}

@keyframes fade-in-up {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fade-in-up {
  animation: fade-in-up 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  opacity: 0;
}

/* Custom scrollbar override for more subtle look */
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.1);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.2);
}
</style>
