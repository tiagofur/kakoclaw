<template>
  <router-view />
</template>

<script setup>
import { useAuthStore } from './stores/authStore'
import { useUIStore } from './stores/uiStore'
import { onMounted } from 'vue'

const authStore = useAuthStore()
const uiStore = useUIStore()

onMounted(() => {
  // Restore session and UI preferences
  authStore.restoreSession()
  uiStore.restoreUIPreferences()

  // Request Notification permission for task push events
  if ('Notification' in window && Notification.permission === 'default') {
    Notification.requestPermission()
  }
})
</script>

<style scoped>
</style>
