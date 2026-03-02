<template>
  <div class="space-y-6 animate-fade-in-up pb-10 max-w-5xl mx-auto">
    <!-- Section Header Decor -->
    <div class="flex items-center gap-4 mb-2 opacity-50">
      <div class="h-[1px] flex-1 bg-gradient-to-r from-transparent to-makoclaw-border" />
      <span class="text-[9px] font-medium uppercase tracking-wide text-makoclaw-text-secondary">
        Identity & Access
      </span>
      <div class="h-[1px] flex-1 bg-gradient-to-l from-transparent to-makoclaw-border" />
    </div>

    <div class="glass-panel rounded-3xl p-6 border border-makoclaw-border/50 transition-all hover:border-makoclaw-accent/30 relative overflow-hidden group">
      <!-- Decor -->
      <div class="absolute -top-12 -right-12 w-32 h-32 bg-makoclaw-accent/10 rounded-full blur-[50px] group-hover:bg-makoclaw-accent/20 transition-colors duration-700" />

      <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4 mb-6 relative z-10">
        <div class="flex items-center gap-4">
          <!-- The title and icon have been removed to prevent duplication with SettingsView -->
        </div>
        <button 
          class="px-5 py-2.5 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded-xl transition-all shadow-lg shadow-makoclaw-accent/20 text-xs font-bold uppercase tracking-widest flex items-center justify-center gap-2 active:scale-95 group/btn"
          @click="$emit('add-user')"
        >
          <PlusIcon class="w-4 h-4 group-hover/btn:rotate-90 transition-transform duration-300" />
          Register Node
        </button>
      </div>
      
      <div class="overflow-x-auto custom-scrollbar relative z-10 bg-makoclaw-bg/20 rounded-2xl border border-makoclaw-border/30">
        <table class="w-full text-left text-sm whitespace-nowrap">
          <thead class="bg-makoclaw-surface/50 border-b border-makoclaw-border/50 text-[10px] text-makoclaw-text-secondary/80 font-black uppercase tracking-widest">
            <tr>
              <th class="px-6 py-4">
                ID Cluster
              </th>
              <th class="px-6 py-4">
                Identity
              </th>
              <th class="px-6 py-4">
                Authorization
              </th>
              <th class="px-6 py-4">
                Link Status
              </th>
              <th class="px-6 py-4">
                Timestamp
              </th>
              <th class="px-6 py-4 text-right">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-makoclaw-border/30 text-makoclaw-text">
            <tr
              v-for="u in usersList"
              :key="u.id"
              class="hover:bg-makoclaw-accent/5 transition-colors group"
            >
              <td class="px-6 py-4 font-mono text-[10px] text-makoclaw-text-secondary/50">
                {{ u.id }}
              </td>
              <td class="px-6 py-4">
                <span class="font-bold text-sm tracking-tight group-hover:text-makoclaw-accent transition-colors">{{ u.username }}</span>
              </td>
              <td class="px-6 py-4">
                <span
                  class="px-3 py-1 text-[9px] font-black tracking-widest uppercase rounded-lg shadow-sm"
                  :class="u.role === 'admin' ? 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20' : 'bg-makoclaw-accent/10 text-makoclaw-accent border border-makoclaw-accent/20'"
                >
                  {{ u.role }}
                </span>
              </td>
              <td class="px-6 py-4">
                <span
                  v-if="u.blocked"
                  class="px-3 py-1 text-[9px] font-black tracking-widest uppercase rounded-lg bg-red-500/10 text-red-500 border border-red-500/20 flex items-center gap-1.5 w-max"
                >
                  <NoSymbolIcon class="w-3 h-3" /> Offline
                </span>
                <span
                  v-else
                  class="flex items-center gap-1.5 px-3 py-1 text-[9px] font-black tracking-widest uppercase rounded-lg bg-makoclaw-success/10 text-makoclaw-success border border-makoclaw-success/20 w-max shadow-[0_0_10px_rgba(16,185,129,0.1)]"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-makoclaw-success animate-pulse shadow-[0_0_5px_rgba(16,185,129,0.8)]" />
                  Active
                </span>
              </td>
              <td class="px-6 py-4 text-[10px] font-bold text-makoclaw-text-secondary/60">
                {{ formatDate(u.created_at) }}
              </td>
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end gap-2">
                  <button
                    class="p-2 rounded-xl bg-makoclaw-surface border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-makoclaw-accent hover:border-makoclaw-accent/30 transition-all active:scale-95 shadow-sm"
                    title="Modify Identity"
                    @click="$emit('edit-user', u)"
                  >
                    <PencilSquareIcon class="w-4 h-4" />
                  </button>
                  <button
                    v-if="!u.blocked"
                    :disabled="authStore.user?.username === u.username"
                    class="p-2 rounded-xl bg-makoclaw-surface border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-orange-400 hover:border-orange-400/30 transition-all active:scale-95 disabled:opacity-30 shadow-sm"
                    title="Sever Protocol"
                    @click="$emit('block-user', u)"
                  >
                    <ShieldExclamationIcon class="w-4 h-4" />
                  </button>
                  <button
                    v-else
                    class="p-2 rounded-xl bg-makoclaw-surface border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-makoclaw-success hover:border-makoclaw-success/30 transition-all active:scale-95 shadow-sm"
                    title="Restore Integration"
                    @click="$emit('unblock-user', u)"
                  >
                    <CheckCircleIcon class="w-4 h-4" />
                  </button>
                  <button
                    :disabled="authStore.user?.username === u.username"
                    class="p-2 rounded-xl bg-makoclaw-surface border border-makoclaw-border/50 text-makoclaw-text-secondary hover:text-red-500 hover:border-red-500/30 transition-all active:scale-95 disabled:opacity-30 shadow-sm"
                    title="Delete Node"
                    @click="$emit('delete-user', u)"
                  >
                    <TrashIcon class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        
        <div
          v-if="!usersList || usersList.length === 0"
          class="p-8 text-center border-t border-makoclaw-border/30"
        >
          <p class="text-xs font-bold uppercase tracking-widest text-makoclaw-text-secondary/50">
            No external nodes registered
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { 
  PlusIcon,
  PencilSquareIcon,
  ShieldExclamationIcon,
  CheckCircleIcon,
  TrashIcon,
  NoSymbolIcon
} from '@heroicons/vue/24/outline'

import { useAuthStore } from '../../stores/authStore'

defineProps({
  usersList: { type: Array, default: () => [] }
})

defineEmits(['add-user', 'edit-user', 'block-user', 'unblock-user', 'delete-user'])

const authStore = useAuthStore()

const formatDate = (dateString) => {
  if (!dateString) return 'Unknown Protocol'
  const d = new Date(dateString)
  return d.toLocaleDateString('en-US', { 
    year: 'numeric', month: 'short', day: 'numeric', 
    hour: '2-digit', minute: '2-digit' 
  })
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

.custom-scrollbar::-webkit-scrollbar { height: 6px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(156, 163, 175, 0.2); border-radius: 10px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
</style>
