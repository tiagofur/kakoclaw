<template>
  <div class="min-h-screen flex items-center justify-center p-4 relative overflow-hidden bg-gradient-to-br from-slate-100 via-blue-50/40 to-indigo-50/30 dark:from-makoclaw-bg dark:via-makoclaw-surface dark:to-makoclaw-bg">
    <!-- Decorative background elements -->
    <div class="absolute inset-0 pointer-events-none overflow-hidden">
      <div class="absolute -top-32 -left-32 w-96 h-96 bg-indigo-200/30 dark:bg-cyan-500/5 rounded-full blur-3xl"></div>
      <div class="absolute -bottom-32 -right-32 w-96 h-96 bg-blue-200/40 dark:bg-blue-500/5 rounded-full blur-3xl"></div>
      <div class="absolute top-1/3 right-1/4 w-[500px] h-[500px] bg-cyan-100/20 dark:bg-teal-500/3 rounded-full blur-3xl"></div>
    </div>

    <div class="w-full max-w-md relative z-10">
      <!-- Back to home -->
      <router-link to="/" class="inline-flex items-center gap-1.5 text-sm text-slate-400 hover:text-blue-500 dark:text-makoclaw-text-secondary dark:hover:text-blue-400 transition-colors mb-6 group">
        <svg class="w-4 h-4 transition-transform group-hover:-translate-x-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
        </svg>
        Back to home
      </router-link>

      <!-- Logo & Branding -->
      <div class="text-center mb-8">
        <router-link to="/" class="inline-flex items-center gap-3 mb-4 group">
          <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white text-2xl shadow-lg shadow-blue-500/25 group-hover:shadow-blue-500/40 transition-shadow">🦈</div>
          <h1 class="text-4xl font-bold bg-gradient-to-r from-blue-500 to-cyan-500 bg-clip-text text-transparent">
            MakoClaw
          </h1>
        </router-link>
        <p class="text-slate-500 dark:text-makoclaw-text-secondary text-lg">Create your account</p>
      </div>

      <!-- Form Card -->
      <div class="rounded-2xl p-8 border shadow-2xl bg-white/80 backdrop-blur-sm border-slate-200/80 shadow-slate-300/30 dark:bg-makoclaw-surface/80 dark:backdrop-blur-sm dark:border-makoclaw-border dark:shadow-black/10">
        <h2 class="text-2xl font-bold mb-6 text-slate-900 dark:text-makoclaw-text">Sign Up</h2>

        <form @submit.prevent="handleSignup" class="space-y-5">
          <!-- Username -->
          <div>
            <label for="username" class="block text-sm font-medium mb-2 text-slate-700 dark:text-makoclaw-text">
              Username
            </label>
            <div class="relative">
              <div class="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-makoclaw-text-secondary">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path>
                </svg>
              </div>
              <input
                v-model="form.username"
                id="username"
                type="text"
                placeholder="Choose a username"
                class="w-full pl-11 pr-4 py-3 rounded-xl transition-all outline-none border bg-slate-50 border-slate-200 text-slate-900 placeholder-slate-400 focus:ring-2 focus:ring-blue-500/30 focus:border-blue-400 focus:bg-white dark:bg-makoclaw-bg/50 dark:border-makoclaw-border dark:text-makoclaw-text dark:placeholder-makoclaw-text-secondary/60 dark:focus:ring-blue-500/40 dark:focus:border-blue-500"
                required
                :disabled="isLoading"
              />
            </div>
          </div>

          <!-- Email -->
          <div>
            <label for="email" class="block text-sm font-medium mb-2 text-slate-700 dark:text-makoclaw-text">
              Email
            </label>
            <div class="relative">
              <div class="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-makoclaw-text-secondary">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path>
                </svg>
              </div>
              <input
                v-model="form.email"
                id="email"
                type="email"
                placeholder="your@email.com"
                class="w-full pl-11 pr-4 py-3 rounded-xl transition-all outline-none border bg-slate-50 border-slate-200 text-slate-900 placeholder-slate-400 focus:ring-2 focus:ring-blue-500/30 focus:border-blue-400 focus:bg-white dark:bg-makoclaw-bg/50 dark:border-makoclaw-border dark:text-makoclaw-text dark:placeholder-makoclaw-text-secondary/60 dark:focus:ring-blue-500/40 dark:focus:border-blue-500"
                required
                :disabled="isLoading"
              />
            </div>
          </div>

          <!-- Password -->
          <div>
            <label for="password" class="block text-sm font-medium mb-2 text-slate-700 dark:text-makoclaw-text">
              Password
            </label>
            <div class="relative">
              <div class="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-makoclaw-text-secondary">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path>
                </svg>
              </div>
              <input
                v-model="form.password"
                id="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="At least 8 characters"
                class="w-full pl-11 pr-12 py-3 rounded-xl transition-all outline-none border bg-slate-50 border-slate-200 text-slate-900 placeholder-slate-400 focus:ring-2 focus:ring-blue-500/30 focus:border-blue-400 focus:bg-white dark:bg-makoclaw-bg/50 dark:border-makoclaw-border dark:text-makoclaw-text dark:placeholder-makoclaw-text-secondary/60 dark:focus:ring-blue-500/40 dark:focus:border-blue-500"
                required
                minlength="8"
                :disabled="isLoading"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                :disabled="isLoading"
                class="absolute right-3.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:text-makoclaw-text-secondary dark:hover:text-makoclaw-text disabled:opacity-50 disabled:cursor-not-allowed transition-colors p-0.5"
                :title="showPassword ? 'Hide password' : 'Show password'"
              >
                <svg v-if="showPassword" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                </svg>
                <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-4.803m5.596-3.856A9.006 9.006 0 0112 5c4.478 0 8.268 2.943 9.542 7a10.06 10.06 0 01-4.293 5.036M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 3l18 18"></path>
                </svg>
              </button>
            </div>
            <p class="text-xs text-slate-400 dark:text-makoclaw-text-secondary/70 mt-1.5 ml-1">Must be at least 8 characters long</p>
          </div>

          <!-- Confirm Password -->
          <div>
            <label for="confirmPassword" class="block text-sm font-medium mb-2 text-slate-700 dark:text-makoclaw-text">
              Confirm Password
            </label>
            <div class="relative">
              <div class="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400 dark:text-makoclaw-text-secondary">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path>
                </svg>
              </div>
              <input
                v-model="form.confirmPassword"
                id="confirmPassword"
                :type="showConfirmPassword ? 'text' : 'password'"
                placeholder="Re-enter your password"
                class="w-full pl-11 pr-12 py-3 rounded-xl transition-all outline-none border bg-slate-50 border-slate-200 text-slate-900 placeholder-slate-400 focus:ring-2 focus:ring-blue-500/30 focus:border-blue-400 focus:bg-white dark:bg-makoclaw-bg/50 dark:border-makoclaw-border dark:text-makoclaw-text dark:placeholder-makoclaw-text-secondary/60 dark:focus:ring-blue-500/40 dark:focus:border-blue-500"
                required
                :disabled="isLoading"
              />
              <button
                type="button"
                @click="showConfirmPassword = !showConfirmPassword"
                :disabled="isLoading"
                class="absolute right-3.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:text-makoclaw-text-secondary dark:hover:text-makoclaw-text disabled:opacity-50 disabled:cursor-not-allowed transition-colors p-0.5"
                :title="showConfirmPassword ? 'Hide password' : 'Show password'"
              >
                <svg v-if="showConfirmPassword" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                </svg>
                <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-4.803m5.596-3.856A9.006 9.006 0 0112 5c4.478 0 8.268 2.943 9.542 7a10.06 10.06 0 01-4.293 5.036M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 3l18 18"></path>
                </svg>
              </button>
            </div>
          </div>

          <!-- Error Message -->
          <Transition name="fade">
            <div v-if="errorMessage" :class="[
              'p-4 border rounded-xl text-sm transition-all duration-300',
              isBlockedError
                ? 'bg-red-50 border-red-300 text-red-700 shadow-lg shadow-red-100 dark:bg-makoclaw-error/10 dark:border-makoclaw-error/50 dark:text-makoclaw-error dark:shadow-makoclaw-error/10'
                : 'bg-red-50 border-red-200 text-red-600 dark:bg-makoclaw-error/10 dark:border-makoclaw-error/30 dark:text-makoclaw-error'
            ]">
              <div class="flex items-start gap-3">
                <svg v-if="isBlockedError" class="w-5 h-5 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636"></path>
                </svg>
                <svg v-else class="w-5 h-5 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
                </svg>
                <div class="flex-1">
                  <p :class="isBlockedError ? 'font-bold' : 'font-medium'">{{ errorMessage }}</p>
                  <button
                    v-if="canDismissError"
                    @click="clearError"
                    class="mt-2 text-xs underline hover:no-underline opacity-75 hover:opacity-100 transition-opacity"
                  >
                    Dismiss
                  </button>
                </div>
              </div>
            </div>
          </Transition>

          <!-- Submit Button -->
          <button
            type="submit"
            :disabled="isLoading"
            class="w-full py-3.5 bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 text-white font-semibold rounded-xl transition-all shadow-lg shadow-blue-500/25 hover:shadow-blue-500/40 disabled:opacity-50 disabled:cursor-not-allowed disabled:shadow-none active:scale-[0.98]"
          >
            <span v-if="isLoading" class="inline-flex items-center gap-2">
              <svg class="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
              </svg>
              Creating account...
            </span>
            <span v-else>Create Account</span>
          </button>
        </form>

        <!-- Divider -->
        <div class="relative my-8">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-slate-200 dark:border-makoclaw-border"></div>
          </div>
          <div class="relative flex justify-center">
            <span class="px-4 text-sm text-slate-400 dark:text-makoclaw-text-secondary bg-white/80 dark:bg-makoclaw-surface">Already a member?</span>
          </div>
        </div>

        <!-- Login Link -->
        <router-link
          to="/login"
          class="block w-full py-3 text-center border-2 rounded-xl transition-all font-medium border-slate-200 hover:border-blue-400 text-slate-700 hover:text-blue-600 hover:bg-blue-50/50 dark:border-makoclaw-border dark:hover:border-blue-500/50 dark:text-makoclaw-text dark:hover:text-makoclaw-text dark:hover:bg-blue-500/5"
        >
          Sign In Instead
        </router-link>
      </div>

      <!-- Footer -->
      <p class="text-center text-slate-400/80 dark:text-makoclaw-text-secondary/60 text-xs mt-6">
        Secured with JWT authentication
      </p>
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
      let data
      const contentType = response.headers.get('content-type')
      if (contentType && contentType.includes('application/json')) {
        data = await response.json().catch(() => ({ error: 'Registration failed' }))
      } else {
        const text = await response.text().catch(() => 'Registration failed')
        data = { error: text }
      }

      const isBlocked = data.blocked === true
      let errorMsg = data.error || 'Registration failed'

      if (response.status === 409) {
        errorMsg = data.error || 'Username or email already exists. Please try different credentials.'
      } else if (response.status === 429) {
        errorMsg = 'Too many registration attempts. Please try again later.'
      } else if (response.status === 400) {
        errorMsg = data.error || 'Invalid registration data. Please check your input.'
      }

      errorMessage.value = errorMsg
      isBlockedError.value = isBlocked
      canDismissError.value = true

      if (!isBlocked) {
        setTimeout(() => {
          if (!isBlockedError.value) {
            clearError()
          }
        }, 15000)
      }

      throw new Error(errorMsg)
    }

    const data = await response.json()

    authStore.setCredentials(
      {
        username: data.user.username,
        email: data.user.email,
        role: data.user.role
      },
      data.token
    )

    await router.push('/dashboard')
  } catch (error) {
    console.error('Signup error:', error)
    errorMessage.value = error.message || 'Registration failed. Please try again.'
  } finally {
    isLoading.value = false
  }
}
</script>
