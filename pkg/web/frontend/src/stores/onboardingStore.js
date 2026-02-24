import { defineStore } from 'pinia'
import { useAuthStore } from './authStore'

export const useOnboardingStore = defineStore('onboarding', {
  state: () => ({
    currentStep: 0,
    isFirstLogin: false,
    onboardingCompleted: false,
    statusChecked: false,
    loading: false,
    error: null
  }),

  getters: {
    needsOnboarding: (state) => {
      const authStore = useAuthStore()
      // Need onboarding if:
      // 1. User is authenticated
      // 2. Onboarding not completed
      // 3. User info loaded
      return authStore.isAuthenticated && 
             authStore.user && 
             !state.onboardingCompleted
    },

    canSkipCurrentStep: (state) => {
      // Step 2 (Workspace) and Step 3 (Channel) are optional
      return state.currentStep === 2 || state.currentStep === 3
    },

    isOptionalStep: () => (step) => {
      // Steps: 0 = Welcome, 1 = Provider (required), 2 = Workspace (optional), 3 = Channel (optional), 4 = Preview, 5 = Success
      return step === 2 || step === 3
    }
  },

  actions: {
    async checkOnboardingStatus() {
      const authStore = useAuthStore()
      
      if (!authStore.isAuthenticated) {
        this.onboardingCompleted = true
        return
      }

      this.loading = true
      this.error = null

      try {
        const response = await fetch('/api/v1/auth/me', {
          headers: {
            'Authorization': `Bearer ${authStore.token}`
          }
        })

        if (!response.ok) {
          throw new Error('Failed to check onboarding status')
        }

        const data = await response.json()
        this.onboardingCompleted = data.onboarding_completed || false
        
        // If onboarding not completed on first check after login, mark as first login
        if (!this.onboardingCompleted && !this.isFirstLogin) {
          this.isFirstLogin = true
        }
      } catch (err) {
        console.error('Failed to check onboarding status:', err)
        this.error = err.message
      } finally {
        this.loading = false
        this.statusChecked = true
      }
    },

    async completeOnboarding() {
      const authStore = useAuthStore()
      
      if (!authStore.isAuthenticated) {
        throw new Error('User not authenticated')
      }

      this.loading = true
      this.error = null

      try {
        const response = await fetch('/api/v1/me/onboarding/complete', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${authStore.token}`,
            'Content-Type': 'application/json'
          }
        })

        if (!response.ok) {
          throw new Error('Failed to mark onboarding as complete')
        }

        const data = await response.json()
        
        if (data.success) {
          this.onboardingCompleted = true
          this.isFirstLogin = false
          return true
        }
        
        throw new Error(data.message || 'Failed to complete onboarding')
      } catch (err) {
        console.error('Failed to complete onboarding:', err)
        this.error = err.message
        throw err
      } finally {
        this.loading = false
      }
    },

    skipStep() {
      if (this.canSkipCurrentStep) {
        this.currentStep++
      }
    },

    goToStep(step) {
      if (step >= 0 && step <= 5) {
        this.currentStep = step
      }
    },

    nextStep() {
      if (this.currentStep < 5) {
        this.currentStep++
      }
    },

    previousStep() {
      if (this.currentStep > 0) {
        this.currentStep--
      }
    },

    resetOnboarding() {
      this.currentStep = 0
      this.isFirstLogin = false
      this.error = null
    }
  }
})
