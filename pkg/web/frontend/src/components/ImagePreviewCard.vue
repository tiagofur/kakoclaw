<template>
  <figure
    class="glass-panel group/glow relative overflow-hidden rounded-3xl border border-makoclaw-border/50 p-3 sm:p-4 transition-all duration-300 hover:border-indigo-500/30 hover:shadow-xl hover:shadow-indigo-500/10"
  >
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute inset-0 opacity-25 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-indigo-500/20 via-transparent to-transparent" />
      <div class="absolute inset-0 opacity-20 bg-[radial-gradient(ellipse_at_bottom_left,_var(--tw-gradient-stops))] from-violet-500/20 via-transparent to-transparent" />
    </div>

    <div class="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-0 bg-gradient-to-b from-transparent via-indigo-500/40 to-transparent group-hover/glow:h-2/3 transition-all duration-300" />
    <div class="absolute bottom-0 left-0 h-[2px] w-0 bg-gradient-to-r from-indigo-500 to-violet-500 group-hover/glow:w-full transition-all duration-500 opacity-40" />
    <div class="absolute -top-6 -right-6 w-16 h-16 bg-gradient-to-br from-indigo-500 to-violet-500 rounded-full opacity-0 blur-[15px] group-hover/glow:opacity-10 transition-all duration-500" />

    <div class="relative z-10 overflow-hidden rounded-2xl border border-white/10 bg-makoclaw-bg/40 ring-1 ring-white/10">
      <img
        v-if="!hasError"
        :src="src"
        :alt="caption || 'Image preview'"
        class="w-full h-auto max-w-full max-h-[70vh] object-contain bg-makoclaw-bg/20"
        @error="handleImageError"
      >

      <div
        v-else
        class="flex min-h-[220px] flex-col items-center justify-center gap-3 px-6 py-10 text-center text-makoclaw-text-secondary"
      >
        <div class="flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-makoclaw-error/15 to-orange-500/15 ring-1 ring-white/10">
          <ExclamationTriangleIcon class="h-8 w-8 text-makoclaw-error" />
        </div>
        <div class="space-y-1">
          <p class="text-sm font-semibold text-makoclaw-text">Preview unavailable</p>
          <p class="text-xs sm:text-sm text-makoclaw-text-secondary">The image could not be loaded.</p>
        </div>
      </div>

      <div
        v-if="!hasError && effectiveDownloadUrl"
        class="pointer-events-none absolute inset-0 bg-gradient-to-t from-slate-950/70 via-slate-900/10 to-transparent opacity-0 transition-all duration-300 group-hover/glow:opacity-100 group-focus-within/glow:opacity-100"
      />

      <div
        v-if="!hasError && effectiveDownloadUrl"
        class="pointer-events-none absolute inset-x-0 bottom-0 flex justify-end p-3 sm:p-4 opacity-0 transition-all duration-300 group-hover/glow:opacity-100 group-focus-within/glow:opacity-100"
      >
        <button
          type="button"
          class="pointer-events-auto inline-flex min-h-[40px] items-center gap-2 rounded-xl border border-white/10 bg-makoclaw-surface/30 px-3 py-2 text-xs font-bold text-white shadow-lg shadow-indigo-500/20 ring-1 ring-white/10 backdrop-blur-xl transition-all duration-300 hover:-translate-y-[1px] hover:bg-makoclaw-surface/40 active:scale-95"
          @click="downloadImage"
        >
          <ArrowDownTrayIcon class="h-4 w-4" />
          <span>Download</span>
        </button>
      </div>
    </div>

    <figcaption
      v-if="caption"
      class="relative z-10 mt-3 flex items-start gap-2 text-xs sm:text-sm text-makoclaw-text-secondary"
    >
      <PhotoIcon class="mt-0.5 h-4 w-4 flex-shrink-0 text-indigo-400" />
      <span class="leading-relaxed">{{ caption }}</span>
    </figcaption>
  </figure>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import {
  ArrowDownTrayIcon,
  ExclamationTriangleIcon,
  PhotoIcon
} from '@heroicons/vue/24/outline'

const props = defineProps({
  src: {
    type: String,
    required: true
  },
  caption: {
    type: String,
    default: ''
  },
  downloadUrl: {
    type: String,
    default: ''
  }
})

const hasError = ref(false)

const effectiveDownloadUrl = computed(() => props.downloadUrl || props.src)

watch(() => props.src, () => {
  hasError.value = false
})

function handleImageError() {
  hasError.value = true
}

function downloadImage() {
  if (!effectiveDownloadUrl.value) return

  const link = document.createElement('a')
  link.href = effectiveDownloadUrl.value
  link.download = ''
  link.rel = 'noopener noreferrer'
  link.target = '_blank'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
</script>
