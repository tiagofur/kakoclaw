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

    <!-- Platform Sections -->
    <div class="space-y-3">
      <div
        v-for="platform in availablePlatforms"
        :key="platform.id"
        class="glass-panel rounded-2xl overflow-hidden border border-makoclaw-border/50 transition-all duration-300 hover:border-makoclaw-accent/30"
      >
        <!-- Platform Header -->
        <div
          class="flex items-center gap-3 p-4 cursor-pointer hover:bg-makoclaw-surface/50 transition-colors select-none"
          @click="togglePlatform(platform.id)"
        >
          <!-- Background Glow -->
          <div
            class="w-9 h-9 rounded-xl flex items-center justify-center text-xl shadow-md border transition-all duration-300"
            :class="getAccounts(platform.id).length > 0
              ? 'bg-makoclaw-accent border-makoclaw-accent/20 text-white shadow-makoclaw-accent/30'
              : 'bg-makoclaw-surface border-makoclaw-border/50 text-makoclaw-text-secondary'"
          >
            {{ platform.icon }}
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="font-semibold text-sm text-makoclaw-text">{{ platform.name }}</h3>
            <span class="text-[10px] text-makoclaw-text-secondary/60 uppercase tracking-wide">
              {{ getAccounts(platform.id).length > 0
                ? `${getAccounts(platform.id).length} account${getAccounts(platform.id).length > 1 ? 's' : ''}`
                : platform.subtitle }}
            </span>
          </div>
          <div
            v-if="getAccounts(platform.id).length > 0"
            class="w-5 h-5 rounded-full bg-makoclaw-accent flex items-center justify-center text-white text-[10px] font-bold mr-1"
          >
            {{ getAccounts(platform.id).length }}
          </div>
          <ChevronDownIcon
            class="w-4 h-4 text-makoclaw-text-secondary transition-transform duration-300"
            :class="{ 'rotate-180': expanded[platform.id] }"
          />
        </div>

        <!-- Expanded: Account Rows + Add button -->
        <Transition name="expand">
          <div
            v-if="expanded[platform.id]"
            class="border-t border-makoclaw-border/30"
          >
            <!-- Account rows -->
            <div
              v-for="{ alias, configured } in getAccounts(platform.id)"
              :key="alias"
              class="flex items-center gap-3 px-4 py-3 hover:bg-makoclaw-surface/30 transition-colors border-b border-makoclaw-border/20 last:border-b-0"
            >
              <div class="flex-1 min-w-0">
                <span class="text-sm font-medium text-makoclaw-text truncate block">{{ alias }}</span>
                <span class="text-[10px] text-makoclaw-text-secondary/50 font-mono">{{ platform.id }}:{{ alias }}</span>
              </div>

              <span
                v-if="configured"
                class="inline-flex items-center gap-1 text-makoclaw-accent text-[9px] font-black uppercase tracking-widest"
              >
                <CheckCircleIcon class="w-3.5 h-3.5" /> active
              </span>
              <span
                v-else
                class="inline-flex items-center gap-1 text-makoclaw-text-secondary/50 text-[9px] font-black uppercase tracking-widest"
              >
                <ExclamationCircleIcon class="w-3.5 h-3.5" /> incomplete
              </span>

              <!-- Edit -->
              <button
                class="p-1.5 rounded-lg hover:bg-makoclaw-surface-hover transition-colors text-makoclaw-text-secondary hover:text-makoclaw-text"
                title="Edit"
                @click.stop="$emit('config', { platform, alias })"
              >
                <PencilSquareIcon class="w-3.5 h-3.5" />
              </button>
              <!-- Delete -->
              <button
                class="p-1.5 rounded-lg hover:bg-red-500/10 transition-colors text-makoclaw-text-secondary/40 hover:text-red-400"
                title="Remove"
                @click.stop="$emit('delete', { platform, alias })"
              >
                <TrashIcon class="w-3.5 h-3.5" />
              </button>
            </div>

            <!-- Add account row -->
            <div class="px-4 py-3">
              <button
                class="flex items-center gap-2 text-xs font-black text-makoclaw-accent hover:text-makoclaw-accent/80 transition-colors uppercase tracking-widest"
                @click.stop="$emit('config', { platform, alias: null })"
              >
                <PlusCircleIcon class="w-4 h-4" />
                Add account
              </button>
            </div>
          </div>
        </Transition>
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
          Each account is referenced as <code class="text-makoclaw-accent">platform:alias</code> (e.g. <code class="text-makoclaw-accent">twitter:personal</code>).
          Credentials are stored securely in your personal config.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import {
  CheckCircleIcon,
  ExclamationCircleIcon,
  InformationCircleIcon,
  ChevronDownIcon,
  PlusCircleIcon,
  PencilSquareIcon,
  TrashIcon
} from '@heroicons/vue/24/outline'

const props = defineProps({
  configData: { type: Object, default: () => ({}) },
  saving: { type: Boolean, default: false }
})

defineEmits(['config', 'delete'])

const availablePlatforms = [
  { id: 'twitter',  name: 'Twitter / X', icon: '𝕏',  subtitle: 'Microblogging' },
  { id: 'bluesky',  name: 'Bluesky',     icon: '🦋', subtitle: 'AT Protocol' },
  { id: 'linkedin', name: 'LinkedIn',    icon: '💼', subtitle: 'Professional Network' },
  { id: 'facebook', name: 'Facebook',    icon: '📘', subtitle: 'Pages' }
]

// Expanded state — auto-open platforms that have accounts configured
const expanded = ref(
  Object.fromEntries(availablePlatforms.map(p => [p.id, false]))
)

const togglePlatform = (id) => {
  expanded.value[id] = !expanded.value[id]
}

const getAccounts = (id) => {
  const accounts = props.configData?.tools?.social_media?.[id] || {}
  return Object.entries(accounts).map(([alias, info]) => ({
    alias,
    configured: info?.configured === true
  }))
}
</script>

<style scoped>
@keyframes fade-in-up {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
.animate-fade-in-up {
  animation: fade-in-up 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.expand-enter-active,
.expand-leave-active {
  transition: all 0.25s ease;
  overflow: hidden;
}
.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
}
.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  max-height: 600px;
}
</style>
