<template>
  <div class="bg-kakoclaw-bg/40 border border-kakoclaw-border rounded-xl overflow-hidden transition-all duration-300">
    <button 
      @click="tc.expanded = !tc.expanded"
      class="w-full flex items-center justify-between px-3 md:px-4 py-2 md:py-2.5 text-[10px] md:text-xs font-mono hover:bg-kakoclaw-accent/10 transition-colors"
    >
      <div class="flex items-center gap-2 md:gap-3">
        <div :class="[
          'w-2 h-2 rounded-full',
          tc.status === 'started' ? 'bg-amber-400 animate-pulse' : 
          tc.status === 'error' ? 'bg-red-500' : 'bg-emerald-500'
        ]"></div>
        <div class="flex items-center gap-1.5">
          <span class="text-kakoclaw-text-secondary">Tool:</span>
          <span class="text-kakoclaw-accent font-bold">{{ tc.name }}</span>
        </div>
        <span v-if="tc.status === 'started'" class="text-[9px] text-amber-500/80 italic hidden sm:inline">executing...</span>
      </div>
      <svg 
        class="w-3.5 md:w-4 h-3.5 md:h-4 text-kakoclaw-text-secondary transition-transform duration-300" 
        :class="{ 'rotate-180': tc.expanded }"
        fill="none" stroke="currentColor" viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>
    
    <div v-if="tc.expanded" class="px-3 md:px-4 pb-3 md:pb-4 border-t border-kakoclaw-border/30 animate-fadeIn bg-kakoclaw-bg/20">
      <div class="mt-3 space-y-3">
        <div>
          <div class="text-[9px] md:text-[10px] text-kakoclaw-text-secondary uppercase tracking-wider font-bold mb-1.5 flex items-center gap-1.5">
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" /></svg>
            Arguments
          </div>
          <pre class="bg-black/30 p-2 md:p-3 rounded-lg text-kakoclaw-text/90 overflow-x-auto custom-scrollbar text-[9px] md:text-[11px] leading-tight font-mono">{{ JSON.stringify(tc.args, null, 2) }}</pre>
        </div>
        
        <div v-if="tc.result">
          <div class="text-[9px] md:text-[10px] text-kakoclaw-text-secondary uppercase tracking-wider font-bold mb-1.5 flex items-center gap-1.5">
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            Result
          </div>
          <div class="bg-black/20 p-2 md:p-3 rounded-lg text-kakoclaw-text/80 whitespace-pre-wrap max-h-48 md:max-h-64 overflow-y-auto custom-scrollbar font-mono text-[9px] md:text-[11px] leading-normal border border-kakoclaw-border/10">
            {{ tc.result }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  tc: {
    type: Object,
    required: true
  }
})
</script>
