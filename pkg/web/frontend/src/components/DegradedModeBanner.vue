<template>
  <div
    v-if="shouldShowBanner"
    class="bg-makoclaw-warning/10 border-b-2 border-makoclaw-warning animate-slideDown sticky top-0 z-banner backdrop-blur-sm px-6 py-4 shadow-sm"
  >
    <div class="flex items-center gap-4 max-w-[1200px] mx-auto flex-wrap md:flex-nowrap">
      <div class="text-[32px] md:text-2xl shrink-0">
        ⚠️
      </div>
      <div class="flex-1">
        <h3 class="text-makoclaw-text font-semibold text-lg md:text-base mb-1">
          LLM Provider Not Configured
        </h3>
        <p class="text-makoclaw-text-secondary text-sm">
          {{ configStore.reason || 'Configure your LLM provider to enable AI features.' }}
        </p>
      </div>
      <button
        class="bg-makoclaw-warning text-white hover:bg-makoclaw-warning/80 rounded-lg px-5 py-2.5 font-semibold transition-all active:scale-[0.97] whitespace-nowrap w-full md:w-auto"
        @click="showSetupWizard"
      >
        Configure Now
      </button>
      <button
        v-if="dismissible"
        class="text-makoclaw-text-secondary hover:text-makoclaw-text text-xl transition-colors px-2 leading-none"
        @click="dismiss"
      >
        ×
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useConfigStore } from '../stores/configStore'
import { useOnboardingStore } from '../stores/onboardingStore'
import { useRouter } from 'vue-router'

defineProps({
  dismissible: {
    type: Boolean,
    default: false
  }
})

const configStore = useConfigStore()
const onboardingStore = useOnboardingStore()
const router = useRouter()
const dismissed = ref(false)

// Only show degraded mode banner if:
// 1. Configuration is needed (no provider)
// 2. Not dismissed by user
// 3. Onboarding is not needed (if onboarding needed, router will redirect)
const shouldShowBanner = computed(() => {
  return configStore.needsConfiguration() &&
         !dismissed.value &&
         !onboardingStore.needsOnboarding
})

function showSetupWizard() {
  router.push('/settings?tab=providers')
}

function dismiss() {
  dismissed.value = true
}
</script>
