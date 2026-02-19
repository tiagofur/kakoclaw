<template>
  <Teleport to="body">
    <div class="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
      <div class="bg-picoclaw-surface border border-picoclaw-border rounded-lg max-w-md w-full shadow-lg">
        <!-- Header -->
        <div class="flex items-center justify-between p-4 border-b border-picoclaw-border">
          <h3 class="text-lg font-semibold">Change Password</h3>
          <button
            @click="$emit('close')"
            class="p-1 hover:bg-picoclaw-border rounded transition-smooth"
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
            <input
              v-model="form.current"
              id="current"
              type="password"
              placeholder="Enter current password"
              class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded focus-ring text-sm"
              required
              :disabled="isLoading"
            />
          </div>

          <!-- New Password -->
          <div>
            <label for="new" class="block text-sm font-medium mb-2">
              New Password
            </label>
            <input
              v-model="form.new"
              id="new"
              type="password"
              placeholder="Enter new password (min 10 chars)"
              class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded focus-ring text-sm"
              required
              minlength="10"
              :disabled="isLoading"
            />
            <p class="text-xs text-picoclaw-text-secondary mt-1">Minimum 10 characters</p>
          </div>

          <!-- Confirm Password -->
          <div>
            <label for="confirm" class="block text-sm font-medium mb-2">
              Confirm Password
            </label>
            <input
              v-model="form.confirm"
              id="confirm"
              type="password"
              placeholder="Confirm new password"
              class="w-full px-3 py-2 bg-picoclaw-bg border border-picoclaw-border rounded focus-ring text-sm"
              required
              :disabled="isLoading"
            />
          </div>

          <!-- Error Message -->
          <div v-if="errorMessage" class="p-3 bg-picoclaw-error/20 border border-picoclaw-error rounded text-picoclaw-error text-sm">
            {{ errorMessage }}
          </div>

          <!-- Success Message -->
          <div v-if="successMessage" class="p-3 bg-picoclaw-success/20 border border-picoclaw-success rounded text-picoclaw-success text-sm">
            {{ successMessage }}
          </div>

          <!-- Actions -->
          <div class="flex gap-3 pt-4 border-t border-picoclaw-border">
            <button
              type="button"
              @click="$emit('close')"
              class="flex-1 px-3 py-2 border border-picoclaw-border rounded hover:bg-picoclaw-border transition-smooth"
              :disabled="isLoading"
            >
              Cancel
            </button>
            <button
              type="submit"
              class="flex-1 px-3 py-2 bg-picoclaw-accent hover:bg-picoclaw-accent-hover text-white rounded transition-smooth disabled:opacity-50"
              :disabled="isLoading"
            >
              {{ isLoading ? 'Updating...' : 'Update Password' }}
            </button>
          </div>
        </form>
      </div>
    </div>
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
