<template>
  <div
    :class="[
      'flex w-full',
      msg.role === 'user' ? 'justify-end' : 'justify-start'
    ]"
  >
    <div
      :class="[
        'relative max-w-[92%] sm:max-w-[88%] md:max-w-[85%] lg:max-w-3xl xl:max-w-4xl 2xl:max-w-5xl px-3 sm:px-4 md:px-5 py-2 sm:py-2.5 md:py-3 shadow-lg transition-all duration-300 transform hover:scale-[1.002] animate-slideUp overflow-hidden group/bubble',
        msg.role === 'user'
          ? 'bg-gradient-to-br from-makoclaw-accent to-makoclaw-accent-hover text-white rounded-2xl rounded-br-none shadow-makoclaw-accent/10'
          : 'glass-panel text-makoclaw-text rounded-2xl rounded-bl-none shadow-lg shadow-black/[0.03] dark:shadow-black/20'
      ]"
    >
      <!-- Hover accent line for assistant messages -->
      <!-- Interactive Hover Effects (Assistant only) -->
      <template v-if="msg.role === 'assistant'">
        <!-- Left edge glow line -->
        <div class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-0 bg-gradient-to-b from-transparent via-makoclaw-accent/40 to-transparent group-hover/bubble:h-2/3 transition-all duration-300" />
        <!-- Bottom sweeper line -->
        <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-makoclaw-accent to-blue-500 group-hover/bubble:w-full transition-all duration-500 opacity-40" />
        <!-- Soft corner glow -->
        <div class="absolute -top-6 -right-6 w-16 h-16 bg-gradient-to-br from-makoclaw-accent to-blue-500 rounded-full opacity-0 blur-[15px] group-hover/bubble:opacity-10 transition-all duration-500" />
      </template>

      <p
        v-if="msg.role === 'user'"
        class="text-sm md:text-base whitespace-pre-wrap break-words leading-relaxed"
      >
        {{ msg.content }}
      </p>
      <template v-else>
        <!-- Tool Calls Rendering -->
        <div
          v-if="msg.toolCalls && msg.toolCalls.length > 0"
          class="mb-4 space-y-2"
        >
          <ToolCallItem
            v-for="tc in msg.toolCalls"
            :key="tc.id"
            :tc="tc"
          />
        </div>

        <!-- Contenido Segmentado (cuando hay múltiples agentes) -->
        <div
          v-if="msg.segments && msg.segments.length > 0"
          class="segmented-content"
        >
          <div
            v-for="segment in msg.segments"
            :key="segment.segmentId"
            class="content-segment mb-3 pb-3 border-b border-makoclaw-border/30 last:border-b-0"
          >
            <!-- Header de atribución -->
            <div class="flex items-center gap-2 mb-2">
              <SpecialistBadge
                :name="segment.agent"
                class="text-xs"
              />
              <span class="text-xs text-makoclaw-text-secondary">contributed:</span>
            </div>

            <!-- Contenido del segmento -->
            <MarkdownRenderer
              :content="segment.content"
              class="text-sm md:text-base pl-6"
            />
          </div>
        </div>

        <!-- Fallback: contenido unificado (cuando no hay segmentos) -->
        <div v-else>
          <!-- Streaming Content -->
          <p
            v-if="msg.streaming"
            class="text-sm md:text-base whitespace-pre-wrap break-words leading-relaxed"
          >
            {{ msg.content }}<span class="streaming-cursor" />
          </p>
          <!-- Final Markdown Content -->
          <MarkdownRenderer
            v-else
            :content="msg.content"
            class="text-sm md:text-base"
          />
        </div>
      </template>
      
      <!-- Agent Badges -->
      <div
        v-if="msg.agents && msg.agents.length > 0"
        class="agents-badge-row"
      >
        <SpecialistBadge 
          v-for="agent in msg.agents" 
          :key="agent"
          :name="agent"
          class="agent-badge"
        />
      </div>
      
      <div class="flex items-center justify-between mt-1 sm:mt-1.5">
        <p class="text-[10px] opacity-40 font-medium group-hover:opacity-70 transition-opacity">
          {{ formatTime(msg.timestamp || msg.created_at) }}
        </p>
        <div class="flex items-center gap-0.5 sm:gap-1">
          <!-- Fork button -->
          <button
            v-if="currentSessionId && msg.id"
            :disabled="isLoading"
            class="opacity-0 group-hover:opacity-100 transition-opacity p-1.5 sm:p-2 rounded-lg hover:bg-makoclaw-bg/80 text-makoclaw-text-secondary hover:text-makoclaw-accent disabled:opacity-30"
            title="Ramificar conversación (Continuar desde aquí)"
            @click="$emit('fork', msg)"
          >
            <svg
              class="w-3 sm:w-3.5 h-3 sm:h-3.5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
              />
            </svg>
          </button>
          <!-- Copy button -->
          <button
            v-if="msg.role === 'assistant' && !msg.streaming"
            class="opacity-0 group-hover:opacity-100 transition-opacity p-1.5 sm:p-2 rounded-lg hover:bg-makoclaw-bg/80 text-makoclaw-text-secondary hover:text-makoclaw-accent"
            title="Copiar respuesta"
            @click="$emit('copy', msg.content)"
          >
            <svg
              class="w-3 sm:w-3.5 h-3 sm:h-3.5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
              />
            </svg>
          </button>
          <!-- Regenerate button -->
          <button
            v-if="msg.role === 'assistant' && isLastAssistantMessage"
            :disabled="isLoading"
            class="opacity-0 group-hover:opacity-100 transition-opacity p-1.5 sm:p-2 rounded-lg hover:bg-makoclaw-bg/80 text-makoclaw-text-secondary hover:text-makoclaw-accent disabled:opacity-30"
            title="Regenerar respuesta"
            @click="$emit('regenerate')"
          >
            <svg
              class="w-3 sm:w-3.5 h-3 sm:h-3.5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
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
import SpecialistBadge from './Chat/SpecialistBadge.vue'

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

<style scoped>
.agents-badge-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid rgb(var(--pc-border) / 0.2);
}

.agent-badge {
  font-size: 0.75rem;
}
</style>

