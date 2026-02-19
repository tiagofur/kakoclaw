<template>
  <div class="flex h-screen bg-kakoclaw-bg text-kakoclaw-text overflow-hidden">
    <!-- Sidebar -->
    <Sidebar />

    <!-- Toast Notifications -->
    <ToastContainer />

    <!-- Main Content Area -->
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden relative">
      
      <!-- Mobile Header -->
      <header class="md:hidden h-14 bg-kakoclaw-surface border-b border-kakoclaw-border flex items-center justify-between px-4 flex-shrink-0 z-20">
         <div class="font-bold text-lg text-kakoclaw-accent">KakoClaw</div>
         <button @click="uiStore.toggleSidebar()" class="p-2 text-kakoclaw-text hover:bg-kakoclaw-border rounded">
           <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
             <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16m-7 6h7" />
           </svg>
         </button>
      </header>

      <!-- Page Content -->
      <main class="flex-1 overflow-auto relative scroll-smooth p-4 md:p-6">
        <router-view v-slot="{ Component }">
          <transition 
            enter-active-class="transition ease-out duration-300 transform" 
            enter-from-class="opacity-0 translate-y-2" 
            enter-to-class="opacity-100 translate-y-0"
            leave-active-class="transition ease-in duration-200 transform" 
            leave-from-class="opacity-100 translate-y-0" 
            leave-to-class="opacity-0 -translate-y-2"
            mode="out-in"
          >
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useUIStore } from '../../stores/uiStore'
import Sidebar from './Sidebar.vue'
import ToastContainer from './ToastContainer.vue'

const uiStore = useUIStore()

onMounted(() => {
  // Any global init
})
</script>

<style scoped>
/* Custom Scrollbar */
main::-webkit-scrollbar {
  width: 8px;
}
main::-webkit-scrollbar-track {
  background: transparent;
}
main::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5); /* gray-400 */
  border-radius: 4px;
}
main::-webkit-scrollbar-thumb:hover {
  background-color: rgba(156, 163, 175, 0.8);
}
</style>
