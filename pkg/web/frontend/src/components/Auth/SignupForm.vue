<template>
  <div class="min-h-screen bg-makoclaw-bg flex items-center justify-center p-4">
    <div class="w-full max-w-md">
      <!-- Logo -->
      <div class="text-center mb-8">
        <h1 class="text-4xl font-bold text-makoclaw-accent mb-2">makoclaw</h1>
        <p class="text-makoclaw-text-secondary">Create your account</p>
      </div>

      <!-- Form Card -->
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-lg p-8 shadow-lg">
        <h2 class="text-2xl font-bold mb-6 text-makoclaw-text">Sign Up</h2>

        <form @submit.prevent="handleSignup" class="space-y-4">
          <!-- Username -->
          <div>
            <label for="username" class="block text-sm font-medium mb-2 text-makoclaw-text">
              Username
            </label>
            <input
              v-model="form.username"
              id="username"
              type="text"
              placeholder="Choose a username"
              class="w-full px-4 py-2 bg-makoclaw-bg border border-makoclaw-border rounded focus-ring text-makoclaw-text placeholder-makoclaw-text-secondary"
              required
              :disabled="isLoading"
            />
          </div>

          <!-- Email -->
          <div>
            <label for="email" class="block text-sm font-medium mb-2 text-makoclaw-text">
              Email
            </label>
            <input
              v-model="form.email"
              id="email"
              type="email"
              placeholder="your@email.com"
              class="w-full px-4 py-2 bg-makoclaw-bg border border-makoclaw-border rounded focus-ring text-makoclaw-text placeholder-makoclaw-text-secondary"
              required
              :disabled="isLoading"
            />
          </div>

          <!-- Password -->
          <div>
            <label for="password" class="block text-sm font-medium mb-2 text-makoclaw-text">
              Password
            </label>
            <div class="relative">
              <input
                v-model="form.password"
                id="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="At least 8 characters"
                class="w-full px-4 py-2 bg-makoclaw-bg border border-makoclaw-border rounded focus-ring text-makoclaw-text placeholder-makoclaw-text-secondary"
                required
                minlength="8"
                :disabled="isLoading"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                :disabled="isLoading"
                class="absolute right-3 top-1/2 transform -translate-y-1/2 text-makoclaw-text-secondary hover:text-makoclaw-text disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                :title="showPassword ? 'Hide password' : 'Show password'"
              >
                <svg v-if="showPassword" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                </svg>
                <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-4.803m5.596-3.856a3.375 3.375 0 11-4.753 4.753m4.753-4.753L9.172 9.172m5.656 5.656l1.414 1.414M9 9h.008v.008H9V9m12 0a10.05 10.05 0 01-9.458 15M3 9c-1.657 0-3 1.343-3 3s1.343 3 3 3"></path>
                </svg>
              </button>
            </div>
            <p class="text-xs text-makoclaw-text-secondary mt-1">Must be at least 8 characters long</p>
          </div>

          <!-- Confirm Password -->
          <div>
            <label for="confirmPassword" class="block text-sm font-medium mb-2 text-makoclaw-text">
              Confirm Password
            </label>
            <div class="relative">
              <input
                v-model="form.confirmPassword"
                id="confirmPassword"
                :type="showConfirmPassword ? 'text' : 'password'"
                placeholder="Re-enter your password"
                class="w-full px-4 py-2 bg-makoclaw-bg border border-makoclaw-border rounded focus-ring text-makoclaw-text placeholder-makoclaw-text-secondary"
                required
                :disabled="isLoading"
              />
              <button
                type="button"
                @click="showConfirmPassword = !showConfirmPassword"
                :disabled="isLoading"
                class="absolute right-3 top-1/2 transform -translate-y-1/2 text-makoclaw-text-secondary hover:text-makoclaw-text disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                :title="showConfirmPassword ? 'Hide password' : 'Show password'"
              >
                <svg v-if="showConfirmPassword" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                </svg>
                <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-4.803m5.596-3.856a3.375 3.375 0 11-4.753 4.753m4.753-4.753L9.172 9.172m5.656 5.656l1.414 1.414M9 9h.008v.008H9V9m12 0a10.05 10.05 0 01-9.458 15M3 9c-1.657 0-3 1.343-3 3s1.343 3 3 3"></path>
                </svg>
              </button>
            </div>
          </div>

          <!-- Error Message -->
          <div v-if="errorMessage" :class="[
            'p-4 border rounded-lg text-sm transition-all duration-300',
            isBlockedError ? 'bg-red-500/20 border-red-500 text-red-400 shadow-lg shadow-red-500/20' : 'bg-red-500/20 border-red-500/50 text-red-400'
          ]">
            <div class="flex items-start gap-2">
              <svg v-if="isBlockedError" class="w-5 h-5 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
              </svg>
              <div class="flex-1">
                <p :class="isBlockedError ? 'font-bold' : 'font-medium'">{{ errorMessage }}</p>
                <button 
                  v-if="canDismissError" 
                  @click="clearError" 
                  class="mt-2 text-xs underline hover:no-underline opacity-75 hover:opacity-100"
                >
                  Cerrar mensaje
                </button>
              </div>
            </div>
          </div>

          <!-- Submit Button -->
          <button
            type="submit"
            :disabled="isLoading"
            class="w-full px-4 py-2 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white font-medium rounded transition-smooth disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ isLoading ? 'Creating account...' : 'Create Account' }}
          </button>
        </form>

        <!-- Login Link -->
        <p class="text-sm text-makoclaw-text-secondary text-center mt-6">
          Already have an account?
          <router-link to="/login" class="text-makoclaw-accent hover:text-makoclaw-accent-hover font-medium">
            Sign In
          </router-link>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/authStore'

const router = useRouter()
const authStore = useAuthStore()

const form = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})
const isLoading = ref(false)
const errorMessage = ref('')
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const isBlockedError = ref(false)
const canDismissError = ref(false)

const clearError = () => {
  errorMessage.value = ''
  isBlockedError.value = false
  canDismissError.value = false
}

const handleSignup = async () => {
  errorMessage.value = ''
  isBlockedError.value = false
  canDismissError.value = false

  // Validation
  if (!form.value.username || !form.value.email || !form.value.password || !form.value.confirmPassword) {
    errorMessage.value = 'Please fill in all fields'
    canDismissError.value = true
    return
  }

  if (form.value.password.length < 8) {
    errorMessage.value = 'Password must be at least 8 characters long'
    canDismissError.value = true
    return
  }

  if (form.value.password !== form.value.confirmPassword) {
    errorMessage.value = 'Passwords do not match'
    canDismissError.value = true
    return
  }

  isLoading.value = true

  try {
    const response = await fetch('/api/v1/auth/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username: form.value.username,
        email: form.value.email,
        password: form.value.password
      })
    })

    if (!response.ok) {
      const data = await response.json().catch(() => ({ error: 'Registration failed' }))
      
      // Check if it's a blocked email error
      const isBlocked = data.blocked === true
      const errorMsg = data.error || 'Registration failed'
      
      errorMessage.value = errorMsg
      isBlockedError.value = isBlocked
      canDismissError.value = true
      
      // Auto-clear non-blocked errors after 15 seconds
      if (!isBlocked) {
        setTimeout(() => {
          if (!isBlockedError.value) {
            clearError()
          }
        }, 15000)
      }
      // Blocked errors persist until manually dismissed
      
      throw new Error(errorMsg)
    }

    const data = await response.json()
    
    // Auto-login with returned token
    authStore.setCredentials(
      {
        username: data.user.username,
        email: data.user.email,
        role: data.user.role
      },
      data.token
    )

    // Redirect to dashboard
    await router.push('/dashboard')
  } catch (error) {
    console.error('Signup error:', error)
    errorMessage.value = error.message || 'Registration failed. Please try again.'
  } finally {
    isLoading.value = false
  }
}
</script>

