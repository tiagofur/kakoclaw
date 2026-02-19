<template>
  <div class="h-full flex flex-col bg-kakoclaw-bg">
    <!-- Header -->
    <div class="flex-none p-4 border-b border-kakoclaw-border bg-kakoclaw-surface">
      <h2 class="text-xl font-bold bg-gradient-to-r from-kakoclaw-accent to-purple-500 bg-clip-text text-transparent">Dashboard</h2>
      <p class="text-sm text-kakoclaw-text-secondary mt-1">Overview of your KakoClaw workspace</p>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6 space-y-6 custom-scrollbar">

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-kakoclaw-accent"></div>
      </div>

      <template v-else>
        <!-- Stats Grid -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-4">
            <div class="text-sm text-kakoclaw-text-secondary">Total Tasks</div>
            <div class="text-2xl font-bold mt-1">{{ stats.totalTasks }}</div>
          </div>
          <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-4">
            <div class="text-sm text-kakoclaw-text-secondary">In Progress</div>
            <div class="text-2xl font-bold mt-1 text-blue-400">{{ stats.inProgress }}</div>
          </div>
          <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-4">
            <div class="text-sm text-kakoclaw-text-secondary">Chat Sessions</div>
            <div class="text-2xl font-bold mt-1 text-emerald-400">{{ stats.chatSessions }}</div>
          </div>
          <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-4">
            <div class="text-sm text-kakoclaw-text-secondary">Total Messages</div>
            <div class="text-2xl font-bold mt-1 text-purple-400">{{ stats.totalMessages }}</div>
          </div>
        </div>

        <!-- Task Status Breakdown -->
        <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5">
          <h3 class="font-semibold mb-4">Tasks by Status</h3>
          <div class="grid grid-cols-2 sm:grid-cols-5 gap-3">
            <div v-for="s in statusBreakdown" :key="s.status" class="text-center p-3 rounded-lg bg-kakoclaw-bg">
              <div class="text-lg font-bold" :class="s.color">{{ s.count }}</div>
              <div class="text-xs text-kakoclaw-text-secondary mt-1 capitalize">{{ s.label }}</div>
            </div>
          </div>
        </div>

        <!-- Two Column Layout -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">

          <!-- Recent Tasks -->
          <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5">
            <div class="flex items-center justify-between mb-4">
              <h3 class="font-semibold">Recent Tasks</h3>
              <router-link to="/tasks" class="text-sm text-kakoclaw-accent hover:underline">View all</router-link>
            </div>
            <div v-if="recentTasks.length === 0" class="text-sm text-kakoclaw-text-secondary py-4 text-center">No tasks yet</div>
            <div v-else class="space-y-2">
              <div
                v-for="task in recentTasks"
                :key="task.id"
                class="flex items-center gap-3 p-3 rounded-lg bg-kakoclaw-bg hover:bg-kakoclaw-border/50 transition-colors"
              >
                <span
                  class="w-2 h-2 rounded-full flex-shrink-0"
                  :class="{
                    'bg-gray-400': task.status === 'backlog',
                    'bg-yellow-400': task.status === 'todo',
                    'bg-blue-400': task.status === 'in_progress',
                    'bg-orange-400': task.status === 'review',
                    'bg-green-400': task.status === 'done'
                  }"
                ></span>
                <div class="flex-1 min-w-0">
                  <div class="text-sm font-medium truncate">{{ task.title }}</div>
                  <div class="text-xs text-kakoclaw-text-secondary capitalize">{{ task.status.replace('_', ' ') }}</div>
                </div>
                <div class="text-xs text-kakoclaw-text-secondary flex-shrink-0">
                  {{ formatDate(task.created_at) }}
                </div>
              </div>
            </div>
          </div>

          <!-- Recent Sessions -->
          <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5">
            <div class="flex items-center justify-between mb-4">
              <h3 class="font-semibold">Recent Conversations</h3>
              <router-link to="/history" class="text-sm text-kakoclaw-accent hover:underline">View all</router-link>
            </div>
            <div v-if="recentSessions.length === 0" class="text-sm text-kakoclaw-text-secondary py-4 text-center">No conversations yet</div>
            <div v-else class="space-y-2">
              <div
                v-for="session in recentSessions"
                :key="session.session_id"
                class="flex items-center gap-3 p-3 rounded-lg bg-kakoclaw-bg hover:bg-kakoclaw-border/50 transition-colors"
              >
                <svg v-if="session.session_id.startsWith('web:chat:')" class="w-5 h-5 flex-shrink-0 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                </svg>
                <svg v-else class="w-5 h-5 flex-shrink-0 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
                <div class="flex-1 min-w-0">
                  <div class="text-sm font-medium truncate">{{ sessionLabel(session.session_id) }}</div>
                  <div class="text-xs text-kakoclaw-text-secondary">{{ session.message_count }} messages</div>
                </div>
                <div class="text-xs text-kakoclaw-text-secondary flex-shrink-0">
                  {{ formatDate(session.last_activity) }}
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Quick Actions -->
        <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl p-5">
          <h3 class="font-semibold mb-4">Quick Actions</h3>
          <div class="flex flex-wrap gap-3">
            <router-link
              to="/chat"
              class="flex items-center gap-2 px-4 py-2 bg-kakoclaw-accent text-white rounded-lg hover:bg-kakoclaw-accent/90 transition-colors text-sm"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
              New Chat
            </router-link>
            <router-link
              to="/tasks"
              class="flex items-center gap-2 px-4 py-2 bg-kakoclaw-bg border border-kakoclaw-border text-kakoclaw-text rounded-lg hover:bg-kakoclaw-border/50 transition-colors text-sm"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
              New Task
            </router-link>
            <router-link
              to="/history"
              class="flex items-center gap-2 px-4 py-2 bg-kakoclaw-bg border border-kakoclaw-border text-kakoclaw-text rounded-lg hover:bg-kakoclaw-border/50 transition-colors text-sm"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
              Browse History
            </router-link>
            <router-link
              to="/memory"
              class="flex items-center gap-2 px-4 py-2 bg-kakoclaw-bg border border-kakoclaw-border text-kakoclaw-text rounded-lg hover:bg-kakoclaw-border/50 transition-colors text-sm"
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
import taskService from '../services/taskService'
import { useToast } from '../composables/useToast'

const toast = useToast()
const loading = ref(true)
const tasks = ref([])
const sessions = ref([])

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
    { status: 'backlog', label: 'Backlog', count: counts.backlog, color: 'text-gray-400' },
    { status: 'todo', label: 'To Do', count: counts.todo, color: 'text-yellow-400' },
    { status: 'in_progress', label: 'In Progress', count: counts.in_progress, color: 'text-blue-400' },
    { status: 'review', label: 'Review', count: counts.review, color: 'text-orange-400' },
    { status: 'done', label: 'Done', count: counts.done, color: 'text-green-400' }
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

onMounted(async () => {
  try {
    const [tasksData, sessionsData] = await Promise.all([
      taskService.fetchTasks(false),
      taskService.fetchChatSessions()
    ])
    tasks.value = tasksData.tasks || []
    sessions.value = sessionsData.sessions || []
  } catch (err) {
    console.error('Failed to load dashboard data:', err)
    toast.error('Failed to load dashboard data')
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 8px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 4px;
}
</style>
