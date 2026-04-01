<template>
  <aside
    class="glass-panel border-r border-makoclaw-border flex flex-col transition-all duration-300 ease-in-out h-full z-sidebar"
    :class="[
      uiStore.sidebarCollapsed ? 'w-20' : 'w-64',
      isMobile && !uiStore.sidebarCollapsed ? 'absolute inset-y-0 left-0 shadow-xl' : 'relative',
      isMobile && uiStore.sidebarCollapsed ? 'hidden' : 'flex'
    ]"
  >
    <!-- Logo/Brand -->
    <div class="h-16 flex items-center justify-between px-4 border-b border-makoclaw-border">
      <div
        v-if="!uiStore.sidebarCollapsed"
        class="font-bold text-xl flex items-center gap-2 group cursor-pointer"
      >
        <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-makoclaw-accent to-blue-600 flex items-center justify-center shadow-lg shadow-makoclaw-accent/20 transition-transform group-hover:rotate-12 duration-500">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-5 h-5 text-white"
            viewBox="0 0 24 24"
            fill="currentColor"
          >
            <path d="M20.8,10c-0.5-1-1.5-1.5-2.5-1.5c-0.5,0-1,0.1-1.5,0.4L14,10.4L11,8L4,6v2l5.5,1.5L7,13l-2.5-1L2,11.5v2l2,1l1.5,0.5l-0.5,2h2l1-2h5l1,2h2l-0.5-2l1.5-0.5l2-1v-2l-2.5,0.5L14,13l-1.5-3.5l2.8-1.5c0.2-0.1,0.4-0.2,0.7-0.2c0.5,0,0.9,0.3,1.1,0.7c0.2,0.4,0.2,0.9,0,1.3L16,12l1.5,0.5l1-2.2C19.1,9.3,19.2,8.1,20.8,10z M10,9.5c-0.3,0-0.5,0.2-0.5,0.5s0.2,0.5,0.5,0.5s0.5-0.2,0.5-0.5S10.3,9.5,10,9.5z" />
          </svg>
        </div>
        <span class="bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-makoclaw-accent bg-clip-text text-transparent">MakoClaw</span>
      </div>
      <div
        v-else
        class="w-8 h-8 rounded-lg bg-gradient-to-br from-makoclaw-accent to-blue-600 flex items-center justify-center mx-auto shadow-lg shadow-makoclaw-accent/20"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="w-5 h-5 text-white"
          viewBox="0 0 24 24"
          fill="currentColor"
        >
          <path d="M20.8,10c-0.5-1-1.5-1.5-2.5-1.5c-0.5,0-1,0.1-1.5,0.4L14,10.4L11,8L4,6v2l5.5,1.5L7,13l-2.5-1L2,11.5v2l2,1l1.5,0.5l-0.5,2h2l1-2h5l1,2h2l-0.5-2l1.5-0.5l2-1v-2l-2.5,0.5L14,13l-1.5-3.5l2.8-1.5c0.2-0.1,0.4-0.2,0.7-0.2c0.5,0,0.9,0.3,1.1,0.7c0.2,0.4,0.2,0.9,0,1.3L16,12l1.5,0.5l1-2.2C19.1,9.3,19.2,8.1,20.8,10z M10,9.5c-0.3,0-0.5,0.2-0.5,0.5s0.2,0.5,0.5,0.5s0.5-0.2,0.5-0.5S10.3,9.5,10,9.5z" />
        </svg>
      </div>
      <button
        class="p-1.5 hover:bg-makoclaw-bg rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
        title="Toggle sidebar"
        @click="uiStore.toggleSidebar()"
      >
        <svg
          class="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            v-if="!uiStore.sidebarCollapsed"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M11 19l-7-7 7-7m8 14l-7-7 7-7"
          />
          <path
            v-else
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 5l7 7-7 7M5 5l7 7-7 7"
          />
        </svg>
      </button>
    </div>

    <div class="flex-1 overflow-y-auto">
      <!-- Workspace Nav -->
      <nav class="p-3 space-y-1">
        <div
          v-if="!uiStore.sidebarCollapsed"
          class="px-3 py-2 text-[10px] font-bold uppercase tracking-[0.1em] text-makoclaw-text-secondary/50"
        >
          Workspace
        </div>

        <router-link
          to="/chat"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <div class="relative flex-shrink-0">
            <svg
              class="w-5 h-5 transition-transform group-hover:scale-110"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
              />
            </svg>
            <span
              v-if="chatStore.isWorking"
              class="absolute -top-1 -right-1 flex h-2.5 w-2.5"
            >
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-makoclaw-accent opacity-75" />
              <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-makoclaw-accent" />
            </span>
          </div>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm flex items-center gap-2"
          >
            Chat
            <span
              v-if="chatStore.isWorking"
              class="text-[10px] text-makoclaw-accent animate-pulse font-normal"
            >working...</span>
          </span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Chat{{ chatStore.isWorking ? ' (working...)' : '' }}
          </div>
        </router-link>

        <router-link
          to="/tasks"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01"
            />
          </svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Tasks</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Tasks
          </div>
        </router-link>

        <router-link
          to="/marketing"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Marketing</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Marketing
          </div>
        </router-link>

        <router-link
          v-if="uiStore.devStudioEnabled"
          to="/dev-studio"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Dev Studio</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Dev Studio
          </div>
        </router-link>

        <router-link
          to="/cron"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          /><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4.93 4.93l1.41 1.41M19.07 4.93l-1.41 1.41"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Cron Jobs</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Cron Jobs
          </div>
        </router-link>

        <router-link
          to="/workflows"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 6h16M4 6l4 6-4 6M20 6l-4 6 4 6M10 12h4"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Workflows</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Workflows
          </div>
        </router-link>
      </nav>

      <div class="px-3 py-2">
        <div class="h-px bg-makoclaw-border my-2" />
      </div>

      <!-- Resources Nav -->
      <nav class="p-3 pt-0 space-y-1">
        <div
          v-if="!uiStore.sidebarCollapsed"
          class="px-3 py-2 text-[10px] font-bold uppercase tracking-[0.1em] text-makoclaw-text-secondary/50"
        >
          Resources
        </div>

        <router-link
          to="/knowledge"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Knowledge</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Knowledge
          </div>
        </router-link>

        <router-link
          to="/files"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Files</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Files
          </div>
        </router-link>

        <router-link
          to="/skills"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 10V3L4 14h7v7l9-11h-7z"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Skills</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Skills
          </div>
        </router-link>

        <router-link
          to="/agents"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Agents</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Agents
          </div>
        </router-link>

        <router-link
          to="/mcp"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >MCP</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            MCP Servers
          </div>
        </router-link>
      </nav>

      <div class="px-3 py-2">
        <div class="h-px bg-makoclaw-border my-2" />
      </div>

      <!-- Insights Nav -->
      <nav class="p-3 pt-0 space-y-1">
        <div
          v-if="!uiStore.sidebarCollapsed"
          class="px-3 py-2 text-[10px] font-bold uppercase tracking-[0.1em] text-makoclaw-text-secondary/50"
        >
          Insights
        </div>

        <router-link
          to="/dashboard"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"
            />
          </svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Dashboard</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Dashboard
          </div>
        </router-link>

        <router-link
          to="/history"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >History</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            History
          </div>
        </router-link>

        <router-link
          to="/memory"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M19.428 15.428a2 2 0 00-1.022-.547l-2.384-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Memory</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Memory
          </div>
        </router-link>

        <router-link
          to="/metrics"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M16 8v8m-4-5v5m-4-2v2m-2 4h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Metrics</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Metrics
          </div>
        </router-link>

        <router-link
          to="/reports"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Reports</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Reports
          </div>
        </router-link>
      </nav>

      <div class="px-3 py-2">
        <div class="h-px bg-makoclaw-border my-2" />
      </div>

      <!-- System Nav -->
      <nav class="p-3 pt-0 space-y-1">
        <router-link
          to="/settings"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative"
          active-class="bg-makoclaw-accent/15 text-makoclaw-accent shadow-sm shadow-makoclaw-accent/5"
          inactive-class="text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          @click="closeMobileSidebar"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
          /><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Settings</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Settings
          </div>
        </router-link>

        <a
          href="/help"
          target="_blank"
          rel="noopener"
          class="flex items-center gap-3 px-3 py-2.5 rounded-xl transition-all duration-300 group relative text-makoclaw-text-secondary hover:bg-makoclaw-accent/5 hover:text-makoclaw-text hover:translate-x-1"
          title="Help & Guides"
        >
          <svg
            class="w-5 h-5 flex-shrink-0 transition-transform group-hover:scale-110"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          /></svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="font-medium whitespace-nowrap text-sm"
          >Help</span>
          <svg
            v-if="!uiStore.sidebarCollapsed"
            class="w-3 h-3 ml-auto opacity-50"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          ><path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
          /></svg>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Help
          </div>
        </a>
      </nav>


      <!-- User Section -->
      <div class="border-t border-makoclaw-border p-3 space-y-2">
        <button
          class="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-makoclaw-bg transition-colors text-sm group relative"
          :title="uiStore.theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
          @click="uiStore.toggleTheme()"
        >
          <svg
            v-if="uiStore.theme === 'dark'"
            class="w-5 h-5 flex-shrink-0 text-yellow-400 transition-transform duration-300"
            :class="{ 'rotate-180': uiStore.theme === 'dark' }"
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
          </svg>
          <svg
            v-else
            class="w-5 h-5 flex-shrink-0 text-orange-400 transition-transform duration-300"
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path
              fill-rule="evenodd"
              d="M10 2a1 1 0 011 1v2a1 1 0 11-2 0V3a1 1 0 011-1zM4.22 4.22a1 1 0 011.415 0l1.414 1.414a1 1 0 00-1.415 1.415L4.22 5.636a1 1 0 010-1.415zm11.314 0a1 1 0 011.415 0l1.414 1.414a1 1 0 11-1.415 1.415l-1.414-1.414a1 1 0 010-1.415zM4 10a1 1 0 011-1h2a1 1 0 110 2H5a1 1 0 01-1-1zm12 0a1 1 0 011-1h2a1 1 0 110 2h-2a1 1 0 01-1-1z"
              clip-rule="evenodd"
            />
          </svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="whitespace-nowrap"
          >{{ uiStore.theme === 'dark' ? 'Light Mode' : 'Dark Mode' }}</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Toggle Theme
          </div>
        </button>

        <button
          v-if="uiStore.canInstallPwa"
          class="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-makoclaw-accent/10 text-makoclaw-accent transition-colors text-sm group relative"
          title="Install MakoClaw app"
          @click="uiStore.installPwa()"
        >
          <svg
            class="w-5 h-5 flex-shrink-0"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
            />
          </svg>
          <span
            v-if="!uiStore.sidebarCollapsed"
            class="whitespace-nowrap font-medium"
          >Install App</span>
          <div
            v-if="uiStore.sidebarCollapsed"
            class="tooltip"
          >
            Install App
          </div>
        </button>

        <div class="relative">
          <button
            class="w-full flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-makoclaw-bg transition-colors text-sm group"
            @click="showProfileMenu = !showProfileMenu"
          >
            <div class="w-8 h-8 rounded-full bg-makoclaw-accent flex items-center justify-center text-white font-bold flex-shrink-0">
              {{ userInitials }}
            </div>
            <div
              v-if="!uiStore.sidebarCollapsed"
              class="flex flex-col text-left overflow-hidden"
            >
              <span class="font-medium truncate">{{ authStore.user?.username || 'User' }}</span>
              <span class="text-xs text-makoclaw-text-secondary">View Profile</span>
            </div>
          </button>

          <!-- Profile Menu -->
          <div
            v-if="showProfileMenu"
            class="absolute bottom-full left-0 mb-2 w-full min-w-[12rem] bg-makoclaw-surface border border-makoclaw-border rounded-lg shadow-lg p-1 z-dropdown"
          >
            <button
              class="w-full text-left px-3 py-2 hover:bg-makoclaw-bg rounded transition-colors text-sm flex items-center gap-2"
              @click="showChangePasswordModal = true; showProfileMenu = false"
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
                d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
              /></svg>
              Change Password
            </button>
            <div class="h-px bg-makoclaw-border my-1" />
            <button
              class="w-full text-left px-3 py-2 hover:bg-red-500/10 text-red-500 rounded transition-colors text-sm flex items-center gap-2"
              @click="handleLogout"
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
                d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
              /></svg>
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
  </aside>

  <!-- Mobile Overlay - Teleported to body to escape aside's stacking context -->
  <Teleport to="body">
    <div
      v-if="isMobile && !uiStore.sidebarCollapsed"
      class="fixed inset-0 bg-black/50 z-overlay-backdrop pointer-events-auto"
      @click.stop="closeMobileSidebar"
    />
  </Teleport>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '../../stores/authStore'
import { useUIStore } from '../../stores/uiStore'
import { useChatStore } from '../../stores/chatStore'
import { useRouter } from 'vue-router'
import ChangePasswordModal from '../Auth/ChangePasswordModal.vue'

const router = useRouter()
const authStore = useAuthStore()
const uiStore = useUIStore()
const chatStore = useChatStore()

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

const closeMobileSidebar = () => {
  if (isMobile.value && !uiStore.sidebarCollapsed) {
    uiStore.toggleSidebar()
  }
}
</script>

<style scoped>
/* Tooltip styles for collapsed sidebar */
.tooltip {
  @apply absolute left-full ml-3 px-3 py-1.5 bg-makoclaw-surface/95 backdrop-blur-xl border border-makoclaw-border/50 rounded-lg text-xs font-medium text-makoclaw-text whitespace-nowrap shadow-lg ring-1 ring-white/10 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-dropdown;
}

.tooltip::before {
  content: '';
  @apply absolute right-full top-1/2 -translate-y-1/2 border-4 border-transparent border-r-makoclaw-border/50;
}

/* Scrollbar styling */
::-webkit-scrollbar {
  width: 4px;
}
::-webkit-scrollbar-track {
  background: transparent;
}
::-webkit-scrollbar-thumb {
  @apply bg-makoclaw-border/30 rounded-full;
}
::-webkit-scrollbar-thumb:hover {
  @apply bg-makoclaw-border/50;
}
</style>
