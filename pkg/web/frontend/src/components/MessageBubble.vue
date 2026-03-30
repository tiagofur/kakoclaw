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
        <!-- Agent Activity Blocks (specialist agents working) -->
        <div
          v-if="msg.agentActivities && msg.agentActivities.length > 0"
          class="mb-4 space-y-2"
        >
          <AgentActivityItem
            v-for="act in msg.agentActivities"
            :key="act.id"
            :activity="act"
            :msg="msg"
          />
        </div>

        <!-- Tool Calls Rendering -->
        <div
          v-if="msg.toolCalls && msg.toolCalls.length > 0"
          class="mb-4 space-y-2"
        >
          <ToolCallItem
            v-for="tc in msg.toolCalls"
            :key="tc.id"
            :tool-call="tc"
          />
        </div>

        <div
          v-if="msg.thinkingBlocks && msg.thinkingBlocks.length > 0"
          class="mb-4 space-y-2"
        >
          <ThinkingBlock
            v-for="block in msg.thinkingBlocks"
            :key="block.id"
            :thinking-block="block"
            :is-streaming="msg.streaming"
          />
        </div>

        <!-- Specialist segments render before main content -->
        <div
          v-if="msg.segments && msg.segments.length > 0"
          class="mb-4 space-y-3"
        >
          <SpecialistSegment
            v-for="(segment, index) in msg.segments"
            :key="segment.segmentId || `${segment.agent || segment.specialist || 'segment'}-${index}`"
            :segment="segment"
            :is-streaming="chatStore.isStreaming"
          />
        </div>

        <div v-if="msg.content || inlineImages.length > 0">
          <!-- Streaming Content -->
          <p
            v-if="msg.content && msg.streaming"
            class="text-sm md:text-base whitespace-pre-wrap break-words leading-relaxed"
          >
            {{ msg.content }}<span class="streaming-cursor" />
          </p>
          <!-- Final Markdown Content -->
          <MarkdownRenderer
            v-else-if="msg.content"
            :content="msg.content"
            class="text-sm md:text-base"
          />

          <div
            v-if="inlineImages.length > 0"
            class="mt-4 flex flex-wrap gap-3"
          >
            <ImagePreviewCard
              v-for="image in inlineImages"
              :key="image.key"
              :src="image.src"
              :caption="image.caption"
              :download-url="image.downloadUrl"
            />
          </div>
        </div>

        <!-- Fallback: no content and no segments (shouldn't happen but safety) -->
        <div v-else-if="!msg.segments || msg.segments.length === 0">
          <p class="text-sm text-makoclaw-text-secondary italic">No content</p>
        </div>

        <!-- Legacy: Collapsible agent workflow details (from agentActivity snapshot) -->
        <!-- Now handled by AgentActivityItem blocks above -->
      </template>
      
      <div class="flex items-center justify-between mt-1 sm:mt-1.5">
        <p class="text-[10px] opacity-40 font-medium group-hover:opacity-70 transition-opacity">
          {{ formatTime(msg.timestamp || msg.created_at) }}
        </p>
        <div class="flex items-center gap-1 sm:gap-1.5">
          <!-- Agent badges (right-aligned, before action buttons) -->
          <SpecialistBadge
            v-for="agent in (msg.agents || [])"
            :key="agent"
            :name="agent"
            class="text-[10px]"
          />
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
          <!-- Speak button (TTS) -->
          <button
            v-if="msg.role === 'assistant' && !msg.streaming && msg.content"
            :disabled="isSynthesizing"
            class="opacity-0 group-hover:opacity-100 transition-opacity p-1.5 sm:p-2 rounded-lg hover:bg-makoclaw-bg/80 text-makoclaw-text-secondary hover:text-makoclaw-accent disabled:opacity-30"
            :title="isSynthesizing ? 'Synthesizing...' : 'Read aloud'"
            @click="synthesizeAudio(msg.content)"
          >
            <svg
              v-if="!isSynthesizing"
              class="w-3 sm:w-3.5 h-3 sm:h-3.5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M15.536 8.464a5 5 0 010 7.072M17.95 6.05a8 8 0 010 11.9M6.5 8.8l4.29-3.575A1 1 0 0112.5 6v12a1 1 0 01-1.71.705L6.5 15.2H4a1 1 0 01-1-1v-4.4a1 1 0 011-1h2.5z"
              />
            </svg>
            <svg
              v-else
              class="w-3 sm:w-3.5 h-3 sm:h-3.5 animate-pulse"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m-4-4h8"
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
import { computed, ref } from 'vue'
import MarkdownRenderer from './Chat/MarkdownRenderer.vue'
import ThinkingBlock from './Chat/ThinkingBlock.vue'
import ToolCallItem from './ToolCallItem.vue'
import SpecialistBadge from './Chat/SpecialistBadge.vue'
import AgentActivityItem from './Chat/AgentActivityItem.vue'
import SpecialistSegment from './Chat/SpecialistSegment.vue'
import ImagePreviewCard from './ImagePreviewCard.vue'
import { useAuthStore } from '../stores/authStore'
import { useChatStore } from '../stores/chatStore'
import { buildFilesApiUrl, deriveWorkspaceRelativePath, extractImageResultData } from '../utils/imagePreview'

const isSynthesizing = ref(false)
const authStore = useAuthStore()
const chatStore = useChatStore()

const props = defineProps({
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

const workspaceRoot = computed(() => authStore.user?.workspace || '')

const inlineImages = computed(() => {
  const toolCalls = props.msg?.toolCalls || props.msg?.tool_calls || []

  return toolCalls
    .filter(toolCall => toolCall?.name === 'image_generate' || toolCall?.function?.name === 'image_generate')
    .map((toolCall, index) => {
      const parsed = extractImageResultData(toolCall.result || toolCall.content)
      if (!parsed) return null

      const candidatePath = parsed.local_path || parsed.url || ''
      const relPath = parsed.local_path
        ? deriveWorkspaceRelativePath(parsed.local_path, workspaceRoot.value)
        : ''

      return {
        key: toolCall.id || `${candidatePath}-${index}`,
        src: parsed.local_path
          ? buildFilesApiUrl(relPath, authStore.token, { workspaceRoot: workspaceRoot.value })
          : parsed.url,
        downloadUrl: parsed.local_path
          ? buildFilesApiUrl(relPath, authStore.token, { download: true, workspaceRoot: workspaceRoot.value })
          : parsed.url,
        caption: parsed.revised_prompt || ''
      }
    })
    .filter(Boolean)
})

async function synthesizeAudio(text) {
  if (isSynthesizing.value || !text) return
  isSynthesizing.value = true
  try {
    const resp = await fetch('/api/v1/voice/synthesize', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authStore.token}`
      },
      body: JSON.stringify({ text: text.slice(0, 4096) })
    })
    if (!resp.ok) {
      const err = await resp.json().catch(() => ({}))
      console.error('TTS error:', err.error || resp.statusText)
      return
    }
    const blob = await resp.blob()
    const url = URL.createObjectURL(blob)
    const audio = new Audio(url)
    audio.onended = () => URL.revokeObjectURL(url)
    audio.play()
  } catch (e) {
    console.error('TTS synthesis failed:', e)
  } finally {
    isSynthesizing.value = false
  }
}

const formatTime = (isoString) => {
  if (!isoString) return ''
  const date = new Date(isoString)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
/* Agent badges are now inline in the footer row */
</style>
