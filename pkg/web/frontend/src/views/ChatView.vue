<template>
  <div class="flex flex-col h-full bg-makoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-emerald-500/30 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-teal-500/20 via-transparent to-transparent" />
    </div>

    <!-- Header -->
    <div class="glass-sticky top-0 z-20 border-b border-makoclaw-border/20">
      <div class="px-4 sm:px-6 pt-4 sm:pt-5 pb-3">
        <!-- Title Row -->
        <div class="flex items-center gap-3">
          <!-- Sidebar toggle -->
          <button
            class="p-2 min-h-[40px] min-w-[40px] rounded-xl bg-makoclaw-surface/50 border border-makoclaw-border/50 hover:bg-makoclaw-surface-hover hover:border-emerald-400/30 transition-all flex items-center justify-center active:scale-95"
            title="Toggle Sidebar"
            @click="toggleSidebar"
          >
            <svg
              class="w-4 h-4 text-makoclaw-text-secondary transition-transform duration-500"
              :class="{'rotate-180': !showSidebar}"
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
          <div class="w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 flex items-center justify-center ring-1 ring-white/10 shadow-lg shadow-emerald-500/10">
            <svg
              class="w-5 h-5 sm:w-6 sm:h-6 text-emerald-400"
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
          </div>

          <div class="flex-1 min-w-0">
            <h1 class="text-xl sm:text-2xl font-bold bg-gradient-to-r from-makoclaw-text via-makoclaw-text to-emerald-400 bg-clip-text text-transparent">
              Chat
            </h1>
            <p class="text-xs sm:text-sm text-makoclaw-text-secondary mt-0.5 hidden sm:block">
              Converse with your AI assistant
            </p>
          </div>

          <!-- Agent status indicator -->
          <div
            v-if="chatStore.globalIsLoading"
            class="flex items-center gap-2 px-3 py-1.5 bg-gradient-to-r from-emerald-500/15 to-transparent border border-emerald-500/30 rounded-xl flex-shrink-0"
          >
            <div class="w-5 h-5 rounded-lg bg-emerald-500/20 flex items-center justify-center flex-shrink-0">
              <svg
                class="w-3 h-3 text-emerald-400 animate-spin"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13 10V3L4 14h7v7l9-11h-7z"
                />
              </svg>
            </div>
            <span class="text-xs font-medium text-emerald-400 truncate">Agent working...</span>
          </div>

          <!-- Model selector -->
          <div class="flex items-center gap-2 flex-shrink-0">
            <div class="w-7 h-7 rounded-lg bg-makoclaw-surface/50 flex items-center justify-center">
              <svg
                class="w-3.5 h-3.5 text-makoclaw-text-secondary"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                />
              </svg>
            </div>
            <select
              v-model="chatStore.selectedModel"
              :disabled="chatStore.allModels.length === 0"
              class="bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl px-3 py-1.5 text-xs text-makoclaw-text focus:ring-2 focus:ring-emerald-400/30 focus:border-emerald-400 transition-all cursor-pointer max-w-[160px] md:max-w-[280px]"
            >
              <option
                v-if="chatStore.allModels.length === 0"
                value=""
              >
                No models available
              </option>
              <option
                v-for="model in chatStore.allModels"
                :key="model.id"
                :value="model.id"
              >
                {{ model.provider }}/{{ model.label }}{{ model.isDefault ? ' (default)' : '' }}
              </option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- Content: Sidebar + Main Chat -->
    <div class="flex-1 flex overflow-hidden relative z-10">
      <!-- Sidebar (History) -->
      <div
        :class="[
          'flex-shrink-0 border-r border-makoclaw-border/20 bg-makoclaw-surface/30 backdrop-blur-sm transition-all duration-500 ease-[cubic-bezier(0.4,0,0.2,1)] flex flex-col ring-1 ring-white/5 custom-scrollbar',
          showSidebar ? 'w-72 opacity-100' : 'w-0 opacity-0 border-none overflow-hidden scale-95 origin-left'
        ]"
      >
        <div class="p-3 sm:p-4 border-b border-makoclaw-border/30 bg-makoclaw-surface/30">
          <div class="flex justify-between items-center gap-2 mb-3">
            <div class="flex items-center gap-2">
              <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-emerald-500/20 to-teal-500/20 flex items-center justify-center ring-1 ring-white/10">
                <svg
                  class="w-4 h-4 text-emerald-400"
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
              <h2 class="text-sm font-bold text-makoclaw-text uppercase tracking-wider">Chat History</h2>
            </div>
            <button
              class="p-2 hover:bg-emerald-500/10 rounded-lg text-makoclaw-text-secondary hover:text-emerald-400 transition-all flex items-center justify-center ring-1 ring-white/5 active:scale-95"
              title="New Chat"
              @click="startNewChat"
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
                d="M12 4v16m8-8H4"
              /></svg>
            </button>
          </div>

          <!-- Search bar in Sidebar -->
          <div class="relative group/search">
            <svg
              class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-makoclaw-text-secondary group-focus-within/search:text-emerald-400 transition-colors"
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
              class="w-full pl-10 pr-8 py-1.5 bg-makoclaw-bg/50 border border-makoclaw-border/50 rounded-xl text-xs outline-none focus:border-emerald-400/50 focus:ring-2 focus:ring-emerald-400/20 transition-all font-medium"
            >
            <button
              v-if="searchQuery"
              class="absolute right-2 top-1/2 -translate-y-1/2 p-1 hover:bg-makoclaw-surface rounded-lg text-makoclaw-text-secondary"
              @click="searchQuery = ''"
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
        </div>
      
        <div class="flex-1 overflow-y-auto p-2 space-y-1 custom-scrollbar">
          <div
            v-if="sessions.length === 0"
            class="flex flex-col items-center justify-center py-8 text-center"
          >
            <div class="w-10 h-10 rounded-xl bg-makoclaw-surface/50 flex items-center justify-center mb-3">
              <svg
                class="w-5 h-5 text-makoclaw-text-secondary/40"
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
            </div>
            <p class="text-xs text-makoclaw-text-secondary">
              No history yet
            </p>
          </div>
          <div
            v-for="session in sessions"
            :key="session.session_id"
            class="relative group/session list-item-interactive overflow-hidden cursor-pointer"
            :class="currentSessionId === session.session_id ? 'bg-makoclaw-bg border-makoclaw-accent/30 shadow-sm' : ''"
          >
            <!-- Hover glow line (left edge) -->
            <div class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-0 bg-gradient-to-b from-transparent via-makoclaw-accent/40 to-transparent group-hover/session:h-2/3 transition-all duration-300" />
            <!-- Animated bottom-line on hover -->
            <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover/session:w-full transition-all duration-500 opacity-40" />
            <!-- Soft gradient glow (top-right) -->
            <div class="absolute -top-6 -right-6 w-16 h-16 bg-gradient-to-br from-makoclaw-accent to-blue-500 rounded-full opacity-0 blur-[15px] group-hover/session:opacity-10 transition-all duration-500" />

            <!-- Inline rename -->
            <div
              v-if="renamingSession === session.session_id"
              class="flex items-center gap-1 px-2 py-1.5"
            >
              <input
                v-model="renameInput"
                class="flex-1 bg-makoclaw-bg border border-makoclaw-accent/50 rounded-lg px-2.5 py-1.5 text-xs text-makoclaw-text focus:outline-none focus:ring-2 focus:ring-makoclaw-accent/30"
                autofocus
                placeholder="Session title..."
                @keyup.enter="submitRename(session.session_id)"
                @keyup.escape="cancelRename"
                @blur="submitRename(session.session_id)"
              >
            </div>
            <!-- Normal session button -->
            <button
              v-else
              class="w-full text-left px-2.5 py-2 rounded-lg text-sm transition-all duration-200"
              :class="currentSessionId === session.session_id ? 'text-makoclaw-text' : 'text-makoclaw-text-secondary hover:text-makoclaw-text'"
              @click="loadSession(session.session_id)"
            >
              <div class="flex items-center gap-2.5">
                <div
                  :class="[
                    'w-6 h-6 rounded-lg flex items-center justify-center flex-shrink-0',
                    session.session_id.startsWith('web:task:') ? 'bg-amber-500/10' : 'bg-makoclaw-accent/10'
                  ]"
                >
                  <svg
                    v-if="session.session_id.startsWith('web:task:')"
                    class="w-3 h-3 text-amber-400"
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
                    class="w-3 h-3 text-makoclaw-accent"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  ><path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                  /></svg>
                </div>
                <span class="truncate flex-1 text-xs font-medium">{{ session.title || session.last_message || 'Empty session' }}</span>
                <!-- Context menu trigger -->
                <button
                  class="p-1 rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-text hover:bg-makoclaw-surface/50 transition-all flex-shrink-0 opacity-0 group-hover/session:opacity-100"
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
              <div class="text-[10px] opacity-60 mt-1 pl-8 flex justify-between">
                <span>{{ formatTime(session.updated_at) }}</span>
                <span
                  v-if="session.message_count"
                  class="text-makoclaw-text-secondary"
                >{{ session.message_count }} msg{{ session.message_count !== 1 ? 's' : '' }}</span>
              </div>
            </button>
          </div>
        </div>

        <!-- Context Menu Overlay -->
        <Teleport to="body">
          <div
            v-if="contextMenu.show"
            class="fixed inset-0 z-dropdown"
            @click="closeContextMenu"
          >
            <div
              class="absolute bg-makoclaw-surface/95 backdrop-blur-xl border border-makoclaw-border/50 rounded-xl shadow-2xl py-1.5 min-w-[160px] animate-scaleIn ring-1 ring-white/10"
              :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
              @click.stop
            >
              <button
                class="w-full text-left px-3 py-2 text-sm hover:bg-makoclaw-accent/10 hover:text-makoclaw-accent transition-all flex items-center gap-2.5"
                @click="startRename(contextMenu.sessionId)"
              >
                <svg
                  class="w-4 h-4 opacity-70"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                /></svg>
                Rename
              </button>
              <button
                class="w-full text-left px-3 py-2 text-sm hover:bg-makoclaw-accent/10 hover:text-makoclaw-accent transition-all flex items-center gap-2.5"
                @click="archiveSession(contextMenu.sessionId)"
              >
                <svg
                  class="w-4 h-4 opacity-70"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"
                /></svg>
                Archive
              </button>
              <div class="border-t border-makoclaw-border/50 my-1.5 mx-2" />
              <button
                class="w-full text-left px-3 py-2 text-sm hover:bg-makoclaw-error/10 text-makoclaw-error transition-all flex items-center gap-2.5"
                @click="deleteSession(contextMenu.sessionId)"
              >
                <svg
                  class="w-4 h-4 opacity-70"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                /></svg>
                Delete
              </button>
            </div>
          </div>
        </Teleport>
      </div>

      <!-- Main Chat Area -->
      <div class="flex-1 flex flex-col min-w-0 relative">
        <!-- Specialists Panel -->
        <div class="px-2 sm:px-3 md:px-4 pt-2 sm:pt-3">
          <SpecialistsPanel />
        </div>

        <!-- Messages Area -->
        <div
          ref="messagesContainer"
          class="flex-1 overflow-y-auto p-3 sm:p-4 md:p-6 space-y-4 z-10 custom-scrollbar"
        >
          <div
            v-if="messages.length === 0"
            class="flex flex-col items-center justify-center h-full text-makoclaw-text-secondary animate-fadeIn"
          >
            <div class="relative">
              <!-- Glow effect -->
              <div class="absolute inset-0 bg-gradient-to-br from-emerald-500/30 to-teal-500/30 rounded-3xl blur-2xl opacity-50" />
              <div class="relative glass-panel p-8 sm:p-10 rounded-2xl shadow-2xl shadow-emerald-500/10 ring-1 ring-white/10 group">
                <!-- Gradient glow behind icon -->
                <div class="absolute -top-8 -right-8 w-24 h-24 bg-gradient-to-br from-emerald-500 to-teal-500 rounded-full opacity-15 blur-[25px] group-hover:opacity-30 group-hover:scale-110 transition-all duration-500" />
                <div class="w-16 h-16 sm:w-20 sm:h-20 mx-auto rounded-2xl bg-gradient-to-br from-emerald-500/20 to-teal-500/20 flex items-center justify-center ring-1 ring-white/20 relative z-10 transition-transform duration-500 group-hover:scale-110 group-hover:rotate-3">
                  <svg
                    class="w-8 h-8 sm:w-10 sm:h-10 text-emerald-400 drop-shadow-sm"
                    style="animation: pulse 3s ease-in-out infinite;"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="1.5"
                      d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                    />
                  </svg>
                </div>
                <!-- Animated bottom-line -->
                <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-emerald-500 to-teal-500 group-hover:w-full transition-all duration-500 opacity-70 rounded-b-2xl" />
              </div>
            </div>
            <h3 class="text-lg sm:text-xl font-bold text-makoclaw-text mt-6 bg-gradient-to-r from-makoclaw-text to-makoclaw-text-secondary bg-clip-text">
              Start a conversation
            </h3>
            <p class="text-sm text-makoclaw-text-secondary/70 mt-2 max-w-xs text-center">
              Ask anything, run a task, or use slash commands
            </p>
            <div class="flex items-center gap-2 mt-4 text-xs text-makoclaw-text-secondary/50">
              <span class="px-2 py-1 bg-makoclaw-surface/50 rounded-lg font-mono">/task</span>
              <span class="px-2 py-1 bg-makoclaw-surface/50 rounded-lg font-mono">/help</span>
              <span class="px-2 py-1 bg-makoclaw-surface/50 rounded-lg font-mono">/search</span>
            </div>
          </div>

          <!-- Messages -->
          <div
            v-for="msg in messages"
            :key="msg.id || msg.timestamp"
            class="animate-fadeIn group w-full"
          >
            <MessageBubble
              :msg="msg"
              :current-session-id="currentSessionId"
              :is-loading="isLoading"
              :is-last-assistant-message="isLastAssistantMessage(msg)"
              @fork="forkAtMessage"
              @copy="copyMessageContent"
              @regenerate="regenerateResponse"
            />
          </div>

          <!-- Loading indicator (only when not streaming — streaming shows the message directly) -->
          <div
            v-if="isLoading && !chatStore.isStreaming"
            class="flex justify-start"
          >
            <div class="bg-makoclaw-surface/60 backdrop-blur-xl border border-makoclaw-border/50 rounded-2xl rounded-bl-md px-4 py-3 shadow-lg ring-1 ring-white/5">
              <div class="flex items-center gap-2">
                <div class="flex gap-1">
                  <div
                    class="w-2 h-2 bg-makoclaw-accent rounded-full animate-bounce"
                    style="animation-delay: 0s"
                  />
                  <div
                    class="w-2 h-2 bg-makoclaw-accent/70 rounded-full animate-bounce"
                    style="animation-delay: 0.15s"
                  />
                  <div
                    class="w-2 h-2 bg-makoclaw-accent/40 rounded-full animate-bounce"
                    style="animation-delay: 0.3s"
                  />
                </div>
                <span class="text-xs text-makoclaw-text-secondary ml-1">Thinking...</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Agent Status Indicator -->
        <div class="px-2.5 md:px-4">
          <AgentStatusIndicator />
        </div>

        <!-- Team Activity Panel -->
        <div class="px-2.5 md:px-4">
          <TeamActivityPanel
            ref="teamActivityRef"
            :agent-status="teamAgentStatus"
            :team-communications="teamCommunications"
            :involved-agents-list="involvedAgentsList"
            :specialist-report="chatStore.lastSpecialistReport"
          />
        </div>

        <!-- Input Area -->
        <div class="border-t border-makoclaw-border/30 bg-makoclaw-surface/50 backdrop-blur-2xl p-3 sm:p-4 z-20 relative ring-1 ring-white/[0.05] group/input">
          <!-- Gradient accent line at the top of input area -->
          <div class="absolute top-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent via-blue-500 to-indigo-500 group-focus-within/input:w-full transition-all duration-700 opacity-60" />

          <!-- Slash Command Autocomplete -->
          <div
            v-if="showAutocomplete"
            class="absolute bottom-full left-4 right-4 max-w-4xl mx-auto mb-2"
          >
            <div class="bg-makoclaw-surface/95 backdrop-blur-xl border border-makoclaw-border/50 rounded-2xl shadow-2xl overflow-hidden max-h-64 overflow-y-auto ring-1 ring-white/10 custom-scrollbar">
              <button
                v-for="(cmd, idx) in filteredCommands"
                :key="cmd.command"
                class="w-full text-left px-4 py-3 text-sm transition-all flex items-start gap-3 border-b border-makoclaw-border/30 last:border-0"
                :class="idx === selectedCommandIndex ? 'bg-gradient-to-r from-makoclaw-accent/15 to-transparent text-makoclaw-accent' : 'hover:bg-makoclaw-surface-hover text-makoclaw-text'"
                @click="selectCommand(cmd)"
              >
                <span class="font-mono text-xs bg-makoclaw-bg/70 px-2 py-1 rounded-lg border border-makoclaw-border/50 flex-shrink-0 mt-0.5">{{ cmd.command }}</span>
                <div>
                  <div class="font-semibold text-sm">
                    {{ cmd.label }}
                  </div>
                  <div class="text-xs text-makoclaw-text-secondary mt-0.5">
                    {{ cmd.description }}
                  </div>
                </div>
              </button>
            </div>
          </div>

          <form
            class="flex flex-col gap-1.5 sm:gap-2 max-w-4xl mx-auto w-full"
            @submit.prevent="sendMessage"
          >
            <!-- Compact Toolbar Row -->
            <div class="flex items-center justify-between px-1">
              <!-- Left group: file, prompts, tools -->
              <div class="flex items-center gap-0.5 sm:gap-1">
                <!-- Attach File -->
                <button
                  type="button"
                  :disabled="!isConnected || isLoading || uploadingFile"
                  class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-accent hover:bg-makoclaw-accent/10 transition-all disabled:opacity-40"
                  title="Attach file"
                  @click="fileInputRef?.click()"
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
                      d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"
                    />
                  </svg>
                </button>
                <!-- Prompt Library -->
                <button
                  type="button"
                  :disabled="!isConnected || isLoading"
                  class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-lg text-makoclaw-text-secondary hover:text-makoclaw-accent hover:bg-makoclaw-accent/10 transition-all disabled:opacity-40"
                  title="Prompt Library"
                  @click="showPromptLibrary = true"
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
                      d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                    />
                  </svg>
                </button>
                <!-- Tools Manager -->
                <div class="relative">
                  <button
                    type="button"
                    :class="[
                      'p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-lg transition-all relative',
                      chatStore.enabledTools.length < chatStore.availableTools.length
                        ? 'text-amber-600 hover:bg-amber-500/10'
                        : 'text-makoclaw-text-secondary hover:text-makoclaw-accent hover:bg-makoclaw-accent/10'
                    ]"
                    title="Manage AI Tools"
                    @click="showToolsPopover = !showToolsPopover"
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
                        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                      />
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                      />
                    </svg>
                    <span
                      v-if="chatStore.enabledTools.length < chatStore.availableTools.length"
                      class="absolute -top-0.5 -right-0.5 flex h-3 w-3"
                    >
                      <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75" />
                      <span class="relative inline-flex rounded-full h-3 w-3 bg-amber-500" />
                    </span>
                  </button>

                  <!-- Tools Popover (keep existing Teleport) -->
                  <Teleport to="body">
                    <div
                      v-if="showToolsPopover"
                      class="fixed inset-0 z-modal"
                      @click="showToolsPopover = false"
                    />
                    <div
                      v-if="showToolsPopover"
                      class="fixed bottom-24 right-4 md:right-auto md:left-1/2 md:-translate-x-1/2 w-72 md:w-80 bg-makoclaw-surface/95 backdrop-blur-2xl border border-makoclaw-border/50 rounded-2xl shadow-2xl z-modal-nested overflow-hidden animate-scaleIn ring-1 ring-white/10"
                    >
                      <div class="p-4 border-b border-makoclaw-border/30 bg-gradient-to-r from-makoclaw-surface/50 to-transparent">
                        <div class="flex items-center gap-2">
                          <div class="w-7 h-7 rounded-lg bg-gradient-to-br from-makoclaw-accent/20 to-purple-500/20 flex items-center justify-center">
                            <svg
                              class="w-3.5 h-3.5 text-makoclaw-accent"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                              />
                            </svg>
                          </div>
                          <div>
                            <h3 class="text-sm font-bold text-makoclaw-text">
                              AI Tools
                            </h3>
                            <p class="text-[10px] text-makoclaw-text-secondary">
                              {{ chatStore.enabledTools.length }}/{{ chatStore.availableTools.length }} enabled
                            </p>
                          </div>
                        </div>
                      </div>
                      <div class="max-h-64 overflow-y-auto p-2 space-y-1 custom-scrollbar">
                        <div
                          v-if="chatStore.availableTools.length === 0"
                          class="flex flex-col items-center justify-center py-6"
                        >
                          <div class="w-8 h-8 border-2 border-makoclaw-accent/30 border-t-makoclaw-accent rounded-full animate-spin mb-2" />
                          <p class="text-xs text-makoclaw-text-secondary">
                            Loading tools...
                          </p>
                        </div>
                        <button
                          v-for="tool in chatStore.availableTools"
                          :key="tool"
                          class="w-full flex items-center justify-between px-3 py-2.5 rounded-xl text-xs transition-all"
                          :class="chatStore.enabledTools.includes(tool) ? 'bg-gradient-to-r from-makoclaw-accent/15 to-transparent text-makoclaw-accent' : 'hover:bg-makoclaw-surface-hover text-makoclaw-text-secondary'"
                          @click="chatStore.toggleTool(tool)"
                        >
                          <div class="flex items-center gap-2.5">
                            <div
                              class="w-5 h-5 rounded-md border-2 flex items-center justify-center transition-all"
                              :class="chatStore.enabledTools.includes(tool) ? 'bg-makoclaw-accent border-makoclaw-accent text-white' : 'border-makoclaw-border/50 bg-makoclaw-bg/30'"
                            >
                              <svg
                                v-if="chatStore.enabledTools.includes(tool)"
                                class="w-3 h-3"
                                fill="currentColor"
                                viewBox="0 0 20 20"
                              ><path
                                fill-rule="evenodd"
                                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                                clip-rule="evenodd"
                              /></svg>
                            </div>
                            <span class="font-mono text-xs">{{ tool }}</span>
                          </div>
                        </button>
                      </div>
                    </div>
                  </Teleport>
                </div>
              </div>

              <!-- Right group: mic, send/stop -->
              <div class="flex items-center gap-0.5 sm:gap-1">
                <!-- Mic Button -->
                <button
                  type="button"
                  :disabled="!isConnected || isLoading || isTranscribing"
                  :class="[
                    'p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-lg transition-all',
                    isRecording
                      ? 'bg-makoclaw-error text-white shadow-lg shadow-makoclaw-error/30 animate-pulse'
                      : isTranscribing
                        ? 'text-makoclaw-text-secondary cursor-wait opacity-50'
                        : 'text-makoclaw-text-secondary hover:text-makoclaw-accent hover:bg-makoclaw-accent/10'
                  ]"
                  :title="isRecording ? 'Suelta para transcribir' : isTranscribing ? 'Transcribiendo...' : 'Mantén presionado para grabar'"
                  @mousedown="startRecording"
                  @mouseup="stopRecording"
                  @mouseleave="stopRecording"
                  @touchstart.prevent="startRecording"
                  @touchend.prevent="stopRecording"
                >
                  <svg
                    v-if="isTranscribing"
                    class="w-4 h-4 animate-spin"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      class="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      stroke-width="4"
                    />
                    <path
                      class="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                    />
                  </svg>
                  <svg
                    v-else
                    class="w-4 h-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"
                    />
                  </svg>
                </button>
                <!-- Send Button -->
                <button
                  v-if="!isLoading"
                  type="submit"
                  :disabled="!isConnected || !messageInput.trim()"
                  class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-lg bg-makoclaw-accent hover:bg-makoclaw-accent-hover disabled:bg-makoclaw-surface disabled:text-makoclaw-text-secondary text-white transition-all shadow-md shadow-makoclaw-accent/20 hover:shadow-makoclaw-accent/40"
                  title="Enviar mensaje"
                >
                  <svg
                    class="w-4 h-4 transform rotate-90"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
                    />
                  </svg>
                </button>
                <!-- Stop Button -->
                <button
                  v-else
                  type="button"
                  class="p-2 min-h-[36px] min-w-[36px] flex items-center justify-center rounded-lg bg-makoclaw-error/10 hover:bg-makoclaw-error/20 text-makoclaw-error transition-all"
                  title="Detener agente"
                  @click="cancelExecution"
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
                      d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"
                    />
                  </svg>
                </button>
              </div>
            </div>

            <!-- File Attachment Preview Strip -->
            <div
              v-if="attachments.length > 0"
              class="flex flex-wrap gap-2 px-1"
            >
              <div
                v-for="(att, idx) in attachments"
                :key="idx"
                class="flex items-center gap-2 bg-makoclaw-bg border border-makoclaw-border rounded-lg px-2.5 py-1.5 text-xs max-w-[200px]"
              >
                <svg
                  class="w-3.5 h-3.5 text-makoclaw-accent flex-shrink-0"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                ><path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                /></svg>
                <span class="truncate text-makoclaw-text flex-1">{{ att.name }}</span>
                <button
                  type="button"
                  class="text-makoclaw-text-secondary hover:text-makoclaw-error transition-colors flex-shrink-0"
                  @click="removeAttachment(idx)"
                >
                  <svg
                    class="w-3 h-3"
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
              <div
                v-if="uploadingFile"
                class="flex items-center gap-1.5 px-2.5 py-1.5 text-xs text-makoclaw-text-secondary"
              >
                <div class="w-3 h-3 border-2 border-makoclaw-accent border-t-transparent rounded-full animate-spin" />
                Uploading...
              </div>
            </div>

            <!-- Hidden file input -->
            <input
              ref="fileInputRef"
              type="file"
              class="hidden"
              accept=".txt,.md,.json,.csv,.html,.xml,.yaml,.yml,.py,.go,.js,.ts,.java,.c,.cpp,.h,.cs,.rb,.rs,.php,.log,.pdf"
              @change="handleFileAttach"
            >

            <!-- Textarea (full width) -->
            <textarea
              ref="chatInput"
              v-model="messageInput"
              placeholder="Type a message or / for commands..."
              rows="1"
              class="w-full px-4 py-3 bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl focus:ring-2 focus:ring-makoclaw-accent/30 focus:border-makoclaw-accent/50 transition-all text-sm shadow-inner resize-none overflow-hidden backdrop-blur-sm placeholder:text-makoclaw-text-secondary/50"
              :disabled="!isConnected || isLoading"
              style="max-height: 120px;"
              @input="onInputChange"
              @keydown="onInputKeydown"
            />
          </form>

          <!-- Connection Status -->
          <div class="absolute top-0 right-0 -mt-9 mr-4 px-2.5 py-1 rounded-lg text-[10px] font-medium bg-makoclaw-surface/80 backdrop-blur-xl border border-makoclaw-border/30 shadow-sm">
            <span
              v-if="isConnected"
              class="text-makoclaw-success flex items-center gap-1.5"
            >
              <span class="relative flex h-2 w-2">
                <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-makoclaw-success opacity-75" />
                <span class="relative inline-flex rounded-full h-2 w-2 bg-makoclaw-success" />
              </span>
              Connected
            </span>
            <span
              v-else
              class="text-makoclaw-error flex items-center gap-1.5"
            >
              <span class="w-2 h-2 rounded-full bg-makoclaw-error" />
              Disconnected
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Prompt Library Modal -->
  <PromptLibrary
    :show="showPromptLibrary"
    @close="showPromptLibrary = false"
    @use="insertPrompt"
  />
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { storeToRefs } from 'pinia'
import MessageBubble from '../components/MessageBubble.vue'
import AgentStatusIndicator from '../components/Chat/AgentStatusIndicator.vue'
import SpecialistsPanel from '../components/Chat/SpecialistsPanel.vue'
import TeamActivityPanel from '../components/Chat/TeamActivityPanel.vue'
import PromptLibrary from '../components/PromptModal.vue'
import { useChatStore } from '../stores/chatStore'
import { getChatWebSocket } from '../services/websocketService'
import taskService from '../services/taskService'
import advancedService from '../services/advancedService'
import { useRoute, useRouter } from 'vue-router'
import { useToast } from '../composables/useToast'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const chatStore = useChatStore()
const messagesContainer = ref(null)
const messageInput = ref('')
const isConnected = ref(false)
const isLoading = ref(false)
const showSidebar = ref(localStorage.getItem('chat.sidebar') !== 'false')
const sessions = ref([])

const toggleSidebar = () => {
  showSidebar.value = !showSidebar.value
  localStorage.setItem('chat.sidebar', showSidebar.value)
}
const currentSessionId = ref(null)
const contextMenu = ref({ show: false, sessionId: null, x: 0, y: 0 })
const renamingSession = ref(null)
const renameInput = ref('')
const showToolsPopover = ref(false)
const showPromptLibrary = ref(false)

// Team activity state
const teamActivityRef = ref(null)
const teamAgentStatus = ref(null)
const teamCommunications = ref([])
const involvedAgentsList = ref([])

// File attachments state
const fileInputRef = ref(null)
const attachments = ref([])
const uploadingFile = ref(false)

const handleFileAttach = async (e) => {
  const file = e.target.files?.[0]
  if (!file) return
  uploadingFile.value = true
  try {
    const result = await advancedService.uploadChatAttachment(file)
    attachments.value.push({ name: result.name, content: result.content, truncated: result.truncated })
    if (result.truncated) toast.error('File was truncated to 50,000 characters')
  } catch (err) {
    toast.error('Failed to attach file: ' + (err.response?.data?.error || err.message))
  } finally {
    uploadingFile.value = false
    if (fileInputRef.value) fileInputRef.value.value = ''
  }
}

const removeAttachment = (idx) => {
  attachments.value.splice(idx, 1)
}

const insertPrompt = (content) => {
  messageInput.value = content
  nextTick(() => {
    if (chatInput.value) {
      chatInput.value.style.height = 'auto'
      chatInput.value.style.height = Math.min(chatInput.value.scrollHeight, 120) + 'px'
      chatInput.value.focus()
    }
  })
}

const { messages } = storeToRefs(chatStore)
const chatWs = getChatWebSocket()
const chatInput = ref(null)

// Voice recording state
const isRecording = ref(false)
const isTranscribing = ref(false)
let mediaRecorder = null
let audioChunks = []

// Slash command autocomplete
const showAutocomplete = ref(false)
const selectedCommandIndex = ref(0)
const slashCommands = [
  { command: '/task create', label: 'Create Task', description: 'Create a new task for the agent to work on' },
  { command: '/task list', label: 'List Tasks', description: 'Show all current tasks and their status' },
  { command: '/task run', label: 'Run Task', description: 'Execute a specific task by ID — /task run <id>' },
  { command: '/task move', label: 'Move Task', description: 'Change task status — /task move <id> <status>' },
  { command: '/list', label: 'List (shortcut)', description: 'Alias for /task list' },
  { command: '/archive', label: 'Archive Task', description: 'Archive a task by ID — /archive <id>' },
  { command: '/help', label: 'Help', description: 'Ask the agent for help with available commands' },
  { command: '/summarize', label: 'Summarize', description: 'Ask the agent to summarize recent activity' },
  { command: '/search', label: 'Search', description: 'Search through conversation history' },
]

const filteredCommands = computed(() => {
  const input = messageInput.value.trim().toLowerCase()
  if (!input.startsWith('/')) return []
  return slashCommands.filter(cmd => cmd.command.startsWith(input))
})

const onInputChange = () => {
  const input = messageInput.value.trim()
  if (input.startsWith('/') && input.length >= 1) {
    showAutocomplete.value = filteredCommands.value.length > 0
    selectedCommandIndex.value = 0
  } else {
    showAutocomplete.value = false
  }
  // Auto-resize textarea
  if (chatInput.value) {
    chatInput.value.style.height = 'auto'
    chatInput.value.style.height = Math.min(chatInput.value.scrollHeight, 120) + 'px'
  }
}

const onInputKeydown = (e) => {
  if (showAutocomplete.value) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      selectedCommandIndex.value = (selectedCommandIndex.value + 1) % filteredCommands.value.length
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      selectedCommandIndex.value = (selectedCommandIndex.value - 1 + filteredCommands.value.length) % filteredCommands.value.length
    } else if (e.key === 'Tab' || (e.key === 'Enter' && !e.shiftKey)) {
      e.preventDefault()
      selectCommand(filteredCommands.value[selectedCommandIndex.value])
    } else if (e.key === 'Escape') {
      showAutocomplete.value = false
    }
  } else if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
}

const selectCommand = (cmd) => {
  messageInput.value = cmd.command + ' '
  showAutocomplete.value = false
  nextTick(() => chatInput.value?.focus())
}

const fetchSessions = async () => {
  try {
    const data = await taskService.fetchChatSessions()
    // Explicitly filter out tasks: anything starting with 'task:' or containing ':task:'
    sessions.value = (data.sessions || []).filter(s => {
      const id = s.session_id || ''
      return !id.startsWith('task:') && !id.includes(':task:')
    })
    
    // If current session was a task and is now filtered out, reset view
    const currentId = currentSessionId.value || ''
    if (currentId && (currentId.startsWith('task:') || currentId.includes(':task:'))) {
        startNewChat()
    }
  } catch (error) {
    console.error('Failed to fetch sessions:', error)
  }
}

const normalizeSessionId = (value) => {
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : ''
  }
  return typeof value === 'string' ? value : ''
}

const loadSession = async (sessionId, options = { updateRoute: true }) => {
  const normalizedSessionId = normalizeSessionId(sessionId)
  if (!normalizedSessionId || currentSessionId.value === normalizedSessionId) return
  currentSessionId.value = normalizedSessionId
  if (options.updateRoute) {
    router.replace({ query: { id: normalizedSessionId } })
  }

  // Reset team activity for new session
  teamAgentStatus.value = null
  teamCommunications.value = []
  involvedAgentsList.value = []
  if (teamActivityRef.value?.reset) {
    teamActivityRef.value.reset()
  }

  try {
    const data = await taskService.fetchSessionMessages(normalizedSessionId)
    chatStore.setMessages((data.messages || []).map(m => {
      // Parse metadata to extract agents
      let agents = []
      if (m.metadata && typeof m.metadata === 'string') {
        try {
          const meta = JSON.parse(m.metadata)
          if (meta.agents && Array.isArray(meta.agents)) {
            agents = meta.agents
          }
        } catch (e) {
          // Ignore parse errors
        }
      }
      return {
        ...m,
        timestamp: m.created_at, // Normalize timestamp
        agents: agents
      }
    }))
    // Close sidebar on mobile
    showSidebar.value = false
  } catch (error) {
    console.error('Failed to load session:', error)
  }
}

const startNewChat = () => {
  currentSessionId.value = null
  if (route.query.id) {
    router.replace({ query: {} })
  }
  chatStore.clearMessages()
  showSidebar.value = false
  // Reset team activity for new chat
  teamAgentStatus.value = null
  teamCommunications.value = []
  involvedAgentsList.value = []
  if (teamActivityRef.value?.reset) {
    teamActivityRef.value.reset()
  }
  // Focus input
  // nextTick(() => document.querySelector('input')?.focus())
}

// Session context menu
const openContextMenu = (e, sessionId) => {
  e.preventDefault()
  e.stopPropagation()
  contextMenu.value = { show: true, sessionId, x: e.clientX, y: e.clientY }
}

const closeContextMenu = () => {
  contextMenu.value = { show: false, sessionId: null, x: 0, y: 0 }
}

const startRename = (sessionId) => {
  const session = sessions.value.find(s => s.session_id === sessionId)
  renameInput.value = session?.title || ''
  renamingSession.value = sessionId
  closeContextMenu()
}

const submitRename = async (sessionId) => {
  if (renamingSession.value !== sessionId) return
  try {
    await taskService.updateSession(sessionId, { title: renameInput.value.trim() })
    const session = sessions.value.find(s => s.session_id === sessionId)
    if (session) session.title = renameInput.value.trim()
    toast.success('Sesión renombrada')
  } catch (error) {
    console.error('Failed to rename session:', error)
    toast.error('Error al renombrar la sesión')
  }
  renamingSession.value = null
  renameInput.value = ''
}

const cancelRename = () => {
  renamingSession.value = null
  renameInput.value = ''
}

const archiveSession = async (sessionId) => {
  closeContextMenu()
  try {
    await taskService.updateSession(sessionId, { archived: true })
    sessions.value = sessions.value.filter(s => s.session_id !== sessionId)
    if (currentSessionId.value === sessionId) {
      startNewChat()
    }
    toast.success('Sesión archivada')
  } catch (error) {
    console.error('Failed to archive session:', error)
    toast.error('Error al archivar la sesión')
  }
}

const deleteSession = async (sessionId) => {
  closeContextMenu()
  if (!confirm('¿Eliminar esta sesión y todos sus mensajes? Esta acción no se puede deshacer.')) return
  try {
    await taskService.deleteSession(sessionId)
    sessions.value = sessions.value.filter(s => s.session_id !== sessionId)
    if (currentSessionId.value === sessionId) {
      startNewChat()
    }
    toast.success('Sesión eliminada')
  } catch (error) {
    console.error('Failed to delete session:', error)
    toast.error('Error al eliminar la sesión')
  }
}

const generateSessionId = () => {
  return 'web:chat:' + Date.now().toString(36) + Math.random().toString(36).substr(2)
}

const copyMessageContent = async (content) => {
  try {
    await navigator.clipboard.writeText(content)
    toast.success('Copiado al portapapeles')
  } catch {
    toast.error('Error al copiar')
  }
}

const handleMessage = (message) => {
  if (message.type === 'message') {
    chatStore.addMessage({
      role: message.role || 'assistant',
      content: message.content,
      timestamp: new Date().toISOString(),
      agents: message.agents || []
    })
    // Refresh sessions to show latest message/time
    fetchSessions()
  }
  if (message.type === 'stream_start') {
    chatStore.startStreamingMessage()
    chatStore.clearAgentStatus() // Resetear para nuevo mensaje
    // Reset team activity for new message
    teamAgentStatus.value = null
    teamCommunications.value = []
    involvedAgentsList.value = []
  }
  if (message.type === 'stream') {
    chatStore.appendStreamToken(message.content || '')
  }
  if (message.type === 'stream_end') {
    if (message.error) {
      chatStore.endStreamingMessage(`Error: ${message.error}`, [])
    } else {
      chatStore.endStreamingMessage(message.content || '', message.agents || [])
    }
    chatStore.clearAgentStatus() // Limpiar cuando termina
    fetchSessions()
  }
  if (message.type === 'agent_status') {
    chatStore.setAgentStatus(
      message.agent,
      message.status,
      message.specialist_name,
      message.reason
    )
    // Update team activity panel
    teamAgentStatus.value = {
      agent: message.agent,
      status: message.status,
      specialistName: message.specialist_name,
      reason: message.reason
    }
    // Track involved agents
    if (message.agent && !involvedAgentsList.value.includes(message.agent)) {
      involvedAgentsList.value.push(message.agent)
    }
    // Track inter-specialist communications
    if (message.status === 'requesting_help' && message.specialist_name) {
      teamCommunications.value.push({
        from: message.agent,
        to: message.specialist_name,
        message: message.reason || 'Requesting assistance',
        timestamp: new Date().toISOString()
      })
    }
    if (message.status === 'colleague_complete') {
      teamCommunications.value.push({
        from: message.agent,
        to: 'orchestrator',
        message: 'Completed assistance',
        timestamp: new Date().toISOString()
      })
    }
  }
  if (message.type === 'content_segment') {
    chatStore.addContentSegment(message)
  }
  if (message.type === 'tool_call') {
    chatStore.addToolCall(message)
  }
  if (message.type === 'specialist_report') {
    // Handle structured specialist report
    chatStore.addSpecialistReport(message)

    // Update team activity panel with report info
    teamAgentStatus.value = {
      agent: message.specialist_name,
      status: message.status,
      confidence: message.confidence,
      requestHelp: message.request_help
    }

    // Track involved agents
    if (message.specialist_name && !involvedAgentsList.value.includes(message.specialist_name)) {
      involvedAgentsList.value.push(message.specialist_name)
    }

    // If specialist needs help, add to communications
    if (message.status === 'needs_help' && message.request_help) {
      teamCommunications.value.push({
        from: message.specialist_name,
        to: message.request_help,
        message: message.suggestions?.[0] || message.help_context || 'Requesting assistance',
        timestamp: new Date().toISOString(),
        type: 'help_request'
      })
    }
  }
  if (message.type === 'ready') {
    isLoading.value = false
    chatStore.setGlobalLoading(false) // Clear global loading state when response is ready
  }
}

const handleDisconnected = () => {
  isConnected.value = false
  chatStore.setConnected(false)
}

const handleConnected = () => {
  isConnected.value = true
  chatStore.setConnected(true)
}

onMounted(async () => {
  // Tell MainLayout's background listener to yield — we're handling messages now
  window.__makoclaw_setChatViewActive?.(true)

  await fetchSessions()

  // Fetch available models
  try {
    const modelsData = await advancedService.fetchModels()
    chatStore.setModelsData(modelsData)
  } catch (error) {
    console.error('Failed to fetch models:', error)
    chatStore.setModelsData({ current_model: '', providers: [] })
  }

  // Fetch available tools
  try {
    const toolsData = await advancedService.fetchTools()
    chatStore.setAvailableTools(toolsData.tools || [])
  } catch (error) {
    console.error('Failed to fetch tools:', error)
  }
  
  // Determine which session to show:
  // 1. Route param, 2. Previously active session (from store), 3. Most recent session
  const routeSessionId = normalizeSessionId(route.query.id)
  const storedSessionId = chatStore.activeSessionId
  if (routeSessionId) {
    const routeSessionExists = sessions.value.some(s => s.session_id === routeSessionId)
    if (routeSessionExists) {
      await loadSession(routeSessionId, { updateRoute: false })
    } else if (sessions.value.length > 0) {
      await loadSession(sessions.value[0].session_id)
    }
  } else if (storedSessionId && sessions.value.some(s => s.session_id === storedSessionId)) {
    // Restore session that was active before navigation
    await loadSession(storedSessionId, { updateRoute: true })
  } else if (sessions.value.length > 0) {
    await loadSession(sessions.value[0].session_id)
  }

  try {
    await chatWs.connect()
    isConnected.value = true
    chatStore.setConnected(true)

    // Listen for messages
    chatWs.on('message', handleMessage)

    // Listen for connection events
    chatWs.on('disconnected', handleDisconnected)
    chatWs.on('connected', handleConnected)

    chatStore.setWebSocket(chatWs)
  } catch (error) {
    console.error('Failed to connect to chat:', error)
  }

  // Flush any messages that arrived while ChatView was not mounted
  const pending = chatStore.flushPendingMessages()
  if (pending.length > 0) {
    // Sync isLoading based on whether agent was still working
    const hasReadyEvent = pending.some(m => m.type === 'ready')
    if (!hasReadyEvent) {
      isLoading.value = true
    }
    for (const msg of pending) {
      handleMessage(msg)
    }
    // After processing, scroll to bottom
    await nextTick()
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  }
})

onBeforeUnmount(() => {
  // Remove listeners to prevent duplicates, but DON'T disconnect
  // This allows the agent to continue working even when navigating away from chat
  chatWs.off('message', handleMessage)
  chatWs.off('disconnected', handleDisconnected)
  chatWs.off('connected', handleConnected)
  // Tell MainLayout background listener to resume capturing messages
  window.__makoclaw_setChatViewActive?.(false)
})

// Auto-scroll to bottom
watch(messages, async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}, { deep: true })

watch(() => route.query.id, (newId) => {
  const normalizedId = normalizeSessionId(newId)
  if (normalizedId && normalizedId !== currentSessionId.value) {
    loadSession(normalizedId, { updateRoute: false })
  } else if (!newId) {
    startNewChat()
  }
})

const sendMessage = async () => {
  const content = messageInput.value.trim()
  if (!content) return

  showAutocomplete.value = false

  // Generate session ID if new
  if (!currentSessionId.value) {
    currentSessionId.value = generateSessionId()
    router.replace({ query: { id: currentSessionId.value } })
  }

  // Parse @specialist_name mentions for direct invocation
  // Syntax: @specialist_name message content
  const mentionMatch = content.match(/^@(\w+)\s+(.+)$/s)
  let targetSpecialist = null
  let messageContent = content

  if (mentionMatch) {
    targetSpecialist = mentionMatch[1]
    messageContent = mentionMatch[2]
  }

  // Add attachment content to message if present
  let finalContent = messageContent
  if (attachments.value.length > 0) {
    const attachmentText = attachments.value.map(a =>
      `\n\n--- Attached file: ${a.name} ---\n${a.content}\n--- End of ${a.name} ---`
    ).join('')
    finalContent = messageContent + attachmentText
    attachments.value = []
  }

  // Add user message locally (show original content with @mention)
  chatStore.addMessage({
    role: 'user',
    content: content,
    timestamp: new Date().toISOString()
  })

  messageInput.value = ''
  isLoading.value = true
  chatStore.setIsWorking(true)  // Persist loading state for background navigation
  chatStore.setActiveSessionId(currentSessionId.value) // Persist active session for restoration

  // Reset textarea height
  if (chatInput.value) {
    chatInput.value.style.height = 'auto'
  }

  // Send via WebSocket
  if (chatWs.isConnected()) {
    const excludeTools = (chatStore.availableTools || []).filter(
      tool => !(chatStore.enabledTools || []).includes(tool)
    )

    chatWs.send({
      type: 'message',
      content: finalContent,
      session_id: currentSessionId.value,
      model: chatStore.selectedModel || undefined,
      web_search: chatStore.webSearchEnabled,
      exclude_tools: excludeTools,
      target_specialist: targetSpecialist  // Direct specialist invocation
    })
    // Refresh sessions to show new thread
    setTimeout(fetchSessions, 500)
  } else {
    isLoading.value = false
    chatStore.setGlobalLoading(false)
  }
}

// Cancel current execution
const cancelExecution = async () => {
  if (!currentSessionId.value) return
  
  try {
    const response = await fetch('/api/v1/chat/cancel', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('auth.token')}`
      },
      body: JSON.stringify({
        session_id: currentSessionId.value
      })
    })
    
    if (response.ok) {
      toast.success('Execution canceled')
      isLoading.value = false
      chatStore.setGlobalLoading(false)
    } else {
      toast.error('Failed to cancel execution')
    }
  } catch (error) {
    console.error('Failed to cancel execution:', error)
    toast.error('Failed to cancel execution')
  }
}

const isLastAssistantMessage = (msg) => {
  const assistantMsgs = messages.value.filter(m => m.role === 'assistant')
  if (assistantMsgs.length === 0) return false
  const last = assistantMsgs[assistantMsgs.length - 1]
  return (msg.id && msg.id === last.id) || (msg.timestamp && msg.timestamp === last.timestamp)
}

const regenerateResponse = async () => {
  // Find the last user message
  const userMsgs = messages.value.filter(m => m.role === 'user')
  if (userMsgs.length === 0) return

  const lastUserMsg = userMsgs[userMsgs.length - 1]

  // Remove the last assistant message
  const lastAssistantIdx = messages.value.map(m => m.role).lastIndexOf('assistant')
  if (lastAssistantIdx >= 0) {
    messages.value.splice(lastAssistantIdx, 1)
  }

  isLoading.value = true

  // Resend the last user message
  if (chatWs.isConnected()) {
    chatWs.send({
      type: 'message',
      content: lastUserMsg.content,
      session_id: currentSessionId.value,
      model: chatStore.selectedModel || undefined,
      web_search: chatStore.webSearchEnabled
    })
  } else {
    isLoading.value = false
  }
}

const forkAtMessage = async (msg) => {
  if (!currentSessionId.value || !msg.id) return

  if (!confirm('¿Ramificar conversación desde este mensaje? Se creará una nueva sesión con todos los mensajes hasta este punto.')) return

  try {
    const result = await advancedService.forkChat(currentSessionId.value, msg.id)
    toast.success(`¡Conversación ramificada! Nueva sesión creada con ${result.messages_copied} mensaje(s)`)
    // Navigate to the forked session
    router.push({ query: { id: result.new_session_id } })
    await fetchSessions() // Refresh sessions to show the new one
  } catch (error) {
    console.error('Fork failed:', error)
    toast.error('Error al ramificar la conversación')
  }
}

// Voice recording
const startRecording = async () => {
  if (isRecording.value || isTranscribing.value) return

  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    audioChunks = []
    mediaRecorder = new MediaRecorder(stream, { mimeType: getMimeType() })

    mediaRecorder.ondataavailable = (e) => {
      if (e.data.size > 0) audioChunks.push(e.data)
    }

    mediaRecorder.onstop = async () => {
      // Stop all tracks to release the mic
      stream.getTracks().forEach(track => track.stop())

      if (audioChunks.length === 0) return

      const audioBlob = new Blob(audioChunks, { type: mediaRecorder.mimeType })
      isTranscribing.value = true

      try {
        const result = await advancedService.transcribeAudio(audioBlob)
        if (result.text && result.text.trim()) {
          messageInput.value = (messageInput.value ? messageInput.value + ' ' : '') + result.text.trim()
          // Auto-resize textarea
          nextTick(() => {
            if (chatInput.value) {
              chatInput.value.style.height = 'auto'
              chatInput.value.style.height = Math.min(chatInput.value.scrollHeight, 120) + 'px'
            }
          })
        }
      } catch (error) {
        console.error('Transcription failed:', error)
        const errMsg = error.response?.data?.error || 'Transcription failed'
        chatStore.addMessage({
          role: 'assistant',
          content: `Voice transcription error: ${errMsg}`,
          timestamp: new Date().toISOString()
        })
      } finally {
        isTranscribing.value = false
      }
    }

    mediaRecorder.start()
    isRecording.value = true
  } catch (error) {
    console.error('Microphone access denied:', error)
    chatStore.addMessage({
      role: 'assistant',
      content: 'Microphone access denied. Please allow microphone access in your browser settings.',
      timestamp: new Date().toISOString()
    })
  }
}

const stopRecording = () => {
  if (!isRecording.value || !mediaRecorder) return
  isRecording.value = false
  mediaRecorder.stop()
  mediaRecorder = null
}

const getMimeType = () => {
  const types = ['audio/webm;codecs=opus', 'audio/webm', 'audio/ogg;codecs=opus', 'audio/mp4']
  for (const type of types) {
    if (MediaRecorder.isTypeSupported(type)) return type
  }
  return 'audio/webm' // Fallback
}

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-fadeIn {
  animation: fadeIn 0.3s ease-in;
}

/* Streaming cursor */
@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

.streaming-cursor {
  display: inline-block;
  width: 2px;
  height: 1em;
  background: currentColor;
  margin-left: 1px;
  vertical-align: text-bottom;
  animation: blink 0.8s ease-in-out infinite;
}

</style>



<style scoped>
@keyframes float {
  0%, 100% { transform: translateY(0) translateX(0); }
  50% { transform: translateY(-20px) translateX(10px); }
}
</style>
