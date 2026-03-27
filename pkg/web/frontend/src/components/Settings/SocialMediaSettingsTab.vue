<template>
  <div class="space-y-5 max-w-4xl mx-auto animate-fade-in-up pb-10">
    <!-- Section Header Decor -->
    <div class="flex items-center gap-4 mb-2 opacity-50">
      <div class="h-[1px] flex-1 bg-gradient-to-r from-transparent to-makoclaw-border" />
      <span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary">
        Social Platforms
      </span>
      <div class="h-[1px] flex-1 bg-gradient-to-l from-transparent to-makoclaw-border" />
    </div>

    <!-- Platforms Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div
        v-for="platform in availablePlatforms"
        :key="platform.id"
        class="glass-panel rounded-2xl p-5 transition-all duration-300 hover:shadow-xl hover:shadow-makoclaw-accent/5 hover:-translate-y-1 group relative overflow-hidden flex flex-col h-full border border-makoclaw-border/50 hover:border-makoclaw-accent/40"
      >
        <!-- Background Glow -->
        <div
          class="absolute -top-12 -right-12 w-32 h-32 blur-[50px] rounded-full transition-all duration-700 opacity-0 group-hover:opacity-20"
          :class="isConfigured(platform.id) ? 'bg-makoclaw-accent' : 'bg-makoclaw-text-secondary/50'"
        />

        <div class="flex items-center justify-between mb-8 relative z-10">
          <div class="flex items-center gap-3">
            <div
              class="w-10 h-10 rounded-xl flex items-center justify-center bg-makoclaw-surface border border-makoclaw-border/50 text-makoclaw-text-secondary transition-all duration-300 group-hover:scale-105 group-hover:rotate-3 shadow-md text-xl"
              :class="{'!bg-makoclaw-accent !border-makoclaw-accent/20 !text-white shadow-makoclaw-accent/30': isConfigured(platform.id)}"
            >
              {{ platform.icon }}
            </div>
            <div>
              <h3 class="font-semibold text-base text-makoclaw-text tracking-tight">
                {{ platform.name }}
              </h3>
              <span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/50">
                {{ platform.subtitle }}
              </span>
            </div>
          </div>

          <!-- Configured indicator -->
          <div
            v-if="isConfigured(platform.id)"
            class="w-7 h-7 rounded-full bg-makoclaw-accent/10 border border-makoclaw-accent/20 flex items-center justify-center text-makoclaw-accent text-xs"
            title="Configured"
          >
            <CheckCircleIcon class="w-4 h-4" />
          </div>
          <div
            v-else
            class="w-7 h-7 rounded-full bg-makoclaw-text-secondary/10 border border-makoclaw-border/30 flex items-center justify-center"
            title="Not configured"
          >
            <ExclamationCircleIcon class="w-4 h-4 text-makoclaw-text-secondary/50" />
          </div>
        </div>

        <div class="mt-auto flex items-center gap-3 relative z-10">
          <button
            class="flex-1 py-3 text-xs font-black uppercase tracking-widest bg-makoclaw-surface border border-makoclaw-border/10 rounded-xl hover:border-makoclaw-accent/40 hover:bg-makoclaw-surface transition-all flex items-center justify-center text-makoclaw-text group/btn active:scale-95 shadow-lg"
            @click="$emit('config', platform)"
          >
            <AdjustmentsHorizontalIcon class="h-4 w-4 mr-3 transition-transform group-hover/btn:rotate-180 duration-1000" />
            Configure
          </button>

          <div
            v-if="isConfigured(platform.id)"
            class="w-12 h-12 rounded-xl bg-makoclaw-accent/10 border border-makoclaw-accent/20 flex items-center justify-center text-makoclaw-accent shadow-inner shrink-0"
            title="Connected"
          >
            <div class="relative">
              <SignalIcon class="w-5 h-5 animate-pulse" />
              <span class="absolute -top-1 -right-1 w-2.5 h-2.5 bg-makoclaw-accent rounded-full animate-ping" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Info Note -->
    <div
      class="glass-panel p-4 rounded-2xl bg-blue-500/5 border border-blue-500/10 flex items-center gap-4 mt-6 animate-fade-in-up"
      style="animation-delay: 0.2s"
    >
      <div class="p-2.5 rounded-xl bg-blue-500/10 text-blue-400 shadow-lg shadow-blue-500/5 shrink-0">
        <InformationCircleIcon class="w-5 h-5" />
      </div>
      <div>
        <h4 class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text mb-1">
          Platform Credentials
        </h4>
        <p class="text-xs font-medium text-makoclaw-text-secondary/50 leading-relaxed">
          Credentials are stored securely in your personal config. The agent can publish to configured platforms using the <code class="text-makoclaw-accent">social_post</code> tool.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import {
  AdjustmentsHorizontalIcon,
  CheckCircleIcon,
  ExclamationCircleIcon,
  SignalIcon,
  InformationCircleIcon
} from '@heroicons/vue/24/outline'

const props = defineProps({
  configData: { type: Object, default: () => ({}) },
  saving: { type: Boolean, default: false }
})

const availablePlatforms = [
  { id: 'twitter',  name: 'Twitter / X',  icon: '𝕏',  subtitle: 'Microblogging' },
  { id: 'bluesky',  name: 'Bluesky',      icon: '🦋', subtitle: 'AT Protocol' },
  { id: 'linkedin', name: 'LinkedIn',     icon: '💼', subtitle: 'Professional Network' },
  { id: 'facebook', name: 'Facebook',     icon: '📘', subtitle: 'Pages' }
]

const isConfigured = (id) => props.configData?.tools?.social_media?.[id]?.configured === true

defineEmits(['config'])
</script>

<style scoped>
@keyframes fade-in-up {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
.animate-fade-in-up {
  animation: fade-in-up 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>
