<template>
  <div class="bg-makoclaw-bg/40 border border-makoclaw-border rounded-xl overflow-hidden transition-all duration-300">
    <button
      class="w-full flex items-center justify-between px-3 md:px-4 py-2 md:py-2.5 text-[10px] md:text-xs font-mono hover:bg-makoclaw-accent/10 transition-colors"
      @click="toggleExpanded"
    >
      <div class="flex items-center gap-2 md:gap-3 min-w-0">
        <div
          :class="[
            'w-2 h-2 rounded-full flex-shrink-0',
            toolCall.status === 'started' ? 'bg-makoclaw-warning animate-pulse' :
            toolCall.status === 'error' ? 'bg-makoclaw-error' : 'bg-makoclaw-success'
          ]"
        />
        <div class="flex items-center gap-1.5 min-w-0">
          <span class="text-makoclaw-text-secondary">Tool:</span>
          <span class="text-makoclaw-accent font-bold truncate">{{ toolCall.name }}</span>
        </div>
        <span
          class="text-[9px] px-1.5 py-0.5 rounded-full hidden sm:inline flex-shrink-0"
          :class="statusBadgeClass"
        >{{ statusLabel }}</span>
      </div>
      <svg
        class="w-3.5 md:w-4 h-3.5 md:h-4 text-makoclaw-text-secondary transition-transform duration-300"
        :class="{ 'rotate-180': isExpanded }"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M19 9l-7 7-7-7"
        />
      </svg>
    </button>

    <Transition name="expand">
      <div
        v-if="isExpanded"
        class="px-3 md:px-4 pb-3 md:pb-4 border-t border-makoclaw-border/30 animate-fadeIn bg-makoclaw-bg/20"
      >
        <div class="mt-3 space-y-3">
          <div>
            <div class="text-[9px] md:text-[10px] text-makoclaw-text-secondary uppercase tracking-wider font-bold mb-1.5 flex items-center gap-1.5">
              <svg
                class="w-3 h-3"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
              /></svg>
              Arguments
            </div>
            <pre class="bg-black/30 p-2 md:p-3 rounded-lg text-makoclaw-text/90 overflow-x-auto custom-scrollbar text-[9px] md:text-[11px] leading-tight font-mono">{{ formattedArgs }}</pre>
          </div>

          <div v-if="formattedResult">
            <div class="text-[9px] md:text-[10px] text-makoclaw-text-secondary uppercase tracking-wider font-bold mb-1.5 flex items-center gap-1.5">
              <svg
                class="w-3 h-3"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              ><path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              /></svg>
              Result
            </div>

            <!-- Image preview for image_generate tool -->
            <template v-if="imageResult && !imgError">
              <div class="rounded-2xl overflow-hidden border border-purple-500/20 bg-purple-500/5">
                <img
                  :src="imageDisplayUrl"
                  alt="Generated image"
                  class="w-full max-w-lg object-contain"
                  @error="imgError = true"
                >
                <div class="flex items-center justify-between px-4 py-2.5 border-t border-purple-500/15">
                  <p
                    v-if="imageResult.revised_prompt"
                    class="text-[10px] text-purple-300/60 italic truncate flex-1 mr-3"
                  >
                    {{ imageResult.revised_prompt }}
                  </p>
                  <button
                    class="flex items-center gap-1.5 px-3 py-1.5 bg-purple-600/20 hover:bg-purple-600/40 border border-purple-500/30 rounded-lg text-[10px] font-bold text-purple-300 uppercase tracking-widest transition-all active:scale-95 flex-none"
                    @click.stop="downloadImage"
                  >
                    <ArrowDownTrayIcon class="w-3.5 h-3.5" />
                    Download
                  </button>
                </div>
              </div>
            </template>

            <!-- Default: raw text/JSON result -->
            <div
              v-else
              class="bg-black/20 p-2 md:p-3 rounded-lg text-makoclaw-text/80 whitespace-pre-wrap max-h-48 md:max-h-64 overflow-y-auto custom-scrollbar font-mono text-[9px] md:text-[11px] leading-normal border border-makoclaw-border/10"
            >
              {{ formattedResult }}
            </div>
          </div>

          <div
            v-if="formattedTime"
            class="text-[9px] text-makoclaw-text-secondary/60 font-medium"
          >
            {{ formattedTime }}
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { ArrowDownTrayIcon } from '@heroicons/vue/24/outline'
import { useAuthStore } from '../stores/authStore'

const props = defineProps({
  toolCall: {
    type: Object,
    required: true
  }
})

const authStore = useAuthStore()
const localExpanded = ref(false)
const imgError = ref(false)

const isExpanded = computed(() => localExpanded.value)

const statusLabel = computed(() => {
  if (props.toolCall.status === 'started') return 'executing…'
  if (props.toolCall.status === 'error') return 'error'
  return 'done'
})

const statusBadgeClass = computed(() => {
  if (props.toolCall.status === 'started') return 'bg-makoclaw-warning/15 text-makoclaw-warning ring-1 ring-makoclaw-warning/30'
  if (props.toolCall.status === 'error') return 'bg-makoclaw-error/15 text-makoclaw-error ring-1 ring-makoclaw-error/30'
  return 'bg-makoclaw-success/15 text-makoclaw-success ring-1 ring-makoclaw-success/30'
})

const formattedArgs = computed(() => JSON.stringify(props.toolCall.args ?? {}, null, 2))

const formattedResult = computed(() => {
  if (typeof props.toolCall.result === 'string') return props.toolCall.result
  if (props.toolCall.result == null) return ''
  return JSON.stringify(props.toolCall.result, null, 2)
})

const formattedTime = computed(() => {
  if (!props.toolCall.timestamp) return ''
  return new Date(props.toolCall.timestamp).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
})

// Detect image_generate results and parse them
const imageResult = computed(() => {
  if (props.toolCall.name !== 'image_generate') return null
  const raw = props.toolCall.result
  if (!raw) return null
  try {
    const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
    if (parsed.local_path || parsed.url) return parsed
  } catch {}
  return null
})

// Build display URL from local_path using the files API
const imageDisplayUrl = computed(() => {
  if (!imageResult.value) return null
  const path = imageResult.value.local_path
  if (!path) return imageResult.value.url || null
  const encoded = encodeURIComponent(path).replace(/%2F/g, '/')
  return `/api/v1/files/${encoded}?token=${authStore.token}`
})

// Auto-expand when an image result arrives
watch(imageResult, (val) => {
  if (val) localExpanded.value = true
}, { immediate: true })

const downloadImage = () => {
  const path = imageResult.value?.local_path
  if (!path) return
  const encoded = encodeURIComponent(path).replace(/%2F/g, '/')
  window.open(`/api/v1/files/${encoded}?download=true&token=${authStore.token}`, '_blank')
}

function toggleExpanded() {
  localExpanded.value = !localExpanded.value
}
</script>
