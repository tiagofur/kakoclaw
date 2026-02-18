<template>
  <aside 
    class="glass-panel border-r border-picoclaw-border flex flex-col transition-all duration-300 ease-in-out h-full z-30"
    :class="[
      uiStore.sidebarCollapsed ? 'w-16' : 'w-64',
      isMobile && !uiStore.sidebarCollapsed ? 'absolute inset-y-0 left-0 shadow-xl' : 'relative',
      isMobile && uiStore.sidebarCollapsed ? 'hidden' : 'flex'
    ]"
  >
    <!-- Logo/Brand -->
    <div class="h-16 flex items-center justify-between px-4 border-b border-picoclaw-border">
      <div v-if="!uiStore.sidebarCollapsed" class="font-bold text-xl bg-gradient-to-r from-picoclaw-accent to-purple-500 bg-clip-text text-transparent truncate">
        PicoClaw
      </div>
      <button
        @click="uiStore.toggleSidebar()"
        class="p-1.5 hover:bg-picoclaw-bg rounded-lg text-picoclaw-text-secondary hover:text-picoclaw-text transition-colors"
        title="Toggle sidebar"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path v-if="!uiStore.sidebarCollapsed" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
          <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
        </svg>
      </button>
    </div>

    <div class="flex-1 overflow-y-auto">
      <!-- Navigation Items -->
      <nav class="p-3 space-y-1">
        <router-link
        to="/dashboard"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
        </svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Dashboard</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">
          Dashboard
        </div>
      </router-link>

      <router-link
        to="/chat"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
        </svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Chat</span>
        <!-- Tooltip for collapsed state -->
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">
          Chat
        </div>
      </router-link>

      <router-link
        to="/tasks"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
        </svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Tasks</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">
          Tasks
        </div>
        </router-link>
      </nav>
      
      <div class="px-3 py-2">
        <div class="h-px bg-picoclaw-border my-2"></div>
      </div>

      <!-- Tools Nav -->
      <nav class="p-3 pt-0 space-y-1">
        <div v-if="!uiStore.sidebarCollapsed" class="px-3 py-1 text-[10px] font-semibold uppercase tracking-wider text-picoclaw-text-secondary/60">Tools</div>

      <router-link
        to="/skills"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Skills</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">Skills</div>
      </router-link>

      <router-link
        to="/cron"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.93 4.93l1.41 1.41M19.07 4.93l-1.41 1.41" /></svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Cron Jobs</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">Cron Jobs</div>
      </router-link>

      <router-link
        to="/channels"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.858 15.355-5.858 21.213 0" /></svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Channels</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">Channels</div>
      </router-link>

      <router-link
        to="/files"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" /></svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Files</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">Files</div>
      </router-link>

      <router-link
        to="/knowledge"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" /></svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Knowledge</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">Knowledge</div>
      </router-link>

      <router-link
        to="/mcp"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" /></svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">MCP</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">MCP Servers</div>
      </router-link>

      <router-link
        to="/workflows"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 6l4 6-4 6M20 6l-4 6 4 6M10 12h4" /></svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Workflows</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">Workflows</div>
        </router-link>
      </nav>

      <div class="px-3 py-2">
        <div class="h-px bg-picoclaw-border my-2"></div>
      </div>

      <!-- Secondary Nav -->
      <nav class="p-3 pt-0 space-y-1">
      <router-link
        to="/history"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">History</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">History</div>
      </router-link>

      <router-link
        to="/memory"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" /></svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Memory</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">Memory</div>
      </router-link>

      <router-link
        to="/reports"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Reports</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">Reports</div>
      </router-link>

      <router-link
        to="/metrics"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 8v8m-4-5v5m-4-2v2m-2 4h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" /></svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Metrics</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">Metrics</div>
      </router-link>

      <router-link
        to="/settings"
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-picoclaw-bg transition-colors group relative"
        active-class="bg-picoclaw-accent/10 text-picoclaw-accent hover:bg-picoclaw-accent/20"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
        <span v-if="!uiStore.sidebarCollapsed" class="font-medium whitespace-nowrap">Settings</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">Settings</div>
        </router-link>
      </nav>


      <!-- User Section -->
      <div class="border-t border-picoclaw-border p-3 space-y-2">
      <button
        @click="uiStore.toggleTheme()"
        class="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-picoclaw-bg transition-colors text-sm group relative"
        :title="uiStore.theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
      >
        <svg v-if="uiStore.theme === 'dark'" class="w-5 h-5 flex-shrink-0 text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
          <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
        </svg>
        <svg v-else class="w-5 h-5 flex-shrink-0 text-orange-400" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M10 2a1 1 0 011 1v2a1 1 0 11-2 0V3a1 1 0 011-1zM4.22 4.22a1 1 0 011.415 0l1.414 1.414a1 1 0 00-1.415 1.415L4.22 5.636a1 1 0 010-1.415zm11.314 0a1 1 0 011.415 0l1.414 1.414a1 1 0 11-1.415 1.415l-1.414-1.414a1 1 0 010-1.415zM4 10a1 1 0 011-1h2a1 1 0 110 2H5a1 1 0 01-1-1zm12 0a1 1 0 011-1h2a1 1 0 110 2h-2a1 1 0 01-1-1z" clip-rule="evenodd" />
        </svg>
        <span v-if="!uiStore.sidebarCollapsed" class="whitespace-nowrap">{{ uiStore.theme === 'dark' ? 'Light Mode' : 'Dark Mode' }}</span>
        <div v-if="uiStore.sidebarCollapsed" class="absolute left-full ml-2 px-2 py-1 bg-gray-800 text-white text-xs rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap">
          Toggle Theme
        </div>
      </button>

      <div class="relative">
        <button
          @click="showProfileMenu = !showProfileMenu"
          class="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-picoclaw-bg transition-colors text-sm group"
        >
          <div class="w-8 h-8 rounded-full bg-picoclaw-accent flex items-center justify-center text-white font-bold flex-shrink-0">
            {{ userInitials }}
          </div>
          <div v-if="!uiStore.sidebarCollapsed" class="flex flex-col text-left overflow-hidden">
             <span class="font-medium truncate">{{ authStore.user?.username || 'User' }}</span>
             <span class="text-xs text-picoclaw-text-secondary">View Profile</span>
          </div>
        </button>

        <!-- Profile Menu -->
        <div v-if="showProfileMenu" class="absolute bottom-full left-0 mb-2 w-full min-w-[12rem] bg-picoclaw-surface border border-picoclaw-border rounded-lg shadow-lg p-1 z-50">
          <button
            @click="showChangePasswordModal = true; showProfileMenu = false"
            class="w-full text-left px-3 py-2 hover:bg-picoclaw-bg rounded transition-colors text-sm flex items-center gap-2"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" /></svg>
            Change Password
          </button>
          <div class="h-px bg-picoclaw-border my-1"></div>
          <button
            @click="handleLogout"
            class="w-full text-left px-3 py-2 hover:bg-red-500/10 text-red-500 rounded transition-colors text-sm flex items-center gap-2"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" /></svg>
            Logout
          </button>
        </div>
      </div>
      </div>
    </div>

    <!-- Change Password Modal -->
    <ChangePasswordModal
      v-if="showChangePasswordModal"
      @close="showChangePasswordModal = false"
    />
    
    <!-- Mobile Overlay -->
    <div 
        v-if="isMobile && !uiStore.sidebarCollapsed" 
        class="fixed inset-0 bg-black/50 z-20"
        @click="uiStore.toggleSidebar()"
    ></div>
  </aside>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '../../stores/authStore'
import { useUIStore } from '../../stores/uiStore'
import { useRouter } from 'vue-router'
import ChangePasswordModal from '../Auth/ChangePasswordModal.vue'

const router = useRouter()
const authStore = useAuthStore()
const uiStore = useUIStore()

const showProfileMenu = ref(false)
const showChangePasswordModal = ref(false)
const isMobile = ref(false)

const userInitials = computed(() => {
    const name = authStore.user?.username || 'U'
    return name.substring(0, 2).toUpperCase()
})

const checkMobile = () => {
    isMobile.value = window.innerWidth < 768
    if (isMobile.value && !uiStore.sidebarCollapsed && window.innerWidth < 768) {
        uiStore.sidebarCollapsed = true // Default to collapsed on mobile
    }
}

onMounted(() => {
    checkMobile()
    window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
    window.removeEventListener('resize', checkMobile)
})

const handleLogout = async () => {
  authStore.logout()
  await router.push('/login')
}
</script>

<style scoped>
/* Add any specific styles here if tailwind isn't enough */
</style>
