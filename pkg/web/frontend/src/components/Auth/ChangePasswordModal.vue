<template>
  <Teleport to="body">
    <Transition name="modal">
    <div class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4 z-modal">
      <div class="bg-makoclaw-surface border border-makoclaw-border rounded-lg max-w-md w-full shadow-lg">
        <!-- Header -->
        <div class="flex items-center justify-between p-4 border-b border-makoclaw-border">
          <h3 class="text-lg font-semibold">Change Password</h3>
          <button
            @click="$emit('close')"
            class="p-1 hover:bg-makoclaw-border rounded transition-smooth"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Content -->
        <form @submit.prevent="handleChangePassword" class="p-4 space-y-4">
          <!-- Current Password -->
          <div>
            <label for="current" class="block text-sm font-medium mb-2">
              Current Password
            </label>
            <div class="relative">
              <input
                v-model="form.current"
                id="current"
                :type="showCurrent ? 'text' : 'password'"
                placeholder="Enter current password"
                class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded focus-ring text-sm"
                required
                :disabled="isLoading"
              />
              <button
                type="button"
                @click="showCurrent = !showCurrent"
                :disabled="isLoading"
                class="absolute right-3 top-1/2 transform -translate-y-1/2 text-makoclaw-text-secondary hover:text-makoclaw-text disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                :title="showCurrent ? 'Hide password' : 'Show password'"
              >
                <svg v-if="showCurrent" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-4.803m5.596-3.856a3.375 3.375 0 11-4.753 4.753m4.753-4.753L9.172 9.172m5.656 5.656l1.414 1.414M9 9h.008v.008H9V9m12 0a10.05 10.05 0 01-9.458 15M3 9c-1.657 0-3 1.343-3 3s1.343 3 3 3"></path>
                </svg>
              </button>
            </div>
          </div>

          <!-- New Password -->
          <div>
            <label for="new" class="block text-sm font-medium mb-2">
              New Password
            </label>
            <div class="relative">
              <input
                v-model="form.new"
                id="new"
                :type="showNew ? 'text' : 'password'"
                placeholder="Enter new password (min 10 chars)"
                class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded focus-ring text-sm"
                required
                minlength="10"
                :disabled="isLoading"
              />
              <button
                type="button"
                @click="showNew = !showNew"
                :disabled="isLoading"
                class="absolute right-3 top-1/2 transform -translate-y-1/2 text-makoclaw-text-secondary hover:text-makoclaw-text disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                :title="showNew ? 'Hide password' : 'Show password'"
              >
                <svg v-if="showNew" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-4.803m5.596-3.856a3.375 3.375 0 11-4.753 4.753m4.753-4.753L9.172 9.172m5.656 5.656l1.414 1.414M9 9h.008v.008H9V9m12 0a10.05 10.05 0 01-9.458 15M3 9c-1.657 0-3 1.343-3 3s1.343 3 3 3"></path>
                </svg>
              </button>
            </div>
            <p class="text-xs text-makoclaw-text-secondary mt-1">Minimum 10 characters</p>
          </div>

          <!-- Confirm Password -->
          <div>
            <label for="confirm" class="block text-sm font-medium mb-2">
              Confirm Password
            </label>
            <div class="relative">
              <input
                v-model="form.confirm"
                id="confirm"
                :type="showConfirm ? 'text' : 'password'"
                placeholder="Confirm new password"
                class="w-full px-3 py-2 bg-makoclaw-bg border border-makoclaw-border rounded focus-ring text-sm"
                required
                :disabled="isLoading"
              />
              <button
                type="button"
                @click="showConfirm = !showConfirm"
                :disabled="isLoading"
                class="absolute right-3 top-1/2 transform -translate-y-1/2 text-makoclaw-text-secondary hover:text-makoclaw-text disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                :title="showConfirm ? 'Hide password' : 'Show password'"
              >
                <svg v-if="showConfirm" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-4.803m5.596-3.856a3.375 3.375 0 11-4.753 4.753m4.753-4.753L9.172 9.172m5.656 5.656l1.414 1.414M9 9h.008v.008H9V9m12 0a10.05 10.05 0 01-9.458 15M3 9c-1.657 0-3 1.343-3 3s1.343 3 3 3"></path>
                </svg>
              </button>
            </div>
          </div>

          <!-- Error Message -->
          <div v-if="errorMessage" class="p-3 bg-makoclaw-error/20 border border-makoclaw-error rounded text-makoclaw-error text-sm">
            {{ errorMessage }}
          </div>

          <!-- Success Message -->
          <div v-if="successMessage" class="p-3 bg-makoclaw-success/20 border border-makoclaw-success rounded text-makoclaw-success text-sm">
            {{ successMessage }}
          </div>

          <!-- Actions -->
          <div class="flex gap-3 pt-4 border-t border-makoclaw-border">
            <button
              type="button"
              @click="$emit('close')"
              class="flex-1 px-3 py-2 border border-makoclaw-border rounded hover:bg-makoclaw-border transition-smooth"
              :disabled="isLoading"
            >
              Cancel
            </button>
            <button
              type="submit"
              class="flex-1 px-3 py-2 bg-makoclaw-accent hover:bg-makoclaw-accent-hover text-white rounded transition-smooth disabled:opacity-50"
              :disabled="isLoading"
            >
              {{ isLoading ? 'Updating...' : 'Update Password' }}
            </button>
          </div>
        </form>
      </div>
    </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref } from 'vue'
import authService from '../../services/authService'

const emit = defineEmits(['close'])

const form = ref({
  current: '',
  new: '',
  confirm: ''
})
const isLoading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const showCurrent = ref(false)
const showNew = ref(false)
const showConfirm = ref(false)

const handleChangePassword = async () => {
  errorMessage.value = ''
  successMessage.value = ''

  if (!form.value.current || !form.value.new || !form.value.confirm) {
    errorMessage.value = 'Please fill in all fields'
    return
  }

  if (form.value.new.length < 10) {
    errorMessage.value = 'New password must be at least 10 characters'
    return
  }

  if (form.value.new !== form.value.confirm) {
    errorMessage.value = 'Passwords do not match'
    return
  }

  isLoading.value = true

  try {
    await authService.changePassword(form.value.current, form.value.new)
    successMessage.value = 'Password changed successfully'
    form.value = { current: '', new: '', confirm: '' }
    
    setTimeout(() => {
      emit('close')
    }, 1500)
  } catch (error) {
    console.error('Password change error:', error)
    errorMessage.value = error.response?.data?.message || 'Failed to change password'
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
</style>

