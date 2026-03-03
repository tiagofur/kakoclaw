<template>
  <div class="h-full flex flex-col bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-orange-500/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-amber-500/20 via-transparent to-transparent" />
    </div>


    <!-- Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <!-- Title Row -->
        <div class="flex items-center gap-3">
          <!-- Sidebar toggle -->
          <button
            class="p-2 min-h-[40px] min-w-[40px] rounded-xl bg-makoclaw-surface/50 border border-makoclaw-border/50 hover:bg-makoclaw-surface-hover hover:border-orange-400/30 transition-all flex items-center justify-center active:scale-95"
            title="Toggle Sidebar"
            @click="toggleSidebar"
          >
            <svg
              class="w-4 h-4 text-makoclaw-text-secondary transition-transform duration-500"
              :class="{'rotate-180': isSidebarCollapsed}"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M11 19l-7-7 7-7m8 14l-7-7 7-7"
              />
            </svg>
          </button>

          <!-- Icon Container -->
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-orange-500/20 to-amber-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-orange-500/10">
            <svg
              class="w-5 h-5 sm:w-6 sm:h-6 text-orange-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          </div>

          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-orange-400 bg-clip-text text-transparent">
              Chat History
            </h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5 hidden sm:block">
              Browse, search, and manage past conversations
            </p>
          </div>

          <!-- Action Buttons -->
          <div class="flex items-center gap-2 flex-shrink-0">
            <div
              ref="exportDropdownRef"
              class="relative"
            >
              <button
                class="p-2.5 min-h-[40px] min-w-[40px] rounded-xl bg-makoclaw-surface/50 border border-makoclaw-border/50 hover:bg-makoclaw-surface-hover hover:border-orange-400/30 transition-all flex items-center justify-center active:scale-95"
                title="Export chat history"
                @click="showExportMenu = !showExportMenu"
              >
                <svg
                  class="w-4 h-4 text-makoclaw-text-secondary"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                /></svg>
              </button>
              <div
                v-if="showExportMenu"
                class="absolute right-0 top-full mt-2 w-52 bg-makoclaw-surface/95 backdrop-blur-xl border border-makoclaw-border/50 rounded-xl shadow-2xl p-1.5 z-50 ring-1 ring-white/10"
              >
                <button
                  class="w-full text-left px-3 py-2.5 hover:bg-orange-500/10 hover:text-orange-400 rounded-lg text-sm transition-all"
                  @click="handleExportAll"
                >
                  Export All Chats
                </button>
                <button
                  v-if="activeSessionId"
                  class="w-full text-left px-3 py-2.5 hover:bg-orange-500/10 hover:text-orange-400 rounded-lg text-sm transition-all"
                  @click="handleExportSession"
                >
                  Export Current Session
                </button>
              </div>
            </div>
            <button
              class="p-2.5 min-h-[40px] min-w-[40px] rounded-xl bg-makoclaw-surface/50 border border-makoclaw-border/50 hover:bg-makoclaw-surface-hover hover:border-orange-400/30 transition-all flex items-center justify-center active:scale-95"
              title="Import conversations (ChatGPT, Claude, makoclaw)"
              @click="triggerImport"
            >
              <svg
                class="w-4 h-4 text-makoclaw-text-secondary"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
              /></svg>
            </button>
            <input
              ref="importFileInput"
              type="file"
              accept=".json"
              class="hidden"
              @change="handleImportFile"
            >
            <button
              class="p-2.5 min-h-[40px] min-w-[40px] rounded-xl bg-makoclaw-surface/50 border border-makoclaw-border/50 hover:bg-makoclaw-surface-hover hover:border-orange-400/30 transition-all flex items-center justify-center active:scale-95"
              title="Refresh"
              @click="loadSessions"
            >
              <svg
                class="w-4 h-4 text-makoclaw-text-secondary"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              /></svg>
            </button>
          </div>
        </div>

      <div class="px-4 sm:px-6 pb-3">
        <!-- AppBar Subtitle -->
        <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5">
          Browse, search, and manage past conversations
        </p>
      </div>
      </div>
    </div>

    <div class="flex-1 flex overflow-hidden relative z-10">
      <!-- Session List / Search Results -->
      <div 
        class="border-r border-makoclaw-border/30 overflow-hidden bg-makoclaw-surface/30 backdrop-blur-sm transition-all duration-500 ease-[cubic-bezier(0.4,0,0.2,1)] flex flex-col ring-1 ring-white/5"
        :class="isSidebarCollapsed ? 'w-0 opacity-0 border-none overflow-hidden scale-95 origin-left' : 'w-72 opacity-100 flex-shrink-0'"
      >
        <!-- Mini Header inside Sidebar -->
        <div class="p-3 sm:p-4 border-b border-makoclaw-border/30 bg-makoclaw-surface/30">
          <div class="flex items-center gap-2 mb-3">
            <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-orange-500/20 to-amber-500/20 flex items-center justify-center ring-1 ring-white/10">
              <svg
                class="w-4 h-4 text-orange-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <h2 class="text-sm font-bold text-makoclaw-text uppercase tracking-wider">Session History</h2>
          </div>

          <!-- Search bar in Sidebar -->
          <div class="relative group/search mb-3">
            <svg
              class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-makoclaw-text-secondary group-focus-within/search:text-orange-400 transition-colors"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            /></svg>
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search history..."
              class="w-full pl-10 pr-8 py-1.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-xs outline-none focus:border-orange-400/50 focus:ring-2 focus:ring-orange-400/20 transition-all font-medium"
              @input="onSearchInput"
            >
            <button
              v-if="searchQuery"
              class="absolute right-2 top-1/2 -translate-y-1/2 p-1 hover:bg-makoclaw-surface rounded-lg text-makoclaw-text-secondary"
              @click="clearSearch"
            >
              <svg
                class="w-3.5 h-3.5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              /></svg>
            </button>
          </div>

          <!-- Filters Selects in Sidebar -->
          <div class="flex items-center gap-2">
            <select
              v-model="filterType"
              class="flex-1 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-lg px-2 py-1 text-[10px] outline-none focus:border-orange-400/50 cursor-pointer appearance-none text-center font-bold text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
            >
              <option value="all">All Sessions</option>
              <option value="chat">Chats Only</option>
              <option value="task">Tasks Only</option>
            </select>
            <select
              v-model="filterDate"
              class="flex-1 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-lg px-2 py-1 text-[10px] outline-none focus:border-orange-400/50 cursor-pointer appearance-none text-center font-bold text-makoclaw-text-secondary hover:text-makoclaw-text transition-colors"
            >
              <option value="all">All Time</option>
              <option value="today">Today</option>
              <option value="7d">Last 7d</option>
              <option value="30d">Last 30d</option>
            </select>
            <button
              :class="[
                'px-2 py-1 rounded-lg text-[10px] font-bold transition-colors flex-shrink-0',
                showArchived
                  ? 'bg-amber-500/20 text-amber-400 border border-amber-500/50'
                  : 'bg-makoclaw-bg/50 border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-makoclaw-text'
              ]"
              @click="toggleShowArchived"
            >
              {{ showArchived ? 'Archived' : 'Active' }}
            </button>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto p-2 space-y-1 custom-scrollbar">
        <div
          v-if="loading"
          class="text-center py-4 text-makoclaw-text-secondary animate-pulse"
        >
          Loading sessions...
        </div>
        <div
          v-else-if="searching"
          class="text-center py-4 text-makoclaw-text-secondary animate-pulse"
        >
          Searching...
        </div>

        <!-- Search results mode -->
        <template v-else-if="isSearchMode">
          <div
            v-if="filteredSearchResults.length === 0"
            class="text-center py-4 text-makoclaw-text-secondary"
          >
            No matching messages found for current filters.
          </div>
          <div
            v-for="(result, idx) in filteredSearchResults"
            :key="idx"
            class="p-3 rounded-lg cursor-pointer transition-all border border-transparent hover:border-makoclaw-border hover:bg-makoclaw-bg/50"
            :class="selectedSearchResult === idx ? 'bg-makoclaw-bg border-makoclaw-accent/30 shadow-sm' : ''"
            @click="selectSearchResult(result)"
          >
            <div class="flex items-center gap-2 mb-1">
              <span
                class="text-[10px] font-medium px-1.5 py-0.5 rounded"
                :class="result.role === 'user' ? 'bg-makoclaw-accent/20 text-makoclaw-accent' : 'bg-blue-500/20 text-blue-400'"
              >
                {{ result.role }}
              </span>
              <span
                class="text-[10px] px-1.5 py-0.5 rounded"
                :class="isTaskSession(result) ? 'bg-amber-500/20 text-amber-400' : 'bg-blue-500/20 text-blue-400'"
              >
                {{ isTaskSession(result) ? 'Task' : 'Chat' }}
              </span>
            </div>
            <div
              class="text-sm truncate text-makoclaw-text"
              v-html="highlightMatch(result.content, searchQuery)"
            />
            <div class="text-[10px] text-makoclaw-text-secondary mt-1">
              {{ formatTime(result.created_at) }}
            </div>
          </div>
        </template>

        <!-- Normal sessions mode -->
        <template v-else>
          <div
            v-if="filteredSessions.length === 0"
            class="text-center py-4 text-makoclaw-text-secondary"
          >
            No matching history found.
          </div>

          <div
            v-for="session in filteredSessions"
            :key="session.session_id"
            class="p-3 rounded-lg cursor-pointer transition-all border border-transparent hover:border-makoclaw-border group/session list-item-interactive relative overflow-hidden"
            :class="selectedSession?.session_id === session.session_id ? 'bg-makoclaw-bg border-makoclaw-accent/30 shadow-sm' : ''"
            @click="selectSession(session)"
          >
            <!-- Hover glow line (left edge) -->
            <div class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-0 bg-gradient-to-b from-transparent via-makoclaw-accent/40 to-transparent group-hover/session:h-2/3 transition-all duration-300 pointer-events-none" />
            <!-- Animated bottom-line on hover -->
            <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover/session:w-full transition-all duration-500 opacity-40 pointer-events-none" />
            <!-- Soft gradient glow (top-right) -->
            <div class="absolute -top-6 -right-6 w-16 h-16 bg-gradient-to-br from-makoclaw-accent to-blue-500 rounded-full opacity-0 blur-[15px] group-hover/session:opacity-10 transition-all duration-500 pointer-events-none" />

            <!-- Inline rename -->
            <div
              v-if="renamingSession === session.session_id"
              class="relative z-10 flex items-center gap-1"
              @click.stop
            >
              <input
                v-model="renameInput"
                class="flex-1 bg-makoclaw-bg border border-makoclaw-accent rounded px-2 py-1 text-xs text-makoclaw-text focus:outline-none"
                autofocus
                placeholder="Session title..."
                @keyup.enter="submitRenameSession(session.session_id)"
                @keyup.escape="cancelRenameSession"
                @blur="submitRenameSession(session.session_id)"
              >
            </div>
            <template v-else>
              <div class="flex items-center gap-2 mb-1">
                <!-- Session type icon + badge -->
                <svg
                  v-if="isTaskSession(session)"
                  class="w-4 h-4 text-amber-400 flex-shrink-0"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
                /></svg>
                <svg
                  v-else
                  class="w-4 h-4 text-blue-400 flex-shrink-0"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                /></svg>
                <div class="font-medium truncate text-sm flex-1">
                  {{ getSessionTitle(session) }}
                </div>
                <!-- Context menu trigger -->
                <button
                  class="p-1 rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text hover:bg-makoclaw-surface/50 transition-all flex-shrink-0 opacity-0 group-hover/session:opacity-100 touch-visible"
                  title="Session actions"
                  @click.stop="openContextMenu($event, session.session_id)"
                >
                  <svg
                    class="w-4 h-4"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  ><path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" /></svg>
                </button>
              </div>
              <div class="text-xs text-makoclaw-text-secondary flex justify-between items-center mt-1.5 pl-6">
                <span>{{ formatTime(session.updated_at) }}</span>
                <div class="flex items-center gap-2">
                  <span
                    v-if="session.message_count"
                    class="text-[10px] text-makoclaw-text-secondary"
                  >
                    {{ session.message_count }} msg{{ session.message_count !== 1 ? 's' : '' }}
                  </span>
                  <span
                    class="px-1.5 py-0.5 rounded text-[10px] font-medium"
                    :class="isTaskSession(session) ? 'bg-amber-500/20 text-amber-400' : 'bg-blue-500/20 text-blue-400'"
                  >
                    {{ isTaskSession(session) ? 'Task' : 'Chat' }}
                  </span>
                </div>
              </div>
            </template>
          </div>
        </template>

        <!-- Session Context Menu -->
        <SessionContextMenu
          :show="contextMenu.show"
          :session-id="contextMenu.sessionId"
          :x="contextMenu.x"
          :y="contextMenu.y"
          :show-archived="showArchived"
          @close="closeContextMenu"
          @rename="handleRename"
          @archive="handleArchive"
          @delete="handleDelete"
        />

        <!-- Load More Button -->
        <div
          v-if="!isSearchMode && hasMoreSessions && filteredSessions.length > 0"
          class="py-3 px-2"
        >
          <button
            :disabled="loading"
            class="w-full py-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg text-xs font-semibold text-makoclaw-accent hover:bg-makoclaw-surface transition-colors disabled:opacity-50"
            @click="loadMore"
          >
            {{ loading ? 'Loading...' : 'Load More Sessions' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Message View -->
      <div class="flex-1 overflow-y-auto p-4 bg-makoclaw-bg/30 relative custom-scrollbar">
        <div
          v-if="!selectedSession && !selectedSearchResultSession"
          class="h-full flex flex-col items-center justify-center text-makoclaw-text-secondary"
        >
          <div class="relative glass-panel p-6 sm:p-8 rounded-2xl mb-4 shadow-lg shadow-orange-500/5 group/empty">
            <!-- Gradient glow behind icon -->
            <div class="absolute -top-8 -right-8 w-24 h-24 bg-gradient-to-br from-orange-500 to-amber-500 rounded-full opacity-15 blur-[25px] group-hover/empty:opacity-30 group-hover/empty:scale-110 transition-all duration-500" />
            <svg
              class="w-12 h-12 text-orange-400 drop-shadow-sm relative z-10 transition-transform duration-500 group-hover/empty:scale-110 group-hover/empty:rotate-3"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.5"
              d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"
            /></svg>
            <!-- Animated bottom-line -->
            <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover/empty:w-full transition-all duration-500 opacity-70 rounded-b-2xl" />
          </div>
          <p class="text-sm font-semibold text-makoclaw-text">
            Select a session to view the conversation
          </p>
          <p class="text-xs mt-1 opacity-60">
            Or search for specific messages above
          </p>
        </div>

        <div
          v-else
          class="space-y-4 max-w-3xl mx-auto pb-20"
        >
          <div class="text-center mb-6 sticky top-0 z-10 bg-makoclaw-bg/95 backdrop-blur py-2 border-b border-makoclaw-border/50 flex justify-between items-center px-4 rounded-b-xl shadow-sm">
            <div class="flex items-center gap-2 text-xs text-makoclaw-text-secondary">
              <svg
                v-if="activeSessionIsTask"
                class="w-3.5 h-3.5 text-amber-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"
              /></svg>
              <svg
                v-else
                class="w-3.5 h-3.5 text-blue-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
              /></svg>
              <span class="font-mono">{{ activeSessionId }}</span>
            </div>
            <button
              class="flex items-center gap-2 px-3 py-1.5 bg-makoclaw-accent text-white text-xs rounded-lg hover:bg-makoclaw-accent-hover transition-colors shadow-sm"
              @click="continueChat"
            >
              <span>Continue Chat</span>
              <svg
                class="w-3 h-3"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M14 5l7 7m0 0l-7 7m7-7H3"
              /></svg>
            </button>
          </div>

          <div
            v-if="loadingMessages"
            class="text-center py-8"
          >
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-makoclaw-accent mx-auto" />
          </div>

          <div
            v-for="(msg, index) in messages"
            v-else
            :key="index"
            class="flex flex-col gap-1 group animate-fadeUp opacity-0"
            :class="msg.role === 'user' ? 'items-end' : 'items-start'"
            :style="{ animationDelay: Math.min(index * 30, 300) + 'ms' }"
          >
            <div class="text-xs text-makoclaw-text-secondary px-1 opacity-70 flex items-center gap-2">
              <span>{{ msg.role }}</span>
              <button
                v-if="msg.role === 'assistant'"
                class="opacity-0 group-hover:opacity-100 transition-opacity p-0.5 hover:bg-makoclaw-surface rounded text-makoclaw-text-secondary hover:text-makoclaw-accent"
                title="Copy response"
                @click="copyMessageContent(msg.content)"
              >
                <svg
                  class="w-3.5 h-3.5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                /></svg>
              </button>
              <button
                class="opacity-0 group-hover:opacity-100 transition-opacity p-0.5 hover:bg-makoclaw-surface rounded text-makoclaw-text-secondary hover:text-makoclaw-accent"
                title="Fork conversation from this point"
                @click="forkAtMessage(msg)"
              >
                <svg
                  class="w-3.5 h-3.5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
                /></svg>
              </button>
            </div>
            <div
              class="max-w-[85%] rounded-2xl px-4 py-3 text-sm shadow-sm leading-relaxed relative overflow-hidden group/msg"
              :class="msg.role === 'user'
                ? 'bg-gradient-to-br from-makoclaw-accent to-makoclaw-accent-hover text-white rounded-tr-none whitespace-pre-wrap'
                : 'glass-panel border-makoclaw-border rounded-tl-none text-makoclaw-text'"
            >
              <!-- Hover accent line for assistant messages -->
              <div
                v-if="msg.role !== 'user'"
                class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover/msg:w-full transition-all duration-500 opacity-50"
              />

              <span v-if="msg.role === 'user'">{{ msg.content }}</span>
              <MarkdownRenderer
                v-else
                :content="msg.content"
              />
            </div>
            <div class="text-[10px] text-gray-400 px-1">
              {{ formatTime(msg.created_at || msg.timestamp) }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, onBeforeUnmount } from 'vue'
import api from '../services/api'
import taskService from '../services/taskService'
import advancedService from '../services/advancedService'
import MarkdownRenderer from '../components/Chat/MarkdownRenderer.vue'
import SessionContextMenu from '../components/Chat/SessionContextMenu.vue'
import { useRouter } from 'vue-router'
import { useToast } from '../composables/useToast'

const router = useRouter()
const toast = useToast()
const isSidebarCollapsed = ref(localStorage.getItem('history.sidebar') === 'true')
const sessions = ref([])

const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
  localStorage.setItem('history.sidebar', isSidebarCollapsed.value)
}
const selectedSession = ref(null)
const messages = ref([])
const loading = ref(false)
const loadingMessages = ref(false)
const filterType = ref('all')
const filterDate = ref('all')
const showArchived = ref(false)

// Pagination state
const sessionLimit = 30
const sessionOffset = ref(0)
const hasMoreSessions = ref(true)

// Session management
const renamingSession = ref(null)
const renameInput = ref('')

// Context menu state
const contextMenu = ref({ show: false, sessionId: null, x: 0, y: 0 })

// Search state
const searchQuery = ref('')
const searchResults = ref([])
const searching = ref(false)
const selectedSearchResult = ref(null)
const selectedSearchResultSession = ref(null)
let searchTimeout = null

// Export state
const showExportMenu = ref(false)
const exportDropdownRef = ref(null)

// Import state
const importFileInput = ref(null)
const importing = ref(false)


const handleExportAll = () => {
  advancedService.exportChat()
  showExportMenu.value = false
}

const handleExportSession = () => {
  const sid = activeSessionId.value
  if (sid) {
    advancedService.exportChat(sid)
  }
  showExportMenu.value = false
}

const triggerImport = () => {
  importFileInput.value?.click()
}

const handleImportFile = async (event) => {
  const file = event.target.files?.[0]
  if (!file) return

  importing.value = true
  try {
    const text = await file.text()
    const data = JSON.parse(text)
    const result = await advancedService.importChat(data)
    toast.success(`Imported ${result.sessions_imported} session(s) with ${result.messages_imported} message(s)`)
    await loadSessions()
  } catch (error) {
    console.error('Import failed:', error)
    toast.error('Import failed: ' + (error.response?.data || error.message || 'Unknown error'))
  } finally {
    importing.value = false
    // Reset file input so the same file can be re-selected
    if (importFileInput.value) importFileInput.value.value = ''
  }
}

const handleClickOutside = (e) => {
  if (exportDropdownRef.value && !exportDropdownRef.value.contains(e.target)) {
    showExportMenu.value = false
  }
}

const isSearchMode = computed(() => searchQuery.value.trim().length >= 2)

const filteredSearchResults = computed(() => {
  const q = searchQuery.value.trim()
  if (q.length < 2) return []
  
  let result = searchResults.value
  
  // Apply type filter to search results too
  if (filterType.value !== 'all') {
    result = result.filter(s => {
      const isTask = isTaskSession(s)
      return filterType.value === 'chat' ? !isTask : isTask
    })
  }
  
  return result
})

const isTaskSession = (session) => {
  const id = session.session_id || ''
  return id.startsWith('task:') || id.includes(':task:')
}

const activeSessionId = computed(() => {
  if (selectedSession.value) return selectedSession.value.session_id
  if (selectedSearchResultSession.value) return selectedSearchResultSession.value
  return ''
})

const activeSessionIsTask = computed(() => {
  return activeSessionId.value.startsWith('web:task:')
})

const filteredSessions = computed(() => {
  let result = sessions.value

  // Filter by type
  if (filterType.value !== 'all') {
    result = result.filter(s => {
      const isTask = isTaskSession(s)
      return filterType.value === 'chat' ? !isTask : isTask
    })
  }

  // Filter by date
  if (filterDate.value !== 'all') {
    const now = new Date()
    let cutoff
    if (filterDate.value === 'today') {
      cutoff = new Date(now.getFullYear(), now.getMonth(), now.getDate())
    } else if (filterDate.value === '7d') {
      cutoff = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
    } else if (filterDate.value === '30d') {
      cutoff = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
    }
    if (cutoff) {
      result = result.filter(s => new Date(s.updated_at) >= cutoff)
    }
  }

  return result
})

const continueChat = () => {
  const sid = activeSessionId.value
  if (sid) {
    router.push({ path: '/chat', query: { id: sid } })
  }
}

const forkAtMessage = async (msg) => {
  const sid = activeSessionId.value
  if (!sid) return

  if (!confirm('Fork conversation from this message? A new session will be created with all messages up to this point.')) return

  try {
    const result = await advancedService.forkChat(sid, msg.id)
    toast.success(`Forked! New session created with ${result.messages_copied} message(s)`)
    await loadSessions()
    // Navigate to the forked session in chat
    router.push({ path: '/chat', query: { id: result.new_session_id } })
  } catch (error) {
    console.error('Fork failed:', error)
    toast.error('Fork failed: ' + (error.response?.data || error.message || 'Unknown error'))
  }
}

const loadSessions = async (reset = true) => {
  if (reset) {
    sessionOffset.value = 0
    sessions.value = []
    hasMoreSessions.value = true
  }
  loading.value = true
  try {
    const data = await taskService.fetchChatSessions({ 
      archived: showArchived.value,
      limit: sessionLimit,
      offset: sessionOffset.value
    })
    const newSessions = data.sessions || []
    if (reset) {
      sessions.value = newSessions
    } else {
      sessions.value = [...sessions.value, ...newSessions]
    }
    hasMoreSessions.value = newSessions.length >= sessionLimit
  } catch (error) {
    console.error('Failed to load sessions:', error)
    toast.error('Failed to load sessions')
  } finally {
    loading.value = false
  }
}

const loadMore = () => {
  if (!loading.value && hasMoreSessions.value) {
    sessionOffset.value += sessionLimit
    loadSessions(false)
  }
}

// Session actions
const deleteSessionById = async (sessionId) => {
  if (!confirm('Delete this session and all its messages? This cannot be undone.')) return
  try {
    await taskService.deleteSession(sessionId)
    sessions.value = sessions.value.filter(s => s.session_id !== sessionId)
    if (selectedSession.value?.session_id === sessionId) {
      selectedSession.value = null
      messages.value = []
    }
    toast.success('Session deleted')
  } catch (error) {
    console.error('Failed to delete session:', error)
    toast.error('Failed to delete session')
  }
}

const archiveSessionById = async (sessionId) => {
  const session = sessions.value.find(s => s.session_id === sessionId)
  if (!session) return
  const newArchivedState = !session.archived
  try {
    await taskService.updateSession(sessionId, { archived: newArchivedState })
    sessions.value = sessions.value.filter(s => s.session_id !== sessionId)
    if (selectedSession.value?.session_id === sessionId) {
      selectedSession.value = null
      messages.value = []
    }
    toast.success(newArchivedState ? 'Session archived' : 'Session unarchived')
  } catch (error) {
    console.error('Failed to update session:', error)
    toast.error('Failed to update session')
  }
}

const startRenameSession = (session) => {
  renamingSession.value = session.session_id
  renameInput.value = session.title || ''
}

const submitRenameSession = async (sessionId) => {
  if (renamingSession.value !== sessionId) return
  try {
    await taskService.updateSession(sessionId, { title: renameInput.value.trim() })
    const session = sessions.value.find(s => s.session_id === sessionId)
    if (session) session.title = renameInput.value.trim()
    toast.success('Session renamed')
  } catch (error) {
    console.error('Failed to rename session:', error)
    toast.error('Failed to rename session')
  }
  renamingSession.value = null
  renameInput.value = ''
}

const cancelRenameSession = () => {
  renamingSession.value = null
  renameInput.value = ''
}

const toggleShowArchived = () => {
  showArchived.value = !showArchived.value
  loadSessions()
}

// Context menu handlers
const openContextMenu = (e, sessionId) => {
  e.preventDefault()
  e.stopPropagation()
  contextMenu.value = { show: true, sessionId, x: e.clientX, y: e.clientY }
}

const closeContextMenu = () => {
  contextMenu.value = { show: false, sessionId: null, x: 0, y: 0 }
}

const handleRename = (sessionId) => {
  closeContextMenu()
  const session = sessions.value.find(s => s.session_id === sessionId)
  if (session) startRenameSession(session)
}

const handleArchive = (sessionId) => {
  closeContextMenu()
  archiveSessionById(sessionId)
}

const handleDelete = (sessionId) => {
  closeContextMenu()
  deleteSessionById(sessionId)
}

const selectSession = async (session) => {
  selectedSession.value = session
  selectedSearchResult.value = null
  selectedSearchResultSession.value = null
  loadingMessages.value = true
  messages.value = []

  try {
    const response = await api.get(`/chat/sessions/${encodeURIComponent(session.session_id)}`)
    messages.value = response.data.messages || []
  } catch (error) {
    console.error('Failed to load messages:', error)
  } finally {
    loadingMessages.value = false
  }
}

const selectSearchResult = async (result) => {
  const idx = searchResults.value.indexOf(result)
  selectedSearchResult.value = idx
  selectedSession.value = null
  selectedSearchResultSession.value = result.session_id
  loadingMessages.value = true
  messages.value = []

  try {
    const response = await api.get(`/chat/sessions/${encodeURIComponent(result.session_id)}`)
    messages.value = response.data.messages || []
  } catch (error) {
    console.error('Failed to load session messages:', error)
  } finally {
    loadingMessages.value = false
  }
}

const onSearchInput = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  const q = searchQuery.value.trim()
  if (q.length < 2) {
    searchResults.value = []
    selectedSearchResult.value = null
    selectedSearchResultSession.value = null
    return
  }
  searchTimeout = setTimeout(async () => {
    searching.value = true
    try {
      const data = await taskService.searchMessages(q)
      searchResults.value = data.messages || []
    } catch (error) {
      console.error('Search failed:', error)
      searchResults.value = []
    } finally {
      searching.value = false
    }
  }, 300)
}

const clearSearch = () => {
  searchQuery.value = ''
  searchResults.value = []
  selectedSearchResult.value = null
  selectedSearchResultSession.value = null
}

const highlightMatch = (text, query) => {
  if (!query || !text) return text
  const truncated = text.length > 120 ? text.substring(0, 120) + '...' : text
  const escaped = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return truncated.replace(
    new RegExp(`(${escaped})`, 'gi'),
    '<mark class="bg-makoclaw-accent/30 text-makoclaw-text rounded px-0.5">$1</mark>'
  )
}

const getSessionTitle = (session) => {
  if (session.title) return session.title
  if (session.last_message) {
    if (session.last_message.length > 50) {
      return session.last_message.substring(0, 50) + '...'
    }
    return session.last_message
  }
  return `Session ${session.session_id}`
}

const copyMessageContent = async (content) => {
  try {
    await navigator.clipboard.writeText(content)
    toast.success('Copied to clipboard')
  } catch {
    toast.error('Failed to copy')
  }
}

const formatTime = (ts) => {
  if (!ts) return ''
  return new Date(ts).toLocaleString()
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  loadSessions()
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
@keyframes float {
  0%, 100% { transform: translateY(0) translateX(0); }
  50% { transform: translateY(-20px) translateX(10px); }
}

.list-item-interactive {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.glass-sticky {
  background: rgba(var(--makoclaw-surface-rgb), 0.7);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

/* Touch-friendly action buttons - visible on touch devices */
@media (hover: none) {
  .touch-visible {
    opacity: 0.6 !important;
  }
}
</style>
