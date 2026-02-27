<template>
  <div class="space-y-5 max-w-5xl mx-auto animate-fade-in-up">
     <!-- Section Header Decor -->
     <div class="flex items-center gap-4 mb-2 opacity-50">
        <div class="h-[1px] flex-1 bg-gradient-to-r from-transparent to-makoclaw-border"></div>
        <span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary">Channels</span>
        <div class="h-[1px] flex-1 bg-gradient-to-l from-transparent to-makoclaw-border"></div>
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
          ></div>

          <div class="flex items-center justify-between mb-4 relative z-10">
            <div class="flex items-center gap-3">
              <div 
                class="w-10 h-10 rounded-xl flex items-center justify-center bg-makoclaw-surface border border-makoclaw-border/50 text-makoclaw-text-secondary transition-all duration-300 group-hover:scale-105 group-hover:rotate-3 shadow-md"
                :class="{'!bg-makoclaw-accent !border-makoclaw-accent/20 !text-white shadow-makoclaw-accent/30': channels[channel.id]?.enabled}"
                v-html="channel.icon"
              ></div>
              <div>
                <h3 class="font-semibold text-base text-makoclaw-text tracking-tight">{{ channel.name }}</h3>
                <span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary/50">Uplink Channel</span>
              </div>
            </div>

            <!-- Modern Toggle -->
            <button 
              @click="$emit('toggle', channel.id)"
              class="relative inline-flex h-7 w-12 items-center rounded-full transition-all duration-500 focus:outline-none"
              :class="channels[channel.id]?.enabled ? 'bg-makoclaw-accent' : 'bg-makoclaw-bg/80 border-2 border-makoclaw-border'"
            >
              <span 
                class="inline-block h-5 w-5 rounded-full bg-white shadow-lg transform transition-transform duration-500"
                :class="channels[channel.id]?.enabled ? 'translate-x-[22px]' : 'translate-x-[2px]'"
              ></span>
            </button>
          </div>

          <p class="text-xs font-medium text-makoclaw-text-secondary/60 mb-5 leading-relaxed h-10 line-clamp-2 relative z-10 group-hover:text-makoclaw-text-secondary transition-colors">
            {{ channel.description }}
          </p>

          <div class="mt-auto flex items-center gap-3 relative z-10">
            <button 
              @click="$emit('config', channel)"
              class="flex-1 py-2.5 text-[10px] font-medium uppercase tracking-wide bg-makoclaw-bg/40 border border-makoclaw-border/50 rounded-xl hover:border-makoclaw-accent/40 hover:bg-makoclaw-surface transition-all flex items-center justify-center text-makoclaw-text group/btn active:scale-95"
            >
              <IconConfig class="h-4 w-4 mr-3 transition-transform group-hover/btn:rotate-180 duration-1000" />
              Tune Channel
            </button>
            
            <div 
              v-if="channels[channel.id]?.enabled"
              class="w-10 h-10 rounded-xl bg-makoclaw-accent/10 border border-makoclaw-accent/20 flex items-center justify-center text-makoclaw-accent"
              title="Active Link"
            >
              <div class="relative">
                <IconActive class="w-5 h-5 animate-pulse" />
                <span class="absolute -top-1 -right-1 w-2 h-2 bg-makoclaw-accent rounded-full animate-ping"></span>
              </div>
            </div>
          </div>
        </div>
     </div>

     <!-- Security Note -->
     <div class="glass-panel p-4 rounded-2xl bg-amber-500/5 border border-amber-500/10 flex items-center gap-4 mt-6 animate-fade-in-up" style="animation-delay: 0.2s">
        <div class="p-2 rounded-xl bg-amber-500/10 text-amber-500">
           <IconShield class="w-4 h-4" />
        </div>
        <div>
           <h4 class="text-xs font-medium tracking-wide text-makoclaw-text">Protocol Warning</h4>
           <p class="text-xs font-medium text-makoclaw-text-secondary/50 mt-1">Changing neural channel configurations may temporarily sever active uplinks. Ensure backup encryption is enabled for sensitive transmissions.</p>
        </div>
     </div>
  </div>
</template>

<script setup>
import { h } from 'vue'

defineProps({
  availableChannels: { type: Array, required: true },
  channels: { type: Object, required: true }
})
defineEmits(['toggle', 'config'])

// Premium Icons
const IconConfig = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z' }), h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M15 12a3 3 0 11-6 0 3 3 0 016 0z' })]) }
const IconActive = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M13 10V3L4 14h7v7l9-11h-7z' })]) }
const IconShield = { render: () => h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24', class: 'w-full h-full' }, [h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2.5', d: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z' })]) }
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
