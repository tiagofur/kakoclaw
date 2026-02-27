<template>
  <div class="space-y-5 max-w-5xl mx-auto animate-fade-in-up">
    <!-- Section Header Decor -->
    <div class="flex items-center gap-4 mb-2 opacity-50">
      <div class="h-[1px] flex-1 bg-gradient-to-r from-transparent to-makoclaw-border" />
      <span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary">
        Channels
      </span>
      <div class="h-[1px] flex-1 bg-gradient-to-l from-transparent to-makoclaw-border" />
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div
        v-for="channel in availableChannels"
        :key="channel.id"
        class="glass-panel rounded-2xl p-5 transition-all duration-300 hover:shadow-xl hover:shadow-makoclaw-accent/5 hover:-translate-y-1 group relative overflow-hidden flex flex-col h-full border border-makoclaw-border/50 hover:border-makoclaw-accent/40"
      >
        <!-- Background Glow -->
        <div
          class="absolute -top-12 -right-12 w-32 h-32 blur-[50px] rounded-full transition-all duration-700 opacity-0 group-hover:opacity-20"
          :class="channels[channel.id]?.enabled ? 'bg-makoclaw-accent' : 'bg-makoclaw-text-secondary/50'"
        />

        <div class="flex items-center justify-between mb-4 relative z-10">
          <div class="flex items-center gap-3">
            <div
              class="w-10 h-10 rounded-xl flex items-center justify-center bg-makoclaw-surface border border-makoclaw-border/50 text-makoclaw-text-secondary transition-all duration-300 group-hover:scale-105 group-hover:rotate-3 shadow-md"
              :class="{'!bg-makoclaw-accent !border-makoclaw-accent/20 !text-white shadow-makoclaw-accent/30': channels[channel.id]?.enabled}"
              v-html="channel.icon"
            />
            <div>
              <h3 class="font-semibold text-base text-makoclaw-text tracking-tight">
                {{ channel.name }}
              </h3>
              <span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/50">
                Uplink Channel
              </span>
            </div>
          </div>

          <!-- Modern Toggle -->
          <button
            class="relative inline-flex h-7 w-12 items-center rounded-full transition-all duration-500 focus:outline-none"
            :class="channels[channel.id]?.enabled ? 'bg-makoclaw-accent' : 'bg-makoclaw-bg/80 border-2 border-makoclaw-border'"
            @click="$emit('toggle', channel.id)"
          >
            <span
              class="inline-block h-5 w-5 rounded-full bg-white shadow-lg transform transition-transform duration-500"
              :class="channels[channel.id]?.enabled ? 'translate-x-[22px]' : 'translate-x-[2px]'"
            />
          </button>
        </div>

        <p class="text-xs font-medium text-makoclaw-text-secondary/60 mb-5 leading-relaxed h-10 line-clamp-2 relative z-10 group-hover:text-makoclaw-text-secondary transition-colors">
          {{ channel.description }}
        </p>

        <div class="mt-auto flex items-center gap-3 relative z-10">
          <button
            class="flex-1 py-3 text-xs font-black uppercase tracking-widest bg-makoclaw-surface border border-makoclaw-border/10 rounded-xl hover:border-makoclaw-accent/40 hover:bg-makoclaw-surface transition-all flex items-center justify-center text-makoclaw-text group/btn active:scale-95 shadow-lg"
            @click="$emit('config', channel)"
          >
            <AdjustmentsHorizontalIcon class="h-4 w-4 mr-3 transition-transform group-hover/btn:rotate-180 duration-1000" />
            Tune Channel
          </button>
          
          <div
            v-if="channels[channel.id]?.enabled"
            class="w-12 h-12 rounded-xl bg-makoclaw-accent/10 border border-makoclaw-accent/20 flex items-center justify-center text-makoclaw-accent shadow-inner"
            title="Active Link"
          >
            <div class="relative">
              <BoltIcon class="w-5 h-5 animate-pulse" />
              <span class="absolute -top-1 -right-1 w-2.5 h-2.5 bg-makoclaw-accent rounded-full animate-ping" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Security Note -->
    <div
      class="glass-panel p-4 rounded-2xl bg-amber-500/5 border border-amber-500/10 flex items-center gap-4 mt-6 animate-fade-in-up"
      style="animation-delay: 0.2s"
    >
      <div class="p-2.5 rounded-xl bg-amber-500/10 text-amber-500 shadow-lg shadow-amber-500/5">
        <ShieldCheckIcon class="w-5 h-5" />
      </div>
      <div>
        <h4 class="text-[10px] font-black uppercase tracking-widest text-makoclaw-text mb-1">
          Protocol Warning
        </h4>
        <p class="text-xs font-medium text-makoclaw-text-secondary/50 leading-relaxed">
          Changing neural channel configurations may temporarily sever active uplinks. Ensure backup encryption is enabled for sensitive transmissions.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { 
  AdjustmentsHorizontalIcon, 
  BoltIcon, 
  ShieldCheckIcon 
} from '@heroicons/vue/24/outline'

defineProps({
  availableChannels: { type: Array, required: true },
  channels: { type: Object, required: true }
})
defineEmits(['toggle', 'config'])
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
