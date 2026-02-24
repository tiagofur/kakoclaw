<template>
  <router-view />
</template>

<script setup>
import { useAuthStore } from './stores/authStore'
import { useUIStore } from './stores/uiStore'
import { useConfigStore } from './stores/configStore'
import { useOnboardingStore } from './stores/onboardingStore'
import { onMounted } from 'vue'

const authStore = useAuthStore()
const uiStore = useUIStore()
const configStore = useConfigStore()
const onboardingStore = useOnboardingStore()

onMounted(async () => {
  // Restore session and UI preferences
  authStore.restoreSession()
  uiStore.restoreUIPreferences()

  // Check configuration status and onboarding
  if (authStore.isAuthenticated) {
    try {
      await configStore.checkStatus()
      await onboardingStore.checkOnboardingStatus()
    } catch (error) {
      console.error('Failed to check configuration status:', error)
    }
  }

  // Request Notification permission for task push events
  if ('Notification' in window && Notification.permission === 'default') {
    Notification.requestPermission()
  }
})
</script>

<style scoped>
</style>
