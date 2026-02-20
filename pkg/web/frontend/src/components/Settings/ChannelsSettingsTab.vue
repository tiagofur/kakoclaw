<template>
  <div class="space-y-6 max-w-5xl mx-auto animate-fadeIn">
     <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <div 
          v-for="channel in availableChannels" 
          :key="channel.id"
          class="glass-panel rounded-2xl p-6 transition-all duration-300 hover:shadow-xl hover:-translate-y-1 group"
        >
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center space-x-4">
              <div 
                class="w-12 h-12 rounded-xl flex items-center justify-center bg-kakoclaw-bg border border-kakoclaw-border/50 text-kakoclaw-text-secondary transition-all"
                :class="{'bg-kakoclaw-accent !text-white border-transparent shadow-lg shadow-kakoclaw-accent/30': channels[channel.id]?.enabled}"
                v-html="channel.icon"
              ></div>
              <h3 class="font-bold text-sm text-kakoclaw-text tracking-wide">{{ channel.name }}</h3>
            </div>
            <button 
              @click="$emit('toggle', channel.id)"
              class="relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors duration-200 focus:outline-none"
              :class="channels[channel.id]?.enabled ? 'bg-kakoclaw-success' : 'bg-kakoclaw-border'"
            >
              <span 
                aria-hidden="true" 
                class="pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition duration-200"
                :class="channels[channel.id]?.enabled ? 'translate-x-5' : 'translate-x-0'"
              ></span>
            </button>
          </div>
          <p class="text-[11px] text-kakoclaw-text-secondary mb-8 h-8 line-clamp-2 leading-relaxed opacity-70">{{ channel.description }}</p>
          <div class="flex gap-2">
            <button 
              @click="$emit('config', channel)"
              class="flex-1 py-2 text-[10px] font-bold uppercase tracking-widest border border-kakoclaw-border/50 rounded-xl hover:bg-kakoclaw-bg transition-all flex items-center justify-center text-kakoclaw-text group/btn"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5 mr-2 transition-transform group-hover/btn:rotate-45" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              Config
            </button>
            <div 
              v-if="channels[channel.id]?.enabled"
              class="px-3 py-2 border border-emerald-500/20 rounded-xl bg-emerald-500/10 flex items-center justify-center"
              title="Channel Active"
            >
              <span class="w-2 h-2 rounded-full bg-emerald-500 shadow-lg shadow-emerald-500/50 animate-pulse"></span>
            </div>
          </div>
        </div>
     </div>
  </div>
</template>

<script setup>
defineProps({
  availableChannels: {
    type: Array,
    required: true
  },
  channels: {
    type: Object,
    required: true
  }
})
defineEmits(['toggle', 'config'])
</script>
