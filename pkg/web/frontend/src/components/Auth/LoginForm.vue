<template>
  <div class="min-h-screen bg-kakoclaw-bg flex items-center justify-center p-4">
    <div class="w-full max-w-md">
      <!-- Logo -->
      <div class="text-center mb-8">
        <h1 class="text-4xl font-bold text-kakoclaw-accent mb-2">KakoClaw</h1>
        <p class="text-kakoclaw-text-secondary">AI Agent Control Panel</p>
      </div>

      <!-- Form Card -->
      <div class="bg-kakoclaw-surface border border-kakoclaw-border rounded-lg p-8 shadow-lg">
        <h2 class="text-2xl font-bold mb-6 text-kakoclaw-text">Login</h2>

        <form @submit.prevent="handleLogin" class="space-y-4">
          <!-- Username -->
          <div>
            <label for="username" class="block text-sm font-medium mb-2">
              Username
            </label>
            <input
              v-model="form.username"
              id="username"
              type="text"
              placeholder="Enter your username"
              class="w-full px-4 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded focus-ring text-kakoclaw-text placeholder-kakoclaw-text-secondary"
              required
              :disabled="isLoading"
            />
          </div>

          <!-- Password -->
          <div>
            <label for="password" class="block text-sm font-medium mb-2">
              Password
            </label>
            <input
              v-model="form.password"
              id="password"
              type="password"
              placeholder="Enter your password"
              class="w-full px-4 py-2 bg-kakoclaw-bg border border-kakoclaw-border rounded focus-ring text-kakoclaw-text placeholder-kakoclaw-text-secondary"
              required
              :disabled="isLoading"
            />
          </div>

          <!-- Error Message -->
          <div v-if="errorMessage" class="p-3 bg-kakoclaw-error/20 border border-kakoclaw-error rounded text-kakoclaw-error text-sm">
            {{ errorMessage }}
          </div>

          <!-- Submit Button -->
          <button
            type="submit"
            :disabled="isLoading"
            class="w-full px-4 py-2 bg-kakoclaw-accent hover:bg-kakoclaw-accent-hover text-white font-medium rounded transition-smooth disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {{ isLoading ? 'Signing in...' : 'Sign In' }}
          </button>
        </form>

        <!-- Info -->
        <p class="text-xs text-kakoclaw-text-secondary text-center mt-4">
          Default credentials: admin / (set during setup)
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/authStore'
import authService from '../../services/authService'

const router = useRouter()
const authStore = useAuthStore()

const form = ref({
  username: '',
  password: ''
})
const isLoading = ref(false)
const errorMessage = ref('')

const handleLogin = async () => {
  if (!form.value.username || !form.value.password) {
    errorMessage.value = 'Please fill in all fields'
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    await authService.login(form.value.username, form.value.password)
    await router.push('/dashboard')
  } catch (error) {
    console.error('Login error:', error)
    errorMessage.value = error.response?.data?.message || 'Login failed. Please try again.'
  } finally {
    isLoading.value = false
  }
}
</script>
