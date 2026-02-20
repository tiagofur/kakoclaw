<template>
  <div class="flex h-full bg-kakoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh (Subtle) -->
    <div class="absolute inset-0 pointer-events-none opacity-20 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-kakoclaw-accent/30 via-transparent to-transparent"></div>

    <!-- Sidebar (History) -->
    <div 
      :class="[
        'flex-shrink-0 border-r border-kakoclaw-border bg-kakoclaw-surface/50 backdrop-blur-md transition-all duration-500 ease-[cubic-bezier(0.4,0,0.2,1)] flex flex-col',
        showSidebar ? 'w-56 md:w-64 opacity-100' : 'w-0 opacity-0 border-none overflow-hidden scale-95 origin-left'
      ]"
    >
      <div class="p-2 md:p-4 border-b border-kakoclaw-border flex justify-between items-center gap-2">
        <h2 class="font-semibold text-xs md:text-sm text-kakoclaw-text-secondary">History</h2>
        <button @click="startNewChat" class="p-1 md:p-1.5 hover:bg-kakoclaw-bg rounded-md text-kakoclaw-accent transition-colors flex-shrink-0" title="Nuevo Chat">
          <svg class="w-4 md:w-5 h-4 md:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
        </button>
      </div>
      
      <div class="flex-1 overflow-y-auto p-1.5 md:p-2 space-y-0.5 md:space-y-1">
        <div v-if="sessions.length === 0" class="text-[10px] md:text-xs text-kakoclaw-text-secondary text-center py-4">
          No history yet
        </div>
        <div
          v-for="session in sessions"
          :key="session.session_id"
          class="relative group"
        >
          <!-- Inline rename -->
          <div v-if="renamingSession === session.session_id" class="flex items-center gap-1 px-2 py-1">
            <input
              v-model="renameInput"
              @keyup.enter="submitRename(session.session_id)"
              @keyup.escape="cancelRename"
              @blur="submitRename(session.session_id)"
              class="flex-1 bg-kakoclaw-bg border border-kakoclaw-accent rounded px-2 py-1 text-xs text-kakoclaw-text focus:outline-none"
              autofocus
              placeholder="Session title..."
            />
          </div>
          <!-- Normal session button -->
          <button
            v-else
            @click="loadSession(session.session_id)"
            :class="[
              'w-full text-left px-2 md:px-3 py-1.5 md:py-2 rounded-lg text-xs md:text-sm transition-colors',
              currentSessionId === session.session_id ? 'bg-kakoclaw-accent/10 text-kakoclaw-accent' : 'hover:bg-kakoclaw-bg text-kakoclaw-text-secondary hover:text-kakoclaw-text'
            ]"
          >
            <div class="flex items-center gap-2">
              <svg v-if="session.session_id.startsWith('web:task:')" class="w-3 md:w-3.5 h-3 md:h-3.5 text-amber-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" /></svg>
              <svg v-else class="w-3 md:w-3.5 h-3 md:h-3.5 text-emerald-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
              <span class="truncate flex-1 text-[11px] md:text-sm">{{ session.title || session.last_message || 'Empty session' }}</span>
              <!-- Context menu trigger -->
              <button
                @click.stop="openContextMenu($event, session.session_id)"
                class="opacity-0 group-hover:opacity-100 p-0.5 hover:bg-kakoclaw-border rounded transition-opacity flex-shrink-0"
                title="Acciones de sesión"
              >
                <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20"><circle cx="10" cy="4" r="1.5"/><circle cx="10" cy="10" r="1.5"/><circle cx="10" cy="16" r="1.5"/></svg>
              </button>
            </div>
            <div class="text-[9px] md:text-[10px] opacity-60 mt-0.5 pl-4 md:pl-5 flex justify-between">
              <span>{{ formatTime(session.updated_at) }}</span>
              <span v-if="session.message_count" class="text-kakoclaw-text-secondary">{{ session.message_count }} msg{{ session.message_count !== 1 ? 's' : '' }}</span>
            </div>
          </button>
        </div>
      </div>

      <!-- Context Menu Overlay -->
      <Teleport to="body">
        <div v-if="contextMenu.show" class="fixed inset-0 z-50" @click="closeContextMenu">
          <div
            class="absolute bg-kakoclaw-surface border border-kakoclaw-border rounded-lg shadow-xl py-1 min-w-[140px]"
            :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
            @click.stop
          >
            <button @click="startRename(contextMenu.sessionId)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-kakoclaw-bg transition-colors flex items-center gap-2">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
              Rename
            </button>
            <button @click="archiveSession(contextMenu.sessionId)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-kakoclaw-bg transition-colors flex items-center gap-2">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" /></svg>
              Archive
            </button>
            <div class="border-t border-kakoclaw-border my-1"></div>
            <button @click="deleteSession(contextMenu.sessionId)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-kakoclaw-error/10 text-kakoclaw-error transition-colors flex items-center gap-2">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
              Delete
            </button>
          </div>
        </div>
      </Teleport>
    </div>

    <!-- Main Chat Area -->
    <div class="flex-1 flex flex-col min-w-0 relative bg-kakoclaw-bg/50">
      <!-- Top Bar: Mobile toggle + Model selector -->
      <div class="flex items-center justify-between px-2 md:px-4 py-1.5 md:py-2 border-b border-kakoclaw-border/30 bg-kakoclaw-surface/30 backdrop-blur-sm z-20 gap-2 flex-wrap">
        <div class="flex items-center gap-2">
          <button 
            @click="toggleSidebar"
            class="p-2 hover:bg-kakoclaw-accent/10 rounded-xl text-kakoclaw-text-secondary hover:text-kakoclaw-accent transition-all duration-300 glass border border-transparent hover:border-kakoclaw-accent/30 flex items-center justify-center group"
            title="Toggle Sidebar"
          >
            <svg class="w-5 h-5 transition-transform duration-500" :class="{'rotate-180': !showSidebar}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
            </svg>
          </button>
          <div class="hidden md:block"></div>
        </div>

        <!-- Global Loading Indicator -->
        <div v-if="chatStore.globalIsLoading" class="flex items-center gap-1.5 md:gap-2 px-2 md:px-3 py-1 md:py-1.5 bg-kakoclaw-accent/10 border border-kakoclaw-accent rounded-lg order-3 md:order-2 w-full md:w-auto text-center md:text-left">
          <svg class="w-3.5 md:w-4 h-3.5 md:h-4 text-kakoclaw-accent animate-spin flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          <span class="text-[10px] md:text-xs font-medium text-kakoclaw-accent">Agent working...</span>
        </div>

        <!-- Model Selector -->
        <div class="flex items-center gap-1 md:gap-2 flex-shrink-0 order-2 md:order-3">
          <svg class="w-3.5 md:w-4 h-3.5 md:h-4 text-kakoclaw-text-secondary flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
          <select
            v-model="chatStore.selectedModel"
            :disabled="chatStore.allModels.length === 0"
            class="bg-kakoclaw-bg/50 border border-kakoclaw-border rounded-lg px-2 md:px-3 py-1 md:py-1.5 text-[10px] md:text-xs text-kakoclaw-text focus:ring-2 focus:ring-kakoclaw-accent/50 focus:border-kakoclaw-accent transition-all cursor-pointer max-w-[180px] md:max-w-[280px]"
          >
            <option v-if="chatStore.allModels.length === 0" value="">
              No models available
            </option>
            <option v-for="model in chatStore.allModels" :key="model.id" :value="model.id">
              {{ model.provider }}/{{ model.label }}{{ model.isDefault ? ' (default)' : '' }}
            </option>
          </select>
        </div>
      </div>

      <!-- Messages Area -->
      <div 
        ref="messagesContainer"
        class="flex-1 overflow-y-auto p-3 md:p-4 space-y-4 md:space-y-6 z-10"
      >
        <div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full text-kakoclaw-text-secondary opacity-60">
          <div class="bg-kakoclaw-surface/50 p-6 rounded-full mb-4">
            <svg class="w-12 h-12 text-kakoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <p class="text-lg font-medium">Start a conversation</p>
          <p class="text-sm">Ask anything or run a task</p>
        </div>

        <!-- Messages -->
        <div v-for="msg in messages" :key="msg.id || msg.timestamp" class="animate-fadeIn group w-full">
          <MessageBubble
            :msg="msg"
            :currentSessionId="currentSessionId"
            :isLoading="isLoading"
            :isLastAssistantMessage="isLastAssistantMessage(msg)"
            @fork="forkAtMessage"
            @copy="copyMessageContent"
            @regenerate="regenerateResponse"
          />
        </div>

        <!-- Loading indicator (only when not streaming — streaming shows the message directly) -->
        <div v-if="isLoading && !chatStore.isStreaming" class="flex justify-start">
          <div class="bg-kakoclaw-surface/80 border border-kakoclaw-border rounded-2xl rounded-bl-sm px-2 sm:px-4 py-1.5 sm:py-3 shadow-sm">
            <div class="flex gap-1">
              <div class="w-1.5 sm:w-2 h-1.5 sm:h-2 bg-kakoclaw-text-secondary/50 rounded-full animate-bounce" style="animation-delay: 0s"></div>
              <div class="w-1.5 sm:w-2 h-1.5 sm:h-2 bg-kakoclaw-text-secondary/50 rounded-full animate-bounce" style="animation-delay: 0.2s"></div>
              <div class="w-1.5 sm:w-2 h-1.5 sm:h-2 bg-kakoclaw-text-secondary/50 rounded-full animate-bounce" style="animation-delay: 0.4s"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Input Area -->
      <div class="border-t border-kakoclaw-border/50 bg-kakoclaw-surface/80 backdrop-blur-md p-2.5 md:p-4 z-20 relative">
        <!-- Slash Command Autocomplete -->
        <div v-if="showAutocomplete" class="absolute bottom-full left-4 right-4 max-w-4xl mx-auto mb-1">
          <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-xl shadow-xl overflow-hidden max-h-64 overflow-y-auto">
            <button
              v-for="(cmd, idx) in filteredCommands"
              :key="cmd.command"
              @click="selectCommand(cmd)"
              class="w-full text-left px-4 py-2.5 text-sm transition-colors flex items-start gap-3 border-b border-kakoclaw-border/50 last:border-0"
              :class="idx === selectedCommandIndex ? 'bg-kakoclaw-accent/10 text-kakoclaw-accent' : 'hover:bg-kakoclaw-bg text-kakoclaw-text'"
            >
              <span class="font-mono text-xs bg-kakoclaw-bg px-1.5 py-0.5 rounded border border-kakoclaw-border flex-shrink-0 mt-0.5">{{ cmd.command }}</span>
              <div>
                <div class="font-medium text-xs">{{ cmd.label }}</div>
                <div class="text-[10px] text-kakoclaw-text-secondary mt-0.5">{{ cmd.description }}</div>
              </div>
            </button>
          </div>
        </div>

        <form @submit.prevent="sendMessage" class="flex flex-col gap-2 md:gap-3 max-w-4xl mx-auto w-full">
          <!-- File Attachment Preview Strip -->
          <div v-if="attachments.length > 0" class="flex flex-wrap gap-2 px-1">
            <div
              v-for="(att, idx) in attachments"
              :key="idx"
              class="flex items-center gap-2 bg-kakoclaw-bg border border-kakoclaw-border rounded-lg px-2.5 py-1.5 text-xs max-w-[200px]"
            >
              <svg class="w-3.5 h-3.5 text-kakoclaw-accent flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
              <span class="truncate text-kakoclaw-text flex-1">{{ att.name }}</span>
              <button type="button" @click="removeAttachment(idx)" class="text-kakoclaw-text-secondary hover:text-red-400 transition-colors flex-shrink-0">
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            <div v-if="uploadingFile" class="flex items-center gap-1.5 px-2.5 py-1.5 text-xs text-kakoclaw-text-secondary">
              <div class="w-3 h-3 border-2 border-kakoclaw-accent border-t-transparent rounded-full animate-spin"></div>
              Uploading...
            </div>
          </div>

          <div class="flex gap-2 md:gap-3 md:items-end">
            <!-- Hidden file input -->
            <input ref="fileInputRef" type="file" class="hidden" accept=".txt,.md,.json,.csv,.html,.xml,.yaml,.yml,.py,.go,.js,.ts,.java,.c,.cpp,.h,.cs,.rb,.rs,.php,.log,.pdf" @change="handleFileAttach">

            <div class="flex-1 relative min-w-0">
            <textarea
              ref="chatInput"
              v-model="messageInput"
              @input="onInputChange"
              @keydown="onInputKeydown"
              placeholder="Type a message or / for commands..."
              rows="1"
              class="w-full px-3 md:px-4 py-2 md:py-3 bg-kakoclaw-bg/50 border border-kakoclaw-border rounded-lg md:rounded-xl focus:ring-2 focus:ring-kakoclaw-accent/50 focus:border-kakoclaw-accent transition-all text-sm shadow-inner resize-none overflow-hidden"
              :disabled="!isConnected || isLoading"
              style="max-height: 120px;"
            ></textarea>
            </div>

            <!-- Action Buttons Row -->
            <div class="flex gap-2 md:gap-3 flex-shrink-0 flex-wrap sm:flex-nowrap">
              <!-- Attach File Button -->
              <button
                type="button"
                @click="fileInputRef?.click()"
                :disabled="!isConnected || isLoading || uploadingFile"
                class="flex-none px-2 md:px-3 py-2 md:py-3 rounded-lg md:rounded-xl bg-kakoclaw-surface border border-kakoclaw-border text-kakoclaw-text-secondary hover:text-kakoclaw-accent hover:bg-kakoclaw-bg transition-all flex items-center justify-center min-h-[2.5rem] md:min-h-auto"
                title="Attach file"
              >
                <svg class="w-4 md:w-5 h-4 md:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
                </svg>
              </button>

              <!-- Prompt Library Button -->
              <button
                type="button"
                @click="showPromptLibrary = true"
                :disabled="!isConnected || isLoading"
                class="flex-none px-2 md:px-3 py-2 md:py-3 rounded-lg md:rounded-xl bg-kakoclaw-surface border border-kakoclaw-border text-kakoclaw-text-secondary hover:text-kakoclaw-accent hover:bg-kakoclaw-bg transition-all flex items-center justify-center min-h-[2.5rem] md:min-h-auto"
                title="Prompt Library"
              >
                <svg class="w-4 md:w-5 h-4 md:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
              </button>

              <!-- Tools Manager Popover Toggle -->
            <div class="relative">
              <button
                type="button"
                @click="showToolsPopover = !showToolsPopover"
                :class="[
                  'flex-1 md:flex-none px-2 md:px-3 py-2 md:py-3 rounded-lg md:rounded-xl transition-all font-medium flex items-center justify-center min-h-[2.5rem] md:min-h-auto md:min-w-[3rem] border text-sm md:text-base',
                  chatStore.enabledTools.length < chatStore.availableTools.length
                    ? 'bg-amber-500/10 border-amber-500/40 text-amber-600 hover:bg-amber-500/20'
                    : 'bg-kakoclaw-surface border-kakoclaw-border text-kakoclaw-text-secondary hover:text-kakoclaw-accent hover:bg-kakoclaw-bg'
                ]"
                title="Manage AI Tools"
              >
                <svg class="w-4 md:w-5 h-4 md:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <span v-if="chatStore.enabledTools.length < chatStore.availableTools.length" class="absolute -top-1 -right-1 flex h-3 w-3">
                  <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75"></span>
                  <span class="relative inline-flex rounded-full h-3 w-3 bg-amber-500"></span>
                </span>
              </button>

              <!-- Tools Popover -->
              <Teleport to="body">
                <div v-if="showToolsPopover" class="fixed inset-0 z-[60]" @click="showToolsPopover = false"></div>
                <div 
                  v-if="showToolsPopover" 
                  class="fixed bottom-24 right-4 md:right-auto md:left-1/2 md:-translate-x-1/2 w-64 md:w-72 bg-kakoclaw-surface border border-kakoclaw-border rounded-2xl shadow-2xl z-[70] overflow-hidden animate-slideUp"
                >
                  <div class="p-3 border-b border-kakoclaw-border bg-kakoclaw-bg/50">
                    <h3 class="text-xs font-bold uppercase tracking-wider text-kakoclaw-text-secondary">AI Tools</h3>
                  </div>
                  <div class="max-h-64 overflow-y-auto p-2 space-y-1 custom-scrollbar">
                    <div v-if="chatStore.availableTools.length === 0" class="text-center py-4 text-xs text-kakoclaw-text-secondary">
                      Loading tools...
                    </div>
                    <button
                      v-for="tool in chatStore.availableTools"
                      :key="tool"
                      @click="chatStore.toggleTool(tool)"
                      class="w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs transition-colors"
                      :class="chatStore.enabledTools.includes(tool) ? 'bg-kakoclaw-accent/10 text-kakoclaw-accent' : 'hover:bg-kakoclaw-bg text-kakoclaw-text-secondary'"
                    >
                      <div class="flex items-center gap-2">
                        <div 
                          class="w-4 h-4 rounded border flex items-center justify-center transition-colors"
                          :class="chatStore.enabledTools.includes(tool) ? 'bg-kakoclaw-accent border-kakoclaw-accent text-white' : 'border-kakoclaw-border'"
                        >
                          <svg v-if="chatStore.enabledTools.includes(tool)" class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                        </div>
                        <span class="font-mono">{{ tool }}</span>
                      </div>
                    </button>
                  </div>
                </div>
              </Teleport>
            </div>
            <!-- Mic Button -->
            <button
              type="button"
              @mousedown="startRecording"
              @mouseup="stopRecording"
              @mouseleave="stopRecording"
              @touchstart.prevent="startRecording"
              @touchend.prevent="stopRecording"
              :disabled="!isConnected || isLoading || isTranscribing"
              :class="[
                'flex-1 md:flex-none px-2 md:px-3 py-2 md:py-3 rounded-lg md:rounded-xl transition-all font-medium flex items-center justify-center min-h-[2.5rem] md:min-h-auto md:min-w-[3rem] text-sm md:text-base',
                isRecording
                  ? 'bg-red-500 hover:bg-red-600 text-white shadow-lg shadow-red-500/30 animate-pulse'
                  : isTranscribing
                    ? 'bg-kakoclaw-surface text-kakoclaw-text-secondary cursor-wait'
                    : 'bg-kakoclaw-surface hover:bg-kakoclaw-bg border border-kakoclaw-border text-kakoclaw-text-secondary hover:text-kakoclaw-accent'
              ]"
              :title="isRecording ? 'Suelta para transcribir' : isTranscribing ? 'Transcribiendo...' : 'Mantén presionado para grabar'"
            >
              <svg v-if="isTranscribing" class="w-4 md:w-5 h-4 md:h-5 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
              </svg>
              <svg v-else class="w-4 md:w-5 h-4 md:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
              </svg>
            </button>
            <!-- Send/Stop Button (transforms based on loading state) -->
            <button
              v-if="!isLoading"
              type="submit"
              :disabled="!isConnected || !messageInput.trim()"
              class="flex-1 md:flex-none px-3 md:px-5 py-2 md:py-3 bg-kakoclaw-accent hover:bg-kakoclaw-accent-hover disabled:bg-kakoclaw-surface disabled:text-kakoclaw-text-secondary text-white rounded-lg md:rounded-xl transition-all shadow-lg shadow-kakoclaw-accent/20 hover:shadow-kakoclaw-accent/40 font-medium flex items-center justify-center min-h-[2.5rem] md:min-h-auto md:min-w-[3rem] text-sm md:text-base"
              title="Enviar mensaje"
            >
              <svg class="w-4 md:w-5 h-4 md:h-5 transform rotate-90" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
              </svg>
            </button>
            <!-- Stop Button (visible when loading) -->
            <button
              v-else
              type="button"
              @click="cancelExecution"
              class="flex-1 md:flex-none px-3 md:px-5 py-2 md:py-3 bg-red-500 hover:bg-red-600 text-white rounded-lg md:rounded-xl transition-all shadow-lg font-medium flex items-center justify-center min-h-[2.5rem] md:min-h-auto md:min-w-[3rem] text-sm md:text-base"
              title="Detener agente"
            >
              <svg class="w-4 md:w-5 h-4 md:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" />
              </svg>
            </button>
          </div>
          </div><!-- end flex row wrapper -->
        </form>

        <!-- Connection Status -->
        <div class="absolute top-0 right-0 -mt-8 mr-4 px-2 py-0.5 rounded text-[10px] font-mono glass">
           <span v-if="isConnected" class="text-kakoclaw-success flex items-center gap-1"><span class="w-1.5 h-1.5 rounded-full bg-kakoclaw-success animate-pulse"></span> Connected</span>
           <span v-else class="text-kakoclaw-error flex items-center gap-1"><span class="w-1.5 h-1.5 rounded-full bg-kakoclaw-error"></span> Disconnected</span>
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
  
  try {
    const data = await taskService.fetchSessionMessages(normalizedSessionId)
    chatStore.setMessages((data.messages || []).map(m => ({
      ...m,
      timestamp: m.created_at // Normalize timestamp
    })))
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
      timestamp: new Date().toISOString()
    })
    // Refresh sessions to show latest message/time
    fetchSessions()
  }
  if (message.type === 'stream_start') {
    chatStore.startStreamingMessage()
  }
  if (message.type === 'stream') {
    chatStore.appendStreamToken(message.content || '')
  }
  if (message.type === 'stream_end') {
    chatStore.endStreamingMessage(message.content || '')
    fetchSessions()
  }
  if (message.type === 'tool_call') {
    chatStore.addToolCall(message)
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
  window.__kakoclaw_setChatViewActive?.(true)

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
  window.__kakoclaw_setChatViewActive?.(false)
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

  // Add attachment content to message if present
  let finalContent = content
  if (attachments.value.length > 0) {
    const attachmentText = attachments.value.map(a =>
      `\n\n--- Attached file: ${a.name} ---\n${a.content}\n--- End of ${a.name} ---`
    ).join('')
    finalContent = content + attachmentText
    attachments.value = []
  }

  // Add user message locally
  chatStore.addMessage({
    role: 'user',
    content: finalContent,
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
    chatWs.send({
      type: 'message',
      content: finalContent,
      session_id: currentSessionId.value,
      model: chatStore.selectedModel || undefined,
      web_search: chatStore.webSearchEnabled
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

/* Custom scrollbar */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
::-webkit-scrollbar-track {
  background: transparent;
}
::-webkit-scrollbar-thumb {
  background: rgba(139, 92, 246, 0.2);
  border-radius: 3px;
}
::-webkit-scrollbar-thumb:hover {
  background: rgba(139, 92, 246, 0.4);
}
</style>
