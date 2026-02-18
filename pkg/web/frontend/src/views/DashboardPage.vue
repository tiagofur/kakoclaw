<template>
  <Sidebar>
    <!-- Desktop Two-Pane Layout -->
    <div class="hidden md:flex h-full">
      <!-- Chat Pane (50%) -->
      <div class="flex-1 border-r border-picoclaw-border overflow-hidden">
        <ChatTab />
      </div>

      <!-- Tasks Pane (50%) -->
      <div class="flex-1 overflow-hidden">
        <TasksTab />
      </div>
    </div>

    <!-- Mobile Tab Layout -->
    <div class="md:hidden flex flex-col h-full">
      <!-- Tab Buttons -->
      <div class="flex border-b border-picoclaw-border bg-picoclaw-surface">
        <button
          @click="setActiveTab('chat')"
          class="flex-1 px-4 py-3 font-medium transition-smooth"
          :class="activeTab === 'chat' 
            ? 'border-b-2 border-picoclaw-accent text-picoclaw-accent'
            : 'text-picoclaw-text-secondary hover:text-picoclaw-text'"
        >
          Chat
        </button>
        <button
          @click="setActiveTab('tasks')"
          class="flex-1 px-4 py-3 font-medium transition-smooth"
          :class="activeTab === 'tasks' 
            ? 'border-b-2 border-picoclaw-accent text-picoclaw-accent'
            : 'text-picoclaw-text-secondary hover:text-picoclaw-text'"
        >
          Tasks
        </button>
      </div>

      <!-- Content Area -->
      <div class="flex-1 overflow-hidden">
        <ChatTab v-if="activeTab === 'chat'" />
        <TasksTab v-else-if="activeTab === 'tasks'" />
      </div>
    </div>
  </Sidebar>
</template>

<script setup>
import { onMounted } from 'vue'
import { useUIStore } from '../stores/uiStore'
import Sidebar from '../components/Layout/Sidebar.vue'
import ChatTab from './ChatTab.vue'
import TasksTab from './TasksTab.vue'

const uiStore = useUIStore()
const activeTab = uiStore.activeTab

const setActiveTab = (tab) => uiStore.setActiveTab(tab)

onMounted(() => {
  uiStore.restoreUIPreferences()
})
</script>
