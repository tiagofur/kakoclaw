<template>
  <div class="onboarding-view">
    <SetupWizard />
  </div>
</template>

<script setup>
/**
 * OnboardingView - First-time user onboarding experience
 * 
 * This view is shown when:
 * - User completes registration and logs in for the first time
 * - onboarding_completed = false in database
 * - Router guard automatically redirects to /onboarding
 * 
 * Features:
 * - Full 6-step wizard: Welcome → Provider → Workspace → Channel → Preview → Success
 * - Workspace setup with skills selection (optional, can skip)
 * - Channel setup (optional, can skip)
 * - Marks onboarding_completed = true when finished
 * 
 * Difference from SetupView:
 * - SetupView is for degraded mode (quick 2-step provider config when provider missing)
 * - OnboardingView is comprehensive first-time setup including workspace/channels
 */
import SetupWizard from '../components/Setup/SetupWizard.vue'
</script>

<style scoped>
</style>
