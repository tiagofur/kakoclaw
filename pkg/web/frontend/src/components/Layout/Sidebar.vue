<template>
  <div class="flex h-screen bg-picoclaw-bg text-picoclaw-text">
    <!-- Sidebar -->
    <div 
      class="transition-smooth"
      :class="[
        sidebarCollapsed ? 'w-16' : 'w-64',
        'bg-picoclaw-surface border-r border-picoclaw-border flex flex-col'
      ]"
    >
      <!-- Logo/Brand -->
      <div class="p-4 border-b border-picoclaw-border flex items-center justify-between">
        <div v-if="!sidebarCollapsed" class="font-bold text-lg text-picoclaw-accent">
          PicoClaw
        </div>
        <button
          @click="toggleSidebar"
          class="p-1 hover:bg-picoclaw-border rounded transition-smooth"
          title="Toggle sidebar"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
      </div>

      <!-- Navigation Items -->
      <nav class="flex-1 p-3 space-y-2">
        <router-link
          to="/dashboard"
          class="flex items-center gap-3 px-3 py-2 rounded hover:bg-picoclaw-border transition-smooth"
          @click="setActiveTab('chat')"
          :class="{ 'bg-picoclaw-border': activeTab === 'chat' }"
        >
          <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
          <span v-if="!sidebarCollapsed">Chat</span>
        </router-link>

        <router-link
          to="/dashboard"
          class="flex items-center gap-3 px-3 py-2 rounded hover:bg-picoclaw-border transition-smooth"
          @click="setActiveTab('tasks')"
          :class="{ 'bg-picoclaw-border': activeTab === 'tasks' }"
        >
          <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
          </svg>
          <span v-if="!sidebarCollapsed">Tasks</span>
        </router-link>
      </nav>

      <!-- User Section -->
      <div class="border-t border-picoclaw-border p-3 space-y-2">
        <button
          @click="toggleTheme"
          class="w-full flex items-center gap-3 px-3 py-2 rounded hover:bg-picoclaw-border transition-smooth text-sm"
          :title="theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
        >
          <svg v-if="theme === 'dark'" class="w-5 h-5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
            <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
          </svg>
          <svg v-else class="w-5 h-5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 2a1 1 0 011 1v2a1 1 0 11-2 0V3a1 1 0 011-1zM4.22 4.22a1 1 0 011.415 0l1.414 1.414a1 1 0 00-1.415 1.415L4.22 5.636a1 1 0 010-1.415zm11.314 0a1 1 0 011.415 0l1.414 1.414a1 1 0 11-1.415 1.415l-1.414-1.414a1 1 0 010-1.415zM4 10a1 1 0 011-1h2a1 1 0 110 2H5a1 1 0 01-1-1zm12 0a1 1 0 011-1h2a1 1 0 110 2h-2a1 1 0 01-1-1z" clip-rule="evenodd" />
          </svg>
          <span v-if="!sidebarCollapsed" class="text-xs">{{ theme === 'dark' ? 'Light' : 'Dark' }}</span>
        </button>

        <button
          @click="showProfileMenu = !showProfileMenu"
          class="w-full flex items-center gap-3 px-3 py-2 rounded hover:bg-picoclaw-border transition-smooth text-sm"
        >
          <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
          <span v-if="!sidebarCollapsed">{{ authStore.user?.username || 'Profile' }}</span>
        </button>

        <!-- Profile Menu -->
        <div v-if="showProfileMenu" class="absolute bottom-16 left-3 right-3 bg-picoclaw-border rounded p-2 space-y-1 text-sm">
          <button
            @click="showChangePasswordModal = true; showProfileMenu = false"
            class="w-full text-left px-3 py-2 hover:bg-picoclaw-bg rounded transition-smooth"
          >
            Change Password
          </button>
          <button
            @click="handleLogout"
            class="w-full text-left px-3 py-2 hover:bg-picoclaw-bg rounded transition-smooth text-picoclaw-error"
          >
            Logout
          </button>
        </div>
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Top Bar -->
      <div class="bg-picoclaw-surface border-b border-picoclaw-border h-14 flex items-center px-6 justify-between">
        <div class="text-lg font-semibold">
          {{ activeTab === 'chat' ? 'Chat' : 'Tasks' }}
        </div>
        <div class="flex items-center gap-4">
          <span v-if="authStore.sessionExpiry" class="text-xs text-picoclaw-text-secondary">
            Session expires: {{ formatExpiry() }}
          </span>
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-auto">
        <slot />
      </div>
    </div>

    <!-- Change Password Modal -->
    <ChangePasswordModal
      v-if="showChangePasswordModal"
      @close="showChangePasswordModal = false"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../../stores/authStore'
import { useUIStore } from '../../stores/uiStore'
import { useRouter } from 'vue-router'
import ChangePasswordModal from '../Auth/ChangePasswordModal.vue'

const router = useRouter()
const authStore = useAuthStore()
const uiStore = useUIStore()

const showProfileMenu = ref(false)
const showChangePasswordModal = ref(false)

const sidebarCollapsed = uiStore.sidebarCollapsed
const theme = uiStore.theme
const activeTab = uiStore.activeTab

const toggleSidebar = () => uiStore.toggleSidebar()
const toggleTheme = () => uiStore.toggleTheme()
const setActiveTab = (tab) => uiStore.setActiveTab(tab)

const handleLogout = async () => {
  authStore.logout()
  await router.push('/login')
}

const formatExpiry = () => {
  if (!authStore.sessionExpiry) return ''
  const expiry = new Date(authStore.sessionExpiry)
  const now = new Date()
  const diff = expiry - now
  
  if (diff <= 0) return 'Expired'
  
  const hours = Math.floor(diff / 3600000)
  const minutes = Math.floor((diff % 3600000) / 60000)
  
  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }
  return `${minutes}m`
}
</script>

<style scoped>
.router-link-active {
  @apply bg-picoclaw-accent/20 text-picoclaw-accent;
}
</style>
