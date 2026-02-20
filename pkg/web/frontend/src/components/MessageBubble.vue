<template>
  <div
    :class="[
      'flex w-full',
      msg.role === 'user' ? 'justify-end' : 'justify-start'
    ]"
  >
    <div
      :class="[
        'max-w-[90%] sm:max-w-[85%] lg:max-w-2xl px-4 md:px-5 py-2.5 md:py-3 shadow-lg transition-all duration-300 transform hover:scale-[1.002] animate-slideUp',
        msg.role === 'user'
          ? 'bg-gradient-to-br from-kakoclaw-accent to-emerald-600 text-white rounded-2xl rounded-br-none shadow-kakoclaw-accent/10'
          : 'glass-panel text-kakoclaw-text rounded-2xl rounded-bl-none shadow-black/5'
      ]"
    >
      <p v-if="msg.role === 'user'" class="text-sm md:text-base whitespace-pre-wrap break-words leading-relaxed">{{ msg.content }}</p>
      <template v-else>
        <!-- Tool Calls Rendering -->
        <div v-if="msg.toolCalls && msg.toolCalls.length > 0" class="mb-4 space-y-2">
          <ToolCallItem v-for="tc in msg.toolCalls" :key="tc.id" :tc="tc" />
        </div>

        <!-- Streaming Content -->
        <p v-if="msg.streaming" class="text-sm md:text-base whitespace-pre-wrap break-words leading-relaxed">{{ msg.content }}<span class="streaming-cursor"></span></p>
        <!-- Final Markdown Content -->
        <MarkdownRenderer v-else :content="msg.content" class="text-sm md:text-base" />
      </template>
      
      <div class="flex items-center justify-between mt-1 sm:mt-1.5">
        <p class="text-[10px] opacity-40 font-medium group-hover:opacity-70 transition-opacity">
          {{ formatTime(msg.timestamp || msg.created_at) }}
        </p>
        <div class="flex items-center gap-0.5 sm:gap-1">
          <!-- Fork button -->
          <button
            v-if="currentSessionId && msg.id"
            @click="$emit('fork', msg)"
            :disabled="isLoading"
            class="opacity-0 group-hover:opacity-100 transition-opacity p-0.5 sm:p-1 rounded-md hover:bg-kakoclaw-bg/80 text-kakoclaw-text-secondary hover:text-kakoclaw-accent disabled:opacity-30"
            title="Ramificar conversación (Continuar desde aquí)"
          >
            <svg class="w-3 sm:w-3.5 h-3 sm:h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
            </svg>
          </button>
          <!-- Copy button -->
          <button
            v-if="msg.role === 'assistant' && !msg.streaming"
            @click="$emit('copy', msg.content)"
            class="opacity-0 group-hover:opacity-100 transition-opacity p-0.5 sm:p-1 rounded-md hover:bg-kakoclaw-bg/80 text-kakoclaw-text-secondary hover:text-kakoclaw-accent"
            title="Copiar respuesta"
          >
            <svg class="w-3 sm:w-3.5 h-3 sm:h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
          </button>
          <!-- Regenerate button -->
          <button
            v-if="msg.role === 'assistant' && isLastAssistantMessage"
            @click="$emit('regenerate')"
            :disabled="isLoading"
            class="opacity-0 group-hover:opacity-100 transition-opacity p-0.5 sm:p-1 rounded-md hover:bg-kakoclaw-bg/80 text-kakoclaw-text-secondary hover:text-kakoclaw-accent disabled:opacity-30"
            title="Regenerar respuesta"
          >
            <svg class="w-3 sm:w-3.5 h-3 sm:h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import MarkdownRenderer from './Chat/MarkdownRenderer.vue'
import ToolCallItem from './ToolCallItem.vue'

defineProps({
  msg: {
    type: Object,
    required: true
  },
  currentSessionId: {
    type: String,
    default: null
  },
  isLoading: {
    type: Boolean,
    default: false
  },
  isLastAssistantMessage: {
    type: Boolean,
    default: false
  }
})

defineEmits(['fork', 'copy', 'regenerate'])

const formatTime = (isoString) => {
  if (!isoString) return ''
  const date = new Date(isoString)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
</script>
