<template>
  <div class="flex h-full bg-picoclaw-bg relative overflow-hidden">
    <!-- Background Gradient Mesh (Subtle) -->
    <div class="absolute inset-0 pointer-events-none opacity-20 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-picoclaw-accent/30 via-transparent to-transparent"></div>

    <!-- Sidebar (History) -->
    <div 
      :class="[
        'w-64 flex-shrink-0 border-r border-picoclaw-border bg-picoclaw-surface/50 backdrop-blur-sm transition-all duration-300 flex flex-col',
        showSidebar ? 'translate-x-0' : '-translate-x-full absolute h-full z-20 md:relative md:translate-x-0'
      ]"
    >
      <div class="p-4 border-b border-picoclaw-border flex justify-between items-center">
        <h2 class="font-semibold text-sm text-picoclaw-text-secondary">History</h2>
        <button @click="startNewChat" class="p-1.5 hover:bg-picoclaw-bg rounded-md text-picoclaw-accent transition-colors" title="New Chat">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
        </button>
      </div>
      
      <div class="flex-1 overflow-y-auto p-2 space-y-1">
        <div v-if="sessions.length === 0" class="text-xs text-picoclaw-text-secondary text-center py-4">
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
              class="flex-1 bg-picoclaw-bg border border-picoclaw-accent rounded px-2 py-1 text-xs text-picoclaw-text focus:outline-none"
              autofocus
              placeholder="Session title..."
            />
          </div>
          <!-- Normal session button -->
          <button
            v-else
            @click="loadSession(session.session_id)"
            :class="[
              'w-full text-left px-3 py-2 rounded-lg text-sm transition-colors',
              currentSessionId === session.session_id ? 'bg-picoclaw-accent/10 text-picoclaw-accent' : 'hover:bg-picoclaw-bg text-picoclaw-text-secondary hover:text-picoclaw-text'
            ]"
          >
            <div class="flex items-center gap-2">
              <svg v-if="session.session_id.startsWith('web:task:')" class="w-3.5 h-3.5 text-amber-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" /></svg>
              <svg v-else class="w-3.5 h-3.5 text-emerald-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>
              <span class="truncate flex-1">{{ session.title || session.last_message || 'Empty session' }}</span>
              <!-- Context menu trigger -->
              <button
                @click.stop="openContextMenu($event, session.session_id)"
                class="opacity-0 group-hover:opacity-100 p-0.5 hover:bg-picoclaw-border rounded transition-opacity"
                title="Session actions"
              >
                <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20"><circle cx="10" cy="4" r="1.5"/><circle cx="10" cy="10" r="1.5"/><circle cx="10" cy="16" r="1.5"/></svg>
              </button>
            </div>
            <div class="text-[10px] opacity-60 mt-0.5 pl-5 flex justify-between">
              <span>{{ formatTime(session.updated_at) }}</span>
              <span v-if="session.message_count" class="text-picoclaw-text-secondary">{{ session.message_count }} msg{{ session.message_count !== 1 ? 's' : '' }}</span>
            </div>
          </button>
        </div>
      </div>

      <!-- Context Menu Overlay -->
      <Teleport to="body">
        <div v-if="contextMenu.show" class="fixed inset-0 z-50" @click="closeContextMenu">
          <div
            class="absolute bg-picoclaw-surface border border-picoclaw-border rounded-lg shadow-xl py-1 min-w-[140px]"
            :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
            @click.stop
          >
            <button @click="startRename(contextMenu.sessionId)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-picoclaw-bg transition-colors flex items-center gap-2">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" /></svg>
              Rename
            </button>
            <button @click="archiveSession(contextMenu.sessionId)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-picoclaw-bg transition-colors flex items-center gap-2">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" /></svg>
              Archive
            </button>
            <div class="border-t border-picoclaw-border my-1"></div>
            <button @click="deleteSession(contextMenu.sessionId)" class="w-full text-left px-3 py-1.5 text-sm hover:bg-picoclaw-error/10 text-picoclaw-error transition-colors flex items-center gap-2">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
              Delete
            </button>
          </div>
        </div>
      </Teleport>
    </div>

    <!-- Main Chat Area -->
    <div class="flex-1 flex flex-col min-w-0 relative bg-picoclaw-bg/50">
      <!-- Top Bar: Mobile toggle + Model selector -->
      <div class="flex items-center justify-between px-4 py-2 border-b border-picoclaw-border/30 bg-picoclaw-surface/30 backdrop-blur-sm z-20">
        <!-- Mobile Sidebar Toggle -->
        <button 
          @click="showSidebar = !showSidebar"
          class="md:hidden p-2 bg-picoclaw-surface border border-picoclaw-border rounded-lg shadow-sm"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" /></svg>
        </button>
        <div class="hidden md:block"></div>

        <!-- Model Selector -->
        <div class="flex items-center gap-2">
          <svg class="w-4 h-4 text-picoclaw-text-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
          <select
            v-model="chatStore.selectedModel"
            :disabled="chatStore.allModels.length === 0"
            class="bg-picoclaw-bg/50 border border-picoclaw-border rounded-lg px-3 py-1.5 text-xs text-picoclaw-text focus:ring-2 focus:ring-picoclaw-accent/50 focus:border-picoclaw-accent transition-all cursor-pointer max-w-[280px]"
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
        class="flex-1 overflow-y-auto p-4 space-y-6 z-10"
      >
        <div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full text-picoclaw-text-secondary opacity-60">
          <div class="bg-picoclaw-surface/50 p-6 rounded-full mb-4">
            <svg class="w-12 h-12 text-picoclaw-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <p class="text-lg font-medium">Start a conversation</p>
          <p class="text-sm">Ask anything or run a task</p>
        </div>

        <!-- Messages -->
        <div v-for="msg in messages" :key="msg.id || msg.timestamp" class="animate-fadeIn group">
          <div
            :class="[
              'flex w-full',
              msg.role === 'user' ? 'justify-end' : 'justify-start'
            ]"
          >
            <div
              :class="[
                'max-w-[85%] lg:max-w-2xl px-5 py-3 shadow-md',
                msg.role === 'user'
                  ? 'bg-gradient-to-br from-picoclaw-accent to-picoclaw-accent-hover text-white rounded-2xl rounded-br-sm'
                  : 'bg-picoclaw-surface/90 border border-picoclaw-border text-picoclaw-text rounded-2xl rounded-bl-sm'
              ]"
            >
              <p v-if="msg.role === 'user'" class="text-sm md:text-base whitespace-pre-wrap break-words leading-relaxed">{{ msg.content }}</p>
              <template v-else>
                <!-- Streaming: show raw text with cursor while streaming, markdown when done -->
                <p v-if="msg.streaming" class="text-sm md:text-base whitespace-pre-wrap break-words leading-relaxed">{{ msg.content }}<span class="streaming-cursor"></span></p>
                <MarkdownRenderer v-else :content="msg.content" class="text-sm md:text-base" />
              </template>
              <div class="flex items-center justify-between mt-1.5">
                <p class="text-[10px] opacity-0 group-hover:opacity-70 transition-opacity">
                  {{ formatTime(msg.timestamp || msg.created_at) }}
                </p>
                <div class="flex items-center gap-1">
                  <!-- Fork button (on any message) -->
                  <button
                    v-if="currentSessionId && msg.id"
                    @click="forkAtMessage(msg)"
                    :disabled="isLoading"
                    class="opacity-0 group-hover:opacity-100 transition-opacity p-1 rounded-md hover:bg-picoclaw-bg/80 text-picoclaw-text-secondary hover:text-picoclaw-accent disabled:opacity-30"
                    title="Fork conversation from this point"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                    </svg>
                  </button>
                  <!-- Copy button (on assistant messages) -->
                  <button
                    v-if="msg.role === 'assistant' && !msg.streaming"
                    @click="copyMessageContent(msg.content)"
                    class="opacity-0 group-hover:opacity-100 transition-opacity p-1 rounded-md hover:bg-picoclaw-bg/80 text-picoclaw-text-secondary hover:text-picoclaw-accent"
                    title="Copy response"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                    </svg>
                  </button>
                  <!-- Regenerate button (only on last assistant message) -->
                  <button
                    v-if="msg.role === 'assistant' && isLastAssistantMessage(msg)"
                    @click="regenerateResponse"
                    :disabled="isLoading"
                    class="opacity-0 group-hover:opacity-100 transition-opacity p-1 rounded-md hover:bg-picoclaw-bg/80 text-picoclaw-text-secondary hover:text-picoclaw-accent disabled:opacity-30"
                    title="Regenerate response"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Loading indicator (only when not streaming — streaming shows the message directly) -->
        <div v-if="isLoading && !chatStore.isStreaming" class="flex justify-start">
          <div class="bg-picoclaw-surface/80 border border-picoclaw-border rounded-2xl rounded-bl-sm px-4 py-3 shadow-sm">
            <div class="flex gap-1.5">
              <div class="w-2 h-2 bg-picoclaw-text-secondary/50 rounded-full animate-bounce" style="animation-delay: 0s"></div>
              <div class="w-2 h-2 bg-picoclaw-text-secondary/50 rounded-full animate-bounce" style="animation-delay: 0.2s"></div>
              <div class="w-2 h-2 bg-picoclaw-text-secondary/50 rounded-full animate-bounce" style="animation-delay: 0.4s"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Input Area -->
      <div class="border-t border-picoclaw-border/50 bg-picoclaw-surface/80 backdrop-blur-md p-4 z-20 relative">
        <!-- Slash Command Autocomplete -->
        <div v-if="showAutocomplete" class="absolute bottom-full left-4 right-4 max-w-4xl mx-auto mb-1">
          <div class="bg-picoclaw-surface border border-picoclaw-border rounded-xl shadow-xl overflow-hidden max-h-64 overflow-y-auto">
            <button
              v-for="(cmd, idx) in filteredCommands"
              :key="cmd.command"
              @click="selectCommand(cmd)"
              class="w-full text-left px-4 py-2.5 text-sm transition-colors flex items-start gap-3 border-b border-picoclaw-border/50 last:border-0"
              :class="idx === selectedCommandIndex ? 'bg-picoclaw-accent/10 text-picoclaw-accent' : 'hover:bg-picoclaw-bg text-picoclaw-text'"
            >
              <span class="font-mono text-xs bg-picoclaw-bg px-1.5 py-0.5 rounded border border-picoclaw-border flex-shrink-0 mt-0.5">{{ cmd.command }}</span>
              <div>
                <div class="font-medium text-xs">{{ cmd.label }}</div>
                <div class="text-[10px] text-picoclaw-text-secondary mt-0.5">{{ cmd.description }}</div>
              </div>
            </button>
          </div>
        </div>

        <form @submit.prevent="sendMessage" class="flex gap-3 max-w-4xl mx-auto items-end">
          <div class="flex-1 relative">
            <textarea
              ref="chatInput"
              v-model="messageInput"
              @input="onInputChange"
              @keydown="onInputKeydown"
              placeholder="Type a message or / for commands..."
              rows="1"
              class="w-full px-4 py-3 bg-picoclaw-bg/50 border border-picoclaw-border rounded-xl focus:ring-2 focus:ring-picoclaw-accent/50 focus:border-picoclaw-accent transition-all text-sm shadow-inner resize-none overflow-hidden"
              :disabled="!isConnected || isLoading"
              style="max-height: 120px;"
            ></textarea>
          </div>
          <!-- Web Search Toggle -->
          <button
            type="button"
            @click="chatStore.setWebSearchEnabled(!chatStore.webSearchEnabled)"
            :class="[
              'px-3 py-3 rounded-xl transition-all font-medium flex items-center justify-center min-w-[3rem] border',
              chatStore.webSearchEnabled
                ? 'bg-picoclaw-accent/15 border-picoclaw-accent/40 text-picoclaw-accent hover:bg-picoclaw-accent/25'
                : 'bg-picoclaw-surface border-picoclaw-border text-picoclaw-text-secondary hover:text-picoclaw-text hover:bg-picoclaw-bg'
            ]"
            :title="chatStore.webSearchEnabled ? 'Web search enabled (click to disable)' : 'Web search disabled (click to enable)'"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </button>
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
              'px-3 py-3 rounded-xl transition-all font-medium flex items-center justify-center min-w-[3rem]',
              isRecording
                ? 'bg-red-500 hover:bg-red-600 text-white shadow-lg shadow-red-500/30 animate-pulse'
                : isTranscribing
                  ? 'bg-picoclaw-surface text-picoclaw-text-secondary cursor-wait'
                  : 'bg-picoclaw-surface hover:bg-picoclaw-bg border border-picoclaw-border text-picoclaw-text-secondary hover:text-picoclaw-accent'
            ]"
            :title="isRecording ? 'Release to transcribe' : isTranscribing ? 'Transcribing...' : 'Hold to record'"
          >
            <svg v-if="isTranscribing" class="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
            </svg>
            <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
            </svg>
          </button>
          <button
            type="submit"
            :disabled="!isConnected || isLoading || !messageInput.trim()"
            class="px-5 py-3 bg-picoclaw-accent hover:bg-picoclaw-accent-hover disabled:bg-picoclaw-surface disabled:text-picoclaw-text-secondary text-white rounded-xl transition-all shadow-lg shadow-picoclaw-accent/20 hover:shadow-picoclaw-accent/40 font-medium flex items-center justify-center min-w-[3rem]"
          >
            <svg class="w-5 h-5 transform rotate-90" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
            </svg>
          </button>
        </form>

        <!-- Connection Status -->
        <div class="absolute top-0 right-0 -mt-8 mr-4 px-2 py-0.5 rounded text-[10px] font-mono glass">
           <span v-if="isConnected" class="text-picoclaw-success flex items-center gap-1"><span class="w-1.5 h-1.5 rounded-full bg-picoclaw-success animate-pulse"></span> Connected</span>
           <span v-else class="text-picoclaw-error flex items-center gap-1"><span class="w-1.5 h-1.5 rounded-full bg-picoclaw-error"></span> Disconnected</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { storeToRefs } from 'pinia'
import MarkdownRenderer from '../components/Chat/MarkdownRenderer.vue'
import { useChatStore } from '../stores/chatStore'
import { ChatWebSocket } from '../services/websocketService'
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
const showSidebar = ref(false)
const sessions = ref([])
const currentSessionId = ref(null)
const contextMenu = ref({ show: false, sessionId: null, x: 0, y: 0 })
const renamingSession = ref(null)
const renameInput = ref('')

const { messages } = storeToRefs(chatStore)
const chatWs = new ChatWebSocket()
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
    toast.success('Session renamed')
  } catch (error) {
    console.error('Failed to rename session:', error)
    toast.error('Failed to rename session')
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
    toast.success('Session archived')
  } catch (error) {
    console.error('Failed to archive session:', error)
    toast.error('Failed to archive session')
  }
}

const deleteSession = async (sessionId) => {
  closeContextMenu()
  if (!confirm('Delete this session and all its messages? This cannot be undone.')) return
  try {
    await taskService.deleteSession(sessionId)
    sessions.value = sessions.value.filter(s => s.session_id !== sessionId)
    if (currentSessionId.value === sessionId) {
      startNewChat()
    }
    toast.success('Session deleted')
  } catch (error) {
    console.error('Failed to delete session:', error)
    toast.error('Failed to delete session')
  }
}

const generateSessionId = () => {
  return 'web:chat:' + Date.now().toString(36) + Math.random().toString(36).substr(2)
}

const copyMessageContent = async (content) => {
  try {
    await navigator.clipboard.writeText(content)
    toast.success('Copied to clipboard')
  } catch {
    toast.error('Failed to copy')
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
  if (message.type === 'ready') {
    isLoading.value = false
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
  await fetchSessions()

  // Fetch available models
  try {
    const modelsData = await advancedService.fetchModels()
    chatStore.setModelsData(modelsData)
  } catch (error) {
    console.error('Failed to fetch models:', error)
    chatStore.setModelsData({ current_model: '', providers: [] })
  }
  
  // Check for session ID in route
  const routeSessionId = normalizeSessionId(route.query.id)
  if (routeSessionId) {
    const routeSessionExists = sessions.value.some(s => s.session_id === routeSessionId)
    if (routeSessionExists) {
      await loadSession(routeSessionId, { updateRoute: false })
    } else if (sessions.value.length > 0) {
      await loadSession(sessions.value[0].session_id)
    }
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
})

onBeforeUnmount(() => {
  chatWs.off('message', handleMessage)
  chatWs.off('disconnected', handleDisconnected)
  chatWs.off('connected', handleConnected)
  chatWs.disconnect()
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

  // Add user message locally
  chatStore.addMessage({
    role: 'user',
    content,
    timestamp: new Date().toISOString()
  })

  messageInput.value = ''
  isLoading.value = true

  // Reset textarea height
  if (chatInput.value) {
    chatInput.value.style.height = 'auto'
  }

  // Send via WebSocket
  if (chatWs.isConnected()) {
    chatWs.send({
      type: 'message',
      content,
      session_id: currentSessionId.value,
      model: chatStore.selectedModel || undefined,
      web_search: chatStore.webSearchEnabled
    })
    // Refresh sessions to show new thread
    setTimeout(fetchSessions, 500)
  } else {
    isLoading.value = false
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

  if (!confirm('Fork conversation from this message? A new session will be created.')) return

  try {
    const result = await advancedService.forkChat(currentSessionId.value, msg.id)
    // Navigate to the forked session
    currentSessionId.value = result.new_session_id
    await loadSession(result.new_session_id)
    await fetchSessions()
    toast.success('Conversation forked')
  } catch (error) {
    console.error('Fork failed:', error)
    toast.error('Fork failed: ' + (error.response?.data || error.message || 'Unknown error'))
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
